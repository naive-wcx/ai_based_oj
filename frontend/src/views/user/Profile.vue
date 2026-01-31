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
      
      <div class="right-column">
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

        <div class="card">
          <h3>修改密码</h3>
          <el-form :model="passwordForm" label-width="100px" style="max-width: 400px">
            <el-form-item label="当前密码">
              <el-input v-model="passwordForm.old_password" type="password" show-password autocomplete="current-password" />
            </el-form-item>
            <el-form-item label="新密码">
              <el-input v-model="passwordForm.new_password" type="password" show-password autocomplete="new-password" />
            </el-form-item>
            <el-form-item label="确认新密码">
              <el-input v-model="passwordForm.confirm_password" type="password" show-password autocomplete="new-password" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">
                修改密码
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from '@/utils/message'
import { useUserStore } from '@/stores/user'
import { userApi } from '@/api/user'

const userStore = useUserStore()

const loading = ref(false)
const form = reactive({
  email: '',
  student_id: '',
})
const passwordLoading = ref(false)
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

async function handleUpdate() {
  loading.value = true
  try {
    await userApi.updateProfile(form)
    message.success('更新成功')
    await userStore.fetchProfile()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleChangePassword() {
  if (!passwordForm.old_password || !passwordForm.new_password) {
    message.error('请输入完整的密码信息')
    return
  }
  if (passwordForm.new_password.length < 6 || passwordForm.new_password.length > 20) {
    message.error('新密码长度应为 6-20')
    return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    message.error('两次输入的新密码不一致')
    return
  }
  if (passwordForm.old_password === passwordForm.new_password) {
    message.error('新密码不能与原密码相同')
    return
  }

  passwordLoading.value = true
  try {
    await userApi.changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    message.success('密码修改成功')
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
  } catch (e) {
    console.error(e)
  } finally {
    passwordLoading.value = false
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

.right-column {
  display: grid;
  gap: 20px;
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
