package fastq

import (
	"strings"
	"testing"
)

func TestValidateStream(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want Result
	}{
		{
			name: "valid single record",
			in:   "@r1\nACTG\n+\nIIII\n",
			want: Result{Valid: true, Records: 1},
		},
		{
			name: "valid two records",
			in:   "@r1\nACTG\n+\nIIII\n@r2\nNN\n+\n!!\n",
			want: Result{Valid: true, Records: 2},
		},
		{
			name: "empty file",
			in:   "",
			want: Result{Valid: false, FailureReason: "empty_file"},
		},
		{
			name: "missing at header",
			in:   "r1\nACTG\n+\nIIII\n",
			want: Result{Valid: false, FailureReason: "header_missing_at", FailureRecord: 1},
		},
		{
			name: "quality mismatch",
			in:   "@r1\nACTG\n+\nIII\n",
			want: Result{Valid: false, FailureReason: "quality_length_mismatch", FailureRecord: 1},
		},
		{
			name: "missing quality line",
			in:   "@r1\nACTG\n+\n",
			want: Result{Valid: false, FailureReason: "missing_quality", FailureRecord: 1},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ValidateStream(strings.NewReader(tt.in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got=%+v want=%+v", got, tt.want)
			}
		})
	}
}
