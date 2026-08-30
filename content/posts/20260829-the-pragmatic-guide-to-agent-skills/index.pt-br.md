---
categories:
  - Agentic Coding
date: 2026-08-29
heroStyle: big
title: "O Guia Pragmático para Agent Skills"
slug: "the-pragmatic-guide-to-agent-skills"
aliases:
  - "/pt-br/posts/20260829-the-pragmatic-guide-to-agent-skills/"
summary: "Dominar a criação de Agent Skills personalizadas é o maior salto de produtividade para desenvolvedores. Explore casos práticos em documentação, fluxos e traços."
description: "Um guia pragmático para criar Agent Skills personalizadas: explore casos de uso práticos em documentação, automação de processos e traços de personalidade."
tags:
  - agent-skills
  - antigravity
  - gemini-cli
  - vibe-coding
proficiencyLevel: "Intermediate"
dependencies:
  - "Antigravity 2.0"
  - "Antigravity CLI"
  - "Gemini CLI"
  - "Claude Code"
---

Pode parecer difícil de acreditar, mas o padrão [**Agent Skills**](https://agentskills.io) não tem sequer um ano de vida. O [artigo original da Anthropic](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills) foi publicado em outubro de 2025, apresentando inicialmente as Agent Skills como uma extensão do Claude Code antes de se consolidar como um padrão aberto em dezembro de 2025. A especificação provou ser tão prática que, em pouco tempo, foi adotada pelos principais agentes de código e frameworks de desenvolvimento de agentes para equipar modelos de IA com instruções modulares, scripts determinísticos e fluxos especializados de domínio.

Avançando para os dias de hoje, a maioria dos desenvolvedores já conhece as skills e as vantagens do modelo de *progressive disclosure* (revelação progressiva), mas será que estamos usando e criando skills de forma realmente eficaz? Neste artigo, vamos explorar os principais desafios ao lidar com skills, desde a descoberta e proveniência até o desenvolvimento e otimização.

Todas as skills e exemplos mencionados neste texto foram testados usando o Antigravity CLI e o Gemini 3.7 Flash, mas incentivo você a experimentá-los mesmo que utilize outro harness ou modelo. Sem mais delongas, vamos mergulhar no universo das Agent Skills!

## Criar skills vs consumir skills prontas

O eterno dilema da engenharia de software: devo construir a minha própria solução ou adotar algo pronto? Existem várias abordagens para esse problema, mas vou compartilhar minha visão pragmática:

```goat
      +-----------------------------------------+
      |  Este processo é específico do seu repo |
      |              ou codebase?               |
      +-----------------------------------------+
            |                             |
       Sim  |                             |  Não
            |                             v
            |                 +-----------------------+
            |                 | Conhece uma skill que |<-------------------+
            |                 |    faça esse papel?   |                    |
            |                 +-----------------------+                    |
            |                    |                 |                       |
            |               Sim  |                 |  Não                  |
            |                    v                 v                       |
            |         +--------------------+  +------------------+         |
            |         | Passa em checagens |  | Gastou tempo     |         |
            |         | de segurança/integr|  | demais buscando? |         |
            |         +--------------------+  +------------------+         |
            |            |              |        |            |            |
            |       Sim  |          Não |    Sim |        Não |            |
            |            |              |        |            v            |
            |            |              |        |    +---------------+    |
            |            |              |        |    |  Procurar por |    |
            |            |              |        |    |   uma skill   |    |
            |            |              |        |    +---------------+    |
            |            |              |        |            |            |
            |            |              v        v            +------------+
            |            |         +------------------+
            |            |         |  Construa a sua  |
            +--------------------->|  própria skill   |
                         |         +------------------+
                         |                   |
                         v                   v
                  +-----------------------------------+
                  |             Sucesso!              |
                  +-----------------------------------+
```

Admito que o diagrama é um pouco caótico, mas a regra de ouro é simples: a menos que você tenha uma fonte autoritativa para a skill, ou confiança absoluta de que não está se expondo a um vetor de ataque, é melhor construir a sua própria.

O esforço para criar uma skill é muito baixo, então há pouquíssimas circunstâncias em que eu desaconselharia essa criação. A principal delas seria a falta de conhecimento no domínio para avaliar a qualidade técnica da skill, já que isso cria um grande risco de alimentar uma câmara de eco de maus hábitos.

Se você domina o domínio, construir skills é o maior multiplicador de produtividade ao trabalhar com agentes de código — e é exatamente por isso que acredito que todo desenvolvedor sério deveria dominar a habilidade de criar skills (com o perdão do trocadilho).

## Dominando a habilidade de criar skills

O segredo para criar skills úteis reside na observação atenta dos padrões do agente. É provável que você não consiga conceber uma skill logo na primeira sessão de código, mas quanto mais usar agentes e observar seu comportamento, mais notará lacunas de conhecimento e de execução que podem ser preenchidas com skills customizadas.

Isso ficará mais claro com alguns exemplos práticos, mas antes vamos analisar os diferentes níveis e tipos de skills que podemos construir:

### Nível 1: Documentação

Uma skill que explica uma tecnologia específica, ampliando ou substituindo o conhecimento prévio do modelo. É extremamente comum, por exemplo, que modelos recomendem bibliotecas obsoletas ou padrões ultrapassados simplesmente porque eram prevalentes nos seus dados de treinamento. Usar skills para atualizar o conhecimento do modelo sobre um ecossistema é um dos usos mais fundamentais da tecnologia.

O formato típico de uma skill de documentação é um único arquivo `SKILL.md` com as diretrizes atualizadas, incluindo linhas de comando, snippets de código e referências externas. Dependendo da complexidade do assunto, dividir o `SKILL.md` em arquivos separados de referência e assets é a melhor estratégia para maximizar o potencial do *progressive disclosure*.

Crie esse tipo de skill quando notar seu agente usando padrões depreciados, versões antigas de SDKs ou quando ele gerar repetidamente implementações longe do ideal. A maioria das skills fornecidas oficialmente por empresas se encaixa nessa categoria, como o repositório [Google Skills](https://github.com/google/skills), que reúne guias para a maioria dos produtos do Google.

Um exemplo de skill de documentação que criei é a [ebitengineer](https://skills.danicat.dev/game-dev/ebitengineer/). Essa skill nasceu após eu notar problemas recorrentes no desenvolvimento de jogos 2D com Ebitengine — como a ordem incorreta de operações de matriz, modularização deficiente, ausência de gerenciamento de estado e o abuso de texto de depuração em vez de fontes de produção.

### Nível 2: Processo

Uma skill que estabelece um processo ou fluxo de trabalho rigoroso, como realizar uma revisão de código, conduzir uma auditoria de segurança ou analisar métricas de desempenho. Esse tipo de skill costuma empacotar não apenas conhecimento teórico, mas scripts customizados ou ferramentas CLI prontas para executar tarefas, poupando o modelo de reconstruí-las do zero a cada sessão.

Você deve criar skills de processo sempre que se pegar instruindo o modelo a fazer a mesma sequência de tarefas repetidamente, especialmente quando houver passos determinísticos que podem ser automatizados via script. Algumas skills de processo que utilizo com frequência fazem parte do conjunto de [analytics](https://skills.danicat.dev/analytics), usadas para coletar e consolidar dados de redes sociais, Google Analytics e Google Search Console.

Otimizar minha presença online é uma parte importante do meu trabalho (de pouco adianta produzir conteúdo se ninguém o encontra), então projetei essas skills para automatizar o que antes eu fazia manualmente em cada painel. Hoje tenho uma verdadeira "central de comando" alimentada por essas skills, economizando um tempo valioso que posso direcionar para a produção de conteúdo técnico de maior qualidade.

Olhe para o seu próprio fluxo de trabalho: quais são as tarefas repetitivas que você executa todos os dias, semanas ou meses? Crie skills para elas a fim de poupar tempo, elevar a consistência técnica das entregas, ou ambos!

### Nível 3: Traços de Personalidade

Uma skill de traço (*trait*) altera a personalidade, postura crítica ou modo de raciocínio do agente.

Uma das skills mais famosas dessa categoria é a [`/grill-me`](https://github.com/mattpocock/skills), criada por Matt Pocock. Ficou tão popular que a maioria dos harnesses de agentes já inclui uma versão nativa dela. O objetivo dessa skill é forçar o agente a extrair informações e requisitos do próprio usuário por meio de perguntas direcionadas, preenchendo lacunas que, de outra forma, seriam adivinhadas ou alucinadas pelo modelo.

Algumas das minhas skills favoritas baseadas em traços são: [swarm-coding](https://skills.danicat.dev/agents/swarm-coding), [double-diamond](https://skills.danicat.dev/agents/double-diamond) e [uno-reverse](https://skills.danicat.dev/agents/uno-reverse).

A skill swarm coding foi a primeira que escrevi para orquestrar subagentes em paralelo. Detalhei todo o processo de criação dela em [A Ascensão dos Subagentes]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}), recomendo a leitura se você tiver curiosidade sobre a mecânica.

A double diamond é a sucessora espiritual da swarm coding, incorporando etapas estruturadas de descoberta (*discovery*) e concepção. Desenvolvedores familiarizados com métodos ágeis e *design thinking* reconhecerão a dinâmica de fases sucessivas de divergência e convergência: primeiro você diverge para explorar o espaço do problema, e depois converge para uma solução técnica sólida. Por exemplo, ao planejar uma grande evolução neste site, posso iniciar a sessão com:

```text
/double-diamond Gostaria de planejar a fase 3 deste website. Quero melhorar a visualização dos cards de skills, adicionar métricas e um botão de like/estrela para ranquear skills por popularidade.
```

A uno reverse é a minha abordagem para criar um agente deliberadamente adversário, rompendo a "câmara de eco" dos agentes responsáveis pela implementação. O agente adversário criticará a solução proposta e buscará ativamente redundâncias e desperdícios, funcionando como um contraponto eficaz contra o *overengineering*. Eu a utilizo com frequência em revisões de arquitetura:

```text
Por favor, execute um processo /double-diamond para refinar a ADR-0002. Na etapa de revisão técnica, utilize um agente /uno-reverse para contrapor a proposta e apresentar os pontos de atenção.
```

Essas skills já são muito poderosas isoladamente, mas ficam ainda mais impressionantes quando combinadas no mesmo fluxo. Adoro disparar `/grill-me` + `/double-diamond` + `/uno-reverse` no mesmo prompt: isso faz o modelo me sabatinar com perguntas de clarificação, entrar em uma fase de pesquisa profunda para explorar alternativas e, em seguida, criticar as opções para definir a arquitetura ideal.

Para encontrar inspiração para novas skills, resgate momentos da sua trajetória profissional em que você aprendeu uma técnica (humana) que elevou o patamar de engenharia do time. Especialmente após o surgimento dos subagentes, estamos vendo cada vez mais habilidades de liderança e facilitação técnica sendo aplicadas a IAs. Não é coincidência: modelos de IA, quando deixados por conta própria, enfrentam as mesmas armadilhas de coordenação que os humanos.

## O ciclo de melhoria contínua

Criar skills é ótimo, mas você só colherá todo o potencial dessa abordagem se mantiver um ciclo contínuo de refinamento. Meu hábito é dar uma polida nas minhas skills quase toda vez que as utilizo em um projeto real, pois é no uso diário que descubro arestas, lacunas de contexto ou pequenas falhas. Outra excelente oportunidade para revisá-las é no lançamento de novos modelos de ponta, permitindo atualizar as instruções para tirar proveito de raciocínios mais avançados.

Faço isso com tanta frequência que criei a skill [`/skill-optimizer`](https://skills.danicat.dev/agents/skill-optimizer). Ela sintetiza as boas práticas publicadas no [agentskills.io](https://agentskills.io) com alguns dos meus próprios padrões de qualidade. Ela assegura que as skills sigam um modelo consistente de revelação progressiva, usando um script auxiliar para estimar a contagem de tokens e manter cada seção enxuta. Ela também aprimora as descrições focando nos benefícios práticos em vez de minúcias internas. Antes dessa otimização, minha skill [search-analytics](https://skills.danicat.dev/analytics/search-analytics), por exemplo, continha parágrafos desnecessários explicando como a busca funcionava internamente — detalhes irrelevantes para o uso da skill que apenas consumiam tokens à toa. O que realmente importa é como carregar e consultar os dados.

Além da otimização manual, existe uma abordagem automatizada muito promissora: o conceito de **agent dreaming** (o processo de extrair aprendizados a partir do histórico de sessões anteriores). No Antigravity, você pode disparar esse processo com o comando `/learn`. Ao iniciar uma sessão de aprendizado, o agente analisa o histórico recente em busca de padrões e propõe melhorias — sugerindo ajustes em arquivos de contexto, calibrando guardrails de configuração ou gerando novas skills especializadas. Por exemplo, após uma sessão intensa de depuração, você pode simplesmente rodar:

```text
/learn analise nossas sessões de depuração com matrizes e gerenciamento de estado no Ebitengine. Extraia os bugs recorrentes e soluções em uma nova skill de documentação com anti-patterns e exemplos claros.
```

## Ativação: explícito é melhor do que implícito

Passei mais tempo do que gostaria de admitir tentando calibrar descrições para garantir que os modelos sempre ativassem as skills certas no momento exato. Falhei em todas as tentativas. Você até consegue fazer o modelo ativá-las na maior parte das vezes, mas uma taxa de ativação de 100% ainda é inalcançável hoje.

Por isso, adotei uma máxima clássica do *Zen of Python* no meu fluxo de trabalho: *explícito é melhor do que implícito*. Em vez de contar com a sorte de o agente deduzir qual skill deve ser chamada, eu indico exatamente o que deve ser executado e quando.

No Antigravity temos ainda a conveniência do mapeamento automático de skills para comandos de barra: digitar `/swarm-coding` ativa imediatamente a skill. Essa notação funciona tanto no início de um comando quanto no meio de um prompt, servindo como uma dica (*hint*) explícita para o modelo de que aquela palavra prefixada com barra representa uma skill. Por exemplo, você pode incluir `/grill-me` ao final de uma mensagem e o agente acionará o comportamento correspondente. Em versões recentes do Antigravity CLI, notei até que o próprio harness adiciona `/grill-me` ao final dos meus comandos quando percebe instruções mais complexas ou ambíguas.

## O desafio do excesso de skills

Assim que você adota o ecossistema de skills em larga escala, inevitavelmente esbarra em um problema: por melhor que seja o modelo de revelação progressiva, você ainda sofrerá com o inchaço de contexto (*context bloat*) se mantiver dezenas de skills instaladas permanentemente — o mesmo desafio comum aos servidores MCP.

Existem dois caminhos para contornar isso: um que exige disciplina manual e outro que introduz uma camada superior de abstração.

O primeiro é direto: evite instalar skills globalmente, a menos que sejam de uso universal. As skills podem ser instaladas no escopo do seu usuário (em `~/.agents/skills`) ou no escopo de um workspace específico (`<projeto>/.agents/skills`).

Essa separação atende bem a cenários de pequeno a médio porte, mas se torna difícil de gerenciar quando você tem dezenas de repositórios com várias skills locais. Como mantê-las atualizadas? Como garantir que uma melhoria em uma skill seja propagada para todos os projetos que a utilizam? E a pasta `.agents` deve ser versionada no Git ou ignorada?

A meu ver, a solução ideal combina centralização (ou federação) com o rastreamento rigoroso de proveniência. A primeira peça dessa engrenagem é um catálogo de skills. Enquanto a distribuição tradicional ainda depende de clonar repositórios Git ou copiar pastas manualmente, o padrão oficial de Agent Skills deixa a descoberta remota a cargo das ferramentas de suporte. Para resolver isso, criei um catálogo online em `skills.danicat.dev` que publica um manifesto de índice via web. Você pode conferir o manifesto em produção aqui: [https://skills.danicat.dev/catalog.json](https://skills.danicat.dev/catalog.json).

O catálogo resolve a descoberta: você aponta seus agentes para o catálogo e eles passam a conhecer as skills disponíveis sem precisar instalá-las previamente. No entanto, isso ainda não elimina o inchaço de contexto, a menos que você tenha um mecanismo de busca eficiente que evite injetar o catálogo inteiro a cada consulta.

Além disso, o catálogo resolve apenas a via de ida (do catálogo para a skill). Manter as skills atualizadas exige o caminho inverso (da skill de volta ao catálogo). É a isso que chamo de "proveniência da skill": a origem declarada daquele conhecimento. A especificação não impõe regras sobre proveniência, mas podemos gerenciá-la de duas formas: 1) por meio de um gerenciador externo que registra a origem de cada instalação, ou 2) adicionando as informações do catálogo diretamente nos metadados do frontmatter da skill. A segunda opção é muito mais elegante por ser puramente declarativa: não há mágica oculta, o que está escrito no arquivo é o que vale. Veja um exemplo real do meu catálogo:

```yaml
---
name: swarm-coding
description: >
  Orchestrates multi-agent hierarchical swarms using a divide-and-conquer
  architecture for complex, multi-system, or orthogonal engineering initiatives
  (e.g., concurrent backend, frontend, database, QA). Manages hierarchical Lead
  Agents and Specialists, disjoint work allocations, and strict parent-child
  communication. Activate whenever the user mentions 'swarm', requests
  multi-agent team coordination, or needs context isolation across multiple
  technical domains.
license: Apache-2.0
metadata:
  category: agents
  tags: "swarm, subagents, parallel, orchestration, strategy, complexity, coordination"
  author: Daniela Petruzalek (daniela@danicat.dev)
  version: "0.2.0"
  catalog: https://skills.danicat.dev
---
```

Os campos `name`, `description`, `license` e `metadata` são definidos pela especificação padrão. Embora `metadata` seja flexível, adotei `version` e `author` como campos padrão, estendendo-os com `category`, `tags` e `catalog`. Categorias e tags auxiliam na indexação e busca, enquanto `catalog` registra a proveniência exata. Isso torna a atualização trivial: o gerenciador de skills só precisa consultar a URL do catálogo e verificar se há uma versão mais nova disponível.

Se imaginarmos que os agentes consultarão o catálogo apenas esporadicamente para instalar uma skill pontual, o impacto de carregar o catálogo não é tão grave. Mas o cenário ideal permite uma dinâmica muito mais inteligente. É aí que entra o `kungfu`.

## "I know kung fu."

[`kungfu`](https://github.com/danicat/kungfu) é a ferramenta CLI que desenvolvi para gerenciar minhas skills. O nome é uma homenagem à clássica cena de *Matrix* em que Neo tem a arte marcial "instalada" diretamente em seu cérebro. Esse era exatamente o meu sonho para os agentes de código: poder injetar conhecimentos especializados sob demanda a qualquer instante.

Tenho uma skill chamada `kungfu` que ensina o próprio agente a usar a CLI do `kungfu`, sendo esta a única skill permanente que preciso manter. Quando o agente precisa de uma nova habilidade, ele usa `kungfu find` para pesquisar o catálogo por nome, categoria, tags ou descrição, contando com tolerância a erros de digitação via distância de Levenshtein e sugestões do tipo "você quis dizer...". O comando retorna as melhores opções ranqueadas por relevância, e o agente pode carregá-las sob demanda (*just-in-time* / JIT) via `kungfu load` ou instalá-las permanentemente via `kungfu learn` (globalmente ou no workspace). A imagem abaixo ilustra esse fluxo JIT:

![Terminal screenshot showing kungfu JIT loading the seo-optimizer skill](image.png "kungfu JIT loading the seo-optimizer skill")

O `kungfu` rastreia todas as skills instaladas, permitindo consultar o que está disponível (local ou remotamente) via `kungfu list`. O comando `kungfu update` cuida de atualizar as skills que receberam novas versões no servidor, com mecanismos de segurança para não sobrescrever customizações locais.

Como o `kungfu` se baseia na convenção do `catalog.json` e nas extensões de metadados no frontmatter, ele requer que os repositórios exponham um manifesto ou que as skills declarem sua origem. Para funcionar perfeitamente com repositórios arbitrários da comunidade, adicionei controle de estado local em `~/.config/kungfu/state.json`, registrando a proveniência de cada skill independentemente de como foi instalada.

Você pode se perguntar: por que adicionar o catálogo aos metadados se já existe o manifesto? Primeiro, porque prefiro abordagens declarativas; segundo, porque isso permite reconciliação e gerenciamento de ciclo de vida mesmo quando a skill foi instalada manualmente sem o `kungfu`.

## Conclusão

As Agent Skills representam uma das formas mais eficazes de conectar modelos fundacionais genéricos aos seus fluxos reais de engenharia do dia a dia. Seja criando uma skill simples de documentação para eliminar padrões ultrapassados de SDKs, automatizando tarefas analíticas repetitivas com skills de processo ou explorando traços comportamentais para orquestrar agentes, o retorno sobre o investimento é imediato.

Fique à vontade para se inspirar nos exemplos de [skills.danicat.dev](https://skills.danicat.dev) ou usar o [`kungfu`](https://github.com/danicat/kungfu) para carregar e gerenciar skills sob demanda. Independentemente do caminho escolhido, o mais importante é continuar refinando seus processos para extrair todo o potencial do desenvolvimento com agentes.
