import { create } from 'zustand'
import { api } from '@/shared/lib/api'

export interface Todo {
  id: number
  user_id: number
  title: string
  description: string
  due_date: string | null
  status: string
  created_at: string
}

interface TodoState {
  todos: Todo[]
  isLoading: boolean
  error: string | null
  fetchTodos: () => Promise<void>
  addTodo: (todo: { title: string; description?: string; due_date?: string }) => Promise<void>
  deleteTodo: (id: number) => Promise<void>
}

export const useTodoStore = create<TodoState>((set, get) => ({
  todos: [],
  isLoading: false,
  error: null,

  fetchTodos: async () => {
    set({ isLoading: true, error: null })
    try {
      const response = await api.getTodos()
      set({ todos: response.data, isLoading: false })
    } catch (error: any) {
      set({ error: error.message, isLoading: false })
    }
  },

  addTodo: async (todo) => {
    try {
      await api.addTodo(todo)
      await get().fetchTodos()
    } catch (error: any) {
      set({ error: error.message })
    }
  },

  deleteTodo: async (id) => {
    try {
      await api.deleteTodo(id)
      set({ todos: get().todos.filter((t) => t.id !== id) })
    } catch (error: any) {
      set({ error: error.message })
    }
  },
}))

