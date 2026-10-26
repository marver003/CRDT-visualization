const CLIENT_IDS = ["1", "2"];

const TIMELINE_CLIENT_COLORS: Record<string, string> = { "1": "#0072B2", "2": "#E69F00" };
const TIMELINE_NEUTRAL_COLOR = "#6b778c";

function clientColor(clientId: string | undefined) {
  return (clientId && TIMELINE_CLIENT_COLORS[clientId]) || TIMELINE_NEUTRAL_COLOR;
}

interface ClientView {
  id: string;
  select: HTMLSelectElement;
  element: HTMLInputElement;
  amount: HTMLInputElement;
  writeButton: HTMLButtonElement;
  readButton: HTMLButtonElement;
  result: HTMLElement;
}

const clientViews: ClientView[] = [];
const clientResults = new Map<string, ClientResult>();
const clientLocked = new Map<string, boolean>();
let clientsHaveNodes = false;

function createClient(id: string): HTMLElement {
  const client = document.createElement("div");
  client.className = "client";
  client.dataset.clientId = id;
  client.style.setProperty("--client-color", clientColor(id));
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
      <input class="client-element" type="text" maxlength="12" placeholder="element" autocomplete="off" aria-label="Element for client ${id}" hidden />
      <input class="client-amount" type="number" min="1" max="1000" step="1" value="1" title="Amount to add (1-1000)" aria-label="Amount for client ${id}" />
      <div class="client-actions">
        <button class="client-button client-write" type="button" title="Send a write request">Write</button>
        <button class="client-button client-read" type="button" title="Send a read request">Read</button>
      </div>
      <div class="client-result" aria-live="polite"></div>
    </div>
  `;

  const view: ClientView = {
    id,
    select: client.querySelector(".client-select") as HTMLSelectElement,
    element: client.querySelector(".client-element") as HTMLInputElement,
    amount: client.querySelector(".client-amount") as HTMLInputElement,
    writeButton: client.querySelector(".client-write") as HTMLButtonElement,
    readButton: client.querySelector(".client-read") as HTMLButtonElement,
    result: client.querySelector(".client-result") as HTMLElement,
  };

  view.writeButton.addEventListener("click", () => sendClientRequest(view, writeWithElement));
  for (const input of [view.element, view.amount]) {
    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter") {
        sendClientRequest(view, writeWithElement);
      }
    });
  }
  view.readButton.addEventListener("click", () => sendClientRequest(view, readAPI));

  view.select.disabled = true;
  view.writeButton.disabled = true;
  view.readButton.disabled = true;

  clientViews.push(view);
  return client;
}

const MAX_COUNTER_AMOUNT = 1000;

function writeWithElement(clientId: string, nodeId: string, element?: string, amount?: number) {
  return writeAPI(clientId, nodeId, element, amount);
}

function readAmount(view: ClientView) {
  const parsed = Math.floor(Number(view.amount.value));
  const amount = Number.isFinite(parsed) ? Math.min(MAX_COUNTER_AMOUNT, Math.max(1, parsed)) : 1;

  view.amount.value = String(amount);

  return amount;
}

async function sendClientRequest(
  view: ClientView,
  api: (clientId: string, nodeId: string, element?: string, amount?: number) => Promise<Response | null>,
) {
  if (isAdvancing || clientLocked.get(view.id) || !view.select.value) {
    return;
  }

  const element = view.element.value.trim();

  if (currentCrdt === "gset" && api === writeWithElement && element === "") {
    view.element.focus();
    return;
  }

  const isWrite = api === writeWithElement;
  const amount = usesAmount() && isWrite ? readAmount(view) : undefined;
  const sent = await api(view.id, view.select.value, currentCrdt === "gset" ? element : undefined, amount);

  if (!sent) {
    return;
  }

  if (api === writeWithElement) {
    view.element.value = "";
  }

  clientLocked.set(view.id, true);
  refreshClientControls();

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

    clientLocked.set(client.id, client.locked);

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
        : result.elements
          ? `Node ${result.nodeId} = {${result.elements.join(", ")}}`
          : `Node ${result.nodeId} = ${result.value}`;
  }

  refreshClientControls();
}

function refreshClientControls() {
  for (const clientView of clientViews) {
    clientView.element.hidden = currentCrdt !== "gset";
    clientView.amount.hidden = !usesAmount();

    const blocked = !clientsHaveNodes || isAdvancing || clientLocked.get(clientView.id) === true;

    clientView.select.disabled = !clientsHaveNodes;
    clientView.writeButton.disabled = blocked;
    clientView.readButton.disabled = blocked;
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
  clientsHaveNodes = sorted.length > 0;

  for (const clientView of clientViews) {
    const previous = clientView.select.value;

    clientView.select.innerHTML = sorted
      .map((nodeId) => `<option value="${nodeId}">Node ${nodeId}</option>`)
      .join("");

    if (sorted.includes(previous)) {
      clientView.select.value = previous;
    }
  }
}

const clientsContainer = document.getElementById("clients");

if (clientsContainer) {
  for (const id of CLIENT_IDS) {
    clientsContainer.appendChild(createClient(id));
  }
  updateClientNodeOptions([]);
}

function usesAmount() {
  return currentCrdt === "gcounter" || currentCrdt === "gcounter-op";
}
