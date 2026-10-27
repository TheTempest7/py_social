package dto

// RegisterRequest — что приходит на /register
type RegisterRequest struct {
	Email     string `json:"email"     validate:"required,email"`
	Username  string `json:"username"  validate:"required,min=3,max=50"`
	Password  string `json:"password"  validate:"required,min=8,max=72"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest — что приходит на /login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UserResponse — что отдаём клиенту
type UserResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
}

// AuthResponse — что возвращаем после логина
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserOutput struct {
	ID        int
	Email     string
	Username  string
	FirstName string
	LastName  string
	IsActive  bool
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}
