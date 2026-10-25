package profile // 👈 было package services

import (
	"context"

	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/models"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
)

type ProfileService struct {
	repo repository.UserRepo // используем тот же интерфейс, что и в UserService
}

func NewProfileService(repo repository.UserRepo) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetByID(ctx context.Context, id int) (*dto.UserOutput, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserOutput(user), nil
}

// toUserOutput — конвертер из модели БД в бизнес-структуру
func toUserOutput(u *models.User) *dto.UserOutput {
	return &dto.UserOutput{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		IsActive:  u.IsActive,
	}
}
