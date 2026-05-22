---
title: "Antigravity 2.0 への銀河ヒッチハイク・ガイド"
date: 2026-05-21T11:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["antigravity", "agy-cli", "agy-sdk", "google-io", "agentic-coding"]
summary: "Google I/O 2026 で発表された Google Antigravity 2.0 エコシステムのガイドです。スタンドアロンのデスクトップアプリケーション、Go ベースのターミナル CLI、およびプログラムによる Python SDK について検証します。"
heroStyle: "big"
---

[Google I/O 2026](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/) が無事に終了した今、新しいリリースやそれらが現在および近い将来のワークフローにどのように影響するのかを整理して分析する時期が来ました。多くの興味深い発表がありましたが、今日フォーカスしたいのは、開発者に最も大きな影響を与える [Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0) のリリースと、[Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli) や [Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk) を含む拡張された Antigravity (agy) エコシステム (詳細は [Google I/O 2026 Antigravity highlights](https://antigravity.google/blog/google-io-2026) を参照) です。

技術的な詳細に入る前に、今回のリリースによってウェブ上で多くのノイズ（残念ながら好意的なものばかりではありません）が発生している点に触れておく必要があります。その主な理由は、Antigravity 2.0 が主要な Antigravity デスクトップアプリから IDE 環境を分離したことをはじめ、開発フローの多くの側面で破壊的変更 (breaking changes) を導入したためです。

第二に、Gemini CLI の廃止と Antigravity CLI への移行の [アナウンス](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) も、ユーザーに与えられた移行期限が非常に短かったため（および後述するいくつかの癖のため）、好意的に受け止められませんでした。本質的に、ユーザーは2026年6月18日までの移行を求められており、これは I/O から実質的にわずか1ヶ月しかなく、率直に言って十分な期間とは言えません。

これについては以前にも書きましたが、お気に入りのプロダクトが廃止されるときの不満はよく理解できます。例えば、私自身、Gmail との競争に敗れて久しく忘れ去られたメールクライアントである Google Inbox を今でも懐かしんでいます。ここで取り繕うつもりはありません。Google には確かに優れたプロダクトを廃止してきたという定評があります。しかし、個人的な好みを脇に置いて大局的に見れば、Google が大胆にプロダクトを打ち切るその決断力には、実は敬意を抱いています。

多くの人々は、Google があらゆるテクノロジー関連の分野において破壊的イノベーションをリードすることを期待しています。特に AI の進歩による今日のダイナミックな環境においては、ある方向から別の方向へとピボットするには大きな勇気と決断力が必要です。私は Agile についてよく話しますが、Google は一般的な形式的 Agile 手続きとは結びつけられないものの、経験豊富なアジャイル推進者が最も価値ある企業の特性として認識する能力を体現しています。それは、迅速な進路修正、ピボット、実験、失敗からの学習、そしてイテレーションを素早く行う能力です。

コンフォートゾーンにとどまるのではなく、自分自身を絶えず再定義し続ける能力こそが、Google を常に最先端に留めている要因です。たとえすべての実験が成功するわけではなくても、むしろ多くの実験が失敗することは当然想定されています。そこから何が機能し、何が機能しないのかを学びます。その教訓を活かして次の目標へ進防、より新しいプロダクトにそれを組み込んでいくのです。

今回のリリースからは多くの教訓が得られるでしょうが、最終的に技術そのものに目を向ければ、そのエンドゲームが明確になるはずです。私たちはより高度なプロダクトを構築するための取り組みを統合しつつ、エージェンティック（自律的エージェント）時代に注力しているのです。

## 新しい Antigravity デスクトップアプリの解説

デスクトップアプリにおける最大の変更点は、IDE コンポーネントの削除です。Antigravity 1.x では、アプリは VS Code のフォークに基づいて構築されていたため、コードのナビゲートや編集を行う使い慣れた IDE 機能と、エージェントと対話するためのアシスタントボックスが一体となっていました。

それだけでなく、「Agent Manager」と呼ばれるセカンダリ UI も用意されており、さまざまなチャットセッション（別名「conversations」）を俯瞰して見ることができました。これにより、このビューでエージェントを監視し、入力待ち状態になった際に対応することで、複数のプロジェクトを並行して進めることが可能でした。

新しいデスクトップアプリの最大の変化は、Antigravity 2.0 がエージェントマネージャーの体験を前面かつ中心に据え、IDE 部分を完全に削除した点です（IDE 部分は独立したオプションのアプリとなりました）。

![新しい Agent Manager インターフェース](image.png "新しいエージェントマネージャーのインターフェースはよりクリーンで、プロジェクトと会話に焦点を当てています")

経験豊富な開発者にとって、これは大きな摩擦点となりました。長年頼りにしてきた使い慣れたエディタツールが突然すべて失われてしまったからです。agy 2.0 の UI 上でファイルを確認することは依然として可能ですが、それは agy が現在作業しているものに限られ、直接編集することはできません。すべてのインタラクションはプロンプトまたはファイル上のアノテーションを介して行われます。

![agy 2.0 のファイルビュー](image-2.png "UI上でファイルを見ることはできますが、直接編集することはできません")

エージェントとのやり取りは、この1年でエージェントによるコーディングを行ってきた人にはすでにお馴染みのものでしょう。プロンプトを与えると、エージェントは実装計画（implementation plan）を作成し、インラインコメントまたはトップレベルのプロンプトでそれをレビューできます。承認されると、エージェントは自律的に実行に移ります。UI の設定によっては、エージェントが時折、処理の実行許可を求めてくることがあり、ユーザーはそれを許可するか、オプションで指示を追加して却下（コース修正）することができます。

![ユーザーの入力を求める Agent Manager](image-1.png "リクエストを却下する際にステアリングコメントを追加できます")

拡張性の面において、agy 2.0 は MCP や Skills を含むこの1年で普及した一般的な標準規格に加え、1.x から引き継がれた独自の「Rules」メカニズム（本質的には構成可能な AGENTS.md）や、旧 Gemini CLI の拡張システムに基づいた新しいプラグインシステムをサポートしています。プラグインを使用すると、追加のルール、スラッシュコマンド、MCP サーバー、スキル、およびサブエージェントをまとめてパッケージ化でき、Gemini CLI 拡張機能との後方互換性も維持されています（つまり、CLI 拡張機能を agy にインストールすることはできますが、その逆はできません）。

全体として、IDE がなくなることを惜しむ人々の不満は理解できますが、私自身の第一印象としては、**同じ**アプリに IDE が内蔵されていなくても困ることはありません。Gemini CLI を使用していたときも、手動で編集を行いたいときのために常に VS Code を並行して実行しており、この流れは agy 2.0 でも同じです。実際のところ、最近では VS Code を主にテキストエディタとして使用しており、まともな IDE 機能はほとんど使っていません。メモ帳に置き換えても大きな違いはなく、いくつかのショートカットのキーボードショートカットの記憶が失われることくらいが、今でも VS Code を使い続けている唯一の理由です。

agy 2.0 自体には、agy 1.x や他のコーディングエージェントと比較して画期的な点はないと言わざるを得ませんが、このすっきりとした外観は非常に気に入っています。そして、独自のプラグインでカスタマイズを始めたときにこそ、その真の力を引き出せるのだと信じています。現在、godoctor と speedgrapher を Gemini CLI 拡張機能の形態から agy プラグインへアップグレードする作業を進めており、成果が出次第また報告します。

## Antigravity CLI

ターミナルユーザー向けには、コマンドライン体験が新しい [**Antigravity CLI**](https://antigravity.google/blog/introducing-google-antigravity-cli) (別名 `agy CLI`) として再構築されています。最初は少し混乱するかもしれませんが、CLI のみを使用する予定であっても、同じ認証プロセスを共有しているため、agy 2.0 アプリをインストールする必要があります。agy CLI は Gemini CLI の自然な後継であり、100% の機能一致は提供していませんが、主要な機能である hooks、skills、MCP、sub-agents、プラグインなどはすでに揃っています。

CLI 全体が Go で再書き換えられたため (Gemini CLI は TypeScript でした)、より軽快な動作が期待できる点には大いに満足しています。一方で、現在 agy CLI がクローズドソースであることは主要な批判の一つであり、Gemini CLI からの退行と感じられるかもしれません。少し前までは Gemini CLI のコードを一般に「リーク」させるという冗談を交わしていましたが、今や主要なコーディングエージェント自体がクローズドソースになってしまい、この冗談は皮肉にも笑えなくなってしまいました。

これについて自分には一切コントロールできないことを考慮し、私も気にしないことに決めました。これが良い決断であるか悪い決断であるかを判断するのは時期尚早ですが、これまで Gemini CLI に貢献してきたコミュニティのフラストレーションは理解できます。慰めになるとすれば、プラグインシステムを取り巻くオープンソースコミュニティは今後も活発であり続けるということです。少なくとも私は、実用的な Go のエキスパートサブエージェントと vibe-writing コンパニオンを近いうちに提供できるよう、自身の開発に取り組んでいます。

![agy CLI インターフェース](image-3.png "Gemini CLI や Claude Code から移行したユーザーにとって使い慣れた UI です")

UI については、これまでに CLI コーディングエージェントを使用したことがある人なら驚くことはないでしょう。最初の印象として、レンダリング品質は確かに Gemini CLI の TS によるものより向上していると感じます。また、agy 2.0 と同様に、よりシンプルな外観も非常に評価しています。個人的には、Gemini CLI は機能が増えすぎて UI も肥大化しつつあったため、このすっきりとしたインターフェースは新鮮に感じられます。「Less is more (より少ないことは、より豊かなこと)」は私の好きな格言の一つですが、agy CLI はまさにこの点において優れています。

現時点で（今のところ）不十分な点は、主に拡張機能との互換性に関連しています。移行パスは用意されていますが、常に期待通りに機能するとは限らないため、私は自動移行に頼らず、今週のほとんどを godoctor と speedgrapher の書き換えに費やしてきました。次に、プロジェクトベースの認証にも問題があり、これについては早期の修正を期待しています。現在のところ、プロジェクトベースの認証が機能しなかったため、Google Pro サブスクリプションで使用しています。

Gemini CLI から移行するユーザーにとってのもう一つの懸念事項である課金の複雑さについてはここでは触れませんが、私の個人的な意見としては、agy CLI にはいくつかの問題があるものの、素晴らしい可能性を秘めていると思います。現時点では（変更の多くが内部的なものであるため）革新的な要素はありませんが、一方で重大な欠点も見当たりません。Gemini CLI で行っていたことはすべて agy CLI でも実行可能であり、新しく学ぶべきこともほとんどありません。したがって、Gemini CLI の移行猶予期間がもっと長かったとしても、ワークフローを将来に備えておくために、できるだけ早い移行をお勧めします。

## Antigravity SDK

これまでの議論は古いプロダクトを新しいものに置き換える話が中心で、これは革命的というよりは段階的な進歩のように感じられます。そのため、[**Antigravity SDK**](https://antigravity.google/blog/introducing-google-antigravity-sdk) の発表こそが私にとって最もエキサイティングなものでした。水面下で多くの変化が起きていると述べたのは、エージェントを支援するこの統一されたプラットフォームの作成についてであり、開発者がそれを利用するための手段こそが Antigravity SDK なのです。

以下は、エージェントが 15 行未満のコードでワークスペースを照会する実用的な例です。

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

この Python ライブラリを使用すると、開発者は同じエージェントランタイムおよびオーケストレーションハネースにプログラムからアクセスできます。SDK はランタイムに依存せず、15 行未満のコードでステートフルなエージェントループを立ち上げることが可能です。組み込みツール、カスタム関数、Model Context Protocol サーバー、サブエージェント、再利用可能なスキルなど、モジュール化された機能を統一されたパイプラインの下でサポートしています。

## はじめに

Antigravity に関連するすべてのリリースにおける共通のトレンドは、コードファーストからデザインファーストへの移行です。ソフトウェア開発体験全体が、コードを編集することからエージェントを調整することを中心に再設計されています。このシフトに備えて開発環境を準備するには、以下の手順を検討してください。

1.  **デスクトップアプリのダウンロード**: [antigravity.google](https://antigravity.google) にアクセスして、デスクトップアプリケーションをインストールします。
2.  **ターミナルワークフローの移行**: Gemini CLI の非推奨化期限である **2026年6月18日** より前に、`agy` CLI をインストールしてインポートコマンドを実行し、Gemini CLI の設定を移行します (詳細は [migration announcement](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) を参照)。
3.  **SDK の探索**: Python ライブラリをインストールし、[Antigravity の機能](https://antigravity.google/docs/features) を確認して、agy SDK を使用したカスタムエージェントの構築を開始します：
    ```bash
    pip install google-antigravity
    ```

## 追加リソース

このリリースについて詳しく知り、追加の技術情報にアクセスするには、以下のリソースを参照してください：
* **[Introducing Google Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0)**: 2.0 エコシステムの公式発表。
* **[Introducing Google Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli)**: 新しい Go ベースのターミナルインターフェースの詳細。
* **[An Important Update: Transitioning Gemini CLI to Antigravity CLI](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/)**: Gemini CLI ユーザー向けの移行タイムラインとガイドラインの詳細。
* **[Introducing Google Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk)**: Python でプログラムからエージェントをオーケストレーションする方法を学びます。
* **[Google I/O 2026 Developer Highlights](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/)**: 今年の Google I/O における主要な開発者向け発表。
* **[Google I/O 2026: Antigravity Announcement](https://antigravity.google/blog/google-io-2026)**: Google I/O における Antigravity の主な最新情報とハイライト。
* **[Google Antigravity Documentation & Features](https://antigravity.google/docs/features)**: Antigravity の機能と安全対策に関する包括的ガイド。
