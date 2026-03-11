---
title: "Como Fazer Code Reviews na Era dos Agentes"
date: 2026-03-06T00:00:00Z
summary: "Um guia prático para code reviews modernos. Aprenda onde investir seu tempo e esforço para escrever software de qualidade de forma consistente no mundo dos agentes."
categories: ["Workflow & Best Practices"]
tags: ["code-review", "vibe-coding", "agentic-coding"]
heroStyle: "big"
---

Em 2025 vimos a ascensão do agentic coding (aparentemente o termo "vibe coding" está obsoleto). Entre assistentes de IA e workflows agentic, as features estão voando das prateleiras a um ritmo que nunca vimos antes. Não é incomum que as empresas se gabem de quantos pontos percentuais de sua codebase são agora escritos inteiramente por IA.

Se isso é uma coisa boa ou ruim ainda vamos ver (eu, por exemplo, acho que é uma coisa boa), mas esse ganho na velocidade de escrita não vem sem consequências: revisar a enorme quantidade de código produzido é exaustivo, e os code reviews estão rapidamente se tornando o gargalo. Alguns times / projetos open source até adotaram a opção nuclear de não aceitar PRs gerados por IA de forma alguma.

Embora banir a IA possa dar um respiro às pessoas, não acho que seja uma boa opção a longo prazo. "A resistência é inútil", como diria minha espécie fictícia favorita. Para sobreviver a este novo nível de produtividade temos que parar de fazer o trabalho que as máquinas podem fazer melhor. Combater IA com IA. Mas não apenas IA, um bom conjunto de ferramentas determinísticas da velha escola também pode fazer maravilhas: se um linter pode detectar um problema, eu não deveria estar olhando para isso. Se um formatter pode consertar automaticamente, eu realmente não me importo com isso.

Minha opinião "impopular": eu realmente não me importo se um humano ou um agente escreveu o código. Em open source, as contribuições são zero-trust. Se o código foi escrito por um engenheiro sênior da FAANG ou por um estudante do ensino médio no Sri Lanka, realmente não deveria importar. Então, por que deveríamos nos importar se foi escrito por IA?

Na teoria, PRs gerados por humanos seriam menores, mas depois de uns 20 anos trabalhando nesta indústria eu já tive a minha cota de mega-PRs, então posso dizer com confiança, por estar tão acostumada com isso, que lidar com PRs gigantes e/ou desleixados não é um problema novo.

Eu avalio o código pelo seu valor nominal. Ele funciona? É seguro? Ele corrige um problema conhecido? Está alinhado com o nosso roadmap? Ele atende aos nossos padrões?

É por isso que no artigo de hoje eu gostaria de falar um pouco sobre como estou focada em abordar os code reviews, não apenas ao lidar com contribuições externas, mas também ao lidar com o meu próprio código gerado por IA, pois, de fato, codar com uma IA significa revisar o código da IA o tempo todo.

## Com o que eu realmente me importo

Quando eu reviso código hoje em dia, estou me tornando cada vez mais de alto nível. De certa forma, quanto menos código eu escrevo manualmente, menos eu me importo com os aspectos individuais do código. Eu sempre disse, em cada time em que assumi um papel de liderança em uma capacidade ou outra: código é descartável. Isso nunca foi tão verdadeiro quanto hoje. Repito: código é descartável. O que não é descartável é o conhecimento do sistema que você adquiriu ao desenvolver certo código. Esse conhecimento é o que geralmente se traduz bem de uma implementação para outra ou, por exemplo, o que fica quando migramos da v1 para a v2 da nossa API.

Escrever algo pela segunda vez é mais fácil porque você já passou pelas dores do crescimento de descobrir um monte de coisas e reduzir muita ambiguidade. Você aprendeu o que funcionou bem e o que não funcionou. O que foi overengineered e o que foi underengineered. Esta é a parte importante na engenharia de software: reunir conhecimento, iterar, evoluir. E esse é o tipo de conhecimento que sobreviverá à era da IA. Código é apenas um detalhe de implementação.

Com base nessa filosofia, esta é uma lista não exaustiva das coisas com as quais me importo ao fazer um code review:

### Arquitetura e design de sistema
Modelos de IA têm dificuldade com o quadro geral e também têm a tendência de pegar muitos atalhos. Meu processo de revisão procura por esses sinais como valores e configurações hardcoded, simplificação excessiva do espaço do problema (IA muitas vezes trata as requisições de código como protótipos ou demos) e, paradoxalmente, over-engineering. Modelos de IA também têm a característica irritante de assumir que estar pronto para produção é igual a complexidade. Em outras palavras, eles têm dificuldade com equilíbrio e pragmatismo, coisas que nós aprendemos com a experiência e que muitas vezes é difícil traduzir em palavras. Como uma engenheira experiente, fico sempre atenta a esses detalhes.

### API Pública e módulos
A ergonomia do que estamos construindo importa. A API pública precisa "parecer certa" para o desenvolvedor médio que tem que usá-la. Uma interface bem projetada é intuitiva, difícil de usar mal, e esconde os internos bagunçados do resto da codebase. Eu verifico se as interfaces são robustas e têm o escopo correto, visando a menor área de superfície possível. Se a API é desajeitada, não importa o quão elegante seja o código subjacente. O código é fácil de usar e bem documentado? Uma boa dica de que a API pública é boa é se a qualidade do teste é boa. Um design ruim de API é inerentemente difícil de testar.

### Algoritmos e padrões
LLMs geralmente usam como padrão a forma mais ingênua e de força bruta para resolver um problema. É comum um agente tentar executar uma migração massiva de dados usando loops aninhados e fazendo commits a cada poucas linhas, quando uma estratégia de bulk insert é a abordagem correta. Ou em um nível mais básico: usar uma lista quando um mapa ou dicionário é a estrutura de dados correta. Verificar se as estruturas de dados e os algoritmos realmente se encaixam no espaço do problema previne quedas massivas de performance. O objetivo é código que escala, não apenas código que passa nos testes. No entanto, otimização prematura ainda é um risco. Se uma abordagem mais simples for um pouco mais lenta mas muito mais fácil de ler e estivermos lidando com um conjunto de dados pequeno e limitado, a legibilidade geralmente vence.

### Dependências
Todo novo pacote traz risco externo, possíveis falhas de segurança e overhead de manutenção. Manter o app pequeno reduz a nossa superfície de ataque. Ferramentas principais como nossos SDKs de GenAI ou os principais frameworks web passam mais rápido, mas todo o resto é escrutinado. Uma pequena cópia (ou reimplementação) é melhor do que uma pequena dependência. Quanto mais fácil fica gerar e manter o código, menos eu me preocupo com reuso, especialmente se isso significa adicionar um novo vetor de ataque à minha codebase.

### Anti-patterns e problemas de qualidade
Apenas para citar alguns: ignorar ou silenciar erros, side effects, estado global, estado mutável, vazamentos de recursos, funções ou variáveis não utilizadas, e assim por diante. Expressões idiomáticas da linguagem também importam. Embora eu me importe muito com isso, esses também são alguns dos mais fáceis de automatizar com o uso de análise estática (linters) como [golangci-lint](https://golangci-lint.run/) (Go) e [ruff](https://docs.astral.sh/ruff/) (Python).

### Testabilidade
Código difícil de testar geralmente é mal projetado e resistirá a mudanças no futuro. Separação clara de preocupações, entradas limpas e funções puras são o ideal. Bons testes provam que o código funciona e fornecem uma rede de segurança para nossas futuras modificações. Para componentes de UI e sistemas complexos, eu aceito estratégias práticas de teste ao invés de cobertura rigorosa de unidade, mas a lógica central deve ser coberta.

Eu parei de tentar adicionar uma meta de cobertura para cada projeto individual porque cada caso é um caso, mas eu preciso saber que o que quer que deva ser testado está sendo testado. Idealmente 100% do happy path e uma boa porcentagem do sad path, mas eu não vou tentar atingir 100% de todo o código ou algo perto disso. Desde que você tenha uma boa estratégia de observabilidade e boas mensagens de erro, você está se preparando para o sucesso, já que novos modos de erro podem ser adicionados à suíte de testes mais tarde.

### Benchmarking
Para caminhos críticos, precisamos de números reais ao invés de adivinhações sobre performance. Benchmarks claros para quaisquer mudanças que afetem componentes de alto tráfego são necessários para impedir que código lento chegue à produção.

### Logging enxuto
Logs devem ser acionáveis. Logs desnecessários aumentam nossas faturas na nuvem e podem vazar informações privadas. Logging verboso é bom durante o desenvolvimento, mas deve ser limpo antes do merge.

## Com o que eu não me importo (na maior parte)

Eu deixo que as ferramentas automatizadas lidem com os detalhes para que eu possa me concentrar nos problemas difíceis. Se uma máquina pode fazer isso, um humano não deveria estar fazendo.

### Cada linha individual de código
Revisar cada linha gerada por um LLM é trabalho de um compilador ou analisador estático. Em vez disso, eu me concentro na lógica e nos pontos de conexão.

### Formatação
Desde que comecei a usar Go, nunca entrei em uma discussão sobre estilos de formatação, mas eu sei que eles ainda existem em certos espaços. A melhor coisa que você pode fazer é definir um padrão e deixar o linter e o formatter lidarem com isso. Uma vez que o padrão é conhecido, o agente de código também pode ser mais compatível com ele. Se o pipeline de CI passar, está tudo bem.

### Pequenos detalhes de sintaxe e código
Existem muitas maneiras de resolver um problema, e forçar escolhas específicas de sintaxe limita a liberdade do desenvolvedor. Eu não me importo se é um `for` loop ou uma list comprehension, contanto que a lógica seja sólida.

### Debugging
Eu quase nunca faço sessões de debug (no sentido estrito de realmente usar um debugger). Fazer debugging para mim é o último recurso, e é sinônimo de adicionar um monte de statements print "ESTOU AQUI", que na verdade deveriam ter sido linhas de log.

Se algo não está funcionando, eu crio um novo teste para simular o problema. Se, após reproduzir o problema, eu não conseguir descobrir o que está acontecendo, isso significa que minha observabilidade e logs são insuficientes, então eu me concentro em melhorá-los.

### Nomes não exportados
Quando um nome é local a uma função ou tem escopo limitado, eu me importo muito menos com isso do que se fosse usado em muitas funções ou arquivos. Eu vou passar os olhos no código e, se eu tiver um vislumbre de algo absurdo, talvez eu pare para uma limpeza, mas caso contrário, eu estou tranquila com o que o modelo decidiu usar.

### Pequenas dependências
Aquelas que não são suas principais dependências de framework ou client library. Elas são menos preocupantes se atenderem à nossa baseline de segurança. Verificar se há exploits de segurança e licenças problemáticas ainda é obrigatório, no entanto. Se eu estiver importando algo apenas por uma função "helper", eu vou 100% das vezes reimplementar essa função no meu código e me livrar da dependência.

## Conclusões

Este não é um protocolo que serve para tudo. Também há muito o que se dizer sobre como você está instrumentando a sua codebase. Code reviews sozinhos não capturarão todos os problemas em potencial, e é por isso que eu defendo fortemente a automação, agora com agentic coding mais do que nunca.

Agentes de codificação modernos têm muitos padrões de extensão que permitem restringir o modelo e obter resultados mais determinísticos: [Agent Skills]({{< ref "/posts/20260128-agent-skills-gemini-cli/" >}}), hooks, [MCP tools]({{< ref "/posts/20250817-hello-mcp-world/" >}}), políticas, regras... Use essas ferramentas para colocar limites bem definidos no escopo de seus agentes e a sua vida se tornará muito mais fácil.

Um carro só pode correr tão rápido quanto seus freios suportam. Invista em aprender os guardrails para o seu agente de codificação favorito e use o seu precioso tempo para revisar o que não pode ser automatizado.

Happy coding!

Dani =^.^=