<template>
  <div class="lulu-viewport">
    <section
      ref="stageRef"
      class="lulu-stage"
      :class="[`scene-${scenePeriod}`, { 'has-custom-scene': Boolean(currentScene.image) }]"
      @pointermove="handleStagePointerMove"
      @pointerleave="resetLuluMotion"
    >
      <div
        :key="currentScene.name"
        class="scene-background"
        :class="{ 'has-scene-image': Boolean(currentScene.image) }"
        :style="sceneBackgroundStyle"
      ></div>

      <div class="top-status">
        <div class="identity-block">
          <span class="level-badge">Lv.{{ displayLevel }}</span>
          <div>
            <h2>{{ petData.name || '噜噜' }}</h2>
            <p>{{ formatState(petData.currentState) }}</p>
            <div v-if="syncNotice.visible" class="sync-notice" :class="`sync-${syncNotice.mode}`" role="status">
              <span>{{ syncNotice.text }}</span>
              <button v-if="syncNotice.retry" type="button" @click="fetchStatus()">重试</button>
            </div>
          </div>
        </div>

        <div class="quick-stats">
          <div class="stat-pill stat-hunger">
            <span>饱腹</span>
            <strong>{{ displayStat(petData.hunger) }}</strong>
            <div class="stat-bar">
              <div class="stat-fill" :style="{ width: `${petData.hunger || 0}%` }"></div>
            </div>
          </div>
          <div class="stat-pill stat-energy">
            <span>体力</span>
            <strong>{{ displayStat(petData.energy) }}</strong>
            <div class="stat-bar">
              <div class="stat-fill" :style="{ width: `${petData.energy || 0}%` }"></div>
            </div>
          </div>
          <div class="stat-pill stat-mood">
            <span>心情</span>
            <strong>{{ displayStat(petData.mood) }}</strong>
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
          <span>本月陪伴</span>
          <strong>{{ monthlyCompanionship.visitedDays }}/{{ monthlyCompanionship.elapsedDays }} 天</strong>
          <div class="companion-progress" :title="`本月已过 ${monthlyCompanionship.elapsedDays} 天`">
            <div class="companion-fill" :style="{ width: `${monthlyCompanionshipProgress}%` }"></div>
          </div>
          <small>缺席 {{ monthlyCompanionship.missedDays }} 天 · 连续 {{ careStreak }} 天</small>
        </div>
        <button class="residency-card" :class="`status-${worldData.residency.status.toLowerCase()}`" type="button" @click="openWorldPanel">
          <span>噜妹入住计划 · Lv.55</span>
          <strong>{{ worldData.residency.title }}</strong>
          <div class="residency-progress">
            <div class="residency-fill" :style="{ width: `${lumeiResidencyProgress}%` }"></div>
          </div>
          <small v-if="worldData.residency.resident">
            已经常驻 · 本周任务 {{ worldData.weekly.completedCount }}/{{ worldData.weekly.totalCount }}
          </small>
          <small v-else>当前 Lv.{{ worldData.residency.currentLevel }} · 还差 {{ worldData.residency.remainingLevels }} 级</small>
        </button>
        <button class="community-card" type="button" @click="openWorldPanel">
          <span>全站共同目标</span>
          <strong>{{ worldData.communityGoal.title }}</strong>
          <div class="community-progress">
            <div class="community-fill" :class="{ completed: worldData.communityGoal.completed }" :style="{ width: `${communityGoalProgress}%` }"></div>
          </div>
          <small>
            {{ worldData.communityGoal.current }}/{{ worldData.communityGoal.target }}
            · {{ worldData.memory.title }}
          </small>
        </button>
      </div>

      <aside v-if="showLumeiLetter" class="npc-letter-popover" aria-live="polite">
        <img :src="npcImages.letter" alt="噜妹抱着一封信" />
        <div class="npc-letter-copy">
          <span>噜妹来啦</span>
          <strong>{{ activeNPCEvent.title }}</strong>
          <p>{{ activeNPCEvent.content }}</p>
          <div>
            <button type="button" @click="openWorldPanel">打开回忆册</button>
            <button type="button" class="letter-dismiss" @click="dismissNPCEvent">收到啦</button>
          </div>
        </div>
      </aside>

      <aside v-if="isLumeiAway" class="lumei-away-note" aria-live="polite">
        <span>噜妹的小便签</span>
        <strong>{{ activeNPCEvent.title }}</strong>
        <p>{{ activeNPCEvent.content }}</p>
        <button type="button" @click="openWorldPanel">查看往来记录</button>
      </aside>

      <div
        class="lulu-entity"
        :class="[`state-${visualState.toLowerCase()}`, `action-${actionCategory}`, { 'is-away-event': isLuluAway }]"
        :style="luluMotionStyle"
      >
        <div v-if="bondCombo > 1" class="bond-badge">默契 ×{{ bondCombo }}</div>
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

        <img
          class="lulu-img"
          :class="animationClass"
          :src="currentLuluImage"
          :alt="isLuluAway ? '噜噜和噜妹一起外出' : '噜噜'"
          @error="handleLuluImageError"
        />

        <button
          v-if="lumeiIsResident && !isLuluAway && !isLumeiAway"
          class="lumei-resident"
          :class="[`pose-${lumeiPose}`, { 'is-talking': Boolean(lumeiSpeechText) }]"
          type="button"
          aria-label="和常驻的噜妹说说话"
          @click.stop="interactWithLumei"
        >
          <span v-if="lumeiSpeechText" class="lumei-speech" aria-live="polite">{{ lumeiSpeechText }}</span>
          <img :src="currentLumeiImage" :alt="lumeiPose === 'sitting' ? '坐在噜噜旁边的噜妹' : '站在噜噜旁边的噜妹'" @error="handleLuluImageError" />
          <small>噜妹</small>
        </button>

        <div v-if="!isLuluAway" class="body-hotspots" aria-label="点击噜噜互动">
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
          <p class="care-advice"><strong>噜噜心声</strong>{{ smartCareAdvice.text }}</p>
          <button class="smart-care-btn" type="button" @click="smartCareLulu" :disabled="isActionDisabled">
            <span>智能照顾</span>
            <small>{{ smartCareAdvice.short }}</small>
          </button>
          <button class="log-btn" type="button" @click="openLogPanel">噜噜 日志</button>
        </div>

        <div class="action-grid">
          <button class="action-btn feed-btn" @click="feedLulu" :disabled="isActionDisabled">
            <span>喂食</span>
            <small>饱腹 +30</small>
          </button>
          <button class="action-btn play-btn" @click="playLulu" :disabled="isActionDisabled">
            <span>玩耍</span>
            <small>心情 +20</small>
          </button>
          <button class="action-btn sleep-btn" @click="sleepLulu" :disabled="isActionDisabled">
            <span>{{ sleepBtnText }}</span>
            <small>恢复体力</small>
          </button>
          <button class="action-btn touch-btn" @click="touchLulu" :disabled="isActionDisabled">
            <span>摸摸</span>
            <small>陪伴一下</small>
          </button>
          <button class="action-btn bath-btn" @click="bathLulu" :disabled="isActionDisabled">
            <span>洗澡</span>
            <small>清爽状态</small>
          </button>
          <button class="action-btn music-btn" @click="musicLulu" :disabled="isActionDisabled">
            <span>听音乐</span>
            <small>放松心情</small>
          </button>
          <button class="action-btn wish-btn" @click="makeWish" :disabled="isActionDisabled">
            <span>许愿</span>
            <small>{{ wishText }}</small>
          </button>
          <button class="action-btn dress-btn" @click="changeAccessory" :disabled="isActionDisabled">
            <span>换装</span>
            <small>{{ currentAccessory.name }}</small>
          </button>
          <button class="action-btn scene-btn" @click="changeScene" :disabled="isActionDisabled">
            <span>换场景</span>
            <small>{{ currentScene.name }}</small>
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
        <button class="message-refresh" @click="fetchMessages(messagePagination.page)" :disabled="isMessageLoading">刷新</button>
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
      <nav class="pager" aria-label="留言分页">
        <button type="button" @click="changeMessagePage(messagePagination.page - 1)" :disabled="isMessageLoading || messagePagination.page <= 1">上一页</button>
        <span>第 {{ messagePagination.page }}/{{ messagePagination.totalPages }} 页 · {{ messagePagination.total }} 条</span>
        <button type="button" @click="changeMessagePage(messagePagination.page + 1)" :disabled="isMessageLoading || messagePagination.page >= messagePagination.totalPages">下一页</button>
      </nav>
    </aside>

    <div v-if="isLogPanelOpen" class="log-overlay" @click.self="isLogPanelOpen = false">
      <section class="log-panel">
        <div class="log-panel-header">
          <div>
            <p class="message-eyebrow">互动日志</p>
            <h3>噜噜 的照顾记录 <small>{{ logPagination.total }} 条</small></h3>
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
        <nav class="pager log-pager" aria-label="日志分页">
          <button type="button" @click="changeLogPage(logPagination.page - 1)" :disabled="isLogLoading || logPagination.page <= 1">上一页</button>
          <span>第 {{ logPagination.page }}/{{ logPagination.totalPages }} 页</span>
          <button type="button" @click="changeLogPage(logPagination.page + 1)" :disabled="isLogLoading || logPagination.page >= logPagination.totalPages">下一页</button>
        </nav>
      </section>
    </div>

    <div v-if="isWorldPanelOpen" class="world-overlay" @click.self="isWorldPanelOpen = false">
      <section class="world-panel">
        <div class="world-panel-header">
          <div>
            <p class="message-eyebrow">噜噜世界</p>
            <h3>噜妹来信与我们的回忆</h3>
          </div>
          <button class="message-refresh" type="button" @click="isWorldPanelOpen = false">关闭</button>
        </div>

        <div v-if="isWorldLoading" class="empty-message">回忆册读取中...</div>

        <template v-else>
          <article class="residency-world-card" :class="`status-${worldData.residency.status.toLowerCase()}`">
            <img :src="npcImages.resident" alt="噜妹入住形象" />
            <div class="residency-world-copy">
              <span>噜妹入住计划 · Lv.{{ worldData.residency.targetLevel }}</span>
              <strong>{{ worldData.residency.title }}</strong>
              <p>{{ worldData.residency.message }}</p>
              <div class="residency-progress large">
                <div class="residency-fill" :style="{ width: `${lumeiResidencyProgress}%` }"></div>
              </div>
              <small v-if="worldData.residency.resident">
                {{ worldData.residency.residentSince ? `${worldData.residency.residentSince} 正式入住` : '已经正式入住' }}
              </small>
              <small v-else>{{ worldData.residency.nextMilestoneText }}</small>
            </div>
            <div class="residency-story" aria-label="噜妹入住剧情进度">
              <div
                v-for="chapter in worldData.residency.chapters"
                :key="chapter.level"
                class="story-chapter"
                :class="{ unlocked: chapter.unlocked, current: chapter.current }"
              >
                <span>Lv.{{ chapter.level }}</span>
                <strong>{{ chapter.title }}</strong>
                <small>{{ chapter.unlocked ? chapter.description : '继续陪伴噜噜后解锁' }}</small>
              </div>
            </div>
          </article>

          <article v-if="lumeiIsResident" class="move-in-memory-card">
            <img :src="npcImages.moveInMemory" alt="噜噜欢迎噜妹入住的纪念画面" />
            <div>
              <span>入住纪念</span>
              <strong>从远方来信，到每天都在身边</strong>
              <p>{{ worldData.residency.residentSince || 'Lv.55 解锁日' }} · 噜噜和噜妹拥有了共同的家</p>
            </div>
          </article>

          <article v-if="worldData.weekly.unlocked" class="weekly-task-card">
            <div class="weekly-task-heading">
              <div>
                <span>你和噜妹的本周双人任务</span>
                <strong>{{ worldData.weekly.allCompleted ? '本周的小约定全部完成啦' : '和噜噜、噜妹一起完成' }}</strong>
              </div>
              <b>{{ worldData.weekly.completedCount }}/{{ worldData.weekly.totalCount }}</b>
            </div>
            <small>{{ worldData.weekly.weekStart }} — {{ worldData.weekly.weekEnd }} · 每周一自动更新</small>
            <div class="weekly-task-list">
              <div v-for="task in worldData.weekly.tasks" :key="task.id" class="weekly-task-item" :class="{ completed: task.completed }">
                <div class="weekly-task-title">
                  <span>{{ task.completed ? '✓' : task.order }}</span>
                  <div>
                    <strong>{{ task.title }}</strong>
                    <p>{{ task.description }}</p>
                  </div>
                  <b>{{ Math.min(task.current, task.target) }}/{{ task.target }}</b>
                </div>
                <div class="weekly-task-progress">
                  <div :style="{ width: `${task.progress}%` }"></div>
                </div>
                <small>{{ task.completed ? task.reward : `还差 ${task.remaining} 次 · ${task.reward}` }}</small>
              </div>
            </div>
          </article>

          <article class="memory-card">
            <div class="memory-heading">
              <span>当前网络的专属记忆</span>
              <strong>{{ worldData.memory.title }}</strong>
            </div>
            <p class="memory-signature">{{ worldData.memory.signature }}</p>
            <ul>
              <li v-for="line in worldData.memory.lines" :key="line">{{ line }}</li>
            </ul>
          </article>

          <article class="world-goal-card">
            <div class="world-goal-heading">
              <div>
                <span>今天的全站共同目标</span>
                <strong>{{ worldData.communityGoal.title }}</strong>
              </div>
              <b>{{ worldData.communityGoal.current }}/{{ worldData.communityGoal.target }}</b>
            </div>
            <div class="community-progress large">
              <div class="community-fill" :class="{ completed: worldData.communityGoal.completed }" :style="{ width: `${communityGoalProgress}%` }"></div>
            </div>
            <p>
              {{ worldData.communityGoal.participants }} 位朋友参与
              · {{ worldData.communityGoal.completed ? worldData.communityGoal.reward : `还差 ${worldData.communityGoal.remaining} 次` }}
            </p>
          </article>

          <article class="npc-history-card">
            <div class="npc-history-heading">
              <img :src="npcImages.letter" alt="噜妹" />
              <div>
                <span>噜妹往来日志</span>
                <strong>{{ worldData.history.length ? '这些小事都被保存下来了' : (lumeiIsResident ? '从远方到身边，故事还会继续' : '第一封信还在路上') }}</strong>
              </div>
            </div>
            <div v-if="!worldData.history.length" class="empty-message compact">
              {{ lumeiIsResident ? '噜妹已经住下，过去没有留下的信件不影响接下来的陪伴。' : '噜妹会偶尔寄信，也可能带噜噜出去玩。' }}
            </div>
            <div v-else class="npc-history-list">
              <div v-for="event in worldData.history" :key="event.id" class="npc-history-item">
                <span>{{ event.type === 'OUTING' ? '游' : (event.type === 'LUMEI_OUTING' ? '出' : '信') }}</span>
                <div>
                  <strong>{{ event.title }}</strong>
                  <p>{{ event.content }}</p>
                  <time>{{ event.eventDate }}</time>
                </div>
              </div>
            </div>
          </article>
        </template>
      </section>
    </div>

    <div v-if="showLumeiMoveIn" class="move-in-overlay" role="dialog" aria-modal="true" aria-labelledby="lumei-move-in-title">
      <section class="move-in-card">
        <div class="move-in-sparkles" aria-hidden="true"><span>✦</span><span>♥</span><span>✦</span></div>
        <img class="move-in-memory-image" :src="npcImages.moveInMemory" alt="噜噜欢迎噜妹入住的纪念画面" />
        <p>Lv.55 常驻角色解锁</p>
        <h3 id="lumei-move-in-title">噜妹正式入住啦！</h3>
        <strong>“以后不用只在信里见面了，我会和噜噜一起在这里等你。”</strong>
        <small>从现在起，噜妹会常驻在噜噜身边。点击她，还能听到新的悄悄话。</small>
        <button type="button" @click="acknowledgeLumeiMoveIn">欢迎回家</button>
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
const statusSyncState = ref('loading');
const coreImagesReady = ref(false);
const failedImageUrls = ref(new Set());
const isMessageLoading = ref(false);
const actionCategory = ref('idle');
const actionLabel = ref('');
const actionFrames = ref([]);
const actionFrameIndex = ref(0);
const floatingEffects = ref([]);
const messages = ref([]);
const messageInput = ref('');
const messageListRef = ref(null);
const messagePagination = ref({ page: 1, pageSize: 8, total: 0, totalPages: 1 });
const logs = ref([]);
const logPagination = ref({ page: 1, pageSize: 10, total: 0, totalPages: 1 });
const isLogLoading = ref(false);
const isLogPanelOpen = ref(false);
const isWorldLoading = ref(false);
const isWorldPanelOpen = ref(false);
const dismissedNPCEventId = ref(null);
const showLumeiMoveIn = ref(false);
const residentLine = ref('');
const lumeiPose = ref(Math.random() < 0.46 ? 'sitting' : 'standing');
const worldData = ref({
  event: null,
  history: [],
  residency: {
    status: 'NPC',
    resident: false,
    currentLevel: 1,
    targetLevel: 55,
    remainingLevels: 54,
    progress: 1,
    title: '噜妹还住在远方',
    message: '她会偶尔寄信，也会悄悄来找噜噜玩。',
    nextMilestoneLevel: 45,
    nextMilestoneText: '45 级时，噜妹会说出想留下来的心愿。',
    residentSince: '',
    chapters: [
      { level: 45, title: '想留下来的信', description: '', unlocked: false, current: false },
      { level: 50, title: '准备一个小房间', description: '', unlocked: false, current: false },
      { level: 54, title: '最后一只搬家箱', description: '', unlocked: false, current: false },
      { level: 55, title: '从远方到身边', description: '', unlocked: false, current: false }
    ]
  },
  memory: {
    title: '今天认识的新朋友',
    signature: '噜噜正在翻开新的回忆页。',
    lines: []
  },
  communityGoal: {
    title: '大家一起陪陪噜噜',
    current: 0,
    target: 20,
    remaining: 20,
    participants: 0,
    completed: false,
    reward: '全站解锁一整天的温暖心情'
  },
  weekly: {
    unlocked: false,
    weekKey: '',
    weekStart: '',
    weekEnd: '',
    completedCount: 0,
    totalCount: 0,
    allCompleted: false,
    tasks: []
  }
});
const careStreak = ref(1);
const monthlyCompanionship = ref({
  month: '',
  visitedDays: 0,
  missedDays: new Date().getDate(),
  elapsedDays: new Date().getDate(),
  daysInMonth: new Date(new Date().getFullYear(), new Date().getMonth() + 1, 0).getDate(),
  visitedDates: []
});
const stageRef = ref(null);
const luluMotion = ref({ x: 0, y: 0, rotate: 0 });
const ambientThought = ref('');
const bondCombo = ref(0);
const scenePeriod = ref((() => {
  const hour = new Date().getHours();
  if (hour < 6) return 'night';
  if (hour < 11) return 'morning';
  if (hour < 18) return 'day';
  if (hour < 22) return 'evening';
  return 'night';
})());
const dailyMission = ref({
  type: 'feed',
  title: '给噜噜准备一顿饭',
  current: 0,
  target: 1,
  reward: '完成后心情会亮一下'
});
const accessoryIndex = ref(0);
const sceneIndex = ref((() => {
  if (typeof window === 'undefined') return 0;
  const stored = Number(window.localStorage.getItem('luluSceneIndex'));
  return Number.isInteger(stored) && stored >= 0 && stored < 6 ? stored : 0;
})());
const wishText = ref('抽一句');

let actionTimer = null;
let frameTimer = null;
let ambientTimer = null;
let thoughtTimer = null;
let thoughtClearTimer = null;
let comboTimer = null;
let motionFrame = null;
let pollerTimer = null;
let worldRefreshTimer = null;
let residentLineTimer = null;
let lumeiPoseTimer = null;
let effectIdCounter = 0;
let lastInteractionAt = 0;
let lastAnnouncedNPCEventId = null;

const img = (name) => `/picture/lulu/benti/${name}.webp`;

const npcImages = {
  letter: '/picture/lulu/npc/lumei-letter.webp',
  outing: '/picture/lulu/npc/lulu-lumei-outing.webp',
  resident: '/picture/lulu/npc/lumei-resident.webp',
  residentSitting: '/picture/lulu/npc/lumei-resident-sitting.webp',
  moveInMemory: '/picture/lulu/npc/lulu-lumei-move-in-memory.webp'
};

const luluImages = {
  idle: img('lulu_fadai'),
  happy: img('lulu_kaixin'),
  feed: img('lulu_eat'),
  feedNoodle: img('lulu_eat_noodle'),
  feedCookie: img('lulu_eat_cookie'),
  play: img('lulu_play'),
  playChase: img('lulu_play_chase'),
  playDance: img('lulu_play_dance'),
  idleBook: img('lulu_idle_book'),
  idleStretch: img('lulu_idle_stretch'),
  idleBubble: img('lulu_idle_bubble'),
  idleHoodie: img('lulu_fadai_hoodie'),
  idleOveralls: img('lulu_fadai_overalls'),
  idleScarf: img('lulu_fadai_scarf'),
  idleCrossbody: img('lulu_fadai_crossbody'),
  idlePajamas: img('lulu_fadai_pajamas'),
  idleGreenpants: img('lulu_fadai_greenpants'),
  idleGreenpantsPuffer: img('lulu_fadai_greenpants_puffer'),
  idleWatermelon: img('lulu_fadai_watermelon'),
  idleSharkSlippers: img('lulu_fadai_shark_slippers'),
  idleDinosaurBox: img('lulu_fadai_dinosaur_box'),
  idleCockroach: img('lulu_fadai_cockroach'),
  idleTie: img('lulu_fadai_tie'),
  sleepFloor: img('lulu_sleep_floor'),
  sleepHoodie: img('lulu_sleep_hoodie'),
  sleepOveralls: img('lulu_sleep_overalls'),
  sleepScarf: img('lulu_sleep_scarf'),
  sleepCrossbody: img('lulu_sleep_crossbody'),
  sleepPajamas: img('lulu_sleep_pajamas'),
  sleepGreenpants: img('lulu_sleep_greenpants'),
  sleepGreenpantsPuffer: img('lulu_sleep_greenpants_puffer'),
  sleepWatermelon: img('lulu_sleep_watermelon'),
  sleepSharkSlippers: img('lulu_sleep_shark_slippers'),
  sleepDinosaurBox: img('lulu_sleep_dinosaur_box'),
  sleepCockroach: img('lulu_sleep_cockroach'),
  sleepTie: img('lulu_sleep_tie'),
  touch: img('lulu_touch'),
  bath: img('lulu_bath'),
  music: img('lulu_music'),
  tapHead: img('lulu_tap_head'),
  tapBelly: img('lulu_tap_belly'),
  tapFoot: img('lulu_tap_foot')
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
  { label: '正在吃小蛋糕', frames: [luluImages.feed, luluImages.feedCookie, luluImages.feed], effect: ['点心时间', 'type-food'] },
  { label: '正在吸溜面条', frames: [luluImages.feed, luluImages.feedNoodle, luluImages.feed], effect: ['热乎乎', 'type-food'] },
  { label: '正在认真干饭', frames: [luluImages.feed, luluImages.feedCookie, luluImages.feed], effect: ['吃饱啦', 'type-food'] }
];

const playActions = [
  { label: '追球中', frames: [luluImages.play, luluImages.playChase, luluImages.play], effect: ['跑起来', 'type-mood'] },
  { label: '开心跳舞', frames: [luluImages.play, luluImages.playDance, luluImages.play], effect: ['心情明亮', 'type-mood'] },
  { label: '蹦蹦跳跳', frames: [luluImages.play, luluImages.playDance, luluImages.playChase], effect: ['玩疯了', 'type-mood'] }
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
  { name: '初始衣服', image: luluImages.idle, sleepImage: luluImages.sleepFloor },
  { name: '蓝色卫衣', image: luluImages.idleHoodie, sleepImage: luluImages.sleepHoodie },
  { name: '绿色背带裤', image: luluImages.idleOveralls, sleepImage: luluImages.sleepOveralls },
  { name: '红围巾套装', image: luluImages.idleScarf, sleepImage: luluImages.sleepScarf },
  { name: '斜挎小包', image: luluImages.idleCrossbody, sleepImage: luluImages.sleepCrossbody },
  { name: '睡衣噜', image: luluImages.idlePajamas, sleepImage: luluImages.sleepPajamas },
  { name: '绿裤衩噜', image: luluImages.idleGreenpants, sleepImage: luluImages.sleepGreenpants },
  { name: '河豚包噜', image: luluImages.idleGreenpantsPuffer, sleepImage: luluImages.sleepGreenpantsPuffer },
  { name: '西瓜噜', image: luluImages.idleWatermelon, sleepImage: luluImages.sleepWatermelon },
  { name: '鲨鱼拖鞋噜', image: luluImages.idleSharkSlippers, sleepImage: luluImages.sleepSharkSlippers },
  { name: '恐龙噜', image: luluImages.idleDinosaurBox, sleepImage: luluImages.sleepDinosaurBox },
  { name: '蟑螂噜', image: luluImages.idleCockroach, sleepImage: luluImages.sleepCockroach },
  { name: '领带噜', image: luluImages.idleTie, sleepImage: luluImages.sleepTie }
];

const sceneModes = [
  { name: '原始背景', image: '' },
  { name: '暖暖小屋', image: '/picture/lulu/scenes/scene-cozy-room.webp' },
  { name: '阳光花园', image: '/picture/lulu/scenes/scene-sunny-garden.webp' },
  { name: '森林空地', image: '/picture/lulu/scenes/scene-forest-clearing.webp' },
  { name: '海边露台', image: '/picture/lulu/scenes/scene-seaside-terrace.webp' },
  { name: '星空营地', image: '/picture/lulu/scenes/scene-starry-camp.webp' }
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

const activeNPCEvent = computed(() => worldData.value.event?.active ? worldData.value.event : null);

const isLuluAway = computed(() => activeNPCEvent.value?.type === 'OUTING');

const isLumeiAway = computed(() => activeNPCEvent.value?.type === 'LUMEI_OUTING');

const showLumeiLetter = computed(() => {
  return activeNPCEvent.value?.type === 'LETTER' && dismissedNPCEventId.value !== activeNPCEvent.value.id;
});

const lumeiIsResident = computed(() => Boolean(worldData.value.residency.resident));

const currentLumeiImage = computed(() => {
  if (petData.value.currentState === 'SLEEPING' || lumeiPose.value === 'sitting') {
    return npcImages.residentSitting;
  }
  return npcImages.resident;
});

const lumeiSpeechText = computed(() => {
  if (residentLine.value) return residentLine.value;
  if (petData.value.currentState === 'SLEEPING') return '嘘，噜噜睡着啦，我们小声一点。';
  return '';
});

const lumeiResidencyProgress = computed(() => {
  return Math.max(0, Math.min(100, Number(worldData.value.residency.progress) || 0));
});

const currentLuluImage = computed(() => {
  let requestedImage;
  if (isLuluAway.value) {
    requestedImage = npcImages.outing;
  } else if (actionFrames.value.length) {
    requestedImage = actionFrames.value[actionFrameIndex.value % actionFrames.value.length];
  } else if (petData.value.currentState === 'SLEEPING') {
    requestedImage = currentAccessory.value.sleepImage || luluImages.sleepFloor;
  } else {
    requestedImage = currentAccessory.value.image || luluImages.idle;
  }
  if (!failedImageUrls.value.has(requestedImage)) return requestedImage;
  if (petData.value.currentState === 'SLEEPING' && !failedImageUrls.value.has(luluImages.sleepFloor)) {
    return luluImages.sleepFloor;
  }
  return luluImages.idle;
});

const animationClass = computed(() => {
  if (isLuluAway.value) return 'anim-away';
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
  if (isLuluAway.value) return activeNPCEvent.value?.content || '噜噜和噜妹出去玩啦，晚点回来。';
  return actionLabel.value || ambientThought.value || formatState(visualState.value);
});

const luluMotionStyle = computed(() => ({
  '--lulu-shift-x': `${luluMotion.value.x}px`,
  '--lulu-shift-y': `${luluMotion.value.y}px`,
  '--lulu-rotate': `${luluMotion.value.rotate}deg`
}));

const sleepBtnText = computed(() => {
  return petData.value.currentState === 'SLEEPING' ? '唤醒' : '睡觉';
});

const isActionDisabled = computed(() => {
  return isLuluAway.value || isLoading.value || !coreImagesReady.value || !['ready', 'stale'].includes(statusSyncState.value);
});

const syncNotice = computed(() => {
  if (!coreImagesReady.value) {
    return { visible: true, mode: 'loading', text: '正在准备动作图片…', retry: false };
  }
  if (statusSyncState.value === 'loading') {
    return { visible: true, mode: 'loading', text: '正在同步噜噜状态…', retry: false };
  }
  if (statusSyncState.value === 'error') {
    return { visible: true, mode: 'error', text: '状态加载失败，互动已暂停', retry: true };
  }
  if (statusSyncState.value === 'stale') {
    return { visible: true, mode: 'stale', text: '状态同步中断，当前展示上次结果', retry: true };
  }
  return { visible: false, mode: 'ready', text: '', retry: false };
});

const floatingMessages = computed(() => {
  return messages.value.slice(0, 8).reverse();
});

const currentAccessory = computed(() => accessoryModes[accessoryIndex.value] || accessoryModes[0]);

const currentScene = computed(() => sceneModes[sceneIndex.value] || sceneModes[0]);

const sceneBackgroundStyle = computed(() => {
  if (!currentScene.value.image) return {};
  return { backgroundImage: `url("${currentScene.value.image}")` };
});

const dailyMissionProgress = computed(() => {
  if (!dailyMission.value.target) return 0;
  return Math.min(100, (dailyMission.value.current / dailyMission.value.target) * 100);
});

const monthlyCompanionshipProgress = computed(() => {
  if (!monthlyCompanionship.value.elapsedDays) return 0;
  return Math.min(100, (monthlyCompanionship.value.visitedDays / monthlyCompanionship.value.elapsedDays) * 100);
});

const communityGoalProgress = computed(() => {
  if (!worldData.value.communityGoal.target) return 0;
  return Math.min(100, (worldData.value.communityGoal.current / worldData.value.communityGoal.target) * 100);
});

const moodWeather = computed(() => {
  const mood = petData.value.mood || 0;
  if (mood >= 80) return { text: '今天是闪闪发亮日' };
  if (mood >= 50) return { text: '噜噜状态不错' };
  if (mood >= 25) return { text: '噜噜想被多陪陪' };
  return { text: '噜噜有点低落' };
});

const smartCareAdvice = computed(() => {
  if (petData.value.currentState === 'SLEEPING') {
    return { action: 'RESTING', short: '守护睡眠', text: '噜噜睡得很香，安静陪着它就好。' };
  }
  if (petData.value.hunger <= 45) {
    return { action: 'FEED', short: '建议加餐', text: '小肚子有点空，智能照顾会先准备加餐。' };
  }
  if (petData.value.energy <= 35) {
    return { action: 'REST', short: '建议休息', text: '体力不多了，智能照顾会安排噜噜休息。' };
  }
  if (petData.value.mood <= 65) {
    return { action: 'COMFORT', short: '需要陪伴', text: '噜噜想要一个摸摸，陪伴能让心情变好。' };
  }
  return { action: 'STROLL', short: '适合散步', text: '状态正好，今天很适合和噜噜出去走走。' };
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

const displayStat = (value) => {
  if (!['ready', 'stale'].includes(statusSyncState.value)) return '—';
  return Number.isFinite(Number(value)) ? Number(value) : 0;
};

const pickOne = (items) => items[Math.floor(Math.random() * items.length)];

const randomBetween = (min, max) => Math.floor(min + Math.random() * (max - min + 1));

const handleStagePointerMove = (event) => {
  if (event.pointerType === 'touch' || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  const stage = stageRef.value;
  if (!stage) return;
  const rect = stage.getBoundingClientRect();
  const normalizedX = Math.max(-1, Math.min(1, (event.clientX - (rect.left + rect.width / 2)) / (rect.width / 2)));
  const normalizedY = Math.max(-1, Math.min(1, (event.clientY - (rect.top + rect.height / 2)) / (rect.height / 2)));
  window.cancelAnimationFrame(motionFrame);
  motionFrame = window.requestAnimationFrame(() => {
    luluMotion.value = {
      x: Number((normalizedX * 8).toFixed(2)),
      y: Number((normalizedY * 5).toFixed(2)),
      rotate: Number((normalizedX * 1.6).toFixed(2))
    };
  });
};

const resetLuluMotion = () => {
  window.cancelAnimationFrame(motionFrame);
  luluMotion.value = { x: 0, y: 0, rotate: 0 };
};

const registerBondInteraction = () => {
  const now = Date.now();
  bondCombo.value = now - lastInteractionAt <= 8000 ? Math.min(5, bondCombo.value + 1) : 1;
  lastInteractionAt = now;
  window.clearTimeout(comboTimer);
  comboTimer = window.setTimeout(() => {
    bondCombo.value = 0;
  }, 8000);
  if (bondCombo.value > 1) {
    triggerEffect('默', `默契 ×${bondCombo.value}`, 'type-level');
  }
};

const getAmbientThoughts = () => {
  const thoughts = ['我在听哦。', '今天也一起慢慢来。', '你来啦，我刚好没有睡着。'];
  if (petData.value.hunger <= 35) thoughts.push('小肚子好像在咕咕叫。', '饭碗今天会出现吗？');
  if (petData.value.energy <= 35) thoughts.push('眼皮有一点点打架。', '要不要一起休息五分钟？');
  if (petData.value.mood >= 80) thoughts.push('今天的心情闪闪发亮！', '想和你多玩一会儿。');
  if (monthlyCompanionship.value.visitedDays > 1) thoughts.push(`这个月已经见到你 ${monthlyCompanionship.value.visitedDays} 天啦。`);
  if (scenePeriod.value === 'night') thoughts.push('夜深啦，别忘了早点休息。');
  if (scenePeriod.value === 'morning') thoughts.push('早呀，今天也要元气满满。');
  return thoughts;
};

const scheduleThought = () => {
  window.clearTimeout(thoughtTimer);
  thoughtTimer = window.setTimeout(() => {
    if (actionCategory.value === 'idle' && petData.value.currentState !== 'SLEEPING') {
      ambientThought.value = pickOne(getAmbientThoughts());
      window.clearTimeout(thoughtClearTimer);
      thoughtClearTimer = window.setTimeout(() => {
        ambientThought.value = '';
      }, 4200);
    }
    scheduleThought();
  }, randomBetween(6500, 12000));
};

const preloadImage = (url) => new Promise((resolve, reject) => {
  const image = new Image();
  image.onload = () => resolve(url);
  image.onerror = () => reject(new Error(`图片加载失败: ${url}`));
  image.src = url;
});

const preloadAccessoryImages = async (accessory) => {
  const urls = [...new Set([accessory?.image, accessory?.sleepImage].filter(Boolean))];
  const results = await Promise.allSettled(urls.map(preloadImage));
  const failed = results
    .filter((result) => result.status === 'rejected')
    .map((result) => String(result.reason?.message || result.reason).replace('图片加载失败: ', ''));
  if (failed.length) {
    failedImageUrls.value = new Set([...failedImageUrls.value, ...failed]);
  }
  return failed;
};

const preloadCoreImages = async () => {
  const coreImages = [
    luluImages.idle,
    luluImages.happy,
    luluImages.feed,
    luluImages.play,
    luluImages.touch,
    luluImages.bath,
    luluImages.music,
    luluImages.tapBelly,
    luluImages.sleepFloor,
    npcImages.letter,
    npcImages.outing,
    npcImages.resident,
    npcImages.residentSitting,
    npcImages.moveInMemory
  ];
  const results = await Promise.allSettled(coreImages.map(preloadImage));
  const failed = results
    .filter((result) => result.status === 'rejected')
    .map((result) => String(result.reason?.message || result.reason).replace('图片加载失败: ', ''));
  if (failed.length) {
    failedImageUrls.value = new Set([...failedImageUrls.value, ...failed]);
    ElMessage.warning(`有 ${failed.length} 张动作图片未能加载，已自动使用默认图`);
  }
  coreImagesReady.value = true;
};

const handleLuluImageError = (event) => {
  const failedUrl = event?.currentTarget?.getAttribute('src');
  if (!failedUrl || failedImageUrls.value.has(failedUrl)) return;
  failedImageUrls.value = new Set([...failedImageUrls.value, failedUrl]);
  actionLabel.value = '动作图片走丢了，已经换回默认图';
  ElMessage.warning('动作图片加载失败，已自动回退');
};

const getField = (data, upperKey, lowerKey, fallback) => {
  return data?.[upperKey] ?? data?.[lowerKey] ?? fallback;
};

const normalizeNPCEvent = (event) => {
  if (!event) return null;
  return {
    id: Number(getField(event, 'ID', 'id', 0)),
    type: getField(event, 'EVENT_TYPE', 'eventType', ''),
    eventDate: getField(event, 'EVENT_DATE', 'eventDate', ''),
    title: getField(event, 'TITLE', 'title', ''),
    content: getField(event, 'CONTENT', 'content', ''),
    expiresAt: getField(event, 'EXPIRES_AT', 'expiresAt', ''),
    active: Boolean(getField(event, 'ACTIVE', 'active', false)),
    createTime: getField(event, 'CREATE_TIME', 'createTime', '')
  };
};

const normalizeLumeiResidency = (residency) => {
  const rawChapters = getField(residency, 'CHAPTERS', 'chapters', []);
  return {
    status: getField(residency, 'STATUS', 'status', 'NPC'),
    resident: Boolean(getField(residency, 'RESIDENT', 'resident', false)),
    currentLevel: Number(getField(residency, 'CURRENT_LEVEL', 'currentLevel', displayLevel.value)),
    targetLevel: Number(getField(residency, 'TARGET_LEVEL', 'targetLevel', 55)),
    remainingLevels: Number(getField(residency, 'REMAINING_LEVELS', 'remainingLevels', Math.max(0, 55 - displayLevel.value))),
    progress: Number(getField(residency, 'PROGRESS', 'progress', Math.min(100, displayLevel.value * 100 / 55))),
    title: getField(residency, 'TITLE', 'title', '噜妹还住在远方'),
    message: getField(residency, 'MESSAGE', 'message', '她会偶尔寄信，也会悄悄来找噜噜玩。'),
    nextMilestoneLevel: Number(getField(residency, 'NEXT_MILESTONE_LEVEL', 'nextMilestoneLevel', 45)),
    nextMilestoneText: getField(residency, 'NEXT_MILESTONE_TEXT', 'nextMilestoneText', '45 级时，噜妹会说出想留下来的心愿。'),
    residentSince: getField(residency, 'RESIDENT_SINCE', 'residentSince', ''),
    chapters: Array.isArray(rawChapters) ? rawChapters.map((chapter) => ({
      level: Number(getField(chapter, 'LEVEL', 'level', 0)),
      title: getField(chapter, 'TITLE', 'title', ''),
      description: getField(chapter, 'DESCRIPTION', 'description', ''),
      unlocked: Boolean(getField(chapter, 'UNLOCKED', 'unlocked', false)),
      current: Boolean(getField(chapter, 'CURRENT', 'current', false))
    })) : []
  };
};

const normalizeLumeiWeekly = (weekly) => {
  const rawTasks = getField(weekly, 'TASKS', 'tasks', []);
  const tasks = Array.isArray(rawTasks) ? rawTasks.map((task, index) => {
    const current = Number(getField(task, 'CURRENT', 'current', 0));
    const target = Number(getField(task, 'TARGET', 'target', 1)) || 1;
    return {
      id: getField(task, 'ID', 'id', `weekly-${index}`),
      taskType: getField(task, 'TASK_TYPE', 'taskType', ''),
      title: getField(task, 'TITLE', 'title', ''),
      description: getField(task, 'DESCRIPTION', 'description', ''),
      current,
      target,
      remaining: Number(getField(task, 'REMAINING', 'remaining', Math.max(0, target - current))),
      completed: Boolean(getField(task, 'COMPLETED', 'completed', false)),
      reward: getField(task, 'REWARD', 'reward', ''),
      order: Number(getField(task, 'ORDER', 'order', index + 1)),
      progress: Math.max(0, Math.min(100, current * 100 / target))
    };
  }) : [];
  return {
    unlocked: Boolean(getField(weekly, 'UNLOCKED', 'unlocked', false)),
    weekKey: getField(weekly, 'WEEK_KEY', 'weekKey', ''),
    weekStart: getField(weekly, 'WEEK_START', 'weekStart', ''),
    weekEnd: getField(weekly, 'WEEK_END', 'weekEnd', ''),
    completedCount: Number(getField(weekly, 'COMPLETED_COUNT', 'completedCount', 0)),
    totalCount: Number(getField(weekly, 'TOTAL_COUNT', 'totalCount', tasks.length)),
    allCompleted: Boolean(getField(weekly, 'ALL_COMPLETED', 'allCompleted', false)),
    tasks
  };
};

const applyLuluWorld = (result) => {
  const payload = result?.result ?? result;
  if (!payload || payload?.isError) return;

  const memory = getField(payload, 'MEMORY', 'memory', {});
  const goal = getField(payload, 'COMMUNITY_GOAL', 'communityGoal', {});
  const nextEvent = normalizeNPCEvent(getField(payload, 'NPC_EVENT', 'npcEvent', null));
  const history = getField(payload, 'NPC_HISTORY', 'npcHistory', []);
  const residency = normalizeLumeiResidency(getField(payload, 'LUMEI_RESIDENCY', 'lumeiResidency', {}));
  const weekly = normalizeLumeiWeekly(getField(payload, 'LUMEI_WEEKLY', 'lumeiWeekly', {}));

  worldData.value = {
    event: nextEvent,
    history: Array.isArray(history) ? history.map(normalizeNPCEvent).filter(Boolean) : [],
    residency,
    weekly,
    memory: {
      title: getField(memory, 'TITLE', 'title', '今天认识的新朋友'),
      signature: getField(memory, 'SIGNATURE', 'signature', '噜噜正在翻开新的回忆页。'),
      lines: getField(memory, 'LINES', 'lines', [])
    },
    communityGoal: {
      title: getField(goal, 'TITLE', 'title', '大家一起陪陪噜噜'),
      current: Number(getField(goal, 'CURRENT', 'current', 0)),
      target: Number(getField(goal, 'TARGET', 'target', 20)),
      remaining: Number(getField(goal, 'REMAINING', 'remaining', 20)),
      participants: Number(getField(goal, 'PARTICIPANTS', 'participants', 0)),
      completed: Boolean(getField(goal, 'COMPLETED', 'completed', false)),
      reward: getField(goal, 'REWARD', 'reward', '全站解锁一整天的温暖心情')
    }
  };

  if (residency.resident && typeof window !== 'undefined' && window.localStorage.getItem('lumeiResidentIntroSeenV2') !== '1') {
    showLumeiMoveIn.value = true;
  } else if (!residency.resident) {
    showLumeiMoveIn.value = false;
  }

  if (nextEvent?.active && nextEvent.id !== lastAnnouncedNPCEventId) {
    lastAnnouncedNPCEventId = nextEvent.id;
    if (nextEvent.type === 'LETTER') {
      triggerEffect('信', '噜妹来信啦', 'type-mood');
    } else if (nextEvent.type === 'OUTING') {
      triggerEffect('游', '发现外出彩蛋', 'type-level');
    } else if (nextEvent.type === 'LUMEI_OUTING') {
      triggerEffect('出', '噜妹留下了出门便签', 'type-mood');
    }
  }
};

const fetchLuluWorld = async (silent = false) => {
  if (!silent) isWorldLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/world', { userNum: currentUserId });
    applyLuluWorld(result);
  } catch (error) {
    console.error('获取噜噜世界状态失败:', error);
  } finally {
    if (!silent) isWorldLoading.value = false;
  }
};

const openWorldPanel = () => {
  isWorldPanelOpen.value = true;
  fetchLuluWorld(true);
};

const dismissNPCEvent = () => {
  dismissedNPCEventId.value = activeNPCEvent.value?.id ?? null;
};

const lumeiResidentDialogues = [
  '你来啦！我和噜噜刚刚还在说你。',
  '住在这里以后，每天都能等你回来啦。',
  '我把橘子分了一半给噜噜，另一半留给你。',
  '今天也要摸摸噜噜，它其实特别期待。',
  '我的小房间收拾好啦，花边一点也没有弄皱。',
  '外面的云很好看，不过在这里陪你们也很好。',
  '悄悄告诉你：噜噜刚才又打了一个大哈欠。',
  '以后有开心的事，要同时讲给我和噜噜听哦。'
];

const lumeiFeedReactions = [
  '这顿闻起来好香，噜噜要慢慢吃哦。',
  '我来帮它数数吃了几口。',
  '噜噜今天的饭量还是圆滚滚的。',
  '吃完记得擦擦嘴，我可看见啦。',
  '这一口看起来最好吃，留给噜噜！'
];

const lumeiPlayReactions = [
  '加油加油，我来当裁判！',
  '噜噜跑反方向啦，快回来！',
  '下一局也要带上我哦。',
  '我宣布：今天的冠军是开心！',
  '慢一点，我的花边都快笑歪啦。'
];

const showLumeiReaction = (lines, duration = 4200, forcePose = '') => {
  if (!lumeiIsResident.value || isLumeiAway.value) return;
  if (forcePose) lumeiPose.value = forcePose;
  window.clearTimeout(residentLineTimer);
  residentLine.value = pickOne(lines);
  residentLineTimer = window.setTimeout(() => {
    residentLine.value = '';
  }, duration);
};

const interactWithLumei = () => {
  showLumeiReaction(lumeiResidentDialogues);
  registerBondInteraction();
  triggerEffect('妹', '噜妹回应了你', 'type-mood');
};

const acknowledgeLumeiMoveIn = () => {
  if (typeof window !== 'undefined') {
    window.localStorage.setItem('lumeiResidentIntroSeenV2', '1');
  }
  showLumeiMoveIn.value = false;
  showLumeiReaction(['我真的住下来啦，以后请多关照！'], 4600, 'standing');
  triggerEffect('家', '噜妹正式入住', 'type-level');
};

const scheduleLumeiPose = () => {
  window.clearTimeout(lumeiPoseTimer);
  lumeiPoseTimer = window.setTimeout(() => {
    if (lumeiIsResident.value && !isLumeiAway.value && petData.value.currentState !== 'SLEEPING') {
      lumeiPose.value = Math.random() < 0.48 ? 'sitting' : 'standing';
    }
    scheduleLumeiPose();
  }, randomBetween(28000, 52000));
};

const scheduleWorldRefresh = () => {
  window.clearTimeout(worldRefreshTimer);
  worldRefreshTimer = window.setTimeout(() => fetchLuluWorld(true), 700);
};

const setSceneIndex = (value) => {
  const numericValue = Number(value);
  const normalized = Number.isFinite(numericValue)
    ? ((Math.trunc(numericValue) % sceneModes.length) + sceneModes.length) % sceneModes.length
    : 0;
  sceneIndex.value = normalized;
  if (typeof window !== 'undefined') {
    window.localStorage.setItem('luluSceneIndex', String(normalized));
  }
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
    await preloadAccessoryImages(currentAccessory.value);
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

const startSleepStill = () => {
  clearActionTimers();
  finishAction();
  actionCategory.value = 'sleep';
  actionFrames.value = [currentAccessory.value.sleepImage || luluImages.sleepFloor];
  actionLabel.value = '睡得很香';
  triggerEffect('Z', '睡得很香', 'type-sleep');

  actionTimer = window.setTimeout(finishAction, 1800);
};

const maybeStartAmbientAction = () => {
  if (isLoading.value) return;
  if (petData.value.currentState === 'SLEEPING') return;
  if (actionCategory.value !== 'idle') return;

  startFrameAction('ambient', pickOne(ambientActions), 3000, false, 1350);
};

const scheduleAmbientAction = () => {
  window.clearTimeout(ambientTimer);
  ambientTimer = window.setTimeout(() => {
    maybeStartAmbientAction();
    scheduleAmbientAction();
  }, randomBetween(11000, 24000));
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

const normalizePagedResult = (result, normalizer, fallbackPageSize) => {
  if (result?.isError) {
    ElMessage.warning(result.errMsg || result.message || result.msg || '操作失败');
    return null;
  }
  const payload = result?.result ?? result;
  if (Array.isArray(payload)) {
    return {
      items: payload.map(normalizer),
      page: 1,
      pageSize: fallbackPageSize,
      total: payload.length,
      totalPages: 1
    };
  }
  const rawItems = getField(payload, 'ITEMS', 'items', getField(payload, 'DATA', 'data', []));
  const pageSize = Number(getField(payload, 'PAGE_SIZE', 'pageSize', fallbackPageSize)) || fallbackPageSize;
  const total = Number(getField(payload, 'TOTAL', 'total', Array.isArray(rawItems) ? rawItems.length : 0)) || 0;
  return {
    items: Array.isArray(rawItems) ? rawItems.map(normalizer) : [],
    page: Number(getField(payload, 'PAGE', 'page', 1)) || 1,
    pageSize,
    total,
    totalPages: Math.max(1, Number(getField(payload, 'TOTAL_PAGES', 'totalPages', Math.ceil(total / pageSize))) || 1)
  };
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

const fetchLogs = async (page = logPagination.value.page) => {
  isLogLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/logs', {
      userNum: currentUserId,
      page: Math.max(1, Number(page) || 1),
      pageSize: logPagination.value.pageSize
    });
    const pageResult = normalizePagedResult(result, normalizeLog, logPagination.value.pageSize);
    if (pageResult) {
      logs.value = pageResult.items;
      logPagination.value = {
        page: pageResult.page,
        pageSize: pageResult.pageSize,
        total: pageResult.total,
        totalPages: pageResult.totalPages
      };
    }
  } catch (error) {
    console.error('获取 噜噜 日志失败:', error);
    ElMessage.error('日志读取失败');
  } finally {
    isLogLoading.value = false;
  }
};

const openLogPanel = async () => {
  isLogPanelOpen.value = true;
  await fetchLogs(1);
};

const changeLogPage = (page) => {
  if (page < 1 || page > logPagination.value.totalPages || isLogLoading.value) return;
  fetchLogs(page);
};

const getLogIcon = (actionType) => {
  const map = {
    FEED: '食',
    PLAY: '玩',
    AUTO_CARE: '护',
    VISIT: '来',
    SLEEP: '睡',
    WAKE: '醒',
    NPC_LETTER: '信',
    NPC_OUTING: '游',
    NPC_SOLO_OUTING: '出',
    NPC_RESIDENT: '家'
  };
  return map[actionType] || '记';
};

const applyMonthlyCompanionship = (result) => {
  if (!result || result?.isError) return;
  const payload = result?.result ?? result;
  monthlyCompanionship.value = {
    month: getField(payload, 'MONTH', 'month', ''),
    visitedDays: Number(getField(payload, 'VISITED_DAYS', 'visitedDays', 0)),
    missedDays: Number(getField(payload, 'MISSED_DAYS', 'missedDays', 0)),
    elapsedDays: Number(getField(payload, 'ELAPSED_DAYS', 'elapsedDays', new Date().getDate())),
    daysInMonth: Number(getField(payload, 'DAYS_IN_MONTH', 'daysInMonth', 30)),
    visitedDates: getField(payload, 'VISITED_DATES', 'visitedDates', [])
  };
};

const fetchMonthlyCompanionship = async () => {
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/companionship/monthly', { userNum: currentUserId });
    applyMonthlyCompanionship(result);
  } catch (error) {
    console.error('获取月度陪伴统计失败:', error);
  }
};

const recordDailyVisit = async () => {
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/visit', { userNum: currentUserId });
    applyMonthlyCompanionship(result);
  } catch (error) {
    console.error('记录噜噜来访失败:', error);
    fetchMonthlyCompanionship();
  }
};

const changeAccessory = async () => {
  registerBondInteraction();
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/clothes/change', { userNum: currentUserId });
    const nextIndex = Number(getField(result, 'CLOTHES_INDEX', 'clothesIndex', 0)) % accessoryModes.length;
    const failed = await preloadAccessoryImages(accessoryModes[nextIndex]);
    if (failed.length) {
      console.error('换装图片加载失败:', failed);
    }
    applyFunState(result);
    triggerEffect('装', `换成${currentAccessory.value.name}`, 'type-mood');
    const displayImage = petData.value.currentState === 'SLEEPING'
      ? currentAccessory.value.sleepImage
      : currentAccessory.value.image;
    startFrameAction('ambient', { label: `噜噜换上了${currentAccessory.value.name}`, frames: [displayImage] }, 2200, false, 1000);
  } catch (error) {
    console.error('切换 噜噜 衣服失败:', error);
    ElMessage.error('换装失败');
  }
};

const changeScene = async () => {
  const nextIndex = (sceneIndex.value + 1) % sceneModes.length;
  const nextScene = sceneModes[nextIndex];
  if (nextScene.image) {
    try {
      await preloadImage(nextScene.image);
    } catch (error) {
      console.error('场景图片加载失败:', error);
      ElMessage.error('场景图片没有加载成功，请稍后重试');
      return;
    }
  }

  registerBondInteraction();
  setSceneIndex(nextIndex);
  triggerEffect('景', nextIndex === 0 ? '回到原始背景' : `来到${nextScene.name}`, 'type-mood');
};

const makeWish = () => {
  registerBondInteraction();
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

  registerBondInteraction();
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

const fetchStatus = async (silent = false) => {
  const hadStatus = ['ready', 'stale'].includes(statusSyncState.value);
  if (!silent && !hadStatus) {
    statusSyncState.value = 'loading';
  }
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/status', { userNum: currentUserId });
    updatePetData(result);
    statusSyncState.value = 'ready';
  } catch (error) {
    console.error('获取 噜噜 状态失败:', error);
    statusSyncState.value = hadStatus ? 'stale' : 'error';
  }
};

const fetchMessages = async (page = messagePagination.value.page) => {
  isMessageLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/messages', {
      userNum: currentUserId,
      page: Math.max(1, Number(page) || 1),
      pageSize: messagePagination.value.pageSize
    });
    const pageResult = normalizePagedResult(result, normalizeMessage, messagePagination.value.pageSize);
    if (pageResult) {
      messages.value = pageResult.items;
      messagePagination.value = {
        page: pageResult.page,
        pageSize: pageResult.pageSize,
        total: pageResult.total,
        totalPages: pageResult.totalPages
      };
      await scrollMessagesToTop();
    }
  } catch (error) {
    console.error('获取 噜噜 留言失败:', error);
    ElMessage.error('留言读取失败');
  } finally {
    isMessageLoading.value = false;
  }
};

const changeMessagePage = (page) => {
  if (page < 1 || page > messagePagination.value.totalPages || isMessageLoading.value) return;
  fetchMessages(page);
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
    if (result?.isError) {
      ElMessage.warning(result.errMsg || '留言发送失败');
      return;
    }
    messageInput.value = '';
    triggerEffect('信', '留言已飘出去', 'type-mood');
    await fetchMessages(1);
    scheduleWorldRefresh();
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
    if (result?.isError) {
      ElMessage.warning(result.errMsg || '删除失败');
      return;
    }
    const targetPage = messages.value.length === 1 && messagePagination.value.page > 1
      ? messagePagination.value.page - 1
      : messagePagination.value.page;
    await fetchMessages(targetPage);
    scheduleWorldRefresh();
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

const showBlockedReaction = (reason) => {
  const reactions = {
    sleeping: {
      label: '噜噜翻了个身，继续睡觉',
      frames: [currentAccessory.value.sleepImage || luluImages.sleepFloor],
      effect: ['正在做美梦', 'type-sleep'],
      icon: 'Z'
    },
    hungry: {
      label: '噜噜捂着咕咕叫的小肚子',
      frames: [luluImages.tapBelly],
      effect: ['先喂我嘛', 'type-food'],
      icon: '食'
    },
    tired: {
      label: '噜噜累得坐不住啦',
      frames: [currentAccessory.value.sleepImage || luluImages.sleepFloor],
      effect: ['需要休息', 'type-sleep'],
      icon: 'Z'
    }
  };
  const reaction = reactions[reason];
  if (!reaction) return;
  startFrameAction('touch', reaction, 2400, false, 900);
  triggerEffect(reaction.icon, reaction.effect[0], reaction.effect[1]);
};

const blockUnavailableActivity = () => {
  if (petData.value.currentState === 'SLEEPING') {
    showBlockedReaction('sleeping');
    return true;
  }
  if (petData.value.hunger < 15) {
    showBlockedReaction('hungry');
    return true;
  }
  if (petData.value.energy < 15) {
    showBlockedReaction('tired');
    return true;
  }
  return false;
};

const feedLulu = async () => {
  if (isLoading.value) return;
  registerBondInteraction();

  if (petData.value.hunger >= 90) {
    startFrameAction('touch', { label: '噜噜拍拍圆滚滚的小肚子', frames: [luluImages.happy] }, 2200, false, 900);
    triggerEffect('饱', '已经吃饱啦', 'type-food');
    if (Math.random() < 0.45) {
      showLumeiReaction(['它的小肚子已经装不下啦。', '先让噜噜消化一会儿吧。']);
    }
    return;
  }

  startFrameAction('feed', pickOne(feedActions), 4200);
  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/feed', { userNum: currentUserId });
    updatePetData(result);
    triggerEffect('XP', '经验 +20', 'type-level');
    if (Math.random() < 0.45) {
      showLumeiReaction(lumeiFeedReactions);
    }
    advanceMission('feed');
    fetchMonthlyCompanionship();
    scheduleWorldRefresh();
  } catch (error) {
    console.error('喂食失败:', error);
    ElMessage.error('喂食没有保存，请稍后重试');
  } finally {
    isLoading.value = false;
  }
};

const playLulu = async () => {
  if (isLoading.value) return;
  registerBondInteraction();
  if (blockUnavailableActivity()) return;

  startFrameAction('play', pickOne(playActions), 4200);
  isLoading.value = true;

  try {
    const result = await sendAxiosRequest('/blog-api/lulu/play', { userNum: currentUserId, actionName: '玩耍' });
    updatePetData(result);
    triggerEffect('XP', '经验 +40', 'type-level');
    showLumeiReaction(lumeiPlayReactions);
    advanceMission('play');
    fetchMonthlyCompanionship();
    scheduleWorldRefresh();
  } catch (error) {
    console.error('玩耍失败:', error);
    ElMessage.error('玩耍状态没有保存，请稍后重试');
  } finally {
    isLoading.value = false;
  }
};

const runPlayLikeAction = async (category, actionList, expText) => {
  if (isLoading.value) return;
  registerBondInteraction();
  if (blockUnavailableActivity()) return;

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
    fetchMonthlyCompanionship();
    scheduleWorldRefresh();
  } catch (error) {
    console.error(`${category} 互动失败:`, error);
    ElMessage.error('互动状态没有保存，请稍后重试');
  } finally {
    isLoading.value = false;
  }
};

const touchLulu = () => runPlayLikeAction('touch', touchActions, '经验 +40');

const bathLulu = () => runPlayLikeAction('bath', bathActions, '经验 +40');

const musicLulu = () => runPlayLikeAction('music', musicActions, '经验 +40');

const playSmartCareAnimation = (action) => {
  const animationMap = {
    FEED: () => startFrameAction('feed', pickOne(feedActions), 4200),
    REST: () => startFrameAction('touch', { label: '智能照顾正在安排休息', frames: [currentAccessory.value.sleepImage || luluImages.sleepFloor] }, 3600, false, 900),
    RESTING: () => startFrameAction('touch', { label: '安静守护噜噜的美梦', frames: [currentAccessory.value.sleepImage || luluImages.sleepFloor] }, 3000, false, 900),
    COMFORT: () => startFrameAction('touch', pickOne(touchActions), 3800, true, 900),
    STROLL: () => startFrameAction('play', { label: '和噜噜一起散步', frames: [luluImages.play, luluImages.playChase] }, 4200, true, 720)
  };
  (animationMap[action] || animationMap.COMFORT)();
};

const smartCareLulu = async () => {
  if (isLoading.value) return;

  registerBondInteraction();
  playSmartCareAnimation(smartCareAdvice.value.action);
  isLoading.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/care', { userNum: currentUserId });
    const careAction = getField(result, 'CARE_ACTION', 'careAction', smartCareAdvice.value.action);
    const careMessage = getField(result, 'CARE_MESSAGE', 'careMessage', '噜噜感受到了你的照顾。');
    const missionType = getField(result, 'CARE_MISSION_TYPE', 'careMissionType', '');
    const expGain = Number(getField(result, 'CARE_EXP_GAIN', 'careExpGain', 0));
    const actionAccepted = Boolean(getField(result, 'ACTION_ACCEPTED', 'actionAccepted', true));

    updatePetData(result);
    playSmartCareAnimation(careAction);
    triggerEffect('护', careMessage, careAction === 'REST' || careAction === 'RESTING' ? 'type-sleep' : 'type-level');
    if (actionAccepted && expGain > 0) {
      triggerEffect('XP', `经验 +${expGain}`, 'type-level');
    }
    if (actionAccepted && missionType) {
      advanceMission(missionType);
    }
    fetchMonthlyCompanionship();
    scheduleWorldRefresh();
  } catch (error) {
    console.error('智能照顾失败:', error);
    ElMessage.error('智能照顾暂时不可用，请稍后重试');
  } finally {
    isLoading.value = false;
  }
};

const sleepLulu = async () => {
  if (isLoading.value) return;
  registerBondInteraction();

  const isWaking = petData.value.currentState === 'SLEEPING';

  if (!isWaking) {
    startSleepStill();
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
      showLumeiReaction(['噜噜醒啦，刚才做了一个很香的梦。', '早呀噜噜，睡饱以后再一起玩吧。'], 4200, 'standing');
    } else {
      showLumeiReaction(['嘘，噜噜刚刚睡着，我们小声一点。', '让它好好睡一会儿，我会在旁边看着。'], 5200, 'sitting');
    }
    fetchMonthlyCompanionship();
    scheduleWorldRefresh();
  } catch (error) {
    console.error('切换睡眠状态失败:', error);
    ElMessage.error('睡眠状态没有保存，请稍后重试');
  } finally {
    isLoading.value = false;
  }
};

const initializeLuluWorld = async () => {
  await recordDailyVisit();
  await fetchLuluWorld();
};

onMounted(() => {
  preloadCoreImages();
  fetchFunState();
  fetchStatus();
  fetchMessages();
  initializeLuluWorld();
  scheduleThought();
  scheduleAmbientAction();
  scheduleLumeiPose();
  pollerTimer = window.setInterval(() => fetchStatus(true), 10000);
});

onBeforeUnmount(() => {
  clearActionTimers();
  window.clearTimeout(ambientTimer);
  window.clearTimeout(thoughtTimer);
  window.clearTimeout(thoughtClearTimer);
  window.clearTimeout(comboTimer);
  window.clearTimeout(worldRefreshTimer);
  window.clearTimeout(residentLineTimer);
  window.clearTimeout(lumeiPoseTimer);
  window.cancelAnimationFrame(motionFrame);
  window.clearInterval(pollerTimer);
});
</script>

<style scoped>
.lulu-viewport {
  width: 100%;
  height: 100vh;
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  overflow: hidden;
  background: #f5f6fa;
  color: #303744;
}

.lulu-stage {
  position: relative;
  height: 100vh;
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
  background-position: center;
  background-repeat: no-repeat;
  background-size: cover;
  transition: background 1.2s ease;
}

.scene-background.has-scene-image {
  animation: sceneReveal 0.42s ease-out both;
}

.scene-background.has-scene-image::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.09), transparent 42%, rgba(244, 248, 255, 0.08));
  pointer-events: none;
}

.scene-background.has-scene-image ~ .top-status .identity-block {
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.76);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.7);
  box-shadow: 0 12px 30px rgba(35, 45, 62, 0.1);
  backdrop-filter: blur(14px);
}

.scene-morning .scene-background:not(.has-scene-image) {
  background:
    radial-gradient(circle at 45% 42%, rgba(255, 211, 105, 0.45), transparent 31%),
    linear-gradient(135deg, #fff7e8 0%, #ebf7ff 54%, #fff1e5 100%);
}

.scene-day .scene-background:not(.has-scene-image) {
  background:
    radial-gradient(circle at 48% 44%, rgba(255, 220, 118, 0.36), transparent 30%),
    linear-gradient(135deg, #f4fbff 0%, #eaf8ff 52%, #fff8eb 100%);
}

.scene-evening .scene-background:not(.has-scene-image) {
  background:
    radial-gradient(circle at 48% 44%, rgba(255, 172, 105, 0.34), transparent 31%),
    linear-gradient(135deg, #fff1e6 0%, #eeeafa 54%, #ffecef 100%);
}

.scene-night .scene-background:not(.has-scene-image) {
  background:
    radial-gradient(circle at 48% 44%, rgba(136, 154, 255, 0.22), transparent 30%),
    linear-gradient(135deg, #e8eafa 0%, #dfeaf7 52%, #f1e7f6 100%);
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

.sync-notice {
  margin-top: 7px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #7b8798;
  font-size: 11px;
  font-weight: 800;
}

.sync-notice::before {
  content: '';
  width: 7px;
  height: 7px;
  flex: none;
  border-radius: 50%;
  background: #54a0ff;
  box-shadow: 0 0 0 4px rgba(84, 160, 255, 0.12);
}

.sync-notice.sync-error,
.sync-notice.sync-stale {
  color: #b56a43;
}

.sync-notice.sync-error::before,
.sync-notice.sync-stale::before {
  background: #ff9f43;
  box-shadow: 0 0 0 4px rgba(255, 159, 67, 0.14);
}

.sync-notice button {
  border: none;
  padding: 2px 8px;
  border-radius: 999px;
  color: #a4562e;
  background: rgba(255, 159, 67, 0.14);
  font: inherit;
  cursor: pointer;
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
.streak-card,
.residency-card,
.community-card {
  padding: 13px 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(255, 255, 255, 0.82);
  box-shadow: 0 14px 34px rgba(40, 49, 66, 0.08);
  backdrop-filter: blur(14px);
}

.daily-card span,
.streak-card span,
.residency-card span,
.community-card span {
  display: block;
  color: #7d8796;
  font-size: 12px;
  font-weight: 900;
}

.daily-card strong,
.streak-card strong,
.residency-card strong,
.community-card strong {
  display: block;
  margin-top: 4px;
  color: #303744;
  font-size: 15px;
  line-height: 1.35;
}

.daily-card small,
.streak-card small,
.residency-card small,
.community-card small {
  display: block;
  margin-top: 5px;
  color: #8a94a6;
  font-size: 12px;
  font-weight: 800;
}

.mission-progress,
.companion-progress,
.residency-progress,
.community-progress {
  width: 100%;
  height: 7px;
  margin-top: 9px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(48, 55, 68, 0.08);
}

.mission-fill,
.companion-fill {
  height: 100%;
  border-radius: inherit;
  transition: width 0.28s ease-out;
}

.mission-fill {
  background: linear-gradient(90deg, #ff9f43, #ff6b81);
}

.companion-fill {
  background: linear-gradient(90deg, #6c8cff, #b76cff);
}

.residency-card,
.community-card {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.86);
  text-align: left;
  font: inherit;
  cursor: pointer;
  pointer-events: auto;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.residency-card:hover,
.community-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 38px rgba(92, 71, 128, 0.13);
}

.residency-card {
  background:
    radial-gradient(circle at 92% 12%, rgba(255, 215, 121, 0.33), transparent 35%),
    rgba(255, 249, 252, 0.78);
}

.residency-card.status-resident {
  border-color: rgba(255, 202, 222, 0.92);
  box-shadow: 0 15px 36px rgba(226, 101, 151, 0.13);
}

.residency-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #ffbd64, #ed75a5);
  transition: width 0.4s ease;
}

.status-resident .residency-fill {
  background: linear-gradient(90deg, #ef7ba4, #9c7ae8, #65b6e8);
}

.community-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #ff9d68, #ef6f9c);
  transition: width 0.35s ease;
}

.community-fill.completed {
  background: linear-gradient(90deg, #5bc990, #65a8ff);
}

.npc-letter-popover {
  position: absolute;
  left: calc(50% + 145px);
  top: clamp(155px, 20vh, 205px);
  z-index: 7;
  width: min(380px, 32vw);
  display: grid;
  grid-template-columns: 128px minmax(0, 1fr);
  align-items: end;
  gap: 2px;
  pointer-events: none;
  animation: letterArrive 0.5s cubic-bezier(0.2, 0.82, 0.2, 1) both;
}

.npc-letter-popover > img {
  width: 148px;
  max-height: 245px;
  object-fit: contain;
  align-self: end;
  filter: drop-shadow(0 18px 22px rgba(87, 58, 23, 0.2));
}

.npc-letter-copy {
  position: relative;
  margin-bottom: 34px;
  padding: 13px 14px;
  border: 1px solid rgba(255, 255, 255, 0.88);
  border-radius: 18px 18px 18px 6px;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 16px 38px rgba(56, 45, 75, 0.14);
  backdrop-filter: blur(16px);
  pointer-events: auto;
}

.npc-letter-copy > span {
  display: block;
  color: #dd6d94;
  font-size: 11px;
  font-weight: 900;
}

.npc-letter-copy > strong {
  display: block;
  margin-top: 2px;
  color: #353947;
  font-size: 15px;
  font-weight: 900;
}

.npc-letter-copy p {
  margin: 6px 0 9px;
  color: #697386;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.55;
}

.npc-letter-copy div {
  display: flex;
  gap: 7px;
}

.npc-letter-copy button {
  padding: 6px 9px;
  border: none;
  border-radius: 999px;
  color: #fff;
  background: #ef7ba4;
  font-size: 11px;
  font-weight: 900;
  cursor: pointer;
}

.npc-letter-copy .letter-dismiss {
  color: #7b7181;
  background: rgba(115, 103, 128, 0.1);
}

.lumei-away-note {
  position: absolute;
  left: calc(50% + 165px);
  top: clamp(155px, 20vh, 205px);
  z-index: 7;
  width: min(300px, 28vw);
  padding: 15px 16px;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 20px 20px 6px 20px;
  color: #5c6372;
  background:
    linear-gradient(150deg, rgba(255, 252, 239, 0.95), rgba(255, 238, 247, 0.94));
  box-shadow: 0 17px 38px rgba(64, 49, 71, 0.15);
  backdrop-filter: blur(15px);
  animation: letterArrive 0.4s ease both;
}

.lumei-away-note > span {
  color: #da6e96;
  font-size: 11px;
  font-weight: 900;
}

.lumei-away-note > strong {
  display: block;
  margin-top: 3px;
  color: #373c49;
  font-size: 15px;
  font-weight: 900;
}

.lumei-away-note > p {
  margin: 7px 0 10px;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.55;
}

.lumei-away-note > button {
  padding: 7px 10px;
  border: 0;
  border-radius: 999px;
  color: #fff;
  background: #e9799f;
  font-size: 11px;
  font-weight: 900;
  cursor: pointer;
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
  transform: translate3d(
    var(--lulu-shift-x, 0),
    calc(var(--lulu-shift-y, 0px) + var(--scene-lulu-offset-y, 0px)),
    0
  ) rotate(var(--lulu-rotate, 0));
  transform-origin: 50% 78%;
  transition: transform 0.28s cubic-bezier(0.2, 0.8, 0.2, 1);
  will-change: transform;
}

.lulu-stage.has-custom-scene {
  --scene-lulu-offset-y: clamp(22px, 3vh, 34px);
}

.lulu-stage.has-custom-scene .lulu-entity::before {
  bottom: 2%;
  width: 60%;
  height: 10%;
  background: rgba(55, 49, 35, 0.24);
  filter: blur(10px);
}

.lulu-entity.is-away-event {
  width: min(58vw, 620px);
}

.lulu-entity.is-away-event::before {
  bottom: 4%;
  width: 62%;
  background: rgba(55, 49, 35, 0.18);
}

.lulu-entity.is-away-event .body-hotspots {
  display: none;
}

.lulu-entity::before {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 7%;
  z-index: 0;
  width: 54%;
  height: 9%;
  border-radius: 50%;
  background: rgba(55, 49, 35, 0.16);
  filter: blur(13px);
  transform: translateX(-50%);
  pointer-events: none;
}

.bond-badge {
  position: absolute;
  top: 9%;
  right: 5%;
  z-index: 11;
  padding: 7px 11px;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 999px;
  color: #8654c7;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 12px 28px rgba(76, 55, 114, 0.14);
  font-size: 12px;
  font-weight: 900;
  pointer-events: none;
  animation: badgePop 0.28s ease-out;
}

.lulu-img {
  position: relative;
  z-index: 2;
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter: drop-shadow(0 28px 38px rgba(93, 63, 15, 0.22));
  transition: opacity 0.18s ease;
}

.lulu-img.anim-away {
  animation: awayBreeze 2.6s ease-in-out infinite;
}

.lumei-resident {
  position: absolute;
  right: -38%;
  bottom: 2%;
  z-index: 10;
  width: 60%;
  height: 72%;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 0;
  border: 0;
  background: transparent;
  font: inherit;
  cursor: pointer;
  transform-origin: center bottom;
  animation: lumeiBreathe 3.1s ease-in-out infinite;
}

.lumei-resident::before {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 2%;
  z-index: -1;
  width: 54%;
  height: 8%;
  border-radius: 50%;
  background: rgba(55, 49, 35, 0.14);
  filter: blur(8px);
  transform: translateX(-50%);
}

.lumei-resident > img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter: drop-shadow(0 20px 25px rgba(93, 63, 15, 0.2));
  transition: transform 0.25s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.lumei-resident:hover > img,
.lumei-resident.is-talking > img {
  transform: translateY(-4px) rotate(1.5deg) scale(1.025);
}

.lumei-resident.pose-sitting {
  right: -41%;
  bottom: -1%;
  width: 64%;
  height: 70%;
}

.lumei-resident > small {
  position: absolute;
  right: 13%;
  bottom: 5%;
  padding: 4px 8px;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 999px;
  color: #c45f89;
  background: rgba(255, 255, 255, 0.84);
  box-shadow: 0 8px 18px rgba(75, 53, 87, 0.12);
  font-size: 10px;
  font-weight: 900;
}

.lumei-speech {
  position: absolute;
  right: -8%;
  top: -9%;
  z-index: 3;
  width: max-content;
  max-width: 230px;
  padding: 9px 12px;
  border: 1px solid rgba(255, 255, 255, 0.94);
  border-radius: 16px 16px 5px 16px;
  color: #575d6b;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 13px 30px rgba(72, 54, 81, 0.15);
  font-size: 12px;
  font-weight: 800;
  line-height: 1.45;
  text-align: left;
  animation: badgePop 0.25s ease-out;
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

.care-advice {
  margin: 0;
  color: #718096;
  font-size: 11px;
  line-height: 1.45;
  font-weight: 700;
}

.care-advice strong {
  display: block;
  margin-bottom: 2px;
  color: #416d9e;
  font-size: 11px;
  font-weight: 900;
}

.smart-care-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border: none;
  border-radius: 12px;
  padding: 9px 11px;
  color: #563f79;
  background: linear-gradient(145deg, rgba(154, 118, 255, 0.2), rgba(255, 205, 92, 0.26));
  box-shadow: inset 0 0 0 1px rgba(121, 88, 193, 0.08);
  font-size: 12px;
  font-weight: 900;
  cursor: pointer;
  transition: transform 0.18s ease, opacity 0.18s ease;
}

.smart-care-btn small {
  overflow: hidden;
  color: rgba(86, 63, 121, 0.68);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 10px;
}

.smart-care-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.smart-care-btn:disabled {
  cursor: not-allowed;
  opacity: 0.58;
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
  grid-template-columns: repeat(9, minmax(0, 1fr));
  gap: 8px;
}

.action-btn {
  min-width: 0;
  min-height: 70px;
  padding: 12px 6px;
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

.scene-btn {
  background: linear-gradient(145deg, rgba(93, 173, 226, 0.16), rgba(120, 224, 143, 0.16));
}

.message-board {
  position: relative;
  z-index: 8;
  height: 100vh;
  min-height: 0;
  padding: 24px 22px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
  box-sizing: border-box;
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

.pager {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 34px;
  color: #7b8798;
  font-size: 12px;
  font-weight: 800;
}

.pager button {
  flex: none;
  border: none;
  border-radius: 999px;
  padding: 7px 11px;
  color: #416d9e;
  background: rgba(84, 160, 255, 0.12);
  font: inherit;
  cursor: pointer;
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.pager button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.pager button:disabled {
  cursor: not-allowed;
  opacity: 0.42;
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

.world-overlay {
  position: fixed;
  inset: 0;
  z-index: 42;
  display: flex;
  justify-content: flex-end;
  background: rgba(38, 45, 58, 0.3);
  backdrop-filter: blur(7px);
}

.world-panel {
  width: min(580px, 100%);
  height: 100%;
  overflow-y: auto;
  padding: 24px;
  box-sizing: border-box;
  background:
    radial-gradient(circle at 88% 6%, rgba(255, 188, 210, 0.35), transparent 28%),
    linear-gradient(180deg, rgba(255, 253, 251, 0.98), rgba(248, 247, 255, 0.97));
  box-shadow: -22px 0 50px rgba(31, 40, 56, 0.18);
}

.world-panel-header,
.world-goal-heading,
.npc-history-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.world-panel-header h3 {
  margin: 0;
  color: #2f3440;
  font-size: 22px;
  font-weight: 900;
}

.memory-card,
.residency-world-card,
.world-goal-card,
.npc-history-card {
  margin-top: 16px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 22px;
  box-shadow: 0 16px 38px rgba(47, 54, 72, 0.08);
}

.residency-world-card {
  display: grid;
  grid-template-columns: 132px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
  overflow: hidden;
  background:
    radial-gradient(circle at 14% 20%, rgba(255, 210, 119, 0.28), transparent 34%),
    linear-gradient(145deg, rgba(255, 239, 246, 0.96), rgba(245, 239, 255, 0.92));
}

.residency-world-card.status-resident {
  background:
    radial-gradient(circle at 14% 20%, rgba(255, 210, 119, 0.35), transparent 34%),
    linear-gradient(145deg, rgba(255, 232, 242, 0.98), rgba(238, 244, 255, 0.95));
}

.residency-world-card > img {
  width: 142px;
  height: 142px;
  margin: -8px 0 -18px -10px;
  object-fit: contain;
  filter: drop-shadow(0 15px 20px rgba(86, 57, 24, 0.16));
}

.residency-world-copy > span {
  color: #d66d96;
  font-size: 11px;
  font-weight: 900;
}

.residency-world-copy > strong {
  display: block;
  margin-top: 3px;
  color: #353947;
  font-size: 17px;
  font-weight: 900;
}

.residency-world-copy > p {
  margin: 7px 0 0;
  color: #687286;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.5;
}

.residency-progress.large {
  height: 8px;
  margin-top: 10px;
}

.residency-world-copy > small {
  display: block;
  margin-top: 7px;
  color: #8b7e91;
  font-size: 11px;
  font-weight: 800;
}

.residency-story {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
  margin-top: 5px;
}

.story-chapter {
  position: relative;
  min-width: 0;
  padding: 10px;
  border-radius: 14px;
  opacity: 0.55;
  background: rgba(255, 255, 255, 0.65);
}

.story-chapter.unlocked {
  opacity: 1;
}

.story-chapter.current {
  outline: 2px solid rgba(235, 116, 160, 0.4);
  background: rgba(255, 255, 255, 0.9);
}

.story-chapter > span {
  color: #d46e96;
  font-size: 10px;
  font-weight: 900;
}

.story-chapter > strong {
  display: block;
  margin-top: 3px;
  color: #454a58;
  font-size: 11px;
  font-weight: 900;
  line-height: 1.35;
}

.story-chapter > small {
  display: block;
  margin-top: 5px;
  color: #818999;
  font-size: 9px;
  line-height: 1.45;
}

.move-in-memory-card {
  position: relative;
  margin-top: 16px;
  min-height: 230px;
  overflow: hidden;
  border-radius: 22px;
  box-shadow: 0 18px 42px rgba(61, 46, 49, 0.14);
}

.move-in-memory-card > img {
  display: block;
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
}

.move-in-memory-card > div {
  position: absolute;
  inset: auto 0 0;
  padding: 38px 18px 16px;
  color: #fff;
  background: linear-gradient(transparent, rgba(64, 47, 44, 0.84));
}

.move-in-memory-card span {
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.1em;
}

.move-in-memory-card strong {
  display: block;
  margin-top: 3px;
  font-size: 17px;
  font-weight: 900;
}

.move-in-memory-card p {
  margin: 4px 0 0;
  font-size: 11px;
  font-weight: 700;
}

.weekly-task-card {
  margin-top: 16px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.92);
  border-radius: 22px;
  background: linear-gradient(145deg, rgba(235, 249, 244, 0.96), rgba(239, 242, 255, 0.94));
  box-shadow: 0 16px 38px rgba(47, 54, 72, 0.08);
}

.weekly-task-heading,
.weekly-task-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.weekly-task-heading span {
  color: #5e9581;
  font-size: 11px;
  font-weight: 900;
}

.weekly-task-heading strong {
  display: block;
  margin-top: 3px;
  color: #343b48;
  font-size: 16px;
  font-weight: 900;
}

.weekly-task-heading > b {
  flex: none;
  color: #6686d7;
  font-size: 20px;
  font-weight: 900;
}

.weekly-task-card > small {
  display: block;
  margin-top: 7px;
  color: #84909c;
  font-size: 10px;
  font-weight: 800;
}

.weekly-task-list {
  display: grid;
  gap: 9px;
  margin-top: 13px;
}

.weekly-task-item {
  padding: 11px;
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.72);
}

.weekly-task-item.completed {
  background: rgba(239, 255, 246, 0.88);
}

.weekly-task-title > span {
  flex: none;
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  color: #fff;
  background: linear-gradient(145deg, #7aa8ed, #a17ee8);
  font-size: 11px;
  font-weight: 900;
}

.weekly-task-item.completed .weekly-task-title > span {
  background: linear-gradient(145deg, #55bd87, #7bd3a0);
}

.weekly-task-title > div {
  min-width: 0;
  flex: 1;
}

.weekly-task-title strong {
  color: #424958;
  font-size: 13px;
  font-weight: 900;
}

.weekly-task-title p {
  margin: 2px 0 0;
  color: #747f8e;
  font-size: 11px;
  line-height: 1.4;
}

.weekly-task-title > b {
  flex: none;
  color: #68778d;
  font-size: 12px;
}

.weekly-task-progress {
  height: 6px;
  margin-top: 9px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(72, 85, 102, 0.09);
}

.weekly-task-progress > div {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #63c792, #76a8ee);
  transition: width 0.35s ease;
}

.weekly-task-item > small {
  display: block;
  margin-top: 6px;
  color: #7e8795;
  font-size: 10px;
  font-weight: 800;
}

.memory-card {
  background: linear-gradient(145deg, rgba(255, 239, 226, 0.92), rgba(255, 245, 250, 0.9));
}

.memory-heading span,
.world-goal-heading span,
.npc-history-heading span {
  display: block;
  color: #8e8192;
  font-size: 11px;
  font-weight: 900;
}

.memory-heading strong,
.world-goal-heading strong,
.npc-history-heading strong {
  display: block;
  margin-top: 3px;
  color: #333846;
  font-size: 16px;
  font-weight: 900;
}

.memory-signature {
  margin: 12px 0 8px;
  color: #dc6f91;
  font-size: 14px;
  font-weight: 900;
}

.memory-card ul {
  display: grid;
  gap: 7px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.memory-card li {
  padding-left: 18px;
  color: #626d7f;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.55;
}

.memory-card li::before {
  content: '·';
  float: left;
  margin-left: -14px;
  color: #ef779e;
  font-weight: 900;
}

.world-goal-card {
  background: linear-gradient(145deg, rgba(231, 245, 255, 0.94), rgba(240, 235, 255, 0.92));
}

.world-goal-heading b {
  flex: none;
  color: #6b70cb;
  font-size: 20px;
  font-weight: 900;
}

.community-progress.large {
  height: 9px;
  margin-top: 14px;
}

.world-goal-card > p {
  margin: 9px 0 0;
  color: #6e7890;
  font-size: 12px;
  font-weight: 800;
}

.npc-history-card {
  background: rgba(255, 255, 255, 0.82);
}

.npc-history-heading {
  justify-content: flex-start;
}

.npc-history-heading img {
  width: 62px;
  height: 62px;
  object-fit: contain;
  object-position: top;
}

.npc-history-list {
  display: grid;
  gap: 10px;
  margin-top: 14px;
}

.npc-history-item {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr);
  gap: 10px;
  padding: 11px;
  border-radius: 15px;
  background: rgba(245, 242, 250, 0.76);
}

.npc-history-item > span {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  color: #fff;
  background: linear-gradient(145deg, #f58aad, #ffac75);
  font-size: 12px;
  font-weight: 900;
}

.npc-history-item strong {
  color: #424858;
  font-size: 13px;
  font-weight: 900;
}

.npc-history-item p {
  margin: 3px 0;
  color: #6c7586;
  font-size: 12px;
  line-height: 1.5;
}

.npc-history-item time {
  color: #9aa2af;
  font-size: 10px;
  font-weight: 800;
}

.empty-message.compact {
  padding: 24px 8px 10px;
}

.move-in-overlay {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  padding: 18px;
  box-sizing: border-box;
  background: rgba(43, 43, 58, 0.38);
  backdrop-filter: blur(10px);
  animation: moveInReveal 0.3s ease both;
}

.move-in-card {
  position: relative;
  width: min(430px, 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 30px 28px;
  box-sizing: border-box;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 30px;
  text-align: center;
  background:
    radial-gradient(circle at 50% 16%, rgba(255, 213, 112, 0.35), transparent 31%),
    linear-gradient(160deg, rgba(255, 255, 255, 0.98), rgba(255, 237, 247, 0.98));
  box-shadow: 0 34px 80px rgba(55, 42, 68, 0.26);
  animation: moveInCard 0.48s cubic-bezier(0.2, 0.85, 0.25, 1.1) both;
}

.move-in-card > img {
  width: min(260px, 72vw);
  height: 250px;
  margin: -16px 0 -24px;
  object-fit: contain;
  filter: drop-shadow(0 20px 24px rgba(95, 60, 20, 0.18));
}

.move-in-card > .move-in-memory-image {
  width: 100%;
  height: auto;
  aspect-ratio: 16 / 9;
  margin: 0 0 14px;
  border-radius: 20px;
  object-fit: cover;
  filter: none;
  box-shadow: 0 16px 30px rgba(91, 62, 41, 0.16);
}

.move-in-card > p {
  margin: 0;
  color: #d76d95;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.08em;
}

.move-in-card > h3 {
  margin: 6px 0 9px;
  color: #333743;
  font-size: 27px;
  font-weight: 900;
}

.move-in-card > strong {
  color: #62697a;
  font-size: 14px;
  line-height: 1.65;
}

.move-in-card > small {
  margin-top: 9px;
  color: #8b91a0;
  font-size: 12px;
  line-height: 1.55;
}

.move-in-card > button {
  margin-top: 18px;
  min-width: 150px;
  padding: 11px 20px;
  border: 0;
  border-radius: 999px;
  color: #fff;
  background: linear-gradient(135deg, #ed7aa4, #9c79e8);
  box-shadow: 0 12px 24px rgba(190, 93, 145, 0.25);
  font-size: 14px;
  font-weight: 900;
  cursor: pointer;
}

.move-in-sparkles {
  position: absolute;
  inset: 40px 34px auto;
  display: flex;
  justify-content: space-between;
  color: #f09bb7;
  font-size: 20px;
  pointer-events: none;
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

.log-panel-header h3 small {
  color: #8a94a6;
  font-size: 12px;
  font-weight: 800;
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

@keyframes badgePop {
  0% {
    opacity: 0;
    transform: translateY(5px) scale(0.85);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes sceneReveal {
  0% {
    opacity: 0.45;
    transform: scale(1.012);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes letterArrive {
  0% {
    opacity: 0;
    transform: translate3d(28px, 12px, 0) scale(0.96);
  }
  100% {
    opacity: 1;
    transform: translate3d(0, 0, 0) scale(1);
  }
}

@keyframes awayBreeze {
  0%, 100% {
    transform: translate3d(-4px, 0, 0) rotate(-0.6deg);
  }
  50% {
    transform: translate3d(5px, -4px, 0) rotate(0.8deg);
  }
}

@keyframes lumeiBreathe {
  0%, 100% {
    transform: translateY(0) rotate(-0.4deg);
  }
  50% {
    transform: translateY(-4px) rotate(0.7deg);
  }
}

@keyframes moveInReveal {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes moveInCard {
  from {
    opacity: 0;
    transform: translateY(22px) scale(0.94);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .lulu-entity {
    transform: none !important;
    transition: none;
  }

  .bond-badge {
    animation: none;
  }

  .scene-background.has-scene-image {
    animation: none;
  }

  .npc-letter-popover,
  .lumei-away-note,
  .lumei-resident,
  .move-in-overlay,
  .move-in-card,
  .lulu-img.anim-away {
    animation: none;
  }
}

@media (max-width: 1180px) {
  .lulu-viewport {
    height: auto;
    grid-template-columns: 1fr;
    overflow: auto;
  }

  .lulu-stage {
    height: auto;
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

  .community-card,
  .residency-card {
    grid-column: auto;
  }

  .message-board {
    height: auto;
    min-height: 420px;
    overflow: visible;
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

  .sync-notice {
    margin-top: 4px;
    overflow: hidden;
    white-space: nowrap;
  }

  .sync-notice span {
    overflow: hidden;
    text-overflow: ellipsis;
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

  .npc-letter-popover {
    left: 12px;
    right: 12px;
    top: 36%;
    width: auto;
    grid-template-columns: 86px minmax(0, 1fr);
  }

  .npc-letter-popover > img {
    width: 100px;
    max-height: 170px;
  }

  .npc-letter-copy {
    margin-bottom: 12px;
    padding: 10px 11px;
  }

  .lumei-away-note {
    left: 12px;
    right: 12px;
    top: 34%;
    width: auto;
    max-width: 360px;
  }

  .lulu-entity.is-away-event {
    width: min(92vw, 520px);
  }

  .daily-card,
  .streak-card,
  .residency-card,
  .community-card {
    min-width: 0;
    padding: 9px 10px;
    border-radius: 16px;
  }

  .daily-card strong,
  .streak-card strong,
  .residency-card strong,
  .community-card strong {
    margin-top: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
  }

  .daily-card small,
  .streak-card small,
  .residency-card small,
  .community-card small {
    margin-top: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
  }

  .mission-progress,
  .companion-progress,
  .residency-progress,
  .community-progress {
    height: 5px;
    margin-top: 6px;
  }

  .lulu-entity {
    grid-row: 3;
    align-self: center;
    width: min(76vw, 42svh, 360px);
  }

  .lumei-resident {
    right: -27%;
    width: 55%;
    height: 68%;
  }

  .lumei-resident.pose-sitting {
    right: -29%;
    width: 58%;
  }

  .lumei-speech {
    right: -2%;
    max-width: min(210px, 54vw);
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

  .care-advice {
    grid-column: 1 / -1;
    grid-row: 3;
  }

  .smart-care-btn {
    grid-column: 1;
    grid-row: 4;
  }

  .exp-row {
    grid-column: 1 / -1;
  }

  .log-btn {
    grid-row: 4;
    grid-column: 2;
    padding: 8px 10px;
  }

  .exp-bar {
    grid-row: 2;
    grid-column: 1 / -1;
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
  .streak-card,
  .residency-card,
  .community-card {
    padding: 8px 9px;
  }

  .daily-card span,
  .streak-card span,
  .residency-card span,
  .community-card span {
    font-size: 10px;
  }

  .lulu-entity {
    width: min(84vw, 40svh, 330px);
  }

  .lumei-resident {
    right: -8%;
    width: 48%;
  }

  .lumei-resident.pose-sitting {
    right: -10%;
    width: 52%;
  }

  .residency-world-card {
    grid-template-columns: 92px minmax(0, 1fr);
    padding: 14px;
  }

  .residency-world-card > img {
    width: 104px;
    height: 112px;
    margin-left: -12px;
  }

  .residency-story {
    grid-auto-flow: column;
    grid-auto-columns: 132px;
    grid-template-columns: none;
    overflow-x: auto;
    padding: 2px 2px 7px;
    scroll-snap-type: x proximity;
  }

  .story-chapter {
    scroll-snap-align: start;
  }

  .weekly-task-card {
    padding: 14px;
  }

  .weekly-task-title p {
    display: none;
  }

  .move-in-card {
    padding-inline: 20px;
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

  .world-panel {
    padding: 18px 14px;
  }

  .log-title-row {
    flex-direction: column;
    gap: 3px;
  }

}
</style>
