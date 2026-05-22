---
title: "The Hitchhiker's Guide to Antigravity 2.0"
date: 2026-05-21T11:00:00Z
categories: ["AI & Development", "Workflow & Best Practices"]
tags: ["antigravity", "agy-cli", "agy-sdk", "google-io", "agentic-coding"]
summary: "A guide to the Google Antigravity 2.0 ecosystem announced at Google I/O 2026. We examine the standalone desktop application, the Go-based terminal CLI, and the programmatic Python SDK."
heroStyle: "big"
---

Since [Google I/O 2026](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/) just wrapped up it is now the time to decompress all the new releases and how they will affect our workflows now and in the near future. While lots of interesting things were announced, today I want to focus on what impacts developers the most which is the release of [Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0) and the expanded Antigravity (agy) ecosystem (see the [Google I/O 2026 Antigravity highlights](https://antigravity.google/blog/google-io-2026)) which includes [Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli) and [Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk).

Before getting into the technical details, I feel the need to address that there is a lot of noise on the web due to this launch and, unfortunately, not of the good kind. The main reason for this is that Antigravity 2.0 introduces breaking changes in many aspects of the development flow, starting with the separation of the IDE environment from the main Antigravity desktop app.

Secondly, the [announcement of the deprecation of Gemini CLI](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) in favour of Antigravity CLI was also not well received due to the short deadline given for users to migrate from one to the other (plus a few quirks we are going to see below). In essence, users are given until June 18, 2026 to migrate, basically a month from I/O, which frankly is not a lot.

I wrote about this before and I understand how it is frustrating when your favourite product is deprecated. I, for example, still grieve Google's Inbox which is a long forgotten email client that lost the battle against Gmail. I'm not here to sugarcoat anything, Google indeed has the reputation of killing good products. But, personal preferences aside, when you look at the big picture, I actually admire Google for being bold enough to kill products the way it does.

I believe most people expect from Google to lead the disruption in all things technology related, and specially today, in such a dynamic environment due to the advancements in AI, it takes a lot of courage and determination to pivot from one direction to another. I talk a lot about Agile, and while Google is not typically associated with formal Agile procedures, this is a characteristic that all experienced agilists will recognise as one of the most valuable in a company: the capability to course correct fast, pivot, experiment, learn from mistakes and iterate.

Instead of staying in the comfort zone, this is what keeps Google always at the edge: its capability to reinvent itself, even if not every experiment succeeds. Actually, it is expected that many experiments will fail. This is how we learn about what works and what doesn't. We take the lessons and move on to the next goal, incorporating those in the newer products.

There will be many lessons to be learned from this release, but ultimately, if you look at the tech itself, hopefully it will become clear what is the endgame: we are doubling down in the agentic-era, while consolidating efforts to build more advanced products.

## The new Antigravity desktop app explained

The biggest change in the desktop app is the removal of the IDE component. In Antigravity 1.x, the app was based on a fork of VS Code, so you had all those familiar IDE features to navigate and edit code paired with an assistant box that you could use to interact with the agent.

Not only that, you also had a secondary UI called the "Agent Manager" where you could have a 10,000-foot view of different chat sessions (aka "conversations"), meaning you could work on many projects in parallel by monitoring the agents in this view and reacting when they were waiting for input.

The biggest change in the new desktop app is that Antigravity 2.0 puts the agent manager experience front and center, removing the IDE part completely (the IDE part became a separate and optional app). 

![The new Agent Manager interface](image.png "The new agent manager interface is cleaner, focused on projects and conversations")

For seasoned developers, this became a huge friction point because suddenly all the familiar editor tools they had relied on for so long were simply gone. You can still see files in the agy 2.0 UI, but only the ones agy is currently working on and you cannot edit them directly. Every interaction is made through a prompt or an annotation on the file.

![File view in agy 2.0](image-2.png "You can still see files on the UI, but not edit them directly")

The interactions with the agent should be already familiar to anyone coding with agents in the past year. Once you give it a prompt, it will elaborate an implementation plan, which you can review using in-line comments or a top level prompt, and once approved the agent goes on its own to execute. Depending on how you configure the UI, every now and then it might interrupt you asking for a permission, which you can either give or reject with an optional course correction.

![Agent Manager asking for user input](image-1.png "When you reject a request you can add an steering comment")

In terms of extensibility, agy 2.0 supports common standards we have grown used to in this past year including MCP and Skills, but also its own "Rules" mechanism which was inherited from 1.x (essentially a composable AGENTS.md) and a new plugin system based on the extension system of the former Gemini CLI. Plugins allow you to pack together additional rules, slash commands, MCP servers, skills and sub-agents, and they are backward compatible with Gemini CLI extensions (meaning you can install CLI extensions on agy, but not the other way around).

Overall, while I understand the frustration of people who miss the IDE, my initial impression is that I don't miss it being in the **same** app. Even when working with Gemini CLI, I always had VS Code running in parallel for when I want to do manual edits, and this is the same flow I'm applying to agy 2.0. In fact, I use VS Code mostly as a text editor and hardly ever use any proper IDE features nowadays. I could replace it with Notepad and it wouldn't make much of a difference, except for the lost muscle memory for a few shortcuts, which is the only reason why I keep using VS Code these days.

While I must say there is nothing groundbreaking in agy 2.0 in comparison with agy 1.x or even other coding agents, I am quite enjoying the cleaner look, and I believe I will only extract its full power when I start customising it with my own plugins. I'm currently working on upgrading godoctor and speedgrapher from their Gemini CLI extension forms to agy plugins and will report back as soon as I have something to show.

## Antigravity CLI 

For terminal users, the command-line experience is rebuilt under the new [**Antigravity CLI**](https://antigravity.google/blog/introducing-google-antigravity-cli) (aka. `agy CLI`). It can be a bit confusing at first, but you need to install the agy 2.0 app even if you only plan to use the CLI, as they share the same authentication process. agy CLI is the natural replacement for Gemini CLI and while it doesn't provide a 100% feature match, the main suspects are already there, namely hooks, skills, MCP, sub-agents and plugins.

The whole CLI was rewritten in Go (Gemini CLI was in TypeScript), which makes me profoundly happy as we can expect a snappier experience. At the same time, one of the main criticisms is that agy CLI, as of today, is closed source, which might feel as a regression from Gemini CLI. Not long ago we were making jokes about "leaking" Gemini CLI code to the public. Unfortunately this joke didn't age well as now our main coding agent is also closed source.

Considering I have zero control over this, it is also something I decided not to bother about. It is too early to say whether this is a good or bad decision, but again, I recognise the frustration of the community especially if you contributed to Gemini CLI before. If it serves as a consolation, we will still have a thriving open source community around the plugin system. At least, I am working on mine to make sure we have both a decent Go expert sub-agent and a vibe-writing companion very soon.

![agy CLI interface](image-3.png "The UI will feel familiar for people coming from Gemini CLI or Claude Code")

In terms of the UI, it shouldn't surprise anyone who has used a CLI coding agent before. My first impression is that the render indeed feels better than the TS render in Gemini CLI. Also, similarly to agy 2.0, I really appreciate the cleaner look. In my personal opinion Gemini CLI was getting a bit too big for its own good, too many features and a bloated UI, so this cleaner interface feels refreshing to me. One of my all-time favourite sayings is "less is more", and agy CLI delivers in this aspect.

Where it doesn't deliver very well (for now), is mostly related to compatibility with extensions. Although there is a migration path, it doesn't always work as expected, and this is why I have been dedicating most of my week to rewriting godoctor and speedgrapher, as I prefer to not rely on the automatic migration. Second to that, I also had issues with project based authentication, which I hope it is fixed soon. For now, I have been using it with my Google Pro subscription as project based authentication didn't work for me.

Without getting into the complexities of billing, which is another pain point for users coming from Gemini CLI, my personal opinion is that agy CLI has its issues, but also has great potential. So far there is nothing revolutionary about it (as most of the changes are happening under the hood), but also I cannot see any dealbreakers. Everything that I did on Gemini CLI can be done with agy CLI and there is very little to be learned so, even if Gemini CLI had a longer migration window, I would recommend migrating to it as soon as possible to ensure your workflow is future-proof.

## Antigravity SDK

The discussion so far has been about replacing old products with new ones, which feels more incremental than revolutionary. This is why the release of [**Antigravity SDK**](https://antigravity.google/blog/introducing-google-antigravity-sdk) was the most exciting announcement for me. When I spoke about the majority of the changes happening under the hood, it was about creating this unified platform to support agents, and Antigravity SDK is how you as a developer can also have access to it.

Here is a functional example of an agent querying its workspace in under 15 lines of code:

```python
import asyncio
from google.antigravity import Agent, LocalAgentConfig

async def main():
    config = LocalAgentConfig()
    async with Agent(config) as agent:
        response = await agent.chat("What files are in the current directory?")
        print(await response.text())

if __name__ == "__main__":
    asyncio.run(main())
```

This Python library gives developers programmatic access to the same agentic runtime and orchestration harness. The SDK is runtime-agnostic and allows spinning up a stateful agent loop in under 15 lines of code. It supports modular capabilities including built-in tools, custom functions, Model Context Protocol servers, sub-agents, and reusable skills under a unified pipeline.

## Getting started

A common trend among all releases relating to Antigravity is this shift from code-first to design-first. The whole software development experience is being redesigned about coordinating agents instead of editing code. To prepare your development environment for this shift, consider the following actions:

1.  **Download the desktop app**: Visit [antigravity.google](https://antigravity.google) to install the desktop application.
2.  **Migrate terminal workflows**: Install the `agy` CLI and run the import command to migrate your Gemini CLI configurations before the **June 18, 2026** deprecation date (refer to the [migration announcement](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/) for details).
3.  **Explore the SDK**: Install the Python library, check out the [Antigravity features](https://antigravity.google/docs/features), and begin building custom agents powered by agy SDK:
    ```bash
    pip install google-antigravity
    ```

## Additional resources
To learn more about this release and access additional technical details, check out the following resources:
* **[Introducing Google Antigravity 2.0](https://antigravity.google/blog/introducing-google-antigravity-2-0)**: The official announcement of the 2.0 ecosystem.
* **[Introducing Google Antigravity CLI](https://antigravity.google/blog/introducing-google-antigravity-cli)**: A deep dive into the new Go-based terminal interface.
* **[An Important Update: Transitioning Gemini CLI to Antigravity CLI](https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/)**: Detailed migration timeline and guidelines for Gemini CLI users.
* **[Introducing Google Antigravity SDK](https://antigravity.google/blog/introducing-google-antigravity-sdk)**: Learn how to programmatically orchestrate agents in Python.
* **[Google I/O 2026 Developer Highlights](https://blog.google/innovation-and-ai/technology/developers-tools/google-io-2026-developer-highlights/)**: The major developer announcements from this year's Google I/O.
* **[Google I/O 2026: Antigravity Announcement](https://antigravity.google/blog/google-io-2026)**: Key updates and highlights for Antigravity from Google I/O.
* **[Google Antigravity Documentation & Features](https://antigravity.google/docs/features)**: The comprehensive guide to Antigravity features and safety controls.
