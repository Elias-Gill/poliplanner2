document.addEventListener('DOMContentLoaded', () => {
  const toggleDirecto = document.getElementById('toggleDirecto');
  const panelDirecto = document.getElementById('panelDirecto');
  const panelDesglose = document.getElementById('panelDesglose');
  
  const inputDirecto = document.getElementById('inputPonderadoDirecto');
  
  // Parciales (Fijo 50%)
  const notaP1 = document.getElementById('notaP1');
  const notaP2 = document.getElementById('notaP2');
  
  // Laboratorio
  const checkAplicaLab = document.getElementById('checkAplicaLab');
  const wrapperPesoLab = document.getElementById('wrapperPesoLab');
  const contentLab = document.getElementById('contentLab');
  const notaLab = document.getElementById('notaLab');
  const pesoLab = document.getElementById('pesoLab');

  // Tareas
  const checkAplicaTareas = document.getElementById('checkAplicaTareas');
  const wrapperPesoTareas = document.getElementById('wrapperPesoTareas');
  const contentTareas = document.getElementById('contentTareas');
  const notaTareas = document.getElementById('notaTareas');
  const pesoTareas = document.getElementById('pesoTareas');

  // Elementos de Interfaz
  const valPromedio = document.getElementById('valPromedio');
  const badgeEstado = document.getElementById('badgeEstado');
  const msgEstado = document.getElementById('msgEstado');
  const tablaResultados = document.getElementById('tablaResultados');

  /**
   * Limita y autocorrige el valor de un input entre sus atributos min y max
   */
  function clampInput(input) {
    if (!input || input.value === '') return NaN;
    
    const min = input.hasAttribute('min') ? parseFloat(input.getAttribute('min')) : 0;
    const max = input.hasAttribute('max') ? parseFloat(input.getAttribute('max')) : 100;
    
    let val = parseFloat(input.value);
    if (isNaN(val)) return NaN;

    if (val > max) {
      val = max;
      input.value = max;
    } else if (val < min) {
      val = min;
      input.value = min;
    }
    return val;
  }

  // Modo Directo vs Desglose
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

  // Listener para Laboratorios
  checkAplicaLab.addEventListener('change', () => {
    if (checkAplicaLab.checked) {
      contentLab.classList.remove('hidden');
    } else {
      contentLab.classList.add('hidden');
      notaLab.value = '';
    }
    ajustarPesosProceso();
    calcular();
  });

  // Listener para Tareas
  checkAplicaTareas.addEventListener('change', () => {
    if (checkAplicaTareas.checked) {
      contentTareas.classList.remove('hidden');
    } else {
      contentTareas.classList.add('hidden');
      notaTareas.value = '';
    }
    ajustarPesosProceso();
    calcular();
  });

  // Sincronización de pesos (entre 0% y 50%)
  pesoLab.addEventListener('input', () => {
    let val = clampInput(pesoLab);
    if (isNaN(val)) val = 0;
    pesoTareas.value = 50 - val;
    calcular();
  });

  pesoTareas.addEventListener('input', () => {
    let val = clampInput(pesoTareas);
    if (isNaN(val)) val = 0;
    pesoLab.value = 50 - val;
    calcular();
  });

  // Gestión de visibilidad y reajuste de pesos
  function ajustarPesosProceso() {
    const labActivo = checkAplicaLab.checked;
    const tareasActivas = checkAplicaTareas.checked;

    if (labActivo && tareasActivas) {
      wrapperPesoLab.classList.remove('hidden');
      wrapperPesoTareas.classList.remove('hidden');
      
      const valLab = clampInput(pesoLab) || 0;
      const valTareas = clampInput(pesoTareas) || 0;

      if (valLab + valTareas !== 50) {
        pesoLab.value = 25;
        pesoTareas.value = 25;
      }
    } else if (labActivo) {
      wrapperPesoLab.classList.add('hidden');
      wrapperPesoTareas.classList.add('hidden');
      pesoLab.value = 50;
      pesoTareas.value = 0;
    } else if (tareasActivas) {
      wrapperPesoLab.classList.add('hidden');
      wrapperPesoTareas.classList.add('hidden');
      pesoLab.value = 0;
      pesoTareas.value = 50;
    } else {
      wrapperPesoLab.classList.add('hidden');
      wrapperPesoTareas.classList.add('hidden');
      pesoLab.value = 0;
      pesoTareas.value = 0;
    }
  }

  // Event listener general para todos los inputs numéricos (valida límites y recalcula)
  document.querySelectorAll('input[type="number"]').forEach(input => {
    input.addEventListener('input', () => {
      clampInput(input);
      calcular();
    });
  });

  function calcular() {
    let promedio = 0;

    if (toggleDirecto.checked) {
      promedio = clampInput(inputDirecto) || 0;
    } else {
      const pLab = checkAplicaLab.checked ? (clampInput(pesoLab) || 0) : 0;
      const nLab = checkAplicaLab.checked ? (clampInput(notaLab) || 0) : 0;

      const pTareas = checkAplicaTareas.checked ? (clampInput(pesoTareas) || 0) : 0;
      const nTareas = checkAplicaTareas.checked ? (clampInput(notaTareas) || 0) : 0;

      // Parciales (50% fijo)
      const np1 = clampInput(notaP1);
      const np2 = clampInput(notaP2);
      let promParciales = 0;

      if (!isNaN(np1) && !isNaN(np2)) {
        promParciales = (np1 + np2) / 2;
      } else if (!isNaN(np1)) {
        promParciales = np1;
      } else if (!isNaN(np2)) {
        promParciales = np2;
      }

      // PP = (PromParciales * 0.5) + (nLab * pLab/100) + (nTareas * pTareas/100)
      promedio = (promParciales * 0.50) + 
                 (nLab * (pLab / 100)) + 
                 (nTareas * (pTareas / 100));
    }

    promedio = Math.min(100, Math.max(0, promedio));
    valPromedio.textContent = `${promedio.toFixed(1)}%`;

    actualizarFirmas(promedio);
    actualizarTablaExamen(promedio);
  }

  function actualizarFirmas(promedio) {
    if (promedio < 60) {
      badgeEstado.className = "inline-block px-3 py-1 rounded-sm text-xs font-bold bg-red-50 text-red-900 mb-3";
      badgeEstado.textContent = "Sin Firma (PP < 60%)";
      msgEstado.textContent = "No alcanzás el 60% requerido. Aun podes rendir el 3er Parcial Recuperatorio.";
    } else {
      badgeEstado.className = "inline-block px-3 py-1 rounded-sm text-xs font-bold bg-green-100 text-green-700 mb-3";
      badgeEstado.textContent = "Habilitado (Firma Aprobada)";
      msgEstado.textContent = "";
    }
  }

  function actualizarTablaExamen(promedio) {
    if (promedio < 60) {
      tablaResultados.innerHTML = `
        <tr>
          <td colspan="2" class="px-3 py-4 text-center text-xs text-red-900 font-medium">
            No habilitado.
          </td>
        </tr>`;
      return;
    }

    // Escala oficial de notas según Puntuación Final (PF = 0.6*EF + 0.4*PP)
    const limitesNotas = [
      { nota: 2, minPF: 60 },
      { nota: 3, minPF: 71 },
      { nota: 4, minPF: 81 },
      { nota: 5, minPF: 91 }
    ];

    let html = '';

    limitesNotas.forEach(item => {
      let reqExamen = (item.minPF - (promedio * 0.4)) / 0.6;
      let examenFinalRequerido = Math.max(50, Math.ceil(reqExamen));

      if (examenFinalRequerido > 100) {
        html += `
          <tr>
            <td class="px-3 py-2 text-center font-bold text-gray-400">Nota ${item.nota}</td>
            <td class="px-3 py-2 text-right text-xs text-gray-400 italic">Inalcanzable</td>
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

  // Inicializar cálculo al cargar
  ajustarPesosProceso();
  calcular();
});
