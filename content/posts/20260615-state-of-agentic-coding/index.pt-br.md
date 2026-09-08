---
categories:
- Agentic Coding
date: 2026-06-15 00:00:00+00:00
heroStyle: big
summary: Explore a evolução do desenvolvimento agentivo. Uma atualização sobre a transição para sistemas de planejamento, skills estratégicas e a orquestração de subagentes para alcançar melhores resultados.
tags:
  - agent-skills
  - antigravity
  - hooks
  - mcp
  - subagents
title: "O Estado do Desenvolvimento Agentivo em 2026: Skills, Hooks e Subagentes"
slug: "state-of-agentic-coding"
aliases:
  - "/pt-br/posts/20260615-state-of-agentic-coding/"
description: "Retrospectiva sobre desenvolvimento agentivo em 2026. Compara sistemas de planejamento, Agent Skills vs MCP, orquestração de subagentes e hooks determinísticos em produção."
proficiencyLevel: "Intermediate"
dependencies:
  - "Google Antigravity 2.0"
  - "Model Context Protocol"
  - "Agent Skills"
---

Já se passaram mais de seis meses desde que publiquei o artigo [Taming Vibe Coding]({{< ref "/posts/20251206-taming-vibe-coding" >}}), que consolidou as principais práticas que eu vinha utilizando para aumentar minha produtividade com agentes de programação.

Embora a maior parte do conteúdo daquele artigo continue relevante hoje, muita coisa aconteceu nessa área no último semestre. Por isso, decidi publicar uma atualização rápida sobre como minha forma de pensar evoluiu desde então.

## Prompting e engenharia de contexto

O prompting continua importante, e um bom prompt estruturado economiza bastante tempo, mas já não é o divisor de águas de antes graças ao surgimento dos sistemas de planejamento. A maioria dos agentes de código hoje já vem com um modo de planejamento (plan mode), no qual passam alguns turnos fazendo "brainstorming" sobre a tarefa para depois elaborar um plano de implementação antes de partir direto para o código.

Isso lhe dá a oportunidade de revisar o plano e direcionar o agente antes que o código seja escrito, economizando muito tempo e tokens. O que não mudou foi a necessidade de critérios de aceitação sólidos para garantir resultados consistentes. Essa é a parte à qual costumo prestar mais atenção quando reviso um plano.

A engenharia de contexto melhorou absurdamente com a adoção generalizada de agent skills, o que faz com que hoje eu mal me preocupe em escrever um AGENTS.md (ou GEMINI.md, CLAUDE.md, etc.). No Antigravity temos o conceito de Rules, que é essencialmente otimização de contexto, mas quase não as uso, preferindo as skills. Ainda não encontrei nada que precise fazer em AGENTS.md ou em regras que não consiga fazer usando outra técnica.

Os aspectos de engenharia de contexto relacionados à Geração Aumentada por Recuperação (RAG) — seja via busca semântica ou qualquer outro meio — continuam relevantes para conhecimento especializado. É assim que os sistemas de memória funcionam, e eu ainda uso essa técnica para injetar documentação de pacotes no meu contexto sempre que trabalho com uma dependência externa.

## A ascensão e queda (?) do Model Context Protocol

Em vez de subir servidores MCP, cada vez mais pessoas estão preferindo usar agent skills combinadas com ferramentas CLI. Embora eu respeite isso, tenho um certo receio em dar aos meus agentes acesso direto ao shell, por isso prefiro empacotar minhas ferramentas em servidores MCP que rodo localmente (como o [GoDoctor]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}}) ou o [Speedgrapher]({{< ref "/posts/20250805-introducing-speedgrapher" >}})). A única coisa que mudou para mim é que estou muito mais seletiva ao instalar servidores MCP, e na maior parte do tempo mantenho apenas os meus próprios configurados.

Também quase nunca uso um servidor MCP sozinho; sempre uso uma combinação de skill mais MCP. A skill descreve o processo, enquanto o MCP expõe o ferramental. Isso costuma ser bem melhor do que tentar otimizar as instruções do próprio MCP — algo em que gastei centenas de horas fazendo, apenas para ver os agentes de código ignorá-las solenemente.

Embora eu use MCP + skills na maior parte do tempo, para coisas realmente sérias recorro ao "super combo": MCP + skills + hooks. A parte dos hooks é responsável por forçar o agente na direção que desejo seguir. Às vezes chamo isso de "colocar o agente nos trilhos" ou "reduzir a agência do agente". O sistema de hooks me permite bloquear ações indesejadas e dar um "empurrãozinho gentil" para que o agente use a ferramenta exata que pretendo, removendo o elemento de probabilidade de uma ferramenta ser chamada ou não. Em outras palavras, ele força um comportamento determinístico do agente.

## Skills para tudo

Crio skills para processos que quero que sejam repetíveis. Por exemplo, algumas das minhas skills mais usadas são relacionadas a escrita técnica e revisão, já que estou sempre produzindo conteúdo (como este blog). Também escrevo skills sobre tecnologias com as quais sei que o agente terá dificuldades para lidar. Normalmente, isso inclui tecnologias mais recentes, projetos de nicho ou ferramentas customizadas.

Por exemplo, recentemente passei por maus bocados preparando meu workshop de A2UI. O A2UI é um protocolo relativamente novo para desenvolver interfaces de usuário agentivas. Por ser um conceito tão novo e com necessidades didáticas muito específicas, os agentes não conseguiam entender sem muita ajuda e tentativa e erro. Depois de superar esses obstáculos iniciais, empacotar esse conhecimento em uma skill facilitou muito as coisas para as próximas vezes em que eu precisar fazer algo semelhante.

Pelo lado negativo, sou péssima em manter minhas skills atualizadas e organizadas. Acredito que resolver esse problema pode ser o momento de redenção do MCP, já que existe uma proposta atual para adicionar skills à especificação do protocolo. Infelizmente, não há previsão de quando isso vai acontecer — se é que vai —, então, por enquanto, ficamos encarregados de gerenciar as skills por conta própria. Em teoria, seria possível criar uma experiência parecida com skills usando MCP com prompts para a parte de revelação progressiva e tools para o que exigir scripts, mas, com o tamanho do meu backlog atual, ainda não tentei fazer isso.

## A ascensão (e sem queda) dos subagentes

[Subagentes]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}) são a "próxima grande novidade" sobre a qual todo mundo na praça está comentando. A ideia é paralelizar tarefas iniciando-as como agentes próprios, com uma janela de contexto segregada. Isso traz a vantagem de um uso otimizado do contexto, evitando contaminação entre tarefas e a degradação precoce do contexto. Isso também reduz a necessidade de compressão, já que cada tarefa é autocontida e não polui sua janela de contexto principal. Em termos de suporte nos harnesses de programação, alguns ambientes suportam pré-declarar agentes assim como você declara uma skill, cada um com seu próprio system prompt, modelo e configurações, enquanto outros, como o Antigravity, favorecem a criação ad-hoc de agentes, onde cada agente é um "clone" da sessão principal, mas com sua própria janela de contexto.

Embora a criação ad-hoc de agentes permita fazer loucuras, como usar um prompt para iniciar 3 agentes diferentes rodando em paralelo, sinto falta de ter agentes pré-declarados no Antigravity, pois eles me permitiriam empacotar meus agentes "curados" de forma portátil. Além disso, paralelizar tarefas entre agentes não é uma habilidade trivial de dominar — no último ano me acostumei a delegar tarefas para agentes, mas ainda não tinha pensado a fundo sobre como orquestrá-los de maneira eficiente.

De muitas maneiras, esse é o mesmo músculo que um product owner ou tech lead exercita ao quebrar tarefas e planejar como o time vai abordá-las, mas com subagentes, em vez de uma equipe fixa, você pode criar quantos "membros de equipe" quiser. No fim das contas, me importo menos em paralelizar agentes só porque é "legal", e muito mais com a pergunta: "isso vai produzir resultados melhores?"

Se a resposta para essa pergunta for **não**, é muito melhor rodar um agente por vez, pois o esforço investido no planejamento e o desgaste mental de dividir cuidadosamente as tarefas não compensarão. É por isso que prefiro poder pré-definir meus agentes: eu apenas os especializo para quando forem necessários, sem me preocupar se vão rodar em paralelo ou não.

## Hooks são seus melhores amigos

Deixei o meu favorito para o final: hooks. Hooks são callbacks disparados mediante eventos específicos no ciclo de vida do agente. Escrevi um [artigo inteiro sobre hooks]({{< ref "/posts/20260610-mastering-hooks" >}}) na semana passada, que recomendo conferir logo após terminar este. No fim das contas, os modelos são imprevisíveis e podem sair dos trilhos com bastante facilidade. Hooks são uma ótima maneira de colocar guardrails nos modelos para que você possa direcioná-los para onde quer sem depender da sorte. Mais do que isso, eles permitem plugar monitores ao harness agentivo para coletar dados e melhorar a qualidade das respostas, como, por exemplo, adicionar um sistema de memória persistente.

## Conclusões

O setor está se movendo rápido e, para continuarmos relevantes, precisamos ser adaptáveis para abraçar novos processos e técnicas à medida que surgem e nos desfazermos da bagagem que está nos atrasando. Ainda assim, não encare nenhum guia por aí como a verdade absoluta (inclusive este). É uma experiência de aprendizado para todo mundo, já que essa tecnologia ainda está em seus estágios iniciais. O segredo é experimentar e ver que tipo de tecnologia e fluxo de trabalho funciona melhor no seu ambiente.

Neste artigo compartilhei o que tem funcionado para mim e como meu pensamento vem evoluindo, mas não sei de tudo e estou sempre aprendendo. Falamos muito sobre treinar IA, mas não se esqueça de que treinar o seu cérebro é muito mais importante do que isso. Não use a IA como desculpa para desligar o cérebro: continue experimentando, aprendendo e iterando. E, por favor, compartilhe qualquer coisa interessante que descobrir!
