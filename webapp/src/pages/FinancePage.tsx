import { useEffect } from 'react'
import { tg } from '@/shared/lib/telegram'

export default function FinancePage() {
  useEffect(() => {
    if (tg?.BackButton) {
      tg.BackButton.show()
      tg.BackButton.onClick(() => {
        window.history.back()
      })
    }

    return () => {
      if (tg?.BackButton) {
        tg.BackButton.hide()
      }
    }
  }, [])

  return (
    <div className="min-h-screen bg-secondary-bg p-4">
      <div className="max-w-md mx-auto">
        <h1 className="text-2xl font-bold mb-6 text-text">Финансы</h1>
        <div className="text-hint">Страница в разработке</div>
      </div>
    </div>
  )
}

