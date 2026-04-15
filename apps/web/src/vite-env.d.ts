/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  readonly VITE_FEISHU_SDK_URL?: string
}

type FeishuRequestAccessOptions = {
  appID: string
  scopeList: string[]
  success: (result: { code: string }) => void
  fail: (error: { errno?: number; errString?: string }) => void
}

type FeishuRequestAuthCodeOptions = {
  appId: string
  success: (result: { code: string }) => void
  fail: (error: { errno?: number; errString?: string }) => void
}

interface Window {
  tt?: {
    requestAccess?: (options: FeishuRequestAccessOptions) => void
    requestAuthCode?: (options: FeishuRequestAuthCodeOptions) => void
  }
}
