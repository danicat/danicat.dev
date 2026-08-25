---categories:
- Applied GenAI
date: 2026-08-08
heroStyle: big
series:
- Gemini for Go Developers
series_order: 1
summary: O primeiro capítulo da série Gemini para Desenvolvedores Go, explorando os
  diferentes modelos da família Gemini e as superfícies de API para utilizá-los.
tags:
  - gemini
  - golang
title: "Gemini para Desenvolvedores Go: A Família de Modelos Gemini"
slug: "gemini-for-go-developers-part-1-model-family"
aliases:
  - "/pt-br/posts/20260808-gemini-for-go-developers-part-1-model-family/"
description: "Parte 1 de Gemini para Go: compare os modelos Gemini 3.x, Flash, Pro e Nano Banana, conheça as APIs e escreva seu primeiro código com o Go GenAI SDK."
proficiencyLevel: "Beginner"
dependencies:
  - "Go 1.24+"
  - "google.golang.org/genai"
---

Boas-vindas ao **Gemini para Desenvolvedores Go**! Esta série é o seu guia completo para construir aplicações com inteligência artificial em Go. Ao longo de sete capítulos práticos, abordaremos desde o desenvolvimento agentivo e a criação de agentes autônomos com **Genkit** e **ADK**, até o desenvolvimento de jogos e o uso da **Stack G3** completa (Go, Gemini e GCP) para implantar aplicações na nuvem.

Neste Capítulo 1, preparamos a base explorando a família de modelos Gemini, suas configurações essenciais e escrevendo nossas primeiras linhas de código com o [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) oficial.

## A família de modelos Gemini

Geralmente usamos "Gemini" como um termo guarda-chuva para os produtos de IA do Google, da mesma forma que dizemos "dar um Google" para pesquisas. Na prática, o Gemini é uma família de modelos distintos, cada um projetado com compromissos operacionais (*trade-offs*) específicos.

Embora os modelos de fronteira (*frontier models*) dominem as manchetes, saber quando optar por modelos menores ou especializados é um requisito fundamental para construir sistemas eficientes e econômicos. Além disso, a escolha do modelo [impacta diretamente a experiência do usuário](https://services.google.com/fh/files/blogs/google_delayexp.pdf) e a taxa de adoção do produto, já que a latência varia significativamente entre os diferentes níveis (*tiers*).

Recorrer a um modelo Gemini Pro com níveis altos de raciocínio (*thinking*) para qualquer demanda é tentador, mas nem sempre é a melhor decisão de arquitetura. Em muitos cenários, isso apenas encarece os custos de API e eleva o tempo de resposta sem trazer ganhos proporcionais de qualidade.

### Esquema de nomenclatura dos modelos

Para navegar pelo catálogo do Gemini com facilidade, vale a pena entender a lógica que o Google adota na nomenclatura de seus modelos:

{{< katex >}}
\[
\text{[família]}-\text{[versão]}-\text{[nível]}{-\text{[modificador]}}
\]

Por exemplo: `gemini-3.6-flash` ou `gemini-3-pro-image`.

* **Família**: Embora a maioria dos modelos pertença à família Gemini, o Google também mantém outras linhas especializadas, como Veo e Lyria.
* **Versão**: Representa saltos geracionais em inteligência, suporte à janela de contexto e fidelidade às instruções do prompt.
* **Nível (*Tier*)**:
  * **Pro**: Projetado para raciocínio complexo em múltiplas etapas.
  * **Flash**: Modelo balanceado, priorizando velocidade e custo-benefício.
  * **Flash-Lite**: Otimizado para altíssima vazão (*throughput*) e tarefas objetivas.
* **Modificadores**: Indicam especializações ou subfamílias, como `image` em `gemini-3.1-flash-image` ou `live` em `gemini-3.1-flash-live-preview`. Também identificam fases do ciclo de vida, como `-preview` ou `-exp` (experimental).

### Visão geral dos modelos

Abaixo está um resumo dos principais modelos disponíveis, a começar pela linha de fronteira Gemini 3.x:

#### Gemini 3.x

A linha Gemini 3.x representa a atual geração de modelos de fronteira, distribuída nos níveis Pro, Flash e Flash-Lite. Esses modelos de propósito geral são também a escolha prioritária para geração de código e tarefas avançadas de engenharia de software.

Modelos em destaque:
- `gemini-3.6-flash`: O cavalo de batalha de alta velocidade para raciocínio multimodal e tarefas agentivas.
- `gemini-3.5-flash-lite`: A opção mais econômica e veloz para microsserviços de alto volume.
- `gemini-3.1-pro-preview`: Nível avançado para raciocínio profundo em múltiplos passos e análise de grandes bases de código.

#### Modelos de imagem Gemini (Nano Banana)

Apesar de pertencerem à família Gemini, estes são modelos especializados na geração e edição de imagens, oferecendo entrada e saída multimodais (imagem e texto). São capazes de criar imagens do zero ou modificar ilustrações existentes.

Modelos atuais:
- `gemini-2.5-flash-image` (conhecido como Nano Banana)
- `gemini-3-pro-image` (conhecido como Nano Banana Pro)
- `gemini-3.1-flash-image` (conhecido como Nano Banana 2)
- `gemini-3.1-flash-lite-image` (conhecido como Nano Banana 2 Lite)

#### Veo

Modelo dedicado à [geração de vídeos com áudio nativo](https://ai.google.dev/gemini-api/docs/veo). Gera vídeos a partir de prompts de texto e imagens-chave para guiar transições (quadro inicial e final) e referências de estilo. O Veo 3.1 cria cenas de até 8 segundos, com suporte a extensões de até 20 vezes em blocos adicionais de 7 segundos.

Modelos atuais:
- `veo-3.1-generate-preview`
- `veo-3.1-lite-generate-preview` (geração rápida)

#### Lyria

O [Lyria](https://ai.google.dev/gemini-api/docs/music-generation) é especializado na criação musical, gerando tanto composições instrumentais quanto faixas com vocais completos. O modelo aceita texto e imagens como entrada (as imagens servem de inspiração conceitual para a composição). Você pode fornecer sua própria letra ou deixar que o modelo crie uma para você.

Modelos atuais:
- `lyria-3-pro-preview`
- `lyria-3-clip-preview` (vinhetas curtas de até 30s)

#### Gemma

A linha [Gemma](https://ai.google.dev/gemma/docs) reúne os modelos de pesos abertos (*open weights*) do Google. Criada com a mesma base tecnológica do Gemini, foi desenvolvida para ser implantada na sua própria infraestrutura. Além das versões oficiais, a linha conta com um [ecossistema comunitário vibrante](https://deepmind.google/models/gemma/gemmaverse/) de modelos com ajustes finos (*fine-tuning*) para os mais diversos nichos.

As versões menores do Gemma são leves o bastante para rodar localmente, viabilizando soluções em ambientes com conectividade restrita. Já as variantes maiores oferecem alta precisão, atendendo a requisitos rígidos de soberania de dados e isolamento de rede.

#### Outros destaques

- **Modelos Live**: Enquanto modelos padrão respondem em lote ou no formato de requisição e resposta, os modelos Live operam com fluxos contínuos de áudio e vídeo em tempo real via streaming (identificados pelo sufixo `-live`, como `gemini-3.1-flash-live-preview`).
- **Text-to-Speech**: Converte texto em fala com suporte a marcações de áudio para controle expressivo da narração (`gemini-3.1-flash-tts-preview`).
- **Computer Use**: Modelo com capacidade de interpretar a tela e automatizar tarefas visuais no navegador (`gemini-2.5-computer-use-preview-10-2025`).

Como se vê, o Gemini vai muito além de um único modelo: é uma plataforma completa para atender desde chatbots básicos até criação multimodal e fluxos agentivos autônomos.

Para detalhes técnicos sobre cada variante, consulte a [documentação oficial dos modelos Gemini](https://ai.google.dev/gemini-api/docs/models).

## Recursos avançados

Além da geração tradicional de texto, os modelos Gemini oferecem funcionalidades essenciais para a construção de aplicações robustas. Vamos analisar as principais:

### Raciocínio (*Thinking*)

A partir da versão 2.5, os modelos Gemini contam com um processo de raciocínio interno que aprimora o planejamento em múltiplas etapas, a resolução lógica, a escrita de código e o cálculo matemático. Antes de formular a resposta final, o modelo produz internamente "tokens de pensamento" (*thinking tokens*) para avaliar cenários de borda e estruturar a estratégia de resolução.

Esse comportamento pode ser calibrado por meio dos parâmetros de configuração: `thinking budget` no Gemini 2.5 e `thinking level` no Gemini 3.x. Quanto maior o valor estipulado, mais tempo e tokens o modelo dedicará ao raciocínio prévio.

Vale ressaltar que os tokens de pensamento são contabilizados no faturamento da requisição. Calibrar o nível de raciocínio conforme a complexidade de cada tarefa é uma etapa fundamental para a eficiência em produção.

### Ferramentas integradas e chamadas de função (*Function Calling*)

O mecanismo de *function calling* permite que os modelos Gemini interajam com sistemas externos, bancos de dados e APIs. O Gemini suporta tanto ferramentas prontas e integradas (como `google_search` e `code_execution`) quanto funções customizadas declaradas na sua própria aplicação.

O uso de ferramentas atende a três necessidades fundamentais:
- **Executar ações:** Interagir com serviços externos via API (agendar compromissos, enviar e-mails, emitir pedidos ou controlar dispositivos IoT).
- **Enriquecer o conhecimento:** Consultar informações dinâmicas ou confidenciais em bancos de dados corporativos e sistemas internos.
- **Estender capacidades:** Realizar cálculos exatos, conversões de formato ou compilação de relatórios estruturados que fogem ao escopo estocástico de um LLM.

#### Como funciona o ciclo de *function calling*

A comunicação ocorre em 4 passos entre o seu código e o modelo:

1. **Declaração das ferramentas**: Sua aplicação descreve as funções disponíveis (nome, descrição clara e os JSON Schemas dos parâmetros) na configuração da requisição.
2. **Identificação da intenção**: O modelo avalia o prompt junto às declarações recebidas. Se o uso de uma ferramenta for necessário, ele retorna uma chamada estruturada com o nome da função e os argumentos calculados.
3. **Execução local**: O modelo *não* executa o código diretamente. Sua aplicação recebe a requisição estruturada, roda a lógica correspondente no seu ambiente e obtém o resultado.
4. **Envio do resultado**: O retorno da função é enviado de volta ao modelo em uma etapa de resultado. O modelo utiliza essa informação para compor a resposta final em linguagem natural ou para encadear novas chamadas, se necessário.

### Saídas estruturadas (*Structured Outputs*)

Você pode restringir as respostas dos modelos Gemini para que sigam rigorosamente um [JSON Schema](https://ai.google.dev/gemini-api/docs/structured-output) fornecido. Isso elimina a fragilidade do tratamento de texto livre, garantindo que o retorno seja desserializado diretamente nas estruturas de dados da sua aplicação.

Além de trabalhar com JSON Schemas manuais em requisições REST, os SDKs do Google GenAI permitem declarar esquemas diretamente em código nativo, utilizando *struct tags* em Go ou [Pydantic](https://docs.pydantic.dev/) em Python.

## Consumindo modelos via código

Após conhecermos os modelos e seus recursos, vejamos como integrar essas APIs em projetos Go.

Assim como há modelos dedicados a diferentes necessidades, existem também diferentes superfícies de API para consumi-los:

### Generate Content API

É a [API generativa](https://ai.google.dev/api/generate-content#method:-models.generatecontent) fundamental. Opera de maneira *stateless* (sem estado), recebendo uma requisição isolada e devolvendo a resposta correspondente. Para fluxos de conversa com múltiplos turnos, sua aplicação precisa reenviar o histórico completo a cada chamada.

Essa abordagem exige gerenciar ativamente o histórico para não estourar a janela de contexto, resumindo mensagens antigas quando necessário. Para reduzir custos de entrada em sessões longas, a API do Gemini oferece suporte nativo a [caching implícito](https://ai.google.dev/gemini-api/docs/caching) em todos os modelos desde o Gemini 2.5, além de [caching explícito](https://ai.google.dev/gemini-api/docs/generate-content/caching) para grandes volumes de dados.

Apesar de muito prática para chamadas simples, a Generate Content API vem sendo complementada pela moderna Interactions API.

### Interactions API

> **Nota**: No momento, a Interactions API ainda não está disponível no Go GenAI SDK oficial. O andamento da implementação pode ser acompanhado nesta [issue no GitHub](https://github.com/googleapis/go-genai/issues/658).

A [Interactions API](https://ai.google.dev/gemini-api/docs/interactions-overview) é a interface unificada do Google pensada para todos os tipos de interação — desde conversas simples e chamadas de ferramentas até orquestrações agentivas complexas —, gerenciando o histórico de conversas diretamente no servidor.

### Live API

A [Live API](https://ai.google.dev/gemini-api/docs/live-api) viabiliza conversas bidirecionais de áudio e vídeo com baixa latência via WebSockets. O sistema reconhece automaticamente quando a usuária fala ou interrompe a resposta, criando uma dinâmica natural de conversa com suporte a chamadas de ferramentas durante a própria sessão.

### Batch API

A [Batch API](https://ai.google.dev/gemini-api/docs/batch-api) permite processar grandes volumes de dados em lote de forma assíncrona, com 50% de desconto no custo dos tokens. As tarefas são executadas em segundo plano fora dos horários de pico (geralmente concluídas em até 24 horas), sendo a escolha ideal para processamentos não urgentes.

### Managed Agents API

Os [agentes gerenciados](https://ai.google.dev/gemini-api/docs/agents) oferecem um ambiente de execução totalmente hospedado na nuvem, onde agentes autônomos planejam e executam tarefas. Uma única chamada de API provisiona um sandbox Linux isolado com runtimes pré-configurados (como Python e Node), permitindo ao agente rodar código, gerenciar arquivos e pesquisar na web.

O Google disponibiliza dois agentes gerenciados prontos para uso:
- **Antigravity Agent** (`antigravity-preview-05-2026`): Agente geral padrão baseado no Gemini 3.6 Flash (configurável para Gemini 3.5 Flash ou Flash-Lite) para desenvolvimento de software, manipulação de arquivos e acesso à rede.
- **Deep Research Agent** (`deep-research-preview-04-2026`): Agente autônomo para pesquisa aprofundada, capaz de sintetizar relatórios detalhados a partir de múltiplas fontes na web.

Você também pode customizar o agente Antigravity declarando regras inline, anexando um arquivo `AGENTS.md`, vinculando diretórios com skills (`SKILL.md`) ou montando arquivos locais, repositórios Git e buckets do Cloud Storage diretamente no diretório de trabalho remoto (`/workspace`).

## Formas de acesso e faturamento

Ao utilizar os modelos Gemini, o Google disponibiliza duas modalidades principais de autenticação e faturamento:

1. **Google AI Studio (Google AI)**: Utiliza chaves de API simples (`GEMINI_API_KEY` ou `GOOGLE_API_KEY`). Perfeito para prototipagem rápida, projetos pessoais, aplicativos independentes e testes de conceito.
2. **Gemini Enterprise (antigo Vertex AI)**: Roteia o tráfego pela infraestrutura do Google Cloud utilizando IAM, Application Default Credentials (ADC), contas de serviço ou tokens OAuth 2.0. Indicado para ambientes corporativos que exigem SLAs formais, governança de dados, isolamento de rede e descontos por uso contínuo.

## Go GenAI SDK

Vejamos agora como colocar tudo isso em prática com código Go.

O pacote oficial para integração com o Gemini em Go é o [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai).

Ele é frequentemente chamado de SDK "unificado", pois foi desenvolvido para atender a todos os modelos do ecossistema e suportar autenticação tanto via Google AI quanto pelo Gemini Enterprise. Ele substitui o antigo pacote `github.com/google/generative-ai-go`, que foi descontinuado (*deprecated*).

Para adicioná-lo ao seu projeto:

```bash
go get google.golang.org/genai
```

Aqui está um exemplo completo utilizando autenticação corporativa com o Gemini Enterprise:

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

Para executar este exemplo, você precisará de um projeto no Google Cloud com a API AI Platform ativada. Por exemplo:

```sh
export GOOGLE_CLOUD_PROJECT="your-project-id-goes-here"
go run main.go
```

Veja o resultado gerado:

![Saída de imagem do gato mago gerada no terminal Go](image.png "O verdadeiro propósito da IA: geração infinita de fotos de gatinhos")

Este exemplo básico demonstra os primeiros passos com o SDK. Ao longo da nossa série, exploraremos integrações mais sofisticadas tanto com o Go GenAI SDK quanto com frameworks de orquestração de alto nível, como o [Genkit](https://genkit.dev/) e o [Agent Development Kit (ADK)](https://adk.dev/).

## O que vem a seguir?

Na [**Parte 2: Programando com o Gemini**]({{< ref "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini" >}}) da série **Gemini para Desenvolvedores Go**, vamos nos aprofundar nos agentes de programação e em como estruturar seu ambiente para acelerar o desenvolvimento em bases de código Go. Não perca!
