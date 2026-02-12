<template>
  <div class="problem-list-wrapper">
    <div class="container">
      <div class="page-header">
        <div>
          <h1 class="page-title">题目列表</h1>
          <p class="page-subtitle">支持关键词、难度与标签筛选，点击行可快速进入题目详情。</p>
        </div>
        <div class="header-stats">
          <div class="stat-chip">
            <span>总题数</span>
            <strong>{{ pagination.total }}</strong>
          </div>
          <div class="stat-chip">
            <span>本页 AC</span>
            <strong>{{ pageAcceptedCount }}</strong>
          </div>
          <div class="stat-chip">
            <span>AI 题</span>
            <strong>{{ pageAICount }}</strong>
          </div>
        </div>
      </div>

      <div class="filter-panel">
        <el-input
          v-model="filters.keyword"
          clearable
          placeholder="搜索题目标题或描述"
          class="search-input"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <el-select
          v-model="filters.difficulty"
          placeholder="所有难度"
          clearable
          class="filter-select"
          @change="handleSearch"
        >
          <el-option label="简单" value="easy" />
          <el-option label="中等" value="medium" />
          <el-option label="困难" value="hard" />
        </el-select>

        <el-input
          v-model="filters.tag"
          clearable
          placeholder="标签（如 dp）"
          class="tag-input"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />

        <div class="filter-actions">
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button plain @click="handleReset">重置</el-button>
        </div>
      </div>

      <div class="table-container" v-loading="loading">
        <el-table
          v-if="problems.length"
          :data="problems"
          class="swiss-table"
          @row-click="goProblemDetail"
        >
          <el-table-column prop="id" label="编号" width="80" align="center" header-align="center">
            <template #default="{ row }">
              <span class="id-text">{{ row.id }}</span>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="80" align="center" header-align="center">
            <template #default="{ row }">
              <span v-if="row.has_accepted" class="status-text success">AC</span>
            </template>
          </el-table-column>

          <el-table-column label="标题" min-width="420" align="center" header-align="center">
            <template #default="{ row }">
              <div class="title-cell">
                <router-link :to="`/problem/${row.id}`" class="problem-link" @click.stop>
                  {{ row.title }}
                </router-link>
                <div class="badges">
                  <span v-if="row.is_public === false" class="badge hidden" title="仅管理员或比赛可见">隐藏</span>
                  <span v-if="row.has_ai_judge" class="badge ai">AI</span>
                  <span v-if="row.has_file_io" class="badge file">FILE</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="标签" min-width="140" align="center" header-align="center">
            <template #default="{ row }">
              <div class="tags-wrapper">
                <span v-for="tag in (row.tags || []).slice(0, 3)" :key="tag" class="text-tag">
                  #{{ tag }}
                </span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="通过率" width="110" align="center" header-align="center">
            <template #default="{ row }">
              <span class="rate-text">{{ getAcceptRate(row) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="难度" width="110" align="center" header-align="center">
            <template #default="{ row }">
              <div class="difficulty-wrapper">
                <span :class="['difficulty-dot', row.difficulty]"></span>
                <span class="difficulty-text">{{ formatDifficulty(row.difficulty) }}</span>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else description="暂无匹配题目，请调整筛选条件" />
      </div>

      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSearch"
          @current-change="fetchProblems"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { problemApi } from '@/api/problem'
import { message } from '@/utils/message'
import { useRouter, useRoute } from 'vue-router'

const loading = ref(true)
const problems = ref([])
const router = useRouter()
const route = useRoute()
const defaultPageSize = 50

const filters = reactive({
  keyword: route.query.keyword || '',
  difficulty: route.query.difficulty || '',
  tag: route.query.tag || '',
})

const pagination = reactive({
  page: parseInt(route.query.page, 10) || 1,
  size: parseInt(route.query.size, 10) || defaultPageSize,
  total: 0,
})

const pageAcceptedCount = computed(
  () => problems.value.filter((problem) => problem.has_accepted).length
)

const pageAICount = computed(
  () => problems.value.filter((problem) => problem.has_ai_judge).length
)

const formatDifficulty = (val) => {
  const map = { easy: '简单', medium: '中等', hard: '困难' }
  return map[val] || val
}

const updateUrl = () => {
  router.replace({
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

const handleReset = () => {
  filters.keyword = ''
  filters.difficulty = ''
  filters.tag = ''
  pagination.page = 1
  pagination.size = defaultPageSize
  fetchProblems()
}

function getAcceptRate(row) {
  if (!row.submit_count) return '-'
  const rate = ((row.accepted_count / row.submit_count) * 100).toFixed(1)
  return `${rate}%`
}

function goProblemDetail(row) {
  router.push(`/problem/${row.id}`)
}

onMounted(() => {
  fetchProblems()
})
</script>

<style lang="scss" scoped>
.problem-list-wrapper {
  padding: 40px 0;
  min-height: 100vh;
  background-color: var(--swiss-bg-base);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--swiss-border-light);
}

.page-title {
  font-size: 32px;
  color: var(--swiss-text-main);
  letter-spacing: -0.02em;
  margin: 0;
  line-height: 1.2;
}

.page-subtitle {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--swiss-text-secondary);
}

.header-stats {
  display: flex;
  gap: 10px;
}

.stat-chip {
  min-width: 90px;
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
    color: var(--swiss-text-main);
    line-height: 1;
  }
}

.filter-panel {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 14px;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  background: #fff;
}

.search-input {
  width: 280px;
}

.filter-select {
  width: 140px;
}

.tag-input {
  width: 160px;
}

.filter-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.table-container {
  background: #fff;
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  overflow: hidden;
  min-height: 220px;
}

.id-text {
  color: var(--swiss-text-secondary);
  font-family: var(--font-mono);
  font-size: 13px;
}

.status-text {
  font-size: 12px;
  font-weight: 700;
  &.success { color: var(--swiss-success); }
}

.title-cell {
  display: flex;
  align-items: center;
  justify-content: center; /* 标题内容居中 */
  gap: 12px;
}

.problem-link {
  font-size: 15px;
  font-weight: 500;
  color: var(--swiss-text-main);
  text-decoration: none;
  transition: color 0.2s;
  
  &:hover {
    color: var(--swiss-primary);
  }
}

.badges {
  display: flex;
  gap: 6px;
}

.badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 2px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  
  &.hidden {
    background: #FEE2E2;
    color: #DC2626;
  }
  
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
  justify-content: center; /* 标签内容居中 */
  gap: 8px;
}

.text-tag {
  font-size: 13px;
  color: var(--swiss-text-secondary);
  cursor: default;
  transition: color 0.2s;
  white-space: nowrap;
  
  &:hover {
    color: var(--swiss-text-main);
  }
}

.rate-text {
  color: var(--swiss-text-secondary);
  font-size: 13px;
}

.difficulty-wrapper {
  display: flex;
  align-items: center;
  justify-content: center; /* 难度内容居中 */
  gap: 8px;
}

.difficulty-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  
  &.easy { background-color: var(--swiss-success); }
  &.medium { background-color: var(--swiss-warning); }
  &.hard { background-color: var(--swiss-danger); }
}

.difficulty-text {
  font-size: 13px;
  color: var(--swiss-text-main);
}

.pagination-wrapper {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}

@media (max-width: 1024px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 14px;
  }

  .filter-panel {
    flex-wrap: wrap;
  }

  .search-input {
    width: 100%;
    min-width: 220px;
  }

  .filter-actions {
    margin-left: 0;
  }
}
</style>
