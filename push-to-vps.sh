#!/bin/bash
# ============================================
# MAUTRADE — Local to VPS Deployer
# Bypasses GitHub Actions entirely.
# Runs rsync to copy code to VPS, then executes deploy script.
# ============================================

set -euo pipefail

# Konfigurasi VPS Lo
VPS_HOST="109.123.233.56"
VPS_USER="root"
# Kalo pake private key spesifik, masukin path-nya di sini, misal: ~/.ssh/id_rsa
SSH_KEY_OPT="" 

echo -e "\033[0;36m[1/3] Preparing to deploy to $VPS_USER@$VPS_HOST...\033[0m"

# Pastikan directory tujuan ada di VPS
ssh $SSH_KEY_OPT -o StrictHostKeyChecking=no $VPS_USER@$VPS_HOST "mkdir -p ~/Mautrade"

echo -e "\033[0;33m[2/3] Syncing files to VPS (excluding envs & heavy folders)...\033[0m"
rsync -avz --delete \
  --exclude '.git' \
  --exclude '.env' \
  --exclude 'backend/.env' \
  --exclude 'frontend/.env' \
  --exclude 'logs' \
  --exclude 'letsencrypt' \
  --exclude '.deploy.lock' \
  --exclude '**/node_modules' \
  --exclude '**/.nuxt' \
  --exclude '**/.output' \
  --exclude '**/.nitro' \
  --exclude '**/dist' \
  --exclude '**/build' \
  --exclude '**/.cache' \
  --exclude '**/target' \
  --exclude '**/__pycache__' \
  --exclude '**/*.log' \
  -e "ssh $SSH_KEY_OPT -o StrictHostKeyChecking=no" \
  ./ $VPS_USER@$VPS_HOST:~/Mautrade/

echo -e "\033[0;36m[3/3] Running deployment script on VPS...\033[0m"
ssh $SSH_KEY_OPT -o StrictHostKeyChecking=no $VPS_USER@$VPS_HOST "cd ~/Mautrade && chmod +x deploy-vps.sh && ./deploy-vps.sh"

echo -e "\033[0;32m==============================================\033[0m"
echo -e "\033[0;32m  ✓ Deployment Triggered Successfully!\033[0m"
echo -e "\033[0;32m==============================================\033[0m"
