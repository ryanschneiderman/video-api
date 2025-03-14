#!/bin/bash
# create_project_structure.sh
# This script creates the project folder structure for the video upload & storage demo.

set -e

echo "Creating directories..."

mkdir -p terraform
mkdir -p cmd/api
mkdir -p cmd/cli
mkdir -p internal/config
mkdir -p internal/handlers
mkdir -p internal/models
mkdir -p internal/storage
mkdir -p internal/db
mkdir -p internal/middleware
mkdir -p pkg
mkdir -p scripts

echo "Creating files..."

# Terraform files
touch terraform/main.tf
touch terraform/variables.tf

# API and CLI entry points
touch cmd/api/main.go
touch cmd/cli/main.go

# Internal configuration
touch internal/config/config.go

# Internal handlers
touch internal/handlers/video.go

# Data models
touch internal/models/video.go

# Storage implementations
touch internal/storage/s3_storage.go
touch internal/storage/local_storage.go

# Database access
touch internal/db/db.go

# Middleware (e.g., logging)
touch internal/middleware/logging.go

# Project root files
touch Dockerfile
touch go.mod
touch go.sum

echo "Project structure created successfully."
