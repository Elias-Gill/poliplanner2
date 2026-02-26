(function() {
    const hint = document.getElementById("theme-hint");
    const toggle = document.getElementById("theme-toggle");

    if (hint && toggle) {
        const hasSeenHint = localStorage.getItem("theme-hint-seen");

        // Por defecto oculto
        hint.hidden = true;

        // Mostrar hint automáticamente solo la primera vez
        if (!hasSeenHint) {
            hint.hidden = false;

            // Ocultar el hint al primer mouseleave o touchend y guardar en localStorage
            const hideAndSave = () => {
                hint.hidden = true;
                localStorage.setItem("theme-hint-seen", "true");
                toggle.removeEventListener("mouseleave", hideAndSave);
                toggle.removeEventListener("touchend", hideAndSave);
            };

            toggle.addEventListener("mouseleave", hideAndSave);
            toggle.addEventListener("touchend", hideAndSave);
        }

        // Eventos normales para mostrar/ocultar hint
        toggle.addEventListener("mouseenter", () => {
            hint.hidden = false;
        });

        toggle.addEventListener("mouseleave", () => {
            // Solo ocultar si ya vio el hint la primera vez
            if (localStorage.getItem("theme-hint-seen")) {
                hint.hidden = true;
            }
        });

        // Soporte táctil básico
        toggle.addEventListener("touchstart", () => {
            hint.hidden = false;
        });

        toggle.addEventListener("touchend", () => {
            if (localStorage.getItem("theme-hint-seen")) {
                hint.hidden = true;
            }
        });
    }
})();
