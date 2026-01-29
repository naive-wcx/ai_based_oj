<template>
  <el-card shadow="never" class="page-container">
    <template #header>
      <div class="page-header">
        <span class="page-title">🏆 排行榜</span>
      </div>
    </template>
    
    <el-table :data="users" v-loading="loading" stripe style="width: 100%">
      <el-table-column label="排名" width="100" align="center">
        <template #default="{ $index }">
          <span :class="['rank', getRankClass((pagination.page - 1) * pagination.size + $index + 1)]">
            {{ (pagination.page - 1) * pagination.size + $index + 1 }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="username" label="用户" min-width="200">
        <template #default="{ row }">
          <div class="user-cell">
            <el-avatar :size="32" :src="row.avatar">{{ row.username[0]?.toUpperCase() }}</el-avatar>
            <span class="username">{{ row.username }}</span>
            <el-tag v-if="row.role === 'admin'" size="small" type="danger" effect="dark">管理员</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="solved_count" label="解题数" width="150" sortable align="center" />
      <el-table-column prop="submit_count" label="提交数" width="150" sortable align="center" />
      <el-table-column label="通过率" width="150" align="center">
        <template #default="{ row }">
          <span class="accept-rate">{{ getAcceptRate(row) }}</span>
        </template>
      </el-table-column>
    </el-table>
    
    <div class="pagination-container" v-if="pagination.total > 0">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[50, 100, 200]"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="fetchRank"
        @current-change="fetchRank"
      />
    </div>
  </el-card>
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
  if (!row.submit_count) return 'N/A'
  const rate = (row.solved_count / row.submit_count * 100).toFixed(1)
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
.page-container {
  border: none;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.rank {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  font-weight: 600;
  font-size: 14px;
  
  &.rank-1 {
    background: linear-gradient(135deg, #ffd700, #ffb347);
    color: white;
    box-shadow: 0 0 8px rgba(255, 215, 0, 0.6);
  }
  
  &.rank-2 {
    background: linear-gradient(135deg, #c0c0c0, #a8a8a8);
    color: white;
  }
  
  &.rank-3 {
    background: linear-gradient(135deg, #cd7f32, #b87333);
    color: white;
  }
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  .username {
    font-weight: 500;
  }
}

.accept-rate {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
