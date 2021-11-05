package main

import (
	"kinger/proto/pb"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/timer"
	"time"
	"kinger/gopuppy/common/eventhub"
	"math/rand"
	"kinger/gopuppy/apps/logic"
	"kinger/common/consts"
	//"kinger/gopuppy/common/glog"
)

var (
	fieldMatchMgr = &fieldMatchMgrSt{}
	cityMatchMgr = &cityMatchMgrSt{}
)

type fieldMatchMgrSt struct {
	readyUidSet common.UInt64Set
	readyQueue []*matchTeam
	uid2MatchTeam map[common.UUid]*matchTeam
	road2Queue map[int]map[uint32][]*matchTeam

	ticker *timer.Timer
}

func (mm *fieldMatchMgrSt) initialize() {
	mm.readyUidSet = common.UInt64Set{}
	mm.uid2MatchTeam = map[common.UUid]*matchTeam{}
	mm.road2Queue = map[int]map[uint32][]*matchTeam{}

	if warMgr.isInWar() {
		mm.onWarBegin()
	}

	eventhub.Subscribe(evWarBegin, mm.onWarBegin)
	eventhub.Subscribe(evWarEnd, mm.onWarEnd)
}

func (mm *fieldMatchMgrSt) onWarBegin(_ ...interface{}) {
	if mm.ticker != nil && mm.ticker.IsActive() {
		return
	}

	mm.ticker = timer.AddTicker(time.Second, func() {
		if warMgr.isPause() {
			return
		}
		if len(mm.readyQueue) <= 0 {
			return
		}

		for _, t := range mm.readyQueue {
			if _, ok := mm.uid2MatchTeam[t.p.getUid()]; ok {
				continue
			}

			mm.uid2MatchTeam[t.p.getUid()] = t
			c2ts, ok := mm.road2Queue[t.roadID]
			if !ok {
				c2ts = map[uint32][]*matchTeam{}
				mm.road2Queue[t.roadID] = c2ts
			}

			ts := c2ts[t.countryID]
			c2ts[t.countryID] = append(ts, t)
		}
		mm.readyQueue = []*matchTeam{}
		mm.readyUidSet = common.UInt64Set{}
	})
}

func (mm *fieldMatchMgrSt) onWarEnd(_ ...interface{}) {
	if mm.ticker != nil {
		mm.ticker.Cancel()
		mm.ticker = nil
	}

	mm.readyUidSet = common.UInt64Set{}
	mm.readyQueue = []*matchTeam{}
	mm.uid2MatchTeam = map[common.UUid]*matchTeam{}
	mm.road2Queue = map[int]map[uint32][]*matchTeam{}
}

func (mm *fieldMatchMgrSt) onMatchDone(t1, t2 *matchTeam) {
	logic.PushBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, &pb.BeginBattleArg{
		BattleType:         int32(consts.BtCampaign),
		Fighter1:           t1.fighterData,
		Fighter2:           t2.fighterData,
		NeedFortifications: true,
		BonusType:             2,
		NeedVideo:          true,
		UpperType: 3,
	})
}

func (mm *fieldMatchMgrSt) onMatchTick() {
	for _, c2ts := range mm.road2Queue {
		for countryID, ts := range c2ts {
			ts = c2ts[countryID]
			if len(ts) <= 0 {
				continue
			}

			for countryID2, ts2 := range c2ts {
				if countryID == countryID2 {
					continue
				}

				for len(ts) > 0 && len(ts2) > 0 {
					i1 := rand.Intn(len(ts))
					mt1 := ts[i1]
					t1 := sceneMgr.getTeam(mt1.tid)
					if t1 == nil || t1.getState() != pb.TeamState_NormalTS {
						ts = append(ts[:i1], ts[i1+1:]...)
						c2ts[countryID] = ts
						delete(mm.uid2MatchTeam, mt1.p.getUid())
						continue
					}

					i2 := rand.Intn(len(ts2))
					mt2 := ts2[i2]
					t2 := sceneMgr.getTeam(mt2.tid)
					if t2 == nil || t2.getState() != pb.TeamState_NormalTS {
						ts2 = append(ts2[:i2], ts2[i2+1:]...)
						c2ts[countryID2] = ts2
						delete(mm.uid2MatchTeam, mt2.p.getUid())
						continue
					}

					ts = append(ts[:i1], ts[i1+1:]...)
					c2ts[countryID] = ts
					delete(mm.uid2MatchTeam, mt1.p.getUid())
					t1.onFieldMatchDone()

					ts2 = append(ts2[:i2], ts2[i2+1:]...)
					c2ts[countryID2] = ts2
					delete(mm.uid2MatchTeam, mt2.p.getUid())
					t2.onFieldMatchDone()

					mm.onMatchDone(mt1, mt2)
				}
			}
		}
	}
}

func (mm *fieldMatchMgrSt) stopMatch(uid common.UUid) {
	if mm.readyUidSet.Contains(uint64(uid)) {
		mm.readyUidSet.Remove(uint64(uid))
		for i, t := range mm.readyQueue {
			if t.p.getUid() == uid {
				mm.readyQueue = append(mm.readyQueue[:i], mm.readyQueue[i+1:]...)
				break
			}
		}
	}

	mt, ok := mm.uid2MatchTeam[uid]
	if !ok {
		return
	}

	delete(mm.uid2MatchTeam, uid)
	c2ts, ok := mm.road2Queue[mt.roadID]
	if !ok {
		return
	}

	ts, ok := c2ts[mt.countryID]
	if !ok {
		return
	}

	for i, t := range ts {
		if t.p.getUid() == uid {
			c2ts[mt.countryID] = append(ts[:i], ts[i+1:]...)
			break
		}
	}
}

func (mm *fieldMatchMgrSt) beginMatch(roadID int, t *team) {
	if t.fighterData == nil {
		return
	}
	mm.stopMatch(t.getOwner().getUid())
	mt := newMatchTeam(roadID, 0, t)
	mm.readyUidSet.Add(uint64(mt.p.getUid()))
	mm.readyQueue = append(mm.readyQueue, mt)
}

const (
	attackCityIdx = 0
	defCityIdx = 1
)
type cityMatchMgrSt struct {
	readyUidSet   common.UInt64Set
	atkReadyQueue []*matchTeam
	defReadyQueue []*matchTeam
	uid2MatchTeam map[common.UUid]*matchTeam
	city2Queue    map[int][][]*matchTeam

	ticker *timer.Timer
}

func (mm *cityMatchMgrSt) initialize() {
	mm.readyUidSet = common.UInt64Set{}
	mm.uid2MatchTeam = map[common.UUid]*matchTeam{}
	mm.city2Queue = map[int][][]*matchTeam{}

	if warMgr.isInWar() {
		mm.onWarBegin()
	}

	eventhub.Subscribe(evWarBegin, mm.onWarBegin)
	eventhub.Subscribe(evWarEnd, mm.onWarEnd)
}

func (mm *cityMatchMgrSt) onWarBegin(_ ...interface{}) {
	if mm.ticker != nil && mm.ticker.IsActive() {
		return
	}

	mm.ticker = timer.AddTicker(time.Second, func() {
		if warMgr.isPause() {
			return
		}
		if len(mm.atkReadyQueue) <= 0 && len(mm.defReadyQueue) <= 0 {
			return
		}

		for i, queue := range [][]*matchTeam{ mm.atkReadyQueue, mm.defReadyQueue } {
			for _, t := range queue {
				if _, ok := mm.uid2MatchTeam[t.p.getUid()]; ok {
					continue
				}

				mm.uid2MatchTeam[t.p.getUid()] = t
				c2ts, ok := mm.city2Queue[t.cityID]
				if !ok {
					c2ts = [][]*matchTeam{ []*matchTeam{}, []*matchTeam{} }
					mm.city2Queue[t.cityID] = c2ts
				}

				ts := c2ts[i]
				c2ts[i] = append(ts, t)

				//glog.Infof("cityMatchMgrSt ready tick, tid=%d", t.tid)
			}
		}

		mm.atkReadyQueue = []*matchTeam{}
		mm.defReadyQueue = []*matchTeam{}
		mm.readyUidSet = common.UInt64Set{}
	})
}

func (mm *cityMatchMgrSt) onWarEnd(_ ...interface{}) {
	if mm.ticker != nil {
		mm.ticker.Cancel()
		mm.ticker = nil
	}

	mm.readyUidSet = common.UInt64Set{}
	mm.atkReadyQueue = []*matchTeam{}
	mm.defReadyQueue = []*matchTeam{}
	mm.uid2MatchTeam = map[common.UUid]*matchTeam{}
	mm.city2Queue = map[int][][]*matchTeam{}
}

func (mm *cityMatchMgrSt) onMatchDone(atkMt, defMt *matchTeam) {
	cty := cityMgr.getCity(defMt.cityID)
	if cty != nil {
		defMt.fighterData.CasterSkills = cty.getGameData().Castle
	}

	logic.PushBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, &pb.BeginBattleArg{
		BattleType:         int32(consts.BtCampaign),
		Fighter1:           atkMt.fighterData,
		Fighter2:           defMt.fighterData,
		NeedFortifications: true,
		BonusType:             2,
		NeedVideo:          true,
		UpperType: 1,
	})

	defMt.fighterData.CasterSkills = []int32{}
}

func (mm *cityMatchMgrSt) onMatchTick() {
	for _, c2ts := range mm.city2Queue {
		atkTs := c2ts[attackCityIdx]
		defTs := c2ts[defCityIdx]

		for len(atkTs) > 0 && len(defTs) > 0 {
			i1 := rand.Intn(len(atkTs))
			mt1 := atkTs[i1]
			t1 := sceneMgr.getTeam(mt1.tid)
			if t1 == nil || t1.getState() != pb.TeamState_AttackingCityTS {
				atkTs = append(atkTs[:i1], atkTs[i1+1:]...)
				c2ts[attackCityIdx] = atkTs
				delete(mm.uid2MatchTeam, mt1.p.getUid())
				continue
			}

			i2 := rand.Intn(len(defTs))
			mt2 := defTs[i2]
			t2 := sceneMgr.getDefTeam(mt2.tid)
			if t2 == nil || t2.getState() != pb.TeamState_NormalTS {
				defTs = append(defTs[:i1], defTs[i1+1:]...)
				c2ts[defCityIdx] = defTs
				delete(mm.uid2MatchTeam, mt2.p.getUid())
				continue
			}

			atkTs = append(atkTs[:i1], atkTs[i1+1:]...)
			c2ts[attackCityIdx] = atkTs
			delete(mm.uid2MatchTeam, mt1.p.getUid())
			t1.onCityMatchDone(true)

			defTs = append(defTs[:i2], defTs[i2+1:]...)
			c2ts[defCityIdx] = defTs
			delete(mm.uid2MatchTeam, mt2.p.getUid())
			t2.onCityMatchDone(false)

			mm.onMatchDone(mt1, mt2)
		}
	}
}

func (mm *cityMatchMgrSt) stopMatch(uid common.UUid) {
	if mm.readyUidSet.Contains(uint64(uid)) {
		mm.readyUidSet.Remove(uint64(uid))
	L1: for idx, queue := range [][]*matchTeam{mm.atkReadyQueue, mm.defReadyQueue} {
			for i, t := range queue {
				if t.p.getUid() == uid {
					if idx == attackCityIdx {
						mm.atkReadyQueue = append(queue[:i], queue[i+1:]...)
					} else {
						mm.defReadyQueue = append(queue[:i], queue[i+1:]...)
					}
					break L1
				}
			}
		}
	}

	mt, ok := mm.uid2MatchTeam[uid]
	if !ok {
		return
	}

	delete(mm.uid2MatchTeam, uid)
	c2ts, ok := mm.city2Queue[mt.cityID]
	if !ok {
		return
	}

L2:	for idx, ts := range c2ts {
		for i, t := range ts {
			if t.p.getUid() == uid {
				c2ts[idx] = append(ts[:i], ts[i+1:]...)
				break L2
			}
		}
	}
}

func (mm *cityMatchMgrSt) beginMatch(cityID int, t *team) {
	if t.fighterData == nil {
		return
	}
	mm.stopMatch(t.getOwner().getUid())
	mt := newMatchTeam(0, cityID, t)
	mm.readyUidSet.Add(uint64(mt.p.getUid()))
	if t.isDefTeam() {
		//glog.Infof("cityMatchMgrSt def beginMatch cityID=%d, t=%s", cityID, t)
		mm.defReadyQueue = append(mm.defReadyQueue, mt)
	} else {
		//glog.Infof("cityMatchMgrSt atk beginMatch cityID=%d, t=%s", cityID, t)
		mm.atkReadyQueue = append(mm.atkReadyQueue, mt)
	}
}

type matchTeam struct {
	tid int
	p *player
	roadID int
	cityID int
	countryID uint32
	fighterData *pb.FighterData
}

func newMatchTeam(roadID, cityID int, t *team) *matchTeam {
	return &matchTeam{
		tid: t.getID(),
		p: t.getOwner(),
		roadID: roadID,
		cityID: cityID,
		countryID: t.getOwner().getCountryID(),
		fighterData: t.fighterData,
	}
}
