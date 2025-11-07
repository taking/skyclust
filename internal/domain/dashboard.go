package domain

// DashboardSummary: 대시보드 요약 정보를 나타내는 도메인 타입
type DashboardSummary struct {
	WorkspaceID string          `json:"workspace_id"`
	VMs         VMStats         `json:"vms"`
	Clusters    ClusterStats    `json:"clusters"`
	Networks    NetworkStats    `json:"networks"`
	Credentials CredentialStats `json:"credentials"`
	Members     MemberStats     `json:"members"`
	LastUpdated string          `json:"last_updated"`
}

// VMStats: VM 통계 정보를 나타내는 타입
type VMStats struct {
	Total      int            `json:"total"`
	Running    int            `json:"running"`
	Stopped    int            `json:"stopped"`
	ByProvider map[string]int `json:"by_provider"`
}

// ClusterStats: Kubernetes 클러스터 통계 정보를 나타내는 타입
type ClusterStats struct {
	Total      int            `json:"total"`
	Healthy    int            `json:"healthy"`
	ByProvider map[string]int `json:"by_provider"`
}

// NetworkStats: 네트워크 리소스 통계 정보를 나타내는 타입
type NetworkStats struct {
	VPCs           int `json:"vpcs"`
	Subnets        int `json:"subnets"`
	SecurityGroups int `json:"security_groups"`
}

// CredentialStats: 자격 증명 통계 정보를 나타내는 타입
type CredentialStats struct {
	Total      int            `json:"total"`
	ByProvider map[string]int `json:"by_provider"`
}

// MemberStats: 워크스페이스 멤버 통계 정보를 나타내는 타입
type MemberStats struct {
	Total int `json:"total"`
}
