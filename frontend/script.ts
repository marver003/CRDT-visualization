window.addEventListener("load", renderState);

const refreshButton = document.querySelector(
  ".btn-refresh",
) as HTMLButtonElement;

const resetButton = document.getElementById(
  "reset-button",
) as HTMLButtonElement;

async function reset() {
  await resetAPI();
  await renderState();
}

refreshButton.addEventListener("click", renderState);
resetButton.addEventListener("click", reset);
