#!/bin/bash

IMAGE_NAME="mansoor1/plausible-public-proxy"
# Get the latest Git commit SHA
GIT_SHA=$(git rev-parse --short HEAD)

# Build the Docker image
docker build -t $IMAGE_NAME:$GIT_SHA -t $IMAGE_NAME:latest .
docker push $IMAGE_NAME:$GIT_SHA
docker push $IMAGE_NAME:latest

echo "Docker images pushed: $IMAGE_NAME:$GIT_SHA and $IMAGE_NAME:latest"
