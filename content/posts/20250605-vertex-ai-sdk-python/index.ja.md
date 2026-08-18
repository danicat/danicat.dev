---
categories:
- Agent Development
date: '2025-06-05T00:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 2
summary: Vertex AI SDK for Python を使用して、クライアントコードと Gemini API 間の通信モデルとメッセージ構造、低レイヤーでの手動 Function Calling を詳しく解説します。
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: Vertex AI SDK for Python を深く掘り下げる
---
## はじめに

この記事では、[Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) を使用したクライアントコードと Gemini API 間の通信モデルについて詳しく探ります。メッセージの構造、モデルがコンテキストを理解する仕組み、Function Calling（関数呼び出し）によるモデル機能の拡張方法などを取り上げます。今回は Gemini を中心に解説しますが、ここで説明する概念は Gemma や他の LLM にも同様に適用できます。

[シリーズ第1弾]({{< ref "/posts/20250531-diagnostic-agent" >}}) では、ローカルマシンの診断に関する質問に答える、シンプルながら強力な AI エージェントの作り方を解説しました。ごくわずかなコード行数で、「PCのCPU使用率はどれくらいか」「マルウェアの兆候はないか」といった質問に回答するエージェントを実現しました。

これは Python SDK の優れた抽象化によるものでした。例えば、関数をいつ呼び出すかをエージェント自身に判断させる [Automatic Function Calling（自動関数呼び出し）](https://ai.google.dev/gemini-api/docs/function-calling?example=weather#automatic_function_calling_python_only) を利用したため、通常の Python 関数を定義するだけで SDK が関数のシグネチャや docstring を自動的に解釈してくれました。ただし、この自動機能は現在のところ Python SDK 向けに特化しており、Go、JavaScript、Java などの他の言語で開発する場合は手動での設定が必要です。

そこで今回は、Python 以外の SDK（Go、JS、Java）を扱う際にも役立つよう、Gemini API の内部動作とプロトコルを掘り下げて解説します。比較しやすいようにコード例は引き続き Python を使用しますが、ここで解説する通信モデルはすべてのプログラミング言語に共通する基礎知識です。

取り上げる主なトピックは以下の2つです：
*   クライアントとモデル間の対話（ターン）がどのように処理されるか
*   低レイヤーで手動で Function Calling を実装する方法

Python 開発者にとっても、メッセージの流れを低レイヤーで理解しておくことは、Live API などの高度な機能を扱ったり、実用的なエージェントを構築する上で不可欠です。

## API の仕組みを理解する

AI エージェントは一般的にクライアント・サーバー構成で動作します。リクエストを準備して送信するクライアントコンポーネントと、モデルランタイムをホストしてリクエストを処理するサーバープロセスが存在します。

Vertex AI には主に2つの API グループがあります：クライアントがリクエストを送信して応答を待つ標準的なリクエスト/レスポンス型の REST API と、WebSockets を介して双方向ストリーミングを行う [Live API](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) です。今回はまず基本となる REST API に焦点を当てます。

コンテンツ生成では、テキスト、画像、音声、動画の各モダリティを扱えます。最新のモデルはマルチモーダルネイティブであり、複数のモダリティを入力・出力で同時に処理できます。まずは最もシンプルなテキストから見ていきましょう。

単発のリクエストを行う標準的なコードは次のようになります：

```python
from google import genai

client = genai.Client(
    vertexai=True,
    project="daniela-genai-sandbox",
    location="us-central1"
)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="今日の調子はいかがですか？"
)
print(response.text)
```

出力：

```
順調に稼働しています、お気遣いありがとうございます！大規模言語モデルとして感情を持つことはありませんが、正常に動作しており、サポートの準備ができています。本日はどのようなご用件でしょうか？
```

まず Vertex AI モード（`vertexai=True`）または Gemini API キーでクライアントを初期化します。

初期化後、`client.models.generate_content` メソッドでプロンプトを送信します。使用するモデル（ここでは `gemini-2.0-flash`）とプロンプト文字列を `contents` 引数に指定します。

Python のコードはシンプルに見えますが、内部では重要な抽象化が行われています。実は、**コンテンツは単なる文字列ではありません**。

`contents` は実際には `Content` 構造体のリストであり、各構造体は **role（役割）** と 1 つ以上の **parts（パーツ）** で構成されています。型ライブラリでの定義は以下のようになっています：

```python
from google.genai import types

contents = [types.Content(
  role = "user",
  parts = [ types.Part.from_text("今日の調子はいかがですか？") ]
)]
```

私たちが `contents="今日の調子はいかがですか？"` と指定すると、Python SDK が自動的に「テキストパートを持つコンテンツオブジェクト」に変換してくれています。

もう 1 つの重要なポイントは、`generate_content` を呼び出すたびに、モデルは事前の記憶を持たずにゼロから推論を開始するということです。つまり、過去のメッセージのコンテキストをプロンプトに付与するのは**クライアント側の責務**です。同じ質問を2回連続で行う実験をしてみましょう：

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="今日は何日ですか？"
)
print(response.text)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="今日は何日ですか？"
)
print(response.text)
```

出力：

```
$ python3 main.py 
今日は2023年11月5日日曜日です。

今日は2024年11月2日土曜日です。
```

ここには2つの問題があります：1) モデルは時計を持たないためハルシネーションを起こしている点、2) 同じ質問に対して2回とも異なる日付を返している点です。1つ目の問題はツール（時刻取得関数や Google 検索）によるグラウンディングで解決できますが、2つ目の問題は、モデルが直前に自分が何を答えたかすら覚えていないことを示しています。

会話のコンテキストを維持するようにコードを修正してみます：

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="今日は何日ですか？"
)
print(response.text)

# contents 配列の各要素は会話の「ターン（Turn）」と呼ばれます
contents = [
    {
        "role": "user",
        "parts": [{
            "text": "今日は何日ですか？"
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
            "text": "今日は何日ですか？"
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
今日は2023年11月15日水曜日です。

今日は2023年11月15日水曜日です。
```

2回目の呼び出しでは、過去のやり取り全体を `contents` に含めて送信しています。`role` が `"user"` と `"model"` で交互に変化している点に注目してください。モデルはこの役割と順序を見て会話の流れ（ターン）を理解します。

SDK には `types.UserContent` や `types.Part.from_text` などのヘルパーメソッドも用意されています。

画像などのマルチモーダルデータを渡す場合は、専用のパート構造体を使用します：

```python
from google.genai import types

contents = types.Part.from_uri(
  file_uri='gs://generativeai-downloads/images/scones.jpg',
  mime_type='image/jpeg',
)
```

会話履歴を手動で管理するのは手間がかかるため、Vertex AI SDK にはこれを自動化する `chats` API が用意されています：

```python
chat = client.chats.create(model='gemini-2.0-flash')
response = chat.send_message('今日は何日ですか？')
print(response.text)
response = chat.send_message('今日は何日ですか？')
print(response.text)
```

出力：

```
$ python3 main.py 
今日は2023年10月14日土曜日です。

今日は2023年10月14日土曜日です。
```

`chat` オブジェクトが裏で過去のターン履歴を保持してくれているため、一貫した対話が可能になります。

## 手動での Function Calling（関数呼び出し）

メッセージ構造とコンテキスト管理がわかったところで、Function Calling が低レイヤーでどのようにやり取りされているかを見ていきます。

モデルに関数を使わせるには、まず「どのような関数が存在し、何の引数を取るか」をモデルに宣言（Function Declaration）する必要があります。

例えば、`get_random_number` 関数の宣言は次のようになります：

```python
get_random_number_decl = {
    "name": "get_random_number",
    "description": "ランダムな数値を返します",
}
```

宣言には `name`、`description`、`parameters` の3つのフィールドがあります（引数を取らない関数の場合は parameters を省略できます）。モデルはこの説明文を読み、いつどの関数を呼び出すべきかを判断します。

前回の記事では docstring から SDK に自動推論させましたが、今回は手動で明示的に宣言を作成します：

```python
def get_random_number():
    return 4 # 公平なサイコロの出目によって決定
             # 完全にランダムであることが保証されています (https://xkcd.com/221/)

get_random_number_decl = {
    "name": "get_random_number",
    "description": "ランダムな数値を返します",
}
```

この宣言をツールとしてモデルの設定に追加します：

```python
tools = types.Tool(function_declarations=[get_random_number_decl])
config = types.GenerateContentConfig(tools=[tools])

contents = [types.Part.from_text(text="今日の私のラッキーナンバーは何ですか？")]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config,
)

print(response.candidates[0].content.parts[0])
```

実行すると、次のような結果が得られます：

```
$ python3 main.py 
video_metadata=None thought=None inline_data=None file_data=None thought_signature=None code_execution_result=None executable_code=None function_call=FunctionCall(id=None, args={}, name='get_random_number') function_response=None text=None
```

返ってきたレスポンスの各フィールドは空（`None`）で、`function_call` フィールドだけが埋まっています。これは、モデルが最終回答を出力する代わりに、「`get_random_number` 関数を実行して結果を教えてください」と要求してきたことを意味します。

モデルは関数の存在を知っていますが、Google のサーバー上から直接私たちのローカルコードを実行することはできません。そのため、クライアント側で関数を実行して結果を返すよう要求してきます。

そこで、ローカルで関数を実行し、その結果を会話履歴に含めて再度モデルへ送信します：

```python
# 1. モデルが要求した関数を実行
result = get_random_number()

# 2. 直前のモデルの応答（関数呼び出しリクエスト）を履歴に追加
contents.append(types.ModelContent(parts=response.candidates[0].content.parts))

# 3. 関数の実行結果をユーザーコンテンツとして履歴に追加
contents.append(types.UserContent(parts=types.Part.from_function_response(name="get_random_number", response={"result": result})))

# 4. 更新された履歴全体をモデルに送信
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
今日のあなたのラッキーナンバーは 4 です。
```

## おわりに

この記事では、エージェントクライアントとモデルがどのようなメッセージ構造とターンで通信しているか、そして手動での Function Calling の流れを解明しました。

SDK の自動機能は生産性を高める上で非常に便利ですが、プロトコルレベルでの動作を把握しておくことで、エージェント開発におけるトラブルシューティングや他言語 SDK の活用が格段にスムーズになります。

シリーズ第3弾の [システム指示とエージェントツールの実践ガイド]({{< ref "/posts/20250611-system-prompt" >}}) では、この知識を活用して、セッション履歴の保持とテーブルスキーマの動的注入を備えた本格的なターミナルチャットアプリケーションを作成します。

ご質問やご意見がありましたら、ぜひコメント欄でお知らせください！
