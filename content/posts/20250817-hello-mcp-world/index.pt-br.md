---
categories:
- Agent Development
date: 2025-08-17 15:00:00+00:00
summary: Baseado na minha keynote na GopherCon UK 2025, este artigo é uma introdução
  ao Model Context Protocol (MCP), explorando seus conceitos centrais, arquitetura
  e os blocos fundamentais — Tools, Prompts e Resources — usados para criar aplicações
  com suporte a IA em Go.
tags:
  - gemini
  - golang
  - keynote
  - mcp
title: "Olá, Mundo MCP: Arquitetura Model Context Protocol em Go"
slug: "hello-mcp-world"
aliases:
  - "/pt-br/posts/20250817-hello-mcp-world/"
description: "Introdução à arquitetura do Model Context Protocol (MCP) em Go baseada na keynote do GopherCon UK 2025. Aborda Hosts, Clients, Servers, Tools, Prompts e Resources."
proficiencyLevel: "Beginner"
dependencies:
  - "Go 1.24+"
  - "MCP Go SDK"
---
## Introdução

Neste artigo, vamos explorar o Model Context Protocol (MCP), um protocolo criado pela Anthropic para padronizar a comunicação entre Large Language Models (LLMs) e aplicações. Este artigo é baseado na [keynote de mesmo nome que apresentei na GopherCon UK](https://speakerdeck.com/danicat/hello-mcp-world) na semana passada.

Para construir um entendimento sólido, começaremos pelos fundamentos, passando pelos principais componentes de arquitetura, camadas de transporte e blocos fundamentais (tools, prompts e resources). Ao longo do caminho, veremos alguns exemplos práticos baseados em servidores que escrevi anteriormente (GoDoctor e Speedgrapher). Por fim, vamos ver como você pode criar seu próprio servidor usando o Go SDK oficial para MCP por meio de um exemplo simples "vibe-coded" com o Gemini CLI.

Seja esta a primeira vez que você ouve falar sobre o protocolo ou se você já implementou um servidor ou dois, este artigo busca trazer informações valiosas para os mais diversos níveis de experiência.

## Um Novo Padrão Nasce

Sempre que falamos sobre padrões, a clássica tirinha do XKCD é a primeira coisa que me vem à mente:

![Standards](image.png)
*Fonte: [xkcd.com](https://xkcd.com/927)*

Curiosamente, esta talvez seja a primeira vez na indústria em que essa piada não se aplica totalmente (pelo menos por enquanto). Para a nossa sorte, a indústria convergiu rapidamente para o MCP como padrão para fornecer contexto a LLMs.

De acordo com a especificação, o MCP é:

> O MCP é um protocolo aberto que padroniza como as aplicações fornecem contexto para grandes modelos de linguagem (LLMs). Pense no MCP como uma porta USB-C para aplicações de IA. Assim como o USB-C oferece uma forma padronizada de conectar seus dispositivos a vários periféricos e acessórios, o MCP oferece uma forma padronizada de conectar modelos de IA a diferentes fontes de dados e ferramentas. O MCP permite criar agentes e fluxos de trabalho complexos sobre LLMs e conecta seus modelos com o mundo.

Embora eu entenda a analogia com o USB-C, prefiro enxergar o MCP como o novo HTTP/REST. Da mesma forma que o HTTP forneceu uma linguagem universal para serviços web se comunicarem, o MCP fornece uma base comum para modelos de IA interagirem com sistemas externos. Como engenheiras e engenheiros de software, passamos praticamente as últimas duas décadas tornando tudo "API-first", permitindo que nossos sistemas de software se tornassem interconectados e impulsionando novos patamares de automação. Talvez não seja pelos próximos 20 anos, mas acredito que nos próximos 5 a 10 anos dedicaremos muito esforço de engenharia para adaptar todos esses sistemas (e criar novos) para que sejam integrados com IA — e o MCP é uma peça central nesse processo.

## Arquitetura do MCP

Olhando para o diagrama abaixo, a arquitetura do MCP pode parecer mais complexa do que realmente é:

![MCP Architecture](image-1.png)
*Fonte: [Especificação do MCP](https://modelcontextprotocol.io/docs/learn/architecture)*

Os principais componentes da arquitetura do MCP são:

*   **MCP Host:** A aplicação principal de IA, como sua IDE ou um agente de código.
*   **MCP Server:** Um processo que fornece acesso a determinada capacidade (por exemplo, ferramentas ou prompts).
*   **MCP Client:** Conecta o host a um servidor específico.

Em essência, uma aplicação host cria e gerencia múltiplos clientes, com cada cliente tendo uma relação 1:1 com um servidor em particular.

## Camadas do MCP

A comunicação acontece em duas camadas:

* **Camada de dados (*data layer*)**: é um protocolo baseado em JSON-RPC. Você pode ver exemplos do formato de mensagens na próxima seção.
* **Camada de transporte (*transport layer*)**: define os canais de comunicação, sendo os principais:
  - Standard I/O (stdio): para servidores locais
  - Streamable HTTPS: para comunicações pela rede (substitui HTTPS+SSE).
  - HTTPS+SSE: depreciado na versão mais recente da especificação por questões de segurança.

A camada de dados é gerenciada pelo SDK e, exceto para fins de testes, você não precisará montar essas mensagens manualmente. A escolha do transporte dependerá do seu caso de uso, mas, em geral, recomendo começar com stdio e adicionar HTTPS mais tarde. Existem até adaptadores open source que convertem servidores MCP de stdio para HTTPS e vice-versa, mas adicionar esse recurso é tão trivial que eu só usaria esses adaptadores para servidores cujo código-fonte eu não controlo.

## Fluxo de Inicialização

O cliente e o servidor realizam um *handshake* para estabelecer uma conexão. Esse processo envolve três mensagens principais:

1. O cliente envia uma requisição `initialize` ao servidor, especificando a versão do protocolo que suporta. (O servidor envia uma mensagem de resposta de inicialização de volta ao cliente.)
2. O cliente confirma a inicialização com uma mensagem `notifications/initialized`.
3. O cliente pode então começar a fazer requisições, como `tools/list`, para descobrir as capacidades do servidor.

É assim que o fluxo de inicialização se parece na rede (ou no pipe stdio) a partir do lado do cliente, usando a representação JSON-RPC:

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}
{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
```

Observe que você não pode simplesmente enviar uma mensagem "tools/list" ou "tools/call" diretamente, ou receberá um erro do tipo "servidor não pronto".

Quando estou desenvolvendo um servidor MCP com a ajuda de um agente de código, como o Gemini CLI, costumo instruí-lo a enviar essas mensagens via shell desta forma:

```sh
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}';
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | ./bin/godoctor
```

Gosto de fazer isso para garantir que a implementação está sólida, pois antes de entender completamente esse fluxo, meus agentes de código frequentemente faziam suposições erradas, como "o servidor precisa de mais tempo para inicializar, então vou adicionar um sleep antes da chamada da ferramenta". Quanto mais cedo você ensinar seu agente de código a se comunicar adequadamente com o servidor MCP que você está desenvolvendo, melhor!

## Os Blocos Fundamentais de um Servidor MCP

Em sua essência, a funcionalidade de um servidor MCP é exposta por meio de três blocos de construção fundamentais, às vezes também chamados de "primitivos" ou "conceitos de servidor":

| Bloco Fundamental | Finalidade               | Quem Controla           | Exemplo no Mundo Real                            |
| :---------------- | :----------------------- | :---------------------- | :----------------------------------------------- |
| **Tools**         | Para ações de IA         | Controlado pelo modelo  | Buscar voos, enviar mensagens, revisar código    |
| **Resources**     | Para dados de contexto   | Controlado pela aplicação| Documentos, calendários, e-mails, dados de clima |
| **Prompts**       | Para templates de interação | Controlado pela pessoa usuária | "Planejar férias", "Resumir minhas reuniões"     |

Vamos examinar mais de perto cada um deles.

### Tools (Ferramentas)

Tools são funções que permitem a um modelo de IA executar ações — por exemplo, expor uma API, banco de dados ou ferramenta de linha de comando.

O servidor que escrevi para experimentar com o conceito de tools chama-se GoDoctor, projetado para fornecer ferramentas que aprimoram as capacidades de LLMs ao escrever código Go. O nome GoDoctor vem de um trocadilho com a ferramenta de linha de comando `go doc`, que exibe a documentação de pacotes Go.

Minha hipótese era de que, ao fornecer a documentação correta, os LLMs alucinariam menos e escreveriam códigos melhores. Ou, no mínimo, teriam os recursos necessários para aprender e autocorrigir seus próprios erros.

A implementação de tools consiste em dois componentes principais: registrar a ferramenta no seu servidor MCP e implementar um *handler*.

O registro é feito usando a função `mcp.AddTool`:

{{< github user="danicat" repo="godoctor" path="internal/tools/get_documentation/get_documentation.go" lang="golang" start="35" end="40" >}}

O *handler* é um adaptador que chama uma API, comando ou função e retorna a resposta de forma compatível com o protocolo (uma struct `mcp.CallToolResult`).

Aqui está o handler para a ferramenta de documentação do GoDoctor:

{{< github user="danicat" repo="godoctor" path="internal/tools/get_documentation/get_documentation.go" lang="golang" start="49" end="86" >}}

### Prompts

Prompts fornecem templates reutilizáveis e parametrizáveis controlados pela pessoa usuária. Frequentemente aparecem como comandos de barra (*slash commands*) em um agente de IA, permitindo invocar um fluxo de trabalho complexo com um comando simples.

Para ver isso em ação, vamos olhar para outro servidor MCP que escrevi, chamado `speedgrapher`, que é uma coleção de prompts e ferramentas para ajudar na minha escrita técnica.

Um dos prompts mais simples no `speedgrapher` é o `/haiku`. Assim como nas tools, o processo envolve definir o prompt e depois implementar um *handler* para ele.

{{< github user="danicat" repo="speedgrapher" path="internal/prompts/haiku.go" lang="golang" start="24" end="54" >}}

### Resources (Recursos)

Resources expõem dados a partir de arquivos, APIs ou bancos de dados, fornecendo o contexto de que uma IA precisa para realizar uma tarefa. Conceitualmente, uma **Tool** serve para executar uma ação, enquanto um **Resource** serve para fornecer informações.

Dito isso, no mundo real ainda não vi uma implementação realmente expressiva de resources, já que a maioria das pessoas desenvolvedoras tem usado tools para expor dados (da mesma forma que você faria em uma API com uma requisição GET). Acredito que este seja um dos casos em que a especificação tentou ser esperta demais, mas talvez vejamos bons usos para resources no futuro, à medida que a comunidade for se acostumando mais com eles.

## Conceitos de Cliente (Client Concepts)

Além dos blocos de construção do servidor, o protocolo também define **Client Concepts** (Conceitos de Cliente), que são capacidades que o servidor pode solicitar ao cliente. Eles incluem:

*   **Sampling:** Permite que um servidor solicite completações de LLM a partir do modelo do cliente. Isso é promissor sob a perspectiva de segurança e faturamento, já que autores de servidores não precisam usar suas próprias chaves de API para chamar modelos.
*   **Roots:** Um mecanismo para o cliente comunicar limites no sistema de arquivos, informando ao servidor em quais diretórios ele tem permissão para operar.
*   **Elicitation:** Uma forma estruturada para o servidor solicitar informações específicas da pessoa usuária, pausando sua operação para coletar a entrada necessária.

Este é outro caso em que a maioria das aplicações do mundo real que explorei ainda não alcançou a especificação, incluindo tanto servidores quanto clientes. Pode levar algum tempo até que esses recursos estejam amplamente disponíveis. É um dos desafios de trabalhar com tecnologia de ponta (*bleeding edge*)... Por exemplo, o Gemini CLI adicionou suporte a roots há cerca de uma semana: https://github.com/google-gemini/gemini-cli/pull/5856

## Demonstração Prática: Vibe Coding de um Servidor MCP

Aqui está um prompt que você pode passar para o seu agente de código favorito para gerar um servidor no estilo "Hello World". Como os agentes atuais são não determinísticos, pode ser que não funcione 100% de primeira e você precise orientar o LLM com alguns prompts adicionais após o inicial, mas é um ótimo ponto de partida:

```text
Your task is to create a Model Context Protocol (MCP) server to expose a "hello world" tool. For the MCP implementation, you should use the official Go SDK for MCP and use the stdio transport.

Read these references to gather information about the technology and project structure before writing any code:
- https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/refs/heads/main/README.md
- https://go.dev/doc/modules/layout

To test the server, use shell commands like these:
`( 
	echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}';
	echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
	echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | ./bin/hello`
```

Se o agente concluir a tarefa com sucesso, peça a ele para executar um `tools/call` para sua nova ferramenta para conferir os resultados!

## Um Olhar Para o Futuro

A comunidade Go está investindo ativamente no ecossistema MCP. Dois projetos fundamentais para acompanhar são:

*   **O Go SDK para MCP:** O SDK oficial que usei na demonstração, fruto de uma parceria entre o Google e a Anthropic. Ele ainda é experimental (a versão atual é a 0.20), mas é funcional e está em desenvolvimento ativo. Você pode encontrá-lo em [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk).
*   **Suporte a MCP no `gopls`:** O language server do Go, `gopls`, está recebendo suporte a MCP para fornecer recursos avançados de codificação em Go para modelos de IA. O projeto ainda está em estágios iniciais, e você pode acompanhar o progresso em [tip.golang.org/gopls/features/mcp](https://tip.golang.org/gopls/features/mcp).

## Servidores MCP Úteis

Aqui estão alguns servidores notáveis construídos pela comunidade:

*   **Playwright:** Mantido pela Microsoft, este servidor permite que um agente de IA navegue por páginas web, capture capturas de tela e automatize tarefas no navegador. Você pode encontrá-lo em [https://github.com/microsoft/playwright-mcp](https://github.com/microsoft/playwright-mcp).
*   **Context7:** Similar ao GoDoctor, este servidor fornece documentação aos modelos para reduzir alucinações e melhorar as respostas. Ele obtém documentação a partir de um repositório colaborativo (*crowdsourced*). Saiba mais em [https://context7.com/](https://context7.com/).

## Que Tal Construir o Seu Próprio?

O Model Context Protocol fornece uma forma padronizada de estender as capacidades dos agentes de IA. Ao construir seus próprios servidores, você pode criar assistentes especializados e conscientes do contexto, sob medida para seus fluxos de trabalho específicos.

Se quiser começar, criei um Google Codelab que guiará você passo a passo no processo de construir seu próprio servidor MCP do zero:

[**Como Construir um Assistente de Programação com Gemini CLI, MCP e Go**](https://codelabs.developers.google.com/codelabs/gemini-cli-mcp-go)

## Considerações Finais

Espero que tenha gostado deste artigo. Se você tiver alguma dúvida ou comentário, fique à vontade para entrar em contato na seção de comentários abaixo ou em qualquer uma das minhas redes sociais. Muito obrigada!