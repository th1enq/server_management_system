#!/bin/bash

# Test script để kiểm tra flow mới với API Gateway

echo "=== Testing new API Gateway flow ==="

# Build project
echo "Building project..."
go build -o bin/server cmd/server/main.go

if [ $? -ne 0 ]; then
    echo "Build failed! Please check wire generation"
    exit 1
fi

echo "Build successful!"

# Generate wire
echo "Generating wire dependencies..."
cd internal/wiring
go generate
cd ../..

if [ $? -ne 0 ]; then
    echo "Wire generation failed! Please check dependencies"
    exit 1
fi

echo "Wire generation successful!"

echo "=== Flow Summary ==="
echo "1. Server gửi metrics → API Gateway (ProcessServerMetrics)"
echo "2. API Gateway kiểm tra cache và database status"
echo "3. Nếu status thay đổi → gửi StatusChangeMessage qua Kafka → Elasticsearch"
echo "4. Luôn gửi MonitoringMessage qua Kafka → Server UseCase → update status thành ON"
echo ""
echo "=== Files created/modified ==="
echo "- /internal/infrastructure/mq/producer/status_change.go (NEW)"
echo "- /internal/infrastructure/services/api_gateway_service.go (NEW)"
echo "- /internal/delivery/consumers/status_change.go (NEW)"
echo "- /internal/delivery/http/controllers/server_controller.go (MODIFIED)"
echo "- /internal/usecases/server_usecase.go (MODIFIED)"
echo "- Various wireset files (MODIFIED)"
