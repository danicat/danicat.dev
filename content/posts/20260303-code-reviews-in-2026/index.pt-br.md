---categories:
- Agentic Coding
date: 2026-03-06 00:00:00+00:00
heroStyle: big
summary: Um guia prático para code reviews modernos. Aprenda onde investir seu tempo
  e esforço para escrever software de qualidade de forma consistente no mundo dos
  agentes.
tags:
  - code-review
  - vibe-coding
title: "Como Fazer Code Reviews na Era Agêntica"
slug: "code-reviews-in-2026"
aliases:
  - "/pt-br/posts/20260303-code-reviews-in-2026/"
description: "Guia prático para code reviews na era da IA. Saiba quais áreas de arquitetura e API inspecionar manualmente e o que delegar para linters e formatadores."
proficiencyLevel: "Intermediate"
dependencies:
  - "golangci-lint"
  - "Ruff"
---

Em 2025, vimos a ascensão do desenvolvimento agêntico (aparentemente o termo "vibe coding" já ficou obsoleto). Entre assistentes de IA e fluxos de trabalho agênticos, funcionalidades estão saindo do forno em um ritmo nunca antes visto. Não é raro ver empresas se gabando da porcentagem de sua base de código escrita inteiramente por IA.

Se isso é bom ou ruim, o tempo dirá (eu, particularmente, acho ótimo), mas esse ganho na velocidade de escrita traz consequências: revisar o volume absurdo de código produzido é exaustivo, e os code reviews estão rapidamente se tornando o grande gargalo. Algumas equipes e projetos open source chegaram a adotar a opção radical de recusar sumariamente pull requests gerados por IA.

Embora banir a IA possa dar um fôlego temporário às pessoas, não acho que seja uma boa solução a longo prazo. "A resistência é inútil", como diria minha espécie de ficção favorita. Para sobreviver a esse novo patamar de produtividade, precisamos parar de fazer o trabalho que as máquinas fazem melhor. Combata IA com IA. Mas não apenas com IA: um bom conjunto de ferramentas determinísticas tradicionais também faz maravilhas. Se um linter é capaz de pegar um problema, eu não deveria perder tempo olhando para isso. Se um formatador corrige automaticamente, não tenho por que me preocupar.

Minha opinião "impopular": não me importo se o código foi escrito por um humano ou por um agente. No mundo open source, contribuições são tratadas com confiança zero (*zero-trust*). Se o código foi escrito por uma pessoa engenheira sênior de Big Tech ou por uma estudante de ensino médio no Sri Lanka, não deveria fazer diferença. Então por que deveríamos nos importar se foi gerado por IA?

Em teoria, PRs gerados por humanos seriam menores, mas depois de quase 20 anos nesta indústria, já vi minha cota de mega-PRs — então posso afirmar com tranquilidade que lidar com PRs gigantescos e/ou desleixados não é um problema novo.

Eu avalio o código pelo seu valor real. Ele funciona? É seguro? Resolve um problema conhecido? Está alinhado ao nosso roadmap? Cumpre os nossos padrões?

É por isso que, no artigo de hoje, quero falar um pouco sobre como venho abordando code reviews — não apenas ao lidar com contribuições externas, mas também ao avaliar meu próprio código gerado por IA. Afinal, codificar com IA significa, na prática, fazer code review da IA o tempo todo.

## Com o que eu realmente me importo

Quando reviso código hoje em dia, meu olhar é cada vez mais de alto nível. De certa forma, quanto menos código escrevo manualmente, menos me apego aos detalhes microscópicos da implementação. Sempre disse, em todas as equipes em que atuei em papéis de liderança técnica: código é descartável. Isso nunca foi tão verdadeiro quanto hoje. Repito: código é descartável. O que não é descartável é o conhecimento do sistema que você adquire ao desenvolvê-lo. Esse conhecimento é o que realmente se transfere de uma implementação para outra — ou o que permanece, por exemplo, ao migrar da v1 para a v2 da sua API.

Reescrever algo pela segunda vez é mais fácil porque você já passou pelas dores de aprendizado, descobriu armadilhas e eliminou boa parte da ambiguidade. Você aprendeu o que funcionou bem e o que não funcionou. O que ficou com engenharia em excesso (*overengineered*) e o que ficou simplista demais (*underengineered*). Essa é a parte essencial da engenharia de software: acumular conhecimento, iterar e evoluir. E esse é o tipo de conhecimento que vai sobreviver à era da IA. Código é apenas um detalhe de implementação.

Com base nessa filosofia, apresento uma lista (não exaustiva) do que realmente observo ao fazer code review:

### Arquitetura e design de sistemas
Modelos de IA têm dificuldade com a visão sistêmica ampla e adoram pegar atalhos. Meu processo de revisão busca ativamente sinais como valores e configurações fixados no código (*hardcoded*), simplificação excessiva do escopo do problema (a IA frequentemente trata pedidos de código como meros protótipos ou demos) e, paradoxalmente, complexidade desnecessária (*over-engineering*). Modelos de IA também têm o traço irritante de presumir que código pronto para produção é sinônimo de código complexo. Em outras palavras, eles sofrem para dosar equilíbrio e pragmatismo — virtudes que nós, desenvolvedoras, aprendemos com a experiência e que dificilmente cabem em palavras.

### API pública e módulos
A ergonomia do que construímos é fundamental. A API pública precisa soar natural para qualquer pessoa desenvolvedora que precise consumi-la. Uma interface bem projetada é intuitiva, difícil de usar de forma errada e esconde as entranhas da implementação do restante da base de código. Avalio se as interfaces são robustas e têm o escopo correto, buscando sempre a menor superfície de contato possível. Se a API é desajeitada, não importa o quão elegante seja o código por baixo dos panos. O código é simples de usar e bem documentado? Um ótimo indicativo de uma boa API pública é a qualidade dos testes: uma API mal projetada é inerentemente difícil de testar.

### Algoritmos e padrões
LLMs frequentemente recorrem à solução mais ingênua e de força bruta para resolver um problema. Não é raro ver um agente tentar rodar uma migração massiva de dados usando loops aninhados com commits a cada poucas linhas, quando a estratégia correta seria um *bulk insert*. Ou em um nível mais elementar: usar uma lista quando um mapa ou dicionário seria a estrutura de dados ideal. Garantir que estruturas de dados e algoritmos sejam adequados ao problema evita quedas bruscas de desempenho. O objetivo é código que escala, e não apenas código que passa nos testes. Ainda assim, otimização prematura continua sendo um risco: se uma abordagem mais simples for ligeiramente mais lenta, mas muito mais legível, e estivermos lidando com um volume pequeno e delimitado de dados, a legibilidade costuma vencer.

### Dependências
Cada novo pacote introduz risco externo, potenciais falhas de segurança e custo de manutenção. Manter a aplicação enxuta reduz nossa superfície de ataque. Ferramentas essenciais, como nossos SDKs de GenAI ou grandes frameworks web, passam com mais facilidade na triagem, mas todo o restante é minuciosamente inspecionado. Uma pequena duplicação (ou reimplementação) é bem melhor do que uma dependência desnecessária. Quanto mais fácil fica gerar e manter código, menos me preocupo em reutilizar bibliotecas externas a qualquer custo, especialmente se isso significar adicionar um novo vetor de ataque à minha base de código.

### Anti-patterns e problemas de qualidade
Para citar alguns: erros ignorados ou silenciados, efeitos colaterais ocultos, estado global, mutabilidade descontrolada, vazamento de recursos, funções ou variáveis não utilizadas, e assim por diante. Idiomatismos da linguagem também contam muito. Embora eu dê enorme importância a esses pontos, eles também estão entre os mais fáceis de automatizar com análise estática (linters), como [golangci-lint](https://golangci-lint.run/) (Go) e [ruff](https://docs.astral.sh/ruff/) (Python).

### Testabilidade
Código difícil de testar geralmente foi mal projetado e vai resistir a alterações no futuro. Separação clara de responsabilidades, entradas limpas e funções puras são o ideal. Bons testes comprovam que o código funciona e criam uma rede de segurança para alterações futuras. Para componentes de interface e sistemas complexos, prefiro estratégias práticas de teste a metas rígidas de cobertura unitária, mas a lógica central precisa estar coberta.

Parei de estipular metas fixas de cobertura para todo e qualquer projeto porque cada caso é um caso, mas exijo ter a certeza de que tudo o que precisa ser testado está realmente testado. Idealmente, 100% do caminho feliz e uma boa parcela dos cenários de erro — sem a obsessão ingênua de alcançar 100% de cobertura total da base. Desde que você tenha uma estratégia sólida de observabilidade e boas mensagens de erro, estará pavimentando o caminho para o sucesso, já que novos cenários de falha podem ser incorporados à suíte de testes com facilidade mais tarde.

### Benchmarking
Em caminhos críticos, precisamos de dados concretos em vez de palpites sobre desempenho. Benchmarks claros para quaisquer alterações que afetem componentes de alto tráfego são obrigatórios para impedir que código lento chegue à produção.

### Logging enxuto
Logs precisam ser acionáveis. Logs desnecessários inflam a conta de nuvem e podem expor informações confidenciais. Logging detalhado é ótimo durante o desenvolvimento, mas deve ser limpo antes do merge.

## Com o que eu (quase) não me importo

Deixo as ferramentas automatizadas cuidarem dos pormenores para poder me concentrar nos problemas realmente difíceis. Se uma máquina pode fazer, um ser humano não deveria estar fazendo.

### Cada linha individual de código
Revisar linha por linha o que um LLM gerou é trabalho para compiladores e analisadores estáticos. Meu foco está na lógica e nos pontos de integração.

### Formatação
Desde que comecei a programar em Go, nunca mais entrei em discussões sobre estilo de formatação, embora saiba que elas ainda existam em certos círculos. A melhor coisa a fazer é definir um padrão e deixar o linter e o formatador resolverem. Uma vez estabelecido o padrão, o agente de codificação também consegue aderir melhor a ele. Se o pipeline de CI passou, está ótimo.

### Detalhes menores de sintaxe e estilo
Existem inúmeras maneiras de resolver um mesmo problema, e forçar preferências estritas de sintaxe só limita a liberdade de quem desenvolve. Não me importa se foi usado um loop `for` ou uma *list comprehension*, desde que a lógica seja impecável.

### Debugging
Quase nunca faço sessões formais de depuração com debugger interativo. Para mim, usar debugger é o último recurso — e muitas vezes acaba sendo sinônimo de espalhar dezenas de prints "ESTOU AQUI", que na verdade deveriam ser linhas de log estruturadas.

Se algo não funciona, crio um novo teste reproduzindo o problema. Se, após reproduzir, ainda não consigo entender a causa raiz, significa que minha observabilidade e meus logs estão insuficientes — e é exatamente em melhorá-los que foco meu esforço.

### Nomes de escopo interno (não exportados)
Quando um identificador é local a uma função ou tem escopo restrito, me preocupo muito menos do que com nomes expostos em múltiplos arquivos e funções. Passo o olho rapidamente e, se notar algo absurdo, posso até sugerir um ajuste, mas fora isso não tenho problemas com o que o modelo escolheu.

### Dependências secundárias
Refiro-me àquelas que não são seus frameworks principais ou SDKs de clientes. Elas geram menos preocupação desde que atendam aos requisitos mínimos de segurança — auditoria contra vulnerabilidades e licenças problemáticas continua sendo obrigatória. Se estou importando um pacote apenas por causa de uma função utilitária simples, em 100% dos casos prefiro reimplementar essa função no meu código e eliminar a dependência externa.

## Conclusões

Este não é um roteiro engessado para todas as situações. Há também muito a ser considerado sobre a instrumentação da sua base de código. Code reviews sozinhos não capturam todos os problemas potenciais, e é por isso que defendo com tanto entusiasmo a automação — ainda mais na era do desenvolvimento agêntico.

Agentes modernos oferecem diversos mecanismos de extensão que permitem impor restrições e obter resultados mais determinísticos: [Agent Skills]({{< ref "/posts/20260128-agent-skills-gemini-cli/" >}}), hooks, [MCP tools]({{< ref "/posts/20250817-hello-mcp-world/" >}}), políticas e regras de contexto. Use essas ferramentas para delimitar com precisão o escopo dos seus agentes e a sua rotina será infinitamente mais tranquila.

Um carro só pode correr na velocidade em que seus freios conseguem pará-lo. Invista em aprender as barreiras de proteção (*guardrails*) do seu agente de codificação favorito e use seu tempo valioso para revisar aquilo que não pode ser automatizado.

Happy coding!

Dani =^.^=