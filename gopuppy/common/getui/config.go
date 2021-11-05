package getui

import "errors"

const (
	BASE_API_URL = "https://restapi.getui.com/v1/"
)

var (
	API_URL = BASE_API_URL
	config *GetuiConfig
	noConfigError = errors.New("getui no config")
)

type GetuiConfig struct {
	AppId string
	AppKey string
	MasterSecret string
}

func Initialize(appId, appKey, masterSecret string) {
	config = &GetuiConfig{
		AppId:        appId,
		AppKey:       appKey,
		MasterSecret: masterSecret,
	}
	API_URL = BASE_API_URL + appId + "/"
}
