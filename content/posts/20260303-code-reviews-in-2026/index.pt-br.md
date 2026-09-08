---
categories:
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
description: "Guia prático para revisar código gerado por IA e pull requests. Saiba quais áreas de arquitetura inspecionar e quais tarefas mecânicas delegar para linters."
proficiencyLevel: "Intermediate"
dependencies:
  - "golangci-lint"
  - "Ruff"
---

Em 2025, vimos a ascensão do agentic coding (aparentemente o termo "vibe coding" já ficou obsoleto). Entre assistentes de IA e fluxos de trabalho agênticos, funcionalidades estão saindo do forno em um ritmo nunca antes visto. Não é raro ver empresas se gabando da porcentagem de sua base de código escrita inteiramente por IA.

Se isso é bom ou ruim, ainda vamos ver (eu, particularmente, acho ótimo), mas esse ganho na velocidade de escrita traz consequências: revisar o volume absurdo de código produzido é exaustivo, e os code reviews estão rapidamente se tornando o grande gargalo. Algumas equipes e projetos open source chegaram a adotar a opção radical de simplesmente não aceitar pull requests gerados por IA.

Embora banir a IA possa dar um fôlego temporário às pessoas, não acho que seja uma boa opção a longo prazo. "A resistência é inútil", como diria minha espécie de ficção científica favorita. Para sobreviver a esse novo patamar de produtividade, precisamos parar de fazer o trabalho que as máquinas fazem melhor. Combata IA com IA. Mas não apenas com IA: um bom conjunto de ferramentas determinísticas tradicionais também faz maravilhas: se um linter consegue pegar um problema, eu não deveria perder tempo olhando para isso. Se um formatador corrige automaticamente, eu realmente não me importo com isso.

Minha opinião "impopular": eu não ligo se foi um humano ou um agente que escreveu o código. No open source, contribuições são tratadas com confiança zero (*zero-trust*). Se o código foi escrito por uma pessoa engenheira sênior de Big Tech ou por um estudante de ensino médio no Sri Lanka, não deveria fazer diferença. Então por que deveríamos nos importar se foi escrito por IA?

Em teoria, PRs gerados por humanos seriam menores, mas depois de quase 20 anos trabalhando nesta indústria, já vi minha cota de mega-PRs — então posso dizer com tranquilidade que lidar com PRs gigantescos e/ou desleixados não é um problema novo.

Eu avalio o código pelo seu valor real. Ele funciona? É seguro? Resolve um problema conhecido? Está alinhado ao nosso roadmap? Cumpre os nossos padrões?

É por isso que, no artigo de hoje, gostaria de falar um pouco sobre como venho abordando code reviews, não apenas ao lidar com contribuições externas, mas também ao avaliar meu próprio código gerado por IA — já que, na verdade, programar com IA significa fazer code review da IA o tempo todo.

## Com o que eu realmente me importo

Quando reviso código hoje em dia, meu olhar é cada vez mais de alto nível. Em certo sentido, quanto menos código escrevo manualmente, menos me apego aos detalhes microscópicos da implementação. Sempre disse, em cada equipe em que assumi um papel de liderança de uma forma ou de outra: código é descartável. Isso nunca foi tão verdadeiro quanto hoje. Eu repito: código é descartável. O que não é descartável é o conhecimento do sistema que você adquiriu ao desenvolver determinado código. Esse conhecimento é o que normalmente se transfere bem de uma implementação para outra ou, por exemplo, o que permanece ao migrar da v1 para a v2 da sua API.

Escrever algo pela segunda vez é mais fácil porque você já passou pelas dores de aprendizado de descobrir um monte de coisas e reduzir boa parte da ambiguidade. Você aprendeu o que funcionou bem e o que não funcionou. O que ficou com complexidade excessiva (*overengineered*) e o que ficou simplista demais (*underengineered*). Essa é a parte essencial da engenharia de software: acumular conhecimento, iterar, evoluir. E esse é o tipo de conhecimento que sobreviverá à era da IA. Código é apenas um detalhe de implementação.

Com base nessa filosofia, esta é uma lista não exaustiva das coisas com as quais me importo ao fazer code review:

### Arquitetura e design de sistemas
Modelos de IA têm dificuldade com a visão sistêmica ampla (*the big picture*) e também têm a tendência de pegar muitos atalhos. Meu processo de revisão busca ativamente esses sinais, como valores e configurações fixados no código (*hardcoded*), simplificação excessiva do espaço do problema (a IA frequentemente trata pedidos de código como meros protótipos ou demos) e, paradoxalmente, complexidade desnecessária (*over-engineering*). Modelos de IA também têm o traço irritante de presumir que código pronto para produção é sinônimo de complexidade. Em outras palavras, eles sofrem para dosar equilíbrio e pragmatismo — virtudes que aprendemos com a experiência e que muitas vezes são difíceis de traduzir em palavras.

### API pública e módulos
A ergonomia daquilo que estamos construindo é fundamental. A API pública precisa soar natural para o desenvolvedor médio que terá que usá-la. Uma interface bem projetada é intuitiva, difícil de usar incorretamente e esconde os detalhes internos do restante da base de código. Avalio se as interfaces são robustas e têm o escopo correto, buscando a menor superfície de contato (*surface area*) possível. Se a API for desajeitada, não importa o quão elegante seja o código por baixo dos panos. O código é fácil de usar e bem documentado? Um ótimo indicativo de uma boa API pública é a qualidade dos testes: uma API mal projetada é inerentemente difícil de testar.

### Algoritmos e padrões
Os LLMs frequentemente recorrem à forma mais ingênua e de força bruta para resolver um problema. É comum um agente tentar rodar uma migração massiva de dados usando loops aninhados e fazendo commit a cada poucas linhas, quando uma estratégia de *bulk insert* seria a abordagem correta. Ou em um nível mais elementar: usar uma lista quando um mapa ou dicionário é a estrutura de dados correta. Garantir que as estruturas de dados e os algoritmos realmente se adequem ao espaço do problema evita quedas brutais de desempenho. O objetivo é código que escala, e não apenas código que passa nos testes. No entanto, otimização prematura continua sendo um risco: se uma abordagem mais simples for ligeiramente mais lenta, mas muito mais legível, e estivermos lidando com um conjunto de dados pequeno e delimitado, a legibilidade geralmente vence.

### Dependências
Cada novo pacote traz risco externo, potenciais falhas de segurança e custo de manutenção. Manter a aplicação pequena reduz nossa superfície de ataque. Ferramentas essenciais, como nossos SDKs de GenAI ou grandes frameworks web, passam com mais facilidade na triagem, mas todo o restante é minuciosamente inspecionado. Uma pequena duplicação (ou reimplementação) é bem melhor do que uma dependência desnecessária (*a little copying is better than a little dependency*). Quanto mais fácil fica gerar e manter código, menos me preocupo em reutilizar bibliotecas externas a qualquer custo, especialmente se isso significar adicionar um novo vetor de ataque à minha base de código.

### Anti-patterns e problemas de qualidade
Apenas para citar alguns: erros ignorados ou silenciados, efeitos colaterais ocultos, estado global, estado mutável, vazamentos de recursos, funções ou variáveis não utilizadas, e assim por diante. Os idiomatismos da linguagem também importam. Embora eu dê enorme importância a esses pontos, eles também estão entre os mais fáceis de automatizar com o uso de análise estática (linters), como [golangci-lint](https://golangci-lint.run/) (Go) e [ruff](https://docs.astral.sh/ruff/) (Python).

### Testabilidade
Código difícil de testar geralmente foi mal projetado e vai resistir a alterações no futuro. Separação clara de responsabilidades, entradas limpas e funções puras são o ideal. Bons testes provam que o código funciona e criam uma rede de segurança para alterações futuras. Para componentes de UI e sistemas complexos, prefiro estratégias práticas de teste a metas rígidas de cobertura unitária, mas a lógica central precisa estar coberta.

Parei de tentar estabelecer uma meta de cobertura para cada projeto, porque cada caso é um caso, mas preciso saber que o que deve ser testado está realmente testado. Idealmente, 100% do caminho feliz (*happy path*) e uma boa porcentagem dos fluxos de exceção (*sad path*), mas sem a obsessão ingênua de tentar alcançar 100% de todo o código ou algo perto disso. Desde que você tenha uma estratégia sólida de observabilidade e boas mensagens de erro, você se prepara para o sucesso, já que novos modos de falha podem ser adicionados à suíte de testes com facilidade mais tarde.

### Benchmarking
Para caminhos críticos, precisamos de números reais em vez de palpites sobre desempenho. Benchmarks claros para quaisquer alterações que afetem componentes de alto tráfego são obrigatórios para impedir que código lento chegue à produção.

### Logging enxuto
Logs devem ser acionáveis. Logs desnecessários inflam a conta de nuvem e podem expor informações privadas. Logging detalhado é aceitável durante o desenvolvimento, mas precisa ser limpo antes do merge.

## Com o que eu (quase) não me importo

Deixo as ferramentas automatizadas cuidarem dos pormenores para poder me concentrar nos problemas realmente difíceis. Se uma máquina pode fazer, um ser humano não deveria estar fazendo.

### Cada linha individual de código
Revisar linha por linha o que um LLM gerou é trabalho para compiladores e analisadores estáticos. Em vez disso, eu foco na lógica e nos pontos de integração.

### Formatação
Desde que comecei a programar em Go, nunca mais me envolvi em discussões sobre estilos de formatação, embora saiba que elas ainda existam em certos lugares. A melhor coisa a fazer é definir um padrão e deixar o linter e o formatador cuidarem disso. Uma vez estabelecido o padrão, o agente de código também consegue ser mais aderente a ele. Se o pipeline de CI passou, está ótimo.

### Detalhes menores de sintaxe e código
Existem muitas maneiras de resolver um problema, e forçar escolhas específicas de sintaxe limita a liberdade da pessoa desenvolvedora. Não me importo se é um loop `for` ou uma *list comprehension*, desde que a lógica seja impecável.

### Debugging
Quase nunca faço sessões formais de depuração (no sentido estrito de realmente usar um debugger). Debugging para mim é o último recurso — e muitas vezes acaba sendo sinônimo de espalhar dezenas de prints "ESTOU AQUI", que na verdade deveriam ser linhas de log estruturadas.

Se algo não está funcionando, crio um novo teste para simular o problema. Se, após reproduzir o problema, ainda não consigo entender a causa raiz, significa que minha observabilidade e meus logs estão insuficientes — e é exatamente em melhorá-los que foco meu esforço.

### Nomes de escopo interno (não exportados)
Quando um identificador é local a uma função ou tem escopo restrito, me preocupo muito menos do que com nomes expostos em múltiplos arquivos e funções. Passo o olho rapidamente e, se notar algo absurdo, posso até sugerir um ajuste, mas fora isso estou tranquila com o que o modelo escolheu usar.

### Dependências secundárias
Aquelas que não são seus frameworks principais ou SDKs de clientes. Elas geram menos preocupação desde que atendam aos requisitos mínimos de segurança — auditoria contra vulnerabilidades e licenças problemáticas continua sendo obrigatória. Se estou importando um pacote apenas por causa de uma única função "helper", em 100% das vezes prefiro reimplementar essa função no meu código e eliminar a dependência externa.

## Conclusões

Este não é um protocolo engessado para todas as situações. Há também muito a ser considerado sobre a instrumentação da sua base de código. Code reviews sozinhos não capturam todos os problemas potenciais, e é por isso que defendo veementemente a automação — ainda mais na era do agentic coding.

Agentes de codificação modernos oferecem diversos mecanismos de extensão que permitem restringir o modelo e obter saídas mais determinísticas: [Agent Skills]({{< ref "/posts/20260829-the-pragmatic-guide-to-agent-skills" >}}), [hooks]({{< ref "/posts/20260610-mastering-hooks" >}}), [MCP tools]({{< ref "/posts/20250817-hello-mcp-world/" >}}), políticas, regras... Use essas ferramentas para delimitar com precisão o escopo dos seus agentes e a sua rotina será infinitamente mais tranquila.

Um carro só pode correr na velocidade em que seus freios conseguem pará-lo. Invista em aprender as barreiras de proteção (*guardrails*) do seu agente de codificação favorito e use seu tempo valioso para revisar aquilo que não pode ser automatizado.

Happy coding!

Dani =^.^=