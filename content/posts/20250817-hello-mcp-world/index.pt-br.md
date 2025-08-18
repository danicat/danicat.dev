---
title: "Olá, Mundo MCP!"
date: 2025-08-17T15:00:00Z
categories: ["AI & Development"]
tags: ["mcp", "gemini", "golang"]
summary: "Uma introdução ao Model Context Protocol (MCP), explorando seus conceitos centrais, arquitetura e os blocos de construção—Tools, Prompts e Resources—usados para criar aplicações habilitadas para IA com Go."
---

> Este artigo é baseado na keynote que apresentei na Gophercon UK 2025 em 14 de agosto. Para os slides da keynote, por favor, verifique este [link](https://speakerdeck.com/danicat/hello-mcp-world).

Neste artigo, vamos explorar o Model Context Protocol (MCP), um protocolo desenvolvido pela Anthropic para padronizar as comunicações entre Large Language Models (LLMs) e aplicações.

Como uma boa prática, vamos começar com algumas definições, depois explicar os principais componentes arquitetônicos ilustrando com exemplos práticos de servidores que implementei ao longo da minha própria jornada de aprendizado. Finalmente, vamos ver como você pode escrever seu próprio servidor usando o SDK Go para MCP através de um exemplo simples, "vibe-coded" usando o Gemini CLI.

Seja esta a primeira vez que você ouve falar sobre este protocolo, ou talvez você já tenha escrito um ou dois servidores, este artigo visa fornecer informações úteis para vários níveis de experiência.

## Nasce um Novo Padrão

Sempre que falamos sobre padrões, esta tirinha do XKCD é a primeira coisa que me vem à mente:

![Padrões](image.png)
*Fonte: [xkcd.com](https://xkcd.com/927)*

Engraçado o suficiente, esta pode ser a primeira vez na indústria que essa piada não se aplica totalmente (pelo menos por enquanto). Para nossa sorte, a indústria convergiu rapidamente para o MCP como o padrão para adicionar contexto aos LLMs.

Da especificação, o MCP é:

> O MCP é um protocolo aberto que padroniza como as aplicações fornecem contexto para large language models (LLMs). Pense no MCP como uma porta USB-C para aplicações de IA. Assim como o USB-C fornece uma maneira padronizada de conectar seus dispositivos a vários periféricos e acessórios, o MCP fornece uma maneira padronizada de conectar modelos de IA a diferentes fontes de dados e ferramentas. O MCP permite que você construa agentes e fluxos de trabalho complexos sobre LLMs e conecta seus modelos com o mundo.

Embora eu entenda a analogia com o USB-C, prefiro pensar no MCP como o novo HTTP/REST. Como engenheiros, passamos aproximadamente as últimas duas décadas tornando tudo "API-first", permitindo que nossos sistemas de software se tornem interconectados e impulsionando novos níveis de automação. Talvez não seja para os próximos 20 anos, mas acredito que nos próximos 5 a 10 anos gastaremos bastante poder de engenharia para adaptar todos esses sistemas (e criar novos) para se tornarem habilitados para IA, e o MCP é um componente chave desse processo.

## Arquitetura MCP

Este diagrama representa a arquitetura MCP:

![Arquitetura MCP](image-1.png)
*Fonte: [Especificação MCP](https://modelcontextprotocol.io/docs/learn/architecture)*

Os principais componentes da arquitetura MCP são:

*   **MCP Host:** A aplicação de IA principal, como seu IDE ou um agente de codificação. O host se comunica com servidores MCP usando clientes.
*   **MCP Server:** Um processo que fornece acesso a alguma capacidade.
*   **MCP Client:** A ponte que conecta o host a um único servidor.

## Camadas MCP

A comunicação acontece em duas camadas:

* **Camada de Dados**: é um protocolo baseado em JSON-RPC. Você pode ver exemplos do formato da mensagem na próxima seção.
* **Camada de Transporte**: define os canais de comunicação, sendo os principais:
  - Standard I/O: para servidores locais
  - Streamable HTTPS: para comunicações pela rede. (Substitui HTTPS+SSE).
  - HTTPS+SSE: obsoleto na última versão da especificação por questões de segurança.

## Fluxo de Inicialização

É assim que o fluxo de inicialização se parece na transmissão usando a representação JSON-RPC:

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}
{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
```
Por favor, note que você não pode simplesmente enviar uma mensagem "tools/list" ou "tools/call" diretamente, ou você receberá um erro do tipo "servidor não pronto".

Se estou codificando um servidor MCP através de um agente de codificação, como por exemplo o Gemini CLI, costumo instruí-los a enviar essas mensagens via shell assim:

```sh
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}';
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | ./bin/godoctor
```

Gosto de fazer isso para garantir que a implementação esteja correta, pois antes de entender completamente esse fluxo, meus agentes de codificação muitas vezes faziam suposições erradas como "o servidor precisa de mais tempo para iniciar, então vou adicionar um sleep antes da chamada da ferramenta". Quanto antes você ensinar seu agente de codificação a se comunicar adequadamente com o servidor MCP que você está desenvolvendo, melhor!

## Os Blocos de Construção de um Servidor MCP

O protocolo define três blocos de construção fundamentais, às vezes também chamados de "primitivos" ou "conceitos de servidor":

| Bloco de Construção | Propósito                | Quem Controla           | Exemplo do Mundo Real                            |
| :------------------ | :----------------------- | :---------------------- | :----------------------------------------------- |
| **Tools**           | Para ações de IA         | Controlado pelo Modelo  | Pesquisar voos, enviar mensagens, revisar código |
| **Resources**       | Para dados de contexto   | Controlado pela Aplicação | Documentos, calendários, e-mails, dados do tempo |
| **Prompts**         | Para modelos de interação| Controlado pelo Usuário | "Planejar umas férias," "Resumir minhas reuniões" |

Vamos dar uma olhada mais de perto em cada um deles.

### Tools

Tools são funções que permitem que um modelo de IA execute ações, por exemplo, expondo uma API, banco de dados ou ferramenta de linha de comando.

O servidor que escrevi para experimentar o conceito de tools é chamado GoDoctor, que é projetado para fornecer ferramentas para melhorar as capacidades dos LLMs na escrita de código Go. O nome GoDoctor vem de um jogo de palavras com a ferramenta de linha de comando "go doc" que expõe a documentação sobre pacotes Go.

Minha hipótese era que, ao fornecer a documentação correta, os LLMs alucinariam menos e escreveriam um código melhor. Ou, pelo menos, teriam os recursos para aprender e autocorrigir seus erros.

A implementação de tools consiste em dois componentes principais: registrar a tool com seu servidor MCP e implementar um handler.

O registro é feito usando a função `mcp.AddTool`:

{{< github user="danicat" repo="godoctor" path="internal/tools/get_documentation/get_documentation.go" lang="golang" start="35" end="40" >}}

O handler é um adaptador que chama uma API, comando ou função e retorna a resposta de uma forma compatível com o protocolo (uma struct `mcp.CallToolResult`).

Aqui está o handler para a ferramenta de documentação do GoDoctor:

{{< github user="danicat" repo="godoctor" path="internal/tools/get_documentation/get_documentation.go" lang="golang" start="49" end="86" >}}

### Prompts

Prompts fornecem modelos reutilizáveis e controlados pelo usuário que podem ser parametrizados. Eles geralmente aparecem como comandos de barra em um agente de IA, permitindo que um usuário invoque um fluxo de trabalho complexo com um comando simples.

Para ver isso em ação, vamos olhar para um servidor MCP diferente que escrevi chamado `speedgrapher`, que é uma coleção de prompts e tools para ajudar na minha escrita técnica.

Um dos prompts mais simples no `speedgrapher` é `/haiku`. Assim como com as tools, o processo envolve a definição do prompt e, em seguida, a implementação de um handler para ele.

{{< github user="danicat" repo="speedgrapher" path="internal/prompts/haiku.go" lang="golang" start="24" end="54" >}}

### Resources

Resources expõem dados de arquivos, APIs ou bancos de dados, fornecendo o contexto que uma IA precisa para executar uma tarefa.

Para ser honesta, ainda não encontrei um exemplo prático onde os resources sejam mais adequados, pois a maioria das pessoas simplesmente opta por usar tools para expor dados. Não ajuda o fato de que os agentes que estou usando também ainda não implementaram resources, então vou me abster de aprofundar mais neste assunto até encontrar casos de uso adequados.

Este artigo não estaria completo sem mencioná-los, no entanto.

## Conceitos do Cliente

Do outro lado da conexão, o protocolo também define **Conceitos do Cliente**, que são capacidades que o servidor pode solicitar do cliente. Estes incluem:

*   **Sampling:** Permite que um servidor solicite conclusões de LLM do modelo do cliente. Isso é promissor do ponto de vista de segurança e faturamento, já que os autores de servidores não precisam usar suas próprias chaves de API para chamar modelos.
*   **Roots:** Um mecanismo para um cliente comunicar limites do sistema de arquivos, dizendo a um servidor em quais diretórios ele tem permissão para operar.
*   **Elicitation:** Uma maneira estruturada para um servidor solicitar informações específicas do usuário, pausando sua operação para coletar a entrada quando necessário.

Assim como para os Resources na seção anterior, ainda não explorei essas capacidades, mas estou ansiosa para experimentar o sampling em seguida.

## Demonstração ao Vivo: Vibe Coding um Servidor MCP

Aqui está um prompt que você pode dar ao seu agente de codificação favorito para produzir um tipo de servidor "Hello World". Como os agentes hoje em dia não são determinísticos, pode não funcionar 100% na primeira tentativa e você pode precisar guiar o LLM com alguns prompts extras após o inicial, mas é um bom começo:

> Sua tarefa é criar um servidor Model Context Protocol (MCP) para expor uma ferramenta "hello world". Para a implementação do MCP, você deve usar o SDK Go oficial para MCP e usar o transporte stdio.
>
> Leia estas referências para coletar informações sobre a tecnologia e a estrutura do projeto antes de escrever qualquer código:
> - https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/refs/heads/main/README.md
> - https://go.dev/doc/modules/layout
>
> Para testar o servidor, use comandos de shell como estes:
> `( echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}' ; echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'; echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'; ) | ./bin/hello`

Se o agente for bem-sucedido em completar esta tarefa, peça a ele para executar um "method tools/call" para sua nova ferramenta para ver os resultados!

## Uma Espiada no Futuro

A comunidade Go está investindo ativamente no ecossistema MCP. Dois projetos importantes para observar são:

*   **O Go SDK para MCP:** O SDK oficial que usei na demonstração, que é uma parceria entre o Google e a Anthropic. Você pode encontrá-lo em [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk).
*   **Suporte MCP para `gopls`:** O servidor de linguagem Go, `gopls`, está ganhando suporte MCP, o que trará integrações de IA ainda mais profundas na experiência de desenvolvimento Go. Você pode acompanhar seu progresso em [tip.golang.org/gopls/features/mcp](https://tip.golang.org/gopls/features/mcp).

## Servidores MCP Úteis

O ecossistema de servidores MCP está crescendo. Além dos que eu construí, existem muitos outros disponíveis que você pode usar para aprimorar seus fluxos de trabalho. Aqui estão alguns exemplos notáveis:

*   **Playwright:** Mantido pela Microsoft, este servidor permite que um agente de IA navegue em páginas da web, tire screenshots e automatize tarefas do navegador. Você pode encontrá-lo em [https://github.com/microsoft/playwright-mcp](https://github.com/microsoft/playwright-mcp).
*   **Context7:** Este servidor recupera documentação de um repositório crowdsourced, fornecendo outra rica fonte de contexto para seu agente. Saiba mais em [https://context7.com/](https://context7.com/).

## Que Tal Construir o Seu Próprio?

O Model Context Protocol fornece uma maneira padronizada de estender as capacidades dos agentes de IA. Ao construir seus próprios servidores, você pode criar assistentes especializados e cientes do contexto, adaptados aos seus fluxos de trabalho específicos.

Se você quiser começar, criei um Google Codelab que o guiará pelo processo de construção de seu próprio servidor MCP do zero.

[**Como Construir um Assistente de Codificação com Gemini CLI, MCP e Go**](https://codelabs.developers.google.com/codelabs/gemini-cli-mcp-go)

## Palavras Finais

Espero que você tenha gostado deste artigo. Se tiver alguma dúvida ou comentário, sinta-se à vontade para entrar em contato na seção de comentários abaixo ou em qualquer uma das minhas redes sociais. Obrigada!
