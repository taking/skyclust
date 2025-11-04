package network

import (
	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupGCPRoutes sets up GCP network resource routes
func SetupGCPRoutes(router *gin.RouterGroup, networkService *networkservice.Service, credentialService domain.CredentialService, logger *zap.Logger) {
	handler := NewGCPHandler(networkService, credentialService, logger)

	// GCP VPC management
	// Path: /api/v1/gcp/network/vpcs
	router.GET("/vpcs", handler.ListGCPVPCs)
	router.POST("/vpcs", handler.CreateGCPVPC)
	router.GET("/vpcs/:id", handler.GetGCPVPC)
	router.PUT("/vpcs/:id", handler.UpdateGCPVPC)
	router.DELETE("/vpcs/:id", handler.DeleteGCPVPC)

	// GCP Subnet management
	// Path: /api/v1/gcp/network/subnets
	router.GET("/subnets", handler.ListGCPSubnets)
	router.POST("/subnets", handler.CreateGCPSubnet)
	router.GET("/subnets/:id", handler.GetGCPSubnet)
	router.PUT("/subnets/:id", handler.UpdateGCPSubnet)
	router.DELETE("/subnets/:id", handler.DeleteGCPSubnet)

	// GCP Firewall Rules management
	// Path: /api/v1/gcp/network/firewall-rules
	// Note: GCP uses "Firewall Rules" terminology, not "Security Groups"
	router.GET("/firewall-rules", handler.ListGCPSecurityGroups)
	router.POST("/firewall-rules", handler.CreateGCPSecurityGroup)
	router.GET("/firewall-rules/:id", handler.GetGCPSecurityGroup)
	router.PUT("/firewall-rules/:id", handler.UpdateGCPSecurityGroup)
	router.DELETE("/firewall-rules/:id", handler.DeleteGCPSecurityGroup)

	// GCP Firewall Rule ports management (Individual port addition/removal)
	// Path: /api/v1/gcp/network/firewall-rules/:id/ports
	router.POST("/firewall-rules/:id/ports", handler.AddGCPFirewallRule)
	router.DELETE("/firewall-rules/:id/ports", handler.RemoveGCPFirewallRule)
}
