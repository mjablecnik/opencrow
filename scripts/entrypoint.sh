#!/bin/sh
#
# Entrypoint script for OpenCrow bot
# Initializes default agent identity files if they don't exist
#

set -e

AGENT_DIR="/app/workplace/agent"
DEFAULT_AGENT_DIR="/app/default-agent"

echo "🚀 Starting OpenCrow bot..."

# Check if agent directory exists, if not create it
if [ ! -d "$AGENT_DIR" ]; then
    echo "📁 Creating agent directory..."
    mkdir -p "$AGENT_DIR"
fi

# Copy default identity files if they don't exist
echo "🔍 Checking for identity files..."
for file in IDENTITY.md PERSONALITY.md SOUL.md USER.md TOOLS.md; do
    if [ ! -f "$AGENT_DIR/$file" ]; then
        if [ -f "$DEFAULT_AGENT_DIR/$file" ]; then
            echo "📄 Creating default $file..."
            cp "$DEFAULT_AGENT_DIR/$file" "$AGENT_DIR/$file"
        else
            echo "⚠️  Warning: Default $file not found in image"
        fi
    else
        echo "✓ $file already exists"
    fi
done

echo "✅ Identity files ready"
echo ""

# Start the bot
exec /app/bot
