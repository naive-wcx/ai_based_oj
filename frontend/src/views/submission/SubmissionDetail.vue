<template>
  <div class="submission-detail-wrapper" v-loading="loading">
    <div class="container" v-if="submission">
      <!-- 1. 页头：标题与状态 -->
      <div class="detail-header">
        <div class="header-left">
          <router-link :to="`/problem/${submission.problem_id}`" class="back-link">
            ← 返回题目
          </router-link>
          <h1 class="page-title">
            提交记录 <span class="sub-id">#{{ submission.id }}</span>
          </h1>
        </div>
        <div class="header-right">
           <div class="status-badge" :class="getStatusClass(submission.status)">
             {{ statusMap[submission.status]?.label || submission.status }}
           </div>
        </div>
      </div>

      <!-- 2. 核心指标仪表盘 -->
      <div class="stats-dashboard">
        <div class="stat-card">
          <div class="stat-label">得分</div>
          <div class="stat-value score-value" :class="getScoreClass(submission.score)">
            {{ submission.score != null ? submission.score : '-' }}
          </div>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-card">
          <div class="stat-label">运行时间</div>
          <div class="stat-value">{{ submission.time_used != null ? submission.time_used + ' ms' : '-' }}</div>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-card">
          <div class="stat-label">内存占用</div>
          <div class="stat-value">{{ formatMemory(submission.memory_used) }}</div>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-card">
          <div class="stat-label">语言</div>
          <div class="stat-value">{{ languageLabels[submission.language] || submission.language }}</div>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-card">
          <div class="stat-label">提交者</div>
          <div class="stat-value">{{ submission.username }}</div>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-card">
          <div class="stat-label">提交时间</div>
          <div class="stat-value time-value">{{ formatTime(submission.created_at) }}</div>
        </div>
      </div>

      <!-- 3. 最终评测消息 -->
      <div class="message-banner" v-if="submission.final_message" :class="submission.status === 'Accepted' ? 'success' : 'warning'">
        <span class="icon">{{ submission.status === 'Accepted' ? '🎉' : '⚠️' }}</span>
        {{ submission.final_message }}
      </div>

      <!-- 4. 测试点结果 -->
      <div class="section-block" v-if="submission.testcase_results?.length">
        <h3 class="section-title">测试点详情</h3>
        <div class="testcases-container">
          <TestcaseResults :results="submission.testcase_results" />
        </div>
      </div>

      <!-- 5. 编译错误 -->
      <div class="section-block error-block" v-if="submission.compile_error">
         <h3 class="section-title text-danger">编译错误</h3>
         <pre class="error-content">{{ submission.compile_error }}</pre>
      </div>

      <!-- 6. AI 智能分析 -->
      <div class="section-block" v-if="submission.ai_judge_result?.enabled">
        <h3 class="section-title">
          AI 智能分析
        </h3>
        <AIJudgeResult :result="submission.ai_judge_result" />
      </div>

      <!-- 7. 源代码 -->
      <div class="section-block">
        <h3 class="section-title">源代码</h3>
        <div class="code-wrapper">
          <CodeEditor
            :model-value="submission.code"
            :language="submission.language"
            :readonly="true"
            style="min-height: 300px; font-family: var(--font-mono);"
          />
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { submissionApi } from '@/api/submission'
import TestcaseResults from '@/components/submission/TestcaseResults.vue'
import AIJudgeResult from '@/components/submission/AIJudgeResult.vue'
import CodeEditor from '@/components/common/CodeEditor.vue'

const route = useRoute()
const loading = ref(true)
const submission = ref(null)
let pollTimer = null

const languageLabels = {
  c: 'C',
  cpp: 'C++',
  python: 'Python',
  java: 'Java',
  go: 'Go',
}

const statusMap = {
  'Pending': { label: '等待中', class: 'waiting' },
  'Judging': { label: '评测中', class: 'judging' },
  'Accepted': { label: 'Accepted', class: 'ac' },
  'Submitted': { label: '已提交', class: 'waiting' },
  'Wrong Answer': { label: 'Wrong Answer', class: 'wa' },
  'Time Limit Exceeded': { label: 'Time Limit Exceeded', class: 'tle' },
  'Memory Limit Exceeded': { label: 'Memory Limit Exceeded', class: 'mle' },
  'Runtime Error': { label: 'Runtime Error', class: 're' },
  'Compile Error': { label: 'Compile Error', class: 'ce' },
  'System Error': { label: 'System Error', class: 'uqe' },
}

function getStatusClass(status) {
  return statusMap[status]?.class || 'waiting'
}

function formatMemory(kb) {
  if (kb == null) return '-'
  if (kb < 1024) return `${kb} KB`
  return `${(kb / 1024).toFixed(1)} MB`
}

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

function getScoreClass(score) {
  if (score == null) return ''
  if (score === 100) return 'text-success'
  if (score >= 60) return 'text-warning'
  return 'text-danger'
}

async function fetchSubmission() {
  try {
    const res = await submissionApi.getById(route.params.id)
    submission.value = res.data
    
    if (res.data.status === 'Pending' || res.data.status === 'Judging') {
      startPolling()
    } else {
      stopPolling()
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
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
  await fetchSubmission()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style lang="scss" scoped>
.submission-detail-wrapper {
  padding: 40px 0;
  background-color: var(--swiss-bg-base);
  min-height: 100vh;
}

/* Header */
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--swiss-border-light);
}

.back-link {
  font-size: 14px;
  color: var(--swiss-text-secondary);
  margin-bottom: 8px;
  display: block;
  font-weight: 500;
  transition: color 0.2s;
  &:hover { color: var(--swiss-primary); }
}

.page-title {
  font-size: 32px;
  margin: 0;
  color: var(--swiss-text-main);
  letter-spacing: -0.02em;
  
  .sub-id {
    color: var(--swiss-text-secondary);
    font-weight: 400;
    font-size: 24px;
    margin-left: 8px;
  }
}

.status-badge {
  font-size: 15px;
  font-weight: 700;
  padding: 8px 20px;
  border-radius: var(--radius-xs);
  letter-spacing: 0.02em;
  color: #fff;
  
  &.ac { background-color: var(--status-ac); }
  &.wa { background-color: var(--status-wa); }
  &.tle { background-color: var(--status-tle); }
  &.mle { background-color: var(--status-mle); }
  &.re { background-color: var(--status-re); }
  &.ce { background-color: var(--status-ce); color: #fff; /* Yellow bg, white text might be hard, but let's stick to simple badge */ }
  &.uqe { background-color: var(--status-uqe); }
  &.waiting { background-color: var(--status-waiting); color: #fff; }
  &.judging { background-color: var(--status-judging); }
}

/* Stats Dashboard */
.stats-dashboard {
  display: flex;
  align-items: stretch;
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 24px;
  margin-bottom: 30px;
  flex-wrap: wrap;
  gap: 20px;
}

.stat-card {
  flex: 1;
  min-width: 120px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.stat-divider {
  width: 1px;
  background: var(--swiss-border-light);
  margin: 4px 0;
}

.stat-label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--swiss-text-secondary);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--swiss-text-main);
  
  &.score-value { font-size: 24px; }
  
  &.time-value {
    font-size: 13px;
    font-weight: 400;
    color: var(--swiss-text-secondary);
    white-space: nowrap;
  }
}

.text-success { color: var(--status-ac); }
.text-warning { color: var(--swiss-warning); }
.text-danger { color: var(--status-wa); }

/* Message Banner */
.message-banner {
  padding: 16px 20px;
  border-radius: var(--radius-xs);
  margin-bottom: 30px;
  font-size: 15px;
  display: flex;
  align-items: center;
  gap: 12px;
  
  &.success { background: rgba(82, 196, 26, 0.1); color: var(--status-ac); border: 1px solid rgba(82, 196, 26, 0.2); }
  &.warning { background: rgba(231, 76, 60, 0.1); color: var(--status-wa); border: 1px solid rgba(231, 76, 60, 0.2); }
  
  .icon { font-size: 18px; }
}

/* Common Section */
.section-block {
  margin-bottom: 40px;
}

.section-title {
  font-size: 18px;
  margin-bottom: 16px;
  color: var(--swiss-text-main);
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 12px;
  
  &.text-danger { color: var(--status-ce); } /* Compile Error Title Color */
}

.testcases-container {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 20px;
}

.code-wrapper {
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  overflow: hidden;
  max-width: 100%;
}

.error-block {
  .error-content {
    background: #fff0f0;
    color: var(--swiss-danger);
    padding: 20px;
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: 13px;
    white-space: pre-wrap;
    word-break: break-all;
    border: 1px solid rgba(255, 59, 48, 0.2);
  }
}

@media (max-width: 768px) {
  .stats-dashboard {
    flex-direction: column;
    gap: 16px;
  }
  
  .stat-divider {
    display: none;
  }
  
  .stat-card {
    flex-direction: row;
    align-items: center;
    border-bottom: 1px solid var(--swiss-border-light);
    padding-bottom: 12px;
    
    &:last-child { border-bottom: none; padding-bottom: 0; }
  }
  
  .stat-label { margin-bottom: 0; }
}
</style>