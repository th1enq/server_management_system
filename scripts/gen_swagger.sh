#!/bin/bash

echo "Generating Swagger documentation..."

# Generate swagger docs
swag init -g cmd/server/main.go -o docs

if [ $? -eq 0 ]; then
    echo "Swagger documentation generated successfully!"
    echo ""
    echo "Generated files:"
    echo "- docs/docs.go"
    echo "- docs/swagger.json"
    echo "- docs/swagger.yaml"
    echo ""
    echo "You can access the Swagger UI at: http://localhost:8080/swagger/index.html"
    echo "Make sure your server is running first with: go run cmd/server/main.go"
else
    echo "Failed to generate Swagger documentation"
    exit 1
fi
