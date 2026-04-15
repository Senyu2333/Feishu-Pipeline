import { useEffect, useRef, useState } from 'react'
import { Graph } from '@antv/x6'
import { Button, Input, Tag, Space, Tooltip, message } from 'antd'
import {
  ZoomInOutlined,
  ZoomOutOutlined,
  PlayCircleOutlined,
  SaveOutlined,
  UndoOutlined,
  RedoOutlined,
  PlusOutlined,
  SettingOutlined,
  DeleteOutlined,
  AppstoreOutlined,
  ThunderboltOutlined,
  CodeOutlined,
  FlagOutlined,
  NodeExpandOutlined,
} from '@ant-design/icons'
import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'

// 节点类型定义
interface NodeTemplate {
  type: string
  label: string
  icon: React.ReactNode
  color: string
  category: string
}

interface WorkflowNode {
  id: string
  type: string
  label: string
  x: number
  y: number
  config: Record<string, any>
}

// 节点模板
const nodeTemplates: NodeTemplate[] = [
  { type: 'input', label: '用户输入', icon: <FlagOutlined />, color: '#722ed1', category: 'trigger' },
  { type: 'parsing', label: '需求解析', icon: <FlagOutlined />, color: '#722ed1', category: 'ai' },
  { type: 'design', label: '方案设计', icon: <NodeExpandOutlined />, color: '#1890ff', category: 'ai' },
  { type: 'coding', label: '编码实现', icon: <CodeOutlined />, color: '#52c41a', category: 'tool' },
  { type: 'testing', label: '自动化测试', icon: <ThunderboltOutlined />, color: '#fa8c16', category: 'tool' },
  { type: 'review', label: '人工评审', icon: <AppstoreOutlined />, color: '#eb2f96', category: 'logic' },
  { type: 'delivery', label: '交付归档', icon: <FlagOutlined />, color: '#13c2c2', category: 'trigger' },
]

export default function Workflows() {
  const containerRef = useRef<HTMLDivElement>(null)
  const graphRef = useRef<Graph | null>(null)
  
  const [selectedNode, setSelectedNode] = useState<WorkflowNode | null>(null)
  const [zoom, setZoom] = useState(100)
  const [nodes, setNodes] = useState<WorkflowNode[]>([
    { id: 'input-1', type: 'input', label: '用户输入', x: 50, y: 200, config: { description: '用户提交需求' } },
    { id: 'parsing-1', type: 'parsing', label: '需求解析', x: 250, y: 200, config: { description: '解析用户需求，生成需求文档' } },
    { id: 'design-1', type: 'design', label: '方案设计', x: 450, y: 200, config: { description: '设计技术方案和实现路径' } },
    { id: 'coding-1', type: 'coding', label: '编码实现', x: 650, y: 200, config: { description: 'AI 辅助代码生成' } },
    { id: 'testing-1', type: 'testing', label: '自动化测试', x: 850, y: 200, config: { description: '运行单元测试和集成测试' } },
    { id: 'review-1', type: 'review', label: '人工评审', x: 1050, y: 200, config: { description: '人工代码评审，重点节点' } },
    { id: 'delivery-1', type: 'delivery', label: '交付归档', x: 1250, y: 200, config: { description: '交付代码并归档' } },
  ])
  const [nodeConfig, setNodeConfig] = useState<Record<string, any>>({})
  const [contextMenuPos, setContextMenuPos] = useState<{ x: number; y: number } | null>(null)

  // 初始化图 - 只在挂载时执行
  useEffect(() => {
    if (!containerRef.current) return

    const graph = new Graph({
      container: containerRef.current,
      width: containerRef.current.clientWidth,
      height: containerRef.current.clientHeight,
      background: { color: '#f7f8fa' },
      grid: { visible: true, type: 'dot', args: { color: '#e5e5e5', thickness: 1, scaleFactor: 5 } },
      panning: true,
      mousewheel: { enabled: true, modifiers: ['ctrl', 'meta'], factor: 1.1 },
      connecting: {
        snap: true,
        anchor: 'center',
        connectionPoint: 'anchor',
        allowBlank: false,
        allowLoop: false,
        highlight: true,
      },
    })

    // 初始化添加节点
    nodes.forEach(node => {
      graph.addNode({
        id: node.id,
        x: node.x,
        y: node.y,
        width: 180,
        height: 72,
        draggable: true,
        attrs: {
          body: {
            fill: '#fff',
            stroke: '#d9d9d9',
            strokeWidth: 1,
            rx: 8,
            ry: 8,
          },
          label: {
            text: node.label,
            fill: '#262626',
            fontSize: 13,
            fontWeight: 600,
            refX: 0.5,
            refY: 0.4,
            textAnchor: 'middle',
          },
        },
        data: node,
      })
    })

    // 注册事件
    graph.on('node:selected', ({ node }) => {
      const id = node.id
      const nodeData = nodes.find(n => n.id === id)
      if (nodeData) {
        setSelectedNode(nodeData)
        setNodeConfig(nodeData.config)
      }
    })
    graph.on('blank:click', () => {
      setSelectedNode(null)
    })
    graph.on('node:moved', () => {
      // X6 自动管理位置，不触发 React 重新渲染
    })
    graph.on('node:contextmenu', ({ e, x, y }) => {
      e.preventDefault()
      setContextMenuPos({ x, y })
    })
    graph.on('blank:contextmenu', ({ e, x, y }) => {
      e.preventDefault()
      setContextMenuPos({ x, y })
    })

    // 添加默认边 - 使用简化的直线连接，避免干扰节点拖拽
    graph.addEdge({
      source: 'input-1',
      target: 'parsing-1',
      sourceCell: 'input-1',
      targetCell: 'parsing-1',
      sourcePort: 'out',
      targetPort: 'in',
      attrs: {
        line: {
          stroke: '#b37feb',
          strokeWidth: 2,
        },
      },
      connector: { name: 'normal' },
      router: { name: 'orth' },
    })

    graph.addEdge({
      source: 'parsing-1',
      target: 'design-1',
      sourceCell: 'parsing-1',
      targetCell: 'design-1',
      sourcePort: 'out',
      targetPort: 'in',
      attrs: {
        line: {
          stroke: '#b37feb',
          strokeWidth: 2,
        },
      },
      connector: { name: 'normal' },
      router: { name: 'orth' },
    })

    graph.addEdge({
      source: 'design-1',
      target: 'coding-1',
      sourceCell: 'design-1',
      targetCell: 'coding-1',
      sourcePort: 'out',
      targetPort: 'in',
      attrs: {
        line: {
          stroke: '#b37feb',
          strokeWidth: 2,
        },
      },
      connector: { name: 'normal' },
      router: { name: 'orth' },
    })

    graph.addEdge({
      source: 'coding-1',
      target: 'testing-1',
      sourceCell: 'coding-1',
      targetCell: 'testing-1',
      sourcePort: 'out',
      targetPort: 'in',
      attrs: {
        line: {
          stroke: '#b37feb',
          strokeWidth: 2,
        },
      },
      connector: { name: 'normal' },
      router: { name: 'orth' },
    })

    graph.addEdge({
      source: 'testing-1',
      target: 'review-1',
      sourceCell: 'testing-1',
      targetCell: 'review-1',
      sourcePort: 'out',
      targetPort: 'in',
      attrs: {
        line: {
          stroke: '#b37feb',
          strokeWidth: 2,
        },
      },
      connector: { name: 'normal' },
      router: { name: 'orth' },
    })

    graph.addEdge({
      source: 'review-1',
      target: 'delivery-1',
      sourceCell: 'review-1',
      targetCell: 'delivery-1',
      sourcePort: 'out',
      targetPort: 'in',
      attrs: {
        line: {
          stroke: '#b37feb',
          strokeWidth: 2,
        },
      },
      connector: { name: 'normal' },
      router: { name: 'orth' },
    })

    graphRef.current = graph

    const handleResize = () => {
      if (containerRef.current) {
        graph.resize(containerRef.current.clientWidth, containerRef.current.clientHeight)
      }
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      graph.dispose()
    }
  }, [])  // 空依赖，只初始化一次

  // 节点选择状态更新
  useEffect(() => {
    const graph = graphRef.current
    if (!graph) return

    // 遍历所有节点，更新选中状态样式
    graph.getNodes().forEach(node => {
      const nodeData = node.getData() as WorkflowNode
      const isSelected = selectedNode?.id === node.id
      const template = nodeTemplates.find(t => t.type === nodeData.type)
      
      if (isSelected && template) {
        node.attr('body/stroke', template.color)
        node.attr('body/strokeWidth', 2)
      } else {
        node.attr('body/stroke', '#d9d9d9')
        node.attr('body/strokeWidth', 1)
      }
    })
  }, [selectedNode])

  // 添加/删除节点时同步到画布
  useEffect(() => {
    const graph = graphRef.current
    if (!graph) return

    // 获取当前画布中的节点 ID
    const existingIds = new Set(graph.getNodes().map(n => n.id))
    const targetIds = new Set(nodes.map(n => n.id))

    // 删除不存在的节点
    existingIds.forEach(id => {
      if (!targetIds.has(id)) {
        graph.getCellById(id)?.remove()
      }
    })

    // 添加新节点
    nodes.forEach(node => {
      if (!existingIds.has(node.id)) {
        graph.addNode({
          id: node.id,
          x: node.x,
          y: node.y,
          width: 180,
          height: 72,
          draggable: true,
          attrs: {
            body: {
              fill: '#fff',
              stroke: '#d9d9d9',
              strokeWidth: 1,
              rx: 8,
              ry: 8,
            },
            label: {
              text: node.label,
              fill: '#262626',
              fontSize: 13,
              fontWeight: 600,
              refX: 0.5,
              refY: 0.4,
              textAnchor: 'middle',
            },
          },
          data: node,
        })
      }
    })
  }, [nodes])

  // 缩放控制
  const handleZoomIn = () => {
    const graph = graphRef.current
    if (graph) {
      const newZoom = Number(graph.zoom(0.1))
      setZoom(Math.round(newZoom * 100))
    }
  }

  const handleZoomOut = () => {
    const graph = graphRef.current
    if (graph) {
      const newZoom = Number(graph.zoom(-0.1))
      setZoom(Math.round(newZoom * 100))
    }
  }

  const handleFit = () => {
    const graph = graphRef.current
    if (graph) {
      graph.center()
      graph.zoomTo(1)
      setZoom(100)
    }
  }

  // 添加节点
  const handleAddNode = (type: string) => {
    const template = nodeTemplates.find(t => t.type === type)
    if (!template) return

    const newNode: WorkflowNode = {
      id: `${type}-${Date.now()}`,
      type,
      label: template.label,
      x: contextMenuPos ? contextMenuPos.x : 400,
      y: contextMenuPos ? contextMenuPos.y : 200,
      config: { description: '' },
    }
    setNodes(prev => [...prev, newNode])
    setContextMenuPos(null)
    message.success(`添加 ${template.label} 节点`)
  }

  // 删除节点
  const handleDeleteNode = () => {
    if (!selectedNode) return
    setNodes(prev => prev.filter(n => n.id !== selectedNode.id))
    setSelectedNode(null)
    message.success('节点已删除')
  }

  // 更新节点配置
  const handleUpdateConfig = (key: string, value: any) => {
    if (!selectedNode) return
    setNodeConfig(prev => ({ ...prev, [key]: value }))
    setNodes(prev => prev.map(n => 
      n.id === selectedNode.id ? { ...n, config: { ...n.config, [key]: value } } : n
    ))
  }

  // 运行工作流
  const handleRun = () => {
    message.loading({ content: '正在运行工作流...', key: 'run' })
    setTimeout(() => {
      message.success({ content: '工作流执行完成', key: 'run' })
    }, 2000)
  }

  // 保存
  const handleSave = () => {
    message.success('工作流已保存')
  }

  // 右键菜单项
  // const getContextMenuItems = (): MenuProps['items'] => [
  // 右键菜单通过自定义弹窗实现

  // 节点右键菜单
  // const getNodeContextMenuItems = (): MenuProps['items'] => [

  return (
    <div className="min-h-screen bg-background">
      <TopNav showSearch />
      <Sidebar />
      <main className="ml-64 mt-14 h-[calc(100vh-3.5rem)] flex flex-col overflow-hidden bg-surface-dim">
        {/* Toolbar */}
        <div className="flex items-center justify-between px-4 py-2 bg-surface-container-low border-b border-outline-variant">
          <div className="flex items-center gap-2">
            <Space>
              <Tooltip title="撤销">
                <Button type="text" icon={<UndoOutlined />} />
              </Tooltip>
              <Tooltip title="重做">
                <Button type="text" icon={<RedoOutlined />} />
              </Tooltip>
              <div className="w-px h-6 bg-outline-variant mx-1" />
              <Tooltip title="缩小">
                <Button type="text" icon={<ZoomOutOutlined />} onClick={handleZoomOut} />
              </Tooltip>
              <span className="text-sm text-on-surface-variant min-w-12 text-center">{zoom}%</span>
              <Tooltip title="放大">
                <Button type="text" icon={<ZoomInOutlined />} onClick={handleZoomIn} />
              </Tooltip>
              <Tooltip title="适应画布">
                <Button type="text" icon={<AppstoreOutlined />} onClick={handleFit} />
              </Tooltip>
            </Space>
          </div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-on-surface">Alpha Release Pipeline</span>
            <Tag color="default">草稿</Tag>
          </div>
          <div className="flex items-center gap-2">
            <Space>
              <Button icon={<SaveOutlined />} onClick={handleSave} className="!rounded-lg">
                保存
              </Button>
              <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleRun} className="!rounded-lg">
                运行
              </Button>
            </Space>
          </div>
        </div>

        {/* Main Content */}
        <div className="flex-1 flex overflow-hidden">
          {/* Left Panel - Node Templates */}
          <aside className="w-56 bg-surface-container-lowest border-r border-outline-variant overflow-y-auto">
            <div className="p-3 space-y-4">
              <div>
                <div className="text-xs font-semibold text-on-surface-variant tracking-wider mb-2">基础节点</div>
                <div className="space-y-1">
                  {nodeTemplates.filter(t => t.category === 'trigger').map(template => (
                    <div
                      key={template.type}
                      className="flex items-center gap-2 p-2 rounded-lg hover:bg-surface-variant cursor-pointer transition-colors"
                      onClick={() => handleAddNode(template.type)}
                    >
                      <div className="w-6 h-6 rounded flex items-center justify-center text-xs" style={{ background: template.color + '20', color: template.color }}>
                        {template.icon}
                      </div>
                      <span className="text-sm text-on-surface">{template.label}</span>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-xs font-semibold text-on-surface-variant tracking-wider mb-2">AI 节点</div>
                <div className="space-y-1">
                  {nodeTemplates.filter(t => t.category === 'ai').map(template => (
                    <div
                      key={template.type}
                      className="flex items-center gap-2 p-2 rounded-lg hover:bg-surface-variant cursor-pointer transition-colors"
                      onClick={() => handleAddNode(template.type)}
                    >
                      <div className="w-6 h-6 rounded flex items-center justify-center text-xs" style={{ background: template.color + '20', color: template.color }}>
                        {template.icon}
                      </div>
                      <span className="text-sm text-on-surface">{template.label}</span>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-xs font-semibold text-on-surface-variant tracking-wider mb-2">工具节点</div>
                <div className="space-y-1">
                  {nodeTemplates.filter(t => t.category === 'tool').map(template => (
                    <div
                      key={template.type}
                      className="flex items-center gap-2 p-2 rounded-lg hover:bg-surface-variant cursor-pointer transition-colors"
                      onClick={() => handleAddNode(template.type)}
                    >
                      <div className="w-6 h-6 rounded flex items-center justify-center text-xs" style={{ background: template.color + '20', color: template.color }}>
                        {template.icon}
                      </div>
                      <span className="text-sm text-on-surface">{template.label}</span>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-xs font-semibold text-on-surface-variant tracking-wider mb-2">逻辑节点</div>
                <div className="space-y-1">
                  {nodeTemplates.filter(t => t.category === 'logic').map(template => (
                    <div
                      key={template.type}
                      className="flex items-center gap-2 p-2 rounded-lg hover:bg-surface-variant cursor-pointer transition-colors"
                      onClick={() => handleAddNode(template.type)}
                    >
                      <div className="w-6 h-6 rounded flex items-center justify-center text-xs" style={{ background: template.color + '20', color: template.color }}>
                        {template.icon}
                      </div>
                      <span className="text-sm text-on-surface">{template.label}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </aside>

            {/* Center - Canvas with Context Menu */}
            <div className="flex-1 relative" ref={containerRef}>
              {/* Hint */}
              <div className="absolute top-4 left-1/2 -translate-x-1/2 z-10 px-4 py-2 bg-gray-800 text-white text-sm rounded-full opacity-60">
                右键画布添加节点
              </div>

              {/* Mini Map */}
              <div className="absolute bottom-4 right-4 w-36 bg-white rounded-lg shadow-lg border border-gray-200 overflow-hidden z-10">
                <div className="px-2 py-1 text-xs font-semibold text-gray-500 border-b border-gray-100">缩略图</div>
                <div className="relative h-20 p-2">
                  {nodes.map(node => {
                    const template = nodeTemplates.find(t => t.type === node.type)
                    return (
                      <Tooltip key={node.id} title={node.label}>
                        <div
                          className={`absolute w-2 h-2 rounded-sm ${selectedNode?.id === node.id ? 'ring-2 ring-blue-500' : ''}`}
                          style={{
                            left: `${(node.x / 12)}%`,
                            top: `${(node.y / 8)}%`,
                            background: template?.color
                          }}
                        />
                      </Tooltip>
                    )
                  })}
                </div>
              </div>

              {/* Context Menu */}
              {contextMenuPos && (
                <>
                  <div className="fixed inset-0 z-20" onClick={() => setContextMenuPos(null)} />
                  <div
                    className="fixed bg-white rounded-lg shadow-xl border border-gray-200 py-1 z-30 min-w-48"
                    style={{ left: contextMenuPos.x, top: contextMenuPos.y }}
                  >
                    <div className="px-3 py-2 text-xs font-semibold text-gray-400 tracking-wider border-b border-gray-100">
                      <PlusOutlined /> 添加节点
                    </div>
                    <div className="py-1">
                      {nodeTemplates.map(template => (
                        <div
                          key={template.type}
                          className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer"
                          onClick={() => handleAddNode(template.type)}
                        >
                          <div className="w-5 h-5 rounded flex items-center justify-center text-xs" style={{ background: template.color + '20', color: template.color }}>
                            {template.icon}
                          </div>
                          <span className="text-sm text-gray-700">{template.label}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </div>

            {/* Right Panel - Properties */}
            <aside className={`w-72 bg-white border-l border-gray-200 overflow-y-auto transition-all ${selectedNode ? '' : 'flex items-center justify-center'}`}>
              {selectedNode ? (
                <div className="p-4">
                  <div className="flex items-center justify-between mb-4">
                    <span className="font-semibold text-gray-800">节点配置</span>
                    <Tag color="blue">{nodeTemplates.find(t => t.type === selectedNode.type)?.label}</Tag>
                  </div>
                  <div className="space-y-4">
                    <div>
                      <label className="text-xs font-semibold text-gray-500">节点名称</label>
                      <Input
                        value={selectedNode.label}
                        onChange={(e) => {
                          const value = e.target.value
                          setNodes(prev => prev.map(n => n.id === selectedNode.id ? { ...n, label: value } : n))
                          setSelectedNode(prev => prev ? { ...prev, label: value } : null)
                        }}
                        className="mt-1 rounded-lg"
                      />
                    </div>
                    <div>
                      <label className="text-xs font-semibold text-gray-500">描述</label>
                      <Input.TextArea
                        rows={2}
                        placeholder="节点功能描述..."
                        value={nodeConfig.description || ''}
                        onChange={(e) => handleUpdateConfig('description', e.target.value)}
                        className="mt-1 rounded-lg"
                      />
                    </div>
                    <div className="pt-3 border-t border-gray-100 flex gap-2">
                      <Button danger icon={<DeleteOutlined />} onClick={handleDeleteNode} className="flex-1">
                        删除
                      </Button>
                      <Button type="primary" onClick={handleSave} className="flex-1">
                        保存配置
                      </Button>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-center text-gray-400 p-4">
                  <SettingOutlined style={{ fontSize: 48 }} />
                  <p className="mt-3 text-sm">选择节点以编辑配置</p>
                  <p className="text-xs mt-1">点击画布上的节点查看和修改其属性</p>
                </div>
              )}
            </aside>
          </div>

          {/* Status Bar */}
          <div className="flex items-center justify-between px-4 py-1.5 bg-white border-t border-gray-200 text-xs text-gray-500">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-green-500" />
              <span>就绪</span>
            </div>
            <div>
              <span>自动保存</span>
            </div>
            <div className="flex items-center gap-4">
              <span>节点: {nodes.length}</span>
              <span>连线: {nodes.length - 1}</span>
            </div>
          </div>
        </main>
    </div>
  )
}
