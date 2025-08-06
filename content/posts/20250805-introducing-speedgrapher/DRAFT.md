# Introducing Speedgrapher: An MCP Server for Vibe Writing

I have a confession to make: I love building things, but I don’t always love the boilerplate that comes with it. My mind is often buzzing with ideas for new articles, but the process of structuring them, ensuring they meet my own editorial standards, and even just getting the tone right can sometimes feel like a chore. This is the story of how a deep dive into a technical specification led me to build a tool that has fundamentally changed my writing process. It’s a story about a side quest that became the main quest, and how I built Speedgrapher, my personal MCP server for "vibe writing."

It all started with a bit of curiosity. I was researching the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) for a previous article, focusing on how it allows AI agents to use tools. But as I read the spec more carefully, a small detail caught my eye: besides tools, the protocol also mentioned prompts and resources. A lightbulb went off. I realized that the prompts I had scattered across `GEMINI.md` files—my personal collection of best practices and instructions for the Gemini CLI—could be formalized, packaged, and made portable.

Then came the second "aha!" moment, a bit of pure serendipity. The very same day I was exploring this idea, the Gemini CLI introduced support for custom slash commands. Suddenly, my scattered prompts could become a set of powerful, reusable commands, accessible directly from my terminal. The idea for Speedgrapher was born: a dedicated MCP server to house my personal writing toolkit.

## Chapter 1: The First Experiment - A "Haiku" and a Hard-Won Lesson

Every good technical journey starts with a "Hello, World." For Speedgrapher, my "Hello, World" was a haiku. I needed a simple, low-stakes way to prove that I could expose a prompt as a slash command. What could be simpler than asking an AI to write a poem?

My first attempt was naive. I created a `/haiku` prompt that took a `--theme` argument. The prompt itself was simple: `"generate a haiku based on the theme %s"`. I fired up the Gemini CLI, my Speedgrapher project loaded as context, and I typed:

`/haiku --theme=flowers`

The result was... not a poem. The model, seeing the Go code in my project, interpreted my request as an instruction to *add a haiku feature to Speedgrapher*. It started planning to edit my Go files. I quickly hit `ESC` to abort, a bit startled but also amused.

This was my first hard-won lesson in prompt engineering: **context is everything, and ambiguity is the enemy.** The model wasn't wrong; it was just making a logical inference based on the information it had. I needed to be more explicit.

After a few more tries, I landed on a much better prompt:

```go
// The final, working prompt for the haiku command.
prompt = fmt.Sprintf("The user wants to have some fun and has requested a haiku about the following topic: %s", topic)
```

This version worked perfectly. By adding the phrase "The user wants to have some fun," I gave the model crucial context about my *intent*. I wasn't asking it to code; I was asking it to play. That small change made all the difference. The agent now understood its role and happily generated a haiku about flowers. It was a small victory, but it proved the core concept was sound.

## Chapter 2: From Experiments to Essentials - Building a Writer's Toolkit

With the haiku success under my belt, I was ready to move on to more practical applications. My `GEMINI.md` files were getting cluttered with instructions for reviewing, translating, and outlining my articles. These prompts were useful, but they were tied to specific projects. Every time I started a new project, I'd forget to copy them over. An MCP server was the perfect solution to make these tools portable.

I started by migrating three of my most-used prompts into Speedgrapher: `interview`, `review`, and `localize`. The `localize` prompt was particularly important for my blog, which is available in three languages. The core of these prompts was a set of "editorial guidelines." For example, my localization guideline includes a crucial rule:

> "When translating, do not translate technical terms or product names. For example, 'Gemini CLI' should remain 'Gemini CLI' in all languages."

This is an example of "editorial guidelines as code." Just as a linter enforces coding standards, these guidelines ensure my content maintains a consistent voice and quality, no matter the language.

The `review` prompt was the most ambitious. I wanted an AI editor that understood my personal style. To achieve this, I used a meta-approach: I asked Gemini to read all my past articles and generate a set of editorial guidelines based on my writing. The result was a surprisingly accurate and insightful prompt that now lives in Speedgrapher, ready to review my drafts with a click of a button.

## Chapter 3: The Heart of the Machine - Readability as a Science

Of all the tools in Speedgrapher, the `readability` prompt is the one I'm most proud of. As a technical writer, my biggest challenge is finding the sweet spot between clarity and complexity. If a text is too simple, it can feel childish. If it's too complex, it becomes unreadable. Readability isn't just about making things easy; it's about making them engaging and intellectually stimulating.

The beauty of readability is that it can be measured. While no metric is perfect, the [Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index) is a great tool for getting a baseline. It calculates a score based on sentence length and the number of "complex" words (words with three or more syllables). I decided to implement this as a tool in Speedgrapher.

I started by asking the model to implement the Gunning Fog Index for me, but the initial results were overly complex. I didn't need scientific precision; I needed a compass. So, I took a "vibe-coded" approach. I asked the model to create a simpler heuristic: count vowel groups to estimate syllables and consider any word with more than two syllables to be complex. My guiding principle was simple: **it is far better to overestimate complexity than to underestimate it.**

The result was a `fog` tool that could analyze a block of text and return a Gunning Fog score, along with the raw metrics like sentence count and complex word count.

## Chapter 4: Weaving It All Together - Prompts as a User-Friendly API

The final piece of the puzzle was to connect the `readability` prompt to the `fog` tool. I wanted the experience to be seamless. I didn't want to have to manually run the `fog` tool and then interpret the results. I wanted to simply type `/readability` and get a helpful analysis.

This is where the power of modern LLMs shines. I learned that I didn't need any special SDK features to wire the prompt and the tool together. I just needed to write a prompt that clearly explained the steps. This is the prompt that powers the `/readability` command:

```
**Objective: Evaluate Readability**

You are an expert editor. Your task is to analyze the most recent text you have generated in this session and assess its readability using the Gunning Fog Index.

**Analysis Steps:**

1.  **Identify the Text:** Use the most recent, complete text block you generated in this session as the source material.
2.  **Assess Current Readability:** Use the `fog` tool to calculate the current Gunning Fog Index and classification for the text.

**Your Task:**

Now, execute the plan. First, call the `fog` tool on the text you just wrote. Then, provide your analysis.
```

That's it. The model understands this prompt, calls the `fog` tool, and then presents the results in a human-readable format. It's a beautiful example of the Unix philosophy: "Do one thing, but do one thing well." The `fog` tool calculates the score, and the `readability` prompt provides the user-friendly interface.

## Conclusion: The Power of Portable Prompts

This journey with Speedgrapher has taught me a valuable lesson: the most powerful AI tools are the ones we build for ourselves. By treating my personal prompts as reusable, portable assets, I've created a toolkit that grows with me. My best practices are no longer lost in scattered notes; they are now a set of powerful commands, accessible from any MCP-compliant agent.

The fact that these prompts map so cleanly to slash commands has also opened up new possibilities for more complex workflows. I'm no longer just writing; I'm "vibe writing"—having a conversation with my tools, iterating on ideas, and building a better writing process, one prompt at a time.

And it all started with a haiku. =^.^=

***

### Resources

*   **[Speedgrapher Project](https://github.com/danicat/godoctor):** The source code for the MCP server discussed in this article.
*   **[Model Context Protocol (MCP)](https://modelcontextprotocol.io/):** The official website for the protocol.
*   **[Gunning Fog Index](https://en.wikipedia.org/wiki/Gunning_fog_index):** Learn more about the readability metric.