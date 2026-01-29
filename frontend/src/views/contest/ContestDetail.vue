<template>
  <div class="contest-detail" v-loading="loading">
    <div class="header card" v-if="contest">
      <div class="header-main">
        <h1>{{ contest.title }}</h1>
        <el-tag size="small" :type="contest.type === 'oi' ? 'warning' : 'success'">
          {{ contest.type?.toUpperCase() }}
        </el-tag>
      </div>
      <div class="header-meta">
        <span>开始时间：{{ formatDate(contest.start_at) }}</span>
        <span>结束时间：{{ formatDate(contest.end_at) }}</span>
      </div>
      <div class="header-score" v-if="myTotal !== null">
        我的总分：{{ myTotal }}
      </div>
      <p v-if="contest.description" class="desc">{{ contest.description }}</p>
    </div>

    <div class="card">
      <h3>比赛题目</h3>
      <el-table :data="problems" stripe>
        <el-table-column prop="id" label="编号" width="80" />
        <el-table-column label="题目" min-width="300">
          <template #default="{ row }">
            <router-link :to="`/problem/${row.id}`" class="problem-link">
              {{ row.title }}
            </router-link>
            <el-tag v-if="row.has_accepted" type="success" size="small" class="accepted-tag">已通过</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="难度" width="120">
          <template #default="{ row }">
            <DifficultyBadge :difficulty="row.difficulty" />
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="card" v-if="userStore.isAdmin">
      <div class="leaderboard-header">
        <h3>比赛排行榜</h3>
        <el-button type="primary" :loading="exporting" @click="handleExport">
          导出成绩
        </el-button>
      </div>
      <el-table :data="leaderboardEntries" v-loading="leaderboardLoading" stripe>
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column prop="username" label="用户名" width="160" />
        <el-table-column prop="group" label="分组" width="120" />
        <el-table-column prop="total" label="总分" width="100" />
        <el-table-column
          v-for="(pid, index) in leaderboardProblemIds"
          :key="pid"
          :label="`P${pid}`"
          :min-width="100"
        >
          <template #default="{ row }">
            {{ row.scores?.[index] ?? 0 }}
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { contestApi } from '@/api/contest'
import { adminApi } from '@/api/admin'
import DifficultyBadge from '@/components/problem/DifficultyBadge.vue'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const loading = ref(false)
const contest = ref(null)
const problems = ref([])
const myTotal = ref(null)
const userStore = useUserStore()
const leaderboardLoading = ref(false)
const exporting = ref(false)
const leaderboardProblemIds = ref([])
const leaderboardEntries = ref([])

function formatDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

async function fetchContest() {
  loading.value = true
  try {
    const res = await contestApi.getById(route.params.id)
    contest.value = res.data.contest
    problems.value = res.data.problems || []
    myTotal.value = res.data.my_total ?? null
    if (userStore.isAdmin) {
      fetchLeaderboard()
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchLeaderboard() {
  leaderboardLoading.value = true
  try {
    const res = await adminApi.getContestLeaderboard(route.params.id)
    leaderboardProblemIds.value = res.data.problem_ids || []
    leaderboardEntries.value = res.data.entries || []
  } catch (e) {
    console.error(e)
  } finally {
    leaderboardLoading.value = false
  }
}

async function handleExport() {
  exporting.value = true
  try {
    const res = await adminApi.exportContestLeaderboard(route.params.id)
    const blob = new Blob([res.data], { type: 'text/csv;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `contest_${route.params.id}_leaderboard.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success({ message: '导出成功', duration: 1000 })
  } catch (e) {
    console.error(e)
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  fetchContest()
})
</script>

<style lang="scss" scoped>
.header {
  margin-bottom: 20px;
}

.header-main {
  display: flex;
  align-items: center;
  gap: 12px;

  h1 {
    margin: 0;
  }
}

.header-meta {
  margin-top: 8px;
  color: #909399;
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.header-score {
  margin-top: 6px;
  font-weight: 600;
  color: #303133;
}

.desc {
  margin-top: 12px;
  color: #606266;
}

.problem-link {
  color: #303133;
  font-weight: 500;
  text-decoration: none;

  &:hover {
    color: #409eff;
  }
}

.accepted-tag {
  margin-left: 8px;
}

.leaderboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
</style>
