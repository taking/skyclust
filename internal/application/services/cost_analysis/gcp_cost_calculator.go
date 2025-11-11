package cost_analysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"skyclust/internal/domain"

	billingpb "cloud.google.com/go/billing/apiv1/billingpb"
	"go.uber.org/zap"
)

// GCP Cost Calculator Functions
// All GCP-specific cost calculation operations are implemented in this file

func (s *Service) getGCPCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
	// Create billing client
	billingClient, projectID, err := s.setupGCPBillingClient(ctx, credential)
	if err != nil {
		return nil, err
	}
	defer billingClient.Close()

	// Get billing account for the project
	projectName := fmt.Sprintf("projects/%s", projectID)
	req := &billingpb.GetProjectBillingInfoRequest{
		Name: projectName,
	}

	projectBillingInfo, err := billingClient.GetProjectBillingInfo(ctx, req)
	if err != nil {
		s.logger.Warn("Failed to get GCP billing info, falling back to estimated costs",
			zap.Error(err))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Check if billing is enabled
	if !projectBillingInfo.BillingEnabled {
		s.logger.Warn("Billing not enabled for project, falling back to estimated costs",
			zap.String("project_id", projectID))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Get billing account name
	billingAccountName := projectBillingInfo.BillingAccountName
	if billingAccountName == "" {
		s.logger.Warn("No billing account found for project, falling back to estimated costs",
			zap.String("project_id", projectID))
		return s.calculateEstimatedCosts(vm, startDate, endDate)
	}

	// Note: GCP Cloud Billing API doesn't provide direct cost queries like AWS Cost Explorer
	// For detailed cost data, we would need to use BigQuery Billing Export or Cloud Billing Budget API
	// Since that requires additional setup, we'll use estimated costs based on VM specifications
	// but mark it as coming from GCP pricing
	s.logger.Info("GCP billing account found, using estimated costs based on VM specifications",
		zap.String("billing_account", billingAccountName))

	return s.calculateEstimatedCosts(vm, startDate, endDate)
}

func (s *Service) getGCPKubernetesCosts(ctx context.Context, credential *domain.Credential, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, error) {
	// Create billing client
	billingClient, projectID, err := s.setupGCPBillingClient(ctx, credential)
	if err != nil {
		return nil, nil, err
	}
	defer billingClient.Close()

	// Get billing account for the project
	projectName := fmt.Sprintf("projects/%s", projectID)
	projectInfo, err := billingClient.GetProjectBillingInfo(ctx, &billingpb.GetProjectBillingInfoRequest{
		Name: projectName,
	})
	if err != nil {
		// Check if it's a permission/API disabled error
		errMsg := err.Error()
		var warnings []CostWarning
		if strings.Contains(errMsg, "SERVICE_DISABLED") || strings.Contains(errMsg, "not been used") || strings.Contains(errMsg, "disabled") {
			warnings = append(warnings, CostWarning{
				Code:         "API_NOT_ENABLED",
				Message:      "GCP Cloud Billing API is not enabled. Please enable it in the GCP Console.",
				Provider:     ProviderGCP,
				ResourceType: ResourceTypeCluster,
			})
		} else if strings.Contains(errMsg, "PermissionDenied") || strings.Contains(errMsg, "permission") {
			warnings = append(warnings, CostWarning{
				Code:         "API_PERMISSION_DENIED",
				Message:      "GCP service account does not have permission to access Cloud Billing API.",
				Provider:     ProviderGCP,
				ResourceType: ResourceTypeCluster,
			})
		}
		return nil, warnings, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get project billing info: %v", err), 502)
	}

	if projectInfo.BillingAccountName == "" {
		return nil, nil, domain.NewDomainError(domain.ErrCodeNotFound, "no billing account associated with project", 404)
	}

	// For GKE, costs are tracked under Container Service
	// Since GCP Billing API is complex, we'll use a simplified approach
	// and estimate based on cluster count (similar to VM estimation)
	// In production, you would use Cloud Billing Export to BigQuery for detailed costs
	s.logger.Info("GKE costs retrieved",
		zap.String("project_id", projectID),
		zap.String("billing_account", projectInfo.BillingAccountName))

	// Return empty costs for now - GCP Billing API requires more complex setup
	// In production, use Cloud Billing Export API or BigQuery
	var costs []CostData
	var warnings []CostWarning
	warnings = append(warnings, CostWarning{
		Code:         "GKE_COST_NOT_IMPLEMENTED",
		Message:      "GKE cost calculation requires Cloud Billing Export setup. Currently returning empty costs.",
		Provider:     ProviderGCP,
		ResourceType: ResourceTypeCluster,
	})
	s.logger.Info("GKE cost calculation requires Cloud Billing Export setup - returning empty costs")
	return costs, warnings, nil
}
