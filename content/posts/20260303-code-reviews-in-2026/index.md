---
title: "How to Do Code Reviews in the Agentic Era"
date: 2026-03-06T00:00:00Z
summary: "A practical guide to modern code reviews. Learn where to invest your time and effort to consistently write quality software in the agentic world."
categories: ["Workflow & Best Practices"]
tags: ["code-review", "vibe-coding", "agentic-coding"]
heroStyle: "big"
---

In 2025 we've seen the rise of agentic coding (apparently the term "vibe coding" is obsolete). Between AI assistants and agentic workflows, features are flying off the shelf at a pace we've never seen before. It's not uncommon for companies to brag that any amount of percentage points of their codebase is now written entirely by AI.

Whether this is a good or a bad thing we're yet to see (I, for one, think it's good), but this gain in writing speed isn't without consequences: reviewing the sheer amount of code produced is draining, and code reviews are quickly becoming the bottleneck. Some teams / open source projects have even adopted the nuclear option of not accepting AI-generated pull requests entirely.

While banning AI can give people a breather, I don't think it's a good option in the long run. "Resistance is futile", as my favourite fictional species would say. To survive this new level of productivity we have to stop doing the work that machines can do better. Fight AI with AI. But not only AI, a good set of old-school deterministic tooling can also do wonders: if a linter can catch an issue, I shouldn't be looking at it. If a formatter can fix it automatically, I really don't care about it.

Here is my "unpopular" take. I don't care if a human or an agent wrote the code. In open source, everything is zero-trust anyway. Whether the code was written by a state of the art model or a teenager in Sri Lanka shouldn't really matter. In theory, human-generated PRs would be smaller, but after 20-ish years working in this industry I had my fair share of mega-PRs so I can ensure you that dealing with mega and/or sloppy PRs isn't a new problem.

I evaluate code at face value. Does it work? Is it safe? Does it fix a known issue? Does it align with our roadmap? Does it comply with our standards?

This is why in today's article I would like to talk a little bit about how I am approaching code reviews, not only when dealing with external contributions, but also when dealing with my own AI-generated code as, in fact, coding with an AI means code reviewing the AI the whole time.

## What I actually care about

When I review code nowadays, I am becoming increasingly higher level. In a sense, the less code that I write manually, the less I care about the individual aspects of the code. I always said, in every single team I took a leadership role in one capacity or another: code is disposable. This has never been more truthful than today. I repeat: code is disposable. What's not disposable is the system knowledge you acquired when developing certain code. This knowledge is what usually translates well from one implementation to another or, for example, what sticks when migrating from v1 to v2 of your API. 

Writing something the second time is easier because you already went through the growing pains of discovering a bunch of stuff and reducing a lot of the ambiguity. You learned what worked well and what didn't. What was overengineered and what was underengineered. This is the important bit in software engineering: gathering knowledge, iterating, evolving. And this is the kind of knowledge that will survive the AI age. Code is just an implementation detail.

Based on this philosophy, this is a non-exhaustive list of the things I care about when code reviewing:

### Architecture and system design
AI models struggle with the big picture and also have the tendency of taking many shortcuts. My review process looks for those signals like hardcoded values and configurations, over-simplification of the problem space (AI often treats coding requests as prototypes or demos) and, paradoxically, over-engineering. AI models also have the annoying trait of assuming production-readiness is equal to complexity. Or in other words, they struggle with balance and pragmatism, stuff that we learn with experience and it's often hard to translate into words.

### Public API and modules
The ergonomics of what we're building matter. The public API needs to "feel right" to the average developer who has to use it. A well-designed interface is intuitive, hard to misuse, and hides the messy internals from the rest of the codebase. I check if the interfaces are robust and correctly scoped, aiming for the smallest possible surface area. If the API is clunky, it doesn't matter how elegant the underlying code is. Is the code easy to use and well documented? A good hint that the public API is good is if the test quality is good. Bad API design is inherently hard to test.

### Algorithms and patterns
LLMs often default to the most naive, brute-force way to solve a problem. It’s common for an agent to try running a massive data migration using nested loops and committing every few rows, when a bulk insert strategy is the correct approach. Or at a core level: using a list when a map or dictionary is the right data structure. Verifying that the data structures and algorithms actually fit the problem space prevents massive performance drops. The goal is code that scales, not just code that passes the tests. Premature optimisation is still a risk, though. If a simpler approach is slightly slower but much easier to read and we are dealing with a small, bounded data set, readability usually wins.

### Dependencies
Agents can easily pull in a massive third-party library for a task that the standard library could solve in three lines of code. Every new package brings outside risk, potential security flaws, and maintenance overhead. Every addition must be actively maintained, secure, and truly necessary. Keeping the app small reduces our attack surface. Core tools like our GenAI SDKs or major web frameworks get a faster pass, but everything else gets scrutinised. A little copying (or reimplementation) is better than a little dependency. The easier it gets to generate and maintain code, the less I worry about re-use, especially if this means adding a new attack vector to my codebase.

### Anti-patterns and quality issues
Just to name a few: god objects, hidden state changes, side effects, global state, resource leaks, ignoring errors, unused functions or variables, and so on. Language idioms also matter; what looks like a trap in one language might be the standard way of doing things in another. While I do care about these a lot, these are also some of the easiest to automate with the use of static analysis (linters) like [golangci-lint](https://golangci-lint.run/) (Go) and [ruff](https://docs.astral.sh/ruff/) (Python).

### Testability
Code that's hard to test is usually designed poorly and will resist change in the future. Clear separation of concerns, clean inputs, and pure functions are ideal. Good tests prove the code works and provide a safety net for our future modifications. For UI components and complex systems, I accept practical testing strategies over strict unit coverage, but the core logic must be covered. I stopped trying to add a coverage target for every single project because each case is a case, but I need to know that whatever can be tested is tested. Ideally 100% of the happy path and a good percentage of the sad path, but I won't try to achieve 100% or anywhere close to that. As long as you have a good observability strategy and good error messages, you are setting yourself up for success as new error modes can be added to the test suite later.

### Benchmarking
For critical paths, we need actual numbers rather than guesses about performance. Clear benchmarks for any changes affecting high-traffic components are required to stop slow code from reaching production. This is only necessary for hot paths; asking for benchmarks on every minor helper function is a waste of everyone's time.

### Lean logging
Logs must be actionable. Unnecessary logs increase our cloud bills and can leak private info. I check log levels to ensure we capture exactly what's needed without including secrets or PII. Verbose logging is fine during development, but it must be cleaned up before merging.

## What I don't care about (mostly)

I let automated tooling handle the details so I can focus on the hard problems. If a machine can do it, a human shouldn't be doing it.

### Formatting
Ever since I started doing Go I have never been into a discussion about formatting styles, but I know they still exist in certain spaces. The best thing you can do is set up a standard and let the linter and formatter handle it. Once the standard is known the code agent can also be more compliant to it. If the CI pipeline passes, it is fine. This speeds up the review and stops pointless style debates.

### Minor syntax and code details
There are many ways to solve a problem, and forcing specific syntax choices limits developer freedom. I don't care if it's a `for` loop or a list comprehension, as long as the logic is sound. 

### Every individual line of code
Reviewing every line generated by an LLM is the job of a compiler or static analyzer. I focus on the logic and connection points instead.

### Debugging
I almost never do debug sessions. If something isn't working, I create a new test to simulate the problem. If after reproducing the problem I can't figure out what's happening, it means that my observability and logs are insufficient, so I focus on improving those. Debugging for me is a last resort, and it's a synonym for adding a bunch of "I AM HERE" print statements, which really should have been log lines. Beware of mutable state and keep your transitional states at least temporarily. This will make your life much easier and dismiss the need for debuggers.

### Unexported names (with caveats)
Internal names have limited scope and rarely impact the overall design. I skim them as long as the public API is solid and the context is clear. This keeps the review moving and avoids nitpicking over minor choices. However, bad names almost always lead to bad design, and always lead to code that is hard to maintain, so I still care a little about them.

### Minor dependencies
The ones that aren't your major framework or client library dependencies. These are less of a concern if they meet our security baseline. Checking for security exploits and problematic licenses is still mandatory, though. If I'm importing something just for one "helper" function I will 100% of the time re-implement that function in my code and get rid of the dependency.

## Conclusions

This isn't a one-size-fits-all protocol, but it's how I am approaching code reviews these days. There is also a lot that can be said about how you are instrumenting your code base. Code reviews alone won't catch all potential problems, and this is why I strongly advocate for automation, now with agentic coding more than never. Modern coding agents have many extension patterns that allow you to constrain the model and get more deterministic outputs: [[Agent Skills]({{< ref "/posts/20260128-agent-skills-gemini-cli/" >}}), hooks, [MCP tools]({{< ref "/posts/20250817-hello-mcp-world/" >}}), policies, rules... It can get a bit confusing since many of these aren't standards, but we're getting to the point where many of these are being standardised.

A car can only run as fast as its brakes support it. Invest in learning the guardrails for your favourite coding agent, and use your precious time to review what can't be automated.

Happy coding!

Dani =^.^=