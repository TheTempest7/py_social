from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from app.core.config import settings

# Подключение к БД (SQLite)
engine = create_engine(
    settings.DATABASE_URL,
    connect_args={"check_same_thread": False}  # Нужно для SQLite
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()

# Функция для получения сессии БД (будем использовать в роутерах)
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()