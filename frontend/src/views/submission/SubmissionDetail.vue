<template>
  <div class="submission-detail" v-loading="loading">
    <template v-if="submission">
      <h1 class="page-title">提交 #{{ submission.id }}</h1>
      
      <!-- 基本信息 -->
      <div class="card info-card">
        <div class="info-grid">
          <div class="info-item">
            <label>状态</label>
            <StatusBadge :status="submission.status" />
          </div>
          <div class="info-item">
            <label>题目</label>
            <router-link :to="`/problem/${submission.problem_id}`">
              题目 #{{ submission.problem_id }}
            </router-link>
          </div>
          <div class="info-item">
            <label>语言</label>
            <span>{{ getLanguageLabel(submission.language) }}</span>
          </div>
          <div class="info-item">
            <label>用时</label>
            <span>{{ submission.time_used ? `${submission.time_used}ms` : '-' }}</span>
          </div>
          <div class="info-item">
            <label>内存</label>
            <span>{{ submission.memory_used ? formatMemory(submission.memory_used) : '-' }}</span>
          </div>
          <div class="info-item">
            <label>得分</label>
            <span :class="getScoreClass(submission.score)">{{ submission.score }} 分</span>
          </div>
          <div class="info-item">
            <label>提交时间</label>
            <span>{{ formatTime(submission.created_at) }}</span>
          </div>
        </div>
        
        <!-- 最终判定消息 -->
        <div v-if="submission.final_message" class="final-message">
          <el-alert :type="submission.status === 'Accepted' ? 'success' : 'warning'" :closable="false">
            {{ submission.final_message }}
          </el-alert>
        </div>
      </div>

      <!-- 测试点结果统计 -->
      <div class="card" v-if="submission.testcase_results?.length">
        <div class="section-header">
          <h3>测试点结果</h3>
          <span class="pass-rate" :class="getPassRateClass()">
            通过率: {{ passedCount }}/{{ submission.testcase_results.length }} 
            ({{ passRate }}%)
          </span>
        </div>
        <TestcaseResults :results="submission.testcase_results" />
      </div>

      <!-- AI 判题结果 -->
      <div class="card" v-if="submission.ai_judge_result?.enabled">
        <h3>🤖 AI 智能判题结果</h3>
        <AIJudgeResult :result="submission.ai_judge_result" />
      </div>

      <!-- 编译错误 -->
      <div class="card error-card" v-if="submission.compile_error">
        <h3>❌ 编译错误</h3>
        <pre class="compile-error">{{ submission.compile_error }}</pre>
      </div>

      <!-- 代码 -->
      <div class="card" v-if="submission.code">
        <div class="section-header">
          <h3>提交代码</h3>
          <span class="code-lang">{{ getLanguageLabel(submission.language) }}</span>
        </div>
        <pre class="code-block"><code>{{ submission.code }}</code></pre>
      </div>

      <!-- 没有代码时的提示 -->
      <div class="card" v-if="!submission.code && submission.user_id">
        <el-alert type="info" :closable="false">
          您只能查看自己提交的代码
        </el-alert>
      </div>
    </template>

    <!-- 轮询提示 -->
    <div v-if="isPolling" class="polling-indicator">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>正在评测中，请稍候...</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { Loading } from '@element-plus/icons-vue'
import { submissionApi } from '@/api/submission'
import StatusBadge from '@/components/problem/StatusBadge.vue'
import TestcaseResults from '@/components/submission/TestcaseResults.vue'
import AIJudgeResult from '@/components/submission/AIJudgeResult.vue'

const route = useRoute()

const loading = ref(false)
const submission = ref(null)
const isPolling = ref(false)
let pollTimer = null

const languageLabels = {
  c: 'C',
  cpp: 'C++',
  python: 'Python',
  java: 'Java',
  go: 'Go',
}

// 计算通过的测试点数量
const passedCount = computed(() => {
  if (!submission.value?.testcase_results) return 0
  return submission.value.testcase_results.filter(r => r.status === 'Accepted').length
})

// 计算通过率
const passRate = computed(() => {
  if (!submission.value?.testcase_results?.length) return 0
  return Math.round(passedCount.value / submission.value.testcase_results.length * 100)
})

function getLanguageLabel(lang) {
  return languageLabels[lang] || lang
}

function formatMemory(kb) {
  if (kb < 1024) return `${kb}KB`
  return `${(kb / 1024).toFixed(1)}MB`
}

function formatTime(time) {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN')
}

function getScoreClass(score) {
  if (score === 100) return 'score-full'
  if (score >= 60) return 'score-pass'
  return 'score-fail'
}

function getPassRateClass() {
  if (passRate.value === 100) return 'rate-full'
  if (passRate.value >= 60) return 'rate-pass'
  return 'rate-fail'
}

async function fetchSubmission() {
  try {
    const res = await submissionApi.getById(route.params.id)
    submission.value = res.data
    
    // 如果还在评测中，继续轮询
    if (res.data.status === 'Pending' || res.data.status === 'Judging') {
      isPolling.value = true
      startPolling()
    } else {
      isPolling.value = false
      stopPolling()
    }
  } catch (e) {
    console.error(e)
  }
}

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(fetchSubmission, 2000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  loading.value = true
  await fetchSubmission()
  loading.value = false
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style lang="scss" scoped>
.info-card {
  .info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 16px;
  }
  
  .info-item {
    label {
      display: block;
      font-size: 12px;
      color: #909399;
      margin-bottom: 4px;
    }
    
    span, a {
      font-size: 16px;
      font-weight: 500;
    }
    
    .score-full {
      color: #67c23a;
      font-weight: 700;
    }
    
    .score-pass {
      color: #e6a23c;
    }
    
    .score-fail {
      color: #f56c6c;
    }
  }
  
  .final-message {
    margin-top: 16px;
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  
  h3 {
    margin: 0;
  }
  
  .pass-rate {
    font-size: 14px;
    font-weight: 600;
    padding: 4px 12px;
    border-radius: 12px;
    
    &.rate-full {
      background: #f0f9eb;
      color: #67c23a;
    }
    
    &.rate-pass {
      background: #fdf6ec;
      color: #e6a23c;
    }
    
    &.rate-fail {
      background: #fef0f0;
      color: #f56c6c;
    }
  }
  
  .code-lang {
    font-size: 14px;
    color: #909399;
    background: #f5f7fa;
    padding: 4px 12px;
    border-radius: 4px;
  }
}

.card {
  h3 {
    margin-bottom: 12px;
    color: #303133;
  }
}

.error-card {
  border-left: 4px solid #f56c6c;
}

.compile-error {
  background: #fef0f0;
  color: #f56c6c;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

.code-block {
  background: #282c34;
  color: #abb2bf;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  
  code {
    font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
    font-size: 14px;
    line-height: 1.6;
  }
}

.polling-indicator {
  position: fixed;
  bottom: 20px;
  right: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: #409eff;
  color: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4);
  
  .el-icon {
    font-size: 18px;
  }
}
</style>
