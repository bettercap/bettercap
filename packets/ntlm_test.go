package packets

import (
	"reflect"
	"testing"
)

func TestNTLMConstants(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{NTLM_SIG_OFFSET, 0},
		{NTLM_TYPE_OFFSET, 8},
		{NTLM_TYPE1_FLAGS_OFFSET, 12},
		{NTLM_TYPE1_DOMAIN_OFFSET, 16},
		{NTLM_TYPE1_WORKSTN_OFFSET, 24},
		{NTLM_TYPE1_DATA_OFFSET, 32},
		{NTLM_TYPE1_MINSIZE, 16},
		{NTLM_TYPE2_TARGET_OFFSET, 12},
		{NTLM_TYPE2_FLAGS_OFFSET, 20},
		{NTLM_TYPE2_CHALLENGE_OFFSET, 24},
		{NTLM_TYPE2_CONTEXT_OFFSET, 32},
		{NTLM_TYPE2_TARGETINFO_OFFSET, 40},
		{NTLM_TYPE2_DATA_OFFSET, 48},
		{NTLM_TYPE2_MINSIZE, 32},
		{NTLM_TYPE3_LMRESP_OFFSET, 12},
		{NTLM_TYPE3_NTRESP_OFFSET, 20},
		{NTLM_TYPE3_DOMAIN_OFFSET, 28},
		{NTLM_TYPE3_USER_OFFSET, 36},
		{NTLM_TYPE3_WORKSTN_OFFSET, 44},
		{NTLM_TYPE3_SESSIONKEY_OFFSET, 52},
		{NTLM_TYPE3_FLAGS_OFFSET, 60},
		{NTLM_TYPE3_DATA_OFFSET, 64},
		{NTLM_TYPE3_MINSIZE, 52},
		{NTLM_BUFFER_LEN_OFFSET, 0},
		{NTLM_BUFFER_MAXLEN_OFFSET, 2},
		{NTLM_BUFFER_OFFSET_OFFSET, 4},
		{NTLM_BUFFER_SIZE, 8},
		{NtlmV1, 1},
		{NtlmV2, 2},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestNTLMChallengeResponse(t *testing.T) {
	r := NTLMChallengeResponse{}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{r.Challenge, ""},
		{r.Response, ""},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestNTLMChallengeResponseParsed(t *testing.T) {
	r := NTLMChallengeResponseParsed{}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{r.Type, 0},
		{r.ServerChallenge, ""},
		{r.User, ""},
		{r.Domain, ""},
		{r.LmHash, ""},
		{r.NtHashOne, ""},
		{r.NtHashTwo, ""},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestNTLMResponseHeader(t *testing.T) {
	r := NTLMResponseHeader{}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{r.Sig, ""},
		{r.Type, uint32(0)},
		{r.LmLen, uint16(0)},
		{r.LmMax, uint16(0)},
		{r.LmOffset, uint16(0)},
		{r.NtLen, uint16(0)},
		{r.NtMax, uint16(0)},
		{r.NtOffset, uint16(0)},
		{r.DomainLen, uint16(0)},
		{r.DomainMax, uint16(0)},
		{r.DomainOffset, uint16(0)},
		{r.UserLen, uint16(0)},
		{r.UserMax, uint16(0)},
		{r.UserOffset, uint16(0)},
		{r.HostLen, uint16(0)},
		{r.HostMax, uint16(0)},
		{r.HostOffset, uint16(0)},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestNTLMState(t *testing.T) {
	r := map[uint32]string{}
	p := []NTLMChallengeResponse{}
	s := NTLMState{
		Responses: r,
		Pairs:     p,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{s.Responses, r},
		{s.Pairs, p},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func BuildExampleNTLMState() NTLMState {
	return NTLMState{
		Responses: map[uint32]string{},
		Pairs:     []NTLMChallengeResponse{},
	}
}

func TestNTLMStateAddServerResponse(t *testing.T) {
	s := BuildExampleNTLMState()

	s.AddServerResponse(uint32(0), "picat")

	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{s.Responses[uint32(0)], "picat"},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

// TODO: add tests for the rest of NTLM :P
