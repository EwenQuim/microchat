#!/bin/bash

set -e

echo "Building MicroChat..."

# Build frontend
echo "Building frontend..."
cd app
npm run build
cd ..

# Copy frontend to static
echo "Copying frontend build to static..."
rm -rf static/*
cp -r app/dist/* static/

# Build server
echo "Building server..."
mkdir -p bin
go build -o bin/server ./cmd/server

# Build CLI
echo "Building CLI..."
go build -o bin/cli ./cmd/cli

echo "Build complete!"
echo "Run './bin/server' to start the server"
echo "Run './bin/cli -cmd rooms' to use the CLI"
