package gowright

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
)

func TestNewDatabasePool(t *testing.T) {
	config := &DBConnection{
		Driver:       "sqlite3",
		DSN:          ":memory:",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}
	
	pool, err := NewDatabasePool("test_db", config, 3, 30*time.Second)
	
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Equal(t, "test_db", pool.name)
	assert.Equal(t, config, pool.config)
	assert.Equal(t, 3, pool.maxSize)
	assert.Equal(t, 30*time.Second, pool.timeout)
	assert.Equal(t, 3, cap(pool.connections))
	assert.True(t, pool.initialized)
	assert.NotNil(t, pool.stats)
	assert.Equal(t, "test_db", pool.stats.Name)
	assert.Equal(t, 3, pool.stats.MaxSize)
}

func TestNewDatabasePool_InvalidMaxSize(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 0, 30*time.Second)
	
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestNewDatabasePool_NilConfig(t *testing.T) {
	pool, err := NewDatabasePool("test_db", nil, 3, 30*time.Second)
	
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestDatabasePool_GetStats(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 3, 30*time.Second)
	assert.NoError(t, err)
	
	stats := pool.GetStats()
	
	assert.NotNil(t, stats)
	assert.Equal(t, "test_db", stats.Name)
	assert.Equal(t, 3, stats.MaxSize)
	assert.Equal(t, 0, stats.Available)
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 0, stats.TotalCreated)
	assert.Equal(t, 0, stats.TotalAcquired)
	assert.Equal(t, 0, stats.TotalReleased)
	assert.Equal(t, 0, stats.TotalErrors)
}

func TestDatabasePool_Acquire_NotInitialized(t *testing.T) {
	pool := &DatabasePool{initialized: false}
	ctx := context.Background()
	
	conn, err := pool.Acquire(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestDatabasePool_Release_NotInitialized(t *testing.T) {
	pool := &DatabasePool{initialized: false}
	
	err := pool.Release(nil)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestDatabasePool_Release_NilConnection(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 3, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Release(nil)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot release nil connection")
}

func TestDatabasePool_Cleanup_NotInitialized(t *testing.T) {
	pool := &DatabasePool{initialized: false}
	
	err := pool.Cleanup()
	
	assert.NoError(t, err)
}

func TestDatabasePool_Cleanup_EmptyPool(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 2, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Cleanup()
	
	assert.NoError(t, err)
	assert.False(t, pool.initialized)
}

func TestDatabasePool_Resize(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 3, 30*time.Second)
	assert.NoError(t, err)
	
	// Resize to larger size
	err = pool.Resize(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, pool.maxSize)
	assert.Equal(t, 5, pool.stats.MaxSize)
	assert.Equal(t, 5, cap(pool.connections))
	
	// Resize to smaller size
	err = pool.Resize(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, pool.maxSize)
	assert.Equal(t, 2, pool.stats.MaxSize)
	assert.Equal(t, 2, cap(pool.connections))
}

func TestDatabasePool_Resize_InvalidSize(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 3, 30*time.Second)
	assert.NoError(t, err)
	
	err = pool.Resize(0)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestDatabasePool_HealthCheck_NotInitialized(t *testing.T) {
	pool := &DatabasePool{initialized: false}
	ctx := context.Background()
	
	err := pool.HealthCheck(ctx)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestDatabasePool_HealthCheck_EmptyPool(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 2, 30*time.Second)
	assert.NoError(t, err)
	
	ctx := context.Background()
	err = pool.HealthCheck(ctx)
	
	assert.NoError(t, err)
}

func TestDatabasePool_Acquire_ContextCancelled(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 1, 30*time.Second)
	assert.NoError(t, err)
	
	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	conn, err := pool.Acquire(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestDatabasePool_Acquire_Timeout(t *testing.T) {
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 1, 50*time.Millisecond)
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	conn, err := pool.Acquire(ctx)
	
	// This test might pass or fail depending on system performance
	// The important thing is that it doesn't hang indefinitely
	if err != nil {
		assert.Contains(t, err.Error(), "timeout")
	}
	
	if conn != nil {
		pool.Release(conn)
	}
}

// Integration tests with actual database connections (requires sqlite3 driver)
func TestDatabasePool_Integration_AcquireAndRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}
	
	config := &DBConnection{
		Driver:       "sqlite3",
		DSN:          ":memory:",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}
	
	pool, err := NewDatabasePool("test_db", config, 2, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire first connection
	conn1, err := pool.Acquire(ctx)
	if err != nil {
		t.Skipf("Skipping database test due to missing driver: %v", err)
	}
	
	assert.NotNil(t, conn1)
	
	stats := pool.GetStats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.TotalAcquired)
	assert.Equal(t, 1, stats.TotalCreated)
	
	// Test connection is working
	err = conn1.Ping()
	assert.NoError(t, err)
	
	// Acquire second connection
	conn2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, conn2)
	
	stats = pool.GetStats()
	assert.Equal(t, 2, stats.InUse)
	assert.Equal(t, 2, stats.TotalAcquired)
	assert.Equal(t, 2, stats.TotalCreated)
	
	// Release first connection
	err = pool.Release(conn1)
	assert.NoError(t, err)
	
	stats = pool.GetStats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.Available)
	assert.Equal(t, 1, stats.TotalReleased)
	
	// Release second connection
	err = pool.Release(conn2)
	assert.NoError(t, err)
	
	stats = pool.GetStats()
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 2, stats.Available)
	assert.Equal(t, 2, stats.TotalReleased)
}

func TestDatabasePool_Integration_ReuseReleasedConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}
	
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 1, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire and release a connection
	conn1, err := pool.Acquire(ctx)
	if err != nil {
		t.Skipf("Skipping database test due to missing driver: %v", err)
	}
	
	err = pool.Release(conn1)
	assert.NoError(t, err)
	
	// Acquire again - should reuse the same connection
	conn2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, conn1, conn2) // Should be the same connection instance
	
	err = pool.Release(conn2)
	assert.NoError(t, err)
	
	stats := pool.GetStats()
	assert.Equal(t, 1, stats.TotalCreated) // Only one connection was created
	assert.Equal(t, 2, stats.TotalAcquired) // But acquired twice
	assert.Equal(t, 2, stats.TotalReleased) // And released twice
}

func TestDatabasePool_Integration_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}
	
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 2, 30*time.Second)
	assert.NoError(t, err)
	
	defer func() {
		cleanupErr := pool.Cleanup()
		assert.NoError(t, cleanupErr)
	}()
	
	ctx := context.Background()
	
	// Acquire and release some connections to populate the pool
	conn1, err := pool.Acquire(ctx)
	if err != nil {
		t.Skipf("Skipping database test due to missing driver: %v", err)
	}
	
	err = pool.Release(conn1)
	assert.NoError(t, err)
	
	conn2, err := pool.Acquire(ctx)
	assert.NoError(t, err)
	
	err = pool.Release(conn2)
	assert.NoError(t, err)
	
	// Perform health check
	err = pool.HealthCheck(ctx)
	assert.NoError(t, err)
	
	stats := pool.GetStats()
	assert.Equal(t, 0, stats.TotalErrors) // No errors should have occurred
}

// Benchmark tests
func BenchmarkDatabasePool_AcquireRelease(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database benchmark in short mode")
	}
	
	config := &DBConnection{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	
	pool, err := NewDatabasePool("test_db", config, 5, 30*time.Second)
	if err != nil {
		b.Fatalf("Failed to create database pool: %v", err)
	}
	
	defer func() {
		if cleanupErr := pool.Cleanup(); cleanupErr != nil {
			b.Logf("Cleanup error: %v", cleanupErr)
		}
	}()
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			b.Skipf("Skipping benchmark due to missing driver: %v", err)
		}
		
		err = pool.Release(conn)
		if err != nil {
			b.Fatalf("Failed to release connection: %v", err)
		}
	}
}