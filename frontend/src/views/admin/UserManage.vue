<template>
  <div class="user-manage">
    <div class="page-header">
      <div>
        <h2 class="page-title">用户管理</h2>
        <p class="page-subtitle">支持账户维护、管理员授权与批量导入。</p>
      </div>
      <div class="page-actions">
        <el-button type="primary" @click="openCreateDialog">创建用户</el-button>
        <el-button @click="openBatchDialog">批量导入</el-button>
      </div>
    </div>

    <div class="header-stats">
      <div class="stat-chip">
        <span>总用户</span>
        <strong>{{ pagination.total }}</strong>
      </div>
      <div class="stat-chip">
        <span>本页管理员</span>
        <strong>{{ pageAdminCount }}</strong>
      </div>
      <div class="stat-chip">
        <span>本页解题</span>
        <strong>{{ pageSolvedCount }}</strong>
      </div>
    </div>
    
    <div class="card table-card" v-loading="loading">
      <el-table v-if="users.length" :data="users" stripe class="swiss-table">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="student_id" label="学号" width="120" />
        <el-table-column label="分组" width="140">
          <template #default="{ row }">
            {{ row.group || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)">
              {{ getRoleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="solved_count" label="解题数" width="100" />
        <el-table-column prop="submit_count" label="提交数" width="100" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button
              v-if="userStore.isSuperAdmin && row.role !== 'super_admin' && row.id !== userStore.user?.id"
              size="small"
              :type="row.role === 'admin' ? 'warning' : 'danger'"
              :loading="roleUpdatingId === row.id"
              @click="handleToggleAdmin(row)"
            >
              {{ row.role === 'admin' ? '取消管理员' : '设为管理员' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="暂无用户数据" />
      
      <div class="pagination" v-if="pagination.total > 0">
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

    <el-dialog v-model="createDialogVisible" title="创建用户" width="520px" @closed="resetCreateForm">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-position="top">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="createForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="可选" />
        </el-form-item>
        <el-form-item label="学号" prop="student_id">
          <el-input v-model="createForm.student_id" placeholder="可选" />
        </el-form-item>
        <el-form-item label="分组" prop="group">
          <el-input v-model="createForm.group" placeholder="可选，例如：ClassA" />
        </el-form-item>
        <el-form-item label="初始密码" prop="password">
          <el-input v-model="createForm.password" type="password" show-password placeholder="请输入初始密码" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input v-model="createForm.confirm_password" type="password" show-password placeholder="请再次输入密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editDialogVisible" title="编辑用户" width="520px" @closed="resetEditForm">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-position="top">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="editForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" placeholder="可留空" />
        </el-form-item>
        <el-form-item label="学号" prop="student_id">
          <el-input v-model="editForm.student_id" placeholder="可留空" />
        </el-form-item>
        <el-form-item label="分组" prop="group">
          <el-input v-model="editForm.group" placeholder="可留空，例如：ClassA" />
        </el-form-item>
        <el-form-item label="重置密码" prop="password">
          <el-input v-model="editForm.password" type="password" show-password placeholder="留空则不修改" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input v-model="editForm.confirm_password" type="password" show-password placeholder="留空则不修改" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="updating" @click="handleUpdate">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDialogVisible" title="批量导入用户" width="640px" @closed="resetBatchForm">
      <div class="batch-hint">
        支持 CSV 或 JSON。CSV 默认列顺序：username,password,student_id,email,group（后三列可省略，且不支持逗号转义）。
      </div>
      <el-form label-position="top">
        <el-form-item label="默认分组（可选）">
          <el-input v-model="batchForm.defaultGroup" placeholder="为空则不设置" />
        </el-form-item>
        <el-form-item label="数据内容">
          <el-input
            v-model="batchForm.text"
            type="textarea"
            :rows="10"
            placeholder="username,password,student_id,email,group&#10;alice,pass123,2025001,,ClassA"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="batching" @click="handleBatchImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from '@/utils/message'
import { adminApi } from '@/api/admin'
import { useUserStore } from '@/stores/user'

const loading = ref(false)
const users = ref([])
const roleUpdatingId = ref(null)
const userStore = useUserStore()

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

const pageAdminCount = computed(
  () => users.value.filter((user) => user.role === 'admin' || user.role === 'super_admin').length
)

const pageSolvedCount = computed(
  () => users.value.reduce((sum, user) => sum + (user.solved_count || 0), 0)
)

function getRoleLabel(role) {
  if (role === 'super_admin') return '超级管理员'
  if (role === 'admin') return '管理员'
  return '普通用户'
}

function getRoleTagType(role) {
  if (role === 'super_admin') return 'warning'
  if (role === 'admin') return 'danger'
  return 'info'
}

const createDialogVisible = ref(false)
const createFormRef = ref()
const creating = ref(false)
const createForm = reactive({
  username: '',
  email: '',
  student_id: '',
  group: '',
  password: '',
  confirm_password: '',
})

const editDialogVisible = ref(false)
const editFormRef = ref()
const updating = ref(false)
const editForm = reactive({
  id: null,
  username: '',
  email: '',
  student_id: '',
  group: '',
  password: '',
  confirm_password: '',
})
const editOriginal = reactive({
  username: '',
  email: '',
  student_id: '',
  group: '',
})

const batchDialogVisible = ref(false)
const batching = ref(false)
const batchForm = reactive({
  text: '',
  defaultGroup: '',
})

const confirmPasswordValidator = (rule, value, callback) => {
  if (!value) {
    callback(new Error('请再次输入密码'))
    return
  }
  if (value !== createForm.password) {
    callback(new Error('两次输入的密码不一致'))
    return
  }
  callback()
}

const optionalEmailValidator = (rule, value, callback) => {
  if (!value) {
    callback()
    return
  }
  const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailPattern.test(value)) {
    callback(new Error('邮箱格式不正确'))
    return
  }
  callback()
}

const editConfirmPasswordValidator = (rule, value, callback) => {
  if (!editForm.password) {
    callback()
    return
  }
  if (!value) {
    callback(new Error('请再次输入密码'))
    return
  }
  if (value !== editForm.password) {
    callback(new Error('两次输入的密码不一致'))
    return
  }
  callback()
}

const createRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ validator: optionalEmailValidator, trigger: 'blur' }],
  password: [{ required: true, message: '请输入初始密码', trigger: 'blur' }],
  confirm_password: [{ validator: confirmPasswordValidator, trigger: 'blur' }],
}

const editRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ validator: optionalEmailValidator, trigger: 'blur' }],
  confirm_password: [{ validator: editConfirmPasswordValidator, trigger: 'blur' }],
}

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

function openCreateDialog() {
  createDialogVisible.value = true
}

async function handleToggleAdmin(row) {
  if (!row || row.role === 'super_admin') return
  roleUpdatingId.value = row.id
  try {
    const nextRole = row.role === 'admin' ? 'user' : 'admin'
    await adminApi.setUserRole(row.id, nextRole)
    message.success(nextRole === 'admin' ? '已设为管理员' : '已取消管理员')
    await fetchUsers()
  } catch (e) {
    console.error(e)
  } finally {
    roleUpdatingId.value = null
  }
}

function resetCreateForm() {
  createForm.username = ''
  createForm.email = ''
  createForm.student_id = ''
  createForm.group = ''
  createForm.password = ''
  createForm.confirm_password = ''
  createFormRef.value?.clearValidate()
}

async function handleCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return

  creating.value = true
  try {
    await adminApi.createUser({
      username: createForm.username,
      email: createForm.email,
      student_id: createForm.student_id,
      group: createForm.group,
      password: createForm.password,
    })
    message.success('创建成功')
    createDialogVisible.value = false
    fetchUsers()
  } catch (e) {
    console.error(e)
  } finally {
    creating.value = false
  }
}

function openEditDialog(row) {
  editForm.id = row.id
  editForm.username = row.username || ''
  editForm.email = row.email || ''
  editForm.student_id = row.student_id || ''
  editForm.group = row.group || ''
  editForm.password = ''
  editForm.confirm_password = ''

  editOriginal.username = editForm.username
  editOriginal.email = editForm.email
  editOriginal.student_id = editForm.student_id
  editOriginal.group = editForm.group

  editDialogVisible.value = true
}

function resetEditForm() {
  editForm.id = null
  editForm.username = ''
  editForm.email = ''
  editForm.student_id = ''
  editForm.group = ''
  editForm.password = ''
  editForm.confirm_password = ''
  editFormRef.value?.clearValidate()
}

async function handleUpdate() {
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) return

  const payload = {}
  if (editForm.username !== editOriginal.username) payload.username = editForm.username
  if (editForm.email !== editOriginal.email) payload.email = editForm.email
  if (editForm.student_id !== editOriginal.student_id) payload.student_id = editForm.student_id
  if (editForm.group !== editOriginal.group) payload.group = editForm.group
  if (editForm.password) payload.password = editForm.password

  if (Object.keys(payload).length === 0) {
    message.info('没有需要更新的内容')
    return
  }

  updating.value = true
  try {
    await adminApi.updateUser(editForm.id, payload)
    message.success('更新成功')
    editDialogVisible.value = false
    fetchUsers()
  } catch (e) {
    console.error(e)
  } finally {
    updating.value = false
  }
}

function openBatchDialog() {
  batchDialogVisible.value = true
}

function resetBatchForm() {
  batchForm.text = ''
  batchForm.defaultGroup = ''
}

function parseBatchInput(text) {
  const trimmed = text.trim()
  if (!trimmed) return { users: [], errors: ['数据为空'] }

  if (trimmed.startsWith('[') || trimmed.startsWith('{')) {
    try {
      const parsed = JSON.parse(trimmed)
      const users = Array.isArray(parsed) ? parsed : parsed.users
      if (!Array.isArray(users)) {
        return { users: [], errors: ['JSON 必须是数组或包含 users 数组'] }
      }
      return { users, errors: [] }
    } catch (e) {
      return { users: [], errors: ['JSON 解析失败'] }
    }
  }

  const lines = trimmed.split(/\r?\n/).map((line) => line.trim()).filter(Boolean)
  if (!lines.length) return { users: [], errors: ['数据为空'] }

  let startIndex = 0
  let headerMap = null
  const header = lines[0].toLowerCase()
  if (header.includes('username') && header.includes('password')) {
    headerMap = header.split(',').map((item) => item.trim())
    startIndex = 1
  }

  const users = []
  const errors = []
  for (let i = startIndex; i < lines.length; i += 1) {
    const cols = lines[i].split(',').map((item) => item.trim())
    const getValue = (key, index) => {
      if (!headerMap) return cols[index] || ''
      const idx = headerMap.indexOf(key)
      if (idx === -1) return ''
      return cols[idx] || ''
    }
    const user = {
      username: getValue('username', 0),
      password: getValue('password', 1),
      student_id: getValue('student_id', 2),
      email: getValue('email', 3),
      group: getValue('group', 4) || batchForm.defaultGroup,
    }
    if (!user.username || !user.password) {
      errors.push(`第 ${i + 1} 行缺少用户名或密码`)
      continue
    }
    users.push(user)
  }

  return { users, errors }
}

async function handleBatchImport() {
  const { users: parsedUsers, errors } = parseBatchInput(batchForm.text)
  if (errors.length) {
    message.error(errors[0])
    return
  }
  if (!parsedUsers.length) {
    message.warning('没有可导入的用户')
    return
  }

  batching.value = true
  try {
    const res = await adminApi.createUsersBatch({ users: parsedUsers })
    const { created, failed } = res.data
    if (failed > 0) {
      message.warning(`导入完成：成功 ${created}，失败 ${failed}`)
      console.warn(res.data.errors)
    } else {
      message.success(`导入成功：共 ${created} 个用户`)
    }
    batchDialogVisible.value = false
    fetchUsers()
  } catch (e) {
    console.error(e)
  } finally {
    batching.value = false
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style lang="scss" scoped>
.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 14px;
}

.page-title {
  margin: 0;
}

.page-subtitle {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--swiss-text-secondary);
}

.page-actions {
  display: flex;
  gap: 10px;
}

.header-stats {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.stat-chip {
  min-width: 100px;
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
    line-height: 1;
    color: var(--swiss-text-main);
  }
}

.batch-hint {
  margin-bottom: 12px;
  color: #909399;
  font-size: 13px;
}

.table-card {
  border: 1px solid var(--swiss-border-light);
  border-radius: var(--radius-sm);
  background: #fff;
  padding: 12px;
  min-height: 260px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

@media (max-width: 1024px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .page-actions {
    flex-wrap: wrap;
  }

  .header-stats {
    flex-wrap: wrap;
  }
}
</style>
