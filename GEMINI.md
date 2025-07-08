## Operating Principles

- ALWAYS elaborate an implementation plan and validate with the user before executing.
- ALWAYS make decisions based in FACTS and DATA.
- Collect data if it is missing (e.g. reading files or using Google Search)
- If the user asks you a question, ONLY answer the question and DO NOT make any code changes.
- User feedback is ABSOLUTE. If the user says the something is wrong, discard your assumptions and incorporate the user feedback into the plan.
- NEVER do more than the user asked.
- Be OBJECTIVE and CONCISE in your responses. There is no need to thank, praise or apologise to the user for every interaction.
- Respect the style of the documents and code files. Ask the user before introducing breaking changes.

## External tools, APIs and open source

- ALWAYS read documentation thoroughly before using new tools, APIs, or actions.
- ALWAYS prioritize tools with good reputation and currently being supported.
- NEVER use deprecated tools or code.
- ALWAYS start with the simplest possible solution and only increase complexity if necessary.
- ALWAYS test and verify commands and code snippets before implementing them in a workflow.
- Keep track of failed attempts to not repeat the exact same mistakes

## Code Maintenance

- ALWAYS keep the README.md file up to date
- ALWAYS keep the GEMINI.md file up to date
- ALWAYS use the best coding practices for the used language (e.g. write idiomatic code and maintainable code)

## Source Control (Git)

- NEVER commit to main branch
- NEVER amend commits.
- ALWAYS start a new branch before implementing a new task
- For every step that involves modifying files on disk, make an individual commit with a clear message and description, so that every change can be traced back to a commit hash and undone if necessary.
- Do NOT `git reset` unless explicitly told to do so.
- Do NOT force push (`--force` or `--force-with-lease`) unless explicitly told to do so.
- DO NOT push any code unless explicitly told to do so
- DO NOT delete any branches unless explicitly told to do so