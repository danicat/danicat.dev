---
title: "Como Construir um Servidor MCP com Gemini CLI e Go"
date: 2025-07-29
author: "Daniela Petruzalek"
tags: ["go", "gemini-cli", "mcp", "ai", "codelab"]
categories: ["Tutoriais"]
summary: "Um guia passo a passo sobre como eu construí o GoDoctor, um assistente de desenvolvimento Go com tecnologia de IA, usando o Model Context Protocol (MCP) e o Gemini CLI."
---

{{< translation-notice >}}

### Introdução

Como muitos de vocês, tenho mergulhado fundo no desenvolvimento assistido por IA. A jornada é muitas vezes uma montanha-russa de momentos "uau" e obstáculos frustrantes. Esta é a história de uma dessas jornadas — uma história que começou comigo tentando construir uma coisa, sendo completamente desviada por um problema frustrante, e terminando com uma ferramenta que melhorou fundamentalmente meu fluxo de trabalho assistido por IA.

Meu objetivo original era construir um servidor de Model Context Protocol (MCP) para o [osquery](https://www.osquery.io/), uma ferramenta que permite consultar o estado de uma máquina usando SQL. Eu estava animada para usar o Gemini CLI para me ajudar a escrever o código em Go. No entanto, rapidamente me deparei com um obstáculo. O código em Go que o agente produzia muitas vezes não era idiomático. Cometia erros de iniciante, criava abstrações excessivas e frequentemente "alucinava" APIs inteiras que não existiam. Minha suspeita era que o modelo subjacente não havia sido treinado no mais novo [Go SDK for MCP](https://github.com/modelcontextprotocol/go-sdk), então preferia inventar respostas a admitir que não sabia.

Essa experiência me levou a uma percepção fundamental: em vez de lutar contra a ferramenta, eu poderia *ensiná-la*. Decidi pausar meu projeto do osquery e embarcar em uma "missão secundária": construir um servidor MCP dedicado cujo único propósito era ser um especialista em desenvolvimento Go. Este projeto paralelo, que acabei chamando de [GoDoctor](https://github.com/danicat/godoctor), forneceria as ferramentas que o Gemini CLI precisava para escrever um código Go melhor.

Neste post, vou guiá-los pela história da construção do GoDoctor. Isto é menos um tutorial tradicional e mais uma jornada "orientada por prompts". Vamos nos concentrar em como traduzir os requisitos do projeto em prompts eficazes e guiar uma IA através dos detalhes da implementação, aprendendo com os erros inevitáveis ao longo do caminho.

### Preparando o Terreno: O `GEMINI.md`

Antes de escrever uma única linha de código, o primeiro passo foi estabelecer as regras básicas. Embora o `GEMINI.md` seja um arquivo específico para o Gemini CLI, a prática de criar um arquivo de contexto é um padrão para muitos agentes de codificação de IA (por exemplo, o Jules usa `AGENTS.md` e o Claude usa `CLAUDE.md`). Na verdade, há um esforço emergente para padronizar isso com um arquivo chamado [`AGENT.md`](https://ampcode.com/AGENT.md). Este arquivo é crucial porque fornece à IA uma compreensão fundamental dos padrões do seu projeto e das suas expectativas para o comportamento dela.

Como este projeto era totalmente novo, eu ainda não tinha detalhes arquitetônicos específicos para compartilhar. Portanto, comecei com um conjunto de diretrizes genéricas focadas na criação de código Go idiomático e de alta qualidade. À medida que um projeto evolui, é comum adicionar instruções mais específicas sobre a estrutura do projeto, comandos de build ou bibliotecas principais. Para um exemplo de um arquivo mais específico do projeto, você pode ver o `GEMINI.md` que uso para o meu projeto [`testquery`](https://github.com/danicat/testquery/blob/main/GEMINI.md).

Aqui está o `GEMINI.md` inicial que serviu como nossa constituição para a IA nesta jornada:

```markdown
# Go Development Guidelines
All code contributed to this project must adhere to the following principles.

### 1. Formatting
All Go code **must** be formatted with `gofmt` before being submitted.

### 2. Naming Conventions
- **Packages:** Use short, concise, all-lowercase names.
- **Variables, Functions, and Methods:** Use `camelCase` for unexported identifiers and `PascalCase` for exported identifiers.
- **Interfaces:** Name interfaces for what they do (e.g., `io.Reader`), not with a prefix like `I`.

### 3. Error Handling
- Errors are values. Do not discard them.
- Handle errors explicitly using the `if err != nil` pattern.
- Provide context to errors using `fmt.Errorf("context: %w", err)`.

### 4. Simplicity and Clarity
- "Clear is better than clever." Write code that is easy to understand.
- Avoid unnecessary complexity and abstractions.
- Prefer returning concrete types, not interfaces.

### 5. Documentation
- All exported identifiers (`PascalCase`) **must** have a doc comment.
- Comments should explain the *why*, not the *what*.

# Agent Guidelines
- **Reading URLs:** ALWAYS read URLs provided by the user. They are not optional.
```

Este arquivo estabelece uma base para qualidade e estilo desde o início.

### Entendendo o Model Context Protocol (MCP)

No coração deste projeto está o Model Context Protocol (MCP). Algumas pessoas o descrevem como um "padrão USB para ferramentas de LLM", mas eu gosto de pensar de outra forma: **o que o HTTP e o REST fizeram pela padronização de APIs web, o MCP está fazendo pelas ferramentas de LLM.** Assim como o REST forneceu uma arquitetura previsível que desbloqueou um ecossistema massivo de serviços web, o MCP está fornecendo uma linguagem comum muito necessária para o mundo dos agentes de IA. É um protocolo baseado em JSON-RPC que cria um terreno comum, permitindo que qualquer agente que "fale" MCP descubra e use qualquer ferramenta compatível sem a necessidade de uma integração personalizada e única.

O protocolo define diferentes maneiras para o agente e o servidor de ferramentas se comunicarem, conhecidas como transportes. Os dois mais comuns são:
*   **HTTP:** O familiar modelo de requisição/resposta, ideal para ferramentas implantadas como serviços remotos (por exemplo, no Cloud Run).
*   **stdio:** Um transporte mais simples que usa a entrada e saída padrão, perfeito para executar uma ferramenta como um processo local em sua máquina.

Com o transporte `stdio`, o agente e o servidor trocam uma série de mensagens JSON-RPC. O processo começa com um handshake crucial de três vias para estabelecer a conexão. Somente após a conclusão deste handshake, o cliente pode começar a fazer chamadas de ferramentas.

A sequência se parece com isto:

![MCP stdio sequence diagram](images/mcp-stdio-sequence-diagram.png)
*<p align="center">Figura 1: Diagrama de sequência do transporte stdio da documentação oficial do MCP (2025-06-18).</p>*

Aqui está como as mensagens JSON para esse handshake inicial se parecem, com base na especificação oficial:

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
O servidor responde com um resultado, confirmando a versão do protocolo e anunciando suas capacidades e informações. Esta é a resposta real do binário `godoctor`:
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
Finalmente, o cliente confirma que a configuração está completa enviando uma notificação `initialized`. Note que esta é uma notificação, então não tem campo `id`, e o método está em um namespace.
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized",
  "params": {}
}
```

Uma vez que esta troca esteja completa, a sessão é estabelecida, e o cliente pode prosseguir com as chamadas de ferramentas. Por exemplo, para pedir ao servidor uma lista de suas ferramentas disponíveis, você pode enviar uma requisição `tools/list`. A chave é que todas as três mensagens do handshake devem ser enviadas na ordem correta antes desta requisição.

Você pode ver a sequência completa em ação com este script de shell:
```bash
#!/bin/bash
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"Manual Test Client","version":"1.0.0"}}}';
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | godoctor
```

Enviar este script para o binário `godoctor` produz primeiro o resultado de `initialize`, seguido pelo resultado de `tools/list`, que lista corretamente todas as ferramentas do GoDoctor. Entender este fluxo de três passos foi a chave para resolver meu maior obstáculo inicial, que descreverei na próxima seção.

Para qualquer pessoa nova no MCP, eu recomendo fortemente a leitura da documentação oficial. Os dois documentos que foram mais cruciais para mim foram as páginas sobre o [ciclo de vida cliente/servidor](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle) e a [camada de transporte](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports). (Ou se você estiver se sentindo com preguiça, dê essas URLs para o CLI e faça ele ler para você =^_^=)

### O Primeiro Avanço: Um Agente Que Lê a Documentação

Meu primeiro objetivo era resolver o problema da alucinação de API. Como muitas primeiras tentativas de prompt, minha requisição inicial foi simples e um pouco vaga:

> "Crie um servidor MCP em Go que tenha uma ferramenta chamada `godoc`. Esta ferramenta deve receber um nome de pacote e um nome de símbolo opcional e executar o comando `go doc`."

Os resultados não foram ótimos. O agente passou muito tempo tentando descobrir quais ferramentas usar e qual protocolo era o melhor. Até mesmo a sigla "MCP" não era um conceito óbvio para ele; muitas vezes inferia que significava coisas diferentes antes de eu esclarecer que era "Model Context Protocol". Ele ficou preso em ciclos chamando as ferramentas Google Search e WebFetch, tentando diferentes SDKs, falhando em produzir um exemplo funcional, e então mudando para outro SDK, repetidamente. É aqui que o verdadeiro trabalho de "vibe coding" começa. É um processo iterativo de refinar suas instruções. Após algumas horas de tentativa e erro, cheguei a um prompt muito mais eficaz. Aprendi que fornecer recursos específicos e de alta qualidade era a chave.

Aqui está o prompt final, muito melhorado:

> Sua tarefa é criar um servidor de Model Context Protocol (MCP) para expor o comando go doc, dando aos LLMs a capacidade de consultar a documentação do Go. A ferramenta deve se chamar go-doc e deve receber dois argumentos: package_path (obrigatório) e symbol_name (opcional). Para a parte da documentação, use o comando de shell `go doc`. Para a implementação do MCP, você deve usar o Go SDK oficial para MCP e escrever um servidor MCP pronto para produção que se comunique através de um transporte stdio. Você também deve criar um cliente CLI simples para me permitir testar o servidor.
>
> Leia estas referências para coletar informações sobre a tecnologia e a estrutura do projeto antes de escrever qualquer código:
> - https://github.com/modelcontextprotocol/go-sdk
> - https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle
> - https://go.dev/doc/modules/layout

Este prompt é melhor por várias razões: ele especifica o SDK exato a ser usado, dita o transporte (`stdio`) e, o mais importante, dá ao agente uma lista de leitura. Ao fornecer links para o código-fonte do SDK e a especificação do MCP, reduzi drasticamente a tendência do agente a alucinar.

Mesmo com este prompt melhor, a jornada não foi tranquila. O maior obstáculo apareceu quando tentei usar o transporte `stdio` simples. Minhas chamadas de ferramenta falhavam consistentemente com um erro enigmático: `server initialisation is not complete`. Após muita depuração dolorosa, descobri que o problema não estava no código do meu servidor. A questão era que o transporte `stdio` do MCP requer o handshake específico de três passos que detalhei acima. Meu cliente estava tentando chamar a ferramenta antes que o handshake estivesse completo. Essa experiência me ensinou uma lição valiosa: ao construir ferramentas para IA, você não está apenas depurando código, está depurando o próprio protocolo de conversação.

### Criando um Loop de Feedback com um Revisor de Código de IA

Com uma ferramenta `godoc` funcional, o próximo passo lógico era ensinar o agente não apenas a ler sobre código, mas a raciocinar sobre sua qualidade. Isso levou à ferramenta `code_review`. A experiência desta vez foi muito mais tranquila, um resultado direto do trabalho que já havíamos feito.

Meu prompt estava focado no objetivo, não na implementação:

> Eu quero adicionar uma nova ferramenta ao meu projeto chamada code_review. Esta ferramenta deve usar a API Gemini para analisar o código Go e fornecer uma lista de melhorias em formato json de acordo com as melhores práticas aceitas pela comunidade Go. A ferramenta deve receber o conteúdo do arquivo Go e uma dica opcional...
>
> Use este SDK para chamar o Gemini: https://github.com/googleapis/go-genai

O agente ainda teve que aprender o `genai` Go SDK, mas desta vez ele tinha nossa ferramenta `godoc` em sua caixa de ferramentas. Eu podia vê-lo usando a ferramenta para procurar o SDK, corrigir seus próprios erros e aprender em tempo real. O processo ainda era iterativo, mas foi significativamente mais rápido e eficiente.

O resultado mais importante não foi apenas a ferramenta em si, mas a nova capacidade que ela desbloqueou. Pela primeira vez, eu podia usar a ferramenta para revisar seu próprio código, desbloqueando outro nível de desenvolvimento orientado por IA. **Eu havia criado um loop de feedback positivo.**

Meu fluxo de trabalho agora tinha um novo passo poderoso. Depois que o agente gerava um novo trecho de código, eu podia imediatamente pedir para ele criticar seu próprio trabalho:

> "Agora, use a ferramenta `code_review` no código que você acabou de escrever e aplique as sugestões."

O agente então analisaria sua própria saída e a refatoraria com base no feedback gerado pela IA. Este é o verdadeiro poder de construir ferramentas para IA: você não está apenas automatizando tarefas, está criando sistemas para autoaperfeiçoamento.

### O Capítulo Final: Implantando na Nuvem

Uma ferramenta local é ótima, mas o verdadeiro poder do MCP vem da implantação de ferramentas como serviços escaláveis. A fase final do projeto foi containerizar o GoDoctor e implantá-lo no Cloud Run.

Primeiro, eu instruí o agente a refatorar o servidor de `stdio` para o transporte `streamable HTTP`. Em seguida, pedi para ele criar um `Dockerfile` de múltiplos estágios e pronto para produção.

> Por favor, crie um Dockerfile de múltiplos estágios que compile o binário Go e o copie para uma imagem golang mínima como golang:1.24-alpine.

Finalmente, era hora da implantação.

> Agora, por favor, implante esta imagem no Cloud Run e me retorne uma URL que eu possa usar para chamar a ferramenta MCP. Implante-a em us-central1 e use o projeto atualmente configurado no ambiente.

O agente forneceu os comandos `gcloud` corretos e, após alguns minutos, o GoDoctor estava ativo na internet, acessível a qualquer cliente compatível com MCP.

### Minhas Principais Lições sobre Vibe Coding com o GoDoctor

Esta jornada foi menos sobre escrever código e mais sobre aprender a colaborar eficazmente com uma IA. Minha maior lição foi mudar minha mentalidade de "programadora" para "professora" ou "piloto". Aqui estão algumas das lições mais importantes que aprendi:

*   **Você é o piloto.** A IA às vezes proporá ações com as quais você não concorda. Não tenha medo de pressionar `ESC` para cancelar e fornecer um novo prompt para guiá-la na direção certa.
*   **Lembre, não repita.** Assim como os humanos, a IA pode esquecer detalhes em uma conversa longa. Se ela esquecer uma instrução, um simples lembrete ("Lembre-se, estamos usando o transporte stdio") geralmente é o suficiente.
*   **Se precisar codificar, avise a IA.** Tente fazer com que a IA realize todo o trabalho. Mas se você fizer uma alteração manual, avise a IA sobre isso para que ela possa atualizar seu contexto.
*   **Na dúvida, reinicie.** No caso raro de a IA ficar presa, o comando `/compress` ou mesmo reiniciar o CLI com um contexto limpo pode fazer maravilhas.

Ao dar ao agente o contexto certo e as ferramentas certas, ele se tornou um parceiro muito mais capaz. A jornada se transformou de eu simplesmente tentando fazer com que o código fosse escrito, para eu construindo um sistema que pudesse aprender e se autoaperfeiçoar.

### O Que Vem a Seguir?

A jornada com o GoDoctor está longe de terminar. Ainda é um projeto experimental, e estou aprendendo mais a cada nova ferramenta e cada nova interação. Meu objetivo é continuar a evoluí-lo para um assistente de codificação genuinamente útil para desenvolvedores Go em todos os lugares.

Para aqueles interessados em como esses conceitos estão sendo aplicados na toolchain oficial do Go, eu recomendo fortemente a leitura sobre o servidor `gopls` MCP, que compartilha muitos dos mesmos objetivos. Você pode encontrar mais informações no [site oficial de documentação do Go](https://tip.golang.org/gopls/features/mcp).

### Recursos e Links

Aqui estão alguns dos principais recursos que mencionei ao longo deste post. Espero que sejam tão úteis para você quanto foram para mim.

*   **[Repositório do Projeto GoDoctor](https://github.com/danicat/godoctor):** O código-fonte completo da ferramenta que discutimos.
*   **[Página Inicial do Model Context Protocol](https://modelcontextprotocol.io/):** O melhor ponto de partida para aprender sobre o MCP.
*   **[Especificação do MCP (2025-06-18)](https://modelcontextprotocol.io/specification/2025-06-18):** A especificação técnica completa.
*   **[Documentação do Ciclo de Vida do MCP](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle):** Uma leitura crucial para entender o handshake cliente/servidor.
*   **[Documentação de Transporte do MCP](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports):** Essencial para entender a diferença entre os transportes `stdio` e `http`.
