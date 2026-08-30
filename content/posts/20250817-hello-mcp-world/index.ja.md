---
categories:
- Agent Development
date: 2025-08-17 15:00:00+00:00
summary: GopherCon UK 2025 での基調講演をベースに、Model Context Protocol（MCP）の基本概念、アーキテクチャ、そして Go で AI 対応アプリケーションを構築するためのビルディングブロック（Tools、Prompts、Resources）を詳しく解説します。
tags:
  - gemini
  - golang
  - keynote
  - mcp
title: "Hello, MCP World!：Goで始めるModel Context Protocol入門"
slug: "hello-mcp-world"
aliases:
  - "/ja/posts/20250817-hello-mcp-world/"
description: "GopherCon UK 2025 の登壇に基づく Model Context Protocol（MCP）の入門解説。Host、Client、Server のアーキテクチャと Tool/Resource/Prompt を解説。"
proficiencyLevel: "Beginner"
dependencies:
  - "Go 1.24+"
  - "MCP Go SDK"
---

## はじめに

この記事では、Anthropic が開発した、大規模言語モデル（LLM）とアプリケーション間の通信を標準化するプロトコル「Model Context Protocol（MCP）」について詳しく解説します。本稿は、先週 [GopherCon UK で発表した同名の基調講演（Keynote）](https://speakerdeck.com/danicat/hello-mcp-world) をベースに構成しています。

理解を深めるため、まずは基礎からスタートし、主要なアーキテクチャコンポーネント、トランスポート、そして 3 つのビルディングブロック（Tools、Prompts、Resources）について順を追って説明します。道中では、私が過去に作成したサーバー（`godoctor` や `speedgrapher`）の実践的なコード例も紹介します。最後に、Gemini CLI を使ったシンプルな「バイブコーディング（vibe coding）」の実例を通して、Go SDK for MCP を用いた独自サーバーの作成手順を見ていきます。

MCP について初めて耳にする方から、すでに自作サーバーをいくつか開発した経験のある方まで、さまざまなスキルレベルの開発者にとって役立つ知見をお届けします。

## 新しい標準の誕生

標準（スタンダード）の話になると、いつも XKCD のこのコミックが真っ先に頭に浮かびます。

![Standards](image.png)
*出典: [xkcd.com](https://xkcd.com/927)*

面白いことに、この業界においてこのジョークが完全には当てはまらない稀有な事例が、今回の MCP かもしれません（少なくとも今のところは）。幸いなことに、業界全体が LLM にコンテキストを渡す標準として、急速に MCP へと収束していきました。

公式の仕様書では、MCP は次のように説明されています。

> MCP は、アプリケーションが大規模言語モデル（LLM）にコンテキストを提供する方法を標準化するオープンプロトコルです。MCP は、いわば AI アプリケーションにおける USB-C ポートのようなものだと考えてください。USB-C がデバイスとさまざまな周辺機器やアクセサリを接続する標準規格となっているのと同様に、MCP は AI モデルを多様なデータソースやツールと接続するための標準的な手段を提供します。MCP を使うことで、LLM 上にエージェントや複雑なワークフローを構築し、モデルを世界とつなぐことができます。

USB-C の例えもよく分かりますが、個人的には MCP を「新たな HTTP/REST」と捉える方がしっくりきます。HTTP が Web サービス同士が通信するための共通言語となったように、MCP は AI モデルが外部システムとやり取りするための共通フレームワークを提供します。エンジニアとして、私たちは過去 20 年ほどをかけてあらゆるものを「API ファースト」にし、ソフトウェアシステムを相互接続して新たな次元の自動化を実現してきました。この先 20 年とまではいかなくても、今後 5〜10 年の間、既存のシステムを AI 対応へと改修し（そして新たなシステムを立ち上げ）、莫大なエンジニアリングリソースを投じることになるはずです。そして MCP こそが、そのプロセスの要となるコンポーネントなのです。

## MCP アーキテクチャ

以下の図を見ると、MCP のアーキテクチャは一見複雑そうに見えるかもしれません。

![MCP Architecture](image-1.png)
*出典: [MCP Specification](https://modelcontextprotocol.io/docs/learn/architecture)*

MCP アーキテクチャの主要コンポーネントは以下のとおりです。

*   **MCP Host（ホスト）:** IDE やコーディングエージェントなどのメインとなる AI アプリケーション。
*   **MCP Server（サーバー）:** 特定の機能（ツールやプロンプトなど）へのアクセスを提供するプロセス。
*   **MCP Client（クライアント）:** ホストを個々のサーバーに接続するクライアント。

要するに、ホストアプリケーションが複数のクライアントを生成・管理し、各クライアントが特定のサーバーと 1:1 の関係で接続します。

## MCP レイヤー

通信は 2 つのレイヤーで構成されています。

* **データレイヤー（Data layer）**: JSON-RPC ベースのプロトコルです。メッセージ形式の具体例は次のセクションで紹介します。
* **トランスポートレイヤー（Transport layer）**: 通信チャネルを定義します。主要なものは以下のとおりです。
  - **Standard I/O（標準入出力 / stdio）**: ローカルサーバー向け
  - **Streamable HTTPS**: ネットワーク越しの通信向け（HTTPS+SSE を置き換えるもの）
  - **HTTPS+SSE**: セキュリティ上の懸念により、最新の仕様では非推奨（deprecated）

データレイヤーは SDK 側で管理されるため、テスト目的を除けば開発者が手動でメッセージを組み立てる必要はありません。トランスポートの選択はユースケースによりますが、基本的にはまず stdio から始めて、必要に応じて後から HTTPS を追加するのがおすすめです。stdio の MCP サーバーを HTTPS に（あるいはその逆に）変換するオープンソースのアダプターも存在しますが、自作サーバーに HTTPS を組み込むこと自体非常に簡単なので、そうしたアダプターを使うのは自分でソースコードを管理できないサーバーの場合くらいでしょう。

## 初期化フロー

クライアントとサーバーは、接続確立時にハンドシェイクを行います。これには 3 つの主要なメッセージが関わります。

1. クライアントがサポートするプロトコルバージョンを指定して、サーバーへ `initialize` リクエストを送信する（サーバーはクライアントへ初期化レスポンスを返します）。
2. クライアントが `notifications/initialized` 通知を送信して初期化を確定する。
3. これにより、クライアントは `tools/list` などのリクエストを送信し、サーバーが提供する機能を取得できるようになります。

クライアント側から実際にやり取りされる初期化フロー（JSON-RPC 形式）は次のようになります。

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}
{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
```

いきなり `tools/list` や `tools/call` メッセージを直接送信することはできない点に注意してください。ハンドシェイクを経ずに送信すると、「server not ready」といったエラーが返ってきます。

Gemini CLI などのコーディングエージェントを使って MCP サーバーを開発しているとき、私はよくシェル経由で以下のようにメッセージを流し込んでテストするようエージェントに指示します。

```sh
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}';
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | ./bin/godoctor
```

このフローを完全に把握する前は、コーディングエージェントが「サーバーの起動に時間がかかっているようなので、tool call の前に sleep を挟みます」といった見当違いな推測をしてしまうことがよくありました。そのため、実装が正しく動いているかを確実に検証するためにこの方法を好んで使っています。開発中の MCP サーバーと正しく通信する作法をエージェントへいち早く教え込むことが、開発をスムーズに進める秘訣です！

## MCP サーバーのビルディングブロック

MCP サーバーの機能は、3 つの基本的なビルディングブロック（「プリミティブ」や「サーバーコンセプト」とも呼ばれます）を通じて公開されます。

| ビルディングブロック | 目的                  | 制御主体         | 実世界の例                               |
| :------------- | :----------------------- | :---------------------- | :----------------------------------------------- |
| **Tools（ツール）**      | AI のアクション実行           | モデル主導（Model-controlled）        | フライト検索、メッセージ送信、コードレビュー       |
| **Resources（リソース）**  | コンテキストデータの提供         | アプリケーション主導（Application-controlled）  | ドキュメント、カレンダー、メール、気象データ       |
| **Prompts（プロンプト）**    | 対話テンプレート| ユーザー主導（User-controlled）         | 「旅行の計画を立てる」「ミーティングを要約する」       |

それぞれを詳しく見ていきましょう。

### Tools（ツール）

Tools は、AI モデルに特定のアクションを実行させるための関数です。たとえば、API やデータベース、コマンドラインツールなどを外部に公開します。

Tools の概念を実験するために私が作成したサーバーが `GoDoctor` です。これは、LLM が Go コードを書く際の支援能力を高めるためのツール群を提供します。`GoDoctor` という名前は、Go パッケージのドキュメントを表示する CLI ツール `go doc` にちなんだ言葉遊びです。

私の仮説は、「正確な公式ドキュメントを提供すれば、LLM のハルシネーション（幻覚）が減り、より品質の高いコードを書けるようになるのではないか」というものでした。あるいは少なくとも、自らの間違いを学習し、自己修正するための手掛かりが得られるはずだと考えました。

ツールの実装は、MCP サーバーへのツールの登録と、ハンドラーの実装という 2 つの主要コンポーネントで構成されます。

登録には `mcp.AddTool` 関数を使用します。

{{< github user="danicat" repo="godoctor" path="internal/tools/get_documentation/get_documentation.go" lang="golang" start="35" end="40" >}}

ハンドラーは、API やコマンド、関数を呼び出して、プロトコルと互換性のある形式（`mcp.CallToolResult` 構造体）でレスポンスを返すアダプターの役割を果たします。

`GoDoctor` のドキュメント取得ツールのハンドラー実装は以下のようになっています。

{{< github user="danicat" repo="godoctor" path="internal/tools/get_documentation/get_documentation.go" lang="golang" start="49" end="86" >}}

### Prompts（プロンプト）

Prompts は、パラメータを受け取ることができる再利用可能なユーザー主導のテンプレートを提供します。AI エージェント内ではスラッシュコマンドとして表示されることが多く、ユーザーはシンプルなコマンドひとつで複雑なワークフローをトリガーできます。

実際の動作例として、私が作成したもうひとつの MCP サーバー `speedgrapher` を見てみましょう。これはテクニカルライティングを支援するためのプロンプトとツールを集めたものです。

`speedgrapher` の中でも最もシンプルなプロンプトのひとつが `/haiku` です。Tools と同様に、プロンプトの定義とそのハンドラーの実装を行います。

{{< github user="danicat" repo="speedgrapher" path="internal/prompts/haiku.go" lang="golang" start="24" end="54" >}}

### Resources（リソース）

Resources はファイル、API、データベースなどのデータを公開し、AI がタスクを遂行するために必要なコンテキストを提供します。概念としては、**Tool が「アクションを実行するもの」**であるのに対し、**Resource は「情報を提供するもの」**です。

とはいえ、現実の開発シーンでは、ほとんどの開発者が（通常の API で GET リクエストを使うように）データの公開にも Tools を使っているため、Resources の優れた実装例にはまだあまりお目にかかれていません。これは仕様策定側が少し凝りすぎてしまった部分かもしれませんが、今後コミュニティが MCP に慣れていくにつれて、Resources の有効な活用パターンが見出されていくでしょう。

## クライアントコンセプト（Client Concepts）

サーバー側のビルディングブロックに加えて、MCP 仕様にはサーバーがクライアントに対して要求できる機能である **Client Concepts** も定義されています。具体的には以下のような機能があります。

*   **Sampling（サンプリング）:** サーバー側からクライアントのモデルへ LLM のテキスト生成（補完）を要求できるようにする機能。サーバー作成者がモデル呼び出しのために自前の API キーを用意する必要がなくなるため、セキュリティやコスト管理（課金）の観点から非常に有望です。
*   **Roots（ルート境界）:** クライアントがファイルシステムの操作境界を伝え、サーバーに対して操作が許可されているディレクトリを明示するメカニズム。
*   **Elicitation（入力要求）:** サーバーがユーザーに対して特定の情報を構造化された形式で要求し、必要な入力を得るまで処理を一時停止する仕組み。

これもまた、私が調べた限りでは（サーバー・クライアント双方ともに）現実の実装が仕様に追いついていない領域のひとつです。これらの機能が広く普及するにはもう少し時間がかかるかもしれません。これぞ最先端（ブリーディングエッジ）技術を扱う醍醐味でもあり、苦労でもあります……。たとえば、Gemini CLI に Roots サポートが追加されたのも、つい先週のことでした（[PR #5856](https://github.com/google-gemini/gemini-cli/pull/5856)）。

## ライブデモ: MCP サーバーのバイブコーディング

お気に入りのコーディングエージェントに渡して「Hello World」サーバーを作らせるプロンプトの例を紹介します。現在のエージェントは非決定的（non-deterministic）なため、1 回で 100% 完璧にいかない場合もあります。その際は追加のプロンプトで微調整が必要になることもありますが、最初の足がかりとしては最適です。

```text
Your task is to create a Model Context Protocol (MCP) server to expose a "hello world" tool. For the MCP implementation, you should use the official Go SDK for MCP and use the stdio transport.

Read these references to gather information about the technology and project structure before writing any code:
- https://raw.githubusercontent.com/modelcontextprotocol/go-sdk/refs/heads/main/README.md
- https://go.dev/doc/modules/layout

To test the server, use shell commands like these:
`( 
	echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}';
	echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}';
	echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}';
) | ./bin/hello`
```

エージェントがサーバーの作成に成功したら、新しく作ったツールに対して `tools/call` メソッドを実行させ、結果を確認してみましょう！

## 今後の展望

Go コミュニティは MCP エコシステムへ積極的に投資を行っています。特に注目すべき 2 つのプロジェクトを紹介します。

*   **Go SDK for MCP:** デモで使用した公式 SDK で、Google と Anthropic のパートナーシップによって開発されています。現時点ではまだ実験的（執筆時点のバージョンは v0.20）ですが、十分に実用的であり活発に開発が進められています。[github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) で公開されています。
*   **`gopls` の MCP サポート:** Go 公式の Language Server である `gopls` に MCP サポートを追加し、AI モデルによる Go のコーディング支援能力を大幅に強化する取り組みが進んでいます。プロジェクトはまだ初期段階ですが、[tip.golang.org/gopls/features/mcp](https://tip.golang.org/gopls/features/mcp) で進捗を追うことができます。

## 注目の MCP サーバー

コミュニティによって構築された注目すべきサーバーをいくつか紹介します。

*   **Playwright:** Microsoft が開発・メンテナンスを行っているサーバーで、AI エージェントが Web ページを操作（ナビゲーション）したり、スクリーンショットを撮影したり、ブラウザ操作を自動化できるようにします（[microsoft/playwright-mcp](https://github.com/microsoft/playwright-mcp)）。
*   **Context7:** GoDoctor と同様に、モデルに公式ドキュメントを提供することでハルシネーションを抑制し、回答精度を高めるサーバーです。クラウドソーシングされたリポジトリからドキュメントを取得します。詳細は [context7.com](https://context7.com/) をご覧ください。

## 独自のサーバーを作ってみよう

Model Context Protocol は、AI エージェントの機能を拡張するための標準化された手段を提供します。独自の MCP サーバーを構築すれば、普段の開発ワークフローにぴったり合わせた、コンテキストを深く理解する専用アシスタントを作り出すことができます。

実際に手を動かしてみたい方のために、独自の MCP サーバーをゼロから構築するステップを解説した Google Codelab を作成しました。

[**How to Build a Coding Assistant with Gemini CLI, MCP and Go（Gemini CLI、MCP、Go でコーディングアシスタントを構築する）**](https://codelabs.developers.google.com/codelabs/gemini-cli-mcp-go)

## おわりに

最後までお読みいただきありがとうございました！ご質問や感想などがあれば、ページ下部のコメント欄や各種 SNS でぜひお気軽にお知らせください。