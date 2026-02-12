<template>
  <div class="problem-edit-wrapper">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">{{ isEdit ? '编辑题目' : '创建题目' }}</h1>
        <div class="actions">
          <el-button @click="$router.back()">取消</el-button>
        </div>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        v-loading="loading"
        class="edit-form"
      >
        <!-- 1. 基础信息 -->
        <div class="form-section">
          <div class="section-header">
            <h3>基础信息</h3>
          </div>
          <el-card shadow="never" class="form-card">
            <el-form-item label="题目标题" prop="title">
              <el-input v-model="form.title" placeholder="请输入清晰的题目标题" size="large" />
            </el-form-item>
            
            <el-row :gutter="24">
              <el-col :span="8">
                <el-form-item label="难度">
                  <el-select v-model="form.difficulty" style="width: 100%">
                    <el-option label="简单" value="easy" />
                    <el-option label="中等" value="medium" />
                    <el-option label="困难" value="hard" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="16">
                <el-form-item label="标签">
                  <el-select
                    v-model="form.tags"
                    multiple
                    filterable
                    allow-create
                    default-first-option
                    placeholder="输入或选择标签"
                    style="width: 100%"
                  >
                    <el-option v-for="tag in commonTags" :key="tag" :label="tag" :value="tag" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="24">
              <el-col :span="12">
                <el-form-item label="可见性">
                  <div class="switch-wrapper">
                    <el-switch v-model="form.is_public" active-text="公开" inactive-text="隐藏" />
                    <span class="hint-text">隐藏题目仅对管理员或正在进行的比赛可见</span>
                  </div>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="文件 IO">
                  <div class="switch-wrapper">
                    <el-switch v-model="form.file_io_enabled" @change="handleFileIOToggle" />
                    <div class="file-inputs" v-if="form.file_io_enabled">
                      <el-input v-model="form.file_input_name" placeholder="input.in" size="small" />
                      <span>→</span>
                      <el-input v-model="form.file_output_name" placeholder="output.out" size="small" />
                    </div>
                  </div>
                </el-form-item>
              </el-col>
            </el-row>
          </el-card>
        </div>

        <!-- 2. 题目内容 -->
        <div class="form-section">
          <div class="section-header">
            <h3>题目内容</h3>
            <span class="sub-text">左侧编辑 Markdown，右侧实时预览</span>
          </div>
          
          <el-card shadow="never" class="form-card content-card">
            <el-form-item label="题目描述" prop="description">
              <div class="md-row big">
                <el-input
                  v-model="form.description"
                  type="textarea"
                  placeholder="在此输入题目描述..."
                  class="md-input"
                  resize="none"
                />
                <div class="md-preview">
                  <MarkdownPreview :content="form.description" />
                </div>
              </div>
            </el-form-item>
            
            <el-form-item label="输入格式">
              <div class="md-row">
                <el-input
                  v-model="form.input_format"
                  type="textarea"
                  placeholder="输入格式说明..."
                  class="md-input"
                  resize="none"
                />
                <div class="md-preview">
                  <MarkdownPreview :content="form.input_format" />
                </div>
              </div>
            </el-form-item>

            <el-form-item label="输出格式">
              <div class="md-row">
                <el-input
                  v-model="form.output_format"
                  type="textarea"
                  placeholder="输出格式说明..."
                  class="md-input"
                  resize="none"
                />
                <div class="md-preview">
                  <MarkdownPreview :content="form.output_format" />
                </div>
              </div>
            </el-form-item>

            <el-form-item label="提示 / 数据范围 / 样例解释">
              <div class="md-row">
                <el-input
                  v-model="form.hint"
                  type="textarea"
                  placeholder="可选：在此补充数据范围或提示..."
                  class="md-input"
                  resize="none"
                />
                <div class="md-preview">
                  <MarkdownPreview :content="form.hint" />
                </div>
              </div>
            </el-form-item>
          </el-card>
        </div>

        <!-- 3. 样例数据 -->
        <div class="form-section">
          <div class="section-header">
            <h3>样例数据</h3>
          </div>
          <el-card shadow="never" class="form-card">
            <div v-for="(sample, index) in form.samples" :key="index" class="sample-row">
              <div class="sample-header">
                <span>样例 #{{ index + 1 }}</span>
                <el-button type="danger" link @click="removeSample(index)">删除</el-button>
              </div>
              <div class="sample-inputs">
                <el-input
                  v-model="sample.input"
                  type="textarea"
                  :rows="4"
                  placeholder="输入数据"
                  resize="vertical"
                />
                <el-input
                  v-model="sample.output"
                  type="textarea"
                  :rows="4"
                  placeholder="输出数据"
                  resize="vertical"
                />
              </div>
            </div>
            <el-button class="add-btn" @click="addSample" plain>+ 添加样例</el-button>
          </el-card>
        </div>

        <!-- 4. 评测配置 -->
        <div class="form-section">
          <div class="section-header">
            <h3>评测配置</h3>
          </div>
          <el-card shadow="never" class="form-card">
            <el-row :gutter="24">
	              <el-col :span="12">
	                <el-form-item label="时间限制">
	                  <div class="limit-input">
	                    <el-input-number v-model="form.time_limit" :min="100" :max="10000" :step="100" style="width: 100%" />
	                    <div class="unit-text">ms</div>
	                  </div>
	                </el-form-item>
	              </el-col>
	              <el-col :span="12">
	                <el-form-item label="内存限制">
	                  <div class="limit-input">
	                    <el-input-number v-model="form.memory_limit" :min="16" :max="1024" :step="16" style="width: 100%" />
	                    <div class="unit-text">MB</div>
	                  </div>
	                </el-form-item>
	              </el-col>
            </el-row>

            <el-divider />
            
            <div class="ai-config">
              <div class="ai-header">
                <el-checkbox v-model="form.ai_judge_config.enabled">启用 AI 智能辅助判题</el-checkbox>
              </div>

              <div class="ai-body" v-if="form.ai_judge_config.enabled">
                <el-row :gutter="24">
                  <el-col :span="12">
                    <el-form-item label="要求算法">
                      <el-input v-model="form.ai_judge_config.required_algorithm" placeholder="例如：动态规划" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="要求语言">
                      <el-select v-model="form.ai_judge_config.required_language" clearable style="width: 100%">
                        <el-option label="不限" value="" />
                        <el-option label="C++" value="C++" />
                        <el-option label="Python" value="Python" />
                        <el-option label="Java" value="Java" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
                
                <el-form-item label="禁止特性">
                  <el-select
                    v-model="form.ai_judge_config.forbidden_features"
                    multiple
                    filterable
                    allow-create
                    placeholder="例如：STL sort"
                    style="width: 100%"
                  >
                    <el-option label="STL sort" value="STL sort" />
                    <el-option label="递归" value="递归" />
                  </el-select>
                </el-form-item>
                
                <el-form-item label="自定义 Prompt">
                  <el-input
                    v-model="form.ai_judge_config.custom_prompt"
                    type="textarea"
                    :rows="2"
                    placeholder="给 AI 的额外指令"
                  />
                </el-form-item>
              </div>
            </div>
          </el-card>
        </div>

        <!-- 重点：保存操作栏 -->
        <div class="form-actions-bar sticky-actions">
          <el-button type="primary" size="large" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? '保存所有修改' : '立即创建题目' }}
          </el-button>
        </div>

        <!-- 5. 测试数据 (仅编辑时) -->
        <div class="form-section" v-if="isEdit">
          <div class="section-header">
            <h3>测试点管理</h3>
            <span class="sub-text">上传判题所需的测试用例文件</span>
          </div>
          <el-card shadow="never" class="form-card">
            <el-tabs type="card" class="testcase-tabs">
	              <el-tab-pane label="文件上传">
	                <div class="upload-row">
	                  <div class="upload-group">
	                    <el-upload
	                      ref="inputUploadRef"
	                      action=""
	                      :auto-upload="false"
	                      :on-change="handleInputFileChange"
	                      :show-file-list="false"
	                    >
	                      <el-button :type="inputFile ? 'success' : 'default'">
	                        {{ inputFile ? inputFile.name : '选择 Input (.in)' }}
	                      </el-button>
                    </el-upload>
                    <span class="plus">+</span>
	                    <el-upload
	                      ref="outputUploadRef"
	                      action=""
	                      :auto-upload="false"
	                      :on-change="handleOutputFileChange"
	                      :show-file-list="false"
	                    >
	                      <el-button :type="outputFile ? 'success' : 'default'">
	                        {{ outputFile ? outputFile.name : '选择 Output (.out)' }}
	                      </el-button>
                    </el-upload>
                  </div>
                  
                  <div class="score-input">
                    <span>分值:</span>
                    <el-input-number v-model="testcaseScore" :min="1" :max="100" style="width: 100px" />
                  </div>
                  
	                  <el-button type="primary" @click="uploadTestcase" :loading="uploadingTestcase" :disabled="!inputFile || !outputFile">
	                    上传
	                  </el-button>
	                </div>
	                <div class="upload-progress" v-if="uploadingTestcase || testcaseProgress > 0">
	                  <el-progress
	                    :percentage="testcaseProgress"
	                    :status="testcaseProgressStatus"
	                    :stroke-width="10"
	                  />
	                </div>
	              </el-tab-pane>
	              
	              <el-tab-pane label="Zip 批量上传">
	                <div class="upload-row">
	                  <el-upload
	                    ref="zipUploadRef"
	                    action=""
	                    :auto-upload="false"
	                    :on-change="handleZipFileChange"
	                    :show-file-list="false"
	                  >
	                    <el-button :type="zipFile ? 'success' : 'warning'" plain>
	                      {{ zipFile ? zipFile.name : '选择 Zip 压缩包' }}
                    </el-button>
                  </el-upload>
                  <el-button type="primary" @click="uploadZip" :loading="uploadingZip" :disabled="!zipFile">
                    一键覆盖上传
                  </el-button>
                </div>
	                <div class="zip-tip">
	                  注意：批量上传将<b>删除所有现有测试点</b>。Zip 包内应包含成对的 .in 和 .out 文件。
	                </div>
	                <div class="upload-progress" v-if="uploadingZip || zipProgress > 0">
	                  <el-progress
	                    :percentage="zipProgress"
	                    :status="zipProgressStatus"
	                    :stroke-width="10"
	                  />
	                </div>
	              </el-tab-pane>
	            </el-tabs>

            <el-table :data="testcases" stripe border style="margin-top: 20px" size="small">
              <el-table-column prop="order_num" label="#" width="60" align="center" />
              <el-table-column prop="score" label="分数" width="80" align="center" />
              <el-table-column label="输入文件" min-width="200">
                <template #default="{ row }">
                  <span class="mono">{{ row.input_file }}</span>
                </template>
              </el-table-column>
              <el-table-column label="样例" width="80" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.is_sample" size="small">是</el-tag>
                </template>
              </el-table-column>
            </el-table>
            
	            <div class="testcase-actions" v-if="testcases.length > 0">
	              <el-popconfirm title="确定对该题全部历史提交执行重测？" @confirm="rejudgeProblem">
	                <template #reference>
	                  <el-button type="warning" text :loading="rejudgingProblem">整题重测</el-button>
	                </template>
	              </el-popconfirm>
	              <el-popconfirm title="确定清空所有测试点？" @confirm="deleteAllTestcases">
	                <template #reference>
	                  <el-button type="danger" text>清空测试点</el-button>
	                </template>
	              </el-popconfirm>
	            </div>
          </el-card>
        </div>

      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from '@/utils/message'
import { problemApi } from '@/api/problem'
import MarkdownPreview from '@/components/common/MarkdownPreview.vue'
import { Document, Folder } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)

const loading = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({
  title: '',
  description: '',
  input_format: '',
  output_format: '',
  hint: '',
  samples: [{ input: '', output: '' }],
  time_limit: 1000,
  memory_limit: 256,
  difficulty: 'easy',
  tags: [],
  is_public: true,
  file_io_enabled: false,
  file_input_name: '',
  file_output_name: '',
  ai_judge_config: {
    enabled: false,
    required_algorithm: '',
    required_language: '',
    forbidden_features: [],
    custom_prompt: '',
    strict_mode: false,
  },
})

const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  description: [{ required: true, message: '请输入描述', trigger: 'blur' }],
}

const commonTags = [
  '数组', '字符串', '链表', '栈', '队列', '哈希表',
  '树', '图', '排序', '搜索', '动态规划', '贪心',
  '回溯', '分治', '数学', '位运算', '模拟',
]

// 测试数据相关
const testcases = ref([])
const loadingTestcases = ref(false)
const uploadingTestcase = ref(false)
const inputFile = ref(null)
const outputFile = ref(null)
const testcaseScore = ref(10)
const testcaseProgress = ref(0)
const testcaseProgressStatus = ref('')
const inputUploadRef = ref()
const outputUploadRef = ref()

// Zip 上传
const zipFile = ref(null)
const uploadingZip = ref(false)
const zipProgress = ref(0)
const zipProgressStatus = ref('')
const zipUploadRef = ref()
const rejudgingProblem = ref(false)

function addSample() {
  form.samples.push({ input: '', output: '' })
}

function removeSample(index) {
  form.samples.splice(index, 1)
}

function handleFileIOToggle(value) {
  if (value) {
    if (!form.file_input_name) form.file_input_name = 'data.in'
    if (!form.file_output_name) form.file_output_name = 'data.out'
  } else {
    form.file_input_name = ''
    form.file_output_name = ''
  }
}

function handleInputFileChange(file) {
  inputFile.value = file.raw || null
  inputUploadRef.value?.clearFiles()
}

function handleOutputFileChange(file) {
  outputFile.value = file.raw || null
  outputUploadRef.value?.clearFiles()
}

function handleZipFileChange(file) {
  zipFile.value = file.raw || null
  zipUploadRef.value?.clearFiles()
}

async function fetchProblem() {
  if (!isEdit.value) return
  
  loading.value = true
  try {
    const res = await problemApi.getById(route.params.id)
    Object.assign(form, res.data)
    if (!form.ai_judge_config) {
      form.ai_judge_config = {
        enabled: false,
        required_algorithm: '',
        required_language: '',
        forbidden_features: [],
        custom_prompt: '',
        strict_mode: false,
      }
    }
    if (!form.samples || form.samples.length === 0) {
      form.samples = [{ input: '', output: '' }]
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchTestcases() {
  if (!isEdit.value) return
  
  loadingTestcases.value = true
  try {
    const res = await problemApi.getTestcases(route.params.id)
    testcases.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loadingTestcases.value = false
  }
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  
  const submitData = { ...form }
  submitData.samples = form.samples.filter(s => s.input || s.output)

  if (submitData.file_io_enabled) {
    if (!submitData.file_input_name || !submitData.file_output_name) {
      message.warning('请填写输入/输出文件名')
      return
    }
  } else {
    submitData.file_input_name = ''
    submitData.file_output_name = ''
  }
  
  if (!submitData.ai_judge_config.enabled) {
    submitData.ai_judge_config = null
  }
  
  submitting.value = true
  try {
    if (isEdit.value) {
      await problemApi.update(route.params.id, submitData)
      message.success('保存成功')
    } else {
      const res = await problemApi.create(submitData)
      message.success('创建成功')
      router.replace(`/admin/problem/${res.data.id}/edit`)
    }
  } catch (e) {
    console.error(e)
  } finally {
    submitting.value = false
  }
}

async function uploadTestcase() {
  if (!inputFile.value || !outputFile.value) return

  const formData = new FormData()
  formData.append('input', inputFile.value)
  formData.append('output', outputFile.value)
  formData.append('score', testcaseScore.value)
  
  uploadingTestcase.value = true
  testcaseProgress.value = 0
  testcaseProgressStatus.value = ''
  try {
    await problemApi.uploadTestcase(route.params.id, formData, {
      timeout: 600000,
      onUploadProgress: (event) => {
        if (!event.total) return
        testcaseProgress.value = Math.min(99, Math.round((event.loaded * 100) / event.total))
      },
    })
    testcaseProgress.value = 100
    testcaseProgressStatus.value = 'success'
    message.success('上传成功')
    inputFile.value = null
    outputFile.value = null
    inputUploadRef.value?.clearFiles()
    outputUploadRef.value?.clearFiles()
    fetchTestcases()
  } catch (e) {
    testcaseProgressStatus.value = 'exception'
    console.error(e)
  } finally {
    uploadingTestcase.value = false
  }
}

async function uploadZip() {
  if (!zipFile.value) return

  const formData = new FormData()
  formData.append('zip_file', zipFile.value)
  
  uploadingZip.value = true
  zipProgress.value = 0
  zipProgressStatus.value = ''
  try {
    await problemApi.uploadTestcaseZip(route.params.id, formData, {
      timeout: 600000,
      onUploadProgress: (event) => {
        if (!event.total) return
        zipProgress.value = Math.min(99, Math.round((event.loaded * 100) / event.total))
      },
    })
    zipProgress.value = 100
    zipProgressStatus.value = 'success'
    message.success('批量上传成功')
    zipFile.value = null
    zipUploadRef.value?.clearFiles()
    fetchTestcases()
  } catch (e) {
    zipProgressStatus.value = 'exception'
    console.error(e)
  } finally {
    uploadingZip.value = false
  }
}

async function rejudgeProblem() {
  if (!isEdit.value) return

  rejudgingProblem.value = true
  try {
    const res = await problemApi.rejudge(route.params.id)
    const queued = res.data?.queued ?? 0
    const failed = res.data?.failed ?? 0
    if (failed > 0) {
      message.warning(`已入队 ${queued} 条，失败 ${failed} 条`)
    } else {
      message.success(res.message || `已入队 ${queued} 条重测任务`)
    }
  } catch (e) {
    console.error(e)
  } finally {
    rejudgingProblem.value = false
  }
}

async function deleteAllTestcases() {
  try {
    await problemApi.deleteTestcases(route.params.id)
    message.success('删除成功')
    fetchTestcases()
  } catch (e) {
    console.error(e)
  }
}

onMounted(() => {
  fetchProblem()
  fetchTestcases()
})
</script>

<style lang="scss" scoped>
.problem-edit-wrapper {
  padding: 40px 0;
  background-color: var(--swiss-bg-base);
  min-height: 100vh;
}

/* Fluid Container Override for Editor */
.problem-edit-wrapper .container {
  max-width: 96%;
  min-width: 1000px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.page-title {
  font-size: 28px;
  color: var(--swiss-text-main);
  margin: 0;
}

.form-section {
  margin-bottom: 40px;
  
  .section-header {
    display: flex;
    align-items: baseline;
    gap: 12px;
    margin-bottom: 16px;
    
    h3 {
      font-size: 18px;
      color: var(--swiss-text-main);
      margin: 0;
    }
    
    .sub-text {
      font-size: 13px;
      color: var(--swiss-text-secondary);
    }
  }
}

.form-card {
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
}

.switch-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hint-text {
  font-size: 12px;
  color: var(--swiss-text-secondary);
}

.file-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: 12px;
  
  .el-input {
    width: 120px;
  }
}

/* Markdown Row - Dual Column */
.md-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  height: 250px; /* Default height */
  width: 100%; /* 确保占满容器 */
  
  &.big {
    height: 600px; /* Taller for Description */
  }
  
  /* Breakpoint lowered to 768px */
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
    height: auto;
    
    /* Stack them instead of hiding preview */
    .md-preview {
      height: 300px;
      margin-top: 12px;
    }
    
    .md-input textarea {
      height: 300px !important;
    }
  }
}

.md-input :deep(.el-textarea__inner) {
  height: 100%;
  font-family: var(--font-mono);
  font-size: 14px;
  line-height: 1.6;
  padding: 16px;
  background: #fafafa;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  
  &:focus {
    background: #fff;
    border-color: var(--swiss-primary);
  }
}

.md-preview {
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 16px;
  overflow-y: auto;
  background: #fff;
  height: 100%; /* 确保高度填满 */
  width: 100%;  /* 确保宽度填满 */
  box-sizing: border-box;
}

/* 强制 Grid 在大屏下始终生效 - 移除了重复定义 */
.form-actions-bar {
  display: flex;
  justify-content: center;
  margin: 40px 0;
  
  .el-button {
    width: 240px;
    height: 48px;
    font-size: 16px;
  }
}

/* Sample */
.sample-row {
  background: #f9f9f9;
  border-radius: var(--radius-sm);
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid var(--swiss-border-light);
  
  .sample-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 12px;
    font-size: 13px;
    font-weight: 600;
    color: var(--swiss-text-secondary);
  }
  
  .sample-inputs {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
}

.add-btn {
  width: 100%;
  border-style: dashed;
}

.limit-input {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;

  .el-input-number {
    flex: 1;
  }
}

.unit-text {
  color: var(--swiss-text-secondary);
  font-size: 12px;
  pointer-events: none;
  min-width: 24px;
  text-align: right;
}

.ai-config {
  background: #f0f7ff;
  border: 1px solid #d6e4ff;
  border-radius: var(--radius-sm);
  padding: 20px;
  
  .ai-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }
}

/* Upload */
.upload-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 0;
}

.upload-progress {
  margin-top: 4px;
  margin-bottom: 6px;
}

.upload-group {
  display: flex;
  align-items: center;
  gap: 8px;
  
  .plus { color: var(--swiss-text-secondary); }
}

.score-input {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
  font-size: 14px;
}

.testcase-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.zip-tip {
  margin-top: 12px;
  font-size: 13px;
  color: var(--swiss-warning);
}

.mono {
  font-family: var(--font-mono);
  font-size: 12px;
}
</style>
