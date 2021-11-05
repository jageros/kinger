package types

import (
	"kinger/gopuppy/common/utils"
)

var (
	ChristmasBegin, _ = utils.StringToTime("2019-02-04 00:00:00", utils.TimeFormat2)
	ChristmasEnd, _   = utils.StringToTime("2019-02-10 23:59:59", utils.TimeFormat2)
)

const (
	SeasonPvpRefreshChooseCardJade = 20
	//RechargeHdKey = "yuandan"
	RechargeHdKey = "spring"
)
