import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    # База данных
    DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./social_app.db")

    # JWT
    SECRET_KEY = os.getenv("SECRET_KEY", "your-secret-key-here-change-in-production")
    ALGORITHM = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES = 30


settings = Settings()