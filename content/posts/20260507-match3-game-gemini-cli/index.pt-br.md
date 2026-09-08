---
categories:
- Agentic Coding
date: 2026-05-07 09:00:00+00:00
heroStyle: big
summary: Aprenda como construir um jogo Match-3 totalmente funcional usando agentic coding, o Gemini CLI e Go. Exploramos o plan mode e subagentes customizados.
tags:
  - codelab
  - gemini-cli
  - golang
  - subagents
  - vibe-coding
title: "Construa um Jogo Arcade Match 3 Usando o Gemini CLI"
slug: "match3-game-gemini-cli"
aliases:
  - "/pt-br/posts/20260507-match3-game-gemini-cli/"
description: "Tutorial prático de desenvolvimento de um jogo Match-3 2D em Go e Ebitengine com Gemini CLI. Aborda plan mode, subagentes para testes visuais e deploy no Cloud Run."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.22+"
  - "Ebitengine v2"
  - "Gemini CLI / Antigravity CLI"
  - "Google Cloud Run"
---

{{< alert "circle-info" >}}
**Nota:** Este artigo foi escrito para o Gemini CLI, que foi descontinuado e substituído pelo **Google Antigravity 2.0**. Para saber mais sobre a nova Antigravity CLI (`agy`), o SDK e o ecossistema mais amplo do Antigravity, confira [O Guia do Mochileiro para o Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}).
{{< /alert >}}

{{< alert "circle-info" >}}
**Codelab Atualizado:** Acompanhe a versão moderna deste tutorial atualizada para o Antigravity em [goo.gle/cloud-crush-agy](https://goo.gle/cloud-crush-agy).
{{< /alert >}}

O verdadeiro motivo de eu ter me tornado uma desenvolvedora de software foi o meu amor por videogames quando criança. Eu passava horas e horas jogando e ficava fascinada tentando entender como eles eram feitos. Meu pai fazia o melhor que podia para me explicar como televisores e computadores funcionavam, mas aquilo nunca entrava direito na minha cabeça.

Foi só na adolescência, quando finalmente tivemos acesso à internet, que comecei a entender um pouco melhor. Enquanto outros adolescentes passavam o tempo em salas de bate-papo, mandando mensagens no ICQ e personalizando seus perfis no Orkut, eu passava horas procurando tutoriais de desenvolvimento de jogos. Aqueles sim eram bons tempos.

Os anos passaram e eu nunca me tornei uma desenvolvedora profissional de jogos. Minha carreira me levou para a área de bancos de dados, engenharia de dados, serviços de backend e cloud. Não me arrependo das minhas escolhas. Mesmo assim, de vez em quando ainda me pego pensando em como seria a sensação de criar meu próprio jogo indie.

E adivinha? Com a ascensão do agentic coding, construir aplicações complexas — incluindo jogos — se tornou tão acessível que não precisamos mais ficar só na imaginação. Podemos criar um jogo totalmente funcional e com deploy na nuvem hoje mesmo, como vou te mostrar agora.

Existem duas formas de ler este artigo: como alguém querendo criar jogos que deseja experimentar com GenAI, ou como uma pessoa desenvolvedora experiente que vê no desenvolvimento de jogos uma forma divertida de aprender novas habilidades de agentic coding. Seja qual for o seu caminho, ao longo deste artigo vou te mostrar duas funcionalidades específicas do Gemini CLI: o plan mode e subagentes. Mas antes disso, vamos falar um pouco sobre tecnologia.

## Como escolher a tecnologia certa para o seu projeto

Essa sempre foi uma decisão crítica em qualquer equipe de software. Devemos usar as ferramentas com as quais já temos familiaridade? Devemos seguir novas tendências de mercado? Devemos criar algo próprio? Grandes empresas normalmente se apegam às ferramentas que já conhecem. Para justificar uma mudança, é preciso um motivo muito forte. Esse motivo pode vir de fora — como mudanças nos custos do mercado ou escassez de talentos qualificados. Ou pode vir de dentro, como o alto custo de retreinar o time para dar suporte a uma nova stack.

O agentic coding muda completamente essa dinâmica. Como a IA pode cuidar do boilerplate, a escolha da linguagem de programação hoje importa muito menos do que a arquitetura geral do sistema. Para nós, pessoas desenvolvedoras, isso é um alívio enorme. Podemos trocar de stack técnica para resolver o problema sem passar meses aprendendo uma sintaxe nova.

Você deve estar se perguntando: quando a linguagem perde o protagonismo, o que sobra? Minha resposta é: os padrões. A forma como estruturamos o software, não como um silo isolado, mas como um conjunto de sistemas interconectados. Isso se aplica tanto no nível macro (system design) quanto no micro (design de programas). Você não precisa saber o que cada linha de código faz, mas você **precisa** entender como as diferentes peças do seu software conversam entre si, e você **precisa** saber como direcionar o agente rumo à implementação **correta**.

Isso significa que podemos voltar a escrever tudo em BASIC? Não, porque uma linguagem nunca é uma escolha isolada. Ela carrega consigo um conjunto específico de recursos e todo um ecossistema. Continuamos comprometidos a escolher a tecnologia que melhor atende ao que estamos tentando alcançar. A única coisa que deixou de ser um obstáculo é a capacidade imediata do time de escrever aquele código específico. Isso pode ser facilmente mitigado com coding agents modernos, desde que a equipe tenha fundamentos sólidos de engenharia de software.

Enquanto um critério perde relevância, novos surgem. Neste caso, vamos prestar muita atenção no quão fácil é para o coding agent gerar software de alta qualidade na linguagem escolhida.

Para este projeto específico, escolhi Go por dois motivos principais: é uma linguagem enxuta com a qual os coding agents lidam muito bem (meu MCP godoctor também ajuda!) e conta com um ecossistema open source maduro para desenvolvimento de jogos em torno da ebitengine.

Eu poderia ter feito em Three.js? Sim. No entanto, eu queria muito chegar o mais perto possível da experiência de um jogo de arcade / console, então um jogo compilado é essencial para mim. Além disso, o foco aqui é puramente 2D, dispensando engines pesadas como Unity ou Unreal. Por fim, a ebitengine já tem jogos comerciais publicados na Nintendo eShop (para Nintendo Switch), o que alimenta meu sonho de um dia publicar um jogo próprio (obviamente não este aqui).

Falando um pouco das vantagens de Go: ser uma linguagem compilada nos ajuda a pegar boa parte dos erros logo cedo no processo de desenvolvimento. Python oferece capacidades parecidas para desenvolvimento de jogos, mas o fato de ser interpretada deixa o meu ciclo de testes mais lento. Além disso, Go pode ser compilada nativamente para a sua máquina local ou compilada para WebAssembly (WASM) para a web. Isso significa que também posso fazer o deploy do jogo como um serviço web com pouquíssimas alterações.

## O retorno do analista de software

Enquanto o agente faz o trabalho pesado de escrever o código em Go e compilar tanto o servidor quanto os binários em WASM, nós continuamos com responsabilidades rigorosas no que diz respeito ao design.

A engenharia de software está mudando. Passamos menos tempo nos preocupando com truques de sintaxe e mais tempo pensando em padrões de alto nível.

De certa forma, parece que estamos voltando à era do clássico 'Analista de Software'. Em vez de digitar cada linha manualmente, nosso trabalho principal passa a ser traduzir requisitos humanos em um conjunto preciso de instruções para que a IA escreva o código de verdade.

Eu não tenho experiência profissional em desenvolvimento de jogos propriamente dita, mas como gamer e entusiasta, tenho familiaridade com a **linguagem de domínio** necessária para descrever o que quero alcançar no meu jogo. Ao ancorar meu prompt em certas palavras-chave (por exemplo: arcade game, match 3) ou usar referências consagradas (por exemplo: "Preciso de uma trilha sonora inspirada nas gerações 16-bit e 32-bit de jogos de puzzle, mas com uma roupagem moderna"), consigo comunicar minhas intenções ao agente de forma muito mais eficaz do que alguém tentando criar um jogo sem nenhuma vivência com games.

Estou destacando isso para reforçar um ponto: mesmo que codificar acabe se tornando uma habilidade secundária, a capacidade de descrever padrões e funcionalidades continua sendo uma competência crítica da engenharia de software. Você precisa dominar a linguagem de domínio da sua área, seja backend, frontend ou qualquer ponto intermediário.

## Do design à implementação com o plan mode

A linguagem de domínio é um bom começo, mas escrever o prompt one-shot perfeito de primeira é praticamente impossível. Em Developer Relations, nós usamos prompts one-shot o tempo todo em palestras e demonstrações, mas o que raramente contamos é quantas horas passamos lapidando aquele prompt antes de subir no palco para mostrá-lo ao público.

Construir o prompt ideal é um misto de arte e ciência, e mesmo dominando a linguagem de domínio sempre existirão lacunas. Felizmente, fora do circuito de apresentações e demos, não precisamos acertar tudo de primeira. Além do mais, não precisamos quebrar a cabeça com os prompts sozinhos, já que os próprios agentes podem nos ajudar nisso. É aí que o **plan mode** entra em ação.

No plan mode, o Gemini CLI elabora primeiro um plano de implementação antes de tocar em qualquer linha de código. Isso abre espaço para uma conversa iterativa com o agente, refinando o plano e garantindo que a implementação siga exatamente o rumo que você deseja.

Em uma conversa comum com o agente, ele pode sugerir entrar no plan mode a partir do fluxo do diálogo (por exemplo, ao responder a um prompt que contenha algo como "vamos fazer um plano"). Mas se você não quiser depender da decisão do agente sobre quando entrar no plan mode, é possível ativá-lo manualmente a qualquer momento com o comando `/plan`.

No plan mode, o agente não apenas elabora um plano de implementação com base no seu pedido, mas também pode fazer perguntas de esclarecimento usando a ferramenta `ask_user`. Quando o plano fica pronto, ele pede a sua revisão e te dá a oportunidade de direcionar o plano para onde quiser — corrigindo premissas, ajustando detalhes e adicionando ou removendo funcionalidades.

Por exemplo, um prompt razoavelmente refinado — mas longe de perfeito — para o meu jogo Match 3 é apresentado abaixo:

```txt
Build a Match-3 game called 'Cloud Crush' in Go using Ebitengine v2.
The entire game screen should have background.png as background.
The play area should be an 8x8 grid with white background. 
On the right side of the play area include a side panel with UI elements 
like player score and how to play instructions.
The side panel should have a solid background colour to help with readability of the UI.

Use standard GCP product logos (e.g. Compute Engine, Cloud Storage, BigQuery, etc.)
as the game gems. These logos are provided in the gcp_sprites.png file.

The logos are saved as 64x64 sprites but scale them as necessary
based on the screen resolution. Implement swapping, clearing 3+ gems, and gravity.

Use ebitengine native font rendering (size 48 for titles and size
24 for normal text) for all text and not the debug print.

The font should be monospaced (golang.org/x/image/font/gofont/gomono).
Keep the UI tidy and harmonic, e.g. centered text should always be
adjusted based on text length, not just guess based on estimates.
```

Embora esse prompt cubra muitos aspectos do jogo, é perfeitamente normal que o agente peça detalhes adicionais, como "qual deve ser a resolução da tela?" ou "você prefere animações fluidas ou transições estáticas?".

Quando estivermos satisfeitos com o nível de detalhe do plano, podemos pedir ao agente para começar a codificar, o que encerra o plan mode. A partir daí, o processo não difere de qualquer tarefa comum de desenvolvimento. Após algumas rodadas de interação, temos um jogo funcionando parecido com este:

![Cloud Crush Gameplay Screenshot](cloud-crush-gameplay.png)

## Automatizando testes web com o browser agent

Uma das tarefas mais difíceis no desenvolvimento de jogos é testar. Não tem como escrever um teste unitário padrão para cobrir todos os estados visuais possíveis de um jogo, ou conferir se as suas funções de renderização estão desenhando os elementos certos na tela. Até daria para tentar, mas garanto que seria um processo exaustivo, frágil e que consumiria um tempo absurdo.

Isso não significa que não devamos escrever testes automatizados, mas sim que existem fronteiras entre o que faz sentido cobrir com código puro e o que precisa de playtesting humano. Por exemplo: criar testes unitários para algoritmos de colisão ou pathfinding faz todo o sentido, mas validar a UI em diferentes resoluções é uma tarefa que uma pessoa faz muito melhor (afinal, como você cria um teste unitário para "essa fonte está legível?").

Ou, pelo menos, era assim até agora... _um subagente entra na conversa_

Com as capacidades multimodais dos frontier models e o uso inteligente de agentes, hoje já conseguimos automatizar a validação visual. No Gemini CLI, um subagente é uma persona especializada que roda de maneira independente da conversa principal, dentro da sua própria janela de contexto. Subagentes podem ser usados para adicionar todo tipo de superpoder ao seu fluxo base de programação.

No nosso cenário de testes, podemos usar um agente experimental integrado ao CLI chamado `@browser_agent`. Como se trata de um recurso experimental, você precisa [habilitá-lo manualmente](https://geminicli.com/docs/core/subagents/#enabling-the-browser-agent) editando o arquivo `settings.json`. Veja abaixo um exemplo minimalista de `settings.json` habilitando o browser agent com um modelo visual:

```json
{
  "agents": {
    "overrides": {
      "browser_agent": {
        "enabled": true
      }
    },
    "browser": {
      "visualModel": "gemini-2.5-computer-use-preview-10-2025"
    }
  }
}
```

Normalmente, o browser agent navega por uma página web inspecionando a árvore de acessibilidade — a estrutura interna utilizada por leitores de tela. No entanto, o nosso jogo Match 3 é renderizado inteiramente dentro de um único elemento canvas do HTML. Para a árvore de acessibilidade, aquilo não passa de uma grande caixa em branco.

É exatamente aqui que adicionar um modelo de visão muda o jogo. Ao configurar o agente com um `visualModel` (como o `gemini-2.5-computer-use-preview-10-2025`), ele literalmente aprende a enxergar. Ele tira capturas de tela, analisa o layout visual e descobre as coordenadas X e Y exatas onde precisa clicar na tela.

Em vez de ficar clicando manualmente na aplicação rodando no Cloud Run, você pode simplesmente digitar `@browser_agent please test the live URL...` para instruí-lo a navegar pelo site, jogar uma rodada e tirar screenshots das telas em funcionamento.

Isso não substitui o playtesting humano para avaliar o "game feel" (a sensação do jogo), mas automatiza a validação visual, comprovando que a UI renderiza perfeitamente sem que você precise sair do terminal.

## Terceirizando minha ansiedade com segurança

Com a implementação funcionando e a UI validada, não podemos nos esquecer da segurança.

Eu não sou especialista em segurança de aplicações, o que faz de mim a pior pessoa para avaliar a postura de segurança de um web app. Porém, assim como o agentic coding compensou minha falta de experiência com game engines, subagentes podem compensar minha falta de expertise em segurança. Como orquestradora, não preciso saber de cabeça cada vetor de cross-site scripting; só preciso saber como instanciar um especialista com um contexto limpo para caçá-los.

Podemos criar um ambiente de execução isolado definindo um [agente customizado](https://geminicli.com/docs/core/subagents/#creating-custom-subagents) em um arquivo Markdown (`.gemini/agents/security-auditor.md`) que pode ser acionado via `@security_auditor`:

```markdown
---
name: security_auditor
description: Specialized in finding security vulnerabilities in code.
kind: local
tools:
  - read_file
  - grep_search
model: gemini-3-flash-preview
temperature: 0.2
max_turns: 10
---

You are a ruthless Security Auditor. Your job is to analyze code for potential
vulnerabilities.

Focus on:

1.  SQL Injection
2.  XSS (Cross-Site Scripting)
3.  Hardcoded credentials
4.  Unsafe file operations

When you find a vulnerability, explain it clearly and suggest a fix. Do not fix
it yourself; just report it.
```

Nós passamos um system prompt específico para ele (o corpo do arquivo markdown) e ferramentas como `read_file` e `grep_search` (definidas no frontmatter). Como ele roda no seu próprio loop de contexto, não polui o histórico da conversa principal.

Apontei esse auditor para a base de código do *Cloud Crush* para verificar a existência de credenciais hardcoded, operações de arquivo inseguras e riscos de deploy. Mesmo que um subagente customizado de segurança não substitua um profissional dedicado de segurança, ele oferece uma camada essencial de proteção que, de outra forma, eu não teria.

## Um novo fluxo de desenvolvimento

Esse fluxo define o que considero ser o novo padrão para o desenvolvimento de software. Estamos usando agentes para escrever o código e construindo ativamente ferramentas, skills e subagentes customizados para fazer valer nossos padrões arquiteturais e de qualidade.

E para quem lê com atenção: você deve ter notado que fui intencionalmente econômica nas instruções passo a passo ao longo deste artigo. Isso porque temos um codelab completo dedicado a essa experiência prática, que você pode acessar no link abaixo. Nele, você poderá testar tudo o que discutimos aqui seguindo um guia detalhado, construindo no final a sua própria versão deste jogo Match-3.

**Codelab: [Construa um Jogo Arcade Match 3 Com o Gemini CLI](https://codelabs.developers.google.com/next26/gemini-cli-match3-golang#0)**

E claro, se tiver qualquer dúvida ou quiser trocar uma ideia, fique à vontade para me procurar em qualquer uma das minhas redes sociais.
