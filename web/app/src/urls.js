let url = new URL(window.location);
url.port = "8080";
const flamencoAPIURL = url.href;

url.protocol = "ws:";
const websocketURL = url.href;

const URLs = {
  api: flamencoAPIURL,
  ws: websocketURL,
};

console.log("Flamenco API:", URLs.api);
console.log("Websocket   :", URLs.ws);

export default URLs;
