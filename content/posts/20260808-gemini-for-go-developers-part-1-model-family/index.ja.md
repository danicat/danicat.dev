---
categories:
- Applied GenAI
date: 2026-08-08
heroStyle: big
series:
- Gemini for Go Developers
series_order: 1
summary: "Geminiファミリーの各種モデル、その能力、そしてそれらをプログラムから利用する方法について学びます。"
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

**Go開発者のためのGemini** へようこそ！このシリーズは、GoでAIを活用したソフトウェアを構築するための完全ガイドです。全7章の実践的なチュートリアルを通じて、エージェンティック・コーディングから、**Genkit** や **ADK** を使用した自律型エージェントの構築、ゲーム開発、そして **G3 Stack**（Go、Gemini、GCP）を活用したクラウドへのアプリケーションのデプロイまでを網羅します。

第1章では、Geminiモデルファミリーの構成やモデル設定を探り、公式の [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) を使って最初のコードを書きながら基礎を固めていきます。

## Geminiモデルファミリー

「Google」を検索の代名詞として使うのと同じように、私たちは「Gemini」をGoogleのAI製品全般を指す単一の名前として扱いがちです。しかし実際には、Geminiは運用上の異なるトレードオフに合わせて構築された、複数のモデルからなるファミリーです。

見出しを飾るのは最先端のフロンティアモデルですが、費用対効果の高いエンジニアリングを実現するには、小型モデルや特化型モデルをいつ活用すべきかを理解しておくことが不可欠です。モデルの階層（Tier）によってレイテンシは大きく異なるため、モデルの選定は[ユーザー体験](https://services.google.com/fh/files/blogs/google_delayexp.pdf)やプロダクトの定着率にも直接影響を与えます。

あらゆるタスクに対して思考レベル（thinking levels）を高く設定したGemini Proモデルを選びたくなるかもしれませんが、それが常に正しい判断とは限りません。多くの場合、より良い結果が得られないまま、リクエストごとのレイテンシとAPIコストが無駄に増加するだけです。

### モデルの命名規則

Geminiカタログを把握するには、Googleがモデルにどのような命名を行っているかを理解しておくと役立ちます。標準的なモデル識別文字列は次のパターンに従います：

{{< katex >}}
\[
\text{[family]}-\text{[version]}-\text{[tier]}{-\text{[modifier]}}
\]

例えば、`gemini-3.7-flash` や `gemini-3-pro-image` のようになります。

**Family（ファミリー）**: ほとんどのモデルはGeminiファミリーに属しますが、GoogleにはVeoやLyriaといった他のモデルファミリーもあります。
**Version Numbers（バージョン番号）**: 知能、コンテキストウィンドウの処理能力、指示追従性における世代ごとの飛躍を表します。
**Model Tiers（モデル階層）**:
* **Pro**: 複雑なマルチステップ推論向けに設計されています。
* **Flash**: 速度を重視したバランスの取れたモデルです。
* **Flash-Lite**: 速度と高スループットが求められるシンプルなタスク向けに最適化されています。
**Modifiers（修飾子）**: `gemini-3.1-flash-image` の `image` や `gemini-3.1-flash-live-preview` の `live` のように、サブファミリーや特化分野を示します。また、`-preview` や `-exp`（experimental / 実験的）などのライフサイクル修飾子が含まれる場合もあります。

### モデル概要

フロンティアモデルであるGemini 3.xをはじめとする、主要なGeminiモデルの概要は以下の通りです：

#### Gemini 3.x

Gemini 3.xは主力となるフロンティアモデルラインであり、Pro、Flash、Flash-Liteの各階層で提供されています。これらの汎用モデルは、コード生成やソフトウェアエンジニアリングタスクにおける第一選択肢でもあります。

現在の主なモデル：
- `gemini-3.6-flash`: マルチモーダル推論やエージェンティックタスク向けの高速な主力モデル
- `gemini-3.5-flash-lite`: 高スループットなマイクロサービス向けの最安・超高速モデル
- `gemini-3.1-pro-preview`: 複雑なマルチステップ推論や深層コードベース分析向けの高度なモデル

#### Gemini画像モデル（Nano Banana）

技術的にはGeminiファミリーの一部ですが、画像生成に特化したモデルであり、マルチモーダルな入力と出力（画像およびテキスト）の両方を提供します。ゼロからの画像生成だけでなく、既存画像の編集も可能です。

現在の主なモデル：
- `gemini-2.5-flash-image`（通称 Nano Banana）
- `gemini-3-pro-image`（通称 Nano Banana Pro）
- `gemini-3.1-flash-image`（通称 Nano Banana 2）
- `gemini-3.1-flash-lite-image`（通称 Nano Banana 2 Lite）

#### Veo

[ネイティブ音声付き動画生成](https://ai.google.dev/gemini-api/docs/veo)に特化したモデルです。動画はテキストプロンプトや、トランジションを指定するキー画像（開始フレームおよび終了フレーム）、参照画像に基づいて生成されます。Veo 3.1は最大8秒のクリップを生成しますが、7秒単位で最大20回まで拡張可能です。

現在の主なモデル：
- `veo-3.1-generate-preview`
- `veo-3.1-lite-generate-preview`（高速生成）

#### Lyria

[Lyria](https://ai.google.dev/gemini-api/docs/music-generation) は音楽生成に特化しており、インストゥルメンタルとボーカルの両方の楽曲を制作できます。Lyriaはテキストと画像の両方を入力として受け付け、画像を着想（インスピレーション）として作曲に活用します。歌詞は自ら指定することも、モデルに作成させることも可能です。

現在の主なモデル：
- `lyria-3-pro-preview`
- `lyria-3-clip-preview`（30秒のショートクリップ）

#### Gemma

[Gemma](https://ai.google.dev/gemma/docs) は、Googleが提供するオープンウェイト（open-weights）モデルファミリーです。Geminiを支える技術と同じテクノロジーでトレーニングされていますが、自前のインフラストラクチャにデプロイして運用できるように設計されています。Googleが提供するモデルにとどまらず、Gemmaには多種多様なユースケースに合わせてファインチューニングされたモデルを生み出す[強力なコミュニティ](https://deepmind.google/models/gemma/gemmaverse/)が存在します。

一部のGemmaモデルはローカルマシン上で実行できるほど軽量であり、ネットワーク接続が制限されている環境や存在しない環境でのユースケースを可能にします。より大規模なモデルは非常に高い能力を備えており、データの主権（sovereignty）や完全なネットワーク分離が求められるユースケースに対応できます。

#### その他の注目モデル

- Liveモデル: 初期のモデルがバッチやリクエスト・レスポンス処理を担うのに対し、Googleはリアルタイムストリーミング用のライブモデルも提供しています。名前に `-live` が含まれているのが特徴です（例：`gemini-3.1-flash-live-preview`）。
- Text-to-speech: オーディオタグを使用したナレーション制御により、テキストから音声を生成します（`gemini-3.1-flash-tts-preview`）。
- Computer use: 画面を「見て」、ブラウザタスクを自動化できるモデルです（`gemini-2.5-computer-use-preview-10-2025`）。

このように、Geminiは単一のモデルをはるかに超えた存在です。基本的なチャットボットからマルチモーダルな生成、高度なエージェンティック機能までを網羅する包括的なスイートとなっています。

各モデルの詳細な仕様については、[公式Geminiモデルドキュメント](https://ai.google.dev/gemini-api/docs/models) を参照してください。

## 追加機能

標準的なテキスト補完にとどまらず、Geminiモデルは複雑なアプリを構築するための追加機能をサポートしています。ここでは特に重要な機能をいくつか紹介します。

### 思考（Thinking）

Gemini 2.5以降のモデルは、マルチステップの計画、論理的思考、コーディング、および数学的処理能力を大幅に向上させる内部推論プロセスを使用します。最終的なレスポンスを生成する前に、モデルは内部で「思考トークン（thinking tokens）」を生成してエッジケースを分析し、複数ステップの戦略を立案します。

思考は、2.5の `thinking budget` や 3.x の `thinking level` という設定パラメータを使用して制御できる機能です。バジェットや思考レベルを高く設定するほど、モデルは推論フェーズにより多くの時間とトークンを費やします。

思考が有効な場合、課金対象となる総出力トークン数には、生成された出力テキストとモデルが生成した思考トークンの両方が含まれます。タスクの複雑さに基づいて思考レベルを調整することは、プロダクション環境のサービスにとって極めて重要なステップです。

### 組み込みツールとファンクションコーリング（Function calling）

ファンクションコーリングを使用すると、Geminiモデルを外部ツール、API、データベースと連携させることができます。Geminiは、組み込みツール（`google_search` や `code_execution` など）と、アプリケーションレベルで定義されたカスタム関数の両方をサポートしています。

ファンクションコーリングには主に3つのユースケースがあります：
- **アクションの実行（Take Actions）:** 会議のスケジュール設定、メール送信、請求書作成、スマートホーム機器の操作など、APIを介して外部システムと連携します。
- **知識の拡張（Augment Knowledge）:** 外部データベース、マイクロサービス、ナレッジベースから、リアルタイム情報やプライベートな情報を取得します。
- **能力の拡張（Extend Capabilities）:** LLMの限界を超える高精度の数学計算、データ変換、グラフ生成などを実行します。

#### ファンクションコーリングの仕組み

ファンクションコーリングは、アプリケーションとモデルの間で次の4ステップの実行プロセスに従います：

1. **ツールの宣言（Declare tools）**: 関数の宣言（関数名、明確な説明、パラメータのJSON Schema）を定義し、リクエスト設定で渡します。
2. **モデルがツールの意図を特定（Model identifies tool intent）**: モデルはプロンプトとツール宣言を検査します。ツールが必要な場合、関数名と引数を含む構造化されたツール呼び出しの意図（intent）を返します。
3. **関数コードの実行（Execute function code）**: モデル自身はコードを実行*しません*。アプリケーションが関数呼び出しリクエストを受け取り、対応するローカルロジックを実行して結果を取得します。
4. **関数の結果を返却（Return function result）**: 実行結果を関数結果ステップとしてモデルに送り返します。モデルはこのデータを使用して最終的な自然言語レスポンスを生成するか、追加のツール呼び出しが必要かを判断します。

### 構造化出力（Structured outputs）

Geminiモデルを設定して、提供された [JSON Schema](https://ai.google.dev/gemini-api/docs/structured-output) に厳密に準拠したレスポンスを生成させることができます。これにより、テキストからの構造化データ抽出が簡単になり、モデルのレスポンスをデータ構造に変換する際の壊れやすいパース処理が不要になります。

RESTペイロードに生のJSON Schemaを記述するだけでなく、Google GenAI SDKを使用すると、Pythonの [Pydantic](https://docs.pydantic.dev/) やGoの構造体タグ（struct tags）など、言語ネイティブの構文構造を使用してスキーマを定義できます。

## プログラムからモデルを利用する

モデルのエコシステムと各種機能を把握したところで、これらのAPIをGoアプリケーションに組み込む方法を見ていきましょう。

用途に応じて異なるモデルが存在するのと同様に、複数のAPIサーフェスが提供されています。まずは最も基本的な「Generate Content」から見ていきましょう。

### Generate Content API

これは最も基本的な[生成API](https://ai.google.dev/api/generate-content#method:-models.generatecontent)です。単一のリクエストを受け取ってレスポンスを返すステートレスなインターフェースです。複数ターンにわたる会話の場合、アプリケーションは呼び出しごとに完全なチャット履歴を送信する必要があります。

そのため、コンテキストウィンドウの制限内に収まるように会話履歴を能動的に管理する必要があります。一般的に、アプリケーションは履歴が一定のしきい値に達した時点で要約を行います。長時間のセッションにおける入力コストを削減するために、Gemini APIはGemini 2.5以降のすべてのモデルで[暗黙的キャッシュ（implicit caching）](https://ai.google.dev/gemini-api/docs/caching)をサポートしているほか、大容量ペイロード向けの[明示的キャッシュ（explicit caching）](https://ai.google.dev/gemini-api/docs/generate-content/caching)も利用できます。

Generate Content APIはシンプルなステートレス生成に適していますが、より新しく高機能なInteractions APIへと段階的に置き換えが進んでいます。

### Interactions API

> 注：現時点では、Interactions APIは公式のGo GenAI SDKでまだサポートされていません。実装の進捗状況はこの [GitHub issue](https://github.com/googleapis/go-genai/issues/658) で追跡されています。

[Interactions API](https://ai.google.dev/gemini-api/docs/interactions-overview) は、シンプルなチャットやツール利用から複雑なエージェンティックワークフローまで、あらゆるタスク向けに設計されたGoogleの統合インターフェースです。会話履歴をサーバー側で管理できるため、アプリケーション側で履歴を管理する必要がありません。

### Live API

[Live API](https://ai.google.dev/gemini-api/docs/live-api) は、WebSocketを介したリアルタイムかつ双方向の音声・動画の会話を可能にします。ユーザーの発話や割り込みを自動的に検知し、ライブセッション内でWeb検索やファンクションコーリングなどのツールを直接サポートしながら、自然な音声インタラクションを実現します。

### Batch API

[Batch API](https://ai.google.dev/gemini-api/docs/batch-api) を使用すると、大量のデータを半額の料金で非同期に処理できます。ジョブはオフピーク時間帯にバックグラウンドで実行され（通常24時間以内に完了）、緊急性の低いワークロードに最適です。

### Managed Agents API

[マネージドエージェント（Managed agents）](https://ai.google.dev/gemini-api/docs/agents) は、AIエージェントが自律的にタスクを計画・実行できるフルホスト型のランタイム環境を提供します。単一のAPI呼び出しを行うだけで、PythonやNodeなどの事前インストール済みランタイムを備えたOS分離のLinuxサンドボックスがプロビジョニングされ、エージェントがコードの実行、ファイル管理、Webの閲覧を自律的に行えるようになります。

Googleは、最初から使える2つの事前構築済みマネージドエージェントを提供しています：
- **Antigravity Agent** (`antigravity-preview-05-2026`): コード実行、ファイル管理、Webアクセスをこなす、Gemini 3.6 Flash（Gemini 3.5 FlashまたはFlash-Liteにも設定可能）駆動のデフォルト汎用エージェント。
- **Deep Research Agent** (`deep-research-preview-04-2026`): 複数のWebソースからデータを収集し、詳細なリサーチレポートをバックグラウンドでまとめる自律型リサーチエージェント。

また、システムルールをインラインで定義したり、`AGENTS.md` ファイルをマウントしたり、構造化されたスキルディレクトリ（`SKILL.md`）をアタッチしたり、ローカルファイル、Cloud Storageバケット、Gitリポジトリをリモートワークスペース（`/workspace`）に直接マウントすることで、Antigravityエージェントを拡張することもできます。

## アクセスと請求

Geminiを統合する際、Googleはニーズに応じて2つの主要なアクセスおよび請求モードを提供しています：

1. **Google AI Studio (Google AI)**: APIキー（`GEMINI_API_KEY` または `GOOGLE_API_KEY`）を使用してGemini API経由でリクエストをルーティングします。プロトタイピング、個人プロジェクト、個人開発アプリ、迅速な開発着手に最適です。
2. **Gemini Enterprise（旧 Vertex AI）**: Google Cloud IAM、Application Default Credentials (ADC)、サービスアカウントキー、またはOAuth 2.0ユーザートークンを使用して、Google Cloudエンドポイント経由でリクエストをルーティングします。厳格なデータプライバシー、セキュリティコンプライアンス、SLA、GCPリソース管理、確約利用割引が必要なエンタープライズの本番ワークロードに最適です。

## Go GenAI SDK

それでは、実際のGoコードでの動作を見ていきましょう。

GoアプリケーションにGeminiを統合するための公式SDKは [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai) です。

すべてのGoogleモデルに対応し、Google AIとGemini Enterpriseの両方の認証方式をサポートするよう設計されているため、「統合SDK（unified SDK）」と呼ばれることもあります。すでに非推奨（deprecated）となった旧パッケージ `github.com/google/generative-ai-go` を置き換えるものです。

`go get` でインストールします：

```bash
go get google.golang.org/genai
```

Gemini Enterpriseの認証と請求を利用する実装例は以下の通りです：

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

このサンプルを実行するには、AI Platform APIが有効化されたGoogle Cloudプロジェクトが必要です。例：

```sh
export GOOGLE_CLOUD_PROJECT="your-project-id-goes-here"
go run main.go
```

実行結果は次のようになります：

![Goターミナルで生成された魔法使いの猫の画像出力](image.png "AIの真の目的：無限の猫画像生成")

これはSDKの使い方を示すシンプルな例にすぎませんが、本シリーズを通じて、Go GenAI SDKと [Genkit](https://genkit.dev/) や [Agent Development Kit (ADK)](https://adk.dev/) などの高レベルフレームワークの両方の例をさらに見ていきます。

## 次のステップ

**Go開発者のためのGemini** シリーズの [**パート2: Geminiでコーディングする**]({{< ref "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini" >}}) では、コーディングエージェントと、Goコードベースで作業するための環境を準備する方法について深く掘り下げます。お楽しみに！
