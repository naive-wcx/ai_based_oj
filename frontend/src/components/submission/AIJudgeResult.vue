<template>
  <div class="ai-report-container" v-if="result">
    <!-- 状态头 -->
    <div class="ai-header" :class="result.passed ? 'passed' : 'failed'">
      <div class="status-indicator"></div>
      <span class="status-text">{{ result.passed ? '符合题目要求' : '未满足题目要求' }}</span>
    </div>

    <!-- 总结 -->
    <div class="ai-summary">
      <p>{{ result.reason || result.summary }}</p>
    </div>

    <!-- 详情网格 -->
    <div class="ai-details">
      <!-- 算法 -->
      <div class="detail-row">
        <span class="label">检测到的算法</span>
        <span class="value">{{ result.algorithm_detected || '未检测到特定算法' }}</span>
      </div>

      <!-- 语言合规 -->
      <div class="detail-row">
        <span class="label">语言/特性检查</span>
        <span class="value" :class="result.language_check === 'passed' ? 'text-success' : 'text-danger'">
          {{ result.language_check === 'passed' ? '通过' : '不通过' }}
        </span>
      </div>

      <!-- 对比 (仅失败时显示) -->
      <div class="comparison-box" v-if="result.details && !result.passed">
        <div class="comparison-row">
          <span class="sub-label">题目要求</span>
          <span class="sub-value text-success">{{ result.details.required }}</span>
        </div>
        <div class="comparison-row">
          <span class="sub-label">实际检测</span>
          <span class="sub-value text-danger">{{ result.details.detected }}</span>
        </div>
        <div class="comparison-row">
          <span class="sub-label">置信度</span>
          <span class="sub-value">{{ (result.details.confidence * 100).toFixed(0) }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  result: Object
})
</script>

<style lang="scss" scoped>
.ai-report-container {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 24px;
}

.ai-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  font-weight: 600;
  font-size: 16px;

  .status-indicator {
    width: 12px;
    height: 12px;
    border-radius: 50%;
  }

  &.passed {
    color: var(--swiss-success);
    .status-indicator { background-color: var(--swiss-success); }
  }

  &.failed {
    color: var(--swiss-danger);
    .status-indicator { background-color: var(--swiss-danger); }
  }
}

.ai-summary {
  font-size: 15px;
  color: var(--swiss-text-main);
  line-height: 1.6;
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--swiss-border-light);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  font-size: 14px;
  
  .label { color: var(--swiss-text-secondary); }
  .value { font-weight: 500; }
}

.comparison-box {
  margin-top: 16px;
  background: #fafafa;
  padding: 16px;
  border-radius: var(--radius-xs);
  
  .comparison-row {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
    font-size: 13px;
    
    &:last-child { margin-bottom: 0; }
  }
  
  .sub-label { color: var(--swiss-text-secondary); text-transform: uppercase; font-size: 11px; letter-spacing: 0.05em; }
}

.text-success { color: var(--swiss-success); }
.text-danger { color: var(--swiss-danger); }
</style>