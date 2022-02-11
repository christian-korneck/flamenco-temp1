<template>
  <div id="app">
    <chat-navbar></chat-navbar>
    <br />
    <chat-chatbox
      @sendMessage="sendMessage"
      :chatHistory="messages"
    ></chat-chatbox>
  </div>
</template>

<script>
import io from "socket.io-client";
import ChatNavbar from "./components/ChatNavbar.vue";
import ChatChatbox from "./components/ChatChatbox.vue";

export default {
  name: "App",
  components: {
    ChatNavbar,
    ChatChatbox,
  },
  data: () => {
    return {
      socket: null,
      serverUrl: process.env.VUE_APP_SOCKET_URL || "ws://localhost:8081",
      messages: [],
    };
  },
  mounted: function () {
    this.connectToWebsocket();
  },
  methods: {
    connectToWebsocket() {
      console.log("connecting to WS", this.serverUrl);
      this.socket = io(this.serverUrl, {
        transports: ["websocket"],
      });
      this.socket.on("/message", (message) => {
        console.log("message received: ", message);
        this.messages.push(message);
      });
    },
    sendMessage(message) {
      const payload = { name: "Nikita", text: message };
      console.log("sending:", payload);
      this.socket.emit("/chat", payload);
    },
  },
};
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

html,
body,
#app,
.card {
  height: 100%;
}
</style>
