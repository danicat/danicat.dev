---
title: "Construa um Jogo Arcade Match 3 Usando o Gemini CLI"
date: 2026-05-07T09:00:00Z
categories: ["Workflow & Best Practices", "AI & Development"]
tags: ["gemini-cli", "golang", "codelab", "vibe-coding", "sub-agents"]
summary: "Aprenda como construir um jogo Match-3 totalmente funcional usando programação agêntica, o Gemini CLI e Go. Nós exploramos o modo plano e sub-agents personalizados."
heroStyle: "big"
---

O único motivo de eu ter me tornado um desenvolvedor de software foi porque eu amava videogames quando criança. Eu passava incontáveis horas jogando e ficava profundamente intrigado sobre como eles eram construídos. Meu pai tentava ao máximo explicar como TVs e computadores funcionavam, mas isso nunca entrava de verdade na minha cabeça.

Foi só durante a minha adolescência, quando finalmente tivemos acesso à internet, que eu comecei a entender um pouco mais. Enquanto os adolescentes normais estavam enchendo salas de bate-papo, correndo atrás das pessoas no ICQ e arrumando seus perfis no Orkut, eu estava pesquisando tutoriais de desenvolvimento de jogos. Aqueles foram bons tempos.

Os anos se passaram e eu nunca me tornei um desenvolvedor profissional de jogos. Minha carreira me levou para o lado de bancos de dados, engenharia de dados, serviços backend e nuvem. Eu não me arrependo das minhas escolhas. Ainda assim, de vez em quando eu me pergunto como seria a sensação de construir meu próprio jogo indie.

E adivinha? Com a ascensão do agentic coding, construir aplicações complexas — incluindo jogos — se tornou tão acessível que não precisamos mais apenas imaginar. Nós podemos construir um jogo totalmente funcional, implantado na nuvem hoje mesmo, como estou prestes a te mostrar.

Existem duas formas de você ler este artigo: como um aspirante a desenvolvedor de jogos querendo experimentar com GenAI, ou como um desenvolvedor profissional usando o desenvolvimento de jogos como uma forma divertida de aprender novas habilidades de agentic coding. Seja qual for o caminho que você escolher, durante este artigo eu vou te mostrar duas funcionalidades específicas do Gemini CLI: plan mode e sub-agents. Mas antes disso, vamos falar um pouco sobre tecnologia.

## Como escolher a tecnologia certa para o seu projeto

Essa sempre foi uma decisão importante em toda equipe de software. Devemos usar as ferramentas com as quais estamos familiarizados? Devemos seguir novas tendências de mercado? Devemos construir a nossa própria? Grandes empresas geralmente se apegam às ferramentas que já conhecem. Para justificar uma mudança, você precisa de um motivo muito forte. Esse motivo pode vir de fora — como mudanças nos custos do mercado ou a falta de talentos disponíveis. Ou pode vir de dentro, como o alto custo de retreinar a sua equipe para dar suporte a uma nova stack.

O agentic coding muda completamente essa dinâmica. Como a IA pode lidar com o boilerplate, a sua escolha de linguagem de programação importa muito menos hoje em dia do que a arquitetura geral do seu sistema. Para nós, desenvolvedores, isso é um alívio enorme. Nós podemos trocar as tech stacks para se adequarem ao problema sem precisar passar meses aprendendo uma sintaxe nova.

Você pode estar se perguntando: quando a linguagem perde a importância, o que fica? Minha resposta é: os padrões. A forma como estruturamos o software, não como um silo, mas como um coletivo de sistemas. Isso funciona tanto em um nível macro (design de sistemas) quanto micro (design de programas). Você não precisa saber o que cada linha de código faz, mas você **precisa** saber como as diferentes partes do seu software interagem entre si, e você **precisa** saber como direcionar o agente para o caminho da implementação **correta**.

Isso significa que podemos voltar a escrever tudo em BASIC? Não, porque uma linguagem nunca é uma escolha isolada. Uma linguagem traz com ela um conjunto específico de funcionalidades e todo um ecossistema. Nós estamos sempre atrelados a escolher a tecnologia que melhor se encaixa no que estamos tentando alcançar. A única coisa que não é mais tão relevante é a capacidade técnica da equipe de escrever o software em si. Isso pode ser facilmente mitigado com coding agents modernos, desde que a equipe tenha fundamentos sólidos de engenharia de software.

Enquanto um critério sai de moda, novos aparecem. Neste caso, nós vamos prestar muita atenção no quão fácil é para o coding agent gerar um software de alta qualidade na linguagem alvo.

Para este projeto em particular, eu escolhi Go por dois motivos principais: é uma linguagem leve com a qual os coding agents lidam muito bem (meu MCP godoctor também ajuda!) e possui um ecossistema de desenvolvimento de jogos open source maduro em torno da ebitengine.

Eu poderia ter feito em Three.js? Sim. No entanto, eu realmente queria chegar o mais perto possível da experiência de um jogo de console / arcade, então um jogo compilado é essencial para mim. Além disso, eu só me importo com 2D, então não há necessidade de engines grandes como Unity ou Unreal. Finalmente, a ebitengine tem jogos comerciais publicados na loja da Nintendo (para o Nintendo Switch), o que alimenta meu sonho de um dia publicar um jogo (claro, não este aqui).

Falando um pouco dos pontos fortes de Go: por ser uma linguagem compilada, isso nos ajuda a pegar grande parte dos erros no início do processo de desenvolvimento. Python tem capacidades semelhantes para desenvolvimento de jogos, mas ser interpretada significa que isso atrasa o meu ciclo de testes. Além disso, Go pode ser compilada nativamente para a sua máquina local, ou compilada para WebAssembly (WASM) para a web. Isso significa que eu também posso fazer o deploy do meu jogo como um serviço web com algumas pequenas alterações.

## O retorno do analista de software

Enquanto o agente está fazendo o trabalho pesado de escrever o código em Go e compilar tanto o servidor quanto os binários WASM, nós ainda temos responsabilidades rígidas quando o assunto é o design. 

A engenharia de software está mudando. Nós estamos gastando menos tempo nos preocupando com sintaxes espertas e mais tempo pensando em padrões de alto nível. 

De certa forma, parece que estamos voltando à era do clássico 'Analista de Software'. Em vez de escrever cada linha manualmente, nosso trabalho principal agora é traduzir requisitos humanos em um conjunto preciso de instruções para que a IA possa escrever o código de fato.

Eu não tenho experiência em desenvolvimento de jogos em si, mas como um gamer e entusiasta, estou familiarizado com a **linguagem de domínio** usada para descrever o que eu quero alcançar com o meu jogo. Ao embasar o meu prompt em certas palavras-chave (por exemplo, arcade game, match 3) ou usando exemplos conhecidos (por exemplo, "Eu preciso de uma trilha inspirada nas gerações 16-bit e 32-bit de jogos de puzzle, mas com um toque moderno"), eu consigo comunicar minhas intenções ao agente de uma forma muito melhor do que alguém tentando construir um jogo com absolutamente nenhuma experiência com jogos.

Só estou deixando isso aqui para destacar um ponto. Mesmo que a programação em si se torne uma habilidade secundária, a capacidade de descrever padrões e funcionalidades continua sendo uma habilidade crítica da engenharia de software. Você precisa conhecer a linguagem de domínio da sua área, seja ela backend, frontend ou qualquer coisa no meio do caminho.

## Movendo do design para a implementação com o plan mode

A linguagem de domínio é um começo, mas escrever o prompt one-shot perfeito raramente é viável. Trabalhando com developer relations, nós usamos prompts one-shot o tempo todo em demos e apresentações, mas o que geralmente não contamos é sobre como frequentemente gastamos horas refinando esse prompt one-shot antes de podermos mostrá-lo ao público.

Criar o prompt perfeito é uma mistura de arte e ciência e, mesmo que você tenha um entendimento profundo da linguagem de domínio, sempre haverá lacunas. Felizmente, fora do mundo das demos e apresentações, não precisamos resolver tudo com um só tiro. Além disso, não precisamos trabalhar nos prompts sozinhos, já que os agentes também podem nos ajudar com eles. É aqui que o **plan mode** entra para ajudar.

No plan mode, o Gemini CLI vai primeiro elaborar um plano de implementação antes de escrever qualquer código. Isso cria uma oportunidade para você ter uma conversa de ida e volta com o agente, refinando o plano e garantindo que a implementação está indo na direção que você deseja.

Numa conversa normal com o agente, ele pode pedir para entrar no plan mode com base no fluxo da conversa (por exemplo, respondendo a um prompt que inclui a frase "vamos fazer um plano"), mas se você não quiser depender do agente para decidir quando entrar no plan mode, você pode sempre ativá-lo manualmente com o comando `/plan`.

No plan mode, o agente não só vai elaborar um plano de implementação com base no seu pedido, mas também pode fazer uma ou mais perguntas de esclarecimento usando a ferramenta `ask_user`. Quando o plano estiver pronto, ele pedirá sua revisão e te dará a oportunidade de guiar o plano em qualquer direção, incluindo corrigir suposições e adicionar ou remover funcionalidades.

Por exemplo, um prompt razoavelmente polido - mas longe de ser perfeito - para o meu jogo Match 3 é mostrado abaixo:

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

Apesar deste prompt cobrir muitos aspectos do jogo, é comum que o agente peça por mais detalhes, como "qual deve ser a resolução da tela" ou "você preferiria animações suaves ou estáticas".

Quando estivermos satisfeitos com o nível de detalhes no plano, podemos pedir para o agente começar a construir, o que fará com que ele saia do plan mode. Essa parte não é diferente de nenhuma tarefa típica de programação. Depois de alguns turnos, deveremos ter um jogo rodando parecido com este:

![Cloud Crush Gameplay Screenshot](cloud-crush-gameplay.png)

## Automatizando testes web com o browser agent

Uma das coisas mais difíceis de se fazer no desenvolvimento de jogos é testar. Você não pode escrever um teste de unidade padrão para verificar todos os estados possíveis do jogo ou para checar se suas funções de renderização estão desenhando os elementos certos na tela. Você poderia até tentar, mas eu garanto que seria um processo tedioso, frágil e muito demorado.

Isso não significa que não devemos escrever nenhum teste automatizado, mas sim que existem limites entre o que deve ser feito com puro código e o que requer testes sendo jogados por humanos. Por exemplo, testar unitariamente algoritmos como colisão e path finding me parece ok, mas validar a sua UI em diferentes resoluções pode ser melhor feito por um humano (por exemplo, como você testa unitariamente "essa fonte está legível?").

Ou, pelo menos, era assim até agora... _um sub-agent entra no chat_

Com as capacidades multimodais dos modelos de fronteira e o uso inteligente de agentes, nós podemos, de fato, automatizar a verificação visual. No Gemini CLI, um sub-agent é uma persona especializada que roda independentemente da conversa principal, em sua própria janela de contexto. Sub-agents podem ser usados para adicionar todos os tipos de capacidades ao seu fluxo de trabalho básico de programação.

No nosso cenário de testes, podemos usar um agente experimental que já vem embutido no CLI chamado de `@browser_agent`. Por ser experimental, você precisa [habilitá-lo manualmente](https://geminicli.com/docs/core/subagents/#enabling-the-browser-agent) editando o seu arquivo `settings.json`. Por exemplo, este é um `settings.json` minimalista que habilita o browser agent com um modelo visual:

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

Normalmente, o browser agent navega em uma página web lendo a sua árvore de acessibilidade — a estrutura oculta que leitores de tela usam. No entanto, o nosso jogo Match 3 é desenhado inteiramente sobre um único canvas HTML. Para a árvore de acessibilidade, ele parece apenas uma caixa gigante em branco. 

É aqui que adicionar um modelo de visão muda o jogo. Ao configurar o agente com um `visualModel` (como o `gemini-2.5-computer-use-preview-10-2025`), ele literalmente aprende a ver. Ele tira screenshots, analisa o layout visual e descobre as coordenadas exatas X e Y que precisa clicar na tela.

Em vez de clicar manualmente pela aplicação implantada no Cloud Run, você pode digitar `@browser_agent please test the live URL...` para instruí-lo a navegar no site, jogar uma rodada do jogo e tirar screenshots das telas funcionando. 

Isso não substitui os testes feitos por humanos (playtesting) para avaliar a "sensação" do jogo (game feel), mas automatiza a validação visual, provando que a UI renderiza corretamente sem precisarmos sair do terminal.

## Terceirizando minha ansiedade com segurança

Com a implementação funcionando e a UI verificada, não podemos nos esquecer da segurança. 

Eu não sou um especialista em segurança de aplicações, o que me torna a pessoa errada para avaliar a postura de segurança de um web app. No entanto, assim como o agentic coding mitigou minha falta de experiência com game engines, sub-agents podem mitigar minha falta de expertise em segurança. Como um orquestrador, eu não preciso conhecer cada vetor de cross-site scripting; eu só preciso saber como iniciar um especialista com um contexto limpo para procurá-los.

Nós podemos criar um ambiente de execução isolado definindo um [agente personalizado](https://geminicli.com/docs/core/subagents/#creating-custom-subagents) em um arquivo Markdown (`.gemini/agents/security-auditor.md`) que pode ser invocado usando `@security_auditor`. 

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

Nós damos a ele um system prompt específico (o corpo do arquivo markdown) e ferramentas como `read_file` e `grep_search` (definidas no frontmatter). Pelo fato de rodar em seu próprio loop de contexto, ele não polui o histórico principal da nossa conversa.

Eu apontei este auditor para a base de código do *Cloud Crush* para checar por credenciais chumbadas no código, operações inseguras de arquivo e riscos de deployment. Mesmo que um agente de segurança personalizado não substitua um profissional humano dedicado, ele fornece uma camada base de defesa da qual, de outra forma, eu sentiria falta.

## Um novo fluxo de desenvolvimento

Este fluxo de trabalho define o que eu considero ser o novo padrão para desenvolvimento de software. Nós estamos usando agentes para escrever o código e ativamente construindo ferramentas personalizadas, skills e sub-agents para reforçar nossos padrões arquiteturais e de qualidade.

E para os leitores mais atentos, vocês podem ter notado que eu fui intencionalmente sucinto nas instruções passo a passo neste artigo. Isso é porque temos um codelab inteiro dedicado a essa experiência, que você pode acessar usando o link abaixo. Neste codelab, você vai poder testar tudo o que discutimos neste artigo seguindo instruções passo a passo, construindo por fim a sua própria versão deste jogo Match 3.

**Codelab: [Construa um Jogo Arcade Match 3 Com o Gemini CLI](https://codelabs.developers.google.com/next26/gemini-cli-match3-golang#0)**

Claro que, se você tiver qualquer dúvida, fique à vontade para me mandar uma mensagem em qualquer uma das minhas redes sociais.
