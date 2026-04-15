import { useEffect, useRef } from 'react'
import { Graph } from '@antv/x6'
import { Card, Button, Slider, Select, Avatar, Tag } from 'antd'
import {
  ZoomInOutlined,
  ZoomOutOutlined,
  CaretRightFilled,
  FileTextOutlined,
  StarOutlined,
  ThunderboltFilled,
  CloseOutlined,
  PlusOutlined,
  CheckOutlined,
  BugOutlined,
} from '@ant-design/icons'
import TopNav from '../components/TopNav'
import Sidebar from '../components/Sidebar'

function createNodeHTML(title: string, content: string, icon: string, extra?: string) {
  return `
    <div class="workflow-node" style="width:220px;background:#fff;border-radius:12px;padding:14px;box-shadow:0 2px 8px rgba(0,0,0,0.06);border:1px solid #e5e9ef;">
      <div style="display:flex;align-items:flex-start;gap:10px;margin-bottom:10px;">
        <div style="width:28px;height:28px;border-radius:6px;background:#e8f2fc;color:#0066ff;display:flex;align-items:center;justify-content:center;flex-shrink:0;font-size:14px;">${icon}</div>
        <div style="flex:1;min-width:0;">
          <div style="font-size:13px;font-weight:600;color:#1a1a2e;">${title}</div>
          <div style="font-size:11px;color:#8b95a8;line-height:1.4;margin-top:2px;">${content}</div>
        </div>
        <button style="width:20px;height:20px;border:none;background:transparent;color:#8b95a8;cursor:pointer;font-size:14px;">⋯</button>
      </div>
      ${extra || ''}
    </div>
  `
}

export default function Workflows() {
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!containerRef.current) return

    const graph = new Graph({
      container: containerRef.current,
      width: containerRef.current.clientWidth,
      height: containerRef.current.clientHeight,
      background: { color: '#f8fafc' },
      grid: { visible: true, type: 'dot', args: { color: '#e5e9ef', thickness: 1 } },
      panning: true,
      mousewheel: { enabled: true, modifiers: ['ctrl', 'meta'] },
    })

    graph.addNode({
      shape: 'html',
      x: 60,
      y: 180,
      width: 220,
      height: 100,
      html: () => {
        const div = document.createElement('div')
        div.innerHTML = createNodeHTML(
          'Requirement Analysis',
          'Analyzing and generating specifications using LLM.',
          '📄'
        )
        return div
      },
    })

    graph.addNode({
      shape: 'html',
      x: 360,
      y: 180,
      width: 220,
      height: 120,
      html: () => {
        const div = document.createElement('div')
        div.innerHTML = createNodeHTML(
          'AI Architect',
          'Generating system diagrams...',
          '⭐',
          `<div style="display:flex;align-items:center;gap:8px;margin-bottom:8px;">
            <div style="flex:1;height:4px;background:#e5e9ef;border-radius:2px;overflow:hidden;">
              <div style="width:65%;height:100%;background:#0066ff;border-radius:2px;"></div>
            </div>
            <span style="font-size:11px;font-weight:600;color:#0066ff;">65%</span>
          </div>
          <div style="font-size:11px;color:#5a6478;">Generating system diagrams...</div>`
        )
        return div
      },
    })

    graph.addNode({
      shape: 'html',
      x: 660,
      y: 180,
      width: 220,
      height: 110,
      html: () => {
        const div = document.createElement('div')
        div.innerHTML = createNodeHTML(
          'Automated Review',
          '',
          '</>',
          `<div style="display:flex;flex-direction:column;gap:4px;">
            <div style="display:flex;justify-content:space-between;font-size:11px;color:#5a6478;">
              <span>main.py</span>
              <span style="font-size:10px;font-weight:500;padding:1px 6px;border-radius:4px;color:#10b981;background:#d1fae5;">Ready</span>
            </div>
            <div style="display:flex;justify-content:space-between;font-size:11px;color:#5a6478;">
              <span>api.ts</span>
              <span style="font-size:10px;font-weight:500;padding:1px 6px;border-radius:4px;color:#8b95a8;background:#f0f4f8;">Queued</span>
            </div>
          </div>`
        )
        return div
      },
    })

    graph.addEdge({
      source: { x: 280, y: 230 },
      target: { x: 360, y: 230 },
      attrs: { line: { stroke: '#dbe4ee', strokeWidth: 2, targetMarker: { name: 'circle', r: 4, fill: '#0066ff', stroke: '#0066ff' } } },
    })

    graph.addEdge({
      source: { x: 580, y: 230 },
      target: { x: 660, y: 230 },
      attrs: { line: { stroke: '#dbe4ee', strokeWidth: 2, targetMarker: { name: 'circle', r: 4, fill: '#0066ff', stroke: '#0066ff' } } },
    })

    const handleResize = () => {
      graph.resize(containerRef.current!.clientWidth, containerRef.current!.clientHeight)
    }
    window.addEventListener('resize', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
      graph.dispose()
    }
  }, [])

  return (
    <div className="app-container">
      <TopNav showSearch />
      <div className="main-layout">
        <Sidebar />
        <main className="workflow-page">
          {/* Toolbar */}
          <div className="workflow-toolbar">
            <div className="toolbar-left">
              <Button icon={<ZoomOutOutlined />} className="toolbar-icon-btn" />
              <Button icon={<ZoomInOutlined />} className="toolbar-icon-btn" />
              <span className="zoom-level">100%</span>
            </div>
            <div className="toolbar-center">
              <span className="pipeline-name">Alpha Release Pipeline</span>
              <Tag className="pipeline-badge">DRAFT</Tag>
            </div>
            <div className="toolbar-right">
              <Button type="primary" icon={<CaretRightFilled />} className="run-workflow-btn">
                Run Workflow
              </Button>
            </div>
          </div>

          {/* Canvas + Properties */}
          <div className="workflow-body">
            <div className="workflow-canvas" ref={containerRef} />

            {/* Node Properties Panel */}
            <aside className="node-properties">
              <div className="properties-header">
                <h3>Node Properties</h3>
                <Button type="text" icon={<CloseOutlined />} className="close-btn" />
              </div>

              <div className="properties-section">
                <div className="properties-node-info">
                  <Avatar size="small" icon={<StarOutlined />} style={{ background: '#e8f2fc', color: '#0066ff' }} />
                  <div>
                    <div className="properties-node-title">AI Architect</div>
                    <div className="properties-node-id">ID: NODE-482-AI</div>
                  </div>
                </div>
              </div>

              <div className="properties-section">
                <label className="properties-label">MODEL SELECTION</label>
                <Select defaultValue="gpt4o" style={{ width: '100%' }}>
                  <Select.Option value="gpt4o">GPT-4o Vision High</Select.Option>
                </Select>
              </div>

              <div className="properties-section">
                <label className="properties-label">TEMPERATURE</label>
                <div className="slider-row">
                  <Slider defaultValue={70} style={{ flex: 1 }} />
                  <span className="slider-value">0.7</span>
                </div>
              </div>

              <div className="properties-section">
                <label className="properties-label">CONTEXT SOURCES</label>
                <div className="context-chips">
                  <Tag icon={<FileTextOutlined />} color="blue">spec_v2.md</Tag>
                  <Tag icon={<FileTextOutlined />} color="blue">db_schema.json</Tag>
                  <Button shape="circle" size="small" icon={<PlusOutlined />} className="add-context-btn" />
                </div>
              </div>

              <Card className="ai-insight-card" bordered={false}>
                <div className="ai-insight-header">
                  <ThunderboltFilled style={{ color: '#c2410c' }} />
                  AI INSIGHTS
                </div>
                <p className="ai-insight-text">
                  This node is currently experiencing higher latency due to the complexity of the current schema. Consider splitting into two sub-processes.
                </p>
              </Card>

              <div className="properties-footer">
                <Button className="cancel-btn">Cancel</Button>
                <Button type="primary" className="save-btn">Save Node</Button>
              </div>
            </aside>
          </div>

          {/* Bottom Status Bar */}
          <div className="workflow-statusbar">
            <div className="status-left">
              <span className="status-indicator healthy" />
              <span>System Healthy: 22ms</span>
            </div>
            <div className="status-center">
              <CheckOutlined />
              Changes saved automatically
            </div>
            <div className="status-right">
              <span><BugOutlined /> Debug Console</span>
              <span>v4.2.0-beta</span>
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}
