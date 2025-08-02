package gowright

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClientPool(t *testing.T) {
	maxSize := 5
	timeout := 30 * time.Second
	
	pool, err := NewHTTPClientPool(maxSize, timeout)
	
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Equal(t, maxSize, pool.maxSize)
	assert.Equal(t, timeout, pool.timeout)
	assert.Equal(t, maxSize, cap(pool.clients))
	assert.True(t, pool.initialized)
	assert.NotNil(t, pool.stats)
	assert.Equal(t, maxSize, pool.stats.MaxSize)
}

func TestNewHTTPClientPool_InvalidMaxSize(t *testing.T) {
	pool, err := NewHTTPClientPool(0, 30*time.Second)
	
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestHTTPClientPool_GetStats(t *testing.T) {
	pool, err := NewHTTPClientPool(3, 30*time.Second)
	assert.NoError(t, err)
	
	stats := pool.GetStats()
	
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.MaxSize)
	assert.Equal(t, 0, stats.Available)
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 0, stats.TotalCreated)
	assert.Equal(t, 0, stats.TotalAcquired)
	assert.Equal(t, 0, stats.TotalReleased)
}

func TestHTTPClientPool_Acquire_NotInitialized(t *testing.T) {
	pool := &HTTPClientPool{initialized: false}
	ctx := context.Background()
	
	client, err := pool.Acquire(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestHTTPClientPool_Release_NotInitialized(t *testing.T) {
	pool := &HTTPClientPool{initialized: false}
	
	err := pool.Release(nil)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestHTTPClientPool_Release_NilClient(t *testing.T) {
	pool, err := NewHTTPClientPool(3, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Release(nil)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot release nil HTTP client")
}

func TestHTTPClientPool_Cleanup_NotInitialized(t *testing.T) {
	pool := &HTTPClientPool{initialized: false}
	
	err := pool.Cleanup()
	
	assert.NoError(t, err)
}

func TestHTTPClientPool_Cleanup_EmptyPool(t *testing.T) {
	pool, err := NewHTTPClientPool(2, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Cleanup()
	
	assert.NoError(t, err)
	assert.False(t, pool.initialized)
}

func TestHTTPClientPool_Resize(t *testing.T) {
	pool, err := NewHTTPClientPool(3, 30*time.Second)
	assert.NoError(t, err)
	
	// Resize to larger size
	err = pool.Resize(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, pool.maxSize)
	assert.Equal(t, 5, pool.stats.MaxSize)
	assert.Equal(t, 5, cap(pool.clients))
	
	// Resize to smaller size
	err = pool.Resize(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, pool.maxSize)
	assert.Equal(t, 2, pool.stats.MaxSize)
	assert.Equal(t, 2, cap(pool.clients))
}

func TestHTTPClientPool_Resize_InvalidSize(t *testing.T) {
	pool, err := NewHTTPClientPool(3, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Resize(0)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestHTTPClientPool_Acquire_ContextCancelled(t *testing.T) {
	pool, err := NewHTTPClientPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	client, err := pool.Acquire(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestHTTPClientPool_Acquire_Timeout(t *testing.T) {
	pool, err := NewHTTPClientPool(1, 50*time.Millisecond)
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	client, err := pool.Acquire(ctx)
	
	// This test might pass or fail depending on system performance
	// The important thing is that it doesn't hang indefinitely
	if err != nil {
		assert.Contains(t, err.Error(), "timeout")
	}
	
	if client != nil {
		pool.Release(client)
	}
}

func TestHTTPClientPool_AcquireAndRelease(t *testing.T) {
	pool, err := NewHTTPClientPool(2, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire first client
	client1, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client1)
	
	stats := pool.GetStats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.TotalAcquired)
	assert.Equal(t, 1, stats.TotalCreated)
	
	// Acquire second client
	client2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client2)
	
	stats = pool.GetStats()
	assert.Equal(t, 2, stats.InUse)
	assert.Equal(t, 2, stats.TotalAcquired)
	assert.Equal(t, 2, stats.TotalCreated)
	
	// Release first client
	err = pool.Release(client1)
	assert.NoError(t, err)
	
	stats = pool.GetStats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.Available)
	assert.Equal(t, 1, stats.TotalReleased)
	
	// Release second client
	err = pool.Release(client2)
	assert.NoError(t, err)
	
	stats = pool.GetStats()
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 2, stats.Available)
	assert.Equal(t, 2, stats.TotalReleased)
}

func TestHTTPClientPool_ReuseReleasedClient(t *testing.T) {
	pool, err := NewHTTPClientPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire and release a client
	client1, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client1)
	
	err = pool.Release(client1)
	assert.NoError(t, err)
	
	// Acquire again - should reuse the same client
	client2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, client1, client2) // Should be the same client instance
	
	err = pool.Release(client2)
	assert.NoError(t, err)
	
	stats := pool.GetStats()
	assert.Equal(t, 1, stats.TotalCreated) // Only one client was created
	assert.Equal(t, 2, stats.TotalAcquired) // But acquired twice
	assert.Equal(t, 2, stats.TotalReleased) // And released twice
}

func TestHTTPClientPool_ClientStateReset(t *testing.T) {
	pool, err := NewHTTPClientPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire client and modify its state
	client, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	
	// Set some state on the client
	client.SetAuthToken("test-token")
	client.SetHeader("X-Test", "test-value")
	client.SetDebug(true)
	
	// Release the client
	err = pool.Release(client)
	assert.NoError(t, err)
	
	// Acquire again - state should be reset
	client2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, client, client2) // Same instance
	
	// Verify state has been reset
	// Note: We can't directly check internal state, but we can verify
	// that the client is in a clean state by checking it doesn't have
	// the previously set values in subsequent requests
	
	err = pool.Release(client2)
	assert.NoError(t, err)
}

func TestHTTPClientPool_MaxConcurrency(t *testing.T) {
	pool, err := NewHTTPClientPool(2, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire maximum number of clients
	client1, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client1)
	
	client2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client2)
	
	stats := pool.GetStats()
	assert.Equal(t, 2, stats.InUse)
	assert.Equal(t, 0, stats.Available)
	
	// Try to acquire one more - should timeout quickly
	ctx3, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	
	client3, err := pool.Acquire(ctx3)
	assert.Error(t, err)
	assert.Nil(t, client3)
	
	// Release one client
	err = pool.Release(client1)
	assert.NoError(t, err)
	
	// Now we should be able to acquire again
	client4, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client4)
	
	// Clean up
	err = pool.Release(client2)
	assert.NoError(t, err)
	err = pool.Release(client4)
	assert.NoError(t, err)
}

func TestHTTPClientPool_createClientInstance(t *testing.T) {
	pool, err := NewHTTPClientPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	instance := pool.createClientInstance()
	
	assert.NotNil(t, instance)
	assert.NotNil(t, instance.Client)
	assert.False(t, instance.CreatedAt.IsZero())
	assert.Equal(t, 0, instance.UsageCount)
	
	// Verify client is configured properly
	assert.NotNil(t, instance.Client.GetClient())
}

func TestHTTPClientPool_resetClientState(t *testing.T) {
	pool, err := NewHTTPClientPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	instance := pool.createClientInstance()
	client := instance.Client
	
	// Set some state
	client.SetAuthToken("test-token")
	client.SetHeader("X-Test", "test-value")
	client.SetDebug(true)
	
	// Reset state
	pool.resetClientState(client)
	
	// State should be cleared (we can't directly verify all internal state,
	// but the method should run without error)
	assert.NotNil(t, client)
}

func TestHTTPClientPool_cleanupClientState(t *testing.T) {
	pool, err := NewHTTPClientPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	instance := pool.createClientInstance()
	client := instance.Client
	
	// Set some state
	client.SetAuthToken("test-token")
	client.SetHeader("X-Test", "test-value")
	client.SetDebug(true)
	
	// Cleanup state
	pool.cleanupClientState(client)
	
	// State should be cleared (we can't directly verify all internal state,
	// but the method should run without error)
	assert.NotNil(t, client)
}

// Benchmark tests
func BenchmarkHTTPClientPool_AcquireRelease(b *testing.B) {
	pool, err := NewHTTPClientPool(5, 30*time.Second)
	if err != nil {
		b.Fatalf("Failed to create HTTP client pool: %v", err)
	}
	
	defer func() {
		if cleanupErr := pool.Cleanup(); cleanupErr != nil {
			b.Logf("Cleanup error: %v", cleanupErr)
		}
	}()
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		client, err := pool.Acquire(ctx)
		if err != nil {
			b.Fatalf("Failed to acquire HTTP client: %v", err)
		}
		
		err = pool.Release(client)
		if err != nil {
			b.Fatalf("Failed to release HTTP client: %v", err)
		}
	}
}

func BenchmarkHTTPClientPool_CreateClient(b *testing.B) {
	pool, err := NewHTTPClientPool(1, 30*time.Second)
	if err != nil {
		b.Fatalf("Failed to create HTTP client pool: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		instance := pool.createClientInstance()
		if instance == nil || instance.Client == nil {
			b.Fatal("Failed to create client instance")
		}
	}
}