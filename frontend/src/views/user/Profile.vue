<template>
  <div class="profile-page">
    <h1 class="page-title">个人中心</h1>
    
    <div class="profile-content">
      <div class="card profile-card">
        <div class="profile-header">
          <el-avatar :size="80">{{ userStore.username[0]?.toUpperCase() }}</el-avatar>
          <div class="profile-info">
            <h2>{{ userStore.user?.username }}</h2>
            <el-tag v-if="userStore.isAdmin" type="danger">管理员</el-tag>
          </div>
        </div>
        
        <el-descriptions :column="1" border>
          <el-descriptions-item label="用户名">{{ userStore.user?.username }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ userStore.user?.email }}</el-descriptions-item>
          <el-descriptions-item label="学号">{{ userStore.user?.student_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="解题数">{{ userStore.user?.solved_count || 0 }}</el-descriptions-item>
          <el-descriptions-item label="提交数">{{ userStore.user?.submit_count || 0 }}</el-descriptions-item>
        </el-descriptions>
      </div>
      
      <div class="card">
        <h3>修改资料</h3>
        <el-form ref="formRef" :model="form" label-width="80px" style="max-width: 400px">
          <el-form-item label="邮箱">
            <el-input v-model="form.email" />
          </el-form-item>
          <el-form-item label="学号">
            <el-input v-model="form.student_id" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="loading" @click="handleUpdate">
              保存修改
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { userApi } from '@/api/user'

const userStore = useUserStore()

const loading = ref(false)
const form = reactive({
  email: '',
  student_id: '',
})

async function handleUpdate() {
  loading.value = true
  try {
    await userApi.updateProfile(form)
    ElMessage.success('更新成功')
    await userStore.fetchProfile()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await userStore.fetchProfile()
  if (userStore.user) {
    form.email = userStore.user.email || ''
    form.student_id = userStore.user.student_id || ''
  }
})
</script>

<style lang="scss" scoped>
.profile-content {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 20px;
  
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}

.profile-card {
  .profile-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 24px;
    
    h2 {
      margin: 0;
    }
  }
}

.card h3 {
  margin-bottom: 20px;
}
</style>
