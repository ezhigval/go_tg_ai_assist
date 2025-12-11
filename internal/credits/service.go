package credits

import (
	"context"
	"fmt"
)

// Service предоставляет бизнес-логику для работы с кредитами.
type Service struct {
	repo Repository
}

// NewService создаёт новый экземпляр сервиса кредитов.
func NewService(r Repository) *Service {
	return &Service{repo: r}
}

// Add создаёт новый кредит и возвращает его ID.
func (s *Service) Add(ctx context.Context, c *Credit) (int, error) {
	return s.repo.AddCredit(ctx, c)
}

// List возвращает список всех кредитов пользователя.
func (s *Service) List(ctx context.Context, userID int64) ([]Credit, error) {
	return s.repo.ListCredits(ctx, userID)
}

// GetByID возвращает кредит по ID, если он принадлежит пользователю.
func (s *Service) GetByID(ctx context.Context, id int, userID int64) (*Credit, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// Delete удаляет кредит по ID, если он принадлежит пользователю.
func (s *Service) Delete(ctx context.Context, id int, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

// Copy создаёт копию существующего кредита с пометкой "(копия)" в названии.
func (s *Service) Copy(ctx context.Context, id int, userID int64) (int, error) {
	original, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return 0, fmt.Errorf("кредит не найден: %w", err)
	}

	copy := &Credit{
		UserID:    userID,
		Title:     original.Title + " (копия)",
		Principal: original.Principal,
		Rate:      original.Rate,
		Months:    original.Months,
		CreatedAt: original.CreatedAt,
	}

	return s.repo.AddCredit(ctx, copy)
}
