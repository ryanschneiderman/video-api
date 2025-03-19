#!/bin/bash
set -e

# Configuration
REGION="us-east-1"
ACCOUNT_ID="498061775412"

# ECR repository URIs
API_REPO="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/twelve-labs/api"
WORKER_REPO="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/twelve-labs/video-processor"

# Build the API container
echo "Building API container..."
docker build -f cmd/api/Dockerfile -t twelve-labs-api .

# Tag the API container for ECR
echo "Tagging API container..."
docker tag twelve-labs-api:latest ${API_REPO}:latest

# Build the Worker container
echo "Building Worker container..."
docker build -f cmd/video_processor/Dockerfile -t twelve-labs-video-processor .

# Tag the Worker container for ECR
echo "Tagging Worker container..."
docker tag twelve-labs-video-processor:latest ${WORKER_REPO}:latest

# Authenticate Docker with ECR
echo "Authenticating to AWS ECR..."
aws ecr get-login-password --region ${REGION} | docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com

# Push the API container to ECR
echo "Pushing API container to ECR..."
docker push ${API_REPO}:latest

# Push the Worker container to ECR
echo "Pushing Worker container to ECR..."
docker push ${WORKER_REPO}:latest

echo "Deployment completed successfully!"
