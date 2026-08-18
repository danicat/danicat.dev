---
categories:
- Agent Development
date: '2025-10-21T15:44:03+01:00'
series:
- Building the Diagnostic Agent
series_order: 4
summary: Google の Agent Development Kit (ADK) を使用して診断エージェントを構築し、Vertex AI RAG を活用して回答精度を劇的に向上させる方法を解説します。
tags:
  - adk
  - gemini
  - python
  - rag
  - tutorial
  - vertex-ai
title: Agent Development Kit (ADK) で診断エージェントを作成する
---
## はじめに

ヨーロッパ各地（および南米への出張）でのカンファレンスやミートアップ登壇が重なり、前回の記事から少し間が空いてしまいました。開発者リレーションズ（DevRel）の活動において、9月下旬から12月初旬は最も活発なシーズンです。

各地で多くの開発者と出会い対話することがブログ記事のインスピレーションになり、それがまた新しい講演へと繋がっていきます。

今回は、[シリーズ第3弾の「緊急診断エージェント」]({{< ref "/posts/20250611-system-prompt" >}}) をさらに進化させます。低レイヤーの [Vertex AI SDK](https://cloud.google.com/vertex-ai/docs/python-sdk/overview?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) で書いていたコードを、Google の [Agent Development Kit (ADK)](https://github.com/google/agent-development-kit) フレームワークを使ってリファクタリングします。ADK を導入することで、これまで手作業で書いていたセッション管理やツール連携などのボイラープレートコードが大幅に削減されます。

もちろん、過去の記事で学んだプロトコルやメッセージ構造の知識が無駄になるわけではありません。問題が発生した際のデバッグやトラブルシューティングには、低レイヤーの理解が不可欠です。ADK は、エージェント開発を格段に快適にする強力な高レベル抽象化レイヤーとして機能します。

## プロジェクトの振り返り

本シリーズでは、スタートレックの宇宙船でクルーがコンピューターと音声で対話しながら診断を行う世界観を、最新の生成 AI 技術で再現することを目指しています。

自然言語での診断を実現するために、次の2つを組み合わせています：
- リクエストを解釈する大規模言語モデル（Gemini）
- OS の状態を SQL クエリ形式で取得できるオープンソースツール [Osquery](https://osquery.io/)

エージェントは以下のコンポーネントで構成されています：
- 言語モデル（Gemini）
- 振る舞いを定義するシステムプロンプト
- マシン上の Osquery バイナリ
- Osquery をプログラムから実行する Python ライブラリ
- Gemini にツールとして渡すラッパー関数

前回の実装ではテーブル名の一覧を渡したものの、各診断レベルの具体的な手順やカラム詳細スキーマまでは指定していませんでした。今回は ADK と [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) を活用してこの課題を解決します。

## ADK によるエージェントのリファクタリング

ADK への移行は非常にシンプルです。SDK をインストールし、ルートエージェント（`root_agent`）の定義を記述して、付属の CLI（`adk`）で実行するだけです。

まずは仮想環境を用意して ADK をインストールします：

```shell
mkdir adk-tutorial && cd adk-tutorial
python3 -m venv .venv
source .venv/bin/activate
pip install google-adk
```

*(パッケージマネージャー `uv` をお使いの場合は、`uv init && uv add google-adk` を実行し、`uv run adk` でコマンドを実行できます)*。

インストール後、エージェントの雛形を作成します：

```shell
adk create hello-agent
```

ウィザードに従ってモデルとバックエンドを選択します：

```shell
Choose a model for the root agent:
1. gemini-2.5-flash
2. Other models (fill later)
Choose model (1, 2): 1
1. Google AI
2. Vertex AI
Choose a backend (1, 2): 2
```

Vertex AI の場合は、リージョンを `global` または `us-central1` などに指定します。

完了すると、以下のファイルが生成されます：
- `.env`: 環境変数とプロジェクト設定
- `agent.py`: エージェントのエントリーポイント

生成された `agent.py` のコードは非常にシンプルです：

```python
from google.adk.agents.llm_agent import Agent

root_agent = Agent(
    model='gemini-2.5-flash',
    name='root_agent',
    description='A helpful assistant for user questions.',
    instruction='Answer user questions to the best of your knowledge',
)
```

ターミナルで `adk web` を実行するだけで、ローカル（`http://localhost:8000`）に Dev-UI 開発画面が立ち上がり、すぐに動作テストを行えます。

## Osquery 診断機能の組み込み

続いて、診断機能を追加します。まず Osquery の Python バインディングをインストールします：

```shell
pip install osquery
```

ADK では、1つのプロジェクトフォルダ内に複数のエージェントを同居させることができます：

```shell
adk create diag-agent
```

ADK の Web 画面右上のコンボボックスから、作成したエージェントを簡単に切り替えることができます：

![ADK のエージェント切り替え画面](image.png)

`agent.py` に Osquery 実行ツールと指示を追加します：

```python
import json
import osquery
import platform
from google.adk.agents.llm_agent import Agent
from google.adk.tools import FunctionTool

def run_osquery(query: str) -> str:
    """osquery を使用してクエリを実行します。

    Args:
        query: 実行する SQL クエリ文字列

    Returns:
        クエリ結果の JSON 文字列
    """
    instance = osquery.SpawnInstance()
    instance.open()
    result = instance.client.query(query)
    return json.dumps(result.response)

root_agent = Agent(
    model='gemini-2.5-flash',
    name='emergency_diagnostic_agent',
    description='コンピューターのトラブルシューティングを行う診断アシスタント。',
    instruction=f"""あなたは緊急診断エージェントです。
目的は、ユーザーのマシン診断をサポートすることです。
osquery 経由で OS の内部情報にアクセスできます。
現在の OS は {platform.system()} です。
ユーザーから具体的な指示がない場合は、「診断が必要な緊急事態の内容を教えてください」と尋ねてください。""",
    tools=[FunctionTool(run_osquery)],
)
```

`adk web` を起動して質問してみましょう：

!["ADK Web 画面で OS バージョンと稼働時間を問い合わせる様子"](image-1.png)

## システムプロンプトにおける診断レベルの定義

「レベル1診断」「レベル2診断」といった構造化された指示に一貫して応答できるよう、システムプロンプト内に診断レベルの定義を明記します：

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

レベル1診断を指示すると、エージェントは自動的に複数の Osquery クエリを連続して発行し、包括的なレポートをまとめてくれます：

![レベル1診断を実行する ADK UI 画面](image-2.png)

## Vertex AI RAG による応答精度の向上

テストを重ねる中で、macOS 上で実行した際に存在しない（または空の）テーブル（例: `memory_info`）へクエリを発行してしまうケースが見られました：

![空のクエリ結果が返る様子](image-5.png "macOS では memory_info が空になるが、モデルはそれを事前に把握できない")

そこで、[Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) を使用して、Osquery の全テーブルスキーマをベクトル検索できるようにします。

ユーザーが「メモリ」や「ネットワーク」について尋ねた際、エージェントはまず RAG を使って関連する Osquery テーブルスキーマをベクトル空間から検索・取得し、正しいテーブルとカラムを把握した上でクエリを発行します。

### Vertex AI RAG コーパスのセットアップ

1. [Osquery 公式リポジトリの specs フォルダ](https://github.com/osquery/osquery/tree/master/specs) からテーブル定義ファイルを取得します。
2. 拡張子が `.table` のため、Vertex AI RAG で処理できるように `.txt` に一括変換します：
   ```shell
   for f in *.table; do mv -- "$f" "${f%.table}.txt"; done
   ```
3. Cloud Storage バケットにアップロードし、Vertex AI コンソールからコーパスを作成します（Vertex AI -> RAG Engine -> Create corpus）。

![Vertex AI RAG のコーパス作成画面](image-3.png)

インポートが完了すると、スキーマがベクトルインデックス化されます：

![インポート完了後の Osquery スキーマコーパス](image-4.png)

### スキーマ探索ツール（discover_schema）の実装

コンソールの詳細タブからコーパスの URI（`projects/[PROJECT-ID]/locations/[LOCATION]/ragCorpora/[CORPORA_ID]`）を取得し、`.env` に設定します：

```txt
GOOGLE_GENAI_USE_VERTEXAI=1
GOOGLE_CLOUD_PROJECT=[PROJECT-ID]
GOOGLE_CLOUD_LOCATION=[LOCATION]
RAG_CORPORA_URI=projects/[PROJECT-ID]/locations/[LOCATION]/ragCorpora/[CORPORA_ID]
```

`agent.py` にスキーマ探索関数を追加します：

```python
import json
import os
import vertexai
from google.protobuf.json_format import MessageToDict
from vertexai.preview import rag

vertexai.init()

def discover_schema(search_phrase: str) -> str:
    """検索フレーズに基づいて osquery のテーブル名とスキーマを探索します。

    Args:
        search_phrase: 取得したい情報の検索フレーズ（例: 'user login events', 'memory usage'）

    Returns:
        関連テーブル名とスキーマ定義の JSON 文字列
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

エージェント定義に新しいツールを追加します：

```python
root_agent = Agent(
    model='gemini-2.5-flash',
    name='emergency_diagnostic_agent',
    description='A helpful assistant for diagnosing computer problems.',
    instruction=... # 診断レベルの指示
    tools=[
        FunctionTool(run_osquery),
        FunctionTool(discover_schema), # RAG ツール
    ],
)
```

確実にスキーマ探索を実行させるため、システムプロンプトに以下のディレクティブを追記します：

```txt
テーブルの正確なスキーマが事前に判明している場合を除き、すべてのリクエストで必ず discover_schema を実行してください。
```

エージェントを再起動すると、RAG によるスキーマ探索が動作するようになります：

![RAG スキーマ探索が動作している様子](image-6.png)

事前にスキーマ情報を取得することで、SQL の構文エラーや無効なテーブルへのアクセスが劇的に減少します。

## おわりに

Google ADK へのリファクタリングにより、エージェントのコードベースが大幅に洗練され、拡張性が向上しました。さらに Vertex AI RAG を組み合わせることで、膨大な Osquery スキーマの中から適切なテーブルを動的に特定できるようになりました。

シリーズ第5弾の [Dev-UI を超えて：ADK エージェントのカスタムインターフェース構築]({{< ref "/posts/20251031-building-aida" >}}) では、標準のデバッグ UI を脱却し、FastAPI とリアルタイムストリーミング、AI 生成アバターを備えたカスタム UI（AIDA）を構築します。

## 参考リンク

*   [Agent Development Kit (ADK)](https://github.com/google/agent-development-kit)
*   [Osquery 公式サイト](https://osquery.io/)
*   [Osquery GitHub リポジトリ](https://github.com/osquery/osquery)
*   [Vertex AI RAG 公式ドキュメント](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)
