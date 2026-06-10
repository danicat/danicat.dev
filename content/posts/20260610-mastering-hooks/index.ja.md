---
title: "コーディングエージェントにおけるフックのマスター"
date: 2026-06-10T00:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["antigravity", "agy-cli", "agentic-coding"]
summary: "エージェントフックを使用して、エンジニアリングのベストプラクティスを自律型コーディングループに組み込む方法について解説します。"
heroStyle: "big"
---

コーディングエージェントの機能は急速に進歩しています。私が初めてエージェントに触れたのは、Googleに入社して間もない約1年前のことでした。当時、大きな話題を呼んでいたのはModel Context Protocol（MCP）でした。これは、ツールのその場限りの個別実装を、ポータブルな共通実装へと置き換える（その他の機能も備えた）柔軟な技術です。

それから12か月が経ち、現在では多くの人々がMCPから離れ、次のトレンドとして『Agent Skills（エージェントスキル）』に注目しているようです。スキルは『段階的開示（progressive disclosure）』を導入することで、コンテキストのより効率的な利用を可能にし、結果として全体的なトークン消費を抑えられます。インファレンス（推論）コストが高騰する中、スキルがこれほど普及したのも不思議ではありません。

MCPとスキルの両方にそれぞれの『ファン層』が存在しますが、この期間に登場しながらも、これら人気の標準規格ほど頻繁には言及されていない、もう一つのコーディングエージェントの標準があります。それが『フック（hooks）』です。

MCPやスキルが（ツールや知識を追加することによって）エージェントの機能を拡張することに焦点を当てているのに対し、フックは異なるレイヤーで動作し、エージェントのループや開発プロセス全体をより高度に制御できるようにします。

## エージェントフックの仕組み

『フック』という名前は最初はピンとこないかもしれませんが、フックとは要するに『コールバック』のことです。つまり、エージェントの処理ライフサイクルにおける特定の瞬間に呼び出されるプロシージャ（手続き）です。

フックは次の3つの要素で構成されます：
- **トリガーイベント（Trigger Event）**：フックが呼び出されるタイミングです。通常は、実行フェーズ（preまたはpost）と、ツール呼び出しやモデル呼び出しなどのコンテキストで構成されます。たとえばAntigravityでは、`PreToolUse`や`PostInvocation`といったイベントがあります。一部のエンジンでは、エージェント終了時の`Stop`フックや、自動コンパクション実行前の`PreCompact`フック（Claudeベースの実装に特有のもの）もサポートされています。
- **条件またはフィルター（Condition or Filter）**：トリガーイベントに基づく正規表現です。たとえば、ツール呼び出しの場合、フィルターをツール名に設定したり、ツールの引数を含めたりすることができます。具体的には、`run_command(git)`というツール呼び出しに対してフックを作成することが可能です。
- **プロシージャ（Procedure）**：シェルスクリプトやコマンドとして記述される、フックの本体処理です。このプロシージャを使用して、操作の許可・禁止を制御したり、モデルやツールの呼び出しを完全にオーバーライド（上書き）したり、ログ記録やテレメトリ（データ収集）などの副作用を発生させたりできます。

## フックをいつ使用すべきか

フックは、エージェントのライフサイクルの特定の瞬間をインターセプト（遮断）して、カスタムコマンドやスクリプトを注入します。適切な瞬間をインターセプトすることで、処理の流れを制御し、本来なら非決定論的（再現性がない）になりがちなプロセスに、決定論的（確実）な結果をもたらすことができます。

たとえば、開発者はシステムプロンプトやAGENTS.md（または同様のファイル）を介してコーディングガイドラインを適用しようとすることがよくあります。しかし、大規模言語モデル（LLM）の非決定論的な性質上、プロンプトベースのガイドラインには実行の保証がありません。まったく同じプロンプトであっても異なる結果が生じることがあり、エージェントがプロンプトの一部を都合よく無視することもあります。

プロンプトの代わりにフックを使用することで、特定のアクションを強制的に実行させることができます。たとえば、コード編集のたびに、コードが常にクリーンであることを確認するために、必ず静的解析ツール（いわゆるリンター）を実行させたいとします。プロンプトに『編集後は必ずリンターを実行すること』と記述した場合、エージェントはリンターを実行するかどうかを判断する『自律性（agency）』を持ってしまい、編集が『軽微』であると判断した場合には検証を完全にスキップしてしまう可能性があります。しかし、代わりにフック（この場合はファイル編集ツールをフィルタリングする`PostToolUse`）を作成すれば、編集後に静的解析ツールを実行することを決定論的に担保できます。

これらのライフサイクルをインターセプトすることで、エージェントの動作制御、メトリクスの収集、ワークフローの安全確保など、いくつかの強力なパターンを実装できます。

### エージェントを専門のツールへと誘導する

フックは多くのシナリオで役立ちますが、私のお気に入りのユースケースは、エージェントの周囲にガードレールを設置すること、言い換えれば『エージェントの自律性を制限する』ことです。

これは、`PreToolUse`フックと、ツールへのアクセスを拒否してコーディングエージェントに『誘導ヒント（steering hint）』を返すスクリプトを組み合わせることで実装できます。この誘導ヒントには、代わりに実行させたい指示を含めます。たとえば、エージェントがGoファイルを読み込むためにシェルコマンドを使用するのを防ぎたい場合、誘導ヒントは次のようになります："Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead"（ツール呼び出しがブロックされました - run_command(cat)：*.goファイルの読み込みに'cat'を使用しないでください。代わりに'smart_read'を使用してください）。

### テレメトリデータの収集

フックシステムは、テレメトリコレクターやロガーを配置するのにも適した場所であり、エージェントの内部動作に対する優れた可視性を提供します。

### 悪意のあるプロンプトのインターセプト

`PreInvocation`フックは、入力されたプロンプトをインターセプトし、安全性のヒューリスティクスや軽量な分類モデルに照らし合わせて評価できます。プロンプトがジェイルブレイク（脱獄）の試みのように見える場合、フックはリクエストを即座にブロックし、実行ループに到達する前にバックエンドシステムを保護できます。

### 認証情報の漏洩防止

開発者が誤ってenvファイルや資格情報を、コーディングエージェントが読み取るアクティブなファイルに貼り付けてしまうことがあります。ファイルの読み込みを監視する`PostToolUse`フックや、LLMに送信されるペイロードをスキャンする`PreInvocation`フックは、信頼性の高いデータ損失防止（DLP）ゲートとして機能します。高エントロピーのキーや標準的なAPI形式に一致する文字列を検出した場合、フックは動的にシークレットを隠蔽（マスク）するか、実行を中断して認証情報の安全性を保ちます。

### メモリの管理

[Agent Platform Memory Bank](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank)や[MemPalace](https://github.com/mempalace/mempalace)のような外部メモリシステムに接続されていない限り、エージェントは通常ステートレス（状態を持たない）です。

エージェントにメモリ機能を追加する一つの方法は、記憶と検索をツールとして登録することですが、そうすると、エージェントがそれらのツールを呼び出すという決定を明示的に下すかどうかに依存することになります。

フックシステムを使用すると、メモリの永続化と検索を自動化できます。セッションの終了時（`Stop`フックを使用）や、特定のターン数経過後（ステップ数やモデルの呼び出し回数を監視）に[メモリ生成](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/generate-memories#triggering-memory-generation)を紐付けることができます。

逆に、セッションの開始時やモデルの呼び出し前（例：`PreInvocation`フック）に、自動的にメモリを検索させることも可能です。たとえばAgent Platform Memory Bankでは、[スコープ](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#retrieve-all)（例：ユーザーIDなど）や、クエリに基づく[類似度](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#similarity-search)によってメモリを検索できます。これは本質的に、メモリをベースにした検索拡張生成（RAG）と言えます。

## Antigravityにおけるフックの設定

エージェントのエンジンによってコールバックに独自用語が使用されますが、このセクションでは、具体的な**Antigravity方言**のフックに焦点を当てます。

仕様の全貌については、公式の[Antigravityフックドキュメント](https://antigravity.google/docs/hooks)を参照してください。

Antigravityは、ワークスペースの`.agents/`ディレクトリ内にある`hooks.json`ファイル（またはホームディレクトリのグローバル設定`~/.gemini/config/hooks.json`）を探します。

先ほど説明した誘導ヒントと静的解析を実装する例を以下に示します：

```json
{
  "linter-safety-gate": {
    "PostToolUse": [
      {
        "matcher": "write_to_file|replace_file_content|multi_replace_file_content",
        "hooks": [
          {
            "type": "command",
            "command": "./scripts/run-linter.sh",
            "timeout": 15
          }
        ]
      }
    ]
  },
  "restrict-cat-on-go": {
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "command": "./scripts/steer-go-reads.py"
          }
        ]
      }
    ]
  }
}
```

これらのフックへの入力は、ツールの引数（`toolCall.args`）、アクティブな`workspacePaths`、現在のセッションログのファイルパス（`transcriptPath`）などのコンテキストを含むJSONオブジェクトとして`stdin`経由で渡されます。作成したスクリプトでこれらを評価・計算し、Antigravityに対して実行を許可する（`"allow"`）、拒否する（`"deny"`）、またはユーザーに確認する（`"ask"`）かを伝えるJSONレスポンスを`stdout`に出力します。

たとえば、その入力ペイロードを解析してエージェントを誘導するシンプルなPythonスクリプト（`steer-go-reads.py`）は、以下のように記述できます：

```python
import sys
import json

def main():
    # Read and parse the incoming trigger event payload from stdin
    try:
        payload = json.load(sys.stdin)
    except Exception as e:
        # Standard safety gate fallback
        print(json.dumps({
            "decision": "deny",
            "reason": f"Failed to parse stdin payload: {e}"
        }))
        return

    tool_call = payload.get("toolCall", {})
    tool_name = tool_call.get("name")
    tool_args = tool_call.get("args", {})

    # Match the specific tool and check arguments
    if tool_name == "run_command":
        command_line = tool_args.get("CommandLine", "")
        
        # Detect if command attempts to cat any Go source files
        if "cat" in command_line and ".go" in command_line:
            response = {
                "decision": "deny",
                "reason": "Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead."
            }
            print(json.dumps(response))
            return

    # Default to allow if no rules are violated
    print(json.dumps({
        "decision": "allow"
    }))

if __name__ == "__main__":
    main()
```


## コントロールの奪還

エージェントはよりスマートに、より速くなっており、これまで以上に迅速なコード作成を可能にしています。しかし、コントロールを欠いたスピードは災害の元です。私は講演でよくこの比喩を使います。速い車が好きなら、最も気にすべきなのはエンジンではなく、ブレーキです。もしブレーキがエンジンよりも弱ければ、止まることができず、安全が損なわれてしまいます。

コーディングエージェントにも全く同じ考え方を適用すべきです。コードを迅速に書きたいのであれば、品質を犠牲にしてバグを混入させないための強力な制御システムが必要です。品質を犠牲にすれば、遅かれ早かれアプリケーションに深刻な問題が発生することになります。

フックは、AIの自律性と堅牢なソフトウェアエンジニアリングとの間のギャップを埋め、コントロールを取り戻すためのガードレールを実装するのに最適な場所です。最近、[Joe Bertolami氏によるこちらの記事](https://venturebeat.com/technology/agentic-ai-solved-coding-and-exposed-every-other-problem-in-software-engineering)で目にしたように、『コードを書くことは決してボトルネック（制限要因）ではなかった』のです。何十年にもわたって培われてきたエンジニアリングのベストプラクティスを忘れることなく、エージェントに適切なツールを装備させ、現代の自律型コーディング体験を存分に楽しみましょう。
