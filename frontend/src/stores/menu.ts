import { defineStore } from 'pinia'
import type { MenuNode } from '@/types/menu'

export const useMenuStore = defineStore('menu', {
  state: () => ({
    menuTree: [] as MenuNode[],
    routesAdded: false,
  }),
  persist: true,
  actions: {
    setMenus(menus: MenuNode[]) {
      this.menuTree = menus || []
    },
    markRoutesAdded() {
      this.routesAdded = true
    },
    reset() {
      this.menuTree = []
      this.routesAdded = false
    },
  },
})

