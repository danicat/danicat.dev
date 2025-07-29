# Blog Editorial and Interview Guide

This document contains the editorial guidelines for writing articles for this blog and the process for interviewing the author to extract the necessary details.

## Editorial Guidelines

### Core Philosophy
Every article is a personal story about a technical journey. It's not just a tutorial; it's a narrative that shares the "why" and the "how," including the struggles, the "aha!" moments, and the hard-won lessons. The goal is to be cozy, helpful, and relatable.

### Tone of Voice
- **Personal and Narrative:** Start with a personal story or a relatable frustration. Connect with the reader on a human level.
- **Honest About the Struggle:** Do not present a sanitized, perfect process. Highlight the "pain and payoff." Talk about the cryptic error messages, the flawed initial prompts, and the hours of trial-and-error. These struggles contain the most valuable lessons.
- **Professional, Not Overly Casual:** The tone is that of an experienced peer sharing knowledge. Avoid overly simplistic or patronizing language. Use humor and personal touches (like the `=^_^=` emoticon) sparingly and effectively.
- **Empower the Reader:** Present information objectively and avoid subjective judgments (e.g., calling a protocol "simple"). Allow the reader to form their own opinions based on the facts and the story.

### Article Structure
A typical article should follow this narrative flow:
1.  **Introduction:** Hook the reader with a personal story about a problem or frustration. Set the stage for the journey.
2.  **Context-Setting:** If the topic is complex (like MCP), provide a clear, concise explanation with helpful analogies and links to official documentation.
3.  **The Journey (Body):** Walk through the process chronologically. Each section should represent a phase of the journey, complete with the prompts used, the results (good and bad), and the lessons learned.
4.  **Key Takeaways:** Conclude with a summary of the most important, high-level lessons learned from the entire experience.
5.  **What's Next?:** A brief, forward-looking section that discusses the future of the project and provides links to related official or community efforts.
6.  **Resources and Links:** A final, comprehensive list of all URLs mentioned in the article.

### Titles and Headings
- **Main Title:** Should be a compelling hook. It can be a conversational question, a playful declaration, or a pop-culture reference, but it must be professional.
- **Headings:** Use headings as narrative signposts to guide the reader through the story. Use clever or funny headings very sparingly (1-2 per article, maximum) to emphasize key, surprising moments. The rest should be grounded, descriptive, and professional.

### Technical Accuracy
- **Precision is Paramount:** All technical details, especially protocol messages and code snippets, must be 100% accurate.
- **Cite Your Sources:** Always link to the official documentation, specifications, and SDKs you reference.
- **Use Real-World Examples:** Whenever possible, use the *actual* output from tools and commands for authenticity. If a diagram is used, credit the source in a caption.

---

## The Interview Process

When the author has an idea for an article, the following process should be used to flesh out the details and create a draft that aligns with the editorial guidelines.

**1. Establish the Core Idea:**
   - Begin by understanding the author's high-level goal for the article.

**2. Create a Baseline Draft:**
   - Based on the core idea and an analysis of the subject matter, create an initial, high-level draft. This draft should follow the standard article structure.
   - Pepper the draft with specific, targeted questions (marked with `***`) to identify the gaps in the narrative.

**3. Conduct the Interview (One Question at a Time):**
   - Present one question at a time to the author.
   - Listen carefully to the answers, paying close attention to details about struggles, frustrations, and "aha!" moments.

**4. Focus on the "Pain and Payoff":**
   - The most important details are often in the struggle. Ask follow-up questions to uncover:
     - What was the initial, less-successful prompt?
     - What were the specific, "not great" results? (e.g., "The AI was stuck in a loop calling Google Search...")
     - What was the specific, cryptic error message that caused a roadblock?
     - What was the key piece of information that finally solved the problem?

**5. Iteratively Refine and Integrate:**
   - After each answer, rewrite the relevant section of the article to weave the author's story and technical details into the narrative.
   - Present the updated section to the author for review to ensure it captures their voice and experience accurately.
   - Repeat this process until all questions are answered and all sections are refined.

**6. Final Review:**
   - Once the content is complete, perform a final review of the entire article with the author to ensure it meets all editorial guidelines and is ready for publication.

---

## Localization Guidelines

When translating articles from English (en) to other languages (e.g., pt-br, ja), the following rules must be strictly followed to maintain consistency and clarity.

**1. Do Not Translate Technical Terms:**
   - All technical computer science and software engineering terms must remain in English. This is not an exhaustive list; use your best judgment for similar jargon.
   - Examples: `API`, `backend`, `CLI`, `commit`, `database`, `frontend`, `JSON`, `LLM`, `prompt`.

**2. Do Not Translate Product & Brand Names:**
   - All product, company, and brand names must remain in their original form.
   - Examples: `Claude`, `Gemini CLI`, `Go`, `GoDoctor`, `Google Cloud`, `Jules`, `osquery`.

**3. Maintain Formatting:**
   - Preserve all markdown formatting, including headings, lists, bold/italic text, and links.
   - Do not translate content within code blocks (```). Comments within code may be translated.
   - Keep all URLs and links unchanged.

**4. File Naming and Structure:**
   - The translated file must be saved as `index.<lang-code>.md` within the same directory as the original `index.md`.
   - Example: `index.pt-br.md`.

**5. Add Translation Notice:**
   - Add the `{{< translation-notice >}}` shortcode to the top of every translated page, immediately after the front matter.

**6. Tone and Style:**
   - Review existing articles in the target language to match the established professional yet approachable tone.
