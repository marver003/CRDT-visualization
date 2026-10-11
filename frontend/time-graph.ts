type TimelineEvent = {
  tick: number;
  type: string;
  node?: string;
  clientId?: string;
  seq?: number;
  from?: string;
};

type TimelineEdge = {
  id: string;
  fromLane: string;
  toLane: string;
  sentAt: number;
  deliverAt: number;
  clientId?: string;
};

type BucketKind = "write" | "read" | "merge" | "send";

type Bucket = {
  node: string;
  kind: BucketKind;
  start: number;
  end: number;
  processedAt?: number;
  rightLeg: boolean;
  clients: string[];
  seqs: number[];
};

type Lane = {
  id: string;
  label: string;
  axisY: number;
  height: number;
};

const TICK_WIDTH = 60;
const BUCKET_HEIGHT = 34;
const LANE_TOP = 12;
const LANE_BOTTOM = 16;
const DASH = 9;
const LABEL_ROW = 16;
const LABEL_OVERLAP = 30;
const LEG_DX = 14;
const HEADER_HEIGHT = 8;
const LEFT_PAD = 28;
const LANES_WIDTH = 52;

const timeGraphContainer = document.getElementById("time-graph-container") as HTMLElement;
const timeGraphLanes = document.getElementById("time-graph-lanes") as HTMLElement;
const timeGraphToggle = document.getElementById("time-graph-toggle") as HTMLButtonElement;
const timeGraphWidget = document.querySelector(".time-graph-widget") as HTMLElement;

let timelineEvents: TimelineEvent[] = [];
const timelineEdges = new Map<string, TimelineEdge>();
const timelineSums = new Map<number, Record<string, number>>();
let timelineBaseTick: number | null = null;
let timelineLastTick = 0;

function resetTimeGraph() {
  timelineEvents = [];
  timelineEdges.clear();
  timelineSums.clear();
  timelineBaseTick = null;
  timelineLastTick = 0;
  renderTimeGraph();
}

function sumsOf(state: SimulatorState["state"]): Record<string, number> {
  const sums: Record<string, number> = {};

  for (const [nodeId, counts] of Object.entries(state)) {
    sums[nodeId] = Object.values(counts).reduce((total, value) => total + value, 0);
  }

  return sums;
}

function rememberTimelineEdges(messages: InFlightMessage[]) {
  for (const message of messages) {
    if (message.kind !== "gossip" || timelineEdges.has(message.id)) {
      continue;
    }

    timelineEdges.set(message.id, {
      id: message.id,
      fromLane: message.from.id,
      toLane: message.to.id,
      sentAt: message.sentAt,
      deliverAt: message.deliverAt,
      clientId: message.clientId,
    });
  }
}

function recordTimelineTick(before: SimulatorState, after: SimulatorState, batch: Step[]) {
  if (timelineBaseTick === null) {
    timelineBaseTick = before.now;
    timelineSums.set(before.now, sumsOf(before.state));
  }

  const tick = after.now;
  timelineSums.set(tick, sumsOf(after.state));
  timelineLastTick = tick;

  for (const step of batch) {
    timelineEvents.push(toTimelineEvent(tick, step));
  }

  renderTimeGraph();
}

function toTimelineEvent(tick: number, step: Step): TimelineEvent {
  const event: TimelineEvent = { tick, type: step.type, clientId: step.clientId, seq: step.seq };

  switch (step.type) {
    case "send_gossip":
    case "send_gossip_ack":
      event.node = step.from;
      break;
    case "receive_gossip":
    case "receive_gossip_ack":
    case "merge":
      event.node = step.to;
      event.from = step.from;
      break;
    default:
      event.node = step.replicaId;
  }

  return event;
}

function findBucket(buckets: Bucket[], node: string, kind: BucketKind, seq: number | undefined, open = false) {
  for (let i = buckets.length - 1; i >= 0; i--) {
    const bucket = buckets[i];

    if (bucket.node !== node || bucket.kind !== kind || bucket.seqs[0] !== seq) {
      continue;
    }

    if (!open || bucket.processedAt === undefined) {
      return bucket;
    }
  }

  return undefined;
}

function newBucket(buckets: Bucket[], event: TimelineEvent, kind: BucketKind): Bucket {
  const bucket: Bucket = {
    node: event.node ?? "",
    kind,
    start: event.tick,
    end: event.tick,
    rightLeg: false,
    clients: event.clientId ? [event.clientId] : [],
    seqs: event.seq ? [event.seq] : [],
  };

  buckets.push(bucket);
  return bucket;
}

function buildBuckets(): Bucket[] {
  const buckets: Bucket[] = [];

  for (const event of timelineEvents) {
    if (!event.node) {
      continue;
    }

    const seq = event.seq || undefined;

    switch (event.type) {
      case "receive_write_request":
        newBucket(buckets, event, "write");
        break;

      case "receive_read_request":
        newBucket(buckets, event, "read");
        break;

      case "increment": {
        const bucket = findBucket(buckets, event.node, "write", seq) ?? newBucket(buckets, event, "write");
        bucket.processedAt = event.tick;
        bucket.end = Math.max(bucket.end, event.tick);
        break;
      }

      case "read": {
        const bucket = findBucket(buckets, event.node, "read", seq) ?? newBucket(buckets, event, "read");
        bucket.processedAt = event.tick;
        bucket.end = Math.max(bucket.end, event.tick);
        break;
      }

      case "send_ack":
      case "send_response": {
        const kind = event.type === "send_ack" ? "write" : "read";
        const bucket = findBucket(buckets, event.node, kind, seq) ?? newBucket(buckets, event, kind);
        bucket.end = Math.max(bucket.end, event.tick);
        bucket.rightLeg = true;
        break;
      }

      case "send_gossip": {
        const bucket = seq ? findBucket(buckets, event.node, "write", seq) : undefined;

        if (bucket) {
          bucket.end = Math.max(bucket.end, event.tick);
          bucket.rightLeg = true;
        } else if (!seq) {
          const sender = newBucket(buckets, event, "send");
          sender.rightLeg = true;
        }
        break;
      }

      case "receive_gossip":
        newBucket(buckets, event, "merge");
        break;

      case "send_gossip_ack": {
        const bucket = findBucket(buckets, event.node, "merge", seq);

        if (bucket) {
          bucket.end = Math.max(bucket.end, event.tick);
          bucket.rightLeg = true;
        }
        break;
      }

      case "merge": {
        const bucket = findBucket(buckets, event.node, "merge", seq, true) ?? newBucket(buckets, event, "merge");
        bucket.processedAt = event.tick;
        bucket.end = Math.max(bucket.end, event.tick);
        break;
      }
    }
  }

  return buckets;
}

function clusterBuckets(buckets: Bucket[]): Bucket[][] {
  const byNode = new Map<string, Bucket[]>();

  for (const bucket of buckets) {
    byNode.set(bucket.node, [...(byNode.get(bucket.node) ?? []), bucket]);
  }

  const clusters: Bucket[][] = [];

  for (const nodeBuckets of byNode.values()) {
    nodeBuckets.sort((a, b) => a.start - b.start || a.end - b.end);
    let current: Bucket[] = [];
    let currentEnd = -Infinity;

    for (const bucket of nodeBuckets) {
      if (current.length > 0 && bucket.start < currentEnd) {
        current.push(bucket);
        currentEnd = Math.max(currentEnd, bucket.end);
        continue;
      }

      if (current.length > 0) {
        clusters.push(current);
      }

      current = [bucket];
      currentEnd = bucket.end;
    }

    if (current.length > 0) {
      clusters.push(current);
    }
  }

  for (const cluster of clusters) {
    cluster.sort((a, b) => (a.seqs[0] ?? Infinity) - (b.seqs[0] ?? Infinity));
  }

  return clusters;
}

function escapeXml(text: string) {
  return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

function nodePresence(nodeId: string, base: number, last: number) {
  let first: number | null = null;
  let final: number | null = null;

  for (let tick = base; tick <= last; tick++) {
    if (timelineSums.get(tick)?.[nodeId] !== undefined) {
      first ??= tick;
      final = tick;
    }
  }

  return first === null || final === null ? null : { first, final };
}

function renderTimeGraph() {
  if (timelineBaseTick === null) {
    timeGraphLanes.replaceChildren();
    timeGraphContainer.innerHTML = '<p class="time-graph-empty">No steps yet</p>';
    return;
  }

  const base = timelineBaseTick;
  const last = timelineLastTick;
  const clusters = clusterBuckets(buildBuckets());

  const nodeIds = new Set<string>();
  for (const sums of timelineSums.values()) {
    for (const nodeId of Object.keys(sums)) {
      nodeIds.add(nodeId);
    }
  }

  const laneHeight = LANE_TOP + BUCKET_HEIGHT + LANE_BOTTOM;
  const lanes: Lane[] = [...nodeIds].sort().map((nodeId, index) => ({
    id: nodeId,
    label: nodeId,
    axisY: index * laneHeight + LANE_TOP + BUCKET_HEIGHT,
    height: laneHeight,
  }));

  const totalHeight = lanes.length * laneHeight + 4;
  const x = (tick: number) => LEFT_PAD + (tick - base) * TICK_WIDTH;
  const width = x(last) + TICK_WIDTH;
  const laneById = new Map(lanes.map((lane) => [lane.id, lane]));

  const defs = [...CLIENT_IDS, undefined]
    .map(
      (clientId) =>
        `<marker id="arrow-${clientId ?? "n"}" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M0 0 L10 5 L0 10 z" fill="${clientColor(clientId)}"/></marker>`,
    )
    .join("");

  const parts: string[] = [];

  for (const lane of lanes) {
    const presence = nodePresence(lane.id, base, last);

    if (!presence) {
      continue;
    }

    parts.push(`<line class="tg-axis" x1="${x(presence.first) - 10}" y1="${lane.axisY}" x2="${x(presence.final) + TICK_WIDTH * 0.6}" y2="${lane.axisY}"/>`);

    for (let tick = presence.first; tick <= presence.final; tick++) {
      parts.push(`<line class="tg-axis" x1="${x(tick)}" y1="${lane.axisY - 4}" x2="${x(tick)}" y2="${lane.axisY + 4}"/>`);
    }
  }

  for (const edge of timelineEdges.values()) {
    const from = laneById.get(edge.fromLane);
    const to = laneById.get(edge.toLane);

    if (!from || !to || edge.deliverAt > last || edge.sentAt < base) {
      continue;
    }

    parts.push(
      `<line class="tg-edge" x1="${x(edge.sentAt)}" y1="${from.axisY}" x2="${x(edge.deliverAt)}" y2="${to.axisY}" stroke="${clientColor(edge.clientId)}" marker-end="url(#arrow-${edge.clientId ?? "n"})"/>`,
    );
  }

  for (const cluster of clusters) {
    const lane = laneById.get(cluster[0].node);

    if (lane) {
      parts.push(renderCluster(cluster, lane, x));
    }
  }

  const svg = `<svg class="tg-svg" width="${width}" height="${totalHeight}" viewBox="0 0 ${width} ${totalHeight}"><defs>${defs}</defs>${parts.join("")}</svg>`;
  const laneLabels = lanes
    .map(
      (lane) =>
        `<text class="tg-lane-label" x="${LANES_WIDTH / 2}" y="${lane.axisY + 7}" text-anchor="middle">${escapeXml(lane.label)}</text>`,
    )
    .join("");

  timeGraphLanes.innerHTML = `<svg width="${LANES_WIDTH}" height="${totalHeight}" viewBox="0 0 ${LANES_WIDTH} ${totalHeight}">${laneLabels}</svg>`;
  timeGraphContainer.innerHTML = svg;
  timeGraphContainer.scrollLeft = timeGraphContainer.scrollWidth;
}

function bucketColor(bucket: Bucket) {
  return bucket.clients.length > 0 ? clientColor(bucket.clients[0]) : TIMELINE_NEUTRAL_COLOR;
}

function renderCluster(cluster: Bucket[], lane: Lane, x: (tick: number) => number): string {
  const flatY = lane.axisY;
  const tipY = flatY - BUCKET_HEIGHT;
  const legLength = Math.hypot(LEG_DX, BUCKET_HEIGHT);
  const mixed = new Set(cluster.map(bucketColor)).size > 1;
  const period = DASH * cluster.length;
  const parts: string[] = [];
  const labels = new Map<string, { x: number; text: string; colors: Set<string> }>();

  cluster.forEach((bucket, index) => {
    const left = x(bucket.start);
    const right = Math.max(x(bucket.end), left + 12);
    const color = bucketColor(bucket);

    const outline =
      `M ${left - LEG_DX} ${tipY} L ${left} ${flatY} L ${right} ${flatY}` +
      (bucket.rightLeg ? ` L ${right + LEG_DX} ${tipY}` : "");
    const offset = (((left - legLength - index * DASH) % period) + period) % period;
    const dash = mixed ? ` stroke-dasharray="${DASH} ${period - DASH}" stroke-dashoffset="${offset}"` : "";

    const sum = bucket.processedAt !== undefined ? sumOfNode(bucket.node, bucket.processedAt) : undefined;
    const seqText = bucket.seqs.length > 0 ? ` #${bucket.seqs.join("/")}` : "";
    const what =
      bucket.kind === "write" ? "Write" : bucket.kind === "read" ? "Read" : bucket.kind === "merge" ? "Merge" : "Gossip";
    const detail = `${what}${seqText}, arrives t${bucket.start}${bucket.processedAt !== undefined ? `, processed t${bucket.processedAt}` : ""}${bucket.rightLeg ? `, sends t${bucket.end}` : ""}${sum !== undefined ? `, sum ${sum}` : ""}`;

    parts.push(
      `<path d="${outline}" fill="none" stroke="${color}" stroke-width="3" stroke-linejoin="round"${dash}><title>${escapeXml(detail)}</title></path>`,
    );

    if (bucket.processedAt !== undefined) {
      parts.push(`<circle cx="${x(bucket.processedAt)}" cy="${flatY}" r="4" fill="${color}"/>`);
    }

    if (sum !== undefined) {
      const labelX = (left + right) / 2;
      const key = `${Math.round(labelX)}|${sum}`;
      const label = labels.get(key) ?? { x: labelX, text: `${sum}`, colors: new Set<string>() };
      label.colors.add(color);
      labels.set(key, label);
    }
  });

  const placed: { x: number; row: number }[] = [];

  for (const label of labels.values()) {
    let row = 0;

    while (placed.some((item) => item.row === row && Math.abs(item.x - label.x) < LABEL_OVERLAP)) {
      row++;
    }

    placed.push({ x: label.x, row });
    const fill = label.colors.size === 1 ? [...label.colors][0] : "#202938";
    parts.push(
      `<text class="tg-bucket-sum" x="${label.x}" y="${flatY - 9 - row * LABEL_ROW}" text-anchor="middle" fill="${fill}">${label.text}</text>`,
    );
  }

  return `<g class="tg-bucket">${parts.join("")}</g>`;
}

function sumOfNode(nodeId: string, tick: number): number | undefined {
  return timelineSums.get(tick)?.[nodeId];
}

timeGraphToggle.addEventListener("click", () => {
  const collapsed = timeGraphWidget.classList.toggle("time-graph-widget--collapsed");
  timeGraphToggle.setAttribute("aria-expanded", String(!collapsed));
  timeGraphToggle.textContent = collapsed ? "▸" : "▾";
});
