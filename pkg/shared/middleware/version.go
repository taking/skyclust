package middleware

import (
	"net/http"
	"strings"

	"cmp/pkg/shared/errors"
	"cmp/pkg/shared/logger"

	"github.com/gin-gonic/gin"
)

// APIVersion represents an API version
type APIVersion struct {
	Version     string
	Deprecated  bool
	SunsetDate  string
	Description string
}

// VersionManager manages API versions
type VersionManager struct {
	versions       map[string]APIVersion
	defaultVersion string
}

// NewVersionManager creates a new version manager
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions: map[string]APIVersion{
			"v1": {
				Version:     "v1",
				Deprecated:  false,
				SunsetDate:  "",
				Description: "Current stable version",
			},
			"v2": {
				Version:     "v2",
				Deprecated:  false,
				SunsetDate:  "",
				Description: "Next generation API",
			},
		},
		defaultVersion: "v1",
	}
}

// APIVersionMiddleware provides API version management
func APIVersionMiddleware(versionManager *VersionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract version from URL path
		path := c.Request.URL.Path
		version := extractVersionFromPath(path)

		// If no version in path, check header
		if version == "" {
			version = c.GetHeader("API-Version")
		}

		// If still no version, use default
		if version == "" {
			version = versionManager.defaultVersion
		}

		// Validate version
		if !versionManager.IsValidVersion(version) {
			apiErr := errors.NewAPIError(
				errors.ErrCodeValidationFailed,
				"Unsupported API version",
				http.StatusBadRequest,
			)
			_ = apiErr.WithDetails("supported_versions", versionManager.GetSupportedVersions())
			_ = apiErr.WithDetails("requested_version", version)
			_ = c.Error(apiErr)
			return
		}

		// Check if version is deprecated
		if versionManager.IsDeprecated(version) {
			apiInfo := versionManager.GetVersion(version)
			c.Header("API-Version", version)
			c.Header("API-Deprecated", "true")
			if apiInfo.SunsetDate != "" {
				c.Header("API-Sunset", apiInfo.SunsetDate)
			}

			logger.Warnf("Deprecated API version used: %s", version)
		}

		// Set version in context
		c.Set("api_version", version)
		c.Header("API-Version", version)

		c.Next()
	}
}

// extractVersionFromPath extracts API version from URL path
func extractVersionFromPath(path string) string {
	// Look for /api/v1/, /api/v2/, etc.
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "api" && i+1 < len(parts) {
			version := parts[i+1]
			if strings.HasPrefix(version, "v") {
				return version
			}
		}
	}
	return ""
}

// IsValidVersion checks if a version is valid
func (vm *VersionManager) IsValidVersion(version string) bool {
	_, exists := vm.versions[version]
	return exists
}

// IsDeprecated checks if a version is deprecated
func (vm *VersionManager) IsDeprecated(version string) bool {
	if apiVersion, exists := vm.versions[version]; exists {
		return apiVersion.Deprecated
	}
	return false
}

// GetVersion returns version information
func (vm *VersionManager) GetVersion(version string) APIVersion {
	if apiVersion, exists := vm.versions[version]; exists {
		return apiVersion
	}
	return APIVersion{}
}

// GetSupportedVersions returns list of supported versions
func (vm *VersionManager) GetSupportedVersions() []string {
	versions := make([]string, 0, len(vm.versions))
	for version := range vm.versions {
		versions = append(versions, version)
	}
	return versions
}

// GetDefaultVersion returns the default version
func (vm *VersionManager) GetDefaultVersion() string {
	return vm.defaultVersion
}

// SetDefaultVersion sets the default version
func (vm *VersionManager) SetDefaultVersion(version string) {
	vm.defaultVersion = version
}

// AddVersion adds a new version
func (vm *VersionManager) AddVersion(version APIVersion) {
	vm.versions[version.Version] = version
}

// DeprecateVersion marks a version as deprecated
func (vm *VersionManager) DeprecateVersion(version, sunsetDate string) {
	if apiVersion, exists := vm.versions[version]; exists {
		apiVersion.Deprecated = true
		apiVersion.SunsetDate = sunsetDate
		vm.versions[version] = apiVersion
	}
}

// VersionHeaderMiddleware adds version headers to responses
func VersionHeaderMiddleware(versionManager *VersionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Add version information to response headers
		c.Header("API-Version", c.GetString("api_version"))
		c.Header("API-Supported-Versions", strings.Join(versionManager.GetSupportedVersions(), ", "))

		// Add deprecation warnings if applicable
		version := c.GetString("api_version")
		if versionManager.IsDeprecated(version) {
			apiInfo := versionManager.GetVersion(version)
			c.Header("API-Deprecated", "true")
			if apiInfo.SunsetDate != "" {
				c.Header("API-Sunset", apiInfo.SunsetDate)
			}
		}
	}
}

// VersionInfoHandler provides version information
func VersionInfoHandler(versionManager *VersionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		versions := make(map[string]APIVersion)
		for version, info := range versionManager.versions {
			versions[version] = info
		}

		c.JSON(http.StatusOK, gin.H{
			"current_version":    c.GetString("api_version"),
			"default_version":    versionManager.GetDefaultVersion(),
			"supported_versions": versionManager.GetSupportedVersions(),
			"versions":           versions,
		})
	}
}
