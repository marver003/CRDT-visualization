const CLIENT_COUNT = 2;

interface ClientView {
  select: HTMLSelectElement;
  writeButton: HTMLButtonElement;
  readButton: HTMLButtonElement;
  result: HTMLElement;
}

const clientViews: ClientView[] = [];

function createClient(index: number): HTMLElement {
  const client = document.createElement("div");
  client.className = "client";
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
      <strong class="client-title">Client ${index + 1}</strong>
      <select class="client-select" aria-label="Node for client ${index + 1}"></select>
      <div class="client-actions">
        <button class="client-button client-write" type="button" title="Send a write (increment) request">Write</button>
        <button class="client-button client-read" type="button" title="Send a read request">Read</button>
      </div>
      <div class="client-result" aria-live="polite"></div>
    </div>
  `;

  const view: ClientView = {
    select: client.querySelector(".client-select") as HTMLSelectElement,
    writeButton: client.querySelector(".client-write") as HTMLButtonElement,
    readButton: client.querySelector(".client-read") as HTMLButtonElement,
    result: client.querySelector(".client-result") as HTMLElement,
  };

  view.writeButton.addEventListener("click", () => {
    if (view.select.value) {
      incrementNode(view.select.value);
    }
  });
  view.readButton.addEventListener("click", () => readNode(view));

  clientViews.push(view);
  return client;
}

async function readNode(view: ClientView) {
  const nodeId = view.select.value;

  if (!nodeId) {
    return;
  }

  const response = await getStateAPI<{ state: Record<string, Record<string, number>>; }>();

  if (!response) {
    return;
  }

  const counts = response.state[nodeId];

  if (!counts) {
    view.result.textContent = `Node ${nodeId} no longer exists`;
    return;
  }

  const sum = Object.values(counts).reduce((total, value) => total + value, 0);
  view.result.textContent = `Node ${nodeId} = ${sum}`;
}

function updateClientNodeOptions(nodeIds: string[]) {
  const sorted = [...nodeIds].sort();

  for (const view of clientViews) {
    const previous = view.select.value;

    view.select.innerHTML = sorted
      .map((nodeId) => `<option value="${nodeId}">Node ${nodeId}</option>`)
      .join("");

    if (sorted.includes(previous)) {
      view.select.value = previous;
    }

    const disabled = sorted.length === 0;
    view.select.disabled = disabled;
    view.writeButton.disabled = disabled;
    view.readButton.disabled = disabled;
  }
}

const clientsContainer = document.getElementById("clients");

if (clientsContainer) {
  for (let i = 0; i < CLIENT_COUNT; i++) {
    clientsContainer.appendChild(createClient(i));
  }
  updateClientNodeOptions([]);
}
