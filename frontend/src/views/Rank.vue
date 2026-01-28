<template>
  <div class="rank-page">
    <h1 class="page-title">排行榜</h1>
    
    <div class="card">
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column label="排名" width="80">
          <template #default="{ $index }">
            <span :class="['rank', `rank-${$index + 1}`]">
              {{ (pagination.page - 1) * pagination.size + $index + 1 }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户" min-width="150">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="32">{{ row.username[0]?.toUpperCase() }}</el-avatar>
              <span>{{ row.username }}</span>
              <el-tag v-if="row.role === 'admin'" size="small" type="danger">管理员</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="solved_count" label="解题数" width="120" sortable />
        <el-table-column prop="submit_count" label="提交数" width="120" sortable />
        <el-table-column label="通过率" width="120">
          <template #default="{ row }">
            {{ getAcceptRate(row) }}
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

const loading = ref(false)
const users = ref([])

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

function getAcceptRate(row) {
  if (!row.submit_count) return '-'
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
.rank {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-weight: 600;
  
  &.rank-1 {
    background: linear-gradient(135deg, #ffd700, #ffb347);
    color: white;
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
  gap: 8px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
