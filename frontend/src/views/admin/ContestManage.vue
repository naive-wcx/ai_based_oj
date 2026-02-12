<template>
  <div class="contest-manage">
    <div class="page-header">
      <div>
        <h2 class="page-title">比赛管理</h2>
        <p class="page-subtitle">维护比赛时间、赛制和题目编排。</p>
      </div>
      <el-button type="primary" @click="$router.push('/admin/contest/create')">
        创建比赛
      </el-button>
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

    <div class="card table-card" v-loading="loading">
      <el-table v-if="contests.length" :data="contests" stripe class="swiss-table">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="比赛名称" min-width="240" />
	        <el-table-column label="赛制" width="100">
	          <template #default="{ row }">
	            <el-tag size="small" :type="row.type === 'oi' ? 'warning' : 'success'">
	              {{ row.type?.toUpperCase() }}
	            </el-tag>
	          </template>
	        </el-table-column>
	        <el-table-column label="计时" width="170">
	          <template #default="{ row }">
	            <span v-if="row.timing_mode === 'window'">窗口期 + {{ row.duration_minutes || 0 }} 分钟</span>
	            <span v-else>固定起止</span>
	          </template>
	        </el-table-column>
        <el-table-column label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.start_at) }}
          </template>
        </el-table-column>
        <el-table-column label="结束时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.end_at) }}
          </template>
        </el-table-column>
        <el-table-column label="题目数" width="100">
          <template #default="{ row }">
            {{ row.problem_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <span :class="['status-pill', getStatusClass(row)]">
              {{ getStatusLabel(row) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="$router.push(`/admin/contest/${row.id}/edit`)">
              编辑
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="暂无比赛数据" />

      <div class="pagination" v-if="pagination.total > 0">
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
import { ElMessageBox } from 'element-plus'
import { message } from '@/utils/message'
import { contestApi } from '@/api/contest'
import { adminApi } from '@/api/admin'

const loading = ref(false)
const contests = ref([])

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
  const statusMap = {
    upcoming: '未开始',
    running: '进行中',
    ended: '已结束',
  }
  return statusMap[getStatus(contest)]
}

function getStatusClass(contest) {
  const classMap = {
    upcoming: 'status-upcoming',
    running: 'status-running',
    ended: 'status-ended',
  }
  return classMap[getStatus(contest)]
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

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除比赛 "${row.title}" 吗？`, '提示', {
      type: 'warning',
    })
    await adminApi.deleteContest(row.id)
    message.success('删除成功')
    fetchContests()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

onMounted(() => {
  fetchContests()
})
</script>

<style lang="scss" scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 14px;
}

.page-title {
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
  margin-bottom: 16px;
}

.stat-chip {
  min-width: 92px;
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
    line-height: 1;
    color: var(--swiss-text-main);
  }
}

.table-card {
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  background: #fff;
  padding: 12px;
  min-height: 240px;
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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

@media (max-width: 1024px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-stats {
    flex-wrap: wrap;
  }
}
</style>
