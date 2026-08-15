#!/bin/bash

echo "========================================="
echo "  Generating Swagger Documentation"
echo "========================================="

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Swag is not installed. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Clean existing docs
echo "Cleaning existing docs..."
rm -rf docs/

# Generate docs - This will create docs/docs.go, docs/swagger.json, docs/swagger.yaml
echo "Generating Swagger docs..."
swag init \
    --dir . \
    --generalInfo cmd/api/main.go \
    --output docs \
    --parseDependency \
    --parseInternal \
    --parseDepth 5 \
    --exclude vendor

# Check if generated successfully
if [ -f "docs/swagger.json" ] && [ -f "docs/docs.go" ]; then
    echo "✓ Swagger docs generated successfully!"
    
    # Count operations
    if command -v python3 &> /dev/null; then
        OPERATIONS=$(python3 -c "
import json
with open('docs/swagger.json', 'r') as f:
    data = json.load(f)
    paths = data.get('paths', {})
    count = sum(len(methods) for methods in paths.values())
    print(count)
" 2>/dev/null)
        echo "✓ Operations found: $OPERATIONS"
    fi
    
    echo ""
    echo "Files generated:"
    echo "  - docs/docs.go"
    echo "  - docs/swagger.json"
    echo "  - docs/swagger.yaml"
    echo ""
    echo "View Swagger UI at: http://localhost:8080/swagger/index.html"
else
    echo "✗ Failed to generate Swagger docs"
    echo "Keeping existing docs if any..."
    exit 1
fi