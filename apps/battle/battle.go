package main

import (
	"kinger/common/consts"
	kutils "kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"math/rand"
	"time"
)

var _ iBattle = &battle{}

type iBattle interface {
	packMsg() interface{}
	checkResult() (common.UUid, bool)
	battleEnd(winUid common.UUid, isSurrender, isCancel bool)
	boutReadyDone(f *fighter)
	getBattleID() common.UUid
	getBattleType() int
	getSituation() *battleSituation
	doAction(f *fighter, useCardObjID, gridID int) (*pb.FightBoutResult, int32)
	packAttr() *attribute.AttrMgr
	save(needReply bool)
	isEnd() bool
	onFigherLogin(agent *logic.PlayerAgent)
	onFighterLogout()
	packRestoredMsg() *pb.RestoredFightDesk
	restoredFromAttr(attr *attribute.AttrMgr, agent *logic.PlayerAgent)
	readyDone(cards ...uint32) error
	beginWaitClientTimer(t time.Duration)
	getBoutBeginTime() time.Time

	canSaveVideo() bool
	needBoutTiming() bool
}

type battle struct {
	i             iBattle
	battleID      common.UUid
	battleType    int
	situation     *battleSituation
	end           bool
	needVideo     bool
	videoData     *pb.VideoBattleData
	isFirstPvp    bool
	levelID       int
	boutBeginTime time.Time
	indexDiff     int

	waitClientTimer     *timer.Timer
	boutTimer           *timer.Timer
	saveTimer           *timer.Timer
	surrenderWaitClient chan struct{}
}

func newBattle(battleID common.UUid, battleType int, fighterData1, fighterData2 *pb.FighterData, upperType, bonusType,
	scale, battleRes int, needVideo bool, needFortifications, isFirstPvp bool, seasonData *gamedata.SeasonPvp,
	indexDiff int) *battle {

	battleObj := &battle{
		battleID:      battleID,
		battleType:    battleType,
		needVideo:     needVideo,
		isFirstPvp:    isFirstPvp,
		boutBeginTime: time.Now(),
		indexDiff:     indexDiff,
	}
	battleObj.i = battleObj
	battleObj.situation = newBattleSituation(fighterData1, fighterData2, battleID, upperType, bonusType, scale, battleRes,
		needFortifications, battleType, seasonData)
	f1 := battleObj.situation.fighter1
	f2 := battleObj.situation.fighter2
	f1.initDrawCard(battleType, fighterData1)
	f2.initDrawCard(battleType, fighterData2)
	return battleObj
}

func (b *battle) restoredFromAttr(attr *attribute.AttrMgr, agent *logic.PlayerAgent) {
	b.i = b
	b.battleID = common.UUid(attr.GetUInt64("battleID"))
	b.battleType = attr.GetInt("battleType")
	b.needVideo = attr.GetBool("needVideo")
	b.isFirstPvp = attr.GetBool("isFirstPvp")
	b.indexDiff = attr.GetInt("indexDiff")
	b.situation = &battleSituation{}

	if b.needVideo {
		strVideoData := attr.GetStr("videoData")
		if strVideoData != "" {
			b.videoData = &pb.VideoBattleData{}
			b.videoData.Unmarshal([]byte(strVideoData))
		}
	}

	b.situation.restoredFromAttr(attr.GetMapAttr("situation"), agent, b.battleType)
}

func (b *battle) canSaveVideo() bool {
	return mgr.isPvp(b.battleType) || b.battleType == consts.BtTraining
}

func (b *battle) needBoutTiming() bool {
	return mgr.isPvp(b.battleType) || b.battleType == consts.BtTraining
}

func (b *battle) getBoutBeginTime() time.Time {
	return b.boutBeginTime
}

func (b *battle) isEnd() bool {
	return b.end
}

func (b *battle) getBattleType() int {
	return b.battleType
}

func (b *battle) getSituation() *battleSituation {
	return b.situation
}

func (b *battle) getBattleID() common.UUid {
	return b.battleID
}

func (b *battle) packMsg() interface{} {
	msg := &pb.FightDesk{
		DeskId:     uint64(b.battleID),
		Type:       int32(b.battleType),
		Fighter1:   b.situation.getFighter1().packMsg(),
		Fighter2:   b.situation.getFighter2().packMsg(),
		Scale:      int32(b.situation.scale),
		BattleRes:  int32(b.situation.battleRes),
		IsFirstPvp: b.isFirstPvp,
	}

	for _, objID := range b.situation.grids {
		gMsg := &pb.Grid{}
		t := b.situation.getTargetMgr().getTarget(objID)
		if t.getType() == stEmptyGrid {
			gMsg.ObjId = int32(t.getObjID())
			gridObj := t.(*deskGrid)
			for _, effect := range gridObj.effects {
				gMsg.Effect = append(gMsg.Effect, effect.packMsg())
			}
		} else {
			c := t.(*fightCard)
			gMsg.ObjId = int32(c.getGridObj().getObjID())
			gMsg.InGridCard = c.packMsg()
			gMsg.Owner = uint64(c.getControllerUid())
		}

		msg.Grids = append(msg.Grids, gMsg)
	}

	return msg
}

func (b *battle) videoBegin() {
	if !b.needVideo {
		return
	}
	b.videoData = &pb.VideoBattleData{}
	beginAction := &pb.VideoAction{
		ID: pb.VideoAction_Begin,
	}
	data := b.packMsg().(*pb.FightDesk)
	beginAction.Data, _ = data.Marshal()
	b.videoData.Actions = append(b.videoData.Actions, beginAction)
}

func (b *battle) videoBoutBegin(data *pb.FightBoutBegin) {
	if !b.needVideo {
		return
	}
	action := &pb.VideoAction{
		ID: pb.VideoAction_BoutBegin,
	}
	action.Data, _ = data.Marshal()
	b.videoData.Actions = append(b.videoData.Actions, action)
}

func (b *battle) addVideoBoutAction(data *pb.FightBoutResult) {
	if !b.needVideo {
		return
	}
	action := &pb.VideoAction{
		ID: pb.VideoAction_BoutAction,
	}
	action.Data, _ = data.Marshal()
	b.videoData.Actions = append(b.videoData.Actions, action)
}

func (b *battle) isNeedVideo() bool {
	return b.needVideo && b.videoData != nil
}

func (b *battle) videoEnd(data *pb.BattleResult) {
	if !b.needVideo {
		return
	}
	action := &pb.VideoAction{
		ID: pb.VideoAction_End,
	}
	action.Data, _ = data.Marshal()
	b.videoData.Actions = append(b.videoData.Actions, action)

	attr := attribute.NewAttrMgr("battleVideo", b.battleID, true)

	video, _ := b.videoData.Marshal()
	attr.SetStr("data", string(video))
	attr.Save(false)
}

func (b *battle) readyDone(cards ...uint32) error {
	b.situation.setState(bsWaitClient)
	f1 := b.situation.getFighter1()
	f2 := b.situation.getFighter2()
	f1.setInitialHand()
	f2.setInitialHand()
	// TODO maybe fight card

	if b.needVideo {
		b.videoBegin()
	}

	b.situation.setState(bsWaitClient)
	b.syncReadyFight()

	if f1.isRobot {
		b.i.boutReadyDone(f1)
	}
	if f2.isRobot {
		b.i.boutReadyDone(f2)
	}

	return nil
}

func (b *battle) boutReadyDone(fter *fighter) {
	if b.situation.getState() != bsWaitClient {
		return
	}

	f1 := b.situation.getFighter1()
	f2 := b.situation.getFighter2()
	if f1 == fter {
		f1.ready()
	} else if f2 == fter {
		f2.ready()
	}

	if b.waitClientTimer != nil {
		b.waitClientTimer.Cancel()
		b.waitClientTimer = nil
	}

	if f1.isReady() && f2.isReady() {
		b.boutBegin()
	} else if !fter.isRobot {
		b.waitClientTimer = timer.AfterFunc(3*time.Second, b.onWaitClientTimeout)
	}
}

func (b *battle) boutBegin() {
	if b.situation.getState() != bsWaitClient {
		return
	}

	b.boutBeginTime = time.Now()
	actions := b.situation.boutBegin()
	b.syncBoutBegin(actions)
	b.beginBoutTimer()

	winUid, _ := b.checkResult()
	if winUid != 0 {
		timer.AfterFunc(2*time.Second, func() {
			if !b.isEnd() {
				b.battleEnd(winUid, false, false)
				mgr.delBattle(b.getBattleID())
			}
		})
		return
	}

	if b.situation.getCurBoutFighter().isRobot {
		b.aiThink()
	}
}

func (b *battle) boutEnd() {
	if b.situation.getState() != bsInBout {
		return
	}
	b.situation.boutEnd()
	// 等待客户端做完表现
	b.beginWaitClientTimer(35 * time.Second)
}

func (b *battle) syncReadyFight() {
	if b.getBattleType() == consts.BtGuide {
		return
	}

	msg := b.packMsg()
	fighter1 := b.situation.getFighter1()
	fighter2 := b.situation.getFighter2()
	if !fighter1.isRobot && fighter1.agent != nil {
		fighter1.agent.PushClient(pb.MessageID_S2C_READY_FIGHT, msg)
	}
	if !fighter2.isRobot && fighter2.agent != nil {
		fighter2.agent.PushClient(pb.MessageID_S2C_READY_FIGHT, msg)
	}
}

func (b *battle) syncBoutBegin(actions []*clientAction) {
	msg := &pb.FightBoutBegin{
		BoutUid:  uint64(b.situation.getCurBoutFighter().getUid()),
		BattleID: uint64(b.battleID),
	}

	msg.Actions = make([]*pb.ClientAction, len(actions))
	for i, act := range actions {
		msg.Actions[i] = act.packMsg()
	}

	if b.needBoutTiming() {
		msg.BoutTimeout = int32(b.situation.getCurBoutFighter().getBoutTimeout())
	}

	f1 := b.situation.getFighter1()
	f2 := b.situation.getFighter2()

	b.videoBoutBegin(msg)

	if !f1.isRobot && f1.agent != nil {
		f1.agent.PushClient(pb.MessageID_S2C_FIGHT_BOUT_BEGIN, msg)
	}
	if !f2.isRobot && f2.agent != nil {
		f2.agent.PushClient(pb.MessageID_S2C_FIGHT_BOUT_BEGIN, msg)
	}
}

func (b *battle) onWaitClientTimeout() {
	if b.waitClientTimer == nil {
		return
	}
	if b.end {
		return
	}

	state := b.situation.getState()
	if state == bsReady || state == bsWaitClient {
		b.waitClientTimer = nil
		b.boutBegin()
	}
}

func (b *battle) beginWaitClientTimer(t time.Duration) {
	b.situation.fighter1.wait()
	b.situation.fighter2.wait()

	if b.situation.fighter1.isRobot || b.situation.fighter2.isRobot {
		return
	}

	b.waitClientTimer = timer.AfterFunc(t, b.onWaitClientTimeout)
}

func (b *battle) beginBoutTimer() {
	if !b.needBoutTiming() {
		return
	}

	bout := b.situation.getCurBout()
	curBoutFighter := b.situation.getCurBoutFighter()
	boutTimeout := 4
	if curBoutFighter.getHandAmount() > 0 {
		boutTimeout = curBoutFighter.getBoutTimeout()
	}

	b.boutTimer = timer.AfterFunc(time.Duration(boutTimeout)*time.Second, func() {

		if b.boutTimer == nil {
			return
		}

		if b.end {
			return
		}

		if b.situation.getState() == bsInBout && bout == b.situation.getCurBout() {

			b.boutTimer = nil
			cardObjID := curBoutFighter.randomHandCard()
			curBoutFighter.setBoutTimeout(boutTimeOut2)
			gridID := -1
			if cardObjID > 0 {
				gridID = b.situation.randomUnUseGrid(curBoutFighter)
			}

			result, _ := b.doAction(curBoutFighter, cardObjID, gridID)
			if result == nil {
				return
			}

			if !curBoutFighter.isRobot && curBoutFighter.agent != nil {
				curBoutFighter.agent.PushClient(pb.MessageID_S2C_FIGHT_BOUT_RESULT, result)
			}
		}
	})

}

func (b *battle) doAction(f *fighter, useCardObjID, gridID int) (*pb.FightBoutResult, int32) {
	if b.end {
		return nil, 1
	}

	reply, card, errcode := b.situation.doAction(f, useCardObjID, gridID)
	if reply == nil {
		return nil, errcode
	}

	if card != nil {
		f.addUseCard(card.gcardID)
	}
	if b.boutTimer != nil {
		b.boutTimer.Cancel()
		b.boutTimer = nil
	}

	if b.situation.nextBoutFighter == nil {
		acts := b.situation.bonusBoutEnd(common.UUid(reply.WinUid))
		for _, act := range acts {
			reply.Actions = append(reply.Actions, act.packMsg())
		}
	}

	enemyFighter := b.situation.getEnemyFighter(f)
	if !enemyFighter.isRobot && enemyFighter.agent != nil {
		enemyFighter.agent.PushClient(pb.MessageID_S2C_FIGHT_BOUT_RESULT, reply)
	}

	if reply.WinUid == 0 {
		b.boutEnd()
	}

	b.addVideoBoutAction(reply)

	if reply.WinUid != 0 && b.situation.fighter1.isRobot && b.situation.fighter2.isRobot {
		b.battleEnd(common.UUid(reply.WinUid), false, false)
		return nil, 2
	}

	if b.situation.fighter1.isRobot {
		b.i.boutReadyDone(b.situation.fighter1)
	}
	if b.situation.fighter2.isRobot {
		b.i.boutReadyDone(b.situation.fighter2)
	}

	return reply, 0
}

func (b *battle) getAiThinkTime() time.Duration {
	if !b.needBoutTiming() {
		return 3 * time.Second
	}

	switch b.situation.getCurBout() {
	case 1:
		return time.Duration(rand.Intn(3)+1) * time.Second
	case 2:
		return time.Duration(rand.Intn(3)+6) * time.Second
	case 3:
	case 4:
	case 5:
		return time.Duration(rand.Intn(4)+6) * time.Second
	case 6:
		return time.Duration(rand.Intn(3)+6) * time.Second
	case 7:
		return time.Duration(rand.Intn(3)+5) * time.Second
	case 8:
		return time.Duration(rand.Intn(4)+4) * time.Second
	case 9:
		return time.Duration(rand.Intn(2)+1) * time.Second
	default:
		return 3 * time.Second
	}
	return 3 * time.Second
}

func (b *battle) aiThink() {
	curFighter := b.situation.getCurBoutFighter()
	if curFighter.getHandAmount() <= 0 {
		timer.AfterFunc(time.Second, func() {
			b.doAction(curFighter, 0, 0)
		})
		return
	}

	enemyFighter := b.situation.getEnemyFighter(curFighter)
	situation := b.situation.copy()

	go func() {
		var act *battleAction

		utils.CatchPanic(func() {
			var positive, negative int
			battleType := b.getBattleType()
			if (battleType == consts.BtPvp || battleType == consts.BtGuide) && !enemyFighter.isRobot {
				positive, negative = enemyFighter.getAiIQ(battleType)
			}
			act = searchBattleAction(situation, positive, negative)
			if !enemyFighter.isRobot {
				time.Sleep(b.getAiThinkTime())
			}
		})

		var cardObjID int
		var girdID int
		if act != nil {
			cardObjID = act.getCardObjID()
			girdID = act.getGridID()
		}
		evq.CallLater(func() {
			b.doAction(curFighter, cardObjID, girdID)
		})
	}()
}

func (b *battle) checkResult() (common.UUid, bool) {
	return b.situation.checkResult()
}

func (b *battle) battleEnd(winUid common.UUid, isSurrender, isCancel bool) {
	if b.end {
		if b.surrenderWaitClient != nil {
			c := b.surrenderWaitClient
			b.surrenderWaitClient = nil
			close(c)
		}
		return
	}
	b.end = true

	if isSurrender || isCancel {
		glog.Infof("battleEnd, winUid=%d isSurrender=%v isCancel=%v", winUid, isSurrender, isCancel)
	}

	fighter1 := b.situation.getFighter1()
	fighter2 := b.situation.getFighter2()

	var fighter1IndexDiff int
	var fighter2IndexDiff int
	if !fighter1.isRobot && !fighter2.isRobot {
		fighter1IndexDiff = b.indexDiff
		fighter2IndexDiff = -b.indexDiff
	}

	winer := fighter1
	loser := fighter2
	winerIndexDiff := fighter1IndexDiff
	loserIndexDiff := fighter2IndexDiff
	if fighter2.getUid() == winUid {
		winer, loser = fighter2, fighter1
		winerIndexDiff, loserIndexDiff = fighter2IndexDiff, fighter1IndexDiff
	}

	if isSurrender {
		acts, _, _ := b.situation.triggerMgr.trigger(map[int][]iTarget{surrenderTrigger: []iTarget{loser}}, &triggerContext{
			triggerType:   surrenderTrigger,
			surrenderor:   loser,
			beSurrenderor: winer,
		})

		if len(acts) > 0 {
			msg := &pb.FightBoutResult{WinUid: uint64(winer.getUid()), BattleID: uint64(b.battleID)}
			for _, act := range acts {
				msg.Actions = append(msg.Actions, act.packMsg())
			}
			if !fighter1.isRobot {
				fighter1.agent.PushClient(pb.MessageID_S2C_FIGHT_BOUT_RESULT, msg)
			}
			if !fighter2.isRobot {
				fighter2.agent.PushClient(pb.MessageID_S2C_FIGHT_BOUT_RESULT, msg)
			}
			b.addVideoBoutAction(msg)

			tr := timer.AfterFunc(10*time.Second, func() {
				if b.surrenderWaitClient != nil {
					c := b.surrenderWaitClient
					b.surrenderWaitClient = nil
					close(c)
				}
			})

			b.surrenderWaitClient = make(chan struct{})
			evq.Await(func() {
				<-b.surrenderWaitClient
				tr.Cancel()
			})

		}
	}

	var videoID common.UUid
	needVideo := b.isNeedVideo()
	if needVideo {
		videoID = b.battleID
	}

	winnerEndFighter := newEndFighterData(winer, false, winerIndexDiff)
	loserEndFighter := newEndFighterData(loser, isSurrender, loserIndexDiff)
	//self.endHandler.HandleBattleEnd(winnerEndFighter, loserEndFighter, videoID)
	isWonderful := (winnerEndFighter.isFighter1 && b.situation.isWonderful1) || (!winnerEndFighter.isFighter1 &&
		b.situation.isWonderful2)

	mqMsg := &pb.RmqBattleEnd{
		BattleID:    uint64(b.battleID),
		BattleType:  int32(b.battleType),
		Winner:      winnerEndFighter.packEndFighterMsg(),
		Loser:       loserEndFighter.packEndFighterMsg(),
		LevelID:     int32(b.levelID),
		IsWonderful: isWonderful,
	}

	for _, fter := range []*fighter{fighter1, fighter2} {
		if !fter.isRobot && (!isCancel || fter == winer) {
			kutils.PlayerMqPublish(fter.getUid(), pb.RmqType_BattleEnd, mqMsg)
			if fter.agent != nil {
				fter.agent.SetDispatchApp(consts.AppBattle, 0)
			}
		}
	}

	if needVideo {
		b.videoEnd(&pb.BattleResult{
			WinUid: uint64(winer.getUid()),
		})

		if b.canSaveVideo() {
			arg := &pb.SaveVideoArg{
				VideoID: uint64(videoID),
				Winner:  uint64(winnerEndFighter.uid),
			}
			if winnerEndFighter.isFighter1 {
				arg.Fighter1 = winnerEndFighter.packVideoFighterData()
				arg.Fighter2 = loserEndFighter.packVideoFighterData()
			} else {
				arg.Fighter2 = winnerEndFighter.packVideoFighterData()
				arg.Fighter1 = loserEndFighter.packVideoFighterData()
			}
			logic.PushBackend("", 0, pb.MessageID_B2V_SAVE_VIDEO, arg)
		}
	}

	if b.battleType == consts.BtPvp {
		for _, fter := range []*fighter{fighter1, fighter2} {
			if fter.isRobot {
				logic.BroadcastBackend(pb.MessageID_B2L_ON_ROBOT_BATTLE_END, &pb.OnRobotBattleEndArg{
					RobotID: uint64(fter.robotID),
					IsWin:   winnerEndFighter.isFighter1,
				})
			}
		}
	}
}

func (b *battle) packAttr() *attribute.AttrMgr {
	attr := attribute.NewAttrMgr("battle", b.battleID)
	attr.SetUInt64("battleID", uint64(b.battleID))
	attr.SetInt("battleType", b.battleType)
	attr.SetBool("needVideo", b.needVideo)
	attr.SetBool("isFirstPvp", b.isFirstPvp)
	attr.SetMapAttr("situation", b.situation.packAttr())
	attr.SetInt("indexDiff", b.indexDiff)
	if b.videoData != nil {
		data, _ := b.videoData.Marshal()
		attr.SetStr("videoData", string(data))
	}
	return attr
}

func (b *battle) save(needReply bool) {
	//glog.Infof("battle save 1111111111")
	if b.end {
		return
	}
	if b.situation.state == bsCreate {
		return
	}
	//glog.Infof("battle save 22222222222")

	attr := b.i.packAttr()
	attr.SetInt("version", battleAttrVersion)
	if err := attr.Save(needReply); err != nil {
		glog.Errorf("battle %d save error %s", b.battleID, err)
		return
	}
	//glog.Infof("battle save 333333333333")
	fighter1 := b.situation.fighter1
	if !fighter1.isRobot {
		fighter1Attr := attribute.NewAttrMgr("playerBattleInfo", fighter1.getUid())
		fighter1Attr.SetUInt64("battleID", uint64(b.battleID))
		fighter1Attr.SetUInt32("appID", bService.AppID)
		fighter1Attr.Save(needReply)
	}

	fighter2 := b.situation.fighter2
	if !fighter2.isRobot {
		fighter2Attr := attribute.NewAttrMgr("playerBattleInfo", fighter2.getUid())
		fighter2Attr.SetUInt64("battleID", uint64(b.battleID))
		fighter2Attr.SetUInt32("appID", bService.AppID)
		fighter2Attr.Save(needReply)
	}
}

func (b *battle) onFigherLogin(agent *logic.PlayerAgent) {
	if b.saveTimer != nil {
		b.saveTimer.Cancel()
		b.saveTimer = nil
	}

	if b.situation.fighter1.getUid() == agent.GetUid() {
		agent.SetDispatchApp(consts.AppBattle, bService.AppID)
		b.situation.fighter1.agent = agent
		b.situation.fighter1.wait()
	} else if b.situation.fighter2.getUid() == agent.GetUid() {
		agent.SetDispatchApp(consts.AppBattle, bService.AppID)
		b.situation.fighter2.agent = agent
		b.situation.fighter2.wait()
	}
}

func (b *battle) packRestoredMsg() *pb.RestoredFightDesk {
	return &pb.RestoredFightDesk{
		Desk:       b.packMsg().(*pb.FightDesk),
		CurBoutUid: uint64(b.situation.getCurBoutFighter().getUid()),
		CurBout:    int32(b.situation.getCurBout()),
	}
}

func (b *battle) onFighterLogout() {
	if b.needBoutTiming() || b.saveTimer != nil {
		return
	}

	b.saveTimer = timer.AfterFunc(15*time.Minute, func() {
		b.saveTimer = nil
		b.i.save(false)
		mgr.delBattle(b.battleID)
	})
}
