package gowright

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// DatabasePool manages a pool of database connections for concurrent testing
type DatabasePool struct {
	name        string
	config      *DBConnection
	connections chan *sql.DB
	maxSize     int
	timeout     time.Duration
	mutex       sync.RWMutex
	stats       *DBPoolStats
	initialized bool
}

// DBPoolStats holds statistics about database pool usage
type DBPoolStats struct {
	Name          string `json:"name"`
	MaxSize       int    `json:"max_size"`
	Available     int    `json:"available"`
	InUse         int    `json:"in_use"`
	TotalCreated  int    `json:"total_created"`
	TotalAcquired int    `json:"total_acquired"`
	TotalReleased int    `json:"total_released"`
	TotalErrors   int    `json:"total_errors"`
}

// NewDatabasePool creates a new database connection pool
func NewDatabasePool(name string, config *DBConnection, maxSize int, timeout time.Duration) (*DatabasePool, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("database pool max size must be positive")
	}

	if config == nil {
		return nil, fmt.Errorf("database connection config cannot be nil")
	}

	pool := &DatabasePool{
		name:        name,
		config:      config,
		connections: make(chan *sql.DB, maxSize),
		maxSize:     maxSize,
		timeout:     timeout,
		stats: &DBPoolStats{
			Name:    name,
			MaxSize: maxSize,
		},
		initialized: true,
	}

	return pool, nil
}

// Acquire gets a database connection from the pool
func (dp *DatabasePool) Acquire(ctx context.Context) (*sql.DB, error) {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	if !dp.initialized {
		return nil, fmt.Errorf("database pool not initialized")
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	var conn *sql.DB

	// Try to get an existing connection from the pool
	select {
	case conn = <-dp.connections:
		dp.stats.Available--
	default:
		// No available connections, create a new one if we haven't reached max size
		if dp.stats.TotalCreated < dp.maxSize {
			var err error
			conn, err = dp.createConnection()
			if err != nil {
				dp.stats.TotalErrors++
				return nil, fmt.Errorf("failed to create database connection: %w", err)
			}
			dp.stats.TotalCreated++
		} else {
			// Wait for a connection to become available
			select {
			case conn = <-dp.connections:
				dp.stats.Available--
			case <-ctx.Done():
				return nil, fmt.Errorf("timeout waiting for database connection: %w", ctx.Err())
			case <-time.After(dp.timeout):
				return nil, fmt.Errorf("timeout waiting for database connection")
			}
		}
	}

	// Test the connection to make sure it's still valid
	if err := conn.PingContext(ctx); err != nil {
		// Connection is invalid, try to create a new one
		if newConn, createErr := dp.createConnection(); createErr == nil {
			// Close the invalid connection
			_ = conn.Close()
			conn = newConn
		} else {
			// Return the invalid connection to the pool and return error
			dp.returnConnectionToPool(conn)
			dp.stats.TotalErrors++
			return nil, fmt.Errorf("database connection is invalid and failed to create new one: %w", err)
		}
	}

	dp.stats.InUse++
	dp.stats.TotalAcquired++

	return conn, nil
}

// Release returns a database connection to the pool
func (dp *DatabasePool) Release(conn *sql.DB) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	if !dp.initialized {
		return fmt.Errorf("database pool not initialized")
	}

	if conn == nil {
		return fmt.Errorf("cannot release nil connection")
	}

	// Test the connection before returning it to the pool
	if err := conn.Ping(); err != nil {
		// Connection is invalid, close it
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close invalid database connection: %v\n", closeErr)
		}
		dp.stats.TotalErrors++
	} else {
		// Connection is valid, return to pool
		dp.returnConnectionToPool(conn)
	}

	dp.stats.InUse--
	dp.stats.TotalReleased++

	return nil
}

// createConnection creates a new database connection
func (dp *DatabasePool) createConnection() (*sql.DB, error) {
	conn, err := sql.Open(dp.config.Driver, dp.config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	if dp.config.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(dp.config.MaxOpenConns)
	}

	if dp.config.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(dp.config.MaxIdleConns)
	}

	// Set connection lifetime to prevent stale connections
	conn.SetConnMaxLifetime(30 * time.Minute)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	// Test the connection
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}

// returnConnectionToPool returns a connection to the pool
func (dp *DatabasePool) returnConnectionToPool(conn *sql.DB) {
	select {
	case dp.connections <- conn:
		dp.stats.Available++
	default:
		// Pool is full, close the connection
		if err := conn.Close(); err != nil {
			fmt.Printf("Warning: failed to close database connection: %v\n", err)
		}
	}
}

// Cleanup closes all connections in the pool
func (dp *DatabasePool) Cleanup() error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	if !dp.initialized {
		return nil
	}

	var errors []error

	// Close all connections in the pool
	for {
		select {
		case conn := <-dp.connections:
			if err := conn.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close database connection: %w", err))
			}
		default:
			// No more connections in pool
			goto cleanup_done
		}
	}

cleanup_done:
	dp.initialized = false

	if len(errors) > 0 {
		return fmt.Errorf("database pool cleanup errors: %v", errors)
	}

	return nil
}

// GetStats returns current pool statistics
func (dp *DatabasePool) GetStats() *DBPoolStats {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()

	// Create a copy of stats to avoid race conditions
	return &DBPoolStats{
		Name:          dp.stats.Name,
		MaxSize:       dp.stats.MaxSize,
		Available:     dp.stats.Available,
		InUse:         dp.stats.InUse,
		TotalCreated:  dp.stats.TotalCreated,
		TotalAcquired: dp.stats.TotalAcquired,
		TotalReleased: dp.stats.TotalReleased,
		TotalErrors:   dp.stats.TotalErrors,
	}
}

// Resize changes the maximum size of the database pool
func (dp *DatabasePool) Resize(newSize int) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	if newSize <= 0 {
		return fmt.Errorf("database pool size must be positive")
	}

	// Create new channel with new size
	newConnections := make(chan *sql.DB, newSize)

	// Collect existing connections
	var existingConnections []*sql.DB
	for {
		select {
		case conn := <-dp.connections:
			existingConnections = append(existingConnections, conn)
			dp.stats.Available--
		default:
			// No more connections to collect
			goto collect_done
		}
	}

collect_done:
	// If shrinking, close excess connections
	if newSize < len(existingConnections) {
		for i := newSize; i < len(existingConnections); i++ {
			if err := existingConnections[i].Close(); err != nil {
				fmt.Printf("Warning: failed to close database connection during resize: %v\n", err)
			}
		}
		existingConnections = existingConnections[:newSize]
	}

	// Put remaining connections into new channel
	for _, conn := range existingConnections {
		select {
		case newConnections <- conn:
			dp.stats.Available++
		default:
			// This shouldn't happen since we sized the channel correctly
			if err := conn.Close(); err != nil {
				fmt.Printf("Warning: failed to close excess database connection: %v\n", err)
			}
		}
	}

	// Update pool configuration
	dp.maxSize = newSize
	dp.stats.MaxSize = newSize
	dp.connections = newConnections

	return nil
}

// HealthCheck performs a health check on all connections in the pool
func (dp *DatabasePool) HealthCheck(ctx context.Context) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	if !dp.initialized {
		return fmt.Errorf("database pool not initialized")
	}

	var errors []error
	var healthyConnections []*sql.DB

	// Check all connections in the pool
	for {
		select {
		case conn := <-dp.connections:
			if err := conn.PingContext(ctx); err != nil {
				// Connection is unhealthy, close it
				if closeErr := conn.Close(); closeErr != nil {
					errors = append(errors, fmt.Errorf("failed to close unhealthy connection: %w", closeErr))
				}
				dp.stats.TotalErrors++
				dp.stats.Available--
			} else {
				// Connection is healthy, keep it
				healthyConnections = append(healthyConnections, conn)
			}
		default:
			// No more connections to check
			goto health_check_done
		}
	}

health_check_done:
	// Return healthy connections to the pool
	for _, conn := range healthyConnections {
		select {
		case dp.connections <- conn:
		default:
			// Pool is somehow full, close the connection
			if err := conn.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close excess healthy connection: %w", err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("database pool health check errors: %v", errors)
	}

	return nil
}
