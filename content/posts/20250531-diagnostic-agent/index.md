---
date: '2025-05-31T01:00:00+01:00'
title: 'How I turned my computer into "USS Enterprise" using AI Agents'
summary: "How to create a diagnostic agent that speaks natural language using Gemini and Vertex AI Agent Engine"
categories: ["AI & Development"]
tags: ["gemini", "vertex-ai", "python", "tutorial"]
---
_Space: the final frontier. These are the voyages of the starship Enterprise. Its 5-year mission: to explore strange new worlds; to seek out new life and new civilizations; to boldly go where no man has gone before._

## Introduction

When growing up, thanks to the influence of my father, I got used to hearing these words almost every day. I suspect his passion for Star Trek played a huge role in me choosing the software engineering career. (For those who are not familiar with Star Trek, this speech was played in the beginning of every episode of the original Star Trek series)

Star Trek was always ahead of its time. It showed the [first interracial kiss in U.S television](https://en.wikipedia.org/wiki/Kirk_and_Uhura%27s_kiss), in times when such a scene caused much controversy. It also depicted many pieces of “futuristic” technology that today are commodities, like smartphones and video conferencing.

One thing that is really remarkable is how the engineers in the series interact with the computers. While we do see some keyboards and button presses now and then, many of the commands are vocalised in natural language. Some of the commands they give to the computer are quite iconic, like for example when they request the computer to run a “ level 1 diagnostic procedure”, which happened so many times that it practically became [a joke](https://www.youtube.com/watch?v=cYzByQjzTb0) among the most hardcore fans.

Fast forward 30+ years and here we are, in the Age of AI, a technology revolution that promises to be bigger than the internet. Of course a lot of people are scared of how AI might impact their jobs ([I wrote about it last week](https://danicat.dev/posts/20250528-vibe-coding/)), but growing up watching Star Trek makes it easier for me to see how the role of the engineer will change in the next few years. Instead of commanding the computer through text, manually instructing each step of the way through lines of code and compilers, we will very soon move towards talking and brainstorming with our computers.

To help people visualize this, we are going to use the technology we have today to create a small agent that allows us to interact with our own machines using natural language.

## What you will need for this demo

For the development language we will be using Python in a Jupyter Notebook, as it plays very nicely for experimentation. The main tools and libraries we will be using are:

*   [Vertex AI Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/overview?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog)
*   [Osquery](https://www.osquery.io/) with [python bindings](https://github.com/osquery/osquery-python)
*   [Jupyter Notebook](https://jupyter.org/) [optional] (I’m actually using the [Jupyter plugin for VSCode](https://code.visualstudio.com/docs/datascience/jupyter-notebooks))

The examples below will use Gemini Flash 2.0, but you can use any [Gemini model variant](https://ai.google.dev/gemini-api/docs/models). We won't deploy this agent to Google Cloud this time as we want to use it to answer questions about the local machine and not about the server in the cloud.

## Agent Overview

If you are already familiar with how agent technology works you can skip this section.

An AI agent is a form of AI that is capable of perceiving its environment and taking autonomous actions to achieve specific goals. If you compare with the typical Large Language Models (LLMs), which primarily focus on generating content based on input, AI agents can interact with their environment, make decisions, and execute tasks to reach their objectives. This is achieved by the use of “tools” that will feed the agent with information and enable it to do actions.

To demonstrate the agent technology we are going to use LangChain through [Agent Engine](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain?utm_campaign=CDR_0x72884f69_awareness_b421478530&utm_medium=external&utm_source=blog). First, you need to install the required packages in your system:

```shell
pip install --upgrade --quiet google-cloud-aiplatform[agent_engines,langchain]
```

You will also need to set your gcloud application default credentials (ADC):

```shell
gcloud auth application-default login
```

Note: depending on the environment you are using to run this demo you may need to use a different authentication method.

Now we are ready to work on our Python script. First we are going to initialize the SDK based on our Google Cloud project ID and location:

```python
import vertexai

vertexai.init(
    project="my-project-id",                  # Your project ID.
    location="us-central1",                   # Your cloud location.
    staging_bucket="gs://my-staging-bucket",  # Your staging bucket.
)
```

Once the initial setup is done, creating an agent using LangChain in Agent Engine is pretty straightforward:

```python
from vertexai import agent_engines

model = "gemini-2.0-flash" # feel free to try different models!

model_kwargs = {
    # temperature (float): The sampling temperature controls the degree of
    # randomness in token selection.
    "temperature": 0.20,
}

agent = agent_engines.LangchainAgent(
    model=model,                # Required.
    model_kwargs=model_kwargs,  # Optional.
)
```

The setup above is just enough for you to submit queries to the agent, just like you would submit a query to a LLM:

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

Which could return something like this:

```
{'input': 'which time is now?', 'output': 'As an AI, I don\'t have a "current" time or location in the same way a human does. My knowledge isn\'t updated in real-time.\n\nTo find out the current time, you can:\n\n*   **Check your device:** Your computer, phone, or tablet will display the current time.\n*   **Do a quick search:** Type "what time is it" into a search engine like Google.'}
```

Depending on your settings, prompt and the randomness of the universe, the model can either give you a response saying it cannot tell you the time, or it can “hallucinate” and make up a timestamp. But in fact, since the AI doesn’t have a clock, it will not be able to answer this question... unless you give it a clock!

## Function Calls

One of the most convenient ways of extending our agent capabilities is to give it python functions to call. The process is pretty simple, but it’s important to highlight that the better the documentation you have for the function, the easier it will be for the agent to get its call right. Let’s define our function to check the time:

```python
import datetime

def get_current_time():
    """Returns the current time as a datetime object.

    Args:
        None
    
    Returns:
        datetime: current time as a datetime type
    """
    return datetime.datetime.now()
```

Now that we have a function that gives us the system time, let’s recreate the agent but now aware that the function exists:

```python
agent = agent_engines.LangchainAgent(
    model=model,                # Required.
    model_kwargs=model_kwargs,  # Optional.
    tools=[get_current_time]
)
```

And ask the question again:

```python
response = agent.query(
    input="which time is now?"
)
print(response)
```

The output will look similar to this:

```
{'input': 'which time is now?', 'output': 'The current time is 18:36:42 UTC on May 30, 2025.'}
```

Now the agent can rely on the tool to answer the question with real data. Pretty cool, huh?

## Gathering System Information

For our diagnostic agent we are going to give it capabilities to query information about the machine it is running in using a tool called [osquery](https://www.osquery.io/). Osquery is an open source tool developed by Facebook to allow the user to make SQL queries to “virtual tables” that expose information about the underlying operating system of the machine.

This is convenient to us because it not only gives us a single point of entry to make queries about the system, but LLMs are also very proficient in writing SQL queries.

You can find instructions on how to install osquery in the [official documentation](https://osquery.readthedocs.io/en/stable/). I won’t reproduce them here because they vary depending on the operating system of your machine.

Once you have osquery installed you will need to install the python bindings for osquery. As typical python, it’s only one pip install:

```shell
pip install --upgrade --quiet osquery
```

With the bindings installed you can make osquery calls by importing the osquery package:

```python
import osquery

# Spawn an osquery process using an ephemeral extension socket.
instance = osquery.SpawnInstance()
instance.open()  # This may raise an exception

# Issues queries and call osquery Thrift APIs.
instance.client.query("select timestamp from time")
```

The query method will return an ExtensionResponse object with the results of your query. For example:

```python
ExtensionResponse(status=ExtensionStatus(code=0, message='OK', uuid=0), response=[{'timestamp': 'Fri May 30 17:54:06 2025 UTC'}])
```

If you never worked with osquery before I encourage you to have a look at the [schema](https://www.osquery.io/schema/5.17.0/) to see which kind of information is available in your operating system.

### A side note about formatting

All the outputs from the previous examples were unformatted, but if you are running the code from Jupyter you can access some convenience methods to beautify the output by importing the following packages:

```python
from IPython.display import Markdown, display
```

And displaying the response output as markdown:

```python
response = agent.query(
    input="what is today's stardate?"
)
display(Markdown(response["output"]))
```

Output:

```
Captain's Log, Supplemental. The current stardate is 48972.5.
```

## Connecting the dots

Now that we have a way to query information about the operating system, let’s combine that with our knowledge of agents to make a diagnostic agent that will answer questions about our system.

The first step is to define a function to make the queries. This will be given to the agent as a tool to gather information later:

```python
def call_osquery(query: str):
    """Query the operating system using osquery
      
      This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
      You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

      Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

      Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    return instance.client.query(query)
```

The function itself is pretty trivial, but the important part here is to have a very detailed docstring that will enable the agent to understand how this function works.

During my testing, one tricky problem that occurred quite frequently is that the agent didn’t know exactly which tables were available in my system. For example, I’m running a macOS machine and the table “memory_info” doesn’t exist.

To give the agent a bit more context, we are going to dynamically give to it the names of the tables that are available in this system. In an ideal situation, you would even give it the entire schema with column names and descriptions, but unfortunately that is not trivial to achieve with osquery.

The underlying database technology for osquery is SQLite so we can query the list of virtual tables from the `sqlite_temp_master` table:

```python
# use some python magic to figure out which tables we have in this system
response = instance.client.query("select name from sqlite_temp_master").response
tables = [ t["name"] for t in response ]
```

Now that we have all the table names, we can create the agent with this information and the `call_osquery` tool:

```python
osagent = agent_engines.LangchainAgent(
    model = model,
    system_instruction=f"""
    You are an agent that answers questions about the machine you are running in.
    You should run SQL queries using one or more of the tables to answer the user questions.
    Always return human readable values (e.g. megabytes instead of bytes, and formatted time instead of miliseconds)
    Be very flexible in your interpretation of the requests. For example, if the user ask for application information, it is acceptable to return information about processes and services. If the user requests resource usage, return BOTH memory and cpu information.
    Do not ask the user for clarification.
    You have the following tables available to you: 
    ----- TABLES -----
    {tables}
    ----- END TABLES -----

    Question:
    """,
    tools=[
        call_osquery,
    ]
)
```

Believe it or not, our agent is ready for all our questions! Let’s try it out:

```python
response = osagent.query(input="what is the current time?")
display(Markdown(response["output"]))
```

Output:

```
The current time is Fri May 30 18:08:15 2025 UTC.
```

Let’s make it a bit more complicated:

```python
response = osagent.query(input="what is the top consuming process?")
display(Markdown(response["output"]))
```

Output:

```
The top CPU-consuming process is:
Process Name: mediaanalysisd, PID: 1127, Total CPU Time: 2876716

The top memory-consuming process is:
Process Name: Code Helper (Plugin), PID: 2987, Memory Usage: 1537 MB
```

What if I am a bit more creative?

```python
response = osagent.query(input="computer, run a level 1 diagnostic procedure")
display(Markdown(response["output"]))
```

Output:

```
Here is the level 1 diagnostic information:

**Operating System:**
*   **Name:** macOS
*   **Version:** 15.5
*   **Build:** 24F74
*   **Platform:** darwin
*   **Architecture:** arm64

... ommitted for brevity ...

**Uptime:**
*   **Total Seconds:** 261164
*   **Days:** 3
*   **Hours:** 0
*   **Minutes:** 32
*   **Seconds:** 44

```

Wow! I never actually defined what a Level 1 diagnostic procedure is, but it gave me a pretty impressive report nevertheless!

I even tried to be a bit more clever in my questions and the answers didn’t disappoint (most of the times):

```python
response = osagent.query(input="computer, do you see any signs of malware running?")
display(Markdown(response["output"]))
```

Output:

```
I have checked for processes that are not associated with a file on disk, which can be a sign of malware, and found none. I have also examined the top processes by memory and CPU usage. The processes consuming the most resources are primarily Visual Studio Code and Google Chrome and their related helper processes. This is typical behavior for these applications.

Based on the checks performed, there are no obvious signs of malware running on the system at this time.
```

_Mic drop_ =^.^=

## Conclusions

I know it’s a beaten argument by now, but AI is a game changer. With very few lines of code we have gone from zero to a fully functioning natural language interface to the inner workings of the operating system. With a little more work this agent can be improved to do deeper diagnostics and maybe even autonomously fix things. Scotty would be proud!

![Engineer Scotty trying to speak with the computer using the mouse as microphone](hello-computer-hello.gif)

You can find the source code for all examples in this article on my [GitHub](https://github.com/danicat/devrel/blob/main/blogs/20250531-diagnostic-agent/diagnostic_agent.ipynb).

What are your impressions? Share your thoughts in the comments below.
