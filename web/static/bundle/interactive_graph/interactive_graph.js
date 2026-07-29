// ================= CONSTANTES DE DIMENSIONES =================

const DIMENSIONS = {
  SEMESTER_WIDTH: 300, // Ancho asignado por semestre
  NODE_START_Y: 80, // Posición vertical inicial de los nodos
  NODE_SPACING_Y: 90, // Separación vertical entre nodos
  NODE_MARGIN: 10, // Margen interno de los nodos
  NODE_MAX_WIDTH: 220, // ≈ 28–32 caracteres dependiendo de la fuente
};

// ================= CONSTANTES DE COLORES =================

function getColors() {
  const isDark = document.documentElement.classList.contains("dark-mode");
  if (isDark) {
    return {
      NODE_TEXT:            "#e6edf5", // primary-100
      NODE_BG_ODD:          "#1f3550", // primary-800
      NODE_BG_EVEN:         "#2f4c6d", // primary-700
      NODE_BORDER:          "#3f638a", // primary-600
      NODE_HIGHLIGHT_BG:    "#3f638a", // primary-600
      NODE_HIGHLIGHT_BORDER:"#7fa3c4", // primary-400
      NODE_CHILD_BG:        "#557fab", // primary-500
      NODE_CHILD_BORDER:    "#a9c1da", // primary-300
      NODE_PARENT_BG:       "#2f4c6d", // primary-700
      NODE_PARENT_BORDER:   "#557fab", // primary-500
      EDGE_NORMAL:          "#3f638a", // primary-600
      EDGE_OPACITY:         0.85,
    };
  }
  return {
    NODE_TEXT:            "#142536", // primary-900
    NODE_BG_ODD:          "#e6edf5", // primary-100
    NODE_BG_EVEN:         "#cddaea", // primary-200
    NODE_BORDER:          "#a9c1da", // primary-300
    NODE_HIGHLIGHT_BG:    "#a9c1da", // primary-300
    NODE_HIGHLIGHT_BORDER:"#557fab", // primary-500
    NODE_CHILD_BG:        "#7fa3c4", // primary-400
    NODE_CHILD_BORDER:    "#3f638a", // primary-600
    NODE_PARENT_BG:       "#557fab", // primary-500
    NODE_PARENT_BORDER:   "#2f4c6d", // primary-700
    EDGE_NORMAL:          "#7fa3c4", // primary-400
    EDGE_OPACITY:         0.85,
  };
}

// ================= CONSTANTES DE FUENTES =================

const FONTS = {
  NODE_SIZE: 14,
  NODE_FACE: "Inter, Arial, sans-serif",
};

// ============== CONFIGURACIÓN BASE PARA NODOS Y ARISTAS ==============

function getNodeCommonConfig(COLORS, DIMENSIONS, FONTS) {
  return {
    shape: "box",
    fixed: true,
    physics: false,
    margin: DIMENSIONS.NODE_MARGIN,
    font: {
      size: FONTS.NODE_SIZE,
      face: FONTS.NODE_FACE,
      color: COLORS.NODE_TEXT,
      multi: true,
    },
    borderWidth: 1,
    widthConstraint: {
      maximum: DIMENSIONS.NODE_MAX_WIDTH,
    },
  };
}

function getEdgesCommonConfig(COLORS) {
  return {
    arrows: {
      to: {
        enabled: true,
        scaleFactor: 0.8,
        type: "arrow",
      },
    },
    color: {
      color: COLORS.EDGE_NORMAL,
      opacity: COLORS.EDGE_OPACITY,
      highlight: COLORS.EDGE_NORMAL,
      hover: COLORS.EDGE_NORMAL,
      inherit: false,
    },
    width: 1,
    hoverWidth: 1,
    selectionWidth: 0,
  };
}

const SEMESTERS_TITLE_STYLE = {
  y: 10,
  fixed: true,
  physics: false,
  font: {
    size: 28,
    bold: true,
    color: "#3f638a", // primary-600
  },
  shape: "text",
};

// ================= ESTILOS DE NODOS (para hover / parent / child) =================

function getNodeStyles(COLORS, FONTS) {
  const base = {
    font: {
      color: COLORS.NODE_TEXT,
      bold: false,
      size: FONTS.NODE_SIZE,
    },
    borderWidth: 1,
    opacity: 1,
  };

  return {
    NODE_STYLE_BASE: base,
    NODE_STYLE_DEFAULT_FADED: { opacity: 0.12 },
    NODE_STYLE_HOVER: {
      color: {
        background: COLORS.NODE_HIGHLIGHT_BG,
        border: COLORS.NODE_HIGHLIGHT_BORDER,
      },
      font: {
        ...base.font,
        bold: true,
        size: FONTS.NODE_SIZE + 2,
      },
      opacity: 1,
    },
    NODE_STYLE_PARENT: {
      color: {
        background: COLORS.NODE_PARENT_BG,
        border: COLORS.NODE_PARENT_BORDER,
      },
      font: { ...base.font, bold: true },
      opacity: 1,
    },
    NODE_STYLE_CHILD: {
      color: {
        background: COLORS.NODE_CHILD_BG,
        border: COLORS.NODE_CHILD_BORDER,
      },
      font: { ...base.font, bold: true },
      opacity: 1,
    },
  };
}

// ================= VARIABLES GLOBALES =================

let allSubjects = {};
let nodes;
let edges;
let network;
let semestersCount = 0;

// ===================== PUBLIC API =====================

function prepareGraphData(data) {
  semestersCount = data.career.totalSemesters;

  const subjectsBySemester = _groupSubjectsBySemester(
    data.subjects,
    semestersCount,
  );
  const semesterTitles = _createSemesterTitles();

  const created = _createSubjectNodes(subjectsBySemester);
  allSubjects = created.allSubjects;

  nodes = new vis.DataSet([...semesterTitles, ...created.nodesArray]);
  edges = new vis.DataSet(_createEdges(allSubjects));
}

function renderGraph(elementId) {
  const element = document.getElementById(elementId);
  if (!element) return;

  const options = _createVisOptions();

  network = new vis.Network(element, { nodes, edges }, options);
  _setupNetworkEvents(network, allSubjects, nodes, edges);
  network.fit();
}

// ================= FUNCIONES AUXILIARES =================

function _createSubjectNodes(semesters) {
  const COLORS = getColors();
  const nodeConfig = getNodeCommonConfig(COLORS, DIMENSIONS, FONTS);
  const allSubjectsLocal = {};
  const nodesArray = [];

  for (let sem = 1; sem <= semestersCount; sem++) {
    const subjects = semesters[sem];
    const ids = Object.keys(subjects).sort((a, b) =>
      subjects[a].name.localeCompare(subjects[b].name),
    );

    const xBase = (sem - 1) * DIMENSIONS.SEMESTER_WIDTH;

    ids.forEach((id, i) => {
      const subject = subjects[id];

      allSubjectsLocal[id] = {
        ...subject,
        sem,
        pre: subject.prerequisites || [],
      };

      nodesArray.push({
        id,
        label: subject.name,
        group: `sem${sem}`,
        x: xBase,
        y: DIMENSIONS.NODE_START_Y + i * DIMENSIONS.NODE_SPACING_Y,
        ...nodeConfig,
      });
    });
  }

  return { allSubjects: allSubjectsLocal, nodesArray };
}

function _createEdges(allSubjectsLocal) {
  const COLORS = getColors();
  const edgeConfig = getEdgesCommonConfig(COLORS);
  const edgesArray = [];

  for (const [id, subject] of Object.entries(allSubjectsLocal)) {
    for (const pre of subject.pre) {
      edgesArray.push({
        from: pre,
        to: id,
        ...edgeConfig,
      });
    }
  }

  return edgesArray;
}

function _createSemesterTitles() {
  const titles = [];

  for (let i = 1; i <= semestersCount; i++) {
    titles.push({
      id: `title-sem-${i}`,
      label: `Semestre ${i}`,
      x: (i - 1) * DIMENSIONS.SEMESTER_WIDTH,
      ...SEMESTERS_TITLE_STYLE,
    });
  }

  return titles;
}

function _createVisOptions() {
  const COLORS = getColors();
  const groups = {};

  for (let i = 1; i <= semestersCount; i++) {
    const isEven = i % 2 === 0;
    groups[`sem${i}`] = {
      color: {
        background: isEven ? COLORS.NODE_BG_EVEN : COLORS.NODE_BG_ODD,
        border: COLORS.NODE_BORDER,
        highlight: {
          background: COLORS.NODE_HIGHLIGHT_BG,
          border: COLORS.NODE_HIGHLIGHT_BORDER,
        },
        hover: {
          background: COLORS.NODE_HIGHLIGHT_BG,
          border: COLORS.NODE_HIGHLIGHT_BORDER,
        },
      },
      font: { color: COLORS.NODE_TEXT },
    };
  }

  return {
    interaction: {
      hover: true,
      dragView: true,
      zoomView: true,
    },
    physics: false,
    layout: { hierarchical: false },
    groups,
    edges: {
      arrows: {
        to: {
          enabled: true,
        },
      },
      color: {
        color: COLORS.EDGE_NORMAL,
        highlight: COLORS.EDGE_NORMAL,
        hover: COLORS.EDGE_NORMAL,
        inherit: false,
      },
    },
  };
}

// ================= EVENTOS =================

function _setupNetworkEvents(net, allSubs, nodesDs, edgesDs) {
  const isSemesterTitle = (nodeId) => String(nodeId).startsWith("title-sem-");

  const isTouch = "ontouchstart" in window || navigator.maxTouchPoints > 0;

  function showRelations(id) {
    const COLORS = getColors();
    const styles = getNodeStyles(COLORS, FONTS);
    const parents = _getAllParents(id, allSubs);
    const children = _getDirectChildren(id, allSubs);
    const relatedNodes = new Set([id, ...parents, ...children]);

    const nodeUpdates = nodesDs.map((node) => {
      if (String(node.id).startsWith("title-sem-")) return { id: node.id };

      if (node.id === id) return { id: node.id, ...styles.NODE_STYLE_HOVER };
      if (parents.has(node.id)) return { id: node.id, ...styles.NODE_STYLE_PARENT };
      if (children.has(node.id)) return { id: node.id, ...styles.NODE_STYLE_CHILD };

      return { id: node.id, ...styles.NODE_STYLE_DEFAULT_FADED };
    });

    const edgeUpdates = edgesDs.map((edge) => {
      const fromRelated = relatedNodes.has(edge.from);
      const toRelated = relatedNodes.has(edge.to);
      return {
        id: edge.id,
        hidden: !(fromRelated && toRelated),
      };
    });

    nodesDs.update(nodeUpdates.filter((u) => Object.keys(u).length > 1));
    edgesDs.update(edgeUpdates);
  }

  function resetGraphStyles() {
    const COLORS = getColors();
    const styles = getNodeStyles(COLORS, FONTS);
    const allNodeUpdates = nodesDs.map((node) => {
      if (isSemesterTitle(node.id)) return { id: node.id };
      return { id: node.id, ...styles.NODE_STYLE_BASE, opacity: 1 };
    });

    const allEdgeUpdates = edgesDs.map((edge) => ({
      id: edge.id,
      hidden: false,
    }));

    nodesDs.update(allNodeUpdates.filter((u) => Object.keys(u).length > 1));
    edgesDs.update(allEdgeUpdates);
  }

  // ================= DESKTOP =================
  if (!isTouch) {
    net.on("hoverNode", (params) => {
      const id = params.node;
      if (isSemesterTitle(id)) return;
      showRelations(id);
    });

    net.on("blurNode", resetGraphStyles);

    net.on("click", (params) => {
      if (!params.nodes.length) return;
      const id = params.nodes[0];
      if (isSemesterTitle(id)) return;

      _openModal(id);
    });
  }

  // ================= MOBILE =================
  if (isTouch) {
    net.on("selectNode", (params) => {
      if (!params.nodes.length) return;
      const id = params.nodes[0];
      if (isSemesterTitle(id)) return;

      showRelations(id);
    });

    net.on("deselectNode", resetGraphStyles);

    // long press → modal
    net.on("hold", (params) => {
      if (!params.nodes.length) return;
      const id = params.nodes[0];
      if (isSemesterTitle(id)) return;

      _openModal(id);
    });
  }
}

// ================= FUNCIONES PARA NODOS RELACIONADOS =================

// Búsqueda recursiva para obtener todos los pre-requisitos de un nodo. Utiliza "depth first
// search" para buscar los ancestros del nodo seleccionado.
function _getAllParents(id, allSubs, visited = new Set()) {
  if (visited.has(id)) return new Set();
  visited.add(id);

  const parents = new Set();

  const subject = allSubs[id];
  if (!subject || !subject.pre) return parents;

  for (const preId of subject.pre) {
    parents.add(preId);
    const grandParents = _getAllParents(preId, allSubs, visited);
    for (const gp of grandParents) parents.add(gp);
  }

  return parents;
}

// Obtiene todos los nodos que tengan como pre-requisito al nodo seleccionado
function _getDirectChildren(id, allSubs) {
  const children = new Set();

  for (const [sid, subj] of Object.entries(allSubs)) {
    if (subj.pre.includes(id)) children.add(sid);
  }

  return children;
}

// Agrupa las materias por semestre
function _groupSubjectsBySemester(subjects, count) {
  const grouped = {};
  for (let i = 1; i <= count; i++) grouped[i] = {};

  for (const [id, subject] of Object.entries(subjects)) {
    if (grouped[subject.semester]) {
      grouped[subject.semester][id] = subject;
    }
  }

  return grouped;
}

function _openModal(id) {
  const subject = allSubjects[id];
  if (!subject) return;

  const existing = document.getElementById("subject-modal-overlay");
  if (existing) existing.remove();

  // ================= OVERLAY =================
  const overlay = document.createElement("div");
  overlay.id = "subject-modal-overlay";

  Object.assign(overlay.style, {
    position: "fixed",
    inset: "0",
    background: "rgba(0,0,0,0.45)",
    backdropFilter: "blur(4px)",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    zIndex: "9999",
    animation: "fadeIn 0.2s ease",
  });

  // ================= MODAL =================
  const modal = document.createElement("div");

  Object.assign(modal.style, {
    width: "min(480px, 92vw)",
    maxHeight: "85vh",
    background: "var(--color-surface, #ffffff)",
    border: "1px solid var(--color-border, #e5e7eb)",
    borderRadius: "4px",
    boxShadow: "0 20px 40px rgba(0,0,0,0.2)",
    fontFamily: "Inter, Arial, sans-serif",
    position: "relative",
    display: "flex",
    flexDirection: "column",
    overflow: "hidden",
  });

  modal.addEventListener("click", (e) => e.stopPropagation());

  // ================= HEADER =================
  const header = document.createElement("div");
  Object.assign(header.style, {
    padding: "20px 24px",
    background: "linear-gradient(135deg, var(--color-primary-600, #3b82f6) 0%, var(--color-primary-800, #1e40af) 100%)",
    color: "#ffffff",
  });

  const title = document.createElement("h2");
  title.textContent = subject.name;
  Object.assign(title.style, {
    margin: "0",
    fontSize: "20px",
    fontWeight: "600",
    lineHeight: "1.3",
    paddingRight: "24px",
  });
  header.appendChild(title);

  // ================= BOTÓN CERRAR =================
  const close = document.createElement("button");
  close.innerHTML = "&times;";
  Object.assign(close.style, {
    position: "absolute",
    top: "16px",
    right: "16px",
    border: "none",
    background: "rgba(255,255,255,0.2)",
    borderRadius: "50%",
    width: "28px",
    height: "28px",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontSize: "18px",
    cursor: "pointer",
    color: "#ffffff",
    transition: "background 0.2s",
  });
  close.onmouseover = () => close.style.background = "rgba(255,255,255,0.3)";
  close.onmouseout = () => close.style.background = "rgba(255,255,255,0.2)";
  close.onclick = () => overlay.remove();

  // ================= CONTENIDO (BODY) =================
  const body = document.createElement("div");
  Object.assign(body.style, {
    padding: "24px",
    overflowY: "auto",
  });

  // SVG icons
  const SVG = {
    book: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="#557fab" stroke-width="2" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25"/></svg>`,
    star: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="#557fab" stroke-width="2" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557l-4.204-3.602a.562.562 0 0 1 .321-.988l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z"/></svg>`,
    clock: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="#557fab" stroke-width="2" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>`,
    info: `<svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" style="vertical-align:middle;flex-shrink:0"><path stroke-linecap="round" stroke-linejoin="round" d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"/></svg>`,
  };

  function createStat(label, value, iconSvg) {
    const col = document.createElement("div");
    Object.assign(col.style, {
      background: "var(--color-bg, #e6edf5)",
      border: "1px solid var(--color-border, #a9c1da)",
      padding: "12px 10px",
      borderRadius: "12px",
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      textAlign: "center",
      gap: "4px",
    });

    const icon = document.createElement("div");
    icon.innerHTML = iconSvg;

    const v = document.createElement("div");
    v.textContent = value ?? "-";
    Object.assign(v.style, {
      fontWeight: "700",
      color: "var(--color-text, #142536)",
      fontSize: "17px",
      lineHeight: "1",
    });

    const l = document.createElement("div");
    l.textContent = label;
    Object.assign(l.style, {
      fontSize: "11px",
      color: "var(--color-text-muted, #557fab)",
      fontWeight: "600",
      textTransform: "uppercase",
      letterSpacing: "0.05em",
    });

    col.appendChild(icon);
    col.appendChild(v);
    col.appendChild(l);
    return col;
  }

  const statsGrid = document.createElement("div");
  Object.assign(statsGrid.style, {
    display: "grid",
    gridTemplateColumns: "repeat(3, 1fr)",
    gap: "12px",
    marginBottom: "20px",
  });

  statsGrid.appendChild(createStat("Semestre", subject.semester, SVG.book));
  statsGrid.appendChild(createStat("Créditos", subject.credits, SVG.star));
  statsGrid.appendChild(createStat("Hrs/Sem", subject.weekly_hours, SVG.clock));

  const desc = document.createElement("div");
  Object.assign(desc.style, {
    fontSize: "14px",
    lineHeight: "1.6",
    color: "var(--color-text-muted, #4b5563)",
    background: "var(--color-bg, #f9fafb)",
    borderLeft: "4px solid #557fab",
    padding: "12px 16px",
    borderRadius: "0 4px 4px 0",
    marginBottom: "20px",
  });
  desc.innerHTML = `<strong>Descripción:</strong><br/>${subject.desc ?? "Sin descripción disponible."}`;

  // ================= PRE-REQUISITOS =================
  const prereqSection = document.createElement("div");
  Object.assign(prereqSection.style, {
    marginTop: "0",
    paddingTop: "20px",
    borderTop: "1px solid var(--color-border, #a9c1da)",
  });

  const prereqTitle = document.createElement("div");
  prereqTitle.textContent = "Pre-requisitos";
  Object.assign(prereqTitle.style, {
    fontSize: "11px",
    fontWeight: "700",
    textTransform: "uppercase",
    letterSpacing: "0.08em",
    color: "var(--color-text-muted, #6b7280)",
    marginBottom: "10px",
  });
  prereqSection.appendChild(prereqTitle);

  const prereqList = subject.pre || [];

  if (prereqList.length === 0) {
    const none = document.createElement("p");
    none.textContent = "Esta materia no tiene pre-requisitos.";
    Object.assign(none.style, {
      fontSize: "13px",
      color: "var(--color-text-muted, #9ca3af)",
      fontStyle: "italic",
      margin: "0",
    });
    prereqSection.appendChild(none);
  } else {
    const chipsWrapper = document.createElement("div");
    Object.assign(chipsWrapper.style, {
      display: "flex",
      flexWrap: "wrap",
      gap: "8px",
    });

    prereqList.forEach((preId) => {
      const prereqSubject = allSubjects[preId];
      const chip = document.createElement("button");
      chip.textContent = prereqSubject ? prereqSubject.name : preId;
      Object.assign(chip.style, {
        display: "inline-flex",
        alignItems: "center",
        gap: "4px",
        padding: "5px 12px",
        borderRadius: "999px",
        border: "1px solid var(--color-border, #a9c1da)",
        background: "var(--color-bg, #e6edf5)",
        color: "var(--color-primary-500, #557fab)",
        fontSize: "12px",
        fontWeight: "600",
        cursor: prereqSubject ? "pointer" : "default",
        transition: "background 0.15s, transform 0.1s",
        fontFamily: "Inter, Arial, sans-serif",
      });
      chip.title = prereqSubject
        ? `Ver ${prereqSubject.name}`
        : preId;

      if (prereqSubject) {
        chip.innerHTML = `↗ ${prereqSubject.name}`;
        chip.addEventListener("mouseenter", () => {
          chip.style.background = "#cddaea";
          chip.style.transform = "translateY(-1px)";
        });
        chip.addEventListener("mouseleave", () => {
          chip.style.background = "var(--color-bg, #e6edf5)";
          chip.style.transform = "translateY(0)";
        });
        chip.addEventListener("click", (e) => {
          e.stopPropagation();
          overlay.remove();
          _openModal(preId);
        });
      }

      chipsWrapper.appendChild(chip);
    });

    prereqSection.appendChild(chipsWrapper);
  }

  body.appendChild(desc);
  body.appendChild(prereqSection);

  if (subject.required_credits) {
    const req = document.createElement("div");
    Object.assign(req.style, {
      marginTop: "20px",
      fontSize: "13px",
      color: "var(--color-text, #2f4c6d)",
      background: "var(--color-bg, #e6edf5)",
      border: "1px solid var(--color-border, #a9c1da)",
      padding: "10px 14px",
      borderRadius: "10px",
      display: "flex",
      alignItems: "center",
      gap: "8px",
      lineHeight: "1.5",
    });
    req.innerHTML = `${SVG.info} <span><strong>Nota:</strong> Requiere ${subject.required_credits} créditos aprobados.</span>`;
    body.appendChild(req);
  }

  modal.appendChild(close);
  modal.appendChild(header);
  body.appendChild(statsGrid);
  modal.appendChild(body);
  overlay.appendChild(modal);
  document.body.appendChild(overlay);

  // cerrar al hacer click fuera
  overlay.addEventListener("click", () => overlay.remove());

  // cerrar con ESC
  const escHandler = (e) => {
    if (e.key === "Escape") {
      overlay.remove();
      document.removeEventListener("keydown", escHandler);
    }
  };

  document.addEventListener("keydown", escHandler);
}
