---
categories:
- Agentic Coding
date: 2026-08-08
heroStyle: big
summary: 「Go開発者のためのGemini」シリーズの第1章。Geminiモデルファミリーのさまざまなモデルと、それらを利用するためのAPIサーフェスに焦点を当てます。
tags:
  - gemini
  - golang
title: 'Go開発者のためのGemini - 第1部: Geminiモデルファミリー'
---

**Go開発者のためのGemini**へようこそ！このシリーズは、GoとGeminiを使用したプロダクションレベルのAIアプリケーション構築に関する完全ガイドです。7つの実践的な章を通じて、エージェンティック・コーディングから、**Genkit**や**ADK**を使用した自律型エージェントの構築、ゲーム開発、そしてフル**G3スタック**（Go、Gemini、GCP）を使用したクラウドへのデプロイまでをカバーします。

第1章では、Geminiモデルファミリー、モデル設定、そして公式の [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) を使用した最初のコード記述を通じて、基礎を固めます。

## Geminiモデルファミリー

私たちは「Gemini」をGoogleのAI製品全体を指す単一の名称として扱いがちですが、実際には、異なる運用的トレードオフのために構築された個別のモデルのファミリーです。

フロンティアモデルが注目を集めがちですが、費用対効果の高いエンジニアリングを行うためには、より小さく専門化されたモデルをいつ使うべきかを知ることが不可欠です。モデルの階層（Tier）によってレイテンシが大きく異なるため、モデルの選択はユーザー体験や製品の採用にも直接影響を与えます。

すべてのタスクに対して高い思考レベル（thinking levels）を持つGemini Proモデルに頼りたくなるかもしれませんが、それが常に正しい選択とは限りません。多くの場合、より良い結果をもたらすことなく、リクエストごとのレイテンシとAPIコストを増加させるだけです。

### モデルの命名規則

Geminiカタログをナビゲートするには、Googleがどのようにモデルに命名しているかを理解することが役立ちます。標準的なモデル文字列は次のパターンに従います：

{{< katex >}}
\[
\text{[ファミリー]}-\text{[バージョン]}-\text{[階層]}{-\text{[修飾子]}}
\]

例：`gemini-3.6-flash` または `gemini-3-pro-image`

* **ファミリー（Family）**: ほとんどのモデルはGeminiファミリーに属しますが、GoogleにはVeoやLyriaなどの他のモデルファミリーもあります。
* **バージョン番号（Version Numbers）**: 知能、コンテキストウィンドウの処理能力、および指示に従う能力の世代的な飛躍を表します。
* **モデル階層（Model Tiers）**:
  * **Pro**: 複雑なマルチステップ推論向けに設計されています。
  * **Flash**: 速度に重点を置いたバランス型モデル。
  * **Flash-Lite**: 速度と高スループットのシンプルなタスク向けに最適化されています。
* **修飾子（Modifiers）**: `gemini-3.1-flash-image` の `image` や `gemini-3.1-flash-live-preview` の `live` のように、サブファミリーや専門化を示す場合があります。また、`-preview` や `-exp`（実験的）などのライフサイクル修飾子が含まれることもあります。

### モデル概要

フロンティアモデルGemini 3.xをはじめとする、主要なGeminiモデルの概要は以下の通りです：

#### Gemini 3.x

Gemini 3.xは主要なフロンティアモデルラインであり、Pro、Flash、Flash-Liteの各階層で利用可能です。これらの汎用モデルは、コード生成やソフトウェアエンジニアリングタスクの第一選択肢でもあります。

現在のモデル：
- `gemini-3.6-flash`: マルチモーダル推論およびエージェンティックタスク向けの高速ワークホース
- `gemini-3.5-flash-lite`: 高スループットマイクロサービス向けの最安・超高速階層
- `gemini-3.1-pro-preview`: 複雑なマルチステップ推論および深層コードベース分析向けの高度な階層

#### Gemini画像モデル（Nano Banana）

技術的にはGeminiファミリーの一部ですが、これは画像生成に特化したモデルであり、マルチモーダルな入力と出力（画像およびテキスト）の両方を提供します。ゼロから画像を生成したり、既存の画像を編集したりすることができます。

現在のモデル：
- `gemini-2.5-flash-image`（別名 Nano Banana）
- `gemini-3-pro-image`（別名 Nano Banana Pro）
- `gemini-3.1-flash-image`（別名 Nano Banana 2）
- `gemini-3.1-flash-lite-image`（別名 Nano Banana 2 Lite）

#### Veo

[ネイティブオーディオ付き動画生成](https://ai.google.dev/gemini-api/docs/veo)に特化したモデル。テキストプロンプトと、トランジションをマークする参照画像（開始および終了フレーム）に基づいて動画が生成されます。Veo 3.1は最大8秒のクリップを生成しますが、7秒刻みで最大20回まで拡張可能です。

現在のモデル：
- `veo-3.1-generate-preview`
- `veo-3.1-lite-generate-preview`（高速生成）

#### Lyria

[Lyria](https://ai.google.dev/gemini-api/docs/music-generation) は音楽生成に特化しており、インストゥルメンタルおよびヴォーカルの楽曲を提供します。Lyriaはテキストと画像の両方を入力として受け入れ、画像は楽曲のインスピレーションとして機能します。自分で歌詞を提供することも、モデルに作成させることも可能です。

現在のモデル：
- `lyria-3-pro-preview`
- `lyria-3-clip-preview`（30秒のショートクリップ）

#### Gemma

[Gemma](https://ai.google.dev/gemma/docs) はGoogleのオープンウェイト（open-weights）モデルファミリーです。Geminiと同じ技術でトレーニングされていますが、自社のインフラにデプロイできるように設計されています。Googleが提供するモデルに加えて、Gemmaにはあらゆるユースケース向けにファインチューニングされたバージョンを制作する[強力なコミュニティ](https://deepmind.google/models/gemma/gemmaverse/)が存在します。

一部のGemmaモデルはローカルマシン上で実行できるほど軽量であり、ネットワーク接続が制限されている、または存在しないユースケースを可能にします。より大きなモデルは非常に強力であり、データ主権やネットワーク隔離が必要なユースケースを可能にします。

#### 注目すべきモデル

- ライブモデル（Live models）: 初期のモデルがバッチまたはリクエスト/レスポンスジョブを処理するのに対し、Googleはリアルタイムストリーミング用のライブモデルも提供しています。名称に `-live` が含まれています（例：`gemini-3.1-flash-live-preview`）。
- Text-to-speech（テキスト読み上げ）: オーディオタグを使用してナレーション制御を行い、テキストから音声を生を生成します（`gemini-3.1-flash-tts-preview`）。
- Computer use（コンピュータ使用）: 画面を「見て」、ブラウザタスクを自動化できるモデル（`gemini-2.5-computer-use-preview-10-2025`）。

このように、Geminiは単一のモデルを遥かに超えた存在です。基本的なチャットボットからマルチモーダル作成、エージェンティック機能まで、すべてをカバーする完全なスイートです。

各モデルの詳細な仕様については、[公式Geminiモデルドキュメント](https://ai.google.dev/gemini-api/docs/models)を参照してください。

## 追加機能

標準的なテキスト補完に加えて、Geminiモデルは複雑なアプリを構築するための追加機能をサポートしています。最も重要なものをいくつか紹介します。

### 思考（Thinking）

Gemini 2.5以降のモデルは、マルチステップの計画、論理、コーディング、および数学的能力を大幅に向上させる内部推論プロセスを使用します。最終的な応答を生成する前に、モデルは「思考トークン（thinking tokens）」を生成してエッジケースを分析し、マルチステップの戦略を計画することで内部的に推論します。

思考は、2.5では `thinking budget`、3.xでは `thinking level` という設定パラメータを使用して制御できる機能です。バジェットや思考レベルが高いほど、モデルは推論フェーズにより多くの時間とトークンを費やします。

思考がアクティブな場合、課金対象となる出力トークンの総数には、生成された出力テキストとモデルの思考トークンの両方が含まれます。タスクの複雑さに応じて思考レベルを調整することは、プロダクションサービス最適化のための重要なステップです。

### 組み込みツールとファンクションコーリング（Function Calling）

ファンクションコーリングにより、Geminiモデルは外部ツール、API、およびデータベースと連携できます。Geminiは、組み込みツール（`google_search` や `code_execution` など）と、アプリケーションレベルで定義されたカスタム機能の両方をサポートしています。

ファンクションコーリングには3つの主要なユースケースがあります：
- **アクションの実行:** ミーティングのスケジュール設定、メールの送信、請求書の作成、スマートホームデバイスの制御など、APIを介して外部システムと連携します。
- **知識の拡張:** 外部データベース、マイクロサービス、およびナレッジベースからリアルタイムまたはプライベートな情報を取得します。
- **機能の拡張:** LLMの限界を超える正確な数学計算、データ変換、またはグラフ生成を実行します。

#### ファンクションコーリングの仕組み

ファンクションコーリングは、アプリケーションとモデルの間で4ステップの実行プロセスに従います：

1. **ツールの宣言**: 関数宣言（名前、明確な説明、およびパラメータのJSON Schema）を定義し、リクエスト設定で渡します。
2. **モデルがツールの意図を特定**: モデルはプロンプトとツール宣言を検査します。ツールが必要な場合、関数名と引数を含む構造化されたツール呼び出しインテントを返します。
3. **関数コードの実行**: モデル自体はコードを実行*しません*。アプリケーションが関数呼び出しリクエストを受け取り、対応するローカルロジックを実行して結果を取得します。
4. **関数結果の返却**: 実行結果を関数結果ステップとしてモデルに送り返します。モデルはこのデータを使用して最終的な自然言語レスポンスを生成するか、追加のツール呼び出しが必要かを判断します。

### 構造化出力（Structured Outputs）

提供された [JSON Schema](https://ai.google.dev/gemini-api/docs/structured-output) に準拠したレスポンスを生成するようにGeminiモデルを設定できます。これにより、テキストからの構造化データの抽出がシンプルになり、モデルのレスポンスをデータ構造に変換する際の脆いパース処理が不要になります。

RESTペイロードに生のJSON Schemaを記述するだけでなく、Google GenAI SDKを使用すると、開発者はPythonの [Pydantic](https://docs.pydantic.dev/) やGoの構造体タグ（struct tags）などの言語組み込み構造を使用してスキーマを定義できます。

## プログラムによるモデルの利用

モデルのエコシステムと機能について説明したので、これらのAPIをGoアプリケーションに組み込む方法を見てみましょう。

ユースケースごとに異なるモデルが存在するのと同様に、いくつかのAPIサーフェスが利用可能です。最も基本的な「Generate Content」から始めましょう。

### Generate Content API

これは最も基本的な[生成API](https://ai.google.dev/api/generate-content#method:-models.generatecontent)です。単一のリクエストを受け取りレスポンスを返すステートレスなインターフェースです。複数ターン（マルチターン）の会話では、アプリケーションが呼び出しごとに完全なチャット履歴を送信する必要があります。

これには、コンテキストウィンドウの制限内に収めるために会話履歴をアクティブに管理することが求められます。通常、アプリケーションは履歴がしきい値に達すると要約を行います。長いセッションでの入力コストを削減するために、Gemini APIはGemini 2.5以降のすべてのモデルで[暗黙的キャッシュ（implicit caching）](https://ai.google.dev/gemini-api/docs/caching)をサポートしているほか、重いペイロード向けに[明示的キャッシュ（explicit caching）](https://ai.google.dev/gemini-api/docs/generate-content/caching)もサポートしています。

Generate Content APIはシンプルなステートレス生成に適していますが、より新しい機能を持つInteractions APIに徐々に置き換えられつつあります。

### Interactions API

> 注：本日時点で、Interactions APIは公式のGo GenAI SDKではまだサポートされていません。実装の進捗状況はこの [GitHub issue](https://github.com/googleapis/go-genai/issues/658) でトラッキングされています。

[Interactions API](https://ai.google.dev/gemini-api/docs/interactions-overview)は、シンプルなチャットやツールの利用から複雑なエージェンティックワークフローまで、すべてのタスク向けに設計されたGoogleの統一インターフェースです。会話履歴をサーバー側で管理できるため、アプリケーションで管理する必要がなくなります。

### Live API

[Live API](https://ai.google.dev/gemini-api/docs/live-api)は、WebSocketsを介したリアルタイムの双方向音声および動画会話を可能にします。ユーザーが話したり割り込んだりしたことを自動的に検出し、ライブセッション内でWeb検索やファンクションコーリングなどのツールをサポートしながら、自然な音声対話を実現します。

### Batch API

[Batch API](https://ai.google.dev/gemini-api/docs/batch-api)を使用すると、大量のデータを半額で非同期処理できます。ジョブはオフピーク時間帯にバックグラウンドで実行され（通常24時間以内に完了）、緊急性の低いワークロードに最適です。

### Managed Agents API

[Managed Agents](https://ai.google.dev/gemini-api/docs/agents)は、AIエージェントが自律的にタスクを計画・実行する完全ホスト型の実行環境を提供します。単一のAPI呼び出しで、PythonやNodeなどの実行環境がプリインストールされたOS隔離のLinuxサンドボックスがプロビジョニングされ、エージェントがコードの実行、ファイルの管理、Webの閲覧を行えるようになります。

Googleは、箱から出してすぐに使える2つの事前構築済みマネージドエージェントを提供しています：
- **Antigravity Agent** (`antigravity-preview-05-2026`): コード実行、ファイル管理、およびWebアクセスのための、Gemini 3.6 Flash（Gemini 3.5 FlashまたはFlash-Liteに変更可能）駆動のデフォルト汎用エージェント。
- **Deep Research Agent** (`deep-research-preview-04-2026`): 複数のソースからWebデータをクエリし、バックグラウンドで詳細な調査レポートをまとめる自律型調査エージェント。

また、システムルールをインラインで定義したり、`AGENTS.md` ファイルをマウントしたり、構造化されたスキルディレクトリ（`SKILL.md`）を添付したり、ローカルファイル、Cloud Storageバケット、Gitリポジトリをリモートワークスペース（`/workspace`）に直接マウントすることで、Antigravityエージェントを拡張することも可能です。

## アクセスと請求

Geminiを統合する際、Googleはニーズに応じて2つの主要なアクセスおよび請求モードを提供しています：

1. **Google AI Studio (Google AI)**: APIキー（`GEMINI_API_KEY` または `GOOGLE_API_KEY`）を使用してGemini API経由でリクエストをルーティングします。プロトタイピング、個人プロジェクト、個人開発アプリ、および迅速な開発者のオンボーディングに最適です。
2. **Gemini Enterprise (旧 Vertex AI)**: Google Cloud IAM、アプリケーションデフォルト認証情報（ADC）、サービスアカウントキー、またはOAuth 2.0ユーザーアクセストークンを使用して、Google Cloudエンドポイント経由でリクエストをルーティングします。厳格なデータプライバシー、セキュリティコンプライアンス、SLA、GCPリソース管理、および確約利用割引を必要とするプロダクションエンタープライズワークロードに最適です。

## Go GenAI SDK

それでは、これがGoコードでどのように機能するかを見てみましょう。

GoアプリケーションにGeminiを統合するための公式SDKは [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai) です。

すべてのGoogleモデルと、Google AIおよびGemini Enterpriseの両方の認証をサポートするように設計されているため、時々「統一」SDKと呼ばれることがあります。非推奨（deprecated）となったレガシーパッケージ `github.com/google/generative-ai-go` を置き換えるものです。

`go get` でインストールします：

```bash
go get google.golang.org/genai
```

Gemini Enterpriseの認証と請求を使用する例を以下に示します：

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

次のコマンドを使用してこのプログラムを実行できます：

```sh
export GOOGLE_CLOUD_PROJECT="your-project-id-goes-here"
go run main.go
```

結果は以下の通りです：

![Goターミナルで生成された魔法使いの猫の画像出力](image.png "AIの真の目的：無限の猫画像生成")

これはSDKの基本的な使い方を示すシンプルな例ですが、このシリーズを通じて、Go GenAI SDKと [Genkit](https://genkit.dev/) や [Agent Development Kit (ADK)](https://adk.dev/) などの上位フレームワークの両方の例をさらに見ていきます。

## 次のステップ

**Go開発者のためのGemini** シリーズの**第2部**では、コーディングエージェントと、Goコードベースで作業するための環境を準備する方法について深く掘り下げます。お楽しみに！
