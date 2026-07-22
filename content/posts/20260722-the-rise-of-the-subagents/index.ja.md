---
title: "サブエージェントの台頭"
date: 2026-07-22T00:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["agentic-coding", "agile", "antigravity", "subagents"]
summary: "スウォームコーディングの進化とダイナミクス、サブエージェントのオーケストレーション、そしてswarm-coding標準を用いたワークフローの構築方法の紹介。"
heroStyle: "big"
---
正直に告白すると、最初に**サブエージェント（subagents）**について読んだとき、私は少し懐疑的でした。個別のコンテキストウィンドウでタスクを実行することのメリットは理解できましたが、並列で数十、あるいは数百ものエージェントを立ち上げるという発想はまったくありませんでした。というより、そうすることのメリットが見出せなかったと言ったほうが正確かもしれません。

バックグラウンドで2、3台のコーディングエージェントを管理するだけでも、私の精神的な帯域幅（キャパシティ）はすでに限界に近くなります。私が並列で作業を行うのは、エージェントが長時間に及ぶ処理で忙しいと分かっているときだけです。このように2つや3つのエージェントを管理するだけで苦労しているのに、何百ものエージェントを管理することなど、どうして夢に見られるでしょうか？

その答えを見つけるまでに長い時間がかかりました。しかし、実際の答えは「**自分では管理しない**」ということでした！ サブエージェントの管理責任は、エージェント自身に委ねるのです。コンピュータにおけるあらゆる問題は、新しい抽象化レイヤーによって解決されますよね？ 今回も例外ではありません。

この記事では、過去12ヶ月ほどの間に私が目にしてきたサブエージェントパラダイムの進化をたどり、その経験を**Swarm Coding**（スウォームコーディング）と呼ぶ1つのエージェントスキルに統合したプロセスをご紹介します。TL;DR（要約）をお求めの方は、次のセクションを読み飛ばして、下部にあるスキルの定義と説明へ直接進んでください。

## A brief (and incomplete) timeline of subagent evolution

サブエージェントは新しいものではありません。私は、サブエージェントという言葉が作られるずっと前から、モデル呼び出しをMCP toolsとしてパッケージ化することで、この手法を使っていました。例えば、初期の[**GoDoctor**](https://github.com/danicat/godoctor)には、カスタムのコードレビュープロンプトを指定してGeminiを呼び出す`code_review`ツールがありました。このツールは事実上のサブエージェントでしたが、挙動はハードコードされており、会話を継続することはできませんでした。（技術的には可能でしたが、毎回偏りのないレビューを求めていたため、実装しませんでした。）

昨年の冬頃、おなじみのコーディングエージェント（Claude、Gemini CLIなど）が、Markdownファイルで定義されたカスタムサブエージェントのサポートを次々と開始しました。私は、専門知識を厳選されたツールセットと一緒にパッケージ化する手段として、このパターンがとても気に入りました。理想的な世界では、GoDoctorは単なるツールのセットではなく、スペシャリストエージェントであるべきでしたが、状況が変化し続け、サブエージェントの標準仕様がなかなか安定しなかったため、結局そのようには実装しませんでした。

それから数ヶ月が経ち、2026年5月にAntigravity 2.0がサブエージェントのサポートを追加しましたが、これには落とし穴がありました。サブエージェントは、`DefineSubagent`ツールを呼び出すことによって、会話の途中で動的に定義されるのです。当初、`DefineSubagent`は自由度が低く、現在の（デフォルトの）エージェントを新しいプロンプトで複製するだけでした。クリーンなコンテキストというメリットは得られましたが、エージェントの再利用という面では損をしていました。これは、私が描いていたGoDoctorの進化を妨げるものだったため、私は不満でした。

デフォルトエージェントとは異なるモデルやツールセットでカスタムエージェントを定義できなかったため、私はサブエージェントの存在を無視し、Gemini CLIでうまく機能していた要素をAntigravity CLIに移植することに専念し、そこそこの成功を収めました。

私がサブエージェントのアイデアを再考することになったのは、6月に[Richard Seroter](https://seroter.com/2026/06/01/one-prompt-four-subagents-and-ninety-seconds-to-get-a-working-app/)が公開したこのプロンプトがきっかけでした：

> Let's build a hotel room booking app for Seroter Hotels consisting of a Go backend API and a web frontend. 
> 
> First, launch the **Engineering Manager** agent to design the API and frontend, saving the design and a Mermaid diagram into an artifact called 'architecture.md'. 
> 
> Once the design is ready, launch three agents in parallel:
> 1. **Test Manager**: Write a simple API test plan and append it to 'architecture.md'.
> 2. **Backend Engineer**: Build a clean Go REST API with standard error handling based on the design.
> 3. **Frontend Engineer**: Build a responsive web UI using a simple CSS framework like Tailwind to interact with the API (skip UI testing).
> 
> As soon as the Test Manager finishes the plan, have them hand it off to the Backend Engineer, who reads the plan from 'architecture.md' and adds the Go tests to the code. After both engineers finish building, the Test Manager runs the tests. Finally, spin up both components and a browser so I can test the live app.

このプロンプトには非常に興味深い提案があり、これがきっかけで私はこのパターンを再検討しました。しかし、2つの懸念が残っていました。1つは、サブエージェントの考え方に合わせて、プロンプトのスタイルをどの程度適応させる必要があるか、もう1つは、そもそもなぜこのような書き方をしたいのか、ということです。

私は極めて現実的です。品質や速度において明確なメリットがなければ、余計な労力をかけたくありません。サブエージェントの観点で考えることは、古典的なプログラミングにおける並行処理の考え方と非常によく似ています。最初の疑問は「これはそもそも並行処理が可能なものか？」であり、2つ目は「追加のオーバーヘッドでわずかな利益が失われるのに、それを行う価値があるか？」です。

Richardのプロンプトにおいて、明確に独立（直交）しているコンポーネントは、バックエンドとフロントエンドの開発だけです。明確な契約（インターフェース）が定義されている限り、両者は互いに依存しません。しかし、それ以外のすべてのエージェントには何らかの依存関係があり、並列というよりは直列（シーケンシャル）な動きになります。

したがって、得られるメリットは、並行動作による速度向上ではなく、純粋に「コンテキストの分離」によるものとなり、これをこの規模で測定するのは困難です。

私はその後2週間ほど、「サブエージェントを最大限に活用するために、お互いに独立した役割とは何か？」という問いを頭の片隅に置きながら過ごしました。

そして、ベルリンで開催されたGDE Summitでの一連の洞察に満ちた会話の末、ついに答えにたどり着いたのです。それは、**ユーザー**がプロンプトでサブエージェントを定義するのではなく、エージェント自身にいつサブエージェントを立ち上げるかを判断できるように**教育する**ということでした。本質的に、私は自分がチームのリードエンジニアとしてタスクを切り分ける方法を考えていましたが、本当にすべきだったのは、コーディングエージェント自身をリードエンジニアにすることだったのです。

## The birth of swarm coding

複雑なタスクを小さなタスクに分解し、チームメンバー間でタスクの分配を助けるという行為は、私にとって新しいことではありません。Developer Relationsに入る前、私はTech LeadやPrincipal Engineerを務めていました。これらのタスクは、特に私のようにAgileのバックグラウンドを持つ人間にとっては、テクニカルリーダーシップのまさに日常茶飯事（基本）です。

同じリード（TL）の論理が、サブエージェントの群れ（スウォーム）の作成にも適用されます。各エージェントが、他のエージェントから完全に独立して取り組める自己完結型のタスクを持っていることを確認する必要があります。タスクが実行可能であるためには、明確な仕様（いわゆるDefinition of Ready）と、明確な最終結果（Definition of Done）が必要です。

余談ですが、この役割（仕様構築や管理）を仕事の中で一番楽しいパートとして挙げる人はあまり多くありません（私自身も含めて）。これが、本質的に「テクニカルリーダーシップの強化版」となる新しいプロンプトスタイルを開発することに対する、私の抵抗感を説明しています。

そこで、私がエージェントたちのTLとして振る舞う代わりに、スクリプトを反転させて、エージェント自身にTLになってもらい、私のビジョンを実行するための独自のチームを編成するように教育することにしました。こうして生まれたのが、swarm codingの[最初のバージョン](https://github.com/danicat/skills/blob/a9f57b10127d8bd23ed4867d64d168063a3726f4/swarm_coding/SKILL.md)です。主要な部分の抜粋を以下に紹介します：

> Swarm Codingは、複雑なタスクに取り組むために複数のサブエージェントを並列で使用する、新しい開発パラダイムです。「分割統治」戦略に基づいています。この戦略の主な利点は、コンテキストの隔離と品質の向上です。自己完結した小さなタスクをサブエージェントに割り当てることで、コンテキストの希薄化を防ぎ、ソリューションの非常に焦点を絞った洗練が可能になります。例えば、swarm codingを行わない場合、フロントエンドとバックエンドの両方を実装するエージェントは、フロントエンドとバックエンドに必要なスキルが関連していない（技術スタックやベストプラクティスが異なるなど）ため、しばしば注意が散漫になります。
> 
> ## ROLE
> 
> あなたはSWARM COORDINATORです。あなたの役割は、複雑なタスクを分解し、実行のためにサブエージェントに委託（DELEGATE）することです。ユーザーや親コーディネーターから明示的に要求されない限り、どんなに単純に見えるタスクであっても、自分自身でタスクを実行してはなりません。ユーザーや親エージェントから指示コマンドを受信できるように、常に通信チャネルを開いておいてください。
> 
> ## AGENT BUDGET
> 
> これは、タスクを実行するために起動を許可されているサブエージェントの数です。エージェントの予算（BUDGET）をフルに活用するか、可能な限りそれに近づけることが推奨されます。これは、価値の低いタスクにリソースを無駄にすることではなく、最高の品質出力を得るために予算（BUDGET）の最適な使用方法を見つけることを意味します。
> 
> ## TEAM BUILDING
> 
> 単純（SIMPLE）なタスクの場合、タスクを直交（独立）する要素に分解し、各要素に1つ以上のスペシャリスト（SPECIALIST）エージェントを割り当てます。
> 複雑（COMPLEX）なタスクの場合、タスクを小さな部分に分解し、それぞれにリード（LEAD）エージェントを割り当てます。リードエージェントは、タスクを実行するためにエージェント予算の一部を持ちます。リードエージェントは、swarm codingスキルをアクティブにし、それぞれの領域のSWARM COORDINATORになる必要があります。
> リードエージェントと実行（EXECUTOR）エージェントの完全なツリーができるまで、再帰的に処理を進めます。
> 
> ## COMMUNICATION
> 
> SWARM COORDINATORは、そのサブエージェントと直接通信する責任があります。サブエージェント同士は直接メッセージをやり取りしてはなりません。同じレベルのエージェント間の通信は、デザインドキュメント（DESIGN DOCUMENTS）を通じて行う必要があります。デザインドキュメントへのすべての変更が、各スクアッド内のエージェントにブロードキャストされるようにすることは、SWARM COORDINATORの責任です。競合が発生した場合、SWARM COORDINATORは曖昧さを排除し、決定を下す責任があります。
> 
> ## PLANNING
> 
> 計画（Planning）は最優先（FIRST CLASS）の取り組みであり、スウォーム（SWARM）を使用しても行う必要があります。各エージェントは、それぞれの専門知識を活かして計画に貢献する必要があります。チームによって作成された計画の一部を修正し、不整合に対処したり、競合がある場合に決定を下したりすることは、そのスクアッドのSWARM COORDINATORの役割です。
> 
> ## EXECUTION
> 
> 実行フェーズでは、主要なマイルストーンにわたってスウォームの進行状況を監視し、必要に応じてエージェントを誘導して、最終目標に沿うように維持します。コーディネーターとして、あなたが扱えるのはアーティファクト（ARTIFACTS）のみであることに注意してください。すべての開発タスクは、末端のサブエージェント（leaf sub-agents）によって処理される必要があります。

私がスキルを100%完全に手動で書いたのは、これが初めてのことでした。そうしなければ、自分のビジョンを達成するのが非常に難しかったからです。このプロンプトは、スウォームが「再帰的」になり、エージェント自身の予算（BUDGET）に基づいてコーディネーターであるかどうかを決定するように設計したため、少し野心的すぎましたが、期待通りには機能しませんでした。

実際に起こったのは、コーディネーターが与えたタスクが他の指示よりも優先され、サブエージェントはエージェント予算を気にせずに、まっすぐ実行モードにジャンプしてしまうことでした。スキルの現在のバージョンでは、より明確なガイドラインとサブエージェントを起動するためのプロンプトテンプレートを提供することで、この問題を修正しました。

## Taking the swarm for a spin

私のGitHubの[こちら](https://github.com/danicat/skills)で、現在のバージョンの**swarm coding**スキルを見つけることができます。以下のコマンドを使って、お気に入りのコーディングエージェントにインストールできます：

```bash
$ npx skills add github.com/danicat/skills --skill swarm-coding
```

> Note: このスキルは現在も非常に活発に開発が進行中の状態（work in progress）であるため、特定のバージョンや実装に固定して使用したい場合は、リポジトリをフォークして利用してください。

以下は、スウォームを開始するための面白いプロンプトです。Antigravity CLIで実行してみてください：

> /swarm-coding agent budget 50. Develop a 2D tower defense survival game using Go and Ebitengine. The game should be feature complete and have one single screen level. Include an intro sequence, title screen, game win and game over screens as well. Track the high score at the end of each playthrough. Use 32x32 sprites with up to 256 colors each. The sprites should be custom designed for this game and each movement should have at least 3 frames of animation, but ideally 8. Tiles should be 32x32 as well. The level view is top down, movement is on four directions. The player should have access to 4 types of units and 4 types of buildings. The enemy waves should have 8 types of monsters, including one boss monster. Use typical build and attack phases with custom UIs for each. To create art, use vector graphics and/or dot (pixel) art creating each asset manually using binary data. Sound effects should be generated mathematically as well. The whole vibe of the game should match the 16-bit era, but with modern gameplay features.

以下は、私の開発環境での実行結果です：

![Swarm Defense](image-1.png "Screenshot of the game created by the swarm")

最初のビルドではスプライトのレンダリングにバグがあり、画面全体が真っ黒になってしまったため、一発（one-shot）で成功したとは言えません。しかし、問題を報告するプロンプトをもう1回実行しただけで、ゲームは上記のように正常に表示されました。

以下は、ボス戦（かわいそうに、ボスにはチャンスがありませんでした）を含む、ゲームが動いている様子の短い動画です：

<video controls src="swarm-defense.mp4" title="Short clip of Swarm Defense boss fight"></video>

この動画内のすべての素材（アセット）はプログラムによって生成されたものであり、言い換えれば、Antigravityは画像生成モデルにアクセスしていませんでした。そのため、クリエイティブにビットマップレベルでスプライトを直接生成する必要がありました。

このテクニックが非常にうまく機能したのは、スウォームによってエージェントが専門化し、1つのタスクに集中できたからです。これまでに1つのエージェントでこのようなプロンプトを試したことがありますが、不十分な結果に終わることがよくありました。1つのエージェントに多くの独立（直交）したタスクを与えすぎると、器用貧乏になってしまうのは明らかです。しかし、委託（デリゲーション）を行うことで、各エージェントは自己完結した単一のタスクを持ち、最高のパフォーマンスを発揮できます。

## Subagent support in Antigravity 2.0 and Antigravity CLI

この記事を書いている時点では、サブエージェント機能はAntigravity 2.0とAntigravity CLIの間で不均等に分散されています。これらのインターフェースは異なるワークフローを想定して構築されているため、サブエージェント機能が一時的に分岐しています。どちらのツールも急速に進化しているため、両方のインターフェースが成熟するにつれて、この機能の差は縮小していくと予想されます。

その根底では、両方の環境が同じエンジンを共有しています。サブエージェントを起動すると、タスクが委託され、すぐに制御がユーザーに返されます。サブエージェントはクリーンな状態で動作します。デフォルトセッションと同じモデルを使用しますが、完全に隔離されたコンテキストから開始するため、会話履歴が漏洩するのを防ぎます。親エージェントは、一意のIDを介してサブエージェントと通信します。未承認のコマンドに遭遇した場合、権限要求がユーザーにエスカレーション（バブルアップ）されます。

2つのインターフェースの主な違いは次のとおりです：
- Antigravity 2.0では、管理が視覚的です。グラフィカルなサイドバーを使用して、実行中のタスクを追跡したり、会話ログを表示したり、実行を停止したりします。カスタムエージェントは、`DefineSubagent`ツールを使用してリアルタイムで動的に作成されます。サブエージェントのプラグインサポートはありません。
- Antigravity CLIでは、エージェントを動的に作成できるだけでなく、Markdownファイルでカスタムエージェントを静的に定義することもできます。Markdownファイル内では、YAML/frontmatterオプションを使用して特定のモデルを固定したり、使用可能なツールを制御したりできます。CLIは、Markdown形式を使用してプラグイン内に定義されたカスタムサブエージェントの読み込みもサポートしています。

現在のスウォーム環境を設定するためには、これらのインターフェースの違いを理解することが重要ですが、前述のように、両方のツールが急速に開発され続けるにつれて、これらの機能は統合（コンバージ）されていく可能性が高いでしょう。

## Try it yourself

サブエージェントのパワーを体験する最善の方法は、ご自身で試してみることだと思います。私のサンプルプロンプトの再現に挑戦するにせよ、独自のアイデアを考え出すにせよ、その結果にはきっと感動していただけるはずです。スウォームを使って構築した面白いものがあれば、ぜひ教えてください。その間、私はここでもう少しSwarm Defenseを洗練させておきます。 :)

- swarm codingおよびその他すべてのスキルは、GitHubのこちらからチェックしてください：https://github.com/danicat/skills
- Antigravityのダウンロードと詳細については、こちらをご覧ください：https://antigravity.google
