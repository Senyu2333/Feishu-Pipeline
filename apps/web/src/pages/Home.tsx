import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import Sidebar from '../components/Sidebar'

const promptItems = [
  { key: 'user-story', label: 'User Story Mapping', prompt: 'I want to create a user story for...' },
  { key: 'tech-spec', label: 'Technical Specification', prompt: 'I need a technical specification for...' },
  { key: 'sla', label: 'SLA Definition', prompt: 'Define SLA requirements for...' },
]

export default function Home() {
  const navigate = useNavigate()
  const [input, setInput] = useState('')
  const [creating, setCreating] = useState(false)
  const [convCollapsed, setConvCollapsed] = useState(false)

  // 创建新会话
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
        {/* Welcome Header */}
        <div className="pt-12 pb-6 px-12 text-center">
          <h1 className="text-4xl font-extrabold tracking-tight text-on-surface mb-2">Hello, Designer</h1>
          <p className="text-on-surface-variant font-medium text-lg">Define your new enterprise requirements through guided dialogue.</p>
        </div>

        {/* Chat Conversation Container */}
        <div className="flex-1 overflow-y-auto px-6 md:px-24 py-8 flex flex-col gap-8 scroll-smooth">
          {/* AI Message */}
          <div className="flex items-start gap-4 max-w-[80%]">
            <div className="w-8 h-8 rounded-lg bg-surface-container-high flex-shrink-0 flex items-center justify-center">
              <span className="material-symbols-outlined text-primary text-sm" style={{ fontVariationSettings: "'FILL' 1" }}>auto_awesome</span>
            </div>
            <div className="bg-surface-container-lowest p-5 rounded-2xl rounded-tl-none shadow-sm border border-outline-variant/10">
              <p className="text-sm leading-relaxed text-on-surface">Welcome to the Requirement Architect. I'm ready to help you structure your project. What type of requirement are we looking at today?</p>
              <div className="flex flex-wrap gap-2 mt-4">
                {promptItems.map(item => (
                  <button
                    key={item.key}
                    onClick={() => handlePromptClick(item.prompt)}
                    disabled={creating}
                    className="px-4 py-2 bg-surface-container-low hover:bg-surface-container-high text-primary text-xs font-bold rounded-full transition-all border border-primary/5 disabled:opacity-50"
                  >
                    {item.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Input Area */}
        <div className="px-6 md:px-24 pb-6">
          <div className="bg-surface-container-lowest rounded-2xl border border-outline-variant p-2 flex items-center gap-2">
            <button type="button" className="w-8 h-8 rounded-full border-0 bg-transparent text-on-surface-variant cursor-pointer flex items-center justify-center hover:bg-surface-variant/50">
              <span className="material-symbols-outlined">attach_file</span>
            </button>
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSend()}
              placeholder="Describe your requirement..."
              disabled={creating}
              className="flex-1 bg-transparent border-0 outline-none text-on-surface placeholder:text-on-surface/40 disabled:opacity-50"
            />
            <button
              type="button"
              onClick={handleSend}
              disabled={!input.trim() || creating}
              className="w-8 h-8 rounded-full border-0 bg-primary text-white cursor-pointer flex items-center justify-center hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span className="material-symbols-outlined text-sm">{creating ? 'progress_activity' : 'send'}</span>
            </button>
          </div>
          <div className="text-center text-xs text-on-surface/40 mt-2">AetherFlow AI can make mistakes. Verify critical project details.</div>
        </div>
      </main>
    </div>
  )
}
