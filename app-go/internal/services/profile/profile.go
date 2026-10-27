package profile // 👈 было package services

import (
	"context"

	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/models"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
)

type UpdateProfileInput struct {
	FirstName *string
	LastName  *string
}

func (s *ProfileService) UpdateProfile(ctx context.Context, id int, input UpdateProfileInput) (*dto.UserOutput, error) {
	// 1. Достаем текущего пользователя из БД
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Меняем только те поля, которые пришли (не nil)
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}

	// 3. Сохраняем в БД
	updated, err := s.repo.UpdateProfile(ctx, user)
	if err != nil {
		return nil, err
	}

	// 4. Возвращаем DTO
	return toUserOutput(updated), nil
}

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
