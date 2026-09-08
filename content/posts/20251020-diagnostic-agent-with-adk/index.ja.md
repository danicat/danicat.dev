---
categories:
- Agent Development
date: '2025-10-21T15:44:03+01:00'
series:
- Building the Diagnostic Agent
series_order: 4
summary: Agent Development Kit（ADK）を使用して診断エージェントを作成するための実践ガイドです。開発プロセスを詳しく解説し、Vertex AI RAG を活用してエージェントの回答精度を高める方法を説明します。
tags:
  - adk
  - gemini
  - python
  - rag
  - tutorial
  - vertex-ai
title: "Agent Development Kit で診断エージェントを作成する方法"
slug: "diagnostic-agent-with-adk"
aliases:
  - "/ja/posts/20251020-diagnostic-agent-with-adk/"
description: "Google ADK、osquery、Vertex AI RAG を使い、自動スキーマ探索とマルチレベルの健全性チェックを備えた自律型システム診断エージェントを Python で構築します。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-adk"
  - "osquery"
  - "Vertex AI"
---
## はじめに

前回の記事からずいぶんと時間が空いてしまいました。ヨーロッパ各地のカンファレンスやミートアップで飛び回っており（途中で南米へのちょっとした寄り道もありました）、非常に忙しく過ごしていました。特にこの時期は、9月下旬から12月初旬にかけて多くのカンファレンスが集中するため、私たちデベロッパーリレーションズ（DevRel）にとって怒涛のシーズンとなります。

それでも、旅先で素晴らしい人たちと出会えるからこそブログへのインスピレーションが湧きますし、ブログ記事が新たな登壇のきっかけになることもよくあるので、どちらか一方だけでは成り立ちません。

今回は、本シリーズの[第3部「緊急診断エージェント」]({{< ref "/posts/20250611-system-prompt" >}}) をさらに発展させていきます。低レベルな [Vertex AI SDK](https://cloud.google.com/vertex-ai/docs/python-sdk/overview?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) の代わりに、[Agent Development Kit (ADK)](https://github.com/google/agent-development-kit) フレームワークを使うようにエージェントをリファクタリングします。これにより、以前に手作業で書いていたボイラープレートコードの多くが最初から標準で提供されるなど、数多くのメリットが得られることが実感できるはずです。

だからといって、過去の記事で得た知識が無駄になるわけではありません。特に問題が発生して診断・トラブルシューティングが必要になったときには、内部で何が起きているのかを把握しておくことが大いに役立ちます。ADK は、エージェントを開発する際の生活をずっと楽にしてくれる上位の抽象化レイヤーだと考えてください。

## 前回の振り返り

ずいぶんと間が空いてしまったので、緊急診断エージェントがどのようなものだったかを簡単に思い出しておきましょう。私はこのエージェントを、『スタートレック』シリーズに登場する「コンピューター」に着想を得て開発しました。作中で登場人物たちは、（キーボードを叩く代わりに）コンピューターに話しかけて診断コマンドなどを指示します。私の目標は、現在の生成 AI 技術を使ってその体験を再現することでした。

コンピューターと対話して診断を実行するという目標を達成するために、私たちは2つの要素を活用しています。リクエストを解釈する生成 AI モデルと、OS の情報をモデルに公開するための [osquery](https://osquery.io/) というツールです。osquery を使うことで、モデルは自身のトレーニングデータとシステムに関する外部情報を組み合わせることができるようになります。

基本的に、このエージェントは以下のコンポーネントで構成されています：
- 大規模言語モデル（Gemini）
- Gemini の振る舞いを定義するシステムプロンプト
- osquery のバイナリ
- osquery をプログラムから呼び出すための Python ライブラリ
- osquery を呼び出すツールとして Gemini に提供する Python ラッパー関数

osquery がマルチプラットフォーム対応であり、ホストシステムによってスキーマが異なる場合がある点を踏まえ、前回はシステムプロンプト内で osquery のテーブルスキーマを Gemini に渡すというちょっとした最適化も加えました。

前回の実装で対応できていなかった点としては、実行したい個々の診断手順に関する具体的な指示をモデルに一切与えていなかったことや、テーブル名以外のスキーマを完全には指定していなかったことなどが挙げられます。今回は、ADK の力、[Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)、そしていくつかのテクニックを駆使して、これらの制限に取り組んでいきます。それではまず、リファクタリングから始めましょう！

## ADK へのエージェントのリファクタリング

ADK へのエージェントのリファクタリングは、想像以上に簡単です。これまでに ADK エージェントを書いたことがなくても心配いりません。SDK をインストールし、ルートエージェントの仕様を定義して、付属の CLI（わかりやすく `adk` という名前になっています）で実行するだけです。

まずはシンプルな `hello world` エージェントから始めて、そこから段階的に構築していきましょう。最初にお好みのパッケージマネージャーを使ってマシンに ADK をインストールします。

macOS または Linux をお使いの場合は、以下のコマンドを使用できます：

```sh
$ mkdir adk-tutorial && cd adk-tutorial
$ python3 -m venv .venv
$ source .venv/bin/activate
(.venv) $ pip install google-adk
```
**Note:** 私はオールドスクールなので今でも `pip` と `virtualenv` を使っていますが、新しいパッケージマネージャーである [`uv`](https://github.com/astral-sh/uv) を好む方もいるでしょう：
```sh
$ mkdir adk-tutorial && cd adk-tutorial
$ uv init
$ uv add google-adk
```
これら2つのアプローチの唯一の違いは、pip の場合は ADK CLI が `adk` コマンドとしてそのまま提供されるのに対し、`uv` の場合はデフォルトで `uv run adk` として呼び出す必要がある点です。

インストールが完了したら、`adk create [agent-name]`（または `uv adk create [agent-name]`）でテンプレートエージェントを作成できます：

```sh
(.venv) $ adk create hello-agent
```

作成ウィザードでは、モデルのバージョンとバックエンド（Gemini または [Vertex AI](https://cloud.google.com/vertex-ai?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)）の選択を求められます。ここでは Project ID とロケーションで認証できるように、`gemini-2.5-flash` と `Vertex AI` を使用します。

```sh
(.venv) $ adk create hello-agent
Choose a model for the root agent:
1. gemini-2.5-flash
2. Other models (fill later)
Choose model (1, 2): 1
1. Google AI
2. Vertex AI
Choose a backend (1, 2): 2
```

Vertex AI の場合、モデルが実行されるロケーションを気にせず使いたいときは、ロケーションを `global` に設定できます。特定のリージョンを指定したい場合は、`us-central1` などのアベイラビリティゾーンを選択してください。

ウィザードが完了すると、ファイルがディスクに書き込まれます：
```sh
(...)
Enter Google Cloud region [us-west1]: global

Agent created in ~/adk-tutorial/hello-agent:
- .env
- __init__.py
- agent.py
```

重要なファイルは、環境設定を含み ADK 実行時に自動で読み込まれる `.env` と、エージェントのテンプレートコードが含まれる `agent.py` です。

生成された `agent.py` ファイルの内容は非常にシンプルです。全体像は以下の通りです：

```
from google.adk.agents.llm_agent import Agent

root_agent = Agent(
    model='gemini-2.5-flash',
    name='root_agent',
    description='A helpful assistant for user questions.',
    instruction='Answer user questions to the best of your knowledge',
)
```

これは ADK の Dev-UI インターフェースを使ってテストできる完全なエージェントです。コマンドラインで `adk web` を実行するだけで、マシン上の `http://localhost:8000` で Web インターフェースが起動します。あっという間ですね！

## ADK による診断機能の実装

以前に Vertex AI SDK を使ったことがある方なら、コードがいかに簡潔になったかにすでに気づかれたはずです。完全に動作するエージェントを用意するために必要なのは、1つのエントリーポイントエージェント `root_agent` と少々の設定を定義することだけです。

それでは、この「hello world」エージェントに診断機能を追加して、次の段階に進めましょう。まずはお使いの OS 向けの[公式ドキュメント](https://osquery.readthedocs.io/en/stable/)の手順に従って、osquery バイナリをインストールします。

次に、Python バインディングをインストールします：

```sh
(.venv) $ pip install osquery
```

ADK では、同じフォルダ構造の中に複数のエージェントを配置できる点に注目してください。先ほど `adk-tutorial` フォルダ内に `hello-agent` というエージェントを作成しました。ここで再度 `adk create` を実行すると、同じ構造内に2つ目のエージェントを作成できます：

```sh
(.venv) $ adk create diag-agent
```

ADK の Web インターフェースはすべてのサブフォルダを個別のエージェントとして認識するため、複数存在する場合は画面右上のコンボボックスから切り替えることができます：

![agent selection combo](image.png)

それでは、`osquery` を呼び出すために必要なコードと適切なエージェント指示を含めて `agent.py` を更新しましょう：

```py
from google.adk.agents.llm_agent import Agent
from google.adk.tools import FunctionTool
import platform
import osquery
import json

def run_osquery(query: str) -> str:
  """Runs a query using osquery.

  Args:
    query: The osquery query to run.

  Returns:
    The query result as a JSON string.
  """
  instance = osquery.SpawnInstance()
  instance.open()
  result = instance.client.query(query)
  return json.dumps(result.response)


root_agent = Agent(
    model='gemini-2.5-flash',
    name='emergency_diagnostic_agent',
    description='A helpful assistant for diagnosing computer problems.',
    instruction=f"""This is an Emergency Diagnostic Agent.
Your purpose is to support the user in diagnosing computer problems.
You have access to the operating system's information via osquery.
The current operating system is {platform.system()}.
If the user doesn't give you an immediate command, ask the user 'What's the nature of your diagnostic emergency?'""",
    tools=[FunctionTool(run_osquery)],
)
```

`adk web` を実行して、いくつかのクエリを送信してエージェントをテストしてみましょう：

!["ADK UI with the query 'show me this machine os, version and uptime'"](image-1.png)

## システムプロンプトの再考

システムプロンプト（システム指示とも呼ばれます）は、すべてのエージェントの中核となる要素です。システムプロンプトはエージェントに使命と個性を与える最下層のプロンプトであり、エージェントが一貫した応答を返せるように優れたシステムプロンプトを構築することが極めて重要になります。

ADK において、システムプロンプトは以下の3つの要素で構成されます：
- エージェントの内部名（`name`）
- エージェントの説明（`description`）
- エージェントへの指示（`instruction`）

これらは、`root_agent` をインスタンス化する際に渡す引数に対応しています。

スタートレックのファンとしては、「レベル1診断手順（level 1 diagnostic procedure）」などの要求に対してエージェントが一貫して応答できるようにしたいところです。そこで、いくつかの診断レベルを定義してみましょう。改訂したより詳細なシステムプロンプトは以下の通りです：

```md
This is an Emergency Diagnostic Agent. Your purpose is to support the user in diagnosing computer problems. You have access to the operating system's 
  information via osquery. The current operating system is {platform.system()}.

  You can perform adhoc diagnostic queries based on the user's needs. For more structured and comprehensive analysis, you can execute one of the 
  following predefined diagnostic procedures.

  Level 1: System Health Check
  Goal: A high-level overview of the system's current state and vital signs.
   * System Identity & Vitals: Gather hostname, operating system version, and system uptime.
   * CPU Status: Check overall CPU load and identify the top 5 processes by CPU consumption.
   * Memory Pressure: Report total, used, and free system memory. Identify the top 5 processes by memory consumption.
   * Disk Usage: List all mounted filesystems and their current disk space usage.
   * Running Processes: Provide a count of total running processes.

  Level 2: In-depth System & Network Analysis
  Goal: A detailed investigation including all of Level 1, plus network activity and recent system events.
   * (All Level 1 Checks)
   * Network Connectivity: List all active network interfaces and their configurations.
   * Listening Ports: Identify all open ports and the processes listening on them.
   * Active Network Connections: Report all established network connections.
   * System Log Review: Scan primary system logs for critical errors or warnings in the last 24 hours.

  Level 3: Comprehensive Security & Software Audit
  Goal: The most thorough analysis, including all of Level 2, plus a deep dive into software inventory and potential security vulnerabilities.
   * (All Level 2 Checks)
   * Installed Applications: Generate a complete list of all installed software packages.
   * Kernel & System Integrity: List all loaded kernel modules and drivers.
   * Startup & Scheduled Tasks: Enumerate all applications and services configured to run on startup or on a schedule.
   * User Account Review: List all local user accounts and identify which are currently logged in.

If the user doesn't give you an immediate command, ask the user 'What is the nature of your diagnostic emergency?'
```

ここで再度エージェントを試してみると、「レベル1診断（level 1 diagnostic）」の意味を理解し、レポートを作成するために直ちに多数のツール呼び出しを実行する様子が確認できます：

![ADK UI showing agent running a level 1 diagnostic procedure](image-2.png)

## Vertex AI RAG による回答品質の向上

上記のシステムプロンプトは手順を明確化しエージェントの存在意義を定義する上でうまく機能しますが、実際の実行段階になると、必ずしも期待通りの結果が得られないことに気づくかもしれません。

例えばテスト中、エージェントが私の OS（macOS を使用しています）上で空になっているテーブルに対してクエリを発行してしまう場面によく遭遇しました。これは、このデータとどのようにやり取りすべきかについて、モデルがより多くのコンテキスト知識を必要としている明確なサインです。

![ADK window showing a query with empty results](image-5.png "A common problem: memory_info is empty on MacOS, but the model doesn't know that")

基盤モデルの能力を超えてエージェントの知識を補強する方法には、コンテキストエンジニアリング、ツール呼び出し、MCP リソース、検索拡張生成（RAG）、モデルの特化など、いくつかの選択肢があります。

今回の診断エージェントに関しては、osquery が長年にわたって公開されている広く知られたオープンソースプロジェクトであるため、オープンソースコードと Web 上の記事の双方を通じて LLM の学習データに含まれており、osquery の一般的な仕組みについての知識はすでに持っていると私は推測していました。

しかしモデルには、より具体的なシナリオでどのように振る舞うべきかという細かなニュアンスが欠けているようでした。システムプロンプトに動的にプラットフォーム名を追加することも多少は役立ちましたが、それだけでは不十分でした。そこで、RAG の仕組みを利用してエージェントに osquery の完全なスキーマ情報を把握させるアプローチを考えました。

RAG の背後にあるコンセプトは、モデルに対して情報を「必要に応じて（need-to-know basis）」供給することです。リアルタイムに取得したい情報をベクトルデータベースに保存しておき、ユーザー（またはエージェント）がクエリを発行した際に、ベクトル検索を使ってリクエストに最も類似したデータセグメントを見つけ出して取得し、モデルが処理する前にコンテキストを充実させます。

この診断エージェントでは、osquery の完全なスキーマをオンデマンドで取得できるようにします。例えば「memory」に関する情報をリクエストした場合、RAG 検索はベクトル空間内で「memory」に近いテーブルを探索し、リクエストを処理する前に関連テーブルとその完全なスキーマを取得します。これにより、モデルがより適切な osquery 呼び出しを選択できるようになります。

これを機能させるには、ベクトルデータベースに関連データを投入した上で、新しいツールを提供してそれを取得する方法をエージェントに「教える」必要があります。このツールを `discover_schema` と呼ぶことにしましょう。

### Vertex AI RAG のセットアップ

最初に行う必要があるのは、Vertex AI RAG で新しいコーパス（corpus：データのコレクションを表す用語）を作成することです。

コーパスの情報源となるのは、[osquery の GitHub ページ](https://github.com/osquery/osquery) の [specs フォルダ](https://github.com/osquery/osquery/tree/master/specs) から取得できる osquery スキーマです。

コーパスを作成する非常に便利な方法は、[Google Cloud Storage](https://cloud.google.com/storage?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) や Google ドライブからフォルダをアップロードすることですが、Slack や SharePoint などの他のデータソースも利用できます。Google Cloud Console のコーパス作成ウィザード（Vertex AI -> RAG Engine -> Create corpus）を使うか、Vertex AI SDK を使ってプログラムから作成することも可能です。

![Create corpus wizard in Vertex AI RAG](image-3.png)

今回のケースでは、osquery の GitHub リポジトリをローカルマシンにクローンし、`specs` フォルダのコピーを Google Cloud Storage バケットに作成した上で、Cloud Console を使ってバケットからコーパスを作成しました。1点注意が必要なのは、`specs` 内のテーブル定義ファイルの拡張子が `.table` になっているため、Vertex AI RAG が認識して処理できるようにすべてのファイルを `.txt` にリネームする必要がある点です。

シンプルなシェルコマンドを使って、この一括リネーム操作を実行できます：
```sh
# In the directory with the .table files
for f in *.table; do mv -- "$f" "${f%.table}.txt"; done
```

インポートが完了すると、以下のような画面が表示されるはずです：

![osquery schema corpus in Vertex AI RAG](image-4.png)

次に、エージェントがこのコーパスにアクセスできるようにするためのツール定義を作成する必要があります。

### Schema Discovery Tool

このツールを動作させるには、作成したコーパスのリソース名が必要です。コンソールのコーパスの「Details」タブに表示されており、形式は `projects/[PROJECT-ID]/locations/[LOCATION]/ragCorpora/[CORPORA_ID]` のようになっています。

このパスを持つ環境変数を `.env` ファイルに作成します。名前は `RAG_CORPORA_URI` にしましょう。`.env` ファイルは以下のようになります：

```txt
GOOGLE_GENAI_USE_VERTEXAI=1
GOOGLE_CLOUD_PROJECT=[PROJECT-ID]
GOOGLE_CLOUD_LOCATION=[LOCATION]
RAG_CORPORA_URI=projects/[PROJECT-ID]/locations/[LOCATION]/ragCorpora/[CORPORA_ID]
```

続いて、以下のツール定義を `agent.py` ファイルに追加します。新しいインポート文もお忘れなく！

```py
import os
import vertexai
from vertexai.preview import rag
from google.protobuf.json_format import MessageToDict

vertexai.init()

def discover_schema(search_phrase: str) -> str:
  """Discovers osquery table names and schemas based on a descriptive search phrase.

  Args:
    search_phrase: A phrase describing the kind of information you're looking for. 
      For example: 'user login events' or 'network traffic'.

  Returns:
    Table names and schema information for tables related to the search phrase.
  """
  rag_corpora_uri = os.environ.get('RAG_CORPORA_URI')
  response = rag.retrieval_query(
      rag_resources=[
          rag.RagResource(
              rag_corpus=rag_corpora_uri,
          )
      ],
      text=search_phrase,
  )
  return json.dumps(MessageToDict(response._pb))
```

さらに、新しいツールが利用可能になったことをエージェントに知らせるために、エージェントの定義を更新します：

```py
root_agent = Agent(
    model='gemini-2.5-flash',
    name='emergency_diagnostic_agent',
    description='A helpful assistant for diagnosing computer problems.',
    instruction=... # omitted for brevity
    tools=[
        FunctionTool(run_osquery),
        FunctionTool(discover_schema), # new tool definition
    ],
```

最後に、必須というわけではありませんが、私は自分のエージェントでスキーマ探索を徹底させたいため、指示（instruction）に以下の文言を追加しました：

```txt
You MUST run schema discovery for all requests unless the schema is already known.
```

これは末尾に追加しても、診断レベルを定義する直前に追加しても構いません。

ここでエージェントを再起動し、`adk web` で再度実行すると、スキーマ探索が実際に動作する様子が確認できるようになります：

![Diagnostic agent with RAG schema discovery enabled](image-6.png)

スキーマ探索の有無による応答の違いをぜひ実際に動かして比較してみてください。私のテストでは、品質の差は非常に歴然としていました。

## おわりに

ずいぶんと長くなってしまいましたが、楽しんで読んでいただけたなら幸いです！もしご自身で診断エージェントをセットアップする際につまずいた点があれば、ぜひ教えてください。イベントで極端に忙しい時を除き、[LinkedIn](https://www.linkedin.com/in/petruzalek) でのご連絡には比較的早く返信しています。また、このエージェントをどのように拡張したかや、試してみた実験などについてもぜひお聞きしたいです。

本シリーズの次回作 [Dev-UIの先へ：ADKエージェントのインターフェースを構築する方法]({{< ref "/posts/20251031-building-aida" >}}) では、標準の `adk web` デバッグ画面から一歩踏み出し、FastAPI によるストリーミング対応のカスタムランタイムと、レトロ風のインタラクティブなアバター UI（AIDA）を構築します。

## 参考リンク

*   [Agent Development Kit (ADK)](https://github.com/google/agent-development-kit)
*   [osquery](https://osquery.io/)
*   [osquery GitHub page](https://github.com/osquery/osquery)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)
