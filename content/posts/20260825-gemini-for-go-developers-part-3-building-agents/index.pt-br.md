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

Agora é hora de inverter os papéis e explorar o outro lado da equação: como construir agentes autônomos em Go. Neste capítulo, vamos dissecar a mecânica fundamental de um agente, definir um domínio concreto — um **Avaliador de Jogos Retrô** — e construí-lo passo a passo em três paradigmas distintos em Go:

1. Um loop de agente de baixo nível construído diretamente com o **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)**.
2. Um pipeline estruturado e orientado a fluxos (*flows*) construído com o **[Genkit](https://genkit.dev)**.
3. Um sistema multiagente modular orientado a sessões construído com o **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)** do Google.

Por fim, analisaremos os ambientes de execução (*runtimes*) disponíveis para publicar e hospedar agentes Go em produção na nuvem com alta confiabilidade.

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
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"
)

// GameItem representa um item colecionável no inventário pessoal.
type GameItem struct {
	Title     string  `json:"title"`
	Platform  string  `json:"platform"`
	Year      int     `json:"year"`
	Condition string  `json:"condition"` // ex: "Loose Cartridge", "CIB (Complete in Box)", "Mint"
	PricePaid float64 `json:"price_paid"`
	Notes     string  `json:"notes"`
}

// localCatalog simula um banco de dados de inventário de jogos retrô.
var localCatalog = []GameItem{
	{
		Title:     "Chrono Trigger",
		Platform:  "Super Nintendo (SNES)",
		Year:      1995,
		Condition: "CIB (Complete in Box)",
		PricePaid: 210.00,
		Notes:     "Inclui mapa original e cartão de registro.",
	},
	{
		Title:     "EarthBound",
		Platform:  "Super Nintendo (SNES)",
		Year:      1994,
		Condition: "Loose Cartridge",
		PricePaid: 180.00,
		Notes:     "Placa autêntica verificada; label em excelente estado.",
	},
	{
		Title:     "Castlevania: Symphony of the Night",
		Platform:  "Sony PlayStation",
		Year:      1997,
		Condition: "CIB (Black Label)",
		PricePaid: 135.00,
		Notes:     "Disco original com trilha sonora incluso.",
	},
}

// searchCatalogTool busca jogos correspondentes na coleção local.
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
			"message": fmt.Sprintf("Nenhum item correspondente a %q encontrado na coleção.", query),
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

	// Inicializa o cliente GenAI para Gemini Enterprise / Vertex AI
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("falha ao criar cliente: %v", err)
	}

	// 1. Declaração do schema da função personalizada de inventário
	catalogToolDecl := &genai.FunctionDeclaration{
		Name:        "search_catalog",
		Description: "Busca jogos no inventário pessoal do colecionador por título ou plataforma.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "Título do jogo ou plataforma a buscar (ex: 'EarthBound', 'SNES').",
				},
			},
			Required: []string{"query"},
		},
	}

	// 2. Configuração das ferramentas: função Go local + Google Search grounding
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: "Você é um Avaliador especialista em Jogos Retrô. Ao avaliar compras, consulte o " +
					"catálogo do usuário primeiro para verificar se ele já possui o item, e depois pesquise " +
					"preços atuais na Busca do Google para avaliar se a oferta é justa, cara ou uma pechincha."},
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

	// 3. Prompt inicial do usuário
	prompt := "Encontrei uma cópia de EarthBound para SNES completa na caixa (CIB) em estado impecável por $350. " +
		"Eu já tenho esse jogo na coleção, e $350 é um bom negócio comparado aos preços de mercado atuais?"

	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText(prompt)},
		},
	}

	model := "gemini-3.7-flash"
	maxTurns := 6

	// 4. O Loop do Agente (Requisição LLM -> Despacho de Tool -> Retorno do Resultado -> Nova Requisição)
	for turn := 0; turn < maxTurns; turn++ {
		resp, err := client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			log.Fatalf("erro ao gerar conteúdo: %v", err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			log.Fatal("candidato de resposta vazio recebido do modelo")
		}

		// Adiciona a resposta do modelo ao histórico da conversa
		modelContent := resp.Candidates[0].Content
		contents = append(contents, modelContent)

		// Verifica se o modelo solicitou a execução de ferramentas locais
		funcCalls := resp.FunctionCalls()
		if len(funcCalls) == 0 {
			// Condição de parada: modelo gerou a resposta final em linguagem natural
			fmt.Println("\n=== Veredito do Avaliador ===")
			fmt.Println(resp.Text())
			return
		}

		// Executa cada ferramenta requisitada e monta as partes de resposta
		var responseParts []*genai.Part
		for _, call := range funcCalls {
			fmt.Printf("[Harness] Executando tool: %s(args=%v)\n", call.Name, call.Args)

			var result map[string]any
			switch call.Name {
			case "search_catalog":
				result = searchCatalogTool(call.Args)
			default:
				result = map[string]any{"error": fmt.Sprintf("ferramenta não suportada: %s", call.Name)}
			}

			responseParts = append(responseParts, genai.NewPartFromFunctionResponse(call.Name, result))
		}

		// Retorna os resultados da execução como um turno do usuário
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: responseParts,
		})
	}

	log.Fatal("limite máximo de turnos atingido sem alcançar estado terminal")
}
```

### Executando o agente com o SDK

Para executar este exemplo, certifique-se de ter configurado seu projeto no Google Cloud e efetuado o login com Application Default Credentials (`gcloud auth application-default login`):

```sh
export GOOGLE_CLOUD_PROJECT="seu-projeto-gcp"
go run main.go
```

A saída ilustra o harness coordenando a consulta local ao catálogo e a fundamentação com a Busca do Google no servidor:

```text
[Harness] Executing tool: search_catalog(args=map[query:EarthBound])

=== Veredito do Avaliador ===
Aqui está a verificação da sua coleção e a avaliação para **EarthBound (SNES)**:

1. **Status na Coleção Atual**:
   - Você já possui **EarthBound** para Super Nintendo como **Cartucho Avulso (Loose)**, adquirido por **$180.00**.

2. **Avaliação de Preço de Mercado**:
   - Vendas verificadas recentes para uma cópia autêntica e **Completa na Caixa (CIB)** de EarthBound costumam variar entre **$1.200,00 e $1.500,00**, dependendo do estado da caixa, berço e guia original.

3. **Recomendação**:
   - A **$350.00**, uma cópia CIB autêntica e impecável é um **negócio excepcional** (mais de 70% abaixo do valor de mercado).
   - **Atenção**: Como EarthBound é um dos jogos de SNES mais pirateados, verifique com cuidado a impressão da caixa, cartão de registro e placa PCB antes de fechar a compra. Sendo autêntico, é uma excelente oportunidade para fazer um upgrade do seu cartucho avulso.
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

// GameItem representa um item colecionável no inventário pessoal.
type GameItem struct {
	Title     string  `json:"title"`
	Platform  string  `json:"platform"`
	Year      int     `json:"year"`
	Condition string  `json:"condition"`
	PricePaid float64 `json:"price_paid"`
	Notes     string  `json:"notes"`
}

// localCatalog simula um banco de dados de inventário de jogos retrô.
var localCatalog = []GameItem{
	{
		Title:     "Chrono Trigger",
		Platform:  "Super Nintendo (SNES)",
		Year:      1995,
		Condition: "CIB (Complete in Box)",
		PricePaid: 210.00,
		Notes:     "Inclui mapa original e cartão de registro.",
	},
	{
		Title:     "EarthBound",
		Platform:  "Super Nintendo (SNES)",
		Year:      1994,
		Condition: "Loose Cartridge",
		PricePaid: 180.00,
		Notes:     "Placa autêntica verificada; label em excelente estado.",
	},
	{
		Title:     "Castlevania: Symphony of the Night",
		Platform:  "Sony PlayStation",
		Year:      1997,
		Condition: "CIB (Black Label)",
		PricePaid: 135.00,
		Notes:     "Disco original com trilha sonora incluso.",
	},
}

type CatalogRequest struct {
	Query string `json:"query" jsonschema:"description=O título do jogo ou plataforma a buscar no inventário"`
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

	// 1. Inicializa o Genkit com o plugin da Vertex AI / Gemini Enterprise
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.VertexAI{
			ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
			Location:  "global",
		}),
		genkit.WithDefaultModel("vertexai/gemini-3.7-flash"),
	)

	// 2. Define ferramenta fortemente tipada com geração automática de schema
	catalogTool := genkit.DefineTool(
		g,
		"search_catalog",
		"Busca jogos no inventário pessoal do colecionador por título ou plataforma.",
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
					Message: fmt.Sprintf("Nenhum item correspondente a %q encontrado na coleção.", req.Query),
				}, nil
			}

			return CatalogResponse{
				Found:   true,
				Count:   len(matches),
				Results: matches,
			}, nil
		},
	)

	// 3. Define o fluxo estruturado de avaliação
	appraiserFlow := genkit.DefineFlow(
		g,
		"appraise_game",
		func(ctx context.Context, input string) (string, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithSystemPrompt(
					"Você é um Avaliador especialista em Jogos Retrô. Ajude colecionadores avaliando compras em potencial, "+
						"cruzando dados com o inventário pessoal e verificando cotações justas de mercado. "+
						"Sempre consulte o catálogo usando a ferramenta search_catalog antes de emitir recomendações.",
				),
				ai.WithPrompt(input),
				ai.WithTools(catalogTool),
			)
			if err != nil {
				return "", fmt.Errorf("falha ao gerar avaliação: %w", err)
			}
			return resp.Text(), nil
		},
	)

	// 4. Endpoint HTTP da API com encerramento gracioso
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/appraise", func(w http.ResponseWriter, req *http.Request) {
		var body AppraiserRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, "requisição inválida", http.StatusBadRequest)
			return
		}

		result, err := appraiserFlow.Run(req.Context(), body.Prompt)
		if err != nil {
			log.Printf("erro na execução do fluxo: %v", err)
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

	// Gerencia shutdown gracioso em sinais SIGINT (Ctrl+C) e SIGTERM
	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Avaliador de Jogos Retrô (Genkit) ouvindo em :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("falha no servidor: %v", err)
		}
	}()

	<-serverCtx.Done()
	log.Println("\nEncerrando servidor graciosamente...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("encerramento forçado do servidor: %v", err)
	}
	log.Println("Servidor finalizado com sucesso.")
}
```

### Executando o fluxo com o Genkit

Inicie o servidor Genkit exportando a variável do projeto Google Cloud:

```sh
export GOOGLE_CLOUD_PROJECT="seu-projeto-gcp"
export PORT=8080
go run main.go
```

Em outro terminal, envie uma solicitação de avaliação:

```sh
curl -s -X POST http://localhost:8080/api/appraise \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Encontrei EarthBound de SNES por $350. Eu já tenho ele, e é um bom negócio?"}' | jq .
```

O fluxo executa, consulta a ferramenta de catálogo e retorna o resultado estruturado:

```json
{
  "appraisal": "### 1. Verificação do Catálogo\n**Sim, você já possui este jogo.**\n* **Título:** *EarthBound* (SNES, 1994)\n* **Status na Coleção:** Cartucho Avulso (Loose)\n* **Estado/Notas:** Placa autêntica verificada; label em excelente estado.\n* **Preço Pago:** $180\n\n---\n\n### 2. Avaliação de Mercado\n* **Cartucho Avulso:** A cotação atual para uma fita avulsa autêntica fica entre **$320 e $380**. A **$350**, o valor está exatamente no **preço justo de mercado**.\n* **Completo na Caixa (CIB):** Caso o anúncio inclua a caixa original e o guia de estratégia, $350 seria uma oportunidade imperdível (cópias CIB são vendidas por **$1.500 a $2.500+**).\n\n---\n\n### 3. Recomendação\n* **Passe adiante (se for Avulso):** Como você já possui uma cópia autêntica em ótimo estado, pagar o valor de tabela por um cartucho avulso duplicado não faz sentido.\n* **Compre na hora (se for CIB):** Vale muito a pena se vier com embalagem original completa ou representar um grande upgrade de conservação."
}
```

O Genkit elimina a necessidade do loop manual. Ele cuida da desserialização dos argumentos em structs nativas de Go, executa a função, encaminha a resposta de volta ao modelo e gera spans de telemetria automaticamente para cada etapa. É a escolha ideal para fluxos determinísticos e estruturados.

## Implementando o agente com o Agent Development Kit (ADK)

Enquanto o Genkit se concentra em pipelines de aplicação estruturados, o **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)** (`google.golang.org/adk/v2`) do Google opera em um **alto nível de abstração**, projetado especificamente para agentes conversacionais autônomos, orquestração multiagente e arquiteturas complexas que demandam memória persistente e RAG corporativo.

O ADK padroniza ciclos de vida de agentes, delegação para subagentes e protocolos de comunicação agente a agente (A2A). Assim como o Genkit, ele suporta múltiplos provedores através de adaptadores modulares de modelo.

Quanto ao ambiente de hospedagem, a **Gemini Enterprise Agent Platform** é o destino preferencial caso você deseje usufruir de persistência gerenciada de sessões, conectores corporativos de aterramento e serviços de memória sem precisar programar bancos de dados próprios. Caso prefira gerenciar seu próprio estado em microsserviços conteinerizados, o **Google Cloud Run** é a escolha ideal.

Abaixo está o Avaliador de Jogos Retrô implementado com o ADK v2:

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

// GameItem representa um item colecionável no inventário pessoal.
type GameItem struct {
	Title     string  `json:"title"`
	Platform  string  `json:"platform"`
	Year      int     `json:"year"`
	Condition string  `json:"condition"`
	PricePaid float64 `json:"price_paid"`
	Notes     string  `json:"notes"`
}

// localCatalog simula um banco de dados de inventário de jogos retrô.
var localCatalog = []GameItem{
	{
		Title:     "Chrono Trigger",
		Platform:  "Super Nintendo (SNES)",
		Year:      1995,
		Condition: "CIB (Complete in Box)",
		PricePaid: 210.00,
		Notes:     "Inclui mapa original e cartão de registro.",
	},
	{
		Title:     "EarthBound",
		Platform:  "Super Nintendo (SNES)",
		Year:      1994,
		Condition: "Loose Cartridge",
		PricePaid: 180.00,
		Notes:     "Placa autêntica verificada; label em excelente estado.",
	},
	{
		Title:     "Castlevania: Symphony of the Night",
		Platform:  "Sony PlayStation",
		Year:      1997,
		Condition: "CIB (Black Label)",
		PricePaid: 135.00,
		Notes:     "Disco original com trilha sonora incluso.",
	},
}

type CatalogRequest struct {
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

	// 1. Inicializa o adaptador de modelo Gemini para Gemini Enterprise / Vertex AI
	model, err := gemini.NewModel(ctx, "gemini-3.7-flash", &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("falha ao criar modelo Gemini: %v", err)
	}

	// 2. Encapsula a consulta ao inventário como uma Function Tool do ADK
	catalogTool, err := functiontool.New(functiontool.Config{
		Name:        "search_catalog",
		Description: "Busca jogos no inventário pessoal do colecionador por título ou plataforma.",
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
				Message: fmt.Sprintf("Nenhum item correspondente a %q encontrado na coleção.", req.Query),
			}, nil
		}

		return CatalogResponse{
			Found:   true,
			Count:   len(matches),
			Results: matches,
		}, nil
	})
	if err != nil {
		log.Fatalf("falha ao criar tool de catálogo: %v", err)
	}

	// 3. Define o Agente LLM autônomo
	appraiserAgent, err := llmagent.New(llmagent.Config{
		Name:        "retro_game_appraiser",
		Model:       model,
		Description: "Avaliador especialista que analisa compras de jogos retrô e inventário pessoal.",
		Instruction: "Você é um Avaliador especialista em Jogos Retrô. Ajude colecionadores verificando o " +
			"status da coleção com a ferramenta search_catalog, avaliando variações de conservação e oferecendo recomendações objetivas.",
		Tools: []tool.Tool{catalogTool},
	})
	if err != nil {
		log.Fatalf("falha ao criar agente avaliador: %v", err)
	}

	// 4. Inicializa o Runner do ADK com serviço de sessão para controle de estado
	r, err := runner.New(runner.Config{
		AppName:           "retro_game_vault",
		Agent:             appraiserAgent,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		log.Fatalf("falha ao criar runner ADK: %v", err)
	}

	// 5. Handler HTTP com suporte a streaming e continuação de sessões multi-turno
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/chat", func(w http.ResponseWriter, req *http.Request) {
		var chatReq AppraiserRequest
		if err := json.NewDecoder(req.Body).Decode(&chatReq); err != nil {
			http.Error(w, "requisição inválida", http.StatusBadRequest)
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
				log.Printf("erro no runner: %v", err)
				http.Error(w, fmt.Sprintf("erro no agente: %v", err), http.StatusInternalServerError)
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

	// Shutdown gracioso em SIGINT ou SIGTERM
	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Avaliador de Jogos Retrô (ADK) ouvindo em :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("falha no servidor: %v", err)
		}
	}()

	<-serverCtx.Done()
	log.Println("\nEncerrando servidor graciosamente...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("encerramento forçado do servidor: %v", err)
	}
	log.Println("Servidor finalizado com sucesso.")
}
```

### Executando o agente com o ADK

Inicie o servidor do ADK com as credenciais do seu projeto no Google Cloud:

```sh
export GOOGLE_CLOUD_PROJECT="seu-projeto-gcp"
export PORT=8081
go run main.go
```

Em outro terminal, envie uma pergunta inicial para iniciar a conversa:

```sh
curl -s -X POST http://localhost:8081/api/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Eu tenho Chrono Trigger na minha coleção?"}' | jq .
```

O runner do ADK cuida da inicialização da sessão, executa a tool registrada `search_catalog` e devolve a resposta acompanhada do `session_id` persistido:

```json
{
  "appraisal": "Sim, você possui **Chrono Trigger** na sua coleção! Aqui estão os detalhes do seu inventário:\n\n* **Título:** Chrono Trigger\n* **Plataforma:** Super Nintendo (SNES)\n* **Ano:** 1995\n* **Estado:** Completo na Caixa (CIB)\n* **Preço Pago:** $210.00\n* **Notas:** Inclui mapa original e cartão de registro.",
  "session_id": "session-1787674771186589000"
}
```

Como o `runner` do ADK preserva o estado através do `SessionService`, você pode dar continuidade ao diálogo simplesmente enviando o mesmo `session_id` nas rodadas seguintes:

```sh
curl -s -X POST http://localhost:8081/api/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Quanto eu paguei por ele?", "session_id": "session-1787674771186589000"}' | jq .
```

```json
{
  "appraisal": "Você pagou **$210.00** por ele.",
  "session_id": "session-1787674771186589000"
}
```

O ADK entrega uma separação arquitetural limpa entre a definição do agente (`llmagent`), as ferramentas (`functiontool`) e o gerenciamento de estado da sessão (`runner` + `SessionService`), tornando-se a escolha ideal para hierarquias multiagente complexas e assistentes conversacionais.

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
