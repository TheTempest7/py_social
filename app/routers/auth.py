from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from sqlalchemy.orm import Session
from datetime import timedelta

from fastapi.responses import JSONResponse

from app.core.database import get_db
from app.schemas.user import UserCreate, UserResponse, Token
from app.services.user_service import UserService
from app.core.security import create_access_token
from app.core.config import settings

router = APIRouter()
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/auth/login")


# 📝 Регистрация
@router.post("/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
async def register(user_data: UserCreate, db: Session = Depends(get_db)):
    service = UserService(db)
    try:
        user = service.create_user(user_data)
        return user
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


# 🔑 Логин (получение токена)
@router.post("/login", response_model=Token)
async def login(form_data: OAuth2PasswordRequestForm = Depends(), db: Session = Depends(get_db)):
    service = UserService(db)
    user = service.authenticate_user(form_data.username, form_data.password)
    print("🔍 [1] Вход в /login")
    if not user:
        return JSONResponse(
            status_code=status.HTTP_401_UNAUTHORIZED,
            content={
                "is_authenticated": False,
                "detail": "Incorrect username or password"
            },
            headers={"WWW-Authenticate": "Bearer"},
        )
    print("🔍 [2] Вход в /login")
    access_token_expires = timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)
    access_token = create_access_token(
        data={"sub": user.username}, expires_delta=access_token_expires
    )
    print("🔍 [3] Вход в /login")
    print("👤 [USER DATA] ==========")
    print(f"  🆔 ID: {user.id}")
    print(f"  📛 Username: {user.username}")
    print(f"  📧 Email: {user.email}")
    print(f"  👤 First name: {user.first_name}")
    print(f"  👤 Last name: {user.last_name}")
    print(f"  ✅ Active: {user.is_active}")
    print(f"  📅 Created: {user.created_at}")
    print(f"  📅 Updated: {user.updated_at}")
    print("👤 [END USER DATA] ======")
    response = JSONResponse({
        "message": "Login successful",
        "is_authenticated": True,
        "user": {
            "id": user.id,
            "username": user.username,
            "email": user.email,
            "first_name": user.first_name,
            "last_name": user.last_name,
            "is_active": user.is_active,
            "created_at": user.created_at.isoformat() if user.created_at else None,  # 👈
            "updated_at": user.updated_at.isoformat() if user.updated_at else None  # 👈
        }
    })
    print("🔍 [4] Вход в /login")
    response.set_cookie(
        key="access_token",
        value=access_token,
        httponly=True,
        secure=False,
        samesite="lax",
        max_age=settings.ACCESS_TOKEN_EXPIRE_MINUTES * 60
    )

    return response