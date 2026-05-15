---
title: "Do Ágil ao Agêntico: Um Guia para o Desenvolvimento Corporativo Moderno"
date: 2026-05-15T12:00:00Z
categories: ["Workflow & Best Practices"]
tags: ["agile", "agentic-coding", "gemini-cli", "mcp", "architecture"]
summary: "As transformações Ágeis estão obsoletas? Saiba como as práticas tradicionais do Ágil se mapeiam para os fluxos de trabalho agênticos e como escalá-los por toda a corporação."
---

Se você passou algum tempo significativo na indústria de software na última década, provavelmente já vivenciou uma transformação Ágil. Você participou de intermináveis reuniões de planejamento de sprint, compartilhou suas atualizações em daily standups e talvez até tenha se questionado se tudo isso fazia algum sentido.

Essa frustração é comum e geralmente decorre de uma desconexão entre a gestão de produtos e a realidade da engenharia. Os líderes de negócios querem previsibilidade, mas o desenvolvimento de software é inerentemente imprevisível. Quando as empresas tentam forçar a previsibilidade por meio de métricas e dashboards rigorosos, o Ágil falha. A metodologia torna-se uma burocracia rígida. Os desenvolvedores ficam frustrados porque as cerimônias parecem um fardo inútil, e práticas como a escrita de histórias se transformam em tarefas tediosas.

Em implementações Ágeis maduras, por outro lado, a liderança reconhece a incerteza e capacita as equipes para fazerem as melhores escolhas para o negócio. Métricas e dashboards deixam de ser o objetivo principal; em vez disso, o foco muda para a entrega de valor comercial. Isso não significa deixar tudo nas mãos dos desenvolvedores, mas trabalhar em estreita colaboração com eles em equipes verdadeiramente multifuncionais.

Este preâmbulo é necessário porque, para ler este artigo corretamente, quero que você se reconecte com as raízes do manifesto Ágil. Tenha em mente os [valores fundamentais](https://agilemanifesto.org/):

> Indivíduos e interações mais que processos e ferramentas  
> Software em funcionamento mais que documentação abrangente  
> Colaboração com o cliente mais que negociação de contratos  
> Responder a mudanças mais que seguir um plano

Neste artigo, vamos explorar como podemos transitar essas práticas Ágeis fundamentais diretamente para a era moderna dos fluxos de trabalho agênticos.

## O fluxo de trabalho de desenvolvimento agêntico

Antes de falarmos sobre escala corporativa, precisamos estabelecer a base. Cobri esses tópicos em meus artigos anteriores, mas acho que vale a pena colocá-los aqui novamente por uma questão de completude:

1.  **Escrever histórias é usar prompts:** A parte mais difícil de comandar um agente é fornecer o contexto certo. Desenvolvedores frequentemente têm dificuldade em escrever boas histórias. No entanto, essa é a habilidade mais importante que você pode cultivar hoje. A habilidade necessária para escrever uma boa história é exatamente a mesma necessária para escrever um bom prompt. Trata-se de organizar as informações de uma maneira fácil de consumir, com justificativa de negócios clara e resultados esperados. Também conhecido em termos Ágeis como **Definition of Ready (DoR)**.
2.  **A priorização dita o fluxo de trabalho:** O refinamento tradicional do backlog se traduz diretamente na gestão de suas ferramentas de IA. **Ao priorizar** as tarefas com base nas dimensões de *Valor de Negócio* versus *Certeza Técnica*, você pode decidir no que trabalhar de forma síncrona em primeiro plano (trabalhando em par com ferramentas como a [Gemini CLI](https://geminicli.com/)) versus o que delegar para agentes assíncronos em segundo plano (como o [Jules](https://jules.google/)).
3.  **O ciclo de codificação agêntica:** Os agentes de codificação são muito poderosos, mas muitas vezes carecem de consistência. Eles são não-determinísticos por definição. Podemos mitigar esse problema com o uso de ferramentas determinísticas. Costumo descrever isso como "reduzir a agência do agente". Por exemplo, se você sabe que o seu processo de build é sempre build, test, lint e deploy, o que você não quer fazer é especificar isso num prompt. O agente inevitavelmente esquecerá um ou mais passos à medida que a sessão avança. O que você realmente deseja fazer é empacotar esse processo como uma ferramenta personalizada e fornecê-la ao agente em vez disso, removendo completamente sua opção de esquecer qualquer um desses passos.

Se você domina essas três práticas, você já é um desenvolvedor agêntico eficaz. Mas como tornamos toda uma organização de engenharia eficaz?

## Arquitetura conversacional e o compartilhamento do conhecimento institucional

Escrever software corporativo é difícil, mas lembrar o *porquê* de determinado código ter sido escrito pode ser igualmente desafiador. O conhecimento tribal decai rapidamente. Quando um engenheiro sênior vai embora, seu conhecimento institucional muitas vezes vai com ele. Uma solução tradicional para esse problema é manter documentação exaustiva em wikis internas, mas estas tendem a ser difíceis de manter, descobrir e impor.

Por muitos anos, um dos meus textos favoritos sobre o compartilhamento de conhecimento corporativo tem sido este post do blog de Martin Fowler: [Scaling Architecture Conversationally](https://martinfowler.com/articles/scaling-architecture-conversationally.html). Os autores argumentam que uma boa arquitetura se espalha através da conversa, não apenas de mandatos de cima para baixo. O post também explora como formalizar essas conversas em Registros de Decisões de Arquitetura (ADRs) para que não se percam no tempo.

Os ADRs vão além das wikis simples; eles fornecem um instantâneo histórico do momento em que uma decisão foi tomada. Eles capturam as condições específicas, premissas e restrições que a justificaram. Essa noção pode parecer simples à primeira vista, mas empodera equipes futuras a fazerem alterações quando necessário. Porque têm um registro de *porquê* a escolha original foi feita, eles têm as ferramentas para avaliar se a decisão ainda se mantém e podem substituí-la com confiança (emitindo um novo ADR) quando essas restrições evoluírem.

Com o passar dos anos, cada vez mais acredito que a parte mais importante deste trabalho é gerenciar a incerteza. Os ADRs são uma das ferramentas que nos permitem ser honestos sobre o que sabemos e o que não sabemos. Quanto mais cedo percebermos que não há problema em não saber tudo, melhor. Essa é a essência de ser Ágil. Precisamos saber apenas o suficiente para progredir, acumular aprendizados para reduzir a incerteza e iterar. O software é um organismo vivo: ele nunca está pronto.

Embora os ADRs ofereçam múltiplas vantagens sobre um wiki não estruturado, eles ainda compartilham uma grande falha: a dependência dos humanos para estarem cientes e cumpri-los. Especialmente em grandes organizações, a comunicação torna-se o gargalo. Silos de informação são difundidos e muito esforço é gasto sincronizando diferentes partes do negócio.

### Distribuir conhecimento via agentes

Para escalar uma arquitetura hoje, devemos injetar esse conhecimento institucional diretamente nos agentes. Em vez de depender exclusivamente dos canais oficiais de transmissão pelos humanos, podemos usar a tecnologia para transmitir regulamentações, ADRs, procedimentos de controle e padrões corporativos diretamente aos nossos agentes. Se o seu contexto organizacional vive dentro da janela de contexto do agente, você garante que essas práticas sejam sempre aplicadas e atualizadas.

Do ponto de vista arquitetônico, um servidor MCP é o meio ideal para expor esse tipo de informação. Ele pode ser gerenciado centralmente e atualizado toda vez que um conselho de arquitetura, conselho de segurança ou outro comitê emite uma decisão. Prompts, ferramentas e skills são formas eficazes de mudar o comportamento do agente e podem ser consumidas tanto pelos agentes de codificação nas mãos dos engenheiros quanto por agentes automatizados em pipelines CI/CD.

É lamentável que os agent skills ainda não façam parte da especificação do MCP, mas há um grupo de trabalho dedicado a isso. Uma vez que pudermos usar servidores MCP para transmitir skills diretamente aos agentes, o desafio de mantê-las atualizadas será resolvido, reduzindo o atrito na transmissão de novos padrões aos desenvolvedores.

### A documentação do produto também é um produto consumível

Além das regras internas, esse exato mecanismo se aplica à documentação do produto. Tradicionalmente, se a Equipe A constrói uma API interna, ela publica um arquivo de especificação OpenAPI num portal de desenvolvedores e espera que a Equipe B leia o manual. Na era agêntica, a documentação estática cria atrito. Se o seu produto se destina a ser consumido por outras equipes, sua documentação deve ser consumível por suas ferramentas.

Quando a Equipe A envia seu serviço, eles também devem enviar um servidor MCP dedicado que exponha o schema da API, exemplos de integração e verificações de conformidade como ferramentas. Quando um desenvolvedor na Equipe B precisar se integrar com o serviço, ele simplesmente conecta seu agente de codificação ao servidor MCP da Equipe A. O agente pode consultar a estrutura da API, ler as regras de integração e escrever o código do cliente automaticamente. Passamos dos humanos lendo manuais para agentes lendo APIs, garantindo que a intenção arquitetônica e os padrões de integração sejam perfeitamente preservados em toda a empresa.

## Automatizando as cerimônias: Agentes não-codificadores

Embora gastemos muito tempo falando sobre agentes de codificação, há muitas oportunidades de otimização usando agentes não-codificadores para reduzir a sobrecarga de gerenciamento que muitas vezes assola a maioria das implementações Ágeis.

Desde tarefas simples, como fazer anotações em reuniões e resumi-las, até a repriorização do backlog, refinamento de histórias e criação de spikes, podemos usar agentes não-codificadores para recuperar grande parte do esforço gasto em administração e voltar o foco para a engenharia.

Aqui estão algumas maneiras pelas quais esses agentes focados em processos podem elevar uma equipe:

*   **Refinamento de Histórias e Quebra de Tarefas:** Se um Product Owner escreve um rascunho de um épico, um agente pode revisá-lo para identificar edge cases que faltam, premissas técnicas implícitas e caminhos de erro não tratados. Dê a ele acesso a skills específicos e ele será automaticamente compatível com os padrões organizacionais. Os pontos incertos podem ser automaticamente transformados em tickets de spike para exploração adicional.
*   **Auditando as Definition of Ready e Done:** Em muitos ambientes Ágeis, o DoR e o DoD são tratados como meros checklists numa wiki que são frequentemente esquecidos. Podemos tornar a conformidade proativa integrando agentes nos nossos quadros Kanban existentes (como o Jira ou GitHub Projects). Quando um ticket é movido para "Ready for Dev", um agente em segundo plano pode escaneá-lo para garantir que todo o contexto necessário, como schemas da API e mockups da interface, estejam realmente anexados. Se não estiverem, ele sinaliza a transição. De forma semelhante, antes que um ticket seja fechado, um agente pode verificar se testes foram adicionados e se a documentação foi atualizada.
*   **Retrospectivas Baseadas em Dados:** As retrospectivas costumam sofrer de viés de recência. Agentes não-codificadores podem atuar como analistas de dados objetivos, revisando as transições de tickets, comentários de pull requests e threads de chat do sprint. Por exemplo, ele pode apontar que tickets que tocavam um microsserviço específico passaram em média quatro dias em revisão, levando a equipe a questionar se há um silo de conhecimento de domínio que precisa ser abordado.

## Escalando o fluxo de trabalho com gerenciadores de agentes

Ao longo do último ano a indústria esteve praticamente focada no refinamento da experiência com o agente único, especialmente os agentes de codificação, e vimos o surgimento e a consolidação de novos padrões (MCP, skills, hooks, etc.).

Como uma consequência, a principal habilidade da carreira de engenharia de software passou da escrita de código para a orquestração de agentes. No entanto, há um gargalo oculto aqui: a carga cognitiva humana. Já existem relatos de um [novo tipo de burnout causado pelo uso da IA](https://techcrunch.com/2026/02/09/the-first-signs-of-burnout-are-coming-from-the-people-who-embrace-ai-the-most/).

Delegar tarefas para agentes assíncronos soa muito bem, mas toda tarefa em execução em segundo plano consome a sua largura de banda mental. Você ainda precisa lembrar que a tarefa está ativa, revisar seu output quando termina e mesclar o contexto de volta para o seu fluxo de trabalho principal. Quando essas são tarefas não relacionadas, a penalidade é ainda maior, pois exige uma mudança completa de contexto. É irônico como nós humanos, assim como a IA, também sofremos de problemas de contexto.

Mas como diz o [teorema fundamental da engenharia de software](https://en.wikipedia.org/wiki/Fundamental_theorem_of_software_engineering):

> Todos os problemas em ciência da computação podem ser resolvidos com outro nível de indireção... Exceto o problema de ter muitas camadas de indireção.

Este ano, estamos vendo o surgimento dos "Gerenciadores de Agentes": agentes responsáveis por gerenciar outros agentes. Embora esse conceito tenha sido visto pela primeira vez em código (por exemplo, [`Antigravity`](https://antigravity.google/) e [`scion`](https://github.com/googlecloudplatform/scion)), ele possui implicações muito mais amplas.

No entanto, isso cria um problema ainda maior: se já é difícil revisar o trabalho de um agente, ou de alguns poucos, como é possível revisar o trabalho de frotas de agentes? Não há uma resposta fácil para essa pergunta, mas na minha opinião, precisamos trabalhar nossa maneira de construir confiança nos sistemas multi-agentes. Ou melhor ainda, como se diz na área de segurança: confie, mas verifique.

Da mesma forma que podemos aumentar a confiança nos sistemas de agente único aplicando técnicas de prompts, criando hooks, sandboxes e ferramentas determinísticas, precisaremos encontrar a nossa maneira de adicionar portões de qualidade (quality gates) aos gerenciadores de agentes. A avaliação de agentes, a auditabilidade e práticas de engenharia mais maduras serão fundamentais para essa mudança.

Apesar disso, acredito que chegaremos lá. Anos de práticas de engenharia é o que nos dá confiança para não inspecionar o código de saída de nossos compiladores para ver se o código assembly correto foi gerado. Isso não será diferente.

## Um olhar para o futuro: Kanban agêntico

O que acontece quando tiramos o código do agente de codificação e focamos no produto que estamos construindo? Fiz isso como um experimento mental e percebi que não seria muito diferente do que temos hoje num quadro kanban, mas em vez de humanos pegando os tickets, teríamos predominantemente agentes realizando as interações:

![O Kanban Agêntico Unificado](kanban.png "Um mockup conceitual de um quadro de desenvolvimento agêntico")

Nós teríamos as colunas típicas para as diferentes fases de desenvolvimento (backlog, to do, in progress, etc.), mas cada coluna tem um conjunto de agentes que trabalhará de forma colaborativa para trazer o ticket para o próximo estágio. A adição de skills globais à coluna pode fornecer um contexto importante para todos os agentes envolvidos, como, por exemplo, padrões de arquitetura e procedimentos de controle. Toda etapa pode ser auditada clicando no ticket e acompanhando o fluxo da conversa entre os agentes. Precisa direcionar os agentes? Adicione um comentário ao ticket. Deseja ter uma etapa de revisão humana? Adicione você mesmo como um dos "agentes".

Combinando a gestão visual do Ágil com o poder de execução dos Gerenciadores de Agentes, podemos solucionar o limite de carga cognitiva. Encerramos o ciclo reunindo o mundo Ágil e o Agêntico. As cerimônias do passado evoluem para os dashboards do futuro, provando que tudo o que aprendemos durante aquelas infinitas reuniões de sprint foi apenas preparação para o que vem por aí.

O que você acha dessa abordagem? Eu adoraria ouvir a sua opinião nos comentários abaixo ou em qualquer uma das minhas redes sociais.