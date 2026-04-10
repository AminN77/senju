package healthcheck

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Runner executes registered [Probe] implementations in registration order.
type Runner struct {
	mu     sync.RWMutex
	probes []Probe
}

// NewRunner returns an empty runner. Register probes before serving traffic.
func NewRunner() *Runner {
	return &Runner{}
}

// Register adds a probe. It is safe to call before serving; typical use is wiring in main.
func (r *Runner) Register(p Probe) {
	if p == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.probes = append(r.probes, p)
}

// Check runs all registered probes. If none are registered, the system is considered ready.
func (r *Runner) Check(ctx context.Context) Result {
	r.mu.RLock()
	probes := append([]Probe(nil), r.probes...)
	r.mu.RUnlock()

	if len(probes) == 0 {
		return Result{OK: true}
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	var errs []string
	for _, p := range probes {
		if err := p.Check(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p.Name(), err))
		}
	}
	return Result{OK: len(errs) == 0, Errors: errs}
}
