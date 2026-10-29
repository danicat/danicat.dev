---
title: "Além de Escrever Código: Bom Gosto, Skills e Tech Leads"
date: 2026-10-28
draft: false
categories:
  - Perspectives
tags:
  - agent-skills
  - antigravity
  - career
  - code-review
  - subagents
  - vibe-coding
heroStyle: big
slug: "beyond-writing-code"
summary: "Quando modelos têm todo o conhecimento do mundo, o que resta para nós? Se tudo é conhecido, tudo parece prioritário. Nosso papel mudou da execução para a direção."
description: "Por que a engenharia de software na era dos agentes de código migra da sintaxe para o bom gosto arquitetural, subagentes, agent skills e a postura de tech lead."
proficiencyLevel: "Intermediate"
---

Quando os modelos de fundação trazem todo o conhecimento do mundo embutido em seus pesos, o que sobra para os humanos?

Por décadas, nossa identidade como engenheiros de software esteve atrelada à arte de escrever código: dominar sintaxes, memorizar bibliotecas padrão e nos orgulhar dos sistemas que construíamos com nossas próprias mãos. Disseram-nos que a IA simplesmente cuidaria do boilerplate chato e nos devolveria horas livres toda semana. Em vez disso, ver um modelo gerar em segundos o que levamos anos para aprender desencadeou uma verdadeira crise de identidade no setor. Se a máquina pode escrever o código, qual é o nosso lugar?

Os humanos devem ser o farol. Justamente porque os modelos sabem *tudo*, qualquer caminho parece igualmente plausível. Eles não conseguem priorizar e encadear informações para criar com **bom gosto**. É aqui que entra o nosso novo trabalho: como se estivéssemos explorando uma caverna escura, nós seguramos a lanterna, apontando para o caminho e decidindo onde o modelo deve focar a seguir.

Nosso papel evoluiu da escrita de código para a orquestração. Para muitos de nós, isso não é novidade, mesmo antes da codificação com IA: é a progressão natural da carreira de engenharia de software. Conforme você sobe na carreira, você põe menos a mão na massa e veste o chapéu de mentor cada vez mais. Você se torna responsável pelas entregas da sua equipe, para de ler cada linha de código para focar nas partes críticas, aprende a delegar e a lidar com as vantagens e desvantagens dessa escolha. Você se torna um tech lead.

## Todo mundo foi promovido a tech lead

![Em 2026 Todo Engenheiro Foi Promovido a Tech Lead](everyone-is-a-tech-lead.png "Parabéns pela sua promoção a Tech Lead. (P.S.: O RH confirmou que sua faixa salarial permanece inalterada.)")

Na trajetória tradicional de carreira, essa transição acontecia de forma gradual. Você passava anos nas trincheiras evoluindo de Júnior para Sênior antes de finalmente assumir o papel de lead, engineering manager ou arquiteto, focando na visão macro enquanto engenheiros mais juniores ficavam com os detalhes de implementação.

Agora, todo mundo tem sua equipe virtual de agentes de código. Portanto, até o engenheiro mais júnior precisa aprender a se tornar um bom tech lead. Não para liderar humanos, mas para guiar os agentes.

"Guiar para fazer o quê?" é a pergunta mais importante. Ainda estamos construindo programas, mas a forma como os descrevemos mudou.

Existem poucos livros que li na vida que considero verdadeiramente atemporais. Um deles é *Structure and Interpretation of Computer Programs* (aka SICP). Por que esse livro é relevante agora? Pela maneira como ele descreve programas e como nós, engenheiros, "conjuramos feitiços" para construí-los:

> *"Os processos computacionais são seres abstratos que habitam os computadores. À medida que evoluem, os processos manipulam outras coisas abstratas chamadas **dados**. A evolução de um processo é dirigida por um padrão de regras chamado **programa**. As pessoas criam programas para dirigir processos. Na prática, conjuramos os espíritos do computador com nossos feitiços."*
>
> *"... Os programas que usamos para conjurar processos são como os feitiços de um feiticeiro. Eles são cuidadosamente compostos a partir de expressões simbólicas em linguagens de programação arcanas e esotéricas que prescrevem as tarefas que queremos que nossos processos executem."*
>
> — Harold Abelson e Gerald Jay Sussman, *Structure and Interpretation of Computer Programs* (1985)

Ainda estamos no negócio de conjurar processos para instruir os computadores. Apenas a linguagem na qual criamos nossos feitiços mudou. Agora que escrevemos esses feitiços em linguagem natural em vez de expressões simbólicas arcanas, orquestrar agentes e garantir que eles estejam no caminho certo é a parte mais importante do trabalho. Isso implica escrever menos código e ler muito mais. Você também escreve mais especificações (prompts) do que nunca. Se você já foi tech lead antes da IA, tudo isso vai soar como o dia a dia de sempre.

Como ficar bom nisso? No mundo antigo, os bons tech leads e gerentes passavam muito tempo treinando e mentorando suas equipes para que desenvolvessem o mesmo pensamento crítico e bom gosto adquiridos ao longo de suas carreiras. Eles transferiam gradualmente o conhecimento de processos, comportamentos e técnicas, muitas vezes desenvolvidos à base de muitas cicatrizes de batalha. Aquele code review detalhista que acabei de fazer? Foi porque minha própria falta de atenção aos detalhes me fez derrubar a produção muitos anos atrás.

No mundo de hoje, além de aprender com nossos mentores e nossas próprias cicatrizes, precisamos observar atentamente o comportamento dos nossos agentes e aprender com os erros **deles**. Como esses erros se acumulam em velocidade vertiginosa, os agentes oferecem o ciclo de feedback comprimido definitivo: você assiste a dezenas de modos sutis de falha, condições de corrida e padrões alucinados se desenrolarem em uma única tarde, construindo discernimento arquitetural a dez vezes a velocidade dos ciclos tradicionais de lançamento.

Em essência, agora trabalhamos orquestrando o fluxo de informações, garantindo que os agentes sempre tenham o ambiente ideal para construir os resultados que desejamos.

Dada a taxa de evolução que estamos vendo, o que é ideal hoje muito provavelmente não será ideal amanhã. Portanto, precisamos refinar constantemente nossas ferramentas e processos para acompanhar o ritmo dessa evolução. Mais do que nunca, software é descartável. Nossos feitiços precisam se tornar rápidos e precisos, pois amanhã precisaremos de novos.

## Estruturando a informação para que os modelos possam agir

> *"Engenheiros de software mestres têm a capacidade de organizar programas para que possam ter certeza razoável de que os processos resultantes executarão as tarefas pretendidas. Eles conseguem visualizar o comportamento de seus sistemas com antecedência. Sabem estruturar programas para que problemas imprevistos não levem a consequências catastróficas e, quando surgem problemas, conseguem **depurar** seus programas. Sistemas computacionais bem projetados, assim como automóveis ou reatores nucleares bem projetados, são concebidos de forma modular, para que as peças possam ser construídas, substituídas e depuradas separadamente."*
>
> — Harold Abelson e Gerald Jay Sussman, *Structure and Interpretation of Computer Programs* (1985)

Se você pegar a citação acima e substituir a palavra "programas" por "agentes", todo o raciocínio continua válido. Precisamos aprender a estruturar nossos processos de codificação agêntica para que produzam sistemas computacionais bem projetados.

Uma das maneiras mais importantes de fornecer conhecimento a um agente é por meio de [Agent Skills]({{< ref "/posts/20260829-the-pragmatic-guide-to-agent-skills" >}}). Como os modelos se comportam de maneira muito semelhante aos humanos (o que não deveria surpreender, já que foram treinados com dados que nós produzimos), as mesmas técnicas que aplicávamos para aumentar a produtividade e a qualidade das entregas de nossas equipes são aplicáveis aos agentes. Isso cria um terreno fértil para a criação de skills que influenciarão o comportamento do modelo, inspiradas em suas próprias experiências e cicatrizes de batalha, bem como nas melhores práticas da indústria.

Uma dessas técnicas é quebrar tarefas complexas em partes menores. Um dos pilares do processo Ágil, que agora pode ser aplicado à codificação agêntica por meio de spec-driven development, planos de implementação e [subagentes]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}).

Pessoalmente, trato subagentes como os blocos fundamentais do fluxo de informação que estou desenhando. Isso é, em muitos aspectos, semelhante a lidar com concorrência em programas tradicionais, mas também desempenha um papel crucial no isolamento de contexto. Afinal, "atenção é tudo de que *eles* precisam". Ao quebrar tarefas em unidades ortogonais que podem ser processadas individualmente pelos subagentes, você garante que a atenção preciosa do agente fique focada exclusivamente naquela tarefa única, melhorando a qualidade e a aderência aos requisitos.

Sempre que você estiver criando fluxos de trabalho manuais ou construindo skills para a posteridade, considero estes pontos particularmente úteis:

1. **Papel de coordenador**: É útil manter a sessão principal livre garantindo que o coordenador sempre delegue a execução para subagentes, liberando a sessão principal o mais rápido possível. Dessa forma, você pode continuar interagindo com ela enquanto os subagentes trabalham, enviando comentários para orientá-los na direção certa em tempo real.
2. **Ortogonalidade de tarefas**: Divida tarefas em unidades ortogonais para minimizar dependências ou conflitos entre si. Primeiro, entenda e respeite dependências sequenciais. Depois, você pode paralelizar com a confiança de que os subagentes não sobrescreverão o trabalho uns dos outros nem tomarão premissas incompatíveis.
3. **Recriando fluxos comprovados**: Assim como no coaching e na mentoria, padrões que funcionam para equipes humanas e para a engenharia de software clássica também funcionam para agentes. Revisite constantemente a literatura em busca de padrões antigos e novos que você possa ensinar aos agentes.

Tenho construído gradualmente meu [catálogo de skills](https://skills.danicat.dev) ao longo dos últimos meses. Como as skills *apodrecem* rápido à medida que modelos e harnesses de agentes evoluem, é importante notar que isso não é algo do tipo "configure e esqueça". Eu as evoluo constantemente, a ponto de decidir adicionar um "freshness score" às minhas skills para me lembrar quais estão se tornando obsoletas, para que eu possa tentar dar uma vida nova a elas ou removê-las do catálogo.

Aqui estão alguns exemplos de skills que utilizo ativamente no meu dia a dia:

- **`double-diamond`**: fluxo de desenvolvimento de produto baseado na metodologia do duplo diamante das práticas ágeis, forçando a fase de descoberta antes da implementação.
- **`parallel`**: padrão multiúso de fan-out / map-reduce que quebra uma tarefa em várias threads e reconcilia os resultados no final.
- **`uno-reverse`**: revisão adversarial para combater o excesso de confiança do modelo e a tendência a complicar a arquitetura.
- **`less-is-more`**: filosofia de design que instrui o modelo a fazer uma coisa de cada vez, e bem feita, em vez de criar aplicações inchadas.
- **`intercom`**: ensina meus agentes a conversarem entre si quando estão trabalhando em projetos correlatos.
- **`show-me-the-money`**: minera o histórico da conversa em busca de funcionalidades que discuti com o modelo, mas esqueci completamente de implementar.
- **`abstract`**: generaliza exemplos e circunstâncias em conhecimento estruturado usando princípios filosóficos.

## Janelas quebradas no código agêntico

> *"Um processo computacional, em um computador operando corretamente, executa programas de forma precisa e acurada. Assim, como o aprendiz de feiticeiro, os programadores novatos precisam aprender a entender e a antecipar as consequências de sua conjuração. Mesmo pequenos erros (geralmente chamados de **bugs** ou **glitches**) nos programas podem ter consequências complexas e imprevistas. ... A programação no mundo real, no entanto, exige cuidado, experiência e sabedoria."*
>
> — Harold Abelson e Gerald Jay Sussman, *Structure and Interpretation of Computer Programs* (1985)

Se não anteciparmos as consequências do que conjuramos, um agente sem orientação multiplicará pequenas falhas estruturais mais rápido do que podemos nos dar conta. Esse problema não é novo; na verdade, sempre aconteceu mesmo com código escrito por humanos, mas não percebíamos tanto porque era um processo bem mais lento. Alguns tentavam impedir isso e eram sempre tachados de revisores de código pedantes, amados por poucos escolhidos, mas incompreendidos pela maioria.

Na psicologia, esse fenômeno é conhecido como a Teoria das Janelas Quebradas, baseada em um experimento de 1969. Pesquisadores deixaram um carro abandonado em um bairro nobre por uma semana. Nada aconteceu com ele. Na segunda semana, um dos pesquisadores usou uma marreta para quebrar uma das janelas. Em poucas horas, o carro foi completamente depenado.

O vidro quebrado foi o gatilho. Sinalizou que ninguém se importava, e assim a comunidade, coletivamente, também parou de se importar. Venho defendendo há anos que o mesmo acontece com o código.

Uma vez que surge uma janela quebrada no código (por exemplo: erros de digitação, nomes ruins, testes fracos ou abstrações equivocadas), ela ativa nosso mecanismo psicológico que deixa de se importar. E o problema escala com o tempo, fazendo com que pequenas falhas se transformem em bolas de neve.

No mundo agêntico, o problema é amplificado, já que o código existente enviesará o modelo ao gerar o novo. Se o código existente for ruim, a menos que devidamente instruído, o agente apenas ampliará seus problemas, pois aquilo é interpretado como um padrão aceitável a ser seguido.

Code reviews são a parte menos glamourosa do trabalho de engenharia, a ponto de muitos engenheiros nem considerarem isso como parte do trabalho. No entanto, é uma das mais importantes.

Isso **não** significa revisar cada linha de código.

Ninguém gosta de revisar milhares de linhas de código gerado por IA, inclusive eu. Bom gosto não é sobre verificar sintaxe ou caçar erros mecânicos — é para isso que servem as ferramentas determinísticas. Como escrevi em [Como Fazer Code Reviews na Era Agêntica]({{< ref "/posts/20260303-code-reviews-in-2026" >}}), se um compilador, linter ou teste automatizado pode pegar um problema, um ser humano não deveria estar olhando para aquilo.

Temos que combater IA com IA e automação: use barreiras determinísticas para eliminar o ruído mecânico, foque seu tempo de revisão nas fronteiras arquiteturais críticas e faça amostragem do restante. Como os tempos de ciclo hoje são ordens de magnitude mais rápidos do que na "programação artesanal", através da amostragem, mesmo que você não pegue um problema de imediato, terá rapidamente outra chance de capturá-lo. Portanto, direcione suas revisões para as partes realmente críticas e confie no processo para pegar o resto.

## Para onde vai o tempo livre prometido?

Estamos construindo software mais rápido do que nunca. Ao remover os gargalos de implementação, todas aquelas ideias que costumavam mofar na gaveta até serem substituídas por algo mais urgente de repente se tornaram viáveis. Agora podemos escrever protótipos em rápida sucessão para ver o que funciona e o que não funciona, muito antes de irem parar na pilha de projetos abandonados.

Mover-se mais rápido, contudo, não substitui a necessidade de construir modelos mentais. Nosso trabalho central é estruturar contexto, restrições e ciclos de feedback para que os agentes construam o que realmente pretendíamos. Se mudou algo, é que nosso trabalho ficou ainda mais exigente: o conhecimento técnico se tornou uma commodity, mas o bom gosto e o senso crítico continuam escassos.

Em última análise, ainda somos feiticeiros conjurando os espíritos dos computadores. Foram apenas nossos feitiços que cruzaram mais um nível de abstração.
