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
			"Returns appropriately formatted frequencies on valid input",
			`wlan1     11 channels in total; available frequencies :
			Channel 01 : 2.412 GHz
			Channel 02 : 2.417 GHz
			Channel 03 : 2.422 GHz
			Channel 04 : 2.427 GHz
			Channel 05 : 2.432 GHz
			Channel 06 : 2.437 GHz
			Channel 07 : 2.442 GHz
			Channel 08 : 2.447 GHz
			Channel 09 : 2.452 GHz
			Channel 10 : 2.457 GHz
			Channel 11 : 2.462 GHz
			Current Frequency:2.437 GHz (Channel 6)`,
			nil,
			[]int{2412, 2417, 2422, 2427, 2432, 2437, 2442, 2447, 2452, 2457, 2462},
			false,
		},
		{
			"Returns empty with an error",
			"Doesn't matter",
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
