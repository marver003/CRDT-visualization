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
