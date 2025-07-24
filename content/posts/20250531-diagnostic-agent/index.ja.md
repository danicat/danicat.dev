+++
date = '2025-05-31T01:00:00+01:00'
title = 'AIエージェントを使って自分のコンピューターを「USSエンタープライズ」に変えた方法'
summary = "GeminiとVertex AI Agent Engineを使って自然言語を話す診断エージェントを作成する方法"
tags = ["gemini", "vertex ai", "python"]
+++
{{< translation-notice >}}

_宇宙：最後のフロンティア。これらは宇宙船エンタープライズ号の航海である。その5年間の任務は、奇妙な新しい世界を探検し、新しい生命と新しい文明を探し、これまで誰も行ったことのない場所に大胆に行くことである。_

## はじめに

子供の頃、父の影響で、ほぼ毎日これらの言葉を聞いて育ちました。彼のスタートレックへの情熱が、私がソフトウェアエンジニアリングのキャリアを選ぶ上で大きな役割を果たしたのではないかと疑っています。（スタートレックに詳しくない方のために、このスピーチはオリジナルのスタートレックシリーズのすべてのエピソードの冒頭で再生されました）

スタートレックは常に時代を先取りしていました。そのようなシーンが多くの論争を引き起こした時代に、[米国テレビで初めての人種間のキス](https://en.wikipedia.org/wiki/Kirk_and_Uhura%27s_kiss)を放映しました。また、スマートフォンやビデオ会議など、今日では当たり前になっている多くの「未来的な」テクノロジーを描写しました。

本当に注目すべきことの1つは、シリーズのエンジニアがコンピューターとどのように対話するかです。時々キーボードやボタンを押すのを見かけますが、コマンドの多くは自然言語で発声されます。彼らがコンピューターに与えるコマンドのいくつかは非常に象徴的で、たとえば、コンピューターに「レベル1の診断手順」を実行するように要求するときなどです。これは非常に頻繁に起こったため、最も熱心なファンの間では[冗談](https://www.youtube.com/watch?v=cYzByQjzTb0)になりました。

30年以上早送りして、私たちはAIの時代にいます。これはインターネットよりも大きな革命になると約束されている技術革命です。もちろん、多くの人々がAIが自分の仕事にどのように影響を与えるかを恐れていますが（[先週それについて書きました](https://danicat.dev/posts/20250528-vibe-coding/)）、スタートレックを見て育った私にとっては、今後数年間でエンジニアの役割がどのように変化するかを想像するのは簡単です。テキストを介してコンピューターに命令し、コードとコンパイラの行を介して各ステップを手動で指示する代わりに、私たちはすぐにコンピューターと話したりブレインストーミングしたりする方向に進むでしょう。

これを人々に視覚化してもらうために、今日は私たちが持っているテクノロジーを使用して、自然言語を使用して自分のマシンと対話できる小さなエージェントを作成します。

## このデモに必要なもの

開発言語には、実験に非常に適しているため、Jupyter NotebookでPythonを使用します。使用する主なツールとライブラリは次のとおりです。

*   [Vertex AI Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/overview?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog)
*   [Osquery](https://www.osquery.io/)と[pythonバインディング](https://github.com/osquery/osquery-python)
*   [Jupyter Notebook](https://jupyter.org/) [オプション]（私は実際には[VSCode用のJupyterプラグイン](https://code.visualstudio.com/docs/datascience/jupyter-notebooks)を使用しています）

以下の例ではGemini Flash 2.0を使用しますが、[Geminiモデルのバリアント](https://ai.google.dev/gemini-api/docs/models)はどれでも使用できます。今回は、クラウド上のサーバーではなくローカルマシンに関する質問に答えるために使用したいため、このエージェントをGoogle Cloudにデプロイしません。

## エージェントの概要

エージェントテクノロジーの仕組みにすでに精通している場合は、このセクションをスキップできます。

AIエージェントは、環境を認識し、特定の目標を達成するために自律的な行動をとることができるAIの一形態です。主に入力に基づいてコンテンツを生成することに焦点を当てている典型的な大規模言語モデル（LLM）と比較して、AIエージェントは環境と対話したり、決定を下したり、目的を達成するためにタスクを実行したりできます。これは、エージェントに情報を提供し、アクションを実行できるようにする「ツール」を使用することで実現されます。

エージェントテクノロジーを実証するために、[Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog)を介してLangChainを使用します。まず、システムに必要なパッケージをインストールする必要があります。

```shell
pip install --upgrade --quiet google-cloud-aiplatform[agent_engines,langchain]
```

また、gcloudアプリケーションのデフォルト認証情報（ADC）を設定する必要があります。

```shell
gcloud auth application-default login
```

注：このデモを実行している環境によっては、別の認証方法を使用する必要がある場合があります。

これで、Pythonスクリプトの作業準備が整いました。まず、Google CloudプロジェクトIDと場所に基づいてSDKを初期化します。

```python
import vertexai

vertexai.init(
    project="my-project-id",                  # Your project ID.
    location="us-central1",                   # Your cloud location.
    staging_bucket="gs://my-staging-bucket",  # Your staging bucket.
)
```

初期設定が完了したら、Agent EngineでLangChainを使用してエージェントを作成するのは非常に簡単です。

```python
from vertexai import agent_engines

model = "gemini-2.0-flash" # feel free to try different models!

model_kwargs = {
    # temperature (float): The sampling temperature controls the degree of
    # randomness in token selection.
    "temperature": 0.20,
}

agent = agent_engines.LangchainAgent(
    model=model,                # Required.
    model_kwargs=model_kwargs,  # Optional.
)
```

上記の設定は、LLMにクエリを送信するのと同じように、エージェントにクエリを送信するのに十分です。

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

これは次のようなものを返す可能性があります。

```
{'input': 'which time is now?', 'output': 'As an AI, I don\'t have a "current" time or location in the same way a human does. My knowledge isn\'t updated in real-time.\n\nTo find out the current time, you can:\n\n*   **Check your device:** Your computer, phone, or tablet will display the current time.\n*   **Do a quick search:** Type "what time is it" into a search engine like Google.'}
```

設定、プロンプト、宇宙のランダム性によっては、モデルは時間を教えられないという応答を返すか、「幻覚」を起こしてタイムスタンプを作成する可能性があります。しかし、実際には、AIには時計がないため、この質問に答えることはできません…時計を与えない限り！

## 関数呼び出し

エージェントの機能を拡張する最も便利な方法の1つは、呼び出すPython関数を与えることです。プロセスは非常に単純ですが、関数のドキュメントが優れているほど、エージェントが呼び出しを正しく行うのが簡単になることを強調することが重要です。時間をチェックする関数を定義しましょう。

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

システム時刻を返す関数ができたので、エージェントを再作成しますが、今回は関数が存在することを認識させます。

```python
agent = agent_engines.LangchainAgent(
    model=model,                # Required.
    model_kwargs=model_kwargs,  # Optional.
    tools=[get_current_time]
)
```

そして、もう一度質問します。

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

出力は次のようになります。

```
{'input': 'which time is now?', 'output': 'The current time is 18:36:42 UTC on May 30, 2025.'}
```

これで、エージェントはツールに依存して、実際のデータで質問に答えることができます。かなりクールでしょう？

## システム情報の収集

診断エージェントには、[osquery](https://www.osquery.io/)というツールを使用して、実行中のマシンに関する情報をクエリする機能を与えます。Osqueryは、Facebookが開発したオープンソースツールで、ユーザーがマシンの基盤となるオペレーティングシステムに関する情報を公開する「仮想テーブル」にSQLクエリを実行できるようにします。

これは、システムに関するクエリを実行するための単一のエントリポイントを提供するだけでなく、LLMもSQLクエリの作成に非常に習熟しているため、私たちにとって便利です。

osqueryのインストール方法については、[公式ドキュメント](https://osquery.readthedocs.io/en/stable/)を参照してください。マシンのオペレーティングシステムによって異なるため、ここでは再現しません。

osqueryをインストールしたら、osqueryのPythonバインディングをインストールする必要があります。典型的なPythonと同様に、pipインストールは1つだけです。

```shell
pip install --upgrade --quiet osquery
```

バインディングをインストールすると、osqueryパッケージをインポートしてosquery呼び出しを行うことができます。

```python
import osquery

# Spawn an osquery process using an ephemeral extension socket.
instance = osquery.SpawnInstance()
instance.open()  # This may raise an exception

# Issues queries and call osquery Thrift APIs.
instance.client.query("select timestamp from time")
```

queryメソッドは、クエリの結果を含むExtensionResponseオブジェクトを返します。たとえば、次のようになります。

```python
ExtensionResponse(status=ExtensionStatus(code=0, message='OK', uuid=0), response=[{'timestamp': 'Fri May 30 17:54:06 2025 UTC'}])
```

osqueryを初めて使用する場合は、[スキーマ](https://www.osquery.io/schema/5.17.0/)を見て、オペレーティングシステムでどのような情報が利用できるかを確認することをお勧めします。

### フォーマットに関する補足

前の例の出力はすべてフォーマットされていませんでしたが、Jupyterからコードを実行している場合は、次のパッケージをインポートすることで、出力を美しくするための便利なメソッドにアクセスできます。

```python
from IPython.display import Markdown, display
```

そして、応答出力をマークダウンとして表示します。

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

## 点と点をつなぐ

オペレーティングシステムに関する情報をクエリする方法がわかったので、それをエージェントの知識と組み合わせて、システムに関する質問に答える診断エージェントを作成しましょう。

最初のステップは、クエリを実行する関数を定義することです。これは、後で情報を収集するためのツールとしてエージェントに与えられます。

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

関数自体は非常に些細なものですが、ここで重要なのは、エージェントがこの関数がどのように機能するかを理解できるように、非常に詳細なdocstringを用意することです。

テスト中に頻繁に発生した厄介な問題の1つは、エージェントがシステムでどのテーブルが利用できるかを正確に知らなかったことです。たとえば、私はmacOSマシンを実行しており、「memory_info」テーブルは存在しません。

エージェントにもう少しコンテキストを与えるために、このシステムで利用可能なテーブルの名前を動的に与えます。理想的な状況では、列名と説明を含むスキーマ全体を与えることさえできますが、残念ながら、osqueryでこれを実現するのは簡単ではありません。

osqueryの基盤となるデータベーステクノロジーはSQLiteなので、`sqlite_temp_master`テーブルから仮想テーブルのリストをクエリできます。

```python
# use some python magic to figure out which tables we have in this system
response = instance.client.query("select name from sqlite_temp_master").response
tables = [ t["name"] for t in response ]
```

すべてのテーブル名がわかったので、この情報と`call_osquery`ツールを使用してエージェントを作成できます。

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

信じられないかもしれませんが、私たちのエージェントはすべての質問に対応できます！試してみましょう。

```python
response = osagent.query(input="what is the current time?")
display(Markdown(response["output"]))
```

出力：

```
The current time is Fri May 30 18:08:15 2025 UTC.
```

もう少し複雑にしてみましょう。

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

もう少しクリエイティブになったらどうでしょう？

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

すごい！レベル1の診断手順が何であるかを実際に定義したことはありませんが、それでもかなり印象的なレポートをくれました！

質問にもう少し巧妙になろうとしましたが、答えは（ほとんどの場合）期待を裏切りませんでした。

```python
response = osagent.query(input="computer, do you see any signs of malware running?")
display(Markdown(response["output"]))
```

出力：

```
I have checked for processes that are not associated with a file on disk, which can be a sign of malware, and found none. I have also examined the top processes by memory and CPU usage. The processes consuming the most resources are primarily Visual Studio Code and Google Chrome and their related helper processes. This is typical behavior for these applications.

Based on the checks performed, there are no obvious signs of malware running on the system at this time.
```

_マイクドロップ_ =^.^=

## 結論

今では使い古された議論であることは承知していますが、AIはゲームチェンジャーです。ほんの数行のコードで、オペレーティングシステムの内部動作に対する完全に機能する自然言語インターフェイスをゼロから作成しました。もう少し作業すれば、このエージェントを改善して、より詳細な診断を行い、さらには自律的に物事を修正することもできます。スコッティも誇りに思うでしょう！

![エンジニアのスコッティがマウスをマイクとして使ってコンピューターと話そうとしている](hello-computer-hello.gif)

この記事のすべての例のソースコードは、私の[GitHub](https://github.com/danicat/devrel/blob/main/blogs/20250531-diagnostic-agent/diagnostic_agent.ipynb)にあります。

あなたの感想はどうですか？以下のコメントであなたの考えを共有してください。
