---
categories:
- Agentic Coding
date: '2025-05-21T17:45:07+01:00'
summary: O novo agente autônomo de programação que toda pessoa desenvolvedora precisa conhecer.
tags:
  - jules
  - vibe-coding
title: "Precisamos falar sobre o Jules!"
slug: "jules-autonomous-coding-agent"
aliases:
  - "/pt-br/posts/20250521-jules/"
description: "Primeiras impressões sobre o Google Jules, o agente autônomo e assíncrono de programação. Veja a resolução de issues no GitHub, geração automatizada de PRs e o fluxo de trabalho."
proficiencyLevel: "Beginner"
---
Olá, pessoal! Vamos falar sobre o Jules! Recém-saído do forno do Google I/O, este é o que o Google está chamando de agente autônomo de programação… mas o que é um agente autônomo de programação? Pense no [NotebookLM](https://notebooklm.google/), só que para código, uma IA especializada para te ajudar em tarefas de programação. A principal diferença em relação à abordagem tradicional de “vibe coding” é que, com o Jules, você pode importar seu projeto inteiro como contexto para a IA, de modo que todas as respostas ficam ancoradas no código real em que você está trabalhando!

## Como o Jules funciona

Assim que o projeto é importado, você pode interagir com o Jules enviando “tarefas” (tasks), que podem ser qualquer coisa: correções de bugs, atualizações de dependências, novas funcionalidades, planejamento, documentação, testes e muito mais. Assim que recebe uma tarefa, o Jules planeja sua execução assincronamente em etapas e realiza diferentes subtarefas para garantir o resultado esperado. Por exemplo, garantir que nenhum teste seja quebrado pela nova alteração.

Ele se integra diretamente ao GitHub, então há pouquíssimo atrito para começar a usar. Ele ainda não vai substituir a IDE por completo, mas você pode realizar muitas tarefas diretamente pelo Jules, até o ponto em que ele cria uma branch com todas as alterações solicitadas, pronta para virar um pull request.

## Primeiras impressões e primeiros testes

A consequência infeliz do anúncio do Jules ontem é que a ferramenta está sob carga pesadíssima no momento, então pode demorar um pouco para você ver os resultados após enviar uma tarefa. Mas o Jules faz o trabalho em segundo plano e, se você ativar as notificações do navegador, ele vai te avisar assim que estiver pronto.

Diante disso, não consegui fazer nenhum experimento grande com ele, mas uma das coisas que fiz foi gerar o [README do projeto do meu blog no GitHub](https://github.com/danicat/danicat.dev/pull/1) (o código-fonte desta mesma página que você está lendo agora). Também tentei algumas iterações mais complexas, como ajustar o template do blog. [Ele gerou os arquivos corretos](https://github.com/danicat/danicat.dev/pull/2), mas demorou um pouco para responder aos pedidos, então precisei fazer algumas alterações manualmente.

## Uma nova era para as IDEs

Nada mal para o primeiro dia, eu diria, e há muito potencial a ser destravado nas próximas semanas e meses. O recurso matador é a capacidade de trabalhar em uma base de código completa, em vez daquele fluxo tradicional de fazer uma pergunta ao Gemini (ou ChatGPT), copiar o código para a IDE, rodar, copiar e colar o resultado de volta no LLM e iterar. Claro que ferramentas como o Code Assist e o Copilot oferecem algumas dessas funcionalidades sem você sair da IDE, mas ainda sinto que a IDE não é o ambiente ideal para o vibe coding, parecendo mais um hack.

Nesse espírito, talvez o Jules seja a injeção de inspiração de que precisávamos para uma nova era de IDEs que vai destravar o potencial da IA para desenvolvedores de todo o mundo de forma muito mais natural. Pelo menos é por isso que estou torcendo!

## Como começar

O Jules está atualmente em beta público e você pode testá-lo hoje mesmo se inscrevendo em [https://jules.google](https://jules.google).
