const SPEED_OPTIONS = [0.5, 1, 2] as const;
type Speed = (typeof SPEED_OPTIONS)[number];

const DEFAULT_SPEED: Speed = 1;

// How long the gossip packet animation takes to travel from sender to
// receiver (steps.ts's playGossipSendAnimation).
const BASE_GOSSIP_SEND_ANIMATION_MS = 2_000;

// How long a node card shows its "receiving" highlight while a
// receive_gossip step is applied (steps.ts's playGossipReceiveAnimation).
const BASE_GOSSIP_RECEIVE_ANIMATION_MS = 1_000;

// Extra time the per-value "changed" highlight lingers after a receive, on
// top of BASE_GOSSIP_RECEIVE_ANIMATION_MS (steps.ts's highlightReceivedState).
const BASE_VALUE_CHANGE_HIGHLIGHT_MS = 850;

// Pause between one auto-play batch finishing and the next one starting.
// 1x is defined as one step every 2 seconds.
const BASE_STEP_INTERVAL_MS = 2_000;

let currentSpeed: Speed = DEFAULT_SPEED;

function setSpeed(speed: Speed) {
  currentSpeed = speed;
}

function getSpeed(): Speed {
  return currentSpeed;
}

function scaled(baseMs: number): number {
  return baseMs / currentSpeed;
}

function gossipSendAnimationDuration(): number {
  return scaled(BASE_GOSSIP_SEND_ANIMATION_MS);
}

function gossipReceiveAnimationDuration(): number {
  return scaled(BASE_GOSSIP_RECEIVE_ANIMATION_MS);
}

function valueChangeHighlightDuration(): number {
  return scaled(BASE_VALUE_CHANGE_HIGHLIGHT_MS);
}

// Only meaningful in auto-play - manual "Next Step" clicks are user-paced
// and never wait on this.
function stepInterval(): number {
  return scaled(BASE_STEP_INTERVAL_MS);
}
