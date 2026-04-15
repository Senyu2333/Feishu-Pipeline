import { Badge, Avatar, Space, Button, Input } from 'antd'
import {
  BellOutlined,
  QuestionCircleOutlined,
  SettingOutlined,
  SearchOutlined,
} from '@ant-design/icons'

export default function TopNav({ showSearch }: { showSearch?: boolean }) {
  return (
    <header className="top-nav">
      <div className="top-nav-left">
        <div className="logo">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="10" stroke="#0066ff" strokeWidth="2"/>
            <path d="M12 6v6l4 2" stroke="#0066ff" strokeWidth="2" strokeLinecap="round"/>
          </svg>
          <span className="logo-text">AetherFlow AI</span>
        </div>
        <nav className="top-nav-links">
          <a href="#" className="nav-link">Documents</a>
          <a href="#" className="nav-link active">Workspaces</a>
          <a href="#" className="nav-link">Templates</a>
        </nav>
      </div>
      <div className="top-nav-right">
        {showSearch && (
          <Input
            prefix={<SearchOutlined />}
            placeholder="Search workflows..."
            className="top-search"
            style={{ width: 200, marginRight: 8 }}
          />
        )}
        <Space>
          <Badge dot>
            <Button type="text" icon={<BellOutlined />} className="icon-btn" />
          </Badge>
          <Button type="text" icon={<QuestionCircleOutlined />} className="icon-btn" />
          <Button type="text" icon={<SettingOutlined />} className="icon-btn" />
          <Avatar size="small" style={{ background: '#0066ff' }}>JD</Avatar>
        </Space>
      </div>
    </header>
  )
}
