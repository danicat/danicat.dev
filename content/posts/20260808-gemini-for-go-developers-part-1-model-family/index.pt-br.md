---
categories:
- Agentic Coding
date: 2026-08-08
heroStyle: big
series:
- Gemini for Go Developers
series_order: 1
summary: O primeiro capítulo da série Gemini para Desenvolvedores Go, focando nos
  diferentes modelos da família Gemini e nas superfícies de API para consumi-los.
tags:
  - gemini
  - golang
title: 'Gemini para Desenvolvedores Go - Parte 1: A Família de Modelos Gemini'
---

Boas-vindas ao **Gemini para Desenvolvedores Go**! Esta série é o seu guia completo para construir aplicações de IA em Go prontas para produção. Ao longo de sete capítulos práticos, cobriremos desde programação agentic, até a construção de agentes autônomos com **Genkit** e **ADK**, desenvolvimento de jogos e o uso da **Stack G3** completa (Go, Gemini, GCP) para implantar aplicações na nuvem.

No Capítulo 1, preparamos a base explorando a família de modelos Gemini, configurações de modelos e escrevendo nosso primeiro código com o [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) oficial.

## A família de modelos Gemini

Geralmente tratamos "Gemini" como um nome único para as ofertas de IA do Google, assim como usamos "Google" como sinônimo de busca. Na realidade, o Gemini é uma família de modelos distintos criados para diferentes compensações operacionais (trade-offs).

Embora os modelos de fronteira sejam os que ganham as manchetes, saber quando usar modelos menores ou especializados é essencial para uma engenharia econômica. A seleção do modelo também [influencia diretamente a experiência do usuário](https://services.google.com/fh/files/blogs/google_delayexp.pdf) e a adoção do produto, pois a latência varia significativamente entre os níveis (tiers) de modelos.

Recorrer a um modelo Gemini Pro com altos níveis de raciocínio (thinking) para todas as tarefas é tentador, mas nem sempre é a escolha certa. Em muitos casos, isso apenas aumenta a latência por requisição e os custos de API sem entregar um resultado melhor.

### Esquema de nomenclatura dos modelos

Para navegar pelo catálogo do Gemini, ajuda entender como o Google nomeia seus modelos. Uma string de modelo padrão segue este padrão:

{{< katex >}}
\[
\text{[família]}-\text{[versão]}-\text{[nível]}{-\text{[modificador]}}
\]

Por exemplo: `gemini-3.6-flash` ou `gemini-3-pro-image`.

* **Família**: Embora a maioria dos modelos esteja na família Gemini, o Google também possui outras famílias de modelos como Veo e Lyria.
* **Números de Versão**: Representam saltos geracionais em inteligência, manipulação da janela de contexto e aderência a instruções.
* **Níveis de Modelo (Tiers)**:
  * **Pro**: Projetado para raciocínio complexo em múltiplas etapas.
  * **Flash**: Modelo balanceado com foco em velocidade.
  * **Flash-Lite**: Otimizado para velocidade e tarefas simples de alto volume (throughput).
* **Modificadores**: Podem indicar uma subfamília ou especialização, como `image` em `gemini-3.1-flash-image` ou `live` em `gemini-3.1-flash-live-preview`. Também podem incluir modificadores de ciclo de vida como `-preview` ou `-exp` (para experimental).

### Visão geral dos modelos

Aqui está uma visão geral dos principais modelos Gemini, começando com o modelo de fronteira Gemini 3.x:

#### Gemini 3.x

O Gemini 3.x é a linha principal de modelos de fronteira, disponível nos níveis Pro, Flash e Flash-Lite. Esses modelos de uso geral também são a escolha principal para geração de código e tarefas de engenharia de software.

Os modelos atuais incluem:
- `gemini-3.6-flash`: Modelo de alta velocidade para raciocínio multimodal e tarefas agentic
- `gemini-3.5-flash-lite`: Nível ultra-rápido de menor custo para microsserviços de alta vazão
- `gemini-3.1-pro-preview`: Nível avançado para raciocínio complexo e análise profunda de código

#### Modelos de imagem Gemini (Nano Banana)

Embora tecnicamente ainda façam parte da família Gemini, esses são modelos especializados para geração de imagens, oferecendo entrada e saída multimodais (imagem e texto). Eles são capazes de produzir imagens do zero e realizar edições em imagens existentes.

Os modelos atuais incluem:
- `gemini-2.5-flash-image` (também conhecido como Nano Banana)
- `gemini-3-pro-image` (também conhecido como Nano Banana Pro)
- `gemini-3.1-flash-image` (também conhecido como Nano Banana 2)
- `gemini-3.1-flash-lite-image` (também conhecido como Nano Banana 2 Lite)

#### Veo

Um modelo especializado em [geração de vídeo com áudio nativo](https://ai.google.dev/gemini-api/docs/veo). Os vídeos são gerados com base em prompts de texto e imagens de referência para marcar transições (quadro inicial e final). O Veo 3.1 gera clipes de até 8 segundos, mas é possível estendê-los em até 20 vezes em incrementos de 7 segundos.

Modelos atuais:
- `veo-3.1-generate-preview`
- `veo-3.1-lite-generate-preview` (geração rápida)

#### Lyria

O [Lyria](https://ai.google.dev/gemini-api/docs/music-generation) é especializado em geração de música, entregando composições instrumentais e vocais. O Lyria aceita tanto texto quanto imagens como entrada, com as imagens servindo de inspiração para a composição. Você também pode fornecer as letras ou deixar o modelo criá-las para você.

Modelos atuais:
- `lyria-3-pro-preview`
- `lyria-3-clip-preview` (clipes curtos de 30s)

#### Gemma

O [Gemma](https://ai.google.dev/gemma/docs) é a família de modelos de pesos abertos (open weights) do Google. Ele é treinado com a mesma tecnologia por trás do Gemini, mas projetado para ser implantado em sua própria infraestrutura. Além dos modelos oferecidos pelo Google, o Gemma conta com uma [forte comunidade](https://deepmind.google/models/gemma/gemmaverse/) que produz versões ajustadas (fine-tuned) para diversos casos de uso.

Alguns modelos Gemma são pequenos o suficiente para rodar em máquinas locais, permitindo casos de uso em que a conectividade de rede é limitada ou inexistente. Os modelos maiores são muito capazes, permitindo casos de uso em que soberania de dados e isolamento de rede são necessários.

#### Menções notáveis

- Modelos Live: Enquanto os modelos anteriores lidam com tarefas em lote ou requisição-resposta, o Google também oferece modelos ao vivo para streaming em tempo real. Procure por `-live` no nome (ex: `gemini-3.1-flash-live-preview`).
- Text-to-speech (Texto para fala): Gera fala a partir de texto com controle de narração usando tags de áudio (`gemini-3.1-flash-tts-preview`).
- Computer use (Uso de computador): Um modelo que consegue "ver" a tela e automatizar tarefas no navegador (`gemini-2.5-computer-use-preview-10-2025`).

Como você pode ver, o Gemini é muito mais do que um único modelo. É uma suíte completa que cobre desde chatbots básicos até criação multimodal e capacidades agentic.

Para especificações detalhadas de cada modelo, consulte a [documentação oficial dos modelos Gemini](https://ai.google.dev/gemini-api/docs/models).

## Capacidades adicionais

Além da conclusão de texto padrão, os modelos Gemini oferecem recursos adicionais para construir aplicações complexas. Aqui estão alguns dos mais importantes.

### Raciocínio (Thinking)

O Gemini 2.5 e modelos posteriores utilizam um processo de raciocínio interno que melhora significativamente o planejamento em múltiplas etapas, lógica, programação e capacidades matemáticas. Antes de gerar a resposta final, o modelo raciocina internamente gerando "tokens de pensamento" (thinking tokens) para analisar casos de borda e planejar estratégias.

O raciocínio é um recurso que pode ser controlado usando os parâmetros de configuração `thinking budget` no 2.5 e `thinking level` no 3.x. Quanto maior o orçamento ou nível de raciocínio, mais tempo e tokens o modelo dedicará durante a fase de raciocínio.

Quando o raciocínio está ativo, o total de tokens de saída faturáveis inclui tanto o texto de saída gerado quanto os tokens de pensamento do modelo. Ajustar o nível de raciocínio com base na complexidade da tarefa é um passo crítico para otimizar serviços em produção.

### Ferramentas integradas e chamadas de função (Function Calling)

A chamada de função permite que os modelos Gemini se integrem a ferramentas externas, APIs e bancos de dados. O Gemini suporta tanto ferramentas integradas (como `google_search` e `code_execution`) quanto funções personalizadas definidas no nível da aplicação.

A chamada de função tem três casos de uso principais:
- **Executar Ações:** Interagir com sistemas externos via APIs, como agendar reuniões, enviar e-mails, criar faturas ou controlar dispositivos de casa inteligente.
- **Aumentar Conhecimento:** Buscar informações em tempo real ou privadas em bancos de dados externos, microsserviços e bases de conhecimento.
- **Estender Capacidades:** Realizar cálculos matemáticos precisos, conversão de dados ou geração de gráficos que excedem os limites dos LLMs.

#### Como funciona a chamada de função

A chamada de função segue um processo de execução em 4 etapas entre sua aplicação e o modelo:

1. **Declarar ferramentas**: Defina as declarações de função (nome, descrição clara e JSON Schemas de parâmetros) e passe-as na configuração da requisição.
2. **Modelo identifica a intenção da ferramenta**: O modelo inspeciona o prompt e as declarações de ferramentas. Se uma ferramenta for necessária, ele retorna uma intenção estruturada contendo o nome da função e os argumentos.
3. **Executar código da função**: O modelo *não* executa código por conta própria. Sua aplicação recebe a requisição de chamada de função, executa a lógica local correspondente e captura o resultado.
4. **Retornar resultado da função**: Envie a saída da execução de volta ao modelo como uma etapa de resultado. O modelo usa esses dados para gerar sua resposta final em linguagem natural ou decidir se chamadas adicionais são necessárias.

### Saídas estruturadas (Structured Outputs)

Você pode configurar os modelos Gemini para gerar respostas que aderem a um [JSON Schema](https://ai.google.dev/gemini-api/docs/structured-output) fornecido. Isso simplifica a extração de dados estruturados a partir de texto, eliminando parsing frágil ao converter respostas do modelo em estruturas de dados.

Além de escrever JSON Schemas brutos em payloads REST, os SDKs do Google GenAI permitem que os desenvolvedores definam schemas usando construções nativas da linguagem, como [Pydantic](https://docs.pydantic.dev/) em Python e struct tags em Go.

## Consumindo modelos programaticamente

Com o ecossistema de modelos e suas capacidades cobertos, vamos ver como incorporar essas APIs em aplicações Go.

Assim como existem modelos diferentes para casos de uso distintos, várias superfícies de API estão disponíveis. Vamos começar pela mais básica: Generate Content.

### API Generate Content

Esta é a [API generativa](https://ai.google.dev/api/generate-content#method:-models.generatecontent) mais básica. É uma interface sem estado (stateless) que aceita uma única requisição e retorna uma resposta. Para conversas com múltiplos turnos, sua aplicação deve enviar todo o histórico do chat a cada chamada.

Isso exige gerenciar ativamente o histórico da conversa para permanecer dentro dos limites da janela de contexto. As aplicações normalmente resumem o histórico assim que ele atinge um limite. Para reduzir os custos de entrada em sessões longas, a API do Gemini suporta [caching implícito](https://ai.google.dev/gemini-api/docs/caching) para todos os modelos desde o Gemini 2.5, bem como [caching explícito](https://ai.google.dev/gemini-api/docs/generate-content/caching) para payloads pesados.

Embora a API Generate Content seja boa para gerações simples e sem estado, ela está sendo gradualmente substituída pela nova e mais capaz API de Interações.

### API de Interações (Interactions API)

> Nota: Até a data de hoje, a API de Interações ainda não é suportada pelo SDK oficial do Go GenAI. O progresso da implementação está sendo acompanhado nesta [issue do GitHub](https://github.com/googleapis/go-genai/issues/658).

A [API de Interações](https://ai.google.dev/gemini-api/docs/interactions-overview) é a interface unificada do Google projetada para todas as tarefas, desde chats simples e uso de ferramentas até fluxos de trabalho agentic complexos. Ela pode gerenciar o histórico de conversas no lado do servidor para que sua aplicação não precise fazer isso.

### Live API

A [API Live](https://ai.google.dev/gemini-api/docs/live-api) permite conversas em tempo real por voz e vídeo bidirecionais sobre WebSockets. Ela detecta automaticamente quando o usuário fala ou interrompe, fazendo com que as interações por voz pareçam naturais, enquanto suporta ferramentas como busca na web e chamada de função diretamente na sessão ao vivo.

### Batch API

A [API Batch](https://ai.google.dev/gemini-api/docs/batch-api) permite processar grandes volumes de dados de forma assíncrona pela metade do preço. Os trabalhos são executados em segundo plano durante horários fora de pico (geralmente concluídos em até 24 horas), sendo ideal para cargas de trabalho não urgentes.

### API de Agentes Gerenciados (Managed Agents API)

Os [agentes gerenciados](https://ai.google.dev/gemini-api/docs/agents) oferecem um ambiente de execução totalmente hospedado, onde agentes de IA planejam e executam tarefas de forma autônoma. Uma única chamada de API provisiona um sandbox Linux isolado com runtimes pré-instalados como Python e Node, permitindo que o agente execute código, gerencie arquivos e navegue na web.

O Google oferece dois agentes gerenciados pré-construídos prontos para uso:
- **Agente Antigravity** (`antigravity-preview-05-2026`): O agente de uso geral padrão alimentado pelo Gemini 3.6 Flash (configurável para Gemini 3.5 Flash ou Flash-Lite) para execução de código, gerenciamento de arquivos e acesso à web.
- **Agente Deep Research** (`deep-research-preview-04-2026`): Um agente de pesquisa autônomo que consulta dados da web de múltiplas fontes e compila relatórios detalhados em segundo plano.

Você também pode estender o agente Antigravity definindo regras de sistema inline ou montando um arquivo `AGENTS.md`, anexando diretórios de habilidades estruturados (`SKILL.md`), ou montando arquivos locais, buckets do Cloud Storage e repositórios Git diretamente no workspace remoto (`/workspace`).

## Acesso e faturamento

Ao integrar o Gemini, o Google oferece dois modos principais de acesso e faturamento, dependendo das suas necessidades:

1. **Google AI Studio (Google AI)**: Roteia requisições através da API do Gemini usando uma chave de API (`GEMINI_API_KEY` ou `GOOGLE_API_KEY`). Ideal para prototipagem, projetos pessoais, aplicações indie e integração rápida de desenvolvedores.
2. **Gemini Enterprise (antigo Vertex AI)**: Roteia requisições através de endpoints do Google Cloud usando Google Cloud IAM, Application Default Credentials (ADC), chaves de conta de serviço ou tokens de usuário OAuth 2.0. Ideal para cargas de trabalho empresariais em produção que exigem privacidade rigorosa de dados, conformidade de segurança, SLAs, gerenciamento de recursos do GCP e descontos por uso comprometido.

## Go GenAI SDK

Agora vamos ver como isso funciona em código Go.

O SDK oficial para integrar o Gemini em aplicações Go é [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai). 

Às vezes você o verá sendo chamado de SDK "unificado", pois foi projetado para suportar todos os modelos do Google e autenticação tanto no Google AI quanto no Gemini Enterprise. Ele substitui o pacote legado `github.com/google/generative-ai-go`, que foi descontinuado (deprecated).

Instale-o com `go get`:

```bash
go get google.golang.org/genai
```

Aqui está um exemplo usando autenticação e faturamento do Gemini Enterprise:

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

Você pode executar este programa usando o seguinte:

```sh
export GOOGLE_CLOUD_PROJECT="seu-id-de-projeto-aqui"
go run main.go
```

Aqui está o resultado:

![Saída de imagem do gato mago gerada no terminal Go](image.png "O verdadeiro propósito da IA: geração infinita de fotos de gatos")

Embora este seja apenas um exemplo simples para mostrar como trabalhar com o SDK, ao longo desta série veremos mais exemplos do Go GenAI SDK e de frameworks de nível mais alto, como [Genkit](https://genkit.dev/) e o [Agent Development Kit (ADK)](https://adk.dev/).

## O que vem a seguir?

Na **Parte 2** da série **Gemini para Desenvolvedores Go**, faremos um mergulho profundo em agentes de programação e como preparar seu ambiente para trabalhar em bases de código Go. Fique ligado!
