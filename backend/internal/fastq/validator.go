// Package fastq validates FASTQ content using a streaming parser.
package fastq

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

const maxLineBytes = 1 << 20 // 1 MiB hard cap to keep memory bounded.

// Result captures validation outcome and details used for job output_ref.
type Result struct {
	Valid         bool   `json:"valid"`
	Records       int64  `json:"records"`
	FailureReason string `json:"failure_reason,omitempty"`
	FailureRecord int64  `json:"failure_record,omitempty"`
}

// ValidateStream validates FASTQ syntax from r without loading the full file in memory.
func ValidateStream(r io.Reader) (Result, error) {
	br := bufio.NewReaderSize(r, 64*1024)
	var rec int64
	for {
		h, ok, err := readLineBounded(br, maxLineBytes)
		if err != nil {
			return Result{}, err
		}
		if !ok {
			if rec == 0 {
				return Result{Valid: false, FailureReason: "empty_file"}, nil
			}
			return Result{Valid: true, Records: rec}, nil
		}

		rec++
		if !strings.HasPrefix(h, "@") {
			return Result{Valid: false, Records: rec - 1, FailureReason: "header_missing_at", FailureRecord: rec}, nil
		}

		seq, seqOK, err := readLineBounded(br, maxLineBytes)
		if err != nil {
			return Result{}, err
		}
		if !seqOK || seq == "" {
			return Result{Valid: false, Records: rec - 1, FailureReason: "missing_sequence", FailureRecord: rec}, nil
		}

		plus, plusOK, err := readLineBounded(br, maxLineBytes)
		if err != nil {
			return Result{}, err
		}
		if !plusOK || !strings.HasPrefix(plus, "+") {
			return Result{Valid: false, Records: rec - 1, FailureReason: "plus_missing", FailureRecord: rec}, nil
		}

		qual, qualOK, err := readLineBounded(br, maxLineBytes)
		if err != nil {
			return Result{}, err
		}
		if !qualOK {
			return Result{Valid: false, Records: rec - 1, FailureReason: "missing_quality", FailureRecord: rec}, nil
		}
		if len(seq) != len(qual) {
			return Result{Valid: false, Records: rec - 1, FailureReason: "quality_length_mismatch", FailureRecord: rec}, nil
		}
	}
}

func readLineBounded(r *bufio.Reader, limit int) (string, bool, error) {
	var total int
	var b strings.Builder
	for {
		part, err := r.ReadSlice('\n')
		total += len(part)
		if total > limit {
			return "", false, errors.New("fastq line exceeds maximum allowed length")
		}
		if len(part) > 0 {
			b.Write(part)
		}
		switch {
		case err == nil:
			line := strings.TrimRight(b.String(), "\r\n")
			return line, true, nil
		case errors.Is(err, bufio.ErrBufferFull):
			continue
		case errors.Is(err, io.EOF):
			if total == 0 {
				return "", false, nil
			}
			line := strings.TrimRight(b.String(), "\r\n")
			return line, true, nil
		default:
			return "", false, err
		}
	}
}
