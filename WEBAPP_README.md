# Telegram Mini App - Веб-версия бота

Полнофункциональная веб-версия Telegram Assistant Bot, встроенная как Telegram Mini App.

## Архитектура

### Frontend (webapp/)
- **Vite + React + TypeScript** — современный стек разработки
- **Zustand** — управление состоянием
- **TailwindCSS** — стилизация с поддержкой Telegram тем
- **Framer Motion** — плавные анимации
- **Telegram WebApp SDK** — интеграция с Telegram

### Backend API
- **REST API** на Go с JWT авторизацией
- **WebSocket** для real-time синхронизации
- **Telegram initData** проверка через HMAC
- **CORS** настроен для Telegram доменов

## Установка и запуск

### 1. Backend

```bash
# Установка зависимостей
go mod download

# Настройка .env
# Добавьте в .env:
# JWT_SECRET=your-secret-key-here
# API_PORT=8080
# WEBAPP_URL=https://your-domain.com/webapp

# Запуск
go run main.go
```

Backend запустится на порту 8080 (или указанном в API_PORT).

### 2. Frontend

```bash
cd webapp

# Установка зависимостей
npm install

# Настройка .env
# Создайте webapp/.env:
# VITE_API_URL=http://localhost:8080

# Запуск dev сервера
npm run dev
```

Frontend запустится на `http://localhost:5173`.

### 3. Production сборка

```bash
cd webapp
npm run build
```

Собранные файлы будут в `webapp/dist/`. Разместите их на веб-сервере (nginx, Apache, или любой статический хостинг).

## Настройка Telegram Mini App

1. **Создайте бота** через [@BotFather](https://t.me/BotFather)

2. **Настройте WebApp URL**:
   ```
   /newapp
   Выберите бота
   Введите название приложения
   Введите URL: https://your-domain.com/webapp
   ```

3. **Добавьте Menu Button** (автоматически через код, или вручную):
   ```
   /mybots → Ваш бот → Bot Settings → Menu Button → Configure
   ```

4. **Для локальной разработки** используйте ngrok или аналогичный сервис:
   ```bash
   ngrok http 5173
   # Используйте полученный HTTPS URL в BotFather
   ```

## API Endpoints

### Авторизация
- `POST /api/auth` — авторизация через Telegram initData
  ```json
  {
    "initData": "query_id=..."
  }
  ```
  Возвращает: `{ "token": "...", "user_id": 123 }`

### TODO
- `GET /api/todo/list` — список задач
- `POST /api/todo/add` — добавить задачу
- `POST /api/todo/delete` — удалить задачу

### Finance
- `GET /api/finance/list` — список операций
- `POST /api/finance/add` — добавить операцию
- `POST /api/finance/stats` — статистика
- `POST /api/finance/recurring/add` — добавить регулярный платёж

### Credits
- `GET /api/credits/list` — список кредитов
- `POST /api/credits/add` — добавить кредит
- `POST /api/credits/close` — закрыть кредит
- `GET /api/credits/schedule?id=123` — график платежей

### WebSocket
- `ws://localhost:8080/ws?token=JWT_TOKEN` — real-time обновления

Все запросы (кроме `/api/auth`) требуют заголовок:
```
Authorization: Bearer <JWT_TOKEN>
```

## Безопасность

### Telegram initData проверка
- HMAC-SHA256 проверка подписи
- Проверка времени (auth_date не старше 24 часов)
- Извлечение user_id из проверенных данных

### JWT
- Секретный ключ из `JWT_SECRET` в .env
- Срок действия: 7 дней
- Алгоритм: HS256

### CORS
Разрешены только:
- `https://*.telegram.org`
- `http://localhost:5173` (для разработки)

## Синхронизация Бот ⟷ WebApp

### WebSocket Events

События синхронизации:
```json
{
  "type": "todo_added" | "finance_added" | "credit_added" | ...,
  "user_id": 123,
  "data": { ... },
  "timestamp": 1234567890
}
```

### Как это работает

1. **Действие в WebApp**:
   - Запрос к API
   - Создание записи в БД
   - Broadcast через WebSocket Hub
   - Все открытые WebApp клиенты получают обновление

2. **Действие в боте**:
   - Обработка через Handler
   - Обновление через сервис
   - Broadcast через WebSocket Hub (если Hub доступен)
   - WebApp получает обновление в реальном времени

## Структура проекта

```
tg_bot_asist/
├── webapp/                 # Frontend
│   ├── src/
│   │   ├── app/           # App компонент
│   │   ├── pages/         # Страницы
│   │   ├── stores/        # Zustand stores
│   │   ├── shared/        # Общие утилиты
│   │   └── main.tsx       # Entry point
│   ├── package.json
│   └── vite.config.ts
├── internal/
│   ├── api/               # Backend API
│   │   ├── handlers/      # HTTP handlers
│   │   ├── middleware/    # Middleware (auth, CORS, etc.)
│   │   ├── auth/          # JWT + Telegram auth
│   │   ├── websocket/     # WebSocket Hub
│   │   └── router.go      # API роутер
│   └── ...                # Существующие модули
└── main.go                # Точка входа (бот + API)
```

## Разработка

### Добавление нового API endpoint

1. Создайте handler в `internal/api/handlers/`
2. Добавьте маршрут в `internal/api/router.go`
3. Обновите frontend store и API client

### Добавление новой страницы

1. Создайте компонент в `webapp/src/pages/`
2. Добавьте маршрут в `webapp/src/app/App.tsx`
3. Добавьте ссылку в Dashboard

## HTTPS для локальной разработки

Используйте [mkcert](https://github.com/FiloSottile/mkcert):

```bash
# Установка mkcert
brew install mkcert  # macOS
# или следуйте инструкциям для вашей ОС

# Создание локального CA
mkcert -install

# Генерация сертификатов
cd webapp
mkcert localhost 127.0.0.1 ::1

# Обновите vite.config.ts для использования HTTPS
```

## Troubleshooting

### WebApp не открывается
- Проверьте, что URL в BotFather правильный
- Убедитесь, что используется HTTPS (Telegram требует HTTPS)
- Проверьте CORS настройки

### Авторизация не работает
- Проверьте, что `TELEGRAM_BOT_API` в .env правильный
- Убедитесь, что initData передаётся корректно
- Проверьте логи backend

### WebSocket не подключается
- Проверьте, что токен передаётся в query параметре
- Убедитесь, что Hub запущен (должен быть в main.go)
- Проверьте логи подключений

## Production Deployment

1. **Backend**: Разверните на сервере (Docker, systemd, etc.)
2. **Frontend**: Соберите и разместите на CDN/статический хостинг
3. **HTTPS**: Обязательно используйте HTTPS (Let's Encrypt)
4. **Environment**: Настройте все переменные окружения
5. **Monitoring**: Настройте логирование и мониторинг

## TODO

- [ ] Полная реализация Finance и Credits страниц
- [ ] WebSocket синхронизация с ботом
- [ ] Оптимизация bundle size
- [ ] PWA поддержка
- [ ] Офлайн режим
- [ ] Push уведомления

## Лицензия

MIT

