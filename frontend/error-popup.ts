const errorPopup = document.getElementById("critical-error-popup") as HTMLElement;
const errorPopupTitle = document.getElementById("critical-error-title") as HTMLElement;
const errorPopupMessage = document.getElementById("critical-error-message") as HTMLElement;
const errorPopupCloseButton = document.getElementById(
  "critical-error-close",
) as HTMLButtonElement;

function showCriticalError(title: string, message: string) {
  errorPopupTitle.textContent = title;
  errorPopupMessage.textContent = message;
  errorPopup.hidden = false;
  errorPopupCloseButton.focus();
}

function hideCriticalError() {
  errorPopup.hidden = true;
}

errorPopupCloseButton.addEventListener("click", hideCriticalError);
errorPopup.addEventListener("click", (event) => {
  if (event.target === errorPopup) {
    hideCriticalError();
  }
});
