import { DateTime } from "luxon";

// Do a full refresh once per hour. This is just to make sure that long-lived
// displays (like the TV in the hallway at Blender HQ) pick up on HTML/JS/CSS
// changes eventually.
const reloadAfter = {minute: 60};

function getReloadDeadline() {
  return DateTime.now().plus(reloadAfter);
}

let reloadAt = getReloadDeadline();

// Every activity (mouse move, keyboard, etc.) defers the reload.
function deferReload() {
  reloadAt = getReloadDeadline();
}

function maybeReload() {
  const now = DateTime.now();
  if (now < reloadAt) return;

  window.location.reload();
}

export default function autoreload() {
  // Check whether reloading is needed every minute.
  window.setInterval(maybeReload, 60 * 1000);

  window.addEventListener("resize", deferReload);
  window.addEventListener("mousedown", deferReload);
  window.addEventListener("mouseup", deferReload);
  window.addEventListener("mousemove", deferReload);
  window.addEventListener("keydown", deferReload);
  window.addEventListener("keyup", deferReload);
}
