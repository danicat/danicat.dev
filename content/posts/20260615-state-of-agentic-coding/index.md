---
title: "State of Agentic Coding"
date: 2026-06-15T00:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["antigravity", "agy-cli", "agentic-coding"]
summary: "Explore the evolution of agentic coding. Update on the shift to planning systems, strategic agent skills, and orchestrating subagents for better outcomes."
heroStyle: "big"
---

It has been more than six months since I published the article [Taming Vibe Coding]({{< ref "/posts/20251206-taming-vibe-coding" >}}) which consolidated all the main practices I had been using to increase my productivity with coding agents.

While most of the content in that article is still relevant today, in the past six months a lot has happened in this space, so I decided to post a quick update on how my thinking evolved since then.

## Prompting & context engineering

Prompting is still important, and a good structured prompt will save you a lot of time, but it is not the deal breaker it used to be due to the emergence of planning systems. Most coding agents ship today with a plan (or planning) mode where they spend a few turns "brainstorming" the task to then elaborate an implementation plan before jumping to coding.

This gives you the opportunity to review the plan and steer the agent before code is written, saving a lot of time and tokens. What didn't change is the need for a strong acceptance criteria to ensure consistent results. This is the part that I'm usually paying the most attention when I review a plan.

Context engineering has massively improved with the widespread adoption of agent skills, which makes me hardly care about writing an AGENTS.md (or GEMINI.md, CLAUDE.md, etc.) nowadays. In Antigravity we have the concept of Rules which is essentially context optimisation, but I hardly use them in favour of skills. I have yet to find something that I need to do in AGENTS.md or rules that I can't do using another technique.

The aspects of context engineering related to Retrieval Augmented Generation (RAG), either via semantic search or any other means, are still relevant for specialised knowledge. This is how memory systems work and I still use the technique to inject package documentation in my context whenever working with an external dependency.

## The rise and fall (?) of Model Context Protocol

Instead of spinning up MCP servers, more and more people are preferring to use agent skills combined with CLI tools. While I respect that, I do have some trust issues giving my agents direct shell access and I prefer to package my tools in MCP servers that I run locally (like [godoctor](https://github.com/danicat/godoctor) or [speedgrapher](https://github.com/danicat/speedgrapher)). The only thing that changed for me is that I am much more selective in installing MCP servers, and most of the time I have only my custom made ones configured.

I also almost never use an MCP server alone, but I use a combination of skill plus MCP. The skill describes the process, the MCP exposes the tooling. This is often better than trying to optimise the MCP instructions themselves, which I spent hundreds of hours doing, only for coding agents to ignore them completely.

While I use MCP + skills for most of the time, for really serious stuff I use the "super combo" MCP + skills + hooks. The hooks part is responsible for forcing the agent in the direction I want to go. Sometimes I call this "putting the agent on rails" or "reducing the agency of the agent." The hook system allows me to block the undesirable actions and give the agent a "gentle nudge" to use the tool I want it to use, removing the element of probability of having a tool called or not. Or in other words, it forces a deterministic behaviour of the agent.

## Skills for everything

I create skills for processes that I want to be repeatable. For example, some of my most used skills are related to tech writing and reviewing because I am always producing content (like this blog). I also write skills about technologies that I know the agent is going to have trouble dealing with. Typically those include newer technologies, niche projects or custom made ones.

For example, just recently I had a very hard time preparing my A2UI workshop. A2UI is a new-ish protocol to develop agentic user interfaces. Because it is such a novel concept, and I had very specific teaching needs, the agents couldn't figure it out without a lot of help and trial and error. Once the initial hurdles were overcome, packaging this knowledge in a skill helps smoothing things over the next time I need to do something similar.

On the negative side, I am terrible at keeping my skills up to date and organised. I think solving this problem can be the redemption moment for MCP as there is a current proposal to add skills to the protocol specification. Unfortunately, there is no estimation of when this will happen, if ever, so for now we are stuck managing skills on our own. In theory you could create a "skill-like" experience with MCP using prompts for the progressive disclosure part and tools for whatever needs scripting, but given my current backlog I haven't tried doing this yet.

## The rise (and no fall) of subagents

Subagents are the "next cool thing" that everyone in the village is talking about. The idea is to parallelise tasks by launching them as their own agents with a segregated context window. This has the benefit of optimal usage of context, avoiding contamination between tasks and early context rot. This also reduces the need for compression as each task is self contained and won't pollute your main context window. In terms of coding harness support, some harnesses support pre-declaring agents just like you can declare a skill, each with their own system prompt, model and configurations, while others like Antigravity favour adhoc creation of agents with each agent being a "clone" of the main session, but with their own context windows.

While ad-hoc creation of agents allow you to do crazy things like using a prompt to spin up 3 different agents to run things in parallel, I do miss having pre-declared agents in Antigravity as they allow me to pack my "curated" agents in portable ways. Also, parallelising tasks between agents is not a trivial skill to master - in the last year I got used to give tasks to agents, but I haven't really thought much about how to efficiently orchestrate them.

In many ways this is the same muscle as a product owner or team lead exercises when breaking down tasks and thinking about how the team will tackle them, but with sub-agents instead of having a fixed team you can create as many "team members" as you want. Ultimately, I care less about parallelising agents because its "cool", and care more about the question "is this going to produce better outcomes?"

If the answer to the question is **no**, then you are better off running one agent at a time, as the effort spent in planning and the mental toil of carefully dividing the tasks will not pay off. This is why I prefer being able to pre-define my agents, as I am just specialising them for when they are needed, but it doesn't really matter to me if they run in parallel or not.

## Hooks are your best friends

I saved my favourite one for last: hooks. Hooks are callbacks that will be triggered upon specific events on the agent lifecycle. I wrote an [entire article about hooks]({{< ref "/posts/20260610-mastering-hooks" >}}) last week, which I encourage you to check right after you finish this one. The bottom line is, models are unpredictable and can go off the rails quite easily. Hooks are a good way to add guardrails to the models so you can steer them in the direction you want without relying on chance. More than that, they allow you to plug monitors to the agentic harness to collect data and improve the response quality, like for example adding a persistent memory system.

## Conclusions

The industry is moving fast and to stay relevant we need to be adaptive to embrace new processes and techniques as they come and release the baggage that is holding us back. Still, do not take any of guides out there as the source of truth (including this one). It is a learning experience for everyone as this technology is still in its early stages. The key is to experiment and see what kind of technology and workflow works best for your environment.

In this article I shared what is working for me, and how my thinking has been evolving, but I do not know everything and I am always learning. We talk a lot about training AI, but don't forget that training your brain is much more important than that. Do not use AI as an excuse to disable your brain, keep experimenting, learning and iterating. And please, share any interesting stuff that you find out!
