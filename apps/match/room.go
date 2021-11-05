package main

import (
	"kinger/gopuppy/apps/logic"
	"kinger/common/consts"
	"kinger/proto/pb"
)

type iMatchRoom interface {
	getID() int32
	beginBattle()
	getPlayerAgent1() *logic.PlayerAgent
	getPlayerAgent2() *logic.PlayerAgent
}

var _ iMatchRoom = &matchRoom{}
var _ iMatchRoom = &guideMatchRoom{}

type matchRoom struct {
	roomID       int32
	battleType   int
	upperType    int
	isFirstPvp   bool
	seasonDataID int
	indexDiff int
	player1      iMatchPlayer
	player2      iMatchPlayer
}

func newMatchRoom(roomID int32, battleType, upperType, indexDiff int) *matchRoom {
	return &matchRoom{
		roomID:     roomID,
		battleType: battleType,
		upperType: upperType,
		indexDiff: indexDiff,
	}
}

func (mr *matchRoom) getID() int32 {
	return mr.roomID
}

func (mr *matchRoom) beginBattle() {
	region := mr.player1.getAgent().GetRegion()
	region2 := mr.player2.getAgent().GetRegion()
	if region == 1 || region2 == 1 {
		region = 1
	}

	logic.PushBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, &pb.BeginBattleArg{
		BattleType:         int32(consts.BtPvp),
		Fighter1:           mr.player1.packFighterData(),
		Fighter2:           mr.player2.packFighterData(),
		NeedFortifications: true,
		BonusType:             2,
		NeedVideo:          true,
		UpperType: int32(mr.upperType),
		IsFirstPvp: mr.isFirstPvp,
		SeasonPvpSession: int32(mr.seasonDataID),
		IndexDiff: int32(mr.indexDiff),
	}, region)

	if !mr.player1.getAgent().IsRobot() {
		mr.player1.getAgent().SetDispatchApp(consts.AppMatch, 0)
	}
	if !mr.player2.getAgent().IsRobot() {
		mr.player2.getAgent().SetDispatchApp(consts.AppMatch, 0)
	}
}

func (mr *matchRoom) packMsg() *pb.MatchInfo {
	msg := &pb.MatchInfo{
		RoomId: int32(mr.roomID),
	}
	//if mr.player1 != nil {
	//	msg.Player1 = mr.player1.packMsg()
	//}
	//if mr.player2 != nil {
	//	msg.Player2 = mr.player2.packMsg()
	//}
	return msg
}

func (mr *matchRoom) getPlayerAgent1() *logic.PlayerAgent {
	return mr.player1.getAgent()
}

func (mr *matchRoom) getPlayerAgent2() *logic.PlayerAgent {
	return mr.player2.getAgent()
}

func (mr *matchRoom) getPlayerAmount() int {
	amount := 0
	if mr.player1 != nil {
		amount += 1
	}
	if mr.player2 != nil {
		amount += 1
	}
	return amount
}

func (mr *matchRoom) syncMatchRoomInfo() {
	msg := mr.packMsg()
	if mr.player1 != nil && !mr.player1.getAgent().IsRobot() {
		mr.player1.getAgent().PushClient(pb.MessageID_S2C_UPDATE_MATCH_INFO, msg)
	}
	if mr.player2 != nil && !mr.player2.getAgent().IsRobot() {
		mr.player2.getAgent().PushClient(pb.MessageID_S2C_UPDATE_MATCH_INFO, msg)
	}
}

func (mr *matchRoom) addPlayer(player iMatchPlayer) {
	if mr.player1 == nil {
		mr.player1 = player
	} else {
		mr.player2 = player
	}

	if mr.player1 != nil && mr.player2 != nil {
		mr.syncMatchRoomInfo()
	}
}

type guideMatchRoom struct {
	roomID    int32
	agent     *logic.PlayerAgent
	battleArg *pb.BeginBattleArg
}

func newGuideMatchRoom(roomID int32, agent *logic.PlayerAgent, battleArg *pb.BeginBattleArg) *guideMatchRoom {
	return &guideMatchRoom{
		roomID:    roomID,
		agent:     agent,
		battleArg: battleArg,
	}
}

func (gr *guideMatchRoom) getID() int32 {
	return gr.roomID
}

func (gr *guideMatchRoom) getPlayerAgent1() *logic.PlayerAgent {
	return gr.agent
}

func (gr *guideMatchRoom) getPlayerAgent2() *logic.PlayerAgent {
	return nil
}

func (gr *guideMatchRoom) beginBattle() {
	reply, _ := logic.CallBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, gr.battleArg, gr.agent.GetRegion())
	gr.agent.PushClient(pb.MessageID_S2C_READY_FIGHT, reply)
	gr.agent.SetDispatchApp(consts.AppMatch, 0)
}

func (gr *guideMatchRoom) syncMatchInfo() {
	gr.agent.PushClient(pb.MessageID_S2C_UPDATE_MATCH_INFO, &pb.MatchInfo{
		RoomId: int32(gr.roomID),
	})
}
