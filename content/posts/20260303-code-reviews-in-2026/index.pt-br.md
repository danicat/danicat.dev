---
title: "Como Fazer Code Reviews na Era Agêntica"
date: 2026-03-06T00:00:00Z
summary: "Um guia prático para code reviews modernos. Aprenda onde investir seu tempo e esforço para escrever consistentemente software de qualidade no mundo agêntico."
categories: ["Workflow & Best Practices"]
tags: ["code-review", "vibe-coding", "agentic-coding"]
heroStyle: "big"
---

Em 2025 vimos a ascensão da codificação agêntica (aparentemente o termo "vibe coding" está obsoleto). Entre assistentes de IA e fluxos de trabalho agênticos, as funcionalidades estão saindo do forno a um ritmo que nunca vimos antes. Não é incomum que empresas se gabem de que qualquer quantidade de pontos percentuais de sua base de código agora é escrita inteiramente por IA.

Ainda não sabemos se isso é uma coisa boa ou ruim (eu, por exemplo, acho que é bom), mas esse ganho na velocidade de escrita não vem sem consequências: revisar a enorme quantidade de código produzido é exaustivo, e os code reviews estão rapidamente se tornando o gargalo. Algumas equipes e projetos de código aberto até adotaram a opção nuclear de não aceitar pull requests gerados por IA.

Embora banir a IA possa dar às pessoas um respiro, não acho que seja uma boa opção a longo prazo. "A resistência é inútil", como diria minha espécie fictícia favorita. Para sobreviver a esse novo nível de produtividade, temos que parar de fazer o trabalho que as máquinas podem fazer melhor. Combata IA com IA. Mas não apenas IA, um bom conjunto de ferramentas determinísticas à moda antiga também pode fazer maravilhas: se um linter pode pegar um problema, eu não deveria estar olhando para ele. Se um formatador pode corrigi-lo automaticamente, eu realmente não me importo com isso.

Aqui está a minha opinião "impopular". Eu não me importo se um humano ou um agente escreveu o código. Em código aberto, tudo é zero-trust (confiança zero) de qualquer maneira. Se o código foi escrito por um modelo de última geração ou por um adolescente no Sri Lanka, isso realmente não deveria importar. Em teoria, PRs gerados por humanos seriam menores, mas depois de uns 20 anos trabalhando nesta indústria eu tive minha cota de mega-PRs, então posso garantir que lidar com PRs gigantes e/ou desleixados não é um problema novo!

Eu avalio o código pelo seu valor de face. Ele funciona? É seguro? Corrige um problema conhecido? Está alinhado com o nosso roadmap? Está em conformidade com os nossos padrões?

É por isso que no artigo de hoje eu gostaria de falar um pouco sobre como estou abordando os code reviews, não apenas ao lidar com contribuições externas, mas também ao lidar com meu próprio código gerado por IA, pois, na verdade, codificar com uma IA significa fazer code review da IA o tempo todo.

## O Que Eu Realmente Me Importo

Quando eu reviso código hoje em dia, estou me tornando cada vez mais focada em alto nível. De certa forma, quanto menos código eu escrevo manualmente, menos me importo com os aspectos individuais do código. Eu sempre disse, em cada equipe em que assumi um papel de liderança em uma capacidade ou outra: código é descartável. Isso nunca foi tão verdadeiro quanto hoje. Repito: código é descartável. O que não é descartável é o conhecimento do sistema que você adquiriu ao desenvolver certo código. Esse conhecimento é o que geralmente se traduz bem de uma implementação para outra ou, por exemplo, o que fica quando você migra da v1 para a v2 da sua API.

Escrever algo pela segunda vez é mais fácil porque você já passou pelas dores de crescimento de descobrir um monte de coisas e reduzir grande parte da ambiguidade. Você aprendeu o que funcionou bem e o que não funcionou. O que foi excessivamente planejado e o que foi subplanejado. Essa é a parte importante na engenharia de software: reunir conhecimento, iterar, evoluir. E este é o tipo de conhecimento que sobreviverá à era da IA. O código é apenas um detalhe de implementação.

Com base nessa filosofia, esta é uma lista não exaustiva das coisas com as quais me importo ao fazer code reviews:

### Arquitetura e Design de Sistemas
Modelos de IA têm dificuldade com o panorama geral e também têm a tendência de pegar muitos atalhos. Meu processo de revisão procura por esses sinais como valores e configurações hardcoded, simplificação excessiva do espaço do problema (a IA frequentemente trata pedidos de codificação como protótipos ou demos) e, paradoxalmente, engenharia excessiva (over-engineering). Modelos de IA também têm a característica irritante de assumir que prontidão para produção é igual a complexidade. Ou em outras palavras, eles lutam com equilíbrio e pragmatismo, coisas que aprendemos com a experiência e que muitas vezes são difíceis de traduzir em palavras.

### APIs Públicas e Módulos
A ergonomia do que estamos construindo importa. A API pública precisa "parecer certa" para o desenvolvedor médio que tem que usá-la. Uma interface bem projetada é intuitiva, difícil de usar incorretamente e esconde os internos bagunçados do resto da base de código. Eu verifico se as interfaces são robustas e têm o escopo correto, visando a menor área de superfície possível. Se a API é desajeitada, não importa o quão elegante seja o código subjacente. O código é fácil de usar e bem documentado? Uma boa dica de que a API pública é boa é se a qualidade do teste é boa. Um design de API ruim é inerentemente difícil de testar.

### Algoritmos e Padrões
LLMs costumam usar por padrão a maneira mais ingênua e de força bruta para resolver um problema. É comum um agente tentar executar uma migração massiva de dados usando loops aninhados e fazendo commit a cada poucas linhas, quando uma estratégia de inserção em lote (bulk insert) é a abordagem correta. Ou em um nível central: usar uma lista quando um mapa ou dicionário é a estrutura de dados correta. Verificar se as estruturas de dados e os algoritmos realmente se encaixam no espaço do problema evita quedas massivas de desempenho. O objetivo é código que escala, não apenas código que passa nos testes. No entanto, otimização prematura ainda é um risco. Se uma abordagem mais simples é um pouco mais lenta, mas muito mais fácil de ler, e estamos lidando com um conjunto de dados pequeno e limitado, a legibilidade geralmente vence.

### Dependências
É incrivelmente fácil para um agente puxar uma biblioteca massiva de terceiros para uma tarefa que a biblioteca padrão poderia resolver em três linhas de código. Cada novo pacote traz riscos externos, possíveis falhas de segurança e sobrecarga extra de manutenção. Toda adição deve ser ativamente mantida, segura e verdadeiramente necessária. Manter o aplicativo pequeno reduz a nossa superfície de ataque. Ferramentas essenciais como nossos SDKs de GenAI ou grandes frameworks web ganham um passe mais rápido. Tudo o mais é examinado. Um pouco de cópia (ou reimplementação) é melhor do que uma pequena dependência. Quanto mais fácil fica gerar e manter código, menos me preocupo com a reutilização, especialmente se isso significar adicionar um novo vetor de ataque à minha base de código.

### Anti-Padrões e Problemas de Qualidade
Apenas para citar alguns: objetos deus (god objects), mudanças de estado ocultas, efeitos colaterais, estado global, vazamentos de recursos, ignorar erros, funções ou variáveis não utilizadas e assim por diante. Expressões idiomáticas da linguagem também importam; o que parece uma armadilha em um idioma pode ser a maneira padrão de fazer as coisas em outro. Embora eu me importe muito com isso, esses também são alguns dos mais fáceis de automatizar com o uso de análise estática (linters) como golanci-lint (Go) e ruff (python).

### Testabilidade
Código que é difícil de testar geralmente é mal projetado e resistirá a mudanças no futuro. Separação clara de preocupações, entradas limpas e funções puras são ideais. Bons testes provam que o código funciona e fornecem uma rede de segurança para nossas modificações futuras. Para componentes de UI e sistemas complexos, aceito estratégias de teste práticas em vez de cobertura de unidade estrita, mas a lógica central deve ser coberta. Parei de tentar adicionar uma meta de cobertura para cada projeto, porque cada caso é um caso, mas preciso saber que o que quer que possa ser testado está sendo testado. Idealmente, 100% do caminho feliz e uma boa porcentagem do caminho triste, mas não vou tentar alcançar 100% ou chegar perto disso. Contanto que você tenha uma boa estratégia de observabilidade e boas mensagens de erro, você está se preparando para o sucesso, pois novos modos de erro podem ser adicionados à suíte de testes mais tarde.

### Benchmarking
Para caminhos críticos, precisamos de números reais em vez de suposições sobre desempenho. Benchmarks claros para quaisquer alterações que afetem componentes de alto tráfego são necessários para impedir que código lento chegue à produção. Isso só é necessário para os caminhos quentes (hot paths); pedir benchmarks para cada função auxiliar menor é perda de tempo para todos.

### Logging Enxuto
Os logs devem ser acionáveis. Logs desnecessários aumentam nossas contas na nuvem e podem vazar informações privadas. Eu verifico os níveis de log para garantir que capturemos exatamente o que é necessário, sem incluir segredos ou PII. Logs detalhados (verbose) são bons durante o desenvolvimento, mas devem ser limpos antes do merge.

## O Que Eu Não Me Importo (Na Maioria das Vezes)

Eu deixo as ferramentas automatizadas lidarem com os detalhes para que eu possa focar nos problemas difíceis. Se uma máquina pode fazer isso, um humano não deveria estar fazendo.

### Formatação
Desde que comecei a usar Go, nunca estive em uma discussão sobre estilos de formatação, mas sei que elas ainda existem em certos espaços. A melhor coisa que você pode fazer é estabelecer um padrão e deixar o linter e o formatador lidarem com isso. Uma vez que o padrão seja conhecido, o agente de código também pode ser mais compatível com ele. Se o pipeline de CI passar, está bom. Isso acelera a revisão e impede debates inúteis sobre estilo.

### Pequenos Detalhes de Sintaxe e Código
Existem muitas maneiras de resolver um problema, e forçar escolhas de sintaxe específicas limita a liberdade do desenvolvedor. Eu não me importo se é um loop `for` ou uma list comprehension, desde que a lógica seja sólida.

### Cada Linha de Código Individual
Revisar cada linha gerada por um LLM é trabalho de um compilador ou analisador estático. Em vez disso, eu foco na lógica e nos pontos de conexão.

### Depuração (Debugging)
Eu quase nunca faço sessões de depuração. Se algo não está funcionando, eu crio um novo teste para simular o problema. Se, após reproduzir o problema, não consigo descobrir o que está acontecendo, isso significa que minha observabilidade e logs são insuficientes, então eu foco em melhorá-los. Para mim, a depuração é um último recurso e é sinônimo de adicionar um monte de instruções de impressão "EU ESTOU AQUI", que na verdade deveriam ter sido linhas de log. Cuidado com o estado mutável e mantenha seus estados transicionais pelo menos temporariamente. Isso facilitará muito a sua vida e dispensará a necessidade de depuradores.

### Nomes Não Exportados (Com Ressalvas)
Nomes internos têm escopo limitado e raramente afetam o design geral. Eu passo o olho neles desde que a API pública seja sólida e o contexto esteja claro. Isso mantém a revisão em andamento e evita o nitpicking sobre escolhas menores. No entanto, nomes ruins quase sempre levam a um design ruim, e sempre levam a um código difícil de manter, então eu ainda me importo um pouco com eles.

### Dependências Menores
Aquelas que não são suas principais dependências de framework ou biblioteca de cliente. Elas são de menor preocupação se atenderem à nossa linha de base de segurança. No entanto, verificar a existência de vulnerabilidades de segurança e licenças problemáticas ainda é obrigatório. Se estou importando algo apenas por uma função "auxiliar", em 100% das vezes reimplementarei essa função no meu código e me livrarei da dependência.

## Conclusões

Este não é um protocolo que serve para todos, mas é como estou abordando as revisões de código nos dias de hoje. Também há muito a ser dito sobre como você está instrumentando sua base de código. Os code reviews sozinhos não capturarão todos os problemas potenciais e é por isso que defendo fortemente a automação, agora com a codificação agêntica mais do que nunca. Agentes de codificação modernos têm muitos padrões de extensão que permitem restringir o modelo e obter saídas mais determinísticas: Agent Skills, hooks, ferramentas MCP, políticas, regras... Pode ficar um pouco confuso, já que muitos deles não são padrões, mas estamos chegando ao ponto em que muitos deles estão sendo padronizados.

Um carro só pode correr o mais rápido que seus freios suportarem. Invista em aprender as proteções (guardrails) do seu agente de codificação favorito e use seu precioso tempo para revisar o que não pode ser automatizado.

Happy coding!

Dani =^.^=