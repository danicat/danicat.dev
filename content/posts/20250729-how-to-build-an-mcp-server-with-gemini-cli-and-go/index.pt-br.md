---
author: Daniela Petruzalek
categories:
- Agent Development
date: 2025-07-29
summary: Um guia passo a passo sobre como construí o GoDoctor, um assistente de desenvolvimento Go baseado em IA, usando o Model Context Protocol (MCP) e o Gemini CLI.
tags:
  - ai
  - codelab
  - gemini-cli
  - golang
  - mcp
  - tutorial
title: "Construindo o GoDoctor: Como Criar um Servidor MCP com Gemini CLI e Go"
slug: "how-to-build-an-mcp-server-with-gemini-cli-and-go"
aliases:
  - "/pt-br/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go/"
description: "Tutorial passo a passo construindo o GoDoctor, um servidor MCP em Go com o Gemini CLI. Aborda transportes stdio e HTTP, handshake e deploy no Cloud Run."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "Gemini CLI / Antigravity CLI"
  - "MCP Go SDK"
---

{{< alert "circle-info" >}}
**Nota:** Este artigo foi escrito para o Gemini CLI, que foi descontinuado e substituído pelo **Google Antigravity 2.0**. Para saber mais sobre a nova Antigravity CLI (`agy`), o SDK e o ecossistema Antigravity mais amplo, confira [O Guia do Mochileiro para o Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}).
{{< /alert >}}

## Introdução

Como muitos de vocês, tenho mergulhado fundo no desenvolvimento assistido por IA. A jornada é muitas vezes uma montanha-russa de momentos "uau" e obstáculos frustrantes. Esta é a história de uma dessas jornadas — uma história que começou comigo tentando construir uma coisa, me desviando completamente por causa de um problema frustrante e terminando com uma ferramenta que melhorou fundamentalmente meu fluxo de trabalho assistido por IA.

Meu objetivo original era construir um servidor de Model Context Protocol (MCP) para o [osquery](https://www.osquery.io/), uma ferramenta que permite consultar o estado de uma máquina usando SQL. Eu estava empolgada para usar o Gemini CLI para me ajudar a escrever o código em Go. No entanto, rapidamente bati de frente com uma parede. O código em Go que o agente produzia frequentemente não era idiomático. Ele cometia erros de iniciante, criava abstrações excessivas e com frequência "alucinava" APIs inteiras que simplesmente não existiam. Minha suspeita era de que o modelo subjacente não havia sido treinado no mais recente [Go SDK for MCP](https://github.com/modelcontextprotocol/go-sdk), então ele preferia inventar respostas a admitir que não sabia.

Essa experiência me levou a uma constatação fundamental: em vez de lutar contra a ferramenta, eu poderia *ensiná-la*. Decidi pausar meu projeto com o osquery e embarcar em uma "side quest": construir um servidor MCP dedicado cujo único propósito fosse ser um especialista em desenvolvimento Go. Esse projeto paralelo, que eventualmente chamei de [GoDoctor](https://github.com/danicat/godoctor), forneceria as ferramentas de que o Gemini CLI precisava para escrever um código Go muito melhor.

Neste post, vou guiar vocês pela história da construção do GoDoctor. Este texto é menos um tutorial tradicional e mais uma jornada "orientada por prompts". Vamos focar em como traduzir os requisitos do projeto em prompts eficazes e guiar a IA pelos detalhes de implementação, aprendendo com os erros inevitáveis ao longo do caminho.

## Preparando o Terreno: O `GEMINI.md`

Antes de escrever uma única linha de código, o primeiro passo foi estabelecer as regras básicas. Embora o `GEMINI.md` seja um arquivo específico do Gemini CLI, a prática de criar um arquivo de contexto é comum em muitos agentes de programação com IA (por exemplo, o Jules usa `AGENTS.md` e o Claude usa `CLAUDE.md`). Na verdade, existe um esforço emergente para padronizar isso com um arquivo chamado [`AGENT.md`](https://ampcode.com/AGENT.md). Esse arquivo é crucial porque fornece à IA uma compreensão fundamental dos padrões do seu projeto e das suas expectativas em relação ao comportamento dela.

Como este projeto era novinho em folha, eu ainda não tinha nenhum detalhe arquitetural específico para compartilhar. Por isso, comecei com um conjunto genérico de diretrizes focadas na criação de código Go idiomático e de alta qualidade. Conforme o projeto evolui, é comum adicionar instruções mais específicas sobre a estrutura do projeto, comandos de build ou bibliotecas essenciais. Para ver um exemplo de arquivo mais específico de projeto, você pode conferir o `GEMINI.md` que uso no meu [projeto `testquery`](https://github.com/danicat/testquery/blob/main/GEMINI.md).

Aqui está o `GEMINI.md` inicial que serviu como a constituição para a IA nesta jornada:

```markdown
# Go Development Guidelines
All code contributed to this project must adhere to the following principles.

## 1. Formatting
All Go code **must** be formatted with `gofmt` before being submitted.

## 2. Naming Conventions
- **Packages:** Use short, concise, all-lowercase names.
- **Variables, Functions, and Methods:** Use `camelCase` for unexported identifiers and `PascalCase` for exported identifiers.
- **Interfaces:** Name interfaces for what they do (e.g., `io.Reader`), not with a prefix like `I`.

## 3. Error Handling
- Errors are values. Do not discard them.
- Handle errors explicitly using the `if err != nil` pattern.
- Provide context to errors using `fmt.Errorf("context: %w", err)`.

## 4. Simplicity and Clarity
- "Clear is better than clever." Write code that is easy to understand.
- Avoid unnecessary complexity and abstractions.
- Prefer returning concrete types, not interfaces.

## 5. Documentation
- All exported identifiers (`PascalCase`) **must** have a doc comment.
- Comments should explain the *why*, not the *what*.

# Agent Guidelines
- **Reading URLs:** ALWAYS read URLs provided by the user. They are not optional.
```

Esse arquivo estabelece uma base de qualidade e estilo desde o início.

## Entendendo o Model Context Protocol (MCP)

No coração deste projeto está o Model Context Protocol (MCP). Algumas pessoas o descrevem como um "padrão USB para ferramentas de LLM", mas gosto de pensar de outra forma: **o que o HTTP e o REST fizeram pela padronização de APIs web, o MCP está fazendo pelas ferramentas de LLM.** Da mesma forma que o REST forneceu uma arquitetura previsível que desbloqueou um ecossistema massivo de serviços web, o MCP está trazendo uma linguagem comum fundamental para o universo dos agentes de IA. Trata-se de um protocolo baseado em JSON-RPC que estabelece um terreno comum, permitindo que qualquer agente que "fale" MCP descubra e use qualquer ferramenta compatível sem a necessidade de uma integração personalizada e pontual.

O protocolo define diferentes maneiras de comunicação entre o agente e o servidor de ferramentas, conhecidas como transportes (transports). Os dois mais comuns são:
*   **HTTP:** O modelo familiar de requisição/resposta (request/response), ideal para ferramentas implantadas como serviços remotos (por exemplo, no [Cloud Run](https://cloud.google.com/run?utm_campaign=CDR_0x72884f69_default_b421852297&utm_medium=external&utm_source=blog)).
*   **stdio:** Um transporte robusto que utiliza a entrada e saída padrão (standard input/output), perfeito para executar uma ferramenta como um processo local na sua máquina.

Com o transporte `stdio`, o agente e o servidor trocam uma série de mensagens JSON-RPC. O processo se inicia com um handshake de três etapas fundamental para estabelecer a conexão. Somente após a conclusão desse handshake o cliente pode começar a fazer chamadas de ferramentas.

A sequência funciona assim:

![Diagrama de sequência do stdio do MCP](images/mcp-stdio-sequence-diagram.png)
*<p align="center">Figura 1: Diagrama de sequência do transporte stdio da documentação oficial do MCP (2025-06-18).</p>*

Aqui está como são as mensagens JSON desse handshake inicial, com base na especificação oficial:

**1. Cliente → Servidor: Requisição `initialize`**
O cliente inicia a conversa enviando uma requisição `initialize`.
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-06-18",
    "clientInfo": {
      "name": "Gemini CLI",
      "version": "1.0.0"
    }
  }
}
```

**2. Servidor → Cliente: Resultado de `initialize`**
O servidor responde com o resultado, confirmando a versão do protocolo e anunciando suas capacidades e informações. Esta é a resposta real do binário do `godoctor`:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "capabilities": {
      "completions": {},
      "logging": {},
      "tools": {
        "listChanged": true
      }
    },
    "protocolVersion": "2025-06-18",
    "serverInfo": {
      "name": "godoctor",
      "version": "0.2.0"
    }
  }
}
```

**3. Cliente → Servidor: Notificação `initialized`**
Por fim, o cliente confirma que a configuração foi concluída enviando uma notificação `initialized`. Observe que isso é uma notificação, portanto não possui o campo `id`, e o método possui namespace.
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized",
  "params": {}
}
```

Uma vez concluída essa troca, a sessão está estabelecida e o cliente pode prosseguir com as chamadas de ferramentas. Por exemplo, para solicitar ao servidor a lista de ferramentas disponíveis, você pode enviar uma requisição `tools/list`. O segredo é que todas as três mensagens do handshake devem ser enviadas na ordem correta antes dessa requisição.

Você pode conferir a sequência completa em ação com este script shell:
```bash
#!/bin/bash
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"Manual Test Client","version":"1.0.0"}}}';
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | godoctor
```

Passar esse script via pipe para o binário do `godoctor` produz primeiro o resultado de `initialize`, seguido pelo resultado de `tools/list`, que lista corretamente todas as ferramentas do GoDoctor. Entender esse fluxo obrigatório em três etapas foi a chave para resolver meu maior obstáculo inicial, que descreverei na próxima seção.

Para quem está conhecendo o MCP agora, recomendo fortemente a leitura da documentação oficial. Os dois documentos mais cruciais para mim foram as páginas sobre o [ciclo de vida de cliente/servidor](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle) e a [camada de transporte](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports). (Ou, se estiver com preguiça, passe essas URLs para a CLI e deixe que ela leia para você =^_^=)

## A Primeira Conquista: Um Agente que Lê a Documentação

Meu primeiro objetivo era resolver o problema de alucinação de APIs. Como acontece em muitas primeiras tentativas de elaboração de prompts, meu pedido inicial foi simples e um pouco vago:

> "Crie um servidor MCP em Go que tenha uma ferramenta chamada `godoc`. Esta ferramenta deve receber o nome de um pacote e um nome de símbolo opcional e executar o comando `go doc`."

Os resultados não foram nada bons. O agente perdeu muito tempo tentando descobrir quais ferramentas usar e qual protocolo seria o melhor. Até mesmo a sigla "MCP" não era um conceito óbvio para ele; com frequência, ele deduzia que ela significava outras coisas até que esclareci que se tratava de "Model Context Protocol". Ele ficou preso em loops chamando o Google Search e o WebFetch, testando SDKs diferentes, falhando em produzir um exemplo funcional e então mudando para outro SDK, várias e várias vezes. É aqui que o trabalho real de "vibe coding" começa. Trata-se de um processo iterativo de refinamento das instruções. Após algumas horas de tentativa e erro, cheguei a um prompt muito mais eficaz. Aprendi que fornecer recursos específicos e de alta qualidade era o grande segredo.

Aqui está o prompt final, muito melhorado:

> Sua tarefa é criar um servidor de Model Context Protocol (MCP) para expor o comando go doc, dando aos LLMs a capacidade de consultar a documentação de Go. A ferramenta deve se chamar go-doc e deve receber dois argumentos: package_path (obrigatório) e symbol_name (opcional). Para a parte de documentação, use o comando de shell `go doc`. Para a implementação do MCP, você deve usar o SDK oficial de Go para MCP e escrever um servidor MCP pronto para produção que se comunique por meio de um transporte stdio. Você também deve criar um cliente CLI simples para me permitir testar o servidor.
>
> Leia estas referências para coletar informações sobre a tecnologia e a estrutura do projeto antes de escrever qualquer código:
> - https://github.com/modelcontextprotocol/go-sdk
> - https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle
> - https://go.dev/doc/modules/layout

Esse prompt é superior por vários motivos: ele especifica o SDK exato a ser utilizado, dita o transporte (`stdio`) e, acima de tudo, entrega uma lista de leitura ao agente. Ao fornecer links para o código-fonte do SDK e para a especificação do MCP, reduzi drasticamente a tendência do agente de alucinar.

Mesmo com esse prompt aprimorado, o caminho não foi tranquilo. O maior obstáculo surgiu quando tentei usar o transporte `stdio`. Minhas chamadas de ferramentas falhavam sistematicamente com um erro enigmático: `server initialisation is not complete`. Depois de muito debugging doloroso, descobri que o problema não estava no código do meu servidor. A questão era que o transporte `stdio` do MCP exige o handshake específico de três etapas que descrevi acima. Meu cliente estava tentando chamar a ferramenta antes da conclusão do handshake. Essa experiência me ensinou uma lição valiosa: ao criar ferramentas para IA, você não está apenas depurando código, você está depurando o próprio protocolo de conversação.

Com o servidor rodando e falando o protocolo corretamente, a próxima peça do quebra-cabeça era apresentá-lo ao Gemini CLI. Isso é gerenciado por um arquivo `.gemini/settings.json` na raiz do projeto, que indica à CLI quais ferramentas carregar. Adicionei a seguinte configuração a ele:

```json
{
  "mcpServers": {
    "godoctor": {
      "command": "./bin/godoctor"
    }
  }
}
```

Com isso configurado, sempre que eu iniciava o Gemini CLI nesse diretório, ele inicializava automaticamente meu servidor `godoctor` em segundo plano e disponibilizava suas ferramentas para o agente.

## Criando um Loop de Feedback com um Revisor de Código de IA

Com uma ferramenta `godoc` funcional, o próximo passo lógico era ensinar o agente não apenas a ler sobre código, mas a raciocinar sobre sua qualidade. Isso nos levou à ferramenta `code_review`. Dessa vez, a experiência foi muito mais suave, como resultado direto do trabalho prévio que já havíamos feito.

Meu prompt focou no objetivo, e não na implementação:

> Eu quero adicionar uma nova ferramenta ao meu projeto chamada code_review. Esta ferramenta deve usar a Gemini API para analisar código Go e fornecer uma lista de melhorias em formato json de acordo com as melhores práticas aceitas pela comunidade Go. A ferramenta deve receber o conteúdo do código Go e uma dica opcional como entrada...
>
> Use este SDK para chamar o Gemini: https://github.com/googleapis/go-genai

O agente ainda precisava aprender o Go SDK de `genai`, mas dessa vez ele contava com a nossa ferramenta `godoc` na sua caixa de ferramentas. Eu conseguia vê-lo usando a ferramenta para consultar a documentação do SDK, corrigir seus próprios erros e aprender em tempo real. O processo ainda foi iterativo, mas muito mais rápido e eficiente.

O resultado mais importante não foi apenas a ferramenta em si, mas a nova capacidade que ela desbloqueou. Pela primeira vez, eu podia usar a ferramenta para revisar o próprio código que ela gerava, destravando outro patamar de desenvolvimento orientado por IA. **Eu havia criado um loop de feedback positivo.**

Meu fluxo de trabalho agora contava com um novo passo poderoso. Depois que o agente gerava um trecho de código, eu podia pedir imediatamente que ele fizesse uma crítica ao próprio trabalho:

> "Agora, use a ferramenta `code_review` no código que você acabou de escrever e aplique as sugestões."

O agente então analisava a própria saída e a refatorava com base no feedback gerado pela IA. Esse é o verdadeiro poder de construir ferramentas para IA: você não está apenas automatizando tarefas, está criando sistemas para autoaprimoramento.

## O Capítulo Final: Implantando na Nuvem

{{< warning >}}
Se você implantar seu próprio servidor MCP no Cloud Run, certifique-se de configurar a autenticação adequada. **Não faça deploy de um servidor acessível publicamente**, especialmente se ele usar uma chave de API do Gemini. Um endpoint público pode ser explorado por agentes mal-intencionados, potencialmente gerando uma fatura de nuvem gigantesca e inesperada.
{{< /warning >}}

Uma ferramenta local rodando via `stdio` é excelente para uso pessoal, mas o propósito do Model Context Protocol é criar um ecossistema de ferramentas compartilhadas e localizáveis. A próxima fase dessa "side quest" foi transformar o GoDoctor de um binário local no meu notebook em um serviço web escalável usando o [Google Cloud Run](https://cloud.google.com/run?utm_campaign=CDR_0x72884f69_default_b421852297&utm_medium=external&utm_source=blog).

Isso significava ensinar ao agente duas novas habilidades de desenvolvimento em nuvem: como conteinerizar uma aplicação e como implantá-la.

Primeiro, precisávamos migrar do transporte `stdio` para o `HTTP`. Meu prompt foi direto ao ponto, aproveitando o trabalho que já havíamos feito:

> "Chegou a hora de levar nosso servidor para a web. Por favor, refatore o servidor MCP do transporte `stdio` para usar o transporte `streamable HTTP`."

Com o servidor agora falando HTTP, o próximo passo era empacotá-lo para a nuvem. Pedi ao agente para criar um `Dockerfile` multi-stage pronto para produção, que é a forma padrão de criar imagens de contêiner leves e seguras.

> "Por favor, crie um Dockerfile multi-stage que compile o binário Go e o copie para uma imagem mínima `golang:1.24-alpine`."

Com o `Dockerfile` pronto, era hora do deploy. Esse é o momento em que a prova de conceito local se transforma em uma peça real de infraestrutura na nuvem.

> "Agora, por favor, faça o deploy dessa imagem no Cloud Run. Faça o deploy em `us-central1` e use o projeto atualmente configurado no ambiente. Quando terminar, me dê a URL que posso usar para chamar a ferramenta MCP."

O agente forneceu os comandos `gcloud` corretos e, em poucos minutos, o GoDoctor estava no ar na internet. Para finalizar a configuração, precisei avisar ao meu Gemini CLI local sobre o servidor remoto. Isso significou atualizar o arquivo `.gemini/settings.json`, trocando o `command` local pelo `httpUrl` remoto:

```json
{
  "mcpServers": {
    "godoctor": {
      "httpUrl": "https://<your-cloud-run-url>.run.app"
    }
  }
}
```

E, simples assim, minha CLI estava utilizando uma ferramenta implantada e rodando na nuvem. Esse foi o momento em que a prova de conceito pareceu verdadeiramente completa. O processo iterativo de guiar o agente havia compensado todo o esforço, levando uma ideia por todo o ciclo de vida de uma aplicação moderna — de um conceito local a um serviço cloud-native escalável.

Dito isso, para o meu trabalho no dia a dia, continuo usando a versão `stdio` — para uma equipe de uma pessoa só, fazer deploy no Cloud Run é um exagero.

## Minhas Principais Lições do Vibe Coding com o GoDoctor

Esta jornada foi menos sobre escrever código e mais sobre aprender a colaborar de maneira eficaz com uma IA. Minha maior lição foi mudar minha mentalidade de "programadora" para "professora" ou "piloto". Aqui estão algumas das lições mais importantes que aprendi:

*   **Você é o piloto.** A IA às vezes vai propor ações com as quais você não concorda. Não tenha medo de apertar `ESC` para cancelar e fornecer um novo prompt para guiá-la na direção certa.
*   **Mantenha a IA por dentro.** O ideal é deixar o agente executar todo o trabalho, mas às vezes uma edição manual é necessária. Quando você mesma altera o código, o contexto da IA fica desatualizado. Lembre-se de avisar o que você mudou para que ela continue sincronizada com a base de código.
*   **Quando tudo falhar, desligue e ligue de novo.** No caso (nada raro) de a IA travar completamente, a solução clássica de TI faz milagres. No Gemini CLI, isso significa usar o comando `/compress` para enxugar o histórico da conversa ou, em situações extremas, reiniciar a CLI para recomeçar com um contexto limpo.

Ao fornecer ao agente o contexto correto e as ferramentas certas, ele se tornou um parceiro muito mais capaz. A jornada deixou de ser sobre mim tentando fazer com que o código fosse escrito, e passou a ser sobre a construção de um sistema capaz de aprender e se aprimorar continuamente.

## O Que Vem a Seguir?

A jornada com o GoDoctor está longe de terminar. Ele ainda é um projeto experimental, e aprendo mais a cada nova ferramenta e interação. Meu objetivo é continuar evoluindo o projeto para que ele se torne um assistente de desenvolvimento genuinamente útil para desenvolvedores Go em qualquer lugar.

Se você quiser vivenciar essa jornada e construir seu próprio servidor MCP do zero, eu criei um workshop prático para guiar você ao longo do processo. Confira o codelab **[Build an MCP Server with Go and Gemini CLI](https://codelabs.developers.google.com/cloud-gemini-cli-mcp-go)**.

Para quem tiver interesse em como esses conceitos estão sendo aplicados na toolchain oficial do Go, recomendo fortemente a leitura sobre o servidor MCP do `gopls`, que compartilha de muitos desses mesmos objetivos. Você pode encontrar mais informações no [site oficial de documentação de Go](https://tip.golang.org/gopls/features/mcp).

## Recursos e Links

Aqui estão alguns dos principais recursos que mencionei ao longo deste artigo. Espero que sejam tão úteis para você quanto foram para mim.

*   **[Repositório do Projeto GoDoctor](https://github.com/danicat/godoctor):** O código-fonte completo da ferramenta que discutimos.
*   **[Página Oficial do Model Context Protocol](https://modelcontextprotocol.io/):** O melhor ponto de partida para aprender sobre o MCP.
*   **[Especificação do MCP (2025-06-18)](https://modelcontextprotocol.io/specification/2025-06-18):** A especificação técnica completa.
*   **[Documentação do Ciclo de Vida do MCP](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle):** Uma leitura essencial para compreender o handshake entre cliente e servidor.
*   **[Documentação de Transporte do MCP](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports):** Essencial para entender as diferenças entre os transportes `stdio` e `http`.