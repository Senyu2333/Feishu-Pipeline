import { useState } from 'react'
import Sidebar from '../components/Sidebar'
import {
  Card,
  Progress,
  Tag,
  Button,
  Badge,
  Segmented,
  Space,
  Divider,
  Tooltip,
} from 'antd'
import {
  PauseCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons'

const pipelineSteps = [
  { key: 'ingestion', title: 'Data Ingestion', desc: 'Source: Feishu Cloud Workspace', duration: '0.2s', status: 'done', icon: 'cloud_upload' },
  { key: 'mapping', title: 'Semantic Mapping', desc: 'Vectorizing 42 nodes', duration: '1.4s', status: 'done', icon: 'description' },
  { key: 'inference', title: 'Logical Inference', desc: '64% Completed', duration: '', status: 'running', icon: 'psychology', progress: 64 },
  { key: 'formatting', title: 'Output Formatting', desc: 'Waiting for inference...', duration: '--', status: 'pending', icon: 'format_paint' },
]

const consoleLogs = [
  { time: '14:22:01', level: 'INIT', text: 'Loading knowledge base v1.2', color: '#a0aec0' },
  { time: '14:22:05', level: 'INFO', text: 'Connection established to Feishu API', color: '#a0aec0' },
  { time: '14:22:08', level: 'INFO', text: 'Fetching document metadata...', color: '#a0aec0' },
  { time: '14:22:12', level: 'INFO', text: 'Analyzing paragraph headers (Node 1-12)', color: '#a0aec0' },
  { time: '14:22:45', level: 'INFO', text: 'Generation started using model: Aether-Large', color: '#a0aec0' },
  { time: '14:23:01', level: 'WARN', text: 'Latency spike detected in Region 4', color: '#f6ad55' },
]

const metricsBars = [28, 32, 45, 38, 52, 48, 55, 50, 58, 62, 54, 68, 64, 58, 48, 42, 55, 60, 52, 48, 72]

export default function Monitoring() {
  const [metricTab, setMetricTab] = useState('CPU')
  // 左侧导航固定 80px
  const sidebarWidth = 80

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="h-screen overflow-y-auto p-6 transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        <div className="flex justify-between items-start mb-5">
          <div>
            <div className="text-sm text-on-surface-variant mb-1">
              Agents <span className="text-on-surface/30">›</span> Content Optimizer v2.4
            </div>
            <h1 className="text-2xl font-bold text-on-surface m-0">Agent Status Monitor</h1>
          </div>
          <Space>
            <Button icon={<PauseCircleOutlined />} size="large" className="!rounded-lg">Pause Execution</Button>
            <Button type="primary" icon={<ReloadOutlined />} size="large" className="!rounded-lg">Force Restart</Button>
          </Space>
        </div>

        <div className="grid grid-cols-2 gap-5 mb-5">
          <Card className="!rounded-xl !shadow-sm">
            <div className="flex justify-between items-start mb-5">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 rounded-lg bg-surface-container-high flex items-center justify-center">
                  <span className="material-symbols-outlined text-primary text-2xl">smart_toy</span>
                </div>
                <div>
                  <div className="font-semibold text-on-surface">Data Synthesis Agent</div>
                  <div className="text-sm text-on-surface-variant">
                    Cluster ID: <Tag className="!bg-surface-container-low !text-primary !border-0 !text-xs">AF-992-DELTA</Tag>
                  </div>
                </div>
              </div>
              <div className="text-right">
                <Badge status="success" text={<span className="text-xs font-semibold text-green-600">ACTIVE RUNNING</span>} />
                <div className="text-xs text-on-surface-variant mt-1">Uptime: 14h 22m 04s</div>
              </div>
            </div>
            <div className="mb-5">
              <div className="flex justify-between items-center mb-2">
                <span className="text-sm text-on-surface-variant">Total Progress</span>
                <span className="text-sm font-semibold text-on-surface">78%</span>
              </div>
              <Progress percent={78} strokeColor="#0066ff" trailColor="#dbe8f6" showInfo={false} />
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div className="p-3 bg-surface-container-low rounded-lg text-center">
                <div className="text-xs font-semibold text-on-surface-variant tracking-wider mb-1">TOKENS PROCESSED</div>
                <div className="text-lg font-bold text-on-surface">1.2M</div>
              </div>
              <div className="p-3 bg-surface-container-low rounded-lg text-center">
                <div className="text-xs font-semibold text-on-surface-variant tracking-wider mb-1">LATENCY</div>
                <div className="text-lg font-bold text-on-surface">142ms</div>
              </div>
              <div className="p-3 bg-primary/5 rounded-lg text-center">
                <div className="text-xs font-semibold text-on-surface-variant tracking-wider mb-1">CONFIDENCE</div>
                <div className="text-lg font-bold text-primary">99.4%</div>
              </div>
            </div>
          </Card>

          <Card className="!rounded-xl !shadow-sm" title={<span className="text-xs font-bold text-on-surface-variant tracking-wider">FEISHU SYNC STATUS</span>}>
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-lg bg-surface-container-high flex items-center justify-center">
                <span className="material-symbols-outlined text-primary">cloud_sync</span>
              </div>
              <div>
                <div className="font-medium text-on-surface">Doc: Q3_Strategy_Review</div>
                <div className="text-xs text-on-surface-variant">Synchronized 2m ago</div>
              </div>
            </div>
            <Divider style={{ margin: '12px 0' }} />
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-sm text-on-surface-variant">Write Permissions</span>
                <Tag className="!bg-green-50 !text-green-700 !border-0 !text-xs">Active</Tag>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-on-surface-variant">Update Frequency</span>
                <span className="text-sm font-medium text-on-surface">Real-time</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-on-surface-variant">Failover System</span>
                <Tag className="!bg-surface-container-low !text-on-surface !border-0 !text-xs">Stable</Tag>
              </div>
            </div>
          </Card>
        </div>

        <div className="grid grid-cols-[1fr_380px] gap-5 mb-5">
          <Card className="!rounded-xl !shadow-sm">
            <div className="flex items-center gap-2 mb-5">
              <span className="material-symbols-outlined text-primary text-lg">smart_toy</span>
              <span className="font-semibold text-on-surface">Execution Pipeline</span>
            </div>
            <div className="space-y-4">
              {pipelineSteps.map((step, idx) => (
                <div key={step.key} className="flex gap-4">
                  <div className="flex flex-col items-center">
                    <div className={`w-10 h-10 rounded-full flex items-center justify-center ${step.status === 'done' ? 'bg-green-100' : step.status === 'running' ? 'bg-primary/10' : 'bg-surface-container-low'}`}>
                      {step.status === 'done' ? (
                        <span className="material-symbols-outlined text-green-500 text-xl">check_circle</span>
                      ) : step.status === 'running' ? (
                        <span className="material-symbols-outlined text-primary animate-spin" style={{ animationDuration: '2s' }}>progress_activity</span>
                      ) : (
                        <span className="material-symbols-outlined text-on-surface-variant text-xl">more_horiz</span>
                      )}
                    </div>
                    {idx < pipelineSteps.length - 1 && <div className="w-0.5 flex-1 bg-outline-variant my-1" />}
                  </div>
                  <div className="flex-1 pb-5">
                    <div className="flex justify-between items-start mb-1">
                      <div>
                        <div className="font-semibold text-on-surface">{step.title}</div>
                        {step.status === 'running' && step.progress ? (
                          <div className="flex items-center gap-3 mt-2">
                            <Progress percent={step.progress} size="small" strokeColor="#0066ff" trailColor="#dbe8f6" showInfo={false} className="!w-28" />
                            <span className="text-xs text-on-surface-variant">{step.progress}% Completed</span>
                          </div>
                        ) : (
                          <div className="text-sm text-on-surface-variant mt-0.5">{step.desc}</div>
                        )}
                      </div>
                      <div>
                        {step.status === 'running' ? (
                          <Tag color="blue" className="!text-xs">RUNNING</Tag>
                        ) : (
                          <span className="text-sm text-on-surface-variant">{step.duration}</span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Card>

          <div className="flex flex-col gap-4">
            <Card className="!rounded-xl !shadow-sm">
              <div className="flex justify-between items-center mb-3">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                  <span className="font-semibold text-on-surface">Live Console</span>
                </div>
                <Button type="text" size="small" className="!text-on-surface-variant !text-xs">CLEAR</Button>
              </div>
              <div className="bg-gray-900 rounded-lg p-3 font-mono text-xs space-y-1 max-h-48 overflow-y-auto">
                {consoleLogs.map((log, idx) => (
                  <div key={idx} className="text-gray-300">
                    <span className="text-gray-500">[{log.time}]</span>
                    <span className="ml-2" style={{ color: log.color }}>{log.level}</span>
                    <span className="ml-2 text-gray-400">{log.text}</span>
                  </div>
                ))}
              </div>
            </Card>

            <Card className="!rounded-xl !shadow-sm">
              <div className="flex justify-between items-center mb-3">
                <span className="font-semibold text-on-surface">System Health</span>
                <Tag color="warning" className="!text-xs">ATTENTION</Tag>
              </div>
              <div className="space-y-3">
                <div className="flex gap-3 p-3 bg-orange-50 rounded-lg">
                  <span className="material-symbols-outlined text-orange-500 text-lg mt-0.5">warning</span>
                  <div>
                    <div className="font-medium text-on-surface text-sm">Memory usage high (88%)</div>
                    <div className="text-xs text-on-surface-variant mt-0.5">Performance might be throttled shortly.</div>
                  </div>
                </div>
                <div className="flex gap-3 p-3 bg-green-50 rounded-lg">
                  <span className="material-symbols-outlined text-green-500 text-lg mt-0.5">check_circle</span>
                  <div>
                    <div className="font-medium text-on-surface text-sm">Network stable</div>
                    <div className="text-xs text-on-surface-variant mt-0.5">Global edge nodes responding @ 24ms.</div>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </div>

        <Card className="!rounded-xl !shadow-sm">
          <div className="flex justify-between items-center mb-4">
            <span className="font-semibold text-on-surface">Real-time Performance Metrics</span>
            <Segmented options={['CPU', 'GPU', 'Network']} value={metricTab} onChange={setMetricTab} />
          </div>
          <div className="flex items-end gap-1 h-24">
            {metricsBars.map((h, i) => (
              <Tooltip title={`${h}%`} key={i}>
                <div
                  className={`flex-1 rounded-t transition-all cursor-pointer ${i === metricsBars.length - 1 ? 'bg-primary' : 'bg-primary/20 hover:bg-primary/40'}`}
                  style={{ height: `${h}%` }}
                />
              </Tooltip>
            ))}
          </div>
        </Card>
      </main>
    </div>
  )
}
