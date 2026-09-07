from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.core.database import engine, Base
from app.routers import auth

# Создаём таблицы
Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="My Social App",
    description="Social network API",
    version="0.1.0"
)

# CORS для React
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 👇 ПОДКЛЮЧАЕМ РОУТЕР
app.include_router(auth.router, prefix="/api/auth", tags=["auth"])

# 👇 ПОДКЛЮЧАЕМ РОУТЕР
@app.get("/")
async def root():
    return {"message": "Hello, World!"}

@app.get("/ping")
def ping():
    return {"status": "pong"}