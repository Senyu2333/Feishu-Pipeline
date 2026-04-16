import { Card, Form, Input, Select, Radio, DatePicker, Button, Space } from 'antd'
import Sidebar from '../components/Sidebar'

export default function NewRequirement() {
  const [form] = Form.useForm()
  // 左侧导航固定 80px
  const sidebarWidth = 80

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="h-screen overflow-y-auto p-6 transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        <div className="mb-6">
          <div className="flex items-center gap-2 text-sm mb-3">
            <span className="text-on-surface-variant">Creation</span>
            <span className="text-on-surface/30">›</span>
            <span className="text-primary font-medium">New Requirement</span>
          </div>
          <div className="flex justify-between items-start gap-5">
            <div>
              <h1 className="text-2xl font-bold text-on-surface mb-1">Create New Requirement</h1>
              <p className="text-sm text-on-surface-variant max-w-xl leading-relaxed">
                Define your functional and technical requirements with precision. Use the Feishu Docs import for existing drafts.
              </p>
            </div>
            <Button icon={<span className="material-symbols-outlined text-sm">upload_file</span>} className="flex items-center gap-2 px-3 py-2 !rounded-lg !border-outline-variant bg-white text-primary text-sm font-medium hover:bg-surface-container-low whitespace-nowrap">
              Import from Feishu Docs
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-[1fr_320px] gap-6 items-start">
          <div>
            <Card className="!rounded-xl !shadow-sm !p-6" bordered={false}>
              <Form form={form} layout="vertical">
                <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">REQUIREMENT TITLE</span>}>
                  <Input placeholder="e.g., Real-time Data Analytics Module" className="!rounded-lg" />
                </Form.Item>
                <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">DETAILED DESCRIPTION</span>}>
                  <Input.TextArea
                    placeholder="Outline the core objectives, functional boundaries, and key constraints..."
                    rows={8}
                    className="!rounded-lg"
                  />
                </Form.Item>
                <Space>
                  <Button icon={<span className="material-symbols-outlined text-sm">attach_file</span>} className="!rounded-lg">
                    Add Attachments
                  </Button>
                  <Button icon={<span className="material-symbols-outlined text-sm">link</span>} className="!rounded-lg">
                    Link Asset
                  </Button>
                </Space>
              </Form>
            </Card>
          </div>

          <aside className="flex flex-col gap-3">
            <Card className="!rounded-xl !shadow-sm !p-4 !bg-surface-container-low" bordered={false}>
              <div className="flex items-center gap-2 text-sm font-semibold text-on-surface mb-4">
                <span className="material-symbols-outlined text-primary">info</span>
                <span>Requirement Metadata</span>
              </div>

              <Form layout="vertical">
                <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">CATEGORY</span>} className="mb-4">
                  <Select defaultValue="feature" suffixIcon={null} className="w-full">
                    <Select.Option value="feature">Product Feature</Select.Option>
                    <Select.Option value="bug">Bug Fix</Select.Option>
                    <Select.Option value="improvement">Improvement</Select.Option>
                  </Select>
                </Form.Item>

                <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">PRIORITY</span>} className="mb-4">
                  <Radio.Group defaultValue="p0" buttonStyle="solid">
                    <Radio.Button value="p0" className="!rounded-l-lg">P0</Radio.Button>
                    <Radio.Button value="p1" className="!rounded-none">P1</Radio.Button>
                    <Radio.Button value="p2" className="!rounded-r-lg">P2</Radio.Button>
                  </Radio.Group>
                </Form.Item>

                <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">TARGET DATE</span>} className="mb-4">
                  <DatePicker className="w-full !rounded-lg" />
                </Form.Item>

                <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">TEAM</span>} className="mb-4">
                  <Select defaultValue="team-alpha" className="w-full">
                    <Select.Option value="team-alpha">Team Alpha</Select.Option>
                    <Select.Option value="team-beta">Team Beta</Select.Option>
                  </Select>
                </Form.Item>
              </Form>
            </Card>

            <Card className="!rounded-xl !shadow-sm !p-4" bordered={false}>
              <div className="flex items-center gap-2 text-sm font-semibold text-on-surface mb-3">
                <span className="material-symbols-outlined text-primary text-base">checklist</span>
                <span>Checklist</span>
              </div>
              <div className="space-y-2">
                {['Use Cases Defined', 'Success Criteria Set', 'Dependencies Listed', 'Stakeholders Notified'].map((item) => (
                  <label key={item} className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" className="w-4 h-4 rounded border-outline text-primary" />
                    <span className="text-sm text-on-surface">{item}</span>
                  </label>
                ))}
              </div>
            </Card>

            <Button type="primary" icon={<span className="material-symbols-outlined text-sm">send</span>} block className="!h-12 !rounded-xl !text-sm !font-semibold">
              Submit Requirement
            </Button>
          </aside>
        </div>
      </main>
    </div>
  )
}
