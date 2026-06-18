#!/bin/bash
set -e

echo "DB_ENCRYPTION_KEY=$(openssl rand -base64 32)"
