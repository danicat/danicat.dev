(function () {
  if (!window.firebaseConfig || !window.firebaseConfig.apiKey || !window.firebaseConfig.projectId) {
    return;
  }

  const apiKey = window.firebaseConfig.apiKey;
  const projectId = window.firebaseConfig.projectId;
  const baseDocPath = `projects/${projectId}/databases/(default)/documents`;
  const TOKEN_KEY = "danicat_fb_id_token";
  const TOKEN_EXP_KEY = "danicat_fb_id_token_exp";

  let authPromise = null;

  function formatNumber(x) {
    return Number(x || 0)
      .toString()
      .replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  }

  function setCounterNode(node, value) {
    if (!node) return;
    node.innerText = formatNumber(value);
    if (node.classList.contains("animate-pulse")) {
      node.className = "";
    }
  }

  async function getToken() {
    try {
      const cachedToken = sessionStorage.getItem(TOKEN_KEY);
      const cachedExp = Number(sessionStorage.getItem(TOKEN_EXP_KEY) || 0);
      if (cachedToken && Date.now() < cachedExp) {
        return cachedToken;
      }
    } catch (_) {
      // sessionStorage unavailable in restricted contexts
    }

    if (!authPromise) {
      authPromise = fetch(`https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=${apiKey}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ returnSecureToken: true }),
      })
        .then(async (res) => {
          if (!res.ok) {
            const errText = await res.text();
            throw new Error(`Firebase anonymous auth failed (${res.status}): ${errText}`);
          }
          const data = await res.json();
          try {
            sessionStorage.setItem(TOKEN_KEY, data.idToken);
            sessionStorage.setItem(TOKEN_EXP_KEY, String(Date.now() + 3000 * 1000));
          } catch (_) {}
          return data.idToken;
        })
        .finally(() => {
          authPromise = null;
        });
    }

    return authPromise;
  }

  async function commitIncrement(collection, docId, delta, rawNodeId) {
    const token = await getToken();
    const docName = `${baseDocPath}/${collection}/${docId}`;
    const res = await fetch(`https://firestore.googleapis.com/v1/${baseDocPath}:commit`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        writes: [
          {
            transform: {
              document: docName,
              fieldTransforms: [
                {
                  fieldPath: collection,
                  increment: { integerValue: String(delta) },
                },
              ],
            },
          },
        ],
      }),
    });

    if (!res.ok) {
      const errText = await res.text();
      throw new Error(`Firestore commit failed (${res.status}): ${errText}`);
    }

    const result = await res.json();
    const newVal = result?.writeResults?.[0]?.transformResults?.[0]?.integerValue;
    if (newVal !== undefined && rawNodeId) {
      const targetNode = document.getElementById(rawNodeId);
      if (targetNode) {
        setCounterNode(targetNode, Math.max(0, Number(newVal)));
      }
    }
    return newVal;
  }

  window.danicatFirebase = {
    getToken,
    commitIncrement,
    setCounterNode,
  };

  async function hydrateCounters() {
    const viewNodes = Array.from(document.querySelectorAll("span[id^='views_']"));
    const likeNodes = Array.from(document.querySelectorAll("span[id^='likes_']"));

    if (viewNodes.length === 0 && likeNodes.length === 0) {
      return;
    }

    const docMap = new Map();

    for (const node of viewNodes) {
      const id = node.id ? node.id.replaceAll("/", "-") : "";
      if (!id) continue;
      const fullPath = `${baseDocPath}/views/${id}`;
      if (!docMap.has(fullPath)) {
        docMap.set(fullPath, { field: "views", nodes: [] });
      }
      docMap.get(fullPath).nodes.push(node);
    }

    for (const node of likeNodes) {
      const id = node.id ? node.id.replaceAll("/", "-") : "";
      if (!id) continue;
      const fullPath = `${baseDocPath}/likes/${id}`;
      if (!docMap.has(fullPath)) {
        docMap.set(fullPath, { field: "likes", nodes: [] });
      }
      docMap.get(fullPath).nodes.push(node);
    }

    const documents = Array.from(docMap.keys());
    if (documents.length === 0) return;

    try {
      const token = await getToken();
      const res = await fetch(`https://firestore.googleapis.com/v1/${baseDocPath}:batchGet`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ documents }),
      });

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(`Firestore batchGet failed (${res.status}): ${errText}`);
      }

      const items = await res.json();
      for (const item of items) {
        if (item.found && item.found.name) {
          const entry = docMap.get(item.found.name);
          if (entry) {
            const val = Number(item.found.fields?.[entry.field]?.integerValue || 0);
            for (const node of entry.nodes) {
              setCounterNode(node, val);
            }
          }
        } else if (item.missing) {
          const entry = docMap.get(item.missing);
          if (entry) {
            for (const node of entry.nodes) {
              setCounterNode(node, 0);
            }
          }
        }
      }
    } catch (err) {
      console.error("Failed to hydrate Firestore counters:", err);
    }
  }

  function scheduleHydration() {
    if ("requestIdleCallback" in window) {
      window.requestIdleCallback(() => hydrateCounters(), { timeout: 2000 });
    } else {
      setTimeout(hydrateCounters, 100);
    }
  }

  if (document.readyState === "complete") {
    scheduleHydration();
  } else {
    window.addEventListener("load", scheduleHydration, { once: true });
  }
})();
