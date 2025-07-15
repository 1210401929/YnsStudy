<!-- Chat.vue -->
<template>
  <div class="chat-window">
    <div class="chat-header">
      {{props.title}}
      <span class="close-btn" @click="closeChat">×</span>
    </div>
    <div class="chat-messages" ref="chatMessagesRef">
      <div
          v-for="(msg, index) in chatMessages"
          :key="index"
          :class="['chat-message', msg.user === '我' ? 'self' : 'other']"
      >
        <div class="message-bubble">
          <strong v-if="msg.user !== '我'" class="sender">{{ msg.user }}：</strong>
          {{ msg.text }}
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
import { ref, nextTick } from 'vue';
const emits = defineEmits(['closeChat'])
const props = defineProps({
  title:String
})
const chatText = ref('');
const chatMessages = ref([
  { user: '小明', text: '有人在吗？' },
  { user: '我', text: '我在！' },
]);

const chatMessagesRef = ref(null);

const closeChat = () => {
  emits('closeChat');
};

const sendChat = () => {
  if (!chatText.value.trim()) return;
  chatMessages.value.push({ user: '我', text: chatText.value });
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
  width: 340px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 25px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 1000;
}

.chat-header {
  background: #409EFF;
  color: white;
  padding: 12px 16px;
  font-size: 16px;
  font-weight: bold;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.close-btn {
  cursor: pointer;
  font-size: 18px;
}

.chat-messages {
  padding: 10px;
  max-height: 240px;
  overflow-y: auto;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: #f8f8f8;
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

.message-bubble {
  max-width: 70%;
  padding: 8px 12px;
  border-radius: 14px;
  background-color: #f0f0f0;
  word-break: break-word;
  font-size: 14px;
}

.chat-message.self .message-bubble {
  background-color: #409eff;
  color: white;
  border-bottom-right-radius: 0;
}

.chat-message.other .message-bubble {
  background-color: #e4e6eb;
  color: #333;
  border-bottom-left-radius: 0;
}

.chat-input-row {
  display: flex;
  padding: 10px;
  border-top: 1px solid #eee;
  background-color: #fff;
  gap: 6px;
}

.chat-input {
  flex: 1;
}

.chat-send-btn {
  white-space: nowrap;
}
</style>
