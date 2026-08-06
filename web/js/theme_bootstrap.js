(function () {
  const root = document.documentElement;

  const theme = localStorage.getItem("theme") || "default-theme";
  const mode = localStorage.getItem("mode");

  root.classList.add(theme);

  if (mode === "dark") {
    root.classList.add("dark-mode");
  } else if (!mode) {
    const prefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)",
    ).matches;
    if (prefersDark) {
      root.classList.add("dark-mode");
    }
  }
})();
