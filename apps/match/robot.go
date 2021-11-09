package main

import (
	"container/list"
	"fmt"
	"kinger/common/aicardpool"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"math/rand"
	"time"
)

var (
	robotMgr             = newMatchRobotMgr()
	_        iMatchRobot = &matchRobot{}
	_        iMatchRobot = &popularMatchRobot{}
)

type iMatchRobot interface {
	getCamp() int
	getHandCards() []*pb.SkinGCard
	getGridCards() []*pb.InGridCard
	getPvpScore() int
	getPvpLevel() int
	getID() common.UUid
	getHeadImgUrl() string
	getHeadFrame() string
}

type matchRobot struct {
	id        common.UUid
	attr      *attribute.AttrMgr
	pvpLevel  int
	gridCards []*pb.InGridCard
}

func newRobotByAttr(id common.UUid, attr *attribute.AttrMgr) *matchRobot {
	return &matchRobot{
		id:   id,
		attr: attr,
	}
}

func newRobotByPvpLevel(pvpLevel int) *matchRobot {
	id := common.GenUUid("matchRobot")
	attr := attribute.NewAttrMgr(fmt.Sprintf("matchRobot%d", mService.AppID), id)
	camps := []int{consts.Wei, consts.Shu, consts.Wu}
	camp := camps[rand.Intn(len(camps))]
	cards := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).RandomPvpRobotCards(pvpLevel, camp, false)
	cardsAttr := attribute.NewListAttr()
	for _, card := range cards {
		cardsAttr.AppendUInt32(card.GetGCardID())
	}
	attr.SetListAttr("handCards", cardsAttr)
	score := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData).Ranks[pvpLevel].LevelUpStar
	attr.SetInt("pvpScore", score)
	attr.SetInt("camp", camp)
	glog.Infof("newRobotByPvpLevel id=%d, pvpLevel=%d, cards=%s", id, pvpLevel, cardsAttr)

	return &matchRobot{
		id:       id,
		attr:     attr,
		pvpLevel: pvpLevel,
	}
}

func newNewbiePvpRobot(camp, pvpScore, pvpLevel int, isFirstBattle bool) *matchRobot {
	attr := attribute.NewAttrMgr("robot", camp)
	battleData := gamedata.GetGameData(consts.NewbiePvp).(*gamedata.NewbiePvpGameData).Camp2Battle[camp]
	utils.ShuffleUInt32(battleData.Hand)
	handCardGids := battleData.Hand[:5]

	cardsAttr := attribute.NewListAttr()
	for _, gcardID := range handCardGids {
		cardsAttr.AppendUInt32(gcardID)
	}
	attr.SetListAttr("handCards", cardsAttr)
	attr.SetInt("pvpScore", pvpScore)
	attr.SetInt("camp", camp)

	var robotGridCards []*pb.InGridCard
	if isFirstBattle {
		poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
		for _, gridCardInfo := range battleData.GridCard {
			cardData := poolGameData.GetCardByGid(uint32(gridCardInfo[0]))
			if cardData != nil {
				robotGridCards = append(robotGridCards, &pb.InGridCard{
					GCardID: uint32(gridCardInfo[0]),
					GridID:  int32(gridCardInfo[1]),
				})
			}
		}
	}

	return &matchRobot{
		attr:      attr,
		pvpLevel:  pvpLevel,
		gridCards: robotGridCards,
	}
}

func (r *matchRobot) getHeadImgUrl() string {
	headImgUrl := r.attr.GetStr("headImgUrl")
	if headImgUrl != "" {
		return headImgUrl
	}

	cardsAttr := r.attr.GetListAttr("handCards")
	i := rand.Intn(cardsAttr.Size())
	gcardID := cardsAttr.GetUInt32(i)
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	cardData := poolGameData.GetCardByGid(gcardID)
	if cardData != nil {
		headImgUrl = fmt.Sprintf("avatar_%s_png", cardData.Head)
		r.attr.SetStr("headImgUrl", headImgUrl)
	}
	return headImgUrl
}

func (r *matchRobot) getHeadFrame() string {
	return "1"
}

func (r *matchRobot) getGridCards() []*pb.InGridCard {
	return r.gridCards
}

func (r *matchRobot) getCamp() int {
	return r.attr.GetInt("camp")
}

func (r *matchRobot) getID() common.UUid {
	return r.id
}

func (r *matchRobot) getHandCards() []*pb.SkinGCard {
	handCardsAttr := r.attr.GetListAttr("handCards")
	var cards []*pb.SkinGCard
	handCardsAttr.ForEachIndex(func(index int) bool {
		cards = append(cards, &pb.SkinGCard{
			GCardID: handCardsAttr.GetUInt32(index),
		})
		return true
	})
	return cards
}

func (r *matchRobot) getPvpScore() int {
	return r.attr.GetInt("pvpScore")
}

func (r *matchRobot) modifyPvpScore(modify int) {
	score := r.getPvpScore() + modify
	if score < 1 {
		score = 1
	}
	r.attr.SetInt("pvpScore", score)
	r.pvpLevel = 0
	r.getPvpLevel()
	//glog.Infof("robot modifyPvpScore id=%d, score=%d", r.getID(), score)
}

func (r *matchRobot) getPvpLevel() int {
	if r.pvpLevel == 0 {
		r.pvpLevel = gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData).GetPvpLevelByStar(r.getPvpScore())
	}

	if r.pvpLevel > maxMatchRobotLevel {
		r.pvpLevel = maxMatchRobotLevel
	}

	return r.pvpLevel
}

type popularMatchRobot struct {
	camp      int
	handCards []*pb.SkinGCard
	pvpLevel  int
	pvpScore  int
}

func newPopularMatchRobot(pvpScore, pvpLevel, playerCamp int, playerHandCards []*pb.SkinGCard) iMatchRobot {
	camp, handCards := aicardpool.RandomCardPool(pvpLevel, playerCamp, playerHandCards, false)
	if camp <= 0 || len(handCards) < len(playerHandCards) {
		return nil
	}

	return &popularMatchRobot{
		camp:      camp,
		handCards: handCards,
		pvpLevel:  pvpLevel,
		pvpScore:  pvpScore,
	}
}

func (r *popularMatchRobot) getCamp() int {
	return r.camp
}

func (r *popularMatchRobot) getHandCards() []*pb.SkinGCard {
	return r.handCards
}

func (r *popularMatchRobot) getGridCards() []*pb.InGridCard {
	return []*pb.InGridCard{}
}

func (r *popularMatchRobot) getPvpScore() int {
	return r.pvpScore
}

func (r *popularMatchRobot) getPvpLevel() int {
	return r.pvpLevel
}

func (r *popularMatchRobot) getID() common.UUid {
	return 1
}

func (r *popularMatchRobot) getHeadImgUrl() string {
	gcardID := r.handCards[rand.Intn(len(r.handCards))].GCardID
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	cardData := poolGameData.GetCardByGid(gcardID)
	if cardData != nil {
		return fmt.Sprintf("avatar_%s_png", cardData.Head)
	}
	return ""
}

func (r *popularMatchRobot) getHeadFrame() string {
	return "1"
}

type pvpLevelRobots struct {
	curRobot *list.Element
	robots   *list.List
}

func newPvpLevelRobots() *pvpLevelRobots {
	return &pvpLevelRobots{
		robots: list.New(),
	}
}

func (plr *pvpLevelRobots) addRobot(r *matchRobot) (delRobot *matchRobot, addElem *list.Element) {
	addElem = plr.robots.PushBack(r)
	if plr.robots.Len() > 150 {
		delElem := plr.robots.Front()
		delRobot = delElem.Value.(*matchRobot)
		plr.robots.Remove(delElem)
		if plr.curRobot == delElem {
			plr.curRobot = plr.robots.Front()
		}
	}

	if plr.curRobot == nil {
		plr.curRobot = plr.robots.Front()
	}

	return
}

func (plr *pvpLevelRobots) getRobotAmount() int {
	return plr.robots.Len()
}

func (plr *pvpLevelRobots) delRobot(elem *list.Element) {
	plr.robots.Remove(elem)
	if plr.curRobot == elem {
		plr.curRobot = plr.robots.Front()
	}
}

func (plr *pvpLevelRobots) getRobot() *matchRobot {
	if plr.curRobot == nil {
		plr.curRobot = plr.robots.Front()
	}

	if plr.curRobot == nil {
		return nil
	} else {
		r := plr.curRobot
		plr.curRobot = plr.curRobot.Next()
		if plr.curRobot == nil {
			plr.curRobot = plr.robots.Front()
		}

		return r.Value.(*matchRobot)
	}
}

type robotMgrSt struct {
	id2Robot        map[common.UUid]*list.Element
	pvpLevel2Robots map[int]*pvpLevelRobots
}

func newMatchRobotMgr() *robotMgrSt {
	return &robotMgrSt{
		id2Robot:        make(map[common.UUid]*list.Element),
		pvpLevel2Robots: make(map[int]*pvpLevelRobots),
	}
}

func (rm *robotMgrSt) addRobotToList(r *matchRobot) {
	pvpLevel := r.getPvpLevel()
	rs, ok := rm.pvpLevel2Robots[pvpLevel]
	var delRobot *matchRobot
	var addElem *list.Element
	if ok {
		delRobot, addElem = rs.addRobot(r)
	} else {
		rs = newPvpLevelRobots()
		rs.addRobot(r)
		rm.pvpLevel2Robots[pvpLevel] = rs
	}

	if delRobot != nil {
		delRobot.attr.Delete(false)
		delete(rm.id2Robot, delRobot.getID())
	}

	if addElem != nil {
		rm.id2Robot[r.getID()] = addElem
	}
}

func (rm *robotMgrSt) getRobot(player iMatchPlayer) iMatchRobot {
	pvpLevel := player.getPvpLevel()
	if pvpLevel > maxMatchRobotLevel {
		if player.getSeasonDataID() > 0 {
			return nil
		}
		return newPopularMatchRobot(player.getPvpScore(), pvpLevel, player.getCamp(), player.getHandCards())
	}

	rs, ok := rm.pvpLevel2Robots[pvpLevel]
	if !ok {
		rs = newPvpLevelRobots()
		rm.pvpLevel2Robots[pvpLevel] = rs
	}

	robotAmount := rs.getRobotAmount()
	if robotAmount < 50 {
		need := 150 - robotAmount
		for i := 0; i < need; i++ {
			robot := newRobotByPvpLevel(pvpLevel)
			rm.addRobotToList(robot)
		}
	}
	return rs.getRobot()
}

func (rm *robotMgrSt) saveRobot() {
	for _, elem := range rm.id2Robot {
		r := elem.Value.(*matchRobot)
		r.attr.Save(false)
	}
}

func (rm *robotMgrSt) delRobotFromList(pvpLevel int, elem *list.Element) {
	rs, ok := rm.pvpLevel2Robots[pvpLevel]
	if ok {
		rs.delRobot(elem)
	}
}

func (rm *robotMgrSt) onRobotBattleEnd(arg *pb.OnRobotBattleEndArg) {
	robotID := common.UUid(arg.RobotID)
	if elem, ok := rm.id2Robot[robotID]; ok {
		r := elem.Value.(*matchRobot)
		oldPvpLevel := r.getPvpLevel()
		if arg.IsWin {
			r.modifyPvpScore(1)
		} else {
			r.modifyPvpScore(-1)
		}
		pvpLevel := r.getPvpLevel()

		if oldPvpLevel != pvpLevel {
			rm.delRobotFromList(oldPvpLevel, elem)
			rm.addRobotToList(r)
		}
	}
}

func (rm *robotMgrSt) loadRobot() {
	attrs, err := attribute.LoadAll(fmt.Sprintf("matchRobot%d", mService.AppID))
	if err != nil {
		panic(err)
	}

	for _, attr := range attrs {
		attrID := attr.GetAttrID()
		var id common.UUid
		switch attrID2 := attrID.(type) {
		case int64:
		case int:
		case uint64:
		case uint32:
		case int32:
			id = common.UUid(attrID2)
		default:
			glog.Errorf("loadRobot attrID=%s", attrID2)
			continue
		}
		robot := newRobotByAttr(id, attr)
		rm.addRobotToList(robot)
	}

	timer.AddTicker(10*time.Minute, rm.saveRobot)
}
