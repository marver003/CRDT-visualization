// Step execution. There's no visible step queue/"Next Step" button anymore -
// the backend still queues every action instead of applying it instantly, so
// runQueuedSteps() auto-drains that queue right after each action, playing
// the gossip send/receive animations below for any step that needs them.
// (Merge is currently removed from the UI, so send_gossip/receive_gossip
// steps don't occur today - this stays ready for when merge comes back.)
//
// The backend can group several operations under one shared StepID and
// executes that whole group in a single /step call - meaning those
// operations are meant to happen simultaneously. /steps now only ever
// returns genuinely pending entries, sorted ascending by stepId, so the
// batch to run next is just "every entry sharing the lowest stepId".

type Step = {
  stepId: number;
  type: string;
  replicaId?: string;
  from?: string;
  to?: string;
};

type SimulatorState = {
  state: Record<string, Record<string, number>>;
};

const GOSSIP_ANIMATION_DURATION = 1_000;
const RECEIVE_ANIMATION_DURATION = 1_000;

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
  } finally {
    isRunningSteps = false;
  }
}

async function drainStepQueue() {
  while (true) {
    const batch = await getNextStepBatch();

    if (batch.length === 0) {
      return;
    }

    const needsPreviousState = batch.some((step) => step.type === "receive_gossip");
    const previousState = needsPreviousState ? await getStateAPI<SimulatorState>() : null;

    // Everything in a batch shares one stepId - the backend executes them
    // together in a single /step call, so their animations run together too
    // (Promise.all, not one-by-one) rather than staggered in sequence.
    await Promise.all(batch.map(playStepAnimation));

    if (!await executeNextStepAPI()) {
      return;
    }

    await renderState();
    recordTimeGraphBatch(batch.map(toTimeGraphOperation).filter((op): op is TimeGraphOperation => op !== null));

    if (previousState) {
      await nextFrame();
      for (const step of batch) {
        if (step.type === "receive_gossip" && step.to) {
          highlightReceivedState(previousState, step.to);
        }
      }
    }
  }
}

function playStepAnimation(step: Step): Promise<void> {
  if (step.type === "send_gossip" && step.from && step.to) {
    return playGossipSendAnimation(step.from, step.to);
  }

  if (step.type === "receive_gossip" && step.to) {
    return playGossipReceiveAnimation(step.to);
  }

  return Promise.resolve();
}

// Maps an executed step to what the Replication Timeline widget
// (time-graph.ts) should show for it. Reads the resulting own-counter
// straight from the just-rendered card rather than another API round trip.
function toTimeGraphOperation(step: Step): TimeGraphOperation | null {
  switch (step.type) {
    case "create_node":
      return step.replicaId ? { nodeId: step.replicaId, kind: "create", value: 0 } : null;

    case "increment":
      return step.replicaId
        ? { nodeId: step.replicaId, kind: "increment", value: getOwnCounter(step.replicaId) }
        : null;

    case "remove_node":
      return step.replicaId ? { nodeId: step.replicaId, kind: "remove" } : null;

    case "receive_gossip":
      return step.to ? { nodeId: step.to, kind: "receive", value: getOwnCounter(step.to) } : null;

    default:
      return null;
  }
}

function getOwnCounter(nodeId: string): number | undefined {
  const card = activeNetwork?.cards.find((item) => item.dataset.nodeId === nodeId);
  const text = card?.querySelector<HTMLElement>(`[data-replica-id="${nodeId}"]`)?.textContent;
  return text ? Number(text) : undefined;
}

// The next batch to run is every entry sharing the lowest stepId. /steps only
// returns pending entries, sorted ascending, so that's just a prefix filter.
async function getNextStepBatch(): Promise<Step[]> {
  const response = await getStepsAPI<{ steps?: Step[] }>();
  const steps = response?.steps ?? [];

  if (steps.length === 0) {
    return [];
  }

  const stepId = steps[0].stepId;
  return steps.filter((step) => step.stepId === stepId);
}

async function playGossipSendAnimation(from: string, to: string) {
  const network = activeNetwork;

  if (!network) {
    return;
  }

  const sourceCard = network.cards.find((card) => card.dataset.nodeId === from);
  const targetCard = network.cards.find((card) => card.dataset.nodeId === to);
  const links = network.stage.querySelector<SVGSVGElement>(
    ".node-network-links",
  );

  if (!sourceCard || !targetCard || !links) {
    return;
  }

  const source = getCardCenter(sourceCard);
  const target = getCardCenter(targetCard);
  const activeLink = document.createElementNS(
    "http://www.w3.org/2000/svg",
    "line",
  );
  const packet = document.createElementNS(
    "http://www.w3.org/2000/svg",
    "circle",
  );

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

async function playGossipReceiveAnimation(nodeId: string) {
  const card = activeNetwork?.cards.find(
    (item) => item.dataset.nodeId === nodeId,
  );

  if (!card) {
    return;
  }

  card.classList.add("node-card--receiving");
  await wait(RECEIVE_ANIMATION_DURATION);
}

function getCardCenter(card: HTMLElement) {
  return {
    x: Number.parseFloat(card.style.left),
    y: Number.parseFloat(card.style.top),
  };
}

function animatePacket(packet: SVGCircleElement, from: Point, to: Point) {
  return new Promise<void>((resolve) => {
    const startedAt = performance.now();

    function move(now: number) {
      const progress = Math.min(
        (now - startedAt) / GOSSIP_ANIMATION_DURATION,
        1,
      );
      const easedProgress = 1 - Math.pow(1 - progress, 3);
      const x = from.x + (to.x - from.x) * easedProgress;
      const y = from.y + (to.y - from.y) * easedProgress;

      packet.setAttribute("cx", `${x}`);
      packet.setAttribute("cy", `${y}`);

      if (progress < 1) {
        requestAnimationFrame(move);
      } else {
        resolve();
      }
    }

    requestAnimationFrame(move);
  });
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
    RECEIVE_ANIMATION_DURATION + 850,
  );
}

function wait(duration: number) {
  return new Promise<void>((resolve) => window.setTimeout(resolve, duration));
}

function nextFrame() {
  return new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
}
