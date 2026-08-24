const API_BASE = "http://localhost:8080";

async function createNodeAPI(replicaId: number) {
  const response = await fetch(`${API_BASE}/createNode`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      id: replicaId,
    }),
  });
}

async function removeNodeAPI(replicaId: number) {
  await fetch(`${API_BASE}/removeNode/${replicaId}`, {
    method: "DELETE",
  });
}

async function incrementNodeAPI(replicaId: number) {
  const response = await fetch(`${API_BASE}/increment`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      id: replicaId,
    }),
  });
}

async function mergeNodesAPI(fromReplicaId: number, toReplicaId: number) {
  await fetch(`${API_BASE}/merge`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      "source-id": fromReplicaId,
      "target-id": toReplicaId
    })
  })
}

async function getStepsAPI() {
  const response = await fetch(`${API_BASE}/steps`);
  return response.json();
}

async function executeNextStepAPI() {
  const response = await fetch(`${API_BASE}/step`, {
    method: "POST",
  });
}

async function getStateAPI() {
  const response = await fetch(`${API_BASE}/state`);
  return response.json();
}

async function resetAPI() {
  await fetch(`${API_BASE}/reset`, {
    method: "POST",
  });
}
