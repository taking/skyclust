/**
 * AWS Metadata Service
 * AWS EKS/EC2 메타데이터 조회 (Kubernetes 버전, Region, Availability Zone, Instance Types, AMI Types)
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
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

// GetEKSVersions returns available Kubernetes versions for EKS in the specified region
func (s *Service) GetEKSVersions(ctx context.Context, credential *domain.Credential, region string) ([]string, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("eks:versions:%s:%s", credentialID, region)

	// 캐시에서 조회 시도 (버전 목록은 자주 변하지 않으므로 긴 TTL 사용)
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedVersions, ok := cachedValue.([]string); ok {
				s.logger.Debug(ctx, "EKS versions retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region))
				return cachedVersions, nil
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

	// Create EKS client
	_ = eks.NewFromConfig(cfg)

	// ListAddonVersions API를 사용하여 지원되는 Kubernetes 버전 정보 확인
	// DescribeVersions API는 아직 SDK에 포함되지 않았을 수 있으므로
	// 대안으로 ListAddonVersions를 사용하거나, 하드코딩된 최신 버전 목록 사용
	// 참고: AWS EKS는 일반적으로 최신 3-4개 버전을 지원
	// 일단 현재 지원되는 일반적인 버전 목록을 반환 (실제로는 AWS API 호출 필요)

	// TODO: AWS SDK에 DescribeVersions가 추가되면 사용
	// 현재는 일반적인 EKS 지원 버전 목록 반환
	// 최신 버전부터 순서대로 (2024년 기준)
	var versions []string = []string{
		"1.34",
		"1.33",
		"1.32",
		"1.31",
	}

	// 향후 AWS API 지원 시 사용할 코드:
	// 중요: 정규화 없이 원본 버전 문자열을 그대로 유지해야 함
	// AWS EKS는 "major.minor" 또는 "major.minor.patch" 형식을 사용할 수 있으며,
	// 정규화하면 정보 손실이 발생할 수 있음
	// eksClient := eks.NewFromConfig(cfg)
	// output, err := eksClient.DescribeVersions(ctx, &eks.DescribeVersionsInput{})
	// if err != nil {
	//     err = s.handleAWSError(err, "get EKS versions")
	//     if err != nil {
	//         return nil, err
	//     }
	// }
	// if output != nil && output.Versions != nil {
	//     for _, versionInfo := range output.Versions {
	//         if versionInfo.Version != nil {
	//             // 정규화 없이 원본 버전 문자열 그대로 사용
	//             versions = append(versions, *versionInfo.Version)
	//         }
	//     }
	//     // 중복 제거 및 정렬 (정규화 없이 원본 유지)
	//     versions = removeDuplicatesAndSortVersions(versions)
	// }

	// versions는 이미 빈 슬라이스로 초기화되어 있으므로 nil 체크 불필요

	// 캐시에 저장 (버전 목록은 자주 변하지 않으므로 긴 TTL 사용 - 1시간)
	if s.cacheService != nil && len(versions) > 0 {
		ttl := 1 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, versions, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache EKS versions",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "EKS versions retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region", region),
		domain.NewLogField("version_count", len(versions)))

	// 중복 제거 및 정렬 (정규화 없이 원본 유지)
	versions = removeDuplicatesAndSortEKSVersions(versions)

	return versions, nil
}

// removeDuplicatesAndSortEKSVersions removes duplicates and sorts EKS Kubernetes versions in descending order
// Supports both "major.minor" and "major.minor.patch" formats
// Important: No normalization is performed - original version strings are preserved
func removeDuplicatesAndSortEKSVersions(versions []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, v := range versions {
		if v == "" {
			continue
		}
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}

	// Sort in descending order (newest first) using semantic version comparison
	// No normalization - original version strings are preserved
	for i := 0; i < len(unique)-1; i++ {
		for j := i + 1; j < len(unique); j++ {
			if compareEKSSemanticVersion(unique[i], unique[j]) < 0 {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}

	return unique
}

// compareEKSSemanticVersion compares two EKS semantic version strings
// Supports "major.minor" and "major.minor.patch" formats
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareEKSSemanticVersion(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &num1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &num2)
		}

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	return 0
}

// GetAWSRegions returns available AWS regions for the account
func (s *Service) GetAWSRegions(ctx context.Context, credential *domain.Credential) ([]string, error) {
	// 캐시 키 생성 (Region 목록은 계정 전체이므로 credential ID만 사용)
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:regions:%s", credentialID)

	// 캐시에서 조회 시도 (Region 목록은 매우 자주 변하지 않으므로 긴 TTL 사용)
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedRegions, ok := cachedValue.([]string); ok {
				s.logger.Debug(ctx, "AWS regions retrieved from cache",
					domain.NewLogField("credential_id", credentialID))
				return cachedRegions, nil
			}
		}
	}

	// Region 목록 조회는 기본 리전(us-east-1) 사용
	creds, err := s.extractAWSCredentials(ctx, credential, "us-east-1")
	if err != nil {
		return nil, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeRegions API 호출
	output, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false), // 활성화된 리전만 반환
	})
	if err != nil {
		err = s.providerErrorConverter.ConvertAWSError(err, "get AWS regions")
		if err != nil {
			return nil, err
		}
	}

	// Region 목록 추출
	var regions []string
	if output != nil && output.Regions != nil {
		for _, region := range output.Regions {
			if region.RegionName != nil {
				regions = append(regions, *region.RegionName)
			}
		}
	}

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환
	if regions == nil {
		regions = []string{}
	}

	// 캐시에 저장 (Region 목록은 매우 자주 변하지 않으므로 긴 TTL 사용 - 24시간)
	if s.cacheService != nil && len(regions) > 0 {
		ttl := 24 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, regions, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache AWS regions",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "AWS regions retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region_count", len(regions)))

	return regions, nil
}

// GetAvailabilityZones returns available Availability Zones for the specified region
func (s *Service) GetAvailabilityZones(ctx context.Context, credential *domain.Credential, region string) ([]string, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:availability-zones:%s:%s", credentialID, region)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedZones, ok := cachedValue.([]string); ok {
				s.logger.Debug(ctx, "Availability zones retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region))
				return cachedZones, nil
			}
		}
	}

	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return nil, err
	}

	// 전달된 region을 강제로 사용 (credential에 저장된 region 무시)
	creds.Region = region

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EC2 client with explicit region
	// AWS DescribeAvailabilityZones는 EC2 client의 region 설정에 따라 해당 region의 zones만 반환
	// region-name 필터는 지원하지 않으므로 EC2 client의 region 설정만 사용
	ec2Client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		o.Region = region
	})

	// DescribeAvailabilityZones API 호출
	// EC2 client의 region 설정에 따라 해당 region의 zones만 반환됨
	output, err := ec2Client.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		err = s.providerErrorConverter.ConvertAWSError(err, "get availability zones")
		if err != nil {
			return nil, err
		}
	}

	// Zone 목록 추출
	var zones []string
	if output != nil && output.AvailabilityZones != nil {
		for _, zone := range output.AvailabilityZones {
			if zone.ZoneName != nil {
				zones = append(zones, *zone.ZoneName)
			}
		}
	}

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환
	if zones == nil {
		zones = []string{}
	}

	// 캐시에 저장 (Zone 목록은 자주 변하지 않으므로 긴 TTL 사용 - 1시간)
	if s.cacheService != nil && len(zones) > 0 {
		ttl := 1 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, zones, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache availability zones",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "Availability zones retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region", region),
		domain.NewLogField("zone_count", len(zones)))

	return zones, nil
}

// InstanceTypeInfo represents EC2 instance type information with GPU support
type InstanceTypeInfo struct {
	InstanceType string `json:"instance_type"`
	VCPU         int32  `json:"vcpu"`
	MemoryInMiB  int32  `json:"memory_in_mib"`
	HasGPU       bool   `json:"has_gpu"`
	GPUCount     int32  `json:"gpu_count,omitempty"`
	GPUName      string `json:"gpu_name,omitempty"`
	Architecture string `json:"architecture"` // x86_64, arm64
}

// GetInstanceTypes returns available EC2 instance types with GPU information for the specified region
func (s *Service) GetInstanceTypes(ctx context.Context, credential *domain.Credential, region string) ([]InstanceTypeInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:instance-types:%s:%s", credentialID, region)

	// 캐시에서 조회 시도 (인스턴스 유형 목록은 자주 변하지 않으므로 긴 TTL 사용)
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedTypes, ok := cachedValue.([]InstanceTypeInfo); ok {
				s.logger.Debug(ctx, "Instance types retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region))
				return cachedTypes, nil
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

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeInstanceTypes API 호출
	// Pagination을 위해 NextToken 사용
	var instanceTypes []InstanceTypeInfo
	var nextToken *string

	for {
		input := &ec2.DescribeInstanceTypesInput{
			MaxResults: aws.Int32(100), // 한 번에 최대 100개 조회
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := ec2Client.DescribeInstanceTypes(ctx, input)
		if err != nil {
			err = s.providerErrorConverter.ConvertAWSError(err, "get instance types")
			if err != nil {
				return nil, err
			}
		}

		if output == nil || output.InstanceTypes == nil {
			break
		}

		// 인스턴스 유형 정보 추출
		for _, instanceType := range output.InstanceTypes {
			info := InstanceTypeInfo{
				InstanceType: string(instanceType.InstanceType),
			}

			// VCPU 정보
			if instanceType.VCpuInfo != nil && instanceType.VCpuInfo.DefaultVCpus != nil {
				info.VCPU = aws.ToInt32(instanceType.VCpuInfo.DefaultVCpus)
			}

			// Memory 정보
			if instanceType.MemoryInfo != nil && instanceType.MemoryInfo.SizeInMiB != nil {
				info.MemoryInMiB = int32(aws.ToInt64(instanceType.MemoryInfo.SizeInMiB))
			}

			// GPU 정보
			if instanceType.GpuInfo != nil && len(instanceType.GpuInfo.Gpus) > 0 {
				info.HasGPU = true
				// 첫 번째 GPU 정보 사용
				gpu := instanceType.GpuInfo.Gpus[0]
				if gpu.Count != nil {
					info.GPUCount = aws.ToInt32(gpu.Count)
				}
				if gpu.Manufacturer != nil {
					info.GPUName = aws.ToString(gpu.Manufacturer)
					if gpu.Name != nil {
						info.GPUName += " " + aws.ToString(gpu.Name)
					}
				}
			} else {
				info.HasGPU = false
			}

			// 아키텍처 정보
			if instanceType.ProcessorInfo != nil && len(instanceType.ProcessorInfo.SupportedArchitectures) > 0 {
				// 첫 번째 지원 아키텍처 사용 (일반적으로 x86_64 또는 arm64)
				arch := string(instanceType.ProcessorInfo.SupportedArchitectures[0])
				info.Architecture = arch
			} else {
				// 기본값: x86_64
				info.Architecture = "x86_64"
			}

			instanceTypes = append(instanceTypes, info)
		}

		// 다음 페이지가 있는지 확인
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환
	if instanceTypes == nil {
		instanceTypes = []InstanceTypeInfo{}
	}

	// 캐시에 저장 (인스턴스 유형 목록은 자주 변하지 않으므로 긴 TTL 사용 - 24시간)
	if s.cacheService != nil && len(instanceTypes) > 0 {
		ttl := 24 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, instanceTypes, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache instance types",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "Instance types retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region", region),
		domain.NewLogField("instance_type_count", len(instanceTypes)))

	return instanceTypes, nil
}

// GetInstanceTypeInfo retrieves information for a specific instance type
func (s *Service) GetInstanceTypeInfo(ctx context.Context, credential *domain.Credential, region, instanceType string) (*InstanceTypeInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:instance-type-info:%s:%s:%s", credentialID, region, instanceType)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if info, ok := cachedValue.(*InstanceTypeInfo); ok {
				s.logger.Debug(ctx, "Instance type info retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region),
					domain.NewLogField("instance_type", instanceType))
				return info, nil
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

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeInstanceTypes API 호출 (특정 인스턴스 타입만)
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []ec2Types.InstanceType{ec2Types.InstanceType(instanceType)},
		MaxResults:    aws.Int32(1),
	}

	output, err := ec2Client.DescribeInstanceTypes(ctx, input)
	if err != nil {
		err = s.providerErrorConverter.ConvertAWSError(err, "get instance type info")
		if err != nil {
			return nil, err
		}
	}

	if output == nil || output.InstanceTypes == nil || len(output.InstanceTypes) == 0 {
		return nil, domain.NewDomainError(
			domain.ErrCodeNotFound,
			fmt.Sprintf("instance type %s not found in region %s", instanceType, region),
			404,
		)
	}

	instanceTypeData := output.InstanceTypes[0]
	info := &InstanceTypeInfo{
		InstanceType: string(instanceTypeData.InstanceType),
	}

	// VCPU 정보
	if instanceTypeData.VCpuInfo != nil && instanceTypeData.VCpuInfo.DefaultVCpus != nil {
		info.VCPU = aws.ToInt32(instanceTypeData.VCpuInfo.DefaultVCpus)
	}

	// Memory 정보
	if instanceTypeData.MemoryInfo != nil && instanceTypeData.MemoryInfo.SizeInMiB != nil {
		info.MemoryInMiB = int32(aws.ToInt64(instanceTypeData.MemoryInfo.SizeInMiB))
	}

	// GPU 정보
	if instanceTypeData.GpuInfo != nil && len(instanceTypeData.GpuInfo.Gpus) > 0 {
		info.HasGPU = true
		info.GPUCount = int32(len(instanceTypeData.GpuInfo.Gpus))
		if instanceTypeData.GpuInfo.Gpus[0].Name != nil {
			info.GPUName = *instanceTypeData.GpuInfo.Gpus[0].Name
		}
	}

	// Architecture 정보
	if instanceTypeData.ProcessorInfo != nil && len(instanceTypeData.ProcessorInfo.SupportedArchitectures) > 0 {
		info.Architecture = string(instanceTypeData.ProcessorInfo.SupportedArchitectures[0])
	} else {
		info.Architecture = "x86_64" // 기본값
	}

	// 캐시에 저장 (1시간 TTL)
	if s.cacheService != nil {
		_ = s.cacheService.Set(ctx, cacheKey, info, 1*time.Hour)
	}

	s.logger.Debug(ctx, "Instance type info retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region", region),
		domain.NewLogField("instance_type", instanceType),
		domain.NewLogField("vcpu", info.VCPU))

	return info, nil
}

// GetEKSAmitTypes returns available EKS AMI types
func (s *Service) GetEKSAmitTypes(ctx context.Context) ([]string, error) {
	// AMI Type 목록은 하드코딩 (AWS EKS에서 지원하는 AMI Type은 고정되어 있음)
	// 사용자가 제공한 목록 사용
	amiTypes := []string{
		"BOTTLEROCKET_x86_64",
		"BOTTLEROCKET_ARM_x86_64",
		"BOTTLEROCKET_x86_64_NVIDIA",
		"BOTTLEROCKET_ARM_x86_64_NVIDIA",
		"AL2023_x86_64_STANDARD",
		"AL2023_ARM_64_STANDARD",
		"AL2023_x86_64_NEURON",
		"AL2023_x86_64_NVIDIA",
		"AL2023_ARM_64_NVIDIA",
	}

	return amiTypes, nil
}

// IsGPUAMIType checks if an AMI type supports GPU
func IsGPUAMIType(amiType string) bool {
	return strings.Contains(amiType, "NVIDIA") || strings.Contains(amiType, "NEURON")
}

// GetCompatibleAMITypes returns AMI types compatible with the given instance type and GPU requirement
func GetCompatibleAMITypes(instanceType string, hasGPU bool, architecture string) []string {
	allAMITypes := []string{
		"BOTTLEROCKET_x86_64",
		"BOTTLEROCKET_ARM_x86_64",
		"BOTTLEROCKET_x86_64_NVIDIA",
		"BOTTLEROCKET_ARM_x86_64_NVIDIA",
		"AL2023_x86_64_STANDARD",
		"AL2023_ARM_64_STANDARD",
		"AL2023_x86_64_NEURON",
		"AL2023_x86_64_NVIDIA",
		"AL2023_ARM_64_NVIDIA",
	}

	var compatible []string
	for _, amiType := range allAMITypes {
		// GPU 매칭
		amiHasGPU := IsGPUAMIType(amiType)
		if hasGPU != amiHasGPU {
			continue
		}

		// 아키텍처 매칭
		isX86 := strings.Contains(amiType, "x86_64")
		isARM := strings.Contains(amiType, "ARM_64")
		if architecture == "x86_64" && !isX86 {
			continue
		}
		if architecture == "arm64" && !isARM {
			continue
		}

		compatible = append(compatible, amiType)
	}

	return compatible
}

// GetRecommendedInstanceType returns a recommended instance type based on GPU requirement
func GetRecommendedInstanceType(instanceTypes []InstanceTypeInfo, useGPU bool, architecture string) string {
	if len(instanceTypes) == 0 {
		if useGPU {
			return "g5.xlarge"
		}
		return "t3.medium"
	}

	// GPU 요구사항에 맞는 인스턴스 필터링
	var candidates []InstanceTypeInfo
	for _, it := range instanceTypes {
		if it.HasGPU != useGPU {
			continue
		}
		if it.Architecture != architecture {
			continue
		}
		candidates = append(candidates, it)
	}

	if len(candidates) == 0 {
		// 폴백: GPU 요구사항만 맞는 것
		for _, it := range instanceTypes {
			if it.HasGPU == useGPU {
				candidates = append(candidates, it)
			}
		}
	}

	if len(candidates) == 0 {
		if useGPU {
			return "g5.xlarge"
		}
		return "t3.medium"
	}

	// 추천 로직: GPU 사용 시 가장 작은 GPU 인스턴스, GPU 미사용 시 t3.medium 또는 가장 작은 인스턴스
	if useGPU {
		// GPU 인스턴스 중 가장 작은 것 (g5.xlarge 우선)
		for _, candidate := range candidates {
			if strings.HasPrefix(candidate.InstanceType, "g5.xlarge") {
				return candidate.InstanceType
			}
		}
		// g5.xlarge가 없으면 첫 번째 GPU 인스턴스
		return candidates[0].InstanceType
	} else {
		// t3.medium 우선
		for _, candidate := range candidates {
			if candidate.InstanceType == "t3.medium" {
				return candidate.InstanceType
			}
		}
		// t3.medium이 없으면 가장 작은 VCPU를 가진 인스턴스
		smallest := candidates[0]
		for _, candidate := range candidates {
			if candidate.VCPU < smallest.VCPU {
				smallest = candidate
			}
		}
		return smallest.InstanceType
	}
}

// GetRecommendedAMIType returns a recommended AMI type based on instance type and GPU requirement
func GetRecommendedAMIType(instanceType string, hasGPU bool, architecture string) string {
	compatible := GetCompatibleAMITypes(instanceType, hasGPU, architecture)
	if len(compatible) == 0 {
		if hasGPU {
			if architecture == "arm64" {
				return "AL2023_ARM_64_NVIDIA"
			}
			return "AL2023_x86_64_NVIDIA"
		} else {
			if architecture == "arm64" {
				return "AL2023_ARM_64_STANDARD"
			}
			return "AL2023_x86_64_STANDARD"
		}
	}

	// AL2023 우선 (Bottlerocket보다 일반적으로 더 많이 사용)
	for _, amiType := range compatible {
		if strings.HasPrefix(amiType, "AL2023") {
			return amiType
		}
	}

	// AL2023이 없으면 첫 번째 호환 AMI
	return compatible[0]
}

// InstanceTypeOfferingInfo represents instance type offering information for a specific availability zone
type InstanceTypeOfferingInfo struct {
	InstanceType     string `json:"instance_type"`
	AvailabilityZone string `json:"availability_zone"`
	LocationType     string `json:"location_type"` // availability-zone, region
}

// GetInstanceTypeOfferings returns available availability zones for the specified instance type in the given region
func (s *Service) GetInstanceTypeOfferings(ctx context.Context, credential *domain.Credential, region string, instanceType string) ([]InstanceTypeOfferingInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:instance-type-offerings:%s:%s:%s", credentialID, region, instanceType)

	// 캐시에서 조회 시도 (인스턴스 타입별 AZ 가용성은 자주 변하지 않으므로 긴 TTL 사용)
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedOfferings, ok := cachedValue.([]InstanceTypeOfferingInfo); ok {
				s.logger.Debug(ctx, "Instance type offerings retrieved from cache",
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region),
					domain.NewLogField("instance_type", instanceType))
				return cachedOfferings, nil
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

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeInstanceTypeOfferings API 호출
	// LocationType을 availability-zone으로 설정하여 AZ별 가용성 확인
	var offerings []InstanceTypeOfferingInfo
	var nextToken *string

	for {
		input := &ec2.DescribeInstanceTypeOfferingsInput{
			LocationType: ec2Types.LocationTypeAvailabilityZone,
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String("instance-type"),
					Values: []string{instanceType},
				},
			},
			MaxResults: aws.Int32(100),
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := ec2Client.DescribeInstanceTypeOfferings(ctx, input)
		if err != nil {
			// Log the actual AWS error for debugging
			s.logger.Warn(ctx, "AWS DescribeInstanceTypeOfferings error",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("instance_type", instanceType),
				domain.NewLogField("error", err.Error()))

			err = s.providerErrorConverter.ConvertAWSError(err, "get instance type offerings")
			if err != nil {
				return nil, err
			}
		}

		if output == nil || output.InstanceTypeOfferings == nil {
			break
		}

		// Offering 정보 추출
		for _, offering := range output.InstanceTypeOfferings {
			info := InstanceTypeOfferingInfo{
				InstanceType: string(offering.InstanceType),
				LocationType: string(offering.LocationType),
			}

			// Availability Zone 추출
			if offering.Location != nil {
				info.AvailabilityZone = *offering.Location
			}

			offerings = append(offerings, info)
		}

		// 다음 페이지가 있는지 확인
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환
	if offerings == nil {
		offerings = []InstanceTypeOfferingInfo{}
	}

	// 캐시에 저장 (인스턴스 타입별 AZ 가용성은 자주 변하지 않으므로 긴 TTL 사용 - 24시간)
	if s.cacheService != nil && len(offerings) > 0 {
		ttl := 24 * time.Hour
		if err := s.cacheService.Set(ctx, cacheKey, offerings, ttl); err != nil {
			s.logger.Warn(ctx, "Failed to cache instance type offerings",
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("instance_type", instanceType),
				domain.NewLogField("error", err))
		}
	}

	s.logger.Info(ctx, "Instance type offerings retrieved",
		domain.NewLogField("credential_id", credentialID),
		domain.NewLogField("region", region),
		domain.NewLogField("instance_type", instanceType),
		domain.NewLogField("offering_count", len(offerings)))

	return offerings, nil
}

// ValidateInstanceTypeAvailabilityZones validates if the specified instance types are available in the given availability zones
func (s *Service) ValidateInstanceTypeAvailabilityZones(ctx context.Context, credential *domain.Credential, region string, instanceTypes []string, availabilityZones []string) (map[string][]string, error) {
	// 결과 맵: instanceType -> []availableAZs
	result := make(map[string][]string)
	unavailableTypes := make([]string, 0)

	// 각 인스턴스 타입에 대해 사용 가능한 AZ 조회
	for _, instanceType := range instanceTypes {
		offerings, err := s.GetInstanceTypeOfferings(ctx, credential, region, instanceType)
		if err != nil {
			// 에러 발생 시 해당 인스턴스 타입은 검증 실패로 처리
			s.logger.Warn(ctx, "Failed to get instance type offerings for validation",
				domain.NewLogField("credential_id", credential.ID.String()),
				domain.NewLogField("region", region),
				domain.NewLogField("instance_type", instanceType),
				domain.NewLogField("error", err))
			unavailableTypes = append(unavailableTypes, instanceType)
			continue
		}

		// 사용 가능한 AZ 목록 추출
		availableAZs := make([]string, 0)
		offeringAZs := make(map[string]bool)
		for _, offering := range offerings {
			if offering.AvailabilityZone != "" {
				offeringAZs[offering.AvailabilityZone] = true
			}
		}

		// 요청된 AZ 중에서 사용 가능한 AZ만 필터링
		for _, az := range availabilityZones {
			if offeringAZs[az] {
				availableAZs = append(availableAZs, az)
			}
		}

		result[instanceType] = availableAZs

		// 요청된 AZ 중 하나도 사용 불가능한 경우
		if len(availableAZs) == 0 {
			unavailableTypes = append(unavailableTypes, instanceType)
		}
	}

	// 사용 불가능한 인스턴스 타입이 있는 경우 에러 반환
	if len(unavailableTypes) > 0 {
		// 사용 가능한 AZ 목록 조회하여 에러 메시지에 포함
		errorDetails := make(map[string]interface{})
		for _, instanceType := range unavailableTypes {
			offerings, err := s.GetInstanceTypeOfferings(ctx, credential, region, instanceType)
			if err == nil {
				availableAZs := make([]string, 0)
				for _, offering := range offerings {
					if offering.AvailabilityZone != "" {
						availableAZs = append(availableAZs, offering.AvailabilityZone)
					}
				}
				errorDetails[instanceType] = availableAZs
			}
		}

		errorMsg := fmt.Sprintf("The following instance types are not available in the selected availability zones: %v. Please select subnets from availability zones that support these instance types.", strings.Join(unavailableTypes, ", "))
		if len(errorDetails) > 0 {
			errorMsg += fmt.Sprintf(" Available zones for these instance types: %v", errorDetails)
		}

		err := domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			errorMsg,
			400,
		)
		for key, value := range errorDetails {
			err = err.WithDetails(key, value)
		}

		return nil, err
	}

	return result, nil
}
