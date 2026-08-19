---categories:
- Applied GenAI
date: 2026-08-08
heroStyle: big
series:
- Gemini for Go Developers
series_order: 1
summary: 「Go開発者のためのGemini」シリーズ第1部。Geminiモデルファミリーの各種モデルと、それらを呼び出すためのAPIサーフェスを詳しく解説します。
tags:
  - gemini
  - golang
title: "Go開発者のためのGemini：モデルファミリー"
slug: "gemini-for-go-developers-part-1-model-family"
aliases:
  - "/ja/posts/20260808-gemini-for-go-developers-part-1-model-family/"
description: "Go開発者のためのGemini第1部：Gemini 3.x、Flash、Pro、Nano Bananaの比較から各種API、公式Go GenAI SDKでの実装までを徹底解説。"
proficiencyLevel: "Beginner"
dependencies:
  - "Go 1.24+"
  - "google.golang.org/genai"
---

**Go開発者のためのGemini** へようこそ！このシリーズは、GoでAI搭載ソフトウェアを構築するための完全ガイドです。全7章の実践的なチュートリアルを通じて、エージェンティック・コーディングから、**Genkit** や **ADK** を使用した自律型エージェントの構築、ゲーム開発、そして **G3スタック**（Go、Gemini、GCP）を活用したクラウドへのデプロイまでを網羅します。

第1章では、Geminiモデルファミリーの構成やモデル設定を把握し、公式の [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) を使って最初のGoコードを作成しながら基礎を固めていきます。

## Geminiモデルファミリー

「Google」を検索の代名詞として使うのと同じように、私たちは「Gemini」をGoogleのAI技術全般を指す単一の名前として扱いがちです。しかし実際には、Geminiは運用上の様々なトレードオフ（速度、コスト、推論能力）に合わせて設計された、複数のモデルからなるファミリーです。

見出しを飾るのは最先端のフロンティアモデルですが、費用対効果の高いエンジニアリングを実現するには、小型モデルや特化型モデルをいつ活用すべきかを把握することが欠かせません。モデルの階層（Tier）によってレイテンシは劇的に変化するため、モデルの選択は[ユーザー体験](https://services.google.com/fh/files/blogs/google_delayexp.pdf)やプロダクトの定着率にも直結します。

あらゆるタスクに対して思考レベル（thinking levels）を高く設定したGemini Proモデルを使いたくなるかもしれませんが、それが常に最適なアプローチとは限りません。多くのユースケースでは、成果物の品質が向上しないまま、リクエストごとのレイテンシとAPIコストが無駄に跳ね上がるだけです。

### モデルの命名規則

Geminiカタログを把握するには、Googleがモデルにどのような命名を行っているかを理解しておくと便利です。標準的なモデル識別文字列は次のパターンに従います：

{{< katex >}}
\[
\text{[family]}-\text{[version]}-\text{[tier]}{-\text{[modifier]}}
\]

例：`gemini-3.6-flash` や `gemini-3-pro-image`

* **ファミリー（Family）**: ほとんどのモデルはGeminiファミリーに属しますが、GoogleにはVeoやLyriaといった他のモデルファミリーも存在します。
* **バージョン番号（Version Numbers）**: 知能、コンテキストウィンドウの処理能力、指示への追従性における世代的な飛躍を表します。
* **モデル階層（Model Tiers）**:
  * **Pro**: 複雑なマルチステップ推論向けに設計されたモデル。
  * **Flash**: 速度を重視したバランス型モデル。
  * **Flash-Lite**: 速度と高スループットが求められるシンプルなタスク向けに最適化されたモデル。
* **修飾子（Modifiers）**: `gemini-3.1-flash-image` の `image` や `gemini-3.1-flash-live-preview` の `live` のように、サブファミリーや特化領域を示します。また、`-preview` や `-exp`（experimental / 実験的）などのライフサイクル修飾子が含まれることもあります。

### モデル概要

フロンティアモデルであるGemini 3.xをはじめとする、主要なGeminiモデルの概要は以下の通りです：

#### Gemini 3.x

Gemini 3.xは主力となるフロンティアモデルラインであり、Pro、Flash、Flash-Liteの各階層で提供されています。これらの汎用モデルは、コード生成やソフトウェアエンジニアリングタスクにおける第一選択肢でもあります。

現在の主なモデル：
- `gemini-3.6-flash`: マルチモーダル推論やエージェンティックタスク向けの高速主力モデル
- `gemini-3.5-flash-lite`: 高スループットなマイクロサービス向けの最安・超高速モデル
- `gemini-3.1-pro-preview`: 複雑なマルチステップ推論や深層コードベース分析向けの高度なモデル

#### Gemini画像モデル（Nano Banana）

技術的にはGeminiファミリーの一部ですが、画像生成に特化したモデルであり、マルチモーダルな入力と出力（画像およびテキスト）の両方をサポートしています。ゼロからの画像生成だけでなく、既存画像の編集も行えます。

現在の主なモデル：
- `gemini-2.5-flash-image`（別名 Nano Banana）
- `gemini-3-pro-image`（別名 Nano Banana Pro）
- `gemini-3.1-flash-image`（別名 Nano Banana 2）
- `gemini-3.1-flash-lite-image`（別名 Nano Banana 2 Lite）

#### Veo

[ネイティブ音声付き動画生成](https://ai.google.dev/gemini-api/docs/veo)に特化したモデルです。テキストプロンプトや、トランジションを指定するキー画像（開始フレームや終了フレーム）、参照画像に基づいて動画を生成します。Veo 3.1は最大8秒のクリップを生成し、7秒刻みで最大20回まで動画を延長（拡張）できます。

現在の主なモデル：
- `veo-3.1-generate-preview`
- `veo-3.1-lite-generate-preview`（高速生成）

#### Lyria

[Lyria](https://ai.google.dev/gemini-api/docs/music-generation) は音楽生成に特化しており、インストゥルメンタルからボーカル楽曲まで生成できます。テキストと画像の両方を入力として受け付け、画像を着想（インスピレーション）として作曲に反映させることも可能です。歌詞を自ら指定することも、モデルに自動生成させることもできます。

現在の主なモデル：
- `lyria-3-pro-preview`
- `lyria-3-clip-preview`（30秒のショートクリップ）

#### Gemma

[Gemma](https://ai.google.dev/gemma/docs) は、Googleが提供するオープンウェイト（open-weights）モデルファミリーです。Geminiの基盤技術と同じ技術でトレーニングされていますが、自前のインフラにデプロイして運用できるように設計されています。Google公式のモデルに加えて、Gemmaには多種多様なユースケース向けにファインチューニングモデルを開発する[強力なコミュニティ](https://deepmind.google/models/gemma/gemmaverse/)が存在します。

一部のGemmaモデルはローカルマシン上で実行できるほど軽量であり、ネットワーク接続が限られている環境やオフライン環境でのユースケースを可能にします。一方、より大規模なモデルは極めて高い能力を備えており、データ主権（sovereignty）や完全なネットワーク分離が求められる現場で活躍します。

#### その他の注目モデル

- ライブモデル（Live models）: バッチ処理や単発のリクエスト/レスポンス処理を行うモデルに加え、Googleはリアルタイムストリーミング用のライブモデルも提供しています。モデル名に `-live` が含まれます（例：`gemini-3.1-flash-live-preview`）。
- Text-to-speech（音声合成）: 音声タグを用いてナレーションを制御し、テキストから自然な音声を生成します（`gemini-3.1-flash-tts-preview`）。
- Computer use（コンピュータ操作）: 画面を「見て」、ブラウザタスクなどを自動化できるモデル（`gemini-2.5-computer-use-preview-10-2025`）。

このように、Geminiは単一のモデルではなく、シンプルなチャットボットからマルチモーダル生成、高度なエージェンティック機能までを網羅する包括的なスイートです。

各モデルの詳細な仕様については、[公式Geminiモデルドキュメント](https://ai.google.dev/gemini-api/docs/models) を参照してください。

## 追加機能

標準的なテキスト補完機能にとどまらず、Geminiモデルは複雑なアプリケーションを構築するための高度な機能を備えています。ここでは特に重要な機能をいくつか紹介します。

### 思考（Thinking）

Gemini 2.5以降のモデルには、マルチステップの計画立案、論理的思考、コーディング、および数学的処理能力を大幅に向上させる内部推論プロセスが導入されています。最終的なレスポンスを出力する前に、モデルは内部で「思考トークン（thinking tokens）」を生成し、エッジケースの分析や複数ステップの戦略立案を行います。

思考機能は、Gemini 2.5では `thinking budget`、Gemini 3.xでは `thinking level` という設定パラメータで制御できます。バジェットや思考レベルを高く設定するほど、モデルは推論フェーズにより多くの時間とトークンを費やします。

思考機能が有効な場合、課金対象となる出力トークン数には、最終的に生成された出力テキストだけでなく、モデルが内部生成した思考トークンも含まれます。タスクの複雑さに応じて思考レベルを適切に調整することが、本番運用におけるコスト最適化の鍵となります。

### 組み込みツールとファンクションコーリング（Function Calling）

ファンクションコーリングを利用すると、Geminiモデルを外部ツール、API、データベースと連携させることができます。Geminiは、組み込みツール（`google_search` や `code_execution` など）と、アプリケーション側で定義したカスタム関数の両方をサポートしています。

ファンクションコーリングには主に3つのユースケースがあります：
- **アクションの実行（Take Actions）:** API経由で外部システムと連携し、会議のスケジュール登録、メール送信、請求書作成、スマートホーム機器の操作などを実行します。
- **知識の拡張（Augment Knowledge）:** 外部データベース、マイクロサービス、プライベートなナレッジベースからリアルタイム情報や独自データを取得します。
- **機能の拡張（Extend Capabilities）:** LLM単体では苦手な高精度の数学計算、データ変換、グラフ生成などを実行します。

#### ファンクションコーリングの仕組み

ファンクションコーリングは、アプリケーションとモデルの間で次の4つのステップで実行されます：

1. **ツールの宣言（Declare tools）**: 関数の宣言（関数名、明確な説明文、パラメータのJSON Schema）を定義し、リクエスト設定に渡します。
2. **モデルによるツールの呼び出し判定（Model identifies tool intent）**: モデルはプロンプトとツール宣言を解析します。ツールの使用が必要と判断されると、関数名と引数を含む構造化されたツール呼び出しインテントを返します。
3. **関数コードの実行（Execute function code）**: モデル自身がコードを実行するわけでは*ありません*。アプリケーション側が関数呼び出しリクエストを受け取り、対応するローカルロジックを実行して結果を取得します。
4. **関数の実行結果を返却（Return function result）**: 実行結果を関数レスポンスとしてモデルに送り返します。モデルはこのデータをもとに最終的な自然言語レスポンスを生成するか、さらなるツール呼び出しが必要かを判断します。

### 構造化出力（Structured Outputs）

指定した [JSON Schema](https://ai.google.dev/gemini-api/docs/structured-output) に厳密に準拠したレスポンスを生成するようにGeminiモデルを設定できます。これにより、テキストからの構造化データ抽出が容易になり、モデルの出力をデータ構造にマッピングする際の壊れやすいパース処理が不要になります。

RESTペイロードに生のJSON Schemaを直接記述する以外にも、Google GenAI SDKを使えば、Pythonの [Pydantic](https://docs.pydantic.dev/) やGoの構造体タグ（struct tags）など、プログラミング言語のネイティブな構造定義を用いてスキーマを定義できます。

## プログラムからモデルを呼び出す

モデルのエコシステムと主要機能を押さえたところで、これらのAPIをGoアプリケーションに組み込む方法を見ていきましょう。

用途に応じて異なるモデルが存在するように、APIサーフェスにもいくつかの選択肢が用意されています。まずは最も基本的な「Generate Content」から見ていきます。

### Generate Content API

これは最も基本的な[コンテンツ生成API](https://ai.google.dev/api/generate-content#method:-models.generatecontent)です。単一のリクエストを受け取ってレスポンスを返すステートレスなインターフェースです。複数ターン（マルチターン）の対話を行う場合、アプリケーション側でリクエストごとに完全な会話履歴を送信する必要があります。

そのため、コンテキストウィンドウの上限を超えないよう、アプリケーション側で会話履歴を適切に管理する必要があります。通常、履歴が一定のしきい値に達した段階で要約処理を行います。また、長時間のセッションにおける入力コストを抑えるため、Gemini APIではGemini 2.5以降の全モデルで[暗黙的キャッシュ（implicit caching）](https://ai.google.dev/gemini-api/docs/caching)がサポートされているほか、大きなペイロード向けに[明示的キャッシュ（explicit caching）](https://ai.google.dev/gemini-api/docs/generate-content/caching)も利用できます。

Generate Content APIはシンプルなステートレス生成に適していますが、より高度で柔軟なInteractions APIへの移行が段階的に進んでいます。

### Interactions API

> 注：現時点において、Interactions APIは公式のGo GenAI SDKではまだサポートされていません。実装状況は公式の [GitHub issue](https://github.com/googleapis/go-genai/issues/658) でトラッキングされています。

[Interactions API](https://ai.google.dev/gemini-api/docs/interactions-overview) は、シンプルなチャットやツール利用から複雑なエージェンティックワークフローまで、あらゆるタスクに対応するGoogleの統合インターフェースです。会話履歴をサーバー側で管理できるため、アプリケーション側で履歴管理ロジックを実装する必要がなくなります。

### Live API

[Live API](https://ai.google.dev/gemini-api/docs/live-api) は、WebSocket経由でリアルタイムの双方向音声・映像対話を可能にします。ユーザーの発話や割り込みを自動検知し、ライブセッション内でWeb検索やファンクションコーリングなどのツールをシームレスに使いながら、極めて自然な対話体験を実現します。

### Batch API

[Batch API](https://ai.google.dev/gemini-api/docs/batch-api) を使うと、大量のデータを半額（50%オフ）のコストで非同期処理できます。ジョブはオフピーク時間帯にバックグラウンドで実行され（通常24時間以内に完了）、リアルタイム性を求めないワークロードに最適です。

### Managed Agents API

[マネージドエージェント（Managed agents）](https://ai.google.dev/gemini-api/docs/agents) は、AIエージェントが自律的にタスクを計画・実行するための完全マネージドな実行環境を提供します。単一のAPI呼び出しを行うだけで、PythonやNodeなどのランタイムがプリインストールされたOS隔離のLinuxサンドボックスがプロビジョニングされ、エージェントがコードの実行、ファイル操作、Webブラウジングなどを自律的に行えるようになります。

Googleは、すぐに利用できる2つの事前構築済みマネージドエージェントを提供しています：
- **Antigravity Agent** (`antigravity-preview-05-2026`): コード実行、ファイル管理、Webアクセスをこなすデフォルトの汎用エージェント。Gemini 3.6 Flash駆動（Gemini 3.5 FlashまたはFlash-Liteにも変更可能）。
- **Deep Research Agent** (`deep-research-preview-04-2026`): 複数のWebソースを横断調査し、詳細なリサーチレポートをバックグラウンドで自動作成する自律型リサーチエージェント。

さらに、システムルールをインラインで定義したり、`AGENTS.md` ファイルをマウントしたり、構造化されたスキルディレクトリ（`SKILL.md`）を読み込ませたり、ローカルファイル、Cloud Storageバケット、Gitリポジトリをリモートワークスペース（`/workspace`）に直接マウントすることで、Antigravityエージェントの機能を柔軟に拡張できます。

## アクセスと請求

Geminiを組み込む際、Googleは要件に応じて2つの主要なアクセスおよび請求モードを用意しています：

1. **Google AI Studio (Google AI)**: APIキー（`GEMINI_API_KEY` または `GOOGLE_API_KEY`）を使用してGemini API経由でリクエストをルーティングします。プロトタイピング、個人開発プロジェクト、インディーアプリ、素早い開発着手に最適です。
2. **Gemini Enterprise（旧 Vertex AI）**: Google Cloud IAM、Application Default Credentials (ADC)、サービスアカウントキー、またはOAuth 2.0ユーザートークンを用いて、Google Cloudエンドポイント経由でリクエストをルーティングします。厳格なデータプライバシー、セキュリティコンプライアンス、SLA、GCPリソース管理、確約利用割引が求められるエンタープライズの本番運用に最適です。

## Go GenAI SDK

それでは、実際のGoコードでの動作を見ていきましょう。

GoアプリケーションにGeminiを統合するための公式SDKは [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai) です。

すべてのGoogleモデルに対応し、Google AIとGemini Enterpriseの両方の認証方式をサポートするよう設計されているため、「統合SDK（unified SDK）」と呼ばれることもあります。すでに非推奨（deprecated）となった旧パッケージ `github.com/google/generative-ai-go` を置き換えるものです。

`go get` でインストールします：

```bash
go get google.golang.org/genai
```

Gemini Enterpriseの認証と課金を利用する実装例は以下の通りです：

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	// Initialize the client for Gemini Enterprise (Vertex AI)
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global", // Nano Banana models are served globally
		Backend:  genai.BackendEnterprise,
	})
	if err != nil {
		log.Fatalf("failed to create GenAI client: %v", err)
	}

	prompt := "Generate a high-resolution, cute image of a fluffy cat wearing a tiny wizard hat."

	// Call Nano Banana 2 Lite (gemini-3.1-flash-lite-image)
	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-3.1-flash-lite-image",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		log.Fatalf("failed to generate image: %v", err)
	}

	// Extract generated image bytes from response parts
	for _, candidate := range resp.Candidates {
		if candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != nil {
				filename := "cute_cat.png"
				if err := os.WriteFile(filename, part.InlineData.Data, 0644); err != nil {
					log.Fatalf("failed to save image: %v", err)
				}
				fmt.Printf("Successfully generated and saved cat picture to %s!\n", filename)
				return
			}
		}
	}

	log.Fatal("no image data returned in response")
}
```

このプログラムは以下のコマンドで実行できます：

```sh
export GOOGLE_CLOUD_PROJECT="your-project-id-goes-here"
go run main.go
```

実行結果は次のようになります：

![Goターミナルで生成された魔法使いの猫の画像出力](image.png "AIの真の目的：無限の猫画像生成")

これはSDKの基本的な使い方を示すシンプルな例ですが、本シリーズを通じて、Go GenAI SDKと [Genkit](https://genkit.dev/) や [Agent Development Kit (ADK)](https://adk.dev/) などの上位フレームワークの両方の例をさらに見ていきます。

## 次のステップ

**Go開発者のためのGemini** シリーズの [**パート2: Geminiでコーディングする**]({{< ref "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini" >}}) では、コーディングエージェントと、Goコードベースで作業するための環境を準備する方法について深く掘り下げます。お楽しみに！
