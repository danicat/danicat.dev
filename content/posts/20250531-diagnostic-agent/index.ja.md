---
categories:
- Agent Development
date: '2025-05-31T01:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 1
summary: Gemini と Vertex AI Agent Engine を使用して、自然言語で対話できるシステム診断エージェントを作成する方法を解説します。
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: "AIエージェントを使って自分のコンピューターを「USSエンタープライズ」に変えた方法"
slug: "diagnostic-agent-uss-enterprise"
aliases:
  - "/ja/posts/20250531-diagnostic-agent/"
description: "Gemini と Vertex AI Agent Engine、osquery の Function Calling を使って、自然言語でPCの健全性を診断するスタートレック風アシスタントを構築。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-cloud-aiplatform"
  - "osquery"
---
_宇宙、それは人類に残された最後の開拓地である。そこには人類の想像を絶する新しい文明、新しい生命が待ち受けているに違いない。これは人類初の試みとして5年間の調査飛行に飛び立った宇宙探査船USSエンタープライズ号の驚異に満ちた物語である。_

## はじめに

子供の頃、父の影響で私は毎日のようにこのオープニングナレーションを聞いて育ちました。父のスタートレックへの情熱が、私がソフトウェアエンジニアの道を選ぶ上で大きなきっかけになったのだと思います。（スタートレックをご存じない方のために補足すると、このナレーションは初代スタートレックシリーズの全エピソードの冒頭で流れていたものです）

スタートレックは常に時代を先取りしていました。そのような描写が大きな議論を呼んだ時代に、[米国のテレビ史上初となる異人種間のキスシーン](https://en.wikipedia.org/wiki/Kirk_and_Uhura%27s_kiss)を放映し、スマートフォンやビデオ会議など、今日では当たり前となった数多くの「未来的」テクノロジーも描いていました。

特に印象的なのは、劇中のエンジニアたちがコンピューターとやり取りする方法です。キーボードを叩いたりボタンを押したりする場面も時折見られますが、多くのコマンドは自然言語の音声で指示されます。「レベル1診断手順（level 1 diagnostic procedure）を実行せよ」といった象徴的な命令は、あまりに頻繁に登場するため、熱心なファンの間では[定番のネタ（ジョーク）](https://www.youtube.com/watch?v=cYzByQjzTb0)になっているほどです。

それから30年以上が経ち、私たちはインターネット以上の変革をもたらすと言われる「AI の時代」を迎えています。AI が仕事に与える影響を不安視する声も多くありますが（[先週の記事「誰でもvibe codeできるのか？」]({{< ref "/posts/20250528-vibe-coding" >}}) でも触れました）、スタートレックを見て育った私には、今後数年でエンジニアの役割がどのように変化していくのかが自然とイメージできます。コードを1行ずつ書き、コンパイラを通じて手動で指示を与える代わりに、私たちは間もなくコンピューターと直接対話し、ブレインストーミングを行いながら開発を進めるようになるでしょう。

これを具体的にイメージできるように、今回は現在利用可能なテクノロジーを使って、自然言語で自分のマシンと対話できる小さな診断エージェントを作成してみます。

## このデモに必要なもの

開発言語には、実験がしやすい Jupyter Notebook 上の Python を使用します。主に使用するツールとライブラリは以下の通りです：

*   [Vertex AI Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/overview?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog)
*   [Osquery](https://www.osquery.io/)（および [Python バインディング](https://github.com/osquery/osquery-python)）
*   [Jupyter Notebook](https://jupyter.org/) [任意]（私は [VS Code の Jupyter 拡張機能](https://code.visualstudio.com/docs/datascience/jupyter-notebooks) を使用しています）

以下のサンプルでは Gemini 2.0 Flash を使用しますが、お好みの [Gemini モデルバリアント](https://ai.google.dev/gemini-api/docs/models) を使用できます。クラウド上のサーバーではなくローカルマシンの診断を行うため、今回はエージェントを Google Cloud にデプロイすることはしません。

## エージェントの概要

エージェント技術の仕組みをすでにご存じの方は、このセクションをスキップして構いません。

AI エージェントとは、周囲の環境を認識し、特定の目標を達成するために自律的なアクションを実行できる AI システムです。入力に基づいてコンテンツを生成することに特化した従来の大規模言語モデル（LLM）と比較すると、AI エージェントは環境と対話し、意思決定を行い、目的を達成するためのタスクを実行できます。これは、エージェントに情報を提供し、アクションを実行可能にする「ツール（tools）」を使用することで実現されます。

エージェント技術を体験するために、今回は [Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog) 経由で LangChain を使用します。まず、必要なパッケージをシステムにインストールします：

```shell
pip install --upgrade --quiet google-cloud-aiplatform[agent_engines,langchain]
```

また、gcloud の Application Default Credentials（ADC）を設定する必要があります：

```shell
gcloud auth application-default login
```

注：このデモを実行する環境によっては、別の認証方法が必要になる場合があります。

これで Python スクリプトの作成準備が整いました。まず、Google Cloud のプロジェクト ID とロケーションに基づいて SDK を初期化します：

```python
import vertexai

vertexai.init(
    project="my-project-id",                  # あなたのプロジェクト ID
    location="us-central1",                   # クラウドのロケーション
    staging_bucket="gs://my-staging-bucket",  # ステージング用バケット
)
```

初期設定が完了したら、Agent Engine で LangChain を使用したエージェントの作成は非常にシンプルです：

```python
from vertexai import agent_engines

model = "gemini-2.0-flash" # ぜひ他のモデルも試してみてください！

model_kwargs = {
    # temperature (float): サンプリング温度。トークン選択のランダム性を制御します。
    "temperature": 0.20,
}

agent = agent_engines.LangchainAgent(
    model=model,                # 必須
    model_kwargs=model_kwargs,  # 任意
)
```

この設定だけで、一般的な LLM に問い合わせるのと同じようにエージェントへクエリを送信できます：

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

すると、次のような応答が返ってきます：

```
{'input': 'which time is now?', 'output': 'As an AI, I don\'t have a "current" time or location in the same way a human does. My knowledge isn\'t updated in real-time.\n\nTo find out the current time, you can:\n\n*   **Check your device:** Your computer, phone, or tablet will display the current time.\n*   **Do a quick search:** Type "what time is it" into a search engine like Google.'}
```

設定やプロンプト、そしてモデルの挙動の揺らぎによっては、時刻を答えられないと返すか、あるいは「ハルシネーション（幻覚）」を起こして架空のタイムスタンプをでっち上げるかのどちらかになります。実際のところ、AI は時計を持っていないため、時計というツールを与えない限りこの質問には答えられません。

## Function Calling（関数呼び出し）

エージェントの機能を拡張する最も便利な方法の 1 つは、呼び出し可能な Python 関数を渡すことです。手順自体は非常にシンプルですが、関数の docstring（ドキュメント）を詳細に記述するほど、エージェントが適切に関数を呼び出しやすくなる点は重要です。まずは現在時刻を確認する関数を定義してみましょう：

```python
import datetime

def get_current_time():
    """Returns the current time as a datetime object.

    Args:
        None
    
    Returns:
        datetime: current time as a datetime type
    """
    return datetime.datetime.now()
```

システム時刻を返す関数が用意できたので、この関数の存在を認識させた上でエージェントを再作成します：

```python
agent = agent_engines.LangchainAgent(
    model=model,                # 必須
    model_kwargs=model_kwargs,  # 任意
    tools=[get_current_time]
)
```

そして、もう一度同じ質問を投げます：

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

出力は次のようになります：

```
{'input': 'which time is now?', 'output': 'The current time is 18:36:42 UTC on May 30, 2025.'}
```

これでエージェントはツールを活用し、実際のデータに基づいて質問に答えられるようになりました。なかなか面白いと思いませんか？

## システム情報の収集

今回の診断エージェントには、[Osquery](https://www.osquery.io/) というツールを使って、実行中のマシンに関する情報を取得する機能を持たせます。Osquery は Facebook（現 Meta）が開発したオープンソースツールで、OS の内部情報を公開する「仮想テーブル」に対して SQL クエリを実行できるようにするものです。

これは非常に便利です。システムに関する問い合わせの窓口が 1 つにまとまるだけでなく、LLM は SQL クエリの作成が非常に得意だからです。

Osquery のインストール手順は [公式ドキュメント](https://osquery.readthedocs.io/en/stable/) を参照してください。マシンの OS によって手順が異なるため、ここでは割愛します。

Osquery をインストールしたら、Osquery の Python バインディングをインストールします。Python らしく、`pip install` 1 つで完了します：

```shell
pip install --upgrade --quiet osquery
```

バインディングをインストールしたら、`osquery` パッケージをインポートして Osquery の呼び出しを行えます：

```python
import osquery

# エフェメラルな拡張ソケットを使用して osquery プロセスを起動
instance = osquery.SpawnInstance()
instance.open()  # 例外が発生する可能性があります

# クエリを発行し osquery Thrift API を呼び出す
instance.client.query("select timestamp from time")
```

`query` メソッドは、クエリ結果を含む `ExtensionResponse` オブジェクトを返します。例えば以下のようになります：

```python
ExtensionResponse(status=ExtensionStatus(code=0, message='OK', uuid=0), response=[{'timestamp': 'Fri May 30 17:54:06 2025 UTC'}])
```

これまで Osquery を使ったことがない方は、[スキーマ](https://www.osquery.io/schema/5.17.0/) を確認して、お使いの OS でどのような情報が取得できるかを見てみることをおすすめします。

### 出力フォーマットに関する補足

これまでのサンプルの出力はフォーマットされていませんでしたが、Jupyter 上でコードを実行している場合は、以下のモジュールをインポートすることで出力を綺麗に整える便利なメソッドが使えます：

```python
from IPython.display import Markdown, display
```

そして、レスポンスの出力を Markdown として表示します：

```python
response = agent.query(
    input="what is today's stardate?"
)
display(Markdown(response["output"]))
```

出力：

```
Captain's Log, Supplemental. The current stardate is 48972.5.
```

## 点と点をつなぐ（診断エージェントの実装）

OS の情報を照会する方法がわかったので、エージェントの知識と組み合わせて、システムに関する質問に答えてくれる診断エージェントを作成しましょう。

最初のステップは、クエリを実行する関数を定義することです。この関数は、後で情報収集用のツールとしてエージェントに渡します：

```python
def call_osquery(query: str):
    """Query the operating system using osquery
      
      This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
      You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

      Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

      Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    return instance.client.query(query)
```

関数自体は極めてシンプルですが、ここで重要なのは、エージェントがこの関数の動作を理解できるように非常に詳細な docstring を記述することです。

テスト中によく発生した厄介な問題として、エージェントがシステム上にどのテーブルが存在するかを正確に把握していないという点がありました。例えば、私は macOS マシンで実行しているのですが、`memory_info` というテーブルは macOS には存在しません。

エージェントにより多くのコンテキストを提供するために、このシステムで利用可能なテーブル名を動的に渡すようにします。理想的な状況であれば、カラム名や説明を含むスキーマ全体を渡すところですが、Osquery ではそれを実現するのは容易ではありません。

Osquery の基盤となるデータベース技術は SQLite なので、`sqlite_temp_master` テーブルから仮想テーブルの一覧を取得できます：

```python
# Python の小技を使って、このシステムに存在するテーブル一覧を特定する
response = instance.client.query("select name from sqlite_temp_master").response
tables = [ t["name"] for t in response ]
```

テーブル名がすべて取得できたので、この情報と `call_osquery` ツールを使用してエージェントを作成します：

```python
osagent = agent_engines.LangchainAgent(
    model = model,
    system_instruction=f"""
    You are an agent that answers questions about the machine you are running in.
    You should run SQL queries using one or more of the tables to answer the user questions.
    Always return human readable values (e.g. megabytes instead of bytes, and formatted time instead of miliseconds)
    Be very flexible in your interpretation of the requests. For example, if the user ask for application information, it is acceptable to return information about processes and services. If the user requests resource usage, return BOTH memory and cpu information.
    Do not ask the user for clarification.
    You have the following tables available to you: 
    ----- TABLES -----
    {tables}
    ----- END TABLES -----

    Question:
    """,
    tools=[
        call_osquery,
    ]
)
```

信じられないかもしれませんが、これで私たちのエージェントはあらゆる質問に答える準備が整いました！実際に試してみましょう：

```python
response = osagent.query(input="what is the current time?")
display(Markdown(response["output"]))
```

出力：

```
The current time is Fri May 30 18:08:15 2025 UTC.
```

もう少し複雑な質問をしてみます：

```python
response = osagent.query(input="what is the top consuming process?")
display(Markdown(response["output"]))
```

出力：

```
The top CPU-consuming process is:
Process Name: mediaanalysisd, PID: 1127, Total CPU Time: 2876716

The top memory-consuming process is:
Process Name: Code Helper (Plugin), PID: 2987, Memory Usage: 1537 MB
```

さらに創造的な質問を投げかけてみたらどうなるでしょうか？

```python
response = osagent.query(input="computer, run a level 1 diagnostic procedure")
display(Markdown(response["output"]))
```

出力：

```
Here is the level 1 diagnostic information:

**Operating System:**
*   **Name:** macOS
*   **Version:** 15.5
*   **Build:** 24F74
*   **Platform:** darwin
*   **Architecture:** arm64

... ommitted for brevity ...

**Uptime:**
*   **Total Seconds:** 261164
*   **Days:** 3
*   **Hours:** 0
*   **Minutes:** 32
*   **Seconds:** 44

```

素晴らしいですね！「レベル1診断手順（Level 1 diagnostic procedure）」が何であるかは一切定義していなかったにもかかわらず、驚くほど本格的な診断レポートを返してくれました。

さらに突っ込んだ質問も試してみましたが、その回答も期待を裏切らないものでした（ほとんどの場合は）：

```python
response = osagent.query(input="computer, do you see any signs of malware running?")
display(Markdown(response["output"]))
```

出力：

```
I have checked for processes that are not associated with a file on disk, which can be a sign of malware, and found none. I have also examined the top processes by memory and CPU usage. The processes consuming the most resources are primarily Visual Studio Code and Google Chrome and their related helper processes. This is typical behavior for these applications.

Based on the checks performed, there are no obvious signs of malware running on the system at this time.
```

_（マイクを落としてステージを去るポーズ）_ =^.^=

## まとめ

使い古された表現かもしれませんが、AI はまさにゲームチェンジャーです。ごくわずかなコード行数で、OS の内部動作と対話できる本格的な自然言語インターフェースをゼロから構築できました。もう少し手を加えれば、より深い診断を行ったり、場合によっては自律的に問題を修復したりするようにエージェントを拡張することもできるでしょう。スコッティ機関長もきっと誇らしく思ってくれるはずです！

![マウスをマイク代わりにしてコンピューターに話しかけようとするスコッティ機関長](hello-computer-hello.gif)

この記事で紹介したすべてのサンプルのソースコードは、私の [GitHub](https://github.com/danicat/devrel/blob/main/blogs/20250531-diagnostic-agent/diagnostic_agent.ipynb) で公開しています。

シリーズの次回記事「[Vertex AI SDK for Python を深く掘り下げる]({{< ref "/posts/20250605-vertex-ai-sdk-python" >}})」では、クライアントと Gemini の通信プロトコルの裏側を探り、低レイヤーでの手動 Function Calling を実装します。

皆さんはどう感じましたか？ぜひ下のコメント欄で感想を共有してください！
