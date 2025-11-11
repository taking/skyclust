package cost_analysis

import (
	"context"
	"fmt"
	"time"

	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"go.uber.org/zap"
)

// AWS Cost Calculator Functions
// All AWS-specific cost calculation operations are implemented in this file

func (s *Service) getAWSCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	ceClient, _, err := s.getAWSCostExplorerClient(ctx, credential, vm.Region)
	if err != nil {
		return nil, err
	}

	// Prepare filters - filter by instance ID if available
	var filter *types.Expression
	if vm.InstanceID != "" {
		filter = &types.Expression{
			Dimensions: &types.DimensionValues{
				Key:    types.DimensionResourceId,
				Values: []string{vm.InstanceID},
			},
		}
	}

	result, err := s.queryAWSCostExplorer(ctx, ceClient, startDate, endDate, filter, nil)
	if err != nil {
		s.logger.Warn("Failed to get AWS costs from Cost Explorer API, falling back to estimated costs",
			zap.Error(err))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	costs := s.parseAWSCostExplorerResults(result, vm.InstanceID, vm.WorkspaceID, ResourceTypeVM)

	// If no costs found, fall back to estimated costs
	if len(costs) == 0 {
		s.logger.Warn("No AWS costs found from Cost Explorer API, falling back to estimated costs")
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	return costs, nil
}

func (s *Service) queryAWSCostExplorer(ctx context.Context, ceClient *costexplorer.Client, startDate, endDate time.Time, filter *types.Expression, groupBy []types.GroupDefinition) (*costexplorer.GetCostAndUsageOutput, error) {
	if groupBy == nil {
		groupBy = []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("REGION"),
			},
		}
	}

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate.Format(DateFormatISO)),
			End:   aws.String(endDate.Format(DateFormatISO)),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"BlendedCost", "UnblendedCost"},
		GroupBy:     groupBy,
	}

	if filter != nil {
		input.Filter = filter
	}

	return ceClient.GetCostAndUsage(ctx, input)
}

func (s *Service) parseAWSCostExplorerResults(result *costexplorer.GetCostAndUsageOutput, resourceID, workspaceID, resourceType string) []CostData {
	var costs []CostData

	for _, resultByTime := range result.ResultsByTime {
		dateStr := aws.ToString(resultByTime.TimePeriod.Start)
		date, err := time.Parse(DateFormatISO, dateStr)
		if err != nil {
			s.logger.Warn("Failed to parse date",
				zap.String("date", dateStr),
				zap.Error(err))
			continue
		}

		for _, group := range resultByTime.Groups {
			var service, region string
			for i, key := range group.Keys {
				if i == GroupKeyIndexService {
					service = key
				} else if i == GroupKeyIndexRegion {
					region = key
				}
			}

			var amount float64
			var currency string
			if blendedCost, ok := group.Metrics["BlendedCost"]; ok {
				amountStr := aws.ToString(blendedCost.Amount)
				var parseErr error
				amount, parseErr = parseFloat(amountStr)
				if parseErr != nil {
					s.logger.Warn("Failed to parse cost amount",
						zap.String("amount", amountStr),
						zap.Error(parseErr))
					continue
				}
				currency = aws.ToString(blendedCost.Unit)
			}

			if amount > 0 {
				costs = append(costs, CostData{
					Date:         date,
					Amount:       amount,
					Currency:     currency,
					Service:      service,
					ResourceID:   resourceID,
					ResourceType: resourceType,
					Provider:     ProviderAWS,
					Region:       region,
					WorkspaceID:  workspaceID,
				})
			}
		}
	}

	return costs
}

func (s *Service) getAWSKubernetesCosts(ctx context.Context, credential *domain.Credential, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, error) {
	ceClient, _, err := s.getAWSCostExplorerClient(ctx, credential, AWSDefaultRegion)
	if err != nil {
		return nil, nil, err
	}

	// Filter by EKS service
	filter := &types.Expression{
		Dimensions: &types.DimensionValues{
			Key: types.DimensionService,
			Values: []string{
				"Amazon Elastic Container Service for Kubernetes", // EKS service name
			},
		},
	}

	groupBy := []types.GroupDefinition{
		{
			Type: types.GroupDefinitionTypeDimension,
			Key:  aws.String("SERVICE"),
		},
		{
			Type: types.GroupDefinitionTypeDimension,
			Key:  aws.String("REGION"),
		},
	}

	result, err := s.queryAWSCostExplorer(ctx, ceClient, startDate, endDate, filter, groupBy)
	if err != nil {
		return nil, nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get AWS Kubernetes costs: %v", err), 502)
	}

	costs := s.parseAWSCostExplorerResults(result, "", workspaceID, ResourceTypeCluster)

	// If node groups are requested, also get EC2 costs for EKS nodes
	var warnings []CostWarning
	if includeNodeGroups {
		// Node groups are EC2 instances, so they're tracked separately
		// We can get them by filtering EC2 costs for instances with EKS tags or by service
		// For now, we'll skip node group specific costs as they're part of EC2 costs
		s.logger.Info("Node group costs are included in EC2 service costs")
	}

	return costs, warnings, nil
}
