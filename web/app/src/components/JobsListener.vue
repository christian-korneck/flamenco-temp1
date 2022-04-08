<template>
  <span />
</template>

<script>
import io from "socket.io-client";

export default {
  emits: ["jobUpdate", "taskUpdate", "message", "reconnected"],
  props: ["websocketURL"],
  data() {
    return {
      socket: null,
    }
  },
  mounted: function () {
    this.connectToWebsocket();
  },
  beforeDestroy: function() {
    this.disconnectWebsocket();
  },
  methods: {
    connectToWebsocket() {
      // The SocketIO client API docs are available at:
      // https://github.com/socketio/socket.io-client/blob/2.4.x/docs/API.md
      console.log("connecting JobsListener to WS", this.websocketURL);
      this.socket = io(this.websocketURL, {
        transports: ["websocket"],
      });

      this.socket.on("reconnect", (attemptNumber) => {
        console.log("socketIO reconnected after", attemptNumber, "attempts");
        this.$emit("reconnected");
      })

      this.socket.on("/jobs", (jobUpdate) => {
        this.$emit("jobUpdate", jobUpdate);
      });

      // Chat system, useful for debugging.
      this.socket.on("/message", (message) => {
        this.$emit("message", message);
      });
    },

    disconnectWebsocket() {
      if (this.socket == null) {
        console.log("no JobListener socket to disconnect");
        return;
      }

      console.log("disconnecting JobsListener WS", this.websocketURL);
      this.socket.disconnect();
      this.socket = null;
    },

    sendBroadcastMessage(name, message) {
      const payload = { name: name, text: message };
      console.log("sending broadcast message:", payload);
      this.socket.emit("/chat", payload);
    },
  },
};
</script>
