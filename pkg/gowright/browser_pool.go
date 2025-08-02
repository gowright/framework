package gowright

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// BrowserPool manages a pool of browser instances for concurrent testing
type BrowserPool struct {
	browsers    chan *BrowserInstance
	maxSize     int
	timeout     time.Duration
	launcher    *launcher.Launcher
	mutex       sync.RWMutex
	stats       *BrowserPoolStats
	initialized bool
}

// BrowserInstance represents a browser instance in the pool
type BrowserInstance struct {
	Browser   *rod.Browser
	CreatedAt time.Time
	UsageCount int
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
		return nil, fmt.Errorf("browser pool max size must be positive")
	}
	
	// Create launcher with optimized settings for parallel execution
	l := launcher.New().
		Headless(true).
		NoSandbox(true).
		Set("disable-web-security").
		Set("disable-features", "VizDisplayCompositor").
		Set("disable-background-timer-throttling").
		Set("disable-backgrounding-occluded-windows").
		Set("disable-renderer-backgrounding").
		Set("disable-background-networking").
		Set("disable-ipc-flooding-protection")
	
	pool := &BrowserPool{
		browsers:    make(chan *BrowserInstance, maxSize),
		maxSize:     maxSize,
		timeout:     timeout,
		launcher:    l,
		stats: &BrowserPoolStats{
			MaxSize: maxSize,
		},
		initialized: true,
	}
	
	return pool, nil
}

// Acquire gets a browser instance from the pool
func (bp *BrowserPool) Acquire(ctx context.Context) (*rod.Browser, *rod.Page, error) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	if !bp.initialized {
		return nil, nil, fmt.Errorf("browser pool not initialized")
	}
	
	var instance *BrowserInstance
	
	// Try to get an existing browser from the pool
	select {
	case instance = <-bp.browsers:
		bp.stats.Available--
	default:
		// No available browsers, create a new one if we haven't reached max size
		if bp.stats.TotalCreated < bp.maxSize {
			var err error
			instance, err = bp.createBrowserInstance()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create browser instance: %w", err)
			}
			bp.stats.TotalCreated++
		} else {
			// Wait for a browser to become available
			select {
			case instance = <-bp.browsers:
				bp.stats.Available--
			case <-ctx.Done():
				return nil, nil, fmt.Errorf("timeout waiting for browser: %w", ctx.Err())
			case <-time.After(bp.timeout):
				return nil, nil, fmt.Errorf("timeout waiting for browser")
			}
		}
	}
	
	// Create a new page for this test
	page := instance.Browser.MustPage()
	if page == nil {
		// If page creation fails, try to create a new browser instance
		if newInstance, createErr := bp.createBrowserInstance(); createErr == nil {
			instance = newInstance
			page = instance.Browser.MustPage()
		}
		
		if page == nil {
			// Return the browser to the pool if page creation still fails
			bp.returnBrowserToPool(instance)
			return nil, nil, fmt.Errorf("failed to create page")
		}
	}
		

	
	instance.UsageCount++
	bp.stats.InUse++
	bp.stats.TotalAcquired++
	
	return instance.Browser, page, nil
}

// Release returns a browser instance to the pool
func (bp *BrowserPool) Release(browser *rod.Browser, page *rod.Page) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	if !bp.initialized {
		return fmt.Errorf("browser pool not initialized")
	}
	
	// Close the page
	if page != nil {
		if err := page.Close(); err != nil {
			// Log the error but don't fail the release
			fmt.Printf("Warning: failed to close page: %v\n", err)
		}
	}
	
	// Find the browser instance
	instance := &BrowserInstance{
		Browser: browser,
	}
	
	// Check if browser is still healthy
	if bp.isBrowserHealthy(browser) {
		// Return to pool if there's space
		select {
		case bp.browsers <- instance:
			bp.stats.Available++
		default:
			// Pool is full, close the browser
			if err := browser.Close(); err != nil {
				fmt.Printf("Warning: failed to close browser: %v\n", err)
			}
		}
	} else {
		// Browser is not healthy, close it
		if err := browser.Close(); err != nil {
			fmt.Printf("Warning: failed to close unhealthy browser: %v\n", err)
		}
	}
	
	bp.stats.InUse--
	bp.stats.TotalReleased++
	
	return nil
}

// createBrowserInstance creates a new browser instance
func (bp *BrowserPool) createBrowserInstance() (*BrowserInstance, error) {
	// Launch browser
	url, err := bp.launcher.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}
	
	// Connect to browser
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}
	
	return &BrowserInstance{
		Browser:   browser,
		CreatedAt: time.Now(),
	}, nil
}

// returnBrowserToPool returns a browser instance to the pool without closing the page
func (bp *BrowserPool) returnBrowserToPool(instance *BrowserInstance) {
	select {
	case bp.browsers <- instance:
		bp.stats.Available++
	default:
		// Pool is full, close the browser
		if err := instance.Browser.Close(); err != nil {
			fmt.Printf("Warning: failed to close browser: %v\n", err)
		}
	}
}

// isBrowserHealthy checks if a browser instance is still healthy
func (bp *BrowserPool) isBrowserHealthy(browser *rod.Browser) bool {
	// Try to get browser version as a health check
	_, err := browser.Version()
	return err == nil
}

// Cleanup closes all browsers in the pool
func (bp *BrowserPool) Cleanup() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	if !bp.initialized {
		return nil
	}
	
	var errors []error
	
	// Close all browsers in the pool
	for {
		select {
		case instance := <-bp.browsers:
			if err := instance.Browser.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close browser: %w", err))
			}
		default:
			// No more browsers in pool
			goto cleanup_done
		}
	}
	
cleanup_done:
	bp.initialized = false
	
	if len(errors) > 0 {
		return fmt.Errorf("browser pool cleanup errors: %v", errors)
	}
	
	return nil
}

// GetStats returns current pool statistics
func (bp *BrowserPool) GetStats() *BrowserPoolStats {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()
	
	// Create a copy of stats to avoid race conditions
	return &BrowserPoolStats{
		MaxSize:       bp.stats.MaxSize,
		Available:     bp.stats.Available,
		InUse:         bp.stats.InUse,
		TotalCreated:  bp.stats.TotalCreated,
		TotalAcquired: bp.stats.TotalAcquired,
		TotalReleased: bp.stats.TotalReleased,
	}
}

// Resize changes the maximum size of the browser pool
func (bp *BrowserPool) Resize(newSize int) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	if newSize <= 0 {
		return fmt.Errorf("browser pool size must be positive")
	}
	
	if newSize < bp.maxSize {
		// Shrinking pool - close excess browsers
		excess := bp.maxSize - newSize
		for i := 0; i < excess; i++ {
			select {
			case instance := <-bp.browsers:
				if err := instance.Browser.Close(); err != nil {
					fmt.Printf("Warning: failed to close browser during resize: %v\n", err)
				}
				bp.stats.Available--
			default:
				// No more browsers to remove
				break
			}
		}
	}
	
	bp.maxSize = newSize
	bp.stats.MaxSize = newSize
	
	// Create new channel with new size
	newBrowsers := make(chan *BrowserInstance, newSize)
	
	// Move existing browsers to new channel
	for {
		select {
		case instance := <-bp.browsers:
			select {
			case newBrowsers <- instance:
			default:
				// New channel is full, close excess browser
				if err := instance.Browser.Close(); err != nil {
					fmt.Printf("Warning: failed to close excess browser: %v\n", err)
				}
				bp.stats.Available--
			}
		default:
			goto resize_done
		}
	}
	
resize_done:
	bp.browsers = newBrowsers
	return nil
}