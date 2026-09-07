from datetime import datetime, timedelta, timezone
from typing import Optional
from jose import JWTError, jwt
from passlib.context import CryptContext
from app.core.config import settings

# Настройка хеширования паролей
pwd_context = CryptContext(schemes=["sha256_crypt"], deprecated="auto")


# Хеширование пароля
def get_password_hash(password: str) -> str:
    # 👇 ЛОГИРУЙ ВСЁ!
    print(f"🔍 [DEBUG] Тип password: {type(password)}")
    print(f"🔍 [DEBUG] Длина password: {len(password)}")
    print(f"🔍 [DEBUG] Содержимое password (первые 20 символов): {password[:20]}")
    return pwd_context.hash(password)


# Проверка пароля
def verify_password(plain_password: str, hashed_password: str) -> bool:
    return pwd_context.verify(plain_password, hashed_password)


# Создание JWT токена
def create_access_token(data: dict, expires_delta: Optional[timedelta] = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.now(timezone.utc) + expires_delta
    else:
        expire =datetime.now(timezone.utc) + timedelta(minutes=15)

    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
    return encoded_jwt


# Декодирование JWT токена
def decode_token(token: str):
    try:
        payload = jwt.decode(token, settings.SECRET_KEY, algorithms=[settings.ALGORITHM])
        return payload
    except JWTError:
        return None