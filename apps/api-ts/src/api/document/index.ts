import axios from 'axios'
import lark from '@larksuiteoapi/node-sdk'
import { feishuClient } from '../../lib/feishu.js'

// 飞书应用配置
const FEISHU_APP_ID = process.env.FEISHU_APP_ID ?? 'cli_a954fa893fb85bc6'
const FEISHU_APP_SECRET = process.env.FEISHU_APP_SECRET ?? 'aYDUH3soLMlwONsU262qpcziZmjVDwOe'

/**
 * 创建飞书文档
 */
export async function createDocument(
  folderToken?: string,
  title: string = '未命名文档',
  userToken?: string
): Promise<unknown> {
  return feishuClient.docx.v1.document.create(
    {
      data: {
        folder_token: folderToken || undefined,
        title,
      },
    },
    userToken ? lark.withUserAccessToken(userToken) : undefined
  )
}

/**
 * 创建文档块
 * @param documentId - 文档 ID
 * @param blocks - 块数组
 * @param blockId - 目标块 ID（不填则添加到文档根级别）
 * @param userToken - 用户 Token
 */
export async function createDocumentBlocks(
  documentId: string,
  blocks: object[],
  blockId?: string,
  userToken?: string
): Promise<unknown> {
  const targetBlockId = blockId || documentId
  const response = await axios.post(
    `https://open.feishu.cn/open-apis/docx/v1/documents/${documentId}/blocks/${targetBlockId}/children`,
    { children: blocks },
    {
      headers: {
        Authorization: `Bearer ${userToken}`,
        'Content-Type': 'application/json; charset=utf-8',
      },
    }
  )
  return response.data
}

/**
 * 获取文档纯文本内容
 */
export async function getDocumentContent(
  documentId: string,
  userToken: string
): Promise<unknown> {
  const response = await axios.get(
    `https://open.feishu.cn/open-apis/docx/v1/documents/${documentId}/raw_content`,
    {
      headers: { Authorization: `Bearer ${userToken}` },
      params: { lang: 0 },
    }
  )
  return response.data
}

// 导出 lark 用于 withUserAccessToken
export { lark }
