---
categories:
- Agentic Coding
date: 2026-06-15 00:00:00+00:00
heroStyle: big
summary: Explore a evolução do desenvolvimento agentivo. Uma atualização sobre a mudança
  para sistemas de planejamento, skills estratégicas e orquestração de subagentes para
  melhores resultados.
tags:
  - agent-skills
  - antigravity
  - hooks
  - mcp
  - subagents
title: "O Estado do Desenvolvimento Agentivo em 2026: Skills e Hooks"
slug: "state-of-agentic-coding"
aliases:
  - "/pt-br/posts/20260615-state-of-agentic-coding/"
description: "Retrospectiva de 2026 sobre desenvolvimento agentivo. Compara sistemas de planejamento, Agent Skills vs MCP, orquestração de subagentes e hooks determinísticos."
proficiencyLevel: "Intermediate"
dependencies:
  - "Google Antigravity 2.0"
  - "Model Context Protocol"
  - "Agent Skills"
---

Já se passaram mais de seis meses desde que publiquei o artigo [Taming Vibe Coding]({{< ref "/posts/20251206-taming-vibe-coding" >}}), que consolidou as principais práticas que eu vinha adotando para aumentar minha produtividade com agentes de programação.

Embora a maior parte daquele artigo continue relevante, muita coisa mudou nessa área ao longo do último semestre. Por isso, decidi trazer uma atualização rápida sobre como minha perspectiva evoluiu desde então.

## Prompting e engenharia de contexto

O trabalho de *prompting* continua importante, e um bom prompt estruturado poupa bastante tempo. No entanto, ele já não é o divisor de águas de outrora, graças ao surgimento dos sistemas de planejamento. A maioria dos agentes de código modernos agora conta com um modo de planejamento (*plan mode*), no qual passam alguns turnos fazendo um *brainstorming* da tarefa para, só então, elaborar um plano de implementação detalhado antes de tocar no código.

Isso dá a você a oportunidade de revisar o plano e direcionar o agente antes que qualquer linha seja escrita, economizando muitos tokens e tempo. O que não mudou foi a necessidade de critérios de aceitação bem definidos para garantir resultados consistentes. Essa é a parte à qual presto mais atenção sempre que reviso um plano.

A engenharia de contexto melhorou drasticamente com a adoção generalizada das *agent skills*, a ponto de eu quase não me preocupar mais em escrever um `AGENTS.md` (ou `GEMINI.md`, `CLAUDE.md`, etc.). No Antigravity, temos o conceito de *Rules*, que é essencialmente uma otimização de contexto, mas raramente as utilizo, preferindo focar em skills. Ainda não encontrei nada que precisasse colocar em `AGENTS.md` ou em regras que não pudesse ser resolvido com outra técnica.

Os aspectos de engenharia de contexto ligados à Geração Aumentada por Recuperação (RAG) — seja via busca semântica ou outros métodos — continuam indispensáveis para conhecimentos especializados. É assim que os sistemas de memória operam, e ainda uso essa abordagem para injetar a documentação de pacotes no contexto sempre que trabalho com uma dependência externa.

## A ascensão e queda (?) do Model Context Protocol

Em vez de subir servidores MCP, cada vez mais pessoas preferem usar *agent skills* combinadas com ferramentas de linha de comando (CLI). Embora compreenda a praticidade, ainda tenho um pé atrás na hora de dar acesso irrestrito ao shell para os agentes, preferindo empacotar minhas ferramentas em servidores MCP locais (como o [godoctor](https://github.com/danicat/godoctor) ou o [speedgrapher](https://github.com/danicat/speedgrapher)). A única mudança prática para mim é que me tornei muito mais seletiva ao instalar servidores MCP: na maior parte do tempo, mantenho apenas os meus próprios configurados.

Também quase nunca utilizo um servidor MCP isolado; combino sempre *skill* mais MCP. A skill descreve o processo; o MCP expõe o ferramental. Essa combinação costuma ser muito mais eficaz do que passar centenas de horas tentando aperfeiçoar as instruções internas do MCP — algo que fiz à exaustão, apenas para ver os agentes ignorarem tudo solenemente.

Embora use MCP + skills no dia a dia, para casos realmente críticos recorro ao "super combo": MCP + skills + hooks. O papel dos hooks é forçar o agente na direção certa. Costumo chamar isso de "colocar o agente nos trilhos" ou "reduzir a agência do agente". O sistema de hooks me permite barrar ações indesejadas e dar um empurrãozinho para que o agente use a ferramenta exata que desejo, eliminando o fator probabilístico de uma ferramenta ser acionada ou não. Em outras palavras: ele impõe um comportamento determinístico ao agente.

## Skills para tudo

Crio skills para qualquer processo que precise ser repetível. Por exemplo, algumas das minhas skills mais utilizadas tratam de escrita técnica e revisão, já que estou sempre produzindo conteúdo (como os artigos deste blog). Também escrevo skills para tecnologias com as quais sei que o agente terá dificuldades — geralmente novidades, projetos de nicho ou ferramentas proprietárias.

Recentemente, por exemplo, passei por maus bocados preparando meu workshop sobre A2UI. O A2UI é um protocolo recente voltado para o desenvolvimento de interfaces de usuário agentivas. Por ser um conceito tão novo e com requisitos didáticos específicos, os agentes não conseguiam entender nada daquilo sem muita orientação e tentativa e erro. Depois de superar esses primeiros obstáculos, empacotei todo o conhecimento em uma skill, garantindo um fluxo tranquilo para as próximas vezes em que precisar fazer algo parecido.

Pelo lado negativo, sou péssima em manter minhas skills organizadas e atualizadas. Acredito que resolver essa manutenção possa ser a redenção do MCP, já que existe uma proposta para incorporar skills à especificação do protocolo. Infelizmente, não há previsão de quando isso se tornará realidade, se é que vai acontecer, então continuamos com a responsabilidade de gerenciar as skills por conta própria. Em teoria, seria possível criar uma experiência parecida com skills no MCP usando prompts para a revelação progressiva e tools para automações com scripts, mas com meu backlog atual ainda não cheguei a testar isso.

## A ascensão (sem queda) dos subagentes

Subagentes são a nova grande sensação da qual todo mundo está falando. A ideia básica é paralelizar tarefas delegando-as para agentes independentes, cada um com sua própria janela de contexto. A principal vantagem é o uso inteligente de contexto, evitando a contaminação entre tarefas e a degradação precoce da janela. Isso também diminui a necessidade de compressão, já que cada tarefa é autocontida e não polui o contexto principal.

Em termos de suporte nos ambientes de execução (*harnesses*), algumas ferramentas permitem pré-declarar subagentes assim como você declara uma skill (cada um com seu próprio prompt de sistema, modelo e ferramentas), enquanto outras, como o Antigravity, favorecem a criação ad-hoc — onde cada novo agente nasce como um "clone" da sessão principal, mas com sua janela de contexto limpa.

Embora a criação ad-hoc permita experiências ousadas (como usar um prompt para disparar 3 agentes simultâneos em paralelo), sinto falta de agentes pré-declarados no Antigravity, pois eles facilitam empacotar meus especialistas favoritos de forma portátil. Além disso, paralelizar tarefas entre agentes está longe de ser trivial: no último ano me acostumei a delegar tarefas para agentes, mas ainda não tinha pensado a fundo sobre como orquestrá-los com eficiência.

Sob vários aspectos, esse processo exige o mesmo raciocínio exercitado por uma liderança técnica ou product owner ao quebrar demandas e coordenar a entrega do time. A diferença é que, com subagentes, em vez de um time fixo, você pode convocar quantas "pessoas" quiser para a equipe. No fim das contas, me importo menos em paralelizar agentes só porque é algo "moderno", e muito mais com a pergunta fundamental: "isso realmente vai gerar resultados melhores?"

Se a resposta for **não**, vale muito mais a pena rodar um agente por vez. O esforço gasto no planejamento e o desgaste mental de fatiar minuciosamente as tarefas simplesmente não compensam. É exatamente por isso que prefiro pré-definir meus agentes: foco em especializar cada um deles para quando forem necessários, sem me preocupar se vão rodar em paralelo ou de forma sequencial.

## Hooks são seus melhores aliados

Deixei o meu favorito para o final: os hooks. Hooks são callbacks disparados em eventos específicos do ciclo de vida do agente. Publiquei um [artigo completo sobre hooks]({{< ref "/posts/20260610-mastering-hooks" >}}) na semana passada, que recomendo a leitura assim que terminar este. Em resumo: modelos são imprevisíveis e podem sair dos trilhos facilmente. Hooks são uma excelente maneira de criar barreiras de proteção, garantindo que o modelo siga a direção planejada sem depender da sorte. Além disso, eles permitem conectar rotinas de monitoramento ao *harness* para coletar dados e enriquecer a qualidade das respostas — como a integração com um sistema de memória persistente.

## Conclusão

Nossa indústria caminha a passos largos e, para continuarmos relevantes, precisamos ser adaptáveis, adotando novos processos e técnicas à medida que surgem e deixando de lado o que não nos serve mais. Mesmo assim, não trate nenhum guia por aí como verdade absoluta (incluindo este). Estamos todos em uma grande jornada de aprendizado, pois essa tecnologia ainda dá seus primeiros passos. O segredo é experimentar e descobrir qual fluxo e ferramental funcionam melhor no seu contexto.

Neste artigo, compartilhei o que tem funcionado para mim e como minha forma de pensar vem evoluindo, mas estou longe de saber tudo e sigo aprendendo constantemente. Fala-se muito em treinar IA, mas nunca se esqueça de que exercitar seu próprio cérebro é infinitamente mais importante. Não use a IA como muleta para desligar o raciocínio: continue testando, aprendendo e iterando. E, por favor, compartilhe qualquer descoberta interessante que fizer pelo caminho!
