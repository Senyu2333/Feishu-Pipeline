import { Link, useLocation } from '@tanstack/react-router'
import { Button } from 'antd'
import {
  HomeOutlined,
  PlusSquareOutlined,
  ApartmentOutlined,
  LineChartOutlined,
  CheckSquareOutlined,
  InboxOutlined,
  HddOutlined,
  SettingOutlined,
  CloudOutlined,
} from '@ant-design/icons'

const menuItems = [
  { icon: HomeOutlined, label: 'Home', to: '/' },
  { icon: PlusSquareOutlined, label: 'Creation', to: '/new-requirement' },
  { icon: ApartmentOutlined, label: 'Workflows', to: '/workflows' },
  { icon: LineChartOutlined, label: 'Monitoring', to: '/monitoring' },
  { icon: CheckSquareOutlined, label: 'Approvals', to: '/approvals' },
  { icon: InboxOutlined, label: 'Delivery', to: '/delivery' },
]

export default function Sidebar() {
  const location = useLocation()
  const pathname = location.pathname

  return (
    <aside className="sidebar">
      <div className="sidebar-top">
        <div className="project-card">
          <div className="project-icon">
            <CloudOutlined style={{ fontSize: 18, color: '#fff' }} />
          </div>
          <div className="project-info">
            <div className="project-name">Project Alpha</div>
            <div className="project-type">Enterprise AI</div>
          </div>
        </div>

        <Link to="/new-requirement">
          <Button type="primary" icon={<PlusSquareOutlined />} block className="new-req-btn">
            New Requirement
          </Button>
        </Link>

        <nav className="sidebar-nav">
          {menuItems.map((item) => {
            const isActive = pathname === item.to || (item.to !== '/' && pathname.startsWith(item.to))
            const Icon = item.icon
            return (
              <Link
                key={item.label}
                to={item.to}
                className={`sidebar-nav-item ${isActive ? 'active' : ''}`}
              >
                <Icon style={{ fontSize: 16 }} />
                <span>{item.label}</span>
              </Link>
            )
          })}
        </nav>
      </div>

      <div className="sidebar-bottom">
        <a href="#" className="sidebar-nav-item">
          <HddOutlined style={{ fontSize: 16 }} />
          <span>Archive</span>
        </a>
        <a href="#" className="sidebar-nav-item">
          <SettingOutlined style={{ fontSize: 16 }} />
          <span>Settings</span>
        </a>
      </div>
    </aside>
  )
}
