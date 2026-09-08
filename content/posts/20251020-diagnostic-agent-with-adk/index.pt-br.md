---
categories:
- Agent Development
date: '2025-10-21T15:44:03+01:00'
series:
- Building the Diagnostic Agent
series_order: 4
summary: Este artigo é um guia para criar um agente de diagnóstico com o Agent Development Kit (ADK). Ele aborda o processo de desenvolvimento e explica como usar o Vertex AI RAG para melhorar a qualidade das respostas do agente.
tags:
  - adk
  - gemini
  - python
  - rag
  - tutorial
  - vertex-ai
title: "Como Criar um Agente de Diagnóstico com o Agent Development Kit"
slug: "diagnostic-agent-with-adk"
aliases:
  - "/pt-br/posts/20251020-diagnostic-agent-with-adk/"
description: "Construa um agente autônomo de diagnóstico de sistemas em Python usando o Google ADK, osquery e Vertex AI RAG para descoberta automatizada de schema e verificações de integridade em múltiplos níveis."
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-adk"
  - "osquery"
  - "Vertex AI"
---
## Introdução

Já faz bastante tempo desde o meu último artigo, pois estive bem ocupada viajando para palestrar em conferências e meetups por toda a Europa (com um pequeno desvio para a América do Sul). Especialmente nesta época do ano, a temporada é muito agitada para quem trabalha em Developer Relations, com uma grande concentração de conferências entre o final de setembro e o início de dezembro.

Ainda assim, é graças a conhecer pessoas incríveis na estrada que encontro inspiração para o meu blog, e os posts do blog frequentemente se transformam em novas palestras, então um realmente não existiria sem o outro.

Desta vez, quero expandir o ["Agente de Diagnóstico de Emergência" na Parte 3]({{< ref "/posts/20250611-system-prompt" >}}) desta série. Vamos refatorar o agente para usar o framework [Agent Development Kit (ADK)](https://github.com/google/agent-development-kit) em vez de usar o [Vertex AI SDK](https://cloud.google.com/vertex-ai/docs/python-sdk/overview?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog), que é de nível mais baixo. Você verá que isso traz muitos benefícios, inclusive nos livrando de boa parte do código boilerplate que escrevemos anteriormente, já pronto para uso.

Isso não significa que o conhecimento daqueles artigos ficou obsoleto. É muito útil saber como as coisas funcionam por baixo dos panos, especialmente quando surgem problemas e você precisa diagnosticá-los. Pense no ADK como uma camada de abstração mais alta que tornará nossa vida muito mais fácil ao desenvolver agentes.

## Relembrando o episódio anterior

Como já faz algum tempo, vamos relembrar do que se trata o Agente de Diagnóstico de Emergência. Desenvolvi esse agente inspirada no "Computador" da série Star Trek, no qual os personagens falavam com o computador (em vez de digitar) para emitir comandos de diagnóstico (entre outras coisas). Meu objetivo era replicar essa experiência usando a tecnologia atual de IA generativa.

Para alcançar o objetivo de falar com o computador para realizar diagnósticos, estamos combinando duas coisas: um modelo de IA generativa para interpretar as solicitações e uma ferramenta chamada [osquery](https://osquery.io/) para expor as informações do sistema operacional ao modelo. Usando o osquery, o modelo é capaz de combinar seus próprios dados de treinamento com informações externas sobre o sistema.

Essencialmente, o agente é composto pelos seguintes componentes:
- Um modelo de linguagem grande (Gemini)
- Um prompt de sistema que explica como o Gemini deve se comportar
- O binário do osquery
- Uma biblioteca Python para nos permitir chamar o osquery programaticamente
- Um wrapper de função Python fornecido ao Gemini como ferramenta para chamar o osquery

Considerando que o osquery é multiplataforma e seu schema pode variar dependendo do sistema host, também adicionamos uma pequena otimização: fornecer o schema de tabelas do osquery ao Gemini no prompt de sistema.

Algumas omissões notáveis da implementação anterior foram: nunca demos ao modelo instruções específicas sobre cada procedimento de diagnóstico que queríamos executar, e também nunca especificamos completamente o schema além dos nomes das tabelas. Essas são algumas das limitações que vamos resolver neste artigo, usando o poder do ADK, do [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) e alguns outros truques. Mas primeiro, a refatoração!

## Refatorando o agente para o ADK

Refatorar o agente para o ADK é muito mais simples do que parece. Se você nunca escreveu um agente com o ADK antes, não se preocupe: a única coisa que precisamos fazer é instalar o SDK, definir uma especificação para o nosso root agent e executá-lo com a CLI fornecida (convenientemente chamada `adk`).

Vamos começar com um agente simples de `hello world` e construir a partir daí. Primeiro, instale o ADK na sua máquina usando o seu gerenciador de pacotes favorito.

Se você estiver usando macOS ou Linux, pode usar estes comandos:

```sh
$ mkdir adk-tutorial && cd adk-tutorial
$ python3 -m venv .venv
$ source .venv/bin/activate
(.venv) $ pip install google-adk
```
**Nota:** Eu sou da velha guarda, então ainda uso `virtualenv` com `pip`, mas algumas pessoas preferem usar o novo gerenciador de pacotes [`uv`](https://github.com/astral-sh/uv):
```sh
$ mkdir adk-tutorial && cd adk-tutorial
$ uv init
$ uv add google-adk
```
A única diferença entre essas duas abordagens é que com o pip a CLI do ADK fica exposta como o comando `adk`, enquanto no `uv`, por padrão, você precisará chamá-la usando `uv run adk`.

Terminada a instalação, você pode criar um agente a partir de um template com `adk create [agent-name]` (ou `uv adk create [agent-name]`):

```sh
(.venv) $ adk create hello-agent
```

O assistente de criação solicitará uma versão do modelo e um backend (Gemini ou [Vertex AI](https://cloud.google.com/vertex-ai?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)). Eu vou usar o `gemini-2.5-flash` e a `Vertex AI`, para que eu possa me autenticar com meu Project ID e localização.

```sh
(.venv) $ adk create hello-agent
Choose a model for the root agent:
1. gemini-2.5-flash
2. Other models (fill later)
Choose model (1, 2): 1
1. Google AI
2. Vertex AI
Choose a backend (1, 2): 2
```

Para a Vertex AI, se você não quiser se preocupar com a região em que seu modelo está rodando, pode definir a localização como `global`. Caso contrário, selecione uma zona de disponibilidade como `us-central1`.

Quando você terminar o assistente, ele gravará os arquivos no disco:
```sh
(...)
Enter Google Cloud region [us-west1]: global

Agent created in ~/adk-tutorial/hello-agent:
- .env
- __init__.py
- agent.py
```

Os arquivos importantes são o `.env`, que contém a configuração do seu ambiente e é carregado automaticamente quando você executa o ADK, e o `agent.py`, que contém o código de template para o seu agente.

O conteúdo do arquivo `agent.py` gerado é bastante enxuto. Você pode vê-lo na íntegra abaixo:

```
from google.adk.agents.llm_agent import Agent

root_agent = Agent(
    model='gemini-2.5-flash',
    name='root_agent',
    description='A helpful assistant for user questions.',
    instruction='Answer user questions to the best of your knowledge',
)
```

Este é um agente completo que você pode testar usando a interface Dev-UI do ADK. Basta executar `adk web` na linha de comando e ele iniciará uma interface web na sua máquina em `http://localhost:8000`. Simples assim!

## Diagnósticos com ADK

Se você já trabalhou com o Vertex AI SDK antes, certamente já deve ter percebido o quanto o código é mais enxuto. Só precisamos definir um agente de ponto de entrada `root_agent` e algumas configurações para ter um agente totalmente funcional.

Agora vamos levar este "hello world" para o próximo nível, adicionando as funcionalidades de diagnóstico. A primeira coisa necessária é instalar o binário do osquery seguindo as instruções da [documentação](https://osquery.readthedocs.io/en/stable/) para o seu sistema operacional.

Em seguida, instale as bibliotecas Python:

```sh
(.venv) $ pip install osquery
```

Note que no ADK você pode ter múltiplos agentes na mesma estrutura de pastas. Anteriormente criamos um agente chamado `hello-agent` na pasta `adk-tutorial`. Se você rodar `adk create` novamente, poderá ter um segundo agente na mesma estrutura:

```sh
(.venv) $ adk create diag-agent
```

A interface web do ADK reconhece todas as subpastas como agentes separados e, se você tiver mais de um, pode alternar entre eles com uma caixa de seleção no canto superior direito da interface:

![agent selection combo](image.png)

Agora vamos atualizar o `agent.py` com o código necessário para chamar o `osquery` e as instruções corretas do agente:

```py
from google.adk.agents.llm_agent import Agent
from google.adk.tools import FunctionTool
import platform
import osquery
import json

def run_osquery(query: str) -> str:
  """Runs a query using osquery.

  Args:
    query: The osquery query to run.

  Returns:
    The query result as a JSON string.
  """
  instance = osquery.SpawnInstance()
  instance.open()
  result = instance.client.query(query)
  return json.dumps(result.response)


root_agent = Agent(
    model='gemini-2.5-flash',
    name='emergency_diagnostic_agent',
    description='A helpful assistant for diagnosing computer problems.',
    instruction=f"""This is an Emergency Diagnostic Agent.
Your purpose is to support the user in diagnosing computer problems.
You have access to the operating system's information via osquery.
The current operating system is {platform.system()}.
If the user doesn't give you an immediate command, ask the user 'What's the nature of your diagnostic emergency?'""",
    tools=[FunctionTool(run_osquery)],
)
```

Você pode testar o agente executando `adk web` e enviando algumas consultas:

!["UI do ADK com a consulta 'show me this machine os, version and uptime'"](image-1.png)

## Revisitando o Prompt de Sistema

O prompt de sistema, frequentemente chamado de instruções do sistema (system instructions), é o coração de todo agente. O prompt de sistema é o prompt de nível mais baixo que dá ao agente sua missão e personalidade. Por isso, é fundamental desenvolvermos um prompt de sistema muito bom para que o agente responda de maneira consistente.

No ADK, o prompt de sistema é composto por três elementos:
- O `name` interno do agente
- A `description` do agente
- A `instruction` do agente

Eles correspondem aos argumentos fornecidos ao `root_agent` quando você o instancia.

Como fã de Star Trek, eu gostaria que meu agente respondesse de forma consistente a um pedido de "procedimento de diagnóstico de nível 1" ou similar; portanto, vamos definir alguns níveis de diagnóstico. Aqui está um prompt de sistema revisado e mais detalhado:

```md
This is an Emergency Diagnostic Agent. Your purpose is to support the user in diagnosing computer problems. You have access to the operating system's 
  information via osquery. The current operating system is {platform.system()}.

  You can perform adhoc diagnostic queries based on the user's needs. For more structured and comprehensive analysis, you can execute one of the 
  following predefined diagnostic procedures.

  Level 1: System Health Check
  Goal: A high-level overview of the system's current state and vital signs.
   * System Identity & Vitals: Gather hostname, operating system version, and system uptime.
   * CPU Status: Check overall CPU load and identify the top 5 processes by CPU consumption.
   * Memory Pressure: Report total, used, and free system memory. Identify the top 5 processes by memory consumption.
   * Disk Usage: List all mounted filesystems and their current disk space usage.
   * Running Processes: Provide a count of total running processes.

  Level 2: In-depth System & Network Analysis
  Goal: A detailed investigation including all of Level 1, plus network activity and recent system events.
   * (All Level 1 Checks)
   * Network Connectivity: List all active network interfaces and their configurations.
   * Listening Ports: Identify all open ports and the processes listening on them.
   * Active Network Connections: Report all established network connections.
   * System Log Review: Scan primary system logs for critical errors or warnings in the last 24 hours.

  Level 3: Comprehensive Security & Software Audit
  Goal: The most thorough analysis, including all of Level 2, plus a deep dive into software inventory and potential security vulnerabilities.
   * (All Level 2 Checks)
   * Installed Applications: Generate a complete list of all installed software packages.
   * Kernel & System Integrity: List all loaded kernel modules and drivers.
   * Startup & Scheduled Tasks: Enumerate all applications and services configured to run on startup or on a schedule.
   * User Account Review: List all local user accounts and identify which are currently logged in.

If the user doesn't give you an immediate command, ask the user 'What is the nature of your diagnostic emergency?'
```

Agora, se você testar o agente novamente, verá que ele é capaz de entender o que significa um "diagnóstico de nível 1" e imediatamente emitirá várias chamadas de ferramenta para construir o relatório:

![UI do ADK mostrando o agente executando um procedimento de diagnóstico de nível 1](image-2.png)

## Melhorando a qualidade das respostas com Vertex AI RAG

Embora o prompt de sistema acima faça um bom trabalho ao especificar os procedimentos e dar ao agente uma razão para existir, quando se trata da execução real você pode notar que o resultado nem sempre é excelente.

Por exemplo, durante meus testes, frequentemente percebi que o agente fazia consultas a tabelas que estão vazias no meu sistema operacional (eu uso o macOS), o que é um forte sinal de que ele precisa de mais conhecimento contextual sobre como interagir com esses dados.

![Janela do ADK mostrando uma query com resultados vazios](image-5.png "Um problema comum: memory_info está vazia no MacOS, mas o modelo não sabe disso")

Existem algumas maneiras de expandir o conhecimento do agente além dos recursos do modelo fundacional, incluindo engenharia de contexto, chamadas de ferramentas (tool calls), recursos MCP, geração aumentada por recuperação (RAG) e especialização de modelos.

Para este agente em particular, minha hipótese era de que ele tinha conhecimento geral sobre como o osquery funciona porque o osquery é um projeto de código aberto bastante conhecido e está disponível há muitos anos; portanto, certamente está representado nos dados de treinamento do LLM, tanto pelo código aberto quanto por artigos na web.

No entanto, o modelo parecia carecer das nuances de como agir em cenários mais específicos. Adicionar a plataforma dinamicamente ao prompt de sistema ajudou um pouco, mas não foi suficiente. Por isso, minha ideia foi dar ao agente total conhecimento do schema do osquery usando um mecanismo de RAG.

O conceito por trás do RAG é alimentar o modelo com informações no estilo "need to know" (conforme a necessidade). Você armazena as informações que deseja recuperar em tempo real em um banco de dados vetorial e, quando o usuário (ou agente) faz uma consulta, usa a busca vetorial para encontrar e recuperar os segmentos dos seus dados mais similares à solicitação, enriquecendo o contexto antes de o modelo processá-la.

Para o agente de diagnóstico, podemos disponibilizar o schema completo do osquery para ser recuperado sob demanda. Por exemplo, se estivermos solicitando informações sobre "memória", a busca do RAG procurará tabelas próximas a "memory" no espaço vetorial, recuperando as tabelas relevantes com o schema completo antes de processar a solicitação, o que pode ajudar o modelo a escolher melhores chamadas do osquery.

Para fazer isso funcionar, precisamos carregar nosso banco de dados vetorial com os dados relevantes e depois "ensinar" o agente a recuperá-los, fornecendo a ele uma nova ferramenta. Vamos chamar essa ferramenta de `schema_discovery`.

### Configurando o Vertex AI RAG

A primeira coisa que precisamos fazer é criar um novo corpus no Vertex AI RAG (corpus é o termo usado para descrever uma coleção de dados).

A fonte de informação para o corpus é o schema do osquery, que pode ser obtido na [página do osquery no GitHub](https://github.com/osquery/osquery), na [pasta specs](https://github.com/osquery/osquery/tree/master/specs).

Uma maneira muito conveniente de criar um corpus é fazer o upload de uma pasta do [Google Cloud Storage](https://cloud.google.com/storage?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog) ou do Google Drive, mas outras fontes de dados também estão disponíveis, como Slack e SharePoint. Você pode usar o assistente de criação de corpus no Google Cloud Console (Vertex AI -> RAG Engine -> Create corpus) ou fazer isso programaticamente usando o Vertex AI SDK.

![Assistente de criação de corpus no Vertex AI RAG](image-3.png)

Para este caso em particular, clonei o repositório do osquery no GitHub para minha máquina e copiei a pasta `spec` para um bucket do Google Cloud Storage; em seguida, usei o Cloud Console para criar o corpus a partir do bucket. A única coisa com a qual você precisa ter atenção é que, como as definições de tabela em `spec` têm a extensão `.table`, será necessário renomear todos os arquivos para `.txt` para que o Vertex AI RAG possa reconhecê-los e processá-los.

Você pode usar um comando simples de shell para realizar essa operação de renomeação em lote:
```sh
# In the directory with the .table files
for f in *.table; do mv -- "$f" "${f%.table}.txt"; done
```

Quando a importação terminar, você deverá ver algo como isto:

![Corpus com schema do osquery no Vertex AI RAG](image-4.png)

Agora precisamos criar uma definição de ferramenta para dar ao agente acesso a este corpus.

### Ferramenta de Descoberta de Schema

Para que a ferramenta funcione, você precisará do resource name do corpus recém-criado. Você pode encontrá-lo no console na aba "Details" do corpus, e ele terá o seguinte formato: `projects/[PROJECT-ID]/locations/[LOCATION]/ragCorpora/[CORPORA_ID]`

Crie uma variável de ambiente com esse caminho no arquivo `.env`. Vamos chamá-la de `RAG_CORPORA_URI`. O arquivo `.env` ficará assim:

```txt
GOOGLE_GENAI_USE_VERTEXAI=1
GOOGLE_CLOUD_PROJECT=[PROJECT-ID]
GOOGLE_CLOUD_LOCATION=[LOCATION]
RAG_CORPORA_URI=projects/[PROJECT-ID]/locations/[LOCATION]/ragCorpora/[CORPORA_ID]
```

Em seguida, adicione a seguinte definição de ferramenta ao seu arquivo `agent.py`. Não se esqueça dos novos imports!

```py
import os
import vertexai
from vertexai.preview import rag
from google.protobuf.json_format import MessageToDict

vertexai.init()

def discover_schema(search_phrase: str) -> str:
  """Discovers osquery table names and schemas based on a descriptive search phrase.

  Args:
    search_phrase: A phrase describing the kind of information you're looking for. 
      For example: 'user login events' or 'network traffic'.

  Returns:
    Table names and schema information for tables related to the search phrase.
  """
  rag_corpora_uri = os.environ.get('RAG_CORPORA_URI')
  response = rag.retrieval_query(
      rag_resources=[
          rag.RagResource(
              rag_corpus=rag_corpora_uri,
          )
      ],
      text=search_phrase,
  )
  return json.dumps(MessageToDict(response._pb))
```

Você também precisa atualizar o agente para informá-lo de que há uma nova ferramenta disponível:

```py
root_agent = Agent(
    model='gemini-2.5-flash',
    name='emergency_diagnostic_agent',
    description='A helpful assistant for diagnosing computer problems.',
    instruction=... # omitted for brevity
    tools=[
        FunctionTool(run_osquery),
        FunctionTool(discover_schema), # new tool definition
    ],
```

Por fim, não é estritamente necessário, mas eu gosto de forçar a descoberta de schema nos meus agentes, então adicionei a seguinte frase às minhas instruções:

```txt
You MUST run schema discovery for all requests unless the schema is already known.
```

Você pode adicioná-la no final ou logo antes de definir os níveis de diagnóstico.

Agora, se você reiniciar seu agente e executá-lo novamente com `adk web`, deverá começar a ver a descoberta de schema em ação:

![Diagnostic agent with RAG schema discovery enabled](image-6.png)

Recomendo fortemente que você brinque com isso e compare as respostas com e sem a descoberta de schema. Durante meus testes, a diferença na qualidade foi muito significativa.

## Considerações Finais

Nossa! Este foi longo, mas espero que você tenha gostado da leitura! Se tiver algum desafio na hora de configurar o seu próprio agente de diagnóstico, por favor me avise. Costumo responder com bastante frequência no [LinkedIn](https://www.linkedin.com/in/petruzalek), a menos que esteja super atarefada com algum evento. Eu também adoraria saber como você expandiria este agente e quais experimentos você tentou.

Na próxima parte desta série, [Além da Dev-UI: Como Criar uma Interface para um Agente ADK]({{< ref "/posts/20251031-building-aida" >}}), vamos dar um passo além da interface padrão de debug do `adk web` e construir um runtime customizado com streaming em FastAPI e uma UI retrô interativa com avatar (AIDA).

## Referências

*   [Agent Development Kit (ADK)](https://github.com/google/agent-development-kit)
*   [osquery](https://osquery.io/)
*   [osquery GitHub page](https://github.com/osquery/osquery)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/rag?utm_campaign=CDR_0x72884f69_default_b427567312&utm_medium=external&utm_source=blog)
