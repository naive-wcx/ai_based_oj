<template>
  <div class="problem-edit">
    <div class="page-header">
      <h2>{{ isEdit ? '编辑题目' : '创建题目' }}</h2>
    </div>
    
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      class="card"
      v-loading="loading"
    >
      <el-form-item label="题目标题" prop="title">
        <el-input v-model="form.title" placeholder="请输入题目标题" />
      </el-form-item>
      
      <el-form-item label="题目描述" prop="description">
        <div class="editor-with-preview">
          <div class="editor-section">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="10"
              placeholder="支持 Markdown 格式"
            />
          </div>
          <div class="preview-section">
            <div class="preview-header">
              <span>预览</span>
            </div>
            <div class="preview-content">
              <MarkdownPreview :content="form.description" />
            </div>
          </div>
        </div>
      </el-form-item>
      
      <el-form-item label="输入格式">
        <div class="editor-with-preview">
          <div class="editor-section">
            <el-input
              v-model="form.input_format"
              type="textarea"
              :rows="4"
              placeholder="输入格式说明（支持 Markdown）"
            />
          </div>
          <div class="preview-section small">
            <div class="preview-header"><span>预览</span></div>
            <div class="preview-content">
              <MarkdownPreview :content="form.input_format" />
            </div>
          </div>
        </div>
      </el-form-item>
      
      <el-form-item label="输出格式">
        <div class="editor-with-preview">
          <div class="editor-section">
            <el-input
              v-model="form.output_format"
              type="textarea"
              :rows="4"
              placeholder="输出格式说明（支持 Markdown）"
            />
          </div>
          <div class="preview-section small">
            <div class="preview-header"><span>预览</span></div>
            <div class="preview-content">
              <MarkdownPreview :content="form.output_format" />
            </div>
          </div>
        </div>
      </el-form-item>
      
      <el-form-item label="样例">
        <div v-for="(sample, index) in form.samples" :key="index" class="sample-item">
          <div class="sample-inputs">
            <el-input
              v-model="sample.input"
              type="textarea"
              :rows="3"
              placeholder="输入"
            />
            <el-input
              v-model="sample.output"
              type="textarea"
              :rows="3"
              placeholder="输出"
            />
          </div>
          <el-button type="danger" text @click="removeSample(index)">删除</el-button>
        </div>
        <el-button @click="addSample">添加样例</el-button>
      </el-form-item>
      
      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item label="时间限制">
            <el-input-number v-model="form.time_limit" :min="100" :max="10000" :step="100" />
            <span class="unit">ms</span>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="内存限制">
            <el-input-number v-model="form.memory_limit" :min="16" :max="1024" :step="16" />
            <span class="unit">MB</span>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="难度">
            <el-select v-model="form.difficulty" style="width: 100%">
              <el-option label="简单" value="easy" />
              <el-option label="中等" value="medium" />
              <el-option label="困难" value="hard" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      
      <el-form-item label="标签">
        <el-select
          v-model="form.tags"
          multiple
          filterable
          allow-create
          placeholder="选择或输入标签"
          style="width: 100%"
        >
          <el-option v-for="tag in commonTags" :key="tag" :label="tag" :value="tag" />
        </el-select>
      </el-form-item>
      
      <el-form-item label="是否公开">
        <el-switch v-model="form.is_public" />
      </el-form-item>

      <el-divider content-position="left">文件操作</el-divider>

      <el-form-item label="启用文件操作">
        <el-switch v-model="form.file_io_enabled" @change="handleFileIOToggle" />
        <span class="hint">开启后需从指定文件读写（如 data.in / data.out）</span>
      </el-form-item>

      <template v-if="form.file_io_enabled">
        <el-form-item label="输入文件名">
          <el-input
            v-model="form.file_input_name"
            placeholder="例如：data.in"
          />
        </el-form-item>
        <el-form-item label="输出文件名">
          <el-input
            v-model="form.file_output_name"
            placeholder="例如：data.out"
          />
        </el-form-item>
      </template>
      
      <el-divider content-position="left">AI 判题配置</el-divider>
      
      <el-form-item label="启用 AI 判题">
        <el-switch v-model="form.ai_judge_config.enabled" />
      </el-form-item>
      
      <template v-if="form.ai_judge_config.enabled">
        <el-form-item label="要求算法">
          <el-input
            v-model="form.ai_judge_config.required_algorithm"
            placeholder="如：动态规划、DFS、贪心 等"
          />
        </el-form-item>
        
        <el-form-item label="要求语言">
          <el-select v-model="form.ai_judge_config.required_language" clearable style="width: 100%">
            <el-option label="不限" value="" />
            <el-option label="C" value="C" />
            <el-option label="C++" value="C++" />
            <el-option label="Python" value="Python" />
            <el-option label="Java" value="Java" />
            <el-option label="Go" value="Go" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="禁止特性">
          <el-select
            v-model="form.ai_judge_config.forbidden_features"
            multiple
            filterable
            allow-create
            placeholder="选择或输入禁止使用的特性"
            style="width: 100%"
          >
            <el-option label="STL sort" value="STL sort" />
            <el-option label="递归" value="递归" />
            <el-option label="全局变量" value="全局变量" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="自定义说明">
          <el-input
            v-model="form.ai_judge_config.custom_prompt"
            type="textarea"
            :rows="3"
            placeholder="额外的判题说明"
          />
        </el-form-item>
        
        <el-form-item label="严格模式">
          <el-switch v-model="form.ai_judge_config.strict_mode" />
          <span class="hint">开启后，AI 判定不通过将直接判为 Wrong Answer</span>
        </el-form-item>
      </template>
      
      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEdit ? '保存修改' : '创建题目' }}
        </el-button>
        <el-button @click="$router.back()">取消</el-button>
      </el-form-item>
    </el-form>
    
    <!-- 测试数据管理 -->
    <div class="card" v-if="isEdit">
      <h3>测试数据管理</h3>
      
      <el-tabs type="border-card" class="testcase-tabs">
        <el-tab-pane label="逐个上传">
          <div class="upload-panel">
            <el-upload
              ref="inputUploadRef"
              action=""
              :auto-upload="false"
              :limit="1"
              :on-change="handleInputChange"
              :on-remove="() => inputFile = null"
              accept=".in,.txt"
              class="upload-item"
            >
              <template #trigger>
                <el-button icon="Document">选择输入 (.in)</el-button>
              </template>
              <div class="el-upload__tip" v-if="inputFile">已选: {{ inputFile.name }}</div>
            </el-upload>
            
            <el-upload
              ref="outputUploadRef"
              action=""
              :auto-upload="false"
              :limit="1"
              :on-change="handleOutputChange"
              :on-remove="() => outputFile = null"
              accept=".out,.txt"
              class="upload-item"
            >
              <template #trigger>
                <el-button icon="Document">选择输出 (.out)</el-button>
              </template>
              <div class="el-upload__tip" v-if="outputFile">已选: {{ outputFile.name }}</div>
            </el-upload>
            
            <div class="score-input">
              <span class="label">分值:</span>
              <el-input-number v-model="testcaseScore" :min="1" :max="100" style="width: 140px" />
            </div>
            
            <el-button type="primary" @click="uploadTestcase" :loading="uploadingTestcase">
              上传测试点
            </el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="批量上传 (Zip)">
          <div class="upload-panel">
            <el-upload
              ref="zipUploadRef"
              action=""
              :auto-upload="false"
              :limit="1"
              :on-change="handleZipChange"
              :on-remove="() => zipFile = null"
              accept=".zip"
              class="upload-item"
            >
              <template #trigger>
                <el-button type="warning" plain icon="Folder">选择 Zip 包</el-button>
              </template>
              <div class="el-upload__tip" v-if="zipFile">已选: {{ zipFile.name }}</div>
            </el-upload>
            
            <el-button type="warning" @click="uploadZip" :loading="uploadingZip" :disabled="!zipFile">
              一键批量上传
            </el-button>
          </div>
          <el-alert
            title="注意：批量上传将自动覆盖所有现有测试点"
            type="warning"
            :closable="false"
            show-icon
            style="margin-top: 16px"
          >
            <template #default>
              <div>1. 系统会自动识别 Zip 包内的 .in 和 .out/.ans 文件并配对。</div>
              <div>2. 测试点分数将自动均分（例如 10 个测试点，每个 10 分）。</div>
            </template>
          </el-alert>
        </el-tab-pane>
      </el-tabs>
      
      <el-table :data="testcases" v-loading="loadingTestcases" stripe style="margin-top: 16px">
        <el-table-column prop="order_num" label="序号" width="80" />
        <el-table-column prop="score" label="分数" width="80" />
        <el-table-column label="输入文件" min-width="200">
          <template #default="{ row }">
            <span class="file-path">{{ row.input_file }}</span>
          </template>
        </el-table-column>
        <el-table-column label="是否样例" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_sample ? 'success' : 'info'" size="small">
              {{ row.is_sample ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      
      <div style="margin-top: 16px" v-if="testcases.length > 0">
        <el-popconfirm title="确定要删除所有测试数据吗？" @confirm="deleteAllTestcases">
          <template #reference>
            <el-button type="danger" plain>删除所有测试数据</el-button>
          </template>
        </el-popconfirm>
      </div>
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
  title: [{ required: true, message: '请输入题目标题', trigger: 'blur' }],
  description: [{ required: true, message: '请输入题目描述', trigger: 'blur' }],
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
const inputUploadRef = ref()
const outputUploadRef = ref()
const activeTab = ref('0') // 默认第一个 tab (Element Plus tabs don't need v-model strictly if not controlled, but good practice)

// Zip 上传
const zipFile = ref(null)
const uploadingZip = ref(false)
const zipUploadRef = ref()

function handleInputChange(file) {
  inputFile.value = file.raw
}

function handleOutputChange(file) {
  outputFile.value = file.raw
}

function handleZipChange(file) {
  zipFile.value = file.raw
}

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
  
  // 过滤空样例
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
  
  // 如果没有启用 AI 判题，清空配置
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
  if (!inputFile.value || !outputFile.value) {
    message.warning('请选择输入和输出文件')
    return
  }
  
  const formData = new FormData()
  formData.append('input', inputFile.value)
  formData.append('output', outputFile.value)
  formData.append('score', testcaseScore.value)
  
  uploadingTestcase.value = true
  try {
    await problemApi.uploadTestcase(route.params.id, formData)
    message.success('上传成功')
    
    // 清空选择
    inputFile.value = null
    outputFile.value = null
    inputUploadRef.value?.clearFiles()
    outputUploadRef.value?.clearFiles()
    
    fetchTestcases()
  } catch (e) {
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
  try {
    await problemApi.uploadTestcaseZip(route.params.id, formData)
    message.success('批量上传成功')
    
    zipFile.value = null
    zipUploadRef.value?.clearFiles()
    fetchTestcases()
  } catch (e) {
    console.error(e)
  } finally {
    uploadingZip.value = false
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
.page-header {
  margin-bottom: 20px;
  
  h2 {
    margin: 0;
  }
}

.editor-with-preview {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  width: 100%;
  
  @media (max-width: 1200px) {
    grid-template-columns: 1fr;
  }
  
  .editor-section {
    min-width: 0;
  }
  
  .preview-section {
    border: 1px solid #e4e7ed;
    border-radius: 4px;
    overflow: hidden;
    
    &.small {
      max-height: 150px;
      
      .preview-content {
        max-height: 120px;
      }
    }
    
    .preview-header {
      background: #f5f7fa;
      padding: 8px 12px;
      font-size: 12px;
      color: #909399;
      border-bottom: 1px solid #e4e7ed;
    }
    
    .preview-content {
      padding: 12px;
      max-height: 300px;
      overflow-y: auto;
      background: #fff;
    }
  }
}

.sample-item {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  
  .sample-inputs {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
    margin-bottom: 8px;
  }
}

.unit {
  margin-left: 8px;
  color: #909399;
}

.hint {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}

.card h3 {
  margin-bottom: 16px;
}

.upload-panel {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  padding: 12px 0;
}

.upload-item {
  display: inline-flex;
  flex-direction: column;
}

.score-input {
  display: flex;
  align-items: center;
  gap: 8px;
  
  .label {
    font-size: 14px;
    color: #606266;
  }
}

.file-path {
  font-family: monospace;
  font-size: 12px;
  color: #606266;
}
</style>
