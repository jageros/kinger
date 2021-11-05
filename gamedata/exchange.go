package gamedata

import (
	"encoding/json"

	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type Exchange struct {
	ResType int `json:"__id__"`
	Buy     int `json:"buy"`
	Sold    int `json:"sold"`
}

type ExchangeGameData struct {
	baseGameData
	exchangeMap map[int]*Exchange
}

func newExchangeGameData() *ExchangeGameData {
	eg := &ExchangeGameData{}
	eg.i = eg
	return eg
}

func (eg *ExchangeGameData) name() string {
	return consts.Exchange
}

func (eg *ExchangeGameData) init(d []byte) error {
	var _list []*Exchange
	err := json.Unmarshal(d, &_list)
	if err != nil {
		return err
	}

	eg.exchangeMap = make(map[int]*Exchange)
	for _, ex := range _list {
		eg.exchangeMap[ex.ResType] = ex
	}

	//glog.Infof("exchangeMap = %v", eg.exchangeMap)

	return nil
}

func (eg *ExchangeGameData) GetExchangeRes(resType int) *Exchange {
	return eg.exchangeMap[resType]
}
