package base64vlq_test

import (
	"bytes"
	"testing"

	"gopkg.in/sourcemap.v1/base64vlq"
)

func TestEncodeDecode(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := base64vlq.NewEncoder(buf)
	dec := base64vlq.NewDecoder(buf)

	for n := -1000; n < 1000; n++ {
		if err := enc.Encode(n); err != nil {
			panic(err)
		}
	}

	for n := -1000; n < 1000; n++ {
		nn, err := dec.Decode()
		if err != nil {
			panic(err)
		}

		if nn != n {
			t.Errorf("%d != %d", nn, n)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	buf := new(bytes.Buffer)
	enc := base64vlq.NewEncoder(buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := enc.Encode(1000); err != nil {
			panic(err)
		}
	}
}

func BenchmarkEncodeDecode(b *testing.B) {
	buf := new(bytes.Buffer)
	enc := base64vlq.NewEncoder(buf)
	dec := base64vlq.NewDecoder(buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := enc.Encode(1000); err != nil {
			panic(err)
		}

		_, err := dec.Decode()
		if err != nil {
			panic(err)
		}
	}
}
