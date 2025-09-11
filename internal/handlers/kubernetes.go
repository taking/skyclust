package handlers

import (
	"net/http"
	"strconv"

	"cmp/internal/services"
	"cmp/pkg/interfaces"

	"github.com/gin-gonic/gin"
)

// Kubernetes handlers

// ListClusters lists all clusters in a workspace
func ListClusters(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")

		clusters, err := services.Kubernetes.ListClusters(c.Request.Context(), workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"clusters": clusters})
	}
}

// CreateCluster creates a new cluster
func CreateCluster(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")

		var req interfaces.CreateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cluster, err := services.Kubernetes.CreateCluster(c.Request.Context(), workspaceID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, cluster)
	}
}

// GetCluster gets a specific cluster
func GetCluster(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		cluster, err := services.Kubernetes.GetCluster(c.Request.Context(), workspaceID, clusterID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, cluster)
	}
}

// DeleteCluster deletes a cluster
func DeleteCluster(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		err := services.Kubernetes.DeleteCluster(c.Request.Context(), workspaceID, clusterID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cluster deleted successfully"})
	}
}

// ListDeployments lists all deployments in a cluster namespace
func ListDeployments(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespace")

		deployments, err := services.Kubernetes.ListDeployments(c.Request.Context(), workspaceID, clusterID, namespace)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"deployments": deployments})
	}
}

// CreateDeployment creates a new deployment
func CreateDeployment(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		var req interfaces.CreateDeploymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		deployment, err := services.Kubernetes.CreateDeployment(c.Request.Context(), workspaceID, clusterID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, deployment)
	}
}

// GetDeployment gets a specific deployment
func GetDeployment(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		deployment, err := services.Kubernetes.GetDeployment(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, deployment)
	}
}

// UpdateDeployment updates a deployment
func UpdateDeployment(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		var req interfaces.UpdateDeploymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		deployment, err := services.Kubernetes.UpdateDeployment(c.Request.Context(), workspaceID, clusterID, namespace, name, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, deployment)
	}
}

// DeleteDeployment deletes a deployment
func DeleteDeployment(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		err := services.Kubernetes.DeleteDeployment(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deployment deleted successfully"})
	}
}

// ScaleDeployment scales a deployment
func ScaleDeployment(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		replicasStr := c.Param("replicas")
		replicas, err := strconv.ParseInt(replicasStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid replicas parameter"})
			return
		}

		err = services.Kubernetes.ScaleDeployment(c.Request.Context(), workspaceID, clusterID, namespace, name, int32(replicas))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deployment scaled successfully"})
	}
}

// ListServices lists all services in a cluster namespace
func ListServices(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespace")

		services, err := services.Kubernetes.ListServices(c.Request.Context(), workspaceID, clusterID, namespace)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"services": services})
	}
}

// CreateService creates a new service
func CreateService(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		var req interfaces.CreateServiceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		service, err := services.Kubernetes.CreateService(c.Request.Context(), workspaceID, clusterID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, service)
	}
}

// GetService gets a specific service
func GetService(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		service, err := services.Kubernetes.GetService(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, service)
	}
}

// UpdateService updates a service
func UpdateService(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		var req interfaces.UpdateServiceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		service, err := services.Kubernetes.UpdateService(c.Request.Context(), workspaceID, clusterID, namespace, name, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, service)
	}
}

// DeleteService deletes a service
func DeleteService(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		err := services.Kubernetes.DeleteService(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
	}
}

// ListNamespaces lists all namespaces in a cluster
func ListNamespaces(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		namespaces, err := services.Kubernetes.ListNamespaces(c.Request.Context(), workspaceID, clusterID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
	}
}

// CreateNamespace creates a new namespace
func CreateNamespace(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		var req interfaces.CreateNamespaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		namespace, err := services.Kubernetes.CreateNamespace(c.Request.Context(), workspaceID, clusterID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, namespace)
	}
}

// GetNamespace gets a specific namespace
func GetNamespace(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		name := c.Param("namespaceName")

		namespace, err := services.Kubernetes.GetNamespace(c.Request.Context(), workspaceID, clusterID, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, namespace)
	}
}

// DeleteNamespace deletes a namespace
func DeleteNamespace(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		name := c.Param("namespaceName")

		err := services.Kubernetes.DeleteNamespace(c.Request.Context(), workspaceID, clusterID, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Namespace deleted successfully"})
	}
}

// ListPods lists all pods in a cluster namespace
func ListPods(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespace")

		pods, err := services.Kubernetes.ListPods(c.Request.Context(), workspaceID, clusterID, namespace)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"pods": pods})
	}
}

// GetPod gets a specific pod
func GetPod(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		pod, err := services.Kubernetes.GetPod(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, pod)
	}
}

// DeletePod deletes a pod
func DeletePod(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		err := services.Kubernetes.DeletePod(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Pod deleted successfully"})
	}
}

// ListConfigMaps lists all configmaps in a cluster namespace
func ListConfigMaps(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespace")

		configMaps, err := services.Kubernetes.ListConfigMaps(c.Request.Context(), workspaceID, clusterID, namespace)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"configmaps": configMaps})
	}
}

// CreateConfigMap creates a new configmap
func CreateConfigMap(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		var req interfaces.CreateConfigMapRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		configMap, err := services.Kubernetes.CreateConfigMap(c.Request.Context(), workspaceID, clusterID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, configMap)
	}
}

// GetConfigMap gets a specific configmap
func GetConfigMap(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		configMap, err := services.Kubernetes.GetConfigMap(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, configMap)
	}
}

// UpdateConfigMap updates a configmap
func UpdateConfigMap(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		var req interfaces.UpdateConfigMapRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		configMap, err := services.Kubernetes.UpdateConfigMap(c.Request.Context(), workspaceID, clusterID, namespace, name, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, configMap)
	}
}

// DeleteConfigMap deletes a configmap
func DeleteConfigMap(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		err := services.Kubernetes.DeleteConfigMap(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ConfigMap deleted successfully"})
	}
}

// ListSecrets lists all secrets in a cluster namespace
func ListSecrets(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespace")

		secrets, err := services.Kubernetes.ListSecrets(c.Request.Context(), workspaceID, clusterID, namespace)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"secrets": secrets})
	}
}

// CreateSecret creates a new secret
func CreateSecret(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")

		var req interfaces.CreateSecretRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		secret, err := services.Kubernetes.CreateSecret(c.Request.Context(), workspaceID, clusterID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, secret)
	}
}

// GetSecret gets a specific secret
func GetSecret(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		secret, err := services.Kubernetes.GetSecret(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, secret)
	}
}

// UpdateSecret updates a secret
func UpdateSecret(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		var req interfaces.UpdateSecretRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		secret, err := services.Kubernetes.UpdateSecret(c.Request.Context(), workspaceID, clusterID, namespace, name, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, secret)
	}
}

// DeleteSecret deletes a secret
func DeleteSecret(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		clusterID := c.Param("clusterId")
		namespace := c.Param("namespaceName")
		name := c.Param("secretName")

		err := services.Kubernetes.DeleteSecret(c.Request.Context(), workspaceID, clusterID, namespace, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Secret deleted successfully"})
	}
}
