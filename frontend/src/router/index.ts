import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useMenuStore } from '@/stores/menu'
import { getToken } from '@/utils/auth'
import type { MenuNode } from '@/types/menu'

const MainLayout = () => import('@/layouts/MainLayout.vue')
const LoginLayout = () => import('@/layouts/LoginLayout.vue')
const BlankView = () => import('@/layouts/BlankView.vue')

const baseRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    component: LoginLayout,
    children: [
      {
        path: '',
        name: 'Login',
        component: () => import('@/views/login/index.vue'),
      },
    ],
  },
  {
    path: '/',
    name: 'Main',
    component: MainLayout,
    redirect: '/dashboard',
    children: [],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/dashboard/index.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes: baseRoutes,
})

const viewModules = import.meta.glob('../views/**/*.vue')

function addDynamicRoutes(menus: MenuNode[]): boolean {
  const main = router.getRoutes().find((r) => r.name === 'Main')
  if (!main) return false

  const children: RouteRecordRaw[] = []

  const walk = (nodes: MenuNode[]) => {
    for (const m of nodes || []) {
      if (m.hidden) continue
      if (m.component) {
        const key = `../${m.component}`.replace(/^\.\.\//, '../')
        const loader = (viewModules as any)[key]
        const comp = loader || (viewModules as any)[`../views/${m.component.replace(/^views\//, '')}`]

        children.push({
          path: m.path,
          name: m.path,
          component: comp ? comp : () => import('@/views/dashboard/index.vue'),
          meta: { title: m.name, icon: m.icon },
        })
      } else if (m.children && m.children.length) {
        walk(m.children)
      }
    }
  }

  walk(menus)

  let added = false
  for (const r of children) {
    if (!router.hasRoute(r.name!)) {
      router.addRoute('Main', r)
      added = true
    }
  }
  return added
}

router.beforeEach(async (to) => {
  if (to.path === '/login') return true

  const token = getToken()
  if (!token) return { path: '/login', replace: true }

  const menuStore = useMenuStore()
  const userStore = useUserStore()

  // 持久化恢复后可能 routesAdded=true 但 menuTree 为空（或 router 重建后尚未挂载），需重新拉取
  if (menuStore.routesAdded && menuStore.menuTree.length === 0) {
    menuStore.routesAdded = false
  }

  if (!menuStore.routesAdded) {
    try {
      await userStore.fetchMe()
      addDynamicRoutes(menuStore.menuTree)
      menuStore.markRoutesAdded()
      return { path: to.fullPath, replace: true }
    } catch {
      userStore.logout()
      return { path: '/login', replace: true }
    }
  }

  // 刷新时 router 会重建，动态路由未挂载；重新添加后必须用当前 URL（to.fullPath）replace，否则会落到 catch-all 变成仪表盘
  const added = addDynamicRoutes(menuStore.menuTree)
  if (added) return { path: to.fullPath, replace: true }
  return true
})

export default router

