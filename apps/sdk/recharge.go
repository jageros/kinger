package main

import (
	"net/http"
	"strings"
	"kinger/sdk"
	"kinger/gopuppy/common/evq"
	"kinger/common/utils"
	"kinger/proto/pb"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common"
	gutils "kinger/gopuppy/common/utils"
)

func wrapHttpHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		gutils.CatchPanic(func() {
			handler(writer, request)
		})
	}
}

func rechargeHandler(writer http.ResponseWriter, request *http.Request) {
	pathInfo := strings.Split(request.URL.Path, "/")
	if len(pathInfo) < 3 {
		return
	}

	channel := pathInfo[2]
	var loginChannel string
	if len(pathInfo) > 3 {
		loginChannel = pathInfo[3]
	}

	s := sdk.GetSdk(channel, loginChannel)
	if s == nil {
		glog.Errorf("rechargeHandler %s %s no sdk", channel, loginChannel)
		return
	}

	uid, channelUid, cpOrderID, channelOrderID, paymentAmount, reply, needCheckMoney, ok := s.RechargeAuthSign(request)
	writer.Write(reply)

	if ok {
		glog.Infof("RechargeAuthSign sign ok, channel=%s, uid=%d, channelUid=%s, cpOrderID=%s, channelOrderID=%s, " +
			"paymentAmount=%d", channel, uid, channelUid, cpOrderID, channelOrderID, paymentAmount)
		evq.CallLater(func() {
			utils.PlayerMqPublish(common.UUid(uid), pb.RmqType_SdkRecharge, &pb.RmqSdkRecharge{
				ChannelUid: channelUid,
				CpOrderID: cpOrderID,
				ChannelOrderID: channelOrderID,
				PaymentAmount: int32(paymentAmount),
				NeedCheckMoney: needCheckMoney,
			})
		})
	} else {
		glog.Errorf("RechargeAuthSign sign error, channel=%s, uid=%d, channelUid=%s, cpOrderID=%s, channelOrderID=%s, " +
			"paymentAmount=%d, reply=%s", channel, uid, channelUid, cpOrderID, channelOrderID, paymentAmount, reply)
	}
}

func initializeRouter() {
	http.HandleFunc("/recharge/", wrapHttpHandler(rechargeHandler))
}
