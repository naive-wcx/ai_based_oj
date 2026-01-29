<template>
  <div class="problem-detail" v-loading="loading">
    <template v-if="problem">
      <!-- 题目信息 -->
      <div class="problem-header card">
        <div class="header-main">
          <h1>{{ problem.id }}. {{ problem.title }}</h1>
          <div class="header-tags">
            <DifficultyBadge :difficulty="problem.difficulty" />
            <el-tag v-for="tag in problem.tags" :key="tag" size="small">{{ tag }}</el-tag>
            <el-tag v-if="problem.has_accepted" type="success" size="small">已通过</el-tag>
            <span v-if="problem.ai_judge_config?.enabled" class="ai-badge">AI 判题</span>
          </div>
        </div>
        <div class="header-info">
          <span>时间限制: {{ problem.time_limit }}ms</span>
          <span>内存限制: {{ problem.memory_limit }}MB</span>
        </div>
      </div>

      <!-- AI 判题提示 -->
      <el-alert
        v-if="problem.ai_judge_config?.enabled"
        type="info"
        :closable="false"
        class="ai-alert"
      >
        <template #title>
          <div class="ai-alert-content">
            <strong>🤖 本题启用 AI 智能判题</strong>
            <p v-if="problem.ai_judge_config.required_algorithm">
              要求算法：{{ problem.ai_judge_config.required_algorithm }}
            </p>
            <p v-if="problem.ai_judge_config.required_language">
              要求语言：{{ problem.ai_judge_config.required_language }}
            </p>
            <p v-if="problem.ai_judge_config.forbidden_features?.length">
              禁止使用：{{ problem.ai_judge_config.forbidden_features.join(', ') }}
            </p>
            <p v-if="problem.ai_judge_config.custom_prompt">
              {{ problem.ai_judge_config.custom_prompt }}
            </p>
          </div>
        </template>
      </el-alert>

      <div class="problem-content">
        <!-- 左侧：题目描述 -->
        <div class="problem-desc card">
          <h3>题目描述</h3>
          <MarkdownPreview :content="problem.description" />
          
          <h3>输入格式</h3>
          <MarkdownPreview :content="problem.input_format" />
          
          <h3>输出格式</h3>
          <MarkdownPreview :content="problem.output_format" />
          
          <h3>样例</h3>
          <div v-for="(sample, index) in problem.samples" :key="index" class="sample">
            <div class="sample-item">
              <label>输入 #{{ index + 1 }}</label>
              <pre>{{ sample.input }}</pre>
            </div>
            <div class="sample-item">
              <label>输出 #{{ index + 1 }}</label>
              <pre>{{ sample.output }}</pre>
            </div>
          </div>
        </div>

        <!-- 右侧：代码提交 -->
        <div class="submit-panel card">
          <h3>提交代码</h3>
          
          <el-select v-model="submission.language" style="width: 100%; margin-bottom: 12px">
            <el-option label="C++" value="cpp" />
            <el-option label="C" value="c" />
            <el-option label="Python" value="python" />
            <el-option label="Java" value="java" />
            <el-option label="Go" value="go" />
          </el-select>
          
          <div class="code-editor">
            <el-input
              v-model="submission.code"
              type="textarea"
              :rows="20"
              placeholder="在此输入代码..."
            />
          </div>
          
          <el-button
            type="primary"
            :loading="submitting"
            @click="handleSubmit"
            style="width: 100%; margin-top: 12px"
          >
            提交
          </el-button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { problemApi } from '@/api/problem'
import { submissionApi } from '@/api/submission'
import { useUserStore } from '@/stores/user'
import DifficultyBadge from '@/components/problem/DifficultyBadge.vue'
import MarkdownPreview from '@/components/common/MarkdownPreview.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const submitting = ref(false)
const problem = ref(null)

const submission = reactive({
  language: 'cpp',
  code: '',
})

async function fetchProblem() {
  loading.value = true
  try {
    const res = await problemApi.getById(route.params.id)
    problem.value = res.data
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push({ name: 'Login', query: { redirect: route.fullPath } })
    return
  }

  if (!submission.code.trim()) {
    ElMessage.warning('请输入代码')
    return
  }

  submitting.value = true
  try {
    const res = await submissionApi.submit({
      problem_id: parseInt(route.params.id),
      language: submission.language,
      code: submission.code,
    })
    ElMessage.success('提交成功')
    router.push(`/submission/${res.data.id}`)
  } catch (e) {
    console.error(e)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchProblem()
})
</script>

<style lang="scss" scoped>
.problem-header {
  .header-main {
    h1 {
      font-size: 24px;
      margin-bottom: 12px;
    }
    
    .header-tags {
      display: flex;
      gap: 8px;
      align-items: center;
      flex-wrap: wrap;
    }
  }
  
  .header-info {
    margin-top: 12px;
    color: #909399;
    font-size: 14px;
    
    span {
      margin-right: 20px;
    }
  }
}

.ai-alert {
  margin-bottom: 20px;
  
  .ai-alert-content {
    p {
      margin: 4px 0;
      font-size: 14px;
    }
  }
}

.problem-content {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 20px;
  
  @media (max-width: 1200px) {
    grid-template-columns: 1fr;
  }
}

.problem-desc {
  h3 {
    font-size: 18px;
    margin: 24px 0 12px;
    color: #303133;
    
    &:first-child {
      margin-top: 0;
    }
  }
  
  .sample {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    margin-bottom: 16px;
    
    @media (max-width: 768px) {
      grid-template-columns: 1fr;
    }
  }
  
  .sample-item {
    label {
      display: block;
      font-weight: 500;
      margin-bottom: 8px;
      color: #909399;
    }
    
    pre {
      margin: 0;
      padding: 12px;
      background: #f5f7fa;
      border-radius: 4px;
      font-size: 14px;
    }
  }
}

.submit-panel {
  position: sticky;
  top: 80px;
  height: fit-content;
  
  h3 {
    margin-bottom: 16px;
  }
  
  .code-editor {
    :deep(.el-textarea__inner) {
      font-family: 'Fira Code', 'Monaco', monospace;
      font-size: 14px;
    }
  }
}
</style>
