package network

import (
	networkh "skyclust/internal/application/handlers/network/providers"
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes sets up network resource routes for a specific provider using Factory pattern
// provider: "aws", "gcp", "azure", "ncp"
func SetupRoutes(router *gin.RouterGroup, networkService *networkservice.Service, credentialService domain.CredentialService, provider string, logger ...*zap.Logger) {
	var zapLogger *zap.Logger
	if len(logger) > 0 && logger[0] != nil {
		zapLogger = logger[0]
	}

	factory := networkh.NewFactory(networkService, credentialService, zapLogger)
	handler, err := factory.GetHandler(provider)
	if err != nil {
		return
	}

	// VPC management
	// Path: /api/v1/{provider}/network/vpcs
	router.GET("/vpcs", handler.ListVPCs)
	router.POST("/vpcs", handler.CreateVPC)
	router.GET("/vpcs/:id", handler.GetVPC)
	router.PUT("/vpcs/:id", handler.UpdateVPC)
	router.DELETE("/vpcs/:id", handler.DeleteVPC)

	// Subnet management
	// Path: /api/v1/{provider}/network/subnets
	router.GET("/subnets", handler.ListSubnets)
	router.POST("/subnets", handler.CreateSubnet)
	router.GET("/subnets/:id", handler.GetSubnet)
	router.PUT("/subnets/:id", handler.UpdateSubnet)
	router.DELETE("/subnets/:id", handler.DeleteSubnet)

	// Security Group / Firewall Rules management
	// Path: /api/v1/{provider}/network/security-groups or /firewall-rules (GCP)
	if provider == domain.ProviderGCP {
		// GCP uses "Firewall Rules" terminology
		router.GET("/firewall-rules", handler.ListSecurityGroups)
		router.POST("/firewall-rules", handler.CreateSecurityGroup)
		router.GET("/firewall-rules/:id", handler.GetSecurityGroup)
		router.PUT("/firewall-rules/:id", handler.UpdateSecurityGroup)
		router.DELETE("/firewall-rules/:id", handler.DeleteSecurityGroup)

		// Security Group Rule management
		// Path: /api/v1/gcp/network/firewall-rules/:id/rules
		router.POST("/firewall-rules/:id/rules", handler.AddSecurityGroupRule)
		router.DELETE("/firewall-rules/:id/rules", handler.RemoveSecurityGroupRule)
		router.PUT("/firewall-rules/:id/rules", handler.UpdateSecurityGroupRules)
	} else {
		// Other providers use "Security Groups" terminology
		router.GET("/security-groups", handler.ListSecurityGroups)
		router.POST("/security-groups", handler.CreateSecurityGroup)
		router.GET("/security-groups/:id", handler.GetSecurityGroup)
		router.PUT("/security-groups/:id", handler.UpdateSecurityGroup)
		router.DELETE("/security-groups/:id", handler.DeleteSecurityGroup)

		// Security Group Rule management
		// Path: /api/v1/{provider}/network/security-groups/:id/rules
		router.POST("/security-groups/:id/rules", handler.AddSecurityGroupRule)
		router.DELETE("/security-groups/:id/rules", handler.RemoveSecurityGroupRule)
		router.PUT("/security-groups/:id/rules", handler.UpdateSecurityGroupRules)
	}
}
