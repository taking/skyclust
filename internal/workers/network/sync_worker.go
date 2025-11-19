package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

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
	redisClient       *redis.Client
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

	// Extract Redis client from cache service if available
	var redisClient *redis.Client
	if redisService, ok := cacheService.(*cache.RedisService); ok {
		redisClient = redisService.GetClient()
		logger.Info("Network sync worker initialized with Redis support for subscription tracking")
	} else {
		logger.Warn("Network sync worker initialized without Redis (subscription-based sync will be disabled)")
	}

	return &SyncWorker{
		networkService:    networkService,
		credentialService: credentialService,
		credentialRepo:    credentialRepo,
		workspaceRepo:     workspaceRepo,
		cache:             cacheService,
		redisClient:       redisClient,
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
		zap.Int("max_concurrency", w.maxConcurrency),
		zap.Bool("subscription_based_sync", w.redisClient != nil),
		zap.String("sync_mode", "priority-based"))

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

// syncLoop runs the main synchronization loop with subscription-based priority scheduling
func (w *SyncWorker) syncLoop(ctx context.Context) {
	// Create separate tickers for different priorities
	highPriorityTicker := time.NewTicker(1 * time.Minute)
	mediumPriorityTicker := time.NewTicker(3 * time.Minute)
	lowPriorityTicker := time.NewTicker(10 * time.Minute)
	defer highPriorityTicker.Stop()
	defer mediumPriorityTicker.Stop()
	defer lowPriorityTicker.Stop()

	// Initial sync (all credentials)
	w.syncAllCredentials(ctx)

	// Track last sync time for each priority to avoid duplicate syncs
	lastHighSync := make(map[string]time.Time)   // credential_id:region -> time
	lastMediumSync := make(map[string]time.Time) // credential_id:region -> time
	lastLowSync := make(map[string]time.Time)    // credential_id:region -> time

	for {
		select {
		case <-highPriorityTicker.C:
			w.syncWithPriority(ctx, PriorityHigh, lastHighSync)
		case <-mediumPriorityTicker.C:
			w.syncWithPriority(ctx, PriorityMedium, lastMediumSync)
		case <-lowPriorityTicker.C:
			w.syncWithPriority(ctx, PriorityLow, lastLowSync)
		case <-w.stopCh:
			return
		}
	}
}

// syncAllCredentials synchronizes VPCs for all active credentials (fallback for initial sync)
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
	w.logger.Debug("Completed initial sync for all credentials")
}

// syncWithPriority synchronizes credentials based on subscription priority
func (w *SyncWorker) syncWithPriority(ctx context.Context, priority SubscriptionPriority, lastSyncMap map[string]time.Time) {
	startTime := time.Now()

	// Get active subscriptions
	subscriptions, err := w.getActiveSubscriptions(ctx)
	if err != nil {
		w.logger.Warn("Failed to get active subscriptions, falling back to default sync",
			zap.String("priority", priority.String()),
			zap.Error(err))
		// Fallback to default sync if subscription tracking fails
		if priority == PriorityLow {
			w.syncAllCredentials(ctx)
		}
		return
	}

	// Log subscription statistics
	totalSubscriptions := 0
	for _, regions := range subscriptions {
		for _, count := range regions {
			totalSubscriptions += int(count)
		}
	}
	w.logger.Debug("Subscription statistics",
		zap.String("priority", priority.String()),
		zap.Int("total_subscribed_credentials", len(subscriptions)),
		zap.Int("total_active_subscriptions", totalSubscriptions))

	// Get all active credentials
	workspaces, err := w.workspaceRepo.List(ctx, 1000, 0)
	if err != nil {
		w.logger.Error("Failed to get workspaces for priority sync",
			zap.Error(err))
		return
	}

	var credentialsToSync []struct {
		credential *domain.Credential
		regions    []string
	}

	for _, workspace := range workspaces {
		workspaceID, err := uuid.Parse(workspace.ID)
		if err != nil {
			continue
		}

		credentials, err := w.credentialRepo.GetByWorkspaceID(workspaceID)
		if err != nil {
			continue
		}

		for _, cred := range credentials {
			if !cred.IsActive || (cred.Provider != "aws" && cred.Provider != "gcp" && cred.Provider != "azure" && cred.Provider != "ncp") {
				continue
			}

			credentialID := cred.ID.String()
			subscribedRegions := subscriptions[credentialID]

			// Determine which regions to sync based on priority
			var regionsToSync []string

			if priority == PriorityLow {
				// Low priority: sync all regions (including those without subscriptions)
				// Use default regions as fallback
				regionsToSync = w.getRegionsForProvider(cred.Provider)
			} else {
				// High/Medium priority: only sync subscribed regions
				for region, count := range subscribedRegions {
					subPriority := w.getSubscriptionPriority(count)
					if subPriority == priority {
						regionsToSync = append(regionsToSync, region)
					}
				}
			}

			if len(regionsToSync) > 0 {
				credentialsToSync = append(credentialsToSync, struct {
					credential *domain.Credential
					regions    []string
				}{
					credential: cred,
					regions:    regionsToSync,
				})
			}
		}
	}

	if len(credentialsToSync) == 0 {
		return
	}

	// Use semaphore to limit concurrent syncs
	semaphore := make(chan struct{}, w.maxConcurrency)
	var wg sync.WaitGroup

	for _, item := range credentialsToSync {
		wg.Add(1)
		go func(credential *domain.Credential, regions []string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			credentialID := credential.ID.String()
			for _, region := range regions {
				// Check if we should skip this sync (avoid duplicate syncs)
				syncKey := fmt.Sprintf("%s:%s", credentialID, region)
				lastSync, exists := lastSyncMap[syncKey]
				interval := w.getIntervalForPriority(priority)
				if exists && time.Since(lastSync) < interval {
					continue
				}

				w.syncCredentialRegion(ctx, credential, region)
				lastSyncMap[syncKey] = time.Now()
			}
		}(item.credential, item.regions)
	}

	wg.Wait()

	// Calculate statistics
	totalRegions := 0
	for _, item := range credentialsToSync {
		totalRegions += len(item.regions)
	}

	duration := time.Since(startTime)

	w.logger.Info("Completed priority sync",
		zap.String("priority", priority.String()),
		zap.Int("credentials_synced", len(credentialsToSync)),
		zap.Int("total_regions_synced", totalRegions),
		zap.Duration("sync_interval", w.getIntervalForPriority(priority)),
		zap.Duration("sync_duration", duration),
		zap.Int("total_active_subscriptions", totalSubscriptions))
}

// syncCredential synchronizes VPCs for a specific credential (uses default regions)
func (w *SyncWorker) syncCredential(ctx context.Context, credential *domain.Credential) {
	credentialID := credential.ID.String()
	w.logger.Debug("Syncing Network VPCs for credential",
		zap.String("provider", credential.Provider),
		zap.String("credential_id", credentialID))

	// Get regions for this credential/provider (default regions)
	regions := w.getRegionsForProvider(credential.Provider)

	for _, region := range regions {
		w.syncCredentialRegion(ctx, credential, region)
	}
}

// syncCredentialRegion synchronizes VPCs for a specific credential and region
func (w *SyncWorker) syncCredentialRegion(ctx context.Context, credential *domain.Credential, region string) {
	credentialID := credential.ID.String()

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
		return
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
