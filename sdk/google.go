package sdk

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"github.com/pkg/errors"
	"io/ioutil"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"net/http"
	"strings"
)

const _GOOGLE_OAUTH2_CERTS_URL = "https://www.googleapis.com/oauth2/v1/certs"

var pubKeys = map[string]*rsa.PublicKey{}

type claimSet struct {
	Iss   string `json:"iss"`             // email address of the client_id of the application making the access token request
	Scope string `json:"scope,omitempty"` // space-delimited list of the permissions the application requests
	Aud   string `json:"aud"`             // descriptor of the intended target of the assertion (Optional).
	Exp   int64  `json:"exp"`             // the expiration time of the assertion (seconds since Unix epoch)
	Iat   int64  `json:"iat"`             // the time the assertion was issued (seconds since Unix epoch)
	Typ   string `json:"typ,omitempty"`   // token type (Optional).

	// Email for which the application is requesting delegated access (Optional).
	Sub string `json:"sub,omitempty"`

	// The old name of Sub. Client keeps setting Prn to be
	// complaint with legacy OAuth 2.0 providers. (Optional)
	Prn string `json:"prn,omitempty"`
}

type googleSdk struct {
	clientID string
}

func newGoogleSdk() *googleSdk {
	return &googleSdk{
		clientID: "847435633362-00de741eog945eqsk9rebc0mltvs15ar.apps.googleusercontent.com",
	}
}

func (s *googleSdk) LoginAuth(channelUid, token string) error {
	c, err := s.verifyToken(token, 0)
	if err != nil {
		return err
	}

	if c.Sub != channelUid {
		return errors.Errorf("wrong channelUid, sub=%s, channelUid=%s", c.Sub, channelUid)
	}
	return nil
}

func (s *googleSdk) verifyToken(idToken string, tryCnt int) (*claimSet, error) {
	if len(pubKeys) <= 0 {
		var resp *http.Response
		var err error
		evq.Await(func() {
			resp, err = http.Get(_GOOGLE_OAUTH2_CERTS_URL)
		})
		if err != nil {
			return nil, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		certs := map[string]string{}
		err = json.Unmarshal(body, &certs)
		if err != nil {
			glog.Infof("googleSdk LoginAuth 1111111 %s", err)
			return nil, err
		}

		for kid, c := range certs {
			block, _ := pem.Decode([]byte(c))
			if block == nil {
				return nil, errors.New("public key error")
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				glog.Infof("googleSdk LoginAuth 222222222 %s", err)
				return nil, err
			}
			//pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
			//pub := pubInterface.(*rsa.PublicKey)
			pubKeys[kid] = cert.PublicKey.(*rsa.PublicKey)
		}
	}

	c, signedContent, signatureString, err := s.unverifiedDecode(idToken)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(signedContent)
	for _, key := range pubKeys {
		err2 := rsa.VerifyPKCS1v15(key, crypto.SHA256, h.Sum(nil), signatureString)
		if err2 != nil {
			err = err2
		} else {
			err = nil
			break
		}
	}
	if err != nil {
		if tryCnt > 0 {
			return nil, err
		} else {
			pubKeys = map[string]*rsa.PublicKey{}
			return s.verifyToken(idToken, 1)
		}
	}

	isIssOk := false
	for _, iss := range []string{"accounts.google.com", "https://accounts.google.com"} {
		if c.Iss == iss {
			isIssOk = true
			break
		}
	}
	if !isIssOk {
		return nil, errors.New("Wrong issuer.")
	}
	return c, nil
}

func (s *googleSdk) unverifiedDecode(idToken string) (c *claimSet, signedContent, signatureString []byte,
	err error) {

	glog.Infof("googleSdk unverifiedDecode %s", idToken)
	parts := strings.Split(idToken, ".")
	if len(parts) < 3 {
		err = errors.New("unverifiedDecode: invalid token received")
		return
	}

	decoded, err2 := base64.RawURLEncoding.DecodeString(parts[1])
	if err2 != nil {
		err = err2
		return
	}

	c = &claimSet{}
	err = json.NewDecoder(bytes.NewBuffer(decoded)).Decode(c)
	if err != nil {
		return
	}

	signedContent = []byte(parts[0] + "." + parts[1])
	signatureString, err = base64.RawURLEncoding.DecodeString(parts[2])
	return
}

func (s *googleSdk) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {
	return
}
