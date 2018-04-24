package packets

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"unsafe"
)

const (
	NTLM_SIG_OFFSET  = 0
	NTLM_TYPE_OFFSET = 8

	NTLM_TYPE1_FLAGS_OFFSET   = 12
	NTLM_TYPE1_DOMAIN_OFFSET  = 16
	NTLM_TYPE1_WORKSTN_OFFSET = 24
	NTLM_TYPE1_DATA_OFFSET    = 32
	NTLM_TYPE1_MINSIZE        = 16

	NTLM_TYPE2_TARGET_OFFSET     = 12
	NTLM_TYPE2_FLAGS_OFFSET      = 20
	NTLM_TYPE2_CHALLENGE_OFFSET  = 24
	NTLM_TYPE2_CONTEXT_OFFSET    = 32
	NTLM_TYPE2_TARGETINFO_OFFSET = 40
	NTLM_TYPE2_DATA_OFFSET       = 48
	NTLM_TYPE2_MINSIZE           = 32

	NTLM_TYPE3_LMRESP_OFFSET     = 12
	NTLM_TYPE3_NTRESP_OFFSET     = 20
	NTLM_TYPE3_DOMAIN_OFFSET     = 28
	NTLM_TYPE3_USER_OFFSET       = 36
	NTLM_TYPE3_WORKSTN_OFFSET    = 44
	NTLM_TYPE3_SESSIONKEY_OFFSET = 52
	NTLM_TYPE3_FLAGS_OFFSET      = 60
	NTLM_TYPE3_DATA_OFFSET       = 64
	NTLM_TYPE3_MINSIZE           = 52

	NTLM_BUFFER_LEN_OFFSET    = 0
	NTLM_BUFFER_MAXLEN_OFFSET = 2
	NTLM_BUFFER_OFFSET_OFFSET = 4
	NTLM_BUFFER_SIZE          = 8

	NtlmV1 = 1
	NtlmV2 = 2
)

type NTLMChallengeResponse struct {
	Challenge string
	Response  string
}

type NTLMChallengeResponseParsed struct {
	Type            int
	ServerChallenge string
	User            string
	Domain          string
	LmHash          string
	NtHashOne       string
	NtHashTwo       string
}

type NTLMResponseHeader struct {
	Sig          string
	Type         uint32
	LmLen        uint16
	LmMax        uint16
	LmOffset     uint16
	NtLen        uint16
	NtMax        uint16
	NtOffset     uint16
	DomainLen    uint16
	DomainMax    uint16
	DomainOffset uint16
	UserLen      uint16
	UserMax      uint16
	UserOffset   uint16
	HostLen      uint16
	HostMax      uint16
	HostOffset   uint16
}

type NTLMState struct {
	sync.Mutex

	Responses map[uint32]string
	Pairs     []NTLMChallengeResponse
}

func (s *NTLMState) AddServerResponse(key uint32, value string) {
	s.Lock()
	defer s.Unlock()
	s.Responses[key] = value
}

func (s *NTLMState) AddClientResponse(seq uint32, value string, cb func(data NTLMChallengeResponseParsed)) {
	s.Lock()
	defer s.Unlock()

	if chall, found := s.Responses[seq]; found {
		pair := NTLMChallengeResponse{
			Challenge: chall,
			Response:  value,
		}
		s.Pairs = append(s.Pairs, pair)

		if data, err := pair.Parsed(); err == nil {
			cb(data)
		}
	}
}

func NewNTLMState() *NTLMState {
	return &NTLMState{
		Responses: make(map[uint32]string),
		Pairs:     make([]NTLMChallengeResponse, 0),
	}
}

func (sr NTLMChallengeResponse) getServerChallenge() string {
	dataCallenge := sr.getChallengeBytes()
	//offset to the challenge and the challenge is 8 bytes long
	return hex.EncodeToString(dataCallenge[NTLM_TYPE2_CHALLENGE_OFFSET : NTLM_TYPE2_CHALLENGE_OFFSET+8])
}

func (sr NTLMChallengeResponse) getChallengeBytes() []byte {
	dataCallenge, _ := base64.StdEncoding.DecodeString(sr.Challenge)
	return dataCallenge
}

func (sr NTLMChallengeResponse) getResponseBytes() []byte {
	dataResponse, _ := base64.StdEncoding.DecodeString(sr.Response)
	return dataResponse
}
func (sr *NTLMChallengeResponse) Parsed() (NTLMChallengeResponseParsed, error) {
	if sr.isNtlmV1() {
		return sr.ParsedNtLMv1()
	}
	return sr.ParsedNtLMv2()
}

func (sr *NTLMChallengeResponse) ParsedNtLMv2() (NTLMChallengeResponseParsed, error) {
	r := sr.getResponseHeader()
	if r.UserLen == 0 {
		return NTLMChallengeResponseParsed{}, errors.New("No repsponse data")
	}
	b := sr.getResponseBytes()
	nthash := b[r.NtOffset : r.NtOffset+r.NtLen]
	// each char in user and domain is null terminated
	return NTLMChallengeResponseParsed{
		Type:            NtlmV2,
		ServerChallenge: sr.getServerChallenge(),
		User:            strings.Replace(string(b[r.UserOffset:r.UserOffset+r.UserLen]), "\x00", "", -1),
		Domain:          strings.Replace(string(b[r.DomainOffset:r.DomainOffset+r.DomainLen]), "\x00", "", -1),
		NtHashOne:       hex.EncodeToString(nthash[:16]), // first part of the hash is 16 bytes
		NtHashTwo:       hex.EncodeToString(nthash[16:]),
	}, nil
}

func (sr NTLMChallengeResponse) isNtlmV1() bool {
	headerValues := sr.getResponseHeader()
	return headerValues.NtLen == 24
}

func (sr NTLMChallengeResponse) ParsedNtLMv1() (NTLMChallengeResponseParsed, error) {
	r := sr.getResponseHeader()
	if r.UserLen == 0 {
		return NTLMChallengeResponseParsed{}, errors.New("No repsponse data")
	}
	b := sr.getResponseBytes()
	// each char user and domain is null terminated
	return NTLMChallengeResponseParsed{
		Type:            NtlmV1,
		ServerChallenge: sr.getServerChallenge(),
		User:            strings.Replace(string(b[r.UserOffset:r.UserOffset+r.UserLen]), "\x00", "", -1),
		Domain:          strings.Replace(string(b[r.DomainOffset:r.DomainOffset+r.DomainLen]), "\x00", "", -1),
		LmHash:          hex.EncodeToString(b[r.LmOffset : r.LmOffset+r.LmLen]),
	}, nil
}

func is_le() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

func _uint32(b []byte, start, end int) uint32 {
	if is_le() {
		return binary.LittleEndian.Uint32(b[start:end])
	}
	return binary.BigEndian.Uint32(b[start:end])
}

func _uint16(b []byte, start, end int) uint16 {
	if is_le() {
		return binary.LittleEndian.Uint16(b[start:end])
	}
	return binary.BigEndian.Uint16(b[start:end])
}

func (sr NTLMChallengeResponse) getResponseHeader() NTLMResponseHeader {
	b := sr.getResponseBytes()
	if len(b) == 0 {
		return NTLMResponseHeader{}
	}
	return NTLMResponseHeader{
		Sig:          strings.Replace(string(b[NTLM_SIG_OFFSET:NTLM_SIG_OFFSET+8]), "\x00", "", -1),
		Type:         _uint32(b, NTLM_TYPE_OFFSET, NTLM_TYPE_OFFSET+4),
		LmLen:        _uint16(b, NTLM_TYPE3_LMRESP_OFFSET, NTLM_TYPE3_LMRESP_OFFSET+2),
		LmMax:        _uint16(b, NTLM_TYPE3_LMRESP_OFFSET+2, NTLM_TYPE3_LMRESP_OFFSET+4),
		LmOffset:     _uint16(b, NTLM_TYPE3_LMRESP_OFFSET+4, NTLM_TYPE3_LMRESP_OFFSET+6),
		NtLen:        _uint16(b, NTLM_TYPE3_NTRESP_OFFSET, NTLM_TYPE3_NTRESP_OFFSET+2),
		NtMax:        _uint16(b, NTLM_TYPE3_NTRESP_OFFSET+2, NTLM_TYPE3_NTRESP_OFFSET+4),
		NtOffset:     _uint16(b, NTLM_TYPE3_NTRESP_OFFSET+4, NTLM_TYPE3_NTRESP_OFFSET+6),
		DomainLen:    _uint16(b, NTLM_TYPE3_DOMAIN_OFFSET, NTLM_TYPE3_DOMAIN_OFFSET+2),
		DomainMax:    _uint16(b, NTLM_TYPE3_DOMAIN_OFFSET+2, NTLM_TYPE3_DOMAIN_OFFSET+4),
		DomainOffset: _uint16(b, NTLM_TYPE3_DOMAIN_OFFSET+4, NTLM_TYPE3_DOMAIN_OFFSET+6),
		UserLen:      _uint16(b, NTLM_TYPE3_USER_OFFSET, NTLM_TYPE3_USER_OFFSET+2),
		UserMax:      _uint16(b, NTLM_TYPE3_USER_OFFSET+2, NTLM_TYPE3_USER_OFFSET+4),
		UserOffset:   _uint16(b, NTLM_TYPE3_USER_OFFSET+4, NTLM_TYPE3_USER_OFFSET+6),
		HostLen:      _uint16(b, NTLM_TYPE3_WORKSTN_OFFSET, NTLM_TYPE3_WORKSTN_OFFSET+2),
		HostMax:      _uint16(b, NTLM_TYPE3_WORKSTN_OFFSET+2, NTLM_TYPE3_WORKSTN_OFFSET+4),
		HostOffset:   _uint16(b, NTLM_TYPE3_WORKSTN_OFFSET+4, NTLM_TYPE3_WORKSTN_OFFSET+6),
	}
}

func (data NTLMChallengeResponseParsed) LcString() string {
	// NTLM v1 in .lc format
	if data.Type == NtlmV1 {
		return data.User + "::" + data.Domain + ":" + data.LmHash + ":" + data.ServerChallenge + "\n"
	}
	return data.User + "::" + data.Domain + ":" + data.ServerChallenge + ":" + data.NtHashOne + ":" + data.NtHashTwo + "\n"
}
