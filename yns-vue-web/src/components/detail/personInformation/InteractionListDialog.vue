<template>
  <el-dialog
      :model-value="modelValue"
      @update:model-value="$emit('update:modelValue', $event)"
      :title="title"
      width="60%"
      :modal-class="'modern-dialog-wrapper'"
      top="6vh"
      destroy-on-close
  >
    <div class="card-list-scroll">
      <div class="card-list">

        <template v-if="type === 'blog'">
          <div
              v-for="item in listData"
              :key="item.GUID"
              class="modern-card blog-card"
              @click="$emit('item-click', item)"
          >
            <div class="blog-card-content">
              <h4 class="blog-title">{{ item.BLOG_TITLE }}</h4>
              <div class="blog-meta">
                <span class="meta-item author">
                  <span class="meta-dot"></span>
                  {{ item.USERNAME }}
                </span>
                <span class="meta-item time">{{ formatDate(item.CREATE_TIME) }}</span>
              </div>
            </div>
            <div class="card-action-arrow"></div>
          </div>
        </template>

        <template v-else-if="type === 'user'">
          <div
              v-for="item in listData"
              :key="item.CODE"
              class="modern-card user-card"
              @click="$emit('item-click', item)"
          >
            <el-avatar
                :src="item.avatar"
                :size="52"
                class="user-avatar"
                alt="用户头像"
            >
              {{ item.NAME?.charAt(0) }}
            </el-avatar>

            <div class="user-info">
              <div class="user-name">{{ item.NAME }}</div>
              <div class="user-remark" :title="item.REMARK">
                {{ item.REMARK || '这个人很懒，暂时没有留下个性签名～' }}
              </div>
            </div>

            <div class="card-action-arrow"></div>
          </div>
        </template>

      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { pubFormatDate } from '@/utils/common.js'

// 接收父组件传来的参数
defineProps({
  modelValue: { type: Boolean, default: false }, // 控制弹窗显隐
  title: { type: String, default: '列表' },      // 弹窗标题
  listData: { type: Array, default: () => [] },  // 列表数据
  type: { type: String, default: 'blog' }        // 列表类型: 'blog' 或 'user'
})

// 声明抛给父组件的事件：关闭弹窗、点击卡片
defineEmits(['update:modelValue', 'item-click'])

// 时间格式化工具
const formatDate = (dateStr) => pubFormatDate(dateStr)
</script>

<style scoped>
/* ========== 弹窗容器 & 滚动条美化 ========== */
.card-list-scroll {
  max-height: 550px;
  overflow-y: auto;
  padding: 8px 16px 20px 16px;
  box-sizing: border-box;
}

/* 现代感细线滚动条 */
.card-list-scroll::-webkit-scrollbar {
  width: 6px;
}
.card-list-scroll::-webkit-scrollbar-thumb {
  background-color: #dcdfe6;
  border-radius: 4px;
}
.card-list-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.card-list-scroll:hover::-webkit-scrollbar-thumb {
  background-color: #c0c4cc;
}

.card-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* ========== 通用现代卡片样式 ========== */
.modern-card {
  position: relative;
  background: #ffffff;
  border: 1px solid #ebeef5;
  border-radius: 12px;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  overflow: hidden;
}

.modern-card:hover {
  transform: translateY(-2px);
  border-color: #c6e2ff;
  box-shadow: 0 8px 20px rgba(64, 158, 255, 0.08);
}

/* 卡片右侧的箭头指示器 (纯CSS实现，无需引入Icon库) */
.card-action-arrow {
  width: 8px;
  height: 8px;
  border-top: 2px solid #c0c4cc;
  border-right: 2px solid #c0c4cc;
  transform: rotate(45deg);
  transition: border-color 0.3s;
  flex-shrink: 0;
  margin-left: 16px;
}

.modern-card:hover .card-action-arrow {
  border-color: #409eff;
}

/* ========== 博客文章卡片定制 ========== */
.blog-card {
  padding-left: 24px;
}

/* 左侧强调色带 */
.blog-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  background: #409eff;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.blog-card:hover::before {
  opacity: 1;
}

.blog-card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.blog-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  line-height: 1.4;
  /* 标题最多显示两行，超出省略 */
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.blog-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: #909399;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.meta-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #409eff;
}

.author {
  color: #409eff;
  font-weight: 500;
}

/* ========== 用户/粉丝卡片定制 ========== */
.user-card {
  padding: 14px 20px;
}

.user-avatar {
  background-color: #409eff;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  flex-shrink: 0;
  border: 2px solid #ecf5ff;
}

.user-info {
  flex: 1;
  margin-left: 16px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0; /* 确保文字截断生效 */
  gap: 6px;
}

.user-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.user-remark {
  font-size: 13px;
  color: #909399;
  /* 单行文字溢出省略号 */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
}

/* ========== 移动端响应式 ========== */
@media (max-width: 768px) {
  .modern-card {
    padding: 14px;
  }
  .blog-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  .card-action-arrow {
    display: none; /* 移动端空间小，可隐藏引导箭头 */
  }
}
</style>

/* 如果你需要修改全局弹窗样式，可以将这部分放入不带 scoped 的 style 标签或全局 CSS 中 */
<style>
.modern-dialog-wrapper .el-dialog__header {
  border-bottom: 1px solid #ebeef5;
  margin-right: 0;
  padding-bottom: 16px;
}
.modern-dialog-wrapper .el-dialog__body {
  padding: 10px 0; /* 让滚动区域贴紧边缘 */
}
</style>