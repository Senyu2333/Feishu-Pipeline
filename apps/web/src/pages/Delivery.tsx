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
  BugOutlined,
} from '@ant-design/icons'

const miniBars = [35, 28, 45, 38, 52, 48]

const testCases = [
  {
    id: 'TC-4021',
    title: 'User Authentication flow via OAuth2',
    meta: 'Auth-Module / Security / Production-Mirror',
    type: 'Integration',
    status: 'passed',
  },
  {
    id: 'TC-4022',
    title: 'Real-time Data Streaming Latency Test',
    meta: 'Core-Engine / Performance / APAC-Region',
    type: 'Performance',
    status: 'passed',
  },
  {
    id: 'TC-4023',
    title: 'Batch Import CSV File Validation',
    meta: 'File-System / Functional / Regression',
    type: 'Regression',
    status: 'failed',
  },
  {
    id: 'TC-4024',
    title: 'UI Dashboard Responsive Grid Layout',
    meta: 'Frontend / Layout / Cross-Browser',
    type: 'Visual',
    status: 'passed',
  },
]

const criticalIssues = [
  {
    level: 'P0 - BLOCKER',
    levelColor: '#ff4d4f',
    bg: '#fff2f0',
    border: '#ffccc7',
    time: '2h ago',
    title: 'CSV Parser Timeout',
    desc: 'The engine fails to process files over 50MB in production-mirror...',
    assignee: 'Alex Chen',
    avatar: 'A',
  },
  {
    level: 'P1 - CRITICAL',
    levelColor: '#faad14',
    bg: '#fffbe6',
    border: '#ffe58f',
    time: '5h ago',
    title: 'Token Expiration Flaw',
    desc: 'JWT refresh logic occasionally triggers double-logout when latency...',
    assignee: 'Sarah J.',
    avatar: 'S',
  },
]

export default function Delivery() {
  return (
    <div className="app-container">
      <TopNav />
      <div className="main-layout">
        <Sidebar />
        <main className="page-content delivery-page">
          {/* Header */}
          <div className="delivery-header">
            <div>
              <h1 className="delivery-title">Automated Test Report</h1>
              <p className="delivery-subtitle">Sprint 24 Analysis — Version 2.4.0-build.88</p>
            </div>
            <Space>
              <Button icon={<ShareAltOutlined />} size="large">
                Share Report
              </Button>
              <Button type="primary" icon={<FileTextOutlined />} size="large">
                Export to Feishu Docs
              </Button>
            </Space>
          </div>

          {/* Stats Row */}
          <div className="delivery-stats">
            {/* Pass Rate */}
            <Card className="delivery-stat-card pass-rate-card">
              <div className="pass-rate-header">
                <span className="pass-rate-label">OVERALL PASS RATE</span>
                <Tag className="pass-rate-tag">+2.4% VS LAST RUN</Tag>
              </div>
              <div className="pass-rate-body">
                <div className="pass-rate-value">98.2%</div>
                <div className="pass-rate-bars">
                  {miniBars.map((h, i) => (
                    <div
                      key={i}
                      className="pass-rate-bar"
                      style={{ height: `${h}%`, opacity: i >= 4 ? 1 : 0.35 }}
                    />
                  ))}
                </div>
              </div>
              <div className="pass-rate-desc">1,240 cases passed out of 1,263 total executions.</div>
            </Card>

            {/* Run Duration */}
            <Card className="delivery-stat-card duration-card">
              <div className="stat-icon" style={{ background: '#e8f2fc', color: '#0066ff' }}>
                <div className="stat-icon-inner" style={{ borderColor: '#0066ff' }} />
              </div>
              <div className="stat-label">RUN DURATION</div>
              <div className="stat-value">14m 22s</div>
              <div className="stat-desc">Average: 15m 10s</div>
            </Card>

            {/* Failures */}
            <Card className="delivery-stat-card failures-card">
              <div className="stat-icon" style={{ background: '#fff2f0', color: '#ff4d4f' }}>
                <div className="stat-icon-inner" style={{ borderColor: '#ff4d4f' }} />
              </div>
              <div className="stat-label">FAILURES</div>
              <div className="stat-value-row">
                <span className="stat-value" style={{ color: '#1a1a2e' }}>23</span>
                <div className="failures-bar">
                  <div className="failures-fill" style={{ width: '23%' }} />
                </div>
              </div>
            </Card>
          </div>

          {/* Main Content */}
          <div className="delivery-layout">
            {/* Test Cases Table */}
            <Card className="delivery-table-card">
              <div className="delivery-table-header">
                <span className="delivery-table-title">Test Case Executions</span>
                <Space>
                  <Button type="text" icon={<FilterOutlined />} />
                  <Button type="text" icon={<DownloadOutlined />} />
                </Space>
              </div>

              <div className="test-table">
                <div className="test-table-head">
                  <div className="test-th test-th-id">CASE ID</div>
                  <div className="test-th test-th-title">TEST TITLE</div>
                  <div className="test-th test-th-type">TYPE</div>
                  <div className="test-th test-th-status">STATUS</div>
                </div>
                {testCases.map((tc) => (
                  <div key={tc.id} className={`test-table-row ${tc.status}`}>
                    <div className="test-td test-td-id">{tc.id}</div>
                    <div className="test-td test-td-title">
                      <div className="test-title">{tc.title}</div>
                      <div className="test-meta">{tc.meta}</div>
                    </div>
                    <div className="test-td test-td-type">
                      <Tag className="test-type-tag">{tc.type}</Tag>
                    </div>
                    <div className="test-td test-td-status">
                      {tc.status === 'passed' ? (
                        <span className="test-status passed">
                          <span className="test-status-dot" /> PASSED
                        </span>
                      ) : (
                        <span className="test-status failed">
                          <span className="test-status-dot" /> FAILED
                        </span>
                      )}
                    </div>
                  </div>
                ))}
              </div>

              <div className="delivery-table-footer">
                <Button type="link">View All 1,263 Test Cases</Button>
              </div>
            </Card>

            {/* Right Sidebar */}
            <aside className="delivery-sidebar">
              {/* Critical Issues */}
              <div className="delivery-sidebar-section">
                <div className="delivery-sidebar-title">
                  <BugOutlined style={{ color: '#ff4d4f', fontSize: 16 }} />
                  Critical Issues
                </div>
                <div className="issues-list">
                  {criticalIssues.map((issue, idx) => (
                    <Card
                      key={idx}
                      className="issue-card"
                      style={{ background: issue.bg, borderColor: issue.border }}
                    >
                      <div className="issue-header">
                        <span
                          className="issue-level"
                          style={{ color: issue.levelColor }}
                        >
                          {issue.level}
                        </span>
                        <span className="issue-time">{issue.time}</span>
                      </div>
                      <div className="issue-title">{issue.title}</div>
                      <div className="issue-desc">{issue.desc}</div>
                      <div className="issue-assignee">
                        <Avatar size="small" style={{ background: '#1a1a2e' }}>
                          {issue.avatar}
                        </Avatar>
                        <span>Assigned to: {issue.assignee}</span>
                      </div>
                    </Card>
                  ))}
                </div>
                <Button block className="manage-issues-btn">
                  Manage All Issues (12)
                </Button>
              </div>

              {/* Env Card */}
              <Card className="env-card">
                <div className="env-title">ENV: PRODUCTION MIRROR</div>
                <div className="env-metric">
                  <div className="env-metric-label">
                    <span>CPU Load</span>
                    <span>24%</span>
                  </div>
                  <Progress percent={24} strokeColor="#fff" trailColor="rgba(255,255,255,0.25)" showInfo={false} />
                </div>
                <div className="env-metric">
                  <div className="env-metric-label">
                    <span>Memory Usage</span>
                    <span>6.2GB / 16GB</span>
                  </div>
                  <Progress percent={39} strokeColor="#fff" trailColor="rgba(255,255,255,0.25)" showInfo={false} />
                </div>
              </Card>
            </aside>
          </div>
        </main>
      </div>
    </div>
  )
}
