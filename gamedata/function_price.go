package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
	"kinger/gopuppy/common/glog"
	"strconv"
	"strings"
	"time"
)

type FunctionPrice struct {
	ID    int    `json:"__id__"`
	Type  string `json:"type"`
	Price int    `json:"price"`
}

type FunctionPriceGameData struct {
	baseGameData
	ModifyName             int
	TreasureUp             int
	CardMore               int
	DailyDouble            int
	LuckyBag               int
	VipAccTicket           int
	AccountTransfer        int
	SpCardToPiece          int
	SkinToPiece            int
	ShopTreasureMaxBuyCnt  int
	ClearLevel             int
	VipContinuedTime       int           // vip持续时间
	PrivContinuedTime      int           // 勋章持续时间
	RankSeasonRefreshPrice int           // 锦标赛刷新卡池
	ShopGoldCD             time.Duration // 商城买金币cd
	ShopTreasureCD         time.Duration // 商城买军备宝箱cd
	TryPrivContinuedTime   int           // 试用勋章持续时间
	TryVipContinuedTime    int           // 试用vip持续时间
	AccTreasure            float64       // 宝箱加速每1元宝多少秒
	PrivLevel2Num          map[int]int   // 特权等级对应的特权个数
	Team2MissionExtReward  map[int]int   // 任务每个段位额外获得
	JadeToGold             int
	LeagueResetRewardProp  int              // 联赛积分补偿百分比
	LeagueResetRewardMax   int              // 最大补偿积分
	LeagueCycle            *LeagueCycleTime //赛季时间
}

type LeagueCycleTime struct {
	TimeType string // 时间类型：M表示月， W表示周，D表示天
	TimeDay  int    // 表示每月几号/星期几/时间类型为天的时候无效
	TimeNum  int    // 表示从此刻开始第几个符合的时间
}

func newFunctionPriceGameData() *FunctionPriceGameData {
	r := &FunctionPriceGameData{}
	r.i = r
	return r
}

func (gd *FunctionPriceGameData) name() string {
	return consts.FunctionPrice
}

func (gd *FunctionPriceGameData) init(d []byte) error {
	var l []*FunctionPrice
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.PrivLevel2Num = map[int]int{}
	var baseMissionExtReward int
	gd.Team2MissionExtReward = map[int]int{}
	team2MissionExtRewardRare := map[int]int{}

	for _, d := range l {
		switch d.Type {
		case "modifyName":
			gd.ModifyName = d.Price
		case "treasureUp":
			gd.TreasureUp = d.Price
		case "cardMore":
			gd.CardMore = d.Price
		case "dailyDouble":
			gd.DailyDouble = d.Price
		case "luckyBag":
			gd.LuckyBag = d.Price
		case "vipAccTicket":
			gd.VipAccTicket = d.Price
		case "accountTransfer":
			gd.AccountTransfer = d.Price
		case "spToPiece":
			gd.SpCardToPiece = d.Price
		case "skinToPiece":
			gd.SkinToPiece = d.Price
		case "shopTreasureMaxBuyCnt":
			gd.ShopTreasureMaxBuyCnt = d.Price
		case "clearLevel":
			gd.ClearLevel = d.Price
		case "vipContinuedDay":
			gd.VipContinuedTime = d.Price * 24 * 3600
		case "privContinuedDay":
			gd.PrivContinuedTime = d.Price * 24 * 3600
		case "shopGoldCd":
			gd.ShopGoldCD = time.Duration(d.Price) * time.Hour
		case "shopTreasureCd":
			gd.ShopTreasureCD = time.Duration(d.Price) * time.Hour
		case "rankSeasonRefreshPrice":
			gd.RankSeasonRefreshPrice = d.Price
		case "tryPriv":
			gd.TryPrivContinuedTime = d.Price * 24 * 3600
		case "tryVip":
			gd.TryVipContinuedTime = d.Price * 24 * 3600
		case "accTreasure":
			gd.AccTreasure = float64(d.Price) * 60
		case "questRewardRate":
			baseMissionExtReward = d.Price
		case "jadeToGold":
			gd.JadeToGold = d.Price
		case "leagueResetRewardProp":
			gd.LeagueResetRewardProp = d.Price
		case "leagueResetRewardMax":
			gd.LeagueResetRewardMax = d.Price

		default:

			ty := strings.Split(d.Type, "_")
			switch ty[0] {
			case "privLevel":
				lv, err := strconv.Atoi(ty[1])
				if err != nil {
					glog.Infof("function_price init PrivLevel2Num err=%s, arg2=%s", err, ty[1])
				} else {
					gd.PrivLevel2Num[lv] = d.Price
				}

			case "questRewardTeam":
				team, _ := strconv.Atoi(ty[1])
				team2MissionExtRewardRare[team] = d.Price

			case "leagueCycleTime":
				day, _ := strconv.Atoi(ty[2])
				gd.LeagueCycle = &LeagueCycleTime{
					TimeType: ty[1],
					TimeDay:  day,
					TimeNum:  d.Price,
				}
			}

		}
	}

	for team, rate := range team2MissionExtRewardRare {
		gd.Team2MissionExtReward[team] = baseMissionExtReward * rate
	}

	return nil
}
