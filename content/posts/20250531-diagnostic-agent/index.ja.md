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
title: 'AIエージェントを使って自分のPCを「エンタープライズ号」にした話'
---
_宇宙、それは人類に残された最後の開拓地である。そこには人類の想像を絶する新しい文明、新しい生命が待ち受けているに違いない。これは人類初の試みとして5年間の調査飛行に飛び立った宇宙探査船USSエンタープライズ号の驚異に満ちた物語である。_

## はじめに

子供の頃、父の影響で私は毎日のようにこのオープニングナレーションを聞いて育ちました。父のスタートレックへの情熱が、私がソフトウェアエンジニアの道を選ぶきっかけに大きく影響したことは間違いありません。（スタートレックをご存知ない方のために補足すると、このフレーズはスタートレック初代シリーズの全エピソードの冒頭で流れていたものです）。

スタートレックは常に時代を先取りしていました。当時は物議を醸した[米テレビ史上初の異人種間のキスシーン](https://en.wikipedia.org/wiki/Kirk_and_Uhura%27s_kiss)を描き、スマートフォンやビデオ会議など、今日では当たり前となった数多くの「未来的」テクノロジーを先取りしていました。

特に印象的なのは、劇中のエンジニアたちがコンピューターとやり取りする方法です。キーボードを叩くシーンも時折見られますが、多くのコマンドは自然言語による音声で指示されます。「レベル1診断を実行せよ」といった象徴的な命令は、熱心なファンの間で[定番のジョーク](https://www.youtube.com/watch?v=cYzByQjzTb0)になるほど何度も登場しました。

それから30年以上が経ち、私たちはインターネット以上の変革をもたらすと言われる「AIの時代」を迎えています。AIが仕事に与える影響を不安視する声も多くありますが（[先週の記事でもこれについて書きました]({{< ref "/posts/20250528-vibe-coding" >}})）、スタートレックを見て育った私には、今後数年でエンジニアの役割がどのように変化していくのかが明確に見えてきます。コードを1行ずつ書き、コンパイラを通じて手動で指示を与える代わりに、私たちは間もなくコンピューターと直接対話し、ブレインストーミングを行いながら開発を進めるようになるでしょう。

今回はその世界観を体験するために、現在のテクノロジーを使って、自然言語で自分のマシンと対話できるシンプルな診断エージェントを作成してみます。

## このデモに必要なもの

開発言語には、実験がしやすい Jupyter Notebook 上の Python を使用します。主なツールとライブラリは以下の通りです：

*   [Vertex AI Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/overview?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog)
*   [Osquery](https://www.osquery.io/) および [Python バインディング](https://github.com/osquery/osquery-python)
*   [Jupyter Notebook](https://jupyter.org/) [任意]（今回は [VSCode の Jupyter 拡張機能](https://code.visualstudio.com/docs/datascience/jupyter-notebooks)を使用しています）

以下の例では Gemini 2.0 Flash を使用しますが、他の [Gemini モデルバリアント](https://ai.google.dev/gemini-api/docs/models)でも構いません。クラウド上のサーバーではなくローカルマシンの診断を行うため、今回は Google Cloud へのデプロイは行いません。

## エージェントの概要

エージェントの仕組みをすでにご存知の方は、このセクションを読み飛ばしていただいて構いません。

AI エージェントとは、周囲の環境を認識し、特定の目標を達成するために自律的なアクションを実行できる AI システムです。入力に基づいてテキストを生成することに特化した従来の LLM とは異なり、エージェントは環境と対話し、意思決定を行い、タスクを実行して目的を達成します。これは、エージェントに情報を提供し、アクションを実行できるようにする「ツール (tools)」によって実現されます。

今回は [Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog) 経由で LangChain を使用してエージェントを構築します。まず、必要なパッケージをシステムにインストールします：

```shell
pip install --upgrade --quiet google-cloud-aiplatform[agent_engines,langchain]
```

次に、gcloud CLI の Application Default Credentials (ADC) を設定します：

```shell
gcloud auth application-default login
```

*※注意：実行環境によっては、別の認証方法が必要になる場合があります。*

これで Python スクリプトの作成準備が整いました。まず、Google Cloud の Project ID とリージョンを指定して SDK を初期化します：

```python
import vertexai

vertexai.init(
    project="my-project-id",                  # あなたの Project ID
    location="us-central1",                   # Google Cloud リージョン
    staging_bucket="gs://my-staging-bucket",  # ステージング用バケット
)
```

初期設定が完了したら、Agent Engine で LangChain エージェントを作成するのは非常に簡単です：

```python
from vertexai import agent_engines

model = "gemini-2.0-flash" # お好みのモデルを試してみてください！

model_kwargs = {
    # temperature (float): トークン選択のランダム性を制御します
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
    input="今何時ですか？"
)
print(response)
```

実行すると、以下のような応答が返ってきます：

```
{'input': '今何時ですか？', 'output': '私はAIであるため、人間と同じような「現在」の時刻や場所の感覚を持っていません。私の知識はリアルタイムでは更新されません。\n\n現在の時刻を確認するには、以下の方法をお試しください：\n\n*   **デバイスを確認する:** PC、スマートフォン、タブレットに現在時刻が表示されています。\n*   **検索する:** Google などの検索エンジンで「現在時刻」と検索してください。'}
```

設定やプロンプト、モデルの挙動によっては、時刻がわからないと回答するか、あるいは「幻覚（ハルシネーション）」を起こして架空のタイムスタンプを返してくることがあります。AI 自体は時計を持っていないため、時計というツールを与えない限り、この質問に正確に答えることはできません。

## Function Calling（関数呼び出し）

エージェントの機能を拡張する最も便利な方法の 1 つが、Python 関数をツールとして渡すことです。手順はシンプルですが、関数の docstring（ドキュメント）を詳細に記述するほど、エージェントが適切に関数を呼び出しやすくなります。現在時刻を取得する関数を定義してみましょう：

```python
import datetime

def get_current_time():
    """現在時刻を datetime オブジェクトとして返します。

    Args:
        None
    
    Returns:
        datetime: datetime 型の現在時刻
    """
    return datetime.datetime.now()
```

システム時刻を返す関数ができたので、この関数をツールとしてエージェントに渡して再作成します：

```python
agent = agent_engines.LangchainAgent(
    model=model,                # 必須
    model_kwargs=model_kwargs,  # 任意
    tools=[get_current_time]
)
```

もう一度同じ質問をしてみます：

```python
response = agent.query(
    input="今何時ですか？"
)
print(response)
```

出力は以下のようになります：

```
{'input': '今何時ですか？', 'output': '現在の時刻は 2025年5月30日 18:36:42 UTC です。'}
```

エージェントがツールを使用して、実際のシステムデータを基に正確な回答を返せるようになりました。

## システム情報の収集

今回の診断エージェントには、[Osquery](https://www.osquery.io/) を使用して稼働中のマシン情報を取得する機能を持たせます。Osquery は Facebook（現 Meta）が開発したオープンソースツールで、オペレーティングシステムの内部情報を SQL クエリで取得できる「仮想テーブル」を提供します。

これにより、システム情報の取得口が 1 つに統一されるだけでなく、LLM が得意とする SQL クエリの生成能力を最大限に活かすことができます。

Osquery のインストール手順については、[公式ドキュメント](https://osquery.readthedocs.io/en/stable/) を参照してください（OS によって手順が異なります）。

Osquery がインストールできたら、Python バインディングをインストールします：

```shell
pip install --upgrade --quiet osquery
```

バインディングがインストールされると、`osquery` パッケージをインポートしてクエリを実行できるようになります：

```python
import osquery

# 一時的な拡張ソケットを使用して osquery プロセスを起動
instance = osquery.SpawnInstance()
instance.open()  # 例外が発生する可能性があります

# クエリを発行し osquery Thrift API を呼び出す
instance.client.query("select timestamp from time")
```

`query` メソッドは、クエリ結果を含む `ExtensionResponse` オブジェクトを返します：

```python
ExtensionResponse(status=ExtensionStatus(code=0, message='OK', uuid=0), response=[{'timestamp': 'Fri May 30 17:54:06 2025 UTC'}])
```

Osquery を使ったことがない方は、[公式スキーマ](https://www.osquery.io/schema/5.17.0/) を確認して、お使いの OS でどのような情報が取得できるかをぜひ確認してみてください。

### 出力のフォーマットについて

これまでの例では出力がプレーンテキストでしたが、Jupyter Notebook で実行している場合は、以下のモジュールをインポートして Markdown 形式で見やすく表示できます：

```python
from IPython.display import Markdown, display
```

エージェントの応答を Markdown でレンダリングします：

```python
response = agent.query(
    input="今日の宇宙暦（スターデート）を教えてください"
)
display(Markdown(response["output"]))
```

出力例：

```
船長日誌、追伸。現在の宇宙暦は 48972.5 です。
```

## 点と点を結ぶ：診断エージェントの完成

システム情報を取得する方法がわかったので、エージェントの仕組みと組み合わせて、マシンの状態に関する質問に答える診断エージェントを作ります。

まず、Osquery でクエリを実行する関数を定義します。これをツールとしてエージェントに登録します：

```python
def call_osquery(query: str):
    """osquery を使用してオペレーティングシステムに問い合わせを行います。
      
      この関数は、現在のマシン、OS、および実行中のプロセスに関する情報を取得するために osquery プロセスへクエリを送信します。
      また、sqlite_master、sqlite_temp_master などのシステムテーブルや仮想テーブルを使用して、osquery インスタンスに関する情報を調べることもできます。

      Args:
        query: str  osquery テーブルに対する SQL クエリ（例: "select timestamp from time"）

      Returns:
        ExtensionResponse: リクエストのステータスと、成功した場合はクエリ結果を含むレスポンス
    """
    return instance.client.query(query)
```

関数自体はシンプルですが、エージェントがツールの使い方と制約を正しく理解できるよう、docstring を詳細に記述することが極めて重要です。

テスト中によく発生した問題として、エージェントがシステム上に存在しないテーブルを参照しようとするケースがありました（例えば macOS には `memory_info` テーブルが存在しません）。

そこで、システムに存在する仮想テーブルの一覧を動的に取得してエージェントに渡すようにします。理想的にはカラム名や型定義を含むスキーマ全体を渡すのがベストですが、テーブル一覧だけでも大幅に精度が向上します。

Osquery の内部データベースエンジンは SQLite であるため、`sqlite_temp_master` から仮想テーブルの一覧を取得できます：

```python
# 現在のシステムで利用可能なテーブル一覧を取得
response = instance.client.query("select name from sqlite_temp_master").response
tables = [ t["name"] for t in response ]
```

テーブル一覧が取得できたら、システム指示（system instruction）と `call_osquery` ツールを組み込んでエージェントを作成します：

```python
osagent = agent_engines.LangchainAgent(
    model = model,
    system_instruction=f"""
    あなたは実行中のマシンに関する質問に答えるエージェントです。
    利用可能な 1 つ以上のテーブルに対して SQL クエリを実行し、ユーザーの質問に答えてください。
    常に人間が読みやすい形式で値を返してください（例: バイトではなくメガバイト、ミリ秒ではなく整形された時間形式）。
    ユーザーのリクエストは柔軟に解釈してください。例えば、アプリケーションに関する情報を求められた場合は、プロセスやサービスに関する情報を返しても構いません。リソース使用量を求められた場合は、メモリと CPU の両方の情報を返してください。
    ユーザーに確認や聞き返しを行わないでください。
    利用可能なテーブルは以下の通りです：
    ----- テーブル一覧 -----
    {tables}
    ----- テーブル一覧終了 -----

    質問:
    """,
    tools=[
        call_osquery,
    ]
)
```

これで診断エージェントが完成しました！実際に動かしてみましょう：

```python
response = osagent.query(input="現在の時刻は？")
display(Markdown(response["output"]))
```

出力：

```
現在の時刻は Fri May 30 18:08:15 2025 UTC です。
```

もう少し複雑な質問をしてみます：

```python
response = osagent.query(input="最もリソースを消費しているプロセスは何ですか？")
display(Markdown(response["output"]))
```

出力：

```
最も CPU を消費しているプロセス：
プロセス名: mediaanalysisd, PID: 1127, 合計 CPU 時間: 2876716

最もメモリを消費しているプロセス：
プロセス名: Code Helper (Plugin), PID: 2987, メモリ使用量: 1537 MB
```

スタートレック風の指示も試してみましょう：

```python
response = osagent.query(input="コンピューター、レベル1診断手順を実行して")
display(Markdown(response["output"]))
```

出力：

```
レベル1診断レポートをお知らせします：

**オペレーティングシステム:**
*   **名称:** macOS
*   **バージョン:** 15.5
*   **ビルド:** 24F74
*   **プラットフォーム:** darwin
*   **アーキテクチャ:** arm64

...（省略）...

**稼働時間 (Uptime):**
*   **合計秒数:** 261164
*   **日数:** 3日
*   **時間:** 0時間
*   **分:** 32分
*   **秒:** 44秒
```

レベル1診断の定義を明示的にコード化していないにもかかわらず、エージェントが自律的に包括的な診断レポートを作成してくれました！

セキュリティチェックの質問も試しました：

```python
response = osagent.query(input="コンピューター、マルウェアが実行されている兆候はありますか？")
display(Markdown(response["output"]))
```

出力：

```
ディスク上のファイルに関連付けられていないプロセス（マルウェアの兆候としてよく見られるもの）を確認しましたが、検出されませんでした。また、メモリと CPU 使用率の高いプロセスを調査したところ、主なリソース消費源は Visual Studio Code と Google Chrome、およびその関連ヘルパープロセスであり、正常な動作の範囲内です。

実行したチェックに基づき、現時点でシステム上に明らかなマルウェアの兆候は見つかりませんでした。
```

## おわりに

わずか数行のコードで、OS の内部状態を自然言語で対話・調査できるインターフェースをゼロから構築できました。少し拡張すれば、より詳細な診断を実行したり、問題を自律的に修復する機能を持たせることも可能です。チャーティ機関長（スコッティ）もきっと喜んでくれるはずです！

![マウスをマイクにしてコンピューターに話しかけようとするスコッティ機関長](hello-computer-hello.gif)

この記事で使用したすべてのサンプルコードは、私の [GitHub](https://github.com/danicat/devrel/blob/main/blogs/20250531-diagnostic-agent/diagnostic_agent.ipynb) で確認できます。

シリーズ第2弾の [Vertex AI SDK for Python を深く掘り下げる]({{< ref "/posts/20250605-vertex-ai-sdk-python" >}}) では、クライアントと Gemini の通信プロトコルを解明し、低レイヤーでの手動 Function Calling を実装します。

皆さんはどのようなプロンプトや診断を試してみたいですか？ぜひコメント欄で教えてください！
