import { useState } from 'react'
import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'
import {
  Card,
  Progress,
  Tag,
  Button,
  Badge,
  Segmented,
  Avatar,
  Space,
  Divider,
  Tooltip,
} from 'antd'
import {
  PauseCircleOutlined,
  ReloadOutlined,
  CheckCircleFilled,
  LoadingOutlined,
  MoreOutlined,
  CloudUploadOutlined,
  FileTextOutlined,
  RobotOutlined,
  FormatPainterOutlined,
  WarningFilled,
  CheckCircleOutlined,
  CloudSyncOutlined,
} from '@ant-design/icons'

const pipelineSteps = [
  {
    key: 'ingestion',
    title: 'Data Ingestion',
    desc: 'Source: Feishu Cloud Workspace',
    duration: '0.2s',
    status: 'done',
    icon: <CloudUploadOutlined />,
  },
  {
    key: 'mapping',
    title: 'Semantic Mapping',
    desc: 'Vectorizing 42 nodes',
    duration: '1.4s',
    status: 'done',
    icon: <FileTextOutlined />,
  },
  {
    key: 'inference',
    title: 'Logical Inference',
    desc: '64% Completed',
    duration: '',
    status: 'running',
    icon: <RobotOutlined />,
    progress: 64,
  },
  {
    key: 'formatting',
    title: 'Output Formatting',
    desc: 'Waiting for inference...',
    duration: '--',
    status: 'pending',
    icon: <FormatPainterOutlined />,
  },
]

const consoleLogs = [
  { time: '14:22:01', level: 'INIT', text: 'Loading knowledge base v1.2', color: '#a0aec0' },
  { time: '14:22:05', level: 'INFO', text: 'Connection established to Feishu API', color: '#a0aec0' },
  { time: '14:22:08', level: 'INFO', text: 'Fetching document metadata...', color: '#a0aec0' },
  { time: '14:22:12', level: 'INFO', text: 'Analyzing paragraph headers (Node 1-12)', color: '#a0aec0' },
  { time: '14:22:45', level: 'INFO', text: 'Generation started using model: Aether-Large', color: '#a0aec0' },
  { time: '14:23:01', level: 'WARN', text: 'Latency spike detected in Region 4', color: '#f6ad55' },
]

const metricsBars = [
  28, 32, 45, 38, 52, 48, 55, 50, 58, 62, 54, 68, 64, 58, 48, 42, 55, 60, 52, 48, 72,
]

export default function Monitoring() {
  const [metricTab, setMetricTab] = useState('CPU')

  return (
    <div className="app-container">
      <TopNav />
      <div className="main-layout">
        <Sidebar />
        <main className="page-content monitoring-page">
          {/* Header */}
          <div className="monitoring-header">
            <div>
              <div className="monitoring-breadcrumb">
                Agents <span className="breadcrumb-sep">›</span> Content Optimizer v2.4
              </div>
              <h1 className="monitoring-title">Agent Status Monitor</h1>
            </div>
            <Space>
              <Button icon={<PauseCircleOutlined />} size="large">
                Pause Execution
              </Button>
              <Button type="primary" icon={<ReloadOutlined />} size="large">
                Force Restart
              </Button>
            </Space>
          </div>

          {/* Top Grid */}
          <div className="monitoring-grid">
            {/* Agent Status Card */}
            <Card className="monitoring-card agent-status-card">
              <div className="agent-status-header">
                <div className="agent-status-identity">
                  <Avatar
                    size={48}
                    style={{ background: '#dbe8f6', color: '#0066ff' }}
                    icon={<RobotOutlined style={{ fontSize: 24 }} />}
                  />
                  <div>
                    <div className="agent-name">Data Synthesis Agent</div>
                    <div className="agent-meta">
                      Cluster ID: <Tag className="cluster-tag">AF-992-DELTA</Tag>
                    </div>
                  </div>
                </div>
                <div className="agent-status-badge">
                  <Badge status="success" text="ACTIVE RUNNING" />
                  <div className="agent-uptime">Uptime: 14h 22m 04s</div>
                </div>
              </div>

              <div className="agent-progress-section">
                <div className="agent-progress-label">
                  <span>Total Progress</span>
                  <span className="agent-progress-value">78%</span>
                </div>
                <Progress percent={78} strokeColor="#0066ff" trailColor="#dbe8f6" showInfo={false} />
              </div>

              <div className="agent-metrics-row">
                <div className="agent-metric-box">
                  <div className="agent-metric-label">TOKENS PROCESSED</div>
                  <div className="agent-metric-value">1.2M</div>
                </div>
                <div className="agent-metric-box">
                  <div className="agent-metric-label">LATENCY</div>
                  <div className="agent-metric-value">142ms</div>
                </div>
                <div className="agent-metric-box">
                  <div className="agent-metric-label">CONFIDENCE</div>
                  <div className="agent-metric-value" style={{ color: '#0066ff' }}>99.4%</div>
                </div>
              </div>
            </Card>

            {/* Feishu Sync Status */}
            <Card className="monitoring-card sync-card" title="FEISHU SYNC STATUS">
              <div className="sync-doc">
                <Avatar size={40} style={{ background: '#dbe8f6', color: '#0066ff' }} icon={<CloudSyncOutlined />} />
                <div className="sync-doc-info">
                  <div className="sync-doc-title">Doc: Q3_Strategy_Review</div>
                  <div className="sync-doc-time">Synchronized 2m ago</div>
                </div>
              </div>
              <Divider style={{ margin: '16px 0' }} />
              <div className="sync-meta-row">
                <span className="sync-meta-label">Write Permissions</span>
                <span className="sync-meta-value active">Active</span>
              </div>
              <div className="sync-meta-row">
                <span className="sync-meta-label">Update Frequency</span>
                <span className="sync-meta-value">Real-time</span>
              </div>
              <div className="sync-meta-row">
                <span className="sync-meta-label">Failover System</span>
                <span className="sync-meta-value stable">Stable</span>
              </div>
            </Card>
          </div>

          {/* Middle Grid */}
          <div className="monitoring-grid middle-grid">
            {/* Execution Pipeline */}
            <Card className="monitoring-card pipeline-card">
              <div className="pipeline-header">
                <RobotOutlined style={{ color: '#0066ff', fontSize: 18 }} />
                <span className="pipeline-title">Execution Pipeline</span>
              </div>
              <div className="pipeline-steps">
                {pipelineSteps.map((step, idx) => (
                  <div key={step.key} className={`pipeline-step ${step.status}`}>
                    <div className="pipeline-step-left">
                      <div className={`pipeline-step-icon ${step.status}`}>
                        {step.status === 'done' ? (
                          <CheckCircleFilled style={{ color: '#52c41a', fontSize: 20 }} />
                        ) : step.status === 'running' ? (
                          <div className="pipeline-step-running-icon">
                            <LoadingOutlined style={{ color: '#0066ff', fontSize: 16 }} />
                          </div>
                        ) : (
                          <MoreOutlined style={{ color: '#c4cdd9', fontSize: 20 }} />
                        )}
                      </div>
                      {idx < pipelineSteps.length - 1 && <div className="pipeline-step-line" />}
                    </div>
                    <div className={`pipeline-step-content ${step.status}`}>
                      <div className="pipeline-step-top">
                        <div>
                          <div className="pipeline-step-title">{step.title}</div>
                          {step.status === 'running' && step.progress ? (
                            <div className="pipeline-step-progress">
                              <Progress
                                percent={step.progress}
                                size="small"
                                strokeColor="#0066ff"
                                trailColor="#dbe8f6"
                                showInfo={false}
                                style={{ width: 120 }}
                              />
                              <span className="pipeline-step-progress-text">{step.progress}% Completed</span>
                            </div>
                          ) : (
                            <div className="pipeline-step-desc">{step.desc}</div>
                          )}
                        </div>
                        <div className="pipeline-step-duration">
                          {step.status === 'running' ? (
                            <Tag color="blue">RUNNING</Tag>
                          ) : (
                            step.duration
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </Card>

            {/* Right Column */}
            <div className="monitoring-right-col">
              {/* Live Console */}
              <Card className="monitoring-card console-card">
                <div className="console-header">
                  <div className="console-title">
                    <span className="console-dot" />
                    Live Console
                  </div>
                  <Button type="text" size="small" style={{ color: '#8b95a8' }}>
                    CLEAR
                  </Button>
                </div>
                <div className="console-body">
                  {consoleLogs.map((log, idx) => (
                    <div key={idx} className="console-line">
                      <span className="console-time">[{log.time}]</span>
                      <span className="console-level" style={{ color: log.color }}>
                        {log.level}
                      </span>
                      <span className="console-text">{log.text}</span>
                    </div>
                  ))}
                </div>
              </Card>

              {/* System Health */}
              <Card className="monitoring-card health-card">
                <div className="health-header">
                  <span>System Health</span>
                  <Tag color="warning" className="health-tag">ATTENTION</Tag>
                </div>
                <div className="health-item warning">
                  <WarningFilled style={{ color: '#faad14', fontSize: 18 }} />
                  <div className="health-item-content">
                    <div className="health-item-title">Memory usage high (88%)</div>
                    <div className="health-item-desc">Performance might be throttled shortly.</div>
                  </div>
                </div>
                <div className="health-item success">
                  <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 18 }} />
                  <div className="health-item-content">
                    <div className="health-item-title">Network stable</div>
                    <div className="health-item-desc">Global edge nodes responding @ 24ms.</div>
                  </div>
                </div>
              </Card>
            </div>
          </div>

          {/* Bottom Metrics */}
          <Card className="monitoring-card metrics-card">
            <div className="metrics-header">
              <span className="metrics-title">Real-time Performance Metrics</span>
              <Segmented
                options={['CPU', 'GPU', 'Network']}
                value={metricTab}
                onChange={setMetricTab}
                className="metrics-segmented"
              />
            </div>
            <div className="metrics-chart">
              {metricsBars.map((h, i) => (
                <Tooltip title={`${h}%`} key={i}>
                  <div
                    className={`metrics-bar ${i === metricsBars.length - 1 ? 'active' : ''}`}
                    style={{ height: `${h}%` }}
                  />
                </Tooltip>
              ))}
            </div>
          </Card>
        </main>
      </div>
    </div>
  )
}
