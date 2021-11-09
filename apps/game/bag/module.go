package bag

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"strconv"
	"time"
)

var mod *bagModule

type bagModule struct {
}

func (m *bagModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("bag")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("bag", attr)
	}
	return &bagComponent{attr: attr}
}

func (m *bagModule) isDefHeadFrame(headFrame string) bool {
	return headFrame == "" || headFrame == "1"
}

func (m *bagModule) isDefChatPop(chatPop string) bool {
	return chatPop == "" || chatPop == "1"
}

func (m *bagModule) GetAllItemIDsByType(player types.IPlayer, type_ int) []string {
	items := player.GetComponent(consts.BagCpt).(*bagComponent).getAllItemsByType(type_)
	itemIDs := make([]string, len(items))
	for i, it := range items {
		itemIDs[i] = it.GetID()
	}
	return itemIDs
}

func (m *bagModule) GetItem(player types.IPlayer, type_ int, itemID string) types.IItem {
	return player.GetComponent(consts.BagCpt).(*bagComponent).getItem(type_, itemID)
}

func (m *bagModule) HasItem(player types.IPlayer, type_ int, itemID string) bool {
	return m.GetItem(player, type_, itemID) != nil
}

func (m *bagModule) AddCardSkin(player types.IPlayer, skinID string) types.IItem {
	_, ok := gamedata.GetGameData(consts.CardSkin).(*gamedata.CardSkinGameData).ID2CardSkin[skinID]
	if !ok {
		return nil
	}

	if it, ok := player.GetComponent(consts.BagCpt).(*bagComponent).addItem(consts.ItCardSkin, skinID); ok {
		agent := player.GetAgent()
		if agent != nil {
			agent.PushClient(pb.MessageID_S2C_ADD_CARD_SKIN, &pb.AddCardSkinArg{
				Skin: skinID,
			})
		}
		return it
	} else {
		return nil
	}
}

func (m *bagModule) AddHeadFrame(player types.IPlayer, headFrame string) types.IItem {
	_, ok := gamedata.GetGameData(consts.HeadFrame).(*gamedata.HeadFrameGameData).ID2HeadFrame[headFrame]
	if !ok {
		return nil
	}

	it, _ := player.GetComponent(consts.BagCpt).(*bagComponent).addItem(consts.ItHeadFrame, headFrame)
	return it
}

func (m *bagModule) AddEquip(player types.IPlayer, equipID string) {
	_, ok := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData).ID2Equip[equipID]
	if !ok {
		return
	}

	if _, ok := player.GetComponent(consts.BagCpt).(*bagComponent).addItem(consts.ItEquip, equipID); ok {
		agent := player.GetAgent()
		if agent != nil {
			agent.PushClient(pb.MessageID_S2C_ADD_EQUIP, &pb.TargetEquip{
				EquipID: equipID,
			})
		}
	}
}

func (m *bagModule) AddEmoji(player types.IPlayer, emojiTeam int) {
	_, ok := gamedata.GetGameData(consts.Emoji).(*gamedata.EmojiGameData).Team2Emoji[emojiTeam]
	if !ok {
		return
	}
	if _, ok := player.GetComponent(consts.BagCpt).(*bagComponent).addItem(consts.ItEmoji, strconv.Itoa(emojiTeam)); ok {
		agent := player.GetAgent()
		if agent != nil {
			agent.PushClient(pb.MessageID_S2C_UNLOCK_EMOJI, &pb.UnlockEmojiArg{
				EmojiTeam: int32(emojiTeam),
			})
		}
	}
}

func (m *bagModule) AddChatPop(player types.IPlayer, chatPopID string) types.IItem {
	_, ok := gamedata.GetGameData(consts.ChatPopConfig).(*gamedata.ChatPopGameData).ID2ChatPop[chatPopID]
	if !ok {
		return nil
	}

	it, _ := player.GetComponent(consts.BagCpt).(*bagComponent).addItem(consts.ItChatPop, chatPopID)
	return it
}

func (m *bagModule) GetDefChatPop() string {
	return "1"
}

func (m *bagModule) GetDefHeadFrame() string {
	return "1"
}

func (m *bagModule) getDefHeadEmoji() string {
	return "1"
}

func (m *bagModule) GetAllItemsByType(player types.IPlayer, type_ int) []types.IItem {
	return player.GetComponent(consts.BagCpt).(*bagComponent).getAllItemsByType(type_)
}

func onFixServer1Data(args ...interface{}) {
	player := args[0].(types.IPlayer)
	equips := mod.GetAllItemsByType(player, consts.ItEquip)
	rebornSoldEquipGameData := gamedata.GetGameData(consts.RebornSoldEquip).(*gamedata.RebornSoldEquipGameData)
	var equipPrice int
	for _, soldEquipData := range rebornSoldEquipGameData.ID2Equip {
		equipPrice = soldEquipData.Price
		break
	}

	var reputation int
	var equipIDs []string
	bagCpt := player.GetComponent(consts.BagCpt).(*bagComponent)
	for _, it := range equips {
		equip := it.(*equipItem)
		if equip.GetFrom() == consts.FromReborn {
			reputation += equipPrice
			equipIDs = append(equipIDs, equip.GetID())
			bagCpt.delItem(it)
		}
	}

	reputation += module.Player.GetResource(player, consts.Reputation) * equipPrice
	if reputation > 0 {
		module.Player.SetResource(player, consts.Reputation, reputation)
	}

	glog.Infof("equip fixServer1Data uid=%d, reputation=%d, equipIDs=%v", player.GetUid(), reputation, equipIDs)
}

func onEquipDataUpdate(gdata gamedata.IGameData) {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		player.GetComponent(consts.BagCpt).(*bagComponent).checkEquipVersion()
	})
}

func OnServerStop() {
	log.save()
}

func Initialize() {
	mod = &bagModule{}
	registerRpc()
	module.Bag = mod
	eventhub.Subscribe(consts.EvFixServer1Data, onFixServer1Data)
	gamedata.GetGameData(consts.Equip).AddReloadCallback(onEquipDataUpdate)
	timer.AddTicker(5*time.Minute, log.save)
}
