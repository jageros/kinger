package cardpool

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module"
	"strconv"
	"kinger/gopuppy/common/timer"
	"time"
	"kinger/gopuppy/common/eventhub"
	"kinger/common/consts"
	"kinger/apps/game/module/types"
	"kinger/proto/pb"
	"kinger/gamedata"
	"fmt"
)

var (
	accountType2Log = map[pb.AccountTypeEnum]map[int]*logHub{}
)

func getLog(accountType pb.AccountTypeEnum, area int) *logHub {
	if area2Log, ok := accountType2Log[accountType]; ok {
		return area2Log[area]
	} else {
		return nil
	}
}

func getAccountTypeLogs(accountType pb.AccountTypeEnum) map[int]*logHub {
	return accountType2Log[accountType]
}

func getAreaLogs(area int) []*logHub {
	var logs []*logHub
	for _, area2Log := range accountType2Log {
		if l, ok := area2Log[area]; ok {
			logs = append(logs, l)
		}
	}
	return logs
}

func forEachLog(area int, callback func(pb.AccountTypeEnum, *logHub))  {
	for accountType, area2Log := range accountType2Log {
		if area <= 0 {
			for _, log := range area2Log {
				callback(accountType, log)
			}
		} else if l, ok := area2Log[area]; ok {
			callback(accountType, l)
		}
	}
}

func forEachAccountTypeLog(area int, accountType pb.AccountTypeEnum, callback func(*logHub))  {
	area2Log, ok := accountType2Log[accountType]
	if !ok {
		return
	}
	if area <= 0 {
		for _, log := range area2Log {
			callback(log)
		}
	} else if l, ok := area2Log[area]; ok {
		callback(l)
	}
}

func initAreaLog(gdata gamedata.IGameData) {
	areaGameData := gdata.(*gamedata.AreaConfigGameData)
	for accountType, _ := range pb.AccountTypeEnum_name {
		accountType2 := pb.AccountTypeEnum(accountType)
		if accountType2 == pb.AccountTypeEnum_WxgameIos || accountType2 == pb.AccountTypeEnum_UnknowAccountType {
			continue
		}

		area2Log, ok := accountType2Log[accountType2]
		if !ok {
			area2Log = map[int]*logHub{}
			accountType2Log[accountType2] = area2Log
		}

		for _, areaCfg := range areaGameData.Areas {
			area := areaCfg.Area
			if _, ok := area2Log[area]; ok {
				continue
			}
			log := newLog(accountType2, area)
			area2Log[area] = log
		}
	}
}

func initializeLog() {
	areaGameData := gamedata.GetGameData(consts.AreaConfig)
	areaGameData.AddReloadCallback(initAreaLog)
	initAreaLog(areaGameData)

	timer.AddTicker(10 * time.Minute, SaveAllCardLog)

	eventhub.Subscribe(consts.EvEndPvpBattle, func(args ...interface{}) {
		player := args[0].(types.IPlayer)
		fighterData := args[2].(*pb.EndFighterData)
		accountType := player.GetLogAccountType()
		log := getLog(accountType, player.GetArea())
		if log == nil {
			return
		}

		/*pvpLevel := player.GetPvpLevel()
		var cards []uint32
		poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
		for _, card := range fighterData.InitHandCards {
			cardData := poolGameData.GetCardByGid(card.GCardID)
			if cardData != nil {
				cards = append(cards, cardData.CardID)
			}
		}*/
		mod.LogBattleCards(player, fighterData)
		//log.getBattleLog(pvpLevel).battleEnd(cards)
	})
}

func SaveAllCardLog() {
	for _, area2Log := range accountType2Log {
		for _, log := range area2Log {
			log.save()
		}
	}
}

type cardLog struct {
	attr *attribute.MapAttr
	levelAttr *attribute.MapAttr
}

func newCardLog(cardID uint32) *cardLog {
	attr := attribute.NewMapAttr()
	attr.SetUInt32("id", cardID)
	return newCardLogByAttr(attr)
}

func newCardLogByAttr(attr *attribute.MapAttr) *cardLog {
	cl := &cardLog{
		attr: attr,
		levelAttr: attr.GetMapAttr("level"),
	}
	if cl.levelAttr == nil {
		cl.levelAttr = attribute.NewMapAttr()
		attr.SetMapAttr("level", cl.levelAttr)
	}
	return cl
}

func (cl *cardLog) getCardID() uint32 {
	return cl.attr.GetUInt32("id")
}

func (cl *cardLog) getAmount() int {
	return cl.attr.GetInt("amount")
}

func (cl *cardLog) modifyAmount(amount int) {
	cl.attr.SetInt("amount", cl.getAmount() + amount)
}

func (cl *cardLog) modifyLevel(oldLevel, newLevel int) {
	if oldLevel > 0 {
		key := strconv.Itoa(oldLevel)
		cl.levelAttr.SetInt(key, cl.levelAttr.GetInt(key) - 1)
	}

	if newLevel > 0 {
		key := strconv.Itoa(newLevel)
		cl.levelAttr.SetInt(key, cl.levelAttr.GetInt(key) + 1)
	}
}

func (cl *cardLog) forEachLevel(callback func(level, amount int))  {
	cl.levelAttr.ForEachKey(func(key string) {
		level, _ := strconv.Atoi(key)
		callback(level, cl.levelAttr.GetInt(key))
	})
}

type battleLog struct {
	attr *attribute.MapAttr
	cardsAttr *attribute.MapAttr
}

func newBattleLog(pvpLevel int) *battleLog {
	attr := attribute.NewMapAttr()
	attr.SetInt("pvpLevel", pvpLevel)
	return newBattleLogByAttr(attr)
}

func newBattleLogByAttr(attr *attribute.MapAttr) *battleLog {
	bl := &battleLog{
		attr: attr,
		cardsAttr: attr.GetMapAttr("cards"),
	}
	if bl.cardsAttr == nil {
		bl.cardsAttr = attribute.NewMapAttr()
		attr.SetMapAttr("cards", bl.cardsAttr)
	}
	return bl
}

func (bl *battleLog) getPvpLevel() int {
	return bl.attr.GetInt("pvpLevel")
}

func (bl *battleLog) getAmount() int {
	return bl.attr.GetInt("amount")
}

func (bl *battleLog) battleEnd(cards []uint32) {
	bl.attr.SetInt("amount", bl.getAmount() + 1)
	for _, cardID := range cards {
		key := strconv.Itoa(int(cardID))
		bl.cardsAttr.SetInt(key, bl.cardsAttr.GetInt(key) + 1)
	}
}

func (bl *battleLog) forEachCard(callback func(cardID uint32, amount int)) {
	bl.cardsAttr.ForEachKey(func(key string) {
		cardID, _ := strconv.Atoi(key)
		callback(uint32(cardID), bl.cardsAttr.GetInt(key))
	})
}

type logHub struct {
	attr *attribute.AttrMgr
	cardsAttr *attribute.ListAttr
	battlesAttr *attribute.ListAttr
	id2CardLog map[uint32]*cardLog
	//level2BattleLog map[int]*battleLog
}

func newLog(accountType pb.AccountTypeEnum, area int) *logHub {
	attr := attribute.NewAttrMgr(fmt.Sprintf("cardLog_%s_%d", accountType, area), module.Service.GetAppID())
	attr.Load()
	log := &logHub{
		attr: attr,
		id2CardLog: map[uint32]*cardLog{},
		//level2BattleLog: map[int]*battleLog{},
	}

	log.cardsAttr = attr.GetListAttr("cards")
	if log.cardsAttr == nil {
		log.cardsAttr = attribute.NewListAttr()
		attr.SetListAttr("cards", log.cardsAttr)
	}
	log.cardsAttr.ForEachIndex(func(index int) bool {
		cardAttr := log.cardsAttr.GetMapAttr(index)
		cl := newCardLogByAttr(cardAttr)
		log.id2CardLog[cl.getCardID()] = cl
		return true
	})
	/*
	log.battlesAttr = attr.GetListAttr("battles")
	if log.battlesAttr == nil {
		log.battlesAttr = attribute.NewListAttr()
		attr.SetListAttr("battles", log.battlesAttr)
	}
	log.battlesAttr.ForEachIndex(func(index int) bool {
		battleAttr := log.battlesAttr.GetMapAttr(index)
		bl := newBattleLogByAttr(battleAttr)
		log.level2BattleLog[bl.getPvpLevel()] = bl
		return true
	})*/
	return log
}

func (l *logHub) save() {
	l.attr.Save(false)
}

func (l *logHub) getCardLog(cardID uint32) *cardLog {
	if cardID <= 0 {
		return nil
	}
	cl, ok := l.id2CardLog[cardID]
	if !ok {
		cl = newCardLog(cardID)
		l.cardsAttr.AppendMapAttr(cl.attr)
		l.id2CardLog[cardID] = cl
	}
	return cl
}

/*
func (l *logHub) getBattleLog(pvpLevel int) *battleLog {
	if pvpLevel <= 0 {
		return nil
	}
	bl, ok := l.level2BattleLog[pvpLevel]
	if !ok {
		bl = newBattleLog(pvpLevel)
		l.battlesAttr.AppendMapAttr(bl.attr)
		l.level2BattleLog[pvpLevel] = bl
	}
	return bl
}
*/

func (l *logHub) modifyCardAmount(cardID uint32, amount int) {
	if amount == 0 {
		return
	}
	l.getCardLog(cardID).modifyAmount(amount)
}

func (l *logHub) modifyCardLevel(cardID uint32, oldLevel, newLevel int) {
	if oldLevel == newLevel {
		return
	}
	l.getCardLog(cardID).modifyLevel(oldLevel, newLevel)
}

func (l *logHub) forEachCardLog(callback func(cl *cardLog)) {
	for _, cl := range l.id2CardLog {
		callback(cl)
	}
}

/*
func (l *logHub) forEachBattleLog(callback func(bl *battleLog)) {
	for _, bl := range l.level2BattleLog {
		callback(bl)
	}
}
*/
