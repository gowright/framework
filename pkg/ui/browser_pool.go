package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gowright/framework/pkg/core"
)

// BrowserPool manages a pool of browser instances for concurrent testing
type BrowserPool struct {
	browsers    chan *BrowserInstance
	maxSize     int
	timeout     time.Duration
	mutex       sync.RWMutex
	stats       *BrowserPoolStats
	initialized bool
}

// BrowserInstance represents a browser instance in the pool
type BrowserInstance struct {
	ID         string
	CreatedAt  time.Time
	UsageCount int
	// Browser-specific fields would go here
}

// BrowserPoolStats holds statistics about browser pool usage
type BrowserPoolStats struct {
	MaxSize       int `json:"max_size"`
	Available     int `json:"available"`
	InUse         int `json:"in_use"`
	TotalCreated  int `json:"total_created"`
	TotalAcquired int `json:"total_acquired"`
	TotalReleased int `json:"total_released"`
}

// NewBrowserPool creates a new browser pool
func NewBrowserPool(maxSize int, timeout time.Duration) (*BrowserPool, error) {
	if maxSize <= 0 {
		return nil, core.NewGowrightError(core.ConfigurationError, "browser pool max size must be positive", nil)
	}

	return &BrowserPool{
		browsers: make(chan *BrowserInstance, maxSize),
		maxSize:  maxSize,
		timeout:  timeout,
		stats: &BrowserPoolStats{
			MaxSize: maxSize,
		},
	}, nil
}

// Initialize initializes the browser pool
func (bp *BrowserPool) Initialize() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	if bp.initialized {
		return nil
	}

	// Pre-create some browser instances
	for i := 0; i < bp.maxSize/2; i++ {
		instance, err := bp.createBrowserInstance()
		if err != nil {
			return core.NewGowrightError(core.BrowserError, "failed to create browser instance", err)
		}
		bp.browsers <- instance
		bp.stats.TotalCreated++
		bp.stats.Available++
	}

	bp.initialized = true
	return nil
}

// AcquireBrowser acquires a browser instance from the pool
func (bp *BrowserPool) AcquireBrowser(ctx context.Context) (*BrowserInstance, error) {
	if !bp.initialized {
		return nil, core.NewGowrightError(core.BrowserError, "browser pool not initialized", nil)
	}

	select {
	case instance := <-bp.browsers:
		bp.mutex.Lock()
		instance.UsageCount++
		bp.stats.TotalAcquired++
		bp.stats.Available--
		bp.stats.InUse++
		bp.mutex.Unlock()
		return instance, nil
	case <-time.After(bp.timeout):
		return nil, core.NewGowrightError(core.BrowserError, "timeout acquiring browser from pool", nil)
	case <-ctx.Done():
		return nil, core.NewGowrightError(core.BrowserError, "context cancelled while acquiring browser", ctx.Err())
	}
}

// ReleaseBrowser returns a browser instance to the pool
func (bp *BrowserPool) ReleaseBrowser(instance *BrowserInstance) error {
	if instance == nil {
		return core.NewGowrightError(core.BrowserError, "cannot release nil browser instance", nil)
	}

	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	select {
	case bp.browsers <- instance:
		bp.stats.TotalReleased++
		bp.stats.Available++
		bp.stats.InUse--
		return nil
	default:
		// Pool is full, close the browser instance
		return bp.closeBrowserInstance(instance)
	}
}

// GetStats returns current pool statistics
func (bp *BrowserPool) GetStats() *BrowserPoolStats {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &BrowserPoolStats{
		MaxSize:       bp.stats.MaxSize,
		Available:     bp.stats.Available,
		InUse:         bp.stats.InUse,
		TotalCreated:  bp.stats.TotalCreated,
		TotalAcquired: bp.stats.TotalAcquired,
		TotalReleased: bp.stats.TotalReleased,
	}
}

// Cleanup closes all browser instances and cleans up the pool
func (bp *BrowserPool) Cleanup() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	if !bp.initialized {
		return nil
	}

	// Close all browser instances in the pool
	close(bp.browsers)
	for instance := range bp.browsers {
		if err := bp.closeBrowserInstance(instance); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Error closing browser instance: %v\n", err)
		}
	}

	bp.browsers = make(chan *BrowserInstance, bp.maxSize)
	bp.initialized = false
	bp.stats.Available = 0
	bp.stats.InUse = 0

	return nil
}

// createBrowserInstance creates a new browser instance
func (bp *BrowserPool) createBrowserInstance() (*BrowserInstance, error) {
	// This would create an actual browser instance using rod or other automation library
	// For now, this is a placeholder
	return &BrowserInstance{
		ID:        fmt.Sprintf("browser-%d", time.Now().UnixNano()),
		CreatedAt: time.Now(),
	}, nil
}

// closeBrowserInstance closes a browser instance
func (bp *BrowserPool) closeBrowserInstance(instance *BrowserInstance) error {
	// This would close the actual browser instance
	// For now, this is a placeholder
	return nil
}
