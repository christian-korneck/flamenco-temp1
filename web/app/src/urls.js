
let url = new URL(window.location.href);
url.port = "8080";
url.pathname = "/";
const flamencoAPIURL = url.href;

url.protocol = "ws:";
const websocketURL = url.href;

const URLs = {
  api: flamencoAPIURL,
  ws: websocketURL,
};

console.log("Flamenco API:", URLs.api);
console.log("Websocket   :", URLs.ws);

export function ws() {
  return URLs.ws;
}
export function api() {
  return URLs.api;
}
