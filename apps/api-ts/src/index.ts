import Fastify from 'fastify'
import cors from '@fastify/cors'
import axios from 'axios'
import type { FastifyRequest, FastifyReply } from 'fastify'
import { feishuClient, lark } from './lib/feishu.js'

const app = Fastify({
  logger: {
    transport: {
      target: 'pino-pretty',
      options: { colorize: true, translateTime: 'SYS:HH:MM:ss', ignore: 'pid,hostname' },
    },
  },
})

// ── 插件 ────────────────────────────────────────────────────────────
await app.register(cors, {
  // 开发时允许 Vite dev server 跨域；生产按需收紧
  origin: process.env.CORS_ORIGIN ?? '*',
  credentials: true,
})

// ── 健康检查 ─────────────────────────────────────────────────────────
app.get('/api/health2', async () => ({ status: 'ok', service: 'api-ts' }))

// ── 在这里 import 并注册你的路由模块 ────────────────────────────────
import { feishuRoutes } from './routes/feishu.js'
app.register(feishuRoutes)

//获取文件列表（使用飞书 SDK + user token）
app.post("/api/feishu/list-files", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { folder_token, page_size, page_token, order_by, direction, user_token } = request.body as {
      folder_token?: string
      page_size?: number
      page_token?: string
      order_by?: string
      direction?: string
      user_token?: string
    }
    if (!user_token) {
      return reply.status(400).send({ success: false, error: 'user_token is required' })
    }
    const result = await feishuClient.drive.v1.file.list(
      {
        params: {
          folder_token: folder_token || undefined,
          page_size: page_size || 50,
          page_token: page_token || undefined,
          order_by: (order_by as any) || undefined,
          direction: (direction as any) || undefined,
        },
      },
      lark.withUserAccessToken(user_token)
    )
    return reply.send({ success: true, data: result })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 获取飞书 OAuth 授权 URL（从 Go 后端获取 appId）
app.get("/api/feishu/oauth-url", async (_request: FastifyRequest, reply: FastifyReply) => {
  try {
    // 从 Go 后端获取 OAuth 配置
    const res = await axios.get('http://localhost:8080/api/auth/feishu/config')
    const { appId } = res.data.data || {}

    if (!appId) {
      return reply.status(500).send({ success: false, error: 'Failed to get appId from Go backend' })
    }

    // 需要申请的权限 scope
    const scope = 'docx:document:create docx:document docx:document:readonly drive:drive:readonly'
    const redirectUri = encodeURIComponent('http://localhost:3001/api/feishu/callback')
    const oauthUrl = `https://open.feishu.cn/open-apis/authen/v1/authorize?app_id=${appId}&redirect_uri=${redirectUri}&state=ts-auth&scope=${encodeURIComponent(scope)}`

    return reply.send({ success: true, data: { oauthUrl, appId } })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 飞书 OAuth 回调 - 用 code 换取 user_token
app.get("/api/feishu/callback", async (request: FastifyRequest, reply: FastifyReply) => {
  const { code } = request.query as { code?: string }

  if (!code) {
    return reply.redirect('http://localhost:5173/debug?error=no_code')
  }

  try {
    const appId = process.env.FEISHU_APP_ID ?? 'cli_a954fa893fb85bc6'
    const appSecret = process.env.FEISHU_APP_SECRET ?? 'aYDUH3soLMlwONsU262qpcziZmjVDwOe'

    console.log('[OAuth] 换取 token，code:', code.slice(0, 20) + '...')

    // 第一步：获取 app_access_token
    const appTokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal',
      { app_id: appId, app_secret: appSecret }
    )
    const appAccessToken = appTokenRes.data?.app_access_token

    if (!appAccessToken) {
      const debugInfo = encodeURIComponent(JSON.stringify(appTokenRes.data))
      return reply.redirect(`http://localhost:5173/debug?error=no_app_token&detail=${debugInfo}`)
    }

    // 第二步：用 app_access_token + code 换取 user_token
    const tokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/authen/v1/oidc/access_token',
      {
        grant_type: 'authorization_code',
        code,
      },
      {
        headers: {
          Authorization: `Bearer ${appAccessToken}`,
        },
      }
    )

    console.log('[OAuth] 响应:', JSON.stringify(tokenRes.data))

    const { access_token, refresh_token } = tokenRes.data.data || {}

    if (!access_token) {
      const debugInfo = encodeURIComponent(JSON.stringify(tokenRes.data))
      return reply.redirect(`http://localhost:5173/debug?error=no_token&detail=${debugInfo}`)
    }

    return reply.redirect(`http://localhost:5173/debug?token=${encodeURIComponent(access_token)}&refresh_token=${encodeURIComponent(refresh_token || '')}`)
  } catch (err) {
    console.error('[OAuth] 换取 token 失败:', err)
    const error = err as { response?: { data?: unknown } }
    const debugInfo = encodeURIComponent(JSON.stringify(error.response?.data || String(err)))
    return reply.redirect(`http://localhost:5173/debug?error=oauth_failed&detail=${debugInfo}`)
  }
})

// 手动换取 token（用于调试）
app.post("/api/feishu/exchange-token", async (request: FastifyRequest, reply: FastifyReply) => {
  const { code } = request.body as { code?: string }

  if (!code) {
    return reply.status(400).send({ success: false, error: 'code is required' })
  }

  try {
    const appId = process.env.FEISHU_APP_ID ?? 'cli_a954fa893fb85bc6'
    const appSecret = process.env.FEISHU_APP_SECRET ?? 'aYDUH3soLMlwONsU262qpcziZmjVDwOe'

    // 第一步：获取 app_access_token
    const appTokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal',
      { app_id: appId, app_secret: appSecret }
    )
    const appAccessToken = appTokenRes.data?.app_access_token

    if (!appAccessToken) {
      return reply.status(400).send({ success: false, error: 'Failed to get app_access_token', detail: appTokenRes.data })
    }

    // 第二步：用 app_access_token + code 换取 user_token
    const tokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/authen/v1/oidc/access_token',
      {
        grant_type: 'authorization_code',
        code,
      },
      {
        headers: {
          Authorization: `Bearer ${appAccessToken}`,
        },
      }
    )

    const { access_token, refresh_token, expires_in } = tokenRes.data.data || {}

    if (!access_token) {
      return reply.status(400).send({ success: false, error: 'Failed to get token', detail: tokenRes.data })
    }

    return reply.send({
      success: true,
      data: {
        user_token: access_token,
        refresh_token,
        expires_in,
        expires_at: new Date(Date.now() + (expires_in || 7200) * 1000).toISOString(),
      },
    })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

//获取文件（使用 axios + user token）
app.post("/api/feishu/get-drive-files", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { folder_token, page_size, page_token, order_by, direction, user_id_type, user_token } = request.body as {
      folder_token?: string
      page_size?: number
      page_token?: string
      order_by?: string
      direction?: string
      user_id_type?: string
      user_token?: string
    }
    const result = await axios.get("https://open.feishu.cn/open-apis/drive/v1/files", {
      headers: {
        Authorization: `Bearer ${user_token || ''}`,
      },
      params: {
        folder_token,
        page_size,
        page_token,
        order_by,
        direction,
        user_id_type,
      },
    })
    return reply.send(result.data)
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ error: error.message })
  }
})

// ── 飞书文档 API ──────────────────────────────────────────────────────

// 创建文档（docx）
app.post("/api/feishu/create-document", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { folder_token, title, user_token } = request.body as {
      folder_token?: string
      title?: string
      user_token?: string
    }
    if (!user_token) {
      return reply.status(400).send({ success: false, error: 'user_token is required' })
    }
    const result = await feishuClient.docx.v1.document.create(
      {
        data: {
          folder_token: folder_token || undefined,
          title: title || '未命名文档',
        },
      },
      lark.withUserAccessToken(user_token)
    )
    return reply.send({ success: true, data: result })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 获取文档纯文本内容
app.get("/api/feishu/document-content", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, user_token, lang } = request.query as {
      document_id?: string
      user_token?: string
      lang?: number
    }
    if (!document_id || !user_token) {
      return reply.status(400).send({ success: false, error: 'document_id and user_token are required' })
    }
    // 使用 axios 直接调用
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/raw_content`,
      {
        headers: {
          Authorization: `Bearer ${user_token}`,
        },
        params: {
          lang: lang || 0,
        },
      }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// ── 飞书文档块 API ────────────────────────────────────────────────────

// 创建块（追加到指定块的子块末尾）
app.post("/api/feishu/create-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, children } = request.body as {
      document_id?: string
      block_id?: string
      user_token?: string
      children?: object[]
    }
    if (!document_id || !user_token) {
      return reply.status(400).send({ success: false, error: 'document_id and user_token are required' })
    }
    // block_id 为空则追加到文档根级别
    const targetBlockId = block_id || document_id
    const defaultBlock = {
      block_type: 2, // paragraph
      paragraph: { elements: [{ text_run: { content: '新段落', text_element_style: {} } }], style: {} },
    }
    const result = await axios.post(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${targetBlockId}/children`,
      { children: children || [defaultBlock] },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 批量创建嵌套块（支持多层结构）
app.post("/api/feishu/create-nested-blocks", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, blocks } = request.body as {
      document_id?: string
      block_id?: string
      user_token?: string
      blocks?: object[]
    }
    if (!document_id || !user_token || !blocks) {
      return reply.status(400).send({ success: false, error: 'document_id, user_token, and blocks are required' })
    }
    const targetBlockId = block_id || document_id
    console.log('[create-nested-blocks] document_id:', document_id, 'targetBlockId:', targetBlockId, 'blocks count:', blocks.length)
    
    const result = await axios.post(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${targetBlockId}/children`,
      { children: blocks },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    console.log('[create-nested-blocks] success:', result.data)
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    console.error('[create-nested-blocks] error:', error.response?.data || error.message)
    return reply.status(500).send({ success: false, error: error.message, detail: error.response?.data })
  }
})

// 获取块内容
app.get("/api/feishu/get-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token } = request.query as {
      document_id?: string
      block_id?: string
      user_token?: string
    }
    if (!document_id || !block_id || !user_token) {
      return reply.status(400).send({ success: false, error: 'document_id, block_id, and user_token are required' })
    }
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${block_id}`,
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 获取所有子块
app.get("/api/feishu/get-children", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, page_size, page_token } = request.query as {
      document_id?: string
      block_id?: string
      user_token?: string
      page_size?: number
      page_token?: string
    }
    if (!document_id || !block_id || !user_token) {
      return reply.status(400).send({ success: false, error: 'document_id, block_id, and user_token are required' })
    }
    const targetBlockId = block_id || document_id
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${targetBlockId}/children`,
      {
        headers: { Authorization: `Bearer ${user_token}` },
        params: { page_size: page_size || 500, page_token: page_token || undefined },
      }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 更新块内容
app.put("/api/feishu/update-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, update_blocks } = request.body as {
      document_id?: string
      block_id?: string
      user_token?: string
      update_blocks?: object[]
    }
    if (!document_id || !block_id || !user_token || !update_blocks) {
      return reply.status(400).send({ success: false, error: 'document_id, block_id, user_token, and update_blocks are required' })
    }
    const result = await axios.patch(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${block_id}`,
      { update_blocks },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 批量更新块内容
app.put("/api/feishu/batch-update-blocks", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, user_token, requests } = request.body as {
      document_id?: string
      user_token?: string
      requests?: object[]
    }
    if (!document_id || !user_token || !requests) {
      return reply.status(400).send({ success: false, error: 'document_id, user_token, and requests are required' })
    }
    const result = await axios.post(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/batch_update`,
      { requests },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// 删除块
app.delete("/api/feishu/delete-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token } = request.query as {
      document_id?: string
      block_id?: string
      user_token?: string
    }
    if (!document_id || !block_id || !user_token) {
      return reply.status(400).send({ success: false, error: 'document_id, block_id, and user_token are required' })
    }
    const result = await axios.delete(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${block_id}`,
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

// ── 启动 ─────────────────────────────────────────────────────────────
const PORT = Number(process.env.PORT ?? 3001)

try {
  await app.listen({ port: PORT, host: '0.0.0.0' })
} catch (err) {
  app.log.error(err)
  process.exit(1)
}






// 优雅关闭
const shutdown = async () => {
  await app.close()
  process.exit(0)
}
process.on('SIGINT', shutdown)
process.on('SIGTERM', shutdown)
