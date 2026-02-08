<template>
  <div class="settings-wrapper">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">系统设置</h1>
      </div>
      
      <div class="settings-grid">
        <!-- AI Configuration Card -->
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <span>AI 判题配置</span>
            </div>
          </template>
          
          <el-form
            ref="formRef"
            :model="form"
            label-position="top"
            v-loading="loading"
          >
            <el-form-item label="启用 AI 判题">
              <el-switch v-model="form.enabled" active-text="已启用" inactive-text="已禁用" />
            </el-form-item>
            
            <template v-if="form.enabled">
              <div class="config-section">
                <el-form-item label="服务商预设">
                  <el-radio-group v-model="form.provider" @change="handleProviderChange">
                    <el-radio-button label="deepseek">DeepSeek</el-radio-button>
                    <el-radio-button label="openai">OpenAI</el-radio-button>
                    <el-radio-button label="moonshot">Kimi (Moonshot)</el-radio-button>
                    <el-radio-button label="other">自定义</el-radio-button>
                  </el-radio-group>
                </el-form-item>
                
                <el-row :gutter="20">
                  <el-col :span="12">
                     <el-form-item label="模型名称" required>
                      <el-input v-model="form.model" placeholder="例如: gpt-4, deepseek-chat" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                     <el-form-item label="超时时间 (秒)">
                      <el-input-number v-model="form.timeout" :min="10" :max="300" controls-position="right" style="width: 100%" />
                    </el-form-item>
                  </el-col>
                </el-row>

                <el-form-item label="API 接口地址" required>
                  <el-input v-model="form.api_url" placeholder="例如: https://api.deepseek.com/v1/chat/completions" />
                </el-form-item>

                <el-form-item label="API Key" required>
                  <div class="api-key-input">
                    <el-input
                      v-model="form.api_key"
                      type="password"
                      show-password
                      placeholder="sk-..."
                    />
                  </div>
                  <div class="form-helper" v-if="isMaskedKey">
                    Key 已安全保存。如需修改，请直接输入新的 Key。
                  </div>
                </el-form-item>
              </div>
            </template>

            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave">保存设置</el-button>
              <el-button :loading="testing" @click="handleTest" :disabled="!canTest" plain>测试连接</el-button>
            </div>
          </el-form>
        </el-card>

        <!-- Documentation Card -->
        <div class="info-sidebar">
          <el-card shadow="never" class="info-card">
            <template #header>功能说明</template>
            <div class="info-content">
              <p>AI 判题系统使用大语言模型 (LLM) 分析学生提交的代码，主要用于验证：</p>
              <ul>
                <li><strong>算法使用：</strong> 是否按要求使用了特定算法（如动态规划）？</li>
                <li><strong>语言约束：</strong> 是否使用了被禁止的库函数（如 `std::sort`）？</li>
              </ul>
              <h4>费用提示</h4>
              <p>使用商业 API（如 OpenAI 或 DeepSeek）会产生 Token 费用。DeepSeek-V3 提供了极高的性价比，推荐优先使用。</p>
            </div>
          </el-card>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { message } from '@/utils/message'
import { adminApi } from '@/api/admin'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)

const form = reactive({
  enabled: false,
  provider: 'deepseek',
  api_key: '',
  api_url: '',
  model: '',
  timeout: 60,
})

const isMaskedKey = computed(() => form.api_key === '********')
const canTest = computed(() => form.enabled && !!form.api_key)

// Preset Configurations
const presets = {
  deepseek: {
    api_url: 'https://api.deepseek.com/v1/chat/completions',
    model: 'deepseek-chat',
  },
  openai: {
    api_url: 'https://api.openai.com/v1/chat/completions',
    model: 'gpt-4o',
  },
  moonshot: {
    api_url: 'https://api.moonshot.cn/v1/chat/completions',
    model: 'moonshot-v1-8k',
  }
}

function handleProviderChange(val) {
  if (presets[val]) {
    form.api_url = presets[val].api_url
    form.model = presets[val].model
  }
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
    message.success('设置保存成功')
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function handleTest() {
  testing.value = true
  try {
    await adminApi.testAIConnection()
    message.success('连接测试成功！')
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
.settings-wrapper {
  padding: 40px 0;
  min-height: 100vh;
  background-color: var(--swiss-bg-base);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--swiss-border-light);
}

.page-title {
  font-size: 32px;
  color: var(--swiss-text-main);
  letter-spacing: -0.02em;
  margin: 0;
}

.settings-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 30px;
}

.settings-card {
  background: #fff;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.config-section {
  background: var(--swiss-bg-alt);
  padding: 24px;
  border-radius: var(--radius-sm);
  margin-bottom: 24px;
  border: 1px solid var(--swiss-border-light);
}

.form-actions {
  display: flex;
  gap: 12px;
}

.form-helper {
  font-size: 12px;
  color: var(--swiss-text-secondary);
  margin-top: 6px;
}

.info-content {
  font-size: 14px;
  line-height: 1.6;
  color: var(--swiss-text-secondary);
  
  ul {
    padding-left: 20px;
  }
  
  li {
    margin-bottom: 8px;
  }
  
  h4 {
    margin-top: 20px;
    margin-bottom: 8px;
    color: var(--swiss-text-main);
  }
}

@media (max-width: 900px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
}
</style>
