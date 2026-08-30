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

Tenho uma confissão a fazer: adoro criar coisas, mas nem sempre gosto do boilerplate e do trabalho mecânico que vêm junto. Frequentemente tenho dezenas de ideias para novos artigos, mas estruturá-los, garantir que atendam aos meus próprios padrões editoriais e até acertar o tom pode, às vezes, parecer um fardo. Esta é a história de como um mergulho profundo em uma especificação técnica me levou a construir o [Speedgrapher](https://github.com/danicat/speedgrapher), um servidor MCP que me ajuda a trazer uma camada bem-vinda de estrutura ao meu processo de escrita.

A jornada do Speedgrapher começou logo após eu publicar meu artigo anterior, "[Como Construir um Servidor MCP com Gemini CLI e Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}})". Naquele post, foquei totalmente em como o [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) permite que agentes de IA usem ferramentas (tools). Depois de publicá-lo, voltei à especificação do MCP para mais uma leitura. Desta vez, um detalhe que eu havia deixado passar me chamou a atenção: além de `tools`, o protocolo também define explicitamente `prompts` e `resources`. Foi quando caiu a ficha: percebi que aquela coleção de prompts espalhada pelas minhas anotações, arquivos e repositórios do GitHub poderia ser empacotada e tornada portátil usando esse mesmo protocolo.

Em uma feliz coincidência, no mesmo dia em que eu explorava a ideia de um servidor de prompts, a equipe do Gemini CLI anunciou uma nova funcionalidade que disponibiliza prompts expostos por servidores MCP como [slash commands nativos](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/). Isso significava que a minha ideia de um toolkit portátil de backend poderia ganhar uma interface de primeira linha, super prática, direto no terminal. O conceito do Speedgrapher ficou claro: um servidor MCP dedicado a hospedar um kit de ferramentas de escrita, acessível por meio de simples slash commands.

## Vibe Writing Explicado

Antes de mergulharmos nos detalhes técnicos da construção do Speedgrapher, vale explicar o que quero dizer com "vibe writing". Você provavelmente já ouviu falar de "vibe coding" — termo que descreve o hábito cada vez mais comum de guiar uma IA por meio de prompts em linguagem natural para gerar código. É uma abordagem fluida e conversacional, em que a pessoa desenvolvedora define a direção estratégica e a IA cuida do boilerplate e dos detalhes de implementação.

"Vibe writing" é a extensão natural desse conceito para o universo das palavras. Para mim, trata-se de transformar o ato solitário da escrita em uma conversa dinâmica e colaborativa com uma parceira de IA. Em vez de travar na mecânica da estrutura das frases, na gramática ou na busca pela palavra perfeita, posso focar na mensagem central — na "vibe" que quero transmitir. Eu forneço a faísca inicial — uma ideia bruta, uma história pessoal, um problema frustrante — e a IA me ajuda a moldá-la em uma narrativa estruturada e coerente.

Embora eu não seja a primeira pessoa a usar esse termo, trata-se de um conceito ainda emergente. Ele representa uma virada fundamental na forma como abordamos a criação de conteúdo, evoluindo de um processo puramente manual para uma parceria entre humanos e IA.

## Começando Simples: Um Gerador de Haiku

Toda boa jornada técnica começa com um "Hello, World". No caso do Speedgrapher, o meu "Hello, World" foi um haiku. Eu precisava de uma forma simples e sem riscos de provar que era capaz de expor um prompt como slash command. E o que poderia ser mais simples do que pedir para uma IA escrever um poema?

Minha primeira tentativa foi ingênua. Criei um prompt `/haiku` que recebia um argumento `--theme`. O prompt em si era direto: `"generate a haiku based on the theme %s"`. Abri o Gemini CLI com o projeto do Speedgrapher carregado no contexto e digitei:

`/haiku --theme=flowers`

O resultado... não foi um poema. O modelo, ao ver o código Go no projeto, interpretou meu pedido como uma instrução para *adicionar uma funcionalidade de haiku ao Speedgrapher*. Ele começou a planejar edições nos meus arquivos Go. Apertei `ESC` imediatamente para abortar e repensar minha estratégia.

Essa experiência serviu como um ótimo lembrete de um princípio essencial da engenharia de prompts: o equilíbrio entre ambiguidade e contexto. Em vários dos meus prompts, uso intencionalmente certo grau de ambiguidade para dar ao modelo flexibilidade para raciocinar e inferir informações. Por exemplo, meu prompt `/review` diz simplesmente "revise o artigo no qual estivemos trabalhando". Ele não especifica um nome de arquivo rígido como `DRAFT.md`. Essa ambiguidade é um recurso poderoso em fluxos conversacionais, pois permite ao modelo identificar o texto relevante a partir das nossas interações recentes sem depender de um caminho de arquivo explícito.

No caso do haiku, porém, a ambiguidade não tinha freios. O contexto principal era um projeto Go, o que levou o modelo a uma dedução lógica, mas errada: a de que eu queria alterar o código-fonte. Ele não estava "errado"; apenas fez uma inferência plausível. Como eu queria um resultado bastante específico e sem relação com código, meu trabalho foi reduzir a ambiguidade fornecendo um contexto muito mais claro sobre a minha intenção.

Após algumas tentativas, cheguei ao seguinte prompt:

```go
// The final, working prompt for the haiku command.
prompt = fmt.Sprintf("The user wants to have some fun and has requested a haiku about the following topic: %s", topic)
```

Não sei se essa é a melhor redação possível para expressar minha intenção, mas ela funcionou perfeitamente para o meu objetivo, e o modelo passou a gerar haikus de forma consistente. Com a prova de conceito validada, eu estava pronta para criar prompts mais práticos.

## Construindo um Toolkit de Escrita

O experimento com o haiku confirmou que a ideia central funcionava, então parti para aplicações mais práticas. Meus arquivos `GEMINI.md` tinham virado uma coleção de prompts úteis, porém nada portáteis, para tarefas como revisar, traduzir e estruturar artigos. Por estarem atrelados a projetos específicos, eu vivia esquecendo de copiá-los para novos repositórios. Um servidor MCP era o próximo passo lógico para tornar essas ferramentas portáteis.

Comecei migrando três dos meus prompts mais usados para o Speedgrapher: `interview`, `review` e `localize`. O coração desses prompts é um conjunto de "diretrizes editoriais". Por exemplo, a diretriz de localização inclui uma regra para nunca traduzir termos técnicos, garantindo consistência entre os três idiomas que meu blog suporta. Essa abordagem de "diretrizes editoriais como código" permite construir um sistema estruturado que preserva tom de voz e qualidade constantes — exatamente como um linter faz com o código.

Todos os prompts do Speedgrapher foram gerados com a ajuda do Gemini, mas no caso do `review` adotei uma estratégia um pouco diferente: pedi ao modelo que analisasse meus artigos anteriores e gerasse um conjunto de diretrizes editoriais baseado no meu estilo de escrita. O resultado foi um ótimo primeiro rascunho, e sigo refinando esse prompt continuamente.

Aqui está a versão atual do prompt, incorporada diretamente do código-fonte do Speedgrapher no GitHub:

{{< github user="danicat" repo="speedgrapher" path="internal/prompts/review.go" start="18" end="28" >}}

Com os prompts fundamentais prontos, chegou a hora de automatizar outras partes importantes do meu fluxo.

## A Legibilidade Importa

Como escritora técnica, meu maior desafio é encontrar o ponto ideal entre clareza e complexidade. Se o texto for simples demais, soa infantil; se for excessivamente complexo, torna-se ilegível. Legibilidade não é apenas simplificar tudo, mas tornar o conteúdo envolvente e intelectualmente estimulante.

A boa notícia é que legibilidade pode ser medida. Embora nenhuma métrica seja perfeita, o [Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index) (Índice de Nebulosidade de Gunning) é uma ótima ferramenta para estabelecer uma linha de base. Ele estima os anos de educação formal necessários para que alguém compreenda um texto em uma primeira leitura. Uma pontuação de 12, por exemplo, indica um nível correspondente ao último ano do ensino médio nos EUA.

O índice é calculado com base no seguinte algoritmo:
*   Selecione um trecho de texto com 100 ou mais palavras.
*   Calcule o comprimento médio das frases.
*   Conte a quantidade de palavras "complexas" (palavras com três ou mais sílabas).
*   Some o comprimento médio das frases à porcentagem de palavras complexas.
*   Multiplique o resultado por 0,4.

Para quem prefere a formulação matemática, esse algoritmo se traduz na seguinte equação:

{{< katex >}}
\[
 0.4 \times \left[ \left( \frac{\text{palavras}}{\text{frases}} \right) + 100 \left( \frac{\text{palavras complexas}}{\text{palavras}} \right) \right]
\]

Embora o objetivo original do Fog Index seja estimar anos de escolaridade, acho pouco prático enquadrá-lo dessa forma. Por isso, tomei a liberdade de customizá-lo para as minhas próprias necessidades. Primeiro, simplifiquei o cálculo ignorando casos especiais: uma das partes mais complexas do algoritmo original é definir o que conta como palavra complexa (sem trocadilhos). A regra geral considera complexa qualquer palavra com três ou mais sílabas, mas abre exceções para certos sufixos como *-ing*, *-ed* e *-es*.

Essas exceções causaram uma quantidade surpreendente de dores de cabeça na implementação. Como eu não precisava de precisão acadêmica e preferia superestimar a complexidade em nome da simplicidade, decidi ignorar todas as exceções e adotar duas regras básicas para contagem de sílabas: 1) o número de sílabas em uma palavra é estimado pela quantidade de grupos de vogais, e 2) palavras complexas são simplesmente aquelas com três ou mais sílabas (sem exceções).

Também criei um sistema de classificação que desloca o foco dos anos de estudo para uma abordagem mais pragmática de legibilidade:

| Pontuação | Classificação | Descrição |
| :--- | :--- | :--- |
| >= 22 | Ilegível | Provavelmente incompreensível para a maioria das pessoas. |
| 18-21 | Difícil de Ler | Exige esforço significativo, mesmo de especialistas. |
| 13-17 | Público Profissional | Ideal para leitores com conhecimento especializado. |
| 9-12 | Público Geral | Claro e acessível para a maior parte do público. |
| < 9 | Simplista | Pode soar infantil ou excessivamente básico. |

Com o Gunning Fog Index customizado implementado na ferramenta `fog`, o passo final foi criar uma interface amigável para ele. Desenvolvi o prompt `/readability`, que invoca a ferramenta `fog` e apresenta os resultados de maneira clara. Isso segue a filosofia de design do Speedgrapher: construir ferramentas enxutas, de propósito único, e combiná-las em fluxos de trabalho mais poderosos e agradáveis de usar.

## Automatizando o Fluxo de Escrita

Os prompts individuais já ajudavam muito, mas eu ainda tinha bastante trabalho para alcançar o fluxo dos meus sonhos. Nas iterações seguintes, testei os prompts no dia a dia e mapeei gargalos do processo para criar novos comandos ou calibrar os existentes. Aqui estão os prompts que utilizo atualmente:

**Fluxo Principal**
* `/interview`: Entrevista a autora para coletar ideias e material para um artigo. Costuma ser o ponto de partida de uma sessão de escrita.
* `/outline`: Gera uma estrutura detalhada (outline) a partir do rascunho atual, conceito ou relatório de entrevista.
* `/voice`: Analisa o tom de voz e o estilo de escrita da usuária para replicá-los nos textos gerados.
* `/expand`: Expande um outline ou rascunho inicial em um artigo mais completo. Também aceita um argumento `hint` para expandir um parágrafo ou seção específica com foco direcionado.
* `/review`: Revisa o artigo em andamento com base nas diretrizes editoriais.
* `/readability`: Avalia a legibilidade do último texto gerado utilizando o Gunning Fog Index.
* `/localize`: Traduz o artigo em andamento para um idioma de destino.
* `/publish`: Publica a versão final do artigo.

**Opcionais**
* `/context`: Carrega o artigo em andamento no contexto para comandos subsequentes. Serve para "relembrar" o modelo sobre o rascunho atual, sendo bastante útil antes de executar comandos que operam no texto completo, como `/readability` ou `/review`.
* `/reflect`: Analisa a sessão de trabalho e sugere melhorias para o processo de escrita — excelente para aprimorar prompts e diretrizes editoriais.

A meta era sair de um punhado de comandos isolados e chegar a um processo coeso, capaz de conduzir um artigo desde a ideia embrionária até uma publicação multilíngue e refinada.

O diagrama abaixo ilustra de forma simplificada esse fluxo de trabalho:

{{< mermaid >}}
flowchart TD
    A[Ideia] -->|/interview| B[Transcrição da Entrevista]
    B -->|/outline & /voice| C[Outline Estruturado]
    C -->|/expand| D[Rascunho do Artigo]
    D -->|/review & /readability| E[Rascunho Revisado]
    E -->|/localize| F[Versões Traduzidas]
    F -->|/publish| G[Artigo Publicado]
{{< /mermaid >}}

O processo começa com um `/interview` para extrair os conceitos fundamentais da ideia. A transcrição resultante é transformada em um plano estruturado via `/outline` e calibrada com meu estilo pessoal por meio do `/voice`. Com essa base sólida, entro em um ciclo iterativo usando `/expand` para desenvolver o rascunho, e `/review` e `/readability` para poli-lo.

Com o artigo aprovado, utilizo `/localize` para gerar versões em outros idiomas e `/publish` para concluir a publicação. O prompt opcional `/reflect` pode ser acionado para avaliar a sessão e gerar anotações de melhoria, alimentando um ciclo de evolução contínua.

## Conclusão

Assim como usamos linters e testes para dar estrutura ao código, podemos aplicar princípios parecidos aos nossos fluxos criativos. O processo de escrita envolve muitas etapas mecânicas que podem perfeitamente ser automatizadas. Ao construir um toolkit pessoal de prompts, conseguimos delegar o trabalho repetitivo e nos concentrar no que realmente importa: as ideias.

Esse é o verdadeiro valor de uma ferramenta como o Speedgrapher: o "vibe writing" não busca substituir quem escreve, mas sim potencializar o processo de escrita. Ao integrar um servidor MCP ao fluxo, ganhamos uma camada valiosa de organização em um processo que muitas vezes pode ser caótico, garantindo a aplicação consistente de boas práticas. O mesmo princípio vale para qualquer fluxo assistido por IA: ao tratar seus próprios prompts como ativos reutilizáveis e portáteis, você cria um sistema que evolui junto com o seu método de trabalho — permitindo focar na criatividade, um prompt de cada vez.

## O Que Vem Pela Frente?

A jornada com o Speedgrapher está apenas começando. Embora o toolkit atual seja focado em texto, o próximo passo natural é abraçar a multimodalidade. Estou explorando a integração de ferramentas para geração de imagens de capa (hero images), criação de diagramas mais sofisticados a partir do texto e até sugestões de otimização de layout. O objetivo é seguir expandindo um kit de ferramentas pessoal que assuma cada vez mais tarefas paralelas, deixando-me livre para focar exclusivamente no conteúdo.

## Recursos

*   **[Projeto Speedgrapher](https://github.com/danicat/speedgrapher):** O código-fonte do servidor MCP discutido neste artigo.
*   **[Como Construir um Servidor MCP com Gemini CLI e Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}}):** O artigo anterior que inspirou esta jornada.
*   **[Model Context Protocol (MCP)](https://modelcontextprotocol.io/):** O site oficial do protocolo.
*   **[Anúncio do Gemini CLI](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/):** O post oficial que anunciou o suporte a slash commands customizados.
*   **[Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index):** Saiba mais sobre a métrica de legibilidade.
