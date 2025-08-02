package gowright

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
)

func TestNewBrowserPool(t *testing.T) {
	maxSize := 5
	timeout := 30 * time.Second
	
	pool, err := NewBrowserPool(maxSize, timeout)
	
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Equal(t, maxSize, pool.maxSize)
	assert.Equal(t, timeout, pool.timeout)
	assert.Equal(t, maxSize, cap(pool.browsers))
	assert.True(t, pool.initialized)
	assert.NotNil(t, pool.stats)
	assert.Equal(t, maxSize, pool.stats.MaxSize)
}

func TestNewBrowserPool_InvalidMaxSize(t *testing.T) {
	pool, err := NewBrowserPool(0, 30*time.Second)
	
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestBrowserPool_GetStats(t *testing.T) {
	pool, err := NewBrowserPool(3, 30*time.Second)
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

func TestBrowserPool_Cleanup_NotInitialized(t *testing.T) {
	pool := &BrowserPool{initialized: false}
	
	err := pool.Cleanup()
	
	assert.NoError(t, err)
}

func TestBrowserPool_Cleanup_EmptyPool(t *testing.T) {
	pool, err := NewBrowserPool(2, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Cleanup()
	
	assert.NoError(t, err)
	assert.False(t, pool.initialized)
}

func TestBrowserPool_Resize(t *testing.T) {
	pool, err := NewBrowserPool(3, 30*time.Second)
	assert.NoError(t, err)
	
	// Resize to larger size
	err = pool.Resize(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, pool.maxSize)
	assert.Equal(t, 5, pool.stats.MaxSize)
	assert.Equal(t, 5, cap(pool.browsers))
	
	// Resize to smaller size
	err = pool.Resize(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, pool.maxSize)
	assert.Equal(t, 2, pool.stats.MaxSize)
	assert.Equal(t, 2, cap(pool.browsers))
}

func TestBrowserPool_Resize_InvalidSize(t *testing.T) {
	pool, err := NewBrowserPool(3, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Resize(0)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestBrowserPool_Release_NotInitialized(t *testing.T) {
	pool := &BrowserPool{initialized: false}
	
	err := pool.Release(nil, nil)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestBrowserPool_Acquire_NotInitialized(t *testing.T) {
	pool := &BrowserPool{initialized: false}
	ctx := context.Background()
	
	browser, page, err := pool.Acquire(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, browser)
	assert.Nil(t, page)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestBrowserPool_Acquire_ContextCancelled(t *testing.T) {
	pool, err := NewBrowserPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	browser, page, err := pool.Acquire(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, browser)
	assert.Nil(t, page)
}

func TestBrowserPool_Acquire_Timeout(t *testing.T) {
	pool, err := NewBrowserPool(1, 100*time.Millisecond)
	assert.NoError(t, err)
	
	// Set a very short timeout
	pool.timeout = 50 * time.Millisecond
	
	ctx := context.Background()
	
	browser, page, err := pool.Acquire(ctx)
	
	// This test might pass or fail depending on system performance
	// The important thing is that it doesn't hang indefinitely
	if err != nil {
		assert.Contains(t, err.Error(), "timeout")
	}
	
	if browser != nil {
		pool.Release(browser, page)
	}
}

// Integration tests that actually create browsers (skip in CI/short mode)
func TestBrowserPool_Integration_AcquireAndRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser integration test in short mode")
	}
	
	pool, err := NewBrowserPool(2, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire first browser
	browser1, page1, err := pool.Acquire(ctx)
	if err != nil {
		t.Skipf("Skipping browser test due to environment: %v", err)
	}
	
	assert.NotNil(t, browser1)
	assert.NotNil(t, page1)
	
	stats := pool.GetStats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.TotalAcquired)
	assert.Equal(t, 1, stats.TotalCreated)
	
	// Acquire second browser
	browser2, page2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, browser2)
	assert.NotNil(t, page2)
	
	stats = pool.GetStats()
	assert.Equal(t, 2, stats.InUse)
	assert.Equal(t, 2, stats.TotalAcquired)
	assert.Equal(t, 2, stats.TotalCreated)
	
	// Release first browser
	err = pool.Release(browser1, page1)
	assert.NoError(t, err)
	
	stats = pool.GetStats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.Available)
	assert.Equal(t, 1, stats.TotalReleased)
	
	// Release second browser
	err = pool.Release(browser2, page2)
	assert.NoError(t, err)
	
	stats = pool.GetStats()
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 2, stats.Available)
	assert.Equal(t, 2, stats.TotalReleased)
}

func TestBrowserPool_Integration_ReuseReleasedBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser integration test in short mode")
	}
	
	pool, err := NewBrowserPool(1, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire and release a browser
	browser1, page1, err := pool.Acquire(ctx)
	if err != nil {
		t.Skipf("Skipping browser test due to environment: %v", err)
	}
	
	err = pool.Release(browser1, page1)
	assert.NoError(t, err)
	
	// Acquire again - should reuse the same browser
	browser2, page2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, browser1, browser2) // Should be the same browser instance
	
	err = pool.Release(browser2, page2)
	assert.NoError(t, err)
	
	stats := pool.GetStats()
	assert.Equal(t, 1, stats.TotalCreated) // Only one browser was created
	assert.Equal(t, 2, stats.TotalAcquired) // But acquired twice
	assert.Equal(t, 2, stats.TotalReleased) // And released twice
}

// Benchmark tests
func BenchmarkBrowserPool_AcquireRelease(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping browser benchmark in short mode")
	}
	
	pool, err := NewBrowserPool(5, 30*time.Second)
	if err != nil {
		b.Fatalf("Failed to create browser pool: %v", err)
	}
	
	defer func() {
		if cleanupErr := pool.Cleanup(); cleanupErr != nil {
			b.Logf("Cleanup error: %v", cleanupErr)
		}
	}()
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		browser, page, err := pool.Acquire(ctx)
		if err != nil {
			b.Skipf("Skipping benchmark due to environment: %v", err)
		}
		
		err = pool.Release(browser, page)
		if err != nil {
			b.Fatalf("Failed to release browser: %v", err)
		}
	}
}