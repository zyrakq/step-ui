@echo off
REM Step-CA Web UI Setup Script for Windows

echo 🚀 Setting up Step-CA Web UI...

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker is not installed. Please install Docker first.
    exit /b 1
)

docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker Compose is not installed. Please install Docker Compose first.
    exit /b 1
)

REM Create necessary directories
echo 📁 Creating directories...
if not exist data mkdir data
if not exist secrets mkdir secrets

REM Copy environment file if it doesn't exist
if not exist .env (
    echo 📝 Creating .env file...
    copy .env.example .env
    echo ✅ Created .env file. Please edit it with your Step-CA configuration.
) else (
    echo ✅ .env file already exists.
)

REM Check if secrets directory has required files
echo 🔐 Checking secrets...
if not exist secrets\root_ca.crt (
    echo ⚠️  secrets\root_ca.crt not found. Please add your root CA certificate.
)

if not exist secrets\provisioner_pass.txt (
    echo ⚠️  secrets\provisioner_pass.txt not found. Please add your provisioner password.
)

REM Build and start services
echo 🔨 Building and starting services...
docker-compose up -d --build

echo.
echo 🎉 Setup complete!
echo.
echo 📋 Next steps:
echo 1. Edit .env file with your Step-CA configuration
echo 2. Add your root CA certificate to secrets\root_ca.crt
echo 3. Add your provisioner password to secrets\provisioner_pass.txt
echo 4. Restart services: docker-compose restart
echo.
echo 🌐 Access the application:
echo    Frontend: http://localhost:3000
echo    Backend API: http://localhost:8080
echo.
echo 📖 For more information, see README.md

pause
