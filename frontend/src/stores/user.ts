import { defineStore } from 'pinia'
import { request } from '@/utils/request'
import type { LoginResp, UserInfo } from '@/types/user'
import type { ApiResp } from '@/types/api'
import { clearToken, setToken } from '@/utils/auth'
import { useMenuStore } from './menu'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '' as string,
    userInfo: null as UserInfo | null,
    roles: [] as string[],
  }),
  persist: true,
  actions: {
    async login(username: string, password: string) {
      const resp = await request.post<ApiResp<LoginResp>>('/login', { username, password })
      const data = resp.data.data
      this.token = data.token
      this.userInfo = data.userInfo
      this.roles = data.userInfo.roles || []
      setToken(this.token)
      return data
    },
    async fetchMe() {
      const resp = await request.get<ApiResp<{ userInfo: UserInfo; menus: any[] }>>('/me')
      const { userInfo, menus } = resp.data.data
      this.userInfo = userInfo
      this.roles = userInfo.roles || []
      useMenuStore().setMenus(menus as any)
      return resp.data.data
    },
    logout() {
      this.token = ''
      this.userInfo = null
      this.roles = []
      useMenuStore().reset()
      clearToken()
    },
  },
})

