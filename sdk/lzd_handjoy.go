package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"kinger/gopuppy/common/glog"
	"net/http"
)

type facebookLoginReply struct {
	Data struct {
		AppID     string   `json:"app_id"`
		ExpiresAt int64    `json:"expires_at"`
		IsValid   bool     `json:"is_valid"`
		IssuedAt  int64    `json:"issued_at"`
		Scopes    []string `json:"scopes"`
		UserID    string   `json:"user_id"`
	} `json:"data"`
}

type handJoy struct {
	appID       string
	appSecret   string
	accessToken string
}

func newHandJoy() *handJoy {
	s := &handJoy{
		appID:     "740714252962921",
		appSecret: "df65cff17b0ee0542dec4af08ed768af",
	}
	s.accessToken = fmt.Sprintf("%s|%s", s.appID, s.appSecret)
	return s
}

func (s *handJoy) LoginAuth(channelUid, token string) error {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/debug_token", nil)
	if err != nil {
		return err
	}
	cli := &http.Client{}
	q := req.URL.Query()
	q.Add("access_token", s.accessToken)
	q.Add("input_token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.Errorf("handJoy LoginAuth StatusCode %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	reply := &facebookLoginReply{}
	if err := json.Unmarshal(body, reply); err != nil {
		glog.Errorf("handJoy LoginAuth json.Unmarshal error body=%s, err=%s", body, err)
		return err
	}

	if !reply.Data.IsValid {
		return errors.Errorf("handJoy LoginAuth not IsValid")
	}
	if reply.Data.UserID != channelUid {
		return errors.Errorf("handJoy LoginAuth valid channelUid, channelUid=%s, reply.Data.UserID=%s", channelUid, reply.Data.UserID)
	}
	if reply.Data.AppID != s.appID {
		return errors.Errorf("handJoy LoginAuth not valid appID, appID=%s, appID2=%s", s.appID, reply.Data.AppID)
	}
	glog.Infof("handJoy LoginAuth ok %s", body)

	return nil
}

func (s *handJoy) RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
	paymentAmount int, reply []byte, needCheckMoney, ok bool) {
	return
}
