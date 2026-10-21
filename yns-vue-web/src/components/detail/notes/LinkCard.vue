<!-- 链接卡片：网站图标、网站名、标题、简介、配图。点击在新标签页打开（用户分享的外链加 nofollow ugc） -->
<template>
  <a class="link-card" :class="{ compact, 'no-image': !showImage }" :href="link.url"
     target="_blank" rel="nofollow ugc noopener noreferrer" @click.stop>
    <span class="link-text">
      <span class="link-site">
        <img v-if="showIcon" :src="link.icon" alt="" class="link-icon" referrerpolicy="no-referrer" @error="iconFailed = true">
        <span class="link-site-name">{{ link.siteName || host }}</span>
      </span>
      <span class="link-title">{{ link.title || link.url }}</span>
      <span v-if="link.description" class="link-desc">{{ link.description }}</span>
    </span>
    <span v-if="showImage" class="link-image">
      <img :src="link.image" alt="" loading="lazy" referrerpolicy="no-referrer" @error="imageFailed = true">
    </span>
  </a>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  // { url, title, description, image, siteName, icon }
  link: { type: Object, required: true },
  // 便签墙卡片里用小一号的样式
  compact: { type: Boolean, default: false }
})

// 对方网站的图片可能失效或禁止外链，加载失败就不显示
const iconFailed = ref(false)
const imageFailed = ref(false)
const showIcon = computed(() => !!props.link.icon && !iconFailed.value)
const showImage = computed(() => !!props.link.image && !imageFailed.value)
const host = computed(() => {
  try {
    return new URL(props.link.url).hostname.replace(/^www\./, '')
  } catch {
    return props.link.url
  }
})
</script>

<style scoped>
/* 牛皮纸色的小卡片，像夹在便签上的一张剪报 */
.link-card {
  display: flex;
  align-items: stretch;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid #e2d5bb;
  border-left: 3px solid #d8c49d;
  border-radius: 3px;
  background: #fbf7ee;
  color: var(--j-ink);
  text-decoration: none;
  transition: border-color 0.15s, background-color 0.15s;
  overflow: hidden;
}

.link-card:hover {
  border-color: var(--j-pen);
  background: #fff;
}

.link-text {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.link-site {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--j-muted);
}

.link-icon {
  width: 14px;
  height: 14px;
  object-fit: contain;
  flex-shrink: 0;
}

.link-site-name {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.link-title {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.link-desc {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-size: 13px;
  line-height: 1.55;
  color: var(--j-ink-soft);
  overflow-wrap: anywhere;
}

.link-image {
  flex-shrink: 0;
  width: 120px;
  align-self: center;
}

.link-image img {
  display: block;
  width: 100%;
  aspect-ratio: 1.6;
  object-fit: cover;
  border-radius: 2px;
}

/* 便签墙卡片里：小一号，配图更窄，简介只留一行 */
.link-card.compact {
  gap: 10px;
  padding: 8px 10px;
}

.compact .link-title {
  font-size: 13.5px;
}

.compact .link-desc {
  -webkit-line-clamp: 1;
  font-size: 12px;
}

.compact .link-image {
  width: 72px;
}

@media (max-width: 640px) {
  .link-image {
    width: 84px;
  }
}
</style>
