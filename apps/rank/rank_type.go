package main

import (
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
)

type iRankType interface {
	getRank(ri *rankItem, isTotalBoard bool) int
	getMsgData(ri *rankItem, msg *pb.RankItem, isTotalBoard bool) *pb.RankItem
	getMaxRank() int
	setUpdateTime(ri *rankItem, t int64)
	rankLess(rankList *rankList, i, j int) bool
	getRankAttrName(isToday bool) string
	setRankItemRank(rb *rankBoard, ri *rankItem, curRank, lastRank int)
	clear(rb *rankBoard, ri *rankItem)
}

type rtLadderst struct{}

type rtSeason struct{}

type rtCrossArea struct{}

func getIRankType(rankType pb.RankType) iRankType {
	var r iRankType
	switch rankType {
	case pb.RankType_RtLadder:
		r = &rtLadderst{}
	case pb.RankType_RtSeason:
		r = &rtSeason{}
	case pb.RankType_RtCrossArea:
		r = &rtCrossArea{}

	default:
		r = nil
	}
	return r
}

func (rl *rtLadderst) getRank(ri *rankItem, isTotalBoard bool) int {
	if isTotalBoard {
		return ri.getTodayTotalRank()
	} else {
		return ri.getTodayRank()
	}
}

func (rl *rtLadderst) clear(rb *rankBoard, ri *rankItem) {}

func (rl *rtLadderst) getMsgData(ri *rankItem, msg *pb.RankItem, isTotalBoard bool) *pb.RankItem {
	if isTotalBoard {
		msg.Rank = int32(ri.getTodayTotalRank())
		msg.LastRank = int32(ri.getLastTotalRank())
	} else {
		msg.Rank = int32(ri.getTodayRank())
		msg.LastRank = int32(ri.getLastRank())
	}
	return msg
}

func (rl *rtLadderst) getMaxRank() int {
	return 200
}

func (rl *rtLadderst) setUpdateTime(ri *rankItem, t int64) {
	ri.setLadderUpdateTime(t)
}

func (rl *rtLadderst) rankLess(rankList *rankList, i, j int) bool {
	it1 := rankList.items[i]
	it2 := rankList.items[j]
	rebornCnt1 := it1.getRebornCnt()
	rebornCnt2 := it2.getRebornCnt()
	if rebornCnt1 > rebornCnt2 {
		return true
	} else if rebornCnt1 < rebornCnt2 {
		return false
	}

	score1 := it1.getPvpScore()
	score2 := it2.getPvpScore()
	rgd := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	team1 := rgd.GetTeamByStar(score1)
	team2 := rgd.GetTeamByStar(score2)
	if team1 >= 9 && team2 >= 9 {
		score1 = it1.getRankScore()
		score2 = it2.getRankScore()
	}
	if score1 > score2 {
		return true
	} else if score1 < score2 {
		return false
	}

	updateTime1 := it1.getUpdateTime()
	updateTime2 := it2.getUpdateTime()
	return updateTime1 >= updateTime2
}

func (rl *rtLadderst) getRankAttrName(isToday bool) string {
	if isToday {
		return attrTodayRank
	} else {
		return attrCurRank
	}
}

func (rl *rtLadderst) setRankItemRank(rb *rankBoard, ri *rankItem, curRank, lastRank int) {
	if rb.isTotalBoard() {
		ri.setTodayTotalRank(curRank)
		ri.setLastTotalRank(lastRank)
	} else {
		ri.setTodayRank(curRank)
		ri.setLastRank(lastRank)
	}
}

func (rs *rtSeason) clear(rb *rankBoard, ri *rankItem) {}

func (rs *rtSeason) getRank(ri *rankItem, isTotalBoard bool) int {
	return ri.getTodaySeasonRank()
}

func (rs *rtSeason) getMsgData(ri *rankItem, msg *pb.RankItem, isTotalBoard bool) *pb.RankItem {
	msg.Rank = int32(ri.getTodaySeasonRank())
	msg.LastRank = int32(ri.getLastSeasonRank())
	return msg
}

func (rs *rtSeason) getMaxRank() int {
	return 200
}

func (rs *rtSeason) setUpdateTime(ri *rankItem, t int64) {
	ri.setSeasonUpdateTime(t)
}

func (rs *rtSeason) rankLess(rankList *rankList, i, j int) bool {
	it1 := rankList.items[i]
	it2 := rankList.items[j]
	winDiff1 := it1.getSeasonWinDiff()
	winDiff2 := it2.getSeasonWinDiff()
	if winDiff1 > winDiff2 {
		return true
	} else if winDiff1 < winDiff2 {
		return false
	}

	winCnt1 := it1.getSeasonWinCnt()
	winCnt2 := it2.getSeasonWinCnt()
	if winCnt1 > winCnt2 {
		return true
	} else if winCnt1 < winCnt2 {
		return false
	}

	return it1.getSeasonUpdateTime() >= it2.getSeasonUpdateTime()
}

func (rs *rtSeason) getRankAttrName(isToday bool) string {
	if isToday {
		return attrTodaySeasonRank
	} else {
		return attrCurSeasonRank
	}
}

func (rs *rtSeason) setRankItemRank(rb *rankBoard, ri *rankItem, curRank, lastRank int) {
	ri.setTodaySeasonRank(curRank)
	ri.setLastSeasonRank(lastRank)
}

func (rc *rtCrossArea) getRank(ri *rankItem, isTotalBoard bool) int {
	return ri.getTodayCrossAreaRank()
}

func (rc *rtCrossArea) getMsgData(ri *rankItem, msg *pb.RankItem, isTotalBoard bool) *pb.RankItem {
	msg.Rank = int32(ri.getTodayCrossAreaRank())
	msg.LastRank = int32(ri.getLastCrossAreaRank())
	return msg
}

func (rc *rtCrossArea) getMaxRank() int {
	return 50
}

func (rc *rtCrossArea) setUpdateTime(ri *rankItem, t int64) {
	ri.setCrossAreaUpdateTime(t)
}

func (rc *rtCrossArea) rankLess(rankList *rankList, i, j int) bool {
	it1 := rankList.items[i]
	it2 := rankList.items[j]
	crossAreaHonor1 := it1.getCrossAreaHonor()
	crossAreaHonor2 := it2.getCrossAreaHonor()

	if crossAreaHonor1 > 0 || crossAreaHonor2 > 0 {
		if crossAreaHonor1 > crossAreaHonor2 {
			return true
		} else if crossAreaHonor1 < crossAreaHonor2 {
			return false
		}
	} else if crossAreaHonor1 < 0 && crossAreaHonor2 < 0 {
		if crossAreaHonor1 > crossAreaHonor2 {
			return false
		} else if crossAreaHonor1 < crossAreaHonor2 {
			return true
		}
	}

	return it1.getCrossAreaUpdateTime() >= it2.getCrossAreaUpdateTime()
}

func (rc *rtCrossArea) getRankAttrName(isToday bool) string {
	if isToday {
		return attrTodayCrossAreaRank
	} else {
		return attrCurCrossAreaRank
	}
}

func (rc *rtCrossArea) setRankItemRank(rb *rankBoard, ri *rankItem, curRank, lastRank int) {
	ri.setTodayCrossAreaRank(curRank)
	ri.setLastCrossAreaRank(curRank)
}

func (rc *rtCrossArea) clear(rb *rankBoard, ri *rankItem) {
	rc.setRankItemRank(rb, ri, 0, 0)
	ri.setCrossAreaHonor(0)
}
