# Editorial Guidelines

This document outlines the editorial philosophy for `danicat.dev`. Our goal is to publish technical content that is deeply human, rigorously accurate, and genuinely helpful. We aim for the "cozy web" aesthetic—a space that feels like a thoughtful conversation with a competent peer, rather than a corporate broadcast or a dry academic paper.

## Core Philosophy

*   **The "Cozy Web" Vibe:** We value authenticity over polish. The site should feel like a well-tended digital garden, inviting and personal.
*   **Competent Peer Audience:** Assume the reader is smart but lacks specific context. They don't need "Hello World" explained, but they do need clear reasoning for architectural choices or new patterns.
*   **Narrative First:** Every article needs a narrative thread. We avoid purely functional, context-free tutorials. We want to know *why* you built this, what hurt along the way, and what you learned.

## Tone of Voice: Grounded Expressivity

We want expressive, engaging writing, but it must remain grounded in reality.

*   **Honest (Pain and Payoff):** Do not sanitize the process. Highlight cryptic error messages, flawed initial assumptions, and the hours lost to silly mistakes. These struggles often contain the most valuable lessons.
*   **Show, Don't Sell:** Avoid marketing hype. Don't tell us a technology "shines" or is "revolutionary." Show us the specific problem it solved elegantly. Let the reader be impressed by the facts, not your adjectives.
*   **Professional, Not Corporate:** Speak directly to the reader as an equal. Avoid stiff, passive corporate language ("It was determined that..."). Use "I" and "we."
*   **Objective Empowerment:** Avoid subjective ease-of-use judgments (e.g., "simply," "just," "easy"). What is easy for you might be hard for someone else. State the facts and let the reader decide.

## Article Structure (Narrative Models)

Avoid formulaic, cookie-cutter patterns. The structure should serve the story. While standard elements (Intro/Hook, Context, Body, Takeaways, Resources) are usually present, how you arrange them depends on the narrative.

Consider these models as inspiration, not rigid templates:

*   **The Debugging Mystery:** A detective story starting with a baffling error and tracing the clues to a solution.
*   **The Personal Experience Report:** A chronological journey of building or learning something new, focusing on the "aha" moments and the dead ends.
*   **The Deep-Dive Exploration:** A thorough, curiosity-driven examination of a specific technology's internals.
*   **The Event Summary:** A report from a conference or meetup, focusing on atmosphere, key takeaways, and personal connections rather than just listing talks.
*   **The Interview:** A structured conversation distilled into key insights.

## Capitalization Standards

To maintain a consistent, modern, and conversational feel, we adhere to the following capitalization rules:

*   **Main Title (H1):** Use **Title Case**. Capitalize the first word, last word, and all major words.
    *   *Example:* `Beyond the Dev-UI: How to Build an Interface for an ADK Agent`
*   **Section Headers (H2, H3, etc.):** Use **Sentence case**. Capitalize only the first word and any proper nouns. This feels less formal and more aligned with our "cozy web" vibe.
    *   *Example:* `Building the backend with the ADK runtime`
    *   *Example:* `Vibe coding the UI with Gemini`

## Technical Standards

*   **Code with Context:** Code snippets must be accurate and idiomatic. Crucially, explain *why* the code does what it does, not just *what* it does. Comments should focus on non-obvious logic.
*   **Runnable Examples:** Whenever possible, snippets should be copy-paste runnable.
*   **Visuals:** Use Mermaid.js for flowcharts and sequence diagrams, or high-quality screenshots for UI elements. Visuals should clarify complex concepts, not just break up text.
*   **Citations:** Always link to official documentation, specifications, or source code when referenced.

## Style: Watch Words & Lazy Patterns

To maintain our grounded tone, we avoid "lazy" writing constructs—clichés, hype, and filler that dilute the message. These words aren't strictly "forbidden," but their presence is a strong signal to stop and find a more precise, authentic way to express the idea.

### 1. The Hype Train (Avoid Grandiosity)
*Words that try too hard to impress.*
*   **Watch list:** `revolutionary`, `shines`, `perfect`, `unparalleled`, `game-changer`, `fundamentally changed`.
*   **Fix:** Replace with specific evidence. Instead of "It shines at data processing," use "It processed 10GB of data in 3 seconds."

### 2. The Cliché Crutch
*Overused metaphors that have lost their impact.*
*   **Watch list:** `aha! moment`, `missing link`, `silver bullet`, `weaving it all together`, `piece of the puzzle`.
*   **Fix:** Describe the actual realization or connection. What specifically clicked into place?

### 3. Emotional Filler
*Words that tell the reader how to feel instead of showing them why.*
*   **Watch list:** `surprisingly`, `interestingly`, `sadly`, `unfortunately`.
*   **Fix:** Just state the surprising or interesting fact directly.

### 4. The Patronizing Presumption
*Words that assume the reader's skill level.*
*   **Watch list:** `simply`, `just` (as in "just run this"), `obviously`, `easy`, `straightforward`.
*   **Fix:** Remove them. "Run this command" is stronger and kinder than "Just run this command."