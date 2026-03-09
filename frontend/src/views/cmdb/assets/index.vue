<template>
  <div class="page page--table">
    <SearchForm @search="onSearch" @reset="onReset">
      <div class="filters">
        <el-input v-model="query.service_name" placeholder="服务名称" clearable />
        <el-input v-model="query.ip" placeholder="IP（私网/公网）" clearable />
        <el-input v-model="query.owner" placeholder="负责人" clearable />
        <el-input v-model="query.tags" placeholder="标签 tags（逗号分隔）" clearable />
        <el-input v-model="query.label" placeholder='labels 模糊匹配（如 "env" / "prod"）' clearable />
      </div>
      <template #extra-actions>
        <el-button type="success" plain @click="openCreate">新增</el-button>
        <el-button type="danger" plain :disabled="!selection.length" @click="batchDelete">批量删除</el-button>
        <el-button plain @click="exportExcel">导出Excel</el-button>
      </template>
    </SearchForm>

    <div class="table-block">
      <ElTablePro :total="total" v-model:page="page" v-model:pageSize="pageSize" @change="fetchList">
        <template #toolbar>
          <div class="title">CMDB资产</div>
          <div class="meta">共 {{ total }} 条</div>
        </template>

        <div class="table-inner">
          <el-table
            :data="items"
            border
            height="100%"
            @selection-change="selection = $event"
          >
        <el-table-column type="selection" width="48" />
        <el-table-column prop="service_name" label="服务名称" min-width="160" />
        <el-table-column prop="private_ip" label="私网IP" min-width="160" />
        <el-table-column prop="public_ip" label="公网IP" min-width="140" />
        <el-table-column label="labels" min-width="240">
          <template #default="{ row }">
            <div class="tags">
              <el-tag
                v-for="(v, k) in row.labels || {}"
                :key="k"
                effect="dark"
                size="small"
                style="margin-right: 6px; margin-bottom: 6px"
              >
                {{ k }}:{{ v }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="tags" label="tags" min-width="140" />
        <el-table-column prop="owner" label="负责人" min-width="120" />
        <el-table-column prop="cloud_provider" label="云供应商" min-width="120" />
        <el-table-column prop="region" label="地域" min-width="110" />
        <el-table-column prop="status" label="状态" min-width="110" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" text @click="removeOne(row)">删除</el-button>
          </template>
        </el-table-column>
          </el-table>
        </div>
      </ElTablePro>
    </div>

    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑资产' : '新增资产'" width="720px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="服务名称" prop="service_name">
          <el-input v-model="form.service_name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="私网IP">
          <el-input v-model="form.private_ip" placeholder="支持多个逗号分隔" />
        </el-form-item>
        <el-form-item label="公网IP">
          <el-input v-model="form.public_ip" />
        </el-form-item>
        <el-form-item label="labels">
          <LabelsEditor v-model="form.labels" />
        </el-form-item>
        <el-form-item label="tags">
          <el-input v-model="form.tags" placeholder="prod,important" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.owner" />
        </el-form-item>
        <el-form-item label="云供应商">
          <el-select v-model="form.cloud_provider" filterable clearable placeholder="选择云厂商">
            <el-option v-for="p in providers" :key="p" :label="p" :value="p" />
          </el-select>
        </el-form-item>
        <el-form-item label="地域">
          <el-input v-model="form.region" />
        </el-form-item>
        <el-form-item label="实例规格">
          <el-input v-model="form.instance_type" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" clearable>
            <el-option label="running" value="running" />
            <el-option label="stopped" value="stopped" />
            <el-option label="terminated" value="terminated" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
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
import SearchForm from '@/components/SearchForm.vue'
import ElTablePro from '@/components/ElTablePro.vue'
import LabelsEditor from '@/components/LabelsEditor.vue'
import { request } from '@/utils/request'
import type { ApiResp } from '@/types/api'
import type { CMDBAsset } from '@/types/asset'

const items = ref<CMDBAsset[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const selection = ref<CMDBAsset[]>([])
const providers = ref<string[]>([])

const query = reactive({
  service_name: '',
  ip: '',
  owner: '',
  tags: '',
  label: '',
})

const dialog = reactive({
  visible: false,
  isEdit: false,
  loading: false,
  id: 0,
})

const formRef = ref<FormInstance>()
const form = reactive({
  service_name: '',
  private_ip: '',
  public_ip: '',
  labels: {} as Record<string, string>,
  tags: '',
  owner: '',
  cloud_provider: '',
  region: '',
  instance_type: '',
  status: '',
  remark: '',
})

const rules: FormRules = {
  service_name: [{ required: true, message: '服务名称必填', trigger: 'blur' }],
}

function resetForm() {
  form.service_name = ''
  form.private_ip = ''
  form.public_ip = ''
  form.labels = {}
  form.tags = ''
  form.owner = ''
  form.cloud_provider = ''
  form.region = ''
  form.instance_type = ''
  form.status = ''
  form.remark = ''
}

async function fetchProviders() {
  const resp = await request.get<ApiResp<{ items: string[] }>>('/cmdb/cloud-providers')
  providers.value = resp.data.data.items || []
}

async function fetchList() {
  const resp = await request.get<
    ApiResp<{ items: CMDBAsset[]; total: number; page: number; page_size: number }>
  >('/cmdb/assets', {
    params: {
      page: page.value,
      page_size: pageSize.value,
      ...query,
    },
  })
  items.value = resp.data.data.items || []
  total.value = resp.data.data.total || 0
}

function onSearch() {
  page.value = 1
  fetchList()
}

function onReset() {
  query.service_name = ''
  query.ip = ''
  query.owner = ''
  query.tags = ''
  query.label = ''
  page.value = 1
  fetchList()
}

function openCreate() {
  dialog.visible = true
  dialog.isEdit = false
  dialog.id = 0
  resetForm()
}

function openEdit(row: CMDBAsset) {
  dialog.visible = true
  dialog.isEdit = true
  dialog.id = row.id
  form.service_name = row.service_name
  form.private_ip = row.private_ip
  form.public_ip = row.public_ip
  form.labels = (row.labels || {}) as any
  form.tags = row.tags
  form.owner = row.owner
  form.cloud_provider = row.cloud_provider
  form.region = row.region
  form.instance_type = row.instance_type
  form.status = row.status
  form.remark = row.remark
}

async function submit() {
  await formRef.value?.validate()
  dialog.loading = true
  try {
    if (dialog.isEdit) {
      await request.put<ApiResp>('/cmdb/assets/' + dialog.id, { ...form })
    } else {
      await request.post<ApiResp>('/cmdb/assets', { ...form })
    }
    ElMessage.success('保存成功')
    dialog.visible = false
    fetchList()
  } finally {
    dialog.loading = false
  }
}

async function removeOne(row: CMDBAsset) {
  await ElMessageBox.confirm(`确认删除「${row.service_name}」？`, '提示', { type: 'warning' })
  await request.delete<ApiResp>('/cmdb/assets/' + row.id)
  ElMessage.success('删除成功')
  fetchList()
}

async function batchDelete() {
  await ElMessageBox.confirm(`确认批量删除选中的 ${selection.value.length} 条资产？`, '提示', {
    type: 'warning',
  })
  await request.post<ApiResp>('/cmdb/assets/batch-delete', { ids: selection.value.map((x) => x.id) })
  ElMessage.success('删除成功')
  selection.value = []
  fetchList()
}

async function exportExcel() {
  try {
    const resp = await request.get('/cmdb/assets/export', {
      params: { ...query },
      responseType: 'blob',
    })
    const contentType = resp.headers?.['content-type'] || ''
    if (contentType.includes('application/json')) {
      const text = await (resp.data as Blob).text()
      const json = JSON.parse(text) as ApiResp
      ElMessage.error(json.message || '导出失败')
      return
    }
    const blob = resp.data instanceof Blob ? resp.data : new Blob([resp.data])
    if (blob.size === 0) {
      ElMessage.error('导出结果为空')
      return
    }
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `cmdb_assets_${new Date().toISOString().slice(0, 19).replace(/[-:T]/g, '')}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (e: any) {
    if (e?.response?.data instanceof Blob) {
      const text = await e.response.data.text()
      try {
        const json = JSON.parse(text)
        ElMessage.error(json.message || '导出失败')
      } catch {
        ElMessage.error('导出失败')
      }
    } else {
      ElMessage.error('导出失败')
    }
  }
}

onMounted(async () => {
  await fetchProviders()
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
  grid-template-columns: repeat(12, 1fr);
  gap: 10px;
}
.filters :deep(.el-input),
.filters :deep(.el-select) {
  grid-column: span 12;
}
.title {
  font-weight: 800;
}
.meta {
  opacity: 0.75;
  font-size: 12px;
  align-self: center;
}
.tags {
  display: flex;
  flex-wrap: wrap;
}
@media (min-width: 960px) {
  .filters :deep(.el-input) {
    grid-column: span 3;
  }
  .filters :deep(.el-select) {
    grid-column: span 3;
  }
}
</style>

