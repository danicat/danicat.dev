---categories:
- Agentic Coding
date: '2025-05-21T17:45:07+01:00'
summary: O novo agente autônomo de programação que todo desenvolvedor precisa conhecer.
tags:
  - jules
  - vibe-coding
title: "Precisamos falar sobre o Jules!"
slug: "jules-autonomous-coding-agent"
aliases:
  - "/pt-br/posts/20250521-jules/"
description: "Primeiras impressões sobre o Google Jules, o agente autônomo de código. Veja como ele resolve issues no GitHub, cria pull requests e transforma o fluxo de trabalho."
proficiencyLevel: "Beginner"
---

Olá, pessoal! Vamos falar sobre o Jules! Recém-saído do forno do Google I/O, este é o que o Google está chamando de agente autônomo de programação... mas o que é exatamente um agente autônomo de programação? Pense no [NotebookLM](https://notebooklm.google/), só que para código — uma IA especializada para te ajudar em tarefas de desenvolvimento. A grande diferença em relação à abordagem tradicional de “vibe coding” é que, com o Jules, você pode importar seu projeto inteiro como contexto para a IA, de modo que todas as respostas ficam ancoradas no código real em que você está trabalhando!

## Como o Jules funciona

Assim que o projeto é importado, você passa a interagir com o Jules enviando “tarefas” (tasks), que podem ser qualquer coisa: correções de bugs, atualizações de dependências, novas funcionalidades, planejamento, documentação, testes e muito mais. Ao receber uma tarefa, o Jules planeja sua execução assincronamente em etapas e realiza diferentes subtarefas para garantir o resultado esperado — como verificar se nenhum teste quebrou com a alteração, por exemplo.

Ele se integra diretamente ao GitHub, o que torna a curva de adesão super suave. Ele ainda não substitui a IDE por completo, mas você consegue realizar várias tarefas direto pela interface do Jules, até o ponto em que ele cria uma branch com todas as alterações solicitadas pronta para virar um pull request.

## Primeiras impressões e primeiros testes

A consequência (já esperada) do anúncio de ontem é que a ferramenta está enfrentando uma carga pesadíssima no momento. Por isso, pode demorar um pouco para você ver o resultado após enviar uma tarefa. Mesmo assim, o Jules faz todo o trabalho em segundo plano e, se você ativar as notificações do navegador, ele te avisa assim que terminar.

Por conta dessa fila, ainda não consegui fazer nenhum experimento monumental com ele, mas uma das primeiras coisas que testei foi gerar o [README do repositório deste blog no GitHub](https://github.com/danicat/danicat.dev/pull/1) (o código-fonte desta página em que você está agora). Também arrisquei algumas iterações mais complexas, como ajustar o template do blog. [Ele gerou os arquivos corretos](https://github.com/danicat/danicat.dev/pull/2), mas estava respondendo de forma um pouco lenta às requisições, então acabei fazendo alguns ajustes finos manualmente.

## Uma nova era para as IDEs

Para um primeiro dia, o resultado foi impressionante, e há um potencial gigantesco a ser explorado nas próximas semanas e meses. O grande diferencial é a capacidade de trabalhar sobre uma base de código inteira, aposentando aquele fluxo cansativo de perguntar algo ao Gemini (ou ChatGPT), copiar o código para a IDE, rodar, copiar os erros de volta para o LLM e iterar. Claro, ferramentas como Code Assist e GitHub Copilot trazem parte desses recursos para dentro da IDE, mas ainda sinto que a IDE tradicional não é o ambiente ideal para o vibe coding — muitas vezes parece um remendo (um hack improvisado).

Nesse sentido, talvez o Jules seja a dose de inspiração que faltava para uma nova era de ambientes de desenvolvimento que destravará o potencial da IA para pessoas desenvolvedoras do mundo todo de forma muito mais natural. Pelo menos é por isso que estou torcendo!

## Como experimentar

O Jules está atualmente em beta público e você já pode experimentar se cadastrando em [https://jules.google](https://jules.google).
