<!--背景图片和背景音乐-->
<template>
  <div class="left-dock"
       v-if="isSelf"
       :class="{ collapsed: dockCollapsed }">
    <el-card shadow="hover" class="dock-card">
      <div class="dock-title">🎨 背景设置</div>

      <div class="dock-block">
        <div class="dock-subtitle">图片</div>
        <el-input
            v-model="bgImageInput"
            size="small"
            placeholder="粘贴图片URL后回车"
            @keyup.enter="applyBgImageUrl"
            style="margin-top:8px"/>
        <div class="dock-actions">
          <el-button size="small" @click="resetBgImage">恢复默认</el-button>
          <el-button size="small" link @click="previewBgImage" :disabled="!bgImage">预览</el-button>
        </div>
      </div>

      <el-divider/>

      <div class="dock-title">
        {{ isSelf ? "🎵 音乐设置" : `${userName}的背景音乐` }}
      </div>
      <div class="dock-block">
        <el-input
            v-model="bgAudioInput"
            size="small"
            placeholder="粘贴音频URL后回车"
            @keyup.enter="applyBgAudioUrl"
            style="margin-top:8px"/>

        <div class="audio-row">
          <el-button size="small" :type="isAudioPlaying ? 'danger' : 'primary'" @click="toggleAudio">
            {{ isAudioPlaying ? '暂停' : '播放' }}
          </el-button>
          <el-switch
              v-model="audioMuted"
              size="small"
              active-text="静音"
              style="margin-left:12px"
              @change="applyMute"/>
        </div>

        <div class="audio-row" style="margin-top:8px">
          <span class="vol-label">音量</span>
          <el-slider v-model="audioVolume" :min="0" :max="100" :step="1" style="flex:1" @change="applyVolume"/>
        </div>
      </div>
    </el-card>

    <button class="dock-toggle" @click="dockCollapsed = !dockCollapsed">
      {{ dockCollapsed ? '▶' : '◀' }}
    </button>
  </div>

  <audio
      ref="bgAudioRef"
      :src="bgAudio || ''"
      loop
      preload="metadata"
      playsinline
  ></audio>

  <div class="music-mini" v-if="!isSelf && bgAudio">
    <div class="music-mini__title">🎵 {{ userName || 'Ta' }} 的音乐</div>
    <div class="music-mini__row mini-controls">
      <el-button
          class="icon-btn"
          circle
          size="small"
          :title="isAudioPlaying ? '暂停' : '播放'"
          :type="isAudioPlaying ? 'danger' : 'primary'"
          @click="toggleAudio"
      >
        {{ isAudioPlaying ? '⏸' : '▶' }}
      </el-button>

      <el-button
          class="icon-btn"
          circle
          size="small"
          :title="audioMuted ? '取消静音' : '静音'"
          @click="audioMuted = !audioMuted; applyMute()"
      >
        {{ audioMuted ? '🔇' : '🔊' }}
      </el-button>

      <el-slider
          class="mini-slider"
          v-model="audioVolume"
          :min="0"
          :max="100"
          :step="1"
          :show-tooltip="false"
          @input="applyVolume"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { sendAxiosRequest } from '@/utils/common.js'
import { useUserStore } from '@/stores/main/user.js'

const props = defineProps({
  isSelf: { type: Boolean, default: false },
  userName: { type: String, default: '' },
  initBgImage: { type: String, default: '' },
  initBgAudio: { type: String, default: '' }
})

const emit = defineEmits(['updateBgStyle'])
const userStore = useUserStore()

// 面板折叠状态
const dockCollapsed = ref(false)

// ==== 背景图片逻辑 ====
const bgPresets = {
  softSky: "radial-gradient(1000px 700px at 20% 15%, rgba(210,235,255,.90) 0%, rgba(210,235,255,0) 60%), radial-gradient(800px 600px at 80% 70%, rgba(180,205,230,.35) 0%, rgba(180,205,230,0) 60%), linear-gradient(180deg,#f8fbff 0%, #ffffff 68%)"
}
const bgImage = ref(bgPresets.softSky)
const bgImageInput = ref("")

const bgStyle = computed(() => {
  const v = (bgImage.value || '').trim();
  if (!v) return {};
  if (/gradient\(/i.test(v) || v.includes(',')) {
    return { backgroundImage: v };
  }
  return { backgroundImage: `url('${v}')` };
})

watch(bgStyle, (newStyle) => {
  emit('updateBgStyle', newStyle)
}, { immediate: true })

watch(() => props.initBgImage, async (newVal) => {
  if (newVal !== undefined) {
    bgImageInput.value = newVal
    await setBgImageSafely(newVal)
  }
})

const applyBgImageUrl = () => {
  bgImage.value = bgImageInput.value.trim()
  sendAxiosRequest("/blog-api/userInformation/setPersonInfo", {
    userCode: userStore.userBean.code,
    fieldName: "BGIMAGEURL",
    fieldValue: bgImage.value
  })
  ElMessage.success("已应用")
}

const resetBgImage = () => {
  bgImageInput.value = ""
  applyBgImageUrl()
  ElMessage.success("已恢复默认背景")
}

const previewBgImage = () => {
  if (!bgImage.value) return;
  window.open(bgImage.value, "_blank")
}

async function setBgImageSafely(url) {
  if (!url) {
    bgImage.value = '';
    return;
  }
  const img = new Image();
  img.src = url;
  try {
    if (img.decode) await img.decode();
    else await new Promise(r => { img.onload = r; img.onerror = r; });
  } finally {
    bgImage.value = url;
    await nextTick();
    const root = document.querySelector('.person-info');
    if (root) {
      root.classList.add('bg-fade');
      setTimeout(() => root.classList.remove('bg-fade'), 260);
    }
  }
}

// ==== 背景音乐逻辑 ====
const bgAudioRef = ref(null)
const bgAudio = ref("")
const bgAudioInput = ref("")
const isAudioPlaying = ref(false)
const audioVolume = ref(70)
const audioMuted = ref(false)

watch(() => props.initBgAudio, (newVal) => {
  if (newVal !== undefined) {
    bgAudioInput.value = newVal
    setAudioSrc(newVal)
  }
})

const applyBgAudioUrl = () => {
  sendAxiosRequest("/blog-api/userInformation/setPersonInfo", {
    userCode: userStore.userBean.code,
    fieldName: "BGMUSICURL",
    fieldValue: bgAudioInput.value.trim()
  })
  setAudioSrc(bgAudioInput.value.trim())
  ElMessage.success("背景音乐已设置")
}

function setAudioSrc(src) {
  bgAudio.value = src
  nextTickPlayIfWanted()
}

const toggleAudio = async () => {
  const el = bgAudioRef.value
  if (!el) return
  if (isAudioPlaying.value) {
    el.pause()
    isAudioPlaying.value = false
  } else {
    try {
      await el.play()
      isAudioPlaying.value = true
    } catch (e) {
      console.error("播放音乐失败,音乐链接或已失效")
    }
  }
}

const applyVolume = () => {
  if (bgAudioRef.value) bgAudioRef.value.volume = Math.min(1, Math.max(0, audioVolume.value / 100))
}

const applyMute = () => {
  if (bgAudioRef.value) bgAudioRef.value.muted = !!audioMuted.value
}

const nextTickPlayIfWanted = async () => {
  await Promise.resolve()
  applyVolume()
  applyMute()
}

// 暴露给父组件的方法
const playMusicForce = async () => {
  const el = bgAudioRef.value
  if (el && bgAudio.value) {
    applyVolume()
    applyMute()
    el.currentTime = 0
    try {
      await el.play()
      isAudioPlaying.value = true
    } catch (e) {
      console.error("播放音乐失败,音乐链接或已失效")
    }
  }
}

defineExpose({ playMusicForce })

onMounted(() => {
  applyVolume()
  applyMute()
})
</script>

<style scoped>
.left-dock {
  position: fixed;
  left: 20px;
  top: 120px;
  width: 260px;
  z-index: 20;
  transition: transform 0.25s ease, opacity 0.25s ease;
}

.left-dock.collapsed {
  transform: translateX(-280px);
}

.dock-card {
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: saturate(180%) blur(6px);
  border-radius: 12px;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.12);
  padding-bottom: 12px;
}

.dock-title {
  font-weight: 700;
  color: #333;
  margin-bottom: 8px;
}

.dock-subtitle {
  font-weight: 600;
  font-size: 13px;
  color: #666;
  margin-bottom: 6px;
}

.dock-block {
  margin-bottom: 12px;
}

.dock-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.audio-row {
  padding-top: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.vol-label {
  font-size: 12px;
  color: #666;
  width: 36px;
}

.dock-toggle {
  position: absolute;
  right: -18px;
  top: 10px;
  width: 22px;
  height: 28px;
  border-radius: 0 6px 6px 0;
  border: 1px solid #dcdfe6;
  background: #fff;
  color: #666;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.music-mini {
  position: fixed;
  left: 20px;
  top: 20px;
  width: 220px;
  z-index: 19;
  padding: 8px 10px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: saturate(180%) blur(6px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
  transition: transform .18s ease, box-shadow .18s ease, background .18s ease;
}

.music-mini__title {
  font-size: 12px;
  font-weight: 600;
  color: #333;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.music-mini__row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  width: 28px;
  height: 28px;
  padding: 0;
  font-size: 14px;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.mini-slider {
  flex: 1;
  margin-left: 6px;
}

.mini-slider :deep(.el-slider__runway),
.mini-slider :deep(.el-slider__bar) {
  height: 3px;
  border-radius: 999px;
}

.mini-slider :deep(.el-slider__button-wrapper) {
  width: 12px;
  height: 12px;
}

.mini-slider :deep(.el-slider__button) {
  width: 10px;
  height: 10px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

@media (max-width: 768px) {
  .left-dock { transform: translateX(-280px); }
  .left-dock.collapsed { transform: translateX(-280px); }
  .left-dock .dock-toggle { transform: translateX(0px); }
  .music-mini {
    left: 50%;
    transform: translateX(-50%);
    top: 12px;
    width: calc(100vw - 100px);
    padding: 8px 10px;
  }
}
</style>