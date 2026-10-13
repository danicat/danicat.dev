---
title: "Criando Ferramentas Visuais com Antigravity UI Extensions"
date: 2026-10-13 10:00:00+01:00
draft: false
categories:
  - Agentic Coding
tags:
  - agent-skills
  - antigravity
  - gemini
  - tutorial
  - vibe-coding
heroStyle: big
slug: "building-sidecars-and-ui-extensions-in-antigravity"
description: "Conheça o novo recurso de UI Extensions no Google Antigravity 2.22.0 e aprenda a criar ferramentas visuais bidirecionais no painel lateral."
summary: "Veja como as UI Extensions no Google Antigravity 2.22.0 permitem criar áreas de trabalho visuais, de telas de desenho em camadas a depuradores de jogos 2D, no painel lateral."
proficiencyLevel: "Intermediate"
dependencies:
  - "Google Antigravity 2.22.0+"
video: "https://www.youtube.com/watch?v=A2zOv5Xoh0M"
software:
  name: "Paint Studio (agy-paint-studio)"
  category: "DeveloperApplication"
  os: "Cross-platform"
  repo: "https://github.com/danicat/agy-paint-studio"
---

Eu costumo ser bem pé no chão nos meus posts, mas desta vez estou *realmente* empolgada. Na semana passada, a equipe do Antigravity lançou o [Antigravity 2.22.0](https://antigravity.google/releases?tab=hub&version=2.22.0), que traz um recurso novo que explodiu a minha cabeça: o suporte a **UI Extensions** em plugins do Antigravity. Por que isso é importante? Porque muda toda a experiência de desenvolvimento com agentes de programação para melhor. Vamos ver como.

As UI Extensions permitem renderizar interfaces de usuário personalizadas dentro do painel lateral do Antigravity. O principal motivo pelo qual isso é incrível é que agora podemos ampliar a interface do Antigravity com todo tipo de recurso, trazendo para dentro da ferramenta funcionalidades que antes exigiriam aplicativos separados para executar, como edição de imagens, testes de software, mineração de logs e análise de dados.

Por exemplo, imagine que você está desenhando a interface do seu próximo aplicativo. Explicar isso em texto quase sempre é uma experiência horrível. Com um desenho você consegue ser muito mais preciso do que com texto e o Gemini, sendo multimodal, entende o desenho nativamente. Antes das UI Extensions, eu abria um editor de imagens ou uma ferramenta de diagramas como o [Excalidraw](https://excalidraw.com/), fazia o meu rascunho e depois copiava e colava no meu prompt. Não era uma experiência terrível em si, mas era meio trabalhosa. Agora, com as UI Extensions, posso ter uma ferramenta de diagramação completa dentro da interface do Antigravity e, com um clique de botão, enviar imagens anotadas diretamente para a sessão do meu agente de programação, que reage a elas imediatamente.

Isso é apenas a ponta do iceberg. A capacidade de anotar o seu trabalho e explicar problemas visualmente, com imagens, logs e todo tipo de dado de apoio, ali mesmo sem sair do Antigravity, foi o que me convenceu de que finalmente estamos "vivendo no futuro".

Para provar isso para você, construí dois plugins open-source com UI Extensions que você pode testar agora mesmo: o [**Antigravity Paint Studio (`agy-paint-studio`)**](https://github.com/danicat/agy-paint-studio) para desenho em múltiplas camadas, edição e refinamento de imagens com IA, e o [**Antigravity Game Studio (`agy-game-studio`)**](https://github.com/danicat/agy-game-studio) para criar e depurar visualmente jogos 2D em Go.

Vamos ver como eles funcionam por baixo dos panos.

## Anatomia de uma UI extension

Uma UI extension é um plugin contendo um [**sidecar**](https://antigravity.google/docs/sidecars/) que expõe painéis web interativos e pode ser renderizado dentro da interface do Antigravity. Assim como plugins normais, ela também pode empacotar skills, MCPs e regras, permitindo construir experiências completas em torno de um determinado domínio (como edição de imagens).

Os plugins podem ser instalados tanto no nível do workspace (em `.agents/plugins`) quanto no nível global (em `~/.gemini/config/plugins`). De forma semelhante às skills, eles precisam seguir uma estrutura de diretórios específica.

```text
plugins/<plugin-name>/
├── plugin.json             # Arquivo obrigatório de manifesto e identificação
├── mcp_config.json         # Definições opcionais de servidores MCP
├── hooks.json              # Definição opcional de hooks
├── skills/                 # Diretório opcional de skills
│   └── <skill-name>/
│       └── SKILL.md
├── agents/                 # Templates opcionais de definição de subagentes
│   └── <agent-name>.md
├── rules/                  # Diretório opcional de regras
│   └── <rule-name>.md
└── sidecars/
    └── <sidecar-name>/
        └── sidecar.json    # Obrigatório para UI extensions
```

Embora sidecars comuns possam ser escritos em qualquer linguagem, para as UI extensions vamos utilizar [Node.js](https://nodejs.org/), pois o Antigravity já vem com um **Sidecar SDK** embutido que simplifica a comunicação entre o processo principal do agente e o processo do sidecar. Coisas como descoberta de portas e troca de mensagens são resolvidas pelo SDK como num passe de mágica.

Como nota à parte: na verdade, é possível alcançar a mesma experiência sem usar Node.js, o Sidecar SDK ou qualquer suporte nativo a UI Extensions, como mostrei na minha [palestra State of Agentic Coding no Devoxx Belgium](https://www.youtube.com/watch?v=nHRSwiFNQ9E). O problema da abordagem manual é que você acaba tendo que lidar com alocação de portas (e possíveis conflitos), além de descoberta e comunicação de sessão do jeito difícil. O Sidecar SDK torna tudo isso muito mais simples.

![Arquitetura de uma UI extension e sidecar no Antigravity](architecture-diagram.png "Como o Antigravity supervisiona o processo sidecar de um plugin e conecta o painel lateral auxiliar à conversa ativa")

O ponto de entrada é `sidecars/<sidecar-name>/sidecar.json`, que informa ao Antigravity como iniciar o processo em segundo plano e renderizar sua interface web no painel lateral:

```json
{
  "command": "node",
  "args": ["main.mjs"],
  "restart_policy": "always",
  "display_name": "My Extension",
  "has_web_ui": true,
  "ui_config": {
    "views": [
      {
        "path": "/",
        "entrypoint": "SIDECAR_UI_ENTRYPOINT_AUX_PANE",
        "title": "My Extension"
      }
    ]
  }
}
```

Aqui vale destacar o atributo `has_web_ui` e as propriedades de `ui_config`, pois são eles que transformam um sidecar comum em uma UI extension.

Como `"command": "node"` utiliza o runtime Node.js embutido do Antigravity e o `sidecar_sdk`, sua extensão não precisa de `npm install` nem de `node_modules`. Um backend completo em `main.mjs` leva apenas algumas linhas de JavaScript:

```javascript
import { readFileSync } from 'node:fs';
import { SidecarApp } from 'sidecar_sdk';

const app = new SidecarApp();

// Re-reading on each request lets edits appear when you reopen.
app.page('/', () =>
  readFileSync(new URL('./index.html', import.meta.url), 'utf8')
);

// POST routes return JSON and verify the sidecar token automatically.
app.api('/api/save', (data) => ({ saved: true, name: data.name }));

app.run();
```

No frontend, adicionar `<script src="/preload.js"></script>` ao `<head>` do seu HTML sincroniza as variáveis de tema claro e escuro do Antigravity e expõe o objeto `window.sidecar`, permitindo que sua interface chame o backend (`fetch()`), inspecione o workspace aberto (`getWorkspaceUris()`) e envie mensagens direto para o turno ativo do chat (`agent.sendMessage()`). Uma vez ativada, você pode abrir o painel pelo menu **⋮** em **Extensions**, clicando em qualquer arquivo correspondente a `file_extensions`, ou por meio de um atalho clicável `sidecar://` no chat.

## Criando suas próprias UI extensions do jeito fácil

Esses detalhes de implementação são ótimos quando você precisa resolver algo no braço, mas, sinceramente, espero que nenhum de vocês escreva essas extensões manualmente. Isso porque o Antigravity é "autoconsciente": ele sabe como construir essas extensões e você só precisa pedir para ele fazer isso por você. Se quiser ser bem explícito, pode pedir para ele invocar as skills `/plugin` e `/ui-extension`. Se quiser "ensinar" a ele sobre personalizações em geral, você também pode usar a skill `/agy-customizations`.

Foi assim que construí tanto o "Antigravity Paint Studio" quanto o "Antigravity Game Studio". Vamos dar uma olhada nas duas extensões agora.

## Editando imagens sem sair do Antigravity com o Paint Studio

O **Paint Studio** nasceu exatamente daquele caso de uso que mencionei na introdução: eu queria poder descrever meus problemas usando imagens ("um desenho vale mais que mil palavras"), mas sem trocar de contexto. Isso significava ter um lugar dentro do Antigravity onde eu pudesse rabiscar uma ideia rápida do zero, ou clicar em qualquer imagem do meu workspace ou sessão de chat para abri-la no painel lateral, anotar algumas observações por cima e enviá-la de volta para o agente.

![Paint Studio acoplado ao painel lateral auxiliar refinando o rascunho de um gato em uma ilustração em aquarela](paint-studio-sidepane.png "Paint Studio acoplado ao painel lateral auxiliar, transformando o rabisco de 10 segundos de um gato à esquerda em uma ilustração em aquarela à direita")

O que torna essa configuração muito divertida é que você acaba tendo **IA dos dois lados** do fluxo de trabalho. Você tem o seu agente de programação no painel de chat à esquerda e também tem o Gemini (neste caso, o Nano Banana 2 Lite, também conhecido como `gemini-3.1-flash-lite-image`) integrado diretamente ao Paint Studio à direita. Na captura de tela acima, desenhei aquele gatinho "smol void" em cerca de dez segundos; depois, após algumas etapas de refinamento, o Nano Banana 2 Lite o transformou na pintura em aquarela que você vê no painel lateral.

Vale mencionar que o Paint Studio também aproveita outro recurso de interface bem interessante do Antigravity: as interfaces generativas (aliás, [o meu primeiro protótipo do Paint Studio](https://x.com/danicat83/status/2100298197931372652) foi feito puramente com `/generative_ui`). As Generative UIs diferem das UI Extensions porque são renderizadas inline dentro da própria janela de chat. Elas são ativadas pela skill `/generative_ui` e você pode usá-las tanto para criar superfícies reutilizáveis (como as do Paint Studio) quanto descartáveis, por exemplo, para uma visualização pontual de dados.

{{< x user="danicat83" id="2100298197931372652" >}}

A Generative UI funciona emitindo um bloco `<agent-embed>` que renderiza um iframe em sandbox com limite de 500px de altura. Você pode colocar **quase** qualquer coisa dentro desse iframe usando JavaScript inline e as classes semânticas do Tailwind fornecidas pelo host (como `bg-[var(--card)]` e `text-[var(--foreground)]`). No entanto, sua Content Security Policy (CSP) bloqueia todas as CDNs externas, exceto pelo script do Tailwind embutido no Antigravity (`https://www.gstatic.com/antigravity/web/dev/tailwindcss.min.js`), e não permite a execução de WebAssembly (WASM).

Para o Paint Studio, tenho um endpoint específico capaz de renderizar tanto a interface completa (para o painel lateral ou modo destacado) quanto uma versão simplificada (para o painel inline via `/generative_ui`). Incluí no plugin a skill `/paint-studio`, que você pode acionar para instruir o agente a abrir o estúdio, garantindo um atalho prático sempre que precisar de um diagrama rápido.

![Canvas inline do Paint Studio renderizado via /paint-studio e /generative_ui mostrando o rascunho do gato smol void](paint-studio-inline.png "Invocar a skill /paint-studio carrega um painel inline via /generative_ui para rascunhos rápidos")

Para ainda mais praticidade, você pode registrar tipos de arquivo que abrem com a sua extensão no `sidecar.json` (como `*.psd`, `*.png` e outros), e o aplicativo também pode ser iniciado com uma tela em branco pelo menu de extensões:

```json
{
  "ui_config": {
    "display_name": "Paint Studio",
    "views": [
      {
        "path": "/",
        "entrypoint": "SIDECAR_UI_ENTRYPOINT_AUX_PANE",
        "title": "Paint Studio",
        "file_extensions": [".psd", ".png", ".jpg", ".jpeg", ".webp"]
      }
    ]
  }
}
```

Dessa forma, garantimos uma comunicação bidirecional completa: podemos clicar em qualquer arquivo para abri-lo no estúdio, fazer anotações, retoques e melhorias com IA, e enviar para o chat com o clique de um botão.

## Testando e depurando jogos 2D com o Antigravity Game Studio

Desde que criei minhas skills de desenvolvimento de jogos, venho pensando em como levar isso para o próximo nível. Skills, que são basicamente prompts, só conseguem te levar até certo ponto. No desenvolvimento de jogos, existe um componente visual enorme, que vai desde a direção de arte, criação de assets e design de fases até o HUD e tantos outros detalhes.

Depois de explorar a experiência redesenhada com o Paint Studio, comecei a pensar em como poderíamos redesenhar a experiência de desenvolvimento para criadores de jogos também, e foi assim que surgiu o **Antigravity Game Studio**. No começo, tentei empacotar todo tipo de ferramenta nele, mas acabou virando uma grande bagunça. Então decidi focar na funcionalidade que mais se beneficiaria de feedback visual do que qualquer outra: o playtesting.

A UI extension — **Playtest Studio** — faz uma única coisa: ela compila o seu jogo em Go para WebAssembly e permite jogar e depurá-lo ao vivo a 60 FPS ali mesmo no painel lateral. Enquanto você joga, o Playtest Studio grava um **buffer de replay contínuo de 10 segundos** (como um DVR).

![Playtest Studio pausado no modo de anotação DVR](playtest-studio-annotate.png "Pausando um jogo no Playtest Studio, voltando na linha do tempo de 10 segundos e desenhando uma anotação sobre um bug de spawn")

Agora, sempre que vejo algo estranho acontecer no jogo, o ciclo de depuração leva segundos:

1. Pressiono `F8` (ou `` ` ``) para **Pausar** e volto quadro a quadro na linha do tempo de 10 segundos até o momento exato em que o bug aconteceu.
2. Desenho uma caixa ou seta diretamente sobre o sprite com problema, ou escrevo uma nota rápida na margem ao redor da tela.
3. Clico em **📸 Frame** ou **🎞️ 10s GIF** para enviar a captura de tela anotada (ou um replay animado de 10 segundos em GIF), junto com os logs de console capturados do WebAssembly, direto para a sessão de chat para o agente corrigir.

O estúdio envia de volta ao agente uma mensagem estruturada contendo todas essas informações para que ele possa trabalhar na correção imediatamente, com muito menos ambiguidade do que um prompt em texto.

![Mensagem estruturada de relatório de bug enviada pelo Playtest Studio na sessão de chat do Antigravity](playtest-bug-report.png "Captura de tela do momento em que o Antigravity recebe uma mensagem do Playtest Studio com um relatório de bug")

Você pode ver todo esse fluxo de trabalho em ação no vídeo abaixo, onde abro o Playtest Studio para dar minha primeira olhada em um jogo 2D de Tower Defense que o agente tinha acabado de construir, envio feedback visual e um GIF de 10 segundos para o chat e rebalanceio a economia do jogo na hora:

{{< youtube "A2zOv5Xoh0M" >}}

## Considerações finais

Poder criar ferramentas visuais personalizadas que vivem bem ao lado do seu agente — e conversam com ele com um único clique — muda completamente o que um ambiente de programação pode ser. Em vez de ficar alternando de contexto para aplicativos externos e copiando e colando capturas de tela nos prompts, você pode moldar a interface do Antigravity em torno de qualquer problema que esteja resolvendo hoje.

Se você quiser testar esses plugins ou usá-los como inspiração para pedir ao Antigravity que construa suas próprias UI extensions, ambos os repositórios estão disponíveis no GitHub: [**`danicat/agy-paint-studio`**](https://github.com/danicat/agy-paint-studio) e [**`danicat/agy-game-studio`**](https://github.com/danicat/agy-game-studio).

Bons rascunhos!
