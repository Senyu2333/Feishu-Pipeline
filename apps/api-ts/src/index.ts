import Fastify from 'fastify'
import cors from '@fastify/cors'
import axios from 'axios'
import type { FastifyRequest, FastifyReply } from 'fastify'
import { feishuClient, lark, getDepartmentChildren, batchGetDepartments, batchGetUserNames, sendMessage } from './lib/feishu.js'

const app = Fastify({
  logger: {
    transport: {
      target: 'pino-pretty',
      options: { colorize: true, translateTime: 'SYS:HH:MM:ss', ignore: 'pid,hostname' },
    },
  },
})

await app.register(cors, {
  origin: process.env.CORS_ORIGIN ?? '*',
  credentials: true,
})

app.get('/api/health2', async () => ({ status: 'ok', service: 'api-ts' }))

import { feishuRoutes } from './routes/feishu.js'
app.register(feishuRoutes)

app.post("/api/feishu/list-files", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { folder_token, page_size, page_token, order_by, direction, user_token } = request.body as any
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

app.get("/api/feishu/oauth-url", async (_request: FastifyRequest, reply: FastifyReply) => {
  try {
    const res = await axios.get('http://localhost:8080/api/auth/feishu/config')
    const { appId } = res.data.data || {}

    if (!appId) {
      return reply.status(500).send({ success: false, error: 'Failed to get appId from Go backend' })
    }

    // 需要申请的权限 scope（包括获取部门名称和用户姓名、发送消息所需的权限）
    const scope = [
      'docx:document:create',
      'docx:document',
      'docx:document:readonly',
      'drive:drive:readonly',
      'contact:contact.base:readonly',
      'contact:department.base:readonly',
      'contact:department.organize:readonly',
      'contact:user.basic_profile:readonly',
      'im:message',
      'im:message:send_as_bot',
    ].join(' ')
    const redirectUri = encodeURIComponent('http://localhost:3001/api/feishu/callback')
    const oauthUrl = `https://open.feishu.cn/open-apis/authen/v1/authorize?app_id=${appId}&redirect_uri=${redirectUri}&state=ts-auth&scope=${encodeURIComponent(scope)}`

    return reply.send({ success: true, data: { oauthUrl, appId } })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

app.get("/api/feishu/callback", async (request: FastifyRequest, reply: FastifyReply) => {
  const { code } = request.query as { code?: string }

  if (!code) {
    return reply.redirect('http://localhost:5173/debug?error=no_code')
  }

  try {
    const appId = process.env.FEISHU_APP_ID ?? 'cli_a954fa893fb85bc6'
    const appSecret = process.env.FEISHU_APP_SECRET ?? 'aYDUH3soLMlwONsU262qpcziZmjVDwOe'

    const appTokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal',
      { app_id: appId, app_secret: appSecret }
    )
    const appAccessToken = appTokenRes.data?.app_access_token

    if (!appAccessToken) {
      const debugInfo = encodeURIComponent(JSON.stringify(appTokenRes.data))
      return reply.redirect(`http://localhost:5173/debug?error=no_app_token&detail=${debugInfo}`)
    }

    const tokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/authen/v1/oidc/access_token',
      { grant_type: 'authorization_code', code },
      { headers: { Authorization: `Bearer ${appAccessToken}` } }
    )

    const { access_token, refresh_token } = tokenRes.data.data || {}

    if (!access_token) {
      const debugInfo = encodeURIComponent(JSON.stringify(tokenRes.data))
      return reply.redirect(`http://localhost:5173/debug?error=no_token&detail=${debugInfo}`)
    }

    return reply.redirect(`http://localhost:5173/debug?token=${encodeURIComponent(access_token)}&refresh_token=${encodeURIComponent(refresh_token || '')}`)
  } catch (err) {
    const error = err as { response?: { data?: unknown } }
    const debugInfo = encodeURIComponent(JSON.stringify(error.response?.data || String(err)))
    return reply.redirect(`http://localhost:5173/debug?error=oauth_failed&detail=${debugInfo}`)
  }
})

app.post("/api/feishu/exchange-token", async (request: FastifyRequest, reply: FastifyReply) => {
  const { code } = request.body as { code?: string }

  if (!code) {
    return reply.status(400).send({ success: false, error: 'code is required' })
  }

  try {
    const appId = process.env.FEISHU_APP_ID ?? 'cli_a954fa893fb85bc6'
    const appSecret = process.env.FEISHU_APP_SECRET ?? 'aYDUH3soLMlwONsU262qpcziZmjVDwOe'

    const appTokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal',
      { app_id: appId, app_secret: appSecret }
    )
    const appAccessToken = appTokenRes.data?.app_access_token

    if (!appAccessToken) {
      return reply.status(400).send({ success: false, error: 'Failed to get app_access_token' })
    }

    const tokenRes = await axios.post(
      'https://open.feishu.cn/open-apis/authen/v1/oidc/access_token',
      { grant_type: 'authorization_code', code },
      { headers: { Authorization: `Bearer ${appAccessToken}` } }
    )

    const { access_token, refresh_token, expires_in } = tokenRes.data.data || {}

    if (!access_token) {
      return reply.status(400).send({ success: false, error: 'Failed to get token' })
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

app.post("/api/feishu/get-drive-files", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { folder_token, page_size, page_token, order_by, direction, user_id_type, user_token } = request.body as any
    const result = await axios.get("https://open.feishu.cn/open-apis/drive/v1/files", {
      headers: { Authorization: `Bearer ${user_token || ''}` },
      params: { folder_token, page_size, page_token, order_by, direction, user_id_type },
    })
    return reply.send(result.data)
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ error: error.message })
  }
})

app.post("/api/feishu/create-document", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { folder_token, title, user_token } = request.body as any
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

app.get("/api/feishu/document-content", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, user_token, lang } = request.query as any
    if (!document_id || !user_token) {
      return reply.status(400).send({ success: false, error: 'document_id and user_token are required' })
    }
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/raw_content`,
      {
        headers: { Authorization: `Bearer ${user_token}` },
        params: { lang: lang || 0 },
      }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

app.get("/api/feishu/department-children", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { department_id, user_token, fetch_child, page_size, page_token, user_id_type, department_id_type } = request.query as any
    if (!department_id) {
      return reply.status(400).send({ success: false, error: 'department_id is required' })
    }
    const result = await getDepartmentChildren(department_id, user_token, {
      fetch_child,
      page_size,
      page_token,
      user_id_type,
      department_id_type,
    })
    return reply.send({ success: true, data: result })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

app.get("/api/feishu/batch-departments", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { department_ids, user_token, department_id_type, user_id_type } = request.query as any
    
    if (!department_ids) {
      return reply.status(400).send({ success: false, error: 'department_ids is required (comma-separated or array)' })
    }
    
    // 支持逗号分隔的字符串或数组
    const ids = Array.isArray(department_ids) ? department_ids : department_ids.split(',')
    
    if (ids.length > 50) {
      return reply.status(400).send({ success: false, error: 'Maximum 50 department IDs allowed' })
    }
    
    const result = await batchGetDepartments(ids, user_token, {
      department_id_type,
      user_id_type,
    })
    return reply.send({ success: true, data: result })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

app.post("/api/feishu/batch-departments", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { department_ids, user_token, department_id_type, user_id_type } = request.body as any
    
    if (!department_ids || !Array.isArray(department_ids)) {
      return reply.status(400).send({ success: false, error: 'department_ids is required (array)' })
    }
    
    if (department_ids.length > 50) {
      return reply.status(400).send({ success: false, error: 'Maximum 50 department IDs allowed' })
    }
    
    const result = await batchGetDepartments(department_ids, user_token, {
      department_id_type,
      user_id_type,
    })
    return reply.send({ success: true, data: result })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

app.post("/api/feishu/batch-user-names", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { user_ids, user_token, user_id_type } = request.body as any
    
    if (!user_ids || !Array.isArray(user_ids)) {
      return reply.status(400).send({ success: false, error: 'user_ids is required (array)' })
    }
    
    if (user_ids.length > 10) {
      return reply.status(400).send({ success: false, error: 'Maximum 10 user IDs allowed' })
    }
    
    const result = await batchGetUserNames(user_ids, user_token, {
      user_id_type,
    })
    return reply.send({ success: true, data: result })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

app.post("/api/feishu/send-message", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { receive_id, receive_id_type, msg_type, content, uuid } = request.body as any
    
    if (!receive_id) {
      return reply.status(400).send({ success: false, error: 'receive_id is required' })
    }
    
    if (!msg_type) {
      return reply.status(400).send({ success: false, error: 'msg_type is required' })
    }
    
    if (!content) {
      return reply.status(400).send({ success: false, error: 'content is required' })
    }
    
    // 使用 tenant_access_token（机器人身份）发送消息，不使用 user_token
    const result = await sendMessage(receive_id, receive_id_type || 'open_id', msg_type, content, undefined, uuid)
    return reply.send({ success: true, data: result })
  } catch (err) {
    const error = err as Error
    return reply.status(500).send({ success: false, error: error.message })
  }
})

app.post("/api/feishu/create-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, children } = request.body as any
    if (!document_id || !user_token) {
      return reply.status(400).send({ success: false, error: 'document_id and user_token are required' })
    }
    const targetBlockId = block_id || document_id
    const defaultBlock = {
      block_type: 2,
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

app.post("/api/feishu/create-nested-blocks", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, blocks } = request.body as any
    if (!document_id || !user_token || !blocks) {
      return reply.status(400).send({ success: false, error: 'document_id, user_token, and blocks are required' })
    }
    const targetBlockId = block_id || document_id
    const result = await axios.post(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${targetBlockId}/children`,
      { children: blocks },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return reply.send({ success: true, data: result.data })
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return reply.status(500).send({ success: false, error: error.message, detail: error.response?.data })
  }
})

app.get("/api/feishu/get-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token } = request.query as any
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

app.get("/api/feishu/get-children", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, page_size, page_token } = request.query as any
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

app.put("/api/feishu/update-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token, update_blocks } = request.body as any
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

app.put("/api/feishu/batch-update-blocks", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, user_token, requests } = request.body as any
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

app.delete("/api/feishu/delete-block", async (request: FastifyRequest, reply: FastifyReply) => {
  try {
    const { document_id, block_id, user_token } = request.query as any
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

const PORT = Number(process.env.PORT ?? 3001)

try {
  await app.listen({ port: PORT, host: '0.0.0.0' })
} catch (err) {
  app.log.error(err)
  process.exit(1)
}

const shutdown = async () => {
  await app.close()
  process.exit(0)
}
process.on('SIGINT', shutdown)
process.on('SIGTERM', shutdown)
