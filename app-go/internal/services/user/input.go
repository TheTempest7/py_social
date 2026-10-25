package user // 👈 было package services

// RegisterInput — то, что принимает сервис для регистрации.
// Это не DTO из HTTP-слоя, это бизнес-сущность.
type RegisterInput struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
}

// LoginInput — то, что принимает сервис для логина.
type LoginInput struct {
	Username string
	Password string
}

// UserOutput — то, что сервис отдает наружу (в хендлер).
// Без привязки к HTTP/DTO.
type UserOutput struct {
	ID        int
	Email     string
	Username  string
	FirstName string
	LastName  string
	IsActive  bool
}
