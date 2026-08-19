---categories:
- Agentic Coding
date: 2026-07-22 00:00:00+00:00
heroStyle: big
summary: Uma exploração sobre o paradigma de subagentes no Antigravity, desde sua evolução
  e capacidades até aplicações práticas. Compartilho também a skill 'swarm-coding'
  para ajudar você a orquestrar subagentes especializados em tarefas complexas de engenharia.
tags:
  - agile
  - antigravity
  - subagents
title: "A Ascensão dos Subagentes"
slug: "the-rise-of-the-subagents"
aliases:
  - "/pt-br/posts/20260722-the-rise-of-the-subagents/"
description: "Entenda a evolução dos subagentes no Antigravity e aprenda a orquestrar enxames autônomos de código em Go com a skill Swarm Coding."
proficiencyLevel: "Intermediate"
dependencies:
  - "Google Antigravity 2.0 / agy CLI"
  - "Go 1.24+"
---

Devo confessar que, quando li pela primeira vez sobre **subagentes**, fiquei um pouco cética. Eu entendia perfeitamente os benefícios de rodar tarefas em janelas de contexto separadas, mas nunca tinha me ocorrido disparar dezenas ou centenas de agentes em paralelo. Ou, para ser mais precisa: eu não via vantagem prática nisso.

Gerenciar dois ou três agentes de código em segundo plano já consome uma fatia enorme da minha capacidade mental. Costumo paralelizar tarefas apenas quando sei que um agente está ocupado com uma operação demorada. Se coordenar três agentes assim já é exaustivo, como eu poderia sonhar em gerenciar centenas deles?

Levei algum tempo para encontrar a resposta, e ela é surpreendentemente simples: **você não gerencia**! Você delega a responsabilidade de coordenar os subagentes a um próprio agente. Todo problema em computação é resolvido adicionando uma nova camada de abstração, certo? Desta vez não é diferente.

Neste artigo, quero compartilhar como acompanhei a evolução do paradigma de subagentes ao longo dos últimos doze meses, consolidando essa experiência em uma skill que chamo de **Swarm Coding**. Se você está aqui pelo resumo direto (TL;DR), pode pular a seção histórica e ir direto para a definição da skill e suas explicações no final do texto.

## Uma breve (e incompleta) linha do tempo da evolução dos subagentes

Subagentes não são novidade. Eu já usava esse recurso muito antes de o termo "subagente" virar moda, encapsulando chamadas a modelos como ferramentas MCP. Por exemplo, as primeiras versões do [**GoDoctor**](https://github.com/danicat/godoctor) incluíam uma ferramenta `code_review`, que nada mais era do que uma chamada ao Gemini com um prompt especializado em revisão de código. Essa ferramenta funcionava, na prática, como um subagente — embora com comportamento fixo e sem suporte a conversas continuadas. (Tecnicamente era possível estender, mas preferi não implementar para garantir uma avaliação limpa e imparcial a cada chamada.)

Por volta do inverno passado, os principais agentes de código do mercado (Claude, Gemini CLI, etc.) começaram a introduzir suporte a subagentes personalizados definidos em arquivos Markdown. Gostei muito desse padrão como uma forma prática de empacotar conhecimento especializado com um conjunto refinado de ferramentas. No cenário ideal, o GoDoctor seria um agente especialista completo e não apenas uma coleção de ferramentas, mas acabei não seguindo por esse caminho na época porque o ecossistema mudava constantemente e o padrão de subagentes ainda não havia se estabilizado.

Avançando alguns meses: em maio de 2026, o Antigravity 2.0 trouxe suporte a subagentes, mas com uma limitação importante: os subagentes eram instanciados dinamicamente via ferramenta `DefineSubagent`. De início, a `DefineSubagent` dava pouca flexibilidade: ela apenas clonava o agente padrão ativo com um novo prompt. Ganhávamos a vantagem do contexto limpo, mas perdíamos o reúso estruturado de agentes. Não fiquei satisfeita, pois isso barrava a evolução do GoDoctor da forma como eu havia planejado.

Sem a possibilidade de definir agentes personalizados com modelos e ferramentas diferentes do agente principal, preferi deixar os subagentes de lado por um tempo e foquei em portar as automações que funcionavam bem no Gemini CLI para o Antigravity CLI, com relativo sucesso.

Minha visão sobre subagentes só mudaria depois de ler este prompt, publicado por [Richard Seroter](https://seroter.com/2026/06/01/one-prompt-four-subagents-and-ninety-seconds-to-get-a-working-app/) em junho:

> Let's build a hotel room booking app for Seroter Hotels consisting of a Go backend API and a web frontend. 
> 
> First, launch the **Engineering Manager** agent to design the API and frontend, saving the design and a Mermaid diagram into an artifact called 'architecture.md'. 
> 
> Once the design is ready, launch three agents in parallel:
> 1. **Test Manager**: Write a simple API test plan and append it to 'architecture.md'.
> 2. **Backend Engineer**: Build a clean Go REST API with standard error handling based on the design.
> 3. **Frontend Engineer**: Build a responsive web UI using a simple CSS framework like Tailwind to interact with the API (skip UI testing).
> 
> As soon as the Test Manager finishes the plan, have them hand it off to the Backend Engineer, who reads the plan from 'architecture.md' and adds the Go tests to the code. After both engineers finish building, the Test Manager runs the tests. Finally, spin up both components and a browser so I can test the live app.

Esse prompt trazia propostas muito interessantes e me fez revisitar o padrão. Ainda assim, eu continuava preocupada com dois pontos: primeiro, o quanto eu precisaria adaptar meu estilo de prompting para pensar em termos de subagentes; segundo, por que eu me daria ao trabalho de escrever prompts dessa forma?

Sou muito pragmática: se não enxergo um ganho claro em qualidade e/ou velocidade, prefiro não gastar energia extra. Pensar em subagentes é muito parecido com raciocinar sobre concorrência na programação tradicional: a primeira pergunta é "isso é realmente paralelizável?"; a segunda é "o ganho compensa o esforço?", já que o custo de coordenação frequentemente anula pequenos ganhos.

No prompt do Richard, as únicas frentes verdadeiramente ortogonais são o desenvolvimento do backend e do frontend: ambos trabalham de forma independente desde que haja um contrato bem estabelecido. Mas todos os outros agentes dependem uns dos outros, operando em um fluxo muito mais sequencial do que paralelo.

O benefício real, portanto, vinha quase que exclusivamente do isolamento de contexto, e não de ganhos de velocidade por paralelismo — algo difícil de mensurar nessa escala.

Passei as duas semanas seguintes com essa questão maturando na cabeça: "Quais papéis são de fato ortogonais para que eu possa aproveitar o poder dos subagentes?"

Foi só após uma série de conversas inspiradoras no GDE Summit em Berlim que a ficha finalmente caiu: não cabe a **você** definir cada subagente no prompt; o segredo é **ensinar** o próprio agente a decidir quando instanciar seus subagentes. Em essência, eu estava pensando como uma líder de engenharia dividindo demandas para o time, quando o que precisava fazer era transformar o próprio agente na liderança técnica da operação.

## O nascimento do swarm coding

O ato de decompor projetos complexos em tarefas menores e distribuí-las entre membros da equipe é algo natural para mim. Antes de migrar para Developer Relations, atuei como Tech Lead e depois como Principal Engineer. Essas atividades são o arroz com feijão da liderança técnica, especialmente para quem tem bagagem em métodos ágeis.

A mesma lógica de TL se aplica à criação de um enxame (*swarm*) de subagentes: você precisa garantir que cada agente receba uma tarefa autocontida para trabalhar com total autonomia. Para que a tarefa possa ser executada, ela precisa de especificações claras (o bom e velho *Definition of Ready*) e critérios de aceitação objetivos (*Definition of Done*).

Entre parênteses: quase ninguém descreve essa parte do trabalho como a sua favorita (eu inclusa), o que explica minha resistência inicial em adotar um estilo de prompting que parecia uma liderança técnica elevada à máxima potência.

Então, em vez de atuar como TL dos meus agentes, inverti o jogo: ensinei meu agente principal a assumir o papel de TL e montar sua própria equipe para concretizar minha visão. Foi assim que nasceu a [primeira versão da skill swarm coding](https://github.com/danicat/skills/blob/a9f57b10127d8bd23ed4867d64d168063a3726f4/swarm_coding/SKILL.md). Veja um trecho dos pontos principais:

> Swarm Coding is a new development paradigm that employs multiple sub-agents in parallel to work on complex tasks. It is based on the divide to conquer strategy. The main benefits of this strategy is context isolation and quality improvement: by assigning small self contained tasks to sub-agents, you avoid context dilution and enable very focused refinement of the solution. For example, without swarm coding an agent implementing both frontend and backend will often get distracted as the skills required for frontend and backend are often unrelated (different technology stack, different best practices, etc.)
> 
> ## ROLE
> 
> You are the SWARM COORDINATOR, your role is to break down complex tasks and DELEGATE to sub-agents for execution. You should NEVER execute tasks on your own, no matter how simple they seem to be UNLESS it is EXPLICITLY requested by the user or your parent coordinator. ALWAYS keep the communication channel open so the user or parent agent can send you steering commands.
> 
> ## AGENT BUDGET
> 
> It is the number of sub-agents you are allowed to spawn to work on a task. You are encouraged to use the FULL BUDGET of agents, or get as close to it as possible. This doesn't mean to waste resources on low value tasks, but in finding the optimal use of the BUDGET to achieve the best quality output.
> 
> ## TEAM BUILDING
> 
> For SIMPLE tasks, break down the task into orthogonal elements and assign one or more SPECIALIST agents for each element.
> For COMPLEX tasks, break down the task into smaller pieces and assign LEAD AGENTS to each of them. The LEAD AGENTS should have a fraction of the agent budget to execute the task. LEAD AGENTS should activate the swarm coding skill and become the SWARM COORDINATOR for their respective areas
> Proceed recursively until you have a complete tree of LEAD AGENTS and EXECUTOR agents.
> 
> ## COMMUNICATION
> 
> The SWARM COORDINATOR is responsible for communicating directly with its sub-agents. Sub-agents should not message each other,communication between agents at the same level should be made by DESIGN DOCUMENTS. It is the SWARM COORDINATOR responsibility to make sure all changes to design documents are broadcast to the agents in their squad. Upon conflict, the SWARM COORDINATOR is responsible for disambiguating and making a decision.
> 
> ## PLANNING
> 
> Planning is a FIRST CLASS effort and should also be made using the SWARM. Each AGENT should contribute to the plan with their expertise. It is the role of the SWARM COORDINATOR for a squad to revise the part of the plan produced by their team and address inconsistencies or make decisions when there is conflict.
> 
> ## EXECUTION
> 
> In execution phase, monitor the progress of the swarm accross the main milestones, and steer agents if necessary to keep them aligned with the end goal. Remember that as coordinator you are ONLY allowed to handle ARTIFACTS. All development tasks should be handled by leaf sub-agents.

Essa foi a primeira vez em que escrevi uma skill 100% à mão, já que era muito difícil transmitir essa visão de outra forma. Aquele prompt inicial era ambicioso demais: eu queria que o enxame fosse "recursivo", onde o agente decidiria se deveria atuar como coordenador ou executor com base puramente no orçamento de agentes. Na prática, não funcionou tão bem.

O que acontecia era que a tarefa repassada pelo coordenador assumia prioridade absoluta sobre qualquer outra instrução, fazendo com que o subagente pulasse direto para o modo de execução sem respeitar a distribuição de agentes. Corrigi isso na versão atual da skill, fornecendo diretrizes mais explícitas e modelos de prompt específicos para instanciar cada subagente.

## Colocando o enxame para rodar

Você pode encontrar a versão atual da skill **swarm coding** no [repositório de skills da Danicat no GitHub](https://github.com/danicat/skills). Para instalá-la no seu agente de preferência, use o comando abaixo:

```bash
$ npx skills add github.com/danicat/skills --skill swarm-coding
```

> **Nota**: Esta skill é um trabalho em constante evolução; sinta-se à vontade para criar um fork se quiser adaptá-la para um ambiente específico.

Aqui está um prompt divertido para começar. Experimente rodá-lo no Antigravity CLI:

> /swarm-coding agent budget 50. Develop a 2D tower defense survival game using Go and Ebitengine. The game should be feature complete and have one single screen level. Include an intro sequence, title screen, game win and game over screens as well. Track the high score at the end of each playthrough. Use 32x32 sprites with up to 256 colors each. The sprites should be custom designed for this game and each movement should have at least 3 frames of animation, but ideally 8. Tiles should be 32x32 as well. The level view is top down, movement is on four directions. The player should have access to 4 types of units and 4 types of buildings. The enemy waves should have 8 types of monsters, including one boss monster. Use typical build and attack phases with custom UIs for each. To create art, use vector graphics and/or dot (pixel) art creating each asset manually using binary data. Sound effects should be generated mathematically as well. The whole vibe of the game should match the 16-bit era, but with modern gameplay features.

Confira o resultado na minha máquina:

![Swarm Defense](image-1.png "Captura de tela do jogo criado pelo swarm")

Não posso dizer que foi de primeira (*one-shot*), pois a primeira compilação tinha um bug na renderização dos sprites que deixava a tela toda preta. Mas bastou um único prompt reportando o problema para que o jogo funcionasse como na imagem acima.

Aqui está um pequeno vídeo dele em ação na batalha final contra o chefe (o coitado não teve a menor chance):

<video controls src="swarm-defense.mp4" title="Trecho da batalha contra o boss em Swarm Defense"></video>

Todos os recursos visuais e sonoros deste vídeo foram gerados programaticamente. Ou seja: o Antigravity não tinha acesso a modelos de geração de imagens, precisando criar cada sprite diretamente no nível de bitmap.

Essa abordagem funcionou tão bem justamente porque o enxame permitiu aos agentes especialização e foco exclusivo em tarefas isoladas. Já tentei prompts semelhantes com um único agente e o resultado invariavelmente ficava aquém do esperado. Quando você sobrecarrega um agente com tarefas ortogonais demais, ele acaba fazendo de tudo um pouco sem dominar nada. Com a delegação estruturada, cada subagente cuida de uma fatia autocontida do projeto e entrega o seu melhor.

## Suporte a subagentes no Antigravity 2.0 e Antigravity CLI

No momento em que escrevo este artigo, as funcionalidades de subagentes estão distribuídas de forma ligeiramente diferente entre o Antigravity 2.0 e o Antigravity CLI. Como essas duas interfaces atendem a fluxos de trabalho distintos, seus recursos divergiram temporariamente. Com a evolução rápida de ambas as ferramentas, podemos esperar que essa diferença diminua à medida que o ecossistema amadureça.

Em essência, os dois ambientes compartilham o mesmo motor. Criar um subagente delega a tarefa e devolve imediatamente o controle a você. O subagente inicia com uma folha em branco: ele utiliza o mesmo modelo da sessão principal, mas parte de um contexto totalmente isolado, impedindo vazamentos do histórico da conversa. O agente pai se comunica com ele por meio de identificadores únicos. Caso o subagente encontre um comando que exija permissão, essa solicitação é repassada diretamente para você.

As principais diferenças entre as duas interfaces são:
- No **Antigravity 2.0**, a gestão é visual. Você acompanha tarefas ativas, visualiza logs de conversa e interrompe execuções diretamente pela barra lateral. Agentes customizados são instanciados sob demanda por meio da ferramenta `DefineSubagent`. Não há suporte a plugins para subagentes nesta interface.
- No **Antigravity CLI**, além de criar agentes dinamicamente, você pode declará-los estaticamente em arquivos Markdown, usando opções de frontmatter para fixar modelos específicos ou filtrar as ferramentas disponíveis. A CLI também oferece suporte ao carregamento de subagentes personalizados definidos em plugins no formato Markdown.

Compreender essas particularidades é fundamental para organizar seu fluxo de trabalho hoje, embora a tendência natural seja a convergência de recursos entre as duas interfaces ao longo do tempo.

## Experimente na prática

A melhor maneira de sentir o impacto dos subagentes é testando por conta própria. Seja reproduzindo meu prompt de teste ou criando um novo desafio, tenho certeza de que você vai se surpreender com o resultado. Compartilhe comigo o que você conseguir criar com o enxame! Enquanto isso, estarei por aqui refinando o [Swarm Defense](https://github.com/danicat/swarm-defense) mais um pouco. :)

- Confira o *swarm coding* e todas as minhas outras skills no [repositório danicat/skills no GitHub](https://github.com/danicat/skills)
- Baixe e conheça mais sobre o Antigravity no [site oficial do Antigravity](https://antigravity.google)
