package main

import "testing"

func TestJaroWinkler(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want func(float64) bool
		desc string
	}{
		{
			name: "identical",
			a:    "Budi Santoso",
			b:    "Budi Santoso",
			want: func(s float64) bool { return s == 1.0 },
			desc: "identical names must be 1.0",
		},
		{
			name: "identical case and whitespace",
			a:    "  budi santoso ",
			b:    "Budi Santoso",
			want: func(s float64) bool { return s == 1.0 },
			desc: "normalization must ignore case and surrounding whitespace",
		},
		{
			name: "surname typo",
			a:    "Budi Santoso",
			b:    "Budi Santosa",
			want: func(s float64) bool { return s > 0.9 },
			desc: "one-character surname typo must stay above 0.9",
		},
		{
			name: "unrelated",
			a:    "Budi Santoso",
			b:    "Siti Rahmawati",
			want: func(s float64) bool { return s < 0.6 },
			desc: "unrelated names must be below 0.6",
		},
		{
			name: "empty left",
			a:    "",
			b:    "Budi Santoso",
			want: func(s float64) bool { return s == 0 },
			desc: "empty string must score 0",
		},
		{
			name: "empty both",
			a:    "",
			b:    "",
			want: func(s float64) bool { return s == 0 },
			desc: "two empty strings must score 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := jaroWinkler(tc.a, tc.b)
			if !tc.want(got) {
				t.Fatalf("jaroWinkler(%q, %q) = %.6f; %s", tc.a, tc.b, got, tc.desc)
			}
			if reverse := jaroWinkler(tc.b, tc.a); !tc.want(reverse) {
				t.Fatalf("jaroWinkler is not symmetric for %q/%q: forward %.6f, reverse %.6f", tc.a, tc.b, got, reverse)
			}
		})
	}
}

func TestTokenDiff(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "identical", a: "Budi Santoso", b: "Budi Santoso", want: 0},
		{name: "identical different order", a: "Budi Santoso", b: "Santoso Budi", want: 0},
		{name: "case insensitive", a: "BUDI santoso", b: "budi SANTOSO", want: 0},
		{name: "surname typo", a: "Budi Santoso", b: "Budi Santosa", want: 2},
		{name: "one added token", a: "Budi Santoso", b: "Budi Santoso Wijaya", want: 1},
		{name: "fully disjoint", a: "Budi Santoso", b: "Siti Rahmawati", want: 4},
		{name: "empty left", a: "", b: "Budi Santoso", want: 2},
		{name: "duplicate tokens collapse", a: "Budi Budi", b: "Budi", want: 0},
		{name: "extra whitespace", a: "  Budi   Santoso  ", b: "Budi Santoso", want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tokenDiff(tc.a, tc.b); got != tc.want {
				t.Fatalf("tokenDiff(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
