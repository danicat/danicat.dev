+++
date = '2025-06-11T00:00:00+01:00'
title = 'Prompt Audacioso: Um Guia Prático para Instruções de Sistema e Ferramentas de Agente'
summary = "Este artigo explora os conceitos de instrução de sistema, histórico de sessão e ferramentas de agente para criar um assistente de diagnóstico mais inteligente."
tags = ["gemini", "vertex-ai", "python"]
categories = ["AI & Development"]
+++
{{< translation-notice >}}
## Introdução

Neste guia, vamos aprender mais sobre system prompts e ferramentas de agente para que possamos construir uma nova e melhorada experiência de agente de diagnóstico. Trabalharemos com o [Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog), LangChain, Gemini e [osquery](https://www.osquery.io/).

Devo admitir, a [versão inicial do agente de diagnóstico](https://danicat.dev/posts/20250531-diagnostic-agent/) não estava muito pronta para a "Enterprise" (trocadilho intencional). Não tínhamos muita visibilidade sobre o que ele estava fazendo por baixo dos panos (ele estava realmente executando alguma query?), ele não se lembrava de coisas discutidas na mesma "sessão" e, de vez em quando, também ignorava nossos comandos completamente.

Isso está longe da experiência que desejamos de um agente adequado. O agente de diagnóstico ideal precisa ser capaz de lembrar seus erros e executar instruções de forma consistente, por exemplo, aprendendo que certas colunas não estão disponíveis e contornando isso. Além disso, podemos realmente confiar que ele está fazendo o que diz que está fazendo? Devemos ser capazes de ver as queries a qualquer momento para garantir que as informações que ele está retornando estão corretas e atualizadas.

Com esses objetivos em mente, vamos colocar as mãos na massa e começar a construir nosso Agente de Diagnóstico de Emergência ~~Médica Holográfica~~!

## Preparando o Terreno

Da última vez, escrevemos o código em um Jupyter notebook por conveniência, mas desta vez vamos escrever um programa Python regular. O mesmo código também funcionaria no Jupyter com alterações mínimas, mas estamos fazendo isso para que possamos usar o agente de diagnóstico com uma interface de chat adequada.

Para qualquer projeto Python, sempre recomendo começar com um virtual env limpo para manter as dependências autocontidas:

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
  - Create a plan for which tables to query to fullfill the user request
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

Existem duas pequenas alterações em comparação com o código original: primeiro, agora temos um loop principal que manterá o agente em execução até que o usuário digite "exit" ou "quit". Isso criará nossa interface de chat.

Segundo, ajustamos o system prompt para melhorar a consistência do agente. Agora o chamamos de "Emergency Diagnostic Agent" - este nome não apenas funciona como um elegante [easter egg de Star Trek]https://en.wikipedia.org/wiki/The_Doctor_(Star_Trek:_Voyager), mas também define um tom de urgência que, com base em [pesquisas emergentes](https://arxiv.org/pdf/2307.11760), pode incentivá-lo a cumprir nossos pedidos com mais diligência. (Confira também esta [entrevista recente](https://www.reddit.com/r/singularity/comments/1kv7hm2/sergey_brin_we_dont_circulate_this_too_much_in/) (original em inglês))

Não vamos ameaçar nosso pobre Agente de Diagnóstico de Emergência - e posso garantir que nenhum agente foi prejudicado na produção deste texto - mas, chamá-lo de agente de "Emergência" deve definir o tom para que ele tente cumprir nossos pedidos da melhor maneira possível. Na versão anterior do system prompt, tive casos em que o agente se recusou a fazer uma tarefa porque "achava" que não era capaz de fazê-la ou não sabia quais tabelas consultar.

Claro, chamá-lo de agente de emergência não é suficiente para garantir o comportamento desejado, então adicionamos mais algumas instruções para guiar o comportamento do modelo, como veremos abaixo.

## Instruções de Sistema

Instruções de sistema, também chamadas de system prompt, são um conjunto de instruções que guiam o comportamento do LLM durante toda a conversa. As instruções de sistema são especiais, pois têm prioridade mais alta sobre as interações regulares de chat. Uma maneira de imaginar isso é como se as instruções de sistema fossem sempre repetidas junto com o prompt que você está enviando para o modelo.

Não há um consenso forte na literatura sobre como um system prompt deve ser, mas temos alguns padrões testados em batalha que estão surgindo do uso diário. Por exemplo, um que é praticamente um consenso é dedicar o início do system prompt para atribuir um papel ao agente, para que ele esteja ciente de seu propósito e possa produzir respostas mais coerentes.

Para este agente em particular, optei por incluir as seguintes seções em meu system prompt: papel, ferramentas, contexto e tarefa. Essa estrutura funcionou bem durante minha fase de testes, mas não se apegue demais a ela: experimente com seus prompts e veja se consegue resultados melhores. A experimentação é fundamental para alcançar bons resultados com LLMs.

Agora vamos dar uma olhada em cada seção do prompt.

### System Prompt: Papel

Um papel nada mais é do que a razão da existência do agente. Pode ser tão simples quanto "você é um engenheiro de software" ou "você é um agente de diagnóstico", mas podem ser um pouco mais elaborados, incluindo uma descrição detalhada, regras de comportamento, restrições e outros.

Para um modelo de linguagem grande treinado em todos os tipos de dados, o papel ajuda a definir o tom para o conhecimento de domínio que ele precisará acessar para responder às suas perguntas. Em outras palavras, ele dá significado semântico às suas perguntas... Imagine a pergunta "o que são cookies?", por exemplo. Estamos falando de cookies comestíveis ou cookies de navegador? Se o papel do agente for indefinido, essa pergunta é completamente ambígua, mas assim que definimos o papel para algo técnico (por exemplo, "você é um engenheiro de software"), a ambiguidade desaparece.

Para este agente, o papel é descrito como:

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
```

Além da definição direta ("você é o agente de diagnóstico de emergência"), adicionamos uma descrição mais longa para definir o tom do comportamento do modelo e, esperançosamente, influenciá-lo a levar nossos pedidos "a sério", como mencionado antes, a iteração anterior deste agente tinha a má tendência de recusar pedidos.

### System Prompt: Ferramentas

Ferramentas é a seção que explica ao agente suas capacidades de interagir com sistemas externos além de seu modelo principal. As ferramentas podem ser de vários tipos, mas a maneira mais comum de fornecer uma ferramenta ao agente é por meio de chamadas de função.

Os agentes podem usar ferramentas para recuperar informações, executar tarefas e manipular dados. O Vertex AI SDK for Python tem suporte para funções fornecidas pelo usuário e ferramentas integradas como Google Search e execução de código. Você também pode usar extensões mantidas pela comunidade por meio da interface Model Context Protocol (MCP).

Para nosso agente, precisamos dizer a ele que ele pode chamar o _osquery_:

```
Tools:
  - you can call osquery using the call_osquery function.
```

### System Prompt: Contexto

Em seguida, temos o contexto, que informa ao agente sobre o ambiente em que opera. Eu uso esta seção para destacar explicitamente e corrigir comportamentos indesejados que iterações anteriores do agente eram propensas a fazer. Por exemplo, percebi bem cedo no desenvolvimento que o agente tentaria "adivinhar" quais tabelas estavam disponíveis e enviaria queries às cegas, resultando em uma alta taxa de erro. Adicionar a lista de tabelas ao contexto ajudou a mitigar esse problema.

Semelhante é a tendência do agente de tentar adivinhar os nomes das colunas em uma tabela, em vez de tentar descobrir os nomes primeiro. Neste caso particular, resisti à tentação de instruir o agente a sempre usar `SELECT *` porque esta é uma má prática (recupera mais dados do que o necessário), mas em vez disso eu o "ensinei" como descobrir um schema usando a instrução `PRAGMA`.

Desta forma, o agente ainda cometerá erros ao adivinhar nomes de colunas, mas tem uma maneira de corrigir o curso sem intervenção humana.

A seção de contexto revisada do system prompt é mostrada abaixo.

```
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
```

Note que `tables` é uma variável que contém todas as tabelas que descobrimos do osquery antes de iniciar o modelo.

### System Prompt: Tarefa

Finalmente, a tarefa. Esta seção é usada para descrever como o agente deve interpretar seus pedidos e executá-los. As pessoas geralmente usam esta seção para definir as etapas necessárias para realizar a tarefa em questão.

No nosso caso particular, estamos usando esta seção para definir aproximadamente o plano, mas também adicionar algumas diretivas condicionais:

```
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

A etapa "confirmar o plano com o usuário antes de executar" é interessante, pois nos mostra como o agente está pensando sobre o processo, mas pode ser um pouco irritante depois de interagir com o agente por um tempo. Sempre podemos pedir ao agente para nos dizer o plano com um prompt, então a inclusão desta etapa é totalmente opcional.

Inicialmente pensei nesta etapa como uma forma de depurar o agente, mas na seção seguinte vamos explorar uma maneira diferente de fazer isso.

Com a combinação dessas quatro seções, temos o system prompt completo. Este prompt revisado produziu resultados mais consistentes durante meus testes em preparação para este artigo. Ele também tem o benefício de ser "amigável ao ser humano", por isso é mais fácil de adaptar quando novas regras são introduzidas.

Aqui está a visão completa deste system prompt:

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
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

Como nota lateral, acredito que ainda temos muitas oportunidades para melhorá-lo e uma das minhas atuais áreas de interesse é como alcançar um system prompt autoaperfeiçoável. Isso poderia potencialmente ser alcançado pedindo, ao final de uma sessão, para o modelo resumir seus aprendizados em um novo system prompt para sua iteração futura. O prompt poderia ser armazenado em um banco de dados e carregado na próxima sessão. Isso, é claro, levanta preocupações sobre a degradação do system prompt ou, pior ainda, ataques usando injeção de prompt, então não é tão trivial quanto parece. No entanto, é um exercício divertido e posso escrever sobre isso em um futuro próximo.

## Habilitando o Modo de Depuração

Outra preocupação sobre o design original é a falta de observabilidade sobre o que o agente está fazendo por baixo dos panos. Existem duas abordagens diferentes que podemos aplicar aqui, uma um pouco mais dolorosa que a outra: 1) espiar os "pensamentos" do LLM e tentar encontrar as chamadas de ferramenta entre eles (muito doloroso), ou; 2) adicionar alguma funcionalidade de depuração à própria função para que ela produza as informações que queremos durante a execução (a solução mais fácil geralmente é a correta).

Devo admitir, passei uma quantidade doentia de tempo na opção 1, antes de perceber que poderia fazer a opção 2. Se você realmente quer seguir o caminho do raciocínio do LLM, pode fazê-lo por meio de uma configuração chamada [return_intermediate_steps](https://api.python.langchain.com/en/latest/agents/langchain.agents.agent.AgentExecutor.html#langchain.agents.agent.AgentExecutor.return_intermediate_steps). Devo dizer que isso é muito interessante do ponto de vista do aprendizado, mas depois de passar algumas horas tentando descobrir o formato da saída (dica: [não é exatamente json](https://github.com/langchain-ai/langchain/issues/10099)) decidi que analisar não valia realmente a pena.

Então, como funciona a estratégia simples? Estamos adicionando uma flag de depuração e uma ferramenta para ativar e desativar essa flag. Este truque surpreendentemente simples na verdade abre um mundo totalmente novo de potencial: estamos dando ao agente a oportunidade de modificar seu próprio comportamento!

A implementação do modo de depuração é composta por uma variável global e uma função para configurá-la:

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

Também precisamos mencioná-lo no system prompt:

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

Finalmente, precisamos adaptar `call_osquery` para usar a nova flag `debug`:

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

Com todas essas alterações implementadas, vamos dar uma olhada em como o agente chama o _osquery_ usando a flag de depuração recém-implementada:

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

 • Hostname: petruzalek-mac.roam.internal                                                                                                                               
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

Observe que o comando emitido foi "run a level 1 diagnostic procedure **in debug mode**", o que demonstra uma capacidade interessante do agente: invocação de múltiplas ferramentas. Se julgar necessário, ele é capaz de invocar não apenas a mesma ferramenta várias vezes, mas também ferramentas diferentes ao mesmo tempo. Portanto, não foi necessário habilitar o modo de depuração antes de solicitar o relatório: o agente conseguiu fazer tudo de uma vez.

Observe também como o agente falhou inicialmente ao solicitar uma coluna de usuário, mas depois usou a instrução PRAGMA para descobrir o schema correto e tentar novamente a query com sucesso. Esta é uma demonstração perfeita da capacidade do agente de se recuperar de erros devido ao nosso system prompt aprimorado.

## Preservando o Histórico do Chat

Nossa tarefa final hoje é garantir que o agente se lembre do que estamos falando para que possamos fazer perguntas de esclarecimento e investigar mais o sistema seguindo uma linha de investigação coerente.

No [artigo anterior](https://danicat.dev/posts/20250605-vertex-ai-sdk-python/) exploramos como os LLMs são stateless e que precisamos continuar "lembrando-os" do estado atual da conversa usando "turnos". Felizmente com o LangChain não precisamos fazer isso manualmente e podemos contar com um recurso chamado [chat history](https://python.langchain.com/api_reference/core/chat_history.html).

A beleza do chat history é que qualquer coisa que implemente [BaseChatMessageHistory](https://python.langchain.com/api_reference/core/chat_history/langchain_core.chat_history.BaseChatMessageHistory.html#langchain_core.chat_history.BaseChatMessageHistory) pode ser usada aqui, o que nos permite usar todos os tipos de armazenamentos de dados, incluindo a criação dos nossos próprios. Por exemplo, na documentação oficial do Vertex AI, você pode encontrar exemplos de uso de [Firebase, Bigtable e Spanner](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain#chat-history?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog).

Não precisamos de um banco de dados completo no momento, então vamos nos contentar com `InMemoryChatMessageHistory`, que, como o nome sugere, armazenará tudo na memória.

Aqui está uma implementação típica, tecnicamente suportando múltiplas sessões usando o dicionário `chats_by_session_id` para pesquisa (código retirado da [documentação do langchain](https://python.langchain.com/docs/versions/migrating_memory/chat_history/#chatmessagehistory)):

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

Um aviso rápido para que você não cometa o mesmo erro que eu: o argumento `chat_history` espera um tipo `Callable`, então você não deve invocar a função ali, mas passar a própria função. O LangChain usa um padrão factory aqui; ele invoca a função fornecida (`get_chat_history`) sob demanda com um `session_id` para obter ou criar o objeto de histórico correto. Este design é o que permite ao agente gerenciar múltiplas conversas separadas simultaneamente.

A assinatura da função pode incluir um ou dois argumentos. Se um argumento, presume-se que seja um `session_id`, e se forem dois argumentos, eles são interpretados como `user_id` e `conversation_id`. Mais informações sobre isso podem ser encontradas na documentação [RunnableWithMessageHistory](https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.history.RunnableWithMessageHistory.html).

A última peça do quebra-cabeça é passar o `session_id` para o executor do modelo. Isso é feito por meio do argumento `config`, conforme mostrado no código abaixo:

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

Agora, enquanto a sessão estiver ativa, podemos perguntar ao agente sobre informações em sua "memória de curto prazo", já que o conteúdo da sessão é armazenado na memória. Isso será suficiente para que a maioria das interações básicas pareçam mais naturais, mas estamos abrindo precedentes para problemas maiores: agora que podemos armazenar informações da sessão, após cada iteração ela só crescerá e, ao lidar com dados gerados automaticamente a partir de queries, o contexto da sessão crescerá muito rapidamente, atingindo em breve os limites do modelo, e muito antes de atingirmos os limites da memória do nosso computador.

Modelos como o Gemini são bem conhecidos por suas [longas janelas de contexto](https://ai.google.dev/gemini-api/docs/long-context), mas mesmo um milhão de tokens podem ser esgotados muito rapidamente se preenchermos o contexto com dados. O contexto longo também pode representar um problema para alguns modelos, pois a recuperação se torna cada vez mais difícil - também conhecido como o problema da [agulha no palheiro](https://cloud.google.com/blog/products/ai-machine-learning/the-needle-in-the-haystack-test-and-how-gemini-pro-solves-it?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog).

Existem técnicas para lidar com o problema do contexto crescente, incluindo compressão e sumarização, mas, para manter o contexto deste artigo curto (viu o que eu fiz ali?), vamos guardá-las para o próximo artigo.

A versão final de `main.py`, incluindo todas as modificações neste artigo, fica assim:

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
  - Create a plan for which tables to query to fullfill the user request
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

Neste artigo, aprendemos a importância de ajustar um system prompt para obter respostas consistentes de um agente. Também vimos na prática como funciona a chamada de múltiplas ferramentas e como usar ferramentas para habilitar ou desabilitar flags de recursos para alterar o comportamento do agente. Por último, mas não menos importante, aprendemos sobre como gerenciar o estado da sessão usando o histórico de chat na memória.

No próximo artigo da série, veremos como habilitar a persistência entre sessões usando um banco de dados real, revisitaremos a noção de tokens e discutiremos a técnica de compressão de contexto.

## Apêndice: Coisas Divertidas para Tentar

Agora que nosso agente está mais robusto, use esta seção como um guia prático para testar os novos recursos em ação. Observe como ele agora se lembra do contexto entre as perguntas e como você pode pedir para ele explicar seu trabalho.

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
>> compare the current perfomance metrics with the previous snapshot
>> give me a step by step process to fix the issues you found
>> how many osqueryd processes are in memory?
>> give me a script to kill all osqueryd processes
>> who am i?
```

Se você encontrar outros prompts interessantes, por favor, compartilhe suas experiências na seção de comentários abaixo. Até a próxima!
```