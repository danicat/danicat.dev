---
categories:
- Agentic Coding
date: 2026-01-29 12:00:00+00:00
heroStyle: big
series:
- Agent Skills
series_order: 1
slug: "agent-skills-gemini-cli"
summary: Desbloqueie expertise sob demanda para seu agente de IA. Aprenda a usar Agent
  Skills na Gemini CLI para construir fluxos de trabalho modulares, escaláveis e autônomos.
tags:
  - agent-skills
  - gemini-cli
  - mcp
  - vibe-coding
title: "Dominando Agent Skills na Gemini CLI"
aliases:
  - "/pt-br/posts/20260128-agent-skills-gemini-cli/"
description: "Aprenda a construir e configurar Agent Skills na Gemini CLI. Cobre progressive disclosure, carregamento de prompts sob demanda e scripts determinísticos de análise em Python."
proficiencyLevel: "Intermediate"
dependencies:
  - "Gemini CLI >= 0.1.0"
  - "Python 3.10+"
---

{{< alert "circle-info" >}}
**Atualização (2026):** A Gemini CLI evoluiu para o **Google Antigravity 2.0**. Embora os conceitos fundamentais e a estrutura de Agent Skills abordados neste artigo continuem sendo a base, confira [O Guia do Mochileiro para o Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}) para uma visão geral da plataforma, e leia [O Guia Pragmático para Agent Skills]({{< ref "/posts/20260829-the-pragmatic-guide-to-agent-skills" >}}) para padrões modernos de produção, catálogos de skills e otimização de tokens.
{{< /alert >}}

Quando escrevi sobre o [Tenkai]({{< ref "/posts/20260120-improving-agentic-coding-with-science/" >}}) na semana passada, deixei de abordar um aspecto importante na análise de experimentos: como extrair insights deles. Embora eu tenha uma interface agradável com resumos, métricas estatísticas e testes, é muito difícil capturar as nuances de cada configuração apenas a partir de um resumo.

Por exemplo: frequentemente percebo que operações de leitura (como `read_file` ou o `smart_read` do godoctor) estão fortemente correlacionadas com cenários que falharam ou levaram mais tempo para concluir. Isso acontece porque as operações de leitura são ruins? Não: é porque, para se recuperar de um erro, o agente precisou atualizar seu conhecimento do código-fonte relendo-o. Portanto, embora haja uma forte correlação entre leitura, lentidão e falhas, isso de modo algum indica uma relação de causalidade ou, como os estatísticos adoram dizer, "correlação não implica causalidade".

Como venho realizando diversos experimentos ao longo das últimas semanas, percebi rápido que ensinar o modelo a executar análises mais profundas a cada vez não era muito eficaz. Tipicamente, em cenários assim, ou eu adiciono as instruções de análise ao contexto do meu agente (via `GEMINI.md`) ou armazeno os prompts necessários em um servidor MCP para poder mapeá-los para comandos de barra (slash commands).

Embora ambas as alternativas funcionem, elas têm suas limitações. Expandir o contexto do agente para cada tarefa possível que eu queira executar resultará em inchaço de contexto (*context bloat*) e um comportamento menos eficaz. Criar comandos de barra para cada prompt depende de eu invocar explicitamente o comando, já que o agente não tem conhecimento deles por design.

Felizmente, as **Agent Skills** oferecem uma solução que combina o poder de ambos. Agent Skills são um novo recurso na [Gemini CLI](https://geminicli.com) projetado para dar ao agente capacidades sob demanda. Elas se comportam de maneira semelhante a uma ferramenta de agente (na verdade, as skills são ativadas por uma tool call), mas a skill permite acesso sob demanda a um prompt e a arquivos de suporte para permitir que o agente realize tarefas especializadas, inserindo-os no contexto apenas quando forem necessários.

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

O arquivo `SKILL.md` é onde reside o prompt da skill. Ele traz um pequeno frontmatter para definir o nome e a descrição; fora isso, é um arquivo markdown comum:

```text
---
name: <nome-unico>
description: <o que a skill faz e quando o Gemini deve usá-la>
---

<suas instruções sobre como o agente deve se comportar / usar a skill>
```

Para adicionar uma skill ao seu projeto, você pode criar uma pasta dentro de `.gemini/skills`. Por exemplo, a `my-skill` acima ficaria em `.gemini/skills/my-skill`. A Gemini CLI buscará automaticamente por skills na seguinte ordem de precedência:

1. Workspace (<meu-projeto>/.gemini/skills)
2. Usuário (~/.gemini/skills)
3. Extensões (~/.gemini/extensions/<nome-da-extensao>/skills)

O ponto fundamental a notar é que, quando a Gemini CLI inicia, ela só tem conhecimento do nome e da descrição da skill. Todo o restante será carregado **sob demanda** assim que a skill for ativada.

Agora vamos dar uma olhada em como estou usando uma skill para aprimorar meu próprio fluxo de trabalho de análise de experimentos.

## A skill `experiment-analyst`

Projetei a skill `experiment-analyst` para ser ativada sempre que peço à Gemini CLI para avaliar um experimento. Ela é organizada da seguinte forma:

```text
experiment-analyst/
├── SKILL.md                     <-- As diretrizes de análise
├── references/
│   └── tenkai_db_schema.md      <-- O esquema do banco de dados, para que o agente não precise descobri-lo toda vez
└── scripts/
    ├── analyze_experiment.py    <-- Replica parte da análise que tenho no frontend
    ├── analyze_patterns.py      <-- Mergulhos profundos em padrões comuns para extrair insights
    ├── get_experiment_config.py <-- Recupera os detalhes da configuração do experimento
    └── success_determinants.py  <-- Análise de chamadas de ferramentas e correlação
```

### Definindo a persona especialista

O arquivo `SKILL.md` define o procedimento analítico. Ele busca atingir um equilíbrio ao ensinar o agente sobre o que fazer, mas sem seguir uma fórmula engessada ("cookie-cutter"). Um dos aspectos importantes é evitar que o agente tire conclusões precipitadas, definindo uma persona mais fundamentada. Ainda valido todas as afirmações e recebo todas as conclusões com certa cautela, mas essa versão me proporcionou insights interessantes que, de outra forma, exigiriam muito trabalho manual para descobrir.

```text
---
name: experiment-analyst
description: Expertise in analysing Tenkai agent experiments. Use when asked to "analyse experiment X" to determine success factors, failure modes, and behavioural patterns.
---

# Experiment Analyst

## Core Mandates
1. **Evidence-Based:** Never make claims without data. Cite specific Run IDs.
2. **Correlation ≠ Causation:** A tool might be correlated with failure (e.g., `read_file`) because it's used for recovery. Always investigate the *context* of usage.
3. **Comparative:** Always contrast the performance of alternatives.
```

Nota: você pode clicar aqui para ver o arquivo [SKILL.md](https://github.com/danicat/skills/blob/main/experiment-analyst/SKILL.md) completo.

### Os recursos da skill

Você vai me ouvir falar muito sobre isso nas próximas semanas: ao lidar com agentes, que são inerentemente **não-determinísticos**, a única maneira de assegurar qualidade é fornecendo ferramentas **determinísticas**. As skills se encaixam perfeitamente nessa filosofia porque podemos agrupá-las com scripts para realizar tarefas de maneira consistente, em vez de deixar para o agente "adivinhar" como deve ser feito.

Para a skill de análise de experimentos, eu queria que o agente tivesse liberdade para explorar, mas também não queria que ele ficasse reinventando a roda o tempo todo. Por isso, ela vem com alguns scripts pré-empacotados:

- `analyse_experiment.py`: reproduz um resumo do experimento semelhante ao que tenho no frontend, mas inclui alguns agrupamentos de chamadas de ferramentas para comandos de shell
- `analyse_patterns.py`: extrai amostras da conversa do agente para tentar identificar padrões de uso de ferramentas
- `get_experiment_config.py`: ajuda o agente a entender o experimento recuperando sua definição
- `success_determinants.py`: calcula a correlação entre desfechos bem-sucedidos e chamadas de ferramentas

Eu forneço o esquema do banco de dados em `references/tenkai_db_schema.md` para quando o agente decidir fazer consultas ad-hoc, evitando que precise redescobrir o schema toda vez (esse schema é bastante estável entre as execuções).

Não vou afirmar que essa configuração seja perfeita, já que não passei um tempo significativo refinando-a, mas essa combinação de informações e scripts pré-empacotados cobre a maioria das perguntas que normalmente peço ao agente para explorar.

## Considerações finais

As Agent Skills representam uma mudança significativa na forma como projetamos fluxos de trabalho agênticos. Ao nos afastarmos de prompts de contexto gigantescos (como adicionar tudo ao `GEMINI.md`) em direção a capacidades modulares e sob demanda, resolvemos dois problemas de uma só vez: mantemos o contexto do nosso agente limpo (menos tokens) e viabilizamos uma expertise profunda e especializada que não dilui o desempenho geral.

No meu caso, a skill `experiment-analyst` foi fundamental para transformar uma tarefa repetitiva em um fluxo semiautomatizado. Ela me dá consistência e flexibilidade ideais para realizar as análises que desejo. Agora estou considerando atualizar outras partes do meu fluxo de trabalho para skills, superando a minha abordagem anterior de usar servidores MCP apenas como "bancos de dados de prompts".

Estou animada para ver o que a comunidade vai construir. Então, dê uma olhada em seus próprios fluxos de trabalho: onde você está constantemente repetindo instruções? Onde você precisa de um especialista? Essa é a sua próxima skill esperando para ser escrita.

**Atualização:** Você já pode conferir a [Parte 2: Criando Agent Skills com o skill-creator]({{< ref "/posts/20260227-gemini-cli-skills-part-2/" >}}), onde mergulhamos em exemplos práticos.

Bons códigos!
