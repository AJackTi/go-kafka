package mysql

import "testing"

func TestNormalizeDSN(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		want  string
	}{
		"driver DSN": {
			input: "mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
			want:  "mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
		},
		"migration URL": {
			input: "mysql://mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
			want:  "mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
		},
		"surrounding whitespace": {
			input: "  mysql://mysql:password@tcp(localhost:3306)/go_kafka  ",
			want:  "mysql:password@tcp(localhost:3306)/go_kafka",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeDSN(test.input); got != test.want {
				t.Fatalf("normalizeDSN(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
