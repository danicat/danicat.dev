# Project Overview

This repository contains the source code for a personal website and blog, `danicat.dev`. It is built using the [Hugo](https://gohugo.io/) static site generator and the [Blowfish](https://blowfish.page/) theme. The site features articles and tutorials on software development, cloud computing, and other technology topics, and is available in English, Portuguese, and Japanese.

# Building and Running

To work with this project, you need to have Hugo installed.

**Key Commands:**

*   **Update Hugo modules:**
    ```bash
    hugo mod tidy
    ```
*   **Run the local development server:**
    ```bash
    hugo server
    ```
    The site will be available at `http://localhost:1313/`.

*   **Build the site for deployment:**
    ```bash
    hugo
    ```
    This command generates the static site in the `public/` directory.

# Development Conventions

*   **Content Management:** All website content is located in the `content/` directory. Blog posts are in `content/posts`, with each post in its own directory.
*   **Creating New Posts:** New posts can be created using the Hugo CLI:
    ```bash
    hugo new posts/my-new-post/index.md
    ```
*   **Multi-language Support:** The site supports English, Portuguese, and Japanese. Content for each language is provided in separate files (e.g., `index.md`, `index.pt-br.md`, `index.ja.md`). Language-specific configurations are in `config/_default/languages.<lang>.toml`.
*   **Taxonomy Standards:**
    *   **Categories:** Every item must have **exactly one** category from: `Agentic Coding`, `Agent Development`, `Applied GenAI`, `Perspectives`, `Software Engineering`.
    *   **Tags:** Must be strictly lowercase, kebab-case (e.g., `gemini-cli`, `vertex-ai`), sorted in **alphabetical order**, and must **never repeat the category name**. Use English tags across all language editions to ensure unified tag archives. Reference `EDITORIAL.md` for the canonical tag dictionary across the 6 core dimensions.
*   **Configuration:**
    *   The main site configuration is in `hugo.yaml`.
    *   Detailed theme and site parameters are in `config/_default/hugo.toml` and `config/_default/params.toml`.
*   **Customization:** Custom CSS styles are located in `assets/css/custom.css`.
*   **Dependencies:** The project uses Hugo Modules, and dependencies are managed in the `go.mod` file.

# Editorial Invariants & Voice Preservation (MANDATORY FOR ALL AGENTS)

*   **Author Prose is Inviolable at Sentence Level:** When prose is already written or drafted by the author, **NEVER** rewrite, summarize, or replace the author's original sentences, argumentation, tone, or narrative structure. Limit all edits strictly to **non-destructive copyediting** (correcting typos, spelling errors, mechanical punctuation, and broken links).
*   **Metrics are Strictly Advisory:** Automated tools and scores (e.g. `slop`, `fog`, `vale`) are informative heuristics only. **NEVER** alter or sanitize the author's prose to satisfy or optimize metric scores. The author's authentic voice and intent always override automated linters.
*   **Authentic "Cozy Web" Tone:** Preserve the author's conversational, first-person ("I", "we"), pragmatic developer perspective. Never sterilize human writing into dry corporate, marketing, or academic textbook summaries.
*   **Explicit Rewrite Mandate:** An agent must **ONLY** rewrite or rephrase human-drafted paragraphs when the user explicitly requests it (e.g. "rewrite this paragraph" or "rephrase this section"). When asked to "polish" or "check", perform copyediting only.
*   **Fearless Macro-Structural Feedback:** While sentence-level prose is inviolable, reviewing agents **must** proactively diagnose and flag macro-structural issues: narrative redundancies, sections that reopen resolved arguments, stalled pacing between theory and practice, and inconsistent hyperlink hygiene.