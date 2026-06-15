---
title: "Estado da Codificação Agente"
date: 2026-06-15T00:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["antigravity", "agy-cli", "agentic-coding"]
summary: "Explore a evolução da codificação agente. Atualização sobre a mudança para sistemas de planejamento, habilidades estratégicas de agentes e orquestração de subagentes para melhores resultados."
heroStyle: "big"
---

Já se passaram mais de seis meses desde que publiquei o artigo [Taming Vibe Coding]({{< ref "/posts/20251206-taming-vibe-coding" >}}), que consolidou todas as principais práticas que eu vinha usando para aumentar minha produtividade com agentes de codificação.

Embora a maior parte do conteúdo daquele artigo ainda seja relevante hoje, nos últimos seis meses muita coisa aconteceu neste espaço, por isso decidi postar uma atualização rápida sobre como meu pensamento evoluiu desde então.

## Prompting e engenharia de contexto

O prompting ainda é importante, e um bom prompt estruturado economizará muito tempo, mas não é mais o divisor de águas que costumava ser devido ao surgimento de sistemas de planejamento. A maioria dos agentes de codificação hoje é enviada com um modo de plano (ou planejamento), no qual eles passam alguns turnos fazendo um "brainstorming" da tarefa para então elaborar um plano de implementação antes de pular para a codificação.

Isso lhe dá a oportunidade de revisar o plano e orientar o agente antes que o código seja escrito, economizando muito tempo e tokens. O que não mudou é a necessidade de critérios de aceitação fortes para garantir resultados consistentes. Esta é a parte à qual costumo prestar mais atenção quando reviso um plano.

A engenharia de contexto melhorou massivamente com a adoção generalizada de habilidades de agente, o que faz com que eu mal me importe em escrever um AGENTS.md (ou GEMINI.md, CLAUDE.md, etc.) hoje em dia. No Antigravity, temos o conceito de Rules, que é essencialmente otimização de contexto, mas raramente as uso em favor de habilidades (skills). Ainda não encontrei nada que eu precise fazer em AGENTS.md ou regras que não possa fazer usando outra técnica.

Os aspectos da engenharia de contexto relacionados à Geração Aumentada de Recuperação (RAG), seja por meio de busca semântica ou qualquer outro meio, ainda são relevantes para conhecimentos especializados. É assim que os sistemas de memória funcionam e ainda uso a técnica para injetar documentação de pacotes em meu contexto sempre que trabalho com uma dependência externa.

## A ascensão e queda (?) do Model Context Protocol

Em vez de criar servidores MCP, cada vez mais pessoas preferem usar habilidades de agente combinadas com ferramentas de CLI. Embora eu respeite isso, tenho alguns problemas de confiança em dar aos meus agentes acesso direto ao shell e, por isso, ainda prefiro empacotar minhas ferramentas em servidores MCP que executo localmente (como [godoctor](https://github.com/danicat/godoctor) ou [speedgrapher](https://github.com/danicat/speedgrapher)). A única coisa que mudou para mim é que sou muito mais seletivo na instalação de servidores MCP e, na maioria das vezes, tenho apenas meus personalizados configurados.

Eu também quase nunca uso um servidor MCP sozinho, mas uso uma combinação de habilidade mais MCP. A habilidade descreve o processo, o MCP expõe o ferramental. Isso costuma ser melhor do que tentar otimizar as próprias instruções do MCP, o que gastei centenas de horas fazendo, apenas para que os agentes de codificação as ignorassem completamente.

Embora eu use MCP + habilidades na maior parte do tempo, para coisas realmente sérias uso o "super combo" MCP + habilidades + hooks. A parte de hooks é responsável por forçar o agente na direção que desejo seguir. Às vezes chamo isso de "colocar o agente nos trilhos" ou "reduzir a agência do agente." O sistema de hooks me permite bloquear as ações indesejáveis e dar ao agente um "empurrãozinho gentil" para usar a ferramenta que desejo que ele use, removendo o elemento de probabilidade de ter uma ferramenta chamada ou não. Ou, em outras palavras, força um comportamento determinístico do agente.

## Habilidades para tudo

Crio habilidades para processos que desejo que sejam repetíveis. Por exemplo, algumas das minhas habilidades mais usadas estão relacionadas à escrita técnica e revisão, porque estou sempre produzindo conteúdo (como este blog). Também escrevo habilidades sobre tecnologias que sei que o agente terá dificuldade em lidar. Normalmente, elas incluem tecnologias mais novas, projetos de nicho ou personalizados.

Por exemplo, recentemente tive muita dificuldade em preparar meu workshop de A2UI. A2UI é um protocolo novo para desenvolver interfaces de usuário agentes. Por ser um conceito tão novo, e eu ter necessidades de ensino muito específicas, os agentes não consegui-lo entender sem muita ajuda e tentativa e erro. Uma vez superados os obstáculos iniciais, empacotar esse conhecimento em uma habilidade ajuda a suavizar as coisas na próxima vez que eu precisar fazer algo semelhante.

Pelo lado negativo, sou terrível em manter minhas habilidades atualizadas e organizadas. Acho que resolver esse problema pode ser o momento de redenção para o MCP, pois há uma proposta atual para adicionar habilidades à especificação do protocolo. Infelizmente, não há estimativa de quando isso acontecerá, se é que acontecerá, então, por enquanto, estamos presos ao gerenciamento de habilidades por conta própria. Em teoria, você poderia criar uma experiência "semelhante a uma habilidade" com o MCP usando prompts para a parte de divulgação progressiva e ferramentas para o que quer que precise de script, mas dado o meu backlog atual, ainda não tentei fazer isso.

## A ascensão (e não queda) de subagentes

Os subagentes são a "próxima coisa legal" de que todos na vila estão falando. A ideia é paralelizar tarefas lançando-as como seus próprios agentes com uma janela de contexto segregada. Isso tem o benefício do uso ideal do contexto, evitando a contaminação entre tarefas e a deterioração precoce do contexto. Isso também reduz a necessidade de compressão, já que cada tarefa é autocontida e não poluirá sua janela de contexto principal. Em termos de suporte ao harness de codificação, alguns harnesses suportam a pré-declaração de agentes assim como você pode declarar uma habilidade, cada um com seu próprio prompt de sistema, modelo e configurações, enquanto outros, como o Antigravity, favorecem a criação ad-hoc de agentes, sendo cada agente um "clone" da sessão principal, mas com suas próprias janelas de contexto.

Embora a criação ad-hoc de agentes permita que você faça coisas malucas, como usar um prompt para criar 3 agentes diferentes para executar coisas em paralelo, sinto falta de ter agentes pré-declarados no Antigravity, pois eles me permitem empacotar meus agentes "curados" de forma portátil. Além disso, paralelizar tarefas entre agentes não é uma habilidade trivial de dominar - no último ano me acostumei a dar tarefas aos agentes, mas não pensei muito sobre como orquestrá-los com eficiência.

De muitas maneiras, este é o mesmo músculo que um proprietário de produto ou líder de equipe exercita ao decompor tarefas e pensar em como a equipe irá abordá-las, mas com subagentes, em vez de ter uma equipe fixa, você pode criar quantos "membros de equipe" quiser. No final das contas, me importo menos em paralelizar agentes porque é "legal", e me importo mais com a pergunta "isso vai produzir melhores resultados?"

Se a resposta à pergunta for **não**, então é melhor executar um agente de cada vez, pois o esforço gasto no planejamento e o desgaste mental de dividir cuidadosamente as tarefas não valerão a pena. É por isso que prefiro poder pré-definir meus agentes, pois estou apenas especializando-os para quando forem necessários, mas realmente não me importa se eles rodam em paralelo ou não.

## Hooks são seus melhores amigos

Deixei o meu favorito por último: hooks. Hooks são callbacks que serão acionados em eventos específicos no ciclo de vida do agente. Escrevi um [artigo inteiro sobre hooks]({{< ref "/posts/20260610-mastering-hooks" >}}) na semana passada, que encorajo você a conferir logo após terminar este. O ponto principal é que os modelos são imprevisíveis e podem sair dos trilhos com facilidade. Hooks são uma boa maneira de adicionar proteções aos modelos para que você possa orientá-los na direção desejada sem depender do acaso. Mais do que isso, eles permitem conectar monitores ao harness agente para coletar dados e melhorar a qualidade da resposta, como, por exemplo, adicionar um sistema de memória persistente.

## Conclusões

A indústria está se movendo rápido e, para nos mantermos relevantes, precisamos ser adaptáveis para adotar novos processos e técnicas à medida que surgem e liberar a bagagem que está nos atrasando. Ainda assim, não tome nenhum dos guias por aí como a fonte da verdade (incluindo este). É uma experiência de aprendizado para todos, pois essa tecnologia ainda está em seus estágios iniciais. A chave é experimentar e ver que tipo de tecnologia e fluxo de trabalho funciona melhor para o seu ambiente.

Neste artigo, compartilhei o que está funcionando para mim e como meu pensamento tem evoluído, mas não sei tudo e estou sempre aprendendo. Falamos muito sobre treinar IA, mas não se esqueça de que treinar seu cérebro é muito mais importante do que isso. Não use a IA como desculpa para desativar seu cérebro, continue experimentando, aprendendo e iterando. E, por favor, compartilhe qualquer coisa interessante que você descobrir!
