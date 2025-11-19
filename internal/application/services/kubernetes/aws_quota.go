/**
 * AWS Quota Service
 * AWS Service Quotas API를 사용한 GPU 인스턴스 quota 확인
 */

package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
)

// GPUInstanceQuotaCode represents quota code mapping for GPU instance types
// AWS Service Quotas에서 사용하는 quota code 매핑
// 참고: https://docs.aws.amazon.com/servicequotas/latest/userguide/reference_ec2-quotas.html
var GPUInstanceQuotaCodeMap = map[string]string{
	// G4 인스턴스 (NVIDIA T4 GPU)
	"g4dn.xlarge":   "L-1216C47A", // Running On-Demand G instances
	"g4dn.2xlarge":  "L-1216C47A",
	"g4dn.4xlarge":  "L-1216C47A",
	"g4dn.8xlarge":  "L-1216C47A",
	"g4dn.12xlarge": "L-1216C47A",
	"g4dn.16xlarge": "L-1216C47A",
	"g4dn.metal":    "L-1216C47A",
	"g4ad.xlarge":   "L-1216C47A",
	"g4ad.2xlarge":  "L-1216C47A",
	"g4ad.4xlarge":  "L-1216C47A",
	"g4ad.8xlarge":  "L-1216C47A",
	"g4ad.16xlarge": "L-1216C47A",

	// G5 인스턴스 (NVIDIA A10G GPU)
	"g5.xlarge":    "L-DB2E81BA", // Running On-Demand G and VT instances
	"g5.2xlarge":   "L-DB2E81BA",
	"g5.4xlarge":   "L-DB2E81BA",
	"g5.8xlarge":   "L-DB2E81BA",
	"g5.12xlarge":  "L-DB2E81BA",
	"g5.16xlarge":  "L-DB2E81BA",
	"g5.24xlarge":  "L-DB2E81BA",
	"g5.48xlarge":  "L-DB2E81BA",
	"g5g.xlarge":   "L-DB2E81BA",
	"g5g.2xlarge":  "L-DB2E81BA",
	"g5g.4xlarge":  "L-DB2E81BA",
	"g5g.8xlarge":  "L-DB2E81BA",
	"g5g.16xlarge": "L-DB2E81BA",

	// P3 인스턴스 (NVIDIA V100 GPU)
	"p3.2xlarge":    "L-417A185B", // Running On-Demand P instances
	"p3.8xlarge":    "L-417A185B",
	"p3.16xlarge":   "L-417A185B",
	"p3dn.24xlarge": "L-417A185B",

	// P4 인스턴스 (NVIDIA A100 GPU)
	"p4d.24xlarge":  "L-4EE23FB8", // Running On-Demand P instances
	"p4de.24xlarge": "L-4EE23FB8",

	// P5 인스턴스 (NVIDIA H100 GPU)
	"p5.48xlarge": "L-4EE23FB8",

	// Inf1 인스턴스 (AWS Inferentia)
	"inf1.xlarge":   "L-1945791B", // Running On-Demand Inf instances
	"inf1.2xlarge":  "L-1945791B",
	"inf1.6xlarge":  "L-1945791B",
	"inf1.24xlarge": "L-1945791B",

	// Trn1 인스턴스 (AWS Trainium)
	"trn1.2xlarge":   "L-1945791B", // Running On-Demand Inf instances
	"trn1.32xlarge":  "L-1945791B",
	"trn1n.32xlarge": "L-1945791B",
}

// GetGPUQuotaCode returns the quota code for a given GPU instance type
func GetGPUQuotaCode(instanceType string) (string, bool) {
	// 정확한 매칭 시도
	if code, ok := GPUInstanceQuotaCodeMap[strings.ToLower(instanceType)]; ok {
		return code, true
	}

	// 패턴 매칭 (g4.*, g5.*, p3.*, p4.*, p5.*, inf1.*, trn1.*)
	instanceTypeLower := strings.ToLower(instanceType)
	if strings.HasPrefix(instanceTypeLower, "g4") {
		return "L-1216C47A", true // G4 instances
	}
	if strings.HasPrefix(instanceTypeLower, "g5") {
		return "L-DB2E81BA", true // G5 instances
	}
	if strings.HasPrefix(instanceTypeLower, "p3") {
		return "L-417A185B", true // P3 instances
	}
	if strings.HasPrefix(instanceTypeLower, "p4") {
		return "L-4EE23FB8", true // P4 instances
	}
	if strings.HasPrefix(instanceTypeLower, "p5") {
		return "L-4EE23FB8", true // P5 instances
	}
	if strings.HasPrefix(instanceTypeLower, "inf1") {
		return "L-1945791B", true // Inf1 instances
	}
	if strings.HasPrefix(instanceTypeLower, "trn1") {
		return "L-1945791B", true // Trn1 instances
	}

	return "", false
}

// GPUQuotaInfo represents GPU instance quota information
type GPUQuotaInfo struct {
	InstanceType string  `json:"instance_type"`
	Region       string  `json:"region"`
	QuotaCode    string  `json:"quota_code"`
	QuotaValue   float64 `json:"quota_value"`
	QuotaUnit    string  `json:"quota_unit"`
	Adjustable   bool    `json:"adjustable"`
	HasQuota     bool    `json:"has_quota"`
	QuotaName    string  `json:"quota_name,omitempty"`
}

// GPUQuotaAvailability represents quota availability check result
type GPUQuotaAvailability struct {
	InstanceType      string  `json:"instance_type"`
	Region            string  `json:"region"`
	Available         bool    `json:"available"`
	QuotaValue        float64 `json:"quota_value"`
	CurrentUsage      float64 `json:"current_usage,omitempty"`
	AvailableQuota    float64 `json:"available_quota"`
	RequiredCount     int32   `json:"required_count"`
	QuotaInsufficient bool    `json:"quota_insufficient"`
	Message           string  `json:"message,omitempty"`
}

// AvailableRegion represents a region with available GPU quota
type AvailableRegion struct {
	Region         string  `json:"region"`
	AvailableQuota float64 `json:"available_quota"`
	QuotaValue     float64 `json:"quota_value"`
	CurrentUsage   float64 `json:"current_usage,omitempty"`
}

// GetGPUInstanceQuota retrieves GPU instance quota for a specific region and instance type
func (s *Service) GetGPUInstanceQuota(ctx context.Context, credential *domain.Credential, region, instanceType string) (*GPUQuotaInfo, error) {
	// GPU 인스턴스 타입인지 확인
	quotaCode, isGPU := GetGPUQuotaCode(instanceType)
	if !isGPU {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			fmt.Sprintf("instance type %s is not a GPU instance type", instanceType),
			HTTPStatusBadRequest,
		)
	}

	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:gpu-quota:%s:%s:%s:%s", credentialID, region, instanceType, quotaCode)

	// 캐시에서 조회 시도 (quota는 자주 변하지 않으므로 긴 TTL 사용 - 1시간)
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if quotaInfo, ok := cachedValue.(*GPUQuotaInfo); ok {
				s.logger.Debug(ctx, "GPU quota retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region),
					domain.NewLogField("instance_type", instanceType))
				return quotaInfo, nil
			}
		}
	}

	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return nil, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create Service Quotas client
	servicequotasClient := servicequotas.NewFromConfig(cfg)

	// GetServiceQuota API 호출
	input := &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String("ec2"),
		QuotaCode:   aws.String(quotaCode),
	}

	output, err := servicequotasClient.GetServiceQuota(ctx, input)
	if err != nil {
		// Service Quotas API 에러 처리
		err = s.providerErrorConverter.ConvertAWSError(err, "get GPU instance quota")
		if err != nil {
			// Quota 정보를 찾을 수 없는 경우 (quota가 설정되지 않았거나 권한 부족)
			// 빈 quota 정보 반환 (quota가 0인 것으로 간주)
			quotaInfo := &GPUQuotaInfo{
				InstanceType: instanceType,
				Region:       region,
				QuotaCode:    quotaCode,
				QuotaValue:   0,
				QuotaUnit:    "Count",
				Adjustable:   false,
				HasQuota:     false,
			}

			// 캐시에 저장 (짧은 TTL - 15분, 에러는 자주 재시도하지 않음)
			if s.cacheService != nil {
				ttl := 15 * time.Minute
				_ = s.cacheService.Set(ctx, cacheKey, quotaInfo, ttl)
			}

			return quotaInfo, nil
		}
		return nil, err
	}

	// Quota 정보 추출
	quotaInfo := &GPUQuotaInfo{
		InstanceType: instanceType,
		Region:       region,
		QuotaCode:    quotaCode,
		HasQuota:     true,
		Adjustable:   false,
	}

	if output.Quota != nil {
		if output.Quota.Value != nil {
			quotaInfo.QuotaValue = *output.Quota.Value
		}
		if output.Quota.Unit != nil {
			quotaInfo.QuotaUnit = *output.Quota.Unit
		}
		quotaInfo.Adjustable = output.Quota.Adjustable
		if output.Quota.QuotaName != nil {
			quotaInfo.QuotaName = *output.Quota.QuotaName
		}
	}

	// 캐시에 저장 (quota는 자주 변하지 않으므로 긴 TTL 사용 - 1시간)
	if s.cacheService != nil {
		ttl := 1 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, quotaInfo, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache GPU quota",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("instance_type", instanceType),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "GPU quota retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region", region),
		domain.NewLogField("instance_type", instanceType),
		domain.NewLogField("quota_value", quotaInfo.QuotaValue))

	return quotaInfo, nil
}

// CheckGPUQuotaAvailability checks if GPU quota is available for the requested instance count
func (s *Service) CheckGPUQuotaAvailability(ctx context.Context, credential *domain.Credential, region, instanceType string, requiredCount int32) (*GPUQuotaAvailability, error) {
	// GPU quota 정보 조회
	quotaInfo, err := s.GetGPUInstanceQuota(ctx, credential, region, instanceType)
	if err != nil {
		return nil, err
	}

	// 현재 사용량 조회 (EC2 DescribeInstances 사용)
	currentUsage, err := s.getCurrentGPUInstanceUsage(ctx, credential, region, instanceType)
	if err != nil {
		// 사용량 조회 실패 시 0으로 간주 (에러를 반환하지 않고 계속 진행)
		s.logger.Warn(ctx, "Failed to get current GPU instance usage, assuming 0",
			domain.NewLogField("credential_id", credential.ID.String()),
			domain.NewLogField("region", region),
			domain.NewLogField("instance_type", instanceType),
			domain.NewLogField("error", err))
		currentUsage = 0
	}

	// 사용 가능한 quota 계산
	availableQuota := quotaInfo.QuotaValue - currentUsage
	available := availableQuota >= float64(requiredCount)
	quotaInsufficient := quotaInfo.QuotaValue < float64(requiredCount) || availableQuota < float64(requiredCount)

	var message string
	if quotaInsufficient {
		if quotaInfo.QuotaValue == 0 {
			message = fmt.Sprintf("GPU instance type %s has no quota allocated in region %s. Please request a quota increase or select a different region.", instanceType, region)
		} else if availableQuota < float64(requiredCount) {
			message = fmt.Sprintf("Insufficient GPU quota for instance type %s in region %s. Available: %.0f, Required: %d. Current usage: %.0f, Total quota: %.0f", instanceType, region, availableQuota, requiredCount, currentUsage, quotaInfo.QuotaValue)
		} else {
			message = fmt.Sprintf("GPU quota limit reached for instance type %s in region %s. Total quota: %.0f, Required: %d", instanceType, region, quotaInfo.QuotaValue, requiredCount)
		}
	}

	return &GPUQuotaAvailability{
		InstanceType:      instanceType,
		Region:            region,
		Available:         available,
		QuotaValue:        quotaInfo.QuotaValue,
		CurrentUsage:      currentUsage,
		AvailableQuota:    availableQuota,
		RequiredCount:     requiredCount,
		QuotaInsufficient: quotaInsufficient,
		Message:           message,
	}, nil
}

// getCurrentGPUInstanceUsage retrieves current usage of GPU instances in the region
func (s *Service) getCurrentGPUInstanceUsage(ctx context.Context, credential *domain.Credential, region, instanceType string) (float64, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return 0, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return 0, err
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeInstances로 현재 실행 중인 인스턴스 조회
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "pending"},
			},
			{
				Name:   aws.String("instance-type"),
				Values: []string{instanceType}, // 정확한 인스턴스 타입 매칭
			},
		},
	}

	var totalCount int32 = 0
	var nextToken *string

	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := ec2Client.DescribeInstances(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to describe instances: %w", err)
		}

		if output == nil {
			break
		}

		// 실행 중인 인스턴스 수 계산
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if instance.InstanceType == ec2Types.InstanceType(instanceType) {
					totalCount++
				}
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return float64(totalCount), nil
}

// GetAvailableRegionsForGPU retrieves regions with available GPU quota
func (s *Service) GetAvailableRegionsForGPU(ctx context.Context, credential *domain.Credential, instanceType string, requiredCount int32) ([]AvailableRegion, error) {
	// GPU 인스턴스 타입인지 확인
	_, isGPU := GetGPUQuotaCode(instanceType)
	if !isGPU {
		return nil, domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			fmt.Sprintf("instance type %s is not a GPU instance type", instanceType),
			HTTPStatusBadRequest,
		)
	}

	// 주요 AWS region 목록
	regions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
		"ap-southeast-1", "ap-southeast-2", "ap-south-1",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"ca-central-1", "sa-east-1",
	}

	var availableRegions []AvailableRegion

	// 각 region에 대해 quota 확인 (병렬 처리 고려 가능하나, rate limit 고려하여 순차 처리)
	for _, region := range regions {
		availability, err := s.CheckGPUQuotaAvailability(ctx, credential, region, instanceType, requiredCount)
		if err != nil {
			// 에러 발생 시 해당 region은 스킵
			s.logger.Warn(ctx, "Failed to check GPU quota for region",
				domain.NewLogField("credential_id", credential.ID.String()),
				domain.NewLogField("region", region),
				domain.NewLogField("instance_type", instanceType),
				domain.NewLogField("error", err))
			continue
		}

		// 사용 가능한 region만 추가
		if availability.Available {
			availableRegions = append(availableRegions, AvailableRegion{
				Region:         region,
				AvailableQuota: availability.AvailableQuota,
				QuotaValue:     availability.QuotaValue,
				CurrentUsage:   availability.CurrentUsage,
			})
		}
	}

	return availableRegions, nil
}

// CPUQuotaInfo represents vCPU quota information
type CPUQuotaInfo struct {
	Region     string  `json:"region"`
	QuotaCode  string  `json:"quota_code"`
	QuotaValue float64 `json:"quota_value"`
	QuotaUnit  string  `json:"quota_unit"`
	Adjustable bool    `json:"adjustable"`
	HasQuota   bool    `json:"has_quota"`
	QuotaName  string  `json:"quota_name,omitempty"`
}

// CPUQuotaAvailability represents vCPU quota availability check result
type CPUQuotaAvailability struct {
	Region            string  `json:"region"`
	Available         bool    `json:"available"`
	QuotaValue        float64 `json:"quota_value"`
	CurrentUsage      float64 `json:"current_usage,omitempty"`
	AvailableQuota    float64 `json:"available_quota"`
	RequiredVCPU      int32   `json:"required_vcpu"`
	QuotaInsufficient bool    `json:"quota_insufficient"`
	Message           string  `json:"message,omitempty"`
}

// InstanceFamilyVCPUQuotaCodeMap: 인스턴스 패밀리별 vCPU quota 코드 매핑
// AWS Service Quotas에서 사용하는 vCPU quota code 매핑
// 참고: https://docs.aws.amazon.com/servicequotas/latest/userguide/reference_ec2-quotas.html
var InstanceFamilyVCPUQuotaCodeMap = map[string]string{
	// Standard 패밀리 (A, C, D, H, I, M, R, T, Z)
	"standard": "L-34B43A08", // Running On-Demand Standard (A, C, D, H, I, M, R, T, Z) instances

	// G 패밀리 (GPU 인스턴스)
	"g": "L-DB2E81BA", // Running On-Demand G and VT instances

	// P 패밀리 (GPU 인스턴스)
	"p": "L-4EE23FB8", // Running On-Demand P instances

	// Inf 패밀리 (Inferentia/Trainium)
	"inf": "L-1945791B", // Running On-Demand Inf instances
}

// mapGPUQuotaCodeToVCPUQuotaCode: GPU quota 코드를 vCPU quota 코드로 매핑
// GPU 인스턴스는 GPU quota와 vCPU quota가 별도로 관리됨
func mapGPUQuotaCodeToVCPUQuotaCode(gpuQuotaCode string) string {
	switch gpuQuotaCode {
	case "L-1216C47A": // G4 instances (GPU quota)
		return "L-DB2E81BA" // G 패밀리 vCPU quota
	case "L-DB2E81BA": // G5 instances (GPU quota)
		return "L-DB2E81BA" // G 패밀리 vCPU quota
	case "L-417A185B": // P3 instances (GPU quota)
		return "L-4EE23FB8" // P 패밀리 vCPU quota
	case "L-4EE23FB8": // P4/P5 instances (GPU quota)
		return "L-4EE23FB8" // P 패밀리 vCPU quota
	case "L-1945791B": // Inf/Trn instances (GPU quota)
		return "L-1945791B" // Inf 패밀리 vCPU quota
	default:
		return "L-34B43A08" // 기본값: Standard 패밀리
	}
}

// GetInstanceFamilyVCPUQuotaCode: 인스턴스 타입에서 패밀리를 추출하여 vCPU quota 코드 반환
func GetInstanceFamilyVCPUQuotaCode(instanceType string) string {
	instanceTypeLower := strings.ToLower(instanceType)

	// GPU 인스턴스인 경우 GPU quota 코드를 통해 vCPU quota 코드 결정
	if gpuQuotaCode, isGPU := GetGPUQuotaCode(instanceType); isGPU {
		return mapGPUQuotaCodeToVCPUQuotaCode(gpuQuotaCode)
	}

	// 일반 인스턴스 패밀리 식별
	// 패밀리는 인스턴스 타입의 첫 번째 문자 또는 접두사로 식별
	if strings.HasPrefix(instanceTypeLower, "a") {
		return "L-34B43A08" // Standard 패밀리
	}
	if strings.HasPrefix(instanceTypeLower, "c") {
		return "L-34B43A08" // Standard 패밀리
	}
	if strings.HasPrefix(instanceTypeLower, "d") {
		return "L-34B43A08" // Standard 패밀리
	}
	if strings.HasPrefix(instanceTypeLower, "h") {
		return "L-34B43A08" // Standard 패밀리
	}
	if strings.HasPrefix(instanceTypeLower, "i") && !strings.HasPrefix(instanceTypeLower, "inf") {
		return "L-34B43A08" // Standard 패밀리 (I 인스턴스, Inf 제외)
	}
	if strings.HasPrefix(instanceTypeLower, "m") {
		return "L-34B43A08" // Standard 패밀리
	}
	if strings.HasPrefix(instanceTypeLower, "r") {
		return "L-34B43A08" // Standard 패밀리
	}
	if strings.HasPrefix(instanceTypeLower, "t") && !strings.HasPrefix(instanceTypeLower, "trn") {
		return "L-34B43A08" // Standard 패밀리 (T 인스턴스, Trn 제외)
	}
	if strings.HasPrefix(instanceTypeLower, "z") {
		return "L-34B43A08" // Standard 패밀리
	}

	// 기본값: Standard 패밀리
	return "L-34B43A08"
}

// GetCPUQuotaCode returns the quota code for vCPU limits (deprecated, use GetInstanceFamilyVCPUQuotaCode instead)
// L-34B43A08: Running On-Demand Standard (A, C, D, H, I, M, R, T, Z) instances
// This is the general vCPU quota for most standard instance types
func GetCPUQuotaCode() string {
	return "L-34B43A08"
}

// GetCPUQuota retrieves vCPU quota for a specific region and instance type
// instanceType이 제공되면 해당 인스턴스 타입의 패밀리에 맞는 quota 코드를 사용
// instanceType이 빈 문자열이면 기본 Standard 패밀리 quota 코드 사용 (하위 호환성)
func (s *Service) GetCPUQuota(ctx context.Context, credential *domain.Credential, region string, instanceType string) (*CPUQuotaInfo, error) {
	var quotaCode string
	if instanceType != "" {
		quotaCode = GetInstanceFamilyVCPUQuotaCode(instanceType)
	} else {
		quotaCode = GetCPUQuotaCode() // 하위 호환성
	}

	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:cpu-quota:%s:%s:%s", credentialID, region, quotaCode)

	// 캐시에서 조회 시도 (quota는 자주 변하지 않으므로 긴 TTL 사용 - 1시간)
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if quotaInfo, ok := cachedValue.(*CPUQuotaInfo); ok {
				s.logger.Debug(ctx, "CPU quota retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region))
				return quotaInfo, nil
			}
		}
	}

	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return nil, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create Service Quotas client
	servicequotasClient := servicequotas.NewFromConfig(cfg)

	// GetServiceQuota API 호출
	input := &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String("ec2"),
		QuotaCode:   aws.String(quotaCode),
	}

	output, err := servicequotasClient.GetServiceQuota(ctx, input)
	if err != nil {
		// Service Quotas API 에러 처리
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "NoSuchResourceException") {
			// Quota 정보가 없는 경우 (신규 계정 등)
			quotaInfo := &CPUQuotaInfo{
				Region:     region,
				QuotaCode:  quotaCode,
				QuotaValue: 0,
				QuotaUnit:  "vCPU",
				Adjustable: false,
				HasQuota:   false,
				QuotaName:  "Running On-Demand Standard instances",
			}

			// 캐시에 저장 (짧은 TTL - 5분)
			if s.cacheService != nil {
				_ = s.cacheService.Set(ctx, cacheKey, quotaInfo, 5*time.Minute)
			}

			return quotaInfo, nil
		}

		// 다른 에러는 변환하여 반환
		err = s.providerErrorConverter.ConvertAWSError(err, "get CPU quota")
		if err != nil {
			return nil, err
		}
	}

	// Quota 정보 추출
	var quotaValue float64
	var quotaUnit string = "vCPU"
	var adjustable bool = false
	var quotaName string = "Running On-Demand Standard instances"

	if output != nil && output.Quota != nil {
		if output.Quota.Value != nil {
			quotaValue = *output.Quota.Value
		}
		if output.Quota.Unit != nil {
			quotaUnit = *output.Quota.Unit
		}
		adjustable = output.Quota.Adjustable
		if output.Quota.QuotaName != nil {
			quotaName = *output.Quota.QuotaName
		}
	}

	quotaInfo := &CPUQuotaInfo{
		Region:     region,
		QuotaCode:  quotaCode,
		QuotaValue: quotaValue,
		QuotaUnit:  quotaUnit,
		Adjustable: adjustable,
		HasQuota:   quotaValue > 0,
		QuotaName:  quotaName,
	}

	// 캐시에 저장 (1시간 TTL)
	if s.cacheService != nil {
		_ = s.cacheService.Set(ctx, cacheKey, quotaInfo, 1*time.Hour)
	}

	s.logger.Info(ctx, "CPU quota retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region", region),
		domain.NewLogField("quota_value", quotaInfo.QuotaValue))

	return quotaInfo, nil
}

// FamilyQuotaCheck represents quota check result for a specific instance family
type FamilyQuotaCheck struct {
	QuotaCode      string   `json:"quota_code"`
	InstanceTypes  []string `json:"instance_types"`
	QuotaValue     float64  `json:"quota_value"`
	CurrentUsage   float64  `json:"current_usage"`
	AvailableQuota float64  `json:"available_quota"`
	RequiredVCPU   int32    `json:"required_vcpu"`
	Available      bool     `json:"available"`
	QuotaName      string   `json:"quota_name,omitempty"`
}

// CheckCPUQuotaAvailability checks if vCPU quota is available for the requested instance types and count
// 인스턴스 타입들을 패밀리별로 그룹화하여 각 패밀리의 quota를 확인
func (s *Service) CheckCPUQuotaAvailability(ctx context.Context, credential *domain.Credential, region string, instanceTypes []string, requiredCount int32) (*CPUQuotaAvailability, error) {
	// 인스턴스 타입들을 패밀리별로 그룹화 (같은 quota 코드를 사용하는 인스턴스들을 그룹화)
	familyQuotaMap := make(map[string][]string) // quotaCode -> instanceTypes

	for _, instanceType := range instanceTypes {
		quotaCode := GetInstanceFamilyVCPUQuotaCode(instanceType)
		familyQuotaMap[quotaCode] = append(familyQuotaMap[quotaCode], instanceType)
	}

	// 각 패밀리별로 quota 확인
	var familyChecks []FamilyQuotaCheck
	totalRequiredVCPU := int32(0)
	allAvailable := true
	var insufficientFamilies []string

	for quotaCode, familyInstanceTypes := range familyQuotaMap {
		// 패밀리별 vCPU 수 계산
		familyRequiredVCPU := int32(0)
		for _, instanceType := range familyInstanceTypes {
			instanceTypeInfo, err := s.GetInstanceTypeInfo(ctx, credential, region, instanceType)
			if err != nil {
				s.logger.Warn(ctx, "Failed to get instance type info for vCPU calculation, using default",
					domain.NewLogField("credential_id", credential.ID.String()),
					domain.NewLogField("region", region),
					domain.NewLogField("instance_type", instanceType),
					domain.NewLogField("error", err))
				familyRequiredVCPU += 4 * requiredCount // 기본값 4 vCPU per instance
				continue
			}

			if instanceTypeInfo != nil && instanceTypeInfo.VCPU > 0 {
				familyRequiredVCPU += instanceTypeInfo.VCPU * requiredCount
			} else {
				familyRequiredVCPU += 4 * requiredCount
			}
		}

		// 패밀리별 quota 정보 조회 (첫 번째 인스턴스 타입 사용)
		quotaInfo, err := s.GetCPUQuota(ctx, credential, region, familyInstanceTypes[0])
		if err != nil {
			s.logger.Warn(ctx, "Failed to get CPU quota for family, assuming insufficient",
				domain.NewLogField("credential_id", credential.ID.String()),
				domain.NewLogField("region", region),
				domain.NewLogField("quota_code", quotaCode),
				domain.NewLogField("instance_types", familyInstanceTypes),
				domain.NewLogField("error", err))
			allAvailable = false
			insufficientFamilies = append(insufficientFamilies, quotaCode)
			continue
		}

		// 패밀리별 현재 사용량 조회
		familyCurrentUsage, err := s.getCurrentVCPUUsageByFamily(ctx, credential, region, quotaCode, familyInstanceTypes)
		if err != nil {
			s.logger.Warn(ctx, "Failed to get current vCPU usage for family, assuming 0",
				domain.NewLogField("credential_id", credential.ID.String()),
				domain.NewLogField("region", region),
				domain.NewLogField("quota_code", quotaCode),
				domain.NewLogField("error", err))
			familyCurrentUsage = 0
		}

		// 패밀리별 사용 가능한 quota 계산
		familyAvailableQuota := quotaInfo.QuotaValue - familyCurrentUsage
		familyAvailable := familyAvailableQuota >= float64(familyRequiredVCPU)

		if !familyAvailable {
			allAvailable = false
			insufficientFamilies = append(insufficientFamilies, quotaCode)
		}

		familyChecks = append(familyChecks, FamilyQuotaCheck{
			QuotaCode:      quotaCode,
			InstanceTypes:  familyInstanceTypes,
			QuotaValue:     quotaInfo.QuotaValue,
			CurrentUsage:   familyCurrentUsage,
			AvailableQuota: familyAvailableQuota,
			RequiredVCPU:   familyRequiredVCPU,
			Available:      familyAvailable,
			QuotaName:      quotaInfo.QuotaName,
		})

		totalRequiredVCPU += familyRequiredVCPU
	}

	// 전체 사용량 계산 (모든 패밀리 합계)
	totalCurrentUsage := float64(0)
	for _, check := range familyChecks {
		totalCurrentUsage += check.CurrentUsage
	}

	// 전체 quota는 각 패밀리별 quota의 합계가 아니라, 가장 제한적인 패밀리의 quota를 기준으로 판단
	// 하지만 실제로는 각 패밀리별로 독립적으로 관리되므로, 모든 패밀리가 충족되어야 함
	quotaInsufficient := !allAvailable

	var message string
	if quotaInsufficient {
		var insufficientDetails []string
		for _, check := range familyChecks {
			if !check.Available {
				if check.QuotaValue == 0 {
					insufficientDetails = append(insufficientDetails,
						fmt.Sprintf("Family %s (quota code: %s): vCPU quota not configured (required: %d vCPU)",
							strings.Join(check.InstanceTypes, ", "), check.QuotaCode, check.RequiredVCPU))
				} else {
					insufficientDetails = append(insufficientDetails,
						fmt.Sprintf("Family %s (quota code: %s): Current quota: %.0f vCPU, Current usage: %.0f vCPU, Available: %.0f vCPU, Required: %d vCPU",
							strings.Join(check.InstanceTypes, ", "), check.QuotaCode, check.QuotaValue, check.CurrentUsage, check.AvailableQuota, check.RequiredVCPU))
				}
			}
		}

		message = fmt.Sprintf("Insufficient vCPU quota in region %s. %s Please request a quota increase at http://aws.amazon.com/contact-us/ec2-request",
			region, strings.Join(insufficientDetails, "; "))
	} else {
		message = fmt.Sprintf("vCPU quota is sufficient for all instance families. Total required: %d vCPU", totalRequiredVCPU)
	}

	return &CPUQuotaAvailability{
		Region:            region,
		Available:         allAvailable,
		QuotaValue:        totalCurrentUsage, // 전체 사용량 (각 패밀리별로 독립적이므로 의미 있는 합계는 아님)
		CurrentUsage:      totalCurrentUsage,
		AvailableQuota:    float64(totalRequiredVCPU), // 사용 가능한 quota는 패밀리별로 다르므로 의미 있는 합계는 아님
		RequiredVCPU:      totalRequiredVCPU,
		QuotaInsufficient: quotaInsufficient,
		Message:           message,
	}, nil
}

// getCurrentVCPUUsageByFamily retrieves current vCPU usage for a specific instance family in the region
func (s *Service) getCurrentVCPUUsageByFamily(ctx context.Context, credential *domain.Credential, region string, quotaCode string, familyInstanceTypes []string) (float64, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return 0, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return 0, err
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeInstances로 현재 실행 중인 인스턴스 조회 (특정 패밀리만)
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "pending"},
			},
		},
	}

	var totalVCPU int32 = 0
	var nextToken *string

	// 패밀리 인스턴스 타입 목록을 map으로 변환하여 빠른 조회
	familyInstanceTypeMap := make(map[string]bool)
	for _, it := range familyInstanceTypes {
		familyInstanceTypeMap[strings.ToLower(it)] = true
	}

	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := ec2Client.DescribeInstances(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to describe instances: %w", err)
		}

		if output == nil {
			break
		}

		// 실행 중인 인스턴스의 vCPU 수 계산 (패밀리에 속하는 인스턴스만)
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if instance.InstanceType != "" {
					instanceTypeStr := strings.ToLower(string(instance.InstanceType))
					// 패밀리에 속하는 인스턴스 타입인지 확인
					if familyInstanceTypeMap[instanceTypeStr] {
						instanceTypeInfo, err := s.GetInstanceTypeInfo(ctx, credential, region, string(instance.InstanceType))
						if err == nil && instanceTypeInfo != nil && instanceTypeInfo.VCPU > 0 {
							totalVCPU += instanceTypeInfo.VCPU
						} else {
							totalVCPU += 4 // 기본값
						}
					}
				}
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return float64(totalVCPU), nil
}

// getCurrentVCPUUsage retrieves current vCPU usage in the region
func (s *Service) getCurrentVCPUUsage(ctx context.Context, credential *domain.Credential, region string) (float64, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return 0, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return 0, err
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeInstances로 현재 실행 중인 인스턴스 조회
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "pending"},
			},
		},
	}

	var totalVCPU int32 = 0
	var nextToken *string

	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := ec2Client.DescribeInstances(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to describe instances: %w", err)
		}

		if output == nil {
			break
		}

		// 실행 중인 인스턴스의 vCPU 수 계산
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if instance.InstanceType != "" {
					// 인스턴스 타입 정보 조회 (캐시 활용)
					instanceTypeInfo, err := s.GetInstanceTypeInfo(ctx, credential, region, string(instance.InstanceType))
					if err == nil && instanceTypeInfo != nil && instanceTypeInfo.VCPU > 0 {
						totalVCPU += instanceTypeInfo.VCPU
					} else {
						// 기본값 사용 (보수적 접근)
						totalVCPU += 4
					}
				}
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return float64(totalVCPU), nil
}
