import { useEffect, useMemo, useState } from 'react'
import { Avatar, Button, Card, Input, message as antdMessage } from 'antd'
import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'

const promptItems = [
  { key: 'user-story', label: 'User Story Mapping' },
  { key: 'tech-spec', label: 'Technical Specification' },
  { key: 'sla', label: 'SLA Definition' },
]

type ChatRole = 'user' | 'assistant' | 'system'

type ChatMessage = {
  id: string
  role: ChatRole
  content: string
}

type SessionDetail = {
  session: {
    id: string
    title: string
  }
  messages: Array<{
    id: string
    role: ChatRole
    content: string
  }>
}

type LocalDraftSession = {
  localID: string
  title: string
  hasChatted: boolean
  serverSessionID?: string
}

const activeSessionKey = 'activeRequirementSessionId'
const draftSessionKey = 'activeRequirementSessionDraft'

function readDraftSession(): LocalDraftSession | null {
  const raw = localStorage.getItem(draftSessionKey)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw) as LocalDraftSession
  } catch {
    localStorage.removeItem(draftSessionKey)
    return null
  }
}

function saveDraftSession(session: LocalDraftSession): void {
  localStorage.setItem(draftSessionKey, JSON.stringify(session))
  localStorage.setItem(activeSessionKey, session.localID)
}

function mapMessages(messages: SessionDetail['messages']): ChatMessage[] {
  return messages.map((item) => ({
    id: item.id,
    role: item.role,
    content: item.content,
  }))
}

export default function Home() {
  const [draftSession, setDraftSession] = useState<LocalDraftSession | null>(readDraftSession())
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)

  useEffect(() => {
    const listener = () => {
      setDraftSession(readDraftSession())
      setMessages([])
    }
    window.addEventListener('requirement:session-created', listener)
    return () => window.removeEventListener('requirement:session-created', listener)
  }, [])

  useEffect(() => {
    if (!draftSession?.serverSessionID) {
      return
    }

    void (async () => {
      try {
        const response = await fetch(`/api/sessions/${draftSession.serverSessionID}`, {
          credentials: 'include',
        })
        if (!response.ok) {
          return
        }
        const payload = (await response.json()) as { data?: SessionDetail }
        if (payload.data) {
          setMessages(mapMessages(payload.data.messages))
        }
      } catch {
        // ignore stale session fetch errors
      }
    })()
  }, [draftSession?.serverSessionID])

  const chatMessages = useMemo(() => {
    if (messages.length > 0) {
      return messages
    }
    return [
      {
        id: 'welcome',
        role: 'assistant' as const,
        content: 'Welcome to the Requirement Architect. I am ready to help you structure your project. Describe your requirement to start.',
      },
    ]
  }, [messages])

  const handleSend = async () => {
    const content = input.trim()
    if (!content || sending) {
      return
    }

    const current = draftSession ?? {
      localID: `draft-${Date.now()}`,
      title: `新需求 ${new Date().toLocaleString()}`,
      hasChatted: false,
    }

    if (!draftSession) {
      saveDraftSession(current)
      setDraftSession(current)
    }

    setInput('')
    setSending(true)

    const optimisticUserMessage: ChatMessage = {
      id: `local-${Date.now()}`,
      role: 'user',
      content,
    }
    setMessages((prev) => [...prev, optimisticUserMessage])

    try {
      let response: Response
      if (!current.serverSessionID) {
        response = await fetch('/api/sessions', {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            title: current.title,
            prompt: content,
          }),
        })
      } else {
        response = await fetch(`/api/sessions/${current.serverSessionID}/messages`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ content }),
        })
      }

      if (!response.ok) {
        throw new Error('发送消息失败')
      }

      const payload = (await response.json()) as { data?: SessionDetail }
      if (!payload.data) {
        throw new Error('响应数据不完整')
      }

      const nextDraft: LocalDraftSession = {
        ...current,
        hasChatted: true,
        serverSessionID: payload.data.session.id,
      }
      saveDraftSession(nextDraft)
      setDraftSession(nextDraft)
      setMessages(mapMessages(payload.data.messages))

      try {
        const autoPublishResponse = await fetch(`/api/sessions/${payload.data.session.id}/auto-publish-check`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ content }),
        })
        if (autoPublishResponse.ok) {
          const autoPayload = (await autoPublishResponse.json()) as {
            data?: { triggered?: boolean; reason?: string }
          }
          if (autoPayload.data?.triggered) {
            antdMessage.success('检测到排期需求，已自动触发发布流程')
          }
        }
      } catch {
        // auto publish check is best-effort
      }
    } catch (error) {
      setMessages((prev) => prev.filter((item) => item.id !== optimisticUserMessage.id))
      antdMessage.error(error instanceof Error ? error.message : '发送失败')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <TopNav />
      <Sidebar />
      <main className="ml-64 mt-14 h-[calc(100vh-3.5rem)] flex flex-col relative overflow-hidden">
        {/* Welcome Header */}
        <div className="pt-12 pb-6 px-12 text-center">
          <h1 className="text-4xl font-extrabold tracking-tight text-on-surface mb-2">Hello, Designer</h1>
          <p className="text-on-surface-variant font-medium text-lg">
            {draftSession ? `当前会话：${draftSession.title}` : 'Define your new enterprise requirements through guided dialogue.'}
          </p>
        </div>

        {/* Chat Conversation Container */}
        <div className="flex-1 overflow-y-auto px-6 md:px-24 py-8 flex flex-col gap-8 scroll-smooth">
          {chatMessages.map((item) => {
            const isUser = item.role === 'user'
            return (
              <div key={item.id} className={`flex items-start gap-4 max-w-[85%] ${isUser ? 'self-end flex-row-reverse' : ''}`}>
                <div className={`w-8 h-8 rounded-lg flex-shrink-0 flex items-center justify-center ${isUser ? 'bg-primary shadow-md' : 'bg-surface-container-high'}`}>
                  {isUser ? (
                    <span className="material-symbols-outlined text-white text-sm">person</span>
                  ) : (
                    <span className="material-symbols-outlined text-primary text-sm" style={{ fontVariationSettings: "'FILL' 1" }}>
                      auto_awesome
                    </span>
                  )}
                </div>
                <div className={`p-5 rounded-2xl shadow-sm border ${isUser ? 'bg-primary text-white rounded-tr-none border-primary/10' : 'bg-surface-container-lowest rounded-tl-none border-outline-variant/10'}`}>
                  <p className={`text-sm leading-relaxed whitespace-pre-wrap ${isUser ? 'text-white' : 'text-on-surface'}`}>{item.content}</p>
                </div>
              </div>
            )
          })}

          {messages.length === 0 ? (
            <Card size="small" className="!rounded-2xl !border-0 !shadow-sm max-w-[420px]" style={{ background: 'linear-gradient(135deg, #f0f7ff 0%, #e8f2fc 100%)' }}>
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2 font-semibold">
                  <Avatar size="small" className="!bg-primary" icon={<span className="material-symbols-outlined text-white text-xs">auto_awesome</span>} />
                  推荐起手
                </div>
              </div>
              <div className="flex flex-wrap gap-2">
                {promptItems.map((item) => (
                  <Button key={item.key} size="small" onClick={() => setInput(item.label)}>
                    {item.label}
                  </Button>
                ))}
              </div>
            </Card>
          ) : null}
        </div>

        {/* Input Area */}
        <div className="px-6 md:px-24 pb-6">
          <div className="bg-surface-container-lowest rounded-2xl border border-outline-variant p-2 flex items-center gap-2">
            <button type="button" className="w-8 h-8 rounded-full border-0 bg-transparent text-on-surface-variant cursor-pointer flex items-center justify-center hover:bg-surface-variant/50">
              <span className="material-symbols-outlined">attach_file</span>
            </button>
            <input
              type="text"
              placeholder="Describe your requirement..."
              value={input}
              onChange={(event) => setInput(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' && !event.shiftKey) {
                  event.preventDefault()
                  void handleSend()
                }
              }}
              className="flex-1 bg-transparent border-0 outline-none text-on-surface placeholder:text-on-surface/40"
            />
            <button
              type="button"
              disabled={sending}
              onClick={() => void handleSend()}
              className="w-8 h-8 rounded-full border-0 bg-primary text-white cursor-pointer flex items-center justify-center hover:opacity-90 disabled:opacity-50"
            >
              <span className="material-symbols-outlined text-sm">send</span>
            </button>
          </div>
          <div className="text-center text-xs text-on-surface/40 mt-2">AetherFlow AI can make mistakes. Verify critical project details.</div>
        </div>
      </main>
    </div>
  )
}
