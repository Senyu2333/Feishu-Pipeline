import axios from 'axios'
import { importOpenApiFromUrl, importOpenApiFromSpec, ApifoxImportOptions } from './apifox.js'

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
  department_id?: string
  fetch_child?: boolean
  user_id_type?: string
  department_id_type?: string
  // Apifox 参数
  project_id?: string
  openapi_url?: string
  openapi_spec?: Record<string, unknown>
  target_endpoint_folder_id?: number
  target_schema_folder_id?: number
  endpoint_overwrite_behavior?: 'OVERWRITE_EXISTING' | 'AUTO_MERGE' | 'KEEP_EXISTING' | 'CREATE_NEW'
  schema_overwrite_behavior?: 'OVERWRITE_EXISTING' | 'AUTO_MERGE' | 'KEEP_EXISTING' | 'CREATE_NEW'
  update_folder_of_changed_endpoint?: boolean
  prepend_base_path?: boolean
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

// ── 通讯录工具函数 ──────────────────────────────────────────────────

export async function getDepartmentChildren(params: ToolParams): Promise<ToolResult> {
  try {
    const { user_token, department_id, fetch_child, page_size, page_token, user_id_type, department_id_type } = params
    if (!department_id) {
      return { success: false, error: 'department_id is required' }
    }
    const result = await axios.get(
      `https://open.feishu.cn/open-apis/contact/v3/departments/${department_id}/children`,
      {
        headers: { Authorization: `Bearer ${user_token || ''}` },
        params: {
          fetch_child: fetch_child,
          page_size: page_size || 50,
          page_token: page_token || undefined,
          user_id_type: user_id_type || 'open_id',
          department_id_type: department_id_type || 'open_department_id',
        },
      }
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
  },
  {
    type: "function",
    function: {
      name: "getDepartmentChildren",
      description: "获取指定部门的子部门列表",
      parameters: {
        type: "object",
        properties: {
          user_token: { type: "string", description: "飞书用户访问令牌（可选，不填则使用 tenant token）" },
          department_id: { type: "string", description: "部门 ID，根部门为 0" },
          fetch_child: { type: "boolean", description: "是否递归获取子部门，默认 false" },
          page_size: { type: "number", description: "每页数量，最大 50，默认 50" },
          page_token: { type: "string", description: "分页标记" },
          user_id_type: { type: "string", enum: ["open_id", "union_id", "user_id"], description: "用户 ID 类型" },
          department_id_type: { type: "string", enum: ["open_department_id", "department_id"], description: "部门 ID 类型" }
        },
        required: ["department_id"]
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
  getDepartmentChildren,
  createBlock,
  createNestedBlocks,
  getBlock,
  getChildren,
  updateBlock,
  batchUpdateBlocks,
  deleteBlock,
  // Apifox 工具
  importOpenApiFromUrl: importOpenApiFromUrlTool,
  importOpenApiFromSpec: importOpenApiFromSpecTool,
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

// ── Apifox 工具函数 ──────────────────────────────────────────────────

export async function importOpenApiFromUrlTool(params: ToolParams): Promise<ToolResult> {
  try {
    const {
      project_id,
      openapi_url,
      target_endpoint_folder_id,
      target_schema_folder_id,
      endpoint_overwrite_behavior,
      schema_overwrite_behavior,
      update_folder_of_changed_endpoint,
      prepend_base_path,
    } = params

    if (!project_id) {
      return { success: false, error: 'project_id is required' }
    }

    if (!openapi_url) {
      return { success: false, error: 'openapi_url is required' }
    }

    const options: ApifoxImportOptions = {
      targetEndpointFolderId: target_endpoint_folder_id,
      targetSchemaFolderId: target_schema_folder_id,
      endpointOverwriteBehavior: endpoint_overwrite_behavior as any,
      schemaOverwriteBehavior: schema_overwrite_behavior as any,
      updateFolderOfChangedEndpoint: update_folder_of_changed_endpoint,
      prependBasePath: prepend_base_path,
    }

    const result = await importOpenApiFromUrl(project_id, openapi_url, options)

    if (result.success) {
      const counters = result.data?.counters
      return {
        success: true,
        data: {
          message: `导入成功：创建 ${counters?.endpointCreated || 0} 个接口，更新 ${counters?.endpointUpdated || 0} 个接口`,
          counters,
        },
      }
    } else {
      return { success: false, error: result.error }
    }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

export async function importOpenApiFromSpecTool(params: ToolParams): Promise<ToolResult> {
  try {
    const {
      project_id,
      openapi_spec,
      target_endpoint_folder_id,
      target_schema_folder_id,
      endpoint_overwrite_behavior,
      schema_overwrite_behavior,
      update_folder_of_changed_endpoint,
      prepend_base_path,
    } = params

    if (!project_id) {
      return { success: false, error: 'project_id is required' }
    }

    if (!openapi_spec || typeof openapi_spec !== 'object') {
      return { success: false, error: 'openapi_spec is required and must be an object (OpenAPI 规范 JSON)' }
    }

    const options: ApifoxImportOptions = {
      targetEndpointFolderId: target_endpoint_folder_id,
      targetSchemaFolderId: target_schema_folder_id,
      endpointOverwriteBehavior: endpoint_overwrite_behavior as any,
      schemaOverwriteBehavior: schema_overwrite_behavior as any,
      updateFolderOfChangedEndpoint: update_folder_of_changed_endpoint,
      prependBasePath: prepend_base_path,
    }

    const result = await importOpenApiFromSpec(project_id, openapi_spec, options)

    if (result.success) {
      const counters = result.data?.counters
      return {
        success: true,
        data: {
          message: `导入成功：创建 ${counters?.endpointCreated || 0} 个接口，更新 ${counters?.endpointUpdated || 0} 个接口`,
          counters,
        },
      }
    } else {
      return { success: false, error: result.error }
    }
  } catch (err) {
    const error = err as { message?: string; response?: { data?: unknown } }
    return { success: false, error: error.message, data: error.response?.data }
  }
}

const apifoxTools = [
  {
    type: "function",
    function: {
      name: "importOpenApiFromUrl",
      description: "从 URL 导入 OpenAPI/Swagger 规范到 Apifox 项目。当需要从远程 URL（如 Swagger Hub、GitHub 等）获取 OpenAPI 规范时使用此工具。",
      parameters: {
        type: "object",
        properties: {
          project_id: { type: "string", description: "Apifox 项目 ID" },
          openapi_url: { type: "string", description: "OpenAPI/Swagger 规范文件的 URL" },
          target_endpoint_folder_id: { type: "number", description: "目标接口文件夹 ID（可选）" },
          target_schema_folder_id: { type: "number", description: "目标 Schema 文件夹 ID（可选）" },
          endpoint_overwrite_behavior: {
            type: "string",
            enum: ["OVERWRITE_EXISTING", "AUTO_MERGE", "KEEP_EXISTING", "CREATE_NEW"],
            description: "接口覆盖行为: OVERWRITE_EXISTING(覆盖现有)、AUTO_MERGE(自动合并)、KEEP_EXISTING(保留现有)、CREATE_NEW(创建新)"
          },
          schema_overwrite_behavior: {
            type: "string",
            enum: ["OVERWRITE_EXISTING", "AUTO_MERGE", "KEEP_EXISTING", "CREATE_NEW"],
            description: "Schema 覆盖行为"
          },
          update_folder_of_changed_endpoint: { type: "boolean", description: "是否更新变更接口的文件夹，默认 true" },
          prepend_base_path: { type: "boolean", description: "是否在路径前追加 basePath，默认 false" },
        },
        required: ["project_id", "openapi_url"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "importOpenApiFromSpec",
      description: "从 OpenAPI JSON 规范对象导入到 Apifox 项目。当已经生成了 OpenAPI 规范 JSON 对象时使用此工具。",
      parameters: {
        type: "object",
        properties: {
          project_id: { type: "string", description: "Apifox 项目 ID" },
          openapi_spec: {
            type: "object",
            description: "OpenAPI 3.0 规范对象（完整的 OpenAPI JSON 对象，包含 openapi、info、paths、components 等字段）"
          },
          target_endpoint_folder_id: { type: "number", description: "目标接口文件夹 ID（可选）" },
          target_schema_folder_id: { type: "number", description: "目标 Schema 文件夹 ID（可选）" },
          endpoint_overwrite_behavior: {
            type: "string",
            enum: ["deleteUnmatchedResources", "ignoreUnmatchedResources", "coverUnmatchedResources"],
            description: "接口覆盖行为"
          },
          schema_overwrite_behavior: {
            type: "string",
            enum: ["KEEP_EXISTING", "COVER_EXISTING"],
            description: "Schema 覆盖行为"
          },
          update_folder_of_changed_endpoint: { type: "boolean", description: "是否更新变更接口的文件夹，默认 true" },
          prepend_base_path: { type: "boolean", description: "是否在路径前追加 basePath，默认 false" },
        },
        required: ["project_id", "openapi_spec"],
      },
    },
  },
]

const allTools = [...tools, ...blockTools, ...apifoxTools]

export default allTools
