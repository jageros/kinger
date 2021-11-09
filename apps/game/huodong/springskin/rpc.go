package springskin

import (
	"fmt"
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/proto/pb"
	"strconv"
)

func rpc_C2S_SpringSkinLottery(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSpringSkin)
	if hd == nil || !hd.IsOpen() {
		return nil, gamedata.GameError(1)
	}

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	funcPriceGameData := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if !resCpt.HasResource(consts.Jade, funcPriceGameData.LuckyBag) {
		return nil, gamedata.GameError(2)
	}

	resCpt.ModifyResource(consts.Jade, -funcPriceGameData.LuckyBag, consts.RmrLuckBag)
	reply := lotterySkin(player)

	module.Shop.LogShopBuyItem(player, "luckBag", "luckBag", 1, pb.HuodongTypeEnum_HSpringSkin.String(),
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), funcPriceGameData.LuckyBag,
		fmt.Sprintf("%v", reply.CardSkins))

	return reply, nil
}

func rpc_C2S_GetSpringSkin(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSpringSkin)
	if hd == nil || !hd.IsOpen() {
		return nil, gamedata.GameError(1)
	}

	hpd := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(pb.HuodongTypeEnum_HSpringSkin)
	if hpd == nil {
		return nil, gamedata.GameError(2)
	}

	arg2 := arg.(*pb.GetSpringSkinArg)
	if module.Bag.HasItem(player, consts.ItCardSkin, arg2.SkinID) {
		return nil, gamedata.GameError(1)
	}

	var canGet bool
	if arg2.SkinID == skinLvBu {
		canGet = isPlayerHasAllSkinByCamp(player, consts.Wei) || isPlayerHasAllSkinByCamp(player, consts.Shu) ||
			isPlayerHasAllSkinByCamp(player, consts.Wu)
	} else {
		canGet = isPlayerHasAllSkinByCamp(player, consts.Wei) && isPlayerHasAllSkinByCamp(player, consts.Shu) &&
			isPlayerHasAllSkinByCamp(player, consts.Wu) && isPlayerHasAllSkinByCamp(player, consts.Heroes)
	}

	if !canGet {
		return nil, gamedata.GameError(2)
	}

	module.Bag.AddCardSkin(player, arg2.SkinID)
	module.Shop.LogShopBuyItem(player, arg2.SkinID, arg2.SkinID, 1, pb.HuodongTypeEnum_HSpringSkin.String(),
		"", "", 0, "")
	hpd.(*springSkinHdPlayerData).onGetSkin(arg2.SkinID)
	return nil, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SPRING_SKIN_LOTTERY, rpc_C2S_SpringSkinLottery)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_SPRING_SKIN, rpc_C2S_GetSpringSkin)
}
