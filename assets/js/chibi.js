(function () {
  "use strict";

  const SESSION_KEY = "chibi_dani_session_id";

  function sendTelemetry(name, params) {
    if (!navigator.webdriver && typeof window.gtag === "function") {
      if (params && typeof params === "object") {
        window.gtag("event", name, params);
      } else {
        window.gtag("event", name);
      }
    }
  }

  function escapeHtml(str) {
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function sanitizeUrl(rawUrl) {
    const trimmed = String(rawUrl || "").trim();
    if (!trimmed) return "#";
    if (trimmed.startsWith("/") && !trimmed.startsWith("//")) return trimmed;
    try {
      const parsed = new URL(trimmed, window.location.origin);
      if (parsed.protocol === "https:" || parsed.protocol === "http:" || parsed.protocol === "mailto:") {
        return parsed.href;
      }
    } catch {
      return "#";
    }
    return "#";
  }

  function renderInline(escaped) {
    return escaped
      .replace(/\*\*([^*\n]+)\*\*/g, "<strong>$1</strong>")
      .replace(/(^|[^\*])\*([^*\n]+)\*(?!\*)/g, "$1<em>$2</em>")
      .replace(/\[([^\]\n]+)\]\(((?:[^()\s\n]|\([^()\s\n]*\))+)\)/g, function (_m, label, url) {
        const href = sanitizeUrl(url);
        if (href === "#") return label;
        const isExternal = !href.startsWith(window.location.origin + "/") && !href.startsWith("/");
        const attrs = isExternal ? ' target="_blank" rel="noopener"' : "";
        return '<a href="' + escapeHtml(href) + '"' + attrs + ' class="chibi-link">' + label + "</a>";
      });
  }

  function renderMarkdown(rawText) {
    if (!rawText) return "";
    let text = String(rawText).replace(/\r\n/g, "\n").replace(/\u0000/g, "");

    const blocks = [];
    text = text.replace(/```[a-zA-Z0-9_-]*\n?([\s\S]*?)```/g, function (_m, code) {
      const idx = blocks.length;
      blocks.push("<pre tabindex=\"0\"><code>" + escapeHtml(code.replace(/\n$/, "")) + "</code></pre>");
      return "\n\u0000B" + idx + "\u0000\n";
    });

    const codes = [];
    text = text.replace(/`([^`\n]+)`/g, function (_m, code) {
      const idx = codes.length;
      codes.push('<code class="chibi-inline-code">' + escapeHtml(code) + "</code>");
      return "\u0000C" + idx + "\u0000";
    });

    text = escapeHtml(text);
    const out = [];
    let para = [];
    let listTag = "";

    function flushPara() {
      if (para.length > 0) {
        out.push("<p>" + renderInline(para.join("<br>")) + "</p>");
        para = [];
      }
    }

    function closeList() {
      if (listTag) {
        out.push("</" + listTag + ">");
        listTag = "";
      }
    }

    for (const rawLine of text.split("\n")) {
      const line = rawLine.trim();
      if (!line || /^\u0000B\d+\u0000$/.test(line)) {
        flushPara();
        closeList();
        if (line) out.push(line);
        continue;
      }

      const hMatch = line.match(/^#{1,6}\s+(.+)$/);
      if (hMatch) {
        flushPara();
        closeList();
        out.push("<p><strong>" + renderInline(hMatch[1]) + "</strong></p>");
        continue;
      }

      const ulMatch = line.match(/^[-*+]\s+(.+)$/);
      const olMatch = !ulMatch && line.match(/^\d+\.\s+(.+)$/);
      if (ulMatch || olMatch) {
        flushPara();
        const wantTag = ulMatch ? "ul" : "ol";
        if (listTag !== wantTag) {
          closeList();
          listTag = wantTag;
          out.push("<" + listTag + ">");
        }
        out.push("<li>" + renderInline((ulMatch || olMatch)[1]) + "</li>");
        continue;
      }

      closeList();
      para.push(line);
    }

    flushPara();
    closeList();

    return out
      .join("\n")
      .replace(/\u0000C(\d+)\u0000/g, (_m, i) => codes[Number(i)] || "")
      .replace(/\u0000B(\d+)\u0000/g, (_m, i) => blocks[Number(i)] || "");
  }

  function appendStreamingCursor(html) {
    const cursor = '<span class="chibi-cursor" aria-hidden="true"></span>';
    if (!html) return cursor;
    return html.endsWith("</p>") ? html.slice(0, -4) + cursor + "</p>" : html + cursor;
  }

  const GENERIC_STATUS_MESSAGES = new Set([
    "cancelled",
    "unknown",
    "invalid argument",
    "deadline exceeded",
    "not found",
    "already exists",
    "permission denied",
    "unauthenticated",
    "resource exhausted",
    "failed precondition",
    "aborted",
    "out of range",
    "unimplemented",
    "internal",
    "unavailable",
    "data loss",
  ]);

  function parseSSEEvent(payload) {
    if (!payload || payload === "[DONE]") return { text: "", hasTools: false, sessionId: "" };
    if (!payload.startsWith("{")) return { text: payload, hasTools: false, sessionId: "" };

    const parsed = JSON.parse(payload);
    if (parsed.error) {
      const errMsg = typeof parsed.error === "string" ? parsed.error : parsed.error.message || "";
      const err = new Error(errMsg || "Request error");
      err.publicServerMessage = Boolean(errMsg);
      throw err;
    }

    const chunk = parsed.message && parsed.message.modelChunk;
    if (chunk && Array.isArray(chunk.content)) {
      const role = chunk.role || "model";
      let text = "";
      let hasTools = false;
      for (const part of chunk.content) {
        if (part && part.toolRequest) hasTools = true;
        if (role === "model" && part && typeof part.text === "string") text += part.text;
      }
      return { text, hasTools, sessionId: "" };
    }

    if (parsed.result) {
      const errMsg = typeof parsed.result.error === "string"
        ? parsed.result.error
        : parsed.result.error?.message || "";
      let resultText = "";
      for (const part of parsed.result.message?.content || []) {
        if (part && typeof part.text === "string") resultText += part.text;
      }
      if (parsed.result.error || parsed.result.finishReason === "failed" || (parsed.result.finishReason === "aborted" && !resultText)) {
        const err = new Error(errMsg || "Blocked by guardrail");
        err.publicServerMessage = Boolean(errMsg);
        throw err;
      }
      return { text: "", resultText, hasTools: false, sessionId: parsed.result.sessionId || "" };
    }

    return { text: "", hasTools: false, sessionId: "" };
  }

  function classifyError(err) {
    const msg = (err && err.message ? err.message : String(err)).trim();
    const raw = msg.toLowerCase();
    if (raw.includes("encoded") || raw.includes("obfuscated")) {
      return "Oops! That message looks encoded or obfuscated. Please ask in plain text about Daniela's articles, talks, codelabs, or Go projects.";
    }
    if (raw.includes("model armor") || raw.includes("modelarmor") || raw.includes("security policy") || raw.includes("blocked")) {
      return "Oops! That request tripped Google Cloud Model Armor safety checks. Please ask about Daniela's articles, talks, codelabs, or Go projects.";
    }
    if (raw.includes("429") || raw.includes("rate_limited") || raw.includes("too many requests") || raw.includes("resource exhausted")) {
      return "Hold on! Too many questions came in at once. Wait a few seconds and try again.";
    }
    if (raw.includes("session_not_found") || raw.includes("http 404")) {
      return "This chat session expired while idle. Send your question again to start a fresh chat.";
    }
    if (raw.includes("invalid_request") || raw.includes("payload_too_large")) {
      return "Please keep questions concise (under 1,200 characters) and plain text.";
    }
    if (raw.includes("failed to fetch") || raw.includes("network")) {
      return "I could not reach the Chibi Dani backend service. Please check your connection and try again.";
    }
    if (err && err.publicServerMessage && msg && !GENERIC_STATUS_MESSAGES.has(raw)) {
      return msg;
    }
    return "Something went wrong while answering your question. Please try again in a moment.";
  }

  function initStage() {
    const root = document.getElementById("chibi-stage-root");
    if (!root) return;

    const endpoint = root.dataset.chibiEndpoint || "";
    const spriteImg = document.getElementById("chibi-sprite");
    const portraitFrame = document.getElementById("chibi-portrait-frame");
    const messagesEl = document.getElementById("chibi-messages");
    const form = document.getElementById("chibi-form");
    const input = document.getElementById("chibi-input");
    const sendBtn = document.getElementById("chibi-send");
    const clearBtn = document.getElementById("chibi-clear");
    const counterEl = document.getElementById("chibi-char-counter");

    let lipFlapTimer = null;
    let settleTimer = null;
    let isGenerating = false;
    let turnIndex = 0;
    const overlayLayers = new Map();
    if (portraitFrame) {
      portraitFrame.querySelectorAll(".chibi-portrait-layer[data-expr]").forEach((layer) => {
        const expr = layer.dataset.expr;
        if (!expr) return;
        overlayLayers.set(expr, layer);
        if (typeof layer.decode === "function") layer.decode().catch(function () {});
      });
    }
    if (spriteImg && typeof spriteImg.decode === "function") {
      spriteImg.decode().catch(function () {});
    }

    function clearSettleTimer() {
      if (settleTimer) {
        clearTimeout(settleTimer);
        settleTimer = null;
      }
    }

    function setMood(mood) {
      if (portraitFrame) portraitFrame.setAttribute("data-mood", mood);
      for (const [expr, layer] of overlayLayers) {
        layer.classList.toggle("is-active", expr === mood);
      }
    }

    function startLipFlap() {
      if (lipFlapTimer) return;
      clearSettleTimer();
      let mouthOpen = true;
      setMood("talk");
      if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
        return;
      }
      const talkLayer = overlayLayers.get("talk");
      if (!talkLayer) return;
      const step = () => {
        if (!lipFlapTimer) return;
        mouthOpen = !mouthOpen;
        talkLayer.classList.toggle("is-active", mouthOpen);
        lipFlapTimer = setTimeout(step, mouthOpen ? 180 + Math.random() * 70 : 140 + Math.random() * 60);
      };
      lipFlapTimer = setTimeout(step, 200);
    }

    function stopLipFlap(finalMood) {
      if (lipFlapTimer) {
        clearTimeout(lipFlapTimer);
        lipFlapTimer = null;
      }
      clearSettleTimer();
      const targetMood = finalMood || "idle";
      setMood(targetMood);
      if (targetMood === "happy") {
        settleTimer = setTimeout(() => {
          settleTimer = null;
          if (!isGenerating) setMood("idle");
        }, 2400);
      }
    }

    function scrollMessages() {
      if (messagesEl) messagesEl.scrollTop = messagesEl.scrollHeight;
    }

    function appendBubble(role, initialText) {
      const row = document.createElement("div");
      row.className = "chibi-msg chibi-msg--" + role;

      if (role === "assistant") {
        const avatar = document.createElement("img");
        avatar.src = "/images/persona/chibi-icon.webp";
        avatar.alt = "";
        avatar.width = 40;
        avatar.height = 40;
        avatar.className = "chibi-msg-avatar nozoom";
        row.appendChild(avatar);
      }

      const bubble = document.createElement("div");
      bubble.className = "chibi-bubble chibi-bubble--" + role;
      row.appendChild(bubble);

      if (role === "user") {
        bubble.textContent = initialText;
        messagesEl.appendChild(row);
        scrollMessages();
        return { bubble };
      }

      const statusLine = document.createElement("div");
      statusLine.className = "chibi-tool-status";
      statusLine.hidden = true;

      const body = document.createElement("div");
      body.className = "chibi-bubble-body";
      body.innerHTML = '<span class="chibi-thinking-dots"><span></span><span></span><span></span></span>';

      bubble.appendChild(statusLine);
      bubble.appendChild(body);
      messagesEl.appendChild(row);
      scrollMessages();
      return { bubble, statusLine, body };
    }

    async function sendPrompt(promptText) {
      if (isGenerating || !promptText) return;
      if (!endpoint) {
        const ui = appendBubble("assistant");
        ui.bubble.classList.add("is-error");
        ui.body.textContent = "Chibi Dani endpoint is not configured on this page.";
        setMood("error");
        return;
      }

      isGenerating = true;
      turnIndex += 1;
      clearSettleTimer();
      if (sendBtn) sendBtn.disabled = true;
      if (input) input.disabled = true;

      sendTelemetry("chibi_ask", {
        lang: (document.documentElement.lang || "en").toLowerCase(),
        turn_index: turnIndex,
      });
      appendBubble("user", promptText);
      setMood("thinking");
      const ui = appendBubble("assistant");

      const sessionId = sessionStorage.getItem(SESSION_KEY) || "";
      let accumulated = "";
      let settledMood = "happy";

      try {
        const res = await fetch(endpoint, {
          method: "POST",
          headers: { "Content-Type": "application/json", Accept: "text/event-stream" },
          body: JSON.stringify({
            init: sessionId ? { sessionId } : {},
            data: { message: { role: "user", content: [{ text: promptText }] } },
          }),
        });

        if (!res.ok) {
          if (res.status === 404) {
            sessionStorage.removeItem(SESSION_KEY);
          }
          let detail = "HTTP " + res.status;
          try {
            const errJson = await res.json();
            detail = errJson.error || errJson.message || detail;
          } catch {}
          throw new Error(detail);
        }

        const reader = res.body.getReader();
        const decoder = new TextDecoder("utf-8");
        let buffer = "";

        while (true) {
          const { value, done } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          for (const rawLine of lines) {
            const trimmed = rawLine.trim();
            if (!trimmed.startsWith("data:")) continue;
            const dataStr = trimmed.slice(5).trim();
            if (dataStr === "[DONE]") continue;

            const ev = parseSSEEvent(dataStr);
            if (ev.sessionId) sessionStorage.setItem(SESSION_KEY, ev.sessionId);
            if (ev.hasTools) {
              ui.statusLine.hidden = false;
              ui.statusLine.textContent = "📖 Reading from danicat.dev...";
            }
            const delta = ev.text || (!accumulated && ev.resultText ? ev.resultText : "");
            if (delta) {
              ui.statusLine.hidden = true;
              accumulated += delta;
              startLipFlap();
              ui.body.innerHTML = appendStreamingCursor(renderMarkdown(accumulated));
              scrollMessages();
            }
          }
        }

        ui.statusLine.hidden = true;
        ui.body.innerHTML = renderMarkdown(accumulated || "I could not find a published note on that topic.");
      } catch (err) {
        settledMood = "error";
        ui.statusLine.hidden = true;
        ui.bubble.classList.add("is-error");
        ui.body.textContent = classifyError(err);
      } finally {
        stopLipFlap(settledMood);
        isGenerating = false;
        if (sendBtn) sendBtn.disabled = false;
        if (input) {
          input.disabled = false;
          input.focus();
        }
        scrollMessages();
      }
    }

    if (messagesEl) {
      messagesEl.addEventListener("click", (e) => {
        const chip = e.target.closest(".chibi-starter-chip");
        if (chip) {
          const promptText = (chip.getAttribute("data-prompt") || chip.textContent || "").trim();
          if (promptText && !isGenerating) {
            if (input) {
              input.value = "";
              input.dispatchEvent(new Event("input"));
            }
            sendPrompt(promptText);
          }
          return;
        }
        const link = e.target.closest("a.chibi-link");
        if (link) {
          sendTelemetry("chibi_citation_click", {
            citation_url: link.href,
            is_external: link.origin !== window.location.origin,
          });
        }
      });
    }

    if (clearBtn) {
      clearBtn.addEventListener("click", () => {
        if (isGenerating) return;
        clearSettleTimer();
        turnIndex = 0;
        sessionStorage.removeItem(SESSION_KEY);
        const welcomeNode = messagesEl.querySelector('[data-welcome="true"]');
        messagesEl.textContent = "";
        if (welcomeNode) messagesEl.appendChild(welcomeNode);
        setMood("idle");
      });
    }

    if (input) {
      input.addEventListener("input", () => {
        if (counterEl) counterEl.textContent = input.value.length + " / 1200";
      });
      input.addEventListener("keydown", (event) => {
        if (event.key === "Enter" && !event.shiftKey) {
          if (event.isComposing || event.keyCode === 229) return;
          event.preventDefault();
          form.requestSubmit();
        }
      });
    }

    if (form) {
      form.addEventListener("submit", (event) => {
        event.preventDefault();
        if (isGenerating || !input) return;
        const text = input.value.trim();
        if (!text) return;
        input.value = "";
        if (counterEl) counterEl.textContent = "0 / 1200";
        sendPrompt(text);
      });
    }

    setMood("idle");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initStage);
  } else {
    initStage();
  }
})();
