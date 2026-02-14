<template>
  <div class="contest-list-wrapper">
    <div class="container">
      <div class="page-header">
        <div>
          <h1 class="page-title">比赛列表</h1>
          <p class="page-subtitle">按时间查看比赛安排，点击行可进入比赛详情。</p>
        </div>
        <div class="header-stats">
          <div class="stat-chip">
            <span>总场次</span>
            <strong>{{ pagination.total }}</strong>
          </div>
          <div class="stat-chip">
            <span>进行中</span>
            <strong>{{ pageRunningCount }}</strong>
          </div>
          <div class="stat-chip">
            <span>未开始</span>
            <strong>{{ pageUpcomingCount }}</strong>
          </div>
        </div>
      </div>

      <div class="table-container" v-loading="loading">
        <el-table
          v-if="contests.length"
          :data="contests"
          class="swiss-table"
          @row-click="goContestDetail"
        >
          <el-table-column prop="id" label="ID" width="80" align="center" header-align="center">
            <template #default="{ row }">
              <span class="id-text">{{ row.id }}</span>
            </template>
          </el-table-column>

          <el-table-column label="比赛名称" min-width="260" align="center" header-align="center">
            <template #default="{ row }">
              <router-link :to="`/contest/${row.id}`" class="contest-title" @click.stop>
                {{ row.title }}
              </router-link>
            </template>
          </el-table-column>

	          <el-table-column label="赛制" width="100" align="center" header-align="center">
	            <template #default="{ row }">
	              <span class="rule-type">{{ row.type?.toUpperCase() }}</span>
	            </template>
	          </el-table-column>
	          
	          <el-table-column label="计时" width="180" align="center" header-align="center">
	            <template #default="{ row }">
	              <span class="time-mode-text">
	                {{ row.timing_mode === 'window' ? `窗口期 + ${row.duration_minutes || 0} 分钟` : '固定起止' }}
	              </span>
	            </template>
	          </el-table-column>
	          <el-table-column label="开始时间" width="180" align="center" header-align="center">
            <template #default="{ row }">
              <span class="time-text">{{ formatDate(row.start_at) }}</span>
            </template>
          </el-table-column>
          
          <el-table-column label="结束时间" width="180" align="center" header-align="center">
            <template #default="{ row }">
              <span class="time-text">{{ formatDate(row.end_at) }}</span>
            </template>
          </el-table-column>
          
          <el-table-column label="题目数" width="100" align="center" header-align="center">
            <template #default="{ row }">
              <span class="count-text">{{ row.problem_count || 0 }}</span>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="120" align="center" header-align="center">
            <template #default="{ row }">
              <span :class="['status-pill', getStatusClass(row)]">
                {{ getStatusLabel(row) }}
              </span>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else description="暂无比赛" />
      </div>

      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchContests"
          @current-change="fetchContests"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { contestApi } from '@/api/contest'

const loading = ref(false)
const contests = ref([])
const router = useRouter()

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

const pageRunningCount = computed(
  () => contests.value.filter((contest) => getStatus(contest) === 'running').length
)

const pageUpcomingCount = computed(
  () => contests.value.filter((contest) => getStatus(contest) === 'upcoming').length
)

function formatDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

function getStatus(contest) {
  const now = Date.now()
  const startAt = new Date(contest.start_at).getTime()
  const endAt = new Date(contest.end_at).getTime()
  if (now < startAt) return 'upcoming'
  if (now > endAt) return 'ended'
  return 'running'
}

function getStatusLabel(contest) {
  const map = {
    upcoming: '未开始',
    running: '进行中',
    ended: '已结束',
  }
  return map[getStatus(contest)]
}

function getStatusClass(contest) {
  const map = {
    upcoming: 'status-upcoming',
    running: 'status-running',
    ended: 'status-ended',
  }
  return map[getStatus(contest)]
}

function goContestDetail(row) {
  router.push(`/contest/${row.id}`)
}

async function fetchContests() {
  loading.value = true
  try {
    const res = await contestApi.getList({
      page: pagination.page,
      size: pagination.size,
    })
    contests.value = res.data.list || []
    pagination.total = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchContests()
})
</script>

<style lang="scss" scoped>
.contest-list-wrapper {
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

.contest-title {
  color: var(--swiss-text-main);
  font-weight: 500;
  text-decoration: none;
  font-size: 15px;
  transition: color 0.2s;

  &:hover {
    color: var(--swiss-primary);
  }
}

.rule-type {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--swiss-text-secondary);
  background: rgba(0,0,0,0.04);
  padding: 2px 6px;
  border-radius: 4px;
}

.time-mode-text {
  font-size: 12px;
  color: var(--swiss-text-secondary);
}

.time-text {
  font-size: 13px;
  color: var(--swiss-text-secondary);
  white-space: nowrap;
}

.count-text {
  font-size: 13px;
  color: var(--swiss-text-main);
}

.status-pill {
  display: inline-block;
  border-radius: 999px;
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 600;
}

.status-upcoming {
  color: #1d4ed8;
  background: #dbeafe;
}

.status-running {
  color: #166534;
  background: #dcfce7;
}

.status-ended {
  color: #6b7280;
  background: #f3f4f6;
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
}
</style>
