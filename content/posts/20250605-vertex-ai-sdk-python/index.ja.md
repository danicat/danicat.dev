---
date: '2025-06-05T00:00:00+01:00'
title: 'Vertex AI SDK for Pythonを深く掘り下げる'
summary: "この記事では、Vertex AI SDK for Pythonを使用したクライアントコードとGemini API間の通信モデルについて説明します"
categories: ["AI & Development"]
tags: ["gemini", "vertex-ai", "python", "tutorial"]
---
{{< translation-notice >}}

## はじめに

この記事では、[Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog)を使用したクライアントコードとGemini API間の通信モデルについて説明します。メッセージの構造、モデルが質問のコンテキストをどのように理解するか、関数呼び出しでモデルの機能を拡張する方法などの概念について説明します。この記事の焦点はGeminiですが、ここで説明する同じ概念はGemmaや他のLLMにも適用できます。

[前回の投稿](https://danicat.dev/posts/20250531-diagnostic-agent/)では、ローカルマシンに関する診断の質問に答える、シンプルでありながら驚くほど強力なAIエージェントを作成する方法を説明しました。非常に少ないコード行（そしてそれほど少なくないコメント行）で、「マシンにどれくらいのCPUがありますか」や「マルウェアの兆候がないか確認してください」などのクエリに応答するエージェントを取得できました。

もちろん、それはPython SDKの美しさによるものであり、物事を大幅に簡素化してくれました。たとえば、エージェントがいつ関数を呼び出すかを決定させるために、[自動関数呼び出し](https://ai.google.dev/gemini-api/docs/function-calling?example=weather#automatic_function_calling_python_only)という機能に依存していました。この機能は、関数をプレーンなPython関数として定義するだけで、SDKがそのシグネチャと説明を動的に把握してくれるという点でも役立ちました。残念ながら、この機能はPython SDKでのみ利用可能であるため、他の言語の開発者はもう少し作業を行う必要があります。

そのため、今日の記事では少し異なるアプローチを取り、Gemini APIがどのように機能するかを説明し、Pythonだけでなく、利用可能なSDK（JS、Go、Java）のいずれかを使用する準備を整えることができるようにします。例には引き続きPythonを使用するため、前の記事と比較できますが、ここで説明する概念はすべての異なる言語で有効です。

2つの主要なトピックについて説明します。
*   クライアントとモデル間の会話の仕組み
*   手動で関数呼び出しを実装する方法

Python開発者であっても、この記事から何も得られないわけではないことに注意してください。実際、会話の流れを理解することは、SDKのより高度な概念（Live APIなど）を使用したり、一般的にLLMを操作したりするために重要になります。

## APIの仕組みを理解する

エージェントは通常、クライアントサーバーアプリケーションと同じように機能します。リクエストの準備と作成を担当するクライアントコンポーネントと、モデルランタイムをホストしてクライアントリクエストを処理するサーバープロセスがあります。

Vertex AIには、主に2つのAPIグループがあります。1つは、クライアントがリクエストを送信し、続行する前に応答を待つ、典型的なリクエスト/レスポンススタイルのコンテンツ生成用のREST APIです。もう1つは、Webソケットを使用してリアルタイム情報を処理する新しい[Live API](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog)です。Live APIは正しく動作させるために少し準備が必要なため、まずはREST APIに焦点を当てます。

通常、コンテンツは、テキスト、画像、音声、ビデオのいずれかのモダリティで生成されます。最新のモデルの多くはマルチモーダルでもあり、入力と出力の両方で複数のモダリティを同時に扱うことができます。物事を単純にするために、まずはテキストから始めましょう。

典型的な1回限りのプロンプトアプリケーションは次のようになります。

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

最初に行う必要があるのは、Vertex AIモード（`vertexai=True`）を使用するか、Gemini APIキーを使用してクライアントをインスタンス化することです。この場合、私はVertex AIモードを使用しています。

クライアントが初期化されると、`client.models.generate_content`メソッドを使用してプロンプトを送信できます。どのモデルを呼び出しているか（この場合は`gemini-2.0-flash`）と、contents引数にプロンプト（例：「How are you today?」）を指定する必要があります。

このコードを見ると、Pythonのおかげで多くの抽象化が無料で提供されているため、内部で何が起こっているのかを想像するのは難しいかもしれません。この場合、最も重要なことは、**コンテンツは文字列ではない**ということです。

コンテンツは実際にはコンテンツ構造のリストであり、コンテンツ構造は**ロール**と1つ以上の**パート**で構成されています。この構造の基になる型は、typesライブラリで定義されており、次のようになります。


```python
from google.genai import types

contents = [types.Content(
  role = "user",
  parts = [ types.Part_from_text("How are you today?")
)]
```

したがって、`contents="How are you today?"`と入力するたびに、Python SDKは文字列から「文字列パートを持つコンテンツ」へのこの変換を自動的に行います。

もう1つ注意すべき重要な点は、generate_contentを呼び出すたびに、モデルはゼロから開始しているということです。これは、次のプロンプトに前のメッセージのコンテキストを追加するのは私たちの責任であることを意味します。モデルに今日が何日かを2回連続で尋ねる簡単なテストをしてみましょう。

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

上記の応答には2つの問題があります。1）モデルは日付を知る方法がないため、幻覚を起こしました。2）同じ質問に対して2つの異なる答えを返しました。1）はdatetime呼び出しやGoogle検索などのツールでグラウンディングすることで修正できますが、2）に焦点を当てたいと思います。なぜなら、モデルが言ったことを覚えていないことを明確に示しており、会話についてモデルを最新の状態に保つのは**私たちの**責任であるという上記の点を実証しているからです。

コードを少し変更してみましょう。

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)

# each element in the contents array is usually referred as a "turn"
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

2回目のモデルへの呼び出しでは、contents属性にコンテキスト全体を含めていることに注意してください。また、各パートのロールが「user」から「model」に変わり、再び「user」に変わることに注意してください（「user」と「model」はロールの唯一の可能な値です）。これが、モデルが会話のどの時点にいるか、つまり「ターン」を理解する方法です。たとえば、質問を繰り返す最後のパートを省略した場合、モデルは最新の状態であると考え、最後のターンが「user」ではなく「model」になるため、別の応答を生成しません。

上記のcontents変数は「辞書」形式で記述されていますが、SDKは`types.UserContent`（ロールフィールドを自動的に「user」に設定）や`types.Part.from_text`（プレーンな文字列をパートに変換）など、いくつかの便利なメソッドも提供しています。

他の種類の入力や出力を扱うには、関数呼び出し、バイナリデータなど、他の種類のパートを使用できます。モデルがマルチモーダルの場合、同じメッセージ内で異なるコンテンツタイプのパートを混在させることができます。

バイナリデータは、インラインまたはURIからフェッチできます。mime_typeフィールドを使用して、さまざまな種類のデータを区別できます。たとえば、画像パートは次のように取得できます。

```python
from google.genai import types

contents = types.Part.from_uri(
  file_uri: 'gs://generativeai-downloads/images/scones.jpg',
  mime_type: 'image/jpeg',
)
```

またはインライン：

```python
contents = types.Part.from_bytes(
  data: my_cat_picture, # binary data
  mime_type: 'image/jpeg',
)
```

要約すると、会話の各ターンで、前のモデルの応答と新しいユーザーの質問の両方に対して、新しいコンテンツ行を追加します。

良いニュースは、チャットボットエクスペリエンスは非常に重要なユースケースであるため、Vertex AI SDKがこのフローの実装をすぐに提供していることです。`chat`機能を使用すると、上記の動作を非常に少ないコード行で再現できます。

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

今回は、チャットインターフェイスが履歴を自動的に処理しているため、モデルは日付を覚えていました。

## 非自動関数呼び出し

APIがクライアントメッセージを構築し、コンテキストを管理する方法を見てきたので、次に関数呼び出しをどのように処理するかを探ります。基本レベルでは、モデルに関数が利用可能であることを指示し、その後、関数を呼び出して結果の値をモデルに返すリクエストを処理する必要があります。これは、関数呼び出しによってエージェントが外部システムや現実世界と対話し、単にテキストを生成するだけでなく、データの取得や特定のプロセスのトリガーなどのアクションを作成できるため、重要です。

関数宣言は、モデルに何ができるかを伝えるものです。関数名、説明、引数をモデルに伝えます。たとえば、以下は`get_random_number`関数の関数宣言です。

```python
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Returns a random number",
}
```

モデルがどの関数を呼び出すかを決定するために知る必要があるのは、この宣言です。関数宣言には、名前、説明、パラメータの3つのフィールドがあります。この場合、関数はパラメータを受け付けないため、このフィールドは省略されています。モデルは、関数の説明とその引数の説明を使用して、各関数をいつ、どのように呼び出すかを決定します。

前の記事では、モデルに関数宣言を与える代わりに、怠けてSDKにdocstringに基づいてそれを把握させました。今回は、基になるフローをよりよく理解するために、明示的に関数を宣言するという異なる方法をとります。

宣言を含む関数は次のようになります。

```python
def get_random_number():
    return 4 # chosen by fair dice roll
             # guaranteed to be random (https://xkcd.com/221/)

# the declaration tells the model what it needs to know about the function
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Returns a random number",
}
```

関数宣言の他の例は[こちら](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling#schema-examples?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog)で確認できます。

次に、モデルにこの関数にアクセスできることを伝える必要があります。これは、モデル構成を介して、関数をツールとして追加することで行います。

```python
tools = types.Tool(function_declarations=[get_random_number_decl])
config = types.GenerateContentConfig(tools=[tools])

# my initial prompt
contents = [types.Part.from_text(text="what is my lucky number today?")]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config, # note how we are adding the config to the model call
)

print(response.candidates[0].content.parts[0])
```

上記のコードを実行すると、次のようになります。

```
$ python3 main.py 
video_metadata=None thought=None inline_data=None file_data=None thought_signature=None code_execution_result=None executable_code=None function_call=FunctionCall(id=None, args={}, name='get_random_number') function_response=None text=None
```

ここで見ているのはモデル応答の最初の部分であり、この部分には`function_call`フィールドを除くすべてのフィールドが空（`None`）であることがわかります。これは、モデルが**私たち**にこの関数呼び出しを行い、その結果をモデルに返すことを望んでいることを意味します。

これは最初は私を困惑させましたが、考えてみれば理にかなっています。モデルは関数が存在することを知っていますが、それを呼び出す方法をまったく知りません。モデルの観点からすると、関数は同じマシンで実行されているわけではないため、モデルは「丁寧に」私たちに代理で呼び出しを行うように頼む以外に何もできません。

前の記事では、自動関数呼び出しが引き継いで物事を簡素化してくれたため、これを行う必要はありませんでした。呼び出しは同じフローに従いましたが、SDKはこの複雑さをすべて隠していました。

今すぐ行うべき明らかなことは、実際の関数を呼び出して結果をモデルに返すことですが、コンテキストがないとモデルは前のリクエストについて何も知らないため、関数の結果だけを返しても、それで何をすべきかわからないことを覚えておいてください！

そのため、これまでの対話の履歴を送信する必要があり、少なくともモデルがその値を要求したことを知っている時点までさかのぼる必要があります。以下のコードは、関数呼び出しメッセージを受け取り、完全な情報を含む新しいリクエストを送信する必要があることを前提としています。

```python
# assuming we already inspected the response and know what the model wants
result = get_random_number() # makes the actual function call

# contents still contain the original prompt so we will add the model response...
contents.append(types.ModelContent(parts=response.candidates[0].content.parts))
# ... and the result of the function call
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

## 結論

この記事では、エージェントクライアントがサーバー側のモデルとどのように通信するか、つまりLLM通信の「ドメインモデル」を見てきました。また、Python SDKが私たちのために行う「魔法」のカーテンを取り除きました。

自動化は常に便利で、結果をはるかに速く達成するのに役立ちますが、実際にどのように機能するかを知ることは、通常、独自のエージェントを実装する際の円滑な道のりと途切れ途切れの道のりの大きな違いです。特に、特殊なケースは*決して簡単ではない*ためです。

vibe codingの時代に、一見するとこのようなことを言うのはほとんど皮肉なことだとわかっていますが、vibe codingで私がすぐに学んだことの1つは、AIと話すときにより正確であれば、はるかに短い時間ではるかに良い結果が得られるということです。ですから、今は知識の価値を軽視する時ではなく、AIのせいではなく**AIのおかげで**知識を倍増させる時です。

これまでの旅を楽しんでいただけたことを願っています。次の記事では、この知識を基に、診断エージェントを次のレベルに引き上げ、これまでエージェントが行ったことのない場所に到達させます！（あるいは、行ったことがあるかもしれませんが、確かに私のエージェントではありません=^_^=）

以下のコメントにあなたのコメントを書いてください！ピースアウトo/
