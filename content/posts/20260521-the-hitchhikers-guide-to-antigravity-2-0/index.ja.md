---
categories:
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

[Google I/O 2026](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/) が幕を閉じた今、発表された数々の新リリースを振り返り、それらが現在および近い将来の開発ワークフローにどう影響するのかを整理してみましょう。多くの魅力的な発表がありましたが、今回は開発者に最も大きなインパクトを与える [Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0) のリリースと、[Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli) や [Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk) を含む拡張された Antigravity（agy）エコシステム（[Google I/O 2026 Antigravity ハイライト](https://antigravity.google/blog/google-io-2026) を参照）にフォーカスします。

技術的な詳細に入る前に、今回のローンチに関してネット上で多くの騒ぎ（残念ながら好意的なものばかりではありません）が起きていることに触れておく必要があります。その最大の理由は、Antigravity 2.0 がメインの Antigravity デスクトップアプリから IDE 環境を切り離したことをはじめ、開発フローの多くの面で破壊的変更（breaking changes）を導入したためです。

さらに、Antigravity CLI への移行に伴う [Gemini CLI の非推奨化のアナウンス](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) も、移行に与えられた猶予期間の短さ（加えて後述するいくつかの癖）が原因で、コミュニティではあまり歓迎されませんでした。基本的にユーザーには 2026 年 6 月 18 日までの移行期限が与えられており、I/O から実質 1 か月しかありません。率直に言って、これは決して十分な期間とは言えません。

これについては以前にも書きましたが、愛用していたプロダクトが非推奨・終了になる悔しさは痛いほどよくわかります。私自身、Gmail との争いに敗れ、すっかり忘れ去られてしまったメールクライアントである Google Inbox を今でも惜しんでいます。綺麗事を言うつもりはありません。Google には優れたプロダクトを終了させてきた評判（悪名）が確かにあります。しかし、個人的な好みを別にして大局的に見れば、自らのプロダクトをこれほど大胆に幕引きできる Google の姿勢には、むしろ感銘を受けています。

多くの人は Google に対して、テクノロジーに関連するあらゆる領域で創造的破壊（ディスラプション）を牽引することを期待していると思います。特に AI の進歩によって環境がめまぐるしく変化する今日においては、ある方向から別の方向へと舵を切る（ピボットする）のには多大な勇気と決断力が必要です。私は普段からアジャイルについて多くを語っていますが、Google は形式ばったアジャイル手法とは結びつけられにくいものの、経験豊かなアジリストなら誰もが組織にとって最も価値ある特質の一つとして認める特徴を備えています。それは、素早く進路を修正し、ピボットし、実験し、失敗から学び、反復（イテレーション）を繰り返す能力です。

現状維持（コンフォートゾーン）に甘んじることなく、自らを再発明し続ける力こそが、Google を常に最前線に立たせている理由です。すべての実験が成功するわけではなく、むしろ多くの実験が失敗すること自体が織り込み済みです。そうして何が機能し、何が機能しないかを学んでいきます。得られた教訓を胸に次の目標へと進み、より新しいプロダクトへと反映させていくのです。

今回のリリースからも多くの教訓が得られるはずですが、最終的にテクノロジーそのものを見つめれば、目指しているゴール（エンドゲーム）は自ずと明らかになるでしょう。より高度なプロダクトを構築するためにリソースを集約しつつ、私たちは本格的なエージェンティック（自律型エージェント）時代へと舵を切っているのです。

## 新しい Antigravity デスクトップアプリの解説

デスクトップアプリにおける最大の変更点は、IDE コンポーネントの削除です。Antigravity 1.x ではアプリが VS Code のフォークをベースにしていたため、コードのナビゲーションや編集を行うおなじみの IDE 機能と、エージェントと対話するためのアシスタントボックスが一体となっていました。

それだけでなく、「Agent Manager」と呼ばれるセカンダリ UI も用意されており、複数のチャットセッション（いわゆる「conversations」）を高い視点から俯瞰できました。このビューでエージェントを監視し、入力待ちになったときに対応することで、多数のプロジェクトを並行して進めることができたのです。

新しいデスクトップアプリの最大の変化は、Antigravity 2.0 がこの Agent Manager 体験をメインに据え、IDE 部分を完全に切り離した点です（IDE 部分は独立したオプションのアプリとなりました）。

![新しい Agent Manager のインターフェース](image.png "プロジェクトと会話に特化した、クリーンな新しい Agent Manager インターフェース")

ベテラン開発者にとって、これは非常に大きな摩擦（フリクション）となりました。長年頼りにしてきた使い慣れたエディタツールが、突如として消え去ってしまったからです。agy 2.0 の UI 上でもファイルを見ることはできますが、それは agy が現在作業しているファイルに限られ、直接編集することはできません。すべての操作は、プロンプトまたはファイルへのアノテーション（注釈）を通じて行われます。

![agy 2.0 のファイルビュー](image-2.png "UI 上でファイルを確認できますが、直接編集することはできません")

エージェントとの対話フロー自体は、過去 1 年間にエージェントとともにコーディングしてきた人ならすでに馴染みのあるものでしょう。プロンプトを与えると実装計画（implementation plan）が作成され、インラインコメントやトップレベルのプロンプトを使ってそれをレビューできます。承認されると、エージェントは自律して実行に移ります。UI の設定にもよりますが、時折許可を求めて処理を一時停止することがあり、許可を与えることも、任意の進路修正を添えて却下することもできます。

![ユーザー入力を求める Agent Manager](image-1.png "リクエストを却下する際に、指示を与えるステアリングコメントを追加できます")

拡張性の面では、agy 2.0 は MCP や Skills といったこの 1 年で広く親しまれるようになった共通標準規格をサポートしているほか、1.x から引き継がれた独自の「Rules」メカニズム（本質的にはコンポーザブルな AGENTS.md）や、旧 Gemini CLI の拡張機能をベースにした新しいプラグインシステムも備えています。プラグインを使用すると、追加ルール、スラッシュコマンド、MCP サーバー、スキル、サブエージェントをまとめてパッケージ化でき、Gemini CLI の拡張機能との後方互換性も維持されています（つまり、CLI 拡張機能を agy にインストールできますが、その逆はできません）。

全体として、IDE の統合を惜しむ人たちのフラストレーションは理解できるものの、私個人の第一印象としては、**同じ**アプリ内に IDE が統合されていなくても全く困りませんでした。Gemini CLI を使っていた頃から、手動で編集したいときのために常に VS Code を並行して立ち上げていましたし、agy 2.0 でも全く同じワークフローを適用しています。実際、最近の私は VS Code をほぼテキストエディタとして使っており、本格的な IDE 機能は滅多に使いません。メモ帳に変えても大して変わらないくらいですが、いくつかのショートカットがマッスルメモリーとして染み付いていることだけが、今でも VS Code を使い続けている唯一の理由です。

正直に言って、agy 1.x や他のコーディングエージェントと比較して agy 2.0 に何か画期的な変化があるわけではありませんが、よりすっきりとした外観はとても気に入っていますし、自作プラグインによるカスタマイズを始めてこそ真価を発揮するだろうと考えています。現在、[GoDoctor]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}}) と [Speedgrapher]({{< ref "/posts/20250805-introducing-speedgrapher" >}}) を Gemini CLI 拡張機能から agy プラグインへとアップグレードする作業を進めており、何かお見せできるものができ次第、報告したいと思います。

## Antigravity CLI

ターミナルユーザー向けには、コマンドライン体験が新しい [**Antigravity CLI**](https://antigravity.google/blog/introducing-google-antigravity-cli)（通称 `agy CLI`）として再構築されました。最初は少し戸惑うかもしれませんが、CLI だけを使用する予定であっても、認証プロセスを共有しているため agy 2.0 アプリをインストールしておく必要があります。agy CLI は Gemini CLI の正当な後継であり、100% の機能一致とまではいかないものの、主要な機能はすでにしっかりと揃っています。具体的には、[hooks]({{< ref "/posts/20260610-mastering-hooks" >}})、[skills]({{< ref "/posts/20260829-the-pragmatic-guide-to-agent-skills" >}})、[MCP]({{< ref "/posts/20250817-hello-mcp-world" >}})、[サブエージェント]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}})、そしてプラグインです。

CLI 全体が Go で書き直されたこと（Gemini CLI は TypeScript 製でした）は、よりキビキビとした動作が期待できるため、私を大いに喜ばせてくれました。一方で、最大の批判点の一つは、現時点で agy CLI がクローズドソースであることです。これはオープンだった Gemini CLI からの後退のように感じられるかもしれません。少し前までは Gemini CLI のコードが一般に「流出」したことについて冗談を言い合っていましたが、今や私たちのメインのコーディングエージェント自体がクローズドソースになってしまったため、残念ながらこのジョークは笑えないものになってしまいました。

私自身にはどうすることもできないため、気に病まないことに決めました。これが良い決断か悪い決断かを判断するのは時期尚早ですが、以前 Gemini CLI に貢献した方々をはじめ、コミュニティの悔しさは痛いほどよくわかります。せめてもの救いは、プラグインシステムを中心に活発なオープンソースコミュニティが今後も発展していくであろう点です。少なくとも私自身は、頼もしい Go エキスパート・サブエージェントと vibe-writing コンパニオンを近いうちに皆さんにお届けできるよう、自分のプラグイン開発に取り組んでいます。

![agy CLI のインターフェース](image-3.png "Gemini CLI や Claude Code を使ってきた人なら違和感なく馴染める UI")

UI については、これまでに CLI コーディングエージェントを使ったことがある人なら驚くようなことはないでしょう。第一印象として、ターミナルの描画は Gemini CLI の TypeScript による描画よりも確かに優れていると感じられます。また、agy 2.0 と同様に、すっきりとした外観も非常に好印象です。個人的には、Gemini CLI は機能が増えすぎて UI も肥大化しすぎていたと感じていたので、このクリーンなインターフェースはとても新鮮に感じられます。「Less is more（少ないことは、より豊かなこと）」は私の大好きな言葉の一つですが、agy CLI はこの点において期待に応えてくれています。

（今のところ）あまりうまく機能していない点としては、主に拡張機能との互換性が挙げられます。移行パスは用意されているものの、必ずしも期待通りに動作するとは限らず、私が今週のほとんどの時間を費やして godoctor と speedgrapher の書き直しに専念してきたのも、自動移行に頼りたくなかったからです。それに加えて、プロジェクトベースの認証にも問題が発生しており、これについては近いうちに修正されることを期待しています。現在のところ、プロジェクトベースの認証が動作しなかったため、個人の Google Pro サブスクリプションで使用しています。

Gemini CLI からの移行ユーザーにとって頭の痛いもう一つの問題である請求周りの複雑さには立ち入りませんが、私個人の見解としては、agy CLI にはいくつかの課題があるものの、素晴らしい可能性も秘めています。今のところ革新的な部分は多くありませんが（変更の大部分は内部で行われているため）、致命的な欠点（ディールブレイカー）も見当たりません。Gemini CLI で行っていたことはすべて agy CLI でも可能であり、新たに学ぶべきこともごくわずかです。そのため、仮に Gemini CLI の移行猶予期間がもっと長かったとしても、ワークフローの将来性を確保するために、できるだけ早く移行することをおすすめします。

## Antigravity SDK

ここまでの議論は古いプロダクトを新しいものに置き換える話が中心であり、革新的というよりは漸進的なアップデートのように感じられたかもしれません。だからこそ、[**Antigravity SDK**](https://antigravity.google/blog/introducing-google-antigravity-sdk) のリリースは私にとって最も刺激的な発表でした。変更の大部分が内部で行われていると述べましたが、それこそがエージェントを支援するこの統合プラットフォームの構築であり、Antigravity SDK は開発者である皆さんがその基盤にアクセスするための手段なのです。

以下は、わずか 15 行足らずのコードでワークスペースを問い合わせるエージェントの実装例です。

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

この [Python](https://xkcd.com/353/ "import antigravity") ライブラリにより、開発者はまったく同じエージェント向けランタイムとオーケストレーション・ハーネスにプログラムからアクセスできるようになります。この SDK はランタイム非依存であり、15 行足らずのコードでステートフルなエージェントループを立ち上げることが可能です。組み込みツール、カスタム関数、Model Context Protocol サーバー、サブエージェント、再利用可能なスキルなど、モジュール化された機能を単一のパイプラインの下でサポートします。

## はじめに

Antigravity に関連するすべてのリリースに共通するトレンドは、「コードファースト」から「デザインファースト」へのシフトです。ソフトウェア開発体験全体が、コードを編集することからエージェントを調整・協調させることを中心に再設計されています。この変化に向けて開発環境を準備するために、次のアクションを検討してみてください。

1.  **デスクトップアプリをダウンロードする**: [antigravity.google](https://antigravity.google) にアクセスしてデスクトップアプリケーションをインストールします。
2.  **ターミナルのワークフローを移行する**: `agy` CLI をインストールし、インポートコマンドを実行して、**2026 年 6 月 18 日**の非推奨期限までに Gemini CLI の設定を移行します（詳細は [移行アナウンス](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) を参照）。
3.  **SDK を試してみる**: Python ライブラリをインストールし、[Antigravity の機能](https://antigravity.google/docs/features) をチェックして、agy SDK を使用したカスタムエージェントの構築を始めましょう。
    ```bash
    pip install google-antigravity
    ```

## 参考リソース
今回のリリースについてさらに詳しく知り、追加の技術情報を確認するには、以下のリソースをチェックしてください。
* **[Introducing Google Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0)**: 2.0 エコシステムの公式アナウンス。
* **[Introducing Google Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli)**: Go 製の新しいターミナルインターフェースに関する詳細解説。
* **[An Important Update: Transitioning Gemini CLI to Antigravity CLI](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/)**: Gemini CLI ユーザー向けの詳細な移行スケジュールとガイドライン。
* **[Introducing Google Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk)**: Python でエージェントをプログラムからオーケストレーションする方法。
* **[Google I/O 2026 Developer Highlights](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/)**: 今年の Google I/O における主要な開発者向け発表。
* **[Google I/O 2026: Antigravity Announcement](https://antigravity.google/blog/google-io-2026)**: Google I/O における Antigravity の重要な最新情報とハイライト。
* **[Google Antigravity Documentation & Features](https://antigravity.google/docs/features)**: Antigravity の機能と安全性コントロールに関する包括的ガイド。
