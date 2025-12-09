# Editorial Guidelines

This document outlines the editorial philosophy for `danicat.dev`. Our goal is to publish technical content that is deeply human, rigorously accurate, and genuinely helpful. We aim for the "cozy web" aesthetic—a space that feels like a thoughtful conversation with a competent peer, rather than a corporate broadcast or a dry academic paper.

## Core philosophy

*   **Authenticity:** We value authenticity over polish. The site is a personal space, not a content farm. We value human insight over polished marketing copy.
*   **Competent Peer Audience:** Assume the reader is smart but lacks specific context. They don't need "Hello World" explained, but they do need clear reasoning for architectural choices or new patterns.
*   **Narrative First:** Every article needs a narrative thread. We avoid purely functional, context-free tutorials. We want to know *why* you built this, what hurt along the way, and what you learned.

## Tone of voice

We want expressive, engaging writing that remains grounded in reality.

*   **Transparency:** Show the work, including the mistakes. Highlight cryptic error messages, flawed initial assumptions, and the hours lost to silly mistakes. These struggles often contain the most valuable lessons.
*   **Show, Don't Sell:** Avoid marketing hype. Don't tell us a technology "shines" or is "revolutionary." Show us the specific problem it solved elegantly. Focus on the evidence.
*   **Professional, Not Corporate:** Speak directly to the reader as an equal. Avoid stiff, passive corporate language ("It was determined that..."). Use "I" and "we."
*   **Precision:** Avoid subjective judgments (e.g., "simply," "just," "easy"). What is easy for you might be hard for someone else. State the facts and let the reader decide.

## Article structure (Narrative models)

The structure should serve the story.

Consider these models as inspiration, not rigid templates:

*   **The Debugging Mystery:** A detective story starting with a baffling error and tracing the clues to a solution.
*   **The Personal Experience Report:** A chronological journey of building or learning something new, focusing on the breakthroughs and the dead ends.
*   **The Deep-Dive Exploration:** A thorough, curiosity-driven examination of a specific technology's internals.
*   **The Event Summary:** A report from a conference or meetup, focusing on atmosphere, key takeaways, and personal connections rather than just listing talks.
*   **The Interview:** A structured conversation distilled into key insights.

## Taxonomy standards

We use a strict taxonomy to keep content organized and discoverable.

### Categories (Broad themes)
Use **only** these four core categories. Do not create new ones.
*   **`AI & Development`**: For all technical deep dives, coding, GenAI, MCP, etc.
*   **`Workflow & Best Practices`**: For process, tools, and "vibe coding" philosophy.
*   **`Career & Personal`**: For non-technical reflections, milestones, and career advice.
*   **`Meta`**: For posts about the blog itself.

### Tags (Specific topics)
*   **Format:** Always **lowercase** and **kebab-case** (e.g., `vibe-coding`, `google-cloud`).
*   **Language Agnostic:** Use English tags even for translated posts to ensure unified linking across languages.
*   **Content Types as Tags:** Use tags for content types like `tutorial`, `keynote`, or `interview` instead of categories.

## Capitalization standards

To maintain a consistent, modern, and conversational feel, we adhere to the following capitalization rules:

*   **Main Title (H1):** Use **Title Case**. Capitalize the first word, last word, and all major words.
    *   *Example:* `Beyond the Dev-UI: How to Build an Interface for an ADK Agent`
*   **Section Headers (H2, H3, etc.):** Use **Sentence case**. Capitalize only the first word and any proper nouns. This feels less formal and more aligned with our "cozy web" vibe.
    *   *Example:* `Building the backend with the ADK runtime`
    *   *Example:* `Vibe coding the UI with Gemini`

## Spelling standards

*   **British English:** Use British spelling conventions (e.g., `colour`, `behaviour`, `prioritise`, `centre`).
    *   *Exception:* Keep American spelling if it is part of a specific proper noun, command, or technical term (e.g., `Program Files`, `disk`).

## Technical standards

*   **Code with Context:** Code snippets must be accurate and idiomatic. Crucially, explain *why* the code does what it does, not just *what* it does. Comments should focus on non-obvious logic.
*   **Runnable Examples:** Whenever possible, snippets should be copy-paste runnable.
*   **Visuals:** Use Mermaid.js for flowcharts and sequence diagrams, or high-quality screenshots for UI elements. Visuals should clarify complex concepts, not just break up text.
*   **Citations:** Always link to official documentation, specifications, or source code when referenced.
*   **Technical Term Formatting:**
    *   **Backticks (`code style`):** Use only for syntactical elements you would literally type in a terminal or editor (exact command binaries like `ollama run`, file paths, function names, specific package identifiers).
    *   **Plain Text (Capitalized):** Use for the abstract concept, project, product, or community (e.g., "The Osquery community," "We used Ollama").
    *   **First Mention:** Bold or link the first significant mention of a key tool to introduce it.

## Visual standards

We aim for a consistent, high-quality visual identity that blends modern tech aesthetics with a touch of personal flair.

### Hero Images (Featured)
*   **Style:** High-quality **2D Anime**. Avoid generic cartoons or overly gloomy styles unless thematically appropriate.
*   **Aspect Ratio:** 16:9.
*   **Themes:** The Engineer as Hero. Triumph over chaos (bugs, "AI slop", deadlines). Celebration of mastery.
*   **Elements:**
    *   **Magic/Tech:** Glowing code, magical swords (e.g., with the 4-point star Gemini logo), data streams.
    *   **Contrast:** Play with light (order/engineer) vs. dark (chaos/bugs).
    *   **Vibe:** Excitement, vibrancy, high energy. Don't be afraid of anime tropes (power-ups, beam struggles, sparkles).

### In-Article Illustrations (UI/Screenshots)
*   **Style:** Modern, flat, clean digital UI mockups (resembling tools like Linear, Trello, or modern IDEs).
*   **Perspective:** Direct, flat screen capture view. Avoid 3D tilts, desk backgrounds, or clutter.
*   **Content:** Text within the image matters. Use specific copy to illustrate points.
*   **Theme:** Light mode or clean Dark mode. Consistent flat design.

### Diagrams
*   **Style:** Clean, minimalist, professional, flat design.
*   **Background:** White or very light grey.
*   **Palette:** Modern, warm, or pastel colors. Avoid harsh primaries.
*   **Typography:** High contrast, legible, sans-serif.
*   **Layout:** Balanced usage of space.
*   **Rule:** **No titles embedded in the image.** Let the article text or caption handle the title.
*   **Consistency:** If using multiple related diagrams, they **must** share the exact same visual style, dimensions, and font sizes.

## Style guide

**Important Note:** These words are signals to reflect, not automatic deletions. In our context, conversational fillers (like "naturally" or "great!") can enhance warmth and are acceptable. However, words that minimize complexity or patronize the reader must be removed. Distinguish between sounding like a human peer (good) and sounding dismissive of effort (bad).

### 1. Grandiosity
*   **Watch list:** `revolutionary`, `shines`, `perfect`, `unparalleled`, `game-changer`, `fundamentally changed`.
*   **Fix:** Replace with specific evidence. Instead of "It shines at data processing," use "It processed 10GB of data in 3 seconds."

### 2. Overused metaphors
*   **Watch list:** `aha! moment`, `missing link`, `silver bullet`, `weaving it all together`, `piece of the puzzle`.
*   **Fix:** Describe the actual realization or connection. What specifically clicked into place?

### 3. Filler words
*   **Watch list:** `surprisingly`, `interestingly`, `sadly`, `unfortunately`.
*   **Fix:** Just state the surprising or interesting fact directly.

### 4. Minimizing language
*   **Watch list:** `simply`, `just` (as in "just run this"), `obviously`, `easy`, `straightforward`.
*   **Fix:** Remove them. "Run this command" is stronger and kinder than "Just run this command."

## Current taxonomy

| Category | Associated Tags |
| :--- | :--- |
| **AI & Development** | `adk`, `agent`, `ai`, `cloud-assist`, `codelab`, `frontend`, `gemini`, `gemini-cli`, `genai`, `golang`, `keynote`, `mcp`, `python`, `rag`, `terraform`, `tutorial`, `vertex-ai` |
| **Workflow & Best Practices** | `gemini-cli`, `jules`, `opinion`, `tutorial`, `vibe-coding` |
| **Career & Personal** | `community`, `devrel`, `google`, `mental-health`, `mental-models`, `milestone`, `psychology`, `time-perception` |
| **Meta** | `blogging`, `hugo` |


