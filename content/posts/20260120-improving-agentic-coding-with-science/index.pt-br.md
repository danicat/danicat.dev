---
categories:
- Agentic Coding
date: 2026-01-21 00:00:00+00:00
draft: false
summary: Domar agentes de IA exige ciência, não apenas vibes. Descubra como testes
  A/B e rigor estatístico transformam agentes de programação em uma disciplina de
  engenharia mensurável.
tags:
  - agents
  - ai
  - gemini-cli
  - golang
  - mcp
  - vibe-coding
title: Melhorando Agentes de Programação com Ajuda da Ciência
---

A transição do [determinismo para o não-determinismo](https://newsletter.pragmaticengineer.com/p/martin-fowler) no desenvolvimento de software é um dos maiores desafios que enfrentamos como indústria hoje.

Todo mundo já passou por isso: você está trabalhando com um agente de IA e pensa: *"E se eu adicionar isso ao meu system prompt?"* ou *"Talvez eu devesse dar a ele acesso a esta nova ferramenta MCP?"*

Você faz a alteração, executa uma vez e funciona. Sucesso! Você se sente genial e compartilha a novidade com colegas de trabalho, seguidores, sua avó e seu gato. Mas aí, uma hora depois, você tenta de novo e ele falha miseravelmente na mesmíssima tarefa. A mudança foi realmente boa ou você só deu sorte na primeira tentativa? "Talvez eu não tenha feito direito desta vez..." — desliga e liga de novo, tenta mais uma vez e funciona... ou será que não?

O fato é que não há apenas uma aleatoriedade inerente à forma como os LLMs funcionam, mas também diversos fatores de confusão que podem contribuir para esses resultados.

Como vivemos no mundo determinístico por tanto tempo (será mesmo?), nós, na engenharia de software, não estamos com a mente preparada para lidar com esse nível de incerteza. Esperamos que a computação seja exata, já que nos dizem desde o início que ela é uma [ciência exata](https://pt.wikipedia.org/wiki/Ci%C3%AAncias_exatas).

## A vibe vs. a ciência

Falo muito sobre [vibe coding com disciplina]({{< ref "/posts/20251206-taming-vibe-coding" >}}) ou, em outras palavras, como adicionar metodologia ao vibe coding é essencial para alcançar resultados de qualidade. Este não é um problema novo — lidamos com desafios de qualidade de software há décadas —, mas a redução da barreira de entrada e o aumento expressivo na velocidade de produção de código elevaram a questão a uma nova ordem de magnitude.

Sempre me surpreendo quando "descubro" que uma metodologia antiga, criada para ajudar times de humanos a gerarem melhores resultados, costuma funcionar incrivelmente bem com agentes de IA. Mas, na verdade, isso não deveria surpreender ninguém: afinal, seres humanos também são não-determinísticos por natureza. A IA está apenas amplificando padrões bem conhecidos.

Embora eu tenha conseguido estruturar um fluxo de trabalho razoável com base nesses princípios, nunca cheguei a ter a segurança de afirmar "este é o jeito CERTO de fazer isso". E isso acontece porque aquela falha ocasional ou regressão inesperada corrói a minha confiança — "Será que este `GEMINI.md` é mesmo o prompt definitivo para quem desenvolve em Go?" "Será que cheguei às melhores instruções de sistema para esta tarefa?" "Minhas ferramentas MCP são as melhores APIs que eu poderia criar?" — todos os dias sou inundada por tantas dúvidas que, na maior parte do conteúdo que publiquei no ano passado, evitei ser prescritiva e posicionei tudo como **estudos de caso** — "Eu fiz _isso_", "aconteceu _aquilo_". Fim de papo. Na literatura científica, estudos de caso representam um dos [níveis mais baixos de evidência](https://pt.wikipedia.org/wiki/Medicina_baseada_em_evid%C3%AAncias).

Sei que este não é o argumento típico em um blog técnico, então preciso abrir um pequeno parêntese sobre a minha vida anterior: antes de me firmar na carreira de engenheira de software, cursei a faculdade de Medicina (nunca cheguei a me formar, mas essa história fica para outro momento). Na Medicina e em outras áreas biológicas, as pessoas estão muito mais acostumadas à experimentação e a lidar com vieses, ruído e aleatoriedade, pois precisam extrair "verdades" de sistemas de origem desconhecida (basicamente fazendo engenharia reversa do mundo, o que é fascinante!). Não basta apresentar os dados de um estudo; é preciso validá-los estatisticamente para assegurar a eliminação de qualquer contaminação potencial. Tamanho da amostra, alfa, p-valores, teste t de Student... Eu odiava esses termos nas provas, mas mal sabia como eles seriam úteis hoje.

É claro que isso não é exclusivo da Biologia. Na engenharia, aplicamos métodos estatísticos em pesquisas com frequência, mas sinto que isso não é tão disseminado na nossa área quanto em outras. Quando trabalhei com sistemas de recomendação, por exemplo, os testes A/B eram uma parte crítica do trabalho para otimizar os algoritmos de machine learning. Especialistas em pesquisa de UX também utilizam testes A/B extensivos para determinar quais interfaces performam melhor, e assim por diante.

Eu queria encontrar uma maneira de eliminar o "achismo" dos meus experimentos com agentes de programação. Percebi que a melhor saída seria criar um framework de experimentação para coletar dados e rodar análises estatísticas, o que me permitiria sair do "acho que funciona" para o "**sei** que funciona (_com 95% de confiança_)".

## Apresentando o Tenkai: o framework de experimentação para agentes

Para resolver esse desafio, construí o **Tenkai** (palavra japonesa para "desdobramento" ou "expansão", um pequeno easter egg para quem curte animes). Trata-se de um framework em Go projetado para avaliar e testar diferentes configurações de agentes de programação com rigor estatístico.

![Visão de experimento no Tenkai](image-3.png "Interface do Tenkai com um experimento em execução")

Pense nele como um laboratório para seus agentes de IA. Em vez de rodar um prompt uma única vez e torcer pelo melhor, o Tenkai permite executar experimentos que repetem as mesmas tarefas diversas vezes (até um número N de repetições — o tamanho da sua amostra) e compara alternativas (diferentes conjuntos de configurações) entre si por meio de testes estatísticos.

Digamos que a Alternativa A seja a sua configuração padrão (o "controle" do experimento) e a Alternativa B seja um novo prompt de sistema que você queira testar. O experimento executará ambas N vezes e gerará um relatório indicando se B é significativamente mais rápida, mais eficiente ou mais precisa. Se a diferença decorrer de mera coincidência ou ruído, o framework não a classificará como significativa, e você poderá descartar os resultados com tranquilidade.

### Como funciona

O fluxo de trabalho no Tenkai:

1.  **Definir cenários:** Um cenário é uma tarefa de programação padronizada (por exemplo, "Corrigir um bug neste pacote Go" ou "Implementar um novo componente React"). Ele inclui regras de validação como "Compilou?" ou "Os testes passaram?". Para que um cenário seja considerado bem-sucedido, ele precisa atender a todos os critérios de validação especificados.

![Visão de lista de cenários](image-2.png "Cenários são suas tarefas de programação. Você pode adicionar um ou mais cenários aos seus experimentos e todos serão validados com os mesmos critérios")

2.  **Criar um template:** Você define o que quer testar. Por exemplo, sua "Alternativa A" (controle) pode ser o modelo Gemini 2.5 Flash, e sua "Alternativa B" pode ser o Gemini 2.5 Flash com um servidor MCP configurado. É possível incluir até 10 alternativas no mesmo template de experimento (todas serão comparadas com o controle).

![Editor de alternativas](image-1.png "O editor de alternativas permite sobrescrever o comando do agente, flags de linha de comando, system prompt, GEMINI.md e settings.json")

3.  **Executar o experimento:** O Tenkai executa cada cenário múltiplas vezes para cada alternativa. Ele isola as execuções em workspaces temporários, gerencia timeouts e registra cada evento — de chamadas de ferramentas a saídas do shell. Você também pode definir um nível de simultaneidade para rodar tarefas em paralelo, sem precisar esperar uma eternidade pelos resultados.

### Rigor estatístico (A parte da ciência)

Estes são os testes estatísticos atualmente integrados ao framework:

*   **Teste t de Welch**: Para métricas contínuas como duração, consumo de tokens ou quantidade de problemas de lint. Usamos a versão de Welch em vez do teste t padrão porque ela não assume variância igual entre os grupos — o que é essencial ao comparar modelos com perfis de performance bem discrepantes.
*   **Teste Exato de Fisher**: Para taxas de sucesso. Ao trabalhar com tamanhos de amostra menores (como 10 ou 20 repetições), os testes qui-quadrado padrão podem perder precisão. O teste de Fisher nos fornece um p-valor confiável para determinar se uma taxa de sucesso de 80% é realmente superior a uma de 60%, ou se foi apenas uma sequência de sorte.
*   **Teste U de Mann-Whitney**: É o nosso cavalo de batalha não paramétrico. Nós o usamos para comparar a *distribuição* de chamadas de ferramentas entre execuções bem-sucedidas e com falha. Como as contagens de tool calls não seguem uma distribuição normal (muitos zeros!), o Mann-Whitney ajuda a identificar se uma ferramenta específica está sendo acionada de forma significativamente maior nas execuções vencedoras.
*   **Rho de Spearman (Correlação)**: Usamos para identificar os "Determinantes de Sucesso". Ao calcular a correlação entre chamadas de ferramentas específicas e métricas como duração ou tokens, o Tenkai consegue apontar se uma nova ferramenta MCP atua como um **Impulsionador de Sucesso** ou se é apenas uma distração que infla os custos.

## Insights em tempo real e o dashboard

Conforme o experimento é executado, você pode acompanhar a linha de raciocínio do agente, ver quais ferramentas ele está acionando e identificar onde ele trava. É como poder "olhar dentro da mente" do seu agente ao longo de dezenas de execuções simultâneas.

Um dos recursos mais poderosos do dashboard é a capacidade de filtrar a análise em tempo real sob três "lentes" distintas:

*   **Todas as Execuções:** A verdade nua e crua. Inclui cada timeout crítico e erro de sistema. É a métrica primária da **confiabilidade** geral do sistema.
*   **Apenas Completas:** Filtra execuções que atingiram um estado terminal (Sucesso ou Falha de Validação). É onde você avalia métricas de **qualidade**, como problemas de lint ou tempo de execução, removendo o ruído de timeouts externos.
*   **Apenas Sucessos:** A visão "Padrão Ouro". Ao olhar exclusivamente para as execuções vencedoras, você começa a deduzir *por que* elas deram certo. É aqui que calculamos o **Rho de Spearman** e os p-valores de **Mann-Whitney U** para descobrir quais ferramentas estão altamente correlacionadas com o sucesso.

## Primeiros resultados dos meus próprios experimentos

Tenho usado o Tenkai para refinar o [godoctor](https://github.com/danicat/godoctor) e transformá-lo no meu servidor MCP dos sonhos para desenvolvimento em Go. Minha hipótese é que, ao fornecer ferramentas especializadas aos modelos, eles se tornarão mais eficazes na execução de tarefas de código. Por exemplo: em vez de dar aos modelos a liberdade de decidir quando ler a documentação para descobrir a API de uma biblioteca cliente, estou entregando a documentação diretamente a eles sempre que o modelo chama `go get`. Isso preveniu massivamente alucinações de API e aqueles "loops de inferno de dependências", em que o modelo fica batalhando com `go get` e `go mod` achando que obteve a versão errada do pacote.

Os primeiros resultados também mostram que essa diferença diminui se eu fixar a versão do modelo na geração mais recente — a Gemini CLI por padrão é iniciada no modo "auto", o que significa que ela decide automaticamente qual modelo chamar entre o Gemini 2.5 e o Gemini 3 (tanto versões Flash quanto Pro). Ao fixar a versão para o Gemini 3 Pro (ID do modelo `gemini-3-pro-preview`), ele se torna muito mais inteligente e frequentemente busca a documentação por conta própria (rodando `go doc` na linha de comando) nesses cenários de conflito, tornando o conjunto de ferramentas do godoctor menos impactante.

![Tabela de resumo do experimento comparando a configuração padrão com duas configurações do godoctor (com e sem ferramentas principais)](image.png "Ter o godoctor ativado em paralelo com as ferramentas principais não teve impacto significativo e potencialmente desperdiçou tokens (p < 0,1), mas após desativar as ferramentas principais — forçando o uso do godoctor —, a duração média da tarefa reduziu em mais de 30% (p < 0,01)")

Também enfrentei muitos desafios com a adoção de ferramentas. Os modelos têm, por padrão, uma forte preferência por usar as ferramentas nativas com as quais foram treinados. Mesmo que eu forneça uma ferramenta "mais inteligente", eles têm dificuldade em sair da sua zona de conforto. Errei bastante tentando projetar experimentos em torno disso até perceber que meu processo de experimentação estava completamente falho. **Adoção de ferramentas** e **eficácia de ferramentas** são conceitos ortogonais: ao tentar testar ambos no mesmo experimento, eu não conseguia produzir resultados válidos. Decidi pivotar para testar apenas a eficácia da ferramenta, bloqueando o acesso do modelo às ferramentas nativas. Dessa forma, finalmente comecei a obter sinais melhores sobre o quão boas (ou ruins) são as ferramentas do godoctor.

Se você tem curiosidade sobre a evolução do godoctor e como cheguei à API atual, darei uma palestra sobre o tema na [FOSDEM no próximo dia 1º de fevereiro](https://fosdem.org/2026/schedule/event/3BD3Z9-making_of_godoctor_an_mcp_server_for_go_development/). E se você não puder comparecer, não se preocupe: também escreverei mais sobre isso nas próximas semanas.

## Conclusões

Estamos em um ponto de virada no desenvolvimento de software. Estamos migrando de um mundo onde escrevemos cada linha de código para um cenário onde **orquestramos** inteligência.

Mas a orquestração requer medição. Não podemos melhorar o que não podemos medir. A transição de ser apenas uma escritora de código para uma orquestradora de inteligência não significa que trabalhamos menos — significa que fazemos um trabalho **diferente**. Nossa responsabilidade primária não é mais apenas a sintaxe; é o **Contexto, as Ferramentas e as Diretrizes (Guardrails)**.

Se não medirmos o impacto de nossas mudanças com rigor, não estamos fazendo engenharia; estamos apenas jogando com a sorte. Ao avançar para uma abordagem baseada em evidências para agentes de programação, podemos finalmente construir sistemas que não sejam apenas "legais" quando funcionam, mas confiáveis o suficiente para sustentar um negócio.

Se você tiver interesse em conferir o código ou realizar seus próprios experimentos, pode encontrar o projeto [aqui](https://github.com/danicat/tenkai).

Bons experimentos! o/

Dani =^.^=