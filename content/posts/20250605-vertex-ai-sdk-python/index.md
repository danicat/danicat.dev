---
date: '2025-06-05T00:00:00+01:00'
title: 'Digging deeper into the Vertex AI SDK for Python'
summary: "This article explores the communication model between the client code and the Gemini API using the Vertex AI SDK for Python"
tags: ["gemini", "vertex-ai", "python"]
categories: ["AI & Development"]
---
## Introduction

This article explores the communication model between the client code and the Gemini API using the [Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog). We will cover concepts like how the messages are structured, how the model understands the context of the question and how to augment the model capabilities with function calls. While Gemini is the focus of this article, the same concepts you will see here can also be applied to Gemma and other LLMs.

[In my previous post](https://danicat.dev/posts/20250531-diagnostic-agent/) I’ve explained how to write a simple - but surprisingly powerful - AI Agent that responds to diagnostic questions about your local machine. In very few lines of code (and not so few lines of comments) we were able to get our agent to respond to queries like “how much CPU I have in my machine” or “please check for any signs of malware”.

That was, of course, due to the beauty of the Python SDK as it simplified things a lot. For example, I relied on a feature called [Automatic Function Calling](https://ai.google.dev/gemini-api/docs/function-calling?example=weather#automatic_function_calling_python_only) to let the agent decide when to call a function. This feature also helped me get away with defining the functions as plain Python functions and the SDK figured out its signature and description dynamically for me. This capability unfortunately is only available for the Python SDK, so developers in other languages need to do a bit more work.

This is why in today’s article we are going to take a slightly different approach and discuss how the Gemini API works so that you can be better prepared for using not only Python, but any of the available SDKs out there (JS, Go and Java). I’ll still be using Python for the examples so you can compare with the previous article, but the concepts discussed here are valid across all different languages.

We are going to cover two main topics: 
*   How the conversation between client and model works
*   How to implement function calls the manual way

Note that if you are a Python developer this doesn’t mean you won’t get anything from this article either. Actually, understanding the flow of the conversation will be important to use more advanced concepts of the SDK (like the Live API) and working with LLMs in general.

## Understanding How the API works

Agents typically work in the same way as client server applications - you have a client component who is responsible for preparing and making the requests and a server process that hosts the model runtime and processes the client requests.

For Vertex AI there are two main groups of APIs: a REST APIs for the typical request/response style of content generation, where the client sends a request and waits for the response before continuing, and a new [Live API](https://cloud.google.com/vertex-ai/generative-ai/docs/live-api?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog) that processes real time information using websockets. We are going to focus on the REST APIs first, as the Live API requires a bit more groundwork to get it right. 

We typically generate content in one of the following modalities: text, image, audio and video. Many of the most recent models are also multi-modal, which means you can deal with more than one modality for input and/or output at the same time. To keep things simple, let’s start with text.

A typical one-off prompt application looks like this:

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

Output:

```
I am doing well, thank you for asking! As a large language model, I don't experience emotions like humans do, but I am functioning optimally and ready to assist you. How can I help you today?
```

The first thing we need to do is instantiate the client, either using the Vertex AI mode (`vertexai=True`) or using a Gemini API key. In this case I’m using Vertex AI mode.

Once the client is initialized, we can send it a prompt using the `client.models.generate_content` method. We need to specify which model we are calling (in this case `gemini-2.0-flash)` and the prompt in the contents argument (e.g. `"How are you today?"`).

Looking at this code it might be hard to imagine what is going on under the hood, as we are getting many abstractions for free thanks to Python. The most important thing in this case is that the **content is not a string**.

Contents is actually a list of content structures, and content structures are composed by a **role** and one or more **parts**. The underlying type for this structure is defined in the types library and looks like this:


```python
from google.genai import types

contents = [types.Content(
  role = "user",
  parts = [ types.Part_from_text("How are you today?")
)]
```

So whenever we type `contents="How are you today?",` the Python SDK does this transformation from string to “content with a string part” automatically for us.

Another important thing to note is that whenever we make a call to generate_content, the model is starting from zero. This means that it is our responsibility to add the context of previous messages to the next prompt. Let’s do a simple test by asking the model what day is today two times in a row:

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

Output:

```
$ python3 main.py 
Today is Sunday, November 5th, 2023.

Today is Saturday, November 2nd, 2024.
```

There are two problems with the response above: 1) it hallucinated, as the model has no way of knowing the date, and 2) it gave two different answers to the same question. We can fix 1) by grounding in a tool like a datetime call or Google Search, but I want to focus on 2) because it clearly shows that the model doesn’t remember what he just said and demonstrates the point above that it is **our** responsibility to keep the model up to date on the conversation.

Let’s make a small modification to the code:

```python
response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents="what day is today?"
)
print(response.text)

# each element in the contents array is usually referred as a "turn"
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

Output:

```
$ python3 main.py 
Today is Wednesday, November 15, 2023.

Today is Wednesday, November 15, 2023.
```

Note that on the second call to the model we are including the whole context in the contents attribute. Also note that the role from each part changes from “user” to “model” and then to “user” again (“user” and “model” are the only possible values for role). That’s how the model understands which point of the conversation it is, a.k.a. “turn”. If, for example, we omitted the last part that repeats the question, the model would think it’s up to date and wouldn’t produce another response, as the last turn would be from “model” and not “user”.

The contents variable above is written in the “dictionary” form, but the SDK also provides several convenience methods like `types.UserContent` (sets the role field to “user” automatically) and `types.Part.from_text` (converts a plain string into a part), among others.

To deal with other types of inputs and/or outputs, we can use other types of parts like function calls, binary data, etc. If a model is multi-modal, you can mix parts of different content types in the same message.

Binary data can be both inline or be fetched from an URI. You can differentiate between different types of data using the mime_type field. For example, an image part can be retrieved like this:

```python
from google.genai import types

contents = types.Part.from_uri(
  file_uri: 'gs://generativeai-downloads/images/scones.jpg',
  mime_type: 'image/jpeg',
)
```

Or inlined:

```python
contents = types.Part.from_bytes(
  data: my_cat_picture, # binary data
  mime_type: 'image/jpeg',
)
```

In summary, for every turn of the conversation, we will be adding a new line of content for both the previous model response and the new user question.

The good news is that the chatbot experience is such an important use case that the Vertex AI SDK provides an implementation for this flow out of the box. Using the `chat` feature, we can reproduce the behavior above in very few lines of code:

```python
chat = client.chats.create(model='gemini-2.0-flash')
response = chat.send_message('what day is today?')
print(response.text)
response = chat.send_message('what day is today?')
print(response.text)
```

Output:

```
$ python3 main.py 
Today is Saturday, October 14th, 2023.

Today is Saturday, October 14th, 2023.
```

This time the model remembered the date because the chat interface is handling the history automatically for us.

## Non-automatic function calling

Now that we saw how the API works to build client messages and manage context, it’s time to explore how it deals with function calls. At a basic level, we will need to instruct the model that it has a function at its disposition and then process its requests to call the function and return the resulting values to the model. This is important because function calls allow agents to interact with external systems and the real world, creating actions such as retrieving data or triggering specific processes, going beyond just generating text. 

The function declaration is what tells the model what it can do. It tells the model the function name, description and its arguments. For example, below is a function declaration for the `get_random_number` function:

```python
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Returns a random number",
}
```

It is this declaration that the model needs to know to decide which functions to call. The function declaration has three fields: name, description and parameters - in this case the function doesn’t accept parameters so this field is omitted. The model uses the function description and the description of its arguments to decide when and how to call each function.

In the previous article, instead of giving the model a function declaration, I was lazy and let the SDK figure it out for me based on the docstring of my function. This time, we are going to do differently and explicitly declare a function to get a better understanding of the underlying flow.

The function including its declaration look like this:

```python
def get_random_number():
    return 4 # chosen by fair dice roll
             # guaranteed to be random (https://xkcd.com/221/)

# the declaration tells the model what it needs to know about the function
get_random_number_decl = {
    "name": "get_random_number",
    "description": "Returns a random number",
}
```

You can see other examples of function declarations [here](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling#schema-examples?utm_campaign=CDR_0x72884f69_awareness_b422727650&utm_medium=external&utm_source=blog).

Next we need to tell the model it has access to this function. We do this through the model configuration, adding the function as a tool.

```python
tools = types.Tool(function_declarations=[get_random_number_decl])
config = types.GenerateContentConfig(tools=[tools])

# my initial prompt
contents = [types.Part.from_text(text="what is my lucky number today?")]

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config, # note how we are adding the config to the model call
)

print(response.candidates[0].content.parts[0])
```

If you run the code above you will get something like this:

```
$ python3 main.py 
video_metadata=None thought=None inline_data=None file_data=None thought_signature=None code_execution_result=None executable_code=None function_call=FunctionCall(id=None, args={}, name='get_random_number') function_response=None text=None
```

What you are seeing here is the first part of the model response, and we can see that this part has all fields empty (`None`) except for the `function_call` field. This means the model wants **us** to make this function call, and then return its result back to the model.

This initially puzzled me, but if you think about it makes total sense. The model knows that the function exists, but has absolutely no idea how to call it. From the models perspective, the function is not running on the same machine either, so the model cannot do anything except “politely asking” us to do the call on their behalf.

We didn’t have to do this in my previous article because Automatic Function Calling took over and simplified things for us. The call still followed the same flow, but the SDK had hidden all this complexity from us.

The obvious thing to do now is to call the actual function and return the result to the model, but remember, without context the model knows nothing about our previous request, so if you only send the function results back it has no idea of what to do with it!

That’s why we need to send the history of the interaction up to now, and at least as far back as the point where the model knows it requested that value. The code below assumes we got a function call message and we need to send a new request with the complete information:

```python
# assuming we already inspected the response and know what the model wants
result = get_random_number() # makes the actual function call

# contents still contain the original prompt so we will add the model response...
contents.append(types.ModelContent(parts=response.candidates[0].content.parts))
# ... and the result of the function call
contents.append(types.UserContent(parts=types.Part.from_function_response(name="get_random_number", response={"result": result})))

response = client.models.generate_content(
    model="gemini-2.0-flash",
    contents=contents,
    config=config,
)
print(response.text)
```

Output:

```
$ python3 main.py 
Today's lucky number is 4.
```

## Conclusions

In this article we have seen how the agent client communicates with the model on the server-side or, in other words, the “domain model” of LLM communications. We also removed the curtain on the “magic” that the Python SDK does for us.

Automation is always convenient and helps us achieve results much faster, but knowing how it actually works is usually the big difference between a smooth journey and a patchy one when implementing your own agent, specially because the special cases are _never that easy_.

I know that in times of vibe coding at first glance it is almost ironic to say something like this, but one of the things I quickly learned when vibe coding is that if you are more precise when speaking with the AI you get much better results in much less time. So now is not the time to downplay the value of knowledge, but double down on it - not besides AI but **because of** it.

Hopefully you enjoyed the journey so far. In the next article we will build up on this knowledge to take the diagnostic agent to the next level, where no agent has ever gone before! (or maybe has, but certainly not mine =^_^=)

Please write your comments below! Peace out o/