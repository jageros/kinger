package giftcode

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"strconv"
	"time"
)

var _ types.IPlayerComponent = &giftCodeComponent{}

type giftCodeComponent struct {
	player types.IPlayer
	attr   *attribute.MapAttr
}

func (gc *giftCodeComponent) OnLogin(isRelogin, isRestore bool) {
}

func (gc *giftCodeComponent) OnLogout() {
}

func (gc *giftCodeComponent) ComponentID() string {
	return consts.GiftCodeCpt
}

func (gc *giftCodeComponent) GetPlayer() types.IPlayer {
	return gc.player
}

func (gc *giftCodeComponent) OnInit(player types.IPlayer) {
	gc.player = player
}

func (gc *giftCodeComponent) exchange(code string) (*pb.ExchangeCodeReward, error) {
	reply := &pb.ExchangeCodeReward{}
	giftcodeAttr := attribute.NewAttrMgr("giftcode", code, true)
	err := giftcodeAttr.Load()
	if err != nil {
		return nil, gamedata.GameError(1)
	}

	codeType := giftcodeAttr.GetInt("type")
	codeGameData := gamedata.GetGameData(consts.GiftCode).(*gamedata.GiftCodeGameData)
	codeData, ok := codeGameData.Type2Code[codeType]
	if !ok {
		return nil, gamedata.GameError(1)
	}

	cnt := giftcodeAttr.GetInt("cnt")
	if codeData.Cnt > 0 && cnt >= codeData.Cnt {
		return nil, gamedata.GameError(2)
	}

	key := strconv.Itoa(codeType)
	repeat := gc.attr.GetInt(key)
	if codeData.Repeat > 0 && repeat >= codeData.Repeat {
		return nil, gamedata.GameError(3)
	}

	if codeData.BeginTimeStr != "" {
		beginTime, err := utils.StringToTime(codeData.BeginTimeStr, utils.TimeFormat2)
		if err != nil {
			return nil, gamedata.GameError(1)
		}
		if time.Now().Before(beginTime) {
			reply.ExStatus = pb.ExchangeStatus_NotEffective
			reply.ExTime = codeData.BeginTimeStr
			return reply, nil
		}
	}
	if codeData.EndTimeStr != "" {
		endTime, err := utils.StringToTime(codeData.EndTimeStr, utils.TimeFormat2)
		if err != nil {
			return nil, gamedata.GameError(1)
		}
		if time.Now().After(endTime) {
			reply.ExStatus = pb.ExchangeStatus_Expired
			reply.ExTime = codeData.EndTimeStr
			return reply, nil
		}
	}
	reply.Rewards = module.Reward.GiveRewardList(gc.player, codeData.Reward, consts.RmrGiftCodeExchange)
	reply.ExStatus = pb.ExchangeStatus_ExchangeSuccess

	giftcodeAttr.SetInt("cnt", cnt+1)
	giftcodeAttr.Save(false)
	gc.attr.SetInt(key, repeat+1)

	glog.Infof("exchange giftCode uid=%d, code=%s, type=%d", gc.player.GetUid(), code, codeType)
	return reply, nil
}
