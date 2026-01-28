<template>
  <div class="user-manage">
    <div class="page-header">
      <h2>用户管理</h2>
    </div>
    
    <div class="card">
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="student_id" label="学号" width="120" />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'info'">
              {{ row.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="solved_count" label="解题数" width="100" />
        <el-table-column prop="submit_count" label="提交数" width="100" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.role !== 'admin'"
              size="small"
              type="warning"
              @click="setRole(row, 'admin')"
            >
              设为管理员
            </el-button>
            <el-button
              v-else
              size="small"
              @click="setRole(row, 'user')"
            >
              取消管理员
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
          @size-change="fetchUsers"
          @current-change="fetchUsers"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const loading = ref(false)
const users = ref([])

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

async function fetchUsers() {
  loading.value = true
  try {
    const res = await adminApi.getUserList({
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

async function setRole(row, role) {
  const action = role === 'admin' ? '设为管理员' : '取消管理员'
  try {
    await ElMessageBox.confirm(`确定要将用户 "${row.username}" ${action}吗？`, '提示', {
      type: 'warning',
    })
    
    await adminApi.setUserRole(row.id, role)
    ElMessage.success('设置成功')
    fetchUsers()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style lang="scss" scoped>
.page-header {
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
