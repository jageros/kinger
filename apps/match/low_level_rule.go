package main

import (
	"container/list"
	"time"
	"math"
	"kinger/gopuppy/common"
	//"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
)

// 专家以下的匹配规则

const (
	crossMmrInterval = 2 * time.Second
	maxMmrSectionCount   = math.MaxInt32

	matchStateNewPlayer         = iota     // 等待 crossMmrInterval 秒内
	matchStateWaitPlayer                   // 等待 2 * crossMmrInterval 秒内
	matchStateWaitLongPlayer               // 等待超过 2 * crossMmrInterval
	matchStateTimeout                      // 等待不小于 matchTimeout

	newbieMatchRobotInterval = 3 * time.Second    // 初级 newbieMatchRobotInterval 秒后匹配机器人
	matchRobotInterval       = 10 * time.Second   // 中级 matchRobotInterval 秒后匹配机器人
	matchRobotInterval2      = 18 * time.Second   // 高级 matchRobotInterval2 秒后匹配机器人
)

type lowLevelPlayer struct {
	*matchPlayer
	mmrSection int
	cardStrengthSection int
	state int
}

func (mp *lowLevelPlayer) getMmrSection() int {
	return mp.mmrSection
}

func (mp *lowLevelPlayer) setMmrSection(sec int) {
	mp.mmrSection = sec
}

func (mp *lowLevelPlayer) getCardStrengthSection() int {
	return mp.cardStrengthSection
}

func (mp *lowLevelPlayer) setCardStrengthSection(sec int) {
	mp.cardStrengthSection = sec
}

func (mp *lowLevelPlayer) getState() int {
	return mp.state
}

func (mp *lowLevelPlayer) setState(state int) {
	mp.state = state
}

type cardStreCollection struct {
	r                *lowLevelRuleSt
	// 最小卡等区间
	minCardSection int
	// 最大卡等区间
	maxCardSection   int
	cardStreSec2list map[int]*list.List // map[cardStrengthSection]playerList
}

func (cc *cardStreCollection) add(mplayer *lowLevelPlayer) *list.Element {
	cardStrengthSection := mplayer.getCardStrengthSection()
	if cardStrengthSection > cc.maxCardSection {
		cc.maxCardSection = cardStrengthSection
	} else if cardStrengthSection < cc.minCardSection {
		cc.minCardSection = cardStrengthSection
	}

	if cc.minCardSection < 0 {
		cc.minCardSection = cardStrengthSection
	}

	l, ok := cc.cardStreSec2list[cardStrengthSection]
	if !ok {
		l = list.New()
		cc.cardStreSec2list[cardStrengthSection] = l
	}
	return l.PushBack(mplayer)
}

func (cc *cardStreCollection) remove(elem *list.Element) {
	mplayer := elem.Value.(*lowLevelPlayer)
	if l, ok := cc.cardStreSec2list[mplayer.getCardStrengthSection()]; ok {
		l.Remove(elem)
	}
}

func (cc *cardStreCollection) heartBeat(now time.Time) {
	for _, l := range cc.cardStreSec2list {

		for elem := l.Front(); elem != nil; {

			curElem := elem
			mplayer := curElem.Value.(*lowLevelPlayer)
			canDel := cc.r.tryMatchLadder(mplayer, now)
			elem = elem.Next()
			if canDel {
				l.Remove(curElem)
			}
		}

	}
}

func (cc *cardStreCollection) newIterator(cardStreSec int) func() *list.List {
	upMax := false
	downMin := false
	diff := -1
	sec := cardStreSec

	return func() *list.List {
		for !upMax || !downMin {
			if sec >= cc.maxCardSection {
				upMax = true
			}
			if sec <= cc.minCardSection || cc.minCardSection < 0 {
				downMin = true
			}

			l, ok := cc.cardStreSec2list[sec]

			sec = cardStreSec + diff
			if diff > 0 {
				diff = - diff - 1
			} else {
				diff = - diff
			}

			if ok {
				return l
			}
		}

		return nil
	}
}


type mmrCollection struct {
	r                *lowLevelRuleSt
	minMmrSec int
	maxMmrSec int
	mmrSec2CardStreColl map[int]*cardStreCollection // map[mmrSection]*cardStreCollection
}

func (mc *mmrCollection) heartBeat(now time.Time) {
	for _, cc := range mc.mmrSec2CardStreColl {
		cc.heartBeat(now)
	}
}

func (mc *mmrCollection) getCardStreColl(mmrSec int) *cardStreCollection {
	return mc.mmrSec2CardStreColl[mmrSec]
}

func (mc *mmrCollection) getOrNewCardStreColl(mmrSec int) *cardStreCollection {
	if coll, ok := mc.mmrSec2CardStreColl[mmrSec]; ok {
		return coll
	} else {
		coll = &cardStreCollection{
			r: mc.r,
			minCardSection: -1,
			cardStreSec2list: map[int]*list.List{},
		}
		mc.mmrSec2CardStreColl[mmrSec] = coll

		if mc.minMmrSec < 0 {
			mc.minMmrSec = mmrSec
		}
		if mmrSec > mc.maxMmrSec {
			mc.maxMmrSec = mmrSec
		} else if mmrSec < mc.minMmrSec {
			mc.minMmrSec = mmrSec
		}

		return coll
	}
}

func (mc *mmrCollection) newIterator(mmrSec, crossSecCount int) func() *cardStreCollection {
	sec := mmrSec
	isUpMax := false
	isDownMax := false
	diff := -1

	return func() *cardStreCollection {
		for !isUpMax || !isDownMax {
			if sec >= mc.maxMmrSec {
				isUpMax = true
			}
			if sec <= mc.minMmrSec || mc.minMmrSec < 0 {
				isDownMax = true
			}

			coll, ok := mc.mmrSec2CardStreColl[sec]

			sec = mmrSec + diff
			if diff > 0 {
				diff = - diff - 1
				crossSecCount --
			} else {
				diff = - diff
			}

			if ok {
				return coll
			} else if crossSecCount < 0 {
				return nil
			}
		}

		return nil
	}
}


type lowLevelRuleMatchQueue struct {
	r                *lowLevelRuleSt
	pvpLevel2MmrColl map[int]*mmrCollection // map[pvpLevel]*mmrCollection
}

func newLowLevelRuleMatchQueue(r *lowLevelRuleSt) *lowLevelRuleMatchQueue {
	return &lowLevelRuleMatchQueue{
		r: r,
		pvpLevel2MmrColl: map[int]*mmrCollection{},
	}
}

func (q *lowLevelRuleMatchQueue) heartBeat(now time.Time) {
	for _, mc := range q.pvpLevel2MmrColl {
		mc.heartBeat(now)
	}
}

func (q *lowLevelRuleMatchQueue) getMMrColl(pvpLevel int) *mmrCollection {
	return q.pvpLevel2MmrColl[pvpLevel]
}

func (q *lowLevelRuleMatchQueue) getOrNewMMrColl(pvpLevel int) *mmrCollection {
	if coll, ok := q.pvpLevel2MmrColl[pvpLevel]; ok {
		return coll
	} else {
		coll = &mmrCollection{
			r: q.r,
			minMmrSec: -1,
			mmrSec2CardStreColl: map[int]*cardStreCollection{},
		}
		q.pvpLevel2MmrColl[pvpLevel] = coll
		return coll
	}
}

type lowLevelRuleSt struct {
	allPlayers map[common.UUid]*list.Element

	matchQueue1    *lowLevelRuleMatchQueue
	matchQueue2    *lowLevelRuleMatchQueue
	matchQueue3    *lowLevelRuleMatchQueue

	allMatchQueues  []*lowLevelRuleMatchQueue // crossMmrInterval 秒内的玩家，可匹配
	canMatchQueues1 []*lowLevelRuleMatchQueue // 2 * crossMmrInterval 秒内的玩家，可匹配
	canMatchQueues2 []*lowLevelRuleMatchQueue // 超过2 * crossMmrInterval 秒的玩家，可匹配
}

func newLowLevelRule() iMatchRule {
	r := &lowLevelRuleSt{
		allPlayers: map[common.UUid]*list.Element{},
	}

	r.matchQueue1 = newLowLevelRuleMatchQueue(r)
	r.matchQueue2 = newLowLevelRuleMatchQueue(r)
	r.matchQueue3 = newLowLevelRuleMatchQueue(r)

	r.allMatchQueues = []*lowLevelRuleMatchQueue{
		r.matchQueue1,
		r.matchQueue2,
		r.matchQueue3,
	}

	r.canMatchQueues1 = []*lowLevelRuleMatchQueue{
		r.matchQueue3,
		r.matchQueue2,
	}

	r.canMatchQueues2 = []*lowLevelRuleMatchQueue{
		r.matchQueue3,
	}

	return r
}

func (r *lowLevelRuleSt) beginHeartBeat() {
	timer.AddTicker(800 * time.Millisecond, func() {

		now := time.Now()
		for _, q := range r.allMatchQueues {
			q.heartBeat(now)
		}

	})
}

// 可以跨多少个mmr区间匹配
func (r *lowLevelRuleSt) getMmrSectionCanCross(state int) int {
	switch state {
	case matchStateNewPlayer:
		return 1
	case matchStateWaitPlayer:
		return 2
	case matchStateWaitLongPlayer:
		fallthrough
	default:
		return maxMmrSectionCount
	}
}

func (r *lowLevelRuleSt) getQueuesCanMatch(state int) []*lowLevelRuleMatchQueue {
	switch state {
	case matchStateNewPlayer:
		return r.allMatchQueues
	case matchStateWaitPlayer:
		return r.canMatchQueues1
	case matchStateWaitLongPlayer:
		fallthrough
	default:
		return r.canMatchQueues2
	}
}

func (r *lowLevelRuleSt) timeToMatchRobot(mplayer *lowLevelPlayer, now time.Time) bool {
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

func (r *lowLevelRuleSt) getPlayerState(mplayer *lowLevelPlayer, now time.Time) int {
	if mplayer.isMatchTimout(now) {
		return matchStateTimeout
	}

	beginMatchTime := mplayer.getBeginMatchTime()
	matchTime := now.Sub(beginMatchTime)
	if matchTime >= 2 * crossMmrInterval {
		return matchStateWaitLongPlayer
	} else if matchTime >= crossMmrInterval {
		return matchStateWaitPlayer
	} else {
		return matchStateNewPlayer
	}
}

func (r *lowLevelRuleSt) getMatchQueue(state int) *lowLevelRuleMatchQueue {
	switch state {
	case matchStateNewPlayer:
		return r.matchQueue1
	case matchStateWaitPlayer:
		return r.matchQueue2
	case matchStateWaitLongPlayer:
		fallthrough
	default:
		return r.matchQueue3
	}
}

func (r *lowLevelRuleSt) tryMatchLadder(mplayer *lowLevelPlayer, now time.Time) bool {
	pvpLevel := mplayer.getPvpLevel()
	uid := mplayer.getUid()
	mmrSec := mplayer.getMmrSection()
	cardStreSec := mplayer.getCardStrengthSection()
	state := mplayer.getState()
	mmrSectionCount := r.getMmrSectionCanCross(state)
	matchQueues := r.getQueuesCanMatch(state)
	lvs := gMatchMgr.getCanMatchLevels(pvpLevel)
	area := mplayer.getArea()

	for _, queue := range matchQueues {
		for _, lv := range lvs {
			mmrColl := queue.getMMrColl(lv)
			if mmrColl == nil {
				continue
			}

			//glog.Infof("lowLevelRuleSt tryMatchLadder 111 uid=%d, mmrColl=%p", mplayer.getUid(), mmrColl)
			// 找到合适pvplevel的人

			cardStreCollIter := mmrColl.newIterator(mmrSec, mmrSectionCount)
			for cardStreColl := cardStreCollIter(); cardStreColl != nil; cardStreColl = cardStreCollIter() {
				// 找到合适mmr区间的人

				//glog.Infof("lowLevelRuleSt tryMatchLadder 222 uid=%d, cardStreColl=%p", mplayer.getUid(), cardStreColl)
				plistIter := cardStreColl.newIterator(cardStreSec)
				for l := plistIter(); l != nil; l = plistIter() {
					// 找到接近卡等区间的人

					//glog.Infof("lowLevelRuleSt tryMatchLadder 333 uid=%d, l=%p", mplayer.getUid(), l)
					for oppElem := l.Front(); oppElem != nil; oppElem = oppElem.Next() {
						oppmPlayer := oppElem.Value.(*lowLevelPlayer)
						//glog.Infof("lowLevelRuleSt tryMatchLadder 444 uid=%d, opp=%d", mplayer.getUid(), oppmPlayer.getUid())
						if oppmPlayer.getUid() != uid && oppmPlayer.getArea() == area {
							gMatchMgr.onMatchLadder(mplayer, oppmPlayer, 0)
							l.Remove(oppElem)
							delete(r.allPlayers, oppmPlayer.getUid())
							delete(r.allPlayers, uid)
							return true
						}
					}
				}
			}
		}
	}

	if r.timeToMatchRobot(mplayer, now) {
		robot := robotMgr.getRobot(mplayer)
		if robot != nil {
			gMatchMgr.doMatchRobot(mplayer, robot, 0, false)
			delete(r.allPlayers, uid)
			return true
		}
	}

	newState := r.getPlayerState(mplayer, now)
	if newState == matchStateTimeout {
		gMatchMgr.onMatchTimeout(mplayer)
		delete(r.allPlayers, uid)
		return true
	}

	if newState == state {
		return false
	}

	// 状态切换，将玩家移到别的queue
	mplayer.setState(newState)
	pElem := r.getMatchQueue(newState).getOrNewMMrColl(pvpLevel).getOrNewCardStreColl(mmrSec).add(mplayer)
	r.allPlayers[uid] = pElem
	return true
}

func (r *lowLevelRuleSt) getMmrSection(mmr int) int {
	return mmr / 100
}

func (r *lowLevelRuleSt) getCardStrengthSection(cardStrength int) int {
	return cardStrength / 4
}

func (r *lowLevelRuleSt) addPlayer(player *matchPlayer) {
	//glog.Infof("lowLevelRuleSt addPlayer uid=%d", mplayer.getUid())
	mplayer := &lowLevelPlayer{state: matchStateNewPlayer}
	mplayer.matchPlayer = player
	mplayer.setMmrSection( r.getMmrSection(mplayer.getMmr()) )
	mplayer.setCardStrengthSection( r.getCardStrengthSection(mplayer.getCardStrength()) )
	pElem := r.matchQueue1.getOrNewMMrColl(mplayer.getPvpLevel()).getOrNewCardStreColl(mplayer.getMmrSection()).add(mplayer)
	r.allPlayers[mplayer.getUid()] = pElem
}

func (r *lowLevelRuleSt) delPlayer(uid common.UUid) {
	pElem, ok := r.allPlayers[uid]
	if !ok {
		return
	}
	delete(r.allPlayers, uid)

	//glog.Infof("lowLevelRuleSt delPlayer uid=%d", uid)
	mplayer := pElem.Value.(*lowLevelPlayer)
	pvpLevel := mplayer.getPvpLevel()
	mmrSection := mplayer.getMmrSection()
	for _, queue := range r.allMatchQueues {
		mmrColl := queue.getMMrColl(pvpLevel)
		if mmrColl != nil {
			cardStreColl := mmrColl.getCardStreColl(mmrSection)
			if cardStreColl != nil {
				cardStreColl.remove(pElem)
			}
		}
	}
}

func (r *lowLevelRuleSt) isTargetPlayer(player *matchPlayer, streakLoseCnt int) bool {
	return player.getPvpTeam() < 5
}
