package profile // 👈 было package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"

	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/models"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/storage"
)

type UpdateProfileInput struct {
	FirstName *string
	LastName  *string
	Username  *string
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

	if input.Username != nil {
		user.Username = *input.Username
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
	repo    repository.UserRepo // используем тот же интерфейс, что и в UserService
	storage *storage.Storage
}

func NewProfileService(repo repository.UserRepo, storage *storage.Storage) *ProfileService {
	return &ProfileService{repo: repo, storage: storage}
}

func (s *ProfileService) GetByID(ctx context.Context, id int) (*dto.UserOutput, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserOutput(user), nil
}

func (s *ProfileService) UploadAvatar(
	ctx context.Context,
	objectName string,
	file multipart.File,
	size int64,
) (string, error) {
	return s.storage.UploadFile(ctx, file, objectName, size)
}

func (s *ProfileService) SetAvatar(
	ctx context.Context,
	id int,
	url string,
) (*dto.UserOutput, error) {

	updated, err := s.repo.UpdateAvatar(ctx, id, url)
	if prettyJSON, err := json.MarshalIndent(updated, "", "  "); err == nil {
		fmt.Printf("Updated User:\n%s\n", string(prettyJSON))
	}
	if err != nil {
		return nil, err
	}
	return toUserOutput(updated), nil
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
		AvatarURL: u.AvatarURL,
	}
}
