---
categories:
  - Agent Development
date: 2026-08-25
heroStyle: big
series:
  - Gemini for Go Developers
series_order: 3
summary: "Learn how to build autonomous agents in Go using the Go GenAI SDK, Genkit, and ADK through a practical Retro Game Appraiser example."
tags:
  - adk
  - gemini
  - genkit
  - golang
title: "Gemini for Go Developers: Building Agents in Go"
slug: "gemini-for-go-developers-part-3-building-agents"
aliases:
  - "/posts/20260825-gemini-for-go-developers-part-3-building-agents/"
  - "/posts/20260826-gemini-for-go-developers-part-3-building-agents/"
description: "Part 3 of Gemini for Go Developers: build a Retro Game Appraiser agent using the Go GenAI SDK, Genkit, and Google ADK before deploying to Cloud Run."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "github.com/firebase/genkit/go"
  - "google.golang.org/adk/v2"
  - "google.golang.org/genai"
---

Welcome back to **Gemini for Go Developers**! In [Part 1: The Gemini Model Family]({{< ref "/posts/20260808-gemini-for-go-developers-part-1-model-family" >}}), we looked at Gemini's capabilities across different model tiers and, in [Part 2: Coding with Gemini]({{< ref "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini" >}}), we explored how to configure our coding agents for Go development.

Now it is time to turn the tables and explore the other side of the equation: how to build AI-enabled applications and autonomous agents in Go. In this chapter, we will dissect the fundamental mechanics of an agent, define a concrete agent domain — a **Retro Game Appraiser** — and build it step by step across three distinct Go paradigms:

1. A low-level agent loop built directly with the **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)**.
2. A structured, flow-centric pipeline built with **[Genkit](https://genkit.dev)**.
3. A modular, session-driven multi-agent system built with Google's **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)**.

Finally, we will review the operational runtimes available for deploying Go agents reliably to the cloud.

All runnable examples and complete source code for this article are available on GitHub in the companion repository: [**danicat/gemini-for-go-developers**](https://github.com/danicat/gemini-for-go-developers/tree/main/part-3).

## The anatomy of an agent

The word *agent* is often thrown around loosely, but in modern AI engineering it has a precise architectural definition: an autonomous system composed of a **large language model**, one or more **tools** (callable functions or APIs), and an **execution harness** operating inside a feedback loop.

Without tools, a model is merely a text generator or chatbot. It can only generate responses based on its static training weights or the context provided in the prompt. To grant a model *agency* — the capacity to inspect external state, verify hypotheses, and execute changes in the real world — we must connect it to executable tools.

An essential architectural detail is that the large language model never executes code or external APIs directly. Instead, the local **agent harness** (our Go program) acts as the intermediary. When the model determines that it needs external information or an action performed, it yields a structured *tool call request*. The harness executes the corresponding Go function, captures the output, and feeds the result back to the model turn. The model then evaluates the new context and either decides to call another tool or generates its final answer for the user.

{{< mermaid >}}
sequenceDiagram
    autonumber
    actor User
    participant Harness as Agent Harness (Go)
    participant Model as Gemini LLM (with Search Grounding)
    participant Catalog as Local Catalog Database

    User->>Harness: Prompt: "I found EarthBound on SNES for $350. Is it a good deal?"
    Harness->>Model: Request (System Prompt + search_catalog Tool + Google Search Grounding)
    Model-->>Harness: Tool Call: search_catalog(query="EarthBound")
    Harness->>Catalog: Query local collection
    Catalog-->>Harness: Returns: owned=true, condition="Loose Cartridge", price_paid=$180
    Harness->>Model: Tool Response: {owned: true, format: "Loose", paid: 180}
    Note over Model: Gemini executes Google Search Grounding server-side for market pricing
    Model-->>Harness: Final Natural Language Response (Grounded with Search citations)
    Harness->>User: "You already own a Loose copy. At $350 for a Mint CIB copy, this is an outstanding deal..."
{{< /mermaid >}}

Not all tools require local execution. Gemini supports built-in tools like **Google Search grounding**, which execute server-side on Google infrastructure. When declared in your request configuration, the API transparently resolves search queries and injects grounding metadata into the model's response without requiring a local network roundtrip. For custom domain logic, however, your Go harness is fully responsible for schema declaration, dispatching arguments, and returning serialized results.

## Agent design: the Retro Game Appraiser

To compare our three implementation options directly, we will build the exact same agent in each stack: the **Retro Game Appraiser**.

If you collect retro video games like I do, this is a very familiar pain point. Whenever I visit retro game markets, more than once I have ended up buying a copy of a game I love only to get home and realise I already have it in my collection. On top of that, retro game collecting is notorious for volatile market prices, multiple condition variations (Loose cartridge vs. Complete in Box / CIB vs. Factory Sealed), and frequent counterfeit listings. 

During a recent market trip, I also realised that double-checking asking prices with Gemini in real time was surprisingly productive: I'm always happy to pay a reasonable premium for the in-person buying experience and supporting local vendors, but nobody wants to get scammed or massively overpay for a counterfeit cart.

Our appraiser agent solves both sides of this equation: cross-referencing our personal inventory while evaluating fair market values against live web data.

### Capabilities and user interaction

The collector interacts with the agent with natural questions such as:

* *"Do I have Chrono Trigger in my collection?"*
* *"I found a copy of EarthBound for SNES in mint Complete-in-Box condition for $350. Do I already own it, and is this a good price compared to current market values?"*
* *"What did I pay for Castlevania: Symphony of the Night, and has its market value gone up?"*

### Tool contracts

To answer these questions accurately without hallucinating inventory or prices, our agent relies on two distinct sources of truth:

1. **`search_catalog` (Local Tool):** A client-side Go function that queries the collector's local database. It matches keywords against titles and platforms, returning whether the item is already owned, its physical condition, acquisition date, and original purchase price.
2. **`google_search` (Search Grounding):** A server-side tool that searches online marketplaces and auction trackers to gather current fair market values, recent verified sales, and condition benchmarks.

### Reasoning strategy

When asked about a potential purchase, the agent follows a systematic multi-step workflow:
1. Inspect the local catalog to check whether the game is already in the collection and in what condition.
2. Query Google Search to determine current market prices for the specific platform and condition tier.
3. Synthesize the findings into an actionable appraisal: compare the asking price against market averages, highlight upgrade opportunities (for example, replacing a loose cartridge with a boxed copy), and provide a clear recommendation.

Let's begin by implementing this agent using the bare Go GenAI SDK.

## Implementing the agent with the Go GenAI SDK

Building an agent directly with the **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)** (`google.golang.org/genai`) represents the **lowest abstraction level**, mapping 1:1 directly to the Gemini API wire protocol. Because there are no framework abstractions between your code and the API, you manage the tool dispatch loop, conversation history, and loop termination criteria explicitly.

This low-level approach is ideal for small projects, utility scripts, learning the core mechanics of function calling and grounding, or when you have hyper-specific custom loop requirements that don't fit standard framework patterns. Because it compiles to a single, self-contained Go binary with no framework dependencies, it can be deployed to any typical web service or API hosting platform (such as Google Cloud Run, Kubernetes, or a virtual machine).

Here is the complete, runnable implementation:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Initialise GenAI client for Gemini Enterprise / Vertex AI
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendVertexAI,
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

	// 3. Prepare initial user prompt
	prompt := "I found a copy of EarthBound for SNES in mint Complete-in-Box (CIB) condition for $350. " +
		"Do I already own it, and is $350 a good deal compared to current market prices?"

	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText(prompt)},
		},
	}

	model := "gemini-3.7-flash"
	maxTurns := 6

	// 4. The Agent Loop (LLM Request -> Tool Dispatch -> Tool Result -> LLM Request)
	for turn := 0; turn < maxTurns; turn++ {
		resp, err := client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			log.Fatalf("error generating content: %v", err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			log.Fatal("received empty response candidate from model")
		}

		// Append the model's response to the conversation history
		modelContent := resp.Candidates[0].Content
		contents = append(contents, modelContent)

		// Check if the model requested any client-side tool executions
		funcCalls := resp.FunctionCalls()
		if len(funcCalls) == 0 {
			// Terminal condition: model produced its final natural language answer
			fmt.Println("\n=== Appraiser Verdict ===")
			fmt.Println(resp.Text())
			return
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

	log.Fatal("exceeded maximum conversation turns without reaching terminal state")
}
```

### Running the SDK agent

To run this example, ensure you have configured your Google Cloud project and authenticated with Application Default Credentials (`gcloud auth application-default login`):

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
go run main.go
```

The output demonstrates the harness coordinating between local catalog execution and server-side search grounding:

```text
[Harness] Executing tool: search_catalog(args=map[query:EarthBound])

=== Appraiser Verdict ===
Here is your collection check and appraisal for **EarthBound (SNES)**:

1. **Current Collection Status**:
   - You currently own **EarthBound** on Super Nintendo as a **Loose Cartridge**, purchased for **$180.00**.

2. **Market Price Appraisal**:
   - Verified market sales for an authentic, **Complete-in-Box (CIB)** copy of EarthBound typically range between **$1,200.00 and $1,500.00** depending on the condition of the box, tray, and original player's guide.

3. **Recommendation**:
   - At **$350.00**, a genuine Mint CIB copy is an **exceptional deal** (more than 70% below prevailing market value).
   - **Caution**: Because EarthBound is one of the most heavily counterfeited SNES titles, inspect the box printing, registration card, and PCB board carefully before completing the transaction. If verified authentic, this is an outstanding opportunity to upgrade your loose copy to CIB.
```

The SDK implementation makes the mechanical control flow explicit and straightforward. Today, with state-of-the-art models and generous token windows, it is deceptively easy to write the plumbing, the tool dispatch loops, and all the supporting scaffolding yourself.

However, just because you *can* build all the plumbing from scratch doesn't mean you *should*. Every single line of custom framework code you write becomes an ongoing maintenance burden and a potential source of subtle bugs. Writing your own agent framework is at best a distraction from shipping features, and at worst a massive pile of technical debt that diverts effort away from the actual business outcome you are aiming to achieve. The best code is no-code; the second best code is the one that solves your problem with the least amount of custom code.

## Agent development frameworks

While the raw SDK loop is unbeatable for learning low-level mechanics or hyper-specific bespoke loops, production applications demand higher levels of abstraction:

* **Automatic Schema Reflection:** Defining JSON schemas manually with `&genai.Schema{...}` is tedious and error-prone. Production frameworks infer schemas directly from native Go structs and doc comments.
* **Observability & Distributed Tracing:** In production, you need OpenTelemetry traces, latency breakdowns per tool invocation, and token consumption metrics out of the box.
* **Prompt Management:** Hardcoding prompts in Go strings hinders collaboration with prompt engineers and prevents versioning prompt templates independently of binary releases.
* **Session Persistence & State:** Managing multi-turn conversation histories across stateless HTTP requests requires thread-safe, decoupled storage backends.
* **Model Portability:** While the Go SDK is Gemini-specific, frameworks allow you to switch model providers or test local models without rewriting your business logic.

To address these needs across different architectural requirements, we have two primary open-source frameworks in the Go ecosystem: **Genkit** and **Agent Development Kit (ADK)**.

| Dimension | Go GenAI SDK | Genkit Go | Agent Development Kit (ADK) |
| :--- | :--- | :--- | :--- |
| **Abstraction Level** | **Low** (1:1 Gemini API mapping) | **Medium** (Structured workflows) | **High** (Autonomous multi-agent systems) |
| **Core Architecture** | Explicit `for` loop & dispatch | **Flows** (`genkit.DefineFlow`) & Tools | **Agents**, Runners & Session Services |
| **Sweet Spot & Use Cases** | Small projects, learning mechanics, bespoke low-level loops | One-off GenAI apps (CLIs, web services), deterministic pipelines, single-domain agents | Conversational chat agents, multi-agent orchestration, complex memory & RAG |
| **Model Ecosystem** | Gemini-specific | Multi-model (Google GenAI, Vertex AI, Ollama, etc.) | Multi-model (via ADK model adapters) |
| **Target Deployment** | Any web service / API host (Cloud Run, K8s, VMs) | Any backend platform; **Cloud Run** (preferred) | **Gemini Enterprise** (preferred for sessions/RAG/memory) or **Cloud Run** |

Let's examine how our Retro Game Appraiser is implemented in each framework.

## Implementing the agent with Genkit

**[Genkit](https://genkit.dev)** operates at a **medium abstraction level**, bringing software engineering discipline and structured observability to AI applications. In Genkit, everything is organised around **Flows** (strongly-typed, observable pipelines) and **Tools** (type-safe Go functions with automatic schema generation).

Genkit is the ideal choice for applications with **one-off generation** (such as CLI tools, batch processing, and webhook endpoints), deterministic workflows, and single-domain agents. It natively supports multiple model providers through plugins, and because a Genkit app is simply a standard Go HTTP server, it can run on any backend platform — with **Google Cloud Run** being the preferred deployment target for seamless container hosting and automatic scaling.

Here is the Retro Game Appraiser implemented with Genkit Go:

```go
package main

import (
	"context"
	"encoding/json"
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
	Prompt string `json:"prompt"`
}

type AppraiserResponse struct {
	Appraisal string `json:"appraisal"`
}

func main() {
	ctx := context.Background()

	// 1. Initialise Genkit with Vertex AI / Gemini Enterprise plugin
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.VertexAI{
			ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
			Location:  "global",
		}),
		genkit.WithDefaultModel("vertexai/gemini-3.7-flash"),
	)

	// 2. Define strongly-typed tool with automatic schema generation
	catalogTool := genkit.DefineTool(
		g,
		"search_catalog",
		"Search the collector's personal inventory for owned games by title or platform.",
		func(ctx *ai.ToolContext, req CatalogRequest) (CatalogResponse, error) {
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
		},
	)

	// 3. Define structured appraisal flow
	appraiserFlow := genkit.DefineFlow(
		g,
		"appraise_game",
		func(ctx context.Context, input string) (string, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithSystem(
					"You are an expert Retro Game Appraiser. Assist collectors by evaluating prospective purchases, "+
						"cross-referencing their personal inventory, and assessing fair market valuations. "+
						"Always search the collection catalog using search_catalog before providing purchase recommendations.",
				),
				ai.WithPrompt(input),
				ai.WithTools(catalogTool),
			)
			if err != nil {
				return "", fmt.Errorf("appraisal generation failed: %w", err)
			}
			return resp.Text(), nil
		},
	)

	// 4. HTTP API Endpoint with graceful shutdown
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/appraise", func(w http.ResponseWriter, req *http.Request) {
		var body AppraiserRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		result, err := appraiserFlow.Run(req.Context(), body.Prompt)
		if err != nil {
			log.Printf("flow execution error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AppraiserResponse{Appraisal: result})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Handle graceful shutdown on SIGINT (Ctrl+C) and SIGTERM
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

### Running the Genkit flow

Start the Genkit server by providing your Google Cloud project:

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
export PORT=8080
go run main.go
```

In another terminal, send an appraisal request:

```sh
curl -s -X POST http://localhost:8080/api/appraise \
  -H "Content-Type: application/json" \
  -d '{"prompt": "I found a copy of EarthBound for SNES for $350. Do I own it, and is it a good deal?"}' | jq .
```

The flow executes, queries the catalog tool, and returns a structured appraisal payload:

```json
{
  "appraisal": "### 1. Catalog Check\n**Yes, you already own it.**\n* **Title:** *EarthBound* (SNES, 1994)\n* **Status in Collection:** Loose Cartridge\n* **Condition/Notes:** Authentic board verified; label in excellent shape.\n* **Price Paid:** $180\n\n---\n\n### 2. Market Appraisal & Deal Analysis\n* **Loose Cartridge:** The current going market rate for an authentic loose copy ranges between **$320 and $380**. At **$350**, it is priced right at **fair market value**—neither an overpriced listing nor a significant bargain.\n* **Complete in Box (CIB) / Boxed with Guide:** If this listing happens to include the original big box and strategy guide with scratch-and-sniff cards, $350 would be an extraordinary steal (CIB copies regularly sell for **$1,500–$2,500+**).\n\n---\n\n### 3. Recommendation\n* **Pass (if Loose):** Since you already have an authentic copy in excellent condition, paying retail market price ($350) for a duplicate loose cart does not offer strong value or upside.\n* **Buy immediately (if Complete/Boxed):** Only pull the trigger if it includes the original packaging or represents a major condition upgrade/variant.\n* **Buyer Beware:** If you do ever consider another copy, always inspect the PCB (printed circuit board) screws and chips, as *EarthBound* is one of the most frequently counterfeited games on the SNES."
}
```

Genkit eliminates the manual dispatch loop. It handles argument unmarshalling into native Go structs, executes the tool, feeds the payload back to the model, and automatically emits OpenTelemetry spans for every step. It is the sweet spot whenever your workload follows a structured, deterministic flow.

## Implementing the agent with Agent Development Kit (ADK)

While Genkit focuses on structured application pipelines, Google's **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)** (`google.golang.org/adk/v2`) operates at a **high abstraction level**, architected specifically for autonomous conversational agents, multi-agent orchestration, and complex architectures that demand long-term memory and enterprise RAG.

ADK standardises agent lifecycles, subagent delegation, and agent-to-agent (A2A) communication protocols. Like Genkit, it supports multiple model providers through modular model adapters.

When it comes to deployment, **Gemini Enterprise Agent Platform** is the preferred target if you are leveraging managed session persistence, enterprise grounding connectors, and memory services without writing custom database layers. If you are managing your own state stores or running containerised microservices, **Google Cloud Run** is the recommended choice.

Here is the Retro Game Appraiser implemented with ADK v2:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
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

type CatalogRequest struct {
	// Query is the game title or platform to search in the inventory.
	Query string `json:"query"`
}

type CatalogResponse struct {
	Found   bool       `json:"found"`
	Message string     `json:"message,omitempty"`
	Count   int        `json:"count,omitempty"`
	Results []GameItem `json:"results,omitempty"`
}

type AppraiserRequest struct {
	Prompt    string `json:"prompt"`
	SessionID string `json:"session_id,omitempty"`
}

type AppraiserResponse struct {
	Appraisal string `json:"appraisal"`
	SessionID string `json:"session_id,omitempty"`
}

func main() {
	ctx := context.Background()

	// 1. Initialise Gemini Model adapter for Gemini Enterprise / Vertex AI
	model, err := gemini.NewModel(ctx, "gemini-3.7-flash", &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendVertexAI,
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
		Tools: []tool.Tool{catalogTool},
	})
	if err != nil {
		log.Fatalf("failed to create appraiser agent: %v", err)
	}

	// 4. Initialise ADK Runner with Session Service for state management
	r, err := runner.New(runner.Config{
		AppName:           "retro_game_vault",
		Agent:             appraiserAgent,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		log.Fatalf("failed to create ADK runner: %v", err)
	}

	// 5. HTTP Handler streaming multi-turn responses with session continuation
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/chat", func(w http.ResponseWriter, req *http.Request) {
		var chatReq AppraiserRequest
		if err := json.NewDecoder(req.Body).Decode(&chatReq); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		userMsg := &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: chatReq.Prompt},
			},
		}

		sessionID := chatReq.SessionID
		if sessionID == "" {
			sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
		}

		var responseBuilder strings.Builder
		for event, err := range r.Run(req.Context(), "collector_user", sessionID, userMsg, agent.RunConfig{}) {
			if err != nil {
				log.Printf("runner error: %v", err)
				http.Error(w, fmt.Sprintf("agent error: %v", err), http.StatusInternalServerError)
				return
			}
			if event != nil && event.Content != nil {
				for _, part := range event.Content.Parts {
					if part.Text != "" {
						responseBuilder.WriteString(part.Text)
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AppraiserResponse{
			Appraisal: responseBuilder.String(),
			SessionID: sessionID,
		})
	})

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
		log.Printf("Retro Game Appraiser (ADK) listening on :%s", port)
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

### Running the ADK agent

Start the ADK streaming server with your Google Cloud project configured:

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
export PORT=8081
go run main.go
```

In another terminal, send an initial question to start a conversation:

```sh
curl -s -X POST http://localhost:8081/api/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Do I have Chrono Trigger in my collection?"}' | jq .
```

The ADK runner orchestrates session initialisation, executes the registered `search_catalog` function tool, and returns the response alongside the persisted `session_id`:

```json
{
  "appraisal": "Yes, you have **Chrono Trigger** in your collection! Here are the details from your inventory:\n\n* **Title:** Chrono Trigger\n* **Platform:** Super Nintendo (SNES)\n* **Release Year:** 1995\n* **Condition:** CIB (Complete in Box)\n* **Price Paid:** $210.00\n* **Notes:** Includes original map and registration card.",
  "session_id": "session-1787674771186589000"
}
```

Because ADK's `runner` maintains state via `SessionService`, you can continue the conversation simply by passing the returned `session_id` in subsequent turns:

```sh
curl -s -X POST http://localhost:8081/api/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "What did I pay for it?", "session_id": "session-1787674771186589000"}' | jq .
```

```json
{
  "appraisal": "You paid **$210.00** for it.",
  "session_id": "session-1787674771186589000"
}
```

ADK provides clean architectural separation between agent definition (`llmagent`), tool bindings (`functiontool`), and conversation state management (`runner` + `SessionService`), making it the framework of choice when building complex multi-agent hierarchies or conversational assistants.

## Agent runtimes

Writing your agent's code in Go is only half the journey. Once built, you need an execution environment that can handle long-lived streaming responses, background tool invocations, secure credentials, and traffic spikes without operational headaches.

Because Go compiles to single, self-contained binaries with virtually zero runtime overhead, Go agents are exceptionally fast and cost-effective to host. Depending on your framework choice and state requirements, you have two primary deployment targets:

### Cloud Run: the universal backend sweet spot

**[Google Cloud Run](https://cloud.google.com/run)** is the prime deployment target for hosting containerised Go agents — whether built directly with the Go GenAI SDK, Genkit flows, or ADK runners.

Using multi-stage Docker builds, you can compile your Go agent into a minimal `scratch` or `distroless` container image with zero runtime dependencies. Because the Go binary has no heavy interpreter or virtual machine to bootstrap, your service scales from zero rapidly, minimizing the cold start overhead that often plagues heavier runtimes.

Key architectural and operational advantages of Cloud Run for Go agents include:

* **True Scale-to-Zero & Fast Concurrency:** Pay only for the exact CPU milliseconds consumed while your agent is actively executing or processing tool responses. Go's lightweight goroutines allow a single container instance to concurrently process hundreds of active agent turns with negligible memory footprint.
* **Extended Request Timeouts (Up to 60 Minutes):** While default timeouts are 5 minutes, Cloud Run supports request timeouts up to 60 minutes (3,600 seconds) — giving multi-step reasoning loops, deep research tasks, and iterative subagent swarms ample headroom to finish without premature termination.
* **Bidirectional Streaming & WebSockets:** Cloud Run natively supports HTTP/2 chunked transfer encoding, Server-Sent Events (SSE), and WebSockets. For voice-to-voice or real-time multimodal agents interacting with the Gemini Live API, WebSockets allow continuous, bidirectional streaming between client and server.
* **Session Affinity (Sticky Sessions):** If your agent maintains temporary in-memory caches or local context across consecutive conversational turns, you can enable session affinity (`--session-affinity`) via client IP or cookies to consistently route follow-up requests to the same container instance.
* **Stateless Resilience:** For production deployments, state should be decoupled into external managed datastores (such as Firestore, Redis, or Cloud SQL). If a network disconnect occurs during a long-running turn, the agent can resume execution via its interaction ID without losing context.
* **Secretless Workload Identity IAM:** Zero hardcoded API keys or credentials. Your Go agent authenticates seamlessly to Gemini Developer API, Vertex AI, and Cloud storage using the ambient Cloud Run service account.

### Gemini Enterprise Agent Platform: managed sessions and enterprise RAG

When building with **Agent Development Kit (ADK)**, especially for conversational chat agents or enterprise solutions requiring persistent sessions, managed long-term memory, and enterprise knowledge grounding, the **Gemini Enterprise Agent Platform** (the evolution of Vertex AI Agent Engine) provides a fully managed, serverless agent execution environment.

Instead of provisioning databases and writing custom session storage adapters yourself, the platform offers managed agent infrastructure with:

* **Decoupled Session Persistence:** Seamless architectural separation between persistent long-term conversation storage (`SessionService` in ADK) and ephemeral streaming execution loops (`LiveSession` with automatic session resumption handles across reconnections).
* **Enterprise Grounding & Vector Search:** Turnkey connectors across enterprise knowledge bases (Google Drive, BigQuery, intranet repositories) alongside Vertex AI Vector Search indexes (featuring storage-optimised tiers) for high-performance semantic retrieval.
* **Sandboxed Code Execution:** Secure, isolated sandbox environments where agents can dynamically write and execute code (e.g. data analysis, Python/Go scripts) without exposing your host infrastructure to risk.
* **Agent-to-Agent (A2A) Protocols:** Standardised communication protocols that allow independent enterprise agents to discover capabilities, negotiate schemas, and delegate tasks to one another across organisational boundaries.
* **Security, Agent Identity & Model Armor:** Granular agent IAM permissions, VPC Service Controls perimeters, and Model Armor runtime protection to inspect inputs and outputs against prompt injection, data exfiltration, and policy violations.

## What's next?

It is difficult to do justice to all of these frameworks or runtime platforms in a single article, but don't worry — across the remainder of this series we will explore them in depth:

* **Part 4**: Deep dive into **Genkit for Go** — dotprompt templates, custom plugins, streaming, and observability with the Dev UI.
* **Part 5**: Deep dive into **Agent Development Kit (ADK)** — building autonomous multi-agent hierarchies, subagent delegation, and session state.
* **Part 6**: Deploying Go agents to **Cloud Run** and the **Gemini Enterprise Agent Platform** with production-grade CI/CD and IAM.
* **Part 7**: Shifting gears into game development in Go with **[Ebitengine](https://ebitengine.org/)**.

Stay tuned, and happy hacking!