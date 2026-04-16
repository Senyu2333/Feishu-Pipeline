import { useState, useEffect } from 'react'

// TS 后端 API 地址
const API_BASE = 'http://localhost:3001'
const USER_TOKEN_KEY = 'feishu_user_token'
const USER_TOKEN_EXPIRES_KEY = 'feishu_user_token_expires'

interface ApiResult {
  success?: boolean
  data?: unknown
  error?: string
  code?: number
  msg?: string
  message?: string
}

export default function Debug() {
  const [results, setResults] = useState<{ api: string; result: ApiResult; time: string }[]>([])
  const [loading, setLoading] = useState(false)

  // Token 状态
  const [userToken, setUserToken] = useState(() => localStorage.getItem(USER_TOKEN_KEY) || '')
  const [tokenExpires, setTokenExpires] = useState(() => localStorage.getItem(USER_TOKEN_EXPIRES_KEY) || '')
  const [folderToken, setFolderToken] = useState('')
  const [documentId, setDocumentId] = useState('')
  const [manualCode, setManualCode] = useState('')

  // 检查 URL 中的 token 参数（OAuth 回调）
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const token = params.get('token')
    const expiresAt = params.get('expires_at')
    const error = params.get('error')

    if (token) {
      setUserToken(token)
      localStorage.setItem(USER_TOKEN_KEY, token)
      if (expiresAt) {
        setTokenExpires(expiresAt)
        localStorage.setItem(USER_TOKEN_EXPIRES_KEY, expiresAt)
      }
      // 清除 URL 参数
      window.history.replaceState({}, '', '/debug')
      addResult('OAuth 回调', { success: true, data: { message: 'Token 已获取并保存' } }, '自动')
    } else if (error) {
      addResult('OAuth 回调', { success: false, error: `错误: ${error}` }, '自动')
      window.history.replaceState({}, '', '/debug')
    }
  }, [])

  const addResult = (api: string, result: ApiResult, time: string) => {
    setResults(prev => [...prev, { api, result, time }])
  }

  // 获取飞书 OAuth URL
  const getOAuthUrl = async () => {
    setLoading(true)
    const start = Date.now()
    try {
      const res = await fetch(`${API_BASE}/api/feishu/oauth-url`)
      const data = await res.json()
      addResult('/api/feishu/oauth-url', data, `${Date.now() - start}ms`)

      if (data.success && data.data?.oauthUrl) {
        // 跳转到飞书授权页面
        window.location.href = data.data.oauthUrl
      }
    } catch (err) {
      addResult('/api/feishu/oauth-url', { error: String(err) }, `${Date.now() - start}ms`)
    }
    setLoading(false)
  }

  // 手动兑换 token
  const exchangeToken = async () => {
    if (!manualCode.trim()) {
      alert('请输入 code')
      return
    }
    setLoading(true)
    const start = Date.now()
    try {
      const res = await fetch(`${API_BASE}/api/feishu/exchange-token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: manualCode }),
      })
      const data = await res.json()
      addResult('/api/feishu/exchange-token', data, `${Date.now() - start}ms`)

      if (data.success && data.data?.user_token) {
        setUserToken(data.data.user_token)
        setTokenExpires(data.data.expires_at || '')
        localStorage.setItem(USER_TOKEN_KEY, data.data.user_token)
        if (data.data.expires_at) {
          localStorage.setItem(USER_TOKEN_EXPIRES_KEY, data.data.expires_at)
        }
        setManualCode('')
      }
    } catch (err) {
      addResult('/api/feishu/exchange-token', { error: String(err) }, `${Date.now() - start}ms`)
    }
    setLoading(false)
  }

  // 调用 POST API
  const callApi = async (api: string, body: object) => {
    setLoading(true)
    const start = Date.now()
    try {
      const res = await fetch(`${API_BASE}${api}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      const data = await res.json()
      addResult(api, data, `${Date.now() - start}ms`)
    } catch (err) {
      addResult(api, { error: String(err) }, `${Date.now() - start}ms`)
    }
    setLoading(false)
  }

  // 调用 GET API
  const callGetApi = async (api: string, params: Record<string, string>) => {
    setLoading(true)
    const start = Date.now()
    try {
      const query = new URLSearchParams(params).toString()
      const res = await fetch(`${API_BASE}${api}?${query}`)
      const data = await res.json()
      addResult(api + '?' + query, data, `${Date.now() - start}ms`)
    } catch (err) {
      addResult(api, { error: String(err) }, `${Date.now() - start}ms`)
    }
    setLoading(false)
  }

  // 测试列表
  const [docTitle, setDocTitle] = useState('测试文档-' + new Date().toLocaleString())
  
  const tests = [
    {
      name: '健康检查',
      api: '/api/health2',
      body: {},
    },
    {
      name: '获取根目录文件',
      api: '/api/feishu/list-files',
      body: { order_by: 'EditedTime', direction: 'DESC' },
    },
    {
      name: '获取文件夹文件',
      api: '/api/feishu/list-files',
      body: { folder_token: folderToken, page_size: 20 },
    },
    {
      name: '创建文档',
      api: '/api/feishu/create-document',
      body: { folder_token: folderToken || undefined, title: docTitle },
    },
  ]

  // 模拟创建完整文档
  const [createdDocUrl, setCreatedDocUrl] = useState('')
  const createFullDocument = async () => {
    if (!userToken) {
      alert('请先获取 Token')
      return
    }
    setLoading(true)
    const start = Date.now()

    try {
      // 1. 创建文档
      const title = 'AI 生成报告 - ' + new Date().toLocaleString()
      addResult('1. 创建文档', { message: '开始创建文档...' }, '0ms')

      const createRes = await fetch(`${API_BASE}/api/feishu/create-document`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, user_token: userToken }),
      })
      const createData = await createRes.json()
      addResult('1. 创建文档', createData, `${Date.now() - start}ms`)

      if (!createData.success || !createData.data?.data?.document?.document_id) {
        alert('创建文档失败')
        setLoading(false)
        return
      }

      const documentId = createData.data.data.document.document_id
      const docUrl = `https://feishu.cn/docx/${documentId}`
      setCreatedDocUrl(docUrl)

      // 等待一小段时间让文档创建完成
      await new Promise(r => setTimeout(r, 500))

      // 简化的块结构测试 - 使用飞书正确的 block 结构
      const allBlocks = [
        {
          block_type: 3, // heading1
          heading1: {
            elements: [{ text_run: { content: '📊 测试报告摘要' } }],
          },
        },
        {
          block_type: 2, // text
          text: {
            elements: [{ text_run: { content: '这是一个由 AI 自动生成的测试文档。' } }],
          },
        },
        {
          block_type: 2, // text
          text: {
            elements: [{ text_run: { content: '文档创建时间：' + new Date().toLocaleString() } }],
          },
        },
      ]

      addResult('2. 添加块', { message: `正在添加 ${allBlocks.length} 个块...` }, '')

      const blocksRes = await fetch(`${API_BASE}/api/feishu/create-nested-blocks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          document_id: documentId,
          user_token: userToken,
          blocks: allBlocks,
        }),
      })
      const blocksData = await blocksRes.json()
      addResult('2. 添加块', blocksData, `${Date.now() - start}ms`)

      // 完成
      addResult('完成', {
        success: true,
        data: {
          message: '文档创建成功！',
          url: docUrl,
          document_id: documentId,
        },
      }, `${Date.now() - start}ms`)

    } catch (err) {
      addResult('错误', { error: String(err) }, `${Date.now() - start}ms`)
    }

    setLoading(false)
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-800 mb-6">🔧 TS 后端接口调试</h1>

        {/* Token 获取 */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="font-semibold text-gray-700 mb-3">🔑 Token 获取</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-gray-600 mb-1">方法1: OAuth 授权登录</label>
              <button
                onClick={getOAuthUrl}
                disabled={loading}
                className="w-full px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
              >
                一键登录飞书获取 Token
              </button>
              <p className="text-xs text-gray-500 mt-1">点击后会跳转到飞书授权，授权后自动返回</p>
            </div>
            <div>
              <label className="block text-sm text-gray-600 mb-1">方法2: 手动兑换 Code</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={manualCode}
                  onChange={e => setManualCode(e.target.value)}
                  placeholder="输入飞书授权码"
                  className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  onClick={exchangeToken}
                  disabled={loading}
                  className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
                >
                  兑换
                </button>
              </div>
              <p className="text-xs text-gray-500 mt-1">从飞书 API 调试台获取的临时 code</p>
            </div>
          </div>
        </div>

        {/* 模拟创建完整文档 */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="font-semibold text-gray-700 mb-3">🤖 模拟创建完整文档</h2>
          <p className="text-sm text-gray-600 mb-3">自动创建文档并添加标题、段落、列表、引用、代码块等</p>
          <div className="flex items-center gap-4">
            <button
              onClick={createFullDocument}
              disabled={loading || !userToken}
              className="px-6 py-3 bg-purple-500 text-white rounded-lg hover:bg-purple-600 disabled:opacity-50 font-medium"
            >
              {loading ? '创建中...' : '🚀 一键创建飞书文档'}
            </button>
            {createdDocUrl && (
              <a
                href={createdDocUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="px-4 py-2 bg-green-100 text-green-700 rounded-lg hover:bg-green-200 font-medium"
              >
                📄 查看文档
              </a>
            )}
          </div>
          {createdDocUrl && (
            <div className="mt-3 p-3 bg-gray-50 rounded">
              <label className="block text-sm text-gray-600 mb-1">文档链接：</label>
              <a
                href={createdDocUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-600 hover:text-blue-800 break-all"
              >
                {createdDocUrl}
              </a>
            </div>
          )}
        </div>

        {/* Token 状态 */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="flex items-center justify-between mb-3">
            <h2 className="font-semibold text-gray-700">📋 Token 状态</h2>
            <button
              onClick={() => {
                setUserToken('')
                setTokenExpires('')
                localStorage.removeItem(USER_TOKEN_KEY)
                localStorage.removeItem(USER_TOKEN_EXPIRES_KEY)
              }}
              className="px-3 py-1 text-sm bg-red-100 text-red-600 rounded hover:bg-red-200"
            >
              清空 Token
            </button>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="block text-sm text-gray-600 mb-1">User Token</label>
              <div className="text-xs text-gray-800 break-all bg-gray-50 p-2 rounded max-h-20 overflow-y-auto">
                {userToken || '未获取'}
              </div>
            </div>
            <div>
              <label className="block text-sm text-gray-600 mb-1">过期时间</label>
              <div className="text-xs text-gray-800 bg-gray-50 p-2 rounded">
                {tokenExpires ? new Date(tokenExpires).toLocaleString() : '未获取或无过期时间'}
              </div>
            </div>
            <div>
              <label className="block text-sm text-gray-600 mb-1">Folder Token</label>
              <input
                type="text"
                value={folderToken}
                onChange={e => setFolderToken(e.target.value)}
                placeholder="云盘文件夹 token"
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-xs"
              />
            </div>
            <div>
              <label className="block text-sm text-gray-600 mb-1">Document ID</label>
              <input
                type="text"
                value={documentId}
                onChange={e => setDocumentId(e.target.value)}
                placeholder="文档 ID (docx...)"
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-xs"
              />
            </div>
          </div>
        </div>

        {/* 测试按钮 */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="font-semibold text-gray-700 mb-3">🧪 接口测试</h2>
          
          {/* 文档标题输入 */}
          <div className="mb-3">
            <label className="block text-sm text-gray-600 mb-1">创建文档标题</label>
            <input
              type="text"
              value={docTitle}
              onChange={e => setDocTitle(e.target.value)}
              placeholder="输入文档标题"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            />
          </div>
          
          <div className="flex flex-wrap gap-2">
            {tests.map((test, i) => (
              <button
                key={i}
                onClick={() => {
                  const body = test.api === '/api/health2'
                    ? {}
                    : { ...test.body, user_token: userToken }
                  callApi(test.api, body)
                }}
                disabled={loading || (test.api !== '/api/health2' && !userToken)}
                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50 transition"
                title={!userToken && test.api !== '/api/health2' ? '请先获取 Token' : ''}
              >
                {test.name}
              </button>
            ))}
            <button
              onClick={() => setResults([])}
              className="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600 transition"
            >
              清空结果
            </button>
          </div>
        </div>

        {/* GET 接口测试 */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h2 className="font-semibold text-gray-700 mb-3">📖 文档内容获取</h2>
          <button
            onClick={() => {
              if (!documentId) {
                alert('请输入 Document ID')
                return
              }
              callGetApi('/api/feishu/document-content', {
                document_id: documentId,
                user_token: userToken,
              })
            }}
            disabled={loading || !userToken || !documentId}
            className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50 transition"
            title={!documentId ? '请输入 Document ID' : !userToken ? '请先获取 Token' : ''}
          >
            获取文档纯文本内容
          </button>
        </div>

        {/* 结果展示 */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-4 border-b border-gray-200">
            <h2 className="font-semibold text-gray-700">📊 测试结果</h2>
          </div>
          <div className="divide-y divide-gray-100 max-h-[600px] overflow-y-auto">
            {results.length === 0 ? (
              <p className="p-4 text-gray-500 text-center">暂无测试结果</p>
            ) : (
              results.map((r, i) => (
                <div key={i} className="p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-mono text-sm text-blue-600">{r.api}</span>
                    <span className="text-xs text-gray-400">{r.time}</span>
                  </div>
                  <pre className="bg-gray-50 p-3 rounded text-xs overflow-x-auto max-h-64">
                    {JSON.stringify(r.result, null, 2)}
                  </pre>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
