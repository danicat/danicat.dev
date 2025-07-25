---
title: "AI対応の世界のための現代的な開発者ワークフロー"
date: 2025-07-11
author: "ダニエラ・ペトルザレク"
tags: ["gemini-cli", "vibe-coding", "jules"]
categories: ["働き方"]
summary: "AIツールと単純な優先順位付け演習を使用した現代的な開発者ワークフローの提案。"
---
{{< translation-notice >}}

## はじめに

ベルリンで開催されたWeAreDevelopers World Congress 2025から戻ってきたばかりで、ヨーロッパ中や世界中から集まった多くの開発者に刺激を受けています。もちろん、今年のメインテーマはAIでした。AIはどこにでもあります！今では、クラウド、コンピューター、サングラス、トースター、キッチンシンク、トイレットペーパーにAIがあります。誰もAIから逃れることはできません…JSフレームワークでさえAIよりも速く成長することはできません！私たちは破滅です！！！>.<

あるいはそうではないかもしれません！私たちは恐ろしい時代に生きていることを知っています。テクノロジー業界は変化しています。企業は、AIが人々をより生産的にしている、あるいは人々を完全に置き換えているという前提で、あちこちで人々を解雇しています。これが解雇の本当の理由なのか、それともAIが隠された議題のスケープゴートにすぎないのかは、このブログ投稿ではなくパブでの会話ですが、変化が起こっていることを認めましょう。

時間と経験から学んだ重要な人生の教訓の1つは、制御できないことを気にするべきではないということです。AIは避けられないので、将来私たちの仕事がどうなるかを心配する代わりに、今日AIが実際にあなたの仕事を改善するために何ができるかについて考えてみませんか。これは有名なグラス半分の水のアプローチです。危機を機会に変えましょう。ですから、少し警戒を解いて、私たちのプロセスに「バイブ」を取り入れることで、現代の開発者のワークフローがどのようになるべきかを想像してみましょう。

これは実際に私が過去4週間適用してきたワークフローであり、非常にうまく機能していますが、もちろん、まだ完璧ではないので、鵜呑みにしないでください。とはいえ、このシーンは日々進化しており、良くなる一方だと信じています。

## 優先順位付けへの2次元アプローチ

実際の「AIネイティブ」な働き方に入る前に、私がキャリアの過去7年以上使用してきた優先順位付けモデルを簡単に紹介します。これは、ThoughtWorksでアジャイル変換プロジェクトに取り組んでいるときに学んだ方法であり、その後、自分のニーズに合わせて調整しました。これには、関連する人々（エンジニアと利害関係者の両方を含む）を部屋に集めて議論し、バックログを実行するための論理的な順序を作成することが含まれます。

これは、`技術的な確実性`と`ビジネス価値`という2つの直交する原則に基づいています。

![優先順位付け演習の結果](image-3.png "演習結果に基づく実装の優先順位")

技術的な確実性とは、機能の実装パスがどれほど明確であるかです。技術的な確実性が高い場合、機能を実装するためのすべての（またはほぼすべての）ステップがわかっていることを意味します。技術的な確実性が低い場合、機能の実装方法がわからない、またはいくつかの最初のステップしかわからない可能性があることを意味します。

ビジネス価値とは、この機能がチームの目標を達成するためにどれほど重要であるかです。ビジネス価値が高いということは、その機能がビジネスの成功に不可欠であることを意味し、ビジネス価値が低いということは、「あればいいな」という種類の機能である可能性が高いことを意味します。

この演習の最終的な目標は、**すべて**が重要と見なされる行き詰まりを打破することです。すべてが重要であっても、それらを並べて置くことで、最も厳格な利害関係者でさえ、他の優先事項に対して自分の立場を再考するのに役立ちます。また、価値は同じでも技術的な確実性が異なるものにも実行順序があります。最初に「簡単に手に入るもの」を選ぶことで、チームは残りの機能の不確実性を減らす（技術的な確実性を高める）ための時間を稼ぐことができます。

![優先順位別の作業モード](image-2.png "優先順位付けに基づく推奨作業モード")

このプロセスはAIネイティブな作業とどのように関連しているのでしょうか？私は自分自身を多くのAIの上司だと考えています。私は自分のバックログを優先順位付けし、各タスクにどのツールを使用するかを整理しています。機能に「投資」している場合は、同期的に作業することを優先しますが、そうでない場合は、非同期ワーカーに委任できます。

## 基本的な「AIネイティブ」ワークフロー

システムの新しい機能を実装する必要があるとしましょう。私には、インタラクティブモードとバッチモード（別名「撃ちっぱなし」モード）という2つの主要な操作モードがあります。どちらを選ぶかは、特定の機能の実装方法に関する`技術的な確実性`と、現時点でそれにどれだけ投資しているか（`ビジネス価値`）に大きく依存します。

![優先順位付けの例](image.png "このブログの優先機能の例")

たとえば、このブログでは、前回ホームページに「特集記事」を実装した方法について書きました。それに取り組み始めたとき、特集記事を実装する方法については何も知りませんでしたが、達成したいことのアイデアはありました。これは、どのテクノロジーを使用すればよいかわからず、コードのどこを変更すればよいかわからなかったため、技術的な確実性が低い問題です。同時に、ブログをよりプロフェッショナルで読者にとって魅力的にするという仮説があったため、私にとってはビジネス価値が高いものでした。

それを考えると、自然な選択は短いフィードバックループを持つことでした。そうすれば、AIが提案するすべての変更に対して、すぐに結果を確認し、正しい方向に導くことができます。一方、優先度の低いものについては、長いフィードバックループで問題ないので、バッチモードに適しています。

## 低/中程度の技術的確実性または高いビジネス価値=インタラクティブモード（同期）

技術的な確実性が低い問題は、物事を正しく行うためにより多くの監督を必要とするため、CLIツールを使用したインタラクティブなプロセスを使用することを好みます。現在の私の選択ツールは[Gemini CLI](https://cloud.google.com/gemini/docs/codeassist/gemini-cli?utm_campaign=CDR_0x72884f69_default_b431747616&utm_medium=external&utm_source=blog)です。これはGoogleによってわずか2週間前にリリースされ、すでに開発の世界を席巻しています。

Gemini CLIは、AI対応のREPL（Repeat Eval Print Loop）のようなコマンドラインアプリケーションです。プロンプトを入力すると、CLIは応答で反応します。これは、コードやテキストだけでなく、Model Context Protocol（MCP）のサポートにより、基本的に何でもかまいません。そのおかげで、コーヒーの購入からデータベースの更新まで、基本的に何でもCLIを使用できます。もちろん、CLIの自然な使用例はコーディングですが、ご存知のように、人々は。:)

Gemini CLIには自動化用に設計されたYOLOモードがありますが、正直なところ、監督なしで物事を行うには十分に信頼できません（これについては後で詳しく説明します）。そのため、解決策を思いつく前に問題空間をブレインストーミングして調査する必要がある場合は、CLIを使用することを好みます。機能の計画、実装オプションの調査、さらにはすぐに実装するように依頼することもあります。ただし、最初の実装から学んだことに基づいて、実装を破棄してクリーンな状態でやり直すだけです。これは「プロトタイピング」とも呼ばれます。

プロンプトを正しく取得するには数回試行する必要があります。手動でコーディングして満足のいく方法で何かをプロトタイプ化するのに数回試行するのと同じです。主な違いは、プロトタイプごとに1週間かかるのではなく、通常30分から1時間かかることです。1日で、問題の3〜4つの異なる実装をカバーでき、1日の終わりには、情報に基づいた決定を下すための多くのデータとともに、1つにコミットする準備が整います。

技術的な確実性が低い問題にCLIを使用する主な理由は、フィードバックループがほぼ即座に閉じるためです。仮説をテストし、粗削りな部分を修正し、反復します。唯一の遅延は、モデルがリクエストを処理する時間です。

## 高い技術的確実性かつ低い/中程度のビジネス価値=バッチモード（非同期）

前述のように、技術的な確実性が高い問題とは、実装に必要なすべての（またはほぼすべての）ステップがすでにわかっている問題です。これにより、すでにステップがわかっている場合は、自分で実行する代わりに、AIに実行するように指示するだけで済むため、生活がはるかに楽になります。

ここでは、YOLOモードでGemini CLIを使用するケースが1つありますが、実際には[Jules](https://jules.google/)というより優れたツールがあります。Julesは、今年のI/OでGoogleによって発表され、すぐに私のお気に入りのツールになりました（Gemini CLIは僅差で2位です）。

Julesは、GitHubに接続してバックグラウンドでタスクを実行できる非同期エージェントです。最初にJulesを発見したとき、細かい部分にあまり注意を払わず、その遅さに少しイライラしたことを告白しますが、しばらくして、タスクをバックグラウンドで実行するように設定して、切断して自分の生活を続けることが要点であることに気づきました。

JulesはGitHubに接続されているため、プロジェクトの完全なコンテキストがすでにあります。そのため、「依存関係のバージョンをアップグレードする」や「単体テストを実装する」、さらには「この特定のバグを修正する」などのメンテナンスタスクを実行するように依頼できます。重要な注意点は、フィードバックループが長いため（すぐに結果が得られないため）、段階的に何をすべきか明確になっているタスクにこのツールを予約することをお勧めします。

## 高い技術的確実性かつ高いビジネス価値=インタラクティブモード（同期）

上記のカテゴリに基づいて、高い技術的確実性と高いビジネス価値は常に同期的であると私が特定したことにお気づきかもしれません。その唯一の理由は、この結果（高いビジネス価値）を非常に重視しているため、個人的に監督し、できるだけ早く準備ができていることを確認したいからです。ビジネス価値は、これらのパラメータが同じレベルである場合、常に技術的な確実性よりも優先されるため、私にとっては自然なことです。

## 低い技術的確実性かつ低いビジネス価値=本当にこれをすべきか？

これらは通常、バックログ地獄で失われるものです。AI以前の時代には、私はそれらを忘れていましたが、AI対応の世界では、それらに遭遇した場合、不確実性を減らすか、ビジネス価値を高めることを期待して、いくつかの可能性を探るためにJulesタスクを起動するだけです。認知コストが低いため、考えられる最も単純なプロンプトでJulesタスクを起動しても、文字通り失うものは何もありません。結果は素晴らしいものではないかもしれませんが、とにかくやらないので、そこから得られる良いものはすべて利益です。

## 例外

もちろん、例外のない優れたプロセスはありません。高いビジネス価値のあるタスクをJulesに委任する場合があり、それは通常、そうでなければ実行できない場合です。例えば：
1. イベントに参加していて、新しいアイデアを実装するためにキーボードに手を置くことができないが、スマートフォンは持っている
2. パッチの多い、または遅いインターネットで旅行している
3. スーパーマーケットの列に並んでいて、忘れていたことを思い出し、家に帰ったときにそれを始めるためにキックスタートしたい

要約すると、私の選択肢がゼロの進捗を遂げるか、Julesを起動するかである場合、私はJulesを起動して私のためにいくつかの作業をさせます。

一方、これまでのところ、従来のIDEが私のワークフローにどこで入るかについては触れていません。それらを放棄したわけではありません。実際、現在VS Codeでこのブログ投稿を書いています。IDE/手動編集は、コードの最後のマイルまたは最終的な仕上げのために予約しています。「vibe coding」セッションの途中でマイナーな編集を行う場合は、LLMをすぐに脱線させる傾向があるため、非常に注意してください。しかし、これが最後に行うことである場合は、かなり安全です。

## ボーナス：不確実性を減らすことについてのメモ

トピックを調査する必要がある場合がありますが、コードベースとまだ統合する方法がわからない場合もあります。Gemini CLIやJulesにコーディング以外のタスクを強制的に実行させることもできますが、壁にネジを固定するためにハンマーを使用しているように感じます。この場合、純粋な調査を行う必要がある場合は、代わりに[Gemini Deep Research](https://gemini.google/overview/deep-research/?hl=en-GB)を使用することを好みます。Julesと同様に、Gemini Deep Researchは非同期なので、バックグラウンドで調査をトリガーして、1日を過ごすことができます。

キーボードに手があり、待ちたくない場合は、Google検索でグラウンディングされた通常のGeminiを使用することも驚くべき効果を発揮します。どちらもわずかに冗長な出力を生成する傾向があるため、私のように忍耐力がない場合は、物事をさらに良くする方法は、得られた調査出力を[NotebookLM](https://notebooklm.google/)に入れて、要約したり、外出先で聞くことができるようにポッドキャストを作成したりするように依頼することです。

## 結論

優先順位付け演習に基づくAIツール選択プロセス全体と作業モデルは、次のように要約できます。

![優先順位別の推奨ツールの概要](image-1.png "推奨ツールと作業モードの概要")

1. 高い技術的確実性+高いビジネス価値= Gemini CLIとの同期プロセスまたはペアプログラミング。マイナーな明確化のために検索でグラウンディングされたGeminiを使用します。
1. 低/中程度の技術的確実性+高いビジネス価値= Gemini CLIとの同期プロセスに加えて、技術的確実性を高めるための非同期調査。
1. 高い技術的確実性+低い/中程度のビジネス価値= Julesとの非同期プロセス。必要に応じて詳細な調査。
1. 低い技術的確実性+低いビジネス価値= ほとんどの場合実行しませんが、本当にJulesまたは詳細な調査を使用してパラメータの1つを増やしたい場合。

このプロセスについてどう思いますか？以下にコメントを残してください。
