<template>
  <div class="swiss-layout">
    <div class="swiss-header">
      <h1 class="swiss-title">题目列表</h1>
      <div class="filter-group">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索题目..."
          style="width: 200px"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <el-select
          v-model="filters.difficulty"
          placeholder="难度"
          clearable
          style="width: 120px"
          @change="handleSearch"
        >
          <el-option label="简单" value="easy" />
          <el-option label="中等" value="medium" />
          <el-option label="困难" value="hard" />
        </el-select>
      </div>
    </div>

    <el-table 
      :data="problems" 
      v-loading="loading" 
      class="swiss-table"
    >
      <el-table-column prop="id" label="#" width="80" align="center">
        <template #default="{ row }">
          <span class="swiss-font-mono" style="color: var(--color-text-secondary)">{{ row.id }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="状态" width="60" align="center">
        <template #default="{ row }">
          <el-icon v-if="row.has_accepted" style="color: var(--color-success); font-size: 16px;"><Check /></el-icon>
        </template>
      </el-table-column>

      <el-table-column label="标题" min-width="400">
        <template #default="{ row }">
          <router-link :to="`/problem/${row.id}`" class="problem-link">
            {{ row.title }}
            <span v-if="row.has_ai_judge" class="indicator ai" title="AI 判题已启用">AI</span>
            <span v-if="row.has_file_io" class="indicator file" title="文件 IO">FILE</span>
          </router-link>
        </template>
      </el-table-column>

      <el-table-column label="标签" min-width="200">
        <template #default="{ row }">
          <div class="tags-wrapper">
            <span v-for="tag in (row.tags || []).slice(0, 3)" :key="tag" class="minimal-tag">
              {{ tag }}
            </span>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="通过率" width="120" align="right">
        <template #default="{ row }">
          <span class="swiss-font-mono" style="font-size: 13px;">{{ getAcceptRate(row) }}</span>
        </template>
      </el-table-column>

      <el-table-column label="难度" width="100" align="right">
        <template #default="{ row }">
          <span :class="['difficulty-dot', row.difficulty]"></span>
          <span style="font-size: 13px; color: var(--color-text-secondary);">{{ formatDifficulty(row.difficulty) }}</span>
        </template>
      </el-table-column>
    </el-table>

    <div class="swiss-pagination" v-if="pagination.total > 0">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="prev, pager, next"
        @size-change="fetchProblems"
        @current-change="fetchProblems"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search, Check } from '@element-plus/icons-vue'
import { problemApi } from '@/api/problem'
import { message } from '@/utils/message'
import { useRouter, useRoute } from 'vue-router'

const loading = ref(true)
const problems = ref([])
const router = useRouter()
const route = useRoute()

const filters = reactive({
  keyword: route.query.keyword || '',
  difficulty: route.query.difficulty || '',
  tag: route.query.tag || '',
})

const pagination = reactive({
  page: parseInt(route.query.page, 10) || 1,
  size: parseInt(route.query.size, 10) || 50,
  total: 0,
})

const formatDifficulty = (val) => {
  const map = { easy: '简单', medium: '中等', hard: '困难' }
  return map[val] || val
}

const updateUrl = () => {
  router.push({
    query: {
      page: pagination.page,
      size: pagination.size,
      ...(filters.keyword && { keyword: filters.keyword }),
      ...(filters.difficulty && { difficulty: filters.difficulty }),
      ...(filters.tag && { tag: filters.tag }),
    },
  })
}

async function fetchProblems() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      keyword: filters.keyword,
      difficulty: filters.difficulty,
      tag: filters.tag,
    }
    const res = await problemApi.getList(params)
    problems.value = res.data.list || []
    pagination.total = res.data.total
    updateUrl()
  } catch (e) {
    message.error('加载题目失败')
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchProblems()
}

function getAcceptRate(row) {
  if (!row.submit_count) return '-'
  const rate = ((row.accepted_count / row.submit_count) * 100).toFixed(1)
  return `${rate}%`
}

onMounted(() => {
  fetchProblems()
})
</script>

<style lang="scss" scoped>
.filter-group {
  display: flex;
  gap: 24px;
}

.problem-link {
  font-size: 15px;
  font-weight: 500;
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
  
  &:hover {
    color: var(--color-primary);
  }
}

.indicator {
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 2px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  
  &.ai {
    background: #EEF2FF;
    color: #4338CA;
  }
  
  &.file {
    background: #FEF3C7;
    color: #D97706;
  }
}

.tags-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.minimal-tag {
  font-size: 12px;
  color: var(--color-text-secondary);
  background: rgba(0,0,0,0.04);
  padding: 2px 8px;
  border-radius: 12px;
}

.difficulty-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 8px;
  vertical-align: middle;
  
  &.easy { background-color: var(--color-success); }
  &.medium { background-color: var(--color-warning); }
  &.hard { background-color: var(--color-danger); }
}
</style>