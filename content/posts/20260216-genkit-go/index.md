---
title: "Bringing Old Photos to Life with Genkit and Gemini 3"
date: 2026-02-16
summary: "Learn how to build a high-fidelity photo restoration tool using Go, Genkit, and Nano Banana Pro (Gemini 3 Pro Image) native 4K capabilities."
tags: ["genkit", "golang", "nano-banana", "gemini", "tutorial"]
categories: ["AI & Development"]
heroStyle: "big"
---

As part of my job I get to know all sorts of people and one very common topic of conversations is about my heritage. Not only I have an obviously sounding Eastern European surname but I also look the part, so people often assume I'm either Polish or Czech. Lots of people are surprised when I say that in reality I'm originally from Brazil.

Nobody really knows where my family came from, as we were really **terrible** at keeping track of historical records. Maybe because we are self-conscious of this, we often have conversations in the family about the part of history that we know and how it is also fading. As we are all aging, memories are the first to go, and then goes the documents and pictures. There is an inherent feeling of sadness when you realise that the last time you saw your grandmother was 30 years ago, and that her face is often no more than a blur. This is why photographs are so important to me, as they are the stronghold to fight the degradation of my own memories.

Anything made in the recent years could be easily duplicated and stored as many redundant copies in the cloud as I wish, but we are talking about mementos from before the digital age. Even if I had scanned them years ago, many of them have already gone through decades of dust, mold, wear and tear. They are frozen in time, but not getting any better.

Thanks to the evolution of generative AI, not all is lost and I can finally give these photos a breath of fresh air, not only restoring the damages due to the passage of time, but also colorising and upscaling them to bring them up to modern standards. This is how a small software called "GlowUp" was born.


Below is one example of such restoration:

![](original.jpg "Original: my grandmother preparing her world-famous banana pie")

![](restored.png "Restored: restoration and colorisation by Nano Banana Pro")

In this article, I'm going to show how to build GlowUp from scratch using [Gemini Nano Banana Pro](https://ai.google.dev/gemini-api/docs/image-generation) and [Genkit Go](https://genkit.dev/docs/get-started/?lang=go).

## The building blocks

I've chosen to use Nano Banana Pro (aka. Gemini 3 Pro Image Preview) because it is currently the most advanced image processing model in the Gemini family. While the regular Nano Banana (Gemini 2.5 Flash Image) is also a great model, I find the pro version has better quality outputs and it is also better at following instructions, even if it takes a bit of trial and error.

On the client side, instead of going for a low level SDK like [go-genai](https://pkg.go.dev/google.golang.org/genai), I decided to use Genkit as it provides a few quality of life improvements over the lower level code such as:

- Model agnostic: I can test different models if I want, even non-Google or local ones, with a single plugin replacement
- Out of the box Dev UI support for conveniences like testing models, prompts and tracing of model calls
- Easy to convert from prompt to CLI application to web server (if I decide to go that route)

For the first version of GlowUp, I'm making it available exclusively as a command line tool, but having the flexibility of deploying it as a server will allow me to package this code into a nice app that even my father could use to restore his collection of photos without my intervention.

## A first look at Genkit Go

[Genkit](https://firebase.google.com/docs/genkit) is an open-source framework designed to bring production standards to AI development. If you are a Go developer*, think of it as the "standard library" for AI features. _(* And if you are **not** a Go developer, check out the docs as Genkit also supports JS and Python.)_

Here is what a minimal "Hello World" looks like in Genkit for Go. Note how we use the `googlegenai` plugin to initialize the framework.

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

This short snippet is doing a lot of heavy lifting. Let's look at it a bit more carefully.


### Plugins
Adapters that connect your code to providers like Vertex AI, Google AI, or Ollama. For Google models, we should use the `googlegenai` plugin. It supports both backends:

* **Google AI (Studio):** Uses an API Key. Best for prototyping and personal projects.
```go
// Use Google AI (API Key)
googlegenai.Init(ctx, &googlegenai.Config{APIKey: "MY_KEY"})
```

* **Vertex AI (Google Cloud):** Uses Google Cloud IAM authentication. Recommended for production workloads and enterprise features.
```go
// Use Vertex AI (Cloud Auth)
googlegenai.Init(ctx, &googlegenai.VertexAI{ProjectID: "my-project", Location: "us-central1"})
```

**Note:** If you are migrating from older versions of Genkit, you might be familiar with separate `vertexai` and `googleai` plugins. These have been consolidated into the single `googlegenai` plugin.

### Models
The actual LLMs (e.g., Gemini, Claude) that generate content. You reference them by name strings like `vertexai/gemini-2.5-flash`.

```go
	resp, err := genkit.GenerateText(ctx, g, 
		ai.WithModel("vertexai/gemini-2.5-flash"),
		ai.WithTextPrompt("Tell me a joke"))
```

### Prompts

While nothing prevents you from hardcoding prompts, like in the example above, it is a good practice to keep them in separate files for better maintainability. Genkit uses [dotprompt](https://github.com/google/dotprompt) to load external prompts. 

A `dotprompt` file (*.prompt) consists of two main parts: the **Frontmatter** and the **Template**.

**1. Frontmatter (Configuration)**
* **`model`**: The model identifier (e.g., `vertexai/gemini-2.5-flash`).
* **`config`**: Generation parameters like `temperature`, `topK`, or model-specific settings (e.g., `imageConfig`).
* **`input`**: A JSON schema defining the variables expected from your Go code.
* **`output`**: For structured outputs.

**2. Template (Instructions)**
The body uses Handlebars syntax to construct the prompt:
* **Variables**: Placeholders like `{{theme}}` are replaced by the values defined in your input schema.
* **Roles**: The `{{role "system"}}` and `{{role "user"}}` helpers structure the conversation, separating system instructions from user queries.
* **Media**: The `{{media url=myImage}}` helper injects multimodal data (images, video) directly into the model context.

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
In Genkit, a **Flow** is the fundamental unit of execution that provides:
1.  **Observability**: Every flow execution automatically generates traces and metrics (latency, token usage, success rate) viewable in the Genkit Developer UI or Google Cloud Trace.
2.  **Type Safety**: Flows are strictly typed with input and output schemas, preventing runtime errors when chaining multiple AI operations.
3.  **Deployability**: Flows are strictly separated from their serving logic. To deploy them, wrap them with `genkit.Handler`, which converts a flow into a standard `http.Handler`. This makes it straightforward to serve them using the standard library or any Go web framework:

```go
    // Define a flow
    myFlow := genkit.DefineFlow(g, "myFlow", func(ctx context.Context, input string) (string, error) {
        return "Processed: " + input, nil
    })

    // Expose it as an HTTP handler
    http.HandleFunc("/myFlow", genkit.Handler(myFlow))
```


## Nano Banana Pro

The engine behind our restoration is **Gemini 3 Pro Image**, affectionately (and internally) known as "Nano Banana Pro".

It represents a significant leap over previous generations (and even the current "Flash" models). While Gemini 2.5 Flash is incredibly fast and capable of basic image generation (`gemini-2.5-flash-image`), **Nano Banana Pro** (`gemini-3-pro-image-preview`) is built for deep multimodal reasoning.

It doesn't just "see" pixels; it understands semantic context. It can differentiate between a "scratch on the paper" and a "scar on a face". It knows that a 1950s kitchen likely has linoleum floors, not modern hardwood.

### Key differences

*   **Flash (gemini-2.5-flash-image)**: Optimized for speed and cost. Great for thumbnails or simple illustrations. Max resolution 1024x1024.
*   **Pro (gemini-3-pro-image-preview)**: Optimized for fidelity and reasoning. Supports native **4K resolution** generation (up to 4096px), which is non-negotiable for photo restoration.

The model also accepts `imageConfig` parameters to fine-tune the output:
*   `imageSize`: "4K" or "2K".
*   `aspectRatio`: "16:9", "4:3", "1:1", etc.

One important detail to note is that this model always returns interleaved responses containing both text and images. Unlike other generation models, image-only output is not supported. This is why our extraction logic (which we'll see below) needs to be flexible enough to find the image data within the multi-part response message.

**Note:** At the time of writing, this model is only available in the `global` location on Vertex AI. You must configure your Vertex AI client accordingly.


## Connecting the parts

Now, let's look at how GlowUp connects these pieces. We use a **prompt file** to define the restoration expert persona and a **flow** to handle the image processing.

### The prompt

We use a `.prompt` file to define our model configuration and instructions. Notice how we enforce the `4K` resolution here, keeping our code clean.

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

### The flow

The Go code is remarkably simple. It acts as the orchestrator: preparing the image data, loading the prompt, and executing the request.

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

The `PrepareImageData` helper (which you can find in the [source code](https://github.com/danicat/glowup/blob/main/internal/image.go)) is responsible for normalizing the various input types — whether it's a local file path, a remote URL, or a raw Base64 string — into a standard Data URI that the Gemini API expects.

Note how we use `http.DetectContentType` to dynamically determine the MIME type, rather than assuming everything is a JPEG. This is critical for maintaining fidelity across different scan formats.

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


Because Nano Banana Pro is smart enough to infer the aspect ratio from the input image, we don't need complex logic to calculate and inject it. We provide the pixels and let the model do its job.


## How to run it

If you have a collection of fading memories that serve as fragile anchors to your family history, I encourage you to try this. It’s a way to reclaim those moments from time and give them the clarity they deserve.

1.  **Clone the Repository**:



    ```bash
    git clone https://github.com/danicat/glowup
    cd glowup
    ```

2.  **Set Up Credentials** (Remember: `global` location!):
    ```bash
    export GOOGLE_CLOUD_PROJECT=your-project-id
    export GOOGLE_CLOUD_LOCATION=global
    ```

3.  **Run the Restoration**:
    ```bash
    go run main.go --file old_photo.jpg
    ```

## Known issues and limitations

While the results can be stunning, GlowUp isn't without its quirks. Here are a few things to keep in mind if you are planning to build upon this:

*   **Instruction adherence:** Even though Nano Banana Pro is a vanguard model, it still occasionally misses an instruction. You might find it requires a few attempts before you get the desired result. I haven't spent much time fine-tuning the prompt, so there are likely opportunities for further optimization there.
*   **Models in the Dev UI:** The Dev UI doesn't automatically populate the available models when using Genkit Go, which adds a bit of friction to the experimentation process (the JS version does this just fine). This is [a known bug](https://github.com/firebase/genkit/issues/4783) currently being tracked by the team.


## Conclusions

Building GlowUp was a satisfying experiment in using AI to solve a very specific, personal problem. It is a reminder that while we often focus on the massive, general-purpose applications of LLMs, their true strength often lies in these bespoke, small-scale tools that we can build for ourselves in an afternoon.

The ability to take a fading family memory and give it more clarity is a practical win, but the real takeaway is how much the barrier to entry for building such tools has fallen. I hope this inspires you to look at your own niche problems — technical or personal — and see what you can build to solve them.

For more details, check out the [Genkit documentation](https://firebase.google.com/docs/genkit) and the [GlowUp source code](https://github.com/danicat/glowup).

**Happy coding!**

Dani =^.^=


