window.addEventListener("load", renderState);

const resetButton = document.getElementById(
  "reset-button",
) as HTMLButtonElement;

async function reset() {
  if (isAdvancing || !await resetAPI()) {
    return;
  }

  resetNetworkView();
  resetTimeGraph();
  await renderState();
}

resetButton.addEventListener("click", reset);

const simulatorSelect = document.getElementById(
  "simulator-select",
) as HTMLSelectElement;

function syncSimulatorSelect() {
  simulatorSelect.value = currentCrdt;
  document.title = `${simulatorSelect.selectedOptions[0]?.textContent ?? "CRDT Simulator"}`;
}

simulatorSelect.addEventListener("change", async () => {
  const requested = simulatorSelect.value;

  if (isAdvancing || requested === currentCrdt) {
    simulatorSelect.value = currentCrdt;
    return;
  }

  if (!await switchSimulatorAPI(requested)) {
    simulatorSelect.value = currentCrdt;
    return;
  }

  resetNetworkView();
  resetTimeGraph();
  await renderState();
});

const loadConfigButton = document.getElementById(
  "load-config-button",
) as HTMLButtonElement;
const loadConfigInput = document.getElementById(
  "load-config-input",
) as HTMLInputElement;

loadConfigButton.addEventListener("click", () => {
  if (!isAdvancing) {
    loadConfigInput.click();
  }
});

loadConfigInput.addEventListener("change", async () => {
  const file = loadConfigInput.files?.[0];
  loadConfigInput.value = "";

  if (!file || isAdvancing) {
    return;
  }

  const error = await loadConfigAPI(await file.text());

  if (error) {
    showCriticalError(`Could not load ${file.name}`, error);
    return;
  }

  resetNetworkView();
  resetTimeGraph();
  await renderState();
});
