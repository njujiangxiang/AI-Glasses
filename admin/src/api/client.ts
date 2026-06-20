// 后台前端共用 HTTP 客户端。它把管理员 JWT 保存在 localStorage 中，为受保护接口附加
// Bearer 凭证，并解包后端统一的 { data } 响应结构，同时把后端错误消息暴露给页面展示。
const browserStorage = typeof globalThis.localStorage !== 'undefined' && typeof globalThis.localStorage.getItem === 'function' ? globalThis.localStorage : null
const storage = browserStorage
let token: string | null = storage?.getItem('admin_token') || null

export type CurrentUser = {
  id: number
  username: string
  display_name: string
  name: string
  avatar_size: number
  has_avatar?: boolean
  org_code?: string
  org_name?: string
  company_name?: string
  role_id?: number
}

// setToken 保存登录成功后返回的管理员访问令牌。
export function setToken(t: string) {
  token = t
  storage?.setItem('admin_token', t)
}

// setCurrentUser 保存登录成功后返回的用户、公司和头像状态，并通知应用壳刷新右上角信息。
export function setCurrentUser(user: CurrentUser) {
  storage?.setItem('admin_user', JSON.stringify(user))
  window.dispatchEvent(new CustomEvent('admin-user-updated'))
}

// getCurrentUser 读取当前登录用户信息，用于顶部用户区展示。
export function getCurrentUser(): CurrentUser | null {
  const raw = storage?.getItem('admin_user')
  if (!raw) return null
  try {
    return JSON.parse(raw) as CurrentUser
  } catch {
    return null
  }
}

// clearToken 清除本地登录态，用于退出登录或令牌失效后的清理。
export function clearToken() {
  token = null
  storage?.removeItem('admin_token')
  storage?.removeItem('admin_user')
}

// authHeaders 根据当前 token 生成请求头。
function authHeaders(): Record<string, string> {
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// apiGet 发送带认证信息的 GET 请求，并返回后端 data 字段。
export async function apiGet<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, { ...init, headers: { ...authHeaders(), ...(init?.headers || {}) } })
  return unwrap<T>(res)
}

// apiPost 发送 JSON POST 请求，并返回后端 data 字段。
export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(body || {})
  })
  return unwrap<T>(res)
}

// apiUpload 发送 multipart/form-data POST 请求，并返回后端 data 字段。
export async function apiUpload<T>(path: string, formData: FormData): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: authHeaders(),
    body: formData
  })
  return unwrap<T>(res)
}

// unwrap 统一解析后端响应，失败时抛出后端返回的错误消息。401 自动清除登录态并跳转登录页。
async function unwrap<T>(res: Response): Promise<T> {
  if (res.status === 401) {
    clearToken()
    if (window.location.pathname !== '/login') {
      window.location.replace('/login')
    }
    throw new ApiError(401, '登录已过期，请重新登录')
  }
  let json: any = null
  try {
    json = await res.json()
  } catch {
    json = null
  }
  if (!res.ok) throw new ApiError(res.status, json?.error?.message || 'request failed')
  return json.data as T
}
