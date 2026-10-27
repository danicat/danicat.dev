---
title: "A2UI と Genkit でエージェント向けインターフェースを設計する"
date: 2026-10-27
draft: false
categories:
  - Agent Development
tags:
  - a2ui
  - agent-skills
  - ai-agents
  - gemini
  - genkit
  - golang
  - osquery
series:
  - Gemini for Go Developers
series_order: 5
heroStyle: big
slug: "designing-agentic-interfaces-with-a2ui-and-genkit"
description: "A2UI、カスタムコンポーネントカタログ、Human-in-the-loop インタラプトを活用して、Genkit Go エージェントをテキストチャットボットからジェネレーティブ UI へと進化させる方法を解説します。"
summary: "テキスト主体のターミナルチャットボットだった AIDA を、Genkit の A2UI ミドルウェア、カスタムカタログ、ツールインタラプトを駆使して対話型のジェネレーティブ UI へと進化させます。"
proficiencyLevel: "Advanced"
dependencies:
  - "Go 1.25+"
  - "github.com/firebase/genkit/go"
  - "google.golang.org/genai"
---

[前回の記事]({{< ref "/posts/20261006-building-agentic-backends-with-genkit-go" >}}) では、AI 診断エージェントである AIDA のバックエンドを再構築し、旧来の Python と SQLite による RAG 構成から、コンパイル済みの Genkit Go バイナリとオンデマンドの Agent Skills による構成へと刷新しました。

バックエンドが整ったところで、次はいよいよフロントエンドです。Go による頭脳とオンデマンドスキルを手に入れたとはいえ、AIDA はまだ実質的にはチャットボットに過ぎませんでした。プロンプトを入力すれば `osquery` を実行し、長文のテキストをそのまま返してくるだけです。いくら AIDA がレトロな世界観を持っているとはいえ、時代遅れの体験に縛り付けておく必要はありません。

最新のエージェント向け UI は、より表現力豊かな UI 要素を活用できるだけでなく、一律で静的な画面構成を押し付けるのではなく、ユーザーのその瞬間のニーズに応じて柔軟に適応するべきです。

本記事では、AIDA のインターフェースをモダナイズします。コードの解説に入る前に、まず現代的なエージェント向け UI に求められる要件と、[Agent to User Interface (A2UI)](https://a2ui.org/) プロトコルの仕組みを見ていきましょう。

## A2UI とは？

A2UI は Google が提唱するオープンプロトコルであり、AI エージェントがネイティブな UI コンポーネントを通じて直接対話できるようにする仕組みです。テキストのみによる返答にとどまらず、現在取り組んでいるタスクに合わせて、ダッシュボード、グラフ、テーブル、操作ボタンなどを動的に組み立てて提示できます。

UI のスタイリングや描画の主導権はフロントエンド側が完全に保持したまま、エージェントは会話の文脈やデータに基づいて「何を表示すべきか」を判断します。

A2UI の連携は、主に次の3つの概念を中心に成り立っています。

- **カタログ（Catalog）**: パネル、ゲージ、チャート、テーブル、ボタンといった UI の構成要素と、そのスキーマを定義した共通メニューです。バックエンドのエージェントとフロントエンドのクライアント双方が、あらかじめこのカタログに合意しておきます。組み込みの基本カタログを使うことも、ドメインに特化したカスタムコンポーネントを定義することも可能です。
- **サーフェス（Surfaces）**: クライアント上で描画される動的な UI キャンバス（`<a2ui-surface>`）です。エージェントは処理を進めながら、どのカタログコンポーネントを画面に配置し、そのプロパティをどう更新するかを指示する構造化 JSON メッセージ（`application/a2ui+json`）をストリーミング送信します。
- **アクション（Actions）**: ユーザーがコンポーネントを操作したとき（ボタンのクリック、テーブルの行選択、3D グラフのノード選択など）に、サーフェスからエージェントへ送り返される双方向イベントです。アクションにより、ユーザーが次のプロンプトを入力することなく、クリック操作がそのまま次の会話ターンへと直結します。

## AIDA の新しいフロントエンド仕様

私たちが目指すのは、A2UI を用いて、ハードコードされた固定 UI からエージェントが文脈に応じて自在に画面を構成できる仕組みへの移行です。そのためには、AIDA 専用のコンポーネントカタログを定義して利用可能にし、フロントエンド側で A2UI サーフェスを適切に描画できるよう再配線する必要があります。

以下が、新しく生まれ変わった診断コックピットの実際の動作画面です。AIDA は必要に応じて 3D プロセスグラフ、HUD ゲージ、そしてリアルタイムの `osquery` プロセステーブルを動的に組み立てます。

![3Dプロセスグラフ、HUDゲージ、インタラクティブなosqueryプロセスマトリクスを備えたリアルタイムA2UI v0.9診断サーフェスを描画するAIDA](aida-diagnostic-surface.png)

以下のプロンプトは、この要件を仕様へと落とし込んだものです。

```markdown
Upgrade AIDA's web interface from a static chat UI to a generative diagnostic cockpit using A2UI and Genkit.

Requirements:
1. Keep the retro NEC PC-98 cyberpunk look and feel (AIDA's pixel avatar, CRT scanlines, and neon palette), with a fixed host-status sidebar showing live osquery system vitals.
2. Replace the plain Markdown chat output with an A2UI surface powered by a custom "aida" component catalog for system diagnostics: HUD panels, gauges, charts, filterable data tables, 3D system visualisations, text blocks, and action buttons.
3. Wire up two-way A2UI actions so clicking a table row, 3D node, or button sends a follow-up diagnostic action back to AIDA.
4. Add a human-in-the-loop shell command tool so AIDA can actually fix issues she finds on the host, pausing for explicit Approve/Deny confirmation in the UI before executing any command.
/grill-me
```

このプロンプトだけで一発で完璧な出力を得ようと期待してはいけません。実際には、後述の「推奨スキル」セクションで説明するように、コーディングエージェントに適切なスキル、MCP、ツールを渡しておくことが極めて重要です。

プロンプトの末尾にある `/grill-me` スキルのヒントにも注目してください。このスキルは Antigravity（および多くの最新コーディングエージェント）に標準搭載されており、指示が曖昧だったり詳細が不足していたりする場合にエージェントから質問を投げかけさせる役割を持ちます。これにより構想が洗練され、最初から完璧なプロンプトを書かなければならないというプレッシャーから解放されます。

上記のプロンプトは非常にシンプルですが、`/grill-me` が裏で多くの作業を肩代わりしてくれます。後述するように、コンポーネントカタログの全詳細をプロンプトに書き込まなくても済むのは、不足している情報を Antigravity が対話を通じて質問し補ってくれるからです。

### 推奨スキル

Genkit を使用するため、公式の [Genkit Skills](https://github.com/genkit-ai/skills)（`developing-genkit-go` や `developing-genkit-js` を含む）をインストールして有効化しておくことを強くお勧めします。これらのスキルには、エージェントやフローの構築方法だけでなく、クライアント側およびサーバー側の双方で A2UI 要素を扱うノウハウが含まれています。

さらに、`gemini-api-dev` スキルを導入しておくことで、エージェントが古いモデルへ後退せず最新モデルを選択できるようになります。このスキルは [Google Skills](https://github.com/google-gemini/gemini-skills/blob/main/skills/gemini-api-dev/SKILL.md) リポジトリから単体で導入することも、[Gemini Cloud Developer Plugin](https://cloud.google.com/blog/topics/developers-practitioners/introducing-the-google-cloud-developer-plugin-for-ai-coding-agents) の一部として導入することも可能です。

## A2UI カタログの設計

従来のフロントエンド開発では、満たすべきユーザージャーニーに合わせてインターフェースを設計します。A2UI アプリケーションでもその基本は変わりませんが、従来のような固定されたアプリケーション画面を設計するわけではありません。その代わりに、目的を達成するために必要な個々のコンポーネントを設計することに集中します。

これらのコンポーネントが集まってエージェントの**カタログ**を構成し、エージェントがオンデマンドで UI を適応させる柔軟性をもたらします。A2UI には一般的な用途をカバーする**基本カタログ（Basic Catalog）**が付属しているため、必ずしもゼロからカタログを設計する必要はありません。しかし、今回は独自の世界観と体験を作り出すため、AIDA 専用のカスタムカタログを設計するのが最適な選択肢でした。

AIDA のために設計した7つのコンポーネントは以下のとおりです。

- **Panel**: 関連するコンポーネントを行、列、グリッド状にグループ化するフレーム付き HUD コンテナです。ステータスバッジ（`nominal`、`warning`、`critical`）を備えており、異常のあるサブシステムを一目で把握できます。
- **Scene3D**: システムトポロジー（プロセスツリー、CPU コア、メモリマトリクス、ネットワークメッシュ、ディスクセクター、サーマルゾーンなど）を視覚化するためのインタラクティブな 3D ワイヤーフレームキャンバスです。クリック可能なノードを備え、ドリルダウン調査が可能です。
- **CyberChart**: CPU 負荷、ネットワークスループット、レイテンシなどのメトリクスを描画する複数系列チャート（折れ線、棒、面、散布図、レーダー）です。
- **HudGauge**: ディスク使用率、バッテリー残量、スワップ負荷などの容量やパーセンテージを視覚的に表示するメーター（円弧、プラズマバー、セグメントバー）です。
- **DataMatrix**: `osquery` の調査結果を表示するソート・フィルタリング可能なテーブルです。行をクリックすると、その項目に対する追加調査がトリガーされます。
- **TextBlock**: 要約、コマンド出力、コピー可能なスニペットを表示する等幅フォントの診断ノートです。
- **ActionButton**: 不審なプロセスの詳細調査や修復処理の実行など、AIDA がワンクリックで実行できる次のアクションを提示するための対話型ボタンです。

基本カタログは見た目の美しさや豊かな体験よりも汎用的な機能性を重視しているため、こうしたドメイン特化型の計測器・計器類を設計することは大きな差別化につながります。

## Go バックエンドでの A2UI の組み込み

バックエンドで AIDA に新しい UI 構成要素へのアクセス権を与えるため、起動時に Genkit A2UI プラグインを有効化し、診断ツールと並んでカスタムコンポーネントカタログを読み込みます。

```go
// internal/agent/agent.go
g := genkit.Init(ctx,
	genkit.WithExperimental(),
	genkit.WithPromptDir("./prompts"),
	genkit.WithDefaultModel("vertexai/gemini-3.8-flash"),
	genkit.WithPlugins(
		&googlegenai.VertexAI{},
		&middleware.Middleware{},
		&a2uix.A2UI{}, // A2UI ミドルウェアの登録
	),
)

// 1. カスタム "aida" A2UI コンポーネントカタログの登録
if _, err := a2uix.LoadCatalog(g, AidaCatalog()); err != nil {
	return nil, err
}

// 2. 読み取り専用診断ツール (osquery) と HITL シェルコマンド (command) の登録
osquery.RegisterTool(g, runner)
command.RegisterTool(g)

// 3. prompts/aida.prompt から単一の AIDA エージェントを定義
aidaAgent := genkitx.DefinePromptAgent(g, "aida",
	aix.WithDescription("AIDA — AI Diagnostic Agent"),
	aix.WithSessionStore(sessionStore),
	aix.WithNamedPrompt("aida", PromptInput{OS: runtime.GOOS}),
)
```

ここでは、Gemini Enterprise（Vertex AI）およびミドルウェアプラグインと並んで `&a2uix.A2UI{}` で Genkit を初期化し、診断・修復ツールを登録した上で、`aida` プロンプトエージェントを `prompts/aida.prompt` にバインドしています。最大のポイントは `a2uix.LoadCatalog(g, AidaCatalog())` です。これによりカスタムコンポーネントカタログが Genkit ランタイムに登録され、A2UI ミドルウェアがそれらのコンポーネントをモデルに提示し、モデルの出力を検証できるようになります。

### カタログと A2UI ミドルウェアの定義

Go では、カスタム A2UI カタログ（`a2uix.Catalog`）はコンポーネント定義のリストとして表現されます。各エントリにはコンポーネント名 `Name`、人間向けの説明 `Description`、および期待されるプロパティ `Props` を指定します。

```go
// internal/agent/catalog.go
func AidaCatalog() *a2uix.Catalog {
	return &a2uix.Catalog{
		ID: "aida",
		Components: []a2uix.CatalogComponent{
			{
				Name:        "HudGauge",
				Description: "A visual meter for capacity and percentage metrics.",
				Props:       "label: string (required); value: number (required); unit?: string; status?: 'nominal' | 'warning' | 'critical'",
			},
			{
				Name:        "ActionButton",
				Description: "An interactive button that triggers a follow-up diagnostic or remediation action.",
				Props:       "label: string (required); action: string (required); variant?: 'primary' | 'warning' | 'danger'",
			},
			// ... Panel, Scene3D, CyberChart, DataMatrix, TextBlock
		},
	}
}
```

バックエンド側では、A2UI は Genkit ミドルウェアとして機能します。モデルは通常の Markdown テキストを返す代わりに、描画したいコンポーネントを記述した `a2ui` 形式の JSON ブロックを出力します。ミドルウェアはストリーム上でこれらのブロックをインターセプトし、登録済みカタログに照らして検証した上で、ブラウザ向けに `application/a2ui+json` メッセージとしてパッケージ化します。

このパイプラインは、`prompts/aida.prompt` 内で宣言的に構成します。

```yaml
---
model: vertexai/gemini-3.8-flash
description: AIDA — AI Diagnostic Agent
input:
  schema:
    os: string
tools:
  - run_osquery
  - runCommand
config:
  thinkingConfig:
    thinkingLevel: LOW
  tools:
    - googleSearch: {}
use:
  - genkit-middleware/skills:
      skillPaths:
        - ./skills
  - a2ui:
      catalogId: aida
      validate: strict
---
```

デフォルトでは、`a2ui` ミドルウェアは登録された `"aida"` カタログからコンポーネントの説明とプロパティ仕様を読み出し、それらをシステムプロンプト（`instructions: system`）へ直接注入します。これにより、`internal/agent/catalog.go` が単一の信頼できる情報源（Single Source of Truth）として保たれます。Go 側でコンポーネントを修正すれば、手動でスキーマを同期させることなくプロンプト側も自動的に更新されます。

モデルが `a2ui` ブロックをストリーミング出力する際、`validate: strict` によって各 JSON メッセージが有効な A2UI エンベロープを持ち、`"aida"` カタログに存在するコンポーネント型のみを使用しているかが厳格にチェックされます。モデルが存在しないコンポーネント型を捏造したり、壊れた JSON を出力したりした場合は、即座にエラーとなって早期失敗（Fail-Fast）します。

## ジェネレーティブなフロントエンドの構築

Go バックエンドから検証済みの A2UI メッセージがストリーミングされると、ブラウザ側ではそれらの JSON ペイロードを実際の DOM 要素に変換するクライアントランタイムが必要になります。`web/` 配下のフロントエンドは、TypeScript と [Lit](https://lit.dev/) で作られた軽量アプリケーションであり、以下の3つの役割を担います。

1. **コンポーネントカタログの対応付け**: クライアント側では、7つの `"aida"` 計器のそれぞれを Lit の Web コンポーネント（`web/src/aida-catalog/`）として登録し、クライアント側のプロパティスキーマとペアにして管理します。Go バックエンドはコンポーネントの型名のみを検証するため、ブラウザ側がこのスキーマを使って受け取ったプロパティを検証し、デフォルト値を適用してデータをバインドしてから描画します。
2. **ストリームからのサーフェス描画**: Genkit のクライアント SDK が `/agents/aida` に接続し、受信した A2UI メッセージを A2UI の `MessageProcessor` に渡します。プロセッサはチャンクが届くたびに、画面上の `<a2ui-surface>` 要素をリアルタイムで更新します。
3. **ユーザーアクションのエージェントへの送信**: アクションボタン、テーブルの行、または 3D ノードをクリックすると、A2UI アクションがディスパッチされます。クライアントはそのイベントを、チャットの次のターンとして Go バックエンドに送り返します。

```ts
// web/src/aida-cockpit.ts
const agent = remoteAgent({ url: "/agents/aida" });
const chat = agent.chat();

const processor = new MessageProcessor([aidaCatalog], async (action) => {
  // UI のクリックイベント（ボタン、テーブル行、3D ノード）を Go エージェントへ送り返す
  const turn = chat.sendStream({ message: actionToMessage(action as any) });
  await consumeTurnStream(turn);
});
```

コックピットを最初に開いたとき、左側のサイドバーは `/api/system-status` をポーリングしてリアルタイムの `osquery` システム状態を表示し、メインエリアにはクイック診断ランチャーが並びます。

![PC-98風ピクセルアバター、osqueryによるリアルタイムホスト状態サイドバー、クイック診断ランチャーカードを備えたAIDA診断コックピットの初期画面](aida-cockpit-landing.png)

ランチャーをクリックするか質問を入力すると、プロンプトが `/agents/aida` に送信されます。AIDA が `osquery` の結果をストリーミングで返送するにつれて、`MessageProcessor` がランチャー画面を先ほど確認した動的な診断サーフェスへと置き換えていきます。

## エージェントインタラプトによる Human-in-the-loop の実現

`osquery` によるマシンの調査は完全に読み取り専用であるため、AIDA が自律的にクエリを実行しても安全です。しかし、双方向の A2UI アクションを導入したことで、暴走したプロセスの強制終了やキャッシュの消去など、発見した問題の修復を提案する `ActionButton` を描画できるようにもなりました。

ホスト上でのシェルコマンドの実行を野放しにするわけにはいきません。こうしたアクションを安全に制御するため、Genkit はネイティブな Human-in-the-loop 機構である**ツールインタラプト（tool interrupts）**を提供しています。ツールは現在のターンを一時停止し、ユーザーの明示的な承認を待ってから、要求された副作用を実行します。

### ツールの実行一時停止と再開

Go では、`genkitx.DefineInterruptibleTool` を使って `runCommand` を定義します。この関数は、AIDA がツールを初回実行しているのか、それともユーザー入力後に再開しているのかを示す確認ペイロード（オプション）を受け取ります。

1. **初回呼び出し（一時停止）:** AIDA が最初に `runCommand` を呼び出した時点では、確認ペイロードは `nil` です。シェルコマンドを即座に実行するのではなく、ツールは提案されたコマンド、実行理由、リスクレベルを含む `tool.Interrupt(...)` を返してターンを一時停止します。
2. **再開呼び出し（実行または拒否）:** ユーザーが UI 上で **Approve（承認）** または **Deny（拒否）** をクリックすると、Genkit はユーザーの決定内容を渡してツール関数を再度呼び出します。

```go
// internal/command/tool.go
package command

func RegisterTool(g *genkit.Genkit) {
	genkitx.DefineInterruptibleTool(g, "runCommand",
		ai.WithDescription("Runs a shell command on the host with operator approval"),
		func(ctx *ai.ToolContext, input CommandInput, res *CommandConfirmation) (CommandOutput, error) {
			// 1. 初回呼び出し：ターンを一時停止し、オペレーターに承認を求める
			if res == nil {
				return CommandOutput{}, tool.Interrupt(CommandInterrupt{
					Command: input.Command,
					Reason:  input.Reason,
					Risk:    input.Risk,
				})
			}

			// 2. 再開呼び出し：拒否された場合、エラーではなく構造化フィードバックを返す
			if !res.Approved {
				return CommandOutput{
					Command:  input.Command,
					Status:   "rejected",
					Stderr:   fmt.Sprintf("Operator denied execution: %s", res.Feedback),
					Executed: false,
				}, nil
			}

			// 3. 承認時：30秒のタイムアウト付きでコマンドを実行する
			execCtx, cancel := context.WithTimeout(ctx.Context, 30*time.Second)
			defer cancel()
			out, err := exec.CommandContext(execCtx, "sh", "-c", input.Command).CombinedOutput()
			// ... stdout と終了コードを含む CommandOutput を返す
		},
	)
}
```

`tool.Interrupt` はターンを一時停止する際、`CommandInterrupt` 構造体（`Command`、`Reason`、`Risk`）をレスポンスストリームに直接シリアライズして流し込むため、フロントエンドは追加の API エンドポイントを用意することなくユーザー確認用プロンプトを表示するために必要な全情報を取得できます。

### コックピットでのコマンド承認と拒否

`runCommand` がインタラプトをトリガーすると、ストリーミング中のターンは一時停止し、ブラウザはレスポンスオブジェクトを通じてインタラプトペイロードを受け取ります。コックピットは AIDA のステータスバッジを `APPROVAL REQUIRED`（承認待ち）へと切り替え、インラインの承認カードを描画します。そのカードには、実行予定の正確なコマンド、理由、リスクレベルが表示され、任意のフィードバック入力欄と **Approve** / **Deny** ボタンが配置されます。

![提案されたシェルコマンド、リスクバッジ、フィードバック入力欄、承認・拒否ボタンを表示し、runCommandのHuman-in-the-loopインタラプトで一時停止したAIDAコックピット](aida-hitl-interrupt.png)

いずれかのボタンをクリックすると、一時停止されていたツールリクエストに判定結果が紐付けられ、Genkit クライアントの `chat.resumeStream(...)` が呼び出されます。Go バックエンドはコマンドの実行（またはスキップ）を完了し、更新されたゲージやコマンド出力を備えた AIDA の検証ダッシュボードをサーフェスへ直接ストリーミングで返送します。

![runCommandインタラプトを承認した後のAIDAコックピット。ディスク容量ゲージとシェルコマンド出力を備えた検証用A2UIサーフェスを描画している様子](aida-hitl-resumed.png)

## まとめ

[`go-aida`](https://github.com/danicat/go-aida) における AIDA の書き直しが完了したことで、フルスタックのアーキテクチャが完成しました。前編・後編の2回にわたる記事を通じて、私たちは Python と SQLite によるかつての RAG プロトタイプと静的なチャット画面を、単一バイナリの Genkit Go サービスとリアルタイムの A2UI 診断コックピットへと完全に刷新しました。

この開発フェーズを通じて、特に際立った設計上の教訓が2つあります。

1. **ドメイン特化型カタログは汎用 UI ブロックに勝る:** AIDA に7つのカスタム診断計器を与えたことで、ネストされたカードの中に Markdown の長文が並ぶだけだった出力は、生きたシステムダッシュボードへと生まれ変わりました。カタログはフロントエンドとバックエンドの双方を結ぶ共有コントラクトとして機能します。フロントエンドのコンポーネントを更新したら Go 側のカタログも合わせ（逆も同様）、`a2ui` ミドルウェアがそれらのスキーマをシステムプロンプトに注入して `validate: strict` を強制することで、モデルは常に有効でサポートされた計器を描画できるようになります。
2. **ツールインタラプトがホスト修復の安全性と双方向性を両立する:** Genkit のツールインタラプトを利用することで、Go バックエンドはシェルコマンドを実行する前に一時停止し、ユーザーが承認または拒否した後にクリーンに処理を再開できます。致命的なエラーではなく構造化された拒否フィードバックを返すことで、エージェントは会話の流れを壊すことなく状況に適応し、より安全な代替案を提示できます。
