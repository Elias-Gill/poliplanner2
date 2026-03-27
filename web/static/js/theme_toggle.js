document.addEventListener("DOMContentLoaded", () => {
    const root = document.documentElement;

    const desktopToggle = document.getElementById("theme-toggle-desktop");
    const mobileToggle = document.getElementById("theme-toggle-mobile");

    const storedMode = localStorage.getItem("mode");

    if (storedMode === "dark") {
        root.classList.add("dark-mode");
    } else {
        root.classList.remove("dark-mode");
    }

    const toggleTheme = () => {
        const isDark = root.classList.toggle("dark-mode");
        localStorage.setItem("mode", isDark ? "dark" : "light");
    };

    [desktopToggle, mobileToggle].forEach(btn => {
        if (!btn) return;
        btn.addEventListener("click", toggleTheme);
    });
});
