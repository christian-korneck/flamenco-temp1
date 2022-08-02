
/**
 * The duration in milliseconds of the "flash" effect, when an element has been
 * copied.
 *
 * Also check `base.css`, `.copied` rule, which defines transition durations.
 */
const flashAfterCopyDuration = 150;

/**
 * Copy the inner text of an element to the clipboard.
 *
 * @param {Event } clickEvent the click event that triggered this function call.
 */
export function copyElementText(clickEvent) {
  const sourceElement = clickEvent.target;
  const inputElement = document.createElement("input");
  document.body.appendChild(inputElement);
  inputElement.setAttribute("value", sourceElement.innerText);
  inputElement.select();

  // Note that the `navigator.clipboard` interface is only available when using
  // a secure (HTTPS) connection, which Flamenco Manager will likely not have.
  // This is why this code falls back to the deprecated `document.execCommand()`
  // call.
  // Source: https://developer.mozilla.org/en-US/docs/Web/API/Clipboard
  document.execCommand("copy");

  document.body.removeChild(inputElement);

  flashElement(sourceElement);
}

function flashElement(element) {
  element.classList.add("copied");
  window.setTimeout(() => {
    element.classList.remove("copied");
  }, 150);
}
