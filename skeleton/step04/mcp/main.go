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
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "godoc",
		Version: "v0.1.0",
	}, nil)

	// ツールを登録する
	mcp.AddTool(server, &mcp.Tool{
		Name: "godoc",
		// TODO: Descriptionを設定する
		// ヒント: go help doc を実行してオプションを確認しましょう
		Description: `execute go doc command
Optionis
-all
        Show all the documentation for the package.
-c
        Respect case when matching symbols.
-cmd
        Treat a command (package main) like a regular package.
        Otherwise package main's exported symbols are hidden
        when showing the package's top-level documentation.
-short
        One-line representation for each symbol.
-src
        Show the full source code for the symbol. This will
        display the full Go source of its declaration and
        definition, such as a function definition (including
        the body), type declaration or enclosing const
        block. The output may therefore include unexported
        details.
-u
        Show documentation for unexported as well as exported
        symbols, methods, and fields.
`,
	}, GoDoc)

	// サーバーを実行する
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
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
