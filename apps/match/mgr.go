package main

import (
	"math/rand"
	"time"

	"kinger/common/aicardpool"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
)

const (
	matchTimeout       = time.Minute
	maxMatchRobotLevel = 10
)

var gMatchMgr *matchMgr

type iMatchRule interface {
	beginHeartBeat()
	addPlayer(mplayer *matchPlayer)
	delPlayer(uid common.UUid)
	isTargetPlayer(player *matchPlayer, streakLoseCnt int) bool
}

type matchMgr struct {
	maxRoomID           int32
	id2Room             map[int32]iMatchRoom
	uid2Room            map[common.UUid]iMatchRoom
	allRules            []iMatchRule
	newbieMatchingTimer map[common.UUid]*timer.Timer

	readyPlayerList     []*matchPlayer
	readyDonePlayerList []*matchPlayer
	uid2ReadyPlayer     map[common.UUid]*matchPlayer

	canMatchLevels map[int][]int // map[pvpLevel][]pvpLevel
}

func newMatchMgr() *matchMgr {
	m := &matchMgr{
		id2Room:             make(map[int32]iMatchRoom),
		uid2Room:            make(map[common.UUid]iMatchRoom),
		newbieMatchingTimer: make(map[common.UUid]*timer.Timer),
		canMatchLevels:      map[int][]int{},
		uid2ReadyPlayer:     map[common.UUid]*matchPlayer{},
	}

	m.allRules = []iMatchRule{newFoolRule(), newHighLevelRule()}

	gamedata.GetGameData(consts.Rank).AddReloadCallback(func(data gamedata.IGameData) {
		m.canMatchLevels = map[int][]int{}
	})

	return m
}

func (m *matchMgr) genRoomID() int32 {
	m.maxRoomID++
	return m.maxRoomID
}

func (m *matchMgr) addRoom(room iMatchRoom) {
	m.id2Room[room.getID()] = room
	p1 := room.getPlayerAgent1()
	if p1 != nil && !p1.IsRobot() {
		m.uid2Room[p1.GetUid()] = room
	}
	p2 := room.getPlayerAgent2()
	if p2 != nil && !p2.IsRobot() {
		m.uid2Room[p2.GetUid()] = room
	}
}

func (m *matchMgr) onMatchReadyDone(roomID int32) {
	if r, ok := m.id2Room[roomID]; ok {
		delete(m.id2Room, roomID)
		r.beginBattle()
		p1 := r.getPlayerAgent1()
		if p1 != nil && !p1.IsRobot() {
			delete(m.uid2Room, p1.GetUid())
		}
		p2 := r.getPlayerAgent2()
		if p2 != nil && !p2.IsRobot() {
			delete(m.uid2Room, p2.GetUid())
		}
	}
}

func (m *matchMgr) beginHeartbeat() {
	for _, r := range m.allRules {
		r.beginHeartBeat()
	}
	timer.AddTicker(200*time.Millisecond, m.onHeartBeat)
}

func (m *matchMgr) onMatchTimeout(mplayer iMatchPlayer) {
	mplayer.getAgent().PushClient(pb.MessageID_S2C_MATCH_TIMEOUT, nil)
}

func (m *matchMgr) onMatchLadder(mplayer1 iMatchPlayer, mplayer2 iMatchPlayer, indexDiff int) {
	//glog.Debugf("onMatchLadder 33333333333 %d %d", mplayer1.getUid(), mplayer1.getPvpLevel())
	r := newMatchRoom(m.genRoomID(), consts.BtPvp, 0, indexDiff)
	r.addPlayer(mplayer1)
	r.addPlayer(mplayer2)
	if mplayer1.getSeasonDataID() == mplayer2.getSeasonDataID() {
		r.seasonDataID = mplayer1.getSeasonDataID()
	}
	m.addRoom(r)
}

func (m *matchMgr) getCanMatchLevels(pvpLevel int) []int {
	if levels, ok := m.canMatchLevels[pvpLevel]; ok {
		return levels
	}

	var levels = []int{pvpLevel}
	rankData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData).Ranks[pvpLevel]
	if rankData != nil {
		isUpMax := false
		isDownMax := false
		diff := -1

		for !isUpMax || !isDownMax {
			level := pvpLevel + diff
			if diff > 0 {
				diff = -diff - 1
			} else {
				diff = -diff
			}

			if level > rankData.MatchUpper {
				isUpMax = true
				continue
			}

			if level < rankData.MatchLower {
				isDownMax = true
				continue
			}

			levels = append(levels, level)
		}
	}

	m.canMatchLevels[pvpLevel] = levels
	return levels
}

func (m *matchMgr) doMatchRobot(mPlayer iMatchPlayer, robot iMatchRobot, upperType int, isFirstPvp bool) {
	rmPlayer := newMatchPlayerByRobot(robot, mPlayer)
	rmPlayer.area = mPlayer.getArea()
	r := newMatchRoom(m.genRoomID(), consts.BtPvp, upperType, 0)
	r.isFirstPvp = isFirstPvp
	r.addPlayer(mPlayer)
	r.addPlayer(rmPlayer)
	m.addRoom(r)
}

func (m *matchMgr) doMatchNewbiePvp(agent *logic.PlayerAgent, matchArg *pb.BeginNewbiePvpMatchArg) {

	uid := agent.GetUid()
	if t, ok := m.newbieMatchingTimer[uid]; ok {
		t.Cancel()
	}

	upperType := 0
	if matchArg.IsFirstBattle {
		upperType = 1
	}

	m.newbieMatchingTimer[uid] = timer.AfterFunc(time.Duration(rand.Intn(3)+1)*time.Second, func() {
		delete(m.newbieMatchingTimer, uid)
		mPlayer := newMatchPlayer(agent, matchArg, nil)
		r := newNewbiePvpRobot(int(matchArg.EnemyCamp), int(matchArg.PvpScore), mPlayer.getPvpLevel(), matchArg.IsFirstBattle)
		m.doMatchRobot(mPlayer, r, upperType, matchArg.IsFirstBattle)
	})
}

func (m *matchMgr) onHeartBeat() {
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, player := range m.readyDonePlayerList {
		delete(m.uid2ReadyPlayer, player.getUid())
		for _, r := range m.allRules {
			if r.isTargetPlayer(player, player.getStreakLoseCnt()) {
				r.addPlayer(player)

				handCards := player.getHandCards()
				cardIDs := make([]uint32, len(handCards))
				for i, card := range handCards {
					cardData := poolGameData.GetCardByGid(card.GCardID)
					if cardData == nil {
						break
					}
					cardIDs[i] = cardData.CardID
				}

				aicardpool.AddCardPool(player.getPvpLevel(), player.getCamp(), cardIDs)
				break
			}
		}
	}
	m.readyDonePlayerList = m.readyPlayerList
	m.readyPlayerList = []*matchPlayer{}
}

func (m *matchMgr) beginMatch(agent *logic.PlayerAgent, matchArg *pb.BeginMatchArg) error {
	m.stopMatch(agent.GetUid())
	glog.Infof("matchMgr beginMatch %d", agent.GetUid())
	player := newMatchPlayer(agent, matchArg, matchArg)
	m.uid2ReadyPlayer[agent.GetUid()] = player
	m.readyPlayerList = append(m.readyPlayerList, player)
	return nil
}

func (m *matchMgr) onPlayerLogout(uid common.UUid) {
	delete(m.uid2Room, uid)
	glog.Infof("onPlayerLogout %d", uid)
	m.stopMatch(uid)
}

func (m *matchMgr) stopMatch(uid common.UUid) error {
	if player, ok := m.uid2ReadyPlayer[uid]; ok {
		delete(m.uid2ReadyPlayer, uid)
		for i, player2 := range m.readyPlayerList {
			if player == player2 {
				m.readyPlayerList = append(m.readyPlayerList[:i], m.readyPlayerList[i+1:]...)
				return nil
			}
		}

		for i, player2 := range m.readyDonePlayerList {
			if player == player2 {
				m.readyDonePlayerList = append(m.readyDonePlayerList[:i], m.readyDonePlayerList[i+1:]...)
				return nil
			}
		}
	}

	//glog.Infof("stopMatch 111111 %d", uid)
	if t, ok := m.newbieMatchingTimer[uid]; ok {
		t.Cancel()
		delete(m.newbieMatchingTimer, uid)
		return nil
	}

	if _, ok := m.uid2Room[uid]; ok {
		return gamedata.GameError(1)
	}

	glog.Infof("matchMgr stopMatch %d", uid)
	for _, r := range m.allRules {
		r.delPlayer(uid)
	}
	return nil
}
