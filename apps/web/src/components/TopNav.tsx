import { useState, useEffect } from 'react'
import { Badge, Avatar, Space, Button, Input, Dropdown, message } from 'antd'
import {
  BellOutlined,
  QuestionCircleOutlined,
  SettingOutlined,
  SearchOutlined,
  UserOutlined,
  LoginOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'

interface UserInfo {
  id: string
  name: string
  avatar?: string
  role: string
}

const API_BASE = '/api'

export default function TopNav({ showSearch }: { showSearch?: boolean }) {
  const [user, setUser] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(true)

  // 检查登录状态
  useEffect(() => {
    checkAuth()
  }, [])

  const checkAuth = async () => {
    try {
      const res = await fetch(`${API_BASE}/me`)
      if (res.ok) {
        const data = await res.json()
        setUser(data)
      } else if (res.status === 401 || res.status === 403) {
        // 未登录或无权限
        setUser(null)
      }
    } catch (e) {
      // 网络错误，未登录
    } finally {
      setLoading(false)
    }
  }

  const handleLogin = () => {
    // 跳转到飞书授权登录
    window.location.href = `${API_BASE}/auth/feishu/login`
  }

  // 用户下拉菜单（只显示信息，无登出）
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'name',
      label: <span style={{ fontWeight: 600 }}>{user?.name || '用户'}</span>,
      disabled: true,
    },
    {
      key: 'role',
      label: <span style={{ color: '#8b95a8', fontSize: 12 }}>{user?.role || '访客'}</span>,
      disabled: true,
    },
    { type: 'divider' },
    {
      key: 'profile',
      label: '个人设置',
      icon: <UserOutlined />,
    },
  ]

  const getInitials = (name: string) => {
    return name.slice(0, 2).toUpperCase()
  }

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
          
          {loading ? (
            <Avatar size="small" style={{ background: '#d9d9d9' }}>
              <span style={{ fontSize: 10 }}>...</span>
            </Avatar>
          ) : user ? (
            <Dropdown menu={{ items: userMenuItems }} trigger={['click']} placement="bottomRight">
              <Avatar 
                size="small" 
                style={{ background: '#0066ff', cursor: 'pointer' }}
                src={user.avatar}
              >
                {user.avatar ? null : getInitials(user.name)}
              </Avatar>
            </Dropdown>
          ) : (
            <Button 
              type="primary" 
              size="small" 
              icon={<LoginOutlined />}
              onClick={handleLogin}
              style={{ borderRadius: 6 }}
            >
              登录
            </Button>
          )}
        </Space>
      </div>
    </header>
  )
}
