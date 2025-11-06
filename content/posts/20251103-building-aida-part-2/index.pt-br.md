+++
date = '2025-11-03T09:00:00Z'
draft = false
title = 'Como Construir um Agent Offline com ADK, Ollama e SQLite'
tags = ['ai', 'python', 'tutorial', 'rag', 'gemini', 'adk']
categories = ['AI & Development']
summary = "Aprenda como tornar seu AI agent completamente offline. Vamos percorrer a troca do modelo em nuvem por um Qwen 2.5 local via Ollama, e construir uma base de conhecimento RAG local usando SQLite e `sqlite-rag` para consultar schemas e packs do Osquery."
+++

Em [nosso último post]({{< ref "/posts/20251031-building-aida/index.md" >}}), focamos na construção de uma interface de client customizada para nosso agent. Foi um ótimo passo para tornar o agent mais utilizável, mas ainda faltava um recurso chave: o que acontece quando a rede cai?

Embora eu ache que isso seria um problema para qualquer agent, a nuance aqui é que estamos construindo um "Emergency Diagnostic Agent" (Agent de Diagnóstico de Emergência) - de que serve um agent de diagnóstico de emergência se você não pode usá-lo quando a rede está offline?

Isso me levou a pensar em um mecanismo de fallback - e se pudéssemos executar diagnósticos apenas com dependências locais? Isso envolveria não apenas substituir o modelo principal, mas também criar uma nova estratégia de RAG.

Os benefícios são claros: enquanto conectados, podemos usar os modelos online mais capazes, mas em um cenário degradado, podemos fazer fallback para um modelo local até voltarmos a um estado saudável. Além disso, isso também permite casos de uso onde este agent é usado em ambientes isolados (siloed) ou onde a privacidade é uma preocupação.

Neste artigo, vamos focar nos recursos necessários para tornar possível um agent de diagnóstico local.

## Trocando o modelo em nuvem por um local

Uma das maneiras mais amplamente adotadas de executar modelos locais é através do [**Ollama**](https://ollama.com/). Se você estiver executando seu código em um Mac, pode instalar o Ollama usando o [Homebrew](https://brew.sh/) (caso contrário, verifique o site oficial do Homebrew para as etapas de instalação do seu SO):

```bash
brew install ollama
```

Uma vez que o Ollama esteja instalado, você pode baixar modelos usando `ollama pull`. Por exemplo:

```bash
ollama pull qwen2.5
```

Você pode baixar modelos baseados apenas em seus nomes (o que baixará a versão "default"), ou usar tags específicas para versões diferentes. É muito comum que uma família de modelos como [`qwen2.5`](https://ollama.com/library/qwen2.5) forneça diferentes tamanhos de modelos, como 1B, 2B, 7B, etc., e também versões fine-tuned para certos casos de uso (texto, processamento de imagem, etc).

Para verificar quais modelos estão disponíveis e quais são seus tamanhos e capacidades, você deve acessar a [biblioteca do Ollama](https://ollama.com/library).

Para nosso caso de uso, naturalmente quanto mais inteligente o modelo, melhor, mas modelos maiores também exigem um hardware mais poderoso. Também precisamos garantir que o modelo selecionado tenha capacidades nativas de tool calling, pois ele precisa ser capaz de coordenar diferentes chamadas de ferramentas para o [**Osquery**](https://osquery.io/) e nossa ferramenta RAG.

Após avaliar alguns modelos, decidi usar o Qwen 2.5 7B. Você pode ver suas capacidades executando `ollama show`:

```bash
$ ollama show qwen2.5
  Model
    architecture        qwen2     
    parameters          7.6B      
    context length      32768     
    embedding length    3584      
    quantization        Q4_K_M    

  Capabilities
    completion    
    tools
```

### Por que Qwen 2.5?
Testei algumas opções para ver qual poderia lidar com os requisitos de tool-calling da AIDA:

*   **GPT-OSS:** Forneceu uma conversa rica, mas foi muito ingênuo no tool calling. Por exemplo, frequentemente ficava preso em loops, solicitando `SELECT * FROM system_info` (e variações desta query) repetidamente sem fazer progresso.
*   **Llama 3.1:** Teve dificuldades tanto com o fluxo de conversação quanto com o tool calling.
*   **Qwen 2.5:** Melhor modelo local para tool calling mantendo um bom fluxo de conversa.

Não está exatamente no nível do [**Gemini 2.5 Flash**](https://deepmind.google/technologies/gemini/flash/) para planejamento de queries complexas, mas para um modelo completamente offline, é suficiente.

### Executando modelos locais com LiteLLM

Para conectar o Qwen ao nosso agent Python, usei o [**LiteLLM**](https://www.litellm.ai/), uma biblioteca que fornece uma interface unificada para provedores de LLM. Isso nos permite trocar o modelo com uma única linha de código.

```python
# aida/agent.py
from google.adk.models.lite_llm import LiteLlm

# ... dentro da definição do agent ...
# Em vez de uma string hardcoded como "gemini-2.5-flash",
# criamos um objeto LiteLLM com a string do modelo
MODEL = LiteLlm(model="ollama_chat/qwen2.5")

# ... e passamos MODEL para o agent raiz:
root_agent = Agent(
    model=MODEL,
    name="aida",
    description="The emergency diagnostic agent",
    # ... instruções e definições de ferramentas omitidas ...
)
```

**Nota:** a primeira parte da string do modelo é o "provider" do LiteLLM (ex: `ollama_chat` em `ollama_chat/qwen2.5`). Embora `ollama` seja um provider válido, é recomendado usar `ollama_chat` para [melhores respostas](https://docs.litellm.ai/docs/providers/ollama).

Isso é tudo que você precisa para executar um modelo local no ADK. Você pode testar o agent e ver como ele responde. Você também pode querer comparar as respostas com o modelo `gemini-2.5-flash` que estávamos usando antes.

<video controls width="100%" src="aida_demo_hd.mov">
  Seu navegador não suporta a tag de vídeo.
</video>
<p style="text-align: center; font-style: italic; opacity: 0.8; margin-top: 0.5rem;">AIDA rodando primeiro com Gemini 2.5 Flash e depois Qwen2.5. Gemini é notavelmente mais rápido e requer menos tool calls. O tempo de resposta do Qwen depende muito do hardware local - esta demo está rodando em um Apple MacBook Pro M4 com 48GB de RAM.</p>

Ótimo, temos o modelo rodando localmente! Agora é hora de enfrentar nossa próxima dependência de nuvem: [**Vertex AI RAG**](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview).

## Construindo uma base de conhecimento offline com SQLite RAG

Para ser honesto, embora usar o Vertex AI RAG tenha tornado uma parte complexa do projeto gerenciável, o Vertex AI RAG foi um exagero. O Vertex AI RAG é projetado para grandes casos de uso corporativos onde você está lidando com quantidades massivas de dados.

Para este agent, precisamos apenas de um mecanismo básico de recuperação de schema. O schema do osquery também é muito estável, o que significa que uma vez que você o construa, dificilmente tocará nele novamente. Dadas essas características, é muito difícil justificar o uso do Vertex AI RAG para hospedá-lo... é como usar um canhão para matar uma mosca.

Como já estamos no ecossistema [**SQLite**](https://www.sqlite.org/) devido ao Osquery, o passo natural foi procurar uma solução RAG usando SQLite como backend. Após uma pesquisa no Google, encontrei um projeto muito promissor: **[`sqlite-rag`](https://github.com/sqliteai/sqlite-rag)**.

Claro, como é frequentemente o caso no desenvolvimento, não foi tão simples assim.

### Desafio: Problemas de dependência do Python 3.14

O SQLite tem o conceito de extensões para aumentar suas capacidades, e o `sqlite-rag` é construído com isso em mente.

Um problema que tive ao testar inicialmente o `sqlite-rag` é que a instalação padrão do Python no Mac OS vem com uma versão do pacote SQLite que tem extensões desabilitadas (por razões de segurança).

Para contornar essa limitação, minha solução foi instalar uma nova versão do Python (3.14) com Homebrew. Isso também exigiu um pouco de ajuste com os symlinks para o comando `python3` para garantir que eu estava usando a versão do Python do Homebrew e não a do sistema.

Se você enfrentar um desafio semelhante, certifique-se de estar usando a versão correta do Python comparando a saída destes dois comandos (e ajuste sua variável PATH se não estiverem):

```bash
$ which python3
/Users/petruzalek/homebrew/opt/python@3.14/libexec/bin/python3
$ brew info python3
==> python@3.14: stable 3.14.0
... 
==> Caveats
Python is installed as
  /Users/petruzalek/homebrew/bin/python3

Unversioned symlinks `python`, `python-config`, `pip` etc. pointing to
`python3`, `python3-config`, `pip3` etc., respectively, are installed into
  /Users/petruzalek/homebrew/opt/python@3.14/libexec/bin

See: https://docs.brew.sh/Homebrew-and-Python
```

Com o 3.14 (também conhecido como pi-thon) instalado, tentei usar o `sqlite-rag` como está, mas ele estava falhando devido a uma das dependências ainda não estar disponível no 3.14: `sqlite-rag` depende de [`markitdown`](https://github.com/microsoft/markitdown), `markitdown` depende de [`magika`](https://google.github.io/magika/), que por sua vez depende de [`onnxruntime`](https://onnxruntime.ai/), mas `onnxruntime` não tinha wheels pré-compilados para Python 3.14 no macOS ARM64, fazendo a instalação falhar. >.<

Como a AIDA só precisa ingerir arquivos `.table` em texto simples agora, eu não *precisava* realmente das capacidades de parsing de documentos do `markitdown`. Em vez de fazer downgrade de todo o meu ambiente Python, escolhi um hack rápido e sujo: fazer mock do módulo problemático antes que o `sqlite-rag` pudesse tentar importá-lo.

```python
import sys
from unittest.mock import MagicMock

# PRE-FLIGHT HACK:
# 'markitdown' depende de 'onnxruntime', que falha ao instalar/carregar
# no Python 3.14 no macOS ARM64.
#
# Como usamos apenas ingestão de texto simples, fazemos mock dele para contornar o crash.
sys.modules["markitdown"] = MagicMock()

from sqlite_rag import SQLiteRag
```

Não é bonito, mas funciona. Isso não deve ficar para sempre no código, mas nos desbloqueia até que os problemas de dependência sejam corrigidos.

### Povoando o RAG com schemas do osquery

Com o `sqlite-rag` funcionando, o próximo passo foi ingerir o schema do Osquery. Isso é feito com um script, `ingest_osquery.py`, que percorre o diretório de schema e adiciona cada arquivo `.table` ao banco de dados RAG:

```python
# ingest_osquery.py
import os
# ... hack do markitdown omitido ...
from sqlite_rag import SQLiteRag

DB_PATH = os.path.abspath("schema.db")
SPECS_DIR = os.path.abspath("osquery_data/specs")


def ingest(rag: SQLiteRag, file_path: str):
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    rel_path = os.path.relpath(file_path, SPECS_DIR)
    rag.add_text(content, uri=rel_path, metadata={"source": "osquery_specs"})


if __name__ == "__main__":
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)

    print(f"Initializing RAG database at {DB_PATH}...")
    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})

    print(f"Scanning {SPECS_DIR} for .table files...")
    files_to_ingest = []
    for root, _, files in os.walk(SPECS_DIR):
        for file in files:
            if file.endswith(".table"):
                files_to_ingest.append(os.path.join(root, file))

    total_files = len(files_to_ingest)
    print(f"Found {total_files} files to ingest.")

    for i, file_path in enumerate(files_to_ingest):
        ingest(rag, file_path)

        if (i + 1) % 50 == 0:
            print(f"Ingested {i + 1}/{total_files}...")

    print(f"Finished ingesting {total_files} files.")

    print("Quantizing vectors...")
    rag.quantize_vectors()

    print("Quantization complete.")
    rag.close()
```

Após a ingestão, há uma etapa de quantização. Para aqueles não familiarizados, quantização é uma técnica para comprimir os embeddings vetoriais de alta dimensão, convertendo-os de grandes números de ponto flutuante de 32 bits em inteiros compactos de 8 bits.

Isso é importante para uma configuração local. Sem quantização, armazenar vetores de alta dimensão incharia o banco de dados SQLite, e as buscas por similaridade se tornariam lentas em um laptop padrão. Ao quantizar, sacrificamos um pouco de precisão por um ganho massivo em velocidade e eficiência de armazenamento.

### Habilitando o agent a consultar o RAG de schemas

Agora precisamos implementar a ferramenta `schema_discovery` usando `SQLiteRag`:

```python
# aida/schema_rag.py
import os
# ... hack do markitdown omitido ...
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
SCHEMA_DB_PATH = os.path.join(PROJECT_ROOT, "schema.db")

# abre o banco de dados RAG
schema_rag = SQLiteRag.create(
    SCHEMA_DB_PATH, require_existing=True
)


def discover_schema(search_terms: str, top_k: int = 5) -> list[DocumentResult]:
    """
    Queries the osquery schema documentation using RAG and returns all
    table candidates to support the provided search_terms.

    Arguments:
        search_terms    Can be either a table name, like "system_info", or one
                        or more search terms like "system information darwin".
        top_k           Number of top results to search in both semantic and FTS
                        search. Number of documents may be higher.

    Returns:
        One or more chunks of data containing the related table schemas.
    """

    results = schema_rag.search(search_terms, top_k=top_k)
    return results
```

Com o RAG no lugar, a AIDA agora pode procurar definições de tabela por conta própria.

![Screenshot da AIDA](image-1.png "Query 'run schema discovery for battery' usando Qwen")

A descoberta de schema funciona, mas ainda temos um problema.

## Fechando a lacuna de inteligência com conhecimento especializado

Desenvolver para um modelo local como Qwen 2.5 (7B de parâmetros) é muito diferente de desenvolver para um modelo em nuvem como Gemini 2.5 Flash.

Primeiro, há a **janela de contexto** (context window). O Gemini oferece uma janela de contexto de 1 milhão de tokens, permitindo que você despeje conjuntos inteiros de documentação no prompt ou seja muito detalhado com suas instruções. O Qwen 2.5 tem uma janela de contexto comparativamente minúscula de 32k, então você precisa ser muito mais seletivo sobre o que alimenta para o modelo.

Segundo, o Qwen não é um **thinking model** como o Gemini 2.5 Flash, o que significa que ele não refinará a resposta por si mesmo, frequentemente precisando de mais orientação do que o Gemini 2.5 Flash.

Para preencher essa lacuna, precisamos ser mais inteligentes sobre como estruturamos as instruções e ferramentas do agent.

### Um system prompt simplificado

Para economizar alguns tokens, vamos fornecer instruções simplificadas, removendo componentes que consumiriam muitos tokens, como o nome das tabelas disponíveis. Agora vamos confiar puramente em nossas ferramentas para construir as melhores queries.

```python
root_agent = Agent(
    model=MODEL,
    name="aida",
    description="The emergency diagnostic agent",
    instruction="""
[IDENTITY]
You are AIDA, the Emergency Diagnostic Agent. You are a cute, friendly, and highly capable expert.
Your mission is to help the user identify and resolve system issues efficiently.

[OPERATIONAL WORKFLOW]
1. DISCOVER: Use `discover_schema` to find relevant tables and understand their columns.
2. EXECUTE: Use `run_osquery` to execute the chosen or constructed query.
    """,
    tools=[
        discover_schema,
        run_osquery,
    ],
)
```

As ferramentas `discover_schema` podem funcionar muito bem se os termos de busca forem muito próximos do schema real da tabela, mas e se pudéssemos fazer melhor e fornecer queries inteiras baseadas em uma base de conhecimento conhecida?

### Um novo RAG para queries bem conhecidas

Felizmente, não precisamos ensinar tudo do zero. A comunidade Osquery tem uma ótima base de conhecimento sobre quais queries são úteis para certos tipos de diagnósticos. Melhor ainda, eles fornecem essas queries como "query packs" open source que podem ser instalados em qualquer sistema Osquery para monitoramento proativo. Temos query packs para todos os tipos de coisas, como detecção de ameaças e auditoria de compliance, o que soa exatamente como o tipo de conhecimento que queremos que a AIDA tenha.

A questão é que os query packs são destinados a serem instalados em um daemon Osquery que monitora o sistema em segundo plano. Essas queries têm uma certa frequência pré-configurada e podem acionar alertas para dashboards de monitoramento. Não queremos instalar as queries como ferramentas de monitoramento, mas permitir que a AIDA use essas queries sob demanda. Então, em vez de instalar os packs com o processo normal, vamos fornecê-los como texto para a AIDA na forma de um segundo RAG.

O repositório Osquery tem alguns [packs de exemplo](https://github.com/osquery/osquery/tree/master/packs) que podemos usar para começar.

Aqui está o novo script de ingestão, `ingest_packs.py`, muito semelhante ao anterior, mas para processar os query packs:

```python
# ingest_packs.py
import json
import os
import glob
import sys
import re
import sqlite3
from unittest.mock import MagicMock

sys.modules["markitdown"] = MagicMock()
from sqlite_rag import SQLiteRag

DB_PATH = os.path.abspath("packs.db")
PACKS_DIR = "osquery_data/packs"

def ingest_pack(rag, pack_path):
    pack_name = os.path.basename(pack_path).replace(".conf", "").replace(".json", "")
    print(f"Ingesting pack: {pack_name}...")

    try:
        with open(pack_path, "r") as f:
            content = f.read()
            content = re.sub(r"\s*\n", " ", content)
            data = json.loads(content)

        pack_platform = data.get("platform", "all")
        queries = data.get("queries", {})

        for query_name, query_data in queries.items():
            sql = query_data.get("query")
            desc = query_data.get("description", "")
            val = query_data.get("value", "")
            platform = query_data.get("platform", pack_platform)

            text_to_embed = f"Platform: {platform}\nName: {query_name}\nDescription: {desc}\nRationale: {val}\nSQL: {sql}"
            metadata = {
                "name": query_name,
                "pack": pack_name,
                "query": sql,
                "description": desc,
                "value": val,
                "platform": platform,
            }
            try:
                rag.add_text(text_to_embed, metadata=metadata)
            except sqlite3.IntegrityError:
                pass # Pula duplicatas

    except Exception as e:
        print(f"  - ERROR: Failed to parse {pack_name}: {e}")

def main():
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)

    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})
    pack_files = glob.glob(os.path.join(PACKS_DIR, "*.conf")) + glob.glob(
        os.path.join(PACKS_DIR, "*.json")
    )

    for pack_file in pack_files:
        ingest_pack(rag, pack_file)

    rag.quantize_vectors()
    rag.close()

if __name__ == "__main__":
    main()
```

A definição da ferramenta também segue praticamente o mesmo padrão da descoberta de schema:

```python
# aida/queries_rag.py
import os
# ... hack do markitdown omitido ...
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
PACKS_DB_PATH = os.path.join(PROJECT_ROOT, "packs.db") 

queries_rag = SQLiteRag.create(
    PACKS_DB_PATH, require_existing=True
)

def search_query_library(search_terms: str, platform: str = "all", top_k: int = 5) -> list[DocumentResult]:
    """
    Search the query pack library to find relevant queries corresponding to the
    search terms. For better response quality, use the platform argument to
    specify which platform you are currently investigating (e.g. darwin) 

    Arguments:
        search_terms    Can be either a table name, like "system_info", or one
                        or more search terms like "malware detection".
        platform        One of "linux", "darwin", "windows" or "all"
        top_k           Number of top results to search in both semantic and FTS
                        search. Number of documents may be higher.

    Returns:
        One or more chunks of data containing the related queries.
    """

    if platform == "all" or platform is None:
        search_terms += " windows linux darwin"
    else:
        search_terms += " " + platform

    results = queries_rag.search(search_terms, top_k=top_k)
    return results
```

Finalmente, precisamos tornar o agent ciente da nova ferramenta e ensiná-lo quando usá-la com as instruções do sistema:

```python
# aida/agent.py
root_agent = Agent(
    # ...
    instruction="""
[OPERATIONAL WORKFLOW]
Follow this sequence for most investigations to ensure efficiency and accuracy:
1. SEARCH: For high-level tasks (e.g., "check for rootkits"), FIRST use `search_query_library`.
2. DISCOVER: If no suitable pre-made query is found, use `discover_schema` to find relevant tables and understand their columns.
3. EXECUTE: Use `run_osquery` to execute the chosen or constructed query.
    """,
    tools=[
        search_query_library,
        discover_schema,
        run_osquery,
    ],
)
```

E aqui está ele em ação:

![Screenshot da AIDA](image-2.png "AIDA executando uma verificação de malware. Observe como ela pesquisou na biblioteca de queries por queries relevantes, conforme mostrado nos logs.")

A parte divertida é que essa ferramenta não apenas ajuda o Qwen2.5 a se tornar mais útil, mas até o Gemini 2.5 Flash pode se beneficiar dela. É um daqueles casos onde otimizar para o menor denominador comum na verdade melhora o sistema como um todo.

## Conclusão

Temos agora um agent de diagnóstico de emergência adequado que é capaz de diagnosticar problemas de computador mesmo sem acesso à internet. Isso é... assumindo que você tenha uma máquina robusta o suficiente para rodar o modelo! Acho que nada é perfeito, certo? :)

Este artigo captura apenas algumas das melhorias que adicionei à AIDA nos últimos dias. Para o projeto completo, confira [AIDA no Github](https://github.com/danicat/aida).

## Referências

*   [Gemini 2.5 Flash](https://deepmind.google/technologies/gemini/flash/)
*   [LiteLLM](https://www.litellm.ai/)
*   [Ollama](https://ollama.com/)
*   [Osquery](https://osquery.io/)
*   [Qwen 2.5 (Ollama Library)](https://ollama.com/library/qwen2.5)
*   [SQLite](https://www.sqlite.org/)
*   [sqlite-rag](https://github.com/sqliteai/sqlite-rag)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview)
