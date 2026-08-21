---categories:
- Applied GenAI
date: 2026-02-16
heroStyle: big
summary: Go、Genkit、そして Nano Banana Pro（Gemini 3 Pro Image）のネイティブ 4K 機能を活用して、高精細な写真復元ツールを構築する方法を解説します。
tags:
  - gemini
  - genkit
  - golang
  - nano-banana
  - tutorial
title: "Genkit と Gemini 3 で古い写真に命を吹き込む"
slug: "genkit-go-photo-restoration"
aliases:
  - "/ja/posts/20260216-genkit-go/"
description: "Genkit Go と Gemini 3 Pro Image (Nano Banana Pro) を使ってネイティブ4K写真復元ツールGlowUpを構築するチュートリアル。dotpromptテンプレートとFlow実装を解説。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "Genkit Go >= 0.1.0"
  - "Google Cloud Vertex AI"
---

仕事柄、さまざまな人と知り合う機会が多いのですが、よく話題にのぼるのが私のルーツについてです。東欧系の響きを持つ苗字であるだけでなく、見た目もまさにその系統なので、ポーランドかチェコの出身だろうとよく思われます。そのため、実際にはブラジル出身だと伝えると、大抵の人は驚きます。

我が家は歴史的な記録を残すのが本当に**苦手**で、家族の誰も自分たちのルーツを正確には把握していません。おそらく自分たちでもそれを自覚しているからこそ、把握できているわずかな家族の歴史や、それすらも薄れつつあることについてよく話をします。歳を重ねるにつれて、真っ先に失われていくのは記憶であり、やがて書類や写真も劣化していきます。最後に祖母に会ってから30年が経ち、彼女の顔がおぼろげな輪郭でしか思い出せないと気づいたときには、何とも言えない切なさがこみ上げてきます。だからこそ、私にとって写真はかけがえのない存在なのです。薄れゆく自分自身の記憶の風化に抗うための、最後の砦だからです。

近年撮影されたものなら、簡単に複製してクラウドにいくらでも冗長なバックアップを保存できます。しかし、ここで話しているのはデジタル時代以前の大切な形見です。たとえ何年も前にスキャンしていたとしても、多くの写真はすでに何十年分もの埃やカビ、経年劣化に見舞われています。時間は止まっていても、状態が自然に良くなることはありません。

ですが、生成AIの進化のおかげで、決してすべてが失われたわけではありませんでした。経年変化によるダメージを修復するだけでなく、カラー化してアップスケールし、現代のクオリティへと蘇らせることができるようになったのです。こうして生まれたのが、「GlowUp」という小さなツールです。


以下はそのような復元の一例です。

![バナナパイを作る祖母の傷んだモノクロ写真（オリジナル）](original.jpg "オリジナル：特製バナナパイを作る私の祖母")

![Nano Banana Pro を使って高精細4Kで復元・カラー化した写真](restored.png "復元後：Nano Banana Pro による修復とカラー化")

この記事では、[Gemini Nano Banana Pro](https://ai.google.dev/gemini-api/docs/image-generation) と [Genkit Go](https://genkit.dev/docs/get-started/?lang=go) を使って、GlowUp をゼロから構築する方法を紹介します。

## 構成要素

今回 Nano Banana Pro（別名 Gemini 3 Pro Image Preview）を選んだのは、現在 Gemini ファミリーの中で最も高度な画像処理モデルだからです。標準の Nano Banana（Gemini 2.5 Flash Image）も素晴らしいモデルですが、Pro バージョンの方がより高品質な出力が得られ、多少の試行錯誤は必要なものの、指示に対する追従性（instruction following）も優れていると感じています。

クライアントサイドの実装では、[go-genai](https://pkg.go.dev/google.golang.org/genai) のような低レベル SDK を直接使うのではなく、Genkit を採用しました。低レベルなコードを記述する場合と比べて、次のような優れた開発者体験（QoL）が得られるためです。

- **モデルに依存しない設計（Model-agnostic）**: プラグインを1行差し替えるだけで、Google 以外のモデルやローカルモデルを含め、異なるモデルを自在に検証可能
- **Dev UI の標準サポート**: モデルやプロンプトのテスト、モデル呼び出しのトレースなどがブラウザ上の GUI で手軽に行える
- **柔軟なアーキテクチャ**: CLI アプリケーションとしても Web サーバーとしても同一ロジックを容易にパッケージ化可能

GlowUp は、コマンドラインツールとしても Web サーバーとしても実行できる単一バイナリとして構築されています。この柔軟性のおかげで、手元のターミナルからローカルでサクッと修復を実行することも、同じコードをクラウドサービスとしてデプロイすることもできます。将来的には、父でも自分の写真コレクションを簡単に修復できるような Web アプリに発展させることも可能です。

## Genkit Go の基本

[Genkit](https://firebase.google.com/docs/genkit) は、AI アプリケーション開発に本番水準の信頼性と開発プラクティスをもたらすために設計されたオープンソースフレームワークです。Go エンジニア*なら、AI 機能のための「標準ライブラリ」のような存在と考えると分かりやすいでしょう。_（* もし Go 以外の言語をお使いの場合でも、Genkit は JavaScript / TypeScript や Python もサポートしていますので公式ドキュメントをチェックしてみてください。）_

Go 版 Genkit での最小限の「Hello World」は以下のようになります。フレームワークの初期化に `googlegenai` プラグインを使用している点に注目してください。

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/server" // Import the server plugin
)

func main() {
	ctx := context.Background()
	// Initialize Genkit with the Google GenAI plugin (Vertex AI)
	g := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.VertexAI{}))

	// Define a simple Flow
	genkit.DefineFlow(g, "hello", func(ctx context.Context, name string) (string, error) {
		// Generate text using a model
		resp, err := genkit.GenerateText(ctx, g,
			ai.WithModelName("vertexai/gemini-2.5-flash"),
			ai.WithPrompt(fmt.Sprintf("Say hello to %s", name)))
		if err != nil {
			return "", err
		}
		return resp, nil
	})

	// Start the flow server manually
	mux := http.NewServeMux()
	// Register all flows defined in 'g'
	for _, flow := range genkit.ListFlows(g) {
		mux.HandleFunc("POST /"+flow.Name(), genkit.Handler(flow))
	}

	if err := server.Start(ctx, ":8080", mux); err != nil {
		log.Fatal(err)
	}
}
```

この短いコードスニペットの中には、多くの重要な処理が凝縮されています。各要素を少し詳しく見ていきましょう。

### プラグイン（Plugins）
コードを Vertex AI、Google AI、Ollama などのプロバイダーに接続するアダプターです。Google のモデルを利用する場合、`googlegenai` プラグインを使用します。このプラグインは両方のバックエンドに対応しています。

* **Google AI (Studio):** API キーを使用。プロトタイピングや個人プロジェクトに最適です。
```go
// Use Google AI (API Key)
googlegenai.Init(ctx, &googlegenai.Config{APIKey: "MY_KEY"})
```

* **Vertex AI (Google Cloud):** Google Cloud IAM 認証を使用。本番環境のワークロードやエンタープライズ機能に推奨されます。
```go
// Use Vertex AI (Cloud Auth)
googlegenai.Init(ctx, &googlegenai.VertexAI{ProjectID: "my-project", Location: "us-central1"})
```

**注:** 旧バージョンの Genkit を使ったことがある方は、個別の `vertexai` プラグインと `googleai` プラグインに見覚えがあるかもしれません。現在これらは単一の `googlegenai` プラグインに統合されています。

### モデル（Models）
コンテンツを生成する実際の LLM（Gemini や Claude など）です。`vertexai/gemini-2.5-flash` のような文字列でモデル名を指定して参照します。

```go
	resp, err := genkit.GenerateText(ctx, g, 
		ai.WithModel("vertexai/gemini-2.5-flash"),
		ai.WithTextPrompt("Tell me a joke"))
```

### プロンプト（Prompts）

上の例のようにコード内にプロンプトを直接ハードコードすることもできますが、保守性を高めるためには独立したファイルに切り出すのがベストプラクティスです。Genkit は外部プロンプトの読み込みに `dotprompt` を使用します。

`dotprompt` ファイル（`*.prompt`）は、**Frontmatter** と **Template** の2つの主要部分で構成されます。

**1. Frontmatter（設定）**
* **`model`**: モデル識別子（例: `vertexai/gemini-2.5-flash`）
* **`config`**: `temperature` や `topK` などの生成パラメーター、またはモデル固有の設定（例: `imageConfig`）
* **`input`**: Go コードから渡される引数を定義する JSON スキーマ
* **`output`**: 構造化出力用のスキーマ

**2. Template（指示）**
プロンプト本文には Handlebars 構文を使用します。
* **変数（Variables）**: `{{theme}}` のようなプレースホルダーは、入力スキーマで定義した値に置換されます。
* **ロール（Roles）**: `{{role "system"}}` や `{{role "user"}}` ヘルパーにより、システム指示とユーザーのクエリを明確に構造化できます。
* **メディア（Media）**: `{{media url=myImage}}` ヘルパーにより、マルチモーダルデータ（画像や動画）をモデルのコンテキストに直接注入できます。

```yaml
---
model: vertexai/gemini-2.5-flash
input:
  schema:
    theme: string
---
{{role "system"}}
You are a helpful assistant.

{{role "user"}}
Tell me a joke about {{theme}}.
```

### フロー（Flows）
Genkit において **Flow（フロー）** は処理実行の基本単位であり、以下の機能を提供します。
1. **可観測性（Observability）**: すべてのフロー実行時に、Genkit Developer UI や Google Cloud Trace で確認できるトレースとメトリクス（レイテンシ、トークン消費量、成功率など）が自動生成されます。
2. **型安全性（Type Safety）**: フローは入力・出力スキーマによって厳密に型付けされており、複数の AI 処理をパイプラインとして連結する際のランタイムエラーを防ぎます。
3. **デプロイの容易さ（Deployability）**: フローの実行ロジックはサーバー配信ロジックから完全に分離されています。デプロイする際は `genkit.Handler` でラップするだけで標準の `http.Handler` に変換できるため、Go の標準ライブラリ（`net/http`）や好みの Web フレームワークにそのまま組み込めます。

```go
    // Define a flow
    myFlow := genkit.DefineFlow(g, "myFlow", func(ctx context.Context, input string) (string, error) {
        return "Processed: " + input, nil
    })

    // Expose it as an HTTP handler
    http.HandleFunc("/myFlow", genkit.Handler(myFlow))
```

## Nano Banana Pro

今回の写真復元を支えるエンジンは、愛称として「Nano Banana Pro」と呼ばれる **Gemini 3 Pro Image** です。

これは、前世代（および現行の「Flash」モデル）と比べても飛躍的な進化を遂げています。Gemini 2.5 Flash は非常に高速で基本的な画像生成（`gemini-2.5-flash-image`）に対応していますが、**Nano Banana Pro**（`gemini-3-pro-image-preview`）は高度なマルチモーダル推論を行うために設計されています。

単にピクセルを「見る」だけでなく、意味論的なコンテキストを理解します。「紙の表面についた傷」と「顔の傷跡」を正確に見分けることができ、1950年代のキッチンなら現代的なフローリングではなくリノリウムの床が敷かれているはずだといった背景知識も把握しています。

### 主な違い

* **Flash (`gemini-2.5-flash-image`)**: 速度とコスト効率に最適化。サムネイルやシンプルなイラスト作成に向いています。最大解像度は 1024x1024。
* **Pro (`gemini-3-pro-image-preview`)**: 忠実度と推論力に最適化。写真復元には欠かせないネイティブ **4K 解像度**（最大 4096px）の生成に対応しています。

このモデルは出力を微調整するための `imageConfig` パラメーターも受け付けます。
* `imageSize`: `"4K"` または `"2K"`
* `aspectRatio`: `"16:9"`、`"4:3"`、`"1:1"` など

注意すべき重要な点として、このモデルは常にテキストと画像の両方を含むインターリーブ形式（interleaved）のレスポンスを返します。他の画像生成モデルとは異なり、画像データ単体での出力はサポートされていません。そのため、後述する抽出ロジックのように、マルチパートレスポンスの中から画像データを適切に見つけ出す処理が必要になります。

**注:** 執筆時点では、このモデルは Vertex AI の `global` リージョンでのみ利用可能です。Vertex AI クライアントの設定時にはご注意ください。

## 各コンポーネントの連携

続いて、GlowUp がこれらのパーツをどのように連携させているかを見ていきましょう。写真修復の専門家としてのペルソナを定義する**プロンプトファイル**と、画像処理を実行する**フロー**を使用します。

### プロンプト

モデルの設定と指示を定義するために `.prompt` ファイルを使用します。ここで `4K` 解像度を指定しているため、Go のコード側をシンプルに保てている点に注目してください。

```yaml
---
model: vertexai/gemini-3-pro-image-preview
config:
  imageConfig:
    imageSize: "4K"
input:
  schema:
    url: string
    contentType: string
---

{{role "system"}}
You are GlowUp, a professional-grade photo restorer.
Your goal is to provide a "surgical" restoration service that transforms vintage, damaged, or monochrome photographs into high-fidelity 4K colourised versions.

RULES:
1. **Grounding**: You are strictly grounded in the original source pixels. Do NOT add new objects (trees, people, buildings, etc.) that are not present in the source. Additionally, do NOT remove any elements from the source, unless they are clearly defects that do not belong in the original scene.
2. **Fidelity**: Preserve the original facial expressions and identity of subjects. Do NOT "beautify" or alter features in a way that changes the person's identity.
3. **Background**: Preserve background fidelity. Overexposed light sources (like windows) must remain as light sources. Do not "fill in" missing details with invented scenery.
4. **Colourisation**: If the image is monochrome, colourize it realistically, respecting historical accuracy where possible.
5. **Upscaling**: Output a high-fidelity image.

{{role "user"}}
Restore this photo.
Image: {{media url=url contentType=contentType}}
```

### フロー

Go のコードは非常にシンプルで無駄がありません。この構成では、フロー定義がプロンプトをロードし、マルチモーダル入力をモデルに渡す役割を担います。

```go
// main.go (Flow Definition)
type Input struct {
	URL         string `json:"url,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

func defineGlowUpFlow(g *genkit.Genkit) *core.Flow[Input, string, struct{}] {
	return genkit.DefineFlow(g, "glowUp", func(ctx context.Context, input Input) (string, error) {
		prompt := genkit.LookupPrompt(g, "glowup")
		if prompt == nil {
			return "", errors.New("prompt 'glowup' not found")
		}

		resp, err := prompt.Execute(ctx, ai.WithInput(input))
		if err != nil {
			return "", fmt.Errorf("generation failed: %w", err)
		}

		return resp.Media(), nil
	})
}
```

ローカルファイルをそのまま扱えるように、`fileToDataURI` ヘルパー関数を用意しています。この関数はローカルファイルを読み込み、`http.DetectContentType` で MIME タイプを自動判別した上で、Gemini API が要求する標準の Base64 Data URI 形式にエンコードします。これにより、拡張子をハードコードすることなく、さまざまなスキャン形式の画像を忠実度を保ったまま処理できます。

```go
func fileToDataURI(path string) (uri, contentType string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	contentType = http.DetectContentType(data)
	encoded := base64.StdEncoding.EncodeToString(data)
	uri = fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
	return uri, contentType, nil
}
```

Nano Banana Pro は入力画像からアスペクト比を自動で推測できるため、比率を計算して注入するような複雑なロジックは不要です。ピクセルデータをそのまま渡し、モデルに処理を任せるだけで機能します。

## 実行方法

家族の歴史をつなぐ大切な思い出の写真が色あせてしまっているなら、ぜひ試してみてください。時間の経過に埋もれかけていた瞬間を取り戻し、本来あるべき鮮明さを蘇らせる素晴らしい手段になります。

1. **リポジトリのクローン**:
    ```bash
    git clone https://github.com/danicat/glowup
    cd glowup
    ```

2. **認証情報の設定**（ロケーションは `global` を指定します）:
    ```bash
    export GOOGLE_CLOUD_PROJECT=your-project-id
    export GOOGLE_CLOUD_LOCATION=global
    ```

3. **修復の実行**:
    ```bash
    go run main.go restore --file old_photo.jpg
    ```

## 既知の課題と制限事項

復元処理は十分に機能しますが、いくつかの注意点や癖もあります。開発中に見つかった課題を共有します。

* **指示の遵守精度（Instruction Adherence）:** Nano Banana Pro は最先端モデルですが、プロンプトの指示を完全には反映しきれないことがたまにあります。意図通りの結果を得るまでに数回試行が必要になる場合があります。今回はプロンプトの細かなチューニングにそこまで時間をかけていないため、プロンプトエンジニアリングによってさらに改善できる余地があります。
* **Dev UI でのモデル一覧表示:** `googlegenai` プラグインに、Dev UI 上で利用可能なモデル一覧が自動展開されない不具合があります。モデル名を手動指定して「動的」に参照することは可能ですが、試行錯誤の段階で少し手間がかかります（JS 版では問題なく動作します）。すでに [Issue を報告済み](https://github.com/firebase/genkit/issues/4783)で修正もマージされていますが、古いバージョンを利用している場合はご注意ください。

## おわりに

GlowUp の構築は、AI を使って自分の過去や家族の記憶と感情的につながり直す、非常に有意義な実験となりました。AI に対しては悲観論も少なくありませんが、私自身がそもそも AI にワクワクしていたのは、まさにこうした心動かされるものづくりができるからでした。

この記事で使用した写真は、この技術の可能性を示すほんの一例に過ぎません。すでに第2弾の記事も執筆中であり、そこではさらに応用して、子供の頃に大好きだった思い出のカードゲームを復元するプロジェクトに挑戦しています。

要するに、その可能性は無限大です。技術的な課題であれ個人的なテーマであれ、この記事が皆さん自身の身近な課題に向き合い、解決するためのツールを作るきっかけになれば幸いです。

**自分でも作ってみたいですか？** この写真復元ツールをゼロから作れる[ステップ・バイ・ステップの Codelab](https://codelabs.developers.google.com/cloud-genkit-go-nano-banana?hl=en#0) を公開しています。

詳細については、[Genkit ドキュメント](https://firebase.google.com/docs/genkit) や [GlowUp のソースコード](https://github.com/danicat/glowup) もぜひご覧ください。

**Happy coding!**

Dani =^.^=
