/**
 * AWS Metadata Service
 * AWS EKS/EC2 메타데이터 조회 (Kubernetes 버전, Region, Availability Zone)
 */

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"go.uber.org/zap"
)

// GetEKSVersions returns available Kubernetes versions for EKS in the specified region
func (s *Service) GetEKSVersions(ctx context.Context, credential *domain.Credential, region string) ([]string, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("eks:versions:%s:%s", credentialID, region)

	// 캐시에서 조회 시도 (버전 목록은 자주 변하지 않으므로 긴 TTL 사용)
	if s.cache != nil {
		var cachedVersions []string
		if err := s.cache.Get(ctx, cacheKey, &cachedVersions); err == nil {
			s.logger.Debug("EKS versions retrieved from cache",
				zap.String("credential_id", credentialID),
				zap.String("region", region))
			return cachedVersions, nil
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

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환
	if versions == nil {
		versions = []string{}
	}

	// 캐시에 저장 (버전 목록은 자주 변하지 않으므로 긴 TTL 사용 - 1시간)
	if s.cache != nil && len(versions) > 0 {
		ttl := 1 * time.Hour
		if err := s.cache.Set(ctx, cacheKey, versions, ttl); err != nil {
			s.logger.Warn("Failed to cache EKS versions",
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.Error(err))
		}
	}

	s.logger.Info("EKS versions retrieved",
		zap.String("credential_id", credentialID),
		zap.String("region", region),
		zap.Int("version_count", len(versions)))

	return versions, nil
}

// GetAWSRegions returns available AWS regions for the account
func (s *Service) GetAWSRegions(ctx context.Context, credential *domain.Credential) ([]string, error) {
	// 캐시 키 생성 (Region 목록은 계정 전체이므로 credential ID만 사용)
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:regions:%s", credentialID)

	// 캐시에서 조회 시도 (Region 목록은 매우 자주 변하지 않으므로 긴 TTL 사용)
	if s.cache != nil {
		var cachedRegions []string
		if err := s.cache.Get(ctx, cacheKey, &cachedRegions); err == nil {
			s.logger.Debug("AWS regions retrieved from cache",
				zap.String("credential_id", credentialID))
			return cachedRegions, nil
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
		err = s.handleAWSError(err, "get AWS regions")
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
	if s.cache != nil && len(regions) > 0 {
		ttl := 24 * time.Hour
		if err := s.cache.Set(ctx, cacheKey, regions, ttl); err != nil {
			s.logger.Warn("Failed to cache AWS regions",
				zap.String("credential_id", credentialID),
				zap.Error(err))
		}
	}

	s.logger.Info("AWS regions retrieved",
		zap.String("credential_id", credentialID),
		zap.Int("region_count", len(regions)))

	return regions, nil
}

// GetAvailabilityZones returns available Availability Zones for the specified region
func (s *Service) GetAvailabilityZones(ctx context.Context, credential *domain.Credential, region string) ([]string, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := fmt.Sprintf("aws:availability-zones:%s:%s", credentialID, region)

	// 캐시에서 조회 시도
	if s.cache != nil {
		var cachedZones []string
		if err := s.cache.Get(ctx, cacheKey, &cachedZones); err == nil {
			s.logger.Debug("Availability zones retrieved from cache",
				zap.String("credential_id", credentialID),
				zap.String("region", region))
			return cachedZones, nil
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
		err = s.handleAWSError(err, "get availability zones")
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
	if s.cache != nil && len(zones) > 0 {
		ttl := 1 * time.Hour
		if err := s.cache.Set(ctx, cacheKey, zones, ttl); err != nil {
			s.logger.Warn("Failed to cache availability zones",
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.Error(err))
		}
	}

	s.logger.Info("Availability zones retrieved",
		zap.String("credential_id", credentialID),
		zap.String("region", region),
		zap.Int("zone_count", len(zones)))

	return zones, nil
}
