package docs

import (
	"encoding/json"
	"strings"
)

// OpenAPISpec represents the OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Servers    []Server            `json:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components,omitempty"`
	Tags       []Tag               `json:"tags,omitempty"`
}

// Info provides metadata about the API
type Info struct {
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Version        string  `json:"version"`
	TermsOfService string  `json:"termsOfService,omitempty"`
	Contact        Contact `json:"contact,omitempty"`
	License        License `json:"license,omitempty"`
}

// Contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server information
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable for server URL templating
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem represents a path in the API
type PathItem struct {
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Callbacks   map[string]Callback   `json:"callbacks,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	Servers     []Server              `json:"servers,omitempty"`
}

// Parameter represents a parameter
type Parameter struct {
	Name            string             `json:"name"`
	In              string             `json:"in"`
	Description     string             `json:"description,omitempty"`
	Required        bool               `json:"required,omitempty"`
	Deprecated      bool               `json:"deprecated,omitempty"`
	AllowEmptyValue bool               `json:"allowEmptyValue,omitempty"`
	Style           string             `json:"style,omitempty"`
	Explode         bool               `json:"explode,omitempty"`
	AllowReserved   bool               `json:"allowReserved,omitempty"`
	Schema          *Schema            `json:"schema,omitempty"`
	Example         interface{}        `json:"example,omitempty"`
	Examples        map[string]Example `json:"examples,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

// Response represents a response
type Response struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Links       map[string]Link      `json:"links,omitempty"`
}

// Header represents a header
type Header struct {
	Description     string             `json:"description,omitempty"`
	Required        bool               `json:"required,omitempty"`
	Deprecated      bool               `json:"deprecated,omitempty"`
	AllowEmptyValue bool               `json:"allowEmptyValue,omitempty"`
	Style           string             `json:"style,omitempty"`
	Explode         bool               `json:"explode,omitempty"`
	AllowReserved   bool               `json:"allowReserved,omitempty"`
	Schema          *Schema            `json:"schema,omitempty"`
	Example         interface{}        `json:"example,omitempty"`
	Examples        map[string]Example `json:"examples,omitempty"`
}

// MediaType represents a media type
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty"`
}

// Schema represents a JSON Schema
type Schema struct {
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Example              interface{}        `json:"example,omitempty"`
	Examples             []interface{}      `json:"examples,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	Const                interface{}        `json:"const,omitempty"`
	MultipleOf           *float64           `json:"multipleOf,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	ExclusiveMaximum     *float64           `json:"exclusiveMaximum,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	ExclusiveMinimum     *float64           `json:"exclusiveMinimum,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty"`
	MinItems             *int               `json:"minItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty"`
	MaxProperties        *int               `json:"maxProperties,omitempty"`
	MinProperties        *int               `json:"minProperties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	PatternProperties    map[string]*Schema `json:"patternProperties,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AllOf                []*Schema          `json:"allOf,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty"`
	AnyOf                []*Schema          `json:"anyOf,omitempty"`
	Not                  *Schema            `json:"not,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
}

// Example represents an example
type Example struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

// Encoding represents encoding information
type Encoding struct {
	ContentType   string            `json:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty"`
	Style         string            `json:"style,omitempty"`
	Explode       bool              `json:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty"`
}

// Link represents a link
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty"`
	OperationId  string                 `json:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Server       *Server                `json:"server,omitempty"`
}

// Callback represents a callback
type Callback map[string]PathItem

// SecurityRequirement represents a security requirement
type SecurityRequirement map[string][]string

// Components represents reusable components
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty"`
	Examples        map[string]*Example        `json:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"`
	Headers         map[string]*Header         `json:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
	Links           map[string]*Link           `json:"links,omitempty"`
	Callbacks       map[string]*Callback       `json:"callbacks,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type             string      `json:"type"`
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIdConnectUrl string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents OAuth flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth flow
type OAuthFlow struct {
	AuthorizationUrl string            `json:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty"`
	RefreshUrl       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// Tag represents a tag
type Tag struct {
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	ExternalDocs *ExternalDoc `json:"externalDocs,omitempty"`
}

// ExternalDoc represents external documentation
type ExternalDoc struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// GenerateOpenAPISpec generates the OpenAPI specification for CMP
func GenerateOpenAPISpec() *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       "Cloud Management Platform API",
			Description: "A plugin-based cloud management portal that supports multiple cloud providers through a unified interface.",
			Version:     "1.0.0",
			Contact: Contact{
				Name:  "CMP Team",
				Email: "support@cmp.local",
			},
			License: License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
		Servers: []Server{
			{
				URL:         "http://localhost:8080",
				Description: "Development server",
			},
		},
		Tags: []Tag{
			{
				Name:        "Authentication",
				Description: "User authentication and authorization",
			},
			{
				Name:        "Workspaces",
				Description: "Workspace management",
			},
			{
				Name:        "Cloud Providers",
				Description: "Cloud provider management",
			},
			{
				Name:        "Virtual Machines",
				Description: "VM lifecycle management",
			},
			{
				Name:        "Credentials",
				Description: "Cloud provider credentials management",
			},
			{
				Name:        "Infrastructure as Code",
				Description: "OpenTofu execution management",
			},
			{
				Name:        "Kubernetes",
				Description: "Kubernetes cluster management",
			},
		},
		Paths:      generatePaths(),
		Components: generateComponents(),
	}
}

// generatePaths generates the API paths
func generatePaths() map[string]PathItem {
	return map[string]PathItem{
		"/health": {
			Get: &Operation{
				Summary:     "Health Check",
				Description: "Check the health status of the API",
				Tags:        []string{"System"},
				Responses: map[string]Response{
					"200": {
						Description: "Service is healthy",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"status": {
											Type:        "string",
											Description: "Health status",
											Example:     "healthy",
										},
										"version": {
											Type:        "string",
											Description: "API version",
											Example:     "1.0.0",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"/api/v1/auth/login": {
			Post: &Operation{
				Summary:     "User Login",
				Description: "Authenticate a user and return a JWT token",
				Tags:        []string{"Authentication"},
				RequestBody: &RequestBody{
					Description: "Login credentials",
					Required:    true,
					Content: map[string]MediaType{
						"application/json": {
							Schema: &Schema{
								Type: "object",
								Properties: map[string]*Schema{
									"email": {
										Type:        "string",
										Format:      "email",
										Description: "User email",
									},
									"password": {
										Type:        "string",
										Description: "User password",
									},
								},
								Required: []string{"email", "password"},
							},
						},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "Login successful",
						Content: map[string]MediaType{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"token": {
											Type:        "string",
											Description: "JWT token",
										},
										"user": {
											Ref: "#/components/schemas/User",
										},
									},
								},
							},
						},
					},
					"401": {
						Description: "Invalid credentials",
					},
				},
			},
		},
		// Add more paths as needed
	}
}

// generateComponents generates the reusable components
func generateComponents() Components {
	return Components{
		Schemas: map[string]*Schema{
			"User": {
				Type: "object",
				Properties: map[string]*Schema{
					"id": {
						Type:        "string",
						Format:      "uuid",
						Description: "User ID",
					},
					"email": {
						Type:        "string",
						Format:      "email",
						Description: "User email",
					},
					"name": {
						Type:        "string",
						Description: "User name",
					},
					"created_at": {
						Type:        "string",
						Format:      "date-time",
						Description: "Creation timestamp",
					},
				},
			},
			"Error": {
				Type: "object",
				Properties: map[string]*Schema{
					"code": {
						Type:        "string",
						Description: "Error code",
					},
					"message": {
						Type:        "string",
						Description: "Error message",
					},
					"details": {
						Type:        "object",
						Description: "Additional error details",
					},
					"timestamp": {
						Type:        "string",
						Format:      "date-time",
						Description: "Error timestamp",
					},
					"request_id": {
						Type:        "string",
						Description: "Request ID for tracking",
					},
				},
			},
		},
		SecuritySchemes: map[string]*SecurityScheme{
			"BearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "JWT token authentication",
			},
		},
	}
}

// ToJSON converts the OpenAPI spec to JSON
func (spec *OpenAPISpec) ToJSON() ([]byte, error) {
	return json.MarshalIndent(spec, "", "  ")
}

// ToYAML converts the OpenAPI spec to YAML (simplified)
func (spec *OpenAPISpec) ToYAML() (string, error) {
	// This is a simplified YAML conversion
	// In a real implementation, you would use a proper YAML library
	jsonData, err := spec.ToJSON()
	if err != nil {
		return "", err
	}

	// Convert JSON to YAML-like format (simplified)
	yaml := strings.ReplaceAll(string(jsonData), "\"", "")
	yaml = strings.ReplaceAll(yaml, ":", ": ")
	yaml = strings.ReplaceAll(yaml, ",", "")
	yaml = strings.ReplaceAll(yaml, "{", "")
	yaml = strings.ReplaceAll(yaml, "}", "")

	return yaml, nil
}
