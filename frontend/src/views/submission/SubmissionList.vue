<template>
  <div class="submission-list-wrapper">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">提交记录</h1>
        <div class="filter-group">
          <el-input
            v-model="filters.problem_id"
            placeholder="题目 ID"
            style="width: 120px"
            @keyup.enter="handleSearch"
          >
             <template #prefix>#</template>
          </el-input>
          
          <el-input
            v-if="userStore.isAdmin"
            v-model="filters.username"
            placeholder="用户名"
            style="width: 140px"
            @keyup.enter="handleSearch"
          />
          
          <el-select
            v-model="filters.language"
            placeholder="所有语言"
            clearable
            style="width: 120px"
            @change="handleSearch"
          >
            <el-option
              v-for="(label, lang) in languageLabels"
              :key="lang"
              :label="label"
              :value="lang"
            />
          </el-select>
          
          <el-select
            v-model="filters.status"
            placeholder="所有状态"
            clearable
            style="width: 140px"
            @change="handleSearch"
          >
            <el-option
              v-for="(item, status) in statusMap"
              :key="status"
              :label="item.label"
              :value="status"
            />
          </el-select>
        </div>
      </div>

      <div class="table-container">
        <el-table :data="submissions" v-loading="loading" class="swiss-table">
          <!-- 编号 -->
          <el-table-column prop="id" label="编号" width="100" align="center" header-align="center">
            <template #default="{ row }">
              <span class="id-text">{{ row.id }}</span>
            </template>
          </el-table-column>
          
          <!-- 题目 -->
          <el-table-column label="题目" min-width="200" align="center" header-align="center">
            <template #default="{ row }">
              <router-link :to="`/problem/${row.problem_id}`" class="problem-link">
                {{ row.problem_id }}. {{ row.problem_title }}
              </router-link>
            </template>
          </el-table-column>
          
          <!-- 状态 -->
          <el-table-column label="状态" width="180" align="center" header-align="center">
            <template #default="{ row }">
              <router-link :to="`/submission/${row.id}`" class="status-link">
                <span :class="['status-tag', getStatusClass(row.status)]">
                  {{ statusMap[row.status]?.label || row.status }}
                </span>
              </router-link>
            </template>
          </el-table-column>
          
          <!-- 用时 -->
          <el-table-column label="用时" width="100" align="center" header-align="center">
            <template #default="{ row }">
              <span class="stat-text">{{ row.time_used != null ? row.time_used + ' ms' : '-' }}</span>
            </template>
          </el-table-column>
          
          <!-- 内存 -->
          <el-table-column label="内存" width="100" align="center" header-align="center">
            <template #default="{ row }">
              <span class="stat-text">{{ formatMemory(row.memory_used) }}</span>
            </template>
          </el-table-column>
          
          <!-- 用户 -->
          <el-table-column v-if="userStore.isAdmin" prop="username" label="用户" width="150" align="center" header-align="center">
             <template #default="{ row }">
              <span class="user-text">{{ row.username }}</span>
            </template>
          </el-table-column>
          
          <!-- 提交时间 -->
          <el-table-column label="提交时间" width="180" align="center" header-align="center">
            <template #default="{ row }">
              <span class="time-text">{{ formatTime(row.created_at) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="prev, pager, next"
          @size-change="handleSearch"
          @current-change="fetchSubmissions"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { submissionApi } from '@/api/submission'
import { useRoute, useRouter } from 'vue-router'
import { message } from '@/utils/message'
import { useUserStore } from '@/stores/user'

const loading = ref(true)
const submissions = ref([])
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const filters = reactive({
  problem_id: route.query.problem_id || '',
  username: route.query.username || '',
  language: route.query.language || '',
  status: route.query.status || '',
})

const pagination = reactive({
  page: parseInt(route.query.page, 10) || 1,
  size: parseInt(route.query.size, 10) || 50,
  total: 0,
})

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

const getStatusClass = (status) => {
  return statusMap[status]?.class || 'waiting'
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

const updateUrl = () => {
  const query = {
    page: pagination.page,
    size: pagination.size,
    ...Object.fromEntries(
      Object.entries(filters).filter(([, value]) => value)
    ),
  }
  router.push({ query })
}

async function fetchSubmissions() {
  loading.value = true
  try {
    const params = { ...filters, ...pagination }
    const res = await submissionApi.getList(params)
    submissions.value = res.data.list || []
    pagination.total = res.data.total
    updateUrl()
  } catch (e) {
    message.error('提交记录加载失败')
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchSubmissions()
}

onMounted(() => {
  fetchSubmissions()
})
</script>

<style lang="scss" scoped>
.submission-list-wrapper {
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

.filter-group {
  display: flex;
  gap: 16px;
}

.table-container {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.id-text {
  color: var(--swiss-text-secondary);
  font-family: var(--font-mono);
  font-size: 13px;
}

.problem-link {
  color: var(--swiss-text-main);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
  &:hover { color: var(--swiss-primary); }
}

.status-link {
  text-decoration: none;
}

/* Status Colors */
.status-tag {
  font-size: 13px;
  font-weight: 700;
  padding: 4px 8px;
  border-radius: 4px;
  background: transparent;
  
  &.ac { color: var(--status-ac); }
  &.wa { color: var(--status-wa); }
  &.tle { color: var(--status-tle); }
  &.mle { color: var(--status-mle); }
  &.re { color: var(--status-re); }
  &.ce { color: var(--status-ce); }
  &.uqe { color: var(--status-uqe); }
  &.waiting { color: var(--status-waiting); }
  &.judging { color: var(--status-judging); }
}

.stat-text {
  font-size: 13px;
  color: var(--swiss-text-secondary);
}

.time-text {
  font-size: 13px;
  color: var(--swiss-text-secondary);
  white-space: nowrap;
}

.user-text {
  color: var(--swiss-text-main);
  font-size: 13px;
}

.pagination-wrapper {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}
</style>