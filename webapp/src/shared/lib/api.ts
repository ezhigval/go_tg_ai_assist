import axios, { AxiosInstance } from 'axios'
import { getInitData } from './telegram'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

class ApiClient {
  private client: AxiosInstance
  private token: string | null = null

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Interceptor для добавления токена
    this.client.interceptors.request.use((config) => {
      if (this.token) {
        config.headers.Authorization = `Bearer ${this.token}`
      }
      return config
    })

    // Interceptor для обработки ошибок
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Токен истёк, нужно переавторизоваться
          this.token = null
          localStorage.removeItem('token')
        }
        return Promise.reject(error)
      }
    )
  }

  setToken(token: string) {
    this.token = token
    localStorage.setItem('token', token)
  }

  getToken() {
    return this.token || localStorage.getItem('token')
  }

  async auth() {
    const initData = getInitData()
    if (!initData) {
      throw new Error('Telegram initData not available')
    }

    const response = await this.client.post('/api/auth', { initData })
    const { token } = response.data
    this.setToken(token)
    return response.data
  }

  // User
  async getMe() {
    return this.client.get('/api/user/me')
  }

  // TODO
  async getTodos() {
    return this.client.get('/api/todo/list')
  }

  async addTodo(data: { title: string; description?: string; due_date?: string }) {
    return this.client.post('/api/todo/add', data)
  }

  async deleteTodo(id: number) {
    return this.client.post('/api/todo/delete', { id })
  }

  // Finance
  async getFinanceList() {
    return this.client.get('/api/finance/list')
  }

  async addFinance(data: {
    amount: number
    category: string
    type: 'income' | 'expense'
    note?: string
  }) {
    return this.client.post('/api/finance/add', data)
  }

  async getFinanceStats() {
    return this.client.post('/api/finance/stats')
  }

  async addRecurring(data: {
    title: string
    amount: number
    category: string
    period: string
    next_payment: string
  }) {
    return this.client.post('/api/finance/recurring/add', data)
  }

  // Credits
  async getCredits() {
    return this.client.get('/api/credits/list')
  }

  async addCredit(data: {
    title: string
    principal: number
    rate: number
    months: number
  }) {
    return this.client.post('/api/credits/add', data)
  }

  async closeCredit(id: number) {
    return this.client.post('/api/credits/close', { id })
  }

  async getCreditSchedule(id: number) {
    return this.client.get(`/api/credits/schedule?id=${id}`)
  }
}

export const api = new ApiClient()

