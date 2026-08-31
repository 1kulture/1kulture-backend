#!/bin/bash

echo "========================================="
echo "  Generating Swagger Documentation"
echo "========================================="

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Swag not found. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Clean existing docs (but we'll keep a backup)
rm -rf docs/

# Try automatic generation
echo "Attempting automatic generation..."
swag init \
    --dir . \
    --generalInfo cmd/api/main.go \
    --output docs \
    --parseDependency \
    --parseInternal \
    --parseDepth 5 \
    --exclude vendor \
    --ot go,json,yaml

# Check if generation succeeded
if [ -f "docs/docs.go" ] && [ -f "docs/swagger.json" ]; then
    echo "✅ Automatic generation succeeded!"
    echo "Files created:"
    echo "  - docs/docs.go"
    echo "  - docs/swagger.json"
    echo "  - docs/swagger.yaml"
else
    echo "⚠️  Automatic generation failed. Creating minimal docs.go..."
    mkdir -p docs
    cat > docs/docs.go <<'EOF'
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "openapi": "3.0.0",
    "info": {
        "title": "1Kulture API",
        "version": "1.0",
        "description": "Enterprise Event Management System API"
    },
    "host": "localhost:8080",
    "basePath": "/api/v1",
    "schemes": ["http", "https"],
    "paths": {}
}`

var SwaggerInfo = &swag.Spec{
    Version:          "1.0",
    Host:             "localhost:8080",
    BasePath:         "/api/v1",
    Schemes:          []string{"http", "https"},
    Title:            "1Kulture API",
    Description:      "Enterprise Event Management System API",
    InfoInstanceName: "swagger",
    SwaggerTemplate:  docTemplate,
}

func init() {
    swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
EOF
    echo "✅ Created minimal docs/docs.go"
    echo "ℹ️  You can later replace it with full generated docs if automatic generation works."
fi