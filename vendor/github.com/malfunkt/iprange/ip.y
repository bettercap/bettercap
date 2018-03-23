%{

package iprange

import (
    "encoding/binary"
    "net"

    "github.com/pkg/errors"
)

type AddressRangeList []AddressRange

type AddressRange struct {
    Min net.IP
    Max net.IP
}

type octetRange struct {
    min byte
    max byte
}

%}

%union {
    num         byte
    octRange    octetRange
    addrRange   AddressRange
    result      AddressRangeList
}

%token  <num> num
%type   <addrRange> address target
%type   <octRange>  term octet_range
%type   <result>    result

%%

result: target
            {
                $$ = append($$, $1)
                iplex.(*ipLex).output = $$
            }
      | result comma target
            {
                $$ = append($1, $3)
                iplex.(*ipLex).output = $$
            }

comma: ',' | ',' ' '

target:     address '/' num
                {
                    mask := net.CIDRMask(int($3), 32)
                    min := $1.Min.Mask(mask)
                    maxInt := binary.BigEndian.Uint32([]byte(min)) +
                                0xffffffff -
                                binary.BigEndian.Uint32([]byte(mask))
                    maxBytes := make([]byte, 4)
                    binary.BigEndian.PutUint32(maxBytes, maxInt)
                    maxBytes = maxBytes[len(maxBytes)-4:]
                    max := net.IP(maxBytes)
                    $$ = AddressRange {
                        Min: min.To4(),
                        Max: max.To4(),
                    }
                }
      |     address
                {
                    $$ = $1
                }

address:    term '.' term '.' term '.' term
                {
                    $$ = AddressRange {
                        Min: net.IPv4($1.min, $3.min, $5.min, $7.min).To4(),
                        Max: net.IPv4($1.max, $3.max, $5.max, $7.max).To4(),
                    }
                }

term:   num         { $$ = octetRange { $1, $1 } }
    |   '*'         { $$ = octetRange { 0, 255 } }
    |   octet_range { $$ = $1 }

octet_range:    num '-' num { $$ = octetRange { $1, $3 } }

%%

// ParseList takes a list of target specifications and returns a list of ranges,
// even if the list contains a single element.
func ParseList(in string) (AddressRangeList, error) {
    lex := &ipLex{line: []byte(in)}
    errCode := ipParse(lex)
    if errCode != 0 || lex.err != nil {
        return nil, errors.Wrap(lex.err, "could not parse target")
    }
    return lex.output, nil
}

// Parse takes a single target specification and returns a range. It effectively calls ParseList
// and returns the first result
func Parse(in string) (*AddressRange, error) {
    l, err := ParseList(in)
    if err != nil {
        return nil, err
    }
    return &l[0], nil
}
