package docs

import (
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
)

// SwaggerUIHandler serves the Swagger UI
func SwaggerUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		swaggerHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>CMP API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '/api/v1/docs/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                onComplete: function() {
                    console.log('Swagger UI loaded');
                }
            });
        };
    </script>
</body>
</html>`

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, swaggerHTML)
	}
}

// SwaggerJSONHandler serves the OpenAPI JSON specification
func SwaggerJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		spec := GenerateOpenAPISpec()
		jsonData, err := spec.ToJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate OpenAPI specification",
			})
			return
		}

		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, string(jsonData))
	}
}

// ReDocHandler serves the ReDoc documentation
func ReDocHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		redocHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>CMP API Documentation - ReDoc</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <redoc spec-url='/api/v1/docs/swagger.json'></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@2.0.0/bundles/redoc.standalone.js"></script>
</body>
</html>`

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, redocHTML)
	}
}

// APIVersionHandler serves API version information
func APIVersionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"api_version": "v1",
			"version":     "1.0.0",
			"build_time":  "2024-01-01T00:00:00Z",
			"git_commit":  "unknown",
			"go_version":  "1.21",
		})
	}
}

// HealthCheckHandler provides detailed health check information
func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, you would check various services
		health := gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
			"version":   "1.0.0",
			"services": gin.H{
				"database": gin.H{
					"status":  "healthy",
					"latency": "1ms",
				},
				"cache": gin.H{
					"status":  "healthy",
					"latency": "0.5ms",
				},
				"nats": gin.H{
					"status":  "healthy",
					"latency": "0.2ms",
				},
			},
		}

		c.JSON(http.StatusOK, health)
	}
}

// MetricsHandler serves basic metrics (simplified)
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := gin.H{
			"requests_total":      1000,
			"requests_per_second": 10.5,
			"response_time_avg":   "50ms",
			"error_rate":          0.01,
			"active_connections":  25,
			"memory_usage":        "128MB",
			"cpu_usage":           "15%",
		}

		c.JSON(http.StatusOK, metrics)
	}
}

// OpenAPITemplateHandler serves a custom OpenAPI template
func OpenAPITemplateHandler() gin.HandlerFunc {
	tmpl := `
# CMP API Documentation

## Overview
The Cloud Management Platform (CMP) API provides a unified interface for managing multiple cloud providers.

## Base URL
{{.BaseURL}}

## Authentication
All API requests require authentication using JWT tokens.

### Getting a Token
POST /api/v1/auth/login

## Rate Limiting
API requests are rate limited to prevent abuse:
- 60 requests per minute per IP
- 1000 requests per hour per user
- 10000 requests per day per API key

## Error Handling
All errors follow a consistent format:

` + "```" + `json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {},
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456"
  }
}
` + "```" + `

## Common Error Codes
- ` + "`UNAUTHORIZED`" + `: Authentication required
- ` + "`FORBIDDEN`" + `: Access denied
- ` + "`NOT_FOUND`" + `: Resource not found
- ` + "`VALIDATION_FAILED`" + `: Invalid input
- ` + "`RATE_LIMIT_EXCEEDED`" + `: Too many requests
- ` + "`INTERNAL_ERROR`" + `: Server error

## SDKs and Examples
- [Go SDK](https://github.com/cmp/go-sdk)
- [Python SDK](https://github.com/cmp/python-sdk)
- [JavaScript SDK](https://github.com/cmp/js-sdk)

## Support
- Documentation: https://docs.cmp.local
- Support: support@cmp.local
- Issues: https://github.com/cmp/issues
`

	return func(c *gin.Context) {
		t, err := template.New("openapi").Parse(tmpl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate documentation template",
			})
			return
		}

		data := map[string]interface{}{
			"BaseURL": "http://localhost:8080",
		}

		c.Header("Content-Type", "text/plain")
		if err := t.Execute(c.Writer, data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to execute template",
			})
		}
	}
}
