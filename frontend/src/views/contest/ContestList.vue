<template>
  <div class="contest-list">
    <h1 class="page-title">比赛</h1>

    <div class="card">
      <el-table :data="contests" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="比赛名称" min-width="260">
          <template #default="{ row }">
            <router-link :to="`/contest/${row.id}`" class="contest-title">
              {{ row.title }}
            </router-link>
          </template>
        </el-table-column>
        <el-table-column label="赛制" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.type === 'oi' ? 'warning' : 'success'">
              {{ row.type?.toUpperCase() }}
            </el-tag>
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
      </el-table>

      <div class="pagination">
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
  return date.toLocaleString()
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
  color: #303133;
  font-weight: 500;
  text-decoration: none;

  &:hover {
    color: #409eff;
  }
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
