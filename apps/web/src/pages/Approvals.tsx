import { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Empty, Input, Skeleton, Tag, Timeline, message } from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  FileSearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import Sidebar from '../components/Sidebar'
import {
  type Artifact,
  type PipelineRunCurrent,
  type PipelineRunTimeline,
  approveCheckpoint,
  checkpointLabel,
  fetchPipelineCurrent,
  fetchPipelineRuns,
  fetchPipelineTimeline,
  formatDateTime,
  latestArtifact,
  nextActionLabel,
  parseJSON,
  rejectCheckpoint,
  runStatusMeta,
  stageLabel,
  stageStatusMeta,
} from '../lib/pipeline'

interface ApprovalItem {
  current: PipelineRunCurrent
}

const sidebarWidth = 80

export default function Approvals() {
  const [items, setItems] = useState<ApprovalItem[]>([])
  const [selectedRunId, setSelectedRunId] = useState('')
  const [timeline, setTimeline] = useState<PipelineRunTimeline | null>(null)
  const [comment, setComment] = useState('')
  const [loading, setLoading] = useState(true)
  const [loadingTimeline, setLoadingTimeline] = useState(false)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [error, setError] = useState('')

  const selected = useMemo(
    () => items.find(item => item.current.run.id === selectedRunId) ?? items[0],
    [items, selectedRunId],
  )
  const reviewArtifact = latestArtifact(timeline ?? undefined)
  const reviewText = reviewArtifact ? artifactSummary(reviewArtifact) : ''

  const loadApprovals = async () => {
    setLoading(true)
    setError('')
    try {
      const runs = await fetchPipelineRuns()
      const currents = await Promise.all(
        runs.slice(0, 30).map(run => fetchPipelineCurrent(run.id).catch(() => null)),
      )
      const pending = currents
        .filter((item): item is PipelineRunCurrent => Boolean(item?.checkpoint && item.checkpoint.status === 'pending'))
        .map(current => ({ current }))
      setItems(pending)
      setSelectedRunId(prev => pending.some(item => item.current.run.id === prev) ? prev : pending[0]?.current.run.id || '')
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载审批列表失败')
      setItems([])
    } finally {
      setLoading(false)
    }
  }

  const loadTimeline = async (runId: string) => {
    if (!runId) {
      setTimeline(null)
      return
    }
    setLoadingTimeline(true)
    try {
      setTimeline(await fetchPipelineTimeline(runId))
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载审批上下文失败')
      setTimeline(null)
    } finally {
      setLoadingTimeline(false)
    }
  }

  useEffect(() => {
    void loadApprovals()
  }, [])

  useEffect(() => {
    void loadTimeline(selectedRunId)
    setComment('')
  }, [selectedRunId])

  const decide = async (decision: 'approve' | 'reject') => {
    const checkpoint = selected?.current.checkpoint
    if (!checkpoint) return
    setActionLoading(decision)
    try {
      if (decision === 'approve') {
        await approveCheckpoint(checkpoint.id, comment)
      } else {
        await rejectCheckpoint(checkpoint.id, comment)
      }
      message.success(decision === 'approve' ? '已审批通过' : '已驳回重做')
      await loadApprovals()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '审批失败')
    } finally {
      setActionLoading(null)
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="h-screen overflow-hidden p-5 transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        <div className="mb-4 flex items-start justify-between">
          <div>
            <h1 className="m-0 text-2xl font-bold text-on-surface">人工审批</h1>
            <p className="m-0 mt-1 text-sm text-on-surface-variant">Human-in-the-Loop Checkpoints</p>
          </div>
          <Button icon={<ReloadOutlined />} onClick={loadApprovals} loading={loading}>
            刷新
          </Button>
        </div>

        {error ? <Alert className="mb-4" type="error" showIcon message={error} /> : null}

        <div className="grid h-[calc(100%-72px)] grid-cols-[340px_1fr_360px] gap-4 overflow-hidden">
          <aside className="overflow-y-auto rounded-lg border border-outline-variant bg-surface-container-lowest">
            <div className="border-b border-outline-variant px-4 py-3">
              <div className="text-sm font-semibold text-on-surface">待审批队列</div>
              <div className="text-xs text-on-surface-variant">{items.length} 个 checkpoint</div>
            </div>
            <div className="p-3">
              {loading ? <Skeleton active paragraph={{ rows: 8 }} /> : null}
              {!loading && items.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无待审批" /> : null}
              <div className="space-y-2">
                {items.map(item => {
                  const { run, checkpoint, nextAction } = item.current
                  const active = run.id === selectedRunId
                  return (
                    <button
                      key={run.id}
                      type="button"
                      onClick={() => setSelectedRunId(run.id)}
                      className={`w-full rounded-lg border p-3 text-left transition ${
                        active ? 'border-primary bg-primary/5' : 'border-outline-variant bg-white hover:border-primary/50'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <div className="truncate text-sm font-semibold text-on-surface">{run.title}</div>
                          <div className="mt-1 text-xs text-on-surface-variant">{checkpointLabel(checkpoint?.checkpointType)}</div>
                        </div>
                        <Tag color="warning">{nextActionLabel(nextAction)}</Tag>
                      </div>
                      <div className="mt-3 flex items-center justify-between text-xs text-on-surface-variant">
                        <span>{stageLabel(run.currentStageKey)}</span>
                        <span>{formatDateTime(run.updatedAt)}</span>
                      </div>
                    </button>
                  )
                })}
              </div>
            </div>
          </aside>

          <section className="min-w-0 overflow-y-auto">
            {!selected ? (
              <Card className="!h-full !rounded-lg">
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
              </Card>
            ) : (
              <div className="space-y-4">
                <Card className="!rounded-lg">
                  <div className="mb-4 flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <div className="mb-2 flex flex-wrap gap-2">
                        <Tag color={runStatusMeta(selected.current.run.status).color}>{runStatusMeta(selected.current.run.status).label}</Tag>
                        <Tag color="warning">{checkpointLabel(selected.current.checkpoint?.checkpointType)}</Tag>
                      </div>
                      <h2 className="m-0 truncate text-xl font-bold text-on-surface">{selected.current.run.title}</h2>
                      <p className="m-0 mt-2 text-sm leading-6 text-on-surface-variant">{selected.current.run.requirementText}</p>
                    </div>
                    <FileSearchOutlined className="mt-1 text-2xl text-primary" />
                  </div>
                  <div className="grid grid-cols-3 gap-3">
                    <InfoPill label="当前阶段" value={stageLabel(selected.current.stage?.stageKey)} />
                    <InfoPill label="目标分支" value={selected.current.run.targetBranch} />
                    <InfoPill label="工作分支" value={selected.current.run.workBranch} />
                  </div>
                </Card>

                <Card className="!rounded-lg" title={reviewArtifact?.title || '审批产物'}>
                  {loadingTimeline ? <Skeleton active paragraph={{ rows: 6 }} /> : null}
                  {!loadingTimeline && reviewArtifact ? (
                    <div className="rounded-lg border border-outline-variant bg-surface-container-lowest p-4">
                      <div className="mb-2 flex items-center justify-between gap-3">
                        <Tag>{reviewArtifact.artifactType}</Tag>
                        <span className="text-xs text-on-surface-variant">{formatDateTime(reviewArtifact.createdAt)}</span>
                      </div>
                      <p className="m-0 whitespace-pre-wrap text-sm leading-6 text-on-surface">{reviewText || reviewArtifact.contentText}</p>
                    </div>
                  ) : null}
                  {!loadingTimeline && !reviewArtifact ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
                </Card>

                <Card className="!rounded-lg" title="阶段状态">
                  <div className="grid grid-cols-2 gap-2">
                    {(timeline?.stages ?? []).map(stage => {
                      const meta = stageStatusMeta(stage.status)
                      return (
                        <div key={stage.id} className="flex items-center justify-between rounded-lg border border-outline-variant bg-white px-3 py-2">
                          <div className="min-w-0">
                            <div className="truncate text-sm font-semibold text-on-surface">{stageLabel(stage.stageKey)}</div>
                            <div className="text-xs text-on-surface-variant">attempt {stage.attempt}</div>
                          </div>
                          <Tag color={meta.color}>{meta.label}</Tag>
                        </div>
                      )
                    })}
                  </div>
                </Card>
              </div>
            )}
          </section>

          <aside className="overflow-y-auto">
            <Card className="!rounded-lg" title="审批操作">
              {selected?.current.checkpoint ? (
                <>
                  <div className="mb-3 rounded-lg bg-surface-container-low p-3">
                    <div className="text-xs font-semibold text-on-surface-variant">CHECKPOINT</div>
                    <div className="mt-1 text-sm font-semibold text-on-surface">{selected.current.checkpoint.id}</div>
                  </div>
                  <Input.TextArea
                    value={comment}
                    onChange={event => setComment(event.target.value)}
                    rows={5}
                    placeholder="填写审批意见"
                    className="mb-3"
                  />
                  <div className="grid grid-cols-2 gap-2">
                    <Button
                      danger
                      icon={<CloseCircleOutlined />}
                      loading={actionLoading === 'reject'}
                      onClick={() => decide('reject')}
                    >
                      驳回
                    </Button>
                    <Button
                      type="primary"
                      icon={<CheckCircleOutlined />}
                      loading={actionLoading === 'approve'}
                      onClick={() => decide('approve')}
                    >
                      通过
                    </Button>
                  </div>
                </>
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>

            <Card className="!mt-4 !rounded-lg" title="审批记录">
              <Timeline
                items={(timeline?.checkpoints ?? []).map(checkpoint => ({
                  key: checkpoint.id,
                  color: checkpoint.status === 'approved' ? 'green' : checkpoint.status === 'rejected' ? 'red' : 'orange',
                  children: (
                    <div>
                      <div className="text-sm font-semibold text-on-surface">{checkpointLabel(checkpoint.checkpointType)}</div>
                      <div className="text-xs text-on-surface-variant">{checkpoint.status} · {checkpoint.decision || 'pending'}</div>
                      {checkpoint.comment ? <div className="mt-1 rounded bg-surface-container-low p-2 text-xs text-on-surface-variant">{checkpoint.comment}</div> : null}
                    </div>
                  ),
                }))}
              />
              {timeline?.checkpoints.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
            </Card>
          </aside>
        </div>
      </main>
    </div>
  )
}

function InfoPill({ label, value }: { label: string; value?: string }) {
  return (
    <div className="rounded-lg border border-outline-variant bg-white p-3">
      <div className="text-xs font-semibold text-on-surface-variant">{label}</div>
      <div className="mt-1 truncate text-sm font-semibold text-on-surface">{value || '-'}</div>
    </div>
  )
}

function artifactSummary(artifact: Artifact): string {
  const payload = parseJSON<Record<string, unknown>>(artifact.contentJson)
  const summary = payload && typeof payload.summary === 'string' ? payload.summary : ''
  return summary || artifact.contentText || ''
}
