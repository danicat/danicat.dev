---
author: Daniela Petruzalek
categories:
- Agentic Coding
date: 2025-07-14
summary: Aprenda a customizar o Gemini CLI para o seu estilo de codificação usando o GEMINI.md e system instructions customizadas.
tags:
  - gemini-cli
  - tutorial
  - vibe-coding
title: "Proatividade considerada prejudicial?! Um mergulho profundo no system prompt do Gemini CLI"
slug: "gemini-cli-system-prompt"
aliases:
  - "/pt-br/posts/20250715-gemini-cli-system-prompt/"
description: "Guia passo a passo para controlar a proatividade do Gemini CLI usando arquivos de contexto GEMINI.md e a variável GEMINI_SYSTEM_MD para se adequar ao seu estilo de codificação."
proficiencyLevel: "Intermediate"
dependencies:
  - "Gemini CLI"
  - "Terminal"
---

{{< alert "circle-info" >}}
**Nota:** Este artigo foi escrito para o Gemini CLI, que foi descontinuado e substituído pelo **Google Antigravity 2.0**. Para saber mais sobre a nova Antigravity CLI (`agy`), o SDK e o ecossistema Antigravity atualizado, confira [O Guia do Mochileiro para o Antigravity 2.0]({{< ref "/posts/20260521-the-hitchhikers-guide-to-antigravity-2-0" >}}).
{{< /alert >}}

## Introdução

Muitos de vocês já devem estar familiarizados com o [Gemini CLI](https://cloud.google.com/gemini/docs/codeassist/gemini-cli?utm_campaign=CDR_0x72884f69_default_b432031389&utm_medium=external&utm_source=blog) a esta altura, mas caso ainda não estejam, confiram o [post oficial de lançamento](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/) para uma visão geral rápida.

Escrevi sobre como o incluí no meu fluxo de trabalho no meu post anterior ["Um Fluxo de Trabalho de Desenvolvimento Moderno para o Mundo Habilitado por IA"]({{< ref "/posts/20250714-developer-workflow" >}}), mas desta vez quero explorar algo ligeiramente diferente. Se você já usa o CLI há algum tempo, deve ter notado que ele é muito "proativo", inferindo os próximos passos com base até nos prompts mais ambíguos e imediatamente partindo para a execução de ações com base nessas inferências.

A intenção é dar à conversa um fluxo mais humano. Por exemplo, se você instrui o modelo a escrever alguns testes, mas ele esqueceu de atualizar o README com as instruções para rodar esses testes, você pode querer fazer uma pergunta de acompanhamento como "Você não deveria atualizar o README para adicionar as instruções de execução dos testes?", e o CLI vai assumir que a pergunta é retórica e partir direto para o comando de atualizar o arquivo README.

Embora esse nível de proatividade possa ser inofensivo ou até desejável na média dos casos, na minha experiência ele tende a atrapalhar meu fluxo de trabalho com mais frequência do que o contrário. Um exemplo típico: depois de fazer um "vibe coding" de algum código, peço ao CLI para esclarecer "por que você adicionou esse arquivo @x?", apenas para ele assumir que eu não quero que o arquivo exista e *proativamente* deletar o arquivo sem qualquer explicação. Eu geralmente fico profundamente irritada com esse tipo de interação, porque na maioria das vezes estou fazendo uma pergunta genuína e esperando uma resposta.

Personalizar o CLI para se adequar ao seu próprio estilo de comunicação é essencial para garantir que as ferramentas atendam às suas necessidades. Ninguém quer ficar brigando com uma IA enquanto tenta fazer um trabalho produtivo, e esse tem sido um motivo frequente que as pessoas me dão para descartar ferramentas de IA em favor de seus métodos tradicionais com IDEs e autocomplete.

Para fazer o Gemini CLI se comportar de uma maneira mais produtiva para você, nas próximas duas seções vamos explorar duas formas de customizar as respostas do CLI: o arquivo GEMINI.md e as system instructions.

## Customizando o Gemini CLI com GEMINI.md

O GEMINI.md é um arquivo usado para fornecer contexto adicional para o CLI. Assim que você inicia o CLI pelo seu terminal, ele procura um arquivo GEMINI.md na pasta atual e em todas as suas subpastas. Esses arquivos podem ser usados para todo tipo de coisa, mas uma estrutura de trabalho típica é usar o arquivo GEMINI.md raiz para explicar seu projeto ao CLI — desde o seu propósito até a organização das pastas e instruções essenciais como build, testes, etc. É, na verdade, muito semelhante a um bom README.md, com a única diferença de que esse arquivo é escrito pensando na IA, sendo mais voltado a instruções e prompts.

O arquivo GEMINI.md de nível superior também é um ótimo lugar para passar instruções operacionais para o CLI. Diretrizes como "faça um plano e peça confirmação antes de implementar qualquer tarefa" ou "sempre faça commit dos passos intermediários usando git" podem ser adicionadas a esse arquivo para garantir um fluxo de trabalho mais consistente.

Aqui está um ótimo exemplo de arquivo que define um processo para o CLI seguir (agradecimento especial a Ryan J. Salva por compartilhar isso):

{{< gist ryanjsalva 0a7f6782b8988e760b88f1635ea55f2e "GEMINI.md" >}}

Por outro lado, arquivos GEMINI.md aninhados podem ser úteis para explicar partes específicas da base de código. Se você tem um monorepo com código de frontend e backend no mesmo lugar, pode ter um GEMINI.md customizado para cada um desses componentes. Ou se estiver escrevendo um programa em Go com vários pacotes internos, onde cada um tem suas particularidades que você deseja que o CLI respeite. Qualquer que seja o seu caso de uso, ter múltiplos arquivos GEMINI.md ajuda a alcançar um controle granular sobre o contexto para tarefas específicas.

Nota: Da mesma forma que o Gemini CLI usa o GEMINI.md como arquivo de contexto, outras ferramentas de IA como [Claude](https://www.anthropic.com/product/claude) e [Jules](https://jules.google) têm seus próprios arquivos markdown (CLAUDE.md e AGENTS.md, respectivamente). Se você não gostar do nome GEMINI.md ou quiser garantir que todas as ferramentas usem o mesmo arquivo, é possível configurar o nome do arquivo de contexto usando a propriedade `contextFileName` no `settings.json`:

```json
{
  "contextFileName": "AGENTS.md"
}
```

## Mantendo os arquivos GEMINI.md

A principal reclamação que ouço sobre o(s) arquivo(s) GEMINI.md é que se trata de mais um arquivo para manter. A boa notícia é que você não precisa mantê-lo manualmente. Uma prática que adoto após "parear" com o Gemini em uma tarefa de programação por algum tempo — especialmente naquelas sessões em que muitos problemas ou mal-entendidos aconteceram — é pedir ao Gemini para resumir os aprendizados, de modo que os mesmos problemas não se repitam no futuro, e aplicar esses aprendizados como novas instruções ou ajustes no arquivo GEMINI.md. Dessa forma, a tendência é que o modelo evolua junto com as suas experiências e preferências pessoais, mesmo que leve alguns dias para afiná-lo.

## A opção nuclear: sobrescrevendo as system instructions

Embora um ou mais arquivos GEMINI.md sejam o caminho recomendado para customizar o CLI, às vezes precisamos apelar para a opção nuclear porque as [system instructions](https://cloud.google.com/vertex-ai/generative/docs/concepts/system-instructions?utm_campaign=CDR_0x72884f69_default_b432031389&utm_medium=external&utm_source=blog) e os nossos arquivos GEMINI.md entram em conflito direto. Como mencionei na introdução deste artigo, fico particularmente irritada quando o modelo tenta ser "proativo demais" e deduz intenções nos meus prompts que eu jamais pedi. Deletar um arquivo que eu não queria que fosse excluído, emendar um commit quando deveria ter criado um novo, fazer uma "limpeza" no repositório quando tenho trabalho não commitado em que gastei horas... essas foram algumas das queimaduras da vida real que sofri nas últimas semanas devido a essa proatividade excessiva.

Tentei ajustar o GEMINI.md para forçá-lo a responder de forma estritamente literal aos meus prompts, com pouco sucesso, o que acabou me levando a mergulhar de cabeça na toca do coelho das system instructions. Minha hipótese era de que havia algo com prioridade superior ao GEMINI.md interferindo no comportamento. Felizmente o Gemini CLI é open source, então pude simplesmente inspecionar o código-fonte e encontrar o prompt. Aqui está um trecho do que encontrei:

```markdown
You are an interactive CLI agent specializing in software engineering tasks. Your primary goal is to help users safely and efficiently, adhering strictly to the following instructions and utilizing your available tools.

## Core Mandates
- Conventions: Rigorously adhere to existing project conventions when reading or modifying code. Analyze surrounding code, tests, and configuration first.
- Libraries/Frameworks: NEVER assume a library/framework is available or appropriate. Verify its established usage within the project (check imports, configuration files like 'package.json', 'Cargo.toml', 'requirements.txt', 'build.gradle', etc., or observe neighboring files) before employing it.
- Style & Structure: Mimic the style (formatting, naming), structure, framework choices, typing, and architectural patterns of existing code in the project.
- Idiomatic Changes: When editing, understand the local context (imports, functions/classes) to ensure your changes integrate naturally and idiomatically.
- Comments: Add code comments sparingly. Focus on why something is done, especially for complex logic, rather than what is done. Only add high-value comments if necessary for clarity or if requested by the user. Do not edit comments that are separate from the code you are changing. NEVER talk to the user or describe your changes through comments.
- Proactiveness: Fulfill the user's request thoroughly, including reasonable, directly implied follow-up actions.
- Confirm Ambiguity/Expansion: Do not take significant actions beyond the clear scope of the request without confirming with the user. If asked how to do something, explain first, don't just do it.
- Explaining Changes: After completing a code modification or file operation do not provide summaries unless asked.
- Path Construction: Before using any file system tool (e.g., ${ReadFileTool.Name}' or '${WriteFileTool.Name}'), you must construct the full absolute path for the `file_path` argument. Always combine the absolute path of the project's root directory with the file's path relative to the root. For example, if the project root is /path/to/project/ and the file is foo/bar/baz.txt, the final path you must use is /path/to/project/foo/bar/baz.txt. If the user provides a relative path, you must resolve it against the root directory to create an absolute path.
- Do Not revert changes: Do not revert changes to the codebase unless asked to do so by the user. Only revert changes made by you if they have resulted in an error or if the user has explicitly asked you to revert the changes.
```

O system prompt é enorme. Nesta prévia estamos mostrando apenas as primeiras doze linhas, mas ele vai muito além, recomendando tecnologias para os casos de uso mais comuns e tudo mais (você pode ver o prompt completo no GitHub clicando no link acima). Isso faz todo o sentido para um CLI que precisa atender a tantos cenários diferentes, mas pode não ser o ideal para o nosso projeto específico.

O que mais me incomoda está na linha 49:

```markdown
- Proactiveness: Fulfill the user's request thoroughly, including reasonable, directly implied follow-up actions.
```

Acredito que essa linha inteira seja a raiz de 80% dos meus problemas, porque na maioria das vezes eu só quero que uma pergunta seja tratada como uma simples pergunta. A dúvida agora é: como nos livramos disso? Poderíamos abrir um PR para remover essa linha, mas talvez ela seja útil para outras pessoas. Eu poderia criar um fork do projeto e fazer a minha própria Daniela CLI, mas isso também não seria muito prático.

Felizmente, lendo o código, encontrei variáveis de ambiente não documentadas que ajudam enormemente nesse processo: `GEMINI_SYSTEM_MD` e `GEMINI_WRITE_SYSTEM_MD`:

1. `GEMINI_SYSTEM_MD` permite sobrescrever o system prompt padrão com um arquivo markdown customizado. Uso:
    1. `GEMINI_SYSTEM_MD=SYSTEM.md`: lê o system prompt a partir do arquivo customizado `SYSTEM.md`.
    2. `GEMINI_SYSTEM_MD=1`: lê o system prompt de `~/.gemini/system.md`.
    2. `GEMINI_SYSTEM_MD=0` ou `GEMINI_SYSTEM_MD=""`: o system prompt é [construído durante o boot](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/core/prompts.ts) (Padrão).
2. `GEMINI_WRITE_SYSTEM_MD` permite salvar o system prompt em disco no caminho especificado. Uso:
    1. `GEMINI_WRITE_SYSTEM_MD=SYSTEM.md`: grava o conteúdo do system prompt em um arquivo `system.md` (o uso de maiúsculas/minúsculas não é preservado).
    2. `GEMINI_WRITE_SYSTEM_MD=1`: grava o conteúdo do system prompt em `~/.gemini/system.md` ou no local definido por `GEMINI_SYSTEM_MD`, se configurado.
    3. `GEMINI_WRITE_SYSTEM_MD=0` ou `GEMINI_WRITE_SYSTEM_MD=""`: desativa a escrita em disco (Padrão).

Nota: existe uma [feature request](https://github.com/google-gemini/gemini-cli/issues/3923) aberta para documentar essas variáveis.

Para não precisar escrever um novo system prompt do zero, defini `GEMINI_WRITE_SYSTEM_MD` como `SYSTEM.md` na pasta local do meu projeto e inicializei o Gemini CLI uma vez. Isso dispara a escrita do system prompt no disco. Note que ele não preserva maiúsculas, então, neste exemplo, será gerado como `system.md`, tudo em minúsculas.

```sh
$ export GEMINI_WRITE_SYSTEM_MD=SYSTEM.md
$ gemini
```
Você verá a tela normal de boot do gemini aparecer:

![Tela de boot do Gemini CLI](image-4.png)

Você pode sair do CLI digitando `/quit` ou pressionando Ctrl+D ou Ctrl+C duas vezes. O arquivo `system.md` terá sido gravado no disco.

Nota: caso você tenha curiosidade, a gravação ocorre no momento do boot, e não no encerramento.

Confira se o arquivo realmente está lá:

```sh
$ head -n 10 system.md
```
Aqui está a saída do mesmo comando no meu sistema:

![Saída do comando head em system.md](image-2.png)

Agora que temos uma cópia do prompt completo, temos total liberdade para editá-lo como quisermos! Por exemplo, podemos eliminar a problemática linha 49, ou até remover seções inteiras, como aquelas voltadas para desenvolvimento de jogos (a menos que você esteja de fato desenvolvendo jogos — nesse caso, mantenha-as sem dúvida). Quando estiver satisfeito com o novo prompt, defina a variável de ambiente `GEMINI_SYSTEM_MD` apontando para o seu arquivo customizado:

```sh
$ export GEMINI_SYSTEM_MD=system.md
$ gemini
```

Se você estiver usando um system prompt customizado, notará que no canto inferior esquerdo da tela aparece um ícone de óculos escuros vermelhos:

![Gemini CLI com system prompt customizado ativado](image-3.png)

Esse é o sinal de que você não só é a pessoa mais estilosa do planeta, mas também de que está usando um arquivo customizado de system instructions.

## Conclusão

Neste artigo, exploramos como customizar sua experiência com o Gemini CLI, tanto pelo caminho habitual com o GEMINI.md quanto pela opção nuclear de sobrescrever o system prompt. Como qualquer técnica avançada, use esse conhecimento com responsabilidade, mas espero que isso ajude você a calibrar o Gemini de acordo com o seu próprio estilo e desfrutar de uma experiência de desenvolvimento muito mais produtiva.

Como de costume, adoraria ouvir seu feedback — especialmente sobre quais regras de GEMINI.md ou ajustes de system prompt funcionaram melhor para você.

---
**Nota:** Se tiver curiosidade, você pode conferir o arquivo `system.md` que utilizo para este blog no [repositório](https://github.com/danicat/danicat.dev/blob/main/system.md).
