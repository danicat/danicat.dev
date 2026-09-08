---
categories:
- Agentic Coding
date: 2026-05-15 12:00:00+00:00
heroStyle: big
summary: As transformações Ágeis ficaram obsoletas? Entenda como as práticas tradicionais
  do Ágil se traduzem em fluxos de trabalho agênticos e como escalá-las por toda a
  organização.
tags:
  - agile
  - gemini-cli
  - mcp
  - software-engineering
title: "Do Ágil ao Agêntico: Um Guia para o Desenvolvimento Corporativo Moderno"
slug: "from-agile-to-agentic"
aliases:
  - "/pt-br/posts/20260515-from-agile-to-agentic/"
description: "Como traduzir práticas Ágeis em fluxos de trabalho agênticos. Aborda escrita de histórias como prompts, arquitetura conversacional com MCP e agentes não programadores."
proficiencyLevel: "Intermediate"
dependencies:
  - "Gemini CLI / Antigravity CLI"
  - "Model Context Protocol"
---

Se você passou algum tempo significativo na indústria de software na última década, muito provavelmente já vivenciou uma transformação Ágil. Já participou de reuniões intermináveis de planejamento de sprint, compartilhou atualizações em daily standups e talvez até tenha se perguntado se tudo aquilo fazia algum sentido.

Essa frustração é comum e geralmente nasce do descompasso entre a gestão de produtos e a realidade da engenharia. Lideranças de negócio buscam previsibilidade, mas o desenvolvimento de software é inerentemente imprevisível. Quando as empresas tentam forçar previsibilidade por meio de métricas rígidas e painéis de controle, o Ágil falha. A metodologia se converte em pura burocracia. As pessoas desenvolvedoras ficam frustradas porque as cerimônias parecem mero overhead inútil, e práticas como a escrita de histórias viram tarefas maçantes.

Em implementações Ágeis maduras, por outro lado, a liderança reconhece a incerteza e dá autonomia aos times para tomarem as melhores decisões para o negócio. Métricas e dashboards deixam de ser o objetivo principal; em vez disso, o foco se volta para a entrega de valor de negócio. Isso não significa deixar tudo nas mãos dos desenvolvedores, mas trabalhar em estreita colaboração com eles em times verdadeiramente multidisciplinares.

Esse preâmbulo é necessário porque, para aproveitar ao máximo este artigo, quero que você se reconecte com as raízes do Manifesto Ágil. Lembre-se dos [valores fundamentais](https://agilemanifesto.org/):

> Indivíduos e interações mais que processos e ferramentas  
> Software em funcionamento mais que documentação abrangente  
> Colaboração com o cliente mais que negociação de contratos  
> Responder a mudanças mais que seguir um plano

Neste artigo, vamos explorar como podemos transportar essas práticas Ágeis fundamentais diretamente para a era moderna dos fluxos de trabalho agênticos.

## O fluxo de trabalho de desenvolvimento agêntico

Antes de falarmos sobre escala corporativa, precisamos estabelecer uma base sólida. Já abordei esses tópicos em artigos anteriores, mas vale a pena resumi-los aqui para contextualizar:

1.  **Escrever histórias é fazer prompting:** A parte mais difícil ao comandar um agente é fornecer o contexto correto. Desenvolvedores frequentemente têm dificuldade em redigir boas histórias. No entanto, essa é a habilidade mais importante que você pode cultivar hoje. A capacidade necessária para escrever uma boa história é exatamente a mesma exigida para criar um bom prompt: estruturar informações de forma clara e fácil de consumir, com justificativa de negócio evidente e resultados esperados bem definidos. No jargão Ágil, isso é o bom e velho **Definition of Ready (DoR)**.
2.  **A priorização dita o fluxo de trabalho:** O refinamento tradicional de backlog se traduz diretamente em como gerenciamos nossas ferramentas de IA. **Ao priorizar** tarefas com base nas dimensões de *Valor de Negócio* versus *Certeza Técnica*, você decide o que trabalhar de forma síncrona em primeiro plano (em pair programming com ferramentas como a [Gemini CLI](https://geminicli.com/)) e o que delegar para agentes assíncronos em segundo plano (como o [Jules]({{< ref "/posts/20250521-jules" >}})).
3.  **O ciclo de codificação agêntica:** Agentes de codificação são extremamente poderosos, mas costumam pecar na consistência. Eles são não determinísticos por definição. Podemos mitigar esse problema com o uso de [práticas determinísticas de engenharia]({{< ref "/posts/20251206-taming-vibe-coding" >}}). Costumo descrever isso como "reduzir a agência do agente". Por exemplo: se você sabe que seu processo de build envolve invariavelmente compilar, rodar testes, rodar linter e fazer deploy, você não deve especificar isso como um prompt de texto. O agente inevitavelmente esquecerá um ou mais passos no decorrer da sessão. O que você realmente deve fazer é empacotar esse processo como uma ferramenta customizada e fornecê-la ao agente, eliminando completamente qualquer margem para que ele se esqueça de executar etapas críticas.

Com essas três práticas dominadas, você já é uma pessoa desenvolvedora agêntica eficaz. Mas como tornar toda uma organização de engenharia eficiente?

## Arquitetura conversacional e compartilhamento de conhecimento institucional

Desenvolver software corporativo é difícil, mas lembrar o *porquê* de determinado código ter sido escrito pode ser tão desafiador quanto. O conhecimento tácito (*tribal knowledge*) se dissipa rápido. Quando uma pessoa engenheira sênior sai da empresa, seu conhecimento institucional costuma ir embora junto. A resposta tradicional para esse problema sempre foi manter documentações exaustivas em wikis internas — que, na prática, tendem a ser difíceis de manter, descobrir e fazer cumprir.

Por muitos anos, um dos meus artigos favoritos sobre compartilhamento de conhecimento corporativo tem sido este texto no blog de Martin Fowler: [Scaling Architecture Conversationally](https://martinfowler.com/articles/scaling-architecture-conversationally.html). Os autores defendem que boa arquitetura se dissemina por meio de conversas, e não apenas por imposições verticais (*top-down*). O artigo também explora como formalizar essas conversas em Registros de Decisão de Arquitetura (ADRs, do inglês *Architecture Decision Records*), impedindo que se percam no tempo.

ADRs vão muito além de wikis simples: eles registram um snapshot histórico e pontual do momento em que a decisão foi tomada. Capturam as condições, premissas e restrições específicas que a justificaram. Essa ideia pode parecer simples à primeira vista, mas dá autonomia para que equipes futuras façam alterações quando necessário. Como existe um registro de *por que* a escolha original foi feita, o time tem as ferramentas necessárias para avaliar se aquela decisão ainda se sustenta ou se pode ser revogada com segurança (emitindo um novo ADR) conforme o cenário evolui.

Com o passar dos anos, acredito cada vez mais que a parte mais importante do nosso trabalho é gerenciar incertezas. ADRs são uma das ferramentas que nos permitem ser honestos sobre o que sabemos e o que não sabemos. Quanto antes percebermos que não há problema em não saber tudo de antemão, melhor. Esse é o coração do Ágil: precisamos saber apenas o suficiente para avançar, acumular aprendizados para reduzir a incerteza e iterar. Software é um organismo vivo: ele nunca está concluído.

Embora os ADRs tragam inúmeras vantagens sobre uma wiki não estruturada, eles ainda compartilham uma grande falha: a dependência de que as pessoas humanas estejam cientes deles e os sigam. Especialmente em grandes organizações, a comunicação vira o gargalo. Silos de informação se espalham e um esforço enorme é gasto apenas sincronizando diferentes partes do negócio.

### Distribuir conhecimento através de agentes

Para escalar uma arquitetura hoje, precisamos injetar esse conhecimento institucional diretamente nos agentes. Em vez de depender apenas de canais oficiais de comunicação entre pessoas, podemos usar tecnologia para transmitir regulamentações, ADRs, procedimentos de governança e padrões corporativos diretamente aos nossos agentes. Quando o contexto organizacional vive dentro da janela de contexto do agente, você garante que essas práticas sejam sempre aplicadas e estejam atualizadas.

Do ponto de vista arquitetural, um servidor MCP é o veículo ideal para expor esse tipo de informação. Ele pode ser gerenciado de forma centralizada e atualizado sempre que um comitê de arquitetura, segurança ou governança tomar uma decisão. Prompts, ferramentas e skills são formas eficazes de moldar o comportamento do agente, e podem ser consumidos tanto pelos agentes de codificação nas mãos das pessoas desenvolvedoras quanto por agentes automatizados em pipelines de CI/CD.

É uma pena que Agent Skills ainda não façam parte da especificação oficial do MCP, mas já existe um grupo de trabalho dedicado a isso. Assim que pudermos usar servidores MCP para transmitir skills diretamente aos agentes, o desafio de mantê-las atualizadas estará resolvido, reduzindo o atrito para disseminar novos padrões entre as pessoas desenvolvedoras.

### Documentação de produto também é um produto consumível

Além das regras internas, esse exato mesmo mecanismo se aplica à documentação de produto. Tradicionalmente, quando o Time A cria uma API interna, ele publica um arquivo de especificação OpenAPI em um portal de desenvolvedores e espera que o Time B leia o manual. Na era agêntica, a documentação estática cria atrito. Se o seu produto foi feito para ser consumido por outros times, a documentação dele deve ser consumível pelas ferramentas desses times.

Quando o Time A lançar seu serviço, ele também deveria disponibilizar um servidor MCP dedicado que exponha o schema da API, exemplos de integração e verificações de conformidade na forma de ferramentas. Quando alguém do Time B precisar fazer a integração com o serviço, basta conectar seu agente de codificação ao servidor MCP do Time A. O agente pode consultar a estrutura da API, ler as regras de integração e escrever o código do cliente automaticamente. Migramos de humanos lendo manuais para agentes lendo APIs, garantindo que a intenção arquitetural e os padrões de integração sejam perfeitamente preservados por toda a empresa.

## Automatizando cerimônias: agentes não programadores

Embora passemos muito tempo falando sobre agentes de codificação, existem inúmeras oportunidades de otimização usando agentes não programadores (*non-coding agents*) para reduzir a sobrecarga de gestão que frequentemente afeta a maioria das implementações Ágeis.

Desde tarefas mais simples, como atas automáticas de reuniões e sumarização, até a repriorização de backlog, refinamento de histórias e criação de spikes técnicos, podemos usar agentes não programadores para recuperar muito do esforço gasto em tarefas administrativas e redirecioná-lo para a engenharia.

Aqui estão algumas formas pelas quais esses agentes focados em processos podem elevar o nível do time:

*   **Refinamento e quebra de histórias:** Se a pessoa responsável pelo produto (Product Owner) redige o rascunho de um épico, um agente pode revisá-lo para identificar cenários de borda esquecidos, premissas técnicas implícitas e fluxos de erro não tratados. Dê a ele acesso a skills específicas e ele estará automaticamente em conformidade com os padrões organizacionais. Pontos de incerteza podem se transformar automaticamente em tickets de spike para investigação técnica mais aprofundada.
*   **Auditoria de Definition of Ready e Definition of Done:** Em muitas configurações Ágeis, o DoR e o DoD são tratados como meros checklists em wikis que são frequentemente esquecidos. Podemos tornar essa conformidade proativa integrando agentes aos quadros Kanban existentes (como Jira ou GitHub Projects). Quando um ticket é movido para "Ready for Dev", um agente em background pode escaneá-lo para garantir que todo o contexto necessário, como schemas de API e mockups de UI, esteja realmente anexado. Se não estiver, ele sinaliza a transição. Da mesma forma, antes de fechar um ticket, um agente pode verificar se testes foram adicionados e se a documentação foi atualizada.
*   **Retrospectivas orientadas a dados:** Retrospectivas costumam sofrer de viés de recência. Agentes não programadores podem atuar como analistas de dados imparciais, revisando o histórico de movimentação dos tickets na sprint, comentários em pull requests e conversas em canais de chat. Por exemplo: o agente pode apontar que tickets relacionados a um microsserviço específico levaram em média quatro dias em revisão, incentivando o time a investigar se existe um silo de conhecimento que precisa ser resolvido.

## Escalando o fluxo de trabalho com Gerenciadores de Agentes

Ao longo do último ano, a indústria se concentrou fortemente no aprimoramento da experiência com agente único — especialmente os agentes de codificação — e vimos novos padrões surgirem e se consolidarem (MCP, skills, hooks, etc.).

Como consequência, a competência central da carreira de engenharia de software migrou da escrita direta de código para a orquestração de agentes. No entanto, existe um gargalo oculto aqui: a carga cognitiva humana. Já existem relatos de um [novo tipo de burnout causado pelo uso de IA](https://techcrunch.com/2026/02/09/the-first-signs-of-burnout-are-coming-from-the-people-who-embrace-ai-the-most/).

Delegar tarefas para agentes assíncronos parece ótimo, mas cada tarefa rodando em segundo plano consome sua largura de banda mental. Você ainda precisa lembrar que a tarefa está ativa, revisar o resultado quando ela conclui e reinserir esse contexto no seu fluxo de trabalho principal. Quando são tarefas não correlatas, a penalidade é ainda maior por exigir uma troca de contexto completa. É irônico como nós, humanos, sofremos de problemas de contexto de forma muito semelhante às IAs.

Mas, como diz o [teorema fundamental da engenharia de software](https://en.wikipedia.org/wiki/Fundamental_theorem_of_software_engineering):

> Todos os problemas da ciência da computação podem ser resolvidos com mais um nível de indireção... exceto o problema de termos camadas de indireção demais.

Este ano estamos testemunhando a ascensão dos "Gerenciadores de Agentes": agentes responsáveis por coordenar outros agentes. Embora esse conceito tenha surgido na codificação (como no [Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}) e no [`scion`](https://github.com/googlecloudplatform/scion)), suas implicações são muito mais amplas.

Isso nos coloca diante de um problema ainda maior: se já é difícil [revisar o trabalho de um único agente]({{< ref "/posts/20260303-code-reviews-in-2026" >}}), ou de alguns poucos, como poderemos revisar o trabalho de frotas inteiras de agentes? Não existe resposta fácil para essa pergunta, mas, na minha visão, precisamos trilhar o caminho para construir confiança em sistemas multiagente. Ou, melhor ainda, como dizem na segurança: confie, mas verifique.

Da mesma forma como ampliamos a confiança em sistemas de agente único aplicando técnicas de prompting, criando hooks, sandboxes e ferramentas determinísticas, precisaremos encontrar formas de adicionar barreiras de qualidade (*quality gates*) aos gerenciadores de agentes. Avaliação de agentes (*evals*), auditabilidade e práticas de engenharia mais maduras serão fundamentais para essa mudança.

Ainda assim, tenho certeza de que chegaremos lá. Anos de práticas consolidadas de engenharia são o que nos dá a tranquilidade de não precisar inspecionar a saída do compilador para checar se o código assembly gerado está correto. Com agentes, o processo não será diferente.

## Uma prévia do futuro: kanban agêntico

O que acontece quando tiramos o código do agente de codificação e focamos no produto que estamos construindo? Fiz esse exercício mental e percebi que o resultado não seria muito diferente do que temos hoje em um quadro Kanban — mas, em vez de pessoas humanas puxando tickets, teríamos predominantemente agentes realizando as interações:

![Mockup de kanban agêntico](image.png "Mockup conceitual de um quadro kanban agêntico criado com Google Stitch")

Temos as colunas típicas para os diferentes estágios de desenvolvimento (backlog, to do, in progress, etc.), mas cada coluna conta com um conjunto de agentes que trabalharão de forma colaborativa para levar o ticket ao próximo estágio. Adicionar skills globais à coluna fornece um contexto importante para todos os agentes envolvidos — como, por exemplo, padrões de arquitetura e procedimentos de controle. Cada etapa pode ser auditada clicando no ticket e acompanhando o fluxo de conversa entre os agentes. Precisa direcionar os agentes? Adicione um comentário ao ticket. Quer incluir uma etapa de revisão humana? Adicione a si mesma como um dos "agentes".

Ao combinar a gestão visual do Ágil com o poder de execução dos Gerenciadores de Agentes, podemos superar o limite da carga cognitiva. Fechamos o ciclo entre os mundos Ágil e Agêntico. As cerimônias do passado evoluem para os dashboards do futuro, provando que tudo o que aprendemos naquelas reuniões intermináveis de planejamento de sprint era apenas preparação para o que vem a seguir.

O que você acha dessa abordagem? Adoraria ouvir suas reflexões nos comentários abaixo ou em qualquer uma das minhas redes sociais!