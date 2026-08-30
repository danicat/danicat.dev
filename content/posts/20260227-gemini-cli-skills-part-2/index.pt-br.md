---
categories:
- Agentic Coding
date: 2026-02-27 12:00:00+00:00
heroStyle: big
series:
- Agent Skills
series_order: 2
summary: Aprenda a usar o skill-creator nativo da Gemini CLI para gerar, refinar
  e estruturar automaticamente suas próprias Agent Skills personalizadas usando exemplos
  práticos.
tags:
  - agent-skills
  - gemini-cli
  - vibe-coding
title: "Criando Agent Skills com o skill-creator"
slug: "gemini-cli-skills-part-2"
aliases:
  - "/pt-br/posts/20260227-gemini-cli-skills-part-2/"
description: "Guia de design de Agent Skills com o skill-creator na Gemini CLI. Exemplos práticos: latest-version, pyhd com Ruff, find-examples e de-sloppify."
proficiencyLevel: "Intermediate"
dependencies:
  - "Gemini CLI >= 0.1.0"
  - "Node.js >= 18"
  - "Python 3.10+"
---

{{< alert "circle-info" >}}
**Atualização (2026):** A Gemini CLI evoluiu para o **Google Antigravity 2.0**. Embora os conceitos fundamentais e a estrutura de Agent Skills apresentados neste artigo continuem válidos, confira [O Guia do Mochileiro para o Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}) para uma visão geral da plataforma atualizada e de seus novos recursos.
{{< /alert >}}

Na [Parte 1: Dominando Agent Skills na Gemini CLI]({{< ref "/posts/20260128-agent-skills-gemini-cli" >}}), exploramos como as Agent Skills adicionam novos recursos à Gemini CLI. Analisamos a skill `experiment-analyst` como um exemplo prático de como manter o contexto do agente limpo, fornecendo instruções específicas para uma tarefa.

Hoje, vamos mergulhar nos princípios fundamentais do design de skills e analisar alguns exemplos práticos de skills que utilizo diariamente no meu fluxo de trabalho.

## O que são Agent Skills

Vamos fazer uma rápida recapitulação da parte 1, caso você tenha perdido. Agent Skills são um padrão aberto projetado para dar aos agentes de codificação conhecimento especializado "just in time". Elas são desenhadas para que o conhecimento especializado seja adicionado ao contexto apenas quando necessário, ajudando a evitar o chamado *context bloat* (inchaço de contexto). O termo técnico para isso é **Progressive Disclosure** (Divulgação Progressiva): mantemos as instruções principais (`SKILL.md`) o mais enxutas possível e usamos arquivos separados para referências detalhadas ou scripts que são carregados apenas *quando necessário*.

No disco, uma skill é uma pasta com um arquivo `SKILL.md` e, opcionalmente, recursos empacotados:

```text
skill-name/
├── SKILL.md (Obrigatório: apenas nome, descrição e instruções principais)
└── Bundled Resources (Opcional)
    ├── scripts/    (Código executável para tarefas repetíveis)
    ├── references/ (Documentação carregada sob demanda, ex: esquemas de API)
    └── assets/     (Templates ou arquivos binários usados nas saídas)
```

## A skill `skill-creator`

Você sempre pode escrever skills manualmente, mas a Gemini CLI já vem com uma meta-skill integrada chamada `skill-creator`, que facilita bastante as coisas.

Você pode ativar essa skill pedindo à Gemini CLI para criar (ou refatorar) uma skill:
> *"Quero criar uma nova skill para buscar a versão real mais recente de um pacote de software para que você pare de alucinar versões."*

Qualquer solicitação relacionada a "criar skills" deve acionar automaticamente o `skill-creator`, mas, caso você esteja lidando com um modelo mais "rabugento", também pode ser mais explícita:

> *"Use o skill-creator para escrever uma skill para des-sloppificar textos gerados por IA (por favor, não leve para o lado pessoal)"*

A Gemini CLI pode pedir alguns detalhes antes de gerar o boilerplate da skill. Ela aprendeu recentemente a interagir com a ferramenta [`ask_user`](https://geminicli.com/docs/tools/ask-user/#ask_user-ask-user), e é muito legal ver isso em ação.

## Quando criar skills

No meu fluxo de trabalho pessoal, tenho dois usos principais para skills:

1. Para documentar um processo específico do meu trabalho (ex.: como fazer code review do jeito que eu gosto, como inicializar um repositório, como avaliar um post de blog, etc.)
2. Para adicionar conhecimento especializado sobre uma ferramenta, linguagem ou tecnologia específica (ex.: como funciona um projeto Genkit, como trabalhar com ADK para desenvolver agentes, etc.)

Até certo ponto, você pode pensar em skills como um conceito intermediário entre comandos de barra (que costumo armazenar como prompts MCP) e ferramentas. Ao criar comandos de barra, meu objetivo é ter um "processo repetível"; ao criar ferramentas, quero dar ao modelo uma forma determinística de executar algo. Como as skills podem conter tanto prompts quanto scripts, elas conseguem atender a ambos os cenários, com os scripts desempenhando o papel de ferramentas.

Claro, se você estiver empacotando suas skills como parte de uma extensão, há uma grande chance de que elas sejam distribuídas com um servidor MCP que também exponha ferramentas. Você também pode aproveitar essa integração na definição da skill, ensinando o modelo a usar suas ferramentas MCP quando estiverem disponíveis.

Também costumo criar skills em dois momentos principais:

1. Após uma sessão dolorosamente longa tentando ensinar o modelo a fazer algo para mim (ex.: "por favor, consolide o que acabamos de fazer em uma skill que possamos reutilizar depois")
2. Logo após ter uma nova ideia que acredito que me tornará mais eficiente no trabalho (ex.: "vamos escrever uma skill de de-slopify para melhorar sua escrita")

Em ambos os casos, a skill dificilmente estará perfeita de primeira, mas assim que começo a usá-la, vou refinando até ter certeza de que agrega valor real — ou descarto e guardo a ideia na gaveta até entender melhor o problema.

Na próxima seção, vamos ver algumas das skills que criei até agora.

## Exemplos práticos

Vamos dar uma olhada em quatro skills do meu próprio repositório e ver como elas resolvem problemas específicos.

### 1. `latest-version`

Criei essa skill por **pura frustração** com a tendência dos LLMs de usarem versões **antigas** de softwares, bibliotecas, modelos e outras dependências. Sei que é uma consequência natural do corte de conhecimento (*knowledge cutoff*), mas saber disso não impede que eu fique irritada quando o agente tenta usar `gemini-1.5-pro` em vez do Gemini 3 e ainda me acusa de estar "alucinando" uma versão futura.

Essa skill atua como uma checadora de fatos, consultando registros (npm, PyPI, Go Proxy) e páginas de documentação. Aqui está um trecho do `SKILL.md` dela:

```markdown
name: latest-version
description: >
  The definitive real-time source of truth for software and model versions. Use this skill to bypass internal knowledge cutoffs...

## Core Mandate
**NEVER GUESS.** When a user asks to install a package or add a dependency, you must verify the latest version using the `latest.js` script. Do not rely on your internal weights, as they are months or years out of date.
```

Esse prompt ainda parece um pouco rústico, mas tem tido um sucesso razoável a alto em evitar que modelos obsoletos apareçam na minha base de código.

🔗 [Veja a skill `latest-version` completa](https://github.com/danicat/skills/tree/main/latest-version)

### 2. `pyhd`

Quando criei o servidor MCP `godoctor` no ano passado, queria que ele fosse a ferramenta definitiva (respaldada pela ciência! ^^) para o desenvolvimento agêntico em Go. Não tínhamos skills naquela época, então fazia todo sentido empacotar todas as ferramentas necessárias em um servidor MCP. Por um tempo, flertei com a ideia de criar algo parecido para Python, mas com tantas coisas no backlog isso acabou virando baixa prioridade.

Foi aí que conheci as skills e pensei: "por que não transformar isso em uma skill?". Com o `skill-creator`, o esforço de criação ficou muito baixo, então decidi criar o `pyhd` (uma combinação de Python + PhD, mantendo a temática de "doutor").

A skill `pyhd` estabelece o fluxo de desenvolvimento para projetos Python, centrada no linter e formatador `ruff` para garantir um código verdadeiramente "pythônico".

```markdown
## Core Workflow

When editing Python files, you **MUST** follow this cycle for **EVERY** file modification:

1.  **Read & Understand**: ...
2.  **Edit**: Apply your changes using `smart_edit` or `replace`.
3.  **Sanitize (Ruff)**:
    Immediately after editing, run the following commands to format and fix linting issues:
    `uv run ruff check --fix <filename>`
    `uv run ruff format <filename>`
4.  **Verify**: Run tests...
```

Essa skill garante que toda alteração em arquivos Python seja imediatamente seguida de linting e formatação padronizados, o que ajuda a pegar problemas logo no início. Enquanto não encontro tempo para implementar um "pydoctor" completo, essa é a minha skill de cabeceira para desenvolvimento em Python.

🔗 [Veja a skill `pyhd` completa](https://github.com/danicat/skills/tree/main/pyhd)

### 3. `find-examples`

Às vezes, precisamos ver como uma biblioteca específica é usada no mundo real. A skill `find-examples` utiliza um script Python (`github_search.py`) para pesquisar no GitHub códigos reais que usam a dependência desejada no projeto. Criei essa skill para ajudar a combater o hábito dos modelos de alucinar APIs quando poderiam simplesmente consultar documentação ou exemplos reais.

Por usar apenas a busca do GitHub, ela não precisa de personal access token e costuma ter um desempenho muito melhor do que uma busca no Google.

```markdown
### 1. Search for Repositories (Multi-Language)
Run the `github_search.py` script. If you can't find many examples in your target language, add related languages supported by the SDK.

### 4. Clone and Inspect
Clone the selected repositories into the `_examples` folder.
Once cloned, use `list_files`, `smart_read`, or `grep_search` to find relevant implementation details.
```

Também adicionei um recurso em que ela tenta encontrar exemplos em diferentes linguagens para SDKs poliglotas. Como é uma das minhas skills mais recentes, ainda não a testei tão a fundo, mas achei um ótimo exemplo para compartilhar aqui.

🔗 [Veja a skill `find-examples` completa](https://github.com/danicat/skills/tree/main/find-examples)

### 4. `de-sloppify`

Uso essa skill para identificar padrões comuns de escrita gerada por IA. Ela traz um script que calcula um "slop score" com base na escolha de palavras, variação no tamanho das frases e repetições estruturais.

O script usa o NLTK para realizar tagging morfossintático (POS tagging), o que ajuda a detectar a alta densidade de substantivos e o uso excessivo de voz passiva típicos de textos de IA sem edição. Ele roda localmente e gera um relatório detalhado dos marcadores encontrados.

🔗 [Veja a skill `de-sloppify` completa](https://github.com/danicat/skills/tree/main/de-sloppify)

## Conclusões

Skills raramente nascem perfeitas na primeira tentativa. A melhor forma de refiná-las é pelo uso no dia a dia. Quando notar que o agente está se batendo em alguma etapa ou buscando o contexto errado, peça para ele atualizar a skill usando o `skill-creator` novamente.

Dê uma olhada nos seus fluxos de trabalho diários. Em quais tarefas você precisa ficar lembrando a IA das mesmas regras o tempo todo? Essas são as candidatas ideais para a sua próxima skill personalizada.

Pronta para construir sua primeira skill? Confira a [documentação oficial](https://geminicli.com/docs/cli/skills/) para aprender o básico e, se quiser inspiração, dê uma olhada no [repositório danicat/skills no GitHub](https://github.com/danicat/skills).

Happy coding!

Dani =^.^=
