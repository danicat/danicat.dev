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

    localStorage.setItem(id, "true");
    window.danicatFirebase
      .commitIncrement("views", id, 1, oid)
      .catch((err) => {
        localStorage.removeItem(id);
        console.error("Failed to record page view:", err);
      });
  }

  function scheduleViewIncrement() {
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        setTimeout(recordViewIfNeeded, 300);
      });
    });
  }

  if (document.readyState === "complete") {
    scheduleViewIncrement();
  } else {
    window.addEventListener("load", scheduleViewIncrement, { once: true });
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
