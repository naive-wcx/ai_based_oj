<template>
  <div class="swiss-layout">
    <div class="swiss-header">
      <h1 class="swiss-title">比赛列表</h1>
    </div>

    <el-table :data="contests" v-loading="loading" class="swiss-table">
      <el-table-column prop="id" label="ID" width="80" align="center">
        <template #default="{ row }">
          <span class="swiss-font-mono" style="color: var(--color-text-secondary)">{{ row.id }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="比赛名称" min-width="260">
        <template #default="{ row }">
          <router-link :to="`/contest/${row.id}`" class="contest-title">
            {{ row.title }}
          </router-link>
        </template>
      </el-table-column>
      
      <el-table-column label="赛制" width="100" align="center">
        <template #default="{ row }">
          <span class="rule-type">{{ row.type?.toUpperCase() }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="开始时间" width="180" align="right">
        <template #default="{ row }">
          <span class="swiss-font-mono time-text">{{ formatDate(row.start_at) }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="结束时间" width="180" align="right">
        <template #default="{ row }">
          <span class="swiss-font-mono time-text">{{ formatDate(row.end_at) }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="题目数" width="100" align="center">
        <template #default="{ row }">
          <span class="swiss-font-mono">{{ row.problem_count || 0 }}</span>
        </template>
      </el-table-column>
    </el-table>

    <div class="swiss-pagination">
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
.contest-title {
  color: var(--color-text-primary);
  font-weight: 500;
  text-decoration: none;
  font-size: 15px;

  &:hover {
    color: var(--color-primary);
  }
}

.rule-type {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--color-text-secondary);
  background: rgba(0,0,0,0.04);
  padding: 2px 6px;
  border-radius: 4px;
}

.time-text {
  font-size: 13px;
  color: var(--color-text-secondary);
}
</style>