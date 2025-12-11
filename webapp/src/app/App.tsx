import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { initTelegramWebApp } from '@/shared/lib/telegram'
import { useAuthStore } from '@/stores/authStore'
import Dashboard from '@/pages/Dashboard'
import TodoPage from '@/pages/TodoPage'
import FinancePage from '@/pages/FinancePage'
import CreditsPage from '@/pages/CreditsPage'

function App() {
  const { isAuthenticated, login, isLoading } = useAuthStore()

  useEffect(() => {
    initTelegramWebApp()
    
    // Автоматическая авторизация при загрузке
    if (!isAuthenticated && !isLoading) {
      login()
    }
  }, [isAuthenticated, isLoading, login])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-hint">Загрузка...</div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-hint">Ошибка авторизации</div>
      </div>
    )
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/todo" element={<TodoPage />} />
        <Route path="/finance" element={<FinancePage />} />
        <Route path="/credits" element={<CreditsPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App

