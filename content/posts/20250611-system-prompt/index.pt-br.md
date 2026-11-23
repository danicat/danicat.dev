---
categories:
- Agent Development
date: '2025-06-11T00:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 3
summary: Este artigo explora os conceitos de instruções de sistema, histórico de sessão e ferramentas para criar um assistente de diagnóstico mais inteligente.
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: "Prompt Audacioso: Um Guia Prático para Instruções de Sistema e Ferramentas"
slug: "system-instructions-agent-tools"
aliases:
  - "/pt-br/posts/20250611-system-prompt/"
description: "Aprenda a combinar instruções de sistema, histórico de chat e ferramentas osquery em Python com a Vertex AI para construir um assistente de diagnóstico autônomo."
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-cloud-aiplatform"
  - "osquery"
  - "LangChain"
status: "deprecated"
superseded_by:
  - "/posts/20261006-building-agentic-backends-with-genkit-go"
  - "/posts/20261120-ai-agents-from-scratch-with-google-adk"
status_note: "O template de prompt Role/Tools/Context/Task, a ferramenta de depuração auto-ativável e a recuperação de schema com PRAGMA table_info continuam úteis, mas o código legado de vertexai com LangChain e o ID de modelo preview estão obsoletos."
---
## Introdução

Neste guia, vamos aprender mais sobre prompts de sistema (system prompts) e ferramentas de agentes (agent tools) para construir uma experiência nova e aprimorada para o nosso agente de diagnóstico. Vamos trabalhar com o [SDK da Vertex AI para Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog), LangChain, Gemini e [osquery](https://www.osquery.io/).

Preciso admitir que a [versão inicial do agente de diagnóstico]({{< ref "/posts/20250531-diagnostic-agent" >}}) não estava muito pronta para a "Enterprise" (com o perdão do trocadilho). Não tínhamos muita visibilidade do que ele estava fazendo por baixo dos panos (ele estava mesmo rodando alguma query?), ele não se lembrava das coisas discutidas na mesma "sessão" e, de tempos em tempos, também ignorava completamente nossos comandos.

Isso está longe da experiência que desejamos de um agente de verdade. O agente de diagnóstico ideal precisa ser capaz de lembrar dos seus erros e executar instruções de forma consistente, por exemplo, aprendendo que certas colunas não estão disponíveis e contornando isso. Além disso, podemos mesmo confiar que ele está fazendo o que diz estar fazendo? Deveríamos poder ver as queries a qualquer momento para ter certeza de que a informação que ele está retornando é correta e atualizada.

Com esses objetivos em mente, vamos colocar a mão na massa e começar a construir o nosso ~~Holograma Médico~~ Agente de Diagnóstico de Emergência!

## Preparando o ambiente

Da última vez escrevemos o código em um notebook Jupyter pela conveniência, mas desta vez vamos escrever um programa Python normal. O mesmo código também funcionaria no Jupyter com alterações mínimas, mas estamos fazendo isso para podermos usar o agente de diagnóstico com uma interface de chat adequada.

Para qualquer projeto Python, eu sempre recomendo começar com um ambiente virtual limpo para manter as dependências isoladas:

```
$ mkdir -p ~/projects/diagnostic-agent
$ cd ~/projects/diagnostic-agent
$ python3 -m venv venv
$ source venv/bin/activate
$ pip install --upgrade google-cloud-aiplatform[agent_engines,langchain]
```

Aqui está a versão inicial de `main.py` que reproduz o agente do artigo anterior:

```py
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown
import os

PROJECT_ID = os.environ.get("GCP_PROJECT")
LOCATION = os.environ.get("GCP_REGION", "us-central1")
STAGING_BUCKET = os.environ.get("STAGING_BUCKET_URI")

vertexai.init(
    project=PROJECT_ID,
    location=LOCATION,
    staging_bucket=STAGING_BUCKET
)

MODEL = os.environ.get("GEMINI_MODEL", "gemini-2.0-flash")

instance = osquery.SpawnInstance()

def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()  # This may raise an exception

    result = instance.client.query(query)
    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()  # This may raise an exception
    
    response = instance.client.query("select name from sqlite_temp_master").response
    tables = [ t["name"] for t in response ]
    return f"""
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fulfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
    """

def main():
    agent = agent_engines.LangchainAgent(
        model = MODEL,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
        ],
    )
    
    console = Console()

    print("Welcome to the Emergency Diagnostic Agent\n")
    print("What is the nature of your diagnostic emergency?")

    while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query)
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)

    print("Goodbye!")

if __name__ == "__main__":
    main()
```

Você pode executar o agente com `python main.py`:

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

What is the nature of your diagnostic emergency?
>> 
```

Há duas pequenas mudanças em comparação com o código original: primeiro, agora temos um loop principal que mantém o agente em execução até que a pessoa usuária digite "exit" ou "quit". Isso cria a nossa interface de chat.

Segundo, ajustamos o prompt de sistema para melhorar a consistência do agente. Agora o chamamos de "Emergency Diagnostic Agent" (Agente de Diagnóstico de Emergência). Esse nome não só funciona como um ótimo [easter egg de Star Trek](https://en.wikipedia.org/wiki/The_Doctor_(Star_Trek:_Voyager%29), mas, mais importante ainda, estabelece um tom de urgência que, com base em [pesquisas recentes](https://arxiv.org/pdf/2307.11760), pode incentivá-lo a atender aos nossos pedidos com mais diligência. (Confira também esta [entrevista recente](https://www.reddit.com/r/singularity/comments/1kv7hm2/sergey_brin_we_dont_circulate_this_too_much_in/))

Não vamos ameaçar nosso pobre Agente de Diagnóstico de Emergência (e posso garantir que nenhum agente foi ferido durante a produção deste texto), mas chamá-lo de agente de "Emergência" deve dar o tom para que ele tente atender aos nossos pedidos da melhor maneira possível. Na versão anterior do prompt de sistema, tive casos em que o agente se recusava a fazer uma tarefa porque "achava" que não era capaz de fazê-la ou não sabia quais tabelas consultar.

Claro que apenas chamá-lo de agente de emergência não basta para garantir o comportamento desejado, então adicionamos mais algumas instruções para orientar o comportamento do modelo, como veremos a seguir.

## Instruções de sistema

Instruções de sistema, também chamadas de prompt de sistema (system prompt), são um conjunto de instruções que guiam o comportamento do LLM durante toda a conversa. As instruções de sistema são especiais porque têm prioridade maior sobre as interações comuns do chat. Uma forma de imaginar isso é como se as instruções de sistema fossem sempre repetidas junto com o prompt que você está enviando para o modelo.

Não existe um consenso forte na literatura sobre qual deve ser o formato exato de um prompt de sistema, mas temos alguns padrões testados em batalha que vêm surgindo no uso do dia a dia. Por exemplo, um que é praticamente consenso é dedicar logo o início do prompt de sistema para atribuir um papel (role) ao agente, para que ele tenha consciência do seu propósito e consiga produzir respostas mais coerentes.

Para este agente em particular, optei por incluir as seguintes seções no meu prompt de sistema: papel (role), ferramentas (tools), contexto (context) e tarefa (task). Essa estrutura funcionou bem durante a minha fase de testes, mas não se apegue demais a ela: experimente com seus prompts e veja se consegue resultados melhores. Experimentação é a chave para alcançar bons resultados com LLMs.

Agora vamos dar uma olhada em cada seção do prompt.

### Prompt de sistema: papel (Role)

Um papel nada mais é do que a razão de existir do agente. Pode ser algo simples como "você é um engenheiro de software" ou "você é um agente de diagnóstico", mas também pode ser um pouco mais elaborado, incluindo uma descrição detalhada, regras de comportamento, restrições e outros pontos.

Para um modelo de linguagem grande treinado em todo tipo de dado, o papel ajuda a definir o tom do domínio de conhecimento que ele precisará acessar para responder às suas perguntas. Em outras palavras, ele dá significado semântico às suas consultas... Imagine a pergunta "o que são cookies?", por exemplo. Estamos falando de biscoitos comestíveis ou de cookies de navegador? Se o papel do agente estiver indefinido, essa pergunta é completamente ambígua, mas assim que definimos o papel como algo técnico (por exemplo, "você é uma pessoa engenheira de software"), a ambiguidade desaparece.

Para este agente, o papel é descrito assim:

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
```

Além da definição direta ("você é o agente de diagnóstico de emergência"), adicionamos uma descrição mais longa para definir o tom do comportamento do modelo e, com sorte, influenciá-lo a levar nossos pedidos "a sério", já que, como mencionei antes, a iteração anterior desse agente tinha o mau hábito de recusar pedidos.

### Prompt de sistema: ferramentas (Tools)

Tools é a seção que explica ao agente suas capacidades de interagir com sistemas externos além do seu modelo central. Ferramentas podem ser de vários tipos, mas a forma mais comum de fornecer uma ferramenta ao agente é por meio de chamadas de função (function calls).

Os agentes podem usar ferramentas para recuperar informações, executar tarefas e manipular dados. O SDK da Vertex AI para Python oferece suporte tanto para funções fornecidas pelo usuário quanto para ferramentas integradas, como Busca do Google e execução de código. Você também pode usar extensões mantidas pela comunidade por meio da interface do Model Context Protocol (MCP).

Para o nosso agente, precisamos dizer a ele que pode chamar o _osquery_:

```
Tools:
  - you can call osquery using the call_osquery function.
```

### Prompt de sistema: contexto (Context)

Em seguida temos o contexto, que informa ao agente sobre o ambiente em que ele opera. Eu uso essa seção para apontar e corrigir explicitamente comportamentos indesejados que iterações anteriores do agente costumavam ter. Por exemplo, percebi bem cedo no desenvolvimento que o agente tentava "adivinhar" quais tabelas estavam disponíveis e enviava queries às cegas, resultando em uma alta taxa de erros. Adicionar a lista de tabelas ao contexto ajudou a mitigar esse problema.

Semelhante é a tendência do agente de tentar adivinhar os nomes das colunas em uma tabela, em vez de tentar descobrir os nomes primeiro. Nesse caso específico, resisti à tentação de instruir o agente a usar sempre `SELECT *`, porque isso é uma má prática (recupera mais dados do que você precisa), e em vez disso "ensinei" ele a descobrir um schema usando a instrução `PRAGMA`.

Dessa forma, o agente ainda vai cometer erros ao adivinhar nomes de colunas, mas tem uma maneira de corrigir o curso sem intervenção humana.

A seção de contexto revisada do prompt de sistema é mostrada abaixo.

```
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
```

Note que `tables` é uma variável que contém todas as tabelas que descobrimos no osquery antes de iniciar o modelo.

### Prompt de sistema: tarefa (Task)

Por fim, a tarefa. Essa seção é usada para descrever como o agente deve interpretar seus pedidos e executá-los. As pessoas geralmente usam essa seção para listar os passos necessários para concluir a tarefa em questão.

No nosso caso específico, estamos usando essa seção para traçar o plano em linhas gerais, mas também adicionando algumas diretivas condicionais:

```
Task:
  - Create a plan for which tables to query to fulfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

O passo "confirm the plan with the user before executing" (confirmar o plano com o usuário antes de executar) é interessante porque nos mostra como o agente está pensando sobre o processo, mas pode ficar um pouco irritante depois de interagir com o agente por algum tempo. Sempre podemos pedir em um prompt para o agente nos contar o plano, então a inclusão desse passo é totalmente opcional.

Eu inicialmente pensei nesse passo como uma forma de depurar o agente, mas na próxima seção vamos explorar uma maneira diferente de fazer isso.

Com a combinação dessas quatro seções, temos o prompt de sistema completo. Esse prompt revisado produziu resultados mais consistentes durante os meus testes na preparação deste artigo. Ele também tem a vantagem de ser "amigável para humanos", então fica mais fácil adaptá-lo quando novas regras são introduzidas.

Aqui está a visão completa desse prompt de sistema:

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fulfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

Fazendo um parêntese, acredito que ainda temos muitas oportunidades para melhorá-lo, e uma das minhas áreas atuais de interesse é como alcançar um prompt de sistema autoaprimorável (self-improving system prompt). Isso poderia ser feito pedindo, ao final de uma sessão, para o modelo resumir seus aprendizados em um novo prompt de sistema para sua próxima iteração. O prompt poderia ser armazenado em um banco de dados e carregado na sessão seguinte. Isso, é claro, levanta preocupações com a degradação do prompt de sistema ou, pior ainda, ataques usando prompt injection, portanto não é tão trivial quanto parece. Ainda assim, é um exercício divertido e talvez eu escreva sobre ele em um futuro próximo.

## Habilitando o modo de depuração (Debug Mode)

Outra preocupação sobre o design original é a falta de observabilidade sobre o que o agente está fazendo por baixo dos panos. Há duas abordagens diferentes que podemos aplicar aqui, uma um pouco mais dolorosa que a outra: 1) espiar os "pensamentos" do LLM e tentar encontrar as chamadas de ferramentas no meio deles (muito doloroso), ou; 2) adicionar alguma funcionalidade de depuração à própria função para que ela imprima a informação que queremos em tempo de execução (a solução mais simples geralmente é a certa).

Preciso confessar que passei uma quantidade nada saudável de tempo na opção 1 antes de perceber que podia fazer a opção 2. Se você realmente quiser seguir o caminho do raciocínio do LLM, pode fazer isso por meio de uma configuração chamada [return_intermediate_steps](https://api.python.langchain.com/en/latest/agents/langchain.agents.agent.AgentExecutor.html#langchain.agents.agent.AgentExecutor.return_intermediate_steps). Devo dizer que isso é muito interessante do ponto de vista de aprendizado, mas depois de passar algumas horas tentando descobrir o formato da saída (dica: [não é bem JSON](https://github.com/langchain-ai/langchain/issues/10099)), decidi que fazer o parsing disso realmente não valia o trabalho.

Então, como funciona a estratégia simples? Adicionamos uma flag de debug e uma ferramenta para ligar e desligar essa flag. Esse truque surpreendentemente simples na verdade abre todo um novo mundo de potencial: estamos dando ao agente a oportunidade de modificar o seu próprio comportamento!

A implementação do modo de debug é composta por uma variável global e uma função para defini-la:

```py
debug = False

def set_debug_mode(debug_mode: bool):
    """Toggle debug mode. Call this function to enable or disable debug mode.
    
    Args:
        debug_mode (bool): True to enable debug mode, False to disable it.


    Returns:
        None
    """
    global debug
    debug = debug_mode
```

Também precisamos mencioná-la no prompt de sistema:

```
...
   Tools:
    - you can call osquery using the call_osquery function.
    - you can enable or disable the debug mode using the set_debug_mode function.
    Context:
...
```

E adicionar a função à lista de ferramentas na instanciação do agente:

```py
   agent = agent_engines.LangchainAgent(
        model = model,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode,
        ],
    )

```

Por fim, precisamos adaptar `call_osquery` para usar a nova flag `debug`:

```py
def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()

    if debug:
        print("Executing query: ", query)

    result = instance.client.query(query)
    if debug:
        print("Query result: ", {
            "status": result.status.message if result.status else None, 
            "response": result.response if result.response else None
        })

    return result
```

Com todas essas mudanças no lugar, vamos dar uma olhada em como o agente chama o _osquery_ usando a flag de debug recém-implementada:

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

What is the nature of your diagnostic emergency?
>> run a level 1 diagnostic procedure in debug mode
Executing query:  SELECT * FROM system_info
Query result:  {'status': 'OK', 'response': [{...}]}
Executing query:  SELECT pid, name, user, cpu_percent FROM processes ORDER BY cpu_percent DESC LIMIT 10
Query result:  {'status': 'no such column: user', 'response': None}
Executing query:  SELECT pid, name, user, resident_size FROM processes ORDER BY resident_size DESC LIMIT 10
Query result:  {'status': 'no such column: user', 'response': None}
Executing query:  PRAGMA table_info(processes)
Query result:  {'status': 'OK', 'response': [{'cid': '0', 'dflt_value': '', 'name': 'pid', 'notnull': '1', 'pk': '1', 'type': 'BIGINT'}, ...]}
(...)

System Information:                                                                                                                                                     

 • Hostname: workstation.localdomain                                                                                                                               
 • CPU Type: arm64e                                                                                                                                                     
 • Physical Memory: 51539607552 bytes                                                                                                                                   

Top 5 Processes by CPU Usage:                                                                                                                                          
                                                  
  PID     Name                         CPU Usage  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 
  1127    mediaanalysisd               95627517   
  43062   mediaanalysisd-access        66441942   
  54099   Google Chrome                3005046    
  54115   Google Chrome Helper (GPU)   2092500    
  81270   Electron                     1688335    

Top 5 Processes by Memory Usage (Resident Size):                                                                                                                       
                                                                   
  PID     Name                              Resident Size (Bytes)  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 
  43062   mediaanalysisd-access             3933536256             
  54099   Google Chrome                     1313669120             
  59194   Code Helper (Plugin)              1109508096             
  59025   Code Helper (Renderer)            915456000              
  19681   Google Chrome Helper (Renderer)   736329728                         
                                                                   
>> 
```

Observe que o comando dado foi "run a level 1 diagnostic procedure **in debug mode**" ("execute um procedimento de diagnóstico de nível 1 em modo debug"), o que mostra uma capacidade interessante do agente: invocação de múltiplas ferramentas (multi-tool invocation). Se julgar necessário, ele consegue invocar não apenas a mesma ferramenta várias vezes, mas também ferramentas diferentes ao mesmo tempo. Assim, não foi preciso ativar o modo debug antes de pedir o relatório: o agente conseguiu fazer tudo de uma vez só.

Note também como o agente falhou inicialmente ao solicitar uma coluna `user`, mas depois usou a instrução `PRAGMA` para descobrir o schema correto e tentar novamente a query com sucesso. Essa é uma demonstração perfeita da capacidade do agente de se recuperar de erros graças ao nosso prompt de sistema aprimorado.

## Preservando o histórico do chat

Nossa tarefa final hoje é garantir que o agente se lembre do que estamos falando para que possamos fazer perguntas de acompanhamento e investigar ainda mais o sistema seguindo uma linha de raciocínio coerente.

No [artigo anterior]({{< ref "/posts/20250605-vertex-ai-sdk-python" >}}) exploramos como os LLMs são stateless (sem estado) e como precisamos continuar "lembrando" a eles o estado atual da conversa usando "turnos". Felizmente, com o LangChain não precisamos fazer isso manualmente e podemos contar com um recurso chamado [chat history](https://python.langchain.com/api_reference/core/chat_history.html).

A beleza do chat history é que qualquer coisa que implemente [BaseChatMessageHistory](https://python.langchain.com/api_reference/core/chat_history/langchain_core.chat_history.BaseChatMessageHistory.html#langchain_core.chat_history.BaseChatMessageHistory) pode ser usada aqui, o que nos permite utilizar todo tipo de armazenamento de dados, inclusive criar o nosso próprio. Por exemplo, na documentação oficial da Vertex AI você encontra exemplos de uso do [Firebase, Bigtable e Spanner](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog#chat-history).

Não precisamos de um banco de dados completo no momento, então vamos ficar com o `InMemoryChatMessageHistory`, que, como o nome sugere, armazena tudo em memória.

Aqui está uma implementação típica, que tecnicamente suporta múltiplas sessões usando o dicionário `chats_by_session_id` para consulta (código obtido da [documentação do langchain](https://python.langchain.com/docs/versions/migrating_memory/chat_history/#chatmessagehistory)):

```py
chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history
```

E aqui está nossa nova função `main` instanciando o agente com o histórico de chat habilitado:

```py
import uuid

def main():
    session_id = uuid.uuid4()
    agent = agent_engines.LangchainAgent(
        model = model,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode,
        ],
        chat_history=get_chat_history,
    )
```

Um aviso rápido para que você não cometa o mesmo erro que eu: o argumento `chat_history` espera um tipo `Callable`, então você não deve invocar a função ali, mas sim passar a própria função. O LangChain usa um padrão factory aqui; ele invoca a função fornecida (`get_chat_history`) sob demanda com um `session_id` para obter ou criar o objeto de histórico correto. Esse design é o que permite ao agente gerenciar múltiplas conversas separadas simultaneamente.

A assinatura da função pode incluir um ou dois argumentos. Se tiver um argumento, assume-se que é um `session_id`, e se tiver dois argumentos, eles são interpretados como `user_id` e `conversation_id`. Mais informações sobre isso podem ser encontradas na documentação do [RunnableWithMessageHistory](https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.history.RunnableWithMessageHistory.html).

A última peça do quebra-cabeça é passar o `session_id` para o executor do modelo. Isso é feito por meio do argumento `config`, como mostrado no código abaixo:

```py
# (...)
   while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query, config={"configurable": {"session_id": session_id}})
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)
```

Agora, enquanto a sessão estiver ativa, podemos perguntar ao agente sobre informações em sua "memória de curto prazo", já que o conteúdo da sessão fica armazenado em memória. Isso será suficiente para que a maioria das interações básicas pareça mais natural, mas estamos abrindo precedente para problemas maiores: agora que podemos armazenar informações de sessão, após cada iteração ela só vai crescer e, ao lidar com dados gerados automaticamente por queries, o contexto da sessão crescerá muito rápido, atingindo logo os limites do modelo, muito antes de atingirmos os limites da memória do nosso computador.

Modelos como o Gemini são bem conhecidos por suas [janelas de contexto longo](https://ai.google.dev/gemini-api/docs/long-context), mas até um milhão de tokens pode se esgotar bem rápido se preenchermos o contexto com dados. Um contexto longo também pode representar um problema para alguns modelos, pois a recuperação de informações fica cada vez mais difícil, algo também conhecido como o problema da [agulha no palheiro (needle in a haystack)](https://cloud.google.com/blog/products/ai-machine-learning/the-needle-in-the-haystack-test-and-how-gemini-pro-solves-it?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog).

Existem técnicas para lidar com o problema do contexto crescente, incluindo compressão e sumarização, mas, para manter o contexto deste artigo curto (viu o que eu fiz aqui?), vamos guardar esses tópicos para o próximo artigo.

A versão final de `main.py` incluindo todas as modificações deste artigo fica assim:

```py
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown
from langchain_core.chat_history import InMemoryChatMessageHistory
import os
import uuid

PROJECT_ID = os.environ.get("GCP_PROJECT")
LOCATION = os.environ.get("GCP_REGION", "us-central1")
STAGING_BUCKET = os.environ.get("STAGING_BUCKET_URI")

vertexai.init(
    project=PROJECT_ID,
    location=LOCATION,
    staging_bucket=STAGING_BUCKET
)

MODEL = os.environ.get("GEMINI_MODEL", "gemini-2.5-pro-preview-05-06")

instance = osquery.SpawnInstance()
debug = False

def set_debug_mode(debug_mode: bool):
    """Toggle debug mode. Call this function to enable or disable debug mode.
    
    Args:
        debug_mode (bool): True to enable debug mode, False to disable it.


    Returns:
        None
    """
    global debug
    debug = debug_mode

def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()  # This may raise an exception

    if debug:
        print("Executing query: ", query)

    result = instance.client.query(query)
    if debug:
        print("Query result: ", {
            "status": result.status.message if result.status else None, 
            "response": result.response if result.response else None
        })

    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()  # This may raise an exception
    
    response = instance.client.query("select name from sqlite_temp_master").response
    tables = [ t["name"] for t in response ]
    return f"""
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
  - you can use the set_debug_mode function to enable or disable debug mode.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fulfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
    """

chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history

def main():
    session_id = uuid.uuid4()
    agent = agent_engines.LangchainAgent(
        model = MODEL,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode
        ],
        chat_history=get_chat_history,
    )
    
    console = Console()

    print("Welcome to the Emergency Diagnostic Agent\n")
    print("What is the nature of your diagnostic emergency?")

    while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query, config={"configurable": {"session_id": session_id}})
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)

    print("Goodbye!")

if __name__ == "__main__":
    main()
```

## Conclusões

Neste artigo aprendemos a importância de ajustar um prompt de sistema para obter respostas consistentes de um agente. Também vimos na prática como funciona a chamada de múltiplas ferramentas e como usar ferramentas para ativar ou desativar feature flags para alterar o comportamento do agente. Por fim, aprendemos como gerenciar o estado da sessão usando o histórico de chat em memória.

No próximo artigo da série, veremos como habilitar a persistência entre sessões usando um banco de dados real, revisitaremos a noção de tokens e discutiremos a técnica de compressão de contexto.

## Apêndice: Coisas divertidas para testar

Agora que nosso agente está mais robusto, use esta seção como um guia prático para testar os novos recursos em ação. Observe como ele agora se lembra do contexto entre perguntas e como você pode pedir para ele explicar o seu próprio trabalho.

```
>> run a level 1 diagnostic procedure
>> run a level 2 diagnostic procedure
>> explain the previous procedure step by step
>> find any orphan processes
>> show me the top resource consuming processes
>> write a system prompt to transfer your current knowledge to another agent
>> search the system for malware
>> is this computer connected to the internet?
>> why is my computer slow?
>> take a snapshot of the current performance metrics
>> compare the current performance metrics with the previous snapshot
>> give me a step by step process to fix the issues you found
>> how many osqueryd processes are in memory?
>> give me a script to kill all osqueryd processes
>> who am i?
```

Se você encontrar outros prompts interessantes, compartilhe suas experiências na seção de comentários abaixo.

Na próxima parte desta série, [Como Criar um Agente de Diagnóstico com o Agent Development Kit]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}), vamos refatorar o agente de diagnóstico do código boilerplate puro com o SDK para o Agent Development Kit (ADK) do Google e integrar o Vertex AI RAG para descoberta dinâmica de schemas de tabelas. Até a próxima!
