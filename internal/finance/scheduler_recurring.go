package finance

import (
	"context"
	"fmt"
	"time"

	"tg_bot_asist/internal/logger"
	"tg_bot_asist/internal/todo"
)

type RecurringScheduler struct {
	repo       *RecurringRepo
	todoSvc    *todo.Service
	financeSvc *Service
}

func NewRecurringScheduler(repo *RecurringRepo, todoSvc *todo.Service, financeSvc *Service) *RecurringScheduler {
	return &RecurringScheduler{
		repo:       repo,
		todoSvc:    todoSvc,
		financeSvc: financeSvc,
	}
}

// RunDailyCheck проверяет все регулярные платежи и выполняет необходимые действия.
// Вызывается каждый час из main.go.
func (s *RecurringScheduler) RunDailyCheck(ctx context.Context) {
	logger.Debug("RecurringScheduler: starting daily check")

	// Получаем все регулярные платежи (для всех пользователей)
	// userID = 0 означает "все пользователи" (для scheduler)
	allPayments, err := s.repo.GetUserPayments(ctx, 0)
	if err != nil {
		logger.Error("Failed to get recurring payments: " + err.Error())
		return
	}

	now := time.Now().Truncate(24 * time.Hour) // Округляем до начала дня
	processed := 0

	for _, payment := range allPayments {
		paymentDate := payment.NextPayment.Truncate(24 * time.Hour)

		// Проверяем, наступила ли дата платежа
		if paymentDate.Before(now) || paymentDate.Equal(now) {
			logger.Info(
				fmt.Sprintf("Processing recurring payment: %s for user %d", payment.Title, payment.UserID),
			)

			// 1. Создаём финансовую запись
			if err := s.financeSvc.AddRecurringExecution(ctx, payment); err != nil {
				logger.Error("Failed to create finance entry: " + err.Error())
				continue
			}

			// 2. Обновляем дату следующего платежа
			nextPayment := calcNextRecurringDate(payment)
			if err := s.repo.UpdateNextPayment(ctx, payment.ID, nextPayment); err != nil {
				logger.Error("Failed to update next payment date: " + err.Error())
				continue
			}

			// 3. Создаём напоминание в TODO (опционально, можно настроить)
			reminderTitle := fmt.Sprintf("Платёж выполнен: %s — %.2f ₽", payment.Title, payment.Amount)
			if err := s.todoSvc.CreateAuto(ctx, payment.UserID, reminderTitle); err != nil {
				logger.Warn("Failed to create todo reminder: " + err.Error())
				// Не критично, продолжаем
			}

			processed++
		}
	}

	if processed > 0 {
		logger.Info(fmt.Sprintf("RecurringScheduler: processed %d payments", processed))
	} else {
		logger.Debug("RecurringScheduler: no payments due today")
	}
}

func calcNextRecurringDate(p *RecurringPayment) time.Time {
	switch p.Period {
	case "daily":
		return p.NextPayment.AddDate(0, 0, 1)
	case "weekly":
		return p.NextPayment.AddDate(0, 0, 7)
	case "monthly":
		return p.NextPayment.AddDate(0, 1, 0)
	case "yearly":
		return p.NextPayment.AddDate(1, 0, 0)
	}
	return p.NextPayment.AddDate(0, 1, 0)
}
