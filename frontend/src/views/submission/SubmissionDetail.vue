<template>
  <div class="submission-detail-container" v-loading="loading">
    <template v-if="submission">
      <div class="page-header">
        <h1 class="page-title">提交记录 #{{ submission.id }}</h1>
        <div v-if="isPolling" class="polling-indicator">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>正在评测</span>
        </div>
      </div>

      <!-- 基本信息 -->
      <el-card shadow="never" class="info-card">
        <el-descriptions :column="4" border>
          <el-descriptions-item label="状态" label-align="center" align="center">
            <el-tag :type="getStatusTagType(submission.status)" effect="light">
              {{ statusMap[submission.status]?.label || submission.status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="题目" label-align="center" align="center">
            <router-link :to="`/problem/${submission.problem_id}`" class="problem-link">
              {{ submission.problem_id }}. {{ submission.problem_title }}
            </router-link>
          </el-descriptions-item>
          <el-descriptions-item label="用户" label-align="center" align="center">
            {{ submission.username }}
          </el-descriptions-item>
          <el-descriptions-item label="语言" label-align="center" align="center">
            {{ languageLabels[submission.language] || submission.language }}
          </el-descriptions-item>
           <el-descriptions-item label="用时" label-align="center" align="center">
            <span v-if="submission.time_used != null">{{ submission.time_used }}ms</span>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="内存" label-align="center" align="center">
            <span v-if="submission.memory_used != null">{{ formatMemory(submission.memory_used) }}</span>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="分数" label-align="center" align="center">
             <span :class="getScoreClass(submission.score)">
               {{ submission.score != null ? submission.score : '-' }}
             </span>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间" label-align="center" align="center">
            {{ formatTime(submission.created_at) }}
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="submission.final_message" class="final-message">
          <el-alert :type="submission.status === 'Accepted' ? 'success' : 'warning'" :closable="false">
            {{ submission.final_message }}
          </el-alert>
        </div>
      </el-card>

      <!-- 测试点结果 -->
      <el-card shadow="never" class="card-section" v-if="submission.testcase_results?.length">
        <template #header><h3>测试点结果</h3></template>
        <TestcaseResults :results="submission.testcase_results" />
      </el-card>

      <!-- AI 判题结果 -->
      <el-card shadow="never" class="card-section" v-if="submission.ai_judge_result?.enabled">
        <template #header><h3>🤖 AI 智能判题结果</h3></template>
        <AIJudgeResult :result="submission.ai_judge_result" />
      </el-card>

      <!-- 编译错误 -->
      <el-card shadow="never" class="card-section" v-if="submission.compile_error">
        <template #header><h3>❌ 编译错误</h3></template>
        <pre class="compile-error">{{ submission.compile_error }}</pre>
      </el-card>

      <!-- 代码 -->
      <el-card shadow="never" class="card-section" v-if="submission.code">
        <template #header><h3>提交代码</h3></template>
        <CodeEditor
          :model-value="submission.code"
          :language="submission.language"
          :readonly="true"
          style="min-height: 300px"
        />
      </el-card>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { Loading } from '@element-plus/icons-vue'
import { submissionApi } from '@/api/submission'
import TestcaseResults from '@/components/submission/TestcaseResults.vue'
import AIJudgeResult from '@/components/submission/AIJudgeResult.vue'
import CodeEditor from '@/components/common/CodeEditor.vue'

const route = useRoute()
const loading = ref(true)
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

const statusMap = {
  'Pending': { label: '等待中', type: 'info' },
  'Judging': { label: '评测中', type: 'primary' },
  'Accepted': { label: '通过', type: 'success' },
  'Wrong Answer': { label: '答案错误', type: 'danger' },
  'Time Limit Exceeded': { label: '超时', type: 'warning' },
  'Memory Limit Exceeded': { label: '内存超限', type: 'warning' },
  'Runtime Error': { label: '运行错误', type: 'danger' },
  'Compile Error': { label: '编译错误', type: 'danger' },
  'System Error': { label: '系统错误', type: 'danger' },
}

const getStatusTagType = (status) => {
  return statusMap[status]?.type || 'info'
}

function formatMemory(kb) {
  if (kb == null) return '-'
  if (kb < 1024) return `${kb} KB`
  return `${(kb / 1024).toFixed(1)} MB`
}

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', { hour12: false })
}

function getScoreClass(score) {
  if (score == null) return ''
  if (score === 100) return 'score-full'
  if (score >= 60) return 'score-pass'
  return 'score-fail'
}

async function fetchSubmission() {
  try {
    const res = await submissionApi.getById(route.params.id)
    submission.value = res.data
    
    if (res.data.status === 'Pending' || res.data.status === 'Judging') {
      isPolling.value = true
      startPolling()
    } else {
      isPolling.value = false
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
.submission-detail-container {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.polling-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--el-color-primary);
}

.card-section, .info-card {
  margin-bottom: 20px;
  border: none;
  background-color: #ffffff;
}

.final-message {
  margin-top: 20px;
}

.problem-link {
  color: var(--el-text-color-primary);
  text-decoration: none;
  &:hover {
    color: var(--el-color-primary);
  }
}

.score-full {
  color: var(--el-color-success);
  font-weight: bold;
}

.score-pass {
  color: var(--el-color-warning);
}

.score-fail {
  color: var(--el-color-danger);
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
</style>
