---
author: Daniela Petruzalek
categories:
- Agentic Coding
date: 2025-07-03
summary: "Um relato detalhado do meu processo iterativo trabalhando com o Jules, um assistente de programação com IA, para implementar uma nova seção de post em destaque na página inicial do meu blog."
tags:
  - jules
  - tutorial
  - vibe-coding
title: "Como Usei o Jules para Adicionar um Post em Destaque Neste Blog"
slug: "jules-featured-post"
aliases:
  - "/pt-br/posts/20250703-jules-featured-post/"
description: "Estudo de caso sobre o uso do agente Jules para implementar uma seção de post em destaque no Hugo com Tailwind CSS. Aborda iteração de prompts, estilização e aprendizados."
proficiencyLevel: "Intermediate"
dependencies:
  - "Hugo"
  - "Tailwind CSS"
  - "Jules AI Agent"
---

> *Nota da autora:* cerca de 90% deste post foi escrito por IA, mas revisei e editei o texto para garantir uma leitura agradável e fluida. Foi engraçado notar a tendência do Jules de se autoelogiar. Precisei guiá-lo com vários prompts até chegar a este resultado final, embora a última edição tenha sido mais fácil de fazer manualmente. Você pode conferir todo o histórico de edições no [histórico de commits do PR](https://github.com/danicat/danicat.dev/pull/6). Vale mencionar que ele se recusou terminantemente a traduzir este post para o português (Brasil), alegando que não tinha capacidades de tradução — mesmo com a totalidade deste blog tendo sido traduzida usando o Jules em uma interação anterior. Acho que ele só não estava no clima mesmo. :)

## Introdução

Recentemente, decidi atualizar a página inicial do meu blog para destacar melhor o conteúdo mais recente. Como engenheira de backend, mergulhar nas minúcias do frontend não faz parte do meu dia a dia. Por isso, em vez de codificar manualmente todas as mudanças em um domínio com o qual tenho menos familiaridade, recorri à ajuda do [Jules](https://jules.google), um assistente de programação com IA.

Este post detalha a nossa jornada iterativa: os sucessos, os mal-entendidos (alguns bem divertidos) e o que aprendi sobre como trabalhar de forma eficaz com IA no desenvolvimento web, especialmente para preencher lacunas de habilidades.

## O Objetivo: Uma Seção de Post em Destaque

Minha solicitação inicial para o Jules foi bem direta:
> "Change the layout of the main page so that it displays the most recent blog post in highlight instead of it being in the recent posts list. The recent posts should contain all other posts except the most recent one. This behaviour should be seen only on the blog landing page (home). If the user clicks on the Blog menu it should still see all the posts in reverse chronological order, including the most recent one."

O Jules compreendeu rapidamente e propôs um plano envolvendo a exploração da base de código do Hugo, a identificação dos templates e sua modificação.

## Destaques das Iterações: O Bom, o Ruim e a IA

Nossa colaboração envolveu várias iterações para acertar todos os detalhes.

### Iteração 1: Configuração Inicial — Acertando a Base
O Jules identificou corretamente os partials do tema Blowfish e estruturou os overrides. A lógica para separar o post mais recente dos demais na lista de "Posts Recentes" foi muito bem implementada.

*   **O que funcionou:** Compreensão da estrutura central do Hugo, busca de posts e modificações básicas nos templates. A capacidade do Jules de navegar pelos arquivos do tema e do projeto poupou um tempo significativo aqui.

![Tela de aguardo de execução de tarefas no agente Jules](images/image001.png "Houve bastante tempo de espera entre as tarefas")

### Iteração 2: Estilização com Tailwind — A Dança da Tentativa e Erro
Passamos a focar na aparência: título, largura e dimensões da imagem. Isso envolveu uma série de prompts para afinar os visuais. Por exemplo:

> "Change the featured post title to 'Featured Post'. Adjust its width to be about 80% of the view. The image is too tall/narrow, let's try a 4:3 aspect ratio. That's still not quite right, make it wider/less tall."

Foi aqui que a natureza iterativa de trabalhar com o Jules em elementos visuais ficou muito evidente.

*   **A abordagem do Jules:** Modificou arquivos de i18n para os títulos, utilizou diversas classes de largura do Tailwind (por exemplo, `md:w-4/5`, `md:w-2/3`, `max-w-xl`, `max-w-2xl`) e manipulou `padding-bottom` para controlar o aspect ratio das imagens.

*   **Desafio e Frustração:** Um obstáculo específico, especialmente para alguém como eu que trabalha primariamente no backend, foi a natureza de tentativa e erro ao estilizar com Tailwind por meio de um intermediário. Embora o Jules pudesse aplicar as classes que julgava apropriadas, o resultado visual nem sempre correspondia de imediato ao que eu tinha em mente. Mudanças nas classes do Tailwind frequentemente não se traduziam em uma diferença claramente visível na primeira tentativa, ou o efeito não saía como esperado. Isso levou a algumas rodadas de "tente esta classe", "não, faça mais estreito/mais largo/mais alto/mais baixo", o que, embora bem-sucedido no fim das contas, foi um tanto frustrante em certos momentos. Ficou evidente o descompasso entre mexer no código e obter feedback visual imediato nesse fluxo de trabalho assíncrono mediado por IA.

*   **Aprendizado:** Fazer o ajuste fino da estética visual é a parte de que menos gosto nessa experiência, pois as instruções frequentemente geram interpretações imprecisas. Fornecer um feedback claro e descritivo é fundamental, mas também é preciso reconhecer que certo vai-e-vem é inevitável quando não se pode apontar diretamente para a tela ou fazer microajustes em tempo real por conta própria. Ainda assim, o Jules aplicou com dedicação cada alteração solicitada, ajudando a suprir minha menor familiaridade com o frontend.

### Iteração 3: CSS Customizado vs. Tailwind — Um Breve Desvio
A certa altura, para ter um controle muito específico sobre as dimensões do card, passei o seguinte prompt:

> "jules, instead of trying to use an existing style class, create an unique style class for the featured post card. This style should use relative width and height of 75% of the container..."

*   **Resposta do Jules:** O Jules criou corretamente as regras de CSS customizado e refatorou o partial do card para utilizá-las.

*   **Resultado e Aprendizado:** Embora o Jules tenha implementado exatamente o solicitado, o resultado pareceu um tanto estranho em relação ao restante do design do blog, que é fortemente baseado em Tailwind. O CSS customizado não harmonizou muito bem, e rapidamente decidi que manter a consistência com o Tailwind era mais importante. Foi uma ótima lição sobre garantir que soluções geradas por IA respeitem a linguagem de design existente e a preferência por aderir ao framework já estabelecido. O Jules retornou prontamente ao Tailwind quando solicitei:

> "undo the last change and restore the tailwind style of formatting. apply the same style guidelines using tailwind best practices"

### Iteração 4: O Grande Mal-Entendido dos "Comentários"!
Esta foi talvez a parte mais ilustrativa da interação entre humano e IA. Eu mencionei:

> "the comments are rendering in the featured post. please remove all the comments or make them invisible"

*   **A interpretação do Jules:** O Jules presumiu que eu me referia ao *sistema de comentários dos usuários* do blog (como Utterances ou Giscus) ou a metadados como contadores de visualizações/curtidas. Isso gerou uma sequência de etapas em que o Jules tentou investigar e ocultar condicionalmente os metadados de visualizações/curtidas.
*   **Meu esclarecimento:** Após essas mudanças, esclareci fornecendo um exemplo concreto:
    > "you are wrong, I never said I wanted to remove the views and likes - I'm referring to the code comments in rendering as {/* Adjusted padding ... */} and {/* Removed prose classes ... */}"
*   **Resolução:** Assim que o Jules compreendeu que eu me referia a *comentários literais de código em Go template/HTML* incorretamente formatados (usando `{/*...*/}`, que não é uma sintaxe válida de comentário no Hugo e por isso renderizava como texto comum) e não `{{/* ... */}}`, a correção foi imediata: remover o texto indevido dos templates.
*   **O que funcionou:** A persistência e a abordagem sistemática do Jules para depurar o problema (mesmo tendo partido de uma premissa incorreta) foram louváveis.
*   **Desafio e Aprendizado:** O caso evidenciou um aspecto crucial da interação com IA: a ambiguidade na linguagem natural. A palavra "comentários" tem múltiplos significados. Meu relato inicial não havia sido preciso o suficiente.

### Iteração 5: Polimento Final
Após resolvermos os comentários visíveis do template, fizemos os retoques finais:

> "Remove the 'Featured Post' title. Change card width to 50%. Increase title and summary font sizes. Make the image's aspect ratio 16:9."

Isso levou aos ajustes finais na largura do card, tamanhos de fonte e proporção da imagem. A porcentagem de largura em si não surtiu efeito, mas alterar o aspect ratio resolveu com perfeição.

### Iteração Bônus: O Jules Rascunha Este Post

> "This is perfect. No more code changes are needed. Now I want you to create a new blog post entry describing the iteration we just did..."

E cá estamos! Este próprio post foi redigido com a assistência do Jules, a partir do nosso histórico de interações e das minhas orientações de refinamento — incluindo as exatas reflexões que você acabou de ler.

## O Que Funcionou Bem com o Jules

*   **Preenchendo Lacunas de Habilidades:** Como engenheira de backend, o suporte do Jules foi valioso para encarar tarefas de frontend envolvendo templates do Hugo e Tailwind CSS, áreas onde tenho menos experiência no dia a dia. O Jules compensou minha falta de conhecimento aprofundado em frontend, propondo e implementando soluções que coube a mim apenas direcionar e refinar.
*   **Velocidade de Implementação:** Para alterações bem delineadas, o Jules modifica código, cria arquivos e refatora estruturas com muito mais agilidade do que a digitação manual.
*   **Processamento de Instruções Complexas:** De modo geral, o Jules compreendeu solicitações em múltiplas etapas e metas complexas de layout.
*   **Resolução Sistemática de Problemas:** Mesmo diante de mal-entendidos, o Jules frequentemente seguiu uma linha de raciocínio lógica e estruturada.
*   **Refinamento Iterativo:** O Jules se manteve consistentemente receptivo aos feedbacks para ajustes.

## Desafios e Aprendizados

*   **Precisão da Linguagem:** O episódio dos "comentários" ressalta o quão crítica é a precisão na comunicação. O que parece óbvio para um desenvolvedor, ou uma simplificação comum, pode soar ambíguo para uma IA.
*   **Loop de Feedback Visual e Tailwind:** O processo de tentativa e erro ao estilizar com Tailwind foi um desafio central. Como o Jules não "enxerga" o resultado final, descrever o visual desejado ou explicar por que determinado grupo de classes não estava funcionando exigiu paciência e descrições detalhadas. Essa é uma característica intrínseca da interação puramente textual para tarefas visuais.
*   **Interpretação Equivocada e Correção de Rota:** Quando o Jules interpretava mal uma tarefa, ele avançava diligentemente por aquele caminho incorreto. Não havia como interrompê-lo no meio do processo; era necessário aguardar a conclusão da sequência de ações em andamento para fornecer o feedback corretivo.
*   **Fluxo de Trabalho Assíncrono e Ritmo:** O trabalho é predominantemente assíncrono. Cada solicitação e execução do Jules podia levar de alguns minutos até meia hora em sequências mais complexas. Isso torna o ciclo de iteração mais compassado do que programar diretamente com feedback instantâneo ou fazer pair programming ao vivo.

## Recursos Sugeridos

Para quem tiver interesse em conhecer mais sobre o Jules:

*   [Site Oficial do Jules](https://jules.google)
*   [Documentação do Jules](https://jules.google/docs)

## Conclusão

No geral, trabalhar com o Jules para construir o recurso da homepage foi uma experiência produtiva. Teve um gostinho autêntico de "vibe-coding" — uma troca dinâmica conduzindo a IA. O segredo para o sucesso está em manter uma comunicação clara e iterativa, demonstrar paciência diante de mal-entendidos e fornecer feedbacks pontuais e acionáveis.

As frustrações, sobretudo com as tentativas e erros no Tailwind e as eventuais interpretações equivocadas da IA, fazem parte do cenário atual do desenvolvimento assistido por inteligência artificial. Contudo, apesar da natureza assíncrona e do tempo de espera em certas gerações, delegar a parte mecânica da programação e obter sugestões para áreas fora da minha especialidade (como implementações específicas no Tailwind ou na estrutura do Hugo) ainda gerou um saldo extremamente positivo. Foi consideravelmente mais rápido e prático do que se eu tivesse tentado aprender do zero todos os princípios de design de frontend, as peculiaridades do Hugo e os detalhes do Tailwind CSS apenas para entregar essa funcionalidade.

Assistentes de IA como o Jules são ferramentas poderosas. Eles não substituem o julgamento crítico nem a intenção arquitetural do engenheiro, mas se tornam aceleradores extraordinários quando combinados com a mentalidade e a estratégia de comunicação corretas.