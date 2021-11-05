package getui

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"kinger/gopuppy/common/glog"
)

//post请求
func sendPost(url string, authToken string, params interface{}) (result []byte, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var bodyByte []byte
	bodyByte, err = json.Marshal(params)
	if err != nil {
		return
	}
	body := bytes.NewBuffer(bodyByte)

	var req *http.Request
	req, err = http.NewRequest("POST", url, body)
	if err != nil {
		return
	}

	req.Header.Add("authtoken", authToken)
	req.Header.Add("Charset", "UTF-8")
	req.Header.Add("Content-Type", "application/json")

	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	//读取响应
	result, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		glog.Errorf("request getui fail %s", resp)
		return
	}

	return
}

//生成请求参数对应的JSON
func makeReqBody(parmar interface{}) ([]byte, error) {
	body, err := json.Marshal(parmar)
	if err != nil {
		return nil, err
	}

	return body, nil
}
