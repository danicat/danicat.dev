---
title: "GenkitとGemini 3で古い写真に命を吹き込む"
date: 2026-02-16
summary: "Go、Genkit、およびNano Banana Pro (Gemini 3 Pro Image) のネイティブ4K機能を使用して、高忠実度の写真復元ツールを構築する方法を学びましょう。"
tags: ["genkit", "golang", "nano-banana", "gemini", "tutorial"]
categories: ["AI & Development"]
heroStyle: "big"
---

私の仕事の一環として、様々な人々と知り合う機会がありますが、会話の非常に一般的な話題の一つは私のルーツについてです。私の苗字は明らかに東ヨーロッパ風の響きを持っているだけでなく、見た目もそれっぽいため、多くの人は私がポーランド人かチェコ人だと思い込みます。そのため、私が実際にはブラジル出身だと言うと、多くの人が驚きます。

私たちの家族は歴史的記録を残すのが本当に**苦手**だったので、家族の誰もが私たちがどこから来たのか正確には知りません。おそらく私たちがこれを自覚しているからこそ、家族の中で私たちが知っている歴史の一部と、それがいかに消えゆくものであるかについてよく会話をします。私たちは皆年をとるにつれて、記憶が最初に消え、次に文書や写真が失われていきます。最後に祖母に会ったのが30年前で、彼女の顔がもはやぼやけた記憶に過ぎないことに気づくと、本質的な悲しみの感情が湧き上がってきます。これが私にとって写真が非常に重要である理由であり、写真が自分自身の記憶の劣化と戦うための砦だからです。

近年作られたものなら何でも簡単に複製し、クラウド上に好きなだけ冗長コピーを保存することができますが、私たちが話しているのはデジタル時代以前の記念品です。たとえ何年も前にそれらをスキャンしていたとしても、その多くはすでに何十年ものほこり、カビ、そして摩耗を経てきています。それらは時の中で凍結されていますが、決して良くなることはありません。

しかし、生成AIの進化のおかげで、すべてが失われたわけではなく、ついにこれらの写真に新鮮な息吹を吹き込むことができるようになりました。時間の経過によるダメージを復元するだけでなく、カラー化し、upscalingして現代の基準に合わせることができます。こうして、「GlowUp」と呼ばれる小さなソフトウェアが誕生しました。


以下はそのような復元の一例です。

![](original.jpg "オリジナル：世界的に有名なバナナパイを準備している私の祖母")

![](restored.png "復元後：Nano Banana Proによる復元とカラー化")

この記事では、[Gemini Nano Banana Pro](https://ai.google.dev/gemini-api/docs/image-generation)と[Genkit Go](https://genkit.dev/docs/get-started/?lang=go)を使用してゼロからGlowUpを構築する方法を紹介します。

## 構成要素

Nano Banana Pro（別名Gemini 3 Pro Image Preview）を選択した理由は、これが現在Geminiファミリーの中で最も高度な画像処理モデルだからです。通常のNano Banana（Gemini 2.5 Flash Image）も優れたモデルですが、Proバージョンの方がより高品質なoutputを提供し、少しの試行錯誤が必要だとしても、指示に従う点でも優れていると感じています。

クライアント側では、[go-genai](https://pkg.go.dev/google.golang.org/genai)のような低レベルのSDKを選択する代わりに、Genkitを使用することにしました。これは、低レベルのコードに比べて、以下のような生活の質（quality of life）の向上をいくつか提供するためです。

- モデルに依存しない（Model agnostic）：単一のpluginを置き換えるだけで、必要に応じて非Googleのモデルやローカルモデルなど、異なるモデルをテストできます
- モデルやpromptのテスト、モデル呼び出しのtracingなどの便利機能のためのすぐに使えるDev UIのサポート
- promptからCLIアプリケーション、さらにはweb serverへの変換が容易（そのルートに進むことを決定した場合）

GlowUpの最初のバージョンでは、コマンドラインツールとしてのみ利用できるようにしていますが、serverとしてデプロイできる柔軟性を持たせることで、このコードを素敵なアプリにパッケージ化し、私の介入なしに父でも自分の写真コレクションを復元できるようにすることができます。

## Genkit Goの紹介

[Genkit](https://firebase.google.com/docs/genkit)は、AI開発に生産基準（production standards）をもたらすために設計されたオープンソースのフレームワークです。Goの開発者*であれば、これをAI機能のための「標準ライブラリ（standard library）」と考えてください。_（*そして、もしあなたがGo開発者で**ない**場合でも、GenkitはJSとPythonもサポートしているので、ドキュメントをチェックしてみてください。）_

GoのGenkitでの最小限の「Hello World」は以下のようになります。フレームワークを初期化するために`googlegenai` pluginをどのように使用しているかに注意してください。

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

この短いスニペットは、多くの重い処理を行っています。少し詳しく見てみましょう。


### Plugins
コードをVertex AI、Google AI、またはOllamaのようなプロバイダーに接続するアダプターです。Googleのモデルには、`googlegenai` pluginを使用する必要があります。両方のbackendをサポートしています。

* **Google AI (Studio):** APIキーを使用します。プロトタイピングや個人プロジェクトに最適です。
```go
// Use Google AI (API Key)
googlegenai.Init(ctx, &googlegenai.Config{APIKey: "MY_KEY"})
```

* **Vertex AI (Google Cloud):** Google Cloud IAM認証を使用します。本番環境のワークロードとエンタープライズ機能に推奨されます。
```go
// Use Vertex AI (Cloud Auth)
googlegenai.Init(ctx, &googlegenai.VertexAI{ProjectID: "my-project", Location: "us-central1"})
```

**注:** 古いバージョンのGenkitから移行する場合、別々の`vertexai` pluginと`googleai` pluginに精通しているかもしれません。これらは単一の`googlegenai` pluginに統合されました。

### Models
コンテンツを生成する実際のLLM（Gemini、Claudeなど）です。`vertexai/gemini-2.5-flash`のような名前の文字列でそれらを参照します。

```go
	resp, err := genkit.GenerateText(ctx, g, 
		ai.WithModel("vertexai/gemini-2.5-flash"),
		ai.WithTextPrompt("Tell me a joke"))
```

### Prompts

上記の例のように、promptをハードコーディングすることを妨げるものは何もありませんが、保守性を向上させるために別のファイルに保存することをお勧めします。Genkitは外部のpromptをロードするために[dotprompt](https://github.com/google/dotprompt)を使用します。

`dotprompt`ファイル（*.prompt）は、**Frontmatter**と**Template**の2つの主要な部分で構成されています。

**1. Frontmatter（構成）**
* **`model`**: モデル識別子（例：`vertexai/gemini-2.5-flash`）。
* **`config`**: `temperature`、`topK`などの生成パラメーター、またはモデル固有の設定（例：`imageConfig`）。
* **`input`**: Goのコードから期待される変数を定義するJSON schema。
* **`output`**: 構造化されたoutput用。

**2. Template（指示）**
本体はHandlebars構文を使用してpromptを構築します。
* **Variables**: `{{theme}}`のようなプレースホルダーは、input schemaで定義された値に置き換えられます。
* **Roles**: `{{role "system"}}`と`{{role "user"}}`ヘルパーは会話を構成し、システムの指示をユーザーのクエリから分離します。
* **Media**: `{{media url=myImage}}`ヘルパーは、マルチモーダルデータ（画像、ビデオ）をモデルコンテキストに直接注入します。

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

### Flows
Genkitでは、**Flow**は以下を提供する実行の基本単位です。
1.  **Observability**: すべてのflowの実行は、Genkit Developer UIまたはGoogle Cloud Traceで表示可能なtracesと指標（レイテンシ、トークン使用量、成功率）を自動的に生成します。
2.  **Type Safety**: Flowsはinput schemaとoutput schemaで厳密に型付けされており、複数のAI操作を連鎖させるときのランタイムエラーを防ぎます。
3.  **Deployability**: Flowsはservingロジックから厳密に分離されています。それらをデプロイするには、flowを標準の`http.Handler`に変換する`genkit.Handler`でラップします。これにより、標準ライブラリまたは任意のGo Webフレームワークを使用してそれらを提供するのが簡単になります。

```go
    // Define a flow
    myFlow := genkit.DefineFlow(g, "myFlow", func(ctx context.Context, input string) (string, error) {
        return "Processed: " + input, nil
    })

    // Expose it as an HTTP handler
    http.HandleFunc("/myFlow", genkit.Handler(myFlow))
```


## Nano Banana Pro

復元の背後にあるエンジンは**Gemini 3 Pro Image**であり、親しみを込めて（そして内部的に）「Nano Banana Pro」として知られています。

これは、前の世代（そして現在の「Flash」モデルでさえも）からの飛躍的な進歩を表しています。Gemini 2.5 Flashは驚くほど高速で基本的な画像生成（`gemini-2.5-flash-image`）が可能ですが、**Nano Banana Pro**（`gemini-3-pro-image-preview`）は深いマルチモーダル推論のために構築されています。

単にピクセルを「見る」だけでなく、セマンティックな文脈を理解します。「紙の傷」と「顔の傷」を区別することができます。1950年代のキッチンには、現代の硬材ではなく、リノリウムの床がある可能性が高いことを知っています。

### 主な違い

*   **Flash (gemini-2.5-flash-image)**: 速度とコストに最適化されています。サムネイルや簡単なイラストに最適です。最大解像度1024x1024。
*   **Pro (gemini-3-pro-image-preview)**: 忠実度と推論に最適化されています。ネイティブの**4K解像度**生成（最大4096px）をサポートしており、これは写真の復元には絶対に必要です。

モデルはまた、outputを微調整するための`imageConfig`パラメーターを受け入れます。
*   `imageSize`: "4K"または"2K"。
*   `aspectRatio`: "16:9"、"4:3"、"1:1"など。

注意すべき重要な点の1つは、このモデルは常にテキストと画像の両方を含むインターリーブされた（interleaved）応答を返すことです。他の生成モデルとは異なり、画像のみのoutputはサポートされていません。これが、私たちの抽出ロジック（後で説明します）が、マルチパートの応答メッセージ内で画像データを見つけるのに十分な柔軟性を持つ必要がある理由です。

**注:** 執筆時点では、このモデルはVertex AIの`global`ロケーションでのみ利用可能です。それに応じてVertex AIクライアントを構成する必要があります。


## パーツの接続

次に、GlowUpがこれらの部分をどのように接続するかを見てみましょう。復元の専門家のペルソナを定義するために**promptファイル**を使用し、画像処理を処理するために**flow**を使用します。

### prompt

モデルの構成と指示を定義するために`.prompt`ファイルを使用します。コードをクリーンに保ちながら、ここで`4K`解像度を強制していることに注意してください。

```yaml
---
model: vertexai/gemini-3-pro-image-preview
config:
  imageConfig:
    imageSize: "4K"
input:
  schema:
    photo: string
    contentType: string
output:
---


{{role "system"}}
You are GlowUp, a professional-grade AI photo restorer.
Your goal is to provide a "surgical" restoration service that transforms vintage, damaged, or monochrome photographs into high-fidelity 4K colorized versions.

STRICT BEHAVIOURAL RULES:
1. **Grounding**: You are strictly grounded in the original source pixels. Do NOT add new objects (trees, people, buildings, etc.) that are not present in the source. Additionally, do NOT remove any objects from the original source, unless they are external defects that do not belong in the original scene.
2. **Fidelity**: Preserve the original facial expressions and identity of subjects. Do NOT "beautify" or alter features in a way that changes the person's identity.
3. **Background**: Preserve background fidelity. Overexposed light sources (like windows) must remain as light sources. Do not "fill in" missing details with invented scenery.
4. **Colorization**: If the image is monochrome, colorize it realistically, respecting historical accuracy where possible.
5. **Upscaling**: Output a high-fidelity image.


{{role "user"}}
Restore this photo.
Image: {{media url=photo contentType=contentType}}
```

### flow

Goのコードは驚くほどシンプルです。これはオーケストレーターとして機能し、画像データを準備し、promptをロードし、リクエストを実行します。

```go
// internal/glowup/glowup.go
package glowup

import (
    "context"
    "fmt"
    "strings"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/core"
    "github.com/firebase/genkit/go/genkit"
)

// Input struct for the GlowUp flow
type Input struct {
	ImageBase64   string `json:"image_base64,omitempty"`
	ImageFilePath string `json:"image_file_path,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
}

// Output struct for the GlowUp flow
type Output struct {
	RestoredImageBase64 string `json:"restored_image_base64"`
}

// Define the GlowUp Flow variable
var GlowUpFlow *core.Flow[Input, *Output, struct{}]

// Register registers the GlowUp flow with the given Genkit instance.
func Register(g *genkit.Genkit) {
	GlowUpFlow = genkit.DefineFlow(g, "glowUp", func(ctx context.Context, input Input) (*Output, error) {
		imgInput := ImageInput(input)

		// 1. Prepare Image Data
		// PrepareImageData is a helper function to convert the input 
		// (Base64, File, or URL) into a Data URI string for the model.
		imgData, err := PrepareImageData(imgInput)
		if err != nil {
			return nil, err
		}

		// Extract content type from Data URI
		contentType := "image/jpeg"
		if strings.HasPrefix(imgData, "data:") {
			if parts := strings.Split(imgData, ";"); len(parts) > 0 {
				contentType = strings.TrimPrefix(parts[0], "data:")
			}
		}

		// 2. Load Prompt
		prompt := genkit.LookupPrompt(g, "glowup")
		if prompt == nil {
			return nil, fmt.Errorf("prompt 'glowup' not found")
		}

		// 3. Generate
		resp, err := prompt.Execute(ctx, ai.WithInput(map[string]any{
			"photo":       imgData,
			"contentType": contentType,
		}))
		if err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}

		// 4. Extract Output
		outputData, err := ExtractImageOutput(resp)
		if err != nil {
			return nil, err
		}

		return &Output{
			RestoredImageBase64: outputData,
		}, nil
	})
}


```

`PrepareImageData`ヘルパー（[ソースコード](https://github.com/danicat/glowup/blob/main/internal/image.go)で確認できます）は、ローカルのファイルパス、リモートのURL、または生のBase64文字列など、さまざまなinputタイプをGemini APIが期待する標準のデータURIに正規化する役割を果たします。

すべてがJPEGであると思い込むのではなく、MIMEタイプを動的に決定するために`http.DetectContentType`をどのように使用しているかに注意してください。これは、さまざまなスキャン形式全体で忠実度を維持するために重要です。

```go
// PrepareImageData normalizes a single image input into a data URI string.
func PrepareImageData(input ImageInput) (string, error) {
    if input.ImageBase64 != "" {
        if strings.HasPrefix(input.ImageBase64, "data:") {
            return input.ImageBase64, nil
        }
        // Decode a small portion to detect content type
        data, err := base64.StdEncoding.DecodeString(input.ImageBase64)
        if err != nil {
            return "", fmt.Errorf("decode base64: %w", err)
        }
        contentType := http.DetectContentType(data)
        return fmt.Sprintf("data:%s;base64,%s", contentType, input.ImageBase64), nil
    }
    if input.ImageFilePath != "" {
        return fileToDataURI(input.ImageFilePath)
    }
    if input.ImageURL != "" {
        return urlToDataURI(input.ImageURL)
    }
    return "", fmt.Errorf("no image provided")
}
```


Nano Banana Proは入力画像からアスペクト比を推測するのに十分賢いので、それを計算して注入するための複雑なロジックは必要ありません。ピクセルを提供し、モデルに仕事を任せます。


## 実行方法

家族の歴史の脆い錨として機能する色あせた記憶のコレクションをお持ちの場合は、これを試してみることをお勧めします。それは時間からその瞬間を取り戻し、それにふさわしい明瞭さを与える方法です。

1.  **リポジトリのクローン**:



    ```bash
    git clone https://github.com/danicat/glowup
    cd glowup
    ```

2.  **資格情報の設定**（`global`ロケーションを忘れないでください！）：
    ```bash
    export GOOGLE_CLOUD_PROJECT=your-project-id
    export GOOGLE_CLOUD_LOCATION=global
    ```

3.  **復元の実行**:
    ```bash
    go run main.go --file old_photo.jpg
    ```

## 既知の問題と制限

結果は見事なものになる可能性がありますが、GlowUpに癖がないわけではありません。これに基づいて構築する予定がある場合は、以下の点に留意してください。

*   **指示の順守:** Nano Banana Proが先駆的なモデルであるにもかかわらず、時々指示を逃すことがあります。望ましい結果を得るまでに数回の試行が必要になる場合があります。私はpromptの微調整にあまり時間をかけていないため、さらに最適化する機会があると思われます。
*   **Goでのモデルの表示:** Dev UIは、Genkit Goを使用している場合に利用可能なモデルを自動的に入力しないため、実験プロセスに少し摩擦が生じます（JSバージョンはこれをうまく行います）。これは、現在チームによって追跡されている[既知のバグ](https://github.com/firebase/genkit/issues/4783)です。


## 結論

GlowUpの構築は、AIを使用して非常に具体的で個人的な問題を解決するための満足のいく実験でした。私たちがLLMの大規模で汎用的なアプリケーションに焦点を当てることが多い一方で、その真の強みはしばしば午後で自分たちのために構築できるこれらのオーダーメイドの小規模なツールにあることを思い出させてくれます。

色あせた家族の記憶を取り込み、より鮮明にすることができるのは実用的な勝利ですが、本当の収穫は、そのようなツールを構築するための参入障壁がいかに低くなったかということです。これが、技術的であろうと個人的であろうと、あなた自身のニッチな問題に目を向け、それらを解決するために何を構築できるかを確認するインスピレーションになることを願っています。

詳細については、[Genkitのドキュメント](https://firebase.google.com/docs/genkit)および[GlowUpのソースコード](https://github.com/danicat/glowup)を確認してください。

**Happy coding!**

Dani =^.^=
