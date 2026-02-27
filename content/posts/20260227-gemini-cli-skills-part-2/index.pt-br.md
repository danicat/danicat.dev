---
title: "Construindo Agent Skills com skill-creator"
date: 2026-02-27T12:00:00Z
summary: "Aprenda a usar o skill-creator integrado da Gemini CLI para gerar, refinar e estruturar automaticamente suas próprias Agent Skills personalizadas usando exemplos práticos."
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["gemini-cli", "agent-skills", "vibe-coding"]
series: ["Agent Skills"]
series_order: 2
heroStyle: "big"
---

No [artigo anterior]({{< ref "/posts/20260128-agent-skills-gemini-cli/" >}}), exploramos como as Agent Skills adicionam novos recursos à Gemini CLI. Analisamos a skill `experiment-analyst` como um exemplo prático de como manter o contexto do agente limpo, fornecendo-lhe instruções específicas para uma tarefa.

Hoje, vamos mergulhar nos princípios fundamentais do design de skills e analisar alguns exemplos práticos de skills que utilizo diariamente.

## O que são Agent Skills

Vamos fazer uma rápida recapitulação da parte 1, caso você tenha perdido. Agent Skills são um padrão aberto projetado para dar aos agentes de codificação conhecimento especializado "just in time". Elas são projetadas para que o conhecimento especializado seja adicionado ao contexto apenas quando necessário, ajudando a evitar o chamado inchaço de contexto (context bloat). O termo técnico para isso é **Progressive Disclosure** (Divulgação Progressiva): mantemos as instruções principais (`SKILL.md`) o mais enxutas possível e usamos arquivos separados para referências detalhadas ou scripts que são carregados apenas *quando necessário*.

No disco, uma skill é uma pasta com um arquivo `SKILL.md` e, opcionalmente, recursos integrados:

```text
nome-da-skill/
├── SKILL.md (Obrigatório: Apenas nome, descrição e instruções principais)
└── Recursos Integrados (Opcional)
    ├── scripts/    (Código executável para tarefas repetíveis)
    ├── references/ (Docs carregados sob demanda, ex: esquemas de API)
    └── assets/     (Templates ou arquivos binários usados nas saídas)
```

## A skill `skill-creator`

Você sempre pode escrever skills manualmente, mas a Gemini CLI vem com uma meta-skill integrada chamada `skill-creator`, que torna as coisas muito mais fáceis.

Você pode ativar esta skill pedindo à Gemini CLI para criar (ou refatorar) uma skill:
> *"Quero criar uma nova skill para buscar a versão real mais recente de um pacote de software para que você pare de alucinar versões."*

Quaisquer solicitações relacionadas a "criar skills" devem acionar automaticamente o `skill-creator`, mas, caso você esteja falando com um modelo "rabugento", também pode ser mais explícita:

> *"Use o skill-creator para escrever uma skill para des-sloppificar textos gerados por IA (por favor, não leve para o lado pessoal)"*

A Gemini CLI pode pedir alguns detalhes antes de escrever o boilerplate da skill. Ela aprendeu recentemente a perguntar ao usuário com a ferramenta [`ask_user`](https://geminicli.com/docs/tools/ask-user/#ask_user-ask-user) e é muito legal vê-la em ação.

## Quando criar skills

No meu fluxo de trabalho pessoal, tenho dois usos principais para as skills:

1. Para documentar um processo que é específico para o meu trabalho (ex: como fazer um code review do jeito que eu gosto, como inicializar um repositório, como avaliar um post de blog, etc.)
2. Para adicionar conhecimento especializado sobre uma ferramenta, linguagem ou tecnologia específica (ex: como funciona um projeto genkit, como trabalhar com adk para desenvolver agentes, etc.)

Até certo ponto, você pode pensar em skills como um conceito intermediário entre comandos de barra (que eu costumo armazenar como prompts MCP) e ferramentas. Ao construir comandos de barra, quero criar um "processo repetível"; ao criar ferramentas, muitas vezes quero dar ao modelo uma maneira determinística de fazer algo. Como as skills podem ter tanto prompts quanto scripts, elas podem fazer ambos, com os scripts desempenhando o papel de ferramentas.

Claro, se você estiver empacotando suas skills como parte de uma extensão, há uma grande chance de que elas sejam enviadas com um servidor MCP que também expõe ferramentas. Você também pode aproveitar essa integração na definição da skill, ensinando o modelo a usar suas ferramentas MCP, se disponíveis.

Existem também dois momentos principais em que crio skills:

1. Após uma sessão dolorosamente longa tentando ensinar um modelo a fazer algo para mim (ex: "por favor, consolide o conhecimento do que acabamos de fazer em uma skill que possamos usar mais tarde")
2. Logo após ter uma nova ideia de algo que acredito que me ajudará a ser mais eficiente no meu trabalho (ex: "vamos escrever uma skill de de-slopify para melhorar suas capacidades de escrita")

Em ambos os casos, a skill nunca estará pronta na primeira tentativa, mas assim que eu começar a usá-la, irei refiná-la até o ponto em que estiver confiante de que ela traz algum valor, ou matá-la e estacionar a ideia até aprender mais sobre o problema.

Na próxima seção, veremos algumas das skills que construí até agora.

## Exemplos práticos

Vamos dar uma olhada em quatro skills do meu próprio repositório e ver como elas resolvem problemas específicos.

### 1. `latest-version`

Criei esta skill por **pura frustração** com a forma como os LLMs em geral tendem a usar versões **antigas** de software, bibliotecas, modelos e outras dependências. Sei que é uma consequência natural dos cortes de conhecimento (knowledge cutoffs), mas saber disso não me impede de ficar irritada quando o agente tenta usar o `gemini-1.5-pro` em vez do Gemini 3 e acusa a **MIM** de "alucinar" uma versão futura.

Esta skill atua como um verificador de fatos consultando registros (npm, PyPI, Go Proxy) e páginas de documentação. Aqui está um trecho do seu `SKILL.md`:

```markdown
name: latest-version
description: >
  A fonte definitiva da verdade em tempo real para versões de software e modelos. Use esta skill para contornar cortes de conhecimento internos...

## Mandato Principal
**NUNCA ADIVINHE.** Quando um usuário pedir para instalar um pacote ou adicionar uma dependência, você deve verificar a versão mais recente usando o script `latest.js`. Não confie em seus pesos internos, pois eles estão meses ou anos desatualizados.
```

Este prompt ainda parece um pouco desleixado, mas tem tido sucesso moderado a alto em evitar que alguns modelos obsoletos apareçam em minhas bases de código.

🔗 [Veja a skill `latest-version` completa](https://github.com/danicat/skills/tree/main/latest-version)

### 2. `pyhd`

Quando criei o servidor MCP `godoctor` no ano passado, queria que ele fosse a ferramenta **definitiva** (apoiada pela ciência! ^^) para o desenvolvimento de Go agêntico. Não tínhamos skills naquela época, então parecia natural empacotar todas as ferramentas de que eu precisava em um servidor MCP. Por um tempo, flertei com a ideia de criar algo semelhante para Python, mas tenho tantas coisas no backlog que isso se tornou uma prioridade baixa para mim.

Então me deparei com as skills e pensei "por que não fazer uma skill em vez disso?". Com o `skill-creator`, tornou-se muito fácil escrevê-la, então decidi criar o `pyhd` (uma combinação de Python + PhD, apenas para manter o tema "doctor").

A skill `pyhd` contém o fluxo de trabalho de desenvolvimento para projetos Python, centrando-se no linter e formatador `ruff` para garantir um código "pythonic" adequado.

```markdown
## Fluxo de Trabalho Principal

Ao editar arquivos Python, você **DEVE** seguir este ciclo para **CADA** modificação de arquivo:

1.  **Ler e Entender**: ...
2.  **Editar**: Aplique suas alterações usando `smart_edit` ou `replace`.
3.  **Sanitizar (Ruff)**:
    Imediatamente após a edição, execute os seguintes comandos para formatar e corrigir problemas de linting:
    `uv run ruff check --fix <arquivo>`
    `uv run ruff format <arquivo>`
4.  **Verificar**: Execute testes...
```

Esta skill incentiva que cada edição em Python seja imediatamente seguida por linting e formatação padrão, o que ajuda a detectar alguns problemas precoces. Até que eu encontre tempo para uma implementação adequada de um "pydoctor", esta é a minha skill preferida para o desenvolvimento em Python.

🔗 [Veja a skill `pyhd` completa](https://github.com/danicat/skills/tree/main/pyhd)

### 3. `find-examples`

Às vezes, você precisa saber como uma biblioteca específica é usada no mundo real. A skill `find-examples` usa um script Python (`github_search.py`) para pesquisar no GitHub códigos que usam a dependência que você deseja usar em seu projeto. Desenvolvi esta skill para ajudar a resolver o problema de modelos alucinando APIs, quando eles claramente poderiam fazer melhor apenas olhando para alguma documentação ou exemplos.

Como ela usa apenas a pesquisa do GitHub, não precisa de um token de acesso pessoal e tende a ter um desempenho melhor do que uma pesquisa no Google.

```markdown
### 1. Pesquisar Repositórios (Multi-Linguagem)
Execute o script `github_search.py`. Se não conseguir encontrar muitos exemplos em sua linguagem de destino, adicione linguagens relacionadas suportadas pelo SDK.

### 4. Clonar e Inspecionar
Clone os repositórios selecionados na pasta `_examples`.
Uma vez clonados, use `list_files`, `smart_read` ou `grep_search` para encontrar detalhes de implementação relevantes.
```

Também adicionei um recurso onde ela tenta encontrar exemplos em diferentes linguagens para SDKs poliglotas. É uma das minhas skills mais recentes, então não a testei muito, mas achei que era um exemplo interessante de adicionar.

🔗 [Veja a skill `find-examples` completa](https://github.com/danicat/skills/tree/main/find-examples)

### 4. `de-sloppify`

Uso esta skill para verificar padrões comuns de escrita de IA. Ela inclui um script que calcula uma "pontuação de slop" (AI slop score) com base na escolha de palavras, variação no comprimento da frase e repetição estrutural.

Ela usa o NLTK para marcar partes do discurso, o que ajuda a identificar a alta densidade de substantivos e a voz passiva frequentemente encontradas em saídas de modelos não editadas. O script roda localmente e fornece um relatório sobre os marcadores específicos que encontrou.

🔗 [Veja a skill `de-sloppify` completa](https://github.com/danicat/skills/tree/main/de-sloppify)

## Conclusões

Skills raramente são perfeitas na primeira tentativa. Uma maneira eficaz de refiná-las é através do uso real. Quando você notar o agente lutando com uma etapa ou buscando o contexto errado, peça para ele atualizar a skill usando o `skill-creator` novamente.

Dê uma olhada em seus fluxos de trabalho diários. Quais tarefas exigem que você lembre constantemente a IA das regras? Essas são as candidatas ideais para sua próxima skill personalizada.

Pronta para construir sua primeira skill? Confira a [documentação oficial](https://geminicli.com/docs/cli/skills/) para aprender o básico e, se precisar de inspiração, verifique o [repositório danicat/skills no GitHub](https://github.com/danicat/skills).

Happy coding!

Dani =^.^=
