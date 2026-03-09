<template>
  <template v-for="n in nodes" :key="n.path">
    <el-sub-menu v-if="hasChildren(n)" :index="n.path">
      <template #title>
        <el-icon v-if="n.icon"><component :is="n.icon" /></el-icon>
        <span>{{ n.name }}</span>
      </template>
      <SideMenu :nodes="n.children || []" />
    </el-sub-menu>

    <el-menu-item v-else :index="n.path">
      <el-icon v-if="n.icon"><component :is="n.icon" /></el-icon>
      <span>{{ n.name }}</span>
    </el-menu-item>
  </template>
</template>

<script setup lang="ts">
import type { MenuNode } from '@/types/menu'

defineProps<{ nodes: MenuNode[] }>()

function hasChildren(n: MenuNode) {
  return !!(n.children && n.children.filter((c) => !c.hidden).length)
}
</script>

