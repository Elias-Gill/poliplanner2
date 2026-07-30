document.addEventListener("DOMContentLoaded", () => {
    const root = document.documentElement;

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

    document.querySelectorAll(".theme-toggle").forEach(btn => {
        btn.addEventListener("click", toggleTheme);
    });
});
