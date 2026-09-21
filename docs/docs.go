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
