# STEP 1: はじめてのエージェント

**目安時間: 30分**

## ADK Goとは

ADK (Agent Development Kit) は、Googleが提供するAIエージェントを構築するためのフレームワークです。
ADK Goを使うことで、Gemini APIを活用したAIエージェントをGoで簡単に作成できます。

このステップでは、ユーザーの入力をそのままオウム返しするシンプルなエージェントを作成します。

## AIエージェントとは

AIエージェントとは、LLM（大規模言語モデル）を中心に、外部ツールやデータソースと連携して、ユーザーの指示に基づいてタスクを自律的に実行するプログラムです。

単純なチャットボットとの違いは、エージェントが「ツール」を使って外部の情報を取得したり、アクションを実行したりできる点です。
例えば、「明日の天気を調べて」という質問に対して、天気APIを呼び出して最新の情報を取得し、回答することができます。

## LLMとモデル

LLM（Large Language Model）は、大量のテキストデータで学習した言語モデルです。
Gemini、GPT、Claudeなどが代表的なLLMです。

ADK Goでは、`gemini.NewModel`で作成した「モデル」を通じてLLMにアクセスします。
モデルはLLMとの通信を抽象化したもので、プロンプトを送信し、応答を受け取る役割を担います。

## モデルの作成

まず、Gemini APIを使用するためのモデルを作成します。
`gemini.NewModel`関数を使って、使用するモデル名とAPIキーを指定します。

```go
const modelName = "gemini-2.5-flash-lite"
model, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{APIKey: apikey})
if err != nil {
    return err
}
```

モデル名は[Gemini APIのドキュメント](https://ai.google.dev/gemini-api/docs/models)を参照してください。

## エージェントの作成

次に、`llmagent.New`関数を使ってエージェントを作成します。
`llmagent.Config`には以下の項目を設定します。

- `Name`: エージェントの名前
- `Model`: 使用するモデル
- `Description`: エージェントの説明
- `Instruction`: 生成AIへの指示（システムプロンプト）

```go
echoAgent, err := llmagent.New(llmagent.Config{
    Name:        "hello-agent",
    Model:       model,
    Description: "echo agent",
    Instruction: "あなたはユーザの入力をそのままオウム返しをするエージェントです。",
})
if err != nil {
    return err
}
```

`Instruction`はエージェントの振る舞いを決定する重要な設定です。
ここでは、オウム返しをするように指示しています。

`Instruction`は「システムプロンプト」とも呼ばれ、ユーザーからの入力とは別に、LLMに対して事前に与える指示です。
エージェントの役割、応答のスタイル、制約事項などを記述します。

## ランチャーの設定と実行

作成したエージェントを実行するために、ランチャーを設定します。

ランチャーは、エージェントを起動するためのエントリーポイントです。
`full.NewLauncher`は、コンソール（CLI）、Web API、Web UIなど複数のインターフェースを統合したランチャーを作成します。
コマンドライン引数（例: `web api webui`）で、どのインターフェースを有効にするかを指定します。

```go
config := &launcher.Config{
    AgentLoader: agent.NewSingleLoader(echoAgent),
}

l := full.NewLauncher()
if err = l.Execute(ctx, config, args); err != nil {
    return err
}
```

`full.NewLauncher`は、CLI、Web API、Web UIなど複数の実行方法をサポートするランチャーを作成します。

## 実行

`TODO`を埋めると次のコマンドで実行することができます。

```
$ go run . web api webui
```

`localhost:8080`でWebサーバーが起動し、ブラウザでエージェントと対話できます。
左上のドロップダウンからエージェントを選択して、メッセージを入力してみましょう。

入力した内容がそのまま返ってくれば成功です。

なお、ADK Webは開発およびデバッグ目的のためのものです。

## トラブルシューティング

### Rate Limit

Gemini APIの無料枠では、モデルごとに1日あたりのリクエスト数に上限があります。
例えば、`gemini-2.5-flash`や`gemini-2.5-flash-lite`は1日あたり20リクエストまでです。

上限に達した場合は、以下のいずれかの方法で対処してください。

* Google AI Studioでプロジェクトに支払い情報を紐づける
* 新しいプロジェクトを作成し、APIキーを再取得する

### APIキーのエラー

`GOOGLE_API_KEY`が設定されていない、または無効な場合はエラーになります。

```
$ echo $GOOGLE_API_KEY
```

で環境変数が設定されているか確認してください。設定されていない場合は以下を実行します。

```bash
export GOOGLE_API_KEY="your-api-key"
```

### ポート8080が使用中

別のプロセスがポート8080を使用している場合、サーバーが起動できません。

```bash
# 使用中のプロセスを確認
lsof -i :8080
```

該当プロセスを終了するか、別のターミナルで既に起動しているサーバーを停止してください。
