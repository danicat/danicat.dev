---
title: "Go開発者のためのGemini: Genkitによるエージェンティック・バックエンドの構築"
date: 2026-10-06
draft: false
categories:
  - Agent Development
tags:
  - agent-skills
  - gemini
  - genkit
  - golang
  - osquery
series:
  - Gemini for Go Developers
series_order: 4
heroStyle: big
slug: "building-agentic-backends-with-genkit-go"
description: "Genkit Goでエージェンティック・バックエンドを構築する方法を学びます。型付きフロー、dotpromptテンプレート、ミドルウェアとAgent Skills、マルチエージェント委譲、Cloud Runへのデプロイまでを解説します。"
summary: "Genkitを使ってGoでエージェンティック・バックエンドを構築する方法を、型付きフロー、dotpromptテンプレート、ミドルウェア、Agent Skills、ステートフルなマルチエージェント委譲を通じて学びます。"
proficiencyLevel: "Advanced"
dependencies:
  - "Go 1.25+"
  - "github.com/firebase/genkit/go"
  - "google.golang.org/genai"
---

[「Go開発者のためのGemini」の前回の章]({{< ref "/posts/20260825-gemini-for-go-developers-part-3-building-agents" >}})では、[GenAI SDK](https://pkg.go.dev/google.golang.org/genai)、[Genkit](https://genkit.dev)、そして[ADK](https://adk.dev)という3つの異なるライブラリを使って、Goでエージェントを構築する方法を探求しました。それぞれの開発スタイルの使い勝手を横並びで比較できるよう、あえてエージェントはシンプルな構成にとどめていました。ここからは、各フレームワークをさらに深く掘り下げていく時間です。

本記事では、Genkitを使ったエージェンティック・バックエンドの構築に焦点を当てます。開発環境のセットアップからGenkitアプリのクラウドへのデプロイまで、開発プロセスをエンドツーエンドでカバーします。また、Genkitのコア要素（フロー、ツール、プロンプト、プラグイン、ミドルウェアなど）の背後にある理論にも触れますが、本当に重要な実践から離れないよう、これらすべての概念を実際のアプリケーションと組み合わせて解説していきます。その題材として、昨年このブログで紹介したAI診断エージェント「AIDA（AI Diagnostic Agent）」を再訪し、モダナイズしていきます。

## AIDAの再訪：AI診断エージェント

AIDAは、[私が初めて作ったエージェント]({{< ref "/posts/20250531-diagnostic-agent" >}})でした。その目的は、[osquery](https://osquery.io/)というオープンソースツールを介してオペレーティングシステムにクエリを実行し、コンピュータのトラブルを診断することです。AIDAが誕生してから1年余りが経ち、その間に何度かアップデートを重ねてきましたが、今日に至るまでそのアーキテクチャ自体を根本から問い直すことはありませんでした。

AIDAを作った当時、モデルの能力は今とは異なる水準にあり、Agent Skillsのような手法もまだ発明されていませんでした。応答の品質を高めるため、[SQLiteを使った簡易的なRAGソリューション]({{< ref "/posts/20251103-building-aida-part-2" >}})を用いて、スキーマやクエリの知識をオンデマンドでエージェントのコンテキストに注入する必要がありました。ついこの間までクールで「最先端」だったものが、今ではすっかり古臭くぎこちなく感じられるのは面白いものです。もっとも、問題はRAGだけではありませんでした。AIDAのコードベース全体が、エンジニアリングされたというよりは場当たり的にハックされたものだったのです。さらに悪いことに、すべてが黎明期の「バイブコーディング（vibe coding）」で作られていました。

私は10年以上前から言い続けていますが、今こそこれまで以上に当てはまる真理があります。もしコードベースが足かせになっているなら、思い切って更地にしてゼロから作り直しましょう。コードベースの成長痛を経験した後にゼロからやり直すと、v2（あるいはN+1）は必ずと言っていいほど良くなります。何をすべきでないかを学び、最適化し、シンプルにできるからです。そして最高な点は何かというと、今の時代なら昼食前に3つのバージョンを作れてしまうことです。かつては書き直し（リライト）といえば数ヶ月もの期間と、多大な社内政治コストが必要でした。

AIDAをGenkitで書き直すことで、彼女を現代の標準に引き上げられるだけでなく、コードの保守性と可読性を劇的に向上させることができます。AIDAは単一のバイナリにコンパイルされ、スタンドアロンのアプリとしても、クラウドにデプロイ可能なWebサーバーとしても動作するようになります。

## コーディングエージェントとローカルワークフローのセットアップ

アプリケーションコードを書く前に、2つのものが必要です。開発用UIとコマンドラインユーティリティを提供するローカルの `genkit` CLI、そしてコーディングエージェントにGenkit Goの使い方を教えるAgent Skillsです。

### Genkit CLIとDev UI

`genkit` CLIをインストールするには、公式のインストールスクリプトを実行します。

```sh
curl -sL cli.genkit.dev | bash
```

`genkit` CLIは、Genkit Goにおけるローカル開発の中枢を担います。プログラムを `go run .` で直接実行すれば通常のGoバイナリとして動作しますが、実行コマンドの先頭に `genkit start --` を付けることで、**Genkit Developer UI**（デフォルトは `http://localhost:4000`）にアクセスできるようになります。ここでは、Genkitプログラムの各コンポーネントを個別に視覚的に確認し、テストすることができます。

```sh
genkit start -- go run .
```

Dev UIでは、登録されたフロー、プロンプト、エージェントを対話的にテストできるほか、プロンプト、ツール呼び出し、トークン使用量、レイテンシのトレースを検査できます。

Dev UIに加えて、`genkit` CLIはフローの実行、トレースの可視化、ドキュメントの閲覧を行うためのコマンドラインヘルパーも提供しています。最後のドキュメント閲覧機能は、実装フェーズにおけるモデルのハルシネーションを防ぐのに特に役立ちます。

以下はいくつかの使用例です。

```sh
# Run a specific flow once from the terminal and exit (--stream or --wait optional)
genkit flow:run diagnose '{"userQuestion": "how is my disk usage?"}' --non-interactive -- go run .

# Browse and read the embedded Genkit Go documentation offline
genkit docs:list go
genkit docs:read go/middleware.md
```

### 推奨されるAgent Skills

最近のSDKにおけるゴールドスタンダードは、公式のAgent Skillsを同梱して提供することであり、Genkitも例外ではありません。Genkit Goプログラムを開発するなら、探すべきスキルはこちらです。

- [`developing-genkit-go`](https://github.com/genkit-ai/skills): フロー、dotprompt、組み込みミドルウェア、実験的なエージェント機能を網羅した公式のGenkit Goスキルです。

また、Geminiモデルを使って開発するため、`gemini-api-dev` スキルもインストールしておくと便利です。
- [`gemini-api-dev`](https://github.com/google-gemini/gemini-skills): 最新のモデルIDや思考（thinking）設定を扱うための公式Geminiスキルです。

これらはVercelの [`skills` CLI](https://github.com/vercel-labs/skills)を使ってワークスペースにインストールできます。

```sh
npx skills add genkit-ai/skills --skill developing-genkit-go
npx skills add google-gemini/gemini-skills --skill gemini-api-dev
```

あるいは、[`kungfu`](https://github.com/danicat/kungfu) にカタログを登録して、ジャストインタイム（JIT）で読み込むこともできます。

```sh
kungfu catalog add genkit-ai/skills
kungfu catalog add google-gemini/gemini-skills
```

スキルの代わりにMCPサーバーを使いたい場合は、CLIに[MCPサーバー](https://genkit.dev/docs/go/mcp-server/)（`genkit mcp`）も同梱されています。最近の私は個人的にスキルの方へ傾倒していますが、特定のニッチなユースケース（ドキュメント検索など）においてはMCPにも依然として価値があります。

## コアコンセプト

Genkitは、もともとFirebaseチームによって作られ、その後独立したプロジェクトへと成長した生成AIアプリケーション向けのオープンソースフレームワークです。オリジナルの言語はJSですが、Genkit Goは熟練したGo開発者にとっても驚くほどイディオマティック（Goらしい自然な書き味）に感じられます。もっとも、ところどころにまだ少し粗削りな部分は残っていますが。`genkit.Handler` のようなアダプターや、ミドルウェア、強く型付けされたフローやプロンプトといったおなじみの構成要素のおかげで、典型的なGenkitプログラムは従来のWebサービスとさほど変わらない姿になります。

### フロー（Flows）

Genkitにおける基本単位は**フロー（flow）**です。これは、生成AIの呼び出しと前処理・後処理のロジックをまとめてラップする、強く型付けされた関数です。まずは、`osqueryi` を介してオペレーティングシステムにクエリを実行するツールを備えた、シングルショットの診断フローとしてAIDAをスタートさせましょう。

```go
type AIDARequest struct {
	OS           string `json:"os,omitempty" jsonschema_description:"The target operating system (defaults to runtime.GOOS)"`
	UserQuestion string `json:"userQuestion" jsonschema_description:"The diagnostic question to investigate"`
}

type AIDAResponse struct {
	Message string `json:"message" jsonschema_description:"The diagnostic findings and recommendations"`
}

type OsqueryInput struct {
	Query string `json:"query" jsonschema_description:"The SQL query to execute in osquery."`
}

g := genkit.Init(ctx,
	genkit.WithPlugins(&googlegenai.VertexAI{}),
	genkit.WithDefaultModel("vertexai/gemini-3.8-flash"),
)

runOsquery := genkit.DefineTool(g, "runOsquery", "Runs a system inspection query using osquery.",
	func(ctx *ai.ToolContext, in OsqueryInput) (string, error) {
		out, err := exec.CommandContext(ctx, "osqueryi", "--json", in.Query).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("osqueryi failed: %w: %s", err, out)
		}
		return string(out), nil
	},
)

diagnoseFlow := genkit.DefineFlow(
	g,
	"diagnose",
	func(ctx context.Context, req AIDARequest) (AIDAResponse, error) {
		if req.OS == "" {
			req.OS = runtime.GOOS
		}
		resp, err := genkit.Generate(ctx, g,
			ai.WithSystem(fmt.Sprintf("You are the Emergency Diagnostic Agent for diagnosing %s operating system failures. Your name is AIDA.", req.OS)),
			ai.WithPrompt(req.UserQuestion),
			ai.WithTools(runOsquery),
		)
		if err != nil {
			return AIDAResponse{}, fmt.Errorf("diagnose flow failed: %w", err)
		}
		return AIDAResponse{Message: resp.Text()}, nil
	},
)
```

`genkit.DefineTool` は `OsqueryInput` 構造体を使ってJSONスキーマを生成し、`runOsquery` ツールを登録します。一方、`genkit.DefineFlow` はコールバック関数とともに名前付きでフローを登録します。コールバックのシグネチャを見ると、型付きのHTTPハンドラーに非常によく似ていることがわかります。`context.Context` と強く型付けされたリクエスト構造体を受け取り、型付きのレスポンスと標準的なGoの `error` を返します。

`DefineTool`、`DefineFlow`、そして `Generate` の第1引数として渡されている変数 `g` に注目してください。`g` は `*genkit.Genkit` インスタンスであり、プログラム内のすべてのフロー、プロンプト、ツール、モデル、ミドルウェアの中央レジストリとして機能します。`main` 内で `genkit.Init` を使って一度だけ `g` を初期化し、それを明示的に渡していきます。

正直なところ、私はコードが明示的であることを好むため、`genkit.WithDefaultModel` はあまり好きではないのですが、多数のフローを定義する場合にはボイラープレートを少し減らすのに役立ちます。フロー内にモデルをハードコードするよりも優れた選択肢については、このあとプロンプトのセクションで見ていきます。

`genkit.Init` は、ランタイムレジストリと可観測性（オブザーバビリティ）スタックの両方をセットアップします。フロー内のすべてのモデル呼び出しとツール呼び出しは、そのフローのトレース配下の子スパンとして自動的に記録されます。また、フロー内でデータベースへのクエリや外部APIの呼び出しといったカスタムの前処理・後処理を行う場合、そのロジックを `genkit.Run` でラップすれば、トレースのウォーターフォール内に独立した名前付きサブステップとして記録されます。

開発中は、CLIやDev UIからフローを実行できます。本番環境では、フローを通常のGo関数呼び出し（例：`myFlow.Run`）として呼び出すことも、HTTPエンドポイントとして公開することもできます。`genkit.Handler` を使えば、たった1行でフローをHTTPハンドラーに変換できます。

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /diagnose", genkit.Handler(diagnoseFlow))

port := os.Getenv("PORT")
if port == "" {
	port = "3400"
}

// Note: "server" requires the server plugin:
// import "github.com/firebase/genkit/go/plugins/server"
log.Fatal(server.Start(ctx, "0.0.0.0:"+port, mux))
```

アプリケーションに登録されているすべてのフローを一度に公開したい場合は、`genkit.ListFlows(g)` をループ処理できます。

```go
mux := http.NewServeMux()
for _, flow := range genkit.ListFlows(g) {
	mux.HandleFunc("POST /"+flow.Name(), genkit.Handler(flow))
}
log.Fatal(server.Start(ctx, "0.0.0.0:"+port, mux))
```

なお、`genkit.Handler` 自体は認証処理を一切行わないため、公開エンドポイントは標準の `net/http` 認証ミドルウェアでラップするようにしてください。

### プロンプト（Prompts）

フローのコード内にプロンプトやモデル設定をハードコードする方法は、小規模なプロジェクトでは機能しますが、プロジェクトが大きくなるにつれて保守が難しくなります。今日ではコーディングエージェントが作業の大部分を担ってくれるとはいえ、長年培われてきたエンジニアリングのベストプラクティスを捨て去るべきではありません。コードの整理・構造化もその1つです。

Genkitは、[dotprompt](https://genkit.dev/docs/go/dotprompt/)（`*.prompt` ファイル）と呼ばれるフォーマットを使ったプロンプトテンプレートをサポートしています。`dotprompt` のプロンプトには、モデルへの指示だけでなく、モデルの選択や設定のためのメタデータ、スキーマ定義、システム指示、さらには会話履歴などを含めることができます。

AIDAのプロンプト、ツール、モデル設定を `diagnoseFlow` から切り出して `prompts/aida.prompt` に抽出すると、次のようになります。

```markdown
---
model: vertexai/gemini-3.8-flash
tools:
  - runOsquery
config:
  thinkingConfig:
    thinkingLevel: low
input:
  schema:
    os: string
    userQuestion: string
output:
  format: json
  schema:
    message: string
---
{{role "system"}}
You are the Emergency Diagnostic Agent for diagnosing {{os}} operating system failures. Your name is AIDA.

{{role "user"}}
{{userQuestion}}
```

デフォルトでは、`genkit.Init` は作業ディレクトリ内の `./prompts` から `.prompt` ファイルを自動的に読み込みます。エージェントを単一バイナリとして配布したい場合は、`//go:embed` を使って `prompts` ディレクトリをバイナリに直接埋め込み、`genkit.WithPromptFS` に渡すことができます。

```go
import "embed"

//go:embed prompts/*
var promptsFS embed.FS

func main() {
	g := genkit.Init(context.Background(), genkit.WithPromptFS(promptsFS))
	// ...
}
```

プロンプトにはコードから `genkit.LookupPrompt` メソッドでアクセスでき、`myPrompt.Execute` で実行します。

```go
aida := genkit.LookupPrompt(g, "aida")
resp, err := aida.Execute(ctx, ai.WithInput(map[string]any{
	"os":           runtime.GOOS,
	"userQuestion": "how is my disk usage?",
}))
```

しばらくGoを書いている方なら、あの `map[string]any` は私と同じくらい気になるはずです。ありがたいことに、Genkitでは**データプロンプト（data prompts）**（`genkit.LookupDataPrompt`）と `genkit.DefineSchemaFor` を使うことで、先ほど定義した `AIDARequest` と `AIDAResponse` 構造体を直接 `aida.prompt` にバインドできます。

```go
// In prompts/aida.prompt, reference the Go types by name:
//
// input:
//   schema: AIDARequest
// output:
//   schema: AIDAResponse
//
// Register both schemas with Genkit so the prompt file can resolve them:
genkit.DefineSchemaFor[AIDARequest](g)
genkit.DefineSchemaFor[AIDAResponse](g)

aida := genkit.LookupDataPrompt[AIDARequest, *AIDAResponse](g, "aida")
resp, _, err := aida.Execute(ctx, AIDARequest{
	OS:           runtime.GOOS,
	UserQuestion: "how is my disk usage?",
})
// resp is of type *AIDAResponse
// Note: The second return value is the raw *ai.ModelResponse
```

`DefineSchemaFor` と `LookupDataPrompt` を使えば、Go側で一度スキーマを定義するだけで、入力からパース済みの出力までコンパイル時の完全な型安全性を維持できます。

### プラグイン（Plugins）

Genkitは `plugins/` 名前空間配下のパッケージによって拡張します。プラグインの最も一般的な用途はモデルファミリーの登録ですが、`server` や `middleware` のようにより幅広い機能を提供するプラグインもあります。`server` は登録しなくても使えるという点で特殊なケースですが、ほとんどのプラグインは `genkit.Init` の呼び出し時に初期化します。

```go
import (
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/middleware"
)

g := genkit.Init(ctx,
	genkit.WithPlugins(
		&googlegenai.VertexAI{},  // Gemini via Vertex AI
		&middleware.Middleware{}, // Registers Retry, Fallback, Skills, etc.
	),
)
```

ミドルウェアという概念はほとんどのGo開発者にとっておなじみのはずですが、このプラグインはフォールバック、リトライ、そしてAgent Skillsといった重要な機能への入り口となるため、特筆しておく価値があります。次に、これを詳しく見ていきましょう。

### ミドルウェアとAgent Skills

AIDAは先ほど定義した `runOsquery` ツールを通じてすでに `osquery` コマンドを実行できますが、各オペレーティングシステム（`darwin`、`linux`、`windows`）にどのテーブルやカラムが存在するかについては、依然としてモデルの一般的な学習データに頼って推測しています。AIDAに正確なスキーマ知識を与えるため、`./skills` フォルダ内にOSごとのスキルを1つずつ用意します。これらのスキルは、Antigravityに[osqueryリポジトリ](https://github.com/osquery/osquery)へのリンクを渡し、テーブルの利用可能状況に基づいてOS固有のスキルを生成するよう依頼して作成したものです。スキルには各テーブルとそれぞれのスキーマが含まれています。

Genkitでは `ai.WithUse` を使って生成呼び出しにミドルウェアをアタッチします。そして `plugins/middleware` パッケージは、リトライ、モデルのフォールバック、および [Agent Skills](https://agentskills.io) の組み込み実装を提供しています。

```go
resp, err := genkit.Generate(ctx, g,
	ai.WithSystem(fmt.Sprintf("You are AIDA, a %s operating system diagnostic assistant.", runtime.GOOS)),
	ai.WithPrompt("My machine is running out of file descriptors. How do I inspect open files with osquery?"),
	ai.WithTools(runOsquery),
	ai.WithUse(
		&middleware.Retry{MaxRetries: 3, InitialDelayMs: 1000, BackoffFactor: 2},
		&middleware.Fallback{
			Models: []ai.ModelRef{
				googlegenai.ModelRef("vertexai/gemini-3.5-flash-lite", nil),
			},
		},
		&middleware.Skills{
			SkillPaths: []string{"./skills"},
		},
	),
)
```

HTTPミドルウェアと同様に、`ai.WithUse` は指定した順序で合成されます。`Fallback` の前に `Retry` を配置することで、プライマリモデルをリトライした上で、それでも失敗した場合にセカンダリモデルへフォールバックします。

`&middleware.Skills{}` に `./skills` を指定すると、Genkitは各サブディレクトリをスキャンし、スキルの説明文（description）をシステムプロンプトに注入するとともに、`use_skill` ツールを登録します。これにより、AIDAは適切なOSスキーマやクエリパターンをオンデマンドで読み込めるようになります。

なお、`dotprompt` とは異なり、現時点では `Skills` ミドルウェアは埋め込みファイルシステム（`embed.FS`）をサポートしていないため、デプロイメントには `skills` ディレクトリを同梱する必要があります。

また、`genkit.Init` 時に `&middleware.Middleware{}` を登録しているため、これらのミドルウェアを `prompts/aida.prompt` 内で直接宣言することも可能です。

```yaml
---
model: vertexai/gemini-3.8-flash
tools:
  - runOsquery
use:
  - name: genkit-middleware/retry
    config:
      maxRetries: 2
  - genkit-middleware/skills
---
```

さらに高度なユースケースでは、`ai.Middleware` インターフェースを実装してカスタムミドルウェアを作成し、生成、モデル、またはツール呼び出しを独自のロジックでインターセプトしてラップすることもできます。

### Developer UIでの確認

ここまでの完全なコードは、コンパニオンリポジトリの [`github.com/danicat/gemini-for-go-developers/part-4/01-flow`](https://github.com/danicat/gemini-for-go-developers/tree/main/part-4/01-flow) にあります。

Genkit Developer UIでAIDAをテストするには、そのディレクトリから `genkit start` でアプリケーションを起動します。

```sh
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_LOCATION=global # required for Gemini 3.x
genkit start -- go run .
```

`http://localhost:4000` を開き、サイドバーから **diagnose** フローを選択して、以下のテストペイロードで実行してみてください。

```json
{
  "userQuestion": "what is my battery status?"
}
```

私のマシンでこのプロンプトを実行した結果がこちらです（右側のパネルのトレースで、スキルがアクティベートされている様子に注目してください）。

![diagnoseフローを実行中のGenkit Developer UI](image-1.png)

## ステートフルなエージェント（実験的機能）

私は普段、実験的（experimental）な機能には手を出さないようにしているのですが、これだけは例外です。フローだけを使ってエージェントを実装することも十分に可能ですが、チャットボットや直列エージェント、並列エージェントといったより複雑なパターンを開発しようとすると、ゴルーチン、同期処理、エラーハンドリングなどを駆使して適切なパターンを自ら作り込むのにそれなりの労力がかかります。

エージェント抽象化の導入により、エージェント作成の宣言的な側面に焦点を当てたはるかに使い勝手の良いAPIが提供され、このプロセスが大幅に簡素化されます。

### エージェントの定義

このAPIはまだ実験的機能であるため、Genkitの初期化時に `genkit.WithExperimental()` を指定して有効化する必要があります。セッションストアをバックエンドに持つ、ステートフルなマルチターンのターミナルエージェントへとAIDAを進化させる方法は以下の通りです。

```go
import (
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/localstore"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
)

g := genkit.Init(ctx,
	genkit.WithExperimental(),
	genkit.WithPlugins(&googlegenai.VertexAI{}, &middleware.Middleware{}),
)

store := localstore.NewInMemorySessionStore[struct{}]()

aida := genkitx.DefineAgent(g, "aida",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem(fmt.Sprintf("You are AIDA, a %s operating system diagnostic assistant. Use runOsquery to inspect the host.", runtime.GOOS)),
		ai.WithTools(runOsquery),
		ai.WithUse(&middleware.Skills{SkillPaths: []string{"./skills"}}),
	},
	aix.WithSessionStore(store),
)

// Interactive terminal session maintaining state across turns
scanner := bufio.NewScanner(os.Stdin)
var sessionID string

fmt.Print("AIDA> ")
for scanner.Scan() {
	input := strings.TrimSpace(scanner.Text())
	if input == "" || input == "exit" {
		break
	}

	var opts []aix.InvocationOption[struct{}]
	if sessionID != "" {
		opts = append(opts, aix.WithSessionID[struct{}](sessionID))
	}

	out, err := aida.RunText(ctx, input, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Print("AIDA> ")
		continue
	}

	sessionID = out.SessionID
	fmt.Printf("\n%s\n\nAIDA> ", out.Message.Text())
}
```

`out.SessionID` を保持して後続のターンに渡すことで、セッション全体にわたって会話履歴が維持されます。型パラメータの `[struct{}]` は、カスタムのセッション状態を付与せず、メッセージ履歴のみを永続化したいことをGenkitに伝えています。

フロー周辺のコードを改善するために使ったのと同じテクニックが、エージェントにもそのまま使えます。プロンプトのハードコードを避けるには、`genkitx.DefinePromptAgent` を使って `prompts/aida.prompt` をエージェントに直接バインドできます。

また、リモートフロントエンドから接続できるようにAIDAをHTTP経由で公開するには、`genkitx.AllAgentRoutes` を使って標準のGoルーターにターン実行、スナップショット、中断（abort）の各エンドポイントをマウントします。

```go
mux := http.NewServeMux()
for _, route := range genkitx.AllAgentRoutes(g) {
	mux.HandleFunc(route.Pattern(), route.Handler())
}
```

ステートフルなエージェントの完全な実装は、コンパニオンリポジトリの [`github.com/danicat/gemini-for-go-developers/part-4/02-agent`](https://github.com/danicat/gemini-for-go-developers/tree/main/part-4/02-agent) で確認できます。

### マルチエージェント委譲

1つのタスクを遂行するのに複数のエージェントが必要になるケースもあります。マルチエージェント・アーキテクチャの利点には、関心の分離（エージェントの注意力を高め、コンテキストの劣化＝コンテキストロットを減らすため）や、並列化（応答時間を短縮するため）などがあります。

Genkitの実験的な `middlewarex.Agents` ミドルウェアを使うと、`aida` をメインのコーディネーター（調整役）ペルソナとして維持しつつ、AIDAの調査作業を2つのスペシャリストエージェントに分割できます。

- **`inspector`**: `runOsquery` とOS固有の `osquery` スキルを介して、ホストのテレメトリをクエリすることだけに専念します。
- **`researcher`**: Google検索グラウンディングを使用して、AIDAのメインの会話履歴を肥大化させることなく、見慣れないプロセス、バイナリ、エラーメッセージ、CVEをWeb上で調査します。

```go
// requires:
// import (
// 	"google.golang.org/genai"
// 	middlewarex "github.com/firebase/genkit/go/plugins/middleware/exp"
// )

inspector := genkitx.DefineAgent(g, "inspector",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem(fmt.Sprintf("Inspect %s system telemetry using osquery and report diagnostic findings.", runtime.GOOS)),
		ai.WithTools(runOsquery),
		ai.WithUse(
			&middleware.Retry{MaxRetries: 5},
			&middleware.Skills{SkillPaths: []string{"./skills"}},
		),
	},
	aix.WithDescription[struct{}]("Inspects host metrics, running processes, and system configuration using osquery."),
)

researcher := genkitx.DefineAgent(g, "researcher",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem("Research unfamiliar processes, binaries, error messages, and CVEs using Google Search and report concise findings."),
		ai.WithConfig(&genai.GenerateContentConfig{
			Tools: []*genai.Tool{
				{GoogleSearch: &genai.GoogleSearch{}},
			},
		}),
	},
	aix.WithDescription[struct{}]("Searches the web for information on unfamiliar processes, binaries, error codes, and CVEs."),
)

aida := genkitx.DefineAgent(g, "aida",
	aix.InlinePrompt{
		ai.WithModelName("vertexai/gemini-3.8-flash"),
		ai.WithSystem(fmt.Sprintf("You are AIDA, a %s operating system diagnostic assistant. Delegate host inspection to inspector and web research to researcher, then synthesize findings for the user.", runtime.GOOS)),
		ai.WithUse(
			&middleware.Retry{MaxRetries: 5},
			&middlewarex.Agents{
				Agents:         []aix.AgentRef{inspector.Ref(), researcher.Ref()},
				MaxDelegations: 5,
			},
		),
	},
	aix.WithSessionStore(store),
)
```

委譲（delegation）はデフォルトでは同期的に実行されますが、AIDAに長時間実行されるスペシャリストのタスクをバックグラウンドで開始させ、ターンをまたいでその結果を結合させたい場合は、`middlewarex.Agents` で `Async: true` を設定することもできます。

1つ心に留めておくべきなのは、レイテンシのトレードオフです。まったく同じ `"what is my battery status?"` というクエリを両方のバージョン（`thinkingLevel` を `low` に設定）で計測したところ、シングルエージェント版（`02-agent`）は3回のLLM呼び出しで **9.56秒** で完了したのに対し、このマルチエージェント版（`03-multi-agent`）は5回のLLM呼び出しで **17.02秒** かかりました。1回あたりの呼び出しレイテンシはほぼ同じ（約3.3秒）であるため、追加の約7.5秒は純粋に、`aida` が `inspector` に委譲してからそのレポートを要約するための、コーディネーターの2回分の追加往復（ラウンドトリップ）から生じています。

並列プログラミングとよく似ていて、小さな *N* に対するコーディネーションのオーバーヘッドは、単純なタスクをかえって遅くすることがあります。ちょっとしたクエリであれば単一のエージェントの方が高速です。しかし実際のシステムにおいては、ノイズの多いツール出力を隔離し、コンテキスト使用量を最適化し、全体的な応答品質を向上させるために、1つの「巨大エージェント（uber-agent）」を特化型のサブエージェントに分割することは、多くの場合その追加の往復に見合うだけの価値があります。

マルチエージェントの完全なサンプルは、コンパニオンリポジトリの [`github.com/danicat/gemini-for-go-developers/part-4/03-multi-agent`](https://github.com/danicat/gemini-for-go-developers/tree/main/part-4/03-multi-agent) にあります。

## 本番環境へのデプロイ

### ランタイム

Genkit Goアプリケーションは `net/http` を使用する標準的なGoバイナリにコンパイルされるため、独自のランタイムへのロックインはありません。つまり、HTTPサーバーをホストするために普段使っているプラットフォームなら基本的にどれでも、Genkitアプリの実行環境としてそのまま利用できるということです。

私の第一選択肢は常にサーバーレスのコンテナランタイムです。Google社内での私は主に **Google Cloud Run** を使用しており、Cloud Runでは提供されていない機能がどうしても必要な致命的なブロッカーにぶつかった場合にのみ、より複雑なプラットフォーム（**Google Kubernetes Engine (GKE)** や **Gemini Enterprise** など）を選択します。

また、かつてはこうしたプラットフォーム選定の意思決定が今よりもずっと重要だった、という点にも触れておく価値があります。以前はプラットフォームの乗り換え（リプラットフォーミング）に多大な労力がかかったため、皆最初から「正解」を選びたがる傾向がありました。しかしコーディングエージェントがある今、アプリを新しいプラットフォームに適応させる作業は数日（場合によっては数時間）で終わります。ですから、プロジェクトの初期段階から複雑さを持ち込む言い訳はもはや存在しないのです。

これが理論上の話です。そして実践においては、過去1年間でCloud Runでは実現できないケースに遭遇したことは一度もありませんでした。ただただ、それくらい便利なのです！もっとも、今後の記事でADKについて議論するときには少し話が変わってきます。Gemini Enterpriseエコシステムには、ADKとの間に注目すべきシナジーがいくつか存在するからです。

なお、通常はエージェント（特にAntigravityのようなGeminiベースのもの）がCloud Runへのデプロイで助けを必要とすることはありません。ですが、もしあなたのエージェントが機嫌を損ねているようなら、[github.com/google/skills](https://github.com/google/skills/tree/main/skills/cloud/cloud-run-basics) からCloud Runスキルをインストールして、少し背中を押してあげるとよいでしょう。

### 可観測性（Observability）

ローカル開発中、`genkit start` は完全な実行トレースと構造化ログを `.genkit/traces`（これは `.gitignore` に追加しておくべきです）に記録するため、Dev UIでそれらを検査できます。本番環境でも、監視バックエンド上でトークン使用量、レイテンシ、そしてステップごとのトレースに対する同じ可視性を確保したいところです。

Genkitは[OpenTelemetry](https://opentelemetry.io/)の上に構築されているため、主に2つの選択肢があります。Google Cloudにデプロイする場合は、`googlecloud` プラグイン（または内部で同じエクスポーターを実行する `firebase` プラグイン）を使えば、すべてがCloud Logging、Cloud Trace、Cloud Monitoringに送信されます。チームでDatadog、Grafana、Honeycombなどの別のバックエンドを使用している場合は、標準のOTLPエクスポーターをGenkitのトレーサープロバイダーに直接アタッチできます。

典型的なCloud Runデプロイメントでは、`genkit.Init` を呼び出す前にGoogle Cloudエクスポーターを有効にします。

```go
import "github.com/firebase/genkit/go/plugins/googlecloud"

func main() {
	ctx := context.Background()

	googlecloud.EnableGoogleCloudTelemetry(&googlecloud.GoogleCloudTelemetryOptions{
		// Raw prompts and responses are logged by default; disable this for sensitive data:
		DisableLoggingInputAndOutput: true,
		// Set to true if you want to test Cloud export locally under `genkit start`:
		ForceDevExport: false,
	})

	g := genkit.Init(ctx,
		// if you are deploying to GCP, auth via Gemini Enterprise (aka Vertex AI) is preferred
		genkit.WithPlugins(&googlegenai.VertexAI{}),
	)
	// ...
}
```

代わりにOTLPバックエンドを使用したい場合は、`genkit.Init` の前に標準のOpenTelemetryスパンプロセッサを登録し、そのシャットダウン関数を `defer` します。

```go
import (
	"context"
	"log"

	"github.com/firebase/genkit/go/core/tracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func registerOTLP(ctx context.Context) func(context.Context) error {
	// Configured via standard OTEL_EXPORTER_OTLP_* environment variables.
	exp, err := otlptracegrpc.New(ctx)
	if err != nil {
		log.Fatalf("failed to build OTLP exporter: %v", err)
	}

	tp := tracing.TracerProvider()
	tp.RegisterSpanProcessor(sdktrace.NewBatchSpanProcessor(exp))
	return tp.Shutdown
}
```

## まとめ

正直に言うと、AIDAを再構築するこの演習を通じてGenkitのコア機能の大部分に触れることはできましたが、当初予定していた究極のディープダイブにはまだ程遠いものです。とはいえ、これは嬉しい悲鳴というやつです。Genkitには興味深い機能があまりにも多く、1つの記事ですべてを語り尽くすのは不可能なのです。だからこそ、今回の記事ではエージェント設計のバックエンドの側面に焦点を絞り、A2UIを使ったフロントエンドやインタラクション設計については次回の記事に取っておくことにしました。

それでもなお、本記事はGenkitで何ができるかを示しただけでなく、過去12ヶ月ほどの間に業界そのものがいかに進化したかも浮き彫りにしています。かつてはカスタムのRAGパイプライン、エンベディングモデル、そして品質に疑問の残る数百行のPythonコード（これはPythonへの批判ではなく、私自身のコードへの批判です）を必要としたものが、今では型安全なGoとAgent Skillsミドルウェアへの呼び出しだけで解決できてしまいます。特に、コンパイルされたエージェントのキビキビとした動作感は最高で、Gemini 3.8 Flashのような高速なモデルと組み合わせたときにはなおさらです。

この記事をPython対Goの議論にするつもりはありませんし、どちらか一方を使うよう説得するつもりもありません。ですが、理由が何であれGoでエージェントを構築する予定があるなら、私はGenkitを強くお勧めします。この記事で見てきたすべての理由、次回取り上げる理由、そして私がまだ発見すらしていないその他すべての理由からね。 :)
