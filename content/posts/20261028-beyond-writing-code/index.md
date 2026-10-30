---
title: "Beyond Writing Code: Taste, Skills, and Tech Leads"
date: 2026-10-28
draft: false
categories:
  - Perspectives
tags:
  - agent-skills
  - antigravity
  - career
  - code-review
  - subagents
  - vibe-coding
heroStyle: big
slug: "beyond-writing-code"
aliases:
  - "/posts/20261028-beyond-writing-code/"
summary: "When models have all the knowledge of the world, what is left for the humans? If you know it all, everything seems equally important. Our role has evolved from execution to guidance."
description: "Why software engineering in the age of coding agents shifts from writing syntax to architectural taste, subagents, agent skills, and acting as tech leads."
proficiencyLevel: "Intermediate"
---

When foundation models have all the knowledge of the world baked into their weights, what's left for the humans?

For decades, our identity as software engineers was tied to the craft of writing code: mastering syntax, memorising standard libraries, and taking pride in the systems we built with our own hands. We were told AI would simply take care of the boring boilerplate and hand us back hours of free time every week. Instead, watching a model generate in seconds what took us years to learn has triggered an identity crisis across the industry. If the machine can write the code, where do we fit in?

Humans should be the beacon. Exactly because models know *all*, every path looks equally plausible. They cannot prioritise and sequence information to create with **taste**. This is where our new job kicks in: as if we were spelunking in a dark cave, we hold the flashlight, pointing at the path and deciding where the model should focus next.

Our role evolved from writing code to orchestration. This is actually not new for many of us, even before AI coding: it is the natural progression of the software engineering career. When you move up the ladder, you get less hands-on and start wearing the coach hat more and more. You become responsible for the output of your team, you stop reading all lines of code and focus on the critical parts, you learn to delegate and to deal with the pros and cons of that. You become a tech lead.

## Everyone has been promoted to tech lead

![In 2026 Every Engineer Has Been Promoted to Tech Lead](everyone-is-a-tech-lead.png "Congratulations on your promotion to Tech Lead. (P.S.: HR confirmed your compensation band remains unchanged.)")

In the traditional career path, that transition happened gradually. You spent years in the trenches evolving from Junior to Senior before eventually stepping into a lead, engineering manager, or architect role, focused on the big picture while more junior engineers were tasked with the implementation details.

Now everyone has their virtual team of coding agents, so even the most junior engineer needs to learn how to become a good tech lead. Not to guide humans, but to guide the agents.

"Guide to do what?" is the most important question. We are still building programs, but the way we describe them has changed.

There are few books that I've read in my life that I consider to be evergreen. One of those is *Structure and Interpretation of Computer Programs* (aka SICP). Why is this book relevant now? Because of the way it describes programs and how we as engineers "conjure spells" to build them:

> *"Computational processes are abstract beings that inhabit computers. As they evolve, processes manipulate other abstract things called **data**. The evolution of a process is directed by a pattern of rules called a **program**. People create programs to direct processes. In effect, we conjure the spirits of the computer with our spells."*
>
> *"... The programs we use to conjure processes are like a sorcerer's spells. They are carefully composed from symbolic expressions in arcane and esoteric programming languages that prescribe the tasks we want our processes to perform."*
>
> — Harold Abelson and Gerald Jay Sussman, *Structure and Interpretation of Computer Programs* (1985)

We are still in the business of conjuring processes to instruct the computers. Only the language in which we craft our spells has changed. Now that we write those spells in plain English instead of arcane symbolic expressions, orchestrating agents and ensuring they are on the right track is the most important part of the job, which implies less writing code and a lot more reading. You are also writing more specifications (prompts) than ever. If you have been a tech lead before AI, all of this will simply sound like business as usual.

How do you get good at this? In the old world, the good tech leads and managers would spend a lot of time coaching and mentoring their teams so they would develop the same critical thinking and taste that they developed during their careers. They would gradually transfer the knowledge of processes, behaviours and techniques often developed at the cost of many battle scars. That pedantic code review I just did? It was because my own lack of attention to details made me crash production many years ago.

In today's world, beyond learning from our mentors and our own battle scars, we need to carefully observe the behaviour of our agents and learn from **their** mistakes. Because those mistakes accumulate at blinding speed, agents provide the ultimate compressed feedback loop: you watch dozens of subtle failure modes, race conditions, and hallucinated patterns unfold in an afternoon, building architectural discernment at ten times the speed of traditional release cycles.

In essence, we are now in the business of orchestrating the flow of information, ensuring that agents always have the optimal environment for building the outcomes we desire.

Given the rate of evolution we are seeing, what is optimal today will most likely not be optimal tomorrow, so we need to constantly refine our tools and processes to keep up with the pace of evolution. More than ever, software is disposable. Our spells need to become rapid and precise, as tomorrow we will need new ones.

## Structuring information so models can act on it

> *"Master software engineers have the ability to organize programs so that they can be reasonably sure that the resulting processes will perform the tasks intended. They can visualize the behavior of their systems in advance. They know how to structure programs so that unanticipated problems do not lead to catastrophic consequences, and when problems do arise, they can **debug** their programs. Well-designed computational systems, like well-designed automobiles or nuclear reactors, are designed in a modular manner, so that the parts can be constructed, replaced, and debugged separately."*
>
> — Harold Abelson and Gerald Jay Sussman, *Structure and Interpretation of Computer Programs* (1985)

If you take the quote above and replace the word "programs" with "agents," the whole framework still stands. We need to learn how to structure our agentic coding processes so that they produce well-designed computational systems.

One of the most important ways of giving knowledge to an agent is via [Agent Skills]({{< ref "/posts/20260829-the-pragmatic-guide-to-agent-skills" >}}). Because models behave very similarly to humans (shouldn't be a surprise, they were trained on data that we produced), the same techniques we applied to improve the productivity and output quality of our teams are applicable to agents. This makes for a fertile ground in creating skills that will influence model behaviour inspired by your own experiences and battle scars, but also industry best practices.

One such technique is to break down complex tasks into smaller pieces. A cornerstone of the Agile process, which now can be applied to agentic coding via the constructs of spec-driven development, implementation plans and [subagents]({{< ref "/posts/20260722-the-rise-of-the-subagents" >}}).

I personally treat subagents as the building blocks of the information flow I'm designing. This is in many senses similar to dealing with concurrency in traditional programs, but also plays an important role in isolating context. After all, "attention is all *they* need." By breaking down tasks into orthogonal pieces that can be individually processed by the subagents, you are ensuring that the agent's precious attention span is focused on that single task alone, improving quality and task adherence.

Whenever you are manually weaving workflows or building your skills for posterity, I've found these points to be particularly helpful:
1. **Coordinator role**: It is useful to keep the main session free by making sure the coordinator always delegates execution to subagents, releasing the main session as fast as possible. That way you can keep interacting with it while the subagents are working, sending comments to steer them in the right direction in real time.
2. **Orthogonality of tasks**: Slice tasks in orthogonal units to minimise dependencies or conflicts with each other. First, you have to understand and respect sequential dependencies. Then, you can parallelise with confidence that subagents won't overwrite each other's work or make incompatible assumptions. 
3. **Recreating proven workflows**: Just like with coaching and mentoring, patterns that work for human teams and classical software engineering also work for agents. Constantly revisit the literature for patterns old and new that you can teach to the agents.

I have been gradually building my [skill catalog](https://skills.danicat.dev) over the past few months. Because skills *rot* fast as models and agent harnesses evolve, it's important to notice that this isn't a fire-and-forget type of thing. I'm constantly evolving them, to the point I decided to add a "freshness score" to my skills to remind me which ones are becoming obsolete so I can either try to give them a new life or trim them from my catalog.

Here are a few examples of skills I actively use on a daily basis:

- **`double-diamond`**: a product development workflow based on the double diamond methodology from Agile practices that forces discovery before implementation.
- **`parallel`**: a multi-purpose fan-out / map-reduce pattern that breaks a task into multiple threads and reconciles the results at the end.
- **`uno-reverse`**: adversarial review to counter the model's overconfidence and tendency to over-engineer.
- **`less-is-more`**: a design philosophy telling the model to do one thing at a time, and do it well, instead of creating bloated apps.
- **`intercom`**: a skill that teaches my agents to talk with each other when they are working on related projects.
- **`show-me-the-money`**: a skill to mine the conversation history for features that I discussed with the model, but completely forgot to implement.
- **`abstract`**: generalises examples and circumstances into knowledge using philosophical principles.

## Broken windows in agentic code

> *"A computational process, in a correctly working computer, executes programs precisely and accurately. Thus, like the sorcerer's apprentice, novice programmers must learn to understand and to anticipate the consequences of their conjuring. Even small errors (usually called **bugs** or **glitches**) in programs can have complex and unanticipated consequences. ... Real-world programming, however, requires care, expertise, and wisdom."*
>
> — Harold Abelson and Gerald Jay Sussman, *Structure and Interpretation of Computer Programs* (1985)

If we don't anticipate the consequences of what we conjure, an unguided agent will multiply small structural flaws faster than we can even realise them. This problem is not new; it actually always happened even with human-written code, but we didn't notice it that much because it was a much slower process. Some people tried to stop it and were always deemed pedantic code reviewers, loved by a selected few but hated by most.

In psychology, it is a phenomenon known as the Broken Window Theory based on an experiment dating back to 1969. Researchers left an abandoned car in a good neighbourhood for a week. Nothing happened to it. On week two, one of the researchers used a sledgehammer to break a window. Within hours, the car was stripped bare.

The broken glass was the trigger. It signalled that nobody cared, so collectively the community stopped caring as well. I've been advocating for years that the same thing happens with code.

Once there is a broken window in code (for example: typos, bad names, poor tests, or wrong abstractions), it triggers our psychological system that stops caring. And the problem escalates with time, so minor flaws can snowball into huge ones.

In the agentic world, the problem is amplified, as existing code will bias the model when generating the new one. If existing code is bad, unless properly instructed, the agent will only amplify its problems, as it's deemed an acceptable standard to follow.

Code reviews are the less glorious part of the engineering work, to the point that many engineers don't even consider it to be part of the work. It is, in fact, one of the most important ones.

This **doesn't** mean reviewing every single line of code.

Nobody likes to review thousands of lines of AI-generated code, including myself. Taste isn't about checking syntax or catching mechanical bugs—that's what deterministic tooling is for. As I wrote in [How to Do Code Reviews in the Agentic Era]({{< ref "/posts/20260303-code-reviews-in-2026" >}}), if a compiler, linter, or automated test can catch an issue, a human shouldn't be looking at it.

We have to fight AI with AI and automation: use deterministic gates to eliminate the mechanical noise, focus your review time on critical architectural boundaries, and sample the rest. Because cycle times today are orders of magnitude faster than "artisanal coding," by sampling, even if you don't catch an issue immediately, you will quickly have another chance to catch it. Therefore, direct your reviews to the really critical parts and trust the process to catch the rest.

## Where the promised free time actually goes

We're building software faster than ever. By removing implementation bottlenecks, all the ideas that used to stay on the shelf until they were replaced by something more urgent suddenly became viable. We can now write prototypes in quick succession to see what works and what doesn't long before they end up in the pile of abandoned projects.

Moving faster, however, doesn't replace the need to build mental models. Our core work is structuring context, constraints, and feedback loops so agents build what we actually intended. If anything, our job has become harder: technical knowledge has become a commodity, but taste and critical sense are still sparse.

Ultimately, we're still sorcerers conjuring the spirits of computers. It's just our spells that have crossed yet another level of abstraction.
