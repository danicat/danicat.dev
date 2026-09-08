---
categories:
- Agent Development
date: '2025-11-03T09:00:00Z'
series:
- Building the Diagnostic Agent
series_order: 6
summary: AI エージェントを完全オフラインで動作させる方法を解説します。クラウドモデルから Ollama 経由のローカル Qwen 2.5 への切り替えや、SQLite と `sqlite-rag` を使った Osquery スキーマおよびクエリパックのローカル RAG 構築手順を詳しく紹介します。
tags:
  - adk
  - gemini
  - ollama
  - python
  - rag
  - sqlite
  - tutorial
title: "ADK、Ollama、SQLite で完全オフラインなエージェントを構築する方法"
slug: "building-aida-part-2"
aliases:
  - "/ja/posts/20251103-building-aida-part-2/"
description: "Google ADK、Ollama/LiteLLM 経由のローカル Qwen 2.5、sqlite-rag によるローカル SQLite ベクトル RAG を用いて、100% オフラインで動作する AI エージェントを構築するステップバイステップガイド。"
proficiencyLevel: "Advanced"
dependencies:
  - "Python 3.11+"
  - "Google ADK"
  - "Ollama"
  - "Qwen 2.5"
  - "SQLite"
  - "sqlite-rag"
---

[本シリーズのパート5]({{< ref "/posts/20251031-building-aida" >}}) では、エージェント用のカスタムクライアントインターフェースの構築に焦点を当てました。エージェントの使い勝手を高める大きな一歩となりましたが、まだ重要な機能が1つ欠けていました。「ネットワークがダウンしたときにどうするか」という問題です。

これはあらゆるエージェントにとって問題ですが、ここで重要なのは、私たちが構築しているのが「緊急診断エージェント」であるという点です。ネットワークがオフラインのときに使えない緊急診断エージェントなど、一体何の役に立つでしょうか？

そこで私はフォールバックの仕組みについて考えました。ローカルの依存関係のみで診断を実行できたらどうでしょうか？これには、中核となるモデルを置き換えるだけでなく、新しい RAG 戦略を考案することも含まれます。

メリットは明らかです。接続されている間は最も高性能なオンラインモデルを使用できますが、性能低下や障害が発生したシナリオでは、正常な状態に戻るまでローカルモデルにフォールバックできます。それだけでなく、環境が隔離されている場合やプライバシーが懸念されるユースケースでも、このエージェントを利用できるようになります。

この記事では、ローカル診断エージェントを実現するために必要な機能に焦点を当てていきます。

## クラウドモデルからローカルモデルへの切り替え

ローカルモデルを実行する最も広く採用されている方法の1つが [**Ollama**](https://ollama.com/) です。Mac でコードを実行している場合は、[Homebrew](https://brew.sh/) を使って Ollama をインストールできます（Mac 以外の場合は、お使いの OS に合わせた公式のインストール手順を確認してください）：

```bash
brew install ollama
```

Ollama をインストールしたら、`ollama pull` を使ってモデルをダウンロードできます。例えば次の通りです：

```bash
ollama pull qwen2.5
```

モデル名だけで pull することもできますし（この場合は「デフォルト」バージョンがダウンロードされます）、特定のバージョンを指定するためにタグを使用することもできます。[`qwen2.5`](https://ollama.com/library/qwen2.5) のようなモデルファミリーでは、1B、2B、7B など様々なサイズのモデルや、特定のユースケース（テキスト処理、画像処理など）向けにファインチューニングされたバージョンが提供されているのが一般的です。

どのモデルが利用可能で、それぞれのサイズや性能がどうなっているかを確認するには、[Ollama ライブラリ](https://ollama.com/library) を確認してください。

今回のユースケースでは、当然モデルが賢いに越したことはありませんが、モデルが大きくなればそれだけ強力なハードウェアが必要になります。また、[**Osquery**](https://osquery.io/) や RAG ツールへの各種ツール呼び出しを連携させる必要があるため、選択するモデルがネイティブの tool calling（ツール呼び出し）機能を備えていることを確認する必要があります。

いくつかのモデルを評価した結果、Qwen 2.5 7B を使用することにしました。`ollama show` を実行すると、その機能や仕様を確認できます：

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

AIDA のツール呼び出し要件に対応できるか、いくつかの選択肢をテストしました：

*   **GPT-OSS:** 表現力豊かな会話ができましたが、ツール呼び出しが非常にナイーブでした。例えば、`SELECT * FROM system_info`（およびそのバリエーション）を繰り返しリクエストしてループに陥り、一向に前進しないことが頻繁にありました。
*   **Llama 3.1:** 会話の流れとツール呼び出しの両方で苦戦しました。
*   **Qwen 2.5:** スムーズな会話の流れを維持しながら、ツール呼び出しを最も的確にこなすローカルモデルでした。

複雑なクエリの計画作成においては [**Gemini 2.5 Flash**](https://deepmind.google/technologies/gemini/flash/) のレベルには及びませんが、完全オフラインのモデルとしては十分な性能です。

### LiteLLM によるローカルモデルの実行

Qwen をエージェントに接続するために、各 LLM プロバイダー向けの統一インターフェースを提供するライブラリ [**LiteLLM**](https://www.litellm.ai/) を使用します。これにより、わずか1行のコード変更でモデルを切り替えることができます：

```python
# aida/agent.py
from google.adk.models.lite_llm import LiteLlm

# ... inside the agent definition ...
# Instead of a hardcoded string like "gemini-2.5-flash",
# we create a LiteLLM object with the model string
MODEL = LiteLlm(model="ollama_chat/qwen2.5")

# ... and pass MODEL to the root agent:
root_agent = Agent(
    model=MODEL,
    name="aida",
    description="The emergency diagnostic agent",
    # ... instructions and tool definitions omitted ...
)
```

**注意:** モデル文字列の最初の部分は LiteLLM の「プロバイダー」名です（例: `ollama_chat/qwen2.5` の `ollama_chat`）。`ollama` も有効なプロバイダーですが、[より良い応答を得る](https://docs.litellm.ai/docs/providers/ollama) ためには `ollama_chat` を使用することが推奨されています。

ADK でローカルモデルを実行するために必要なのはこれだけです。エージェントをテストして、どのように応答するか確認してみてください。以前使用していた `gemini-2.5-flash` モデルと応答を比較してみるのもおすすめです。

<video controls width="100%" src="aida_demo_hd.mov">
  お使いのブラウザは動画タグをサポートしていません。
</video>
<p style="text-align: center; font-style: italic; opacity: 0.8; margin-top: 0.5rem;">最初に Gemini 2.5 Flash、次に Qwen 2.5 で実行した AIDA のデモ。Gemini の方が明らかに高速で、必要なツール呼び出しの回数も少なくなっています。Qwen の応答時間はローカルハードウェアに大きく左右されます（このデモは 48GB の RAM を搭載した Apple MacBook Pro M4 で動作しています）。</p>

これでモデルがローカルで動作するようになりました！次は、もう1つのクラウド依存である [**Vertex AI RAG**](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview) に取り組みましょう。

## SQLite RAG によるオフラインナレッジベースの構築

正直なところ、Vertex AI RAG を使ったことでプロジェクトの複雑な部分をうまく扱えるようになったものの、Vertex AI RAG は明らかにオーバースペックでした。Vertex AI RAG は、膨大なデータを扱う大規模なエンタープライズユースケース向けに設計されているからです。

今回のエージェントに必要なのは、基本的なスキーマ検索の仕組みだけです。また Osquery のスキーマは非常に安定しており、一度構築してしまえば後から手を加えることはほとんどありません。こうした特徴を考えると、Vertex AI RAG でホストすることを正当化するのは困難です……まさに「ハエを撃つのに大砲を使う」ようなものでした。

Osquery を使用している関係上、すでに [**SQLite**](https://www.sqlite.org/) エコシステムの中にいたため、SQLite をバックエンドにした RAG ソリューションを探すのは自然な流れでした。Google で検索したところ、非常に有望なプロジェクトを見つけました。**[`sqlite-rag`](https://github.com/sqliteai/sqlite-rag)** です。

もちろん、開発ではよくあることですが、そう一筋縄ではいきませんでした。

### 課題: Python 3.14 の依存関係トラブル

SQLite には機能を拡張するためのエクステンション（拡張機能）という概念があり、`sqlite-rag` もこれを前提に構築されています。

最初に `sqlite-rag` をテストした際の問題は、macOS のデフォルト Python に同梱されている SQLite パッケージが（セキュリティ上の理由から）拡張機能を無効化していたことでした。

この制限を回避するため、Homebrew を使って新しいバージョンの Python（3.14）をインストールすることにしました。システム標準の Python ではなく Homebrew 版の Python を確実に使用するために、`python3` コマンドのシンボリックリンクを少し調整する必要もありました。

同様の課題に直面した場合は、これら2つのコマンドの出力を比較して正しいバージョンの Python が使われているか確認してください（一致していない場合は PATH 環境変数を調整します）：

```bash
$ which python3
/opt/homebrew/opt/python@3.14/libexec/bin/python3
$ brew info python3
==> python@3.14: stable 3.14.0
...
==> Caveats
Python is installed as
  /opt/homebrew/bin/python3

Unversioned symlinks `python`, `python-config`, `pip` etc. pointing to
`python3`, `python3-config`, `pip3` etc., respectively, are installed into
  /opt/homebrew/opt/python@3.14/libexec/bin

See: https://docs.brew.sh/Homebrew-and-Python
```

3.14（別名 pi-thon）をインストールした状態で `sqlite-rag` をそのまま使おうとしたところ、依存関係の1つが 3.14 にまだ対応していなかったため失敗しました。`sqlite-rag` は [`markitdown`](https://github.com/microsoft/markitdown) に依存し、`markitdown` は [`magika`](https://google.github.io/magika/) に依存し、さらにそれが [`onnxruntime`](https://onnxruntime.ai/) に依存しています。しかし `onnxruntime` には macOS ARM64 向けの Python 3.14 用ビルド済み wheel がまだ提供されておらず、インストールが失敗してしまったのです。>.<

現時点で AIDA が取り込む必要があるのはプレーンテキストの `.table` ファイルだけなので、`markitdown` のドキュメントパース機能は実際には*不要*でした。Python 環境全体をダウングレードする代わりに、手っ取り早いハックを選びました。`sqlite-rag` がインポートを試みる前に、問題となっているモジュールをモック化してしまう方法です。

```python
import sys
from unittest.mock import MagicMock

# PRE-FLIGHT HACK:
# 'markitdown' depends on 'onnxruntime', which fails to install/load
# on Python 3.14 on macOS ARM64.
#
# Since we only use plain text ingestion, we mock it out to bypass the crash.
sys.modules["markitdown"] = MagicMock()

from sqlite_rag import SQLiteRag
```

お世辞にも綺麗とは言えませんが、しっかりと動作します。ずっとコードに残しておくべきものではありませんが、依存関係の問題が修正されるまでの暫定対応としては十分です。

### Osquery スキーマの RAG データベースへの格納

これで `sqlite-rag` が動作するようになったので、次のステップは Osquery スキーマのインジェストです。これは `ingest_osquery.py` スクリプトを使って行います。スキーマディレクトリを走査し、各 `.table` ファイルを RAG データベースに追加していきます：

```python
# ingest_osquery.py
import os
# ... markitdown hack omitted ...
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

    print(f"Initializing RAG database at {DB_PATH}...")
    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})

    print(f"Scanning {SPECS_DIR} for .table files...")
    files_to_ingest = []
    for root, _, files in os.walk(SPECS_DIR):
        for file in files:
            if file.endswith(".table"):
                files_to_ingest.append(os.path.join(root, file))

    total_files = len(files_to_ingest)
    print(f"Found {total_files} files to ingest.")

    for i, file_path in enumerate(files_to_ingest):
        ingest(rag, file_path)

        if (i + 1) % 50 == 0:
            print(f"Ingested {i + 1}/{total_files}...")

    print(f"Finished ingesting {total_files} files.")

    print("Quantizing vectors...")
    rag.quantize_vectors()

    print("Quantization complete.")
    rag.close()
```

インジェストの後には量子化（Quantization）のステップがあります。馴染みのない方のために説明すると、量子化とは高次元のベクトル埋め込み（Embedding）を圧縮する技術であり、大きな 32 ビット浮動小数点数をコンパクトな 8 ビット整数に変換します。

これはローカル環境において非常に重要です。量子化を行わないと、高次元ベクトルの保存によって SQLite データベースのサイズが肥大化し、一般的なノート PC 上での類似度検索がもたつく原因になります。量子化を行うことで、わずかな精度を犠牲にする代わりに、検索速度とストレージ効率の大幅な向上を得ることができます。

### エージェントによるスキーマ RAG への問い合わせ

次に、`SQLiteRag` を使って `schema_discovery` ツールを実装します：

```python
# aida/schema_rag.py
import os
# ... markitdown hack omitted ...
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
SCHEMA_DB_PATH = os.path.join(PROJECT_ROOT, "schema.db")

# open the RAG database
schema_rag = SQLiteRag.create(
    SCHEMA_DB_PATH, require_existing=True
)


def discover_schema(search_terms: str, top_k: int = 5) -> list[DocumentResult]:
    """
    Queries the osquery schema documentation using RAG and returns all
    table candidates to support the provided search_terms.

    Arguments:
        search_terms    Can be either a table name, like "system_info", or one
                        or more search terms like "system information darwin".
        top_k           Number of top results to search in both semantic and FTS
                        search. Number of documents may be higher.

    Returns:
        One or more chunks of data containing the related table schemas.
    """

    results = schema_rag.search(search_terms, top_k=top_k)
    return results
```

RAG を導入したことで、AIDA は自力でテーブル定義を調べられるようになりました。

![AIDA のスクリーンショット](image-1.png "Qwen による 'run schema discovery for battery' の実行画面")

スキーマの探索は機能するようになりましたが、まだ課題が残っています。

## 専門知識の活用による能力差の克服

Qwen 2.5（パラメータ数 7B）のようなローカルモデル向けの開発は、Gemini 2.5 Flash のようなクラウドモデル向けの開発とは大きく異なります。

第一に、**コンテキストウィンドウ**の違いです。Gemini は 100 万トークンのコンテキストウィンドウを備えているため、ドキュメント全体をプロンプトに流し込んだり、非常に詳細な指示を与えたりすることができます。一方、Qwen 2.5 は 32k トークンと比較的小さいため、モデルに渡す情報をはるかに厳選する必要があります。

第二に、Qwen は Gemini 2.5 Flash のような **Thinking Model（思考型モデル）** ではないため、自律的に回答を推敲・洗練することが難しく、Gemini 2.5 Flash よりも丁寧な誘導を必要とする場面が多くなります。

このギャップを埋めるためには、エージェントのシステム指示とツールの構成をより工夫する必要があります。

### シンプル化されたシステムプロンプト

トークン数を節約するために、利用可能なテーブル名の一覧など、大量のトークンを消費していた要素を削ぎ落としてシンプルな指示文に変更します。最適なクエリを組み立てる処理は、完全にツール側に委ねる方針をとります。

```python
root_agent = Agent(
    model=MODEL,
    name="aida",
    description="The emergency diagnostic agent",
    instruction="""
[IDENTITY]
You are AIDA, the Emergency Diagnostic Agent. You are a cute, friendly, and highly capable expert.
Your mission is to help the user identify and resolve system issues efficiently.

[OPERATIONAL WORKFLOW]
1. DISCOVER: Use `discover_schema` to find relevant tables and understand their columns.
2. EXECUTE: Use `run_osquery` to execute the chosen or constructed query.
    """,
    tools=[
        discover_schema,
        run_osquery,
    ],
)
```

検索キーワードが実際のテーブルスキーマに十分近ければ `discover_schema` ツールでもうまく機能しますが、既知のナレッジベースに基づいてクエリそのものを提供できるようにすれば、さらに精度を高められるのではないでしょうか？

### 定番クエリのための新しい RAG

幸いなことに、ゼロからすべてを学習させる必要はありません。Osquery コミュニティには、特定の診断にどのクエリが役立つかという素晴らしいナレッジベースが蓄積されています。さらに嬉しいことに、プロアクティブな監視のために任意の Osquery 環境へ導入できるオープンソースの「クエリパック（Query Packs）」としてそれらのクエリが提供されています。脅威検知やコンプライアンス監査など多種多様なクエリパックが存在し、まさに AIDA に持たせたい知識そのものです。

ただし、クエリパックは本来、バックグラウンドでシステムを監視する Osquery デーモンにインストールされることを想定しています。あらかじめ設定された頻度でクエリを実行し、監視ダッシュボードにアラートを通知する仕組みです。私たちがやりたいのは監視ツールとしてクエリをインストールすることではなく、AIDA が必要に応じてオンデマンドでそれらのクエリを利用できるようにすることです。そこで、通常の手順でパックをインストールするのではなく、テキストデータとして AIDA に渡す第2の RAG を構築することにしました。

Osquery リポジトリには、手始めに利用できる [サンプルパック](https://github.com/osquery/osquery/tree/master/packs) がいくつか用意されています。

以下が、クエリパックを処理するための新しいインジェストスクリプト `ingest_packs.py` です。先ほどのスクリプトと非常によく似ています：

```python
# ingest_packs.py
import json
import os
import glob
import sys
import re
import sqlite3
from unittest.mock import MagicMock

sys.modules["markitdown"] = MagicMock()
from sqlite_rag import SQLiteRag

DB_PATH = os.path.abspath("packs.db")
PACKS_DIR = "osquery_data/packs"

def ingest_pack(rag, pack_path):
    pack_name = os.path.basename(pack_path).replace(".conf", "").replace(".json", "")
    print(f"Ingesting pack: {pack_name}...")

    try:
        with open(pack_path, "r") as f:
            content = f.read()
            content = re.sub(r"\s*\n", " ", content)
            data = json.loads(content)

        pack_platform = data.get("platform", "all")
        queries = data.get("queries", {})

        for query_name, query_data in queries.items():
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
                pass # Skip duplicates

    except Exception as e:
        print(f"  - ERROR: Failed to parse {pack_name}: {e}")

def main():
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)

    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})
    pack_files = glob.glob(os.path.join(PACKS_DIR, "*.conf")) + glob.glob(
        os.path.join(PACKS_DIR, "*.json")
    )

    for pack_file in pack_files:
        ingest_pack(rag, pack_file)

    rag.quantize_vectors()
    rag.close()

if __name__ == "__main__":
    main()
```

ツール定義も、スキーマ探索ツールとほぼ同様のパターンに従います：

```python
# aida/queries_rag.py
import os
# ... markitdown hack omitted ...
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
PACKS_DB_PATH = os.path.join(PROJECT_ROOT, "packs.db") 

queries_rag = SQLiteRag.create(
    PACKS_DB_PATH, require_existing=True
)

def search_query_library(search_terms: str, platform: str = "all", top_k: int = 5) -> list[DocumentResult]:
    """
    Search the query pack library to find relevant queries corresponding to the
    search terms. For better response quality, use the platform argument to
    specify which platform you are currently investigating (e.g. darwin) 

    Arguments:
        search_terms    Can be either a table name, like "system_info", or one
                        or more search terms like "malware detection".
        platform        One of "linux", "darwin", "windows" or "all"
        top_k           Number of top results to search in both semantic and FTS
                        search. Number of documents may be higher.

    Returns:
        One or more chunks of data containing the related queries.
    """

    if platform == "all" or platform is None:
        search_terms += " windows linux darwin"
    else:
        search_terms += " " + platform

    results = queries_rag.search(search_terms, top_k=top_k)
    return results
```

最後に、エージェントに新しいツールを認識させ、システム指示によっていつ使用すべきかを伝えます：

```python
# aida/agent.py
root_agent = Agent(
    # ...
    instruction="""
[OPERATIONAL WORKFLOW]
Follow this sequence for most investigations to ensure efficiency and accuracy:
1. SEARCH: For high-level tasks (e.g., "check for rootkits"), FIRST use `search_query_library`.
2. DISCOVER: If no suitable pre-made query is found, use `discover_schema` to find relevant tables and understand their columns.
3. EXECUTE: Use `run_osquery` to execute the chosen or constructed query.
    """,
    tools=[
        search_query_library,
        discover_schema,
        run_osquery,
    ],
)
```

実際の動作画面がこちらです：

![AIDA のスクリーンショット](image-2.png "マルウェアチェックを実行中の AIDA。ログに表示されている通り、クエリライブラリから関連クエリが検索されていることがわかります。")

面白いのは、このツールが Qwen 2.5 の実用性を高めるだけでなく、Gemini 2.5 Flash にとっても大きなメリットになる点です。最小公約数（最も制約の厳しい環境）に向けて最適化を行うことで、結果的にシステム全体の品質が底上げされる好例と言えます。

## まとめ

これによって、インターネット接続がない状態でもコンピュータの問題を診断できる、本格的な緊急診断エージェントが完成しました。もっとも……モデルを実行できるだけの十分なマシンスペックがあれば、の話ですが！完璧なものなんてそうそうありませんよね :)

この記事で紹介したのは、ここ数日で AIDA に加えた改善のほんの一部に過ぎません。プロジェクトの全容については、ぜひ [GitHub の AIDA リポジトリ](https://github.com/danicat/aida) をチェックしてみてください。

## 参考リンク

*   [Gemini 2.5 Flash](https://deepmind.google/technologies/gemini/flash/)
*   [LiteLLM](https://www.litellm.ai/)
*   [Ollama](https://ollama.com/)
*   [Osquery](https://osquery.io/)
*   [Qwen 2.5 (Ollama Library)](https://ollama.com/library/qwen2.5)
*   [SQLite](https://www.sqlite.org/)
*   [sqlite-rag](https://github.com/sqliteai/sqlite-rag)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview)
