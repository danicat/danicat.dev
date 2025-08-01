---
title: "Gemini CLIとGoでMCPサーバーを構築する方法"
date: 2025-07-29
author: "Daniela Petruzalek"
tags: ["go", "gemini-cli", "mcp", "ai", "codelab"]
categories: ["チュートリアル"]
summary: "Model Context Protocol（MCP）とGemini CLIを使用して、AI搭載のGo開発アシスタントであるGoDoctorを構築した方法のステップバイステップガイド。"
---
{{< translation-notice >}}

### はじめに

多くの皆さんと同様に、私もAI支援開発に深く没頭してきました。その道のりは、しばしば「すごい」という瞬間とイライラする障害のジェットコースターです。これは、そのような旅の物語の一つです。あるものを構築しようとして始まり、イライラする問題で完全に脇道に逸れ、最終的には私のAI支援ワークフローを根本的に改善するツールを手に入れることになった物語です。

私の当初の目標は、SQLを使用してマシンの状態をクエリできるツールである[osquery](https://www.osquery.io/)用のModel Context Protocol（MCP）サーバーを構築することでした。Goコードを書くのを手伝ってもらうためにGemini CLIを使うのを楽しみにしていました。しかし、私はすぐに壁にぶつかりました。エージェントが生成したGoコードは、しばしば慣用的ではありませんでした。初心者の間違いを犯し、過剰な抽象化を作成し、存在しないAPI全体を頻繁に「幻覚」しました。私の推測では、基礎となるモデルが新しい[Go SDK for MCP](https://github.com/modelcontextprotocol/go-sdk)でトレーニングされていなかったため、知らないと認めるよりも答えをでっち上げることを好んだのでしょう。

この経験から、私は重要な気づきを得ました。ツールと戦うのではなく、ツールに*教える*ことができるということです。私はosqueryプロジェクトを一時停止し、「サイドクエスト」に乗り出すことにしました。その唯一の目的は、Go開発のエキスパートである専用のMCPサーバーを構築することです。最終的に[GoDoctor](https://github.com/danicat/godoctor)と名付けたこのサイドプロジェクトは、Gemini CLIがより良いGoコードを書くために必要なツールを提供するでしょう。

この記事では、GoDoctorを構築した物語を順を追って説明します。これは従来のチュートリアルというよりも、「プロンプト駆動」の旅です。プロジェクトの要件を効果的なプロンプトに変換し、実装の詳細を通じてAIをガイドし、その過程で避けられない間違いから学ぶことに焦点を当てます。

### 舞台設定：`GEMINI.md`

コードを一行も書く前に、最初のステップは基本ルールを設定することでした。`GEMINI.md`はGemini CLIに固有のファイルですが、コンテキストファイルを作成する習慣は多くのAIコーディングエージェントで一般的です（たとえば、Julesは`AGENTS.md`を使用し、Claudeは`CLAUDE.md`を使用します）。実際、これを[`AGENT.md`](https://ampcode.com/AGENT.md)というファイルで標準化しようという新たな取り組みがあります。このファイルは、AIにプロジェクトの標準とその動作に対する期待の基礎的な理解を提供するため、非常に重要です。

このプロジェクトは全く新しいものだったので、まだ共有できる特定のアーキテクチャの詳細は何もありませんでした。したがって、私は高品質で慣用的なGoコードを作成することに焦点を当てた一般的なガイドラインのセットから始めました。プロジェクトが進化するにつれて、プロジェクトの構造、ビルドコマンド、または主要なライブラリに関するより具体的な指示を追加するのが一般的です。よりプロジェクト固有のファイルの例として、私の[`testquery`プロジェクト](https://github.com/danicat/testquery/blob/main/GEMINI.md)で使用している`GEMINI.md`を見ることができます。

これが、この旅でAIの憲法として機能した最初の`GEMINI.md`です。

```markdown
# Go開発ガイドライン
このプロジェクトに貢献するすべてのコードは、以下の原則に従う必要があります。

### 1. フォーマット
すべてのGoコードは、送信する前に`gofmt`でフォーマットする**必要があります**。

### 2. 命名規則
- **パッケージ:** 短く、簡潔で、すべて小文字の名前を使用してください。
- **変数、関数、メソッド:** エクスポートされていない識別子には`camelCase`を使用し、エクスポートされた識別子には`PascalCase`を使用してください。
- **インターフェース:** インターフェースには、`I`のようなプレフィックスではなく、何をするかに基づいて名前を付けます（例：`io.Reader`）。

### 3. エラー処理
- エラーは値です。それらを破棄しないでください。
- `if err != nil`パターンを使用して、エラーを明示的に処理してください。
- `fmt.Errorf("コンテキスト: %w", err)`を使用して、エラーにコンテキストを提供してください。

### 4. シンプルさと明快さ
- 「賢いよりも明確が良い。」理解しやすいコードを書いてください。
- 不必要な複雑さと抽象化を避けてください。
- インターフェースではなく、具象型を返すことを好みます。

### 5. ドキュメント
- エクスポートされたすべての識別子（`PascalCase`）には、ドキュメントコメントが**必要です**。
- コメントは、*何を*ではなく、*なぜ*を説明する必要があります。

# エージェントガイドライン
- **URLの読み取り:** ユーザーから提供されたURLは常に読んでください。それらはオプションではありません。
```

このファイルは、最初から品質とスタイルのベースラインを確立します。

### Model Context Protocol（MCP）を理解する

このプロジェクトの中心にあるのは、Model Context Protocol（MCP）です。一部の人々はそれを「LLMツールのためのUSB標準」と表現しますが、私は別の見方をしています。**HTTPとRESTがWeb APIの標準化のために行ったことを、MCPはLLMツールのために行っています。**RESTが大規模なWebサービスのエコシステムを解き放った予測可能なアーキテクチャを提供したように、MCPはAIエージェントの世界に待望の共通言語を提供しています。これはJSON-RPCベースのプロトコルであり、MCPを「話す」エージェントが、カスタムの一回限りの統合を必要とせずに、準拠するツールを発見して使用できる共通の基盤を作成します。

プロトコルは、エージェントとツールサーバーが通信するためのさまざまな方法を定義しており、これらはトランスポートとして知られています。最も一般的な2つは次のとおりです。
*   **HTTP:** 使い慣れたリクエスト/レスポンスモデルで、リモートサービスとしてデプロイされたツールに最適です（例：[Cloud Run](https://cloud.google.com/run?utm_campaign=CDR_0x72884f69_default_b421852297&utm_medium=external&utm_source=blog)）。
*   **stdio:** 標準入力と出力を使用するよりシンプルなトランスポートで、マシン上でローカルプロセスとしてツールを実行するのに最適です。

`stdio`トランスポートを使用すると、エージェントとサーバーは一連のJSON-RPCメッセージを交換します。プロセスは、接続を確立するための重要な3ウェイハンドシェイクから始まります。このハンドシェイクが完了した後にのみ、クライアントはツール呼び出しを開始できます。

シーケンスは次のようになります。

![MCP stdioシーケンス図](images/mcp-stdio-sequence-diagram.png)
*<p align="center">図1：公式MCPドキュメント（2025-06-18）からのstdioトランスポートシーケンス図。</p>*

公式仕様に基づくと、その最初のハンドシェイクのJSONメッセージは次のようになります。

**1. クライアント→サーバー：`initialize`リクエスト**
クライアントは、`initialize`リクエストを送信して会話を開始します。
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-06-18",
    "clientInfo": {
      "name": "Gemini CLI",
      "version": "1.0.0"
    }
  }
}
```

**2. サーバー→クライアント：`initialize`結果**
サーバーは、プロトコルバージョンを確認し、その機能と情報をアドバタイズする結果で応答します。これは、`godoctor`バイナリからの実際の応答です。
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "capabilities": {
      "completions": {},
      "logging": {},
      "tools": {
        "listChanged": true
      }
    },
    "protocolVersion": "2025-06-18",
    "serverInfo": {
      "name": "godoctor",
      "version": "0.2.0"
    }
  }
}
```

**3. クライアント→サーバー：`initialized`通知**
最後に、クライアントは`initialized`通知を送信してセットアップが完了したことを確認します。これは通知であるため、`id`フィールドがなく、メソッドは名前空間化されていることに注意してください。
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized",
  "params": {}
}
```

この交換が完了すると、セッションが確立され、クライアントはツール呼び出しに進むことができます。たとえば、サーバーに利用可能なツールのリストを要求するには、`tools/list`リクエストを送信できます。重要なのは、このリクエストの前に3つのハンドシェイクメッセージすべてが正しい順序で送信される必要があるということです。

このシェルスクリプトで完全なシーケンスを実際に確認できます。
```bash
#!/bin/bash
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"Manual Test Client","version":"1.0.0"}}}';
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | godoctor
```

このスクリプトを`godoctor`バイナリにパイプすると、最初に`initialize`の結果が出力され、次に`tools/list`の結果が出力され、GoDoctorのすべてのツールが正しくリストされます。この必要な3ステップのフローを理解することが、次のセクションで説明する私の最大の最初の障害を解決するための鍵でした。

MCPを初めて使用する人には、公式ドキュメントを読むことを強くお勧めします。私にとって最も重要だった2つのドキュメントは、[クライアント/サーバーのライフサイクル](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle)と[トランスポート層](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports)に関するページでした。（あるいは、面倒なら、それらのURLをCLIに渡して、あなたのために読ませてください=^_^=）

### 最初のブレークスルー：ドキュメントを読むエージェント

私の最初の目標は、APIの幻覚問題を解決することでした。多くの最初のプロンプトの試みと同様に、私の最初のリクエストはシンプルで少し曖昧でした。

> 「`godoc`という名前のツールを1つ持つGoのMCPサーバーを作成してください。このツールは、パッケージ名とオプションのシンボル名を受け取り、`go doc`コマンドを実行する必要があります。」

結果はあまり良くありませんでした。エージェントは、どのツールを使用し、どのプロトコルが最適かを判断するのに多くの時間を費やしました。頭字語「MCP」でさえ、それにとって明白な概念ではありませんでした。私が「Model Context Protocol」を意味すると明らかにするまで、しばしばそれが異なるものを表すと推測しました。Google検索とWebFetchを呼び出すサイクルにはまり、さまざまなSDKを試し、動作する例を作成できずに、別のSDKにピボットすることを何度も繰り返しました。ここからが「バイブコーディング」の本当の作業の始まりです。それは、指示を洗練させる反復的なプロセスです。数時間の試行錯誤の末、私ははるかに効果的なプロンプトにたどり着きました。具体的で高品質なリソースを提供することが鍵であることを学びました。

これが最終的な、はるかに改善されたプロンプトです。

> あなたのタスクは、go docコマンドを公開するModel Context Protocol（MCP）サーバーを作成し、LLMにGoのドキュメントをクエリする機能を提供することです。ツール名はgo-docとし、2つの引数、package_path（必須）とsymbol_name（オプション）を取る必要があります。ドキュメント部分には、`go doc`シェルコマンドを使用してください。MCP実装には、公式のGo SDK for MCPを使用し、stdioトランスポートを介して通信する本番環境対応のMCPサーバーを作成してください。また、サーバーをテストできるように、シンプルなCLIクライアントも作成してください。
>
> コードを書く前に、これらのリファレンスを読んで、テクノロジーとプロジェクトの構造に関する情報を収集してください。
> - https://github.com/modelcontextprotocol/go-sdk
> - https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle
> - https://go.dev/doc/modules/layout

このプロンプトは、いくつかの理由で優れています。使用する正確なSDKを指定し、トランスポート（`stdio`）を指示し、そして最も重要なことに、エージェントに読書リストを与えます。SDKのソースコードとMCP仕様へのリンクを提供することで、エージェントが幻覚を起こす傾向を劇的に減らしました。

このより良いプロンプトを使っても、旅は順調ではありませんでした。最大のハードルは、シンプルな`stdio`トランスポートを使用しようとしたときに現れました。私のツール呼び出しは、`サーバーの初期化が完了していません`という不可解なエラーで一貫して失敗しました。多くのつらいデバッグの後、問題はサーバーコードにはまったくないことを発見しました。問題は、MCP `stdio`トランスポートが、上で詳述した特定の3ステップのハンドシェイクを必要とすることでした。私のクライアントは、ハンドシェイクが完了する前にツールを呼び出そうとしていました。この経験は、私に貴重な教訓を教えてくれました。AI用のツールを構築するときは、コードをデバッグするだけでなく、会話プロトコル自体をデバッグしているのです。

サーバーが実行され、プロトコルを正しく話すようになったら、パズルの次のピースは、それをGemini CLIに導入することでした。これは、プロジェクトのルートにある`.gemini/settings.json`ファイルによって処理され、CLIにどのツールをロードするかを指示します。私はそれに次の設定を追加しました。

```json
{
  "mcpServers": {
    "godoctor": {
      "command": "./bin/godoctor"
    }
  }
}
```

これを設定すると、このディレクトリでGemini CLIを起動するたびに、バックグラウンドで`godoctor`サーバーが自動的に起動し、そのツールがエージェントで利用できるようになります。

### AIコードレビュアーとのフィードバックループの作成

動作する`godoc`ツールができたので、次の論理的なステップは、エージェントにコードについて読むだけでなく、その品質について推論するように教えることでした。これが`code_review`ツールにつながりました。今回は、すでに行った作業の直接の結果として、経験はずっとスムーズでした。

私のプロンプトは、実装ではなく目標に焦点を当てていました。

> 私のプロジェクトにcode_reviewという新しいツールを追加したいです。このツールは、Gemini APIを使用してGoコードを分析し、Goコミュニティで受け入れられているベストプラクティスに従って、json形式で改善点のリストを提供する必要があります。このツールは、Goコードのコンテンツとオプションのヒントを入力として受け取る必要があります...
>
> Geminiを呼び出すには、このSDKを使用してください：https://github.com/googleapis/go-genai

エージェントはまだGoの`genai` SDKを学ぶ必要がありましたが、今回はツールボックスに`godoc`ツールがありました。SDKを調べ、自分の間違いを修正し、リアルタイムで学習するためにツールを使用しているのを見ることができました。プロセスはまだ反復的でしたが、大幅に高速で効率的でした。

最も重要な結果は、ツール自体だけでなく、それが解き放った新しい機能でした。初めて、ツールを使用して独自のコードを確認し、AI駆動開発の別のレベルを解き放つことができました。**私は正のフィードバックループを作成しました。**

私のワークフローには、強力な新しいステップが加わりました。エージェントが新しいコードを生成した後、すぐに自分の作業を批評するように依頼できるようになりました。

> 「今、あなたが書いたコードで`code_review`ツールを使用して、提案を適用してください。」

エージェントは、独自の出力を分析し、AIが生成したフィードバックに基づいてリファクタリングします。これが、AI用のツールを構築する真の力です。タスクを自動化するだけでなく、自己改善のためのシステムを作成しているのです。

### 最終章：クラウドへのデプロイ

{{< warning >}}
独自のMCPサーバーをCloud Runにデプロイする場合は、適切な認証が設定されていることを確認してください。**公開アクセス可能なサーバーをデプロイしないでください**。特に、Gemini APIキーを使用している場合はそうです。公開エンドポイントは悪意のある攻撃者によって悪用される可能性があり、非常に高額で予期しないクラウド請求につながる可能性があります。
{{< /warning >}}

`stdio`を介して実行されるローカルツールは個人使用には最適ですが、Model Context Protocolの真の目的は、共有され、発見可能なツールのエコシステムを作成することです。この「サイドクエスト」の次のフェーズは、GoDoctorをラップトップのローカルバイナリから、[Google Cloud Run](https://cloud.google.com/run?utm_campaign=CDR_0x72884f69_default_b421852297&utm_medium=external&utm_source=blog)を使用してスケーラブルなWebサービスにすることでした。

これは、エージェントにクラウド開発のための2つの新しいスキル、つまりアプリケーションをコンテナ化する方法とデプロイする方法を教えることを意味しました。

まず、シンプルな`stdio`トランスポートから`HTTP`に切り替える必要がありました。私のプロンプトは、以前の作業に基づいて直接的なものでした。

> 「サーバーをWebに公開する時が来ました。MCPサーバーを`stdio`トランスポートから`streamable HTTP`トランスポートを使用するようにリファクタリングしてください。」

サーバーがHTTPを話すようになったので、次のステップはクラウド用にパッケージ化することでした。私はエージェントに、軽量で安全なコンテナイメージを構築する標準的な方法である、本番環境対応のマルチステージ`Dockerfile`を作成するように依頼しました。

> 「Goバイナリをコンパイルし、それを最小限の`golang:1.24-alpine`イメージにコピーするマルチステージDockerfileを作成してください。」

`Dockerfile`が定義されたので、デプロイの時間です。これは、ローカルの概念実証が実際のクラウドインフラストラクチャの一部になる瞬間です。

> 「今、このイメージをCloud Runにデプロイしてください。`us-central1`にデプロイし、現在環境で設定されているプロジェクトを使用してください。完了したら、MCPツールを呼び出すために使用できるURLを教えてください。」

エージェントは正しい`gcloud`コマンドを提供し、数分後、GoDoctorはインターネット上で稼働していました。セットアップを完了するには、ローカルのGemini CLIにリモートサーバーについて通知する必要がありました。これは、`.gemini/settings.json`ファイルを更新し、ローカルの`command`をリモートの`httpUrl`に交換することを意味しました。

```json
{
  "mcpServers": {
    "godoctor": {
      "httpUrl": "https://<your-cloud-run-url>.run.app"
    }
  }
}
```

そして、ちょうどそのように、私のCLIはデプロイされ、クラウドで実行されているツールを使用していました。これは、概念実証が本当に完了したと感じた瞬間でした。エージェントを導く反復的なプロセスは報われ、シンプルなアイデアを現代のアプリケーションのライフサイクル全体、つまりローカルの概念からスケーラブルなクラウドネイティブサービスへと導きました。

とはいえ、日常業務では、現在`stdio`バージョンを使用しています。1人のチームにとって、Cloud Runへのデプロイは単にやり過ぎでした。

### GoDoctorでのバイブコーディングからの私の重要な学び

この旅は、コードを書くことよりも、AIと効果的に協力する方法を学ぶことでした。私の最大の教訓は、考え方を「コーダー」から「教師」または「パイロット」に変えることでした。私が学んだ最も重要な教訓のいくつかを次に示します。

*   **あなたがパイロットです。** AIは強力な協力者ですが、それでもツールです。同意できないアクションを提案した場合は、ためらわずに`ESC`を押してキャンセルし、正しい方向に導くための新しいプロンプトを提供してください。
*   **ちょっとしたリマインダーが大きな違いを生みます。** 長い会話では、AIは以前の指示を見失うことがあります。元のプロンプト全体を繰り返す代わりに、記憶を呼び覚ますための短く的を絞ったリマインダーで十分なことがよくあります（例：「`stdio`トランスポートを使用することを忘れないでください」）。
*   **AIをループに入れておきます。** エージェントにすべての作業を実行させるのが最善ですが、手動での編集が必要になる場合があります。自分でコードを変更すると、AIのコンテキストが古くなります。コードベースと同期できるように、変更した内容を必ず伝えてください。
*   **他に何も機能しない場合は、オフにしてから再度オンにします。** （それほど）まれなケースではありませんが、AIが完全に動かなくなった場合、古典的なITソリューションが驚くほどうまく機能します。Gemini CLIの場合、これは`/compress`コマンドを使用して会話履歴を要約するか、極端な場合はCLIを再起動してクリーンなコンテキストで開始することを意味します。

エージェントに適切なコンテキストと適切なツールを与えることで、はるかに有能なパートナーになりました。旅は、私が単にコードを書いてもらおうとすることから、学習して自己改善できるシステムを構築することへと変わりました。

### 次は何ですか？

GoDoctorとの旅はまだ終わっていません。まだ実験的なプロジェクトであり、新しいツールや新しいインタラクションのたびに、私はさらに学んでいます。私の目標は、世界中のGo開発者にとって真に役立つコーディングアシスタントに進化させ続けることです。

これらの概念が公式のGoツールチェーンでどのように適用されているかに関心のある方は、多くの同じ目的を共有する`gopls` MCPサーバーについて読むことを強くお勧めします。[公式のGoドキュメントWebサイト](https://tip.golang.org/gopls/features/mcp)で詳細情報を見つけることができます。

### リソースとリンク

この記事全体で言及した主要なリソースのいくつかを次に示します。これらが私にとってそうであったように、あなたにとっても役立つことを願っています。

*   **[GoDoctorプロジェクトリポジトリ](https://github.com/danicat/godoctor):** 議論したツールの完全なソースコード。
*   **[Model Context Protocolホームページ](https://modelcontextprotocol.io/):** MCPについて学ぶための最良の出発点。
*   **[MCP仕様（2025-06-18）](https://modelcontextprotocol.io/specification/2025-06-18):** 完全な技術仕様。
*   **[MCPライフサイクルドキュメント](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle):** クライアント/サーバーのハンドシェイクを理解するための重要な読み物。
*   **[MCPトランスポートドキュメント](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports):** `stdio`トランスポートと`http`トランスポートの違いを理解するために不可欠です。