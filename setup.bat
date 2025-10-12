@echo off
REM Step-CA Web UI Setup Script for Windows

echo 🚀 Setting up Step-CA Web UI...
echo 📦 Repository: https://github.com/marcin-kruszynski/step-ui.git

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker is not installed. Please install Docker first.
    exit /b 1
)

docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker Compose is not available. Please ensure Docker Desktop is running.
    exit /b 1
)

REM Create necessary directories
echo 📁 Creating directories...
if not exist data mkdir data

REM Copy environment file if it doesn't exist
if not exist .env (
    echo 📝 Creating .env file...
    copy example.env .env
    echo ✅ Created .env file. Please edit it with your Step-CA configuration.
) else (
    echo ✅ .env file already exists.
)

REM Check environment configuration
echo 🔐 Checking configuration...
if "%CA_URL%"=="" (
    echo ⚠️  CA_URL not set in .env file. Please configure your Step-CA URL.
)

if "%PROVISIONER_PASSWORD%"=="" (
    echo ⚠️  PROVISIONER_PASSWORD not set in .env file. Please configure your provisioner password.
)

REM Build and start services
echo 🔨 Building and starting services...
docker compose up -d --build

echo.
echo 🎉 Setup complete!
echo.
echo 📋 Next steps:
echo 1. Edit .env file with your Step-CA configuration
echo 2. Restart services: docker compose restart
echo.
echo 🌐 Access the application:
echo    Frontend: http://localhost:3000
echo    Backend API: http://localhost:8080
echo.
echo 📖 For more information, see README.md

pause
