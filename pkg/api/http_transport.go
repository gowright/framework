package api

import (
	"net"
	"net/http"
	"time"
)

// HTTPTransportConfig holds configuration for HTTP transport
type HTTPTransportConfig struct {
	MaxIdleConns          int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `json:"max_idle_conns_per_host"`
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout"`
	DialTimeout           time.Duration `json:"dial_timeout"`
	TLSHandshakeTimeout   time.Duration `json:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout"`
	DisableKeepAlives     bool          `json:"disable_keep_alives"`
	DisableCompression    bool          `json:"disable_compression"`
}

// DefaultHTTPTransportConfig returns default HTTP transport configuration
func DefaultHTTPTransportConfig() *HTTPTransportConfig {
	return &HTTPTransportConfig{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		DialTimeout:           30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
	}
}

// Build creates an HTTP transport with the specified configuration
func (config *HTTPTransportConfig) Build() *http.Transport {
	transport := &http.Transport{
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
		DisableCompression:    config.DisableCompression,
		DialContext: (&net.Dialer{
			Timeout: config.DialTimeout,
		}).DialContext,
	}

	// Set defaults if not specified
	if transport.MaxIdleConns == 0 {
		transport.MaxIdleConns = 100
	}
	if transport.MaxIdleConnsPerHost == 0 {
		transport.MaxIdleConnsPerHost = 10
	}
	if transport.IdleConnTimeout == 0 {
		transport.IdleConnTimeout = 90 * time.Second
	}
	if transport.TLSHandshakeTimeout == 0 {
		transport.TLSHandshakeTimeout = 10 * time.Second
	}
	if transport.ResponseHeaderTimeout == 0 {
		transport.ResponseHeaderTimeout = 30 * time.Second
	}
	if config.DialTimeout == 0 {
		transport.DialContext = (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext
	}

	return transport
}

// OptimizedTransport creates an HTTP transport optimized for testing
func OptimizedTransport() *http.Transport {
	config := &HTTPTransportConfig{
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       60 * time.Second,
		DialTimeout:           10 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
	}
	return config.Build()
}

// FastTransport creates an HTTP transport optimized for speed
func FastTransport() *http.Transport {
	config := &HTTPTransportConfig{
		MaxIdleConns:          500,
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       30 * time.Second,
		DialTimeout:           5 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    true, // Disable compression for speed
	}
	return config.Build()
}

// SecureTransport creates an HTTP transport with enhanced security settings
func SecureTransport() *http.Transport {
	config := &HTTPTransportConfig{
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       30 * time.Second,
		DialTimeout:           15 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DisableKeepAlives:     true, // Disable keep-alives for security
		DisableCompression:    false,
	}
	return config.Build()
}
