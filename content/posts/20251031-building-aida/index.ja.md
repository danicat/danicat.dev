---
categories:
- Agent Development
date: '2025-10-31T11:43:35Z'
series:
- Building the Diagnostic Agent
series_order: 5
summary: Google ADK エージェントにレトロスタイルのカスタムインターフェースを構築する実践ガイド。FastAPI とバニラ JavaScript を用いて標準の Dev-UI を置き換え、Gemini CLI で生成したアバターをリアルタイムストリーミングに合わせてアニメーションさせる手法を解説します。
tags:
  - adk
  - fastapi
  - gemini
  - javascript
  - python
  - tutorial
title: "Dev-UI の先へ：ADK エージェントのカスタムインターフェースを構築する"
slug: "building-aida"
aliases:
  - "/ja/posts/20251031-building-aida/"
description: "Google ADK の Dev-UI を脱却し、FastAPI バックエンド、Vanilla JS クライアント、Gemini CLI で生成したリアルタイムストリーミング同期アバターを構築する方法を解説します。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.11+"
  - "Google ADK"
  - "FastAPI"
  - "Uvicorn"
  - "Gemini CLI"
---

この半年間、Google での DevRel 業務の一環として、GenAI、バイブコーディング、エージェント開発、そしてその間にあるあらゆる領域を探求してきました。新しい技術を学ぶときは、実際に何かを作ってみるのが一番の近道だといつも感じています。この期間に私が情熱を注いできたプロジェクトのひとつが「診断エージェント」でした。これは、自然言語を使ってコンピューターのトラブルを診断できるソフトウェアです。

本シリーズの[パート4]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}})では、Google の [ADK](https://google.github.io/adk-docs/) を使うように診断エージェントをリファクタリングしました。今回の記事では、プロジェクトにもう少し個性を吹き込むために、ADK エージェント向けのカスタムフロントエンドを作成する方法を探っていきます。

## 新しいユーザーインターフェースを求めて

これまでは、ADK、[Gemini](https://gemini.google.com/)、[osquery](https://osquery.io/)、[Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform) といった主要コンポーネントを組み合わせて、小さな MVP を作り上げてきました。エージェントにはいくつかクセもありましたが、十分に面白かったため、ここ数か月のいくつかの登壇トークのネタとしても活用していました。

次のステップをどうしようかと少し行き詰まっていたとき、8月に見かけた [Sito さん](https://x.com/Sikino_Sito) のこのツイートを思い出しました：

![Sito-san's Avatar UI](image.png "https://x.com/Sikino_Sito/status/1957645002533925235")

アニメやレトロゲームの大ファンである私にとって、このビジュアルはたまらなく魅力的でしたが、当時はまだ自分のプロジェクトと直接結びついてはいませんでした。それから数か月後、スペイン・マラガの BiznagaFest での登壇準備をしていたとき、ADK の開発用 UI を超えて、エージェントのための本格的なクライアントを作ってみたいと考えました。そのとき、ついに点と点がカチッと繋がったのです。

Sito さんは [Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core) というオープンソースプロジェクトを公開されていますが、当時の私にはすぐに統合できるほどの知識がありませんでした。前述の通り、学ぶための最善の方法は「自分で作ること」です。そこで、Sito さんの作品にインスパイアされながら独自の UI を自作することにしました。また、レトロな世界観は大好きですが、8-bit というよりは 16-bit 時代のような、少しだけモダンさを残したテイストにしたいと考えました。

## ADK ランタイムを探る

独自の UI を構築するための最初のステップは、エージェントのランタイムを作成することでした。ランタイムとは、エージェントを実際に実行し、ユーザーのリクエストをエージェントへルーティングするとともに、エージェントから返ってくるイベントをキャプチャしてモデルのレスポンスを処理するコンポーネントです。

これまでは `adk web` コマンドで起動できる ADK の開発用 UI に頼っていたため、ランタイムを自分で書く必要がありませんでした：

![ADK dev UI](image-1.png)

開発用 UI は非常に便利です。モデルへのリクエストやレスポンスのインスペクト、評価セットの作成といった多数のデバッグツールが揃っているほか、画像や双方向ストリーミングを扱うマルチモーダル機能も標準で備わっています。

ある意味、Dev-UI が最初から完璧に動いてくれたおかげで、私はエージェントの機能やツールの開発だけに専念でき、ランタイム自体の探索を後回しにしてしまっていたとも言えます。しかし、独自の本格的な UI を作るとなると、Dev-UI を自前の実装に置き換え、エージェントの Runner を自分で直接制御しなければなりません。

ソリューションの全体アーキテクチャは以下のようになります：

{{< mermaid >}}
flowchart LR
    frontend["`Frontend
    (HTML/CSS + JS)
    `"]
    runtime[Runtime]
    root["Root Agent"]
    osquery["osquery"]
    schema["schema"]
    rag[("Schema RAG")]
    os("Operating System")
    
    frontend -->|GET/POST| runtime
    
    subgraph be ["`Backend
    (FastAPI)`"]
    runtime -- query --> root
    root -- events --> runtime
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

HTML/CSS と JavaScript で書かれた軽量なフロントエンドが、Python の [FastAPI](https://fastapi.tiangolo.com/) で書かれたバックエンドに対してリクエストを送信します。バックエンド側では ADK の Runner を起動し、root agent とのやり取りを制御します。

root agent は AIDA の「頭脳」であり、リクエストの処理（LLM へのルーティング）と必要なツール呼び出しを担当します。root agent の処理が完了すると、ランタイムが処理するためのイベントが発行されます。

まずは最小限の実装（ベアボーン実装）から見ていきましょう。なお、root agent の定義自体は[前回の記事]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}})で解説しているため、簡潔にするためにここでは割愛します。

ユーザーセッションを管理するために、`Runner` クラスとセッションサービスが必要になります。セッションサービスにはさまざまな実装がありますが、今回は単一ユーザー向けでありセッションも一時的なものであるため、シンプルにするために `InMemorySessionService` を使用し、ユーザー ID とセッション ID をコード内にハードコードします。

セッションと Runner の宣言は以下の通りです：

```py
from fastapi import FastAPI
from google.adk.runners import Runner
from google.adk.sessions import InMemorySessionService
from dotenv import load_dotenv

load_dotenv()

# --- Agent Definition ---
from aida.agent import root_agent

APP_NAME="aida"

# --- Services and Runner Setup ---
session_service = InMemorySessionService()
runner = Runner(
    app_name=APP_NAME, agent=root_agent, session_service=session_service
)
app = FastAPI()
```

次に、エージェントにメッセージを送信するためのエンドポイントを実装します。名前は `chat` としましょう：

```py
from fastapi import Request
from fastapi.responses import JSONResponse
from google.genai.types import Content, Part

# --- API Endpoint for Chat Logic ---
@app.post("/chat")
async def chat_endpoint(request: Request):
    """Handles the chat logic, returning the agent's final response."""
    body = await request.json()
    query = body.get("query")
    user_id = "demo_user"
    session_id = "demo_session"

    # Ensure a session exists
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

冒頭の数行は典型的なリクエスト処理コードと、ハードコードされたセッション管理です。このコードの肝は `runner.run_async` の呼び出しで、root agent からイベントが発行されます。ここでは最終レスポンスのみに関心があるため、それを呼び出し元に JSON レスポンスとして返します。

以下のコマンドを実行して、この小さなアプリをテストできます：

```sh
$ uvicorn main:app
...
INFO:     Started server process [86669]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
```

そして別のターミナルから `/chat` に POST リクエストを送ってみます：

```sh
$ curl -X POST localhost:8000/chat -d '{"query":"hello"}'
{"response":"Hello! Please state the nature of the diagnostic emergency."}
```

今のところは UI が大きく進化したようには感じられないかもしれませんが、着実に前進しています。洗練されてきれいになる前には、まず無骨な姿を経験するものです！ :)

## エージェントのフロントエンド

UI をレンダリングする方法はたくさんありますが、私はフロントエンド開発が特別得意というわけではありません。そのため、デザイン周りはすべて Gemini CLI に任せることにしました。結果として、私が自分で作るよりもはるかに素晴らしいものに仕上げてくれました。

ここでは、PoC（概念実証）として最初の UI を表示するための非常にシンプルなアプローチを紹介します。ですがその後は、実際にフロントエンドに詳しい人に相談するか、私と同じように Gemini CLI を試してみることを強くおすすめします。

何らかの形で HTML を配信する必要があるため、手っ取り早い方法として FastAPI に新しいエンドポイントを作成します：

```py
from fastapi.responses import HTMLResponse

# see the full content below
HTML_CONTENT="""
...
"""

# --- Web UI Endpoint ---
@app.get("/", response_class=HTMLResponse)
async def get_chat_ui():
    return HTML_CONTENT
```

もちろん、ここで重要なのは実際の HTML コンテンツです。チャットウィンドウ、入力ボックス、エージェントにメッセージを送信するボタンの3つの要素を定義します。

スニペットを短く保つため、以下のコードからはスタイル情報をすべて省いています。無骨ですが機能的です：

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <title>AIDA Chat</title>
</head>
<body>
    <h1>AIDA Chat</h1>
    <div id="chat-window"></div>
    <form id="input-area">
        <input type="text" id="user-input" placeholder="Type your message..." autocomplete="off">
        <button type="submit">Send</button>
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
                appendMessage('Error communicating with the bot.', 'error-message');
            }
        });
    </script>
</body>
</html>
```

再び `uvicorn main:app` を実行してホームページにアクセスすると、以下のような画面が表示されます。メッセージを送信してみてください。

![Barebones chat page](image-2.png "This is how the world would look without designers")

シンプルにするためにすべてを1つのファイルにまとめていますが、実際には HTML、CSS、JS、アセットファイル用のフォルダ（通常は `static`）に分けるのが適切です。正しい拡張子を使うことで、IDE がコードを理解しやすくなります。Gemini CLI にとってシンタックスハイライトは関係ありませんが、コードを手動で確認したり微調整したりするときに役立ちます。

## デザインを洗練させる

正直なところ、ここからの作業は完全に Gemini CLI による魔法です。私の目標は、Sito さんが作ったようなレトロ・サイバーパンク・キュート・アニメなインターフェースを、自分好みのテイストで再現することでした。

そこで少しズルをして、Sito さんのツイートから保存したスクリーンショットのスタイルを再現するよう Gemini CLI に依頼しました。スクリーンショットを `image.png` として保存し、CLI に次のプロンプトを与えました：

> I would like to update the UI in @demo.py to an aesthetic that resembles this interface @image.png  
> （@demo.py の UI を、このインターフェース @image.png に似たテイストにアップデートしてください）

Gemini CLI では、`@` 文字をつけることで指定したリソース（ファイル）を読み込ませることができます。

![Screenshot of the Gemini CLI showing the prompt used to generate the new UI.](image-4.png)

そして、Gemini が作成してくれたのがこちらです：

![The new UI for the AIDA agent, generated by Gemini CLI, featuring a retro-cyberpunk aesthetic with a chat window and an avatar.](image-3.png)

これが可能なのは、Gemini 2.5 がマルチモーダルであり、画像を実際に「理解」できるからです。言葉でうまく説明できないとき、私はよくこの手法を使ってモデルにやりたいことを伝えています。百聞は一見にしかず、ですよね？

### Nano Banana によるアセット生成

インターフェースは良くなりましたが、決定的なパーツが欠けています。それがアバターです。この問題を解決するために、もうひとつの CLI テクニックを使いました — Gemini CLI 向けの [Nano Banana](https://github.com/gemini-cli-extensions/nanobanana) 拡張機能をインストールしたのです。

この拡張機能を使うと、別のツールに切り替えることなく画像を生成できます。Nano Banana は新規の画像生成だけでなく編集にも優れているため、ベース画像から変更を加えて新しいフレームを生成し、アニメーションを作成するツールとして非常に効果的です。

私が使ったプロンプトは次の通りです：
> create an avatar for the agent in @demo.py. the avatar should be a 2d anime girl in PC-98 style. make so it is looking at the "camera" in an idle pose  
> （@demo.py のエージェント用アバターを作成してください。PC-98 スタイルの 2D アニメ少女で、待機ポーズで「カメラ」を見ているようにしてください）

Nano Banana 拡張機能がインストールされていれば、Gemini CLI がそれを呼び出して画像を生成します。拡張機能をインストールしたくない場合は、Gemini アプリや Web 上で同じプロンプトを実行しても構いません。

最初に出てきた結果がこの画像でした：

![Initial 2D anime girl avatar in PC-98 style, generated by the Nano Banana extension.](image-5.png)

これに手動でトリミングを加え、顔だけにフォーカスさせました：

![Cropped version of the AIDA avatar, focusing on the face.](image-6.png)

シンプルな会話アニメーション（口パク）を作るために、この画像をベースにして2枚目のフレームを生成させました：

> modify static/assets/aida.png to create a new asset with the exact same pose but the character is with the mouth open, speaking  
> （static/assets/aida.png を修正し、まったく同じポーズで口を開けて話している新しいアセットを作成してください）

その結果がこちらです：

![Second frame of the AIDA avatar with an open mouth, for the talking animation.](image-7.png)

ファイルを整理するため、すべての `png` ファイルを保存する `static/assets` フォルダを作成しました。HTML コンテンツと同様に base64 でインライン化することもできましたが、可哀想な Python スクリプトが肥大化して（しかも散らかって）しまいます。

次に、これらのファイルを配信するコードが必要です：

```py
# --- Static assets ---
@app.get("/idle")
async def idle():
    return FileResponse("static/assets/idle.png")

@app.get("/talk")
async def talk():
    return FileResponse("static/assets/talk.png")
```

そして、これらのエンドポイントのいずれかからの画像で `avatar-container` を埋めるように HTML を編集します：

```html
<div class="avatar-container">
    <img id="avatar-image" src="/idle" alt="AIDA Avatar">
    <div id="avatar-name">AIDA</div>
</div>
```

その結果がこちら：

![The AIDA chat interface with the idle avatar displayed.](image-8.png)

形が見えてきました！

### アニメーションの実装

アニメーションの追加はそれほど難しくありませんが、必要なすべてのポーズのフレームを用意する必要があります。すでに `talk` と `idle` を作成してあるので、これらのフレームを交互に切り替えることでシンプルな発話アニメーションを生成できます。

このロジックはシンプルな状態関数にカプセル化できます：

```js
let talkInterval = null;

function setAvatarState(state) {
    const avatarImg = document.getElementById('avatar-image');
    if (state === 'talking') {
        if (!talkInterval) {
            talkInterval = setInterval(() => {
                // Toggle between talk and idle frames
                const isTalking = avatarImg.src.endsWith('/talk');
                avatarImg.src = isTalking ? '/idle' : '/talk';
            }, 150);
        }
    } else {
        // Stop animation and reset to idle
        if (talkInterval) {
            clearInterval(talkInterval);
            talkInterval = null;
        }
        avatarImg.src = '/idle';
    }
}
```

発話アニメーション効果を実現するには、エージェントがデータの送信を開始したときに `setAvatarState('talking')` を呼び出し、終了したときに `setAvatarState('idle')` を呼び出すだけです。

## ストリーミングでアバターに命を吹き込む

パズルの最後のピースは、アバターの会話アニメーションをエージェントのストリーミング応答と同期させ、本当に生きているかのように見せることです。そのためには、リアルタイムデータを処理できるようにバックエンドとフロントエンドの両方を修正する必要があります。

### バックエンドの変更：エージェントの応答をストリーミングする

標準の HTTP リクエストは、レスポンス全体が準備できるまで何も送り返さずに待ちます。LLM エージェントの場合、これはモデルが「思考」して段落全体を生成する間、静止した画面を見つめ続けることを意味します。アバターを生き生きと感じさせるには、その沈黙を打ち破る必要があります。

FastAPI の `/chat` エンドポイントを `StreamingResponse` を使用するように更新し、テキストチャンクが生成された瞬間に `runner.run_async` イベントから直接 yield します。

```python
from fastapi.responses import StreamingResponse
from google.genai.types import Content, Part

@app.post("/chat")
async def chat_endpoint(request: Request):
    data = await request.json()
    user_query = data.get("query")
    
    # Hardcoded for demo simplicity
    user_id = "demo_user"
    session_id = "demo_session"

    # Ensure session exists
    if not await session_service.get_session(APP_NAME, user_id, session_id):
        await session_service.create_session(APP_NAME, user_id, session_id)

    async def response_stream():
        """Generates text chunks from the agent's events."""
        async for event in runner.run_async(
            user_id=user_id,
            session_id=session_id,
            new_message=Content(role="user", parts=[Part.from_text(text=user_query)]),
        ):
            # We only want the final text response for this simple UI
            if event.is_final_response() and event.content and event.content.parts:
                for part in event.content.parts:
                    if hasattr(part, "text") and part.text:
                        yield part.text

    return StreamingResponse(response_stream(), media_type="text/plain")
```

### フロントエンドの変更：ストリームの受信とアニメーション連動

バックエンドがストリーミングを行うようになったので、フロントエンドの JavaScript もこのストリームを受信してアバターの会話アニメーションをトリガーするように更新する必要があります。`submit` イベントリスナーを変更して `ReadableStream` を使用し、テキストが到着するたびに追加していきます。

また、操作全体を `try/finally` ブロックで囲みます。これにより、ネットワークリクエストやストリーム処理中にエラーが発生した場合でも、常に `setAvatarState('idle')` が呼び出され、アバターが無限に話し続けるループに陥るのを防ぐことができます。

```js
// ... inside the submit handler ...
// Prepare AIDA's message container
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
        
        // Typing effect
        for (const char of chunk) {
            aidaMsg.textContent += char;
            chatWindow.scrollTop = chatWindow.scrollHeight;
            // Tiny delay for retro feel
            await new Promise(r => setTimeout(r, 5)); 
        }
    }
} catch (err) {
    appendMessage(`SYSTEM> Error: ${err.message}`, 'system');
} finally {
    setAvatarState('idle');
}
```

## 完成した結果

これらの変更により、私たちの AIDA エージェントは完全にインタラクティブで視覚的にも魅力的なインターフェースを手に入れました。アバターに命が吹き込まれ、ストリーミングされるレスポンスと同期して話すことで、はるかに没入感のある体験が生まれます。

![AIDA in action](aida-demo.gif "It's alive!")

## おわりに・ソースコード

基本的な ADK 開発用 UI から、ずいぶんと遠くまでやってきました。Gemini CLI や Nano Banana 拡張機能のような生成 AI ツールを活用して、ユニークなレトロ・サイバーパンク・キュート・アニメな世界観を作り出しながら、カスタムフロントエンドを構築する方法を探ってきました。

この記事では ADK エージェント用フロントエンドの構築の基本を取り上げましたが、これはほんの出発点にすぎません。今回構築したデモの完全なソースコードは以下からダウンロードできます：

*   **[demo.py をダウンロード](demo.py)**

エージェントのより発展的なバージョンに興味がある方は、私の GitHub でご覧いただけます：**[github.com/danicat/aida](https://github.com/danicat/aida)**

ぜひリポジトリを探索し、自分で動かしてみて、よければコントリビュートもしてみてください！これらのビルディングブロックが実際のアプリケーションでどのように組み合わさるかを理解する素晴らしい方法になるはずです。

本シリーズの次回「[ADK、Ollama、SQLite で完全オフラインなエージェントを構築する方法]({{< ref "/posts/20251103-building-aida-part-2" >}})」では、Ollama 経由のローカル Qwen 2.5 とローカル SQLite RAG を使用して、ネットワーク障害時でも完全に動作する AIDA の構築を掘り下げます。

## 参考資料

*   **[Agent Development Kit (ADK)](https://google.github.io/adk-docs/)**: Google ADK の公式ドキュメント。
*   **[Gemini](https://gemini.google.com/)**: Google の AI アシスタント。
*   **[osquery](https://osquery.io/)**: osquery の公式サイト。
*   **[Vertex AI](https://cloud.google.com/vertex-ai/docs/start/introduction-unified-platform)**: Google Cloud の統合 AI プラットフォーム。
*   **[FastAPI](https://fastapi.tiangolo.com/)**: Python Web フレームワーク FastAPI の公式サイト。
*   **[Sito さんのツイート](https://x.com/Sikino_Sito/status/1957645002533925235)**: UI デザインの着想の元となったツイート。
*   **[Avatar UI Core](https://github.com/sito-sikino/avatar-ui-core)**: Sito さんのオープンソースプロジェクト。
*   **[Nano Banana 拡張機能](https://github.com/gemini-cli-extensions/nanobanana)**: 画像生成用の Gemini CLI 拡張機能。
