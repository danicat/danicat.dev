---
categories:
- Agentic Coding
date: 2026-05-21 11:00:00+00:00
heroStyle: big
summary: Um guia para o ecossistema Google Antigravity 2.0 anunciado no Google I/O 2026. Analisamos a aplicação desktop independente, a CLI de terminal baseada em Go e o SDK programático em Python.
tags:
  - antigravity
  - cli
  - google-cloud
  - python
  - sdk
title: "O Guia do Mochileiro para o Antigravity 2.0"
slug: "the-hitchhikers-guide-to-antigravity-2-0"
aliases:
  - "/pt-br/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0/"
description: "Guia completo do ecossistema Google Antigravity 2.0 anunciado no Google I/O 2026. Analisa o app desktop Agent Manager, a CLI em Go e o SDK em Python."
proficiencyLevel: "Intermediate"
dependencies:
  - "Google Antigravity 2.0"
  - "google-antigravity Python SDK"
  - "Go 1.22+"
---

Com o encerramento do [Google I/O 2026](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/), chegou a hora de assimilar todos os novos lançamentos e entender como eles afetarão nossos fluxos de trabalho agora e em um futuro próximo. Embora muitas coisas interessantes tenham sido anunciadas, hoje quero focar no que mais impacta desenvolvedores: o lançamento do [Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0) e a expansão do ecossistema Antigravity (agy) (veja os [destaques do Antigravity no Google I/O 2026](https://antigravity.google/blog/google-io-2026)), que inclui a [Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli) e o [Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk).

Antes de entrar nos detalhes técnicos, sinto a necessidade de falar sobre o muito barulho na web causado por esse lançamento e, infelizmente, não do tipo positivo. O principal motivo é que o Antigravity 2.0 introduz breaking changes em vários aspectos do fluxo de desenvolvimento, a começar pela separação do ambiente de IDE do aplicativo desktop principal do Antigravity.

Em segundo lugar, o [anúncio da descontinuação da Gemini CLI](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) em favor da Antigravity CLI também não foi bem recebido devido ao prazo apertado dado aos usuários para migrar de uma para outra (além de algumas peculiaridades que veremos a seguir). Em essência, os usuários têm até 18 de junho de 2026 para migrar — basicamente um mês após o I/O, o que, francamente, não é muito.

Já escrevi sobre isso antes e compreendo como é frustrante quando seu produto favorito é descontinuado. Eu, por exemplo, ainda lamento o Inbox do Google, aquele cliente de e-mail há muito esquecido que perdeu a batalha contra o Gmail. Não estou aqui para dourar a pílula: o Google de fato tem a reputação de descontinuar bons produtos. Mas, preferências pessoais à parte, ao olhar para o quadro geral, eu realmente admiro o Google por ter a coragem de matar produtos da maneira que faz.

Acredito que a maioria das pessoas espera que o Google lidere a disrupção em tudo relacionado à tecnologia e, especialmente hoje, em um ambiente tão dinâmico por conta dos avanços em IA, é preciso muita coragem e determinação para pivotar de uma direção para outra. Eu falo muito sobre Agile, e embora o Google não seja tipicamente associado a processos ágeis formais, essa é uma característica que todos os agilistas experientes reconhecerão como uma das mais valiosas em uma empresa: a capacidade de corrigir o curso rapidamente, pivotar, experimentar, aprender com os erros e iterar.

Em vez de permanecer na zona de conforto, é isso que mantém o Google sempre na vanguarda: sua capacidade de se reinventar, mesmo que nem todo experimento dê certo. Na verdade, já é esperado que muitos experimentos fracassem. É assim que aprendemos o que funciona e o que não funciona. Absorvemos as lições e seguimos para o próximo objetivo, incorporando-as nos produtos mais novos.

Haverá muitas lições a serem aprendidas com este lançamento, mas, em última análise, se você olhar para a tecnologia em si, com sorte ficará claro qual é o objetivo final: estamos dobrando a aposta na era agêntica, consolidando esforços para construir produtos mais avançados.

## O novo aplicativo desktop Antigravity explicado

A maior mudança no aplicativo desktop é a remoção do componente de IDE. No Antigravity 1.x, o app era baseado em um fork do VS Code, oferecendo todos aqueles recursos familiares de IDE para navegar e editar código, combinados a uma caixa de assistente usada para interagir com o agente.

Não apenas isso, você também tinha uma interface secundária chamada "Agent Manager", onde era possível ter uma visão panorâmica de diferentes sessões de chat (as chamadas "conversas"), o que permitia trabalhar em vários projetos em paralelo monitorando os agentes nessa tela e reagindo quando eles estavam esperando por entrada.

A maior mudança no novo aplicativo desktop é que o Antigravity 2.0 coloca a experiência do Agent Manager em primeiro plano, removendo completamente a parte da IDE (que se tornou um aplicativo separado e opcional).

![A nova interface do Agent Manager](image.png "A nova interface do Agent Manager é mais limpa, focada em projetos e conversas")

Para desenvolvedores experientes, isso se tornou um grande ponto de atrito, pois, de repente, todas as ferramentas familiares de edição nas quais confiavam há tanto tempo simplesmente desapareceram. Você ainda consegue visualizar arquivos na interface do agy 2.0, mas apenas aqueles em que o agy está trabalhando no momento, e não é possível editá-los diretamente. Toda interação é feita por meio de um prompt ou de uma anotação no arquivo.

![Visualização de arquivos no agy 2.0](image-2.png "Você ainda pode visualizar arquivos na interface, mas não editá-los diretamente")

As interações com o agente já devem ser familiares para qualquer pessoa que tenha programado com agentes no último ano. Assim que você fornece um prompt, ele elabora um plano de implementação, que você pode revisar usando comentários inline ou um prompt de nível superior. Uma vez aprovado, o agente parte por conta própria para a execução. Dependendo de como você configura a interface, de tempos em tempos ele pode interromper pedindo uma permissão, a qual você pode conceder ou rejeitar com uma correção de rota opcional.

![Agent Manager solicitando entrada do usuário](image-1.png "Ao rejeitar uma solicitação, você pode adicionar um comentário de direcionamento")

Em termos de extensibilidade, o agy 2.0 suporta padrões comuns aos quais nos acostumamos neste último ano, incluindo MCP e Skills, mas também seu próprio mecanismo de "Rules" herdado da versão 1.x (essencialmente um AGENTS.md combinável) e um novo sistema de plugins baseado no sistema de extensões da antiga Gemini CLI. Os plugins permitem empacotar regras adicionais, comandos com barra (slash commands), servidores MCP, skills e subagentes, e são retrocompatíveis com extensões da Gemini CLI (ou seja, você pode instalar extensões da CLI no agy, mas não o contrário).

No geral, embora compreenda a frustração de quem sente falta da IDE, minha impressão inicial é que não sinto falta de tê-la no **mesmo** aplicativo. Mesmo quando trabalhava com a Gemini CLI, eu sempre mantinha o VS Code aberto em paralelo para quando queria fazer edições manuais, e este é o mesmo fluxo que estou aplicando ao agy 2.0. Na verdade, uso o VS Code quase exclusivamente como editor de texto e dificilmente uso algum recurso avançado de IDE hoje em dia. Poderia substituí-lo pelo Bloco de Notas e não faria muita diferença, exceto pela perda da memória muscular de alguns atalhos, que é o único motivo pelo qual continuo usando o VS Code atualmente.

Embora deva admitir que não há nada de revolucionário no agy 2.0 em comparação com o agy 1.x ou mesmo com outros agentes de código, estou gostando bastante do visual mais limpo, e acredito que só vou extrair todo o seu potencial quando começar a customizá-lo com meus próprios plugins. No momento, estou trabalhando na atualização do [GoDoctor]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}}) e do [Speedgrapher]({{< ref "/posts/20250805-introducing-speedgrapher" >}}), migrando-os do formato de extensões da Gemini CLI para plugins do agy, e trarei novidades assim que tiver algo para mostrar.

## Antigravity CLI

Para usuários de terminal, a experiência de linha de comando foi reconstruída sob a nova [**Antigravity CLI**](https://antigravity.google/blog/introducing-google-antigravity-cli) (conhecida como `agy CLI`). Pode ser um pouco confuso no início, mas você precisa instalar o aplicativo agy 2.0 mesmo que pretenda usar apenas a CLI, já que eles compartilham o mesmo processo de autenticação. A agy CLI é a substituta natural da Gemini CLI e, embora não ofereça 100% de paridade de recursos, os principais pilares já estão presentes: [hooks]({{< ref "/posts/20260610-mastering-hooks" >}}), [skills]({{< ref "/posts/20260829-the-pragmatic-guide-to-agent-skills" >}}), [MCP]({{< ref "/posts/20250817-hello-mcp-world" >}}), [subagentes]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}) e plugins.

Toda a CLI foi reescrita em Go (a Gemini CLI era em TypeScript), o que me deixa profundamente feliz, já que podemos esperar uma experiência muito mais ágil. Ao mesmo tempo, uma das principais críticas é que a agy CLI, até hoje, é de código fechado, o que pode parecer uma regressão em relação à Gemini CLI. Não faz muito tempo que fazíamos piadas sobre "vazar" o código da Gemini CLI para o público. Infelizmente, essa piada não envelheceu bem, já que agora nosso principal agente de código também tem código fechado.

Considerando que tenho zero controle sobre isso, é algo com o qual decidi não me preocupar. É cedo demais para dizer se essa foi uma decisão boa ou ruim, mas, novamente, reconheço a frustração da comunidade, especialmente de quem contribuiu para a Gemini CLI antes. Se serve de consolo, continuaremos tendo uma próspera comunidade open source em torno do sistema de plugins. Ao menos da minha parte, estou trabalhando nos meus para garantir que tenhamos tanto um subagente especialista em Go decente quanto um companheiro de vibe-writing muito em breve.

![Interface da agy CLI](image-3.png "A interface será familiar para quem vem da Gemini CLI ou do Claude Code")

Em termos de interface, ela não deve surpreender ninguém que já tenha usado um agente de código no terminal antes. Minha primeira impressão é que a renderização realmente parece melhor do que a renderização em TypeScript da Gemini CLI. Além disso, da mesma forma que no agy 2.0, aprecio muito o visual mais limpo. Na minha opinião pessoal, a Gemini CLI estava ficando grande demais para o seu próprio bem, com recursos em excesso e uma interface inchada, então essa interface mais limpa soa refrescante para mim. Um dos meus ditados favoritos de todos os tempos é "menos é mais", e a agy CLI cumpre o que promete nesse aspecto.

Onde ela não se sai muito bem (por enquanto) é principalmente no que diz respeito à compatibilidade com extensões. Embora exista um caminho de migração, ele nem sempre funciona como esperado, e é por isso que dediquei a maior parte da minha semana reescrevendo o godoctor e o speedgrapher, já que prefiro não depender da migração automática. Além disso, também tive problemas com a autenticação baseada em projeto, o que espero que seja corrigido em breve. Por enquanto, venho utilizando com minha assinatura Google Pro pessoal, já que a autenticação baseada em projeto não funcionou para mim.

Sem entrar nas complexidades de faturamento, que é outro ponto sensível para usuários vindos da Gemini CLI, minha opinião pessoal é de que a agy CLI tem seus problemas, mas também um grande potencial. Até o momento não há nada de revolucionário nela (pois a maior parte das mudanças está acontecendo sob o capô), mas também não vejo nenhum impeditivo grave. Tudo o que eu fazia na Gemini CLI pode ser feito com a agy CLI e há pouquíssimo a ser aprendido, de modo que, mesmo que a Gemini CLI tivesse uma janela de migração mais longa, eu recomendaria migrar o quanto antes para garantir que seu fluxo de trabalho esteja preparado para o futuro.

## Antigravity SDK

A discussão até agora foi sobre a substituição de produtos antigos por novos, o que soa mais incremental do que revolucionário. É por isso que o lançamento do [**Antigravity SDK**](https://antigravity.google/blog/introducing-google-antigravity-sdk) foi o anúncio mais empolgante para mim. Quando falei sobre a maioria das mudanças ocorrer sob o capô, tratava-se justamente da criação dessa plataforma unificada para dar suporte a agentes — e o Antigravity SDK é a forma como você, como pessoa desenvolvedora, também pode ter acesso a ela.

Aqui está um exemplo funcional de um agente consultando seu workspace em menos de 15 linhas de código:

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

Essa biblioteca [Python](https://xkcd.com/353/ "import antigravity") oferece a pessoas desenvolvedoras acesso programático exatamente ao mesmo runtime agêntico e harness de orquestração. O SDK é agnóstico em relação ao ambiente de execução e permite inicializar um loop de agente com estado em menos de 15 linhas de código. Ele oferece suporte a capacidades modulares, incluindo ferramentas integradas, funções personalizadas, servidores Model Context Protocol, subagentes e skills reutilizáveis sob um pipeline unificado.

## Primeiros passos

Uma tendência comum em todos os lançamentos relacionados ao Antigravity é a transição de code-first para design-first. Toda a experiência de desenvolvimento de software está sendo redesenhada em torno da coordenação de agentes em vez da edição de código. Para preparar seu ambiente de desenvolvimento para essa mudança, considere as seguintes ações:

1.  **Baixe o aplicativo desktop**: Acesse [antigravity.google](https://antigravity.google) para instalar a aplicação desktop.
2.  **Migre os fluxos de trabalho do terminal**: Instale a CLI `agy` e execute o comando de importação para migrar suas configurações da Gemini CLI antes da data de descontinuação em **18 de junho de 2026** (consulte o [anúncio de migração](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) para obter detalhes).
3.  **Explore o SDK**: Instale a biblioteca Python, confira os [recursos do Antigravity](https://antigravity.google/docs/features) e comece a criar agentes personalizados desenvolvidos com o SDK do agy:
    ```bash
    pip install google-antigravity
    ```

## Recursos adicionais
Para saber mais sobre este lançamento e ter acesso a detalhes técnicos adicionais, confira os seguintes recursos:
* **[Introducing Google Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0)**: O anúncio oficial do ecossistema 2.0.
* **[Introducing Google Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli)**: Uma análise aprofundada da nova interface de terminal desenvolvida em Go.
* **[An Important Update: Transitioning Gemini CLI to Antigravity CLI](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/)**: Cronograma detalhado de migração e diretrizes para usuários da Gemini CLI.
* **[Introducing Google Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk)**: Saiba como orquestrar agentes programaticamente em Python.
* **[Google I/O 2026 Developer Highlights](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/)**: Os principais anúncios para desenvolvedores do Google I/O deste ano.
* **[Google I/O 2026: Antigravity Announcement](https://antigravity.google/blog/google-io-2026)**: Principais atualizações e destaques do Antigravity no Google I/O.
* **[Google Antigravity Documentation & Features](https://antigravity.google/docs/features)**: O guia abrangente sobre recursos e controles de segurança do Antigravity.
