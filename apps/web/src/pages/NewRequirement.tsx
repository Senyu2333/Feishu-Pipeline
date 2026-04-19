import { Card, Form, Input, Select, Radio, DatePicker, Button, Space, Tag, Timeline } from 'antd'
import { useState, useEffect } from 'react'
import Sidebar from '../components/Sidebar'

// TS 后端 API 地址
const API_BASE = 'http://localhost:3001'
const USER_TOKEN_KEY = 'feishu_user_token'
const USER_OPEN_ID_KEY = 'feishu_user_open_id'

// 部门信息接口
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
  const [generatingDoc, setGeneratingDoc] = useState(false)
  
  // AI 步骤项类型
  interface AiChainItem {
    key: string
    title: string
    description?: string
    status: 'loading' | 'success' | 'error'
    children?: React.ReactNode
  }
  
  // AI 生成状态
  const [aiChainItems, setAiChainItems] = useState<AiChainItem[]>([])
  const [isAiGenerating, setIsAiGenerating] = useState(false)
  
  const sidebarWidth = 80
  
  // 提交需求
  const handleSubmit = async () => {
    const leaders = getLeaderTags()
    if (leaders.length === 0) {
      alert('请先选择至少一个团队')
      return
    }

    console.log('提交的 Leaders:', leaders)

    // 获取表单数据
    const formData = form.getFieldsValue()
    const requirementTitle = formData['requirement_title'] || formData[0] || '未命名需求'

    setSubmitting(true)
    try {
      // 1. AI 生成 API 设计文档（如果有详细描述）
      let docUrl = ''
      const description = formData['description'] || ''
      
      if (description) {
        setGeneratingDoc(true)
        setIsAiGenerating(true)
        // 重置 AI 状态
        setAiChainItems([])
        
        // 初始思考节点
        setAiChainItems([{
          key: 'start',
          title: '🚀 开始分析需求',
          description: '正在理解用户输入...',
          status: 'loading',
        } as any])
        
        try {
          const userToken = localStorage.getItem(USER_TOKEN_KEY) || ''
          const openId = localStorage.getItem(USER_OPEN_ID_KEY) || ''
          
          // 调用 SSE AI 生成文档
          const aiRes = await fetch(`${API_BASE}/api/ai/chat/stream`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
              message: `请根据以下需求描述生成 API 设计文档：\n\n需求标题：${requirementTitle}\n\n详细描述：${description}`,
              user_token: userToken,
              open_id: openId
            }),
          })
          
          if (!aiRes.ok || !aiRes.body) {
            throw new Error('SSE 请求失败')
          }
          
          const reader = aiRes.body.getReader()
          const decoder = new TextDecoder()
          let buffer = ''
          
          while (true) {
            const { done, value } = await reader.read()
            if (done) break
            
            buffer += decoder.decode(value, { stream: true })
            const lines = buffer.split('\n')
            buffer = lines.pop() || ''
            
            for (const line of lines) {
              if (!line.startsWith('data: ')) continue
              try {
                const data = JSON.parse(line.slice(6))
                
                switch (data.event) {
                  case 'start':
                    // 只添加一个分析节点
                    setAiChainItems([{
                      key: 'step_0',
                      title: '🤖 AI 正在生成文档...',
                      description: data.content || '正在分析需求并生成文档',
                      status: 'loading' as const,
                    }])
                    break
                    
                  case 'done':
                    // 完成所有
                    if (data.content) {
                      // 从内容中提取文档链接
                      const docMatch = data.content.match(/https:\/\/feishu\.cn\/docx\/[a-zA-Z0-9_-]+/)
                      if (docMatch) {
                        docUrl = docMatch[0]
                      }
                    }
                    setAiChainItems([{
                      key: 'step_0',
                      title: docUrl ? '✅ 文档生成完成' : '📝 处理完成',
                      description: docUrl ? `文档链接: ${docUrl}` : 'AI 处理已完成',
                      status: 'success' as const,
                    }])
                    break
                    
                  case 'error':
                    console.error('AI Stream Error:', data.message)
                    setAiChainItems([{
                      key: 'step_0',
                      title: '❌ 发生错误',
                      description: data.message,
                      status: 'error' as const,
                    }])
                    break
                }
              } catch (e) {
                // 忽略 JSON 解析错误
              }
            }
          }
        } catch (aiErr) {
          console.error('AI 生成文档失败:', aiErr)
          // 添加错误节点
          const errorKey = `error_${Date.now()}`
          setAiChainItems(prev => [...(prev || []), {
            key: errorKey,
            title: '❌ 请求失败',
            description: String(aiErr),
            status: 'error' as any,
          }])
        }
        setGeneratingDoc(false)
        setIsAiGenerating(false)
      }

      // 2. 发送消息给每个 leader
      for (const leader of leaders) {
        // 构建消息内容
        let messageContent = `${leader.name}，您好！您收到了一条新的开发需求，请及时查看处理。\n\n`
        messageContent += `📋 需求标题：${requirementTitle}\n`
        
        // 添加截止日期
        if (formData['target_date']) {
          const targetDate = new Date(formData['target_date']).toLocaleDateString('zh-CN')
          messageContent += `📅 截止日期：${targetDate}\n`
        }
        
        // 添加优先级
        if (formData['priority']) {
          messageContent += `🔥 优先级：${formData['priority']}\n`
        }
        
        // 添加分类
        if (formData['category']) {
          const categoryMap: Record<string, string> = {
            feature: '功能需求',
            bug: 'Bug修复',
            optimization: '优化改进',
            other: '其他'
          }
          messageContent += `📂 分类：${categoryMap[formData['category']] || formData['category']}\n`
        }
        
        // 如果生成了文档，添加文档链接
        if (docUrl) {
          messageContent += `\n📄 API 设计文档：${docUrl}\n`
        }
        
        messageContent += `\n请及时查看并处理！`
        
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

      let alertMsg = `需求已提交！\n已通知 ${leaders.length} 位团队负责人`
      if (docUrl) {
        alertMsg += `\n\nAI 已自动生成 API 设计文档：${docUrl}`
      }
      alert(alertMsg)
    } catch (err) {
      console.error('提交失败:', err)
      alert('提交失败，请重试')
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
            body: JSON.stringify({ 
              department_ids: batchIds, 
              user_token: userToken,
              user_id_type: 'open_id'  // 明确要求返回 open_id
            }),
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
            <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">需求标题</span>} name="requirement_title">
              <Input placeholder="例如：实时数据分析模块" className="!rounded-lg" />
            </Form.Item>
            <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">详细描述</span>} name="description">
              <Input.TextArea
                placeholder="描述核心目标、功能边界和关键约束...\n可以粘贴飞书文档链接，AI将自动读取文档内容"
                rows={8}
                className="!rounded-lg"
              />
            </Form.Item>

            {/* 需求元数据 */}
            <div className="grid grid-cols-2 gap-4 mb-4">
              <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">优先级</span>} name="priority" className="mb-0">
                <Radio.Group defaultValue="p0" buttonStyle="solid">
                  <Radio.Button value="p0" className="!rounded-l-lg">P0</Radio.Button>
                  <Radio.Button value="p1" className="!rounded-none">P1</Radio.Button>
                  <Radio.Button value="p2" className="!rounded-r-lg">P2</Radio.Button>
                </Radio.Group>
              </Form.Item>
            </div>

            <div className="grid grid-cols-2 gap-4 mb-4">
              <Form.Item label={<span className="text-xs font-semibold text-on-surface-variant tracking-wider">目标日期</span>} name="target_date" className="mb-0">
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

        {/* AI 生成状态展示 */}
        {(generatingDoc || (aiChainItems && aiChainItems.length > 0)) && (
          <Card className="!rounded-xl !shadow-sm !p-4 mb-4" bordered={false}>
            <div className="flex items-center gap-2 mb-4 pb-3 border-b border-gray-100">
              <span className="material-symbols-outlined text-primary" style={{ fontSize: 20 }}>
                {isAiGenerating ? 'psychology' : 'check_circle'}
              </span>
              <span className="font-medium">
                {isAiGenerating ? 'AI 正在处理...' : 'AI 处理完成'}
              </span>
            </div>
            <Timeline
              items={aiChainItems.map(item => ({
                color: item.status === 'loading' ? 'blue' : item.status === 'success' ? 'green' : 'red',
                content: (
                  <div>
                    <div className="font-medium text-sm">{item.title}</div>
                    {item.description && (
                      <div className="text-xs text-gray-500 mt-1">{item.description}</div>
                    )}
                    {item.children}
                  </div>
                ),
              }))}
            />
          </Card>
        )}

        <Button 
          type="primary" 
          icon={<span className="material-symbols-outlined text-sm">send</span>} 
          size="large" 
          className="!rounded-xl !font-semibold"
          onClick={handleSubmit}
          loading={submitting || generatingDoc}
        >
          {generatingDoc ? 'AI 生成文档...' : submitting ? '通知中...' : '提交需求'}
        </Button>
      </main>
    </div>
  )
}
