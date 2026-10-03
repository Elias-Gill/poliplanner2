/**
 * Dashboard global behaviors.
 *
 * Loaded once. Keeps a single MutationObserver and a single htmx listener so the
 * schedule colors survive HTMX swaps without stacking observers on every swap.
 */
(function () {
    if (window.__dashColorsInit) return;
    window.__dashColorsInit = true;

    const colors = [
        { headerBg: "#fee2e2", text: "#991b1b", border: "#fca5a5" },
        { headerBg: "#ffedd5", text: "#9a3412", border: "#fdba74" },
        { headerBg: "#fef9c3", text: "#854d0e", border: "#fde047" },
        { headerBg: "#dcfce7", text: "#166534", border: "#86efac" },
        { headerBg: "#cffafe", text: "#155e75", border: "#67e8f9" },
        { headerBg: "#e0f2fe", text: "#075985", border: "#7dd3fc" },
        { headerBg: "#dbeafe", text: "#1e40af", border: "#93c5fd" },
        { headerBg: "#e0e7ff", text: "#3730a3", border: "#a5b4fc" },
        { headerBg: "#ede9fe", text: "#5b21b6", border: "#c4b5fd" },
        { headerBg: "#fae8ff", text: "#86198f", border: "#f0abfc" },
        { headerBg: "#fce7f3", text: "#9d174d", border: "#f9a8d4" },
        { headerBg: "#ffe4e6", text: "#9f1239", border: "#fda4af" },
    ];

    const colorsDark = [
        { headerBg: "#2a1717", text: "#e2b0b0", border: "#442525" },
        { headerBg: "#2a1c14", text: "#e2bdad", border: "#442d20" },
        { headerBg: "#282313", text: "#ded3aa", border: "#3f381f" },
        { headerBg: "#15241a", text: "#a9d4b8", border: "#213a2a" },
        { headerBg: "#142327", text: "#a8d2dc", border: "#20373e" },
        { headerBg: "#142128", text: "#a7cadb", border: "#1f3441" },
        { headerBg: "#171d29", text: "#b0c0e0", border: "#242f43" },
        { headerBg: "#1c1b29", text: "#b8b6e0", border: "#2c2a43" },
        { headerBg: "#211929", text: "#c5b4e0", border: "#352742" },
        { headerBg: "#271726", text: "#dcafd9", border: "#40243e" },
        { headerBg: "#281720", text: "#ddafc6", border: "#412433" },
        { headerBg: "#28171a", text: "#ddafb5", border: "#412429" },
    ];

    function pickColor(course, isDark) {
        let hash = 0;
        for (let i = 0; i < course.length; i++) {
            hash = course.charCodeAt(i) + ((hash << 5) - hash);
        }
        const idx = Math.abs(hash) % colors.length;
        return isDark ? colorsDark[idx] : colors[idx];
    }

    function applyColors() {
        const isDark = document.documentElement.classList.contains("dark");
        document.querySelectorAll(".class-entry").forEach((el) => {
            const course = el.dataset.course;
            if (!course) return;

            const c = pickColor(course, isDark);
            const headerEl = el.querySelector(".class-header");

            el.style.borderColor = c.border;
            if (headerEl) {
                headerEl.style.backgroundColor = c.headerBg;
                headerEl.style.color = c.text;
            }
        });
    }

    applyColors();

    new MutationObserver(applyColors).observe(document.documentElement, {
        attributes: true,
        attributeFilter: ["class"],
    });

    document.addEventListener("htmx:afterSettle", applyColors);
})();
