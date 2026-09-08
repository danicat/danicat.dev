---
categories:
- Applied GenAI
date: '2025-07-02T00:00:00+01:00'
summary: Como desenhar infraestrutura no Google Cloud usando linguagem natural com o Gemini Cloud Assist
tags:
  - gemini
  - google-cloud
  - tutorial
title: "Do Prompt à Infraestrutura com o Gemini Cloud Assist"
slug: "gemini-cloud-assist"
aliases:
  - "/pt-br/posts/20250702-gemini-cloud-assist/"
description: "Projete infraestrutura no Google Cloud usando linguagem natural com o Gemini Cloud Assist e o Application Design Center. Gere e exporte módulos Terraform prontos."
proficiencyLevel: "Beginner"
dependencies:
  - "Google Cloud Platform"
  - "Terraform"
---
## Introdução

Hoje vamos fazer um pequeno desvio do nosso conteúdo habitual sobre agentes de IA para falar sobre um produto que explorei recentemente durante a minha participação no I/O Connect Berlin 2025, na semana passada.

O evento reuniu mais de 1.000 pessoas desenvolvedoras de toda a Europa, incluindo membros das comunidades do Google (Google Developer Groups) e experts da comunidade. Foi também o meu primeiro evento oficial do Google desde que entrei para a equipe de DevRel em abril, o que tornou tudo ainda mais especial para mim — e é por isso que não tivemos post no blog na semana passada!

Fui responsável por uma demonstração chamada “Design and Deploy”, que apresentou a combinação de dois produtos: o [Application Design Center (ADC)](https://cloud.google.com/application-design-center/docs/overview?utm_campaign=CDR_0x72884f69_awareness_b428663487&utm_medium=external&utm_source=blog) e o [Gemini Cloud Assist (GCA)](https://cloud.google.com/products/gemini/cloud-assist?utm_campaign=CDR_0x72884f69_awareness_b428663487&utm_medium=external&utm_source=blog). A demo foi tão bem recebida que decidi trazer o conteúdo para o blog também, dando a oportunidade de experimentar essa tecnologia a quem não pôde estar lá presencialmente.

O Application Design Center é um produto criado para ajudar arquitetos e desenvolvedores a desenhar a infraestrutura de suas aplicações. Na camada visual, ele oferece uma interface intuitiva na qual você define os componentes da infraestrutura graficamente; por baixo dos panos, porém, tudo o que aparece na tela é representado como um módulo Terraform, permitindo aproveitar todos os benefícios de [Infraestrutura como Código (IaC)](https://en.wikipedia.org/wiki/Infrastructure_as_code).

Um aviso importante é que o ADC está atualmente em [public preview](https://cloud.google.com/products?e=48754805&hl=en#product-launch-stages&utm_campaign=CDR_0x72884f69_awareness_b428663487&utm_medium=external&utm_source=blog). Isso significa que o produto evolui diariamente e pode, eventualmente, quebrar a compatibilidade com versões anteriores. Ele também conta com algumas arestas que precisam ser lapidadas antes da disponibilidade geral (GA), as quais vou destacar adiante.

Já o Gemini Cloud Assist (também em public preview) é o nome oficial do suporte do Gemini no Google Cloud. Por isso, o GCA não é um produto isolado, mas sim uma camada conectiva que permite interagir com qualquer recurso do GCP em linguagem natural, trazendo todas as vantagens da experiência moderna de chat baseada em modelos de linguagem (LLMs).

Vamos ver como combinar essas duas tecnologias para desenhar rapidamente a infraestrutura de uma aplicação.


## Como iniciar uma sessão de design de aplicação

Você sempre pode abrir o Application Design Center manualmente pelo console do Google Cloud, mas onde estaria a graça nisso? O jeito mais prático de acionar o ADC para um novo desenho é abrir o painel do Gemini a partir de qualquer página. Aqui, por exemplo, estou na página inicial (Welcome) do meu projeto:

![Tela inicial no console do Google Cloud](images/image001.png "Tela inicial no console do Google Cloud")


Ao clicar no botão de “estrela” no canto direito da barra de pesquisa, você abre o painel do Gemini Cloud Assist:

![Visão ampliada do botão do Gemini](images/image002.png "Visão ampliada do botão do Gemini")

A interface exibida será esta:

![Tela inicial do Google Cloud Assist](images/image003.png "Tela inicial do Google Cloud Assist")


Nesse painel você pode interagir diretamente com o Gemini. Digite algo como “crie uma aplicação que faz X” e inclua quantos detalhes arquiteturais desejar. Por exemplo, vamos criar uma aplicação que gera imagens de gatos. Aqui está o prompt:

> Crie uma aplicação que gere imagens de gatos com o Gemini e as armazene em um banco de dados Cloud SQL. Os usuários podem solicitar novas imagens usando um serviço de geração e visualizar as imagens criadas por meio de um serviço de imagens. Ambos os serviços devem ser expostos através de um serviço de frontend e um balanceador de carga global.

Após enviar o prompt, o Gemini processará a requisição por alguns instantes e apresentará um resultado como este:

![Resposta do Gemini com diagrama de arquitetura](images/image004.png "Resposta do Gemini com diagrama de arquitetura")

A visualização integrada dá uma boa ideia inicial, mas podemos interagir muito melhor com o desenho clicando no botão “Edit app design”. Isso abrirá a arquitetura em modo expandido para refinamento. (Nota: o restante deste artigo assume que o botão “Edit app design” abre a janela de Preview. Caso isso não aconteça no seu ambiente, consulte as observações ao final do texto).

É assim que o desenho aparece na janela de “Preview”:

![Janela de Preview do Gemini Cloud Assist](images/image005.png "Janela de Preview do Gemini Cloud Assist")

Caso queira ajustar as convenções de nomenclatura ou os detalhes dos componentes gerados, basta clicar sobre qualquer componente para abrir o painel de configuração. Aqui abri os detalhes do meu `frontend-service`:

![Visão do painel de detalhes do componente](images/image006.png "Visão do painel de detalhes do componente")

Repare que a tela também exibe a imagem de contêiner instanciada pelo Cloud Run, que por padrão vem como um contêiner “hello”. Isso acontece porque o Gemini Cloud Assist não sabe qual imagem você pretende rodar, mas você pode informar esse valor para que ele faça a substituição.

Vale reforçar esse ponto para alinhar expectativas: a ferramenta não escreve o código da aplicação por você, ela apenas projeta a infraestrutura necessária para executá-la. Para programar os serviços de frontend e backend, por exemplo, você continuará usando outras ferramentas como a Gemini CLI ou sua IDE de preferência, publicando os artefatos no seu container registry para que o Cloud Run possa acessá-los.

Na janela de Preview é possível editar componentes existentes, mas não adicionar novos blocos manualmente. Se quiser iterar no design, basta pedir as alterações ao Gemini em linguagem natural. Veja este prompt de continuidade:

> Adicione um serviço de streaming que capture eventos para cada imagem de gato gerada. Do outro lado do stream, inclua um serviço consumidor que atualizará uma página estática hospedada no GCS, adicionando as fotos mais recentes a um feed.

E esta foi a resposta do Gemini:

![Resposta do Gemini ao prompt complementar](images/image007.png "Resposta do Gemini ao prompt complementar")

A janela de Preview é atualizada com o novo design, destacando adições (em verde), modificações (em azul) e remoções (em vermelho):

![Alterações propostas no diagrama](images/image008.png "Alterações propostas no diagrama")

Na parte inferior da tela, você tem a opção de aceitar ou rejeitar a sugestão. Mas antes disso, é uma boa oportunidade para inspecionar o código Terraform gerado nos bastidores. Para visualizar o código e comparar as mudanças, clique em “View diff”:

Isso abrirá a janela de Code Diff com ambas as versões lado a lado:

![Janela de diff comparando o código Terraform antes e depois](images/image009.png "Janela de diff comparando o código Terraform antes e depois")

Como você pode ver, cada caixa no diagrama é mapeada para um módulo Terraform específico. Rolando até o fim da página, os módulos adicionados recentemente aparecem destacados em verde.

Se a implementação estiver satisfatória, você pode aceitar a proposta ou rejeitá-la e pedir melhorias ao Gemini. Eu aceitei a sugestão, mas notei algo curioso no módulo “database-secrets” e resolvi questionar o Gemini:

Prompt: “por que você adicionou um database secret se o banco Cloud SQL está usando autenticação IAM?”

Pois é, no fim das contas não era realmente necessário:

![Resposta do Gemini à pergunta sobre autenticação IAM](images/image010.png "Resposta do Gemini à pergunta sobre autenticação IAM")

Na janela de Preview:

![Proposta do Gemini para remover o segredo do banco de dados](images/image011.png "Proposta do Gemini para remover o segredo do banco de dados")

Esse é um lembrete importante: por mais avançada que a IA seja, ela ainda não nos isenta de avaliar e tomar decisões. No final das contas, a IA ainda vai continuar lá, mas é o nosso emprego que está em jogo — então não se esqueça de validar tudo. 🙂

Falando em validações, outro detalhe que me chamou a atenção foi o tipo de instância sugerido pelo Gemini para o Cloud SQL: `db-perf-optimized-N-8`. Vamos tentar outro prompt para otimizar os custos, já que essa máquina é nitidamente exagerada para um protótipo inicial:

> Torne a arquitetura mais econômica (Make it cost effective)

![Resposta do Gemini sugerindo um balanceador regional e trocando Postgres por MySQL](images/image012.png "Resposta do Gemini sugerindo um balanceador regional e trocando Postgres por MySQL")

Hummm… essa resposta me fez pensar. Entendo o ponto sobre balanceador regional vs. global, mas não me convenci de por que ele considera o MySQL mais econômico que o PostgreSQL. Minha preocupação principal era com o tipo de máquina, não com a tecnologia do banco em si.

A explicação do Gemini também não conta a história toda. Ao inspecionar o diff com atenção, vi que ele de fato alterou o tipo de máquina (pelo atributo `tier`), só esqueceu de mencionar na mensagem:

![Diff do Terraform mostrando que o Gemini também alterou o tipo de máquina (tier)](images/image013.png "Diff do Terraform mostrando que o Gemini também alterou o tipo de máquina (tier)")

Não fiquei totalmente satisfeita com isso, então perguntei o porquê:

> Por que você considera o MySQL mais econômico que o PostgreSQL?

![Perguntando ao Gemini por que ele considera o MySQL mais econômico que o Postgres](images/image014.png "Perguntando ao Gemini por que ele considera o MySQL mais econômico que o Postgres")

A resposta alegou que o MySQL seria mais econômico que o Postgres devido a:
1. Diferenças de licenciamento
2. Consumo de recursos
3. Preço do serviço gerenciado

Infelizmente não posso concordar com essa resposta. Quanto ao item 1, ambos têm licenças de código aberto, então não são tão diferentes assim. O item 2 pode até ter algum fundo de verdade, mas eu ainda precisaria de um benchmark adequado. Já o item 3 está errado, porque o Cloud SQL para Postgres e para MySQL possui exatamente o mesmo modelo de preços no GCP. Mais um ponto para os humanos! Vamos reverter a alteração:

> Reverta a alteração de postgres para mysql, mas mantenha o tipo de máquina menor.

Inspeção final: fiquei satisfeita com o Cloud SQL rodando Postgres em um tier menor de banco de dados, mas também notei outra alteração relevante que habilita o recurso de scale-to-zero no Cloud Run:

![Diff do Terraform mostrando que o Cloud Run foi configurado com scale-to-zero (min_instance_count = 0)](images/image015.png "Diff do Terraform mostrando que o Cloud Run foi configurado com scale-to-zero (min_instance_count = 0)")

Essa alteração faz todo o sentido, mas também não foi mencionada no diálogo. Esse é outro lembrete de “confie, mas verifique” tudo o que as ferramentas de IA sugerem. Não queremos surpresas rodando em produção!

## Baixando os arquivos Terraform

Quando estiver tudo pronto com o seu desenho, basta clicar no botão “&lt;&gt; Get Code” no canto superior direito da tela. A ferramenta vai empacotar todo o código Terraform subjacente em um arquivo `.zip` para download na sua máquina local.

Infelizmente, no momento em que escrevo este artigo, o Application Design Center não oferece suporte a integrações com sistemas de versionamento de código como GitHub, GitLab, Google Source, Bitbucket e outros. A única forma de extrair o código da ferramenta é por meio desse download em arquivo zip.

Para quem usa contas corporativas com uma hierarquia de organização completa, é possível pegar esse design e fazer o deploy usando o AppHub. Mas se você estiver usando uma conta pessoal, infelizmente este é o limite do que a ferramenta consegue fazer por você.


## Observações sobre a interface do App Design Center

O botão “Edit app design” tem comportamentos diferentes dependendo de como o seu console do Google Cloud está configurado. Se você estiver testando esse prompt a partir de uma conta pessoal e essa conta não estiver associada a uma organização, ele abrirá uma janela de Preview onde é possível visualizar o design e baixar o código Terraform correspondente, mas você não terá acesso à interface completa do App Design Center.

Para usar a interface completa, você precisa fazer parte de uma organização, pois a configuração do App Design Center exige um tipo especial de pasta configurada, conhecida como pasta “app design center enabled”. Não há como criar pastas em contas sem organização e, dentro de uma organização, essa pasta precisa ser configurada pelo administrador de nuvem.

Infelizmente, isso significa que contas de usuário que não pertencem a nenhuma organização ficarão efetivamente impedidas de acessar o conjunto completo de recursos do ADC, pelo menos por enquanto.

Você ainda poderá usar o Gemini para ajudar a prototipar a arquitetura da sua aplicação como mostrei neste artigo, mas não poderá salvar seu progresso na interface web da nuvem e precisará baixar os arquivos Terraform para a sua máquina local e fazer o deploy usando a sua própria instalação do Terraform.

## Conclusões e próximos passos

Cada novo produto de IA lançado me deixa empolgada com a ideia de viver aquele momento "Tony Stark", em que você projeta seu software apenas com comandos de voz. Ainda não chegamos lá, mas com o Gemini Cloud Assist estamos fazendo um bom progresso, já que agora podemos usar linguagem natural para especificar os componentes de infraestrutura por nós.

Ainda existem algumas arestas a serem aparadas, tanto na interface quanto nas sugestões do Gemini, mas já fico aliviada de não precisar escrever código Terraform manualmente do zero para cada nova aplicação que começo a desenvolver.

Este é claramente um artigo que deve ter data de validade, pois devemos ver essas ferramentas evoluírem muito rápido nos próximos meses. Para se manter atualizado, você sempre pode conferir a página do produto [Application Design Center](https://cloud.google.com/application-design-center/docs/overview?utm_campaign=CDR_0x72884f69_awareness_b428663487&utm_medium=external&utm_source=blog), e com certeza farei o meu melhor para trazer as novas funcionalidades e melhorias interessantes aqui no blog também.

Como sugestão, recomendo experimentar alguns prompts criativos como “make it cost effective” (torne mais econômico), “make it highly available” (garanta alta disponibilidade), “explain why x instead of y” (explique por que x em vez de y), “replace x with y” (substitua x por y), “explain x to me like I’m 5” (me explique x como se eu tivesse 5 anos), e assim por diante.

O que você achou? Achou essa ferramenta empolgante ou assustadora? Descobriu algum prompt interessante? Deixe seus comentários abaixo!
