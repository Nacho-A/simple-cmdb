<template>
  <div class="page">
    <div class="glass wrap">
      <div class="toolbar">
        <div class="title">菜单管理</div>
        <div class="actions">
          <el-button type="success" plain @click="openCreate">新增菜单</el-button>
          <el-button plain @click="fetchList">刷新</el-button>
        </div>
      </div>

      <el-table :data="tree" border row-key="id" tree-props="{ children: 'children' }" height="calc(100vh - 210px)">
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="path" label="路径" min-width="200" />
        <el-table-column prop="component" label="组件" min-width="220" />
        <el-table-column prop="icon" label="图标" width="140" />
        <el-table-column prop="order" label="排序" width="80" />
        <el-table-column prop="hidden" label="隐藏" width="80">
          <template #default="{ row }">
            <el-tag :type="row.hidden ? 'info' : 'success'" effect="dark">{{ row.hidden ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text type="danger" @click="removeOne(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑菜单' : '新增菜单'" width="720px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" placeholder="/cmdb/assets" />
        </el-form-item>
        <el-form-item label="组件">
          <el-input v-model="form.component" placeholder="views/cmdb/assets/index.vue（目录菜单可留空）" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="Element Plus icon 名称，如 Setting/Menu/Monitor" />
        </el-form-item>
        <el-form-item label="父菜单">
          <el-select v-model="form.parent_id" clearable filterable style="width: 100%">
            <el-option :value="null" label="无（一级菜单）" />
            <el-option v-for="m in treeFlat" :key="m.id" :label="`${m.name} (${m.path})`" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.order" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="隐藏">
          <el-switch v-model="form.hidden" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="dialog.loading" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { request } from '@/utils/request'
import type { ApiResp } from '@/types/api'
import type { MenuNode } from '@/types/menu'

const tree = ref<MenuNode[]>([])

const treeFlat = ref<MenuNode[]>([])

const dialog = reactive({ visible: false, isEdit: false, loading: false, id: 0 })
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  path: '',
  component: '',
  icon: '',
  parent_id: null as number | null,
  order: 0,
  hidden: false,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  path: [{ required: true, message: '请输入路径', trigger: 'blur' }],
}

function resetForm() {
  form.name = ''
  form.path = ''
  form.component = ''
  form.icon = ''
  form.parent_id = null
  form.order = 0
  form.hidden = false
  buildTreeFlat()
}

async function fetchList() {
  const resp = await request.get<ApiResp<{ items: MenuNode[] }>>('/menus')
  tree.value = resp.data.data.items || []
  buildTreeFlat()
}

function openCreate() {
  dialog.visible = true
  dialog.isEdit = false
  dialog.id = 0
  resetForm()
}

function openEdit(row: MenuNode) {
  dialog.visible = true
  dialog.isEdit = true
  dialog.id = row.id
  form.name = row.name
  form.path = row.path
  form.component = row.component
  form.icon = row.icon
  form.parent_id = (row.parent_id ?? null) as any
  form.order = row.order
  form.hidden = row.hidden
}

async function submit() {
  await formRef.value?.validate()
  dialog.loading = true
  try {
    const payload = {
      name: form.name,
      path: form.path,
      component: form.component,
      icon: form.icon,
      parent_id: form.parent_id === null ? undefined : form.parent_id,
      order: form.order,
      hidden: form.hidden,
    }
    if (dialog.isEdit) await request.put<ApiResp>('/menus/' + dialog.id, payload)
    else await request.post<ApiResp>('/menus', payload)

    ElMessage.success('保存成功')
    dialog.visible = false
    fetchList()
  } finally {
    dialog.loading = false
  }
}

async function removeOne(row: MenuNode) {
  await ElMessageBox.confirm(`确认删除菜单「${row.name}」？`, '提示', { type: 'warning' })
  await request.delete<ApiResp>('/menus/' + row.id)
  ElMessage.success('删除成功')
  fetchList()
}

function buildTreeFlat() {
  const out: MenuNode[] = []
  const walk = (nodes: MenuNode[]) => {
    for (const n of nodes || []) {
      out.push(n)
      if (n.children?.length) walk(n.children)
    }
  }
  walk(tree.value)
  treeFlat.value = out
}

onMounted(fetchList)
</script>

<style scoped>
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
.actions {
  display: flex;
  gap: 10px;
}
</style>

