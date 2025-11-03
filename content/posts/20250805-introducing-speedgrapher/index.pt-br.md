---
title: "Apresentando o Speedgrapher: Um Servidor MCP para Vibe Writing"
date: 2025-08-05
author: "Daniela Petruzalek"
categories: ["Workflow & Best Practices"]
tags: ["golang", "gemini-cli", "mcp", "ai", "vibe-coding"]
summary: "A história do Speedgrapher, um servidor MCP personalizado para 'vibe writing'. Detalha a jornada de transformar uma coleção pessoal de prompts em um kit de ferramentas portátil e alimentado por IA para automatizar e estruturar o processo criativo."
---

{{< translation-notice >}}

## Introdução

Tenho uma confissão a fazer: adoro construir coisas, mas nem sempre adoro o trabalho repetitivo que vem com isso. Muitas vezes, tenho muitas ideias para novos artigos, mas o processo de estruturá-los, garantir que atendam aos meus próprios padrões editoriais e até mesmo acertar o tom pode, às vezes, parecer uma tarefa árdua. Esta é a história de como um mergulho profundo em uma especificação técnica me levou a construir o [Speedgrapher](https://github.com/danicat/speedgrapher), um servidor MCP que me ajuda a trazer uma camada útil de estrutura para o meu processo de escrita.

A jornada para o Speedgrapher começou logo depois que publiquei meu último artigo, "[Como Construir um Servidor MCP com Gemini CLI e Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}})". Naquele post, foquei inteiramente em como o [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) permite que agentes de IA usem ferramentas. Depois de publicá-lo, voltei à especificação do MCP para outra leitura. Desta vez, um pequeno detalhe que eu havia ignorado anteriormente me chamou a atenção: além de `tools`, o protocolo também define explicitamente `prompts` e `resources`. Uma lâmpada se acendeu. Percebi que a coleção de prompts que eu tinha espalhada por minhas anotações, arquivos e repositórios do GitHub poderia ser empacotada e tornada portátil usando o mesmo protocolo.

Em uma feliz coincidência, no mesmo dia em que eu estava explorando a ideia de um servidor de prompts, a equipe do Gemini CLI anunciou um novo recurso que torna os prompts expostos por servidores MCP disponíveis como [comandos de barra nativos](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/). Isso significava que minha ideia de um kit de ferramentas de backend portátil poderia ter uma interface de primeira classe e fácil de usar diretamente no meu terminal. O conceito para o Speedgrapher estava agora claro: um servidor MCP dedicado para abrigar um kit de ferramentas de escrita, exposto como simples comandos de barra.

## Vibe Writing Explicado

Antes de mergulharmos na jornada técnica de construção do Speedgrapher, quero parar um momento para explicar o que quero dizer com "vibe writing". Você provavelmente já ouviu o termo "vibe coding" por aí — ele descreve a prática cada vez mais comum de desenvolvedores usarem prompts em linguagem natural para guiar uma IA na geração de código. É uma abordagem fluida e conversacional, onde o desenvolvedor define a direção de alto nível, e a IA cuida do trabalho repetitivo e dos detalhes da implementação.

"Vibe writing" é a extensão natural desse conceito para o mundo das palavras. Para mim, trata-se de transformar o ato solitário de escrever em uma conversa dinâmica e colaborativa com um parceiro de IA. Em vez de me prender à mecânica da estrutura das frases, gramática e encontrar a palavra perfeita, posso me concentrar na mensagem principal — a "vibe" que quero criar. Eu forneço a centelha inicial — uma ideia aproximada, uma história pessoal, um problema frustrante — e a IA me ajuda a moldá-la em uma narrativa estruturada e coerente.

Embora eu não seja a primeira pessoa a usar este termo, ainda é um conceito emergente. Ele representa uma mudança fundamental em como abordamos a criação de conteúdo, passando de um processo puramente manual para uma parceria humano-IA.

## Começando Simples: Um Gerador de Haiku

Toda boa jornada técnica começa com um "Olá, Mundo". Para o Speedgrapher, meu "Olá, Mundo" foi um haiku. Eu precisava de uma maneira simples e de baixo risco para provar que poderia expor um prompt como um comando de barra. O que poderia ser mais simples do que pedir a uma IA para escrever um poema?

Minha primeira tentativa foi ingênua. Criei um prompt `/haiku` que recebia um argumento `--theme`. O prompt em si era simples: "gere um haiku baseado no tema %s". Iniciei o Gemini CLI, com meu projeto Speedgrapher carregado como contexto, e digitei:

`/haiku --theme=flowers`

O resultado foi... não um poema. O modelo, vendo o código Go no meu projeto, interpretou meu pedido como uma instrução para *adicionar um recurso de haiku ao Speedgrapher*. Ele começou a planejar a edição dos meus arquivos Go. Eu rapidamente apertei `ESC` para abortar, pois tive que repensar minha estratégia.

Essa experiência foi um lembrete poderoso de um princípio central na engenharia de prompts: a necessidade de equilibrar ambiguidade e contexto. Em muitos dos meus prompts, eu uso intencionalmente um grau de ambiguidade para dar ao modelo a flexibilidade de raciocinar e inferir informações. Por exemplo, meu prompt `/review` simplesmente diz para "revisar o artigo em que estamos trabalhando". Ele não especifica um nome de arquivo como `DRAFT.md`. Essa ambiguidade é uma ferramenta poderosa em um fluxo de trabalho conversacional, pois permite que o modelo identifique o texto relevante de nossas interações recentes sem a necessidade de um caminho de arquivo rígido e explícito.

No caso do haiku, no entanto, a ambiguidade era irrestrita. O contexto principal era um projeto Go, o que levou o modelo a uma conclusão lógica, mas incorreta: que eu queria modificar o código. Não estava errado; estava apenas fazendo uma inferência razoável. Como eu queria um resultado muito específico e não relacionado a código, minha tarefa neste caso era reduzir a ambiguidade, fornecendo um contexto muito mais claro para minha intenção.

Depois de mais algumas tentativas, cheguei ao seguinte prompt:

```go
// O prompt final e funcional para o comando haiku.
prompt = fmt.Sprintf("O usuário quer se divertir um pouco e solicitou um haiku sobre o seguinte tópico: %s", topic)
```

Embora eu não tenha certeza se esta é a melhor maneira de expressar minha intenção, ela atendeu ao meu propósito, e o modelo produziu haikus de forma consistente depois disso. Com o conceito principal comprovado, eu estava pronta para construir prompts mais práticos.

## Construindo um Kit de Ferramentas para Escritores

O experimento do haiku confirmou que o conceito principal era sólido, então passei para aplicações mais práticas. Meus arquivos `GEMINI.md` haviam se tornado uma coleção de prompts úteis, mas não portáteis, para tarefas como revisar, traduzir e delinear meus artigos. Como estavam vinculados a projetos específicos, muitas vezes eu esquecia de copiá-los para novos projetos. Um servidor MCP era o próximo passo lógico para tornar essas ferramentas portáteis.

Comecei migrando três dos meus prompts mais usados para o Speedgrapher: `interview`, `review` e `localize`. O núcleo desses prompts é um conjunto de "diretrizes editoriais". Por exemplo, a diretriz de localização inclui uma regra para não traduzir termos técnicos, o que ajuda a garantir a consistência nos três idiomas que meu blog suporta. Essa abordagem de criar "diretrizes editoriais como código" é uma maneira de construir um sistema estruturado que mantém uma voz e qualidade consistentes, muito parecido com o que um linter faz para o código.

Todos os prompts no Speedgrapher foram gerados com a ajuda do Gemini, mas para o prompt `review`, adotei uma abordagem um pouco diferente. Pedi ao modelo para analisar meus artigos anteriores e gerar um conjunto de diretrizes editoriais com base no meu estilo de escrita. O resultado foi um primeiro rascunho sólido, mas é um prompt que estou constantemente refinando.

Aqui está a versão atual do prompt, incorporada diretamente do código-fonte do Speedgrapher no GitHub:

{{< github user="danicat" repo="speedgrapher" path="internal/prompts/review.go" start="18" end="28" >}}

Com os prompts principais no lugar, era hora de trabalhar na automação de outras partes importantes do meu trabalho.

## A Legibilidade Importa

Como escritora técnica, meu maior desafio é encontrar o ponto ideal entre clareza e complexidade. Se um texto é muito simples, pode parecer infantil. Se é muito complexo, torna-se ilegível. Legibilidade não é apenas sobre tornar as coisas fáceis; é sobre torná-las envolventes e intelectualmente estimulantes.

A boa notícia sobre a legibilidade é que ela pode ser medida. Embora nenhuma métrica seja perfeita, o [Índice de Nebulosidade de Gunning (Gunning Fog Index)](https://en.wikipedia.org/wiki/Gunning_fog_index) é uma ótima ferramenta para obter uma linha de base. O Gunning Fog Index é um teste de legibilidade que estima os anos de educação formal que uma pessoa precisa para entender um texto na primeira leitura. Uma pontuação de 12, por exemplo, significa que o texto está no nível de leitura de um aluno do último ano do ensino médio dos EUA.

O índice é calculado com base no seguinte algoritmo:
*   Pegue uma seção de texto de 100 ou mais palavras.
*   Encontre o comprimento médio da frase.
*   Conte o número de palavras "complexas" (palavras com três ou mais sílabas).
*   Some o comprimento médio da frase à porcentagem de palavras complexas.
*   Multiplique o resultado por 0,4.

Ou, para os inclinados à matemática, este algoritmo se traduz na seguinte equação:

{{< katex >}}
\[
 0.4 \times \left[ \left( \frac{\text{palavras}}{\text{frases}} \right) + 100 \left( \frac{\text{palavras complexas}}{\text{palavras}} \right) \right]
\]

Embora a intenção original do índice de nebulosidade seja estimar os anos de educação necessários para entender o texto, acho que não é útil enquadrá-lo especificamente em termos de anos de educação, então tomei a liberdade de personalizá-lo para minhas próprias necessidades. Primeiro, simplifiquei o cálculo para ignorar casos especiais: uma das partes mais complexas do algoritmo é como definir se uma palavra é complexa (trocadilho intencional). Embora o caso base seja considerar uma palavra complexa se ela tiver três ou mais sílabas, ele cria casos especiais onde você ignora certas terminações de palavras como -ing, -ed e -es.

Isso criou uma quantidade surpreendente de problemas durante a implementação. Eu não precisava ser precisa e estava feliz em superestimar a complexidade em nome da simplicidade. Para este propósito, ignorei todos os casos especiais e considerei duas regras básicas para contar sílabas: 1) o número de sílabas em uma palavra é estimado pelo número de grupos de vogais, e 2) palavras complexas são palavras que têm três ou mais sílabas (sem exceções).

Também criei um sistema de classificação que muda o foco de anos de educação para uma abordagem mais pragmática da legibilidade.

| Pontuação | Classificação | Descrição |
| :--- | :--- | :--- |
| >= 22 | Ilegível | Provavelmente incompreensível para a maioria dos leitores. |
| 18-21 | Difícil de Ler | Requer esforço significativo, mesmo para especialistas. |
| 13-17 | Público Profissional | Melhor para leitores com conhecimento especializado. |
| 9-12 | Público Geral | Claro e acessível para a maioria dos leitores. |
| < 9 | Simplista | Pode ser percebido como infantil ou excessivamente simples. |

Com o Gunning Fog Index personalizado implementado como uma ferramenta `fog`, o passo final foi criar uma interface amigável para ele. Criei um prompt `/readability` que chama a ferramenta `fog` e apresenta os resultados em um formato claro. Isso segue minhas diretrizes de design para o Speedgrapher: construir ferramentas focadas e de propósito único e, em seguida, compô-las em fluxos de trabalho mais poderosos e fáceis de usar.

## Automatizando o Fluxo de Trabalho do Escritor

Os prompts individuais foram úteis, mas eu ainda tinha muito a automatizar até conseguir o fluxo de trabalho dos meus sonhos. Nas próximas iterações, eu testaria os prompts e mapearia as lacunas do processo para criar novos prompts e/ou ajustar os existentes. Aqui estão os prompts que estou usando atualmente:

**Fluxo Principal**
* `/interview`: Entrevista um autor para coletar material para um artigo. Este é geralmente o ponto de partida para uma sessão de escrita.
* `/outline`: Gera um esboço estruturado do rascunho, conceito ou relatório de entrevista atual.
* `/voice`: Analisa a voz e o tom da escrita do usuário para replicá-lo no texto gerado.
* `/expand`: Expande um esboço ou rascunho de trabalho em um artigo mais detalhado. Também pode ser usado com um argumento `hint` para fazer uma expansão focada de um parágrafo ou seção específica.
* `/review`: Revisa o artigo em que se está trabalhando atualmente em relação às diretrizes editoriais.
* `/readability`: Analisa o último texto gerado para legibilidade usando o Gunning Fog Index.
* `/localize`: Traduz o artigo em que se está trabalhando atualmente para um idioma de destino.
* `/publish`: Publica a versão final do artigo.

**Opcional**
* `/context`: Carrega o artigo em andamento atual para o contexto para comandos futuros. Isso é usado para "lembrar" o modelo do rascunho atual, se necessário, e é frequentemente executado antes de comandos como `/readability` ou `/review` que operam no texto completo.
* `/reflect`: Analisa a sessão atual e propõe melhorias no processo de escrita. Isso é útil para melhorar os prompts e as diretrizes editoriais.

O objetivo era passar de uma coleção de comandos úteis para um único processo simplificado que pudesse guiar um artigo de uma ideia simples a uma publicação polida e multilíngue.

O diagrama abaixo é uma representação simplificada do meu fluxo de trabalho:

{{< mermaid >}}
flowchart TD
    A[Ideia] -->|/interview| B[Transcrição da Entrevista]
    B -->|/outline & /voice| C[Esboço Estruturado]
    C -->|/expand| D[Rascunho do Artigo]
    D -->|/review & /readability| E[Rascunho Revisado]
    E -->|/localize| F[Versões Localizadas]
    F -->|/publish| G[Artigo Publicado]
{{< /mermaid >}}

O processo começa com uma `/interview` para detalhar os conceitos principais de uma ideia. A transcrição resultante é então transformada em um plano estruturado usando `/outline`, e alinhada com meu estilo de escrita pessoal com `/voice`. Com esta base estabelecida, entro em um ciclo iterativo de usar `/expand` para construir o rascunho, e `/review` e `/readability` para refiná-lo.

Uma vez que o artigo é aprovado, uso `/localize` para criar versões para outros idiomas, e `/publish` para finalizar o processo. O prompt opcional `/reflect` pode então ser usado para analisar a sessão e gerar notas para melhorias futuras, criando um ciclo de refinamento contínuo.

## Conclusão

Assim como usamos linters e testes para trazer estrutura ao nosso código, podemos aplicar princípios semelhantes aos nossos fluxos de trabalho criativos. O processo de escrita tem muitas tarefas repetitivas que podem ser automatizadas. Ao construir um kit de ferramentas pessoal de prompts, podemos descarregar o trabalho repetitivo e focar nas ideias centrais do nosso trabalho.

Este é o valor de uma ferramenta como o Speedgrapher: "vibe writing" não é sobre substituir o escritor, mas sobre aumentar o processo de escrita. Ao adicionar um servidor MCP à mistura, ele traz uma camada útil de estrutura para um fluxo de trabalho às vezes desorganizado, garantindo que as melhores práticas sejam seguidas. O mesmo pode ser aplicado a qualquer processo assistido por IA: ao tratar seus próprios prompts como ativos reutilizáveis e portáteis, você pode criar um sistema que evolui com seu processo, permitindo que você se concentre nos aspectos criativos do seu trabalho, um prompt de cada vez.

## O Que Vem a Seguir?

A jornada com o Speedgrapher está longe de terminar. Embora o kit de ferramentas atual esteja focado em texto, o próximo passo lógico é abraçar a multimodalidade. Estou explorando como integrar ferramentas para gerar imagens de destaque, criar diagramas mais sofisticados a partir do texto e até mesmo sugerir otimizações de layout. O objetivo é continuar construindo um kit de ferramentas pessoal que lide com mais tarefas que não são de escrita, me liberando para focar no conteúdo em si.

## Recursos

*   **[Projeto Speedgrapher](https://github.com/danicat/speedgrapher):** O código-fonte do servidor MCP discutido neste artigo.
*   **[Como Construir um Servidor MCP com Gemini CLI e Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}}):** O artigo anterior que inspirou esta jornada.
*   **[Model Context Protocol (MCP)](https://modelcontextprotocol.io/):** O site oficial do protocolo.
*   **[Anúncio do Gemini CLI](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/):** O post do blog que anunciou o suporte a comandos de barra personalizados.
*   **[Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index):** Saiba mais sobre a métrica de legibilidade.
