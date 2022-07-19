package main

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// ProviderOpt ...
type ProviderOpt func(*ProviderOpts)

// Opts ...
type ProviderOpts struct {
	Timeout time.Duration
	Logger  *zap.Logger
	URL     string
}

// Configure os configuring the options.
func (o *ProviderOpts) Configure(opts ...ProviderOpt) {
	for _, opt := range opts {
		opt(o)
	}
}

// Provider ...
type Provider interface {
	// CloneWithContext ...
	CloneWithContext(ctx context.Context, url string, folder string) error
}
