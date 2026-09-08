---
author: Daniela Petruzalek
categories:
- Agentic Coding
date: 2025-08-05
summary: A história do Speedgrapher, um servidor MCP personalizado para 'vibe writing'.
  A jornada de transformar uma coleção pessoal de prompts em um kit de ferramentas
  portátil e potencializado por IA para automatizar e estruturar o processo criativo.
tags:
  - ai
  - gemini-cli
  - golang
  - mcp
  - vibe-coding
title: "Apresentando o Speedgrapher: Um Servidor MCP para Vibe Writing"
slug: "introducing-speedgrapher"
aliases:
  - "/pt-br/posts/20250805-introducing-speedgrapher/"
description: "Veja como o Speedgrapher usa prompts e ferramentas do Model Context Protocol para automatizar escrita técnica, pontuação Gunning Fog e revisão editorial."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "Gemini CLI / Antigravity CLI"
---

## Introdução

Tenho uma confissão a fazer: adoro criar coisas, mas nem sempre gosto do boilerplate e do trabalho mecânico que vêm junto. Frequentemente tenho muitas ideias para novos artigos, mas o processo de estruturá-los, garantir que atendam aos meus próprios padrões editoriais e até acertar o tom pode, às vezes, parecer um fardo. Esta é a história de como um mergulho profundo em uma especificação técnica me levou a construir o [Speedgrapher](https://github.com/danicat/speedgrapher), um servidor MCP que me ajuda a trazer uma camada bem-vinda de estrutura ao meu processo de escrita.

A jornada do Speedgrapher começou logo após eu publicar meu artigo anterior, "[Construindo o GoDoctor: Um Servidor MCP com Gemini CLI e Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}})". Naquele post, foquei totalmente em como o [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) permite que agentes de IA usem ferramentas. Depois de publicá-lo, voltei à especificação do MCP para mais uma leitura. Desta vez, um pequeno detalhe que eu havia deixado passar me chamou a atenção: além de `tools`, o protocolo também define explicitamente `prompts` e `resources`. Uma lâmpada se acendeu: percebi que a coleção de prompts que eu tinha espalhada pelas minhas anotações, arquivos e repositórios do GitHub poderia ser empacotada e tornada portátil usando esse mesmo protocolo.

Em uma feliz coincidência, no mesmo dia em que eu explorava a ideia de um servidor de prompts, a equipe da Gemini CLI anunciou uma nova funcionalidade que disponibiliza prompts expostos por servidores MCP como [slash commands nativos](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/). Isso significava que a minha ideia de um toolkit portátil de backend poderia ganhar uma interface de primeira linha, super prática, direto no terminal. O conceito do Speedgrapher ficou claro: um servidor MCP dedicado a hospedar um kit de ferramentas de escrita, acessível por meio de simples slash commands.

## Vibe writing explicado

Antes de mergulharmos nos detalhes técnicos da construção do Speedgrapher, quero tirar um momento para explicar o que quero dizer com "vibe writing". Você provavelmente já ouviu falar do termo "vibe coding" por aí — ele descreve a prática cada vez mais comum de desenvolvedores usando prompts em linguagem natural para guiar uma IA na geração de código. É uma abordagem fluida e conversacional, onde a pessoa desenvolvedora define a direção de alto nível e a IA cuida do boilerplate e dos detalhes de implementação.

"Vibe writing" é a extensão natural desse conceito para o universo das palavras. Para mim, trata-se de transformar o ato solitário da escrita em uma conversa dinâmica e colaborativa com uma parceira de IA. Em vez de travar na mecânica da estrutura das frases, na gramática ou na busca pela palavra perfeita, posso focar na mensagem central — na "vibe" que quero criar. Eu forneço a faísca inicial — uma ideia bruta, uma história pessoal, um problema frustrante — e a IA me ajuda a moldá-la em uma narrativa estruturada e coerente.

Embora eu não seja a primeira pessoa a usar esse termo, trata-se de um conceito ainda emergente. Ele representa uma mudança fundamental na forma como abordamos a criação de conteúdo, evoluindo de um processo puramente manual para uma parceria entre humanos e IA.

## Começando simples: um gerador de haiku

Toda boa jornada técnica começa com um "Hello, World". Para o Speedgrapher, o meu "Hello, World" foi um haiku. Eu precisava de uma forma simples e de baixo risco para provar que conseguiria expor um prompt como slash command. O que poderia ser mais simples do que pedir para uma IA escrever um poema?

Minha primeira tentativa foi ingênua. Criei um prompt `/haiku` que recebia um argumento `--theme`. O prompt em si era simples: `"generate a haiku based on the theme %s"`. Iniciei a Gemini CLI com o meu projeto do Speedgrapher carregado como contexto e digitei:

`/haiku --theme=flowers`

O resultado foi... nada de poema. O modelo, ao ver o código Go no meu projeto, interpretou meu pedido como uma instrução para *adicionar uma funcionalidade de haiku ao Speedgrapher*. Ele começou a planejar edições nos meus arquivos Go. Pressionei `ESC` rapidamente para abortar, pois precisava repensar minha estratégia.

Essa experiência foi um lembrete marcante de um princípio essencial da engenharia de prompts: a necessidade de equilibrar ambiguidade e contexto. Em muitos dos meus prompts, uso intencionalmente certo grau de ambiguidade para dar ao modelo a flexibilidade de raciocinar e inferir informações. Por exemplo, meu prompt `/review` simplesmente diz para "revisar o artigo no qual estivemos trabalhando". Ele não especifica um nome de arquivo rígido como `DRAFT.md`. Essa ambiguidade é uma ferramenta poderosa em um fluxo de trabalho conversacional, pois permite ao modelo identificar o texto relevante a partir das nossas interações recentes, sem precisar de um caminho de arquivo explícito e rígido.

No caso do haiku, porém, a ambiguidade não tinha limites. O contexto principal era um projeto Go, o que levou o modelo a uma conclusão lógica, porém incorreta: a de que eu queria modificar o código. Ele não estava errado; estava apenas fazendo uma inferência plausível. Como eu queria um resultado bastante específico e sem relação com código, minha tarefa nesse caso foi reduzir a ambiguidade fornecendo um contexto muito mais claro sobre a minha intenção.

Depois de mais algumas tentativas, cheguei ao seguinte prompt:

```go
// The final, working prompt for the haiku command.
prompt = fmt.Sprintf("The user wants to have some fun and has requested a haiku about the following topic: %s", topic)
```

Embora eu não tenha certeza se essa é a melhor forma de expressar minha intenção, ela atendeu ao meu propósito, e o modelo passou a produzir haikus consistentemente depois disso. Com o conceito central comprovado, eu estava pronta para construir prompts mais práticos.

## Construindo um kit de ferramentas de escrita

O experimento com o haiku confirmou que o conceito central era sólido, então avancei para aplicações mais práticas. Meus arquivos `GEMINI.md` tinham se tornado uma coleção de prompts úteis, mas não portáteis, para tarefas como revisar, traduzir e criar outlines para os meus artigos. Por estarem vinculados a projetos específicos, eu frequentemente esquecia de copiá-los para novos projetos. Um servidor MCP era o próximo passo lógico para tornar essas ferramentas portáteis.

Comecei migrando três dos meus prompts mais usados para o Speedgrapher: `interview`, `review` e `localize`. O núcleo desses prompts é um conjunto de "diretrizes editoriais". Por exemplo, a diretriz de localização inclui uma regra para não traduzir termos técnicos, o que ajuda a garantir consistência entre os três idiomas suportados pelo meu blog. Essa abordagem de criar "diretrizes editoriais como código" é uma forma de construir um sistema estruturado que mantém tom de voz e qualidade consistentes, de maneira muito semelhante ao que um linter faz pelo código.

Todos os prompts do Speedgrapher foram gerados com a ajuda do Gemini, mas para o prompt `review` adotei uma abordagem ligeiramente diferente. Pedi ao modelo que analisasse meus artigos anteriores e gerasse um conjunto de diretrizes editoriais com base no meu estilo de escrita. O resultado foi um primeiro rascunho sólido, mas é um prompt que estou constantemente refinando.

Aqui está a versão atual do prompt, incorporada diretamente do código-fonte do Speedgrapher no GitHub:

{{< github user="danicat" repo="speedgrapher" path="internal/prompts/review.go" start="18" end="28" >}}

Com os prompts centrais definidos, chegou o momento de trabalhar na automação de outras partes importantes do meu trabalho.

## A legibilidade importa

Como escritora técnica, meu maior desafio é encontrar o ponto de equilíbrio ideal entre clareza e complexidade. Se um texto for simples demais, pode parecer infantil. Se for excessivamente complexo, torna-se ilegível. Legibilidade não se resume a tornar as coisas fáceis; trata-se de torná-las envolventes e intelectualmente estimulantes.

A boa notícia sobre a legibilidade é que ela pode ser medida. Embora nenhuma métrica seja perfeita, o [Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index) é uma excelente ferramenta para se obter uma linha de base. O Gunning Fog Index é um teste de legibilidade que estima os anos de educação formal que uma pessoa necessita para compreender um texto na primeira leitura. Uma pontuação de 12, por exemplo, indica que o texto está no nível de leitura de um estudante do último ano do ensino médio nos Estados Unidos.

O índice é calculado com base no seguinte algoritmo:
*   Selecione um trecho de texto com 100 ou mais palavras.
*   Calcule o comprimento médio das frases.
*   Conte a quantidade de palavras "complexas" (palavras com três ou mais sílabas).
*   Some o comprimento médio das frases à porcentagem de palavras complexas.
*   Multiplique o resultado por 0,4.

Ou, para quem tem inclinação matemática, esse algoritmo se traduz na seguinte equação:

{{< katex >}}
\[
 0.4 \times \left[ \left( \frac{\text{palavras}}{\text{frases}} \right) + 100 \left( \frac{\text{palavras complexas}}{\text{palavras}} \right) \right]
\]

Embora o propósito original do Fog Index seja estimar os anos de escolaridade necessários para compreender o texto, considero pouco prático enquadrá-lo especificamente nesses termos. Por isso, tomei a liberdade de customizá-lo para as minhas próprias necessidades. Primeiro, simplifiquei o cálculo ignorando casos especiais: uma das partes mais complexas do algoritmo é definir o que qualifica uma palavra como complexa (sem trocadilhos). Embora a regra básica considere uma palavra complexa se ela tiver três ou mais sílabas, ela cria casos especiais em que certas terminações de palavras, como *-ing*, *-ed* e *-es*, são ignoradas.

Isso causou uma quantidade surpreendente de problemas durante a implementação. Eu não precisava de precisão exata e estava satisfeita em superestimar a complexidade em nome da simplicidade. Para isso, ignorei todos os casos especiais e considerei duas regras básicas para a contagem de sílabas: 1) o número de sílabas em uma palavra é estimado pela quantidade de grupos de vogais, e 2) palavras complexas são palavras que possuem três ou mais sílabas (sem exceções).

Também criei um sistema de classificação que desloca o foco dos anos de escolaridade para uma abordagem mais pragmática de legibilidade:

| Pontuação | Classificação | Descrição |
| :--- | :--- | :--- |
| >= 22 | Ilegível | Provavelmente incompreensível para a maioria dos leitores. |
| 18-21 | Difícil de ler | Exige esforço significativo, mesmo para especialistas. |
| 13-17 | Público profissional | Ideal para leitores com conhecimento especializado. |
| 9-12 | Público geral | Claro e acessível para a maior parte dos leitores. |
| < 9 | Simplista | Pode ser percebido como infantil ou excessivamente simples. |

Com o Gunning Fog Index customizado implementado como uma ferramenta `fog`, o passo final foi criar uma interface amigável para ele. Criei o prompt `/readability`, que invoca a ferramenta `fog` e apresenta os resultados em um formato claro. Isso segue as minhas diretrizes de design para o Speedgrapher: construir ferramentas focadas, de propósito único, e então combiná-las em fluxos de trabalho mais poderosos e fáceis de usar.

## Automatizando o fluxo de trabalho de escrita

Os prompts individuais já ajudavam bastante, mas eu ainda tinha muito a automatizar até alcançar o fluxo de trabalho dos meus sonhos. Nas iterações seguintes, testei os prompts no dia a dia e mapeei as lacunas do processo para propor novos prompts e/ou ajustar os já existentes. Aqui estão os prompts que utilizo atualmente:

**Fluxo principal**
* `/interview`: Entrevista o autor para coletar material para um artigo. Costuma ser o ponto de partida de uma sessão de escrita.
* `/outline`: Gera um outline estruturado a partir do rascunho atual, conceito ou transcrição da entrevista.
* `/voice`: Analisa a voz e o tom da escrita da pessoa usuária para replicá-los no texto gerado.
* `/expand`: Expande um outline ou rascunho de trabalho em um artigo mais detalhado. Também pode ser usado com o argumento `hint` para fazer uma expansão focada de um parágrafo ou seção específica.
* `/review`: Revisa o artigo atualmente em produção com base nas diretrizes editoriais.
* `/readability`: Analisa o último texto gerado quanto à legibilidade usando o Gunning Fog Index.
* `/localize`: Traduz o artigo atualmente em produção para um idioma de destino.
* `/publish`: Publica a versão final do artigo.

**Opcional**
* `/context`: Carrega o artigo em andamento no contexto para comandos subsequentes. É usado para "lembrar" o modelo sobre o rascunho atual se necessário, e frequentemente executado antes de comandos como `/readability` ou `/review`, que operam sobre o texto completo.
* `/reflect`: Analisa a sessão atual e propõe melhorias para o processo de escrita. Muito útil para aperfeiçoar prompts e diretrizes editoriais.

O objetivo era evoluir de uma coleção de comandos úteis para um processo único e otimizado, capaz de guiar um artigo desde uma ideia simples até uma publicação multilíngue e refinada.

O diagrama abaixo é uma representação simplificada do meu fluxo de trabalho:

{{< mermaid >}}
flowchart TD
    A[Ideia] -->|/interview| B[Transcrição da Entrevista]
    B -->|/outline & /voice| C[Outline Estruturado]
    C -->|/expand| D[Rascunho do Artigo]
    D -->|/review & /readability| E[Rascunho Revisado]
    E -->|/localize| F[Versões Traduzidas]
    F -->|/publish| G[Artigo Publicado]
{{< /mermaid >}}

O processo começa com um `/interview` para desenvolver os conceitos centrais de uma ideia. A transcrição resultante é então transformada em um plano estruturado via `/outline` e alinhada ao meu estilo pessoal de escrita com o `/voice`. Com essa base sólida, entro em um ciclo iterativo usando `/expand` para desenvolver o rascunho, e `/review` e `/readability` para refiná-lo.

Quando o artigo é aprovado, utilizo `/localize` para criar versões em outros idiomas e `/publish` para concluir o processo. O prompt opcional `/reflect` pode ser usado para analisar a sessão e gerar notas para melhorias futuras, criando um ciclo de refinamento contínuo.

## Conclusão

Assim como usamos linters e testes para trazer estrutura ao nosso código, podemos aplicar princípios semelhantes aos nossos fluxos de trabalho criativos. O processo de escrita envolve muitas tarefas repetitivas que podem ser automatizadas. Ao construir um toolkit pessoal de prompts, podemos delegar o boilerplate e focar nas ideias centrais do nosso trabalho.

Esse é o verdadeiro valor de uma ferramenta como o Speedgrapher: "vibe writing" não é sobre substituir quem escreve, mas sobre aumentar o processo de escrita. Ao adicionar um servidor MCP ao fluxo, ganhamos uma camada útil de estrutura em um fluxo de trabalho que muitas vezes pode ser desorganizado, garantindo que as melhores práticas sejam seguidas. O mesmo pode ser aplicado a qualquer processo assistido por IA: ao tratar seus próprios prompts como ativos reutilizáveis e portáteis, você cria um sistema que evolui junto com o seu processo, permitindo que você se concentre nos aspectos criativos do seu trabalho, um prompt de cada vez.

## O que vem a seguir?

A jornada com o Speedgrapher está longe de terminar. Embora o toolkit atual seja focado em texto, o próximo passo lógico é abraçar a multimodalidade. Estou explorando como integrar ferramentas para geração de hero images, criação de diagramas mais sofisticados a partir do texto e até sugestões de otimização de layout. O objetivo é continuar construindo um toolkit pessoal que lide com mais tarefas paralelas à escrita, liberando-me para focar no conteúdo em si.

## Recursos

*   **[Projeto Speedgrapher](https://github.com/danicat/speedgrapher):** O código-fonte do servidor MCP discutido neste artigo.
*   **[Construindo o GoDoctor: Um Servidor MCP com Gemini CLI e Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}}):** O artigo anterior que inspirou esta jornada.
*   **[Model Context Protocol (MCP)](https://modelcontextprotocol.io/):** O site oficial do protocolo.
*   **[Anúncio da Gemini CLI](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/):** O post no blog que anunciou o suporte a slash commands customizados.
*   **[Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index):** Saiba mais sobre a métrica de legibilidade.
