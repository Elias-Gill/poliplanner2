document.addEventListener('DOMContentLoaded', function() {
    const MIN_EXAM = 50;
    const UMBRALES = [
        { nota: 5, min: 91 },
        { nota: 4, min: 81 },
        { nota: 3, min: 71 },
        { nota: 2, min: 60 },
    ];

    const inputPonderado = document.getElementById('inputPonderado');
    const btnPonderado = document.getElementById('btnPonderado');
    const btnComponentes = document.getElementById('btnComponentes');
    const resultadoTabla = document.getElementById('resultadoTabla');
    const resultados = document.getElementById('resultados');

    if (!btnPonderado || !btnComponentes || !resultadoTabla || !resultados) {
        console.error('No se encontraron algunos elementos necesarios');
        return;
    }

    function examenNecesario(ponderado, nota) {
        const u = UMBRALES.find(x => x.nota === nota);
        if (!u) return '-';
        let ex = (u.min - 0.4 * ponderado) / 0.6;
        ex = Math.max(ex, MIN_EXAM);
        ex = Math.floor(ex);
        return ex > 100 ? '-' : ex;
    }

    function renderTabla(ponderado) {
        resultadoTabla.innerHTML = '';
        for (let n = 5; n >= 2; n--) {
            const examenReq = examenNecesario(ponderado, n);
            const fila = document.createElement('tr');
            const esPar = (5 - n) % 2 === 0;
            fila.className = esPar ? 'bg-gray-50' : 'bg-white';
            fila.innerHTML = `
                <td class="px-4 py-3 text-gray-900 font-medium">${n}</td>
                <td class="px-4 py-3 text-gray-700">
                    ${examenReq === '-' ? '<span class="text-red-500 font-medium">—</span>' : '<span class="font-semibold text-gray-900">' + examenReq + '</span>'}
                </td>
            `;
            resultadoTabla.appendChild(fila);
        }
    }

    function mostrarResultados(seccionId) {
        const seccion = document.getElementById(seccionId);
        if (!seccion) return;

        seccion.parentNode.insertBefore(resultados, seccion.nextSibling);
        resultados.classList.remove('hidden');
        resultados.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }

    btnPonderado.addEventListener('click', function(event) {
        event.preventDefault();
        const v = parseFloat(inputPonderado.value);
        if (isNaN(v)) {
            alert('Por favor ingresa un número válido');
            return;
        }
        if (v < 40 || v > 100) {
            alert('El promedio debe estar entre 40 y 100');
            return;
        }
        renderTabla(v);
        mostrarResultados('seccionA');
    });

    btnComponentes.addEventListener('click', function(event) {
        event.preventDefault();
        const notaP1 = document.getElementById('notaP1');
        const notaP2 = document.getElementById('notaP2');
        const pesoParciales = document.getElementById('pesoParciales');
        const notaT = document.getElementById('notaT');
        const pesoT = document.getElementById('pesoT');
        const notaL = document.getElementById('notaL');
        const pesoL = document.getElementById('pesoL');

        const inputs = [notaP1, notaP2, pesoParciales, notaT, pesoT, notaL, pesoL];
        for (let i = 0; i < inputs.length; i++) {
            if (!inputs[i]) {
                alert('Error: No se encontraron todos los campos del formulario');
                return;
            }
        }

        const p1Val = parseFloat(notaP1.value);
        const p2Val = parseFloat(notaP2.value);
        const pesoParVal = parseFloat(pesoParciales.value);
        const tVal = parseFloat(notaT.value);
        const pesoTVal = parseFloat(pesoT.value);
        const lVal = parseFloat(notaL.value);
        const pesoLVal = parseFloat(pesoL.value);

        if (isNaN(p1Val) || isNaN(p2Val) || isNaN(pesoParVal) || 
            isNaN(tVal) || isNaN(pesoTVal) || isNaN(lVal) || isNaN(pesoLVal)) {
            alert('Por favor completa todos los campos');
            return;
        }

        const notas = [p1Val, p2Val, tVal, lVal];
        for (let i = 0; i < notas.length; i++) {
            if (notas[i] < 0 || notas[i] > 100) {
                alert('Las notas deben estar entre 0 y 100');
                return;
            }
        }

        const pesos = [pesoParVal, pesoTVal, pesoLVal];
        for (let i = 0; i < pesos.length; i++) {
            if (pesos[i] < 0 || pesos[i] > 100) {
                alert('Los pesos deben estar entre 0 y 100');
                return;
            }
        }

        const totalPesos = pesoParVal + pesoTVal + pesoLVal;
        if (Math.abs(totalPesos - 100) > 0.01) {
            alert('Los pesos deben sumar 100% (actual: ' + totalPesos.toFixed(1) + '%)');
            return;
        }

        const promedioParciales = (p1Val + p2Val) / 2;
        const ponderado = (promedioParciales * pesoParVal + tVal * pesoTVal + lVal * pesoLVal) / 100;

        renderTabla(ponderado);
        mostrarResultados('seccionB');
    });

    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function(event) {
            event.preventDefault();
            return false;
        });
    });
});
