---
categories:
- Agent Development
date: '2025-06-05T00:00:00+01:00'
series:
- Building the Diagnostic Agent
series_order: 2
summary: Este artigo explora o modelo de comunicação entre o código cliente e a API Gemini usando o SDK da Vertex AI para Python.
tags:
  - gemini
  - python
  - tutorial
  - vertex-ai
title: "Explorando a Fundo o SDK da Vertex AI para Python"
slug: "vertex-ai-sdk-python"
aliases:
  - "/pt-br/posts/20250605-vertex-ai-sdk-python/"
description: "Guia aprofundado sobre a comunicação entre cliente e API no SDK da Vertex AI para Python. Aprenda a estruturar instruções de sistema, function calling e histórico."
proficiencyLevel: "Intermediate"
dependencies:
  - "Python 3.10+"
  - "google-genai"
  - "Google Cloud Vertex AI"
---
## Introdução

Este artigo explora o modelo de comunicação entre o código cliente e a API Gemini usando o [SDK da Vertex AI para Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog). Vamos cobrir como as mensagens são estruturadas, como o modelo compreende o contexto de uma conversa e como expandir as capacidades do modelo com function calls (chamadas de função). Embora o foco aqui seja o Gemini, os mesmos conceitos se aplicam a modelos como o Gemma e outros LLMs modernos.

[Na Parte 1 desta série]({{< ref "/posts/20250531-diagnostic-agent" >}}), mostrei como criar um agente de IA simples — porém surpreendentemente poderoso — que responde a perguntas de diagnóstico sobre a sua máquina local. Com pouquíssimas linhas de código (e comentários bem detalhados), nosso agente já respondia a pedidos como "quanto de CPU tenho disponível" ou "verifique se há sinais de malware".

Isso foi possível graças à facilidade do SDK do Python, que abstrai boa parte da complexidade. Por exemplo, usei o recurso de [Automatic Function Calling](https://ai.google.dev/gemini-api/docs/function-calling?example=weather#automatic_function_calling_python_only) para deixar o agente decidir quando invocar cada função. Esse recurso também me permitiu declarar funções Python normais enquanto o SDK inferia assinaturas e docstrings dinamicamente. No entanto, essa facilidade é exclusiva do SDK Python; desenvolvedores em Go, JavaScript ou Java precisam estruturar essas chamadas de forma manual.

Por isso, neste artigo vamos entender o funcionamento da API Gemini por baixo do capô, preparando você para usar com segurança não apenas o SDK Python, mas qualquer outro SDK disponível (Go, JS, Java). Continuarei usando Python nos exemplos para facilitar a comparação com o post anterior, mas a lógica se aplica a qualquer linguagem.

Abordaremos dois tópicos principais:
*   Como funciona o fluxo de conversa entre o cliente e o modelo
*   Como implementar function calling de forma manual em baixo nível

Mesmo se você programa exclusivamente em Python, entender o ciclo de vida dessas mensagens é fundamental para dominar recursos avançados do SDK (como a Live API) e construir agentes robustos.

## Compreendendo o funcionamento da API

Agentes funcionam como aplicações cliente-servidor clássicas: de um lado, há um cliente encarregado de montar e enviar as requisições; do outro, um processo remoto que hospeda o runtime do modelo e processa os dados recebidos.

Na Vertex AI, encontramos dois grupos principais de APIs: APIs REST para o modelo tradicional de requisição/resposta (onde o cliente envia uma mensagem e aguarda a conclusão antes de continuar) e a [Live API](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog), que processa streaming bidirecional em tempo real via WebSockets. Vamos focar primeiro nas APIs REST.

Podemos gerar conteúdo em diversas modalidades: texto, imagem, áudio e vídeo. Os modelos mais recentes são multimodais nativos, permitindo combinar diferentes tipos de entrada e saída em uma mesma interação. Para manter as coisas simples, vamos começar com texto.

Uma chamada direta e pontual a um modelo se parece com isto:

```python
from google import genai

client = genai.Client(
    vertexai=True,
    project="meu-projeto-sandbox",
    location="us-central1"
)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="Como você está hoje?"
)
print(response.text)
```

Saída:

```
Estou funcionando perfeitamente, obrigado por perguntar! Como um modelo de linguagem, não sinto emoções como os humanos, mas estou pronto para ajudar. Como posso ser útil hoje?
```

Primeiro instanciamos o cliente — seja no modo Vertex AI (`vertexai=True`) ou fornecendo uma chave de API do Gemini Developer.

Em seguida, enviamos o prompt chamando `client.models.generate_content`, especificando o modelo (`gemini-2.0-flash`) e a mensagem no argumento `contents`.

À primeira vista, parece uma simples troca de strings. Mas, por baixo das abstrações do Python, **o conteúdo não é uma string simples**.

`contents` é, na verdade, uma lista de estruturas de conteúdo (`Content`), onde cada estrutura é composta por uma função (`role`) e uma ou mais partes (`parts`). A estrutura interna definida na biblioteca de tipos se parece com:

```python
from google.genai import types

contents = [types.Content(
  role = "user",
  parts = [ types.Part.from_text("Como você está hoje?") ]
)]
```

Quando passamos `contents="Como você está hoje?"`, o SDK do Python converte automaticamente a string para uma estrutura `Content` contendo uma `Part` de texto.

Outro ponto crucial: a cada chamada a `generate_content`, o modelo é executado sem memória prévia. É **nossa** responsabilidade incluir o histórico de mensagens anteriores nas próximas requisições se quisermos manter uma conversa contínua. Vamos comprovar isso perguntando que dia é hoje duas vezes seguidas:

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="que dia é hoje?"
)
print(response.text)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="que dia é hoje?"
)
print(response.text)
```

Saída:

```
$ python3 main.py 
Hoje é domingo, 5 de novembro de 2023.

Hoje é sábado, 2 de novembro de 2024.
```

Temos dois problemas aqui: 1) o modelo alucinou, pois não tem acesso ao relógio do sistema, e 2) forneceu duas respostas totalmente diferentes para a mesma pergunta. O problema 1 é resolvido com ferramentas (como uma função datetime ou busca no Google), mas o problema 2 demonstra que o modelo não se lembra do que acabou de responder. O controle do estado da conversa cabe exclusivamente ao código cliente.

Vamos ajustar o código para enviar o histórico da conversa:

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="que dia é hoje?"
)
print(response.text)

# Cada item na lista contents representa um "turno" (turn) da conversa
contents = [
    {
        "role": "user",
        "parts": [{
            "text": "que dia é hoje?"
        }]
    },
    {
        "role": "model",
        "parts": [{
            "text": response.text
        }]
    },
    {
        "role": "user",
        "parts": [{
            "text": "que dia é hoje?"
        }]
    },
]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents
)
print(response.text)
```

Saída:

```
$ python3 main.py 
Hoje é quarta-feira, 15 de novembro de 2023.

Hoje é quarta-feira, 15 de novembro de 2023.
```

Na segunda chamada, enviamos o histórico completo. Observe que o atributo `role` alterna entre `"user"` e `"model"` (os únicos valores aceitos para papel na conversa). É assim que o modelo identifica em qual turno do diálogo está. Se omitíssemos o último turno do usuário, o modelo assumiria que a conversa terminou e não geraria uma nova resposta.

O SDK também fornece helpers como `types.UserContent` e `types.Part.from_text` para montar essas estruturas de forma mais elegante.

Para trabalhar com dados multimídia, usamos partes específicas (como chamadas de função, dados binários ou URIs do Cloud Storage):

```python
from google.genai import types

contents = types.Part.from_uri(
  file_uri='gs://generativeai-downloads/images/scones.jpg',
  mime_type='image/jpeg',
)
```

Ou dados binários inline:

```python
contents = types.Part.from_bytes(
  data=minha_foto_binaria,
  mime_type='image/jpeg',
)
```

Como gerenciar esse histórico é uma necessidade constante em aplicações conversacionais, o SDK da Vertex AI já traz essa abstração pronta na API `chats`:

```python
chat = client.chats.create(model='gemini-2.0-flash')
response = chat.send_message('que dia é hoje?')
print(response.text)
response = chat.send_message('que dia é hoje?')
print(response.text)
```

Saída:

```
$ python3 main.py 
Hoje é sábado, 14 de outubro de 2023.

Hoje é sábado, 14 de outubro de 2023.
```

Aqui o objeto `chat` gerencia o histórico de turnos automaticamente para nós.

## Implementando Function Calling de forma manual

Agora que entendemos como o modelo consome mensagens e gerencia contexto, vamos ver como funciona o ciclo de chamadas de ferramentas sem depender de mágica automática.

Para que o modelo utilize uma função, precisamos declarar a existência dessa função (nome, descrição e parâmetros aceitos). O modelo analisa essas declarações para decidir quando uma chamada é necessária.

Veja a declaração da função `get_random_number`:

```python
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Retorna um número aleatório",
}
```

Essa declaração possui três campos principais: `name`, `description` e `parameters` (como esta função não recebe argumentos, o campo foi omitido). O modelo se baseia na descrição da função e dos parâmetros para decidir se deve ou não acioná-la.

No artigo anterior, deixamos o SDK extrair esses metadados da docstring da função. Agora, vamos declarar a função e a ferramenta explicitamente:

```python
def get_random_number():
    return 4 # escolhido por um lançamento justo de dados
             # garantidamente aleatório (https://xkcd.com/221/)

get_random_number_decl = {
    "name": "get_random_number",
    "description": "Retorna um número aleatório",
}
```

Você pode conferir outros esquemas de declaração na [documentação de function calling da Vertex AI](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling#schema-examples?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog).

Agora registramos a ferramenta na configuração do modelo:

```python
tools = types.Tool(function_declarations=[get_random_number_decl])
config = types.GenerateContentConfig(tools=[tools])

contents = [types.Part.from_text(text="qual é o meu número da sorte hoje?")]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config,
)

print(response.candidates[0].content.parts[0])
```

Ao executar o código acima, obtemos:

```
$ python3 main.py 
video_metadata=None thought=None inline_data=None file_data=None thought_signature=None code_execution_result=None executable_code=None function_call=FunctionCall(id=None, args={}, name='get_random_number') function_response=None text=None
```

Observe que todos os campos da resposta estão vazios (`None`), exceto o campo `function_call`. O modelo não gerou texto final; ele está nos informando: *"Por favor, execute a função `get_random_number` e me devolva o resultado"*.

O modelo sabe que a função existe, mas não pode executá-la diretamente no servidor do Google. Ele depende do cliente para rodar o código localmente e fornecer a resposta de volta.

Para concluir o ciclo, executamos a função no cliente e enviamos o resultado de volta ao modelo, acompanhado de todo o histórico anterior:

```python
# 1. Executamos a função solicitada pelo modelo
result = get_random_number()

# 2. Anexamos a resposta do modelo (o pedido de execução da função)...
contents.append(types.ModelContent(parts=response.candidates[0].content.parts))

# 3. ... e anexamos o resultado retornado pela nossa função local
contents.append(types.UserContent(parts=types.Part.from_function_response(name="get_random_number", response={"result": result})))

# 4. Enviamos o histórico atualizado de volta ao modelo
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config,
)
print(response.text)
```

Saída:

```
$ python3 main.py 
Seu número da sorte para hoje é 4.
```

## Conclusões

Neste artigo, desmistificamos como o cliente de um agente conversa com o modelo em nível de protocolo e eliminamos a "mágica" das ferramentas automáticas.

A automação é excelente para produtividade, mas compreender o fluxo real de mensagens é o que diferencia quem apenas copia exemplos de quem consegue diagnosticar bugs e arquitetar sistemas de agentes confiáveis.

No contexto atual de vibe coding, saber exatamente o que acontece por baixo dos panos permite criar prompts e especificações muito mais precisas, economizando tempo e recursos de computação.

Na próxima parte desta série, [Instruções de Sistema e Ferramentas para Agentes]({{< ref "/posts/20250611-system-prompt" >}}), vamos usar essa base para criar uma aplicação de chat completa no terminal com histórico persistente de sessão e injeção dinâmica de schemas de tabelas.

Deixe seus comentários e dúvidas abaixo! Até a próxima o/
