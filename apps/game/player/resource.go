package player

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
	"kinger/gopuppy/common/timer"
	"time"
	"kinger/gopuppy/common/evq"
)

var needLogRes = map[int]string{}

func init() {
	needLogRes[consts.Gold] = "金币"
	needLogRes[consts.Jade] = "宝玉"
	needLogRes[consts.Reputation] = "声望"
	needLogRes[consts.Bowlder] = "玉石"
	needLogRes[consts.EventItem1] = "春节活动物品"
	needLogRes[consts.AccTreasureCnt] = "加速劵"
	needLogRes[consts.CardPiece] = "武将碎片"
	needLogRes[consts.SkinPiece] = "皮肤碎片"
	needLogRes[consts.Prestige] = "名望"
	needLogRes[consts.Feats] = "功勋"
}

// implement module.IPlayerComponent
type ResourceComponent struct {
	attr   *attribute.MapAttr
	gdata  *gamedata.ExchangeGameData
	player *Player
}

func (rc *ResourceComponent) OnLogin(isRelogin, isRestore bool) {
	if isRelogin {
		return
	}
	rc.OnCrossDay(timer.GetDayNo())
}

func (rc *ResourceComponent) OnCrossDay(dayno int) {
	lastDayno := rc.player.GetDataDayNo()
	if dayno == lastDayno {
		return
	}

	now := time.Now()
	for lastDayno < dayno {
		if now.Weekday() == time.Monday{
			rc.SetResource(consts.CrossAreaHonor, 0)
			return
		}

		lastDayno += 1
		now = now.Add(- 24 * time.Hour)
	}
}

func (rc *ResourceComponent) OnLogout() {
}

func (rc *ResourceComponent) ComponentID() string {
	return consts.ResourceCpt
}

func (rc *ResourceComponent) GetPlayer() types.IPlayer {
	return rc.player
}

func (rc *ResourceComponent) OnInit(player types.IPlayer) {
	rc.gdata = gamedata.GetGameData(consts.Exchange).(*gamedata.ExchangeGameData)
	rc.player = player.(*Player)
	playerAttr := rc.player.getAttr()
	attr := playerAttr.GetMapAttr("resource")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("resource", attr)
	}
	rc.attr = attr
}

func (rc *ResourceComponent) packMsg() []*pb.Resource {
	return []*pb.Resource{
		{int32(consts.Gold), int32(rc.attr.GetInt(strconv.Itoa(consts.Gold)))},
//		{int32(consts.Forage), int32(rc.attr.GetInt(strconv.Itoa(consts.Forage)))},
//		{int32(consts.Weap), int32(rc.attr.GetInt(strconv.Itoa(consts.Weap)))},
//		{int32(consts.Horse), int32(rc.attr.GetInt(strconv.Itoa(consts.Horse)))},
//		{int32(consts.Mat), int32(rc.attr.GetInt(strconv.Itoa(consts.Mat)))},
//		{int32(consts.Med), int32(rc.attr.GetInt(strconv.Itoa(consts.Med)))},
//		{int32(consts.Ban), int32(rc.attr.GetInt(strconv.Itoa(consts.Ban)))},
//		{int32(consts.Wine), int32(rc.attr.GetInt(strconv.Itoa(consts.Wine)))},
//		{int32(consts.Book), int32(rc.attr.GetInt(strconv.Itoa(consts.Book)))},
		{int32(consts.Score), int32(rc.attr.GetInt(strconv.Itoa(consts.Score)))},
		{int32(consts.GuidePro), int32(rc.attr.GetInt(strconv.Itoa(consts.GuidePro)))},
		{int32(consts.MaxScore), int32(rc.attr.GetInt(strconv.Itoa(consts.MaxScore)))},
		{int32(consts.Mmr), int32(rc.attr.GetInt(strconv.Itoa(consts.Mmr)))},
		{int32(consts.AccTreasureCnt), int32(rc.attr.GetInt(strconv.Itoa(consts.AccTreasureCnt)))},
		{int32(consts.NotSubStarCnt), int32(rc.attr.GetInt(strconv.Itoa(consts.NotSubStarCnt)))},
		{int32(consts.Jade), int32(rc.attr.GetInt(strconv.Itoa(consts.Jade)))},
		{int32(consts.PvpTreasureCnt), int32(rc.attr.GetInt(strconv.Itoa(consts.PvpTreasureCnt)))},
		{int32(consts.PvpGoldCnt), int32(rc.attr.GetInt(strconv.Itoa(consts.PvpGoldCnt)))},
		{int32(consts.Feats), int32(rc.attr.GetInt(strconv.Itoa(consts.Feats)))},
		{int32(consts.Prestige), int32(rc.attr.GetInt(strconv.Itoa(consts.Prestige)))},
		{int32(consts.Reputation), int32(rc.attr.GetInt(strconv.Itoa(consts.Reputation)))},
		{int32(consts.WinDiff), int32(rc.attr.GetInt(strconv.Itoa(consts.WinDiff)))},
		{int32(consts.Bowlder), int32(rc.attr.GetInt(strconv.Itoa(consts.Bowlder)))},
		{int32(consts.SkyBook), int32(rc.attr.GetInt(strconv.Itoa(consts.SkyBook)))},
		{int32(consts.EventItem1), int32(rc.attr.GetInt(strconv.Itoa(consts.EventItem1)))},
		{int32(consts.CardPiece), int32(rc.attr.GetInt(strconv.Itoa(consts.CardPiece)))},
		{int32(consts.SkinPiece), int32(rc.attr.GetInt(strconv.Itoa(consts.SkinPiece)))},
		{int32(consts.CrossAreaHonor), int32(rc.attr.GetInt(strconv.Itoa(consts.CrossAreaHonor)))},
		{int32(consts.MatchScore), int32(rc.attr.GetInt(strconv.Itoa(consts.MatchScore)))},
		{int32(consts.MaxMatchScore), int32(rc.attr.GetInt(strconv.Itoa(consts.MaxMatchScore)))},
		{int32(consts.KingFlag), int32(rc.attr.GetInt(strconv.Itoa(consts.KingFlag)))},
	}
}

func (rc *ResourceComponent) HasResource(resType int, amount int) bool {
	if amount < 0 {
		return false
	}
	return rc.attr.GetInt(strconv.Itoa(resType)) >= amount
}

func (rc *ResourceComponent) _modifyResource(resType, amount int, reason string) (curAmount int) {
	oldAmount := rc.GetResource(resType)
	curAmount = oldAmount + amount
	minAmount := mod.getResourceMinAmount(resType, rc.player)
	if curAmount < minAmount {
		amount = minAmount - oldAmount
		curAmount = minAmount
	} else {
		maxAmount := mod.getResourceMaxAmount(resType)
		if curAmount > maxAmount {
			curAmount = maxAmount
		}
	}

	rc.attr.SetInt(strconv.Itoa(resType), curAmount)

	if amount != 0 {
		if name, ok := needLogRes[resType]; ok {

			if reason == "" {
				if amount > 0 {
					reason = consts.RmrUnknownOutput
				} else {
					reason = consts.RmrUnknownConsume
				}
			}
			glog.JsonInfo("resource", glog.Int("oldAmount", oldAmount), glog.Int("newAmount", curAmount),
				glog.Int("modify", amount), glog.String("name", name), glog.Int("type", resType), glog.Uint64(
				"uid", uint64(rc.player.GetUid())), glog.String("accountType", rc.player.GetLogAccountType().String()),
				glog.String("channel", rc.player.GetChannel()), glog.Int("area", rc.player.GetArea()),
				glog.String("reason", reason), glog.String("subChannel", rc.player.GetSubChannel()))

		} else {
			glog.Infof("player %d _modifyResource type=%d, oldAmount=%d, curAmount=%d, modify=%d",
				rc.player.GetUid(), resType, oldAmount, curAmount, amount)
		}
	}

	if resType == consts.Score {
		maxScore := rc.GetResource(consts.MaxScore)
		if curAmount > maxScore {
			rc.ModifyResource(consts.MaxScore, curAmount-maxScore)
		}
	}

	if resType == consts.MatchScore {
		maxMatchScore := rc.GetResource(consts.MaxMatchScore)
		if curAmount > maxMatchScore {
			rc.ModifyResource(consts.MaxMatchScore, curAmount-maxMatchScore)
		}
	}

	eventhub.Publish(consts.EvResUpdate, rc.player, resType, curAmount, amount)

	if resType == consts.Jade {
		evq.CallLater(func() {
			rc.player.Save(false)
		})
	}

	return curAmount
}

func (rc *ResourceComponent) ModifyResource(resType int, amount int, args ...interface{}) {
	var reason string
	needSync := true
	argsLen := len(args)
	if argsLen > 0 {
		if reason2, ok := args[0].(string); ok {
			reason = reason2
		}
		if argsLen > 1 {
			if needSync2, ok := args[1].(bool); ok {
				needSync = needSync2
			}
		}
	}

	curAmount := rc._modifyResource(resType, amount, reason)

	if needSync {
		msg := &pb.ResourceModify{
			Res: []*pb.Resource{{
				Type:   int32(resType),
				Amount: int32(curAmount),
			}},
		}
		agent := rc.player.GetAgent()
		if agent != nil {
			agent.PushClient(pb.MessageID_S2C_SYNC_RESOURCE, msg)
		}
	}
}

func (rc *ResourceComponent) SetResource(resType int, amount int) {
	rc.attr.SetInt(strconv.Itoa(resType), amount)
	msg := &pb.ResourceModify{
		Res: []*pb.Resource{{
			Type:   int32(resType),
			Amount: int32(amount),
		}},
	}
	eventhub.Publish(consts.EvResUpdate, rc.player, resType, amount)
	agent := rc.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_SYNC_RESOURCE, msg)
	}

	if resType == consts.Jade {
		evq.CallLater(func() {
			rc.player.Save(false)
		})
	}
}

func (rc *ResourceComponent) BatchModifyResource(modify map[int]int, args ...string) []*pb.ChangeResInfo {
	var reason string
	if len(args) > 0 {
		reason = args[0]
	}

	msg := &pb.ResourceModify{}
	var resChange []*pb.ChangeResInfo
	for _type, amount := range modify {
		if amount == 0 {
			continue
		}

		oldAmount := rc.GetResource(_type)
		curAmount := rc._modifyResource(_type, amount, reason)
		if oldAmount == curAmount {
			continue
		}

		msg.Res = append(msg.Res, &pb.Resource{
			Type:   int32(_type),
			Amount: int32(curAmount),
		})

		resChange = append(resChange, &pb.ChangeResInfo{
			Old: &pb.Resource{
				Type:   int32(_type),
				Amount: int32(oldAmount),
			},

			New: &pb.Resource{
				Type:   int32(_type),
				Amount: int32(curAmount),
			},
		})
	}

	agent := rc.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_SYNC_RESOURCE, msg)
	}
	return resChange
}

func (rc *ResourceComponent) GetResource(resType int) int {
	return rc.attr.GetInt(strconv.Itoa(resType))
}

func (rc *ResourceComponent) exchangeGold(resType int, amount int) {
	if amount == 0 {
		return
	}
	exchange := rc.gdata.GetExchangeRes(resType)
	if exchange == nil {
		return
	}

	// map[resType]modify
	resModify := make(map[int]int)
	if amount > 0 {
		needGold := exchange.Buy * amount
		if !rc.HasResource(consts.Gold, needGold) {
			return
		}
		resModify[consts.Gold] = -needGold
		resModify[resType] = amount
	} else {
		if !rc.HasResource(resType, -amount) {
			return
		}
		resModify[consts.Gold] = exchange.Sold * -amount
		resModify[resType] = amount
	}
	rc.BatchModifyResource(resModify)
}
