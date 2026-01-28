<template>
  <div class="testcase-results">
    <div class="results-header">
      <h4>测试点结果</h4>
      <div class="results-summary">
        <span class="passed">✓通过: {{ passedCount }}</span>
        <span class="total">总计: {{ results.length }}</span>
        <span class="max-time" v-if="maxTime">最大用时: {{ maxTime }}ms</span>
      </div>
    </div>
    
    <div class="results-grid">
      <div
        v-for="(result, index) in results"
        :key="index"
        :class="['result-item', getStatusClass(result.status)]"
      >
        <div class="result-header">
          <span class="result-id">#{{ result.id || index + 1 }}</span>
          <span :class="['result-status', getStatusClass(result.status)]">
            {{ getStatusLabel(result.status) }}
          </span>
        </div>
        <div class="result-info" v-if="result.time || result.memory">
          <span v-if="result.time" class="time">{{ result.time }}ms</span>
          <span v-if="result.memory" class="memory">{{ formatMemory(result.memory) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  results: {
    type: Array,
    default: () => [],
  },
})

const passedCount = computed(() => {
  return props.results.filter(r => r.status === 'Accepted').length
})

const maxTime = computed(() => {
  if (!props.results.length) return 0
  return Math.max(...props.results.map(r => r.time || 0))
})

function getStatusClass(status) {
  if (status === 'Accepted') return 'passed'
  return 'failed'
}

function getStatusLabel(status) {
  const labels = {
    'Accepted': '通过',
    'Wrong Answer': '答案错误',
    'Time Limit Exceeded': '超时',
    'Memory Limit Exceeded': '内存超限',
    'Runtime Error': '运行错误',
    'Compile Error': '编译错误',
    'System Error': '系统错误',
  }
  return labels[status] || status
}

function formatMemory(kb) {
  if (!kb) return ''
  if (kb < 1024) return `${kb}KB`
  return `${(kb / 1024).toFixed(1)}MB`
}
</script>

<style lang="scss" scoped>
.testcase-results {
  .results-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 12px;
    
    h4 {
      margin: 0;
      color: #303133;
      font-size: 16px;
    }
  }
  
  .results-summary {
    display: flex;
    gap: 16px;
    font-size: 14px;
    
    .passed {
      color: #67c23a;
      font-weight: 600;
    }
    
    .total {
      color: #909399;
    }
    
    .max-time {
      color: #409eff;
    }
  }
}

.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}

.result-item {
  padding: 12px;
  border-radius: 8px;
  border: 1px solid;
  
  &.passed {
    background: #f0f9eb;
    border-color: #e1f3d8;
  }
  
  &.failed {
    background: #fef0f0;
    border-color: #fde2e2;
  }
  
  .result-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }
  
  .result-id {
    font-size: 12px;
    color: #909399;
    font-weight: 500;
  }
  
  .result-status {
    font-size: 13px;
    font-weight: 600;
    
    &.passed {
      color: #67c23a;
    }
    
    &.failed {
      color: #f56c6c;
    }
  }
  
  .result-info {
    display: flex;
    gap: 8px;
    font-size: 12px;
    color: #606266;
    
    .time, .memory {
      background: rgba(0, 0, 0, 0.05);
      padding: 2px 6px;
      border-radius: 4px;
    }
  }
}
</style>
