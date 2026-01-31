<template>
  <div class="settings-page">
    <div class="page-header">
      <h2>系统设置</h2>
    </div>
    
    <div class="card">
      <h3>
        <span class="section-icon">🤖</span>
        AI 判题设置
      </h3>
      <p class="section-desc">配置 AI 智能判题功能，支持 DeepSeek 等大模型 API</p>
      
      <el-form
        ref="formRef"
        :model="form"
        label-width="120px"
        v-loading="loading"
        class="settings-form"
      >
        <el-form-item label="启用 AI 判题">
          <el-switch v-model="form.enabled" />
          <span class="hint">开启后，题目可以配置 AI 判题规则</span>
        </el-form-item>
        
        <template v-if="form.enabled">
          <el-divider />
          
          <el-form-item label="API 提供商">
            <el-select v-model="form.provider" style="width: 200px">
              <el-option label="DeepSeek" value="deepseek" />
              <el-option label="OpenAI" value="openai" />
              <el-option label="其他" value="other" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="API Key" required>
            <el-input
              v-model="form.api_key"
              type="password"
              :show-password="!isMaskedKey"
              placeholder="请输入 API Key"
              style="width: 400px"
            />
            <el-button v-if="isMaskedKey" text class="inline-action" @click="clearApiKey">
              重新输入
            </el-button>
            <div class="form-tip">
              <div v-if="isMaskedKey">已保存的 API Key 出于安全不会显示，需修改请点击“重新输入”。</div>
              <template v-if="form.provider === 'deepseek'">
                前往 <a href="https://platform.deepseek.com/" target="_blank">DeepSeek 开放平台</a> 获取 API Key
              </template>
              <template v-else-if="form.provider === 'openai'">
                前往 <a href="https://platform.openai.com/" target="_blank">OpenAI 平台</a> 获取 API Key
              </template>
            </div>
          </el-form-item>
          
          <el-form-item label="API 地址">
            <el-input
              v-model="form.api_url"
              placeholder="API 端点地址"
              style="width: 400px"
            />
            <div class="form-tip">
              DeepSeek 默认: https://api.deepseek.com/v1/chat/completions
            </div>
          </el-form-item>
          
          <el-form-item label="模型">
            <el-input
              v-model="form.model"
              placeholder="模型名称"
              style="width: 200px"
            />
            <div class="form-tip">
              DeepSeek 推荐: deepseek-chat
            </div>
          </el-form-item>
          
          <el-form-item label="超时时间">
            <el-input-number
              v-model="form.timeout"
              :min="10"
              :max="300"
              :step="10"
            />
            <span class="unit">秒</span>
          </el-form-item>
        </template>
        
        <el-divider />
        
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">
            保存设置
          </el-button>
          <el-button :loading="testing" @click="handleTest" :disabled="!canTest">
            测试连接
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    
    <div class="card">
      <h3>
        <span class="section-icon">ℹ️</span>
        使用说明
      </h3>
      <div class="help-content">
        <h4>什么是 AI 判题？</h4>
        <p>AI 判题功能可以分析用户提交的代码，检测是否使用了指定的算法或编程语言特性。例如：</p>
        <ul>
          <li>要求使用"动态规划"算法，但用户使用了"暴力枚举" → 判定为不通过</li>
          <li>禁止使用 STL sort 函数，但用户使用了 → 判定为不通过</li>
          <li>要求使用 C++ 语言 → 自动检测代码语言</li>
        </ul>
        
        <h4>如何使用？</h4>
        <ol>
          <li>在此页面配置 AI API（推荐使用 DeepSeek，性价比高）</li>
          <li>创建/编辑题目时，开启"AI 判题"选项</li>
          <li>设置算法要求、语言要求等规则</li>
          <li>用户提交代码后，系统会自动调用 AI 分析</li>
        </ol>
        
        <h4>费用说明</h4>
        <p>AI 判题会调用外部 API，会产生一定费用。DeepSeek API 价格约为 ¥1/百万 tokens，一次判题约消耗 1000-2000 tokens。</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from '@/utils/message'
import { adminApi } from '@/api/admin'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)

const form = reactive({
  enabled: false,
  provider: 'deepseek',
  api_key: '',
  api_url: 'https://api.deepseek.com/v1/chat/completions',
  model: 'deepseek-chat',
  timeout: 60,
})

const isMaskedKey = computed(() => form.api_key === '********')
const canTest = computed(() => form.enabled && !!form.api_key)

function clearApiKey() {
  form.api_key = ''
}

async function fetchSettings() {
  loading.value = true
  try {
    const res = await adminApi.getAISettings()
    Object.assign(form, res.data)
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await adminApi.updateAISettings(form)
    message.success('设置已保存')
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function handleTest() {
  testing.value = true
  try {
    const res = await adminApi.testAIConnection()
    message.success('连接成功！配置有效')
  } catch (e) {
    console.error(e)
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  fetchSettings()
})
</script>

<style lang="scss" scoped>
.page-header {
  margin-bottom: 20px;
  
  h2 {
    margin: 0;
  }
}

.card {
  h3 {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    font-size: 18px;
    
    .section-icon {
      font-size: 24px;
    }
  }
  
  .section-desc {
    color: #909399;
    margin-bottom: 20px;
  }
}

.settings-form {
  max-width: 600px;
}

.hint {
  margin-left: 12px;
  color: #909399;
  font-size: 13px;
}

.unit {
  margin-left: 8px;
  color: #909399;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  
  a {
    color: #409eff;
  }
}

.inline-action {
  margin-left: 8px;
}

.help-content {
  line-height: 1.8;
  color: #606266;
  
  h4 {
    margin: 20px 0 8px;
    color: #303133;
    
    &:first-child {
      margin-top: 0;
    }
  }
  
  ul, ol {
    padding-left: 20px;
    margin: 8px 0;
  }
  
  li {
    margin: 4px 0;
  }
}
</style>
