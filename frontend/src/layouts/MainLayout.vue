<template>
  <div class="layout">
    <el-container class="layout-inner glass">
      <el-aside :width="collapsed ? '72px' : '260px'" class="aside">
        <div class="brand">
          <div class="logo" title="CMDB">CMDB</div>
          <span v-if="!collapsed" class="title">Cursor CMDB</span>
        </div>

        <el-scrollbar class="menu-scroll">
          <el-menu
            :default-active="route.path"
            :collapse="collapsed"
            :collapse-transition="false"
            router
            class="menu"
          >
            <SideMenu :nodes="visibleMenus" />
          </el-menu>
        </el-scrollbar>
      </el-aside>

      <el-container>
        <el-header class="header">
          <div class="left">
            <el-button text class="icon-btn" @click="collapsed = !collapsed">
              <el-icon><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
            </el-button>
            <el-breadcrumb separator="/" class="breadcrumb">
              <el-breadcrumb-item v-for="m in route.matched" :key="m.path">
                {{ (m.meta?.title as string) || ' ' }}
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>

          <div class="right">
            <el-tooltip :content="theme.dark ? '切换为浅色模式' : '切换为暗黑模式'" placement="bottom">
              <el-button text class="icon-btn" @click="theme.toggle()">
                <el-icon><Moon v-if="theme.dark" /><Sunny v-else /></el-icon>
              </el-button>
            </el-tooltip>

            <el-dropdown>
              <span class="user">
                <el-avatar :size="30">{{ avatarText }}</el-avatar>
                <span v-if="!collapsed" class="user-name">{{ displayName }}</span>
                <el-icon class="caret"><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="goMe">个人信息</el-dropdown-item>
                  <el-dropdown-item divided @click="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <el-main class="main">
          <RouterView />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Fold, Expand, Moon, Sunny, ArrowDown } from '@element-plus/icons-vue'
import SideMenu from '@/components/SideMenu.vue'
import { useMenuStore } from '@/stores/menu'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'

const collapsed = ref(false)
const route = useRoute()
const router = useRouter()
const menuStore = useMenuStore()
const userStore = useUserStore()
const theme = useThemeStore()

const visibleMenus = computed(() => menuStore.menuTree || [])
const displayName = computed(() => userStore.userInfo?.nickname || userStore.userInfo?.username || '用户')
const avatarText = computed(() => (displayName.value || 'U').slice(0, 1).toUpperCase())

function logout() {
  userStore.logout()
  router.replace('/login')
}

function goMe() {
  // 预留：后续可以放个人设置页
  router.push('/dashboard')
}
</script>

<style scoped>
.layout {
  height: 100%;
  padding: 14px;
}

.layout-inner {
  height: 100%;
  overflow: hidden;
}

.aside {
  border-right: 1px solid var(--cmdb-layout-border);
}

.brand {
  height: 56px;
  min-height: 56px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  user-select: none;
  flex-shrink: 0;
}

.logo {
  width: 36px;
  height: 36px;
  min-width: 36px;
  min-height: 36px;
  flex-shrink: 0;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(56, 189, 248, 0.35), rgba(168, 85, 247, 0.35));
  border: 1px solid var(--cmdb-glass-border);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
  line-height: 1;
  color: inherit;
}

.title {
  font-weight: 650;
  font-size: 15px;
  letter-spacing: 0.02em;
  opacity: 0.95;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.menu-scroll {
  height: calc(100% - 56px);
}

.menu {
  border-right: none;
  background: transparent;
}

.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--cmdb-layout-border);
  background: var(--cmdb-layout-bg);
}

.left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.breadcrumb {
  opacity: 0.85;
}

.right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  border-radius: 10px;
}

.user {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  border-radius: 12px;
  border: 1px solid var(--cmdb-user-border);
  background: var(--cmdb-user-bg);
  cursor: pointer;
  transition: transform 0.15s ease, background 0.15s ease;
}
.user:hover {
  transform: translateY(-1px);
  background: var(--cmdb-user-hover-bg);
}
.user-name {
  opacity: 0.95;
  font-weight: 600;
}
.caret {
  opacity: 0.8;
}

.main {
  height: calc(100% - 56px);
  padding: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.main :deep(.page) {
  flex: 1;
  min-height: 0;
}
</style>

