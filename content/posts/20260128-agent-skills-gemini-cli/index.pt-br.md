---
title: "Dominando Agent Skills na Gemini CLI"
date: 2026-01-29T12:00:00Z
draft: false
summary: "Desbloqueie expertise sob demanda para seu agente de IA. Aprenda a usar Agent Skills na Gemini CLI para construir fluxos de trabalho modulares, escaláveis e autônomos."
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["gemini-cli", "agent-skills", "mcp", "vibe-coding"]
series: ["Agent Skills"]
series_order: 1
slug: "agent-skills-gemini-cli"
heroStyle: "big"
---

Quando escrevi sobre o [Tenkai]({{< ref "/posts/20260120-improving-agentic-coding-with-science/" >}}) na semana passada, não cobri um aspecto importante sobre a análise de experimentos: como extrair insights dos experimentos. Embora eu tenha um frontend bacana com resumos, métricas estatísticas e testes, é muito difícil capturar as nuances de cada configuração apenas com um resumo.

Por exemplo, frequentemente percebo que operações de leitura (ex: `read_file` ou `smart_read` do godoctor) estão fortemente correlacionadas com cenários que falharam ou levaram mais tempo para completar. Isso acontece porque as operações de leitura são ruins? Não, é porque, para se recuperar de um erro, o agente precisou atualizar seu conhecimento do código-fonte lendo-o novamente. Então, embora exista uma forte correlação entre leitura, lentidão e falha, de forma alguma isso é uma relação de causalidade ou, como os estatísticos gostam de dizer, "correlação não implica causalidade".

Como tenho feito muitos experimentos nas últimas semanas, percebi rapidamente que ensinar o modelo a realizar análises mais profundas toda vez não era muito eficaz. Normalmente, nesses cenários, ou eu adiciono as instruções de análise ao meu agent context (via `GEMINI.md`), ou armazeno os prompts necessários em um servidor MCP para poder mapeá-los para slash commands.

Embora ambas as alternativas funcionem, elas têm suas limitações. Expandir o agent context para cada tarefa possível que eu queira realizar resultará em inchaço do contexto (context bloat) e comportamento menos eficaz. Criar slash commands para cada prompt depende de eu invocar explicitamente o comando, já que o agente não tem conhecimento deles por design.

Felizmente, as **Agent Skills** fornecem uma solução que combina o poder de ambos. Agent Skills são um novo recurso na [Gemini CLI](https://geminicli.com) projetado para dar ao agente capacidades sob demanda. Ele se comporta de maneira semelhante a uma agent tool (na verdade, skills são ativadas por uma tool call), mas a skill permite acesso sob demanda a um prompt e arquivos de suporte para permitir que o agente realize tarefas especializadas, colocando-os no contexto apenas quando são necessários.

Você pode encontrar as especificações técnicas completas na [documentação oficial](https://geminicli.com/docs/cli/skills/), mas neste artigo vou cobrir o básico para você começar.

## Anatomia de uma skill

Uma skill nada mais é do que uma pasta com um prompt e, opcionalmente, arquivos de suporte como documentação e scripts.

```text
my-skill/
├── SKILL.md       (Obrigatório) Instruções e metadados
├── scripts/       (Opcional) Scripts/ferramentas executáveis
├── references/    (Opcional) Documentação estática e exemplos
└── assets/        (Opcional) Templates e recursos binários
```

O arquivo `SKILL.md` é onde vive o prompt da skill. Ele tem um pequeno frontmatter para definir o nome e a descrição da skill, mas, fora isso, é apenas um arquivo markdown comum:

```text
---
name: <nome-unico>
description: <o que a skill faz e quando o Gemini deve usá-la>
---

<suas instruções sobre como o agente deve se comportar / usar a skill>
```

Para adicionar uma skill ao seu projeto, você pode criar uma pasta em `.gemini/skills`. Por exemplo, a `my-skill` acima ficaria em `.gemini/skills/my-skill`. A Gemini CLI procurará automaticamente por skills na seguinte ordem de precedência:

1. Workspace (<meu-nome-do-projeto>/.gemini/skills)
2. User (~/.gemini/skills)
3. Extensions (~/.gemini/extensions/<nome-da-extensao>/skills)

O importante a notar é que, quando a Gemini CLI inicia, ela só tem conhecimento do nome e da descrição da skill. Todo o resto será carregado **sob demanda** quando a skill for ativada.

Agora vamos dar uma olhada em como estou usando uma skill para melhorar meu fluxo de trabalho de análise de experimentos.

## A skill `experiment-analyst`

Projetei a skill `experiment-analyst` para ser ativada quando peço à Gemini CLI para avaliar um experimento. Ela é organizada da seguinte forma:

```text
experiment-analyst/
├── SKILL.md                     <-- As diretrizes de análise
├── references/
│   └── tenkai_db_schema.md      <-- O esquema do banco de dados, para que o agente não precise descobri-lo toda vez
└── scripts/
    ├── analyze_experiment.py    <-- Replica parte da análise que tenho no frontend
    ├── analyze_patterns.py      <-- Alguns mergulhos profundos em padrões comuns para extrair insights
    ├── get_experiment_config.py <-- Recupera os detalhes da configuração do experimento
    └── success_determinants.py  <-- Análise de chamadas de ferramentas e correlação
```

### Definindo a persona especialista

O arquivo `SKILL.md` define o procedimento analítico. Ele tenta alcançar um equilíbrio ao ensinar o agente o que fazer, mas não de uma maneira simples e padronizada ("cookie-cutter"). Um dos aspectos importantes é tentar desviar o agente de tirar conclusões precipitadas, definindo uma persona mais fundamentada. Ainda valido todas as afirmações e recebo todas as conclusões com um grão de sal, mas esta versão me proporcionou alguns insights interessantes que, de outra forma, me dariam muito trabalho para descobrir.

```text
---
name: experiment-analyst
description: Expertise em analisar experimentos do agente Tenkai. Use quando solicitado a "analisar experimento X" para determinar fatores de sucesso, modos de falha e padrões comportamentais.
---

# Experiment Analyst

## Core Mandates
1. **Evidence-Based:** Nunca faça afirmações sem dados. Cite Run IDs específicos.
2. **Correlation ≠ Causation:** Uma ferramenta pode estar correlacionada com falha (ex: `read_file`) porque é usada para recuperação. Sempre investigue o *contexto* de uso.
3. **Comparative:** Sempre compare o desempenho de alternativas.
```

### Os assets da skill

Você vai me ouvir falar muito sobre isso nas próximas semanas, mas ao lidar com agentes, que são inerentemente **não-determinísticos**, a única maneira de garantir qualidade é dar a eles ferramentas **determinísticas**. Skills se encaixam bem nessa filosofia porque podemos agrupá-las com scripts para realizar tarefas de maneira consistente, em vez de deixar para o agente "adivinhar" como é feito.

Para a skill de análise de experimentos, eu queria que o agente tivesse liberdade para explorar, mas também não quero que ele reinvente a roda o tempo todo, então ela vem com alguns scripts pré-empacotados:

- `analyse_experiment.py`: reproduz um resumo do experimento semelhante ao que tenho no frontend, mas inclui algum agrupamento em tool calls para comandos shell
- `analyse_patterns.py`: extrai amostras da conversa do agente para tentar identificar padrões de uso de ferramentas
- `get_experiment_config.py`: ajuda o agente a entender o experimento recuperando sua definição
- `success_determinants.py`: calcula a correlação entre resultados bem-sucedidos e tool calls

Eu forneço o esquema do banco de dados em `references/tenkai_db_schema.md` para quando o agente decide fazer consultas ad-hoc, para que ele não precise redescobrir o esquema toda vez (esse esquema é bastante estável entre as execuções).

Não vou afirmar que essa configuração é perfeita, pois não passei muito tempo refinando-a, mas essa combinação de informações e scripts pré-empacotados cobre a maioria das perguntas que tipicamente peço para o agente explorar.

## Pensamentos finais

Agent Skills representam uma mudança significativa na forma como projetamos fluxos de trabalho agênticos. Ao nos afastarmos de prompts de contexto massivos e monolíticos (ex: `GEMINI.md`) e irmos em direção a capacidades modulares e sob demanda, resolvemos dois problemas de uma só vez: mantemos o contexto do nosso agente limpo (menos tokens) e permitimos expertise profunda e especializada que não dilui o desempenho geral.

No meu caso, a skill `experiment-analyst` foi útil para transformar uma tarefa repetitiva em um fluxo semiautomatizado. Ela me dá consistência e flexibilidade suficientes para realizar as análises que desejo. Agora estou considerando atualizar outras partes do meu fluxo de trabalho para skills, afastando-me da minha abordagem de usar servidores MCP como "bancos de dados de prompts".

Estou animada para ver o que a comunidade vai construir. Então, dê uma olhada em seus próprios fluxos de trabalho. Onde você está constantemente repetindo instruções? Onde você precisa de um especialista? Essa é a sua próxima skill esperando para ser escrita.

**Atualização:** Você já pode ler a [Parte 2: Construindo Agent Skills com skill-creator]({{< ref "/posts/20260227-gemini-cli-skills-part-2/" >}}) onde mergulhamos em exemplos práticos.

Happy coding!
