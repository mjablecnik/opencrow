#!/bin/sh
#
# Setup Fly.io Secrets for OpenCrow
#
# This script only sets up secrets without deploying
#

set -e

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🔐 OpenCrow Secrets Setup"
echo "========================="
echo ""

# Check if flyctl is installed
if ! command -v flyctl >/dev/null 2>&1; then
    echo "❌ Error: flyctl is not installed"
    echo "Install it from: https://fly.io/docs/hands-on/install-flyctl/"
    exit 1
fi

# Check if .env file exists
if [ ! -f "$PROJECT_DIR/.env" ]; then
    echo "❌ Error: .env file not found in $PROJECT_DIR"
    echo "Please create a .env file with your configuration"
    echo "You can copy .env.example and fill in your values"
    exit 1
fi

# Load environment variables from .env
echo "📋 Loading environment variables from .env..."

# Read .env file and export variables (handles Windows line endings)
while IFS='=' read -r key value || [ -n "$key" ]; do
    # Skip comments and empty lines
    case "$key" in
        \#*|'') continue ;;
    esac
    
    # Remove leading/trailing whitespace and carriage returns
    key=$(printf '%s' "$key" | tr -d '\r' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    value=$(printf '%s' "$value" | tr -d '\r' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    # Export the variable
    if [ -n "$key" ] && [ -n "$value" ]; then
        export "$key=$value"
    fi
done < "$PROJECT_DIR/.env" 2>/dev/null

# Check required secrets
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "❌ Error: TELEGRAM_BOT_TOKEN is not set in .env"
    exit 1
fi

if [ -z "$OPENROUTER_API_KEY" ]; then
    echo "❌ Error: OPENROUTER_API_KEY is not set in .env"
    exit 1
fi

echo "✅ Required secrets found"
echo ""

# Set secrets in Fly.io
echo "🔐 Setting secrets in Fly.io..."
flyctl secrets set TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" OPENROUTER_API_KEY="$OPENROUTER_API_KEY" --app opencrow

echo ""
echo "✅ Secrets configured successfully!"
echo ""
echo "📊 View configured secrets:"
echo "  flyctl secrets list --app opencrow"
echo ""
