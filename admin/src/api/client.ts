// 后台前端共用 HTTP 客户端。它把管理员 JWT 保存在 localStorage 中，为受保护接口附加
// Bearer 凭证，并解包后端统一的 { data } 响应结构，同时把后端错误消息暴露给页面展示。
let token: string | null = localStorage.getItem('admin_token')

// setToken 保存登录成功后返回的管理员访问令牌。
export function setToken(t: string) {
  token = t
  localStorage.setItem('admin_token', t)
}

// clearToken 清除本地登录态，用于退出登录或令牌失效后的清理。
export function clearToken() {
  token = null
  localStorage.removeItem('admin_token')
}

// authHeaders 根据当前 token 生成请求头。
function authHeaders(): Record<string, string> {
  return token ? { Authorization: `Bearer ${token}` } : {}
}

// apiGet 发送带认证信息的 GET 请求，并返回后端 data 字段。
export async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(path, { headers: authHeaders() })
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

// unwrap 统一解析后端响应，失败时抛出后端返回的错误消息。
async function unwrap<T>(res: Response): Promise<T> {
  const json = await res.json()
  if (!res.ok) throw new Error(json?.error?.message || 'request failed')
  return json.data as T
}