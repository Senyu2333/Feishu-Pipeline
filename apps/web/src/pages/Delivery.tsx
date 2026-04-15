import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'
import {
  Card,
  Button,
  Progress,
  Avatar,
  Space,
  Tag,
} from 'antd'
import {
  ShareAltOutlined,
  FileTextOutlined,
  FilterOutlined,
  DownloadOutlined,
} from '@ant-design/icons'

const miniBars = [35, 28, 45, 38, 52, 48]

const testCases = [
  { id: 'TC-4021', title: 'User Authentication flow via OAuth2', meta: 'Auth-Module / Security / Production-Mirror', type: 'Integration', status: 'passed' },
  { id: 'TC-4022', title: 'Real-time Data Streaming Latency Test', meta: 'Core-Engine / Performance / APAC-Region', type: 'Performance', status: 'passed' },
  { id: 'TC-4023', title: 'Batch Import CSV File Validation', meta: 'File-System / Functional / Regression', type: 'Regression', status: 'failed' },
  { id: 'TC-4024', title: 'UI Dashboard Responsive Grid Layout', meta: 'Frontend / Layout / Cross-Browser', type: 'Visual', status: 'passed' },
]

const criticalIssues = [
  { level: 'P0 - BLOCKER', levelColor: '#ff4d4f', bg: '#fff2f0', border: '#ffccc7', time: '2h ago', title: 'CSV Parser Timeout', desc: 'The engine fails to process files over 50MB in production-mirror...', assignee: 'Alex Chen', avatar: 'A' },
  { level: 'P1 - CRITICAL', levelColor: '#faad14', bg: '#fffbe6', border: '#ffe58f', time: '5h ago', title: 'Token Expiration Flaw', desc: 'JWT refresh logic occasionally triggers double-logout when latency...', assignee: 'Sarah J.', avatar: 'S' },
]

export default function Delivery() {
  return (
    <div className="min-h-screen bg-background">
      <TopNav />
      <Sidebar />
      <main className="ml-64 mt-14 h-[calc(100vh-3.5rem)] overflow-y-auto p-6">
        <div className="flex justify-between items-start mb-5">
          <div>
            <h1 className="text-2xl font-bold text-on-surface m-0 mb-1">Automated Test Report</h1>
            <p className="text-sm text-on-surface-variant m-0">Sprint 24 Analysis — Version 2.4.0-build.88</p>
          </div>
          <Space>
            <Button icon={<ShareAltOutlined />} size="large" className="!rounded-lg">Share Report</Button>
            <Button type="primary" icon={<FileTextOutlined />} size="large" className="!rounded-lg">Export to Feishu Docs</Button>
          </Space>
        </div>

        <div className="grid grid-cols-3 gap-4 mb-5">
          <Card className="!rounded-xl !shadow-sm">
            <div className="flex justify-between items-start mb-4">
              <span className="text-xs font-bold text-on-surface-variant tracking-wider">OVERALL PASS RATE</span>
              <Tag className="!bg-green-50 !text-green-600 !border-0 !text-xs !font-semibold">+2.4% VS LAST RUN</Tag>
            </div>
            <div className="flex items-end gap-4 mb-3">
              <span className="text-3xl font-bold text-on-surface">98.2%</span>
              <div className="flex items-end gap-1 flex-1 h-10">
                {miniBars.map((h, i) => (
                  <div key={i} className={`w-3 rounded-t ${i >= 4 ? 'bg-green-500' : 'bg-green-200'}`} style={{ height: `${h}%` }} />
                ))}
              </div>
            </div>
            <div className="text-xs text-on-surface-variant">1,240 cases passed out of 1,263 total executions.</div>
          </Card>

          <Card className="!rounded-xl !shadow-sm">
            <div className="flex items-center gap-3 mb-3">
              <div className="w-10 h-10 rounded-lg bg-surface-container-high flex items-center justify-center">
                <div className="w-4 h-4 rounded-full border-2 border-primary" />
              </div>
            </div>
            <div className="text-xs font-bold text-on-surface-variant tracking-wider mb-1">RUN DURATION</div>
            <div className="text-2xl font-bold text-on-surface mb-1">14m 22s</div>
            <div className="text-xs text-on-surface-variant">Average: 15m 10s</div>
          </Card>

          <Card className="!rounded-xl !shadow-sm">
            <div className="flex items-center gap-3 mb-3">
              <div className="w-10 h-10 rounded-lg bg-red-50 flex items-center justify-center">
                <div className="w-4 h-4 rounded-full border-2 border-red-500" />
              </div>
            </div>
            <div className="text-xs font-bold text-on-surface-variant tracking-wider mb-1">FAILURES</div>
            <div className="flex items-center gap-3 mb-1">
              <span className="text-2xl font-bold text-on-surface">23</span>
              <div className="flex-1 h-2 bg-surface-variant rounded-full overflow-hidden">
                <div className="h-full bg-red-500 rounded-full" style={{ width: '23%' }} />
              </div>
            </div>
          </Card>
        </div>

        <div className="grid grid-cols-[1fr_340px] gap-5">
          <Card className="!rounded-xl !shadow-sm">
            <div className="flex justify-between items-center mb-4">
              <span className="font-semibold text-on-surface">Test Case Executions</span>
              <Space>
                <Button type="text" icon={<FilterOutlined />} className="!text-on-surface-variant" />
                <Button type="text" icon={<DownloadOutlined />} className="!text-on-surface-variant" />
              </Space>
            </div>
            <div className="rounded-lg border border-outline-variant overflow-hidden">
              <div className="flex bg-surface-container-low text-xs font-semibold text-on-surface-variant tracking-wider">
                <div className="w-24 p-3">CASE ID</div>
                <div className="flex-1 p-3">TEST TITLE</div>
                <div className="w-28 p-3">TYPE</div>
                <div className="w-24 p-3">STATUS</div>
              </div>
              {testCases.map((tc) => (
                <div key={tc.id} className={`flex items-center border-t border-outline-variant ${tc.status === 'passed' ? 'bg-white' : 'bg-red-50'}`}>
                  <div className="w-24 p-3 text-xs font-semibold text-on-surface">{tc.id}</div>
                  <div className="flex-1 p-3">
                    <div className="text-sm font-medium text-on-surface">{tc.title}</div>
                    <div className="text-xs text-on-surface-variant mt-0.5">{tc.meta}</div>
                  </div>
                  <div className="w-28 p-3">
                    <Tag className="!border-0">{tc.type}</Tag>
                  </div>
                  <div className="w-24 p-3">
                    <span className={`inline-flex items-center gap-1 text-xs font-semibold ${tc.status === 'passed' ? 'text-green-600' : 'text-red-600'}`}>
                      <span className={`w-2 h-2 rounded-full ${tc.status === 'passed' ? 'bg-green-500' : 'bg-red-500'}`} />
                      {tc.status === 'passed' ? 'PASSED' : 'FAILED'}
                    </span>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4 pt-3 border-t border-outline-variant">
              <Button type="link" className="!text-primary !p-0">View All 1,263 Test Cases</Button>
            </div>
          </Card>

          <aside className="flex flex-col gap-4">
            <Card className="!rounded-xl !shadow-sm">
              <div className="flex items-center gap-2 mb-4">
                <span className="material-symbols-outlined !text-red-500 text-base">bug_report</span>
                <span className="font-semibold text-on-surface">Critical Issues</span>
              </div>
              <div className="flex flex-col gap-3 mb-4">
                {criticalIssues.map((issue, idx) => (
                  <div key={idx} className="p-3 rounded-lg border" style={{ background: issue.bg, borderColor: issue.border }}>
                    <div className="flex justify-between items-start mb-2">
                      <span className="text-xs font-bold" style={{ color: issue.levelColor }}>{issue.level}</span>
                      <span className="text-xs text-on-surface-variant">{issue.time}</span>
                    </div>
                    <div className="font-semibold text-sm text-on-surface mb-1">{issue.title}</div>
                    <div className="text-xs text-on-surface-variant mb-2">{issue.desc}</div>
                    <div className="flex items-center gap-2">
                      <Avatar size="small" className="!bg-gray-800 text-white">{issue.avatar}</Avatar>
                      <span className="text-xs text-on-surface-variant">Assigned to: {issue.assignee}</span>
                    </div>
                  </div>
                ))}
              </div>
              <Button block className="!bg-surface-container-low !text-on-surface !border-0 !h-10 !rounded-lg">Manage All Issues (12)</Button>
            </Card>

            <Card className="!rounded-xl !shadow-sm !bg-gray-800 text-white">
              <div className="text-xs font-bold text-on-surface-variant tracking-wider mb-4">ENV: PRODUCTION MIRROR</div>
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between text-xs text-white/70 mb-2">
                    <span>CPU Load</span><span>24%</span>
                  </div>
                  <Progress percent={24} strokeColor="#fff" trailColor="rgba(255,255,255,0.25)" showInfo={false} />
                </div>
                <div>
                  <div className="flex justify-between text-xs text-white/70 mb-2">
                    <span>Memory Usage</span><span>6.2GB / 16GB</span>
                  </div>
                  <Progress percent={39} strokeColor="#fff" trailColor="rgba(255,255,255,0.25)" showInfo={false} />
                </div>
              </div>
            </Card>
          </aside>
        </div>
      </main>
    </div>
  )
}
