"use strict";
// Step execution. There's no visible step queue/"Next Step" button anymore -
// the backend still queues every action instead of applying it instantly, so
// runQueuedSteps() auto-drains that queue right after each action, playing
// the gossip send/receive animations below for any step that needs them.
// (Merge is currently removed from the UI, so send_gossip/receive_gossip
// steps don't occur today - this stays ready for when merge comes back.)
const GOSSIP_ANIMATION_DURATION = 1000;
const RECEIVE_ANIMATION_DURATION = 1000;
let isRunningSteps = false;
let stepsRunAgain = false;
async function runQueuedSteps() {
    if (isRunningSteps) {
        // A drain is already in flight (e.g. another card's button was clicked
        // moments ago) - ask it to loop once more after it finishes instead of
        // starting a second, overlapping drain.
        stepsRunAgain = true;
        return;
    }
    isRunningSteps = true;
    try {
        do {
            stepsRunAgain = false;
            await drainStepQueue();
        } while (stepsRunAgain);
    }
    finally {
        isRunningSteps = false;
    }
}
async function drainStepQueue() {
    while (true) {
        const nextStep = await getNextStep();
        if (!nextStep) {
            return;
        }
        const previousState = nextStep.type === "receive_gossip"
            ? await getStateAPI()
            : null;
        if (nextStep.type === "send_gossip" && nextStep.from && nextStep.to) {
            await playGossipSendAnimation(nextStep.from, nextStep.to);
        }
        if (nextStep.type === "receive_gossip" && nextStep.to) {
            await playGossipReceiveAnimation(nextStep.to);
        }
        if (!await executeNextStepAPI()) {
            return;
        }
        await renderState();
        recordStepInTimeGraph(nextStep);
        if (previousState && nextStep.to) {
            await nextFrame();
            highlightReceivedState(previousState, nextStep.to);
        }
    }
}
// Feeds the executed step into the Replication Timeline widget (time-graph.ts).
// Reads the resulting own-counter straight from the just-rendered card rather
// than making another API round trip.
function recordStepInTimeGraph(step) {
    switch (step.type) {
        case "create_node":
            if (step.replicaId)
                recordTimeGraphStep(step.replicaId, "create", 0);
            break;
        case "increment":
            if (step.replicaId)
                recordTimeGraphStep(step.replicaId, "increment", getOwnCounter(step.replicaId));
            break;
        case "remove_node":
            if (step.replicaId)
                recordTimeGraphStep(step.replicaId, "remove");
            break;
        case "receive_gossip":
            if (step.to)
                recordTimeGraphStep(step.to, "receive", getOwnCounter(step.to));
            break;
    }
}
function getOwnCounter(nodeId) {
    const card = activeNetwork?.cards.find((item) => item.dataset.nodeId === nodeId);
    const text = card?.querySelector(`[data-replica-id="${nodeId}"]`)?.textContent;
    return text ? Number(text) : undefined;
}
async function getNextStep() {
    const response = await getStepsAPI();
    return response?.steps?.[0] ?? null;
}
async function playGossipSendAnimation(from, to) {
    const network = activeNetwork;
    if (!network) {
        return;
    }
    const sourceCard = network.cards.find((card) => card.dataset.nodeId === from);
    const targetCard = network.cards.find((card) => card.dataset.nodeId === to);
    const links = network.stage.querySelector(".node-network-links");
    if (!sourceCard || !targetCard || !links) {
        return;
    }
    const source = getCardCenter(sourceCard);
    const target = getCardCenter(targetCard);
    const activeLink = document.createElementNS("http://www.w3.org/2000/svg", "line");
    const packet = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    activeLink.classList.add("node-network-active-link");
    activeLink.setAttribute("x1", `${source.x}`);
    activeLink.setAttribute("y1", `${source.y}`);
    activeLink.setAttribute("x2", `${target.x}`);
    activeLink.setAttribute("y2", `${target.y}`);
    packet.classList.add("node-network-packet");
    packet.setAttribute("r", "10");
    links.append(activeLink, packet);
    sourceCard.classList.add("node-card--sending");
    await animatePacket(packet, source, target);
    sourceCard.classList.remove("node-card--sending");
    activeLink.remove();
    packet.remove();
}
async function playGossipReceiveAnimation(nodeId) {
    const card = activeNetwork?.cards.find((item) => item.dataset.nodeId === nodeId);
    if (!card) {
        return;
    }
    card.classList.add("node-card--receiving");
    await wait(RECEIVE_ANIMATION_DURATION);
}
function getCardCenter(card) {
    return {
        x: Number.parseFloat(card.style.left),
        y: Number.parseFloat(card.style.top),
    };
}
function animatePacket(packet, from, to) {
    return new Promise((resolve) => {
        const startedAt = performance.now();
        function move(now) {
            const progress = Math.min((now - startedAt) / GOSSIP_ANIMATION_DURATION, 1);
            const easedProgress = 1 - Math.pow(1 - progress, 3);
            const x = from.x + (to.x - from.x) * easedProgress;
            const y = from.y + (to.y - from.y) * easedProgress;
            packet.setAttribute("cx", `${x}`);
            packet.setAttribute("cy", `${y}`);
            if (progress < 1) {
                requestAnimationFrame(move);
            }
            else {
                resolve();
            }
        }
        requestAnimationFrame(move);
    });
}
function highlightReceivedState(previousState, nodeId) {
    const card = activeNetwork?.cards.find((item) => item.dataset.nodeId === nodeId);
    if (!card) {
        return;
    }
    const previousCounts = previousState.state[nodeId] ?? {};
    card.classList.add("node-card--receiving");
    for (const cell of card.querySelectorAll("[data-replica-id]")) {
        const replicaId = cell.dataset.replicaId;
        const currentValue = Number(cell.textContent);
        if (replicaId && previousCounts[replicaId] !== currentValue) {
            cell.classList.add("node-card-value-changed");
        }
    }
    window.setTimeout(() => card.classList.remove("node-card--receiving"), RECEIVE_ANIMATION_DURATION + 850);
}
function wait(duration) {
    return new Promise((resolve) => window.setTimeout(resolve, duration));
}
function nextFrame() {
    return new Promise((resolve) => requestAnimationFrame(() => resolve()));
}
