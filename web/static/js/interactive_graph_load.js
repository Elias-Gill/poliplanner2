(function () {
  const loadingEl = document.getElementById("loading");
  const errorEl = document.getElementById("error");

  async function loadGraph() {
    try {
      loadingEl.classList.remove("hidden");

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 8000);

      const response = await fetch("/static/curriculums/IIN.json", {
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

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", loadGraph);
  } else {
    loadGraph();
  }
})();
