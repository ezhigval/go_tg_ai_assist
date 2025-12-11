package todo

import (
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Add(ctx context.Context, userID int64, title, desc string, due *time.Time) error {
	item := &Item{
		UserID:      userID,
		Title:       title,
		Description: desc,
		DueDate:     due,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	_, err := s.repo.Create(ctx, item)
	return err
}

func (s *Service) List(ctx context.Context, userID int64) ([]Item, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, userID int64, id int) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *Service) CreateAuto(ctx context.Context, userID int64, title string) error {
	return s.Add(ctx, userID, title, "(автоматическое напоминание)", nil)
}
