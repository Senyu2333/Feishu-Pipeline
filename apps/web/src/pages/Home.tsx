import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import Sidebar from '../components/Sidebar'

const promptItems = [
  { key: 'user-story', label: '创建用户故事地图', prompt: '我想为一个电商平台创建用户故事地图...' },
  { key: 'tech-spec', label: '技术规格说明', prompt: '我需要为微服务架构编写技术规格说明...' },
  { key: 'sla', label: 'SLA 定义', prompt: '帮我定义核心服务的 SLA 要求...' },
  { key: 'api-design', label: 'API 设计', prompt: '设计一个 RESTful API 用于订单管理...' },
  { key: 'database', label: '数据库设计', prompt: '为社交网络应用设计数据库 Schema...' },
  { key: 'deployment', label: '部署方案', prompt: '制定 Kubernetes 集群的部署方案...' },
  { key: 'security', label: '安全评估', prompt: '评估 Web 应用的安全风险并提出建议...' },
  { key: 'performance', label: '性能优化', prompt: '分析并优化现有系统的性能瓶颈...' },
]

export default function Home() {
  const navigate = useNavigate()
  const [input, setInput] = useState('')
  const [creating, setCreating] = useState(false)
  const [convCollapsed, setConvCollapsed] = useState(false)

  // 创建新会话 - 立即跳转，首条消息通过 sessionStorage 传给 Session 页面
  const createSession = async (title: string, prompt: string) => {
    if (creating) return
    setCreating(true)
    try {
      const res = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ title, prompt })
      })
      if (res.ok) {
        const data = await res.json()
        if (data.data?.session?.id) {
          // 把首条消息内容暂存，Session 页面读取后立即展示+发送
          sessionStorage.setItem(`pending_msg_${data.data.session.id}`, prompt)
          navigate({ to: `/sessions/${data.data.session.id}` })
        }
      }
    } catch (err) {
      console.error('Failed to create session:', err)
    } finally {
      setCreating(false)
    }
  }

  const handleSend = () => {
    if (!input.trim() || creating) return
    const title = input.slice(0, 50) + (input.length > 50 ? '...' : '')
    createSession(title, input)
  }

  const handlePromptClick = (prompt: string) => {
    createSession(prompt.slice(0, 50), prompt)
  }

  // 计算左边距：折叠时 80px，展开时 336px (80+256)
  const sidebarWidth = convCollapsed ? 80 : 336

  return (
    <div className="min-h-screen bg-background">
      <Sidebar convCollapsed={convCollapsed} onConvCollapse={setConvCollapsed} />
      <main className="h-screen flex flex-col relative overflow-hidden transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        {/* 欢迎页 - 类似豆包风格 */}
        <div className="flex-1 flex flex-col items-center justify-center px-8">
          <h1 className="text-3xl font-bold text-on-surface mb-8">有什么我能帮你的吗？</h1>
          
          {/* 快捷提示按钮 */}
          <div className="flex flex-wrap justify-center gap-3 max-w-3xl">
            {promptItems.map(item => (
              <button
                key={item.key}
                onClick={() => handlePromptClick(item.prompt)}
                disabled={creating}
                className="px-4 py-2.5 bg-surface-container-low hover:bg-surface-container text-on-surface text-sm rounded-xl transition-all border border-outline-variant/50 disabled:opacity-50 hover:shadow-sm"
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

        {/* Input Area */}
        <div className="px-8 md:px-24 pb-8">
          <div className="bg-surface-container-lowest rounded-2xl border border-outline-variant p-4 flex items-start gap-3 shadow-sm">
            <button type="button" className="w-9 h-9 mt-1 rounded-full border-0 bg-transparent text-on-surface-variant cursor-pointer flex items-center justify-center hover:bg-surface-variant/50 flex-shrink-0">
              <span className="material-symbols-outlined">attach_file</span>
            </button>
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSend()}
              placeholder="描述你的需求..."
              disabled={creating}
              className="flex-1 bg-transparent border-0 outline-none text-on-surface text-base placeholder:text-on-surface/40 disabled:opacity-50 py-1"
            />
            <button
              type="button"
              onClick={handleSend}
              disabled={!input.trim() || creating}
              className="w-9 h-9 rounded-full border-0 bg-primary text-white cursor-pointer flex items-center justify-center hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed flex-shrink-0"
            >
              <span className="material-symbols-outlined">{creating ? 'progress_activity' : 'send'}</span>
            </button>
          </div>
          <div className="text-center text-xs text-on-surface/40 mt-2">内容由 AI 生成，请仔细甄别</div>
        </div>
      </main>
    </div>
  )
}
