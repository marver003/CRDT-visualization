const createNodeButton = document.querySelector(
  ".btn-primary",
) as HTMLButtonElement;

const removeNodeButton = document.querySelector(
  "#remove-button",
) as HTMLButtonElement;

const incrementButton = document.getElementById(
  "increment-button",
) as HTMLButtonElement;

const mergeButton = document.getElementById(
  "merge-button",
) as HTMLButtonElement;

const nodesContainer = document.getElementById(
  "nodes-container",
) as HTMLElement;

const nodeRemoveSelection = document.getElementById(
  "remove-node",
) as HTMLSelectElement;

const nodeIncrementSelection = document.getElementById(
  "increment-node",
) as HTMLSelectElement;

const nodeMergeFromSelection = document.getElementById(
  "merge-from",
) as HTMLSelectElement;

const nodeMergeToSelection = document.getElementById(
  "merge-to",
) as HTMLSelectElement;

async function renderState() {
  const { state } = await getStateAPI();

  nodesContainer.innerHTML = "";
  nodeRemoveSelection.innerHTML = "<option>&lt;Select&gt;</option>";
  nodeIncrementSelection.innerHTML = "<option>&lt;Select&gt;</option>";
  nodeMergeFromSelection.innerHTML = "<option>&lt;Select&gt;</option>";
  nodeMergeToSelection.innerHTML = "<option>&lt;Select&gt;</option>";

  for (const [nodeId, counts] of Object.entries(state)) {
    const replicaCounts = counts as Record<string, number>;
    const total = Object.values(replicaCounts).reduce((sum, v) => sum + v, 0);

    const node = document.createElement("div");
    node.className = "node-card";

    const countsRows = Object.entries(replicaCounts)
      .map(
        ([replicaId, count]) => `
                <div>
                    <span>Replica ${replicaId}</span>
                    <span>${count}</span>
                </div>`,
      )
      .join("");

    node.innerHTML = `
            <div class="node-card-header">
                <strong>Node ${nodeId}</strong>
            </div>

            <div class="node-card-body">
                ${countsRows}

                <div class="node-total">
                    <span>Total</span>
                    <strong>${total}</strong>
                </div>
            </div>
        `;

    const nodeOption = document.createElement("option");
    nodeOption.innerText = `${nodeId}`;

    nodesContainer.appendChild(node);

    nodeRemoveSelection.appendChild(nodeOption);
    nodeIncrementSelection.appendChild(nodeOption.cloneNode(true));
    nodeMergeFromSelection.appendChild(nodeOption.cloneNode(true));
    nodeMergeToSelection.appendChild(nodeOption.cloneNode(true));
  }

  await renderStepsQueue();
}

function getReplicaId(): number | null {
  const input = document.getElementById("replica-id");

  if (!(input instanceof HTMLInputElement)) {
    return null;
  }

  const value = input.value.trim();

  if (value === "") {
    return null;
  }

  return Number(value);
}

async function createNode() {
  const id = getReplicaId();

  if (id == null || id == 0) {
    return;
  }

  await createNodeAPI(id);

  await renderStepsQueue();
}

async function removeNode() {
  const id = Number(nodeRemoveSelection.value);

  if (id == null || id == 0) {
    return;
  }

  await removeNodeAPI(id);

  await renderStepsQueue();
}

async function incrementNode() {
  const id = Number(nodeIncrementSelection.value);

  if (id == null || id == 0) {
    return;
  }

  await incrementNodeAPI(id);

  await renderStepsQueue();
}

async function mergeNodes() {
  const fromId = Number(nodeMergeFromSelection.value);
  const toId = Number(nodeMergeToSelection.value);

  if (fromId === toId) {
    return;
  }

  await mergeNodesAPI(fromId, toId);

  await renderStepsQueue();
}

createNodeButton.addEventListener("click", createNode);
removeNodeButton.addEventListener("click", removeNode);
incrementButton.addEventListener("click", incrementNode);
mergeButton.addEventListener("click", mergeNodes);
