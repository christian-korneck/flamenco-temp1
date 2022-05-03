<template>
  <span></span>
</template>

<script>
import io from "socket.io-client";
import * as API from "@/manager-api"

export default {
  emits: [
    // Data from Flamenco Manager:
    "jobUpdate", "taskUpdate", "message",
    // SocketIO events:
    "sioReconnected", "sioDisconnected"
  ],
  props: ["websocketURL", "subscribedJob"],
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
  unmounted() {
    this.disconnectWebsocket();
  },
  beforeDestroy() {
    this.disconnectWebsocket();
  },
  watch: {
    subscribedJob(newJob, oldJob) {
      if (oldJob) {
        this._updateJobSubscription("unsubscribe", oldJob);
      }
      if (newJob) {
        this._updateJobSubscription("subscribe", newJob);
      }
    },
  },
  methods: {
    connectToWebsocket() {
      // The SocketIO client API docs are available at:
      // https://github.com/socketio/socket.io-client/blob/2.4.x/docs/API.md
      console.log("connecting JobsListener to WS", this.websocketURL);
      const ws = io(this.websocketURL, {
        transports: ["websocket"],
      });
      this.socket = ws;

      // For easy debugging. This assigns `ws` and not `this.socket`, as the
      // latter is Vue-reactive, which gets in the way of using in the browser
      // console.
      window.ws = ws;

      this.socket.on('connect_error', (error) => {
        // Don't log the error here, it's too long and noisy for regular logs.
        console.log("socketIO connection error");
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
        if (reason === 'io server disconnect') {
          // The disconnection was initiated by the server, need to reconnect
          // manually. If the disconnect was for other reasons, the socket
          // should automatically try to reconnect.

          // Intentionally commented out function call, because this should
          // happen with some nice exponential backoff instead of hammering the
          // server:
          // socket.connect();
        }
      });
      this.socket.on("reconnect", (attemptNumber) => {
        console.log("socketIO reconnected after", attemptNumber, "attempts");

        // Resubscribe to whatever we want to be subscribed to:
        if (this.subscribedJob) this._updateJobSubscription("subscribe", newJob);

        this.$emit("sioReconnected", attemptNumber);
      });

      this.socket.on("/jobs", (jobUpdate) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiJobUpdate = API.JobUpdate.constructFromObject(jobUpdate)
        this.$emit("jobUpdate", apiJobUpdate);
      });

      this.socket.on("/task", (taskUpdate) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiTaskUpdate = API.SocketIOTaskUpdate.constructFromObject(taskUpdate)
        this.$emit("taskUpdate", apiTaskUpdate);
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

    _updateJobSubscription(operation, jobID) {
      const payload = new API.SocketIOSubscription(operation, "job", jobID);
      console.log("sending job subscription:", payload);
      this.socket.emit("/subscription", payload);
    },
  },
};
</script>
