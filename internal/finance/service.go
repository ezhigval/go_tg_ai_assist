package finance

import (
	"context"
	"time"
)

// repository определяет интерфейс для работы с финансовыми записями.
type repository interface {
	AddEntry(context.Context, *FinanceEntry) error
	ListEntries(context.Context, int64) ([]*FinanceEntry, error)
}

// Service предоставляет бизнес-логику для работы с финансами и регулярными платежами.
type Service struct {
	repo          repository
	recurringRepo *RecurringRepo
}

// NewService создаёт новый экземпляр сервиса финансов.
func NewService(repo repository, recurring *RecurringRepo) *Service {
	return &Service{repo: repo, recurringRepo: recurring}
}

// AddEntry добавляет новую финансовую запись (доход или расход).
func (s *Service) AddEntry(ctx context.Context, e *FinanceEntry) error {
	return s.repo.AddEntry(ctx, e)
}

// ListEntries возвращает список всех финансовых записей пользователя.
func (s *Service) ListEntries(ctx context.Context, userID int64) ([]*FinanceEntry, error) {
	return s.repo.ListEntries(ctx, userID)
}

// AddRecurring добавляет новый регулярный платёж.
func (s *Service) AddRecurring(ctx context.Context, p *RecurringPayment) (int, error) {
	return s.recurringRepo.Add(ctx, p)
}

// DeleteRecurring удаляет регулярный платёж по ID.
func (s *Service) DeleteRecurring(ctx context.Context, id int) error {
	return s.recurringRepo.Delete(ctx, id)
}

// GetRecurringList возвращает список всех регулярных платежей пользователя.
func (s *Service) GetRecurringList(ctx context.Context, userID int64) ([]*RecurringPayment, error) {
	list, err := s.recurringRepo.GetUserPayments(ctx, userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// AddRecurringExecution создаёт финансовую запись из выполненного регулярного платежа.
func (s *Service) AddRecurringExecution(ctx context.Context, p *RecurringPayment) error {
	entry := &FinanceEntry{
		UserID:    p.UserID,
		Amount:    p.Amount,
		Category:  p.Category,
		Type:      "expense",
		Note:      "Регулярный платёж: " + p.Title,
		CreatedAt: time.Now(),
	}
	return s.repo.AddEntry(ctx, entry)
}
