---
categories:
  - Agentic Coding
date: 2026-08-29
heroStyle: big
title: "O Guia Pragmático para Agent Skills"
slug: "the-pragmatic-guide-to-agent-skills"
aliases:
  - "/pt-br/posts/20260829-the-pragmatic-guide-to-agent-skills/"
summary: "Dominar a criação de Agent Skills customizadas é o maior salto de produtividade para desenvolvedores. Explore casos práticos em documentação, fluxos e traços."
description: "Um guia pragmático para criar Agent Skills customizadas para coding agents: explore casos de uso práticos em documentação, automação de processos e traços de personalidade."
tags:
  - agent-skills
  - antigravity
  - gemini-cli
  - vibe-coding
proficiencyLevel: "Intermediate"
---

Pode parecer difícil de acreditar, mas o padrão [**Agent Skills**](https://agentskills.io) não tem sequer um ano de vida. O [artigo original da Anthropic](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills) foi publicado em outubro de 2025, apresentando inicialmente as Agent Skills como uma extensão do Claude Code antes de se consolidar como um padrão aberto em dezembro de 2025. O padrão provou ser tão útil que, em pouquíssimo tempo, foi adotado por todos os principais coding agents e frameworks de desenvolvimento de agentes para equipar modelos de IA com instruções modulares, scripts determinísticos e fluxos especializados de domínio.

Avançando para os dias de hoje, a maioria dos desenvolvedores já conhece as skills e as vantagens do modelo de progressive disclosure, mas será que estamos usando e criando skills de forma realmente eficaz? Neste artigo, vamos explorar os principais desafios ao lidar com skills, desde a descoberta e proveniência até o desenvolvimento e otimização.

Todas as skills e exemplos mencionados neste texto foram testados usando o Antigravity CLI e o Gemini 3.7 Flash, mas incentivo você a experimentar mesmo se utilizar outro harness e/ou modelo. Sem mais delongas, vamos mergulhar fundo no universo das Agent Skills!

## Criar skills vs consumir skills

O eterno dilema da engenharia de software: devo construir a minha própria solução ou adotar algo pronto? Existem várias abordagens para esse problema, mas vou compartilhar minha visão pragmática:

```goat
      +-----------------------------------------+
      |  Is this process specific to your repo  |
      |              or codebase?               |
      +-----------------------------------------+
            |                             |
       Yes  |                             |  No
            |                             v
            |                 +-----------------------+
            |                 |  Do you know a skill  |<-------------------+
            |                 |   that does the job?  |                    |
            |                 +-----------------------+                    |
            |                    |                 |                       |
            |               Yes  |                 |  No                   |
            |                    v                 v                       |
            |         +--------------------+  +------------------+         |
            |         | Passes security &  |  | Spent too much   |         |
            |         |  integrity checks? |  |  time looking?   |         |
            |         +--------------------+  +------------------+         |
            |            |              |        |            |            |
            |       Yes  |           No |    Yes |         No |            |
            |            |              |        |            v            |
            |            |              |        |    +---------------+    |
            |            |              |        |    |  Search for   |    |
            |            |              |        |    |    a skill    |    |
            |            |              |        |    +---------------+    |
            |            |              |        |            |            |
            |            |              v        v            +------------+
            |            |         +------------------+
            |            |         |  Build your own  |
            +--------------------->|      skill       |
                         |         +------------------+
                         |                   |
                         v                   v
                  +-----------------------------------+
                  |              Profit!              |
                  +-----------------------------------+
```

Admito que o algoritmo é um pouco caótico, mas o ponto central é: a menos que você tenha uma fonte autoritativa para a skill, ou confiança absoluta de que não está se expondo a um vetor de ataque, é melhor construir a sua própria.

O esforço para criar skills é muito baixo, então há pouquíssimas circunstâncias em que eu desaconselharia essa criação. A principal seria a falta de conhecimento no domínio para avaliar a qualidade técnica da skill, já que isso cria um grande risco de alimentar uma câmara de eco de comportamentos ruins.

Se você sabe o que está fazendo, construir skills é o maior hack para desbloquear a produtividade com coding agents — e é exatamente por isso que acredito que todo desenvolvedor sério deveria dominar a habilidade de criar skills (com o perdão do trocadilho).

## Dominando a habilidade de criar skills

O segredo para criar skills úteis reside em observar os padrões do agente. Você pode não conseguir conceber uma skill logo na primeira sessão de código, mas quanto mais usar agentes e observar seu comportamento, mais notará lacunas de conhecimento e de comportamento que podem ser preenchidas com skills customizadas.

Isso ficará mais claro com alguns exemplos práticos, mas antes de falar sobre eles, vamos dar uma olhada nos diferentes tipos de skills que podemos criar:

### Nível 1: Documentação

Uma skill que explica uma tecnologia específica, ampliando ou substituindo o conhecimento nativo do modelo. É extremamente comum, por exemplo, que modelos recomendem bibliotecas ou SDKs obsoletos, ou usem padrões ultrapassados de código, simplesmente porque eram prevalentes nos seus dados de treinamento. Usar skills para atualizar o conhecimento do modelo sobre um software é um dos usos mais básicos dessa tecnologia.

O formato típico de uma skill de documentação é um único arquivo `SKILL.md` com o conhecimento atualizado, mostrando linhas de comando, snippets de código e URLs externas. Dependendo da complexidade do domínio, dividir o arquivo `SKILL.md` em arquivos separados de referência e/ou assets é ideal para maximizar o potencial do progressive disclosure.

Crie esse tipo de skill quando notar seu agente usando padrões obsoletos, bibliotecas desatualizadas, versões antigas de SDKs ou quando ele gerar repetidamente implementações longe do ideal. A maioria das skills fornecidas pelos fabricantes se encaixa nessa categoria, como o repositório [Google Skills](https://github.com/google/skills), que inclui skills para a maioria dos produtos do Google.

Um exemplo de skill de documentação que criei é a [ebitengineer](https://skills.danicat.dev/game-dev/ebitengineer/). Essa skill nasceu após eu notar problemas recorrentes enquanto criava jogos 2D com o Ebitengine, como a ordem incorreta de operações de matriz, modularização deficiente, falta de gerenciamento de estado, abuso de texto de debug em vez de fontes de produção e assim por diante.

### Nível 2: Processo

Uma skill que estabelece um processo ou workflow específico, como [fazer um code review]({{< ref "/posts/20260303-code-reviews-in-2026" >}}), conduzir uma auditoria de segurança ou analisar métricas de performance. Esse tipo de skill costuma empacotar não apenas conhecimento, mas scripts customizados ou ferramentas de CLI para executar tarefas, poupando o modelo do esforço de construí-los do zero a cada vez que você precisa deles.

Você deve criar skills de processo sempre que se pegar instruindo o modelo a fazer a mesma tarefa repetidamente, especialmente quando houver passos determinísticos que podem ser automatizados via script. Algumas skills de processo que utilizo com frequência fazem parte do conjunto de [analytics](https://skills.danicat.dev/analytics), usadas para coletar e analisar dados de redes sociais, Google Analytics e Google Search Console.

Otimizar minha presença online é uma parte importante do meu trabalho (de pouco adianta produzir conteúdo que ninguém vê), então criei essas skills para automatizar as tarefas que antes eu fazia manualmente inspecionando o site de analytics de cada plataforma, um por um. Agora tenho uma "central de comando" alimentada por essas skills, economizando um tempo precioso que posso dedicar a produzir mais e melhores conteúdos.

Olhe para o seu próprio processo. Quais são as tarefas que você executa repetidamente durante o seu dia de trabalho, semana ou mês? Crie skills para elas a fim de poupar tempo precioso, melhorar a qualidade do seu trabalho, ou ambos!

### Nível 3: Traços

Uma skill de traço (trait) afeta a personalidade e/ou o modo de operação do agente.

Uma skill famosa nessa categoria é a [`/grill-me`](https://github.com/mattpocock/skills), criada por Matt Pocock. Ficou tão popular que a maioria dos harnesses de agentes já inclui uma versão nativa pronta para uso. O propósito dessa skill é encorajar o agente a extrair conhecimento do usuário, preenchendo lacunas que, de outra forma, seriam adivinhadas ou, pior ainda, alucinadas.

Algumas das minhas skills favoritas baseadas em traços são: [swarm-coding](https://skills.danicat.dev/agents/swarm-coding), [double-diamond](https://skills.danicat.dev/agents/double-diamond) e [uno-reverse](https://skills.danicat.dev/agents/uno-reverse).

Swarm coding foi a primeira skill que escrevi para lidar com a orquestração de subagents. Escrevi sobre todo o processo de criação dela em [A Ascensão dos Subagentes]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}), então recomendo conferir esse artigo se você tiver curiosidade a respeito.

Double diamond é a sucessora espiritual da swarm coding, mas inclui uma fase de concepção (inception) e de descoberta (discovery). Quem está familiarizado com metodologias ágeis reconhecerá essa técnica. Ela se baseia no princípio de fases sucessivas de divergência e convergência. Você diverge para explorar o espaço do problema e depois converge para uma solução. Por exemplo, ao planejar uma grande feature neste site, posso iniciar toda a exploração com:

```text
/double-diamond I would like to plan phase 3 of this website. I want to improve the visualization of skill cards, add metrics, and a like/star button so we can rank skills by popularity.
```

Uno reverse é a minha própria abordagem para ter um agente com uma personalidade adversarial, com o único objetivo de quebrar a câmara de eco dos agentes que estão implementando o código. O agente adversarial criticará a implementação e buscará oportunidades para cortar excessos, funcionando como um contraponto contra o overengineering. Eu frequentemente o utilizo como uma checagem de sanidade durante revisões de arquitetura:

```text
Please run a /double-diamond process to refine ADR-0002. At the review phase, use an /uno-reverse agent to antagonise the proposal and present the results.
```

Como você pode ver, essas skills já são muito poderosas sozinhas, mas ficam ainda mais interessantes quando combinadas entre si. Por exemplo, eu me divirto muito usando `/grill-me` + `/double-diamond` + `/uno-reverse` no mesmo prompt. Isso faz o modelo me fazer perguntas esclarecedoras, partir para uma fase de pesquisa para encontrar diferentes implementações e, em seguida, criticá-las para chegar à solução ideal.

Quanto à inspiração para criar essas skills, procure momentos da sua carreira em que você aprendeu uma habilidade (humana) que ajudou você ou seu time a melhorar a qualidade do trabalho. Especialmente desde o advento dos subagents, estamos vendo cada vez mais o uso de habilidades de gestão sendo aplicadas a IAs, e isso não é coincidência. A IA simplesmente se mostra tão pouco confiável quanto nós quando deixada por conta própria.

## O ciclo de melhoria contínua

Criar skills é ótimo, mas você só vai realmente desbloquear todo o valor das skills se estiver melhorando-as continuamente. Meu processo é dar um tapa no visual e no conteúdo (um glow-up) das minhas skills quase toda vez que as uso, porque frequentemente descubro arestas, lacunas e bugs quando as utilizo em uma nova sessão. Outra desculpa para atualizá-las é quando um novo modelo é lançado, já que é esperado que o novo modelo seja mais capaz e você queira atualizar a skill para aproveitar o melhor de seus recursos.

Faço isso com tanta frequência que criei a skill [`/skill-optimizer`](https://skills.danicat.dev/agents/skill-optimizer). Essa skill é baseada nas melhores práticas publicadas no [agentskills.io](https://agentskills.io), com um toque do meu próprio gosto. Ela garante que as skills tenham um bom modelo de progressive disclosure, usando um script de apoio para estimar a contagem de tokens e manter cada seção em um tamanho razoável. Ela também otimiza a descrição da skill focando em benefícios em vez de detalhes de implementação. Por exemplo, antes da otimização, minha skill [search-analytics](https://skills.danicat.dev/analytics/search-analytics) estava cheia de enrolação sobre como a busca funcionava usando o algoritmo X ou Y, mas isso importa muito pouco para o uso da skill e era um desperdício de espaço. A informação relevante nesse caso é como carregar e consultar os dados. Todo o resto é desperdício de tokens.

Além de criar skills e otimizá-las por conta própria, também existe uma maneira automatizada de fazer isso: um conceito chamado **agent dreaming**, o processo de extrair lições de sessões anteriores dos agentes. No Antigravity, você pode pedir para ele "sonhar" usando o comando de barra `/learn`. Quando você inicia uma sessão de aprendizado, o agente explora o histórico da conversa em busca de padrões e procura maneiras de se aprimorar, oferecendo opções para atualizar arquivos de contexto, adicionar ou calibrar guardrails de configuração, ou gerar skills customizadas. Por exemplo, após uma longa maratona de debug, você poderia simplesmente rodar:

```text
/learn look at our debugging sessions with Ebitengine matrix transforms and state management. Extract the recurring bugs and fixes into a new documentation skill with clear anti-patterns and examples.
```

## Ativação: explícito é melhor do que implícito

Passei mais tempo do que gostaria de admitir tentando calibrar descrições para garantir que as skills fossem sempre ativadas quando o modelo precisasse delas. Falhei em todas as minhas tentativas até agora. Você consegue chegar a um ponto em que os modelos as ativam na maioria das vezes, mas, na minha experiência, taxas de ativação perfeitas ainda são inalcançáveis hoje.

É por isso que estou importando um velho adágio do Zen do Python e mudando a minha própria forma de trabalhar: explícito é melhor do que implícito. Em vez de esperar pela boa vontade do agente para invocar a skill que eu quero, eu digo exatamente o que quero invocado e quando.

No Antigravity também temos o benefício do mapeamento automático de skills para comandos de barra, então quando digito `/swarm-coding`, ele invoca imediatamente a skill de swarm coding. A notação de barra transforma skills em slash commands quando elas são a primeira palavra em um prompt, mas essa notação também funciona no meio de prompts como dicas para o modelo de que o que está prefixado com barra pode ser o nome de uma skill. Por exemplo, você pode adicionar `/grill-me` no final do seu prompt e o agente disparará a skill `grill-me` também, mesmo que tecnicamente não seja um slash command. Em versões recentes do Antigravity CLI, notei inclusive que o harness ocasionalmente adiciona `/grill-me` automaticamente ao final dos meus prompts quando meus comandos são relativamente complexos ou ambíguos.

## O problema de skills demais

Assim que você abraçar o mundo das skills, eventualmente chegará a esta situação. Por melhor que seja o modelo de progressive disclosure, você ainda sofrerá com o mesmo problema de inchaço de contexto (context bloat) que ocorre com servidores MCP se continuar instalando mais e mais skills.

Existem duas maneiras de melhorar isso: uma que exige trabalho ativo da sua parte e outra que emprega um "nível mais alto de abstração".

A primeira é simples: não instale skills globalmente, a menos que seja absolutamente necessário. As skills podem ser instaladas para o seu usuário (elas ficam na sua pasta home, geralmente em `~/.agents/skills`) ou em um workspace (nível de projeto, em `<project-dir>/.agents/skills`).

Essa solução funciona em pequena e média escala, mas começa a trazer problemas se você tem muitos projetos com muitas skills neles. Como mantê-las todas atualizadas? Como garantir que uma melhoria em uma skill seja refletida em todos os outros projetos que usam essa skill? Além disso, a pasta `.agents` deve ser commitada no repositório ou não?

Acho que a solução ideal para esse problema é uma combinação de centralização (ou federação) e rastreamento de proveniência. A primeira parte da solução é um catálogo de skills. Enquanto o compartilhamento típico de skills depende de git clones ou copiar e colar pastas, o padrão oficial de Agent Skills deixa a descoberta remota e a distribuição abertas para as ferramentas resolverem. Para solucionar isso, criei um catálogo manifesto online para `skills.danicat.dev` que expõe um índice pela web. Você pode inspecionar o catálogo manifesto em produção aqui: [https://skills.danicat.dev/catalog.json](https://skills.danicat.dev/catalog.json).

Um catálogo de skills resolve o problema da descoberta de skills. A ideia é que você pode apontar seus agentes para o catálogo e eles saberão imediatamente quais skills estão disponíveis sem precisar instalá-las. Isso não resolve o problema do context bloat, no entanto, a menos que você tenha um procedimento adequado de "buscar skill" que evite carregar o catálogo inteiro toda vez que estiver procurando uma nova skill.

O catálogo resolve uma das direções do fluxo de informação — do catálogo para a skill —, mas manter as skills atualizadas exige a direção oposta — da skill para o catálogo. É a isso que tenho chamado de "proveniência da skill": a fonte da skill. A especificação de Agent Skills não é prescritiva sobre proveniência, mas existem duas maneiras de gerenciar isso: 1) um gerenciador externo de skills que rastreia a origem de cada skill instalada, ou 2) adicionar informações do catálogo aos metadados da skill. Acredito que a segunda opção é mais elegante por ser "declarativa": não é preciso adivinhar qual mágica acontece nos bastidores; o que você vê é o que você tem. Este é um exemplo típico de frontmatter no meu catálogo de skills:

```yaml
---
name: swarm-coding
description: >
  Orchestrates multi-agent hierarchical swarms using a divide-and-conquer architecture for complex, multi-system, or orthogonal engineering initiatives (e.g., concurrent backend, frontend, database, QA). Manages hierarchical Lead Agents and Specialists, disjoint work allocations, and strict parent-child communication. Activate whenever the user mentions 'swarm', requests multi-agent team coordination, or needs context isolation across multiple technical domains.
license: Apache-2.0
metadata:
  category: agents
  tags: "swarm, subagents, parallel, orchestration, strategy, complexity, coordination"
  author: Daniela Petruzalek (daniela@danicat.dev)
  version: "0.2.0"
  catalog: https://skills.danicat.dev
---
```

`name`, `description`, `license` e `metadata` são os campos padrão definidos pela especificação. Embora `metadata` tenha uma definição livre, decidi seguir a recomendação de `version` e `author` como campos padrão. Minhas próprias extensões são `category`, `tags` e `catalog`. Category e tags me ajudam na descoberta e classificação, enquanto com `catalog` consigo rastrear a proveniência. Isso torna particularmente simples implementar um gerenciador de skills que mantém minhas skills atualizadas: ele só precisa acessar o catálogo no endereço fornecido e verificar se a versão da skill foi atualizada ou não.

Se você assumir que os agentes só navegarão pelo catálogo eventualmente (em vez de em toda sessão) para escolher a dedo e instalar skills, poderíamos relevar o context bloat de carregar o catálogo inteiro. Mas, na minha visão, o cenário ideal permitiria uma dinâmica um pouco mais inteligente. É aí que entra o `kungfu`.

## "I know kung fu."

[`kungfu`](https://github.com/danicat/kungfu) é a ferramenta de CLI que desenvolvi para me ajudar a gerenciar minhas skills. O nome vem da clássica cena de *Matrix* onde Neo tem Kung Fu (a arte marcial) "instalado" em seu cérebro. Esse era o sonho que eu tinha para os meus agentes: instalar conhecimento sob demanda a qualquer momento.

Eu tenho uma skill `kungfu` que ensina o agente a usar a CLI do `kungfu`, e essa é a única skill de que preciso. Se o agente precisa de uma nova skill, ele pode usar `kungfu find` para pesquisar no catálogo com base em nome, categoria, tags ou descrição, com tolerância para erros de digitação usando a distância de Levenshtein e sugestões no estilo "você quis dizer?". O comando `find` retornará uma lista dos melhores resultados ranqueados por pontuação, e se o agente precisar de alguma das skills listadas, ele pode carregá-las just-in-time (JIT) via `kungfu load` ou aprendê-las permanentemente via `kungfu learn` (instalando globalmente ou no workspace). A imagem abaixo mostra o fluxo de carregamento JIT:

![Terminal screenshot showing kungfu JIT loading the seo-optimizer skill](image.png "kungfu JIT loading the seo-optimizer skill")

O `kungfu` rastreará todas as skills que instalar e você pode verificar as skills disponíveis (instaladas ou online) via `kungfu list`. O comando `kungfu update` se encarrega de atualizar as skills que têm versões mais novas no servidor, fornecendo alguns guardrails para não sobrescrever customizações locais.

Como o `kungfu` depende dessa convenção do `catalog.json` e das minhas extensões de metadados no frontmatter, ele requer que os registros exponham um manifesto de catálogo ou que as skills declarem sua origem. Para fazê-lo funcionar sem atritos com repositórios de skills arbitrários por aí, adicionei rastreamento de estado local em `~/.config/kungfu/state.json` para registrar a origem das skills independentemente de como foram instaladas.

Você pode estar pensando: por que se dar ao trabalho de adicionar o catálogo aos metadados se você já tem o manifesto? Primeiro, porque gosto de coisas declarativas; segundo, porque isso permite reconciliação e gerenciamento de ciclo de vida mesmo que você não tenha usado o `kungfu` para aprender uma skill.

## Conclusão

As Agent Skills são uma das maneiras mais eficazes de conectar modelos fundacionais estáticos aos seus fluxos reais de engenharia no dia a dia. Seja começando pela criação de uma skill simples de documentação para banir padrões desatualizados de SDKs, automatizando tarefas analíticas repetitivas com skills de processo ou explorando swarms de agentes baseados em traços, o investimento se paga rapidamente.

Sinta-se à vontade para se inspirar nas ideias em [skills.danicat.dev](https://skills.danicat.dev) ou usar o [`kungfu`](https://github.com/danicat/kungfu) para gerenciá-las e carregá-las sob demanda. Seja qual for o caminho escolhido, o mais importante é continuar refinando seus processos para alcançar todo o potencial do agentic coding.
