<template>
  <div v-loading="loading" class="problem-detail-wrapper">
    <template v-if="problem">
      <splitpanes class="minimal-splitpanes" style="height: calc(100vh - 65px)">
        <!-- Problem Description Pane -->
        <pane min-size="30" class="left-pane">
          <div class="paper-content">
            <div class="problem-header">
              <h1 class="title">{{ problem.title }}</h1>
              <div class="meta-line">
                <span class="meta-item">时间限制: {{ problem.time_limit }}ms</span>
                <span class="meta-item">内存限制: {{ problem.memory_limit }}MB</span>
                <span class="meta-item">IO模式: 标准输入输出</span>
              </div>
              <div class="tags-line">
                <span :class="['difficulty-tag', problem.difficulty]">{{ formatDifficulty(problem.difficulty) }}</span>
                <span v-for="tag in problem.tags" :key="tag" class="minimal-tag">{{ tag }}</span>
              </div>
            </div>

            <div v-if="problem.ai_judge_config?.enabled" class="alert-box ai">
              <div class="alert-title">本题已启用 AI 智能分析</div>
              <div class="alert-content">
                <p>AI 判题系统将检查您的代码是否符合以下约束条件：</p>
                <ul>
                    <li v-if="problem.ai_judge_config.required_algorithm">
                        <strong>指定算法:</strong> {{ problem.ai_judge_config.required_algorithm }}
                    </li>
                    <li v-if="problem.ai_judge_config.required_language">
                        <strong>指定语言:</strong> {{ problem.ai_judge_config.required_language }}
                    </li>
                    <li v-if="problem.ai_judge_config.forbidden_features?.length">
                        <strong>禁用特性:</strong> {{ problem.ai_judge_config.forbidden_features.join(', ') }}
                    </li>
                    <li v-if="problem.ai_judge_config.custom_prompt">
                        <strong>额外要求:</strong> {{ problem.ai_judge_config.custom_prompt }}
                    </li>
                </ul>
              </div>
            </div>
            
            <div v-if="problem.file_io_enabled" class="alert-box warning">
              <div class="alert-title">文件读写要求</div>
              <div class="alert-content">
                请从文件 <code class="inline-code">{{ problem.file_input_name }}</code> 读取输入，并将结果输出到 <code class="inline-code">{{ problem.file_output_name }}</code>
              </div>
            </div>

            <div class="markdown-body">
              <div class="section-title">题目描述</div>
              <MarkdownPreview :content="problem.description" />
              
              <div class="section-title">输入格式</div>
              <MarkdownPreview :content="problem.input_format" />
              
              <div class="section-title">输出格式</div>
              <MarkdownPreview :content="problem.output_format" />
              
              <div class="section-title">样例数据</div>
              <div v-for="(sample, index) in problem.samples" :key="index" class="sample-box">
                <div class="sample-header">
                  <span>样例 #{{ index + 1 }}</span>
                  <div class="copy-actions">
                    <span @click="copyToClipboard(sample.input)" class="copy-btn">复制输入</span>
                    <span class="divider">|</span>
                    <span @click="copyToClipboard(sample.output)" class="copy-btn">复制输出</span>
                  </div>
                </div>
                <div class="sample-grid">
                  <div class="sample-col">
                    <div class="col-label">输入</div>
                    <pre>{{ sample.input }}</pre>
                  </div>
                  <div class="sample-col">
                    <div class="col-label">输出</div>
                    <pre>{{ sample.output }}</pre>
                  </div>
                </div>
              </div>

              <template v-if="problem.hint">
                <div class="section-title">提示</div>
                <MarkdownPreview :content="problem.hint" />
              </template>
            </div>
          </div>
        </pane>

        <!-- Code Editor Pane -->
        <pane min-size="30" class="right-pane">
          <div class="editor-container">
            <div class="editor-toolbar">
              <div class="toolbar-left">
                <span class="toolbar-label">语言</span>
                <el-select 
                  v-model="submission.language" 
                  class="language-select" 
                  size="small"
                >
                  <el-option label="C++" value="cpp" />
                  <el-option label="C" value="c" />
                  <el-option label="Python" value="python" />
                  <el-option label="Java" value="java" />
                  <el-option label="Go" value="go" />
                </el-select>

                <span class="toolbar-label" style="margin-left: 16px">字号</span>
                <el-select 
                  v-model="fontSize" 
                  class="language-select" 
                  size="small"
                  style="width: 80px"
                >
                  <el-option label="12px" :value="12" />
                  <el-option label="14px" :value="14" />
                  <el-option label="16px" :value="16" />
                  <el-option label="18px" :value="18" />
                  <el-option label="20px" :value="20" />
                </el-select>

                <span class="toolbar-label" style="margin-left: 16px">Tab Size</span>
                <el-select 
                  v-model="tabSize" 
                  class="language-select" 
                  size="small"
                  style="width: 80px"
                >
                  <el-option label="2 空格" :value="2" />
                  <el-option label="4 空格" :value="4" />
                  <el-option label="8 空格" :value="8" />
                </el-select>
              </div>
              <div class="toolbar-right">
                <el-button 
                  type="primary" 
                  :loading="submitting" 
                  @click="handleSubmit"
                  class="submit-btn"
                >
                  提交运行
                </el-button>
              </div>
            </div>

            <CodeEditor
              v-model="submission.code"
              :language="submission.language"
              :tab-size="tabSize"
              :font-size="fontSize"
              class="minimal-editor"
            />
          </div>
        </pane>
      </splitpanes>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from '@/utils/message'
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
const fontSize = ref(14)

const submission = reactive({
  language: 'cpp',
  code: '',
})

const formatDifficulty = (val) => {
  const map = { easy: '简单', medium: '中等', hard: '困难' }
  return map[val] || val
}

function fallbackCopyTextToClipboard(text) {
  var textArea = document.createElement("textarea");
  textArea.value = text;
  
  // Ensure it's not visible but part of the DOM
  textArea.style.position = "fixed";
  textArea.style.left = "-9999px";
  textArea.style.top = "0";
  
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();

  try {
    var successful = document.execCommand('copy');
    if (successful) {
      message.success('已复制');
    } else {
      message.error('复制失败');
    }
  } catch (err) {
    message.error('复制失败');
  }

  document.body.removeChild(textArea);
}

async function copyToClipboard(text) {
  if (!navigator.clipboard) {
    fallbackCopyTextToClipboard(text);
    return;
  }
  try {
    await navigator.clipboard.writeText(text);
    message.success('已复制');
  } catch (err) {
    fallbackCopyTextToClipboard(text);
  }
}

async function fetchProblem() {
  loading.value = true
  try {
    const res = await problemApi.getById(route.params.id)
    problem.value = res.data
  } catch (e) {
    message.error('加载题目失败')
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!userStore.isLoggedIn) {
    message.warning('请先登录')
    router.push({ name: 'Login', query: { redirect: route.fullPath } })
    return
  }

  if (!submission.code.trim()) {
    message.warning('代码不能为空')
    return
  }

  submitting.value = true
  try {
    const res = await submissionApi.submit({
      problem_id: parseInt(route.params.id),
      language: submission.language,
      code: submission.code,
    })
    message.success('提交成功')
    router.push(`/submission/${res.data.id}`)
  } catch (e) {
    // Error handled by interceptor
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchProblem()
})
</script>

<style lang="scss" scoped>
.problem-detail-wrapper {
  background-color: var(--color-bg);
  height: calc(100vh - 65px);
  overflow: hidden;
}

.left-pane {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  overflow-y: auto;
}

.right-pane {
  background: #1e1e1e; /* Dark theme for editor background usually matches better */
  display: flex;
  flex-direction: column;
}

.paper-content {
  padding: 40px;
  max-width: 900px;
  margin: 0 auto;
}

.problem-header {
  margin-bottom: 32px;
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 24px;
  
  .title {
    font-size: 28px;
    font-weight: 600;
    color: var(--color-text-primary);
    margin: 0 0 12px 0;
    letter-spacing: -0.01em;
  }
  
  .meta-line {
    font-family: var(--font-family-mono);
    font-size: 13px;
    color: var(--color-text-secondary);
    margin-bottom: 16px;
    
    .meta-item {
      margin-right: 24px;
    }
  }
  
  .tags-line {
    display: flex;
    gap: 8px;
  }
}

.difficulty-tag {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  
  &.easy { color: var(--color-success); background: rgba(16, 185, 129, 0.1); }
  &.medium { color: var(--color-warning); background: rgba(245, 158, 11, 0.1); }
  &.hard { color: var(--color-danger); background: rgba(239, 68, 68, 0.1); }
}

.minimal-tag {
  font-size: 12px;
  color: var(--color-text-secondary);
  background: #F3F4F6;
  padding: 2px 8px;
  border-radius: 4px;
}

.alert-box {
  padding: 16px;
  border-radius: 6px;
  margin-bottom: 24px;
  font-size: 14px;
  
  .alert-title {
    font-weight: 600;
    margin-bottom: 8px;
  }
  
  .alert-content {
      line-height: 1.6;
      ul {
          margin: 4px 0 0 20px;
          padding: 0;
      }
      li {
          margin-bottom: 4px;
      }
  }
  
  &.ai {
    background: #EEF2FF;
    color: #4338CA;
    border: 1px solid #E0E7FF;
  }
  
  &.warning {
    background: #FFFBEB;
    color: #B45309;
    border: 1px solid #FEF3C7;
  }
  
  .inline-code {
    background: rgba(0,0,0,0.05);
    padding: 2px 4px;
    border-radius: 3px;
    font-family: var(--font-family-mono);
  }
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-top: 32px;
  margin-bottom: 16px;
}

.sample-box {
  border: 1px solid var(--color-border);
  border-radius: 6px;
  margin-bottom: 20px;
  overflow: hidden;
  
  .sample-header {
    background: #F9FAFB;
    padding: 8px 16px;
    border-bottom: 1px solid var(--color-border);
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-secondary);
    display: flex;
    justify-content: space-between;
    
    .copy-actions {
      display: flex;
      gap: 8px;
      
      .copy-btn {
        cursor: pointer;
        &:hover { color: var(--color-primary); }
      }
      .divider { color: var(--color-border); }
    }
  }
  
  .sample-grid {
    display: flex;
    
    .sample-col {
      flex: 1;
      padding: 16px;
      
      &:first-child {
        border-right: 1px solid var(--color-border);
      }
      
      .col-label {
        font-size: 11px;
        text-transform: uppercase;
        color: var(--color-text-secondary);
        margin-bottom: 8px;
        letter-spacing: 0.05em;
      }
      
      pre {
        margin: 0;
        font-family: var(--font-family-mono);
        font-size: 14px;
        white-space: pre-wrap;
        word-break: break-all;
        color: var(--color-text-primary);
      }
    }
  }
}

.editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e; /* Match editor */
}

.editor-toolbar {
  height: 50px;
  background: #252526; /* VS Code header style */
  border-bottom: 1px solid #333;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  
  .toolbar-label {
    color: #9CA3AF;
    font-size: 13px;
  }
  
  .language-select {
    width: 100px;
    :deep(.el-input__wrapper),
    :deep(.el-select__wrapper) {
      background-color: #374151 !important;
      box-shadow: none !important;
      border: 1px solid #4B5563;
      
      .el-input__inner,
      .el-select__selected-item {
        color: #FFFFFF !important;
      }
    }
  }
}

.minimal-editor {
  flex: 1;
  overflow: hidden;
}

/* Splitpanes overrides */
:deep(.splitpanes__splitter) {
  background-color: var(--color-border) !important;
  width: 1px !important;
  border: none !important;
  
  &:hover {
    background-color: var(--color-primary) !important;
    width: 2px !important;
  }
  
  &::before, &::after {
    display: none !important;
  }
}
</style>
