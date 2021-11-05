package main

import (
	"kinger/gamedata"
	"kinger/common/consts"
	"time"
	"kinger/gopuppy/common"
	//"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
)

// <= maxMatchRobotLevel && (连输3场 || 胜率过低)

type foolRuleSt struct {
	matchAiData *gamedata.AiMatchGameData
	uid2Player  map[common.UUid]*matchPlayer
	uid2Robot map[common.UUid]iMatchRobot
}

func newFoolRule() iMatchRule {
	return &foolRuleSt{
		matchAiData: gamedata.GetGameData(consts.AiMatch).(*gamedata.AiMatchGameData),
		uid2Player:  map[common.UUid]*matchPlayer{},
		uid2Robot: map[common.UUid]iMatchRobot{},
	}
}

func (r *foolRuleSt) isTargetPlayer(player *matchPlayer, streakLoseCnt int) bool {
	pvpLevel := player.getPvpLevel()
	if pvpLevel >= 31 {
		// 王者
		return false
	}

	if !(streakLoseCnt >= 3 || player.getWinRate() <= r.matchAiData.FoolWinRate) {
		return false
	}

	if pvpLevel > maxMatchRobotLevel && player.getStreakWinCnt() >= 1 {
		return false
	}

	robot := robotMgr.getRobot(player)
	if robot == nil {
		return false
	}

	r.uid2Robot[player.getUid()] = robot
	return true
}

func (r *foolRuleSt) beginHeartBeat() {
	timer.AddTicker(time.Second, func() {

		now := time.Now()
		for uid, player := range r.uid2Player {
			robot := r.uid2Robot[uid]
			if robot == nil {
				robot = robotMgr.getRobot(player)
			}

			if robot == nil {
				continue
			}

			gMatchMgr.doMatchRobot(player, robot, 0, false)
			delete(r.uid2Player, uid)
			delete(r.uid2Robot, uid)
		}

		for uid, player := range r.uid2Player {
			if player.isMatchTimout(now) {
				gMatchMgr.onMatchTimeout(player)
				delete(r.uid2Player, uid)
				delete(r.uid2Robot, uid)
			}
		}

	})
}

func (r *foolRuleSt) addPlayer(mplayer *matchPlayer) {
	//glog.Infof("foolRuleSt addPlayer uid=%d", mplayer.getUid())
	r.uid2Player[mplayer.getUid()] = mplayer
}

func (r *foolRuleSt) delPlayer(uid common.UUid) {
	//glog.Infof("foolRuleSt delPlayer uid=%d", uid)
	delete(r.uid2Player, uid)
	delete(r.uid2Robot, uid)
}
