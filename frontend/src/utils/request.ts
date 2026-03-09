import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResp } from '@/types/api'
import { getToken } from './auth'
import { useUserStore } from '@/stores/user'

export const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (resp) => {
    if (resp.config.responseType === 'blob') {
      return resp
    }
    const data = resp.data as ApiResp
    if (data && typeof data.code === 'number' && data.code !== 200) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(data)
    }
    return resp
  },
  (err) => {
    const status = err?.response?.status
    if (status === 401) {
      ElMessage.error('未登录或登录已过期')
      useUserStore().logout()
      if (location.pathname !== '/login') location.href = '/login'
    } else if (status === 403) {
      ElMessage.error('无权限')
    } else {
      ElMessage.error('网络或服务器错误')
    }
    return Promise.reject(err)
  },
)

