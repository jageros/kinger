package main

import (
	"kinger/proto/pb"
	"kinger/gopuppy/common"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/apps/logic"
	"time"
)

var mgr = &battleMgr{
	uid2Fighter: make(map[common.UUid]*fighter),
	id2Battle: make(map[common.UUid]iBattle),
}

type battleMgr struct {
	uid2Fighter map[common.UUid]*fighter
	id2Battle map[common.UUid]iBattle
}

func (bm *battleMgr) publishBeginBattle(battleObj iBattle) {
	mqMsg := &pb.RmqBattleBegin{
		BattleID:   uint64(battleObj.getBattleID()),
		AppID:      bService.AppID,
		BattleType: int32(battleObj.getBattleType()),
	}
	f1 := battleObj.getSituation().getFighter1()
	f2 := battleObj.getSituation().getFighter2()
	if !f1.isRobot {
		utils.PlayerMqPublish(f1.getUid(), pb.RmqType_BattleBegin, mqMsg)
	}
	if !f2.isRobot {
		utils.PlayerMqPublish(f2.getUid(), pb.RmqType_BattleBegin, mqMsg)
	}

	if !f1.isRobot && f1.agent != nil {
		f1.agent.SetDispatchApp(consts.AppBattle, bService.AppID)
	}
	if !f2.isRobot && f2.agent != nil {
		f2.agent.SetDispatchApp(consts.AppBattle, bService.AppID)
	}
	bm.addBattle(battleObj)
}

func (bm *battleMgr) beginBattle(battleType int, fighterData1, fighterData2 *pb.FighterData, upperType, bonusType, scale,
	battleRes int, needVideo bool, needFortifications, isFirstPvp bool, seasonDataID, indexDiff int) (iBattle, error) {

	if upperType <= 0 {
		upperType = 3
	}
	battleID := common.GenUUid("battle")
	var seasonData *gamedata.SeasonPvp
	if seasonDataID > 0 {
		seasonData = gamedata.GetGameData(consts.SeasonPvp).(*gamedata.SeasonPvpGameData).ID2Season[seasonDataID]
	}

	//if seasonData == nil {
	battleObj := newBattle(battleID, battleType, fighterData1, fighterData2, upperType, bonusType, scale, battleRes,
		needVideo, needFortifications, isFirstPvp, seasonData, indexDiff)
	//} else {

	//}

	f1 := battleObj.getSituation().fighter1
	f2 := battleObj.getSituation().fighter2
	if !f1.isRobot && f1.agent != nil {
		f1.agent.SetDispatchApp(consts.AppBattle, bService.AppID)
	}
	if !f2.isRobot && f2.agent != nil {
		f2.agent.SetDispatchApp(consts.AppBattle, bService.AppID)
	}

	battleObj.beginWaitClientTimer(7 * time.Second)
	battleObj.readyDone()
	bm.publishBeginBattle(battleObj)
	return battleObj, nil
}

func (bm *battleMgr) beginLevelBattle(fighterData *pb.FighterData, levelData *gamedata.Level, isHelp bool) (*levelBattle, error) {
	battleID := common.GenUUid("battle")
	battleObj := newLevelBattle(battleID, fighterData, levelData, isHelp)
	bm.addBattle(battleObj)
	return battleObj, nil
}

func (bm *battleMgr) isPvp(battleType int) bool {
	return battleType == consts.BtPvp || battleType == consts.BtFriend || battleType == consts.BtCampaign
}

func (bm *battleMgr) getFighter(uid common.UUid) *fighter {
	return bm.uid2Fighter[uid]
}

func (bm *battleMgr) getBattle(battleID common.UUid) iBattle {
	if b, ok := bm.id2Battle[battleID]; ok {
		return b
	} else {
		return nil
	}
}

func (bm *battleMgr) delBattle(battleID common.UUid) {
	b, ok := bm.id2Battle[battleID]
	if !ok {
		return
	}

	f1Uid := b.getSituation().getFighter1().getUid()
	f2Uid := b.getSituation().getFighter2().getUid()
	delete(bm.id2Battle, battleID)

	if f, ok := bm.uid2Fighter[f1Uid]; ok && f.battleID == battleID {
		delete(bm.uid2Fighter, f1Uid)
	}
	if f, ok := bm.uid2Fighter[f2Uid]; ok && f.battleID == battleID {
		delete(bm.uid2Fighter, f2Uid)
	}
}

func (bm *battleMgr) delFighter(uid common.UUid) {
	delete(bm.uid2Fighter, uid)
}

func (bm *battleMgr) addFighter(f *fighter) {
	bm.uid2Fighter[f.getUid()] = f
}

func (bm *battleMgr) addBattle(battleObj iBattle) {
	f1 := battleObj.getSituation().getFighter1()
	f2 := battleObj.getSituation().getFighter2()

	fs := []*fighter{f1, f2}
	for _, f := range fs {
		if f != nil && !f.isRobot {
			uid := f.getUid()
			oldFighter := bm.getFighter(uid)
			if oldFighter != nil {
				battleObj := oldFighter.getBattle()
				if battleObj != nil {
					if !battleObj.needBoutTiming() {
						bm.delBattle(battleObj.getBattleID())
						bm.delFighter(uid)
					} else {
						bm.surrender(oldFighter)
					}
				}
			}

			bm.uid2Fighter[uid] = f
		}
	}

	bm.id2Battle[battleObj.getBattleID()] = battleObj
}

func (bm *battleMgr) loadBattle(battleID common.UUid, agent *logic.PlayerAgent) *pb.RestoredFightDesk {
	battleObj := bm.getBattle(battleID)
	if battleObj != nil {
		if !battleObj.isEnd() {
			battleObj.onFigherLogin(agent)
			bm.addFighter(battleObj.getSituation().getFighter(agent.GetUid()))
			return battleObj.packRestoredMsg()
		} else {
			return nil
		}
	}

	attr := attribute.NewAttrMgr("battle", battleID)
	if err := attr.Load(); err != nil {
		return nil
	}

	attr.Delete(false)
	ver := attr.GetInt("version")
	if ver != battleAttrVersion {
		return nil
	}

	battleType := attr.GetInt("battleType")
	if battleType == consts.BtLevel || battleType == consts.BtLevelHelp {
		levelBattleObj := &levelBattle{}
		levelBattleObj.boutBeginTime = time.Now()
		battleObj = levelBattleObj
	} else {
		battleObj = &battle{boutBeginTime: time.Now()}
	}
	battleObj.restoredFromAttr(attr, agent)
	bm.addBattle(battleObj)
	battleObj.getSituation().state = bsWaitClient
	agent.SetDispatchApp(consts.AppBattle, bService.AppID)
	return battleObj.packRestoredMsg()
}

func (bm *battleMgr) onFighterLogout(uid common.UUid) {
	f := bm.getFighter(uid)
	if f != nil {
		f.wait()
		bm.delFighter(uid)
		battleObj := f.getBattle()
		if battleObj != nil && !bm.isPvp(battleObj.getBattleType()) {
			state := battleObj.getSituation().state
			if state == bsCreate || battleObj.isEnd() {
				bm.delBattle(battleObj.getBattleID())
			} else {
				battleObj.onFighterLogout()
			}
		}
	}
}

func (bm *battleMgr) cancelBattle(uid, battleID common.UUid) {
	f := bm.getFighter(uid)
	if f == nil {
		return
	}

	battleObj := f.getBattle()
	if battleObj == nil || battleObj.getBattleID() != battleID || battleObj.isEnd() {
		return
	}

	f1Uid := battleObj.getSituation().getFighter1().getUid()
	f2Uid := battleObj.getSituation().getFighter2().getUid()
	if f1Uid != uid && f2Uid != uid {
		return
	}

	winUid := f1Uid
	if winUid == uid {
		winUid = f2Uid
	}

	battleObj.battleEnd(winUid, false, true)
	mgr.delBattle(battleObj.getBattleID())
}

func (bm *battleMgr) surrender(f *fighter) error {
	battleObj := f.getBattle()
	if battleObj.isEnd() {
		return nil
	}

	uid := f.getUid()
	f1Uid := battleObj.getSituation().getFighter1().getUid()
	f2Uid := battleObj.getSituation().getFighter2().getUid()
	if f1Uid != uid && f2Uid != uid {
		return gamedata.InternalErr
	}

	winUid := f1Uid
	if winUid == uid {
		winUid = f2Uid
	}

	battleObj.battleEnd(winUid, true, false)
	mgr.delBattle(battleObj.getBattleID())
	return nil
}
