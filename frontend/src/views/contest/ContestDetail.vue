<template>
  <div class="contest-detail-wrapper" v-loading="loading">
    <div class="container" v-if="contest">
      <!-- 1. 页头：标题与状态 -->
	      <div class="detail-header">
	        <div class="header-left">
	          <router-link to="/contests" class="back-link">
	            ← 返回比赛列表
	          </router-link>
          <h1 class="page-title">
            {{ contest.title }}
          </h1>
	        </div>
	        <div class="header-right">
	          <el-button
	            v-if="canStartWindowContest"
	            type="primary"
	            :loading="startingContest"
	            @click="handleStartContest"
	          >
	            开始比赛（{{ contest.duration_minutes }} 分钟）
	          </el-button>
	           <div class="status-badge" :class="getContestStatusClass(contest)">
	             {{ getContestStatusLabel(contest) }}
	           </div>
	        </div>
	      </div>

      <!-- 2. 核心指标仪表盘 -->
      <div class="stats-dashboard">
        <div class="stat-card">
          <div class="stat-label">赛制</div>
          <div class="stat-value">{{ contest.type?.toUpperCase() }}</div>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-card">
          <div class="stat-label">开始时间</div>
          <div class="stat-value time-value">{{ formatDate(contest.start_at) }}</div>
        </div>
        <div class="stat-divider"></div>
	        <div class="stat-card">
	          <div class="stat-label">结束时间</div>
	          <div class="stat-value time-value">{{ formatDate(contest.end_at) }}</div>
	        </div>
	        <div class="stat-divider"></div>
	        <div class="stat-card">
	          <div class="stat-label">计时模式</div>
	          <div class="stat-value">
	            {{ contest.timing_mode === 'window' ? `窗口期 + ${contest.duration_minutes || 0} 分钟` : '固定起止' }}
	          </div>
	        </div>
	        <div class="stat-divider"></div>
	        <div class="stat-card">
	          <div class="stat-label">我的赛时|赛后</div>
	          <div class="stat-value time-value">
	            {{ myLiveTotal != null || myPostTotal != null ? `${myLiveTotal ?? 0} | ${myPostTotal ?? 0}` : '-' }}
	          </div>
	        </div>
	        <div class="stat-divider" v-if="showUserRemaining"></div>
	        <div class="stat-card" v-if="showUserRemaining">
	          <div class="stat-label">剩余时间</div>
	          <div class="stat-value time-value">
	            {{ formatRemaining(sessionState?.remaining_seconds || 0) }}
	          </div>
	        </div>
	      </div>

	      <div class="session-box" v-if="contest.timing_mode === 'window' && !userStore.isAdmin">
	        <div class="session-item">
	          <span class="session-label">个人状态</span>
	          <span class="session-value">
	            {{ sessionState?.started ? (sessionState?.in_live ? '进行中' : '已结束') : (canStartWindowContest ? '可开始' : '未开始') }}
	          </span>
	        </div>
	        <div class="session-item">
	          <span class="session-label">个人开始</span>
	          <span class="session-value">{{ formatDate(sessionState?.start_at) }}</span>
	        </div>
	        <div class="session-item">
	          <span class="session-label">个人结束</span>
	          <span class="session-value">{{ formatDate(sessionState?.end_at) }}</span>
	        </div>
	        <div class="session-item">
	          <span class="session-label">剩余时间</span>
	          <span class="session-value">{{ formatRemaining(sessionState?.remaining_seconds || 0) }}</span>
	        </div>
	      </div>

      <!-- 3. 比赛描述 -->
      <div class="section-block description-section" v-if="contest.description">
        <h3 class="section-title">比赛说明</h3>
        <div class="description-box">
          <MarkdownPreview :content="contest.description" />
        </div>
      </div>

      <!-- 4. 题目列表 -->
      <div class="section-block">
        <h3 class="section-title">题目列表</h3>
        <div class="table-wrapper">
          <el-table :data="problems" class="swiss-table">
            <!-- 编号：居中 -->
            <el-table-column prop="id" label="编号" width="80" align="center" header-align="center" />
            
            <!-- 题目 -->
            <el-table-column label="题目" min-width="300" align="center" header-align="center">
              <template #default="{ row }">
                <router-link :to="`/problem/${row.id}`" class="table-link">
                  {{ row.title }}
                </router-link>
              </template>
            </el-table-column>
            
            <!-- 状态：居中 -->
            <el-table-column label="状态" width="120" align="center" header-align="center">
              <template #default="{ row }">
                <span v-if="row.has_accepted" class="status-dot success" title="已通过">●</span>
                <span v-else-if="row.has_submitted" class="status-dot warning" title="已提交">●</span>
                <span v-else class="status-dot pending" title="未尝试">●</span>
              </template>
            </el-table-column>
            
	            <!-- 难度：居中 -->
	            <el-table-column label="难度" width="120" align="center" header-align="center">
	              <template #default="{ row }">
	                <DifficultyBadge v-if="row.difficulty" :difficulty="row.difficulty" />
	                <span v-else>-</span>
	              </template>
	            </el-table-column>
          </el-table>
        </div>
      </div>

	      <!-- 5. 管理员排行榜 (仅管理员可见) -->
	      <div class="section-block" v-if="userStore.isAdmin">
	        <div class="section-header">
	          <h3 class="section-title">排行榜 (管理员)</h3>
	          <div class="leaderboard-actions">
	            <el-radio-group v-model="leaderboardMode" size="small" @change="fetchLeaderboard">
	              <el-radio-button label="combined">赛时|赛后</el-radio-button>
	              <el-radio-button label="live">赛时</el-radio-button>
	              <el-radio-button label="post">赛后</el-radio-button>
	            </el-radio-group>
	            <el-button size="small" :loading="exporting" @click="handleExport">导出成绩 CSV</el-button>
	          </div>
	        </div>
	        <div class="table-wrapper">
	          <el-table :data="leaderboardEntries" v-loading="leaderboardLoading" class="swiss-table">
            <el-table-column prop="user_id" label="ID" width="80" align="center" header-align="center" />
            <el-table-column prop="username" label="用户" width="160" align="center" header-align="center" />
            <el-table-column prop="group" label="分组" width="120" align="center" header-align="center" />
            <el-table-column
              v-if="showAdminWindowRemaining"
              label="剩余时间"
              width="140"
              align="center"
              header-align="center"
            >
              <template #default="{ row }">
                {{ getAdminWindowRemainingLabel(row) }}
              </template>
            </el-table-column>
		            <el-table-column prop="total" :label="leaderboardMode === 'combined' ? '赛时|赛后' : '总分'" width="130" align="center" header-align="center">
		              <template #default="{ row }">
		                <span class="score-cell" v-if="leaderboardMode === 'combined'">
		                  {{ row.live_total }} | {{ row.post_total }}
	                </span>
	                <span class="score-cell" v-else>{{ row.total }}</span>
	              </template>
	            </el-table-column>
            <el-table-column
              v-for="(pid, index) in leaderboardProblemIds"
              :key="pid"
              :label="`P${pid}`"
              :min-width="80"
              align="center"
              header-align="center"
	            >
	              <template #default="{ row }">
	                <span v-if="leaderboardMode === 'combined'">
	                  <span :class="getScoreClass(row.live_scores?.[index])">{{ row.live_scores?.[index] ?? 0 }}</span>
	                  |
	                  <span :class="getScoreClass(row.post_scores?.[index])">{{ row.post_scores?.[index] ?? 0 }}</span>
	                </span>
	                <span v-else :class="getScoreClass(row.scores?.[index])">{{ row.scores?.[index] ?? '-' }}</span>
	              </template>
	            </el-table-column>
	          </el-table>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useRoute } from 'vue-router'
import { message } from '@/utils/message'
import { contestApi } from '@/api/contest'
import { adminApi } from '@/api/admin'
import DifficultyBadge from '@/components/problem/DifficultyBadge.vue'
import MarkdownPreview from '@/components/common/MarkdownPreview.vue'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const loading = ref(false)
const contest = ref(null)
const problems = ref([])
const myLiveTotal = ref(null)
const myPostTotal = ref(null)
const sessionState = ref(null)
const userStore = useUserStore()
const leaderboardLoading = ref(false)
const exporting = ref(false)
const startingContest = ref(false)
const leaderboardProblemIds = ref([])
const leaderboardEntries = ref([])
const leaderboardMode = ref('combined')
const countdownTimer = ref(null)
const clockNow = ref(Date.now())

const canStartWindowContest = computed(() =>
  !userStore.isAdmin &&
  contest.value?.timing_mode === 'window' &&
  !!sessionState.value?.can_start
)

const showUserRemaining = computed(() =>
  !userStore.isAdmin && !!sessionState.value?.in_live
)

const showAdminWindowRemaining = computed(() => {
  if (!userStore.isAdmin || !contest.value) return false
  if (contest.value.timing_mode !== 'window') return false
  const endAt = new Date(contest.value.end_at)
  if (Number.isNaN(endAt.getTime())) return false
  return clockNow.value <= endAt.getTime()
})

function formatDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function formatRemaining(seconds) {
  const s = Number(seconds || 0)
  if (s <= 0) return '-'
  const hours = Math.floor(s / 3600)
  const minutes = Math.floor((s % 3600) / 60)
  const remainSeconds = s % 60
  return `${hours}h ${minutes}m ${remainSeconds}s`
}

function getAdminWindowRemainingLabel(row) {
  if (!row?.started_at) return '未开始'
  const startedAt = new Date(row.started_at)
  if (Number.isNaN(startedAt.getTime())) return '未开始'
  const durationSeconds = Number(contest.value?.duration_minutes || 0) * 60
  if (durationSeconds <= 0) return '-'
  const elapsedSeconds = Math.max(0, Math.floor((clockNow.value - startedAt.getTime()) / 1000))
  const remainingSeconds = durationSeconds - elapsedSeconds
  if (remainingSeconds <= 0) return '已结束'
  return formatRemaining(remainingSeconds)
}

function getContestStatusClass(contest) {
  const now = new Date()
  const start = new Date(contest.start_at)
  const end = new Date(contest.end_at)
  
  if (now < start) return 'info' // 未开始
  if (now > end) return 'info'   // 已结束
  return 'success'               // 进行中
}

function getContestStatusLabel(contest) {
  const now = new Date()
  const start = new Date(contest.start_at)
  const end = new Date(contest.end_at)
  
  if (now < start) return '未开始'
  if (now > end) return '已结束'
  return '进行中'
}

function getScoreClass(score) {
  if (!score) return 'score-gray'
  if (score === 100) return 'score-green'
  return 'score-orange'
}

async function fetchContest() {
  loading.value = true
  try {
    const res = await contestApi.getById(route.params.id)
    contest.value = res.data.contest
    problems.value = res.data.problems || []
    sessionState.value = res.data.session || null
    myLiveTotal.value = res.data.my_live_total ?? null
    myPostTotal.value = res.data.my_post_total ?? null
    leaderboardEntries.value = []
    leaderboardProblemIds.value = []
    if (userStore.isAdmin) {
      await fetchLeaderboard()
    }
    setupCountdownTimer()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchLeaderboard() {
  leaderboardLoading.value = true
  try {
    const res = await adminApi.getContestLeaderboard(route.params.id, {
      board_mode: leaderboardMode.value,
    })
    leaderboardProblemIds.value = res.data.problem_ids || []
    leaderboardEntries.value = res.data.entries || []
    leaderboardMode.value = res.data.board_mode || leaderboardMode.value
  } catch (e) {
    console.error(e)
  } finally {
    leaderboardLoading.value = false
  }
}

function clearCountdownTimer() {
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value)
    countdownTimer.value = null
  }
}

function setupCountdownTimer() {
  clearCountdownTimer()
  if (!showUserRemaining.value && !showAdminWindowRemaining.value) return
  countdownTimer.value = setInterval(() => {
    clockNow.value = Date.now()
    if (showUserRemaining.value && sessionState.value) {
      const current = Number(sessionState.value?.remaining_seconds || 0)
      if (current <= 0) {
        sessionState.value = { ...sessionState.value, in_live: false, remaining_seconds: 0 }
      } else {
        sessionState.value = { ...sessionState.value, remaining_seconds: current - 1 }
      }
    }
    if (!showUserRemaining.value && !showAdminWindowRemaining.value) {
      clearCountdownTimer()
    }
  }, 1000)
}

async function handleStartContest() {
  if (!canStartWindowContest.value) return
  startingContest.value = true
  try {
    await contestApi.start(route.params.id)
    message.success({ message: '比赛已开始，祝你好运！', duration: 1000 })
    await fetchContest()
  } catch (e) {
    console.error(e)
  } finally {
    startingContest.value = false
  }
}

async function handleExport() {
  exporting.value = true
  try {
    const res = await adminApi.exportContestLeaderboard(route.params.id, {
      board_mode: leaderboardMode.value,
    })
    const blob = new Blob([res.data], { type: 'text/csv;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `contest_${route.params.id}_leaderboard_${leaderboardMode.value}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    message.success({ message: '导出成功', duration: 1000 })
  } catch (e) {
    console.error(e)
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  fetchContest()
})

onBeforeUnmount(() => {
  clearCountdownTimer()
})
</script>

<style lang="scss" scoped>
.contest-detail-wrapper {
  padding: 40px 0;
  background-color: var(--swiss-bg-base);
  min-height: 100vh;
}

.container {
  text-align: center;
}

/* Header */
.detail-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--swiss-border-light);
}

.header-left {
  text-align: center;
}

.header-right {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 12px;
}

.back-link {
  font-size: 14px;
  color: var(--swiss-text-secondary);
  margin-bottom: 12px;
  display: inline-block;
  font-weight: 500;
  transition: color 0.2s;
  &:hover { color: var(--swiss-primary); }
}

.page-title {
  font-size: 32px;
  margin: 0;
  color: var(--swiss-text-main);
  letter-spacing: -0.02em;
}

.status-badge {
  font-size: 15px;
  font-weight: 600;
  padding: 8px 20px;
  border-radius: var(--radius-xs);
  letter-spacing: 0.02em;
  
  &.success { color: var(--swiss-success); background: rgba(52, 199, 89, 0.1); }
  &.info { color: var(--swiss-text-secondary); background: rgba(0, 0, 0, 0.05); }
}

/* Stats Dashboard */
.stats-dashboard {
  display: flex;
  align-items: stretch;
  justify-content: center;
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 24px;
  margin-bottom: 30px;
  flex-wrap: wrap;
  gap: 20px;
}

.session-box {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  margin-bottom: 30px;
}

.session-item {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 4px;
}

.session-label {
  font-size: 12px;
  color: var(--swiss-text-secondary);
}

.session-value {
  font-size: 15px;
  font-weight: 600;
  color: var(--swiss-text-main);
}

.stat-card {
  flex: 1;
  min-width: 120px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  justify-content: space-between;
}

.stat-divider {
  width: 1px;
  background: var(--swiss-border-light);
  margin: 4px 0;
}

.stat-label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--swiss-text-secondary);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--swiss-text-main);
  
  &.score-value.highlight {
    color: var(--swiss-primary);
    font-size: 24px;
  }
  
  &.time-value {
    font-size: 16px;
  }
}

/* Common Section */
.section-block {
  margin-bottom: 40px;
}

.section-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 16px;
}

.leaderboard-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 10px;
}

.section-title {
  font-size: 18px;
  margin: 0 0 16px;
  color: var(--swiss-text-main);
  font-weight: 600;
  text-align: center;
}

.description-box {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  padding: 24px;
  font-size: 15px;
  color: var(--swiss-text-main);
  line-height: 1.6;
  text-align: left;
}

.description-section {
  text-align: left;
}

.description-box :deep(*) {
  text-align: left;
}

.announcement-block {
  .section-title {
    color: var(--swiss-primary);
  }
  .announcement-content {
    background: #f0f9eb;
    border-color: #e1f3d8;
  }
}

/* Table */
.table-wrapper {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.table-link {
  color: var(--swiss-text-main);
  font-weight: 500;
  transition: color 0.2s;
  
  &:hover { color: var(--swiss-primary); }
}

.status-dot {
  font-size: 12px;
  &.success { color: var(--swiss-success); }
  &.warning { color: var(--swiss-warning); }
  &.pending { color: var(--swiss-border); }
}

.score-cell {
  font-weight: 600;
}

.score-green { color: var(--swiss-success); font-weight: 600; }
.score-orange { color: var(--swiss-warning); }
.score-gray { color: var(--swiss-text-secondary); }

@media (max-width: 768px) {
  .stats-dashboard {
    flex-direction: column;
    gap: 16px;
  }
  
  .stat-divider {
    display: none;
  }
  
  .stat-card {
    flex-direction: row;
    align-items: center;
    border-bottom: 1px solid var(--swiss-border-light);
    padding-bottom: 12px;
    
    &:last-child { border-bottom: none; padding-bottom: 0; }
  }
  
  .stat-label { margin-bottom: 0; }
}
</style>
