#!/usr/bin/env bash
set -euo pipefail

# Deploy cb to Fly.io
# Prerequisites:
#   1. Install flyctl: curl -L https://fly.io/install.sh | sh
#   2. Login: fly auth login

echo "=== Deploying cb to Fly.io ==="

# Check flyctl
if ! command -v fly &>/dev/null; then
    echo "Error: flyctl not found. Install with: curl -L https://fly.io/install.sh | sh"
    exit 1
fi

# Check if app exists
if ! fly status &>/dev/null; then
    echo "Creating new Fly.io app..."
    fly launch --no-deploy --copy-config --yes
fi

# Set secrets
echo ""
echo "Setting JWT secret..."
JWT_SECRET=$(openssl rand -hex 32)
fly secrets set CB_JWT_SECRET="$JWT_SECRET" --yes

echo ""
echo "Deploying..."
fly deploy --yes

echo ""
echo "=== Deployment complete ==="
echo "Your app URL: https://$(fly status --json 2>/dev/null | python3 -c 'import sys,json;print(json.load(sys.stdin)["App"]["Name"])' 2>/dev/null || echo 'cb').fly.dev"
echo ""
echo "Next steps:"
echo "  1. Open https://your-app.fly.dev in browser"
echo "  2. Register the first user (becomes admin)"
echo "  3. CLI: cb login --api-url https://your-app.fly.dev/v1 --user yourname"
