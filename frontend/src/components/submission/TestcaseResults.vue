<template>
  <div class="testcases-grid">
    <div 
      v-for="(result, index) in results" 
      :key="index"
      class="testcase-dot"
      :class="getStatusClass(result.status)"
      :title="`#${index + 1}: ${getStatusLabel(result.status)} (${result.time}ms, ${formatMemory(result.memory)})`"
    >
      <span class="dot-id">{{ index + 1 }}</span>
    </div>
  </div>
  
  <div class="testcases-summary">
    <span>通过: <strong class="text-ac">{{ passedCount }}</strong> / {{ results.length }}</span>
    <span class="divider">|</span>
    <span>最大用时: <strong>{{ maxTime }}ms</strong></span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  results: {
    type: Array,
    default: () => [],
  }
})

const passedCount = computed(() => {
  return props.results.filter(r => r.status === 'Accepted').length
})

const maxTime = computed(() => {
  if (!props.results.length) return 0
  return Math.max(...props.results.map(r => r.time || 0))
})

function getStatusClass(status) {
  const map = {
    'Accepted': 'ac',
    'Wrong Answer': 'wa',
    'Time Limit Exceeded': 'tle',
    'Memory Limit Exceeded': 'mle',
    'Runtime Error': 're',
    'Compile Error': 'ce',
    'System Error': 'uqe',
    'Pending': 'pending',
    'Judging': 'judging'
  }
  return map[status] || 'pending'
}

function getStatusLabel(status) {
  const map = {
    'Accepted': '通过',
    'Wrong Answer': '答案错误',
    'Time Limit Exceeded': '超时',
    'Memory Limit Exceeded': '内存超限',
    'Runtime Error': '运行错误',
    'Compile Error': '编译错误',
    'System Error': '系统错误',
    'Pending': '等待中',
    'Judging': '评测中'
  }
  return map[status] || status
}

function formatMemory(kb) {
  if (!kb) return '-'
  return kb < 1024 ? `${kb}KB` : `${(kb/1024).toFixed(1)}MB`
}
</script>

<style lang="scss" scoped>
.testcases-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
  gap: 8px;
  margin-bottom: 20px;
}

.testcase-dot {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
  cursor: help;
  transition: all 0.2s;
  color: white;
  
  &:hover { 
    transform: scale(1.05);
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  }
  
  &.ac { background-color: var(--status-ac); }
  &.wa { background-color: var(--status-wa); }
  &.tle { background-color: var(--status-tle); }
  &.mle { background-color: var(--status-mle); }
  &.re { background-color: var(--status-re); }
  &.ce { background-color: var(--status-ce); }
  &.uqe { background-color: var(--status-uqe); }
  &.judging { background-color: var(--status-judging); }
  &.pending { 
    background-color: #f0f2f5; 
    color: var(--swiss-text-secondary);
    border: 1px solid var(--swiss-border-light);
  }
}

.testcases-summary {
  font-size: 14px;
  color: var(--swiss-text-secondary);
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--swiss-border-light);
  
  strong { color: var(--swiss-text-main); font-weight: 600; }
  .text-ac { color: var(--status-ac); }
  .divider { color: var(--swiss-border-light); }
}
</style>
