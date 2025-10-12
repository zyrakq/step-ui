#!/bin/bash

# Step-CA Web UI Setup Script

set -e

echo "🚀 Setting up Step-CA Web UI..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p data
mkdir -p secrets

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "✅ Created .env file. Please edit it with your Step-CA configuration."
else
    echo "✅ .env file already exists."
fi

# Check if secrets directory has required files
echo "🔐 Checking secrets..."
if [ ! -f secrets/root_ca.crt ]; then
    echo "⚠️  secrets/root_ca.crt not found. Please add your root CA certificate."
fi

if [ ! -f secrets/provisioner_pass.txt ]; then
    echo "⚠️  secrets/provisioner_pass.txt not found. Please add your provisioner password."
fi

# Build and start services
echo "🔨 Building and starting services..."
docker-compose up -d --build

echo ""
echo "🎉 Setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Edit .env file with your Step-CA configuration"
echo "2. Add your root CA certificate to secrets/root_ca.crt"
echo "3. Add your provisioner password to secrets/provisioner_pass.txt"
echo "4. Restart services: docker-compose restart"
echo ""
echo "🌐 Access the application:"
echo "   Frontend: http://localhost:3000"
echo "   Backend API: http://localhost:8080"
echo ""
echo "📖 For more information, see README.md"
