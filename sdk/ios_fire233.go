package sdk

import (
	"crypto/md5"
	"fmt"
	"io"
	"kinger/gopuppy/common/glog"
	"net/http"
	"strconv"
	"strings"
)

var (
	iosFire233Success = []byte("SUCCESS")
	iosFire233Err     = []byte("FAILURE")
)

type iosFire233 struct {
	gameSecret string
}

func newIosFire233() *iosFire233 {
	return &iosFire233{
		gameSecret: "74d94579fc169cb9a02932243724c263",
	}
}

func (s *iosFire233) LoginAuth(channelUid, token string) error {
	return nil
}

func (s *iosFire233) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {

	err := request.ParseForm()
	if err != nil {
		glog.Infof("iosFire233 RechargeAuthSign ParseForm %s, err=%s", request.Body, err)
		reply = iosFire233Err
		return
	}

	serverID := request.Form.Get("serverId")
	cpOrderID = request.Form.Get("callbackInfo")
	channelUid = request.Form.Get("openId")
	channelOrderID = request.Form.Get("orderId")
	orderStatus := request.Form.Get("orderStatus")
	payType := request.Form.Get("payType")
	amount := request.Form.Get("amount")
	money := request.Form.Get("money")
	remark := request.Form.Get("remark")
	signature := request.Form.Get("sign")

	strUid := cpOrderID[:len(cpOrderID)-16]
	uid, _ = strconv.ParseUint(strUid, 10, 64)
	paymentAmount2, _ := strconv.ParseFloat(money, 64)
	paymentAmount = int(paymentAmount2)

	if orderStatus != "1" {
		reply = iosFire233Success
		return
	}

	sourceStrBuilder := strings.Builder{}
	sourceStrBuilder.WriteString(serverID)
	sourceStrBuilder.WriteString(cpOrderID)
	sourceStrBuilder.WriteString(channelUid)
	sourceStrBuilder.WriteString(channelOrderID)
	sourceStrBuilder.WriteString(orderStatus)
	sourceStrBuilder.WriteString(payType)
	sourceStrBuilder.WriteString(amount)
	sourceStrBuilder.WriteString(remark)
	sourceStrBuilder.WriteString(s.gameSecret)
	h := md5.New()
	io.WriteString(h, sourceStrBuilder.String())
	sign := fmt.Sprintf("%x", h.Sum(nil))
	if sign != signature {
		reply = iosFire233Err
		return
	}

	reply = iosFire233Success
	ok = true
	needCheckMoney = true
	return
}
