---
date: '2025-05-31T01:00:00+01:00'
title: 'Como transformei meu computador na "USS Enterprise" usando Agentes de IA'
summary: "Como criar um agente de diagnóstico que fala linguagem natural usando Gemini e Vertex AI Agent Engine"
categories: ["AI & Development"]
tags: ["gemini", "vertex-ai", "python", "tutorial"]
---
{{< translation-notice >}}
_Espaço: a fronteira final. Estas são as viagens da nave estelar Enterprise. Sua missão de 5 anos: explorar novos mundos estranhos; procurar novas vidas e novas civilizações; audaciosamente ir onde nenhum homem jamais esteve._

## Introdução

Enquanto crescia, graças à influência do meu pai, acostumei-me a ouvir essas palavras quase todos os dias. Suspeito que a paixão dele por Star Trek desempenhou um papel enorme na minha escolha pela carreira de engenharia de software. (Para aqueles que não estão familiarizados com Star Trek, este discurso era reproduzido no início de cada episódio da série original de Star Trek)

Star Trek sempre esteve à frente de seu tempo. Mostrou o [primeiro beijo inter-racial na televisão dos EUA](https://en.wikipedia.org/wiki/Kirk_and_Uhura%27s_kiss), em tempos em que tal cena causava muita controvérsia. Também retratou muitas peças de tecnologia “futurista” que hoje são commodities, como smartphones e videoconferência.

Uma coisa realmente notável é como os engenheiros da série interagem com os computadores. Embora vejamos alguns teclados e pressionamentos de botões de vez em quando, muitos dos comandos são vocalizados em linguagem natural. Alguns dos comandos que eles dão ao computador são bastante icônicos, como por exemplo, quando solicitam ao computador para executar um “procedimento de diagnóstico de nível 1”, o que aconteceu tantas vezes que praticamente se tornou [uma piada](https://www.youtube.com/watch?v=cYzByQjzTb0) entre os fãs mais assíduos.

Avançando mais de 30 anos e aqui estamos nós, na Era da IA, uma revolução tecnológica que promete ser maior que a internet. Claro que muitas pessoas estão com medo de como a IA pode impactar seus empregos ([escrevi sobre isso na semana passada](https://danicat.dev/posts/20250528-vibe-coding/)), mas crescer assistindo Star Trek torna mais fácil para mim ver como o papel do engenheiro mudará nos próximos anos. Em vez de comandar o computador por meio de texto, instruindo manualmente cada etapa do caminho por meio de linhas de código e compiladores, muito em breve passaremos a conversar e fazer brainstorming com nossos computadores.

Para ajudar as pessoas a visualizar isso, vamos usar a tecnologia que temos hoje para criar um pequeno agente que nos permite interagir com nossas próprias máquinas usando linguagem natural.

## O que você precisará para esta demonstração

Para a linguagem de desenvolvimento, usaremos Python em um Jupyter Notebook, pois ele funciona muito bem para experimentação. As principais ferramentas e bibliotecas que usaremos são:

*   [Vertex AI Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/overview?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog)
*   [Osquery](https://www.osquery.io/) com [bindings para Python](https://github.com/osquery/osquery-python)
*   [Jupyter Notebook](https://jupyter.org/) [opcional] (na verdade, estou usando o [plugin Jupyter para VSCode](https://code.visualstudio.com/docs/datascience/jupyter-notebooks))

Os exemplos abaixo usarão o Gemini Flash 2.0, mas você pode usar qualquer [variante do modelo Gemini](https://ai.google.dev/gemini-api/docs/models). Não implantaremos este agente no Google Cloud desta vez, pois queremos usá-lo para responder a perguntas sobre a máquina local e não sobre o servidor na nuvem.

## Visão Geral do Agente

Se você já está familiarizado com o funcionamento da tecnologia de agentes, pode pular esta seção.

Um agente de IA é uma forma de IA capaz de perceber seu ambiente e tomar ações autônomas para atingir objetivos específicos. Se comparado com os típicos Modelos de Linguagem Grande (LLMs), que se concentram principalmente na geração de conteúdo com base na entrada, os agentes de IA podem interagir com seu ambiente, tomar decisões e executar tarefas para atingir seus objetivos. Isso é alcançado pelo uso de “ferramentas” que alimentarão o agente com informações e permitirão que ele realize ações.

Para demonstrar a tecnologia de agente, usaremos o LangChain por meio do [Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog). Primeiro, você precisa instalar os pacotes necessários em seu sistema:

```shell
pip install --upgrade --quiet google-cloud-aiplatform[agent_engines,langchain]
```

Você também precisará definir suas credenciais padrão de aplicativo (ADC) do gcloud:

```shell
gcloud auth application-default login
```

Nota: dependendo do ambiente que você está usando para executar esta demonstração, pode ser necessário usar um método de autenticação diferente.

Agora estamos prontos para trabalhar em nosso script Python. Primeiro, vamos inicializar o SDK com base no ID e local do nosso projeto do Google Cloud:

```python
import vertexai

vertexai.init(
    project="my-project-id",                  # Seu ID de projeto.
    location="us-central1",                   # Sua localização na nuvem.
    staging_bucket="gs://my-staging-bucket",  # Seu bucket de preparo.
)
```

Uma vez feita a configuração inicial, criar um agente usando LangChain no Agent Engine é bastante simples:

```python
from vertexai import agent_engines

model = "gemini-2.0-flash" # sinta-se à vontade para experimentar diferentes modelos!

model_kwargs = {
    # temperature (float): A temperatura de amostragem controla o grau de
    # aleatoriedade na seleção de tokens.
    "temperature": 0.20,
}

agent = agent_engines.LangchainAgent(
    model=model,                # Obrigatório.
    model_kwargs=model_kwargs,  # Opcional.
)
```

A configuração acima é suficiente para você enviar consultas ao agente, assim como enviaria uma consulta a um LLM:

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

O que poderia retornar algo assim:

```
{'input': 'which time is now?', 'output': 'Como uma IA, eu não tenho uma hora ou local "atuais" da mesma forma que um humano. Meu conhecimento não é atualizado em tempo real.\n\nPara saber a hora atual, você pode:\n\n*   **Verificar seu dispositivo:** Seu computador, telefone ou tablet exibirá a hora atual.\n*   **Fazer uma pesquisa rápida:** Digite "que horas são" em um mecanismo de busca como o Google.'}
```

Dependendo de suas configurações, prompt e da aleatoriedade do universo, o modelo pode lhe dar uma resposta dizendo que não pode lhe dizer a hora, ou pode “alucinar” e inventar um timestamp. Mas, na verdade, como a IA não tem relógio, ela não será capaz de responder a essa pergunta... a menos que você lhe dê um relógio!

## Chamadas de Função

Uma das maneiras mais convenientes de estender as capacidades do nosso agente é dar-lhe funções Python para chamar. O processo é bastante simples, mas é importante ressaltar que quanto melhor a documentação que você tiver para a função, mais fácil será para o agente acertar sua chamada. Vamos definir nossa função para verificar a hora:

```python
import datetime

def get_current_time():
    """Retorna a hora atual como um objeto datetime.

    Args:
        Nenhum

    Returns:
        datetime: hora atual como um tipo datetime
    """
    return datetime.datetime.now()
```

Agora que temos uma função que nos dá a hora do sistema, vamos recriar o agente, mas agora ciente de que a função existe:

```python
agent = agent_engines.LangchainAgent(
    model=model,                # Obrigatório.
    model_kwargs=model_kwargs,  # Opcional.
    tools=[get_current_time]
)
```

E faça a pergunta novamente:

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

A saída será semelhante a esta:

```
{'input': 'which time is now?', 'output': 'A hora atual é 18:36:42 UTC de 30 de maio de 2025.'}
```

Agora o agente pode contar com a ferramenta para responder à pergunta com dados reais. Muito legal, hein?

## Coletando Informações do Sistema

Para nosso agente de diagnóstico, vamos dar a ele recursos para consultar informações sobre a máquina em que está sendo executado usando uma ferramenta chamada [osquery](https://www.osquery.io/). Osquery é uma ferramenta de código aberto desenvolvida pelo Facebook para permitir que o usuário faça consultas SQL a “tabelas virtuais” que expõem informações sobre o sistema operacional subjacente da máquina.

Isso é conveniente para nós porque não apenas nos dá um único ponto de entrada para fazer consultas sobre o sistema, mas os LLMs também são muito proficientes em escrever consultas SQL.

Você pode encontrar instruções sobre como instalar o osquery na [documentação oficial](https://osquery.readthedocs.io/en/stable/). Não vou reproduzi-las aqui porque elas variam dependendo do sistema operacional da sua máquina.

Depois de instalar o osquery, você precisará instalar os bindings Python para o osquery. Como é típico em Python, é apenas um `pip install`:

```shell
pip install --upgrade --quiet osquery
```

Com os bindings instalados, você pode fazer chamadas ao osquery importando o pacote `osquery`:

```python
import osquery

# Inicia um processo osquery usando um soquete de extensão efêmero.
instance = osquery.SpawnInstance()
instance.open()  # Isso pode levantar uma exceção

# Emite consultas e chama APIs Thrift do osquery.
instance.client.query("select timestamp from time")
```

O método `query` retornará um objeto `ExtensionResponse` com os resultados de sua consulta. Por exemplo:

```python
ExtensionResponse(status=ExtensionStatus(code=0, message='OK', uuid=0), response=[{'timestamp': 'Sex Mai 30 17:54:06 2025 UTC'}])
```

Se você nunca trabalhou com o osquery antes, encorajo você a dar uma olhada no [schema](https://www.osquery.io/schema/5.17.0/) para ver que tipo de informação está disponível em seu sistema operacional.

### Uma nota lateral sobre formatação

Todas as saídas dos exemplos anteriores não foram formatadas, mas se você estiver executando o código do Jupyter, poderá acessar alguns métodos de conveniência para embelezar a saída importando os seguintes pacotes:

```python
from IPython.display import Markdown, display
```

E exibindo a saída da resposta como markdown:

```python
response = agent.query(
    input="what is today's stardate?"
)
display(Markdown(response["output"]))
```

Saída:

```
Diário do Capitão, Suplementar. A data estelar atual é 48972.5.
```

## Conectando os pontos

Agora que temos uma maneira de consultar informações sobre o sistema operacional, vamos combinar isso com nosso conhecimento de agentes para criar um agente de diagnóstico que responderá a perguntas sobre nosso sistema.

O primeiro passo é definir uma função para fazer as consultas. Isso será dado ao agente como uma ferramenta para coletar informações posteriormente:

```python
def call_osquery(query: str):
    """Consulta o sistema operacional usando osquery

      Esta função é usada para enviar uma consulta ao processo osquery para retornar informações sobre a máquina atual, sistema operacional e processos em execução.
      Você também pode usar esta função para consultar o banco de dados SQLite subjacente para descobrir mais informações sobre a instância do osquery usando tabelas do sistema como sqlite_master, sqlite_temp_master e tabelas virtuais.

      Args:
        query: str  Uma consulta SQL para uma das tabelas do osquery (por exemplo, "select timestamp from time")

      Returns:
        ExtensionResponse: uma resposta do osquery com o status da solicitação e uma resposta à consulta, se bem-sucedida.
    """
    return instance.client.query(query)
```

A função em si é bastante trivial, mas a parte importante aqui é ter um docstring bem detalhado que permitirá ao agente entender como essa função funciona.

Durante meus testes, um problema complicado que ocorreu com bastante frequência foi que o agente não sabia exatamente quais tabelas estavam disponíveis em meu sistema. Por exemplo, estou executando uma máquina macOS e a tabela “memory_info” não existe.

Para dar ao agente um pouco mais de contexto, vamos fornecer dinamicamente os nomes das tabelas que estão disponíveis neste sistema. Em uma situação ideal, você até daria a ele o schema inteiro com nomes de colunas e descrições, mas infelizmente isso não é trivial de se conseguir com o osquery.

A tecnologia de banco de dados subjacente para o osquery é o SQLite, então podemos consultar a lista de tabelas virtuais da tabela `sqlite_temp_master`:

```python
# use um pouco de mágica Python para descobrir quais tabelas temos neste sistema
response = instance.client.query("select name from sqlite_temp_master").response
tables = [ t["name"] for t in response ]
```

Agora que temos todos os nomes das tabelas, podemos criar o agente com esta informação e a ferramenta `call_osquery`:

```python
osagent = agent_engines.LangchainAgent(
    model = model,
    system_instruction=f"""
    Você é um agente que responde a perguntas sobre a máquina em que está sendo executado.
    Você deve executar consultas SQL usando uma ou mais das tabelas para responder às perguntas do usuário.
    Sempre retorne valores legíveis por humanos (por exemplo, megabytes em vez de bytes e hora formatada em vez de milissegundos)
    Seja muito flexível em sua interpretação das solicitações. Por exemplo, se o usuário solicitar informações do aplicativo, é aceitável retornar informações sobre processos e serviços. Se o usuário solicitar o uso de recursos, retorne AMBAS as informações de memória e CPU.
    Não peça esclarecimentos ao usuário.
    Você tem as seguintes tabelas disponíveis para você:
    ----- TABELAS -----
    {tables}
    ----- FIM TABELAS -----

    Pergunta:
    """,
    tools=[
        call_osquery,
    ]
)
```

Acredite ou não, nosso agente está pronto para todas as nossas perguntas! Vamos experimentar:

```python
response = osagent.query(input="what is the current time?")
display(Markdown(response["output"]))
```

Saída:

```
A hora atual é Sex Mai 30 18:08:15 2025 UTC.
```

Vamos complicar um pouco mais:

```python
response = osagent.query(input="what is the top consuming process?")
display(Markdown(response["output"]))
```

Saída:

```
O processo que mais consome CPU é:
Nome do Processo: mediaanalysisd, PID: 1127, Tempo Total de CPU: 2876716

O processo que mais consome memória é:
Nome do Processo: Code Helper (Plugin), PID: 2987, Uso de Memória: 1537 MB
```

E se eu for um pouco mais criativo?

```python
response = osagent.query(input="computer, run a level 1 diagnostic procedure")
display(Markdown(response["output"]))
```

Saída:

```
Aqui estão as informações de diagnóstico de nível 1:

**Sistema Operacional:**
*   **Nome:** macOS
*   **Versão:** 15.5
*   **Build:** 24F74
*   **Plataforma:** darwin
*   **Arquitetura:** arm64

... omitido por questões de brevidade ...

**Tempo de Atividade:**
*   **Segundos Totais:** 261164
*   **Dias:** 3
*   **Horas:** 0
*   **Minutos:** 32
*   **Segundos:** 44

```

Uau! Eu nunca defini o que é um procedimento de diagnóstico de Nível 1, mas ele me deu um relatório bastante impressionante mesmo assim!

Até tentei ser um pouco mais esperto nas minhas perguntas e as respostas não decepcionaram (na maioria das vezes):

```python
response = osagent.query(input="computer, do you see any signs of malware running?")
display(Markdown(response["output"]))
```

Saída:

```
Verifiquei se há processos que não estão associados a um arquivo em disco, o que pode ser um sinal de malware, e não encontrei nenhum. Também examinei os principais processos por uso de memória e CPU. Os processos que consomem mais recursos são principalmente o Visual Studio Code e o Google Chrome e seus processos auxiliares relacionados. Este é um comportamento típico para esses aplicativos.

Com base nas verificações realizadas, não há sinais óbvios de malware em execução no sistema neste momento.
```

_Mic drop_ =^.^=

## Conclusões

Eu sei que já é um argumento batido, mas a IA é um divisor de águas. Com pouquíssimas linhas de código, passamos do zero para uma interface de linguagem natural totalmente funcional para o funcionamento interno do sistema operacional. Com um pouco mais de trabalho, este agente pode ser aprimorado para fazer diagnósticos mais profundos e talvez até mesmo consertar coisas autonomamente. Scotty ficaria orgulhoso!

![Engenheiro Scotty tentando falar com o computador usando o mouse como microfone](hello-computer-hello.gif)

Você pode encontrar o código-fonte de todos os exemplos deste artigo no meu [GitHub](https://github.com/danicat/devrel/blob/main/blogs/20250531-diagnostic-agent/diagnostic_agent.ipynb).

Quais são suas impressões? Compartilhe suas ideias nos comentários abaixo.
