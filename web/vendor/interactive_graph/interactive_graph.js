// ================= CONSTANTES DE DIMENSIONES =================

const DIMENSIONS = {
  SEMESTER_WIDTH: 340,
  NODE_START_Y: 65,
  NODE_SPACING_Y: 100,
  NODE_MAX_WIDTH: 220,
  NODE_HEIGHT: 38,
  LINE_HEIGHT: 17,
};

// ================= CONSTANTES DE COLORES =================

function getColors() {
  const isDark = document.documentElement.classList.contains("dark-mode");
  if (isDark) {
    return {
      NODE_TEXT:             "#e6edf5",
      NODE_BG_ODD:          "#142536",
      NODE_BG_EVEN:         "#2f4c6d",
      NODE_BORDER:          "#3f638a",
      NODE_HIGHLIGHT_BG:    "#3f638a",
      NODE_HIGHLIGHT_BORDER:"#7fa3c4",
      NODE_CHILD_BG:        "#557fab",
      NODE_CHILD_BORDER:    "#a9c1da",
      NODE_PARENT_BG:       "#2f4c6d",
      NODE_PARENT_BORDER:   "#557fab",
      EDGE_NORMAL:          "#3f638a",
      EDGE_OPACITY:         0.85,
      TITLE_COLOR:          "#60a5fa",
    };
  }
  return {
    NODE_TEXT:             "#142536",
    NODE_BG_ODD:          "#f0f4f8",
    NODE_BG_EVEN:         "#cddaea",
    NODE_BORDER:          "#a9c1da",
    NODE_HIGHLIGHT_BG:    "#a9c1da",
    NODE_HIGHLIGHT_BORDER:"#557fab",
    NODE_CHILD_BG:        "#7fa3c4",
    NODE_CHILD_BORDER:    "#3f638a",
    NODE_PARENT_BG:       "#557fab",
    NODE_PARENT_BORDER:   "#2f4c6d",
    EDGE_NORMAL:          "#7fa3c4",
    EDGE_OPACITY:         0.85,
    TITLE_COLOR:          "#3f638a",
  };
}

const FONTS = {
  NODE_SIZE: 13,
  NODE_FACE: "Inter, Arial, sans-serif",
};

// ================= VARIABLES GLOBALES =================

let allSubjects = {};
let nodePositions = {};
let edgeList = [];
let semestersCount = 0;
let _zoomInstance = null;
let _graphWidth = 0;
let _graphHeight = 0;

// Estado de selección táctil en mobile
let _selectedMobileNodeId = null;

// ===================== PUBLIC API =====================

function prepareGraphData(data) {
  semestersCount = data.career.totalSemesters;

  const subjectsBySemester = _groupSubjectsBySemester(
    data.subjects,
    semestersCount,
  );

  const created = _createSubjectNodes(subjectsBySemester);
  allSubjects = created.allSubjects;
  nodePositions = created.nodePositions;
  edgeList = _createEdges(allSubjects);
}

function renderGraph(elementId) {
  const element = document.getElementById(elementId);
  if (!element) return;
  element.innerHTML = "";

  _selectedMobileNodeId = null; // Reiniciar selección

  const COLORS = getColors();
  const isTouch = "ontouchstart" in window || navigator.maxTouchPoints > 0;

  if (semestersCount === 0 || Object.keys(nodePositions).length === 0) return;

  let maxY = 0;
  for (const pos of Object.values(nodePositions)) {
    const bottom = pos.y + (pos.height || DIMENSIONS.NODE_HEIGHT);
    if (bottom > maxY) maxY = bottom;
  }

  const padX = 20;
  const padY = 10;
  const titleArea = 50;
  _graphWidth = semestersCount * DIMENSIONS.SEMESTER_WIDTH + padX * 2;
  _graphHeight = Math.max(maxY + padY, titleArea + 60);

  // ── SVG (sin viewBox — el zoom de D3 escala directamente en CSS) ──
  const svg = d3.select(element)
    .append("svg")
    .attr("width", "100%")
    .attr("height", "100%")
    .style("width", "100%")
    .style("height", "100%")
    .style("display", "block");

  // ── Defs (arrow marker) ──
  svg.append("defs")
    .append("marker")
    .attr("id", "arrow")
    .attr("viewBox", "0 -4 8 8")
    .attr("refX", 14)
    .attr("refY", 0)
    .attr("markerWidth", 6)
    .attr("markerHeight", 6)
    .attr("orient", "auto")
    .append("path")
    .attr("d", "M0,-4L8,0L0,4")
    .attr("fill", COLORS.EDGE_NORMAL);

  // ── Zoom ──
  _zoomInstance = d3.zoom()
    .scaleExtent([0.1, 20])
    .on("zoom", (event) => {
      container.attr("transform", event.transform);
    });

  svg.call(_zoomInstance);

  const container = svg.append("g");

  // ── Semester title columns ──
  for (let sem = 1; sem <= semestersCount; sem++) {
    const x = (sem - 1) * DIMENSIONS.SEMESTER_WIDTH + 50;
    const cx = x + DIMENSIONS.SEMESTER_WIDTH / 2;

    container.append("text")
      .attr("x", cx)
      .attr("y", 32)
      .attr("text-anchor", "middle")
      .attr("font-size", 22)
      .attr("font-weight", "700")
      .attr("font-family", FONTS.NODE_FACE)
      .attr("fill", COLORS.TITLE_COLOR)
      .attr("class", "semester-title")
      .attr("pointer-events", isTouch ? "none" : null)
      .text(`Semestre ${sem}`);
  }

  // ── Edges ──
  const edgePaths = container.append("g")
    .selectAll("path")
    .data(edgeList)
    .enter()
    .append("path")
    .attr("d", (d) => {
      const src = nodePositions[d.from];
      const tgt = nodePositions[d.to];
      if (!src || !tgt) return "";
      const sx = src.x + src.width;
      const sy = src.y + src.height / 2;
      const tx = tgt.x;
      const ty = tgt.y + tgt.height / 2;
      const cx = (sx + tx) / 2;
      return `M${sx},${sy}C${cx},${sy} ${cx},${ty} ${tx},${ty}`;
    })
    .attr("fill", "none")
    .attr("stroke", COLORS.EDGE_NORMAL)
    .attr("stroke-width", 1.2)
    .attr("opacity", COLORS.EDGE_OPACITY)
    .attr("marker-end", "url(#arrow)");

  // ── Nodes ──
  const nodeData = Object.entries(nodePositions).map(([id, pos]) => ({
    id,
    ...pos,
    subject: allSubjects[id],
  }));

  const nodeGroups = container.append("g")
    .selectAll("g")
    .data(nodeData)
    .enter()
    .append("g")
    .attr("data-id", (d) => d.id);

  // tooltip con nombre completo
  nodeGroups.append("title")
    .text((d) => d.subject ? d.subject.name : d.id);

  function getNodeBg(d) {
    return d.subject && d.subject.sem % 2 === 0 ? COLORS.NODE_BG_EVEN : COLORS.NODE_BG_ODD;
  }

  // rect
  nodeGroups.append("rect")
    .attr("x", (d) => d.x)
    .attr("y", (d) => d.y)
    .attr("width", (d) => d.width)
    .attr("height", (d) => d.height)
    .attr("rx", 6)
    .attr("ry", 6)
    .attr("fill", getNodeBg)
    .attr("stroke", COLORS.NODE_BORDER)
    .attr("stroke-width", 1)
    .attr("cursor", "pointer")
    .attr("class", "node-rect");

  // label (multi-line via tspan)
  nodeGroups.each(function (d) {
    const lines = d.lines || [d.subject ? d.subject.name : ""];
    const lineCount = lines.length;
    const totalTextH = lineCount * DIMENSIONS.LINE_HEIGHT;
    const startY = d.y + (d.height - totalTextH) / 2 + DIMENSIONS.LINE_HEIGHT - 3;

    const text = d3.select(this).append("text")
      .attr("text-anchor", "middle")
      .attr("font-size", FONTS.NODE_SIZE)
      .attr("font-family", FONTS.NODE_FACE)
      .attr("fill", COLORS.NODE_TEXT)
      .attr("pointer-events", "none")
      .attr("class", "node-label");

    text.selectAll("tspan")
      .data(lines)
      .enter()
      .append("tspan")
      .attr("x", d.x + d.width / 2)
      .attr("dy", (_, i) => i === 0 ? 0 : DIMENSIONS.LINE_HEIGHT)
      .attr("y", (_, i) => i === 0 ? startY : null)
      .text((l) => l);
  });

  // ── Leyenda adaptada según dispositivo ──
  const legendDiv = document.createElement("div");
  legendDiv.id = "graph-legend";

  const actionLegendText = isTouch
    ? "Toca para conexiones / Toca de nuevo para detalles"
    : "Clic en materia para detalles";

  legendDiv.innerHTML = `
    <div id="legend-title">Navegación</div>
    <div class="legend-item">
      <svg width="8" height="8" viewBox="0 0 8 8" fill="#94a3b8" style="flex-shrink:0;width:8px;height:8px">
        <rect x="1" y="1" width="2.5" height="2.5" rx=".6"/>
        <rect x="4.5" y="1" width="2.5" height="2.5" rx=".6"/>
        <rect x="1" y="4.5" width="2.5" height="2.5" rx=".6"/>
        <rect x="4.5" y="4.5" width="2.5" height="2.5" rx=".6"/>
      </svg>
      Arrastra para mover
    </div>
    <div class="legend-item">
      <svg width="8" height="10" viewBox="0 0 8 10" fill="none" stroke="#94a3b8" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;width:8px;height:10px">
        <path d="M2 1L7 5l-3 .6L3 8z"/>
      </svg>
      ${actionLegendText}
    </div>`;
  element.appendChild(legendDiv);

  // ── Ajuste inicial para que cubra todo el alto ──
  requestAnimationFrame(() => {
    fitGraph(true);
  });

  // ── Events ──
  if (isTouch) {
    // LÓGICA MOBILE / TÁCTIL
    nodeGroups.on("click", (event, d) => {
      event.stopPropagation();

      // Si se toca el mismo nodo que ya está enfocado -> Abrir Modal
      if (_selectedMobileNodeId === d.id) {
        _openModal(d.id);
        return;
      }

      // Si se toca un nodo distinto -> Resaltar sus conexiones
      _selectedMobileNodeId = d.id;
      _highlightRelations(d.id, nodeGroups, edgePaths, COLORS);
    });

    // Tap en el fondo del mapa -> Limpiar resaltado
    svg.on("click", () => {
      if (_selectedMobileNodeId !== null) {
        _selectedMobileNodeId = null;
        _resetStyles(nodeGroups, edgePaths, COLORS);
      }
    });
  } else {
    // LÓGICA DESKTOP (Hover + Clic directo)
    nodeGroups.on("click", (event, d) => {
      event.stopPropagation();
      _openModal(d.id);
    });

    nodeGroups.on("mouseenter", function (event, d) {
      _highlightRelations(d.id, nodeGroups, edgePaths, COLORS);
    });

    nodeGroups.on("mouseleave", function () {
      _resetStyles(nodeGroups, edgePaths, COLORS);
    });
  }

  svg.on("dblclick", () => {
    fitGraph(true);
  });
}

function fitGraph(fillHeight) {
  const svg = d3.select("#graph-container svg");
  if (svg.empty() || !_zoomInstance) return;

  const gc = document.getElementById("graph-container");
  if (!gc) return;
  const rect = gc.getBoundingClientRect();
  const cw = rect.width, ch = rect.height;
  if (cw <= 0 || ch <= 0) return;

  // alto real del contenido vía bounding box del <g>
  let contentH = _graphHeight;
  const gNode = svg.select("g").node();
  if (gNode) {
    try {
      const bb = gNode.getBBox();
      if (bb.height > 0) contentH = bb.height + 6;
    } catch (_) {}
  }

  const effectiveH = ch * 0.85;
  const k = fillHeight
    ? effectiveH / contentH
    : Math.min(cw / _graphWidth, effectiveH / _graphHeight);

  const tx = fillHeight ? 0 : (cw - _graphWidth * k) / 2;
  const ty = fillHeight ? 0 : Math.max((ch - _graphHeight * k) / 2, 0);

  const t = d3.zoomIdentity.translate(tx, ty).scale(k);
  svg.call(_zoomInstance.transform, t);
}

function renderSemesterView() {
  const tabsEl = document.getElementById("semester-tabs");
  const cardsEl = document.getElementById("semester-cards");
  if (!tabsEl || !cardsEl) return;

  const subjectsBySem = {};
  for (const [id, subj] of Object.entries(allSubjects)) {
    const sem = subj.sem;
    if (!subjectsBySem[sem]) subjectsBySem[sem] = [];
    subjectsBySem[sem].push({ id, ...subj });
  }

  for (const sem of Object.keys(subjectsBySem)) {
    subjectsBySem[sem].sort((a, b) => a.name.localeCompare(b.name));
  }

  let activeSem = Math.min(...Object.keys(subjectsBySem).map(Number), 1);

  tabsEl.innerHTML = "";
  const sortedSems = Object.keys(subjectsBySem).map(Number).sort((a, b) => a - b);

  sortedSems.forEach((sem) => {
    const tab = document.createElement("button");
    tab.className = `semester-tab${sem === activeSem ? " active" : ""}`;
    tab.textContent = `${sem}°`;
    tab.setAttribute("data-sem", sem);
    tab.addEventListener("click", () => {
      activeSem = sem;
      tabsEl.querySelectorAll(".semester-tab").forEach((t) => t.classList.remove("active"));
      tab.classList.add("active");
      renderCards(activeSem);
    });
    tabsEl.appendChild(tab);
  });

  function renderCards(sem) {
    cardsEl.innerHTML = "";
    const subjects = subjectsBySem[sem] || [];

    subjects.forEach((subj) => {
      const card = document.createElement("div");
      card.className = "semester-card";
      card.addEventListener("click", () => _openModal(subj.id));

      const header = document.createElement("div");
      header.className = "semester-card-header";

      const name = document.createElement("div");
      name.className = "semester-card-name";
      name.textContent = subj.name;

      const badge = document.createElement("span");
      badge.className = `semester-card-badge badge-${subj.type || "OBL"}`;
      badge.textContent = subj.type || "OBL";

      header.appendChild(name);
      header.appendChild(badge);

      const meta = document.createElement("div");
      meta.className = "semester-card-meta";
      meta.innerHTML = `
        <span>${subj.credits || "–"} créd.</span>
        <span class="meta-sep">·</span>
        <span>${subj.weekly_hours || "–"} hrs/sem</span>
      `;

      const prereqs = document.createElement("div");
      prereqs.className = "semester-card-prereqs";

      const preList = subj.pre || [];
      if (preList.length === 0) {
        const none = document.createElement("span");
        none.className = "semester-card-no-prereqs";
        none.textContent = "Sin pre-requisitos";
        prereqs.appendChild(none);
      } else {
        preList.forEach((preId) => {
          const prereqSubject = allSubjects[preId];
          const chip = document.createElement("button");
          chip.className = "semester-prereq-chip";
          chip.textContent = prereqSubject ? prereqSubject.name : preId;
          if (prereqSubject) {
            chip.addEventListener("click", (e) => {
              e.stopPropagation();
              _openModal(preId);
            });
          }
          prereqs.appendChild(chip);
        });
      }

      card.appendChild(header);
      card.appendChild(meta);
      card.appendChild(prereqs);
      cardsEl.appendChild(card);
    });
  }

  renderCards(activeSem);
}

// ================= FUNCIONES AUXILIARES =================

function _wrapText(text, maxWidth, charWidth) {
  const maxChars = Math.max(Math.floor((maxWidth - 16) / charWidth), 5);
  const words = text.split(" ");
  const lines = [];
  let current = "";
  for (const word of words) {
    const test = current ? current + " " + word : word;
    if (test.length > maxChars && current) {
      lines.push(current);
      current = word;
    } else {
      current = test;
    }
  }
  if (current) lines.push(current);
  return lines;
}

function _createSubjectNodes(semesters) {
  const allSubjectsLocal = {};
  const positions = {};

  for (let sem = 1; sem <= semestersCount; sem++) {
    const subjects = semesters[sem];
    const ids = Object.keys(subjects).sort((a, b) =>
      subjects[a].name.localeCompare(subjects[b].name),
    );

    const xBase = (sem - 1) * DIMENSIONS.SEMESTER_WIDTH + 50;

    ids.forEach((id, i) => {
      const subject = subjects[id];

      allSubjectsLocal[id] = {
        ...subject,
        sem,
        pre: subject.prerequisites || [],
      };

      const charWidth = 7.8;
      const textWidth = subject.name.length * charWidth + 30;
      const nodeWidth = Math.min(Math.max(textWidth, 90), DIMENSIONS.NODE_MAX_WIDTH);
      const lines = _wrapText(subject.name, nodeWidth, charWidth);
      const nodeHeight = lines.length * DIMENSIONS.LINE_HEIGHT + 10;

      positions[id] = {
        x: xBase + (DIMENSIONS.SEMESTER_WIDTH - nodeWidth) / 2,
        y: DIMENSIONS.NODE_START_Y + i * DIMENSIONS.NODE_SPACING_Y,
        width: nodeWidth,
        height: Math.max(nodeHeight, DIMENSIONS.NODE_HEIGHT),
        lines,
      };
    });
  }

  return { allSubjects: allSubjectsLocal, nodePositions: positions };
}

function _createEdges(allSubjectsLocal) {
  const edges = [];
  for (const [id, subject] of Object.entries(allSubjectsLocal)) {
    for (const pre of subject.pre) {
      edges.push({ from: pre, to: id });
    }
  }
  return edges;
}

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

// ================= HIGHLIGHT / RESET =================

function _highlightRelations(id, nodeGroups, edgePaths, COLORS) {
  const parents = _getAllParents(id, allSubjects);
  const children = _getDirectChildren(id, allSubjects);
  const related = new Set([id, ...parents, ...children]);

  nodeGroups.each(function (d) {
    const rect = this.querySelector(".node-rect");
    const label = this.querySelector(".node-label");
    if (!rect) return;

    if (d.id === id) {
      rect.setAttribute("fill", COLORS.NODE_HIGHLIGHT_BG);
      rect.setAttribute("stroke", COLORS.NODE_HIGHLIGHT_BORDER);
      rect.setAttribute("stroke-width", "2");
      if (label) label.setAttribute("font-weight", "bold");
    } else if (parents.has(d.id)) {
      rect.setAttribute("fill", COLORS.NODE_PARENT_BG);
      rect.setAttribute("stroke", COLORS.NODE_PARENT_BORDER);
      if (label) label.setAttribute("font-weight", "bold");
    } else if (children.has(d.id)) {
      rect.setAttribute("fill", COLORS.NODE_CHILD_BG);
      rect.setAttribute("stroke", COLORS.NODE_CHILD_BORDER);
      if (label) label.setAttribute("font-weight", "bold");
    } else {
      rect.setAttribute("opacity", "0.12");
    }
  });

  edgePaths.attr("opacity", (d) => {
    const fromRelated = related.has(d.from);
    const toRelated = related.has(d.to);
    return fromRelated && toRelated ? COLORS.EDGE_OPACITY : 0.04;
  });
}

function _resetStyles(nodeGroups, edgePaths, COLORS) {
  nodeGroups.each(function (d) {
    const rect = this.querySelector(".node-rect");
    const label = this.querySelector(".node-label");
    if (!rect) return;

    const sem = d.subject ? d.subject.sem : 1;
    const bg = sem % 2 === 0 ? COLORS.NODE_BG_EVEN : COLORS.NODE_BG_ODD;
    rect.setAttribute("fill", bg);
    rect.setAttribute("stroke", COLORS.NODE_BORDER);
    rect.setAttribute("stroke-width", "1");
    rect.setAttribute("opacity", "1");
    if (label) label.setAttribute("font-weight", "normal");
  });

  edgePaths.attr("opacity", COLORS.EDGE_OPACITY);
}

// ================= FUNCIONES PARA NODOS RELACIONADOS =================

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

function _getDirectChildren(id, allSubs) {
  const children = new Set();
  for (const [sid, subj] of Object.entries(allSubs)) {
    if (subj.pre.includes(id)) children.add(sid);
  }
  return children;
}

// ================= MODAL =================

function _openModal(id) {
  const subject = allSubjects[id];
  if (!subject) return;

  const existing = document.getElementById("subject-modal-overlay");
  if (existing) existing.remove();

  const overlay = document.createElement("div");
  overlay.id = "subject-modal-overlay";
  overlay.className = "modal-overlay";

  const modal = document.createElement("div");
  modal.className = "modal-content";
  modal.style.fontFamily = "Inter, Arial, sans-serif";
  modal.addEventListener("click", (e) => e.stopPropagation());

  const header = document.createElement("div");
  header.className = "modal-header";
  header.style.color = "#ffffff";

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

  const close = document.createElement("button");
  close.innerHTML = "&times;";
  Object.assign(close.style, {
    position: "absolute",
    top: "14px",
    right: "14px",
    border: "none",
    background: "rgba(255,255,255,0.2)",
    borderRadius: "4px",
    width: "36px",
    height: "36px",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    fontSize: "20px",
    cursor: "pointer",
    color: "#ffffff",
    transition: "background 0.2s",
    zIndex: "2",
  });
  close.onmouseover = () => close.style.background = "rgba(255,255,255,0.3)";
  close.onmouseout = () => close.style.background = "rgba(255,255,255,0.2)";
  close.onclick = () => overlay.remove();

  const body = document.createElement("div");
  body.className = "modal-body";
  body.style.overflowY = "auto";

  const SVG = {
    book: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="#557fab" stroke-width="2" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25"/></svg>`,
    star: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="#557fab" stroke-width="2" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557l-4.204-3.602a.562.562 0 0 1 .321-.988l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z"/></svg>`,
    clock: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="#557fab" stroke-width="2" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>`,
    info: `<svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" style="vertical-align:middle;flex-shrink:0"><path stroke-linecap="round" stroke-linejoin="round" d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"/></svg>`,
  };

  function createStat(label, value, iconSvg) {
    const col = document.createElement("div");
    col.className = "modal-stat-card";
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
  statsGrid.className = "modal-stats-row";
  statsGrid.appendChild(createStat("Semestre", subject.semester, SVG.book));
  statsGrid.appendChild(createStat("Créditos", subject.credits, SVG.star));
  statsGrid.appendChild(createStat("Hrs/Sem", subject.weekly_hours, SVG.clock));

  const desc = document.createElement("div");
  desc.className = "modal-desc";
  desc.innerHTML = `<strong>Descripción:</strong><br/>${subject.desc ?? "Sin descripción disponible."}`;

  const prereqSection = document.createElement("div");
  prereqSection.className = "modal-prereqs";

  const prereqTitle = document.createElement("div");
  prereqTitle.className = "modal-section-title";
  prereqTitle.textContent = "Pre-requisitos";
  prereqSection.appendChild(prereqTitle);

  const prereqList = subject.pre || [];

  if (prereqList.length === 0) {
    const none = document.createElement("p");
    none.className = "modal-no-prereqs";
    none.textContent = "Esta materia no tiene pre-requisitos.";
    prereqSection.appendChild(none);
  } else {
    const chipsWrapper = document.createElement("div");
    chipsWrapper.className = "modal-chips";
    prereqList.forEach((preId) => {
      const prereqSubject = allSubjects[preId];
      const chip = document.createElement("button");
      chip.className = "prereq-chip";
      chip.textContent = prereqSubject ? prereqSubject.name : preId;
      chip.style.cursor = prereqSubject ? "pointer" : "default";
      chip.title = prereqSubject ? `Ver ${prereqSubject.name}` : preId;
      if (prereqSubject) {
        chip.innerHTML = prereqSubject.name;
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
    req.className = "modal-warning";
    req.innerHTML = `${SVG.info} <span><strong>Nota:</strong> Requiere ${subject.required_credits} créditos aprobados.</span>`;
    body.appendChild(req);
  }

  modal.appendChild(close);
  modal.appendChild(header);
  body.appendChild(statsGrid);
  modal.appendChild(body);
  overlay.appendChild(modal);
  document.body.appendChild(overlay);

  overlay.addEventListener("click", () => overlay.remove());

  const escHandler = (e) => {
    if (e.key === "Escape") {
      overlay.remove();
      document.removeEventListener("keydown", escHandler);
    }
  };
  document.addEventListener("keydown", escHandler);
}
