---
categories:
- Agent Development
date: '2025-10-31T11:43:35Z'
series:
- Building the Diagnostic Agent
series_order: 5
summary: Evolua seu agente Google ADK com uma interface personalizada em estilo retrô. Este guia passo a passo mostra como substituir a Dev-UI padrão por FastAPI e JavaScript vanilla, integrando um avatar animado com streaming em tempo real.
tags:
  - adk
  - fastapi
  - frontend
  - gemini
  - python
  - tutorial
title: 'Além da Dev-UI: Como Construir uma Interface para um Agente ADK'
---
Nos últimos seis meses, estive explorando IA Generativa, vibe coding, agentes e afins em minhas atividades de DevRel no Google. Sempre que desejo aprender uma nova tecnologia, acredito que a melhor maneira é construir algo prático com ela. Um dos meus projetos favoritos durante esse período foi o agente de diagnóstico: um software capaz de ajudar desenvolvedores a diagnosticar problemas no sistema operacional usando linguagem natural.

[Na Parte 4 desta série]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}), refatoramos o agente para utilizar o framework [ADK (Agent Development Kit)](https://google.github.io/adk-docs/). Neste artigo, vamos explorar como construir uma interface frontend personalizada para um agente ADK, adicionando mais personalidade e estilo ao projeto.

## Em busca de uma nova interface

Até o momento, estávamos conectando as peças centrais — ADK, [Gemini](https://gemini.google.com/), [Osquery](https://osquery.io/) e [Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform) — em um MVP funcional. O agente funcionava bem, mas eu precisava de uma interface mais interessante para as minhas palestras técnicas.

Foi então que me lembrei de um tweet de [Sito-san](https://x.com/Sikino_Sito) publicado alguns meses antes:

![Avatar UI de Sito-san](image.png "https://x.com/Sikino_Sito/status/1957645002533925235")

Como grande fã de animes e jogos retrô, fiquei encantada com a estética. Durante a preparação para uma palestra no BiznagaFest em Málaga, decidi que era o momento perfeito para ir além da interface padrão de debug do ADK e criar um cliente completo com um avatar interativo — no estilo 16-bit.

Sito-san mantém o projeto open source [Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core). Usando esse conceito como inspiração, decidi criar uma interface web personalizada e integrada ao runtime do ADK.

## Explorando o Runtime do ADK

O primeiro passo foi construir o runtime do agente. O runtime é o componente encarregado de executar o agente, rotear as mensagens dos usuários e capturar os eventos gerados pelo modelo.

Até então, usávamos a Dev-UI do ADK (`adk web`):

![Dev UI do ADK](image-1.png)

A Dev-UI é excelente para depuração, inspeção de chamadas de ferramentas e testes rápidos. No entanto, para criar um produto com design próprio, precisamos instanciar o runner do ADK no nosso backend.

A arquitetura geral do sistema é:

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

Criamos um frontend em HTML/CSS e JavaScript vanilla que se comunica com um backend Python usando [FastAPI](https://fastapi.tiangolo.com/). O backend inicializa o `Runner` do ADK, que gerencia o ciclo de vida e a execução do agente.

Precisamos da classe `Runner` e de um serviço de sessão. Como a aplicação roda localmente para um único usuário, utilizamos o `InMemorySessionService`:

```python
from fastapi import FastAPI
from google.adk.runners import Runner
from google.adk.sessions import InMemorySessionService
from dotenv import load_dotenv

load_dotenv()

from aida.agent import root_agent

APP_NAME = "aida"

session_service = InMemorySessionService()
runner = Runner(
    app_name=APP_NAME,
    agent=root_agent,
    session_service=session_service
)
app = FastAPI()
```

Em seguida, implementamos o endpoint `/chat`:

```python
from fastapi import Request
from fastapi.responses import JSONResponse
from google.genai.types import Content, Part

@app.post("/chat")
async def chat_endpoint(request: Request):
    """Gerencia a conversa com o agente e retorna a resposta final."""
    body = await request.json()
    query = body.get("query")
    user_id = "demo_user"
    session_id = "demo_session"

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

O método `runner.run_async` transmite os eventos emitidos pelo agente. Filtramos a resposta final e a retornamos em formato JSON.

Podemos testar o servidor com `uvicorn main:app`:

```shell
$ curl -X POST localhost:8000/chat -d '{"query":"olá"}'
{"response":"Olá! Por favor, informe a natureza da sua emergência de diagnóstico."}
```

## Construindo a Interface Frontend

Para servir a interface web diretamente pelo FastAPI, definimos uma rota HTML básica:

```python
from fastapi.responses import HTMLResponse

@app.get("/", response_class=HTMLResponse)
async def get_chat_ui():
    return HTML_CONTENT
```

O HTML inicial contém uma janela de chat, um campo de entrada de texto e um botão de envio:

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

Essa estrutura simples já permite enviar e receber mensagens no navegador:

![Interface básica de chat](image-2.png)

## Criando a Estética Retrô-Cyberpunk

Para transformar a interface básica em um terminal retrô estilizado, utilizei o Gemini CLI. Forneci o screenshot de referência com o prompt:

> Gostaria de atualizar a UI em @demo.py para uma estética que se assemelhe a esta interface @image.png

![Prompt no Gemini CLI para estilização da UI](image-4.png)

O modelo gerou a paleta de cores e o layout com contêineres e tipografia adequados:

![Nova UI gerada pelo Gemini CLI com estética retrô](image-3.png)

### Gerando o Avatar com a Extensão Nano Banana

Para criar a personagem do avatar, usei a extensão [Nano Banana](https://github.com/gemini-cli-extensions/nanobanana) para o Gemini CLI, que permite gerar e editar imagens diretamente pelo terminal.

Com o prompt:
> crie um avatar para o agente em @demo.py. O avatar deve ser uma garota de anime 2D no estilo PC-98, olhando para a câmera em pose de repouso (idle).

Obtive o frame inicial:

![Avatar 2D em estilo PC-98](image-5.png)

Recortei o foco no rosto:

![Avatar AIDA recortado no rosto](image-6.png)

E gerei o segundo frame com a boca aberta para criar a animação de fala:

> modifique static/assets/aida.png para criar um novo asset com a mesma pose, mas com a boca aberta, como se estivesse falando.

![Segundo frame com a boca aberta](image-7.png)

Servimos os assets estáticos no FastAPI:

```python
from fastapi.responses import FileResponse

@app.get("/idle")
async def idle():
    return FileResponse("static/assets/idle.png")

@app.get("/talk")
async def talk():
    return FileResponse("static/assets/talk.png")
```

E no frontend, alternamos entre as duas imagens quando o agente estiver respondendo:

```javascript
let talkInterval = null;

function setAvatarState(state) {
    const avatarImg = document.getElementById('avatar-image');
    if (state === 'talking') {
        if (!talkInterval) {
            talkInterval = setInterval(() => {
                const isTalking = avatarImg.src.endsWith('/talk');
                avatarImg.src = isTalking ? '/idle' : '/talk';
            }, 150);
        }
    } else {
        if (talkInterval) {
            clearInterval(talkInterval);
            talkInterval = null;
        }
        avatarImg.src = '/idle';
    }
}
```

## Sincronizando o Avatar com Streaming em Tempo Real

Em vez de esperar o modelo gerar todo o parágrafo antes de exibir o texto, atualizamos o backend para transmitir os tokens em tempo real usando `StreamingResponse`:

```python
from fastapi.responses import StreamingResponse
from google.genai.types import Content, Part

@app.post("/chat")
async def chat_endpoint(request: Request):
    data = await request.json()
    user_query = data.get("query")
    
    user_id = "demo_user"
    session_id = "demo_session"

    if not await session_service.get_session(APP_NAME, user_id, session_id):
        await session_service.create_session(APP_NAME, user_id, session_id)

    async def response_stream():
        """Transmite os pedaços de texto gerados pelos eventos do agente."""
        async for event in runner.run_async(
            user_id=user_id,
            session_id=session_id,
            new_message=Content(role="user", parts=[Part.from_text(text=user_query)]),
        ):
            if event.is_final_response() and event.content and event.content.parts:
                for part in event.content.parts:
                    if hasattr(part, "text") and part.text:
                        yield part.text

    return StreamingResponse(response_stream(), media_type="text/plain")
```

No frontend, consumimos a `ReadableStream` e ativamos a animação enquanto o fluxo de dados estiver ativo:

```javascript
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
        
        for (const char of chunk) {
            aidaMsg.textContent += char;
            chatWindow.scrollTop = chatWindow.scrollHeight;
            await new Promise(r => setTimeout(r, 5));
        }
    }
} catch (err) {
    appendMessage(`SYSTEM> Erro: ${err.message}`, 'system');
} finally {
    setAvatarState('idle');
}
```

## Resultado Final

Com o streaming ativado, o avatar move a boca em perfeita sincronia com a digitação das respostas no terminal retrô:

![AIDA em ação](aida-demo.gif "AIDA funcionando!")

## Conclusões e Código-Fonte

Substituir a interface padrão por um runtime próprio com FastAPI permitiu dar uma identidade visual única ao agente de diagnóstico, transformando o assistente técnico em uma experiência imersiva e responsiva.

O código-fonte completo da demonstração está disponível para download:
*   **[Baixar demo.py](demo.py)**
*   **Repositório Completo no GitHub**: **[github.com/danicat/aida](https://github.com/danicat/aida)**

Na próxima e última parte desta série, [Como Construir um Agente Offline com ADK, Ollama e SQLite]({{< ref "/posts/20251103-building-aida-part-2" >}}), vamos tornar a AIDA completamente independente de internet, rodando modelos locais com Ollama e implementando busca semântica local com SQLite.

## Recursos

*   **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)**
*   **[Gemini](https://gemini.google.com/)**
*   **[Osquery](https://osquery.io/)**
*   **[Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform)**
*   **[FastAPI](https://fastapi.tiangolo.com/)**
*   **[Tweet de Sito-san](https://x.com/Sikino_Sito/status/1957645002533925235)**
*   **[Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core)**
*   **[Extensão Nano Banana](https://github.com/gemini-cli-extensions/nanobanana)**
