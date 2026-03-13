<template>
  <div class="page page--table">
    <SearchForm @search="onSearch" @reset="onReset">
      <div class="filters">
        <el-input v-model="query.service_name" placeholder="服务名称" clearable />
        <el-input v-model="query.private_ip" placeholder="私网IP" clearable />
        <el-input v-model="query.public_ip" placeholder="公网IP" clearable />
      </div>
    </SearchForm>

    <div class="table-block">
      <ElTablePro :total="total" v-model:page="page" v-model:pageSize="pageSize" @change="fetchList">
        <template #toolbar>
          <div class="title">服务列表</div>
          <div class="meta">共 {{ total }} 个服务</div>
        </template>

        <div class="table-inner">
          <el-table :data="items" border height="100%">
            <el-table-column prop="service_name" label="服务名称" min-width="180" />
            <el-table-column label="私网IP" min-width="280">
              <template #default="{ row }">
                <el-tag
                  v-for="ip in row.private_ips"
                  :key="ip"
                  type="info"
                  effect="dark"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 4px"
                >
                  {{ ip }}
                </el-tag>
                <span v-if="!row.private_ips?.length" class="empty-text">-</span>
              </template>
            </el-table-column>
            <el-table-column label="公网IP" min-width="200">
              <template #default="{ row }">
                <el-tag
                  v-for="ip in row.public_ips"
                  :key="ip"
                  type="warning"
                  effect="dark"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 4px"
                >
                  {{ ip }}
                </el-tag>
                <span v-if="!row.public_ips?.length" class="empty-text">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="asset_count" label="资产数量" width="100" align="center" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="primary" text @click="viewAssets(row)">
                  查看资产
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </ElTablePro>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import SearchForm from '@/components/SearchForm.vue'
import ElTablePro from '@/components/ElTablePro.vue'
import { request } from '@/utils/request'
import type { ApiResp } from '@/types/api'
import type { ServiceItem } from '@/types/service'

const router = useRouter()
const items = ref<ServiceItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const query = reactive({
  service_name: '',
  private_ip: '',
  public_ip: '',
})

async function fetchList() {
  const resp = await request.get<
    ApiResp<{ items: ServiceItem[]; total: number; page: number; page_size: number }>
  >('/cmdb/services', {
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
  query.private_ip = ''
  query.public_ip = ''
  page.value = 1
  fetchList()
}

function viewAssets(row: ServiceItem) {
  router.push({ path: '/cmdb/assets', query: { service_name: row.service_name } })
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
.filters {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.title {
  font-weight: 800;
}
.meta {
  opacity: 0.75;
  font-size: 12px;
  align-self: center;
}
.empty-text {
  color: #999;
}
</style>