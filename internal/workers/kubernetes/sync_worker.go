package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/cache"
)

// SyncWorker periodically synchronizes Kubernetes resources with CSP APIs
type SyncWorker struct {
	k8sService        *kubernetesservice.Service
	credentialService domain.CredentialService
	credentialRepo    domain.CredentialRepository
	workspaceRepo     domain.WorkspaceRepository
	cache             cache.Cache
	eventPublisher    *messaging.Publisher
	logger            *zap.Logger

	// Worker configuration
	syncInterval   time.Duration
	maxConcurrency int
	running        bool
	mu             sync.RWMutex
	stopCh         chan struct{}
}

// SyncWorkerConfig holds configuration for the sync worker
type SyncWorkerConfig struct {
	SyncInterval   time.Duration
	MaxConcurrency int
}

// NewSyncWorker creates a new Kubernetes sync worker
func NewSyncWorker(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
	credentialRepo domain.CredentialRepository,
	workspaceRepo domain.WorkspaceRepository,
	cacheService cache.Cache,
	eventBus messaging.Bus,
	logger *zap.Logger,
	config SyncWorkerConfig,
) *SyncWorker {
	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute
	}
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 5
	}

	return &SyncWorker{
		k8sService:        k8sService,
		credentialService: credentialService,
		credentialRepo:    credentialRepo,
		workspaceRepo:     workspaceRepo,
		cache:             cacheService,
		eventPublisher:    messaging.NewPublisher(eventBus, logger),
		logger:            logger,
		syncInterval:      config.SyncInterval,
		maxConcurrency:    config.MaxConcurrency,
		stopCh:            make(chan struct{}),
	}
}

// Start starts the sync worker
func (w *SyncWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("sync worker is already running")
	}
	w.running = true
	w.mu.Unlock()

	w.logger.Info("Starting Kubernetes sync worker",
		zap.Duration("sync_interval", w.syncInterval),
		zap.Int("max_concurrency", w.maxConcurrency))

	go w.syncLoop(ctx)

	return nil
}

// Stop stops the sync worker
func (w *SyncWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	close(w.stopCh)

	w.logger.Info("Stopped Kubernetes sync worker")
}

// syncLoop runs the main synchronization loop
func (w *SyncWorker) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(w.syncInterval)
	defer ticker.Stop()

	// Initial sync
	w.syncAllCredentials(ctx)

	for {
		select {
		case <-ticker.C:
			w.syncAllCredentials(ctx)
		case <-w.stopCh:
			return
		}
	}
}

// syncAllCredentials synchronizes clusters for all active credentials
func (w *SyncWorker) syncAllCredentials(ctx context.Context) {
	w.logger.Debug("Starting sync for all credentials")

	// Get all workspaces (using List with a large limit)
	workspaces, err := w.workspaceRepo.List(ctx, 1000, 0)
	if err != nil {
		w.logger.Error("Failed to get workspaces for sync",
			zap.Error(err))
		return
	}

	// Collect all active credentials from all workspaces
	var k8sCredentials []*domain.Credential
	for _, workspace := range workspaces {
		workspaceID, err := uuid.Parse(workspace.ID)
		if err != nil {
			w.logger.Warn("Failed to parse workspace ID",
				zap.String("workspace_id", workspace.ID),
				zap.Error(err))
			continue
		}

		credentials, err := w.credentialRepo.GetByWorkspaceID(workspaceID)
		if err != nil {
			w.logger.Warn("Failed to get credentials for workspace",
				zap.String("workspace_id", workspace.ID),
				zap.Error(err))
			continue
		}

		// Filter for active Kubernetes provider credentials
		for _, cred := range credentials {
			if cred.IsActive && (cred.Provider == "aws" || cred.Provider == "gcp" || cred.Provider == "azure" || cred.Provider == "ncp") {
				k8sCredentials = append(k8sCredentials, cred)
			}
		}
	}

	if len(k8sCredentials) == 0 {
		w.logger.Debug("No Kubernetes credentials found for sync")
		return
	}

	// Use semaphore to limit concurrent syncs
	semaphore := make(chan struct{}, w.maxConcurrency)
	var wg sync.WaitGroup

	for _, cred := range k8sCredentials {
		wg.Add(1)
		go func(credential *domain.Credential) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			w.syncCredential(ctx, credential)
		}(cred)
	}

	wg.Wait()
	w.logger.Debug("Completed sync for all credentials")
}

// syncCredential synchronizes clusters for a specific credential
func (w *SyncWorker) syncCredential(ctx context.Context, credential *domain.Credential) {
	credentialID := credential.ID.String()
	w.logger.Debug("Syncing Kubernetes clusters for credential",
		zap.String("provider", credential.Provider),
		zap.String("credential_id", credentialID))

	// Get regions for this credential/provider
	regions := w.getRegionsForProvider(credential.Provider)

	for _, region := range regions {
		// Get cached clusters
		keyBuilder := cache.NewCacheKeyBuilder()
		cacheKey := keyBuilder.BuildKubernetesClusterListKey(credential.Provider, credentialID, region)

		var cachedClusters kubernetesservice.ListClustersResponse
		hasCache := false
		if err := w.cache.Get(ctx, cacheKey, &cachedClusters); err == nil {
			hasCache = true
		}

		// Fetch current clusters from CSP API
		currentClusters, err := w.k8sService.ListEKSClusters(ctx, credential, region)
		if err != nil {
			w.logger.Warn("Failed to list clusters from CSP API",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.Error(err))
			continue
		}

		// Compare with cached data and detect changes
		if hasCache {
			w.detectChanges(ctx, credential, region, &cachedClusters, currentClusters)
		}

		// Update cache
		if w.cache != nil {
			ttl := cache.GetDefaultTTL(cache.ResourceKubernetes)
			if err := w.cache.Set(ctx, cacheKey, currentClusters, ttl); err != nil {
				w.logger.Warn("Failed to update cache after sync",
					zap.String("provider", credential.Provider),
					zap.String("credential_id", credentialID),
					zap.String("region", region),
					zap.Error(err))
			}
		}

		// Publish list event for SSE subscribers
		clusterData := map[string]interface{}{
			"clusters": currentClusters.Clusters,
		}
		if err := w.eventPublisher.PublishKubernetesClusterEvent(ctx, credential.Provider, credentialID, region, "list", clusterData); err != nil {
			w.logger.Warn("Failed to publish cluster list event",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.Error(err))
		}
	}
}

// detectChanges detects changes between cached and current clusters
func (w *SyncWorker) detectChanges(ctx context.Context, credential *domain.Credential, region string, cached, current *kubernetesservice.ListClustersResponse) {
	credentialID := credential.ID.String()

	// Create maps for efficient lookup
	cachedMap := make(map[string]kubernetesservice.ClusterInfo)
	for _, cluster := range cached.Clusters {
		cachedMap[cluster.ID] = cluster
	}

	currentMap := make(map[string]kubernetesservice.ClusterInfo)
	for _, cluster := range current.Clusters {
		currentMap[cluster.ID] = cluster
	}

	// Detect new clusters
	for id, cluster := range currentMap {
		if _, exists := cachedMap[id]; !exists {
			w.logger.Info("Detected new Kubernetes cluster",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.String("cluster_id", id),
				zap.String("cluster_name", cluster.Name))

			clusterData := map[string]interface{}{
				"cluster_id": cluster.ID,
				"name":       cluster.Name,
				"version":    cluster.Version,
				"status":     cluster.Status,
				"region":     cluster.Region,
			}
			if err := w.eventPublisher.PublishKubernetesClusterEvent(ctx, credential.Provider, credentialID, region, "created", clusterData); err != nil {
				w.logger.Warn("Failed to publish cluster created event",
					zap.String("provider", credential.Provider),
					zap.String("cluster_id", id),
					zap.Error(err))
			}
		}
	}

	// Detect deleted clusters
	for id, cluster := range cachedMap {
		if _, exists := currentMap[id]; !exists {
			w.logger.Info("Detected deleted Kubernetes cluster",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.String("cluster_id", id),
				zap.String("cluster_name", cluster.Name))

			clusterData := map[string]interface{}{
				"cluster_id": cluster.ID,
				"name":       cluster.Name,
			}
			if err := w.eventPublisher.PublishKubernetesClusterEvent(ctx, credential.Provider, credentialID, region, "deleted", clusterData); err != nil {
				w.logger.Warn("Failed to publish cluster deleted event",
					zap.String("provider", credential.Provider),
					zap.String("cluster_id", id),
					zap.Error(err))
			}
		}
	}

	// Detect updated clusters (status, version changes)
	for id, currentCluster := range currentMap {
		if cachedCluster, exists := cachedMap[id]; exists {
			// Check for status changes
			if currentCluster.Status != cachedCluster.Status {
				w.logger.Info("Detected Kubernetes cluster status change",
					zap.String("provider", credential.Provider),
					zap.String("credential_id", credentialID),
					zap.String("region", region),
					zap.String("cluster_id", id),
					zap.String("old_status", cachedCluster.Status),
					zap.String("new_status", currentCluster.Status))

				clusterData := map[string]interface{}{
					"cluster_id": currentCluster.ID,
					"name":       currentCluster.Name,
					"old_status": cachedCluster.Status,
					"new_status": currentCluster.Status,
					"version":    currentCluster.Version,
					"region":     currentCluster.Region,
				}
				if err := w.eventPublisher.PublishKubernetesClusterEvent(ctx, credential.Provider, credentialID, region, "updated", clusterData); err != nil {
					w.logger.Warn("Failed to publish cluster updated event",
						zap.String("provider", credential.Provider),
						zap.String("cluster_id", id),
						zap.Error(err))
				}
			}
		}
	}
}

// getRegionsForProvider returns default regions for a provider
func (w *SyncWorker) getRegionsForProvider(provider string) []string {
	switch provider {
	case "aws":
		return []string{"ap-northeast-2", "us-east-1", "us-west-2", "eu-west-1"}
	case "gcp":
		return []string{"asia-northeast3", "asia-northeast1", "us-central1", "europe-west1"}
	case "azure":
		return []string{"koreacentral", "eastus", "westus", "westeurope"}
	case "ncp":
		return []string{"KR"}
	default:
		return []string{}
	}
}
