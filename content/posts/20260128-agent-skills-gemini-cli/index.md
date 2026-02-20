---
title: "Mastering Agent Skills in Gemini CLI"
date: 2026-01-29T12:00:00Z
summary: "Unlock on-demand expertise for your AI agent. Learn how to use Agent Skills in Gemini CLI to build modular, scalable, and autonomous workflows."
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["gemini-cli", "agent-skills", "mcp", "vibe-coding"]
slug: "agent-skills-gemini-cli"
---

When I wrote about [Tenkai]({{< ref "/posts/20260120-improving-agentic-coding-with-science/" >}}) last week I didn't cover one important aspect about experiment analysis: how to extract insights from experiments. Even though I have a nice frontend with summaries and statistical metrics and tests, it is really hard to capture the nuances of each configuration from just a summary.

For example, I often find that read operations (e.g. `read_file` or godoctor's `smart_read`) are strongly correlated with scenarios that either failed or took longer to complete. Is this because the read operations are bad? No, this is because to recover from an error, the agent had to refresh its knowledge of the source code by reading it again. So, while there is a strong correlation between read and slowness and failure, by no means is this a relation of causality or, as statisticians like to say, "correlation doesn't mean causation".

Since I have been doing many experiments for the past few weeks, I quickly realised that teaching the model how to perform deeper analysis every single time was not very effective. Typically in these scenarios I either will add the analysis instructions to my agent context (via `GEMINI.md`), or will store the required prompts in an MCP server so I can map them to slash commands.

While both alternatives work, they both have their limitations. Expanding the agent context for each possible task I want to perform will result in context bloat and less effective behaviour. Creating slash commands for each prompt relies on me explicitly invoking the command as the agent is not aware of them by design.

Fortunately, **Agent Skills** provide a solution that combines the power of both. Agent Skills are a new feature in [Gemini CLI](https://geminicli.com) designed to give the agent on-demand capabilities. It behaves similarly to an agent tool (in fact, skills are activated by a tool call), but the skill allows on-demand access to a prompt and supporting files to enable the agent to do specialised tasks, putting them in context only when they are needed. 

You can find the full technical specs in the [official documentation](https://geminicli.com/docs/cli/skills/), but in this article I'll cover the basics to get you started.

## Anatomy of a skill

A skill is nothing more than a folder with a prompt and, optionally, supporting files like documentation and scripts.

```text
my-skill/
├── SKILL.md       (Required) Instructions and metadata
├── scripts/       (Optional) Executable scripts/tools
├── references/    (Optional) Static documentation and examples
└── assets/        (Optional) Templates and binary resources
```

The `SKILL.md` file is where the skill prompt lives. It has a small frontmatter to define the skill name and description, but other than that is just a regular markdown file:

```text
---
name: <unique-name>
description: <what the skill does and when Gemini should use it>
---

<your instructions for how the agent should behave / use the skill>
```

To add an skill to your project, you can create a folder under `.gemini/skills`. For example, the `my-skill` above would live in `.gemini/skills/my-skill`. Gemini CLI will automatically look for skills in the following order of precedence:

1. Workspace (<my-project-name>/.gemini/skills)
2. User (~/.gemini/skills)
3. Extensions (~/.gemini/extensions/<extension-name>/skills)

The important thing to note is that when Gemini CLI starts it is only aware of the skill name and description. Everything else will be loaded **on-demand** when the skill is activated.

Now let's have a look at how I'm using a skill to improve my experiment analysis workflow.

## The `experiment-analyst` skill

I designed the `experiment-analyst` skill to be activated when I ask Gemini CLI to evaluate an experiment. It is organised as follows:

```text
experiment-analyst/
├── SKILL.md                     <-- The analysis guidelines
├── references/
│   └── tenkai_db_schema.md      <-- The database schema, so the agent don't need to discover it every time
└── scripts/
    ├── analyze_experiment.py    <-- Replicates some of the analysis I have on the frontend
    ├── analyze_patterns.py      <-- Some deep dives into common patterns to extract insights
    ├── get_experiment_config.py <-- Retrieves the experiment configuration details
    └── success_determinants.py  <-- Tool call analysis and correlation
```

### Defining the expert persona

The `SKILL.md` file defines the analytical procedure. It tries to achieve a balance in teaching the agent what to do but not in a simple "cookie-cutter" way. One of the important aspects of it is trying to steer the agent away from jumping to conclusions, defining a more grounded persona. I still validate all claims and take all conclusions with a grain of salt, but this version has provided me some interesting insights that otherwise would take me a lot of work to discover.

```text
---
name: experiment-analyst
description: Expertise in analysing Tenkai agent experiments. Use when asked to "analyse experiment X" to determine success factors, failure modes, and behavioural patterns.
---

# Experiment Analyst

## Core Mandates
1. **Evidence-Based:** Never make claims without data. Cite specific Run IDs.
2. **Correlation ≠ Causation:** A tool might be correlated with failure (e.g., `read_file`) because it's used for recovery. Always investigate the *context* of usage.
3. **Comparative:** Always contrast the performance of alternatives.
```

Note: you can click here to see the full [SKILL.md](https://github.com/danicat/skills/blob/main/experiment-analyst/SKILL.md) file.

### The skill assets

You are going to hear me talking a lot about this over the next few weeks, but when dealing with agents, which are inherently **non-deterministic**, the only way we can ensure quality is to give them **deterministic** tooling. Skills fit nicely into this philosophy because we can bundle them with scripts to perform tasks in a consistent way instead of leaving up to the agent to "guess" how it is done.

For the experiment analysis skill I wanted the agent to have freedom to explore, but also I don't want it to re-invent the well all the time, so it comes with some pre-packaged scripts:

- `analyse_experiment.py`: reproduces an experiment summary similar to what I have in the frontend, but includes some grouping in tool calls for shell commands
- `analyse_patterns.py`: extracts samples from the agent conversation to try to identify tool usage patterns
- `get_experiment_config.py`: helps the agent understand the experiment by retrieving its definition
- `success_determinants.py`: calculates the correlation between successful outcomes and tool calls

I provide the database schema in `references/tenkai_db_schema.md` for when the agent decides to do ad-hoc queries, so it doesn't need to rediscover the schema every time (this schema is fairly stable between runs).

I will not claim this setup is perfect as I haven't spent a significant amount of time refining it, but this combination of information and pre-packaged scripts covers most of the questions I typically ask the agent to explore.

## Closing thoughts

Agent Skills represent a significant shift in how we design agentic workflows. By moving away from massive context prompts (e.g.: adding everything to GEMINI.md) towards modular, on-demand capabilities, we solve two problems at once: we keep our agent's context clean (less tokens), and we allow for deep, specialised expertise that doesn't dilute general performance.

In my case, the `experiment-analyst` skill was helpful to turn a repetitive task into a semi-automated flow. It gives me just enough consistency and flexibility to perform the analyses I want. I am now considering upgrading other parts of my workflow to skills, moving away from my approach of using MCP servers as "prompt databases".

I’m excited to see what the community builds. So, take a look at your own workflows. Where are you constantly repeating instructions? Where do you need a specialist? That's your next skill waiting to be written.

Happy coding!