from pydantic import BaseModel, EmailStr, Field
from typing import Optional

# Схема для регистрации
class UserCreate(BaseModel):
    email: EmailStr
    username: str = Field(min_length=3, max_length=50)
    password: str = Field(min_length=8, max_length=72)
    first_name: Optional[str] = None
    last_name: Optional[str] = None

# Схема для логина
class UserLogin(BaseModel):
    username: str
    password: str

# Схема для ответа (что возвращаем клиенту)
class UserResponse(BaseModel):
    id: int
    email: str
    username: str
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    is_active: bool

    class Config:
        from_attributes = True  # Для работы с SQLAlchemy моделями

# Схема для JWT токена
class Token(BaseModel):
    access_token: str
    token_type: str = "bearer"