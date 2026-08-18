---
categories:
- Agentic Coding
date: 2026-08-17
heroStyle: big
series:
- Gemini for Go Developers
series_order: 2
summary: Na Parte 2 de Gemini para Desenvolvedores Go, exploramos a afinidade agentic do Go,
  superfícies do Antigravity e fluxos nativos de IA em Go.
tags:
  - antigravity
  - gemini
  - golang
  - mcp
title: 'Gemini para Desenvolvedores Go - Parte 2: Programando com o Gemini'
---

Boas-vindas de volta à série **Gemini para Desenvolvedores Go**! Na [Parte 1: A Família de Modelos Gemini]({{< ref "/posts/20260808-gemini-for-go-developers-part-1-model-family" >}}), exploramos os diferentes modelos Gemini para casos de uso específicos, examinamos as superfícies de API para consumir modelos e escrevemos nosso primeiro código em Go com o [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai) oficial.

Agora, na Parte 2, vamos explorar como usar o Gemini para programar em Go. Começaremos com uma breve discussão sobre escolhas de linguagem na Era da IA, depois exploraremos o panorama de harnesses e padrões para agentes, finalizando com a configuração recomendada para aumentar a afinidade agentic do Go no seu ambiente de desenvolvimento.

## Por que usar Go na era da IA?

Antes de avançarmos, vamos tirar a dúvida óbvia do caminho: na era da IA, a escolha da linguagem de programação para nossas aplicações ainda importa?

No passado, a seleção de uma linguagem era quase sempre ditada pela experiência existente na equipe. Aprender novas sintaxes, padrões e peculiaridades de ferramentas exigia um tempo valioso; por isso, as equipes permaneciam em suas zonas de conforto, a menos que fossem forçadas a migrar por uma grande mudança tecnológica.

A IA transformou essa dinâmica. A sintaxe não é mais uma barreira quando os modelos geram boilerplate sob demanda, e aprender uma stack desconhecida com um tutor de IA é ordens de grandeza mais rápido do que passar horas rezando para que alguma thread de cinco anos atrás no Stack Overflow corresponda exatamente ao seu erro de compilação.

Por que você se importaria com a seleção de linguagem então? Tudo se resume a dois temas principais: **ecossistema da linguagem** e **afinidade agentic**.

Neste contexto, o ecossistema de uma linguagem refere-se a tudo o que a linguagem atrai por "gravidade": SDKs mantidos ativamente, bibliotecas, documentação, vitalidade da comunidade e conhecimento do setor. Mas a **afinidade agentic** é algo exclusivo do mundo em que vivemos hoje. Eu a defino como "o quão fácil é orientar um agente a programar com essa linguagem". Uma alta afinidade agentic é influenciada principalmente pelo preparo dos modelos para gerar código nessa linguagem e, em segundo lugar, por quão preparados eles estão para adotar as ferramentas necessárias para verificar, testar e manter esse código.

A afinidade agentic depende naturalmente do volume de dados sobre uma determinada linguagem disponível durante o treinamento do modelo, beneficiando linguagens populares que possuem ampla documentação e código público. Dito isso, volume por si só não é tudo: linguagens mais antigas que passaram por grandes mudanças de paradigma sofrem com práticas fragmentadas, fazendo com que os modelos sugiram padrões obsoletos no momento da inferência.

Na minha experiência pessoal, linguagens como Go, Python e JavaScript têm alta afinidade agentic por padrão. O Go se destaca em particular: sua legibilidade, tipagem estática rigorosa e a filosofia quase "Pythonesque" de que "explícito é melhor do que implícito" tornam a geração de código muito menos propensa a alucinações de sintaxe. Mais importante ainda, a compilação rápida do Go, seus testes nativos e suas ferramentas padronizadas e opinativas oferecem aos agentes de programação um loop de feedback imediato e determinístico para detectar e corrigir erros autonomamente. (Para uma análise mais aprofundada sobre como a filosofia de design do Go se alinha com a engenharia de software assistida por IA, confira [este ensaio de Cameron Balahan e Richard Seroter no Google Developers Blog](https://developers.googleblog.com/why-go-is-an-ideal-language-for-ai-assisted-software-engineering/)).

Por outro lado, linguagens como R e C estão no outro extremo do espectro. No caso do R, é uma linguagem acadêmica de nicho em que, mesmo com suporte total de agentes, não consigo manter meu pacote [read.dbc](https://github.com/danicat/read.dbc) no CRAN (o sistema de distribuição de pacotes do R) porque nem eu nem os modelos conseguimos reproduzir o problema relatado pelo pipeline do CRAN. Em C, a falta de salvaguardas faz com que erros simples se transformem em bugs silenciosos e catastróficos que os modelos têm dificuldade de detectar precocemente sem ferramentas ergonômicas.

Em última análise, a afinidade agentic é medida pela velocidade e responsividade do "loop" agentic: se o seu agente consegue compilar, testar, rodar benchmarks e reparar o código prontamente, sem necessidade de microgerenciamento humano constante.

## Selecionando a superfície agentic correta

Embora a corrida dos modelos de fronteira seja empolgante de acompanhar, os modelos em si são apenas uma parte da equação quando se trata de geração de código. A qualidade das respostas e a experiência do desenvolvedor ao programar com um agente são fortemente influenciadas pelo harness que está executando o modelo.

Se você está programando com o Gemini, a escolha natural é o **ecossistema Antigravity**. Com o lançamento do Antigravity 2.0, o Google separou o ecossistema em três superfícies distintas, dependendo de como você prefere trabalhar:

### Antigravity 2.0

No centro do ecossistema Antigravity está a aplicação desktop **Antigravity 2.0**, às vezes também chamada de **Agent Manager**. A maior mudança aqui é colocar a experiência do agente em primeiro plano enquanto o código fica em segundo plano. Para usuários de primeira viagem acostumados com uma IDE, a experiência pode ser um pouco chocante: não há árvores de arquivos para explorar a base de código, nem qualquer forma de editar arquivos manualmente. Cada interação é feita por meio do agente. Seu controle reside em orientar o que o agente deve fazer e anotar o trabalho dele com comentários no estilo "Google Docs".

### O Antigravity CLI (`agy`)

Embora o conceito de Agent Manager seja relativamente novo, UIs de terminal para agentes já existem há mais tempo, popularizadas por Claude, Aider, Cline e outros. Em junho de 2025, o Google também lançou sua própria UI de terminal, o Gemini CLI, mas ela foi descontinuada em favor do novo **Antigravity CLI** (`agy`).

Um detalhe que me deixa particularmente feliz como Gopher é que o `agy` foi escrito em Go (enquanto o Gemini CLI era em TypeScript), resultando em uma experiência de terminal visivelmente mais rápida e ágil.

### O Antigravity IDE

Se você ainda prefere um editor de código visual e dedicado, o Google oferece o **Antigravity IDE** como uma aplicação companheira separada. Ele é baseado no VS Code, portanto todos os elementos familiares da IDE estão onde você lembra que estavam, além de contar com um painel lateral de agente para interações com o Gemini.

Para ser 100% transparente, hoje em dia raramente abro a IDE para programar. As únicas vezes em que realmente uso a IDE são quando estou escrevendo ou revisando artigos (como este), já que ainda faço uma grande parte do meu processo de escrita manualmente. Para código, quase nunca edito nada manualmente hoje em dia.

## Padrões agentic aplicados à programação

Independentemente da superfície escolhida, o Antigravity é muito capaz por padrão, mas tem suas peculiaridades. A melhor forma de aproveitar ao máximo os recursos do Antigravity é equipar seu agente com **customizações**.

O Antigravity suporta tanto padrões agentic bem estabelecidos quanto emergentes: regras, skills, MCP, hooks, subagentes, sidecars e plugins. No entanto, o suporte a essas customizações, infelizmente, **não é uniforme** entre as diferentes superfícies do Antigravity hoje.

Vamos ver como cada uma dessas customizações funciona na prática.

### Instruções e regras para agentes

O conceito de instruções foi padronizado pela iniciativa [**AGENTS.md**](https://agents.md/), suportada pelo Antigravity por meio de arquivos `AGENTS.md` ou `GEMINI.md` (assim como regras modulares em `.gemini/rules/` ou `.agents/rules/`).

Você pode pensar no `AGENTS.md` como o `README.md` para agentes de IA. É o lugar para armazenar o contexto do projeto, restrições arquiteturais, comandos de teste e preferências de estilo que são essenciais para um agente saber, mas que, de outra forma, sobrecarregariam a documentação humana.

Para ser completamente sincero, com o surgimento das Agent Skills, raramente atualizo o `GEMINI.md`, e a maioria dos meus repositórios provavelmente tem arquivos desatualizados (erro meu, pobres agentes!). Dito isso, eles ainda são úteis para direcionar os agentes no caminho certo. Apenas lembre-se de que instruções baseadas em prompts funcionam mais como recomendações do que guardrails: um agente ainda pode ignorar seletivamente ou desviar das regras durante uma sessão longa.

### Model Context Protocol (MCP)

O [**Model Context Protocol**](https://modelcontextprotocol.io/) (MCP) é um padrão aberto para conectar aplicações de IA a ferramentas e fontes de dados externas.

O protocolo expõe três primitivas centrais:
1. **Tools:** Funções executáveis que o agente pode invocar (por exemplo, consultar um banco de dados, rodar um linter, verificar um build).
2. **Resources:** Fontes de dados somente leitura que o agente pode inspecionar (por exemplo, arquivos de documentação, esquemas de banco de dados, logs do sistema).
3. **Prompts:** Modelos de fluxo de trabalho predefinidos.

No mundo real, nos importamos com tools mais do que com qualquer outra coisa, e a maioria dos clientes sequer suportará resources ou prompts. Resources podem ser emulados via tools (uma ferramenta dedicada pode recuperar dados) e prompts caíram em grande parte no esquecimento devido à introdução das Agent Skills, que são muito mais flexíveis.

Como as skills podem ser empacotadas com scripts, algumas pessoas estão até abandonando MCPs completamente em favor de skills. Ainda acho que há muitos casos em que MCPs são melhores que skills. Um dos pontos fortes dos MCPs sobre skills é o gerenciamento de ciclo de vida, especialmente ao usá-los sobre HTTPS. Por exemplo, como empresa, você pode implantar um servidor MCP como um serviço web e seus clientes só precisam configurá-lo uma vez para ter acesso imediato à documentação atualizada. Com skills, por outro lado, garantir que todos os seus clientes usem apenas as versões mais recentes ainda é um problema a ser resolvido em escala.

### Agent skills

Uma [skill](https://agentskills.io) é um diretório contendo instruções (`SKILL.md`), scripts auxiliares opcionais e documentação que ensinam um agente a executar um fluxo de trabalho de engenharia específico.

O conceito arquitetural definidor por trás das skills é a **revelação progressiva (progressive disclosure)**. Em vez de despejar centenas de páginas de documentação na janela de contexto do agente antecipadamente, o sistema injeta apenas o nome e a descrição da skill. Quando o agente determina que uma tarefa corresponde a uma skill, ele carrega as instruções completas do `SKILL.md` e executa os scripts empacotados sob demanda. Esse modelo garante que o conhecimento especializado esteja sempre disponível, sem sobrecarregar a janela de contexto ou distrair seu agente com dados irrelevantes para a tarefa em questão.

Os agentes usam descrições de skills para identificar seus gatilhos de ativação, mas confiar apenas na ativação automática é arriscado. Por padrão, o Antigravity mapeia todas as skills como comandos slash, permitindo que você force a ativação de uma skill digitando `/<nome-da-skill>` em qualquer lugar do seu prompt. Ser proativo com a ativação de skills economizará muitas dores de cabeça se a skill for importante para o seu fluxo de trabalho.

### Hooks

Enquanto regras, prompts e até skills oferecem orientação suave, os **hooks** introduzem controle determinístico no loop do agente. Hooks são callbacks que interceptam o ciclo de vida do agente em momentos específicos, como antes da execução de uma ferramenta (`PreToolUse`), após a execução de uma ferramenta (`PostToolUse`), antes de uma invocação do modelo (`PreInvocation`) ou ao encerrar a sessão (`Stop`).

Como os LLMs são não-determinísticos, dizer a um agente para "sempre rodar um linter após editar código" via prompts deixa a validação à mercê do acaso. Hooks, por outro lado, são controlados pelo harness e sempre são executados para o evento correspondente.

Há apenas uma ressalva à qual você sempre precisa prestar atenção ao projetar seus hooks: os modelos são muito bons em contorná-los. Sim, infelizmente isso acontece. Um hook pode impedir o agente de causar danos, apenas para o agente tentar ser mais esperto que o hook imediatamente depois, seja alterando a configuração do agente, tentando mascarar a condição de disparo ou, pior, reescrevendo o próprio script do hook.

Entristece-me dizer isto: no passado, usei muito os hooks, mas com a nova geração de modelos eles ficaram espertos demais, então estou migrando lentamente de hooks para skills. É como educar uma criança: não proíba, eduque.

### Subagentes

Os **subagentes** oferecem outra solução para o problema da janela de contexto, ao mesmo tempo em que possibilitam paradigmas interessantes, como a execução paralela. Ao gerar subagentes, o agente principal pode segmentar o espaço do problema e criar um agente focado em cada tarefa.

Um exemplo trivial seria trabalhar em um serviço web que possui tanto um frontend quanto um backend. As alterações são essencialmente ortogonais: o frontend requer HTML, CSS e JavaScript, enquanto o backend requer Go, Python e talvez algum SQL. As tarefas de frontend e backend terão padrões de código e pipelines de build diferentes. Com exceção do contrato entre eles, eles não têm nada em comum e, se feitos na mesma janela de contexto, uma parte só gerará ruído para a outra.

Ao dividir as tarefas entre dois (ou mais) subagentes, você garante que cada subagente tenha foco total em sua camada da stack, reduzindo o risco de degradação de contexto devido ao cruzamento de tecnologias completamente diferentes.

Outro bom exemplo é quando você precisa de uma revisão imparcial do trabalho recém-concluído. Peça ao agente para executar a revisão de código em um subagente e você terá o benefício de "olhos novos" avaliando o código.

### Sidecars

> Nota: no momento em que escrevo, os sidecars funcionam apenas no Antigravity 2.0. Eles não estão disponíveis no CLI `agy` ou na IDE.

**Sidecars** são processos em segundo plano que rodam junto com o agente durante a sessão. Eles podem ser processos persistentes com o ciclo de vida gerenciado pelo Antigravity por meio de uma política de reinicialização definida, ou processos que rodam de acordo com um agendamento.

Ainda não explorei os sidecars a fundo, então não há muito que eu possa acrescentar à discussão no momento, mas meu colega Mete Atamel começou recentemente a escrever sobre eles, e encorajo você a conferir o [artigo dele](https://medium.com/google-cloud/where-does-antigravity-look-for-sidecars-20e7002b9246) para mais informações.

### Plugins de agentes

Plugins são, em essência, pacotes de customização, reunindo regras, MCPs, skills, hooks, agentes e sidecars em uma única unidade de distribuição. Tenho experimentado plugins há algum tempo e, infelizmente, eles não funcionam tão suavemente quanto deveriam. Por exemplo, tive muita dificuldade tentando garantir que os hooks empacotados com meus plugins fossem executados corretamente, mas isso nunca funcionou para mim.

Além disso, o suporte a plugins não é padronizado entre as diferentes superfícies: por exemplo, sidecars só são suportados em plugins do Antigravity 2.0, agentes personalizados só são suportados no Antigravity CLI e assim por diante.

Para que eles servem então? MCPs e skills. Seja coincidência ou não, é também para onde a indústria está convergindo com o novo padrão [Agent Plugins](https://agent-plugins.org/specification). A justificativa deles é que hooks, agentes, regras, servidores LSP e outros ainda são muito específicos de cada cliente, ficando fora de escopo por enquanto.

## Montando o seu kit de ferramentas agentic em Go

Agora que cobrimos harnesses e customizações, vamos equipar nosso kit de desenvolvimento em Go. Podemos dividir essas ferramentas entre utilitários essenciais da comunidade e extensões nativas para IA.

### Ferramentas essenciais da comunidade

Embora o conjunto de ferramentas padrão do Go ofereça uma excelente base para os agentes, diversas ferramentas da comunidade elevam a qualidade do código e a automação de releases:

- [**`golangci-lint`**](https://golangci-lint.run/): Reúne dezenas de linters rápidos em uma única passada, capturando erros não tratados, asserções de tipo não verificadas, variáveis sombreadas e armadilhas de concorrência que o `go vet` não detecta.
- [**`goreleaser`**](https://goreleaser.com/): Para projetos que distribuem binários, o `goreleaser` automates a geração de artefatos multiplataforma, o gerenciamento de pipelines de release e a criação de changelogs a partir de um `.goreleaser.yaml`.
- [**`modernize`**](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize) / **`go fix`**: Analisa o código em relação às versões mais novas do Go e atualiza mecanicamente códigos legados, como substituir loops manuais de slice/map ou funções auxiliares de min/max por funções integradas modernas.
- [**`deadcode`**](https://pkg.go.dev/golang.org/x/tools/cmd/deadcode): Utiliza análise de alcance em todo o programa (reachability analysis) para identificar funções não utilizadas e código inacessível em todos os pacotes.
- **`selene` e `testquery` (jabá descarado):** Mantenho duas ferramentas de código aberto para inspecionar e melhorar suítes de teste. O [**`selene`**](https://github.com/danicat/selene) é uma ferramenta de teste de mutação para Go que introduz falhas direcionadas na AST para verificar se os testes realmente capturam defeitos no código. O [**`testquery`**](https://github.com/danicat/testquery) é uma CLI que expõe uma interface SQL para consultar resultados de testes e cobertura por teste em Go. Embora um pouco de nicho, elas me ajudam a otimizar minhas suítes de teste.

### Ferramentas específicas para IA

As ferramentas acima foram construídas para desenvolvedores humanos, mas agentes de programação podem executá-las diretamente por meio de comandos de terminal padrão. Agora vamos dar uma olhada em MCPs e skills especializados para fluxos de trabalho nativos de IA.

#### O servidor MCP oficial do `gopls`

Em uma IDE tradicional, o `gopls` fornece ao editor consciência semântica: verificação de tipos, referências e definições de símbolos. Mas quando um agente opera em um ambiente headless ou no terminal, ele normalmente interage com o código como arquivos de texto brutos.

Para preencher essa lacuna, a equipe do Go adicionou suporte nativo a MCP no [**`gopls`**](https://pkg.go.dev/golang.org/x/tools/gopls). Executar o `gopls` no modo MCP expõe o verificador de tipos e o índice do language server diretamente ao modelo como ferramentas executáveis. Isso permite que o agente navegue pelas hierarquias de pacotes e inspecione assinaturas de tipos usando o próprio modelo semântico do compilador.

#### O servidor MCP do `godoctor`

Uma das minhas ressalvas sobre o servidor MCP do `gopls` é que ele foi projetado sobre a API de LSP, que não foi concebida com agentes em mente: LSPs foram construídos para velocidade de digitação e digitação interativa, enquanto fluxos de trabalho de agentes são transacionais. Por essa razão, construí o [**`godoctor`**](https://github.com/danicat/godoctor) para fornecer ferramentas nativas de IA para o desenvolvimento em Go.

A versão atual do `godoctor` oferece o seguinte:
- **`smart_edit`:** Um editor de arquivos ciente da AST com validação automática via `go vet` e correção de erros de digitação. Se uma edição introduzir um erro de compilação ou sintaxe, o `godoctor` reverte automaticamente a alteração e sugere correções com base em identificadores próximos, como um prompt de orientação no estilo "você quis dizer...?".
- **`smart_build`:** Um pipeline de verificação automatizado que executa higiene de módulos (`go mod tidy`, `modernize`, `goimports`), compila o pacote, roda testes com cobertura e valida o linting via `golangci-lint` em uma única passada.
- **`smart_test`:** Um pipeline de testes opinativo com suporte integrado ao `testquery` e ao `selene`.
- **`read_docs`:** Consulta de documentação baseada no `go doc`, com suporte a exemplos e um sistema de fallback que exibe a documentação independentemente da configuração de módulos.

#### Conhecimento da plataforma e autoaperfeiçoamento

Enquanto o `godoctor` e o `gopls` cuidam da semântica do código local, os agentes de programação também precisam de conhecimento da plataforma em tempo real e guias de fluxo de trabalho para entender os serviços com os quais estão se integrando. Aqui estão os recursos essenciais de nuvem e API para trabalhar com Go, GCP e Gemini:

- **Google Developer Knowledge MCP (`developerknowledge.googleapis.com/mcp`):** Conecta o agente diretamente à documentação oficial do Google Cloud, Gemini Enterprise (Vertex AI) e APIs do Google.
- **Gemini Docs MCP (`gemini-api-docs-mcp.dev`):** Fornece documentação atualizada para endpoints da Gemini API, atualizações de SDK e padrões de configuração (leia mais no [guia de agentes de código do Gemini](https://ai.google.dev/gemini-api/docs/coding-agents)).
- **Skills oficiais do Google:** [**`github.com/google/skills`**](https://github.com/google/skills) e [**`github.com/google-gemini/gemini-skills`**](https://github.com/google-gemini/gemini-skills) contêm skills oficiais mantidas pelo Google (incluindo `gemini-api-dev`, `gemini-live-api-dev` e `gemini-interactions-api`).
- **Catálogo comunitário e pessoal:** Você pode explorar minhas skills pessoais em [**`skills.danicat.dev`**](https://skills.danicat.dev) (ou no [**GitHub**](https://github.com/danicat/skills)), que incluem skills para boas práticas de engenharia, desenvolvimento de jogos 2D, mídia generativa (Lyria, Nano Banana Pro) e muito mais.

E para otimizar seu próprio fluxo de trabalho com extensões customizadas:

- **AgentSkills MCP (`agentskills.io/mcp`):** O mecanismo oficial de busca e recuperação para consultar a [especificação aberta do Agent Skills](https://agentskills.io) e as melhores práticas de autoria. Excelente para quando você cria suas próprias skills no trabalho diário — o que você definitivamente deveria fazer.
- **Skills de desenvolvimento de MCP:** Se existe um MCP para desenvolvimento de agent skills, por que não ter também [agent skills para desenvolvimento de MCP](https://modelcontextprotocol.io/docs/2026-07-28/develop/build-with-agent-skills)? Sim, você leu certo (LOL). Embora seja um pouco de nicho, como você pode ver pelo meu próprio trabalho, criar MCPs para uso próprio também é uma excelente forma de melhorar seu ambiente de desenvolvimento.

## A configuração agentic de 5 minutos do Gopher

Se você quer uma configuração opinativa para começar a programar com o Gemini hoje, aqui está o guia rápido de 5 minutos:

1. Baixe o [Antigravity](https://antigravity.google) no site oficial.
2. Configure os servidores MCP recomendados:
   - [Gemini Docs MCP](https://ai.google.dev/gemini-api/docs/coding-agents): `npx add-mcp "https://gemini-api-docs-mcp.dev"`
   - [Developer Knowledge MCP](https://developers.google.com/knowledge/mcp): ative a API e configure-a seguindo as instruções na página de documentação.
   - [Agent Skills MCP](https://agentskills.io): Expanda o botão de copiar em qualquer página para ver as instruções.
   - [godoctor](https://github.com/danicat/godoctor): Use o script de instalação em uma linha.
3. Adicione as skills recomendadas:
   - [Desenvolvimento com Gemini API](https://ai.google.dev/gemini-api/docs/coding-agents): `npx skills add google-gemini/gemini-skills --skill gemini-api-dev`
   - [Swarm coding]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}): `npx skills add github.com/danicat/skills/agents/swarm-coding`
4. Teste o loop:
   Peça ao seu agente para rodar uma verificação autônoma:
   > Run a smart build on this package with godoctor, address any findings, and evaluate the test suite with selene.

## O que vem a seguir?

Neste capítulo, cobrimos o panorama de harnesses para agentes, padrões de customização (regras, MCP, skills, hooks, subagentes e plugins) e como configurar um ambiente prático para programar em Go com o Gemini.

Na **Parte 3: Desenvolvendo Agentes em Go**, vamos passar para o outro lado da mesa: construindo runtimes autônomos de agentes em Go. Exploraremos loops de chamada de ferramentas (tool calling), engenharia de contexto e frameworks de alto nível como o **Genkit Go** e o **Agent Development Kit (ADK)**. Nos vemos lá!
