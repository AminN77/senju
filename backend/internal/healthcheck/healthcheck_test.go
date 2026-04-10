package healthcheck

import (
	"context"
	"errors"
	"testing"
)

type fakeProbe struct {
	name string
	err  error
}

func (f fakeProbe) Name() string                { return f.name }
func (f fakeProbe) Check(context.Context) error { return f.err }

func TestFormatBody(t *testing.T) {
	t.Parallel()
	ok := Result{OK: true}
	if got := FormatBody(ok); got != "ok\n" {
		t.Fatalf("got %q", got)
	}
	fail := Result{OK: false, Errors: []string{"a: x", "b: y"}}
	got := FormatBody(fail)
	if len(got) < 10 {
		t.Fatalf("unexpected %q", got)
	}
}

func TestRunner_Check(t *testing.T) {
	t.Parallel()
	r := NewRunner()
	res := r.Check(context.Background())
	if !res.OK {
		t.Fatalf("empty runner should be ready, got %+v", res)
	}

	r.Register(fakeProbe{name: "one", err: nil})
	r.Register(fakeProbe{name: "two", err: errors.New("down")})
	res = r.Check(context.Background())
	if res.OK || len(res.Errors) != 1 {
		t.Fatalf("got %+v", res)
	}
}
