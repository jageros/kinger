package gamedata

import (
	//pconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"

	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/rpubsub"
	"kinger/gopuppy/attribute"
	"kinger/common/config"
	"kinger/common/consts"
)

var (
	allGameData  = make(map[string]IGameData)
	allGameDataList []IGameData
	gameDataPath = "jsondata"
	handjoyGameDataName = map[string]string{}
	OnReload func()
)

type IGameData interface {
	load() error
	reload() error
	init([]byte) error
	name() string
	AddReloadCallback(func(data IGameData))
}

type baseGameData struct {
	i               IGameData
	reloadCallbacks []func(data IGameData)
}

func (g *baseGameData) AddReloadCallback(f func(data IGameData)) {
	g.reloadCallbacks = append(g.reloadCallbacks, f)
}

func (g *baseGameData) load() error {
	attr := attribute.NewAttrMgr("gamedata", wrapName( g.i.name() ), true)
	err := attr.Load()
	if err != nil {
		glog.Errorf("load %s ReadFile error, %s", wrapName( g.i.name() ), err)
		return err
	}

	if err := g.i.init([]byte(attr.GetStr("data"))); err != nil {
		glog.Errorf("load %s init error, %s", wrapName( g.i.name() ), err)
		return err
	}

	return nil
}

func (g *baseGameData) reload() error {
	if err := g.load(); err != nil {
		return err
	}

	for _, f := range g.reloadCallbacks {
		f(g.i)
	}

	return nil
}

func GetGameData(name string) IGameData {
	if d, ok := allGameData[wrapName(name)]; ok {
		return d
	} else {
		return nil
	}
}

func SetGameDataPath(path string) {
	gameDataPath = path
}

func onReload(ev evq.IEvent) {
	args := ev.(*evq.CommonEvent).GetData()
	c := args[0].(chan error)

	for _, data := range allGameData {
		err := data.reload()
		if err != nil {
			c <- err
			return
		}
	}

	c <- nil
}

func addGameData(gdata IGameData) {
	name := wrapName( gdata.name() )
	if _, ok := allGameData[name]; !ok {
		allGameData[name] = gdata
		allGameDataList = append(allGameDataList, gdata)
	}
}

func wrapName(name string) string {
	if config.GetConfig().IsMultiLan {
		if name2, ok := handjoyGameDataName[name]; ok {
			return name2
		}
	}
	return name
}

func Load() {
	handjoyGameDataName[consts.CardSkin] = consts.CardSkin + "_handjoy"
	handjoyGameDataName[consts.HeadFrame] = consts.HeadFrame + "_handjoy"
	handjoyGameDataName[consts.SeasonReward] = consts.SeasonReward + "_handjoy"
	handjoyGameDataName[consts.HuodongConfig] = consts.HuodongConfig + "_handjoy"
	handjoyGameDataName[consts.HuodongReward] = consts.HuodongReward + "_handjoy"

	//evq.HandleEvent(pconsts.GAME_DATA_RELOAD_EVENT, onReload)
	rpubsub.Subscribe("reload_json", func(i map[string]interface{}) {
		allRewardTblNames := getAllRewardTblName()
		for _, name := range allRewardTblNames {
			if _, ok := allGameData[name]; !ok {
				addGameData( newRewardTblGameData(name) )
			}
		}

		for i := 0; i < len(allGameDataList); i++ {
			data := allGameDataList[i]
			if err := data.reload(); err != nil {
				glog.Errorf("gamedata %s reload err %s", data.name(), err)
			} else {
				glog.Infof("gamedata %s reload ok", data.name())
			}
		}

		if OnReload != nil {
			OnReload()
		}
	})

	//allGameData[consts.Duel] = newDuelGameData()
	//allGameData[consts.FreeJddeAds] = newFreeJadeAdsGameData()
	addGameData( newAreaConfigGameData() )
	addGameData( newExchangeGameData() )
	addGameData( newTextGameData() )
	addGameData( newLevelGameData() )
	addGameData( newPoolGameData() )
	addGameData( newBonusGameData() )
	//addGameData( newDiyGameData() )
	addGameData( newSkillGameData() )
	addGameData( newTargetGameData() )
	addGameData( newTreasureGameData() )
	addGameData( newTreasureRewardFakeGameData() )
	addGameData( newTreasureDailyFakeGameData() )
	addGameData( newRankGameData() )
	addGameData( newTutorialGameData() )
	addGameData( name1 )
	addGameData( name2 )
	addGameData( name3 )
	addGameData( name4 )
	addGameData( name5 )
	addGameData( newGiftCodeGameData() )
	addGameData( newTreasureShareGameData() )
	addGameData( newIosRechargeGameData() )
	addGameData( newLimitGiftGameData() )
	addGameData( newSoldTreasureGameData() )
	addGameData( newSoldGoldGameData() )
	addGameData( newFreeGoldAdsGameData() )
	addGameData( newFreeTreasureAdsGameData() )
	addGameData( newFreeGoodTreasureAdsGameData() )
	addGameData( newMissionGameData() )
	addGameData( newMissionTreasureGameData() )
	addGameData( newNewbiePvpGameData() )
	addGameData( newWxInviteRewardGameData() )
	addGameData( newWxRechargeGameData() )
	addGameData( newWxLimitGiftGameData() )
	addGameData( newAndroidRechargeGameData() )
	addGameData( newAndroidLimitGiftGameData() )
	addGameData( newSeasonPvpGameData() )
	addGameData( newHeadFrameGameData() )
	addGameData( newCardSkinGameData() )
	addGameData( newSeasonRewardGameData() )
	addGameData( newRebornSoldCardGameData() )
	addGameData( newRebornSoldPrivGameData() )
	addGameData( newRebornSoldSkinGameData() )
	addGameData( newRebornCardCaculGameData() )
	addGameData( newRebornGoldCaculGameData() )
	addGameData( newRebornTreausreGameData() )
	addGameData( newEquipGameData() )
	addGameData( newEmojiGameData() )
	addGameData( newRebornSoldEquipGameData() )
	addGameData( newCityGameData() )
	addGameData( newRoadGameData() )
	addGameData( newCampaignParamGameData() )
	addGameData( newRebornCntGameData() )
	addGameData( newHuodongGameData() )
	addGameData( newHuodongRewardGameData() )
	addGameData( newLuckBagRewardGameData() )
	addGameData( newFunctionPriceGameData() )
	addGameData( newPieceCardGameData() )
	addGameData( newPieceSkinGameData() )
	addGameData( newAiMatchGameData() )
	addGameData( newRecruitTreasureGameData() )
	addGameData(newActivityGameData())
	addGameData(newActivityOpenConditionGameData())
	addGameData(newActivityTimeGameData())
	addGameData( newRandomShopGameData() )
	addGameData( newMatchParamGameData() )
	addGameData( newTreasureEventGameData() )
	addGameData( newPrivilegeGameData() )
	addGameData( newChatPopGameData() )
	addGameData( newRankHonorRewardData() )
	addGameData( newSoldGoldGiftGameData() )
	addGameData( newWinningRateGameData() )
	addGameData( newRecruitRefreshConfigGameData() )
	addGameData( newLeagueGameData() )
	addGameData( newLeagueRewardGameData() )
	addGameData( newLeagueRankRewardGameData() )

	if config.GetConfig().IsMultiLan {
		addGameData( nameEn1 )
		addGameData( nameEn2 )
		addGameData( newAndroidHandJoyLimitGiftGameData() )
		addGameData( newIosHandJoyLimitGiftGameData() )
		addGameData( newSoldGoldHandjoyGameData() )
		addGameData( newSoldTreasureHandjoyGameData() )
		addGameData( newMultiLanSeasonPvpGameData() )
	} else {
		addGameData( newWarShopCardGameData() )
		addGameData( newWarShopEquipGameData() )
		addGameData( newWarShopSkinGameData() )
		addGameData( newWarShopResGameData() )
	}

	allRewardTblNames := getAllRewardTblName()
	for _, name := range allRewardTblNames {
		addGameData( newRewardTblGameData(name) )
	}

	for i := 0; i < len(allGameDataList); i++ {
		data := allGameDataList[i]
		if err := data.load(); err != nil {
			glog.Errorf("gamedata %s load err %s", data.name(), err)
		} else {
			glog.Infof("gamedata %s load ok", data.name())
		}
	}
}
