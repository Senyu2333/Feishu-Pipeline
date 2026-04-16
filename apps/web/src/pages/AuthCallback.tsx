import { useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Spin, message } from 'antd'

const API_BASE = '/api'

// 飞书 OAuth code 仅单次有效；React StrictMode 会重复执行 effect，需合并为同一请求。
const feishuLoginInFlight = new Map<string, Promise<Response>>()

function postFeishuSSOLogin(code: string): Promise<Response> {
  let p = feishuLoginInFlight.get(code)
  if (!p) {
    p = fetch(`${API_BASE}/auth/feishu/sso/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ code }),
    }).finally(() => {
      feishuLoginInFlight.delete(code)
    })
    feishuLoginInFlight.set(code, p)
  }
  return p
}

export default function AuthCallback() {
  const navigate = useNavigate()

  useEffect(() => {
    const handleCallback = async () => {
      // 从 URL 获取 code 和 state
      const params = new URLSearchParams(window.location.search)
      const code = params.get('code')
      const state = params.get('state')
      const savedState = sessionStorage.getItem('feishu_auth_state')

      // 验证 state
      if (state && savedState && state !== savedState) {
        message.error('state 验证失败，请重试')
        navigate({ to: '/' })
        return
      }

      if (!code) {
        message.error('未获取到授权码')
        navigate({ to: '/' })
        return
      }

      try {
        const res = await postFeishuSSOLogin(code)

        if (!res.ok) {
          throw new Error('登录失败')
        }

        // 清除 state
        sessionStorage.removeItem('feishu_auth_state')
        
        message.success('登录成功')
        navigate({ to: '/' })
      } catch (e) {
        message.error('登录失败，请重试')
        navigate({ to: '/' })
      }
    }

    handleCallback()
  }, [navigate])

  return (
    <div className="flex items-center justify-center min-h-screen">
      <Spin size="large" tip="正在登录..." />
    </div>
  )
}
