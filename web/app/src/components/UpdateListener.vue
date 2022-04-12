<template>
  <span></span>
</template>

<script>
import io from "socket.io-client";
import * as API from "../manager-api"

export default {
  emits: [
    // Data from Flamenco Manager:
    "jobUpdate", "taskUpdate", "message",
    // SocketIO events:
    "sioReconnected", "sioDisconnected"
  ],
  props: ["websocketURL"],
  data() {
    return {
      socket: null,
    }
  },
  mounted: function () {
    if (!this.websocketURL) {
      console.log("UpdateListener: no websocketURL given, cannot do anything");
      return;
    }
    this.connectToWebsocket();
  },
  beforeDestroy: function () {
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
      this.socket.on('connect_error', (error) => {
        console.log("socketIO connection error:", error);
      });
      this.socket.on('error', (error) => {
        console.log("socketIO error:", error);
      });
      this.socket.on('connect_timeout', (timeout) => {
        console.log("socketIO connection timeout:", timeout);
      });

      this.socket.on("disconnect", (reason) => {
        console.log("socketIO disconnected:", reason);
        this.$emit("sioDisconnected", reason);
      });
      this.socket.on("reconnect", (attemptNumber) => {
        console.log("socketIO reconnected after", attemptNumber, "attempts");
        this.$emit("sioReconnected", attemptNumber);
      });

      this.socket.on("/jobs", (jobUpdate) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiJobUpdate = API.JobUpdate.constructFromObject(jobUpdate)
        this.$emit("jobUpdate", apiJobUpdate);
      });

      // Chat system, useful for debugging.
      this.socket.on("/message", (message) => {
        this.$emit("message", message);
      });

      // Only connect after the handlers have been set up.
      // this.socket.open();
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
