<template>
  <div class="page">
    <SearchForm @search="fetchList" @reset="resetQ">
      <div class="filters">
        <el-input v-model="q" placeholder="搜索：角色名/描述" clearable />
      </div>
      <template #extra-actions>
        <el-button type="success" plain @click="openCreate">新增角色</el-button>
      </template>
    </SearchForm>

    <div style="height: 12px" />

    <div class="glass wrap">
      <div class="toolbar">
        <div class="title">角色管理</div>
        <div class="meta">仅 admin 可见</div>
      </div>
      <el-table :data="items" border height="calc(100vh - 230px)">
        <el-table-column prop="name" label="角色名" width="160" />
        <el-table-column prop="description" label="描述" min-width="240" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text @click="openMenus(row)">菜单权限</el-button>
            <el-button size="small" text type="danger" @click="removeOne(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑角色' : '新增角色'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="角色名" prop="name">
          <el-input v-model="form.name" :disabled="dialog.isEdit" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="dialog.loading" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="menuDialog.visible" title="角色菜单权限" width="800px">
      <div class="menu-box glass">
        <el-tree
          ref="treeRef"
          :data="menuTree"
          node-key="id"
          show-checkbox
          default-expand-all
          :props="{ label: 'name', children: 'children' }"
          check-strictly
        />
      </div>
      <template #footer>
        <el-button @click="menuDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="menuDialog.loading" @click="saveMenus">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SearchForm from '@/components/SearchForm.vue'
import { request } from '@/utils/request'
import { useUserStore } from '@/stores/user'
import type { ApiResp } from '@/types/api'
import type { MenuNode } from '@/types/menu'

type RoleRow = { id: number; name: string; description: string }

const q = ref('')
const items = ref<RoleRow[]>([])

const dialog = reactive({ visible: false, isEdit: false, loading: false, id: 0 })
const formRef = ref<FormInstance>()
const form = reactive({ name: '', description: '' })
const rules: FormRules = {
  name: [{ required: true, message: '请输入角色名', trigger: 'blur' }],
}

const menuTree = ref<MenuNode[]>([])
const treeRef = ref<any>()
const menuDialog = reactive({ visible: false, loading: false, roleId: 0 })
const userStore = useUserStore()

function resetQ() {
  q.value = ''
  fetchList()
}

function resetForm() {
  form.name = ''
  form.description = ''
}

async function fetchList() {
  const resp = await request.get<ApiResp<{ items: RoleRow[] }>>('/roles', { params: { q: q.value } })
  items.value = resp.data.data.items || []
}

function openCreate() {
  dialog.visible = true
  dialog.isEdit = false
  dialog.id = 0
  resetForm()
}

function openEdit(row: RoleRow) {
  dialog.visible = true
  dialog.isEdit = true
  dialog.id = row.id
  form.name = row.name
  form.description = row.description
}

async function submit() {
  await formRef.value?.validate()
  dialog.loading = true
  try {
    if (dialog.isEdit) {
      await request.put<ApiResp>('/roles/' + dialog.id, { description: form.description })
    } else {
      await request.post<ApiResp>('/roles', { name: form.name.trim(), description: form.description })
    }
    ElMessage.success('保存成功')
    dialog.visible = false
    fetchList()
  } finally {
    dialog.loading = false
  }
}

async function removeOne(row: RoleRow) {
  await ElMessageBox.confirm(`确认删除角色「${row.name}」？`, '提示', { type: 'warning' })
  await request.delete<ApiResp>('/roles/' + row.id)
  ElMessage.success('删除成功')
  fetchList()
}

async function fetchMenus() {
  const resp = await request.get<ApiResp<{ items: MenuNode[] }>>('/menus')
  menuTree.value = resp.data.data.items?.length ? resp.data.data.items : []
}

async function openMenus(row: RoleRow) {
  menuDialog.visible = true
  menuDialog.roleId = row.id
  await fetchMenus()
  if (menuTree.value.length === 0) {
    ElMessage.warning('没有可用的菜单')
    menuDialog.visible = false
    return
  }
  const resp = await request.get<ApiResp<{ menu_ids: number[] }>>(`/roles/${row.id}/menus`)
  const ids = resp.data.data.menu_ids || []
  await nextTick()
  treeRef.value?.setCheckedKeys(ids, false)
}

async function saveMenus() {
  menuDialog.loading = true
  try {
    const ids = (treeRef.value?.getCheckedKeys?.() || []) as number[]
    await request.post<ApiResp>(`/roles/${menuDialog.roleId}/menus`, { menu_ids: ids })
    ElMessage.success('保存成功')
    menuDialog.visible = false
    await userStore.fetchMe()
  } finally {
    menuDialog.loading = false
  }
}

onMounted(fetchList)
</script>

<style scoped>
.filters {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}
.wrap {
  padding: 12px;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.title {
  font-weight: 800;
}
.meta {
  opacity: 0.7;
  font-size: 12px;
  align-self: center;
}
.menu-box {
  padding: 12px;
  max-height: 60vh;
  overflow: auto;
}
</style>

