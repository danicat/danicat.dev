(function () {
  const pageScriptElement =
    document.currentScript || document.querySelector("script[data-oid]");
  const oid = pageScriptElement ? pageScriptElement.getAttribute("data-oid") : null;
  const oid_likes = pageScriptElement ? pageScriptElement.getAttribute("data-oid-likes") : null;

  let liked_page = false;
  const id = oid ? oid.replaceAll("/", "-") : null;
  const id_likes = oid_likes ? oid_likes.replaceAll("/", "-") : null;

  function updateLikeButtonUI(isLiked) {
    const heart = document.getElementById("button_likes_heart");
    const emptyHeart = document.getElementById("button_likes_emtpty_heart");
    const text = document.getElementById("button_likes_text");
    if (heart) heart.style.display = isLiked ? "" : "none";
    if (emptyHeart) emptyHeart.style.display = isLiked ? "none" : "";
    if (text) text.innerText = isLiked ? "" : "\xa0Like";
  }

  if (id_likes && localStorage.getItem(id_likes)) {
    liked_page = true;
    updateLikeButtonUI(true);
  }

  function recordViewIfNeeded() {
    if (!id || localStorage.getItem(id)) return;
    if (!window.danicatFirebase) return;

    // Ignore headless automation (Lighthouse, PageSpeed Insights, Puppeteer, scrapers)
    if (navigator.webdriver) return;

    // Ignore background tabs until the user actually views the page
    if (document.visibilityState !== "visible") {
      document.addEventListener("visibilitychange", function onVisible() {
        if (document.visibilityState === "visible") {
          document.removeEventListener("visibilitychange", onVisible);
          scheduleViewIncrement();
        }
      });
      return;
    }

    localStorage.setItem(id, "true");
    window.danicatFirebase
      .commitIncrement("views", id, 1, oid)
      .catch((err) => {
        localStorage.removeItem(id);
        console.error("Failed to record page view:", err);
      });
  }

  function scheduleViewIncrement() {
    let recorded = false;
    const trigger = () => {
      if (recorded) return;
      recorded = true;
      cleanup();
      recordViewIfNeeded();
    };

    const cleanup = () => {
      window.removeEventListener("scroll", trigger);
      window.removeEventListener("touchstart", trigger);
      clearTimeout(timer);
    };

    // Count view after 3s visible dwell OR upon first scroll/touch
    const timer = setTimeout(trigger, 3000);
    window.addEventListener("scroll", trigger, { once: true, passive: true });
    window.addEventListener("touchstart", trigger, { once: true, passive: true });
  }

  function initScrollDepthMilestones() {
    if (navigator.webdriver || typeof IntersectionObserver !== "function") return;
    const container = document.querySelector(".article-content");
    if (!container) return;

    container.style.position = "relative";

    const milestones = [25, 50, 75, 100];
    const observer = new IntersectionObserver(
      (entries, obs) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue;
          obs.unobserve(entry.target);
          const eventName = entry.target.getAttribute("data-scroll-milestone");
          if (eventName && typeof window.gtag === "function") {
            window.gtag("event", eventName);
          }
        }
      },
      { threshold: 0 }
    );

    for (const pct of milestones) {
      const sentinel = document.createElement("span");
      sentinel.setAttribute("aria-hidden", "true");
      sentinel.setAttribute("data-scroll-milestone", `scroll_${pct}`);
      sentinel.style.cssText = `position:absolute;top:${pct}%;left:0;width:1px;height:1px;pointer-events:none;opacity:0;`;
      container.appendChild(sentinel);
      observer.observe(sentinel);
    }
  }

  function onPageReady() {
    scheduleViewIncrement();
    const scheduleIdle = window.requestIdleCallback || ((cb) => setTimeout(cb, 200));
    scheduleIdle(initScrollDepthMilestones);
  }

  if (document.readyState === "complete") {
    onPageReady();
  } else {
    window.addEventListener("load", onPageReady, { once: true });
  }

  window.process_article = function () {
    if (!id_likes || !window.danicatFirebase) return;
    if (!liked_page) {
      liked_page = true;
      localStorage.setItem(id_likes, "true");
      updateLikeButtonUI(true);
      window.danicatFirebase
        .commitIncrement("likes", id_likes, 1, oid_likes)
        .catch((err) => {
          console.error("Failed to like article:", err);
        });
    } else {
      liked_page = false;
      localStorage.removeItem(id_likes);
      updateLikeButtonUI(false);
      window.danicatFirebase
        .commitIncrement("likes", id_likes, -1, oid_likes)
        .catch((err) => {
          console.error("Failed to unlike article:", err);
        });
    }
  };
})();
