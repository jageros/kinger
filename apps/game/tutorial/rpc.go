package tutorial

import (
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/proto/pb"
)

/*
func rpc_C2S_GetCampID(ses *network.Session, arg interface{}) (interface{}, error) {
	// arg1 := arg.(*pb.GetCampIDArg)
	uid := ses.GetProp("uid").(common.UUid)
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	tutorialComponent := player.GetComponent(consts.TutorialCpt).(*tutorialComponent)
	return &pb.GetCampIDReply{CampID: tutorialComponent.GetCampID()}, nil
}
*/

func rpc_C2S_SetCampID(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg1 := arg.(*pb.SetCampIDArg)
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	tutorialComponent := player.GetComponent(consts.TutorialCpt).(*tutorialComponent)
	return &pb.SetCampIDReply{Ok: tutorialComponent.setCampID(arg1.CampID)}, nil
}

func rpc_C2S_StartTutorialBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg1 := arg.(*pb.StartTutorialBattleArg)
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	tutorialComponent := player.GetComponent(consts.TutorialCpt).(*tutorialComponent)
	camp := tutorialComponent.GetCampID()
	if camp <= 0 {
		if arg1.CampID <= 0 {
			return nil, gamedata.InternalErr
		} else {
			tutorialComponent.setCampID(arg1.CampID)
		}
	}

	reply, err := tutorialComponent.startTutorialBattle()
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func registerRpc() {
	//logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_CAMP_ID, rpc_C2S_GetCampID)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SET_CAMP_ID, rpc_C2S_SetCampID)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_START_TUTORIAL_BATTLE, rpc_C2S_StartTutorialBattle)
}
