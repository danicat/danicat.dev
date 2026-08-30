---
categories:
- Agent Development
date: '2025-10-21T15:44:03+01:00'
series:
- Building the Diagnostic Agent
series_order: 4
summary: Google の Agent Development Kit (ADK) を使用して診断エージェントを構築し、Vertex AI RAG を活用して回答精度を向上させる開発手順を解説します。
tags:
  - adk
  - gemini
  - python
  - rag
  - tutorial
  - vertex-ai
title: "Agent Development Kit (ADK) で診断エージェントを作成する"
slug: "diagnostic-agent-with-adk"
aliases:
  - "/ja/posts/20251020-diagnostic-agent-with-adk/"
description: "Google ADK、osquery、Vertex AI RAG を使って、スキーマ自動探索とマルチレベル健全性診断を行う自律型システム診断エージェントを構築するガイド。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-adk"
  - "osquery"
  - "Vertex AI"
---
## はじめに

ヨーロッパ各地（途中で南米への寄り道もありました）でのカンファレンスやミートアップで飛び回っていたため、前回の記事からずいぶんと間が空いてしまいました。特に9月下旬から12月初旬にかけてはカンファレンスが集中するため、私たちデベロッパーリレーションズ（DevRel）にとって非常に忙しい時期にあたります。

それでも、旅先で素晴らしい人々と出会うことがこのブログへの大きな刺激となり、ブログ記事がまた新しいトーク（登壇）のネタになっていくため、どちらか一方だけでは成り立ちません。

今回は、本シリーズの[第3弾「システム指示とエージェントツールの実践ガイド」]({{< ref "/posts/20250611-system-prompt" >}}) で作成した「緊急診断エージェント」をさらに発展させていきます。低レベルな [Vertex AI SDK](https://cloud.google.com/vertex-ai/docs/python-sdk/overview?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) の代わりに、[Agent Development Kit (ADK)](https://github.com/google/agent-development-kit) フレームワークを使うようにエージェントをリファクタリングします。これにより、以前手作業で書いていた大量のボイラープレートコードが最初から標準で提供されるなど、多くの恩恵が得られることがわかるはずです。

だからといって、これまでの記事で得た知識が無駄になるわけではありません。特に問題が発生してトラブルシューティングが必要になった際、内部で何が起きているのかを把握しておくことは非常に有益です。ADK は、エージェント開発を大幅に快適にしてくれる、より上位の抽象化レイヤーだと考えてください。

## これまでの振り返り

久しぶりですので、まずは緊急診断エージェントがどのようなものだったかを簡単に振り返っておきましょう。このエージェントは、『スタートレック』シリーズに登場する「コンピューター」から着想を得て開発しました。主人公たちがキーボードを叩く代わりにコンピューターに話しかけて診断コマンドなどを実行する、あの体験を最新の生成 AI 技術で再現することが私の目標でした。

コンピューターと対話して診断を行うという目標を達成するために、私たちは2つの要素を活用しています。リクエストを解釈する生成 AI モデルと、OS の情報をモデルに公開するための [osquery](https://osquery.io/) というツールです。osquery を利用することで、モデルは自身の学習データとシステムの外部情報を組み合わせることが可能になります。

基本的に、エージェントは以下のコンポーネントで構成されています：
- 大規模言語モデル（Gemini）
- Gemini の振る舞いを定義するシステムプロンプト
- osquery のバイナリ
- osquery をプログラムから呼び出すための Python ライブラリ
- Gemini に osquery 呼び出しツールとして渡す Python ラッパー関数

osquery はマルチプラットフォーム対応であり、ホストシステムによってスキーマが異なる場合があるため、前回の実装ではシステムプロンプト内で osquery のテーブルスキーマを Gemini に渡すというちょっとした最適化も加えました。

ただし、前回の実装で対応できていなかった点もいくつかあります。実行したい個々の診断手順に関する具体的な指示をモデルに一切与えていなかったことや、テーブル名以外のスキーマ詳細を完全に指定できていなかったことなどです。本記事では、ADK のパワーと [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)、そしていくつかのテクニックを組み合わせてこれらの制約を解消していきます。ではまず、リファクタリングから始めましょう！

## ADK へのエージェントのリファクタリング

ADK へのリファクタリングは、想像以上にシンプルです。これまで ADK エージェントを書いたことがなくても心配いりません。SDK をインストールし、ルートエージェントの仕様を定義して、付属の CLI（わかりやすく `adk` という名前になっています）で実行するだけです。

まずはシンプルな `hello world` エージェントから始めて、段階的に拡張していきましょう。最初にお好みのパッケージマネージャーを使ってマシンに ADK をインストールします。

macOS または Linux をお使いの場合は、以下のコマンドを実行します：

```sh
$ mkdir adk-tutorial && cd adk-tutorial
$ python3 -m venv .venv
$ source .venv/bin/activate
(.venv) $ pip install google-adk
```

**Note:** 私は昔ながらのスタイルで `pip` と `virtualenv` を使っていますが、新しいパッケージマネージャーである [`uv`](https://github.com/astral-sh/uv) を好む方もいるでしょう：
```sh
$ mkdir adk-tutorial && cd adk-tutorial
$ uv init
$ uv add google-adk
```
これら2つの方法の唯一の違いは、pip の場合は ADK CLI が `adk` コマンドとしてそのまま使えるのに対し、`uv` の場合はデフォルトで `uv run adk` として呼び出す必要がある点です。

インストールが完了したら、`adk create [agent-name]`（または `uv adk create [agent-name]`）でテンプレートエージェントを作成できます：

```sh
(.venv) $ adk create hello-agent
```

作成ウィザードで、モデルのバージョンとバックエンド（Gemini または [Vertex AI](https://cloud.google.com/vertex-ai?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)）の選択を求められます。ここではプロジェクト ID とロケーションで認証できるように、`gemini-2.5-flash` と `Vertex AI` を選択します。

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

Vertex AI の場合、モデルが実行されるロケーションを気にせず使いたいときは `global` に設定できます。特定のリージョンを指定したい場合は、`us-central1` などの利用可能なゾーンを選択してください。

ウィザードが完了すると、ファイルがディスクに書き出されます：
```sh
(...)
Enter Google Cloud region [us-west1]: global

Agent created in ~/adk-tutorial/hello-agent:
- .env
- __init__.py
- agent.py
```

重要なファイルは、環境設定を含み ADK 実行時に自動で読み込まれる `.env` と、エージェントのテンプレートコードが含まれる `agent.py` です。

生成された `agent.py` の内容は非常にシンプルです。全体像は以下の通りです：

```py
from google.adk.agents.llm_agent import Agent

root_agent = Agent(
    model='gemini-2.5-flash',
    name='root_agent',
    description='A helpful assistant for user questions.',
    instruction='Answer user questions to the best of your knowledge',
)
```

これは ADK の Dev-UI 画面を使ってそのままテストできる完全なエージェントです。コマンドラインで `adk web` を実行するだけで、マシン上の `http://localhost:8000` で Web インターフェースが立ち上がります。これだけで動作確認の準備は完了です！

## ADK による診断機能の実装

以前に Vertex AI SDK を使ったことがある方なら、コードがいかに簡潔になったかにすでに気づかれたはずです。完全に動作するエージェントを用意するために必要なのは、1つのエントリーポイントエージェント `root_agent` と少々の設定を定義することだけです。

では、この「hello world」エージェントに診断機能を追加して、次の段階に進めましょう。まずはお使いの OS 向けの[公式ドキュメント](https://osquery.readthedocs.io/en/stable/)の手順に従って、osquery バイナリをインストールします。

次に、Python バインディングをインストールします：

```sh
(.venv) $ pip install osquery
```

ADK では、同じフォルダ構造の中に複数のエージェントを配置できる点に注目してください。先ほど `adk-tutorial` フォルダ内に `hello-agent` というエージェントを作成しました。ここで再度 `adk create` を実行すると、同じ構造内に2つ目のエージェントを作成できます：

```sh
(.venv) $ adk create diag-agent
```

ADK の Web インターフェースはすべてのサブフォルダを個別のエージェントとして認識するため、複数存在する場合は画面右上のコンボボックスから切り替えることができます：

![エージェント選択のコンボボックス](image.png)

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

`adk web` を実行して、いくつかクエリを入力してテストしてみましょう：

!["このマシンの OS、バージョン、稼働時間を表示して" というクエリを実行した ADK UI](image-1.png)

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

ここで再度エージェントを試してみると、「レベル1診断（level 1 diagnostic）」の意味を理解し、レポートを生成するために直ちに多数のツール呼び出しを実行する様子が確認できます：

![レベル1診断手順を実行するエージェントを表示した ADK UI](image-2.png)

## Vertex AI RAG による回答品質の向上

上記のシステムプロンプトは手順を明確化しエージェントの存在意義を定義する上でうまく機能しますが、実際の実行段階になると、必ずしも期待通りの結果が得られないことに気づくかもしれません。

例えばテスト中、エージェントが私の OS（macOS）上で空になっているテーブルに対してクエリを発行してしまう場面によく遭遇しました。これは、このデータとどのようにやり取りすべきかについて、モデルがより多くのコンテキスト知識を必要としている明確なサインです。

![空の結果が返るクエリを表示した ADK ウィンドウ](image-5.png "よくある問題：macOS では memory_info が空になりますが、モデルはそれを把握していません")

基盤モデルの能力を超えてエージェントの知識を補強する方法には、コンテキストエンジニアリング、ツール呼び出し、MCP リソース、検索拡張生成（RAG）、モデルの特化など、いくつかの選択肢があります。

今回の診断エージェントに関しては、osquery が長年にわたって公開されている広く知られたオープンソースプロジェクトであるため、オープンソースコードと Web 上の記事の双方を通じて LLM の学習データに含まれており、osquery の一般的な仕組みについての知識はすでに持っていると私は推測していました。

しかしモデルには、より具体的なシナリオでどのように振る舞うべきかという細かなニュアンスが欠けているようでした。システムプロンプトに動的にプラットフォーム名を追加することも多少は役立ちましたが、それだけでは不十分でした。そこで、RAG の仕組みを利用してエージェントに osquery の完全なスキーマ情報を把握させるアプローチを考えました。

RAG の背後にあるコンセプトは、モデルに対して情報を「必要に応じて（need-to-know basis）」供給することです。リアルタイムに取得したい情報をベクトルデータベースに保存しておき、ユーザー（またはエージェント）がクエリを発行した際に、ベクトル検索を使ってリクエストに最も類似したデータセグメントを見つけ出して取得し、モデルが処理する前にコンテキストを充実させます。

この診断エージェントでは、osquery の完全なスキーマをオンデマンドで取得できるようにします。例えば「memory」に関する情報をリクエストした場合、RAG 検索はベクトル空間内で「memory」に近いテーブルを探索し、リクエストを処理する前に関連テーブルとその完全なスキーマを取得します。これにより、モデルがより適切な osquery 呼び出しを選択できるようになります。

これを機能させるには、ベクトルデータベースに関連データを投入した上で、新しいツールを提供してそれを取得する方法をエージェントに「教える」必要があります。このツールを `discover_schema` と呼ぶことにしましょう。

### Vertex AI RAG のセットアップ

まず最初に行うべきことは、Vertex AI RAG で新しいコーパス（corpus：データのコレクションを表す用語）を作成することです。

コーパスの情報源となるのは、[osquery の GitHub ページ](https://github.com/osquery/osquery) の [specs フォルダ](https://github.com/osquery/osquery/tree/master/specs) から取得できる osquery スキーマです。

コーパスを作成する非常に便利な方法は、[Google Cloud Storage](https://cloud.google.com/storage?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) や Google ドライブからフォルダをアップロードすることですが、Slack や SharePoint などの他のデータソースも利用可能です。Google Cloud Console のコーパス作成ウィザード（Vertex AI -> RAG Engine -> コーパスの作成）を使うか、Vertex AI SDK を使ってプログラムから作成することができます。

![Vertex AI RAG のコーパス作成ウィザード](image-3.png)

今回のケースでは、osquery の GitHub リポジトリをローカルマシンにクローンし、`specs` フォルダのコピーを Google Cloud Storage バケットにアップロードした上で、クラウドコンソールを使ってそのバケットからコーパスを作成しました。1点注意が必要なのは、`specs` 内のテーブル定義ファイルの拡張子が `.table` になっているため、Vertex AI RAG が認識して処理できるようにすべてのファイルを `.txt` にリネームする必要がある点です。

シンプルなシェルコマンドを使って、この一括リネーム操作を実行できます：
```sh
# .table ファイルがあるディレクトリで実行
for f in *.table; do mv -- "$f" "${f%.table}.txt"; done
```

インポートが完了すると、以下のような画面が表示されます：

![Vertex AI RAG の osquery スキーマコーパス](image-4.png)

これで、エージェントがこのコーパスにアクセスできるようにするためのツール定義を作成する準備が整いました。

### スキーマ探索ツール（discover_schema）の実装

ツールを動作させるには、作成したコーパスのリソース名が必要です。コンソールのコーパスの「詳細」タブに表示されており、形式は `projects/[PROJECT-ID]/locations/[LOCATION]/ragCorpora/[CORPORA_ID]` のようになっています。

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

さらに、新しいツールが利用可能になったことをエージェントに知らせるためにエージェント定義を更新します：

```py
root_agent = Agent(
    model='gemini-2.5-flash',
    name='emergency_diagnostic_agent',
    description='A helpful assistant for diagnosing computer problems.',
    instruction=... # 簡潔さのため省略
    tools=[
        FunctionTool(run_osquery),
        FunctionTool(discover_schema), # 新しいツール定義
    ],
)
```

最後に、必須ではありませんが、私はエージェントにスキーマ探索を徹底させたいため、指示（instruction）に以下の文言を追加しました：

```txt
You MUST run schema discovery for all requests unless the schema is already known.
```

この指示は末尾に追加しても、診断レベルを定義する直前に追加しても構いません。

ここでエージェントを再起動し、`adk web` で再度実行すると、スキーマ探索が実際に動作する様子が確認できるようになります：

![RAG スキーマ探索が有効化された診断エージェント](image-6.png)

スキーマ探索の有無による応答の違いをぜひ実際に試して比較してみてください。私のテストでは、品質の差は非常に歴然としていました。

## おわりに

少し長くなってしまいましたが、楽しんで読んでいただけたなら幸いです！もしご自身で診断エージェントをセットアップする際につまずいた点があれば、ぜひ教えてください。イベントで極端に忙しい時を除き、[LinkedIn](https://www.linkedin.com/in/petruzalek) でのご連絡には比較的早く返信しています。また、このエージェントをどのように拡張したかや、試してみた実験などについてもぜひお聞きしたいです。

本シリーズの次回作 [Dev-UIの先へ：ADKエージェントのインターフェースを構築する方法]({{< ref "/posts/20251031-building-aida" >}}) では、標準の `adk web` デバッグ画面から一歩踏み出し、FastAPI によるストリーミング対応のカスタムランタイムと、レトロ風のインタラクティブなアバター UI（AIDA）を構築します。

## 参考リンク

*   [Agent Development Kit (ADK)](https://github.com/google/agent-development-kit)
*   [osquery](https://osquery.io/)
*   [osquery GitHub page](https://github.com/osquery/osquery)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)
