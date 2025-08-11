package api

import (
	"context"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gowright/framework/pkg/core"
)

// HTTPClientPool manages a pool of HTTP clients for concurrent testing
type HTTPClientPool struct {
	clients     chan *HTTPClientInstance
	maxSize     int
	timeout     time.Duration
	mutex       sync.RWMutex
	stats       *HTTPClientPoolStats
	initialized bool
}

// HTTPClientInstance represents an HTTP client instance in the pool
type HTTPClientInstance struct {
	Client     *resty.Client
	CreatedAt  time.Time
	UsageCount int
}

// HTTPClientPoolStats holds statistics about HTTP client pool usage
type HTTPClientPoolStats struct {
	MaxSize       int `json:"max_size"`
	Available     int `json:"available"`
	InUse         int `json:"in_use"`
	TotalCreated  int `json:"total_created"`
	TotalAcquired int `json:"total_acquired"`
	TotalReleased int `json:"total_released"`
}

// NewHTTPClientPool creates a new HTTP client pool
func NewHTTPClientPool(maxSize int, timeout time.Duration) (*HTTPClientPool, error) {
	if maxSize <= 0 {
		return nil, core.NewGowrightError(core.ConfigurationError, "HTTP client pool max size must be positive", nil)
	}

	pool := &HTTPClientPool{
		clients: make(chan *HTTPClientInstance, maxSize),
		maxSize: maxSize,
		timeout: timeout,
		stats: &HTTPClientPoolStats{
			MaxSize: maxSize,
		},
	}

	// Initialize the pool
	if err := pool.Initialize(); err != nil {
		return nil, err
	}

	return pool, nil
}

// Initialize initializes the HTTP client pool
func (hcp *HTTPClientPool) Initialize() error {
	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	if hcp.initialized {
		return nil
	}

	// Don't pre-create instances, create them on demand
	hcp.initialized = true
	return nil
}

// AcquireClient acquires an HTTP client instance from the pool
func (hcp *HTTPClientPool) AcquireClient(ctx context.Context) (*HTTPClientInstance, error) {
	if !hcp.initialized {
		return nil, core.NewGowrightError(core.APIError, "HTTP client pool not initialized", nil)
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, core.NewGowrightError(core.APIError, "context cancelled while acquiring HTTP client", ctx.Err())
	default:
	}

	// Try to get an available instance first
	select {
	case instance := <-hcp.clients:
		hcp.mutex.Lock()
		instance.UsageCount++
		hcp.stats.TotalAcquired++
		hcp.stats.Available--
		hcp.stats.InUse++
		hcp.mutex.Unlock()
		return instance, nil
	default:
		// No available instance, check if we can create a new one
		hcp.mutex.Lock()
		if hcp.stats.InUse+hcp.stats.Available < hcp.maxSize {
			// We can create a new instance
			instance := hcp.createClientInstance()
			hcp.stats.TotalCreated++
			hcp.stats.TotalAcquired++
			hcp.stats.InUse++
			hcp.mutex.Unlock()
			return instance, nil
		}
		hcp.mutex.Unlock()

		// Pool is at capacity, wait for an available instance
		select {
		case instance := <-hcp.clients:
			hcp.mutex.Lock()
			instance.UsageCount++
			hcp.stats.TotalAcquired++
			hcp.stats.Available--
			hcp.stats.InUse++
			hcp.mutex.Unlock()
			return instance, nil
		case <-time.After(hcp.timeout):
			return nil, core.NewGowrightError(core.APIError, "timeout acquiring HTTP client from pool", nil)
		case <-ctx.Done():
			return nil, core.NewGowrightError(core.APIError, "context cancelled while acquiring HTTP client", ctx.Err())
		}
	}
}

// ReleaseClient returns an HTTP client instance to the pool
func (hcp *HTTPClientPool) ReleaseClient(instance *HTTPClientInstance) error {
	if instance == nil {
		return core.NewGowrightError(core.APIError, "cannot release nil HTTP client instance", nil)
	}

	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	select {
	case hcp.clients <- instance:
		hcp.stats.TotalReleased++
		hcp.stats.Available++
		hcp.stats.InUse--
		return nil
	default:
		// Pool is full, just update stats
		hcp.stats.InUse--
		return nil
	}
}

// GetStats returns current pool statistics
func (hcp *HTTPClientPool) GetStats() *HTTPClientPoolStats {
	hcp.mutex.RLock()
	defer hcp.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &HTTPClientPoolStats{
		MaxSize:       hcp.stats.MaxSize,
		Available:     hcp.stats.Available,
		InUse:         hcp.stats.InUse,
		TotalCreated:  hcp.stats.TotalCreated,
		TotalAcquired: hcp.stats.TotalAcquired,
		TotalReleased: hcp.stats.TotalReleased,
	}
}

// Cleanup closes all HTTP client instances and cleans up the pool
func (hcp *HTTPClientPool) Cleanup() error {
	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	if !hcp.initialized {
		return nil
	}

	// Close all HTTP client instances in the pool
	close(hcp.clients)
	for range hcp.clients {
		// HTTP clients don't need explicit cleanup
	}

	hcp.clients = make(chan *HTTPClientInstance, hcp.maxSize)
	hcp.initialized = false
	hcp.stats.Available = 0
	hcp.stats.InUse = 0

	return nil
}

// Acquire is an alias for AcquireClient for backward compatibility
func (hcp *HTTPClientPool) Acquire(ctx context.Context) (*HTTPClientInstance, error) {
	return hcp.AcquireClient(ctx)
}

// Release is an alias for ReleaseClient for backward compatibility
func (hcp *HTTPClientPool) Release(instance *HTTPClientInstance) error {
	if !hcp.initialized {
		return core.NewGowrightError(core.APIError, "HTTP client pool not initialized", nil)
	}
	return hcp.ReleaseClient(instance)
}

// Resize changes the maximum size of the HTTP client pool
func (hcp *HTTPClientPool) Resize(newSize int) error {
	if newSize <= 0 {
		return core.NewGowrightError(core.ConfigurationError, "HTTP client pool max size must be positive", nil)
	}

	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	if newSize == hcp.maxSize {
		return nil // No change needed
	}

	// Create new channel with new size
	newClients := make(chan *HTTPClientInstance, newSize)

	// Transfer existing clients to new channel (up to new capacity)
	transferred := 0
	for transferred < newSize {
		select {
		case instance := <-hcp.clients:
			newClients <- instance
			transferred++
		default:
			// No more clients to transfer
			goto transferComplete
		}
	}
transferComplete:

	// Drain any remaining clients from old channel if downsizing
	for len(hcp.clients) > 0 {
		<-hcp.clients
		hcp.stats.Available--
	}

	// Update pool configuration
	hcp.clients = newClients
	hcp.maxSize = newSize
	hcp.stats.MaxSize = newSize
	hcp.stats.Available = transferred

	return nil
}

// createClientInstance creates a new HTTP client instance
func (hcp *HTTPClientPool) createClientInstance() *HTTPClientInstance {
	client := resty.New()

	// Configure client with reasonable defaults
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	return &HTTPClientInstance{
		Client:    client,
		CreatedAt: time.Now(),
	}
}
