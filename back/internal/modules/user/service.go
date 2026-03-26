package user

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Delete(ctx context.Context, userID string) error {
	return s.repo.DeleteByID(ctx, userID)
}

