package td

import (
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/async"
	//"strconv"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kinger/gopuppy/common/glog"
	"net/http"
	"time"
)

const jobTdPush = "jobTdPush"

func OnPay(os, tdkey, gameServer, channel string, uid common.UUid, level int, orderID string, money float64,
	jade int, currency ...string) {

	async.AppendAsyncJob(jobTdPush, func() (res interface{}, err error) {
		accountID := fmt.Sprintf("%s_%d", channel, uid)
		args := []map[string]interface{}{map[string]interface{}{
			"msgID":  orderID,
			"status": "success",
			"OS":     os,
			//"accountID": strconv.FormatUint(uint64(uid), 10),
			"accountID":             accountID,
			"orderID":               orderID,
			"currencyAmount":        money,
			"currencyType":          "CNY",
			"virtualCurrencyAmount": float32(jade),
			"chargeTime":            time.Now().Unix() * 1000,
			"gameServer":            "1",
			"level":                 level,
			"partner":               channel,
		}}
		if len(currency) > 0 {
			args[0]["currencyType"] = currency[0]
		}

		postData, err := json.Marshal(args)
		//glog.Infof("OnPay 1111111 %s    %s", tdkey, postData)
		if err != nil {
			glog.Errorf("td OnPay err, args=%v, err=%s", args, err)
			return
		}

		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		_, err = writer.Write(postData)
		if err != nil {
			glog.Errorf("td OnPay gzip err, args=%v, err=%s", args, err)
			writer.Close()
			return
		}
		err = writer.Close()
		if err != nil {
			glog.Errorf("td OnPay gzip err, args=%v, err=%s", args, err)
			return
		}

		url := "http://api.talkinggame.com/api/charge/" + tdkey
		resp, err := http.Post(url, "application/json", &buffer)
		if err != nil {
			glog.Errorf("td OnPay http.Post err, args=%v, err=%s", args, err)
			return
		}

		if resp.StatusCode != 200 {
			glog.Errorf("td OnPay http.Post err, args=%v, status=%d", args, resp.StatusCode)
			return
		}

		_, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			glog.Errorf("td OnPay http.Post readbody err, args=%v, err=%d", args, err)
			return
		}

		//reader, _ := gzip.NewReader(bytes.NewReader(reply))
		//reply2, _ := ioutil.ReadAll(reader)
		//glog.Infof("td OnPay %s", reply2)

		return
	}, nil)
}
