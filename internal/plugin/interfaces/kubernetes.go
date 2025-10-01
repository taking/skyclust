package interfaces

import (
	"context"
	"time"
)

// KubernetesProvider defines the Kubernetes management interface
type KubernetesProvider interface {
	// Cluster management
	CreateCluster(ctx context.Context, req CreateClusterRequest) (*Cluster, error)
	GetCluster(ctx context.Context, clusterID string) (*Cluster, error)
	ListClusters(ctx context.Context) ([]Cluster, error)
	DeleteCluster(ctx context.Context, clusterID string) error

	// Workload management
	CreateDeployment(ctx context.Context, req CreateDeploymentRequest) (*Deployment, error)
	GetDeployment(ctx context.Context, clusterID, namespace, name string) (*Deployment, error)
	ListDeployments(ctx context.Context, clusterID, namespace string) ([]Deployment, error)
	UpdateDeployment(ctx context.Context, clusterID, namespace, name string, req UpdateDeploymentRequest) (*Deployment, error)
	DeleteDeployment(ctx context.Context, clusterID, namespace, name string) error
	ScaleDeployment(ctx context.Context, clusterID, namespace, name string, replicas int32) error

	// Service management
	CreateService(ctx context.Context, req CreateServiceRequest) (*Service, error)
	GetService(ctx context.Context, clusterID, namespace, name string) (*Service, error)
	ListServices(ctx context.Context, clusterID, namespace string) ([]Service, error)
	UpdateService(ctx context.Context, clusterID, namespace, name string, req UpdateServiceRequest) (*Service, error)
	DeleteService(ctx context.Context, clusterID, namespace, name string) error

	// Namespace management
	CreateNamespace(ctx context.Context, req CreateNamespaceRequest) (*Namespace, error)
	GetNamespace(ctx context.Context, clusterID, name string) (*Namespace, error)
	ListNamespaces(ctx context.Context, clusterID string) ([]Namespace, error)
	DeleteNamespace(ctx context.Context, clusterID, name string) error

	// Pod management
	ListPods(ctx context.Context, clusterID, namespace string) ([]Pod, error)
	GetPod(ctx context.Context, clusterID, namespace, name string) (*Pod, error)
	DeletePod(ctx context.Context, clusterID, namespace, name string) error

	// ConfigMap and Secret management
	CreateConfigMap(ctx context.Context, req CreateConfigMapRequest) (*ConfigMap, error)
	GetConfigMap(ctx context.Context, clusterID, namespace, name string) (*ConfigMap, error)
	ListConfigMaps(ctx context.Context, clusterID, namespace string) ([]ConfigMap, error)
	UpdateConfigMap(ctx context.Context, clusterID, namespace, name string, req UpdateConfigMapRequest) (*ConfigMap, error)
	DeleteConfigMap(ctx context.Context, clusterID, namespace, name string) error

	CreateSecret(ctx context.Context, req CreateSecretRequest) (*Secret, error)
	GetSecret(ctx context.Context, clusterID, namespace, name string) (*Secret, error)
	ListSecrets(ctx context.Context, clusterID, namespace string) ([]Secret, error)
	UpdateSecret(ctx context.Context, clusterID, namespace, name string, req UpdateSecretRequest) (*Secret, error)
	DeleteSecret(ctx context.Context, clusterID, namespace, name string) error
}

// Cluster represents a Kubernetes cluster
type Cluster struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Status    string            `json:"status"`
	Endpoint  string            `json:"endpoint"`
	Region    string            `json:"region"`
	Provider  string            `json:"provider"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Deployment represents a Kubernetes deployment
type Deployment struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	Replicas      int32             `json:"replicas"`
	ReadyReplicas int32             `json:"ready_replicas"`
	Image         string            `json:"image"`
	Labels        map[string]string `json:"labels"`
	Annotations   map[string]string `json:"annotations"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// Service represents a Kubernetes service
type Service struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"cluster_ip"`
	Ports       []ServicePort     `json:"ports"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ServicePort represents a service port
type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Status      string            `json:"status"`
	Ready       bool              `json:"ready"`
	Restarts    int32             `json:"restarts"`
	Image       string            `json:"image"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
}

// ConfigMap represents a Kubernetes ConfigMap
type ConfigMap struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Data        map[string]string `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Secret represents a Kubernetes Secret
type Secret struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	Data        map[string][]byte `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Request types
type CreateClusterRequest struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Region   string            `json:"region"`
	Provider string            `json:"provider"`
	Labels   map[string]string `json:"labels"`
}

type CreateDeploymentRequest struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Replicas    int32             `json:"replicas"`
	Image       string            `json:"image"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type UpdateDeploymentRequest struct {
	Replicas    *int32            `json:"replicas,omitempty"`
	Image       string            `json:"image,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type CreateServiceRequest struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	Ports       []ServicePort     `json:"ports"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type UpdateServiceRequest struct {
	Type        string            `json:"type,omitempty"`
	Ports       []ServicePort     `json:"ports,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type CreateNamespaceRequest struct {
	Name        string            `json:"name"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type CreateConfigMapRequest struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Data        map[string]string `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type UpdateConfigMapRequest struct {
	Data        map[string]string `json:"data,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type CreateSecretRequest struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	Data        map[string][]byte `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type UpdateSecretRequest struct {
	Type        string            `json:"type,omitempty"`
	Data        map[string][]byte `json:"data,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
