---categories:
- Agent Development
date: '2025-10-31T11:43:35Z'
series:
- Building the Diagnostic Agent
series_order: 5
summary: Evolua seu agente Google ADK com uma interface personalizada em estilo retrô. Este guia passo a passo mostra como substituir a Dev-UI padrão por FastAPI e JavaScript vanilla, integrando um avatar animado com streaming em tempo real.
tags:
  - adk
  - fastapi
  - gemini
  - javascript
  - python
  - tutorial
title: "Além da Dev-UI: Construindo uma Interface Customizada para o ADK"
slug: "building-aida"
aliases:
  - "/pt-br/posts/20251031-building-aida/"
description: "Aprenda a substituir a Dev-UI do Google ADK por um backend FastAPI customizado, cliente Vanilla JS e avatar animado sincronizado com streaming de áudio."
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.11+"
  - "Google ADK"
  - "FastAPI"
  - "Uvicorn"
  - "Gemini CLI"
---
Nos últimos seis meses, estive explorando IA Generativa, vibe coding, agentes e afins em minhas atividades de DevRel no Google. Sempre que desejo aprender uma nova tecnologia, acredito que a melhor maneira é construir algo prático com ela. Um dos meus projetos favoritos durante esse período foi o agente de diagnóstico: um software capaz de ajudar pessoas a diagnosticar problemas no computador usando linguagem natural.

[Na Parte 4 desta série]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}), refatoramos o agente para utilizar o framework [ADK (Agent Development Kit)](https://google.github.io/adk-docs/). Neste artigo, vamos explorar como construir uma interface frontend personalizada para um agente ADK, adicionando mais personalidade e estilo ao projeto.

## Em busca de uma nova interface

Até o momento, estive trabalhando com os componentes centrais — ADK, [Gemini](https://gemini.google.com/), [osquery](https://osquery.io/) e [Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform) — para montar um pequeno MVP. O agente tinha algumas peculiaridades, mas era interessante o suficiente para que eu o usasse como conteúdo para algumas das minhas palestras nos últimos meses.

Eu estava um pouco sem saber para onde levar o projeto a seguir, quando me lembrei deste tweet de [Sito-san](https://x.com/Sikino_Sito) que vi em agosto:

![Avatar UI de Sito-san](image.png "https://x.com/Sikino_Sito/status/1957645002533925235")

Como grande fã de animes e jogos retrô, fiquei muito empolgada com a estética, mas naquela época ainda não tinha conectado todos os pontos. Avançando alguns meses, durante a preparação para minha palestra no BiznagaFest em Málaga, decidi que era o momento de ir além da UI de desenvolvimento do ADK e criar um cliente de verdade para meu agente. Foi aí que as coisas finalmente se encaixaram.

O Sito-san tem um projeto open source chamado [Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core), mas eu não tinha conhecimento suficiente para conectá-lo de imediato. Como disse acima, acredito que a melhor forma de aprender é construindo, então decidi criar minha própria UI usando o trabalho dele como inspiração. Além disso, por mais que eu ame a estética retrô, também queria fazer algo um pouco mais moderno, mas sem exagerar... algo moderno como nos tempos dos 16-bits em vez de 8-bits.

## Explorando o runtime do ADK

O primeiro passo para construir minha própria UI foi criar o runtime do agente. O runtime é o componente que realmente executa o agente, responsável por rotear as requisições do usuário para o agente e capturar os eventos de volta para então processar as respostas dos modelos.

Eu não precisava fazer isso antes porque dependia da UI de desenvolvimento do ADK, que você pode iniciar com o comando `adk web`:

![Dev UI do ADK](image-1.png)

A UI de desenvolvimento é muito conveniente, pois oferece muitas ferramentas de depuração, permitindo inspecionar requisições e respostas dos modelos e criar conjuntos de avaliação, além de recursos multimodais prontos para uso para lidar com imagens e até streaming bidirecional.

A Dev-UI é em parte culpada por eu ter demorado tanto para explorar o runtime do ADK, já que, como tudo funcionava direto da caixa, eu estava felizmente focada em construir as capacidades e ferramentas do agente. Agora que preciso de uma UI adequada, tive que substituir a Dev-UI por algo personalizado, o que significava lidar com o executor do agente por conta própria.

A arquitetura geral da solução é assim:

{{< mermaid >}}
flowchart LR
    frontend["`Frontend
    (HTML/CSS + JS)`"]
    runtime[Runtime]
    root["Root Agent"]
    osquery["osquery"]
    schema["schema"]
    rag[("Schema RAG")]
    os("Sistema Operacional")
    
    frontend -->|GET/POST| runtime
    
    subgraph be ["`Backend
    (FastAPI)`"]
    runtime -- consulta --> root
    root -- eventos --> runtime
    root --> osquery
    root --> schema
    subgraph tools
    osquery
    schema
    end
    end
    
    osquery --> os
    schema --> rag
{{< /mermaid >}}

Temos um pequeno frontend escrito em HTML/CSS e JavaScript que fará requisições ao nosso backend escrito em Python usando o [FastAPI](https://fastapi.tiangolo.com/). O backend inicializará o runner do ADK, que controlará as interações com o nosso agente raiz (root agent).

O agente raiz é o "cérebro" da AIDA e é responsável por processar as requisições (roteando-as para um LLM) e fazer as chamadas de ferramentas necessárias. Sempre que o agente raiz termina, ele emite eventos que o runtime processará.

Vamos dar uma olhada em uma implementação básica primeiro. Por brevidade, vou omitir a definição do agente raiz, já que ela foi explicada no [artigo anterior]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}).

Vamos precisar da classe `Runner` e de um serviço de sessão para controlar as sessões dos usuários. Existem muitas implementações de serviços de sessão que você pode explorar, mas neste caso o agente foi feito para ser usado por um único usuário e as sessões são efêmeras, então usaremos o `InMemorySessionService` e fixaremos os IDs de usuário e sessão por simplicidade.

Você pode ver as declarações de sessão e runner abaixo:

```py
from fastapi import FastAPI
from google.adk.runners import Runner
from google.adk.sessions import InMemorySessionService
from dotenv import load_dotenv

load_dotenv()

# --- Definição do Agente ---
from aida.agent import root_agent

APP_NAME = "aida"

# --- Configuração dos Serviços e do Runner ---
session_service = InMemorySessionService()
runner = Runner(
    app_name=APP_NAME, agent=root_agent, session_service=session_service
)
app = FastAPI()
```

Em seguida, precisamos implementar um endpoint para enviar mensagens ao agente. Vamos chamá-lo de `chat`:

```py
from fastapi import Request
from fastapi.responses import JSONResponse
from google.genai.types import Content, Part

# --- Endpoint da API para a Lógica de Chat ---
@app.post("/chat")
async def chat_endpoint(request: Request):
    """Gerencia a lógica do chat, retornando a resposta final do agente."""
    body = await request.json()
    query = body.get("query")
    user_id = "demo_user"
    session_id = "demo_session"

    # Garante que a sessão exista
    session = await session_service.get_session(app_name=APP_NAME, user_id=user_id, session_id=session_id)
    if not session:
        session = await session_service.create_session(app_name=APP_NAME, user_id=user_id, session_id=session_id)

    response_text = ""
    async for event in runner.run_async(
        user_id=user_id,
        session_id=session_id,
        new_message=Content(role="user", parts=[Part.from_text(text=query)]),
    ):
        if event.is_final_response() and event.content and event.content.parts[0].text:
            response_text = event.content.parts[0].text

    return JSONResponse(content={"response": response_text})
```

As primeiras linhas são o código típico de tratamento de requisições somado ao nosso controle de sessão simplificado. O coração deste código é a chamada `runner.run_async`, que emite os eventos do agente raiz. Estamos interessados apenas na resposta final e a retornamos como uma resposta JSON para quem fez a chamada.

Você pode testar esse pequeno app executando o seguinte comando:

```sh
$ uvicorn main:app
...
INFO:     Started server process [86669]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
```

E em um segundo terminal, faça a chamada POST para `/chat`:

```sh
$ curl -X POST localhost:8000/chat -d '{"query":"olá"}'
{"response":"Olá! Por favor, informe a natureza da sua emergência de diagnóstico."}
```

Ainda não parece uma grande melhoria na UI, mas estamos chegando lá. As coisas sempre ficam feias antes de ficarem bonitas! :)

## O frontend do agente

Existem muitas maneiras de renderizar a UI, e eu não sou uma desenvolvedora frontend particularmente talentosa, então acabei delegando toda a parte de design para o Gemini CLI — que acabou fazendo um trabalho muito melhor do que eu jamais conseguiria.

Aqui vou mostrar uma abordagem bem minimalista que você pode usar para renderizar a primeira UI como uma prova de conceito, mas depois recomendo fortemente que converse com alguém que realmente entenda de frontend, ou faça como eu e dê uma chance ao Gemini CLI.

Precisamos servir o HTML de alguma forma, e o jeito mais rápido e direto é criar um novo endpoint FastAPI para isso:

```py
from fastapi.responses import HTMLResponse

# veja o conteúdo completo abaixo
HTML_CONTENT = """
...
"""

# --- Endpoint da UI Web ---
@app.get("/", response_class=HTMLResponse)
async def get_chat_ui():
    return HTML_CONTENT
```

Claro, a parte importante aqui é o conteúdo HTML propriamente dito. Estamos definindo três elementos: uma janela de chat, uma caixa de entrada de texto e um botão para enviar a mensagem ao agente.

Para manter o trecho de código curto, retirei todas as informações de estilo do código abaixo. É simples, mas funcional:

```html
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <title>AIDA Chat</title>
</head>
<body>
    <h1>AIDA Chat</h1>
    <div id="chat-window"></div>
    <form id="input-area">
        <input type="text" id="user-input" placeholder="Digite sua mensagem..." autocomplete="off">
        <button type="submit">Enviar</button>
    </form>

    <script>
        const chatWindow = document.getElementById('chat-window');
        const inputForm = document.getElementById('input-area');
        const userInput = document.getElementById('user-input');

        function appendMessage(text, className) {
            const div = document.createElement('div');
            div.className = className;
            div.textContent = text;
            chatWindow.appendChild(div);
            return div;
        }

        inputForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const query = userInput.value.trim();
            if (!query) return;

            appendMessage(`USER: ${query}`, 'user-message');
            userInput.value = '';

            try {
                const response = await fetch('/chat', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ query })
                });
                const data = await response.json();
                appendMessage(`AIDA: ${data.response}`, 'bot-message');
            } catch (error) {
                appendMessage('Erro ao comunicar com o agente.', 'error-message');
            }
        });
    </script>
</body>
</html>
```

Se você executar `uvicorn main:app` novamente e acessar a página inicial, verá algo assim. Experimente enviar uma mensagem:

![Interface básica de chat](image-2.png "É assim que o mundo seria sem designers")

Observe que estou mantendo tudo em um único arquivo para simplificar, mas no mundo real é muito melhor ter uma pasta separada para arquivos HTML, CSS, JS e assets (geralmente chamada de `static`), já que ter a extensão correta também ajuda sua IDE a entender o código. O Gemini CLI não se importa com syntax highlighting, mas isso é muito útil quando você está revisando o código manualmente ou fazendo ajustes finos.

## Deixando a interface bonita

Não vou mentir, tudo daqui para frente é puro Gemini CLI fazendo sua mágica. Meu objetivo era ter uma interface retrô-cyberpunk-fofa-anime parecida com a que o Sita-san criou, mas com o meu toque pessoal.

Então usei um truque e pedi ao Gemini CLI para replicar o estilo na captura de tela que tirei do tweet do Sita-san. Salvei a captura de tela como `image.png` e dei o seguinte prompt ao CLI:

> I would like to update the UI in @demo.py to an aesthetic that resembles this interface @image.png

O caractere `@` no Gemini CLI indica que ele deve carregar o recurso (arquivo) apontado por ele.

![Captura de tela do Gemini CLI mostrando o prompt usado para gerar a nova UI.](image-4.png)

E foi isso que o Gemini criou:

![A nova UI para o agente AIDA, gerada pelo Gemini CLI, com estética retrô-cyberpunk, janela de chat e avatar.](image-3.png)

Note que isso só é possível porque o Gemini 2.5 é multimodal, conseguindo realmente "entender" a imagem. Costumo usar esse truque com frequência para descrever ao modelo o que quero fazer quando não consigo explicar direito em palavras. Uma imagem vale mais que mil palavras, não é?

### Gerando assets com o Nano Banana

A interface ficou melhor, mas falta uma peça-chave: o avatar. Para resolver esse problema, usei um segundo truque no CLI — instalei a extensão [Nano Banana](https://github.com/gemini-cli-extensions/nanobanana) para o Gemini CLI.

A extensão me permite gerar imagens sem precisar alternar para outra ferramenta. O Nano Banana não serve apenas para geração de imagens, mas também para edição, o que o torna uma ferramenta muito eficiente para criar animações — já que posso, a partir de uma imagem base, pedir modificações para gerar novos frames.

O prompt que usei foi:
> create an avatar for the agent in @demo.py. the avatar should be a 2d anime girl in PC-98 style. make so it is looking at the "camera" in an idle pose

Assumindo que você tenha a extensão Nano Banana instalada, o Gemini CLI a invocará para gerar a imagem. Se você não quiser instalar a extensão, também pode fazer o mesmo no app do Gemini ou na web.

O resultado inicial foi esta imagem:

![Avatar inicial de garota de anime 2D em estilo PC-98, gerado pela extensão Nano Banana.](image-5.png)

Que recortei manualmente para focar apenas no rosto:

![Versão recortada do avatar da AIDA, focando no rosto.](image-6.png)

Para criar uma animação simples de fala, pedi para ele gerar um segundo frame com base neste:

> modify static/assets/aida.png to create a new asset with the exact same pose but the character is with the mouth open, speaking

E este foi o resultado:

![Segundo frame do avatar da AIDA com a boca aberta, para a animação de fala.](image-7.png)

Para manter as coisas organizadas, criei uma pasta `static/assets` para armazenar todos os arquivos `.png`. Eu poderia tê-los embutido usando base64, assim como fizemos com o conteúdo HTML, mas ficaria grande demais (e bagunçado) para o meu script Python.

Agora precisamos do código para servir esses arquivos:

```py
# --- Assets estáticos ---
@app.get("/idle")
async def idle():
    return FileResponse("static/assets/idle.png")

@app.get("/talk")
async def talk():
    return FileResponse("static/assets/talk.png")
```

E agora precisamos editar o HTML para preencher o `avatar-container` com a imagem de um desses endpoints:

```html
<div class="avatar-container">
    <img id="avatar-image" src="/idle" alt="Avatar da AIDA">
    <div id="avatar-name">AIDA</div>
</div>
```

O resultado:

![A interface de chat da AIDA com o avatar em repouso (idle) exibido.](image-8.png)

Estamos começando a chegar a algum lugar!

### A animação

Adicionar animação não é particularmente difícil, mas depende de criar frames para todas as poses que você precisa. Já criamos `talk` e `idle`, então é possível gerar uma animação simples de fala alternando esses frames.

Podemos encapsular essa lógica em uma função de estado simples:

```js
let talkInterval = null;

function setAvatarState(state) {
    const avatarImg = document.getElementById('avatar-image');
    if (state === 'talking') {
        if (!talkInterval) {
            talkInterval = setInterval(() => {
                // Alterna entre os frames de fala e repouso
                const isTalking = avatarImg.src.endsWith('/talk');
                avatarImg.src = isTalking ? '/idle' : '/talk';
            }, 150);
        }
    } else {
        // Para a animação e reseta para idle
        if (talkInterval) {
            clearInterval(talkInterval);
            talkInterval = null;
        }
        avatarImg.src = '/idle';
    }
}
```

Alcançar o efeito de animação de fala exige chamar `setAvatarState('talking')` quando o agente começar a enviar dados e `setAvatarState('idle')` quando ele terminar.

## Dando vida ao avatar com streaming

A última peça do quebra-cabeça é fazer nosso avatar realmente ganhar vida, sincronizando sua animação de fala com as respostas transmitidas via streaming pelo agente. Isso exige modificar tanto o backend quanto o frontend para lidar com dados em tempo real.

### Alterações no backend: transmitindo a resposta do agente via streaming

Requisições HTTP convencionais esperam até que a resposta inteira esteja pronta antes de enviar qualquer coisa de volta. Para um agente LLM, isso significa ficar olhando para uma tela estática enquanto ele "pensa" e gera um parágrafo completo. Para fazer o avatar parecer vivo, precisamos quebrar esse silêncio.

Vamos atualizar nosso endpoint `/chat` do FastAPI para usar `StreamingResponse`, entregando pedaços de texto diretamente dos eventos de `runner.run_async` no instante em que forem gerados.

```python
from fastapi.responses import StreamingResponse
from google.genai.types import Content, Part

@app.post("/chat")
async def chat_endpoint(request: Request):
    data = await request.json()
    user_query = data.get("query")
    
    # Valores fixos por simplicidade na demonstração
    user_id = "demo_user"
    session_id = "demo_session"

    # Garante que a sessão exista
    if not await session_service.get_session(APP_NAME, user_id, session_id):
        await session_service.create_session(APP_NAME, user_id, session_id)

    async def response_stream():
        """Gera pedaços de texto a partir dos eventos do agente."""
        async for event in runner.run_async(
            user_id=user_id,
            session_id=session_id,
            new_message=Content(role="user", parts=[Part.from_text(text=user_query)]),
        ):
            # Queremos apenas a resposta final em texto para esta UI simples
            if event.is_final_response() and event.content and event.content.parts:
                for part in event.content.parts:
                    if hasattr(part, "text") and part.text:
                        yield part.text

    return StreamingResponse(response_stream(), media_type="text/plain")
```

### Alterações no frontend: consumindo a stream e animando

Com o backend agora em streaming, nosso JavaScript no frontend precisa ser atualizado para consumir essa stream e disparar a animação de fala do avatar. Vamos modificar o event listener de `submit` para usar uma `ReadableStream` e anexar o texto à medida que ele chega.

Também envolvemos toda a operação em um bloco `try/finally`. Isso garante que, mesmo se ocorrer um erro durante a requisição de rede ou no processamento da stream, `setAvatarState('idle')` seja sempre chamado, evitando que o avatar fique preso em um loop infinito de fala.

```js
// ... dentro do handler de submit ...
// Prepara o contêiner de mensagem da AIDA
const aidaMsg = appendMessage('AIDA> ', 'aida');

try {
    const response = await fetch('/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query })
    });

    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    while (true) {
        const { value, done } = await reader.read();
        if (done) break;

        setAvatarState('talking');
        const chunk = decoder.decode(value, { stream: true });
        
        // Efeito de digitação
        for (const char of chunk) {
            aidaMsg.textContent += char;
            chatWindow.scrollTop = chatWindow.scrollHeight;
            // Pequeno atraso para dar um toque retrô
            await new Promise(r => setTimeout(r, 5)); 
        }
    }
} catch (err) {
    appendMessage(`SYSTEM> Erro: ${err.message}`, 'system');
} finally {
    setAvatarState('idle');
}
```

## O resultado final

Com essas alterações, nosso agente AIDA agora tem uma interface totalmente interativa e visualmente envolvente. O avatar ganha vida, falando em sincronia com as respostas transmitidas via streaming, criando uma experiência muito mais imersiva.

![AIDA em ação](aida-demo.gif "Ela está viva!")

## Considerações finais e código-fonte

Percorremos um longo caminho desde a UI básica de desenvolvimento do ADK. Exploramos como construir um frontend personalizado, aproveitando ferramentas de IA generativa como o Gemini CLI e a extensão Nano Banana para criar uma estética única retrô-cyberpunk-fofa-anime.

Este artigo cobriu os fundamentos da construção de um frontend para um agente ADK, mas é apenas o ponto de partida. Você pode baixar o código-fonte completo da demonstração que construímos aqui:

*   **[Baixar demo.py](demo.py)**

Se você tiver interesse em uma versão mais avançada do agente, pode encontrá-la no meu GitHub: **[github.com/danicat/aida](https://github.com/danicat/aida)**

Encorajo você a explorar o repositório, tentar executá-lo você mesmo e talvez contribuir! É uma ótima maneira de ver como esses blocos de construção se juntam em uma aplicação do mundo real.

Na próxima parte desta série, [Como Construir um Agente Offline com ADK, Ollama e SQLite]({{< ref "/posts/20251103-building-aida-part-2" >}}), vamos mergulhar fundo em como tornar a AIDA completamente resiliente a quedas de rede usando o Qwen 2.5 local via Ollama e RAG local com SQLite.

## Recursos

*   **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)**: Documentação oficial do Google ADK.
*   **[Gemini](https://gemini.google.com/)**: O assistente de IA do Google.
*   **[osquery](https://osquery.io/)**: O site oficial do osquery.
*   **[Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform)**: A plataforma unificada de IA do Google Cloud.
*   **[FastAPI](https://fastapi.tiangolo.com/)**: O site oficial do framework web Python FastAPI.
*   **[Tweet de Sito-san](https://x.com/Sikino_Sito/status/1957645002533925235)**: O tweet original que inspirou o design da UI.
*   **[Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core)**: O projeto de código aberto de Sito-san.
*   **[Extensão Nano Banana](https://github.com/gemini-cli-extensions/nanobanana)**: A extensão do Gemini CLI para geração de imagens.
