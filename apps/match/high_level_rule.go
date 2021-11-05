package main

import (
	"time"
	"kinger/gopuppy/common"
	"container/list"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/glog"
	"kinger/gamedata"
	"kinger/common/consts"
	"math"
	"math/rand"
	"kinger/common/config"
)

type highLevelPlayer struct {
	*matchPlayer
	rankGameData *gamedata.Rank
	index float64                  // 匹配指数
	crossAreaIndex float64         // 跨区的匹配指数
	maxIndexDiff float64           // 同区能匹配相差多少个指数区间的玩家
	crossAreaMaxIndexDiff float64  // 跨区能匹配相差多少个指数区间的玩家
}

func newHighLevelPlayer(mp *matchPlayer, param *gamedata.MatchParamGameData) *highLevelPlayer {
	return &highLevelPlayer{
		matchPlayer: mp,
		index: math.MaxFloat64,
		crossAreaIndex: math.MaxFloat64,
		maxIndexDiff: param.IndexInterval,
	}
}

func (p *highLevelPlayer) getStreakWinOrLoseIndex(param *gamedata.MatchParamGameData) float64 {
	streakWinCnt := p.getStreakWinCnt()
	if streakWinCnt > 1 {
		streakWinCnt -= 1
		if streakWinCnt > 10 {
			streakWinCnt = 10
		}
		return param.WinningStreakIndex * float64(streakWinCnt)
	}

	streakLoseCnt := p.getStreakLoseCnt()
	if streakLoseCnt > 1 {
		streakLoseCnt -= 1
		if streakLoseCnt > 10 {
			streakLoseCnt = 10
		}
		return - param.WinningStreakIndex * float64(streakLoseCnt)
	}
	return 0
}

func (p *highLevelPlayer) getIndex(param *gamedata.MatchParamGameData) float64 {
	if p.index == math.MaxFloat64 {
		/*
		pvpScore := float64(p.getPvpScore())
		if pvpScore > 100 {
			pvpScore = 100
		}

		p.index = float64(p.getCardStrength()) * param.CardStreWeight + pvpScore * param.StarWeight +
			float64(p.getWinRate()) / 100 * param.WinRateWeight + float64(p.getRebornCnt()) * param.RebornWeight +
			float64(p.getEquipAmount()) * param.EquipWeight + p.getStreakWinOrLoseIndex(param) +
			float64(p.getRechargeMatchIndex())
		*/
		matchScore := float64(p.getMatchScore())
		p.index = matchScore + p.getStreakWinOrLoseIndex(param) + float64(p.getRechargeMatchIndex())
		if p.index < 0 {
			p.index = 0
		}
	}
	return p.index
}

func (p *highLevelPlayer) getCrossAreaIndex(param *gamedata.MatchParamGameData) float64 {
	if p.crossAreaIndex == math.MaxFloat64 {
		p.crossAreaIndex = float64(p.getCardStrength()) * param.CardStreWeight + float64(p.getWinRate()) /
			100 * param.WinRateWeight + float64(p.getRebornCnt()) * param.RebornWeight +
			float64(p.getEquipAmount()) * param.EquipWeight  + p.getStreakWinOrLoseIndex(param) +
			float64(p.getRechargeMatchIndex())
	}
	return p.crossAreaIndex
}

func (p *highLevelPlayer) getMaxIndexDiff(isCrossArea bool) float64 {
	if isCrossArea {
		return p.crossAreaMaxIndexDiff
	} else {
		return p.maxIndexDiff
	}
}

func (p *highLevelPlayer) onTick() {
	if p.rankGameData == nil {
		p.rankGameData = gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData).Ranks[p.getPvpLevel()]
	}
	if p.rankGameData == nil {
		return
	}

	p.maxIndexDiff += p.rankGameData.IndexInterval
	//p.crossAreaMaxIndexDiff += param.IndexInterval
	if p.maxIndexDiff > p.rankGameData.IndexRedline {
		p.maxIndexDiff = float64(p.rankGameData.IndexRedline)
	}
	//if p.crossAreaMaxIndexDiff > param.CrossAreaIndexRedLine {
	//	p.crossAreaMaxIndexDiff = param.CrossAreaIndexRedLine
	//}
}

func (p *highLevelPlayer) isOldServerArea() bool {
	return p.getArea() <= 3 && config.GetConfig().IsOldServer()
}

func (p *highLevelPlayer) canPair(opp *highLevelPlayer, param *gamedata.MatchParamGameData,
	seasonGameData gamedata.ISeasonPvpGameData) (can bool, indexDiffSec int, rawIndexDiff float64) {

	if p.getUid() == opp.getUid() {
		return false, 0, 0
	}

	isCrossArea := p.getArea() != opp.getArea()
	if isCrossArea && config.GetConfig().IsOldXfServer() && (p.getArea() == 1 || opp.getArea() == 1) {
		return false, 0, 0
	}

	if !seasonGameData.IsSeasonEqual(p.getSeasonDataID(), opp.getSeasonDataID()) {
		return false, 0, 0
	}

	myIndex := p.getIndex(param)
	oppIndex := opp.getIndex(param)

	rawIndexDiff = myIndex - oppIndex
	var indexDiff float64
	if rawIndexDiff < 0 {
		indexDiff = - rawIndexDiff
	} else {
		indexDiff = rawIndexDiff
	}

	if isCrossArea {
		indexDiff += param.AreaRevise
	}
	if p.getLastOppUid() == opp.getUid() || opp.getLastOppUid() == p.getUid() {
		indexDiff += param.RecentlyOpponentIndex
	}

	//if rebornDiff > param.SameAreaReborn || rebornDiff < - param.SameAreaReborn {
	//	return false, 0, indexDiff
	//}

	canMatchLevels := gMatchMgr.getCanMatchLevels(p.getPvpLevel())
	levelCanMatch := false
	oppPvpLevel := opp.getPvpLevel()
	for _, level := range canMatchLevels {
		if level == oppPvpLevel {
			levelCanMatch = true
			break
		}
	}

	if !levelCanMatch {
		return false, 0, rawIndexDiff
	}

	if indexDiff > p.getMaxIndexDiff(false) || indexDiff > opp.getMaxIndexDiff(false) {
		return false, 0, rawIndexDiff
	}

	return true, int(indexDiff / param.IndexInterval), rawIndexDiff
}


type highLevelRuleSt struct {
	allPlayers map[common.UUid]*list.Element
	playerList *list.List

	param *gamedata.MatchParamGameData
	seasonGameData gamedata.ISeasonPvpGameData

	ticker *timer.Timer
	tickerInterval time.Duration
}

func newHighLevelRule() iMatchRule {
	return &highLevelRuleSt{
		allPlayers: map[common.UUid]*list.Element{},
		playerList: list.New(),
	}
}

func (r *highLevelRuleSt) delMatchingElem(pElem *list.Element) *list.Element {
	mp := pElem.Value.(*highLevelPlayer)
	delete(r.allPlayers, mp.getUid())
	curElem := pElem
	pElem = pElem.Next()
	r.playerList.Remove(curElem)
	return pElem
}

func (r *highLevelRuleSt) beginTick(timeInterval time.Duration) {
	if r.ticker != nil {
		r.ticker.Cancel()
	}

	r.tickerInterval = timeInterval
	r.ticker = timer.AddTicker(timeInterval, func() {

		now := time.Now()
		for pElem := r.playerList.Front(); pElem != nil; {
			mp := pElem.Value.(*highLevelPlayer)
			sec2Opps := map[int][]*list.Element{}
			minSec := math.MaxInt64
			oppUid2IndexDiff := map[common.UUid]int{}

			for oppElem := pElem.Next(); oppElem != nil; oppElem = oppElem.Next() {
				opp := oppElem.Value.(*highLevelPlayer)
				canPair, diffSec, rawIndexDiff := mp.canPair(opp, r.param, r.seasonGameData)
				oppUid2IndexDiff[opp.getUid()] = int(rawIndexDiff)

				if !canPair {
					continue
				}
				if diffSec > minSec {
					continue
				}

				minSec = diffSec
				opps := sec2Opps[diffSec]
				sec2Opps[diffSec] = append(opps, oppElem)
			}

			opps := sec2Opps[minSec]
			var oppElem *list.Element
			if len(opps) > 0 {
				oppElem = opps[ rand.Intn(len(opps)) ]
				opp := oppElem.Value.(*highLevelPlayer)
				gMatchMgr.onMatchLadder(mp, opp, oppUid2IndexDiff[opp.getUid()])

				delete(r.allPlayers, opp.getUid())
				r.playerList.Remove(oppElem)

				pElem = r.delMatchingElem(pElem)
				glog.Infof("highLevelRuleSt onMatchLadder %d %d", mp.getUid(), opp.getUid())

			} else {

				if r.timeToMatchRobot(mp, now) {
					robot := robotMgr.getRobot(mp)
					if robot != nil {
						gMatchMgr.doMatchRobot(mp, robot, 0, false)
						pElem = r.delMatchingElem(pElem)
						continue
					}
				}

				if mp.isMatchTimout(now) {
					pElem = r.delMatchingElem(pElem)
					gMatchMgr.onMatchTimeout(mp)
					glog.Infof("highLevelRuleSt match timeout %d", mp.getUid())

				} else {
					mp.onTick()
					pElem = pElem.Next()
				}

			}
		}

	})
}

func (r *highLevelRuleSt) beginHeartBeat() {
	r.seasonGameData = gamedata.GetSeasonPvpGameData()
	r.param = gamedata.GetGameData(consts.MatchParam).(*gamedata.MatchParamGameData)
	r.beginTick(r.param.TimeInterval)

	r.param.AddReloadCallback(func(data gamedata.IGameData) {
		r.param = data.(*gamedata.MatchParamGameData)
		if r.tickerInterval != r.param.TimeInterval {
			r.beginTick(r.param.TimeInterval)
		}
	})
}

func (r *highLevelRuleSt) addPlayer(player *matchPlayer) {
	mplayer := newHighLevelPlayer(player, r.param)
	pElem := r.playerList.PushBack(mplayer)
	r.allPlayers[mplayer.getUid()] = pElem
	glog.Infof("highLevelRuleSt addPlayer %d", mplayer.getUid())
}

func (r *highLevelRuleSt) delPlayer(uid common.UUid) {
	if pElem, ok := r.allPlayers[uid]; ok {
		delete(r.allPlayers, uid)
		r.playerList.Remove(pElem)
		glog.Infof("highLevelRuleSt delPlayer %d", uid)
	}
}

func (r *highLevelRuleSt) isTargetPlayer(player *matchPlayer, streakLoseCnt int) bool {
	return player.getPvpTeam() >= 2
}

func (r *highLevelRuleSt) timeToMatchRobot(mplayer *highLevelPlayer, now time.Time) bool {
	pvpLevel := mplayer.getPvpLevel()
	if pvpLevel > maxMatchRobotLevel {
		return false
	}

	// 中级
	matchRobotInterval := matchRobotInterval
	if pvpLevel <= 4 {
		// 初级
		matchRobotInterval = newbieMatchRobotInterval
	} else if pvpLevel > 7 {
		// 高级
		matchRobotInterval = matchRobotInterval2
	}

	return now.Sub(mplayer.getBeginMatchTime()) >= matchRobotInterval
}
