import { Bubble, Prompts, Sender } from '@ant-design/x'
import { Card, Badge, Avatar } from 'antd'
import {
  StarOutlined,
  FileTextOutlined,
  SafetyCertificateOutlined,
  PaperClipOutlined,
  AudioOutlined,
  SendOutlined,
} from '@ant-design/icons'
import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'

const aiAvatar = (
  <Avatar size="small" style={{ background: '#dbe8f6', color: '#0066ff' }}>
    <StarOutlined />
  </Avatar>
)

const userAvatar = (
  <Avatar size="small" style={{ background: '#0066ff' }}>JD</Avatar>
)

const promptItems = [
  { key: 'user-story', label: 'User Story Mapping', icon: <FileTextOutlined /> },
  { key: 'tech-spec', label: 'Technical Specification', icon: <FileTextOutlined /> },
  { key: 'sla', label: 'SLA Definition', icon: <SafetyCertificateOutlined /> },
]

export default function Home() {
  return (
    <div className="app-container">
      <TopNav />
      <div className="main-layout">
        <Sidebar />
        <main className="chat-area">
          <div className="chat-header">
            <h1>Hello, Designer</h1>
            <p>Define your new enterprise requirements through guided dialogue.</p>
          </div>
          <div className="chat-messages">
            <Bubble
              placement="start"
              avatar={aiAvatar}
              content="Welcome to the Requirement Architect. I'm ready to help you structure your project. What type of requirement are we looking at today?"
              styles={{ content: { background: '#fff', borderRadius: 14, borderBottomLeftRadius: 4, boxShadow: '0 1px 2px rgba(0,0,0,0.04)' } }}
            />
            <Prompts
              items={promptItems}
              onItemClick={(info: { data: { key: string } }) => console.log(info.data.key)}
              style={{ marginLeft: 40 }}
            />
            <Bubble
              placement="end"
              avatar={userAvatar}
              content="I want to create a new User Story for the authentication flow of the Alpha App. It needs to include biometric support."
              styles={{ content: { background: '#0066ff', color: '#fff', borderRadius: 14, borderBottomRightRadius: 4 } }}
            />
            <Bubble
              placement="start"
              avatar={aiAvatar}
              content={
                <div>
                  <p style={{ margin: 0, marginBottom: 12 }}>
                    Understood. I've initialized a draft for <strong>Alpha App: Biometric Auth Flow</strong>. Here is a summary of the core logic I'm drafting based on your request:
                  </p>
                  <Card
                    size="small"
                    style={{ maxWidth: 420, background: 'linear-gradient(135deg, #f0f7ff 0%, #e8f2fc 100%)', borderColor: '#dbe8f6' }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600 }}>
                        <Avatar size="small" style={{ background: '#0066ff' }} icon={<StarOutlined />} />
                        Draft: Requirement #1042
                      </div>
                      <Badge style={{ background: '#dbe8f6', color: '#0066ff' }}>IN PROGRESS</Badge>
                    </div>
                    <div style={{ display: 'flex', gap: 32, marginBottom: 12 }}>
                      <div>
                        <div style={{ fontSize: 10, fontWeight: 600, color: '#8b95a8', marginBottom: 4 }}>ACTOR</div>
                        <div style={{ fontSize: 12, fontWeight: 500 }}>End-User (Mobile)</div>
                      </div>
                      <div>
                        <div style={{ fontSize: 10, fontWeight: 600, color: '#8b95a8', marginBottom: 4 }}>COMPLEXITY</div>
                        <div style={{ width: 60, height: 4, background: '#dbe4ee', borderRadius: 2, overflow: 'hidden' }}>
                          <div style={{ width: '60%', height: '100%', background: '#0066ff', borderRadius: 2 }} />
                        </div>
                      </div>
                    </div>
                    <blockquote style={{ margin: 0, paddingLeft: 12, borderLeft: '2px solid #0066ff', fontStyle: 'italic', color: '#5a6478', fontSize: 13 }}>
                      "As a user, I want to use FaceID or TouchID so that I can log in securely without typing my password every time."
                    </blockquote>
                  </Card>
                </div>
              }
              styles={{ content: { background: 'transparent', padding: 0, boxShadow: 'none' } }}
            />
          </div>
          <div className="chat-input-wrapper">
            <Sender
              placeholder="Tell me more about the requirement logic or upload a sketch..."
              footer={(
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: 8, padding: '8px 12px' }}>
                  <button type="button" style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: 'transparent', color: '#5a6478', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><PaperClipOutlined /></button>
                  <button type="button" style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: 'transparent', color: '#5a6478', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><AudioOutlined /></button>
                  <button type="button" style={{ width: 32, height: 32, borderRadius: '50%', border: 'none', background: '#0066ff', color: '#fff', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><SendOutlined /></button>
                </div>
              )}
            />
            <div className="chat-footer-note">AetherFlow AI can make mistakes. Verify critical project details.</div>
          </div>
        </main>
      </div>
    </div>
  )
}
