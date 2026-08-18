---categories:
- Agent Development
date: '2025-11-03T09:00:00Z'
series:
- Building the Diagnostic Agent
series_order: 6
summary: Aprenda a tornar seu agente de IA totalmente offline. Substituímos o modelo em nuvem por um modelo local Qwen 2.5 via Ollama e criamos uma base de conhecimento RAG local usando SQLite e sqlite-rag para consultar schemas e packs do Osquery.
tags:
  - adk
  - gemini
  - ollama
  - python
  - rag
  - sqlite
  - tutorial
title: "Como Construir um Agente Offline com ADK, Ollama e SQLite"
slug: "building-aida-part-2"
aliases:
  - "/pt-br/posts/20251103-building-aida-part-2/"
description: "Guia passo a passo para construir um agente de IA 100% offline usando o Google ADK, modelo local Qwen 2.5 via Ollama e LiteLLM, e RAG vetorial no SQLite com sqlite-rag."
proficiencyLevel: "Advanced"
dependencies:
  - "Python 3.11+"
  - "Google ADK"
  - "Ollama"
  - "Qwen 2.5"
  - "SQLite"
  - "sqlite-rag"
---
[Na Parte 5 desta série]({{< ref "/posts/20251031-building-aida" >}}), focamos na criação de uma interface de cliente personalizada para o nosso agente. Foi um grande passo para tornar o agente mais utilizável, mas ainda faltava um recurso fundamental: o que acontece quando a rede cai?

Embora eu ache que isso seja um problema para qualquer agente, a nuance aqui é que estamos construindo um "Agente de Diagnóstico de Emergência" — de que adianta um assistente de emergência se você não consegue usá-lo quando a rede está offline?

Isso me levou a pensar em um mecanismo de fallback: e se pudéssemos rodar diagnósticos usando apenas dependências locais? Isso envolveria não apenas substituir o modelo central, mas também arquitetar uma nova estratégia de RAG.

Os benefícios são claros: enquanto conectados, podemos usar os modelos online mais inteligentes, mas em um cenário degradado, podemos fazer o fallback para um modelo local até voltarmos a um estado saudável. Além disso, isso também viabiliza casos de uso onde o agente é executado em ambientes isolados (air-gapped) ou onde a privacidade é uma preocupação primordial.

Neste artigo, vamos nos concentrar nos recursos necessários para tornar um agente de diagnóstico local possível.

## Substituindo o modelo em nuvem por um local

Uma das formas mais adotadas para rodar modelos locais é através do [**Ollama**](https://ollama.com/). Se você estiver rodando seu código em um Mac, pode instalar o Ollama usando o [Homebrew](https://brew.sh/) (caso contrário, consulte o site oficial do Homebrew ou do Ollama para os passos de instalação no seu sistema operacional):

```bash
brew install ollama
```

Depois que o Ollama estiver instalado, você pode baixar modelos usando `ollama pull`. Por exemplo:

```bash
ollama pull qwen2.5
```

Você pode baixar modelos apenas pelo nome (o que puxará a versão "padrão"), ou usar tags específicas para versões diferentes. É muito comum que uma família de modelos como [`qwen2.5`](https://ollama.com/library/qwen2.5) ofereça modelos de tamanhos variados, como 1B, 2B, 7B, etc., além de versões com fine-tuning para certos casos de uso (texto, processamento de imagem, etc.).

Para verificar quais modelos estão disponíveis e quais são seus tamanhos e capacidades, acesse a [biblioteca do Ollama](https://ollama.com/library).

Para o nosso caso de uso, naturalmente quanto mais inteligente o modelo, melhor, mas modelos maiores também exigem um hardware mais potente. Também precisamos garantir que o modelo selecionado tenha capacidades nativas de tool calling (chamada de ferramentas), já que ele precisa ser capaz de coordenar diferentes chamadas de ferramentas para o [**osquery**](https://osquery.io/) e para nossa ferramenta de RAG.

Após avaliar alguns modelos, decidi usar o Qwen 2.5 7B. Você pode ver as capacidades dele executando `ollama show`:

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

### Por que o Qwen 2.5?

Testei algumas opções para ver quais conseguiam lidar com os requisitos de chamadas de ferramentas da AIDA:

*   **GPT-OSS:** Forneceu uma conversa rica, mas foi muito ingênuo no tool calling. Por exemplo, frequentemente ficava preso em loops, solicitando `SELECT * FROM system_info` (e variações dessa query) repetidamente sem fazer progresso.
*   **Llama 3.1:** Teve dificuldades tanto com o fluxo de conversa quanto com o tool calling.
*   **Qwen 2.5:** O melhor modelo local para chamadas de ferramentas, mantendo um bom fluxo de conversa.

Não está bem no nível do [**Gemini 2.5 Flash**](https://deepmind.google/technologies/gemini/flash/) para planejamento de queries complexas, mas para um modelo completamente offline, é suficiente.

### Executando modelos locais com LiteLLM

Para conectar o Qwen ao agente, usamos o [**LiteLLM**](https://www.litellm.ai/), uma biblioteca que oferece uma interface unificada para provedores de LLM. Isso nos permite trocar o modelo com uma única linha de código:

```python
# aida/agent.py
from google.adk.models.lite_llm import LiteLlm

# ... dentro da definição do agente ...
# Em vez de uma string literal como "gemini-2.5-flash",
# criamos um objeto LiteLLM com a string do modelo
MODEL = LiteLlm(model="ollama_chat/qwen2.5")

# ... e passamos o MODEL para o agente raiz:
root_agent = Agent(
    model=MODEL,
    name="aida",
    description="The emergency diagnostic agent",
    # ... instruções e definições de ferramentas omitidas ...
)
```

**Nota:** a primeira parte da string do modelo é o "provedor" do LiteLLM (ex: `ollama_chat` em `ollama_chat/qwen2.5`). Embora `ollama` seja um provedor válido, é recomendável usar `ollama_chat` para [obter respostas melhores](https://docs.litellm.ai/docs/providers/ollama).

Isso é tudo o que você precisa para rodar um modelo local no ADK. Você pode testar o agente e ver como ele responde. Também vale a pena comparar as respostas com o modelo `gemini-2.5-flash` que estávamos usando antes.

<video controls width="100%" src="aida_demo_hd.mov">
  Seu navegador não suporta a tag de vídeo.
</video>
<p style="text-align: center; font-style: italic; opacity: 0.8; margin-top: 0.5rem;">AIDA executando primeiro com Gemini 2.5 Flash e depois com Qwen 2.5. O Gemini é visivelmente mais rápido e exige menos chamadas de ferramentas. O tempo de resposta do Qwen é altamente dependente do hardware local — esta demonstração está rodando em um Apple MacBook Pro M4 com 48GB de RAM.</p>

Maravilha, já temos o modelo rodando localmente! Agora é hora de resolver nossa próxima dependência de nuvem: o [**Vertex AI RAG**](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview).

## Criando uma base de conhecimento offline com SQLite RAG

Para ser sincera, embora usar o Vertex AI RAG tenha tornado uma parte complexa do projeto gerenciável, ele era um tremendo exagero. O Vertex AI RAG foi projetado para grandes casos de uso corporativos, onde você lida com volumes massivos de dados.

Para o nosso agente, precisamos apenas de um mecanismo básico de recuperação de schemas. Além disso, o schema do osquery é muito estável — depois de construído, você dificilmente precisará mexer nele de novo. Com essas características, é muito difícil justificar o uso do Vertex AI RAG para hospedá-lo... é como usar um canhão para matar uma mosca.

Como já estamos no ecossistema do [**SQLite**](https://www.sqlite.org/) por causa do Osquery, o passo natural foi procurar uma solução de RAG usando o SQLite como backend. Após uma pesquisa no Google, encontrei um projeto muito promissor: **[`sqlite-rag`](https://github.com/sqliteai/sqlite-rag)**.

É claro que, como costuma acontecer no desenvolvimento, a coisa não foi tão simples assim.

### Desafio: Problemas de dependência do Python 3.14

O SQLite possui o conceito de extensões para ampliar suas capacidades, e o `sqlite-rag` foi construído com isso em mente.

Um problema que tive ao testar o `sqlite-rag` inicialmente foi que a instalação padrão do Python no macOS vem com uma versão do SQLite com extensões desabilitadas (por motivos de segurança).

Para contornar essa limitação, minha solução foi instalar uma nova versão do Python (3.14) via Homebrew. Isso também exigiu um pequeno ajuste nos symlinks do comando `python3` para garantir que eu estivesse usando a versão do Homebrew e não a do sistema.

Se você enfrentar um desafio semelhante, certifique-se de estar usando a versão correta do Python comparando a saída destes dois comandos (e ajuste sua variável PATH se não estiverem alinhadas):

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

Com o 3.14 (também conhecido carinhosamente como pi-thon) instalado, tentei usar o `sqlite-rag` como estava, mas ele falhou devido a uma das dependências que ainda não estava disponível para o 3.14: o `sqlite-rag` depende do [`markitdown`](https://github.com/microsoft/markitdown), o `markitdown` depende do [`magika`](https://google.github.io/magika/), que por sua vez depende do [`onnxruntime`](https://onnxruntime.ai/). Acontece que o `onnxruntime` ainda não tinha wheels pré-compilados para o Python 3.14 no macOS ARM64, fazendo com que a instalação quebrasse. >.<

Como a AIDA só precisa fazer a ingestão de arquivos `.table` em texto puro no momento, eu não *precisava* de verdade das capacidades de parsing de documentos do `markitdown`. Em vez de fazer o downgrade de todo o meu ambiente Python, escolhi um hack rápido e direto: fazer um mock do módulo problemático antes que o `sqlite-rag` tentasse importá-lo.

```python
import sys
from unittest.mock import MagicMock

# HACK PRÉ-VOO:
# O 'markitdown' depende do 'onnxruntime', que falha ao instalar/carregar
# no Python 3.14 em macOS ARM64.
#
# Como usamos apenas ingestão de texto puro, fazemos o mock dele para contornar o crash.
sys.modules["markitdown"] = MagicMock()

from sqlite_rag import SQLiteRag
```

Não é bonito, mas funciona. Isso não deve ficar para sempre no código, mas nos desbloqueia até que os problemas de dependência sejam corrigidos.

### Populando o RAG com os schemas do osquery

Com o `sqlite-rag` funcionando, o próximo passo foi fazer a ingestão do schema do Osquery. Isso é feito com um script, `ingest_osquery.py`, que percorre o diretório de schemas e adiciona cada arquivo `.table` ao banco RAG:

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

    print(f"Inicializando banco de dados RAG em {DB_PATH}...")
    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})

    print(f"Escaneando {SPECS_DIR} em busca de arquivos .table...")
    files_to_ingest = []
    for root, _, files in os.walk(SPECS_DIR):
        for file in files:
            if file.endswith(".table"):
                files_to_ingest.append(os.path.join(root, file))

    total_files = len(files_to_ingest)
    print(f"Encontrados {total_files} arquivos para ingestão.")

    for i, file_path in enumerate(files_to_ingest):
        ingest(rag, file_path)

        if (i + 1) % 50 == 0:
            print(f"Ingeridos {i + 1}/{total_files}...")

    print(f"Ingestão de {total_files} arquivos concluída.")

    print("Quantizando vetores...")
    rag.quantize_vectors()

    print("Quantização concluída.")
    rag.close()
```

Após a ingestão, há uma etapa de quantização. Para quem não está familiarizado, quantização é uma técnica para comprimir os embeddings vetoriais de alta dimensão, convertendo-os de grandes números de ponto flutuante de 32 bits em inteiros compactos de 8 bits.

Isso é importante para um setup local. Sem a quantização, armazenar vetores de alta dimensão inflaria o banco de dados SQLite, e as buscas por similaridade ficariam lentas em um notebook comum. Ao quantizar, sacrificamos um pouco de precisão em troca de um ganho massivo de velocidade e eficiência de armazenamento.

### Permitindo que o agente consulte o RAG de schemas

Agora precisamos implementar a ferramenta `discover_schema` usando o `SQLiteRag`:

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
    Consulta a documentação de schemas do osquery usando RAG e retorna
    todas as tabelas candidatas para atender aos termos de busca fornecidos.

    Argumentos:
        search_terms    Pode ser o nome de uma tabela, como "system_info", ou
                        um ou mais termos de busca como "system information darwin".
        top_k           Número de melhores resultados a buscar tanto na busca
                        semântica quanto no FTS. O número de documentos pode ser maior.

    Retorno:
        Um ou mais pedaços de dados contendo os schemas das tabelas relacionadas.
    """

    results = schema_rag.search(search_terms, top_k=top_k)
    return results
```

Com o RAG configurado, a AIDA agora consegue consultar definições de tabelas por conta própria.

![Captura de tela da AIDA](image-1.png "Consulta 'run schema discovery for battery' usando o Qwen")

A descoberta de schemas funciona, mas ainda temos um problema.

## Fechando o gap de inteligência com conhecimento especializado

Desenvolver para um modelo local como o Qwen 2.5 (7B parâmetros) é bem diferente de desenvolver para um modelo em nuvem como o Gemini 2.5 Flash.

Primeiro, há a **janela de contexto**. O Gemini oferece uma janela de contexto de 1 milhão de tokens, permitindo despejar conjuntos inteiros de documentação no prompt ou ser bastante prolixo nas suas instruções. O Qwen 2.5 tem uma janela comparativamente minúscula de 32k tokens, então você precisa ser muito mais seletivo sobre o que alimenta para o modelo.

Segundo, o Qwen não é um **modelo de pensamento (thinking model)** como o Gemini 2.5 Flash, o que significa que ele não refinará a resposta sozinho, muitas vezes precisando de mais direcionamento do que o Gemini 2.5 Flash.

Para preencher essa lacuna, precisamos ser mais inteligentes sobre como estruturamos as instruções e ferramentas do agente.

### Um system prompt simplificado

Para economizar alguns tokens, vamos fornecer instruções simplificadas, removendo componentes que consumiriam muitos tokens, como o nome de todas as tabelas disponíveis. Agora vamos confiar exclusivamente nas nossas ferramentas para construir as melhores queries.

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

A ferramenta `discover_schema` funciona muito bem se os termos de busca forem muito próximos do schema real da tabela, mas e se pudéssemos ir além e fornecer queries completas baseadas em uma base de conhecimento conhecida?

### Um novo RAG para queries consagradas

Felizmente, não precisamos ensinar tudo do zero. A comunidade do Osquery possui uma excelente base de conhecimento sobre quais queries são úteis para determinados tipos de diagnósticos. Melhor ainda, eles fornecem essas queries como "query packs" open source que podem ser instalados em qualquer sistema com Osquery para monitoramento proativo. Temos query packs para todo tipo de coisa, como detecção de ameaças e auditoria de conformidade — exatamente o tipo de conhecimento que queremos que a AIDA tenha.

A questão é que os query packs foram feitos para serem instalados em um daemon do Osquery que monitora o sistema em segundo plano. Essas queries têm uma frequência pré-configurada e podem disparar alertas em painéis de monitoramento. Não queremos instalar as queries como ferramentas de monitoramento contínuo, mas sim permitir que a AIDA use essas queries sob demanda. Então, em vez de instalar os packs pelo processo normal, vamos entregá-los como texto para a AIDA na forma de um segundo RAG.

O repositório do Osquery possui alguns [exemplos de packs](https://github.com/osquery/osquery/tree/master/packs) que podemos usar para começar.

Aqui está o novo script de ingestão, `ingest_packs.py`, muito parecido com o anterior, mas para processar os query packs:

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
    print(f"Ingerindo pack: {pack_name}...")

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
                pass # Ignora duplicatas

    except Exception as e:
        print(f"  - ERRO: Falha ao processar {pack_name}: {e}")

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

A definição da ferramenta também segue praticamente o mesmo padrão da descoberta de schemas:

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
    Pesquisa na biblioteca de query packs para encontrar queries relevantes correspondentes aos
    termos de busca. Para melhor qualidade de resposta, use o argumento platform para
    especificar qual plataforma você está investigando atualmente (ex: darwin).

    Argumentos:
        search_terms    Pode ser o nome de uma tabela, como "system_info", ou um
                        ou mais termos de busca como "malware detection".
        platform        Um entre "linux", "darwin", "windows" ou "all".
        top_k           Número de melhores resultados a buscar tanto na busca semântica
                        quanto no FTS. O número de documentos pode ser maior.

    Retorno:
        Um ou mais pedaços de dados contendo as queries relacionadas.
    """

    if platform == "all" or platform is None:
        search_terms += " windows linux darwin"
    else:
        search_terms += " " + platform

    results = queries_rag.search(search_terms, top_k=top_k)
    return results
```

Por fim, precisamos ensinar o agente sobre a nova ferramenta e instruí-lo sobre quando utilizá-la por meio das instruções de sistema:

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

E aqui está ela em ação:

![Captura de tela da AIDA](image-2.png "AIDA executando uma verificação de malware. Observe como ela buscou queries relevantes na biblioteca de packs, conforme mostrado nos logs.")

A parte divertida é que essa ferramenta não apenas ajuda o Qwen 2.5 a se tornar muito mais útil, mas o próprio Gemini 2.5 Flash também se beneficia dela. É um daqueles casos em que otimizar para o menor denominador comum acaba melhorando o sistema como um todo.

## Conclusão

Agora temos um agente de diagnóstico de emergência completo, capaz de diagnosticar problemas no computador mesmo sem acesso à internet. Isso... supondo que você tenha uma máquina potente o suficiente para rodar o modelo! Imagino que nada seja perfeito, certo? :)

Este artigo captura apenas algumas das melhorias que adicionei à AIDA ao longo dos últimos dias. Para ver o projeto completo, confira o [código da AIDA no GitHub](https://github.com/danicat/aida).

## Referências

*   [Gemini 2.5 Flash](https://deepmind.google/technologies/gemini/flash/)
*   [LiteLLM](https://www.litellm.ai/)
*   [Ollama](https://ollama.com/)
*   [osquery](https://osquery.io/)
*   [Qwen 2.5 (Ollama Library)](https://ollama.com/library/qwen2.5)
*   [SQLite](https://www.sqlite.org/)
*   [sqlite-rag](https://github.com/sqliteai/sqlite-rag)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview)
