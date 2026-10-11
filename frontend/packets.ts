type MessageKind =
  | "gossip"
  | "gossip_ack"
  | "write_request"
  | "write_ack"
  | "read_request"
  | "read_response";

type Endpoint = { kind: "node" | "client"; id: string };

type InFlightMessage = {
  id: string;
  kind: MessageKind;
  from: Endpoint;
  to: Endpoint;
  sentAt: number;
  deliverAt: number;
  clientId?: string;
  seq?: number;
};

type ClientResult = {
  kind: "write_ack" | "read_response";
  nodeId: string;
  value: number;
  at: number;
};

type ClientSnapshot = { id: string; locked: boolean; result: ClientResult | null };

type SimulatorState = {
  now: number;
  state: Record<string, Record<string, number>>;
  clients: ClientSnapshot[];
  inFlight: InFlightMessage[];
};

const PACKET_RADIUS = 10;
const CLIENT_PACKET_OFFSET = 12;
const SPAWN_ANIMATION_BASE_MS = 400;

let currentInFlight: InFlightMessage[] = [];
let currentNow = 0;
let latestSimulatorState: SimulatorState | null = null;

function packetsLayer(network: Network) {
  return network.stage.querySelector<SVGSVGElement>(".node-network-packets");
}

function messageProgress(message: InFlightMessage, tick: number) {
  if (tick >= message.deliverAt) {
    return 1;
  }

  const travelTicks = message.deliverAt - 1 - message.sentAt;

  if (travelTicks <= 0) {
    return 0;
  }

  return Math.min(1, Math.max(0, (tick - message.sentAt) / travelTicks));
}

function endpointCenter(network: Network, endpoint: Endpoint): Point | null {
  if (endpoint.kind === "client") {
    return getClientStagePoint(endpoint.id);
  }

  const card = findCard(network, endpoint.id);
  return card ? getCardCenter(card) : null;
}

function endpointBorder(network: Network, endpoint: Endpoint, toward: Point): Point | null {
  if (endpoint.kind === "client") {
    const center = getClientStagePoint(endpoint.id);

    if (!center) {
      return null;
    }

    const dx = toward.x - center.x;
    const dy = toward.y - center.y;
    const distance = Math.hypot(dx, dy);
    const offset = CLIENT_PACKET_OFFSET / (view?.scale ?? 1);

    if (distance === 0) {
      return center;
    }

    return { x: center.x + (dx / distance) * offset, y: center.y + (dy / distance) * offset };
  }

  const card = findCard(network, endpoint.id);
  return card ? getBorderPoint(card, toward) : null;
}

function messageEnds(network: Network, message: InFlightMessage): { start: Point; end: Point } | null {
  const fromCenter = endpointCenter(network, message.from);
  const toCenter = endpointCenter(network, message.to);

  if (!fromCenter || !toCenter) {
    return null;
  }

  const start = endpointBorder(network, message.from, toCenter);
  const end = endpointBorder(network, message.to, fromCenter);

  return start && end ? { start, end } : null;
}

function lerpPoint(a: Point, b: Point, t: number): Point {
  return { x: a.x + (b.x - a.x) * t, y: a.y + (b.y - a.y) * t };
}

function messagePosition(network: Network, message: InFlightMessage, tick: number): Point | null {
  const ends = messageEnds(network, message);
  return ends ? lerpPoint(ends.start, ends.end, messageProgress(message, tick)) : null;
}

function createPacket(message: InFlightMessage): SVGCircleElement {
  const packet = document.createElementNS(SVG_NAMESPACE, "circle");

  packet.classList.add("node-network-packet", `node-network-packet--${message.kind}`);
  packet.dataset.messageId = message.id;
  packet.setAttribute("r", `${PACKET_RADIUS}`);

  if (message.clientId) {
    packet.style.fill = clientColor(message.clientId);
  }

  return packet;
}

function setPacketPosition(packet: SVGCircleElement, position: Point) {
  packet.setAttribute("cx", `${position.x}`);
  packet.setAttribute("cy", `${position.y}`);
}

function findPacket(network: Network, messageId: string) {
  return network.stage.querySelector<SVGCircleElement>(
    `.node-network-packet[data-message-id="${messageId}"]`,
  );
}

function drawInFlightPackets(network: Network) {
  const layer = packetsLayer(network);

  if (!layer) {
    return;
  }

  layer.replaceChildren();

  for (const message of currentInFlight) {
    const position = messagePosition(network, message, currentNow);

    if (!position) {
      continue;
    }

    const packet = createPacket(message);
    setPacketPosition(packet, position);
    layer.append(packet);
  }
}

function redrawPacketsAfterViewChange() {
  if (activeNetwork && !isAdvancing) {
    drawInFlightPackets(activeNetwork);
  }
}

async function playTickPacketAnimations(before: SimulatorState) {
  const network = activeNetwork;

  if (!network) {
    return;
  }

  const tick = before.now + 1;
  const animations: Promise<void>[] = [];

  for (const message of before.inFlight) {
    const ends = messageEnds(network, message);
    const packet = findPacket(network, message.id);

    if (!ends || !packet) {
      continue;
    }

    const from = lerpPoint(ends.start, ends.end, messageProgress(message, before.now));
    const to = lerpPoint(ends.start, ends.end, messageProgress(message, tick));

    if (message.deliverAt === tick) {
      animations.push(playDelivery(network, message, packet, from, ends.end));
    } else {
      animations.push(animatePacket(packet, from, to, gossipSendAnimationDuration(), PACKET_RADIUS, false));
    }
  }

  await Promise.all(animations);
}

async function playDelivery(
  network: Network,
  message: InFlightMessage,
  packet: SVGCircleElement,
  from: Point,
  border: Point,
) {
  const target = endpointCenter(network, message.to);

  if (Math.hypot(border.x - from.x, border.y - from.y) > 1) {
    await animatePacket(packet, from, border, gossipSendAnimationDuration(), PACKET_RADIUS, false);
  }

  if (!target) {
    return;
  }

  const receiver = setReceiving(network, message.to, true);
  await animatePacket(packet, border, target, gossipReceiveAnimationDuration(), 0);
  receiver();
}

function setReceiving(network: Network, endpoint: Endpoint, receiving: boolean) {
  const element =
    endpoint.kind === "client"
      ? document.querySelector<HTMLElement>(`.client[data-client-id="${endpoint.id}"]`)
      : findCard(network, endpoint.id);

  const className = endpoint.kind === "client" ? "client--receiving" : "node-card--receiving";

  element?.classList.toggle(className, receiving);

  return () => element?.classList.remove(className);
}

function playSpawnAnimations(knownMessages: InFlightMessage[]) {
  const network = activeNetwork;

  if (!network) {
    return;
  }

  const knownIds = new Set(knownMessages.map((message) => message.id));

  for (const message of currentInFlight) {
    if (knownIds.has(message.id)) {
      continue;
    }

    const packet = findPacket(network, message.id);

    if (!packet) {
      continue;
    }

    const position = {
      x: Number(packet.getAttribute("cx")),
      y: Number(packet.getAttribute("cy")),
    };

    packet.setAttribute("r", "0");
    void animatePacketRadius(packet, position, scaled(SPAWN_ANIMATION_BASE_MS));
  }
}

function animatePacketRadius(packet: SVGCircleElement, position: Point, duration: number) {
  return animatePacket(packet, position, position, duration, PACKET_RADIUS, false, 0);
}

function animatePacket(
  packet: SVGCircleElement,
  from: Point,
  to: Point,
  duration: number,
  endRadius = PACKET_RADIUS,
  eased = true,
  startRadius = PACKET_RADIUS,
) {
  return new Promise<void>((resolve) => {
    const startedAt = performance.now();

    function move(now: number) {
      const progress = Math.min((now - startedAt) / duration, 1);
      const easedProgress = eased ? 1 - Math.pow(1 - progress, 3) : progress;

      setPacketPosition(packet, {
        x: from.x + (to.x - from.x) * easedProgress,
        y: from.y + (to.y - from.y) * easedProgress,
      });
      packet.setAttribute("r", `${startRadius + (endRadius - startRadius) * easedProgress}`);

      if (progress < 1) {
        requestAnimationFrame(move);
      } else {
        resolve();
      }
    }

    requestAnimationFrame(move);
  });
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
