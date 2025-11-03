---
title: "Proatividade pode ser prejudicial? Um guia para customizar o Gemini CLI para o seu estilo de codificação"
date: 2025-07-14
author: "Daniela Petruzalek"
categories: ["Workflow & Best Practices"]
tags: ["gemini-cli", "vibe-coding", "tutorial"]
summary: "Aprenda a customizar o Gemini CLI para o seu estilo de codificação usando o GEMINI.md e instruções de sistema customizadas."
---
{{< translation-notice >}}

## Introdução

Muitos de vocês já devem estar familiarizados com o [Gemini CLI](https://cloud.google.com/gemini/docs/codeassist/gemini-cli?utm_campaign=CDR_0x72884f69_default_b432031389&utm_medium=external&utm_source=blog), mas caso não estejam, confiram o [blog de lançamento](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/) oficial para uma visão geral rápida.

Eu escrevi sobre como o incluí no meu fluxo de trabalho no meu post anterior ["Um Fluxo de Trabalho de Desenvolvimento Moderno para o Mundo Habilitado por IA"](../20250714-developer-workflow/), mas desta vez quero explorar algo um pouco diferente. Se você usou o CLI por um tempo, deve ter notado que ele é muito "proativo", pois muitas vezes infere os próximos passos com base até mesmo nos prompts mais ambíguos e imediatamente parte para a ação com base nessas inferências.

A intenção é dar à conversa um fluxo mais humano. Por exemplo, se você o instrui a escrever alguns testes, mas esquece de atualizar o README com as instruções para executar esses testes, pode-se dar um prompt de acompanhamento como uma pergunta do tipo "Você não deveria atualizar o README para adicionar as instruções para executar os testes?" e o CLI assumirá que a pergunta é retórica e prosseguirá com o comando para atualizar o arquivo README.

Embora esse nível de proatividade possa ser inofensivo ou até desejado para o caso médio, na minha experiência, ele tende a atrapalhar meu fluxo de trabalho com mais frequência do que o contrário. Um exemplo típico seria, depois de "vibe coding" de algum código, eu pergunto ao CLI para esclarecer "por que você adicionou este arquivo @x?", para ele apenas assumir que eu não quero que o arquivo exista e *proativamente* o exclua sem uma explicação. Geralmente fico profundamente irritada com esses tipos de interações porque na maioria das vezes estou fazendo uma pergunta genuína e esperando uma resposta.

Personalizar o CLI para se adequar ao seu próprio estilo de comunicação é importante para garantir que as ferramentas atendam às suas necessidades. Ninguém quer brigar com uma IA enquanto tenta fazer um trabalho produtivo, e essa tem sido uma razão frequente que as pessoas me dão para dispensar as ferramentas de IA em favor de seus métodos tradicionais com IDEs e autocompletar.

Para fazer o Gemini CLI se comportar de uma maneira que pareça mais produtiva para você, nas próximas duas seções, exploraremos duas maneiras de customizar as respostas do CLI: o arquivo GEMINI.md e as instruções do sistema.

## Customizando o Gemini CLI com GEMINI.md

O GEMINI.md é um arquivo usado para fornecer contexto adicional para o CLI. Assim que você inicia o CLI do seu terminal, ele procurará um arquivo GEMINI.md na pasta atual e em todas as suas pastas filhas. Esses arquivos podem ser usados para todos os tipos de coisas, mas uma estrutura de trabalho típica é usar o arquivo GEMINI.md raiz para explicar seu projeto ao CLI, desde seu propósito, até a organização de pastas e instruções-chave como compilação, teste, etc. É, na verdade, muito semelhante ao que um bom README.md seria, com a única diferença de que este arquivo é escrito com a IA em mente, então é mais "semelhante a um prompt".

O arquivo GEMINI.md de nível superior também é um bom lugar para dar ao CLI informações sobre como ele operará. Coisas como "faça um plano e peça confirmação antes de implementar qualquer tarefa" ou "sempre faça commit de passos intermediários usando git" podem ser adicionadas a este arquivo para garantir um fluxo de trabalho mais consistente.

Aqui está um bom exemplo de tal arquivo que define um processo para o CLI seguir (agradecimentos especiais a Ryan J. Salva por compartilhar isso):

{{< gist ryanjsalva 0a7f6782b8988e760b88f1635ea55f2e "GEMINI.md" >}}

Os arquivos GEMINI.md aninhados, por outro lado, podem ser úteis para explicar diferentes partes da base de código. Talvez você tenha um monorepo que tenha código de frontend e backend no mesmo lugar, então você pode ter um GEMINI.md customizado para cada um desses componentes. Ou você está escrevendo um programa em Go que tem muitos pacotes internos, onde cada um tem sua própria especificidade que você deseja que o CLI respeite. Seja qual for o seu caso de uso, ter vários arquivos GEMINI.md pode ajudá-lo a obter um controle refinado sobre o contexto para certas tarefas.

Nota: Da mesma forma que o Gemini CLI tem o GEMINI.md para arquivos de contexto, outras ferramentas de IA como [Claude](https://www.anthropic.com/product/claude) e [Jules](https://jules.google) têm seus próprios arquivos markdown (CLAUDE.md e AGENTS.md, respectivamente). Se você não estiver satisfeito com o nome GEMINI.md, ou quiser garantir que todas as ferramentas usem o mesmo arquivo, você sempre pode configurar o nome do arquivo de contexto usando a propriedade `contextFileName` em `settings.json`:

{{< github user="google-gemini" repo="gemini-cli" path="docs/cli/configuration.md" lang="markdown" start="38" end="43" >}}

## Mantendo os arquivos GEMINI.md

A queixa número um que ouvi sobre o(s) arquivo(s) GEMINI.md é que este é mais um arquivo para manter. A boa notícia é que você não precisa mantê-lo sozinho. Uma coisa que gosto de fazer depois de "parear" com o Gemini em uma tarefa de codificação por um tempo, especialmente se for uma daquelas sessões em que ocorreram muitos problemas ou mal-entendidos, é pedir ao Gemini para resumir os aprendizados para que no futuro os mesmos problemas não aconteçam, e aplicar esses aprendizados como instruções novas ou modificadas no arquivo GEMINI.md. Dessa forma, a tendência é que o modelo evolua com suas experiências e preferências pessoais, mesmo que leve alguns dias para ajustá-lo.

## A opção nuclear: sobrescrevendo as instruções do sistema

Embora um ou mais arquivos GEMINI.md seja a maneira recomendada de customizar o CLI, às vezes precisamos ir para a opção nuclear porque as [instruções do sistema](https://cloud.google.com/vertex-ai/generative/docs/concepts/system-instructions?utm_campaign=CDR_0x72884f69_default_b432031389&utm_medium=external&utm_source=blog) e nossos arquivos GEMINI.md estão se contradizendo. Como mencionei na introdução deste artigo, fico particularmente irritada quando o modelo tenta ser "muito proativo" e lê coisas em meus prompts que eu não pedi para ele fazer. Excluir um arquivo que eu não queria que fosse excluído, emendar um commit quando deveria ter criado um novo, "limpar" o repositório quando tenho trabalho não commitado em que passei muito tempo... essas são algumas das queimaduras da vida real que tive nas últimas semanas devido a essa proatividade.

Tentei brincar com o GEMINI.md para forçá-lo a responder muito literalmente aos meus prompts com pouco sucesso, o que eventualmente me jogou na toca do coelho das instruções do sistema. Minha hipótese era que havia algo com prioridade maior que o GEMINI.md que estava me atrapalhando. Felizmente, o Gemini CLI é de código aberto, então eu pude simplesmente ir ao código e encontrar o prompt para inspecioná-lo. Aqui está um trecho do que encontrei:

{{< github user="google-gemini" repo="gemini-cli" path="packages/core/src/core/prompts.ts" lang="markdown" start="40" end="53" >}}

O system prompt é enorme. Nesta prévia, estamos renderizando apenas a primeira dúzia de linhas, mas ele vai até o ponto de recomendar tecnologias para os casos de uso mais comuns e tudo mais (você pode ver o prompt completo no GitHub se clicar no link acima). Isso, é claro, faz muito sentido para um CLI que precisa satisfazer tantos casos de uso diferentes, mas pode não ser benéfico para o nosso próprio projeto, muito especializado.

Meu problema está na linha 49:

{{< github user="google-gemini" repo="gemini-cli" path="packages/core/src/core/prompts.ts" lang="markdown" start="49" end="49" >}}

Acredito que toda essa linha seja a fonte de 80% dos meus problemas, porque na maioria das vezes quero que uma pergunta seja apenas uma pergunta. Agora a questão é como nos livramos disso? Podemos enviar um PR para remover esta linha, mas talvez seja útil para outras pessoas. Eu poderia fazer um fork do projeto e criar meu próprio Daniela CLI, mas isso também não seria muito prático.

Felizmente, ao ler o código, me deparei com variáveis de ambiente não documentadas que podem ser muito úteis nesse processo: `GEMINI_SYSTEM_MD` e `GEMINI_WRITE_SYSTEM_MD`:

1. `GEMINI_SYSTEM_MD` permite que você sobrescreva o system prompt padrão com um arquivo markdown customizado. Uso:
    1. `GEMINI_SYSTEM_MD=SYSTEM.md`: lê o system prompt do arquivo `SYSTEM.md` customizado.
    2. `GEMINI_SYSTEM_MD=1`: lê o system prompt de `~/.gemini/system.md`
    2. `GEMINI_SYSTEM_MD=0` ou `GEMINI_SYSTEM_MD=""`: o system prompt é [construído durante o tempo de inicialização](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/core/prompts.ts) (Padrão)
2. `GEMINI_WRITE_SYSTEM_MD` permite que você escreva um system prompt para o disco no caminho fornecido. Uso:
    1. `GEMINI_WRITE_SYSTEM_MD=SYSTEM.md`: escreve o conteúdo do system prompt em um arquivo `system.md`. (a capitalização não é preservada)
    2. `GEMINI_WRITE_SYSTEM_MD=1`: escreve o conteúdo do system prompt em `~/.gemini/system.md` ou no local especificado por `GEMINI_SYSTEM_MD`, se definido.
    3. `GEMINI_WRITE_SYSTEM_MD=0` ou `GEMINI_WRITE_SYSTEM_MD=""`: desabilita a escrita em disco (Padrão).

Nota: existe uma [solicitação de feature](https://github.com/google-gemini/gemini-cli/issues/3923) aberta para documentar essas variáveis.

Para evitar escrever um novo system prompt do zero, defini `GEMINI_WRITE_SYSTEM_MD` como `SYSTEM.md` na minha pasta de projeto local e iniciei o Gemini CLI uma vez. Isso acionará a escrita do system prompt no disco. Observe que ele não respeita a capitalização, então, neste exemplo, ele escreverá como `system.md`, tudo em minúsculas.

```sh
$ export GEMINI_WRITE_SYSTEM_MD=SYSTEM.md
$ gemini
```
Você deve ver a tela de inicialização normal do gemini aparecer:

![Tela de inicialização do Gemini CLI](image-4.png)

Você pode sair do CLI digitando `/quit` ou digitando Ctrl+D ou Ctrl+C duas vezes. O arquivo `system.md` deve ter sido gravado no disco.

Nota: apenas caso você esteja curioso, a operação de escrita acontece no momento da inicialização, não no desligamento.

Verifique se o arquivo está realmente lá:

```sh
$ head -n 10 system.md
```
Aqui está a saída do mesmo comando no meu sistema:

![Saída do comando head em system.md](image-2.png)

Agora que temos uma cópia do prompt completo, estamos livres para editá-lo o quanto quisermos! Por exemplo, podemos nos livrar da problemática linha 49, ou até mesmo remover seções inteiras como as destinadas à construção de jogos (a menos que você esteja realmente construindo jogos, então, por favor, mantenha-as). Quando estiver satisfeito com seu novo prompt, você pode definir sua variável de ambiente `GEMINI_SYSTEM_MD` para seu arquivo customizado:

```sh
$ export GEMINI_SYSTEM_MD=system.md
$ gemini
```

Se você estiver usando um system prompt customizado, notará que no canto inferior esquerdo da tela há um ícone de óculos de sol vermelhos:

![Gemini CLI com system prompt customizado ativado](image-3.png)

Este é o sinal de que você não só é a pessoa mais legal da Terra, mas também que está usando um arquivo de instruções de sistema customizado.

## Conclusões

Neste artigo, exploramos como customizar sua experiência com o Gemini CLI usando tanto a maneira normal do GEMINI.md quanto a opção nuclear de sobrescrever o system prompt. Como qualquer técnica, use esse conhecimento com responsabilidade, mas espero que isso permita que você ajuste o Gemini para trabalhar de acordo com seu próprio estilo e tenha uma experiência de desenvolvimento melhor no geral.

Como de costume, eu adoraria ouvir seu feedback, especialmente em relação ao GEMINI.md ou system prompts que funcionaram para você.

---
**Nota:** Se você estiver curioso, pode conferir o arquivo `system.md` que uso para este blog neste [repositório](https://github.com/danicat/danicat.dev/blob/main/system.md).
