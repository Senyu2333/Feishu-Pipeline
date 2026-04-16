import axios from 'axios'

type ToolFunctionParams = {
  user_token?: string
  folder_token?: string
  document_id?: string
  name?: string
  title?: string
  page_size?: number
  page_token?: string
  order_by?: string
  direction?: string
  block_id?: string
  content?: string
  blocks?: object[]
  update_blocks?: object[]
  requests?: object[]
}

type ToolParams = ToolFunctionParams

interface ToolResult {
  success: boolean
  data?: unknown
  error?: string
}

export async function getAllDocuments(params: ToolParams): Promise<ToolResult> {
  try {
    const { user_token, folder_token, page_size, page_token, order_by, direction } = params
    if (!user_token) {
      return { success: false, error: 'user_token is required' }
    }
    const result = await axios.get('https://open.feishu.cn/open-apis/drive/v1/files', {
      headers: { Authorization: `Bearer ${user_token}` },
      params: {
        folder_token: folder_token || undefined,
        page_size: page_size || 100,
        page_token: page_token || undefined,
        order_by: order_by || 'EditedTime',
        direction: direction || 'DESC',
      },
    })
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function getDocumentContent(params: ToolParams): Promise<ToolResult> {
  try {
    const { user_token, document_id } = params
    if (!user_token || !document_id) {
      return { success: false, error: 'user_token and document_id are required' }
    }
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/raw_content`,
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function createFileFolder(params: ToolParams): Promise<ToolResult> {
  try {
    const { user_token, folder_token, name } = params
    if (!user_token || !folder_token || !name) {
      return { success: false, error: 'user_token, folder_token, and name are required' }
    }
    const result = await axios.post(
      'https://open.feishu.cn/open-apis/drive/v1/files/create_folder',
      { name, folder_token },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function createDocument(params: ToolParams): Promise<ToolResult> {
  try {
    const { user_token, folder_token, title } = params
    if (!user_token || !title) {
      return { success: false, error: 'user_token and title are required' }
    }
    const result = await axios.post(
      'https://open.feishu.cn/open-apis/docx/v1/documents',
      { folder_token: folder_token || undefined, title },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

const tools = [
  {
    type: "function",
    function: {
      name: "getAllDocuments",
      description: "获取云盘文件列表，支持文件夹。不填folder_token则获取根目录",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          folder_token: { type: "string", description: "文件夹token，不填则获取根目录" },
          page_size: { type: "number", description: "每页数量，最大200，默认100" },
          page_token: { type: "string", description: "分页标记，用于翻下一页" },
          order_by: { type: "string", enum: ["EditedTime", "CreatedTime"], description: "排序字段" },
          direction: { type: "string", enum: ["ASC", "DESC"], description: "排序方向" },
        },
        required: ["user_token"]
      }
    }
  },
  {
    type: "function",
    function: {
      name: "getDocumentContent",
      description: "获取飞书文档的纯文本内容",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID，格式如 doxbcmEtb..." }
        },
        required: ["user_token", "document_id"]
      }
    }
  },
  {
    type: "function",
    function: {
      name: "createFileFolder",
      description: "在飞书云盘中创建文件夹",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          folder_token: { type: "string", description: "父文件夹token" },
          name: { type: "string", description: "文件夹名称" }
        },
        required: ["user_token", "folder_token", "name"]
      }
    }
  },
  {
    type: "function",
    function: {
      name: "createDocument",
      description: "创建飞书文档（docx）",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          folder_token: { type: "string", description: "文件夹token，不填则在根目录创建" },
          title: { type: "string", description: "文档标题" }
        },
        required: ["user_token", "title"]
      }
    }
  }
]

export async function createBlock(params: ToolParams & { block_id?: string; content?: string }): Promise<ToolResult> {
  try {
    const { user_token, document_id, block_id, content } = params
    if (!user_token || !document_id) {
      return { success: false, error: 'user_token and document_id are required' }
    }
    const targetBlockId = block_id || document_id
    const defaultBlock = {
      block_type: 2,
      text: { elements: [{ text_run: { content: content || '新段落', text_element_style: {} } }], style: {} },
    }
    const result = await axios.post(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${targetBlockId}/children`,
      { children: [defaultBlock] },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function createNestedBlocks(params: ToolParams & { block_id?: string; blocks?: object[] }): Promise<ToolResult> {
  try {
    const { user_token, document_id, block_id, blocks } = params
    if (!user_token || !document_id || !blocks) {
      return { success: false, error: 'user_token, document_id, and blocks are required' }
    }
    const targetBlockId = block_id || document_id
    const result = await axios.post(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${targetBlockId}/children`,
      { children: blocks },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function getBlock(params: ToolParams & { block_id: string }): Promise<ToolResult> {
  try {
    const { user_token, document_id, block_id } = params
    if (!user_token || !document_id || !block_id) {
      return { success: false, error: 'user_token, document_id, and block_id are required' }
    }
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${block_id}`,
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function getChildren(params: ToolParams & { block_id?: string }): Promise<ToolResult> {
  try {
    const { user_token, document_id, block_id, page_size, page_token } = params
    if (!user_token || !document_id) {
      return { success: false, error: 'user_token and document_id are required' }
    }
    const targetBlockId = block_id || document_id
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${targetBlockId}/children`,
      {
        headers: { Authorization: `Bearer ${user_token}` },
        params: { page_size: page_size || 500, page_token: page_token || undefined },
      }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function updateBlock(params: ToolParams & { block_id: string; update_blocks: object[] }): Promise<ToolResult> {
  try {
    const { user_token, document_id, block_id, update_blocks } = params
    if (!user_token || !document_id || !block_id || !update_blocks) {
      return { success: false, error: 'user_token, document_id, block_id, and update_blocks are required' }
    }
    const result = await axios.patch(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${block_id}`,
      { update_blocks },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function batchUpdateBlocks(params: ToolParams & { requests: object[] }): Promise<ToolResult> {
  try {
    const { user_token, document_id, requests } = params
    if (!user_token || !document_id || !requests) {
      return { success: false, error: 'user_token, document_id, and requests are required' }
    }
    const result = await axios.post(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/batch_update`,
      { requests },
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function deleteBlock(params: ToolParams & { block_id: string }): Promise<ToolResult> {
  try {
    const { user_token, document_id, block_id } = params
    if (!user_token || !document_id || !block_id) {
      return { success: false, error: 'user_token, document_id, and block_id are required' }
    }
    const result = await axios.delete(
      `https://open.feishu.cn/open-apis/docx/v1/documents/${document_id}/blocks/${block_id}`,
      { headers: { Authorization: `Bearer ${user_token}` } }
    )
    return { success: true, data: result.data }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export const toolFunctions: Record<string, (params: any) => Promise<ToolResult>> = {
  getAllDocuments,
  getDocumentContent,
  createFileFolder,
  createDocument,
  createBlock,
  createNestedBlocks,
  getBlock,
  getChildren,
  updateBlock,
  batchUpdateBlocks,
  deleteBlock,
}


const blockTools = [
  {
    type: "function",
    function: {
      name: "createBlock",
      description: "在飞书文档中创建块（段落）",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID" },
          block_id: { type: "string", description: "父块ID，不填则追加到文档根级别" },
          content: { type: "string", description: "段落文本内容" },
        },
        required: ["user_token", "document_id"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "createNestedBlocks",
      description: "批量创建嵌套块（支持多层级结构）",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID" },
          block_id: { type: "string", description: "父块ID，不填则追加到文档根级别" },
          blocks: {
            type: "array",
            description: "块数组，每项包含 block_type 和内容",
            items: { type: "object" },
          },
        },
        required: ["user_token", "document_id", "blocks"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "getBlock",
      description: "获取飞书文档中指定块的内容",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID" },
          block_id: { type: "string", description: "块ID" },
        },
        required: ["user_token", "document_id", "block_id"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "getChildren",
      description: "获取文档中指定块的所有子块",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID" },
          block_id: { type: "string", description: "块ID，不填则获取文档根级别块" },
          page_size: { type: "number", description: "每页数量，最大500" },
          page_token: { type: "string", description: "分页标记" },
        },
        required: ["user_token", "document_id"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "updateBlock",
      description: "更新飞书文档中指定块的内容",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID" },
          block_id: { type: "string", description: "块ID" },
          update_blocks: {
            type: "array",
            description: "更新块数组",
            items: { type: "object" },
          },
        },
        required: ["user_token", "document_id", "block_id", "update_blocks"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "batchUpdateBlocks",
      description: "批量更新飞书文档中的多个块",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID" },
          requests: {
            type: "array",
            description: "批量更新请求数组",
            items: { type: "object" },
          },
        },
        required: ["user_token", "document_id", "requests"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "deleteBlock",
      description: "删除飞书文档中的指定块",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌" },
          document_id: { type: "string", description: "文档ID" },
          block_id: { type: "string", description: "块ID" },
        },
        required: ["user_token", "document_id", "block_id"],
      },
    },
  },
]

const allTools = [...tools, ...blockTools]

export default allTools
