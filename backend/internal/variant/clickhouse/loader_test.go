package clickhouse

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	t.Parallel()
	line := "chr1\t123\t.\tA\tC,G\t99.5\tPASS\tDP=42"
	vars, err := parseLine("ds1", line)
	if err != nil {
		t.Fatal(err)
	}
	if len(vars) != 2 {
		t.Fatalf("variants=%d", len(vars))
	}
	if vars[0].DatasetID != "ds1" || vars[0].Chrom != "chr1" || vars[0].Pos != 123 {
		t.Fatalf("variant %+v", vars[0])
	}
	if vars[0].Qual == nil || *vars[0].Qual != 99.5 {
		t.Fatalf("qual %+v", vars[0].Qual)
	}
}

func TestParseLine_Invalid(t *testing.T) {
	t.Parallel()
	if _, err := parseLine("ds1", "chr1\tbad"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadVCF_IdempotentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	dsn := strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN"))
	if dsn == "" {
		t.Skip("CLICKHOUSE_DSN not set")
	}
	l, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()

	ctx := context.Background()
	if err := l.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	_, err = l.db.ExecContext(ctx, "ALTER TABLE variants DELETE WHERE dataset_id = ?", "issue13-test")
	if err != nil {
		t.Fatal(err)
	}

	vcf := "##fileformat=VCFv4.2\n#CHROM\tPOS\tID\tREF\tALT\tQUAL\tFILTER\tINFO\nchr1\t10\t.\tA\tC\t50\tPASS\tDP=10\nchr1\t10\t.\tA\tC\t50\tPASS\tDP=10\nchr2\t20\t.\tG\tT\t.\tPASS\tDP=5\n"
	n1, err := l.LoadVCF(ctx, "issue13-test", strings.NewReader(vcf))
	if err != nil {
		t.Fatal(err)
	}
	n2, err := l.LoadVCF(ctx, "issue13-test", strings.NewReader(vcf))
	if err != nil {
		t.Fatal(err)
	}
	if n1 == 0 {
		t.Fatalf("expected inserts in first load, got %d", n1)
	}
	_ = n2
	var count int
	if err := l.db.QueryRowContext(ctx, "SELECT count() FROM variants WHERE dataset_id = ?", "issue13-test").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count=%d want=2", count)
	}
}

func BenchmarkLoadVCF_ParseOnlyThroughput(b *testing.B) {
	sample := "chr1\t100\t.\tA\tG\t99.9\tPASS\tDP=10\n"
	input := "##fileformat=VCFv4.2\n#CHROM\tPOS\tID\tREF\tALT\tQUAL\tFILTER\tINFO\n" + strings.Repeat(sample, 5000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]Variant, 0, 5000)
		lines := strings.Split(input, "\n")
		for _, line := range lines {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			v, err := parseLine("bench", line)
			if err != nil {
				b.Fatal(err)
			}
			s = append(s, v...)
		}
		if len(s) == 0 {
			b.Fatal("parsed zero variants")
		}
	}
}

func TestOpen_NilDSN(t *testing.T) {
	t.Parallel()
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty dsn")
	}
}

func TestNewWithDB_Nil(t *testing.T) {
	t.Parallel()
	if _, err := NewWithDB((*sql.DB)(nil)); err == nil {
		t.Fatal("expected nil db error")
	}
}
