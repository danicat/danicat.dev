+++
date = '2025-06-11T00:00:00+01:00'
title = '大胆なプロンプト：システム指示とエージェントツールの実践ガイド'
summary = "この記事では、システム指示、セッション履歴、エージェントツールの概念を探求し、よりスマートな診断アシスタントを作成します。"
categories: ["AI & Development"]
tags: ["gemini", "vertex-ai", "python", "tutorial"]
+++
{{< translation-notice >}}

## はじめに

このガイドでは、システムプロンプトとエージェントツールについて詳しく学び、新しく改良された診断エージェントエクスペリエンスを構築します。[Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog)、LangChain、Gemini、[osquery](https://www.osquery.io/)を使用します。

[診断エージェントの初期バージョン](https://danicat.dev/posts/20250531-diagnostic-agent/)は、あまり「エンタープライズ」対応ではなかったことを認めなければなりません（しゃれです）。内部で何をしているのか（実際にクエリを実行しているのか？）についてはあまり可視性がなく、同じ「セッション」で議論されたことを覚えておらず、時々、私たちのコマンドを完全に無視することもありました。

これは、適切なエージェントに期待するエクスペリエンスとはかけ離れています。理想的な診断エージェントは、自分の間違いを覚えて、指示を一貫して実行できる必要があります。たとえば、特定の列が利用できないことを学習し、それを回避するなどです。また、それが言っていることを本当に実行していると信頼できるでしょうか？返される情報が正しく、最新であることを確認するために、いつでもクエリを確認できる必要があります。

これらの目標を念頭に置いて、手を汚して、緊急~~医療ホログラム~~診断エージェントの構築を始めましょう！

## 準備

前回は便宜上Jupyterノートブックでコードを作成しましたが、今回は通常のPythonプログラムを作成します。同じコードはJupyterでも最小限の変更で動作しますが、今回は適切なチャットインターフェイスで診断エージェントを使用できるようにするためにこれを行います。

Pythonプロジェクトでは、依存関係を自己完結させるために、常にクリーンな仮想環境から始めることをお勧めします。

```
$ mkdir -p ~/projects/diagnostic-agent
$ cd ~/projects/diagnostic-agent
$ python3 -m venv venv
$ source venv/bin/activate
$ pip install --upgrade google-cloud-aiplatform[agent_engines,langchain]
```

以下は、前の記事のエージェントを再現する`main.py`の初期バージョンです。

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

`python main.py`でエージェントを実行できます。

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

What is the nature of your diagnostic emergency?
>> 
```

元のコードと比較して2つの小さな変更があります。まず、ユーザーが「exit」または「quit」と入力するまでエージェントを実行し続けるメインループができました。これにより、チャットインターフェイスが作成されます。

次に、エージェントの一貫性を向上させるためにシステムプロンプトを調整しました。現在は「緊急診断エージェント」と呼んでいます。この名前は、気の利いた[スタートレックのイースターエッグ](https://en.wikipedia.org/wiki/The_Doctor_(Star_Trek:_Voyager))として機能するだけでなく、より重要なことに、[新たな研究](https://arxiv.org/pdf/2307.11760)に基づいて、私たちの要求により熱心に従うように促す緊急性のトーンを設定します。（この[最近のインタビュー](https://www.reddit.com/r/singularity/comments/1kv7hm2/sergey_brin_we_dont_circulate_this_too_much_in/)もチェックしてください）

私たちは貧しい緊急診断エージェントを脅かすつもりはありません。そして、このテキストの作成中にエージェントが傷つけられなかったことを保証できます。しかし、「緊急」エージェントと呼ぶことで、私たちの要求に最大限の能力で応えようとするトーンが設定されるはずです。システムプロンプトの前のバージョンでは、エージェントがタスクを実行できないと「考えた」ため、またはどのテーブルをクエリすればよいかわからなかったためにタスクを拒否するケースがありました。

もちろん、緊急エージェントと呼ぶだけでは、望ましい動作を保証するのに十分ではないため、以下で説明するように、モデルの動作をガイドするためにいくつかの指示を追加しました。

## システム指示

システム指示（システムプロンプトとも呼ばれる）は、会話全体を通してLLMの動作をガイドする一連の指示です。システム指示は、通常のチャットインタラクションよりも優先度が高いため特別です。これを想像する方法の1つは、システム指示がモデルに送信しているプロンプトと一緒に常に繰り返されているかのように考えることです。

システムプロンプトがどのように見えるべきかについては、文献で強いコンセンサスはありませんが、日常的な使用からいくつかの実証済みのパターンが出現しています。たとえば、事実上コンセンサスとなっているのは、システムプロンプトの冒頭でエージェントに役割を割り当てることです。これにより、エージェントは自分の目的を認識し、より一貫性のある回答を生成できます。

この特定のエージェントでは、システムプロンプトに次のセクションを含めることにしました。役割、ツール、コンテキスト、タスク。この構造は、テスト段階でうまく機能しましたが、あまり固執しないでください。プロンプトを試して、より良い結果が得られるかどうかを確認してください。実験は、LLMで良い結果を得るための鍵です。

それでは、プロンプトの各セクションを見ていきましょう。

### システムプロンプト：役割

役割は、エージェントが存在する理由に他なりません。「あなたはソフトウェアエンジニアです」や「あなたは診断エージェントです」のように単純な場合もあれば、詳細な説明、行動ルール、制限などを含む、より精巧な場合もあります。

あらゆる種類のデータでトレーニングされた大規模言語モデルにとって、役割は、クエリに答えるためにアクセスする必要のあるドメイン知識のトーンを設定するのに役立ちます。言い換えれば、クエリに意味的な意味を与えます…たとえば、「クッキーとは何ですか？」という質問を想像してみてください。食べられるクッキーについて話しているのでしょうか、それともブラウザのクッキーについて話しているのでしょうか？エージェントの役割が未定義の場合、この質問は完全に曖昧ですが、役割を技術的なもの（たとえば、「あなたはソフトウェアエンジニアです」）に設定すると、曖昧さがなくなります。

このエージェントの場合、役割は次のように記述されます。

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
```

単純な定義（「あなたは緊急診断エージェントです」）を超えて、モデルの動作のトーンを設定し、うまくいけば私たちの要求を「真剣に」受け止めるように影響を与えるために、より長い説明を追加しました。前述のように、このエージェントの前のイテレーションでは、要求を拒否する悪い傾向がありました。

### システムプロンプト：ツール

ツールは、エージェントがコアモデルを超えて外部システムと対話する能力を説明するセクションです。ツールにはいくつかの種類がありますが、エージェントにツールを提供する最も一般的な方法は、関数呼び出しを介することです。

エージェントは、ツールを使用して情報を取得したり、タスクを実行したり、データを操作したりできます。Vertex AI SDK for Pythonは、ユーザー提供の関数と、Google検索やコード実行などの組み込みツールの両方をサポートしています。また、Model Context Protocol（MCP）インターフェイスを介して、コミュニティが維持している拡張機能を使用することもできます。

私たちのエージェントには、_osquery_を呼び出すことができることを伝える必要があります。

```
Tools:
  - you can call osquery using the call_osquery function.
```

### システムプロンプト：コンテキスト

次に、エージェントが動作する環境についてエージェントに伝えるコンテキストがあります。このセクションを使用して、エージェントの前のイテレーションが起こしやすかった望ましくない動作を明示的に呼び出して修正します。たとえば、開発の非常に早い段階で、エージェントが利用可能なテーブルを「推測」しようとして、盲目的にクエリを送信し、高いエラー率になることに気づきました。コンテキストにテーブルのリストを追加することで、その問題が軽減されました。

同様に、エージェントが最初に名前を発見しようとするのではなく、テーブル内の列名を推測しようとする傾向があります。この特定のケースでは、エージェントに常にSELECT *を使用するように指示する誘惑に抵抗しました。これは悪い習慣だからです（必要以上のデータを取得します）。代わりに、PRAGMA命令を使用してスキーマを発見する方法を「教え」ました。

これにより、エージェントは列名を推測する際にまだ間違いを犯しますが、人間の介入なしにコースを修正する方法があります。

システムプロンプトの改訂されたコンテキストセクションを以下に示します。

```
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
```

テーブルは、モデルを開始する前にosqueryから発見したすべてのテーブルを含む変数であることに注意してください。

### システムプロンプト：タスク

最後に、タスクです。このセクションは、エージェントがリクエストをどのように解釈し、実行するかを記述するために使用されます。人々は通常、このセクションを使用して、手元のタスクを達成するために必要な手順をレイアウトします。

私たちの特定のケースでは、このセクションを使用して計画を大まかにレイアウトするだけでなく、いくつかの条件付きディレクティブも追加しています。

```
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

「実行する前にユーザーと計画を確認する」というステップは、エージェントがプロセスについてどのように考えているかを示してくれるため興味深いですが、しばらくエージェントと対話した後は少し面倒かもしれません。プロンプトでエージェントに計画を教えてもらうようにいつでも頼むことができるので、このステップを含めることは完全にオプションです。

私は当初、このステップをエージェントをデバッグする方法として考えていましたが、次のセクションでは、それを行う別の方法を探ります。

これら4つのセクションを組み合わせることで、システムプロンプト全体が完成します。この改訂されたプロンプトは、この記事の準備中のテストでより一貫した結果を生み出しました。また、「人間にも優しい」という利点もあり、新しいルールが導入されたときに適応しやすくなっています。

このシステムプロンプトの完全なビューは次のとおりです。

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

余談ですが、まだ改善の機会はたくさんあると信じており、現在の関心分野の1つは、自己改善するシステムプロンプトを実現する方法です。これは、セッションの最後に、モデルにその学習内容を将来のイテレーションのための新しいシステムプロンプトに要約するように依頼することで、潜在的に達成できる可能性があります。プロンプトはデータベースに保存され、次のセッションでロードできます。もちろん、これはシステムプロンプトの劣化や、さらに悪いことに、プロンプトインジェクションを使用した攻撃に対する懸念を引き起こすため、見た目ほど些細なことではありません。それにもかかわらず、それは楽しい演習であり、近い将来それについて書くかもしれません。

## デバッグモードの有効化

元の設計に関するもう1つの懸念は、エージェントが内部で何をしているかについての可観測性の欠如です。ここで適用できる2つの異なるアプローチがあります。1つはもう1つよりも少し面倒です。1）LLMの「思考」を覗き込み、その中からツール呼び出しを見つけようとします（非常に面倒です）。または、2）関数自体にデバッグ機能を追加して、実行時に必要な情報を出力するようにします（最も簡単な解決策は通常、正しいものです）。

オプション2ができることに気づく前に、オプション1で不健康な量の時間を費やしたことを認めなければなりません。LLMの推論の道を本当に進みたい場合は、[return_intermediate_steps](https://api.python.langchain.com/en/latest/agents/langchain.agents.agent.AgentExecutor.html#langchain.agents.agent.AgentExecutor.return_intermediate_steps)という構成を介して行うことができます。学習の観点からは非常に興味深いと言わざるを得ませんが、出力の形式を把握しようとして数時間費やした後（ヒント：[実際にはjsonではありません](https://github.com/langchain-ai/langchain/issues/10099)）、それを解析する価値はあまりないと判断しました。

では、単純な戦略はどのように機能するのでしょうか？デバッグフラグと、そのフラグをオン/オフに切り替えるツールを追加しています。この驚くほど単純なトリックは、実際にはまったく新しい可能性の世界を開きます。エージェントに独自の動作を変更する機会を与えています！

デバッグモードの実装は、グローバル変数とそれを設定する関数で構成されています。

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

また、システムプロンプトで言及する必要があります。

```
...
   Tools:
    - you can call osquery using the call_osquery function.
    - you can enable or disable the debug mode using the set_debug_mode function.
    Context:
...
```

そして、エージェントのインスタンス化でツールリストに関数を追加します。

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

最後に、`call_osquery`を新しい`debug`フラグを使用するように適合させる必要があります。

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

これらすべての変更が適用されたので、新しく実装されたデバッグフラグを使用してエージェントが_osquery_をどのように呼び出すかを見てみましょう。

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

 • Hostname: petruzalek-mac.roam.internal                                                                                                                               
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

発行されたコマンドが「**デバッグモードで**レベル1の診断手順を実行する」であったことに注意してください。これは、エージェントの興味深い機能であるマルチツール呼び出しを示しています。必要と判断した場合、同じツールを複数回呼び出すだけでなく、異なるツールを同時に呼び出すこともできます。したがって、レポートを要求する前にデバッグモードを有効にする必要はありませんでした。エージェントは一度にすべてを行うことができました。

また、エージェントが最初にユーザー列を要求したときに失敗しましたが、その後PRAGMA命令を使用して正しいスキーマを発見し、クエリを正常に再試行したことにも注意してください。これは、改善されたシステムプロンプトのおかげで、エージェントがエラーから回復する能力を完璧に示しています。

## チャット履歴の保持

今日の最後のタスクは、エージェントが話している内容を覚えているようにして、一貫した調査ラインに従って明確な質問をしたり、システムをさらに調査したりできるようにすることです。

[前の記事](https://danicat.dev/posts/20250605-vertex-ai-sdk-python/)では、LLMがステートレスであり、「ターン」を使用して会話の現在の状態を「思い出させる」必要があることを探りました。幸いなことに、LangChainではこれを手動で行う必要はなく、[チャット履歴](https://python.langchain.com/api_reference/core/chat_history.html)という機能に依存できます。

チャット履歴の美しさは、[BaseChatMessageHistory](https://python.langchain.com/api_reference/core/chat_history/langchain_core.chat_history.BaseChatMessageHistory.html#langchain_core.chat_history.BaseChatMessageHistory)を実装するものは何でもここで使用できることです。これにより、独自のデータストアを作成するなど、あらゆる種類のデータストアを使用できます。たとえば、Vertex AIの公式ドキュメントには、[Firebase、Bigtable、Spanner](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain#chat-history?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog)を使用する例があります。

現時点では本格的なデータベースは必要ないので、名前が示すようにすべてをメモリに保存する`InMemoryChatMessageHistory`で妥協します。

以下は、`chats_by_session_id`辞書を使用して複数のセッションを技術的にサポートする典型的な実装です（コードは[langchainドキュメント](https://python.langchain.com/docs/versions/migrating_memory/chat_history/#chatmessagehistory)から取得）。

```py
chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history
```

そして、以下はチャット履歴を有効にしてエージェントをインスタンス化する新しい`main`関数です。

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

私と同じ間違いをしないように、簡単な注意点です。`chat_history`引数は`Callable`型を期待しているため、そこで関数を呼び出すのではなく、関数自体を渡す必要があります。LangChainはここでファクトリパターンを使用します。提供された関数（`get_chat_history`）を`session_id`でオンデマンドで呼び出して、正しい履歴オブジェクトを取得または作成します。この設計により、エージェントは複数の個別の会話を同時に管理できます。

関数シグネチャには、1つまたは2つの引数を含めることができます。引数が1つの場合、それは`session_id`であると想定され、引数が2つの場合、それらは`user_id`と`conversation_id`として解釈されます。これに関する詳細情報は、[RunnableWithMessageHistory](https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.history.RunnableWithMessageHistory.html)ドキュメントにあります。

パズルの最後のピースは、session_idをモデルランナーに渡すことです。これは、以下のコードに示すように、config引数を介して行われます。

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

これで、セッションが生きている限り、セッションの内容がメモリに保存されているため、エージェントに「短期記憶」の情報を尋ねることができます。これにより、ほとんどの基本的な対話がより自然に感じられるようになりますが、より大きな問題の前例を開いています。セッション情報を保存できるようになったため、各イテレーションの後、それは大きくなるだけであり、クエリから自動的に生成されたデータを扱う場合、セッションコンテキストは非常に速く大きくなり、すぐにモデルの制限に達し、コンピューターのメモリの制限に達するはるか前に達します。

Geminiのようなモデルは、[長いコンテキストウィンドウ](https://ai.google.dev/gemini-api/docs/long-context)でよく知られていますが、コンテキストをデータで埋めると、100万トークンでさえ非常に速く使い果たされる可能性があります。長いコンテキストは、取得がますます困難になるため、一部のモデルで問題を引き起こす可能性もあります。これは、[干し草の山の中の針](https://cloud.google.com/blog/products/ai-machine-learning/the-needle-in-the-haystack-test-and-how-gemini-pro-solves-it?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog)問題としても知られています。

増大するコンテキストの問題に対処するための手法には、圧縮や要約などがありますが、この記事のコンテキストを短く保つために（私が何をしたかわかりますか？）、それらは次の記事のために取っておきます。

この記事のすべての変更を含む`main.py`の最終バージョンは次のようになります。

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

## 結論

この記事では、エージェントから一貫した応答を得るためにシステムプロンプトを微調整することの重要性を学びました。また、マルチツール呼び出しがどのように機能するか、ツールを使用して機能フラグを有効または無効にしてエージェントの動作を変更する方法を実際に見てきました。最後に、インメモリチャット履歴を使用してセッション状態を管理する方法について学びました。

シリーズの次の記事では、実際のデータベースを使用してセッション間の永続性を有効にする方法、トークンの概念を再検討し、コンテキスト圧縮手法について説明します。

## 付録：試してみる楽しいこと

エージェントがより堅牢になったので、このセクションを実践的なガイドとして使用して、新しい機能を実際にテストしてください。質問間でコンテキストをどのように覚えているか、作業を説明するように依頼できるかに注目してください。

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

他の興味深いプロンプトを見つけたら、以下のコメントセクションであなたの経験を共有してください。また次回！
