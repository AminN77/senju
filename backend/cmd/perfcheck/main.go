// Command perfcheck compares benchmark output against repository baselines.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type baseline struct {
	Policy struct {
		WarnRegressionPct float64 `json:"warn_regression_pct"`
		FailRegressionPct float64 `json:"fail_regression_pct"`
	} `json:"policy"`
	Benchmarks map[string]struct {
		TargetNSOp float64 `json:"target_ns_op"`
	} `json:"benchmarks"`
}

var benchLine = regexp.MustCompile(`^(Benchmark[^\s]+)\s+\d+\s+([0-9.]+)\s+ns/op`)

func main() {
	const (
		baselinePath = "./perf/baseline.json"
		resultsPath  = "/tmp/bench.txt"
	)
	if len(os.Args) != 1 {
		fail("usage: go run ./cmd/perfcheck")
	}
	base, err := loadBaseline(baselinePath)
	if err != nil {
		fail(err.Error())
	}
	results, err := loadBenchResults(resultsPath)
	if err != nil {
		fail(err.Error())
	}
	hardFail := false
	for name, cfg := range base.Benchmarks {
		got, ok := results[name]
		if !ok {
			hardFail = true
			fmt.Printf("FAIL: benchmark %s not found in results\n", name)
			continue
		}
		if cfg.TargetNSOp <= 0 {
			fmt.Printf("WARN: benchmark %s has non-positive target\n", name)
			continue
		}
		regression := ((got - cfg.TargetNSOp) / cfg.TargetNSOp) * 100
		if regression > base.Policy.FailRegressionPct {
			hardFail = true
			fmt.Printf("FAIL: %s regressed %.2f%% (target %.2f ns/op, got %.2f ns/op)\n", name, regression, cfg.TargetNSOp, got)
			continue
		}
		if regression > base.Policy.WarnRegressionPct {
			fmt.Printf("WARN: %s regressed %.2f%% (target %.2f ns/op, got %.2f ns/op)\n", name, regression, cfg.TargetNSOp, got)
			continue
		}
		fmt.Printf("PASS: %s %.2f ns/op (target %.2f)\n", name, got, cfg.TargetNSOp)
	}
	if hardFail {
		os.Exit(1)
	}
}

func loadBaseline(path string) (baseline, error) {
	// #nosec G304 -- path is provided by trusted CI/local command invocation.
	raw, err := os.ReadFile(path)
	if err != nil {
		return baseline{}, err
	}
	var b baseline
	if err := json.Unmarshal(raw, &b); err != nil {
		return baseline{}, err
	}
	if len(b.Benchmarks) == 0 {
		return baseline{}, errors.New("baseline has no benchmarks")
	}
	return b, nil
}

func loadBenchResults(path string) (map[string]float64, error) {
	// #nosec G304 -- path is provided by trusted CI/local command invocation.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	out := map[string]float64{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		m := benchLine.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		v, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			return nil, err
		}
		name := m[1]
		if idx := strings.LastIndex(name, "-"); idx > 0 {
			if _, err := strconv.Atoi(name[idx+1:]); err == nil {
				name = name[:idx]
			}
		}
		out[name] = v
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(2)
}
