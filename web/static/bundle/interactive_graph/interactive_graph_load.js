(function () {
  const loadingOverlay = document.getElementById("loading-overlay");
  const errorEl = document.getElementById("error");
  const careerSelector = document.getElementById("career-selector");
  let lastLoadedFile = "IIN.json";

  // ── Actualizar el header con datos de la carrera ──
  function updateHeader(data) {
    const nameEl = document.getElementById('career-name');
    const semEl = document.getElementById('stat-semesters');
    const credEl = document.getElementById('stat-credits');
    const subjEl = document.getElementById('stat-subjects');

    if (nameEl) nameEl.textContent = data.career.name || 'Carrera';
    if (semEl) semEl.textContent = `${data.career.totalSemesters || 0} semestres`;
    
    if (credEl) {
      credEl.textContent = data.career.totalCredits
        ? `${data.career.totalCredits} créditos`
        : '';
    }

    if (subjEl && data.subjects) {
      const subjectCount = Object.keys(data.subjects).length;
      subjEl.textContent = `${subjectCount} materias`;
    }
  }

  function showLoading(visible) {
    if (loadingOverlay) {
      loadingOverlay.classList.toggle('hidden', !visible);
    }
  }

  async function loadGraph(subjectFile = "IIN.json") {
    try {
      showLoading(true);

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 8000);

      const response = await fetch(`/static/curriculums/${subjectFile}`, {
        signal: controller.signal,
      });
      clearTimeout(timeoutId);

      if (!response.ok) throw new Error("Error en la respuesta");

      const data = await response.json();

      if (
        typeof prepareGraphData !== "function" ||
        typeof renderGraph !== "function"
      ) {
        throw new Error("Librería no cargada");
      }

      prepareGraphData(data);
      renderGraph("graph-container");
      updateHeader(data);
      lastLoadedFile = subjectFile;

      setTimeout(() => showLoading(false), 300);
    } catch (error) {
      console.error("Error:", error);
      showLoading(false);
      if (errorEl) {
        errorEl.classList.remove("hidden");
        setTimeout(() => errorEl.classList.add("hidden"), 3000);
      }
    }
  }

  // Escucha cambios en el selector de materias
  if (careerSelector) {
    careerSelector.addEventListener("change", () => {
      const selectedFile = careerSelector.value;
      loadGraph(selectedFile);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => loadGraph(careerSelector?.value || "IIN.json"));
  } else {
    loadGraph(careerSelector?.value || "IIN.json");
  }

  // ── Re-renderizar cuando cambia el modo oscuro/claro ──
  const htmlEl = document.documentElement;
  const observer = new MutationObserver((mutations) => {
    for (const m of mutations) {
      if (m.type === "attributes" && m.attributeName === "class") {
        loadGraph(lastLoadedFile);
        break;
      }
    }
  });
  observer.observe(htmlEl, { attributes: true, attributeFilter: ["class"] });
})();
