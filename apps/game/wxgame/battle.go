package wxgame

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
	"time"
)

type inviteRoom struct {
	inviter    common.UUid
	createTime time.Time
}

func newInviteRoom(inviter common.UUid) *inviteRoom {
	return &inviteRoom{
		inviter:    inviter,
		createTime: time.Now(),
	}
}

func (r *inviteRoom) isTimeout(now time.Time) bool {
	if now.Sub(r.createTime) >= time.Minute {
		invitePlayer := module.Player.GetPlayer(r.inviter)
		if invitePlayer != nil && invitePlayer.GetAgent() != nil {
			invitePlayer.GetAgent().PushClient(pb.MessageID_S2C_WX_INVITE_BATTLE_RESULT, &pb.WxInviteBattleResult{
				Result: pb.WxInviteBattleResult_Timeout,
			})
		}
		return true
	}
	return false
}

func (r *inviteRoom) beginBattle(beInviterData *pb.FighterData) bool {
	invitePlayer := module.Player.GetPlayer(r.inviter)
	if invitePlayer == nil || invitePlayer.IsInBattle() {
		return false
	}

	module.Mission.OnInviteBattle(invitePlayer)

	invitePlayer.GetAgent().PushClient(pb.MessageID_S2C_WX_INVITE_BATTLE_RESULT, &pb.WxInviteBattleResult{
		Result: pb.WxInviteBattleResult_OK,
	})

	logic.PushBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, &pb.BeginBattleArg{
		BattleType:         consts.BtFriend,
		Fighter1:           invitePlayer.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpFighterData(),
		Fighter2:           beInviterData,
		NeedFortifications: true,
		NeedVideo:          true,
		UpperType:          3,
	})
	return true
}
