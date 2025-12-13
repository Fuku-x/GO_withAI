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
	// ハンドラを作成する
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

		// go docコマンドを実行する
		doc, err := execGoDoc(ctx, tmpdir, input.Options, input.PkgPath)
		if err != nil {
			slog.Error(err.Error(), errorAttrs(err))
			return nil, ergo.Wrap(err, "failed to execute GoDoc tool")
		}

		return &GoDocResult{Doc: doc}, nil
	}

	// functiontool.Newでツールを作成する
	godocTool, err := functiontool.New(functiontool.Config{
		Name: "GoDoc",
		// TODO: Descriptionを設定する
		// ヒント: go help doc を実行してオプションを確認しましょう
		Description: /* TODO: go docコマンドの説明とオプションを記述 */,
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
