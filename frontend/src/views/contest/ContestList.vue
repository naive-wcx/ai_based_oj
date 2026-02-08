<template>
  <div class="contest-list-wrapper">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">比赛列表</h1>
      </div>

      <div class="table-container">
        <el-table :data="contests" v-loading="loading" class="swiss-table">
          <el-table-column prop="id" label="ID" width="80" align="center" header-align="center">
            <template #default="{ row }">
              <span class="id-text">{{ row.id }}</span>
            </template>
          </el-table-column>
          
          <el-table-column label="比赛名称" min-width="260" align="center" header-align="center">
            <template #default="{ row }">
              <router-link :to="`/contest/${row.id}`" class="contest-title">
                {{ row.title }}
              </router-link>
            </template>
          </el-table-column>
          
          <el-table-column label="赛制" width="100" align="center" header-align="center">
            <template #default="{ row }">
              <span class="rule-type">{{ row.type?.toUpperCase() }}</span>
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
        </el-table>
      </div>

      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="prev, pager, next"
          @size-change="fetchContests"
          @current-change="fetchContests"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { contestApi } from '@/api/contest'

const loading = ref(false)
const contests = ref([])

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

function formatDate(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
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

.time-text {
  font-size: 13px;
  color: var(--swiss-text-secondary);
  white-space: nowrap;
}

.count-text {
  font-size: 13px;
  color: var(--swiss-text-main);
}

.pagination-wrapper {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}
</style>
