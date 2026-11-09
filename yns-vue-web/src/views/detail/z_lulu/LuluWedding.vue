<template>
  <button type="button" class="wedding-entry" aria-haspopup="dialog" @click="open">
    <span class="wedding-entry-copy">
      <small>{{ wedding?.married ? 'Lv.70 · 婚礼纪念' : '婚礼预告 · Lv.70 举行' }}</small>
      <strong>{{ chapter?.title || '两只杯子的约定' }}</strong>
      <small>{{ wedding?.married ? '回看仪式与合照' : wedding?.currentLevel === 69 ? '再升 1 级就举行婚礼' : '打开婚礼手账' }}</small>
    </span>
    <img v-if="entryImage" :src="entryImage" alt="" />
  </button>

  <el-dialog v-model="visible" title="两只杯子的约定" width="min(600px, 92vw)" append-to-body>
    <div class="wedding-dialog">
      <p v-if="error && !wedding" role="status">婚礼手账暂时没能打开。<button type="button" class="wedding-button" @click="fetchWedding">重试</button></p>
      <p v-else-if="!wedding" role="status">正在打开婚礼手账…</p>
      <template v-else-if="chapter">
        <section v-if="wedding.married && ceremonyStep !== null" class="wedding-ceremony" aria-label="婚礼仪式">
          <img :src="ceremonyArt" alt="穿深蓝西装的噜噜和穿白婚纱的噜妹站在花拱下的木质典礼台上" />
          <div class="wedding-ceremony-copy" aria-live="polite">
            <small>婚礼仪式 · {{ ceremonyStep + 1 }}/{{ ceremony.length }}</small>
            <h3>{{ ceremony[ceremonyStep].title }}</h3>
            <p>{{ ceremony[ceremonyStep].text }}</p>
          </div>
          <div v-if="ceremonyStep === 1" class="wedding-vows" aria-label="两封誓词卡">
            <article v-for="card in vowCards" :key="card.id" class="wedding-vow" :class="{ 'is-open': openedVows.has(card.id) }">
              <button type="button" :aria-expanded="openedVows.has(card.id)" :aria-controls="`wedding-vow-${card.id}`" @click="openVow(card.id)">
                {{ card.author }}的誓词卡<span>{{ openedVows.has(card.id) ? '已拆开' : '点击拆封' }}</span>
              </button>
              <div :id="`wedding-vow-${card.id}`" aria-live="polite">
                <p v-if="openedVows.has(card.id)">「{{ card.text }}」</p>
              </div>
            </article>
          </div>
          <button type="button" class="wedding-button wedding-ceremony-next" :disabled="ceremonyStep === 1 && !allVowsOpened"
            @click="advanceCeremony">{{ ceremony[ceremonyStep].button }}</button>
        </section>
        <template v-else>
          <section class="wedding-summary">
            <img v-if="wedding.married" :src="photo" alt="穿深蓝西装的噜噜与穿白婚纱的噜妹的婚礼合照" />
            <div class="wedding-summary-copy">
              <small>{{ wedding.married ? (wedding.marriedAt || 'Lv.70') + ' · 婚礼纪念卡' : `Lv.${wedding.currentLevel} / Lv.${wedding.targetLevel}` }}</small>
              <h3>{{ chapter.title }}</h3>
              <p>{{ chapter.description }}</p>
              <template v-if="wedding.married">
                <button type="button" class="wedding-button" :disabled="!vowCards.length" @click="startCeremony">见证婚礼</button>
                <p class="wedding-next">婚房已经解锁，可通过「换场景」前往。</p>
              </template>
              <p v-else class="wedding-next">下一章 · Lv.{{ wedding.nextLevel }}</p>
            </div>
            <img v-if="!wedding.married && entryImage" class="wedding-preparation" :src="preparation" alt="筹备桌上的两只杯子、花束和信封" />
          </section>

          <details v-if="wedding.chapters.length > 1" class="wedding-journal">
            <summary>翻看之前的筹备记录</summary>
            <ol>
              <li v-for="item in wedding.chapters.slice(0, -1)" :key="item.level">
                <strong>Lv.{{ item.level }} · {{ item.title }}</strong>
                <p>{{ item.description }}</p>
              </li>
            </ol>
          </details>

          <section v-if="wedding.canBless" class="wedding-blessings">
            <div class="wedding-blessings-heading">
              <h4>祝福册 · {{ wedding.blessingCount }} 条</h4>
              <button type="button" class="wedding-text-button" @click="fetchWedding">刷新</button>
            </div>
            <form v-if="!wedding.blessedToday" class="wedding-form" @submit.prevent="sendBlessing">
              <label for="lulu-wedding-blessing">给他们留一句祝福</label>
              <textarea id="lulu-wedding-blessing" v-model="blessing" rows="2" placeholder="愿你们的每个早晨都有热乎乎的一杯…"></textarea>
              <div class="wedding-blessings-heading">
                <small>{{ [...blessing.trim()].length }}/60 · 同一网络每天可留一次</small>
                <button class="wedding-button" type="submit" :disabled="isSending || !blessing.trim() || [...blessing.trim()].length > 60">{{ isSending ? '送出中…' : '送出祝福' }}</button>
              </div>
            </form>
            <p v-else class="wedding-next">今天的祝福已经夹进回忆册了，谢谢你来。</p>
            <p v-if="!wedding.blessings.length" class="wedding-next">祝福册还空着，可以写下第一句。</p>
            <ul v-else class="wedding-wishes" aria-label="最新祝福">
              <li v-for="item in wedding.blessings" :key="item.id">{{ item.content }}</li>
            </ul>
          </section>
        </template>
      </template>
      <p v-else>婚礼邀请信将在 Lv.60 送达。</p>
    </div>
  </el-dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { sendAxiosRequest } from '@/utils/common.js';

const props = defineProps({ level: { type: Number, required: true }, introReady: Boolean });
const photo = '/picture/lulu/npc/lulu-wedding-photo-formal.webp';
const ceremonyArt = '/picture/lulu/npc/lulu-wedding-ceremony.webp';
const preparation = '/picture/lulu/npc/lulu-wedding-preparation.webp';
const introStorageKey = 'luluWeddingCeremonyIntroV2';
const previewStorageKey = 'luluWeddingPreviewSeen';
const visible = ref(false);
const wedding = ref(null);
const error = ref(false);
const blessing = ref('');
const isSending = ref(false);
const ceremonyStep = ref(null);
const openedVows = ref(new Set());
const introPending = ref(false);
const seenWeddingNotices = new Set();
let latestRequest = 0;

const chapter = computed(() => wedding.value?.chapters.at(-1));
const entryImage = computed(() => wedding.value?.married ? photo : (chapter.value?.level >= 65 ? preparation : ''));
const vowCards = computed(() => wedding.value?.vows || []);
const allVowsOpened = computed(() => vowCards.value.length > 0 && vowCards.value.every((card) => openedVows.value.has(card.id)));
const ceremony = [
  { title: '婚礼开始，请你做见证人', text: '花拱下的灯亮起来了。穿着西装的噜噜牵着穿婚纱的噜妹走上典礼台，请你替他们拆开两封誓词卡。', button: '打开誓词桌' },
  { title: '两封誓词，等你拆开', text: '可以先拆任意一封。两封都读完后，就轮到我们一起倒数啦。', button: '一起数三声' },
  { title: '三、二、一，碰杯！', text: '典礼台上的两只杯子叮地碰在一起。噜噜和噜妹结婚啦！这张合照，留给见证了今天的你。', button: '查看纪念合影' }
];

async function fetchWedding() {
  const request = ++latestRequest;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/wedding');
    if (!Array.isArray(result?.chapters) || !Array.isArray(result?.blessings)) throw new Error('无效的婚礼状态');
    if (request === latestRequest) {
      const wasPreparing = wedding.value !== null && !wedding.value.married;
      wedding.value = result;
      error.value = false;
      if (result.married && (wasPreparing || result.currentLevel === result.targetLevel) && !hasSeenNotice(introStorageKey)) {
        introPending.value = true;
      }
    }
  } catch (err) {
    if (request === latestRequest) error.value = true;
    console.error('读取噜噜婚礼手账失败:', err);
  }
}

function open() {
  if (introPending.value && vowCards.value.length) startCeremony();
  else {
    ceremonyStep.value = null;
    visible.value = true;
  }
  fetchWedding();
}

function hasSeenNotice(key) {
  if (seenWeddingNotices.has(key)) return true;
  try { return window.localStorage.getItem(key) === '1'; } catch { return false; }
}

function markNoticeSeen(key) {
  seenWeddingNotices.add(key);
  try { window.localStorage.setItem(key, '1'); } catch { /* Keep the notice dismissed for this visit. */ }
}

function startCeremony() {
  if (!wedding.value?.married || !vowCards.value.length) return;
  openedVows.value = new Set();
  ceremonyStep.value = 0;
  visible.value = true;
  introPending.value = false;
  markNoticeSeen(introStorageKey);
}

function openVow(id) {
  if (ceremonyStep.value !== 1 || !vowCards.value.some((card) => card.id === id)) return;
  openedVows.value.add(id);
}

function advanceCeremony() {
  if (ceremonyStep.value === null || (ceremonyStep.value === 1 && !allVowsOpened.value)) return;
  ceremonyStep.value = ceremonyStep.value + 1 < ceremony.length ? ceremonyStep.value + 1 : null;
}

async function sendBlessing() {
  const content = blessing.value.trim();
  if (!wedding.value?.canBless || wedding.value.blessedToday || isSending.value || !content || [...content].length > 60) return;
  isSending.value = true;
  try {
    const result = await sendAxiosRequest('/blog-api/lulu/wedding/blessing', { content });
    if (result?.isError !== false) {
      ElMessage.error(result?.errMsg || '祝福没能保存，请稍后重试');
      await fetchWedding();
      return;
    }
    blessing.value = '';
    wedding.value = { ...wedding.value, blessedToday: true };
    ElMessage.success('祝福已经收进回忆册啦');
    await fetchWedding();
  } catch (err) {
    console.error('保存噜噜婚礼祝福失败:', err);
    ElMessage.error('祝福没能保存，请稍后重试');
  } finally {
    isSending.value = false;
  }
}

watch(() => props.level, fetchWedding, { immediate: true });
watch([wedding, () => props.introReady, visible], ([state, ready, isOpen]) => {
  if (!state || state.married || !chapter.value || hasSeenNotice(previewStorageKey)) return;
  if (!isOpen && !ready) return;
  markNoticeSeen(previewStorageKey);
  visible.value = true;
});
watch([introPending, () => props.introReady, vowCards], ([pending, ready, cards]) => {
  if (pending && ready && cards.length) startCeremony();
});
defineExpose({ open });
</script>

<style scoped>
.wedding-entry {
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 9px 12px;
  text-align: left;
  border: 1px solid #f2c9bb;
  border-radius: 15px;
  background: #fff2e8;
  color: #66382f;
  cursor: pointer;
}
.wedding-entry-copy { display: grid; gap: 2px; }
.wedding-entry-copy small, .wedding-next { color: #7a574d; }
.wedding-entry img { width: 72px; height: 46px; flex: none; object-fit: contain; border-radius: 8px; }
.wedding-dialog { max-height: 68vh; overflow: auto; display: grid; gap: 12px; padding: 2px; }
.wedding-dialog p, .wedding-dialog h3, .wedding-dialog h4 { margin: 0; }
.wedding-summary, .wedding-ceremony, .wedding-journal, .wedding-blessings {
  border: 1px solid #f2dfd6;
  border-radius: 14px;
  overflow: hidden;
  background: #fffbf7;
  color: #523b35;
}
.wedding-summary > img, .wedding-ceremony > img { width: 100%; height: auto; max-height: 300px; object-fit: contain; display: block; }
.wedding-summary > .wedding-preparation { width: min(44%, 175px); margin: -2px auto 10px; }
.wedding-summary-copy, .wedding-journal, .wedding-blessings { padding: 12px; }
.wedding-summary-copy, .wedding-ceremony, .wedding-blessings, .wedding-form { display: grid; gap: 9px; }
.wedding-ceremony > :not(img) { margin-inline: 12px; }
.wedding-ceremony-copy { display: grid; gap: 6px; }
.wedding-ceremony-next { margin-bottom: 12px; }
.wedding-vows { display: grid; gap: 8px; grid-template-columns: repeat(auto-fit, minmax(min(100%, 190px), 1fr)); }
.wedding-vow { min-width: 0; border: 1px solid #e7c6b7; border-radius: 10px; background: #fff2e8; }
.wedding-vow.is-open { background: #fffaf5; }
.wedding-vow button { width: 100%; display: grid; gap: 4px; padding: 10px; border: 0; border-radius: inherit; background: transparent; color: #703c32; text-align: left; font: inherit; font-weight: 700; cursor: pointer; }
.wedding-vow button span { font-size: 12px; font-weight: 400; }
.wedding-vow p { padding: 0 10px 10px; line-height: 1.65; }
.wedding-blessings-heading { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 6px; }
.wedding-journal summary, .wedding-text-button { cursor: pointer; }
.wedding-journal ol { padding-left: 22px; margin-bottom: 0; }
.wedding-journal li + li { margin-top: 8px; }
.wedding-button, .wedding-text-button { justify-self: start; border: 0; padding: 6px 10px; border-radius: 8px; font: inherit; cursor: pointer; }
.wedding-button { color: #fff; background: #965144; }
.wedding-button:disabled { opacity: .55; cursor: default; }
.wedding-text-button { background: transparent; color: #7e4339; text-decoration: underline; }
.wedding-form textarea { width: 100%; box-sizing: border-box; border: 1px solid #d9b8a9; border-radius: 8px; font: inherit; padding: 7px; resize: vertical; }
.wedding-wishes { list-style: none; margin: 0; padding: 0; display: grid; gap: 6px; }
.wedding-wishes li { padding: 7px 9px; background: #f7eee7; border-radius: 8px; overflow-wrap: anywhere; white-space: pre-wrap; }
</style>
