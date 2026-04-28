import { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Empty, Skeleton, Space, Tag, message } from 'antd'
import {
  BranchesOutlined,
  CopyOutlined,
  FileTextOutlined,
  LinkOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import Sidebar from '../components/Sidebar'
import {
  type GitDelivery,
  type PipelineRunTimeline,
  deliveryStatusMeta,
  fetchGitDelivery,
  fetchPipelineDeliveries,
  fetchPipelineRuns,
  fetchPipelineTimeline,
  formatDateTime,
  formatDuration,
  parseJSON,
  runStatusMeta,
  stageLabel,
  stageStatusMeta,
} from '../lib/pipeline'

const sidebarWidth = 80

export default function Delivery() {
  const [deliveries, setDeliveries] = useState<GitDelivery[]>([])
  const [selectedDeliveryId, setSelectedDeliveryId] = useState('')
  const [selectedDelivery, setSelectedDelivery] = useState<GitDelivery | null>(null)
  const [timeline, setTimeline] = useState<PipelineRunTimeline | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingDetail, setLoadingDetail] = useState(false)
  const [error, setError] = useState('')

  const changedFiles = useMemo(
    () => parseJSON<string[]>(selectedDelivery?.changedFilesJson) ?? [],
    [selectedDelivery?.changedFilesJson],
  )
  const validation = useMemo(
    () => parseJSON<string[]>(selectedDelivery?.validationJson) ?? [],
    [selectedDelivery?.validationJson],
  )
  const readyCount = deliveries.filter(item => item.status === 'ready').length

  const loadDeliveries = async () => {
    setLoading(true)
    setError('')
    try {
      const runs = await fetchPipelineRuns()
      const lists = await Promise.all(
        runs.slice(0, 30).map(run => fetchPipelineDeliveries(run.id).catch(() => [])),
      )
      const items = lists.flat().sort((left, right) => Date.parse(right.createdAt) - Date.parse(left.createdAt))
      setDeliveries(items)
      setSelectedDeliveryId(prev => items.some(item => item.id === prev) ? prev : items[0]?.id || '')
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载交付记录失败')
      setDeliveries([])
    } finally {
      setLoading(false)
    }
  }

  const loadDetail = async (deliveryId: string) => {
    if (!deliveryId) {
      setSelectedDelivery(null)
      setTimeline(null)
      return
    }
    setLoadingDetail(true)
    setError('')
    try {
      const delivery = await fetchGitDelivery(deliveryId)
      setSelectedDelivery(delivery)
      setTimeline(await fetchPipelineTimeline(delivery.pipelineRunId))
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载交付详情失败')
      setSelectedDelivery(null)
      setTimeline(null)
    } finally {
      setLoadingDetail(false)
    }
  }

  useEffect(() => {
    void loadDeliveries()
  }, [])

  useEffect(() => {
    void loadDetail(selectedDeliveryId)
  }, [selectedDeliveryId])

  const copySummary = async () => {
    if (!selectedDelivery?.summaryMarkdown) return
    try {
      await navigator.clipboard.writeText(selectedDelivery.summaryMarkdown)
      message.success('已复制交付摘要')
    } catch {
      message.error('复制失败')
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="h-screen overflow-hidden p-5 transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        <div className="mb-4 flex items-start justify-between">
          <div>
            <h1 className="m-0 text-2xl font-bold text-on-surface">交付审查</h1>
            <p className="m-0 mt-1 text-sm text-on-surface-variant">GitDelivery Drafts</p>
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={loadDeliveries} loading={loading}>刷新</Button>
            <Button icon={<CopyOutlined />} onClick={copySummary} disabled={!selectedDelivery?.summaryMarkdown}>复制摘要</Button>
            <Button
              type="primary"
              icon={<LinkOutlined />}
              disabled={!selectedDelivery?.prmrUrl}
              onClick={() => selectedDelivery?.prmrUrl && window.open(selectedDelivery.prmrUrl, '_blank')}
            >
              打开 PR/MR
            </Button>
          </Space>
        </div>

        {error ? <Alert className="mb-4" type="error" showIcon message={error} /> : null}

        <div className="grid h-[calc(100%-72px)] grid-cols-[340px_1fr] gap-4 overflow-hidden">
          <aside className="overflow-y-auto rounded-lg border border-outline-variant bg-surface-container-lowest">
            <div className="border-b border-outline-variant px-4 py-3">
              <div className="text-sm font-semibold text-on-surface">交付草稿</div>
              <div className="text-xs text-on-surface-variant">{deliveries.length} 条记录 · {readyCount} 条可审查</div>
            </div>
            <div className="p-3">
              {loading ? <Skeleton active paragraph={{ rows: 8 }} /> : null}
              {!loading && deliveries.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无交付记录" /> : null}
              <div className="space-y-2">
                {deliveries.map(delivery => {
                  const meta = deliveryStatusMeta(delivery.status)
                  const active = delivery.id === selectedDeliveryId
                  return (
                    <button
                      key={delivery.id}
                      type="button"
                      onClick={() => setSelectedDeliveryId(delivery.id)}
                      className={`w-full rounded-lg border p-3 text-left transition ${
                        active ? 'border-primary bg-primary/5' : 'border-outline-variant bg-white hover:border-primary/50'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <div className="truncate text-sm font-semibold text-on-surface">{delivery.prmrTitle || delivery.id}</div>
                          <div className="mt-1 truncate text-xs text-on-surface-variant">{delivery.repo} · {delivery.baseBranch}</div>
                        </div>
                        <Tag color={meta.color}>{meta.label}</Tag>
                      </div>
                      <div className="mt-3 flex items-center justify-between text-xs text-on-surface-variant">
                        <span>{delivery.provider}</span>
                        <span>{formatDateTime(delivery.createdAt)}</span>
                      </div>
                    </button>
                  )
                })}
              </div>
            </div>
          </aside>

          <section className="min-w-0 overflow-y-auto">
            {loadingDetail ? <Skeleton active paragraph={{ rows: 12 }} /> : null}
            {!loadingDetail && !selectedDelivery ? (
              <Card className="!h-full !rounded-lg">
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
              </Card>
            ) : null}
            {!loadingDetail && selectedDelivery ? (
              <div className="space-y-4">
                <div className="grid grid-cols-4 gap-3">
                  <Metric title="交付状态" value={deliveryStatusMeta(selectedDelivery.status).label} />
                  <Metric title="Pipeline" value={runStatusMeta(timeline?.run.status).label} />
                  <Metric title="运行耗时" value={formatDuration(timeline?.summary.durationMs)} />
                  <Metric title="变更文件" value={`${changedFiles.length}`} />
                </div>

                <Card className="!rounded-lg">
                  <div className="mb-4 flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <div className="mb-2 flex flex-wrap gap-2">
                        <Tag color={deliveryStatusMeta(selectedDelivery.status).color}>{deliveryStatusMeta(selectedDelivery.status).label}</Tag>
                        <Tag color="blue">{selectedDelivery.provider}</Tag>
                        {!selectedDelivery.prmrUrl ? <Tag>本地草稿</Tag> : null}
                      </div>
                      <h2 className="m-0 break-words text-xl font-bold text-on-surface">{selectedDelivery.prmrTitle || selectedDelivery.id}</h2>
                      <div className="mt-3 flex flex-wrap gap-2 text-sm text-on-surface-variant">
                        <span><BranchesOutlined /> {selectedDelivery.baseBranch}</span>
                        <span>→</span>
                        <span>{selectedDelivery.headBranch}</span>
                      </div>
                    </div>
                    <FileTextOutlined className="mt-1 text-2xl text-primary" />
                  </div>
                  <div className="rounded-lg border border-outline-variant bg-surface-container-lowest p-4">
                    <pre className="m-0 whitespace-pre-wrap break-words text-sm leading-6 text-on-surface">{selectedDelivery.prmrBody || selectedDelivery.summaryMarkdown || '暂无正文'}</pre>
                  </div>
                </Card>

                <div className="grid grid-cols-[1fr_360px] gap-4">
                  <Card className="!rounded-lg" title="变更文件">
                    <div className="space-y-2">
                      {changedFiles.map(file => (
                        <div key={file} className="rounded-lg border border-outline-variant bg-white px-3 py-2 font-mono text-xs text-on-surface">
                          {file}
                        </div>
                      ))}
                      {changedFiles.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
                    </div>
                  </Card>

                  <Card className="!rounded-lg" title="验证摘要">
                    <div className="space-y-2">
                      {validation.map(item => (
                        <div key={item} className="rounded-lg border border-outline-variant bg-white px-3 py-2 text-sm text-on-surface">
                          {item}
                        </div>
                      ))}
                      {validation.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
                    </div>
                  </Card>
                </div>

                <Card className="!rounded-lg" title="Pipeline 阶段">
                  <div className="grid grid-cols-4 gap-2">
                    {(timeline?.stages ?? []).map(stage => {
                      const meta = stageStatusMeta(stage.status)
                      return (
                        <div key={stage.id} className="rounded-lg border border-outline-variant bg-white p-3">
                          <div className="truncate text-sm font-semibold text-on-surface">{stageLabel(stage.stageKey)}</div>
                          <div className="mt-2 flex items-center justify-between">
                            <span className="text-xs text-on-surface-variant">attempt {stage.attempt}</span>
                            <Tag color={meta.color}>{meta.label}</Tag>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </Card>
              </div>
            ) : null}
          </section>
        </div>
      </main>
    </div>
  )
}

function Metric({ title, value }: { title: string; value: string }) {
  return (
    <div className="rounded-lg border border-outline-variant bg-white px-4 py-3">
      <div className="text-xs font-semibold text-on-surface-variant">{title}</div>
      <div className="mt-1 truncate text-xl font-bold text-on-surface">{value}</div>
    </div>
  )
}
