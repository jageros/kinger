package getui

//http://docs.getui.com/getui/server/rest/push/#doc-title-1
type PushListParam struct {
	Cid        []string `json:"cid"` //cid为cid list，与alias list二选一
	Alias      []string `json:"alias"`
	Taskid     string   `json:"taskid"`
	NeedDetail bool     `json:"need_detail"`
}

type PushListResult struct {
	Result       string            `json:"result"`
	Taskid       string            `json:"taskid"`
	Desc         string            `json:"desc"`
	CidDetails   map[string]string `json:"cid_details"`
	AliasDetails map[string]string `json:"alias_details"`
}

func (this *PushListResult) GetResult() string {
	return this.Result
}

//群推
func PushList(param *PushListParam) (*PushListResult, error) {
	result := &PushListResult{}
	err := callGetuiApi("push_list", param, result)
	return result, err
}
