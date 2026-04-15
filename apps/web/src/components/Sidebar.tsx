import { Link, useLocation } from '@tanstack/react-router'
import { Button } from 'antd'

const menuItems = [
  { icon: 'chat_bubble', label: 'Home', to: '/' },
  { icon: 'add_box', label: 'Creation', to: '/new-requirement' },
  { icon: 'account_tree', label: 'Workflows', to: '/workflows' },
  { icon: 'insights', label: 'Monitoring', to: '/monitoring' },
  { icon: 'fact_check', label: 'Approvals', to: '/approvals' },
  { icon: 'package_2', label: 'Delivery', to: '/delivery' },
]

export default function Sidebar() {
  const location = useLocation()
  const pathname = location.pathname

  return (
    <aside className="fixed left-0 top-14 h-[calc(100vh-3.5rem)] flex flex-col p-4 w-64 bg-[#e6f6ff] z-40">
      <div className="flex items-center gap-3 px-3 py-4 mb-4">
        <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center shadow-lg">
          <span className="material-symbols-filled text-white">cloud_done</span>
        </div>
        <div>
          <p className="text-sm font-black text-[#001f2a] leading-none">Project Alpha</p>
          <p className="text-[10px] uppercase tracking-widest text-[#001f2a]/50 font-bold mt-1">Enterprise AI</p>
        </div>
      </div>

      <Link to="/new-requirement">
        <button className="mb-6 mx-2 py-3 px-4 bg-gradient-to-r from-primary to-primary-container text-white rounded-xl font-bold text-sm shadow-md hover:shadow-lg transition-all active:scale-[0.98] w-[calc(100%-1rem)]">
          <span className="flex items-center justify-center gap-2">
            <span className="material-symbols-outlined text-sm">add</span>
            New Requirement
          </span>
        </button>
      </Link>

      <nav className="flex-1 flex flex-col gap-1 overflow-y-auto">
        {menuItems.map((item) => {
          const isActive = pathname === item.to || (item.to !== '/' && pathname.startsWith(item.to))
          return (
            <Link
              key={item.label}
              to={item.to}
              className={`flex items-center gap-3 px-3 py-2 rounded-lg transition-all ${isActive 
                ? 'bg-white text-primary font-bold border-l-4 border-primary translate-x-1' 
                : 'text-[#001f2a]/70 hover:bg-[#c9e7f7]'}`}
            >
              <span className="material-symbols-outlined">{item.icon}</span>
              <span className="font-medium text-sm">{item.label}</span>
            </Link>
          )
        })}
      </nav>

      <div className="mt-auto pt-4 border-t border-[#001f2a]/5 flex flex-col gap-1">
        <a href="#" className="flex items-center gap-3 text-[#001f2a]/70 px-3 py-2 hover:bg-[#c9e7f7] rounded-lg transition-all">
          <span className="material-symbols-outlined">inventory_2</span>
          <span className="font-medium text-sm">Archive</span>
        </a>
        <a href="#" className="flex items-center gap-3 text-[#001f2a]/70 px-3 py-2 hover:bg-[#c9e7f7] rounded-lg transition-all">
          <span className="material-symbols-outlined">settings</span>
          <span className="font-medium text-sm">Settings</span>
        </a>
      </div>
    </aside>
  )
}
