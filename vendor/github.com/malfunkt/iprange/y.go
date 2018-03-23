//line ip.y:2
package iprange

import __yyfmt__ "fmt"

//line ip.y:3
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

//line ip.y:26
type ipSymType struct {
	yys       int
	num       byte
	octRange  octetRange
	addrRange AddressRange
	result    AddressRangeList
}

const num = 57346

var ipToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"num",
	"','",
	"' '",
	"'/'",
	"'.'",
	"'*'",
	"'-'",
}
var ipStatenames = [...]string{}

const ipEofCode = 1
const ipErrCode = 2
const ipInitialStackSize = 16

//line ip.y:88

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

//line yacctab:1
var ipExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const ipNprod = 12
const ipPrivate = 57344

var ipTokenNames []string
var ipStates []string

const ipLast = 22

var ipAct = [...]int{

	4, 5, 12, 20, 2, 10, 6, 18, 11, 14,
	9, 17, 16, 13, 15, 8, 1, 7, 3, 19,
	0, 21,
}
var ipPact = [...]int{

	-3, 5, -1000, -2, 0, -8, -1000, -1000, -3, 3,
	10, -3, 7, -1000, -1000, -1000, -1, -1000, -3, -5,
	-3, -1000,
}
var ipPgo = [...]int{

	0, 18, 4, 0, 17, 16, 15,
}
var ipR1 = [...]int{

	0, 5, 5, 6, 6, 2, 2, 1, 3, 3,
	3, 4,
}
var ipR2 = [...]int{

	0, 1, 3, 1, 2, 3, 1, 7, 1, 1,
	1, 3,
}
var ipChk = [...]int{

	-1000, -5, -2, -1, -3, 4, 9, -4, -6, 5,
	7, 8, 10, -2, 6, 4, -3, 4, 8, -3,
	8, -3,
}
var ipDef = [...]int{

	0, -2, 1, 6, 0, 8, 9, 10, 0, 3,
	0, 0, 0, 2, 4, 5, 0, 11, 0, 0,
	0, 7,
}
var ipTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 6, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 9, 3, 5, 10, 8, 7,
}
var ipTok2 = [...]int{

	2, 3, 4,
}
var ipTok3 = [...]int{
	0,
}

var ipErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	ipDebug        = 0
	ipErrorVerbose = false
)

type ipLexer interface {
	Lex(lval *ipSymType) int
	Error(s string)
}

type ipParser interface {
	Parse(ipLexer) int
	Lookahead() int
}

type ipParserImpl struct {
	lval  ipSymType
	stack [ipInitialStackSize]ipSymType
	char  int
}

func (p *ipParserImpl) Lookahead() int {
	return p.char
}

func ipNewParser() ipParser {
	return &ipParserImpl{}
}

const ipFlag = -1000

func ipTokname(c int) string {
	if c >= 1 && c-1 < len(ipToknames) {
		if ipToknames[c-1] != "" {
			return ipToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func ipStatname(s int) string {
	if s >= 0 && s < len(ipStatenames) {
		if ipStatenames[s] != "" {
			return ipStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func ipErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !ipErrorVerbose {
		return "syntax error"
	}

	for _, e := range ipErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + ipTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := ipPact[state]
	for tok := TOKSTART; tok-1 < len(ipToknames); tok++ {
		if n := base + tok; n >= 0 && n < ipLast && ipChk[ipAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if ipDef[state] == -2 {
		i := 0
		for ipExca[i] != -1 || ipExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; ipExca[i] >= 0; i += 2 {
			tok := ipExca[i]
			if tok < TOKSTART || ipExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if ipExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += ipTokname(tok)
	}
	return res
}

func iplex1(lex ipLexer, lval *ipSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = ipTok1[0]
		goto out
	}
	if char < len(ipTok1) {
		token = ipTok1[char]
		goto out
	}
	if char >= ipPrivate {
		if char < ipPrivate+len(ipTok2) {
			token = ipTok2[char-ipPrivate]
			goto out
		}
	}
	for i := 0; i < len(ipTok3); i += 2 {
		token = ipTok3[i+0]
		if token == char {
			token = ipTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = ipTok2[1] /* unknown char */
	}
	if ipDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", ipTokname(token), uint(char))
	}
	return char, token
}

func ipParse(iplex ipLexer) int {
	return ipNewParser().Parse(iplex)
}

func (iprcvr *ipParserImpl) Parse(iplex ipLexer) int {
	var ipn int
	var ipVAL ipSymType
	var ipDollar []ipSymType
	_ = ipDollar // silence set and not used
	ipS := iprcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	ipstate := 0
	iprcvr.char = -1
	iptoken := -1 // iprcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		ipstate = -1
		iprcvr.char = -1
		iptoken = -1
	}()
	ipp := -1
	goto ipstack

ret0:
	return 0

ret1:
	return 1

ipstack:
	/* put a state and value onto the stack */
	if ipDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", ipTokname(iptoken), ipStatname(ipstate))
	}

	ipp++
	if ipp >= len(ipS) {
		nyys := make([]ipSymType, len(ipS)*2)
		copy(nyys, ipS)
		ipS = nyys
	}
	ipS[ipp] = ipVAL
	ipS[ipp].yys = ipstate

ipnewstate:
	ipn = ipPact[ipstate]
	if ipn <= ipFlag {
		goto ipdefault /* simple state */
	}
	if iprcvr.char < 0 {
		iprcvr.char, iptoken = iplex1(iplex, &iprcvr.lval)
	}
	ipn += iptoken
	if ipn < 0 || ipn >= ipLast {
		goto ipdefault
	}
	ipn = ipAct[ipn]
	if ipChk[ipn] == iptoken { /* valid shift */
		iprcvr.char = -1
		iptoken = -1
		ipVAL = iprcvr.lval
		ipstate = ipn
		if Errflag > 0 {
			Errflag--
		}
		goto ipstack
	}

ipdefault:
	/* default state action */
	ipn = ipDef[ipstate]
	if ipn == -2 {
		if iprcvr.char < 0 {
			iprcvr.char, iptoken = iplex1(iplex, &iprcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if ipExca[xi+0] == -1 && ipExca[xi+1] == ipstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			ipn = ipExca[xi+0]
			if ipn < 0 || ipn == iptoken {
				break
			}
		}
		ipn = ipExca[xi+1]
		if ipn < 0 {
			goto ret0
		}
	}
	if ipn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			iplex.Error(ipErrorMessage(ipstate, iptoken))
			Nerrs++
			if ipDebug >= 1 {
				__yyfmt__.Printf("%s", ipStatname(ipstate))
				__yyfmt__.Printf(" saw %s\n", ipTokname(iptoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for ipp >= 0 {
				ipn = ipPact[ipS[ipp].yys] + ipErrCode
				if ipn >= 0 && ipn < ipLast {
					ipstate = ipAct[ipn] /* simulate a shift of "error" */
					if ipChk[ipstate] == ipErrCode {
						goto ipstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if ipDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", ipS[ipp].yys)
				}
				ipp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if ipDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", ipTokname(iptoken))
			}
			if iptoken == ipEofCode {
				goto ret1
			}
			iprcvr.char = -1
			iptoken = -1
			goto ipnewstate /* try again in the same state */
		}
	}

	/* reduction by production ipn */
	if ipDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", ipn, ipStatname(ipstate))
	}

	ipnt := ipn
	ippt := ipp
	_ = ippt // guard against "declared and not used"

	ipp -= ipR2[ipn]
	// ipp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if ipp+1 >= len(ipS) {
		nyys := make([]ipSymType, len(ipS)*2)
		copy(nyys, ipS)
		ipS = nyys
	}
	ipVAL = ipS[ipp+1]

	/* consult goto table to find next state */
	ipn = ipR1[ipn]
	ipg := ipPgo[ipn]
	ipj := ipg + ipS[ipp].yys + 1

	if ipj >= ipLast {
		ipstate = ipAct[ipg]
	} else {
		ipstate = ipAct[ipj]
		if ipChk[ipstate] != -ipn {
			ipstate = ipAct[ipg]
		}
	}
	// dummy call; replaced with literal code
	switch ipnt {

	case 1:
		ipDollar = ipS[ippt-1 : ippt+1]
		//line ip.y:41
		{
			ipVAL.result = append(ipVAL.result, ipDollar[1].addrRange)
			iplex.(*ipLex).output = ipVAL.result
		}
	case 2:
		ipDollar = ipS[ippt-3 : ippt+1]
		//line ip.y:46
		{
			ipVAL.result = append(ipDollar[1].result, ipDollar[3].addrRange)
			iplex.(*ipLex).output = ipVAL.result
		}
	case 5:
		ipDollar = ipS[ippt-3 : ippt+1]
		//line ip.y:54
		{
			mask := net.CIDRMask(int(ipDollar[3].num), 32)
			min := ipDollar[1].addrRange.Min.Mask(mask)
			maxInt := binary.BigEndian.Uint32([]byte(min)) +
				0xffffffff -
				binary.BigEndian.Uint32([]byte(mask))
			maxBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(maxBytes, maxInt)
			maxBytes = maxBytes[len(maxBytes)-4:]
			max := net.IP(maxBytes)
			ipVAL.addrRange = AddressRange{
				Min: min.To4(),
				Max: max.To4(),
			}
		}
	case 6:
		ipDollar = ipS[ippt-1 : ippt+1]
		//line ip.y:70
		{
			ipVAL.addrRange = ipDollar[1].addrRange
		}
	case 7:
		ipDollar = ipS[ippt-7 : ippt+1]
		//line ip.y:75
		{
			ipVAL.addrRange = AddressRange{
				Min: net.IPv4(ipDollar[1].octRange.min, ipDollar[3].octRange.min, ipDollar[5].octRange.min, ipDollar[7].octRange.min).To4(),
				Max: net.IPv4(ipDollar[1].octRange.max, ipDollar[3].octRange.max, ipDollar[5].octRange.max, ipDollar[7].octRange.max).To4(),
			}
		}
	case 8:
		ipDollar = ipS[ippt-1 : ippt+1]
		//line ip.y:82
		{
			ipVAL.octRange = octetRange{ipDollar[1].num, ipDollar[1].num}
		}
	case 9:
		ipDollar = ipS[ippt-1 : ippt+1]
		//line ip.y:83
		{
			ipVAL.octRange = octetRange{0, 255}
		}
	case 10:
		ipDollar = ipS[ippt-1 : ippt+1]
		//line ip.y:84
		{
			ipVAL.octRange = ipDollar[1].octRange
		}
	case 11:
		ipDollar = ipS[ippt-3 : ippt+1]
		//line ip.y:86
		{
			ipVAL.octRange = octetRange{ipDollar[1].num, ipDollar[3].num}
		}
	}
	goto ipstack /* stack new state and value */
}
