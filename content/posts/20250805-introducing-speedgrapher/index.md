---
title: "Introducing Speedgrapher: An MCP Server for Vibe Writing"
date: 2025-08-05
author: "Daniela Petruzalek"
categories: ["Workflow & Best Practices"]
tags: ["golang", "gemini-cli", "mcp", "ai", "vibe-coding"]
summary: "The story of Speedgrapher, a custom MCP server for 'vibe writing.' It details the journey of turning a personal collection of prompts into a portable, AI-powered toolkit to automate and structure the creative process."

---

## Introduction

I have a confession to make: I love building things, but I don’t always love the boilerplate that comes with it. I often have a lot of ideas for new articles, but the process of structuring them, ensuring they meet my own editorial standards, and even just getting the tone right can sometimes feel like a chore. This is the story of how a deep dive into a technical specification led me to build [Speedgrapher](https://github.com/danicat/speedgrapher), an MCP server that helps me bring a helpful layer of structure to my writing process.

The journey to Speedgrapher began right after I published my last article, "[How to Build an MCP Server with Gemini CLI and Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}})." In that post, I focused entirely on how the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) allows AI agents to use tools. After publishing it, I went back to the MCP specification for another read. This time, a small detail I had previously overlooked caught my eye: besides `tools`, the protocol also explicitly defines `prompts` and `resources`. A lightbulb went off. I realized that the collection of prompts I had scattered across my notes, files and GitHub repos, could be packaged, and made portable using the very same protocol.

In a fortunate coincidence, the same day I was exploring the idea of a prompt server, the Gemini CLI team announced a new feature that makes prompts exposed by MCP servers available as [native slash commands](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/). It meant my idea for a portable, backend toolkit could have a first-class, user-friendly interface right in my terminal. The concept for Speedgrapher was now clear: a dedicated MCP server to house a writing toolkit, exposed as simple slash commands.

## Vibe Writing Explained

Before we dive into the technical journey of building Speedgrapher, I want to take a moment to explain what I mean by "vibe writing." You've likely heard the term "vibe coding" floating around — it describes the increasingly common practice of developers using natural language prompts to guide an AI in generating code. It's a fluid, conversational approach where the developer sets the high-level direction, and the AI handles the boilerplate and implementation details.

"Vibe writing" is the natural extension of this concept into the world of words. For me, it's about transforming the solitary act of writing into a dynamic, collaborative conversation with an AI partner. Rather than getting caught up in the mechanics of sentence structure, grammar, and finding the perfect word, I can focus on the core message — the "vibe" I want to create. I provide the initial spark — a rough idea, a personal story, a frustrating problem — and the AI helps me shape it into a structured, coherent narrative.

While I'm not the first person to use this term, it's still an emerging concept. It represents a fundamental shift in how we approach content creation, moving from a purely manual process to a human-AI partnership.

## Starting Simple: A Haiku Generator

Every good technical journey starts with a "Hello, World." For Speedgrapher, my "Hello, World" was a haiku. I needed a simple, low-stakes way to prove that I could expose a prompt as a slash command. What could be simpler than asking an AI to write a poem?

My first attempt was naive. I created a `/haiku` prompt that took a `--theme` argument. The prompt itself was simple: `"generate a haiku based on the theme %s"`. I fired up the Gemini CLI, my Speedgrapher project loaded as context, and I typed:

`/haiku --theme=flowers`

The result was... not a poem. The model, seeing the Go code in my project, interpreted my request as an instruction to *add a haiku feature to Speedgrapher*. It started planning to edit my Go files. I quickly hit `ESC` to abort, as I had to rethink my strategy.

This experience was a powerful reminder of a core principle in prompt engineering: the need to balance ambiguity and context. In many of my prompts, I intentionally use a degree of ambiguity to give the model the flexibility to reason and infer information. For example, my `/review` prompt simply says to "review the article we have been working on." It doesn't specify a file name like `DRAFT.md`. This ambiguity is a powerful tool in a conversational workflow, as it allows the model to identify the relevant text from our recent interactions without needing a rigid, explicit file path.

In the case of the haiku, however, the ambiguity was unconstrained. The primary context was a Go project, which led the model to a logical but incorrect conclusion: that I wanted to modify the code. It wasn't wrong; it was just making a reasonable inference. Because I wanted a very specific, non-code-related outcome, my task in this case was to reduce the ambiguity by providing a much clearer context for my intent.

After a few more tries, I landed on the following prompt:

```go
// The final, working prompt for the haiku command.
prompt = fmt.Sprintf("The user wants to have some fun and has requested a haiku about the following topic: %s", topic)
```

While I'm not sure this is the best way to phrase my intent, it did fit my purpose, and the model consistently produced haikus after that. With the core concept proven, I was ready to build more practical prompts.

## Building a Writer's Toolkit

The haiku experiment confirmed that the core concept was sound, so I moved on to more practical applications. My `GEMINI.md` files had become a collection of useful but non-portable prompts for tasks like reviewing, translating, and outlining my articles. Because they were tied to specific projects, I would often forget to copy them to new ones. An MCP server was a logical next step to make these tools portable.

I started by migrating three of my most-used prompts into Speedgrapher: `interview`, `review`, and `localize`. The core of these prompts is a set of "editorial guidelines." For example, the localization guideline includes a rule to not translate technical terms, which helps ensure consistency across the three languages my blog supports. This approach of creating "editorial guidelines as code" is a way to build a structured system that maintains a consistent voice and quality, much like a linter does for code.

All the prompts in Speedgrapher were generated with the help of Gemini, but for the `review` prompt, I took a slightly different approach. I asked the model to analyze my past articles and generate a set of editorial guidelines based on my writing style. The result was a solid first draft, but it's a prompt that I'm constantly refining.

Here is the current version of the prompt, embedded directly from the Speedgrapher source code on GitHub:

{{< github user="danicat" repo="speedgrapher" path="internal/prompts/review.go" start="18" end="28" >}}

With the core prompts in place, it was time to work on automating other important parts of my job.

## Readability Matters

As a technical writer, my biggest challenge is finding the sweet spot between clarity and complexity. If a text is too simple, it can feel childish. If it's too complex, it becomes unreadable. Readability isn't just about making things easy; it's about making them engaging and intellectually stimulating.

The good news about readability is that it can be measured. While no metric is perfect, the [Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index) is a great tool for getting a baseline. The Gunning Fog Index is a readability test that estimates the years of formal education a person needs to understand a text on the first reading. A score of 12, for example, means the text is at the reading level of a U.S. high school senior.

The index is calculated based on the following algorithm:
*   Take a section of text of 100 or more words.
*   Find the average sentence length.
*   Count the number of "complex" words (words with three or more syllables).
*   Add the average sentence length to the percentage of complex words.
*   Multiply the result by 0.4.

Or, for the mathematically inclined, this algorithm translates to the following equation:

{{< katex >}}
\[
 0.4 \times \left[ \left( \frac{\text{words}}{\text{sentences}} \right) + 100 \left( \frac{\text{complex words}}{\text{words}} \right) \right]
\]

While the original intent of the fog index is to estimate the years of education required to understand the text, I think it is not helpful to frame it specifically in terms of years of education, so I took the liberty to customise it for my own needs. First, I simplified the calculation to ignore special cases: one of the most complex parts of the algorithm is how to define if a word is complex (pun intended). While the base case is to consider a word complex if it has three or more syllables, it creates special cases where you ignore certain word terminations like -ing, -ed and -es.

This created a surprising amount of problems during implementation. I didn't need to be precise and  I was happy with overestimating complexity on behalf of simplicity. For this purpose, I ignored all special cases and considered two basic rules for counting syllables: 1) the number of sylables in a word is estimated by the number of vowel groups, and 2) complex words are words that have three or more syllables (no exceptions)

I also created a classification system that shifts the focus from years of education towards a more pragmatic approach to readability.

| Score | Classification | Description |
| :--- | :--- | :--- |
| >= 22 | Unreadable | Likely incomprehensible to most readers. |
| 18-21 | Hard to Read | Requires significant effort, even for experts. |
| 13-17 | Professional Audiences | Best for readers with specialized knowledge. |
| 9-12 | General Audiences | Clear and accessible for most readers. |
| < 9 | Simplistic | May be perceived as childish or overly simple. |

With the customized Gunning Fog Index implemented as a `fog` tool, the final step was to create a user-friendly interface for it. I created a `/readability` prompt that calls the `fog` tool and presents the results in a clear format. This follows my design guidelines for Speedgrapher: building focused, single-purpose tools and then composing them into more powerful, user-friendly workflows.

## Automating the Writer Workflow

The individual prompts were helpful, but I still had a lot to automate until I got my dream workflow. For the next few iterations I would test drive the prompts and map the process gaps to come with new prompts and/or fine tune the existing ones. Here are the prompts I'm currently using:

**Main Flow**
* `/interview`: Interviews an author to gather material for an article. This is usually the starting point for a writing session.
* `/outline`: Generates a structured outline of the current draft, concept or interview report.
* `/voice`: Analyzes the voice and tone of the user's writing to replicate it in generated text.
* `/expand`: Expands a working outline or draft into a more detailed article. Can also be used with a `hint` argument to do a focused expansion of a specific paragraph or section.
* `/review`: Reviews the article currently being worked on against the editorial guidelines.
* `/readability`: Analyzes the last generated text for readability using the Gunning Fog Index.
* `/localize`: Translates the article currently being worked on into a target language.
* `/publish`: Publishes the final version of the article.

**Optional**
* `/context`: Loads the current work-in-progress article to context for further commands. This is used to "remind" the model of the current draft if necessary, and is often run before commands like `/readability` or `/review` that operate on the full text.
* `/reflect`: Analyzes the current session and proposes improvements to the writing process. This is useful for improving prompts and editorial guidelines.

The goal was to move from a collection of useful commands to a single, streamlined process that could guide an article from a simple idea to a polished, multi-lingual publication.

The diagram below is a simplified representation of my workflow:

{{< mermaid >}}
flowchart TD
    A[Idea] -->|/interview| B[Interview Transcript]
    B -->|/outline & /voice| C[Structured Outline]
    C -->|/expand| D[Draft Article]
    D -->|/review & /readability| E[Reviewed Draft]
    E -->|/localize| F[Localized Versions]
    F -->|/publish| G[Published Article]
{{< /mermaid >}}

The process begins with an `/interview` to flesh out the core concepts of an idea. The resulting transcript is then transformed into a structured plan using `/outline`, and aligned with my personal writing style with `/voice`. With this foundation in place, I enter an iterative loop of using `/expand` to build out the draft, and `/review` and `/readability` to refine it.

Once the article is approved, I use `/localize` to create versions for other languages, and `/publish` to finalize the process. The optional `/reflect` prompt can then be used to analyze the session and generate notes for future improvements, creating a cycle of continuous refinement.

## Conclusion

Just as we use linters and tests to bring structure to our code, we can apply similar principles to our creative workflows. The writing process has many repetitive tasks that can be automated. By building a personal toolkit of prompts, we can offload the boilerplate and focus on the core ideas of our work.

This is the value of a tool like Speedgrapher: “vibe writing” is not about replacing the writer, but about augmenting the writing process. By adding an MCP server to the mix, it brings a helpful layer of structure to a sometimes disorganized workflow, ensuring that best practices are followed. The same can be applied to any AI-assisted process: by treating your own prompts as reusable, portable assets, you can create a system that evolves with your process, allowing you to focus on the creative aspects of your work, one prompt at a time.

## What's Next?

The journey with Speedgrapher is far from over. While the current toolkit is focused on text, the next logical step is to embrace multimodality. I'm exploring how to integrate tools for generating hero images, creating more sophisticated diagrams from the text, and even suggesting layout optimizations. The goal is to continue building a personal toolkit that handles more of the non-writing tasks, freeing me up to focus on the content itself.

## Resources

*   **[Speedgrapher Project](https://github.com/danicat/speedgrapher):** The source code for the MCP server discussed in this article.
*   **[How to Build an MCP Server with Gemini CLI and Go]({{< ref "/posts/20250729-how-to-build-an-mcp-server-with-gemini-cli-and-go" >}}):** The previous article that inspired this journey.
*   **[Model Context Protocol (MCP)](https://modelcontextprotocol.io/):** The official website for the protocol.
*   **[Gemini CLI Announcement](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/):** The blog post that announced custom slash command support.
*   **[Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index):** Learn more about the readability metric.