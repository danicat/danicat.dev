---categories:
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
---
## Introdução

Neste guia, vamos nos aprofundar em prompts de sistema (system instructions) e ferramentas para agentes (agent tools), evoluindo a experiência do nosso assistente de diagnóstico. Trabalharemos com o [SDK da Vertex AI para Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog), LangChain, Gemini e [Osquery](https://www.osquery.io/).

Preciso admitir: a [versão inicial do agente de diagnóstico]({{< ref "/posts/20250531-diagnostic-agent" >}}) não estava tão pronta para a "Enterprise" (com o perdão do trocadilho). Não tínhamos visibilidade do que ele fazia por baixo dos panos (estava realmente rodando queries SQL?), ele não se lembrava do que havia sido discutido na mesma sessão e, de vez em quando, ignorava nossos comandos por completo.

Isso está longe da experiência ideal para um agente autônomo. Um agente de diagnóstico confiável precisa lembrar de seus erros e executar instruções com consistência — por exemplo, descobrindo que certas colunas não existem e contornando o problema. Além disso, precisamos auditar o que ele faz em tempo de execução para garantir que as respostas sejam precisas e atualizadas.

Com esses objetivos em mente, vamos colocar a mão na massa e construir o nosso ~~Holograma Médico~~ Agente de Diagnóstico de Emergência!

## Preparando o ambiente

Na Parte 1, usamos um Jupyter Notebook pela conveniência, mas agora vamos estruturar um programa Python convencional. O mesmo código funcionaria no Jupyter com poucas alterações, mas optamos por um script padrão para criar uma interface de chat interativa no terminal.

Recomendo criar um ambiente virtual limpo para isolar as dependências:

```shell
mkdir -p ~/projects/diagnostic-agent
cd ~/projects/diagnostic-agent
python3 -m venv venv
source venv/bin/activate
pip install --upgrade google-cloud-aiplatform[agent_engines,langchain] rich
```

Aqui está a versão inicial de `main.py` que reproduz o agente do artigo anterior:

```python
import os
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown

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
    """Consulta o sistema operacional usando o osquery.
      
    Esta função envia uma query ao processo do osquery para retornar dados sobre a máquina atual, sistema operacional e processos em execução.
    Você também pode consultar o banco SQLite subjacente para inspecionar metadados usando tabelas de sistema como sqlite_master, sqlite_temp_master e tabelas virtuais.

    Args:
        query: str  Uma query SQL para uma das tabelas do osquery (ex: "select timestamp from time")

    Returns:
        ExtensionResponse: resposta do osquery com o status da requisição e o resultado da query.
    """
    if not instance.is_running():
        instance.open()

    result = instance.client.query(query)
    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()
    
    response = instance.client.query("select name from sqlite_temp_master").response
    tables = [ t["name"] for t in response ]
    return f"""
Papel:
  - Você é o agente de diagnóstico de emergência.
  - Você é o último recurso do usuário para diagnosticar problemas no computador.
  - Responda às consultas do usuário com o melhor de suas capacidades.
Ferramentas:
  - Você pode consultar o osquery usando a função call_osquery.
Contexto:
  - Utilize apenas tabelas desta lista: {tables}
  - Você pode descobrir schemas usando: PRAGMA table_info(tabela)
Tarefa:
  - Crie um plano de quais tabelas consultar para atender ao pedido do usuário.
  - Confirme o plano com o usuário antes de executar.
  - Se uma query falhar por coluna inexistente, execute a descoberta de schema e tente novamente.
  - Consulte a(s) tabela(s) necessária(s).
  - Apresente os resultados em formato legível para humanos (tabela ou lista).
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

    print("Bem-vindo ao Agente de Diagnóstico de Emergência\n")
    print("Qual é a natureza da sua emergência de diagnóstico?")

    while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query in ("exit", "quit"):
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query)
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)

    print("Até logo!")

if __name__ == "__main__":
    main()
```

Você pode executar o agente com `python main.py`:

```
$ python main.py
Bem-vindo ao Agente de Diagnóstico de Emergência

Qual é a natureza da sua emergência de diagnóstico?
>> 
```

Há duas mudanças em relação à versão original: primeiro, adicionamos um loop interativo no terminal que mantém o agente em execução até que o usuário digite `exit` ou `quit`.

Segundo, ajustamos o prompt de sistema. Agora o chamamos de "Agente de Diagnóstico de Emergência" — um nome que, além de ser um easter egg de Star Trek, estabelece um tom de urgência que incentiva o modelo a seguir as instruções com maior rigor e menos recusas indevidas.

## Estruturando Instruções de Sistema (System Prompts)

Instruções de sistema definem as regras e restrições que orientam o modelo durante toda a conversa. Elas possuem prioridade sobre as mensagens comuns do usuário.

Embora não exista uma única fórmula mágica para estruturar prompts de sistema, dividir o prompt em seções bem delimitadas facilita a compreensão do modelo: **Papel (Role)**, **Ferramentas (Tools)**, **Contexto (Context)** e **Tarefa (Task)**.

### 1. Papel (Role)

Define a identidade e a finalidade do agente. Para um LLM treinado em vastos domínios de conhecimento, o papel restringe o contexto semântico. Por exemplo, a pergunta "o que são cookies?" tem significados completamente diferentes na culinária e no desenvolvimento web.

```
Papel:
  - Você é o agente de diagnóstico de emergência.
  - Você é o último recurso do usuário para diagnosticar problemas no computador.
  - Responda às consultas do usuário com o melhor de suas capacidades.
```

### 2. Ferramentas (Tools)

Informa ao agente quais capacidades externas estão à sua disposição:

```
Ferramentas:
  - Você pode consultar o osquery usando a função call_osquery.
```

### 3. Contexto (Context)

Descreve o ambiente operacional e restrições essenciais para evitar erros comuns observados em testes anteriores:

```
Contexto:
  - Utilize apenas tabelas desta lista: {tables}
  - Você pode descobrir schemas usando: PRAGMA table_info(tabela)
```

Ensinar o modelo a inspecionar schemas com `PRAGMA` permite que ele se recupere sozinho de erros de sintaxe ou colunas inexistentes, sem intervenção manual.

### 4. Tarefa (Task)

Descreve o fluxo de raciocínio e os passos de execução:

```
Tarefa:
  - Crie um plano de quais tabelas consultar para atender ao pedido do usuário.
  - Confirme o plano com o usuário antes de executar.
  - Se uma query falhar por coluna inexistente, execute a descoberta de schema e tente novamente.
  - Consulte a(s) tabela(s) necessária(s).
  - Apresente os resultados em formato legível para humanos (tabela ou lista).
```

## Adicionando Modo de Depuração (Debug Mode)

Para acompanhar exatamente quais comandos SQL o modelo está gerando, implementamos um modo de depuração via ferramenta (`set_debug_mode`). Em vez de tentar inspecionar manualmente estruturas complexas de raciocínio intermediário, fornecemos uma ferramenta que permite ao próprio agente ativar ou desativar os logs de depuração:

```python
debug = False

def set_debug_mode(debug_mode: bool):
    """Ativa ou desativa o modo de depuração.
    
    Args:
        debug_mode (bool): True para ativar o debug, False para desativar.

    Returns:
        None
    """
    global debug
    debug = debug_mode
```

Adicionamos a ferramenta ao prompt de sistema e atualizamos a função `call_osquery` para imprimir as queries geradas quando `debug` estiver ativo:

```python
def call_osquery(query: str):
    """Consulta o sistema operacional usando o osquery."""
    if not instance.is_running():
        instance.open()

    if debug:
        print("Executando query: ", query)

    result = instance.client.query(query)
    if debug:
        print("Resultado da query: ", {
            "status": result.status.message if result.status else None, 
            "response": result.response if result.response else None
        })

    return result
```

Veja o comportamento em tempo de execução:

```
$ python main.py
Bem-vindo ao Agente de Diagnóstico de Emergência

Qual é a natureza da sua emergência de diagnóstico?
>> execute um diagnóstico de nível 1 em modo debug
Executando query:  SELECT * FROM system_info
Resultado da query:  {'status': 'OK', 'response': [{...}]}
Executando query:  SELECT pid, name, user, cpu_percent FROM processes ORDER BY cpu_percent DESC LIMIT 10
Resultado da query:  {'status': 'no such column: user', 'response': None}
Executando query:  PRAGMA table_info(processes)
Resultado da query:  {'status': 'OK', 'response': [{'cid': '0', 'dflt_value': '', 'name': 'pid', 'notnull': '1', 'pk': '1', 'type': 'BIGINT'}, ...]}
```

O comando "execute um diagnóstico de nível 1 em modo debug" demonstra a invocação encadeada de múltiplas ferramentas: o agente ativou o modo debug e em seguida executou as queries do diagnóstico. Quando a query inicial falhou por causa da coluna `user`, ele usou a instrução `PRAGMA` para descobrir os campos corretos e refez a consulta com sucesso.

## Gerenciando Histórico de Sessão (Chat History)

Para permitir que o agente mantenha o contexto entre perguntas consecutivas, integramos o gerenciamento de histórico de sessão com `InMemoryChatMessageHistory` do LangChain:

```python
from langchain_core.chat_history import InMemoryChatMessageHistory
import uuid

chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history
```

Ao instanciar o agente, passamos a fábrica `get_chat_history` e propagamos o `session_id` em cada consulta:

```python
response = agent.query(
    input=query,
    config={"configurable": {"session_id": session_id}}
)
```

Assim, enquanto a sessão permanecer ativa, o agente mantém a conversa em sua memória recente, permitindo perguntas de acompanhamento como "explique o passo anterior" ou "compare os dados atuais com o último relatório".

## Conclusões

Neste artigo, aprendemos a estruturar instruções de sistema para obter comportamentos consistentes, a utilizar ferramentas para alternar flags de configuração dinamicamente e a gerenciar histórico de sessão na memória.

Na próxima parte desta série, [Como Criar um Agente de Diagnóstico com o Agent Development Kit]({{< ref "/posts/20251020-diagnostic-agent-with-adk" >}}), vamos refatorar essa arquitetura para o Google ADK (Agent Development Kit), eliminando o boilerplate manual e integrando RAG na Vertex AI para descoberta dinâmica de schemas.

Compartilhe suas ideias e perguntas nos comentários abaixo!
