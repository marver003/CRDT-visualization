type TimeGraphKind =
  | "create"
  | "increment"
  | "remove"
  | "send"
  | "receive"
  | "merge"
  | "arrive"
  | "read"
  | "ack"
  | "response";

type TimeGraphEntry = {
  step: number;
  nodeId: string;
  kind: TimeGraphKind;
  value?: number;
  label?: string;
  targets?: string[];
  broadcast?: boolean;
};

type TimeGraphOperation = {
  nodeId: string;
  kind: TimeGraphKind;
  value?: number;
  label?: string;
  targets?: string[];
  broadcast?: boolean;
};

const timeGraphContainer = document.getElementById("time-graph-container") as HTMLElement;

let timeGraphRows: string[] = [];
let timeGraphEntries: TimeGraphEntry[] = [];
let timeGraphStepCount = 0;

function resetTimeGraph() {
  timeGraphRows = [];
  timeGraphEntries = [];
  timeGraphStepCount = 0;
  renderTimeGraph();
}

function recordTimeGraphBatch(operations: TimeGraphOperation[]) {
  if (operations.length === 0) {
    return;
  }

  timeGraphStepCount += 1;

  for (const operation of operations) {
    if (!timeGraphRows.includes(operation.nodeId)) {
      timeGraphRows.push(operation.nodeId);
    }

    timeGraphEntries.push({ step: timeGraphStepCount, ...operation });
  }

  renderTimeGraph();
}

function renderTimeGraph() {
  if (timeGraphEntries.length === 0) {
    timeGraphContainer.innerHTML = '<p class="time-graph-empty">No steps yet</p>';
    return;
  }

  const entryAt = new Map<string, TimeGraphEntry[]>();
  const createdAt = new Map<string, number>();
  const removedAt = new Map<string, number>();

  for (const entry of timeGraphEntries) {
    const key = `${entry.nodeId}:${entry.step}`;
    entryAt.set(key, [...(entryAt.get(key) ?? []), entry]);

    if (entry.kind === "create" && !createdAt.has(entry.nodeId)) {
      createdAt.set(entry.nodeId, entry.step);
    }

    if (entry.kind === "remove") {
      removedAt.set(entry.nodeId, entry.step);
    }
  }

  const table = document.createElement("table");
  table.className = "time-graph-table";

  const thead = document.createElement("thead");
  const headRow = document.createElement("tr");
  headRow.appendChild(document.createElement("th"));

  for (let step = 1; step <= timeGraphStepCount; step++) {
    const th = document.createElement("th");
    th.textContent = `${step}`;
    headRow.appendChild(th);
  }

  thead.appendChild(headRow);
  table.appendChild(thead);

  const tbody = document.createElement("tbody");

  const sortedRows = [...timeGraphRows].sort(compareTimeGraphRows);

  for (const nodeId of sortedRows) {
    const row = document.createElement("tr");
    const rowHeader = document.createElement("th");
    rowHeader.scope = "row";
    rowHeader.textContent = timeGraphRowLabel(nodeId);

    if (isClientRow(nodeId)) {
      rowHeader.classList.add("time-graph-row--client");
      rowHeader.title = `Client ${nodeId.slice(CLIENT_ROW_PREFIX.length)}`;
    }
    row.appendChild(rowHeader);

    const firstStep = createdAt.get(nodeId);
    const lastStep = removedAt.get(nodeId);

    for (let step = 1; step <= timeGraphStepCount; step++) {
      const td = document.createElement("td");
      const entries = entryAt.get(`${nodeId}:${step}`);

      if (entries && entries.length === 1) {
        applyEntry(td, entries[0]);
      } else if (entries) {
        td.classList.add("time-graph-cell--multi");

        for (const entry of entries) {
          const chip = document.createElement("span");
          applyEntry(chip, entry);
          td.appendChild(chip);
        }
      } else if (
        (firstStep !== undefined && step < firstStep) ||
        (lastStep !== undefined && step > lastStep)
      ) {
        td.classList.add("time-graph-cell--gone");
      }

      row.appendChild(td);
    }

    tbody.appendChild(row);
  }

  table.appendChild(tbody);
  timeGraphContainer.replaceChildren(table);
  timeGraphContainer.scrollLeft = timeGraphContainer.scrollWidth;
}

function sendLabel(entry: TimeGraphEntry): string {
  if (entry.broadcast) {
    return "→all";
  }

  return `→${(entry.targets ?? []).join("")}`;
}

const CLIENT_ROW_PREFIX = "client:";

function clientRowId(clientId: string) {
  return `${CLIENT_ROW_PREFIX}${clientId}`;
}

function isClientRow(rowId: string) {
  return rowId.startsWith(CLIENT_ROW_PREFIX);
}

function timeGraphRowLabel(rowId: string) {
  return isClientRow(rowId) ? `C${rowId.slice(CLIENT_ROW_PREFIX.length)}` : rowId;
}

function compareTimeGraphRows(a: string, b: string) {
  if (isClientRow(a) !== isClientRow(b)) {
    return isClientRow(a) ? -1 : 1;
  }

  return a.localeCompare(b);
}

function applyEntry(element: HTMLElement, entry: TimeGraphEntry) {
  element.classList.add(`time-graph-cell--${entry.kind}`);
  element.textContent =
    entry.label ??
    (entry.kind === "remove" ? "×" :
    entry.kind === "send" ? sendLabel(entry) :
    entry.value !== undefined ? `${entry.value}` : "");

  if (entry.kind === "send" && entry.targets) {
    element.title = `Sent to ${entry.targets.join(", ")}`;
  }
}
