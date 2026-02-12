<template>
  <div class="submission-list-wrapper">
    <div class="container">
      <div class="page-header">
        <div>
          <h1 class="page-title">提交记录</h1>
          <p class="page-subtitle">支持按题目、用户（管理员）和状态快速筛选。</p>
        </div>
        <div class="header-stats">
          <div class="stat-chip">
            <span>总提交</span>
            <strong>{{ pagination.total }}</strong>
          </div>
          <div class="stat-chip">
            <span>本页 AC</span>
            <strong>{{ pageAcceptedCount }}</strong>
          </div>
          <div class="stat-chip">
            <span>评测中</span>
            <strong>{{ pageJudgingCount }}</strong>
          </div>
        </div>
      </div>

      <div class="filter-panel">
        <el-input
          v-model="filters.problem_id"
          clearable
          placeholder="题目 ID"
          class="id-input"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #prefix>#</template>
        </el-input>

        <el-input
          v-if="userStore.isAdmin"
          v-model="filters.user_id"
          clearable
          placeholder="用户 ID"
          class="id-input"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />

        <el-select
          v-model="filters.status"
          placeholder="所有状态"
          clearable
          class="status-select"
          @change="handleSearch"
        >
          <el-option
            v-for="(item, status) in statusMap"
            :key="status"
            :label="item.label"
            :value="status"
          />
        </el-select>

        <div class="filter-actions">
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button plain @click="handleReset">重置</el-button>
        </div>
      </div>

      <div class="table-container" v-loading="loading">
        <el-table
          v-if="submissions.length"
          :data="submissions"
          class="swiss-table"
          @row-click="goSubmissionDetail"
        >
          <!-- 编号 -->
          <el-table-column prop="id" label="编号" width="100" align="center" header-align="center">
            <template #default="{ row }">
              <span class="id-text">{{ row.id }}</span>
            </template>
          </el-table-column>
          
          <!-- 题目 -->
          <el-table-column label="题目" min-width="200" align="center" header-align="center">
            <template #default="{ row }">
              <router-link :to="`/problem/${row.problem_id}`" class="problem-link" @click.stop>
                {{ row.problem_id }}. {{ row.problem_title }}
              </router-link>
            </template>
          </el-table-column>
          
          <!-- 状态 -->
          <el-table-column label="状态" width="180" align="center" header-align="center">
            <template #default="{ row }">
              <router-link :to="`/submission/${row.id}`" class="status-link" @click.stop>
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

        <el-empty v-else description="暂无提交记录" />
      </div>

      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSearch"
          @current-change="fetchSubmissions"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { submissionApi } from '@/api/submission'
import { useRoute, useRouter } from 'vue-router'
import { message } from '@/utils/message'
import { useUserStore } from '@/stores/user'

const loading = ref(true)
const submissions = ref([])
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const defaultPageSize = 50

const filters = reactive({
  problem_id: route.query.problem_id || '',
  user_id: route.query.user_id || '',
  status: route.query.status || '',
})

const pagination = reactive({
  page: parseInt(route.query.page, 10) || 1,
  size: parseInt(route.query.size, 10) || defaultPageSize,
  total: 0,
})

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

const pageAcceptedCount = computed(
  () => submissions.value.filter((item) => item.status === 'Accepted').length
)

const pageJudgingCount = computed(
  () => submissions.value.filter((item) => item.status === 'Judging' || item.status === 'Pending').length
)

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
    ...(filters.problem_id && { problem_id: filters.problem_id }),
    ...(userStore.isAdmin && filters.user_id && { user_id: filters.user_id }),
    ...(filters.status && { status: filters.status }),
  }
  router.replace({ query })
}

async function fetchSubmissions() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      status: filters.status,
    }
    if (filters.problem_id) params.problem_id = parseInt(filters.problem_id, 10) || undefined
    if (userStore.isAdmin && filters.user_id) params.user_id = parseInt(filters.user_id, 10) || undefined

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

const handleReset = () => {
  filters.problem_id = ''
  filters.user_id = ''
  filters.status = ''
  pagination.page = 1
  pagination.size = defaultPageSize
  fetchSubmissions()
}

function goSubmissionDetail(row) {
  router.push(`/submission/${row.id}`)
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
  align-items: flex-end;
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

.page-subtitle {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--swiss-text-secondary);
}

.header-stats {
  display: flex;
  gap: 10px;
}

.stat-chip {
  min-width: 90px;
  border: 1px solid var(--swiss-border-light);
  background: #fff;
  border-radius: var(--radius-sm);
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;

  span {
    font-size: 11px;
    color: var(--swiss-text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  strong {
    font-size: 18px;
    color: var(--swiss-text-main);
    line-height: 1;
  }
}

.filter-panel {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 14px;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  background: #fff;
}

.id-input {
  width: 130px;
}

.status-select {
  width: 180px;
}

.filter-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.table-container {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  overflow: hidden;
  min-height: 220px;
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

@media (max-width: 1024px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 14px;
  }

  .filter-panel {
    flex-wrap: wrap;
  }

  .filter-actions {
    margin-left: 0;
  }
}
</style>
