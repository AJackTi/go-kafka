package migration

import "testing"

func TestMigrationURLNormalizesMySQLDSN(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		want  string
	}{
		"raw dsn": {
			input: "mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
			want:  "mysql://mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
		},
		"already normalized": {
			input: "mysql://mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
			want:  "mysql://mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true",
		},
		"whitespace": {
			input: "  mysql:password@tcp(localhost:3306)/go_kafka  ",
			want:  "mysql://mysql:password@tcp(localhost:3306)/go_kafka",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := migrationURL(test.input); got != test.want {
				t.Fatalf("migrationURL(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
