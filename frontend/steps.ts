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
  const batch = await getNextStepBatch();

  if (batch.length === 0) {
    return false;
  }

  const needsPreviousState = batch.some((step) => step.type === "receive_gossip");
  const previousState = needsPreviousState ? await getStateAPI<SimulatorState>() : null;

  await Promise.all(batch.map(playStepAnimation));

  if (!await executeNextStepAPI()) {
    return false;
  }

  await renderState();
  recordTimeGraphBatch(toTimeGraphOperations(batch));

  if (previousState) {
    await nextFrame();
    for (const step of batch) {
      if (step.type === "receive_gossip" && step.to) {
        highlightReceivedState(previousState, step.to);
      }
    }
  }

  return true;
}

function playStepAnimation(step: Step): Promise<void> {
  if (step.type === "send_gossip" && step.from && step.to) {
    return playGossipSendAnimation(step.from, step.to);
  }

  if (step.type === "receive_gossip" && step.from && step.to) {
    return playGossipReceiveAnimation(step.from, step.to);
  }

  return Promise.resolve();
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

const PACKET_RADIUS = 10;

const packetsAtBorder = new Map<string, { from: string; to: string }>();

function packetKey(from: string, to: string) {
  return `${from}->${to}`;
}

function clearPendingPackets() {
  packetsAtBorder.clear();
}

function drawPendingPackets(network: Network) {
  const packets = network.stage.querySelector<SVGSVGElement>(".node-network-packets");

  if (!packets) {
    return;
  }

  packets.replaceChildren();

  for (const { from, to } of packetsAtBorder.values()) {
    const sourceCard = findCard(network, from);
    const targetCard = findCard(network, to);

    if (!sourceCard || !targetCard) {
      continue;
    }

    const packet = createPacket(from, to);
    setPacketPosition(packet, getBorderPoint(targetCard, getCardCenter(sourceCard)));
    packets.append(packet);
  }
}

async function playGossipSendAnimation(from: string, to: string) {
  const network = activeNetwork;

  if (!network) {
    return;
  }

  const sourceCard = findCard(network, from);
  const targetCard = findCard(network, to);
  const links = network.stage.querySelector<SVGSVGElement>(".node-network-links");
  const packets = network.stage.querySelector<SVGSVGElement>(".node-network-packets");

  if (!sourceCard || !targetCard || !links || !packets) {
    return;
  }

  const source = getCardCenter(sourceCard);
  const target = getCardCenter(targetCard);
  const activeLink = document.createElementNS(SVG_NAMESPACE, "line");

  activeLink.classList.add("node-network-active-link");
  activeLink.setAttribute("x1", `${source.x}`);
  activeLink.setAttribute("y1", `${source.y}`);
  activeLink.setAttribute("x2", `${target.x}`);
  activeLink.setAttribute("y2", `${target.y}`);

  const packet = createPacket(from, to);
  links.append(activeLink);
  packets.append(packet);
  sourceCard.classList.add("node-card--sending");

  await animatePacket(
    packet,
    getBorderPoint(sourceCard, target),
    getBorderPoint(targetCard, source),
    gossipSendAnimationDuration(),
  );

  sourceCard.classList.remove("node-card--sending");
  activeLink.remove();
  packetsAtBorder.set(packetKey(from, to), { from, to });
}

async function playGossipReceiveAnimation(from: string, to: string) {
  const network = activeNetwork;
  const key = packetKey(from, to);

  if (!network) {
    packetsAtBorder.delete(key);
    return;
  }

  const sourceCard = findCard(network, from);
  const targetCard = findCard(network, to);

  if (!targetCard) {
    packetsAtBorder.delete(key);
    return;
  }

  targetCard.classList.add("node-card--receiving");

  const target = getCardCenter(targetCard);
  const packet =
    network.stage.querySelector<SVGCircleElement>(`.node-network-packet[data-key="${key}"]`) ??
    (sourceCard ? spawnPacketAtBorder(network, from, to, sourceCard, targetCard) : null);

  if (!packet) {
    packetsAtBorder.delete(key);
    await wait(gossipReceiveAnimationDuration());
    return;
  }

  const start = {
    x: Number(packet.getAttribute("cx")),
    y: Number(packet.getAttribute("cy")),
  };

  await animatePacket(packet, start, target, gossipReceiveAnimationDuration(), 0);

  packet.remove();
  packetsAtBorder.delete(key);
}

function spawnPacketAtBorder(
  network: Network,
  from: string,
  to: string,
  sourceCard: HTMLElement,
  targetCard: HTMLElement,
): SVGCircleElement | null {
  const packets = network.stage.querySelector<SVGSVGElement>(".node-network-packets");

  if (!packets) {
    return null;
  }

  const packet = createPacket(from, to);
  setPacketPosition(packet, getBorderPoint(targetCard, getCardCenter(sourceCard)));
  packets.append(packet);

  return packet;
}

function createPacket(from: string, to: string): SVGCircleElement {
  const packet = document.createElementNS(SVG_NAMESPACE, "circle");

  packet.classList.add("node-network-packet");
  packet.dataset.key = packetKey(from, to);
  packet.setAttribute("r", `${PACKET_RADIUS}`);

  return packet;
}

function setPacketPosition(packet: SVGCircleElement, position: Point) {
  packet.setAttribute("cx", `${position.x}`);
  packet.setAttribute("cy", `${position.y}`);
}

function findCard(network: Network, nodeId: string): HTMLElement | undefined {
  return network.cards.find((card) => card.dataset.nodeId === nodeId);
}

function getCardCenter(card: HTMLElement): Point {
  return {
    x: Number.parseFloat(card.style.left),
    y: Number.parseFloat(card.style.top),
  };
}

function getBorderPoint(card: HTMLElement, toward: Point): Point {
  const center = getCardCenter(card);
  const dx = toward.x - center.x;
  const dy = toward.y - center.y;
  const distance = Math.hypot(dx, dy);

  if (distance === 0) {
    return center;
  }

  const scale = Math.min(
    dx === 0 ? Infinity : card.offsetWidth / 2 / Math.abs(dx),
    dy === 0 ? Infinity : card.offsetHeight / 2 / Math.abs(dy),
  );

  return {
    x: center.x + dx * scale + (dx / distance) * PACKET_RADIUS,
    y: center.y + dy * scale + (dy / distance) * PACKET_RADIUS,
  };
}

function animatePacket(
  packet: SVGCircleElement,
  from: Point,
  to: Point,
  duration: number,
  endRadius = PACKET_RADIUS,
) {
  return new Promise<void>((resolve) => {
    const startedAt = performance.now();

    function move(now: number) {
      const progress = Math.min((now - startedAt) / duration, 1);
      const easedProgress = 1 - Math.pow(1 - progress, 3);

      setPacketPosition(packet, {
        x: from.x + (to.x - from.x) * easedProgress,
        y: from.y + (to.y - from.y) * easedProgress,
      });
      packet.setAttribute("r", `${PACKET_RADIUS + (endRadius - PACKET_RADIUS) * easedProgress}`);

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
    gossipReceiveAnimationDuration() + valueChangeHighlightDuration(),
  );
}

function wait(duration: number) {
  return new Promise<void>((resolve) => window.setTimeout(resolve, duration));
}

function nextFrame() {
  return new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
}
