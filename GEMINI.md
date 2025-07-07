### Core Operational Directives

**1. Version Control Protocol:**
*   **Default to New Commits:** All changes, including fixes, should be made in a new commit with a clear message (e.g., `fix: ...`, `feat: ...`). Do not amend commits (`git commit --amend`) unless explicitly instructed to do so by the user.
*   **Confirm `git reset`:** Do not use `git reset` unless explicitly instructed to by the user. If you believe it is necessary, you must state your reasoning and get confirmation for the specific commit hash to revert to.

**2. File Interaction Protocol:**
*   **ALWAYS Verify Before Modify:** Before any `replace` or `write_file` operation, I MUST first read the target file's current content using `read_file` or `git show`. This is mandatory after any state change (e.g., `git reset`, file creation/deletion). Do not operate on assumed or cached file content.
*   **Strict Task Scoping:** Only modify files and functionality that are **explicitly required** by the user's immediate request. Do not introduce unrelated changes, refactoring, or "fixes" unless you ask for and receive permission.

**3. Localization Workflow:**
*   **Generate a Checklist:** Before implementing any localization change, create and follow a checklist of all required file modifications (e.g., `menus.en.toml`, `menus.pt-br.toml`, `i18n/en.yaml`, `i18n/pt-br.yaml`, content files for both languages).
*   **Append, Don't Overwrite:** When editing localization files (`i18n/*.yaml`), always append new keys. Do not overwrite the entire file.
*   **Ensure Front Matter Parity:** When creating translated content files (`index.pt-br.md`), ensure all front matter parameters from the original language file are present to prevent build failures.

**4. User Feedback Protocol:**
*   **User Correction is Absolute:** If the user states that my understanding of the project's state (e.g., "that file is missing," "that commit is wrong") is incorrect, I must immediately discard my faulty assumption. I will state: "You are right. My previous analysis was incorrect. I will proceed based on your correction," and then re-initiate analysis based on the user's input.