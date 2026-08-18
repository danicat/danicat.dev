---categories:
- Agentic Coding
date: 2026-05-07 09:00:00+00:00
heroStyle: big
summary: Aprenda como construir um jogo Match-3 totalmente funcional usando programação
  agêntica, o Gemini CLI e Go. Nós exploramos o modo de planejamento e subagentes personalizados.
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
description: "Tutorial prático de desenvolvimento de um jogo Match-3 2D em Go e Ebitengine com Gemini CLI. Aborda modo de planejamento, subagentes visuais e deploy no Cloud Run."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.22+"
  - "Ebitengine v2"
  - "Gemini CLI / Antigravity CLI"
  - "Google Cloud Run"
---

{{< alert "circle-info" >}}
**Nota:** Este artigo foi escrito para o Gemini CLI, que foi descontinuado e substituído pelo **Google Antigravity 2.0**. Para saber mais sobre a nova Antigravity CLI (`agy`), o SDK e o ecossistema Antigravity atualizado, confira [O Guia do Mochileiro para o Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}).
{{< /alert >}}

{{< alert "circle-info" >}}
**Codelab Atualizado:** Acompanhe a versão moderna deste tutorial atualizada para o Antigravity em [goo.gle/cloud-crush-agy](https://goo.gle/cloud-crush-agy).
{{< /alert >}}

O único motivo de eu ter me tornado uma desenvolvedora de software foi porque eu amava videogames quando criança. Eu passava incontáveis horas jogando e ficava profundamente intrigada sobre como eles eram construídos. Meu pai tentava ao máximo explicar como TVs e computadores funcionavam, mas isso nunca entrava de verdade na minha cabeça.

Foi só durante a minha adolescência, quando finalmente tivemos acesso à internet, que eu comecei a entender um pouco mais. Enquanto as pessoas da minha idade estavam enchendo salas de bate-papo, conversando no ICQ e arrumando seus perfis no Orkut, eu pesquisava tutoriais de desenvolvimento de jogos. Aqueles foram bons tempos.

Os anos se passaram e eu nunca me tornei uma desenvolvedora profissional de jogos. Minha carreira me levou para o lado de bancos de dados, engenharia de dados, serviços de backend e nuvem. Eu não me arrependo das minhas escolhas. Ainda assim, de vez em quando eu me pergunto como seria a sensação de construir meu próprio jogo indie.

E adivinha? Com a ascensão do agentic coding, construir aplicações complexas — incluindo jogos — se tornou tão acessível que não precisamos mais apenas imaginar. Nós podemos construir um jogo totalmente funcional, implantado na nuvem hoje mesmo, como estou prestes a te mostrar.

Existem duas formas de você ler este artigo: como alguém aspirando a desenvolver jogos que queira experimentar com GenAI, ou como profissional de desenvolvimento usando o desenvolvimento de jogos como uma forma divertida de aprender novas habilidades de programação agêntica. Seja qual for o caminho que você escolher, ao longo deste artigo eu vou te mostrar duas funcionalidades específicas do Gemini CLI: o plan mode e subagentes. Mas antes disso, vamos falar um pouco sobre tecnologia.

## Como escolher a tecnologia certa para o seu projeto

Essa sempre foi uma decisão importante em qualquer equipe de software. Devemos usar as ferramentas com as quais estamos familiarizados? Devemos seguir novas tendências de mercado? Devemos construir a nossa própria? Grandes empresas geralmente se apegam às ferramentas que já conhecem. Para justificar uma mudança, você precisa de um motivo muito forte. Esse motivo pode vir de fora — como mudanças nos custos do mercado ou escassez de talentos disponíveis. Ou pode vir de dentro, como o alto custo de retreinar a sua equipe para dar suporte a uma nova stack.

O agentic coding muda completamente essa dinâmica. Como a IA pode lidar com o boilerplate, a sua escolha de linguagem de programação importa muito menos hoje em dia do que a arquitetura geral do seu sistema. Para nós que desenvolvemos software, isso é um alívio enorme. Podemos trocar de stack técnica para se adequar ao problema sem passar meses aprendendo uma sintaxe nova.

Você pode estar se perguntando: quando a linguagem perde a importância, o que sobra? Minha resposta é: os padrões. A forma como estruturamos o software, não como um silo, mas como um coletivo de sistemas. Isso funciona tanto em um nível macro (design de sistemas) quanto micro (design de programas). Você não precisa saber o que cada linha de código faz, mas você **precisa** saber como as diferentes partes do seu software interagem entre si, e você **precisa** saber como direcionar o agente para o caminho da implementação **correta**.

Isso significa que podemos voltar a escrever tudo em BASIC? Não, porque uma linguagem nunca é uma escolha isolada. Uma linguagem traz consigo um conjunto específico de funcionalidades e todo um ecossistema. Estamos sempre atrelados a escolher a tecnologia que melhor se encaixa no que estamos tentando alcançar. A única coisa que não é mais uma barreira intransponível é a capacidade imediata da equipe de escrever o código em si. Isso pode ser facilmente mitigado com coding agents modernos, desde que a equipe tenha fundamentos sólidos de engenharia de software.

Enquanto um critério sai de moda, novos aparecem. Neste caso, vamos prestar muita atenção no quão fácil é para o coding agent gerar software de alta qualidade na linguagem-alvo.

Para este projeto em particular, escolhi Go por dois motivos principais: é uma linguagem enxuta com a qual os coding agents lidam muito bem (meu MCP godoctor também ajuda!) e possui um ecossistema maduro de desenvolvimento de jogos open source em torno da ebitengine.

Eu poderia ter feito em Three.js? Sim. No entanto, eu realmente queria chegar o mais perto possível da experiência de um jogo de console / arcade, então um jogo compilado é essencial para mim. Além disso, eu só me importo com 2D, então não há necessidade de engines pesadas como Unity ou Unreal. Por fim, a ebitengine tem jogos comerciais publicados na loja da Nintendo (para o Nintendo Switch), o que alimenta meu sonho de um dia publicar um jogo (claro, não este aqui).

Falando um pouco dos pontos fortes de Go: por ser uma linguagem compilada, ela nos ajuda a capturar grande parte dos erros logo no início do processo de desenvolvimento. Python tem capacidades semelhantes para desenvolvimento de jogos, mas ser interpretada significa que isso atrasa meu ciclo de testes. Além disso, Go pode ser compilada nativamente para a sua máquina local ou compilada para WebAssembly (WASM) para a web. Isso significa que eu também posso fazer o deploy do meu jogo como um serviço web com pouquíssimas alterações.

## O retorno do analista de software

Enquanto o agente faz o trabalho pesado de escrever o código em Go e compilar tanto o servidor quanto os binários WASM, nós ainda temos responsabilidades rigorosas quando o assunto é o design. 

A engenharia de software está mudando. Estamos gastando menos tempo nos preocupando com sintaxes mirabolantes e mais tempo pensando em padrões de alto nível. 

De certa forma, parece que estamos voltando à era do clássico 'Analista de Software'. Em vez de escrever cada linha manualmente, nosso trabalho principal agora é traduzir requisitos humanos em um conjunto preciso de instruções para que a IA possa escrever o código de fato.

Eu não tenho experiência profissional em desenvolvimento de jogos em si, mas como gamer e entusiasta, estou familiarizada com a **linguagem de domínio** usada para descrever o que quero alcançar com o meu jogo. Ao embasar meu prompt em certas palavras-chave (por exemplo, jogo arcade, match 3) ou usar referências consagradas (por exemplo, "Preciso de uma trilha sonora inspirada nas gerações 16-bit e 32-bit de jogos de puzzle, mas com uma roupagem moderna"), consigo comunicar minhas intenções ao agente de forma muito mais eficaz do que alguém tentando construir um jogo sem nenhuma bagagem com games.

Estou deixando isso registrado para enfatizar um ponto: mesmo que a codificação em si se torne uma habilidade secundária, a capacidade de descrever padrões e requisitos continua sendo uma competência essencial da engenharia de software. Você precisa dominar a linguagem de domínio da sua área, seja ela backend, frontend ou qualquer área correlata.

## Indo do design à implementação com o plan mode

A linguagem de domínio é um começo, mas escrever o prompt one-shot perfeito raramente é viável. No time de Developer Relations, usamos prompts one-shot o tempo todo em demonstrações e apresentações, mas o que geralmente não contamos é quantas horas passamos refinando esse prompt antes de mostrá-lo ao público.

Elaborar o prompt perfeito é uma mistura de arte e ciência. Mesmo que você domine a linguagem de domínio, sempre existirão lacunas. Felizmente, fora do palco de demos e palestras, não precisamos acertar tudo de primeira. Além disso, não precisamos trabalhar nos prompts sozinhos, já que os próprios agentes podem nos ajudar. É aqui que o **plan mode** se destaca.

No plan mode, o Gemini CLI elabora primeiro um plano de implementação antes de escrever qualquer linha de código. Isso cria uma oportunidade para você iterar com o agente, refinando o plano e garantindo que a implementação siga na direção desejada.

Em uma conversa normal com o agente, ele pode sugerir entrar no plan mode dependendo do contexto (por exemplo, ao responder a um prompt que contenha "vamos fazer um plano"). Mas se você não quiser depender dessa decisão automática, pode ativá-lo manualmente a qualquer momento com o comando `/plan`.

No plan mode, o agente não apenas estrutura o plano de implementação com base no seu pedido, como também pode fazer perguntas de alinhamento usando a ferramenta `ask_user`. Quando o plano estiver pronto, ele solicita sua revisão, permitindo ajustar o escopo, corrigir premissas e adicionar ou remover funcionalidades.

Como exemplo, um prompt razoavelmente refinado — mas longe de perfeito — para o meu jogo Match 3 pode ser visto abaixo:

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

Apesar de este prompt cobrir muitos aspectos do jogo, é comum que o agente peça detalhes adicionais, como "qual deve ser a resolução da tela" ou "você prefere animações fluidas ou estáticas".

Assim que estivermos satisfeitas com o nível de detalhe do plano, podemos pedir ao agente para começar a programar, o que encerra o plan mode. A partir daí, o processo segue como qualquer tarefa comum de programação. Após algumas iterações, temos um jogo funcionando parecido com este:

![Cloud Crush Gameplay Screenshot](cloud-crush-gameplay.png)

## Automatizando testes web com o browser agent

Uma das etapas mais desafiadoras no desenvolvimento de jogos são os testes. Não é possível escrever um teste unitário convencional para cobrir todos os estados visuais possíveis do jogo ou validar se as funções de renderização estão desenhando cada elemento no lugar correto da tela. Você até poderia tentar, mas garanto que seria um processo exaustivo, frágil e improdutivo.

Isso não significa abandonar testes automatizados, mas sim reconhecer a fronteira entre o que deve ser validado via código e o que exige playtesting humano. Por exemplo: testes unitários para lógica de colisão e pathfinding fazem todo o sentido; já validar a legibilidade da interface em diferentes resoluções é uma tarefa muito melhor avaliada visualmente (afinal, como criar um teste unitário para "esta fonte está confortável para leitura?").

Ou, pelo menos, era assim até agora... _um subagente entra na conversa_

Com as capacidades multimodais dos modelos de fronteira e o uso estratégico de agentes, podemos finalmente automatizar a validação visual. No Gemini CLI, um subagente é uma persona especializada que executa de forma independente da conversa principal, dentro de sua própria janela de contexto. Subagentes podem ser utilizados para estender o fluxo básico de desenvolvimento com as mais diversas capacidades.

No nosso cenário de testes, podemos usar um agente experimental integrado ao CLI chamado `@browser_agent`. Por ser experimental, é necessário [habilitá-lo manualmente](https://geminicli.com/docs/core/subagents/#enabling-the-browser-agent) editando seu arquivo `settings.json`. Veja abaixo um exemplo minimalista de configuração que habilita o browser agent com um modelo visual:

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

Normalmente, o browser agent navega por uma página web interpretando a árvore de acessibilidade — a estrutura subjacente utilizada por leitores de tela. No entanto, o nosso jogo Match 3 é renderizado inteiramente dentro de um único canvas HTML. Do ponto de vista da árvore de acessibilidade, trata-se apenas de um grande bloco em branco.

É exatamente aqui que a integração com um modelo de visão faz toda a diferença. Ao configurar o agente com um `visualModel` (como `gemini-2.5-computer-use-preview-10-2025`), ele passa literalmente a "enxergar". O agente captura screenshots da página, analisa o layout visual e deduz as coordenadas X e Y exatas onde precisa clicar na tela.

Em vez de navegar e clicar manualmente na aplicação hospedada no Cloud Run, basta enviar um comando como `@browser_agent please test the live URL...` para que ele abra o site, jogue uma partida e capture screenshots das telas em execução.

Isso não elimina o playtesting humano para avaliar o "feeling" do jogo, mas automatiza a checagem visual, comprovando que a interface está renderizando perfeitamente sem que você precise sair do terminal.

## Terceirizando minha ansiedade com segurança

Com a implementação funcional e a interface validada, não podemos negligenciar a segurança.

Eu não sou uma especialista em segurança de aplicações, o que me torna a pessoa menos indicada para auditar a postura de segurança de um web app. No entanto, assim como o agentic coding supriu minha falta de experiência com game engines, subagentes podem suprir minha carência técnica em segurança. Como orquestradora, não preciso dominar todos os vetores de Cross-Site Scripting; preciso apenas saber como instanciar um especialista com contexto limpo e focado em encontrá-los.

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

Nós fornecemos a ele um prompt de sistema específico (o corpo do arquivo markdown) e ferramentas como `read_file` e `grep_search` (declaradas no frontmatter). Por rodar em seu próprio ciclo de contexto, ele não sobrecarrega o histórico principal da conversa.

Apontei esse auditor para a base de código do *Cloud Crush* para verificar a existência de credenciais hardcoded, operações de arquivo inseguras e riscos de deploy. Embora um agente customizado de segurança não substitua uma equipe humana especializada, ele oferece uma camada essencial de proteção que, de outra forma, eu não teria.

## Um novo fluxo de desenvolvimento

Esse fluxo consolida o que considero ser o novo paradigma do desenvolvimento de software. Usamos agentes para produzir código e construímos ativamente ferramentas, skills e subagentes customizados para garantir nossos padrões arquiteturais e de qualidade.

E para quem lê com atenção: você deve ter notado que fui intencionalmente sucinta nas instruções passo a passo neste artigo. Isso porque temos um codelab completo dedicado a essa prática, que você pode acessar no link abaixo. Nele, você poderá executar cada etapa descrita aqui, construindo sua própria versão do jogo Match 3 do início ao fim.

**Codelab: [Construa um Jogo Arcade Match 3 Com o Gemini CLI](https://codelabs.developers.google.com/next26/gemini-cli-match3-golang#0)**

Se tiver dúvidas ou comentários, sinta-se à vontade para me procurar em qualquer uma das minhas redes sociais.
