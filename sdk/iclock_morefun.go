package sdk

import (
	"crypto/md5"
	"fmt"
	"io"
	"kinger/common/config"
	"kinger/gopuppy/common/glog"
	"net/http"
	"sort"
	"strings"
)

var (
	iClockMoreFunSuccess = []byte("ok")
	iClockMoreFunErr     = []byte("fail")
)

type iClockMoreFun struct {
	appID     string
	appKey    string
	appSecret string
}

func newIClockMoreFun(cfg *config.LoginChannelConfig) ISdk {
	s := &iClockMoreFun{
		appID:     cfg.AppID,
		appKey:    cfg.LoginKey,
		appSecret: cfg.LoginSecret,
	}
	return s
}

func (s *iClockMoreFun) LoginAuth(channelUid, token string) error {
	return nil
}

func (s *iClockMoreFun) getSign(params map[string]string) string {
	var keys []string
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	builder := &strings.Builder{}
	for _, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		builder.WriteString(v)
	}
	builder.WriteString(s.appSecret)

	h := md5.New()
	io.WriteString(h, builder.String())
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *iClockMoreFun) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {

	err := request.ParseForm()
	if err != nil {
		glog.Errorf("iClockMoreFun RechargeAuthSign ParseForm %s, err=%s", request.Body, err)
		reply = iClockMoreFunErr
		return
	}

	channelOrderID = request.Form.Get("trade_no")
	productID := request.Form.Get("product_id")
	url := request.Form.Get("url")
	sign := request.Form.Get("sign")
	appID := request.Form.Get("app_id")
	notifyType := request.Form.Get("notify_type")
	totalFee := request.Form.Get("total_fee")
	//notifyTime := request.Form.Get("notify_time")
	tradeStatus := request.Form.Get("trade_status")
	notifyID := request.Form.Get("notify_id")
	productName := request.Form.Get("product_name")
	//requestCount := request.Form.Get("request_count")
	channelUid = request.Form.Get("uuid")
	extra := request.Form.Get("extra")
	webPay := request.Form.Get("web_pay")

	mySign := s.getSign(map[string]string{
		"trade_no":     channelOrderID,
		"product_id":   productID,
		"url":          url,
		"app_id":       appID,
		"notify_type":  notifyType,
		"total_fee":    totalFee,
		"trade_status": tradeStatus,
		"notify_id":    notifyID,
		"product_name": productName,
		"uuid":         channelUid,
		"extra":        extra,
		"web_pay":      webPay,
	})

	if mySign != sign {
		reply = iClockMoreFunErr
		return
	}

	uid, cpOrderID = parseExt(extra)
	glog.Infof("iClockMoreFun RechargeAuthSign ok, uid=%d, cpOrderID=%s, channelOrderID=%s, money=%s", uid,
		cpOrderID, channelOrderID, totalFee)

	reply = iClockMoreFunSuccess
	ok = true
	return
}
