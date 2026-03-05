document.addEventListener("DOMContentLoaded", () => {
    const root = document.documentElement;

    // Botones desktop y mobile
    const desktopToggle = document.getElementById("theme-toggle-desktop");
    const mobileToggle = document.getElementById("theme-toggle-mobile");

    // Función para actualizar el ícono según modo
    const updateToggleIcon = () => {
        if (desktopToggle) {
            desktopToggle.innerHTML = root.classList.contains("dark-mode")
                ? '<i class="bi bi-sun text-xl"></i>'
                : '<i class="bi bi-moon text-xl"></i>';
        }
        if (mobileToggle) {
            mobileToggle.innerHTML = root.classList.contains("dark-mode")
                ? '<i class="bi bi-sun text-lg"></i>'
                : '<i class="bi bi-moon text-lg"></i>';
        }
    };

    // Inicializar según localStorage
    const storedMode = localStorage.getItem("mode");
    if (storedMode === "dark") {
        root.classList.add("dark-mode");
    } else {
        root.classList.remove("dark-mode");
    }
    updateToggleIcon();

    // Función para alternar modo
    const toggleTheme = () => {
        const isDark = root.classList.toggle("dark-mode");
        localStorage.setItem("mode", isDark ? "dark" : "light");
        updateToggleIcon();
    };

    // Agregar listeners a ambos botones
    [desktopToggle, mobileToggle].forEach(btn => {
        if (!btn) return;

        // Click normal alterna modo
        btn.addEventListener("click", toggleTheme);

        // Opcional: mantener presionado podría abrir paleta si la agregas
        // btn.addEventListener("mousedown", ...);
        // btn.addEventListener("mouseup", ...);
        // btn.addEventListener("touchstart", ...);
        // btn.addEventListener("touchend", ...);
    });
});
