package network

import (
	"errors"
	"reflect"
	"testing"
)

func TestProcessSupportedFrequencies(t *testing.T) {
	// Actually test processSupportedFrequencies; IO is lifted out.
	cases := []struct {
		Name          string
		InputString   string
		InputError    error
		ExpectedFreqs []int
		ExpectedError bool
	}{
		{
			"Returns empty with an error",
			"Shouldn't matter",
			errors.New("iwlist must have failed"),
			[]int{},
			true,
		},
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			freqs, err := processSupportedFrequencies(test.InputString, test.InputError)
			if err != nil && !test.ExpectedError {
				t.Errorf("unexpected error: %s", err)
			}
			if err == nil && test.ExpectedError {
				t.Error("expected error, but got none")
			}
			if !test.ExpectedError && !reflect.DeepEqual(freqs, test.ExpectedFreqs) {
				t.Errorf("got %v, want %v", freqs, test.ExpectedFreqs)
			}
		})
	}
}
