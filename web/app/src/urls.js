
let url = new URL(window.location.href);
// Uncomment this when the web interface is running on a different port than the
// API, for example when using the Vite devserver. Set the API port here.
if (url.port == "8081") {
  url.port = "8080";
}
url.pathname = "/";
const flamencoAPIURL = url.href;

url.protocol = "ws:";
const websocketURL = url.href;

const URLs = {
  api: flamencoAPIURL,
  ws: websocketURL,
};

// console.log("Flamenco API:", URLs.api);
// console.log("Websocket   :", URLs.ws);

export function ws() {
  return URLs.ws;
}
export function api() {
  return URLs.api;
}

// Backend URLs (like task logs, SwaggerUI, etc.) should be relative to the API
// url in order to stay working when the web development server is in use.
export function backendURL(path) {
  const url = new URL(path, URLs.api);
  return url.href;
}
