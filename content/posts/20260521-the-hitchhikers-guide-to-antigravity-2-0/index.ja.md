---categories:
- Agentic Coding
date: 2026-05-21 11:00:00+00:00
heroStyle: big
summary: Google I/O 2026 で発表された Google Antigravity 2.0 エコシステムの包括ガイド。スタンドアロンのデスクトップアプリ、Go 製のターミナル CLI、プログラムから操作可能な Python SDK を詳しく解説します。
tags:
  - antigravity
  - cli
  - google-cloud
  - python
  - sdk
title: "Antigravity 2.0 への銀河ヒッチハイク・ガイド"
slug: "the-hitchhikers-guide-to-antigravity-2-0"
aliases:
  - "/ja/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0/"
description: "Google I/O 2026 で発表された Antigravity 2.0 の完全解説。Agent Manager 特化デスクトップ、Go製 agy CLI、Python SDK によるプログラム制御まで網羅。"
proficiencyLevel: "Intermediate"
dependencies:
  - "Google Antigravity 2.0"
  - "google-antigravity Python SDK"
  - "Go 1.22+"
---

[Google I/O 2026](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/) が幕を閉じた今、発表された数々の新リリースを振り返り、それらが現在および今後の開発ワークフローにどう影響するのかを整理してみたいと思います。多くの魅力的な発表がありましたが、今回は開発者に最も大きなインパクトを与える [Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0) のリリースと、[Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli) や [Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk) を含む Antigravity（agy）エコシステムの拡張（詳細は [Google I/O 2026 Antigravity ハイライト](https://antigravity.google/blog/google-io-2026) を参照）にフォーカスします。

技術的な詳細に入る前に、今回のローンチをめぐってネット上で多くの議論（残念ながらネガティブなものも少なくありません）が巻き起こっていることに触れておく必要があります。その最大の理由は、Antigravity 2.0 がメインの Antigravity デスクトップアプリから IDE 環境を切り離したことを筆頭に、開発フローの様々な面で破壊的変更を導入したからです。

さらに、Antigravity CLI への移行に伴う [Gemini CLI の非推奨化のアナウンス](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) も、移行までの猶予期間が短すぎること（そして後述するいくつかの癖があること）から、コミュニティで波紋を呼びました。具体的には、移行期限は 2026 年 6 月 18 日に設定されており、I/O から実質 1 か月しかありません。率直に言って、これはあまりに短すぎます。

これについては以前にも書きましたが、愛用していたプロダクトが非推奨・終了になる悔しさは痛いほどよくわかります。私自身、Gmail との争いに敗れ、すっかり過去のものとなってしまった Google Inbox を今でも惜しんでいます。綺麗事を言うつもりはありません。Google には優れたプロダクトを惜しげもなく終了させてきた実績（悪名）があります。しかし、個人的な好みを抜きにして大局的に見れば、Google がこれほど大胆にプロダクトの幕引きを行えること自体、実は賞賛に値すると思っています。

多くの人は Google に対して、テクノロジーのあらゆる領域でディスラプション（創造的破壊）をリードすることを期待しています。そして AI の進化によって環境が急速に変化する今日においては、ある方向から別の方向へと舵を切る（ピボットする）のには多大な勇気と決断力が必要です。私は普段からアジャイルについてよく語っていますが、Google は形式ばったアジャイル開発手法とは結びつけられにくいものの、経験豊かなアジリストなら誰もが組織として最も価値があると認める特質を備えています。それは、迅速に進路修正し、ピボットし、実験し、失敗から学び、高速にイテレーションを回す能力です。

現状維持（コンフォートゾーン）に甘んじることなく、自らを再発明し続ける力こそが、Google を常に最前線に立たせている理由です。すべての実験が成功するわけではなく、むしろ多くの実験が失敗すること自体が織り込み済みです。そうして何が機能し、何が機能しないかを学び取ります。その教訓を胸に次の目標へと進み、より新しいプロダクトへと反映させていくのです。

今回のリリースからも多くの教訓が得られるはずですが、最終的にテクノロジーそのものを見つめれば、目指しているゴール（エンドゲーム）は自ずと明らかになるでしょう。より高度なプロダクトを構築するためにリソースを集約しつつ、私たちは本格的なエージェンティック（自律型エージェント）時代へと舵を切っているのです。

## 新しくなった Antigravity デスクトップアプリを解説

デスクトップアプリにおける最大の変更点は、IDE コンポーネントの完全な削除です。Antigravity 1.x のアプリは VS Code のフォークをベースにしていたため、コードのナビゲーションや編集を行う使い慣れた IDE 機能と、エージェントと対話するためのアシスタントパネルが一体となっていました。

それだけでなく、「Agent Manager」と呼ばれるセカンダリ UI も用意されており、複数のチャットセッション（いわゆる「会話 / conversations」）を俯瞰して一括管理できました。このビューで各エージェントの進捗をモニタリングし、ユーザー入力待ちになった際に応答することで、多数のプロジェクトを並行して進められたのです。

新しいデスクトップアプリの最大の変化は、Antigravity 2.0 がこの Agent Manager 体験をメインに据え、IDE 部分を完全に切り離した点です（IDE 部分は別個のオプショナルなアプリになりました）。

![新しい Agent Manager のインターフェース](image.png "プロジェクトと会話（conversations）に特化した、クリーンな新しい Agent Manager インターフェース")

ベテラン開発者にとって、これは非常に大きな戸惑い（フリクション）となりました。長年愛用してきた使い慣れたエディタツールが、ある日突然姿を消してしまったからです。agy 2.0 の UI 上でもファイルを見ることはできますが、agy が現在作業中のファイルに限られ、直接編集することはできません。すべての操作は、プロンプトやファイルへのアノテーション（注釈）を通じて行います。

![agy 2.0 のファイルビュー](image-2.png "UI 上でファイルを確認できますが、直接編集することはできません")

エージェントとの対話フロー自体は、この 1 年で AI エージェントによるコーディングを経験した方ならすでに馴染み深いものでしょう。プロンプトを投げると、エージェントは実装計画（implementation plan）を組み立てます。ユーザーはインラインコメントやトップレベルのプロンプトで計画をレビューし、承認すればエージェントが自律的に実行を開始します。UI の設定次第では、途中で実行許可を求めてエージェントから確認が入ることもあり、許可するか、あるいは軌道修正の指示（コースコレクション）を添えて却下することができます。

![ユーザー入力を求める Agent Manager](image-1.png "リクエストを却下する際に、軌道修正のためのステアリングコメントを追加できます")

拡張性の面では、agy 2.0 は [MCP (Model Context Protocol)](https://modelcontextprotocol.io) や [Agent Skills](https://agentskills.io) といったこの 1 年で定着した標準規格に加え、1.x から引き継がれた独自の「Rules」メカニズム（本質的にはコンポーザブルな [AGENTS.md](https://agents.md)）や、旧 Gemini CLI の拡張システムをベースにした新しいプラグインシステムをサポートしています。プラグインを使うことで、追加ルール、スラッシュコマンド、MCP サーバー、スキル、サブエージェントをひとまとめにパッケージ化できます。また Gemini CLI 拡張機能との後方互換性も保たれているため、Gemini CLI 向けの拡張機能を agy にインストールすることが可能です（逆は不可）。

全体として、IDE の統合を惜しむ人たちのフラストレーションは理解できるものの、私個人のファーストインプレッションとしては、**同じ**アプリ内に IDE が統合されていなくても全く困りませんでした。Gemini CLI を使っていた頃から、手動でコードをいじりたい時のために常に VS Code をバックグラウンドで立ち上げていましたし、agy 2.0 でも全く同じワークフローをとっているからです。実際、最近の私は VS Code をほぼ単なるテキストエディタとして使っており、本格的な IDE 機能は滅多に触りません。メモ帳に変えても大して変わらないくらいですが、いくつか身体に染み付いたキーボードショートカット（マッスルメモリー）があるからこそ、今でも VS Code を使い続けているにすぎません。

正直に言えば、agy 2.0 は 1.x や他のコーディングエージェントと比べて何かが劇的に革新されたわけではありません。それでも、このすっきりとした外観はとても快適ですし、自作プラグインでカスタマイズを重ねていくことで真価を発揮するはずだと感じています。現在、`godoctor` と `speedgrapher` を Gemini CLI 拡張機能から agy プラグインへとアップグレードする作業を進めており、形になり次第またレポートしたいと思います。

## Antigravity CLI

ターミナル派の開発者に向けては、コマンドライン体験が新しい [**Antigravity CLI**](https://antigravity.google/blog/introducing-google-antigravity-cli)（通称 `agy CLI`）として再構築されました。最初は少し戸惑うかもしれませんが、CLI だけを使う予定であっても、認証プロセスを共有しているため agy 2.0 アプリをインストールしておく必要があります。agy CLI は Gemini CLI の正当な後継であり、100% 完全に同一の機能セットではないものの、hooks、skills、MCP、サブエージェント、プラグインといった主要な機能はすでにしっかりと網羅されています。

CLI 全体が Go でゼロから書き直された点（Gemini CLI は TypeScript でした）は、Go 好きの私にとってこの上なく嬉しいニュースであり、よりキビキビとした軽快な動作が期待できます。その一方で、最大の批判点となっているのが、現時点で agy CLI がクローズドソースであることです。これはオープンだった Gemini CLI からの後退（デグレ）と感じられるかもしれません。少し前までは Gemini CLI のコードが公開された件で冗談を言い合っていましたが、今やメインのコーディングエージェントがクローズドソースになってしまったため、このジョークも笑えない皮肉となってしまいました。

とはいえ、自分にどうこうできる問題でもないため、あまり気に病まないことにしました。これが吉と出るか凶と出るかを判断するのはまだ早いですが、これまで Gemini CLI の開発に貢献してきたコミュニティのやり場のない思いは痛いほど理解できます。せめてもの救いは、プラグインシステムを中心に活発なオープンソースコミュニティが育ち続けるであろう点です。少なくとも私自身は、頼れる Go エキスパート・サブエージェントと vibe-writing コンパニオンを近いうちに皆さんにお届けできるよう、開発を進めています。

![agy CLI のインターフェース](image-3.png "Gemini CLI や Claude Code を使ってきた人なら違和感なく馴染める UI")

UI に関しては、これまでに CLI コーディングエージェントを触ったことがある人なら迷うことはないでしょう。第一印象として、ターミナルの描画（レンダリング）は Gemini CLI の TypeScript 実装よりも明らかに軽快で心地よく感じられます。また agy 2.0 と同様に、すっきりとしたミニマルな見た目も気に入っています。個人的には、Gemini CLI は機能が詰め込まれすぎて UI が少々肥大化（ファットに）なりすぎていたと感じていたので、このクリーンなインターフェースはとても新鮮です。「Less is more（少ない方が豊かである）」は私の大好きな格言ですが、agy CLI はまさにそれを体現しています。

一方で、（現時点で）まだ物足りないと感じるのは、主に拡張機能との互換性周りです。移行パスは用意されているものの、必ずしも期待通りに動くとは限りません。自動移行だけに頼りたくなかった私が、今週のほとんどの時間を費やして `godoctor` と `speedgrapher` の書き直しに専念してきたのはそのためです。さらに、プロジェクトベースの認証（GCP プロジェクト認証）にも問題があり、これについては早急な修正を期待したいところです。今のところ、私の環境ではプロジェクトベースの認証が通らなかったため、Google Pro サブスクリプション経由で利用しています。

Gemini CLI からの移行ユーザーにとって頭の痛いもう一つの問題である課金モデルの複雑さには深入りしませんが、私個人の見解としては、agy CLI はいくつかの課題を抱えつつも、非常に大きな可能性を秘めています。変更の大部分が水面下（アンダー・ザ・フード）で行われているため、表面上は目新しい驚きこそ少ないものの、移行を躊躇するような決定的な欠点（ディールブレイカー）は見当たりません。Gemini CLI でできていたことはすべて agy CLI でも実現可能ですし、新たに学習すべきこともごくわずかです。たとえ Gemini CLI の移行猶予期間がもっと長かったとしても、ワークフローを将来に備えておく（フューチャープルーフにする）ために、できるだけ早く移行することをおすすめします。

## Antigravity SDK

ここまでの話は既存プロダクトのリプレイスが中心で、革新というよりは漸進的なアップデートのように感じられたかもしれません。だからこそ、[**Antigravity SDK**](https://antigravity.google/blog/introducing-google-antigravity-sdk) のリリースこそが、私にとって今回最もワクワクした発表でした。変更の大部分が水面下で行われていると述べましたが、それこそがエージェントを支える統一プラットフォームの構築であり、私たち開発者がその基盤に直接アクセスできるようにする鍵がこの Antigravity SDK なのです。

以下は、わずか 15 行足らずのコードでワークスペースを調査するエージェントの実装例です。

```python
import asyncio
from google.antigravity import Agent, LocalAgentConfig

async def main():
    config = LocalAgentConfig()
    async with Agent(config) as agent:
        response = await agent.chat("What files are in the current directory?")
        print(await response.text())

if __name__ == "__main__":
    asyncio.run(main())
```

この [Python](https://xkcd.com/353/ "import antigravity") ライブラリにより、開発者は同じエージェント・ランタイムとオーケストレーション・ハーネスへプログラムから直接アクセスできるようになります。SDK はランタイム非依存であり、15 行足らずのコードでステートフルなエージェントループを起動できます。組み込みツール、カスタム関数、[Model Context Protocol (MCP)](https://modelcontextprotocol.io) サーバー、サブエージェント、そして再利用可能な [Agent Skills](https://agentskills.io) などのモジュール機能を、単一の統一パイプライン上で連携させることが可能です。

## はじめに

Antigravity にまつわるすべての発表に通底するトレンドは、「コードファースト」から「デザインファースト」へのシフトです。ソフトウェア開発体験全体が、直接コードを編集することから、エージェントをオーケストレーション（調整・協調）することを中心に再設計されています。このパラダイムシフトに向けて開発環境を整えるために、まずは以下のステップから始めてみてください。

1. **デスクトップアプリをダウンロードする**: [antigravity.google](https://antigravity.google) にアクセスし、デスクトップアプリケーションをインストールします。
2. **ターミナルのワークフローを移行する**: `agy` CLI をインストールし、インポートコマンドを実行して Gemini CLI の設定を移行します。Gemini CLI の非推奨期限は **2026 年 6 月 18 日** です（詳細は [移行アナウンス](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) を参照）。
3. **SDK を試してみる**: Python ライブラリをインストールし、[Antigravity ドキュメント](https://antigravity.google/docs/features) をチェックして、agy SDK を使ったカスタムエージェントの構築を始めましょう。
   ```bash
   pip install google-antigravity
   ```

## 参考リソース

今回のリリースに関する詳細や技術ドキュメントについては、以下の公式リソースを参照してください。

* **[Introducing Google Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0)**: 2.0 エコシステム全体の公式アナウンス。
* **[Introducing Google Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli)**: Go 製の新しいターミナルインターフェースのディープダイブ。
* **[An Important Update: Transitioning Gemini CLI to Antigravity CLI](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/)**: Gemini CLI ユーザー向けの移行タイムラインとガイドライン。
* **[Introducing Google Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk)**: Python からプログラムでエージェントをオーケストレーションする方法。
* **[Google I/O 2026 Developer Highlights](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/)**: 今年の Google I/O における主要な開発者向け発表まとめ。
* **[Google I/O 2026: Antigravity Announcement](https://antigravity.google/blog/google-io-2026)**: Google I/O における Antigravity 関連の主なアップデートとハイライト。
* **[Google Antigravity Documentation & Features](https://antigravity.google/docs/features)**: Antigravity の機能群とセーフティコントロールに関する包括的な公式ドキュメント。
