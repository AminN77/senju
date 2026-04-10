// Package httpget provides a generic HTTP GET [healthcheck.Probe] (expects HTTP 200).
package httpget

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AminN77/senju/backend/internal/healthcheck"
)

// DefaultClient is a short-timeout client suitable for readiness checks.
func DefaultClient() *http.Client {
	return &http.Client{Timeout: 2 * time.Second}
}

// Probe performs GET url and expects status 200. Name identifies this instance in readiness output.
type Probe struct {
	name   string
	url    string
	client *http.Client
}

// New returns a [healthcheck.Probe]. If client is nil, [DefaultClient] is used.
func New(name, url string, client *http.Client) healthcheck.Probe {
	if client == nil {
		client = DefaultClient()
	}
	return &Probe{name: name, url: url, client: client}
}

// Name implements [healthcheck.Probe].
func (p *Probe) Name() string { return p.name }

// Check implements [healthcheck.Probe].
func (p *Probe) Check(ctx context.Context) error {
	if p.url == "" {
		return fmt.Errorf("url not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
