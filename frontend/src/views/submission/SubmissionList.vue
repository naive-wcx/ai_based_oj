<template>
  <div class="swiss-layout">
    <div class="swiss-header">
      <h1 class="swiss-title">提交记录</h1>
      <div class="filter-group">
        <el-input
          v-model="filters.problem_id"
          placeholder="题目 ID"
          style="width: 100px"
          @keyup.enter="handleSearch"
        />
        <el-input
          v-if="userStore.isAdmin"
          v-model="filters.username"
          placeholder="用户名"
          style="width: 140px"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="filters.language"
          placeholder="语言"
          clearable
          style="width: 100px"
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
          placeholder="状态"
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

    <el-table :data="submissions" v-loading="loading" class="swiss-table">
      <el-table-column prop="id" label="#" width="80" align="center">
        <template #default="{ row }">
          <span class="swiss-font-mono" style="color: var(--color-text-secondary)">{{ row.id }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="题目" min-width="200">
        <template #default="{ row }">
          <router-link :to="`/problem/${row.problem_id}`" class="problem-title">
            {{ row.problem_id }}. {{ row.problem_title }}
          </router-link>
        </template>
      </el-table-column>
      
      <el-table-column label="状态" width="180" align="center">
        <template #default="{ row }">
          <router-link :to="`/submission/${row.id}`" class="status-link">
            <span :class="['status-badge', getStatusClass(row.status)]">
              {{ statusMap[row.status]?.label || row.status }}
            </span>
          </router-link>
        </template>
      </el-table-column>
      
      <el-table-column label="用时" width="100" align="right">
        <template #default="{ row }">
          <span v-if="row.status === 'Accepted'" class="swiss-font-mono stat-text">{{ row.time_used }}ms</span>
          <span v-else class="stat-text">-</span>
        </template>
      </el-table-column>
      
      <el-table-column label="内存" width="100" align="right">
        <template #default="{ row }">
          <span v-if="row.status === 'Accepted'" class="swiss-font-mono stat-text">{{ formatMemory(row.memory_used) }}</span>
           <span v-else class="stat-text">-</span>
        </template>
      </el-table-column>
      
      <el-table-column v-if="userStore.isAdmin" prop="username" label="用户" width="150" align="center" />
      
      <el-table-column label="提交时间" width="180" align="right">
        <template #default="{ row }">
          <span class="swiss-font-mono time-text">{{ formatTime(row.created_at) }}</span>
        </template>
      </el-table-column>
    </el-table>

    <div class="swiss-pagination" v-if="pagination.total > 0">
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
  'Pending': { label: '等待中', class: 'pending' },
  'Judging': { label: '评测中', class: 'judging' },
  'Accepted': { label: '通过', class: 'accepted' },
  'Submitted': { label: '已提交', class: 'submitted' },
  'Wrong Answer': { label: '答案错误', class: 'error' },
  'Time Limit Exceeded': { label: '超时', class: 'warning' },
  'Memory Limit Exceeded': { label: '内存超限', class: 'warning' },
  'Runtime Error': { label: '运行错误', class: 'error' },
  'Compile Error': { label: '编译错误', class: 'error' },
  'System Error': { label: '系统错误', class: 'error' },
}

const getStatusClass = (status) => {
  return statusMap[status]?.class || 'default'
}

function formatMemory(kb) {
  if (!kb) return '-'
  if (kb < 1024) return `${kb}KB`
  return `${(kb / 1024).toFixed(1)}MB`
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
.filter-group {
  display: flex;
  gap: 16px;
}

.problem-title {
  color: var(--color-text-primary);
  text-decoration: none;
  font-weight: 500;
  &:hover {
    color: var(--color-primary);
  }
}

.status-link {
  text-decoration: none;
}

.status-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 4px;
  
  &.accepted { color: var(--color-success); background: rgba(16, 185, 129, 0.1); }
  &.error { color: var(--color-danger); background: rgba(239, 68, 68, 0.1); }
  &.warning { color: var(--color-warning); background: rgba(245, 158, 11, 0.1); }
  &.judging { color: var(--color-primary); background: rgba(15, 82, 186, 0.1); }
  &.pending { color: var(--color-text-secondary); background: rgba(0,0,0,0.05); }
}

.stat-text, .time-text {
  font-size: 13px;
  color: var(--color-text-secondary);
}
</style>