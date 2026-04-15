import { Card, Form, Input, Select, Radio, DatePicker, Button, Avatar, Space } from 'antd'
import {
  FileTextOutlined,
  LinkOutlined,
  SendOutlined,
  InfoCircleFilled,
  StarOutlined,
  CalendarOutlined,
} from '@ant-design/icons'
import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'

export default function NewRequirement() {
  const [form] = Form.useForm()

  return (
    <div className="app-container">
      <TopNav />
      <div className="main-layout">
        <Sidebar />
        <main className="page-content">
          <div className="page-header">
            <div className="breadcrumb">
              <span className="breadcrumb-item">Creation</span>
              <span className="breadcrumb-separator">›</span>
              <span className="breadcrumb-item active">New Requirement</span>
            </div>
            <div className="page-title-row">
              <div>
                <h1 className="page-title">Create New Requirement</h1>
                <p className="page-subtitle">
                  Define your functional and technical requirements with precision. Use the Feishu Docs import for existing drafts.
                </p>
              </div>
              <Button icon={<FileTextOutlined />} className="import-btn">
                Import from Feishu Docs
              </Button>
            </div>
          </div>

          <div className="form-layout">
            <div className="form-main">
              <Card className="form-card" bordered={false}>
                <Form form={form} layout="vertical">
                  <Form.Item label="REQUIREMENT TITLE">
                    <Input placeholder="e.g., Real-time Data Analytics Module" />
                  </Form.Item>
                  <Form.Item label="DETAILED DESCRIPTION">
                    <Input.TextArea
                      placeholder="Outline the core objectives, functional boundaries, and key constraints..."
                      rows={8}
                    />
                  </Form.Item>
                  <Space>
                    <Button icon={<FileTextOutlined />}>Add Attachments</Button>
                    <Button icon={<LinkOutlined />}>Link Asset</Button>
                  </Space>
                </Form>
              </Card>
            </div>

            <aside className="form-sidebar">
              <Card className="metadata-card" bordered={false}>
                <div className="metadata-header">
                  <InfoCircleFilled style={{ color: '#0066ff' }} />
                  <span>Requirement Metadata</span>
                </div>

                <Form layout="vertical">
                  <Form.Item label="CATEGORY">
                    <Select defaultValue="feature" suffixIcon={null}>
                      <Select.Option value="feature">Product Feature</Select.Option>
                    </Select>
                  </Form.Item>

                  <Form.Item label="PRIORITY">
                    <Radio.Group defaultValue="p0" buttonStyle="solid">
                      <Radio.Button value="p0">P0</Radio.Button>
                      <Radio.Button value="p1">P1</Radio.Button>
                      <Radio.Button value="p2">P2</Radio.Button>
                    </Radio.Group>
                  </Form.Item>

                  <Form.Item label="ASSIGNEE">
                    <Select
                      defaultValue="erik"
                      suffixIcon={null}
                      options={[
                        { value: 'erik', label: (
                          <Space>
                            <Avatar size="small" style={{ background: '#a78bfa' }}>EM</Avatar>
                            Erik Magnus
                          </Space>
                        )},
                      ]}
                    />
                  </Form.Item>

                  <Form.Item label="EXPECTED DELIVERY">
                    <DatePicker
                      defaultValue={undefined}
                      placeholder="Oct 24, 2024"
                      suffixIcon={<CalendarOutlined />}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                </Form>
              </Card>

              <Button type="primary" size="large" block icon={<SendOutlined />} className="submit-btn">
                Submit Requirement
              </Button>

              <Button type="text" block className="draft-btn">
                Save as Draft
              </Button>

              <Card className="ai-card" bordered={false}>
                <Avatar size="small" icon={<StarOutlined />} style={{ background: 'rgba(255,255,255,0.1)', color: '#fff', marginBottom: 10 }} />
                <div className="ai-card-title">AI Optimization Active</div>
                <p className="ai-card-desc">
                  Our engine will automatically cross-reference this requirement with current sprint capacity and technical debt logs.
                </p>
              </Card>
            </aside>
          </div>
        </main>
      </div>
    </div>
  )
}
