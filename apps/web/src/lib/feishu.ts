const DEFAULT_FEISHU_SDK_URL = 'https://lf-scm-cn.feishucdn.com/lark/op/h5-js-sdk-1.5.44.js'

type RequestAccessSuccess = {
  code: string
}

type RequestAccessFailure = {
  errno?: number
  errString?: string
}

function sdkURL(): string {
  return import.meta.env.VITE_FEISHU_SDK_URL || DEFAULT_FEISHU_SDK_URL
}

export function isFeishuWebApp(): boolean {
  if (typeof window === 'undefined') {
    return false
  }
  if (window.tt) {
    return true
  }
  return /feishu|lark/i.test(window.navigator.userAgent)
}

export async function ensureFeishuSDKLoaded(): Promise<void> {
  if (typeof window === 'undefined') {
    throw new Error('当前环境不支持飞书网页应用免登。')
  }
  if (window.tt) {
    return
  }

  await new Promise<void>((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>('script[data-feishu-sdk="true"]')
    if (existing) {
      existing.addEventListener('load', () => resolve(), { once: true })
      existing.addEventListener('error', () => reject(new Error('飞书 JS SDK 加载失败。')), { once: true })
      return
    }

    const script = document.createElement('script')
    script.src = sdkURL()
    script.async = true
    script.dataset.feishuSdk = 'true'
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('飞书 JS SDK 加载失败。'))
    document.head.appendChild(script)
  })

  if (!window.tt) {
    throw new Error('飞书 JS SDK 已加载，但未检测到可用的 tt API。')
  }
}

export async function requestFeishuAuthCode(appId: string): Promise<string> {
  if (!appId.trim()) {
    throw new Error('缺少飞书 App ID，无法发起免登。')
  }
  if (!isFeishuWebApp()) {
    throw new Error('请在飞书网页应用内打开当前页面以完成免登。')
  }

  await ensureFeishuSDKLoaded()
  const tt = window.tt
  if (!tt) {
    throw new Error('飞书 JS SDK 不可用。')
  }

  const code = await new Promise<string>((resolve, reject) => {
    const onSuccess = (result: RequestAccessSuccess) => {
      if (!result?.code) {
        reject(new Error('飞书未返回有效的预授权码。'))
        return
      }
      resolve(result.code)
    }

    const onFailure = (error: RequestAccessFailure) => {
      reject(new Error(error.errString || '飞书免登授权失败。'))
    }

    const fallbackToRequestAuthCode = () => {
      if (!tt.requestAuthCode) {
        reject(new Error('当前飞书客户端不支持免登取码。'))
        return
      }
      tt.requestAuthCode({
        appId,
        success: onSuccess,
        fail: onFailure,
      })
    }

    if (tt.requestAccess) {
      tt.requestAccess({
        appID: appId,
        scopeList: [],
        success: onSuccess,
        fail: (error: RequestAccessFailure) => {
          if (error?.errno === 103) {
            fallbackToRequestAuthCode()
            return
          }
          onFailure(error)
        },
      })
      return
    }

    fallbackToRequestAuthCode()
  })

  return code
}
