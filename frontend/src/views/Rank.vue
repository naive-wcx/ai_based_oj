<template>
  <div class="rank-list-wrapper">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">排行榜</h1>
      </div>
      
      <div class="table-container">
        <el-table 
          :data="users" 
          v-loading="loading" 
          class="swiss-table"
        >
          <el-table-column label="排名" width="100" align="center" header-align="center">
            <template #default="{ $index }">
              <span :class="['rank-number', getRankClass((pagination.page - 1) * pagination.size + $index + 1)]">
                {{ (pagination.page - 1) * pagination.size + $index + 1 }}
              </span>
            </template>
          </el-table-column>

          <!-- 用户名：居中 -->
          <el-table-column prop="username" label="用户" min-width="200" align="center" header-align="center">
            <template #default="{ row }">
              <div class="user-cell">
                <span class="user-avatar">{{ row.username[0]?.toUpperCase() }}</span>
                <span class="username">{{ row.username }}</span>
                <span v-if="row.role === 'admin'" class="badge admin">ADMIN</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="solved_count" label="解题数" width="150" align="center" header-align="center">
            <template #default="{ row }">
              <span class="stat-value primary">{{ row.solved_count }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="submit_count" label="提交数" width="150" align="center" header-align="center">
            <template #default="{ row }">
              <span class="stat-value secondary">{{ row.submit_count }}</span>
            </template>
          </el-table-column>

          <el-table-column label="通过率" width="150" align="center" header-align="center">
            <template #default="{ row }">
              <span class="stat-value">{{ getAcceptRate(row) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[50, 100, 200]"
          layout="prev, pager, next"
          @size-change="fetchRank"
          @current-change="fetchRank"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { rankApi } from '@/api/rank'

const loading = ref(true)
const users = ref([])

const pagination = reactive({
  page: 1,
  size: 50,
  total: 0,
})

function getRankClass(rank) {
  if (rank === 1) return 'rank-1'
  if (rank === 2) return 'rank-2'
  if (rank === 3) return 'rank-3'
  return ''
}

function getAcceptRate(row) {
  if (!row.submit_count) return '-'
  const accepted = row.accepted_count !== undefined ? row.accepted_count : row.solved_count
  const rate = (accepted / row.submit_count * 100).toFixed(1)
  return `${rate}%`
}

async function fetchRank() {
  loading.value = true
  try {
    const res = await rankApi.getList({
      page: pagination.page,
      size: pagination.size,
    })
    users.value = res.data.list || []
    pagination.total = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchRank()
})
</script>

<style lang="scss" scoped>
.rank-list-wrapper {
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

.rank-number {
  font-weight: 500;
  font-size: 14px;
  color: var(--swiss-text-secondary);
  font-family: var(--font-mono);
  
  &.rank-1 { color: #D97706; font-weight: 700; } /* Gold */
  &.rank-2 { color: #4B5563; font-weight: 700; } /* Silver */
  &.rank-3 { color: #B45309; font-weight: 700; } /* Bronze */
}

.user-cell {
  display: flex;
  align-items: center;
  justify-content: center; /* 居中对齐 */
  gap: 12px;
}

.user-avatar {
  width: 28px;
  height: 28px;
  background: #F3F4F6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--swiss-text-secondary);
}

.username {
  font-weight: 500;
  color: var(--swiss-text-main);
  font-size: 14px;
}

.badge {
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 2px;
  font-weight: 700;
  letter-spacing: 0.05em;
  
  &.admin {
    background: #FEE2E2;
    color: #DC2626;
  }
}

.stat-value {
  font-size: 14px;
  font-family: var(--font-mono);
  
  &.primary { color: var(--swiss-text-main); font-weight: 600; }
  &.secondary { color: var(--swiss-text-secondary); }
}

.pagination-wrapper {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}
</style>
