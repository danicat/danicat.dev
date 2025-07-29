---
title: "Gemini CLIとGoでMCPサーバーを構築する方法"
date: 2025-07-29
author: "Daniela Petruzalek"
tags: ["go", "gemini-cli", "mcp", "ai", "codelab"]
categories: ["Tutorials"]
summary: "Gemini CLIとModel Context Protocol（MCP）を使用して、AI搭載のGo開発アシスタントであるGoDoctorを構築したステップバイステップガイド。"
---

{{< translation-notice >}}

### はじめに

多くの皆さんと同様に、私もAI支援開発に深く没頭してきました。その道のりは、しばしば「すごい」と思う瞬間とイライラする障害のジェットコースターです。これは、そんな旅の物語の一つです。あるものを構築しようとして、イライラする問題で完全に脇道に逸れ、最終的には私のAI支援ワークフローを根本的に改善するツールを手に入れた物語です。

私の当初の目標は、[osquery](https://www.osquery.io/)（SQLを使用してマシンの状態をクエリできるツール）用のModel Context Protocol（MCP）サーバーを構築することでした。Gemini CLIを使ってGoのコードを書くのを手伝ってもらうことに興奮していました。しかし、すぐに壁にぶつかりました。エージェントが生成するGoのコードは、しばしば慣用的ではありませんでした。初心者のような間違いを犯し、過剰な抽象化を行い、存在しないAPIを頻繁に「幻覚」しました。私の推測では、基礎となるモデルが新しい[Go SDK for MCP](https://github.com/modelcontextprotocol/go-sdk)でトレーニングされていなかったため、知らないと認めるよりも答えをでっち上げることを好んだのでしょう。

この経験から、私は重要な気づきを得ました。ツールと戦うのではなく、ツールに*教える*ことができるということです。私はosqueryプロジェクトを一時中断し、「サイドクエスト」に乗り出すことにしました。それは、Go開発の専門家であることを唯一の目的とする専用のMCPサーバーを構築することです。このサイドプロジェクトは、最終的に[GoDoctor](https://github.com/danicat/godoctor)と名付けられ、Gemini CLIがより良いGoコードを書くために必要なツールを提供することになります。

この記事では、GoDoctorを構築した物語を皆さんにご紹介します。これは従来のチュートリアルというよりは、「プロンプト駆動」の旅です。プロジェクトの要件を効果的なプロンプトに変換し、実装の詳細を通じてAIを導き、その過程で避けられない間違いから学ぶことに焦点を当てます。

### 舞台設定：`GEMINI.md`

コードを一行も書く前に、最初のステップは基本ルールを設定することでした。`GEMINI.md`はGemini CLIに特有のファイルですが、コンテキストファイルを作成する習慣は多くのAIコーディングエージェントで標準的です（例えば、Julesは`AGENTS.md`を、Claudeは`CLAUDE.md`を使用します）。実際、これを[`AGENT.md`](https://ampcode.com/AGENT.md)というファイルで標準化しようという新たな動きもあります。このファイルは、AIにプロジェクトの標準やその振る舞いに対する期待を基礎的に理解させる上で非常に重要です。

このプロジェクトは全くの新品だったので、共有すべき特定のアーキテクチャの詳細はまだありませんでした。そのため、私は高品質で慣用的なGoコードを作成することに焦点を当てた一般的なガイドラインから始めました。プロジェクトが進化するにつれて、プロジェクトの構造、ビルドコマンド、または主要なライブラリに関するより具体的な指示を追加するのが一般的です。よりプロジェクト固有のファイルの例として、私の[`testquery`プロジェクト](https://github.com/danicat/testquery/blob/main/GEMINI.md)で使用している`GEMINI.md`を見ることができます。

これが、この旅でAIの憲法として機能した最初の`GEMINI.md`です。

```markdown
# Go Development Guidelines
All code contributed to this project must adhere to the following principles.

### 1. Formatting
All Go code **must** be formatted with `gofmt` before being submitted.

### 2. Naming Conventions
- **Packages:** Use short, concise, all-lowercase names.
- **Variables, Functions, and Methods:** Use `camelCase` for unexported identifiers and `PascalCase` for exported identifiers.
- **Interfaces:** Name interfaces for what they do (e.g., `io.Reader`), not with a prefix like `I`.

### 3. Error Handling
- Errors are values. Do not discard them.
- Handle errors explicitly using the `if err != nil` pattern.
- Provide context to errors using `fmt.Errorf("context: %w", err)`.

### 4. Simplicity and Clarity
- "Clear is better than clever." Write code that is easy to understand.
- Avoid unnecessary complexity and abstractions.
- Prefer returning concrete types, not interfaces.

### 5. Documentation
- All exported identifiers (`PascalCase`) **must** have a doc comment.
- Comments should explain the *why*, not the *what*.

# Agent Guidelines
- **Reading URLs:** ALWAYS read URLs provided by the user. They are not optional.
```

このファイルは、最初から品質とスタイルのベースラインを確立します。

### Model Context Protocol（MCP）を理解する

このプロジェクトの中心にあるのがModel Context Protocol（MCP）です。これを「LLMツール用のUSB標準」と表現する人もいますが、私は別の見方をしています。**HTTPとRESTがWeb APIの標準化に果たした役割を、MCPはLLMツールに対して果たしているのです。** RESTが登場する前、Webサービスが通信するための普遍的に受け入れられた単一の方法はありませんでした。MCPは、AIエージェントの世界で同じ問題を解決しています。これはJSON-RPCベースのプロトコルで、共通の基盤を作り出し、「MCPを話す」エージェントが、カスタムの、一度きりの統合を必要とせずに、準拠したツールを発見して使用できるようにします。

このプロトコルは、エージェントとツールサーバーが通信するためのさまざまな方法を定義しており、これらはトランスポートとして知られています。最も一般的な2つは次のとおりです。
*   **HTTP:** おなじみのリクエスト/レスポンスモデルで、リモートサービスとしてデプロイされるツール（例：Cloud Run）に最適です。
*   **stdio:** 標準入力と標準出力を使用するよりシンプルなトランスポートで、マシン上でローカルプロセスとしてツールを実行するのに最適です。

`stdio`トランスポートでは、エージェントとサーバーは一連のJSON-RPCメッセージを交換します。プロセスは、接続を確立するための重要な3ウェイハンドシェイクから始まります。このハンドシェイクが完了した後にのみ、クライアントはツール呼び出しを開始できます。

シーケンスは次のようになります。

![MCP stdio sequence diagram](images/mcp-stdio-sequence-diagram.png)
*<p align="center">図1：公式MCPドキュメント（2025-06-18）からのstdioトランスポートシーケンス図。</p>*

公式仕様に基づいた、その最初のハンドシェイクのJSONメッセージは次のようになります。

**1. クライアント → サーバー：`initialize`リクエスト**
クライアントは`initialize`リクエストを送信して会話を開始します。
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

**2. サーバー → クライアント：`initialize`リザルト**
サーバーは結果を返し、プロトコルバージョンを確認し、その機能と情報を通知します。これが`godoctor`バイナリからの実際の応答です。
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

**3. クライアント → サーバー：`initialized`ノーティフィケーション**
最後に、クライアントは`initialized`ノーティフィケーションを送信してセットアップが完了したことを確認します。これはノーティフィケーションなので`id`フィールドがなく、メソッドは名前空間化されています。
```json
{
  "jsonrpc": "2.0",
  "method": "notifications/initialized",
  "params": {}
}
```

このやり取りが完了するとセッションが確立され、クライアントはツール呼び出しに進むことができます。たとえば、サーバーに利用可能なツールのリストを尋ねるには、`tools/list`リクエストを送信できます。重要なのは、このリクエストの前に3つのハンドシェイクメッセージすべてが正しい順序で送信される必要があるということです。

このシェルスクリプトで完全なシーケンスを実際に確認できます。
```bash
#!/bin/bash
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"Manual Test Client","version":"1.0.0"}}}';
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | godoctor
```

このスクリプトを`godoctor`バイナリにパイプすると、最初に`initialize`の結果が出力され、次にGoDoctorのすべてのツールを正しくリストする`tools/list`の結果が出力されます。この必要な3ステップのフローを理解することが、次のセクションで説明する私の最初の最大の障害を解決する鍵でした。

MCPが初めての方は、公式ドキュメントを読むことを強くお勧めします。私にとって最も重要だった2つのドキュメントは、[クライアント/サーバーライフサイクル](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle)と[トランスポート層](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports)に関するページでした。（もし面倒なら、これらのURLをCLIに渡して読ませてみてください =^_^=）

### 最初のブレークスルー：ドキュメントを読むエージェント

私の最初の目標は、APIの幻覚問題を解決することでした。多くの最初のプロンプト試行と同様に、私の最初のリクエストは単純で少し曖昧でした。

> 「`godoc`という名前のツールを1つ持つGoのMCPサーバーを作成してください。このツールはパッケージ名とオプションのシンボル名を受け取り、`go doc`コマンドを実行する必要があります。」

結果は芳しくありませんでした。エージェントは多くの時間を費やして、どのツールを使用し、どのプロトコルが最適かを判断しようとしました。頭字語「MCP」でさえ、それが「Model Context Protocol」を意味することを私が明らかにするまで、しばしばそれが異なるものを指すと推測しました。Google SearchとWebFetchツールを呼び出すサイクルにはまり、さまざまなSDKを試し、動作する例を作成できずに、次から次へと別のSDKにピボットしました。ここからが「vibe coding」の本当の作業の始まりです。指示を洗練させる反復的なプロセスです。数時間の試行錯誤の末、私ははるかに効果的なプロンプトにたどり着きました。具体的で高品質なリソースを提供することが鍵であることを学びました。

これが最終的な、大幅に改善されたプロンプトです。

> あなたのタスクは、go docコマンドを公開するModel Context Protocol（MCP）サーバーを作成し、LLMにGoのドキュメントをクエリする機能を提供することです。ツール名はgo-docとし、2つの引数、package_path（必須）とsymbol_name（オプション）を取る必要があります。ドキュメント部分には、`go doc`シェルコマンドを使用してください。MCP実装には、公式のGo SDK for MCPを使用し、stdioトランスポートを介して通信する本番環境対応のMCPサーバーを作成してください。また、サーバーをテストするための簡単なCLIクライアントも作成してください。
>
> コードを書く前に、これらのリファレンスを読んで、テクノロジーとプロジェクトの構造に関する情報を収集してください。
> - https://github.com/modelcontextprotocol/go-sdk
> - https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle
> - https://go.dev/doc/modules/layout

このプロンプトが優れている理由はいくつかあります。使用するSDKを正確に指定し、トランスポート（`stdio`）を指示し、そして最も重要なことに、エージェントに読書リストを与えている点です。SDKのソースコードとMCP仕様へのリンクを提供することで、エージェントが幻覚を起こす傾向を劇的に減らしました。

この改善されたプロンプトを使っても、道のりは平坦ではありませんでした。最大の障害は、単純な`stdio`トランスポートを使おうとしたときに現れました。私のツール呼び出しは、`server initialisation is not complete`という不可解なエラーで一貫して失敗しました。多くのつらいデバッグの末、問題はサーバーコードにはまったくないことを発見しました。問題は、MCP `stdio`トランスポートが、上記で詳述した特定の3ステップのハンドシェイクを必要とすることでした。私のクライアントは、ハンドシェイクが完了する前にツールを呼び出そうとしていました。この経験から、AI用のツールを構築するときは、コードをデバッグするだけでなく、会話プロトコル自体をデバッグしているのだという貴重な教訓を学びました。

### フィードバックループの作成：AIコードレビュアー

`godoc`ツールが機能するようになったことで、次の論理的なステップは、エージェントにコードについて読むだけでなく、その品質について推論するように教えることでした。これが`code_review`ツールにつながりました。今回は、すでに行った作業の直接的な結果として、経験はずっとスムーズでした。

私のプロンプトは、実装ではなく目標に焦点を当てていました。

> 私のプロジェクトにcode_reviewという新しいツールを追加したいです。このツールは、Gemini APIを使用してGoコードを分析し、Goコミュニティで受け入れられているベストプラクティスに従って、json形式で改善点のリストを提供する必要があります。このツールは、Goファイルの内容とオプションのヒントを入力として受け取ります...
>
> Geminiを呼び出すには、このSDKを使用してください：https://github.com/googleapis/go-genai

エージェントはまだ`genai` Go SDKを学ぶ必要がありましたが、今回はツールボックスに`godoc`ツールがありました。SDKを調べ、自身の誤りを訂正し、リアルタイムで学習するためにツールを使用しているのを見ることができました。プロセスはまだ反復的でしたが、大幅に高速で効率的でした。

最も重要な成果は、ツール自体だけではなく、それが解き放った新しい能力でした。初めて、ツールを使って自身のコードをレビューし、AI駆動開発の別のレベルを解き放つことができました。**私はポジティブなフィードバックループを作成しました。**

私のワークフローには、今や強力な新しいステップがあります。エージェントが新しいコードを生成した後、すぐに自身の作業を批評するように頼むことができます。

> 「今書いたコードに対して`code_review`ツールを使い、提案を適用してください。」

エージェントは自身の出力を分析し、AIが生成したフィードバックに基づいてリファクタリングします。これこそがAI用ツールを構築する真の力です。タスクを自動化するだけでなく、自己改善のためのシステムを作成しているのです。

### 最終章：クラウドへのデプロイ

ローカルツールは素晴らしいですが、MCPの真の力は、スケーラブルなサービスとしてツールをデプロイすることにあります。プロジェクトの最終段階は、GoDoctorをコンテナ化し、Cloud Runにデプロイすることでした。

まず、エージェントにサーバーを`stdio`から`streamable HTTP`トランスポートにリファクタリングするように促しました。次に、本番環境対応のマルチステージ`Dockerfile`を作成するように依頼しました。

> Goバイナリをコンパイルし、golang:1.24-alpineのような最小限のgolangイメージにコピーするマルチステージDockerfileを作成してください。

最後に、デプロイの時間です。

> 今すぐこのイメージをCloud Runにデプロイし、MCPツールを呼び出すために使用できるURLを返してください。us-central1にデプロイし、現在環境で設定されているプロジェクトを使用してください。

エージェントは正しい`gcloud`コマンドを提供し、数分後、GoDoctorはインターネット上で稼働し、MCP準拠のクライアントからアクセスできるようになりました。

### GoDoctorのVibe Codingから得た重要な教訓

この旅は、コードを書くことよりも、AIと効果的に協力する方法を学ぶことでした。私の最大の教訓は、自分の考え方を「コーダー」から「教師」または「パイロット」に変えることでした。私が学んだ最も重要な教訓のいくつかを以下に示します。

*   **あなたがパイロットです。** AIは、あなたが同意しない行動を提案することがあります。`ESC`を押してキャンセルし、正しい方向に導くための新しいプロンプトを提供することを恐れないでください。
*   **繰り返すのではなく、思い出させてください。** 人間と同じように、AIも長い会話の中で詳細を忘れることがあります。指示を忘れた場合は、「stdioトランスポートを使用していることを忘れないでください」といった簡単なリマインダーで十分なことがよくあります。
*   **コーディングする必要がある場合は、AIに伝えてください。** すべての作業をAIに実行させるようにしてください。ただし、手動で変更を加えた場合は、AIにそのことを伝えて、コンテキストを更新できるようにしてください。
*   **困ったときは、再起動してください。** AIが動かなくなったまれなケースでは、`/compress`コマンド、あるいはクリーンなコンテキストでCLIを再起動するだけでも驚くほどうまくいくことがあります。

エージェントに適切なコンテキストと適切なツールを与えることで、はるかに有能なパートナーになりました。旅は、単にコードを書いてもらうことから、学習し自己改善できるシステムを構築することへと変わりました。

### 次は何？

GoDoctorとの旅はまだ終わっていません。これはまだ実験的なプロジェクトであり、新しいツールや新しいインタラクションのたびに、私はさらに多くを学んでいます。私の目標は、世界中のGo開発者にとって真に役立つコーディングアシスタントに進化させ続けることです。

これらの概念が公式のGoツールチェーンでどのように適用されているか興味がある方は、多くの同じ目的を共有する`gopls` MCPサーバーについて読むことを強くお勧めします。[公式のGoドキュメントウェブサイト](https://tip.golang.org/gopls/features/mcp)で詳細情報を見つけることができます。

### リソースとリンク

この記事で言及した主要なリソースのいくつかを以下に示します。これらが私にとってそうであったように、皆さんにとっても役立つことを願っています。

*   **[GoDoctorプロジェクトリポジトリ](https://github.com/danicat/godoctor):** 議論したツールの完全なソースコード。
*   **[Model Context Protocolホームページ](https://modelcontextprotocol.io/):** MCPについて学ぶための最良の出発点。
*   **[MCP仕様（2025-06-18）](https://modelcontextprotocol.io/specification/2025-06-18):** 完全な技術仕様。
*   **[MCPライフサイクルドキュメント](https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle):** クライアント/サーバーハンドシェイクを理解するための重要な読み物。
*   **[MCPトランスポートドキュメント](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports):** `stdio`と`http`トランスポートの違いを理解するために不可欠です。
