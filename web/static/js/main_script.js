document.addEventListener("DOMContentLoaded", () => {
  const root = document.documentElement;
  let paletteTimeout = null;

  /* mobile menu */
  const burger = document.querySelector(".burguer");
  const modal = document.getElementById("mobile-menu-modal");
  const closeBtn = document.getElementById("mobile-menu-close");

  if (burger) {
    burger.addEventListener("click", () => {
      modal.classList.add("open");
    });
  }

  if (closeBtn) {
    closeBtn.addEventListener("click", () => {
      modal.classList.remove("open");
    });
  }

  window.addEventListener("resize", () => {
    if (window.innerWidth > 768 && modal.classList.contains("open")) {
      modal.classList.remove("open");
    }
  });

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
