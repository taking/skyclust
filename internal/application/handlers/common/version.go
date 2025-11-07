package common

// API version constants
// API 버전 상수
const (
	APIVersionV1 = "v1"
	APIVersionV2 = "v2" // Future version
)

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
