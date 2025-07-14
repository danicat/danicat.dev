---
title: "Proactiveness considered harmful?! A deep dive into the Gemini CLI system prompt"
date: 2025-07-15
author: "Daniela Petruzalek"
tags: ["gemini-cli", "vibe-coding"]
categories: ["IDEs"]
summary: ""
---
## Introduction

Many of you should be familiar with the [Gemini CLI](https://cloud.google.com/gemini/docs/codeassist/gemini-cli?utm_campaign=CDR_0x72884f69_default_b431747616&utm_medium=external&utm_source=blog>) by now, but in case you are not, check the official [release blog](https://blog.google/technology/developers/introducing-gemini-cli-open-source-ai-agent/) for a quick overview. I've written about how I've included it in my workflow in my previous post ["A Modern Developer Workflow for the AI Enabled World"](../20250714-developer-workflow/), but this time I want to explore something slightly different.

If you have used the CLI for a while you might have noticed that it is very "proactive" as it will often infer the next steps based on even the most ambiguous prompts and immediately jump to perform actions based on these inferences. This level of proactiveness might be harmless for the average case, but in my experience it tends to get in the way of my workflow more often than not.

A typical example would be, after "vibing" some code I ask the CLI to clarify "why did you add this @x file?", for it only to *proactively* assume I have second intentions behind my question and deleting the file without an explanation. I'm usually deeply annoyed with these kind of interactions because for the majority of times I am asking a genuine question.

I've spent a considerable amount of time over the last week trying to figure out how to customise the GEMINI.md file to prevent this and other behaviours with very little success, so I finally decided to go for the nuclear option: what if we could change the system prompt to make it *less* proactive?

## Customising the Gemini CLI with GEMINI.md

A quick note before you decide to go for the nuclear option yourself. For most cases, what you *really* want to do is to customise your GEMINI.md file before messing up with the system instructions. The GEMINI.md file allows you to provide custom instructions similar to the system instructions, but with lower priority.

The good thing about GEMINI.md is that you can have multiple of these files, one for each folder in your project, which given the right content they will help the CLI to understand your project and pinpoint the exact places where it needs to apply changes. To some extent, GEMINI.md is no different than a well done README.md, but with the difference that it is designed to provide instructions for the AI as primary use case, with the side effect of being human readable.

You can also have a top level GEMINI.md that defines the operating modes of the CLI. Here is an example of such a file:

{{< gist ryanjsalva 0a7f6782b8988e760b88f1635ea55f2e >}}

## Gemini CLI system instructions: where are they?

