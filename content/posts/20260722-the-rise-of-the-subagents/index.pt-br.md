---
title: "A Ascensão dos Subagentes"
date: 2026-07-22T00:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["agentic-coding", "agile", "antigravity", "subagents"]
summary: "Uma exploração do paradigma de subagents no Antigravity, desde sua evolução e capacidades até aplicações práticas. Também compartilhamos uma skill chamada 'swarm-coding' para ajudar você a orquestrar subagentes especializados para tarefas complexas de engenharia."
heroStyle: "big"
---
Devo confessar que quando li pela primeira vez sobre **subagents**, fiquei um pouco cético. Eu conseguia entender os benefícios de rodar tarefas em janelas de contexto separadas, mas nunca tinha me ocorrido iniciar dezenas ou talvez até centenas de agentes em paralelo. Ou talvez seja melhor dizer que eu não via o benefício de fazer isso.

Gerenciar alguns agentes de codificação em segundo plano já consome uma grande parte da minha largura de banda mental. Eu costumo fazer coisas em paralelo apenas quando sei que um agente está ocupado trabalhando em uma operação de longa execução. Se gerenciar dois ou três agentes dessa forma já é doloroso, como posso sonhar em gerenciar centenas deles?

Levei muito tempo para descobrir, mas a resposta é, na verdade, **você não gerencia**! Você delega a responsabilidade de gerenciar os subagents a um próprio agente. Todo problema em computação é resolvido por um novo nível de abstração, certo? Não é diferente desta vez.

Neste artigo, quero guiar você por como vi a evolução do paradigma de subagents nos últimos 12 meses, consolidando depois minha experiência em uma skill de agente que chamo de **Swarm Coding**. Se você está aqui pelo TL;DR, pode pular a seção abaixo e ir direto para a definição e explicação da skill no final do texto.

## A brief (and incomplete) timeline of subagent evolution

Subagents não são novidade. Eu venho usando este truque muito antes do termo subagent ser cunhado, empacotando chamadas de modelo como MCP tools. Por exemplo, as primeiras versões do [**GoDoctor**](https://github.com/danicat/godoctor) tinham uma ferramenta `code_review` que era uma chamada de modelo para o Gemini com um prompt de revisão de código personalizado. Essa ferramenta era efetivamente um subagent, embora com comportamento codificado e sem oportunidade de continuar a conversa. (Tecnicamente era possível, eu apenas nunca implementei pois queria uma revisão imparcial em cada chamada.)

Por volta do inverno passado, todos os suspeitos habituais de agentes de codificação (Claude, Gemini CLI, etc.) começaram a adicionar suporte a subagents personalizados definidos em arquivos Markdown. Gostei muito desse padrão como uma forma de empacotar conhecimento especializado com um conjunto de ferramentas curado. No mundo ideal, o GoDoctor seria um agente especialista e não apenas um conjunto de ferramentas, mas acabei nunca implementando dessa forma porque o cenário continuava mudando e o padrão de subagents nunca se estabilizou de fato.

Avançando alguns meses: em maio de 2026, o Antigravity 2.0 adiciona suporte a subagents, mas com uma pegadinha: os subagents são definidos dinamicamente invocando a ferramenta `DefineSubagent`. No início, a `DefineSubagent` não dava muita flexibilidade: ela clonava o agente atual (padrão) com um novo prompt. Ganhamos o benefício do contexto limpo, mas perdemos no lado da reutilização de agentes. Eu não fiquei feliz, pois isso me impediu de realizar a evolução do GoDoctor da forma que eu tinha planejado.

Como não conseguia definir meus agentes personalizados com um modelo e conjunto de ferramentas diferentes do agente padrão, decidi ignorar a existência de subagents e foquei em portar o que funcionava bem no Gemini CLI para a Antigravity CLI com sucesso moderado.

Eu só revisitaria a ideia de subagents por causa deste prompt, publicado por [Richard Seroter](https://seroter.com/2026/06/01/one-prompt-four-subagents-and-ninety-seconds-to-get-a-working-app/) em junho:

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

Este prompt trazia algumas propostas muito interessantes, o que me fez revisitar o padrão, mas eu ainda estava preocupado com duas coisas: primeiro, o quanto eu precisaria adaptar meu estilo de prompt para pensar em termos de subagents e, segundo, por que eu iria querer escrever as coisas dessa forma?

Sou muito pragmático: se não tenho um benefício claro em qualidade e/ou velocidade, não quero gastar um esforço extra. Pensar em termos de subagents é muito semelhante a como pensamos sobre concorrência na programação clássica: a primeira pergunta é \"isso é sequer 'paralelizável'?\" e a segunda é \"vale a pena?\", já que a sobrecarga adicional muitas vezes anula pequenos ganhos.

No prompt de Richard, os únicos componentes que são claramente ortogonais são o desenvolvimento do backend e do frontend. Eles não dependem um do outro, desde que tenham um contrato claro para implementar. Mas todos os outros agentes têm algum tipo de dependência entre si, o que os torna mais seriais do que paralelos.

O benefício, então, deve vir apenas do isolamento de contexto, em vez de ganhos de velocidade devido a operações paralelas, e isso é difícil de medir nessa escala.

Passei talvez as duas semanas seguintes com esse pensamento rodando no fundo da minha mente: \"Que tipo de papéis são ortogonais entre si para que eu possa aproveitar os subagents?\"

Foi somente após uma série de conversas esclarecedoras no GDE Summit em Berlim que finalmente decifrei o código: Não se trata de **você** definir os subagents no prompt, mas sim de **ensinar** o próprio agente a decidir quando iniciar subagents. Em essência, eu estava pensando como um engenheiro líder dividindo o trabalho para a minha equipe, quando o que eu precisava fazer era fazer do próprio agente o engenheiro líder.

## The birth of swarm coding

O ato de decompor tarefas complexas em tarefas menores e ajudar a distribuí-las entre os membros da equipe não é novidade para mim. Antes de ingressar em Developer Relations, atuei como Tech Lead e depois Principal Engineer. Essas tarefas são literalmente o pão de cada dia da liderança técnica, especialmente se você vem de uma experiência com Agile como eu.

A mesma lógica de TL se aplica à criação do seu enxame de subagents: você quer garantir que cada agente tenha uma tarefa independente que possa executar de forma totalmente autônoma em relação aos outros. Para que a tarefa seja viável, ela deve ter especificações claras (conhecidas como Definição de Preparado ou *Definition of Ready*) e resultados finais claros (Definição de Concluído ou *Definition of Done*).

Como nota lateral, não são muitas as pessoas que descrevem essa parte do trabalho como a que mais gostam (eu inclusive), o que explica minha resistência em desenvolver um novo estilo de prompt que se tornaria essencialmente uma liderança técnica elevada ao extremo.

Então, em vez de agir como um TL para os meus agentes, decidi inverter os papéis e ensinar meu agente a se tornar o TL e montar sua própria equipe para executar minha visão. Foi assim que surgiu a [primeira versão](https://github.com/danicat/skills/blob/a9f57b10127d8bd23ed4867d64d168063a3726f4/swarm_coding/SKILL.md) do swarm coding. Um trecho das partes principais pode ser visto aqui:

> Swarm Coding é um novo paradigma de desenvolvimento que emprega múltiplos subagentes em paralelo para trabalhar em tarefas complexas. Baseia-se na estratégia de dividir para conquistar. Os principais benefícios desta estratégia são o isolamento de contexto e a melhoria da qualidade: ao atribuir pequenas tarefas independentes a subagentes, evita-se a diluição do contexto e permite-se um refinamento muito focado da solução. Por exemplo, sem swarm coding, um agente que implemente tanto o frontend quanto o backend frequentemente se distrairá, pois as habilidades necessárias para frontend e backend geralmente não estão relacionadas (diferentes stacks tecnológicas, diferentes boas práticas, etc.)
> 
> ## ROLE
> 
> Você é o SWARM COORDINATOR, seu papel é decompor tarefas complexas e DELEGAR para subagentes para execução. Você NUNCA deve executar tarefas por conta própria, não importa quão simples pareçam, A MENOS que seja EXPLICITAMENTE solicitado pelo usuário ou pelo seu coordenador pai. SEMPRE mantenha o canal de comunicação aberto para que o usuário ou agente pai possa lhe enviar comandos de direcionamento.
> 
> ## AGENT BUDGET
> 
> É o número de subagentes que você tem permissão para iniciar para trabalhar em uma tarefa. Você é incentivado a usar o ORÇAMENTO TOTAL de agentes, ou chegar o mais próximo possível dele. Isso não significa desperdiçar recursos em tarefas de baixo valor, mas sim encontrar o uso ideal do ORÇAMENTO para obter a melhor qualidade de entrega.
> 
> ## TEAM BUILDING
> 
> Para tarefas SIMPLES, decomponha a tarefa em elementos ortogonais e atribua um ou mais agentes ESPECIALISTAS para cada elemento.
> Para tarefas COMPLEXAS, decomponha a tarefa em partes menores e atribua AGENTES LÍDERES para cada uma delas. Os AGENTES LÍDERES devem ter uma fração do orçamento de agentes para executar a tarefa. Os AGENTES LÍDERES devem ativar a skill de swarm coding e se tornar o SWARM COORDINATOR para suas respectivas áreas.
> Proceda recursivamente até ter uma árvore completa de AGENTES LÍDERES e agentes EXECUTORES.
> 
> ## COMMUNICATION
> 
> O SWARM COORDINATOR é responsável por se comunicar diretamente com seus subagentes. Os subagentes não devem trocar mensagens entre si; a comunicação entre agentes do mesmo nível deve ser feita através de DOCUMENTOS DE DESIGN. É responsabilidade do SWARM COORDINATOR garantir que todas as alterações nos documentos de design sejam transmitidas aos agentes em sua equipe. Em caso de conflito, o SWARM COORDINATOR é responsável por desfazer a ambiguidade e tomar uma decisão.
> 
> ## PLANNING
> 
> O planejamento é um esforço de PRIMEIRA CLASSE e também deve ser feito utilizando o SWARM. Cada AGENTE deve contribuir para o plano com sua especialidade. É papel do SWARM COORDINATOR de uma equipe revisar a parte do plano produzida pelo seu time e corrigir inconsistências ou tomar decisões quando houver conflito.
> 
> ## EXECUTION
> 
> Na fase de execução, monitore o progresso do swarm ao longo dos principais marcos e oriente os agentes se necessário para mantê-los alinhados com o objetivo final. Lembre-se de que, como coordenador, você SÓ tem permissão para lidar com ARTEFATOS. Todas as tarefas de desenvolvimento devem ser tratadas por subagentes folha.

Esta foi a minha primeira vez escrevendo uma skill 100% manualmente, já que seria muito difícil alcançar minha visão de outra forma. Este prompt era um pouco ambicioso demais, pois eu queria que o Swarm fosse \"recursivo\" e que, com base apenas no orçamento de agentes, o agente decidisse se era um coordenador ou não, mas isso não funcionou como esperado.

O que aconteceu na prática foi que a tarefa dada pelo coordenador tinha precedência sobre qualquer outra instrução, e o subagent saltava direto para o modo de execução sem prestar atenção ao orçamento do agente. Corrigi isso na versão atual da skill fornecendo diretrizes mais claras e templates de prompt para iniciar os subagents.

## Taking the swarm for a spin

Você pode encontrar a versão atual da skill **swarm coding** no meu GitHub [aqui](https://github.com/danicat/skills). Você pode instalá-la em seu agente de codificação favorito com o comando abaixo:

```bash
$ npx skills add github.com/danicat/skills --skill swarm-coding
```

> Note: Esta skill é muito um trabalho em andamento, então faça um fork se quiser fixá-la a qualquer implementação específica

Aqui está um prompt divertido para começar. Tente executá-lo no Antigravity CLI:

> /swarm-coding agent budget 50. Develop a 2D tower defense survival game using Go and Ebitengine. The game should be feature complete and have one single screen level. Include an intro sequence, title screen, game win and game over screens as well. Track the high score at the end of each playthrough. Use 32x32 sprites with up to 256 colors each. The sprites should be custom designed for this game and each movement should have at least 3 frames of animation, but ideally 8. Tiles should be 32x32 as well. The level view is top down, movement is on four directions. The player should have access to 4 types of units and 4 types of buildings. The enemy waves should have 8 types of monsters, including one boss monster. Use typical build and attack phases with custom UIs for each. To create art, use vector graphics and/or dot (pixel) art creating each asset manually using binary data. Sound effects should be generated mathematically as well. The whole vibe of the game should match the 16-bit era, but with modern gameplay features.

Aqui estão os resultados na minha máquina:

![Swarm Defense](image-1.png "Screenshot of the game created by the swarm")

Não posso dizer que foi de primeira (*one-shot*) porque o primeiro build tinha um bug de renderização dos sprites, deixando a tela inteira preta, mas após mais um prompt relatando o problema, o jogo renderizou como mostrado acima.

Aqui está um pequeno vídeo dele em ação com a batalha final contra o chefe (coitado, não teve chance):

<video controls src="swarm-defense.mp4" title="Short clip of Swarm Defense boss fight"></video>

Cada asset neste vídeo foi gerado programmaticamente ou, em outras palavras, a Antigravity não tinha acesso a um modelo de geração de imagens. Portanto, ela teve que ser criativa e gerar os sprites no nível de bitmap.

Essa técnica só funcionou tão bem porque o swarm permitiu que os agentes se especializassem e focassem em uma única tarefa. Já tentei esse tipo de prompt antes com um único agente e geralmente produzia resultados insatisfatórios. Dê a um agente muitas tarefas ortogonais e ele claramente se tornará um mestre de nada. Mas com a delegação, cada agente pode ter uma tarefa única e autocontida e ter o melhor desempenho possível.

## Subagent support in Antigravity 2.0 and Antigravity CLI

No momento em que escrevo este artigo, as capacidades de subagents estão distribuídas de forma desigual entre o Antigravity 2.0 e a Antigravity CLI. Como essas interfaces são construídas para fluxos de trabalho diferentes, seus recursos de subagents divergiram temporariamente. Dado que ambas as ferramentas estão evoluindo rapidamente, podemos esperar que essa lacuna de recursos diminua à medida que ambas as interfaces continuem a amadurecer.

Em sua essência, ambos os ambientes compartilham o mesmo mecanismo subjacente. Iniciar um subagent delega a tarefa e retorna imediatamente o controle para você. O subagent roda com uma página em branco: ele usa o mesmo modelo da sessão padrão, mas começa com um contexto totalmente isolado, evitando o vazamento do histórico da conversa. O agente pai se comunica com ele por meio de IDs exclusivos. Se ele atingir um comando não aprovado, ele repassa a solicitação de permissão para você.

As diferenças notáveis entre as duas interfaces são:
- No Antigravity 2.0, o gerenciamento é visual. Você usa uma barra lateral gráfica para acompanhar tarefas em execução, visualizar logs de conversação ou interromper a execução. Agentes personalizados são criados dinamicamente em tempo real usando a ferramenta `DefineSubagent`. Não há suporte para plugins para subagents.
- Na Antigravity CLI, além de criar agentes dinamicamente, agentes personalizados também podem ser definidos estaticamente em arquivos Markdown, onde você pode usar opções de frontmatter para fixar modelos específicos ou controlar ferramentas disponíveis. A CLI também oferece suporte ao carregamento de subagents personalizados definidos dentro de plugins usando o formato Markdown.

Entender essas diferenças de interface é fundamental para configurar seu ambiente de swarm hoje, mas, como afirmado antes, essas capacidades provavelmente devem convergir à medida que ambas as ferramentas continuam a evoluir.

## Try it yourself

Acho que a melhor maneira de experimentar o poder dos subagents é testando você mesmo. Quer você queira tentar reproduzir meu prompt de exemplo ou criar o seu próprio, acredito que ficará impressionado com os resultados. Deixe-me saber sobre quaisquer coisas divertidas que você construir com o swarm. Enquanto isso, estarei por aqui refinando um pouco mais o [Swarm Defense](https://github.com/danicat/swarm-defense). :)

- Confira o swarm coding e todas as minhas outras skills em: https://github.com/danicat/skills
- Baixe e leia mais sobre o Antigravity em: https://antigravity.google
