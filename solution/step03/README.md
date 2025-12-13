# STEP 3: カスタムツールを作る

**目安時間: 45分**

## カスタムツールとは

STEP 2ではGeminiに組み込まれているGoogle検索ツールを使いました。
しかし、実際のアプリケーションでは、独自の処理を行うカスタムツールが必要になることがあります。

このステップでは、`functiontool`パッケージを使って、`go doc`コマンドを実行するカスタムツールを作成します。

## 入出力の型を定義する

まず、ツールの入力と出力を表す型を定義します。

```go
type GoDocInput struct {
    // ドキュメントを取得するパッケージのインポートパス
    PkgPath string `json:"pkgpath"`
    // go docコマンドのオプション
    Options []string `json:"options"`
}

type GoDocResult struct {
    Doc string `json:"doc"`
}
```

入力の型のフィールドには`json`タグを付けます。
これにより、LLMがツールを呼び出す際にJSONでパラメータを渡すことができます。

## ツールのハンドラを作成する

ツールが呼び出されたときに実行される処理をハンドラとして定義します。

```go
handler := func(ctx tool.Context, input *GoDocInput) (*GoDocResult, error) {
    // 一時ディレクトリを作成
    tmpdir, err := os.MkdirTemp("", "godoc-")
    if err != nil {
        return nil, err
    }

    // go mod init を実行
    if err := execGoModInit(ctx, tmpdir); err != nil {
        return nil, err
    }

    // go get でパッケージを取得
    if err := execGoGet(ctx, tmpdir, input.PkgPath); err != nil {
        return nil, err
    }

    // go doc コマンドを実行
    doc, err := execGoDoc(ctx, tmpdir, input.Options, input.PkgPath)
    if err != nil {
        return nil, err
    }

    return &GoDocResult{Doc: doc}, nil
}
```

ハンドラは`tool.Context`と入力の型へのポインタを受け取り、出力の型へのポインタとエラーを返します。

## functiontool.Newでツールを作成する

`functiontool.New`関数を使って、ツールを作成します。

```go
godocTool, err := functiontool.New(functiontool.Config{
    Name:        "GoDoc",
    Description: "...",
}, handler)
if err != nil {
    return nil, err
}
```

### Descriptionの重要性

`Description`には、ツールの説明を記述します。
LLMはこの説明を読んで、いつどのようにツールを使うべきかを判断します。

**Descriptionはエージェントの品質に大きく影響します。**
どのような説明を書けばLLMがツールを適切に使えるか、試行錯誤してみましょう。

ヒント: `go help doc`を実行すると、go docコマンドのオプションが確認できます。

```
$ go help doc
```

オプションの情報をDescriptionに含めることで、LLMがより適切にツールを使えるようになります。

以下は設定例です（そのまま使用しても構いません）：

```go
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
```

## エージェントにツールを追加する

作成したツールをエージェントに追加します。

### Instructionの設計

`Instruction`はエージェントの振る舞いを決定する重要な設定です。
どのような指示を書けばエージェントが期待通りに動作するか、試行錯誤してみましょう。

以下のような点を考慮してみてください：
- エージェントの役割は何か
- どのような質問に答えるべきか
- どのような形式で回答すべきか
- 答えるべきでない質問はあるか

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
    Tools: []tool.Tool{
        godocTool,
    },
})
```

## 実行

`TODO`を埋めると次のコマンドで実行することができます。

```
$ go run . web api webui
```

`localhost:8080`でWebサーバーが起動します。
エージェントを選択して、Goのパッケージについて質問してみましょう。

例えば「fmtパッケージのPrintln関数について教えて」と入力すると、`go doc`コマンドを実行してドキュメントを取得し、その内容を説明してくれます。

## トラブルシューティング

### go getに時間がかかる

初めてパッケージを取得する際は、ダウンロードに時間がかかることがあります。
ネットワーク環境によっては数十秒かかる場合もあります。

### パッケージが見つからない

存在しないパッケージ名を指定するとエラーになります。
正しいインポートパスを指定しているか確認してください。

例: `fmt`、`net/http`、`encoding/json` など

### ツールが呼び出されない

Descriptionが不十分だと、LLMがツールを使うべきタイミングを判断できません。
Descriptionにツールの用途やオプションを詳しく記述してください。
