import { Card, Form, Input, Select, Radio, DatePicker, Button, Space, Tag } from 'antd'
import { useState, useEffect } from 'react'
import Sidebar from '../components/Sidebar'

// TS 后端 API 地址
const API_BASE = 'http://localhost:3001'
const USER_TOKEN_KEY = 'feishu_user_token'

interface FeishuDepartment {
  name?: string
  i18n_name?: { zh_cn?: string; en_us?: string; ja_jp?: string }
  open_department_id?: string
  department_id?: string
  leader_user_id?: string
  leaders?: Array<{ leaderType: number; leaderID: string }>
}

interface DepartmentInfo {
  id: string
  label: string
  leaderUserId?: string
}

export default function NewRequirement() {
  const [form] = Form.useForm()
  const [teams, setTeams] = useState<{ value: string; label: string }[]>([])
  const [selectedTeams, setSelectedTeams] = useState<string[]>([])
  const [loadingTeams, setLoadingTeams] = useState(false)
  const [departmentMap, setDepartmentMap] = useState<Map<string, DepartmentInfo>>(new Map())
  const [leaderNames, setLeaderNames] = useState<Map<string, string>>(new Map())
  const [loadingLeaders, setLoadingLeaders] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const sidebarWidth = 80

  // 提交需求
  const handleSubmit = async () => {
    const leaders = getLeaderTags()
    if (leaders.length === 0) {
      alert('请先选择至少一个团队')
      return
    }

    setSubmitting(true)
    try {
      // 向每个 leader 发送消息
      for (const leader of leaders) {
        const messageContent = `${leader.name}，您好！您收到了一条新的开发需求，请及时查看处理。`
        const content = JSON.stringify({ text: messageContent })
        const uuid = crypto.randomUUID()

        await fetch(`${API_BASE}/api/feishu/send-message`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            receive_id: leader.id,
            receive_id_type: 'open_id',
            msg_type: 'text',
            content,
            uuid,
          }),
        })
      }

      alert(`已成功通知 ${leaders.length} 位团队负责人！`)
    } catch (err) {
      console.error('发送消息失败:', err)
      alert('发送通知失败，请重试')
    } finally {
      setSubmitting(false)
    }
  }

  // 获取部门列表（递归获取所有子部门）
  useEffect(() => {
    const fetchAllDepartments = async () => {
      const userToken = localStorage.getItem(USER_TOKEN_KEY) || ''
      
      if (!userToken) {
        return
      }

      setLoadingTeams(true)
      try {
        // 第一步：递归获取所有部门（处理分页 has_more）
        const allItems: any[] = []
        let pageToken: string | undefined = undefined
        
        do {
          const params = new URLSearchParams({
            department_id: '0',
            fetch_child: 'true',
            page_size: '50',
            user_token: userToken,
          })
          if (pageToken) {
            params.set('page_token', pageToken)
          }
          
          const res = await fetch(`${API_BASE}/api/feishu/department-children?${params.toString()}`)
          const data = await res.json()
          
          if (data.success && data.data?.data?.items) {
            allItems.push(...data.data.data.items)
            pageToken = data.data.data.page_token
            console.log(`获取到 ${data.data.data.items.length} 个部门, has_more: ${data.data.data.has_more}`)
          } else {
            break
          }
        } while (pageToken)
        
        console.log('获取到的所有部门数量:', allItems.length)
        
        if (allItems.length === 0) {
          setLoadingTeams(false)
          return
        }
        
        // 第二步：获取所有部门的完整信息（包括 leader）
        // 提取所有部门的 open_department_id
        const allDeptIds = allItems
          .map((dept: any) => dept.open_department_id || dept.department_id)
          .filter((id: string) => id)
        
        console.log('部门 IDs:', allDeptIds)
        
        // 分批调用批量获取部门信息接口（最多 50 个一批）
        const batchSize = 50
        const allDepartments: any[] = []
        
        for (let i = 0; i < allDeptIds.length; i += batchSize) {
          const batchIds = allDeptIds.slice(i, i + batchSize)
          const batchRes = await fetch(`${API_BASE}/api/feishu/batch-departments`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ department_ids: batchIds, user_token: userToken }),
          })
          const batchData = await batchRes.json()
          
          if (batchData.success && batchData.data?.data?.items) {
            allDepartments.push(...batchData.data.data.items)
          }
        }
        
        console.log('批量获取的部门详情数量:', allDepartments.length)
        
        // 处理部门数据，提取名称和 leader
        const departments: DepartmentInfo[] = allDepartments.map((dept: any) => {
          const id = dept.open_department_id || dept.department_id || ''
          const zhName = dept.i18n_name?.zh_cn?.trim()
          const enName = dept.i18n_name?.en_us?.trim()
          const rawName = dept.name?.trim()
          const label = zhName || enName || rawName || `部门-${id.slice(0, 8)}`
          // 获取主负责人（leaderType === 1）
          const leaderUserId = dept.leaders?.find((l: any) => l.leaderType === 1)?.leaderID || dept.leader_user_id
          return { id, label, leaderUserId }
        })
        
        console.log('处理的部门数据:', departments)
        
        const deptMap = new Map<string, DepartmentInfo>()
        departments.forEach(dept => deptMap.set(dept.id, dept))
        setDepartmentMap(deptMap)
        
        const teamOptions = departments.map(d => ({ value: d.id, label: d.label }))
        if (teamOptions.length > 0) {
          setTeams(teamOptions)
        }
      } catch (err) {
        console.error('Failed to fetch departments:', err)
      } finally {
        setLoadingTeams(false)
      }
    }
    fetchAllDepartments()
  }, [])

  // 当选择的团队变化时，获取 leader 姓名
  useEffect(() => {
    const fetchLeaderNames = async () => {
      if (selectedTeams.length === 0) {
        setLeaderNames(new Map())
        return
      }

      setLoadingLeaders(true)
      try {
        // 从 departmentMap 中获取所有 leader user_id
        const leaderIds = selectedTeams
          .map(teamId => departmentMap.get(teamId)?.leaderUserId)
          .filter((id): id is string => !!id)
          .filter((id, index, arr) => arr.indexOf(id) === index) // 去重

        if (leaderIds.length === 0) {
          setLeaderNames(new Map())
          return
        }

        const userToken = localStorage.getItem(USER_TOKEN_KEY) || ''
        const res = await fetch(`${API_BASE}/api/feishu/batch-user-names`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            user_ids: leaderIds,
            user_token: userToken,
          }),
        })
        const data = await res.json()

        if (data.success && data.data?.data?.users) {
          const nameMap = new Map<string, string>()
          data.data.data.users.forEach((user: any) => {
            // 优先使用外层 name 字段，其次使用 i18n_name.zh_cn
            const name = user.name || user.i18n_name?.zh_cn || user.user_id
            nameMap.set(user.user_id, name)
          })
          setLeaderNames(nameMap)
        }
      } catch (err) {
        console.error('Failed to fetch leader names:', err)
      } finally {
        setLoadingLeaders(false)
      }
    }

    fetchLeaderNames()
  }, [selectedTeams, departmentMap])

  // 获取当前选中团队的 leader 列表
  const getLeaderTags = () => {
    const leaders = selectedTeams
      .map(teamId => departmentMap.get(teamId)?.leaderUserId)
      .filter((id): id is string => !!id)
      .filter((id, index, arr) => arr.indexOf(id) === index) // 去重

    return leaders.map(leaderId => ({
      id: leaderId,
      name: leaderNames.get(leaderId) || leaderId,
    }))
  }

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="h-screen overflow-y-auto p-6 transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        <div className="mb-6">
          <div className="flex items-center gap-2 text-sm mb-3">
            <span className="text-on-surface-variant">创建</span>
            <span className="text-on-surface/30">›</span>
            <span className="text-primary font-medium">新建需求</span>
          </div>
          <div className="flex justify-between items-start gap-5">
            <div>
              <h1 className="text-2xl font-bold text-on-surface mb-1">创建新需求</h1>
              <p className="text-sm text-on-surface-variant max-w-xl leading-relaxed">
                精确描述功能需求和技术需求，可从飞书文档导入现有草稿。
              </p>
            </div>
            <Button icon={<span className="material-symbols-outlined text-sm">upload_file</span>} className="flex items-center gap-2 px-3 py-2 !rounded-lg !border-outline-variant bg-white text-primary text-sm font-medium hover:bg-surface-container-low whitespace-nowrap">
              从飞书文档导入
            </Button>
          </div>
        </div>

        <div className="mb-8">
          <Card className="!rounded-xl !shadow-sm !p-6" bordered={false}>
          <Form form={form} layout="vertical">
            <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">需求标题</span>}>
              <Input placeholder="例如：实时数据分析模块" className="!rounded-lg" />
            </Form.Item>
            <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">详细描述</span>}>
              <Input.TextArea
                placeholder="描述核心目标、功能边界和关键约束..."
                rows={6}
                className="!rounded-lg"
              />
            </Form.Item>

            {/* 需求元数据 */}
            <div className="grid grid-cols-2 gap-4 mb-4">
              <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">分类</span>} className="mb-0">
                <Select defaultValue="feature" suffixIcon={null} className="w-full">
                  <Select.Option value="feature">产品功能</Select.Option>
                  <Select.Option value="bug">缺陷修复</Select.Option>
                  <Select.Option value="improvement">功能改进</Select.Option>
                </Select>
              </Form.Item>
              <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">优先级</span>} className="mb-0">
                <Radio.Group defaultValue="p0" buttonStyle="solid">
                  <Radio.Button value="p0" className="!rounded-l-lg">P0</Radio.Button>
                  <Radio.Button value="p1" className="!rounded-none">P1</Radio.Button>
                  <Radio.Button value="p2" className="!rounded-r-lg">P2</Radio.Button>
                </Radio.Group>
              </Form.Item>
            </div>

            <div className="grid grid-cols-2 gap-4 mb-4">
              <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">目标日期</span>} className="mb-0">
                <DatePicker className="w-full !rounded-lg" />
              </Form.Item>
              <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">团队</span>} className="mb-0">
                <Select
                  mode="multiple"
                  placeholder="选择团队（可多选）"
                  value={selectedTeams}
                  onChange={setSelectedTeams}
                  options={teams}
                  loading={loadingTeams}
                  className="w-full"
                  maxTagCount={2}
                  allowClear
                />
              </Form.Item>
            </div>

            <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">团队Leader</span>} className="mb-4">
              <div className="min-h-[32px] p-2 bg-white rounded border border-gray-200">
                {loadingLeaders ? (
                  <span className="text-gray-400 text-sm">加载中...</span>
                ) : getLeaderTags().length === 0 ? (
                  <span className="text-gray-400 text-sm">选择团队后自动显示</span>
                ) : (
                  <Space wrap>
                    {getLeaderTags().map(leader => (
                      <Tag key={leader.id} color="blue" className="!m-0">
                        {leader.name}
                      </Tag>
                    ))}
                  </Space>
                )}
              </div>
              <div className="text-xs text-gray-400 mt-1">选择团队后自动通知负责人</div>
            </Form.Item>

            {/* 检查清单 */}
            <div className="mb-4">
              <div className="flex items-center gap-2 text-sm font-semibold text-on-surface mb-3">
                <span className="material-symbols-outlined text-primary text-base">checklist</span>
                <span>检查清单</span>
              </div>
              <div className="space-y-2">
                {['已定义用例', '已设置成功标准', '已列出依赖项', '已通知相关方'].map((item) => (
                  <label key={item} className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" className="w-4 h-4 rounded border-outline text-primary" />
                    <span className="text-sm text-on-surface">{item}</span>
                  </label>
                ))}
              </div>
            </div>

            <Space>
              <Button icon={<span className="material-symbols-outlined text-sm">attach_file</span>} className="!rounded-lg">
                添加附件
              </Button>
              <Button icon={<span className="material-symbols-outlined text-sm">link</span>} className="!rounded-lg">
                关联资产
              </Button>
            </Space>
          </Form>
        </Card>
        </div>

        <Button 
          type="primary" 
          icon={<span className="material-symbols-outlined text-sm">send</span>} 
          size="large" 
          className="!rounded-xl !font-semibold"
          onClick={handleSubmit}
          loading={submitting}
        >
          {submitting ? '通知中...' : '提交需求'}
        </Button>
      </main>
    </div>
  )
}
