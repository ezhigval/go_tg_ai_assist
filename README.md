# 🧠 Jarvis Assistant — AI-ассистент на Go
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-6+-DC382D?logo=redis)](https://redis.io/)
[![Telegram Bot](https://img.shields.io/badge/Telegram-Bot_API-26A5E4?logo=telegram)](https://core.telegram.org/bots/api)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](#)

Персональный ассистент уровня JARVIS: Телеграм-бот + Мини-эпп + AI + календарь + финансы + задачи + кредиты + почта.  
Создан как портфолио-проект уровня **Senior Go Backend Engineer**.

---

## 🌟 Возможности

### 🧠 AI
- Понимание естественного языка
- Голосовые команды → Whisper → Intent
- Персональная манера общения

### 🗓 Календарь + напоминания
- Создание событий и напоминаний
- ICS-интеграция с iCloud
- Автогенерация напоминаний (кредиты, финансы, задачи)

### ✔️ Задачи
- todo‑листы, статусы, дедлайны
- Привязка к событиям календаря

### 💰 Финансы
- Учёт доходов и расходов
- Баланс по счетам
- Регулярные операции

### 🏦 Кредиты
- Создание кредитов
- Аннуитетный график
- Связка с напоминаниями и финансами

### 📬 Почта
- IMAP‑подключение
- AI‑фильтрация писем
- Извлечение OTP‑кодов

### 🌐 Telegram Mini App
- React + TS + Vite
- Tailwind + Framer Motion
- Real‑time через WebSocket

---

## 🏗 Архитектура

Модульный монолит:

```
/cmd
  /bot         — Telegram Webhook сервер
  /api         — backend Mini App
  /worker      — AI/ASR/почта/события
  /scheduler   — периодические задачи
  /migrate     — миграции БД
/internal
  /domain      — бизнес‑логика (чистый Go)
  /adapters    — Telegram, AI, Email, ICS, WebAPI
  /infra       — PostgreSQL, Redis/NATS, логирование, конфиги, метрики
/webapp        — Telegram Mini App (React)
```

---

## 🧰 Используемый стэк

### Backend
- Go 1.22+
- PostgreSQL + pgxpool  
- Redis Streams / NATS  
- Chi Router  
- OpenTelemetry  
- Prometheus  
- golang‑migrate  
- Custom Logger  

### AI
- Whisper ASR  
- OpenAI / LM Studio  

### Integrations
- Telegram Bot API  
- Telegram WebApp SDK  
- Email IMAP  
- ICS Calendar  

### Frontend
- React  
- TypeScript  
- Vite  
- Zustand  
- Tailwind  
- Framer Motion  

---

## 🛠 Установка и запуск

### 1. Скачать зависимости
```bash
go mod tidy
```

### 2. Поднять PostgreSQL
```bash
docker compose up -d
```

### 3. Применить миграции
```bash
go run ./cmd/migrate
```

### 4. Запуск подсистем
```bash
go run ./cmd/bot
go run ./cmd/api
go run ./cmd/worker
go run ./cmd/scheduler
```

---

## 📚 Документация

- **01_architecture_and_stack.md** — архитектура и стэк  
- **02_technical_spec.md** — техническое задание  
- **03_roadmap.md** — план разработки  
- **04_senior_go_skills_map.md** — карта навыков Go  
- **instructions.md** — обучающий мануал по проекту  

---

## 🤝 Автор проекта  
**Валентин Ежов**  
Telegram: **@ezhigval**

---

## ⭐ Если вы HR/Team Lead

Данный проект демонстрирует навыки уровня **Senior Go Backend Engineer**:  
архитектура, асинхронность, очереди, интеграции, Event‑Driven, AI, тестирование, WebApp.

