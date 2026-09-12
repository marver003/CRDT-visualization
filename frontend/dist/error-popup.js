"use strict";
const errorPopup = document.getElementById("critical-error-popup");
const errorPopupTitle = document.getElementById("critical-error-title");
const errorPopupMessage = document.getElementById("critical-error-message");
const errorPopupCloseButton = document.getElementById("critical-error-close");
function showCriticalError(title, message) {
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
