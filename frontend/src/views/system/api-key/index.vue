<template>
  <div class="page page--table">
    <SearchForm @search="fetchList" @reset="onReset">
      <template #extra-actions>
        <el-button type="success" plain @click="openCreate">新增 API Key</el-button>
      </template>
    </SearchForm>

    <div class="table-block">
      <ElTablePro :total="total" v-model:page="page" v-model:pageSize="pageSize" @change="fetchList">
        <template #toolbar>
          <div class="title">API Key 管理</div>
          <div class="meta">用于 Jenkins / 脚本 / 其他系统调用</div>
        </template>

        <div class="table-inner">
          <el-table :data="items" border height="100%">
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column prop="scope" label="权限范围" width="120">
              <template #default="{ row }">
                <el-tag :type="row.scope === 'write' ? 'warning' : 'info'" effect="dark">
                  {{ row.scope === 'write' ? '读写' : '只读' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'" effect="dark">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button size="small" text type="danger" @click="removeOne(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </ElTablePro>
    </div>

    <el-dialog v-model="dialog.visible" title="新增 API Key" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如：jenkins-ci" />
        </el-form-item>
        <el-form-item label="权限范围" prop="scope">
          <el-radio-group v-model="form.scope">
            <el-radio value="read">只读 (GET)</el-radio>
            <el-radio value="write">读写 (全部)</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="dialog.loading" @click="submit">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="keyDialog.visible" title="API Key 已创建" width="560px" :close-on-click-modal="false">
      <el-alert type="warning" :closable="false" show-icon style="margin-bottom: 16px">
        <template #title>请立即复制保存，此 Key 仅显示一次</template>
      </el-alert>
      <el-input v-model="keyDialog.key" readonly>
        <template #append>
          <el-button @click="copyKey">复制</el-button>
        </template>
      </el-input>
      <template #footer>
        <el-button type="primary" @click="keyDialog.visible = false">我已保存</el-button>
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

type APIKeyRow = {
  id: number
  name: string
  scope: string
  status: number
  created_at: string
}

const items = ref<APIKeyRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const dialog = reactive({ visible: false, loading: false })
const formRef = ref<FormInstance>()
const form = reactive({ name: '', scope: 'read' })
const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  scope: [{ required: true, message: '请选择权限范围', trigger: 'change' }],
}

const keyDialog = reactive({ visible: false, key: '' })

async function fetchList() {
  const resp = await request.get<ApiResp<{ list: APIKeyRow[]; total: number }>>('/api-keys')
  items.value = resp.data.data.list || []
  total.value = resp.data.data.total || 0
}

function onReset() {
  page.value = 1
  fetchList()
}

function openCreate() {
  form.name = ''
  form.scope = 'read'
  dialog.visible = true
}

async function submit() {
  await formRef.value?.validate()
  dialog.loading = true
  try {
    const resp = await request.post<ApiResp<{ key: string }>>('/api-keys', {
      name: form.name.trim(),
      scope: form.scope,
    })
    dialog.visible = false
    keyDialog.key = resp.data.data.key
    keyDialog.visible = true
    fetchList()
  } finally {
    dialog.loading = false
  }
}

async function removeOne(row: APIKeyRow) {
  await ElMessageBox.confirm(`确认删除 API Key「${row.name}」？删除后无法恢复。`, '提示', { type: 'warning' })
  await request.delete<ApiResp>('/api-keys/' + row.id)
  ElMessage.success('删除成功')
  fetchList()
}

function copyKey() {
  navigator.clipboard.writeText(keyDialog.key)
  ElMessage.success('已复制到剪贴板')
}

onMounted(() => {
  fetchList()
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
.title {
  font-weight: 800;
}
.meta {
  opacity: 0.7;
  font-size: 12px;
  align-self: center;
}
</style>