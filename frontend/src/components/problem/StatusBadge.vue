<template>
  <span :class="['status-badge', statusClass]">{{ label }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    required: true,
  },
})

const statusMap = {
  'Pending': { class: 'pending', label: '等待中' },
  'Judging': { class: 'judging', label: '评测中' },
  'Accepted': { class: 'accepted', label: '通过' },
  'Wrong Answer': { class: 'wrong-answer', label: '答案错误' },
  'Time Limit Exceeded': { class: 'time-limit', label: '超时' },
  'Memory Limit Exceeded': { class: 'memory-limit', label: '内存超限' },
  'Runtime Error': { class: 'runtime-error', label: '运行错误' },
  'Compile Error': { class: 'compile-error', label: '编译错误' },
  'System Error': { class: 'system-error', label: '系统错误' },
}

const statusClass = computed(() => {
  return statusMap[props.status]?.class || 'unknown'
})

const label = computed(() => {
  return statusMap[props.status]?.label || props.status
})
</script>

<style lang="scss" scoped>
.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 600;
  
  &.pending {
    background: #e9e9eb;
    color: #909399;
  }
  
  &.judging {
    background: #ecf5ff;
    color: #409eff;
  }
  
  &.accepted {
    background: #f0f9eb;
    color: #67c23a;
  }
  
  &.wrong-answer {
    background: #fef0f0;
    color: #f56c6c;
  }
  
  &.time-limit,
  &.memory-limit {
    background: #fdf6ec;
    color: #e6a23c;
  }
  
  &.runtime-error,
  &.compile-error {
    background: #fef0f0;
    color: #f56c6c;
  }
  
  &.system-error,
  &.unknown {
    background: #e9e9eb;
    color: #909399;
  }
}
</style>
