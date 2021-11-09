package shop

import (
	"fmt"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"net/url"
	"strings"
	"time"
	//"crypto/hmac"
	//"crypto/sha1"
	//"encoding/base64"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"net/http"
	"sort"
	"strconv"
)

type midasPayReply struct {
	Ret     int     `json:"ret"`
	Msg     string  `json:"msg"`
	Billno  string  `json:"billno"`
	Balance float64 `json:"balance"`
}

func (mpr *midasPayReply) String() string {
	return fmt.Sprintf("[midasPayReply ret=%d, msg=%s, billno=%s, balance=%f]", mpr.Ret, mpr.Msg,
		mpr.Billno, mpr.Balance)
}

// 米大师充值，麻烦死
type midasPaymentSt struct {
	player            types.IPlayer
	cptAttr           *attribute.MapAttr
	pendingOrdersAttr *attribute.MapAttr
	doingOrdersAttr   map[string]*attribute.MapAttr
	channelCfg        *config.LoginChannelConfig
}

func newMidasPayment(player types.IPlayer, cptAttr *attribute.MapAttr) *midasPaymentSt {
	channel := player.GetChannel()
	if channel != "lzd_xianfeng_tx" {
		return nil
	}

	channelCfg := config.GetConfig().GetLoginChannelConfig(channel, player.GetLoginChannel())
	if channelCfg == nil {
		return nil
	}

	return &midasPaymentSt{
		player:            player,
		cptAttr:           cptAttr,
		pendingOrdersAttr: cptAttr.GetMapAttr("midasOrders"),
		channelCfg:        channelCfg,
		doingOrdersAttr:   map[string]*attribute.MapAttr{},
	}
}

func (mp *midasPaymentSt) onLogin(player types.IPlayer) {
	mp.player = player
	mp.doingOrdersAttr = map[string]*attribute.MapAttr{}
	if mp.pendingOrdersAttr == nil {
		return
	}

	pendingOrderIDs := mp.pendingOrdersAttr.Keys()
	shopCpt := mp.player.GetComponent(consts.ShopCpt).(*shopComponent)
	for _, orderID := range pendingOrderIDs {
		order := shopCpt.loadOrder(orderID)
		if order == nil || order.isComplete() {
			glog.Errorf("midasPaymentSt onLogin no order, uid=%d, orderID=%s", mp.player.GetUid(), orderID)
			continue
		}

		midasOrder := mp.pendingOrdersAttr.GetMapAttr(orderID)
		if midasOrder == nil {
			continue
		}
		mp.doingOrdersAttr[orderID] = midasOrder

		evq.CallLater(func() {
			mp.doRecharge(player, order, 0)
		})
	}
}

func (mp *midasPaymentSt) onLogout() {
	mp.player = nil
	mp.doingOrdersAttr = map[string]*attribute.MapAttr{}
}

func (mp *midasPaymentSt) getMidasOrderAttr(orderID string) *attribute.MapAttr {
	if mp.pendingOrdersAttr == nil {
		return nil
	}

	if _, ok := mp.doingOrdersAttr[orderID]; ok {
		return nil
	}

	return mp.pendingOrdersAttr.GetMapAttr(orderID)
}

func (mp *midasPaymentSt) makeSig(method, urlPath string, params map[string]string) string {
	mk := mp.makeSource(method, "/v3/r"+urlPath, params)
	h := hmac.New(sha1.New, []byte(mp.channelCfg.MidasAppKey+"&"))
	h.Write(mk)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (mp *midasPaymentSt) makeSource(method, urlPath string, params map[string]string) []byte {
	strs := strings.ToUpper(method) + "&" + url.QueryEscape(urlPath) + "&"

	var keys []string
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	builder := &strings.Builder{}
	for i, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}

		if i != 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(v)
	}

	source := strs + url.QueryEscape(builder.String())
	return []byte(source)
}

func (mp *midasPaymentSt) doRecharge(player types.IPlayer, order *orderSt, tryCnt int) {
	if mp.player != player {
		return
	}

	orderID := order.getCpOrderID()
	midasOrder, ok := mp.doingOrdersAttr[orderID]
	if !ok || midasOrder.GetBool("ok") {
		delete(mp.doingOrdersAttr, orderID)
		return
	}

	params := map[string]string{
		"openid":  player.GetChannelUid(),
		"openkey": midasOrder.GetStr("openkey"),
		"appid":   mp.channelCfg.MidasOfferID,
		"ts":      strconv.FormatInt(time.Now().Unix(), 10),
		"pf":      midasOrder.GetStr("pf"),
		"pfkey":   midasOrder.GetStr("pfkey"),
		"zoneid":  "1",
		"amt":     strconv.Itoa(order.getPrice() * 10),
	}
	sign := mp.makeSig("GET", "/mpay/pay_m", params)

	var payReply *midasPayReply
	url := "https://ysdktest.qq.com/mpay/pay_m"
	//url := "http://msdk.qq.com/mpay/pay_m"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glog.Errorf("midasPaymentSt doRecharge NewRequest error, uid=%d, order=%s, err=%s", player.GetUid(),
			order, err)
	} else {

		q := req.URL.Query()
		q.Add("sig", sign)
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()

		expires := time.Now().Add(120 * time.Second)
		req.AddCookie(&http.Cookie{
			Name:    "org_loc",
			Value:   "/mpay/pay_m",
			Expires: expires,
		})

		sessionID := "hy_gameid"
		sessionType := "wc_actoken"
		if player.GetLoginChannel() == "qq" {
			sessionID = "openid"
			sessionType = "kp_actoken"
		}
		req.AddCookie(&http.Cookie{
			Name:    "session_id",
			Value:   sessionID,
			Expires: expires,
		})
		req.AddCookie(&http.Cookie{
			Name:    "session_type",
			Value:   sessionType,
			Expires: expires,
		})

		evq.Await(func() {
			var resp *http.Response
			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			if resp.StatusCode != 200 {
				err = errors.New("doRecharge http error")
				return
			}

			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}

			payReply = &midasPayReply{}
			err = json.Unmarshal(body, payReply)
		})

	}

	if err == nil && payReply != nil {
		if payReply.Ret != 0 {
			glog.Errorf("midasPaymentSt doRecharge error, uid=%d, order=%s, reply=%s", player.GetUid(),
				order, payReply)
		} else {

			glog.Infof("midasPaymentSt doRecharge ok, uid=%d, order=%s, reply=%s", player.GetUid(),
				order, payReply)

			midasOrder.SetBool("ok", true)
			delete(mp.doingOrdersAttr, orderID)
			if mp.pendingOrdersAttr != nil {
				mp.pendingOrdersAttr.Del(orderID)
				if mp.pendingOrdersAttr.Size() <= 0 {
					mp.cptAttr.Del("midasOrders")
					mp.pendingOrdersAttr = nil
				}
			}

			player.GetComponent(consts.ShopCpt).(*shopComponent).OnSdkRecharge(player.GetChannelUid(), orderID,
				payReply.Billno, order.getPrice(), true)
			return
		}
	}

	if tryCnt >= 20 {
		delete(mp.doingOrdersAttr, orderID)
		agnet := player.GetAgent()
		if agnet != nil {
			agnet.PushClient(pb.MessageID_S2C_NOTIFY_SDK_RECHARGE_RESULT, &pb.SdkRechargeResult{
				Errcode: pb.SdkRechargeResult_Fail,
			})
		}
	} else {
		timer.AfterFunc(time.Duration(tryCnt/2+1)*time.Second, func() {
			mp.doRecharge(player, order, tryCnt+1)
		})
	}
}

func (mp *midasPaymentSt) onRecharge(orderID, openkey, pf, pfkey string) {
	shopCpt := mp.player.GetComponent(consts.ShopCpt).(*shopComponent)
	order := shopCpt.loadOrder(orderID)
	if order == nil || order.isComplete() {
		glog.Errorf("midasPaymentSt onRecharge no order, uid=%d, orderID=%s", mp.player.GetUid(), orderID)
		return
	}

	if mp.pendingOrdersAttr != nil {
		midasOrder := mp.pendingOrdersAttr.GetMapAttr(orderID)
		if midasOrder != nil {
			return
		}
	}

	midasOrder := attribute.NewMapAttr()
	midasOrder.SetStr("openkey", openkey)
	midasOrder.SetStr("pf", pf)
	midasOrder.SetStr("pfkey", pfkey)
	if mp.pendingOrdersAttr == nil {
		mp.pendingOrdersAttr = attribute.NewMapAttr()
		mp.cptAttr.SetMapAttr("midasOrders", mp.pendingOrdersAttr)
	}
	mp.pendingOrdersAttr.SetMapAttr(orderID, midasOrder)

	mp.doingOrdersAttr[orderID] = midasOrder
	mp.doRecharge(mp.player, order, 0)
}
