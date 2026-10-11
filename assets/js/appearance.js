const sitePreference = document.documentElement.getAttribute("data-default-appearance");
const userPreference = localStorage.getItem("appearance");

if ((sitePreference === "dark" && userPreference === null) || userPreference === "dark") {
  document.documentElement.classList.add("dark");
}

if (document.documentElement.getAttribute("data-auto-appearance") === "true") {
  if (
    window.matchMedia &&
    window.matchMedia("(prefers-color-scheme: dark)").matches &&
    userPreference !== "light"
  ) {
    document.documentElement.classList.add("dark");
  }
  window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", (event) => {
    if (event.matches) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
    updateMeta();
  });
}

var getTargetAppearance = () => {
  return document.documentElement.classList.contains("dark") ? "dark" : "light";
};

var updateMeta = () => {
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) {
    meta.setAttribute(
      "content",
      getTargetAppearance() === "dark" ? "rgb(30, 41, 59)" : "rgb(255, 255, 255)"
    );
  }
};

var updateMermaidTheme = () => {
  if (typeof mermaid !== "undefined") {
    const isDark = document.documentElement.classList.contains("dark");
    const mermaids = document.querySelectorAll("pre.mermaid");
    mermaids.forEach((e) => {
      if (e.getAttribute("data-processed")) {
        e.removeAttribute("data-processed");
        e.innerHTML = e.getAttribute("data-graph");
      } else {
        e.setAttribute("data-graph", e.textContent);
      }
    });
    if (isDark) {
      initMermaidDark();
      mermaid.run();
    } else {
      initMermaidLight();
      mermaid.run();
    }
  }
};

updateMeta();

window.addEventListener("DOMContentLoaded", () => {
  const switcher = document.getElementById("appearance-switcher");
  const switcherMobile = document.getElementById("appearance-switcher-mobile");

  updateMermaidTheme();

  const toggleTheme = () => {
    document.documentElement.classList.toggle("dark");
    const targetAppearance = getTargetAppearance();
    localStorage.setItem("appearance", targetAppearance);
    updateMeta();
    updateMermaidTheme();
  };

  if (switcher) {
    switcher.addEventListener("click", toggleTheme);
    switcher.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      localStorage.removeItem("appearance");
    });
  }
  if (switcherMobile) {
    switcherMobile.addEventListener("click", toggleTheme);
    switcherMobile.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      localStorage.removeItem("appearance");
    });
  }
});
