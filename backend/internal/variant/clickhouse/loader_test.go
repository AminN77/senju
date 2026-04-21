package clickhouse

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestParseLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		line         string
		wantErr      bool
		wantVariants int
		wantChrom    string
		wantPos      uint32
		wantFirstAlt string
		wantNilQual  bool
		wantQual     float64
	}{
		{
			name:         "valid multi alt",
			line:         "chr1\t123\t.\tA\tC,G\t99.5\tPASS\tDP=42",
			wantVariants: 2,
			wantChrom:    "chr1",
			wantPos:      123,
			wantFirstAlt: "C",
			wantQual:     99.5,
		},
		{
			name:         "valid single alt and nil qual",
			line:         "chr2\t20\t.\tG\tT\t.\tPASS\tDP=5",
			wantVariants: 1,
			wantChrom:    "chr2",
			wantPos:      20,
			wantFirstAlt: "T",
			wantNilQual:  true,
		},
		{
			name:         "empty alt entries are skipped",
			line:         "chr3\t31\t.\tC\tA,,G\t12.0\tPASS\tDP=8",
			wantVariants: 2,
			wantChrom:    "chr3",
			wantPos:      31,
			wantFirstAlt: "A",
			wantQual:     12,
		},
		{
			name:    "invalid pos",
			line:    "chr1\tbad\t.\tA\tC\t10\tPASS\tDP=1",
			wantErr: true,
		},
		{
			name:    "invalid qual",
			line:    "chr1\t10\t.\tA\tC\tnot-a-float\tPASS\tDP=1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vars, err := parseLine("ds1", tt.line)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected parse error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(vars) != tt.wantVariants {
				t.Fatalf("variants=%d want=%d", len(vars), tt.wantVariants)
			}
			if len(vars) == 0 {
				return
			}
			if vars[0].DatasetID != "ds1" || vars[0].Chrom != tt.wantChrom || vars[0].Pos != tt.wantPos {
				t.Fatalf("variant %+v", vars[0])
			}
			if vars[0].Alt != tt.wantFirstAlt {
				t.Fatalf("alt=%q want=%q", vars[0].Alt, tt.wantFirstAlt)
			}
			if tt.wantNilQual {
				if vars[0].Qual != nil {
					t.Fatalf("qual=%+v want=nil", vars[0].Qual)
				}
				return
			}
			if vars[0].Qual == nil || *vars[0].Qual != tt.wantQual {
				t.Fatalf("qual=%+v want=%f", vars[0].Qual, tt.wantQual)
			}
		})
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
	datasetID := "issue13-test-" + uuid.NewString()

	vcf := "##fileformat=VCFv4.2\n#CHROM\tPOS\tID\tREF\tALT\tQUAL\tFILTER\tINFO\nchr1\t10\t.\tA\tC\t50\tPASS\tDP=10\nchr1\t10\t.\tA\tC\t50\tPASS\tDP=10\nchr2\t20\t.\tG\tT\t.\tPASS\tDP=5\n"
	n1, err := l.LoadVCF(ctx, datasetID, strings.NewReader(vcf))
	if err != nil {
		t.Fatal(err)
	}
	n2, err := l.LoadVCF(ctx, datasetID, strings.NewReader(vcf))
	if err != nil {
		t.Fatal(err)
	}
	if n1 == 0 {
		t.Fatalf("expected inserts in first load, got %d", n1)
	}
	if n2 != 0 {
		t.Fatalf("expected zero inserts on idempotent rerun, got %d", n2)
	}
	var count int
	if err := l.db.QueryRowContext(ctx, "SELECT count() FROM variants WHERE dataset_id = ?", datasetID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count=%d want=2", count)
	}
}

func BenchmarkLoadVCF_ParseOnlyThroughput(b *testing.B) {
	sample := "chr1\t100\t.\tA\tG\t99.9\tPASS\tDP=10\n"
	input := "##fileformat=VCFv4.2\n#CHROM\tPOS\tID\tREF\tALT\tQUAL\tFILTER\tINFO\n" + strings.Repeat(sample, 5000)
	lines := strings.Split(input, "\n")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := make([]Variant, 0, 5000)
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

func TestNewVCFScanner_LargeLineWithinConfiguredLimit(t *testing.T) {
	t.Parallel()
	largeInfo := strings.Repeat("A", scannerBufferBytes/2)
	line := "chr1\t1\t.\tA\tT\t50\tPASS\t" + largeInfo + "\n"
	sc := newVCFScanner(strings.NewReader(line))
	if !sc.Scan() {
		t.Fatalf("scan failed: %v", sc.Err())
	}
	if got := sc.Text(); got == "" {
		t.Fatal("expected scanned line")
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
