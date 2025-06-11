+++
date = '2025-06-11T00:00:00+01:00'
title = 'Boldly Prompting: A Practical Guide to System Instructions and Agent Tools'
summary = "This article explores the concepts of system instruction, session history and agent tools to create a smarter diagnostic assistant."
tags = ["gemini", "vertex ai", "python"]
+++
## Introduction

In this guide, we are going to learn more about system prompts and agent tools so that we can build a new and improved diagnostic agent experience. We will be working with the [Vertex AI SDK for Python](https://cloud.google.com/vertex-ai/docs/python-sdk/use-vertex-ai-python-sdk?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog), LangChain, Gemini and [osquery](https://www.osquery.io/).

I must admit, the [initial version of the diagnostic agent](https://danicat.dev/posts/20250531-diagnostic-agent/) was not very “Enterprise” ready (pun intended). We didn’t have much visibility on what it was doing under the hood (was it actually running any queries?), it would not remember things discussed in the same “session” and, from time to time, it would also ignore our commands completely.

This is far from the experience we wish from a proper agent. The ideal diagnostic agent needs to be capable of remembering their mistakes and executing instructions consistently, e.g., learning that certain columns are not available and working around that. Also, can we really trust that it is doing what it is saying it is doing? We should be able to see the queries at any time to make sure the information it is returning is correct and up to date.

With these goals in mind, let’s get our hands dirty and start building our Emergency ~~Medical Hologram~~ Diagnostic Agent!

## Setting the Stage

Last time we wrote the code in a Jupyter notebook for the sake of convenience, but this time we are going to write a regular Python program. The same code would also work on Jupyter with minimal changes, but we are doing this so that we can use the diagnostic agent with a proper chat interface.

For any Python projects, I always recommend starting with a clean virtual env to keep dependencies self contained:

```
$ mkdir -p ~/projects/diagnostic-agent
$ cd ~/projects/diagnostic-agent
$ python3 -m venv venv
$ source venv/bin/activate
$ pip install --upgrade google-cloud-aiplatform[agent_engines,langchain]
```

Here is the initial version of `main.py` that reproduces the agent in the previous article:

```py
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown
import os

PROJECT_ID = os.environ.get("GCP_PROJECT")
LOCATION = os.environ.get("GCP_REGION", "us-central1")
STAGING_BUCKET = os.environ.get("STAGING_BUCKET_URI")

vertexai.init(
    project=PROJECT_ID,
    location=LOCATION,
    staging_bucket=STAGING_BUCKET
)

MODEL = os.environ.get("GEMINI_MODEL", "gemini-2.0-flash")

instance = osquery.SpawnInstance()

def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()  # This may raise an exception

    result = instance.client.query(query)
    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()  # This may raise an exception
    
    response = instance.client.query("select name from sqlite_temp_master").response
    tables = [ t["name"] for t in response ]
    return f"""
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
    """

def main():
    agent = agent_engines.LangchainAgent(
        model = MODEL,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
        ],
    )
    
    console = Console()

    print("Welcome to the Emergency Diagnostic Agent\n")
    print("What is the nature of your diagnostic emergency?")

    while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query)
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)

    print("Goodbye!")

if __name__ == "__main__":
    main()
```

You can run the agent with `python main.py`:

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

How can I help you?
>> 
```

There are two minor changes when compared to the original code: first, now we have a main loop which will keep the agent running until the user types “exit” or “quit”. This will create our chat interface.

Second, we have tweaked  the system prompt to improve the agent consistency. We are now calling it “Emergency Diagnostic Agent” - this name not only works as a neat [Star Trek easter egg](https://en.wikipedia.org/wiki/The_Doctor_(Star_Trek:_Voyager)), but more importantly, it sets a tone of urgency that, based on [emerging research](https://arxiv.org/pdf/2307.11760), may encourage it to comply with our requests more diligently. (Also check this [recent interview](https://www.reddit.com/r/singularity/comments/1kv7hm2/sergey_brin_we_dont_circulate_this_too_much_in/))

We are not going to threaten our poor Emergency Diagnostic Agent - and I can guarantee that no agents were harmed in the making of this text - but, calling it an “Emergency” agent should set the tone so that it will try to comply with our requests to the best of our ability. In the previous version of the system prompt I got cases when the agent refused to do a task because it “thought” it was not capable of doing it, or didn’t know which tables to query.

Of course, calling it an emergency agent is not enough to ensure the desired behaviour, so we added a few more instructions to guide the model behaviour, as we will see below.

## System Instructions

System instructions, also called system prompt, are a set of instructions that guide the behaviour of the LLM during the entire conversation. System instructions are special as they have higher priority over regular chat interactions. A way to imagine this is as if the system instructions were always repeated together with the prompt you are sending to the model.

There isn’t a strong consensus in the literature of what a system prompt should look like, but we have a few battle tested patterns that are emerging from day to day use. For example, one that is practically a consensus is to dedicate the very beginning of the system prompt to assign the agent a role, so it is aware of its purpose and can produce more coherent answers.

For this particular agent, I’ve opted to include the following sections in my system prompt: role, tools, context and task. This structure worked well during my testing phase, but don’t get too attached to it: experiment with your prompts and see if you can get better results. Experimentation is key to achieving good results with LLMs.

Now let’s have a look at each section of the prompt.

### System Prompt: Role

A role is nothing more than the reason for the agent to exist. It could be as simple as “you are a software engineer” or “you are a diagnostic agent”, but they can be a bit more elaborate including a detailed description, behavioural rules, restrictions and others. 

For a large language model that is trained on all kinds of data, the role helps setting the tone for the domain knowledge it will need to access to answer your queries. In other words, it gives semantic meaning to your queries… Imagine the question “what are cookies?” for example. Are we talking about edible cookies or browser cookies? If the agent role is undefined this question is completely ambiguous, but once we set the role to something technical (e.g. “you are a software engineer”), the ambiguity disappears.

For this agent, the role is described as:

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
```

Beyond the straightforward definition (“you are the emergency diagnostic agent”) we added a longer description to set the tone for the model behaviour and hopefully influence it to take our requests “seriously”, as mentioned before, the previous iteration of this agent had the bad tendency of refusing requests.

### System Prompt: Tools

Tools is the section that explains to the agent its capabilities to interact with external systems beyond its core model. Tools can be of several kinds, but the most common way to provide a tool to the agent is through function calls.

The agents can use tools to retrieve information, execute tasks and manipulate data. The Vertex AI SDK for Python has support for both user provided functions and built-in tools like Google Search and code execution. You can also use community maintained extensions through the Model Context Protocol (MCP) interface.

For our agent, we need to tell it that it can call _osquery_:

```
Tools:
  - you can call osquery using the call_osquery function.
```

### System Prompt: Context

Next we have the context, which tells the agent how about the environment it operates. I use this section to explicitly call out and correct undesired behaviours that previous iterations of the agent were prone to do. For example, I noticed very early in development that the agent would try to “guess” which tables were available and send queries blindly, resulting in a high error rate. Adding the list of tables to the context helped mitigate that problem.

Similar is the tendency for the agent to try to guess the column names in a table, instead of trying to discover the names first. In this particular case I resisted the temptation to instruct the agent to always use SELECT \* because this is a bad practice (it retrieves more data than you need), but instead I “taught” it how to discover a schema using the PRAGMA instruction.

This way the agent will still make mistakes while guessing column names, but it has a way to course correct without human intervention.

The revised context section of the system prompt is shown below.

```
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
```

Note that tables is a variable that contains all the tables we discovered from osquery before starting the model.

### System Prompt: Task

Finally, the task. This section is used to describe how the agent should interpret your requests and execute them. People usually use this section to lay out the steps required to achieve the task at hand.

In our particular case, we are using this section to roughly lay out the plan, but also add a few conditional directives:

```
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

The step “confirm the plan with the user before executing” is interesting as it shows us how the agent is thinking about the process, but it might be a bit annoying after interacting with the agent for a while. We can always ask for the agent to tell us the plan with a prompt, so the inclusion of this step is entirely optional.

I initially thought of this step as a way to debug the agent, but in the following section we are going to explore a different way to do it.

With the combination of these four sections we have the entire system prompt. This revised prompt produced more consistent results during my tests in preparation for this article. It also has the benefit of being “human-friendly” so it is easier to adapt when new rules are introduced.

Here is the complete view of this system prompt:

```
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
```

As a side note, I believe that we still have many opportunities to improve it and one of my current areas of interest is how to achieve a self-improving system prompt. This could potentially be achieved by asking, at the end of a session, for the model to summarise its learnings into a new system prompt for its future iteration. The prompt could be stored in a database and loaded on the next session. This of course raises concerns for system prompt degradation or, even worse, attacks using prompt injection, so it is not as trivial as it seems. Nevertheless, it is a fun exercise and I might write about it in the near future.

## Enabling Debug Mode

Another concern about the original design is the lack of observability about what the agent is doing under the hood. There are two different approaches we can apply here, one slightly more painful than the other: 1) peek into the “thoughts” of the LLM and try to find the tool calls among them (very painful), or; 2) add some debug functionality to the function itself so it outputs the information we want during runtime (the easiest solution is usually the right one).

I must admit, I spent an unhealthy amount of time in option 1, before realising I could do option 2. If you really want to go through the path of LLM reasoning, you can do so through a configuration called [return_intermediate_steps](https://api.python.langchain.com/en/latest/agents/langchain.agents.agent.AgentExecutor.html#langchain.agents.agent.AgentExecutor.return_intermediate_steps). I should say this is very interesting from a learning point of view, but after spending a couple of hours trying to figure out the format of the output (hint: [it is not really json](https://github.com/langchain-ai/langchain/issues/10099)) I decided that parsing it was not really worth the trouble.

So, how does the simple strategy work? We are adding a debug flag and a tool to toggle that flag on and off. This surprisingly simple trick actually opens an entire new world of potential: we are giving the agent an opportunity to modify its own behaviour!

The implementation of debug mode is composed by a global variable and a function to set it:

```py
debug = False

def set_debug_mode(debug_mode: bool):
    """Toggle debug mode. Call this function to enable or disable debug mode.
    
    Args:
        debug_mode (bool): True to enable debug mode, False to disable it.


    Returns:
        None
    """
    global debug
    debug = debug_mode
```

We also need to mention it in the system prompt:

```
...
   Tools:
    - you can call osquery using the call_osquery function.
    - you can enable or disable the debug mode using the set_debug_mode function.
    Context:
...
```

And add the function to the tool list in the agent instantiation:

```py
   agent = agent_engines.LangchainAgent(
        model = model,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode,
        ],
    )

```

Finally, we need to adapt `call_osquery` to use the new `debug` flag:

```py
def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()

    if debug:
        print("Executing query: ", query)

    result = instance.client.query(query)
    if debug:
        print("Query result: ", {
            "status": result.status.message if result.status else None, 
            "response": result.response if result.response else None
        })

    return result
```

With all of these changes in place, let’s have a look at how the agent calls _osquery_ using the newly implemented debug flag:

```
$ python main.py
Welcome to the Emergency Diagnostic Agent

How can I help you?
>> run a level 1 diagnostic procedure in debug mode
Executing query:  SELECT * FROM system_info
Query result:  {'status': 'OK', 'response': [{...}]}
Executing query:  SELECT pid, name, user, cpu_percent FROM processes ORDER BY cpu_percent DESC LIMIT 10
Query result:  {'status': 'no such column: user', 'response': None}
Executing query:  SELECT pid, name, user, resident_size FROM processes ORDER BY resident_size DESC LIMIT 10
Query result:  {'status': 'no such column: user', 'response': None}
Executing query:  PRAGMA table_info(processes)
Query result:  {'status': 'OK', 'response': [{'cid': '0', 'dflt_value': '', 'name': 'pid', 'notnull': '1', 'pk': '1', 'type': 'BIGINT'}, ...]}
(...)

System Information:                                                                                                                                                     

 • Hostname: petruzalek-mac.roam.internal                                                                                                                               
 • CPU Type: arm64e                                                                                                                                                     
 • Physical Memory: 51539607552 bytes                                                                                                                                   

Top 5 Processes by CPU Usage:                                                                                                                                          
                                                  
  PID     Name                         CPU Usage  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 
  1127    mediaanalysisd               95627517   
  43062   mediaanalysisd-access        66441942   
  54099   Google Chrome                3005046    
  54115   Google Chrome Helper (GPU)   2092500    
  81270   Electron                     1688335    

Top 5 Processes by Memory Usage (Resident Size):                                                                                                                       
                                                                   
  PID     Name                              Resident Size (Bytes)  
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 
  43062   mediaanalysisd-access             3933536256             
  54099   Google Chrome                     1313669120             
  59194   Code Helper (Plugin)              1109508096             
  59025   Code Helper (Renderer)            915456000              
  19681   Google Chrome Helper (Renderer)   736329728                         
                                                                   
>> 
```

Notice that the command issued was “run a level 1 diagnostic procedure **in debug mode**” which showcases an interesting capability of the agent: multi-tool invocation. If it deems necessary, it is able to invocate not only the same tool multiple times, but different tools at the same time as well. So it was not necessary to enable debug mode before requesting the report: the agent was able to do it all in one go.

Also notice how the agent initially failed when requesting a user column, but then used the PRAGMA instruction to discover the correct schema and retry the query successfully. This is a perfect demonstration of the agent's ability to recover from errors due to our improved system prompt.

## Preserving the Chat History

Our final task today is to ensure the agent remembers what we are talking about so that we can ask clarifying questions and further probe the system following a coherent line of investigation. 

In the [previous article](https://danicat.dev/posts/20250605-vertex-ai-sdk-python/) we explored how LLMs are stateless and that we need to keep “reminding” them of the current state of the conversation using “turns”. Luckily with LangChain we don’t need to do this manually and we can rely on a feature called [chat history](https://python.langchain.com/api_reference/core/chat_history.html).

The beauty of chat history is that anything that implements [BaseChatMessageHistory](https://python.langchain.com/api_reference/core/chat_history/langchain_core.chat_history.BaseChatMessageHistory.html#langchain_core.chat_history.BaseChatMessageHistory) can be used here, which allows us to use all sorts of data stores, including creating our own. For example, in the official documentation for Vertex AI you can find examples for using [Firebase, Bigtable and Spanner](https://cloud.google.com/vertex-ai/generative-ai/docs/agent-engine/develop/langchain#chat-history?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog).

We don’t need a full fledged database for the moment so we are going to settle with InMemoryChatMessageHistory, which as the name	suggests will store everything in memory.

Here is a typical implementation, technically supporting multiple sessions using the `chats_by_session_id` dictionary for lookup (code copied from the [langchain documentation](https://python.langchain.com/docs/versions/migrating_memory/chat_history/#chatmessagehistory)):

```py
chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history
```

And here is our new `main` function instantiating the agent with chat history enabled:

```py
import uuid

def main():
    session_id = uuid.uuid4()
    agent = agent_engines.LangchainAgent(
        model = model,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode,
        ],
        chat_history=get_chat_history,
    )
```

A quick callout so that you don’t make the same mistake as me: the `chat_history` argument expects a `Callable` type, so you should not invoke the function there, but pass the function itself. LangChain uses a factory pattern here; it invokes the provided function (`get_chat_history`) on demand with a `session_id` to get or create the correct history object. This design is what enables the agent to manage multiple, separate conversations concurrently.

The function signature can either include one or two arguments. If one argument, it is assumed to be a `session_id`, and if it is two arguments they are interpreted as `user_id` and `conversation_id`. More information about this can be found in the [RunnableWithMessageHistory](https://python.langchain.com/api_reference/core/runnables/langchain_core.runnables.history.RunnableWithMessageHistory.html) documentation.

The last piece of the puzzle is passing the session_id to the model runner. This is made through the config argument as shown in the code below:

```py
# (...)
   while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query, config={"configurable": {"session_id": session_id}})
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)
```

Now as long as the session is alive we can ask the agent about information in its “short-term memory”, as the session contents are stored in memory. This will be enough for most basic interactions to feel more natural, but we are opening precedence for bigger problems: now that we can store session information, after each iteration it will only grow, and while dealing with data generated automatically from queries, the session context will grow very fast soon hitting the limits of the model, and long before we hit the limits of our computer memory.

Models like Gemini are well known for their [long context windows](https://ai.google.dev/gemini-api/docs/long-context), but even a million tokens can be exhausted really fast if we fill the context with data. Long context can also pose a problem for some models as retrieval gets harder and harder - also known as the [needle in a haystack](https://cloud.google.com/blog/products/ai-machine-learning/the-needle-in-the-haystack-test-and-how-gemini-pro-solves-it?utm_campaign=CDR_0x72884f69_awareness_b424142426&utm_medium=external&utm_source=blog) problem.

There are techniques for addressing the growing context problem, including compression and summarisation, but for the sake of keeping the context of this article short (see what I did there?), we are going to save those for the next article.

The final version of `main.py` including all modifications in this article looks like this:

```py
import vertexai
from vertexai import agent_engines
import osquery
from rich.console import Console
from rich.markdown import Markdown
from langchain_core.chat_history import InMemoryChatMessageHistory
import os
import uuid

PROJECT_ID = os.environ.get("GCP_PROJECT")
LOCATION = os.environ.get("GCP_REGION", "us-central1")
STAGING_BUCKET = os.environ.get("STAGING_BUCKET_URI")

vertexai.init(
    project=PROJECT_ID,
    location=LOCATION,
    staging_bucket=STAGING_BUCKET
)

MODEL = os.environ.get("GEMINI_MODEL", "gemini-2.5-pro-preview-05-06")

instance = osquery.SpawnInstance()
debug = False

def set_debug_mode(debug_mode: bool):
    """Toggle debug mode. Call this function to enable or disable debug mode.
    
    Args:
        debug_mode (bool): True to enable debug mode, False to disable it.


    Returns:
        None
    """
    global debug
    debug = debug_mode

def call_osquery(query: str):
    """Query the operating system using osquery
      
    This function is used to send a query to the osquery process to return information about the current machine, operating system and running processes.
    You can also use this function to query the underlying SQLite database to discover more information about the osquery instance by using system tables like sqlite_master, sqlite_temp_master and virtual tables.

    Args:
        query: str  A SQL query to one of osquery tables (e.g. "select timestamp from time")

    Returns:
        ExtensionResponse: an osquery response with the status of the request and a response to the query if successful.
    """
    if not instance.is_running():
        instance.open()  # This may raise an exception

    if debug:
        print("Executing query: ", query)

    result = instance.client.query(query)
    if debug:
        print("Query result: ", {
            "status": result.status.message if result.status else None, 
            "response": result.response if result.response else None
        })

    return result

def get_system_prompt():
    if not instance.is_running():
        instance.open()  # This may raise an exception
    
    response = instance.client.query("select name from sqlite_temp_master").response
    tables = [ t["name"] for t in response ]
    return f"""
Role:
  - You are the emergency diagnostic agent. 
  - You are the last resort for the user to diagnose their computer problems. 
  - Answer the user queries to the best of your capabilities.
Tools:
  - you can call osquery using the call_osquery function.
  - you can use the set_debug_mode function to enable or disable debug mode.
Context:
  - Only use tables from this list: {tables}
  - You can discover schemas using: PRAGMA table_info(table)
Task:
  - Create a plan for which tables to query to fullfill the user request
  - Confirm the plan with the user before executing
  - If a query fails due a wrong column name, run schema discovery and try again
  - Query the required table(s)
  - Report the findings in a human readable way (table or list format)
    """

chats_by_session_id = {}

def get_chat_history(session_id: str) -> InMemoryChatMessageHistory:
    chat_history = chats_by_session_id.get(session_id)
    if chat_history is None:
        chat_history = InMemoryChatMessageHistory()
        chats_by_session_id[session_id] = chat_history
    return chat_history

def main():
    session_id = uuid.uuid4()
    agent = agent_engines.LangchainAgent(
        model = MODEL,
        system_instruction=get_system_prompt(),
        tools=[
            call_osquery,
            set_debug_mode
        ],
        chat_history=get_chat_history,
    )
    
    console = Console()

    print("Welcome to the Emergency Diagnostic Agent\n")
    print("What is the nature of your diagnostic emergency?")

    while True:
        try:
            query = input(">> ")
        except EOFError:
            query = "exit"

        if query == "exit" or query == "quit":
            break

        if query.strip() == "":
            continue
            
        response = agent.query(input=query, config={"configurable": {"session_id": session_id}})
        rendered_markdown = Markdown(response["output"])
        console.print(rendered_markdown)

    print("Goodbye!")

if __name__ == "__main__":
    main()
```

## Conclusions

In this article we have learned the importance of fine tuning a system prompt to achieve consistent responses from an agent. We also have seen in practice how multi-tool calling operates, and how to use tools to enable or disable feature flags to change the behaviour of the agent. Last but not least, we have learned about how to manage session state using the in-memory chat history.

In the next article of the series we will see how to enable persistence between sessions using a real database, revisit the notion of tokens and discuss the context compression technique.

## Appendix: Fun Things to Try

Now that our agent is more robust, use this section as a practical guide to test the new features in action. Notice how it now remembers context between questions and how you can ask it to explain its work.

```
>> run a level 1 diagnostic procedure
>> run a level 2 diagnostic procedure
>> explain the previous procedure step by step
>> find any orphan processes
>> show me the top resource consuming processes
>> write a system prompt to transfer your current knowledge to another agent
>> search the system for malware
>> is this computer connected to the internet?
>> why is my computer slow?
>> take a snapshot of the current performance metrics
>> compare the current perfomance metrics with the previous snapshot
>> give me a step by step process to fix the issues you found
>> how many osqueryd processes are in memory?
>> give me a script to kill all osqueryd processes
>> who am i?
```

If you find other interesting prompts, please share your experiences in the comments section below. See you next time!
