import { useEffect, useRef, useState } from 'react'
import { Graph } from '@antv/x6'
import { Button, Select, Slider, Input, Tag, Space, Tooltip, message } from 'antd'
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
  GlobalOutlined,
  CodeOutlined,
  BranchesOutlined,
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
  const minimapRef = useRef<HTMLDivElement>(null)
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
    <div className="app-container">
      <TopNav showSearch />
      <div className="main-layout">
        <Sidebar />
        <main className="workflow-page">
          {/* Toolbar */}
          <div className="dify-toolbar">
            <div className="toolbar-left">
              <Space>
                <Tooltip title="撤销">
                  <Button type="text" icon={<UndoOutlined />} />
                </Tooltip>
                <Tooltip title="重做">
                  <Button type="text" icon={<RedoOutlined />} />
                </Tooltip>
                <div className="toolbar-divider" />
                <Tooltip title="缩小">
                  <Button type="text" icon={<ZoomOutOutlined />} onClick={handleZoomOut} />
                </Tooltip>
                <span className="zoom-display">{zoom}%</span>
                <Tooltip title="放大">
                  <Button type="text" icon={<ZoomInOutlined />} onClick={handleZoomIn} />
                </Tooltip>
                <Tooltip title="适应画布">
                  <Button type="text" icon={<AppstoreOutlined />} onClick={handleFit} />
                </Tooltip>
              </Space>
            </div>
            <div className="toolbar-center">
              <span className="pipeline-name">Alpha Release Pipeline</span>
              <Tag color="default">草稿</Tag>
            </div>
            <div className="toolbar-right">
              <Space>
                <Button icon={<SaveOutlined />} onClick={handleSave}>
                  保存
                </Button>
                <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleRun}>
                  运行
                </Button>
              </Space>
            </div>
          </div>

          {/* Main Content */}
          <div className="workflow-body dify-layout">
            {/* Center - Canvas with Context Menu */}
            <div className="workflow-canvas" ref={containerRef}>
              {/* Hint */}
              <div className="canvas-hint">
                <span>右键画布添加节点</span>
              </div>

              {/* Mini Map */}
              <div className="minimap" ref={minimapRef}>
                <div className="minimap-title">缩略图</div>
                <div className="minimap-content">
                  {nodes.map(node => {
                    const template = nodeTemplates.find(t => t.type === node.type)
                    return (
                      <Tooltip key={node.id} title={node.label}>
                        <div 
                          className={`minimap-node ${selectedNode?.id === node.id ? 'active' : ''}`}
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
                  <div 
                    className="context-menu-overlay"
                    onClick={() => setContextMenuPos(null)}
                  />
                  <div 
                    className="context-menu"
                    style={{ 
                      left: contextMenuPos.x, 
                      top: contextMenuPos.y 
                    }}
                  >
                    <div className="context-menu-header">
                      <PlusOutlined /> 添加节点
                    </div>
                    <div className="context-menu-items">
                      {nodeTemplates.map(template => (
                        <div 
                          key={template.type}
                          className="context-menu-item"
                          onClick={() => handleAddNode(template.type)}
                        >
                          <span 
                            className="node-icon"
                            style={{ background: `${template.color}20`, color: template.color }}
                          >
                            {template.icon}
                          </span>
                          <span className="node-label">{template.label}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </div>

            {/* Right Panel - Properties */}
            <aside className={`properties-panel ${selectedNode ? 'open' : ''}`}>
              {selectedNode ? (
                <>
                  <div className="properties-header">
                    <div className="properties-title">
                      <span>节点配置</span>
                      <Tag color="blue">{nodeTemplates.find(t => t.type === selectedNode.type)?.label}</Tag>
                    </div>
                  </div>

                  <div className="properties-content">
                    {/* Node Info */}
                    <div className="properties-section">
                      <div className="section-title">基本信息</div>
                      <div className="form-item">
                        <label>节点名称</label>
                        <Input 
                          value={selectedNode.label}
                          onChange={(e) => {
                            const value = e.target.value
                            setNodes(prev => prev.map(n => 
                              n.id === selectedNode.id ? { ...n, label: value } : n
                            ))
                            setSelectedNode(prev => prev ? { ...prev, label: value } : null)
                          }}
                        />
                      </div>
                      <div className="form-item">
                        <label>描述</label>
                        <Input.TextArea 
                          rows={2}
                          placeholder="节点功能描述..."
                          value={nodeConfig.description || ''}
                          onChange={(e) => handleUpdateConfig('description', e.target.value)}
                        />
                      </div>
                    </div>

                    {/* Type-specific Config */}
                    {selectedNode.type === 'llm' && (
                      <div className="properties-section">
                        <div className="section-title">模型配置</div>
                        <div className="form-item">
                          <label>模型</label>
                          <Select 
                            value={nodeConfig.model || 'gpt-4o'}
                            onChange={(v) => handleUpdateConfig('model', v)}
                            style={{ width: '100%' }}
                            options={[
                              { value: 'gpt-4o', label: 'GPT-4o' },
                              { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
                              { value: 'claude-3-opus', label: 'Claude 3 Opus' },
                              { value: 'claude-3-sonnet', label: 'Claude 3 Sonnet' },
                            ]}
                          />
                        </div>
                        <div className="form-item">
                          <label>Temperature: {nodeConfig.temperature || 0.7}</label>
                          <Slider 
                            min={0} 
                            max={2} 
                            step={0.1}
                            value={nodeConfig.temperature || 0.7}
                            onChange={(v) => handleUpdateConfig('temperature', v)}
                          />
                        </div>
                        <div className="form-item">
                          <label>系统提示词</label>
                          <Input.TextArea 
                            rows={4}
                            placeholder="输入系统提示词..."
                            value={nodeConfig.systemPrompt || ''}
                            onChange={(e) => handleUpdateConfig('systemPrompt', e.target.value)}
                          />
                        </div>
                      </div>
                    )}

                    {selectedNode.type === 'code' && (
                      <div className="properties-section">
                        <div className="section-title">代码配置</div>
                        <div className="form-item">
                          <label>语言</label>
                          <Select 
                            value={nodeConfig.language || 'python'}
                            onChange={(v) => handleUpdateConfig('language', v)}
                            style={{ width: '100%' }}
                            options={[
                              { value: 'python', label: 'Python' },
                              { value: 'javascript', label: 'JavaScript' },
                              { value: 'typescript', label: 'TypeScript' },
                            ]}
                          />
                        </div>
                        <div className="form-item">
                          <label>代码</label>
                          <Input.TextArea 
                            rows={6}
                            placeholder="输入代码..."
                            value={nodeConfig.code || ''}
                            onChange={(e) => handleUpdateConfig('code', e.target.value)}
                            style={{ fontFamily: 'monospace', fontSize: 12 }}
                          />
                        </div>
                      </div>
                    )}

                    {selectedNode.type === 'http' && (
                      <div className="properties-section">
                        <div className="section-title">请求配置</div>
                        <div className="form-item">
                          <label>方法</label>
                          <Select 
                            value={nodeConfig.method || 'GET'}
                            onChange={(v) => handleUpdateConfig('method', v)}
                            style={{ width: '100%' }}
                            options={[
                              { value: 'GET', label: 'GET' },
                              { value: 'POST', label: 'POST' },
                              { value: 'PUT', label: 'PUT' },
                              { value: 'DELETE', label: 'DELETE' },
                            ]}
                          />
                        </div>
                        <div className="form-item">
                          <label>URL</label>
                          <Input 
                            placeholder="https://api.example.com"
                            value={nodeConfig.url || ''}
                            onChange={(e) => handleUpdateConfig('url', e.target.value)}
                          />
                        </div>
                      </div>
                    )}

                    {selectedNode.type === 'condition' && (
                      <div className="properties-section">
                        <div className="section-title">条件配置</div>
                        <div className="form-item">
                          <label>条件表达式</label>
                          <Input.TextArea 
                            rows={2}
                            placeholder="e.g., {{variable}} > 0"
                            value={nodeConfig.condition || ''}
                            onChange={(e) => handleUpdateConfig('condition', e.target.value)}
                          />
                        </div>
                      </div>
                    )}

                    {/* Advanced */}
                    <div className="properties-section">
                      <div className="section-title">高级设置</div>
                      <div className="form-item">
                        <label>超时时间 (秒)</label>
                        <Input type="number" value={nodeConfig.timeout || 60} />
                      </div>
                      <div className="form-item">
                        <label>重试次数</label>
                        <Input type="number" value={nodeConfig.retry || 0} />
                      </div>
                    </div>
                  </div>

                  {/* Footer Actions */}
                  <div className="properties-footer">
                    <Button 
                      danger 
                      icon={<DeleteOutlined />}
                      onClick={handleDeleteNode}
                    >
                      删除
                    </Button>
                    <Button type="primary" onClick={handleSave}>
                      保存配置
                    </Button>
                  </div>
                </>
              ) : (
                <div className="properties-empty">
                  <SettingOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
                  <p>选择节点以编辑配置</p>
                  <span>点击画布上的节点查看和修改其属性</span>
                </div>
              )}
            </aside>
          </div>

          {/* Status Bar */}
          <div className="workflow-statusbar">
            <div className="status-left">
              <span className="status-indicator healthy" />
              <span>就绪</span>
            </div>
            <div className="status-center">
              <span>自动保存</span>
            </div>
            <div className="status-right">
              <span>节点: {nodes.length}</span>
              <span>连线: {nodes.length - 1}</span>
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
