import { useEffect, useMemo, useRef, useState } from 'react'
import { Alert, Button, Card, Empty, Progress, Skeleton, Space, Tag, Tooltip, message } from 'antd'
import {
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  StopOutlined,
  SyncOutlined,
} from '@ant-design/icons'
import Sidebar from '../components/Sidebar'
import {
  type AgentRun,
  type Artifact,
  type PipelineRun,
  type PipelineRunTimeline,
  fetchPipelineRuns,
  fetchPipelineTimeline,
  formatDateTime,
  formatDuration,
  isLiveRun,
  latestArtifact,
  nextActionLabel,
  pausePipelineRun,
  resumePipelineRun,
  runStatusMeta,
  stageLabel,
  stageStatusMeta,
  startPipelineRun,
  terminatePipelineRun,
} from '../lib/pipeline'

const sidebarWidth = 80

export default function Workflows() {
  const [runs, setRuns] = useState<PipelineRun[]>([])
  const [selectedRunId, setSelectedRunId] = useState('')
  const [timeline, setTimeline] = useState<PipelineRunTimeline | null>(null)
  const [loadingRuns, setLoadingRuns] = useState(true)
  const [loadingTimeline, setLoadingTimeline] = useState(false)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [error, setError] = useState('')
  const redirectedRunIdRef = useRef('')

  const selectedRun = useMemo(
    () => runs.find(run => run.id === selectedRunId) ?? timeline?.run,
    [runs, selectedRunId, timeline],
  )
  const currentArtifact = latestArtifact(timeline ?? undefined)
  const completionPercent = timeline?.summary.totalStages
    ? Math.round((timeline.summary.completedStages / timeline.summary.totalStages) * 100)
    : 0
  const recentAgentRuns = useMemo(
    () => [...(timeline?.agentRuns ?? [])].slice(-4).reverse(),
    [timeline?.agentRuns],
  )

  const loadRuns = async () => {
    setLoadingRuns(true)
    setError('')
    try {
      const items = await fetchPipelineRuns()
      setRuns(items)
      setSelectedRunId(prev => prev || items[0]?.id || '')
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载流水线失败')
      setRuns([])
    } finally {
      setLoadingRuns(false)
    }
  }

  const loadTimeline = async (runId: string) => {
    if (!runId) {
      setTimeline(null)
      return
    }
    setLoadingTimeline(true)
    setError('')
    try {
      setTimeline(await fetchPipelineTimeline(runId))
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载工作台失败')
      setTimeline(null)
    } finally {
      setLoadingTimeline(false)
    }
  }

  useEffect(() => {
    void loadRuns()
  }, [])

  useEffect(() => {
    void loadTimeline(selectedRunId)
  }, [selectedRunId])

  useEffect(() => {
    if (!timeline || !isLiveRun(timeline.run.status)) return
    const timer = window.setInterval(() => {
      void loadTimeline(timeline.run.id)
    }, 8000)
    return () => window.clearInterval(timer)
  }, [timeline?.run.id, timeline?.run.status])

  useEffect(() => {
    if (!timeline?.run?.id) return
    if (timeline.run.status !== 'waiting_approval') {
      if (redirectedRunIdRef.current === timeline.run.id) {
        redirectedRunIdRef.current = ''
      }
      return
    }
    if (redirectedRunIdRef.current === timeline.run.id) return
    redirectedRunIdRef.current = timeline.run.id
    window.location.assign(`/approvals/${encodeURIComponent(timeline.run.id)}`)
  }, [timeline?.run.id, timeline?.run.status])

  const refreshAll = async () => {
    await loadRuns()
    if (selectedRunId) await loadTimeline(selectedRunId)
  }

  const runAction = async (action: string, fn: () => Promise<void>) => {
    setActionLoading(action)
    try {
      await fn()
      message.success('操作已提交')
      await refreshAll()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '操作失败')
    } finally {
      setActionLoading(null)
    }
  }

  const renderRunActions = () => {
    if (!selectedRun) return null
    const status = selectedRun.status
    return (
      <Space>
        {status === 'draft' || status === 'failed' ? (
          <Tooltip title="启动">
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              loading={actionLoading === 'start'}
              onClick={() => runAction('start', () => startPipelineRun(selectedRun.id))}
            />
          </Tooltip>
        ) : null}
        {status === 'queued' || status === 'running' ? (
          <Tooltip title="暂停">
            <Button
              icon={<PauseCircleOutlined />}
              loading={actionLoading === 'pause'}
              onClick={() => runAction('pause', () => pausePipelineRun(selectedRun.id))}
            />
          </Tooltip>
        ) : null}
        {status === 'paused' || status === 'failed' ? (
          <Tooltip title="恢复">
            <Button
              icon={<PlayCircleOutlined />}
              loading={actionLoading === 'resume'}
              onClick={() => runAction('resume', () => resumePipelineRun(selectedRun.id))}
            />
          </Tooltip>
        ) : null}
        {['draft', 'queued', 'running', 'waiting_approval', 'paused', 'failed'].includes(status) ? (
          <Tooltip title="终止">
            <Button
              danger
              icon={<StopOutlined />}
              loading={actionLoading === 'terminate'}
              onClick={() => runAction('terminate', () => terminatePipelineRun(selectedRun.id))}
            />
          </Tooltip>
        ) : null}
        <Tooltip title="刷新">
          <Button icon={<ReloadOutlined />} onClick={refreshAll} loading={loadingTimeline} />
        </Tooltip>
      </Space>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <Sidebar />
      <main className="h-screen overflow-hidden p-5 transition-all duration-300" style={{ marginLeft: `${sidebarWidth}px` }}>
        <div className="flex h-full gap-4">
          <aside className="w-80 shrink-0 overflow-y-auto rounded-lg border border-outline-variant bg-surface-container-lowest">
            <div className="sticky top-0 z-10 flex items-center justify-between border-b border-outline-variant bg-surface-container-lowest px-4 py-3">
              <div>
                <h1 className="m-0 text-lg font-bold text-on-surface">Pipeline 工作台</h1>
                <div className="text-xs text-on-surface-variant">{runs.length} 个运行记录</div>
              </div>
              <Button type="text" icon={<ReloadOutlined />} onClick={refreshAll} loading={loadingRuns} />
            </div>
            <div className="p-3">
              {loadingRuns ? <Skeleton active paragraph={{ rows: 8 }} /> : null}
              {!loadingRuns && runs.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
              <div className="space-y-2">
                {runs.map(run => {
                  const meta = runStatusMeta(run.status)
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
                          <div className="mt-1 truncate text-xs text-on-surface-variant">{run.targetRepo} · {run.targetBranch}</div>
                        </div>
                        <Tag color={meta.color}>{meta.label}</Tag>
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

          <section className="flex min-w-0 flex-1 flex-col overflow-hidden">
            <div className="mb-4 flex items-start justify-between gap-4">
              <div className="min-w-0">
                <div className="mb-1 flex items-center gap-2">
                  {selectedRun ? <Tag color={runStatusMeta(selectedRun.status).color}>{runStatusMeta(selectedRun.status).label}</Tag> : null}
                  {timeline?.current?.nextAction ? <Tag color="blue">{nextActionLabel(timeline.current.nextAction)}</Tag> : null}
                </div>
                <h2 className="m-0 truncate text-2xl font-bold text-on-surface">{selectedRun?.title || '选择 PipelineRun'}</h2>
                <p className="m-0 mt-1 truncate text-sm text-on-surface-variant">{selectedRun?.requirementText || '暂无运行记录'}</p>
              </div>
              {renderRunActions()}
            </div>

            {error ? <Alert className="mb-4" type="error" showIcon message={error} /> : null}

            <div className="grid grid-cols-4 gap-3">
              <Metric title="完成阶段" value={`${timeline?.summary.completedStages ?? 0}/${timeline?.summary.totalStages ?? 0}`} />
              <Metric title="当前阶段" value={stageLabel(timeline?.summary.currentStageKey)} />
              <Metric title="运行耗时" value={formatDuration(timeline?.summary.durationMs)} />
              <Metric title="AgentRun" value={`${timeline?.agentRuns.length ?? 0}`} />
            </div>

            <div className="mt-4 grid min-h-0 flex-1 grid-cols-[1fr_360px] gap-4 overflow-hidden">
              <div className="min-w-0 overflow-y-auto">
                <Card className="!rounded-lg">
                  <div className="mb-3 flex items-center justify-between">
                    <span className="font-semibold text-on-surface">阶段进度</span>
                    <span className="text-sm text-on-surface-variant">{completionPercent}%</span>
                  </div>
                  <Progress percent={completionPercent} showInfo={false} />
                  <div className="mt-4 grid grid-cols-1 gap-2">
                    {(timeline?.stages ?? []).map(stage => (
                      <StageRow key={stage.id} stage={stage} active={stage.stageKey === timeline?.run.currentStageKey} />
                    ))}
                    {!timeline && !loadingTimeline ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
                    {loadingTimeline ? <Skeleton active paragraph={{ rows: 6 }} /> : null}
                  </div>
                </Card>

                <Card className="!mt-4 !rounded-lg" title="阶段产物">
                  <div className="space-y-2">
                    {(timeline?.artifacts ?? []).slice(-5).reverse().map(artifact => (
                      <ArtifactItem key={artifact.id} artifact={artifact} />
                    ))}
                    {timeline?.artifacts.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
                  </div>
                </Card>
              </div>

              <aside className="min-w-0 overflow-y-auto">
                <Card className="!rounded-lg" title="当前上下文">
                  <div className="space-y-4">
                    <Field label="当前阶段" value={stageLabel(timeline?.current?.stage?.stageKey)} />
                    <Field label="下一动作" value={nextActionLabel(timeline?.current?.nextAction)} />
                    <Field label="最新产物" value={currentArtifact?.title || '-'} />
                    <Field label="交付草稿" value={timeline?.current?.delivery?.prmrTitle || '-'} />
                  </div>
                </Card>

                <Card className="!mt-4 !rounded-lg" title="Agent 观测">
                  <div className="space-y-3">
                    {recentAgentRuns.map(agentRun => <AgentRunItem key={agentRun.id} agentRun={agentRun} />)}
                    {recentAgentRuns.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /> : null}
                  </div>
                </Card>
              </aside>
            </div>
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

function StageRow({ stage, active }: { stage: PipelineRunTimeline['stages'][number]; active: boolean }) {
  const meta = stageStatusMeta(stage.status)
  return (
    <div className={`flex items-center gap-3 rounded-lg border px-3 py-2 ${active ? 'border-primary bg-primary/5' : 'border-outline-variant bg-white'}`}>
      <div className={`flex h-8 w-8 items-center justify-center rounded-full ${stage.status === 'succeeded' ? 'bg-green-50 text-green-600' : 'bg-surface-container-high text-on-surface-variant'}`}>
        {stage.status === 'running' ? <SyncOutlined spin /> : stage.attempt}
      </div>
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-semibold text-on-surface">{stageLabel(stage.stageKey)}</div>
        <div className="truncate text-xs text-on-surface-variant">{stage.stageKey} · attempt {stage.attempt}</div>
      </div>
      <Tag color={meta.color}>{meta.label}</Tag>
    </div>
  )
}

function ArtifactItem({ artifact }: { artifact: Artifact }) {
  return (
    <div className="rounded-lg border border-outline-variant bg-white p-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold text-on-surface">{artifact.title}</div>
          <div className="mt-1 line-clamp-2 text-xs text-on-surface-variant">{artifact.contentText || artifact.artifactType}</div>
        </div>
        <Tag>{artifact.artifactType}</Tag>
      </div>
    </div>
  )
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-xs font-semibold text-on-surface-variant">{label}</div>
      <div className="mt-1 break-words text-sm font-medium text-on-surface">{value}</div>
    </div>
  )
}

function AgentRunItem({ agentRun }: { agentRun: AgentRun }) {
  const ok = agentRun.status === 'succeeded'
  return (
    <div className="rounded-lg border border-outline-variant bg-white p-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold text-on-surface">{agentRun.agentKey}</div>
          <div className="mt-1 truncate text-xs text-on-surface-variant">{agentRun.provider || '-'} · {agentRun.model || '-'}</div>
        </div>
        <Tag color={ok ? 'success' : 'error'}>{ok ? '成功' : '失败'}</Tag>
      </div>
      <div className="mt-2 text-xs text-on-surface-variant">{formatDuration(agentRun.latencyMs)}</div>
    </div>
  )
}
