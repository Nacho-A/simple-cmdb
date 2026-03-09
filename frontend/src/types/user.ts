export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  status: number
  roles: string[]
}

export interface LoginResp {
  token: string
  userInfo: UserInfo
}

