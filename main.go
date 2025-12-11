package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"tg_bot_asist/cmd"
	"tg_bot_asist/internal/api"
	"tg_bot_asist/internal/api/auth"
	"tg_bot_asist/internal/api/websocket"
	"tg_bot_asist/internal/bot"
	"tg_bot_asist/internal/config"
	"tg_bot_asist/internal/credits"
	"tg_bot_asist/internal/finance"
	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/storage"
	"tg_bot_asist/internal/todo"
)

func main() {
	// Инициализация логгера
	logger.Init(logger.LEVEL_INFO, "logs/bot.log")
	logger.Info("=========== Starting telegram assistant ===========")

	// Загружаем переменные окружения
	if err := config.LoadEnv(); err != nil {
		logger.Warn("Failed to load .env file, using system environment variables")
	}

	// Инициализация Telegram бота
	state, err := cmd.StartBotBooting()
	if err != nil {
		logger.Fatal("Bot boot error: " + err.Error())
		return
	}

	// Подключение к базе данных
	db, err := config.ConnectDB()
	if err != nil {
		logger.Fatal("DB connection error: " + err.Error())
	}
	defer db.Close()

	// Выполнение миграций
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := storage.RunMigrations(ctx, db); err != nil {
		logger.Fatal("Migration error: " + err.Error())
		return
	}

	// Инициализация репозиториев
	userRepo := storage.NewUserRepo(db)
	todoRepo := storage.NewTodoRepo(db)
	stateRepo := storage.NewStateRepo(db)
	creditsRepo := storage.NewCreditsRepo(db)
	financeRepo := storage.NewFinanceRepo(db)
	recurringRepo := finance.NewRecurringRepo(db)

	// Инициализация сервисов
	todoService := todo.NewService(todoRepo)
	creditService := credits.NewService(creditsRepo)
	financeService := finance.NewService(financeRepo, recurringRepo)

	// Запуск планировщика регулярных платежей
	scheduler := finance.NewRecurringScheduler(recurringRepo, todoService, financeService)
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	schedulerWg := &sync.WaitGroup{}
	schedulerWg.Add(1)
	go func() {
		defer schedulerWg.Done()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		// Первый запуск сразу
		scheduler.RunDailyCheck(schedulerCtx)

		for {
			select {
			case <-schedulerCtx.Done():
				logger.Info("Scheduler stopped")
				return
			case <-ticker.C:
				scheduler.RunDailyCheck(schedulerCtx)
			}
		}
	}()

	// Инициализация JWT
	auth.InitJWT()

	// Создание WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Создание API роутера
	apiRouter := api.NewRouter(
		userRepo,
		todoService,
		financeService,
		creditService,
		wsHub,
	)

	// Запуск HTTP сервера для API
	apiPort := config.GetWithDefault("API_PORT", "8080")
	apiHandler := apiRouter.SetupRoutes()

	httpServer := &http.Server{
		Addr:    ":" + apiPort,
		Handler: apiHandler,
	}

	// Запуск HTTP сервера в отдельной горутине
	httpServerWg := &sync.WaitGroup{}
	httpServerWg.Add(1)
	go func() {
		defer httpServerWg.Done()
		logger.Info("Starting API server on port " + apiPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("API server error: " + err.Error())
		}
	}()

	// Настройка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	logger.Info("Bot is now waiting for updates...")
	logger.Info("Press Ctrl+C to gracefully shutdown the bot")

	// Запуск обработки обновлений в отдельной горутине
	botWg := &sync.WaitGroup{}
	botWg.Add(1)
	botCtx, botCancel := context.WithCancel(context.Background())
	go func() {
		defer botWg.Done()
		bot.HandleUpdatesWithContext(
			botCtx,
			state.Bot,
			state.Updates,
			userRepo,
			todoService,
			stateRepo,
			creditService,
			financeService,
			recurringRepo,
		)
	}()

	// Ожидание сигнала завершения
	<-sigChan
	logger.Info("Received shutdown signal, gracefully stopping all services...")

	// 1. Останавливаем планировщик
	logger.Info("Stopping scheduler...")
	schedulerCancel()
	schedulerWg.Wait()
	logger.Info("Scheduler stopped")

	// 2. Останавливаем HTTP сервер
	logger.Info("Stopping HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error: " + err.Error())
	} else {
		logger.Info("HTTP server stopped")
	}

	// 3. Останавливаем WebSocket Hub
	logger.Info("Stopping WebSocket hub...")
	wsHub.Stop()
	logger.Info("WebSocket hub stopped")

	// 4. Останавливаем обработку обновлений бота
	logger.Info("Stopping bot updates handler...")
	botCancel()
	botWg.Wait()
	logger.Info("Bot updates handler stopped")

	// 5. Закрываем соединение с БД
	logger.Info("Closing database connection...")
	db.Close()
	logger.Info("Database connection closed")

	logger.Info("All services stopped gracefully. Goodbye!")
}
