<template>
  <el-card shadow="never" class="page-container">
    <template #header>
      <div class="page-header">
        <span class="page-title">提交记录</span>
        <div class="filter-bar">
          <el-input
            v-model="filters.problem_id"
            placeholder="题目 ID"
            clearable
            style="width: 120px"
            @keyup.enter="handleSearch"
          />
          <el-input
            v-if="userStore.isAdmin"
            v-model="filters.username"
            placeholder="用户名"
            clearable
            style="width: 150px"
            @keyup.enter="handleSearch"
          />
          <el-select
            v-model="filters.language"
            placeholder="语言"
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
            placeholder="状态"
            clearable
            style="width: 150px"
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
    </template>

    <el-table :data="submissions" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="#" width="80" />
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
            <el-tag :type="getStatusTagType(row.status)" effect="light">
              {{ statusMap[row.status]?.label || row.status }}
            </el-tag>
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="用时" width="100" align="center">
        <template #default="{ row }">
          <span v-if="row.status === 'Accepted'">{{ row.time_used }}ms</span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="内存" width="100" align="center">
        <template #default="{ row }">
          <span v-if="row.status === 'Accepted'">{{ formatMemory(row.memory_used) }}</span>
           <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column v-if="userStore.isAdmin" prop="username" label="用户" width="150" />
      <el-table-column label="提交时间" width="180" align="center">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container" v-if="pagination.total > 0">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="handleSearch"
        @current-change="fetchSubmissions"
      />
    </div>
  </el-card>
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
  if (!kb) return '-'
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
.page-container {
  border: none;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.problem-title {
  color: var(--el-text-color-primary);
  text-decoration: none;
  &:hover {
    color: var(--el-color-primary);
  }
}

.status-link {
  text-decoration: none;
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
