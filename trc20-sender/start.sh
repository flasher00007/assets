#!/bin/bash

echo "🚀 Starting TRC20 USDT Sender..."
echo ""

# Check if binary exists
if [ ! -f "./trc20-sender" ]; then
    echo "📦 Building application..."
    go build -o trc20-sender main.go
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
    echo "✅ Build successful!"
    echo ""
fi

echo "🌐 Starting server on http://localhost:8090"
echo "📝 Press Ctrl+C to stop the server"
echo ""

./trc20-sender
