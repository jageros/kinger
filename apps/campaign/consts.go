package main

const (
	version = 1
	// 城市数据版本
	cdVersion          = 1
	pageAmount         = 10
	noticeTimeout      = 24 * 60 * 60 * 3
	maxWarContribution = 500
	captiveTimeout     = 36 * 3600
)

const (
	// 资源类型
	resAgriculture = "1" // 农业
	resBusiness    = "2" // 商业
	resDefense     = "3" // 城防
	resForage      = "4" // 粮草
	resGold        = "5" // 金币
)

const (
	createCountryEndHour = 22
	createCountryEndMin  = 0

	coutryWarReadyBeginHour = 20
	coutryWarReadyBeginMin  = 30

	coutryWarBeginHour = 21
	coutryWarBeginMin  = 0

	coutryWarEndHour = 22
	coutryWarEndMin  = 0
)

const (
	maxCounsellorAmount   = 1
	maxGeneralAmount      = 2
	maxPrefectAmount      = 1
	maxDuWeiAmount        = 2
	maxFieldOfficerAmount = 5
)

const (
	// event
	evPlayerLogin = 2000 + iota
	evPlayerChangeCity
	evPlayerChangeCountry
	evPlayerChangeCityJob
	evWarReady
	evWarBegin
	evWarEnd
	evUnified
	evCityChangeCountry
)

const (
	ttSupport = 1 + iota
	ttExpedition
	ttDefCity
)
