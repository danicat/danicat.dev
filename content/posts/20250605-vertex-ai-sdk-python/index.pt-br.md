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
status: "superseded"
superseded_by:
  - "/posts/20261120-ai-agents-from-scratch-with-google-adk"
  - "/posts/20260808-gemini-for-go-developers-part-1-model-family"
status_note: "A explicação sobre payloads Content/Part sem estado em múltiplos turnos e function calling manual usa o SDK unificado google-genai e continua tecnicamente precisa, mas o gemini-2.0-flash, a nomenclatura Vertex AI SDK e a abordagem anterior à Interactions API estão datados."
---
## Introdução

Este artigo explora o modelo de comunicação entre o código cliente e a API Gemini usando o [SDK da Vertex AI para Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog). Vamos cobrir conceitos como a forma como as mensagens são estruturadas, como o modelo compreende o contexto da pergunta e como ampliar as capacidades do modelo com chamadas de função (function calls). Embora o Gemini seja o foco deste artigo, os mesmos conceitos que você verá aqui também podem ser aplicados ao Gemma e a outros LLMs.

[Na Parte 1 desta série]({{< ref "/posts/20250531-diagnostic-agent" >}}), expliquei como escrever um agente de IA simples (porém surpreendentemente poderoso) que responde a perguntas de diagnóstico sobre a sua máquina local. Com pouquíssimas linhas de código (e não tão poucas linhas de comentários), conseguimos fazer nosso agente responder a perguntas como "quanta CPU eu tenho na minha máquina" ou "por favor, verifique se há sinais de malware".

Isso aconteceu, é claro, graças à beleza do SDK em Python, que simplificou bastante as coisas. Por exemplo, eu contei com um recurso chamado [Automatic Function Calling](https://ai.google.dev/gemini-api/docs/function-calling?example=weather#automatic_function_calling_python_only) para deixar o agente decidir quando chamar uma função. Esse recurso também me permitiu definir tudo apenas como funções Python comuns, deixando que o SDK descobrisse dinamicamente a assinatura e a descrição para mim. Infelizmente, essa capacidade está disponível apenas no SDK para Python, então pessoas desenvolvedoras em outras linguagens precisam fazer um pouco mais de trabalho.

É por isso que, no artigo de hoje, vamos adotar uma abordagem um pouco diferente e discutir como a API Gemini funciona para que você tenha uma base melhor ao usar não apenas Python, mas qualquer um dos SDKs disponíveis por aí (JS, Go e Java). Ainda vou usar Python nos exemplos para que você possa comparar com o artigo anterior, mas os conceitos discutidos aqui são válidos para todas as diferentes linguagens.

Vamos cobrir dois tópicos principais:
*   Como funciona a conversa entre o cliente e o modelo
*   Como implementar chamadas de função (function calls) de forma manual

Note que, se você programa em Python, isso não significa que não vá aproveitar nada deste artigo. Na verdade, entender o fluxo da conversa será importante para usar conceitos mais avançados do SDK (como a Live API) e para trabalhar com LLMs em geral.

## Compreendendo o funcionamento da API

Agentes tipicamente funcionam da mesma maneira que aplicações cliente-servidor: você tem um componente cliente responsável por preparar e enviar as requisições e um processo no servidor que hospeda o runtime do modelo e processa as requisições do cliente.

Para a Vertex AI existem dois grupos principais de APIs: APIs REST para o estilo tradicional de requisição/resposta na geração de conteúdo, onde o cliente envia uma requisição e aguarda a resposta antes de continuar, e uma nova [Live API](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) que processa informações em tempo real usando WebSockets. Vamos focar nas APIs REST primeiro, pois a Live API exige um pouco mais de preparação para funcionar direito.

Normalmente geramos conteúdo em uma das seguintes modalidades: texto, imagem, áudio e vídeo. Muitos dos modelos mais recentes também são multimodais, o que significa que você pode lidar com mais de uma modalidade na entrada e/ou na saída ao mesmo tempo. Para manter as coisas simples, vamos começar com texto.

Uma aplicação típica de prompt único se parece com isto:

```python
from google import genai

client = genai.Client(
    vertexai=True,
    project="daniela-genai-sandbox",
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
Estou funcionando perfeitamente, obrigado por perguntar! Como um modelo de linguagem grande, não sinto emoções como os humanos sentem, mas estou operando de forma otimizada e pronto para ajudar você. Como posso ajudar hoje?
```

A primeira coisa que precisamos fazer é instanciar o cliente, seja usando o modo Vertex AI (`vertexai=True`) ou usando uma chave de API do Gemini. Neste caso, estou usando o modo Vertex AI.

Assim que o cliente é inicializado, podemos enviar um prompt usando o método `client.models.generate_content`. Precisamos especificar qual modelo estamos chamando (neste caso `gemini-2.0-flash`) e o prompt no argumento `contents` (por exemplo, `"Como você está hoje?"`).

Olhando para esse código, pode ser difícil imaginar o que está acontecendo por baixo dos panos, pois estamos ganhando muitas abstrações de graça graças ao Python. A coisa mais importante neste caso é que **o conteúdo não é uma string**.

`contents` é, na verdade, uma lista de estruturas de conteúdo (`Content`), e estruturas de conteúdo são compostas por um **papel** (`role`) e uma ou mais **partes** (`parts`). O tipo subjacente dessa estrutura é definido na biblioteca `types` e se parece com isto:

```python
from google.genai import types

contents = [types.Content(
  role = "user",
  parts = [types.Part.from_text(text="Como você está hoje?")]
)]
```

Portanto, sempre que escrevemos `contents="Como você está hoje?"`, o SDK em Python faz essa transformação de string para "conteúdo com uma parte de texto" automaticamente para nós.

Outra coisa importante a observar é que, sempre que fazemos uma chamada para `generate_content`, o modelo está começando do zero. Isso significa que é nossa responsabilidade adicionar o contexto das mensagens anteriores no próximo prompt. Vamos fazer um teste simples perguntando ao modelo que dia é hoje duas vezes seguidas:

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

Há dois problemas com a resposta acima: 1) ele alucinou, já que o modelo não tem como saber a data, e 2) ele deu duas respostas diferentes para a mesma pergunta. Podemos corrigir o item 1) fazendo grounding em uma ferramenta como uma chamada `datetime` ou Busca do Google, mas quero focar no item 2) porque ele mostra claramente que o modelo não se lembra do que acabou de dizer e demonstra o ponto acima de que é **nossa** responsabilidade manter o modelo atualizado sobre a conversa.

Vamos fazer uma pequena modificação no código:

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="que dia é hoje?"
)
print(response.text)

# cada elemento no array contents costuma ser chamado de "turno" (turn)
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

Note que na segunda chamada ao modelo estamos incluindo todo o contexto no atributo `contents`. Note também que o `role` de cada parte muda de `"user"` para `"model"` e depois para `"user"` novamente (`"user"` e `"model"` são os únicos valores possíveis para `role`). É assim que o modelo entende em qual ponto da conversa está, isto é, o "turno". Se, por exemplo, omitíssemos a última parte que repete a pergunta, o modelo acharia que já está atualizado e não produziria outra resposta, pois o último turno seria de `"model"` e não de `"user"`.

A variável `contents` acima está escrita no formato de "dicionário", mas o SDK também fornece vários métodos de conveniência como `types.UserContent` (define o campo `role` como `"user"` automaticamente) e `types.Part.from_text` (converte uma string simples em uma parte), entre outros.

Para lidar com outros tipos de entradas e/ou saídas, podemos usar outros tipos de partes, como chamadas de função, dados binários, etc. Se um modelo for multimodal, você pode misturar partes de diferentes tipos de conteúdo na mesma mensagem.

Dados binários podem ser enviados tanto inline quanto buscados a partir de uma URI. Você pode diferenciar entre os diferentes tipos de dados usando o campo `mime_type`. Por exemplo, uma parte de imagem pode ser recuperada assim:

```python
from google.genai import types

contents = types.Part.from_uri(
  file_uri='gs://generativeai-downloads/images/scones.jpg',
  mime_type='image/jpeg',
)
```

Ou inline:

```python
contents = types.Part.from_bytes(
  data=my_cat_picture, # dados binários
  mime_type='image/jpeg',
)
```

Em resumo, a cada turno da conversa, vamos adicionar uma nova linha de conteúdo tanto para a resposta anterior do modelo quanto para a nova pergunta do usuário.

A boa notícia é que a experiência de chatbot é um caso de uso tão importante que o SDK da Vertex AI já oferece uma implementação pronta para esse fluxo. Usando o recurso `chat`, podemos reproduzir o comportamento acima em pouquíssimas linhas de código:

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

Desta vez o modelo se lembrou da data porque a interface `chat` está gerenciando o histórico automaticamente para nós.

## Implementando function calling de forma manual

Agora que vimos como a API funciona para construir mensagens no cliente e gerenciar o contexto, é hora de explorar como ela lida com chamadas de função (function calls). Em um nível básico, precisaremos instruir o modelo de que ele tem uma função à sua disposição e, em seguida, processar suas requisições para chamar a função e devolver os valores resultantes ao modelo. Isso é importante porque as chamadas de função permitem que os agentes interajam com sistemas externos e com o mundo real, criando ações como recuperar dados ou disparar processos específicos, indo além de apenas gerar texto.

A declaração da função é o que informa ao modelo o que ele pode fazer. Ela diz ao modelo o nome da função, a descrição e seus argumentos. Por exemplo, abaixo está uma declaração de função para a função `get_random_number`:

```python
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Retorna um número aleatório",
}
```

É dessa declaração que o modelo precisa para decidir quais funções chamar. A declaração da função tem três campos: `name`, `description` e `parameters`; neste caso a função não aceita parâmetros, então esse campo foi omitido. O modelo usa a descrição da função e a descrição de seus argumentos para decidir quando e como chamar cada função.

No artigo anterior, em vez de passar uma declaração de função ao modelo, fui preguiçosa e deixei o SDK descobrir tudo para mim com base na docstring da minha função. Desta vez, vamos fazer diferente e declarar explicitamente uma função para entender melhor o fluxo subjacente.

A função junto com sua declaração fica assim:

```python
def get_random_number():
    return 4 # escolhido por um lançamento justo de dados
             # garantidamente aleatório (https://xkcd.com/221/)

# a declaração informa ao modelo o que ele precisa saber sobre a função
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Retorna um número aleatório",
}
```

Você pode ver outros exemplos de declarações de função na [documentação de function calling da Vertex AI](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog#schema-examples).

Em seguida, precisamos informar ao modelo que ele tem acesso a essa função. Fazemos isso por meio da configuração do modelo, adicionando a função como uma ferramenta (`tool`).

```python
tools = types.Tool(function_declarations=[get_random_number_decl])
config = types.GenerateContentConfig(tools=[tools])

# meu prompt inicial
contents = [types.Part.from_text(text="qual é o meu número da sorte hoje?")]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config, # note como estamos adicionando o config na chamada do modelo
)

print(response.candidates[0].content.parts[0])
```

Se você executar o código acima, receberá algo parecido com isto:

```
$ python3 main.py 
video_metadata=None thought=None inline_data=None file_data=None thought_signature=None code_execution_result=None executable_code=None function_call=FunctionCall(id=None, args={}, name='get_random_number') function_response=None text=None
```

O que você está vendo aqui é a primeira parte da resposta do modelo, e podemos notar que essa parte tem todos os campos vazios (`None`), exceto o campo `function_call`. Isso significa que o modelo quer que **nós** façamos essa chamada de função e depois devolvamos o resultado para o modelo.

Isso me deixou confusa no começo, mas se você parar para pensar, faz todo sentido. O modelo sabe que a função existe, mas não faz a menor ideia de como chamá-la. Da perspectiva do modelo, a função também não está rodando na mesma máquina, então o modelo não pode fazer nada além de "pedir educadamente" que executemos a chamada em nome dele.

Não precisamos fazer isso no artigo anterior porque o Automatic Function Calling assumiu o controle e simplificou as coisas para nós. A chamada ainda seguiu o mesmo fluxo, mas o SDK havia escondido toda essa complexidade de nós.

A coisa óbvia a fazer agora é chamar a função de verdade e devolver o resultado ao modelo, mas lembre-se: sem contexto o modelo não sabe nada sobre nossa requisição anterior, então se você enviar apenas os resultados da função de volta ele não terá ideia do que fazer com eles!

É por isso que precisamos enviar o histórico da interação até agora, e pelo menos até o ponto em que o modelo sabe que solicitou aquele valor. O código abaixo pressupõe que recebemos uma mensagem de chamada de função e precisamos enviar uma nova requisição com as informações completas:

```python
# assumindo que já inspecionamos a resposta e sabemos o que o modelo quer
result = get_random_number() # faz a chamada real da função

# contents ainda contém o prompt original, então vamos adicionar a resposta do modelo...
contents.append(types.ModelContent(parts=response.candidates[0].content.parts))
# ... e o resultado da chamada de função
contents.append(types.UserContent(parts=types.Part.from_function_response(name="get_random_number", response={"result": result})))

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
O número da sorte de hoje é 4.
```

## Conclusões

Neste artigo vimos como o cliente do agente se comunica com o modelo no lado do servidor ou, em outras palavras, o "modelo de domínio" da comunicação com LLMs. Também levantamos a cortina sobre a "mágica" que o SDK em Python faz para nós.

A automação é sempre conveniente e nos ajuda a alcançar resultados muito mais rápido, mas saber como ela realmente funciona costuma ser a grande diferença entre uma jornada tranquila e uma cheia de remendos na hora de implementar o seu próprio agente, especialmente porque os casos especiais *nunca são tão fáceis assim*.

Sei que, em tempos de vibe coding, à primeira vista chega a ser quase irônico dizer algo assim, mas uma das coisas que aprendi rapidamente praticando vibe coding é que, se você tiver mais precisão ao conversar com a IA, consegue resultados muito melhores em muito menos tempo. Então agora não é hora de subestimar o valor do conhecimento, mas de dobrar a aposta nele: não apesar da IA, mas **por causa** dela.

Espero que você tenha gostado da jornada até aqui. Na próxima parte desta série, [Prompt Audacioso: Um Guia Prático para Instruções de Sistema e Ferramentas]({{< ref "/posts/20250611-system-prompt" >}}), vamos construir sobre esse conhecimento para criar uma aplicação de chat dedicada no terminal com histórico de sessão e injeção dinâmica de schemas de tabelas.

Por favor, deixe seus comentários abaixo! Paz o/
