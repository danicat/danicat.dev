---
categories:
- Agentic Coding
date: 2026-01-29 12:00:00+00:00
draft: false
heroStyle: big
series:
- Agent Skills
series_order: 1
slug: agent-skills-gemini-cli
summary: Desbloqueie expertise sob demanda para seu agente de IA. Aprenda a usar Agent
  Skills na Gemini CLI para construir fluxos de trabalho modulares, escaláveis e autônomos.
tags:
  - agent-skills
  - gemini-cli
  - mcp
  - vibe-coding
title: Dominando Agent Skills na Gemini CLI
---

{{< alert "circle-info" >}}
**Atualização (2026):** A Gemini CLI evoluiu para o **Google Antigravity 2.0**. Embora os conceitos fundamentais e a estrutura de Agent Skills apresentados neste artigo continuem sendo a base, confira [O Guia do Mochileiro para o Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}) para uma visão geral da nova plataforma e seus recursos.
{{< /alert >}}

Quando escrevi sobre o [Tenkai]({{< ref "/posts/20260120-improving-agentic-coding-with-science/" >}}) na semana passada, deixei de abordar um aspecto essencial na análise de experimentos: como extrair insights deles. Embora eu conte com uma interface amigável, com resumos, métricas estatísticas e testes, é muito difícil capturar as nuances de cada configuração apenas a partir de um resumo.

Por exemplo: frequentemente percebo que operações de leitura (como `read_file` ou o `smart_read` do godoctor) estão fortemente correlacionadas com cenários que falharam ou levaram mais tempo para rodar. Isso acontece porque as operações de leitura são ruins? Não: é porque, para se recuperar de um erro, o agente precisou atualizar seu conhecimento do código-fonte relendo os arquivos. Portanto, embora haja uma forte correlação entre leituras, lentidão e falhas, isso de modo algum indica causalidade ou, como os estatísticos adoram lembrar, "correlação não implica causalidade".

Como venho realizando diversos experimentos ao longo das últimas semanas, percebi rápido que ensinar o modelo a executar análises mais profundas a cada rodada não era nada produtivo. Em cenários assim, normalmente ou eu adiciono as instruções de análise ao contexto do agente (via `GEMINI.md`) ou armazeno os prompts necessários em um servidor MCP para acioná-los por comandos de barra (slash commands).

Embora ambas as alternativas funcionem, elas têm suas limitações. Inflar o contexto do agente para cada tarefa possível resulta em inchaço de contexto (context bloat) e um comportamento menos eficiente. Por outro lado, criar comandos de barra para cada prompt depende de uma invocação manual explícita, já que o agente, por design, não sabe da existência deles.

Felizmente, as **Agent Skills** oferecem uma solução que une o melhor dos dois mundos. Agent Skills são um novo recurso na [Gemini CLI](https://geminicli.com) projetado para dotar o agente de capacidades sob demanda. Elas funcionam de modo semelhante a uma ferramenta (na verdade, uma skill é ativada por uma tool call), mas oferecem acesso dinâmico a um prompt e a arquivos de suporte para que o agente execute tarefas especializadas, inserindo essas informações no contexto apenas no momento em que forem necessárias.

Você pode consultar as especificações técnicas completas na [documentação oficial](https://geminicli.com/docs/cli/skills/), mas neste artigo vou cobrir o básico para você começar.

## Anatomia de uma skill

Uma skill nada mais é do que uma pasta contendo um prompt e, opcionalmente, arquivos de suporte como documentação e scripts.

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

Para adicionar uma skill ao seu projeto, você pode criar uma pasta dentro de `.gemini/skills`. Por exemplo, a `my-skill` acima ficaria em `.gemini/skills/my-skill`. A Gemini CLI buscará automaticamente por skills na seguinte ordem de prioridade:

1. Workspace (<meu-projeto>/.gemini/skills)
2. Usuário (~/.gemini/skills)
3. Extensões (~/.gemini/extensions/<nome-da-extensao>/skills)

O ponto fundamental a notar é que, quando a Gemini CLI inicia, ela só tem conhecimento do nome e da descrição da skill. Todo o restante será carregado **sob demanda** assim que a skill for ativada.

Agora vamos dar uma olhada em como estou usando uma skill para aprimorar meu próprio fluxo de análise de experimentos.

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

O arquivo `SKILL.md` define o procedimento analítico. Ele busca atingir um equilíbrio ao orientar o agente sobre o que fazer, fugindo de fórmulas engessadas ou prontas ("cookie-cutter"). Um aspecto essencial é evitar que o agente tire conclusões precipitadas, delimitando uma persona mais fundamentada e crítica. Ainda valido todas as afirmações e recebo todas as conclusões com uma boa pitada de ceticismo, mas essa versão me proporcionou insights valiosos que, de outra forma, exigiriam muito esforço manual para descobrir.

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

Para a skill de análise de experimentos, eu queria que o agente tivesse liberdade para investigar, mas sem reinventar a roda o tempo todo. Por isso, ela vem com alguns scripts pré-empacotados:

- `analyse_experiment.py`: reproduz um resumo do experimento semelhante ao que tenho no frontend, mas inclui agrupamentos de chamadas de ferramentas para comandos de terminal
- `analyse_patterns.py`: extrai amostras da conversa do agente para tentar identificar padrões de uso de ferramentas
- `get_experiment_config.py`: ajuda o agente a entender o experimento recuperando sua definição
- `success_determinants.py`: calcula a correlação entre desfechos bem-sucedidos e chamadas de ferramentas

Eu forneço o esquema do banco de dados em `references/tenkai_db_schema.md` para quando o agente decidir fazer consultas ad-hoc, evitando que precise redescobrir o schema toda vez (essa estrutura é bastante estável entre as execuções).

Não vou afirmar que essa configuração seja perfeita, já que não passei um tempo significativo refinando cada detalhe, mas essa combinação de informações e scripts pré-empacotados cobre a grande maioria das perguntas que normalmente peço ao agente para explorar.

## Considerações finais

As Agent Skills representam uma mudança significativa na forma como projetamos fluxos de trabalho agênticos. Ao nos afastarmos de prompts de contexto gigantescos e monolíticos (como adicionar tudo ao `GEMINI.md`) em direção a capacidades modulares e sob demanda, resolvemos dois problemas de uma só vez: mantemos o contexto do nosso agente enxuto (menos tokens) e viabilizamos uma expertise profunda e especializada que não dilui o desempenho geral.

No meu caso, a skill `experiment-analyst` foi fundamental para transformar uma tarefa repetitiva em um fluxo semiautomatizado. Ela me dá consistência e flexibilidade ideais para realizar as análises que desejo. Agora estou considerando migrar outras partes do meu fluxo de trabalho para skills, superando a minha abordagem anterior de usar servidores MCP apenas como "bancos de dados de prompts".

Estou animada para ver o que a comunidade vai construir. Então, dê uma olhada em seus próprios fluxos de trabalho: onde você está constantemente repetindo instruções? Em que tarefas você sente falta de um especialista? Essa é exatamente a sua próxima skill esperando para ser escrita.

**Atualização:** Você já pode conferir a [Parte 2: Construindo Agent Skills com skill-creator]({{< ref "/posts/20260227-gemini-cli-skills-part-2/" >}}), onde mergulhamos em exemplos práticos.

Bons códigos!

Dani =^.^=
