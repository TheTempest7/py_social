package profile // 👈 было package services

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"strings"
	"time"

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
	if err != nil {
		return nil, err
	}
	return toUserOutput(updated), nil
}

// GenerateAvatarUploadLink — просто проксирует вызов в storage
func (s *ProfileService) GenerateAvatarUploadLink(ctx context.Context, objectName string) (*storage.PresignedPostData, error) {
	// Ссылка живёт 5 минут
	return s.storage.GeneratePresignedPost(ctx, objectName, 5*time.Minute)
}

// ConfirmAvatarUpload — проверяет файл в MinIO и сохраняет в БД
func (s *ProfileService) ConfirmAvatarUpload(ctx context.Context, userID int, objectName string) (*dto.UserOutput, error) {
	// 1. Проверяем, что файл реально существует и его размер/тип ок
	info, err := s.storage.StatObject(ctx, objectName)
	if err != nil {
		slog.Error("file not found or invalid: %w", err)
		return nil, fmt.Errorf("file not found or invalid: %w", err)
	}

	// Проверка размера (дублируем, на всякий случай)
	if info.Size > 5*1024*1024 {
		_ = s.storage.DeleteObject(ctx, objectName) // удаляем мусор
		return nil, fmt.Errorf("file too large")
	}

	// Проверка типа (MinIO вернёт Content-Type, который был указан при загрузке)
	if !strings.HasPrefix(info.ContentType, "image/") {
		_ = s.storage.DeleteObject(ctx, objectName)
		slog.Error("invalid content type - failed", "err", err)
		return nil, fmt.Errorf("invalid content type")
	}

	// 2. Формируем публичный URL (как ты делал раньше)
	// Внимание: для продакшена лучше генерировать PresignedGet, но пока оставим прямой
	url := fmt.Sprintf("http://%s/%s/%s", s.storage.GetEndpoint(), s.storage.GetBucket(), objectName)

	// 3. Сохраняем в БД через репозиторий (тот самый UpdateAvatar)
	updated, err := s.repo.UpdateAvatar(ctx, userID, url)
	if err != nil {
		slog.Error("UpdateAvatar failed", "err", err)
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
