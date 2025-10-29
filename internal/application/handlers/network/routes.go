package network

import (
	service "skyclust/internal/application/services"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up network resource routes for a specific provider
// provider: "aws", "gcp", "azure", "ncp"
func SetupRoutes(router *gin.RouterGroup, networkService *service.NetworkService, credentialService domain.CredentialService, provider string) {
	handler := NewHandler(networkService, credentialService, provider)

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

	// Security Group management
	// Path: /api/v1/{provider}/network/security-groups
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
