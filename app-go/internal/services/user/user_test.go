package user_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/services/user"
	testutils "github.com/TheTempest7/social_network/app-go/internal/test_utils"

	_ "modernc.org/sqlite"
)

func TestRegisterUserRaceCondition(t *testing.T) {
	repo := testutils.SetupDB(t) // 👈 Вся настройка БД в одной строке
	svc := user.NewUserService(repo)

	const goroutines = 100
	var wg sync.WaitGroup
	var (
		successes int
		conflicts int
		others    int
		mu        sync.Mutex
	)

	input := &user.RegisterInput{
		Email:    "race@test.com",
		Username: "race",
		Password: "12345678",
	}

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.RegisterUser(context.Background(), input)

			mu.Lock()
			defer mu.Unlock()

			switch {
			case err == nil:
				successes++
			case errors.Is(err, repository.ErrEmailTaken),
				errors.Is(err, repository.ErrUsernameTaken):
				conflicts++
			default:
				others++
				t.Logf("неожиданная ошибка: %v", err)
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Errorf("ожидали 1 успех, получили %d", successes)
	}
	if conflicts != goroutines-1 {
		t.Errorf("ожидали %d конфликтов, получили %d", goroutines-1, conflicts)
	}
	if others != 0 {
		t.Errorf("ожидали 0 прочих ошибок, получили %d", others)
	}
}
func TestLoginUserSuccess(t *testing.T) {
	repo := testutils.SetupDB(t)
	svc := user.NewUserService(repo)
	// Сначала регистрируем
	regReq := &user.RegisterInput{
		Email:    "login@test.com",
		Username: "loginuser",
		Password: "12345678",
	}
	_, err := svc.RegisterUser(context.Background(), regReq)
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	// Теперь логинимся
	user, token, err := svc.LoginUser(context.Background(), user.LoginInput{
		Username: "loginuser",
		Password: "12345678",
	})
	if err != nil {
		t.Fatalf("логин: %v", err)
	}

	if user == nil {
		t.Fatal("user == nil")
	}
	if user.Username != "loginuser" {
		t.Errorf("ожидали 'loginuser', получили '%s'", user.Username)
	}
	if token == "" {
		t.Error("токен пустой")
	}
}

func TestLoginUserWrongPassword(t *testing.T) {
	repo := testutils.SetupDB(t)
	svc := user.NewUserService(repo)

	_, err := svc.RegisterUser(context.Background(), &user.RegisterInput{
		Email:    "wrong@test.com",
		Username: "wrongpass",
		Password: "correct123",
	})
	if err != nil {
		t.Fatalf("регистрация: %v", err)
	}

	_, _, err = svc.LoginUser(context.Background(), user.LoginInput{
		Username: "wrongpass",
		Password: "wrong123",
	})

	if !errors.Is(err, user.ErrInvalidCreds) {
		t.Errorf("ожидали ErrInvalidCreds, получили %v", err)
	}
}
