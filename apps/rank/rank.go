package main

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"sort"
	"time"
)

type rankBoard struct {
	msg      *pb.RankInfo
	rankType pb.RankType
	inSeason bool
	area     int

	todayRankAttr *attribute.AttrMgr
	todayRankList *rankList

	curRankAttr *attribute.AttrMgr
	curRankList *rankList
}

func newRankBoard(rankType pb.RankType) *rankBoard {
	return &rankBoard{
		rankType:      rankType,
		todayRankList: newRankList(rankType),
		curRankList:   newRankList(rankType),
	}
}

func (rb *rankBoard) onSeasonPvpBegin() {
	if !rb.isSeason() {
		return
	}

	rb.inSeason = true
	rb.todayRankAttr.SetBool("isInSeason", true)
	rb.curRankList.forEach(func(i int, ri *rankItem) bool {
		ri.onSeasonPvpBegin()
		return true
	})

	now := int32(time.Now().Unix())
	rb.todayRankList = newRankList(rb.rankType)
	rankAttr := attribute.NewListAttr()
	rb.todayRankAttr.SetListAttr("data", rankAttr)
	rb.todayRankAttr.SetInt32("updateTime", now)

	rb.curRankList = newRankList(rb.rankType)
	rankAttr = attribute.NewListAttr()
	rb.curRankAttr.SetListAttr("data", rankAttr)
	rb.curRankAttr.SetInt32("updateTime", now)
	glog.Infof("onSeasonPvpBegin area=%d", rb.area)
}

func (rb *rankBoard) onSeasonPvpEnd() {
	if !rb.isSeason() {
		return
	}
	rb.inSeason = false
	rb.todayRankAttr.SetBool("isInSeason", false)
	glog.Infof("onSeasonPvpEnd area=%d", rb.area)
}

func (rb *rankBoard) onLeagueSeasonEnd() {
	rb.curRankList.forEach(func(i int, ri *rankItem) bool {
		return ri.onLeagueSeasonEnd()
	})
}

func (rb *rankBoard) save() {
	rb.todayRankAttr.Save(false)
	rb.curRankAttr.Save(true)
}

func (rb *rankBoard) getUpdateTime() int32 {
	updateTime := rb.todayRankAttr.GetInt32("updateTime")
	if updateTime > 0 {
		updateTime = int32(time.Now().Unix()) - updateTime
	}
	if updateTime < 0 {
		updateTime = 0
	}
	return updateTime
}

func (rb *rankBoard) packMsg() *pb.RankInfo {
	updateTime := rb.getUpdateTime()
	if rb.msg != nil {
		rb.msg.UpdateTime = updateTime
		return rb.msg
	}

	msg := &pb.RankInfo{
		UpdateTime: updateTime,
	}

	iRt := getIRankType(rb.rankType)
	if iRt == nil {
		return nil
	}

	rb.todayRankList.forEach(func(i int, ri *rankItem) bool {
		msg.Items = append(msg.Items, ri.packItemMsg(iRt, rb.isTotalBoard()))
		return true
	})

	rb.msg = msg
	return msg
}

func (rb *rankBoard) getRankAttrName(isToday bool) string {
	iRt := getIRankType(rb.rankType)
	if iRt == nil {
		return ""
	}
	return iRt.getRankAttrName(isToday)
}

func (rb *rankBoard) loadRankList(isToday bool) {
	attrName := rb.getRankAttrName(isToday)
	rankAttrMgr := attribute.NewAttrMgr(attrName, rb.area)
	if err := rankAttrMgr.Load(); err != nil && err != attribute.NotExistsErr {
		panic(err)
	}

	rankAttr := rankAttrMgr.GetListAttr("data")
	if rankAttr == nil {
		rankAttr = attribute.NewListAttr()
		rankAttrMgr.SetListAttr("data", rankAttr)
		rankAttrMgr.SetInt32("updateTime", int32(time.Now().Unix()))
	}

	if isToday {
		rb.todayRankAttr = rankAttrMgr
	} else {
		rb.curRankAttr = rankAttrMgr
	}

	rankAttr.ForEachIndex(func(index int) bool {
		uid := common.UUid(rankAttr.GetUInt64(index))
		var ri *rankItem
		if isToday {
			ri = rankMgr.loadTodayRankItem(uid)
		} else {
			ri = rankMgr.loadCurRankItem(uid)
		}
		if ri == nil {
			glog.Errorf("loadRankList loadRankItem error, attrName=%s", attrName)
			return true
		}

		if isToday {
			rb.todayRankList.append(ri)
		} else {
			rb.curRankList.append(ri)
		}
		return true
	})
}

func (rb *rankBoard) isSeason() bool {
	return rb.rankType == pb.RankType_RtSeason
}

func (rb *rankBoard) load(area int) {
	rb.area = area
	rb.loadRankList(true)
	rb.loadRankList(false)
	if rb.isSeason() {
		rb.inSeason = rb.todayRankAttr.GetBool("isInSeason")
	}
}

func (rb *rankBoard) refreshCurRankList(needReply bool) common.UInt64Set {
	if rb.isSeason() && !rb.inSeason {
		return rb.getRankPlayersSet(rb.curRankList)
	}

	sort.Sort(rb.curRankList)
	rb.curRankList.fixSize()

	curPlayersSet := common.UInt64Set{}
	var curPlayers []uint64
	attr := attribute.NewListAttr()
	rb.curRankList.forEach(func(_ int, ri *rankItem) bool {
		uid := uint64(ri.getUid())
		attr.AppendUInt64(uid)
		curPlayersSet.Add(uid)
		curPlayers = append(curPlayers, uid)
		return true
	})

	rb.curRankAttr.SetListAttr("data", attr)
	rb.curRankAttr.Save(needReply)

	glog.Infof("refreshCurRankList rankType=%s, area=%d, curPlayers=%v", rb.rankType, rb.area, curPlayers)
	return curPlayersSet
}

func (rb *rankBoard) isTotalBoard() bool {
	return rb.area == 0
}

func (rb *rankBoard) getRankPlayersSet(rl *rankList) common.UInt64Set {
	players := common.UInt64Set{}
	rl.forEach(func(i int, ri *rankItem) bool {
		players.Add(uint64(ri.getUid()))
		return true
	})
	return players
}

func (rb *rankBoard) getRankItemTodayRank(ri *rankItem) int {
	iRt := getIRankType(rb.rankType)
	if iRt == nil {
		return 0
	}
	return iRt.getRank(ri, rb.isTotalBoard())
}

func (rb *rankBoard) setRankItemRank(ri *rankItem, curRank, lastRank int) {
	getIRankType(rb.rankType).setRankItemRank(rb, ri, curRank, lastRank)
}

func (rb *rankBoard) refreshTodayRankList() common.UInt64Set {
	var todayPlayers []common.UUid
	todayPlayersSet := common.UInt64Set{}
	nowRankList := newRankList(rb.rankType)
	attr := attribute.NewListAttr()
	rb.curRankList.forEach(func(i int, ri *rankItem) bool {
		uid := ri.getUid()
		attr.AppendUInt64(uint64(uid))
		todayPlayers = append(todayPlayers, uid)
		todayPlayersSet.Add(uint64(uid))

		todayPlayer := rankMgr.getTodayRankItem(common.UUid(uid))
		var lastRank int
		if todayPlayer == nil {

			todayPlayer = newRankItem(attrTodayRankItem, ri.getUid(), ri.getName(), ri.getArea())
			rankMgr.addTodayRankItem(todayPlayer)

		} else {
			lastRank = rb.getRankItemTodayRank(todayPlayer)
		}

		todayPlayer.update(&pb.UpdatePvpScoreArg{
			PvpScore:       int32(ri.getPvpScore()),
			Camp:           int32(ri.getCamp()),
			HandCards:      ri.getFightCards(),
			Name:           ri.getName(),
			WinDiff:        int32(ri.getSeasonWinDiff()),
			WinCnt:         int32(ri.getSeasonWinCnt()),
			RebornCnt:      int32(ri.getRebornCnt()),
			CrossAreaHonor: int32(ri.getCrossAreaHonor()),
			RankScore:      int32(ri.getRankScore()),
			//CrossAreaBlotHonor: int32(ri.getCrossAreaBlotHonor()),
		})

		rb.setRankItemRank(todayPlayer, i+1, lastRank)

		todayPlayer.attr.Save(false)
		nowRankList.append(todayPlayer)
		return true
	})

	rb.todayRankList = nowRankList
	rb.todayRankAttr.SetListAttr("data", attr)
	//rb.onRefresh()
	rb.todayRankAttr.SetInt32("updateTime", int32(time.Now().Unix()))
	rb.todayRankAttr.Save(false)
	rb.msg = nil

	glog.Infof("refreshTodayRankList rankType=%s, area=%d, todayPlayers=%v", rb.rankType, rb.area, todayPlayers)
	return todayPlayersSet
}

func (rb *rankBoard) clear() {
	rankType := getIRankType(rb.rankType)
	rb.todayRankList.forEach(func(i int, ri *rankItem) bool {
		rankType.clear(rb, ri)
		return true
	})

	rb.todayRankList = newRankList(rb.rankType)
	attr := attribute.NewListAttr()
	rb.todayRankAttr.SetListAttr("data", attr)
	rb.todayRankAttr.SetInt32("updateTime", int32(time.Now().Unix()))
	rb.todayRankAttr.Save(false)
	rb.msg = nil

	rb.curRankList = newRankList(rb.rankType)
	rb.curRankAttr.SetListAttr("data", attribute.NewListAttr())
	rb.curRankAttr.SetInt32("updateTime", int32(time.Now().Unix()))
	rb.curRankAttr.Save(false)
}

func (rb *rankBoard) onRefresh() {
	weekDay := int(time.Now().Weekday())
	if weekDay == 1 && rb.rankType == pb.RankType_RtCrossArea {
		updateTime := rb.todayRankAttr.GetInt32("updateTime")
		updateDay := timer.GetDayNo(int64(updateTime))
		now := time.Now().Unix()
		curDay := timer.GetDayNo(now)

		if subDay := curDay - updateDay; subDay > 0 {
			rb.seedReward()
			rb.clear()
		}
	}
}

func (rb *rankBoard) seedReward() {
	var uidList []uint64
	var honorList []int32

	for _, item := range rb.curRankList.items {
		if item.getCrossAreaHonor() < 0 {
			break
		}
		uidList = append(uidList, uint64(item.getUid()))
		honorList = append(honorList, int32(item.getCrossAreaHonor()))
	}

	logic.PushBackend(gconsts.AppGame, 1, pb.MessageID_R2G_SEND_RANK_HONOR, &pb.RankHonorInfo{
		Uids:   uidList,
		Honors: honorList,
	})
}

func (rb *rankBoard) getCurRankList(maxRank int) []uint64 {
	var rankUids []uint64
	rb.curRankList.forEach(func(i int, ri *rankItem) bool {
		if i+1 > maxRank {
			return false
		}
		rankUids = append(rankUids, uint64(ri.getUid()))
		return true
	})
	return rankUids
}

func (rb *rankBoard) onRankItemUpdate(ri *rankItem) {
	if rb.isSeason() && !rb.inSeason {
		return
	}
	rb.curRankList.append(ri)
}
