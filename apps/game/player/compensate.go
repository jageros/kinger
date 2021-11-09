package player

import (
	"encoding/json"
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
	"math"
	"strconv"
	"time"
)

/**
用来处理玩家补偿
*/

func onCompensateStar(p *Player) {
	pvpCpt := p.GetComponent(consts.PvpCpt).(types.IPvpComponent)
	pvpLevel := pvpCpt.GetPvpLevel()
	if pvpLevel < 16 {
		return
	}
	mod.ModifyResource(p, consts.Score, 20)
}

func onCompensateRebornCard(p *Player) {
	rebornCnt := module.Reborn.GetRebornCnt(p)
	if rebornCnt <= 0 {
		return
	}

	timer.AfterFunc(4*time.Second, func() {
		glog.Infof("onCompensateRebornCard uid=%d, rebornCnt=%d", p.GetUid(), rebornCnt)
		p.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(map[uint32]*pb.CardInfo{
			6:  &pb.CardInfo{Amount: 50},
			17: &pb.CardInfo{Amount: 50},
			36: &pb.CardInfo{Amount: 50},
			38: &pb.CardInfo{Amount: 50},
			56: &pb.CardInfo{Amount: 50},
			57: &pb.CardInfo{Amount: 50},
		})
	})
}

func onCompensateVipCard(p *Player) {
	if !config.GetConfig().IsMultiLan {
		return
	}
	if p.GetPvpLevel() <= 1 {
		return
	}
	timer.AfterFunc(3*time.Second, func() {
		st := module.OutStatus.GetStatus(p, consts.OtVipCard)
		if st == nil {
			module.OutStatus.AddStatus(p, consts.OtVipCard, 24*3600)
		} else {
			st.Over(24 * 3600)
		}
	})
}

func onFixServer1Data(p *Player) {
	if config.GetConfig().IsMultiLan {
		return
	}
	if config.GetConfig().HostID >= 1000 {
		return
	}

	eventhub.Publish(consts.EvFixServer1Data, p)
}

func onCompensateXfCard(p *Player) {
	if config.GetConfig().IsMultiLan {
		return
	}
	if config.GetConfig().HostID < 1001 {
		return
	}

	modifyCards := map[uint32]*pb.CardInfo{}
	allCards := module.Card.GetAllCollectCards(p)
	for _, card := range allCards {
		modifyCards[card.GetCardID()] = &pb.CardInfo{Amount: 20}
	}

	if len(modifyCards) > 0 {
		module.Card.ModifyCollectCards(p, modifyCards)
	}

	glog.Infof("onCompensateXfCard uid=%d, modifyCards=%v", p.GetUid(), modifyCards)
}

func onCompensateXfBook(p *Player) {
	cfg := config.GetConfig()
	if cfg.IsMultiLan || cfg.IsXfServer() {
		return
	}

	rebornCnt := module.Reborn.GetRebornCnt(p)
	if rebornCnt <= 0 {
		return
	}

	var book int
	rebornCntGameData := gamedata.GetGameData(consts.RebornCnt).(*gamedata.RebornCntGameData)
	for i := 1; i <= rebornCnt; i++ {
		book += rebornCntGameData.Cnt2BookAmount[i]
	}

	allCards := module.Card.GetAllCollectCards(p)
	curBook := module.Player.GetResource(p, consts.SkyBook)
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, card := range allCards {
		unlockLevel := card.GetMaxUnlockLevel()
		if unlockLevel <= 0 {
			continue
		}

		cardData := poolGameData.GetCard(card.GetCardID(), unlockLevel-1)
		if cardData == nil {
			continue
		}

		curBook += cardData.ConsumeBook
	}

	book -= curBook
	if book > 0 {
		module.Player.ModifyResource(p, consts.SkyBook, book)
	}
	glog.Infof("onCompensateXfBook uid=%d, book=%d", p.GetUid(), book)
}

func onCompensateServer1WxInvite(p *Player) {
	cfg := config.GetConfig()
	if cfg.IsMultiLan || cfg.IsXfServer() {
		return
	}

	if p.GetArea() != 1 || p.GetChannel() != "wxgame" {
		return
	}

	eventhub.Publish(consts.EvFixServer1WxInvite, p)
}

func onCompensateLevelRechargeUnlock(p *Player) {
	if !config.GetConfig().IsXfServer() {
		return
	}

	eventhub.Publish(consts.EvFixLevelRechargeUnlock, p)
}

func onCompensatePvpScore(p *Player) {
	pvpScore := module.Player.GetResource(p, consts.Score)
	matchScore := pvpScore * 30

	if matchScore > 4000 {
		matchScore = 4000
	}

	module.Player.ModifyResource(p, consts.MatchScore, matchScore)

	glog.Infof("onCompensatePvpScore uid=%d, pvpScore=%d, matchScore=%d",
		p.GetUid(), pvpScore, matchScore)
}

func onCompensateMatchScore(p *Player) {
	pvpScore := module.Player.GetResource(p, consts.Score)
	matchScore := pvpScore * 30
	oldMatchScore := module.Player.GetResource(p, consts.MatchScore)
	module.Player.ModifyResource(p, consts.MatchScore, matchScore-oldMatchScore)
	module.Player.SetResource(p, consts.MaxMatchScore, matchScore)

	glog.Infof("onCompensateMatchScore uid=%d, pvpScore=%d, oldMatchScore=%d, matchScore=%d",
		p.GetUid(), pvpScore, oldMatchScore, matchScore)
}

func onCompensateMatchStartGiveReward(p *Player) {
	if p.GetPvpTeam() < 9 {
		return
	}
	oldRankScore := p.GetRankScore()
	oldMaxSore := p.GetMaxRankScore()
	modifyMaxScore, modifyScore := utils.CrossLeagueSeasonResetScore(oldMaxSore, oldRankScore)
	if modifyMaxScore != 0 {
		module.Player.ModifyResource(p, consts.MaxMatchScore, modifyMaxScore)
	}
	if modifyScore != 0 {
		module.Player.ModifyResource(p, consts.MatchScore, modifyScore)
	}

	sender := module.Mail.NewMailSender(p.GetUid())
	content := `亲爱的主公：
       本次更新后将正式开启王者联赛，届时各位主公将体验到全新的赛制、奖励玩法，达到一定的积分还有珍稀联赛皮肤、限量头像框哦~
       按照联赛的规则，我们将对原王者段位的星级进行统一结算，后续将以积分的形式进行赛制排名及奖励。
       感谢你对我们一如既往的支持与热爱。`

	sender.SetTitleAndContent("联赛统一结算奖励公告", content)
	rewardOdj := sender.GetRewardObj()
	rwSore := oldMaxSore - 4080
	switch {
	case rwSore >= 3000:
		rewardOdj.AddItem(pb.MailRewardType_MrtTreasure, "BX8004", 1)
	case rwSore >= 2000:
		rewardOdj.AddItem(pb.MailRewardType_MrtTreasure, "BX8003", 1)
	case rwSore >= 1000:
		rewardOdj.AddItem(pb.MailRewardType_MrtTreasure, "BX8002", 1)
	default:
		rewardOdj.AddItem(pb.MailRewardType_MrtTreasure, "BX8001", 1)
	}
	sender.Send()
}

func doCompensate(p *Player, isRelogin bool) {
	if isRelogin {
		return
	}

	cpsVer := p.getCompensateVersion()
	for cpsVer < compensateVersion {
		cpsVer++
		p.setCompensateVersion(cpsVer)
		switch cpsVer {
		case 5:
			onFixServer1Data(p)

		case 11:
			onCompensateXfCard(p)

		case 12:
			onCompensateXfBook(p)

		case 13:
			onCompensateServer1WxInvite(p)

		case 14:
			onCompensateLevelRechargeUnlock(p)

		//case 15:
		//	onCompensatePvpScore(p)

		case 16:
			onCompensateMatchScore(p)
			onCompensateMatchStartGiveReward(p)

			/* 新版本不要这些了
			case 1:
				onCompensateStar(player2)
			case 2:
				onCompensateRebornCard(player2)
			case 3:
				onCompensateVipCard(player2)
			*/
		}
	}
}

type bugPlayer1826 struct {
	Uid   common.UUid `json:"uid"`
	Gold  int         `json:"gold"`
	Cards []*struct {
		CardID uint32 `json:"cardID"`
		Amount int    `json:"amount"`
	} `json:"cards"`
}

func fixBug1826() {
	bugData := []byte(``)
	var players []*bugPlayer1826
	json.Unmarshal(bugData, &players)

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, bp := range players {
		playerAttr := attribute.NewAttrMgr("player", bp.Uid)
		err := playerAttr.Load()
		if err != nil {
			glog.Errorf("fixBug1826 uid=%d, err=%s", bp.Uid, err)
			continue
		}

		agent := logic.NewRobotAgent()
		agent.SetUid(bp.Uid)
		player := newPlayer(bp.Uid, agent, playerAttr)
		mod.addPlayer(player)
		if ok := mod.onPlayerLogin(player, false, true); !ok {
			mod.delPlayer(bp.Uid)
			glog.Errorf("fixBug1826 login err uid=%d, err=%s", bp.Uid, err)
			continue
		}

		modifyCards := map[uint32]*pb.CardInfo{}
		gold := bp.Gold
		cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
		for _, cardInfo := range bp.Cards {
			c := cardCpt.GetCollectCard(cardInfo.CardID)
			if c == nil {
				continue
			}

			level := c.GetLevel()
			curLevel := level
			cardMsg := &pb.CardInfo{}
			amount := c.GetAmount()
			curAmount := amount
			for curAmount < cardInfo.Amount && curLevel > 1 {
				curLevel--
				cdata := poolGameData.GetCard(cardInfo.CardID, curLevel)
				gold -= cdata.LevelupGold
				curAmount += cdata.LevelupNum
			}
			curAmount -= cardInfo.Amount
			if curAmount < 0 {
				curAmount = 0
			}

			cardMsg.Level = int32(curLevel - level)
			cardMsg.Amount = int32(curAmount - amount)
			if curLevel == 1 && curAmount == 0 {
				cardMsg.Level -= 1
			}

			modifyCards[cardInfo.CardID] = cardMsg
		}

		modifyCardsMsg, _ := json.Marshal(modifyCards)
		glog.Infof("fixBug1826 uid=%d, gold=%d, modifyCards=%s", bp.Uid, gold, modifyCardsMsg)

		resCpt := player.GetComponent(consts.ResourceCpt).(*ResourceComponent)
		if gold > 0 && resCpt.GetResource(consts.Gold) < gold {
			gold = resCpt.GetResource(consts.Gold)
		}

		resCpt.ModifyResource(consts.Gold, -gold, consts.RmrUnknownConsume)
		cardCpt.ModifyCollectCards(modifyCards)
		//cardCpt.OnReborn()
	}
}

func fixRecharge(uid common.UUid, channelUid, cpOrderID, channelOrderID string, paymentAmount int) {
	timer.AfterFunc(5*time.Second, func() {
		utils.PlayerMqPublish(uid, pb.RmqType_SdkRecharge, &pb.RmqSdkRecharge{
			ChannelUid:     channelUid,
			CpOrderID:      cpOrderID,
			ChannelOrderID: channelOrderID,
			PaymentAmount:  int32(paymentAmount),
		})
	})
}

func fixBug190315() {
	uids := []common.UUid{245399, 270397, 83905, 119453, 33011, 131371, 125863, 235016, 129765, 243115, 37812, 129038, 33421, 127496, 222820, 288062, 74708, 361340, 9331, 12524, 241198, 164748, 34649, 21992, 229043, 95571, 217877, 236807, 317253, 9239, 248106, 11779, 17287, 72327, 354853, 60308, 67405, 68145, 14711, 355854, 228033, 235231, 348927, 128947, 68261, 269723, 50470, 363055, 349410, 235130, 339298, 262805, 186671, 89855, 267609, 43353, 214920, 331664, 68400, 194261, 363986, 129867, 99733, 66896, 234060, 68799, 125141, 369899, 261189, 289659, 314469, 44679, 89878, 81680, 39813, 74061, 19283, 65350, 230500, 224353, 269195, 76348, 361218, 363814, 52255, 66134, 371817, 234886, 329713, 348489, 166414, 287780, 9449, 94785, 232368, 227308, 78114, 227667, 237385, 57831, 313984, 13198, 131286, 355421, 93844, 240545, 226071, 208906, 76533, 215506, 36217, 370726, 100440, 331886, 327647, 233603, 372313, 333390, 263014, 10239, 235047, 73122, 130613, 78448, 308102, 119984, 324875, 74561, 212661, 264340, 353314, 243467, 45100, 88348, 71089, 227027, 208979, 225834, 92612, 349686, 65823, 104862, 232735, 66130, 341689, 73734, 53992, 35177, 225848, 365261, 74928, 14350, 34030, 297324, 248502, 12935, 127429, 355515, 291617, 336520, 104321, 31379, 120820, 111171, 366398, 80574, 362996, 217185, 227002, 359371, 290103, 12519, 25867, 89928, 118537, 72084, 76032, 90340, 67453, 73909, 372186, 223801, 361878, 350163, 64581, 167297, 350659, 75322, 43975, 262735, 274166, 78104, 72176, 360472, 177409, 296605, 287744, 112544, 256827, 358734, 236403, 8982, 75861, 95811, 23084, 72997, 362483, 358930, 226253, 223118, 258103, 78462, 77389, 230802, 102794, 77409, 290170, 229107, 218017, 76258, 323694, 237078, 235057, 220825, 167602, 363226, 365768, 18681, 228925, 359865, 35976, 215422, 74295, 358799, 212908, 37157, 123969, 56106, 232644, 131127, 290567, 67058, 229384, 360543, 349486, 234681, 53676, 291886, 44300, 36353, 335643, 59171, 234909, 206900, 211110, 228176, 278291, 12425, 236472, 368147, 68243, 115975, 94338, 67417, 54771, 18520, 37209, 68818, 264372, 73804, 236007, 235517, 286644, 109076, 94400, 123338, 234568, 20263, 308830, 49282, 255073, 273088, 355432, 73102, 320686, 238518, 78745, 40372, 85080, 72856, 96587, 118329, 103423, 163552, 128915, 267644, 18614, 363149, 184847, 376119, 82999, 171296, 52776, 101609, 47972, 276278, 253458, 31605, 41190, 231406, 348426, 129032, 353346, 74111, 74197, 274911, 260717, 346284, 34094, 352902, 369089, 300812, 79725, 77387, 17556, 13224, 114563, 25285, 304622, 293374, 98558, 67987, 145731, 28108, 282855, 37178, 235189, 359909, 86546, 81112, 249213, 213918, 17721, 48614, 297353, 17098, 235200, 245529, 17946, 62583, 64825, 220705, 61963, 82105, 129087, 114492, 40958, 359366, 15010, 114021, 111148, 349371, 286569, 41977, 73034, 22015, 291778, 362088, 36344, 258813, 108260, 73673, 212181, 214147, 295716, 65845, 218400, 241875, 68717, 90899, 343411, 11925, 236024, 43234, 42880, 343504, 232889, 254747, 123359, 65773, 164615, 69394, 272106, 59508, 9592, 24097, 358814, 289567, 221433, 15289, 177687, 65213, 64532, 364262, 63853, 77712, 87478, 349242, 74980, 355851, 339446, 211000, 260516, 262496, 223341, 256607, 54285, 72860, 131217, 73057, 211289, 206987, 123138, 214931, 72729, 46377, 209938, 260792, 25431, 220939, 297526, 67295, 263754, 76785, 228436, 244688, 232742, 130270, 67329, 366293, 17962, 225948, 29481, 244920, 253341, 211005, 112064, 35375, 37818, 12949, 122648, 226841, 349599, 65730, 77915, 48067, 240176, 300688, 117113, 209147, 107162, 165823, 129545, 33618, 67376, 236853, 76729, 30967, 255474, 47980, 330891, 72037, 38085, 81444, 332215, 28846, 128397, 54498, 34437, 55006, 77061, 110142, 13877, 101509, 263520, 290456, 256377, 324825, 291549, 230371, 271490, 81808, 225987, 322740, 253333, 107448, 246027, 22973, 360581, 27634, 72235, 299415, 261296, 122407, 83945, 164861, 39907, 28744, 105668, 296972, 127836, 75984, 49251, 368133, 273674, 67115, 54876, 40406, 111058, 125063, 42038, 238455, 129678, 40928, 87366, 128928, 55062, 244026, 344263, 119354, 291318, 294267, 291879, 119495, 60650, 116182, 194674, 233367, 287231, 266243, 28253, 78525, 83231, 150610, 363240, 23208, 33513, 212025, 270204, 78830, 221715, 17601, 120378, 357958, 94357, 128036, 92100, 127085, 329368, 58801, 42081, 202173, 69812, 16598, 45216, 163817, 18704, 363010, 224053, 229061, 47741, 38576, 69942, 196010, 366923, 218000, 293110, 362465, 116930, 228926, 90793, 341118, 238378, 18378, 45348, 95080, 368663, 112742, 334713, 73989, 231632, 10447, 353392, 65891, 25070, 240288, 128462, 263724, 246548, 224661, 349002, 80439, 260651, 253134, 221531, 39669, 15138, 109957, 83138, 214516, 356741, 72990, 272721, 42472, 68038, 306349, 309028, 62927, 357251, 164624, 25394, 360834, 293330, 261627, 179779, 43716, 22136, 129648, 70575, 85018, 205860, 99422, 25228, 226033, 125428, 97803, 261846, 209757, 115054, 24166, 237734, 68618, 371169, 325771, 300144, 32429, 95470, 267113, 265508, 10639, 259184, 265141, 126030, 355483, 290825, 59994, 229361, 107591, 30974, 218973, 24591, 355817, 69976, 38480, 114258, 249561, 369196, 94262, 124058, 74592, 354081, 70920, 271793, 65452, 155956, 252108, 268950, 94346, 285494, 39676, 211352, 33474, 246546, 32200, 361544, 93379, 295967, 15422, 65671, 373405, 56182, 16074, 73504, 79814, 178405, 290965, 356939, 230839, 18793, 321467, 123118, 206769, 278485, 220926, 111434, 68336, 229430, 69772, 305187, 270743, 50058, 54598, 9429}

	timer.AfterFunc(3*time.Second, func() {

		for _, uid := range uids {
			sender := module.Mail.NewMailSender(uid)
			sender.SetTitleAndContent("运营补偿", "锦标赛皮肤补偿！")
			rewardObj := sender.GetRewardObj()
			rewardObj.AddItem(pb.MailRewardType_MrtCardSkin, "SK06201", 1)
			sender.Send()
			glog.Infof("fixBug190315 uid=%d", uid)
		}

	})

}

// 老服删档，记录玩家进度，用于补偿
func recordReward190430() {
	timer.AfterFunc(3*time.Second, func() {

		orderAttrs, err := attribute.LoadAll("order")
		if err != nil {
			glog.Errorf("load order err %s", err)
			return
		}

		luckBagRewards := gamedata.GetGameData(consts.LuckyBagReward).(*gamedata.LuckBagRewardGameData).Rewards
		luckBagSkins := common.StringSet{}
		for _, lreward := range luckBagRewards {
			luckBagSkins.Add(lreward.Reward[0])
		}

		uid2pay := map[common.UUid]int{}
		for _, orderAttr := range orderAttrs {
			if !orderAttr.GetBool("isComplete") {
				continue
			}

			uid := common.UUid(orderAttr.GetUInt64("uid"))
			uid2pay[uid] = uid2pay[uid] + orderAttr.GetInt("price")
		}

		workerID := int(module.Service.GetAppID() - 1)
		allAccountAttrs, err := attribute.LoadAll("account")
		if err != nil {
			glog.Errorf("load all account err %s", err)
			return
		}

		for index, attr := range allAccountAttrs {
			if index%6 != workerID {
				continue
			}

			arc := attr.GetMapAttr("1")
			if arc == nil {
				glog.Errorf("recordReward190430 no arc %v", attr.ToMap())
				return
			}

			uid := common.UUid(arc.GetUInt64("uid"))
			if uid <= 0 {
				glog.Errorf("recordReward190430 no uid %v", attr.ToMap())
				return
			}

			playerAttr := attribute.NewAttrMgr("player", uid)
			err := playerAttr.Load()
			if err != nil {
				glog.Errorf("recordReward190430 no player err=%s, %v", err, attr.ToMap())
				return
			}

			p := newPlayer(uid, logic.NewPlayerAgent(&gpb.PlayerClient{
				ClientID: uint64(uid),
				GateID:   1,
				Uid:      uint64(uid),
				Region:   1,
			}), playerAttr)
			ok := mod.onPlayerLogin(p, false, false)
			if !ok {
				glog.Errorf("recordReward190430 player login err %v", attr.ToMap())
				return
			}

			roleProAttr := attribute.NewMapAttr()
			attr.SetMapAttr("rolePro", roleProAttr)
			attr.SetInt("area", 1)

			cumulativePay := p.GetComponent(consts.ShopCpt).(types.IShopComponent).GetCumulativePay()
			if cumulativePay < uid2pay[uid] {
				cumulativePay = uid2pay[uid]
			}
			roleProAttr.SetInt("cumulativePay", cumulativePay)

			vipSt := module.OutStatus.GetStatus(p, consts.OtVipCard)
			if vipSt != nil {
				roleProAttr.SetInt("vip", vipSt.GetRemainTime())
			}

			cardAttr := attribute.NewMapAttr()
			roleProAttr.SetMapAttr("card", cardAttr)
			allCards := module.Card.GetAllCollectCards(p)
			poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
			var cardGold int
			for _, card := range allCards {
				cardData := card.GetCardGameData()
				if cardData == nil {
					continue
				}

				if cardData.IsSpCard() {
					cardAttr.SetInt("sp", cardAttr.GetInt("sp")+1)
				} else {
					rare := strconv.Itoa(cardData.Rare)
					cardID := cardData.CardID
					curLevel := card.GetLevel()
					amount := card.GetAmount()
					for lv := 1; lv < curLevel; lv++ {
						data := poolGameData.GetCard(cardID, lv)
						if data == nil {
							continue
						}
						cardGold += data.LevelupGold
						amount += data.LevelupNum
					}
					cardAttr.SetInt(rare, cardAttr.GetInt(rare)+amount)
				}
			}
			roleProAttr.SetInt("cardGold", cardGold)

			allSkins := module.Bag.GetAllItemIDsByType(p, consts.ItCardSkin)
			var skinAmount int
			for _, skinID := range allSkins {
				if !luckBagSkins.Contains(skinID) {
					skinAmount += 1
				}
			}
			roleProAttr.SetInt("skin", skinAmount)

			equips := module.Bag.GetAllItemIDsByType(p, consts.ItEquip)
			var equipAmount int
			var campaignEquipAmount int
			warShopEquipGameData := gamedata.GetGameData(consts.WarShopEquip).(*gamedata.WarShopEquipGameData)
			for _, equipID := range equips {
				if _, ok := warShopEquipGameData.ID2Goods[equipID]; ok {
					campaignEquipAmount += 1
				} else {
					equipAmount += 1
				}
			}
			roleProAttr.SetInt("equip", equipAmount)
			roleProAttr.SetInt("campaignEquip", campaignEquipAmount)

			roleProAttr.SetInt("rebornCnt", module.Reborn.GetRebornCnt(p))
			roleProAttr.SetInt("team", p.GetPvpTeam())
			roleProAttr.SetInt("skinPiece", module.Player.GetResource(p, consts.SkinPiece))
			roleProAttr.SetInt("cardPiece", module.Player.GetResource(p, consts.CardPiece))
			roleProAttr.SetInt("reputation", module.Player.GetResource(p, consts.Reputation))
			roleProAttr.SetInt("gold", module.Player.GetResource(p, consts.Gold))

			roleProAttr.SetInt("jade", module.Player.GetResource(p, consts.Jade))
			roleProAttr.SetInt("feats", module.Player.GetResource(p, consts.Feats))
			roleProAttr.SetInt("prestige", module.Player.GetResource(p, consts.Prestige))

			reply, err := p.GetAgent().CallBackend(pb.MessageID_C2S_FETCH_CONTRIBUTION, nil)
			if err == nil {
				reply2 := reply.(*pb.ContributionReply)
				roleProAttr.SetInt("contribution", int(reply2.Contribution))
			}

			buffAttr := attribute.NewMapAttr()
			roleProAttr.SetMapAttr("buff", buffAttr)
			module.OutStatus.ForEachClientStatus(p, func(st types.IOutStatus) {
				buff, ok := st.(types.IBuff)
				if ok {
					buffAttr.SetInt(strconv.Itoa(buff.GetBuffID()), buff.GetRemainTime())
				}
			})

			err = attr.Save(true)
			if err != nil {
				glog.Errorf("recordReward190430 save err %s %v", err, attr.ToMap())
				return
			}

			glog.Infof("recordReward190430 ok %v", attr.ToMap())
		}

	})
}

func giveReward190430(account *accountSt, roleProAttr *attribute.MapAttr, p *Player) {
	evq.Await(func() {
		time.Sleep(2 * time.Second)
	})

	cumulativeJade := roleProAttr.GetInt("cumulativePay") * 15

	vipRemainTime := roleProAttr.GetInt("vip")
	var vipRemainDay int
	if vipRemainTime < 0 {
		vipRemainDay = 60
	} else {
		vipRemainDay = int(math.Ceil(float64(vipRemainTime) / (24 * 60 * 60.0)))
	}
	vipJade := vipRemainDay * 20

	cardAttr := roleProAttr.GetMapAttr("card")
	var spJade int
	var cardJade2 float64
	if cardAttr != nil {
		cardAttr.ForEachKey(func(key string) {
			amount := cardAttr.GetInt(key)
			if key == "sp" {
				spJade = amount * 3000
			} else {
				switch key {
				case "1":
					cardJade2 += 0.05 * float64(amount)
				case "2":
					cardJade2 += 0.1 * float64(amount)
				case "3":
					cardJade2 += 0.15 * float64(amount)
				case "4":
					cardJade2 += 0.2 * float64(amount)
				case "5":
					cardJade2 += 0.25 * float64(amount)
				}
			}
		})
	}

	cardJade := int(math.Ceil(cardJade2))
	skinJade := roleProAttr.GetInt("skin") * 2000
	equipJade := roleProAttr.GetInt("equip")*6000 + roleProAttr.GetInt("campaignEquip")*2000

	var teamJade int
	team := roleProAttr.GetInt("team")
	if team > 2 {
		teamJade = (team - 2) * 100
	}

	skinPieceJade := roleProAttr.GetInt("skinPiece") * 40
	cardPieceJade := roleProAttr.GetInt("cardPiece") * 60
	reputationJade := int(math.Ceil(float64(roleProAttr.GetInt("reputation")) / 10))
	goldJade := int(math.Ceil((float64(roleProAttr.GetInt("gold")) + float64(roleProAttr.GetInt("cardGold"))) / 300))
	jadeJade := roleProAttr.GetInt("jade")
	var contributionJade int

	contributionVal := roleProAttr.Get("contribution")
	if contributionVal != nil {
		switch v := contributionVal.(type) {
		case int32:
			contributionJade = int(v)
		case int:
			contributionJade = int(v)
		case int64:
			contributionJade = int(v)
		case float64:
			contributionJade = int(v)
		}
	}

	var buffRemainDay int
	buffAttr := roleProAttr.GetMapAttr("buff")
	if buffAttr != nil {
		buffAttr.ForEachKey(func(key string) {
			remainTime := buffAttr.GetInt(key)
			if remainTime < 0 {
				buffRemainDay += 60
			} else {
				buffRemainDay += int(math.Ceil(float64(remainTime) / (24 * 60 * 60.0)))
			}
		})
	}
	buffJade := buffRemainDay * 5

	feats := float64(roleProAttr.GetInt("feats"))
	spJade += int(math.Ceil(feats/50000.0)) * 2000

	prestige := float64(roleProAttr.GetInt("prestige"))
	buffJade += int(math.Ceil(prestige/20000.0)) * 300
	totalJade := cumulativeJade + vipJade + spJade + cardJade + skinJade + equipJade + teamJade + skinPieceJade +
		cardPieceJade + reputationJade + goldJade + jadeJade + contributionJade + buffJade

	var headFrame string
	switch roleProAttr.GetInt("rebornCnt") {
	case 0:
		headFrame = "9"
	case 1:
		headFrame = "10"
	case 2:
		headFrame = "11"
	case 3:
		headFrame = "12"
	case 4:
		headFrame = "13"
	case 5:
		headFrame = "14"
	default:
		headFrame = "15"
	}

	glog.Infof("doReward190430 uid=%d, %v, cumulativeJade=%d, vipJade=%d, spJade=%d, cardJade=%d, skinJade=%d, "+
		"equipJade=%d, teamJade=%d, skinPieceJade=%d, cardPieceJade=%d, reputationJade=%d, goldJade=%d, "+
		"jadeJade=%d, contributionJade=%d, buffJade=%d, totalJade=%d, headFrame=%s", p.GetUid(), account.attr.ToMap(),
		cumulativeJade, vipJade, spJade, cardJade, skinJade, equipJade, teamJade, skinPieceJade, cardPieceJade,
		reputationJade, goldJade, jadeJade, contributionJade, buffJade, totalJade, headFrame)

	sender := module.Mail.NewMailSender(p.GetUid())
	contentFmt := "亲爱的主公：\n\n全新的成长体验版本已更新完成，正式开启付费不删档测试，各位主公请开启您的策略称霸之路吧~\n本次更新后，" +
		"各位主公将无法登陆原服务器和角色，我们根据大家的充值、拥有的武将、其他资源道具等进行超额补偿，您的补偿明细如下：\n" +
		"『充值补偿』：%d\n" +
		"『月卡补偿』：%d\n" +
		"『限定卡补偿』：%d\n" +
		"『皮肤补偿』：%d\n" +
		"『宝物补偿』：%d\n" +
		"『段位补偿』：%d\n" +
		"『皮肤碎片』：%d\n" +
		"『卡牌碎片』：%d\n" +
		"『名望补偿』：%d\n" +
		"『卡牌补偿』：%d\n" +
		"『金币补偿』：%d\n" +
		"『宝玉补偿』：%d\n" +
		"『勋章补偿』：%d\n" +
		"『战功补偿』：%d\n" +
		"\n如有任何疑问，请加入官方QQ群（2群号：887881215），已添加1群的主公请勿重复添加，私聊机器人还有独家礼包哦~\n感谢" +
		"各位主公对《英雄爱三国》的理解与支持。\n          《英雄爱三国》官方运营团队"
	sender.SetTitleAndContent("版本更替补偿", fmt.Sprintf(contentFmt, cumulativeJade, vipJade, spJade, skinJade,
		equipJade, teamJade, skinPieceJade, cardPieceJade, reputationJade, cardJade, goldJade, jadeJade, buffJade,
		contributionJade))
	reward := sender.GetRewardObj()
	reward.AddAmountByType(pb.MailRewardType_MrtJade, totalJade)
	reward.AddItem(pb.MailRewardType_MrtHeadFrame, headFrame, 1)
	sender.Send()
}

func doReward190430(account *accountSt, p *Player) {
	roleProAttr := account.attr.GetMapAttr("rolePro")
	if roleProAttr == nil {
		return
	}

	account.attr.Del("rolePro")
	err := account.save(true)
	if err != nil {
		glog.Errorf("doReward190430 save err uid=%d, %s %v", p.GetUid(), err, account.attr.ToMap())
		return
	}

	glog.Infof("begin doReward190430 uid=%d, %v", p.GetUid(), account.attr.ToMap())

	evq.CallLater(func() {
		giveReward190430(account, roleProAttr, p)
	})
}

func fixReward190430() {
	timer.AfterFunc(3*time.Second, func() {

		uids := []common.UUid{8965, 8966, 8967, 8968, 8969, 8970, 8971, 8972, 8973, 8974, 8975, 8976, 8977, 8978, 8979, 8980,
			8981, 8982, 8983, 8984, 8985, 8986, 8987, 8988, 8989, 8990, 8991, 8992, 8993, 8994, 8995, 8996, 8997, 8998, 8999,
			9000, 9001, 9002, 9003, 9004, 9005, 9006, 9007, 9008, 9009, 9010, 9011, 9012, 9013, 9014, 9015, 9016, 9017, 9018,
			9019, 9020, 9021, 9022, 9023, 9024, 9025, 9026, 9027, 9028, 9029, 9030, 9031, 9032, 9033, 9034, 9035, 9036, 9037,
			9038, 9039, 9040, 9041, 9042, 9043, 9044, 9045, 9046, 9047, 9048, 9049, 9050, 9051, 9052, 9053, 9054, 9055, 9056,
			9057, 9058, 9059, 9060, 9061, 9062, 9063, 9064, 9065, 9066, 9067, 9068, 9069, 9070, 9071, 9072, 9073, 9074, 9075,
			9076, 9077, 9078, 9079, 9080, 9081, 9082, 9083, 9084, 9085, 9086, 9087, 9088, 9089, 9090, 9091, 9092, 9093, 9094,
			9095, 9096, 9097, 9098, 9099, 9100, 9101, 9102, 9103, 9104, 9105, 9106, 9107, 9108, 9109, 9110, 9111, 9112, 9113,
			9114, 9115, 9116, 9117, 9118, 9119, 9120, 9121, 9122, 9123, 9124, 9125, 9126, 9127, 9128, 9129, 9130, 9131, 9132,
			9133, 9134, 9135, 9136, 9137, 9138}
		for _, uid := range uids {
			sender := module.Mail.NewMailSender(uid)
			sender.SetTitleAndContent("头像框补偿", "现为您补发遗漏的老版本纪念头像框，感谢您一如既往的支持。")
			reward := sender.GetRewardObj()
			reward.AddItem(pb.MailRewardType_MrtHeadFrame, "9", 1)
			sender.Send()
		}

	})
}

func giveRewardFengce(fcpay int, p *Player) {
	evq.Await(func() {
		time.Sleep(2 * time.Second)
	})

	sender := module.Mail.NewMailSender(p.GetUid())
	content1 := "亲爱的主公：\n\n" +
		"仙峰游戏旗下九宫格TCG手游《英雄爱三国》全平台公测5月21日全面开战！\n\n" +
		"非常感谢您在封测期间对我们的热爱与支持，我们现根据您在封测期间的充值情况进行宝玉返还。为表达官方真挚的谢意，我们将返还比例由原先预定的150%提高至200%！\n\n" +
		"返还规则：封测期间的充值金额*200%价值宝玉\n\n" +
		"返还宝玉："
	content2 := "\n\n" +
		"后续我们将会继续不停地优化版本，完善游戏品质，不辜负大家对我们的热爱与期待～\n\n\n" +
		"《英雄爱三国》官方运营团队"
	jade := fcpay * 20
	sender.SetTitleAndContent("封测充值返还", content1+strconv.Itoa(jade)+content2)
	reward := sender.GetRewardObj()
	reward.AddAmountByType(pb.MailRewardType_MrtJade, jade)
	sender.Send()
}

func doRewardFengce(account *accountSt, p *Player) {
	fcpayVal := account.attr.Get("fcpay")
	if fcpayVal == nil {
		return
	}

	var fcpay int
	switch v := fcpayVal.(type) {
	case int32:
		fcpay = int(v)
	case int:
		fcpay = int(v)
	case int64:
		fcpay = int(v)
	case float64:
		fcpay = int(v)
	}

	if fcpay <= 0 {
		return
	}

	account.attr.Del("fcpay")
	err := account.save(true)
	if err != nil {
		glog.Errorf("doRewardFengce save err uid=%d, %s %v", p.GetUid(), err, account.attr.ToMap())
		return
	}

	glog.Infof("begin doRewardFengce uid=%d, fcpay=%d", p.GetUid(), fcpay)

	evq.CallLater(func() {
		giveRewardFengce(fcpay, p)
	})
}

func rewardRecruit() {
	timer.AfterFunc(3*time.Second, func() {
		uid2jade := map[common.UUid]int{27824: 99, 16398: 297, 83580: 99, 16419: 98, 98355: 49, 61105: 2178, 91657: 297, 16464: 99, 16467: 2128, 27322: 49, 84669: 99, 37102: 197, 19135: 2940, 16510: 2673, 20190: 1029, 82054: 49, 75838: 8167, 32909: 49, 16543: 99, 16545: 99, 16548: 148, 16550: 49, 25647: 247, 20737: 3236, 16559: 346, 98488: 2989, 16584: 148, 101069: 49, 98520: 49, 16605: 148, 98538: 49, 19154: 7344, 19155: 1029, 82168: 49, 16637: 693, 19157: 2610, 49408: 14107, 84571: 148, 16643: 99, 92209: 495, 28378: 99, 86208: 1980, 16655: 148, 91664: 147, 16669: 99, 82210: 99, 19165: 49, 82225: 27918, 65844: 2326, 98622: 98, 98623: 49, 18560: 2989, 33099: 11286, 16719: 247, 82263: 6039, 82266: 1485, 16737: 148, 16741: 2475, 98672: 49, 82297: 297, 16762: 198, 16771: 99, 98696: 49, 98697: 98, 82314: 3267, 103671: 49, 21915: 1960, 16804: 246, 24646: 5078, 98727: 49, 33196: 396, 98734: 49, 33199: 198, 82360: 3267, 101109: 1029, 82373: 99, 19191: 3038, 21922: 16758, 98766: 2009, 16848: 792, 82387: 2178, 16852: 6138, 82392: 98, 16861: 99, 82401: 198, 26203: 980, 51963: 2277, 21926: 3087, 27388: 6237, 33263: 49, 49655: 99, 16889: 148, 16890: 99, 19200: 5929, 82437: 99, 16904: 19800, 16916: 99, 82454: 99, 82455: 99, 19204: 2940, 82469: 99, 21938: 2989, 98867: 98, 78580: 49, 98870: 245, 49719: 198, 82489: 99, 98882: 49, 16968: 5742, 49739: 21978, 25661: 99, 98894: 49, 98906: 7889, 19215: 4059, 16989: 99, 33376: 2574, 82532: 297, 19217: 10593, 37975: 49, 82555: 99, 17034: 22254, 66188: 11979, 82576: 99, 82577: 297, 17043: 396, 17046: 12967, 82586: 198, 82587: 297, 33439: 594, 82593: 198, 33456: 99, 17085: 99, 17087: 198, 82624: 10880, 19232: 938, 99011: 49, 25665: 343, 24695: 297, 17107: 444, 17110: 99, 91679: 99, 17121: 49, 82659: 495, 24700: 539, 99052: 1029, 49904: 21285, 17146: 14988, 17147: 2673, 17148: 98, 17154: 49, 17162: 247, 17164: 1089, 99085: 147, 82713: 1980, 82715: 197, 17186: 396, 24710: 23859, 17196: 49, 76595: 297, 99128: 980, 82752: 99, 17217: 148, 17223: 3117, 99147: 98, 82780: 99, 17250: 99, 19259: 3465, 19260: 2207, 82801: 148, 82808: 99, 82820: 2079, 26217: 5533, 93322: 99, 99226: 147, 11077: 198, 99233: 1029, 99234: 49, 19270: 19181, 19271: 49189, 66479: 297, 60232: 297, 99250: 1666, 66484: 495, 17344: 12770, 50114: 99, 99267: 1029, 99268: 49, 17350: 99, 50119: 99, 50121: 297, 17354: 20225, 99281: 49, 66519: 2178, 82905: 495, 17378: 99, 82917: 9999, 19281: 49, 50163: 198, 82932: 4059, 82938: 99, 22402: 1325, 50182: 99, 82953: 2079, 17419: 99, 82961: 198, 22403: 49, 99366: 4214, 33833: 297, 82992: 49, 83001: 1980, 33851: 6039, 99397: 98, 83016: 99, 68232: 2276, 83020: 17919, 83023: 3465, 83024: 2079, 83027: 2029, 101220: 49, 99429: 98, 93329: 8156, 19304: 3602, 83062: 99, 101225: 49, 75295: 2079, 17541: 9999, 91743: 297, 33938: 791, 93876: 2128, 83096: 198, 17562: 10571, 17575: 12177, 17578: 99, 19314: 6483, 50351: 99, 17586: 26837, 17591: 49, 17592: 4256, 17593: 7869, 17594: 21729, 17597: 197, 17600: 99, 17602: 17215, 17604: 296, 17608: 99, 83156: 99, 17634: 23364, 17636: 28995, 34024: 6039, 17644: 21977, 83184: 49, 87593: 99, 17657: 97452, 17665: 10354, 83205: 99, 17672: 147, 83210: 1980, 17675: 544, 95791: 99, 83230: 99, 83241: 198, 22412: 9205, 87603: 99, 99637: 2989, 84873: 49, 24800: 6088, 66882: 99, 19339: 7197, 26232: 297, 83276: 2079, 19991: 3663, 83280: 99, 17745: 99, 17750: 246, 76970: 99, 17755: 49, 19686: 4455, 99685: 1078, 83303: 8118, 83314: 2277, 66934: 5049, 17783: 294, 58166: 99, 83325: 99, 49389: 198, 99730: 49, 83347: 2079, 19354: 49, 83360: 792, 17832: 99, 19356: 5978, 99763: 49, 34237: 2079, 34238: 643, 83397: 2960, 83400: 1881, 83407: 198, 83412: 2079, 83423: 99, 83445: 99, 43945: 99, 17919: 693, 87638: 2079, 83472: 7920, 93889: 148, 87642: 99, 83487: 99, 85161: 49, 17981: 99, 50758: 98, 50761: 99, 83538: 10197, 83539: 297, 24846: 49, 50783: 99, 19387: 48163, 99943: 147, 18025: 99, 92538: 2079, 99952: 49, 22120: 5532, 67188: 99, 83574: 4009, 22121: 5786, 83577: 99, 67196: 1980, 79466: 148, 22123: 99, 99974: 1029, 22124: 49, 83594: 3960, 35778: 98, 83605: 99, 83606: 198, 18075: 98, 18084: 99, 83622: 7216, 18092: 18018, 18094: 98, 100015: 1960, 18096: 49, 93128: 2695, 18098: 49, 67707: 3267, 34485: 396, 18102: 6335, 100031: 49, 83654: 148, 18124: 3404, 100049: 49, 83672: 99, 83680: 247, 83682: 99, 18147: 2079, 100071: 12789, 67307: 2178, 18166: 6138, 18170: 1029, 83707: 99, 83710: 99, 19413: 346, 19414: 13008, 18183: 838, 18193: 296, 83732: 2178, 95877: 148, 18610: 3563, 83753: 99, 18223: 99, 83765: 2009, 83773: 198, 83775: 4059, 95883: 99, 83782: 2178, 85170: 99, 100170: 49, 18255: 1039, 100180: 1029, 83797: 6039, 100183: 49, 19428: 16334, 18271: 246, 51044: 18512, 83813: 99, 83819: 3969, 100205: 147, 18290: 49, 83828: 2079, 82239: 6138, 83840: 99, 83841: 2871, 83843: 49, 100239: 49, 83856: 2623, 83876: 15076, 83878: 99, 100286: 1029, 51349: 99, 18372: 99, 83911: 692, 83914: 3960, 84983: 396, 83920: 495, 51158: 99, 100314: 3969, 100318: 245, 83946: 99, 18425: 98, 18428: 296, 18431: 49, 18442: 23070, 18446: 2079, 83988: 4059, 18453: 15362, 18454: 99, 58990: 15741, 83999: 6484, 18471: 148, 18473: 3531, 30347: 49, 18486: 198, 18497: 148, 82273: 99, 18509: 2178, 18510: 147, 18511: 79820, 18512: 85917, 100434: 49, 18516: 62684, 18518: 40796, 18519: 26728, 18521: 80818, 18522: 148, 18523: 10878, 18525: 9088, 18526: 7751, 18528: 7900, 18529: 246, 18530: 8869, 18531: 1128, 18532: 99, 18534: 35695, 18535: 70903, 18536: 11858, 18537: 14254, 18538: 840, 18539: 148, 67692: 99, 18541: 4998, 18542: 2721, 18543: 1176, 18544: 6039, 18545: 6909, 100466: 49, 18547: 642, 18548: 49, 18549: 2474, 84086: 495, 18551: 85339, 18552: 3107, 18553: 34401, 18554: 40142, 18555: 98, 18556: 5047, 84093: 2079, 18558: 98, 18559: 4059, 84096: 2277, 18561: 5424, 18562: 9979, 18563: 98, 18564: 98, 18565: 1960, 84102: 6583, 18567: 24014, 18568: 8284, 18569: 13106, 18570: 395, 18571: 2355, 84108: 6930, 18573: 5484, 18574: 6424, 18575: 2820, 18576: 4059, 18577: 6909, 18578: 30786, 18579: 5048, 18580: 6146, 18581: 6385, 18582: 7869, 18583: 4067, 18584: 4780, 18585: 49, 18586: 11818, 100507: 49, 18588: 43758, 18589: 7463, 18590: 76538, 18591: 4108, 18592: 346, 18593: 4949, 18594: 7751, 18595: 9541, 18596: 11205, 18597: 39699, 18598: 15937, 18599: 15889, 18600: 2108, 18601: 147326, 18602: 3503, 18603: 18678, 18604: 14946, 18605: 22966, 18606: 7821, 18607: 396, 18608: 148, 18609: 5880, 100530: 49, 18612: 4384, 18614: 98, 18615: 691, 18616: 6196, 18617: 17580, 18619: 8699, 18620: 4157, 18621: 13601, 18622: 49, 18623: 10444, 18624: 8019, 18625: 5245, 18626: 6909, 18627: 1325, 18628: 11798, 18629: 14284, 18630: 6126, 18631: 20441, 18633: 13037, 18634: 3087, 18635: 7126, 18636: 246, 18637: 16335, 100558: 14749, 18639: 9849, 18641: 6780, 18642: 49, 18643: 6039, 18644: 1820, 18645: 593, 84182: 45837, 18650: 26401, 18651: 37796, 18652: 938, 18653: 4455, 18654: 938, 18655: 17096, 18656: 67117, 84193: 99, 18658: 5989, 18660: 691, 18661: 66278, 18662: 6088, 18663: 16186, 100584: 49, 18665: 49, 18666: 4107, 18667: 52903, 18668: 31680, 18669: 79244, 18670: 7840, 18671: 75339, 18672: 7146, 51442: 99, 18675: 16709, 18676: 99, 18677: 246, 18678: 42565, 18679: 2800, 18680: 44359, 18681: 32223, 84218: 19206, 18683: 8929, 18684: 5690, 18685: 89198, 18686: 5641, 18687: 17660, 18688: 13452, 18690: 1029, 18692: 1029, 18693: 5878, 18694: 10098, 18695: 1127, 18696: 3920, 18697: 25478, 18699: 6166, 18700: 36003, 18701: 16292, 18702: 28518, 18703: 20978, 84240: 495, 18705: 295, 18706: 9017, 18707: 9838, 18708: 9423, 18709: 147, 18710: 147, 100631: 147, 18712: 887, 18713: 2918, 18714: 22737, 18715: 1029, 18716: 1980, 18717: 12403, 18718: 74158, 18719: 12778, 18720: 3236, 18721: 643, 18722: 33340, 18723: 37618, 18725: 35558, 18726: 11756, 18727: 20848, 18728: 46604, 18731: 4167, 18732: 31203, 18733: 18312, 91733: 980, 18735: 49, 18736: 3137, 18737: 1820, 18738: 1029, 18739: 5443, 18741: 6909, 18742: 1127, 18743: 15196, 18744: 7107, 18745: 7265, 18747: 14235, 18749: 7987, 18750: 8019, 18751: 4849, 18752: 6909, 18753: 12077, 18754: 78650, 18755: 10829, 18756: 2502, 18758: 345, 84295: 445, 18760: 4800, 18761: 1524, 18762: 15840, 18763: 2940, 18765: 8660, 18766: 7374, 18767: 9036, 18768: 1980, 18769: 14778, 18770: 147, 18771: 43399, 18772: 1127, 18773: 8068, 18774: 85315, 18775: 3960, 67928: 2079, 18777: 2721, 18778: 6483, 18779: 21243, 18780: 4059, 100701: 245, 18783: 12373, 18784: 20501, 18785: 47238, 18786: 80742, 18787: 74480, 18788: 8611, 18789: 2009, 18790: 55535, 59001: 99, 18793: 792, 18794: 30442, 18795: 49, 18797: 35735, 18798: 2602, 18799: 4207, 18800: 49525, 18802: 34786, 18803: 47717, 18804: 10395, 100726: 49, 18807: 2940, 18808: 13472, 18810: 39655, 18811: 38073, 18812: 2574, 84349: 99, 18814: 2602, 51583: 98, 18816: 66852, 18817: 19054, 18818: 49, 18819: 539, 84356: 2079, 18821: 49, 18823: 29947, 18824: 2009, 18825: 3305, 18826: 49, 18827: 12889, 19522: 6760, 18831: 4503, 18832: 1226, 18833: 13026, 18834: 9324, 18835: 26805, 18836: 6236, 18837: 2553, 84374: 1980, 18839: 9136, 18840: 246, 18841: 85618, 18842: 5593, 18843: 2573, 18844: 10593, 18845: 49, 18846: 4513, 84383: 147, 18848: 1225, 18850: 7987, 18851: 3998, 18852: 49, 18853: 938, 18854: 7454, 18855: 791, 18856: 6039, 18857: 11037, 18859: 5929, 18860: 10048, 71410: 3168, 18862: 6958, 84399: 99, 18864: 4137, 18865: 5166, 18866: 13958, 18867: 10156, 100788: 49, 18870: 2989, 18872: 2524, 18873: 10838, 18874: 6138, 18875: 2326, 18876: 1276, 18881: 1869, 18882: 16718, 19531: 3137, 18884: 889, 18885: 12573, 18886: 741, 18888: 3038, 18889: 147, 18890: 2128, 18892: 12324, 24994: 3385, 18894: 6225, 18895: 24737, 18896: 5294, 84433: 99, 84434: 4059, 18899: 16680, 18900: 345, 18901: 938, 18902: 2277, 18903: 13769, 18904: 4018, 18905: 12720, 18906: 2960, 18907: 246, 18908: 41002, 18909: 38920, 18911: 98, 18912: 49, 18914: 40904, 18915: 20344, 18916: 4801, 18917: 2325, 18918: 10839, 18919: 2989, 18921: 3969, 18922: 6682, 18923: 98, 18924: 49, 18925: 3452, 18926: 18709, 18927: 6225, 18928: 71818, 18929: 99, 18930: 94829, 18931: 5989, 18932: 3969, 18933: 7464, 18935: 198, 100856: 2009, 18937: 2107, 18938: 8126, 18939: 246, 18940: 5047, 100862: 49, 18943: 56130, 18944: 2107, 18945: 31470, 18946: 588, 18947: 61238, 18948: 3305, 18949: 147, 18950: 4454, 18951: 13404, 18952: 8096, 18953: 6740, 18954: 5147, 18955: 11155, 18956: 1127, 18957: 88643, 18958: 2128, 18959: 4068, 18960: 44555, 18962: 197, 18963: 1029, 18965: 52271, 18966: 23924, 18967: 24600, 18969: 938, 18970: 23300, 18971: 14304, 68124: 148, 18973: 2940, 18974: 11314, 18977: 5047, 18978: 7352, 18979: 6533, 18980: 1671, 18981: 1029, 100902: 49, 18983: 45192, 18984: 10296, 18985: 7315, 18986: 5929, 18987: 98, 18989: 76326, 18990: 197, 18991: 6780, 18993: 90020, 18994: 5998, 18995: 2177, 18996: 27264, 18997: 6929, 18999: 46677, 19000: 21978, 84537: 2375, 19004: 49, 19005: 938, 19007: 3087, 19008: 41134, 19010: 1325, 19014: 98, 35401: 49, 19018: 49, 19019: 7492, 19020: 10593, 84558: 99, 19023: 2206, 19024: 32487, 19025: 98, 19026: 6780, 19027: 85395, 19028: 3849, 19030: 17820, 19031: 980, 19033: 2009, 19034: 1980, 19035: 1830, 19037: 49, 19038: 1826, 19039: 840, 19040: 99, 19041: 10691, 19042: 9403, 19044: 123698, 19045: 5978, 19046: 2107, 19047: 8264, 19048: 6920, 19049: 1960, 19050: 19402, 19051: 11126, 100972: 49, 19053: 9048, 19054: 53465, 100975: 49, 19056: 36925, 19057: 4949, 19058: 4503, 19059: 4059, 19061: 10927, 19063: 2058, 19064: 71722, 19065: 14145, 19066: 2425, 19067: 6039, 19068: 1226, 19070: 2646, 19072: 49, 19073: 6958, 19074: 18562, 19075: 33759, 19076: 22324, 19077: 1980, 19078: 49, 19079: 2424, 19080: 33826, 19081: 1029, 68234: 98, 19083: 791, 19085: 94162, 19088: 198, 19089: 4108, 19090: 8869, 19093: 31878, 19094: 49, 19095: 147, 19096: 3107, 19098: 493, 19185: 6958, 84636: 99, 19101: 2869, 19102: 5195, 19103: 1137, 19104: 9999, 19105: 12078, 19106: 49, 19107: 9405, 19108: 49, 19109: 3582, 19110: 1029, 19111: 49, 19112: 10185, 19113: 5682, 19114: 98, 19115: 1127, 19116: 2058, 19117: 35498, 19118: 5929, 19119: 49, 19120: 6186, 19122: 1078, 84659: 99, 19124: 9900, 101046: 49, 19128: 77821, 19129: 16946, 19130: 4117, 19131: 11769, 19132: 34489, 19133: 20689, 19134: 1960, 101055: 2107, 19136: 12749, 19137: 3850, 19138: 24718, 19139: 6809, 19140: 980, 19142: 7938, 19143: 27223, 19145: 2325, 19147: 37478, 19148: 21928, 19149: 6039, 19150: 6682, 19151: 9900, 19152: 4137, 19153: 10858, 68306: 4059, 68307: 147, 84693: 99, 19158: 10116, 19159: 12495, 19160: 147, 19162: 13986, 19163: 31401, 85114: 495, 19168: 16879, 19169: 10838, 19170: 7007, 19171: 1036, 19172: 16243, 19173: 69120, 19174: 247, 19176: 32766, 19177: 938, 19178: 5582, 19179: 2226, 19180: 98, 19181: 3038, 19182: 2989, 19183: 33686, 84721: 37620, 19186: 67978, 19189: 49, 19190: 2524, 85999: 99, 19192: 8166, 19194: 98, 19195: 19986, 19196: 7889, 19197: 20090, 19198: 396, 84736: 2079, 19201: 1029, 19202: 643, 19203: 22302, 101124: 2891, 19205: 9146, 19206: 6245, 19207: 3602, 19208: 7889, 19209: 8641, 19210: 14798, 19211: 16066, 19213: 938, 19214: 13769, 51983: 99, 66008: 99, 19218: 16968, 19219: 20916, 19222: 13530, 19223: 17441, 19224: 148, 84762: 49, 19227: 18216, 19228: 6909, 19229: 7497, 19230: 3969, 19231: 3058, 84768: 297, 19233: 1078, 19234: 3107, 19235: 1029, 84772: 297, 19237: 98, 19238: 1960, 84775: 2277, 19240: 44498, 19241: 6226, 19242: 14502, 19243: 441, 19244: 3969, 19245: 687, 19246: 33363, 19247: 98, 19248: 2058, 19251: 4557, 19252: 38757, 19254: 5858, 19255: 42238, 19256: 98, 19257: 6336, 101179: 49, 84796: 4554, 19261: 11880, 19262: 642, 19263: 24106, 19264: 24698, 19265: 1722, 19266: 938, 19267: 1029, 19269: 5929, 101190: 49, 84807: 99, 19272: 1574, 19273: 16046, 19274: 19404, 19275: 2079, 19277: 6019, 19278: 49, 101201: 1029, 19282: 11368, 19283: 188071, 19285: 14988, 19286: 22786, 19288: 742, 19289: 6039, 19290: 19649, 19291: 1029, 19292: 39004, 19294: 296, 19295: 103454, 19296: 49, 19297: 839, 19298: 16758, 19299: 980, 19300: 9740, 19301: 47101, 19302: 8967, 19303: 2401, 35688: 198, 19305: 5978, 19307: 12057, 19308: 2989, 19309: 10838, 19311: 5880, 19312: 43626, 101234: 14749, 19315: 14798, 84852: 445, 19318: 49, 19320: 57666, 19321: 9849, 19322: 12276, 19323: 6077, 19324: 147, 19325: 1128, 19326: 980, 19327: 11086, 19328: 18936, 19329: 247, 19330: 10856, 19331: 980, 19332: 6503, 19333: 99, 19334: 19800, 19335: 12838, 19336: 2206, 19337: 13235, 19338: 3502, 84875: 99, 19341: 1127, 19342: 392, 19345: 6959, 19346: 4502, 19347: 2622, 19348: 13818, 19349: 5929, 19350: 147, 19352: 31341, 19353: 8967, 35738: 49, 19355: 15382, 84892: 99, 19357: 4454, 19358: 8820, 19359: 295, 19361: 49, 19363: 5900, 19364: 1078, 19365: 13423, 19367: 2811, 19368: 52988, 84905: 99, 19370: 4841, 19371: 1029, 19373: 297, 19374: 109430, 19375: 51396, 19376: 7562, 19377: 13095, 19378: 8710, 19380: 1177, 19381: 147, 19382: 4681, 19383: 148, 19384: 49, 19385: 21649, 19386: 147, 101307: 49, 19389: 49, 19390: 4949, 19391: 98, 19392: 4503, 19394: 22580, 19395: 3157, 19396: 4858, 19397: 8918, 19398: 19878, 19399: 10829, 19400: 2128, 19402: 4206, 19404: 9651, 19407: 346, 19408: 56306, 19410: 839, 19411: 260748, 19412: 12769, 84949: 10395, 84950: 396, 19415: 938, 84952: 148, 84953: 6138, 19418: 691, 19419: 980, 84956: 49, 19422: 980, 19423: 73065, 84960: 7365, 19425: 8869, 19426: 4557, 19427: 7056, 84964: 99, 19429: 3482, 19430: 11979, 19431: 1276, 19432: 1029, 19434: 1277, 19435: 10798, 19436: 3108, 19437: 14749, 19439: 1136, 19440: 345, 19442: 4900, 19443: 13769, 19444: 49, 19446: 1325, 19447: 91123, 19449: 11809, 19450: 19431, 19451: 20303, 19452: 494, 19453: 8019, 19454: 14749, 19455: 3038, 19457: 1029, 19458: 6039, 19459: 196, 19460: 198, 84997: 99, 19462: 148, 19463: 18165, 19464: 980, 19465: 1928, 68618: 1188, 19467: 346, 19468: 13008, 19469: 1918, 19470: 1127, 85007: 4158, 19472: 3493, 19473: 1078, 19474: 11979, 27381: 2376, 19476: 99, 101397: 1029, 101398: 49, 19479: 17265, 19481: 20493, 19482: 1325, 19483: 542, 101404: 49, 19485: 3018, 19486: 16137, 19487: 17660, 19488: 147, 19489: 49, 19490: 14255, 19491: 15740, 35876: 13937, 19493: 2999, 19494: 4454, 19495: 22076, 19496: 148, 85033: 297, 19498: 1227, 19499: 641, 19500: 49, 19501: 39876, 19502: 4900, 19504: 543, 19505: 4166, 19506: 7364, 19507: 75879, 19508: 297, 19509: 28620, 19510: 6117, 19512: 1029, 19513: 98, 85050: 148, 19515: 49, 19516: 2058, 19517: 2009, 19518: 17919, 19521: 1474, 101442: 147, 19523: 3969, 19524: 2918, 19525: 1960, 35911: 6385, 85064: 495, 19529: 22680, 19530: 15443, 101451: 49, 19532: 13769, 35917: 99, 19534: 444, 19536: 25529, 85176: 99, 19539: 4018, 19540: 7146, 19541: 980, 19542: 6682, 19543: 197, 19544: 1275, 19546: 4631, 19547: 2079, 19548: 27293, 19549: 49, 19550: 6077, 19551: 8997, 85088: 99, 19553: 5979, 19555: 49, 19556: 4780, 19557: 2108, 19559: 196, 19560: 98, 19561: 5117, 19562: 1980, 68715: 297, 19564: 4166, 19565: 2009, 19566: 938, 19567: 18619, 19568: 3108, 19569: 5929, 19571: 49, 19572: 28906, 68725: 99, 14185: 49, 101496: 49, 19577: 24688, 19578: 5880, 19580: 296, 19581: 4067, 19584: 4206, 19585: 11809, 85122: 4059, 19587: 5038, 19589: 4900, 19590: 12740, 19591: 11809, 19592: 2009, 19595: 197, 19596: 49, 19598: 3920, 19599: 15939, 19600: 21717, 19601: 345, 19602: 9254, 19603: 889, 85142: 7920, 37015: 99, 19609: 1029, 19610: 5929, 19611: 22473, 19612: 791, 19613: 980, 19614: 3038, 19615: 32114, 19616: 12097, 19617: 1029, 19618: 10880, 20839: 1376, 19620: 1324, 68773: 99, 96113: 3038, 19624: 3512, 19625: 49, 19628: 246, 19629: 33907, 19632: 4018, 19633: 13809, 19634: 16115, 19635: 21609, 19636: 980, 19637: 7920, 85174: 4504, 19640: 7920, 19641: 10938, 19642: 14454, 19644: 98, 85181: 20196, 19646: 98, 19647: 14798, 19648: 1980, 19649: 24006, 19650: 1980, 19651: 6039, 19652: 8958, 19653: 7344, 19654: 3038, 19656: 3960, 19657: 1334, 19659: 297, 19660: 5940, 19661: 16086, 19662: 98, 19664: 3960, 19665: 395, 19666: 20511, 19667: 343, 19669: 12888, 19670: 5335, 19671: 4949, 19672: 3207, 85209: 148, 19674: 4949, 85212: 396, 19678: 495, 19679: 7493, 101600: 49, 85217: 49, 19682: 345, 19683: 8869, 19684: 5364, 101606: 49, 19687: 1127, 19688: 2178, 19690: 2989, 19693: 7205, 19694: 639, 19695: 197, 19697: 9622, 19698: 296, 19704: 30865, 19705: 24698, 19708: 10838, 19709: 147, 19712: 4401, 19713: 45299, 19714: 12601, 19716: 10829, 19717: 49, 19718: 3969, 19719: 2157, 19722: 2009, 19724: 8019, 19725: 45369, 19726: 4949, 19727: 493, 19728: 4405, 19729: 17689, 19730: 71379, 19731: 170118, 19732: 7920, 19733: 9898, 19734: 19343, 19735: 98, 19736: 4059, 19737: 642, 19739: 5958, 19740: 11929, 19741: 888, 19742: 594, 19743: 6969, 19745: 1325, 19749: 2277, 19750: 7889, 19751: 4165, 19752: 1226, 52521: 99, 19754: 1029, 19755: 2058, 19756: 1423, 19757: 246, 19758: 3960, 85296: 49, 19761: 1523, 19763: 2989, 19764: 840, 19765: 1226, 19766: 3355, 19769: 17689, 19771: 49, 19772: 246, 19773: 12472, 19774: 18026, 19775: 98, 19680: 5146, 19778: 2671, 19780: 12818, 19782: 11433, 19681: 23064, 19784: 691, 38796: 99, 19790: 19105, 19791: 2206, 85328: 890, 19793: 4949, 19794: 3009, 19795: 3404, 19797: 742, 52566: 99, 19799: 49, 36184: 10197, 19801: 18510, 19802: 49, 19803: 49, 19805: 4557, 101727: 49, 19810: 7117, 85347: 3960, 85349: 495, 19814: 2227, 19817: 3107, 19818: 295, 19819: 7296, 19820: 2206, 19821: 14066, 19822: 49, 19823: 8858, 19824: 17066, 19825: 46864, 19826: 3038, 19828: 4059, 19829: 4117, 19830: 441, 19831: 6435, 19832: 26085, 19833: 7027, 19834: 4107, 19835: 49, 19836: 3960, 101758: 49, 19839: 19379, 19842: 2989, 19843: 1277, 19844: 23906, 19845: 2473, 19846: 7860, 19847: 9226, 19848: 32478, 19849: 642, 19851: 938, 19852: 92621, 82498: 99, 19854: 99, 19855: 196, 19856: 12798, 19857: 8769, 19858: 6027, 19859: 4018, 19861: 345, 19862: 7068, 101783: 1029, 19865: 10296, 19866: 3038, 19867: 2058, 19868: 1226, 19869: 5047, 101790: 49, 19871: 642, 19872: 4158, 19873: 1617, 19874: 9659, 19877: 7553, 19878: 938, 19880: 938, 19881: 4949, 19882: 5532, 19883: 7889, 19884: 1078, 19885: 49, 19888: 296, 101811: 49, 19892: 8967, 19894: 19899, 19895: 2058, 19896: 2277, 19897: 13889, 85434: 4306, 19899: 21293, 19900: 980, 19902: 1078, 19905: 99, 19906: 1226, 19907: 2009, 19909: 49, 19911: 4088, 19912: 17264, 19914: 1029, 19915: 4949, 19916: 8019, 19917: 295, 19918: 147, 19920: 38558, 19921: 81634, 19924: 22887, 19925: 741, 19926: 891, 19927: 75481, 19928: 98, 19931: 8512, 19932: 741, 85469: 3465, 85470: 346, 19935: 18500, 85472: 297, 19937: 147, 19938: 4165, 19939: 5187, 19941: 28016, 19942: 10731, 19943: 11058, 19944: 5384, 19946: 98, 19949: 24254, 19950: 24669, 19952: 23760, 19953: 13037, 19956: 2108, 19959: 6958, 26306: 98, 19961: 4257, 19963: 45279, 19964: 16383, 19967: 3898, 85248: 4059, 101890: 245, 19971: 24401, 19973: 11979, 19974: 1128, 19975: 47564, 19976: 4166, 19977: 840, 19978: 8127, 19979: 8651, 19980: 13403, 19981: 33857, 19982: 2918, 19983: 49, 19985: 2940, 19986: 38726, 19987: 7889, 85524: 297, 19990: 19800, 85527: 2079, 19992: 21906, 19993: 9165, 19994: 8428, 19995: 2128, 19996: 21473, 19997: 5483, 69155: 147, 52772: 99, 20005: 98, 20006: 980, 20007: 1127, 20009: 6236, 20010: 8364, 20012: 14205, 20013: 7393, 20014: 32142, 20016: 297, 20019: 16611, 20021: 2898, 20022: 11818, 85561: 2178, 20026: 6633, 20027: 23618, 20028: 790, 20029: 9163, 20030: 49, 20032: 10048, 20033: 49, 20035: 98, 20037: 2029, 20038: 148, 20039: 1127, 20041: 4483, 20042: 3087, 20044: 2989, 20045: 1029, 20046: 245, 20050: 1960, 20051: 1079, 80320: 98, 20053: 9018, 20054: 10048, 38841: 99, 20056: 49, 20057: 6076, 20058: 14520, 20059: 2720, 20060: 3829, 85599: 148, 20064: 3283, 20066: 4067, 101987: 49, 101988: 1519, 20069: 3087, 20071: 8838, 85608: 99, 20074: 6027, 20075: 15939, 20076: 6830, 20078: 49, 102000: 49, 20082: 31906, 20084: 98, 20086: 3059, 20087: 6920, 20088: 2701, 20089: 539, 85626: 1980, 36475: 4059, 20092: 6226, 20093: 3384, 20096: 393, 20097: 2009, 20098: 98, 20099: 1029, 20100: 1617, 20105: 9800, 20106: 2918, 20107: 49, 20108: 2079, 102031: 49, 20112: 49, 102034: 49, 85651: 297, 20116: 98, 20117: 5920, 20118: 4165, 20119: 14966, 20120: 13769, 20121: 49, 20122: 2989, 20123: 9947, 20124: 2058, 36510: 2178, 20127: 542, 85666: 297, 20131: 49, 20132: 36234, 20135: 90969, 20137: 196, 20138: 16145, 20139: 3235, 20141: 5878, 20142: 3483, 20143: 2166, 52912: 98, 102065: 49, 20147: 10077, 20148: 32460, 60702: 297, 20150: 98, 20151: 101720, 20152: 49, 20153: 49, 20155: 6068, 20156: 10927, 20157: 3920, 20158: 38844, 20159: 10494, 20160: 1029, 20161: 11809, 20164: 16897, 85702: 49, 20168: 12740, 20169: 4920, 20170: 17919, 20171: 9979, 20172: 544, 102094: 49, 20176: 10878, 20177: 445, 20178: 6731, 20181: 98, 20182: 1029, 20183: 8067, 20184: 2009, 20185: 4780, 20186: 98, 20187: 14355, 20188: 22521, 20189: 1128, 85726: 198, 20191: 12937, 20192: 8591, 20193: 18116, 20194: 4018, 20195: 196, 20196: 147, 20197: 17919, 20198: 8710, 20201: 8433, 20202: 10027, 20204: 21579, 20205: 6344, 20206: 3969, 20207: 1029, 20208: 245, 20209: 148, 20210: 840, 20211: 49, 20212: 30096, 20214: 25806, 19753: 3089, 20216: 7504, 102137: 49, 20218: 1127, 20219: 9621, 20222: 642, 20223: 2524, 20224: 2960, 20225: 2602, 20226: 4117, 20227: 39600, 20228: 28094, 20229: 3038, 20230: 294, 20232: 9769, 20233: 99, 20234: 2989, 20235: 148, 20236: 1421, 20237: 98, 20238: 445, 20239: 1078, 20240: 33145, 20241: 1078, 20242: 1078, 20243: 49, 102164: 49, 20245: 10690, 20246: 1960, 20247: 5265, 20248: 4117, 20250: 1078, 20251: 2989, 20253: 294, 20254: 294, 20255: 29837, 20256: 2671, 19760: 147, 20258: 3969, 20260: 5880, 20261: 6958, 20262: 5166, 20263: 345, 20266: 20727, 102187: 49, 20268: 2206, 20269: 2940, 20270: 741, 20271: 4552, 20272: 1029, 20273: 1862, 20274: 444, 20275: 30620, 20276: 7938, 20277: 49, 20279: 6068, 20281: 2009, 20283: 2940, 85820: 6435, 20285: 6860, 102206: 49, 85823: 99, 20288: 49, 20290: 2058, 20292: 15928, 20295: 15907, 20299: 2009, 20300: 19245, 20301: 3355, 20302: 1323, 20304: 49, 20305: 17341, 20306: 49, 20307: 29124, 102228: 1029, 20309: 49, 20311: 148, 20312: 7018, 20313: 8869, 20314: 8979, 20316: 16680, 20317: 839, 20318: 1375, 20319: 2157, 20320: 13678, 20322: 15840, 20323: 49, 20324: 3898, 102245: 1029, 20326: 49, 20327: 4087, 20329: 35640, 20332: 2157, 20334: 20493, 20335: 5087, 20336: 3038, 20337: 20027, 20339: 2553, 20340: 8373, 20341: 5978, 20342: 4356, 20344: 4483, 20345: 49, 20349: 4606, 20350: 6414, 20351: 3969, 20353: 7858, 20354: 5047, 20356: 49, 20358: 11880, 19777: 49, 20361: 7920, 20363: 5979, 20364: 148, 20366: 1980, 20368: 1029, 20369: 2574, 20370: 128662, 85907: 495, 69525: 4257, 85910: 198, 20375: 8968, 20376: 49, 20377: 7650, 20379: 49, 20380: 4653, 20381: 14560, 20382: 2989, 20383: 9849, 36768: 5940, 20385: 2178, 20387: 72138, 20389: 7821, 20392: 15740, 20394: 3185, 85931: 2079, 20396: 4900, 20399: 56218, 20400: 13818, 20401: 4067, 20402: 2079, 20403: 2058, 85940: 49, 20406: 10543, 19785: 3107, 85944: 6138, 20409: 148, 20410: 6860, 20412: 296, 20413: 8612, 20417: 198, 20420: 9106, 20421: 99, 20422: 6039, 20424: 197, 20426: 3186, 20428: 18620, 20429: 1078, 20430: 12423, 20431: 343, 20432: 8037, 20433: 1029, 20434: 494, 20435: 99, 20437: 25778, 20438: 345, 20439: 6186, 20440: 1276, 20441: 5039, 85680: 1980, 20444: 5494, 20445: 1960, 20447: 4158, 20448: 980, 19792: 2079, 20450: 12749, 20452: 14875, 20453: 2503, 20454: 246, 20456: 197, 20457: 49, 20458: 98, 20459: 98, 20460: 1960, 85330: 2475, 20463: 10443, 36848: 5940, 20465: 6880, 36850: 1078, 20467: 98, 20468: 98, 20469: 4998, 20470: 49, 86007: 99, 20474: 4998, 20475: 345, 20476: 3039, 20478: 4087, 20479: 1980, 20481: 10978, 86018: 445, 20486: 741, 20487: 6265, 20488: 18117, 20489: 20145, 20491: 39298, 86028: 2277, 20493: 5978, 20496: 4759, 20497: 19244, 36882: 4059, 20499: 1225, 20500: 938, 20502: 29667, 20503: 12176, 20504: 2940, 20506: 11809, 20507: 1325, 20508: 3382, 20509: 3969, 20511: 4285, 86049: 3960, 20514: 246, 20515: 444, 20516: 494, 20517: 742, 20520: 2009, 20524: 7126, 11613: 99, 20528: 32585, 20529: 9898, 20532: 1474, 20533: 13046, 20534: 4900, 20535: 980, 20536: 20897, 20537: 1138, 20538: 1079, 20539: 296, 20541: 13860, 20542: 99, 20543: 493, 20544: 6236, 20545: 938, 20546: 1078, 102467: 49, 20548: 15927, 20550: 21619, 20551: 490, 20552: 1078, 20553: 5900, 20555: 49, 20556: 1128, 20558: 14482, 20560: 5978, 20561: 980, 20562: 3136, 86099: 6138, 20564: 1960, 20565: 6958, 20566: 246, 86107: 148, 20572: 1128, 20573: 10196, 20574: 8167, 19813: 295, 22177: 197, 20577: 7938, 20578: 643, 20579: 1127, 20580: 32610, 20582: 2722, 20583: 6958, 20586: 20776, 36971: 3108, 20590: 49, 20591: 2849, 20592: 9849, 20593: 495, 20594: 2108, 86133: 27819, 20599: 5632, 20600: 3384, 69753: 99, 86138: 3267, 20603: 980, 20604: 1177, 85354: 1980, 20606: 12789, 22549: 690, 20609: 2276, 20611: 8869, 20613: 14749, 20616: 3087, 20617: 1325, 20618: 345, 20619: 5978, 20620: 196, 20621: 1029, 53390: 198, 20623: 8097, 20624: 3940, 20625: 10880, 20628: 2424, 20629: 2718, 20630: 198, 20631: 2108, 102552: 49, 20633: 19789, 20634: 3960, 20636: 147, 86173: 8118, 20639: 2940, 20640: 24453, 20641: 7889, 20643: 10939, 53412: 6435, 20645: 4257, 20646: 19985, 20647: 392, 20649: 1960, 20650: 98, 20651: 3920, 20653: 1523, 20654: 544, 20655: 13769, 20656: 2107, 20657: 10880, 53427: 593, 20661: 30865, 20663: 157178, 86200: 99, 20665: 543, 20667: 49, 86204: 2079, 20669: 3969, 20670: 1917, 86207: 11374, 20672: 147, 20673: 197, 19783: 76200, 20677: 4147, 20680: 6860, 20681: 1078, 20682: 1960, 20683: 13959, 20685: 148, 20687: 642, 20688: 2989, 20690: 13502, 20691: 1079, 20692: 3237, 20693: 246, 20694: 1225, 20696: 30778, 20697: 543, 20698: 5247, 20699: 8820, 20700: 7987, 20701: 98, 20702: 33954, 20703: 12650, 20704: 1980, 20705: 98, 86243: 4059, 20709: 75306, 20710: 49, 20711: 6633, 20712: 98, 86249: 148, 20714: 9255, 20715: 6463, 20716: 3920, 102638: 49, 20719: 148, 20720: 544, 20721: 4018, 20722: 49, 82643: 99, 20725: 839, 86262: 4158, 20730: 30786, 20733: 938, 20736: 10829, 86273: 4256, 20738: 938, 20739: 13700, 86278: 198, 20743: 9948, 20746: 246, 20748: 23205, 102669: 1029, 20752: 148, 90840: 6920, 20755: 6237, 86294: 1980, 20759: 8789, 20762: 3038, 20763: 12870, 20766: 9978, 20767: 297, 20768: 2009, 20769: 14847, 20770: 3969, 20771: 1078, 20775: 1424, 20777: 246, 102698: 147, 20779: 1127, 37164: 49, 20781: 3157, 86318: 8077, 20785: 6999, 20786: 740, 20787: 6039, 20788: 4018, 20789: 4361, 20790: 7007, 20791: 980, 20792: 693, 37177: 2079, 86330: 99, 20795: 1029, 20796: 13680, 20797: 297, 20798: 297, 20802: 938, 20803: 12581, 20805: 49, 20806: 197, 20807: 99, 20812: 7920, 20813: 3663, 20814: 49, 19853: 32766, 20816: 1178, 20818: 49, 20819: 21917, 20822: 99, 86359: 148, 20825: 98, 20827: 34301, 20828: 1078, 20832: 8380, 20834: 1177, 22587: 5047, 102759: 49, 20841: 3087, 20843: 3038, 20844: 98, 20845: 99, 20846: 1375, 20847: 3108, 20848: 49, 20849: 2009, 20853: 980, 20854: 980, 86392: 297, 20857: 98, 102778: 49, 20859: 49, 20860: 4700, 20863: 7561, 86400: 198, 20866: 13541, 20867: 1078, 20868: 4257, 37253: 495, 86406: 891, 20872: 3087, 20873: 4059, 20874: 5117, 20875: 98, 20876: 197, 20877: 4989, 86414: 198, 20880: 5878, 20882: 1425, 86419: 3663, 86420: 99, 20885: 49, 20886: 15926, 20887: 3186, 20888: 23996, 20890: 98, 20891: 4969, 20892: 99, 70045: 99, 20897: 21275, 20898: 49, 20899: 4346, 20900: 25164, 20902: 7224, 102825: 2009, 20906: 198, 20909: 52664, 20910: 196, 20913: 13493, 20914: 4284, 23610: 49, 20916: 3920, 19870: 15046, 20919: 1960, 20920: 637, 86457: 36134, 20923: 7364, 20924: 24846, 20925: 21946, 20927: 4009, 20928: 98, 20929: 2820, 20930: 12086, 20936: 98, 20937: 98779, 70091: 49, 20940: 3730, 20941: 25876, 20942: 44004, 20943: 14354, 86480: 4059, 86481: 49, 86482: 148, 20947: 18066, 20948: 9542, 20949: 7988, 20950: 543, 86489: 148, 70106: 891, 20955: 8235, 20956: 2107, 20959: 147, 20961: 17670, 94535: 2009, 20963: 2918, 70116: 49, 20965: 20669, 20966: 7028, 20967: 393, 25794: 494, 86505: 297, 20970: 2009, 20972: 3799, 20973: 6771, 20974: 6097, 20975: 297, 20976: 3038, 20979: 2989, 86516: 594, 20982: 49, 20983: 7889, 20984: 49, 20985: 980, 20992: 441, 20994: 98, 20995: 3598, 86532: 2079, 20997: 1226, 86535: 6237, 21000: 3969, 21001: 2107, 102922: 2009, 21004: 147, 21005: 98, 21006: 49, 86543: 198, 21008: 2029, 21013: 9185, 21018: 7969, 21021: 11126, 21022: 3969, 86559: 98, 21024: 5245, 85424: 2079, 21026: 1276, 53796: 1980, 21029: 4018, 21030: 1078, 21032: 49, 21033: 9146, 21035: 4018, 21037: 4306, 21038: 980, 102960: 49, 21042: 1029, 102963: 49, 21044: 4305, 21045: 49197, 21046: 495, 21047: 39399, 21048: 2940, 21049: 3701, 21051: 11225, 21052: 16680, 21054: 10127, 21055: 3849, 21059: 99, 21060: 980, 21062: 4088, 21063: 3306, 102984: 49, 21066: 7939, 21067: 1127, 21068: 980, 86606: 891, 21072: 2009, 21074: 1980, 21075: 197, 25358: 3920, 21078: 980, 21079: 49, 21080: 2107, 21082: 9542, 21084: 3920, 19898: 3969, 21087: 2989, 21089: 593, 21091: 19800, 21094: 345, 86631: 49, 21096: 147, 21098: 735, 21099: 2989, 21101: 246, 21102: 197, 21104: 197, 25363: 13464, 21108: 4059, 69054: 99, 21111: 49, 21112: 8145, 21116: 49, 21117: 345, 21118: 4998, 21120: 2306, 21123: 5137, 21124: 7301, 21125: 345, 103048: 1029, 21129: 5643, 21130: 147, 21131: 11858, 21132: 8316, 21135: 3969, 23378: 49, 21137: 245, 21138: 2058, 21139: 8217, 21142: 4009, 86679: 693, 21144: 246, 21145: 49, 103068: 49, 21149: 1474, 21150: 441, 21151: 641, 21153: 7344, 21155: 9929, 21158: 1960, 21161: 1960, 21164: 5978, 21168: 3087, 21169: 98, 21171: 2255, 21172: 99, 21173: 16165, 21176: 11286, 21177: 1326, 21178: 247, 21179: 98, 21181: 2079, 21182: 1127, 21183: 147, 21184: 99, 21185: 49, 21186: 2326, 21187: 7252, 21188: 18986, 86725: 8316, 21191: 99, 37577: 99, 21195: 3087, 21196: 7838, 86733: 495, 93645: 297, 21200: 6039, 21201: 98, 86740: 49, 21206: 19838, 86743: 980, 8996: 28858, 21210: 6296, 21211: 6760, 21213: 2989, 21214: 2009, 90917: 198, 21216: 147, 21217: 49, 21221: 6237, 21223: 39932, 21224: 147, 21225: 3038, 21226: 444, 21227: 64340, 21228: 6433, 21229: 642, 21233: 297, 21235: 2523, 21236: 637, 21984: 1078, 86774: 693, 21239: 49, 21240: 692, 21241: 1127, 21242: 5857, 21243: 98, 21244: 11769, 21245: 2058, 21247: 2009, 21249: 7077, 21250: 2029, 21252: 297, 21253: 49, 21254: 148, 21255: 6087, 21259: 14244, 21260: 99, 103182: 2989, 21264: 1127, 21265: 2177, 86802: 13959, 21267: 16038, 21270: 147, 21271: 98, 21272: 99, 21273: 3969, 103194: 1029, 21275: 20097, 21278: 3969, 21285: 1623, 103206: 49, 41777: 9900, 21288: 49, 21289: 5978, 21290: 49, 21292: 1176, 37677: 6237, 21296: 99, 86833: 99, 21298: 98, 21300: 1960, 86837: 495, 86843: 1980, 21311: 15939, 70464: 49, 19936: 1980, 21315: 9146, 21316: 2205, 103237: 49, 21320: 98, 71820: 99, 86859: 99, 21324: 4849, 21326: 46132, 21327: 245, 21328: 148, 21333: 49, 21335: 2107, 21338: 6384, 21339: 1127, 25402: 297, 21343: 5940, 21344: 49, 21345: 4067, 21346: 4256, 25403: 198, 21348: 2009, 21351: 445, 21353: 49, 21354: 49, 21355: 9225, 21357: 7541, 21359: 49, 21361: 197, 21362: 443, 21364: 4156, 21366: 49, 21367: 294, 21368: 3773, 21370: 2058, 21373: 245, 21375: 246, 21376: 49, 21377: 1029, 21379: 49, 21380: 196, 21381: 543, 21382: 2107, 21384: 980, 21386: 2305, 21387: 49, 21389: 1326, 21390: 2695, 103312: 49, 21393: 9699, 19807: 5286, 103315: 6909, 21396: 2255, 86933: 99, 21399: 2325, 21400: 4998, 21401: 246, 75455: 99, 21406: 980, 21407: 6235, 21410: 1177, 21412: 32808, 21415: 2127, 21416: 99, 21417: 15084, 21419: 11880, 21420: 1127, 21421: 9702, 86959: 245, 21426: 21027, 21427: 148, 21430: 444, 21434: 99, 21435: 4018, 25418: 3663, 21438: 49, 25517: 49, 21440: 147, 21441: 1324, 25419: 1980, 21444: 8007, 21446: 26136, 21447: 148, 21448: 6265, 21453: 1127, 21454: 197, 21455: 13928, 21456: 10543, 19960: 5394, 21458: 49, 21459: 1227, 21460: 2079, 21461: 28214, 21462: 9919, 86999: 99, 21464: 14017, 21465: 3960, 21466: 147, 25423: 49, 21468: 1226, 22693: 147, 21472: 2079, 35631: 17919, 21474: 17689, 21475: 980, 21478: 13067, 21484: 2178, 21485: 12908, 21487: 99, 21488: 99, 21489: 29253, 21490: 6097, 21491: 98, 21493: 5878, 21495: 49, 21497: 1960, 21499: 10978, 21500: 544, 21504: 980, 21509: 3920, 21511: 49, 87048: 735, 21513: 8918, 21514: 1029, 21515: 4920, 21516: 5374, 19970: 8245, 87055: 593, 85279: 99, 21521: 10345, 21523: 1029, 87060: 396, 21528: 61328, 21529: 99, 21530: 18511, 21532: 3465, 21533: 99, 21534: 3969, 21536: 1029, 21537: 3235, 21539: 7008, 21540: 6909, 21543: 147, 21544: 148, 21545: 1980, 21546: 10195, 21549: 6731, 21551: 4018, 21554: 148, 21555: 1078, 21557: 5315, 21558: 4157, 103479: 1029, 21561: 3087, 21564: 3433, 85514: 99, 21566: 49, 87103: 19503, 21569: 1078, 21570: 10959, 21571: 296, 21572: 2989, 87110: 99, 21576: 2474, 21577: 62562, 21578: 1029, 21579: 16085, 25442: 2721, 87121: 198, 21586: 147, 103507: 49, 21588: 33697, 21591: 9949, 21592: 13018, 103513: 1029, 21595: 4949, 21596: 735, 21598: 3960, 21604: 35529, 21605: 49, 103527: 49, 21608: 19373, 21610: 23618, 21611: 98, 21615: 3108, 21618: 20114, 52755: 49, 21620: 49, 21621: 6125, 21622: 49, 21623: 49, 25450: 4059, 21631: 2989, 21633: 14652, 87170: 99, 21635: 4681, 21636: 147, 21637: 444, 21638: 147, 21639: 6364, 21640: 392, 21641: 11809, 21642: 1176, 21643: 7969, 21644: 245, 70797: 148, 21647: 197, 21650: 1029, 21653: 1523, 21654: 11979, 21655: 98, 21656: 1771, 88525: 2277, 21658: 42278, 25455: 6979, 21660: 197, 21661: 29984, 21666: 6969, 21667: 833, 21668: 10098, 87206: 594, 103591: 4900, 21673: 4384, 21675: 5146, 21679: 9999, 22728: 1980, 44367: 98, 25459: 441, 54452: 49, 21685: 147, 21686: 686, 21690: 98, 21693: 3969, 21695: 3038, 21696: 49, 21698: 3989, 21700: 14354, 90998: 99, 21703: 980, 21705: 10838, 21707: 2079, 21709: 1029, 21711: 9048, 21712: 49, 21715: 642, 87254: 99, 21719: 5880, 21720: 980, 21721: 1078, 21722: 1029, 21725: 99, 21726: 7225, 21728: 5929, 21729: 1335, 21730: 8673, 21733: 16650, 21735: 4265, 21736: 30945, 87274: 49, 21741: 1029, 21743: 6419, 21744: 49, 21746: 938, 21748: 343, 21751: 4800, 21753: 32519, 87290: 4158, 21755: 2504, 21757: 1128, 21758: 98, 21760: 1029, 54530: 99, 70916: 3960, 21765: 1881, 21766: 4236, 85548: 2277, 21770: 43478, 87307: 7425, 21773: 49, 87312: 99, 21783: 4949, 21788: 297, 21792: 1089, 38177: 99, 54566: 692, 93745: 12176, 21800: 147, 21803: 2989, 21808: 148, 21809: 296, 21810: 16660, 21811: 147, 87350: 2079, 21816: 98, 21817: 198, 21818: 16037, 21819: 6493, 21820: 12948, 21826: 2058, 21828: 11977, 103751: 49, 21833: 98, 21834: 49, 21835: 7364, 21836: 343, 21839: 2940, 21841: 6305, 21842: 4059, 21844: 10878, 21847: 2128, 25487: 297, 21852: 147, 39141: 99, 21856: 196, 21858: 3286, 21859: 3969, 21862: 2601, 21863: 49, 21865: 24253, 21866: 16343, 103787: 49, 87404: 99, 21870: 12056, 21872: 196, 21875: 2256, 21878: 49, 25492: 1523, 103803: 49, 21889: 19779, 21890: 3207, 87427: 49, 21895: 7265, 21900: 6534, 103821: 49, 71055: 99, 21904: 7056, 21905: 2009, 103826: 49, 21907: 49, 21908: 1128, 21910: 6186, 21911: 343, 21914: 98, 54683: 5940, 21916: 3187, 103837: 343, 21918: 980, 85573: 99, 21920: 441, 21921: 49, 87458: 99, 21924: 10106, 71078: 99, 21928: 790, 21929: 12127, 21930: 1029, 21932: 1128, 21933: 7057, 21934: 49, 54703: 297, 54706: 99, 21940: 14948, 21942: 197, 21943: 98, 21944: 2009, 21946: 3940, 21948: 98, 21949: 49, 87491: 3058, 21956: 10077, 21957: 1029, 21961: 13092, 21964: 23757, 87502: 99, 21967: 1127, 25507: 245, 87508: 297, 21974: 4780, 21975: 49, 21977: 5346, 21978: 5929, 21982: 3435, 87520: 198, 21986: 29905, 71139: 49, 21989: 246, 21992: 14433, 71147: 8019, 21998: 1029, 87535: 4257, 25512: 6336, 87538: 4108, 22003: 49, 22004: 2108, 22006: 345, 22007: 49, 22009: 98, 22010: 17740, 25515: 1721, 22027: 12452, 22029: 3920, 71183: 4752, 22035: 2059, 22039: 148, 22041: 296, 38428: 49, 22045: 148, 22046: 2058, 68384: 27373, 20878: 99, 22052: 1029, 22053: 1374, 22055: 64774, 22056: 980, 93788: 296, 22059: 49, 22064: 5048, 22066: 49, 74675: 3465, 22069: 16482, 22070: 197, 22072: 445, 82868: 99, 20063: 32618, 22079: 5662, 22080: 6454, 22085: 99, 22091: 2009, 22093: 147, 22094: 49, 22096: 1830, 87635: 9374, 22102: 98, 20068: 296, 22106: 7125, 22109: 27748, 22110: 3038, 22111: 49, 22112: 4257, 22113: 245, 22116: 980, 87654: 7821, 38504: 99, 91068: 148, 22122: 246, 54891: 99, 92937: 296, 22126: 49, 74685: 49, 41918: 2277, 22804: 99, 22140: 2960, 22141: 6315, 22142: 148, 87680: 346, 25536: 4059, 85612: 99, 22159: 98, 22167: 2178, 22171: 444, 22172: 98, 22173: 245, 22174: 99, 22176: 3969, 20080: 6671, 87715: 99, 82886: 12474, 22183: 5445, 22184: 13769, 22185: 3969, 22187: 49, 22189: 396, 22190: 9463, 87728: 4059, 22199: 3969, 22200: 4998, 22201: 692, 22202: 2523, 22203: 2969, 54974: 2376, 87743: 148, 22208: 980, 22209: 76721, 25547: 99, 22212: 3960, 22213: 5483, 22214: 245, 22217: 11471, 22220: 4018, 22223: 2940, 22224: 49, 22226: 5741, 22819: 642, 54997: 49, 38614: 49, 22231: 49, 22233: 1236, 22239: 13323, 22240: 3059, 87777: 4455, 22242: 693, 87779: 147, 55012: 11725, 22245: 980, 22246: 10828, 22251: 16026, 22252: 1029, 22257: 1960, 22258: 16066, 22262: 38737, 22265: 10028, 22269: 396, 87806: 2326, 87808: 4059, 22273: 245, 22275: 18669, 87812: 495, 91095: 4236, 22284: 247, 87821: 16038, 22286: 4455, 22829: 1029, 22288: 4998, 22290: 2009, 87828: 2475, 87831: 99, 22297: 980, 87837: 7474, 22302: 198, 87841: 297, 87845: 2178, 22310: 17066, 22313: 99, 22314: 8166, 87855: 49, 22321: 21015, 22322: 14670, 87859: 99, 19838: 4940, 22327: 245, 87864: 1980, 22329: 3960, 22330: 4138, 22331: 4336, 22332: 4058, 22333: 4374, 22335: 589, 22339: 245, 22341: 3186, 91105: 198, 22344: 12096, 22346: 18728, 22349: 295, 22350: 49, 22354: 1424, 25571: 99, 22356: 3969, 22357: 939, 55126: 99, 22359: 2157, 38934: 99, 20111: 32669, 22365: 4459, 55136: 99, 22376: 4108, 38762: 49, 22379: 296, 22380: 5462, 52882: 98, 22383: 9999, 20115: 12740, 87926: 297, 22568: 346, 22399: 20450, 22400: 7957, 87938: 296, 55171: 8019, 22405: 99, 22406: 49, 22408: 12700, 22410: 3107, 87948: 297, 22413: 147, 22415: 19899, 22417: 5929, 38802: 49, 22420: 490, 22424: 5532, 22426: 1177, 22429: 1960, 38814: 99, 22435: 49, 25585: 99, 87977: 100135, 55211: 49, 22444: 1078, 22446: 98, 22452: 49, 22455: 26967, 22457: 3108, 22458: 99, 22459: 1029, 71613: 99, 22463: 1980, 55232: 49, 22465: 3186, 22467: 29947, 22470: 889, 22471: 2078, 22476: 2940, 20130: 12383, 22478: 3602, 88015: 2178, 91128: 49, 22484: 49, 22487: 98, 22488: 49, 22489: 98, 22495: 2404, 22501: 543, 22502: 4356, 22506: 6138, 22507: 49, 22509: 30478, 22510: 980, 58365: 148, 22513: 495, 22514: 5215, 22516: 4949, 22521: 49, 22524: 2009, 22528: 544, 22529: 21044, 22530: 16086, 22531: 2940, 22533: 9146, 22535: 148, 22537: 2989, 22540: 5513, 22541: 98, 22542: 147, 88082: 246, 85678: 1029, 22550: 2128, 22555: 11880, 22557: 99, 20144: 147, 22562: 49, 22564: 1078, 88103: 49, 88104: 49, 77490: 99, 22577: 2989, 22578: 148, 22579: 49, 22580: 297, 22581: 98, 22583: 42469, 22585: 45577, 22879: 6285, 22590: 2518, 22593: 10789, 22598: 28003, 22600: 4455, 22601: 3940, 22605: 5483, 22606: 3137, 22610: 3479, 22615: 98, 22619: 99, 22620: 42201, 22621: 66850, 22624: 99, 22626: 294, 22627: 12177, 22628: 1980, 22631: 2128, 88172: 99, 71792: 1980, 74270: 99, 22653: 4018, 22654: 1029, 22656: 2079, 22658: 98, 22660: 16236, 22663: 16125, 22665: 49, 22666: 6878, 88203: 49, 88204: 5247, 22669: 2918, 22670: 6336, 91160: 99, 22674: 1960, 22679: 4849, 88216: 10197, 22681: 2940, 22685: 7938, 22686: 9027, 22690: 98, 22691: 49, 22692: 1127, 20166: 3675, 22694: 22558, 22695: 1128, 22700: 14325, 31090: 544, 22702: 49, 71855: 49, 22706: 2206, 22707: 20047, 22708: 3088, 39095: 99, 22714: 980, 22716: 692, 22719: 43200, 22726: 99, 22727: 5920, 88264: 49, 22736: 7920, 22737: 980, 22740: 45478, 22742: 3186, 22744: 148, 22745: 14334, 22749: 6197, 22750: 2009, 22751: 12849, 22755: 36853, 22757: 49, 22759: 98, 22763: 1029, 39293: 99, 22772: 294, 22774: 49, 22777: 5738, 22778: 4356, 22779: 2058, 22780: 10048, 25642: 99, 22782: 98, 22786: 2178, 22787: 3038, 22789: 49, 22791: 49, 22801: 6039, 88340: 1485, 22806: 1960, 22808: 1980, 22811: 2761, 71965: 99, 22816: 3335, 88355: 198, 71973: 99, 22822: 1326, 88359: 49, 22825: 445, 22826: 98, 91186: 99, 22831: 49, 22834: 8196, 22835: 1523, 22837: 396, 22839: 13167, 91188: 19998, 22842: 49, 22847: 20497, 22848: 1029, 22849: 49, 88386: 99, 22852: 99, 22853: 2108, 22854: 4513, 22859: 49, 52962: 4158, 22865: 4661, 88405: 980, 22870: 148, 25657: 7176, 22875: 889, 39263: 4059, 22880: 25975, 22882: 99, 39267: 49, 22889: 5292, 22892: 741, 22895: 742, 22905: 2156, 22906: 10977, 91199: 99, 22908: 10275, 22909: 543, 22912: 742, 22913: 4116, 22919: 494, 22920: 2820, 22921: 97858, 22926: 148, 22928: 16174, 22929: 99, 22931: 693, 22932: 30916, 22935: 23569, 22936: 20154, 22937: 49, 22939: 395, 22940: 49, 22942: 49, 22943: 980, 22944: 8097, 22945: 980, 22946: 247, 22947: 60189, 20884: 18669, 22955: 3267, 36594: 99, 42056: 6039, 22963: 5880, 22965: 1474, 22966: 4207, 25315: 99, 22968: 49, 22969: 27175, 22974: 4341, 22978: 13512, 22981: 147, 22982: 20777, 22985: 4206, 22986: 980, 20215: 10640, 22989: 3871, 22990: 3381, 22991: 147, 22997: 148, 22998: 49, 22999: 148, 23001: 197, 23002: 99, 19575: 12304, 23005: 4137, 23012: 5940, 23016: 1029, 23018: 3087, 23020: 395, 23024: 296, 23025: 15779, 23031: 1127, 23033: 395, 23039: 196, 23044: 8217, 72197: 245, 23046: 13602, 72201: 2474, 23051: 196, 23054: 6138, 23055: 12671, 23056: 49, 23057: 2058, 23058: 3256, 23060: 99, 42073: 49, 23064: 18017, 23067: 296, 23068: 99, 23069: 6187, 23072: 6236, 23074: 8365, 23075: 9555, 23076: 980, 23077: 27918, 23078: 49, 23080: 297, 25693: 4455, 23090: 2277, 39476: 14058, 23095: 1078, 23096: 20047, 23097: 4266, 23098: 3960, 23099: 4285, 23100: 7029, 23101: 13959, 23103: 2375, 23108: 4414, 23109: 3969, 23111: 6860, 23112: 1029, 23113: 4414, 88650: 99, 23118: 147, 39352: 297, 23124: 34984, 23127: 98, 23130: 98, 23132: 18899, 23135: 98, 23137: 4949, 23139: 44061, 23140: 1029, 23141: 9968, 23143: 3187, 23145: 980, 23146: 3960, 23148: 98, 23149: 21748, 23150: 2009, 23151: 7819, 88688: 99, 28435: 49, 23157: 2940, 23160: 4039, 20244: 22176, 23162: 4236, 53691: 99, 23169: 3306, 23170: 19948, 23171: 20093, 23172: 23252, 23174: 49, 23175: 2058, 23182: 49, 23183: 1960, 23184: 49, 23185: 49, 23188: 148, 23189: 25868, 23191: 10046, 23197: 296, 23200: 3762, 25712: 3465, 23209: 790, 70605: 49, 23211: 643, 55981: 22077, 23215: 9800, 23216: 148, 39602: 396, 23225: 5880, 23228: 49, 23231: 395, 23233: 3750, 20257: 13818, 56008: 693, 23241: 33509, 23243: 98, 23244: 49, 39631: 2079, 23248: 2524, 88785: 49, 23253: 2128, 23259: 245, 23260: 980, 23263: 49, 56032: 99, 23267: 49, 23270: 55602, 47569: 99, 23272: 49, 23273: 27817, 39665: 49, 88818: 99, 88827: 49, 23293: 346, 20267: 3969, 23309: 6039, 88847: 4158, 23321: 15897, 23323: 99, 88861: 16285, 74885: 99, 23330: 544, 23331: 9997, 23332: 1128, 23335: 98, 23338: 10958, 23340: 245, 23342: 99, 56112: 148, 85811: 99, 23350: 1089, 72506: 148, 23355: 99, 23356: 539, 23357: 147, 88894: 4059, 23359: 49, 23363: 2058, 88900: 396, 94007: 2079, 23373: 26629, 23374: 3479, 23375: 4425, 23377: 833, 88914: 5039, 23389: 1980, 23391: 98, 23393: 99, 23396: 2574, 88934: 297, 25745: 2178, 76067: 198, 88938: 8019, 88939: 99, 23411: 14750, 23412: 395, 20286: 86898, 88952: 99, 23418: 23997, 20287: 15027, 23424: 49, 23425: 1029, 23428: 1980, 23429: 98, 23434: 198, 88975: 297, 23441: 5039, 39827: 1980, 23448: 49, 23449: 49, 83098: 297, 23454: 49, 31216: 49, 25755: 539, 23461: 345, 88998: 4158, 102215: 49, 23468: 2613, 23471: 26609, 23473: 741, 89011: 99, 23477: 7117, 23482: 1960, 91295: 247, 23484: 1079, 23487: 6958, 23489: 4900, 23492: 49, 85836: 8820, 89034: 99, 23502: 49, 23503: 539, 23504: 99, 89041: 297, 23506: 2059, 23517: 5414, 89055: 49, 23520: 49, 89060: 2178, 56296: 99, 91893: 49, 23547: 346, 23552: 49, 23554: 198, 39939: 346, 23556: 296, 23558: 980, 89095: 49, 89097: 20394, 26426: 198, 83117: 495, 23568: 245, 23570: 99, 89110: 198, 85849: 99, 20217: 4949, 23587: 5068, 23594: 3483, 23600: 13958, 23608: 1177, 56378: 494, 23613: 2178, 89152: 148, 89154: 25839, 23619: 10246, 23622: 99, 23625: 6272, 23626: 17968, 45523: 49, 23632: 49, 23634: 49, 23155: 1029, 23639: 12107, 23655: 147, 23656: 1029, 89197: 49, 23662: 5929, 23667: 42557, 23671: 99, 23674: 7216, 89211: 19899, 23676: 833, 25908: 99, 23679: 10136, 23681: 1029, 72834: 99, 23684: 1128, 23685: 35153, 23687: 495, 23688: 148, 23693: 3286, 25799: 99, 23700: 49, 91332: 198, 89248: 5049, 89249: 493, 23715: 49, 91334: 147, 23720: 980, 23721: 2108, 23723: 99, 23726: 49, 20671: 15939, 72887: 147, 23739: 2058, 89278: 49, 89280: 26037, 72902: 99, 23758: 495, 28536: 8217, 89300: 99, 23765: 980, 23766: 4780, 23769: 49, 91343: 2375, 23783: 3880, 23784: 147, 23788: 49, 23791: 980, 89333: 4059, 23798: 49, 23800: 245, 89340: 4455, 94080: 4059, 23811: 49, 23812: 245, 23815: 2207, 23817: 2079, 23820: 4500, 23821: 49, 40208: 4158, 89365: 49, 23830: 4059, 23832: 98, 91354: 346, 89374: 25938, 23841: 8217, 23842: 4463, 23845: 99, 23852: 539, 40239: 99, 23857: 247, 23858: 297, 23866: 198, 89403: 4257, 23869: 98, 89410: 7623, 23878: 245, 23884: 99, 89425: 445, 89427: 4454, 23895: 99, 88086: 1029, 23900: 148, 89437: 99, 23902: 980, 25829: 37917, 23905: 1960, 89444: 148, 92451: 2227, 23918: 490, 56688: 49, 23923: 13818, 23925: 9949, 89465: 49, 89471: 592, 40320: 99, 89474: 99, 23940: 980, 20374: 295, 89479: 4059, 89483: 2079, 23955: 49, 89496: 148, 23962: 2474, 23970: 49, 89507: 99, 23972: 147, 23973: 9651, 23974: 5096, 89511: 99, 56744: 198, 89521: 99, 40371: 49, 89525: 4059, 73143: 99, 23994: 98, 23999: 297, 89539: 5187, 56776: 99, 24009: 99, 24011: 3335, 19800: 14749, 89559: 198, 24028: 10909, 24034: 1177, 24046: 445, 24048: 99, 24051: 198, 24055: 3087, 85909: 99, 24067: 5243, 24068: 16830, 73221: 99, 24071: 148, 24094: 1673, 25862: 4504, 24109: 196, 89647: 49, 24112: 11828, 91400: 14255, 24116: 147, 89657: 4405, 24124: 441, 53723: 99, 24130: 1960, 56903: 49, 40522: 396, 20407: 37558, 89676: 297, 89681: 2079, 24148: 8008, 24152: 148, 89694: 99, 89716: 18018, 89717: 99, 24190: 49, 24194: 49, 24199: 1881, 24201: 49, 44451: 49, 24208: 13959, 24216: 11770, 25882: 3069, 89761: 2079, 24227: 1573, 24239: 6039, 25903: 245, 24243: 17277, 89784: 12374, 53194: 99, 64117: 297, 24265: 99, 24268: 198, 89805: 99, 89809: 1980, 23161: 2355, 57048: 99, 24283: 296, 40669: 49, 24288: 692, 94160: 6088, 89826: 2277, 24293: 99, 24297: 49, 73450: 198, 23165: 6662, 91432: 99, 24314: 2475, 89855: 3960, 25899: 445, 57096: 47619, 36823: 99, 89870: 99, 24335: 99, 57106: 2277, 88707: 3108, 24342: 343, 89880: 198, 24345: 99, 91439: 99, 24351: 882, 24358: 396, 24359: 6039, 24361: 9949, 40750: 49, 89905: 99, 91444: 6237, 89915: 1980, 24383: 49, 57155: 99, 24391: 2619, 89933: 247, 89935: 4009, 24400: 148, 91212: 49, 24406: 99, 73564: 1683, 94181: 346, 31377: 99, 89965: 99, 89970: 10098, 24442: 8167, 89980: 198, 24448: 98, 57217: 99, 84831: 693, 24453: 196, 24455: 99, 24458: 99, 24464: 3960, 73617: 99, 87014: 297, 24474: 1079, 90011: 396, 24480: 104007, 24481: 2574, 24484: 21175, 91462: 49, 90023: 98, 24488: 2405, 24491: 49, 24495: 197, 90032: 6286, 86004: 2277, 25930: 99, 90046: 8365, 24511: 99, 21002: 13959, 24515: 49, 90055: 99, 24525: 99, 24528: 297, 24533: 98, 24537: 7008, 90088: 297, 24555: 1078, 24556: 50687, 24559: 5273, 24561: 49, 90102: 4018, 24568: 13305, 90106: 297, 90113: 99, 24578: 49, 24580: 148, 25942: 6345, 24587: 1029, 20482: 295, 87999: 198, 24592: 49, 40977: 148, 24595: 49, 24609: 2009, 25947: 2425, 24613: 6177, 57383: 49, 90166: 99, 90171: 49, 90172: 5788, 24638: 198, 24644: 148, 90182: 32520, 20492: 7266, 24653: 245, 78291: 99, 24658: 35243, 25955: 3088, 24664: 198, 23735: 441, 24670: 6138, 83301: 99, 24672: 14879, 24678: 2602, 24679: 49, 24682: 49, 24687: 99, 24688: 3267, 24689: 445, 24691: 3960, 91497: 148, 24699: 99, 90236: 99, 41087: 49, 24705: 247, 90246: 49, 24712: 99, 24716: 637, 24719: 1029, 24721: 99, 90259: 148, 24726: 9948, 24727: 4068, 24728: 49, 24732: 7920, 24733: 247, 24741: 8019, 24743: 99, 88775: 8019, 73904: 198, 24754: 10633, 24757: 2059, 24758: 14157, 24761: 2653, 24763: 198, 24767: 5929, 24769: 148, 20513: 196, 24781: 58062, 24782: 34936, 24783: 1980, 24789: 1128, 25977: 148, 90328: 297, 90329: 99, 24795: 67622, 90336: 2079, 90342: 99, 24808: 5880, 73962: 99, 24811: 99, 90348: 49, 24815: 9444, 24817: 99, 24822: 49, 85389: 99, 91519: 2376, 41213: 99, 24834: 245, 24835: 6336, 24838: 297, 90376: 99, 24843: 148, 90380: 99, 90382: 297, 75139: 99, 90388: 99, 57622: 495, 24858: 2079, 24860: 3969, 24862: 2722, 45104: 99, 24867: 42371, 24868: 14640, 24871: 2940, 24875: 343, 24878: 49, 23742: 297, 24884: 2009, 41270: 6039, 24888: 2721, 24897: 31352, 24905: 297, 57675: 2079, 24911: 148, 24912: 69128, 24913: 46160, 24915: 637, 90452: 2079, 25998: 1960, 41316: 891, 90470: 148, 24938: 17978, 41323: 49, 90476: 49, 90484: 99, 86078: 20592, 24958: 742, 24964: 840, 24965: 148, 86081: 3960, 24968: 49, 75574: 3960, 41365: 2702, 24991: 147, 57762: 6633, 24996: 2989, 90533: 198, 24999: 8019, 90544: 1980, 86088: 197, 25013: 3108, 57787: 98, 74172: 980, 84850: 49, 25022: 1425, 25043: 1683, 25051: 345, 90589: 99, 23291: 247, 25062: 980, 90601: 49, 25073: 297, 20563: 49, 25078: 7920, 25080: 1029, 25084: 98, 26026: 1980, 90630: 148, 25386: 594, 25372: 4454, 25118: 20649, 86490: 99, 41505: 99, 90659: 1079, 25127: 98, 90668: 99, 41522: 148, 90697: 98, 74582: 1980, 26040: 4554, 23311: 2940, 25184: 147, 25185: 2989, 41572: 99, 57964: 2079, 90740: 99, 25209: 99, 25215: 198, 20587: 1226, 90757: 31878, 90759: 4059, 90761: 99, 25227: 99, 25232: 12103, 25238: 297, 25246: 247, 41632: 99, 25253: 49, 25254: 148, 90791: 99, 25260: 198, 74413: 99, 25265: 148, 25270: 148, 26057: 99, 25275: 6385, 25277: 9900, 20597: 5385, 25283: 98, 58828: 49, 25292: 147, 88866: 8118, 75213: 693, 90832: 99, 74450: 99, 25304: 5929, 20602: 3285, 90848: 15939, 90851: 2079, 58086: 99, 90855: 4108, 26066: 2326, 74485: 990, 25335: 99, 28799: 99, 24303: 49, 41730: 15840, 25347: 735, 8965: 10296, 8966: 59994, 8967: 39202, 8968: 79398, 8969: 133579, 8973: 4356, 8974: 297, 8976: 6534, 8977: 2772, 8979: 2079, 90900: 495, 8981: 27918, 8982: 99, 8983: 1187, 8984: 396, 8985: 99, 8986: 12325, 8987: 2673, 8988: 99, 8989: 4026, 8990: 8019, 8991: 99, 8992: 4851, 8993: 99, 8994: 99969, 8995: 23957, 90916: 51579, 8997: 14058, 8998: 6039, 8999: 1089, 9000: 49797, 9002: 445, 9003: 1287, 9007: 7920, 9008: 44153, 90929: 4158, 76984: 5940, 9012: 396, 9013: 50913, 9014: 18018, 9015: 8365, 25400: 99, 9017: 939, 9018: 17957, 9019: 643, 9021: 5049, 9022: 8019, 9024: 33759, 9026: 4455, 9027: 41281, 9028: 8217, 9029: 396, 9030: 8613, 9031: 16137, 9034: 2673, 9035: 9999, 9036: 99, 9037: 6691, 9038: 99, 9039: 2177, 9040: 17968, 9041: 6831, 9042: 22224, 9043: 4059, 9044: 1980, 9045: 6038, 9046: 3365, 9047: 18215, 9048: 6138, 25433: 1980, 9050: 6830, 94351: 539, 9052: 2475, 9053: 4059, 9054: 7920, 9055: 6335, 9056: 346, 9058: 22869, 9059: 2178, 9060: 1336, 9061: 297, 9062: 19800, 74599: 148, 9064: 17919, 9066: 30629, 9068: 840, 9069: 4949, 9071: 34056, 9072: 2376, 9073: 9900, 9074: 2277, 9075: 3315, 9077: 1386, 9078: 99, 9079: 154083, 9080: 543, 9081: 8464, 9082: 2277, 9083: 5841, 9084: 5088, 9085: 106946, 9086: 297, 9087: 10939, 9088: 19214, 9089: 1287, 9090: 8562, 9091: 2079, 9092: 693, 9093: 6979, 9094: 1485, 9096: 2277, 9097: 2079, 9098: 10642, 9099: 6137, 9100: 13860, 41869: 49, 9102: 25194, 9103: 21780, 9104: 8217, 20632: 1029, 9106: 5346, 9107: 99, 9108: 20196, 9109: 6039, 9111: 198, 9112: 643, 9113: 59101, 9114: 4257, 9115: 297, 25500: 1029, 9117: 38115, 9119: 2524, 9120: 4207, 9121: 1980, 9122: 2079, 9123: 86772, 9124: 73160, 9126: 1435, 9127: 18315, 9128: 9999, 9129: 4059, 9130: 9999, 9131: 12819, 9132: 178762, 9133: 7920, 9134: 6039, 9135: 15147, 9136: 247, 9139: 4158, 9141: 198, 9142: 43151, 9143: 445, 9144: 99, 9145: 891, 9146: 6682, 9147: 15840, 9148: 67733, 9149: 495, 9150: 4257, 9151: 31680, 9152: 46370, 9153: 7227, 9156: 5049, 9157: 1287, 9159: 6633, 9160: 4059, 9161: 13959, 9162: 69249, 9163: 7227, 9164: 4851, 9165: 99, 9166: 2178, 9167: 18611, 9168: 197, 9169: 6039, 9170: 9900, 9172: 99, 9174: 2079, 9175: 99, 9176: 14541, 20644: 6571, 9178: 2029, 9179: 4009, 9180: 4653, 9181: 12919, 25566: 148, 9183: 53478, 9184: 2277, 9185: 12522, 9187: 9900, 91108: 495, 9189: 8019, 9190: 1386, 9191: 2574, 9192: 38023, 9193: 247, 9194: 594, 9195: 99, 9197: 8019, 9199: 2673, 9200: 990, 9201: 29799, 9202: 1779, 9203: 12969, 58878: 99, 9208: 643, 9209: 49, 9210: 9900, 9211: 445, 9212: 9256, 91133: 99, 9214: 11236, 9215: 1980, 9217: 39946, 9218: 4257, 9219: 13959, 9220: 2277, 9222: 5266, 9223: 6633, 9224: 3168, 9225: 10147, 9226: 2277, 9228: 495, 9229: 198, 9230: 2524, 9231: 396, 9232: 4059, 9234: 9405, 9235: 4554, 9236: 148, 9237: 2079, 9238: 2128, 9239: 198, 9240: 2871, 9241: 27918, 9242: 99, 9243: 11979, 9245: 16483, 9246: 297, 9247: 99, 9249: 12126, 9251: 990, 9252: 2524, 9254: 544, 9255: 2277, 9256: 17323, 9257: 135580, 9258: 990, 91655: 99, 9260: 3118, 9261: 4059, 9262: 346, 9263: 2524, 26120: 99, 9266: 3960, 9268: 495, 9269: 6137, 9271: 1286, 9272: 2376, 9273: 495, 25658: 49, 9275: 5940, 9276: 2574, 9277: 4653, 9279: 495, 9280: 495, 9281: 1782, 9282: 31878, 9284: 99, 9285: 1287, 9286: 4059, 9287: 1980, 9288: 197, 9289: 13116, 9290: 2821, 9291: 4059, 9292: 20344, 9293: 2277, 9297: 6039, 9298: 692, 9301: 7789, 9303: 5940, 25688: 2227, 9305: 4653, 9309: 39946, 9311: 12078, 9312: 4455, 9313: 396, 9314: 9009, 25700: 6039, 9317: 940, 9318: 3514, 9319: 1779, 9320: 99, 9321: 33045, 9322: 2968, 91243: 99, 91244: 4207, 9325: 10098, 9326: 2574, 9327: 2919, 9328: 29649, 91251: 99, 9335: 2079, 9337: 4207, 9338: 297, 9339: 198, 9341: 1877, 9343: 247, 9345: 98, 9346: 99, 9347: 4454, 9348: 9405, 9349: 6336, 9350: 7325, 91271: 296, 9352: 8910, 9356: 198, 9357: 3861, 9358: 32323, 9359: 38262, 9360: 792, 9361: 198, 9364: 40192, 9367: 16038, 9368: 3069, 9369: 11820, 9371: 4059, 9373: 17788, 9374: 2178, 9375: 99, 9376: 8315, 9377: 1334, 9378: 10741, 9379: 693, 91300: 148, 9381: 99, 9384: 445, 9386: 2079, 9387: 1287, 9388: 2573, 9389: 19899, 9390: 2079, 9391: 71080, 9392: 2376, 9393: 3960, 91314: 49, 9395: 6186, 9396: 2079, 9397: 691, 9398: 12325, 9400: 2673, 9401: 26185, 9403: 4059, 9404: 4356, 9406: 1681, 9407: 10542, 9408: 5940, 91329: 4256, 9410: 841, 9412: 297, 9413: 11979, 9414: 544, 9415: 297, 9416: 2178, 9417: 198, 9418: 891, 9419: 2178, 9421: 495, 9422: 4356, 9423: 297, 9425: 3267, 25810: 980, 9427: 593, 9429: 5143, 91350: 2989, 9431: 63855, 9432: 25938, 9433: 1089, 9434: 2276, 9435: 99, 9436: 23859, 9437: 1980, 9438: 13860, 25823: 99, 9441: 6039, 9442: 2376, 9444: 4554, 9445: 50241, 9447: 792, 9449: 99, 9452: 3960, 9453: 494, 9454: 198, 9455: 2079, 91376: 16215, 9457: 297, 9458: 6039, 9459: 8068, 9460: 10395, 25845: 5385, 9462: 396, 9463: 12028, 25848: 297, 9465: 4653, 9466: 99, 91387: 49, 9469: 5940, 9470: 13959, 9471: 198, 9472: 99, 9475: 495, 9477: 2079, 9478: 2079, 9479: 99, 9480: 17919, 9481: 11880, 9485: 2079, 75022: 594, 9487: 297, 25872: 5940, 9489: 346, 9493: 6019, 9494: 2673, 9495: 2919, 9496: 495, 9497: 99, 9498: 2079, 9499: 99, 9500: 13860, 9502: 116522, 9503: 99, 9504: 65951, 9505: 99, 9506: 890, 9507: 6039, 9508: 25046, 9509: 445, 9510: 544, 9512: 13512, 86236: 198, 9514: 99, 9515: 10345, 9516: 6286, 9517: 2178, 25902: 147, 9519: 1980, 9520: 4059, 9521: 7920, 9523: 346, 9524: 396, 9525: 1980, 9527: 891, 9528: 85505, 9529: 198, 9530: 99, 25915: 49, 9532: 6482, 9533: 12622, 25918: 99, 9535: 24749, 9536: 4900, 94432: 148, 9538: 28658, 9539: 396, 9540: 891, 9541: 8019, 9542: 2474, 9543: 8563, 25928: 4059, 9545: 51480, 9546: 99, 9550: 198, 9551: 10840, 9552: 445, 9553: 15184, 9555: 99, 9556: 8266, 9557: 297, 9558: 17820, 9561: 890, 9563: 14553, 9564: 1980, 9565: 47520, 9566: 24007, 9571: 12226, 25956: 2960, 9573: 99, 9574: 11583, 9575: 8217, 9577: 13563, 9578: 2079, 9581: 841, 9582: 4751, 9583: 25938, 9584: 2177, 9585: 1386, 9586: 5940, 9587: 89793, 9589: 148, 9590: 8364, 20713: 49, 9592: 6237, 9593: 49152, 9594: 395, 9595: 12770, 9596: 3661, 9597: 297, 9598: 5296, 9599: 25838, 9600: 3108, 9601: 1881, 9602: 4356, 9603: 2079, 9605: 346, 25990: 99, 9607: 4059, 9609: 8118, 9610: 2178, 9611: 2178, 9612: 3108, 9613: 99, 9614: 891, 9615: 247, 9616: 495, 9617: 99, 9618: 99, 9619: 4356, 9620: 396, 9621: 2227, 9622: 5544, 9623: 1980, 9625: 693, 91546: 49, 9627: 2375, 9628: 247, 9629: 99, 9630: 2079, 9631: 1980, 9632: 5643, 9633: 396, 91554: 99, 9635: 693, 9636: 6435, 9638: 4257, 9640: 6138, 58793: 99, 9642: 1485, 9643: 21977, 9645: 11038, 9646: 643, 9647: 12078, 9649: 297, 9651: 693, 9652: 2178, 9653: 3267, 9654: 4158, 9655: 77475, 9656: 1287, 26041: 2475, 9660: 3267, 9661: 198, 9665: 8118, 9667: 2079, 9668: 4257, 9669: 643, 9670: 2029, 9672: 24947, 9673: 594, 9674: 14008, 9675: 6039, 9676: 198, 9677: 8019, 26062: 1980, 9679: 11880, 9680: 12870, 9682: 52052, 9683: 16928, 9686: 198, 9687: 38510, 9688: 99, 26073: 742, 9690: 24156, 9691: 51043, 9692: 5940, 9693: 99, 9694: 693, 26192: 297, 9699: 99, 9703: 4455, 9706: 2079, 91627: 5137, 9709: 198, 9710: 4257, 9711: 1287, 9712: 15691, 9713: 4257, 9714: 2970, 9717: 891, 9718: 8811, 9719: 1287, 9721: 4949, 9722: 19998, 9724: 1980, 9725: 3366, 9726: 27918, 9727: 791, 9729: 19899, 9732: 14355, 9734: 4998, 9735: 7011, 9736: 4158, 9737: 198, 9738: 45738, 87069: 13860, 9741: 19800, 9742: 9680, 9743: 1880, 9744: 395, 9745: 15047, 9746: 99, 9747: 297, 9748: 11979, 9749: 2772, 9752: 19531, 9753: 4257, 9754: 42273, 9755: 6879, 9756: 99, 9757: 444, 9758: 8118, 9759: 4653, 9760: 4059, 9761: 297, 9763: 1087, 20742: 4849, 9766: 395, 9768: 99, 9769: 11979, 9771: 1980, 19420: 2376, 9773: 49, 9774: 1681, 9778: 9601, 9780: 5940, 9781: 10098, 9783: 99, 9784: 1583, 9785: 495, 9786: 1039, 9787: 1980, 9789: 4059, 91710: 6167, 9791: 14751, 9793: 10593, 9794: 1089, 9795: 297, 9796: 2376, 9798: 148, 9800: 20492, 9801: 891, 9802: 1980, 9803: 594, 9804: 544, 9805: 1980, 9807: 16681, 9808: 1485, 9811: 1631, 9813: 1980, 9814: 5940, 9816: 3960, 9817: 99, 9818: 2079, 9819: 6633, 9822: 643, 9823: 19800, 9824: 21879, 9826: 9405, 9827: 2424, 91748: 2277, 9829: 29799, 9831: 1287, 9832: 594, 9833: 198, 9836: 5731, 9837: 2079, 9838: 1980, 9841: 692, 9844: 396, 9845: 8811, 9846: 16929, 91767: 247, 9848: 693, 9849: 11285, 9851: 1980, 9852: 396, 91775: 148, 9856: 18018, 9858: 5445, 9859: 99, 9860: 5641, 59013: 346, 9862: 6435, 9863: 396, 9868: 6880, 9869: 544, 9870: 99, 9871: 4158, 9872: 198, 9873: 2574, 9874: 296, 9875: 2079, 9877: 1286, 9879: 99, 9880: 6039, 9882: 346, 9883: 980, 9885: 6039, 59039: 98, 9889: 1584, 9891: 297, 9892: 99, 9894: 2178, 9895: 14454, 9896: 198, 9897: 14502, 9898: 891, 9899: 198, 9901: 495, 9902: 989, 9904: 544, 9905: 41230, 9906: 13067, 9908: 14652, 91829: 49, 9910: 395, 9912: 198, 9915: 14206, 9916: 990, 9918: 1980, 9919: 1980, 9920: 8019, 9921: 20097, 9922: 6039, 9923: 4158, 9924: 6633, 9925: 16731, 9926: 198, 91847: 99, 9928: 8019, 9929: 1137, 9930: 40887, 9931: 495, 9933: 297, 9935: 693, 91857: 1375, 9938: 198, 9939: 345, 91861: 2079, 9942: 99, 9943: 4058, 9944: 99, 9946: 80090, 9947: 297, 9948: 4257, 9949: 2376, 9950: 8414, 9952: 4257, 9953: 4851, 9956: 9404, 9960: 4999, 9961: 1287, 9963: 20493, 9964: 7920, 9965: 1089, 9966: 198, 9967: 148, 26352: 99, 9969: 16928, 9970: 4059, 9973: 15939, 9974: 17919, 9975: 12375, 9978: 4504, 9979: 7821, 9980: 10642, 9981: 99, 9983: 17523, 9987: 198, 9988: 246, 9989: 99, 9991: 4158, 9994: 3069, 9996: 1385, 9998: 841, 10000: 692, 10001: 13860, 91922: 14058, 10003: 6435, 10005: 1980, 10006: 8810, 10007: 8118, 10008: 99, 10009: 396, 10010: 3661, 10012: 99, 10013: 544, 10014: 2079, 10015: 4059, 10019: 8363, 10021: 2079, 10022: 10976, 10024: 86512, 10025: 1980, 10026: 3960, 27612: 37917, 10029: 1089, 10032: 296, 10034: 3960, 10035: 36135, 91958: 148, 10039: 99, 10040: 99, 10041: 2079, 10042: 18414, 10043: 4158, 10045: 6039, 10046: 2970, 10047: 2276, 10048: 49, 10049: 3563, 91970: 99, 10051: 27264, 10053: 4455, 10054: 5940, 10055: 2425, 10056: 7920, 10057: 28164, 10059: 5643, 10062: 1287, 10063: 2573, 10064: 14256, 10067: 99, 10068: 14205, 10069: 1485, 10070: 99, 20793: 13769, 10072: 9900, 10073: 2376, 10076: 6237, 10077: 2079, 10078: 6039, 10079: 12077, 10080: 6138, 10081: 99, 10082: 2166, 10084: 22275, 10086: 99, 42855: 99, 10088: 13860, 10089: 99, 10090: 22077, 10092: 2376, 10093: 4158, 10094: 445, 10096: 198, 10097: 297, 10099: 297, 10100: 2969, 10101: 17619, 10102: 6039, 10103: 1483, 10104: 99, 10106: 99, 10107: 297, 10108: 2128, 10109: 297, 10111: 9999, 10112: 6187, 10113: 8167, 10115: 1980, 10116: 3465, 10117: 5007, 10119: 980, 10120: 99, 10122: 198, 10125: 643, 45379: 9900, 10134: 148, 10135: 16731, 10137: 99, 10139: 2079, 10141: 742, 10143: 297, 10144: 4257, 10145: 1683, 10146: 99, 10147: 3267, 10148: 1434, 10149: 8019, 10150: 198, 10151: 99, 10152: 297, 10153: 99, 10155: 840, 10156: 3663, 10157: 49, 10159: 99, 10160: 297, 10161: 148, 10163: 198, 10166: 99, 10169: 4702, 92090: 1980, 10171: 891, 10173: 2574, 10174: 99, 10175: 8118, 10176: 20986, 10178: 8880, 10182: 396, 10184: 8019, 10186: 396, 92108: 99, 92109: 148, 10190: 99, 10191: 8019, 10192: 8019, 10194: 2425, 10195: 247, 10196: 7029, 10197: 3960, 10199: 2722, 10200: 198, 10201: 346, 10202: 980, 10203: 16929, 10206: 19899, 10207: 792, 10208: 2277, 10209: 297, 10210: 1287, 10211: 2079, 10212: 4059, 10214: 24699, 10215: 7019, 10219: 148, 10220: 99, 12626: 1582, 10222: 2178, 10223: 99, 10224: 99, 10225: 297, 10226: 4603, 10227: 16087, 10228: 12721, 10233: 3960, 10235: 2079, 10237: 2128, 10238: 37818, 10240: 741, 10241: 3364, 10242: 14978, 92163: 99, 10244: 99, 25437: 49, 10248: 14454, 10249: 15719, 10250: 5989, 10251: 1483, 10252: 396, 10253: 3960, 43024: 2178, 10257: 247, 10259: 345, 10260: 297, 10261: 4257, 10263: 17919, 10266: 2276, 10267: 1980, 10268: 2871, 10270: 2128, 10274: 198, 10277: 6336, 10278: 6336, 10279: 495, 10280: 52370, 10281: 4059, 10283: 445, 10285: 198, 10286: 2178, 10287: 13464, 10288: 1386, 10289: 2326, 10290: 4257, 10292: 4355, 92213: 49, 10295: 4059, 10297: 297, 10301: 10098, 10302: 2960, 10303: 1980, 10305: 4554, 10307: 4455, 10310: 9900, 10311: 594, 10317: 6039, 10319: 6633, 10320: 247, 10321: 940, 10323: 5197, 10324: 2821, 10325: 4257, 10326: 4059, 10328: 10593, 10329: 3663, 10335: 3157, 25440: 148, 10339: 99, 10340: 4752, 10342: 8118, 10344: 2079, 10345: 14057, 10347: 297, 26732: 2178, 10349: 99, 10354: 4356, 10355: 1582, 10357: 5940, 10358: 396, 10360: 99, 10361: 792, 10362: 6039, 10363: 18809, 10364: 99, 10365: 28413, 92286: 99, 26752: 343, 10369: 29848, 75906: 396, 86379: 495, 10372: 2178, 10373: 99, 10375: 15442, 10377: 2475, 10378: 198, 10381: 2227, 10383: 2574, 10384: 4455, 92305: 99, 10386: 395, 92307: 9900, 10388: 4653, 10392: 99, 10393: 445, 10398: 48708, 10399: 297, 10400: 1681, 10403: 247, 10404: 3563, 10405: 297, 10407: 23958, 10410: 2079, 43180: 99, 10414: 198, 10416: 198, 10418: 99, 10420: 742, 10423: 99, 10424: 2178, 10426: 2772, 92348: 28116, 91850: 7595, 10430: 3960, 10435: 99, 10436: 99, 10437: 198, 10438: 99, 10439: 198, 10441: 99, 10442: 32965, 10443: 198, 10445: 792, 10447: 8118, 10449: 198, 92372: 22007, 10454: 1980, 92375: 296, 10456: 8019, 92378: 4356, 10460: 294, 10462: 1980, 10464: 99, 10465: 8514, 10467: 8316, 10468: 2673, 10469: 396, 10470: 445, 10475: 198, 10477: 99, 10479: 297, 10480: 6533, 10481: 19948, 10486: 837, 92407: 99, 10488: 45925, 10489: 1980, 10490: 445, 76027: 198, 10492: 2178, 90913: 4554, 10495: 1039, 10496: 10147, 10497: 3762, 26882: 297, 10500: 2474, 10506: 99, 10507: 10839, 10508: 99, 10509: 99, 10515: 4949, 10516: 346, 10517: 247, 10520: 6879, 10521: 4059, 10523: 5049, 10524: 10048, 43293: 99, 10526: 2376, 10529: 148, 10531: 396, 10532: 3960, 10533: 989, 10535: 49, 10536: 693, 10537: 4158, 92461: 14156, 10542: 19837, 10544: 1980, 10546: 494, 10548: 297, 10552: 10296, 10554: 99, 10555: 10147, 10556: 396, 10557: 6039, 10561: 6435, 91873: 99, 10568: 297, 10569: 495, 10570: 495, 10571: 2079, 10572: 396, 10575: 297, 10576: 3069, 10577: 41796, 10578: 792, 10579: 3069, 10580: 8316, 92501: 297, 10582: 1980, 10583: 198, 10584: 247, 10585: 643, 10587: 6920, 10588: 2178, 10589: 594, 10590: 2871, 10591: 6237, 86416: 4158, 10596: 198, 10597: 2871, 10598: 2574, 10599: 297, 10601: 891, 10602: 1980, 10604: 2079, 10605: 495, 10607: 8019, 10608: 495, 10612: 2079, 10613: 10098, 10614: 197, 10615: 15988, 10616: 3108, 10617: 99, 10618: 693, 10619: 198, 10624: 594, 23616: 197, 10626: 99, 91883: 99, 10629: 4059, 10631: 2078, 10634: 99, 10635: 198, 23364: 9849, 59789: 16731, 10638: 9999, 10639: 2277, 10641: 148, 10644: 396, 10645: 99, 92566: 6195, 10648: 8909, 10649: 148, 10650: 99, 10652: 99, 10653: 891, 10654: 4851, 10655: 4405, 10656: 6138, 10657: 495, 10659: 444, 10661: 1980, 10663: 396, 10667: 989, 10668: 8613, 10670: 3960, 20893: 12789, 10672: 495, 90919: 148, 10674: 28848, 92595: 247, 59829: 99, 10678: 297, 92600: 1029, 27066: 99, 10684: 12078, 92607: 197, 43457: 198, 10690: 2079, 10694: 4257, 10698: 2079, 10700: 16929, 10703: 891, 10707: 21582, 23630: 6365, 10710: 148, 86436: 3940, 10714: 99, 10715: 2178, 10718: 5098, 10720: 492, 27108: 198, 10727: 297, 10730: 99, 10734: 297, 10735: 2178, 10737: 2277, 10741: 49, 20905: 196, 10750: 6880, 10751: 494, 10753: 594, 10755: 247, 10758: 198, 10766: 99, 86445: 99, 10768: 6138, 10770: 197, 10771: 2376, 10772: 18858, 10779: 3861, 10784: 99, 59937: 99, 10786: 297, 10787: 8415, 10788: 148, 10792: 594, 10793: 7425, 10795: 6039, 10796: 8316, 10798: 544, 10800: 148, 10801: 1681, 10802: 99, 10803: 99, 92725: 2871, 10807: 495, 10808: 990, 10811: 198, 10815: 99, 10816: 3267, 53646: 99, 10823: 1980, 10826: 593, 76364: 99, 10830: 21978, 92751: 4059, 10832: 297, 10833: 297, 10834: 693, 91470: 2079, 10837: 3465, 92758: 99, 10840: 99, 76377: 692, 10846: 2376, 10849: 19135, 10851: 445, 10855: 5641, 27241: 98, 10858: 297, 10859: 148, 10861: 99, 10863: 99, 10866: 297, 92787: 2227, 10870: 8514, 10874: 297, 10876: 148, 10877: 2227, 10878: 4356, 10880: 8118, 43650: 6483, 92806: 99, 10887: 198, 92809: 49, 10891: 198, 10893: 35937, 10894: 693, 10895: 198, 45507: 49, 92821: 22076, 10903: 99, 10904: 2079, 10909: 2376, 10912: 495, 26396: 99, 92564: 1980, 92848: 11979, 92850: 1980, 10932: 99, 102857: 245, 10938: 15246, 10939: 693, 10940: 297, 10942: 198, 92866: 99, 10947: 297, 10948: 1088, 10949: 891, 10950: 445, 70092: 2079, 10954: 18117, 10956: 396, 86564: 148, 25496: 539, 10964: 297, 27352: 99, 10971: 1485, 10973: 99, 20944: 4900, 10979: 198, 10981: 99, 10983: 99, 20946: 49, 10994: 2178, 10995: 9840, 10996: 247, 25462: 99, 11000: 2178, 11001: 643, 11002: 99, 11004: 495, 11006: 2079, 11007: 495, 11009: 495, 60162: 8316, 11011: 99, 11013: 7474, 11017: 198, 11018: 297, 11019: 197, 11021: 9007, 11023: 198, 11030: 2376, 20953: 3643, 11034: 594, 11037: 6484, 11038: 21669, 76575: 2277, 11041: 99, 11042: 297, 11043: 2326, 11044: 297, 11052: 99, 11055: 99, 11057: 494, 11059: 693, 11060: 8067, 11062: 148, 11066: 297, 11073: 99, 11074: 4851, 11075: 99, 10038: 741, 11078: 4454, 43847: 99, 11080: 99, 91959: 49, 11085: 6039, 93006: 198, 11087: 297, 91960: 99, 11090: 742, 86499: 6039, 11093: 10593, 20964: 9996, 11099: 14058, 11101: 2475, 27490: 99, 43878: 2079, 11115: 297, 11119: 1039, 78313: 99, 11130: 99, 11131: 396, 11135: 148, 11136: 297, 11139: 99, 11141: 148, 76682: 99, 11148: 198, 11150: 544, 93071: 345, 11152: 4356, 11155: 297, 76697: 49, 11166: 99, 11167: 1138, 43939: 98, 27556: 99, 11174: 15939, 11177: 2376, 11178: 8077, 11180: 594, 60335: 2376, 39065: 6039, 11187: 2326, 11193: 495, 11194: 4207, 11197: 297, 11198: 247, 11199: 297, 11203: 99, 27588: 49, 11208: 791, 43977: 1980, 93131: 3663, 11212: 4059, 43985: 49, 11220: 198, 11224: 2079, 11228: 297, 11231: 346, 11238: 3663, 44010: 297, 11245: 297, 11248: 198, 11249: 197, 11250: 197, 11251: 1089, 76793: 99, 93180: 6435, 11263: 445, 86528: 1980, 11266: 11532, 11268: 297, 11269: 11088, 11272: 396, 11273: 99, 11274: 693, 93196: 99, 11278: 1980, 11285: 10740, 11293: 198, 11294: 841, 11295: 891, 11296: 198, 11297: 99, 11299: 396, 93221: 49, 11306: 2772, 20999: 4483, 11308: 495, 93229: 1029, 11311: 297, 11312: 495, 11313: 1980, 85401: 99, 11317: 4356, 11319: 297, 11320: 12770, 26463: 8807, 27708: 99, 11325: 99, 11328: 8316, 42848: 98, 11330: 396, 11332: 396, 44106: 494, 11339: 297, 11341: 99, 27726: 6831, 11346: 99, 11350: 50893, 11354: 990, 93276: 2079, 36022: 9999, 27745: 99, 11362: 297, 11363: 495, 11375: 50489, 11378: 99, 93300: 99, 11381: 6039, 11382: 297, 11385: 18711, 11388: 99, 11393: 495, 11394: 246, 86033: 99, 11397: 198, 11402: 396, 11403: 99, 11404: 4257, 11406: 198, 11409: 792, 11417: 495, 93340: 39649, 11421: 2227, 11422: 495, 11426: 6435, 11429: 198, 11433: 99, 11434: 396, 93359: 99, 11440: 4059, 11442: 2326, 11443: 693, 27828: 198, 11445: 198, 11448: 49, 11455: 495, 11456: 445, 11458: 1980, 93379: 2079, 93380: 495, 93382: 148, 53793: 99, 11464: 346, 75640: 49, 11479: 693, 11480: 99, 21028: 839, 11482: 99, 60635: 247, 19477: 2128, 11491: 495, 11492: 346, 11497: 297, 11498: 346, 75645: 98, 11504: 6385, 11507: 4257, 11511: 297, 11513: 396, 19478: 6909, 11516: 396, 11519: 99, 27906: 198, 93443: 49, 11524: 643, 60677: 3960, 11532: 99, 11535: 31779, 11536: 495, 11537: 10395, 11539: 99, 60634: 49, 11542: 792, 27929: 99, 11549: 495, 11550: 643, 11551: 99, 23771: 148, 20025: 3087, 11558: 99, 11561: 1178, 11563: 198, 60718: 3731, 11568: 297, 11569: 8019, 20571: 2989, 21043: 98, 11572: 297, 11578: 247, 11581: 198, 11585: 346, 44354: 99, 11588: 23760, 74779: 99, 11591: 198, 11592: 198, 11594: 1485, 11595: 99, 44364: 49, 11597: 891, 11599: 445, 11601: 99, 11603: 495, 11604: 148, 93525: 1980, 11608: 12127, 11611: 495, 93533: 49, 11615: 9009, 93537: 49, 11619: 297, 93544: 99, 11626: 19899, 93548: 49, 11632: 198, 11633: 2277, 11635: 642, 11637: 891, 11638: 693, 11641: 148, 53823: 49, 11647: 4158, 11655: 99, 92055: 198, 93580: 297, 11661: 495, 93582: 343, 77199: 99, 11666: 544, 11667: 2376, 11668: 198, 22211: 147, 11675: 247, 11678: 6237, 28066: 99, 11683: 198, 11684: 297, 11685: 49, 11686: 396, 60845: 2375, 21064: 99, 93629: 98, 11710: 99, 92045: 7029, 11724: 495, 11725: 99, 11732: 990, 28124: 99, 11741: 693, 93665: 445, 11747: 396, 11748: 99, 11752: 693, 93675: 99, 11759: 99, 11760: 396, 28146: 99, 86611: 99, 11771: 297, 93693: 15840, 11775: 99, 11777: 891, 11793: 396, 11798: 99, 11801: 2079, 11808: 990, 11811: 12276, 11813: 99, 11817: 4257, 11818: 346, 11820: 594, 11823: 693, 11825: 22077, 28210: 2079, 11827: 1485, 11833: 7326, 11835: 99, 11836: 2871, 11838: 297, 11845: 99, 11846: 99, 11851: 3267, 61004: 2376, 11858: 99, 11859: 297, 93784: 99, 11868: 99, 93798: 2079, 93799: 3207, 11886: 297, 44658: 2079, 11892: 1980, 93813: 99, 46769: 6237, 11899: 2079, 11900: 198, 11901: 1632, 44670: 2178, 72981: 8514, 11909: 297, 93830: 49, 93836: 99, 11917: 99, 11918: 495, 11920: 22888, 11923: 99, 11925: 297, 11928: 15939, 11932: 99, 11936: 99, 11938: 495, 93860: 15630, 11941: 99, 11945: 99, 11948: 6039, 11949: 297, 11951: 2079, 93872: 49, 11953: 693, 11954: 99, 11955: 99, 28340: 49, 11957: 18166, 11961: 198, 11962: 297, 87143: 2622, 90962: 8959, 61117: 6930, 77502: 396, 21109: 2029, 93888: 1980, 11969: 4653, 11970: 21481, 11978: 77836, 11979: 198, 11985: 692, 11987: 4059, 11992: 99, 93914: 20047, 44765: 49, 12004: 49, 12007: 2079, 12013: 99, 12015: 396, 12016: 2425, 12018: 1089, 19416: 1225, 93942: 99, 12043: 99, 12044: 2277, 12050: 99, 12051: 198, 61206: 1980, 12056: 17919, 12059: 99, 12064: 4653, 93990: 5088, 12074: 396, 93998: 148, 12079: 297, 21128: 198, 12087: 99, 12096: 198, 94019: 346, 12108: 5148, 12117: 99, 12122: 99, 12127: 4158, 36411: 495, 12131: 99, 12135: 99, 12138: 4158, 61294: 1980, 12145: 9999, 12152: 198, 12156: 99, 12158: 396, 12160: 494, 94081: 99, 61316: 148, 12166: 891, 77704: 16532, 94090: 99, 21143: 1969, 23874: 99, 61334: 49, 12186: 2376, 29338: 148, 12192: 6039, 28577: 22077, 12196: 99, 21148: 1029, 12203: 99, 12205: 29799, 22229: 12838, 12216: 99, 12217: 99, 12222: 198, 21684: 9520, 12234: 14157, 12240: 297, 12248: 643, 61403: 2277, 94180: 148, 12261: 6028, 94182: 99, 12264: 148, 94189: 99, 12274: 2673, 12275: 6138, 83967: 4257, 12289: 198, 12291: 297, 12299: 13215, 75778: 99, 12305: 4108, 28690: 99, 12308: 4455, 12311: 2277, 12315: 99, 12320: 99, 94243: 3353, 94245: 197, 12330: 693, 12336: 49, 12337: 8217, 12338: 99, 12339: 49, 12341: 148, 26633: 2277, 12346: 594, 12348: 99, 94269: 49, 12353: 198, 12354: 40441, 28741: 2475, 45126: 297, 12361: 99, 94292: 2128, 90976: 4158, 12386: 99, 12388: 197, 94318: 99, 12399: 1980, 12401: 495, 12404: 8415, 23913: 49, 12409: 593, 12414: 8414, 12415: 198, 94350: 49, 12431: 396, 12432: 22175, 12433: 10197, 12437: 99, 75702: 198, 77978: 12177, 12444: 841, 45213: 4653, 12446: 297, 12452: 99, 12456: 3069, 12457: 2178, 86727: 4059, 12463: 148, 12464: 198, 78004: 693, 12470: 1980, 12471: 99, 12476: 99, 12483: 396, 12488: 2177, 12489: 495, 12491: 891, 12492: 99, 79735: 7365, 12496: 297, 12498: 297, 12501: 99, 90980: 13959, 12507: 198, 12509: 23796, 59222: 10246, 12512: 99, 28897: 4751, 24967: 3801, 78060: 2079, 12525: 297, 94447: 196, 75816: 297, 12532: 6138, 12533: 693, 12540: 247, 12545: 495, 12549: 445, 12551: 2079, 12552: 4059, 12554: 4504, 12555: 4257, 12556: 99, 12558: 49, 21209: 5285, 94489: 1128, 94490: 49, 12572: 693, 78119: 4653, 12585: 99, 94506: 49, 12589: 396, 12591: 99, 19514: 10829, 12603: 1980, 45378: 99, 94531: 49, 28999: 1980, 94536: 4157, 12619: 297, 12620: 99, 94546: 1029, 12637: 198, 12641: 99, 12649: 495, 94570: 197, 55404: 495, 12656: 8019, 92222: 21978, 61815: 99, 29056: 5940, 37611: 99, 86765: 2673, 12690: 49, 12691: 99, 12696: 6039, 87713: 7177, 12706: 2277, 94627: 49, 85083: 99, 94644: 13057, 12729: 99, 12730: 1582, 94652: 495, 12733: 1980, 12734: 99, 12739: 4653, 12742: 2376, 12743: 49995, 12747: 4158, 12749: 594, 94670: 99, 12755: 693, 12759: 2720, 12766: 198, 94687: 2058, 12771: 693, 61924: 2277, 12777: 99, 12778: 346, 94706: 1029, 12788: 297, 12792: 395, 12798: 544, 94724: 99, 61960: 99, 94731: 1980, 86786: 495, 12820: 1582, 61973: 594, 45596: 1980, 12833: 4158, 12841: 3861, 12842: 99, 94770: 197, 12856: 4059, 12858: 99, 12863: 99, 12865: 2524, 23991: 198, 94805: 3157, 12888: 99, 94809: 148, 12892: 4108, 12898: 6237, 94822: 33759, 12931: 1187, 12932: 99, 12935: 4306, 94866: 297, 24003: 1524, 12949: 198, 94874: 99, 12956: 495, 21274: 4108, 94882: 49, 12967: 8217, 23891: 49, 18546: 196, 94895: 1029, 84083: 1980, 9707: 99, 12993: 1980, 94915: 49, 94916: 99, 94922: 49, 13007: 495, 94947: 1127, 13029: 9207, 13033: 346, 18557: 48360, 45812: 20592, 94969: 98, 94973: 197, 13057: 99, 94980: 2128, 13067: 99, 94988: 98, 13081: 247, 13082: 12474, 86613: 2178, 18566: 889, 29480: 99, 95017: 98, 13101: 99, 20622: 7448, 13107: 2079, 13118: 99, 13120: 8019, 13123: 99, 13127: 99, 18572: 9800, 62291: 3257, 13142: 99, 13144: 99, 78999: 8613, 21307: 148, 13156: 1029, 67731: 6682, 13175: 99, 13177: 99, 45949: 495, 13183: 297, 13187: 2721, 84118: 99, 13208: 99, 13215: 445, 18587: 1078, 23354: 343, 45996: 99, 95150: 99, 95153: 99, 13236: 49, 29623: 2277, 13248: 49, 13249: 4257, 13258: 2128, 13259: 297, 13262: 297, 13268: 99, 13280: 21382, 13285: 197, 13289: 2652, 13291: 99, 29676: 49, 13296: 148, 35659: 49, 13321: 99, 95244: 99, 95257: 1980, 84143: 99, 95260: 148, 86874: 2227, 13357: 99, 13358: 198, 13359: 495, 97800: 343, 13364: 2178, 84148: 2079, 86879: 99, 29759: 99, 13376: 297, 13377: 9900, 13378: 2475, 13383: 2376, 95313: 99, 21347: 1029, 13400: 297, 54116: 99, 20749: 49, 13416: 346, 95337: 693, 13420: 297, 13425: 198, 100542: 2989, 13431: 297, 13434: 99, 13436: 297, 84160: 99, 95370: 98, 13462: 3267, 13463: 2623, 13467: 1227, 79004: 3465, 13470: 99, 13472: 2474, 13473: 396, 13474: 2256, 78707: 99, 13500: 791, 70519: 49, 13516: 98, 84173: 99, 13524: 3960, 18638: 7106, 13528: 198, 13532: 198, 95458: 49, 20091: 2009, 13542: 6138, 13544: 693, 13557: 297, 18646: 9016, 25548: 99, 55192: 99, 95506: 49, 13599: 99, 44098: 2376, 13609: 6435, 13614: 99, 13625: 495, 58829: 49, 13627: 4207, 92385: 6039, 13641: 2475, 53374: 99, 85017: 148, 13654: 297, 24121: 980, 13663: 99, 13665: 940, 95593: 1177, 13676: 345, 13679: 4059, 18664: 12299, 18829: 7889, 13689: 198, 13690: 297, 13693: 99, 92395: 4059, 25552: 1326, 62856: 2376, 79247: 9900, 30098: 99, 95635: 49, 13727: 495, 13728: 99, 13729: 495, 13732: 792, 79270: 99, 30122: 99, 13742: 99, 13744: 297, 13745: 99, 13755: 198, 95676: 245, 62909: 99, 95679: 98, 79299: 99, 13771: 99, 13772: 99, 30157: 49, 92409: 4207, 13788: 148, 18682: 3038, 62945: 495, 13795: 148, 79333: 49, 13802: 17126, 95723: 98, 20815: 2673, 35069: 99, 13810: 297, 97875: 147, 30197: 2277, 30198: 49, 25556: 10340, 95749: 1029, 13835: 297, 62995: 5395, 30230: 297, 46619: 99, 30236: 8118, 79389: 99, 30239: 4455, 13871: 99, 13875: 297, 13882: 2079, 13884: 594, 13886: 4257, 13891: 99, 13906: 2079, 13908: 297, 13909: 99, 46685: 99, 13918: 198, 18704: 543, 13927: 99, 95848: 99, 95850: 99, 30330: 297, 13948: 1526, 22287: 7399, 13955: 99, 13957: 297, 30345: 4257, 18711: 6138, 13964: 495, 30349: 49, 13969: 23859, 46739: 544, 13976: 198, 95900: 98, 95904: 8918, 13985: 99, 79522: 1078, 95909: 99, 13990: 99, 48754: 49, 95918: 4059, 25562: 99, 14001: 297, 14002: 99, 14006: 594, 14008: 4059, 14016: 297, 30401: 198, 79557: 445, 14022: 297, 14028: 988, 95949: 2475, 14031: 148, 14035: 99, 14039: 297, 14044: 297, 95965: 148, 21199: 2303, 95979: 295, 19563: 6704, 26921: 2079, 63227: 296, 14076: 99, 30463: 9702, 14085: 99, 96013: 99, 96014: 2177, 10541: 198, 91033: 4207, 86123: 49, 79637: 99, 96031: 98, 14118: 1980, 96045: 742, 96051: 2128, 96053: 1980, 76084: 17919, 30522: 198, 14154: 99, 87010: 2079, 96078: 2009, 85034: 99, 79715: 99, 96101: 98, 92476: 49, 92477: 49, 14193: 247, 78825: 2128, 46975: 297, 14208: 693, 14212: 297, 96133: 7056, 14214: 99, 14215: 99, 87021: 2871, 79772: 99, 14238: 4207, 14239: 594, 87765: 2178, 96187: 98, 14268: 198, 14270: 148, 96201: 441, 87033: 148, 92500: 99, 14335: 495, 96258: 2989, 96260: 98, 14342: 197, 79879: 99, 18776: 1029, 14366: 99, 79904: 99, 22301: 4920, 96294: 49, 18781: 3186, 14392: 1980, 25453: 2079, 14406: 99, 14412: 10196, 14413: 14799, 96334: 98, 14418: 1188, 93226: 12127, 23555: 196, 14427: 4455, 14429: 99, 14430: 198, 96351: 1029, 84326: 396, 92519: 49, 14446: 4356, 24253: 7988, 14453: 99, 63606: 99, 47228: 198, 14470: 247, 96397: 49, 87064: 198, 14482: 297, 20668: 29252, 14487: 197, 14489: 99, 14490: 99, 63643: 99, 96420: 98, 54300: 49, 103453: 1029, 54494: 3960, 14520: 9999, 80068: 198, 97996: 49, 96460: 49, 96468: 98, 14557: 4158, 14560: 1980, 96486: 147, 14573: 297, 18813: 21420, 14578: 3960, 96506: 1029, 18815: 15680, 14592: 148, 30983: 49, 46127: 9999, 96544: 196, 37938: 49, 47413: 297, 14647: 99, 96579: 98, 14661: 99, 10637: 99, 14676: 1980, 96602: 1078, 80231: 49, 23403: 294, 14704: 99, 63857: 98, 96626: 49, 14709: 495, 96630: 1029, 51604: 396, 88697: 99, 14721: 49, 18838: 346, 14730: 2673, 96651: 49, 92568: 16038, 47508: 99, 96666: 2009, 40687: 2009, 92570: 198, 24496: 2079, 24305: 245, 80299: 99, 80301: 297, 80312: 99, 18847: 13965, 14784: 2475, 92576: 693, 21581: 1029, 80337: 49, 14804: 99, 96739: 49, 47599: 99, 96752: 98, 14833: 148, 21587: 1078, 96756: 98, 14845: 148, 92588: 98, 14859: 148, 18861: 65219, 14864: 99, 43438: 10929, 21593: 8019, 14872: 198, 14877: 495, 14878: 247, 84400: 594, 25592: 7821, 18868: 2079, 47678: 49, 96850: 49, 24334: 49, 14955: 198, 96877: 147, 14965: 396, 96891: 49, 47742: 99, 88880: 99, 96899: 98, 14984: 49, 85607: 148, 96913: 2989, 23413: 8068, 64159: 99, 80545: 594, 15016: 9999, 15019: 99, 74697: 544, 21619: 49, 84425: 297, 18893: 18134, 15057: 99, 15076: 643, 85610: 495, 18897: 17374, 18898: 11224, 84435: 99, 80633: 198, 27924: 4752, 31485: 643, 97024: 49, 97057: 49, 15144: 495, 97073: 49, 88886: 43708, 64309: 49, 15164: 99, 15172: 27063, 15173: 11979, 15177: 99, 15178: 2376, 15181: 99, 47952: 99, 15185: 99, 15197: 2673, 84454: 297, 15214: 594, 15215: 99, 15216: 99, 15218: 198, 15223: 49, 57151: 99, 97163: 98, 24387: 99, 48023: 99, 101983: 49, 89927: 2079, 31663: 49, 97201: 49, 27787: 49, 97209: 98, 97212: 49, 46241: 99, 18936: 9849, 31702: 99, 80858: 14157, 18942: 13959, 84479: 3960, 19606: 7939, 97285: 49, 97288: 245, 31753: 395, 31755: 49, 15381: 297, 84484: 1980, 15404: 99, 48177: 4059, 48185: 99, 87065: 297, 31812: 99, 103606: 245, 15430: 99, 76893: 297, 52343: 99, 15476: 99, 87232: 6385, 84503: 8613, 93262: 99, 97422: 98, 15504: 23839, 87261: 148, 82351: 49, 15511: 1683, 31901: 99, 97450: 10878, 15532: 99, 97457: 49, 15539: 99, 31939: 1980, 81092: 99, 100898: 17689, 97486: 98, 31951: 494, 84515: 8215, 15585: 445, 18982: 46893, 97513: 1029, 15595: 297, 84521: 4603, 97535: 98, 15616: 198, 97541: 49, 32012: 3960, 32017: 297, 22343: 99, 15635: 297, 15642: 2277, 15646: 19998, 97567: 147, 20707: 392, 48431: 297, 12524: 99, 10805: 1335, 87264: 4851, 86303: 99, 15691: 99, 22145: 2960, 15699: 4603, 19001: 1029, 97635: 49, 64869: 198, 84540: 2079, 32122: 4356, 36330: 198, 48523: 396, 64925: 99, 64928: 2326, 15806: 99, 32194: 198, 97737: 49, 15818: 99, 15819: 247, 64976: 2227, 97747: 1078, 97749: 441, 15840: 99, 15848: 35739, 92754: 99, 48628: 396, 97785: 49, 97788: 1078, 19623: 18133, 65025: 17928, 97795: 7938, 15880: 148, 21762: 4741, 94366: 99, 65454: 49, 97814: 49, 97819: 245, 86183: 2673, 15905: 297, 15907: 297, 15916: 99, 98228: 1960, 15954: 25839, 15955: 99, 48724: 693, 15958: 445, 48737: 98, 97892: 392, 48743: 99, 84583: 7920, 15983: 5978, 20718: 3989, 97906: 49, 97908: 49, 24510: 1029, 68728: 99, 58045: 99, 16004: 99, 19052: 15706, 43424: 99, 24514: 891, 97936: 245, 85096: 49, 87824: 2574, 16038: 99, 84593: 2079, 32428: 296, 91098: 395, 16049: 1079, 16051: 10395, 16057: 2107, 16062: 99, 83796: 495, 81612: 2376, 32462: 148, 26177: 2475, 98009: 49, 16093: 99, 81638: 99, 32497: 99, 85099: 297, 98036: 49, 84606: 99, 48887: 99, 16122: 99, 98046: 49, 25633: 5544, 20723: 196, 95534: 49, 32535: 49, 98074: 49, 86351: 198, 32559: 122958, 98097: 196, 32562: 297, 98103: 4998, 84619: 2871, 22362: 49, 16216: 99, 38184: 2078, 30011: 34155, 98154: 2940, 98162: 49, 98173: 2058, 16269: 99, 49052: 18414, 81829: 99, 91652: 49, 16304: 297, 16308: 16383, 98254: 49, 81871: 2079, 49111: 99, 81882: 99, 16350: 99, 16361: 297, 16374: 297}

		content1 := "在收集、整理了各位主公们的建议与述求后，我们决定将招募玩法中的SP武将卡掉落概率调整为5%，为保证公平与良好的游戏体验，我们将对参与过招募玩法的老玩家进行补偿！\n\n" +
			"补偿方案：5月18日维护前招募消耗的宝玉*50%\n\n" +
			"补偿宝玉："
		content2 := "\n\n" +
			"再次感谢您的热爱与支持"
		for uid, jade := range uid2jade {
			rewardJade := jade / 2
			glog.Infof("rewardRecruit uid=%d, jade=%d", uid, jade)
			sender := module.Mail.NewMailSender(uid)
			sender.SetTitleAndContent("招募SP武将概率调整", content1+strconv.Itoa(rewardJade)+content2)
			reward := sender.GetRewardObj()
			reward.AddAmountByType(pb.MailRewardType_MrtJade, rewardJade)
			sender.Send()
		}
	})
}
