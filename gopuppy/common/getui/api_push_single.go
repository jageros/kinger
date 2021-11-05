package getui

import "fmt"

type PushSingleParam struct {
	Message      *Message      `json:"message"`
	Notification *Notification `json:"notification,omitempty"`
	Link         *Link         `json:"link,omitempty"`
	Notypopload  *NotyPopload  `json:"notypopload,omitempty"`
	Transmission *Transmission `json:"transmission,omitempty"`
	PushInfo     *PushInfo     `json:"push_info,omitempty"`
	Cid          string        `json:"cid,omitempty"`
	Alias        string        `json:"alias,omitempty"`
	RequestId    string        `json:"requestid"`
}

type PushSingleResult struct {
	Result string `json:"result"` //ok 鉴权成功
	TaskId string `json:"taskid"` //任务标识号
	Desc   string `json:"desc"`   //错误信息描述
	Status string `json:"status"` //推送结果successed_offline 离线下发successed_online 在线下发successed_ignore 非活跃用户不下发
}

func (this *PushSingleResult) GetResult() string {
	return this.Result
}

func (this *PushSingleResult) String() string {
	return fmt.Sprintf("result=%s, taskId=%s, desc=%s, status=%s", this.Result, this.TaskId, this.Desc, this.Status)
}

//单推
func PushSingle(param *PushSingleParam) (*PushSingleResult, error) {
	result := &PushSingleResult{}
	err := callGetuiApi("push_single", param, result)
	return result, err
}
