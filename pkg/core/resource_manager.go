package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ResourceType represents different types of resources
type ResourceType string

const (
	ResourceTypeBrowser    ResourceType = "browser"
	ResourceTypeDatabase   ResourceType = "database"
	ResourceTypeHTTPClient ResourceType = "http_client"
	ResourceTypeMobile     ResourceType = "mobile"
	ResourceTypeFile       ResourceType = "file"
	ResourceTypeMemory     ResourceType = "memory"
)

// Resource represents a managed resource
type Resource interface {
	GetID() string
	GetType() ResourceType
	GetCreatedAt() time.Time
	GetLastUsed() time.Time
	IsActive() bool
	Cleanup() error
}

// ResourceInfo holds information about a resource
type ResourceInfo struct {
	ID        string                 `json:"id"`
	Type      ResourceType           `json:"type"`
	CreatedAt time.Time              `json:"created_at"`
	LastUsed  time.Time              `json:"last_used"`
	Active    bool                   `json:"active"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceManager manages the lifecycle of test resources
type ResourceManager struct {
	resources     map[string]Resource
	resourceStats map[ResourceType]*ResourceTypeStats
	mutex         sync.RWMutex
	config        *ResourceManagerConfig
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
}

// ResourceManagerConfig holds configuration for resource management
type ResourceManagerConfig struct {
	MaxResources          int           `json:"max_resources"`
	CleanupInterval       time.Duration `json:"cleanup_interval"`
	ResourceTimeout       time.Duration `json:"resource_timeout"`
	MaxIdleTime           time.Duration `json:"max_idle_time"`
	EnableLeakDetection   bool          `json:"enable_leak_detection"`
	LeakDetectionInterval time.Duration `json:"leak_detection_interval"`
}

// ResourceTypeStats holds statistics for a resource type
type ResourceTypeStats struct {
	Type         ResourceType `json:"type"`
	Active       int          `json:"active"`
	Total        int          `json:"total"`
	Created      int          `json:"created"`
	Cleaned      int          `json:"cleaned"`
	Leaked       int          `json:"leaked"`
	LastActivity time.Time    `json:"last_activity"`
}

// DefaultResourceManagerConfig returns default configuration
func DefaultResourceManagerConfig() *ResourceManagerConfig {
	return &ResourceManagerConfig{
		MaxResources:          1000,
		CleanupInterval:       30 * time.Second,
		ResourceTimeout:       5 * time.Minute,
		MaxIdleTime:           10 * time.Minute,
		EnableLeakDetection:   true,
		LeakDetectionInterval: 1 * time.Minute,
	}
}

// NewResourceManager creates a new resource manager
func NewResourceManager(config *ResourceManagerConfig) *ResourceManager {
	if config == nil {
		config = DefaultResourceManagerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	rm := &ResourceManager{
		resources:     make(map[string]Resource),
		resourceStats: make(map[ResourceType]*ResourceTypeStats),
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Initialize stats for all resource types
	for _, resourceType := range []ResourceType{
		ResourceTypeBrowser, ResourceTypeDatabase, ResourceTypeHTTPClient,
		ResourceTypeMobile, ResourceTypeFile, ResourceTypeMemory,
	} {
		rm.resourceStats[resourceType] = &ResourceTypeStats{
			Type: resourceType,
		}
	}

	// Start cleanup routine
	rm.startCleanupRoutine()

	return rm
}

// RegisterResource registers a new resource
func (rm *ResourceManager) RegisterResource(resource Resource) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if len(rm.resources) >= rm.config.MaxResources {
		return NewGowrightError(ConfigurationError, "maximum number of resources exceeded", nil)
	}

	id := resource.GetID()
	if _, exists := rm.resources[id]; exists {
		return NewGowrightError(ConfigurationError, fmt.Sprintf("resource with ID %s already exists", id), nil)
	}

	rm.resources[id] = resource

	// Update stats
	resourceType := resource.GetType()
	stats := rm.resourceStats[resourceType]
	stats.Active++
	stats.Total++
	stats.Created++
	stats.LastActivity = time.Now()

	return nil
}

// UnregisterResource unregisters a resource
func (rm *ResourceManager) UnregisterResource(resourceID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	resource, exists := rm.resources[resourceID]
	if !exists {
		return NewGowrightError(ConfigurationError, fmt.Sprintf("resource with ID %s not found", resourceID), nil)
	}

	// Cleanup the resource
	if err := resource.Cleanup(); err != nil {
		return NewGowrightError(ConfigurationError, "failed to cleanup resource", err)
	}

	delete(rm.resources, resourceID)

	// Update stats
	resourceType := resource.GetType()
	stats := rm.resourceStats[resourceType]
	stats.Active--
	stats.Cleaned++
	stats.LastActivity = time.Now()

	return nil
}

// GetResource retrieves a resource by ID
func (rm *ResourceManager) GetResource(resourceID string) (Resource, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	resource, exists := rm.resources[resourceID]
	if !exists {
		return nil, NewGowrightError(ConfigurationError, fmt.Sprintf("resource with ID %s not found", resourceID), nil)
	}

	return resource, nil
}

// GetResourcesByType retrieves all resources of a specific type
func (rm *ResourceManager) GetResourcesByType(resourceType ResourceType) []Resource {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var resources []Resource
	for _, resource := range rm.resources {
		if resource.GetType() == resourceType {
			resources = append(resources, resource)
		}
	}

	return resources
}

// GetStats returns resource statistics
func (rm *ResourceManager) GetStats() map[ResourceType]*ResourceTypeStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// Return a copy to avoid race conditions
	stats := make(map[ResourceType]*ResourceTypeStats)
	for resourceType, stat := range rm.resourceStats {
		stats[resourceType] = &ResourceTypeStats{
			Type:         stat.Type,
			Active:       stat.Active,
			Total:        stat.Total,
			Created:      stat.Created,
			Cleaned:      stat.Cleaned,
			Leaked:       stat.Leaked,
			LastActivity: stat.LastActivity,
		}
	}

	return stats
}

// CleanupAll cleans up all resources
func (rm *ResourceManager) CleanupAll() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	var errors []error

	for id, resource := range rm.resources {
		if err := resource.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup resource %s: %w", id, err))
		}
	}

	// Clear all resources
	rm.resources = make(map[string]Resource)

	// Reset stats
	for _, stats := range rm.resourceStats {
		stats.Active = 0
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// Shutdown shuts down the resource manager
func (rm *ResourceManager) Shutdown() error {
	// Stop cleanup routine
	if rm.cleanupTicker != nil {
		rm.cleanupTicker.Stop()
	}

	// Cancel context
	if rm.cancel != nil {
		rm.cancel()
	}

	// Cleanup all resources
	return rm.CleanupAll()
}

// startCleanupRoutine starts the periodic cleanup routine
func (rm *ResourceManager) startCleanupRoutine() {
	rm.cleanupTicker = time.NewTicker(rm.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-rm.cleanupTicker.C:
				// Perform cleanup operations
			case <-rm.ctx.Done():
				return
			}
		}
	}()
}
