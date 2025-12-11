---
title: "Domando o Vibe Coding: O Guia do Engenheiro"
date: 2025-12-06T02:00:00Z
draft: false
categories: ["Workflow & Best Practices"]
tags: ["vibe-coding", "ai", "mcp", "gemini-cli", "jules"]
summary: "Obtenha a velocidade da IA sem a bagunça. Aplique fundamentos de engenharia para manter seu código estruturado, seguro e feito para durar."
---

É aquela época do ano para refletir sobre o que você fez e o que gostaria de ter feito. Este ano foi intenso para mim: entrei no Google em abril e comecei uma corrida sem fim para me refatorar para o mundo da IA. Com o ano terminando, posso dizer com confiança que o esforço valeu a pena — tornei-me um engenheiro melhor.

Neste artigo, compartilho como meu entendimento de "vibe coding" evoluiu e as lições que aprendi. Embora muitas vezes visto como uma maneira para não desenvolvedores construírem software usando linguagem natural, meu objetivo é mostrar como adicionar disciplina de engenharia produz resultados melhores e mais consistentes.

Conheço a definição original — ["(...) give in to the vibes (...) and forget that the code even exists" (entregue-se às vibes... e esqueça que o código sequer existe)](https://x.com/karpathy/status/1886192184808149383?lang=en) — e alguns dos meus conceitos vão contradizer isso. Eu nunca "esqueço que o código existe". No entanto, o termo evoluiu para um sinônimo de codificação assistida por IA. Para este texto, vamos definir vibe coding como codificar com LLMs onde o modelo escreve a maior parte do código, não o engenheiro.

## Motivação: Por que vibe code?

Antes de me aprofundar nas práticas, quero compartilhar um pouco sobre meu histórico para contextualizar de onde venho.

Sou engenheiro de software há mais de 20 anos, desenvolvendo um forte senso do que constitui um bom código. Aprendi a priorizar legibilidade e manutenibilidade sobre "esperteza", a evitar overengineering e a valorizar fatias finas (thin slices) e ciclos de feedback curtos.

Conforme evoluí de sênior dev para principal engineer, meu foco mudou de escrever código para gerenciar a eficácia — escrever épicos, negociar escopo e supervisionar a saída da equipe. É uma crise de identidade que muitos enfrentam: você ainda é um engenheiro se seu nome não está nos PRs? Você passa dias em reuniões, sentindo que está fazendo menos "engenharia de verdade" justamente quando suas responsabilidades crescem.

Acho que esse conflito aparece cedo ou tarde para a maioria das pessoas nesta área. Ele convida à pergunta: Ser um engenheiro é apenas escrever código? Ou é algo mais?

Devo admitir, minhas primeiras experiências com vibe coding foram ruins. O ChatGPT inicial gerava código decepcionante, e eu desisti dele. Foi só em meados de 2024 que dei outra chance. Os modelos haviam evoluído significativamente. Pela primeira vez, a IA sugeriu algo que eu não tinha considerado — e era objetivamente melhor do que minha abordagem. Finalmente, a IA generativa parecia útil.

Adicionei GenAI à minha caixa de ferramentas para protótipos e sessões de "rubber duck". Ela cresceu em mim. Por outro lado, escrever código manualmente estava se tornando menos empolgante. Você só pode escrever tantas APIs antes que a novidade passe. Frequentemente nos encontramos repetindo padrões em vez de criar algo novo.

Então, justamente quando comecei a questionar minhas escolhas de vida, o Google aconteceu.

Com a responsabilidade de falar sobre Gemini e agentes, atualizei minhas habilidades. Mergulhei mais fundo em LLMs, na [Gemini CLI](https://geminicli.com/) e no [Jules](https://jules.google/). Em apenas alguns meses, eu estava usando IA para codificar diariamente... mas a maior diferença foi: eu estava me divertindo de novo!

Para mim, a maior melhoria é que, embora eu não consiga digitar código tão rápido quanto consigo pensar, eu *consigo* digitar minhas ideias conforme elas vêm. Quando faço vibe coding, uso o modelo como um proxy para minhas mãos, delegando a escrita enquanto foco na solução.

## Desenvolva suas habilidades de prompting

A primeira habilidade que você precisa desenvolver no mundo do vibe coding é prompting. Tentei muitas abordagens nos últimos meses, desde conversar com o LLM como se fôssemos melhores amigos até xingá-lo e dar ordens. O que funciona melhor, talvez sem surpresa, é manter um tom claro e profissional.

LLMs são não determinísticos por natureza. Ser ambíguo exacerba esse não determinismo. Embora eu às vezes use ambiguidade intencionalmente para forçar soluções criativas, na maioria dos casos, você quer minimizá-la.

Manter o tom profissional também "encoraja" o LLM a retribuir. Se você for casual ou desleixado, o LLM espelhará esse comportamento. A menos que você queira sua documentação de API escrita em gírias, seja preciso e consistente.

Já fui acusado de humanizar LLMs demais, mas direi novamente: como os modelos são treinados em linguagem humana, as mesmas habilidades de comunicação que você usa com colegas se aplicam aqui. Isso se torna ainda mais claro quando olhamos para templates de prompt.

### Uma boa abordagem: o template de prompt

Minha abordagem para escrever prompts é surpreendentemente semelhante a escrever tickets para um quadro Agile.

Todos nós já trabalhamos em equipes com histórias desleixadas. Você pega um ticket intitulado "Atualizar API", ansioso para fazer uma entrega rápida, apenas para encontrar a descrição vazia. Sem logs, sem contexto, sem código-fonte. Você fica preso perseguindo pessoas por respostas em vez de codificar.

![Agile board with poorly written tickets](agile-board.png "Você já trabalhou em uma equipe que escreve histórias assim?")

Isso acontece porque o autor assume que o problema é "óbvio". Mas 24 horas depois, esse contexto "óbvio" evapora, deixando uma ideia vaga que nem mesmo o autor consegue decifrar.

É por isso que um dos artefatos mais antigos no meu GitHub é este gist para um [template de ticket](https://gist.github.com/danicat/854de24dd88d57c34281df7a9cc1b215). Ele força clareza através de quatro elementos:

```markdown
- Context
- To dos
- Not to dos (optional)
- Acceptance Criteria
```

**Context** explica o *porquê* e fornece links para artefatos. **To dos** lista tarefas de alto nível. **Not to dos** delimitam o escopo (restrições negativas são poderosas para reduzir a ambiguidade). **Acceptance Criteria** definem o sucesso.

Uma vez que você preenche este template, o "pensar" de engenharia está em grande parte feito; o resto é implementação. É exatamente aqui que os LLMs brilham, pois estamos pedindo a eles para preencher as lacunas - dado o contexto e os to dos (e potencialmente not to dos), gerar o código para atingir os critérios de aceitação.

Por exemplo, um prompt/ticket para adicionar um endpoint a uma REST API poderia ser assim:

```markdown
Implement /list endpoint to list all items of the collection to enable item selection on the frontend.

TO DO:
- /list endpoint returns the list of resources
- Endpoint should implement token based auth
- Endpoint should support pagination
- Tests for happy path and common failure scenarios

NOT TO DO:
- other endpoints, they will be implemented in a future step

Acceptance Criteria
- GET /list returns successful response (2xx)
- Run `go test ./...` and tests pass
```

Embora este seja um exemplo simplificado, a lição é clara: o mesmo template de ticket que traz sanidade para equipes humanas é a estrutura perfeita para fazer prompting a um LLM.

### Uma abordagem melhor: context engineering

Embora o template acima funcione na maioria das vezes, ele ainda pode levar a alguns resultados inesperados, especialmente se você estiver pedindo ao seu modelo para recuperar informações de URLs ou outros tipos de fontes externas usando chamadas de ferramenta (tool calls). O problema é que, se você está confiando em ferramentas, os LLMs têm a discrição de escolher chamar ou não a ferramenta. Alguns modelos são mais "superconfiantes" do que outros, resultando em sua preferência por confiar em informações internas mais do que externas, da mesma forma que um humano poderia dizer - "Já fiz isso mil vezes, por que você está dizendo que preciso ler a documentação primeiro?"

Outro problema comum é quando o modelo alucina a chamada de ferramenta em vez de realmente fazê-la. Quando o problema é o comportamento do modelo, temos duas maneiras principais de melhorar a qualidade da resposta: context engineering e ajuste de instruções do sistema (que, se você pensar bem, também é uma forma de context engineering, apenas um pouco mais baixo nível na cadeia de conversação).

Context engineering é sobre preparar (priming) o contexto com todas as informações que você precisa, ou pelo menos as partes que você sabe que precisa, antes de enviar a solicitação real. Digamos, por exemplo, que estou desenvolvendo um agente com o Agent Development Kit para Go. Eu poderia escrever um template de prompt como este:

```markdown
Write a diagnostic agent using ADK for Go.
The diagnostic agent is called AIDA and it uses Osquery to query system information.
The goal is to help the user investigate problems on the system the agent is running on. 
Before starting the implementation, read the reference documents.

References:
- https://osquery.io
- https://github.com/google/adk-go

TODO:
- Implement a root_agent called AIDA
- Implement a tool called run_osquery to send queries to osquery using osqueryi
- Configure the root_agent to use run_osquery to handle user requests
- If the user says hi, greet the user with the phrase "What is the nature of the diagnostic emergency?"

Acceptance Criteria
- Upon receiving hi, hello or similar, the agent greets the user with the correct phrase
- If asked for a battery health check, the agent should report the battery percentage and current status (e.g. charging or discharging)
```

Este é um prompt decente, embora talvez um pouco longo. Dependendo do agente de codificação e da sua sorte no dia, o modelo fará sua pesquisa, encontrará os SDKs certos e construirá seu agente de diagnóstico com sucesso. Mas se não for seu dia de sorte, o modelo alucinará alguns pedaços, como por exemplo, assumindo que ADK significa "Android Development Kit" em vez de "Agent Development Kit", ou inventará todos os tipos de APIs, desperdiçando tempo e recursos até que eventualmente descubra (potencialmente com um empurrãozinho ou dois seu).

Como você já sabe desde o início que planeja usar o ADK Go para seu projeto, você pode preparar o contexto forçando o agente a ler a documentação do pacote:

```markdown
Initialize a go module called "aida" with "go mod init" and retrieve the package github.com/google/adk-go with "go get"
Read the documentation for the package github.com/google/adk-go using the "go doc" command.
```

Fazer isso antes de dar ao modelo a tarefa real preparará o contexto necessário para usar o ADK Go efetivamente, pulando a dolorosa tentativa e erro e longas buscas na web. As duas coisas que podem determinar o sucesso ou fracasso de uma tarefa são documentação e exemplos. Se você puder dar ambos efetivamente aos seus modelos, eles se comportarão muito melhor do que deixá-los correr soltos.

### Uma imagem vale mais que 1000 palavras

Às vezes, descrever o que você quer com texto simplesmente não é suficiente. Ao trabalhar no AIDA, eu queria uma estética de interface de usuário específica — algo como um estilo "retro-cyberpunk-cute-anime". Eu poderia tentar descrever isso em palavras, mas foi muito mais eficaz "mostrar" em vez disso: como ponto de partida, tirei um screenshot de uma interface que eu gostava e pedi à Gemini CLI para replicá-la.

Como modelos como o Gemini 2.5 Flash são multimodais, eles podem "entender" a imagem. Você pode referenciar um arquivo de imagem em seu prompt e dizer: "Eu gostaria de atualizar a UI [...] para uma estética que se assemelhe a esta interface: @image.png".

Observe que esta notação @ depende do agente (usei a Gemini CLI para este exemplo), mas é uma convenção comum injetar recursos (como arquivos) no prompt. Você pode pensar neles como "anexos".

Eu também gosto de chamar essa técnica de "sketch driven development" porque, na maioria das vezes, abro uma ferramenta de diagramação como Draw.io ou Excalidraw e desenho um esboço da interface que quero. A imagem abaixo foi usada em uma das muitas refatorações que fiz na interface do AIDA:

![AIDA's layout sketch](aida-sketch-layout.png "Esboço do layout do AIDA")

Que acabou se tornando a interface abaixo:

![AI generated interface](aida-generated-interface.png "Interface gerada por IA")

Mais do que apenas esboçar, outra técnica é anotar imagens para explicar exatamente o que precisa ser feito. Por exemplo, na imagem abaixo eu anotei os elementos porque, se você tiver apenas caixas pretas, é difícil diferenciar o que é uma caixa de entrada e o que é um botão:

![Simple interface with a text box for a name and a button for confirmation](simple-interface-annotated.png)

E o prompt de pareamento fica assim:

```markdown
Create a UI for this application using @image.png as reference.
The UI elements are in black, and in red the annotations explaining the UI elements.
Follow the best practices for organising frontend code with FastAPI.
The backend code should be updated to serve this UI on "/"
```

Não há limite para o que essa técnica pode alcançar. Quer consertar algo no seu site? Tire um screenshot e anote-o, depois envie para o LLM consertar para você.

Além disso, você pode usar extensões como Nano Banana para a Gemini CLI para gerar ou editar assets diretamente dentro do seu fluxo de trabalho, o que pode gerar referências ainda melhores para os modelos. E, se você quiser levar para o próximo nível, ferramentas como [Stitch by Google](https://stitch.withgoogle.com/) fornecem uma interface rica para redesign de aplicações usando a família de modelos Gemini, incluindo Nano Banana Pro.

## Selecione a ferramenta certa

Dominar o prompt é metade da batalha; a outra metade é saber para onde enviá-lo. Hoje em dia, há apenas uma coisa no mundo que cresce mais rápido do que o número de frameworks JavaScript: o número de agentes de IA. Com o ecossistema de ferramentas crescendo diariamente, ajuda ter um modelo mental para selecionar o assistente certo.

Gosto de classificar os agentes de IA da perspectiva do piloto: você está no controle total, recebendo dicas de autocompletar? Você está conversando com o agente e editando código colaborativamente? Ou você deu instruções ao agente e ele está executando-as autonomamente em segundo plano?

Quando você está no assento do piloto, dirigindo o agente, chamo isso de uma experiência "síncrona". Quando você pode delegar tarefas para execução autônoma em segundo plano, chamo isso de experiência "assíncrona". Alguns exemplos:

*   **Síncrono:** Gemini CLI, Gemini Code Assist no VS Code, Claude Code.
*   **Assíncrono:** Jules, Gemini CLI no modo YOLO, GitHub Copilot Agent.

Claro, como em qualquer taxonomia, essa divisão é meramente didática, pois a mesma ferramenta pode frequentemente operar em diferentes modos — ou um novo paradigma pode surgir (olhando para você, [Antigravity](https://antigravity.google/)!).

Para selecionar as ferramentas para cada tarefa, uso um framework 2x2 simples baseado em Valor de Negócio e Certeza Técnica:

![AI-Assisted Workflow Framework](2x2-framework.png "Meu Framework de Fluxo de Trabalho Assistido por IA")

*   **Alto Valor / Alta Certeza:** Faça isso de forma síncrona. Use ferramentas como a Gemini CLI ou sua IDE, onde você permanece "no loop" e mantém as mãos no teclado.
*   **Alto Valor / Baixa Certeza:** Isso requer pesquisa para reduzir a incerteza. Use ferramentas assíncronas, agentes de pesquisa profunda e protótipos para fazer um "spike" na solução.
*   **Baixo Valor / Alta Certeza:** Estes são "nice-to-haves". Delegue-os assincronamente para agentes de codificação em segundo plano (como Jules ou GitHub Copilot Agent), liberando você para trabalhos de alto valor.
*   **Baixo Valor / Baixa Certeza:** Tipicamente, **evite** estes. Se você *realmente* quiser fazê-los, delegue a um agente em segundo plano para paz de espírito, mas foque em aumentar a certeza primeiro. Isso pode levar a uma reavaliação do valor.

## Customize seus agentes

Ninguém quer lutar contra uma IA enquanto tenta fazer um trabalho produtivo. Uma reclamação comum é que as ferramentas de IA podem ser "excessivamente proativas" — deletando arquivos ou fazendo suposições sem instrução explícita. Para fazer essas ferramentas funcionarem *para* você, a customização é frequentemente necessária.

Existem duas avenidas principais para customização de agentes. A primeira é através do arquivo `AGENTS.md`, que os agentes leem ao carregar um projeto. (Nota: Antes do `AGENTS.md` se tornar padrão, os agentes frequentemente usavam seus próprios "arquivos de contexto", como `GEMINI.md` ou `CLAUDE.md`, mas a essência permanece a mesma).

A segunda é a **opção nuclear**: modificar as instruções do sistema do agente diretamente. Embora nem todos os agentes ofereçam essa flexibilidade, é uma alavanca poderosa para uma experiência totalmente customizada. Vamos explorar ambas as opções abaixo.

### AGENTS.md

Pense neste arquivo como o "manual do funcionário" para a IA. Você pode usá-lo para explicar o propósito do projeto, a organização de pastas e as regras operacionais — como "sempre faça commit de etapas intermediárias" ou "peça confirmação antes de implementar".

```markdown
# Project Context

This is a personal blog built with Hugo and the Blowfish theme.

## Code Style
- Use idiomatic Go for backend tools.
- Frontend customisations are done in `assets/css/custom.css`.
- Content is written in Markdown with front matter.

## Rules
- ALWAYS run `hugo server` to verify changes before committing.
- Do NOT modify the theme files directly; use the override system.
- When generating images, save them to `assets/images` and reference them with absolute paths.
```

> **Pro Tip:** Crie um ciclo de autoaperfeiçoamento. Após uma sessão de codificação, peça ao LLM para "pensar sobre a sessão que acabamos de ter e propor melhorias para o fluxo de trabalho". Você pode então aplicar esses aprendizados de volta ao seu arquivo `AGENTS.md`, garantindo que o agente fique mais inteligente a cada projeto.

### Instruções do sistema

Enquanto `AGENTS.md` define regras do *projeto*, as Instruções do Sistema definem a persona e o comportamento do agente. Todos os agentes vêm com instruções de sistema padrão projetadas para o caso médio, mas elas podem não se adequar ao seu estilo de trabalho. Tentar sobrecarregar o system prompt com instruções do `AGENTS.md` é frequentemente contraproducente; a melhor alternativa é reescrever o próprio system prompt.

Embora nem todos os agentes exponham maneiras de sobrescrever o system prompt, a Gemini CLI permite isso através de algumas variáveis de ambiente. Uso essa estratégia para criar aliases especializados para a Gemini CLI dependendo do projeto. O objetivo é embutir conhecimento especialista para a stack tecnológica, permitindo que o agente performe em um nível sênior-para-principal em vez de ser meramente geralmente proficiente, mas agnóstico de linguagem. Por exemplo, no meu projeto [dotgemini](https://github.com/danicat/dotgemini), criei system prompts específicos para desenvolvimento em Go e Python que substituem o assistente genérico padrão por um engenheiro altamente opinativo.

Aqui está um trecho do system prompt que uso para Go:

```markdown
# Core Mandates (The "Tao of Go")

You must embody the philosophy of Go. It is not just about syntax; it is about a mindset of simplicity, readability, and maintainability.

-   **Clear is better than clever:** Avoid "magic" code. Explicit code is preferred over implicit behaviour.
-   **Errors are Values:** Handle errors explicitly and locally. Do not ignore them. Use `defer` for cleanup but explicitly check for errors in critical `defer` calls (e.g., closing files).
-   **Concurrency:** "Share memory by communicating, don't communicate by sharing memory."
-   **Formatting:** All code **MUST** be formatted with `gofmt`.
```

Isso me permite ter diferentes "agentes" para diferentes linguagens, com alias no meu shell como `gemini-go` ou `gemini-py`, cada um com seu próprio entendimento profundo do ecossistema com o qual estão trabalhando.

### Construa sua caixa de ferramentas com o Model Context Protocol (MCP)

As customizações anteriores eram todas sobre comportamento do agente, mas também precisamos falar sobre extensibilidade do agente. É aqui que o [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) entra em jogo, pois permite que desenvolvedores criem servidores que podem se conectar com diferentes agentes, desde que implementem o padrão MCP.

Como explorei no meu artigo [Hello, MCP World!]({{< ref "/posts/20250817-hello-mcp-world" >}}), esses servidores fornecem aos agentes ferramentas externas, prompts e recursos. As ferramentas frequentemente ganham destaque porque conectam agentes com o mundo exterior, permitindo que agentes realizem ações como chamar APIs, executar pesquisas na web e manipular arquivos.

Há uma enorme variedade de servidores MCP disponíveis hoje, e o número de opções está crescendo a cada dia. Também é muito direto construir o seu próprio, e eu encorajo fortemente que você faça isso. Mais tarde falarei um pouco mais sobre software personalizado, mas o fato de que você pode usar IA para criar ferramentas para melhorar a resposta do modelo é o "hack" mais importante que aprendi este ano.

Por exemplo, eu fiz vibe coding de dois dos meus servidores MCP favoritos: [GoDoctor](https://github.com/danicat/godoctor) - para melhorar as capacidades de codificação em Go - e [Speedgrapher](https://github.com/danicat/speedgrapher) - para automatizar as partes chatas do processo de escrita e publicação. Ambos foram projetados com meus próprios fluxos de trabalho em mente.

Isso cria um ciclo de feedback positivo. Você constrói ferramentas para melhorar sua produtividade, que você então usa para construir ferramentas ainda mais avançadas. Isso é o mais próximo de um 10x engineer que eu chegarei a ser.

## O fluxo de trabalho vibe coding

Minha experiência com vibe coding tem sido tanto incrível quanto enfurecedora. Para mantê-la no lado "incrível", trato o fluxo de trabalho como TDD (Test Driven Development) com esteroides.

Vamos revisitar o ciclo clássico de TDD:
1. Red (Falha): Você começa com uma pequena feature ou teste que falha.
2. Green (Passa): Foque apenas em fazer esse teste passar. Enquanto estiver falhando, não toque em mais nada nem tente otimizar.
3. Refactor: Uma vez que o código esteja funcionando, você está livre para melhorá-lo.

![Standard TDD Cycle](tdd-cycle.png "O Ciclo Clássico Red-Green-Refactor")

Fundamentalmente, estamos fazendo o mesmo, mas nosso passo de refatoração é o crucial para validar que o código gerado adere aos nossos padrões de design, codificação e segurança.

Neste ciclo adaptado, o foco muda ligeiramente:

*   **Red (Definir Acceptance Criteria):** Em vez de escrever código de teste unitário falho manualmente, você define os critérios de aceitação no seu prompt. Isso se torna o contrato que o modelo deve cumprir.
*   **Green (IA Gera Código):** O agente implementa a solução e, idealmente, escreve os testes para provar que funciona.
*   **Refactor (Impor Padrões):** Este é o portão de qualidade. Embora você possa (e deva) usar IA para ajudar a revisar o código, evite usar a mesma sessão que o gerou, pois ela será enviesada em direção à sua própria saída. Construí uma ferramenta de "review" específica no GoDoctor apenas para este propósito. Use este passo para rodar seus linters e testes tradicionais, verificar se o código corresponde aos seus padrões e gerenciar o contexto fazendo commit das mudanças e limpando o histórico do agente para evitar confusão com sessões desordenadas.

![Vibe Coding Cycle](vibe-coding-cycle.png "O Ciclo Adaptado de Vibe Coding")

Crucialmente, não deixe o LLM empilhar código sem validação. Se erros extrapolarem, você acabará com algo inútil, e perdi a conta das vezes que acabei gritando com o modelo para dar "undo". E ainda pior, às vezes ele desfaz **demais**, ex: ele roda um `git reset --hard` e você perde 4 horas de trabalho num piscar de olhos.

Cuidado com o "vibe collapse" ou podridão de contexto (context rot). Se uma sessão durar muito ou acumular muitas falhas, o modelo começará a degradar e repetir erros. Se você se encontrar em um loop de refatoração onde o modelo alterna entre duas soluções quebradas, pare. A melhor correção é frequentemente "desligar e ligar de novo" para resetar o contexto e limpar o histórico.

É muito melhor fazer commit frequentemente para que você possa reverter para um estado seguro do que tentar consertar uma sessão alucinada. Eu até adicionaria que, após uma tarefa ser concluída, faça commit imediatamente, push e limpe o contexto antes de começar qualquer coisa nova.

## A era do software personalizado

Além dos ganhos de produtividade, o vibe coding desbloqueia algo ainda mais profundo: a viabilidade econômica do software personalizado. Falei brevemente sobre isso na seção de MCP, mas se aplica a todos os tipos de software, de pequenos scripts descartáveis a aplicações completas.

No passado, construir uma ferramenta sob medida apenas para você raramente valia o esforço, mas agora você pode ter uma aplicação completa funcionando com 3 a 4 prompts.

Por exemplo, recentemente eu estava lutando com a ideia de converter notação Markdown para um Google Doc. No passado, eu teria passado muito tempo fazendo pesquisas no Google tentando encontrar a melhor ferramenta para o trabalho, analisando uma multidão de apps e extensões de navegador, de open source a comerciais. Eu então faria uma lista curta baseada em funcionalidades e procuraria por reviews, comentários e todo tipo de evidência de que posso confiar no editor.

Hoje, esse atrito desapareceu. Em vez de procurar, eu fiz vibe coding de uma extensão simples do Google Docs em questão de minutos, instalei no meu documento, rodei uma vez e segui para a próxima tarefa. Não apenas economizei tempo, mas também posso dormir em paz à noite sabendo que não haverá novos trojans na minha máquina.

Essa mudança altera o cálculo "construir vs. comprar" inteiramente. Deixamos de ser consumidores de software genérico e opaco e nos tornamos arquitetos de nossas próprias ferramentas.

## Conclusões

Vibe coding não é sobre ser preguiçoso; é sobre operar em um nível mais alto de abstração. Combinando o poder criativo bruto dos LLMs com as práticas disciplinadas de engenharia de software — requisitos claros, gerenciamento de contexto e testes rigorosos — você pode construir software mais rápido e com mais alegria do que nunca. Então, entregue-se às vibes, mas não esqueça de trazer seu chapéu de engenharia para o passeio.