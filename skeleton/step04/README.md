# STEP 4: MCPサーバーを作る

**目安時間: 30分**

## MCPとは

MCP (Model Context Protocol) は、LLMアプリケーションと外部ツールやデータソースを接続するためのオープンなプロトコルです。
2024年11月にAnthropicが公開し、2025年12月にLinux Foundation傘下の[Agentic AI Foundation](https://www.linuxfoundation.org/press/linux-foundation-announces-the-formation-of-the-agentic-ai-foundation)に寄贈されました。
現在はClaude、Gemini、ChatGPTなど主要なAIプラットフォームで採用されています。

STEP 3ではツールをエージェントのプロセス内に直接組み込みました。
しかし、ツールを別プロセス（サーバー）として分離することで、以下のメリットがあります:

- **再利用性**: 同じツールを複数のエージェントやアプリケーションから利用できる
- **言語非依存**: MCPサーバーはどの言語でも実装でき、MCPクライアントから呼び出せる
- **独立したデプロイ**: ツールとエージェントを別々にデプロイ・更新できる

このステップでは、STEP 3で作成したGoDocツールをMCPサーバーとして切り出し、エージェントから接続します。

## MCPサーバーの作成

`mcp`ディレクトリにMCPサーバーを作成します。

### サーバーの初期化

```go
server := mcp.NewServer(&mcp.Implementation{
    Name:    "godoc",
    Version: "v0.1.0",
}, nil)
```

`mcp.NewServer`関数で、サーバーの名前とバージョンを指定してサーバーを作成します。

### ツールの登録

```go
mcp.AddTool(server, &mcp.Tool{
    Name:        "godoc",
    Description: "...",
}, GoDoc)
```

`mcp.AddTool`関数で、サーバーにツールを登録します。
第3引数にはツールのハンドラ関数を指定します。

### Descriptionの重要性

STEP 3と同様に、`Description`はエージェントの品質に大きく影響します。
`go help doc`を実行してオプションを確認し、適切なDescriptionを設定しましょう。

以下は設定例です（そのまま使用しても構いません）：

```go
mcp.AddTool(server, &mcp.Tool{
    Name: "godoc",
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
```

### ハンドラ関数

MCPのハンドラ関数は、以下のシグネチャを持ちます。

```go
func GoDoc(ctx context.Context, req *mcp.CallToolRequest, input *GoDocInput) (*mcp.CallToolResult, *GoDocResult, error)
```

STEP 3の`functiontool`とは異なり、`*mcp.CallToolRequest`を受け取り、`*mcp.CallToolResult`も返すことができます。

### サーバーの実行

```go
if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
    return err
}
```

`StdioTransport`を使うことで、標準入出力を介してエージェントと通信します。

## MCPサーバーのビルド

MCPサーバーは事前にビルドしておく必要があります。

```
$ go build -o godocmcp ./mcp
```

## エージェントからMCPサーバーに接続する

`mcptoolset`パッケージを使って、MCPサーバーに接続します。

```go
godocMcp, err := mcptoolset.New(mcptoolset.Config{
    Transport: &mcp.CommandTransport{
        Command: exec.Command("./godocmcp"),
    },
})
if err != nil {
    return err
}
```

`CommandTransport`を使うことで、ビルドしたMCPサーバーをサブプロセスとして起動し、接続します。

## エージェントにToolsetを追加する

MCPサーバーから取得したツールは、`Toolsets`フィールドに追加します。

`Tools`ではなく`Toolsets`を使う点に注意してください。
`Toolset`は複数のツールをまとめて提供するインターフェースです。

### Instructionの設計

STEP 3と同様に、`Instruction`はエージェントの振る舞いを決定する重要な設定です。
どのような指示を書けばエージェントが期待通りに動作するか、試行錯誤してみましょう。

以下は設定例です（そのまま使用しても構いません）：

```go
godocAgent, err := llmagent.New(llmagent.Config{
    Name:        "go-doc-agent",
    Model:       model,
    Description: "GoDoc Agent",
    Instruction: `
        あなたはGoのパッケージについて調べるエージェントです。
        ユーザの質問にgo docコマンで調べながら答えてください。
        Goに関する質問以外は受けないようにしてください。
        Markdown形式で日本語で答えてください。
    `,
    Toolsets: []tool.Toolset{
        godocMcp,
    },
})
```

## 実行

`TODO`を埋めた後、まずMCPサーバーをビルドします。

```
$ go build -o godocmcp ./mcp
```

次に、エージェントを実行します。

```
$ go run . web api webui
```

`localhost:8080`でWebサーバーが起動します。
STEP 3と同様に、Goのパッケージについて質問してみましょう。

MCPサーバーを経由してツールが実行されますが、動作はSTEP 3と同じです。
MCPを使うことで、ツールを独立したプロセスとして管理でき、他のエージェントやアプリケーションからも再利用できるようになります。

## トラブルシューティング

### MCPサーバーのビルド忘れ

エージェント実行前に、MCPサーバーをビルドしておく必要があります。

```bash
go build -o godocmcp ./mcp
```

ビルドしていない場合、`exec: "./godocmcp": stat ./godocmcp: no such file or directory` のようなエラーが発生します。

### MCPサーバーに接続できない

`godocmcp`のパスが正しいか確認してください。
カレントディレクトリにバイナリがあることを前提としています。

```bash
ls -la ./godocmcp
```

### ToolsではなくToolsetsを使う

MCPサーバーから取得したツールは、`Tools`ではなく`Toolsets`フィールドに追加します。
間違えるとコンパイルエラーになります。
