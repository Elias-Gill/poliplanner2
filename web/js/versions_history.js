      // automaticamente cargado desde "versions.json" (?v= para evitar cache del navegador)
      fetch("/static/versions.json?v=")
        .then((res) => res.json())
        .then((data) => {
          const historyContainer = document.getElementById("news-history");
          if (!historyContainer || !data.entries) return;

          let html = "";
          data.entries.forEach((entry) => {
            html += `<h3>Versión ${entry.version} (${entry.date}) — ${entry.title}</h3>`;
            html += "<ul>";
            entry.changes.forEach((change) => {
              html += `<li>${change}</li>`;
            });
            html += "</ul>";
          });

          historyContainer.innerHTML = html;
        })
        .catch((err) =>
          console.error("Error al cargar historial de noticias:", err),
        );
