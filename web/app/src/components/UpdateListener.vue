<template>
  <span></span>
</template>

<script>
import io from "socket.io-client";
import { ws } from '@/urls'
import * as API from "@/manager-api"
import { useSocketStatus } from '@/stores/socket-status';

const websocketURL = ws();

export default {
  emits: [
    // Data from Flamenco Manager:
    "jobUpdate", "taskUpdate", "taskLogUpdate", "message", "workerUpdate", "lastRenderedUpdate",
    // SocketIO events:
    "sioReconnected", "sioDisconnected"
  ],
  props: [
    "mainSubscription",  // One of the 'allXXX' subscription types, see `SocketIOSubscriptionType` in `flamenco-openapi.yaml`.
    "subscribedJobID",
    "subscribedTaskID",
  ],
  data() {
    return {
      socket: null,
      sockStatus: useSocketStatus(),
    }
  },
  mounted: function () {
    if (!websocketURL) {
      console.warn("UpdateListener: no websocketURL given, cannot do anything");
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
    subscribedJobID(newJobID, oldJobID) {
      if (oldJobID) {
        this._updateJobSubscription("unsubscribe", oldJobID);
      }
      if (newJobID) {
        this._updateJobSubscription("subscribe", newJobID);
      }
    },
    subscribedTaskID(newTaskID, oldTaskID) {
      if (oldTaskID) {
        this._updateTaskLogSubscription("unsubscribe", oldTaskID);
      }
      if (newTaskID) {
        this._updateTaskLogSubscription("subscribe", newTaskID);
      }
    },
    mainSubscription(newType, oldType) {
      if (oldType) {
        this._updateMainSubscription("unsubscribe", oldType);
      }
      if (newType) {
        this._updateMainSubscription("subscribe", newType);
      }
    },
  },
  methods: {
    connectToWebsocket() {
      // The SocketIO client API docs are available at:
      // https://github.com/socketio/socket.io-client/blob/2.4.x/docs/API.md
      // console.log("connecting JobsListener to WS", websocketURL);
      const ws = io(websocketURL, {
        transports: ["websocket"],
      });
      this.socket = ws;

      // For easy debugging. This assigns `ws` and not `this.socket`, as the
      // latter is Vue-reactive, which gets in the way of using in the browser
      // console.
      window.ws = ws;

      this.socket.on('connect', (error) => {
        // console.log("socketIO connection established");
        this.sockStatus.connected();
        this._resubscribe();
      });
      this.socket.on('connect_error', (error) => {
        // Don't log the error here, it's too long and noisy for regular logs.
        console.log("socketIO connection error");
        this.sockStatus.disconnected(error);
      });
      this.socket.on('error', (error) => {
        console.log("socketIO error:", error);
        this.sockStatus.disconnected(error);
      });
      this.socket.on('connect_timeout', (timeout) => {
        console.log("socketIO connection timeout:", timeout);
        this.sockStatus.disconnected("Connection timeout");
      });

      this.socket.on("disconnect", (reason) => {
        // console.log("socketIO disconnected:", reason);
        this.$emit("sioDisconnected", reason);
        this.sockStatus.disconnected(reason);

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
        this.sockStatus.connected();
        this._resubscribe();

        this.$emit("sioReconnected", attemptNumber);
      });

      this.socket.on("/jobs", (jobUpdate) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiJobUpdate = API.SocketIOJobUpdate.constructFromObject(jobUpdate)
        this.$emit("jobUpdate", apiJobUpdate);
      });

      this.socket.on("/last-rendered", (update) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiUpdate = API.SocketIOLastRenderedUpdate.constructFromObject(update)
        this.$emit("lastRenderedUpdate", apiUpdate);
      });

      this.socket.on("/task", (taskUpdate) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiTaskUpdate = API.SocketIOTaskUpdate.constructFromObject(taskUpdate)
        this.$emit("taskUpdate", apiTaskUpdate);
      });

      this.socket.on("/tasklog", (taskLogUpdate) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiTaskLogUpdate = API.SocketIOTaskLogUpdate.constructFromObject(taskLogUpdate)
        this.$emit("taskLogUpdate", apiTaskLogUpdate);
      });

      this.socket.on("/workers", (workerUpdate) => {
        // Convert to API object, in order to have the same parsing of data as
        // when we'd do an API call.
        const apiWorkerUpdate = API.SocketIOWorkerUpdate.constructFromObject(workerUpdate)
        this.$emit("workerUpdate", apiWorkerUpdate);
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

      console.log("disconnecting JobsListener WS", websocketURL);
      this.socket.disconnect();
      this.socket = null;
    },

    sendBroadcastMessage(name, message) {
      const payload = { name: name, text: message };
      this.socket.emit("/chat", payload);
    },

    /**
     * Send main subscription (un)subscription request.
     * @param {string} operation either "subscribe" or "unsubscribe"
     * @param {string} type see `SocketIOSubscriptionType` in `flamenco-openapi.yaml`.
     */
    _updateMainSubscription(operation, type) {
      const payload = new API.SocketIOSubscription(operation, type);
      this.socket.emit("/subscription", payload);
    },

    /**
     * Send job (un)subscription request.
     * @param {string} operation either "subscribe" or "unsubscribe"
     * @param {string} jobID
     */
    _updateJobSubscription(operation, jobID) {
      const payload = new API.SocketIOSubscription(operation, "job");
      payload.uuid = jobID;
      this.socket.emit("/subscription", payload);
    },

    /**
     * Send task log (un)subscription request.
     * @param {string} operation either "subscribe" or "unsubscribe"
     * @param {string} jobID
     */
    _updateTaskLogSubscription(operation, taskID) {
      const payload = new API.SocketIOSubscription(operation, "tasklog");
      payload.uuid = taskID;
      this.socket.emit("/subscription", payload);
    },

    // Resubscribe to whatever we want to be subscribed to:
    _resubscribe() {
      if (this.subscribedJobID) this._updateJobSubscription("subscribe", this.subscribedJobID);
      if (this.subscribedTaskID) this._updateTaskLogSubscription("subscribe", this.subscribedTaskID);
      if (this.mainSubscription) this._updateMainSubscription("subscribe", this.mainSubscription);
    },
  },
};
</script>
