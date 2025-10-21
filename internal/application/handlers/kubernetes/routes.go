package kubernetes

import (
	service "skyclust/internal/application/services"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up Kubernetes routes for a specific provider
// provider: "aws", "gcp", "azure", "ncp"
func SetupRoutes(router *gin.RouterGroup, k8sService *service.KubernetesService, credentialService domain.CredentialService, provider string) {
	handler := NewHandler(k8sService, credentialService, provider)

	// Cluster management
	// Path: /api/v1/{provider}/kubernetes/clusters
	router.POST("/clusters", handler.CreateCluster)
	router.GET("/clusters", handler.ListClusters)
	router.GET("/clusters/:name", handler.GetCluster)
	router.DELETE("/clusters/:name", handler.DeleteCluster)
	router.GET("/clusters/:name/kubeconfig", handler.GetKubeconfig)

	// Node pool management
	// Path: /api/v1/{provider}/kubernetes/clusters/:name/nodepools
	router.POST("/clusters/:name/nodepools", handler.CreateNodePool)
	router.GET("/clusters/:name/nodepools", handler.ListNodePools)
	router.GET("/clusters/:name/nodepools/:nodepool", handler.GetNodePool)
	router.DELETE("/clusters/:name/nodepools/:nodepool", handler.DeleteNodePool)
	router.PUT("/clusters/:name/nodepools/:nodepool/scale", handler.ScaleNodePool)

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
}
