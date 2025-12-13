package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/newmo-oss/ergo"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type GoDocInput struct {
	// ドキュメントを取得するパッケージのインポートパス
	PkgPath string `json:"pkgpath"`
	// go docコマンドのオプション
	Options []string `json:"options"`
}

type GoDocResult struct {
	Doc string `json:"doc"`
}

func NewGodocTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input *GoDocInput) (*GoDocResult, error) {
		tmpdir, err := os.MkdirTemp("", "godoc-")
		if err != nil {
			return nil, ergo.Wrap(err, "failed to create temp dir")
		}

		// go mod initをする
		if err := execGoModInit(ctx, tmpdir); err != nil {
			slog.Error(err.Error(), errorAttrs(err))
			return nil, ergo.Wrap(err, "failed to initlized module")
		}

		// go getをする
		if err := execGoGet(ctx, tmpdir, input.PkgPath); err != nil {
			slog.Error(err.Error(), errorAttrs(err))
			return nil, ergo.Wrap(err, "failed to get package")
		}

		// execGoDocでgo docコマンドを実行する
		doc, err := execGoDoc(ctx, tmpdir, input.Options, input.PkgPath)
		if err != nil {
			slog.Error(err.Error(), errorAttrs(err))
			return nil, ergo.Wrap(err, "failed to execute GoDoc tool")
		}

		// 結果を*GoDocResultとして返す
		return &GoDocResult{Doc: doc}, nil
	}

	godocTool, err := functiontool.New(functiontool.Config{
		Name: "GoDoc",
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
        -http
                Serve HTML docs over HTTP.
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
	}, handler)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to new GoDoc tool")
	}

	return godocTool, nil
}

func execGoDoc(ctx context.Context, dir string, options []string, pkgpath string) (string, error) {
	args := slices.Concat([]string{"doc"}, options, []string{pkgpath})
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", ergo.Wrap(err, "failed to execute go doc", slog.String("options", strings.Join(options, " ")))
	}

	return stdout.String(), nil
}

func execGoGet(ctx context.Context, dir string, pkgpath string) error {
	args := slices.Concat([]string{"get"}, []string{pkgpath})
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return ergo.Wrap(err, "failed to execute go get", slog.String("package", pkgpath))
	}

	return nil
}

func execGoModInit(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "init", "example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return ergo.Wrap(err, "failed to execute go mod init", slog.String("command", cmd.String()))
	}

	return nil
}
