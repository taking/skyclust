package kubernetes

import (
	kubernetesh "skyclust/internal/application/handlers/kubernetes/providers"
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up Kubernetes routes for a specific provider using Factory pattern
// provider: "aws", "gcp", "azure", "ncp"
func SetupRoutes(router *gin.RouterGroup, k8sService *kubernetesservice.Service, credentialService domain.CredentialService, provider string) {
	factory := kubernetesh.NewFactory(k8sService, credentialService)
	handler, err := factory.GetHandler(provider)
	if err != nil {
		return
	}

	// Cluster management
	// Path: /api/v1/{provider}/kubernetes/clusters
	router.POST("/clusters", handler.CreateCluster)
	router.GET("/clusters", handler.ListClusters)
	router.GET("/clusters/:name", handler.GetCluster)
	router.DELETE("/clusters/:name", handler.DeleteCluster)
	router.GET("/clusters/:name/kubeconfig", handler.GetKubeconfig)

	// Node pool management
	// Path: /api/v1/{provider}/kubernetes/clusters/:name/node-pools
	router.POST("/clusters/:name/node-pools", handler.CreateNodePool)
	router.GET("/clusters/:name/node-pools", handler.ListNodePools)
	router.GET("/clusters/:name/node-pools/:nodepool", handler.GetNodePool)
	router.DELETE("/clusters/:name/node-pools/:nodepool", handler.DeleteNodePool)
	router.PUT("/clusters/:name/node-pools/:nodepool/scale", handler.ScaleNodePool)

	// Node group management (EKS specific)
	// Path: /api/v1/{provider}/kubernetes/clusters/:name/node-groups
	router.POST("/clusters/:name/node-groups", handler.CreateNodeGroup)
	router.GET("/clusters/:name/node-groups", handler.ListNodeGroups)
	router.GET("/clusters/:name/node-groups/:nodegroup", handler.GetNodeGroup)
	router.PUT("/clusters/:name/node-groups/:nodegroup", handler.UpdateNodeGroup)
	router.DELETE("/clusters/:name/node-groups/:nodegroup", handler.DeleteNodeGroup)

	// Cluster operations
	// Path: /api/v1/{provider}/kubernetes/clusters/:name/upgrade
	router.POST("/clusters/:name/upgrade", handler.UpgradeCluster)
	router.GET("/clusters/:name/upgrade/status", handler.GetUpgradeStatus)

	// Node management
	// Path: /api/v1/{provider}/kubernetes/clusters/:name/nodes
	router.GET("/clusters/:name/nodes", handler.ListNodes)
	router.GET("/clusters/:name/nodes/:node", handler.GetNode)
	router.POST("/clusters/:name/nodes/:node/drain", handler.DrainNode)
	router.POST("/clusters/:name/nodes/:node/cordon", handler.CordonNode)
	router.POST("/clusters/:name/nodes/:node/uncordon", handler.UncordonNode)

	// SSH access
	// Path: /api/v1/{provider}/kubernetes/clusters/:name/nodes/:node/ssh
	router.GET("/clusters/:name/nodes/:node/ssh", handler.GetNodeSSHConfig)
	router.POST("/clusters/:name/nodes/:node/ssh/execute", handler.ExecuteNodeCommand)

	// Metadata endpoints (AWS only)
	// Path: /api/v1/{provider}/kubernetes/metadata
	if provider == "aws" {
		router.GET("/metadata/versions", handler.GetEKSVersions)
		router.GET("/metadata/regions", handler.GetAWSRegions)
		router.GET("/metadata/availability-zones", handler.GetAvailabilityZones)
		router.GET("/metadata/instance-types", handler.GetInstanceTypes)
		router.GET("/metadata/ami-types", handler.GetEKSAmitTypes)
		router.GET("/metadata/gpu-quota", handler.CheckGPUQuota)
	}
}
