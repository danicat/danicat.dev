(function () {
  "use strict";

  function detectPageLang() {
    var path = window.location.pathname || "/";
    if (path === "/pt-br" || path.indexOf("/pt-br/") === 0) return "pt-br";
    if (path === "/ja" || path.indexOf("/ja/") === 0) return "ja";
    var docLang = (document.documentElement.lang || "").toLowerCase();
    if (docLang.indexOf("pt") === 0) return "pt-br";
    if (docLang.indexOf("ja") === 0) return "ja";
    return "en";
  }

  function langPrefix(lang) {
    if (lang === "pt-br" || lang === "ja") return "/" + lang;
    return "";
  }

  function registerWebMCPTools() {
    var ctx = document.modelContext || navigator.modelContext;
    if (!ctx || typeof ctx.registerTool !== "function") {
      return;
    }

    var cachedByLang = {};
    async function loadSearchIndex(lang) {
      var targetLang = lang === "pt-br" || lang === "ja" || lang === "en" ? lang : detectPageLang();
      if (cachedByLang[targetLang]) {
        return cachedByLang[targetLang];
      }
      var indexUrl = langPrefix(targetLang) + "/index.json";
      var res = await fetch(indexUrl, {
        headers: { Accept: "application/json" },
      });
      if (!res.ok) {
        throw new Error("Failed to fetch " + indexUrl + ": HTTP " + res.status);
      }
      cachedByLang[targetLang] = await res.json();
      return cachedByLang[targetLang];
    }

    ctx.registerTool({
      name: "search_content",
      description:
        "Search danicat.dev technical articles, codelabs, videos, and conference talks by keyword, category, tag, or language.",
      inputSchema: {
        type: "object",
        properties: {
          query: {
            type: "string",
            description:
              "Case-insensitive search query matched against title, summary, tags, and category.",
          },
          category: {
            type: "string",
            description:
              "Optional category filter (e.g. 'Agentic Coding', 'Agent Development', 'Applied GenAI', 'Perspectives', 'Software Engineering').",
          },
          section: {
            type: "string",
            description:
              "Optional content section filter ('posts', 'codelabs', 'videos', 'events', 'about').",
          },
          lang: {
            type: "string",
            description:
              "Optional language code filter ('en', 'pt-br', 'ja'). Defaults to the active page language.",
          },
          limit: {
            type: "number",
            description: "Maximum number of results to return (default 10).",
          },
        },
        required: ["query"],
      },
      annotations: {
        readOnlyHint: true,
      },
      execute: async function (args) {
        var langFilter = args && args.lang ? String(args.lang).toLowerCase() : "";
        var items = await loadSearchIndex(langFilter || detectPageLang());
        var q = String((args && args.query) || "")
          .toLowerCase()
          .trim();
        var catFilter = args && args.category ? String(args.category).toLowerCase() : "";
        var secFilter = args && args.section ? String(args.section).toLowerCase() : "";
        var limit = args && args.limit ? Number(args.limit) : 10;

        var matches = items.filter(function (item) {
          if (secFilter && String(item.section || "").toLowerCase() !== secFilter) {
            return false;
          }
          if (langFilter && String(item.lang || "").toLowerCase() !== langFilter) {
            return false;
          }
          if (catFilter) {
            var cats = (item.categories || []).map(function (c) {
              return String(c).toLowerCase();
            });
            if (cats.indexOf(catFilter) === -1) {
              return false;
            }
          }
          if (!q) {
            return true;
          }
          var haystack = [
            item.title || "",
            item.summary || "",
            (item.categories || []).join(" "),
            (item.tags || []).join(" "),
            (item.series || []).join(" "),
          ]
            .join(" ")
            .toLowerCase();
          return haystack.indexOf(q) !== -1;
        });

        var sliced = matches.slice(0, limit);
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(
                {
                  totalMatches: matches.length,
                  returned: sliced.length,
                  results: sliced,
                },
                null,
                2
              ),
            },
          ],
        };
      },
    });

    ctx.registerTool({
      name: "get_agent_resources",
      description:
        "Retrieve machine-readable entrypoints for danicat.dev (llms.txt, llms-full.txt, ai-catalog.json, and current page Markdown URL).",
      inputSchema: {
        type: "object",
        properties: {},
      },
      annotations: {
        readOnlyHint: true,
      },
      execute: async function () {
        var activeLang = detectPageLang();
        var prefix = langPrefix(activeLang);
        var mdLink = document.querySelector('link[rel="alternate"][type="text/markdown"]');
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(
                {
                  lang: activeLang,
                  currentPageMarkdown: mdLink ? mdLink.href : window.location.pathname.replace(/\/?$/, "/index.md"),
                  llmsTxt: prefix + "/llms.txt",
                  llmsFullTxt: prefix + "/llms-full.txt",
                  aiCatalog: "/.well-known/ai-catalog.json",
                  apiCatalog: "/.well-known/api-catalog",
                  searchIndex: prefix + "/index.json",
                },
                null,
                2
              ),
            },
          ],
        };
      },
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", registerWebMCPTools);
  } else {
    registerWebMCPTools();
  }
})();
