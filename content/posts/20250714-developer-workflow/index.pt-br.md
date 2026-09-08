---
author: Daniela Petruzalek
categories:
- Agentic Coding
date: 2025-07-11
summary: Uma proposta de fluxo de trabalho moderno para pessoas desenvolvedoras usando ferramentas de IA e um exercício simples de priorização.
tags:
  - gemini-cli
  - jules
  - vibe-coding
title: "Um Fluxo de Trabalho Moderno para o Mundo Habilitado por IA"
slug: "developer-workflow-ai-world"
aliases:
  - "/pt-br/posts/20250714-developer-workflow/"
description: "Um framework bidimensional de priorização para fluxos de desenvolvimento com IA. Equilibre certeza técnica e valor de negócio com a Gemini CLI e o Jules."
proficiencyLevel: "Intermediate"
dependencies:
  - "Gemini CLI"
  - "Jules AI Agent"
  - "Git"
---
## Introdução

Acabei de voltar do WeAreDevelopers World Congress 2025 em Berlim e voltei inspirada pelas inúmeras pessoas desenvolvedoras que conheci de toda a Europa e do mundo. Como esperado, o grande tema deste ano foi IA — IA em todo lugar! Agora temos IA na nossa nuvem, no computador, nos óculos de sol, na torradeira, na pia da cozinha e até no rolo de papel higiênico. Ninguém consegue escapar da IA... nem mesmo os frameworks JS conseguem nascer mais rápido que ela! Estamos CONDENADOS!!! >.<

Ou talvez não! Eu sei que estamos vivendo tempos assustadores. A indústria de tecnologia está em transformação. Empresas estão demitindo a torto e a direito sob a premissa de que a IA torna as pessoas mais produtivas ou até substitui pessoas por completo. Se essa é a causa real ou se a IA é apenas um bode expiatório para outros interesses, essa é uma conversa para a mesa de bar e não para este post. Mas vamos ao menos reconhecer que as mudanças estão acontecendo.

Uma lição de vida importante que aprendi com o tempo e a experiência é que a gente não deve se preocupar com as coisas que não pode controlar. A IA é inevitável. Portanto, em vez de nos preocuparmos com o que será dos nossos empregos no futuro, convido você para uma reflexão sobre o que a IA pode de fato fazer para melhorar o seu trabalho hoje. É a clássica visão do copo meio cheio: transformar uma crise em oportunidade. Vamos baixar a guarda por um instante e imaginar como pode ser o fluxo de trabalho de quem desenvolve software ao incorporar o "vibe coding" à rotina.

Este é, na verdade, um fluxo de trabalho que venho aplicando nas últimas quatro semanas e que tem funcionado muito bem — mas, claro, veja com uma certa reserva, pois ainda não está perfeito. Dito isso, esse ecossistema evolui todos os dias e acredito que só vai melhorar.

## Uma abordagem bidimensional de priorização

Antes de entrarmos nas práticas realmente "AI-native" de trabalhar, quero apresentar brevemente um modelo de priorização que utilizo há mais de 7 anos na minha carreira. Aprendi essa abordagem trabalhando na ThoughtWorks em projetos de transformação ágil e a adaptei para as minhas próprias necessidades. O método consiste em reunir as pessoas relevantes na mesma sala para uma discussão — incluindo tanto engenharia quanto stakeholders — e produzir uma sequência lógica para a execução do backlog.

Ele se baseia em dois eixos ortogonais: `certeza técnica` (`technical certainty`) e `valor de negócio` (`business value`).

![Resultados do exercício de priorização](image-3.png "Prioridade de implementação com base nos resultados do exercício")

A **certeza técnica** mede o quanto o caminho de implementação de uma funcionalidade já está claro. Se a certeza técnica é alta, significa que todos (ou quase todos) os passos necessários para implementar a funcionalidade são conhecidos. Se for baixa, significa que não sabemos como resolver o problema ou temos apenas algumas hipóteses iniciais.

O **valor de negócio** indica o quão importante essa funcionalidade é para alcançar os objetivos do time. Alto valor de negócio significa que a feature é crítica para o sucesso do produto; baixo valor indica algo no nível de um "bom ter" (*nice to have*).

O grande objetivo deste exercício é destravar aqueles impasses onde **tudo** é considerado prioritário. Mesmo quando tudo **é** importante, colocar os itens lado a lado ajuda até os stakeholders mais resistentes a repensar suas posições em relação às outras prioridades. Além disso, itens de mesmo valor mas com certezas técnicas distintas ganham uma ordem de execução natural: priorizar os "ganhos rápidos" (*low hanging fruit*) primeiro costuma comprar tempo para o time investigar e realizar *spikes*, reduzindo incertezas e elevando a certeza técnica dos itens restantes.

![Modos de trabalho por prioridade](image-2.png "Modo de trabalho recomendado com base na priorização")

Como esse processo se relaciona com o trabalho AI-native? Gosto de me ver como a chefe de várias IAs. Eu priorizo meu próprio backlog e defino qual ferramenta alocar para cada tarefa. Se estou fortemente envolvida em uma funcionalidade, dou prioridade para trabalhar com ela de forma síncrona; caso contrário, posso delegá-la para um agente assíncrono.

## Um fluxo de trabalho básico "AI-native"

Digamos que eu precise implementar uma nova funcionalidade no meu sistema. Tenho dois modos principais de operação: o **interativo** (síncrono) e o **em lote** (assíncrono, no estilo *fire-and-forget*). A decisão entre um e outro depende diretamente da `certeza técnica` que tenho sobre como implementar aquela funcionalidade e do meu nível de dedicação a ela no momento (`valor de negócio`).

![Exemplo de priorização](image.png "Exemplo de funcionalidades priorizadas para este blog")

Por exemplo, neste mesmo blog, escrevi no post anterior ([Como Usei o Jules para Criar um Post em Destaque no Blog]({{< ref "/posts/20250703-jules-featured-post" >}})) sobre como implementei o "post em destaque" na página inicial. Quando comecei a trabalhar nisso, eu não sabia nada sobre como implementar um post em destaque, mas tinha uma ideia clara do que queria alcançar. Esse era um problema com baixa certeza técnica, porque eu não conhecia a tecnologia nem os pontos exatos do código que precisaria alterar. Ao mesmo tempo, ele tinha um alto valor de negócio para mim, pois a minha hipótese era de que isso deixaria o blog muito mais profissional e atraente para quem lê.

Diante disso, a escolha natural era ter um ciclo de feedback curto: para cada alteração sugerida pela IA, eu precisava avaliar o resultado visual imediatamente e corrigir a rota na direção certa. Por outro lado, para itens de menor prioridade, não há problema em ter um ciclo de feedback longo, o que os torna ideais para o modo em lote.

## Certeza Técnica Baixa/Média OU Alto Valor de Negócio = Modo Interativo (Síncrono)

Problemas com menor certeza técnica exigem mais supervisão para fazer as coisas do jeito certo, então prefiro usar um processo interativo com uma ferramenta de CLI. Minha ferramenta de escolha atual é a [Gemini CLI](https://cloud.google.com/gemini/docs/codeassist/gemini-cli?utm_campaign=CDR_0x72884f69_default_b431747616&utm_medium=external&utm_source=blog), lançada pelo Google há apenas duas semanas e que já vem conquistando a comunidade de desenvolvimento em massa.

A Gemini CLI é uma aplicação de linha de comando no estilo REPL (*Read-Eval-Print Loop*) potencializada por IA. Você digita um prompt e a CLI reage com uma resposta — que pode ser não apenas código e texto, mas praticamente qualquer coisa graças ao suporte ao Model Context Protocol (MCP). Com isso, você pode usar a CLI para quase tudo, desde pedir um café até atualizar seus bancos de dados. É claro que o caso de uso natural da CLI é codificação, mas vocês sabem como são os desenvolvedores. :)

Embora a Gemini CLI conte com um modo YOLO projetado para automação, honestamente ainda não confio nela o suficiente para fazer tarefas sem supervisão (falarei mais sobre isso adiante). Por isso, prefiro usar a CLI quando preciso fazer brainstorming e explorar o espaço do problema antes de chegar a uma solução. Posso pedir para ela planejar uma funcionalidade, pesquisar opções de implementação ou até mesmo implementar direto — apenas para descartar a implementação e recomeçar tudo do zero com o estado limpo, aproveitando os aprendizados da versão inicial (a famosa "prototipagem").

Leva algumas tentativas para eu acertar o prompt, da mesma forma que levaria algumas tentativas para prototipar algo de forma satisfatória codificando manualmente. A diferença brutal é que, em vez de gastar uma semana por protótipo, costumo levar de 30 minutos a uma hora. Em um único dia, consigo validar de 3 a 4 implementações diferentes para o meu problema e, no fim do dia, estou pronta para me comprometer com uma delas, respaldada por dados concretos para tomar uma decisão informada.

A principal razão para usar a CLI em problemas de baixa certeza técnica é que o ciclo de feedback se fecha quase imediatamente. Você testa sua hipótese, corrige as arestas e itera. O único tempo de espera é o que o modelo leva para processar a sua requisição.

## Certeza Técnica Alta E Baixo/Médio Valor de Negócio = Modo em Lote (Assíncrono)

Um problema com alta certeza técnica, como vimos acima, é aquele em que você já conhece todos (ou quase todos) os passos necessários para implementá-lo. Isso facilita imensamente a sua vida, porque se você já conhece os passos, em vez de executá-los manualmente, pode simplesmente instruir a IA a fazer o trabalho pesado por você.

Aqui teríamos um caso para usar a Gemini CLI no modo YOLO, mas na verdade temos uma ferramenta ainda melhor para isso: o [Jules](https://jules.google/). O Jules foi lançado pelo Google no I/O deste ano e rapidamente se tornou a minha ferramenta favorita de todas (com a Gemini CLI logo atrás em segundo lugar).

O Jules é um agente assíncrono que você pode conectar ao GitHub para executar tarefas em background. Confesso que, ao descobrir o Jules inicialmente, não prestei atenção suficiente nas letras miúdas e fiquei um pouco frustrada com o quão lento ele parecia. Só depois de um tempo compreendi que a proposta central é justamente disparar a tarefa em segundo plano para que você possa se desconectar e seguir com a sua vida.

Como o Jules está integrado ao GitHub, ele já possui todo o contexto do projeto. Você pode pedir tarefas de manutenção como "atualizar a versão das dependências", "implementar testes unitários" ou até "corrigir este bug específico". O ponto fundamental é que, como o ciclo de feedback é longo — você não receberá o resultado na hora —, é melhor reservar essa ferramenta para tarefas nas quais você tem total clareza do que precisa ser feito passo a passo.

## Alta Certeza Técnica E Alto Valor de Negócio = Modo Interativo (Síncrono)

Você deve ter notado que, na matriz acima, destaquei que alta certeza técnica combinada com alto valor de negócio é sempre um processo síncrono. O único motivo para isso é que eu me importo demais com esse resultado (alto valor de negócio), então faço questão de supervisioná-lo de perto e garantir que ele fique pronto o mais rápido possível. Para mim, isso é muito natural, pois o valor de negócio sempre se sobrepõe à certeza técnica quando ambos os parâmetros estão no mesmo nível.

## Baixa Certeza Técnica E Baixo Valor de Negócio = Devo realmente fazer isso?

Esses são os itens que costumam se perder no limbo do backlog. No mundo pré-IA, eu simplesmente os esqueceria. Mas em um mundo habilitado por IA, se me deparo com eles, costumo disparar uma tarefa no Jules para explorar possibilidades na esperança de reduzir incertezas ou alavancar o valor de negócio. Como o custo cognitivo é muito baixo, você literalmente não tem nada a perder abrindo uma tarefa no Jules, mesmo com o prompt mais básico que imaginar. O resultado pode não ser perfeito, mas como você não faria a tarefa de qualquer maneira, qualquer avanço que sair dali já é lucro.

## Exceções

É claro que nenhum bom processo estaria completo sem exceções. Há casos em que delego tarefas de alto valor de negócio para o Jules, e isso geralmente acontece quando eu não teria como realizá-las de outra forma. Por exemplo:
1. Estou em um evento e não consigo abrir o laptop para implementar uma ideia, mas estou com o celular em mãos
2. Estou viajando com uma conexão instável ou lenta
3. Estou presa na fila do supermercado, lembro de algo que deixei pendente e quero dar o pontapé inicial para quando chegar em casa

Em resumo: se as alternativas forem fazer zero progresso ou acionar o Jules, prefiro colocar o Jules para trabalhar por mim.

Por outro lado, ainda não comentei onde as IDEs tradicionais se encaixam no meu fluxo de trabalho. Isso não significa que as abandonei — na verdade, estou escrevendo este post no VS Code agora mesmo. Reservo a IDE e a edição manual para a última milha ou o acabamento final do código. Apenas tome MUITO cuidado com pequenas edições manuais no meio de uma sessão de "vibe coding", pois elas tendem a desorientar o contexto dos LLMs com facilidade. Mas se essa for a etapa final do processo, você estará em total segurança.

## Bônus: Uma nota sobre redução de incerteza

Às vezes você só precisa explorar um assunto, mas ainda nem sabe como integrá-lo à sua base de código. Tentar forçar a Gemini CLI ou o Jules a realizar tarefas que não envolvem código é como usar uma marreta para fixar um parafuso na parede. Nesses cenários de pesquisa pura, prefiro recorrer ao [Gemini Deep Research](https://gemini.google/overview/deep-research/?hl=en-GB). Assim como o Jules, o Gemini Deep Research é assíncrono: você dispara a pesquisa em segundo plano e segue o seu dia.

Se você já estiver diante do teclado e não quiser esperar, usar o Gemini padrão com grounding na Busca do Google também faz maravilhas. Ambas as ferramentas tendem a produzir relatórios detalhados e um tanto extensos. Se você tiver pouca paciência para leituras longas (como eu), um ótimo truque é enviar os resultados da pesquisa para o [NotebookLM](https://notebooklm.google/) e pedir para ele gerar um resumo ou produzir um podcast em áudio para você ouvir no caminho.

## Conclusões

Todo o processo de seleção de ferramentas de IA e o modelo de trabalho baseado no exercício de priorização podem ser resumidos da seguinte forma:

![Resumo das ferramentas recomendadas por prioridade](image-1.png "Resumo das ferramentas e modos de trabalho recomendados")

1. **Alta Certeza Técnica + Alto Valor de Negócio** = processo síncrono ou pair programming com a Gemini CLI. Use o Gemini com grounding na busca para esclarecimentos pontuais.
2. **Baixa/Média Certeza Técnica + Alto Valor de Negócio** = processo síncrono com a Gemini CLI somado a pesquisa assíncrona para aumentar a certeza técnica.
3. **Alta Certeza Técnica + Baixo/Médio Valor de Negócio** = processo assíncrono com o Jules. Deep Research se necessário.
4. **Baixa Certeza Técnica + Baixo Valor de Negócio** = na maioria dos casos, não faça; se você realmente quiser avançar, use o Jules ou Deep Research para elevar um dos parâmetros.

O que você acha dessa metodologia de trabalho? Por favor, compartilhe seus comentários e impressões abaixo!
