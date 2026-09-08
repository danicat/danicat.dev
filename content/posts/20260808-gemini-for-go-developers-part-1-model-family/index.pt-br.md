---
categories:
- Applied GenAI
date: 2026-08-08
heroStyle: big
series:
- Gemini for Go Developers
series_order: 1
summary: Conheça os diferentes modelos da família Gemini, suas capacidades e como consumi-los via código.
tags:
  - gemini
  - golang
title: "Gemini para Desenvolvedores Go: A Família de Modelos Gemini"
slug: "gemini-for-go-developers-part-1-model-family"
aliases:
  - "/pt-br/posts/20260808-gemini-for-go-developers-part-1-model-family/"
description: "Parte 1 de Gemini para Desenvolvedores Go: compare os modelos Gemini 3.x, Flash, Pro e Nano Banana, explore as APIs e escreva seu primeiro código Go com o GenAI SDK."
proficiencyLevel: "Beginner"
dependencies:
  - "Go 1.24+"
  - "google.golang.org/genai"
---

Boas-vindas ao **Gemini para Desenvolvedores Go**! Esta série é o seu guia completo para construir software potencializado por IA em Go. Ao longo de sete capítulos práticos, abordaremos desde codificação agentiva e a criação de agentes autônomos com **Genkit** e **ADK**, até o desenvolvimento de jogos e o uso da **G3 Stack** completa (Go, Gemini, GCP) para fazer deploy de aplicações na nuvem.

No Capítulo 1, estabelecemos a base explorando a família de modelos Gemini, configurações de modelos e escrevendo nosso primeiro código com o [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) oficial.

## A família de modelos Gemini

Geralmente tratamos "Gemini" como um nome único para os produtos de IA do Google, muito parecido com usar "Google" como sinônimo de busca. Na realidade, o Gemini é uma família de modelos distintos, construídos para diferentes compensações operacionais (*trade-offs*).

Embora os modelos de fronteira sejam os que ganham as manchetes, saber quando usar modelos menores ou especializados é essencial para uma engenharia com bom custo-benefício. A seleção do modelo também [influencia diretamente a experiência do usuário](https://services.google.com/fh/files/blogs/google_delayexp.pdf) e a adoção do produto, já que a latência varia bruscamente entre os diferentes tiers de modelos.

Recorrer a um modelo Gemini Pro com níveis altos de *thinking* para cada tarefa é tentador, mas nem sempre é a escolha certa. Em muitos casos, isso apenas aumenta a latência por requisição e os custos de API sem entregar um resultado melhor.

### Esquema de nomenclatura dos modelos

Para navegar pelo catálogo do Gemini, ajuda entender como o Google nomeia seus modelos. Uma string de modelo padrão segue este padrão:

{{< katex >}}
\[
\text{[family]}-\text{[version]}-\text{[tier]}{-\text{[modifier]}}
\]

Por exemplo: `gemini-3.7-flash` ou `gemini-3-pro-image`.

**Família (Family)**: Embora a maioria dos modelos pertença à família Gemini, o Google também possui outras famílias de modelos, como Veo e Lyria.
**Números de Versão (Version Numbers)**: Representam saltos geracionais em inteligência, manipulação da janela de contexto e aderência a instruções.
**Tiers de Modelo (Model Tiers)**:
* **Pro**: Projetado para raciocínio complexo em múltiplas etapas.
* **Flash**: Modelo balanceado com foco em velocidade.
* **Flash-Lite**: Otimizado para velocidade e tarefas simples de alto rendimento (*throughput*).
**Modificadores (Modifiers)**: Podem indicar uma subfamília ou especialização, como `image` em `gemini-3.1-flash-image` ou `live` em `gemini-3.1-flash-live-preview`. Também podem incluir modificadores de ciclo de vida como `-preview` ou `-exp` (para experimental).

### Visão geral dos modelos

Aqui está uma visão geral dos principais modelos Gemini, começando com a linha de modelos de fronteira Gemini 3.x:

#### Gemini 3.x

O Gemini 3.x é a principal linha de modelos de fronteira, disponível nos tiers Pro, Flash e Flash-Lite. Esses modelos de uso geral também são a escolha principal para geração de código e tarefas de engenharia de software.

Os modelos atuais incluem:
- `gemini-3.6-flash`: O cavalo de batalha veloz para raciocínio multimodal e tarefas agentivas
- `gemini-3.5-flash-lite`: O tier mais barato e ultrarrápido para microsserviços de alto throughput
- `gemini-3.1-pro-preview`: Tier avançado para raciocínio complexo em múltiplas etapas e análise profunda de bases de código

#### Modelos de imagem Gemini (Nano Banana)

Embora tecnicamente ainda faça parte da família Gemini, este é um modelo especializado para geração de imagens, fornecendo tanto entrada quanto saída multimodais (imagem e texto). Ele é capaz de produzir imagens do zero e realizar edições em imagens existentes.

Os modelos atuais incluem:
- `gemini-2.5-flash-image` (também conhecido como Nano Banana)
- `gemini-3-pro-image` (também conhecido como Nano Banana Pro)
- `gemini-3.1-flash-image` (também conhecido como Nano Banana 2)
- `gemini-3.1-flash-lite-image` (também conhecido como Nano Banana 2 Lite)

#### Veo

Um modelo especializado em [geração de vídeo com áudio nativo](https://ai.google.dev/gemini-api/docs/veo). Os vídeos são gerados com base em prompts de texto e imagens-chave para marcar transições (quadro inicial e final) e como referências. O Veo 3.1 gera clipes de até 8 segundos, mas é possível estendê-los em até 20 vezes em incrementos de 7 segundos.

Modelos atuais:
- `veo-3.1-generate-preview`
- `veo-3.1-lite-generate-preview` (geração rápida)

#### Lyria

O [Lyria](https://ai.google.dev/gemini-api/docs/music-generation) é especializado em geração de música, entregando tanto composições instrumentais quanto vocais. O Lyria aceita texto e imagens como entrada, com as imagens servindo de inspiração para a composição. Você também pode fornecer a letra por conta própria ou deixar que o modelo a crie para você.

Modelos atuais:
- `lyria-3-pro-preview`
- `lyria-3-clip-preview` (clipes curtos de 30s)

#### Gemma

O [Gemma](https://ai.google.dev/gemma/docs) é uma família de modelos de pesos abertos (*open-weights*) do Google. Ele é treinado com a mesma tecnologia por trás do Gemini, mas projetado para ser executado na sua própria infraestrutura. Além dos modelos oferecidos pelo Google, o Gemma também conta com uma [comunidade forte](https://deepmind.google/models/gemma/gemmaverse/) que produz versões com ajuste fino (*fine-tuned*) para todo tipo de caso de uso.

Alguns modelos Gemma são pequenos o suficiente para serem executados em máquinas locais, viabilizando casos de uso em que a conectividade de rede é limitada ou inexistente. Os modelos maiores são muito capazes, atendendo a casos de uso onde soberania e isolamento de rede são necessários.

#### Menções notáveis

- Modelos Live: Enquanto os modelos anteriores lidam com tarefas em lote ou requisição-resposta, o Google também oferece modelos live para streaming em tempo real. Procure por `-live` no nome (por exemplo, `gemini-3.1-flash-live-preview`).
- Text-to-speech: Gera fala a partir de texto com controle de narração usando tags de áudio (`gemini-3.1-flash-tts-preview`).
- Computer use: Um modelo que pode "enxergar" a tela e automatizar tarefas no navegador (`gemini-2.5-computer-use-preview-10-2025`).

Como você pode ver, o Gemini é muito mais do que um único modelo. É uma suíte completa cobrindo tudo, desde chatbots básicos até criação multimodal e recursos agentivos.

Para especificações detalhadas de cada modelo, consulte a [documentação oficial dos modelos Gemini](https://ai.google.dev/gemini-api/docs/models).

## Recursos adicionais

Além da completude padrão de texto, os modelos Gemini suportam recursos adicionais para a construção de aplicações complexas. Aqui estão alguns dos mais importantes.

### Thinking

Os modelos Gemini 2.5 e posteriores utilizam um processo interno de raciocínio que aprimora significativamente as capacidades de planejamento em múltiplas etapas, lógica, programação e matemática. Antes de gerar a resposta final, o modelo raciocina internamente gerando "tokens de pensamento" (*thinking tokens*) para analisar casos extremos (*edge cases*) e planejar estratégias em múltiplos passos.

O *thinking* é um recurso que pode ser controlado usando os parâmetros de configuração `thinking budget` no 2.5 e `thinking level` no 3.x. Quanto maior o *budget* ou o *thinking level*, mais tempo e tokens o modelo gastará durante a fase de raciocínio.

Quando o *thinking* está ativo, o total de tokens de saída faturáveis inclui tanto o texto de saída gerado quanto os tokens de pensamento gerados pelo modelo. Ajustar o nível de *thinking* com base na complexidade da tarefa é um passo fundamental para serviços em produção.

### Ferramentas integradas e chamadas de função (*function calling*)

O *function calling* permite que os modelos Gemini se comuniquem com ferramentas externas, APIs e bancos de dados. O Gemini suporta tanto ferramentas integradas (como `google_search` e `code_execution`) quanto funções customizadas definidas no nível da aplicação.

O *function calling* tem três casos de uso principais:
- **Executar ações:** Interagir com sistemas externos via APIs, como agendar reuniões, enviar e-mails, criar faturas ou controlar dispositivos de casa inteligente.
- **Aumentar conhecimento:** Buscar informações em tempo real ou privadas a partir de bancos de dados externos, microsserviços e bases de conhecimento.
- **Estender capacidades:** Realizar cálculos matemáticos precisos, conversão de dados ou geração de gráficos que superam os limites do LLM.

#### Como o *function calling* funciona

O *function calling* segue um processo de execução em 4 etapas entre sua aplicação e o modelo:

1. **Declarar ferramentas**: Defina as declarações das funções (nome, descrição clara e JSON Schemas dos parâmetros) e passe-as na configuração da requisição.
2. **Modelo identifica a intenção da ferramenta**: O modelo inspeciona o prompt e as declarações de ferramentas. Se uma ferramenta for necessária, ele retorna uma intenção estruturada de chamada de ferramenta contendo o nome da função e os argumentos.
3. **Executar o código da função**: O modelo *não* executa o código diretamente. Sua aplicação recebe a requisição de chamada de função, executa a lógica local correspondente e captura o resultado.
4. **Retornar o resultado da função**: Envie a saída da execução de volta ao modelo como uma etapa de resultado da função. O modelo usa esses dados para gerar sua resposta final em linguagem natural ou decidir se chamadas adicionais de ferramentas são necessárias.

### Saídas estruturadas (*structured outputs*)

Você pode configurar os modelos Gemini para gerar respostas que sigam rigorosamente um [JSON Schema](https://ai.google.dev/gemini-api/docs/structured-output) fornecido. Isso simplifica a extração de dados estruturados a partir de texto, eliminando rotinas frágeis de parsing ao converter respostas do modelo em estruturas de dados.

Além de escrever JSON Schemas brutos em payloads REST, os SDKs do Google GenAI permitem que desenvolvedores definam schemas usando construções nativas da linguagem, como [Pydantic](https://docs.pydantic.dev/) em Python e struct tags em Go.

## Consumindo modelos programaticamente

Com o ecossistema de modelos e suas capacidades cobertos, vamos ver como incorporar essas APIs em aplicações Go.

Assim como modelos diferentes existem para casos de uso diferentes, várias superfícies de API estão disponíveis. Vamos começar com a mais básica: Generate Content.

### Generate Content API

Esta é a [API generativa](https://ai.google.dev/api/generate-content#method:-models.generatecontent) mais básica. Trata-se de uma interface stateless (sem estado) que aceita uma única requisição e retorna uma resposta. Para conversas de múltiplos turnos, sua aplicação precisa enviar todo o histórico de chat a cada chamada.

Isso exige gerenciar ativamente o histórico de conversa para ficar dentro dos limites da janela de contexto. As aplicações normalmente resumem o histórico assim que ele atinge um determinado limite. Para reduzir custos de entrada em sessões longas, a Gemini API suporta [caching implícito](https://ai.google.dev/gemini-api/docs/caching) para todos os modelos a partir do Gemini 2.5, além de [caching explícito](https://ai.google.dev/gemini-api/docs/generate-content/caching) para payloads pesados.

Embora a Generate Content API seja boa para gerações stateless simples, ela está sendo gradualmente substituída pela mais recente e mais poderosa Interactions API.

### Interactions API

> Nota: Até o momento, a Interactions API ainda não é suportada pelo Go GenAI SDK oficial. O progresso da implementação está sendo acompanhado nesta [issue do GitHub](https://github.com/googleapis/go-genai/issues/658).

A [Interactions API](https://ai.google.dev/gemini-api/docs/interactions-overview) é a interface unificada do Google projetada para todas as tarefas, desde chat simples e uso de ferramentas até fluxos de trabalho agentivos complexos. Ela pode gerenciar o histórico de conversas no lado do servidor, de modo que sua aplicação não precise fazer isso.

### Live API

A [Live API](https://ai.google.dev/gemini-api/docs/live-api) permite conversas de voz e vídeo bidirecionais em tempo real via WebSockets. Ela detecta automaticamente quando o usuário fala ou interrompe, fazendo com que as interações de voz pareçam naturais, enquanto suporta ferramentas como busca na web e function calling diretamente na sessão ao vivo.

### Batch API

A [Batch API](https://ai.google.dev/gemini-api/docs/batch-api) permite processar grandes volumes de dados de forma assíncrona pela metade do preço. Os jobs são executados em segundo plano durante horários de menor movimento (geralmente concluídos em até 24 horas), tornando-a ideal para cargas de trabalho não urgentes.

### Managed Agents API

Os [agentes gerenciados](https://ai.google.dev/gemini-api/docs/agents) oferecem um ambiente de execução totalmente hospedado onde agentes de IA planejam e executam tarefas de forma autônoma. Uma única chamada de API provisiona um sandbox Linux isolado no nível do SO com runtimes pré-instalados como Python e Node, permitindo que o agente execute código, gerencie arquivos e navegue na web.

O Google fornece dois agentes gerenciados prontos para uso:
- **Antigravity Agent** (`antigravity-preview-05-2026`): O agente padrão de propósito geral equipado com o Gemini 3.6 Flash (configurável para Gemini 3.5 Flash ou Flash-Lite) para execução de código, gerenciamento de arquivos e acesso à web.
- **Deep Research Agent** (`deep-research-preview-04-2026`): Um agente de pesquisa autônomo que consulta dados da web de múltiplas fontes e compila relatórios detalhados de pesquisa em segundo plano.

Você também pode estender o agente Antigravity definindo regras de sistema inline ou montando um arquivo `AGENTS.md`, anexando diretórios de skills estruturados (`SKILL.md`) ou montando arquivos locais, buckets do Cloud Storage e repositórios Git diretamente no workspace remoto (`/workspace`).

## Acesso e faturamento

Ao integrar o Gemini, o Google oferece dois modos principais de acesso e faturamento, dependendo das suas necessidades:

1. **Google AI Studio (Google AI)**: Roteia requisições através da Gemini API usando uma chave de API (`GEMINI_API_KEY` ou `GOOGLE_API_KEY`). Ideal para prototipagem, projetos pessoais, aplicações indie e onboarding rápido para desenvolvedores.
2. **Gemini Enterprise (anteriormente Vertex AI)**: Roteia requisições através de endpoints do Google Cloud usando Google Cloud IAM, Application Default Credentials (ADC), chaves de conta de serviço ou tokens de usuário OAuth 2.0. Ideal para cargas de trabalho empresariais em produção que exigem rigorosa privacidade de dados, conformidade de segurança, SLAs, gerenciamento de recursos do GCP e descontos por uso contínuo (*committed-use discounts*).

## Go GenAI SDK

Agora vamos ver como isso funciona em código Go.

O SDK oficial para integrar o Gemini em aplicações Go é o [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai).

Às vezes você o verá sendo chamado de SDK "unificado", pois foi projetado para suportar todos os modelos do Google e autenticação tanto via Google AI quanto pelo Gemini Enterprise. Ele substitui o pacote legado `github.com/google/generative-ai-go`, que foi descontinuado (*deprecated*).

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

Para executar este exemplo, você precisará de um projeto no Google Cloud com a API AI Platform habilitada. Por exemplo:

```sh
export GOOGLE_CLOUD_PROJECT="your-project-id-goes-here"
go run main.go
```

Aqui está o resultado:

![Saída de imagem de gato mago gerada no terminal Go](image.png "O verdadeiro propósito da IA: geração infinita de fotos de gatinhos")

Embora este seja apenas um exemplo simples para mostrar como trabalhar com o SDK, ao longo desta série veremos mais exemplos tanto do Go GenAI SDK quanto de frameworks de nível mais alto como o [Genkit](https://genkit.dev/) e o [Agent Development Kit (ADK)](https://adk.dev/).

## O que vem a seguir?

Na [**Parte 2: Programando com o Gemini**]({{< ref "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini" >}}) da série **Gemini para Desenvolvedores Go**, vamos nos aprofundar em agentes de codificação e em como preparar seu ambiente para trabalhar em bases de código Go. Fique ligado!
