<template>
  <div class="lulu-viewport">
    <section class="lulu-scene">
      <div class="scene-background"></div>

      <div class="lulu-entity" :class="[`state-${visualState.toLowerCase()}`, `action-${actionCategory}`]">
        <div class="floating-layer">
          <div
            v-for="effect in floatingEffects"
            :key="effect.id"
            class="floating-item"
            :class="effect.type"
          >
            <span class="effect-icon">{{ effect.icon }}</span>
            <span class="effect-text">{{ effect.text }}</span>
          </div>
        </div>

        <div class="action-props" aria-hidden="true">
          <span v-if="actionCategory === 'feed'" class="food-dot dot-one"></span>
          <span v-if="actionCategory === 'feed'" class="food-dot dot-two"></span>
          <span v-if="actionCategory === 'feed'" class="food-dot dot-three"></span>

          <span v-if="actionCategory === 'play'" class="toy-ball"></span>
          <span v-if="['play', 'ambient', 'touch', 'bath', 'music'].includes(actionCategory)" class="spark spark-one"></span>
          <span v-if="['play', 'ambient', 'touch', 'bath', 'music'].includes(actionCategory)" class="spark spark-two"></span>

          <span v-if="visualState === 'SLEEPING'" class="sleep-mark mark-one">Z</span>
          <span v-if="visualState === 'SLEEPING'" class="sleep-mark mark-two">Z</span>
          <span v-if="visualState === 'SLEEPING'" class="sleep-mark mark-three">Z</span>
        </div>

        <img class="lulu-img" :class="animationClass" :src="currentLuluImage" alt="噜噜" />

        <div class="status-bubble">
          {{ statusText }}
        </div>
      </div>
    </section>

    <aside class="glass-control-panel">
      <div class="panel-header">
        <div class="level-row">
          <span class="level-badge">Lv.{{ displayLevel }}</span>
          <span class="state-chip">{{ formatState(petData.currentState) }}</span>
        </div>
        <h3 class="panel-title">{{ petData.name || '噜噜' }}</h3>

        <div class="exp-container">
          <div class="exp-bar">
            <div class="exp-fill" :style="{ width: `${expPercentage}%` }"></div>
          </div>
          <span class="exp-text">{{ petData.exp || 0 }} / {{ maxExpOfCurrentLevel }} XP</span>
        </div>
      </div>

      <div class="stats-container">
        <div class="stat-item">
          <div class="stat-header">
            <span>饱腹感</span>
            <span>{{ petData.hunger || 0 }}/100</span>
          </div>
          <div class="progress-bar flat-bar">
            <div class="progress-fill fill-hunger" :style="{ width: `${petData.hunger || 0}%` }"></div>
          </div>
        </div>

        <div class="stat-item">
          <div class="stat-header">
            <span>体力值</span>
            <span>{{ petData.energy || 0 }}/100</span>
          </div>
          <div class="progress-bar flat-bar">
            <div class="progress-fill fill-energy" :style="{ width: `${petData.energy || 0}%` }"></div>
          </div>
        </div>

        <div class="stat-item">
          <div class="stat-header">
            <span>心情值</span>
            <span>{{ petData.mood || 0 }}/100</span>
          </div>
          <div class="progress-bar flat-bar">
            <div class="progress-fill fill-mood" :style="{ width: `${petData.mood || 0}%` }"></div>
          </div>
        </div>
      </div>

      <div class="action-grid">
        <button class="flat-btn feed-btn" @click="feedLulu" :disabled="isLoading">
          喂食
        </button>
        <button class="flat-btn play-btn" @click="playLulu" :disabled="isLoading">
          玩耍
        </button>
        <button class="flat-btn generic-btn" @click="sleepLulu" :disabled="isLoading">
          {{ sleepBtnText }}
        </button>
        <button class="flat-btn touch-btn" @click="touchLulu" :disabled="isLoading">
          摸摸
        </button>
        <button class="flat-btn bath-btn" @click="bathLulu" :disabled="isLoading">
          洗澡
        </button>
        <button class="flat-btn music-btn" @click="musicLulu" :disabled="isLoading">
          听音乐
        </button>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { sendAxiosRequest } from '@/utils/common.js';

const currentUserId = 1;
const isLoading = ref(false);
const actionCategory = ref('idle');
const actionLabel = ref('');
const actionFrames = ref([]);
const actionFrameIndex = ref(0);
const activeSleepImage = ref('');
const floatingEffects = ref([]);

let actionTimer = null;
let frameTimer = null;
let ambientTimer = null;
// 【修复2】：补回心跳定时器变量
let pollerTimer = null;
let effectIdCounter = 0;

const img = (name) => `/picture/lulu/benti/${name}`;

const luluImages = {
  idle: img('lulu_fadai.png'),
  happy: img('lulu_kaixin.png'),
  feed: img('lulu_eat.png'),
  feedNoodle: img('lulu_eat_noodle.png'),
  feedCookie: img('lulu_eat_cookie.png'),
  play: img('lulu_play.png'),
  playChase: img('lulu_play_chase.png'),
  playDance: img('lulu_play_dance.png'),
  idleBook: img('lulu_idle_book.png'),
  idleStretch: img('lulu_idle_stretch.png'),
  idleBubble: img('lulu_idle_bubble.png'),
  touch: img('lulu_touch.png'),
  bath: img('lulu_bath.png'),
  music: img('lulu_music.png'),
  sleep: img('lulu_sleep.png'),
  sleepFloor: img('lulu_sleep_floor.png'),
  sleepBed: img('lulu_sleep_bed.png')
};

const petData = ref({
  name: 'Lulu',
  hunger: 0,
  energy: 0,
  mood: 0,
  level: 0,
  exp: 0,
  currentState: 'IDLE'
});

const feedActions = [
  {
    label: '正在吃小蛋糕',
    frames: [luluImages.feedCookie, luluImages.feed, luluImages.feedCookie],
    effect: ['点心时间', 'type-food']
  },
  {
    label: '正在吸溜面条',
    frames: [luluImages.feedNoodle, luluImages.feed, luluImages.feedNoodle],
    effect: ['热乎乎', 'type-food']
  },
  {
    label: '正在认真干饭',
    frames: [luluImages.feed, luluImages.feedCookie, luluImages.feed],
    effect: ['吃饱啦', 'type-food']
  }
];

const playActions = [
  {
    label: '追球中',
    frames: [luluImages.playChase, luluImages.play, luluImages.playChase],
    effect: ['跑起来', 'type-mood']
  },
  {
    label: '开心跳舞',
    frames: [luluImages.playDance],
    effect: ['心情闪亮', 'type-mood']
  },
  {
    label: '蹦蹦跳跳',
    frames: [luluImages.play, luluImages.playDance, luluImages.playChase],
    effect: ['玩疯了', 'type-mood']
  }
];

const sleepActions = [
  {
    label: '在床上睡觉',
    image: luluImages.sleepBed,
    effect: ['盖好被子', 'type-sleep']
  },
  {
    label: '在地上睡着了',
    image: luluImages.sleepFloor,
    effect: ['睡得香', 'type-sleep']
  },
  {
    label: '梦见好吃的',
    image: luluImages.sleep,
    effect: ['做梦中', 'type-sleep']
  }
];

// 【修复1】：去除所有环境动作第一帧的 idle，让换图和特效同步发生
const ambientActions = [
  {
    label: '自己看小书',
    frames: [luluImages.idleBook]
  },
  {
    label: '伸个懒腰',
    frames: [luluImages.idleStretch]
  },
  {
    label: '吹泡泡',
    frames: [luluImages.idleBubble]
  }
];

const touchActions = [
  {
    label: '被摸摸头',
    frames: [luluImages.touch],
    effect: ['舒服', 'type-mood']
  }
];

const bathActions = [
  {
    label: '洗香香',
    frames: [luluImages.bath],
    effect: ['干净啦', 'type-mood']
  }
];

const musicActions = [
  {
    label: '听音乐',
    frames: [luluImages.music],
    effect: ['摇起来', 'type-mood']
  }
];

const displayLevel = computed(() => petData.value.level || 1);

const maxExpOfCurrentLevel = computed(() => displayLevel.value * 50 + 100);

const expPercentage = computed(() => {
  if (!maxExpOfCurrentLevel.value) return 0;
  return Math.min(100, ((petData.value.exp || 0) / maxExpOfCurrentLevel.value) * 100);
});

const visualState = computed(() => {
  if (actionCategory.value === 'feed') return 'FEEDING';
  if (['play', 'ambient', 'touch', 'bath', 'music'].includes(actionCategory.value)) return 'PLAYING';
  if (petData.value.currentState === 'SLEEPING') return 'SLEEPING';
  return petData.value.currentState || 'IDLE';
});

const currentLuluImage = computed(() => {
  if (actionFrames.value.length) {
    return actionFrames.value[actionFrameIndex.value % actionFrames.value.length];
  }

  if (petData.value.currentState === 'SLEEPING') {
    return activeSleepImage.value || luluImages.sleepBed;
  }

  return luluImages.idle;
});

const animationClass = computed(() => {
  if (petData.value.currentState === 'SLEEPING') return 'anim-sleep';

  const map = {
    feed: 'anim-eat',
    play: 'anim-play',
    touch: 'anim-care',
    bath: 'anim-care',
    music: 'anim-care',
    ambient: 'anim-ambient',
    idle: 'anim-breathe'
  };
  return map[actionCategory.value] || 'anim-breathe';
});

const statusText = computed(() => {
  return actionLabel.value || formatState(visualState.value);
});

const sleepBtnText = computed(() => {
  return petData.value.currentState === 'SLEEPING' ? '唤醒' : '睡觉';
});

const formatState = (state) => {
  const map = {
    IDLE: '发呆中...',
    SLEEPING: '呼呼大睡 zZZ',
    FEEDING: '正在吃饭',
    PLAYING: '玩得正开心',
    HAPPY: '心情超好'
  };
  return map[state] || state;
};

const pickOne = (items) => items[Math.floor(Math.random() * items.length)];

const getField = (data, upperKey, lowerKey, fallback) => {
  return data?.[upperKey] ?? data?.[lowerKey] ?? fallback;
};

const triggerEffect = (icon, text, type) => {
  const id = effectIdCounter++;
  floatingEffects.value.push({ id, icon, text, type });

  setTimeout(() => {
    floatingEffects.value = floatingEffects.value.filter((effect) => effect.id !== id);
  }, 1500);
};

const clearActionTimers = () => {
  window.clearTimeout(actionTimer);
  window.clearInterval(frameTimer);
};

const finishAction = () => {
  actionCategory.value = 'idle';
  actionLabel.value = '';
  actionFrames.value = [];
  actionFrameIndex.value = 0;
  window.clearInterval(frameTimer);
};

const startFrameAction = (category, action, duration = 3600, showEffect = true, frameInterval = 620) => {
  clearActionTimers();

  actionCategory.value = category;
  actionLabel.value = action.label;
  actionFrames.value = action.frames;
  actionFrameIndex.value = 0;

  const [effectText, effectType] = action.effect || [];
  if (showEffect && effectText) {
    triggerEffect(category === 'feed' ? '🍰' : '♥', effectText, effectType);
  }

  frameTimer = window.setInterval(() => {
    actionFrameIndex.value += 1;
  }, frameInterval);

  actionTimer = window.setTimeout(finishAction, duration);
};

const startSleepStill = (action) => {
  clearActionTimers();
  finishAction();
  activeSleepImage.value = action.image;
  actionLabel.value = action.label;

  const [effectText, effectType] = action.effect || [];
  if (effectText) {
    triggerEffect('Z', effectText, effectType);
  }

  actionTimer = window.setTimeout(() => {
    actionLabel.value = '';
  }, 1800);
};

const maybeStartAmbientAction = () => {
  if (isLoading.value) return;
  if (petData.value.currentState === 'SLEEPING') return;
  if (actionCategory.value !== 'idle') return;

  startFrameAction('ambient', pickOne(ambientActions), 5400, false, 1350);
};

const updatePetData = (data) => {
  if (!data) return;

  const nextLevel = Number(getField(data, 'LEVEL', 'level', 1));
  const nextExp = Number(getField(data, 'EXP', 'exp', 0));
  const nextState = getField(data, 'CURRENT_STATE', 'currentState', 'IDLE');

  if (petData.value.level > 0 && nextLevel > petData.value.level) {
    triggerEffect('★', 'LEVEL UP!', 'type-level');
  }

  petData.value = {
    name: getField(data, 'NAME', 'name', 'Lulu'),
    hunger: Number(getField(data, 'HUNGER', 'hunger', 0)),
    energy: Number(getField(data, 'ENERGY', 'energy', 0)),
    mood: Number(getField(data, 'MOOD', 'mood', 0)),
    level: nextLevel,
    exp: nextExp,
    currentState: nextState
  };

  if (nextState === 'SLEEPING' && !activeSleepImage.value) {
    activeSleepImage.value = pickOne(sleepActions).image;
  }

  if (nextState !== 'SLEEPING') {
    activeSleepImage.value = '';
  }
};

const fetchStatus = async () => {
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/status', { userNum: currentUserId });
    updatePetData(result);
  } catch (error) {
    console.error('获取 Lulu 状态失败:', error);
  }
};

const feedLulu = async () => {
  if (isLoading.value) return;

  startFrameAction('feed', pickOne(feedActions), 4200);
  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/feed', { userNum: currentUserId });
    updatePetData(result);
    triggerEffect('XP', '经验 +20', 'type-level');
  } catch (error) {
    console.error('喂食失败:', error);
  } finally {
    isLoading.value = false;
  }
};

const playLulu = async () => {
  if (isLoading.value) return;

  if (petData.value.currentState === 'SLEEPING') {
    triggerEffect('', '噜噜 正在睡觉', 'type-sleep');
    return;
  }
  if (petData.value.hunger < 15) {
    triggerEffect('!', '噜噜太饿了，需要投喂~', 'type-sleep');
    return;
  }
  if (petData.value.energy < 15) {
    triggerEffect('!', '噜噜太累了，先睡一会儿', 'type-sleep');
    return;
  }

  startFrameAction('play', pickOne(playActions), 4200);
  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/play', { userNum: currentUserId });
    updatePetData(result);
    triggerEffect('XP', '经验 +40', 'type-level');
  } catch (error) {
    console.error('玩耍失败:', error);
  } finally {
    isLoading.value = false;
  }
};

const runPlayLikeAction = async (category, actionList, expText) => {
  if (isLoading.value) return;

  if (petData.value.currentState === 'SLEEPING') {
    triggerEffect('', '噜噜 正在睡觉', 'type-sleep');
    return;
  }
  if (petData.value.hunger < 15) {
    triggerEffect('!', '噜噜太饿了，需要投喂~', 'type-sleep');
    return;
  }
  if (petData.value.energy < 15) {
    triggerEffect('!', '体力不足，先睡一会儿', 'type-sleep');
    return;
  }

  startFrameAction(category, pickOne(actionList), 4600, true, 1150);
  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/play', { userNum: currentUserId });
    updatePetData(result);
    triggerEffect('XP', expText, 'type-level');
  } catch (error) {
    console.error(`${category} 互动失败:`, error);
  } finally {
    isLoading.value = false;
  }
};

const touchLulu = () => {
  return runPlayLikeAction('touch', touchActions, '经验 +40');
};

const bathLulu = () => {
  return runPlayLikeAction('bath', bathActions, '经验 +40');
};

const musicLulu = () => {
  return runPlayLikeAction('music', musicActions, '经验 +40');
};

const sleepLulu = async () => {
  if (isLoading.value) return;

  const isWaking = petData.value.currentState === 'SLEEPING';
  const sleepAction = isWaking ? null : pickOne(sleepActions);

  if (!isWaking) {
    startSleepStill(sleepAction);
  } else {
    clearActionTimers();
    finishAction();
  }

  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/sleep', { userNum: currentUserId });
    updatePetData(result);

    if (isWaking) {
      triggerEffect('♥', '醒啦', 'type-mood');
    }
  } catch (error) {
    console.error('切换睡眠状态失败:', error);
  } finally {
    isLoading.value = false;
  }
};

onMounted(() => {
  fetchStatus();
  // 【修复2】：补回状态心跳轮询，每10秒同步一次后端进度
  pollerTimer = window.setInterval(fetchStatus, 10000);
  ambientTimer = window.setInterval(maybeStartAmbientAction, 10000);
});

onBeforeUnmount(() => {
  clearActionTimers();
  window.clearInterval(ambientTimer);
  // 【修复2】：卸载时清理心跳定时器
  window.clearInterval(pollerTimer);
});
</script>

<style scoped>
.lulu-viewport {
  display: flex;
  width: 100%;
  min-height: 100vh;
  overflow: hidden;
  position: relative;
  background: #f3f4f7;
}

.lulu-scene {
  flex: 1;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 560px;
}

.scene-background {
  position: absolute;
  inset: 0;
  z-index: 1;
  background:
    radial-gradient(circle at 42% 42%, rgba(255, 205, 92, 0.42), transparent 34%),
    linear-gradient(135deg, #fffaf0 0%, #eef7ff 46%, #f8efff 100%);
}

.scene-background::after {
  content: '';
  position: absolute;
  inset: auto 12% 9%;
  height: 18%;
  border-radius: 50%;
  background: rgba(118, 94, 48, 0.08);
  filter: blur(18px);
}

.lulu-entity {
  position: relative;
  z-index: 2;
  width: min(48vw, 440px);
  aspect-ratio: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.lulu-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter: drop-shadow(0 26px 36px rgba(93, 63, 15, 0.2));
  transition: opacity 0.18s ease;
}

.action-props {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 5;
}

.food-dot {
  position: absolute;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #ffc95a;
  box-shadow: 0 0 14px rgba(255, 201, 90, 0.7);
  animation: crumbFly 1.1s ease-in-out infinite;
}

.dot-one {
  top: 44%;
  left: 45%;
}

.dot-two {
  top: 49%;
  left: 54%;
  animation-delay: 0.22s;
}

.dot-three {
  top: 56%;
  left: 48%;
  animation-delay: 0.44s;
}

.toy-ball {
  position: absolute;
  left: 18%;
  bottom: 18%;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #30d5c8 0 48%, #ff9f43 50% 100%);
  animation: toyBounce 0.78s ease-in-out infinite;
  box-shadow: 0 12px 18px rgba(47, 128, 237, 0.22);
}

.spark {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 3px;
  background: #ffcf5a;
  transform: rotate(45deg);
  animation: sparkle 1.25s ease-in-out infinite;
}

.spark-one {
  top: 24%;
  left: 28%;
}

.spark-two {
  top: 30%;
  right: 20%;
  animation-delay: 0.45s;
}

.sleep-mark {
  position: absolute;
  right: 18%;
  color: #6c7ae0;
  font-weight: 800;
  opacity: 0;
  animation: sleepFloat 2.4s ease-in-out infinite;
}

.mark-one {
  top: 22%;
  font-size: 24px;
}

.mark-two {
  top: 12%;
  right: 12%;
  font-size: 34px;
  animation-delay: 0.55s;
}

.mark-three {
  top: 5%;
  right: 25%;
  font-size: 18px;
  animation-delay: 1s;
}

@keyframes breathe {
  0%,
  100% {
    transform: scaleX(1) scaleY(1);
  }
  50% {
    transform: scaleX(1.025) scaleY(0.975);
  }
}

@keyframes eat {
  0%,
  100% {
    transform: translateY(0) rotate(0deg) scale(1);
  }
  35% {
    transform: translateY(5px) rotate(-1deg) scaleX(1.03) scaleY(0.97);
  }
  70% {
    transform: translateY(-3px) rotate(1deg) scaleX(0.98) scaleY(1.02);
  }
}

@keyframes play {
  0%,
  100% {
    transform: translateY(0) rotate(0deg);
  }
  42% {
    transform: translateY(-20px) rotate(-2deg);
  }
  68% {
    transform: translateY(6px) rotate(2deg);
  }
}

@keyframes ambient {
  0%,
  100% {
    transform: translateY(0) rotate(0deg) scale(1);
  }
  35% {
    transform: translateY(-10px) rotate(-1.5deg) scale(1.02);
  }
  70% {
    transform: translateY(3px) rotate(1.5deg) scale(0.99);
  }
}

@keyframes sleep {
  0%,
  100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(4px) rotate(1.2deg);
  }
}

@keyframes crumbFly {
  0%,
  100% {
    opacity: 0;
    transform: translate(0, 0) scale(0.7);
  }
  40% {
    opacity: 1;
    transform: translate(-8px, -18px) scale(1);
  }
}

@keyframes toyBounce {
  0%,
  100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-18px) rotate(18deg);
  }
}

@keyframes sparkle {
  0%,
  100% {
    opacity: 0;
    transform: scale(0.6) rotate(45deg);
  }
  45% {
    opacity: 1;
    transform: scale(1) rotate(45deg);
  }
}

@keyframes sleepFloat {
  0% {
    opacity: 0;
    transform: translateY(18px) scale(0.75);
  }
  30% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translateY(-34px) scale(1.1);
  }
}

.anim-breathe {
  animation: breathe 2.5s ease-in-out infinite;
  transform-origin: center bottom;
}

.anim-eat {
  animation: eat 0.64s ease-in-out infinite;
  transform-origin: center bottom;
}

.anim-play {
  animation: play 0.78s ease-in-out infinite;
  transform-origin: center bottom;
}

.anim-care {
  animation: ambient 1.1s ease-in-out infinite;
  transform-origin: center bottom;
}

.anim-ambient {
  animation: ambient 0.9s ease-in-out infinite;
  transform-origin: center bottom;
}

.anim-sleep {
  animation: sleep 3.5s ease-in-out infinite;
  transform-origin: center bottom;
}

.floating-layer {
  position: absolute;
  top: 8%;
  left: 50%;
  transform: translateX(-50%);
  z-index: 20;
  pointer-events: none;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.floating-item {
  position: absolute;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 18px;
  font-weight: 800;
  white-space: nowrap;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.86);
  box-shadow: 0 10px 26px rgba(31, 40, 56, 0.12);
  animation: floatUpFade 1.5s cubic-bezier(0.25, 1, 0.5, 1) forwards;
}

.type-food {
  color: #e15f41;
}

.type-mood,
.type-level {
  color: #ff5b7f;
}

.type-sleep {
  color: #6c7ae0;
}

@keyframes floatUpFade {
  0% {
    opacity: 0;
    transform: translateY(20px) scale(0.8);
  }
  20% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
  100% {
    opacity: 0;
    transform: translateY(-84px) scale(0.96);
  }
}

.status-bubble {
  position: absolute;
  top: -5%;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(255, 255, 255, 0.92);
  padding: 9px 16px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 700;
  color: #333;
  box-shadow: 0 12px 28px rgba(44, 55, 75, 0.1);
  white-space: nowrap;
  z-index: 6;
}

.status-bubble::after {
  content: '';
  position: absolute;
  bottom: -6px;
  left: 50%;
  transform: translateX(-50%);
  border-width: 6px 6px 0;
  border-style: solid;
  border-color: rgba(255, 255, 255, 0.92) transparent transparent transparent;
}

.glass-control-panel {
  width: 320px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-left: 1px solid rgba(255, 255, 255, 0.72);
  display: flex;
  flex-direction: column;
  gap: 32px;
  z-index: 10;
}

.panel-header {
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  padding-bottom: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.level-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.level-badge,
.state-chip {
  font-size: 12px;
  font-weight: 800;
  padding: 4px 9px;
  border-radius: 999px;
}

.level-badge {
  color: #ff6b6b;
  background: rgba(255, 107, 107, 0.12);
}

.state-chip {
  color: #5a6c85;
  background: rgba(84, 160, 255, 0.12);
}

.panel-title {
  margin: 0;
  font-size: 22px;
  color: #2f3440;
  font-weight: 800;
}

.exp-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 4px;
}

.exp-bar {
  width: 100%;
  height: 7px;
  background: rgba(0, 0, 0, 0.06);
  border-radius: 999px;
  overflow: hidden;
}

.exp-fill {
  height: 100%;
  background: #54a0ff;
  border-radius: 999px;
  transition: width 0.3s ease-out;
}

.exp-text {
  font-size: 11px;
  color: #7a8494;
  text-align: right;
}

.stats-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-header {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  font-weight: 700;
  color: #525b68;
}

.flat-bar {
  width: 100%;
  height: 12px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.4s ease-out;
}

.fill-hunger {
  background-color: #ff9f43;
}

.fill-energy {
  background-color: #1dd1a1;
}

.fill-mood {
  background-color: #ff6b6b;
}

.action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: auto;
}

.feed-btn {
  grid-column: span 2;
}

.flat-btn {
  min-height: 48px;
  padding: 12px 14px;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  transition: transform 0.18s ease, background 0.18s ease, opacity 0.18s ease;
  background: rgba(0, 0, 0, 0.05);
  color: #4d5562;
}

.flat-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  background: rgba(0, 0, 0, 0.08);
}

.flat-btn:disabled {
  cursor: not-allowed;
  opacity: 0.64;
}

.feed-btn {
  background: rgba(255, 159, 67, 0.16);
  color: #d95f25;
}

.feed-btn:hover:not(:disabled) {
  background: rgba(255, 159, 67, 0.26);
}

.play-btn {
  background: rgba(29, 209, 161, 0.14);
  color: #13996f;
}

.generic-btn {
  background: rgba(84, 160, 255, 0.14);
  color: #2878d9;
}

.touch-btn {
  background: rgba(255, 107, 129, 0.14);
  color: #d64565;
}

.bath-btn {
  background: rgba(72, 219, 251, 0.16);
  color: #0c84a8;
}

.music-btn {
  grid-column: span 2;
  background: rgba(95, 92, 255, 0.12);
  color: #5650d8;
}

@media (max-width: 768px) {
  .lulu-viewport {
    flex-direction: column;
    min-height: 100vh;
  }

  .lulu-scene {
    min-height: 360px;
    flex: none;
  }

  .lulu-entity {
    width: min(76vw, 330px);
  }

  .glass-control-panel {
    width: auto;
    border-left: none;
    border-top: 1px solid rgba(255, 255, 255, 0.72);
  }
}
</style>
