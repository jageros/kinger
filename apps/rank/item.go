package main

import (
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/attribute"
	"strconv"
	"time"
	"kinger/gamedata"
	"kinger/common/consts"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"kinger/common/utils"
)

//const maxRank = 200

type rankList struct {
	rankType pb.RankType
	uid2Item map[common.UUid]*rankItem
	items []*rankItem
}

func newRankList(rankType pb.RankType) *rankList {
	return &rankList{
		rankType: rankType,
		uid2Item: map[common.UUid]*rankItem{},
	}
}

func (rl *rankList) fixSize() {
	size := len(rl.items)
	maxRank := getIRankType(rl.rankType).getMaxRank()
	var index int
	if rl.rankType == pb.RankType_RtCrossArea {
		var honor []*rankItem
		var blot []*rankItem
		for i, item := range rl.items{
			if item.getCrossAreaHonor() < 0 {
				index = i
				break
			}
		}
		if index > 0 {
			if len(rl.items[index:]) > maxRank {
				for i := index+maxRank; i < size; i++ {
					delete(rl.uid2Item, rl.items[i].getUid())
				}
				blot = rl.items[index: index + maxRank]
			}else {
				blot = rl.items[index:]
			}

			itemsSize := len(rl.items[:index])
			if  itemsSize > maxRank {
				for i := maxRank; i < itemsSize; i++ {
					delete(rl.uid2Item, rl.items[i].getUid())
				}
				honor = rl.items[:maxRank]
			}else {
				honor = rl.items[:index]
			}

			rl.items = honor
			rl.items = append(rl.items, blot...)
		}
	}

	if size > maxRank && index == 0{
		for i := maxRank; i < size; i++ {
			delete(rl.uid2Item, rl.items[i].getUid())
		}
		rl.items = rl.items[:maxRank]
	}
}

func (rl *rankList) forEach(callback func(i int, ri *rankItem) bool)  {
	for i, ri := range rl.items {
		if !callback(i, ri) {
			return
		}
	}
}

func (rl *rankList) append(ri *rankItem) {
	if _, ok := rl.uid2Item[ri.getUid()]; ok {
		return
	}

	if rl.rankType == pb.RankType_RtCrossArea && ri.getCrossAreaHonor() == 0{
		return
	}

	rl.items = append(rl.items, ri)
	rl.uid2Item[ri.getUid()] = ri
}

func (rl *rankList) Len() int {
	return len(rl.items)
}

func (rl *rankList) Swap(i, j int) {
	rl.items[i], rl.items[j] = rl.items[j], rl.items[i]
}

func (rl *rankList) Less(i, j int) bool {
	iRt := getIRankType(rl.rankType)
	if iRt == nil {
		return false
	}
	return iRt.rankLess(rl, i, j)
}


type rankItem struct {
	uid  common.UUid
	attr *attribute.AttrMgr
}

func newRankItem(attrName string, uid common.UUid, name string, area int) *rankItem {
	attr := attribute.NewAttrMgr(attrName, uid)
	attr.SetStr("name", name)
	attr.SetInt("area", area)
	return newRankItemByAttr(uid, attr)
}

func newRankItemByAttr(uid common.UUid, attr *attribute.AttrMgr) *rankItem {
	return &rankItem{
		uid:  uid,
		attr: attr,
	}
}

func (ri *rankItem) packItemMsg(rankType iRankType, isTotalBoard bool) *pb.RankItem {
	msg := &pb.RankItem{
		Uid:      uint64(ri.uid),
		Name:     ri.getName(),
		PvpScore: int32(ri.getPvpScore()),
		Rank:     int32(rankType.getRank(ri, isTotalBoard)),
		LastRank: int32(ri.getLastRank()),
		Camp:     int32(ri.getCamp()),
		WinDiff: int32(ri.getSeasonWinDiff()),
		CrossAreaHonor: int32(ri.getCrossAreaHonor()),
		RankScore: int32(ri.getRankScore()),
	}

	msg = rankType.getMsgData(ri, msg, isTotalBoard)

	return msg
}

func (ri *rankItem) packUserMsg() *pb.RankUser {
	return &pb.RankUser{
		FightCards: ri.getFightCards(),
	}
}

func (ri *rankItem) getUid() common.UUid {
	return ri.uid
}

func (ri *rankItem) getCrossAreaHonor() int {
	return ri.attr.GetInt("caHonor")
}

func (ri *rankItem) setCrossAreaHonor(val int) {
	ri.attr.SetInt("caHonor", val)
}

func (ri *rankItem) getName() string {
	return ri.attr.GetStr("name")
}

func (ri *rankItem) getArea() int {
	return ri.attr.GetInt("area")
}

func (ri *rankItem) setName(name string) {
	ri.attr.SetStr("name", name)
}

func (ri *rankItem) getPvpScore() int {
	return ri.attr.GetInt("pvpScore")
}

func (ri *rankItem) getRankScore() int {
	return ri.attr.GetInt("rankScore")
}
func (ri *rankItem) setRankScore(score int) {
	ri.attr.SetInt("rankScore", score)
}

func (ri *rankItem) getSeasonWinDiff() int {
	return ri.attr.GetInt("winDiff")
}

func (ri *rankItem) getSeasonWinCnt() int {
	return ri.attr.GetInt("winCnt")
}

func (ri *rankItem) getCamp() int {
	return ri.attr.GetInt("camp")
}

// 昨天区榜排名
func (ri *rankItem) getLastRank() int {
	return ri.attr.GetInt("lastRank")
}
func (ri *rankItem) setLastRank(r int) {
	ri.attr.SetInt("lastRank", r)
}

// 今天区榜排名
func (ri *rankItem) getTodayRank() int {
	return ri.attr.GetInt("lastAreaRank")
}
func (ri *rankItem) setTodayRank(r int) {
	ri.attr.SetInt("lastAreaRank", r)
}

// 昨天总榜排名
func (ri *rankItem) setLastTotalRank(rank int) {
	ri.attr.SetInt("lastTotalRank", rank)
}
func (ri *rankItem) getLastTotalRank() int {
	return ri.attr.GetInt("lastTotalRank")
}

// 今天总榜排名
func (ri *rankItem) setTodayTotalRank(rank int) {
	ri.attr.SetInt("todayTotalRank", rank)
}
func (ri *rankItem) getTodayTotalRank() int {
	return ri.attr.GetInt("todayTotalRank")
}

// 昨天锦标赛区榜排名
func (ri *rankItem) getLastSeasonRank() int {
	return ri.attr.GetInt("lastSeasonRank")
}
func (ri *rankItem) setLastSeasonRank(r int) {
	ri.attr.SetInt("lastSeasonRank", r)
}

// 今天锦标赛区榜排名
func (ri *rankItem) getTodaySeasonRank() int {
	return ri.attr.GetInt("todaySeasonRank")
}
func (ri *rankItem) setTodaySeasonRank(r int) {
	ri.attr.SetInt("todaySeasonRank", r)
}

// 昨天跨区荣誉榜排名
func (ri *rankItem) getLastCrossAreaRank() int {
	return ri.attr.GetInt("lastCrossAreaRank")
}
func (ri *rankItem) setLastCrossAreaRank(r int) {
	ri.attr.SetInt("lastCrossAreaRank", r)
}

// 今天跨区荣誉榜排名
func (ri *rankItem) getTodayCrossAreaRank() int {
	return ri.attr.GetInt("todayCrossAreaRank")
}
func (ri *rankItem) setTodayCrossAreaRank(r int) {
	ri.attr.SetInt("todayCrossAreaRank", r)
}

func (ri *rankItem) getUpdateTime() int64 {
	return ri.attr.GetInt64("updateTime")
}

func (ri *rankItem) setLadderUpdateTime(t int64) {
	ri.attr.SetInt64("updateTime", t)
}

func (ri *rankItem) getSeasonUpdateTime() int64 {
	return ri.attr.GetInt64("seasonUpdateTime")
}

func (ri *rankItem) setSeasonUpdateTime(t int64) {
	ri.attr.SetInt64("seasonUpdateTime", t)
}

func (ri *rankItem) getCrossAreaUpdateTime() int64 {
	return ri.attr.GetInt64("crossAreaUpdateTime")
}
func (ri *rankItem) setCrossAreaUpdateTime(t int64) {
	ri.attr.SetInt64("crossAreaUpdateTime", t)
}

func (ri *rankItem) getRebornCnt() int {
	return ri.attr.GetInt("rebornCnt")
}

func (ri *rankItem) setRebornCnt(cnt int) {
	ri.attr.SetInt("rebornCnt", cnt)
}

// 兼容老数据
func (ri *rankItem) oldGetFightCards() []*pb.SkinGCard {
	fightCardsAttr := ri.attr.GetListAttr("fightCards")
	fightCardSkinsAttr := ri.attr.GetMapAttr("fightCardSkins")
	var cards []*pb.SkinGCard
	if fightCardsAttr == nil {
		return cards
	}
	fightCardsAttr.ForEachIndex(func(index int) bool {
		c := &pb.SkinGCard{
			GCardID: fightCardsAttr.GetUInt32(index),
		}
		if fightCardSkinsAttr != nil {
			c.Skin = fightCardSkinsAttr.GetStr(strconv.Itoa(int(c.GCardID)))
		}
		cards = append(cards, c)
		return true
	})
	return cards
}

func (ri *rankItem) getFightCards() []*pb.SkinGCard {
	fightCardsAttr :=  ri.attr.GetListAttr("fightCards2")
	if fightCardsAttr == nil {
		return ri.oldGetFightCards()
	}
	var cards []*pb.SkinGCard
	fightCardsAttr.ForEachIndex(func(index int) bool {
		cardAttr := fightCardsAttr.GetMapAttr(index)
		cards = append(cards, &pb.SkinGCard{
			GCardID: cardAttr.GetUInt32("gcardID"),
			Skin: cardAttr.GetStr("skin"),
			Equip: cardAttr.GetStr("equip"),
		})
		return true
	})
	return cards
}

func (ri *rankItem) setUpdateTime(rankType pb.RankType, t int64) {
	getIRankType(rankType).setUpdateTime(ri, t)
}

func (ri *rankItem) update(playerInfo *pb.UpdatePvpScoreArg) (updateRankTypes common.IntSet) {

	oldScore := ri.getPvpScore()
	updateRankTypes = common.IntSet{}
	if oldScore != int(playerInfo.PvpScore) {
		updateRankTypes.Add(int(pb.RankType_RtLadder))
		ri.attr.SetInt("pvpScore", int(playerInfo.PvpScore))
	}

	oldRankScore := ri.getRankScore()
	if oldRankScore != int(playerInfo.RankScore) {
		updateRankTypes.Add(int(pb.RankType_RtLadder))
		ri.setRankScore(int(playerInfo.RankScore))
		//ri.attr.SetInt("rankScore", int(playerInfo.RankScore))
	}
	oldRebornCnt := ri.getRebornCnt()
	if oldRebornCnt != int(playerInfo.RebornCnt) {
		updateRankTypes.Add(int(pb.RankType_RtLadder))
		updateRankTypes.Add(int(pb.RankType_RtSeason))
		if ri.getCrossAreaHonor() > 0 {
			updateRankTypes.Add(int(pb.RankType_RtCrossArea))
		}
		ri.setRebornCnt(int(playerInfo.RebornCnt))
	}

	if playerInfo.Camp > 0 {
		ri.attr.SetInt("camp", int(playerInfo.Camp))
	}

	oldWinDiff := ri.getSeasonWinDiff()
	if oldWinDiff != int(playerInfo.WinDiff) {
		updateRankTypes.Add(int(pb.RankType_RtSeason))
		ri.attr.SetInt("winDiff", int(playerInfo.WinDiff))
	}

	oldWinCnt := ri.getSeasonWinCnt()
	if oldWinCnt != int(playerInfo.WinCnt) {
		updateRankTypes.Add(int(pb.RankType_RtSeason))
		ri.attr.SetInt("winCnt", int(playerInfo.WinCnt))
	}

	oldCrossAreaHonor := ri.getCrossAreaHonor()
	if oldCrossAreaHonor != int(playerInfo.CrossAreaHonor) {
		updateRankTypes.Add(int(pb.RankType_RtCrossArea))
		ri.setCrossAreaHonor(int(playerInfo.CrossAreaHonor))
	}

	now := time.Now().Unix()
	updateRankTypes.ForEach(func(rankType int) {
		ri.setUpdateTime(pb.RankType(rankType), now)
	})

	if len(playerInfo.HandCards) > 0 {
		fightCardsAttr := attribute.NewListAttr()
		ri.attr.SetListAttr("fightCards2", fightCardsAttr)

		for _, c := range playerInfo.HandCards {
			cardAttr := attribute.NewMapAttr()
			cardAttr.SetUInt32("gcardID", c.GCardID)
			cardAttr.SetStr("skin", c.Skin)
			cardAttr.SetStr("equip", c.Equip)
			fightCardsAttr.AppendMapAttr(cardAttr)
		}
	}

	if playerInfo.Name != "" && ri.getName() != playerInfo.Name {
		ri.setName(playerInfo.Name)
	}
	return
}

func (ri *rankItem) onSeasonPvpBegin() {
	ri.setLastSeasonRank(0)
	ri.setTodaySeasonRank(0)
	ri.attr.SetInt("winDiff", 0)
	ri.attr.SetInt("winCnt", 0)
}

func (ri *rankItem) onMultiLanSeasonPvpBegin(now int64) {
	curPvpScore := ri.attr.GetInt("pvpScore")
	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	curPvpLevel := rankGameData.GetPvpLevelByStar(curPvpScore)
	var resetLevel int
	var resetPvpScore int
	if curPvpLevel < 16 {
		return
	}

	/*
	ri.attr.SetInt64("updateTime", now)
	if curPvpLevel >= 31 {
		resetLevel = 21
	} else if curPvpLevel >= 28 {
		resetLevel = 20
	} else if curPvpLevel >= 25 {
		resetLevel = 19
	} else if curPvpLevel >= 22 {
		resetLevel = 18
	} else if curPvpLevel >= 19 {
		resetLevel = 17
	} else {
		resetLevel = 16
	}
	*/

	resetLevel = 16
	rankData, ok := rankGameData.Ranks[resetLevel - 1]
	if !ok {
		resetPvpScore = 1
	} else {
		resetPvpScore = rankData.LevelUpStar + 1
	}
	ri.attr.SetInt("pvpScore", resetPvpScore)
	glog.Infof("onSeasonPvpBegin uid=%d, curPvpLevel=%d, resetLevel=%d, resetPvpScore=%d", ri.getUid(),
		curPvpLevel, resetLevel, resetPvpScore)
}

func (ri *rankItem) onLeagueSeasonEnd() bool {
	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	team := rankGameData.GetTeamByStar(ri.getPvpScore())
	if team < 9 {
		return true
	}

	oldRankScore := ri.getRankScore()
	_, modifyScore := utils.CrossLeagueSeasonResetScore(oldRankScore, oldRankScore)
	ri.setRankScore(oldRankScore + modifyScore)
	glog.Infof("onLeagueSeasonEnd %d %d %d", ri.getUid(), oldRankScore, modifyScore)
	return true
}
