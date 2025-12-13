package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/newmo-oss/ergo"
)

type GoDocInput struct {
	// ドキュメントを取得するパッケージのインポートパス
	PkgPath string `json:"pkgpath" jsonschema:"the import path of pacakge"`
	// go docコマンドのオプション
	Options []string `json:"options" jsonschema:"options of go doc command"`
}

type GoDocResult struct {
	Doc string `json:"doc"`
}

func GoDoc(ctx context.Context, req *mcp.CallToolRequest, input *GoDocInput) (*mcp.CallToolResult, *GoDocResult, error) {
	tmpdir, err := os.MkdirTemp("", "godoc-")
	if err != nil {
		return nil, nil, ergo.Wrap(err, "failed to create temp dir")
	}

	// go mod initをする
	if err := execGoModInit(ctx, tmpdir); err != nil {
		slog.Error(err.Error(), errorAttrs(err))
		return nil, nil, ergo.Wrap(err, "failed to initlized module")
	}

	// go getをする
	if err := execGoGet(ctx, tmpdir, input.PkgPath); err != nil {
		slog.Error(err.Error(), errorAttrs(err))
		return nil, nil, ergo.Wrap(err, "failed to get package")
	}

	// execGoDocでgo docコマンドを実行する
	doc, err := execGoDoc(ctx, tmpdir, input.Options, input.PkgPath)
	if err != nil {
		slog.Error(err.Error(), errorAttrs(err))
		return nil, nil, ergo.Wrap(err, "failed to execute GoDoc tool")
	}

	// 結果を*GoDocResultとして返す
	return nil, &GoDocResult{Doc: doc}, nil
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
