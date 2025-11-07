package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"

	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/cache"
)

// SyncWorker periodically synchronizes Network resources with CSP APIs
type SyncWorker struct {
	networkService    *networkservice.Service
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

// NewSyncWorker creates a new Network sync worker
func NewSyncWorker(
	networkService *networkservice.Service,
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
		networkService:    networkService,
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

	w.logger.Info("Starting Network sync worker",
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

	w.logger.Info("Stopped Network sync worker")
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

// syncAllCredentials synchronizes VPCs for all active credentials
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
	var networkCredentials []*domain.Credential
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

		// Filter for active Network provider credentials
		for _, cred := range credentials {
			if cred.IsActive && (cred.Provider == "aws" || cred.Provider == "gcp" || cred.Provider == "azure" || cred.Provider == "ncp") {
				networkCredentials = append(networkCredentials, cred)
			}
		}
	}

	if len(networkCredentials) == 0 {
		w.logger.Debug("No Network credentials found for sync")
		return
	}

	// Use semaphore to limit concurrent syncs
	semaphore := make(chan struct{}, w.maxConcurrency)
	var wg sync.WaitGroup

	for _, cred := range networkCredentials {
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

// syncCredential synchronizes VPCs for a specific credential
func (w *SyncWorker) syncCredential(ctx context.Context, credential *domain.Credential) {
	credentialID := credential.ID.String()
	w.logger.Debug("Syncing Network VPCs for credential",
		zap.String("provider", credential.Provider),
		zap.String("credential_id", credentialID))

	// Get regions for this credential/provider
	regions := w.getRegionsForProvider(credential.Provider)

	for _, region := range regions {
		// Get cached VPCs
		keyBuilder := cache.NewCacheKeyBuilder()
		cacheKey := keyBuilder.BuildNetworkVPCListKey(credential.Provider, credentialID, region)

		var cachedVPCs networkservice.ListVPCsResponse
		hasCache := false
		if err := w.cache.Get(ctx, cacheKey, &cachedVPCs); err == nil {
			hasCache = true
		}

		// Fetch current VPCs from CSP API
		currentVPCs, err := w.networkService.ListVPCs(ctx, credential, networkservice.ListVPCsRequest{
			Region: region,
		})
		if err != nil {
			w.logger.Warn("Failed to list VPCs from CSP API",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.Error(err))
			continue
		}

		// Compare with cached data and detect changes
		if hasCache {
			w.detectChanges(ctx, credential, region, &cachedVPCs, currentVPCs)
		}

		// Update cache
		if w.cache != nil {
			ttl := cache.GetDefaultTTL(cache.ResourceNetwork)
			if err := w.cache.Set(ctx, cacheKey, currentVPCs, ttl); err != nil {
				w.logger.Warn("Failed to update cache after sync",
					zap.String("provider", credential.Provider),
					zap.String("credential_id", credentialID),
					zap.String("region", region),
					zap.Error(err))
			}
		}

		// Publish list event for SSE subscribers
		vpcData := map[string]interface{}{
			"vpcs": currentVPCs.VPCs,
		}
		if err := w.eventPublisher.PublishVPCEvent(ctx, credential.Provider, credentialID, region, "list", vpcData); err != nil {
			w.logger.Warn("Failed to publish VPC list event",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.Error(err))
		}
	}
}

// detectChanges detects changes between cached and current VPCs
func (w *SyncWorker) detectChanges(ctx context.Context, credential *domain.Credential, region string, cached, current *networkservice.ListVPCsResponse) {
	credentialID := credential.ID.String()

	// Create maps for efficient lookup
	cachedMap := make(map[string]networkservice.VPCInfo)
	for _, vpc := range cached.VPCs {
		cachedMap[vpc.ID] = vpc
	}

	currentMap := make(map[string]networkservice.VPCInfo)
	for _, vpc := range current.VPCs {
		currentMap[vpc.ID] = vpc
	}

	// Detect new VPCs
	for id, vpc := range currentMap {
		if _, exists := cachedMap[id]; !exists {
			w.logger.Info("Detected new Network VPC",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.String("vpc_id", id),
				zap.String("vpc_name", vpc.Name))

			vpcData := map[string]interface{}{
				"vpc_id": vpc.ID,
				"name":   vpc.Name,
				"state":  vpc.State,
				"region": vpc.Region,
			}
			if err := w.eventPublisher.PublishVPCEvent(ctx, credential.Provider, credentialID, region, "created", vpcData); err != nil {
				w.logger.Warn("Failed to publish VPC created event",
					zap.String("provider", credential.Provider),
					zap.String("vpc_id", id),
					zap.Error(err))
			}
		}
	}

	// Detect deleted VPCs
	for id, vpc := range cachedMap {
		if _, exists := currentMap[id]; !exists {
			w.logger.Info("Detected deleted Network VPC",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.String("vpc_id", id),
				zap.String("vpc_name", vpc.Name))

			vpcData := map[string]interface{}{
				"vpc_id": vpc.ID,
				"name":   vpc.Name,
			}
			if err := w.eventPublisher.PublishVPCEvent(ctx, credential.Provider, credentialID, region, "deleted", vpcData); err != nil {
				w.logger.Warn("Failed to publish VPC deleted event",
					zap.String("provider", credential.Provider),
					zap.String("vpc_id", id),
					zap.Error(err))
			}
		}
	}

	// Detect updated VPCs (state changes)
	for id, currentVPC := range currentMap {
		if cachedVPC, exists := cachedMap[id]; exists {
			// Check for state changes
			if currentVPC.State != cachedVPC.State {
				w.logger.Info("Detected Network VPC state change",
					zap.String("provider", credential.Provider),
					zap.String("credential_id", credentialID),
					zap.String("region", region),
					zap.String("vpc_id", id),
					zap.String("old_state", cachedVPC.State),
					zap.String("new_state", currentVPC.State))

				vpcData := map[string]interface{}{
					"vpc_id":    currentVPC.ID,
					"name":      currentVPC.Name,
					"old_state": cachedVPC.State,
					"new_state": currentVPC.State,
					"region":    currentVPC.Region,
				}
				if err := w.eventPublisher.PublishVPCEvent(ctx, credential.Provider, credentialID, region, "updated", vpcData); err != nil {
					w.logger.Warn("Failed to publish VPC updated event",
						zap.String("provider", credential.Provider),
						zap.String("vpc_id", id),
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
