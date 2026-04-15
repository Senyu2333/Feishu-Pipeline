import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'
import {
  Card,
  Button,
  Input,
  Timeline,
  Progress,
  Avatar,
  Badge,
  Space,
} from 'antd'
import {
  UserAddOutlined,
  FilePdfOutlined,
  CloseOutlined,
  EditOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'

const historyItems = [
  {
    color: '#0066ff',
    title: 'Document Generated',
    time: 'Today, 09:12 AM • AI System',
    desc: 'Initial draft generated based on requirement Alpha-01.',
  },
  {
    color: '#dbe4ee',
    title: 'Initial Screening',
    time: 'Today, 10:45 AM • Sarah Chen',
    desc: 'Formatting checked. Content looks solid. Passed to senior review.',
  },
  {
    color: '#faad14',
    title: 'Resource Alert',
    time: 'Today, 11:05 AM • System Monitor',
    desc: 'GPU cluster 4 experiencing 85% load. Verification may be delayed.',
    alert: true,
  },
]

export default function Approvals() {
  return (
    <div className="min-h-screen bg-background">
      <TopNav />
      <Sidebar />
      <main className="ml-64 mt-14 h-[calc(100vh-3.5rem)] overflow-y-auto p-6">
        <div className="flex justify-between items-start mb-5">
          <div>
            <div className="text-sm text-on-surface-variant mb-1">
              Approvals <span className="text-on-surface/30">›</span>{' '}
              <span className="text-primary font-medium">Requirement #AL-9042</span>
            </div>
            <h1 className="text-2xl font-bold text-on-surface m-0 mb-1">Manual Review & Approval</h1>
            <p className="text-sm text-on-surface-variant m-0">
              Review the AI-generated feasibility report for Project Alpha - Q4 Delivery
            </p>
          </div>
          <Space>
            <Button icon={<UserAddOutlined />} size="large" className="!rounded-lg">
              Assign Reviewer
            </Button>
            <Button icon={<FilePdfOutlined />} size="large" className="!rounded-lg">
              Export PDF
            </Button>
          </Space>
        </div>

        <div className="grid grid-cols-[1fr_360px] gap-5">
          <Card className="!rounded-xl !shadow-sm">
            <div className="flex items-center justify-between px-4 py-3 bg-surface-container-low border-b border-outline-variant">
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-on-surface-variant">menu</span>
                <span className="text-sm font-medium text-on-surface">Feasibility_Report_v2.docx</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="material-symbols-outlined text-on-surface-variant cursor-pointer">zoom_out</span>
                <span className="text-sm text-on-surface-variant min-w-12 text-center">100%</span>
                <span className="material-symbols-outlined text-on-surface-variant cursor-pointer">zoom_in</span>
              </div>
            </div>
            <div className="px-10 py-8">
              <div className="flex items-center gap-4 mb-7">
                <Avatar size={48} className="!bg-surface-container-high !text-primary" icon={<span className="material-symbols-outlined">star</span>} />
                <div className="text-right flex-1">
                  <div className="text-xs font-semibold text-gray-400 tracking-wide">GENERATED DOCUMENT</div>
                  <div className="text-sm text-on-surface font-medium">ID: REQ-9042-ALPHA</div>
                </div>
              </div>

              <h2 className="text-xl font-bold text-on-surface mt-6 mb-4">Abstract</h2>
              <p className="text-sm text-on-surface-variant leading-relaxed mb-4">
                This feasibility report outlines the necessary requirements for the integration of high-level LLM agents into the existing Project Alpha infrastructure.
              </p>

              <h2 className="text-xl font-bold text-on-surface mt-6 mb-4">Technical Specifications</h2>
              <ul className="list-disc pl-5 mb-5 space-y-2 text-sm text-on-surface-variant">
                <li>Integration with existing REST APIs through a secure gateway layer.</li>
                <li>Implementation of vector database indexing for real-time retrieval.</li>
                <li>Estimated GPU resource allocation: 400 TFLOPS across 8 clusters.</li>
              </ul>

              <div className="bg-surface-container-low rounded-xl p-5 my-5">
                <div className="flex justify-between items-center mb-3">
                  <span className="text-xs font-bold text-primary tracking-wide">AI CONFIDENCE SCORE</span>
                  <span className="text-lg font-bold text-on-surface">92%</span>
                </div>
                <Progress percent={92} strokeColor="#0066ff" trailColor="#dbe8f6" showInfo={false} />
              </div>
            </div>
          </Card>

          <aside className="flex flex-col gap-4">
            <Card className="!rounded-xl !shadow-sm" title="Review Action">
              <div className="text-xs font-semibold text-on-surface-variant tracking-wide mb-2">COMMENT & FEEDBACK</div>
              <Input.TextArea placeholder="Enter your detailed feedback here..." rows={4} className="!bg-surface-container-low !border-outline-variant !rounded-lg mb-3" />
              <div className="grid grid-cols-2 gap-2 mb-3">
                <Button icon={<CloseOutlined />} className="!text-red-500 !border-red-200 !bg-red-50 !rounded-lg">
                  Reject
                </Button>
                <Button icon={<EditOutlined />} className="!text-primary !border-primary/20 !bg-primary/5 !rounded-lg">
                  Revision
                </Button>
              </div>
              <Button type="primary" icon={<CheckCircleOutlined />} block className="!h-11 !rounded-xl !text-sm !font-semibold">
                Approve Document
              </Button>
            </Card>

            <Card className="!rounded-xl !shadow-sm" title="Review History" extra={<Badge count="3 Events" className="!bg-surface-container-high !text-primary !font-medium" />}>
              <Timeline className="pt-1" items={historyItems.map((item, idx) => ({
                key: idx,
                color: item.color,
                children: (
                  <div className="p-3 bg-surface-container-low rounded-lg">
                    <div className="font-semibold text-sm text-on-surface">{item.title}</div>
                    <div className="text-xs text-on-surface-variant mt-0.5">{item.time}</div>
                    {item.alert ? (
                      <div className="text-xs text-orange-700 bg-orange-50 border border-orange-200 p-2 rounded mt-2">{item.desc}</div>
                    ) : (
                      <div className="text-xs text-on-surface-variant bg-white p-2 rounded mt-2">{item.desc}</div>
                    )}
                  </div>
                ),
              }))} />
            </Card>

            <div className="flex items-center gap-3 bg-white rounded-xl p-3 shadow-sm">
              <Avatar size="small" className="!bg-primary">A</Avatar>
              <div className="flex-1">
                <div className="text-xs text-on-surface-variant">Currently reviewing</div>
                <div className="text-sm font-semibold text-on-surface">You (Admin)</div>
              </div>
              <Badge status="success" />
            </div>
          </aside>
        </div>
      </main>
    </div>
  )
}
