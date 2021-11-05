package springskin

import (
	"kinger/gamedata"
	"kinger/gopuppy/common"
	"kinger/common/consts"
	"strconv"
	"kinger/proto/pb"
	"kinger/apps/game/module/types"
	"math/rand"
	"kinger/apps/game/module"
)

const (
	springSkin2EventItems = 2
	skinLvBu = "SK06399"   // 吕布
	skinDiaoChan = "SK06699"  // 貂蝉
)

var allLuckBagSkins []*luckBagSkin
var camp2Skins map[int]common.StringSet

type luckBagSkin struct {
	data *gamedata.LuckBagReward
	amount int
}

func (ls *luckBagSkin) getSkin() string {
	return ls.data.Reward[0]
}

func (ls *luckBagSkin) getProp() int {
	return ls.data.Prop
}

func addSpringSkin(player types.IPlayer, skins []string) int {
	var eventItemAmount int
	for _, skinID := range skins {
		if module.Bag.HasItem(player, consts.ItCardSkin, skinID) {
			eventItemAmount += springSkin2EventItems
		} else {
			module.Bag.AddCardSkin(player, skinID)
		}
	}

	if eventItemAmount > 0 {
		module.Player.ModifyResource(player, consts.EventItem1, eventItemAmount)
	}
	return eventItemAmount
}

func lotterySkin(player types.IPlayer) (*pb.OpenTreasureReply) {
	var skins []string
	for i := 0; i < 5; i++ {
		ls := allLuckBagSkins[ rand.Intn(len(allLuckBagSkins)) ]
		skinID := ls.getSkin()
		for j := 0; j < ls.amount; j++ {
			skins = append(skins, skinID)
		}
	}

	reply := &pb.OpenTreasureReply{
		CardSkins: skins,
	}
	reply.Resources = append(reply.Resources, &pb.Resource{
		Type: int32(consts.EventItem1),
		Amount: int32(addSpringSkin(player, skins)),
	})

	return reply
}

func isPlayerHasAllSkinByCamp(player types.IPlayer, camp int) bool {
	sks := camp2Skins[camp]
	if sks == nil {
		return false
	}

	has := true
	sks.ForEach(func(skinID string) bool {
		 if !module.Bag.HasItem(player, consts.ItCardSkin, skinID) {
			 has = false
			 return false
		 }
		 return true
	})
	return has
}

func doInitSkins(gdata gamedata.IGameData) {
	skinConfigGameData := gamedata.GetGameData(consts.CardSkin).(*gamedata.CardSkinGameData)
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	camp2Skins2 := map[int]common.StringSet{}
	var allLuckBagSkins2 []*luckBagSkin
	data := gdata.(*gamedata.LuckBagRewardGameData)
	for _, reward := range data.Rewards {
		skinData := skinConfigGameData.ID2CardSkin[reward.Reward[0]]
		if skinData == nil {
			continue
		}

		cardData := poolGameData.GetCard(skinData.CardID, 1)
		if cardData == nil || cardData.Camp <= 0 {
			continue
		}

		ls := &luckBagSkin{data: reward}
		ls.amount, _ = strconv.Atoi(reward.Reward[1])
		allLuckBagSkins2 = append(allLuckBagSkins2, ls)

		sks := camp2Skins2[cardData.Camp]
		if sks == nil {
			sks = common.StringSet{}
			camp2Skins2[cardData.Camp] = sks
		}
		sks.Add(ls.getSkin())
	}

	allLuckBagSkins = allLuckBagSkins2
	camp2Skins = camp2Skins2
}

func initSkins() {
	luckBagGameData := gamedata.GetGameData(consts.LuckyBagReward)
	luckBagGameData.AddReloadCallback(doInitSkins)
	doInitSkins(luckBagGameData)
}
