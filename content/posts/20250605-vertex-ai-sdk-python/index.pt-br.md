+++
date = '2025-06-05T00:00:00+01:00'
title = 'Aprofundando-se no SDK da Vertex AI para Python'
summary = "Este artigo explora o modelo de comunicação entre o código do cliente e a API Gemini usando o SDK da Vertex AI para Python"
tags = ["gemini", "vertex ai", "python"]
+++
## Introdução

Este artigo explora o modelo de comunicação entre o código do cliente e a API Gemini usando o [SDK da Vertex AI para Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog). Cobriremos conceitos como a estrutura das mensagens, como o modelo entende o contexto da pergunta e como aumentar as capacidades do modelo com chamadas de função. Embora o Gemini seja o foco deste artigo, os mesmos conceitos que você verá aqui também podem ser aplicados ao Gemma e outros LLMs.

[No meu post anterior](https://danicat.dev/posts/20250531-diagnostic-agent/), expliquei como escrever um Agente de IA simples - mas surpreendentemente poderoso - que responde a perguntas de diagnóstico sobre sua máquina local. Em poucas linhas de código (e não tão poucas linhas de comentários), conseguimos fazer nosso agente responder a consultas como “quanta CPU eu tenho na minha máquina” ou “por favor, verifique se há sinais de malware”.

Isso, claro, deveu-se à beleza do SDK Python, pois simplificou muito as coisas. Por exemplo, contei com um recurso chamado [Chamada de Função Automática](https://ai.google.dev/gemini-api/docs/function-calling?example=weather#automatic_function_calling_python_only) para permitir que o agente decidisse quando chamar uma função. Esse recurso também me ajudou a definir as funções como funções Python simples e o SDK descobriu sua assinatura e descrição dinamicamente para mim. Essa capacidade, infelizmente, está disponível apenas para o SDK Python, então os desenvolvedores em outras linguagens precisam trabalhar um pouco mais.

É por isso que no artigo de hoje vamos adotar uma abordagem um pouco diferente e discutir como a API Gemini funciona para que você possa estar mais bem preparado para usar não apenas Python, mas qualquer um dos SDKs disponíveis (JS, Go e Java). Continuarei usando Python para os exemplos para que você possa comparar com o artigo anterior, mas os conceitos discutidos aqui são válidos para todas as diferentes linguagens.

Vamos cobrir dois tópicos principais:
*   Como funciona a conversa entre cliente e modelo
*   Como implementar chamadas de função manualmente

Observe que, se você é um desenvolvedor Python, isso também não significa que não aprenderá nada com este artigo. Na verdade, entender o fluxo da conversa será importante para usar conceitos mais avançados do SDK (como a Live API) e trabalhar com LLMs em geral.

## Entendendo como a API funciona

Os agentes normalmente funcionam da mesma forma que os aplicativos cliente-servidor - você tem um componente cliente responsável por preparar e fazer as solicitações e um processo servidor que hospeda o tempo de execução do modelo e processa as solicitações do cliente.

Para a Vertex AI, existem dois grupos principais de APIs: uma API REST para o estilo típico de solicitação/resposta de geração de conteúdo, onde o cliente envia uma solicitação e aguarda a resposta antes de continuar, e uma nova [Live API](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) que processa informações em tempo real usando websockets. Vamos nos concentrar primeiro nas APIs REST, pois a Live API requer um pouco mais de trabalho preparatório para funcionar corretamente.

Normalmente, geramos conteúdo em uma das seguintes modalidades: texto, imagem, áudio e vídeo. Muitos dos modelos mais recentes também são multimodais, o que significa que você pode lidar com mais de uma modalidade de entrada e/ou saída ao mesmo tempo. Para simplificar, vamos começar com texto.

Um aplicativo típico de prompt único se parece com isto:

```python
from google import genai

client = genai.Client(
    vertexai=True,
    project="daniela-genai-sandbox",
    location="us-central1"
)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="How are you today?"
)
print(response.text)

```

Saída:

```
Estou bem, obrigado por perguntar! Como um modelo de linguagem grande, não experimento emoções como os humanos, mas estou funcionando de forma otimizada e pronto para ajudá-lo. Como posso ajudá-lo hoje?
```

A primeira coisa que precisamos fazer é instanciar o cliente, usando o modo Vertex AI (`vertexai=True`) ou usando uma chave de API Gemini. Neste caso, estou usando o modo Vertex AI.

Assim que o cliente é inicializado, podemos enviar-lhe um prompt usando o método `client.models.generate_content`. Precisamos especificar qual modelo estamos chamando (neste caso `gemini-2.0-flash)` e o prompt no argumento `contents` (por exemplo, `"How are you today?"`).

Olhando para este código, pode ser difícil imaginar o que está acontecendo por baixo dos panos, pois estamos obtendo muitas abstrações gratuitamente graças ao Python. A coisa mais importante neste caso é que o **conteúdo não é uma string**.

`Contents` é na verdade uma lista de estruturas de conteúdo, e as estruturas de conteúdo são compostas por uma **função (role)** e uma ou mais **partes (parts)**. O tipo subjacente para esta estrutura é definido na biblioteca `types` e se parece com isto:


```python
from google.genai import types

contents = [types.Content(
  role = "user",
  parts = [ types.Part_from_text("How are you today?")
)]
```

Portanto, sempre que digitamos `contents="How are you today?"`, o SDK Python faz essa transformação de string para “conteúdo com uma parte de string” automaticamente para nós.

Outra coisa importante a notar é que sempre que fazemos uma chamada para `generate_content`, o modelo está começando do zero. Isso significa que é nossa responsabilidade adicionar o contexto das mensagens anteriores ao próximo prompt. Vamos fazer um teste simples pedindo ao modelo que dia é hoje duas vezes seguidas:

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)
```

Saída:

```
$ python3 main.py
Hoje é domingo, 5 de novembro de 2023.

Hoje é sábado, 2 de novembro de 2024.
```

Existem dois problemas com a resposta acima: 1) ela alucinou, pois o modelo não tem como saber a data, e 2) deu duas respostas diferentes para a mesma pergunta. Podemos corrigir o 1) baseando-nos em uma ferramenta como uma chamada de `datetime` ou Pesquisa Google, mas quero focar no 2) porque mostra claramente que o modelo não se lembra do que acabou de dizer e demonstra o ponto acima de que é **nossa** responsabilidade manter o modelo atualizado sobre a conversa.

Vamos fazer uma pequena modificação no código:

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)

# cada elemento no array de conteúdos é geralmente referido como um "turno"
contents = [
    {
        "role": "user",
        "parts": [{
            "text": "what day is today?"
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
            "text": "what day is today?"
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

Observe que na segunda chamada ao modelo estamos incluindo todo o contexto no atributo `contents`. Observe também que a `role` de cada parte muda de “user” para “model” e depois para “user” novamente (“user” e “model” são os únicos valores possíveis para `role`). É assim que o modelo entende em que ponto da conversa está, também conhecido como “turno”. Se, por exemplo, omitíssemos a última parte que repete a pergunta, o modelo pensaria que está atualizado e não produziria outra resposta, pois o último turno seria de “model” e não de “user”.

A variável `contents` acima está escrita na forma de “dicionário”, mas o SDK também fornece vários métodos de conveniência como `types.UserContent` (define o campo `role` como “user” automaticamente) e `types.Part.from_text` (converte uma string simples em uma parte), entre outros.

Para lidar com outros tipos de entradas e/ou saídas, podemos usar outros tipos de partes, como chamadas de função, dados binários, etc. Se um modelo for multimodal, você pode misturar partes de diferentes tipos de conteúdo na mesma mensagem.

Os dados binários podem ser tanto inline quanto buscados de um URI. Você pode diferenciar entre diferentes tipos de dados usando o campo `mime_type`. Por exemplo, uma parte de imagem pode ser recuperada assim:

```python
from google.genai import types

contents = types.Part.from_uri(
  file_uri: 'gs://generativeai-downloads/images/scones.jpg',
  mime_type: 'image/jpeg',
)
```

Ou inline:

```python
contents = types.Part.from_bytes(
  data: my_cat_picture, # dados binários
  mime_type: 'image/jpeg',
)
```

Em resumo, para cada turno da conversa, adicionaremos uma nova linha de conteúdo tanto para a resposta anterior do modelo quanto para a nova pergunta do usuário.

A boa notícia é que a experiência de chatbot é um caso de uso tão importante que o SDK da Vertex AI fornece uma implementação para esse fluxo pronta para uso. Usando o recurso `chat`, podemos reproduzir o comportamento acima em poucas linhas de código:

```python
chat = client.chats.create(model='gemini-2.0-flash')
response = chat.send_message('what day is today?')
print(response.text)
response = chat.send_message('what day is today?')
print(response.text)
```

Saída:

```
$ python3 main.py
Hoje é sábado, 14 de outubro de 2023.

Hoje é sábado, 14 de outubro de 2023.
```

Desta vez, o modelo lembrou a data porque a interface de chat está lidando com o histórico automaticamente para nós.

## Chamada de função não automática

Agora que vimos como a API funciona para construir mensagens do cliente e gerenciar o contexto, é hora de explorar como ela lida com chamadas de função. Em um nível básico, precisaremos instruir o modelo de que ele tem uma função à sua disposição e, em seguida, processar suas solicitações para chamar a função e retornar os valores resultantes ao modelo. Isso é importante porque as chamadas de função permitem que os agentes interajam com sistemas externos e o mundo real, criando ações como recuperar dados ou acionar processos específicos, indo além de apenas gerar texto.

A declaração da função é o que diz ao modelo o que ele pode fazer. Ela informa ao modelo o nome da função, a descrição e seus argumentos. Por exemplo, abaixo está uma declaração de função para a função `get_random_number`:

```python
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Retorna um número aleatório",
}
```

É essa declaração que o modelo precisa saber para decidir quais funções chamar. A declaração da função tem três campos: nome, descrição e parâmetros - neste caso, a função não aceita parâmetros, então este campo é omitido. O modelo usa a descrição da função e a descrição de seus argumentos para decidir quando e como chamar cada função.

No artigo anterior, em vez de dar ao modelo uma declaração de função, fui preguiçoso e deixei o SDK descobrir isso para mim com base no docstring da minha função. Desta vez, vamos fazer diferente e declarar explicitamente uma função para entender melhor o fluxo subjacente.

A função, incluindo sua declaração, se parece com isto:

```python
def get_random_number():
    return 4 # escolhido por um lançamento de dado justo
             # garantido ser aleatório (https://xkcd.com/221/)

# a declaração informa ao modelo o que ele precisa saber sobre a função
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Retorna um número aleatório",
}
```

Você pode ver outros exemplos de declarações de função [aqui](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling#schema-examples?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog).

Em seguida, precisamos dizer ao modelo que ele tem acesso a esta função. Fazemos isso por meio da configuração do modelo, adicionando a função como uma ferramenta.

```python
tools = types.Tool(function_declarations=[get_random_number_decl])
config = types.GenerateContentConfig(tools=[tools])

# meu prompt inicial
contents = [types.Part.from_text(text="what is my lucky number today?")]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config, # observe como estamos adicionando a configuração à chamada do modelo
)

print(response.candidates[0].content.parts[0])
```

Se você executar o código acima, obterá algo assim:

```
$ python3 main.py
video_metadata=None thought=None inline_data=None file_data=None thought_signature=None code_execution_result=None executable_code=None function_call=FunctionCall(id=None, args={}, name='get_random_number') function_response=None text=None
```

O que você está vendo aqui é a primeira parte da resposta do modelo, e podemos ver que esta parte tem todos os campos vazios (`None`), exceto o campo `function_call`. Isso significa que o modelo quer que **nós** façamos essa chamada de função e, em seguida, retornemos seu resultado de volta ao modelo.

Isso inicialmente me intrigou, mas se você pensar bem, faz todo o sentido. O modelo sabe que a função existe, mas não tem absolutamente nenhuma ideia de como chamá-la. Da perspectiva do modelo, a função também não está rodando na mesma máquina, então o modelo não pode fazer nada exceto “pedir educadamente” para que façamos a chamada em seu nome.

Não tivemos que fazer isso no meu artigo anterior porque a Chamada de Função Automática assumiu o controle e simplificou as coisas para nós. A chamada ainda seguiu o mesmo fluxo, mas o SDK escondeu toda essa complexidade de nós.

A coisa óbvia a fazer agora é chamar a função real e retornar o resultado para o modelo, mas lembre-se, sem contexto o modelo não sabe nada sobre nossa solicitação anterior, então se você enviar apenas os resultados da função de volta, ele não terá ideia do que fazer com isso!

É por isso que precisamos enviar o histórico da interação até agora, e pelo menos até o ponto em que o modelo sabe que solicitou esse valor. O código abaixo assume que recebemos uma mensagem de chamada de função e precisamos enviar uma nova solicitação com as informações completas:

```python
# assumindo que já inspecionamos a resposta e sabemos o que o modelo quer
result = get_random_number() # faz a chamada de função real

# contents ainda contém o prompt original, então adicionaremos a resposta do modelo...
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

Neste artigo, vimos como o cliente do agente se comunica com o modelo no lado do servidor ou, em outras palavras, o “modelo de domínio” das comunicações LLM. Também removemos a cortina da “mágica” que o SDK Python faz por nós.

A automação é sempre conveniente e nos ajuda a alcançar resultados muito mais rapidamente, mas saber como ela realmente funciona geralmente é a grande diferença entre uma jornada tranquila e uma irregular ao implementar seu próprio agente, especialmente porque os casos especiais _nunca são tão fáceis_.

Eu sei que em tempos de "vibe coding", à primeira vista, é quase irônico dizer algo assim, mas uma das coisas que aprendi rapidamente ao programar no "vibe coding" é que se você for mais preciso ao falar com a IA, obterá resultados muito melhores em muito menos tempo. Portanto, agora não é hora de menosprezar o valor do conhecimento, mas sim de dobrá-lo - não apesar da IA, mas **por causa** dela.

Espero que você tenha gostado da jornada até agora. No próximo artigo, construiremos sobre este conhecimento para levar o agente de diagnóstico ao próximo nível, onde nenhum agente jamais esteve! (ou talvez tenha estado, mas certamente não o meu =^_^=)

Por favor, escreva seus comentários abaixo! Paz o/
