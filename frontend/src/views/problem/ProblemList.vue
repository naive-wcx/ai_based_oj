<template>
  <div class="problem-list">
    <h1 class="page-title">题目列表</h1>
    
    <!-- 筛选条件 -->
    <div class="filter-bar card">
      <el-input
        v-model="filters.keyword"
        placeholder="搜索题目..."
        clearable
        style="width: 300px"
        @keyup.enter="fetchProblems"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      
      <el-select v-model="filters.difficulty" placeholder="难度" clearable style="width: 120px">
        <el-option label="简单" value="easy" />
        <el-option label="中等" value="medium" />
        <el-option label="困难" value="hard" />
      </el-select>
      
      <el-button type="primary" @click="fetchProblems">搜索</el-button>
    </div>
    
    <!-- 题目表格 -->
    <div class="card">
      <el-table :data="problems" v-loading="loading" stripe>
        <el-table-column prop="id" label="编号" width="80" />
        <el-table-column label="题目" min-width="300">
          <template #default="{ row }">
            <router-link :to="`/problem/${row.id}`" class="problem-title">
              {{ row.title }}
              <span v-if="row.has_ai_judge" class="ai-badge">AI</span>
            </router-link>
          </template>
        </el-table-column>
        <el-table-column label="难度" width="100">
          <template #default="{ row }">
            <DifficultyBadge :difficulty="row.difficulty" />
          </template>
        </el-table-column>
        <el-table-column label="标签" width="200">
          <template #default="{ row }">
            <el-tag
              v-for="tag in (row.tags || []).slice(0, 3)"
              :key="tag"
              size="small"
              style="margin-right: 4px"
            >
              {{ tag }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="通过率" width="120">
          <template #default="{ row }">
            {{ getAcceptRate(row) }}
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchProblems"
          @current-change="fetchProblems"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { problemApi } from '@/api/problem'
import DifficultyBadge from '@/components/problem/DifficultyBadge.vue'

const loading = ref(false)
const problems = ref([])

const filters = reactive({
  keyword: '',
  difficulty: '',
  tag: '',
})

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

async function fetchProblems() {
  loading.value = true
  try {
    const res = await problemApi.getList({
      page: pagination.page,
      size: pagination.size,
      keyword: filters.keyword,
      difficulty: filters.difficulty,
      tag: filters.tag,
    })
    problems.value = res.data.list || []
    pagination.total = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function getAcceptRate(row) {
  if (!row.submit_count) return '-'
  const rate = (row.accepted_count / row.submit_count * 100).toFixed(1)
  return `${rate}% (${row.accepted_count}/${row.submit_count})`
}

onMounted(() => {
  fetchProblems()
})
</script>

<style lang="scss" scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.problem-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #303133;
  font-weight: 500;
  
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
