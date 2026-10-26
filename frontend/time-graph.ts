type TimelineEvent = {
  tick: number;
  type: string;
  node?: string;
  clientId?: string;
  seq?: number;
  from?: string;
  replay?: boolean;
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
  row: number;
  replay: boolean;
  from?: string;
};

type Lane = {
  id: string;
  label: string;
  axisY: number;
  height: number;
};

const TICK_WIDTH_LEVELS = [15, 20, 30, 45, 60, 90, 120];
let tickWidthIndex = 2;
let TICK_WIDTH = TICK_WIDTH_LEVELS[tickWidthIndex];
const BUCKET_HEIGHT = 34;
const LANE_TOP = 10;
const LANE_BOTTOM = 16;
const LINE_GAP = 10;
const ROW_GAP = 10;
const BUCKET_MARGIN = 8;
const MIN_LANE_ROWS = 2;
const LEG_DX = 14;
const HEADER_HEIGHT = 8;
const LEFT_PAD = 28;
const LANES_WIDTH = 52;

const timeGraphContainer = document.getElementById("time-graph-container") as HTMLElement;
const timeGraphLanes = document.getElementById("time-graph-lanes") as HTMLElement;
const timeGraphToggle = document.getElementById("time-graph-toggle") as HTMLButtonElement;
const timeGraphNarrow = document.getElementById("time-graph-narrow") as HTMLButtonElement;
const timeGraphWiden = document.getElementById("time-graph-widen") as HTMLButtonElement;
const timeGraphWidget = document.querySelector(".time-graph-widget") as HTMLElement;

let timelineEvents: TimelineEvent[] = [];
const timelineEdges = new Map<string, TimelineEdge>();
const timelineSums = new Map<number, Record<string, number>>();
const timelineSets = new Map<number, Record<string, string[]>>();
const timelineDisabled = new Map<number, string[]>();
let timelineBaseTick: number | null = null;
let timelineRecording = false;
let timelineLastTick = 0;

function resetTimeGraph() {
  timelineEvents = [];
  timelineEdges.clear();
  timelineSums.clear();
  timelineSets.clear();
  timelineDisabled.clear();
  timelineBaseTick = null;
  timelineLastTick = 0;
  timelineRecording = false;
  renderTimeGraph();
}

function syncTimelinePreview() {
  if (timelineRecording || !latestSimulatorState) {
    return;
  }

  const { now, state, disabled } = latestSimulatorState;

  timelineSums.clear();
  timelineSets.clear();
  timelineDisabled.clear();
  timelineSums.set(now, sumsOf(state));
  timelineSets.set(now, setsOf(state));
  timelineDisabled.set(now, disabled ?? []);
  timelineBaseTick = now;
  timelineLastTick = now;
  renderTimeGraph();
}

function sumsOf(state: SimulatorState["state"]): Record<string, number> {
  const sums: Record<string, number> = {};

  for (const [nodeId, counts] of Object.entries(state)) {
    sums[nodeId] =
      typeof counts === "number"
        ? counts
        : Array.isArray(counts)
          ? counts.length
          : Object.values(counts).reduce((total, value) => total + value, 0);
  }

  return sums;
}

function setsOf(state: SimulatorState["state"]): Record<string, string[]> {
  const sets: Record<string, string[]> = {};

  for (const [nodeId, value] of Object.entries(state)) {
    if (Array.isArray(value)) {
      sets[nodeId] = value;
    }
  }

  return sets;
}

function gainedElements(nodeId: string, tick: number): string[] {
  const now = timelineSets.get(tick)?.[nodeId];

  if (!now) {
    return [];
  }

  const known = new Set(timelineSets.get(tick - 1)?.[nodeId] ?? []);

  return now.filter((element) => !known.has(element));
}

function rememberTimelineEdges(messages: InFlightMessage[]) {
  for (const message of messages) {
    if ((message.kind !== "gossip" && message.kind !== "sync_request") || timelineEdges.has(message.id)) {
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
  if (!timelineRecording) {
    timelineRecording = true;
    timelineSums.clear();
  timelineSets.clear();
    timelineDisabled.clear();
    timelineBaseTick = before.now;
    timelineSums.set(before.now, sumsOf(before.state));
    timelineSets.set(before.now, setsOf(before.state));
    timelineDisabled.set(before.now, before.disabled ?? []);
  }

  const tick = after.now;
  timelineSums.set(tick, sumsOf(after.state));
  timelineSets.set(tick, setsOf(after.state));
  timelineDisabled.set(tick, after.disabled ?? []);
  timelineLastTick = tick;

  const down = new Set(before.disabled ?? []);

  for (const step of batch) {
    const event = toTimelineEvent(tick, step);

    if (event.node && down.has(event.node)) {
      continue;
    }

    timelineEvents.push(event);
  }

  renderTimeGraph();
}

function toTimelineEvent(tick: number, step: Step): TimelineEvent {
  const event: TimelineEvent = { tick, type: step.type, clientId: step.clientId, seq: step.seq, replay: step.replay };

  switch (step.type) {
    case "send_gossip":
      event.node = step.from;
      break;
    case "send_gossip_ack":
      event.node = step.from;
      event.from = step.to;
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

    const sameSeq = seq === undefined ? bucket.seqs.length === 0 : bucket.seqs.includes(seq);

    if (bucket.node !== node || bucket.kind !== kind || !sameSeq) {
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
    row: 0,
    replay: event.replay ?? false,
    from: event.from,
  };

  buckets.push(bucket);
  return bucket;
}

function findReplayMerge(buckets: Bucket[], node: string, from: string | undefined, tick: number, processed: boolean) {
  for (let i = buckets.length - 1; i >= 0; i--) {
    const bucket = buckets[i];

    if (bucket.node !== node || bucket.kind !== "merge" || !bucket.replay || bucket.from !== from) {
      continue;
    }

    if (processed ? bucket.processedAt !== undefined : bucket.start === tick) {
      return bucket;
    }
  }

  return undefined;
}

function addToBucket(bucket: Bucket, event: TimelineEvent) {
  if (event.seq && !bucket.seqs.includes(event.seq)) {
    bucket.seqs.push(event.seq);
  }

  if (event.clientId && !bucket.clients.includes(event.clientId)) {
    bucket.clients.push(event.clientId);
  }
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

      case "increment":
      case "add": {
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
        if (event.replay) {
          const sameTick = buckets.find((item) => item.node === event.node && item.kind === "send" && item.replay && item.start === event.tick);
          const replayBucket = sameTick ?? newBucket(buckets, event, "send");

          addToBucket(replayBucket, event);
          replayBucket.rightLeg = true;
          break;
        }

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

      case "receive_gossip": {
        const replayBucket = event.replay ? findReplayMerge(buckets, event.node, event.from, event.tick, false) : undefined;

        if (replayBucket) {
          addToBucket(replayBucket, event);
        } else {
          newBucket(buckets, event, "merge");
        }
        break;
      }

      case "send_gossip_ack": {
        const bucket = event.replay
          ? findReplayMerge(buckets, event.node, event.from, event.tick, true)
          : findBucket(buckets, event.node, "merge", seq);

        if (bucket) {
          bucket.end = Math.max(bucket.end, event.tick);
          bucket.rightLeg = true;
        }
        break;
      }

      case "merge": {
        const bucket = event.replay
          ? (findReplayMerge(buckets, event.node, event.from, event.tick - 1, false) ??
            findReplayMerge(buckets, event.node, event.from, event.tick, true) ??
            newBucket(buckets, event, "merge"))
          : (findBucket(buckets, event.node, "merge", seq, true) ?? newBucket(buckets, event, "merge"));
        bucket.processedAt = event.tick;
        bucket.end = Math.max(bucket.end, event.tick);
        break;
      }
    }
  }

  return buckets;
}

function bucketExtent(bucket: Bucket, x: (tick: number) => number) {
  const left = x(bucket.start);
  const right = Math.max(x(bucket.end), left + 12);

  return { left, right, from: left - LEG_DX, to: right + (bucket.rightLeg ? LEG_DX : 0) };
}

function assignRows(buckets: Bucket[], x: (tick: number) => number) {
  const byNode = new Map<string, Bucket[]>();

  for (const bucket of buckets) {
    byNode.set(bucket.node, [...(byNode.get(bucket.node) ?? []), bucket]);
  }

  for (const nodeBuckets of byNode.values()) {
    nodeBuckets.sort((a, b) => a.start - b.start || a.end - b.end);
    const rowEnds: number[] = [];

    for (const bucket of nodeBuckets) {
      const extent = bucketExtent(bucket, x);
      let row = rowEnds.findIndex((end) => extent.from - BUCKET_MARGIN > end);

      if (row === -1) {
        row = rowEnds.length;
      }

      rowEnds[row] = extent.to;
      bucket.row = row;
    }
  }
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
    timeGraphContainer.replaceChildren();
    return;
  }

  const base = timelineBaseTick;
  const last = timelineLastTick;
  const x = (tick: number) => LEFT_PAD + (tick - base) * TICK_WIDTH;
  const buckets = buildBuckets();
  assignRows(buckets, x);

  const nodeIds = new Set<string>();
  for (const sums of timelineSums.values()) {
    for (const nodeId of Object.keys(sums)) {
      nodeIds.add(nodeId);
    }
  }

  const lanes: Lane[] = [];
  let y = 0;

  for (const nodeId of [...nodeIds].sort()) {
    const rows = Math.max(MIN_LANE_ROWS, ...buckets.filter((bucket) => bucket.node === nodeId).map((bucket) => bucket.row + 1));
    const height = LANE_TOP + rows * (BUCKET_HEIGHT + ROW_GAP) + LINE_GAP - ROW_GAP + LANE_BOTTOM;

    lanes.push({ id: nodeId, label: nodeId, axisY: y + height - LANE_BOTTOM, height });
    y += height;
  }

  const totalHeight = y + 4;
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

  for (const lane of lanes) {
    let downFrom: number | null = null;

    for (let tick = base; tick <= last + 1; tick++) {
      const isDown = tick <= last && (timelineDisabled.get(tick) ?? []).includes(lane.id);

      if (isDown && downFrom === null) {
        downFrom = tick;
      }

      if (!isDown && downFrom !== null) {
        const end = tick <= last ? x(tick) : x(last) + TICK_WIDTH * 0.6;
        parts.push(`<rect class="tg-down" x="${x(downFrom) - 10}" y="${lane.axisY - (lane.height - LANE_BOTTOM)}" width="${end - x(downFrom) + 10}" height="${lane.height}"/>`);
        downFrom = null;
      }
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

  for (const bucket of buckets) {
    const lane = laneById.get(bucket.node);

    if (lane) {
      parts.push(renderBucket(bucket, lane, x));
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

function renderBucket(bucket: Bucket, lane: Lane, x: (tick: number) => number): string {
  const flatY = lane.axisY - LINE_GAP - bucket.row * (BUCKET_HEIGHT + ROW_GAP);
  const tipY = flatY - BUCKET_HEIGHT;
  const { left, right } = bucketExtent(bucket, x);
  const color = bucketColor(bucket);

  const outline =
    `M ${left - LEG_DX} ${tipY} L ${left} ${flatY} L ${right} ${flatY}` +
    (bucket.rightLeg ? ` L ${right + LEG_DX} ${tipY}` : "");

  const isSet = currentCrdt === "gset";
  const sum = !isSet && bucket.processedAt !== undefined ? sumOfNode(bucket.node, bucket.processedAt) : undefined;
  const gained = isSet && bucket.processedAt !== undefined ? gainedElements(bucket.node, bucket.processedAt) : [];
  const gainedLabel = gained.length > 0 ? gained.join(", ") : undefined;
  const seqText = bucket.seqs.length > 0 ? ` #${bucket.seqs.join("/")}` : "";
  const what =
    bucket.kind === "write" ? "Write" : bucket.kind === "read" ? "Read" : bucket.kind === "merge" ? "Merge" : "Gossip";
  const replayText = bucket.replay ? "Replay: " : "";
  const detail = `${replayText}${what}${seqText}, arrives t${bucket.start}${bucket.processedAt !== undefined ? `, processed t${bucket.processedAt}` : ""}${bucket.rightLeg ? `, sends t${bucket.end}` : ""}${sum !== undefined ? `, sum ${sum}` : ""}${gained.length > 0 ? `, adds ${gained.join(", ")}` : ""}`;

  return [
    `<g class="tg-bucket">`,
    `<path d="${outline}" fill="none" stroke="${color}" stroke-width="3" stroke-linejoin="round"><title>${escapeXml(detail)}</title></path>`,
    bucket.processedAt !== undefined ? `<circle cx="${x(bucket.processedAt)}" cy="${flatY}" r="4" fill="${color}"/>` : "",
    sum !== undefined || gainedLabel !== undefined
      ? `<text class="tg-bucket-sum" x="${(left + right) / 2}" y="${flatY - 9}" text-anchor="middle" fill="${color}">${escapeXml(String(sum ?? gainedLabel))}</text>`
      : "",
    "</g>",
  ].join("");
}

function sumOfNode(nodeId: string, tick: number): number | undefined {
  return timelineSums.get(tick)?.[nodeId];
}

timeGraphToggle.addEventListener("click", () => {
  const collapsed = timeGraphWidget.classList.toggle("time-graph-widget--collapsed");
  timeGraphToggle.setAttribute("aria-expanded", String(!collapsed));
  timeGraphToggle.textContent = collapsed ? "▸" : "▾";
});

function setTickWidthIndex(index: number) {
  tickWidthIndex = Math.min(TICK_WIDTH_LEVELS.length - 1, Math.max(0, index));
  TICK_WIDTH = TICK_WIDTH_LEVELS[tickWidthIndex];
  timeGraphNarrow.disabled = tickWidthIndex === 0;
  timeGraphWiden.disabled = tickWidthIndex === TICK_WIDTH_LEVELS.length - 1;
  renderTimeGraph();
}

timeGraphNarrow.addEventListener("click", () => setTickWidthIndex(tickWidthIndex - 1));
timeGraphWiden.addEventListener("click", () => setTickWidthIndex(tickWidthIndex + 1));
setTickWidthIndex(tickWidthIndex);
