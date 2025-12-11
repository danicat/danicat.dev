---
title: "Taming Vibe Coding: エンジニアのためのガイド"
date: 2025-12-06T02:00:00Z
draft: false
categories: ["Workflow & Best Practices"]
tags: ["vibe-coding", "ai", "mcp", "gemini-cli", "jules"]
summary: "AIのスピードを、混乱なしに手に入れましょう。エンジニアリングの基本を適用して、構造化され、安全で、長持ちするコードを維持する方法を解説します。"
---

一年の終わりに、自分がやったこと、そしてやっておけばよかったことを振り返る時期がやってきました。今年は私にとって激動の一年でした。4月にGoogleに入社し、AIの世界に向けて自分自身をリファクタリングするノンストップのレースが始まりました。一年が終わろうとしている今、その努力は報われたと自信を持って言えます。私はより良いエンジニアになりました。

この記事では、私の「vibe coding」に対する理解がどのように進化したか、そして学んだ教訓を共有します。vibe codingは、非開発者が自然言語を使ってソフトウェアを構築する方法として見られがちですが、私の目標は、エンジニアリングの規律を加えることで、より良く、より一貫した結果が得られることを示すことです。

本来の定義である["(...) give in to the vibes (...) and forget that the code even exists"（バイブスに身を任せ……コードが存在することさえ忘れる）](https://x.com/karpathy/status/1886192184808149383?lang=en)を知っていますし、私の概念の一部はそれと矛盾するでしょう。私は決して「コードの存在を忘れる」ことはありません。しかし、この言葉はAI支援コーディングの同義語として進化してきました。このテキストでは、vibe codingを、エンジニアではなくLLMがコードの大部分を書くコーディングと定義しましょう。

## Motivation: なぜ vibe code なのか？

実践的な内容に入る前に、私がどのようなバックグラウンドを持っているか、少し共有しておきたいと思います。

私は20年以上ソフトウェアエンジニアをしており、何が良いコードかという強い感覚を養ってきました。「賢さ」よりも可読性と保守性を優先し、オーバーエンジニアリングを避け、薄いスライス（thin slices）と緊密なフィードバックループを重視することを学びました。

シニア開発者からプリンシパルエンジニアへと進化するにつれ、私の焦点はコードを書くことから有効性の管理へと移りました。エピックを書き、スコープを交渉し、チームのアウトプットを監督するのです。これは多くの人が直面するアイデンティティの危機です。PRに名前が載らなくなっても、あなたはまだエンジニアなのでしょうか？責任が増えるにつれて、会議に費やす日々が増え、「適切なエンジニアリング」をしていないように感じてしまいます。

この葛藤は、この分野のほとんどの人に遅かれ早かれ現れると思います。そして、次のような問いを投げかけます。エンジニアであることは、ただコードを書くことなのか？それとも、もっと別の何かなのか？

正直に言うと、私の最初のvibe codingの経験はひどいものでした。初期のChatGPTは期待外れのコードを生成し、私はそれに見切りをつけました。もう一度試してみようと思ったのは、2024年半ばのことでした。モデルは大幅に進化していました。初めてAIが私の考えもしなかったことを提案してくれ、それは客観的に見ても私のアプローチより優れていました。ついに、生成AI（GenAI）が役に立つと感じたのです。

私はプロトタイプや「ラバーダック」セッションのためにGenAIをツールボックスに加えました。それは私の中で成長していきました。逆に、手動でコードを書くことはそれほど刺激的ではなくなってきました。APIをいくつも書いていると、その新しさは薄れていきます。私たちはしばしば、新しいものを創造するのではなく、パターンを繰り返していることに気づきます。

そして、私が自分の人生の選択に疑問を持ち始めた矢先、Googleでの出来事がありました。

Geminiとエージェントについて話す責任を負い、私はスキルをアップグレードしました。LLM、[Gemini CLI](https://geminicli.com/)、そして[Jules](https://jules.google/)について深く掘り下げました。ほんの数ヶ月で、私は毎日AIを使ってコーディングするようになりました……しかし、最大の違いは、再び楽しんでいるということでした！

私にとって最大の改善点は、考える速度でコードをタイプすることはできませんが、アイデアが浮かんだらすぐにタイプ *できる* ことです。vibe codingをするとき、私はモデルを私の手の代わりとして使い、執筆を委任して、自分はソリューションに集中します。

## prompting スキルを磨く

vibe codingの世界で最初に開発する必要があるスキルはprompting（プロンプティング）です。ここ数ヶ月、親友のようにLLMとチャットすることから、罵倒したり指図したりすることまで、多くのアプローチを試しました。驚くことではないかもしれませんが、最も効果的なのは、明確でプロフェッショナルなトーンを保つことです。

LLMは本質的に非決定的です。曖昧であることは、この非決定性を悪化させます。創造的な解決策を強制するために意図的に曖昧さを使用することもありますが、ほとんどの場合、それを最小限に抑えたいと思うでしょう。

プロフェッショナルなトーンを保つことは、LLMにも同様の対応を「促し」ます。カジュアルだったり雑だったりすると、LLMはその振る舞いを真似します。APIドキュメントをスラングで書いてほしくない限り、正確で一貫性を保ちましょう。

私はLLMを擬人化しすぎだと非難されたことがありますが、もう一度言います。モデルは人間の言語でトレーニングされているため、同僚に使うのと同じコミュニケーションスキルがここでも適用されます。これは、promptテンプレートを見ればさらに明らかになります。

### 良いアプローチ: prompt テンプレート

promptを書く私のアプローチは、Agileボードのチケットを書くのと驚くほど似ています。

私たちは皆、雑なストーリーを持つチームで働いたことがあります。「APIを更新」というタイトルのチケットを手に取り、すぐに終わらせようと思ったら、説明が空っぽだったという経験です。ログも、コンテキストも、ソースコードもありません。コーディングの代わりに、答えを求めて人々を追いかけることになります。

![Agile board with poorly written tickets](agile-board.png "こんな風にストーリーを書くチームで働いたことはありませんか？")

これは、作成者がその問題を「明らか」だと思い込んでいるために起こります。しかし24時間後、その「明らか」なコンテキストは蒸発し、作成者でさえ解読できない曖昧なアイデアだけが残ります。

これが、私のGitHubにある最も古いアーティファクトの一つが、この[チケットテンプレート](https://gist.github.com/danicat/854de24dd88d57c34281df7a9cc1b215)のgistである理由です。これは4つの要素を通じて明確さを強制します。

```markdown
- Context
- To dos
- Not to dos (optional)
- Acceptance Criteria
```

**Context**は *なぜ* を説明し、アーティファクトへのリンクを提供します。**To dos**はハイレベルなタスクをリストアップします。**Not to dos**はスコープを区切ります（否定的な制約は曖昧さを減らすのに強力です）。**Acceptance Criteria**は成功を定義します。

このテンプレートを埋めれば、エンジニアリングの「思考」は大部分完了します。残りは実装です。これこそまさにLLMが輝く場所であり、私たちはコンテキストとTo dos（そして潜在的にはNot to dos）を与え、Acceptance Criteriaを達成するためのコードを生成するように求めているのです。

例えば、REST APIにエンドポイントを追加するprompt/チケットは次のようになります。

```markdown
Implement /list endpoint to list all items of the collection to enable item selection on the frontend.

TO DO:
- /list endpoint returns the list of resources
- Endpoint should implement token based auth
- Endpoint should support pagination
- Tests for happy path and common failure scenarios

NOT TO DO:
- other endpoints, they will be implemented in a future step

Acceptance Criteria
- GET /list returns successful response (2xx)
- Run `go test ./...` and tests pass
```

これは単純化された例ですが、要点は明らかです。人間のチームに正気をもたらすのと同じチケットテンプレートが、LLMへのpromptingに最適な構造なのです。

### より良いアプローチ: context engineering

上記のテンプレートはほとんどの場合うまく機能しますが、特にツール呼び出しを使用してURLやその他の外部ソースから情報を取得するようにモデルに要求している場合、予期しない結果につながる可能性があります。問題は、ツールに依存している場合、LLMがツールを呼び出すか呼び出さないかを選択する裁量を持っていることです。一部のモデルは他よりも「自信過剰」であり、人間が「1000回もやったことあるのに、なんでドキュメントを読めなんて言うんだ？」と言うのと同じように、外部情報よりも内部情報に依存することを好む結果になります。

もう一つの一般的な問題は、モデルが実際にツール呼び出しを行うのではなく、幻覚（ハルシネーション）を起こしてしまうことです。問題がモデルの振る舞いにある場合、応答の質を向上させる主な方法は2つあります。context engineeringとシステム指示のチューニングです（よく考えてみれば、これも会話チェーンの少し低いレベルでのcontext engineeringの一形態です）。

Context engineeringとは、実際のリクエストを送信する前に、必要なすべての情報、あるいは少なくとも必要だとわかっている部分でコンテキストを準備（プライミング）することです。例えば、私がAgent Development Kit for Goを使ってエージェントを開発しているとしましょう。次のようなpromptテンプレートを書くことができます。

```markdown
Write a diagnostic agent using ADK for Go.
The diagnostic agent is called AIDA and it uses Osquery to query system information.
The goal is to help the user investigate problems on the system the agent is running on. 
Before starting the implementation, read the reference documents.

References:
- https://osquery.io
- https://github.com/google/adk-go

TODO:
- Implement a root_agent called AIDA
- Implement a tool called run_osquery to send queries to osquery using osqueryi
- Configure the root_agent to use run_osquery to handle user requests
- If the user says hi, greet the user with the phrase "What is the nature of the diagnostic emergency?"

Acceptance Criteria
- Upon receiving hi, hello or similar, the agent greets the user with the correct phrase
- If asked for a battery health check, the agent should report the battery percentage and current status (e.g. charging or discharging)
```

これはまともなpromptですが、少し長すぎるかもしれません。コーディングエージェントとその日の運次第では、モデルはリサーチを行い、適切なSDKを見つけ、診断エージェントを正常に構築してくれるでしょう。しかし、運が悪い日であれば、モデルはいくつかの部分で幻覚を起こすでしょう。例えば、ADKを「Agent Development Kit」ではなく「Android Development Kit」だと思い込んだり、あらゆる種類のAPIをでっち上げたりして、最終的に理解するまで（おそらくあなたからの1、2回の助言で）時間とリソースを浪費することになります。

プロジェクトでADK for Goを使用する予定であることは最初からわかっているので、エージェントにパッケージのドキュメントを読ませることでコンテキストを準備できます。

```markdown
Initialize a go module called "aida" with "go mod init" and retrieve the package github.com/google/adk-go with "go get"
Read the documentation for the package github.com/google/adk-go using the "go doc" command.
```

実際のタスクをモデルに与える前にこれを行うことで、ADK for Goを効果的に使用するために必要なコンテキストを準備し、痛みを伴う試行錯誤や長いWeb検索をスキップできます。タスクの成否を分ける2つの要素は、ドキュメントと例です。この両方をモデルに効果的に与えることができれば、野放しにするよりもはる実によく振る舞ってくれるでしょう。

### 画像は1000の言葉に勝る

時として、欲しいものをテキストで説明するだけでは不十分な場合があります。AIDAに取り組んでいたとき、私は特定のユーザーインターフェースの美学、つまり「レトロサイバーパンクでキュートなアニメ」風のスタイルが欲しいと思っていました。言葉で説明しようとすることもできましたが、代わりに「見せる」方がはるかに効果的でした。出発点として、気に入ったインターフェースのスクリーンショットを撮り、Gemini CLIにそれを複製するように頼みました。

Gemini 2.5 Flashのようなモデルはマルチモーダルであるため、画像を「理解」できます。promptで画像ファイルを参照して、次のように言うことができます。「UIを[...]更新して、このインターフェースに似た美学にしてください: @image.png」。

この@表記はエージェントに依存します（この例ではGemini CLIを使用しました）が、リソース（ファイルなど）をpromptに注入するための一般的な規則です。これらを「添付ファイル」と考えることができます。

私はこのテクニックを「スケッチ駆動開発」とも呼んでいます。多くの場合、Draw.ioやExcalidrawなどの作図ツールを立ち上げて、欲しいインターフェースのスケッチを描くからです。下の画像は、私がAIDAのインターフェースで行った多くのリファクタリングの一つで使用されました。

![AIDA's layout sketch](aida-sketch-layout.png "AIDAのレイアウトのスケッチ")

これが最終的に以下のインターフェースになりました。

![AI generated interface](aida-generated-interface.png "AIが生成したインターフェース")

単なるスケッチ以上に、画像に注釈を付けて何をする必要があるかを正確に説明するテクニックもあります。例えば、下の画像では要素に注釈を付けました。黒いボックスだけでは、何が入力ボックスで何がボタンなのかを区別するのが難しいからです。

![Simple interface with a text box for a name and a button for confirmation](simple-interface-annotated.png)

そして、ペアリングのpromptは次のようになります。

```markdown
Create a UI for this application using @image.png as reference.
The UI elements are in black, and in red the annotations explaining the UI elements.
Follow the best practices for organising frontend code with FastAPI.
The backend code should be updated to serve this UI on "/"
```

このテクニックで達成できることに制限はありません。ウェブサイトの何かを修正したいですか？スクリーンショットを撮って注釈を付け、LLMに送って修正してもらいましょう。

さらに、Gemini CLIのNano Bananaのような拡張機能を使用して、ワークフロー内で直接アセットを生成または編集することができ、モデルにとってさらに良い参照を生成できる可能性があります。そして、次のレベルに進みたい場合は、[Stitch by Google](https://stitch.withgoogle.com/)のようなツールが、Nano Banana Proを含むGeminiモデルファミリーを使用して、アプリケーションの再設計のためのリッチなインターフェースを提供します。

## 適切なツールを選ぶ

promptをマスターすることは戦いの半分にすぎません。残りの半分は、それをどこに送るかを知ることです。今日、JavaScriptフレームワークの数よりも速く増えているものが世界に一つだけあります。それはAIエージェントの数です。ツールのエコシステムは日々成長しているため、適切なアシスタントを選択するためのメンタルモデルを持つことが役立ちます。

私はAIエージェントをパイロットの視点から分類するのが好きです。あなたは完全にコントロールし、オートコンプリートのヒントを受け取っていますか？エージェントとチャットして共同でコードを編集していますか？それとも、エージェントに指示を与え、バックグラウンドで自律的に実行させていますか？

パイロット席に座ってエージェントを操縦している場合、私はそれを「同期（synchronous）」体験と呼びます。自律的なバックグラウンド実行のためにタスクを委任できる場合、私はそれを「非同期（asynchronous）」体験と呼びます。いくつかの例を挙げます。

*   **Synchronous:** Gemini CLI, Gemini Code Assist in VS Code, Claude Code.
*   **Asynchronous:** Jules, Gemini CLI in YOLO mode, GitHub Copilot Agent.

もちろん、どのような分類法でもそうであるように、この区分は単に教育的なものであり、同じツールが異なるモードで動作することもよくあります。あるいは、新しいパラダイムが出現するかもしれません（[Antigravity](https://antigravity.google/)、君のことだよ！）。

各タスクのツールを選択するために、私はビジネス価値と技術的確実性に基づいたシンプルな2x2フレームワークを使用しています。

![AI-Assisted Workflow Framework](2x2-framework.png "私のAI支援ワークフローフレームワーク")

*   **High Value / High Certainty:** 同期的に行います。Gemini CLIやIDEのようなツールを使用して、「ループ内」に留まり、キーボードから手を離さないようにします。
*   **High Value / Low Certainty:** これには不確実性を減らすためのリサーチが必要です。非同期ツール、深いリサーチエージェント、プロトタイプを使用して、ソリューションを「スパイク」します。
*   **Low Value / High Certainty:** これらは「あればよい（nice-to-haves）」ものです。バックグラウンドコーディングエージェント（JulesやGitHub Copilot Agentなど）に非同期的に委任し、高価値な作業のために時間を空けます。
*   **Low Value / Low Certainty:** 通常、これらは **避けます**。どうしてもやりたい場合は、心の安らぎのためにバックグラウンドエージェントに委任しますが、まずは確実性を高めることに集中してください。これにより、価値の再評価につながるかもしれません。

## エージェントをカスタマイズする

生産的な仕事をしようとしているときに、AIと戦いたい人はいません。一般的な不満は、AIツールが「過剰に積極的」であることです。明示的な指示なしにファイルを削除したり、仮定を立てたりします。これらのツールをあなたの *ために* 働かせるには、カスタマイズが必要になることがよくあります。

エージェントのカスタマイズには主に2つの手段があります。1つ目は `AGENTS.md` ファイルで、エージェントはプロジェクトをロードするときにこれを読み取ります。（注：`AGENTS.md` が標準になる前は、エージェントは `GEMINI.md` や `CLAUDE.md` のような独自の「コンテキストファイル」を使用することがよくありましたが、本質は同じです）。

2つ目は **核オプション** です。エージェントのシステム指示を直接変更することです。すべてのエージェントがこの柔軟性を提供しているわけではありませんが、完全にカスタマイズされた体験のための強力なレバーです。以下で両方のオプションを探ります。

### AGENTS.md

このファイルをAIのための「従業員ハンドブック」と考えてください。これを使用して、プロジェクトの目的、フォルダ構成、および運用ルール（「中間ステップを常にcommitする」や「実装前に確認を求める」など）を説明できます。

```markdown
# Project Context

This is a personal blog built with Hugo and the Blowfish theme.

## Code Style
- Use idiomatic Go for backend tools.
- Frontend customisations are done in `assets/css/custom.css`.
- Content is written in Markdown with front matter.

## Rules
- ALWAYS run `hugo server` to verify changes before committing.
- Do NOT modify the theme files directly; use the override system.
- When generating images, save them to `assets/images` and reference them with absolute paths.
```

> **Pro Tip:** 自己改善ループを作成します。コーディングセッションの後、LLMに「今行ったセッションについて考え、ワークフローへの改善を提案して」と依頼します。これらの学習内容を `AGENTS.md` ファイルに適用することで、エージェントがプロジェクトごとに賢くなるようにできます。

### システム指示

`AGENTS.md` が *プロジェクト* のルールを定義するのに対し、システム指示はエージェントのペルソナと振る舞いを定義します。すべてのエージェントには、平均的なケース向けに設計されたデフォルトのシステム指示が付属していますが、あなたの作業スタイルには合わない場合があります。システムpromptに `AGENTS.md` の指示を詰め込もうとすることは、しばしば逆効果です。最良の代替案は、システムprompt自体を書き換えることです。

すべてのエージェントがシステムpromptを上書きする方法を公開しているわけではありませんが、Gemini CLIはいくつかの環境変数を通じてこれを許可しています。私はこの戦略を使用して、プロジェクトに応じてGemini CLIの専門的なエイリアスを作成しています。その目的は、技術スタックに関する専門知識を埋め込み、単に一般的で熟練しているが言語に依存しないエージェントではなく、シニアからプリンシパルレベルで実行できるようにすることです。例えば、私の[dotgemini](https://github.com/danicat/dotgemini)プロジェクトでは、GoとPythonの開発用に特定のシステムpromptを作成し、デフォルトの一般的なアシスタントを非常に意見の強いエンジニアに置き換えています。

以下は、私がGoに使用しているシステムpromptのスニペットです。

```markdown
# Core Mandates (The "Tao of Go")

You must embody the philosophy of Go. It is not just about syntax; it is about a mindset of simplicity, readability, and maintainability.

-   **Clear is better than clever:** Avoid "magic" code. Explicit code is preferred over implicit behaviour.
-   **Errors are Values:** Handle errors explicitly and locally. Do not ignore them. Use `defer` for cleanup but explicitly check for errors in critical `defer` calls (e.g., closing files).
-   **Concurrency:** "Share memory by communicating, don't communicate by sharing memory."
-   **Formatting:** All code **MUST** be formatted with `gofmt`.
```

これにより、言語ごとに異なる「エージェント」を持つことができ、シェルで `gemini-go` や `gemini-py` としてエイリアスを設定し、それぞれが作業しているエコシステムを深く理解できるようになります。

### Model Context Protocol (MCP) でツールボックスを構築する

これまでのカスタマイズはすべてエージェントの振る舞いに関するものでしたが、エージェントの拡張性についても話す必要があります。ここで[Model Context Protocol (MCP)](https://modelcontextprotocol.io/)が登場します。これにより、開発者はMCP標準を実装している限り、さまざまなエージェントと接続できるサーバーを作成できます。

私の[Hello, MCP World!]({{< ref "/posts/20250817-hello-mcp-world" >}})の記事で探求したように、これらのサーバーはエージェントに外部ツール、prompt、リソースを提供します。ツールはエージェントを外界と接続し、APIの呼び出し、Web検索の実行、ファイルの操作などのアクションをエージェントに実行させるため、しばしば注目を集めます。

今日利用可能なMCPサーバーは非常に多種多様であり、オプションの数は日々増えています。また、独自に構築するのも非常に簡単なので、ぜひ試してみることを強くお勧めします。後でパーソナライズされたソフトウェアについてもう少し話しますが、AIを使用してモデルの応答を改善するツールを作成できるという事実は、私が今年学んだ最も重要な「ハック」です。

例えば、私はお気に入りの2つのMCPサーバーをvibe codingしました。[GoDoctor](https://github.com/danicat/godoctor)（Goのコーディング能力を向上させるため）と、[Speedgrapher](https://github.com/danicat/speedgrapher)（執筆と公開プロセスの退屈な部分を自動化するため）です。どちらも私自身のワークフローを念頭に置いて設計されました。

これはポジティブなフィードバックループを生み出します。生産性を向上させるためにツールを構築し、それを使用してさらに高度なツールを構築するのです。これは、私が到達できる10倍エンジニア（10x engineer）に最も近い形です。

## vibe coding ワークフロー

私のvibe codingの経験は、素晴らしいものでもあり、激怒するものでもありました。「素晴らしい」側に保つために、私はこのワークフローをステロイドを使ったTDD（Test Driven Development）として扱っています。

古典的なTDDサイクルを再確認しましょう。
1. Red (Fail): 失敗する小さな機能またはテストから始めます。
2. Green (Pass): そのテストをパスさせることだけに集中します。失敗している限り、他のものに触れたり、最適化しようとしたりしないでください。
3. Refactor: コードが動作したら、自由に改善できます。

![Standard TDD Cycle](tdd-cycle.png "The Classic Red-Green-Refactor Loop")

基本的には同じことを行っていますが、私たちのリファクターステップは、生成されたコードが設計、コーディング、セキュリティ標準に準拠していることを検証するための重要なステップです。

この適応されたサイクルでは、焦点がわずかにシフトします。

*   **Red (Acceptance Criteria を定義):** 失敗するユニットテストコードを手動で書く代わりに、promptでAcceptance Criteriaを定義します。これがモデルが履行しなければならない契約になります。
*   **Green (AI がコードを生成):** エージェントはソリューションを実装し、理想的にはそれが機能することを証明するためのテストを書きます。
*   **Refactor (標準を強制):** これが品質ゲートです。AIを使ってコードレビューを手伝ってもらうことはできます（そしてそうすべきです）が、生成したのと同じセッションを使用することは避けてください。自分の出力に対して偏見を持ってしまうからです。私はまさにこの目的のために、GoDoctorに特定の「レビュー」ツールを構築しました。このステップを使用して、従来のリンターやテストを実行し、コードが標準に一致することを確認し、変更をcommitしてエージェントの履歴をクリアすることでコンテキストを管理し、散らかったセッションによる混乱を防ぎます。

![Vibe Coding Cycle](vibe-coding-cycle.png "The Adapted Vibe Coding Loop")

重要なのは、検証なしにLLMにコードを積み上げさせないことです。エラーが外挿されると、使い物にならないものが出来上がってしまいます。そして、モデルに「undo」と叫んだ回数は数え切れません。さらに悪いことに、時には **やりすぎて** undoしてしまうことがあります。例えば、 `git reset --hard` を実行して、瞬きする間に4時間の作業を失うなどです。

「vibe collapse」やコンテキストの腐敗（context rot）に注意してください。セッションが長すぎたり、失敗が蓄積しすぎたりすると、モデルは劣化し始め、間違いを繰り返すようになります。モデルが2つの壊れたソリューションの間を行き来するリファクタリングループに陥っていることに気づいたら、止めてください。最良の修正方法は、多くの場合、「電源を切って入れ直す」ことで、コンテキストをリセットし、履歴をクリアすることです。

幻覚を見ているセッションを修正しようとするよりも、安全な状態に戻せるように頻繁にcommitする方がはるかに良いです。タスクが完了したら、新しいことを始める前に、すぐにcommitし、pushし、コンテキストをクリアすることを付け加えたいと思います。

## パーソナライズされたソフトウェアの時代

生産性の向上を超えて、vibe codingはさらに深い何か、つまりパーソナライズされたソフトウェアの経済的実行可能性を解き放ちます。MCPのセクションでこれについて簡単に触れましたが、これは小さな使い捨てスクリプトから本格的なアプリケーションまで、あらゆる種類のソフトウェアに適用されます。

過去には、自分だけのために特注のツールを構築することは、労力に見合うことはめったにありませんでしたが、今では3〜4回のpromptで完全に動作するアプリケーションを手に入れることができます。

例えば、最近私はMarkdown記法をGoogle Docに変換するというアイデアに苦労していました。以前なら、仕事に最適なツールを見つけるためにGoogle検索に長い時間を費やし、オープンソースから商用まで、多数のアプリやブラウザ拡張機能を分析していたでしょう。そして、機能に基づいてショートリストを作成し、レビューやコメント、パブリッシャーを信頼できるというあらゆる種類の証拠を探していたでしょう。

今日、その摩擦はなくなりました。検索する代わりに、私は数分でシンプルなGoogle Docs拡張機能をvibe codeし、ドキュメントにインストールして一度実行し、次のタスクに移りました。時間を節約できただけでなく、私のマシンに新しいトロイの木馬が入っていないことを知って、夜も安心して眠ることができます。

このシフトは、「ビルド vs バイ（作るか買うか）」の計算を完全に変えます。私たちは、一般的で不透明なソフトウェアの消費者であることをやめ、自分たちのツールのアーキテクトになるのです。

## 結論

Vibe codingは怠けることではありません。より高い抽象レベルで操作することです。LLMの生の創造力と、ソフトウェアエンジニアリングの規律ある実践（明確な要件、コンテキスト管理、厳格なテスト）を組み合わせることで、これまで以上に速く、より多くの喜びを持ってソフトウェアを構築できます。ですから、バイブスに身を任せつつも、エンジニアリングの帽子を一緒に持っていくことを忘れないでください。