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
  document.title = `${currentCrdt === "gset" ? "G-Set" : "G-Counter"} Simulator`;
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
