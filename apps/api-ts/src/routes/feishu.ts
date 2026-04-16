/**
 * 飞书 API 路由示例
 * 展示如何在 TS 后端直接调用 Go 后端没有的飞书 API
 */

import type { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify'
import { feishuClient, lark, getDocumentRawContent } from '../lib/feishu.js'
import { http } from '../lib/http.js'

export async function feishuRoutes(app: FastifyInstance) {
  // ── 示例 1: 获取飞书文档内容（Go 后端没有的 API）──────────────────────
  app.get('/api/feishu/doc/:documentId', async (request: FastifyRequest, reply: FastifyReply) => {
    const { documentId } = request.params as { documentId: string }
    try {
      const content = await getDocumentRawContent(documentId)
      return reply.send({ success: true, data: content })
    } catch (err) {
      const error = err as Error
      return reply.status(500).send({ success: false, error: error.message })
    }
  })

  // ── 示例 2: 获取用户信息（Go 后端已有，但 TS 可直接调）────────────────
  app.get('/api/feishu/user/:userId', async (request: FastifyRequest, reply: FastifyReply) => {
    const { userId } = request.params as { userId: string }
    try {
      // 直接用飞书 SDK 调用（需要 user_access_token）
      // 如果是服务端调用，建议通过 Go 后端获取 token
      const userInfo = await feishuClient.contact.v3.user.get({
        path: { user_id: userId },
        params: { user_id_type: 'open_id' },
      })
      return reply.send({ success: true, data: userInfo })
    } catch (err) {
      const error = err as Error
      return reply.status(500).send({ success: false, error: error.message })
    }
  })

  // ── 示例 3: 调用 Go 后端 API（通过 http.ts）──────────────────────────
  app.get('/api/go/health', async (_request: FastifyRequest, reply: FastifyReply) => {
    try {
      const response = await http.get('/api/health')
      return reply.send({ success: true, data: response.data })
    } catch (err) {
      const error = err as Error
      return reply.status(500).send({ success: false, error: error.message })
    }
  })

  // ── 示例 4: 调用 Go 后端获取当前用户 ─────────────────────────────────
  app.get('/api/go/me', async (request: FastifyRequest, reply: FastifyReply) => {
    try {
      // 从请求头获取 session cookie
      const cookieHeader = request.headers.cookie as string | undefined
      const response = await http.get('/api/me', {
        headers: cookieHeader ? { cookie: cookieHeader } : {},
      })
      return reply.send({ success: true, data: response.data })
    } catch (err) {
      return reply.status(401).send({ success: false, error: '未登录' })
    }
  })

  // ── 示例 5: 调用需要用户 token 的飞书 API ───────────────────────────
  // 先从 Go 后端获取 user token（如果有接口的话）
  app.post('/api/feishu/user-profile', async (request: FastifyRequest<{
    Body: { userToken: string; userId: string }
  }>, reply: FastifyReply) => {
    const { userToken, userId } = request.body
    try {
      // 使用 user token 调用需要用户权限的 API
      const profile = await feishuClient.contact.v3.user.get(
        {
          path: { user_id: userId },
          params: { user_id_type: 'open_id' },
        },
        lark.withUserAccessToken(userToken)
      )
      return reply.send({ success: true, data: profile })
    } catch (err) {
      const error = err as Error
      return reply.status(500).send({ success: false, error: error.message })
    }
  })
}
