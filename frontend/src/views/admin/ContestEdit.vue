<template>
  <div class="contest-edit">
    <div class="page-header">
      <h2>{{ isEdit ? '编辑比赛' : '创建比赛' }}</h2>
    </div>

    <div class="card">
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="比赛名称" prop="title">
          <el-input v-model="form.title" placeholder="请输入比赛名称" />
        </el-form-item>

        <el-form-item label="比赛描述" prop="description">
          <div class="md-row">
            <el-input
              v-model="form.description"
              type="textarea"
              placeholder="支持 Markdown (赛制说明、注意事项等)"
              class="md-input"
              resize="none"
            />
            <div class="md-preview">
              <MarkdownPreview :content="form.description" />
            </div>
          </div>
        </el-form-item>

        <div class="form-row">
          <el-form-item label="赛制" prop="type" class="form-item">
            <el-select v-model="form.type" placeholder="请选择赛制">
              <el-option label="OI" value="oi" />
              <el-option label="IOI" value="ioi" />
            </el-select>
          </el-form-item>
          <el-form-item label="开始时间" prop="start_at" class="form-item">
            <el-date-picker v-model="form.start_at" type="datetime" placeholder="开始时间" />
          </el-form-item>
          <el-form-item label="结束时间" prop="end_at" class="form-item">
            <el-date-picker v-model="form.end_at" type="datetime" placeholder="结束时间" />
          </el-form-item>
        </div>

        <el-form-item label="比赛题目" prop="problem_ids">
          <el-select
            v-model="form.problem_ids"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="请选择题目"
          >
            <el-option
              v-for="problem in problemOptions"
              :key="problem.id"
              :label="`${problem.id} - ${problem.title}`"
              :value="problem.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="允许参赛用户" prop="allowed_users">
          <el-select
            v-model="form.allowed_users"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="请选择参赛用户"
          >
            <el-option
              v-for="user in userOptions"
              :key="user.id"
              :label="`${user.username} (#${user.id})`"
              :value="user.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="允许参赛分组" prop="allowed_groups">
          <el-select
            v-model="form.allowed_groups"
            multiple
            filterable
            allow-create
            default-first-option
            collapse-tags
            collapse-tags-tooltip
            placeholder="请选择参赛分组"
          >
            <el-option
              v-for="group in groupOptions"
              :key="group"
              :label="group"
              :value="group"
            />
          </el-select>
        </el-form-item>

        <div class="form-actions">
          <el-button @click="$router.back()">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSubmit">
            {{ isEdit ? '保存' : '创建' }}
          </el-button>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from '@/utils/message'
import { adminApi } from '@/api/admin'
import { contestApi } from '@/api/contest'
import { problemApi } from '@/api/problem'
import MarkdownPreview from '@/components/common/MarkdownPreview.vue'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id)

const formRef = ref()
const saving = ref(false)

const form = reactive({
  title: '',
  description: '',
  type: 'oi',
  start_at: null,
  end_at: null,
  problem_ids: [],
  allowed_users: [],
  allowed_groups: [],
})

const rules = {
  title: [{ required: true, message: '请输入比赛名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择赛制', trigger: 'change' }],
  start_at: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_at: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
}

const problemOptions = ref([])
const userOptions = ref([])
const groupOptions = ref([])

async function fetchProblems() {
  try {
    const res = await problemApi.getList({ page: 1, size: 1000 })
    problemOptions.value = res.data.list || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchUsers() {
  try {
    const res = await adminApi.getUserList({ page: 1, size: 1000 })
    userOptions.value = res.data.list || []
    const groups = userOptions.value
      .map((user) => user.group)
      .filter((group) => group)
    groupOptions.value = Array.from(new Set(groups))
  } catch (e) {
    console.error(e)
  }
}

async function fetchContest() {
  if (!isEdit.value) return
  try {
    const res = await contestApi.getById(route.params.id)
    const contest = res.data.contest
    form.title = contest.title
    form.description = contest.description || ''
    form.type = contest.type
    form.start_at = contest.start_at ? new Date(contest.start_at) : null
    form.end_at = contest.end_at ? new Date(contest.end_at) : null
    form.problem_ids = contest.problem_ids || []
    form.allowed_users = contest.allowed_users || []
    form.allowed_groups = contest.allowed_groups || []
  } catch (e) {
    console.error(e)
  }
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  if (form.allowed_users.length === 0 && form.allowed_groups.length === 0) {
    message.warning('请至少选择一个参赛用户或分组')
    return
  }

  saving.value = true
  try {
    const payload = {
      title: form.title,
      description: form.description,
      type: form.type,
      start_at: form.start_at,
      end_at: form.end_at,
      problem_ids: form.problem_ids,
      allowed_users: form.allowed_users,
      allowed_groups: form.allowed_groups,
    }

    if (isEdit.value) {
      await adminApi.updateContest(route.params.id, payload)
      message.success({ message: '更新成功', duration: 1000 })
    } else {
      await adminApi.createContest(payload)
      message.success({ message: '创建成功', duration: 1000 })
    }
    router.push('/admin/contests')
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchProblems()
  fetchUsers()
  fetchContest()
})
</script>

<style lang="scss" scoped>
.page-header {
  margin-bottom: 20px;

  h2 {
    margin: 0;
  }
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}

.md-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  width: 100%;
  height: 320px;

  @media (max-width: 768px) {
    grid-template-columns: 1fr;
    height: auto;

    .md-preview {
      height: 260px;
      margin-top: 12px;
    }

    .md-input :deep(.el-textarea__inner) {
      height: 260px;
    }
  }
}

.md-input {
  height: 100%;

  :deep(.el-textarea) {
    height: 100%;
  }

  :deep(.el-textarea__inner) {
    height: 100%;
    font-family: var(--font-mono);
    line-height: 1.6;
  }
}

.md-preview {
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #fafafa;
  overflow-y: auto;
  height: 100%;
}
</style>
