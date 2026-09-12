window.addEventListener("load", renderState);

const refreshButton = document.querySelector(
  ".btn-refresh",
) as HTMLButtonElement;

const resetButton = document.getElementById(
  "reset-button",
) as HTMLButtonElement;

const loadButton = document.getElementById(
  "load-button",
) as HTMLButtonElement;

const loadFileInput = document.getElementById(
  "load-file-input",
) as HTMLInputElement;

async function reset() {
  if (!await resetAPI()) {
    return;
  }

  resetNetworkView();
  resetTimeGraph();
  await renderState();
}

refreshButton.addEventListener("click", renderState);
resetButton.addEventListener("click", reset);
loadButton.addEventListener('click', () => loadFileInput.click());
loadFileInput.addEventListener('change', async (e) => {
  const file = loadFileInput.files?.[0];
  if (!file) return;

  const fileContent = await file.text();

  await loadStateAPI(fileContent);

  resetNetworkView();
  resetTimeGraph();
  await renderState();
});