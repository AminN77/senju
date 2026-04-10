// Package tcp provides a generic TCP dial [healthcheck.Probe].
package tcp

import (
	"context"
	"fmt"
	"net"

	"github.com/AminN77/senju/backend/internal/healthcheck"
)

// Probe dials address on network (typically "tcp"). Name identifies this instance in readiness output.
type Probe struct {
	name    string
	network string
	address string
}

// New returns a [healthcheck.Probe]. Network is usually "tcp".
func New(name, network, address string) healthcheck.Probe {
	return &Probe{name: name, network: network, address: address}
}

// Name implements [healthcheck.Probe].
func (p *Probe) Name() string { return p.name }

// Check implements [healthcheck.Probe].
func (p *Probe) Check(ctx context.Context) error {
	if p.address == "" {
		return fmt.Errorf("address not configured")
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, p.network, p.address)
	if err != nil {
		return err
	}
	if err := conn.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return nil
}
