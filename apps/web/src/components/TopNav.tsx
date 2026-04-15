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
    <header className="fixed top-0 w-full flex justify-between items-center px-6 h-14 bg-[#f4faff]/80 backdrop-blur-xl z-50 border-b border-[#c1c6d7]/20">
      <div className="flex items-center gap-8">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center shadow-md">
            <span className="material-symbols-filled text-white text-sm">cloud_done</span>
          </div>
          <span className="text-xl font-bold tracking-tight text-[#001f2a]">AetherFlow AI</span>
        </div>
        <nav className="hidden md:flex items-center gap-6">
          <a href="#" className="text-[#001f2a]/60 hover:text-primary text-sm font-medium transition-colors">Documents</a>
          <a href="#" className="text-[#001f2a]/60 hover:text-primary text-sm font-medium transition-colors">Workspaces</a>
          <a href="#" className="text-[#001f2a]/60 hover:text-primary text-sm font-medium transition-colors">Templates</a>
        </nav>
      </div>
      <div className="flex items-center gap-3">
        {showSearch && (
          <Input
            prefix={<SearchOutlined className="text-on-surface-variant" />}
            placeholder="Search workflows..."
            className="!w-52 !bg-white/50 !border-0 !rounded-lg hover:!bg-white focus-within:!bg-white"
          />
        )}
        <Space>
          <button className="p-2 hover:bg-[#c9e7f7]/30 rounded-full transition-colors active:scale-95 duration-150">
            <span className="material-symbols-outlined text-on-surface-variant">notifications</span>
          </button>
          <button className="p-2 hover:bg-[#c9e7f7]/30 rounded-full transition-colors active:scale-95 duration-150">
            <span className="material-symbols-outlined text-on-surface-variant">help</span>
          </button>
          <button className="p-2 hover:bg-[#c9e7f7]/30 rounded-full transition-colors active:scale-95 duration-150">
            <span className="material-symbols-outlined text-on-surface-variant">settings</span>
          </button>
          
          {loading ? (
            <Avatar size="small" className="!bg-surface-variant">...</Avatar>
          ) : user ? (
            <Dropdown menu={{ items: userMenuItems }} trigger={['click']} placement="bottomRight">
              <div className="ml-2 w-8 h-8 rounded-full bg-primary-container flex items-center justify-center text-white text-xs font-bold ring-2 ring-surface ring-offset-2 cursor-pointer">
                {getInitials(user.name)}
              </div>
            </Dropdown>
          ) : (
            <Button 
              type="primary" 
              size="small" 
              icon={<LoginOutlined />}
              onClick={handleLogin}
              className="!rounded-lg !font-semibold"
            >
              登录
            </Button>
          )}
        </Space>
      </div>
    </header>
  )
}
