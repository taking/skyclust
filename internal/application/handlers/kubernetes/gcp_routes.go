package kubernetes

import (
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupGCPRoutes sets up GCP GKE-specific routes
func SetupGCPRoutes(router *gin.RouterGroup, k8sService *kubernetesservice.Service, credentialService domain.CredentialService) {
	handler := NewGCPHandler(k8sService, credentialService)

	// GKE Cluster management
	// Path: /api/v1/gcp/kubernetes/clusters
	router.POST("/clusters", handler.CreateGKECluster)
	router.GET("/clusters", handler.ListGKEClusters)
	router.GET("/clusters/:name", handler.GetGKECluster)
	router.DELETE("/clusters/:name", handler.DeleteGKECluster)
	router.GET("/clusters/:name/kubeconfig", handler.GetGKEKubeconfig)

	// GKE Node Pool management
	// Path: /api/v1/gcp/kubernetes/clusters/:name/nodepools
	router.POST("/clusters/:name/nodepools", handler.CreateGKENodePool)
	router.GET("/clusters/:name/nodepools", handler.ListGKENodePools)
	router.GET("/clusters/:name/nodepools/:nodepool", handler.GetGKENodePool)
	router.DELETE("/clusters/:name/nodepools/:nodepool", handler.DeleteGKENodePool)
	router.PUT("/clusters/:name/nodepools/:nodepool/scale", handler.ScaleGKENodePool)

	// GKE Cluster operations
	// Path: /api/v1/gcp/kubernetes/clusters/:name/upgrade
	router.POST("/clusters/:name/upgrade", handler.UpgradeGKECluster)
	router.GET("/clusters/:name/upgrade/status", handler.GetGKEUpgradeStatus)

	// GKE Node management
	// Path: /api/v1/gcp/kubernetes/clusters/:name/nodes
	router.GET("/clusters/:name/nodes", handler.ListGKENodes)
	router.GET("/clusters/:name/nodes/:node", handler.GetGKENode)
	router.POST("/clusters/:name/nodes/:node/drain", handler.DrainGKENode)
	router.POST("/clusters/:name/nodes/:node/cordon", handler.CordonGKENode)
	router.POST("/clusters/:name/nodes/:node/uncordon", handler.UncordonGKENode)

	// GKE SSH access
	// Path: /api/v1/gcp/kubernetes/clusters/:name/nodes/:node/ssh
	router.GET("/clusters/:name/nodes/:node/ssh", handler.GetGKENodeSSHConfig)
	router.POST("/clusters/:name/nodes/:node/ssh/execute", handler.ExecuteGKENodeCommand)
}
