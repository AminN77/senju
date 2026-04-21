package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// QueryFilters controls variant search predicates and pagination.
type QueryFilters struct {
	Chromosome  string
	PositionMin *uint32
	PositionMax *uint32
	Gene        string
	Page        int
	PageSize    int
}

// QueryRow is one variant returned to the API layer.
type QueryRow struct {
	Chromosome string
	Position   uint32
	Ref        string
	Alt        string
	Qual       *float64
	Filter     string
	Info       string
	Gene       string
}

// QueryResult is a paginated variant page.
type QueryResult struct {
	Rows     []QueryRow
	Total    int64
	Page     int
	PageSize int
	HasNext  bool
}

// QueryService is consumed by HTTP handlers for variant retrieval.
type QueryService interface {
	Query(context.Context, QueryFilters) (QueryResult, error)
}

// QueryRepository executes variant queries against ClickHouse.
type QueryRepository struct {
	db     *sql.DB
	ownsDB bool
}

// NewQueryRepository builds a query repository from an existing DB handle.
func NewQueryRepository(db *sql.DB) (*QueryRepository, error) {
	if db == nil {
		return nil, errors.New("clickhouse query repository: db is nil")
	}
	return &QueryRepository{db: db, ownsDB: false}, nil
}

// OpenQueryRepository creates a query repository from a ClickHouse DSN.
func OpenQueryRepository(dsn string) (*QueryRepository, error) {
	l, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	return &QueryRepository{db: l.db, ownsDB: true}, nil
}

// Close releases the underlying ClickHouse connection pool.
func (r *QueryRepository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	if !r.ownsDB {
		return nil
	}
	return r.db.Close()
}

// Query returns a filtered/paginated variants page.
func (r *QueryRepository) Query(ctx context.Context, f QueryFilters) (QueryResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}
	whereSQL, args := buildWhere(f)

	var total int64
	countSQL := "SELECT count() FROM variants" + whereSQL
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return QueryResult{}, err
	}

	offset := (f.Page - 1) * f.PageSize
	// #nosec G202 -- whereSQL is assembled from constant clauses with parameter placeholders only.
	q := `
SELECT chrom, pos, ref, alt, qual, filter, info
FROM variants` + whereSQL + `
ORDER BY chrom, pos, ref, alt
LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), f.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, q, queryArgs...)
	if err != nil {
		return QueryResult{}, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]QueryRow, 0, f.PageSize)
	for rows.Next() {
		var rr QueryRow
		var qual sql.NullFloat64
		if err := rows.Scan(&rr.Chromosome, &rr.Position, &rr.Ref, &rr.Alt, &qual, &rr.Filter, &rr.Info); err != nil {
			return QueryResult{}, err
		}
		if qual.Valid {
			qv := qual.Float64
			rr.Qual = &qv
		}
		rr.Gene = extractGene(rr.Info)
		out = append(out, rr)
	}
	if err := rows.Err(); err != nil {
		return QueryResult{}, err
	}

	return QueryResult{
		Rows:     out,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
		HasNext:  int64(offset+len(out)) < total,
	}, nil
}

func buildWhere(f QueryFilters) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if c := strings.TrimSpace(f.Chromosome); c != "" {
		clauses = append(clauses, "chrom = ?")
		args = append(args, c)
	}
	if f.PositionMin != nil {
		clauses = append(clauses, "pos >= ?")
		args = append(args, *f.PositionMin)
	}
	if f.PositionMax != nil {
		clauses = append(clauses, "pos <= ?")
		args = append(args, *f.PositionMax)
	}
	if g := strings.TrimSpace(f.Gene); g != "" {
		lg := strings.ToLower(g)
		clauses = append(clauses, "(lower(info) LIKE ? OR lower(info) LIKE ?)")
		args = append(args, "%gene="+lg+"%", "%gene_symbol="+lg+"%")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func extractGene(info string) string {
	parts := strings.Split(info, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		up := strings.ToUpper(p)
		if strings.HasPrefix(up, "GENE=") {
			return strings.TrimSpace(p[len("GENE="):])
		}
		if strings.HasPrefix(up, "GENE_SYMBOL=") {
			return strings.TrimSpace(p[len("GENE_SYMBOL="):])
		}
	}
	return ""
}
