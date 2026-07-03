(function () {
  const loadingEl = document.getElementById("loading");
  const errorEl = document.getElementById("error");
  const subjectSelect = document.getElementById("subject-select"); // tu select manual

  async function loadGraph(subjectFile = "IIN.json") {
    try {
      loadingEl.classList.remove("hidden");

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

      setTimeout(() => loadingEl.classList.add("hidden"), 300);
    } catch (error) {
      console.error("Error:", error);
      loadingEl.classList.add("hidden");
      errorEl.classList.remove("hidden");
      setTimeout(() => errorEl.classList.add("hidden"), 3000);
    }
  }

  // Escucha cambios en el selector de materias
  if (subjectSelect) {
    subjectSelect.addEventListener("change", () => {
      const selectedFile = subjectSelect.value;
      loadGraph(selectedFile);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => loadGraph());
  } else {
    loadGraph();
  }
})();
