"use strict";
const API_BASE = "http://localhost:8080";
async function requestAPI(path, options) {
    try {
        const response = await fetch(`${API_BASE}${path}`, options);
        if (!response.ok) {
            showCriticalError("Backend request failed", `The backend returned HTTP ${response.status} - ${response.statusText}`);
            return null;
        }
        return response;
    }
    catch {
        showCriticalError("Backend connection failed", "The frontend could not connect to the backend. Start the backend service and try again.");
        return null;
    }
}
async function requestJSON(path, options) {
    const response = await requestAPI(path, options);
    if (!response) {
        return null;
    }
    try {
        return (await response.json());
    }
    catch {
        showCriticalError("Invalid backend response", "The backend returned data that the simulator could not read.");
        return null;
    }
}
async function createNodeAPI(replicaId) {
    return requestAPI("/createNode", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            id: replicaId,
        }),
    });
}
async function removeNodeAPI(replicaId) {
    return requestAPI(`/removeNode/${encodeURIComponent(replicaId)}`, {
        method: "DELETE",
    });
}
async function incrementNodeAPI(replicaId) {
    return requestAPI("/increment", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            id: replicaId,
        }),
    });
}
async function mergeNodesAPI(fromReplicaId, toReplicaId) {
    return requestAPI("/merge", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            "source-id": fromReplicaId,
            "target-id": toReplicaId,
        }),
    });
}
async function getStepsAPI() {
    return requestJSON("/steps");
}
async function executeNextStepAPI() {
    return requestAPI("/step", {
        method: "POST",
    });
}
async function getStateAPI() {
    return requestJSON("/state");
}
async function resetAPI() {
    return requestAPI("/reset", {
        method: "POST",
    });
}
async function loadStateAPI(stateJson) {
    return requestAPI("/load", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: stateJson
    });
}
