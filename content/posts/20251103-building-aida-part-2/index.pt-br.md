---
categories:
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
title: 'Como Construir um Agente Offline com ADK, Ollama e SQLite'
---
[Na Parte 5 desta série]({{< ref "/posts/20251031-building-aida" >}}), focamos na criação de uma interface de cliente personalizada para o nosso agente. Foi um passo importante para tornar a experiência mais interativa, mas ainda restava uma questão fundamental: e se a rede cair?

Embora a falta de conexão seja um problema para qualquer agente, a nuance aqui é que estamos construindo um "Agente de Diagnóstico de Emergência". De que adianta um assistente de emergência se ele não funciona quando você mais precisa — exatamente quando o computador perde a conexão com a internet?

Isso me levou a desenhar um mecanismo de contingência (fallback): e se pudéssemos rodar todo o diagnóstico apenas com dependências locais? Isso exigiria não apenas trocar o modelo central por um LLM local, mas também repensar nossa estratégia de RAG.

Os benefícios são imediatos: quando online, aproveitamos a inteligência e velocidade de modelos em nuvem como o Gemini; em caso de falha de rede ou em ambientes isolados (air-gapped) com altas exigências de privacidade, o agente faz o fallback para o runtime local.

Neste artigo, vamos implementar todos os componentes necessários para tornar o agente de diagnóstico 100% offline.

## Substituindo o Modelo em Nuvem por um LLM Local

A forma mais simples e popular de rodar modelos locais atualmente é o [**Ollama**](https://ollama.com/). No macOS, você pode instalá-lo via Homebrew:

```bash
brew install ollama
```

Com o Ollama em execução, baixamos o modelo desejado com o comando `ollama pull`:

```bash
ollama pull qwen2.5
```

A família [`qwen2.5`](https://ollama.com/library/qwen2.5) oferece diversos tamanhos (1B, 2B, 7B, 14B, 32B). Para o nosso caso de uso, precisamos de um modelo que ofereça suporte nativo a **Function Calling (Tools)**, já que ele precisará orquestrar consultas ao [**Osquery**](https://osquery.io/) e ferramentas de RAG.

Após avaliar várias opções, optei pelo **Qwen 2.5 7B**:

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
*   **GPT-OSS:** Bom em conversa natural, mas entrava em loops repetindo queries como `SELECT * FROM system_info` sem avançar.
*   **Llama 3.1:** Apresentou dificuldades na precisão dos argumentos das ferramentas.
*   **Qwen 2.5:** O melhor modelo local de 7B para chamadas de função confiáveis, mantendo um diálogo fluido.

Embora não tenha o raciocínio encadeado de um modelo maior em nuvem como o Gemini 2.5 Flash, ele é perfeitamente capaz de conduzir investigações offline.

### Conectando Modelos Locais ao ADK com LiteLLM

Para integrar o Ollama ao Google ADK, utilizamos a biblioteca [**LiteLLM**](https://www.litellm.ai/), que padroniza a interface de múltiplos provedores. Isso nos permite alternar o backend do modelo em poucas linhas:

```python
# aida/agent.py
from google.adk.models.lite_llm import LiteLlm

# Criamos o objeto LiteLlm apontando para o provedor ollama_chat
MODEL = LiteLlm(model="ollama_chat/qwen2.5")

root_agent = Agent(
    model=MODEL,
    name="aida",
    description="The emergency diagnostic agent",
    # ... instruções e ferramentas ...
)
```

<video controls width="100%" src="aida_demo_hd.mov">
  Seu navegador não suporta a tag de vídeo.
</video>
<p style="text-align: center; font-style: italic; opacity: 0.8; margin-top: 0.5rem;">AIDA executando primeiro com Gemini 2.5 Flash e depois com Qwen 2.5 local em um Apple MacBook Pro M4 com 48GB de RAM.</p>

Com o modelo rodando localmente, precisamos agora substituir a dependência de nuvem do [**Vertex AI RAG**](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview).

## Criando uma Base de Conhecimento RAG Local com SQLite

O Vertex AI RAG é uma solução poderosa para grandes corporações que gerenciam petabytes de dados em tempo real. No entanto, para armazenar algumas centenas de schemas estáticos do Osquery, ele é um exagero desnecessário.

Como o Osquery já é baseado no motor do [**SQLite**](https://www.sqlite.org/), a escolha lógica foi construir uma solução RAG embarcada utilizando o projeto **[`sqlite-rag`](https://github.com/sqliteai/sqlite-rag)**.

### Ingestão dos Schemas do Osquery no SQLite

Criamos o script `ingest_osquery.py` para ler os arquivos `.table` da especificação do Osquery e gravá-los no banco `schema.db`:

```python
# ingest_osquery.py
import os
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

    print(f"Inicializando banco RAG em {DB_PATH}...")
    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})

    files_to_ingest = []
    for root, _, files in os.walk(SPECS_DIR):
        for file in files:
            if file.endswith(".table"):
                files_to_ingest.append(os.path.join(root, file))

    total_files = len(files_to_ingest)
    for i, file_path in enumerate(files_to_ingest):
        ingest(rag, file_path)

    print("Quantizando vetores para otimizar busca local...")
    rag.quantize_vectors()
    rag.close()
    print("Ingestão concluída!")
```

A etapa `quantize_vectors()` converte embeddings de ponto flutuante de 32 bits em inteiros de 8 bits. Essa quantização reduz drasticamente o tamanho do banco SQLite e acelera as buscas semânticas na CPU local.

### Implementando a Ferramenta de Descoberta de Schemas Local

Implementamos a ferramenta `discover_schema` para consultar a base SQLite:

```python
# aida/schema_rag.py
import os
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
SCHEMA_DB_PATH = os.path.join(PROJECT_ROOT, "schema.db")

schema_rag = SQLiteRag.create(SCHEMA_DB_PATH, require_existing=True)

def discover_schema(search_terms: str, top_k: int = 5) -> list[DocumentResult]:
    """Consulta os schemas de tabelas do osquery via RAG local.

    Args:
        search_terms: Termos de busca (ex: 'battery', 'system information darwin').
        top_k: Número de resultados relevantes a retornar.

    Returns:
        Lista com as definições de schemas das tabelas encontradas.
    """
    return schema_rag.search(search_terms, top_k=top_k)
```

Agora o Qwen consegue descobrir tabelas diretamente no SQLite local:

![Captura de tela da AIDA](image-1.png "Busca de schema para 'battery' executada localmente pelo Qwen")

## Enriquecendo o Conhecimento com Query Packs do Osquery

Um modelo local de 7B possui uma janela de contexto menor (32k tokens contra 1 milhão no Gemini) e menor capacidade de formular queries complexas do zero sem assistência.

Para compensar essa diferença de escala, integramos uma segunda base RAG contendo os **Query Packs** oficiais da comunidade Osquery (conjuntos de queries prontas e testadas para detecção de ameaças, conformidade e auditoria de segurança).

### Ingestão dos Packs no SQLite

Criamos o script `ingest_packs.py` para processar os arquivos `.conf` e `.json` dos packs:

```python
# ingest_packs.py
import glob
import json
import os
import re
import sqlite3
from sqlite_rag import SQLiteRag

DB_PATH = os.path.abspath("packs.db")
PACKS_DIR = "osquery_data/packs"

def ingest_pack(rag, pack_path):
    pack_name = os.path.basename(pack_path).replace(".conf", "").replace(".json", "")
    with open(pack_path, "r") as f:
        content = re.sub(r"\s*\n", " ", f.read())
        data = json.loads(content)

    pack_platform = data.get("platform", "all")
    for query_name, query_data in data.get("queries", {}).items():
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
            pass

def main():
    if os.path.exists(DB_PATH):
        os.remove(DB_PATH)

    rag = SQLiteRag.create(DB_PATH, settings={"quantize_scan": True})
    pack_files = glob.glob(os.path.join(PACKS_DIR, "*.conf")) + glob.glob(os.path.join(PACKS_DIR, "*.json"))

    for pack_file in pack_files:
        ingest_pack(rag, pack_file)

    rag.quantize_vectors()
    rag.close()

if __name__ == "__main__":
    main()
```

### Ferramenta de Biblioteca de Queries (`search_query_library`)

Disponibilizamos a nova ferramenta para o agente:

```python
# aida/queries_rag.py
import os
from sqlite_rag import SQLiteRag
from sqlite_rag.models.document_result import DocumentResult

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
PACKS_DB_PATH = os.path.join(PROJECT_ROOT, "packs.db")

queries_rag = SQLiteRag.create(PACKS_DB_PATH, require_existing=True)

def search_query_library(search_terms: str, platform: str = "all", top_k: int = 5) -> list[DocumentResult]:
    """Busca queries predefinidas na biblioteca de packs do osquery.

    Args:
        search_terms: Termos descritivos (ex: 'malware detection', 'rootkit').
        platform: 'darwin', 'linux', 'windows' ou 'all'.
        top_k: Quantidade de candidatos.

    Returns:
        Queries SQL testadas e metadados associados.
    """
    if platform == "all" or platform is None:
        search_terms += " windows linux darwin"
    else:
        search_terms += f" {platform}"

    return queries_rag.search(search_terms, top_k=top_k)
```

Instruímos o fluxo de trabalho do agente:

```python
# aida/agent.py
root_agent = Agent(
    model=MODEL,
    name="aida",
    description="Agente de diagnóstico de emergência",
    instruction="""
[FLUXO OPERACIONAL]
1. BUSCA (SEARCH): Para tarefas de alto nível (ex: "verifique se há rootkits"), use PRIMEIRO `search_query_library`.
2. DESCOBERTA (DISCOVER): Se não houver query pronta adequada, use `discover_schema` para inspecionar tabelas e colunas.
3. EXECUÇÃO (EXECUTE): Use `run_osquery` para executar a query SQL selecionada.
    """,
    tools=[
        search_query_library,
        discover_schema,
        run_osquery,
    ],
)
```

Veja o fluxo em ação com o modelo local identificando e executando queries de detecção de malware:

![AIDA executando verificação de malware offline](image-2.png "AIDA local buscando queries na biblioteca de packs e executando no osquery")

## Conclusões

Completamos a jornada da série "Construindo o Agente de Diagnóstico"! Criamos um assistente autônomo, com interface visual interativa e capacidade de operar tanto em nuvem de alta performance quanto em modo 100% offline, resiliente a falhas de rede.

O código-fonte completo de todas as etapas do projeto AIDA está disponível no GitHub: **[github.com/danicat/aida](https://github.com/danicat/aida)**.

Obrigada por acompanhar toda a série! Deixe suas dúvidas, feedbacks e experimentos nos comentários abaixo.

## Referências

*   [Gemini 2.5 Flash](https://deepmind.google/technologies/gemini/flash/)
*   [LiteLLM](https://www.litellm.ai/)
*   [Ollama](https://ollama.com/)
*   [Osquery](https://osquery.io/)
*   [Qwen 2.5 (Ollama Library)](https://ollama.com/library/qwen2.5)
*   [SQLite](https://www.sqlite.org/)
*   [sqlite-rag](https://github.com/sqliteai/sqlite-rag)
*   [Vertex AI RAG](https://cloud.google.com/vertex-ai/docs/generative-ai/grounding/overview)
