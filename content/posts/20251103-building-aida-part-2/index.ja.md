---
categories:
- Agent Development
date: '2025-11-03T09:00:00Z'
series:
- Building the Diagnostic Agent
series_order: 6
summary: AI エージェントを完全オフラインで動作させる方法を解説します。クラウドモデルから Ollama 経由のローカル Qwen 2.5 への切り替え、および SQLite と sqlite-rag を使った Osquery スキーマとクエリパックのローカル RAG 構築手順を詳しく説明します。
tags:
  - adk
  - gemini
  - ollama
  - python
  - rag
  - sqlite
  - tutorial
title: ADK、Ollama、SQLite で完全オフラインなエージェントを構築する方法
---
[シリーズ第5弾]({{< ref "/posts/20251031-building-aida" >}}) では、エージェント用のカスタムインターフェースの構築に焦点を当てました。UI の完成度は大きく向上しましたが、1つ決定的な課題が残っていました。それは「ネットワークが切断されたときにどうするか」という問題です。

ネットワーク障害はあらゆるエージェントにとって課題ですが、私たちが作っているのは「緊急診断エージェント」です。ネットワークが落ちたまさにその時に使えない診断アシスタントでは、緊急時の役に立ちません。

そこで、ローカル環境の依存関係のみで完結するフォールバック構成を設計することにしました。これには、メインの言語モデルをローカル LLM に置き換えるだけでなく、RAG の仕組みもローカルで完結させる必要があります。

オンライン時はクラウド上の高性能な Gemini を活用し、回線障害時やプライバシー重視の環境（エアギャップ環境）ではローカルモデルにシームレスにフォールバックする構成を目指します。

本記事では、完全オフライン対応の診断エージェントを実現する手順を解説します。

## クラウドモデルからローカル LLM への切り替え

ローカルで LLM を手軽に実行する標準的な選択肢として [**Ollama**](https://ollama.com/) を使用します。macOS の場合は Homebrew でインストールできます：

```bash
brew install ollama
```

インストール後、モデルをダウンロードします：

```bash
ollama pull qwen2.5
```

[`qwen2.5`](https://ollama.com/library/qwen2.5) ファミリーには様々なサイズ（1B、2B、7B、14B、32B）が存在します。今回の診断エージェントでは、[**Osquery**](https://osquery.io/) や RAG ツールを適切に呼び出すための **Function Calling（ツール呼び出し）** 機能が不可欠です。

検証の結果、性能とリソース消費のバランスが最も優れていた **Qwen 2.5 7B** を採用しました：

```bash
$ ollama show qwen2.5
  Model
    architecture        qwen2     
    parameters          7.6B      
    context length      32768     
    embedding length    3584      
    quantization        Q4_K_M    

  Capabilities
    completion    
    tools
```

### なぜ Qwen 2.5 なのか？
*   **GPT-OSS:** 自然な対話は可能ですが、ツール呼び出し時に `SELECT * FROM system_info` などのクエリを無限ループする傾向が見られました。
*   **Llama 3.1:** 引数の生成精度が不安定でした。
*   **Qwen 2.5:** 7B クラスの中で最も安定して関数呼び出しを実行し、自然な会話を維持できました。

### LiteLLM による ADK と Ollama の連携

Google ADK にローカルモデルを接続するために [**LiteLLM**](https://www.litellm.ai/) を使用します。これにより、わずか数行の変更でモデルバックエンドを切り替えることができます：

```python
# aida/agent.py
from google.adk.models.lite_llm import LiteLlm

# ollama_chat プロバイダーを指定して LiteLlm オブジェクトを作成
MODEL = LiteLlm(model="ollama_chat/qwen2.5")

root_agent = Agent(
    model=MODEL,
    name="aida",
    description="The emergency diagnostic agent",
    # ... 指示とツール定義 ...
)
```

<video controls width="100%" src="aida_demo_hd.mov">
  お使いのブラウザは動画タグをサポートしていません。
</video>
<p style="text-align: center; font-style: italic; opacity: 0.8; margin-top: 0.5rem;">Gemini 2.5 Flash とローカル Qwen 2.5 の動作比較（Apple MacBook Pro M4 / 48GB RAM で実行）。</p>

続いて、クラウド側の [**Vertex AI RAG**](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview) をローカル RAG に置き換えます。

## SQLite RAG によるローカルナレッジベースの構築

Vertex AI RAG は膨大なデータを扱う大規模エンタープライズ向けのマネージドサービスです。数百件程度の静的な Osquery スキーマを検索するためには、ややオーバースペックでした。

Osquery 自身が [**SQLite**](https://www.sqlite.org/) エンジンに基づいているため、RAG バックエンドにも SQLite を活用するオープンソースプロジェクト **[`sqlite-rag`](https://github.com/sqliteai/sqlite-rag)** を採用しました。

### Osquery スキーマの SQLite へのインジェスト

Osquery の `.table` 定義ファイルを読み込み、`schema.db` に格納するスクリプト `ingest_osquery.py` を作成します：

```python
# ingest_osquery.py
import os
from sqlite_rag import SQLiteRag

DB_PATH = os.path.abspath("schema.db")
SPECS_DIR = os.path.abspath("osquery_data/specs")

def ingest(rag: SQLiteRag, file_path: str):
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    rel_path = os.path.relpath(file_path, SPECS_DIR)
    rag.add_text(content, uri=rel_path, metadata={"source": "osquery_specs"})

if __name__ == "__main__":
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)

    print(f"RAG データベースを初期化中: {DB_PATH}...")
    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})

    files_to_ingest = []
    for root, _, files in os.walk(SPECS_DIR):
        for file in files:
            if file.endswith(".table"):
                files_to_ingest.append(os.path.join(root, file))

    total_files = len(files_to_ingest)
    for i, file_path in enumerate(files_to_ingest):
        ingest(rag, file_path)

    print("ベクトルを量子化中...")
    rag.quantize_vectors()
    rag.close()
    print("インジェストが完了しました。")
```

`quantize_vectors()` による量子化処理で、32ビット浮動小数点ベクトルを8ビット整数に変換します。これにより、データベースサイズが大幅に圧縮され、ローカルマシンの CPU 上でも高速な類似度検索が可能になります。

### ローカルスキーマ探索ツールの実装

SQLite データベースを検索する `discover_schema` ツールを実装します：

```python
# aida/schema_rag.py
import os
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
SCHEMA_DB_PATH = os.path.join(PROJECT_ROOT, "schema.db")

schema_rag = SQLiteRag.create(SCHEMA_DB_PATH, require_existing=True)

def discover_schema(search_terms: str, top_k: int = 5) -> list[DocumentResult]:
    """ローカル RAG を使用して osquery のテーブルスキーマを検索します。

    Args:
        search_terms: 検索ワード（例: 'battery', 'system information darwin'）
        top_k: 取得する最大件数

    Returns:
        関連するテーブルスキーマ定義のリスト
    """
    return schema_rag.search(search_terms, top_k=top_k)
```

ローカルの Qwen から Osquery スキーマを探索できるようになりました：

![AIDA のローカルスキーマ検索画面](image-1.png "Qwen による 'battery' 関連スキーマのローカル検索")

## Osquery クエリパックによる専門ナレッジの補強

7B パラメータのローカルモデルは、コンテキスト長（32k トークン）や複雑な SQL の自律生成力においてクラウドモデルに一歩譲ります。

この差を埋めるため、Osquery コミュニティが蓄積してきた実績のある事前定義クエリ集（**Query Packs**）を第2のローカル RAG として取り込みます。

### クエリパックの SQLite へのインジェスト

```python
# ingest_packs.py
import glob
import json
import os
import re
import sqlite3
from sqlite_rag import SQLiteRag

DB_PATH = os.path.abspath("packs.db")
PACKS_DIR = "osquery_data/packs"

def ingest_pack(rag, pack_path):
    pack_name = os.path.basename(pack_path).replace(".conf", "").replace(".json", "")
    with open(pack_path, "r") as f:
        content = re.sub(r"\s*\n", " ", f.read())
        data = json.loads(content)

    pack_platform = data.get("platform", "all")
    for query_name, query_data in data.get("queries", {}).items():
        sql = query_data.get("query")
        desc = query_data.get("description", "")
        val = query_data.get("value", "")
        platform = query_data.get("platform", pack_platform)

        text_to_embed = f"Platform: {platform}\nName: {query_name}\nDescription: {desc}\nRationale: {val}\nSQL: {sql}"
        metadata = {
            "name": query_name,
            "pack": pack_name,
            "query": sql,
            "description": desc,
            "value": val,
            "platform": platform,
        }
        try:
            rag.add_text(text_to_embed, metadata=metadata)
        except sqlite3.IntegrityError:
            pass

def main():
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)

    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})
    pack_files = glob.glob(os.path.join(PACKS_DIR, "*.conf")) + glob.glob(os.path.join(PACKS_DIR, "*.json"))

    for pack_file in pack_files:
        ingest_pack(rag, pack_file)

    rag.quantize_vectors()
    rag.close()

if __name__ == "__main__":
    main()
```

### クエリライブラリ検索ツール（search_query_library）

```python
# aida/queries_rag.py
import os
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
PACKS_DB_PATH = os.path.join(PROJECT_ROOT, "packs.db")

queries_rag = SQLiteRag.create(PACKS_DB_PATH, require_existing=True)

def search_query_library(search_terms: str, platform: str = "all", top_k: int = 5) -> list[DocumentResult]:
    """クエリパックライブラリから事前定義された関連クエリを検索します。

    Args:
        search_terms: 検索用語（例: 'malware detection', 'rootkit'）
        platform: 'darwin', 'linux', 'windows', または 'all'
        top_k: 取得件数

    Returns:
        検証済みの SQL クエリとメタデータ
    """
    if platform == "all" or platform is None:
        search_terms += " windows linux darwin"
    else:
        search_terms += f" {platform}"

    return queries_rag.search(search_terms, top_k=top_k)
```

エージェントの指示に優先ワークフローを定義します：

```python
# aida/agent.py
root_agent = Agent(
    model=MODEL,
    name="aida",
    description="緊急診断エージェント",
    instruction="""
[OPERATIONAL WORKFLOW]
1. SEARCH: 高レベルな診断タスク（例: 'rootkit の確認'）では、まず `search_query_library` を使用してください。
2. DISCOVER: 適切な事前定義クエリがない場合は、`discover_schema` でテーブルとカラムを確認してください。
3. EXECUTE: `run_osquery` でクエリを実行してください。
    """,
    tools=[
        search_query_library,
        discover_schema,
        run_osquery,
    ],
)
```

ローカル環境でマルウェア検知クエリが的確に検索・実行される様子が確認できます：

![オフラインでマルウェア検索を実行する AIDA](image-2.png "AIDA がクエリパックライブラリから適切なクエリを検索して実行")

## まとめ

本連載を通じて、シンプルなプロトタイプから始まり、システム指示の最適化、Google ADK によるアーキテクチャ刷新、カスタムストリーミング UI、そして完全オフラインでの自律稼働まで、診断エージェントの進化を辿ってきました。

AIDA プロジェクトの全ソースコードは GitHub で公開しています：**[github.com/danicat/aida](https://github.com/danicat/aida)**

連載をお読みいただきありがとうございました！ご意見やご感想、試してみた結果などをぜひコメント欄でお知らせください。

## 参考リンク

*   [Gemini 2.5 Flash](https://deepmind.google/technologies/gemini/flash/)
*   [LiteLLM](https://www.litellm.ai/)
*   [Ollama](https://ollama.com/)
*   [Osquery](https://osquery.io/)
*   [Qwen 2.5 (Ollama Library)](https://ollama.com/library/qwen2.5)
*   [SQLite](https://www.sqlite.org/)
*   [sqlite-rag](https://github.com/sqliteai/sqlite-rag)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview)
