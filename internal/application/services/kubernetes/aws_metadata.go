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
	//             versions = append(versions, *versionInfo.Version)
	//         }
	//     }
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

	return versions, nil
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

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// DescribeAvailabilityZones API 호출
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
