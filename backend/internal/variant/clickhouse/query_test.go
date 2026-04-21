package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuildWhere(t *testing.T) {
	t.Parallel()
	posMin := uint32(10)
	posMax := uint32(20)
	sql, args := buildWhere(QueryFilters{
		Chromosome:  "chr1",
		PositionMin: &posMin,
		PositionMax: &posMax,
		Gene:        "TP53",
	})
	if !strings.Contains(sql, "chrom = ?") || !strings.Contains(sql, "pos >= ?") || !strings.Contains(sql, "lower(info) LIKE ?") {
		t.Fatalf("where %q", sql)
	}
	if len(args) != 5 {
		t.Fatalf("args %+v", args)
	}
}

func TestExtractGene(t *testing.T) {
	t.Parallel()
	if got := extractGene("DP=10;GENE=TP53;AF=0.2"); got != "TP53" {
		t.Fatalf("gene %q", got)
	}
	if got := extractGene("DP=10;GENE_SYMBOL=BRCA1;AF=0.2"); got != "BRCA1" {
		t.Fatalf("gene symbol %q", got)
	}
}

func TestQueryRepository_IntegrationFiltersAndPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	dsn := strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN"))
	if dsn == "" {
		t.Skip("CLICKHOUSE_DSN not set")
	}
	ctx := context.Background()
	loader, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = loader.Close() }()
	if err := loader.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	repo, err := NewQueryRepository(loader.db)
	if err != nil {
		t.Fatal(err)
	}

	chrom := "chr14_" + uuid.NewString()[:8]
	if err := seedVariantsForTests(ctx, loader.db, chrom, 300); err != nil {
		t.Fatal(err)
	}

	posMin := uint32(50)
	posMax := uint32(200)
	res, err := repo.Query(ctx, QueryFilters{
		Chromosome:  chrom,
		PositionMin: &posMin,
		PositionMax: &posMax,
		Gene:        "TP53",
		Page:        1,
		PageSize:    25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) == 0 {
		t.Fatal("expected rows")
	}
	if res.Total == 0 {
		t.Fatal("expected total")
	}
	for _, row := range res.Rows {
		if row.Chromosome != chrom {
			t.Fatalf("chrom %q", row.Chromosome)
		}
		if row.Position < posMin || row.Position > posMax {
			t.Fatalf("position %d out of range", row.Position)
		}
		if !strings.EqualFold(row.Gene, "TP53") {
			t.Fatalf("gene %q", row.Gene)
		}
	}
}

func TestQueryRepository_P95Under500ms(t *testing.T) {
	if testing.Short() || os.Getenv("CI") != "" {
		t.Skip("performance threshold is environment-dependent")
	}
	dsn := strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN"))
	if dsn == "" {
		t.Skip("CLICKHOUSE_DSN not set")
	}
	ctx := context.Background()
	loader, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = loader.Close() }()
	if err := loader.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	repo, err := NewQueryRepository(loader.db)
	if err != nil {
		t.Fatal(err)
	}

	chrom := "chr14perf_" + uuid.NewString()[:8]
	if err := seedVariantsForTests(ctx, loader.db, chrom, 5000); err != nil {
		t.Fatal(err)
	}
	const samples = 200
	lat := make([]time.Duration, 0, samples)
	for i := 0; i < samples; i++ {
		start := time.Now()
		_, err := repo.Query(ctx, QueryFilters{
			Chromosome: chrom,
			Page:       1,
			PageSize:   50,
		})
		if err != nil {
			t.Fatal(err)
		}
		lat = append(lat, time.Since(start))
	}
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	p95 := lat[int(float64(len(lat))*0.95)]
	if p95 > 500*time.Millisecond {
		t.Fatalf("p95 latency %s exceeds 500ms", p95)
	}
}

func seedVariantsForTests(ctx context.Context, db *sql.DB, chrom string, rows int) error {
	parts := make([]string, 0, rows)
	args := make([]any, 0, rows*9)
	for i := 1; i <= rows; i++ {
		gene := "GENE=BRCA1"
		if i%2 == 0 {
			gene = "GENE=TP53"
		}
		parts = append(parts, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			"test_dataset",
			chrom,
			uint32(i),
			"A",
			"G",
			float64(50),
			"PASS",
			"DP=10;"+gene,
			uuid.NewString(),
		)
	}
	q := `
INSERT INTO variants (dataset_id, chrom, pos, ref, alt, qual, filter, info, source_key)
VALUES ` + strings.Join(parts, ",")
	// #nosec G202 -- placeholders are generated from constant tuple tokens only.
	_, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("seed variants: %w", err)
	}
	return nil
}
