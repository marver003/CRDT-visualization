const CLIENT_IDS = ["1", "2"];

interface ClientView {
  id: string;
  select: HTMLSelectElement;
  writeButton: HTMLButtonElement;
  readButton: HTMLButtonElement;
  result: HTMLElement;
}

const clientViews: ClientView[] = [];
const clientResults = new Map<string, ClientResult>();

function createClient(id: string): HTMLElement {
  const client = document.createElement("div");
  client.className = "client";
  client.dataset.clientId = id;
  client.innerHTML = `
    <svg class="client-figure" viewBox="0 0 60 100" aria-hidden="true">
      <circle cx="30" cy="16" r="11" />
      <line x1="30" y1="27" x2="30" y2="62" />
      <line x1="30" y1="38" x2="10" y2="54" />
      <line x1="30" y1="38" x2="50" y2="54" />
      <line x1="30" y1="62" x2="14" y2="92" />
      <line x1="30" y1="62" x2="46" y2="92" />
    </svg>
    <div class="client-panel">
      <strong class="client-title">Client ${id}</strong>
      <select class="client-select" aria-label="Node for client ${id}"></select>
      <div class="client-actions">
        <button class="client-button client-write" type="button" title="Send a write (increment) request">Write</button>
        <button class="client-button client-read" type="button" title="Send a read request">Read</button>
      </div>
      <div class="client-result" aria-live="polite"></div>
    </div>
  `;

  const view: ClientView = {
    id,
    select: client.querySelector(".client-select") as HTMLSelectElement,
    writeButton: client.querySelector(".client-write") as HTMLButtonElement,
    readButton: client.querySelector(".client-read") as HTMLButtonElement,
    result: client.querySelector(".client-result") as HTMLElement,
  };

  view.writeButton.addEventListener("click", () => sendClientRequest(view, writeAPI));
  view.readButton.addEventListener("click", () => sendClientRequest(view, readAPI));

  clientViews.push(view);
  return client;
}

async function sendClientRequest(
  view: ClientView,
  api: (clientId: string, nodeId: string) => Promise<Response | null>,
) {
  if (!view.select.value) {
    return;
  }

  if (!await api(view.id, view.select.value)) {
    return;
  }

  if (!isAdvancing) {
    const knownMessages = currentInFlight;

    await renderState();
    await nextFrame();
    playSpawnAnimations(knownMessages);
  }

  notifyActionQueued();
}

function updateClientResults(clients: ClientSnapshot[]) {
  for (const client of clients) {
    const view = clientViews.find((item) => item.id === client.id);
    const result = client.result;

    if (result) {
      clientResults.set(client.id, result);
    } else {
      clientResults.delete(client.id);
    }

    if (!view) {
      continue;
    }

    view.result.textContent = !result
      ? ""
      : result.kind === "write_ack"
        ? `Ack from ${result.nodeId}`
        : `Node ${result.nodeId} = ${result.value}`;
  }
}

function getClientStagePoint(clientId: string): Point | null {
  const client = document.querySelector<HTMLElement>(`.client[data-client-id="${clientId}"]`);

  if (!client || !activeNetwork || !view) {
    return null;
  }

  const clientRect = client.getBoundingClientRect();
  const stageRect = activeNetwork.stage.getBoundingClientRect();

  return {
    x: (clientRect.right - stageRect.left) / view.scale,
    y: (clientRect.top + clientRect.height / 2 - stageRect.top) / view.scale,
  };
}

function updateClientNodeOptions(nodeIds: string[]) {
  const sorted = [...nodeIds].sort();

  for (const clientView of clientViews) {
    const previous = clientView.select.value;

    clientView.select.innerHTML = sorted
      .map((nodeId) => `<option value="${nodeId}">Node ${nodeId}</option>`)
      .join("");

    if (sorted.includes(previous)) {
      clientView.select.value = previous;
    }

    const disabled = sorted.length === 0;
    clientView.select.disabled = disabled;
    clientView.writeButton.disabled = disabled;
    clientView.readButton.disabled = disabled;
  }
}

const clientsContainer = document.getElementById("clients");

if (clientsContainer) {
  for (const id of CLIENT_IDS) {
    clientsContainer.appendChild(createClient(id));
  }
  updateClientNodeOptions([]);
}
