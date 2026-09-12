"use strict";
const NETWORK_PADDING = 40;
const CARD_GAP = 48;
const MINIMUM_RADIUS = 160;
const SVG_NAMESPACE = "http://www.w3.org/2000/svg";
// DOM references
const createNodeButton = document.querySelector(".btn-primary");
const nodesContainer = document.getElementById("nodes-container");
let activeNetwork = null;
// Pan & zoom. `view` is null until the network has been laid out once, at
// which point it's set to a view that fits the network in the container;
// from then on it's fully user-controlled (wheel to zoom, drag to pan) and
// persists across re-renders so routine actions (increment, merge, step...)
// don't reset the camera. Call resetNetworkView() to drop back to the fit.
let view = null;
const MIN_SCALE = 0.3;
const MAX_SCALE = 3;
function resetNetworkView() {
    view = null;
}
function clamp(value, min, max) {
    return Math.min(max, Math.max(min, value));
}
function getFittedView(size) {
    const containerW = nodesContainer.clientWidth;
    const containerH = nodesContainer.clientHeight;
    const scale = containerW > 0 && containerH > 0
        ? Math.min(1, containerW / size.width, containerH / size.height)
        : 1;
    return {
        scale,
        x: (containerW - size.width * scale) / 2,
        y: (containerH - size.height * scale) / 2,
    };
}
function applyView(stage) {
    if (!view)
        return;
    stage.style.transform = `translate(${view.x}px, ${view.y}px) scale(${view.scale})`;
}
// State rendering
async function renderState() {
    const response = await getStateAPI();
    if (!response) {
        return;
    }
    const { state } = response;
    const network = createNetwork();
    activeNetwork = network;
    for (const [nodeId, counts] of Object.entries(state)) {
        const replicaCounts = counts;
        const card = createNodeCard({ nodeId, state: replicaCounts });
        network.stage.appendChild(card);
        network.cards.push(card);
    }
    requestAnimationFrame(() => {
        if (activeNetwork === network) {
            layoutCircularNetwork(network);
        }
    });
}
function createNetwork() {
    nodesContainer.replaceChildren();
    const stage = document.createElement("div");
    stage.className = "node-network";
    const links = document.createElementNS(SVG_NAMESPACE, "svg");
    links.classList.add("node-network-links");
    stage.appendChild(links);
    nodesContainer.appendChild(stage);
    return { stage, cards: [] };
}
// Circular network layout
function layoutCircularNetwork({ stage, cards }) {
    const links = stage.querySelector(".node-network-links");
    if (!links || cards.length === 0) {
        return;
    }
    const cardSize = getLargestCardSize(cards);
    const radius = getNetworkRadius(cards.length, cardSize);
    const size = getNetworkSize(radius, cardSize);
    const positions = positionCards(cards, { x: size.width / 2, y: size.height / 2 }, radius);
    stage.style.width = `${size.width}px`;
    stage.style.height = `${size.height}px`;
    drawConnections(links, positions, size);
    // First layout (or after resetNetworkView()): start from a view that fits
    // the whole network in the container. After that the user's pan/zoom wins.
    if (view === null) {
        view = getFittedView(size);
    }
    applyView(stage);
}
function getLargestCardSize(cards) {
    return {
        width: Math.max(...cards.map((card) => card.offsetWidth || 260)),
        height: Math.max(...cards.map((card) => card.offsetHeight || 190)),
    };
}
function getNetworkRadius(nodeCount, cardSize) {
    if (nodeCount <= 1) {
        return 0;
    }
    const minimumDistance = Math.hypot(cardSize.width, cardSize.height) + CARD_GAP;
    const desiredRadius = minimumDistance / (2 * Math.sin(Math.PI / nodeCount));
    // Ensure minimum radius keeps nodes well separated as originally intended
    return Math.max(desiredRadius, MINIMUM_RADIUS);
}
function getNetworkSize(radius, cardSize) {
    // Account for card half-sizes plus padding
    const neededWidth = Math.ceil(radius * 2 + cardSize.width + NETWORK_PADDING);
    const neededHeight = Math.ceil(radius * 2 + cardSize.height + NETWORK_PADDING);
    return {
        width: Math.max(nodesContainer.clientWidth, neededWidth),
        height: Math.max(nodesContainer.clientHeight, neededHeight),
    };
}
function positionCards(cards, center, radius) {
    return cards.map((card, index) => {
        const angle = -Math.PI / 2 + (2 * Math.PI * index) / cards.length;
        const position = cards.length === 1
            ? center
            : {
                x: center.x + radius * Math.cos(angle),
                y: center.y + radius * Math.sin(angle),
            };
        card.style.left = `${position.x}px`;
        card.style.top = `${position.y}px`;
        return position;
    });
}
function drawConnections(links, positions, size) {
    links.setAttribute("viewBox", `0 0 ${size.width} ${size.height}`);
    links.setAttribute("width", `${size.width}`);
    links.setAttribute("height", `${size.height}`);
    links.replaceChildren();
    for (let i = 0; i < positions.length; i++) {
        for (let j = i + 1; j < positions.length; j++) {
            const line = document.createElementNS(SVG_NAMESPACE, "line");
            line.setAttribute("x1", `${positions[i].x}`);
            line.setAttribute("y1", `${positions[i].y}`);
            line.setAttribute("x2", `${positions[j].x}`);
            line.setAttribute("y2", `${positions[j].y}`);
            links.appendChild(line);
        }
    }
}
// Node operations
function getNodeId(input) {
    const value = input.value.trim();
    return /^[A-Za-z]$/.test(value) ? value : null;
}
async function createNode() {
    const replicaIdInput = document.getElementById("replica-id");
    const nodeId = replicaIdInput instanceof HTMLInputElement ? getNodeId(replicaIdInput) : null;
    if (nodeId === null) {
        return;
    }
    if (!await createNodeAPI(nodeId.toUpperCase())) {
        return;
    }
    await runQueuedSteps();
}
async function removeNode(nodeId) {
    if (nodeId === null) {
        return;
    }
    if (!await removeNodeAPI(nodeId)) {
        return;
    }
    await runQueuedSteps();
}
async function incrementNode(nodeId) {
    if (nodeId === null) {
        return;
    }
    if (!await incrementNodeAPI(nodeId)) {
        return;
    }
    await runQueuedSteps();
}
// Pan & zoom interaction
let isPanning = false;
let panPointerId = null;
let panStart = { x: 0, y: 0 };
let panViewStart = { x: 0, y: 0 };
nodesContainer.addEventListener("wheel", (e) => {
    if (!activeNetwork || !view)
        return;
    e.preventDefault();
    const rect = nodesContainer.getBoundingClientRect();
    const cursorX = e.clientX - rect.left;
    const cursorY = e.clientY - rect.top;
    const zoomFactor = Math.exp(-e.deltaY * 0.0015);
    const newScale = clamp(view.scale * zoomFactor, MIN_SCALE, MAX_SCALE);
    const ratio = newScale / view.scale;
    // Keep the point under the cursor fixed while the scale changes.
    view = {
        scale: newScale,
        x: cursorX + (view.x - cursorX) * ratio,
        y: cursorY + (view.y - cursorY) * ratio,
    };
    applyView(activeNetwork.stage);
}, { passive: false });
nodesContainer.addEventListener("pointerdown", (e) => {
    if (!activeNetwork || !view || e.button !== 0)
        return;
    // Let clicks on real controls (buttons, selects...) through untouched.
    if (e.target.closest("button, select, input, a"))
        return;
    isPanning = true;
    panPointerId = e.pointerId;
    panStart = { x: e.clientX, y: e.clientY };
    panViewStart = { x: view.x, y: view.y };
    nodesContainer.setPointerCapture(e.pointerId);
    nodesContainer.classList.add("panning");
});
nodesContainer.addEventListener("pointermove", (e) => {
    if (!isPanning || !activeNetwork || !view || e.pointerId !== panPointerId)
        return;
    view = {
        ...view,
        x: panViewStart.x + (e.clientX - panStart.x),
        y: panViewStart.y + (e.clientY - panStart.y),
    };
    applyView(activeNetwork.stage);
});
function stopPanning(e) {
    if (!isPanning || e.pointerId !== panPointerId)
        return;
    isPanning = false;
    panPointerId = null;
    nodesContainer.releasePointerCapture(e.pointerId);
    nodesContainer.classList.remove("panning");
}
nodesContainer.addEventListener("pointerup", stopPanning);
nodesContainer.addEventListener("pointercancel", stopPanning);
// Events
createNodeButton.addEventListener("click", createNode);
window.addEventListener("resize", () => activeNetwork && layoutCircularNetwork(activeNetwork));
