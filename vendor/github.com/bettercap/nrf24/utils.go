package nrf24

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var addrParser = regexp.MustCompile(`(?i)^[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}$`)

func NthChannel(idx int) int {
	return idx % TopChannel
}

func LoopChannels(idx *int) int {
	ch := NthChannel(*idx)
	*idx++
	return ch
}

func ConvertAddress(address string) (error, []byte) {
	if address == "" {
		return fmt.Errorf("no address provided"), nil
	} else if addrParser.MatchString(address) == false {
		return fmt.Errorf("address '%s' is not in the form XX:XX:XX:XX:XX", address), nil
	}

	// remove ':', decode as hex and reverse the bytes order
	clean := strings.Replace(address, ":", "", -1)
	raw, err := hex.DecodeString(clean)
	if err != nil {
		return err, nil
	} else if len(raw) != 5 {
		return fmt.Errorf("address must be composed of 5 octets"), nil
	}
	for i := 5/2 - 1; i >= 0; i-- {
		opp := 4 - i
		raw[i], raw[opp] = raw[opp], raw[i]
	}

	return nil, raw
}
