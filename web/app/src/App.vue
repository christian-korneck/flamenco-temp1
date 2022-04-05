<template>
  <div id="app">
    <chat-navbar></chat-navbar>
    <br />
    <chat-chatbox
      @sendMessage="sendMessage"
      :chatHistory="messages"
    />
    <jobs-listener
      ref="jobsListener"
      :websocketURL="websocketURL"
      @jobUpdate="onJobUpdate"
      @message="onChatMessage"
    />
  </div>
</template>

<script>
import ChatNavbar from "./components/ChatNavbar.vue";
import ChatChatbox from "./components/ChatChatbox.vue";
import JobsListener from "./components/JobsListener.vue";

export default {
  name: "FlamencoWebApp",
  components: {
    ChatNavbar,
    ChatChatbox,
    JobsListener,
  },
  data: () => {
    return {
      websocketURL: "ws://localhost:8080",
      messages: [],
    };
  },
  mounted: function () {
  },
  methods: {
    sendMessage(message) {
      this.$refs.jobsListener.sendBroadcastMessage("typer", message);
    },
    onJobUpdate(jobUpdate) {
      console.log("job update received:", jobUpdate);
      if (jobUpdate.previous_status) {
        this.messages.push(`Job update: ${jobUpdate.id} (${jobUpdate.previous_status} → ${jobUpdate.status})`);
      } else {
        this.messages.push(`New job: ${jobUpdate.id} (${jobUpdate.status})`);
      }
    },
    onChatMessage(message) {
      console.log("chat message received:", message);
      this.messages.push(`${message.text}`);
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
