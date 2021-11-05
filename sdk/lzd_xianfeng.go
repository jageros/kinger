package sdk

import (
	"net/http"
	"io/ioutil"
	"kinger/gopuppy/common/glog"
	"github.com/pkg/errors"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"crypto/md5"
	"io"
	"strconv"
	"kinger/common/config"
	"kinger/gopuppy/common/evq"
)

var (
	xianFengSuccess = []byte("SUCCESS")
	xianFengErr = []byte("FAILURE")
	xianFengAuthNullErr = errors.New("null channelID")
)

type iXianfengSdkAuthor interface {
	getAuthUrl() string
}

type xianFengLoginReply struct {
	Errorcode int `json:"errorcode"`
	Msg string `json:"msg"`
}

type xianFeng struct {
	author iXianfengSdkAuthor
	appID string
	appKey string
	appSecret string
}

func newXianFeng(cfg *config.LoginChannelConfig) ISdk {
	s := &xianFeng{
		appID: cfg.AppID,
		appKey: cfg.LoginKey,
		appSecret: cfg.LoginSecret,
	}
	s.author = s
	return s
}

func (s *xianFeng) getAuthUrl() string {
	return "https://proxy3sdk.5199.com/fire-3sdk/api/auth/login"
}

func (s *xianFeng) getSign(params map[string]string) string {
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
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(v)
		builder.WriteString("&")
	}
	builder.WriteString(s.appSecret)

	h := md5.New()
	io.WriteString(h, builder.String())
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *xianFeng) LoginAuth(channelUid, token string) error {
	if channelUid == "" || channelUid == "null" {
		return xianFengAuthNullErr
	}
	var body []byte
	var err error
	evq.Await(func() {
		var resp *http.Response
		resp, err = http.PostForm(s.author.getAuthUrl(), url.Values{
			"appId": {s.appID},
			"uid": {channelUid},
			"token": {token},
			"sign": {
				s.getSign(map[string]string{
					"appId": s.appID,
					"uid": channelUid,
					"token": token,
				}),
			},
		})

		if err != nil {
			return
		}
		if resp.StatusCode != 200 {
			err = errors.Errorf("xianFeng LoginAuth StatusCode %d", resp.StatusCode)
			return
		}

		body, err = ioutil.ReadAll(resp.Body)
	})

	if err != nil {
		return err
	}

	reply := &xianFengLoginReply{}
	if err := json.Unmarshal(body, reply); err != nil {
		glog.Errorf("xianFeng LoginAuth json.Unmarshal error body=%s, err=%s", body, err)
		return err
	}

	if reply.Errorcode != 0 {
		return errors.Errorf("xianFeng LoginAuth %d %s", reply.Errorcode, reply.Msg)
	}

	return nil
}

func (s *xianFeng) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {

	err := request.ParseForm()
	if err != nil {
		glog.Errorf("xianFeng RechargeAuthSign ParseForm %s, err=%s", request.Body, err)
		reply = xianFengErr
		return
	}

	appId := request.Form.Get("appId")
	packageId := request.Form.Get("packageId")
	channelUid = request.Form.Get("uid")
	sdkId := request.Form.Get("sdkId")
	channelOrderID = request.Form.Get("orderId")
	money := request.Form.Get("money")
	way := request.Form.Get("way")
	ext := request.Form.Get("ext")
	status := request.Form.Get("status")
	createTime := request.Form.Get("createTime")
	sign := request.Form.Get("sign")

	mySign := s.getSign(map[string]string{
		"appId": appId,
		"packageId": packageId,
		"uid": channelUid,
		"sdkId": sdkId,
		"orderId": channelOrderID,
		"money": money,
		"way": way,
		"ext": ext,
		"status": status,
		"createTime": createTime,
	})

	if mySign != sign {
		reply = xianFengErr
		return
	}

	reply = xianFengSuccess
	if status != "0" {
		return
	}

	uid, cpOrderID = parseExt(ext)
	if strings.HasPrefix(cpOrderID, "ios") {
		http.PostForm("http://10.10.26.218:6667/test_recharge", url.Values{
			"uid": []string{strconv.FormatUint(uid, 10)},
			"cpOrderID": []string{cpOrderID},
			"channelUid": []string{channelUid},
			"channelOrderID": []string{channelOrderID},
		})
		return
	}

	paymentAmount, _ = strconv.Atoi(money)
	paymentAmount /= 100
	ok = true
	needCheckMoney = true
	return
}

type multilanXianFeng struct {
	xianFeng
}

func newMultilanXianFeng(cfg *config.LoginChannelConfig) ISdk {
	s := &multilanXianFeng{}
	s.appID = cfg.AppID
	s.appKey = cfg.LoginKey
	s.appSecret = cfg.LoginSecret
	s.author = s
	return s
}

func (s *multilanXianFeng) LoginAuth(channelUid, token string) error {
	return nil
}

func (s *multilanXianFeng) getAuthUrl() string {
	return "https://proxy3sdk.teraent.com/fire-3sdk/api/auth/login"
}
