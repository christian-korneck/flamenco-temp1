<template>
  <b-input-group class="mb-2 mr-sm-2 mb-sm-0">
    <b-form-input
      placeholder="type here.."
      v-model.trim="newMsg"
      @keyup.enter="sendMessage"
    ></b-form-input>
    <b-button
      size="md"
      variant="primary"
      class="mb-2"
      @click="sendMessage"
      :disabled="!newMsg"
    >
      Send
    </b-button>
  </b-input-group>
</template>

<script>
export default {
  emits: ["sendMessage"],
  data() {
    return {
      newMsg: "",
    };
  },
  methods: {
    sendMessage() {
      if (!this.newMsg) return;
      console.log("emitting message to send:", this.newMsg);

      this.$emit("sendMessage", this.newMsg);
      this.newMsg = "";
    },
  },
};
</script>
