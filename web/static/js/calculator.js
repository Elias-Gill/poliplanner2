// calculator.js
document.addEventListener('DOMContentLoaded', function() {
    console.log('Script cargado correctamente'); // Para debugging
    
    const MIN_EXAM = 50;
    const UMBRALES = [
        { nota: 5, min: 91 },
        { nota: 4, min: 81 },
        { nota: 3, min: 71 },
        { nota: 2, min: 60 },
    ];

    // Referencias a inputs y botones
    const inputPonderado = document.getElementById('inputPonderado');
    const btnPonderado = document.getElementById('btnPonderado');
    const btnComponentes = document.getElementById('btnComponentes');
    const resultadoTabla = document.getElementById('resultadoTabla');

    // Verificar que los elementos existen
    console.log('btnPonderado:', btnPonderado);
    console.log('btnComponentes:', btnComponentes);
    console.log('resultadoTabla:', resultadoTabla);

    if (!btnPonderado || !btnComponentes || !resultadoTabla) {
        console.error('No se encontraron algunos elementos necesarios');
        return;
    }

    // Función que calcula examen necesario
    function examenNecesario(ponderado, nota) {
        console.log('Calculando para nota:', nota, 'con ponderado:', ponderado); // Debug
        
        const u = UMBRALES.find(x => x.nota === nota);
        if (!u) return '-';
        
        // Fórmula: (nota_minima - 0.4 * ponderado) / 0.6
        let ex = (u.min - 0.4 * ponderado) / 0.6;
        ex = Math.max(ex, MIN_EXAM);
        ex = Math.floor(ex);
        
        console.log('Resultado examen:', ex); // Debug
        return ex > 100 ? '-' : ex;
    }

    // Renderiza la tabla
    function renderTabla(ponderado) {
        console.log('Renderizando tabla con ponderado:', ponderado); // Debug
        
        if (!resultadoTabla) {
            console.error('resultadoTabla no encontrado');
            return;
        }
        
        // Limpiar tabla
        resultadoTabla.innerHTML = '';
        
        // Generar filas
        for (let n = 5; n >= 2; n--) {
            const examenReq = examenNecesario(ponderado, n);
            const fila = document.createElement('tr');
            fila.innerHTML = `
                <td class="border px-3 py-1">${n}</td>
                <td class="border px-3 py-1">${examenReq}</td>
            `;
            resultadoTabla.appendChild(fila);
        }
        
        // Scroll a resultados
        const resultados = document.getElementById('resultados');
        if (resultados) {
            resultados.scrollIntoView({ behavior: 'smooth' });
        }
    }

    // Opción A: promedio ponderado conocido
    btnPonderado.addEventListener('click', function(event) {
        event.preventDefault(); // Prevenir cualquier comportamiento por defecto
        console.log('Click en btnPonderado'); // Debug
        
        if (!inputPonderado) {
            alert('Error: Campo de entrada no encontrado');
            return;
        }
        
        const v = parseFloat(inputPonderado.value);
        console.log('Valor ingresado:', v); // Debug
        
        if (isNaN(v)) {
            alert('Por favor ingresa un número válido');
            return;
        }
        
        if (v < 40 || v > 100) {
            alert('El promedio debe estar entre 40 y 100');
            return;
        }
        
        renderTabla(v);
    });

    // Opción B: armar promedio ponderado desde componentes
    btnComponentes.addEventListener('click', function(event) {
        event.preventDefault(); // Prevenir cualquier comportamiento por defecto
        console.log('Click en btnComponentes'); // Debug
        
        // Obtener referencias a los inputs
        const notaP1 = document.getElementById('notaP1');
        const notaP2 = document.getElementById('notaP2');
        const pesoParciales = document.getElementById('pesoParciales');
        const notaT = document.getElementById('notaT');
        const pesoT = document.getElementById('pesoT');
        const notaL = document.getElementById('notaL');
        const pesoL = document.getElementById('pesoL');
        
        // Verificar que todos los inputs existen
        const inputs = [notaP1, notaP2, pesoParciales, notaT, pesoT, notaL, pesoL];
        for (let i = 0; i < inputs.length; i++) {
            if (!inputs[i]) {
                alert('Error: No se encontraron todos los campos del formulario');
                return;
            }
        }
        
        // Obtener valores
        const p1Val = parseFloat(notaP1.value);
        const p2Val = parseFloat(notaP2.value);
        const pesoParVal = parseFloat(pesoParciales.value);
        const tVal = parseFloat(notaT.value);
        const pesoTVal = parseFloat(pesoT.value);
        const lVal = parseFloat(notaL.value);
        const pesoLVal = parseFloat(pesoL.value);
        
        console.log('Valores:', { p1Val, p2Val, pesoParVal, tVal, pesoTVal, lVal, pesoLVal }); // Debug
        
        // Validar que todos los campos tengan valor
        if (isNaN(p1Val) || isNaN(p2Val) || isNaN(pesoParVal) || 
            isNaN(tVal) || isNaN(pesoTVal) || isNaN(lVal) || isNaN(pesoLVal)) {
            alert('Por favor completa todos los campos');
            return;
        }
        
        // Validar rangos de notas
        const notas = [p1Val, p2Val, tVal, lVal];
        for (let i = 0; i < notas.length; i++) {
            if (notas[i] < 0 || notas[i] > 100) {
                alert('Las notas deben estar entre 0 y 100');
                return;
            }
        }
        
        // Validar rangos de pesos
        const pesos = [pesoParVal, pesoTVal, pesoLVal];
        for (let i = 0; i < pesos.length; i++) {
            if (pesos[i] < 0 || pesos[i] > 100) {
                alert('Los pesos deben estar entre 0 y 100');
                return;
            }
        }
        
        // Validar suma de pesos
        const totalPesos = pesoParVal + pesoTVal + pesoLVal;
        console.log('Total pesos:', totalPesos); // Debug
        
        // Permitir pequeño margen de error por redondeo
        if (Math.abs(totalPesos - 100) > 0.01) {
            alert('Los pesos deben sumar 100% (actual: ' + totalPesos.toFixed(1) + '%)');
            return;
        }
        
        // Calcular promedio ponderado
        const promedioParciales = (p1Val + p2Val) / 2;
        const ponderado = (promedioParciales * pesoParVal + tVal * pesoTVal + lVal * pesoLVal) / 100;
        
        console.log('Ponderado calculado:', ponderado); // Debug
        
        renderTabla(ponderado);
    });

    // También agregar event listeners a los formularios para prevenir submit
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function(event) {
            event.preventDefault();
            return false;
        });
    });
    
    console.log('Event listeners configurados correctamente');
});
