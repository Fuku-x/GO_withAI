package main

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/newmo-oss/ergo"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, err.Error(), errorAttrs(err)...)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// MCPサーバーを作成する
	server := /* TODO: mcp.NewServerでサーバーを作成 */

	// ツールを登録する
	mcp.AddTool(server, &mcp.Tool{
		Name: "godoc",
		// TODO: Descriptionを設定する
		// ヒント: go help doc を実行してオプションを確認しましょう
		Description: /* TODO: go docコマンドの説明とオプションを記述 */,
	}, GoDoc)

	// サーバーを実行する
	if err := server.Run(ctx, /* TODO: &mcp.StdioTransport{}を指定 */); err != nil {
		return ergo.Wrap(err, "failed to run mcp server")
	}

	return nil
}

// エラーから属性とスタックトレースをとりだす
func errorAttrs(err error) []slog.Attr {
	var frames []slog.Attr
	for i, frame := range ergo.StackTraceOf(err) {
		frames = append(frames, slog.Group(strconv.Itoa(i),
			"package", frame.PkgPath(),
			"file", frame.File(),
			"line", frame.Line(),
			"func", frame.FuncName(),
		))
	}
	return []slog.Attr{
		slog.GroupAttrs("error_attrs", slices.Collect(ergo.AttrsAll(err))...),
		slog.GroupAttrs("stacktrace", frames...),
	}
}
