type Step = {
  stepId: number;
  type: string;
  replicaId?: string;
  clientId?: string;
  from?: string;
  to?: string;
};

let isAdvancing = false;

async function advanceOneStep(): Promise<boolean> {
  if (isAdvancing) {
    return false;
  }

  isAdvancing = true;

  try {
    return await runNextBatch();
  } finally {
    isAdvancing = false;
  }
}

async function runNextBatch(): Promise<boolean> {
  const before = await getStateAPI<SimulatorState>();

  if (!before) {
    return false;
  }

  const batch = await getNextStepBatch(before.now);

  if (batch === null) {
    return false;
  }

  await playTickPacketAnimations(before);

  if (!await executeNextStepAPI()) {
    return false;
  }

  await renderState();
  recordTimeGraphBatch(toTimeGraphOperations(batch));

  await nextFrame();
  playSpawnAnimations(before.inFlight);
  highlightChangedNodes(before);

  return true;
}

function highlightChangedNodes(previousState: SimulatorState) {
  for (const card of activeNetwork?.cards ?? []) {
    const nodeId = card.dataset.nodeId;

    if (nodeId && hasStateChanged(previousState, card, nodeId)) {
      highlightReceivedState(previousState, nodeId);
    }
  }
}

function hasStateChanged(previousState: SimulatorState, card: HTMLElement, nodeId: string) {
  const previousCounts = previousState.state[nodeId];

  if (!previousCounts) {
    return false;
  }

  return Array.from(card.querySelectorAll<HTMLElement>("[data-replica-id]")).some(
    (cell) => previousCounts[cell.dataset.replicaId ?? ""] !== Number(cell.textContent),
  );
}

function toTimeGraphOperations(batch: Step[]): TimeGraphOperation[] {
  const operations: TimeGraphOperation[] = [];
  const sendTargets = new Map<string, string[]>();

  for (const step of batch) {
    if (step.type === "send_gossip" && step.from && step.to) {
      sendTargets.set(step.from, [...(sendTargets.get(step.from) ?? []), step.to]);
      continue;
    }

    const operation = toTimeGraphOperation(step);
    if (operation) {
      operations.push(operation);
    }
  }

  for (const [from, targets] of sendTargets) {
    operations.push({
      nodeId: from,
      kind: "send",
      targets: targets.sort(),
      broadcast: isBroadcast(from, targets),
    });
  }

  return operations;
}

function isBroadcast(from: string, targets: string[]): boolean {
  const otherNodes = (activeNetwork?.cards ?? [])
    .map((card) => card.dataset.nodeId)
    .filter((nodeId): nodeId is string => nodeId !== undefined && nodeId !== from);

  return otherNodes.length > 1 && otherNodes.every((nodeId) => targets.includes(nodeId));
}

function toTimeGraphOperation(step: Step): TimeGraphOperation | null {
  const clientRow = step.clientId ? clientRowId(step.clientId) : undefined;

  switch (step.type) {
    case "create_node":
      return step.replicaId ? { nodeId: step.replicaId, kind: "create", value: 0 } : null;

    case "increment":
      return step.replicaId
        ? { nodeId: step.replicaId, kind: "increment", value: getOwnCounter(step.replicaId) }
        : null;

    case "remove_node":
      return step.replicaId ? { nodeId: step.replicaId, kind: "remove" } : null;

    case "receive_write_request":
    case "receive_read_request":
      return step.replicaId
        ? {
            nodeId: step.replicaId,
            kind: "arrive",
            label: `${step.type === "receive_write_request" ? "W" : "R"}←${step.clientId ?? ""}`,
          }
        : null;

    case "read":
      return step.replicaId
        ? { nodeId: step.replicaId, kind: "read", value: getNodeSum(step.replicaId) }
        : null;

    case "send_ack":
    case "send_response":
      return step.replicaId
        ? {
            nodeId: step.replicaId,
            kind: "ack",
            label: `${step.type === "send_ack" ? "ack" : "val"}→${step.clientId ?? ""}`,
          }
        : null;

    case "receive_write_ack":
      return clientRow ? { nodeId: clientRow, kind: "response", label: "ack" } : null;

    case "receive_read_response":
      return clientRow
        ? { nodeId: clientRow, kind: "response", value: clientResults.get(step.clientId ?? "")?.value }
        : null;

    case "receive_gossip":
      return step.to ? { nodeId: step.to, kind: "receive", label: `←${step.from ?? ""}` } : null;

    case "merge":
      return step.to ? { nodeId: step.to, kind: "merge", value: getNodeSum(step.to) } : null;

    default:
      return null;
  }
}

function getOwnCounter(nodeId: string): number | undefined {
  const card = activeNetwork?.cards.find((item) => item.dataset.nodeId === nodeId);
  const text = card?.querySelector<HTMLElement>(`[data-replica-id="${nodeId}"]`)?.textContent;
  return text ? Number(text) : undefined;
}

function getNodeSum(nodeId: string): number | undefined {
  const card = activeNetwork?.cards.find((item) => item.dataset.nodeId === nodeId);
  const text = card?.querySelector<HTMLElement>(".node-card-sum")?.textContent;
  return text ? Number(text) : undefined;
}

async function getNextStepBatch(now: number): Promise<Step[] | null> {
  const response = await getStepsAPI<{ steps?: Step[] }>();
  const steps = response?.steps ?? [];

  if (steps.length === 0) {
    return null;
  }

  const stepId = steps[0].stepId;

  if (stepId !== now + 1) {
    return [];
  }

  return steps.filter((step) => step.stepId === stepId);
}

function highlightReceivedState(previousState: SimulatorState, nodeId: string) {
  const card = activeNetwork?.cards.find(
    (item) => item.dataset.nodeId === nodeId,
  );

  if (!card) {
    return;
  }

  const previousCounts = previousState.state[nodeId] ?? {};

  card.classList.add("node-card--receiving");
  for (const cell of card.querySelectorAll<HTMLElement>("[data-replica-id]")) {
    const replicaId = cell.dataset.replicaId;
    const currentValue = Number(cell.textContent);

    if (replicaId && previousCounts[replicaId] !== currentValue) {
      cell.classList.add("node-card-value-changed");
    }
  }

  window.setTimeout(
    () => card.classList.remove("node-card--receiving"),
    gossipReceiveAnimationDuration() + valueChangeHighlightDuration(),
  );
}

function wait(duration: number) {
  return new Promise<void>((resolve) => window.setTimeout(resolve, duration));
}

function nextFrame() {
  return new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
}
