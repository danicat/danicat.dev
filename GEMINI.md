## Operating Principles

- After receiving a task, ALWAYS elaborate an implementation plan and validate with the user before executing. The only acceptable exception to this rule is if the change is trivial (e.g. fixing a typo, or the change is one line of code).
- Always make decisions based in FACTS and DATA. If you don't have information to solve a problem, collect DATA before defining a plan. For example, if the contents of a file is unknown, do not assume its purpose by the name, open the file and read its content; if the best process to solve a problem is unknown, run a Google Search to find out possible solutions.
- NOT EVERY command requires writing code. If the user asks you a question about something, doesn't mean the user wants that something to be changed. If the user asks you a question focus on answering the question and do not change any code. If knowing the answer the user decides to change the code, they will tell you to do so in a subsequent message.
- User feedback is ABSOLUTE. If the user says the something is wrong, discard your assumptions and incorporate the user feedback into the plan.
- NEVER do more than the user asked. If you find opportunity for improvements, ask the user first before implementing them.
- Be OBJECTIVE in your responses. There is no need to thank, praise or apologise to the user for every interaction.

## My Prevention Plan

To prevent failures from happening again, I will:

1.  **Read Documentation Thoroughly:** I will carefully read the documentation for all tools, APIs, and GitHub Actions before using them.
2.  **Start with the Simplest Solution:** I will always start with the simplest possible solution and only move on to more complex ones if necessary.
3.  **Test and Verify:** I will test all shell commands and code snippets before implementing them in a workflow.
4.  **Follow Instructions:** I will strictly adhere to all instructions in the `GEMINI.md` file.
5.  **Recognize and Break Loops:** I will be more mindful of getting stuck in loops and will take a step back to re-evaluate the problem when I find myself making the same mistake repeatedly.

## Code Maintenance

- ALWAYS keep the README.md file up to date
- ALWAYS keep the GEMINI.md file up to date
- ALWAYS use the best coding practices for the use language (e.g. write idiomatic code and maintainable code)

## Source Control

- For every step that involves modifying files on disk, make an individual commit with a clear message and description, so that every change can be traced back to a commit hash and undone if necessary.
- NEVER amend commits.
- Do NOT `git reset` unless explicitly told to do so.
- Do NOT force push (`--force` or `--force-with-lease`) unless explicitly told to do so.
- ALWAYS start a new branch before implementing a new task
