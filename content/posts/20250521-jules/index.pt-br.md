+++
date = '2025-05-21T17:45:07+01:00'
title = 'Precisamos falar sobre o Jules!'
summary = "O novo agente de codificação autônomo que todo desenvolvedor precisa conhecer."
tags = ["vibe coding"]
+++
{{< translation-notice >}}
Olá a todos, vamos falar sobre o Jules! Saído direto do forno do Google I/O, isto é o que o Google chama de um agente de codificação autônomo… mas o que é um agente de codificação autônomo? Pense no [NotebookLM](https://notebooklm.google/), mas para codificação - uma IA especializada para ajudá-lo com tarefas de programação. A principal diferença da abordagem tradicional de “vibe coding” é que com o Jules você pode importar todo o seu projeto como contexto para a IA, então todas as respostas são baseadas no código em que você está realmente trabalhando!

Depois que o projeto é importado, você pode interagir com o Jules enviando “tarefas”, que podem ser qualquer coisa, desde correções de bugs, atualizações de dependência, novos recursos, planejamento, documentação, testes e assim por diante. Assim que recebe uma tarefa, o Jules planejará assincronamente sua execução em etapas e realizará diferentes subtarefas para garantir que o resultado desejado seja alcançado. Por exemplo, garantindo que nenhum teste foi quebrado pela nova alteração.

Ele se integra diretamente com o GitHub, então há muito pouco atrito para começar a usá-lo. Ele ainda não substituirá completamente o IDE, mas você pode realizar muitas tarefas diretamente do Jules até o ponto em que ele cria um branch com todas as alterações solicitadas, pronto para ser transformado em um pull request.

A consequência infeliz do anúncio do Jules ontem é que a ferramenta está atualmente sob forte carga, então pode levar um tempo depois que você envia uma tarefa para ver os resultados, mas o Jules fará o trabalho em segundo plano e se você tiver as notificações do navegador ativadas, ele o avisará quando estiver pronto.

Diante disso, não consegui fazer grandes experimentos com ele, mas uma das coisas que fiz foi gerar o [README para o projeto do meu blog no Github](https://github.com/danicat/danicat.dev/pull/1) (a fonte desta mesma página que você está lendo agora). Também tentei algumas iterações mais complexas, como ajustar o template do blog. [Ele gerou os arquivos corretos](https://github.com/danicat/danicat.dev/pull/2), mas demorou um pouco para responder às minhas solicitações, então tive que fazer algumas alterações manualmente.

Nada mal para o primeiro dia, eu diria, e há muito potencial a ser desbloqueado nas próximas semanas e meses. O recurso matador é a capacidade de trabalhar em uma base de código completa, em vez daquele fluxo tradicional de fazer uma pergunta ao Gemini (ou ChatGPT), copiar o código-fonte para o IDE, executar, copiar e colar de volta os resultados no LLM e iterar. Claro, ferramentas como Code Assist e CoPilot fornecerão algumas dessas capacidades sem sair do IDE, mas ainda sinto que o IDE não é o ambiente certo para o vibe coding, pois parece mais um hack.

Nesse espírito, talvez o Jules seja a injeção de inspiração que precisávamos para uma nova era de IDEs que desbloqueará o potencial da IA para desenvolvedores em todo o mundo de uma forma mais natural. Pelo menos é o que espero!

O Jules está atualmente em beta público e você pode brincar com ele hoje inscrevendo-se em [https://jules.google](https://jules.google).
