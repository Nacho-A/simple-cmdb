<template>
  <div class="page">
    <div class="grid">
      <div class="card glass">
        <div class="title">欢迎</div>
        <div class="value">{{ displayName }}</div>
        <div class="sub">登录校验：JWT + Casbin + 动态菜单路由</div>
      </div>

      <div class="card glass">
        <div class="title">当前角色</div>
        <div class="value">
          <el-tag v-for="r in roles" :key="r" type="success" effect="dark" style="margin-right: 8px">
            {{ r }}
          </el-tag>
        </div>
        <div class="sub">菜单级权限由 role_menus 控制</div>
      </div>

      <div class="card glass">
        <div class="title">快捷入口</div>
        <div class="value">
          <el-button type="primary" plain @click="$router.push('/cmdb/assets')">资产管理</el-button>
          <el-button plain @click="$router.push('/system/user')">用户管理</el-button>
        </div>
        <div class="sub">未授权页面会被动态路由拦截</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const displayName = computed(() => userStore.userInfo?.nickname || userStore.userInfo?.username || '用户')
const roles = computed(() => userStore.roles || [])
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 14px;
}

.card {
  grid-column: span 12;
  padding: 16px;
  transition: transform 0.15s ease, background 0.15s ease;
}

.card:hover {
  transform: translateY(-1px);
  background: var(--cmdb-card-hover);
}

.title {
  opacity: 0.8;
  font-size: 13px;
}

.value {
  margin-top: 8px;
  font-size: 18px;
  font-weight: 750;
}

.sub {
  margin-top: 10px;
  opacity: 0.7;
  font-size: 12px;
}

@media (min-width: 960px) {
  .card {
    grid-column: span 4;
  }
}
</style>

