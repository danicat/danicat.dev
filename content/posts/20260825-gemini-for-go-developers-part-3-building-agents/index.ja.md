---
categories:
  - Agent Development
date: 2026-08-25
heroStyle: big
series:
  - Gemini for Go Developers
series_order: 3
summary: "Go GenAI SDK、Genkit、ADKを使用してGoで自律型AIエージェントを構築する方法を、実践的なレトロゲーム鑑定士エージェントを通じて学びます。"
tags:
  - adk
  - gemini
  - genkit
  - golang
title: "Go開発者のためのGemini: Goでエージェントを構築する"
slug: "gemini-for-go-developers-part-3-building-agents"
aliases:
  - "/ja/posts/20260825-gemini-for-go-developers-part-3-building-agents/"
  - "/ja/posts/20260826-gemini-for-go-developers-part-3-building-agents/"
description: "Go開発者のためのGemini 第3部: Go GenAI SDK、Genkit、Google ADKを使ってレトロゲーム鑑定士エージェントを構築し、Cloud Runなどのランタイム環境を比較します。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "github.com/firebase/genkit/go"
  - "google.golang.org/adk/v2"
  - "google.golang.org/genai"
---

**Go開発者のためのGemini**シリーズへようこそ！[第1部: Geminiモデルファミリー]({{< ref "/posts/20260808-gemini-for-go-developers-part-1-model-family" >}})では、モデル階層ごとのGeminiの機能を確認し、[第2部: Geminiを使ったコーディング]({{< ref "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini" >}})では、Go開発向けにコーディングエージェントを設定する方法を探求しました。

今回は立場を逆転させ、GoでAI対応アプリケーションや自律型エージェントを構築する方法を探求します。本記事では、エージェントの根本的な仕組みを紐解き、具体的なユースケースである**レトロゲーム鑑定士（Retro Game Appraiser）**を定義して、Goにおける3つの異なるパラダイムで段階的に構築していきます。

1. **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)** を使って直接構築する低レベルのエージェントループ
2. **[Genkit](https://genkit.dev)** を使って構築する構造化されたフローベースのパイプライン
3. Googleの **[Agent Development Kit (ADK)](https://adk.dev)** を使って構築するモジュール型マルチエージェント・セッション対応システム

最後に、クラウド上でGoエージェントを本番運用するための堅牢なランタイム環境（実行基盤）について比較・考察します。

本記事で紹介したすべてのサンプルの完全なソースコードは、GitHubのコンパニオンリポジトリ [**danicat/gemini-for-go-developers**](https://github.com/danicat/gemini-for-go-developers/tree/main/part-3) で公開しています。

## エージェントの解剖学

「エージェント」という言葉は曖昧に使われがちですが、現代のAIエンジニアリングにおいては明確なアーキテクチャ定義が存在します。それは、**大規模言語モデル（LLM）**、1つ以上の**ツール**（実行可能な関数やAPI）、そしてフィードバックループ内で動作する**実行ハーネス（Harness）**から構成される自律型システムです。

ツールを持たない言語モデルは、単なるテキスト生成器やチャットボットに過ぎません。静的な学習時の重みやプロンプト内で渡されたコンテキストに基づいてのみ応答を生成します。モデルに「主体性（Agency）」— 外部の状態を検査し、仮説を検証し、現実世界でアクションを実行する能力 — を与えるには、実行可能なツールへの接続が不可欠です。

重要なアーキテクチャ上のポイントは、言語モデルが外部コードやAPIを直接実行することは決してないという点です。その代わり、**ローカルハーネス**（私たちが作成するGoプログラム）が仲介役として機能します。モデルが外部情報やアクションが必要だと判断すると、構造化された「関数呼び出し（Tool Call）」リクエストを発行します。ハーネスは対応するGo関数を実行して出力を取得し、その結果をモデルのターンに注入します。モデルは新しいコンテキストを評価し、さらに別のツールを呼び出すか、ユーザーへの最終回答を生成するのに十分なデータが揃ったかを判断します。

{{< mermaid >}}
sequenceDiagram
    autonumber
    actor User as ユーザー
    participant Harness as エージェントハーネス (Go)
    participant Model as Gemini LLM (Search Grounding有効)
    participant Catalog as ローカルカタログDB

    User->>Harness: プロンプト: 「SFC版マザー2の完品が$350で見つかりました。買いですか？」
    Harness->>Model: リクエスト (システムプロンプト + search_catalogツール + Google検索グラウンディング)
    Model-->>Harness: ツール呼び出し: search_catalog(query="EarthBound")
    Harness->>Catalog: ローカルインベントリを検索
    Catalog-->>Harness: 返却: owned=true, condition="カセットのみ", price_paid=$180
    Harness->>Model: ツール実行結果: {owned: true, format: "Loose", paid: 180}
    Note over Model: Geminiがサーバー側でGoogle検索グラウンディングを実行し相場を取得
    Model-->>Harness: 最終自然言語回答 (検索引用メタデータ付き)
    Harness->>User: 「すでにカセットのみを所有しています。箱・説明書付きの完品（CIB）で$350なら非常にお買い得です...」
{{< /mermaid >}}

すべてのツールがローカルクライアントでの実行を必要とするわけではありません。Geminiは、Googleのインフラ上で直接実行される**Google検索グラウンディング（Google Search grounding）**のようなネイティブ統合ツールをサポートしています。リクエスト構成で宣言すると、APIが市場価格の検索を透過的に解決し、ローカルネットワークの往復なしに応答へグラウンディングメタデータを注入します。一方、プライベートな業務データやロジックについては、スキーマ宣言、引数のディスパッチ、結果の返却をGoハーネス側が担当します。

## エージェント設計: レトロゲーム鑑定士

3つの実装手法を直接比較するため、すべてのスタックでまったく同じエージェント「**レトロゲーム鑑定士（Retro Game Appraiser）**」を構築します。

私のようにレトロゲームを収集している方なら、これは非常にあるあるな悩みです。レトロゲーム市場やフリーマーケットを訪れた際、大好きなゲームを見つけて購入し、帰宅してから「あ、すでに持ってた…」と気づくことが一度や二度ではありません。さらに、レトロゲーム市場は価格変動が激しく、状態による価格差（カセットのみ vs 箱説付き完品 / CIB vs 未開封）が大きく、海賊版（ブートレグ）の流通も後を絶ちません。

前回の買い出しの際、Geminiを使ってリアルタイムに店頭価格をダブルチェックしてみたところ、驚くほど役に立ちました。実店舗ならではの購入体験やショップ支援のために多少のプレミアムを払うのは喜ばしいことですが、誰しも偽物を掴まされたり法外なぼったくり価格を払ったりはしたくありません。

私たちの鑑定士エージェントは、個人のコレクションを照合しつつ、Web上の最新相場データを参照してこの課題を解決します。

### 機能とユーザーの対話例

コレクターは以下のような自然言語でエージェントに質問します。

* *「クロノ・トリガーは私のコレクションにありますか？」*
* *「スーパーファミコン版のMOTHER2（EarthBound）で、箱・説明書付きの極上品が$350で見つかりました。すでに持っていますか？また、現在の相場と比べてお買い得ですか？」*
* *「悪魔城ドラキュラX 月下の夜想曲はいくらで購入しましたか？市場価値は上がっていますか？」*

### ツール定義（Tool Contracts）

在庫や価格のハルシネーション（もっともらしい嘘）を防ぐため、エージェントは2つの信頼できる情報源を利用します。

1. **`search_catalog`（ローカルツール）:** コレクターのローカルデータベースを検索するGo関数。タイトルやハード名で検索し、所有状況、状態、購入日、購入価格を返します。
2. **`google_search`（検索グラウンディング）:** Web上のオークション履歴や専門店の相場を検索し、最新の平均取引価格や状態別の価格帯を取得するサーバー側ツール。

### 推論戦略

購入相談を受けたエージェントは以下の手順で推論を進めます。
1. ローカルカタログを検索し、そのゲームをすでに所有しているか（どの状態で持っているか）を確認する。
2. Google検索を実行し、指定されたハードおよび状態における現在の適正相場を調べる。
3. データを統合して実践的なアドバイスを作成する。提示価格と相場を比較し、状態アップグレード（カセットのみから完品への買い替えなど）の価値を判断して、明確な購入判断を提示します。

まずは、Go GenAI SDKを使った直接的な実装から見ていきましょう。

## Go GenAI SDKでエージェントを実装する

**[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)**（`google.golang.org/genai`）を使用したエージェント構築は、Gemini APIプロトコルに1対1で対応する**最も低い抽象度**のアプローチです。フレームワーク層が存在しないため、ツールのディスパッチループ、会話履歴の管理、終了判定をGoコードで明示的に制御します。

この低レベルな手法は、単発のスクリプト、実験的な検証、ファンクションコーリングやグラウンディングの挙動の学習、または完全に独自のループ制御を行いたい場合に最適です。依存関係のない単一のバイナリにコンパイルされるため、Cloud Run、Kubernetes、仮想マシンなど、あらゆる標準的なマイクロサービス環境に容易にデプロイできます。

以下が完全な実装コードです。

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/genai"
)

// GameItem represents a collectible item in the user's personal inventory.
type GameItem struct {
	Title     string  `json:"title"`
	Platform  string  `json:"platform"`
	Year      int     `json:"year"`
	Condition string  `json:"condition"` // e.g. "Loose Cartridge", "CIB (Complete in Box)", "Mint"
	PricePaid float64 `json:"price_paid"`
	Notes     string  `json:"notes"`
}

// localCatalog simulates an inventory database for retro games.
var localCatalog = []GameItem{
	{
		Title:     "Chrono Trigger",
		Platform:  "Super Nintendo (SNES)",
		Year:      1995,
		Condition: "CIB (Complete in Box)",
		PricePaid: 210.00,
		Notes:     "Includes original map and registration card.",
	},
	{
		Title:     "EarthBound",
		Platform:  "Super Nintendo (SNES)",
		Year:      1994,
		Condition: "Loose Cartridge",
		PricePaid: 180.00,
		Notes:     "Authentic board verified; label in excellent shape.",
	},
	{
		Title:     "Castlevania: Symphony of the Night",
		Platform:  "Sony PlayStation",
		Year:      1997,
		Condition: "CIB (Black Label)",
		PricePaid: 135.00,
		Notes:     "Original soundtrack disc included.",
	},
}

// searchCatalogTool searches the local collection for matching games.
func searchCatalogTool(args map[string]any) map[string]any {
	query, _ := args["query"].(string)
	queryLower := strings.ToLower(strings.TrimSpace(query))

	var matches []GameItem
	for _, item := range localCatalog {
		if strings.Contains(strings.ToLower(item.Title), queryLower) ||
			strings.Contains(strings.ToLower(item.Platform), queryLower) {
			matches = append(matches, item)
		}
	}

	if len(matches) == 0 {
		return map[string]any{
			"found":   false,
			"message": fmt.Sprintf("No items matching %q found in your collection.", query),
		}
	}

	return map[string]any{
		"found":   true,
		"count":   len(matches),
		"results": matches,
	}
}

func main() {
	ctx := context.Background()

	// Initialise GenAI client for Gemini Enterprise
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendEnterprise,
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// 1. Declare custom function schema for collection lookup
	catalogToolDecl := &genai.FunctionDeclaration{
		Name:        "search_catalog",
		Description: "Search the collector's personal inventory for owned games by title or platform.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "Game title or platform to search (e.g. 'EarthBound', 'SNES').",
				},
			},
			Required: []string{"query"},
		},
	}

	// 2. Configure model tools: custom function declaration + Google Search grounding
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: "You are an expert Retro Game Appraiser. When evaluating purchases, check the user's " +
					"collection catalog first to see if they already own the item, then check current market " +
					"prices using Google Search to evaluate whether the deal is fair, overpriced, or a bargain."},
			},
		},
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{catalogToolDecl},
			},
			{
				GoogleSearch: &genai.GoogleSearch{},
			},
		},
	}

	// 3. Graceful shutdown on Ctrl+C (SIGINT) or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nGoodbye!")
		os.Exit(0)
	}()

	model := "gemini-3.8-flash"
	var contents []*genai.Content

	fmt.Println("Retro Game Appraiser (SDK Agent)")
	fmt.Println("Type your question below, or 'exit' (Ctrl+C / Ctrl+D) to quit.")
	fmt.Println("-----------------------------------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nUser: ")
		if !scanner.Scan() {
			fmt.Println("\nGoodbye!")
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "exit") {
			fmt.Println("Goodbye!")
			break
		}

		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText(input)},
		})

		// 4. The Agent Loop: model generation -> tool dispatch -> feedback -> until final answer
		for {
			resp, err := client.Models.GenerateContent(ctx, model, contents, config)
			if err != nil {
				log.Printf("error generating content: %v", err)
				break
			}

			if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				log.Println("received empty response candidate from model")
				break
			}

			// Append the model's response to the conversation history
			modelContent := resp.Candidates[0].Content
			contents = append(contents, modelContent)

			// Check if the model requested any client-side tool executions
			funcCalls := resp.FunctionCalls()
			if len(funcCalls) == 0 {
				fmt.Printf("\nAppraiser: %s\n", resp.Text())
				break
			}

			// Execute each requested tool and prepare response parts
			var responseParts []*genai.Part
			for _, call := range funcCalls {
				fmt.Printf("[Harness] Executing tool: %s(args=%v)\n", call.Name, call.Args)

				var result map[string]any
				switch call.Name {
				case "search_catalog":
					result = searchCatalogTool(call.Args)
				default:
					result = map[string]any{"error": fmt.Sprintf("unsupported tool: %s", call.Name)}
				}

				responseParts = append(responseParts, genai.NewPartFromFunctionResponse(call.Name, result))
			}

			// Return tool execution results as a user turn
			contents = append(contents, &genai.Content{
				Role:  "user",
				Parts: responseParts,
			})
		}
	}
}
```

### SDKエージェントを実行する

この例を実行するには、Google Cloudプロジェクトを設定し、Application Default Credentialsでログインします（`gcloud auth application-default login`）。

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
go run main.go
```

インタラクティブなチャットセッションがターミナル上で開始されます。鑑定士エージェントと複数ターンにわたって対話でき、`exit`と入力するか、`Ctrl+C`または`Ctrl+D`（EOF）を押すことでいつでも安全に終了できます。

```text
Retro Game Appraiser (SDK Agent)
Type your question below, or 'exit' (Ctrl+C / Ctrl+D) to quit.
-----------------------------------------------------------------

User: I found a copy of EarthBound for SNES in mint Complete-in-Box (CIB) condition for $350. Do I already own it, and is $350 a good deal compared to current market prices?
[Harness] Executing tool: search_catalog(args=map[query:EarthBound])

Appraiser: Here is your collection check and appraisal for **EarthBound (SNES)**:

1. **Current Collection Status**:
   - You currently own **EarthBound** on Super Nintendo as a **Loose Cartridge**, purchased for **$180.00**.

2. **Market Price Appraisal**:
   - Verified market sales for an authentic, **Complete-in-Box (CIB)** copy of EarthBound typically range between **$1,200.00 and $1,500.00** depending on the condition of the box, tray, and original player's guide.

3. **Recommendation**:
   - At **$350.00**, a genuine Mint CIB copy is an **exceptional deal** (more than 70% below prevailing market value).
   - **Caution**: Because EarthBound is one of the most heavily counterfeited SNES titles, inspect the box printing, registration card, and PCB board carefully before completing the transaction. If verified authentic, this is an outstanding opportunity to upgrade your loose copy to CIB.

User: exit
Goodbye!
```

SDKでの実装は、制御フローの仕組みが非常に分かりやすいのが特徴です。昨今の高性能なモデルと広大なトークンウィンドウを使えば、配管やディスパッチループを自前で書くこと自体は驚くほど簡単です。

しかし、「自前で作れる」からといって「作るべき」とは限りません。自作したフレームワークコードの1行1行は、将来にわたる保守の負担となり、並行処理やシリアライズに関する潜在的なバグの温床になります。独自のエージェントフレームワークを書くことは、機能開発の観点からは余計な回り道であり、最悪の場合は本来達成すべきビジネス目標からリソースを奪う技術的負債になりかねません。「最も優れたコードは書かないコードであり、次に優れたコードは最小限の自作コードで課題を解決するコードである」という原則を思い出す価値があります。

## エージェント開発フレームワーク

生のSDKループは低レベルの仕組みを理解したり特殊なループを組むのには最適ですが、本番環境のアプリケーションではより高い抽象化が求められます。

* **スキーマの自動リフレクション:** `&genai.Schema{...}` でJSONスキーマを手動定義するのは冗長で誤りが発生しやすい作業です。フレームワークはGoの構造体やドキュメントコメントから直接スキーマを推論します。
* **可観測性と分散トレーシング:** 本番環境では、OpenTelemetryのトレース、ツールごとのレイテンシ、トークン消費量の追跡をコードに手を加えずに取得できる必要があります。
* **プロンプト管理:** プロンプトをGoコード内にハードコードすると、プロンプトエンジニアとの協業が難しくなり、バイナリのリリースとは独立したテンプレートのバージョン管理ができません。
* **セッション永続化と状態管理:** ステートレスなHTTPリクエスト間で複数ターンの会話履歴を安全に保持するには、スレッドセーフで疎結合なストレージが必要です。
* **モデルの移植性:** Go SDKはGemini専用ですが、フレームワークを使えばビジネスロジックを変更することなくモデルプロバイダを切り替えたりローカルモデルをテストしたりできます。

こうした要件に応えるため、Goエコシステムには2つの主要なオープンソースフレームワークが存在します。**Genkit** と **Agent Development Kit (ADK)** です。

| 比較軸 | Go GenAI SDK | Genkit Go | Agent Development Kit (ADK) |
| :--- | :--- | :--- | :--- |
| **抽象度** | **低**（Gemini APIへの1:1マッピング） | **中**（構造化ワークフロー） | **高**（自律型マルチエージェント） |
| **コアアーキテクチャ** | 明示的な `for` ループとディスパッチ | **Flows** (`genkit.DefineFlow`) & Tools | **Agents**, Runners & Session Services |
| **得意なユースケース** | スクリプト、仕組みの学習、特殊なループ制御 | 単発AIアプリ（CLI、Webサービス）、決定論的パイプライン、単一ドメイン | 会話型チャットエージェント、マルチエージェント連携、長期メモリ & RAG |
| **モデル対応** | Gemini専用 | マルチモデル（Google GenAI, Vertex AI, Ollamaなど） | マルチモデル（ADKモデルアダプター経由） |
| **推奨デプロイ環境** | 任意のHTTPホスト（Cloud Run, K8s, VM） | 任意のバックエンド; **Cloud Run**（推奨） | **Gemini Enterprise**（セッション/RAG推奨）または **Cloud Run** |

それでは、レトロゲーム鑑定士をそれぞれのフレームワークで実装してみましょう。

## Genkitでエージェントを実装する

**[Genkit](https://genkit.dev)** は**中レベルの抽象度**に位置し、AIアプリケーションにソフトウェアエンジニアリングの規律と組み込みの可観測性をもたらします。Genkitでは、すべてが **Flows**（型安全で観測可能なパイプライン）と **Tools**（自動スキーマ生成を備えたGo関数）を中心に構成されます。

Genkitは、**単発の処理**（CLIツール、バッチ処理、Webhookなど）、決定論的な処理パイプライン、単一ドメインに特化したエージェントに最適です。プラグインを通じて複数のモデルをサポートし、標準的なGo HTTPサーバーとして動作するため、コンテナ管理と自動スケーリングに優れた **Google Cloud Run** が最適なデプロイ先となります。

以下はGenkit Goによる実装です。

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"google.golang.org/genai"
)

// GameItem represents a collectible item in the user's personal inventory.
type GameItem struct {
	Title     string  `json:"title"`
	Platform  string  `json:"platform"`
	Year      int     `json:"year"`
	Condition string  `json:"condition"`
	PricePaid float64 `json:"price_paid"`
	Notes     string  `json:"notes"`
}

// localCatalog simulates an inventory database for retro games.
var localCatalog = []GameItem{
	{
		Title:     "Chrono Trigger",
		Platform:  "Super Nintendo (SNES)",
		Year:      1995,
		Condition: "CIB (Complete in Box)",
		PricePaid: 210.00,
		Notes:     "Includes original map and registration card.",
	},
	{
		Title:     "EarthBound",
		Platform:  "Super Nintendo (SNES)",
		Year:      1994,
		Condition: "Loose Cartridge",
		PricePaid: 180.00,
		Notes:     "Authentic board verified; label in excellent shape.",
	},
	{
		Title:     "Castlevania: Symphony of the Night",
		Platform:  "Sony PlayStation",
		Year:      1997,
		Condition: "CIB (Black Label)",
		PricePaid: 135.00,
		Notes:     "Original soundtrack disc included.",
	},
}

type CatalogRequest struct {
	Query string `json:"query" jsonschema:"description=The game title or platform to search in the inventory"`
}

type CatalogResponse struct {
	Found   bool       `json:"found"`
	Message string     `json:"message,omitempty"`
	Count   int        `json:"count,omitempty"`
	Results []GameItem `json:"results,omitempty"`
}

type AppraiserRequest struct {
	Prompt string `json:"prompt" jsonschema:"description=The collector's question or purchase offer to evaluate"`
}

type AppraiserResponse struct {
	Appraisal string `json:"appraisal"`
}

func main() {
	ctx := context.Background()

	// 1. Initialise Genkit with Vertex AI plugin
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.VertexAI{
			ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
			Location:  "global",
		}),
	)

	// 2. Define strongly-typed tool with automatic schema generation
	catalogTool := genkit.DefineTool(
		g,
		"search_catalog",
		"Search the collector's personal inventory for owned games by title or platform.",
		func(ctx *ai.ToolContext, req CatalogRequest) (CatalogResponse, error) {
			queryLower := strings.ToLower(strings.TrimSpace(req.Query))
			queryWords := strings.Fields(queryLower)
			var matches []GameItem

			for _, item := range localCatalog {
				itemText := strings.ToLower(item.Title + " " + item.Platform)
				allMatch := true
				for _, word := range queryWords {
					if !strings.Contains(itemText, word) {
						allMatch = false
						break
					}
				}
				if allMatch {
					matches = append(matches, item)
				}
			}

			if len(matches) == 0 {
				return CatalogResponse{
					Found:   false,
					Message: fmt.Sprintf("No items matching %q found in personal collection.", req.Query),
				}, nil
			}

			return CatalogResponse{
				Found:   true,
				Count:   len(matches),
				Results: matches,
			}, nil
		},
	)

	// 3. Define structured appraisal flow with typed request and response
	appraiserFlow := genkit.DefineFlow(
		g,
		"appraise_game",
		func(ctx context.Context, req AppraiserRequest) (AppraiserResponse, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithModelName("vertexai/gemini-3.8-flash"),
				ai.WithSystem(
					"You are an expert Retro Game Appraiser. Assist collectors by evaluating prospective purchases, "+
						"cross-referencing their personal inventory, and assessing fair market valuations. "+
						"Always search the collection catalog using search_catalog before providing purchase recommendations.",
				),
				ai.WithConfig(&genai.GenerateContentConfig{
					ThinkingConfig: &genai.ThinkingConfig{IncludeThoughts: true},
					Tools: []*genai.Tool{
						{
							GoogleSearch: &genai.GoogleSearch{},
						},
					},
				}),
				ai.WithPrompt(req.Prompt),
				ai.WithTools(catalogTool),
			)
			if err != nil {
				return AppraiserResponse{}, fmt.Errorf("appraisal generation failed: %w", err)
			}
			return AppraiserResponse{Appraisal: resp.Text()}, nil
		},
	)

	// 4. Mount flow directly using Genkit's built-in HTTP handler
	mux := http.NewServeMux()
	mux.Handle("POST /api/appraise", genkit.Handler(appraiserFlow))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown on Ctrl+C (SIGINT) or SIGTERM
	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Retro Game Appraiser (Genkit) listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-serverCtx.Done()
	log.Println("\nShutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly.")
}
```

### Genkitフローを実行する

Google Cloudプロジェクト環境変数を設定してGenkitサーバーを起動します。

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
export PORT=8080
go run main.go
```

別のターミナルから鑑定リクエストを送信します。

```sh
curl -s -X POST http://localhost:8080/api/appraise \
  -H "Content-Type: application/json" \
  -d '{"data": {"prompt": "I found a copy of EarthBound for SNES for $350. Do I own it, and is it a good deal?"}}' | jq .
```

フローが実行され、カタログツールを呼び出して構造化された鑑定結果が返されます。

```json
{
  "appraisal": "### 1. Catalog Check\n**Yes, you already own it.**\n* **Title:** *EarthBound* (SNES, 1994)\n* **Status in Collection:** Loose Cartridge\n* **Condition/Notes:** Authentic board verified; label in excellent shape.\n* **Price Paid:** $180\n\n---\n\n### 2. Market Appraisal & Deal Analysis\n* **Loose Cartridge:** The current going market rate for an authentic loose copy ranges between **$320 and $380**. At **$350**, it is priced right at **fair market value**—neither an overpriced listing nor a significant bargain.\n* **Complete in Box (CIB) / Boxed with Guide:** If this listing happens to include the original big box and strategy guide with scratch-and-sniff cards, $350 would be an extraordinary steal (CIB copies regularly sell for **$1,500–$2,500+**).\n\n---\n\n### 3. Recommendation\n* **Pass (if Loose):** Since you already have an authentic copy in excellent condition, paying retail market price ($350) for a duplicate loose cart does not offer strong value or upside.\n* **Buy immediately (if Complete/Boxed):** Only pull the trigger if it includes the original packaging or represents a major condition upgrade/variant.\n* **Buyer Beware:** If you do ever consider another copy, always inspect the PCB (printed circuit board) screws and chips, as *EarthBound* is one of the most frequently counterfeited games on the SNES."
}
```

Genkitを使うことで、手動のディスパッチループを書く必要がなくなります。引数をGoの構造体にアンマーシャリングし、関数を実行して結果をモデルに戻し、各ステップのテレメトリトレースを自動的に収集してくれます。

## Agent Development Kit (ADK) でエージェントを実装する

Genkitが構造化されたアプリケーションパイプラインに主眼を置いているのに対し、Googleの **[Agent Development Kit (ADK)](https://adk.dev)** は、自律型会話エージェント、マルチエージェントのオーケストレーション、永続セッションやエンタープライズRAGを必要とする複雑なシステム向けに設計された**高レベルの抽象化**を提供します。

ADKは、エージェントのライフサイクル、サブエージェントへのタスク委譲、エージェント間通信プロトコル（A2A）を標準化しています。Genkitと同様に、モジュール式のモデルアダプターによりマルチモデルに対応しています。

実行基盤としては、自前でデータベースを組むことなくマネージドなセッション永続化やエンタープライズグラウンディングを利用したい場合は **Gemini Enterprise Agent Platform** が最適です。コンテナ化されたステートレスマイクロサービスとして独自に管理したい場合は **Google Cloud Run** が適しています。

以下はADK v2による実装です。

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/cmd/launcher"
	"google.golang.org/adk/v2/cmd/launcher/full"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
	"google.golang.org/adk/v2/tool/geminitool"
)

// GameItem represents a collectible item in the user's personal inventory.
type GameItem struct {
	Title     string  `json:"title"`
	Platform  string  `json:"platform"`
	Year      int     `json:"year"`
	Condition string  `json:"condition"`
	PricePaid float64 `json:"price_paid"`
	Notes     string  `json:"notes"`
}

// localCatalog simulates an inventory database for retro games.
var localCatalog = []GameItem{
	{
		Title:     "Chrono Trigger",
		Platform:  "Super Nintendo (SNES)",
		Year:      1995,
		Condition: "CIB (Complete in Box)",
		PricePaid: 210.00,
		Notes:     "Includes original map and registration card.",
	},
	{
		Title:     "EarthBound",
		Platform:  "Super Nintendo (SNES)",
		Year:      1994,
		Condition: "Loose Cartridge",
		PricePaid: 180.00,
		Notes:     "Authentic board verified; label in excellent shape.",
	},
	{
		Title:     "Castlevania: Symphony of the Night",
		Platform:  "Sony PlayStation",
		Year:      1997,
		Condition: "CIB (Black Label)",
		PricePaid: 135.00,
		Notes:     "Original soundtrack disc included.",
	},
}

type CatalogRequest struct {
	Query string `json:"query" jsonschema:"The game title or platform to search in the inventory."`
}

type CatalogResponse struct {
	Found   bool       `json:"found"`
	Message string     `json:"message,omitempty"`
	Count   int        `json:"count,omitempty"`
	Results []GameItem `json:"results,omitempty"`
}

func main() {
	ctx := context.Background()

	// 1. Initialise Gemini Model adapter for Gemini Enterprise
	model, err := gemini.NewModel(ctx, "gemini-3.8-flash", &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendEnterprise,
	})
	if err != nil {
		log.Fatalf("failed to create Gemini model: %v", err)
	}

	// 2. Wrap collection lookup as an ADK Function Tool
	catalogTool, err := functiontool.New(functiontool.Config{
		Name:        "search_catalog",
		Description: "Search the collector's personal inventory for owned games by title or platform.",
	}, func(ctx agent.Context, req CatalogRequest) (CatalogResponse, error) {
		queryLower := strings.ToLower(strings.TrimSpace(req.Query))
		var matches []GameItem

		for _, item := range localCatalog {
			if strings.Contains(strings.ToLower(item.Title), queryLower) ||
				strings.Contains(strings.ToLower(item.Platform), queryLower) {
				matches = append(matches, item)
			}
		}

		if len(matches) == 0 {
			return CatalogResponse{
				Found:   false,
				Message: fmt.Sprintf("No items matching %q found in personal collection.", req.Query),
			}, nil
		}

		return CatalogResponse{
			Found:   true,
			Count:   len(matches),
			Results: matches,
		}, nil
	})
	if err != nil {
		log.Fatalf("failed to create catalog tool: %v", err)
	}

	// 3. Define autonomous LLM Agent
	appraiserAgent, err := llmagent.New(llmagent.Config{
		Name:        "retro_game_appraiser",
		Model:       model,
		Description: "Expert appraiser that analyzes retro video game purchases and collection inventory.",
		Instruction: "You are an expert Retro Game Appraiser. Assist collectors by verifying collection " +
			"status with search_catalog, assessing condition variants, and offering objective buying recommendations.",
		Tools: []tool.Tool{
			catalogTool,
			geminitool.GoogleSearch{},
		},
	})
	if err != nil {
		log.Fatalf("failed to create appraiser agent: %v", err)
	}

	// 4. Configure launcher and execute
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(appraiserAgent),
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
```

### ADKエージェントを実行する

ADKにはユニバーサルランチャー（`full.NewLauncher()`）が用意されており、HTTPルーティングやJSONのマーシャリング、セッション管理の定型コードを手動で書く必要がありません。コマンドライン引数をランチャーに渡すだけで、アプリケーションコードを1行も変更することなく、複数の対話モードでエージェントを実行できます。

ターミナルで対話型チャットセッションを開始するには、以下を実行します。

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
go run main.go
```

ランチャーがコンソールモードで起動し、ターミナル上で直接鑑定エージェントと会話できます。

```text
User: Do I have Chrono Trigger in my collection?
Agent: Yes, you have Chrono Trigger in your collection! Here are the details from your inventory:

* Title: Chrono Trigger
* Platform: Super Nintendo (SNES)
* Release Year: 1995
* Condition: CIB (Complete in Box)
* Price Paid: $210.00
* Notes: Includes original map and registration card.

User: What did I pay for it?
Agent: You paid $210.00 for it.
```

ADKが会話状態とセッションの継続性を自動的に管理するため、過去のやり取りのコンテキストを維持したまま自然に追加の質問を行えます。

また、ADKに組み込まれた開発者向けWeb UIとREST APIサーバーを起動することも可能です。

```sh
go run main.go web webui api
```

ブラウザで `http://localhost:8080` にアクセスすると対話型チャット画面が開き、リアルタイムのストリーミング応答、セッション履歴、およびツール実行（カスタムの `search_catalog` 関数ツールと組み込みの `GoogleSearch` によるグラウンディングの両方）の詳細な実行ステップを確認できます。

ADKは、エージェント定義（`llmagent`）、ツール（`functiontool` や組み込みの `geminitool.GoogleSearch`）、実行環境（`launcher` と `runner`）の間で明確な責務の分離を提供し、複雑なマルチエージェント階層や会話アシスタントに最適な選択肢となります。

## エージェント実行ランタイム環境

Goでエージェントコードを記述することは行程の半分に過ぎません。完成したエージェントを本番運用するには、長時間のストリーミング応答、バックグラウンドでのツール実行、安全な認証情報の管理、急なトラフィック変動に耐えうる実行基盤が必要です。

Goは最小限のリソース消費で単一バイナリにコンパイルされるため、極めて高速かつ安価にホスティングできます。フレームワークや状態管理の要件に応じて、主に2つのデプロイ先が存在します。

### Cloud Run: コンテナ化バックエンドの最適解

**[Google Cloud Run](https://cloud.google.com/run)** は、SDK直接実装、Genkitフロー、ADK Runnerのいずれであっても、Goエージェントをコンテナとしてホスティングする最適な環境です。

Dockerのマルチステージビルドを活用することで、Goエージェントを実行時依存関係のない軽量な `scratch` や `distroless` イメージにコンパイルできます。重いインタプリタや仮想マシンを起動する必要がないため、ゼロからアクティブインスタンスへと素早くスケールし、重量級ランタイムで発生しがちなコールドスタートのオーバーヘッドを最小限に抑えられます。

Cloud Runの主なメリット:

* **真のゼロスケールと高並行性:** エージェントの実行中やツール応答の処理中など、実際にCPUを消費したミリ秒単位でのみ課金されます。Goの軽量なゴルーチンにより、1つのインスタンスで極小のメモリフットプリントのまま数百の同時ターンを処理できます。
* **最大60分の延長タイムアウト:** デフォルトのタイムアウトは5分ですが、Cloud Runは最大60分（3,600秒）まで設定可能です。多段階の推論、広範なディープリサーチ、サブエージェント群の連携処理であっても、途中で中断されることなく完了できます。
* **双方向ストリーミングとWebSockets:** HTTP/2のチャンク転送、Server-Sent Events（SSE）、WebSocketsをネイティブサポートしています。Gemini Live APIを用いた音声やリアルタイムのマルチモーダル対話においても、シームレスな通信が可能です。
* **セッションアフィニティ（Sticky Sessions）:** エージェントがメモリ上に一時キャッシュを保持する場合、クライアントIPやCookieによるセッションアフィニティ（`--session-affinity`）を有効にして、同一インスタンスへリクエストをルーティングできます。
* **疎結合な耐障害性:** 本番環境では、Firestore、Redis、Cloud SQLなどのマネージドDBに状態を永続化すべきです。長時間のターン中に接続が切断された場合でも、インタラクションIDを使ってコンテキストを損なうことなく再開できます。
* **Workload Identity IAMによるセキュアな認証:** コード内にAPIキーをハードコードする必要はありません。GoエージェントはCloud Runのサービスアカウント環境を利用して、Gemini Developer API、Vertex AI、Cloud Storageへ安全に認証されます。

### Gemini Enterprise Agent Platform: マネージドセッションとエンタープライズRAG

**Agent Development Kit (ADK)** を使用して、長期にわたるセッション、永続メモリ、社内データへのグラウンディングを必要とするエンタープライズソリューションを構築する場合、**Gemini Enterprise Agent Platform**（Vertex AI Agent Engineの進化版）がフルマネージドなサーバーレス基盤を提供します。

独自のデータベースを構築して永続化アダプターを開発する代わりに、以下の機能が標準で提供されます。

* **疎結合なセッション永続化:** 長期的な会話履歴ストレージ（ADKの `SessionService`）とエフェメラルなストリーミング実行ループ（再接続時に自動再開可能な `LiveSession`）の明確な分離。
* **エンタープライズグラウンディングとベクトル検索:** 社内ナレッジベース（Google Drive、BigQuery、社内リポジトリ）へのネイティブコネクタと、高スループットなセマンティック検索を実現するVertex AI Vector Search（ストレージ最適化ティア対応）。
* **安全なサンドボックスコード実行:** エージェントが安全にコードを動的生成・実行できる隔離された環境（データ分析、Python/Goスクリプトなど）。ホストインフラを危険に晒すことはありません。
* **Agent-to-Agent（A2A）プロトコル:** 組織内の異なる領域に存在する独立したエージェント同士が、機能を検出し、スキーマを合意し、タスクを相互に委譲できる標準化されたプロトコル。
* **セキュリティ、エージェントID、Model Armor:** きめ細やかなIAM権限、VPC Service Controls境界、およびプロンプトインジェクションやデータ漏洩、ポリシー違反を検知・防御するModel Armorランタイム保護。

## 次回予告

1つの記事ですべてのフレームワークやランタイム環境を網羅し切ることは困難ですが、ご安心ください。本シリーズの今後の記事でそれぞれを徹底的に深掘りしていきます。

* **第4部**: **Genkit for Go** のディープダイブ — dotpromptテンプレート、カスタムプラグイン、ストリーミング、Dev UIによる可観測性。
* **第5部**: **Agent Development Kit (ADK)** のディープダイブ — 自律型マルチエージェント階層の構築、サブエージェントへの委譲、セッション状態管理。
* **第6部**: 本番環境レベルのCI/CDとIAMを備えた、Goエージェントの **Cloud Run** および **Gemini Enterprise Agent Platform** へのデプロイ。
* **第7部**: **[Ebitengine](https://ebitengine.org/)** を使用したGoによるゲーム開発への展開。

Stay tuned, and happy hacking!
