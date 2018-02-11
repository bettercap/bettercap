package packets

import (
	"errors"
	"strconv"
	"time"

	"encoding/asn1"
	"encoding/hex"
)

const (
	Krb5AsRequestType         = 10
	Krb5Krb5PrincipalNameType = 1
	Krb5CryptDesCbcMd4        = 2
	Krb5CryptDescCbcMd5       = 3
	Krb5CryptRc4Hmac          = 23
)

//https://github.com/heimdal/heimdal/blob/master/lib/asn1/krb5.asn1
var (
	Krb5AsReqParam = "application,explicit,tag:10"
)

type Krb5PrincipalName struct {
	NameType   int      `asn1:"explicit,tag:0"`
	NameString []string `asn1:"general,explicit,tag:1"`
}

type Krb5EncryptedData struct {
	Etype  int    `asn1:"explicit,tag:0"`
	Kvno   int    `asn1:"optional,explicit,tag:1"`
	Cipher []byte `asn1:"explicit,tag:2"`
}

type Krb5Ticket struct {
	TktVno  int               `asn1:"explicit,tag:0"`
	Realm   string            `asn1:"general,explicit,tag:1"`
	Sname   Krb5PrincipalName `asn1:"explicit,tag:2"`
	EncPart Krb5EncryptedData `asn1:"explicit,tag:3"`
}

type Krb5Address struct {
	AddrType    int    `asn1:"explicit,tag:0"`
	Krb5Address []byte `asn1:"explicit,tag:1"`
}

type Krb5PnData struct {
	Krb5PnDataType  int    `asn1:"explicit,tag:1"`
	Krb5PnDataValue []byte `asn1:"explicit,tag:2"`
}

type Krb5ReqBody struct {
	KDCOptions            asn1.BitString    `asn1:"explicit,tag:0"`
	Cname                 Krb5PrincipalName `asn1:"optional,explicit,tag:1"`
	Realm                 string            `asn1:"general,explicit,tag:2"`
	Sname                 Krb5PrincipalName `asn1:"optional,explicit,tag:3"`
	From                  time.Time         `asn1:"generalized,optional,explicit,tag:4"`
	Till                  time.Time         `asn1:"generalized,optional,explicit,tag:5"`
	Rtime                 time.Time         `asn1:"generalized,optional,explicit,tag:6"`
	Nonce                 int               `asn1:"explicit,tag:7"`
	Etype                 []int             `asn1:"explicit,tag:8"`
	Krb5Addresses         []Krb5Address     `asn1:"optional,explicit,tag:9"`
	EncAuthData           Krb5EncryptedData `asn1:"optional,explicit,tag:10"`
	AdditionalKrb5Tickets []Krb5Ticket      `asn1:"optional,explicit,tag:11"`
}

type Krb5Request struct {
	Pvno       int          `asn1:"explicit,tag:1"`
	MsgType    int          `asn1:"explicit,tag:2"`
	Krb5PnData []Krb5PnData `asn1:"optional,explicit,tag:3"`
	ReqBody    Krb5ReqBody  `asn1:"explicit,tag:4"`
}

func (kdc Krb5Request) String() (string, error) {
	var eType, cipher string
	var crypt []string
	realm := kdc.ReqBody.Realm

	if kdc.ReqBody.Cname.NameType == Krb5Krb5PrincipalNameType {
		crypt = kdc.ReqBody.Cname.NameString
	}
	if len(crypt) != 1 {
		return "", errors.New("No crypt alg found")
	}
	for _, pn := range kdc.Krb5PnData {
		if pn.Krb5PnDataType == 2 {
			enc, err := pn.getParsedValue()
			if err != nil {
				return "", errors.New("Failed to extract pnData from as-req")
			}
			eType = strconv.Itoa(enc.Etype)
			cipher = hex.EncodeToString(enc.Cipher)
		}
	}
	if eType == "" || cipher == "" {
		return "", errors.New("No encryption type or cipher found")
	}
	hash := "$krb5$" + eType + "$" + crypt[0] + "$" + realm + "$nodata$" + cipher
	return hash, nil
}

func (pd Krb5PnData) getParsedValue() (Krb5EncryptedData, error) {
	var encData Krb5EncryptedData
	_, err := asn1.Unmarshal(pd.Krb5PnDataValue, &encData)
	if err != nil {
		return Krb5EncryptedData{}, errors.New("Failed to parse pdata value")
	}
	return encData, nil
}
