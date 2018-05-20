package packets

import (
	"encoding/asn1"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestKrb5Contants(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{Krb5AsRequestType, 10},
		{Krb5Krb5PrincipalNameType, 1},
		{Krb5CryptDesCbcMd4, 2},
		{Krb5CryptDescCbcMd5, 3},
		{Krb5CryptRc4Hmac, 23},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestKrb5Vars(t *testing.T) {
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{ErrNoCrypt, errors.New("No crypt alg found")},
		{ErrReqData, errors.New("Failed to extract pnData from as-req")},
		{ErrNoCipher, errors.New("No encryption type or cipher found")},
		{Krb5AsReqParam, "application,explicit,tag:10"},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestKrb5PrincipalName(t *testing.T) {
	str := []string{"example"}
	name := Krb5PrincipalName{NameString: str}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{name.NameType, 0},
		{name.NameString, str},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestKrb5EncryptedData(t *testing.T) {
	cipher := []byte{}
	data := Krb5EncryptedData{Cipher: cipher}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{data.Cipher, cipher},
		{data.Etype, 0},
		{data.Kvno, 0},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestKrb5Ticket(t *testing.T) {
	v := 0
	r := "picat"
	s := Krb5PrincipalName{}
	e := Krb5EncryptedData{}

	ticket := Krb5Ticket{
		TktVno:  v,
		Realm:   r,
		Sname:   s,
		EncPart: e,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{ticket.TktVno, v},
		{ticket.Realm, r},
		{ticket.Sname, s},
		{ticket.EncPart, e},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestKrb5Address(t *testing.T) {
	x := 0
	y := []byte{}
	addr := Krb5Address{
		AddrType:    x,
		Krb5Address: y,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{addr.AddrType, x},
		{addr.Krb5Address, y},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestKrb5PnData(t *testing.T) {
	x := 0
	y := []byte{}
	addr := Krb5PnData{
		Krb5PnDataType:  x,
		Krb5PnDataValue: y,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{addr.Krb5PnDataType, x},
		{addr.Krb5PnDataValue, y},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}
}

func TestKrb5ReqBody(t *testing.T) {
	e := []int{}
	a := []Krb5Address{}
	k := []Krb5Ticket{}
	req := Krb5ReqBody{
		Etype:                 e,
		Krb5Addresses:         a,
		AdditionalKrb5Tickets: k,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{req.KDCOptions, asn1.BitString{}},
		{req.Cname, Krb5PrincipalName{}},
		{req.Realm, ""},
		{req.Sname, Krb5PrincipalName{}},
		{req.From, time.Time{}},
		{req.Till, time.Time{}},
		{req.Rtime, time.Time{}},
		{req.Nonce, 0},
		{req.Etype, e},
		{req.Krb5Addresses, a},
		{req.EncAuthData, Krb5EncryptedData{}},
		{req.AdditionalKrb5Tickets, k},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}

}

func TestKrb5Request(t *testing.T) {
	p := []Krb5PnData{}
	req := Krb5Request{
		Krb5PnData: p,
	}
	var units = []struct {
		got interface{}
		exp interface{}
	}{
		{req.Pvno, 0},
		{req.MsgType, 0},
		{req.Krb5PnData, p},
		{req.ReqBody, Krb5ReqBody{}},
	}
	for _, u := range units {
		if !reflect.DeepEqual(u.exp, u.got) {
			t.Fatalf("expected '%v', got '%v'", u.exp, u.got)
		}
	}

}

// TODO: add test for func (kdc Krb5Request) String()
// TODO: add test for func (pd Krb5PnData) getParsedValue()
