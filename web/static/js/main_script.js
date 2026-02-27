function toggleMobileMenu() {
    const menu = document.getElementById('mobile-menu');
    menu.classList.toggle('hidden');
    if (!menu.classList.contains('hidden')) {
        // Scroll al top para que el menú sea visible inmediatamente
        menu.scrollTop = 0;
    }
}

document.addEventListener("DOMContentLoaded", () => {
  const root = document.documentElement;
  let paletteTimeout = null;

  /* dark / light toggle */
  const themeToggle = document.getElementById("theme-toggle");
  const paletteSelector = document.getElementById("palette-selector");

  const updateToggleIcon = () => {
    themeToggle.innerHTML = root.classList.contains("dark-mode")
      ? '<i class="bi bi-sun"></i>'
      : '<i class="bi bi-moon"></i>';
  };

  updateToggleIcon();

  // Alterna modo claro / oscuro con click normal
  themeToggle.addEventListener("click", (e) => {
    // Solo alternar si no estamos mostrando la paleta
    if (!paletteSelector.classList.contains("visible")) {
      const isDark = root.classList.toggle("dark-mode");
      localStorage.setItem("mode", isDark ? "dark" : "light");
      updateToggleIcon();
    }
  });

  // Función para ocultar la paleta
  const hidePalette = () => {
    paletteSelector.classList.remove("visible");
    clearTimeout(paletteTimeout);
  };

  // Mostrar paleta al mantener presionado (móvil y desktop)
  themeToggle.addEventListener("mousedown", () => {
    paletteTimeout = setTimeout(() => {
      paletteSelector.classList.add("visible");
    }, 500); // 500ms para evitar activaciones accidentales
  });

  themeToggle.addEventListener("mouseup", () => {
    clearTimeout(paletteTimeout);
  });

  themeToggle.addEventListener("mouseleave", () => {
    clearTimeout(paletteTimeout);
  });

  // También para touch devices
  themeToggle.addEventListener("touchstart", () => {
    paletteTimeout = setTimeout(() => {
      paletteSelector.classList.add("visible");
    }, 500);
  });

  themeToggle.addEventListener("touchend", () => {
    clearTimeout(paletteTimeout);
  });

  // Ocultar paleta al hacer clic fuera
  document.addEventListener("click", (e) => {
    if (
      !themeToggle.contains(e.target) &&
      !paletteSelector.contains(e.target)
    ) {
      hidePalette();
    }
  });

  // Ocultar paleta en móvil al hacer scroll
  window.addEventListener("scroll", () => {
    if (window.innerWidth <= 768) {
      hidePalette();
    }
  });

  // Selección de tema
  paletteSelector.addEventListener("click", (e) => {
    if (e.target.tagName === "BUTTON") {
      const theme = e.target.dataset.theme;

      root.classList.remove(
        "default-theme",
        "theme-warm",
        "theme-forest",
        "theme-lavender",
        "theme-pink",
      );
      root.classList.add(theme);
      localStorage.setItem("theme", theme);

      hidePalette();
    }
  });

  /* duplicate menu for mobile */
  const desktopMenu = document.querySelector(".nav-links");
  const mobileMenuContent = document.getElementById("mobile-menu-content");

  if (desktopMenu && mobileMenuContent) {
    mobileMenuContent.innerHTML = "";
    const links = desktopMenu.querySelectorAll("li.right-menu-item > a");
    links.forEach((link) => {
      mobileMenuContent.appendChild(link.cloneNode(true));
    });
  }
});
