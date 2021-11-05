package bag

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
	"time"
)

var _ types.IEquipItem = &equipItem{}

type item struct {
	type_ int
	id    string
	attr  *attribute.MapAttr
}

func newItem(type_ int, id string) types.IItem {
	attr := attribute.NewMapAttr()
	it := newItemByAttr(type_, id, attr)
	if type_ == consts.ItEquip {
		data := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData).ID2Equip[id]
		if data != nil {
			it.(*equipItem).setVersion(data.Version)
		}
	}
	return it
}

func newItemByAttr(type_ int, id string, attr *attribute.MapAttr) types.IItem {
	switch type_ {
	case consts.ItEquip:
		it := &equipItem{}
		it.type_ = type_
		it.id = id
		it.attr = attr
		return it
	default:
		return &item{
			type_: type_,
			id:    id,
			attr:  attr,
		}
	}
}

func (i *item) GetAttr() *attribute.MapAttr {
	return i.attr
}

func (i *item) GetType() int {
	return i.type_
}

func (i *item) GetAmount() int {
	return i.attr.GetInt("ammount")
}

func (i *item) GetID() string {
	return i.id
}

func (i *item) GetGmID() string {
	switch i.type_ {
	case consts.ItHeadFrame:
		return fmt.Sprintf("HF%s", i.id)
	case consts.ItEmoji:
		return fmt.Sprintf("EJ%s", i.id)
	default:
		return i.id
	}
}

func (i *item) ModifyAmount(amount int) {
	i.attr.SetInt("ammount", i.GetAmount()+amount)
}

func (i *item) GetName() string {
	switch i.type_ {
	case consts.ItHeadFrame:
		gdata := gamedata.GetGameData(consts.HeadFrame).(*gamedata.HeadFrameGameData)
		if d, ok := gdata.ID2HeadFrame[i.id]; ok {
			return fmt.Sprintf("头像框-%s", d.GetName())
		} else {
			return ""
		}

	case consts.ItCardSkin:
		gdata := gamedata.GetGameData(consts.CardSkin).(*gamedata.CardSkinGameData)
		if d, ok := gdata.ID2CardSkin[i.id]; ok {
			return fmt.Sprintf("皮肤-%s", d.GetName())
		} else {
			return ""
		}

	case consts.ItEquip:
		gdata := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData)
		if d, ok := gdata.ID2Equip[i.id]; ok {
			return fmt.Sprintf("宝物-%s", d.GetName())
		} else {
			return ""
		}

	case consts.ItEmoji:
		gdata := gamedata.GetGameData(consts.Emoji).(*gamedata.EmojiGameData)
		id, _ := strconv.Atoi(i.id)
		if d, ok := gdata.Team2Emoji[id]; ok {
			return fmt.Sprintf("表情-%s", d.GetTeamName())
		} else {
			return ""
		}

	default:
		return ""
	}
}

type equipItem struct {
	item
	owner uint32
}

func (i *equipItem) GetOwner() uint32 {
	return i.owner
}

func (i *equipItem) SetOwner(cardID uint32) {
	i.owner = cardID
}

func (i *equipItem) GetVersion() int {
	return i.attr.GetInt("version")
}

func (i *equipItem) setVersion(ver int) {
	i.attr.SetInt("version", ver)
}

func (i *equipItem) SetFrom(from int) {
	i.attr.SetInt("from", from)
}

func (i *equipItem) SetObtainTime(tim int64) {
	return
	i.attr.SetInt64("obtainTime", tim)
}

func (i *equipItem) GetObtainTime() int64 {
	return i.attr.GetInt64("obtainTime")
}

func (i *equipItem) GetFrom() int {
	return i.attr.GetInt("from")
}

func (i *equipItem) Reset(player types.IPlayer) {
	equipGameData := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData)
	isNoData := equipGameData.ID2Equip[i.GetID()] == nil

	switch i.GetFrom() {
	case consts.FromReborn:
		rebornSoldEquipGameData := gamedata.GetGameData(consts.RebornSoldEquip).(*gamedata.RebornSoldEquipGameData)
		soldEquipData, ok := rebornSoldEquipGameData.EquipID2Goods[i.GetID()]
		if ok {
			module.Player.ModifyResource(player, consts.Reputation, soldEquipData.Price, consts.RmrResetEquip)

		}

		if isNoData || ok {
			player.GetComponent(consts.BagCpt).(*bagComponent).delItem(i)
		}

	case consts.FromCampaign:
		warShopEquipGameData := gamedata.GetGameData(consts.WarShopEquip).(*gamedata.WarShopEquipGameData)
		goodsData, ok := warShopEquipGameData.ID2Goods[i.GetID()]
		if ok {
			module.Campaign.ModifyContribution(player, goodsData.FightPrice)
		}

		if isNoData || ok {
			player.GetComponent(consts.BagCpt).(*bagComponent).delItem(i)
		}

	default:
		player.GetComponent(consts.BagCpt).(*bagComponent).delItem(i)
	}

	glog.Infof("equip reset, uid=%d, equipID=%d, oldVersion=%d", player.GetUid(), i.GetID(),
		i.GetVersion())
}

func (i *equipItem) getOwnDay() int {
	obTim := i.GetObtainTime()
	if obTim <= 0 {
		obTim = time.Now().Unix()
	}
	ownDay := timer.GetDayNo(time.Now().Unix()) - timer.GetDayNo(obTim) + 1
	return ownDay
}

func (i *equipItem) getBackPro() int {
	ownDay := i.getOwnDay()
	pro := 80
	if ownDay > 0 {
		pro = ownDay * 10
	}
	if pro > 80 {
		return 80
	}
	return 100 - pro
}

func (i *equipItem) backEquip(player types.IPlayer, backPro int) (price int, resType pb.ReturnResType) {
	equipGameData := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData)
	isNoData := equipGameData.ID2Equip[i.GetID()] == nil

	price, resType = i.getEquipPriceAndResType(player, backPro)
	switch resType {
	case pb.ReturnResType_ResTypeReputation :
			module.Player.ModifyResource(player, consts.Reputation, price, consts.RmrResetEquip)
			player.GetComponent(consts.BagCpt).(*bagComponent).delItem(i)

	case pb.ReturnResType_ResTypeContributions:
			module.Campaign.ModifyContribution(player, price)
		 	player.GetComponent(consts.BagCpt).(*bagComponent).delItem(i)
	default:
		if isNoData {
			player.GetComponent(consts.BagCpt).(*bagComponent).delItem(i)
			glog.Infof("equip back, uid=%d, equipID=%d, oldVersion=%d", player.GetUid(), i.GetID(), i.GetVersion())
			return
		}
		return
	}
	glog.Infof("equip back, uid=%d, equipID=%d, oldVersion=%d", player.GetUid(), i.GetID(), i.GetVersion())
	return
}

func (i *equipItem) getEquipPriceAndResType(player types.IPlayer, backPro int) (price int, resType pb.ReturnResType) {
	switch i.GetFrom() {
	case consts.FromReborn:
		rebornSoldEquipGameData := gamedata.GetGameData(consts.RebornSoldEquip).(*gamedata.RebornSoldEquipGameData)
		soldEquipData, ok := rebornSoldEquipGameData.EquipID2Goods[i.GetID()]
		if ok {
			resType = pb.ReturnResType_ResTypeReputation
			price = soldEquipData.Price * backPro / 100
		}
	case consts.FromCampaign:
		warShopEquipGameData := gamedata.GetGameData(consts.WarShopEquip).(*gamedata.WarShopEquipGameData)
		goodsData, ok := warShopEquipGameData.ID2Goods[i.GetID()]
		if ok {
			resType = pb.ReturnResType_ResTypeContributions
			price = goodsData.FightPrice * backPro / 100
		}
	default:
		rebornSoldEquipGameData := gamedata.GetGameData(consts.RebornSoldEquip).(*gamedata.RebornSoldEquipGameData)
		soldEquipData, ok := rebornSoldEquipGameData.EquipID2Goods[i.GetID()]
		if ok {
			resType = pb.ReturnResType_ResTypeReputation
			price = soldEquipData.Price * backPro / 100
		}
		warShopEquipGameData := gamedata.GetGameData(consts.WarShopEquip).(*gamedata.WarShopEquipGameData)
		goodsData, ok := warShopEquipGameData.ID2Goods[i.GetID()]
		if ok {
			resType = pb.ReturnResType_ResTypeContributions
			price = goodsData.FightPrice * backPro / 100
		}
	}
	return
}