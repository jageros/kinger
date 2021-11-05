package wxgame

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
	"time"
	"github.com/gogo/protobuf/proto"
	"kinger/gopuppy/common/glog"
	"kinger/apps/game/module"
	"kinger/gopuppy/common/utils"
	"kinger/gopuppy/common"
)

type wxgameComponent struct {
	player types.IPlayer
	attr   *attribute.MapAttr
	dailyShareAttr *attribute.MapAttr
}

func (wc *wxgameComponent) ComponentID() string {
	return consts.WxgameCpt
}

func (wc *wxgameComponent) GetPlayer() types.IPlayer {
	return wc.player
}

func (wc *wxgameComponent) OnInit(player types.IPlayer) {
	wc.player = player
}

func (wc *wxgameComponent) OnLogin(isRelogin, isRestore bool) {
	if isRelogin {
		return
	}
	shareCntDay := wc.attr.GetInt("shareCntDay")
	dayno := timer.GetDayNo()
	if shareCntDay != dayno {
		wc.attr.SetInt("shareCntDay", dayno)
		wc.attr.SetInt("triggerTreasureShareHDCnt", 0)
		resCpt := wc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		wxCfg := config.GetConfig().Wxgame
		//resCpt.SetResource(consts.AccTreasureCnt, module.Reborn.GetAccTreasureCntPriv(wc.player, wxCfg.AccTreasureCnt))
		resCpt.SetResource(consts.NotSubStarCnt, wxCfg.NotSubStarCnt)
	}

	if wc.dailyShareAttr == nil {
		wc.dailyShareAttr = wc.attr.GetMapAttr("dailyShare")
		if wc.dailyShareAttr == nil {
			wc.dailyShareAttr = attribute.NewMapAttr()
			wc.attr.SetMapAttr("dailyShare", wc.dailyShareAttr)
		}
	}
}

func (wc *wxgameComponent) OnLogout() {
	wc.endShareBattleLose()
}

func (wc *wxgameComponent) getShareTreasureReward(hid int) *pb.OpenTreasureReply {
	hdData, ok := gamedata.GetGameData(consts.TreasureShare).(*gamedata.TreasureShareGameData).Id2TreasureShare[hid]
	if !ok {
		return &pb.OpenTreasureReply{}
	}

	shareTreasureHdAttr := wc.attr.GetMapAttr("shareTreasureHd")
	if shareTreasureHdAttr == nil {
		shareTreasureHdAttr = attribute.NewMapAttr()
		wc.attr.SetMapAttr("shareTreasureHd", shareTreasureHdAttr)
	}

	key := strconv.Itoa(hid)
	if shareTreasureHdAttr.GetBool(key) {
		return &pb.OpenTreasureReply{}
	}
	shareTreasureHdAttr.SetBool(key, true)

	return wc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(hdData.Reward, false)
}

func (wc *wxgameComponent) shareTreasure(treasureID uint32, wxGroupID string) error {
	treasureCpt := wc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent)
	if treasureCpt.IsDailyTreasure(treasureID) {
		return nil
	}

	resCpt := wc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if resCpt.GetResource(consts.AccTreasureCnt) <= 0 {
		return gamedata.GameError(2)
	}

	if wxGroupID != "" {
		shareTreasureGroupsAttr := wc.attr.GetMapAttr("shareTreasureGroups")
		if shareTreasureGroupsAttr == nil {
			shareTreasureGroupsAttr = attribute.NewMapAttr()
			wc.attr.SetMapAttr("shareTreasureGroups", shareTreasureGroupsAttr)
		}

		now := time.Now().Unix()
		lastShareTime := shareTreasureGroupsAttr.GetInt64(wxGroupID)
		if now-lastShareTime < 600 {
			return gamedata.GameError(1)
		}

		shareTreasureGroupsAttr.SetInt64(wxGroupID, now)
	}
	return nil
}

func (wc *wxgameComponent) shareBattleLose(wxGroupID string) error {
	resCpt := wc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if resCpt.GetResource(consts.NotSubStarCnt) <= 0 {
		return gamedata.GameError(2)
	}

	shareBattleLoseGroupsAttr := wc.attr.GetMapAttr("shareBattleLoseGroups")
	if shareBattleLoseGroupsAttr == nil {
		shareBattleLoseGroupsAttr = attribute.NewMapAttr()
		wc.attr.SetMapAttr("shareBattleLoseGroups", shareBattleLoseGroupsAttr)
	}

	now := time.Now().Unix()
	lastShareTime := shareBattleLoseGroupsAttr.GetInt64(wxGroupID)
	if now-lastShareTime < 2 * 60 * 60 {
		return gamedata.GameError(1)
	}

	shareBattleLoseGroupsAttr.SetInt64(wxGroupID, now)
	wc.attr.SetBool("shareBattleLose", true)
	return nil
}

func (wc *wxgameComponent) endShareBattleLose() {
	wc.attr.Del("shareBattleLose")
}

func (wc *wxgameComponent) helpShareBattleLose() {
	if wc.attr.GetBool("shareBattleLose") {
		wc.endShareBattleLose()
		resCpt := wc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		resCpt.BatchModifyResource(map[int]int{
			consts.Score:         1,
			consts.NotSubStarCnt: -1,
		})
		wc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_LOSE_BE_HELP, &pb.BattleLoseBeHelp{
			AddStar: 1,
		})
	}
}

func (wc *wxgameComponent) TriggerTreasureShareHD(treasureID string) int {
	cnt := wc.attr.GetInt("triggerTreasureShareHDCnt")
	if cnt >= config.GetConfig().Wxgame.TriggerTreasureShareHDCnt {
		return 0
	}

	now := time.Now()
	for i := 0; i < len(treasureShareHDs); {
		hd := treasureShareHDs[i]
		if !hd.isOpen(now) {
			treasureShareHDs = append(treasureShareHDs[:i], treasureShareHDs[i+1:]...)
		} else {
			i++
			if hd.canTrigger(treasureID) {
				wc.attr.SetInt("triggerTreasureShareHDCnt", cnt+1)
				return hd.id
			}
		}
	}
	return 0
}

func (wc *wxgameComponent) OnShareBeHelp(shareType int, shareTime int64, data []byte) {
	module.Mission.OnWxShare(wc.player, shareTime)

	var reply proto.Marshaler = nil
	var err error = nil
	switch shareType {
	case stShopGoldAds:
		fallthrough
	case stShopJadeAds:
		fallthrough
	case stShopTreasureAds:
		arg := &pb.WatchShopFreeAdsArg{}
		arg.Unmarshal(data)
		err = wc.player.GetComponent(consts.ShopCpt).(types.IShopComponent).OnShopAdsBeHelp(arg.Type, int(arg.ID))
		reply = arg

	case stUpTreasureRareAds:
		reply, err = wc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).UpTreasureRare(false)

	case stTreasureAddCardAds:
		arg := &pb.TargetTreasure{}
		arg.Unmarshal(data)
		reply, err = wc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).WatchTreasureAddCardAds(
			arg.TreasureID, false)

	case stDailyShare:
		reply = wc.onDailyShareBeHelp(shareTime, data)

	case stDailyTreasureDouble:
		arg := &pb.GWxDailyTreasureShare{}
		arg.Unmarshal(data)
		if module.Treasure.WxHelpDoubleDailyTreasure(wc.player, arg.TreasureID, common.UUid(arg.HelperUid), arg.HelperHeadImg,
			arg.HelperHeadFrame, arg.HelperName) {
			reply = &pb.DailyTreasureShareInfo{
				HelperUid: arg.HelperUid,
				HelperHeadImg: arg.HelperHeadImg,
				HelperHeadFrame: arg.HelperHeadFrame,
				HelperName: arg.HelperName,
			}
		}
	}

	if err == nil && reply != nil {
		msg := &pb.WxShareBeHelpArg{
			ShareType: int32(shareType),
		}
		msg.Data, _ = reply.Marshal()
		wc.player.GetAgent().PushClient(pb.MessageID_S2C_WX_SHARE_BE_HELP, msg)
	}
}

func (wc *wxgameComponent) onShare(shareType int, wxGroupID string, data []byte) error {
	if wxGroupID == "" {
		return nil
	}

	shareGroupsAttr := wc.attr.GetMapAttr("shareGroups")
	if shareGroupsAttr == nil {
		shareGroupsAttr = attribute.NewMapAttr()
		wc.attr.SetMapAttr("shareGroups", shareGroupsAttr)
	}
	key := strconv.Itoa(shareType)
	attr := shareGroupsAttr.GetMapAttr(key)
	if attr == nil {
		attr = attribute.NewMapAttr()
		shareGroupsAttr.SetMapAttr(key, attr)
	}

	now := time.Now().Unix()
	lastShareTime := attr.GetInt64(wxGroupID)
	if now-lastShareTime < 2 * 60 * 60 {
		return gamedata.GameError(1)
	}
	attr.SetInt64(wxGroupID, now)

	//wxconfig := config.GetConfig().Wxgame
	//if wxconfig.IsExamined {
	//	timer.AfterFunc(time.Duration(wxconfig.DelayRewardTime)*time.Second, func() {
	//		wc.OnShareBeHelp(shareType, data)
	//	})
	//}

	return nil
}

func (wc *wxgameComponent) getDailyShareReward() (jade, bowlder, ticket int, err error) {
	day := wc.attr.GetInt("lastIosShareDay")
	curDay := timer.GetDayNo()
	if day != curDay {
		ticket = 2
		wc.attr.SetInt("lastIosShareDay", curDay)
		glog.Infof("getDailyShareReward uid=%d", wc.player.GetUid())
		wc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).BatchModifyResource(map[int]int{
			consts.AccTreasureCnt: ticket,
			//consts.Bowlder: bowlder,
		}, consts.RmrDailyShare)
		return
	} else {
		err = gamedata.GameError(1)
		return
	}
}

func (wc *wxgameComponent) onDailyShareBeHelp(shareTime int64, data []byte) proto.Marshaler {
	if !utils.IsSameDay(shareTime, time.Now().Unix()) {
		return nil
	}

	state := wc.getDailyShareState()
	if state != dstCantReward {
		return nil
	}

	glog.Infof("onDailyShareBeHelp uid=%d", wc.player.GetUid())
	arg := &pb.GWxDailyShare{}
	arg.Unmarshal(data)
	dailyShareAttr := wc.getDailyShareAttr()
	dailyShareAttr.SetBool("canReward", true)
	dailyShareAttr.SetUInt64("uid", uint64(arg.Uid))
	dailyShareAttr.SetStr("headImg", arg.HeadImg)
	dailyShareAttr.SetStr("name", arg.Name)
	return &pb.WxDailyShareReply{
		PlayerName: arg.Name,
	}
}

func (wc *wxgameComponent) getDailyShareHelperUid() common.UUid {
	return common.UUid(wc.dailyShareAttr.GetUInt64("uid"))
}

func (wc *wxgameComponent) getDailyShareAttr() *attribute.MapAttr {
	dayno := timer.GetDayNo()
	if wc.dailyShareAttr.GetInt("day") != dayno {
		wc.dailyShareAttr.SetInt("day", dayno)
		wc.dailyShareAttr.SetBool("canReward", false)
		wc.dailyShareAttr.SetInt("returnCnt", 0)
	}
	return wc.dailyShareAttr
}

func (wc *wxgameComponent) getDailyShareState() int {
	dayno := timer.GetDayNo()
	if wc.attr.GetInt("lastIosShareDay") == dayno {
		return dstHasReward
	}
	dailyShareAttr := wc.getDailyShareAttr()
	if dailyShareAttr.GetBool("canReward") {
		return dstCanReward
	}
	return dstCantReward
}

func (wc *wxgameComponent) packDailyShareMsg() *pb.DailyShareInfo {
	state := wc.getDailyShareState()
	msg := &pb.DailyShareInfo{
		RewardState: int32(state),
	}

	if state != dstCantReward {
		msg.HelperUid = wc.dailyShareAttr.GetUInt64("uid")
		msg.HelperHeadImg = wc.dailyShareAttr.GetStr("headImg")
		msg.HelperName = wc.dailyShareAttr.GetStr("name")
	}
	return msg
}

func (wc *wxgameComponent) returnDailyShareReward(playerName string) {
	var gold int
	dailyShareAttr := wc.getDailyShareAttr()
	cnt :=  dailyShareAttr.GetInt("returnCnt")
	if cnt < dailyShareReturnMaxCnt {
		gold = dailyShareReturnGold
		cnt++
		dailyShareAttr.SetInt("returnCnt", cnt)
		glog.Infof("returnDailyShareReward uid=%d, playerName=%s, cnt=%d", wc.player.GetUid(), playerName, cnt)
		module.Player.ModifyResource(wc.player, consts.Gold, gold, consts.RmrDailyShareReturn)
	}

	agent := wc.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_DAILY_SHARE_RETURN_REWARD, &pb.DailyShareReturnReward{
			PlayerName: playerName,
			Gold: int32(gold),
		})
	}
}
