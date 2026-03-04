#!/bin/sh
#
# Fly.io Deployment Script for OpenCrow
#
# This script sets up secrets and deploys the OpenCrow bot to Fly.io
#

set -e

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🚀 OpenCrow Fly.io Deployment Script"
echo "======================================"
echo ""

# Deploy the application
echo "📦 Deploying application..."
cd "$PROJECT_DIR"
flyctl deploy --app opencrow

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Useful commands:"
echo "  flyctl status --app opencrow          # Check app status"
echo "  flyctl logs --app opencrow            # View logs"
echo "  flyctl ssh console --app opencrow     # SSH into container"
echo "  flyctl secrets list --app opencrow    # List configured secrets"
echo ""
