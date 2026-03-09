<template>
  <div class="page page--table">
    <SearchForm @search="onSearch" @reset="onReset">
      <div class="filters">
        <el-input v-model="q" placeholder="搜索：用户名/昵称/邮箱" clearable />
      </div>
      <template #extra-actions>
        <el-button type="success" plain @click="openCreate">新增用户</el-button>
      </template>
    </SearchForm>

    <div class="table-block">
      <ElTablePro :total="total" v-model:page="page" v-model:pageSize="pageSize" @change="fetchList">
        <template #toolbar>
          <div class="title">用户管理</div>
          <div class="meta">仅 admin 可见</div>
        </template>

        <div class="table-inner">
          <el-table :data="items" border height="100%">
            <el-table-column prop="username" label="用户名" min-width="140" />
            <el-table-column prop="nickname" label="昵称" min-width="140" />
            <el-table-column prop="email" label="邮箱" min-width="180" />
            <el-table-column prop="status" label="状态" width="110">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'" effect="dark">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="角色" min-width="220">
              <template #default="{ row }">
                <el-tag v-for="r in row.roles || []" :key="r" style="margin-right: 6px" effect="dark">
                  {{ r }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <el-button size="small" text type="primary" @click="openEdit(row)">编辑</el-button>
                <el-button size="small" text @click="openBindRoles(row)">绑定角色</el-button>
                <el-button size="small" text type="danger" @click="removeOne(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </ElTablePro>
    </div>

    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑用户' : '新增用户'" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="dialog.isEdit" />
        </el-form-item>
        <el-form-item :label="dialog.isEdit ? '新密码' : '密码'" :prop="dialog.isEdit ? '' : 'password'">
          <el-input v-model="form.password" show-password type="password" placeholder="留空表示不修改" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.enabled" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="dialog.loading" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleDialog.visible" title="绑定角色" width="520px">
      <el-form label-width="90px">
        <el-form-item label="用户">
          <el-input :model-value="roleDialog.username" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="roleDialog.roleIds" multiple filterable style="width: 100%">
            <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="roleDialog.loading" @click="saveUserRoles">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SearchForm from '@/components/SearchForm.vue'
import ElTablePro from '@/components/ElTablePro.vue'
import { request } from '@/utils/request'
import type { ApiResp } from '@/types/api'

type UserRow = {
  id: number
  username: string
  nickname: string
  email: string
  status: number
  roles: string[]
}

type RoleRow = { id: number; name: string; description: string }

const q = ref('')
const items = ref<UserRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const roles = ref<RoleRow[]>([])

const dialog = reactive({ visible: false, isEdit: false, loading: false, id: 0 })
const formRef = ref<FormInstance>()
const form = reactive({ username: '', password: '', nickname: '', email: '', enabled: true })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const roleDialog = reactive({
  visible: false,
  loading: false,
  userId: 0,
  username: '',
  roleIds: [] as number[],
})

function resetForm() {
  form.username = ''
  form.password = ''
  form.nickname = ''
  form.email = ''
  form.enabled = true
}

async function fetchRoles() {
  const resp = await request.get<ApiResp<{ items: RoleRow[] }>>('/roles')
  roles.value = resp.data.data.items || []
}

async function fetchList() {
  const resp = await request.get<ApiResp<{ items: UserRow[]; total: number }>>('/users', {
    params: { q: q.value, page: page.value, page_size: pageSize.value },
  })
  items.value = resp.data.data.items || []
  total.value = resp.data.data.total || 0
}

function onSearch() {
  page.value = 1
  fetchList()
}

function onReset() {
  q.value = ''
  page.value = 1
  fetchList()
}

function openCreate() {
  dialog.visible = true
  dialog.isEdit = false
  dialog.id = 0
  resetForm()
}

function openEdit(row: UserRow) {
  dialog.visible = true
  dialog.isEdit = true
  dialog.id = row.id
  form.username = row.username
  form.password = ''
  form.nickname = row.nickname
  form.email = row.email
  form.enabled = row.status === 1
}

async function submit() {
  if (!dialog.isEdit) await formRef.value?.validate()
  dialog.loading = true
  try {
    if (dialog.isEdit) {
      await request.put<ApiResp>('/users/' + dialog.id, {
        password: form.password || undefined,
        nickname: form.nickname,
        email: form.email,
        status: form.enabled ? 1 : 0,
      })
    } else {
      await request.post<ApiResp>('/users', {
        username: form.username.trim(),
        password: form.password,
        nickname: form.nickname,
        email: form.email,
        status: form.enabled ? 1 : 0,
      })
    }
    ElMessage.success('保存成功')
    dialog.visible = false
    fetchList()
  } finally {
    dialog.loading = false
  }
}

async function removeOne(row: UserRow) {
  await ElMessageBox.confirm(`确认删除用户「${row.username}」？`, '提示', { type: 'warning' })
  await request.delete<ApiResp>('/users/' + row.id)
  ElMessage.success('删除成功')
  fetchList()
}

function openBindRoles(row: UserRow) {
  roleDialog.visible = true
  roleDialog.userId = row.id
  roleDialog.username = row.username
  const name2id = new Map(roles.value.map((r) => [r.name, r.id]))
  roleDialog.roleIds = (row.roles || []).map((n) => name2id.get(n) || 0).filter((x) => x > 0)
}

async function saveUserRoles() {
  roleDialog.loading = true
  try {
    await request.put<ApiResp>(`/users/${roleDialog.userId}/roles`, {
      role_ids: roleDialog.roleIds,
    })
    ElMessage.success('绑定成功')
    roleDialog.visible = false
    fetchList()
  } finally {
    roleDialog.loading = false
  }
}

onMounted(async () => {
  await fetchRoles()
  await fetchList()
})
</script>

<style scoped>
.page--table {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
}
.page--table .table-block {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  margin-top: 12px;
}
.table-inner {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
.filters {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}
.title {
  font-weight: 800;
}
.meta {
  opacity: 0.7;
  font-size: 12px;
  align-self: center;
}
</style>

