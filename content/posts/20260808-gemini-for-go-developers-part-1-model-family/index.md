---
title: "Gemini for Go Developers - Part 1: The Gemini Model Family"
date: 2026-08-08
summary: "The first chapter of the Gemini for Go Developers series, focusing on the different models of the Gemini model family and the API surfaces to consume them."
tags: ["golang", "gemini"]
categories: ["AI & Development"]
heroStyle: "big"
---

Welcome to **Gemini for Go Developers**! This series is your complete guide to building AI-powered software in Go. Across seven hands-on chapters, we will cover everything from agentic coding, to building autonomous agents with **Genkit** and **ADK**, developing games, and using the full **G3 Stack** (Go, Gemini, GCP) to deploy applications to the cloud.

In Chapter 1, we lay the foundation by exploring the Gemini model family, model configurations, and writing our first code with the official [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai).

## The Gemini model family

We often treat "Gemini" as a single name for Google AI offerings, much like using "Google" as a shorthand for search. In reality, Gemini is a family of distinct models built for different operational trade-offs.

While frontier models are the ones that capture the headlines, knowing when to use smaller or specialised models is essential for cost-effective engineering. Model selection also directly [influences user experience](https://services.google.com/fh/files/blogs/google_delayexp.pdf) and product adoption, as latency varies sharply across model tiers.

Reaching for a Gemini Pro model with high thinking levels for every task is tempting, but it isn't always the right move. In many cases, it simply increases per-request latency and API costs without delivering a better outcome.

### Model naming scheme

To navigate the Gemini catalog, it helps to understand how Google names its models. A standard model string follows this pattern:

{{< katex >}}
\[
\text{[family]}-\text{[version]}-\text{[tier]}{-\text{[modifier]}}
\]

For example: `gemini-3.6-flash` or `gemini-3-pro-image`.

* **Family**: While most models will be in the Gemini family, Google also has other model families like Veo and Lyria.
* **Version Numbers**: Represent generational leaps in intelligence, context window handling, and instruction adherence.
* **Model Tiers**:
  * **Pro**: Designed for complex multi-step reasoning.
  * **Flash**: Balanced model with bias towards speed.
  * **Flash-Lite**: Optimized for speed and high-throughput, simple tasks.
* **Modifiers**: Might indicate a sub-family or specialisation, like `image` in `gemini-3.1-flash-image` or `live` in `gemini-3.1-flash-live-preview`. It may also include lifecycle modifiers like `-preview` or `-exp` (for experimental).

### Model overview

Here is an overview of key Gemini models, starting with the frontier model Gemini 3.x:

#### Gemini 3.x

Gemini 3.x is the primary frontier model line, available in Pro, Flash, and Flash-Lite tiers. These general-purpose models are also the primary choice for code generation and software engineering tasks.

Current models include:
- `gemini-3.6-flash`: High-speed workhorse for multimodal reasoning and agentic tasks
- `gemini-3.5-flash-lite`: Lowest-cost, ultra-fast tier for high-throughput microservices
- `gemini-3.1-pro-preview`: Advanced tier for complex multi-step reasoning and deep codebase analysis

#### Gemini image models (Nano Banana)

While technically still part of the Gemini family, this is a specialised model for image generation, providing both multimodal input and output (image and text). It is capable of producing images from scratch and performing edits on existing images.

Current models include:
- `gemini-2.5-flash-image` (aka Nano Banana)
- `gemini-3-pro-image` (aka Nano Banana Pro)
- `gemini-3.1-flash-image` (aka Nano Banana 2)
- `gemini-3.1-flash-lite-image` (aka Nano Banana 2 Lite)

#### Veo

A model specialised in [video generation with native audio](https://ai.google.dev/gemini-api/docs/veo). Videos are generated based on text prompts and key images to mark transitions (start and end frame) and as references. Veo 3.1 generates clips up to 8 seconds, but it is possible to extend them up to 20 times in 7-second increments.

Current models:
- `veo-3.1-generate-preview`
- `veo-3.1-lite-generate-preview` (fast generation)

#### Lyria

[Lyria](https://ai.google.dev/gemini-api/docs/music-generation) is specialised in music generation, delivering both instrumental and vocal compositions. Lyria accepts both text and images as input, with the images serving as inspiration for the composition. You can also provide the lyrics yourself or let the model create them for you.

Current models:
- `lyria-3-pro-preview`
- `lyria-3-clip-preview` (short 30s clips)

#### Gemma

[Gemma](https://ai.google.dev/gemma/docs) is an open-weights model family from Google. It is trained with the same technology behind Gemini, but designed to be deployed on your own infrastructure. Beyond the models offered by Google, Gemma also has a [strong community](https://deepmind.google/models/gemma/gemmaverse/) that produces fine-tuned versions for all sorts of use cases.

Some Gemma models are small enough to be suitable for running on local machines, enabling use cases where network connectivity is limited or non-existent. The bigger models are very capable, enabling use cases where sovereignty and network isolation are required.

#### Notable mentions

- Live models: While earlier models handle batch or request-response jobs, Google also offers live models for real-time streaming. Look for `-live` in the name (e.g. `gemini-3.1-flash-live-preview`).
- Text-to-speech: Generates speech from text with narration control using audio tags (`gemini-3.1-flash-tts-preview`).
- Computer use: A model that can "see" the screen and automate browser tasks (`gemini-2.5-computer-use-preview-10-2025`).

As you can see, Gemini is much more than a single model. It's a complete suite covering everything from basic chatbots to multimodal creation and agentic capabilities.

For detailed specifications on each model, see the [official Gemini models documentation](https://ai.google.dev/gemini-api/docs/models).

## Additional capabilities

Beyond standard text completion, Gemini models support additional features for building complex apps. Here are a few of the most important ones.

### Thinking

Gemini 2.5 and later models use an internal reasoning process that significantly improves multi-step planning, logic, coding, and mathematical capabilities. Before generating the final response, the model reasons internally by generating "thinking tokens" to analyze edge cases and plan multi-step strategies.

Thinking is a feature that can be controlled using the configuration parameters `thinking budget` in 2.5 and `thinking level` in 3.x. The higher budget or thinking level, the more time and tokens the model will spend during the reasoning phase.

When thinking is active, total billable output tokens include both the generated output text and the model's generated thinking tokens. Adjusting the thinking level based on task complexity is a critical step for production services.

### Built-in tools and function calling

Function calling allows Gemini models to interface with external tools, APIs, and databases. Gemini supports both built-in tools (like `google_search` and `code_execution`) and custom functions defined at application level.

Function calling has three primary use cases:
- **Take Actions:** Interact with external systems via APIs, such as scheduling meetings, sending emails, creating invoices, or controlling smart home devices.
- **Augment Knowledge:** Fetch real-time or private information from external databases, microservices, and knowledge bases.
- **Extend Capabilities:** Perform precise math, data conversion, or chart generation that exceeds LLM limits.

#### How function calling works

Function calling follows a 4-step execution process between your application and the model:

1. **Declare tools**: Define function declarations (name, clear description, and parameter JSON Schemas) and pass them in the request configuration.
2. **Model identifies tool intent**: The model inspects the prompt and tool declarations. If a tool is required, it returns a structured tool call intent containing the function name and arguments.
3. **Execute function code**: The model *does not* execute code itself. Your application receives the function call request, executes the corresponding local logic, and captures the result.
4. **Return function result**: Send the execution output back to the model as a function result step. The model uses this data to generate its final natural language response or decide if additional tool calls are needed.

### Structured outputs

You can configure Gemini models to generate responses that adhere to a provided [JSON Schema](https://ai.google.dev/gemini-api/docs/structured-output). This simplifies extracting structured data from text, eliminating fragile parsing when converting model responses into data structures.

Besides writing raw JSON Schemas in REST payloads, the Google GenAI SDKs allow developers to define schemas using native language constructs such as [Pydantic](https://docs.pydantic.dev/) in Python and struct tags in Go.

## Consuming models programmatically

With the model ecosystem and capabilities covered, let's look at how to incorporate these APIs into Go applications.

Just as different models exist for different use cases, several API surfaces are available. Let's start with the most basic one: Generate Content.

### Generate Content API

This is the most basic [generative API](https://ai.google.dev/api/generate-content#method:-models.generatecontent). It is a stateless interface that accepts a single request and returns a response. For multi-turn conversations, your application must send the full chat history with each call.

This requires actively managing conversation history to stay within context window limits. Applications typically summarize history once it reaches a threshold. To reduce input costs on long sessions, the Gemini API supports [implicit caching](https://ai.google.dev/gemini-api/docs/caching) for all models since Gemini 2.5, as well as [explicit caching](https://ai.google.dev/gemini-api/docs/generate-content/caching) for heavy payloads.

While the Generate Content API is good for simple stateless generations, it is being gradually replaced by the newer and more capable Interactions API.

### Interactions API

> Note: As of today, the Interactions API is not yet supported by the official Go GenAI SDK. Implementation progress is being tracked in this [GitHub issue](https://github.com/googleapis/go-genai/issues/658).

The Interactions API is Google's unified interface designed for all tasks, from simple chat and tool use to complex agentic workflows. It can manage conversation history server-side so your application doesn't have to.

### Live API

The Live API enables real-time, two-way voice and video conversations over WebSockets. It automatically detects when a user speaks or interrupts, making voice interactions feel natural while supporting tools like web search and function calling directly within the live session.

### Batch API

The Batch API lets you process large volumes of data asynchronously at half price. Jobs run in the background during off-peak hours (usually completing within 24 hours), making it ideal for non-urgent workloads.

### Managed Agents API

Managed agents offer a fully hosted runtime environment where AI agents plan and execute tasks autonomously. A single API call provisions an OS-isolated Linux sandbox with pre-installed runtimes like Python and Node, enabling the agent to run code, manage files, and browse the web.

Google provides two pre-built managed agents out of the box:
- **Antigravity Agent** (`antigravity-preview-05-2026`): The default general-purpose agent powered by Gemini 3.6 Flash (configurable to Gemini 3.5 Flash or Flash-Lite) for code execution, file management, and web access.
- **Deep Research Agent** (`deep-research-preview-04-2026`): An autonomous research agent that queries multi-source web data and compiles detailed research reports in the background.

You can also extend the Antigravity agent by defining system rules inline or mounting an `AGENTS.md` file, attaching structured skill directories (`SKILL.md`), or mounting local files, Cloud Storage buckets, and Git repositories directly into the remote workspace (`/workspace`).

## Access and billing

When integrating Gemini, Google offers two primary access and billing modes depending on your needs:

1. **Google AI Studio (Google AI)**: Routes requests through the Gemini API using an API key (`GEMINI_API_KEY` or `GOOGLE_API_KEY`). Best for prototyping, personal projects, indie applications, and fast developer onboarding.
2. **Gemini Enterprise (formerly Vertex AI)**: Routes requests through Google Cloud endpoints using Google Cloud IAM, Application Default Credentials (ADC), service account keys, or OAuth 2.0 user tokens. Best for production enterprise workloads requiring strict data privacy, security compliance, SLAs, GCP resource management, and committed-use discounts.

## Go GenAI SDK

Now let's see how this works in Go code.

The official SDK for integrating Gemini in Go applications is [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai). 

Sometimes you will see it being called the "unified" SDK as it is designed to support all Google models and both Google AI and Gemini Enterprise authentication. It replaces the legacy `github.com/google/generative-ai-go` package which has been deprecated.

Install it with `go get`:

```bash
go get google.golang.org/genai
```

Here is an example using Gemini Enterprise authentication and billing:

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

You can run this program using the following:

```sh
export GOOGLE_CLOUD_PROJECT="your-project-id-goes-here"
go run main.go
```

Here is the result:

![Generated wizard cat image output in Go terminal](image.png "The true purpose of AI: infinite cat picture generation")

While this is just a simple example to show how to work with the SDK, throughout this series we will see more examples of both the Go GenAI SDK and higher-level frameworks like [Genkit](https://genkit.dev/) and [Agent Development Kit (ADK)](https://adk.dev/).

## What's next?

In **Part 2** of the **Gemini for Go Developers** series, we'll dive deep into coding agents and how to prepare your environment for working in Go codebases. Stay tuned!