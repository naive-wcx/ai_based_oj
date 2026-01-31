<template>
  <div class="problem-manage">
    <div class="page-header">
      <h2>题目管理</h2>
      <el-button type="primary" @click="$router.push('/admin/problem/create')">
        <el-icon><Plus /></el-icon> 创建题目
      </el-button>
    </div>
    
    <div class="card">
      <el-table :data="problems" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="题目" min-width="200" />
        <el-table-column label="难度" width="100">
          <template #default="{ row }">
            <DifficultyBadge :difficulty="row.difficulty" />
          </template>
        </el-table-column>
        <el-table-column label="AI 判题" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.has_ai_judge" type="success" size="small">已启用</el-tag>
            <el-tag v-else type="info" size="small">未启用</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交/通过" width="120">
          <template #default="{ row }">
            {{ row.submit_count || 0 }} / {{ row.accepted_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="$router.push(`/admin/problem/${row.id}/edit`)">
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
          @size-change="fetchProblems"
          @current-change="fetchProblems"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessageBox } from 'element-plus'
import { message } from '@/utils/message'
import { Plus } from '@element-plus/icons-vue'
import { problemApi } from '@/api/problem'
import DifficultyBadge from '@/components/problem/DifficultyBadge.vue'

const loading = ref(false)
const problems = ref([])

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
    })
    problems.value = res.data.list || []
    pagination.total = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除题目 "${row.title}" 吗？`, '提示', {
      type: 'warning',
    })
    
    await problemApi.delete(row.id)
    message.success('删除成功')
    fetchProblems()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

onMounted(() => {
  fetchProblems()
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
