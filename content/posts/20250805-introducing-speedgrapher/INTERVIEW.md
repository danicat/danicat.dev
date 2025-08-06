# Interview Transcript

**Q: Hello! I'm here to help you craft a compelling story about your technical journey. To start, could you tell me the high-level goal of the article you have in mind? What's the main story you'd like to share with the world?**

A: I would like to write an article entitled "Introducing Speedgrapher: An MCP Server for Vibe Writing"

**Q: That's a great title. It connects a technical tool, "Speedgrapher," with a creative process, "Vibe Writing." I'm curious, what was the personal motivation or the specific problem that led you to build an MCP server for this purpose? Tell me about the moment you realized you needed to create this.**

A: So, speedgrapher is an MCP server that I use for my own writing. It's a collection of prompts that help me streamline my writing process. Everything started when I was doing my research about MCP servers. In my previous article I wrote about writing MCP servers to expose MCP tools, but reading the MCP spec carefully I found out there was more to that. Besides tools, the spec also mentions prompts and resources. Then I started to become curious about how to use prompts effectively. Eventually I realised that prompts could be mapped to "slash commands" in Gemini CLI and I decided to build a conceptual tool to help with my workflow leveraging that. This is how speedgrapher was born.

**Q: That's a fascinating origin story, connecting the dots from the MCP spec to a practical writing tool. It's impressive how you translated a detail from the specification into a real-world application.

Could you walk me through the early days of developing Speedgrapher? What was the very first prompt you created, and what was that "aha!" moment like when you saw it working as a slash command for the first time?**

A: First I needed to prove that the prompt worked, so I didn't want to start with anything complex, as I was not sure about what to expect. So the first prompt I added to the server is not a very practical one... it is called "haiku" and, as you can expect, it asks the model to write a haiku for the user. The first time I tried it was fun - at that moment I didn't know the Gemini CLI supported custom slash commands, but once I configured the server there was a "/haiku" command for me to call. Of course, it didn't work the first time! Since the project that was in context, a model interpreted the haiku prompt as a request to add a "haiku feature" to the code base I was working on. So it took me a few more shots to fine tune the prompt so that it unanbiguously sounds like a "make me a haiku" command to the models.

**Q: That's a great story. It highlights the often-overlooked part of development: the fine-tuning and iteration needed to get something to work as intended. It's one thing to have an idea, and another to make it robust.

You mentioned that the model initially misinterpreted your "haiku" prompt. Could you share the specific wording of the first prompt you tried and the final version that worked? I think readers would find it very insightful to see the evolution and understand the nuances of prompt engineering.**

A: The first prompt was something like "generate a haiku based on the theme %s" and the theme was an argument for a prompt, so it would be called with "/haiku --theme=flowers" - the context I had was the speedgrapher project which is written in Go, so the model tried to edit the /internal/prompts/haiku.go file to update it's prompt. I quickly aborted the operation and came with a better prompt. The final prompt is: `"The user wants to have some fun and has requested a haiku about the following topic: %s"`.

**Q: After solving the "haiku" problem, what was the next, more practical prompt you built for your writing workflow? What real problem were you trying to solve with it?**

A: After I was happy with the haiku prompt I started brainstorming what would be useful prompts to support me in my writing journey. I got a little carried away and came up with three at once: interview, review and localize. Those were the easiest ones because I already had been using them for my blog, but instead of living in an MCP server they were in my GEMINI.md file. It did work for my needs, but I didn't like polluting my GEMINI.md with processes, also, GEMINI.md is not as portable as an MCP server.

**Q: That's a very practical reason to move to an MCP server—making your tools portable and keeping your configuration clean. Let's dive into one of those prompts. The `localize` prompt sounds especially powerful for your multi-language site. Could you describe the process of converting that prompt from your `GEMINI.md` file into a formal MCP prompt? What did the original instruction look like, and what challenges did you face in making it a reliable, reusable command?**

A: The localize and review prompts were all based on applying certain "editorial guidelines". For instance, for localization I often add as a guideline to not translate technical terms or product names. As for editorial, I want the guidelines to help me achieve a balance between professionalism and personalisation. I have clear goals I want to achieve with my content and the editorial guidelines should help me achieve them... just like a code review does for code, I wanted to have an automated process that would check and validate my content.

**Q: That's a really smart approach. Treating your content with the same rigor as code, using "editorial guidelines" like a linter, is a powerful concept. It's a great way to ensure consistency and quality.

Let's focus on the `review` prompt. Could you give me a specific example of one of your editorial guidelines? For instance, what's a rule you have for balancing professionalism and personalization, and how did you translate that into a prompt that the model can understand and apply effectively?**

A: The review prompt is a tricky one, because I didn't write it myself. I'm still in the process of refining it, but as you can see it talks about the article structure and voice. I actually used Gemini to generate this prompt by reading my past content. I think there is still a lot to be improved, but it is a good start.

**Q: That's a fascinating, meta-level approach—using Gemini to analyze your own work and create a personalized editor. It's a powerful testament to how these tools can be adapted to very specific, personal workflows.

You mentioned it's a good start but needs improvement. This is a key part of any technical journey. Could you tell me more about the process of generating these guidelines? What kind of prompt did you give Gemini to get this output? And, based on your experience using it so far, where does this `review` prompt still fall short? What are the specific areas you feel need more refinement?**

A: Those are questions still to be answered. The more I use it the more I am able to notice the rough edges and refine it. It's an iterative process. I don't have the exact prompt that generated it anymore, but it was something as simple of "please read all my content and extract the best editorial practices that capture my voice, style and writing standards - document it in a file for later usage"

**Q: That makes perfect sense. The iterative process of refining prompts is a journey in itself, and it's a great lesson to share. The fact that you started with a simple, high-level prompt and are now in a cycle of continuous improvement is a very relatable story for anyone working with AI.

Before we move on, is there anything else you'd like to add about the `review` prompt or the process of developing your editorial guidelines? I want to make sure we've captured all the important details of this part of your journey.**

A: not really, I think the beauty of speedgrapher is what comes next, which is the readability prompt

**Q: Excellent. Let's dive into that. The "readability" prompt sounds like a significant evolution. You called it "the beauty of speedgrapher," which is a powerful statement. Could you tell me what it does and what makes it so special to you? What problem were you trying to solve when you created it?**

A: when writing technical texts, one of my top concerns is to make it readable. readability is not just about making a text easy to read... you make it too easy it becomes childish and boring, you make it too complex it will be super hard to read, sometimes even unreadable. Readability is about finding that ideal spot where you get the message you want to convey to the audience to be transmitted as clearly as possible, without sacrificing their engagement with the content. It is not just about dumping content, it is about making it interesting, engaging and intellectually stimulant. one of the great things about readability is that, while there is not a perfect measure, we do have ways of measuring it to give good enough guidance to fine tune the final text. One of such a metric is the Gunning Fog Index, which I decided to implement as part of the readability prompt (and companion tool)

**Q: That's a brilliant insight. You've moved beyond simple prompts and into a more sophisticated, data-driven approach to writing. Using a concrete metric like the Gunning Fog Index to guide your work is a fantastic example of blending art and science.

This sounds like a significant step up in complexity. Could you walk me through the process of implementing the Gunning Fog Index? Did you write the calculation logic yourself, or did you find a library to help? And how did you integrate that calculation into a prompt that could provide actionable feedback to you as a writer? I'm especially curious about the technical challenges you faced in bridging the gap between a numerical score and creative guidance.**

A: My implementation of the gunning fog index was essencially a carefully tuned vibe coded implementation. At first I just asked the model to produce one implementation for me, but it was overcomplicating the implementation. I don't really need precision, I just care about a baseline that I can use to tune my writing style. So I eventually asked it to adopt a simple heuristic counting vowel groups and ignoring rules about silent terminations like -es -ed etc. For my goal, anything longer than two syllables is a complex word. This is based on a simple premise: it is far better to overestimate complexity than to underestimate it. Also, in the end, it is how the final text reads that matters. The metric is just a compass.

**Q: That's a brilliant, pragmatic solution. "It is far better to overestimate complexity than to underestimate it" is a fantastic guiding principle for this kind of tool. It really shows a developer's intuition shaping the outcome, rather than just accepting a standard, overly-complex solution.

So, once your tool calculates this Gunning Fog score, how does the `readability` prompt actually use that information? Does it simply report the number back to you, or does it provide specific, actionable advice? For instance, if a paragraph's score is too high, what does the prompt's feedback look like? Does it suggest simpler words or shorter sentences?**

A: I initially started with a prompt that would invoke the tool and then propose improvements on the text based on a "readability" target, but soon dropped that approach as it felt that the prompt was doing too much. Nowadays it not only reports the fog index but also the intermediary metrics like avg sentence length and word and complex word counts. As how to use this information, its best for the user to decide. I believe that Being less opinionated in the prompt makes up for more useful tools. Do one thing, but do one thing well is the design phylosophy for this tool. Kind of like the unix phyolosophy.

**Q: That's a fantastic design philosophy, drawing a parallel to the Unix philosophy. It shows a deep respect for the writer's autonomy, providing data instead of prescriptive changes. It seems like this makes Speedgrapher more of a powerful instrument than an automated editor.

You've given me a great overview of the evolution of the prompts within Speedgrapher, from the initial 'haiku' to this data-driven 'readability' prompt. At the beginning of our conversation, you mentioned that Speedgrapher is a collection of prompts and, as of today, one tool. Could you tell me about this tool? What does it do, and how does it fit into the Speedgrapher ecosystem?**

A: the tool i mentioned is actually the fog index tool. I decided to also add a prompt to "expose" this tool because it would make for a nicer interface, as "readability" becomes a slash command. I was also exploring the interaction between tools and prompts, so this was the natural consequence of this experiment.

**Q: That's a very elegant design, where the `readability` prompt acts as a user-friendly wrapper for the `fog` tool, making a complex calculation feel like a simple command. It's a great example of how prompts and tools can work together to create a seamless user experience.

Could you walk me through how you implemented that connection in Go? Specifically, how does the `readability` prompt handler call the `fog` tool and then present the results? What was the most challenging part of wiring the prompt and the tool together? Were there any surprises or "gotchas" in the MCP SDK that you had to work around?**

A: There isn't much mystery. Calling a tool inside a prompt is done in the same way you do using a normal conversational experience. Here is the full prompt:

`**Objective: Evaluate Readability**

You are an expert editor. Your task is to analyze the most recent text you have generated in this session and assess its readability using the Gunning Fog Index.

**Analysis Steps:**

1.  **Identify the Text:** Use the most recent, complete text block you generated in this session as the source material.
2.  **Assess Current Readability:** Use the ` + "`fog`" + ` tool to calculate the current Gunning Fog Index and classification for the text.

**Your Task:**

Now, execute the plan. First, call the ` + "`fog`" + ` tool on the text you just wrote. Then, provide your analysis.`

**Q: Looking back at the entire journey of creating Speedgrapher, what would you say was the single most important lesson you learned? It could be about prompt engineering, the MCP spec, or even your own writing process. What's the key takeaway you'd want to share with other developers who are just starting to explore this space?**

A: Just to add to your comment, this is not a feature of the MCP SDK itself, but it relies on the underlying model to understand tool calling and make the right choice. Which most modern models will do, so not a big surprise there.

As for the journeys and takeways... I think the most satisfying outcome of this project is to realise that all my prompts can now be mapped to reusable assets and be made portable using MCP, instead of them being lost in random notes or even GitHub repos. Packing your prompts in MCPs makes them easily accessible for any MCP compliant agent. Before realising this I was filling my GEMINI.md files with "best practices" that I would often forget to port to new projects, but now I'm slowly moving away those practices to dedicated MCPs that will help me streamline my coding, writing, and any other important tasks of my job. The fact that prompts map nicely to slash commands also can make them into building blocks for more complex flows.

**Q: This has been an incredibly insightful conversation. We've traced the journey of Speedgrapher from a simple experiment to a sophisticated writing assistant, and you've shared some fantastic lessons about iterative development, prompt engineering, and the power of the Model Context Protocol.

Is there anything else you'd like to add before we conclude?**

A: Just a fun fact - the day I decided to write about prompts was the day that Gemini CLI introduced the support for custom slash commands, so I kind of stumbled into this prompt to command mapping by accident and only a few hours later would realise that it was a brand new feature.
