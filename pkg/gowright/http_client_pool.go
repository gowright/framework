package gowright

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
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
		return nil, fmt.Errorf("HTTP client pool max size must be positive")
	}

	pool := &HTTPClientPool{
		clients: make(chan *HTTPClientInstance, maxSize),
		maxSize: maxSize,
		timeout: timeout,
		stats: &HTTPClientPoolStats{
			MaxSize: maxSize,
		},
		initialized: true,
	}

	return pool, nil
}

// Acquire gets an HTTP client from the pool
func (hcp *HTTPClientPool) Acquire(ctx context.Context) (*resty.Client, error) {
	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	if !hcp.initialized {
		return nil, fmt.Errorf("HTTP client pool not initialized")
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	var instance *HTTPClientInstance

	// Try to get an existing client from the pool
	select {
	case instance = <-hcp.clients:
		hcp.stats.Available--
	default:
		// No available clients, create a new one if we haven't reached max size
		if hcp.stats.TotalCreated < hcp.maxSize {
			instance = hcp.createClientInstance()
			hcp.stats.TotalCreated++
		} else {
			// Wait for a client to become available
			select {
			case instance = <-hcp.clients:
				hcp.stats.Available--
			case <-ctx.Done():
				return nil, fmt.Errorf("timeout waiting for HTTP client: %w", ctx.Err())
			case <-time.After(hcp.timeout):
				return nil, fmt.Errorf("timeout waiting for HTTP client")
			}
		}
	}

	// Reset client state for new usage
	hcp.resetClientState(instance.Client)

	instance.UsageCount++
	hcp.stats.InUse++
	hcp.stats.TotalAcquired++

	return instance.Client, nil
}

// Release returns an HTTP client to the pool
func (hcp *HTTPClientPool) Release(client *resty.Client) error {
	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	if !hcp.initialized {
		return fmt.Errorf("HTTP client pool not initialized")
	}

	if client == nil {
		return fmt.Errorf("cannot release nil HTTP client")
	}

	// Clean up client state before returning to pool
	hcp.cleanupClientState(client)

	// Create instance wrapper
	instance := &HTTPClientInstance{
		Client: client,
	}

	// Return to pool if there's space
	select {
	case hcp.clients <- instance:
		hcp.stats.Available++
	default:
		// Pool is full, client will be garbage collected
	}

	hcp.stats.InUse--
	hcp.stats.TotalReleased++

	return nil
}

// createClientInstance creates a new HTTP client instance
func (hcp *HTTPClientPool) createClientInstance() *HTTPClientInstance {
	client := resty.New()

	// Configure client with optimized settings for parallel execution
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	// Enable connection pooling
	transportConfig := &HTTPTransportConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	client.GetClient().Transport = transportConfig.Build()

	return &HTTPClientInstance{
		Client:    client,
		CreatedAt: time.Now(),
	}
}

// resetClientState resets the client state for new usage
func (hcp *HTTPClientPool) resetClientState(client *resty.Client) {
	// Clear any request-specific settings
	client.SetHeaders(nil)
	client.SetCookies(nil)
	client.SetAuthToken("")
	client.SetBasicAuth("", "")

	// Reset debug mode
	client.SetDebug(false)

	// Clear any custom middleware that might have been added
	client.OnBeforeRequest(nil)
	client.OnAfterResponse(nil)
	client.OnError(nil)
}

// cleanupClientState cleans up client state before returning to pool
func (hcp *HTTPClientPool) cleanupClientState(client *resty.Client) {
	// Clear any sensitive information
	client.SetAuthToken("")
	client.SetBasicAuth("", "")
	client.SetHeaders(nil)
	client.SetCookies(nil)

	// Clear any temporary settings
	client.SetDebug(false)
}

// Cleanup closes all clients in the pool
func (hcp *HTTPClientPool) Cleanup() error {
	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	if !hcp.initialized {
		return nil
	}

	// Clear all clients from the pool
	for {
		select {
		case instance := <-hcp.clients:
			// Close any persistent connections
			if transport, ok := instance.Client.GetClient().Transport.(*http.Transport); ok {
				transport.CloseIdleConnections()
			}
		default:
			// No more clients in pool
			goto cleanup_done
		}
	}

cleanup_done:
	hcp.initialized = false
	return nil
}

// GetStats returns current pool statistics
func (hcp *HTTPClientPool) GetStats() *HTTPClientPoolStats {
	hcp.mutex.RLock()
	defer hcp.mutex.RUnlock()

	// Create a copy of stats to avoid race conditions
	return &HTTPClientPoolStats{
		MaxSize:       hcp.stats.MaxSize,
		Available:     hcp.stats.Available,
		InUse:         hcp.stats.InUse,
		TotalCreated:  hcp.stats.TotalCreated,
		TotalAcquired: hcp.stats.TotalAcquired,
		TotalReleased: hcp.stats.TotalReleased,
	}
}

// Resize changes the maximum size of the HTTP client pool
func (hcp *HTTPClientPool) Resize(newSize int) error {
	hcp.mutex.Lock()
	defer hcp.mutex.Unlock()

	if newSize <= 0 {
		return fmt.Errorf("HTTP client pool size must be positive")
	}

	// Create new channel with new size
	newClients := make(chan *HTTPClientInstance, newSize)

	// Move existing clients to new channel
	movedClients := 0
	for {
		select {
		case instance := <-hcp.clients:
			if movedClients < newSize {
				newClients <- instance
				movedClients++
			} else {
				// New channel is full, close excess client connections
				if transport, ok := instance.Client.GetClient().Transport.(*http.Transport); ok {
					transport.CloseIdleConnections()
				}
				hcp.stats.Available--
			}
		default:
			// No more clients to move
			goto resize_done
		}
	}

resize_done:
	// Update pool configuration
	hcp.clients = newClients
	hcp.maxSize = newSize
	hcp.stats.MaxSize = newSize
	hcp.stats.Available = movedClients

	return nil
}
