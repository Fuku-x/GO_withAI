package main

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"strconv"

	"github.com/newmo-oss/ergo"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	// APIキーを環境変数から取得する
	apikey := os.Getenv("GOOGLE_API_KEY")

	if err := run(ctx, apikey, os.Args[1:]); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, err.Error(), errorAttrs(err)...)
		os.Exit(1)
	}
}

func run(ctx context.Context, apikey string, args []string) error {
	// 利用するモデル名を入れる
	// See: https://ai.google.dev/gemini-api/docs/models
	const modelName = "gemini-2.5-flash-lite"
	model, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{APIKey: apikey})
	if err != nil {
		return ergo.Wrap(err, "failed to create model", slog.String("model_name", modelName))
	}

	echoAgent, err := llmagent.New(llmagent.Config{
		Name:        "hello-agent",
		Model:       model,
		Description: "echo agent",
		// 生成AIへの指示を入れる
		Instruction: "あなたはユーザの入力をそのままオウム返しをするエージェントです。",
	})
	if err != nil {
		return ergo.Wrap(err, "failed to create agent")
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(echoAgent),
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, args); err != nil {
		return ergo.Wrap(err, "faield to execute launcher", slog.String("command_line_syntax", l.CommandLineSyntax()))
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
