/**
 * 飞书 SDK 客户端初始化
 * 文档: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/server-side-sdk/nodejs-sdk/using-the-sdk
 */

import lark from '@larksuiteoapi/node-sdk'

// 飞书应用配置
const FEISHU_APP_ID = process.env.FEISHU_APP_ID ?? 'cli_a954fa893fb85bc6'
const FEISHU_APP_SECRET = process.env.FEISHU_APP_SECRET ?? 'aYDUH3soLMlwONsU262qpcziZmjVDwOe'

// 验证配置
if (!FEISHU_APP_ID || !FEISHU_APP_SECRET) {
  console.warn('[Feishu] 警告: 未配置 FEISHU_APP_ID 或 FEISHU_APP_SECRET')
}

// 创建飞书客户端实例
// disableTokenCache=false: SDK 自动管理租户 Token 的获取与刷新（推荐）
export const feishuClient = new lark.Client({
  appId: FEISHU_APP_ID,
  appSecret: FEISHU_APP_SECRET,
  disableTokenCache: false,
})

// 导出 lark 命名空间，方便使用 withTenantToken / withUserAccessToken 等工具函数
export { lark }

/**
 * 调用需要用户权限的 API（如获取用户信息）
 *
 * @param userToken - 从 Go 后端获取的 user_access_token
 * @param apiCall - 实际调用的飞书 API（传入 SDK 方法）
 *
 * @example
 * const userInfo = await withUserToken(userToken, (token) =>
 *   client.contact.v3.user.get({
 *     path: { user_id: 'user_id' },
 *   }, lark.withUserAccessToken(token))
 * )
 */
export async function withUserToken<T>(
  userToken: string,
  apiCall: (token: string) => Promise<T>
): Promise<T> {
  return apiCall(userToken)
}

/**
 * 获取用户信息（需要 user_access_token）
 * @see https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/user/get
 */
export async function getUserInfo(userToken: string, userId: string): Promise<unknown> {
  return feishuClient.contact.v3.user.get(
    {
      path: { user_id: userId },
      params: { user_id_type: 'open_id' },
    },
    lark.withUserAccessToken(userToken)
  )
}

/**
 * 示例: 调用文档 API（仅需 tenant token，SDK 自动处理）
 * @see https://open.feishu.cn/document/server-docs/docs/docs/docx-v1/document/raw-content
 */
export async function getDocumentRawContent(documentId: string): Promise<unknown> {
  return feishuClient.docx.v1.document.rawContent({
    path: { document_id: documentId },
    params: { lang: 0 },
  })
}

