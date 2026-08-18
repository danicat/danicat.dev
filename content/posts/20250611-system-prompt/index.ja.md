---
categories:
- Agent Development
date: '2025-06-11T00:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 3
summary: システム指示（System Instructions）、セッション履歴、エージェントツールの概念を学び、よりインテリジェントな診断アシスタントを構築します。
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: システム指示とエージェントツールの実践ガイド
---
## はじめに

本ガイドでは、システム指示（System Instructions / プロンプト）とエージェントツール（Agent Tools）の設計を深掘りし、システム診断エージェントの完成度を高めていきます。[Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog)、LangChain、Gemini、および [Osquery](https://www.osquery.io/) を組み合わせて実装します。

正直なところ、[第1弾で作成した初期バージョン]({{< ref "/posts/20250531-diagnostic-agent" >}}) の診断エージェントは実用レベルとは言えませんでした。内部で実際にどのようなクエリを実行しているのか可視化できず、同一セッション内の会話履歴を記憶できず、時折ユーザーの指示を勝手に無視してしまうこともありました。

これでは自律的なエージェントとして不十分です。理想的な診断エージェントは、失敗から学習して指示を一貫して実行できる（例えば、存在しないカラム名を検知した際に自己修正する）必要があります。また、返された情報の正確性を担保するため、実行された SQL クエリをいつでも追跡・確認できる可観測性（オブザーバビリティ）が求められます。

今回はこれらの課題を解決し、本格的な緊急診断エージェント（Emergency Diagnostic Agent）を構築します！

## 環境の準備

前回は手軽さを優先して Jupyter Notebook を使用しましたが、今回はターミナルで対話できる標準的な Python プログラムとして作成します。

まずはクリーンな仮想環境を用意します：

```shell
mkdir -p ~/projects/diagnostic-agent
cd ~/projects/diagnostic-agent
python3 -m venv venv
source venv/bin/activate
pip install --upgrade google-cloud-aiplatform[agent_engines,langchain] rich
```

初期状態の `main.py` は以下のようになります：

```python
import os
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown

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
    """osquery を使用してオペレーティングシステムに問い合わせを行います。
      
    この関数は、現在のマシン、OS、および実行中のプロセスに関する情報を取得するために osquery プロセスへクエリを送信します。
    また、sqlite_master、sqlite_temp_master などのシステムテーブルや仮想テーブルを使用して、osquery インスタンスに関する情報を調べることもできます。

    Args:
        query: str  osquery テーブルに対する SQL クエリ（例: "select timestamp from time"）

    Returns:
        ExtensionResponse: リクエストのステータスとクエリ結果を含むレスポンス
    """
    if not instance.is_running():
        instance.open()

    result = instance.client.query(query)
    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()
    
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

        if query in ("exit", "quit"):
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

`python main.py` でエージェントを起動できます：

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

What is the nature of your diagnostic emergency?
>> 
```

前回のコードからの主な変更点は2つあります：1つ目は `exit` または `quit` が入力されるまで対話を継続するメインループを追加した点、2つ目はシステムプロンプトの設計を大幅に強化した点です。

## システム指示（System Prompt）の設計

システム指示は、会話全体を通じてモデルの振る舞い、制約、優先順位を規定する最も強力なディレクティブです。

効果的なプロンプト設計として、プロンプトを **Role（役割）**、**Tools（ツール）**、**Context（コンテキスト）**、**Task（タスク）** の4つのセクションに明確に分割する構造を採用しています。

### 1. Role（役割）
エージェントの目的とペルソナを定義します。技術的な役割を与えることで、モデルが参照すべき専門知識の範囲が明確になります。

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
```

### 2. Tools（ツール）
エージェントが利用できる外部機能を明示します：

```
Tools:
  - you can call osquery using the call_osquery function.
```

### 3. Context（コンテキスト）
実行環境の制約や過去の試行で判明した注意点を指示します：

```
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
```

`PRAGMA` を使ったスキーマ確認手順を指示しておくことで、カラム名の指定ミスによるクエリエラーが発生しても、エージェントが自律的にスキーマを調べて再試行できるようになります。

### 4. Task（タスク）
リクエストの解釈手順と実行ステップを定義します：

```
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

## デバッグモードの実装

エージェントがどのような SQL クエリを生成しているかをリアルタイムで確認できるように、デバッグモード用のツール（`set_debug_mode`）を実装します。

エージェント自身がツールの実行を通じてデバッグフラグを切り替えられるようにします：

```python
debug = False

def set_debug_mode(debug_mode: bool):
    """デバッグモードを切り替えます。
    
    Args:
        debug_mode (bool): True でデバッグ有効、False で無効
    """
    global debug
    debug = debug_mode
```

`call_osquery` 関数内で `debug` フラグを参照し、発行された SQL クエリと結果を出力します：

```python
def call_osquery(query: str):
    """osquery を使用してオペレーティングシステムに問い合わせを行います。"""
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

実行例：

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

What is the nature of your diagnostic emergency?
>> run a level 1 diagnostic procedure in debug mode
Executing query:  SELECT * FROM system_info
Query result:  {'status': 'OK', 'response': [{...}]}
Executing query:  SELECT pid, name, user, cpu_percent FROM processes ORDER BY cpu_percent DESC LIMIT 10
Query result:  {'status': 'no such column: user', 'response': None}
Executing query:  PRAGMA table_info(processes)
Query result:  {'status': 'OK', 'response': [{'cid': '0', 'dflt_value': '', 'name': 'pid', 'notnull': '1', 'pk': '1', 'type': 'BIGINT'}, ...]}
```

「run a level 1 diagnostic procedure in debug mode」と指示するだけで、エージェントはデバッグツールの有効化と Osquery クエリの実行を同時に判断・実行しています。また、`user` カラムが存在せずエラーになった際、`PRAGMA` でスキーマを確認して即座に自己修復している様子が確認できます。

## チャット履歴（セッション状態）の管理

前後の文脈を保持して「前の手順を詳しく説明して」といった追加の質問に答えられるよう、LangChain の `InMemoryChatMessageHistory` を組み込みます：

```python
from langchain_core.chat_history import InMemoryChatMessageHistory
import uuid

chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history
```

エージェント呼び出し時にセッション ID を渡します：

```python
response = agent.query(
    input=query,
    config={"configurable": {"session_id": session_id}}
)
```

これにより、同一セッション内での対話履歴がメモリ上に保持され、文脈を維持した自然な対話が可能になります。

## おわりに

この記事では、一貫性のある動作を引き出すシステムプロンプトの設計、マルチツールの同時呼び出しによる設定変更、およびメモリ内チャット履歴の管理方法を学びました。

シリーズ第4弾の [Agent Development Kit (ADK) で診断エージェントを作成する]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}) では、手動のボイラープレートコードから Google ADK フレームワークへ移行し、Vertex AI RAG による動的なテーブルスキーマ探索を実装します。

ぜひ感想や試してみたプロンプトをコメント欄で教えてください！
