# Архитектура проекта Jarvis Assistant (Go + Telegram + Web + AI)

Этот документ фиксирует целевую архитектуру проекта — твоего персонального Jarvis-а, который одновременно является демонстрацией навыков Senior+ Go backend инженера.

---

## 1. Общая идея

Jarvis Assistant — это мульти‑модульная система, которая:

- живёт в **Telegram-чате** как умный ассистент (текст + голос);
- имеет **Telegram Mini App** (web-интерфейс);
- интегрируется с **календарём** (iCloud / любой через ICS);
- умеет работать с **задачами, напоминаниями, финансами, кредитами, почтой**;
- использует **ИИ** для понимания естественного языка и фильтрации информации;
- работает **автономно 24/7**, реагирует не только на команды, но и на события (наступил день платежа, пришло важное письмо и т.д.).

Архитектура спроектирована как **модульный монолит**, который можно при желании раздробить на микросервисы, но на данном этапе остаётся в одном репозитории и одном кодбэйзе.

---

## 2. Логическая архитектура (высокий уровень)

### 2.1. Слои

1. **Внешние интерфейсы (Edge / Adapters)**  
   - Telegram Bot Gateway (webhook сервер);
   - Telegram Mini App (web UI + HTTP API);
   - ICS Calendar endpoint (для iCloud/календарей);
   - Email интеграции (Gmail/Yandex/iCloud через IMAP);
   - AI/ASR провайдеры (LLM, Whisper и т.д.).

2. **Ядро (Domain / Core)**  
   Чистый бизнес‑слой, ничего не знает о Telegram, HTTP, IMAP и пр.  
   Модули:
   - user (пользователь, настройки, персональность ассистента);
   - reminder/calendar (напоминания, события, календарь);
   - task (задачи/проекты);
   - finance (доходы, расходы, бюджеты, регулярные платежи);
   - credits (кредиты, графики платежей);
   - mail (абстракция над почтой, “важные письма”);
   - events (внутренние доменные события).

3. **Инфраструктура (Infra)**  
   - PostgreSQL, схемы, миграции;
   - Redis/NATS (очереди, pub/sub);
   - Logger + Metrics + Tracing;
   - Конфигурация (env, файлы).

---

## 3. Процессы (бинарники)

### 3.1. `/cmd/bot` — Telegram Bot Gateway

**Назначение:** принимать обновления Telegram и отдавать быстрый ответ пользователю, не блокируясь на тяжёлых операциях.

Функции:

- HTTP endpoint для Telegram Webhook;
- валидация запросов (token, IP-whitelist / secret);
- базовый rate limiting;
- парсинг апдейтов;
- преобразование в внутренние события (`MessageReceived`, `VoiceReceived` и т.д.) и отправка в очередь (Redis/NATS);
- отправка простых ответов (например, “принято, работаю над запросом”).

**Навыки Go:**  
HTTP, middlewares, контекст, работа с JSON, конкурентность, очереди.

---

### 3.2. `/cmd/worker` — Brain / Worker-процесс

**Назначение:** “мозг” системы. Обрабатывает события из очередей и взаимодействует с доменными сервисами и AI.

Функции:

- подписка на очередь событий (Redis Streams/NATS/Kafka);
- маршрутизация событий в обработчики;
- вызов AI/ASR провайдеров;
- вызов доменных сервисов (`ReminderService`, `FinanceService`, `CreditsService`, `MailService`);
- генерация ответов для пользователя (через Telegram API);
- публикация внутренних событий (например, `ReminderCreated`, `PaymentDue`).

**Навыки Go:**  
worker-пулы, goroutines, channels, retry-логика, idempotency, error handling, backoff стратегии.

---

### 3.3. `/cmd/api` — HTTP API для Web/Mini App

**Назначение:** backend для Telegram Mini App и внешних интеграций.

Функции:

- REST API:  
  - `/api/me`
  - `/api/reminders`
  - `/api/tasks`
  - `/api/finance`
  - `/api/credits`
  - `/api/system/status`
- авторизация через Telegram WebApp initData (HMAC-подпись);
- JWT для внешних клиентов (в перспективе);
- WebSocket для live обновления UI.

**Навыки Go:**  
REST-дизайн, auth, HMAC, WebSocket, сериализация/десериализация, CORS, безопасный HTTPS.

---

### 3.4. `/cmd/scheduler` — планировщик событий

**Назначение:** запускать периодические задачи.

Функции:

- проверка “наступило ли время напоминания”;
- генерация событий `ReminderDue`;
- обработка регулярных платежей (финансы, кредиты);
- периодический обход почты (mail polling);
- “health” внутренних сервисов (опционально).

**Навыки Go:**  
cron‑паттерны, таймеры, context, доступ к БД, транзакции, планирование.

---

## 4. Структура репозитория (целевое дерево)

```bash
/tg_ai_assistant
  /cmd
    /bot
      main.go
    /api
      main.go
    /worker
      main.go
    /scheduler
      main.go
    /migrate
      main.go
  /internal
    /app
      server.go        # общая инициализация, DI
    /domain
      /user
        service.go
        models.go
      /reminder
        service.go
        models.go
      /task
        service.go
        models.go
      /finance
        service.go
        models.go
      /credits
        service.go
        models.go
      /mail
        service.go
        models.go
      /events
        bus.go
        event_types.go
    /adapters
      /telegrambot
        handler.go
        router.go
        client.go
      /webapi
        http.go
        middleware.go
        handlers_reminders.go
        handlers_finance.go
        handlers_credits.go
      /calendar
        ics_handler.go
        ics_generator.go
      /ai
        interpreter.go
        provider_openai.go
        provider_local.go
      /asr
        transcriber.go
        provider_whisper.go
      /email
        imap_client.go
        parser.go
    /infra
      db.go
      redis.go
      queue.go
      logger.go
      config.go
      metrics.go
      tracing.go
  /migrations
    001_init.sql
    002_add_reminders.sql
    003_add_finance.sql
    ...
  /webapp
    (frontend для Mini App)
  README.md
  instructions.md
```

(Остальная часть арх-дока опущена здесь ради краткости — она у тебя в чате и при необходимости можно расширить файл.)

