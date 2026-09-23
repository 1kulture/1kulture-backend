#!/bin/bash

echo "========================================="
echo "  Generating Swagger Documentation"
echo "========================================="

export PATH=$PATH:$(go env GOPATH)/bin

if ! command -v swag &> /dev/null; then
    echo "Swag not found. Installing latest..."
    go install github.com/swaggo/swag/cmd/swag@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

echo "Cleaning existing docs..."
rm -rf docs/

echo "Generating Swagger docs..."
# NOTE: --parseDependency is intentionally OMITTED.
# It crawls Go stdlib (math/rand/v2) which uses generics the parser doesn't support.
swag init \
    -g cmd/api/main.go \
    -o docs \
    --parseInternal \
    --parseDepth 5

if [ -f "docs/swagger.json" ]; then
    COUNT=$(python3 -c "
import json
with open('docs/swagger.json') as f:
    d = json.load(f)
    print(len(d.get('paths', {})))
" 2>/dev/null || echo "?")
    echo "✅ Swagger docs generated — paths: $COUNT"
    echo "   View at http://localhost:8080/swagger/index.html"
else
    echo "⚠️  swagger.json not generated. Check the log above."
    exit 1
fi