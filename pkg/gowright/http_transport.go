package gowright

import (
	"net"
	"net/http"
	"time"
)

// HTTPTransportConfig holds configuration for HTTP transport
type HTTPTransportConfig struct {
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	DialTimeout           time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
}

// Build creates an HTTP transport with the specified configuration
func (config *HTTPTransportConfig) Build() *http.Transport {
	transport := &http.Transport{
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
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

// CloseIdleConnections closes idle connections in the transport
func (config *HTTPTransportConfig) CloseIdleConnections() {
	// This method is called on the config, but we need to call it on the transport
	// In practice, this would be called on the actual transport instance
	// This is a placeholder method for the config struct
}
