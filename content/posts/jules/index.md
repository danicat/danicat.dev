+++
date = '2025-05-21T17:45:07+01:00'
draft = true
title = 'Let’s talk about Jules - is this the tool we needed to make vibe coding the new normal?'
featured_image = "jules.png"
+++
Alright everyone, let's talk about Jules! Hot from the ovens of Google I/O, this is what Google is calling an autonomous coding agent… but what is an autonomous coding agent? Think about [NotebookLM](https://notebooklm.google/), but for coding - a specialised AI to help you with coding tasks. The main difference from the traditional “vibe coding” approach is that with Jules you can import your whole project as context for the AI, so all the responses are grounded on the code you are actually working on!

Once the project is imported, you can interact with Jules by submitting “tasks”, which can be anything from bug fixes, dependency updates, new features, planning, documentation, tests and so on. As soon as it receives a task Jules will asynchronously plan its execution in steps and perform different sub-tasks to ensure that the desired outcome is achieved. For example, ensuring that no tests were broken by the new change.

It integrates directly with GitHub, so there is very little friction to start using it. It still won’t replace the IDE completely, but you can perform many tasks directly from Jules up to the point it creates a branch with all the alterations you requested ready to be made into a pull request.

The unfortunate consequence of Jules announcement yesterday is that the tool is currently under heavy load, so it might take a while after you submit a task to see the results, but Jules will do the work in background and if you have browser notifications enabled it will let you know once it is ready.

Given this fact, I wasn’t able to do any major experiments with it, but one of the things I did was to generate the [README for my blog project on Github](https://github.com/danicat/danicat.dev/pull/1) (the source for this very page you are reading now). I also tried some more complex iterations like adjusting the blog template. [It did generate the correct files](https://github.com/danicat/danicat.dev/pull/2), but it was a bit slow to respond to my requests so I had to do a few changes manually.

Not bad for day 1 I would say, and there is a lot of potential to be unlocked over the next few weeks and months. The killer feature is the ability to work on a complete code base instead of that traditional flow of asking Gemini (or ChatGPT) a question, copying the source to the IDE, running, copying and pasting back the results to the LLM and iterating. Of course, tools like Code Assist and CoPilot will provide some of those capabilities without leaving the IDE, but I still feel that the IDE is not the right environment for vibe coding as it feels more like a hack.

In that spirit, maybe Jules is the injection of inspiration we needed for a new era of IDEs that will unlock the potential of AI for developers all over the world in a more natural way. At least that is what I am hoping for!

Jules is currently in public beta and you can play with it today by signing up at htt[ps://jules.google](https://jules.google).
