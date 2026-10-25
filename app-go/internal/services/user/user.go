package user // 👈 было package services
import (
	"context"
	"errors"
	"fmt"

	"github.com/TheTempest7/social_network/app-go/internal/auth"
	"github.com/TheTempest7/social_network/app-go/internal/dto"
	"github.com/TheTempest7/social_network/app-go/internal/models"
	"github.com/TheTempest7/social_network/app-go/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo repository.UserRepo
}

// Ошибки сервиса
var (
	ErrInvalidCreds    = errors.New("неверный логин или пароль")
	ErrPasswordTooLong = errors.New("пароль слишком длинный (макс 72 байта)")
	ErrUserInactive    = errors.New("пользователь неактивен")
)

func NewUserService(repo repository.UserRepo) *UserService {
	return &UserService{repo: repo}
}

// RegisterUser — регистрация нового пользователя
func (s *UserService) RegisterUser(ctx context.Context, input *RegisterInput) (*dto.UserOutput, error) {

	// 2. Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return nil, ErrPasswordTooLong // 👈 своя ошибка
		}
		return nil, fmt.Errorf("хеширование пароля: %w", err)
	}

	// 3. Создаём модель
	user := &models.User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hash),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		IsActive:     true,
	}

	// 4. Сохраняем в БД
	created, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	// 5. Возвращаем DTO (без пароля)
	return &dto.UserOutput{
		ID:        created.ID,
		Email:     created.Email,
		Username:  created.Username,
		FirstName: created.FirstName,
		LastName:  created.LastName,
		IsActive:  created.IsActive,
	}, nil
}

// LoginUser — вход
func (s *UserService) LoginUser(ctx context.Context, input LoginInput) (*models.User, string, error) {
	// 1. Ищем пользователя
	user, err := s.repo.GetUserByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			bcrypt.CompareHashAndPassword(
				[]byte("$2a$10$dummy.hash.to.waste.time.for.timing.attack"),
				[]byte(input.Password),
			)

			return nil, "", ErrInvalidCreds
		}
		return nil, "", fmt.Errorf("поиск пользователя: %w", err)
	}

	// 2. Проверяем пароль
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		return nil, "", ErrInvalidCreds
	}

	// 3. Проверяем, что активен
	if !user.IsActive {
		return nil, "", ErrUserInactive
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("генерация токена: %w", err)
	}

	return user, token, nil
}
