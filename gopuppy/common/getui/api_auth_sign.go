package getui

import (
	"encoding/json"
)

var authToken string

type SignParam struct {
	Sign      string `json:"sign"`
	Timestamp string `json:"timestamp"`
	AppKey    string `json:"appkey"`
}

//token
type SignResult struct {
	Result    string `json:"result"`
	AuthToken string `json:"auth_token"`
}

//获取Auth签名
//http://docs.getui.com/getui/server/rest/other_if/
func GetGeTuiToken(force bool) (string, error) {
	if config == nil {
		return "", noConfigError
	}

	if authToken != "" && !force {
		return authToken, nil
	}

	signStr, timestamp := Signature(config.AppKey, config.MasterSecret)

	params := &SignParam{
		Sign:      signStr,
		Timestamp: timestamp,
		AppKey:    config.AppKey,
	}

	url := API_URL + "auth_sign"
	result, err := sendPost(url, "", params)
	if err != nil {
		return "", err
	}

	tokenResult := &SignResult{}
	if err := json.Unmarshal(result, &tokenResult); err != nil {
		return "", err
	}

	authToken = tokenResult.AuthToken
	return tokenResult.AuthToken, nil
}

type iGetuiApiResult interface {
	GetResult() string
}

func callGetuiApi(api string, params interface{}, result iGetuiApiResult) error {
	if config == nil {
		return noConfigError
	}

	for _, forceGetToken := range []bool{false, true} {
		authToken, err := GetGeTuiToken(forceGetToken)
		if err != nil {
			return err
		}

		byteResult, err := sendPost(API_URL+api, authToken, params)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(byteResult, result); err != nil {
			return err
		}

		apiResult := result.GetResult()
		if apiResult == "not_auth" || apiResult == "taginvalid_or_noauth" || apiResult == "other_error" {
			continue
		}
		return nil
	}
	return nil
}
