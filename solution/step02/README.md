# STEP 2: Google検索ツールを使う

**目安時間: 15分**

## ツールとは

LLMは学習データに基づいて応答を生成しますが、学習時点以降の情報や、外部システムとの連携はできません。
この制限を克服するために、エージェントは「ツール」を使います。

ツールとは、エージェントが呼び出せる外部機能のことです。
LLMは会話の中で「このツールを使うべきだ」と判断すると、ツールを呼び出し、その結果を使って応答を生成します。

例えば:
- 検索ツール: 最新の情報をWeb検索で取得
- 計算ツール: 複雑な計算を正確に実行
- APIツール: 外部サービス（天気、株価など）にアクセス

このステップでは、Geminiに組み込まれているGoogle検索ツールを使って、検索機能を持つエージェントを作成します。

## Google検索ツールの追加

`llmagent.Config`の`Tools`フィールドに`geminitool.GoogleSearch{}`を追加することで、Google検索機能を持つエージェントを作成できます。

```go
searchAgent, err := llmagent.New(llmagent.Config{
    Name:        "search-agent",
    Model:       model,
    Description: "Google Search Agent",
    Instruction: "あなたはサーチエージェントです。ユーザの入力を検索して検索結果を日本語で要約してください。",
    Tools: []tool.Tool{
        geminitool.GoogleSearch{},
    },
})
if err != nil {
    return err
}
```

`geminitool`パッケージには、Geminiに組み込まれているツールが用意されています。
`GoogleSearch`はその1つで、Google検索を行うことができます。

## Instructionの変更

STEP 1ではオウム返しをするエージェントを作成しましたが、今回は検索エージェントとして振る舞うように`Instruction`を変更しています。

```go
Instruction: "あなたはサーチエージェントです。ユーザの入力を検索して検索結果を日本語で要約してください。",
```

## 実行

`TODO`を埋めると次のコマンドで実行することができます。

```
$ go run . web api webui
```

`localhost:8080`でWebサーバーが起動します。
エージェントを選択して、検索したいキーワードを入力してみましょう。

例えば「プログラミング言語Goの最新バージョン」と入力すると、Google検索を行い、その結果を要約して返してくれます。

## トラブルシューティング

### ツールの登録数

現在のADK Goでは、1つのエージェントに登録できるツールは1つまでです。
複数のツールを使用する場合は、ルートエージェントを作成し、マルチエージェント構成にする必要があります。

```go
// マルチエージェント構成の例
rootAgent, err := llmagent.New(llmagent.Config{
    Name:        "root-agent",
    Model:       model,
    Description: "Root Agent",
    Instruction: "ユーザーのリクエストに応じて適切なサブエージェントに処理を委譲してください",
    SubAgents: []agent.Agent{
        searchAgent,  // 検索エージェント
        godocAgent,   // GoDocエージェント
    },
})
```

### 検索結果が返ってこない

Google検索ツールはネットワーク接続が必要です。
ネットワークに接続されているか確認してください。

また、検索クエリが曖昧すぎると適切な結果が返らないことがあります。
具体的なキーワードで試してみてください。
