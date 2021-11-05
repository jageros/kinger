package web

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	gutils "kinger/gopuppy/common/utils"
	"io"
	"io/ioutil"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/proto/pb"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	success, _ = json.Marshal(map[string]interface{} {
		"errcode": 0,
	})

	httpMethodErr, _ = json.Marshal(map[string]interface{} {
		"errcode": 100,
	})

	signErr, _ = json.Marshal(map[string]interface{} {
		"errcode": 1,
	})

	noPlayerErr, _ = json.Marshal(map[string]interface{} {
		"errcode": 100,
	})

	itemList []byte
	cardList []byte
	areaList []byte
	cardID2CardName map[uint32]string
)

func getCardList() ([]byte, map[uint32]string) {
	if cardList != nil && len(cardList) > 0 {
		return cardList, cardID2CardName
	}

	c := make(chan struct{})
	evq.CallLater(func() {
		cardID2CardName = map[uint32]string{}
		poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
		cardDatas := poolGameData.GetAllCards()
		var cards []map[string]interface{}
		for _, cardData := range cardDatas {
			if cardData.CardID == 0 || cardData.Campaign == 1 {
				continue
			}
			if _, ok := cardID2CardName[cardData.CardID]; ok {
				continue
			}

			cardData2 := poolGameData.GetCard(cardData.CardID, 1)
			if cardData2 == nil {
				continue
			}

			cardID2CardName[cardData.CardID] = cardData.GetName()
			cards = append(cards, map[string]interface{}{
				"name": cardData2.GetName(),
				"id": cardData2.CardID,
			})
		}

		cardList, _ = json.Marshal(cards)
		close(c)
	})

	<- c
	return cardList, cardID2CardName
}

func wrapHttpHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		gutils.CatchPanic(func() {
			handler(writer, request)
		})
	}
}

func prepareRequest(writer http.ResponseWriter, request *http.Request) map[string]interface{} {
	if request.Method == "GET" {
		writer.Write(httpMethodErr)
		return nil
	}

	args := map[string]interface{}{}
	payload , err := ioutil.ReadAll(request.Body)
	if err != nil {
		errResponse, _ := json.Marshal(map[string]interface{}{
			"errcode": 3,
			"errMsg": err.Error(),
		})
		writer.Write(errResponse)
		return nil
	}

	err = json.Unmarshal(payload, &args)
	if err != nil {
		errResponse, _ := json.Marshal(map[string]interface{}{
			"errcode": 3,
			"errMsg": err.Error(),
		})
		writer.Write(errResponse)
		return nil
	}
	
	return args
}

func checkSign(args map[string]string, sign string) bool {
	cfg := config.GetConfig()
	if cfg.Debug {
		return true
	}
	if cfg.GmtoolKey == "" {
		return true
	}

	var keys []string
	for k, _ := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sourceBuilder := &strings.Builder{}
	for _, k := range keys {
		sourceBuilder.WriteString(fmt.Sprintf("%s=%s&", k, args[k]))
	}
	sourceBuilder.WriteString("secretKey=")
	sourceBuilder.WriteString(cfg.GmtoolKey)

	h := md5.New()
	io.WriteString(h, sourceBuilder.String())
	mysign := fmt.Sprintf("%x", h.Sum(nil))
	return mysign == sign
}

func playerInfoHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	uid := common.UUid(args["uid"].(float64))
       var playerAttr *attribute.AttrMgr
       for _, region := range []uint32{1, 2} {
               playerAttr = attribute.NewAttrMgr("player", uid, false, region)
               err := playerAttr.Load(true)
               if err == nil {
                       break
               } else {
                       playerAttr = nil
               }
       }

       if playerAttr == nil {
		writer.Write(noPlayerErr)
		return
	}

	c := make(chan struct{})
	var reply map[string]interface{}
	evq.CallLater(func() {
		player := module.Player.NewPlayerByAttr(uid, playerAttr)
		winRate := int(100 * float64(player.GetFirstHandWinAmount() + player.GetBackHandWinAmount()) / float64(
			player.GetFirstHandAmount() + player.GetBackHandAmount()))

		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
		// 资源
		items := []map[string]interface{}{
			{
				"name": "金币",
				"id": "gold",
				"amount": resCpt.GetResource(consts.Gold),
			},
			{
				"name": "宝玉",
				"id": "jade",
				"amount": resCpt.GetResource(consts.Jade),
			},
			{
				"name": "声望",
				"id": "reputation",
				"amount": resCpt.GetResource(consts.Reputation),
			},
			{
				"name": "武将碎片",
				"id": "cardPiece",
				"amount": resCpt.GetResource(consts.CardPiece),
			},
			{
				"name": "皮肤碎片",
				"id": "skinPiece",
				"amount": resCpt.GetResource(consts.SkinPiece),
			},
		}

		equips := module.Bag.GetAllItemsByType(player, consts.ItEquip)
		cardSkins := module.Bag.GetAllItemsByType(player, consts.ItCardSkin)
		headFrames := module.Bag.GetAllItemsByType(player, consts.ItHeadFrame)
		emojis := module.Bag.GetAllItemsByType(player, consts.ItEmoji)
		for _, its := range [][]types.IItem{equips, cardSkins, headFrames, emojis} {
			for _, it := range its {
				items = append(items, map[string]interface{}{
					"name": it.GetName(),
					"id": it.GetGmID(),
					"amount": it.GetAmount(),
				})
			}
		}

		cards := cardCpt.GetAllCollectCards()
		for _, c := range cards {
			cardData := c.GetCardGameData()
			items = append(items, map[string]interface{}{
				"name": cardData.GetName(),
				"id": fmt.Sprintf("card%d", cardData.CardID),
				"amount": c.GetAmount(),
			})
		}

		reply = map[string]interface{}{
			"errcode": 0,
			"name": player.GetName(),
			"roleLevel": player.GetPvpLevel(),
			"maxRoleLevel": player.GetMaxPvpLevel(),
			"level": player.GetComponent(consts.LevelCpt).(types.ILevelComponent).GetCurLevel(),
			"winRate": winRate,
			"isForbidLogin": player.IsForbidLogin(),
			"isForbidChat": player.IsForbidChat(),
			"lastLoginTime": player.GetLastOnlineTime(),
			"items": items,
			"accountType": player.GetLogAccountType().String(),
			"area": player.GetArea(),
			"channel": player.GetChannel(),
			"channelUid": player.GetChannelUid(),
			"ipAddr": player.GetIP(),
			"createTime": player.GetCreateTime(),
		}
		close(c)
	})

	<- c
	response, _ := json.Marshal(reply)
	writer.Write(response)
}

func forbidLoginHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	uid := common.UUid(args["uid"].(float64))
	isForbid := args["isForbid"].(bool)

	if !checkSign(map[string]string{
		"uid": uid.String(),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	c := make(chan struct{})
	evq.CallLater(func() {
		utils.PlayerMqPublish(uid, pb.RmqType_ForbidLogin, &pb.RmqForbidLogin{
			IsForbid: isForbid,
			OverTime: -1,
		})
		close(c)
	})
	<-c
	writer.Write(success)
}

func forbidChatHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	uid := common.UUid(args["uid"].(float64))
	isForbid := args["isForbid"].(bool)
	if !checkSign(map[string]string{
		"uid": uid.String(),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	c := make(chan struct{})
	evq.CallLater(func() {
		utils.PlayerMqPublish(uid, pb.RmqType_ForbidChat, &pb.RmqForbidChat{
			IsForbid: isForbid,
			OverTime: -1,
		})
		close(c)
	})
	<-c
	writer.Write(success)
}

func fetchItemListHandler(writer http.ResponseWriter, request *http.Request) {
	if len(itemList) > 0 {
		writer.Write(itemList)
		return
	}

	c := make(chan struct{})
	var reply map[string]interface{}
	evq.CallLater(func() {
		items := []map[string]interface{}{
			{
				"name": "金币",
				"id":   "gold",
			},
			{
				"name": "钻石",
				"id": "jade",
			},
			{
				"name": "功勋",
				"id": "feats",
			},
			{
				"name": "名望",
				"id": "prestige",
			},
			{
				"name": "玉石",
				"id": "bowlder",
			},
			{
				"name": "战功",
				"id": "contribution",
			},
			{
				"name": "声望",
				"id": "reputation",
			},
			{
				"name": "武将碎片",
				"id": "cardPiece",
			},
			{
				"name": "皮肤碎片",
				"id": "skinPiece",
			},
			{
				"name": "加速劵",
				"id": "accTicket",
			},
			{
				"name": "星星",
				"id": "pvpScore",
			},
			{
				"name": "匹配积分",
				"id": "matchScore",
			},
		}

		poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
		cardDatas := poolGameData.GetAllCards()
		cardIdSet := common.UInt32Set{}
		for _, cardData := range cardDatas {
			if cardData.CardID == 0 || cardData.Campaign == 1 || cardIdSet.Contains(cardData.CardID) {
				continue
			}

			cardData2 := poolGameData.GetCard(cardData.CardID, 1)
			if cardData2 == nil {
				continue
			}
			cardIdSet.Add(cardData.CardID)
			items = append(items, map[string]interface{}{
				"name": cardData2.GetName(),
				"id": fmt.Sprintf("card%d", cardData2.CardID),
			})
		}

		skinGameData := gamedata.GetGameData(consts.CardSkin).(*gamedata.CardSkinGameData)
		for _, skinData := range skinGameData.CardSkins {
			items = append(items, map[string]interface{}{
				"name": fmt.Sprintf("皮肤-%s", skinData.GetName()),
				"id": skinData.ID,
			})
		}

		headFrameGameData := gamedata.GetGameData(consts.HeadFrame).(*gamedata.HeadFrameGameData)
		for _, headFrameData := range headFrameGameData.HeadFrames {
			items = append(items, map[string]interface{}{
				"name": fmt.Sprintf("头像框-%s", headFrameData.GetName()),
				"id": fmt.Sprintf("HF%s", headFrameData.ID),
			})
		}

		equipGameData := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData)
		for _, eqData := range equipGameData.Equips {
			items = append(items, map[string]interface{}{
				"name": fmt.Sprintf("宝物-%s", eqData.GetName()),
				"id": eqData.ID,
			})
		}

		emojiGameData := gamedata.GetGameData(consts.Emoji).(*gamedata.EmojiGameData)
		for _, eData := range emojiGameData.EmojiTeams {
			items = append(items, map[string]interface{}{
				"name": fmt.Sprintf("表情-%s", eData.GetTeamName()),
				"id": fmt.Sprintf("EJ%d", eData.Team),
			})
		}

		treasureGameData := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData)
		for _, treasureData := range treasureGameData.AllTreasures {
			items = append(items, map[string]interface{}{
				"name": treasureData.GetName(),
				"id": treasureData.ID,
			})
		}

		reply = map[string]interface{}{
			"errcode": 0,
			"items": items,
		}
		close(c)
	})

	<- c
	itemList, _ = json.Marshal(reply)
	writer.Write(itemList)
}

func sendMailHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	uids := args["uid"].([]interface{})
	title := args["title"].(string)
	content := args["content"].(string)
	accountType := pb.AccountTypeEnum(args["accountType"].(float64))
	rewards := args["rewards"].([]interface{})
	var newbieDeadLine int64
	if newbieDeadLinef, ok := args["newbieDeadLine"]; ok {
		newbieDeadLine = int64(newbieDeadLinef.(float64))
	}

	var area int
	if areaf, ok := args["area"]; ok {
		area = int(areaf.(float64))
	}

	if !checkSign(map[string]string{
		"title": title,
		"content": content,
		"accountType": strconv.Itoa(int(accountType)),
		"newbieDeadLine": strconv.FormatInt(newbieDeadLine, 10),
		"area": strconv.Itoa(area),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	//glog.Infof("%v", rewards)
	c := make(chan struct{})
	var mailIDs []int
	evq.CallLater(func() {
		for _, uid2 := range uids {
			var uid common.UUid
			uid3 := uid2.(float64)
			if uid3 >= 0 {
				uid = common.UUid(uid3)
				if uid == 0 {
					continue
				}
			}

			sender := module.Mail.NewMailSender(uid)
			mailReward := sender.GetRewardObj()
			sender.SetTitleAndContent(title, content)
			sender.SetAccountType(accountType)
			sender.SetNewbieDeadLine(newbieDeadLine)
			sender.SetArea(area)
			for _, reward := range rewards {
				reward2 := reward.(map[string]interface{})
				for id, amount := range reward2 {
					amount2 := int(amount.(float64))
					if id == "gold" {
						mailReward.AddGold(amount2)
					} else if id == "jade" {
						mailReward.AddJade(amount2)
					} else if id == "feats" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtFeats, amount2)
					} else if id == "prestige" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtPrestige, amount2)
					} else if id == "bowlder" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtBowlder, amount2)
					} else if id == "contribution" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtContribution, amount2)
					} else if id == "reputation" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtReputation, amount2)
					} else if id == "cardPiece" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtCardPiece, amount2)
					} else if id == "skinPiece" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtSkinPiece, amount2)
					} else if id == "accTicket" {
						mailReward.AddAmountByType(pb.MailRewardType_MrtAccTicket, amount2)
					} else if id == "pvpScore" || id == "matchScore" {
						if uid > 0 {
							resType := consts.Score
							if id == "matchScore" {
								resType = consts.MatchScore
							}
							utils.PlayerMqPublish(uid, pb.RmqType_Bonus, &pb.RmqBonus{
								ChangeRes: []*pb.Resource{ &pb.Resource{
									Type:   int32(resType),
									Amount: int32(amount2),
								} },
							})
						}
					} else if strings.HasPrefix(id, "card") {
						strCardID := id[4:]
						cardID, err := strconv.Atoi(strCardID)
						if err == nil {
							mailReward.AddCard(uint32(cardID), amount2)
						}
					} else if strings.HasPrefix(id, "SK") {
						mailReward.AddItem(pb.MailRewardType_MrtCardSkin, id, 1)
					} else if strings.HasPrefix(id, "BX") {
						mailReward.AddItem(pb.MailRewardType_MrtTreasure, id, 1)
					} else if strings.HasPrefix(id, "HF") {
						mailReward.AddItem(pb.MailRewardType_MrtHeadFrame, id[2:], 1)
					} else if strings.HasPrefix(id, "EQ") {
						mailReward.AddItem(pb.MailRewardType_MrtEquip, id, 1)
					} else if strings.HasPrefix(id, "EJ") {
						emojiTeam, _ := strconv.Atoi(id[2:])
						mailReward.AddEmoji(emojiTeam)
					}
					break
				}
			}

			mailIDs = append(mailIDs, sender.Send())
		}
		close(c)
	})
	<-c

	reply, _ := json.Marshal(map[string]interface{}{
		"errcode": 0,
		"mailIDs": mailIDs,
	})
	writer.Write(reply)
}

func cardAmountHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	accountType := pb.AccountTypeEnum(args["accountType"].(float64))
	var area int
	if area2, ok := args["area"].(float64); ok {
		area = int(area2)
	}
	c := make(chan struct{})
	var reply []*pb.CardsAmountLog
	var at2Logs map[pb.AccountTypeEnum]map[uint32]int

	evq.CallLater(func() {
		at2Logs = module.Card.GetCardAmountLog(accountType, area)
		close(c)
	})

	getCardList()

	<- c
	for at, logs := range at2Logs {
		msg := &pb.CardsAmountLog{AccountType: at}
		reply = append(reply, msg)
		for cardID, amount := range logs {
			msg.Logs = append(msg.Logs, &pb.CardAmountLog{
				CardID: cardID,
				Amount: int32(amount),
				CardName: cardID2CardName[cardID],
			})
		}
	}
	cardAmountList, _ := json.Marshal(reply)
	writer.Write(cardAmountList)
}

func cardLevelHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	accountType := pb.AccountTypeEnum(args["accountType"].(float64))
	cardID := uint32(args["cardID"].(float64))
	var area int
	if area2, ok := args["area"].(float64); ok {
		area = int(area2)
	}
	c := make(chan struct{})
	var reply []*pb.CardsLevelLog
	var at2Logs map[pb.AccountTypeEnum]map[uint32]map[int]int

	evq.CallLater(func() {
		at2Logs = module.Card.GetCardLevelLog(accountType, cardID, area)
		close(c)
	})

	getCardList()

	<- c

	for at, logs := range at2Logs {
		msg := &pb.CardsLevelLog{AccountType: at}
		reply = append(reply, msg)
		for cardID, lv2Amount := range logs {
			cardMsg := &pb.CardLevelLog{CardID: cardID, CardName: cardID2CardName[cardID]}
			msg.Logs = append(msg.Logs, cardMsg)
			for lv := 1; lv <= 5; lv++ {
				cardMsg.Levels = append(cardMsg.Levels, &pb.CardLevelLog_LevelAmount{
					Level: int32(lv),
					Amount: int32(lv2Amount[lv]),
				})
			}
		}
	}
	cardLevelList, _ := json.Marshal(reply)
	writer.Write(cardLevelList)
}

/*
func cardPoolHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	accountType := pb.AccountTypeEnum(args["accountType"].(float64))
	pvpLevel := int(args["pvpLevel"].(float64))
	var area int
	if area2, ok := args["area"].(float64); ok {
		area = int(area2)
	}
	c := make(chan struct{})
	var reply []*pb.CardPoolsLog
	var cardPoolLogs map[pb.AccountTypeEnum]map[int]map[uint32]int
	var battleAmountLogs map[pb.AccountTypeEnum]map[int]int

	evq.CallLater(func() {
		cardPoolLogs, battleAmountLogs = module.Card.GetCardPoolLog(accountType, pvpLevel, area)
		close(c)
	})

	getCardList()

	<- c

	for at, lv2BattleAmount := range battleAmountLogs {
		msg := &pb.CardPoolsLog{AccountType: at}
		reply = append(reply, msg)
		for lv, battleAmount := range lv2BattleAmount {
			lvMsg := &pb.CardPoolLog{PvpLevel: int32(lv), BattleAmount: int32(battleAmount)}
			msg.Logs = append(msg.Logs, lvMsg)

			id2CardAmount := cardPoolLogs[at][lv]
			for cardID, amount := range id2CardAmount {
				lvMsg.CardLogs = append(lvMsg.CardLogs, &pb.CardPoolLog_CardLog{
					CardID: cardID,
					Amount: int32(amount),
					CardName: cardID2CardName[cardID],
				})
			}
		}
	}
	cardPoolList, _ := json.Marshal(reply)
	writer.Write(cardPoolList)
}
*/

func fetchCardListHandler(writer http.ResponseWriter, request *http.Request) {
	getCardList()
	writer.Write(cardList)
}

func updateMailDeadlineHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	var ok bool
	mailID := int(args["mailID"].(float64))
	deadLine := int64(args["deadLine"].(float64))

	if !checkSign(map[string]string{
		"mailID": strconv.Itoa(mailID),
		"deadLine": strconv.FormatInt(deadLine, 10),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	c := make(chan struct{})
	evq.CallLater(func() {
		ok = module.Mail.UpdateMailDeadLine(mailID, deadLine)
		close(c)
	})

	<- c
	if ok {
		writer.Write(success)
	} else {
		writer.Write(noPlayerErr)
	}
}

func fetchAreaListHandler(writer http.ResponseWriter, request *http.Request) {
	if len(areaList) > 0 {
		writer.Write(areaList)
		return
	}

	c := make(chan struct{})
	reply := map[string]interface{} {}
	evq.CallLater(func() {
		var areas []int
		areaGameData := gamedata.GetGameData(consts.AreaConfig).(*gamedata.AreaConfigGameData)
		for _, areaCfg := range areaGameData.Areas {
			areas = append(areas, areaCfg.Area)
		}
		reply["areaList"] = areas
		close(c)
	})

	<- c
	areaList, _ = json.Marshal(reply)
	writer.Write(areaList)
}

func fetchServerStatusHandler(writer http.ResponseWriter, request *http.Request) {
	reply := map[string]interface{} {
		"status": 0,
		"message": "",
	}

	c := make(chan struct{})
	evq.CallLater(func() {
		serverStatus := mod.GetServerStatus()
		if serverStatus != nil {
			reply["status"] = int(serverStatus.Status)
			reply["message"] = serverStatus.Message
		}
		close(c)
	})

	<- c
	status, _ := json.Marshal(reply)
	writer.Write(status)
}

func updateServerStatusHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	status := int(args["status"].(float64))
	message := args["message"].(string)

	if !checkSign(map[string]string{
		"status": strconv.Itoa(status),
		"message": message,
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	evq.CallLater(func() {
		mod.updateServerMaintainMessage(message)
	})

	writer.Write(success)
}

func fetchNoticeHandler(writer http.ResponseWriter, request *http.Request) {
	var reply []map[string]interface{}
	c := make(chan struct{})

	evq.CallLater(func() {

		mod.forEachNotice(func(channel string, notice *pb.LoginNotice) {
			for _, data := range notice.Notices {
				reply = append(reply, map[string]interface{}{
					"channel": channel,
					"title":   data.Title,
					"content": data.Content,
				})
			}
		})

		close(c)
	})
	<- c
	notices, _ := json.Marshal(reply)
	writer.Write(notices)
}

func updateNoticeHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	t := int64(args["time"].(float64))
	notices := args["notices"].([]interface{})
	if !checkSign(map[string]string{
		"time": strconv.FormatInt(t, 10),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	var channel string
	noticeMsg := &pb.LoginNotice{}
	for _, notice := range notices {
		newNotice := notice.(map[string]interface{})
		if channel == "" {
			channel = newNotice["channel"].(string)
		}

		noticeMsg.Notices = append(noticeMsg.Notices, &pb.LoginNoticeData{
			Title:   newNotice["title"].(string),
			Content: newNotice["content"].(string),
		})
	}

	evq.CallLater(func() {
		noticeMsg.Version = int32(mod.genLoginNoticeVersion())
		mod.onLoginNoticeUpdate(channel, noticeMsg)
	})

	writer.Write(success)
}

func importDirtyWordsHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}
	addWords := args["addWords"].(string)
	delWords := args["delWords"].(string)
	wordType := int(args["wordType"].(float64))     // 1: 精确敏感词， 2：模糊敏感词
	if !checkSign(map[string]string{
		"addWords": addWords,
		"delWords": delWords,
		"wordType": strconv.Itoa(wordType),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}
	var isAccurate bool
	isAccurate = wordType == 1
	c := make(chan struct{})
	evq.CallLater(func() {
		utils.AddDirtyWords(addWords, delWords, isAccurate)
		close(c)
	})
	<-c
	writer.Write(success)
}

func forbidAccountHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}
	area := int(args["area"].(float64))
	uid := common.UUid(args["playerID"].(float64))
	forbidType := int(args["forbidType"].(float64))
	isForbid := args["isForbid"].(bool)
	overTime := int64(args["overTime"].(float64))
	opTime := int64(args["opTime"].(float64))
	if !checkSign(map[string]string{
		"playerID": uid.String(),
		"area":strconv.Itoa(int(area)),
		"forbidType": strconv.Itoa(forbidType),
		"overTime": strconv.Itoa(int(overTime)),
		"opTime": strconv.Itoa(int(opTime)),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	var playerAttr *attribute.AttrMgr
	for _, region := range []uint32{1, 2} {
		playerAttr = attribute.NewAttrMgr("player", uid, false, region)
		err := playerAttr.Load(true)
		if err == nil {
			break
		} else {
			playerAttr = nil
		}
	}

	if playerAttr == nil {
		writer.Write(noPlayerErr)
		return
	}

	player := module.Player.NewPlayerByAttr(uid, playerAttr)
	if player.GetArea() != area {
		writer.Write(noPlayerErr)
		return
	}

	c := make(chan struct{})
	evq.CallLater(func() {
		switch forbidType {
		case consts.ForbidAccount:
			utils.PlayerMqPublish(uid, pb.RmqType_ForbidLogin, &pb.RmqForbidLogin{
				IsForbid: isForbid,
				OverTime: overTime,
			})

			if isForbid {
				utils.UpdateForbidList(uid, consts.ForbidMonitor, false, 0, true)
			}else {
				utils.UpdateForbidList(uid, forbidType, false, 0, false)
			}

		case consts.ForbidChat:
			utils.PlayerMqPublish(uid, pb.RmqType_ForbidChat, &pb.RmqForbidChat{
				IsForbid: isForbid,
				OverTime: overTime,
			})
		case consts.ForbidMonitor:
			utils.PlayerMqPublish(uid, pb.RmqType_MonitorAccount, &pb.RmqMonitorAccount{
				IsForbid: isForbid,
				OverTime:overTime,
				OpTime: opTime,
			})
			utils.UpdateForbidList(uid, forbidType, isForbid, opTime, false)
		}
		close(c)
	})
	<-c
	writer.Write(success)
}

//func fetchDirtyWordHandler(writer http.ResponseWriter, request *http.Request) {
//	var accurateWords, fuzzyWords string
//	var items []byte
//	var err error
//	c := make(chan struct{})
//	evq.CallLater(func() {
//		accurateWords, fuzzyWords = utils.FetchDirtyWords()
//		words := map[string]interface{}{
//			"accurateWords": accurateWords,
//			"fuzzyWords": fuzzyWords,
//		}
//		items, err = json.Marshal(words)
//		close(c)
//	})
//	<-c
//	if err == nil {
//		writer.Write(items)
//	}
//}

//func fetchForbidAccounts(writer http.ResponseWriter, request *http.Request) {
//	var replay []byte
//	var err error
//	c := make(chan struct{})
//	evq.CallLater(func() {
//		items := utils.FetchForbidAccounts()
//		replay, err = json.Marshal(items)
//		close(c)
//	})
//	<-c
//	if err == nil {
//		writer.Write(replay)
//	}
//}

func forbidIpAddrHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}
	ipaddr := args["ip"].(string)
	isForbid := args["isForbid"].(bool)
	if !checkSign(map[string]string{
		"ip": ipaddr,
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}
	c := make(chan struct{})
	evq.CallLater(func() {
		utils.ForbidIpAddr(ipaddr, isForbid)
		close(c)
	})
	<-c
	writer.Write(success)
}

func compensateRechargeHandler(writer http.ResponseWriter, request *http.Request) {
	args := prepareRequest(writer, request)
	if args == nil {
		return
	}

	uid := common.UUid(args["uid"].(float64))
	cpOrderID := args["cpOrderID"].(string)
	channelOrderID := args["channelOrderID"].(string)
	goodsID := args["goodsID"].(string)
	t := int64(args["time"].(float64))
	diffTime := time.Now().Unix() - t
	if diffTime > 10 || diffTime < -10 {
		writer.Write(noPlayerErr)
		return
	}

	if !checkSign(map[string]string{
		"uid": uid.String(),
		"cpOrderID": cpOrderID,
		"channelOrderID": channelOrderID,
		"goodsID": goodsID,
		"time": strconv.FormatInt(t, 10),
	}, args["sign"].(string)) {
		writer.Write(signErr)
		return
	}

	c := make(chan struct{})
	evq.CallLater(func() {
		utils.PlayerMqPublish(uid, pb.RmqType_CompensateRecharge, &pb.RmqCompensateRecharge{
			CpOrderID: cpOrderID,
			ChannelOrderID: channelOrderID,
			GoodsID: goodsID,
		})
		close(c)
	})
	<-c
	writer.Write(success)
}

func onGameDataReload() {
	itemList = []byte{}
	cardList = []byte{}
	areaList = []byte{}
}

func initializeGmtool() {
	gamedata.OnReload = onGameDataReload
	http.HandleFunc("/player_info", wrapHttpHandler(playerInfoHandler))
	http.HandleFunc("/forbid_login", wrapHttpHandler(forbidLoginHandler))
	http.HandleFunc("/forbid_chat", wrapHttpHandler(forbidChatHandler))
	http.HandleFunc("/fetch_item_list", wrapHttpHandler(fetchItemListHandler))
	http.HandleFunc("/send_mail", wrapHttpHandler(sendMailHandler))
	http.HandleFunc("/card_amount", wrapHttpHandler(cardAmountHandler))
	http.HandleFunc("/card_level", wrapHttpHandler(cardLevelHandler))
	//http.HandleFunc("/card_pool", wrapHttpHandler(cardPoolHandler))
	http.HandleFunc("/fetch_card_list", wrapHttpHandler(fetchCardListHandler))
	http.HandleFunc("/update_mail_deadline", wrapHttpHandler(updateMailDeadlineHandler))
	http.HandleFunc("/fetch_area_list", wrapHttpHandler(fetchAreaListHandler))
	http.HandleFunc("/fetch_server_status", wrapHttpHandler(fetchServerStatusHandler))
	http.HandleFunc("/update_server_status", wrapHttpHandler(updateServerStatusHandler))
	http.HandleFunc("/fetch_notice", wrapHttpHandler(fetchNoticeHandler))
	http.HandleFunc("/update_notice", wrapHttpHandler(updateNoticeHandler))
	http.HandleFunc("/import_words", wrapHttpHandler(importDirtyWordsHandler))
	http.HandleFunc("/forbid_account", wrapHttpHandler(forbidAccountHandler))
	//http.HandleFunc("/fetch_words", wrapHttpHandler(fetchDirtyWordHandler))
	//http.HandleFunc("/fetch_accounts", wrapHttpHandler(fetchForbidAccounts))
	http.HandleFunc("/forbid_ip", wrapHttpHandler(forbidIpAddrHandler))
	http.HandleFunc("/compensate_recharge", wrapHttpHandler(compensateRechargeHandler))
}
