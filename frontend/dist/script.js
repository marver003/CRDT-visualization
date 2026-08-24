"use strict";
window.addEventListener("load", renderState);
const refreshButton = document.querySelector(".btn-refresh");
const resetButton = document.getElementById("reset-button");
async function reset() {
    await resetAPI();
    await renderState();
}
refreshButton.addEventListener("click", renderState);
resetButton.addEventListener("click", reset);
