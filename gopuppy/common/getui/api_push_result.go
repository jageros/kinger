package getui

type PushResultParam struct {
	Taskidlist []string `json:"taskIdList"` //查询的任务结果列表
}

type GtFeedBack struct {
	Feedback  int    `json:"feedback"`
	Displayed int    `json:"displayed"`
	Result    string `json:"result"`
	Sent      int    `json:"sent"`
	Clicked   int    `json:"clicked"`
}

type PushResult struct {
	Result string `json:"result"`
	Data   []struct {
		Taskid     string `json:"taskId"`
		MsgTotal   int    `json:"msgTotal"`
		MsgProcess int    `json:"msgProcess"`
		ClickNum   int    `json:"clickNum"`
		PushNum    int    `json:"pushNum"`
		APN        string `json:"APN,omitempty"`
		GT         string `json:"GT"`
	} `json:"data"`
}

func (this *PushResult) GetResult() string {
	return this.Result
}

//获取推送结果接口
func GetPushResult(param *PushResultParam) (*PushResult, error) {
	result := &PushResult{}
	err := callGetuiApi("push_result", param, result)
	return result, err
}
