package sdk

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"kinger/common/config"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"net/http"
	"strconv"
	"time"
)

type tencentYSDKLoginReply struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

type iTencentYSDKLoginAuther interface {
	getLoginAuthUrl() string
}

type tencentYSDK struct {
	loginAuther iTencentYSDKLoginAuther
	appID       string
	appKey      string
}

type tencentQQYSDK struct {
	tencentYSDK
}

type tencentWechatYSDK struct {
	tencentYSDK
}

func newTencentQQYSDK(cfg *config.LoginChannelConfig) ISdk {
	s := &tencentQQYSDK{}
	s.loginAuther = s
	s.appID = cfg.AppID
	s.appKey = cfg.LoginKey
	return s
}

func newTencentWechatYSDK(cfg *config.LoginChannelConfig) ISdk {
	s := &tencentWechatYSDK{}
	s.loginAuther = s
	s.appID = cfg.AppID
	s.appKey = cfg.LoginKey
	return s
}

func (s *tencentYSDK) LoginAuth(channelUid, token string) error {
	var body []byte
	var err error

	evq.Await(func() {

		urlPath := s.loginAuther.getLoginAuthUrl()
		//url := "http://ysdk.qq.com" + urlPath
		url := "http://ysdktest.qq.com" + urlPath
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		h := md5.New()
		h.Write([]byte(s.appKey + timestamp))
		uri := fmt.Sprintf("%s?appid=%s&openid=%s&openkey=%s&timestamp=%s&sig=%s", url, s.appID, channelUid,
			token, timestamp, fmt.Sprintf("%x", h.Sum(nil)))

		var resp *http.Response
		resp, err = http.Get(uri)
		if err != nil {
			return
		}
		if resp.StatusCode != 200 {
			err = errors.Errorf("tencentYSDK LoginAuth StatusCode %d", resp.StatusCode)
			return
		}

		body, err = ioutil.ReadAll(resp.Body)
	})

	if err != nil {
		return err
	}

	reply := &tencentYSDKLoginReply{}
	if err := json.Unmarshal(body, reply); err != nil {
		glog.Errorf("tencentYSDK LoginAuth json.Unmarshal error body=%s, err=%s", body, err)
		return err
	}

	if reply.Ret != 0 {
		return errors.Errorf("tencentYSDK LoginAuth %d %s", reply.Ret, reply.Msg)
	}

	return nil
}

func (s *tencentYSDK) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {

	return
}

func (s *tencentQQYSDK) getLoginAuthUrl() string {
	return "/auth/qq_check_token"
}

func (s *tencentWechatYSDK) getLoginAuthUrl() string {
	return "/auth/wx_check_token"
}
