---categories:
- Agent Development
date: '2025-05-31T01:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 1
summary: Como criar um agente de diagnóstico que compreende linguagem natural usando Gemini e Vertex AI Agent Engine.
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: "Como transformei meu computador na \"USS Enterprise\" usando Agentes de IA"
slug: "diagnostic-agent-uss-enterprise"
aliases:
  - "/pt-br/posts/20250531-diagnostic-agent/"
description: "Construa um assistente de diagnóstico inspirado em Star Trek usando Python, Gemini, Vertex AI Agent Engine e function calling do osquery para telemetria em linguagem natural."
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-cloud-aiplatform"
  - "osquery"
---
_Espaço: a fronteira final. Estas são as viagens da nave estelar Enterprise. Em sua missão de cinco anos: para explorar novos mundos; para pesquisar novas vidas e novas civilizações; para audaciosamente ir onde nenhum homem jamais esteve._

## Introdução

Quando eu era criança, graças à influência do meu pai, me acostumei a ouvir essas palavras quase todos os dias. Suspeito que a paixão dele por Star Trek tenha tido um papel enorme na minha decisão de seguir a carreira de engenharia de software. (Para quem não conhece Star Trek, essa introdução passava no início de cada episódio da série clássica).

Star Trek sempre esteve à frente de seu tempo. Mostrou o [primeiro beijo interracial da televisão americana](https://en.wikipedia.org/wiki/Kirk_and_Uhura%27s_kiss), em uma época em que uma cena dessas causava imensa controvérsia. Também retratou inúmeras tecnologias "futuristas" que hoje são rotineiras, como smartphones e videoconferências.

Uma coisa realmente marcante é a forma como os engenheiros da série interagem com os computadores. Embora a gente veja teclados e botões sendo pressionados de vez em quando, muitos dos comandos são dados por voz, em linguagem natural. Alguns dos comandos são icônicos — como quando pedem ao computador para executar um "procedimento de diagnóstico de nível 1", algo que aconteceu tantas vezes que praticamente virou [uma piada](https://www.youtube.com/watch?v=cYzByQjzTb0) entre os fãs mais apaixonados.

Avançando mais de 30 anos no tempo, aqui estamos nós na Era da IA, uma transformação tecnológica que promete ser maior do que a própria internet. Muita gente tem receio de como a IA pode impactar seus empregos ([escrevi sobre isso na semana passada]({{< ref "/posts/20250528-vibe-coding" >}})), mas crescer assistindo Star Trek me ajudou a enxergar com mais clareza como o papel da pessoa engenheira vai mudar nos próximos anos. Em vez de comandar a máquina exclusivamente por texto, instruindo cada detalhe passo a passo por meio de linhas de código e compiladores, muito em breve vamos conversar e fazer brainstorming diretamente com nossos computadores.

Para ajudar a visualizar esse conceito na prática, vamos usar a tecnologia disponível hoje para criar um agente simples que nos permite interagir com a nossa própria máquina usando linguagem natural.

## O que você vai precisar para este tutorial

Para o desenvolvimento, usaremos Python em um Jupyter Notebook, o que facilita bastante a experimentação. As principais ferramentas e bibliotecas que utilizaremos são:

*   [Vertex AI Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/overview?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog)
*   [Osquery](https://www.osquery.io/) com [bindings em Python](https://github.com/osquery/osquery-python)
*   [Jupyter Notebook](https://jupyter.org/) [opcional] (estou usando a [extensão do Jupyter para VSCode](https://code.visualstudio.com/docs/datascience/jupyter-notebooks))

Os exemplos abaixo usam o Gemini 2.0 Flash, mas você pode usar qualquer [variante do Gemini](https://ai.google.dev/gemini-api/docs/models). Não faremos o deploy deste agente no Google Cloud desta vez porque queremos utilizá-lo para responder perguntas sobre a máquina local, e não sobre um servidor remoto na nuvem.

## Visão geral de agentes

Se você já conhece o funcionamento de agentes de IA, pode pular esta seção.

Um agente de IA é um sistema capaz de perceber seu ambiente e tomar decisões autônomas para atingir objetivos específicos. Ao contrário dos Modelos de Linguagem Tradicionais (LLMs), que focam primordialmente em gerar texto a partir de um prompt, agentes de IA podem interagir com seu ambiente, tomar decisões e executar tarefas para concluir seus objetivos. Isso é viabilizado pelo uso de "ferramentas" (tools), que fornecem dados ao agente e concedem a ele a capacidade de executar ações.

Para demonstrar essa tecnologia, vamos usar o LangChain por meio do [Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog). Primeiro, instale os pacotes necessários no seu ambiente:

```shell
pip install --upgrade --quiet google-cloud-aiplatform[agent_engines,langchain]
```

Você também precisará configurar as Application Default Credentials (ADC) da Google Cloud CLI:

```shell
gcloud auth application-default login
```

*Nota:* dependendo do ambiente em que estiver executando esta demonstração, pode ser necessário usar um método de autenticação diferente.

Agora estamos prontos para trabalhar no script Python. Primeiro, inicializamos o SDK informando o Project ID e a região no Google Cloud:

```python
import vertexai

vertexai.init(
    project="meu-project-id",                 # Seu Project ID.
    location="us-central1",                   # Sua região no Google Cloud.
    staging_bucket="gs://meu-staging-bucket", # Seu bucket de staging.
)
```

Com a configuração inicial pronta, criar um agente usando LangChain no Agent Engine é bastante direto:

```python
from vertexai import agent_engines

model = "gemini-2.0-flash" # sinta-se à vontade para testar outros modelos!

model_kwargs = {
    # temperature (float): controla a aleatoriedade na seleção de tokens.
    "temperature": 0.20,
}

agent = agent_engines.LangchainAgent(
    model=model,                # Obrigatório.
    model_kwargs=model_kwargs,  # Opcional.
)
```

O setup acima já é suficiente para você enviar perguntas ao agente, da mesma forma que faria com um LLM comum:

```python
response = agent.query(
    input="que horas são agora?"
)
print(response)
```

O retorno será algo como:

```
{'input': 'que horas são agora?', 'output': 'Como uma IA, não tenho uma noção de horário "atual" ou localização da mesma forma que um ser humano. Meus conhecimentos não são atualizados em tempo real.\n\nPara verificar o horário atual, você pode:\n\n*   **Checar seu dispositivo:** Seu computador, celular ou tablet exibe a hora atual.\n*   **Fazer uma busca rápida:** Pesquise "que horas são" em um mecanismo de busca como o Google.'}
```

Dependendo das configurações, do prompt e da aleatoriedade do modelo, ele pode responder que não sabe a hora ou pode "alucinar" e inventar um timestamp. Como o modelo não tem um relógio interno, ele não consegue responder a essa pergunta... a menos que a gente entregue um relógio a ele!

## Function Calling (Chamadas de Funções)

Uma das maneiras mais práticas de estender a capacidade do nosso agente é fornecer funções Python para ele executar. O processo é simples, mas vale ressaltar: quanto melhor for a documentação da função (docstring), mais fácil será para o modelo entender quando e como chamá-la. Vamos definir a função para checar o horário atual:

```python
import datetime

def get_current_time():
    """Retorna o horário atual como um objeto datetime.

    Args:
        None
    
    Returns:
        datetime: horário atual no formato datetime
    """
    return datetime.datetime.now()
```

Agora que temos uma função que retorna a hora do sistema, vamos recriar o agente, desta vez informando que essa ferramenta existe:

```python
agent = agent_engines.LangchainAgent(
    model=model,                # Obrigatório.
    model_kwargs=model_kwargs,  # Opcional.
    tools=[get_current_time]
)
```

E fazemos a pergunta novamente:

```python
response = agent.query(
    input="que horas são agora?"
)
print(response)
```

A saída será parecida com:

```
{'input': 'que horas são agora?', 'output': 'O horário atual é 18:36:42 UTC em 30 de maio de 2025.'}
```

Agora o agente consulta a ferramenta para responder com dados reais do sistema. Bacana, não?

## Coletando Informações do Sistema

Para o nosso agente de diagnóstico, vamos conceder a ele a capacidade de consultar dados sobre a máquina em que está rodando usando o [Osquery](https://www.osquery.io/). O Osquery é uma ferramenta open source criada originalmente pelo Facebook que permite consultar o sistema operacional por meio de queries SQL aplicadas a "tabelas virtuais".

Isso é muito vantajoso porque nos dá uma interface única e padronizada para inspecionar o sistema, e LLMs são extremamente competentes em escrever código SQL.

Você encontra as instruções de instalação do Osquery na [documentação oficial](https://osquery.readthedocs.io/en/stable/). Não vou reproduzi-las aqui porque os passos variam de acordo com o sistema operacional.

Com o Osquery instalado, instale os bindings Python oficiais:

```shell
pip install --upgrade --quiet osquery
```

Com os bindings instalados, você pode executar consultas importando o pacote `osquery`:

```python
import osquery

# Inicia um processo do osquery usando um socket de extensão efêmero.
instance = osquery.SpawnInstance()
instance.open()  # Pode lançar uma exceção

# Executa queries via Thrift APIs do osquery.
instance.client.query("select timestamp from time")
```

O método `query` retorna um objeto `ExtensionResponse` com o resultado da execução:

```python
ExtensionResponse(status=ExtensionStatus(code=0, message='OK', uuid=0), response=[{'timestamp': 'Fri May 30 17:54:06 2025 UTC'}])
```

Se você nunca trabalhou com Osquery antes, recomendo dar uma olhada no [schema oficial](https://www.osquery.io/schema/5.17.0/) para explorar quais tabelas e dados estão disponíveis para o seu sistema operacional.

### Uma dica sobre formatação

Nos exemplos anteriores, as saídas estavam sem formatação. Se você estiver rodando no Jupyter, pode formatar as respostas com Markdown importando os seguintes módulos:

```python
from IPython.display import Markdown, display
```

E exibindo o resultado formatado:

```python
response = agent.query(
    input="qual é a data estelar de hoje?"
)
display(Markdown(response["output"]))
```

Saída:

```
Diário de Bordo do Capitão, Suplementar. A data estelar atual é 48972.5.
```

## Conectando as peças

Agora que temos uma forma de consultar o sistema operacional, vamos juntar isso ao nosso agente para responder a perguntas reais sobre a saúde da máquina.

O primeiro passo é definir a função que fará as consultas via Osquery. Essa função será entregue como ferramenta ao agente:

```python
def call_osquery(query: str):
    """Consulta o sistema operacional usando o osquery.
      
      Esta função envia uma query ao processo do osquery para retornar informações sobre a máquina atual, sistema operacional e processos em execução.
      Você também pode usar esta função para consultar o banco SQLite subjacente e descobrir metadados da instância do osquery usando tabelas de sistema como sqlite_master, sqlite_temp_master e tabelas virtuais.

      Args:
        query: str  Uma query SQL para uma das tabelas do osquery (ex: "select timestamp from time")

      Returns:
        ExtensionResponse: resposta do osquery com o status da requisição e os dados da query se bem-sucedida.
    """
    return instance.client.query(query)
```

A função em si é bem concisa, mas o ponto crucial é a docstring detalhada, que permite ao modelo entender exatamente o propósito e as restrições da ferramenta.

Durante os meus testes, notei que o agente muitas vezes tentava acessar tabelas que não existiam no meu sistema. Por exemplo, no macOS a tabela `memory_info` não está presente.

Para dar mais contexto ao modelo, vamos listar dinamicamente todas as tabelas virtuais disponíveis no ambiente atual. Em um cenário ideal, forneceríamos o schema completo com colunas e tipos, mas apenas a lista de tabelas já ajuda enormemente.

Como o motor interno do Osquery utiliza SQLite, podemos consultar os nomes das tabelas virtuais em `sqlite_temp_master`:

```python
# Consulta as tabelas virtuais disponíveis nesta máquina
response = instance.client.query("select name from sqlite_temp_master").response
tables = [ t["name"] for t in response ]
```

Com a lista de tabelas em mãos, configuramos o agente com essas instruções e a ferramenta `call_osquery`:

```python
osagent = agent_engines.LangchainAgent(
    model = model,
    system_instruction=f"""
    Você é um agente que responde perguntas sobre a máquina em que está sendo executado.
    Você deve executar queries SQL usando uma ou mais tabelas disponíveis para responder às perguntas do usuário.
    Sempre retorne valores legíveis para humanos (ex: megabytes em vez de bytes, e tempo formatado em vez de milissegundos).
    Seja flexível na interpretação dos pedidos. Por exemplo, se o usuário pedir informações sobre aplicativos, você pode retornar processos e serviços. Se pedir consumo de recursos, retorne TANTO informações de memória QUANTO de CPU.
    Não faça perguntas de esclarecimento ao usuário.
    Você tem as seguintes tabelas disponíveis no sistema:
    ----- TABELAS -----
    {tables}
    ----- FIM DAS TABELAS -----

    Pergunta:
    """,
    tools=[
        call_osquery,
    ]
)
```

Com isso, nosso agente de diagnóstico está pronto! Vamos colocá-lo à prova:

```python
response = osagent.query(input="qual é a hora atual?")
display(Markdown(response["output"]))
```

Saída:

```
A hora atual é Fri May 30 18:08:15 2025 UTC.
```

Vamos testar uma consulta mais detalhada:

```python
response = osagent.query(input="qual processo está consumindo mais recursos?")
display(Markdown(response["output"]))
```

Saída:

```
O processo com maior consumo de CPU é:
Nome do Processo: mediaanalysisd, PID: 1127, Tempo Total de CPU: 2876716

O processo com maior consumo de memória é:
Nome do Processo: Code Helper (Plugin), PID: 2987, Uso de Memória: 1537 MB
```

E se pedirmos algo no espírito de Star Trek?

```python
response = osagent.query(input="computador, execute um procedimento de diagnóstico de nível 1")
display(Markdown(response["output"]))
```

Saída:

```
Aqui está o relatório do diagnóstico de nível 1:

**Sistema Operacional:**
*   **Nome:** macOS
*   **Versão:** 15.5
*   **Build:** 24F74
*   **Plataforma:** darwin
*   **Arquitetura:** arm64

... omitido para brevidade ...

**Tempo de Atividade (Uptime):**
*   **Total de Segundos:** 261164
*   **Dias:** 3
*   **Horas:** 0
*   **Minutos:** 32
*   **Segundos:** 44
```

O agente montou um relatório de diagnóstico completo e estruturado por conta própria, sem que eu precisasse codificar as regras de um diagnóstico de nível 1!

Também testei perguntas de segurança:

```python
response = osagent.query(input="computador, você vê algum sinal de malware em execução?")
display(Markdown(response["output"]))
```

Saída:

```
Verifiquei processos não associados a arquivos em disco (um indicador comum de malware) e nenhum foi encontrado. Também examinei os processos com maior consumo de memória e CPU. Os processos com maior uso são o Visual Studio Code e o Google Chrome, juntamente com seus processos auxiliares, o que representa um comportamento normal.

Com base nas verificações realizadas, não há sinais evidentes de malware em execução no sistema neste momento.
```

## Conclusões

Com pouquíssimas linhas de código, passamos de zero a uma interface em linguagem natural capaz de inspecionar o funcionamento interno do sistema operacional. Com alguns ajustes adicionais, esse agente pode realizar diagnósticos ainda mais profundos e até sugerir correções de forma autônoma. O Scotty ficaria orgulhoso!

![Engenheiro Scotty tentando falar com o computador usando o mouse como microfone](hello-computer-hello.gif)

Você pode conferir o código-fonte de todos os exemplos deste artigo no meu [GitHub](https://github.com/danicat/devrel/blob/main/blogs/20250531-diagnostic-agent/diagnostic_agent.ipynb).

Na próxima parte desta série, [Explorando a Fundo o SDK da Vertex AI para Python]({{< ref "/posts/20250605-vertex-ai-sdk-python" >}}), vamos entender como funciona o protocolo de comunicação entre o cliente e o Gemini, e implementar function calling manual em baixo nível.

O que achou da experiência? Compartilhe suas impressões nos comentários abaixo!
