package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
)

// DatabasePool manages a pool of database connections for concurrent testing
type DatabasePool struct {
	name        string
	config      *config.DatabaseConnection
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
func NewDatabasePool(name string, cfg *config.DatabaseConnection, maxSize int, timeout time.Duration) (*DatabasePool, error) {
	if maxSize <= 0 {
		return nil, core.NewGowrightError(core.ConfigurationError, "database pool max size must be positive", nil)
	}

	if cfg == nil {
		return nil, core.NewGowrightError(core.ConfigurationError, "database connection config cannot be nil", nil)
	}

	pool := &DatabasePool{
		name:        name,
		config:      cfg,
		connections: make(chan *sql.DB, maxSize),
		maxSize:     maxSize,
		timeout:     timeout,
		stats: &DBPoolStats{
			Name:    name,
			MaxSize: maxSize,
		},
	}

	return pool, nil
}

// Initialize initializes the database pool
func (dp *DatabasePool) Initialize() error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	if dp.initialized {
		return nil
	}

	// Pre-create some database connections
	for i := 0; i < dp.maxSize/2; i++ {
		conn, err := dp.createConnection()
		if err != nil {
			return core.NewGowrightError(core.DatabaseError, "failed to create database connection", err)
		}
		dp.connections <- conn
		dp.stats.TotalCreated++
		dp.stats.Available++
	}

	dp.initialized = true
	return nil
}

// AcquireConnection acquires a database connection from the pool
func (dp *DatabasePool) AcquireConnection(ctx context.Context) (*sql.DB, error) {
	if !dp.initialized {
		return nil, core.NewGowrightError(core.DatabaseError, "database pool not initialized", nil)
	}

	select {
	case conn := <-dp.connections:
		// Test the connection
		if err := conn.Ping(); err != nil {
			// Connection is stale, create a new one
			if closeErr := conn.Close(); closeErr != nil {
				fmt.Printf("Error closing stale connection: %v\n", closeErr)
			}
			newConn, createErr := dp.createConnection()
			if createErr != nil {
				dp.mutex.Lock()
				dp.stats.TotalErrors++
				dp.mutex.Unlock()
				return nil, core.NewGowrightError(core.DatabaseError, "failed to create replacement connection", createErr)
			}
			conn = newConn
		}

		dp.mutex.Lock()
		dp.stats.TotalAcquired++
		dp.stats.Available--
		dp.stats.InUse++
		dp.mutex.Unlock()
		return conn, nil

	case <-time.After(dp.timeout):
		return nil, core.NewGowrightError(core.DatabaseError, "timeout acquiring connection from pool", nil)

	case <-ctx.Done():
		return nil, core.NewGowrightError(core.DatabaseError, "context cancelled while acquiring connection", ctx.Err())
	}
}

// ReleaseConnection returns a database connection to the pool
func (dp *DatabasePool) ReleaseConnection(conn *sql.DB) error {
	if conn == nil {
		return core.NewGowrightError(core.DatabaseError, "cannot release nil connection", nil)
	}

	// Test the connection before returning to pool
	if err := conn.Ping(); err != nil {
		// Connection is bad, close it
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Printf("Error closing bad connection: %v\n", closeErr)
		}
		dp.mutex.Lock()
		dp.stats.InUse--
		dp.stats.TotalErrors++
		dp.mutex.Unlock()
		return nil
	}

	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	select {
	case dp.connections <- conn:
		dp.stats.TotalReleased++
		dp.stats.Available++
		dp.stats.InUse--
		return nil
	default:
		// Pool is full, close the connection
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Printf("Error closing excess connection: %v\n", closeErr)
		}
		dp.stats.InUse--
		return nil
	}
}

// GetStats returns current pool statistics
func (dp *DatabasePool) GetStats() *DBPoolStats {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()

	// Return a copy to avoid race conditions
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

// Cleanup closes all connections and cleans up the pool
func (dp *DatabasePool) Cleanup() error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	if !dp.initialized {
		return nil
	}

	// Close all connections in the pool
	close(dp.connections)
	for conn := range dp.connections {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error closing database connection: %v\n", err)
		}
	}

	dp.connections = make(chan *sql.DB, dp.maxSize)
	dp.initialized = false
	dp.stats.Available = 0
	dp.stats.InUse = 0

	return nil
}

// createConnection creates a new database connection
func (dp *DatabasePool) createConnection() (*sql.DB, error) {
	// Build connection string based on driver
	var dsn string
	switch dp.config.Driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			dp.config.Username, dp.config.Password, dp.config.Host, dp.config.Port, dp.config.Database)
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dp.config.Host, dp.config.Port, dp.config.Username, dp.config.Password, dp.config.Database, dp.config.SSLMode)
	case "sqlite3":
		dsn = dp.config.Database
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", dp.config.Driver)
	}

	db, err := sql.Open(dp.config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Printf("Error closing failed connection: %v\n", closeErr)
		}
		return nil, err
	}

	return db, nil
}
