package services

import (
	"errors"
	"fmt"

	"github.com/TheTempest7/social_network/app-go/internal/auth"
	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/models"
	"github.com/TheTempest7/social_network/app-go/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// Ошибки сервиса
var (
	ErrInvalidCreds    = errors.New("неверный логин или пароль")
	ErrPasswordTooLong = errors.New("пароль слишком длинный (макс 72 байта)")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// RegisterUser — регистрация нового пользователя
func (s *UserService) RegisterUser(req dto.RegisterRequest) (*dto.UserResponse, error) {

	// 2. Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return nil, ErrPasswordTooLong // 👈 своя ошибка
		}
		return nil, fmt.Errorf("хеширование пароля: %w", err)
	}

	// 3. Создаём модель
	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	// 4. Сохраняем в БД
	created, err := s.repo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	// 5. Возвращаем DTO (без пароля)
	return toUserResponse(created), nil
}

// LoginUser — вход
func (s *UserService) LoginUser(req dto.LoginRequest) (*models.User, string, error) {
	// 1. Ищем пользователя
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCreds
		}
		return nil, "", fmt.Errorf("поиск пользователя: %w", err)
	}

	// 2. Проверяем пароль
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, "", ErrInvalidCreds
	}

	// 3. Проверяем, что активен
	if !user.IsActive {
		return nil, "", errors.New("пользователь неактивен")
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("генерация токена: %w", err)
	}

	return user, token, nil
}

func (s *UserService) GetByID(id int) (*dto.UserResponse, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

// toUserResponse — конвертер модели в DTO
func toUserResponse(u *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		IsActive:  u.IsActive,
	}
}
