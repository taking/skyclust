package common

import "time"

// API version constants
// API 버전 상수
const (
	APIVersionV1 = "v1"
	APIVersionV2 = "v2" // Future version
)

// APIVersion represents API version information
type APIVersion struct {
	Version        string
	Deprecated     bool
	DeprecationDate *time.Time
	SunsetDate     *time.Time
}

// SupportedVersions contains information about all supported API versions
var SupportedVersions = map[string]APIVersion{
	"v1": {
		Version:    APIVersionV1,
		Deprecated: false,
	},
	"v2": {
		Version:    APIVersionV2,
		Deprecated: false,
	},
}

// CurrentAPIVersion returns the current API version
// 현재 API 버전 반환
func CurrentAPIVersion() string {
	return APIVersionV1
}

// GetAPIVersionFromPath extracts API version from path
// 경로에서 API 버전 추출
func GetAPIVersionFromPath(path string) string {
	// Path format: /api/v1/...
	// Extract version from path
	if len(path) > 5 && path[:5] == "/api/" {
		// Find next slash after /api/
		for i := 5; i < len(path); i++ {
			if path[i] == '/' {
				return path[5:i]
			}
		}
	}
	return APIVersionV1 // Default to v1
}

// IsVersionSupported checks if the given version is supported
func IsVersionSupported(version string) bool {
	_, exists := SupportedVersions[version]
	return exists
}

// IsVersionDeprecated checks if the given version is deprecated
func IsVersionDeprecated(version string) bool {
	if v, exists := SupportedVersions[version]; exists {
		return v.Deprecated
	}
	return false
}

// BuildAPIPath builds an API path with version
// 버전이 포함된 API 경로 생성
func BuildAPIPath(version string, paths ...string) string {
	path := "/api/" + version
	for _, p := range paths {
		if p != "" {
			if p[0] != '/' {
				path += "/"
			}
			path += p
		}
	}
	return path
}
