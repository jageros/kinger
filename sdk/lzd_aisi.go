package sdk

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"kinger/common/config"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	grsa "kinger/gopuppy/common/rsa"
	"net/http"
	"strings"
)

var (
	aiSiPaySuccess = []byte("success")
	aiSiPayErr     = []byte("fail")
)

type aiSiLoginReply struct {
	Status   int    `json:"status"`
	UserName string `json:"username"`
	UserId   int    `json:"userid"`
}

type aiSiSdk struct {
	publickey string
	pub       *rsa.PublicKey
}

func newAiSiSdk(cfg *config.LoginChannelConfig) ISdk {
	s := &aiSiSdk{publickey: cfg.PaySecret}
	return s
	decodePublic, err := base64.StdEncoding.DecodeString(cfg.PaySecret)
	if err != nil {
		glog.Errorf("newAiSiSdk DecodeString err %s", err)
		return nil
	}
	pubInterface, err := x509.ParsePKIXPublicKey(decodePublic)
	if err != nil {
		glog.Errorf("newAiSiSdk ParsePKIXPublicKey err %s", err)
		return nil
	}
	s.pub = pubInterface.(*rsa.PublicKey)
	return s
}

func (s *aiSiSdk) LoginAuth(channelUid, token string) error {
	var body []byte
	var err error
	evq.Await(func() {
		var resp *http.Response
		resp, err = http.Get("https://pay.i4.cn/member_third.action?token=" + token)
		if err != nil {
			return
		}
		if resp.StatusCode != 200 {
			err = errors.Errorf("aiSiSdk LoginAuth StatusCode %d", resp.StatusCode)
			return
		}

		body, err = ioutil.ReadAll(resp.Body)
	})

	if err != nil {
		return err
	}

	reply := &aiSiLoginReply{}
	if err := json.Unmarshal(body, reply); err != nil {
		glog.Errorf("aiSiSdk LoginAuth json.Unmarshal error body=%s, err=%s", body, err)
		return err
	}

	if reply.Status != 0 {
		return errors.Errorf("aiSiSdk LoginAuth %d %s", reply.Status, string(body))
	}

	return nil
}

func (s *aiSiSdk) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {

	err := request.ParseForm()
	if err != nil {
		glog.Errorf("aiSiSdk RechargeAuthSign ParseForm %s, err=%s", request.Body, err)
		reply = xianFengErr
		return
	}

	params := map[string]string{}
	for k, vs := range request.Form {
		if k == "sign" {
			continue
		}

		if len(vs) > 0 {
			v := vs[0]
			if v == "" {
				continue
			}
			params[k] = v
		}
	}

	channelOrderID = request.Form.Get("order_id")
	ext := request.Form.Get("billno")
	status := request.Form.Get("status")
	sign := request.Form.Get("sign")

	if !s.verifySignature(params, sign) {
		glog.Errorf("aisi sign error %v", request.Form)
		reply = aiSiPayErr
		return
	}

	if status != "0" {
		reply = aiSiPaySuccess
		return
	}

	uid, cpOrderID = parseExt(ext)
	reply = aiSiPaySuccess
	ok = true
	return
}

func (s *aiSiSdk) parseSignature(sign string) map[string]string {
	//dcDataStr, err := base64.StdEncoding.Decode(sign)
	//glog.Infof("parseSignature 1111 %s", string(dcDataStr))
	//if err != nil {
	//	return nil
	//}

	plainData, err := grsa.RSAPublicKeyDecryptBase64(s.publickey, []byte(sign))
	if err != nil {
		glog.Errorf("aiSiSdk publicDecrypt error %s", err)
		return nil
	}

	parseString := string(plainData)
	out := map[string]string{}
	for _, kv := range strings.Split(parseString, "&") {
		pair := strings.Split(kv, "=")
		if len(pair) != 2 {
			return nil
		}
		out[pair[0]] = pair[1]
	}
	return out
}

func (s *aiSiSdk) verifySignature(params map[string]string, sign string) bool {
	signature := s.parseSignature(sign)
	if signature == nil {
		return false
	}

	for k, v := range params {
		if signature[k] != v {
			return false
		}
	}
	return true
}
