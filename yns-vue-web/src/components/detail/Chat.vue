<!-- 聊天窗口Chat.vue -->
<template>
  <div class="chat-window">
    <div class="chat-header">
      {{ props.title }}
      <span class="close-btn" @click="closeChat">×</span>
    </div>
    <div class="chat-messages" ref="chatMessagesRef">
      <div
          v-for="(msg, index) in chatMessages"
          :key="index"
          :class="['chat-message', msg.USERCODE === userStore.userBean.code ? 'self' : 'other']"
      >
        <div class="message-bubble">
          <strong v-if="msg.USERCODE != userStore.userBean.code" class="sender">{{ msg.USERNAME }}：</strong>
          {{ msg.TEXT }}
        </div>
      </div>
    </div>
    <div class="chat-input-row">
      <el-input
          v-model="chatText"
          size="small"
          placeholder="输入消息，回车或点击发送"
          @keyup.enter="sendChat"
          class="chat-input"
      />
      <el-button type="primary" size="small" @click="sendChat" class="chat-send-btn">发送</el-button>
    </div>
  </div>
</template>

<script setup>
import {ref, nextTick} from 'vue';
import {useUserStore} from "@/stores/main/user.js";

const userStore = useUserStore();

const emits = defineEmits(['closeChat'])
const props = defineProps({
  title: String
})
const chatText = ref('');
const chatMessages = ref([
]);

const chatMessagesRef = ref(null);

const closeChat = () => {
  emits('closeChat');
};

const sendChat = () => {
  if (!chatText.value.trim()) return;
  chatMessages.value.push({
    USERCODE: userStore.userBean.code,
    USERNAME: userStore.userBean.name,
    TEXT: chatText.value
  });
  chatText.value = '';
  nextTick(() => {
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight;
    }
  });
};

</script>
<style scoped>
.chat-window {
  position: fixed;
  right: 20px;
  bottom: 90px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  width: 340px;
  overflow: hidden;
  border: 1px solid var(--j-rule);
  border-radius: 2px;
  background: var(--j-paper);
  box-shadow: 0 24px 40px -20px rgba(40, 32, 20, 0.55);
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px dashed var(--j-rule-strong);
  background: var(--j-paper-warm);
  font-family: var(--j-hand);
  font-size: 18px;
  color: var(--j-ink);
}

.close-btn {
  font-size: 20px;
  color: var(--j-muted);
  cursor: pointer;
}

.close-btn:hover {
  color: var(--j-stamp);
}

.chat-messages {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 240px;
  min-height: 120px;
  padding: 14px 12px;
  overflow-y: auto;
  font-size: 14px;
  background-color: var(--j-desk);
  background-image: radial-gradient(rgba(120, 104, 80, 0.16) 1px, transparent 1px);
  background-size: 18px 18px;
}

.chat-message {
  display: flex;
}

.chat-message.self {
  justify-content: flex-end;
}

.chat-message.other {
  justify-content: flex-start;
}

/* 每条消息是一张小纸条 */
.message-bubble {
  max-width: 72%;
  padding: 8px 12px;
  border-radius: 2px;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
  box-shadow: 0 1px 2px rgba(60, 50, 30, 0.15);
}

.chat-message.self .message-bubble {
  background-color: var(--j-note);
  color: var(--j-ink);
  transform: rotate(0.8deg);
}

.chat-message.other .message-bubble {
  background-color: #fff;
  color: var(--j-ink);
  transform: rotate(-0.8deg);
}

.sender {
  color: var(--j-pen);
}

.chat-input-row {
  display: flex;
  gap: 6px;
  padding: 10px;
  border-top: 1px dashed var(--j-rule-strong);
  background-color: var(--j-paper);
}

.chat-input {
  flex: 1;
}

.chat-send-btn {
  white-space: nowrap;
}
</style>
