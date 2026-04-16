/**
 * Axios 请求封装
 * 包含请求拦截器、响应拦截器、错误处理
 */

import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'

// ── 配置 ────────────────────────────────────────────────────────────────────

export interface RequestConfig {
  /** 基础 URL */
  baseURL?: string
  /** 请求超时(ms) */
  timeout?: number
  /** 是否携带凭证(cookie) */
  withCredentials?: boolean
}

const defaultConfig: RequestConfig = {
  baseURL: process.env.API_BASE_URL ?? 'http://localhost:8080',
  timeout: 30000,
  withCredentials: true,
}

// ── 创建实例 ─────────────────────────────────────────────────────────────────

export const http: AxiosInstance = axios.create(defaultConfig)

// ── 请求拦截器 ───────────────────────────────────────────────────────────────

http.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 1. 添加时间戳防止缓存(GET 请求)
    if (config.method === 'get' && config.params) {
      config.params._t = Date.now()
    } else if (config.method === 'get') {
      config.params = { _t: Date.now() }
    }

    // 2. 从 localStorage/Session 获取 Token(前端用)
    // if (typeof window !== 'undefined') {
    //   const token = localStorage.getItem('token') || sessionStorage.getItem('token')
    //   if (token && config.headers) {
    //     config.headers.Authorization = `Bearer ${token}`
    //   }
    // }

    // 3. 从 Cookie 获取 Session(服务端用)
    // if (config.withCredentials) {
    //   // axios 会自动携带 cookie
    // }

    console.debug(`[HTTP] → ${config.method?.toUpperCase()} ${config.url}`, {
      params: config.params,
      data: config.data ? '[body]' : undefined,
    })

    return config
  },
  (error: AxiosError) => {
    console.error('[HTTP] 请求配置错误:', error)
    return Promise.reject(error)
  }
)

// ── 响应拦截器 ───────────────────────────────────────────────────────────────

http.interceptors.response.use(
  (response: AxiosResponse) => {
    console.debug(
      `[HTTP] ← ${response.config.method?.toUpperCase()} ${response.config.url}`,
      response.status,
      response.statusText
    )

    // 如果后端返回的数据包装在 data 字段中，直接返回
    if (response.data && typeof response.data === 'object' && 'data' in response.data) {
      return response.data as unknown as AxiosResponse
    }

    return response
  },
  async (error: AxiosError) => {
    const config = error.config as InternalAxiosRequestConfig & { _retry?: number }

    console.error('[HTTP] ← 响应错误:', {
      url: config?.url,
      status: error.response?.status,
      message: error.message,
    })

    // 1. 处理 401 未授权
    if (error.response?.status === 401) {
      // 避免重复跳转
      if (config && !config._retry) {
        config._retry = 1

        // 清除登录状态并跳转登录页
        if (typeof window !== 'undefined') {
          // localStorage.removeItem('token')
          // window.location.href = '/login'
        }
      }
    }

    // 2. 处理 403 禁止访问
    if (error.response?.status === 403) {
      console.warn('[HTTP] 无权限访问:', config?.url)
    }

    // 3. 处理 500 服务器错误
    if (error.response?.status === 500) {
      console.error('[HTTP] 服务器内部错误:', config?.url)
    }

    // 4. 处理网络错误
    if (error.code === 'ECONNABORTED') {
      console.error('[HTTP] 请求超时:', config?.url)
    }

    if (!error.response) {
      console.error('[HTTP] 网络错误: 无法连接到服务器')
    }

    return Promise.reject(error)
  }
)

// ── 辅助方法 ─────────────────────────────────────────────────────────────────

/**
 * 创建带有默认配置的 axios 实例
 * 用于需要不同配置的场景(如文件上传)
 */
export function createHttpClient(config: RequestConfig): AxiosInstance {
  return axios.create({ ...defaultConfig, ...config })
}

/**
 * 格式化错误信息
 */
export function formatError(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError
    if (axiosError.response?.data) {
      const data = axiosError.response.data as Record<string, unknown>
      return (data.message as string) || (data.error as string) || axiosError.message
    }
    return axiosError.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return '未知错误'
}
