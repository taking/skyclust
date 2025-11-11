package errors

import (
	"errors"
	"fmt"
	"strings"

	"skyclust/internal/domain"

	"github.com/aws/smithy-go"
)

// ProviderErrorConverter converts provider-specific errors to domain errors
type ProviderErrorConverter struct{}

// NewProviderErrorConverter creates a new provider error converter
func NewProviderErrorConverter() *ProviderErrorConverter {
	return &ProviderErrorConverter{}
}

// ConvertAWSError converts AWS SDK errors to domain errors
func (c *ProviderErrorConverter) ConvertAWSError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()

	// Check if it's an AWS API error
	var apiErr *smithy.OperationError
	if errors.As(err, &apiErr) {
		// Check for specific AWS error codes
		switch {
		case strings.Contains(errorMsg, "UnrecognizedClientException"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid AWS credentials", 400)
		case strings.Contains(errorMsg, "InvalidUserID.NotFound"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "AWS user not found", 400)
		case strings.Contains(errorMsg, "AccessDenied") || strings.Contains(errorMsg, "UnauthorizedOperation"):
			// Extract the specific action and user from the error message for better context
			if strings.Contains(errorMsg, "not authorized to perform") {
				// Extract action name if available
				if strings.Contains(errorMsg, "ec2:DescribeRegions") {
					return domain.NewDomainError(domain.ErrCodeForbidden, "AWS IAM permission required: The credential does not have permission to list AWS regions. Please add 'ec2:DescribeRegions' permission to the IAM user or role.", 403)
				}
				if strings.Contains(errorMsg, "eks:") {
					return domain.NewDomainError(domain.ErrCodeForbidden, "AWS IAM permission required: The credential does not have permission to access EKS. Please add the required EKS permissions to the IAM user or role.", 403)
				}
				if strings.Contains(errorMsg, "ec2:DeleteVpc") {
					return domain.NewDomainError(domain.ErrCodeForbidden, "Permission denied for VPC deletion. Please check the 'ec2:DeleteVpc' permission in your IAM policy.", 403)
				}
				return domain.NewDomainError(domain.ErrCodeForbidden, "AWS IAM permission required: The credential does not have the required permissions. Please check your IAM policy.", 403)
			}
			return domain.NewDomainError(domain.ErrCodeForbidden, "Access denied to AWS resources", 403)
		case strings.Contains(errorMsg, "NoSuchEntity") || strings.Contains(errorMsg, "InvalidVpcID.NotFound"):
			return domain.NewDomainError(domain.ErrCodeNotFound, "AWS resource not found", 404)
		case strings.Contains(errorMsg, "InvalidParameterValue"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid AWS parameter", 400)
		case strings.Contains(errorMsg, "ThrottlingException"):
			return domain.NewDomainError(domain.ErrCodeProviderQuota, "AWS API rate limit exceeded", 429)
		case strings.Contains(errorMsg, "DependencyViolation"):
			// Special handling for VPC deletion errors
			// Extract VPC ID from operation context if available
			vpcID := ""
			if strings.Contains(operation, "vpc-") {
				// Try to extract from operation string
				parts := strings.Split(operation, "vpc-")
				if len(parts) > 1 {
					vpcID = "vpc-" + strings.Fields(parts[1])[0]
				}
			}
			return c.convertAWSVPCDeleteError(err, errorMsg, vpcID)
		default:
			return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("AWS API error: %s", errorMsg), 400)
		}
	}

	// For other errors, return as internal server error
	return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to %s: %v", operation, err), 500)
}

// convertAWSVPCDeleteError converts AWS VPC deletion errors with detailed dependency information
func (c *ProviderErrorConverter) convertAWSVPCDeleteError(err error, errorMsg string, providedVPCID string) error {
	// Extract VPC ID from provided parameter, error message, or operation context
	vpcID := providedVPCID
	if vpcID == "" && strings.Contains(errorMsg, "vpc-") {
		parts := strings.Split(errorMsg, "'")
		for i, part := range parts {
			if strings.HasPrefix(part, "vpc-") {
				vpcID = part
				break
			}
			if i > 0 && strings.Contains(parts[i-1], "vpc") {
				vpcID = part
				break
			}
		}
	}
	// If still no VPC ID found, use a generic identifier
	if vpcID == "" {
		vpcID = "the VPC"
	}

	// Identify dependency type
	var dependencyType string
	if strings.Contains(errorMsg, "subnet") || strings.Contains(errorMsg, "Subnet") {
		dependencyType = "subnets"
	} else if strings.Contains(errorMsg, "security group") || strings.Contains(errorMsg, "SecurityGroup") {
		dependencyType = "security groups"
	} else if strings.Contains(errorMsg, "network interface") || strings.Contains(errorMsg, "NetworkInterface") {
		dependencyType = "network interfaces"
	} else if strings.Contains(errorMsg, "internet gateway") || strings.Contains(errorMsg, "InternetGateway") {
		dependencyType = "internet gateways"
	} else if strings.Contains(errorMsg, "route table") || strings.Contains(errorMsg, "RouteTable") {
		dependencyType = "route tables"
	} else if strings.Contains(errorMsg, "vpc peering") || strings.Contains(errorMsg, "VpcPeering") {
		dependencyType = "VPC peering connections"
	} else {
		dependencyType = "attached resources"
	}

	message := fmt.Sprintf(
		"VPC '%s' cannot be deleted because it has %s attached. Please delete or detach all attached resources before deleting the VPC.",
		vpcID,
		dependencyType,
	)

	// Add resolution steps
	message += fmt.Sprintf(
		"\n\nResolution steps:\n" +
			"1. Delete all subnets attached to the VPC\n" +
			"2. Delete or detach security groups from other resources\n" +
			"3. Detach and delete internet gateways attached to the VPC\n" +
			"4. Delete network interfaces attached to the VPC\n" +
			"5. Delete any VPC peering connections\n" +
			"6. Remove all dependent resources and try again",
	)

	return domain.NewDomainError(domain.ErrCodeConflict, message, 409)
}

// ConvertAzureError converts Azure SDK errors to domain errors
func (c *ProviderErrorConverter) ConvertAzureError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()

	// Check if it's an Azure ResponseError
	if strings.Contains(errorMsg, "ResponseError") {
		// Check for specific Azure error codes
		switch {
		case strings.Contains(errorMsg, "InvalidAuthenticationToken") || strings.Contains(errorMsg, "AuthenticationFailed"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid Azure credentials", 400)
		case strings.Contains(errorMsg, "AuthorizationFailed") || strings.Contains(errorMsg, "Forbidden"):
			return domain.NewDomainError(domain.ErrCodeForbidden, "Access denied to Azure resources. Please check your Azure RBAC permissions.", 403)
		case strings.Contains(errorMsg, "ResourceNotFound") || strings.Contains(errorMsg, "NotFound"):
			return domain.NewDomainError(domain.ErrCodeNotFound, "Azure resource not found", 404)
		case strings.Contains(errorMsg, "InvalidParameter") || strings.Contains(errorMsg, "BadRequest"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("Invalid Azure parameter: %s", errorMsg), 400)
		case strings.Contains(errorMsg, "TooManyRequests") || strings.Contains(errorMsg, "Throttled"):
			return domain.NewDomainError(domain.ErrCodeProviderQuota, "Azure API rate limit exceeded", 429)
		case strings.Contains(errorMsg, "Conflict"):
			return domain.NewDomainError(domain.ErrCodeConflict, "Azure resource conflict", 409)
		default:
			return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("Azure API error: %s", errorMsg), 400)
		}
	}

	// For other errors, return as internal server error
	return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to %s: %v", operation, err), 500)
}

// ConvertGCPError converts GCP SDK errors to domain errors
func (c *ProviderErrorConverter) ConvertGCPError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()

	// Check for specific GCP error codes
	switch {
	case strings.Contains(errorMsg, "invalid credentials") || strings.Contains(errorMsg, "authentication failed"):
		return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid GCP credentials", 400)
	case strings.Contains(errorMsg, "permission denied") || strings.Contains(errorMsg, "forbidden"):
		return domain.NewDomainError(domain.ErrCodeForbidden, "Access denied to GCP resources. Please check your GCP IAM permissions.", 403)
	case strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "NotFound"):
		return domain.NewDomainError(domain.ErrCodeNotFound, "GCP resource not found", 404)
	case strings.Contains(errorMsg, "invalid argument") || strings.Contains(errorMsg, "bad request"):
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("Invalid GCP parameter: %s", errorMsg), 400)
	case strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "quota exceeded"):
		return domain.NewDomainError(domain.ErrCodeProviderQuota, "GCP API rate limit exceeded", 429)
	case strings.Contains(errorMsg, "already exists") || strings.Contains(errorMsg, "conflict"):
		return domain.NewDomainError(domain.ErrCodeConflict, "GCP resource conflict", 409)
	default:
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("GCP API error: %s", errorMsg), 400)
	}
}

// ConvertNCPError converts NCP SDK errors to domain errors
func (c *ProviderErrorConverter) ConvertNCPError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()

	// Check for specific NCP error codes
	switch {
	case strings.Contains(errorMsg, "invalid credentials") || strings.Contains(errorMsg, "authentication failed"):
		return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid NCP credentials", 400)
	case strings.Contains(errorMsg, "permission denied") || strings.Contains(errorMsg, "forbidden"):
		return domain.NewDomainError(domain.ErrCodeForbidden, "Access denied to NCP resources. Please check your NCP IAM permissions.", 403)
	case strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "NotFound"):
		return domain.NewDomainError(domain.ErrCodeNotFound, "NCP resource not found", 404)
	case strings.Contains(errorMsg, "invalid argument") || strings.Contains(errorMsg, "bad request"):
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("Invalid NCP parameter: %s", errorMsg), 400)
	case strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "quota exceeded"):
		return domain.NewDomainError(domain.ErrCodeProviderQuota, "NCP API rate limit exceeded", 429)
	case strings.Contains(errorMsg, "already exists") || strings.Contains(errorMsg, "conflict"):
		return domain.NewDomainError(domain.ErrCodeConflict, "NCP resource conflict", 409)
	default:
		return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("NCP API error: %s", errorMsg), 400)
	}
}

// ConvertProviderError converts provider-specific errors to domain errors based on provider type
func (c *ProviderErrorConverter) ConvertProviderError(err error, provider string, operation string) error {
	if err == nil {
		return nil
	}

	switch provider {
	case domain.ProviderAWS:
		return c.ConvertAWSError(err, operation)
	case domain.ProviderAzure:
		return c.ConvertAzureError(err, operation)
	case domain.ProviderGCP:
		return c.ConvertGCPError(err, operation)
	case domain.ProviderNCP:
		return c.ConvertNCPError(err, operation)
	default:
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to %s: %v", operation, err), 500)
	}
}
