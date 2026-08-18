---categories:
- Agentic Coding
date: 2026-08-17
heroStyle: big
series:
- Gemini for Go Developers
series_order: 2
summary: In Part 2 of Gemini for Go Developers, we explore Go's agentic affinity,
  Antigravity surfaces, and how to configure a practical AI-native workflow in Go.
tags:
  - antigravity
  - gemini
  - golang
  - mcp
title: "Gemini for Go Developers - Part 2: Coding with Gemini"
slug: "gemini-for-go-developers-part-2-coding-with-gemini"
aliases:
  - "/posts/20260817-gemini-for-go-developers-part-2-coding-with-gemini/"
description: "Part 2 of Gemini for Go: maximize Go's agentic affinity, explore Antigravity surfaces, configure MCP servers, and build an AI-native Go toolkit."
proficiencyLevel: "Intermediate"
dependencies:
  - "Go 1.24+"
  - "Antigravity CLI"
  - "godoctor"
---

Welcome back to **Gemini for Go Developers**! In [Part 1: The Gemini Model Family]({{< ref "/posts/20260808-gemini-for-go-developers-part-1-model-family" >}}), we explored the different Gemini models for specific use cases, looked at API surfaces to consume models, and wrote our first Go code with the official [Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai).

Now, in Part 2, we are going to explore how to use Gemini for coding in Go. We will start with a brief discussion on language choices in the Age of AI, then explore the landscape of agent harnesses and agent standards, finishing with the recommended setup to increase the agentic affinity of Go in your development environment.

## Why use Go in the age of AI?

Before we get any further, let's get the obvious question out of the way: in the age of AI, does our choice of programming language even matter?

In the past, language selection was almost always dictated by existing team expertise. Learning new syntax, idioms, and tooling quirks took valuable time, so teams stayed in their comfort zones unless forced to move by a major tech shift.

AI flipped that dynamic. Syntax is no longer a barrier when models generate boilerplate on demand, and learning an unfamiliar stack with an AI tutor beats spending hours praying that some five-year-old Stack Overflow thread matches your exact compiler error.

Why would you care about language selection then? It boils down to two main themes: **language ecosystem** and **agentic affinity**.

In this context, a language ecosystem refers to everything the language attracts by "gravity": actively maintained SDKs, libraries, documentation, community vitality, and industry knowledge. But **agentic affinity** is unique to the world we are living in today. I define it as "how easy it is to guide an agent to code with this language". High agentic affinity is primarily influenced by how prepared the models are to generate code in this language, and secondly by how prepared they are to adopt the tooling required to verify, test, and maintain that code.

Agentic affinity naturally depends on how much data about a given language was available during model training, benefiting popular languages with extensive public code and literature. That said, volume alone isn't everything: older languages that underwent massive paradigm shifts suffer from fragmented practices, causing models to suggest obsolete patterns at inference time.

In my personal experience, languages like Go, Python, and JavaScript have high agentic affinity by default. Go stands out in particular: its readability, strict static typing, and almost Pythonesque philosophy that "explicit is better than implicit" make code generation far less prone to hallucinated syntax. More importantly, Go's fast compilation, built-in testing, and opinionated standard tooling give coding agents an immediate, deterministic feedback loop to catch and repair errors autonomously. (For a deeper dive on how Go's design philosophy aligns with AI-assisted software engineering, check out [this essay by Cameron Balahan and Richard Seroter on the Google Developers Blog](https://developers.googleblog.com/why-go-is-an-ideal-language-for-ai-assisted-software-engineering/)).

On the other hand, languages like R and C sit on the other end of the spectrum. In R's case, it is a niche academic language where even with full agent support, I am not able to keep my [read.dbc](https://github.com/danicat/read.dbc) package on CRAN (R's package distribution system) because neither I nor the models can reproduce the problem reported by CRAN's pipeline. In C, the lack of guardrails means simple mistakes can become silent, catastrophic bugs that models struggle to catch early without rich tooling ergonomics.

Ultimately, agentic affinity is measured by the speed and responsiveness of the agentic "loop": whether your agent can build, test, benchmark, and repair code out of the box without human micromanagement.

## Selecting the right agentic surface

While the frontier model race is exciting to watch, models alone are only one part of the equation when it comes to code generation. The response quality and the developer experience of coding with an agent are heavily influenced by the harness that is running the model.

If you are coding with Gemini, the natural choice is the **Antigravity ecosystem**. With the release of Antigravity 2.0, Google separated the ecosystem into three distinct surfaces depending on how you like to work:

### Antigravity 2.0

At the centre of the Antigravity ecosystem is the desktop application **Antigravity 2.0**, sometimes also referred to as the **Agent Manager**. The biggest change here is putting the agent experience front and centre while the code sits in the background. For first-time users accustomed to an IDE, the experience might be a bit shocking: there are no file trees to explore the codebase, nor any way to edit files manually. Every interaction is made through the agent. Your control lies in telling the agent what to do instead and annotating their work with "Google Docs"-style comments.

### The Antigravity CLI (`agy`)

While the Agent Manager concept is fairly new, terminal UIs for agents have been around for a bit longer, popularised by Claude, Aider, Cline, and others. In June 2025, Google also launched its own terminal UI, the Gemini CLI, but it has since been deprecated in favour of the new **Antigravity CLI** (`agy`).

One detail that makes me particularly happy as a Gopher is that `agy` is written in Go (whereas the Gemini CLI was in TypeScript), resulting in a noticeably snappier terminal experience.

### The Antigravity IDE

If you still prefer a dedicated, visual code editor, Google offers the **Antigravity IDE** as a separate companion application. It is based on VS Code, so all the familiar IDE elements are where you remember them to be, plus it has an agent side panel for interactions with Gemini.

To be 100% transparent, nowadays I rarely open the IDE for coding purposes. The only times I really use the IDE are when I am writing or reviewing articles (like this one) as I still do a huge part of my writing process manually. For code, I very rarely edit anything by hand these days.

## Agentic standards applied to coding

Regardless of the surface you choose, Antigravity is very capable out of the box, but it is not without its quirks. The best way to fully leverage Antigravity's capabilities is to equip your agent with **customisations**. 

Antigravity supports both well-established and emerging agentic standards: rules, skills, MCP, hooks, subagents, sidecars, and plugins. However, support for these customisations is, unfortunately, **not uniform** across the different Antigravity surfaces today.

Let's look at how each of these customisations works in practice.

### Agent instructions and rules

The concept of instructions was standardised by the [**AGENTS.md**](https://agents.md/) initiative, which Antigravity supports via either `AGENTS.md` or `GEMINI.md` files (as well as modular rules under `.gemini/rules/` or `.agents/rules/`).

You can think of `AGENTS.md` as the `README.md` for AI agents. It is the place to store project context, architectural constraints, testing commands, and style preferences that are essential for an agent to know, but would otherwise clutter human documentation.

To be completely honest, with the advent of Agent Skills, I rarely update `GEMINI.md` any more, and most of my repositories probably have outdated files (my bad, poor agents!). That said, they are still useful for steering agents in the right direction. Just keep in mind that prompt-based instructions act more as recommendations than guardrails: an agent can still selectively overlook or drift away from rules during a long session.

### Model Context Protocol (MCP)

The [**Model Context Protocol**](https://modelcontextprotocol.io/) (MCP) is an open standard for connecting AI applications to external tools and data sources. 

The protocol exposes three core primitives:
1. **Tools:** Executable functions the agent can invoke (e.g. querying a database, running a linter, checking a build).
2. **Resources:** Read-only data sources the agent can inspect (e.g. documentation files, database schemas, system logs).
3. **Prompts:** Predefined workflow templates.

In the real world, we care about tools more than anything else, and most clients won't even support resources or prompts. Resources can be emulated via tools (a dedicated tool can retrieve data) and prompts have mostly fallen into obscurity due to the introduction of Agent Skills, which are much more flexible.

Because skills can be packaged with scripts, some people are even ditching MCPs completely in favour of skills. I still think there are plenty of cases where MCPs are better than skills. One of the strengths of MCPs over skills is lifecycle management, especially when using them over HTTPS. For example, as a company you can deploy an MCP server as a web service and your clients only need to configure it once to have immediate access to up-to-date documentation. With skills, on the other hand, ensuring all your clients only use the latest versions is a problem still to be solved at scale.

### Agent skills

A [skill](https://agentskills.io) is a directory containing instructions (`SKILL.md`), optional helper scripts, and documentation that teach an agent how to execute a specific engineering workflow.

The defining architectural concept behind skills is **progressive disclosure**. Instead of dumping hundreds of pages of documentation into the agent's context window up front, the system only injects the skill's name and description. When the agent determines that a task matches a skill, it loads the full `SKILL.md` instructions and executes the bundled scripts on demand. This model ensures specialist knowledge is always available, but doesn't overwhelm the context window or distract your agent with data that is not relevant to the task at hand.

Agents use skill descriptions to identify their activation triggers, but relying solely on automatic activation is risky. By default, Antigravity maps all skills as slash commands, so you can force a skill activation by typing `/<skill-name>` anywhere within your prompt. Being proactive with skill activation will save you a lot of headaches if the skill is important for your workflow.

### Hooks

While rules, prompts, and even skills offer soft guidance, **hooks** introduce deterministic control into the agent loop. Hooks are callbacks that intercept the agent lifecycle at specific moments, such as before a tool runs (`PreToolUse`), after a tool executes (`PostToolUse`), before a model invocation (`PreInvocation`), or upon session termination (`Stop`).

Because LLMs are non-deterministic, telling an agent to "always run a linter after editing code" via prompts leaves validation up to chance. Hooks, on the other hand, are controlled by the harness and always run for their given event.

There is only one caveat that you need to always pay attention to when designing your hooks: models are very good at bypassing them. Yes, unfortunately, this is a thing. A hook might block the agent from doing harm, only for the agent to try to outsmart the hook immediately afterwards, either by tampering with the agent configuration, trying to mask the trigger condition, or, worse, rewriting the hook script itself.

It saddens me to say this: in the past, I used hooks a lot, but with the new generation of models they got too smart for their own good, so I'm slowly moving away from hooks towards skills. Feels like teaching a child: do not prohibit, educate.

### Subagents

**Subagents** offer another solution to the context window problem, while also enabling some interesting paradigms, like parallel execution. By spawning subagents, the main agent can segment the problem space and create an agent focused on each task.

A trivial example would be to work on a web service that has both a frontend and a backend. The changes are essentially orthogonal to each other: the frontend requires HTML, CSS, and JavaScript, while the backend requires Go, Python, and maybe some SQL. Both frontend and backend tasks will have different coding standards and build pipelines. With the exception of their contract, they have nothing in common and if done in the same context window, one part will only be noise for the other.

By splitting the tasks between two (or more) subagents, you ensure each subagent has full focus on its slice of the stack, reducing the risk of context degradation due to crossing the beams of two completely different tech stacks.

Another good example is when you need an unbiased review of the work you just finished. Ask the agent to run the code review in a subagent, and you have the benefit of "fresh eyes" looking at the code.

### Sidecars

> Note: at the time of this writing, sidecars only work on Antigravity 2.0. They are not available on the `agy` CLI or the IDE.

**Sidecars** are background processes that run alongside the agent for the duration of a session. They can be persistent processes with the lifecycle managed by Antigravity through a defined restart policy, or they can be processes that run on a schedule.

I haven't explored sidecars yet, so there isn't much I can add to the discussion at the moment, but my colleague Mete Atamel just started writing about them, and I encourage you to check [his article](https://medium.com/google-cloud/where-does-antigravity-look-for-sidecars-20e7002b9246) for more information.

### Agent plugins

Plugins are, in essence, customisation bundles, packaging rules, MCPs, skills, hooks, agents, and sidecars in a single distribution unit. I've been experimenting with plugins for a while, and unfortunately, they don't work as smoothly as they are supposed to. For instance, I had a lot of trouble trying to ensure the hooks packaged with my plugins were executed properly, but it never worked for me.

Also, plugin support is non-standard across different surfaces: for example, sidecars are only supported in Antigravity 2.0 plugins, custom agents are only supported in the Antigravity CLI, and so on.

What are they good for then? MCPs and skills. Whether it's a coincidence or not, this is also where the industry is converging with the new [Agent Plugins](https://agent-plugins.org/specification) standard. Their justification is that hooks, agents, rules, LSP servers, and others are still pretty much client-specific, so out of scope for now.

## Equipping your Go agentic toolkit

Now that we have covered agent harnesses and customisations, let's assemble our Go development toolkit. We can divide these into essential community tooling and specialised AI-native extensions.

### Essential community tooling

While Go's standard toolchain gives agents a strong baseline, several community tools elevate code quality and release automation:

- [**`golangci-lint`**](https://golangci-lint.run/): Bundles dozens of fast linters into a single pass, catching unhandled errors, unchecked type assertions, shadowed variables, and concurrency pitfalls that `go vet` misses.
- [**`goreleaser`**](https://goreleaser.com/): For projects that distribute binaries, `goreleaser` automates building cross-platform artifacts, managing release pipelines, and generating changelogs from `.goreleaser.yaml`.
- [**`modernize`**](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize) / **`go fix`**: Analyses code against newer Go releases and mechanically upgrades older boilerplate, like replacing manual slice/map loops or min/max helpers with modern built-in functions.
- [**`deadcode`**](https://pkg.go.dev/golang.org/x/tools/cmd/deadcode): Uses whole-program reachability analysis to identify unused functions and unreachable code across packages.
- **`selene` and `testquery` (shameless plug):** I maintain two open-source tools to inspect and improve test suites. [**`selene`**](https://github.com/danicat/selene) is a mutation testing tool for Go that introduces targeted AST faults to verify whether tests actually catch code defects. [**`testquery`**](https://github.com/danicat/testquery) is a CLI that exposes a SQL interface to query Go test results and per-test coverage. While a bit niche, they help me optimise my test suites.

### AI-specific tooling

The tools above were built for human developers, but coding agents can execute them directly through standard shell commands. Now let's have a look at specialised MCPs and skills for AI-native workflows.

#### The official `gopls` MCP server

In a traditional IDE, `gopls` provides your editor with semantic awareness: type checking, references, and symbol definitions. But when an agent operates in a headless or terminal environment, it typically interacts with code as raw text files.

To bridge this gap, the Go team added native MCP support to [**`gopls`**](https://pkg.go.dev/golang.org/x/tools/gopls). Running `gopls` in MCP mode exposes the language server's type-checker and index directly to the model as executable tools. This allows the agent to navigate package hierarchies and inspect type signatures using the compiler's own semantic model.

#### The `godoctor` MCP server

One of my rants about the `gopls` MCP server is that it was designed on top of the LSP API, which was not designed with agents in mind: LSPs were built for typing speed and interactive keystrokes, while agent workflows are transactional. For this reason, I built [**`godoctor`**](https://github.com/danicat/godoctor) to provide AI-native tools for Go development.

The current version of `godoctor` provides the following:
- **`smart_edit`:** An AST-aware file editor with automatic `go vet` validation and typo correction. If an edit introduces a compile or syntax error, `godoctor` automatically rolls back the change and suggests fixes based on nearby identifiers with a "did you mean?" kind of steering prompt.
- **`smart_build`:** An automated verification pipeline that runs module hygiene (`go mod tidy`, `modernize`, `goimports`), builds the package, executes tests with coverage, and validates linting via `golangci-lint` in a single pass.
- **`smart_test`:** An opinionated test pipeline with support for `testquery` and `selene`.
- **`read_docs`:** Documentation lookup based on `go doc`, which supports examples and includes a fallback system that shows documentation independently of module configuration.

#### Platform knowledge and self-improvement


While `godoctor` and `gopls` handle local code semantics, coding agents also need live platform knowledge and workflow playbooks to understand the services they are integrating with. Here are the essential cloud and API resources when working with Go, GCP, and Gemini:

- **Google Developer Knowledge MCP (`developerknowledge.googleapis.com/mcp`):** Connects the agent directly to official Google Cloud, Gemini Enterprise (Vertex AI), and Google API documentation.
- **Gemini Docs MCP (`gemini-api-docs-mcp.dev`):** Provides live documentation for current Gemini API endpoints, SDK updates, and configuration patterns (read more about it in the [Gemini coding agents guide](https://ai.google.dev/gemini-api/docs/coding-agents)).
- **Official Google Skills:** [**`github.com/google/skills`**](https://github.com/google/skills) and [**`github.com/google-gemini/gemini-skills`**](https://github.com/google-gemini/gemini-skills) contain official skills maintained by Google (including `gemini-api-dev`, `gemini-live-api-dev`, and `gemini-interactions-api`).
- **Community & personal catalogue:** You can browse my personal skills at [**`skills.danicat.dev`**](https://skills.danicat.dev) (or on [**GitHub**](https://github.com/danicat/skills)), which includes skills for engineering best practices, 2D game development, generative media (Lyria, Nano Banana Pro), and others.

And for optimising your own workflow with custom extensions:

- **AgentSkills MCP (`agentskills.io/mcp`):** The official search and retrieval engine for querying the [Agent Skills open specification](https://agentskills.io) and authoring best practices. Great for when you create your own skills as part of your daily job. Which you should.
- **MCP Dev skills:** If there is an MCP for agent skills development, why not also have some [agent skills for MCP development](https://modelcontextprotocol.io/docs/2026-07-28/develop/build-with-agent-skills)? Yes, you read that right (LOL). While a bit niche, as you can see from my own work, creating MCPs for your own use is also a good way to improve your development environment.

## The Gopher's 5-minute agentic setup

If you want an opinionated setup to start working with Gemini today, here is the 5-minute quickstart:

1. Download [Antigravity](https://antigravity.google) from the official website.
2. Configure the recommended MCP servers:
   - [Gemini Docs MCP](https://ai.google.dev/gemini-api/docs/coding-agents): `npx add-mcp "https://gemini-api-docs-mcp.dev"`
   - [Developer Knowledge MCP](https://developers.google.com/knowledge/mcp): enable the API and configure it following the instructions on the documentation page.
   - [Agent Skills MCP](https://agentskills.io): Expand the copy button on any page for instructions.
   - [godoctor](https://github.com/danicat/godoctor): Use the one-line install script.
3. Add the recommended skills:
   - [Gemini API development](https://ai.google.dev/gemini-api/docs/coding-agents): `npx skills add google-gemini/gemini-skills --skill gemini-api-dev`
   - [Swarm coding]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}): `npx skills add github.com/danicat/skills/agents/swarm-coding`
4. Test drive the loop:
   Prompt your agent to run an autonomous verification pass:
   > Run a smart build on this package with godoctor, address any findings, and evaluate the test suite with selene.

## What's next?

In this chapter, we covered the landscape of agent harnesses, customisation standards (rules, MCP, skills, hooks, subagents, and plugins), and how to configure a practical environment for coding in Go with Gemini.

In **Part 3: Developing Agents in Go**, we will cross over to the other side of the table: building autonomous agent runtimes in Go. We will explore tool-calling loops, context engineering, and higher-level agent frameworks like **Genkit Go** and the **Agent Development Kit (ADK)**. See you there!
