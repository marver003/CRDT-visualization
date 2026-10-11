type Step = {
  stepId: number;
  type: string;
  replicaId?: string;
  clientId?: string;
  seq?: number;
  from?: string;
  to?: string;
};

let isAdvancing = false;

async function advanceOneStep(): Promise<boolean> {
  if (isAdvancing) {
    return false;
  }

  setAnimating(true);

  try {
    return await runNextBatch();
  } finally {
    setAnimating(false);
  }
}

function setAnimating(animating: boolean) {
  isAdvancing = animating;
  document.body.classList.toggle("is-animating", animating);
  refreshClientControls();
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
  if (latestSimulatorState) {
    recordTimelineTick(before, latestSimulatorState, batch);
  }

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
