---
title: "Interaction log: How I used Jules to add a featured post to this blog"
date: 2024-08-02
author: "Daniela Petruzalek"
tags: ["jules", "vibe-coding"]
categories: ["AI Native"]
featured: "/images/jules-interaction-post-placeholder.png"
featurealt: "Jules, an AI coding assistant, represented by a purple octopus logo."
summary: "A detailed account of my iterative process working with Jules, an AI coding assistant, to implement a new featured post section on my blog's homepage. We explore what went well, the challenges, and the importance of clear communication in AI-assisted development."
---

I recently decided to update the homepage of my blog to better highlight the most recent content. Instead of manually coding the changes, I enlisted the help of Jules, an AI coding assistant. This post details our iterative journey, the successes, the (sometimes amusing) misunderstandings, and what I learned about working effectively with AI for web development.

## The Goal: A Featured Post Section

My initial request was straightforward:
> "Change the layout of the main page so that it displays the most recent blog post in highlight instead of it being in the recent posts list. The recent posts should contain all other posts except the most recent one. This behaviour should be seen only on the blog landing page (home). If the user clicks on the Blog menu it should still see all the posts in reverse chronological order, including the most recent one."

Jules quickly understood and proposed a plan involving exploring the Hugo codebase, identifying templates, and modifying them.

## Iteration Highlights: The Good, The Bad, and The AI

Our collaboration involved several iterations to get things just right.

### Iteration 1: Initial Setup - Getting the Basics Right
Jules correctly identified the Blowfish theme's partials and set up the override structure. The logic to separate the latest post from the others in the "Recent Posts" list was implemented well.

*   **What worked:** Understanding the core Hugo structure, fetching posts, basic template modifications.

### Iteration 2: Styling the Featured Post - Width and Title
We then focused on the appearance. I requested a specific title "Featured Post" and wanted to adjust its width relative to other elements.

*   **Jules's approach:** Modified i18n files for the title and used Tailwind's percentage width classes (`md:w-4/5`, then `md:w-2/3`).
*   **Challenge:** Getting the visual perception of "80% width" right using Tailwind's responsive classes required a couple of tries. What seems like 80% in code doesn't always translate directly visually without seeing it live.

### Iteration 3: Image Dimensions - Too Tall, Too Narrow
The next challenge was the featured image. I asked for it to be taller. Jules's first attempt made it *too* tall. Then, adjusting the card width made it *too* narrow.

*   **Jules's approach:** Manipulated `padding-bottom` for aspect ratio, and `max-w-*` classes for container width.
*   **Learning:** Fine-tuning visual aesthetics like aspect ratios and container widths is highly iterative with an AI. Clear, descriptive feedback is key. We eventually settled on `max-w-2xl` for the container and aimed for a 4:3 image.

### Iteration 4: Custom CSS vs. Tailwind - A Brief Detour
At one point, to get very specific control, I asked Jules to use a custom CSS class.
> "jules, instead of trying to use an existing style class, create an unique style class for the featured post card. This style should use relative width and height of 75% of the container..."

*   **Jules's response:** Jules created the custom CSS and refactored the partial.
*   **Outcome:** I quickly realized that sticking to Tailwind best practices was preferable for this project's consistency.
*   **What worked:** Jules successfully implemented the custom CSS as requested.
*   **Learning:** It's important for me (the user) to be clear about preferred methodologies (like sticking to a framework like Tailwind) if I have them. Jules adapted back when I asked to revert.

### Iteration 5: The Great "Comments" Misunderstanding!
This was perhaps the most illustrative part of the AI-human interaction. I mentioned:
> "the comments are rendering in the featured post. please remove all the comments or make them invisible"

*   **Jules's interpretation:** Jules, quite reasonably, assumed I meant the blog's *comment system* (like Utterances or Giscus) or metadata like views/likes counts. This led to a series of steps where Jules tried to:
    1.  Investigate which partials might be rendering a comment system.
    2.  Propose hiding views/likes counts, hypothesizing they might be triggering comments via JavaScript.
    3.  Implemented a system with an `IsFeaturedCard` flag to conditionally hide these meta items.
*   **My Clarification:** After these changes, I clarified:
    > "you are wrong, I never said I wanted to remove the views and likes - I'm referring to the code comments in rendering as {/* Adjusted padding ... */} and {/* Removed prose classes ... */}"
*   **Resolution:** Once Jules understood I meant *literal Go template/HTML comments* that were incorrectly formatted and rendering as text, the fix was immediate: remove the offending `{/* ... */}` text from the templates.
*   **What worked:** Jules's persistence and systematic approach to debugging the (misunderstood) problem.
*   **Challenge & Learning:** This highlighted a crucial aspect of AI interaction: ambiguity in natural language. "Comments" has a very specific meaning in web development (user comments) and a different one for code comments. My initial report wasn't precise enough. For Jules, it showed the difficulty in inferring developer shorthand or spotting typos that change meaning.

### Iteration 6: Final Polish
After resolving the comment visibility, we made final tweaks:
*   Removing the "Featured Post" heading above the card.
*   Adjusting the card width to 50% (`w-1/2`) of its parent.
*   Slightly increasing font sizes for the title and summary within the featured card.
*   Setting the image aspect ratio to 16:9.

## What Worked Well with Jules

*   **Speed of Implementation:** Jules can make code changes, create files, and refactor structures much faster than I could manually type.
*   **Handling Complex Instructions:** For the most part, Jules understood multi-step requests and complex layout goals.
*   **Systematic Problem Solving:** Even when misunderstanding the "comments" issue, Jules went through a logical process of elimination and proposed solutions.
*   **Adherence to Plan:** Jules would propose a plan and stick to it, which is great for structured development.
*   **Iterative Refinement:** Jules was receptive to continuous feedback and tweaks.

## Challenges and Learnings

*   **Precision of Language:** The "comments" incident underscores how critical precise language is. What's obvious to a human might not be to an AI that takes things more literally or defaults to common interpretations.
*   **Visual Feedback Loop:** Without Jules "seeing" the output, describing visual discrepancies or desired aesthetics purely through text requires patience and clear, descriptive language from my end.
*   **Over-Correction/Misinterpretation:** Sometimes, in an attempt to solve a problem, the AI might go down a path that's technically correct for its interpretation but not what the user intended (like the extensive work to hide views/likes).
*   **Knowing When to Specify the "How":** While I often want Jules to figure out the "how," sometimes specifying the preferred method (e.g., "use Tailwind classes" vs. "write custom CSS") is more efficient.

## Conclusion

Working with Jules on this homepage feature was a fascinating and ultimately successful experience. It truly felt like "vibe-coding" – a rapid exchange of ideas and implementations. The key to success lies in clear, iterative communication, patience during misunderstandings, and a willingness to guide the AI with specific feedback. While not a replacement for human understanding and design sense, AI assistants like Jules are powerful tools that can significantly accelerate the development process, especially for well-defined tasks and iterative refinements.

---
