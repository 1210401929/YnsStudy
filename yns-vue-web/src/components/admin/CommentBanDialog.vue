<!-- 新增“禁止评论”规则的弹窗：评论限制页、评论管理、用户管理共用 -->
<template>
  <el-dialog :model-value="modelValue" title="禁止评论" width="460px" append-to-body @update:model-value="$emit('update:modelValue', $event)">
    <el-form label-width="80px" size="small" @submit.prevent>
      <el-form-item label="限制方式">
        <el-select v-model="form.type" :disabled="lockType" style="width: 100%">
          <el-option v-for="(info, key) in BAN_TYPES" :key="key" :label="info.label" :value="key"/>
        </el-select>
      </el-form-item>
      <el-form-item :label="BAN_TYPES[form.type].field">
        <el-input v-model="form.value" :placeholder="BAN_TYPES[form.type].placeholder" maxlength="255"/>
        <div class="tip">{{ BAN_TYPES[form.type].tip }}</div>
      </el-form-item>
      <el-form-item label="期限">
        <el-radio-group v-model="form.days">
          <el-radio-button :value="1">1 天</el-radio-button>
          <el-radio-button :value="7">7 天</el-radio-button>
          <el-radio-button :value="30">30 天</el-radio-button>
          <el-radio-button :value="0">永久</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" placeholder="选填，例如：发广告" maxlength="255"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button size="small" @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button size="small" type="danger" :loading="saving" :disabled="!form.value.trim()" @click="save">确认禁止</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { sendAxiosRequestChecked } from '@/utils/common.js'
import { BAN_TYPES } from './commentBanTypes.js'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  presetType: { type: String, default: 'ip' },
  presetValue: { type: String, default: '' },
  // 从评论或用户发起时，限制方式已确定，不允许修改
  lockType: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'saved'])

const form = reactive({ type: 'ip', value: '', days: 7, remark: '' })
const saving = ref(false)

watch(() => props.modelValue, (open) => {
  if (!open) return
  form.type = BAN_TYPES[props.presetType] ? props.presetType : 'ip'
  form.value = props.presetValue || ''
  form.days = 7
  form.remark = ''
})

const save = async () => {
  saving.value = true
  try {
    const result = await sendAxiosRequestChecked('/blog-api/admin/addCommentBan', { ...form, value: form.value.trim() }, '保存失败')
    if (!result) return
    ElMessage.success('已禁止评论')
    emit('update:modelValue', false)
    emit('saved')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.tip {
  margin-top: 4px;
  color: var(--j-muted);
  font-size: 12px;
  line-height: 1.6;
}
</style>
