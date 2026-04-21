// Package clickhouse provides ClickHouse variant schema and VCF loader.
package clickhouse

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2" // register database/sql clickhouse driver
)

const batchSize = 2000

const createTableSQL = `
CREATE TABLE IF NOT EXISTS variants (
    dataset_id String,
    chrom LowCardinality(String),
    pos UInt32,
    ref String,
    alt String,
    qual Nullable(Float64),
    filter String,
    info String,
    source_key String
)
ENGINE = ReplacingMergeTree
ORDER BY (dataset_id, chrom, pos, ref, alt)
`

// Variant is one parsed VCF variant row.
type Variant struct {
	DatasetID string
	Chrom     string
	Pos       uint32
	Ref       string
	Alt       string
	Qual      *float64
	Filter    string
	Info      string
	SourceKey string
}

// Loader ingests VCF data into ClickHouse.
type Loader struct {
	db *sql.DB
}

// Open returns a loader using a ClickHouse DSN.
func Open(dsn string) (*Loader, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("clickhouse loader: dsn is required")
	}
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}
	return &Loader{db: db}, nil
}

// NewWithDB returns a loader from an existing DB handle.
func NewWithDB(db *sql.DB) (*Loader, error) {
	if db == nil {
		return nil, errors.New("clickhouse loader: db is nil")
	}
	return &Loader{db: db}, nil
}

// Close closes the underlying DB.
func (l *Loader) Close() error {
	if l.db == nil {
		return nil
	}
	return l.db.Close()
}

// EnsureSchema creates the variants table if missing.
func (l *Loader) EnsureSchema(ctx context.Context) error {
	_, err := l.db.ExecContext(ctx, createTableSQL)
	return err
}

// LoadVCF parses and inserts variants from VCF stream.
// Duplicate protection is enforced by skipping existing variant keys.
func (l *Loader) LoadVCF(ctx context.Context, datasetID string, r io.Reader) (int, error) {
	if strings.TrimSpace(datasetID) == "" {
		return 0, errors.New("clickhouse loader: dataset_id is required")
	}
	if r == nil {
		return 0, errors.New("clickhouse loader: reader is nil")
	}
	if err := l.EnsureSchema(ctx); err != nil {
		return 0, err
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	inserted := 0
	seenInRun := make(map[string]struct{}, 4096)
	batch := make([]Variant, 0, batchSize)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		variants, err := parseLine(datasetID, line)
		if err != nil {
			return inserted, err
		}
		for _, v := range variants {
			if _, ok := seenInRun[v.SourceKey]; ok {
				continue
			}
			seenInRun[v.SourceKey] = struct{}{}
			batch = append(batch, v)
			if len(batch) >= batchSize {
				n, err := l.flushBatch(ctx, datasetID, batch)
				if err != nil {
					return inserted, err
				}
				inserted += n
				batch = batch[:0]
			}
		}
	}
	if err := sc.Err(); err != nil {
		return inserted, err
	}
	if len(batch) > 0 {
		n, err := l.flushBatch(ctx, datasetID, batch)
		if err != nil {
			return inserted, err
		}
		inserted += n
	}
	return inserted, nil
}

func (l *Loader) flushBatch(ctx context.Context, datasetID string, batch []Variant) (int, error) {
	existing, err := l.fetchExistingKeys(ctx, datasetID, batch)
	if err != nil {
		return 0, err
	}
	toInsert := make([]Variant, 0, len(batch))
	for _, v := range batch {
		if _, ok := existing[v.SourceKey]; ok {
			continue
		}
		toInsert = append(toInsert, v)
	}
	if len(toInsert) == 0 {
		return 0, nil
	}
	if err := l.insertBatch(ctx, toInsert); err != nil {
		return 0, err
	}
	return len(toInsert), nil
}

func (l *Loader) fetchExistingKeys(ctx context.Context, datasetID string, batch []Variant) (map[string]struct{}, error) {
	if len(batch) == 0 {
		return map[string]struct{}{}, nil
	}
	holders := make([]string, 0, len(batch))
	args := make([]any, 0, len(batch)+1)
	args = append(args, datasetID)
	for _, v := range batch {
		holders = append(holders, "?")
		args = append(args, v.SourceKey)
	}
	// #nosec G202 -- placeholders are generated from constant "?" tokens, values remain parameterized.
	q := `
SELECT source_key FROM variants
WHERE dataset_id = ? AND source_key IN (` + strings.Join(holders, ",") + `)`
	rows, err := l.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]struct{}, len(batch))
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		out[key] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (l *Loader) insertBatch(ctx context.Context, batch []Variant) error {
	holders := make([]string, 0, len(batch))
	args := make([]any, 0, len(batch)*9)
	for _, v := range batch {
		holders = append(holders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args, v.DatasetID, v.Chrom, v.Pos, v.Ref, v.Alt, v.Qual, v.Filter, v.Info, v.SourceKey)
	}
	// #nosec G202 -- VALUES placeholder tuples are generated from constant "(?, ...)" tokens only.
	q := `
INSERT INTO variants (dataset_id, chrom, pos, ref, alt, qual, filter, info, source_key)
VALUES ` + strings.Join(holders, ",")
	_, err := l.db.ExecContext(ctx, q, args...)
	return err
}

func parseLine(datasetID, line string) ([]Variant, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 8 {
		return nil, fmt.Errorf("invalid vcf row: expected >=8 cols, got %d", len(fields))
	}
	pos64, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid vcf pos: %w", err)
	}
	qual, err := parseQual(fields[5])
	if err != nil {
		return nil, err
	}
	alts := strings.Split(fields[4], ",")
	out := make([]Variant, 0, len(alts))
	for _, alt := range alts {
		alt = strings.TrimSpace(alt)
		if alt == "" {
			continue
		}
		key := makeKey(datasetID, fields[0], uint32(pos64), fields[3], alt)
		out = append(out, Variant{
			DatasetID: datasetID,
			Chrom:     fields[0],
			Pos:       uint32(pos64),
			Ref:       fields[3],
			Alt:       alt,
			Qual:      qual,
			Filter:    fields[6],
			Info:      fields[7],
			SourceKey: key,
		})
	}
	return out, nil
}

func parseQual(raw string) (*float64, error) {
	if raw == "." {
		return nil, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid vcf qual: %w", err)
	}
	return &v, nil
}

func makeKey(datasetID, chrom string, pos uint32, ref, alt string) string {
	h := sha256.Sum256([]byte(datasetID + "|" + chrom + "|" + strconv.FormatUint(uint64(pos), 10) + "|" + ref + "|" + alt))
	return fmt.Sprintf("%x", h[:])
}
