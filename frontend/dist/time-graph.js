"use strict";
// Replication timeline widget - a lightweight visual history of executed
// steps, loosely modeled on the classic Raft "servers x time" diagram: one
// row per replica, one column per executed step, each cell showing what
// happened to that replica at that step. This replaces the old text-based
// Step Queue log.
//
// The history is kept client-side only. The backend's /steps endpoint only
// ever returns the *pending* queue - once a step executes it's dequeued and
// gone, so there's no server-side log of what already happened. Each entry
// here is recorded by steps.ts right after it successfully executes a step.
const timeGraphContainer = document.getElementById("time-graph-container");
let timeGraphRows = [];
let timeGraphEntries = [];
let timeGraphStepCount = 0;
function resetTimeGraph() {
    timeGraphRows = [];
    timeGraphEntries = [];
    timeGraphStepCount = 0;
    renderTimeGraph();
}
function recordTimeGraphStep(nodeId, kind, value) {
    if (!timeGraphRows.includes(nodeId)) {
        timeGraphRows.push(nodeId);
    }
    timeGraphStepCount += 1;
    timeGraphEntries.push({ step: timeGraphStepCount, nodeId, kind, value });
    renderTimeGraph();
}
function renderTimeGraph() {
    if (timeGraphEntries.length === 0) {
        timeGraphContainer.innerHTML = '<p class="time-graph-empty">No steps yet</p>';
        return;
    }
    const entryAt = new Map();
    const createdAt = new Map();
    const removedAt = new Map();
    for (const entry of timeGraphEntries) {
        entryAt.set(`${entry.nodeId}:${entry.step}`, entry);
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
    for (const nodeId of timeGraphRows) {
        const row = document.createElement("tr");
        const rowHeader = document.createElement("th");
        rowHeader.scope = "row";
        rowHeader.textContent = nodeId;
        row.appendChild(rowHeader);
        const firstStep = createdAt.get(nodeId);
        const lastStep = removedAt.get(nodeId);
        for (let step = 1; step <= timeGraphStepCount; step++) {
            const td = document.createElement("td");
            const entry = entryAt.get(`${nodeId}:${step}`);
            if (entry) {
                td.classList.add(`time-graph-cell--${entry.kind}`);
                td.textContent = entry.kind === "remove" ? "×" : entry.value !== undefined ? `${entry.value}` : "";
            }
            else if ((firstStep !== undefined && step < firstStep) ||
                (lastStep !== undefined && step > lastStep)) {
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
