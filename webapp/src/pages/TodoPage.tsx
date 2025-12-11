import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useTodoStore, Todo } from '@/stores/todoStore'
import { tg } from '@/shared/lib/telegram'

export default function TodoPage() {
  const { todos, isLoading, fetchTodos, addTodo, deleteTodo } = useTodoStore()
  const [showAdd, setShowAdd] = useState(false)
  const [title, setTitle] = useState('')

  useEffect(() => {
    fetchTodos()

    // Настройка Telegram BackButton
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
  }, [fetchTodos])

  const handleAdd = async () => {
    if (!title.trim()) return

    await addTodo({ title: title.trim() })
    setTitle('')
    setShowAdd(false)
    
    if (tg?.HapticFeedback) {
      tg.HapticFeedback.notificationOccurred('success')
    }
  }

  const handleDelete = async (id: number) => {
    await deleteTodo(id)
    if (tg?.HapticFeedback) {
      tg.HapticFeedback.impactOccurred('medium')
    }
  }

  return (
    <div className="min-h-screen bg-secondary-bg">
      <div className="max-w-md mx-auto p-4">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-text">Задачи</h1>
          <button
            onClick={() => setShowAdd(true)}
            className="w-10 h-10 rounded-full bg-button text-button-text flex items-center justify-center text-xl font-bold"
          >
            +
          </button>
        </div>

        {isLoading ? (
          <div className="text-center text-hint py-8">Загрузка...</div>
        ) : todos.length === 0 ? (
          <div className="text-center text-hint py-8">Нет задач</div>
        ) : (
          <div className="space-y-2">
            <AnimatePresence>
              {todos.map((todo) => (
                <motion.div
                  key={todo.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: 20 }}
                  className="p-4 rounded-xl bg-bg shadow-sm"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h3 className="font-semibold text-text">{todo.title}</h3>
                      {todo.description && (
                        <p className="text-sm text-hint mt-1">{todo.description}</p>
                      )}
                    </div>
                    <button
                      onClick={() => handleDelete(todo.id)}
                      className="ml-2 text-red-500"
                    >
                      ✕
                    </button>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>

      {/* Add Modal */}
      <AnimatePresence>
        {showAdd && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/50 flex items-end z-50"
            onClick={() => setShowAdd(false)}
          >
            <motion.div
              initial={{ y: '100%' }}
              animate={{ y: 0 }}
              exit={{ y: '100%' }}
              transition={{ type: 'spring', damping: 25 }}
              className="w-full bg-bg rounded-t-2xl p-6"
              onClick={(e) => e.stopPropagation()}
            >
              <h2 className="text-xl font-bold mb-4 text-text">Новая задача</h2>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Название задачи"
                className="w-full p-3 rounded-lg bg-secondary-bg text-text border-none outline-none mb-4"
                autoFocus
              />
              <div className="flex gap-2">
                <button
                  onClick={() => setShowAdd(false)}
                  className="flex-1 p-3 rounded-lg bg-secondary-bg text-text"
                >
                  Отмена
                </button>
                <button
                  onClick={handleAdd}
                  className="flex-1 p-3 rounded-lg bg-button text-button-text"
                >
                  Добавить
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

