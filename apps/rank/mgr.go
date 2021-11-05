package main

import (
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/apps/logic"
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"time"
)

var rankMgr *rankMgrSt

const (
	attrTodayRankItem = "todayRankItem"
	attrCurRankItem = "curRankItem"
	attrTodayRank = "pvpRank"
	attrCurRank = "curPvpRank"
	attrTodaySeasonRank = "seasonRank"
	attrCurSeasonRank = "curSeasonRank"
	attrTodayCrossAreaRank = "crossAreaRank"
	attrCurCrossAreaRank = "curCrossAreaRank"
)

type rankMgrSt struct {
	todayPlayers map[common.UUid]*rankItem  // 今天所有榜的玩家数据
	curPlayers map[common.UUid]*rankItem    // 当前所有榜的玩家数据
	totalBoard *rankBoard                   // 全服总榜
	type2Boards map[pb.RankType]map[int]*rankBoard  // 分区天梯、分区锦标赛、分区跨区荣誉
}

func newRankMgr() {
	rankMgr = &rankMgrSt{
		todayPlayers: map[common.UUid]*rankItem{},
		curPlayers: map[common.UUid]*rankItem{},
		type2Boards: map[pb.RankType]map[int]*rankBoard{},
	}

	timer.AddTicker(15*time.Minute, func() {
		//rankMgr.refreshCurRankList(false)
		//rankMgr.refreshTodayRankList()
		rankMgr.save()
	})
	timer.RunEveryHour(0, 0, rankMgr.refreshTodayRankList)
	timer.RunEveryHour(30, 1, rankMgr.refreshTodayRankList)
}

func (rm *rankMgrSt) getBoard(rankType pb.RankType, area int) *rankBoard {
	if rankType == pb.RankType_RtLadder && area == 0 {
		return rm.totalBoard
	}

	area2broad, ok := rm.type2Boards[rankType]
	if !ok {
		return nil
	}
	return area2broad[area]
}

func (rm *rankMgrSt) getBoardUsersByArea(rankType pb.RankType, maxRank int, area int)  *pb.Area2UserRanking{
	area2Board, ok := rm.type2Boards[rankType]
	if !ok {
		return nil
	}

	msg := &pb.Area2UserRanking{}
	for area2, board := range area2Board {
		if area > 0 && area != area2 {
			continue
		}

		msg.Areas = append(msg.Areas, int32(area))
		msg2 := &pb.UserRankingInfo{}
		msg2.Uids = board.getCurRankList(maxRank)
		msg.UserRanking = append(msg.UserRanking, msg2)
	}
	return msg
}

func (rm *rankMgrSt) loadRankItem(attrName string, uid common.UUid) *rankItem {
	attr := attribute.NewAttrMgr(attrName, uid)
	if err := attr.Load(); err != nil {
		glog.Errorf("%s load player %d error %s", attrName, uid, err)
		return nil
	}

	return newRankItemByAttr(uid, attr)
}

func (rm *rankMgrSt) loadTodayRankItem(uid common.UUid) *rankItem {
	if ri, ok := rm.todayPlayers[uid]; ok {
		return ri
	} else {
		ri = rm.loadRankItem(attrTodayRankItem, uid)
		if ri != nil {
			rm.todayPlayers[uid] = ri
		}
		return ri
	}
}

func (rm *rankMgrSt) getTodayRankItem(uid common.UUid) *rankItem {
	return rm.todayPlayers[uid]
}

func (rm *rankMgrSt) addTodayRankItem(ri *rankItem) {
	rm.todayPlayers[ri.getUid()] = ri
}

func (rm *rankMgrSt) loadCurRankItem(uid common.UUid) *rankItem {
	if ri, ok := rm.curPlayers[uid]; ok {
		return ri
	} else {
		ri = rm.loadRankItem(attrCurRankItem, uid)
		if ri != nil {
			rm.curPlayers[uid] = ri
		}
		return ri
	}
}

func (rm *rankMgrSt) loadAreaBoard(gdata gamedata.IGameData) {
	areaGameData := gdata.(*gamedata.AreaConfigGameData)
	areaGameData.ForEachOpenedArea(func(config *gamedata.AreaConfig) {

		for _, rankType := range []pb.RankType{pb.RankType_RtLadder, pb.RankType_RtSeason,
			pb.RankType_RtCrossArea} {

			areaToBoard, ok := rm.type2Boards[rankType]
			if !ok {
				areaToBoard = map[int]*rankBoard{}
				rm.type2Boards[rankType] = areaToBoard
			}

			area := config.Area
			if _, ok := areaToBoard[area]; ok {
				return
			}

			areaToBoard[area] = newRankBoard(rankType)
			areaToBoard[area].load(area)
		}
	})
}

func (rm *rankMgrSt) loadBoard() {
	rm.totalBoard = newRankBoard(pb.RankType_RtLadder)
	rm.totalBoard.load(0)
	areaGameData := gamedata.GetGameData(consts.AreaConfig).(*gamedata.AreaConfigGameData)
	areaGameData.AddReloadCallback(rm.loadAreaBoard)
	rm.loadAreaBoard(areaGameData)
}

func (rm *rankMgrSt) refreshCurRankList(needReply bool) {
	var allCurPlayers []common.UInt64Set
	allCurPlayers = append(allCurPlayers, rm.totalBoard.refreshCurRankList(needReply))

	for _, areaToBoard := range rm.type2Boards {
		for _, board := range areaToBoard {
			allCurPlayers = append(allCurPlayers, board.refreshCurRankList(false))
		}
	}


L:	for uid, ri := range rm.curPlayers {
		uid2 := uint64(uid)
		for _, players := range allCurPlayers {
			if players.Contains(uid2) {
				continue L
			}
		}

		ri.attr.Delete(false)
		delete(rm.curPlayers, uid)
	}
}

func (rm *rankMgrSt) refreshTodayRankList() {
	rm.refreshCurRankList(false)

	var allTodayPlayers []common.UInt64Set
	allTodayPlayers = append(allTodayPlayers, rm.totalBoard.refreshTodayRankList())

	for _, areaToBoard := range rm.type2Boards {
		for _, board := range areaToBoard {
			allTodayPlayers = append(allTodayPlayers, board.refreshTodayRankList())
		}
	}

L:	for uid, ri := range rm.todayPlayers {
		uid2 := uint64(uid)
		for _, players := range allTodayPlayers {
			if players.Contains(uid2) {
				continue L
			}
		}

		ri.attr.Delete(false)
		delete(rm.curPlayers, uid)
	}

	api.BroadcastClient(pb.MessageID_S2C_REFRESH_RANK, nil, nil)

	msg := rm.getBoardUsersByArea(pb.RankType_RtLadder, 3, 0)
	logic.PushBackend(gconsts.AppGame, 1, pb.MessageID_R2G_SEND_PLAYER_RANK, msg)
}

func (rm *rankMgrSt) onSeasonPvpBegin(area int) {
	board := rm.getBoard(pb.RankType_RtSeason, area)
	if board == nil {
		glog.Errorf("onSeasonPvpBegin no board area=%d", area)
		return
	}
	board.onSeasonPvpBegin()
}

func (rm *rankMgrSt) updateRankScore(playerInfo *pb.UpdatePvpScoreArg) {

	uid := common.UUid(playerInfo.Uid)
	area := int(playerInfo.Area)
	ri, ok := rm.curPlayers[uid]
	if !ok {
		ri = newRankItem(attrCurRankItem, uid, playerInfo.Name, area)
		rm.curPlayers[uid] = ri
	}

	updateRankTypes := ri.update(playerInfo)
	updateRankTypes.ForEach(func(rankType int) {
		if pb.RankType(rankType) == pb.RankType_RtLadder {
			rm.totalBoard.onRankItemUpdate(ri)
		}
		board := rm.getBoard(pb.RankType(rankType), area)
		if board != nil {
			board.onRankItemUpdate(ri)
		}
	})

	ri.attr.Save(false)
}

func (rm *rankMgrSt) save() {
	rm.totalBoard.save()
	for _, areaToBoard := range rm.type2Boards {
		for _, board := range areaToBoard {
			board.save()
		}
	}
}
