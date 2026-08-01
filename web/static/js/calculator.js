document.addEventListener('DOMContentLoaded', () => {
  const toggleDirecto = document.getElementById('toggleDirecto');
  const panelDirecto = document.getElementById('panelDirecto');
  const panelDesglose = document.getElementById('panelDesglose');
  
  const inputDirecto = document.getElementById('inputPonderadoDirecto');
  
  // Parciales
  const notaP1 = document.getElementById('notaP1');
  const notaP2 = document.getElementById('notaP2');
  const pesoParciales = document.getElementById('pesoParciales');
  
  // Laboratorio y sus toggles
  const checkAplicaLab = document.getElementById('checkAplicaLab');
  const wrapperPesoLab = document.getElementById('wrapperPesoLab');
  const contentLab = document.getElementById('contentLab');
  const notaLab = document.getElementById('notaLab');
  const pesoLab = document.getElementById('pesoLab');

  // Tareas y sus toggles
  const checkAplicaTareas = document.getElementById('checkAplicaTareas');
  const wrapperPesoTareas = document.getElementById('wrapperPesoTareas');
  const contentTareas = document.getElementById('contentTareas');
  const notaTareas = document.getElementById('notaTareas');
  const pesoTareas = document.getElementById('pesoTareas');

  // UI
  const valPromedio = document.getElementById('valPromedio');
  const badgeEstado = document.getElementById('badgeEstado');
  const msgEstado = document.getElementById('msgEstado');
  const tablaResultados = document.getElementById('tablaResultados');
  const sumaPesosVal = document.getElementById('sumaPesosVal');
  const alertaPesos = document.getElementById('alertaPesos');

  // Alternar vista Directa vs Desglose
  toggleDirecto.addEventListener('change', () => {
    if (toggleDirecto.checked) {
      panelDirecto.classList.remove('hidden');
      panelDesglose.classList.add('hidden');
    } else {
      panelDirecto.classList.add('hidden');
      panelDesglose.classList.remove('hidden');
    }
    calcular();
  });

  // Toggle Laboratorio
  checkAplicaLab.addEventListener('change', () => {
    if (checkAplicaLab.checked) {
      wrapperPesoLab.classList.remove('hidden');
      contentLab.classList.remove('hidden');
    } else {
      wrapperPesoLab.classList.add('hidden');
      contentLab.classList.add('hidden');
      notaLab.value = '';
    }
    calcular();
  });

  // Toggle Tareas
  checkAplicaTareas.addEventListener('change', () => {
    if (checkAplicaTareas.checked) {
      wrapperPesoTareas.classList.remove('hidden');
      contentTareas.classList.remove('hidden');
    } else {
      wrapperPesoTareas.classList.add('hidden');
      contentTareas.classList.add('hidden');
      notaTareas.value = '';
    }
    calcular();
  });

  // Escuchar eventos en todos los inputs
  document.querySelectorAll('input').forEach(input => {
    input.addEventListener('input', calcular);
  });

  function calcular() {
    let promedio = 0;

    if (toggleDirecto.checked) {
      promedio = parseFloat(inputDirecto.value) || 0;
    } else {
      const pParciales = parseFloat(pesoParciales.value) || 0;
      
      const aplicaLab = checkAplicaLab.checked;
      const pLab = aplicaLab ? (parseFloat(pesoLab.value) || 0) : 0;
      const nLab = aplicaLab ? (parseFloat(notaLab.value) || 0) : 0;

      const aplicaTareas = checkAplicaTareas.checked;
      const pTareas = aplicaTareas ? (parseFloat(pesoTareas.value) || 0) : 0;
      const nTareas = aplicaTareas ? (parseFloat(notaTareas.value) || 0) : 0;

      const sumaPesos = pParciales + pLab + pTareas;
      sumaPesosVal.textContent = sumaPesos;

      if (sumaPesos !== 100) {
        alertaPesos.className = "text-xs text-center py-1.5 px-3 rounded-sm bg-red-100 text-red-700 font-medium";
      } else {
        alertaPesos.className = "text-xs text-center py-1.5 px-3 rounded-sm bg-gray-100 text-gray-600 font-medium";
      }

      // Promedio de Parciales
      const np1 = parseFloat(notaP1.value);
      const np2 = parseFloat(notaP2.value);
      let promParciales = 0;

      if (!isNaN(np1) && !isNaN(np2)) {
        promParciales = (np1 + np2) / 2;
      } else if (!isNaN(np1)) {
        promParciales = np1;
      } else if (!isNaN(np2)) {
        promParciales = np2;
      }

      promedio = (promParciales * (pParciales / 100)) + 
                 (nLab * (pLab / 100)) + 
                 (nTareas * (pTareas / 100));
    }

    promedio = Math.min(100, Math.max(0, promedio));
    valPromedio.textContent = `${promedio.toFixed(1)}%`;

    actualizarFirmas(promedio);
    actualizarTablaExamen(promedio);
  }

  function actualizarFirmas(promedio) {
    if (promedio < 50) {
      badgeEstado.className = "inline-block px-3 py-1 rounded-sm text-xs font-bold bg-red-100 text-red-700 mb-4";
      badgeEstado.textContent = "Sin Firma (Reprobado)";
      msgEstado.textContent = "No alcanzas el mínimo (50%) para habilitar el examen final.";
    } else if (promedio < 60) {
      badgeEstado.className = "inline-block px-3 py-1 rounded-sm text-xs font-bold bg-yellow-100 text-yellow-800 mb-4";
      badgeEstado.textContent = "Media Firma (2do final)";
      msgEstado.textContent = "Habilitado únicamente para rendir el 2do examen final.";
    } else {
      badgeEstado.className = "inline-block px-3 py-1 rounded-sm text-xs font-bold bg-green-100 text-green-700 mb-4";
      badgeEstado.textContent = "Dos Firmas (Ambos Llamados)";
      msgEstado.textContent = "Habilitado para rendir el 1er y 2do examen final.";
    }
  }

  function actualizarTablaExamen(promedio) {
    if (promedio < 50) {
      tablaResultados.innerHTML = `
        <tr>
          <td colspan="2" class="px-3 py-4 text-center text-xs text-red-500 font-medium">
            Se requiere al menos 50% de promedio ponderado para rendir el examen final.
          </td>
        </tr>`;
      return;
    }

    const limitesNotas = [
      { nota: 2, minFinal: 60 },
      { nota: 3, minFinal: 71 },
      { nota: 4, minFinal: 81 },
      { nota: 5, minFinal: 91 }
    ];

    let html = '';

    limitesNotas.forEach(item => {
      let reqExamen = (item.minFinal - (promedio * 0.4)) / 0.6;
      let examenFinalRequerido = Math.max(50, Math.ceil(reqExamen));

      if (examenFinalRequerido > 100) {
        html += `
          <tr>
            <td class="px-3 py-2 text-center font-bold text-gray-400">Nota ${item.nota}</td>
            <td class="px-3 py-2 text-right text-xs text-gray-400">Inalcanzable</td>
          </tr>`;
      } else {
        html += `
          <tr>
            <td class="px-3 py-2 text-center font-bold text-gray-800">Nota ${item.nota}</td>
            <td class="px-3 py-2 text-right font-bold text-primary-600">${examenFinalRequerido}%</td>
          </tr>`;
      }
    });

    tablaResultados.innerHTML = html;
  }

  // Ejecutar primera evaluación
  calcular();
});
