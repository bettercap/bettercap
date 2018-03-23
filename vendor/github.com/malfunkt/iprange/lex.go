package iprange

import (
	"bytes"
	"errors"
	"log"
	"strconv"
	"unicode/utf8"
)

const eof = 0

type ipLex struct {
	line   []byte
	peek   rune
	output AddressRangeList
	err    error
}

func (ip *ipLex) Lex(yylval *ipSymType) int {
	for {
		c := ip.next()
		switch c {
		case eof:
			return eof
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return ip.byte(c, yylval)
		default:
			return int(c)
		}
	}
}

func (ip *ipLex) byte(c rune, yylval *ipSymType) int {
	add := func(b *bytes.Buffer, c rune) {
		if _, err := b.WriteRune(c); err != nil {
			log.Fatalf("WriteRune: %s", err)
		}
	}
	var b bytes.Buffer
	add(&b, c)
L:
	for {
		c = ip.next()
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			add(&b, c)
		default:
			break L
		}
	}
	if c != eof {
		ip.peek = c
	}
	octet, err := strconv.ParseUint(b.String(), 10, 32)
	if err != nil {
		log.Printf("badly formatted octet")
		return eof
	}
	yylval.num = byte(octet)
	return num
}

func (ip *ipLex) next() rune {
	if ip.peek != eof {
		r := ip.peek
		ip.peek = eof
		return r
	}
	if len(ip.line) == 0 {
		return eof
	}
	c, size := utf8.DecodeRune(ip.line)
	ip.line = ip.line[size:]
	if c == utf8.RuneError && size == 1 {
		log.Print("invalid utf8")
		return ip.next()
	}
	return c
}

func (ip *ipLex) Error(s string) {
	ip.err = errors.New(s)
}
