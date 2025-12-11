// Telegram WebApp SDK типы и утилиты
declare global {
  interface Window {
    Telegram?: {
      WebApp: {
        initData: string
        initDataUnsafe: {
          user?: {
            id: number
            first_name?: string
            last_name?: string
            username?: string
          }
          auth_date: number
        }
        version: string
        platform: string
        colorScheme: 'light' | 'dark'
        themeParams: {
          bg_color?: string
          text_color?: string
          hint_color?: string
          link_color?: string
          button_color?: string
          button_text_color?: string
          secondary_bg_color?: string
        }
        isExpanded: boolean
        viewportHeight: number
        viewportStableHeight: number
        headerColor: string
        backgroundColor: string
        BackButton: {
          isVisible: boolean
          onClick: (callback: () => void) => void
          offClick: (callback: () => void) => void
          show: () => void
          hide: () => void
        }
        MainButton: {
          text: string
          color: string
          textColor: string
          isVisible: boolean
          isActive: boolean
          isProgressVisible: boolean
          setText: (text: string) => void
          onClick: (callback: () => void) => void
          offClick: (callback: () => void) => void
          show: () => void
          hide: () => void
          enable: () => void
          disable: () => void
          showProgress: (leaveActive?: boolean) => void
          hideProgress: () => void
          setParams: (params: {
            text?: string
            color?: string
            text_color?: string
            is_active?: boolean
            is_visible?: boolean
          }) => void
        }
        HapticFeedback: {
          impactOccurred: (style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft') => void
          notificationOccurred: (type: 'error' | 'success' | 'warning') => void
          selectionChanged: () => void
        }
        ready: () => void
        expand: () => void
        close: () => void
        sendData: (data: string) => void
        openLink: (url: string, options?: { try_instant_view?: boolean }) => void
        openTelegramLink: (url: string) => void
        openInvoice: (url: string, callback?: (status: string) => void) => void
        showPopup: (params: {
          title?: string
          message: string
          buttons?: Array<{
            id?: string
            type?: 'default' | 'ok' | 'close' | 'cancel' | 'destructive'
            text: string
          }>
        }, callback?: (id: string) => void) => void
        showAlert: (message: string, callback?: () => void) => void
        showConfirm: (message: string, callback?: (confirmed: boolean) => void) => void
        showScanQrPopup: (params: {
          text?: string
        }, callback?: (data: string) => void) => void
        closeScanQrPopup: () => void
        readTextFromClipboard: (callback?: (text: string) => void) => void
        requestWriteAccess: (callback?: (granted: boolean) => void) => void
        requestContact: (callback?: (granted: boolean) => void) => void
      }
    }
  }
}

export const tg = window.Telegram?.WebApp

export const initTelegramWebApp = () => {
  if (!tg) {
    console.warn('Telegram WebApp not available')
    return
  }

  tg.ready()
  tg.expand()

  // Применяем тему Telegram
  if (tg.themeParams) {
    const root = document.documentElement
    if (tg.themeParams.bg_color) {
      root.style.setProperty('--tg-theme-bg-color', tg.themeParams.bg_color)
    }
    if (tg.themeParams.text_color) {
      root.style.setProperty('--tg-theme-text-color', tg.themeParams.text_color)
    }
    if (tg.themeParams.hint_color) {
      root.style.setProperty('--tg-theme-hint-color', tg.themeParams.hint_color)
    }
    if (tg.themeParams.link_color) {
      root.style.setProperty('--tg-theme-link-color', tg.themeParams.link_color)
    }
    if (tg.themeParams.button_color) {
      root.style.setProperty('--tg-theme-button-color', tg.themeParams.button_color)
    }
    if (tg.themeParams.button_text_color) {
      root.style.setProperty('--tg-theme-button-text-color', tg.themeParams.button_text_color)
    }
    if (tg.themeParams.secondary_bg_color) {
      root.style.setProperty('--tg-theme-secondary-bg-color', tg.themeParams.secondary_bg_color)
    }
  }

  return tg
}

export const getUser = () => {
  return tg?.initDataUnsafe?.user
}

export const getInitData = () => {
  return tg?.initData || ''
}

