#!/bin/bash

# Simple HTTP server script for testing the WASM font access demo
# Usage: ./serve.sh [port]

PORT=${1:-8080}

echo "Starting HTTP server on port $PORT..."
echo "Open http://localhost:$PORT/example.html in Chrome to test the Local Font Access API"
echo ""
echo "Make sure you have:"
echo "1. Built the WASM file: GOOS=js GOARCH=wasm go build -o example.wasm example_font_access.go"
echo "2. Copied wasm_exec.js from your Go installation"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Use Python's built-in HTTP server
if command -v python3 &> /dev/null; then
    python3 -m http.server $PORT
elif command -v python &> /dev/null; then
    python -m SimpleHTTPServer $PORT
else
    echo "Error: Python not found. Please install Python or use another HTTP server."
    exit 1
fi