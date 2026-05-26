<template>
  <div class="lulu-viewport">
    <section class="lulu-stage">
      <div class="scene-background"></div>

      <div class="top-status">
        <div class="identity-block">
          <span class="level-badge">Lv.{{ displayLevel }}</span>
          <div>
            <h2>{{ petData.name || '噜噜' }}</h2>
            <p>{{ formatState(petData.currentState) }}</p>
          </div>
        </div>

        <div class="quick-stats">
          <div class="stat-pill stat-hunger">
            <span>饱腹</span>
            <strong>{{ petData.hunger || 0 }}</strong>
            <div class="stat-bar">
              <div class="stat-fill" :style="{ width: `${petData.hunger || 0}%` }"></div>
            </div>
          </div>
          <div class="stat-pill stat-energy">
            <span>体力</span>
            <strong>{{ petData.energy || 0 }}</strong>
            <div class="stat-bar">
              <div class="stat-fill" :style="{ width: `${petData.energy || 0}%` }"></div>
            </div>
          </div>
          <div class="stat-pill stat-mood">
            <span>心情</span>
            <strong>{{ petData.mood || 0 }}</strong>
            <div class="stat-bar">
              <div class="stat-fill" :style="{ width: `${petData.mood || 0}%` }"></div>
            </div>
          </div>
        </div>
      </div>

      <div class="message-fly-zone" aria-hidden="true">
        <div
          v-for="(message, index) in floatingMessages"
          :key="`${message.id}-fly`"
          class="fly-message"
          :style="getFlyStyle(index)"
        >
          {{ message.content }}
        </div>
      </div>

      <div class="fun-hud">
        <div class="daily-card">
          <span>今日小任务</span>
          <strong>{{ dailyMission.title }}</strong>
          <div class="mission-progress">
            <div class="mission-fill" :style="{ width: `${dailyMissionProgress}%` }"></div>
          </div>
          <small>{{ dailyMission.current }}/{{ dailyMission.target }} · {{ dailyMission.reward }}</small>
        </div>
        <div class="streak-card">
          <span>连续陪伴</span>
          <strong>{{ careStreak }} 天</strong>
          <small>{{ moodWeather.text }}</small>
        </div>
      </div>

      <div
        class="lulu-entity"
        :class="[`state-${visualState.toLowerCase()}`, `action-${actionCategory}`]"
      >
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

        <div class="body-hotspots" aria-label="点击噜噜互动">
          <button class="body-hotspot hotspot-head" type="button" aria-label="摸摸噜噜脑袋" @click="tapLuluBody('head')"></button>
          <button class="body-hotspot hotspot-belly" type="button" aria-label="戳戳噜噜肚子" @click="tapLuluBody('belly')"></button>
          <button class="body-hotspot hotspot-foot" type="button" aria-label="碰碰噜噜脚" @click="tapLuluBody('foot')"></button>
        </div>

        <div class="status-bubble">
          {{ statusText }}
        </div>
      </div>

      <section class="action-dock">
        <div class="progress-strip">
          <div class="exp-row">
            <span>经验</span>
            <strong>{{ petData.exp || 0 }} / {{ maxExpOfCurrentLevel }}</strong>
          </div>
          <div class="exp-bar">
            <div class="exp-fill" :style="{ width: `${expPercentage}%` }"></div>
          </div>
          <button class="log-btn" type="button" @click="openLogPanel">噜噜 日志</button>
        </div>

        <div class="action-grid">
          <button class="action-btn feed-btn" @click="feedLulu" :disabled="isLoading">
            <span>喂食</span>
            <small>饱腹 +30</small>
          </button>
          <button class="action-btn play-btn" @click="playLulu" :disabled="isLoading">
            <span>玩耍</span>
            <small>心情 +20</small>
          </button>
          <button class="action-btn sleep-btn" @click="sleepLulu" :disabled="isLoading">
            <span>{{ sleepBtnText }}</span>
            <small>恢复体力</small>
          </button>
          <button class="action-btn touch-btn" @click="touchLulu" :disabled="isLoading">
            <span>摸摸</span>
            <small>陪伴一下</small>
          </button>
          <button class="action-btn bath-btn" @click="bathLulu" :disabled="isLoading">
            <span>洗澡</span>
            <small>清爽状态</small>
          </button>
          <button class="action-btn music-btn" @click="musicLulu" :disabled="isLoading">
            <span>听音乐</span>
            <small>放松心情</small>
          </button>
          <button class="action-btn wish-btn" @click="makeWish" :disabled="isLoading">
            <span>许愿</span>
            <small>{{ wishText }}</small>
          </button>
          <button class="action-btn dress-btn" @click="changeAccessory" :disabled="isLoading">
            <span>换装</span>
            <small>{{ currentAccessory.name }}</small>
          </button>
        </div>
      </section>
    </section>

    <aside class="message-board">
      <div class="message-board-header">
        <div>
          <p class="message-eyebrow">留言板</p>
          <h3>写给 噜噜 的小纸条</h3>
        </div>
        <button class="message-refresh" @click="fetchMessages" :disabled="isMessageLoading">刷新</button>
      </div>

      <form class="message-form" @submit.prevent="sendMessage">
        <textarea
          v-model="messageInput"
          :disabled="isMessageLoading"
          maxlength="500"
          placeholder="写一条会从 噜噜 身边飘过的留言..."
          @keydown.enter.exact.prevent="sendMessage"
        ></textarea>
        <div class="message-actions">
          <span>{{ messageInput.length }}/500</span>
          <button type="submit" :disabled="isMessageLoading || !messageInput.trim()">
            {{ isMessageLoading ? '发布中' : '发布留言' }}
          </button>
        </div>
      </form>

      <div ref="messageListRef" class="message-list">
        <div v-if="!messages.length" class="empty-message">
          还没有留言。第一张小纸条，等你贴上来。
        </div>

        <article v-for="message in messages" :key="message.id" class="message-card">
          <div class="message-card-main">
            <p>{{ message.content }}</p>
            <time>{{ formatMessageTime(message.createTime) }}</time>
          </div>
          <button class="delete-message" @click="deleteMessage(message)" :disabled="isMessageLoading">删除</button>
        </article>
      </div>
    </aside>

    <div v-if="isLogPanelOpen" class="log-overlay" @click.self="isLogPanelOpen = false">
      <section class="log-panel">
        <div class="log-panel-header">
          <div>
            <p class="message-eyebrow">互动日志</p>
            <h3>噜噜 的照顾记录</h3>
          </div>
          <button class="message-refresh" @click="isLogPanelOpen = false">关闭</button>
        </div>

        <div class="log-list">
          <div v-if="isLogLoading" class="empty-message">日志读取中...</div>
          <div v-else-if="!logs.length" class="empty-message">还没有互动日志。</div>

          <article v-for="log in logs" :key="log.id" class="log-card">
            <div class="log-icon">{{ getLogIcon(log.actionType) }}</div>
            <div class="log-content">
              <div class="log-title-row">
                <strong>{{ log.actionName }}</strong>
                <time>{{ formatMessageTime(log.createTime) }}</time>
              </div>
              <p>{{ log.remark }}</p>
              <div class="log-meta">
                <span>IP：{{ log.ipAddress || '未知' }}</span>
                <span>{{ log.browser || '未知浏览器' }}</span>
                <span>{{ log.deviceModel || '未知设备' }}</span>
              </div>
            </div>
          </article>
        </div>
      </section>
    </div>

  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { sendAxiosRequest } from '@/utils/common.js';

const currentUserId = 1;
const isLoading = ref(false);
const isMessageLoading = ref(false);
const actionCategory = ref('idle');
const actionLabel = ref('');
const actionFrames = ref([]);
const actionFrameIndex = ref(0);
const activeSleepImage = ref('');
const floatingEffects = ref([]);
const messages = ref([]);
const messageInput = ref('');
const messageListRef = ref(null);
const logs = ref([]);
const isLogLoading = ref(false);
const isLogPanelOpen = ref(false);
const careStreak = ref(1);
const dailyMission = ref({
  type: 'feed',
  title: '给噜噜准备一顿饭',
  current: 0,
  target: 1,
  reward: '完成后心情会亮一下'
});
const accessoryIndex = ref(0);
const wishText = ref('抽一句');

let actionTimer = null;
let frameTimer = null;
let ambientTimer = null;
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
  idleHoodie: img('lulu_fadai_hoodie.png'),
  idleOveralls: img('lulu_fadai_overalls.png'),
  idleScarf: img('lulu_fadai_scarf.png'),
  touch: img('lulu_touch.png'),
  bath: img('lulu_bath.png'),
  music: img('lulu_music.png'),
  tapHead: img('lulu_tap_head.png'),
  tapBelly: img('lulu_tap_belly.png'),
  tapFoot: img('lulu_tap_foot.png'),
  sleep: img('lulu_sleep.png'),
  sleepFloor: img('lulu_sleep_floor.png'),
  sleepBed: img('lulu_sleep_bed.png')
};

const petData = ref({
  name: '噜噜',
  hunger: 0,
  energy: 0,
  mood: 0,
  level: 0,
  exp: 0,
  currentState: 'IDLE'
});

const feedActions = [
  { label: '正在吃小蛋糕', frames: [luluImages.feedCookie, luluImages.feed, luluImages.feedCookie], effect: ['点心时间', 'type-food'] },
  { label: '正在吸溜面条', frames: [luluImages.feedNoodle, luluImages.feed, luluImages.feedNoodle], effect: ['热乎乎', 'type-food'] },
  { label: '正在认真干饭', frames: [luluImages.feed, luluImages.feedCookie, luluImages.feed], effect: ['吃饱啦', 'type-food'] }
];

const playActions = [
  { label: '追球中', frames: [luluImages.playChase, luluImages.play, luluImages.playChase], effect: ['跑起来', 'type-mood'] },
  { label: '开心跳舞', frames: [luluImages.playDance], effect: ['心情明亮', 'type-mood'] },
  { label: '蹦蹦跳跳', frames: [luluImages.play, luluImages.playDance, luluImages.playChase], effect: ['玩疯了', 'type-mood'] }
];

const sleepActions = [
  { label: '在床上睡觉', image: luluImages.sleepBed, effect: ['盖好被子', 'type-sleep'] },
  { label: '在地上睡着了', image: luluImages.sleepFloor, effect: ['睡得很香', 'type-sleep'] },
  { label: '梦见好吃的', image: luluImages.sleep, effect: ['做梦中', 'type-sleep'] }
];

const ambientActions = [
  { label: '自己看小书', frames: [luluImages.idleBook] },
  { label: '伸个懒腰', frames: [luluImages.idleStretch] },
  { label: '吹泡泡', frames: [luluImages.idleBubble] }
];

const touchActions = [
  { label: '被摸摸头', frames: [luluImages.touch], effect: ['舒服', 'type-mood'] }
];

const bathActions = [
  { label: '洗香香', frames: [luluImages.bath], effect: ['干净啦', 'type-mood'] }
];

const musicActions = [
  { label: '听音乐', frames: [luluImages.music], effect: ['摇起来', 'type-mood'] }
];

const accessoryModes = [
  { name: '初始衣服', image: luluImages.idle },
  { name: '蓝色卫衣', image: luluImages.idleHoodie },
  { name: '绿色背带裤', image: luluImages.idleOveralls },
  { name: '红围巾套装', image: luluImages.idleScarf }
];

const missionPool = [
  { type: 'feed', title: '给噜噜准备一顿饭', target: 1, reward: '完成后心情会亮一下' },
  { type: 'play', title: '陪噜噜玩两次', target: 2, reward: '完成后撒一把星星' },
  { type: 'touch', title: '摸摸噜噜一次', target: 1, reward: '完成后获得贴贴感' },
  { type: 'wish', title: '和噜噜许个愿', target: 1, reward: '完成后收到小签语' },
];

const wishPool = [
  '今天适合慢慢来，但不要停下来。',
  '噜噜批准你休息五分钟。',
  '今天的小幸运藏在一杯热饮后面。',
  '先把最小的一件事做完，后面会轻很多。',
  '噜噜觉得你已经很努力了。'
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

  return currentAccessory.value.image || luluImages.idle;
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

const statusText = computed(() => actionLabel.value || formatState(visualState.value));

const sleepBtnText = computed(() => {
  return petData.value.currentState === 'SLEEPING' ? '唤醒' : '睡觉';
});

const floatingMessages = computed(() => {
  return messages.value.slice(0, 8).reverse();
});

const currentAccessory = computed(() => accessoryModes[accessoryIndex.value] || accessoryModes[0]);

const dailyMissionProgress = computed(() => {
  if (!dailyMission.value.target) return 0;
  return Math.min(100, (dailyMission.value.current / dailyMission.value.target) * 100);
});

const moodWeather = computed(() => {
  const mood = petData.value.mood || 0;
  if (mood >= 80) return { text: '今天是闪闪发亮日' };
  if (mood >= 50) return { text: '噜噜状态不错' };
  if (mood >= 25) return { text: '噜噜想被多陪陪' };
  return { text: '噜噜有点低落' };
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

const applyFunState = (data) => {
  if (!data) return;
  dailyMission.value = {
    type: getField(data, 'MISSION_TYPE', 'missionType', 'feed'),
    title: getField(data, 'MISSION_TITLE', 'missionTitle', '给噜噜准备一顿饭'),
    current: Number(getField(data, 'MISSION_CURRENT', 'missionCurrent', 0)),
    target: Number(getField(data, 'MISSION_TARGET', 'missionTarget', 1)),
    reward: getField(data, 'MISSION_REWARD', 'missionReward', '完成后心情会亮一下')
  };
  careStreak.value = Number(getField(data, 'STREAK_COUNT', 'streakCount', 1));
  accessoryIndex.value = Number(getField(data, 'CLOTHES_INDEX', 'clothesIndex', 0)) % accessoryModes.length;
};

const fetchFunState = async () => {
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/fun-state', { userNum: currentUserId });
    applyFunState(result);
  } catch (error) {
    console.error('获取 噜噜 玩法状态失败:', error);
  }
};

const advanceMission = async (type, amount = 1) => {
  if (dailyMission.value.type !== type) return;

  const wasCompleted = dailyMission.value.current >= dailyMission.value.target;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/mission/advance', {
      userNum: currentUserId,
      missionType: type,
      amount
    });
    applyFunState(result);

    if (!wasCompleted && Number(getField(result, 'MISSION_CURRENT', 'missionCurrent', 0)) >= Number(getField(result, 'MISSION_TARGET', 'missionTarget', 1))) {
      triggerEffect('星', '今日小任务完成', 'type-level');
      startFrameAction('ambient', { label: '任务完成，噜噜很得意', frames: [luluImages.happy] }, 2600, false, 1000);
    }
  } catch (error) {
    console.error('推进 噜噜 今日任务失败:', error);
  }
};

const getFlyStyle = (index) => {
  const lanes = ['16%', '27%', '39%', '52%', '66%', '78%'];
  return {
    top: lanes[index % lanes.length],
    animationDelay: `${index * 1.4}s`,
    animationDuration: `${18 + (index % 3) * 4}s`
  };
};

const triggerEffect = (icon, text, type) => {
  const id = effectIdCounter++;
  floatingEffects.value.push({ id, icon, text, type });

  window.setTimeout(() => {
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
    triggerEffect(category === 'feed' ? '食' : '心', effectText, effectType);
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

  startFrameAction('ambient', pickOne(ambientActions), 3000, false, 1350);
};

const updatePetData = (data) => {
  if (!data) return;

  const nextLevel = Number(getField(data, 'LEVEL', 'level', 1));
  const nextExp = Number(getField(data, 'EXP', 'exp', 0));
  const nextState = getField(data, 'CURRENT_STATE', 'currentState', 'IDLE');

  if (petData.value.level > 0 && nextLevel > petData.value.level) {
    triggerEffect('*', 'LEVEL UP!', 'type-level');
  }

  petData.value = {
    name: getField(data, 'NAME', 'name', '噜噜'),
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

const normalizeMessage = (message, index) => {
  const createTime = getField(message, 'CREATE_TIME', 'createTime', '');
  return {
    id: getField(message, 'ID', 'id', `message-${createTime}-${index}`),
    content: getField(message, 'CONTENT', 'content', ''),
    createTime,
    ipAddress: getField(message, 'IP_ADDRESS', 'ipAddress', '')
  };
};

const normalizeMessageResult = (result) => {
  if (Array.isArray(result)) {
    return result.map(normalizeMessage);
  }

  if (result?.isError) {
    ElMessage.warning(result.errMsg || result.message || result.msg || '操作失败');
    return messages.value;
  }

  if (Array.isArray(result?.result)) {
    return result.result.map(normalizeMessage);
  }

  return [];
};

const normalizeLog = (log, index) => {
  const createTime = getField(log, 'CREATE_TIME', 'createTime', '');
  return {
    id: getField(log, 'ID', 'id', `log-${createTime}-${index}`),
    actionType: getField(log, 'ACTION_TYPE', 'actionType', ''),
    actionName: getField(log, 'ACTION_NAME', 'actionName', '互动'),
    ipAddress: getField(log, 'IP_ADDRESS', 'ipAddress', ''),
    browser: getField(log, 'BROWSER', 'browser', ''),
    deviceModel: getField(log, 'DEVICE_MODEL', 'deviceModel', ''),
    remark: getField(log, 'REMARK', 'remark', ''),
    createTime
  };
};

const fetchLogs = async () => {
  isLogLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/logs', { userNum: currentUserId });
    logs.value = Array.isArray(result) ? result.map(normalizeLog) : [];
  } catch (error) {
    console.error('获取 噜噜 日志失败:', error);
    ElMessage.error('日志读取失败');
  } finally {
    isLogLoading.value = false;
  }
};

const openLogPanel = async () => {
  isLogPanelOpen.value = true;
  await fetchLogs();
};

const getLogIcon = (actionType) => {
  const map = {
    FEED: '食',
    PLAY: '玩',
    SLEEP: '睡',
    WAKE: '醒'
  };
  return map[actionType] || '记';
};

const changeAccessory = async () => {
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/clothes/change', { userNum: currentUserId });
    applyFunState(result);
    triggerEffect('装', `换成${currentAccessory.value.name}`, 'type-mood');
    startFrameAction('ambient', { label: `噜噜换上了${currentAccessory.value.name}`, frames: [currentAccessory.value.image] }, 2200, false, 1000);
  } catch (error) {
    console.error('切换 噜噜 衣服失败:', error);
    ElMessage.error('换装失败');
  }
};

const makeWish = () => {
  const wish = pickOne(wishPool);
  wishText.value = wish.length > 6 ? `${wish.slice(0, 6)}...` : wish;
  triggerEffect('愿', wish, 'type-level');
  advanceMission('wish');
};

const tapLuluBody = (part) => {
  if (petData.value.currentState === 'SLEEPING') {
    triggerEffect('Z', '噜噜正在睡觉', 'type-sleep');
    return;
  }

  const actionMap = {
    head: {
      label: '别突然摸脑袋啦',
      frames: [luluImages.tapHead],
      effect: ['有点害羞', 'type-mood'],
      icon: '头'
    },
    belly: {
      label: '揉揉肚子中',
      frames: [luluImages.tapBelly],
      effect: ['咕噜咕噜', 'type-food'],
      icon: '肚'
    },
    foot: {
      label: '脚脚被碰到了',
      frames: [luluImages.tapFoot],
      effect: ['痒痒', 'type-mood'],
      icon: '脚'
    }
  };
  const action = actionMap[part];
  if (!action) return;

  startFrameAction('touch', action, 2600, false, 1000);
  triggerEffect(action.icon, action.effect[0], action.effect[1]);
  advanceMission('touch');
};

const scrollMessagesToTop = async () => {
  await nextTick();
  if (messageListRef.value) {
    messageListRef.value.scrollTop = 0;
  }
};

const fetchStatus = async () => {
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/status', { userNum: currentUserId });
    updatePetData(result);
  } catch (error) {
    console.error('获取 噜噜 状态失败:', error);
  }
};

const fetchMessages = async () => {
  isMessageLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/messages', { userNum: currentUserId });
    messages.value = normalizeMessageResult(result);
    await scrollMessagesToTop();
  } catch (error) {
    console.error('获取 噜噜 留言失败:', error);
  } finally {
    isMessageLoading.value = false;
  }
};

const sendMessage = async () => {
  const content = messageInput.value.trim();
  if (!content || isMessageLoading.value) return;

  isMessageLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/message/add', {
      userNum: currentUserId,
      content
    });
    if (!result?.isError) {
      messageInput.value = '';
      triggerEffect('信', '留言已飘出去', 'type-mood');
    }
    messages.value = normalizeMessageResult(result);
    await scrollMessagesToTop();
  } catch (error) {
    console.error('发送 噜噜 留言失败:', error);
    ElMessage.error('留言发送失败');
  } finally {
    isMessageLoading.value = false;
  }
};

const deleteMessage = async (message) => {
  if (isMessageLoading.value) return;

  isMessageLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/message/delete', {
      userNum: currentUserId,
      messageId: message.id
    });
    messages.value = normalizeMessageResult(result);
    ElMessage.success('留言已删除');
  } catch (error) {
    console.error('删除 噜噜 留言失败:', error);
    ElMessage.error('删除失败');
  } finally {
    isMessageLoading.value = false;
  }
};

const formatMessageTime = (time) => {
  if (!time) return '';
  const date = new Date(String(time).replace(' ', 'T'));
  if (Number.isNaN(date.getTime())) {
    return String(time).slice(0, 16);
  }
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
};

const feedLulu = async () => {
  if (isLoading.value) return;

  if (petData.value.hunger >= 90) {
    triggerEffect('!', '噜噜 已经吃饱了', 'type-sleep');
    return;
  }

  startFrameAction('feed', pickOne(feedActions), 4200);
  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/feed', { userNum: currentUserId });
    updatePetData(result);
    triggerEffect('XP', '经验 +20', 'type-level');
    advanceMission('feed');
  } catch (error) {
    console.error('喂食失败:', error);
  } finally {
    isLoading.value = false;
  }
};

const playLulu = async () => {
  if (isLoading.value) return;

  if (petData.value.currentState === 'SLEEPING') {
    triggerEffect('Z', '噜噜 正在睡觉', 'type-sleep');
    return;
  }
  if (petData.value.hunger < 15) {
    triggerEffect('!', '噜噜 太饿了，需要先喂食', 'type-sleep');
    return;
  }
  if (petData.value.energy < 15) {
    triggerEffect('!', '噜噜 太累了，先睡一会儿', 'type-sleep');
    return;
  }

  startFrameAction('play', pickOne(playActions), 4200);
  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/play', { userNum: currentUserId, actionName: '玩耍' });
    updatePetData(result);
    triggerEffect('XP', '经验 +40', 'type-level');
    advanceMission('play');
  } catch (error) {
    console.error('玩耍失败:', error);
  } finally {
    isLoading.value = false;
  }
};

const runPlayLikeAction = async (category, actionList, expText) => {
  if (isLoading.value) return;

  if (petData.value.currentState === 'SLEEPING') {
    triggerEffect('Z', '噜噜 正在睡觉', 'type-sleep');
    return;
  }
  if (petData.value.hunger < 15) {
    triggerEffect('!', '噜噜 太饿了，需要先喂食', 'type-sleep');
    return;
  }
  if (petData.value.energy < 15) {
    triggerEffect('!', '体力不足，先睡一会儿', 'type-sleep');
    return;
  }

  startFrameAction(category, pickOne(actionList), 4600, true, 1150);
  isLoading.value = true;

  try {
    const actionNameMap = {
      touch: '摸摸',
      bath: '洗澡',
      music: '听音乐'
    };
    const result = await sendAxiosRequest('/blog-api/lulu/play', {
      userNum: currentUserId,
      actionName: actionNameMap[category] || '玩耍'
    });
    updatePetData(result);
    triggerEffect('XP', expText, 'type-level');
    advanceMission(category);
  } catch (error) {
    console.error(`${category} 互动失败:`, error);
  } finally {
    isLoading.value = false;
  }
};

const touchLulu = () => runPlayLikeAction('touch', touchActions, '经验 +40');

const bathLulu = () => runPlayLikeAction('bath', bathActions, '经验 +40');

const musicLulu = () => runPlayLikeAction('music', musicActions, '经验 +40');

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
      triggerEffect('心', '醒啦', 'type-mood');
    }
  } catch (error) {
    console.error('切换睡眠状态失败:', error);
  } finally {
    isLoading.value = false;
  }
};

onMounted(() => {
  fetchFunState();
  fetchStatus();
  fetchMessages();
  pollerTimer = window.setInterval(fetchStatus, 10000);
  ambientTimer = window.setInterval(maybeStartAmbientAction, 20000);
});

onBeforeUnmount(() => {
  clearActionTimers();
  window.clearInterval(ambientTimer);
  window.clearInterval(pollerTimer);
});
</script>

<style scoped>
.lulu-viewport {
  width: 100%;
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  overflow: hidden;
  background: #f5f6fa;
  color: #303744;
}

.lulu-stage {
  position: relative;
  min-height: 100vh;
  display: grid;
  grid-template-rows: auto minmax(360px, 1fr) auto;
  align-items: center;
  overflow: hidden;
  padding: 28px clamp(20px, 4vw, 56px);
}

.scene-background {
  position: absolute;
  inset: 0;
  z-index: 0;
  background:
    radial-gradient(circle at 48% 45%, rgba(255, 205, 92, 0.38), transparent 30%),
    linear-gradient(135deg, #fff8ed 0%, #edf8ff 50%, #fff3f6 100%);
}

.scene-background::before {
  content: '';
  position: absolute;
  inset: auto 11% 13%;
  height: 17%;
  border-radius: 50%;
  background: rgba(99, 73, 31, 0.09);
  filter: blur(18px);
}

.top-status {
  position: relative;
  z-index: 4;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18px;
}

.identity-block {
  display: flex;
  align-items: center;
  gap: 14px;
}

.identity-block h2 {
  margin: 0;
  font-size: 30px;
  line-height: 1.1;
  font-weight: 900;
}

.identity-block p {
  margin: 6px 0 0;
  color: #6f7a8a;
  font-size: 14px;
  font-weight: 700;
}

.level-badge {
  min-width: 58px;
  height: 58px;
  border-radius: 18px;
  display: grid;
  place-items: center;
  color: #ff6a3d;
  background: rgba(255, 255, 255, 0.78);
  box-shadow: 0 18px 40px rgba(40, 49, 66, 0.1);
  font-size: 16px;
  font-weight: 900;
}

.quick-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(74px, 1fr));
  gap: 10px;
}

.stat-pill {
  padding: 11px 14px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.74);
  box-shadow: 0 14px 34px rgba(40, 49, 66, 0.08);
}

.stat-pill span,
.exp-row span {
  display: block;
  color: #7d8796;
  font-size: 12px;
  font-weight: 800;
}

.stat-pill strong {
  display: block;
  margin-top: 2px;
  font-size: 22px;
  line-height: 1;
}

.stat-bar {
  width: 100%;
  height: 6px;
  margin-top: 10px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(48, 55, 68, 0.09);
}

.stat-fill {
  height: 100%;
  border-radius: inherit;
  transition: width 0.35s ease-out;
}

.stat-hunger .stat-fill {
  background: #ff9f43;
}

.stat-energy .stat-fill {
  background: #1dd1a1;
}

.stat-mood .stat-fill {
  background: #ff6b81;
}

.message-fly-zone {
  position: absolute;
  inset: 92px 0 190px;
  z-index: 2;
  pointer-events: none;
  overflow: hidden;
}

.fly-message {
  position: absolute;
  right: -42%;
  max-width: 340px;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.68);
  color: #445062;
  border: 1px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 12px 28px rgba(44, 55, 75, 0.08);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;
  font-weight: 700;
  animation: messageDrift linear infinite;
}

.fun-hud {
  position: absolute;
  left: clamp(18px, 4vw, 56px);
  top: 118px;
  z-index: 4;
  width: min(280px, 28vw);
  display: flex;
  flex-direction: column;
  gap: 10px;
  pointer-events: none;
}

.daily-card,
.streak-card {
  padding: 13px 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(255, 255, 255, 0.82);
  box-shadow: 0 14px 34px rgba(40, 49, 66, 0.08);
  backdrop-filter: blur(14px);
}

.daily-card span,
.streak-card span {
  display: block;
  color: #7d8796;
  font-size: 12px;
  font-weight: 900;
}

.daily-card strong,
.streak-card strong {
  display: block;
  margin-top: 4px;
  color: #303744;
  font-size: 15px;
  line-height: 1.35;
}

.daily-card small,
.streak-card small {
  display: block;
  margin-top: 5px;
  color: #8a94a6;
  font-size: 12px;
  font-weight: 800;
}

.mission-progress {
  width: 100%;
  height: 7px;
  margin-top: 9px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(48, 55, 68, 0.08);
}

.mission-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #ff9f43, #ff6b81);
  transition: width 0.28s ease-out;
}

.lulu-entity {
  position: relative;
  z-index: 3;
  justify-self: center;
  width: min(48vw, 470px);
  aspect-ratio: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.lulu-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter: drop-shadow(0 28px 38px rgba(93, 63, 15, 0.22));
  transition: opacity 0.18s ease;
}

.body-hotspots {
  position: absolute;
  inset: 0;
  z-index: 8;
}

.body-hotspot {
  position: absolute;
  border: none;
  padding: 0;
  border-radius: 999px;
  background: transparent;
  cursor: pointer;
}

.body-hotspot:focus-visible {
  outline: 2px solid rgba(84, 160, 255, 0.7);
  outline-offset: 3px;
}

.hotspot-head {
  left: 27%;
  top: 12%;
  width: 46%;
  height: 28%;
}

.hotspot-belly {
  left: 25%;
  top: 42%;
  width: 50%;
  height: 31%;
}

.hotspot-foot {
  left: 28%;
  top: 74%;
  width: 44%;
  height: 18%;
}

.action-dock {
  position: relative;
  z-index: 5;
  display: grid;
  grid-template-columns: 210px minmax(0, 1fr);
  gap: 16px;
  align-items: stretch;
  padding: 16px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.86);
  box-shadow: 0 22px 46px rgba(40, 49, 66, 0.12);
  backdrop-filter: blur(18px);
}

.progress-strip {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 16px;
  background: rgba(84, 160, 255, 0.1);
}

.exp-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}

.exp-row strong {
  white-space: nowrap;
}

.exp-bar {
  width: 100%;
  height: 9px;
  background: rgba(0, 0, 0, 0.07);
  border-radius: 999px;
  overflow: hidden;
}

.exp-fill {
  height: 100%;
  background: #54a0ff;
  border-radius: 999px;
  transition: width 0.3s ease-out;
}

.log-btn {
  width: 100%;
  border: none;
  border-radius: 999px;
  padding: 9px 13px;
  color: #416d9e;
  background: rgba(84, 160, 255, 0.14);
  font-size: 13px;
  font-weight: 900;
  cursor: pointer;
  transition: transform 0.18s ease, background 0.18s ease;
}

.log-btn:hover {
  transform: translateY(-1px);
  background: rgba(84, 160, 255, 0.2);
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 10px;
}

.action-btn {
  min-width: 0;
  min-height: 70px;
  padding: 12px 10px;
  border: none;
  border-radius: 16px;
  cursor: pointer;
  color: #3d4654;
  background: rgba(0, 0, 0, 0.05);
  transition: transform 0.18s ease, background 0.18s ease, opacity 0.18s ease;
}

.action-btn span,
.action-btn small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-btn span {
  font-size: 15px;
  line-height: 1.2;
  font-weight: 900;
}

.action-btn small {
  margin-top: 6px;
  color: rgba(61, 70, 84, 0.64);
  font-size: 11px;
  font-weight: 800;
}

.action-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  background: rgba(0, 0, 0, 0.08);
}

.action-btn:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.feed-btn {
  background: rgba(255, 159, 67, 0.16);
}

.play-btn {
  background: rgba(29, 209, 161, 0.14);
}

.sleep-btn {
  background: rgba(84, 160, 255, 0.14);
}

.touch-btn {
  background: rgba(255, 107, 129, 0.14);
}

.bath-btn {
  background: rgba(72, 219, 251, 0.16);
}

.music-btn {
  background: rgba(95, 92, 255, 0.12);
}

.wish-btn {
  background: rgba(255, 205, 92, 0.2);
}

.dress-btn {
  background: rgba(255, 107, 129, 0.12);
}

.message-board {
  position: relative;
  z-index: 8;
  min-height: 100vh;
  padding: 24px 22px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: rgba(255, 255, 255, 0.76);
  border-left: 1px solid rgba(255, 255, 255, 0.9);
  box-shadow: -18px 0 42px rgba(36, 44, 58, 0.08);
  backdrop-filter: blur(20px);
}

.message-board-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.message-eyebrow {
  margin: 0 0 4px;
  color: #ff8a3d;
  font-size: 12px;
  font-weight: 900;
}

.message-board-header h3 {
  margin: 0;
  color: #2f3440;
  font-size: 20px;
  font-weight: 900;
}

.message-refresh,
.delete-message,
.message-actions button {
  border: none;
  cursor: pointer;
  font-weight: 900;
  transition: opacity 0.18s ease, transform 0.18s ease, background 0.18s ease;
}

.message-refresh {
  border-radius: 999px;
  padding: 8px 13px;
  color: #416d9e;
  background: rgba(84, 160, 255, 0.13);
}

.message-form {
  padding: 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.7);
  box-shadow: 0 14px 32px rgba(42, 50, 64, 0.08);
}

.message-form textarea {
  width: 100%;
  height: 94px;
  resize: none;
  border: 1px solid rgba(47, 52, 64, 0.1);
  border-radius: 14px;
  outline: none;
  padding: 12px;
  color: #2f3440;
  background: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  line-height: 1.55;
  box-sizing: border-box;
}

.message-form textarea:focus {
  border-color: rgba(84, 160, 255, 0.52);
  box-shadow: 0 0 0 3px rgba(84, 160, 255, 0.12);
}

.message-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 9px;
  color: #9aa3b2;
  font-size: 12px;
  font-weight: 800;
}

.message-actions button {
  border-radius: 999px;
  padding: 9px 18px;
  color: #fff;
  background: #ff9f43;
  box-shadow: 0 10px 22px rgba(255, 159, 67, 0.2);
}

.message-list {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding-right: 3px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.empty-message {
  margin: auto;
  color: #8a94a6;
  font-size: 14px;
  text-align: center;
  line-height: 1.6;
}

.message-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  padding: 13px 13px 12px;
  border-radius: 16px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(255, 250, 241, 0.88));
  border: 1px solid rgba(255, 255, 255, 0.92);
  box-shadow: 0 12px 26px rgba(42, 50, 64, 0.07);
}

.message-card-main {
  min-width: 0;
}

.message-card p {
  margin: 0;
  color: #3a4050;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.55;
  font-size: 14px;
}

.message-card time {
  display: block;
  margin-top: 8px;
  color: #9aa3b2;
  font-size: 12px;
  font-weight: 800;
}

.delete-message {
  flex: none;
  padding: 6px 9px;
  border-radius: 999px;
  color: #d75656;
  background: rgba(255, 107, 107, 0.12);
  font-size: 12px;
}

.message-refresh:hover:not(:disabled),
.delete-message:hover:not(:disabled),
.message-actions button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.message-refresh:disabled,
.delete-message:disabled,
.message-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.log-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: flex;
  justify-content: flex-end;
  background: rgba(38, 45, 58, 0.28);
  backdrop-filter: blur(6px);
}

.log-panel {
  width: min(520px, 100%);
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(247, 250, 255, 0.94));
  box-shadow: -22px 0 50px rgba(31, 40, 56, 0.18);
  box-sizing: border-box;
}

.log-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.log-panel-header h3 {
  margin: 0;
  font-size: 22px;
  color: #2f3440;
  font-weight: 900;
}

.log-list {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-right: 3px;
}

.log-card {
  display: flex;
  gap: 12px;
  padding: 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(255, 255, 255, 0.94);
  box-shadow: 0 12px 28px rgba(42, 50, 64, 0.08);
}

.log-icon {
  width: 42px;
  height: 42px;
  flex: none;
  display: grid;
  place-items: center;
  border-radius: 14px;
  color: #ff7d3d;
  background: rgba(255, 159, 67, 0.15);
  font-weight: 900;
}

.log-content {
  min-width: 0;
  flex: 1;
}

.log-title-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: baseline;
}

.log-title-row strong {
  color: #303744;
  font-size: 15px;
}

.log-title-row time {
  flex: none;
  color: #9aa3b2;
  font-size: 12px;
  font-weight: 800;
}

.log-content p {
  margin: 6px 0 9px;
  color: #596475;
  font-size: 13px;
  line-height: 1.55;
}

.log-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.log-meta span {
  max-width: 100%;
  padding: 5px 8px;
  border-radius: 999px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #6f7a8a;
  background: rgba(84, 160, 255, 0.1);
  font-size: 12px;
  font-weight: 800;
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

.status-bubble {
  position: absolute;
  top: -5%;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(255, 255, 255, 0.92);
  padding: 9px 16px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 800;
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

.anim-care,
.anim-ambient {
  animation: ambient 1.1s ease-in-out infinite;
  transform-origin: center bottom;
}

.anim-sleep {
  animation: sleep 3.5s ease-in-out infinite;
  transform-origin: center bottom;
}

@keyframes messageDrift {
  0% {
    opacity: 0;
    transform: translateX(0) translateY(8px);
  }
  8%,
  82% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translateX(calc(-100vw - 520px)) translateY(-8px);
  }
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

@media (max-width: 1180px) {
  .lulu-viewport {
    grid-template-columns: 1fr;
    overflow: auto;
  }

  .lulu-stage {
    min-height: 780px;
  }

  .fun-hud {
    position: relative;
    left: auto;
    top: auto;
    width: 100%;
    grid-row: 2;
    align-self: start;
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .message-board {
    min-height: 420px;
    border-left: none;
    border-top: 1px solid rgba(255, 255, 255, 0.9);
  }
}

@media (max-width: 860px) {
  .lulu-viewport {
    min-height: auto;
    overflow: visible;
  }

  .lulu-stage {
    min-height: 100svh;
    grid-template-rows: auto auto minmax(260px, 1fr) auto;
    align-items: stretch;
    gap: 10px;
    padding: max(10px, env(safe-area-inset-top)) 12px max(10px, env(safe-area-inset-bottom));
    overflow: clip;
  }

  .top-status {
    display: grid;
    grid-template-columns: minmax(112px, 0.72fr) minmax(0, 1.28fr);
    align-items: stretch;
    gap: 8px;
  }

  .identity-block {
    min-width: 0;
    gap: 8px;
    padding: 9px 10px;
    border-radius: 18px;
    background: rgba(255, 255, 255, 0.72);
    box-shadow: 0 12px 28px rgba(40, 49, 66, 0.08);
    backdrop-filter: blur(14px);
  }

  .identity-block h2 {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 20px;
  }

  .identity-block p {
    margin-top: 3px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 12px;
  }

  .level-badge {
    min-width: 42px;
    width: 42px;
    height: 42px;
    border-radius: 14px;
    font-size: 13px;
    box-shadow: none;
  }

  .quick-stats {
    width: 100%;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 6px;
  }

  .stat-pill {
    min-width: 0;
    padding: 8px 7px;
    border-radius: 14px;
  }

  .stat-pill span {
    font-size: 11px;
  }

  .stat-pill strong {
    margin-top: 1px;
    font-size: 17px;
  }

  .stat-bar {
    height: 5px;
    margin-top: 6px;
  }

  .fun-hud {
    grid-row: 2;
    grid-template-columns: minmax(0, 1.35fr) minmax(92px, 0.65fr);
    gap: 8px;
  }

  .daily-card,
  .streak-card {
    min-width: 0;
    padding: 9px 10px;
    border-radius: 16px;
  }

  .daily-card strong,
  .streak-card strong {
    margin-top: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
  }

  .daily-card small,
  .streak-card small {
    margin-top: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
  }

  .mission-progress {
    height: 5px;
    margin-top: 6px;
  }

  .lulu-entity {
    grid-row: 3;
    align-self: center;
    width: min(76vw, 42svh, 360px);
  }

  .action-dock {
    grid-row: 4;
    position: sticky;
    bottom: max(8px, env(safe-area-inset-bottom));
    grid-template-columns: 1fr;
    gap: 8px;
    align-self: end;
    margin-inline: -4px;
    padding: 10px;
    border-radius: 20px;
    box-shadow: 0 16px 40px rgba(40, 49, 66, 0.16);
  }

  .action-grid {
    display: grid;
    grid-auto-flow: column;
    grid-auto-columns: 84px;
    grid-template-columns: none;
    gap: 8px;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    scroll-snap-type: x proximity;
    padding-bottom: 2px;
    scrollbar-width: none;
  }

  .action-grid::-webkit-scrollbar {
    display: none;
  }

  .action-btn {
    min-height: 56px;
    scroll-snap-align: start;
    padding: 9px 8px 8px;
    border-radius: 15px;
  }

  .action-btn span {
    font-size: 14px;
  }

  .action-btn small {
    margin-top: 4px;
    font-size: 10px;
  }

  .progress-strip {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 110px;
    align-items: center;
    gap: 9px;
    padding: 9px 10px;
    border-radius: 15px;
  }

  .exp-row {
    grid-column: 1 / -1;
  }

  .log-btn {
    grid-row: 2;
    grid-column: 2;
    padding: 8px 10px;
  }

  .exp-bar {
    grid-row: 2;
    grid-column: 1;
  }

  .message-board {
    padding: 18px 14px;
  }

  .message-fly-zone {
    inset: 126px 0 118px;
  }

}

@media (max-width: 520px) {
  .lulu-stage {
    grid-template-rows: auto auto minmax(240px, 1fr) auto;
    gap: 8px;
    padding-inline: 10px;
  }

  .top-status {
    grid-template-columns: 1fr;
  }

  .identity-block {
    padding: 8px 9px;
  }

  .quick-stats {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .fun-hud {
    grid-template-columns: minmax(0, 1fr) minmax(86px, 0.56fr);
  }

  .daily-card,
  .streak-card {
    padding: 8px 9px;
  }

  .daily-card span,
  .streak-card span {
    font-size: 10px;
  }

  .lulu-entity {
    width: min(84vw, 40svh, 330px);
  }

  .status-bubble {
    top: -3%;
    max-width: 86vw;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
  }

  .action-dock {
    padding: 8px;
    border-radius: 18px;
  }

  .action-grid {
    grid-auto-columns: 76px;
  }

  .action-btn {
    min-height: 54px;
  }

  .message-board-header {
    align-items: flex-start;
  }

  .message-fly-zone {
    inset: 118px 0 108px;
  }

  .log-panel {
    padding: 18px 14px;
  }

  .log-title-row {
    flex-direction: column;
    gap: 3px;
  }

}
</style>
