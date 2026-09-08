---
categories:
- Agentic Coding
date: 2026-08-17
heroStyle: big
series:
- Gemini for Go Developers
series_order: 2
summary: Na Parte 2 de Gemini para Desenvolvedores Go, exploramos a afinidade agentiva do Go, as superfícies do Antigravity e como configurar um fluxo prático de desenvolvimento nativo de IA em Go.
tags:
  - antigravity
  - gemini
  - golang
  - mcp
title: "Gemini para Desenvolvedores Go: Programando com o Gemini"
slug: "gemini-for-go-developers-part-2-coding-with-gemini"
aliases:
  - "/pt-br/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini/"
description: "Parte 2 de Gemini para Go: maximize a afinidade agentiva do Go, explore as superfícies do Antigravity, configure servidores MCP e monte seu kit de IA para Go."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "Antigravity CLI"
  - "godoctor"
---

Boas-vindas de volta à série **Gemini para Desenvolvedores Go**! Na [Parte 1: A Família de Modelos Gemini]({{< ref "/posts/20260808-gemini-for-go-developers-part-1-model-family" >}}), exploramos os diferentes modelos Gemini para casos de uso específicos, analisamos as superfícies de API para consumir os modelos e escrevemos nosso primeiro código em Go com o [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) oficial.

Agora, na Parte 2, vamos explorar como usar o Gemini para programar em Go. Começaremos com uma breve discussão sobre escolhas de linguagem na Era da IA, depois exploraremos o ecossistema de harnesses e padrões para agentes, finalizando com a configuração recomendada para aumentar a afinidade agentiva do Go no seu ambiente de desenvolvimento.

## Por que usar Go na era da IA?

Antes de irmos mais a fundo, vamos tirar da frente a pergunta óbvia: na era da IA, a nossa escolha de linguagem de programação ainda importa?

No passado, a seleção de uma linguagem era quase sempre ditada pela experiência prévia da equipe. Aprender uma nova sintaxe, novos idiomas e as peculiaridades do ferramental tomava um tempo valioso; por isso, os times permaneciam na sua zona de conforto, a menos que fossem forçados a mudar por uma grande transição tecnológica.

A IA virou essa dinâmica de ponta-cabeça. A sintaxe não é mais uma barreira quando modelos geram boilerplate sob demanda, e aprender uma stack desconhecida com um tutor de IA supera de longe passar horas rezando para que alguma thread de cinco anos atrás no Stack Overflow corresponda exatamente ao seu erro de compilação.

Por que você se importaria com a escolha da linguagem, então? Tudo se resume a dois temas principais: o **ecossistema da linguagem** e a **afinidade agentiva**.

Nesse contexto, o ecossistema da linguagem refere-se a tudo o que a linguagem atrai por "gravidade": SDKs ativamente mantidos, bibliotecas, documentação, a vitalidade da comunidade e o conhecimento acumulado do setor. Mas a **afinidade agentiva** é exclusiva do mundo em que vivemos hoje. Eu a defino como "o quão fácil é guiar um agente para programar nessa linguagem". Uma alta afinidade agentiva é influenciada, em primeiro lugar, por quão preparados os modelos estão para gerar código nessa linguagem e, em segundo lugar, por quão preparados estão para adotar o ferramental necessário para verificar, testar e manter esse código.

A afinidade agentiva depende naturalmente de quanto dado sobre uma determinada linguagem estava disponível durante o treinamento do modelo, beneficiando linguagens populares com código público e literatura extensos. Dito isso, volume sozinho não é tudo: linguagens mais antigas que passaram por mudanças massivas de paradigma sofrem com práticas fragmentadas, fazendo com que os modelos sugiram padrões obsoletos no momento da inferência.

Na minha experiência pessoal, linguagens como Go, Python e JavaScript têm alta afinidade agentiva por padrão. O Go se destaca em particular: sua legibilidade, tipagem estática rigorosa e a filosofia quase pythônica de que "o explícito é melhor do que o implícito" tornam a geração de código muito menos propensa a alucinações de sintaxe. Mais importante ainda: a compilação rápida do Go, seus testes nativos integrados e o ferramental padrão opinativo oferecem aos agentes de código um loop de feedback imediato e determinístico para capturar e corrigir erros de forma autônoma. (Para um mergulho mais profundo em como a filosofia de design do Go se alinha à engenharia de software assistida por IA, confira [este ensaio de Cameron Balahan e Richard Seroter no Google Developers Blog](https://developers.googleblog.com/why-go-is-an-ideal-language-for-ai-assisted-software-engineering/)).

Por outro lado, linguagens como R e C ficam no outro extremo do espectro. No caso do R, é uma linguagem acadêmica de nicho em que, mesmo com suporte total de agentes, não consigo manter meu pacote [read.dbc](https://github.com/danicat/read.dbc) no CRAN (o sistema de distribuição de pacotes do R) porque nem eu nem os modelos conseguimos reproduzir o problema reportado pelo pipeline do CRAN. Em C, a falta de salvaguardas significa que erros simples podem se transformar em bugs catastróficos e silenciosos que os modelos têm dificuldade para capturar cedo sem uma rica ergonomia de ferramentas.

Em última análise, a afinidade agentiva é medida pela velocidade e capacidade de resposta do "loop" agentivo: se o seu agente consegue compilar, testar, rodar benchmarks e corrigir o código sem necessidade de microgerenciamento humano constante.

## Selecionando a superfície agentiva correta

Embora a corrida dos modelos de fronteira seja empolgante de acompanhar, os modelos em si são apenas parte da equação quando se trata de geração de código. A qualidade das respostas e a experiência do desenvolvedor ao programar com um agente são fortemente influenciadas pelo harness que está executando o modelo.

Se você está programando com o Gemini, a escolha natural é o **ecossistema Antigravity**. Com o lançamento do Antigravity 2.0, o Google separou o ecossistema em três superfícies distintas, dependendo de como você prefere trabalhar:

### Antigravity 2.0

No centro do ecossistema Antigravity está o aplicativo desktop **Antigravity 2.0**, às vezes também chamado de **Agent Manager**. A maior mudança aqui é colocar a experiência do agente no centro do palco, enquanto o código fica nos bastidores. Para usuários de primeira viagem acostumados a uma IDE, a experiência pode ser um pouco chocante: não há árvores de arquivos para explorar a base de código, nem qualquer forma de editar arquivos manualmente. Cada interação é feita através do agente. O seu controle reside em dizer ao agente o que fazer e anotar o trabalho dele com comentários no estilo "Google Docs".

### O Antigravity CLI (`agy`)

Embora o conceito de Agent Manager seja bastante novo, interfaces de terminal para agentes já existem há mais tempo, popularizadas por Claude, Aider, Cline e outros. Em junho de 2025, o Google também lançou sua própria interface de terminal, o Gemini CLI, mas desde então ele foi descontinuado em favor do novo **Antigravity CLI** (`agy`).

Um detalhe que me deixa particularmente feliz como Gopher é que o `agy` foi escrito em Go (enquanto o Gemini CLI era em TypeScript), resultando em uma experiência de terminal visivelmente mais rápida e ágil.

### O Antigravity IDE

Se você ainda prefere um editor de código visual e dedicado, o Google oferece o **Antigravity IDE** como uma aplicação companheira separada. Ele é baseado no VS Code, então todos os elementos familiares da IDE continuam onde você se lembra, além de contar com um painel lateral de agente para interações com o Gemini.

Para ser 100% transparente, hoje em dia eu raramente abro a IDE para programar. As únicas vezes em que realmente uso a IDE são quando estou escrevendo ou revisando artigos (como este), já que ainda faço uma parte enorme do meu processo de escrita manualmente. Para código, muito raramente edito qualquer coisa à mão atualmente.

## Padrões agentivos aplicados à programação

Independentemente da superfície escolhida, o Antigravity é muito capaz assim que sai da caixa, mas tem suas peculiaridades. A melhor forma de aproveitar ao máximo as capacidades do Antigravity é equipar o seu agente com **customizações**.

O Antigravity suporta padrões agentivos tanto consolidados quanto emergentes: regras (*rules*), skills, MCP, hooks, subagentes, sidecars e plugins. No entanto, o suporte a essas customizações, infelizmente, **não é uniforme** entre as diferentes superfícies do Antigravity hoje.

Vamos ver como cada uma dessas customizações funciona na prática.

### Instruções e regras para agentes

O conceito de instruções foi padronizado pela iniciativa [**AGENTS.md**](https://agents.md/), que o Antigravity suporta por meio de arquivos `AGENTS.md` ou `GEMINI.md` (assim como regras modulares sob `.gemini/rules/` ou `.agents/rules/`).

Você pode pensar no `AGENTS.md` como o `README.md` para agentes de IA. É o lugar para guardar contexto do projeto, restrições arquiteturais, comandos de teste e preferências de estilo que são essenciais para o agente saber, mas que poluiriam a documentação voltada para humanos.

Para ser completamente sincera, com o advento das Agent Skills, raramente atualizo o `GEMINI.md` hoje em dia, e a maioria dos meus repositórios provavelmente tem arquivos desatualizados (minha culpa, coitados dos agentes!). Dito isso, eles ainda são úteis para direcionar os agentes para o caminho certo. Só tenha em mente que instruções baseadas em prompts atuam mais como recomendações do que como guardrails: um agente ainda pode ignorar seletivamente ou se afastar das regras durante uma sessão longa.

### Model Context Protocol (MCP)

O [**Model Context Protocol**](https://modelcontextprotocol.io/) (MCP) é um padrão aberto para conectar aplicações de IA a ferramentas e fontes de dados externas.

O protocolo expõe três primitivas centrais:
1. **Tools:** Funções executáveis que o agente pode invocar (por exemplo, consultar um banco de dados, rodar um linter, verificar um build).
2. **Resources:** Fontes de dados somente leitura que o agente pode inspecionar (por exemplo, arquivos de documentação, esquemas de banco de dados, logs do sistema).
3. **Prompts:** Modelos pré-definidos de fluxo de trabalho.

No mundo real, nós nos importamos mais com as tools do que com qualquer outra coisa, e a maioria dos clientes nem sequer oferece suporte a resources ou prompts. Resources podem ser emulados por meio de tools (uma tool dedicada pode recuperar dados) e prompts caíram em grande parte no esquecimento devido à introdução das Agent Skills, que são muito mais flexíveis.

Como as skills podem ser empacotadas com scripts, algumas pessoas estão até mesmo abandonando o MCP completamente em favor de skills. Eu ainda acho que há muitos casos em que o MCP é melhor do que skills. Um dos pontos fortes do MCP em relação às skills é o gerenciamento de ciclo de vida, especialmente ao utilizá-lo sobre HTTPS. Por exemplo, como empresa, você pode implantar um servidor MCP como um serviço web e seus clientes só precisam configurá-lo uma vez para ter acesso imediato a documentação atualizada. Com skills, por outro lado, garantir que todos os seus clientes usem apenas as versões mais recentes ainda é um problema a ser resolvido em escala.

### Agent skills

Uma [skill](https://agentskills.io) é um diretório contendo instruções (`SKILL.md`), scripts auxiliares opcionais e documentação que ensinam ao agente como executar um fluxo de trabalho de engenharia específico.

O conceito arquitetural definidor por trás das skills é a **revelação progressiva** (*progressive disclosure*; explore padrões de produção em [O Guia Prático de Agent Skills]({{< ref "/posts/20260829-the-pragmatic-guide-to-agent-skills" >}})). Em vez de despejar centenas de páginas de documentação na janela de contexto do agente logo de início, o sistema apenas injeta o nome e a descrição da skill. Quando o agente determina que uma tarefa corresponde a uma skill, ele carrega as instruções completas do `SKILL.md` e executa os scripts empacotados sob demanda. Esse modelo garante que o conhecimento especializado esteja sempre disponível, sem sobrecarregar a janela de contexto nem distrair seu agente com dados irrelevantes para a tarefa em mãos.

Os agentes usam as descrições das skills para identificar seus gatilhos de ativação, mas depender unicamente da ativação automática é arriscado. Por padrão, o Antigravity mapeia todas as skills como comandos com barra (*slash commands*), para que você possa forçar a ativação de uma skill digitando `/<nome-da-skill>` em qualquer lugar do seu prompt. Ser proativa com a ativação de skills vai lhe poupar muitas dores de cabeça se a skill for importante para o seu fluxo de trabalho.

### Hooks

Enquanto regras, prompts e até mesmo skills oferecem uma orientação branda, os [**hooks**]({{< ref "/posts/20260610-mastering-hooks" >}}) introduzem controle determinístico no loop do agente. Hooks são callbacks que interceptam o ciclo de vida do agente em momentos específicos, como antes da execução de uma ferramenta (`PreToolUse`), após a execução de uma ferramenta (`PostToolUse`), antes de uma invocação do modelo (`PreInvocation`) ou no encerramento da sessão (`Stop`).

Como os LLMs são não determinísticos, dizer a um agente para "sempre rodar um linter após editar o código" via prompts deixa a validação à mercê da sorte. Hooks, por outro lado, são controlados pelo harness e sempre são executados para o seu evento designado.

Há apenas um alerta que você precisa sempre prestar atenção ao projetar seus hooks: os modelos são muito bons em contorná-los. Sim, infelizmente, isso acontece. Um hook pode impedir o agente de fazer algo prejudicial, apenas para o agente tentar ser mais esperto que o hook logo em seguida, seja alterando a configuração do agente, tentando mascarar a condição de disparo ou, pior, reescrevendo o script do hook.

Me entristece dizer isso: no passado, eu usava bastante os hooks, mas com a nova geração de modelos eles ficaram espertos demais para o próprio bem, então estou aos poucos migrando dos hooks para as skills. Parece educar uma criança: não proíba, eduque.

### Subagentes

Os [**subagentes**]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}) oferecem outra solução para o problema da janela de contexto, ao mesmo tempo em que possibilitam paradigmas interessantes, como a execução paralela. Ao criar subagentes, o agente principal pode segmentar o espaço do problema e criar um agente focado em cada tarefa.

Um exemplo trivial seria trabalhar em um serviço web que possui tanto frontend quanto backend. As mudanças são essencialmente ortogonais entre si: o frontend requer HTML, CSS e JavaScript, enquanto o backend requer Go, Python e talvez algum SQL. As tarefas de frontend e backend terão padrões de código e pipelines de build diferentes. Com exceção do contrato entre eles, não há nada em comum e, se feitas na mesma janela de contexto, uma parte será apenas ruído para a outra.

Ao dividir as tarefas entre dois (ou mais) subagentes, você garante que cada subagente tenha foco total na sua fatia da stack, reduzindo o risco de degradação do contexto causada por cruzar os fluxos de duas stacks de tecnologia completamente diferentes.

Outro bom exemplo é quando você precisa de uma revisão imparcial do trabalho que acabou de concluir. Peça ao agente para rodar a revisão de código em um subagente e você terá o benefício de "olhos frescos" analisando o código.

### Sidecars

> Nota: no momento em que escrevo, sidecars funcionam apenas no Antigravity 2.0. Eles não estão disponíveis no CLI `agy` nem na IDE.

**Sidecars** são processos em segundo plano que rodam junto com o agente durante a sessão. Eles podem ser processos persistentes com ciclo de vida gerenciado pelo Antigravity por meio de uma política de reinicialização definida, ou processos que rodam de acordo com um agendamento.

Eu ainda não explorei os sidecars a fundo, então não há muito que eu possa acrescentar à discussão no momento, mas meu colega Mete Atamel acabou de começar a escrever sobre eles, e encorajo você a conferir [o artigo dele](https://medium.com/google-cloud/where-does-antigravity-look-for-sidecars-20e7002b9246) para saber mais.

### Plugins de agentes

Plugins são, em essência, pacotes de customizações, reunindo regras, MCPs, skills, hooks, agentes e sidecars em uma única unidade de distribuição. Venho experimentando plugins há algum tempo e, infelizmente, eles não funcionam tão suavemente quanto deveriam. Por exemplo, tive bastante trabalho tentando garantir que os hooks empacotados com meus plugins fossem executados corretamente, mas isso nunca funcionou para mim.

Além disso, o suporte a plugins não é padronizado entre as diferentes superfícies: por exemplo, sidecars são suportados apenas em plugins do Antigravity 2.0, agentes customizados são suportados apenas no Antigravity CLI, e assim por diante.

Para que eles servem então? MCPs e skills. Seja por coincidência ou não, é também aí que a indústria está convergindo com o novo padrão [Agent Plugins](https://agent-plugins.org/specification). A justificativa deles é que hooks, agentes, regras, servidores LSP e outros ainda são bastante específicos de cada cliente e, portanto, fora do escopo por enquanto.

## Montando o seu kit de ferramentas agentivas em Go

Agora que cobrimos os harnesses de agentes e as customizações, vamos montar nosso kit de ferramentas de desenvolvimento em Go. Podemos dividi-las entre ferramentas essenciais da comunidade e extensões especializadas nativas de IA.

### Ferramentas essenciais da comunidade

Embora a toolchain padrão do Go forneça aos agentes uma linha de base sólida, várias ferramentas da comunidade elevam a qualidade do código e a automação de releases:

- [**`golangci-lint`**](https://golangci-lint.run/): Reúne dezenas de linters rápidos em uma única passada, capturando erros não tratados, asserções de tipo não verificadas, variáveis sombreadas e armadilhas de concorrência que o `go vet` não detecta.
- [**`goreleaser`**](https://goreleaser.com/): Para projetos que distribuem binários, o `goreleaser` automatiza a compilação de artefatos multiplataforma, o gerenciamento de pipelines de release e a geração de changelogs a partir do `.goreleaser.yaml`.
- [**`modernize`**](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize) / **`go fix`**: Analisa o código em relação a versões mais recentes do Go e atualiza mecanicamente o boilerplate antigo, como substituir loops manuais em slices/maps ou funções auxiliares min/max por funções nativas modernas.
- [**`deadcode`**](https://pkg.go.dev/golang.org/x/tools/cmd/deadcode): Usa análise de alcance de programa inteiro para identificar funções não utilizadas e código inalcançável entre os pacotes.
- **`selene` e `testquery` (jabá descarado):** Eu mantenho duas ferramentas de código aberto para inspecionar e aprimorar suítes de testes. O [**`selene`**](https://github.com/danicat/selene) é uma ferramenta de testes de mutação para Go que introduz falhas direcionadas na AST para verificar se os testes realmente detectam defeitos no código. O [**`testquery`**](https://github.com/danicat/testquery) é uma CLI que expõe uma interface SQL para consultar resultados de testes e cobertura por teste em Go. Embora sejam um tanto de nicho, elas me ajudam a otimizar minhas suítes de testes.

### Ferramental específico para IA

As ferramentas acima foram construídas para desenvolvedores humanos, mas agentes de código podem executá-las diretamente por meio de comandos de shell padrão. Agora, vamos ver os MCPs e skills especializados para fluxos de trabalho nativos de IA.

#### O servidor MCP oficial do `gopls`

Em uma IDE tradicional, o `gopls` fornece ao editor percepção semântica: verificação de tipos, referências e definições de símbolos. Mas quando um agente opera em um ambiente headless ou de terminal, ele normalmente interage com o código como arquivos de texto puro.

Para fechar essa lacuna, a equipe do Go adicionou suporte nativo a MCP ao [**`gopls`**](https://pkg.go.dev/golang.org/x/tools/gopls). Executar o `gopls` no modo MCP expõe o verificador de tipos e o índice do language server diretamente ao modelo como ferramentas executáveis. Isso permite que o agente navegue pelas hierarquias de pacotes e inspecione assinaturas de tipos usando o próprio modelo semântico do compilador.

#### O servidor MCP `godoctor`

Uma das minhas queixas sobre o servidor MCP do `gopls` é que ele foi projetado sobre a API do LSP, que não foi pensada com agentes em mente: LSPs foram construídos para velocidade de digitação e toques interativos no teclado, enquanto os fluxos de trabalho de agentes são transacionais. Por essa razão, construí o [**`godoctor`**](https://github.com/danicat/godoctor) para fornecer ferramentas nativas de IA para desenvolvimento em Go.

A versão atual do `godoctor` oferece o seguinte:
- **`smart_edit`:** Um editor de arquivos ciente da AST com validação automática via `go vet` e correção de digitação. Se uma edição introduzir um erro de compilação ou sintaxe, o `godoctor` reverte automaticamente a alteração e sugere correções com base em identificadores próximos, com um prompt de orientação no estilo "você quis dizer...?".
- **`smart_build`:** Um pipeline automatizado de verificação que executa a higiene do módulo (`go mod tidy`, `modernize`, `goimports`), compila o pacote, executa testes com cobertura e valida o linting via `golangci-lint` em uma única passada.
- **`smart_test`:** Um pipeline de testes opinativo com suporte para `testquery` e `selene`.
- **`read_docs`:** Consulta de documentação baseada em `go doc`, com suporte a exemplos e um sistema de fallback que exibe a documentação independentemente da configuração do módulo.

#### Conhecimento de plataforma e autoaperfeiçoamento

Enquanto `godoctor` e `gopls` lidam com a semântica do código local, agentes de programação também precisam de conhecimento em tempo real da plataforma e playbooks de fluxo de trabalho para entender os serviços com os quais estão se integrando. Aqui estão os recursos essenciais de nuvem e API para trabalhar com Go, GCP e Gemini:

- **Google Developer Knowledge MCP (`developerknowledge.googleapis.com/mcp`):** Conecta o agente diretamente à documentação oficial do Google Cloud, Gemini Enterprise (Vertex AI) e APIs do Google.
- **Gemini Docs MCP (`gemini-api-docs-mcp.dev`):** Fornece documentação atualizada para os endpoints da Gemini API, atualizações do SDK e padrões de configuração (leia mais no [guia de agentes de código do Gemini](https://ai.google.dev/gemini-api/docs/coding-agents)).
- **Skills oficiais do Google:** [**`github.com/google/skills`**](https://github.com/google/skills) e [**`github.com/google-gemini/gemini-skills`**](https://github.com/google-gemini/gemini-skills) contêm skills oficiais mantidas pelo Google (incluindo `gemini-api-dev`, `gemini-live-api-dev` e `gemini-interactions-api`).
- **Catálogo comunitário e pessoal:** Você pode explorar minhas skills pessoais em [**`skills.danicat.dev`**](https://skills.danicat.dev) (ou no [**GitHub**](https://github.com/danicat/skills)), que incluem skills para boas práticas de engenharia, desenvolvimento de jogos 2D, mídia generativa (Lyria, Nano Banana Pro) e outras.

E para otimizar seu próprio fluxo de trabalho com extensões personalizadas:

- **AgentSkills MCP (`agentskills.io/mcp`):** O mecanismo oficial de busca e recuperação para consultar a [especificação aberta do Agent Skills](https://agentskills.io) e boas práticas de criação. Excelente para quando você criar suas próprias skills como parte do seu trabalho diário. O que você deveria fazer.
- **Skills para desenvolvimento de MCP:** Se existe um MCP para desenvolvimento de agent skills, por que não ter também [skills de agentes para desenvolvimento de MCP](https://modelcontextprotocol.io/docs/2026-07-28/develop/build-with-agent-skills)? Sim, você leu certo (rs). Embora seja um pouco nichado, como você pode ver pelo meu próprio trabalho, criar MCPs para seu próprio uso também é uma ótima maneira de aprimorar seu ambiente de desenvolvimento.

## O setup agentivo de 5 minutos do Gopher

Se você deseja uma configuração opinativa para começar a trabalhar com o Gemini hoje mesmo, aqui está o guia rápido de 5 minutos:

1. Baixe o [Antigravity](https://antigravity.google) no site oficial.
2. Configure os servidores MCP recomendados:
   - [Gemini Docs MCP](https://ai.google.dev/gemini-api/docs/coding-agents): `npx add-mcp "https://gemini-api-docs-mcp.dev"`
   - [Developer Knowledge MCP](https://developers.google.com/knowledge/mcp): ative a API e configure-a seguindo as instruções na página de documentação.
   - [Agent Skills MCP](https://agentskills.io): expanda o botão de cópia em qualquer página para ver as instruções.
   - [godoctor](https://github.com/danicat/godoctor): use o script de instalação de uma linha só.
3. Adicione as skills recomendadas:
   - [Desenvolvimento com Gemini API](https://ai.google.dev/gemini-api/docs/coding-agents): `npx skills add google-gemini/gemini-skills --skill gemini-api-dev`
   - [Swarm coding]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}): `npx skills add github.com/danicat/skills/agents/swarm-coding`
4. Faça um test drive no loop:
   Oriente seu agente a executar uma passada de verificação autônoma:
   > Run a smart build on this package with godoctor, address any findings, and evaluate the test suite with selene.

## O que vem a seguir?

Neste capítulo, cobrimos o panorama de harnesses de agentes, padrões de customização (regras, MCP, skills, hooks, subagentes e plugins) e como configurar um ambiente prático para programar em Go com o Gemini.

Na **Parte 3: Desenvolvendo Agentes em Go**, vamos atravessar para o outro lado da mesa: construir runtimes de agentes autônomos em Go. Vamos explorar loops de chamada de ferramentas (*tool calling*), engenharia de contexto e frameworks de agentes de nível superior como o **Genkit Go** e o **Agent Development Kit (ADK)**. Até lá!
