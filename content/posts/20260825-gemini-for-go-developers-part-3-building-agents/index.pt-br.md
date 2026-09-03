---
categories:
  - Agent Development
date: 2026-08-25
heroStyle: big
series:
  - Gemini for Go Developers
series_order: 3
summary: "Aprenda a construir agentes autônomos em Go usando o Go GenAI SDK, Genkit e ADK através de um exemplo prático de Avaliador de Jogos Retrô."
tags:
  - adk
  - gemini
  - genkit
  - golang
title: "Gemini para Desenvolvedores Go: Construindo Agentes em Go"
slug: "gemini-for-go-developers-part-3-building-agents"
aliases:
  - "/pt-br/posts/20260825-gemini-for-go-developers-part-3-building-agents/"
  - "/pt-br/posts/20260826-gemini-for-go-developers-part-3-building-agents/"
description: "Parte 3 de Gemini para Desenvolvedores Go: construa um agente Avaliador de Jogos Retrô com o Go GenAI SDK, Genkit e Google ADK antes de publicar no Cloud Run."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "github.com/firebase/genkit/go"
  - "google.golang.org/adk/v2"
  - "google.golang.org/genai"
---

Boas-vindas de volta à série **Gemini para Desenvolvedores Go**! Na [Parte 1: A Família de Modelos Gemini]({{< ref "/posts/20260808-gemini-for-go-developers-part-1-model-family" >}}), conhecemos as capacidades do Gemini entre diferentes categorias de modelos e, na [Parte 2: Programando com o Gemini]({{< ref "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini" >}}), exploramos como configurar nossos agentes de programação para o desenvolvimento em Go.

Agora é hora de inverter os papéis e explorar o outro lado da equação: como construir aplicações habilitadas para IA e agentes autônomos em Go. Neste capítulo, vamos dissecar a mecânica fundamental de um agente, definir um domínio concreto — um **Avaliador de Jogos Retrô** — e construí-lo passo a passo em três paradigmas distintos em Go:

1. Um loop de agente de baixo nível construído diretamente com o **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)**.
2. Um pipeline estruturado e orientado a fluxos (*flows*) construído com o **[Genkit](https://genkit.dev)**.
3. Um sistema multiagente modular orientado a sessões construído com o **[Agent Development Kit (ADK)](https://adk.dev)** do Google.

Por fim, analisaremos os ambientes de execução (*runtimes*) disponíveis para publicar e hospedar agentes Go em produção na nuvem com alta confiabilidade.

O código-fonte completo e executável de todos os exemplos deste artigo está disponível no GitHub no repositório complementar: [**danicat/gemini-for-go-developers**](https://github.com/danicat/gemini-for-go-developers/tree/main/part-3).

## A anatomia de um agente

A palavra *agente* é frequentemente usada de forma vaga, mas na engenharia moderna de IA ela possui uma definição arquitetural precisa: um sistema autônomo composto por um **modelo de linguagem (LLM)**, uma ou mais **ferramentas** (funções ou APIs executáveis) e um **harness de execução** operando dentro de um loop de feedback.

Sem ferramentas, um modelo é apenas um gerador de texto ou chatbot. Ele só consegue produzir respostas com base nos seus pesos estáticos de treinamento ou no contexto fornecido no prompt. Para conceder ao modelo *agência* — a capacidade de inspecionar estados externos, verificar hipóteses e executar ações no mundo real —, precisamos conectá-lo a ferramentas executáveis.

Um detalhe arquitetural crucial é que o modelo de linguagem nunca executa código ou APIs externas diretamente. Em vez disso, o **harness local** (nosso programa em Go) atua como intermediário. Quando o modelo determina que precisa de uma informação externa ou de uma ação executada, ele emite uma requisição estruturada de *chamada de função* (*tool call*). O harness executa a função Go correspondente, captura a saída e injeta o resultado de volta no turno do modelo. O modelo então avalia o novo contexto e decide se precisa chamar outra ferramenta ou se já possui dados suficientes para gerar a resposta final ao usuário.

{{< mermaid >}}
sequenceDiagram
    autonumber
    actor User as Usuário
    participant Harness as Harness do Agente (Go)
    participant Model as Gemini LLM (com Search Grounding)
    participant Catalog as Banco de Dados do Catálogo Local

    User->>Harness: Prompt: "Encontrei EarthBound de SNES por $350. É um bom negócio?"
    Harness->>Model: Requisição (Prompt de Sistema + Tool search_catalog + Google Search Grounding)
    Model-->>Harness: Tool Call: search_catalog(query="EarthBound")
    Harness->>Catalog: Consulta inventário local
    Catalog-->>Harness: Retorna: owned=true, condition="Cartucho Avulso", price_paid=$180
    Harness->>Model: Resposta da Tool: {owned: true, format: "Loose", paid: 180}
    Note over Model: Gemini executa o Google Search Grounding no servidor para obter preços de mercado
    Model-->>Harness: Resposta Final em Linguagem Natural (Com citações de busca)
    Harness->>User: "Você já possui uma cópia avulsa. A $350 por uma cópia Completa na Caixa (CIB), é um excelente negócio..."
{{< /mermaid >}}

Nem todas as ferramentas exigem execução local no seu cliente. O Gemini suporta ferramentas nativas integradas, como o **aterramento com a Busca do Google (Google Search grounding)**, executadas diretamente na infraestrutura do Google. Quando declaradas na configuração da sua requisição, a API resolve as buscas de mercado de forma transparente e injeta os metadados de fundamentação na resposta sem exigir idas e vindas na sua rede local. Para regras de negócio e dados privados, contudo, seu harness em Go é o único responsável pela declaração do schema, despacho de argumentos e retorno dos resultados serializados.

## Design do agente: o Avaliador de Jogos Retrô

Para comparar nossas três opções de implementação de forma direta, construiremos exatamente o mesmo agente em cada uma das stacks: o **Avaliador de Jogos Retrô** (*Retro Game Appraiser*).

Se você coleciona videogames retrô como eu, este é um problema muito familiar. Sempre que visito feiras e mercados de games, mais de uma vez acabei comprando uma cópia de um jogo que adoro só para chegar em casa e perceber que já o tinha na coleção. Para completar, o mercado de colecionismo retrô é famoso pela volatilidade de preços, múltiplas variações de estado de conservação (apenas o cartucho avulso vs. completo na caixa / CIB vs. lacrado de fábrica) e proliferação de cartuchos falsificados (*bootlegs*).

Durante uma visita recente a uma feira, percebi que checar os preços no Gemini em tempo real era incrivelmente produtivo: fico feliz em pagar um valor um pouco maior pela experiência presencial e pelo apoio aos vendedores locais, mas ninguém quer ser enganado ou pagar uma fortuna por uma reprodução não autêntica.

Nosso agente avaliador resolve os dois lados dessa equação: cruza o inventário pessoal do colecionador e analisa valores de mercado atualizados na web.

### Capacidades e interação do usuário

O colecionador conversa com o agente fazendo perguntas naturais como:

* *"Eu tenho Chrono Trigger na minha coleção?"*
* *"Encontrei uma cópia de EarthBound para Super Nintendo completa na caixa (CIB) em estado impecável por $350. Eu já tenho esse jogo, e esse preço está bom comparado ao mercado atual?"*
* *"Quanto eu paguei por Castlevania: Symphony of the Night, e o valor de mercado dele subiu?"*

### Contratos de ferramentas (*tool contracts*)

Para responder a essas dúvidas com precisão sem alucinar inventário ou cotações, o agente utiliza duas fontes de verdade:

1. **`search_catalog` (Ferramenta Local):** Função cliente em Go que consulta o banco de dados do colecionador. Ela busca palavras-chave em títulos e plataformas, informando se o item já foi adquirido, seu estado físico, data de aquisição e preço original pago.
2. **`google_search` (Aterramento de Busca):** Ferramenta no servidor que consulta lojas e rastreadores de leilões online para obter valores médios atuais, vendas verificadas recentes e referências de conservação.

### Estratégia de raciocínio

Ao ser questionado sobre uma oportunidade de compra, o agente segue um fluxo sistemático:
1. Inspeciona o catálogo local para verificar se o jogo já pertence à coleção e em que formato.
2. Consulta a Busca do Google para determinar a faixa de preço atual para a plataforma e condição física específicas.
3. Sintetiza os dados em uma avaliação prática: compara o valor anunciado com as médias de mercado, destaca oportunidades de upgrade (por exemplo, substituir um cartucho avulso por uma cópia completa com caixa e manual) e oferece uma recomendação clara.

Vamos começar implementando esse agente diretamente com o Go GenAI SDK.

## Implementando o agente com o Go GenAI SDK

Construir um agente diretamente com o **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)** (`google.golang.org/genai`) representa o **nível mais baixo de abstração**, mapeando 1:1 diretamente sobre o protocolo da API do Gemini. Como não existem camadas de framework entre seu código e a API, você gerencia o loop de despacho de ferramentas, o histórico da conversa e as condições de parada explicitamente.

Essa abordagem de baixo nível é excelente para scripts utilitários, pequenos experimentos, aprendizado da mecânica de *function calling* e aterramento, ou quando você precisa de um controle de loop hiperpersonalizado. Por compilar em um binário único e autossuficiente sem dependências pesadas, ele pode ser hospedado em qualquer plataforma padrão de microsserviços (como Google Cloud Run, Kubernetes ou máquinas virtuais).

Abaixo está a implementação completa e executável:

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

	model := "gemini-flash-latest"
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

### Executando o agente com o SDK

Para executar este exemplo, certifique-se de ter configurado seu projeto no Google Cloud e efetuado o login com Application Default Credentials (`gcloud auth application-default login`):

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
go run main.go
```

A sessão interativa de chat é iniciada diretamente no seu terminal. Você pode conversar com o avaliador ao longo de múltiplos turnos e encerrar a qualquer momento digitando `exit`, pressionando `Ctrl+C` ou enviando um `EOF` via `Ctrl+D`:

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

A implementação com o SDK torna o fluxo mecânico totalmente explícito. Hoje, com modelos de última geração e janelas generosas de tokens, é tentadoramente fácil escrever o encanamento, os loops de despacho e todo o andaime de suporte por conta própria.

No entanto, só porque você *consegue* construir todo o encanamento do zero, não significa que *deva*. Cada linha de código de framework customizado que você escreve se transforma em um custo contínuo de manutenção e em uma fonte potencial de bugs sutis de concorrência e serialização. Escrever seu próprio framework de agentes é, na melhor das hipóteses, uma distração que atrasa a entrega de valor e, na pior, uma montanha de débito técnico que desvia seu esforço do objetivo de negócio que você quer alcançar. O melhor código é aquele que não precisa ser escrito; o segundo melhor é aquele que resolve seu problema com a menor quantidade de código sob sua responsabilidade.

## Frameworks para desenvolvimento de agentes

Enquanto o loop com o SDK puro é imbatível para entender a mecânica básica ou criar loops altamente customizados, aplicações em produção exigem níveis superiores de abstração:

* **Inferência Automática de Schemas:** Definir schemas JSON manualmente com `&genai.Schema{...}` é repetitivo e suscetível a erros. Frameworks modernos inferem schemas diretamente de structs Go e comentários de documentação.
* **Observabilidade e Rastreamento Distribuído:** Em produção, você precisa de traces do OpenTelemetry, métricas de latência por ferramenta e contagem de tokens sem precisar instrumentar cada função manualmente.
* **Gerenciamento de Prompts:** Manter prompts fixos no código dificulta o trabalho colaborativo com engenheiros de prompt e impede o versionamento de templates separado do binário.
* **Persistência de Sessão e Estado:** Gerenciar históricos de conversas multi-turno entre requisições HTTP sem estado exige camadas de armazenamento desacopladas e seguras para concorrência.
* **Portabilidade de Modelos:** Enquanto o SDK Go é exclusivo para a família Gemini, frameworks permitem trocar provedores ou testar modelos locais sem reescrever a lógica de negócio.

Para atender a essas demandas em diferentes cenários arquiteturais, temos dois frameworks de código aberto no ecossistema Go: **Genkit** e **Agent Development Kit (ADK)**.

| Dimensão | Go GenAI SDK | Genkit Go | Agent Development Kit (ADK) |
| :--- | :--- | :--- | :--- |
| **Nível de Abstração** | **Baixo** (Mapeamento 1:1 com a API Gemini) | **Médio** (Workflows estruturados) | **Alto** (Sistemas multiagente autônomos) |
| **Arquitetura Central** | Loop `for` e despacho explícitos | **Flows** (`genkit.DefineFlow`) e Tools | **Agents**, Runners e Serviços de Sessão |
| **Ponto Ideal de Uso** | Scripts, aprendizado da mecânica, loops sob medida | Aplicações pontuais (CLIs, web services), pipelines determinísticos, agentes de domínio único | Agentes conversacionais, orquestração multiagente, memória de longo prazo e RAG corporativo |
| **Ecossistema de Modelos** | Específico para Gemini | Multimodelo (Google GenAI, Vertex AI, Ollama, etc.) | Multimodelo (via adaptadores do ADK) |
| **Destino Recomendado** | Qualquer host HTTP (Cloud Run, K8s, VMs) | Qualquer backend; **Cloud Run** (preferencial) | **Gemini Enterprise** (preferencial com sessões/RAG/memória) ou **Cloud Run** |

Vejamos como nosso Avaliador de Jogos Retrô é implementado em cada um desses frameworks.

## Implementando o agente com o Genkit

O **[Genkit](https://genkit.dev)** opera em um **nível médio de abstração**, trazendo disciplina de engenharia de software e observabilidade integrada para aplicações com IA. No Genkit, tudo é estruturado em torno de **Flows** (pipelines fortemente tipados e observáveis) e **Tools** (funções Go type-safe com geração automática de schemas).

O Genkit é a escolha perfeita para aplicações com **gerações pontuais** (como ferramentas de linha de comando, processamento em lote e webhooks), fluxos determinísticos e agentes focados em um único domínio. Ele suporta múltiplos provedores através de plugins e, por ser executado como um servidor HTTP padrão em Go, pode rodar em qualquer infraestrutura — sendo o **Google Cloud Run** o destino preferencial pela simplicidade no gerenciamento de contêineres e escalabilidade automática.

Abaixo está o Avaliador de Jogos Retrô construído com o Genkit Go:

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

### Executando o fluxo com o Genkit

Inicie o servidor Genkit exportando a variável do projeto Google Cloud:

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
export PORT=8080
go run main.go
```

Em outro terminal, envie uma solicitação de avaliação:

```sh
curl -s -X POST http://localhost:8080/api/appraise \
  -H "Content-Type: application/json" \
  -d '{"data": {"prompt": "I found a copy of EarthBound for SNES for $350. Do I own it, and is it a good deal?"}}' | jq .
```

O fluxo executa, consulta a ferramenta de catálogo e retorna o resultado estruturado:

```json
{
  "appraisal": "### 1. Catalog Check\n**Yes, you already own it.**\n* **Title:** *EarthBound* (SNES, 1994)\n* **Status in Collection:** Loose Cartridge\n* **Condition/Notes:** Authentic board verified; label in excellent shape.\n* **Price Paid:** $180\n\n---\n\n### 2. Market Appraisal & Deal Analysis\n* **Loose Cartridge:** The current going market rate for an authentic loose copy ranges between **$320 and $380**. At **$350**, it is priced right at **fair market value**—neither an overpriced listing nor a significant bargain.\n* **Complete in Box (CIB) / Boxed with Guide:** If this listing happens to include the original big box and strategy guide with scratch-and-sniff cards, $350 would be an extraordinary steal (CIB copies regularly sell for **$1,500–$2,500+**).\n\n---\n\n### 3. Recommendation\n* **Pass (if Loose):** Since you already have an authentic copy in excellent condition, paying retail market price ($350) for a duplicate loose cart does not offer strong value or upside.\n* **Buy immediately (if Complete/Boxed):** Only pull the trigger if it includes the original packaging or represents a major condition upgrade/variant.\n* **Buyer Beware:** If you do ever consider another copy, always inspect the PCB (printed circuit board) screws and chips, as *EarthBound* is one of the most frequently counterfeited games on the SNES."
}
```

O Genkit elimina a necessidade do loop manual. Ele cuida da desserialização dos argumentos em structs nativas de Go, executa a função, encaminha a resposta de volta ao modelo e gera spans de telemetria automaticamente para cada etapa. É a escolha ideal para fluxos determinísticos e estruturados.

## Implementando o agente com o Agent Development Kit (ADK)

Enquanto o Genkit se concentra em pipelines de aplicação estruturados, o **[Agent Development Kit (ADK)](https://adk.dev)** do Google opera em um **alto nível de abstração**, projetado especificamente para agentes conversacionais autônomos, orquestração multiagente e arquiteturas complexas que demandam memória persistente e RAG corporativo.

O ADK padroniza ciclos de vida de agentes, delegação para subagentes e protocolos de comunicação agente a agente (A2A). Assim como o Genkit, ele suporta múltiplos provedores através de adaptadores modulares de modelo.

Quanto ao ambiente de hospedagem, a **Gemini Enterprise Agent Platform** é o destino preferencial caso você deseje usufruir de persistência gerenciada de sessões, conectores corporativos de aterramento e serviços de memória sem precisar programar bancos de dados próprios. Caso prefira gerenciar seu próprio estado em microsserviços conteinerizados, o **Google Cloud Run** é a escolha ideal.

Abaixo está o Avaliador de Jogos Retrô implementado com o ADK v2:

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
	model, err := gemini.NewModel(ctx, "gemini-flash-latest", &genai.ClientConfig{
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

### Executando o agente com o ADK

O ADK disponibiliza um launcher universal (`full.NewLauncher()`) que elimina a necessidade de construir roteamento HTTP manual, serialização JSON ou boilerplate de gerenciamento de sessões. Passando argumentos pela linha de comando para o launcher, você pode executar o agente em múltiplos modos de interação sem alterar uma única linha de código.

Para iniciar uma sessão interativa no terminal, execute:

```sh
export GOOGLE_CLOUD_PROJECT="your-gcp-project-id"
go run main.go
```

O launcher inicia em modo console, permitindo conversar com o avaliador diretamente pelo terminal:

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

Como o ADK gerencia o estado da conversa e a continuidade da sessão automaticamente, perguntas subsequentes retêm todo o histórico contextual das rodadas anteriores.

Como alternativa, você pode iniciar o servidor com a Web UI de desenvolvedor e a API REST integradas do ADK:

```sh
go run main.go web webui api
```

Ao acessar `http://localhost:8080`, uma interface web interativa de chat é disponibilizada, oferecendo respostas em streaming em tempo real, histórico de sessões e visibilidade transparente da execução das ferramentas (tanto da tool customizada `search_catalog` quanto do aterramento nativo com `GoogleSearch`).

O ADK entrega uma separação arquitetural limpa entre a definição do agente (`llmagent`), as ferramentas (`functiontool` e ferramentas nativas como `geminitool.GoogleSearch`) e os ambientes de execução (`launcher` e `runner`), tornando-se a escolha ideal para hierarquias multiagente complexas e assistentes conversacionais.

## Runtimes para execução de agentes

Programar o código do agente em Go é apenas metade da jornada. Depois de pronto, você precisa de um ambiente de execução capaz de suportar respostas em streaming de longa duração, execuções de ferramentas em segundo plano, credenciais seguras e picos de tráfego sem complicações operacionais.

Como Go compila em binários únicos com consumo mínimo de recursos, os agentes em Go são extremamente rápidos e baratos de hospedar. Dependendo do framework e dos requisitos de estado, você conta com dois destinos principais:

### Cloud Run: o ponto ideal para backends conteinerizados

O **[Google Cloud Run](https://cloud.google.com/run)** é o principal destino para hospedar agentes Go em contêineres — sejam eles construídos diretamente com o Go GenAI SDK, fluxos do Genkit ou runners do ADK.

Utilizando builds em múltiplos estágios no Docker (*multi-stage builds*), você pode compilar seu agente Go em imagens mínimas `scratch` ou `distroless` sem dependências de tempo de execução. Sem a necessidade de inicializar interpretadores pesados ou máquinas virtuais, seu serviço escala de zero rapidamente, minimizando a sobrecarga de *cold start* que frequentemente afeta ambientes mais pesados.

Vantagens operacionais do Cloud Run para agentes Go:

* **Escala a Zero Real e Alta Concorrência:** Pague apenas pelos milissegundos de CPU consumidos durante a execução ativa do agente ou o processamento de respostas de ferramentas. As goroutines leves do Go permitem que uma única instância atenda simultaneamente centenas de turnos com consumo irrisório de memória.
* **Timeouts Estendidos (Até 60 Minutos):** Embora o timeout padrão seja de 5 minutos, o Cloud Run suporta limites de até 60 minutos (3.600 segundos) — garantindo que raciocínios em múltiplas etapas, tarefas profundas de pesquisa e enxames de subagentes terminem sem interrupções precoces.
* **Streaming Bidirecional e WebSockets:** Suporte nativo a HTTP/2 chunked transfer encoding, Server-Sent Events (SSE) e WebSockets. Para agentes multimodais de voz ou tempo real com a Gemini Live API, WebSockets viabilizam fluxos contínuos entre cliente e servidor.
* **Afinidade de Sessão (Sticky Sessions):** Se o seu agente mantém caches temporários em memória ou contexto local entre turnos consecutivos, é possível habilitar a afinidade de sessão (`--session-affinity`) via IP do cliente ou cookies para direcionar requisições à mesma instância.
* **Resiliência Desacoplada:** Para produção, o estado deve ser persistido em bancos gerenciados (como Firestore, Redis ou Cloud SQL). Se ocorrer uma queda de conexão em um turno longo, o agente pode retomar a execução pelo ID da interação sem perda de contexto.
* **Autenticação Segura com Workload Identity IAM:** Zero chaves de API fixas no código. O agente Go se autentica na Gemini Developer API, Vertex AI e Cloud Storage usando a conta de serviço ambiente do Cloud Run.

### Gemini Enterprise Agent Platform: sessões gerenciadas e RAG corporativo

Ao construir com o **Agent Development Kit (ADK)**, especialmente para soluções corporativas que exigem sessões duradouras, memória de longo prazo e aterramento em dados da empresa, a **Gemini Enterprise Agent Platform** (a evolução do Vertex AI Agent Engine) oferece um ecossistema serverless totalmente gerenciado.

Em vez de provisionar bancos de dados e escrever adaptadores manuais de persistência, a plataforma fornece:

* **Persistência Desacoplada de Sessão:** Separação limpa entre o armazenamento duradouro de histórico (`SessionService` no ADK) e os loops efêmeros de streaming (`LiveSession` com recuperação automática em caso de reconexão).
* **Aterramento Corporativo e Busca Vetorial:** Conectores nativos para fontes de conhecimento corporativo (Google Drive, BigQuery, repositórios internos) e índices do Vertex AI Vector Search para recuperação semântica de alto desempenho.
* **Execução Segura em Sandbox:** Ambientes isolados onde os agentes podem gerar e executar código dinamicamente (análise de dados, scripts Python/Go) sem expor a infraestrutura hospedeira a riscos.
* **Protocolos Agente a Agente (A2A):** Protocolos padronizados que permitem a agentes independentes descobrir capacidades, negociar contratos e delegar tarefas entre si através de diferentes áreas da organização.
* **Segurança, Identidade do Agente e Model Armor:** Permissões granulares de IAM, perímetros de VPC Service Controls e proteção em tempo de execução com o Model Armor contra injeção de prompt, vazamento de dados e violações de política.

## O que vem a seguir?

É difícil fazer justiça a todos esses frameworks e plataformas de execução em um único artigo, mas não se preocupe — ao longo dos próximos capítulos desta série vamos nos aprofundar em cada um deles:

* **Parte 4**: Mergulho no **Genkit para Go** — templates dotprompt, plugins customizados, streaming e observabilidade com o Dev UI.
* **Parte 5**: Mergulho no **Agent Development Kit (ADK)** — construção de hierarquias multiagente autônomas, delegação para subagentes e gerenciamento de sessões.
* **Parte 6**: Publicação de agentes Go no **Cloud Run** e na **Gemini Enterprise Agent Platform** com CI/CD e IAM prontos para produção.
* **Parte 7**: Mudando de marcha para o desenvolvimento de jogos em Go com o **[Ebitengine](https://ebitengine.org/)**.

Fique ligado, e boas programações!
