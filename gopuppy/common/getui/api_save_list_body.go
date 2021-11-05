package getui

//消息应用模板 notification、link、notypopload、transmission 四种类型选其一该属性与message下面的msgtype一致
type SaveListBodyParam struct {
	Message      *Message      `json:"message"`                //消息内容
	Notification *Notification `json:"notification,omitempty"` //通知模板
	Link         *Link         `json:"link,omitempty"`
	Notypopload  *NotyPopload  `json:"notypopload,omitempty"`
	Transmission *Transmission `json:"transmission,omitempty"` //透传模板
	PushInfo     *PushInfo     `json:"push_info,omitempty"`    //json串，当手机为ios，并且为离线的时候；或者简化推送的时候，使用该参数
	TaskName     string        `json:"task_name,omitempty"`    //	任务名称 可以给多个任务指定相同的task_name，后面用task_name查询推送结果能得到多个任务的结果  可选
}

type SaveListBodyResult struct {
	Result string `json:"result"`
	TaskId string `json:"taskid"` //	任务标识号
	Desc   string `json:"desc"`   //	错误信息描述
}

func (this *SaveListBodyResult) GetResult() string {
	return this.Result
}

//保存消息共同体,获取推送taskid
func SaveListBody(param *SaveListBodyParam) (*SaveListBodyResult, error) {
	result := &SaveListBodyResult{}
	err := callGetuiApi("save_list_body", param, result)
	return result, err
}
