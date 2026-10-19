var scriptBundle = document.getElementById("script-bundle");
var copyText = scriptBundle?.getAttribute("data-copy") || "Copy";
var copiedText = scriptBundle?.getAttribute("data-copied") || "Copied";

function createCopyButton(highlightWrapper) {
  const button = document.createElement("button");
  button.className = "copy-button";
  button.type = "button";
  button.ariaLabel = copyText;
  button.innerText = copyText;
  button.addEventListener("click", () => copyCodeToClipboard(button, highlightWrapper));
  highlightWrapper.insertBefore(button, highlightWrapper.firstChild);
}

async function copyCodeToClipboard(button, highlightWrapper) {
  const codeToCopy = getCodeText(highlightWrapper);
  try {
    await navigator.clipboard.writeText(codeToCopy);
    button.blur();
    button.innerText = copiedText;
    setTimeout(function () {
      button.innerText = copyText;
    }, 2000);
  } catch (err) {
    console.error("Failed to copy code block to clipboard:", err);
    button.innerText = "Error";
    throw err;
  }
}

function getCodeText(highlightWrapper) {
  const highlightDiv = highlightWrapper.querySelector(".highlight");
  if (!highlightDiv) return "";

  const codeBlock = highlightDiv.querySelector("code");
  const inlineLines = codeBlock?.querySelectorAll(".cl");
  const tableCodeCell = highlightDiv?.querySelector(".lntable .lntd:last-child code");

  if (!codeBlock) return "";

  if (inlineLines.length > 0) {
    const cleanedLines = Array.from(inlineLines).map((line) => line.textContent.replace(/\n$/, ""));
    return cleanedLines.join("\n");
  }

  if (tableCodeCell) {
    return tableCodeCell.textContent.trim();
  }

  return codeBlock.textContent.trim();
}

async function copyPageMarkdown(button) {
  const labelEl = button.querySelector(".copy-md-label");
  const copiedLabel = button.getAttribute("data-copied-label") || "Copied!";
  const defaultLabel = button.getAttribute("data-default-label") || "Copy Markdown";

  try {
    let markdown = button.dataset.cachedMarkdown;
    if (!markdown) {
      const mdUrl = window.location.pathname.endsWith("/")
        ? window.location.pathname + "index.md"
        : window.location.pathname + "/index.md";
      const response = await fetch(mdUrl);
      if (!response.ok) {
        throw new Error(`Failed to fetch ${mdUrl}: HTTP ${response.status} ${response.statusText}`);
      }
      markdown = await response.text();
      button.dataset.cachedMarkdown = markdown;
    }

    await navigator.clipboard.writeText(markdown);

    if (labelEl) {
      labelEl.textContent = "✓ " + copiedLabel;
      setTimeout(function () {
        labelEl.textContent = defaultLabel;
      }, 1800);
    }
  } catch (err) {
    console.error("Failed to fetch or copy page Markdown:", err);
    button.classList.add("border-red-500", "text-red-600", "dark:text-red-400");
    if (labelEl) {
      labelEl.textContent = `✗ Error (${err.message})`;
    }
    throw err;
  }
}

window.copyPageMarkdown = copyPageMarkdown;

window.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".highlight-wrapper").forEach((highlightWrapper) => createCopyButton(highlightWrapper));
  document.querySelectorAll("#copy-markdown-button, [data-copy-markdown]").forEach((button) => {
    button.addEventListener("click", () => copyPageMarkdown(button));
  });
});
