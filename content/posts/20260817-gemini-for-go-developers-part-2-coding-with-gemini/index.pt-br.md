---categories:
- Agentic Coding
date: 2026-08-17
heroStyle: big
series:
- Gemini for Go Developers
series_order: 2
summary: Na Parte 2 de Gemini para Desenvolvedores Go, exploramos a afinidade agentiva do Go,
  as superfícies do Antigravity e como estruturar um fluxo prático de desenvolvimento nativo de IA em Go.
tags:
  - antigravity
  - gemini
  - golang
  - mcp
title: "Gemini para Desenvolvedores Go: Programando com o Gemini"
slug: "gemini-for-go-developers-part-2-coding-with-gemini"
aliases:
  - "/pt-br/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini/"
description: "Parte 2 de Gemini para Go: maximize a afinidade agentiva do Go, explore o Antigravity, configure servidores MCP e monte seu kit de IA para Go."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "Antigravity CLI"
  - "godoctor"
---

Boas-vindas de volta à série **Gemini para Desenvolvedores Go**! Na [Parte 1: A Família de Modelos Gemini]({{< ref "/posts/20260808-gemini-for-go-developers-part-1-model-family" >}}), conhecemos os diferentes modelos Gemini para casos de uso específicos, examinamos as superfícies de API disponíveis e escrevemos nossas primeiras linhas de código em Go com o [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) oficial.

Nesta Parte 2, vamos explorar como utilizar o Gemini para acelerar o desenvolvimento diário em Go. Começaremos com uma reflexão sobre escolhas de linguagem na Era da IA, depois analisaremos o ecossistema de *harnesses* e padrões para agentes, finalizando com a configuração recomendada para maximizar a afinidade agentiva do Go no seu ambiente de trabalho.

## Por que usar Go na era da IA?

Antes de avançarmos, vale responder à dúvida fundamental: na era da inteligência artificial, a escolha da linguagem de programação ainda faz diferença?

No passado, a seleção de uma linguagem dependia quase sempre da bagagem técnica prévia da equipe. Dominar novas sintaxes, bibliotecas e particularidades de compilação exigia semanas de dedicação; por isso, times tendiam a permanecer na sua zona de conforto a menos que fossem impelidos por uma grande virada tecnológica.

A IA transformou essa dinâmica. A sintaxe deixou de ser um gargalo intransponível quando modelos geram código e *boilerplate* sob demanda, e aprender uma nova stack com um tutor de IA é infinitamente mais produtivo do que passar horas torcendo para que uma discussão de cinco anos atrás no Stack Overflow resolva seu erro de compilação.

Por que, então, se importar com a escolha da linguagem? Tudo se resume a dois pilares: o **ecossistema da linguagem** e sua **afinidade agentiva**.

Nesse contexto, o ecossistema de uma linguagem refere-se a tudo o que ela atrai por gravidade natural: SDKs ativamente mantidos, maturidade das bibliotecas, qualidade da documentação e vitalidade da comunidade técnica. Já a **afinidade agentiva** é um conceito próprio dos dias de hoje. Eu a defino como "a facilidade com que conseguimos orientar um agente a produzir código confiável nessa linguagem". Uma alta afinidade agentiva depende, em primeiro lugar, de quão preparados os modelos estão para gerar soluções corretas na linguagem e, em segundo, de quão aptos estão para utilizar o ferramental necessário para compilar, testar e manter esse código.

A afinidade agentiva decorre naturalmente do volume e da qualidade de dados aos quais o modelo teve acesso durante o treinamento, beneficiando linguagens com farto código público e documentação de excelência. Mas volume por si só não basta: linguagens antigas que passaram por rupturas bruscas de paradigma sofrem com práticas desiguais, fazendo com que os modelos sugiram abordagens obsoletas no momento da inferência.

Na minha experiência prática, linguagens como Go, Python e JavaScript apresentam alta afinidade agentiva por padrão. O Go se destaca com louvor: sua clareza sintática, tipagem estática rigorosa e a filosofia quase pythônica de que "o explícito é melhor que o implícito" tornam a geração de código muito menos sujeita a alucinações de sintaxe. Mais importante ainda: a compilação instantânea do Go, seus testes integrados e seu conjunto padronizado e opinativo de ferramentas oferecem aos agentes de código um ciclo de feedback rápido e determinístico para diagnosticar e corrigir falhas de forma autônoma. (Para um mergulho mais detalhado sobre como a filosofia do Go se alinha ao desenvolvimento de software assistido por IA, confira [este ensaio de Cameron Balahan e Richard Seroter no Google Developers Blog](https://developers.googleblog.com/why-go-is-an-ideal-language-for-ai-assisted-software-engineering/)).

Por outro lado, linguagens como R e C ocupam o extremo oposto do espectro. O R é uma linguagem acadêmica de nicho em que, mesmo com suporte total de agentes, não consigo manter meu pacote [read.dbc](https://github.com/danicat/read.dbc) no CRAN (o repositório oficial de pacotes do R), porque nem eu nem os modelos conseguimos reproduzir os erros apontados pelo pipeline do CRAN. Em C, a escassez de salvaguardas nativas faz com que pequenos descuidos gerem comportamentos indefinidos e falhas silenciosas que os modelos têm enorme dificuldade de identificar sem um ambiente ergonômico e assistido.

Em última análise, a afinidade agentiva é medida pela fluidez do **loop agentivo**: a capacidade do agente em compilar, rodar testes, executar benchmarks e corrigir o código por conta própria, sem a necessidade de microgerenciamento humano constante.

## Selecionando a superfície agentiva correta

Embora a disputa entre modelos de fronteira atraia todos os holofotes, o modelo em si é apenas parte da equação no desenvolvimento com IA. A qualidade das entregas e a experiência de uso dependem criticamente do ambiente de execução (*harness*) que orquestra esse modelo.

Para quem desenvolve com o Gemini, o caminho natural é o **ecossistema Antigravity**. Com o lançamento do Antigravity 2.0, o Google organizou esse ecossistema em três superfícies complementares:

### Antigravity 2.0

No centro do ecossistema está o aplicativo desktop **Antigravity 2.0**, também chamado de **Agent Manager**. Sua maior inovação é colocar a interação com o agente em primeiro plano, deixando o código nos bastidores. Para quem está acostumada com IDEs tradicionais, a mudança causa surpresa: não há árvore de diretórios nem janelas de edição manual de arquivos. Todas as alterações são realizadas pelo agente. Seu papel é orientar as metas e revisar o trabalho por meio de anotações e comentários, em um estilo muito parecido com o do Google Docs.

### Antigravity CLI (`agy`)

Embora a proposta do Agent Manager seja recente, interfaces de terminal para agentes já possuem uma história consolidada, impulsionadas por projetos como Claude, Aider e Cline. Em junho de 2025, o Google lançou o Gemini CLI, que mais tarde foi descontinuado em favor da nova **Antigravity CLI** (`agy`).

Um detalhe que me deixa particularmente feliz como Gopher é que o `agy` foi construído em Go (diferente do Gemini CLI, que era em TypeScript), proporcionando uma experiência de terminal visivelmente mais rápida e responsiva.

### Antigravity IDE

Para quem ainda prefere a ergonomia de um editor de código tradicional, o Google disponibiliza a **Antigravity IDE** como aplicação independente. Baseada no VS Code, ela mantém toda a disposição de painéis e atalhos consagrados, integrando um painel lateral nativo para colaboração contínua com o Gemini.

Para ser 100% transparente: hoje em dia, quase nunca abro a IDE para programar. As únicas ocasiões em que recorro a ela são para redigir ou revisar artigos técnicos (como este), já que ainda preservo boa parte da minha escrita de forma manual. Para código propriamente dito, raramente faço edições manuais hoje em dia.

## Padrões agentivos aplicados à programação

Independentemente da interface escolhida, o Antigravity é extremamente potente de fábrica, mas atinge seu potencial máximo quando equipamos o agente com **customizações**.

O Antigravity suporta tanto convenções consolidadas quanto padrões emergentes da indústria: regras (*rules*), skills, MCP, hooks, subagentes, sidecars e plugins. No entanto, o suporte a essas extensões ainda **não é uniforme** entre as diferentes interfaces da plataforma.

Vamos conferir o papel de cada uma dessas customizações na prática.

### Instruções e regras para agentes

O conceito de instruções padronizadas ganhou força com a iniciativa [**AGENTS.md**](https://agents.md/), suportada no Antigravity por meio de arquivos `AGENTS.md` ou `GEMINI.md` (além de regras modulares em `.gemini/rules/` ou `.agents/rules/`).

Podemos pensar no `AGENTS.md` como o `README.md` dos agentes de IA. É o local ideal para registrar convenções arquiteturais, preferências de estilo, comandos de teste e restrições de projeto indispensáveis para o agente, sem poluir a documentação voltada para a equipe humana.

Para ser completamente sincera, com o advento das *Agent Skills*, raramente atualizo o `GEMINI.md` hoje em dia, e muitos dos meus repositórios provavelmente guardam versões defasadas (falha minha, coitados dos agentes!). Dito isso, instruções ainda são ótimas para alinhar expectativas. Lembre-se apenas de que diretrizes em texto funcionam como recomendações e não como barreiras intransponíveis: durante sessões longas, os modelos podem ignorar regras ou se desviar delas.

### Model Context Protocol (MCP)

O [**Model Context Protocol**](https://modelcontextprotocol.io/) (MCP) é um protocolo aberto criado para conectar modelos de IA a ferramentas e fontes de dados externas de maneira padronizada.

A especificação define três primitivas principais:
1. **Tools:** Funções executáveis que o agente pode chamar (como rodar um linter, consultar um banco de dados ou verificar a compilação).
2. **Resources:** Fontes de dados somente leitura acessíveis para consulta (documentações, esquemas de tabelas ou arquivos de log).
3. **Prompts:** Modelos pré-estruturados de fluxos de trabalho.

No dia a dia de engenharia, usamos essencialmente as *tools*, enquanto a maioria dos clientes sequer implementa *resources* ou *prompts*. Recursos podem ser facilmente expostos via ferramentas dedicadas de consulta, e os prompts caíram em desuso com a flexibilidade das Agent Skills.

Como as skills podem incluir scripts executáveis, há quem venha abandonando o MCP em favor exclusivamente de skills. Ainda vejo vantagens claras no MCP: uma de suas maiores forças é a gestão centralizada do ciclo de vida, principalmente via HTTPS. Uma empresa pode hospedar um servidor MCP corporativo na nuvem e permitir que todos os clientes acessem documentações e ferramentas sempre atualizadas com uma única configuração. Já a garantia de versionamento uniforme de skills em larga escala ainda é um desafio em aberto.

### Agent skills

Uma [skill](https://agentskills.io) é uma pasta estruturada contendo diretrizes em Markdown (`SKILL.md`), documentações complementares e scripts auxiliares que ensinam ao agente como executar um fluxo de trabalho técnico específico.

O pilar arquitetural das skills é a **revelação progressiva** (*progressive disclosure*). Em vez de injetar centenas de páginas de manual na janela de contexto logo no início da conversa, o ambiente apresenta apenas o nome e o resumo descritivo de cada skill. Quando o agente identifica que a demanda corresponde àquela habilidade, ele carrega o conteúdo integral do `SKILL.md` e executa os scripts associados sob demanda. Esse padrão mantém o conhecimento especializado sempre acessível, sem esgotar a memória de contexto nem dispersar a atenção do modelo com detalhes irrelevantes.

Os agentes utilizam a descrição das skills para identificar seus gatilhos de ativação, mas depender apenas do reconhecimento automático pode ser arriscado. Por padrão, o Antigravity disponibiliza as skills como comandos com barra: basta digitar `/<nome-da-skill>` no prompt para forçar a execução imediata. Ativar proativamente as skills certas evita retrabalho em etapas críticas do projeto.

### Hooks

Enquanto regras, prompts e skills fornecem orientação consultiva, os **hooks** garantem controle determinístico sobre o loop do agente. Hooks são rotinas de interceptação chamadas em etapas exatas do ciclo de vida: antes da execução de uma ferramenta (`PreToolUse`), após a conclusão de uma ferramenta (`PostToolUse`), antes de invocar o modelo (`PreInvocation`) ou no encerramento da sessão (`Stop`).

Dada a natureza não determinística dos LLMs, pedir via prompt para "sempre rodar o linter após modificar o código" deixa a conferência sujeita a falhas. Já os hooks são gerenciados diretamente pelo *harness*, executando sem exceção para o evento correspondente.

Vale apenas um alerta de segurança: modelos modernos são muito habilidosos em tentar contornar restrições. Se um hook bloqueia uma ação, o agente pode tentar desativar o script, mascarar o padrão de chamada ou alterar sua própria configuração.

Com a nova geração de modelos cada vez mais sofisticada, reduzi o uso de hooks rígidos e venho priorizando a orientação via skills. É como educar: instruir e guiar costuma ser mais sustentável do que apenas proibir.

### Subagentes

Os **subagentes** representam outra solução elegante para a gestão da janela de contexto, viabilizando estratégias produtivas como o processamento paralelo de tarefas. Ao instanciar subagentes, o agente principal divide o escopo do problema e delega cada parte a um especialista dedicado.

Pense no desenvolvimento de uma aplicação web com frontend e backend. As frentes são essencialmente ortogonais: o frontend envolve HTML, CSS e TypeScript, enquanto o backend utiliza Go, PostgreSQL e Docker. Ambas possuem regras de estilo e esteiras de teste distintas. Salvo pelo contrato de integração da API, as duas áreas pouco compartilham; tentar resolver ambas na mesma janela de contexto gera ruído e saturação desnecessária.

Ao delegar as frentes a subagentes distintos, você assegura foco total em cada camada, eliminando contaminações de contexto causadas pelo cruzamento de tecnologias não correlatas.

Outra excelente aplicação é a revisão independente de código: peça ao agente para rodar a checagem em um subagente e aproveite a vantagem de uma avaliação totalmente limpa sobre o que acabou de ser produzido.

### Sidecars

> **Nota**: No momento em que escrevo, sidecars são suportados exclusivamente no Antigravity 2.0, não estando disponíveis no CLI `agy` nem na IDE.

**Sidecars** são processos que rodam em segundo plano de forma contínua durante a sessão do agente. Podem ser serviços persistentes cujo ciclo de vida é gerenciado pelo Antigravity com políticas de reinicialização automática, ou utilitários acionados por agendamento prévio.

Ainda não explorei os sidecars a fundo, mas meu colega Mete Atamel vem compartilhando excelentes conteúdos sobre o tema — recomendo conferir [o artigo dele](https://medium.com/google-cloud/where-does-antigravity-look-for-sidecars-20e7002b9246) para saber mais.

### Plugins de agentes

Plugins funcionam como pacotes de distribuição integrados, reunindo regras, MCPs, skills, hooks, agentes e sidecars em um único módulo instalável. Venho testando plugins há algum tempo e sua maturidade ainda varia bastante entre as diferentes plataformas.

Para que se mostram mais úteis hoje? Para o empacotamento conjunto de servidores MCP e skills. É nessa direção que o mercado vem convergindo com a especificação aberta [Agent Plugins](https://agent-plugins.org/specification), deixando configurações específicas de cada cliente de fora por enquanto.

## Montando o seu kit de ferramentas agentivas em Go

Agora que revisamos as superfícies e os padrões de extensão, vamos estruturar nosso ferramental prático de desenvolvimento em Go. Podemos separá-lo entre ferramentas essenciais da comunidade e extensões nativas de IA.

### Ferramentas essenciais da comunidade

O ecossistema oficial do Go já oferece uma base muito sólida, mas alguns utilitários elevam a qualidade do código e a segurança das entregas:

- [**`golangci-lint`**](https://golangci-lint.run/): Agrega dezenas de linters de alta performance em uma única execução, capturando erros não tratados, asserções de tipo inseguras e problemas de concorrência que escapam ao `go vet`.
- [**`goreleaser`**](https://goreleaser.com/): Para projetos que compilam binários, o `goreleaser` automatiza a geração de artefatos multiplataforma, o gerenciamento de pipelines de release e a criação de changelogs a partir de um arquivo `.goreleaser.yaml`.
- [**`modernize`**](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize) / **`go fix`**: Analisa a compatibilidade do código com versões recentes do Go e atualiza padrões antigos de forma automatizada (como substituir iterações manuais em maps/slices por funções nativas modernas).
- [**`deadcode`**](https://pkg.go.dev/golang.org/x/tools/cmd/deadcode): Realiza análise estática de alcance no programa inteiro para identificar funções órfãs e trechos de código inalcançáveis em todos os pacotes.
- **`selene` e `testquery` (um jabá descarado):** Mantenho dois utilitários de código aberto voltados para suítes de teste. O [**`selene`**](https://github.com/danicat/selene) é uma ferramenta de teste de mutação para Go que introduz falhas controladas na AST para verificar se os testes realmente detectam defeitos. Já o [**`testquery`**](https://github.com/danicat/testquery) é uma CLI que disponibiliza uma interface SQL para consultar resultados de cobertura e métricas por teste. Embora sejam ferramentas de nicho, me ajudam muito a lapidar suítes de teste com os agentes.

### Ferramental especializado para IA

As ferramentas acima foram concebidas originalmente para desenvolvedores humanos, mas os agentes de código conseguem executá-las com facilidade via linha de comando. Vejamos agora os servidores MCP e as skills dedicados a fluxos nativos de IA.

#### O servidor MCP oficial do `gopls`

Nas IDEs convencionais, o `gopls` fornece recursos semânticos essenciais: verificação de tipos, navegação em definições e busca de referências. Mas quando um agente opera sem interface gráfica, ele costuma enxergar o código apenas como texto puro.

Para resolver isso, a equipe do Go incorporou suporte nativo ao protocolo MCP no [**`gopls`**](https://pkg.go.dev/golang.org/x/tools/gopls). Executar o `gopls` em modo MCP expõe o motor de análise e o índice do compilador diretamente ao modelo como ferramentas executáveis. Isso permite ao agente inspecionar hierarquias de pacotes e validar assinaturas de funções com base no modelo semântico real da linguagem.

#### O servidor MCP `godoctor`

Uma das minhas ressalvas em relação ao MCP do `gopls` é que ele foi concebido sobre a API do LSP — que foi desenhada para a digitação interativa de pessoas, enquanto as interações de agentes são transacionais. Por essa razão, desenvolvi o [**`godoctor`**](https://github.com/danicat/godoctor) para oferecer utilitários nativos de IA pensados especificamente para Go.

A versão atual do `godoctor` inclui:
- **`smart_edit`:** Editor de código ciente da AST com validação contínua via `go vet` e autocorreção de nomes. Se uma edição causar falha de compilação ou sintaxe, o `godoctor` desfaz a alteração e sugere correções com base nos identificadores mais próximos, orientando o agente com dicas no estilo "você quis dizer...?".
- **`smart_build`:** Esteira automatizada de verificação que executa a limpeza dos módulos (`go mod tidy`, `modernize`, `goimports`), compila o pacote, roda a suíte de testes com cobertura e aplica as regras do `golangci-lint` em uma única passada.
- **`smart_test`:** Pipeline de testes opinativo com suporte integrado ao `testquery` e ao `selene`.
- **`read_docs`:** Consulta à documentação com base no `go doc`, com suporte a exemplos práticos e mecanismo de fallback independente da configuração dos módulos locais.

#### Conhecimento de plataforma e autoaperfeiçoamento


Enquanto `godoctor` e `gopls` cobrem a semântica do código local, os agentes de programação também precisam de acesso em tempo real à documentação dos serviços com os quais se integram. Aqui estão os principais recursos para quem trabalha com Go, GCP e Gemini:

- **Google Developer Knowledge MCP (`developerknowledge.googleapis.com/mcp`):** Conecta o agente diretamente à documentação oficial do Google Cloud, Gemini Enterprise (Vertex AI) e APIs do Google.
- **Gemini Docs MCP (`gemini-api-docs-mcp.dev`):** Disponibiliza a documentação atualizada dos endpoints da Gemini API, novas versões do SDK e padrões recomendados de configuração (saiba mais no [guia de agentes de código do Gemini](https://ai.google.dev/gemini-api/docs/coding-agents)).
- **Skills oficiais do Google:** Os repositórios [**`github.com/google/skills`**](https://github.com/google/skills) e [**`github.com/google-gemini/gemini-skills`**](https://github.com/google-gemini/gemini-skills) reúnem skills oficiais do time do Google (incluindo `gemini-api-dev`, `gemini-live-api-dev` e `gemini-interactions-api`).
- **Catálogo comunitário e pessoal:** Você pode conferir minhas skills pessoais em [**`skills.danicat.dev`**](https://skills.danicat.dev) (ou no [**GitHub**](https://github.com/danicat/skills)), com coleções para boas práticas de engenharia, criação de jogos 2D e geração multimodal (Lyria, Nano Banana Pro).

Para enriquecer seu próprio fluxo de trabalho com extensões autorais:

- **AgentSkills MCP (`agentskills.io/mcp`):** Mecanismo de busca e consulta para a [especificação aberta do Agent Skills](https://agentskills.io) e guias de boas práticas de autoria — excelente para quando você criar suas próprias skills de trabalho (o que você definitivamente deveria fazer!).
- **MCP Dev skills:** Se existe um MCP para desenvolvimento de agent skills, por que não ter também [skills de agente para desenvolvimento de MCPs](https://modelcontextprotocol.io/docs/2026-07-28/develop/build-with-agent-skills)? Sim, você leu certo (rs). Embora seja um tema especializado, criar MCPs sob medida é uma excelente maneira de potencializar seu ambiente de desenvolvimento.

## A configuração agentiva de 5 minutos para Gophers

Se você procura uma configuração pronta para começar a programar em Go com o Gemini hoje mesmo, aqui está o guia rápido de 5 minutos:

1. Baixe o [Antigravity](https://antigravity.google) no site oficial.
2. Configure os servidores MCP recomendados:
   - [Gemini Docs MCP](https://ai.google.dev/gemini-api/docs/coding-agents): `npx add-mcp "https://gemini-api-docs-mcp.dev"`
   - [Developer Knowledge MCP](https://developers.google.com/knowledge/mcp): ative a API no Google Cloud e configure-a conforme as instruções da documentação.
   - [Agent Skills MCP](https://agentskills.io): copie os parâmetros de configuração disponíveis na página do projeto.
   - [godoctor](https://github.com/danicat/godoctor): execute o script de instalação de linha única.
3. Adicione as skills essenciais:
   - [Desenvolvimento com Gemini API](https://ai.google.dev/gemini-api/docs/coding-agents): `npx skills add google-gemini/gemini-skills --skill gemini-api-dev`
   - [Swarm coding]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}): `npx skills add github.com/danicat/skills/agents/swarm-coding`
4. Teste a esteira completa:
   Oriente seu agente a rodar uma validação autônoma:
   > Run a smart build on this package with godoctor, address any findings, and evaluate the test suite with selene.

## O que vem a seguir?

Neste capítulo, cobrimos o panorama de harnesses para agentes, os principais padrões de customização (regras, MCP, skills, hooks, subagentes e plugins) e como estruturar um ambiente prático e eficiente para programar em Go com o Gemini.

Na **Parte 3: Desenvolvendo Agentes em Go**, passaremos para o outro lado da mesa: como construir runtimes de agentes autônomos utilizando a própria linguagem Go. Exploraremos loops de execução de ferramentas (*tool calling*), engenharia de contexto e frameworks de alto nível como o **Genkit Go** e o **Agent Development Kit (ADK)**. Nos vemos lá!
