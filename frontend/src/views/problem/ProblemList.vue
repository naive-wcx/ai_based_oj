<template>
  <el-card shadow="never" class="page-container">
    <template #header>
      <div class="page-header">
        <span class="page-title">题目列表</span>
        <div class="filter-bar">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索题目 ID 或标题"
            clearable
            style="width: 240px"
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
    </template>

    <el-table :data="problems" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="#" width="80" />
      <el-table-column label="题目" min-width="300">
        <template #default="{ row }">
          <router-link :to="`/problem/${row.id}`" class="problem-title">
            <span>{{ row.title }}</span>
            <span v-if="row.has_ai_judge" class="ai-badge">AI</span>
            <el-tag
              v-if="row.has_accepted"
              type="success"
              size="small"
              disable-transitions
              >已通过</el-tag
            >
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="标签" min-width="200">
        <template #default="{ row }">
          <el-tag
            v-for="tag in (row.tags || []).slice(0, 3)"
            :key="tag"
            size="small"
            effect="plain"
            class="problem-tag"
          >
            {{ tag }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="通过率" width="120" align="center">
        <template #default="{ row }">
          <span class="accept-rate">{{ getAcceptRate(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="难度" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="getDifficultyTag(row.difficulty).type" size="small">{{
            getDifficultyTag(row.difficulty).label
          }}</el-tag>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container" v-if="pagination.total > 0">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="fetchProblems"
        @current-change="fetchProblems"
      />
    </div>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { problemApi } from '@/api/problem'
import { ElMessage } from 'element-plus'
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

const getDifficultyTag = (difficulty) => {
  const settings = {
    easy: { type: 'success', label: '简单' },
    medium: { type: 'warning', label: '中等' },
    hard: { type: 'danger', label: '困难' },
  }
  return settings[difficulty] || { type: 'info', label: difficulty }
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
    ElMessage.error('题目列表加载失败')
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
  if (!row.submit_count) return 'N/A'
  const rate = ((row.accepted_count / row.submit_count) * 100).toFixed(1)
  return `${rate}%`
}

onMounted(() => {
  fetchProblems()
})
</script>

<style lang="scss" scoped>
.page-container {
  border: none;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

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
  color: var(--el-text-color-primary);
  text-decoration: none;
  font-weight: 500;

  &:hover {
    color: var(--el-color-primary);
  }
}

.ai-badge {
  display: inline-block;
  padding: 0 6px;
  font-size: 12px;
  font-weight: bold;
  line-height: 18px;
  border-radius: 4px;
  color: #fff;
  background-color: #409eff;
  border: 1px solid #409eff;
}

.problem-tag {
  margin-right: 4px;
}

.accept-rate {
  color: var(--el-text-color-secondary);
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
