package main

import "testing"

func BenchmarkRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := run()
		if len(m) == 0 {
			b.Fatalf("no results")
		}
	}
}
