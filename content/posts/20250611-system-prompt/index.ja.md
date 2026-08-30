---
categories:
- Agent Development
date: '2025-06-11T00:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 3
summary: システム指示（System Instructions）、セッション履歴、エージェントツールの概念を学び、よりインテリジェントで自律的な診断アシスタントを構築します。
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: "大胆なプロンプト：システム指示とエージェントツールの実践ガイド"
slug: "system-instructions-agent-tools"
aliases:
  - "/ja/posts/20250611-system-prompt/"
description: "システム指示、複数ターンのチャット履歴、osqueryエージェントツールを組み合わせ、Vertex AIで賢い自律型診断アシスタントを構築する実践ガイド。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-cloud-aiplatform"
  - "osquery"
  - "LangChain"
---
## はじめに

本ガイドでは、システムプロンプト（System Prompt / システム指示）とエージェントツール（Agent Tools）の設計を深掘りし、さらに洗練された診断エージェント体験を構築していきます。今回は [Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog)、LangChain、Gemini、そして [Osquery](https://www.osquery.io/) を組み合わせて実装を進めます。

正直なところ、[シリーズ第1弾で作成した初期バージョンの診断エージェント]({{< ref "/posts/20250531-diagnostic-agent" >}}) はお世辞にもエンタープライズ（Enterprise）レディとは言えませんでした（もちろんエンタープライズ号の掛け言葉です）。内部で実際に何が行われているのか（本当にクエリを実行しているのか？）可視化できず、同じ「セッション」内で話した内容を記憶できず、時折ユーザーの指示を完全に無視することさえありました。

これでは本格的なエージェントに期待するユーザー体験とは程遠い状態です。理想的な診断エージェントは、自分の失敗を記憶し、一貫して指示を実行できる必要があります（例えば、特定のカラムが存在しないことを学習して別のアプローチをとるなど）。さらに、「言っている通りの処理を本当に実行しているのか？」という疑問も残ります。返される情報が正確かつ最新であることを確認するために、実行中のクエリをいつでも確認できるようにすべきです。

これらの課題を念頭に置きながら、さっそく手を動かして緊急用~~医療ホログラム~~診断エージェント（Emergency Diagnostic Agent）を作っていきましょう！

## 舞台を整える

前回は手軽さを優先して Jupyter Notebook 上でコードを書きましたが、今回は標準的な Python プログラムとして作成します。わずかな変更で Jupyter 上でも動作しますが、今回は診断エージェントに本格的なチャットインターフェースを持たせるために通常のスクリプトとして構築します。

Python プロジェクトを始める際は、依存関係を独立して管理するためにクリーンな仮想環境を作成することをおすすめします：

```bash
$ mkdir -p ~/projects/diagnostic-agent
$ cd ~/projects/diagnostic-agent
$ python3 -m venv venv
$ source venv/bin/activate
$ pip install --upgrade google-cloud-aiplatform[agent_engines,langchain]
```

前回の記事のエージェントを再現した初期バージョンの `main.py` は以下の通りです：

```py
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown
import os

PROJECT_ID = os.environ.get("GCP_PROJECT")
LOCATION = os.environ.get("GCP_REGION", "us-central1")
STAGING_BUCKET = os.environ.get("STAGING_BUCKET_URI")

vertexai.init(
    project=PROJECT_ID,
    location=LOCATION,
    staging_bucket=STAGING_BUCKET
)

MODEL = os.environ.get("GEMINI_MODEL", "gemini-2.0-flash")

instance = osquery.SpawnInstance()

def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()  # This may raise an exception

    result = instance.client.query(query)
    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()  # This may raise an exception
    
    response = instance.client.query("select name from sqlite_temp_master").response
    tables = [ t["name"] for t in response ]
    return f"""
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
    """

def main():
    agent = agent_engines.LangchainAgent(
        model = MODEL,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
        ],
    )
    
    console = Console()

    print("Welcome to the Emergency Diagnostic Agent\n")
    print("What is the nature of your diagnostic emergency?")

    while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query)
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)

    print("Goodbye!")

if __name__ == "__main__":
    main()
```

エージェントは `python main.py` で実行できます：

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

What is the nature of your diagnostic emergency?
>> 
```

元のコードと比較して、2つの細かな変更点があります。1点目は、ユーザーが「exit」または「quit」と入力するまでエージェントを動かし続けるメインループを追加した点です。これでチャットインターフェースが完成します。

2点目は、エージェントの一貫性を高めるためにシステムプロンプトを調整した点です。ここではエージェントの名前を「Emergency Diagnostic Agent（緊急診断エージェント）」と名付けました。この名前は素晴らしい[『スタートレック』のイースターエッグ](https://en.wikipedia.org/wiki/The_Doctor_(Star_Trek:_Voyager))（緊急用医療ホログラム）として機能するだけでなく、より重要な点として、緊急性のあるトーンを設定することで、[近年の研究論文](https://arxiv.org/pdf/2307.11760)で示されているようにモデルがリクエストにより忠実に従うよう促す効果が期待できます（こちらの[最近のインタビュー](https://www.reddit.com/r/singularity/comments/1kv7hm2/sergey_brin_we_dont_circulate_this_too_much_in/)も併せてご覧ください）。

もちろん哀れな緊急診断エージェントを脅迫するつもりはありませんし、本稿の作成にあたって傷つけられたエージェントは一切存在しないことを保証します。とはいえ、「緊急用（Emergency）」と銘打つことで、モデルが全力を尽くしてこちらの要求に応えようとする雰囲気を作り出すことができます。前バージョンのシステムプロンプトでは、エージェントが自分には能力がないと「思い込んだり」、どのテーブルを照会すればよいかわからずタスクを拒否するケースがありました。

もちろん、単に緊急エージェントと呼ぶだけでは期待通りの振る舞いを保証するのに不十分ですので、以下で見るようにモデルの行動を導くための指示をいくつか追加しています。

## システム指示（System Instructions）

システムプロンプト（システム指示）とは、会話全体を通じて LLM の振る舞いをガイドする一連の指示のことです。通常のチャット対話よりも高い優先度を持つ特別な指示であり、モデルに送信するすべてのプロンプトに常にシステム指示が添えて繰り返されているようなイメージです。

システムプロンプトがどうあるべきかについて、文献上の確固たる統一見解はまだありませんが、日々の実践からいくつかの実績あるパターンが生まれてきています。例えば、ほぼ共通の認識となっているのが、プロンプトの冒頭でエージェントに「役割（Role）」を割り当て、自身の目的を認識させて首尾一貫した回答を生成させるという手法です。

今回のエージェントでは、システムプロンプトに **Role（役割）**、**Tools（ツール）**、**Context（コンテキスト）**、**Task（タスク）** の4つのセクションを含める構成を採用しました。この構造はテスト段階で非常にうまく機能しましたが、これだけに固執する必要はありません。プロンプトを色々試して、より良い結果が得られるか実験してみてください。LLM で優れた成果を上げるには、実験を重ねることが何より重要です。

それでは、プロンプトの各セクションを順に見ていきましょう。

### システムプロンプト：Role（役割）

役割とは、エージェントが存在する理由そのものです。「あなたはソフトウェアエンジニアです」や「あなたは診断エージェントです」といったシンプルなものから、詳細な説明、行動規範、制約事項などを含む手の込んだものまで様々です。

あらゆる種類のデータで訓練された大規模言語モデルにとって、役割を設定することは、質問に答えるためにアクセスすべき専門知識のドメインを定めるのに役立ちます。言い換えれば、質問に意味的（セマンティック）な文脈を与えるのです。例えば「クッキー（cookies）とは何ですか？」という質問を想像してみてください。食べるクッキーのことでしょうか、それともブラウザの Cookie のことでしょうか？ エージェントの役割が未定義であればこの質問は完全に曖昧ですが、役割を「あなたはソフトウェアエンジニアです」といった技術的なものに設定した瞬間に、その曖昧さは解消されます。

今回のエージェントでは、役割を以下のように記述しました：

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
```

単刀直入な定義（「あなたは緊急診断エージェントです」）に加え、モデルの振る舞いのトーンを決定づけ、こちらの要求を「真剣に」受け止めるよう促すための詳細な説明を追加しました。前述の通り、以前のバージョンのエージェントには要求を拒否してしまう悪い癖があったためです。

### システムプロンプト：Tools（ツール）

Tools は、コアモデルの枠を超えて外部システムと対話する能力をエージェントに説明するセクションです。ツールにはいくつかの種類がありますが、最も一般的な提供方法は Function Calling（関数呼び出し）です。

エージェントはツールを使って情報を取得し、タスクを実行し、データを操作できます。Vertex AI SDK for Python は、ユーザー定義関数だけでなく、Google 検索やコード実行などの組み込みツールもサポートしています。さらに、Model Context Protocol（MCP）インターフェースを介してコミュニティが管理する拡張機能を利用することも可能です。

今回構築するエージェントには、_osquery_ を呼び出せることを伝えます：

```
Tools:
  - you can call osquery using the call_osquery function.
```

### システムプロンプト：Context（コンテキスト）

次は Context です。エージェントが動作する環境についての情報を伝えます。私はこのセクションを使って、以前のエージェントが陥りがちだった好ましくない挙動を明示的に指摘し、修正するようにしています。例えば開発のごく初期段階で、エージェントが存在するテーブルを「推測」して闇雲にクエリを送信し、エラー率が跳ね上がることがありました。コンテキストに利用可能なテーブル一覧を追加したことで、その問題は大幅に軽減されました。

同様に、テーブル内のカラム名を事前に確認せず推測しようとする傾向もありました。このケースでは、常に `SELECT *` を使うよう指示したくなる誘惑に駆られましたが、それは悪手（必要以上のデータを取得してしまう）なので思いとどまり、代わりに `PRAGMA` 文を使ってスキーマを探索する方法を「教え」ました。

これにより、エージェントがカラム名を推測して失敗することはあっても、人間の手を煩わせることなく自律的に軌道修正できるようになります。

改善されたシステムプロンプトの Context セクションは以下の通りです：

```
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
```

ここで `{tables}` は、モデルを起動する前に Osquery から取得したすべてのテーブル一覧を格納した変数です。

### システムプロンプト：Task（タスク）

最後は Task です。このセクションでは、エージェントがユーザーのリクエストをどのように解釈し、実行すべきかを記述します。通常は、目の前のタスクを達成するために必要なステップを整理するために使用されます。

今回のケースでは、大まかな実行計画を提示しつつ、いくつかの条件付きディレクティブを追加しています：

```
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

「実行前にユーザーに計画を確認する（Confirm the plan with the user before executing）」というステップは、エージェントがプロセスについてどう考えているかが把握できて興味深い反面、しばらくやり取りを続けていると少し煩わしく感じるかもしれません。実行計画はプロンプトで尋ねればいつでも教えてもらえるため、このステップを含めるかどうかは完全に任意です。

当初はこのステップをエージェントのデバッグ用として考えていましたが、次のセクションでは別の方法でデバッグを実現していきます。

これら4つのセクションを組み合わせることで、完全なシステムプロンプトが出来上がります。この改善されたプロンプトにより、本稿の執筆に向けたテスト期間中、より一貫した結果が得られるようになりました。また「人間にとって読みやすい」構造になっているため、新しいルールを追加する際のメンテナンスも容易です。

完成したシステムプロンプトの全体像は以下の通りです：

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

余談ですが、プロンプトにはまだまだ改善の余地があると考えています。現在私が特に関心を持っている分野の1つが「自己改善型システムプロンプト（Self-improving System Prompt）」の実現です。セッション終了時に、モデル自身が得た学びを要約させて次回イテレーション用の新しいシステムプロンプトを生成させる、といったアプローチが考えられます。プロンプトをデータベースに保存し、次回のセッションで読み込む形です。もちろん、これにはシステムプロンプトの劣化（Degradation）や、さらに深刻なプロンプトインジェクション攻撃への懸念が生じるため、一筋縄ではいきません。それでも非常に面白い試みであり、近いうちに記事にするかもしれません。

## デバッグモードを有効にする

初期設計におけるもう1つの懸念事項は、エージェントが水面下で何を行っているのかについての可観測性（オブザーバビリティ）の欠如でした。ここには2つのアプローチがあり、片方はもう一方よりも少し厄介です。1つ目は、LLM の「思考（Thoughts）」を覗き見してその中からツール呼び出しを探し出す方法（非常に面倒）、2つ目は、関数自体にデバッグ機能を追加して実行時に必要な情報を出力させる方法（大抵の場合、最もシンプルな解決策こそが正解です）です。

白状すると、私はアプローチ2に気づくまで、アプローチ1に不健全なほどの時間を費やしてしまいました。どうしても LLM の推論プロセスを追いたい場合は、[return_intermediate_steps](https://api.python.langchain.com/en/latest/agents/langchain.agents.agent.AgentExecutor.html#langchain.agents.agent.AgentExecutor.return_intermediate_steps) という設定を利用できます。学習の観点からは非常に興味深いのですが、出力フォーマットの解析に数時間費やした結果（ヒント：[実際には JSON ではありません](https://github.com/langchain-ai/langchain/issues/10099)）、わざわざパースするほどの価値はないと判断しました。

では、シンプルな戦略はどのように機能するのでしょうか？ デバッグフラグと、そのフラグをオン・オフするためのツールを追加します。この驚くほどシンプルな工夫によって、まったく新しい可能性の扉が開かれます。つまり、エージェントに自身の振る舞いを変更する権限を与えるのです！

デバッグモードの実装は、グローバル変数とそれを設定する関数で構成されます：

```py
debug = False

def set_debug_mode(debug_mode: bool):
    """Toggle debug mode. Call this function to enable or disable debug mode.
    
    Args:
        debug_mode (bool): True to enable debug mode, False to disable it.


    Returns:
        None
    """
    global debug
    debug = debug_mode
```

システムプロンプトにもこのツールについて記載します：

```
...
   Tools:
    - you can call osquery using the call_osquery function.
    - you can enable or disable the debug mode using the set_debug_mode function.
    Context:
...
```

そして、エージェントのインスタンス化時にツールのリストに関数を追加します：

```py
   agent = agent_engines.LangchainAgent(
        model = model,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode,
        ],
    )
```

最後に、新しい `debug` フラグを使用するように `call_osquery` を修正します：

```py
def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()

    if debug:
        print("Executing query: ", query)

    result = instance.client.query(query)
    if debug:
        print("Query result: ", {
            "status": result.status.message if result.status else None, 
            "response": result.response if result.response else None
        })

    return result
```

これらの変更をすべて反映した上で、新しく実装したデバッグフラグを使ってエージェントがどのように _osquery_ を呼び出すか確認してみましょう：

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

What is the nature of your diagnostic emergency?
>> run a level 1 diagnostic procedure in debug mode
Executing query:  SELECT * FROM system_info
Query result:  {'status': 'OK', 'response': [{...}]}
Executing query:  SELECT pid, name, user, cpu_percent FROM processes ORDER BY cpu_percent DESC LIMIT 10
Query result:  {'status': 'no such column: user', 'response': None}
Executing query:  SELECT pid, name, user, resident_size FROM processes ORDER BY resident_size DESC LIMIT 10
Query result:  {'status': 'no such column: user', 'response': None}
Executing query:  PRAGMA table_info(processes)
Query result:  {'status': 'OK', 'response': [{'cid': '0', 'dflt_value': '', 'name': 'pid', 'notnull': '1', 'pk': '1', 'type': 'BIGINT'}, ...]}
(...)

System Information:                                                                                                                                                     

 • Hostname: workstation.localdomain                                                                                                                               
 • CPU Type: arm64e                                                                                                                                                     
 • Physical Memory: 51539607552 bytes                                                                                                                                   

Top 5 Processes by CPU Usage:                                                                                                                                          
                                                  
  PID     Name                         CPU Usage  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 
  1127    mediaanalysisd               95627517   
  43062   mediaanalysisd-access        66441942   
  54099   Google Chrome                3005046    
  54115   Google Chrome Helper (GPU)   2092500    
  81270   Electron                     1688335    

Top 5 Processes by Memory Usage (Resident Size):                                                                                                                       
                                                                   
  PID     Name                              Resident Size (Bytes)  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 
  43062   mediaanalysisd-access             3933536256             
  54099   Google Chrome                     1313669120             
  59194   Code Helper (Plugin)              1109508096             
  59025   Code Helper (Renderer)            915456000              
  19681   Google Chrome Helper (Renderer)   736329728                         
                                                                   
>> 
```

入力したコマンドが「run a level 1 diagnostic procedure **in debug mode**（デバッグモードでレベル1診断手順を実行せよ）」である点に注目してください。これはエージェントの興味深い能力である「マルチツール呼び出し（Multi-tool invocation）」を示しています。エージェントは必要と判断すれば、同じツールを複数回呼び出すだけでなく、異なるツールを同時に呼び出すこともできます。そのため、レポートを要求する前にあらかじめデバッグモードを有効にしておく必要はなく、エージェントが一度の処理ですべて実行してくれました。

また、エージェントが最初に `user` カラムを要求して失敗した後、`PRAGMA` 文を使って正しいスキーマを探索し、クエリを再試行して成功させている点にもご注目ください。これは、改善されたシステムプロンプトによってエージェントがエラーから自律的に復旧できる能力を完璧に証明しています。

## チャット履歴の保持

本日の最後の課題は、調査の一貫した流れに沿って明確化のための質問をしたりシステムをさらに掘り下げて調査できるよう、エージェントが会話内容を確実に記憶できるようにすることです。

[前回の記事]({{< ref "/posts/20250605-vertex-ai-sdk-python" >}}) では、LLM がステートレス（状態を持たない）であること、そして「ターン（Turns）」を使って会話の現在の状態をモデルに「思い出させ」続ける必要があることを説明しました。幸いなことに、LangChain を使えばこれを手作業で行う必要はなく、[チャット履歴（Chat History）](https://python.langchain.com/api_reference/core/chat_history.html) という機能を活用できます。

チャット履歴の素晴らしいところは、[BaseChatMessageHistory](https://python.langchain.com/api_reference/core/chat_history/langchain_core.chat_history.BaseChatMessageHistory.html#langchain_core.chat_history.BaseChatMessageHistory) を実装しているものであれば何でも利用できる点です。自作のものを含め、あらゆる種類のデータストアを利用できます。例えば Vertex AI の公式ドキュメントには、[Firebase、Bigtable、Spanner](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain#chat-history?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog) を使用する例が掲載されています。

現時点では本格的なデータベースは必要ないため、名前の通りすべてをメモリ上に保存する `InMemoryChatMessageHistory` を採用します。

以下は典型的な実装例であり、ルックアップ用の `chats_by_session_id` 辞書を使用して複数のセッションをサポートしています（コードは [LangChain 公式ドキュメント](https://python.langchain.com/docs/versions/migrating_memory/chat_history/#chatmessagehistory) より引用）：

```py
chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history
```

そして、チャット履歴を有効にしてエージェントをインスタンス化する新しい `main` 関数は以下のようになります：

```py
import uuid

def main():
    session_id = uuid.uuid4()
    agent = agent_engines.LangchainAgent(
        model = model,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode,
        ],
        chat_history=get_chat_history,
    )
```

私と同じ過ちを犯さないための注意点として、`chat_history` 引数は `Callable` 型を想定しているため、ここで関数を実行（invoke）するのではなく、関数そのものを渡す必要があります。LangChain はここでファクトリパターンを採用しており、必要に応じて提供された関数（`get_chat_history`）を `session_id` 付きで呼び出して、適切な履歴オブジェクトを取得または作成します。この設計により、エージェントは複数の独立した会話を並行して管理できるようになっています。

関数のシグネチャには1つまたは2つの引数を含めることができます。引数が1つの場合は `session_id` とみなされ、2つの場合は `user_id` と `conversation_id` として解釈されます。詳細については [RunnableWithMessageHistory](https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.history.RunnableWithMessageHistory.html) のドキュメントを参照してください。

最後のピースは、モデルの実行処理に `session_id` を渡すことです。これは以下のコードのように `config` 引数を介して行います：

```py
# (...)
   while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query, config={"configurable": {"session_id": session_id}})
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)
```

これで、セッションがアクティブである限り、会話内容はメモリ上に保存されるため、エージェントの「短期記憶」にある情報について質問できるようになります。これにより基本的な対話の多くがより自然に感じられるようになりますが、同時にさらに大きな問題への先例を作ることにもなります。セッション情報を保存できるようになったことで、イテレーションを重ねるごとに履歴は肥大化していきます。さらにクエリから自動生成されたデータを扱っていると、セッションコンテキストは急速に増大し、コンピューターのメモリ上限に達するはるか前に、モデルのトークン制限にぶつかってしまいます。

Gemini のようなモデルは[ロングコンテキストウィンドウ](https://ai.google.dev/gemini-api/docs/long-context)で知られていますが、コンテキストをデータで埋め尽くしてしまえば、100万トークンであってもあっという間に使い果たされてしまいます。また、コンテキストが長くなると情報の検索が次第に困難になるという問題（いわゆる[干し草の中の針（Needle in a Haystack）](https://cloud.google.com/blog/products/ai-machine-learning/the-needle-in-the-haystack-test-and-how-gemini-pro-solves-it?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog) 問題）が発生するモデルもあります。

増大するコンテキスト問題に対処する手法としては、圧縮（Compression）や要約（Summarisation）などがありますが、本稿のコンテキストを短く保つためにも（お分かりですね？）、それらは次回の記事に取っておくことにしましょう。

本稿で行ったすべての修正を含めた `main.py` の最終バージョンは以下のようになります：

```py
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown
from langchain_core.chat_history import InMemoryChatMessageHistory
import os
import uuid

PROJECT_ID = os.environ.get("GCP_PROJECT")
LOCATION = os.environ.get("GCP_REGION", "us-central1")
STAGING_BUCKET = os.environ.get("STAGING_BUCKET_URI")

vertexai.init(
    project=PROJECT_ID,
    location=LOCATION,
    staging_bucket=STAGING_BUCKET
)

MODEL = os.environ.get("GEMINI_MODEL", "gemini-2.5-pro-preview-05-06")

instance = osquery.SpawnInstance()
debug = False

def set_debug_mode(debug_mode: bool):
    """Toggle debug mode. Call this function to enable or disable debug mode.
    
    Args:
        debug_mode (bool): True to enable debug mode, False to disable it.


    Returns:
        None
    """
    global debug
    debug = debug_mode

def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()  # This may raise an exception

    if debug:
        print("Executing query: ", query)

    result = instance.client.query(query)
    if debug:
        print("Query result: ", {
            "status": result.status.message if result.status else None, 
            "response": result.response if result.response else None
        })

    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()  # This may raise an exception
    
    response = instance.client.query("select name from sqlite_temp_master").response
    tables = [ t["name"] for t in response ]
    return f"""
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
  - you can use the set_debug_mode function to enable or disable debug mode.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
    """

chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history

def main():
    session_id = uuid.uuid4()
    agent = agent_engines.LangchainAgent(
        model = MODEL,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode
        ],
        chat_history=get_chat_history,
    )
    
    console = Console()

    print("Welcome to the Emergency Diagnostic Agent\n")
    print("What is the nature of your diagnostic emergency?")

    while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query, config={"configurable": {"session_id": session_id}})
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)

    print("Goodbye!")

if __name__ == "__main__":
    main()
```

## まとめ

この記事では、エージェントから一貫した応答を引き出すためのシステムプロンプト微調整の重要性を学びました。また、マルチツール呼び出しが実際にどのように機能するか、ツールを使用してフィーチャーフラグを切り替えエージェントの挙動を動的に変更する方法も実践しました。さらに、インメモリのチャット履歴を使用してセッション状態を管理する方法についても解説しました。

本シリーズの次回以降の記事では、本物のデータベースを使用してセッション間の永続化を実現する方法を見ていき、トークンの概念を再確認しながらコンテキスト圧縮テクニックについて議論します。

## 付録：試してみたい面白いプロンプト

エージェントがより堅牢になったところで、新機能を実際にテストするための実践ガイドとしてこのセクションを活用してください。質問間でコンテキストがどのように記憶されているか、またエージェント自身に行動の理由を説明させることができるかを確認してみてください。

```
>> run a level 1 diagnostic procedure
>> run a level 2 diagnostic procedure
>> explain the previous procedure step by step
>> find any orphan processes
>> show me the top resource consuming processes
>> write a system prompt to transfer your current knowledge to another agent
>> search the system for malware
>> is this computer connected to the internet?
>> why is my computer slow?
>> take a snapshot of the current performance metrics
>> compare the current perfomance metrics with the previous snapshot
>> give me a step by step process to fix the issues you found
>> how many osqueryd processes are in memory?
>> give me a script to kill all osqueryd processes
>> who am i?
```

他にも面白いプロンプトを見つけたら、ぜひ下のコメント欄であなたの体験を共有してください。

シリーズの次回作 [Agent Development Kit (ADK) で診断エージェントを作成する]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}) では、素の SDK によるボイラープレートから Google の Agent Development Kit (ADK) へのリファクタリングを行い、動的なテーブルスキーマ探索のための Vertex AI RAG を統合します。それでは、また次回お会いしましょう！
