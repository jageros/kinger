package sdk

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"kinger/gopuppy/common/glog"
	"net/http"
	"strconv"
	"strings"
)

var (
	fire233Success, _ = json.Marshal(map[string]interface{}{
		"code":    0,
		"message": "充值成功",
	})

	fire233GameErr, _ = json.Marshal(map[string]interface{}{
		"code":    4,
		"message": "游戏服错误",
	})

	fire233SignErr, _ = json.Marshal(map[string]interface{}{
		"code":    3,
		"message": "签名错误",
	})
)

type fire233 struct {
	gameSecret string
}

func newFire233() *fire233 {
	return &fire233{
		gameSecret: "069fb6c64bb486fda74f0ffa6b854cb0",
	}
}

func (s *fire233) LoginAuth(channelUid, token string) error {
	return nil
}

func (s *fire233) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {

	err := request.ParseForm()
	if err != nil {
		glog.Infof("fire233 RechargeAuthSign ParseForm %s, err=%s", request.Body, err)
		reply = fire233GameErr
		return
	}

	roleID := request.Form.Get("role_id")
	channelUid = request.Form.Get("open_id")
	strPaymentAmount := request.Form.Get("payment_amount")
	cpOrderID = request.Form.Get("game_payorder")
	channelOrderID = request.Form.Get("pay_orderno")
	timestamp := request.Form.Get("timestamp")
	signature := request.Form.Get("signature")

	uid, _ = strconv.ParseUint(roleID, 10, 64)
	paymentAmount2, _ := strconv.ParseFloat(strPaymentAmount, 64)
	paymentAmount = int(paymentAmount2)

	sourceStrBuilder := strings.Builder{}
	sourceStrBuilder.WriteString(roleID)
	sourceStrBuilder.WriteString(channelUid)
	sourceStrBuilder.WriteString(strPaymentAmount)
	sourceStrBuilder.WriteString(cpOrderID)
	sourceStrBuilder.WriteString(timestamp)
	sourceStrBuilder.WriteString(s.gameSecret)
	h := md5.New()
	io.WriteString(h, sourceStrBuilder.String())
	sign := fmt.Sprintf("%x", h.Sum(nil))
	if sign != signature {
		reply = fire233SignErr
		return
	}

	reply = fire233Success
	ok = true
	needCheckMoney = true
	return
}
