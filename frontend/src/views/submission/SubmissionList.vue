<template>
  <div class="submission-list">
    <h1 class="page-title">提交记录</h1>
    
    <div class="card">
      <el-table :data="submissions" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="题目" min-width="200">
          <template #default="{ row }">
            <router-link :to="`/problem/${row.problem_id}`">
              {{ row.problem_id }}. {{ row.problem_title }}
            </router-link>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="150">
          <template #default="{ row }">
            <router-link :to="`/submission/${row.id}`">
              <StatusBadge :status="row.status" />
            </router-link>
          </template>
        </el-table-column>
        <el-table-column label="语言" width="100">
          <template #default="{ row }">
            {{ getLanguageLabel(row.language) }}
          </template>
        </el-table-column>
        <el-table-column label="用时" width="100">
          <template #default="{ row }">
            {{ row.time_used ? `${row.time_used}ms` : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="内存" width="100">
          <template #default="{ row }">
            {{ row.memory_used ? formatMemory(row.memory_used) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户" width="120" />
        <el-table-column label="提交时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
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
          @size-change="fetchSubmissions"
          @current-change="fetchSubmissions"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { submissionApi } from '@/api/submission'
import StatusBadge from '@/components/problem/StatusBadge.vue'

const loading = ref(false)
const submissions = ref([])

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

const languageLabels = {
  c: 'C',
  cpp: 'C++',
  python: 'Python',
  java: 'Java',
  go: 'Go',
}

function getLanguageLabel(lang) {
  return languageLabels[lang] || lang
}

function formatMemory(kb) {
  if (kb < 1024) return `${kb}KB`
  return `${(kb / 1024).toFixed(1)}MB`
}

function formatTime(time) {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN')
}

async function fetchSubmissions() {
  loading.value = true
  try {
    const res = await submissionApi.getList({
      page: pagination.page,
      size: pagination.size,
    })
    submissions.value = res.data.list || []
    pagination.total = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchSubmissions()
})
</script>

<style lang="scss" scoped>
.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
