package seasonpvp

import (
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"strconv"
	"time"
)

func NewSeasonPvpHd(area int, attr *attribute.AttrMgr, gdata interface{}) htypes.IHuodong {
	if attr == nil || gdata == nil {
		return nil
	}

	data, ok := gdata.(gamedata.ISeasonPvp)
	if !ok || data == nil {
		return nil
	}

	hd := &seasonPvpHd{
		area:               area,
		sessionDatas:       map[int]*seasonPvpHdSessionData{},
		loadingSessionData: map[int]chan struct{}{},
	}
	hd.I = hd
	hd.SetHtype(pb.HuodongTypeEnum_HSeasonPvp)
	hd.Attr = attr
	hd.SetBeginTime(data.GetStartTime())
	hd.SetEndTime(data.GetStopTime())
	if hd.GetVersion() <= 0 {
		hd.SetVersion(1)
	}
	if !time.Now().Before(hd.GetEndTime()) {
		hd.SetClose(true)
	}
	glog.Infof("NewSeasonPvpHd %s", hd)
	return hd
}

func NewSeasonPvpHdByAttr(area int, attr *attribute.AttrMgr) *seasonPvpHd {
	hd := &seasonPvpHd{
		area:               area,
		sessionDatas:       map[int]*seasonPvpHdSessionData{},
		loadingSessionData: map[int]chan struct{}{},
	}
	hd.I = hd
	hd.SetHtype(pb.HuodongTypeEnum_HSeasonPvp)
	hd.Attr = attr
	hd.SetTime()
	return hd
}

func GetSeasonPvpHandCardInfo(player types.IPlayer) (gamedata.ISeasonPvp, int, pb.BattleHandType, *pb.SeasonPvpChooseCardData,
	*pb.FetchSeasonHandCardReply) {

	hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
	if hd == nil || hd.IsClose() {
		return nil, 0, pb.BattleHandType_UnknowType, nil, nil
	}
	playerData, ok := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(
		pb.HuodongTypeEnum_HSeasonPvp).(*seasonPvpHdPlayerData)
	session := hd.GetVersion()
	if !ok || playerData == nil || playerData.getSession() != session || playerData.isQuit() {
		return nil, 0, pb.BattleHandType_UnknowType, nil, nil
	}

	seasonHd, ok := hd.(*seasonPvpHd)
	if !ok {
		return nil, 0, pb.BattleHandType_UnknowType, nil, nil
	}
	seasonData := seasonHd.getGameData()
	if seasonData == nil || len(seasonData.GetHandCardType()) <= 0 ||
		pb.BattleHandType(seasonData.GetHandCardType()[0]) != pb.BattleHandType_Random {
		return seasonData, 0, pb.BattleHandType_UnknowType, nil, nil
	}

	handCards := playerData.getHandCards()
	if len(handCards) <= 0 {
		return seasonData, playerData.getCamp(), pb.BattleHandType_Random, playerData.getChooseCardData(seasonData), nil
	}

	return seasonData, playerData.getCamp(), pb.BattleHandType_Random, nil, &pb.FetchSeasonHandCardReply{
		ChangeType:   pb.FetchSeasonHandCardReply_ChangeTypeEnum(seasonData.GetChangeHandType()[0]),
		CardIDs:      handCards,
		ChangeMaxPro: int32(seasonData.GetChangeHandType()[1]),
		ChangeCurPro: int32(playerData.getHandCardCurPro()),
		WinCnt:       int32(playerData.getHandCardWinCnt()),
	}
}

func SeasonPvpChooseCamp(player types.IPlayer, camp int) *pb.SeasonPvpChooseCardData {
	hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
	if hd == nil {
		return nil
	}
	playerData := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(
		pb.HuodongTypeEnum_HSeasonPvp).(*seasonPvpHdPlayerData)
	playerData.chooseCamp(camp)
	seasonData := hd.(*seasonPvpHd).getGameData()
	return playerData.randomChooseCards(seasonData)
}

func SeasonPvpChooseCard(player types.IPlayer, cards []uint32) (randCards []uint32, err error) {
	hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
	if hd == nil {
		return
	}
	playerData := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(
		pb.HuodongTypeEnum_HSeasonPvp).(*seasonPvpHdPlayerData)
	randCards, err = playerData.chooseHandCards(hd.(*seasonPvpHd).getGameData(), cards)
	return
}

func RefreshSeasonPvpChooseCard(player types.IPlayer) (*pb.SeasonPvpChooseCardData, error) {
	hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
	if hd == nil {
		return nil, gamedata.GameError(1)
	}

	playerData := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(
		pb.HuodongTypeEnum_HSeasonPvp).(*seasonPvpHdPlayerData)
	var needJade int
	freeRefreshChooseCardCnt := playerData.getFreeRefreshChooseCardCnt()
	if freeRefreshChooseCardCnt <= 0 {
		needJade = getSeasonPvpRefreshChooseCardJade()
	}

	if needJade > 0 {
		jadeRefreshChooseCardCnt := playerData.getJadeRefreshChooseCardCnt()
		if jadeRefreshChooseCardCnt <= 0 {
			return nil, gamedata.GameError(2)
		}

		if !player.HasBowlder(needJade) {
			return nil, gamedata.GameError(3)
		}
		player.SubBowlder(needJade, consts.RmrRefreshSeasonPvp)
		playerData.setJadeRefreshChooseCardCnt(jadeRefreshChooseCardCnt - 1)

		module.Shop.LogShopBuyItem(player, "seasonRefresh", "锦标赛刷卡", 1, "gameplay",
			strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), needJade, "")
	} else {
		playerData.setFreeRefreshChooseCardCnt(freeRefreshChooseCardCnt - 1)
	}

	seasonData := hd.(*seasonPvpHd).getGameData()
	playerData.refreshChooseCards(seasonData)
	return playerData.getChooseCardData(seasonData), nil
}

func GetSeasonPvpWinCnt(player types.IPlayer) int {
	hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
	if hd == nil {
		return 0
	}

	playerData, ok := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(
		pb.HuodongTypeEnum_HSeasonPvp).(*seasonPvpHdPlayerData)
	if !ok {
		return 0
	}
	return playerData.getWinAmount()
}
