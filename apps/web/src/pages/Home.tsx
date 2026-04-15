import { Card, Badge, Avatar } from 'antd'
import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'

const promptItems = [
  { key: 'user-story', label: 'User Story Mapping' },
  { key: 'tech-spec', label: 'Technical Specification' },
  { key: 'sla', label: 'SLA Definition' },
]

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <TopNav />
      <Sidebar />
      <main className="ml-64 mt-14 h-[calc(100vh-3.5rem)] flex flex-col relative overflow-hidden">
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
                  <button key={item.key} className="px-4 py-2 bg-surface-container-low hover:bg-surface-container-high text-primary text-xs font-bold rounded-full transition-all border border-primary/5">
                    {item.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* User Message */}
          <div className="flex items-start gap-4 max-w-[80%] self-end flex-row-reverse">
            <div className="w-8 h-8 rounded-lg bg-primary flex-shrink-0 flex items-center justify-center shadow-md">
              <span className="material-symbols-outlined text-white text-sm">person</span>
            </div>
            <div className="bg-primary text-white p-5 rounded-2xl rounded-tr-none shadow-md">
              <p className="text-sm leading-relaxed">I want to create a new User Story for the authentication flow of the Alpha App. It needs to include biometric support.</p>
            </div>
          </div>

          {/* AI Message with Card */}
          <div className="flex items-start gap-4 max-w-[85%]">
            <div className="w-8 h-8 rounded-lg bg-surface-container-high flex-shrink-0 flex items-center justify-center">
              <span className="material-symbols-outlined text-primary text-sm" style={{ fontVariationSettings: "'FILL' 1" }}>auto_awesome</span>
            </div>
            <div className="flex flex-col gap-4 w-full">
              <div className="bg-surface-container-lowest p-5 rounded-2xl rounded-tl-none shadow-sm border border-outline-variant/10">
                <p className="text-sm leading-relaxed text-on-surface">Understood. I've initialized a draft for <strong>Alpha App: Biometric Auth Flow</strong>. Here is a summary of the core logic I'm drafting based on your request:</p>
              </div>
              <Card
                size="small"
                className="!rounded-2xl !border-0 !shadow-sm max-w-[420px]"
                style={{ background: 'linear-gradient(135deg, #f0f7ff 0%, #e8f2fc 100%)' }}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2 font-semibold">
                    <Avatar size="small" className="!bg-primary" icon={<span className="material-symbols-outlined text-white text-xs">auto_awesome</span>} />
                    Draft: Requirement #1042
                  </div>
                  <Badge className="!bg-surface-container-high !text-primary">IN PROGRESS</Badge>
                </div>
                <div className="flex gap-8 mb-3">
                  <div>
                    <div className="text-xs font-semibold text-gray-400 mb-1">ACTOR</div>
                    <div className="text-sm font-medium">End-User (Mobile)</div>
                  </div>
                  <div>
                    <div className="text-xs font-semibold text-gray-400 mb-1">COMPLEXITY</div>
                    <div className="w-12 h-1 bg-gray-200 rounded overflow-hidden">
                      <div className="w-3/5 h-full bg-primary rounded" />
                    </div>
                  </div>
                </div>
                <blockquote className="m-0 pl-3 border-l-2 text-gray-500 italic text-sm">
                  "As a user, I want to use FaceID or TouchID so that I can log in securely without typing my password every time."
                </blockquote>
              </Card>
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
              placeholder="Describe your requirement..."
              className="flex-1 bg-transparent border-0 outline-none text-on-surface placeholder:text-on-surface/40"
            />
            <button type="button" className="w-8 h-8 rounded-full border-0 bg-primary text-white cursor-pointer flex items-center justify-center hover:opacity-90">
              <span className="material-symbols-outlined text-sm">send</span>
            </button>
          </div>
          <div className="text-center text-xs text-on-surface/40 mt-2">AetherFlow AI can make mistakes. Verify critical project details.</div>
        </div>
      </main>
    </div>
  )
}
