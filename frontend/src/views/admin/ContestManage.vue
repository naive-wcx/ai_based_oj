<template>
  <div class="contest-manage">
    <div class="page-header">
      <h2>比赛管理</h2>
      <el-button type="primary" @click="$router.push('/admin/contest/create')">
        创建比赛
      </el-button>
    </div>

    <div class="card">
      <el-table :data="contests" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="比赛名称" min-width="240" />
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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="$router.push(`/admin/contest/${row.id}/edit`)">
              编辑
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
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
import { ElMessageBox } from 'element-plus'
import { message } from '@/utils/message'
import { contestApi } from '@/api/contest'
import { adminApi } from '@/api/admin'

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

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除比赛 "${row.title}" 吗？`, '提示', {
      type: 'warning',
    })
    await adminApi.deleteContest(row.id)
    message.success('删除成功')
    fetchContests()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

onMounted(() => {
  fetchContests()
})
</script>

<style lang="scss" scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h2 {
    margin: 0;
  }
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
