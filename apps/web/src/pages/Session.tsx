import { useEffect, useState } from 'react'
import { useParams } from '@tanstack/react-router'
import Sidebar from '../components/Sidebar'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  createdAt: string
}

interface SessionDetail {
  session: {
    id: string
    title: string
    status: string
  }
  messages: Message[]
}

export default function Session() {
  const { sessionId } = useParams({ strict: false })
  const [session, setSession] = useState<SessionDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  // 左侧导航固定 80px
  const sidebarWidth = 80

  // 获取会话详情
  useEffect(() => {
    fetch(`/api/sessions/${sessionId}`, { credentials: 'include' })
      .then(res => {
        if (res.ok) return res.json()
        throw new Error('Failed to load session')
      })
      .then(data => {
        if (data.data) setSession(data.data)
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [sessionId])

  // 发送消息
  const handleSend = async () => {
    if (!input.trim() || sending) return
    setSending(true)
    try {
      const res = await fetch(`/api/sessions/${sessionId}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ content: input })
      })
      if (res.ok) {
        const data = await res.json()
        if (data.data) setSession(data.data)
        setInput('')
      }
    } catch (err) {
      console.error('Failed to send message:', err)
    } finally {
      setSending(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <Sidebar />
        <main className="h-screen flex items-center justify-center transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
          <span className="material-symbols-outlined text-primary text-2xl animate-spin">progress_activity</span>
        </main>
      </div>
    )
  }

  if (!session) {
    return (
      <div className="min-h-screen bg-background">
        <Sidebar />
        <main className="h-screen flex items-center justify-center transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
          <div className="text-center">
            <p className="text-on-surface-variant">Session not found</p>
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="h-screen flex flex-col relative overflow-hidden transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        {/* Header */}
        <div className="h-16 px-6 flex items-center border-b border-outline-variant/20 bg-white/50">
          <h1 className="text-lg font-semibold text-on-surface truncate">{session.session.title}</h1>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto px-6 md:px-12 py-8 flex flex-col gap-6">
          {session.messages.length === 0 ? (
            <div className="text-center text-on-surface-variant py-12">
              <p>Start a conversation...</p>
            </div>
          ) : (
            session.messages.map((msg) => (
              <div
                key={msg.id}
                className={`flex items-start gap-4 max-w-[85%] ${msg.role === 'user' ? 'self-end flex-row-reverse' : ''}`}
              >
                <div className={`w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center ${
                  msg.role === 'user' ? 'bg-primary shadow-md' : 'bg-surface-container-high'
                }`}>
                  <span className={`material-symbols-outlined text-sm ${msg.role === 'user' ? 'text-white' : 'text-primary'}`}>
                    {msg.role === 'user' ? 'person' : 'auto_awesome'}
                  </span>
                </div>
                <div className={`p-4 rounded-2xl ${
                  msg.role === 'user'
                    ? 'bg-primary text-white rounded-tr-none shadow-md'
                    : 'bg-surface-container-lowest border border-outline-variant/10 rounded-tl-none shadow-sm'
                }`}>
                  <p className="text-sm leading-relaxed whitespace-pre-wrap">{msg.content}</p>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Input */}
        <div className="px-6 md:px-12 pb-6">
          <div className="bg-surface-container-lowest rounded-2xl border border-outline-variant p-2 flex items-center gap-2">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSend()}
              placeholder="Type your message..."
              className="flex-1 bg-transparent border-0 outline-none text-on-surface placeholder:text-on-surface/40 px-2"
              disabled={sending}
            />
            <button
              onClick={handleSend}
              disabled={!input.trim() || sending}
              className="w-8 h-8 rounded-full border-0 bg-primary text-white cursor-pointer flex items-center justify-center hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span className="material-symbols-outlined text-sm">{sending ? 'progress_activity' : 'send'}</span>
            </button>
          </div>
        </div>
      </main>
    </div>
  )
}
