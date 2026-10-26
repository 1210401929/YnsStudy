<template>
  <LuLu v-if="unlocked" />
  <div v-else class="lulu-gate">
    <form class="gate-card" @submit.prevent="submit">
      <h2>噜噜的小屋</h2>
      <p>请输入暗号进入</p>
      <input
        ref="inputRef"
        v-model="password"
        type="password"
        placeholder="暗号"
        autocomplete="off"
        :class="{ 'is-error': error }"
        @input="error = false"
      />
      <small v-if="error" class="gate-error">暗号不对哦</small>
      <button type="submit">进入</button>
    </form>
  </div>
</template>

<script setup>
import { defineAsyncComponent, onMounted, ref } from 'vue';

// 只是挡一下随手访问，不追求安全，密码直接写在前端
const LULU_PASSWORD = 'ailulu';
const UNLOCK_KEY = 'luluUnlocked';

// 解锁后才加载噜噜页面，避免未输入密码就开始请求接口
const LuLu = defineAsyncComponent(() => import('./LuLu.vue'));

const readUnlocked = () => {
  try {
    return window.localStorage.getItem(UNLOCK_KEY) === '1';
  } catch {
    return false;
  }
};

const unlocked = ref(readUnlocked());
const password = ref('');
const error = ref(false);
const inputRef = ref(null);

const submit = () => {
  if (password.value.trim() !== LULU_PASSWORD) {
    error.value = true;
    return;
  }
  try {
    window.localStorage.setItem(UNLOCK_KEY, '1');
  } catch {
    // 存不了就只本次生效
  }
  unlocked.value = true;
};

onMounted(() => {
  inputRef.value?.focus();
});
</script>

<style scoped>
.lulu-gate {
  width: 100%;
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 16px;
  box-sizing: border-box;
  background: linear-gradient(135deg, #fff8ed 0%, #edf8ff 50%, #fff3f6 100%);
  color: #303744;
}

.gate-card {
  width: min(320px, 100%);
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 28px 24px;
  box-sizing: border-box;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.86);
  box-shadow: 0 12px 32px rgba(48, 55, 68, 0.12);
  text-align: center;
}

.gate-card h2 {
  margin: 0;
  font-size: 22px;
}

.gate-card p {
  margin: 0;
  color: #7a8292;
  font-size: 14px;
}

.gate-card input {
  height: 40px;
  padding: 0 12px;
  border: 1px solid #dfe3ea;
  border-radius: 10px;
  font-size: 15px;
  outline: none;
  transition: border-color 0.2s ease;
}

.gate-card input:focus {
  border-color: #f0a83a;
}

.gate-card input.is-error {
  border-color: #e5604f;
}

.gate-error {
  margin-top: -6px;
  color: #e5604f;
  text-align: left;
}

.gate-card button {
  height: 40px;
  border: none;
  border-radius: 10px;
  background: #f0a83a;
  color: #fff;
  font-size: 15px;
  cursor: pointer;
}

.gate-card button:hover {
  background: #e3962a;
}
</style>
