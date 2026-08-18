---categories:
- Agent Development
date: '2025-06-05T00:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 2
summary: Vertex AI SDK for Python を使用して、クライアントコードと Gemini API 間の通信モデル、メッセージ構造、および低レイヤーでの手動 Function Calling の仕組みを深く掘り下げて解説します。
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: "Vertex AI SDK for Python を深く掘り下げる"
slug: "vertex-ai-sdk-python"
aliases:
  - "/ja/posts/20250605-vertex-ai-sdk-python/"
description: "Vertex AI SDK for Python におけるクライアントとGemini API間の通信モデルを解説。システム指示、Function Calling宣言、チャットセッションの構造を詳解。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-genai"
  - "Google Cloud Vertex AI"
---
## はじめに

この記事では、[Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) を使用したクライアントコードと Gemini API 間の通信モデルについて詳しく探ります。メッセージがどのように構造化されているのか、モデルがどのように質問のコンテキストを理解するのか、そして Function Calling（関数呼び出し）によってモデルの機能をどのように拡張するのか、といった概念を解説します。今回は Gemini を中心に取り上げますが、ここで紹介する概念は Gemma やその他の LLM にも同様に適用できます。

[本シリーズの第1弾「AIエージェントを使って自分のPCを「エンタープライズ号」にした話」]({{< ref "/posts/20250531-diagnostic-agent" >}}) では、ローカルマシンの診断に関する質問に応答する、シンプルでありながら驚くほど強力な AI エージェントの作成方法を解説しました。ごくわずかなコード行数（そして決して少なくはないコメント行）で、「マシンの CPU 使用率はどれくらいか」「マルウェアの兆候がないか確認して」といった問い合わせに応答するエージェントを実現できました。

これはもちろん、処理を大幅に簡略化してくれる Python SDK の恩恵によるものでした。例えば、関数をいつ呼び出すべきかをエージェントに自動判断させる [Automatic Function Calling（自動関数呼び出し）](https://ai.google.dev/gemini-api/docs/function-calling?example=weather#automatic_function_calling_python_only) 機能を活用しました。この機能のおかげで、通常の Python 関数を定義するだけで、SDK が関数のシグネチャと説明（docstring）を動的に自動解釈してくれたのです。ただ、この機能は残念ながら Python SDK 限定であるため、他のプログラミング言語で開発する場合はもう少し手順を踏む必要があります。

そこで今回の記事では少しアプローチを変え、Gemini API 自体の仕組みを解説します。これにより、Python だけでなく、利用可能な各種 SDK（JavaScript、Go、Java など）を使用する際にもスムーズに対応できるようになります。前回の記事と比較しやすいようコード例には引き続き Python を使用しますが、ここで取り上げる概念はすべての言語に共通するものです。

本記事で取り上げる主なトピックは以下の2つです：
*   クライアントとモデル間で対話（会話ターン）がどのように行われるか
*   手動で Function Calling（関数呼び出し）を実装する方法

なお、Python 開発者にとっても得るものは決して少なくありません。対話の内部フローを理解しておくことは、Live API などのより高度な SDK 機能を活用したり、LLM 全般を扱ったりする上で非常に重要な基礎となるからです。

## API の仕組みを理解する

エージェントは一般的に、クライアント・サーバー型アプリケーションと同様の仕組みで動作します。リクエストの準備と送信を担うクライアントコンポーネントと、モデルランタイムをホストしてリクエストを処理するサーバープロセスが存在します。

Vertex AI には主に 2 つの API グループがあります。クライアントがリクエストを送信してレスポンスを待つ標準的なリクエスト/レスポンス型のコンテンツ生成を行う REST API と、WebSocket を使用してリアルタイム情報を処理する新しい [Live API](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) です。Live API を適切に扱うにはもう少し準備が必要になるため、本記事ではまず REST API に焦点を当てます。

コンテンツ生成では通常、テキスト、画像、音声、動画のいずれかのモダリティを扱います。最新のモデルの多くはマルチモーダルであり、入力や出力において複数のモダリティを同時に処理できます。まずはシンプルに、テキストから始めましょう。

単発のリクエストを行う標準的なプロンプトアプリケーションは以下のようになります：

```python
from google import genai

client = genai.Client(
    vertexai=True,
    project="daniela-genai-sandbox",
    location="us-central1"
)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="How are you today?"
)
print(response.text)
```

出力：

```
I am doing well, thank you for asking! As a large language model, I don't experience emotions like humans do, but I am functioning optimally and ready to assist you. How can I help you today?
```

最初に行うべきことは、Vertex AI モード（`vertexai=True`）または Gemini API キーを使用してクライアントを初期化することです。今回の例では Vertex AI モードを使用しています。

クライアントが初期化されたら、`client.models.generate_content` メソッドを使ってプロンプトを送信できます。呼び出すモデル（ここでは `gemini-2.0-flash`）と、`contents` 引数にプロンプト（例: `"How are you today?"`）を指定します。

このコードを見ていると、Python が提供してくれる数多くの抽象化のおかげで、内部で何が起きているのかイメージしづらいかもしれません。ここで最も重要なポイントは、**コンテンツは単なる文字列ではない**ということです。

`contents` は実際にはコンテンツ構造体（`Content`）のリストであり、この構造体は **role（役割）** と 1 つ以上の **parts（パーツ）** で構成されています。この構造体の基底型は types ライブラリで定義されており、以下のようになっています：

```python
from google.genai import types

contents = [types.Content(
  role = "user",
  parts = [ types.Part.from_text("How are you today?") ]
)]
```

つまり、私たちが `contents="How are you today?"` と記述するたびに、Python SDK が自動的に文字列から「文字列パートを持つコンテンツ」への変換を行ってくれているのです。

もう 1 つ注意すべき重要な点は、`generate_content` を呼び出すたびに、モデルは事前の記憶を持たずにゼロから推論を開始するということです。つまり、過去のメッセージのコンテキストを次のプロンプトに含めるのは**私たち（クライアント側）の責務**となります。モデルに対して「今日は何日ですか？（what day is today?）」と 2 回連続で質問する簡単なテストを行ってみましょう：

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)
```

出力：

```
$ python3 main.py 
Today is Sunday, November 5th, 2023.

Today is Saturday, November 2nd, 2024.
```

上記の出力には 2 つの問題があります：1) モデルは実際の日付を知る術がないためハルシネーション（幻覚）を起こしている点、2) 同じ質問に対して 2 回異なる回答を返している点です。1) は日時の取得関数や Google 検索などのツールでグラウンディングすることで解決できますが、ここでは 2) に注目したいと思います。これはモデルが直前に自分が答えた内容すら覚えていないことを明確に示しており、モデルに会話の最新状態を伝え続けるのは**私たち**の責任であるという先ほどのポイントを実証しています。

コードに少し変更を加えてみましょう：

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)

# contents 配列の各要素は通常「ターン（turn）」と呼ばれます
contents = [
    {
        "role": "user",
        "parts": [{
            "text": "what day is today?"
        }]
    },
    {
        "role": "model",
        "parts": [{
            "text": response.text
        }]
    },
    {
        "role": "user",
        "parts": [{
            "text": "what day is today?"
        }]
    },
]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents
)
print(response.text)
```

出力：

```
$ python3 main.py 
Today is Wednesday, November 15, 2023.

Today is Wednesday, November 15, 2023.
```

モデルへの 2 回目の呼び出しでは、`contents` 属性に対話のコンテキスト全体を含めている点に注目してください。また、各パートの role が “user” から “model”、そして再び “user” へと変化していることにも注目してください（role に指定可能な値は “user” と “model” のみです）。モデルはこれによって会話のどの時点にいるのか、すなわち「ターン（turn）」を理解します。もし仮に、質問を繰り返す最後のパートを省略した場合、モデルは会話がすでに完了していると判断し、最後のターンが “user” ではなく “model” であるため新たなレスポンスを生成しません。

上記の `contents` 変数は「辞書（dictionary）」形式で記述されていますが、SDK には `types.UserContent`（role フィールドを自動的に “user” に設定）や `types.Part.from_text`（プレーンな文字列をパートに変換）など、便利なヘルパーメソッドも多数用意されています。

他の種類の入力や出力を扱う場合は、Function Calling（関数呼び出し）やバイナリデータなど、他の種類のパーツを使用できます。モデルがマルチモーダルの場合、同じメッセージ内に異なるコンテンツタイプのパーツを混在させることが可能です。

バイナリデータはインラインで渡すことも、URI から取得することもできます。`mime_type` フィールドを指定することで、異なるデータ形式を区別します。例えば、画像パーツは以下のように URI から取得できます：

```python
from google.genai import types

contents = types.Part.from_uri(
  file_uri='gs://generativeai-downloads/images/scones.jpg',
  mime_type='image/jpeg',
)
```

またはインラインで渡すことも可能です：

```python
contents = types.Part.from_bytes(
  data=my_cat_picture, # バイナリデータ
  mime_type='image/jpeg',
)
```

要約すると、会話のターンごとに、直前のモデルのレスポンスと新しいユーザーの質問の両方に対して、新しいコンテンツ行を追加していくことになります。

嬉しいことに、チャットボット体験は非常に重要なユースケースであるため、Vertex AI SDK にはこのフローの標準実装があらかじめ用意されています。`chat` 機能を使用すれば、上記の挙動をごくわずかなコード行数で再現できます：

```python
chat = client.chats.create(model='gemini-2.0-flash')
response = chat.send_message('what day is today?')
print(response.text)
response = chat.send_message('what day is today?')
print(response.text)
```

出力：

```
$ python3 main.py 
Today is Saturday, October 14th, 2023.

Today is Saturday, October 14th, 2023.
```

今回は、チャットインターフェースが会話履歴を自動的に管理してくれたため、モデルは日付を一貫して記憶できました。

## 手動での Function Calling（関数呼び出し）

クライアントメッセージの構築とコンテキスト管理の仕組みがわかったところで、次は API が Function Calling（関数呼び出し）をどのように処理しているかを見ていきましょう。基本的な流れとしては、利用可能な関数が存在することをモデルに伝え、モデルからの関数呼び出しリクエストを処理し、その実行結果をモデルに返却します。Function Calling はエージェントが外部システムや現実世界と対話し、単なるテキスト生成を超えてデータ取得や特定の処理のトリガーといったアクションを実行できるようにするために不可欠な機能です。

関数宣言（Function Declaration）は、モデルに対して何ができるかを伝えるものです。関数の名前、説明、そして引数をモデルに定義します。例えば、以下は `get_random_number` 関数の宣言です：

```python
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Returns a random number",
}
```

モデルはこの宣言を把握することで、どの関数を呼び出すべきかを判断します。関数宣言には `name`、`description`、`parameters` の 3 つのフィールドがあります（今回の関数は引数を取らないため `parameters` フィールドは省略しています）。モデルは関数の説明と引数の説明を参照して、各関数をいつどのように呼び出すべきかを判断します。

前回の記事では、関数宣言を明示的に記述するのを省き、関数の docstring に基づいて SDK に自動推論させていました。今回は内部フローをより深く理解するために、関数を明示的に宣言するアプローチをとります。

関数とその宣言コードは以下のようになります：

```python
def get_random_number():
    return 4 # 公平なサイコロの出目によって決定
             # 完全にランダムであることが保証されている (https://xkcd.com/221/)

# 宣言によって関数に関する必要な情報をモデルに伝える
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Returns a random number",
}
```

関数宣言のその他の例は、[Vertex AI の Function Calling ドキュメント](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling#schema-examples?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) で確認できます。

次に、モデルがこの関数を利用できることを伝える必要があります。これはモデル設定（`GenerateContentConfig`）を通じて行い、関数をツール（Tool）として追加します。

```python
tools = types.Tool(function_declarations=[get_random_number_decl])
config = types.GenerateContentConfig(tools=[tools])

# 最初のプロンプト
contents = [types.Part.from_text(text="what is my lucky number today?")]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config, # モデル呼び出しに config を渡している点に注目
)

print(response.candidates[0].content.parts[0])
```

上記のコードを実行すると、以下のような結果が出力されます：

```
$ python3 main.py 
video_metadata=None thought=None inline_data=None file_data=None thought_signature=None code_execution_result=None executable_code=None function_call=FunctionCall(id=None, args={}, name='get_random_number') function_response=None text=None
```

ここで出力されているのはモデルレスポンスの最初のパートです。`function_call` フィールド以外のすべてのフィールドが空（`None`）になっていることがわかります。これは、モデルが**私たち（クライアント）**に対してこの関数を呼び出し、その結果をモデルに返してくれるよう求めていることを意味します。

最初はこの挙動に少し戸惑うかもしれませんが、よく考えてみれば完全に理にかなっています。モデルは関数の存在を知ってはいますが、それを実際にどうやって実行するかは一切知りません。モデルの視点から見れば、関数は同じマシン上で動いているわけでもないため、「自分の代わりにこの関数を呼び出してください」と私たちに丁寧に依頼する以外に方法がないのです。

前回の記事でこの手順が不要だったのは、Automatic Function Calling が裏で面倒を見てくれていたからです。内部的な呼び出しフロー自体はまったく同じですが、SDK がこの複雑さをすべて隠蔽してくれていたわけです。

次にやるべきことは当然、実際の関数を呼び出してその結果をモデルに返すことです。ただし思い出してください。コンテキストがなければ、モデルは直前のリクエストについて何も覚えていません。そのため、関数の実行結果だけをポンと送り返しても、モデルは何をすればいいのか理解できません！

だからこそ、これまでの対話履歴（少なくともモデルがその値を要求した時点まで遡る履歴）を一緒に送信する必要があります。以下のコードは、関数呼び出しメッセージを受け取った後、完全な情報を含めて新たなリクエストを送信する実装です：

```python
# レスポンスを検査し、モデルが要求している内容を把握したと仮定
result = get_random_number() # 実際の関数呼び出しを実行

# contents には元のプロンプトが残っているため、モデルの応答を追加し...
contents.append(types.ModelContent(parts=response.candidates[0].content.parts))
# ... さらに関数呼び出しの結果を追加する
contents.append(types.UserContent(parts=types.Part.from_function_response(name="get_random_number", response={"result": result})))

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config,
)
print(response.text)
```

出力：

```
$ python3 main.py 
Today's lucky number is 4.
```

## おわりに

この記事では、エージェントクライアントがサーバー側のモデルとどのように通信しているか、言い換えれば LLM 通信の「ドメインモデル」がどのようになっているかを見てきました。また、Python SDK が裏で行ってくれている「魔法」のベールも剥がしてみました。

自動化は常に便利で、迅速に結果を出すのに役立ちます。しかし、エージェントを自作する際、実際に何が起きているのかを知っているかどうかは、開発がスムーズに進むかトラブルに見舞われるかの決定的な違いになります。とりわけエッジケースの処理は、*決して一筋縄ではいかない*からです。

バイブス駆動開発（Vibe Coding）が全盛の今、このようなことを言うのは一見皮肉に聞こえるかもしれません。しかし、私がバイブコーディングを通じてすぐに学んだことの 1 つは、AI に対してより的確かつ正確に指示を伝えられるほど、はるかに短い時間で圧倒的に優れた結果が得られるということです。だからこそ、今は基礎知識の価値を過小評価する時ではなく、むしろそれを深めるべき時なのです。AI があるから知識が不要になるのではなく、**AI があるからこそ**知識が必要になるのです。

ここまでの道のりを楽しんでいただけたなら幸いです。本シリーズの次回 [システム指示とエージェントツールの実践ガイド]({{< ref "/posts/20250611-system-prompt" >}}) では、今回の知識をベースに、セッション履歴の管理や動的なテーブルスキーマの注入を備えた本格的なターミナルチャットアプリケーションを構築します。

ぜひ下のコメント欄でご意見やご感想をお聞かせください！ それではまた！ o/
