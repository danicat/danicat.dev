# Project Overview

This repository contains the source code for `danicat.dev`, a personal website and blog built with [Hugo](https://gohugo.io/) and the [Blowfish](https://blowfish.page/) theme, published in English, Portuguese, and Japanese.

# Building and Running

*   **Update Hugo modules:** `hugo mod tidy`
*   **Run local dev server:** `hugo server` (available at `http://localhost:1313/`)
*   **Build for production:**
    ```bash
    hugo --baseURL "https://danicat.dev/" --minify
    ```
    Always pass `--baseURL "https://danicat.dev/"` for production builds so `sitemap.xml`, canonical headers, and multilingual `hreflang` alternates emit valid absolute URLs.

# Content, Multilingual & Taxonomy Standards

*   **Content Layout:** Posts live under `content/posts/<slug>/` with one file per language (`index.md`, `index.pt-br.md`, `index.ja.md`). Language configs live in `config/_default/languages.<lang>.toml`; site settings live in `hugo.yaml`, `config/_default/hugo.toml`, and `config/_default/params.toml`; custom CSS lives in `assets/css/custom.css`.
*   **Categories:** Exactly one category per item: `Agentic Coding`, `Agent Development`, `Applied GenAI`, `Perspectives`, `Software Engineering`.
*   **Tags:** Lowercase, kebab-case, sorted alphabetically, English across all language editions, and never repeating the category name. Canonical dictionary:
    *   *Languages & Runtimes:* `golang`, `python`
    *   *Models & Infrastructure:* `gemini`, `vertex-ai`, `genkit`, `google-cloud`, `nano-banana`
    *   *Agent Tooling & Protocols:* `antigravity`, `gemini-cli`, `adk`, `mcp`, `a2ui`, `osquery`, `jules`
    *   *Methodologies & Patterns:* `vibe-coding`, `agent-skills`, `subagents`, `rag`, `code-review`, `agile`, `game-dev`
    *   *Core Engineering:* `testing`, `mutation-testing`
    *   *Content Formats:* `keynote`, `workshop`, `codelab`, `tutorial`, `train-the-trainer`
    *   *Perspectives & Culture:* `career`, `art`, `reflection`
*   **Series:** Multi-part posts use `series: ["<Series Title>"]` and `series_order: <int>`, kept identical across all language editions.
*   **Schema.org Structured Data:** Handled by `layouts/partials/schema.html` (`TechArticle`, `ProfilePage`, `EducationEvent`, `BreadcrumbList`). Supports optional frontmatter fields `dependencies`, `proficiencyLevel`, `recording`/`video`, and `software`/`software_application`.
*   **Multilingual Linking:** Internal links in `index.pt-br.md` and `index.ja.md` must point to the localized target edition with localized anchor text. Never force links between unrelated topics.

# Analytics & Telemetry Events

*   GA4 (`gtag.js`) loads lazily after first user interaction or a 3.5s idle timeout (`layouts/partials/analytics/ga.html`).
*   `assets/js/page.js` dispatches flat scroll-depth events (`scroll_25`, `scroll_50`, `scroll_75`, `scroll_100`); `assets/js/code.js` dispatches `code_copy` and `copy_markdown`. Keep custom event names flat so `google_analytics.py sync` ingests them directly into `daily_events`.

# Blog Editorial, Visual & Sample Code Standards

*   **Standalone Cozy-Web Voice:** Every article must stand on its own (never assume prior posts were read). Preserve the author's first-person ("I", "we"), pragmatic developer voice. Never rewrite author prose above `<!-- human writing stopped here - do not remove this marker -->`, and continue directly from the author's closing segue above it.
*   **Headings, Spelling & Product Naming:** Use Title Case for H1 and Sentence case for H2+. Use British English (`colour`, `behaviour`, `prioritise`), keeping American spelling in proper nouns, code, and commands. Use current product names in prose (`Gemini Enterprise`, or `Gemini Enterprise (formerly known as Vertex AI)` on first mention when helpful) rather than deprecated `Vertex AI`.
*   **Hero Images (16:9):** Invent a fresh scene concept for every post; never clone a previous post's composition. Use grounded 2D editorial comic/illustration scenes in natural daylight. Never use AI-slop style adjectives (`luminous`, `vibrant`, `dynamic`, `radiant`, `neon`, `glowing`, `sparkles`, `magical swords`, `high energy`). Visually inspect every generated image before showing it and regenerate if text is garbled/duplicated or physical logic breaks (for example, drinking through a closed visor).
*   **Architecture Diagrams & Screenshots:** Generate system architecture diagrams as clean, flat Google Cloud diagrams using Nano Banana 2.1 on a light background with no embedded title, matching existing series styling (do not use Mermaid for architecture diagrams). Keep UI screenshots cleanly sized with attribution in the figure caption rather than an extra paragraph.
*   **Social Posts:** Attach only one full-size hero image (never pair hero + diagram), link only to the blog post (not the source repo), avoid bullet-list overload or formulaic "three principles" / "two groups" templates, and never dare or invite readers to break live agents.
*   **Didactic Companion Code & Layout Changes:** Keep blog sample code simple and aligned with official guides rather than overengineering. Never label unreleased prototypes `v1` or `v2`. Never take single-turn shortcuts on security middleware (screen user prompts, tool outputs, and model responses on every turn) and verify framework hooks with end-to-end tests against real runtime behavior rather than synthetic mocks. When changing shared CSS or layouts, visually verify regular blog article pages too.

# Public Repository Security & Git Hygiene (MANDATORY)

`danicat/danicat.dev` is a **PUBLIC** GitHub repository. Everything committed to Git (including `draft: true` files or files deleted later) remains permanently visible in public history.

*   **Never Commit Unapproved Drafts:** Keep WIP drafts in gitignored paths (`DRAFT.md` or `content/drafts/`). Never stage or commit `draft: true` files or new directories under `content/posts/` or `content/codelabs/` without explicit publication approval.
*   **Forbidden Staging Commands:** Never run `git add .`, `git add -A`, or `git commit -a`. Stage explicit file paths and check `git status` and `git diff --staged` before committing.
*   **Pre-Commit Sanitization:** Never commit internal hostnames (`*.roam.internal`, `*.corp.google.com`, `petruzalek-mac`), local user home paths, unannounced/embargoed product identifiers, secrets/tokens/databases, Python virtualenvs/bytecode, or `.gemini/` / `.agents/` directories.
*   **Never Bypass Hooks:** Never pass `--no-verify` (`-n`). Fix any pre-commit hook failure at the source.
