<template>
  <div class="ai-judge-result" v-if="result && result.enabled">
    <div class="ai-status" :class="result.passed ? 'passed' : 'failed'">
      <span class="status-icon">{{ result.passed ? '✓' : '✗' }}</span>
      <span class="status-text">{{ result.passed ? '符合要求' : '不符合要求' }}</span>
    </div>
    
    <div class="ai-content">
      <!-- 检测到的算法 -->
      <div class="ai-item" v-if="result.algorithm_detected">
        <div class="item-label">检测到的算法</div>
        <div class="item-value">
          <el-tag type="primary">{{ result.algorithm_detected }}</el-tag>
        </div>
      </div>
      
      <!-- 语言检查 -->
      <div class="ai-item" v-if="result.language_check">
        <div class="item-label">语言检查</div>
        <div class="item-value">
          <el-tag :type="result.language_check === 'passed' ? 'success' : 'danger'">
            {{ result.language_check === 'passed' ? '通过' : '不通过' }}
          </el-tag>
        </div>
      </div>
      
      <!-- 分析结论 -->
      <div class="ai-item reason" v-if="result.reason">
        <div class="item-label">分析结论</div>
        <div class="item-value reason-text">
          {{ result.reason }}
        </div>
      </div>
      
      <!-- 详细对比 -->
      <div class="ai-comparison" v-if="result.details && !result.passed">
        <div class="comparison-title">要求与实际对比</div>
        <div class="comparison-grid">
          <div class="comparison-item">
            <div class="comparison-label">题目要求</div>
            <div class="comparison-value required">{{ result.details.required || '-' }}</div>
          </div>
          <div class="comparison-item">
            <div class="comparison-label">实际检测</div>
            <div class="comparison-value detected">{{ result.details.detected || '-' }}</div>
          </div>
          <div class="comparison-item" v-if="result.details.confidence">
            <div class="comparison-label">置信度</div>
            <div class="comparison-value">
              <el-progress 
                :percentage="Math.round(result.details.confidence * 100)" 
                :stroke-width="8"
                :color="getConfidenceColor(result.details.confidence)"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 展开 AI 总结 -->
    <el-collapse v-if="result.summary || result.reason" class="raw-response">
      <el-collapse-item title="查看 AI 总结">
        <pre>{{ result.summary || result.reason }}</pre>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup>
defineProps({
  result: {
    type: Object,
    default: null,
  },
})

function getConfidenceColor(confidence) {
  if (confidence >= 0.8) return '#67c23a'
  if (confidence >= 0.6) return '#e6a23c'
  return '#f56c6c'
}

</script>

<style lang="scss" scoped>
.ai-judge-result {
  background: linear-gradient(135deg, #f5f7ff 0%, #f0f5ff 100%);
  border: 1px solid #d9e3ff;
  border-radius: 12px;
  padding: 20px;
}

.ai-status {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  
  &.passed {
    background: linear-gradient(135deg, #f0f9eb 0%, #e1f3d8 100%);
    
    .status-icon {
      background: #67c23a;
    }
    
    .status-text {
      color: #67c23a;
    }
  }
  
  &.failed {
    background: linear-gradient(135deg, #fef0f0 0%, #fde2e2 100%);
    
    .status-icon {
      background: #f56c6c;
    }
    
    .status-text {
      color: #f56c6c;
    }
  }
  
  .status-icon {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: 700;
    font-size: 18px;
  }
  
  .status-text {
    font-size: 18px;
    font-weight: 600;
  }
}

.ai-content {
  .ai-item {
    display: flex;
    margin-bottom: 12px;
    
    .item-label {
      width: 120px;
      flex-shrink: 0;
      color: #909399;
      font-size: 14px;
    }
    
    .item-value {
      flex: 1;
      font-size: 14px;
      color: #303133;
      
      &.reason-text {
        line-height: 1.8;
        padding: 12px;
        background: white;
        border-radius: 8px;
        border: 1px solid #ebeef5;
      }
    }
    
    &.reason {
      flex-direction: column;
      
      .item-label {
        margin-bottom: 8px;
      }
    }
  }
}

.ai-comparison {
  margin-top: 20px;
  padding: 16px;
  background: white;
  border-radius: 8px;
  border: 1px solid #fde2e2;
  
  .comparison-title {
    font-size: 14px;
    font-weight: 600;
    color: #f56c6c;
    margin-bottom: 12px;
  }
  
  .comparison-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 16px;
  }
  
  .comparison-item {
    .comparison-label {
      font-size: 12px;
      color: #909399;
      margin-bottom: 4px;
    }
    
    .comparison-value {
      font-size: 14px;
      font-weight: 500;
      
      &.required {
        color: #67c23a;
      }
      
      &.detected {
        color: #f56c6c;
      }
    }
  }
}

.raw-response {
  margin-top: 16px;
  
  pre {
    background: #282c34;
    color: #abb2bf;
    padding: 12px;
    border-radius: 4px;
    font-size: 12px;
    overflow-x: auto;
    max-height: 300px;
    overflow-y: auto;
  }
}
</style>
