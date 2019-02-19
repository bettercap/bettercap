// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2016 the gousb Authors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gousb

import (
	"fmt"
)

// BCD is a binary-coded decimal version number. Its first 8 bits represent
// the major version number, its last 8 bits represent the minor version number.
// Major and minor are composed of 4+4 bits, where each 4 bits represents
// a decimal digit.
// Example: BCD(0x1234) means major 12 (decimal) and minor 34 (decimal).
type BCD uint16

// Major is the major number of the BCD.
func (s BCD) Major() uint8 {
	maj := uint8(s >> 8)
	return 10*(maj>>4) + maj&0x0f
}

// Minor is the minor number of the BCD.
func (s BCD) Minor() uint8 {
	min := uint8(s & 0xff)
	return 10*(min>>4) + min&0x0f
}

// String returns a dotted representation of the BCD (major.minor).
func (s BCD) String() string {
	return fmt.Sprintf("%d.%02d", s.Major(), s.Minor())
}

// Version returns a BCD version number with given major/minor.
func Version(major, minor uint8) BCD {
	return (BCD(major)/10)<<12 | (BCD(major)%10)<<8 | (BCD(minor)/10)<<4 | BCD(minor)%10
}

// ID represents a vendor or product ID.
type ID uint16

// String returns a hexadecimal ID.
func (id ID) String() string {
	return fmt.Sprintf("%04x", int(id))
}
