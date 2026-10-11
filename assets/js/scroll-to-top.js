(function () {
  let isVisible = false;
  let ticking = false;

  function updateScrollToTop() {
    ticking = false;
    const el = document.getElementById("scroll-to-top");
    if (!el) return;
    const shouldShow = window.scrollY > window.innerHeight * 0.5;
    if (shouldShow === isVisible) return;
    isVisible = shouldShow;
    if (shouldShow) {
      el.classList.remove("translate-y-4", "opacity-0");
      el.classList.add("translate-y-0", "opacity-100");
    } else {
      el.classList.remove("translate-y-0", "opacity-100");
      el.classList.add("translate-y-4", "opacity-0");
    }
  }

  window.addEventListener(
    "scroll",
    function () {
      if (!ticking) {
        window.requestAnimationFrame(updateScrollToTop);
        ticking = true;
      }
    },
    { passive: true }
  );
})();
