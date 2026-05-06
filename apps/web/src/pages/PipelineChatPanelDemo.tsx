import { useLayoutEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'

/** 兼容旧链接：跳转到工作台并在工作流内打开 Mock（`?chatDemo=1`） */
export default function PipelineChatPanelDemo() {
  const navigate = useNavigate()
  useLayoutEffect(() => {
    navigate({ to: '/workflows', search: { chatDemo: '1' }, replace: true })
  }, [navigate])
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-100 text-sm text-slate-500">
      正在进入工作台预览…
    </div>
  )
}
