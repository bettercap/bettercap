package sourcemap_test

import (
	"testing"

	"gopkg.in/sourcemap.v1"
)

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := sourcemap.Parse(jqSourceMapURL, jqSourceMapBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSource(b *testing.B) {
	smap, err := sourcemap.Parse(jqSourceMapURL, jqSourceMapBytes)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			smap.Source(j, 100*j)
		}
	}
}
