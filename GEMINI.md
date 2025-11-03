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
    *   **Categories:** Use only: `AI & Development`, `Workflow & Best Practices`, `Career & Personal`, `Meta`.
    *   **Tags:** Must be lowercase, kebab-case (e.g., `google-cloud`, `vibe-coding`). Use English tags for all languages to ensure unified linking.
*   **Configuration:**
    *   The main site configuration is in `hugo.yaml`.
    *   Detailed theme and site parameters are in `config/_default/hugo.toml` and `config/_default/params.toml`.
*   **Customization:** Custom CSS styles are located in `assets/css/custom.css`.
*   **Dependencies:** The project uses Hugo Modules, and dependencies are managed in the `go.mod` file.