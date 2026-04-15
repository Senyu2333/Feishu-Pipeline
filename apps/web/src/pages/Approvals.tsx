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
  MenuOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  UserAddOutlined,
  FilePdfOutlined,
  CloseOutlined,
  EditOutlined,
  CheckCircleOutlined,
  StarOutlined,
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
    <div className="app-container">
      <TopNav />
      <div className="main-layout">
        <Sidebar />
        <main className="page-content approvals-page">
          {/* Header */}
          <div className="approvals-header">
            <div>
              <div className="approvals-breadcrumb">
                Approvals <span className="breadcrumb-sep">›</span>{' '}
                <span className="breadcrumb-highlight">Requirement #AL-9042</span>
              </div>
              <h1 className="approvals-title">Manual Review & Approval</h1>
              <p className="approvals-subtitle">
                Review the AI-generated feasibility report for Project Alpha - Q4 Delivery
              </p>
            </div>
            <Space>
              <Button icon={<UserAddOutlined />} size="large">
                Assign Reviewer
              </Button>
              <Button icon={<FilePdfOutlined />} size="large">
                Export PDF
              </Button>
            </Space>
          </div>

          {/* Content Grid */}
          <div className="approvals-layout">
            {/* Document Preview */}
            <Card className="approvals-doc-card">
              {/* Doc Toolbar */}
              <div className="doc-toolbar">
                <div className="doc-toolbar-left">
                  <MenuOutlined style={{ color: '#5a6478' }} />
                  <span className="doc-name">Feasibility_Report_v2.docx</span>
                </div>
                <div className="doc-toolbar-right">
                  <ZoomOutOutlined style={{ color: '#5a6478', cursor: 'pointer' }} />
                  <span className="doc-zoom">100%</span>
                  <ZoomInOutlined style={{ color: '#5a6478', cursor: 'pointer' }} />
                </div>
              </div>

              {/* Doc Content */}
              <div className="doc-content">
                <div className="doc-meta-header">
                  <Avatar
                    size={48}
                    style={{ background: '#e8f2fc', color: '#0066ff', borderRadius: 10 }}
                    icon={<StarOutlined style={{ fontSize: 22 }} />}
                  />
                  <div className="doc-meta-info">
                    <div className="doc-meta-label">GENERATED DOCUMENT</div>
                    <div className="doc-meta-id">ID: REQ-9042-ALPHA</div>
                  </div>
                </div>

                <h2 className="doc-section-title">Abstract</h2>
                <p className="doc-paragraph">
                  This feasibility report outlines the necessary requirements for the integration of high-level LLM agents into the existing Project Alpha infrastructure. The proposed architecture leverages a decentralized node structure to minimize latency during peak processing hours.
                </p>

                <h2 className="doc-section-title">Technical Specifications</h2>
                <ul className="doc-list">
                  <li>Integration with existing REST APIs through a secure gateway layer.</li>
                  <li>Implementation of vector database indexing for real-time retrieval.</li>
                  <li>Estimated GPU resource allocation: 400 TFLOPS across 8 clusters.</li>
                </ul>

                <div className="doc-confidence-box">
                  <div className="doc-confidence-header">
                    <span className="doc-confidence-label">AI CONFIDENCE SCORE</span>
                    <span className="doc-confidence-value">92%</span>
                  </div>
                  <Progress percent={92} strokeColor="#0066ff" trailColor="#dbe8f6" showInfo={false} />
                </div>

                <p className="doc-paragraph">
                  Stakeholders are advised to review Section 4.2 regarding compliance and data sovereignty in international regions. The proposed solution adheres to GDPR and CCPA guidelines through localized data masking.
                </p>
              </div>
            </Card>

            {/* Right Sidebar */}
            <aside className="approvals-sidebar">
              {/* Review Action */}
              <Card className="approvals-card" title="Review Action">
                <div className="review-feedback-label">COMMENT & FEEDBACK</div>
                <Input.TextArea
                  placeholder="Enter your detailed feedback here..."
                  rows={4}
                  className="review-feedback-input"
                />
                <div className="review-actions">
                  <Button icon={<CloseOutlined />} className="review-reject-btn">
                    Reject
                  </Button>
                  <Button icon={<EditOutlined />} className="review-revision-btn">
                    Revision
                  </Button>
                </div>
                <Button type="primary" icon={<CheckCircleOutlined />} block className="review-approve-btn">
                  Approve Document
                </Button>
              </Card>

              {/* Review History */}
              <Card
                className="approvals-card"
                title="Review History"
                extra={<Badge count="3 Events" style={{ backgroundColor: '#e8f2fc', color: '#0066ff', fontWeight: 500 }} />}
              >
                <Timeline className="review-timeline">
                  {historyItems.map((item, idx) => (
                    <Timeline.Item key={idx} color={item.color}>
                      <div className="timeline-item">
                        <div className="timeline-title">{item.title}</div>
                        <div className="timeline-time">{item.time}</div>
                        {item.alert ? (
                          <div className="timeline-alert">{item.desc}</div>
                        ) : (
                          <div className="timeline-desc">{item.desc}</div>
                        )}
                      </div>
                    </Timeline.Item>
                  ))}
                </Timeline>
              </Card>

              {/* Current Reviewer */}
              <div className="current-reviewer">
                <Avatar size="small" style={{ background: '#0066ff' }}>A</Avatar>
                <div>
                  <div className="current-reviewer-label">Currently reviewing</div>
                  <div className="current-reviewer-name">You (Admin)</div>
                </div>
                <Badge status="success" />
              </div>
            </aside>
          </div>
        </main>
      </div>
    </div>
  )
}
