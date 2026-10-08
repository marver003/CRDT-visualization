type StepMode = "manual" | "auto";

const stepModeButtons = Array.from(
  document.querySelectorAll<HTMLButtonElement>(".step-mode-btn"),
);
const manualControls = document.getElementById("step-manual-controls") as HTMLElement;
const autoControls = document.getElementById("step-auto-controls") as HTMLElement;
const nextStepButton = document.getElementById("next-step-button") as HTMLButtonElement;
const playPauseButton = document.getElementById("play-pause-button") as HTMLButtonElement;
const playPauseIcon = document.getElementById("play-pause-icon") as HTMLElement;
const playPauseLabel = document.getElementById("play-pause-label") as HTMLElement;
const speedSelect = document.getElementById("speed-select") as HTMLSelectElement;

let mode: StepMode = "manual";
let playing = true;
let autoLoopRunning = false;
let autoLoopRunAgain = false;

function applyModeUI() {
  for (const button of stepModeButtons) {
    const isActive = button.dataset.mode === mode;
    button.setAttribute("aria-checked", `${isActive}`);
    button.classList.toggle("step-mode-btn--active", isActive);
  }

  manualControls.hidden = mode !== "manual";
  autoControls.hidden = mode !== "auto";
}

function setMode(nextMode: StepMode) {
  mode = nextMode;
  applyModeUI();

  if (mode === "auto" && playing) {
    void runAutoPlayLoop();
  }
}

function applyPlayingUI() {
  playPauseButton.setAttribute("aria-label", playing ? "Pause" : "Play");
  playPauseIcon.textContent = playing ? "⏸" : "▶";
  playPauseLabel.textContent = playing ? "Pause" : "Play";
  playPauseButton.classList.toggle("step-play-button--playing", playing);
}

function setPlaying(nextPlaying: boolean) {
  playing = nextPlaying;
  applyPlayingUI();

  if (playing) {
    void runAutoPlayLoop();
  }
}

function notifyActionQueued() {
  if (mode === "auto" && playing) {
    void runAutoPlayLoop();
  }
}

async function runAutoPlayLoop() {
  if (autoLoopRunning) {
    autoLoopRunAgain = true;
    return;
  }

  autoLoopRunning = true;

  try {
    do {
      autoLoopRunAgain = false;

      while (mode === "auto" && playing) {
        const didStep = await advanceOneStep();

        if (!didStep || mode !== "auto" || !playing) {
          break;
        }

        await wait(stepInterval());
      }
    } while (autoLoopRunAgain && mode === "auto" && playing);
  } finally {
    autoLoopRunning = false;
  }
}

async function handleNextStepClick() {
  nextStepButton.disabled = true;
  try {
    await advanceOneStep();
  } finally {
    nextStepButton.disabled = false;
  }
}

for (const button of stepModeButtons) {
  button.addEventListener("click", () => setMode(button.dataset.mode as StepMode));
}

nextStepButton.addEventListener("click", handleNextStepClick);
playPauseButton.addEventListener("click", () => setPlaying(!playing));
speedSelect.addEventListener("change", () => {
  const value = Number(speedSelect.value) as Speed;

  if ((SPEED_OPTIONS as readonly number[]).includes(value)) {
    setSpeed(value);
  }
});

applyModeUI();
applyPlayingUI();
