package bag

import (
	"fmt"
	"kinger/gopuppy/apps/logic"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
)

func rpc_C2S_FetchHeadFrame(agent *logic.PlayerAgent, _ interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return &pb.FetchHeadFrameReply{
		HeadFrames: mod.GetAllItemIDsByType(player, consts.ItHeadFrame),
	}, nil
}

func rpc_C2S_UpdateHeadFrame(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.UpdateHeadFrameArg)
	if mod.isDefHeadFrame(arg2.HeadFrame) || mod.GetItem(player, consts.ItHeadFrame, arg2.HeadFrame) != nil {
		player.SetHeadFrame(arg2.HeadFrame)
		return nil, nil
	}
	return nil, gamedata.GameError(1)
}

func rpc_C2S_UpdateCardSkin(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.SkinCard)
	return nil, module.Card.UpdateCardSkin(player, arg2.CardID, arg2.Skin)
}

func rpc_C2S_WearEquip(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.Equip)
	card := module.Card.GetCollectCard(player, arg2.OwnerCardID)
	if card == nil {
		return nil, gamedata.GameError(1)
	}

	oldEquip := card.GetEquip()
	if oldEquip == arg2.EquipID {
		return nil, nil
	}

	it := mod.GetItem(player, consts.ItEquip, arg2.EquipID)
	if it == nil {
		return nil, gamedata.GameError(2)
	}

	var needJade int
	if oldEquip != "" {
		needJade += deEquipJade
	}

	var equipOwner types.ICollectCard = nil
	eit := it.(*equipItem)
	if eit.GetOwner() > 0 {
		equipOwner = module.Card.GetCollectCard(player, eit.GetOwner())
		if equipOwner != nil {
			needJade += deEquipJade
		}
	}

	if needJade > 0 {
		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if !resCpt.HasResource(consts.Jade, needJade) {
			return nil, gamedata.GameError(3)
		}
		resCpt.ModifyResource(consts.Jade, -needJade, consts.RmrDeEquip)

		module.Shop.LogShopBuyItem(player, "changeEquip", "换装备", 1, "gameplay",
			strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), needJade,
			fmt.Sprint("equipID=%s, cardID=%d", arg2.EquipID, arg2.OwnerCardID))
	}

	if equipOwner != nil {
		equipOwner.DeEquip()
	}
	if oldEquip != "" {
		if oldEit, ok := mod.GetItem(player, consts.ItEquip, oldEquip).(*equipItem); ok {
			oldEit.SetOwner(0)
		}
	}
	card.WearEquip(arg2.EquipID)
	eit.SetOwner(arg2.OwnerCardID)
	return nil, nil
}

func rpc_C2S_DeEquip(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.TargetEquip)
	it := mod.GetItem(player, consts.ItEquip, arg2.EquipID)
	if it == nil {
		return nil, gamedata.GameError(1)
	}

	eit := it.(*equipItem)
	ownerCardID := eit.GetOwner()
	if ownerCardID <= 0 {
		return nil, nil
	}

	card := module.Card.GetCollectCard(player, ownerCardID)
	if card == nil {
		return nil, nil
	}

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Jade, deEquipJade) {
		return nil, gamedata.GameError(3)
	}
	resCpt.ModifyResource(consts.Jade, -deEquipJade, consts.RmrDeEquip)

	module.Shop.LogShopBuyItem(player, "deEquip", "脱装备", 1, "gameplay",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), deEquipJade, "")

	card.DeEquip()
	eit.SetOwner(0)
	return nil, nil
}

func rpc_C2S_FetchEquip(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply := &pb.EquipData{}
	items := player.GetComponent(consts.BagCpt).(*bagComponent).getAllItemsByType(consts.ItEquip)
	for _, it := range items {
		eit := it.(*equipItem)
		reply.Equips = append(reply.Equips, &pb.Equip{
			EquipID:     eit.GetID(),
			OwnerCardID: eit.GetOwner(),
		})
	}
	return reply, nil
}

func rpc_C2S_FetchEmoji(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply := &pb.EmojiData{}
	items := player.GetComponent(consts.BagCpt).(*bagComponent).getAllItemsByType(consts.ItEmoji)
	for _, it := range items {
		ejt, err := strconv.Atoi(it.GetID())
		if err == nil {
			reply.EmojiTeams = append(reply.EmojiTeams, int32(ejt))
		}
	}
	return reply, nil
}

func rpc_C2S_FetchChatPop(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	chatPops := mod.GetAllItemIDsByType(player, consts.ItChatPop)
	reply := &pb.FetchChatPopReply{
		ChatPops:   chatPops,
		CurChatPop: player.GetChatPop(),
	}
	return reply, nil
}

func rpc_C2S_UpdateChatPop(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	chatPop := arg.(*pb.UpdateChatPopArg).ChatPop

	if mod.isDefChatPop(chatPop) || mod.GetItem(player, consts.ItChatPop, chatPop) != nil {
		player.SetChatPop(chatPop)
		return nil, nil
	}
	return nil, gamedata.GameError(1)
}

func rpc_C2S_FetchEquipBackInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	eqid := arg.(*pb.TargetEquip)
	reply := &pb.EquipBackInfo{}
	if equipItem, ok := player.GetComponent(consts.BagCpt).(*bagComponent).getItem(consts.ItEquip, eqid.EquipID).(*equipItem); ok {
		ownDay := equipItem.getOwnDay()
		pro := equipItem.getBackPro()
		price, resType := equipItem.getEquipPriceAndResType(player, pro)
		reply.ObtainDayNum = int32(ownDay)
		reply.LosePro = int32(pro)
		reply.Price = int32(price)
		reply.ResType = resType
	}
	return reply, nil
}

func rpc_C2S_BackBuyEquip(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	eqid := arg.(*pb.TargetEquip)
	reply := &pb.ReturnResData{}
	if equipItem, ok := player.GetComponent(consts.BagCpt).(*bagComponent).getItem(consts.ItEquip, eqid.EquipID).(*equipItem); ok {
		pro := equipItem.getBackPro()
		price, resType := equipItem.backEquip(player, pro)
		reply.ResType = resType
		reply.ResNum = int32(price)
	}
	return reply, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_HEAD_FRAME, rpc_C2S_FetchHeadFrame)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_HEAD_FRAME, rpc_C2S_UpdateHeadFrame)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_CARD_SKIN, rpc_C2S_UpdateCardSkin)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WEAR_EQUIP, rpc_C2S_WearEquip)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DE_EQUIP, rpc_C2S_DeEquip)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_EQUIP, rpc_C2S_FetchEquip)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_EMOJI, rpc_C2S_FetchEmoji)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CHAT_POP, rpc_C2S_FetchChatPop)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_CHAT_POP, rpc_C2S_UpdateChatPop)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_EQUIP_BACK_INFO, rpc_C2S_FetchEquipBackInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BACK_EQUIP, rpc_C2S_BackBuyEquip)
}
