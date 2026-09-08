---
categories:
- Agentic Coding
date: 2026-06-10 00:00:00+00:00
heroStyle: big
summary: エージェントの Hooks（フック）を活用し、自律型コーディングループにエンジニアリングのベストプラクティスを組み込む方法を解説します。
tags:
  - antigravity
  - automation
  - hooks
  - security
title: "コーディングエージェントの Hooks をマスターする：実践ガイド"
slug: "mastering-hooks-in-coding-agents"
aliases:
  - "/ja/posts/20260610-mastering-hooks/"
description: "Google Antigravity をはじめとするコーディングエージェントの Hooks 活用ガイド。PreToolUse や PostToolUse イベントをインターセプトしてリンターや DLP ガードレールを適用する方法を徹底解説。"
proficiencyLevel: "Advanced"
dependencies:
  - "Google Antigravity 2.0 / agy CLI"
  - "Python 3.10+"
---

コーディングエージェントの能力の進化スピードは凄まじいものがあります。私が初めて彼らに触れたのは、Google に入社して間もない約1年前のことでした。当時、大きな話題を集めていたのは Model Context Protocol（MCP）でした。これはツールの場当たり的な個別実装を、ポータブルで再利用可能な実装へと置き換える（ほかにも多くのメリットを持つ）柔軟な技術です。

それから12か月が経ち、今では多くの人が MCP から次のトレンドである Agent Skills へと移行しているように見えます。「段階的開示（progressive disclosure）」を取り入れることで、スキルはコンテキストをより効率的に活用できるようにし、結果として全体的なトークンエコノミーを改善します。推論コストの上昇を考えれば、スキルが一気に脚光を浴びたのも不思議ではありません。

MCP とスキルの双方に熱狂的なファン層が存在しますが、この期間に登場したコーディングエージェントの標準規格の中には、これら人気の技術ほど頻繁には言及されないものの、極めて重要なものがあります。それが Hooks（フック）です。

MCP やスキルが（ツールや知識を追加することで）エージェントの能力を拡張することに焦点を当てているのに対し、Hooks は異なるレイヤーで動作します。エージェントループや開発プロセス全体を、より厳密にコントロールできるようにするのです。

## エージェントの Hooks とは

「フック」という名前は最初はピンとこないかもしれませんが、実質的には「コールバック」と同じです。つまり、エージェントの処理ライフサイクルにおける特定の瞬間に呼び出されるプロシージャ（手続き）を指します。

Hooks は次の3つの要素で構成されます：
- トリガーイベント：フックが呼び出されるタイミングです。通常は実行フェーズ（pre または post）とコンテキスト（ツール呼び出しやモデル呼び出しなど）で構成されます。たとえば Antigravity では、`PreToolUse`、`PostInvocation`、エージェント終了時の `Stop` といったイベントがあります。
- 条件またはフィルター：トリガーイベントに対する正規表現です。たとえばツール呼び出しの場合、フィルターはツール名や引数を含むことができます。たとえば、`run_command(git)` というツール呼び出しに対してフックを作成することが可能です。
- プロシージャ：シェルスクリプトやコマンドとして記述されるフックの本体です。プロシージャは、操作の許可・拒否の制御、モデルやツール呼び出しの完全なオーバーライド、あるいはログ記録やテレメトリ収集といった副作用（side effects）の実行に利用できます。

## どんなときに Hooks を使うべきか

Hooks は、エージェントのライフサイクルの特定の瞬間に介入し、カスタムコマンドやスクリプトを注入します。適切なタイミングを捉えて介入することで、処理フローを制御し、本来なら非決定論的（再現性がない）になりがちなプロセスに決定論的な結果をもたらすことができます。

たとえば、開発者はシステムプロンプトや AGENTS.md ファイル（または同様の設定）を使ってコーディングガイドラインを適用しようとすることがよくあります。しかし、大規模言語モデル（LLM）の非決定論的な性質上、プロンプトベースのガイドラインには実行の保証がありません。まったく同一のプロンプトであっても異なる結果が生じることがあり、エージェントがプロンプトの一部を選択的に無視してしまうことすらあります。

プロンプトの代わりに Hooks を使うことで、特定のアクションを強制できます。たとえば、コード編集のたびにコードが常にクリーンであることを確認するため、静的解析ツール（いわゆるリンター）を必ず実行させたいとしましょう。プロンプトに「編集後は必ずリンターを実行すること」と書いても、検証を実行するかどうかはエージェント任せになってしまいます。エージェントがその編集を「軽微」だと勝手に判断した場合、このステップをスキップしてしまうかもしれません。しかし、代わりにフック（この例ではファイル編集ツールを対象とする `PostToolUse`）を作成すれば、コード編集後に静的解析ツールが実行されることを決定論的に担保できます。

こうしたライフサイクルイベントをインターセプトすることで、エージェントの挙動制御、メトリクス収集、ワークフローの安全性確保など、さまざまなパターンを実装できます。以下でその代表例を見ていきましょう。

### エージェントを専用ツールへ誘導する

Hooks は多彩なシナリオで役立ちますが、私のお気に入りのユースケースは、エージェントの周囲にガードレールを設置すること、言い換えれば「エージェントの自律性（agency）をあえて適度に制限する」ことです。

これは、`PreToolUse` フックと、ツールへのアクセスを拒否してコーディングエージェントに「ステアリングヒント（steering hint）」を返すスクリプトを組み合わせることで実装できます。このステアリングヒントには、代わりに実行させたい指示を含めます。たとえば、Go ソースファイルの読み取りにシェルコマンドを使わせたくない場合、ステアリングヒントは次のようになります："Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead"。

### 悪意あるプロンプトを検知・遮断する

`PreInvocation` フックを使えば、入力されたプロンプトをインターセプトし、セキュリティヒューリスティクスや軽量な分類モデルに照らして評価できます。プロンプトがジェイルブレイク（脱獄）の試みのように見える場合、フックはリクエストを即座にブロックし、実行ループに到達する前にバックエンドシステムを保護できます。

### 認証情報の漏洩を防ぐ

開発者が誤って env ファイルや認証情報を、コーディングエージェントが読み取るアクティブなファイルに貼り付けてしまうことがあります。ファイル読み取りを監視する `PostToolUse` フックや、LLM に送信されるペイロードをスキャンする `PreInvocation` フックは、信頼性の高い DLP（Data Loss Prevention：データ損失防止）ゲートとして機能します。高エントロピーな文字列や標準的な API 形式に一致する文字列を検出した場合、フックはシークレットを動的にマスキング（リダクト）するか、実行を中断して認証情報の安全性を保ちます。

### メモリ管理の自動化

エージェントは通常、[Agent Platform Memory Bank](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank) や [MemPalace](https://github.com/mempalace/mempalace) のような外部メモリシステムに接続されていない限りステートレスです。

エージェントに記憶機能を追加するアプローチの1つとして、記憶の保存と検索をツールとして登録する方法があります。しかしこれを行うと、エージェントがそれらのツールを明示的に呼び出す判断をするかどうかに依存してしまいます。

Hooks の仕組みを利用すれば、メモリの永続化と検索を自動化できます。たとえばセッション終了時（`Stop` フックを使用）や、一定のターン数経過後（ステップ数やモデル呼び出し回数を監視）に[メモリ生成](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/generate-memories#triggering-memory-generation)をトリガーできます。

逆に、セッション開始時やモデル呼び出し前（`PreInvocation` フックなど）にメモリを自動検索することも可能です。たとえば Agent Platform Memory Bank では、[スコープ](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#retrieve-all)（ユーザー ID など）やクエリに基づく[類似度](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#similarity-search)でメモリを取得できます。これは本質的に、メモリを活用した検索拡張生成（RAG）です。

### テレメトリデータの収集

Hooks の仕組みは、テレメトリコレクターやロガーを配置する場所としても最適です。エージェントの内部動作に対する優れた可視性が得られます。個人的には、AI が世界を支配する前にエージェントとより良好な関係を築くべく（冗談です :)）、自分がどれだけ悪態をついているかを自覚するための「罵倒カウンター（curse word counter）」フックを作ってみたいとずっと企んでいます。

## Antigravity での Hooks 設定

エージェントエンジンごとにコールバックの用語は異なりますが、このセクションでは **Antigravity における Hooks の方言（仕様）** に焦点を当てます。

仕様の全容については、公式の [Antigravity Hooks ドキュメント](https://antigravity.google/docs/hooks) を参照してください。

Antigravity は、ワークスペース直下の `.agents/` ディレクトリ内にある `hooks.json`（またはホームディレクトリにあるグローバル設定 `~/.gemini/config/hooks.json`）を読み込みます。

先ほど説明したステアリングヒントと静的解析を実装する設定例を以下に示します：

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

これらのフックへの入力は `stdin` 経由で JSON オブジェクトとして渡されます。ここにはツール引数（`toolCall.args`）、アクティブなワークスペースパス（`workspacePaths`）、現在のセッションログのファイルパス（`transcriptPath`）などのコンテキストが含まれます。スクリプト側でこれらを評価・計算し、Antigravity に対して `"allow"`（許可）、`"deny"`（拒否）、あるいは `"ask"`（ユーザーへの確認要求）を指示する JSON レスポンスを `stdout` に出力します。

たとえば、渡されたペイロードを解析してエージェントを誘導するシンプルな Python スクリプト（`steer-go-reads.py`）は、次のように記述できます：

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

## コントロールを取り戻す

エージェントはますますスマートで高速になり、私たちがかつてないスピードでコードを生成できるよう支援してくれます。しかし、制御を欠いたスピードは大惨事を招く元凶でしかありません。私は登壇時によくこんな比喩を使います。スピードが出る車が好きなら、何よりも気にかけるべきはエンジンではなくブレーキです。もしブレーキがエンジンよりも非力なら、車を止めることができず、安全性が損なわれます。

コーディングエージェントにもまったく同じ考え方を適用すべきです。コードを高速に書きたいのであれば、品質を犠牲にしてバグを混入させないための強力な制御システムが欠かせません。それを怠れば、遅かれ早かれアプリケーションが深刻なトラブルに見舞われることになります。

Hooks は、AI の自律性と堅牢なソフトウェアエンジニアリングとのギャップを埋め、私たちが再び主導権を握るためのガードレールを構築するのに最適な仕組みです。最近 [Joe Bertolami 氏の記事](https://venturebeat.com/technology/agentic-ai-solved-coding-and-exposed-every-other-problem-in-software-engineering) で読んだ通り、「コードを書くこと自体がボトルネックだったことなど一度もない」のです。私たちが何十年もかけて培ってきたエンジニアリングのベストプラクティスを決して忘れず、エージェントに適材適所のツールを備え付けることで、現代のエージェンティック・コーディングの体験を心から楽しみましょう。
