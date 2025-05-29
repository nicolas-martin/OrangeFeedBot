#!/bin/bash

# OrangeFeed Setup Script
# This script helps you set up the OrangeFeed Truth Social Market Intelligence Bot

set -e

echo "🎯 OrangeFeed Setup Script"
echo "=========================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    echo "   Visit: https://golang.org/dl/"
    exit 1
fi

echo "✅ Go is installed: $(go version)"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp config.env.example .env
    echo "✅ .env file created"
else
    echo "⚠️  .env file already exists"
fi

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod tidy
go mod download
echo "✅ Dependencies installed"

# Build the application
echo "🔨 Building OrangeFeed..."
make build
echo "✅ Build complete"

echo ""
echo "🎉 Setup Complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your credentials:"
echo "   - TRUTHSOCIAL_USERNAME and TRUTHSOCIAL_PASSWORD"
echo "   - OPENAI_API_KEY"
echo "   - TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID"
echo ""
echo "2. Test the system:"
echo "   make test"
echo ""
echo "3. Run the application:"
echo "   make run"
echo ""
echo "📚 For detailed instructions, see README.md"
echo ""

# Check if required environment variables are set
echo "🔍 Checking environment configuration..."

if [ -f ".env" ]; then
    source .env
    
    missing_vars=()
    
    if [ -z "$TRUTHSOCIAL_USERNAME" ] || [ "$TRUTHSOCIAL_USERNAME" = "your_username" ]; then
        missing_vars+=("TRUTHSOCIAL_USERNAME")
    fi
    
    if [ -z "$TRUTHSOCIAL_PASSWORD" ] || [ "$TRUTHSOCIAL_PASSWORD" = "your_password" ]; then
        missing_vars+=("TRUTHSOCIAL_PASSWORD")
    fi
    
    if [ -z "$OPENAI_API_KEY" ] || [ "$OPENAI_API_KEY" = "your_openai_api_key" ]; then
        missing_vars+=("OPENAI_API_KEY")
    fi
    
    if [ -z "$TELEGRAM_BOT_TOKEN" ] || [ "$TELEGRAM_BOT_TOKEN" = "your_telegram_bot_token" ]; then
        missing_vars+=("TELEGRAM_BOT_TOKEN")
    fi
    
    if [ -z "$TELEGRAM_CHAT_ID" ] || [ "$TELEGRAM_CHAT_ID" = "your_chat_id" ]; then
        missing_vars+=("TELEGRAM_CHAT_ID")
    fi
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        echo "⚠️  Please configure these variables in .env:"
        for var in "${missing_vars[@]}"; do
            echo "   - $var"
        done
        echo ""
        echo "📝 Edit .env file: nano .env"
    else
        echo "✅ All required environment variables are configured"
        echo ""
        echo "🚀 Ready to run! Execute: make run"
    fi
fi

echo ""
echo "🔗 Useful commands:"
echo "   make help     - Show all available commands"
echo "   make test     - Test the Truth Social connection"
echo "   make run      - Start the monitoring bot"
echo "   make clean    - Clean build artifacts"
echo "" 