---
title: "Mastering Hooks in Coding Agents"
date: 2026-06-10T00:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["antigravity", "agy-cli", "agentic-coding"]
summary: "Learn to use agent hooks to enforce guardrails, run linters, and automate memory in your coding loops."
heroStyle: "big"
---

Coding agent capabilities are moving fast. My first contact with them was roughly one year ago, just a little after I joined Google. Back then the big deal was the Model Context Protocol (MCP), a flexible technology replacing ad hoc implementations of tools with portable ones (among other things).

Fast forward twelve months, and today most people seem to have moved away from MCPs towards Agent Skills as the next cool thing. By introducing "progressive disclosure", skills allow for a more efficient usage of context, which results in better token economy overall. With rising inference costs, it is no surprise that skills became so popular.

Both MCP and skills have their own "fanbases", but there is another coding agent standard that emerged within this time period and is not as often mentioned as their more popular counterparts: the hooks.

While MCPs and skills are focusing on extending agentic capabilities (by adding tools and knowledge), hooks operate at a different level, giving you more control over the agent loop and the development process overall.

## Agent hooks explained

The name hook might not ring a bell at first, but hooks are simply "callbacks" — procedures that are called at specific moments during the agent processing lifecycle.

Hooks have three components:
- A trigger event: when the hook will be called. Typically composed of an execution phase (pre or post) and a context, such as a tool call or model invocation. For example, in Antigravity, we have events like `PreToolUse` and `PostInvocation`. Some engines also support a `Stop` hook (upon agent termination) and a `PreCompact` hook (before auto-compaction, which is unique to Claude-based implementations).
- A condition or filter: a regular expression based on the trigger event. For example, in a tool call, the filter can be the tool name and may include the tool arguments. For example, it is possible to create a hook for the tool call `run_command(git)`.
- A procedure: the body of the hook, either as a shell script or command. The procedure can be used to either allow or disallow an operation, to override model or tool calls completely, and to produce side effects like logging or telemetry.

## When to use hooks

Hooks intercept the agent lifecycle at specific moments to inject custom commands or scripts. By intercepting the right moment, you can control the flow of operation and add deterministic outcomes to what otherwise is mostly a non-deterministic process.

For example, developers often attempt to enforce coding guidelines via either a system prompt or an AGENTS.md file (or similar). However, prompt-based guidelines offer no execution guarantees due to the non-deterministic nature of large language models: the exact same prompt can produce different outcomes, and agents can selectively ignore parts of the prompt.

Using hooks instead of prompts you can enforce a specific action. Let's say for example that you want to ensure your agent always run a static analysis tool in the code (a.k.a. a linter) after every code edit to ensure the code is always sanitised. If you add "always run a linter after edits" to your prompts, the agent has the "agency" to decide if they are going to run a linter or not, and the agent may skip validation entirely if it deems the edit "trivial." But if you create a hook instead — in this case a `PostToolUse` filtering for the file editing tool — you can deterministically ensure your static analysis tool will run after every edit.

By intercepting these lifecycles, we can implement several powerful patterns to control agent behaviour, gather metrics, or keep workflows secure.

### Steer the agent towards specialised tools

Hooks are useful in many scenarios, but my favourite use case is to put guardrails around the agent, or as sometimes I like to say: "constraining agent autonomy".

We can implement this by pairing a `PreToolUse` hook with a script denying the tool access and returning a "steering hint" to the coding agent. This steering hint will contain the instructions you want it to perform instead. For example, let's say you want to prevent the agent from using shell commands to read Go files, a steering hint could look like this: "Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead".

### Collect telemetry data

The hook system is also a good place to put your telemetry collectors and loggers, as it gives you good visibility of the inner workings of the agent.

### Intercept malicious prompts

A `PreInvocation` hook can intercept incoming prompts and evaluate them against safety heuristics or lighter classifier models. If a prompt looks like a jailbreak attempt, the hook can block the request immediately, shielding backend systems before they reach the execution loop.

### Prevent credential leakage

Developers sometimes accidentally paste env files or credentials into active files that the coding agent reads. A `PostToolUse` hook listening to file reads — or a `PreInvocation` hook scanning the payloads sent to the LLM — acts as a reliable Data Loss Prevention (DLP) gate. If the hook detects strings matching high-entropy keys or standard API formats, it can redact the secrets dynamically or abort the run to keep credentials secure.

### Managing memories

Agents are typically stateless unless they are plugged into an external memory system like [Agent Platform Memory Bank](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank) or [MemPalace](https://github.com/mempalace/mempalace).

One way to add memory capabilities to agents is by registering memorization and retrieval as tools, but doing so we depend on the agent explicitly making a decision to call the corresponding tools.

The hook system allows you to automate the memory persistence and retrieval. You can connect the [memory generation](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/generate-memories#triggering-memory-generation) to the end of a session (using a `Stop` hook) or after a certain number of turns (by monitoring step number or number of model invocations).

Conversely, memory retrieval can be done automatically at the start of the session and before invoking models (e.g., a `PreInvocation` hook). For example, in Agent Platform Memory Bank you can retrieve memories by [scope](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#retrieve-all) (e.g., scope could be a user ID) or [similarity](https://docs.cloud.google.com/gemini-enterprise-agent-platform/scale/memory-bank/fetch-memories#similarity-search) (based on a query). This is essentially a memory-based retrieval augmented generation (RAG).

## Configuring hooks in Antigravity

While different agent engines implement their own vocabulary for callbacks, in this section we are going to focus on the specific **Antigravity dialect** of hooks.

For a full look at the specifications, check out the official [Antigravity Hooks Documentation](https://antigravity.google/docs/hooks). 

Antigravity looks for a `hooks.json` file inside your workspace's `.agents/` directory (or globally in your home directory at `~/.gemini/config/hooks.json`).

Here is an example of how to implement the steering hints and static analysis as discussed above:

```json
{
  "linter-safety-gate": {
    "PostToolUse": [
      {
        "matcher": "write_to_file|replace_file_content|multi_replace_file_content",
        "hooks": [
          {
            "type": "command",
            "command": "./scripts/run-linter.sh",
            "timeout": 15
          }
        ]
      }
    ]
  },
  "restrict-cat-on-go": {
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "command": "./scripts/steer-go-reads.py"
          }
        ]
      }
    ]
  }
}
```

The inputs to these hooks are delivered via `stdin` as a JSON object, containing context like the tool arguments (`toolCall.args`), active `workspacePaths`, and the filepath of the current session log (`transcriptPath`). Your scripts can evaluate these, perform computations, and print a JSON response on `stdout` telling Antigravity whether to `"allow"`, `"deny"`, or `"ask"` for user confirmation.

For example, here is how you can write a simple Python script (`steer-go-reads.py`) to parse that incoming payload and steer the agent:

```python
import sys
import json

def main():
    # Read and parse the incoming trigger event payload from stdin
    try:
        payload = json.load(sys.stdin)
    except Exception as e:
        # Standard safety gate fallback
        print(json.dumps({
            "decision": "deny",
            "reason": f"Failed to parse stdin payload: {e}"
        }))
        return

    tool_call = payload.get("toolCall", {})
    tool_name = tool_call.get("name")
    tool_args = tool_call.get("args", {})

    # Match the specific tool and check arguments
    if tool_name == "run_command":
        command_line = tool_args.get("CommandLine", "")
        
        # Detect if command attempts to cat any Go source files
        if "cat" in command_line and ".go" in command_line:
            response = {
                "decision": "deny",
                "reason": "Tool call blocked - run_command(cat): do not use 'cat' for reading *.go files, use 'smart_read' instead."
            }
            print(json.dumps(response))
            return

    # Default to allow if no rules are violated
    print(json.dumps({
        "decision": "allow"
    }))

if __name__ == "__main__":
    main()
```


## Regaining control

Agents are getting smarter and faster, allowing us to produce code faster than ever before. But speed without control is the recipe for disaster. I often use this analogy on my talks: if you like fast cars, the most important thing you need to care about is not the engine, but the brakes. If the brakes are less powerful than the engine, you won't be able to stop and your safety is compromised.

The same thinking should be applied to coding agents. If you want to write code fast, you need a powerful control system that ensures you are not sacrificing quality and introducing bugs, because if you do your application is going to have big problems sooner or later.

Hooks are a good place to implement the guardrails that can put us back in control, bridging the gap between AI agency and robust software engineering. As I read recently in [this article by Joe Bertolami](https://venturebeat.com/technology/agentic-ai-solved-coding-and-exposed-every-other-problem-in-software-engineering): "writing code was never the rate limiter". Let's not forget decades of engineering best practices and equip our agents with the right tools for the job, so that we can fully enjoy the modern agentic coding experience.