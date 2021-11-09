package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kinger/gopuppy/common/glog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"kinger/gopuppy/common"
	"kinger/gopuppy/common/async"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/evq"
)

func post(type_ string, uid common.UUid, seq uint32, msgId, errcode int32, payload []byte) {
	if config.GetConfig().MonitorUrl == "" {
		return
	}

	ts := time.Now().UnixNano() / 1000000
	//glog.Infof("monitor post type=%s uid=%d seq=%d msgid=%d errcode=%d payload=%v",
	//	type_, uid, seq, msgId, errcode, payload)
	async.AppendAsyncJob("player_monitor", func() (res interface{}, err error) {
		args := map[string]string{
			"type":    type_,
			"uid":     uid.String(),
			"seq":     strconv.Itoa(int(seq)),
			"error":   strconv.Itoa(int(errcode)),
			"payload": url.PathEscape(string(payload)),
			"ts":      common.UUid(ts).String(),
			"msgId":   strconv.Itoa(int(msgId)),
		}

		postData, err := json.Marshal(args)
		if err != nil {
			glog.Errorf("monitor post json.Marshal error %s %v", err, args)
		}

		var buffer bytes.Buffer
		buffer.Write(postData)
		http.Post(config.GetConfig().MonitorUrl+"/monitor/post", "application/json", &buffer)
		return
	}, nil)
}

func Login(uid common.UUid, seq uint32, payload []byte) {
	post("login", uid, seq, 0, 0, payload)
}

func RpcReply(uid common.UUid, msgID int32, seq uint32, errcode int32, payload []byte) {
	post("rpcReply", uid, seq, msgID, errcode, payload)
}

func RpcPush(uid common.UUid, msgId int32, payload []byte) {
	post("rpcPush", uid, 0, msgId, 0, payload)
}

func Operation(uid common.UUid, payload []byte) {
	post("operation", uid, 0, 0, 0, payload)
}

func Begin(uid common.UUid) {
	if uid == 0 {
		return
	}
	if config.GetConfig().MonitorUrl == "" {
		return
	}
	evq.Await(func() {
		http.Get(config.GetConfig().MonitorUrl + fmt.Sprintf("/monitor/begin?uid=%d", uid))
	})
}

func End(uid common.UUid) {
	if uid == 0 {
		return
	}
	if config.GetConfig().MonitorUrl == "" {
		return
	}
	evq.Await(func() {
		http.Get(config.GetConfig().MonitorUrl + fmt.Sprintf("/monitor/end?uid=%d", uid))
	})
}
