<template>
  <div v-loading="loading" class="problem-detail-container">
    <template v-if="problem">
      <splitpanes class="default-theme" style="height: calc(100vh - 100px)">
        <pane min-size="30" class="problem-pane">
          <el-card shadow="never" class="problem-card">
            <template #header>
              <div class="problem-header">
                <h1 class="title">{{ problem.id }}. {{ problem.title }}</h1>
                <div class="info">
                  <span>时间限制: {{ problem.time_limit }}ms</span>
                  <span>内存限制: {{ problem.memory_limit }}MB</span>
                </div>
              </div>
            </template>
            <div class="tags">
              <el-tag :type="getDifficultyTag(problem.difficulty).type" size="small">{{
                getDifficultyTag(problem.difficulty).label
              }}</el-tag>
              <el-tag
                v-for="tag in problem.tags"
                :key="tag"
                size="small"
                effect="plain"
                >{{ tag }}</el-tag
              >
              <el-tag v-if="problem.has_accepted" type="success" size="small"
                >已通过</el-tag
              >
            </div>

            <div v-if="problem.ai_judge_config?.enabled" class="ai-section">
              <el-alert type="info" :closable="false" show-icon>
                <template #title>
                  <strong>本题已启用 AI 智能分析</strong>
                </template>
                <div class="ai-requirements">
                  <p>请注意，AI 将对你的代码进行以下维度的评估：</p>
                  <ul>
                    <li v-if="problem.ai_judge_config.required_algorithm">
                      <strong>算法要求:</strong> {{ problem.ai_judge_config.required_algorithm }}
                    </li>
                    <li v-if="problem.ai_judge_config.required_language">
                      <strong>语言要求:</strong> {{ problem.ai_judge_config.required_language }}
                    </li>
                    <li v-if="problem.ai_judge_config.forbidden_features?.length">
                      <strong>禁用特性:</strong>
                      <el-tag
                        v-for="feature in problem.ai_judge_config.forbidden_features"
                        :key="feature"
                        type="danger"
                        effect="light"
                        size="small"
                        class="ai-feature-tag"
                        >{{ feature }}</el-tag
                      >
                    </li>
                    <li v-if="problem.ai_judge_config.custom_prompt">
                      <strong>自定义提示:</strong> {{ problem.ai_judge_config.custom_prompt }}
                    </li>
                  </ul>
                </div>
              </el-alert>
            </div>

            <div class="content-section">
              <h3>题目描述</h3>
              <MarkdownPreview :content="problem.description" />
            </div>
            <div class="content-section">
              <h3>输入格式</h3>
              <MarkdownPreview :content="problem.input_format" />
            </div>
            <div class="content-section">
              <h3>输出格式</h3>
              <MarkdownPreview :content="problem.output_format" />
            </div>
            <div class="content-section">
              <h3>样例</h3>
              <div
                v-for="(sample, index) in problem.samples"
                :key="index"
                class="sample"
              >
                <el-card shadow="never" class="sample-card">
                  <template #header>
                    <div class="sample-header">
                      <span>样例 #{{ index + 1 }}</span>
                    </div>
                  </template>
                  <el-row :gutter="16">
                    <el-col :md="12" :sm="24">
                      <div class="sample-io">
                        <div class="sample-io-header">
                          <label>输入</label>
                          <el-button
                            :icon="CopyDocument"
                            text
                            size="small"
                            @click="copyToClipboard(sample.input)"
                          />
                        </div>
                        <pre>{{ sample.input }}</pre>
                      </div>
                    </el-col>
                    <el-col :md="12" :sm="24">
                       <div class="sample-io">
                         <div class="sample-io-header">
                          <label>输出</label>
                          <el-button
                            :icon="CopyDocument"
                            text
                            size="small"
                            @click="copyToClipboard(sample.output)"
                          />
                        </div>
                        <pre>{{ sample.output }}</pre>
                      </div>
                    </el-col>
                  </el-row>
                </el-card>
              </div>
            </div>
          </el-card>
        </pane>
        <pane min-size="30" class="submit-pane">
          <el-card shadow="never" class="submit-card">
            <template #header>
              <div class="submit-header">
                <div class="submit-header-left">
                  <el-icon><EditPen /></el-icon>
                  <span>提交代码</span>
                </div>
                 <div class="submit-header-right">
                  <span class="tab-size-label">Tab长度</span>
                  <el-select v-model="tabSize" size="small" style="width: 70px;">
                    <el-option :value="2" />
                    <el-option :value="4" />
                    <el-option :value="8" />
                  </el-select>
                </div>
              </div>
            </template>
            <el-select
              v-model="submission.language"
              placeholder="选择语言"
              style="margin-bottom: 16px"
            >
              <el-option label="C++" value="cpp" />
              <el-option label="C" value="c" />
              <el-option label="Python" value="python" />
              <el-option label="Java" value="java" />
              <el-option label="Go" value="go" />
            </el-select>

            <CodeEditor
              v-model="submission.code"
              :language="submission.language"
              :tab-size="tabSize"
              class="code-editor-component"
            />

            <el-button
              type="primary"
              :loading="submitting"
              @click="handleSubmit"
              style="width: 100%; margin-top: 16px"
              size="large"
            >
              提交
            </el-button>
          </el-card>
        </pane>
      </splitpanes>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { EditPen, CopyDocument } from '@element-plus/icons-vue'
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'
import { problemApi } from '@/api/problem'
import { submissionApi } from '@/api/submission'
import { useUserStore } from '@/stores/user'
import MarkdownPreview from '@/components/common/MarkdownPreview.vue'
import CodeEditor from '@/components/common/CodeEditor.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(true)
const submitting = ref(false)
const problem = ref(null)
const tabSize = ref(4)

const submission = reactive({
  language: 'cpp',
  code: '',
})

const getDifficultyTag = (difficulty) => {
  const settings = {
    easy: { type: 'success', label: '简单' },
    medium: { type: 'warning', label: '中等' },
    hard: { type: 'danger', label: '困难' },
  }
  return settings[difficulty] || { type: 'info', label: difficulty }
}

async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success('已复制到剪切板');
  } catch (err) {
    ElMessage.error('复制失败');
    console.error('Failed to copy: ', err);
  }
}

async function fetchProblem() {
  loading.value = true
  try {
    const res = await problemApi.getById(route.params.id)
    problem.value = res.data
  } catch (e) {
    ElMessage.error('题目加载失败')
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录再提交')
    router.push({ name: 'Login', query: { redirect: route.fullPath } })
    return
  }

  if (!submission.code.trim()) {
    ElMessage.warning('提交的代码不能为空')
    return
  }

  submitting.value = true
  try {
    const res = await submissionApi.submit({
      problem_id: parseInt(route.params.id),
      language: submission.language,
      code: submission.code,
    })
    ElMessage.success('提交成功！正在跳转到评测记录...')
    router.push(`/submission/${res.data.id}`)
  } catch (e) {
    // Error message is likely handled by the request interceptor
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchProblem()
})
</script>

<style lang="scss" scoped>
.problem-detail-container {
  height: calc(100vh - 60px); // Full viewport height minus navbar
  overflow: hidden;
  background-color: #f7f8fa;
  padding: 10px;
}

.problem-pane {
  overflow-y: auto;
}

.submit-pane {
  display: flex;
  flex-direction: column;

  .submit-card {
    border: none;
    display: flex;
    flex-direction: column;
    height: 100%;

    :deep(.el-card__body) {
      display: flex;
      flex-direction: column;
      flex-grow: 1;
      padding: 20px !important;
    }
  }
}

.code-editor-component {
  flex-grow: 1;
  border-radius: 4px;
  overflow: hidden;
}

.problem-card {
  border: none;
  :deep(.el-card__body) {
    padding: 20px !important;
  }
}

.problem-header {
  .title {
    font-size: 26px;
    font-weight: 600;
    margin: 0 0 12px;
    color: #303133;
  }
  .info {
    font-size: 14px;
    color: #909399;
    span {
      margin-right: 20px;
    }
  }
}

.tags {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 24px;
}

.ai-section {
  margin-bottom: 24px;
  .ai-requirements {
    margin-top: 8px;
    font-size: 14px;
    p {
      margin: 0 0 8px;
    }
    ul {
      padding-left: 20px;
      margin: 0;
    }
    li {
      margin-bottom: 4px;
    }
    .ai-feature-tag {
      margin: 0 4px;
    }
  }
}

.content-section {
  margin-bottom: 24px;
  h3 {
    font-size: 20px;
    margin: 0 0 16px;
    color: #303133;
    font-weight: 600;
    border-left: 4px solid var(--el-color-primary);
    padding-left: 12px;
  }
}

.sample {
  margin-bottom: 16px;
}
.sample-card {
  background-color: #fafafa;
  border: 1px solid #e4e7ed;
}

.sample-header {
  font-weight: 500;
  color: #606266;
}

.sample-io-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.sample-io {
  label {
    font-weight: 500;
    color: #909399;
    font-size: 14px;
  }

  pre {
    margin: 0;
    padding: 12px;
    background: #ffffff;
    border: 1px solid #e4e7ed;
    border-radius: 4px;
    font-size: 14px;
    font-family: 'Fira Code', 'Monaco', monospace;
    white-space: pre-wrap;
    word-break: break-all;
  }
}

.submit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.submit-header-left, .submit-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tab-size-label {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
}

/* Splitpanes custom theme */
.splitpanes.default-theme .splitpanes__splitter {
  background-color: #f0f2f5;
  border-left: 1px solid #e4e7ed;
  border-right: 1px solid #e4e7ed;
  box-sizing: border-box;
  position: relative;
  width: 7px;
}

.splitpanes.default-theme .splitpanes__splitter:before {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 2px;
  height: 30px;
  background-color: #c0c4cc;
  border-radius: 2px;
}

.splitpanes.default-theme .splitpanes__splitter:hover:before {
  background-color: var(--el-color-primary);
}

@media (max-width: 768px) {
  .splitpanes--vertical > .splitpanes__splitter {
    width: 100% !important;
    height: 7px !important;
  }
  .splitpanes.default-theme .splitpanes__splitter:before {
    width: 30px;
    height: 2px;
  }
}
</style>
