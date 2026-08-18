---
categories:
- Agent Development
date: '2025-10-31T11:43:35Z'
series:
- Building the Diagnostic Agent
series_order: 5
summary: Google ADK エージェントにレトロスタイルのカスタムインターフェースを構築します。FastAPI と JavaScript を用いて標準の Dev-UI を置き換え、リアルタイムストリーミングと連動して動く AI 生成アバターを実装する手順を解説します。
tags:
  - adk
  - fastapi
  - frontend
  - gemini
  - python
  - tutorial
title: Dev-UIの先へ：ADKエージェントのインターフェースを構築する方法
---
過去半年間、Google での DevRel 活動の一環として、生成 AI、バイブコーディング、エージェント開発を探求してきました。新しい技術を学ぶ際、実際に手を動かして動くものを作るのが一番の近道だと考えています。その中で情熱を注いできたプロジェクトが「診断エージェント」です。自然言語でコンピューターの問題を診断できるアシスタントを目指しています。

[シリーズ第4弾]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}) では、Google の [ADK (Agent Development Kit)](https://google.github.io/adk-docs/) を使ってエージェントをリファクタリングしました。今回は、標準の Dev-UI から一歩進んで、ADK エージェント向けのカスタムフロントエンドを作成し、プロジェクトに個性豊かな UI を与える方法を解説します。

## 新しいユーザーインターフェースを求めて

これまでは、ADK、[Gemini](https://gemini.google.com/)、[Osquery](https://osquery.io/)、[Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform) などの主要コンポーネントを組み合わせて MVP を作ってきました。

次のステップを考えていたとき、8月に見かけた [Sitoさん](https://x.com/Sikino_Sito) のツイートを思い出しました：

![Sitoさんの Avatar UI](image.png "https://x.com/Sikino_Sito/status/1957645002533925235")

アニメやレトロゲームが大好きな私にとって、このデザインは非常に魅力的でした。マラガで開催された BiznagaFest での登壇準備中、ADK の開発用 UI を超えて、独自のレトロスタイル（8-bit ではなく 16-bit くらいの懐かしさ）のクライアントを作ろうと思い立ちました。

Sitoさんは [Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core) というオープンソースプロジェクトを公開されています。そのアイデアに刺激を受け、ADK ランタイムと直接通信する専用の Web UI を構築することにしました。

## ADK ランタイムの仕組み

独自の UI を作成する第一歩は、エージェントランタイムを構築することです。ランタイムはエージェントを実行し、ユーザーからのリクエストをルーティングして、モデルが生成したイベントを処理するコンポーネントです。

これまでは `adk web` コマンドで起動する開発用 UI（Dev-UI）を利用していました：

![ADK Dev-UI](image-1.png)

Dev-UI はデバッグやリクエスト/レスポンスの確認には非常に便利ですが、独自の画面を作るには、自分で ADK の Runner を制御する必要があります。

システムの全体アーキテクチャは以下の通りです：

{{< mermaid >}}
flowchart LR
    frontend["`Frontend
    (HTML/CSS + JS)`"]
    runtime[Runtime]
    root["Root Agent"]
    osquery["osquery"]
    schema["schema"]
    rag[("Schema RAG")]
    os("オペレーティングシステム")
    
    frontend -->|GET/POST| runtime
    
    subgraph be ["`Backend
    (FastAPI)`"]
    runtime -- クエリ --> root
    root -- イベント --> runtime
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

HTML/CSS と Vanilla JavaScript で書かれた軽量なフロントエンドが、[FastAPI](https://fastapi.tiangolo.com/) で構築された Python バックエンドにリクエストを送信します。バックエンドでは ADK の Runner を立ち上げ、ルートエージェントとの対話を制御します。

Runner クラスとセッションサービスをセットアップします。今回は単一ユーザーかつ一時的なセッションのため、`InMemorySessionService` を使用します：

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

次に、メッセージを送信するための `/chat` エンドポイントを実装します：

```python
from fastapi import Request
from fastapi.responses import JSONResponse
from google.genai.types import Content, Part

@app.post("/chat")
async def chat_endpoint(request: Request):
    """チャット処理を行い、エージェントの最終応答を返します。"""
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

`runner.run_async` を呼び出すことで、エージェントからイベントが非同期で送られてきます。最終応答（`is_final_response`）を取得して JSON で返します。

`uvicorn main:app` でサーバーを起動し、cURL でテストできます：

```shell
$ curl -X POST localhost:8000/chat -d '{"query":"こんにちは"}'
{"response":"こんにちは！診断が必要な緊急事態の内容を教えてください。"}
```

## フロントエンドの実装

FastAPI から HTML 画面を直接配信します：

```python
from fastapi.responses import HTMLResponse

@app.get("/", response_class=HTMLResponse)
async def get_chat_ui():
    return HTML_CONTENT
```

基本となる HTML 構造はシンプルです：

```html
<!DOCTYPE html>
<html lang="ja">
<head>
    <title>AIDA Chat</title>
</head>
<body>
    <h1>AIDA Chat</h1>
    <div id="chat-window"></div>
    <form id="input-area">
        <input type="text" id="user-input" placeholder="メッセージを入力..." autocomplete="off">
        <button type="submit">送信</button>
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
                appendMessage('エラーが発生しました。', 'error-message');
            }
        });
    </script>
</body>
</html>
```

ブラウザで開くと、シンプルな対話画面が表示されます：

![最小限のチャット画面](image-2.png)

## レトロ・サイバーパンク風デザインの適用

Gemini CLI を活用して、Sitoさんの画面イメージを参考にスタイリングを依頼しました：

> @demo.py の UI を、このスクリーンショット @image.png のようなレトロサイバーパンク風の雰囲気にアップデートしてください。

![Gemini CLI でのプロンプト入力画面](image-4.png)

Gemini CLI が生成したデザインがこちらです：

![Gemini CLI によって生成された新しいレトロ UI](image-3.png)

### Nano Banana 拡張機能によるアバター生成

アバター画像の作成には、Gemini CLI の [Nano Banana](https://github.com/gemini-cli-extensions/nanobanana) 拡張機能を使用しました。

> @demo.py のエージェント用アバターを作成してください。PC-98スタイルの2Dアニメキャラクターで、カメラを見つめる待機ポーズ（idle）にしてください。

生成されたベース画像：

![生成された2Dアニメアバター](image-5.png)

顔を中心にトリミングし：

![顔をトリミングしたアバター画像](image-6.png)

さらに口を開けた会話時用（talk）の差分フレームを生成させました：

> static/assets/aida.png を元に、同じポーズで口を開けて話している差分フレームを作成してください。

![口を開けた差分画像](image-7.png)

FastAPI で画像を配信します：

```python
from fastapi.responses import FileResponse

@app.get("/idle")
async def idle():
    return FileResponse("static/assets/idle.png")

@app.get("/talk")
async def talk():
    return FileResponse("static/assets/talk.png")
```

フロントエンドでアニメーションを制御します：

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

## ストリーミングによるアバターのリアルタイム連動

テキストが生成されるのを待つのではなく、トークンが到着するたびに画面にタイプライター風に出力し、アバターの口パクと連動させます。

バックエンドを `StreamingResponse` に更新します：

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
        """エージェントイベントからテキストチャンクを逐次生成"""
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

フロントエンドで `ReadableStream` を読み取ります：

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
    appendMessage(`SYSTEM> エラー: ${err.message}`, 'system');
} finally {
    setAvatarState('idle');
}
```

## 完成したインターフェース

ストリーミングと連動して、AIDA がテキストを読み上げるように口を動かします：

![動作中の AIDA アバター](aida-demo.gif "AIDA がリアルタイムに対話")

## まとめとソースコード

標準の Dev-UI を脱却し、FastAPI とストリーミングを組み合わせることで、独自の個性を持った魅力的なエージェントインターフェースを構築できました。

デモのソースコードは以下から確認できます：
*   **[demo.py をダウンロード](demo.py)**
*   **GitHub リポジトリ**: **[github.com/danicat/aida](https://github.com/danicat/aida)**

シリーズ最終回となる第6弾 [ADK、Ollama、SQLite で完全オフラインなエージェントを構築する方法]({{< ref "/posts/20251103-building-aida-part-2" >}}) では、ローカル LLM（Qwen 2.5）と SQLite RAG を組み合わせて、ネットワークが切断された環境でも稼働する完全オフライン対応の AIDA を完成させます。

## 参考リンク

*   **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)**
*   **[Gemini](https://gemini.google.com/)**
*   **[Osquery](https://osquery.io/)**
*   **[Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform)**
*   **[FastAPI](https://fastapi.tiangolo.com/)**
*   **[Sitoさんのツイート](https://x.com/Sikino_Sito/status/1957645002533925235)**
*   **[Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core)**
*   **[Nano Banana 拡張機能](https://github.com/gemini-cli-extensions/nanobanana)**
