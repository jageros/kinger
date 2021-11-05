package social

import (
	"fmt"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
	"time"
	"kinger/gopuppy/common/eventhub"
)

type inviteRoom struct {
	inviter    common.UUid
	target     common.UUid
	createTime time.Time
}

func newInviteRoom(inviter, target common.UUid) *inviteRoom {
	return &inviteRoom{
		inviter:    inviter,
		target:     target,
		createTime: time.Now(),
	}
}

func (r *inviteRoom) String() string {
	return fmt.Sprintf("[room inviter=%d, target=%d]", r.inviter, r.target)
}

func (r *inviteRoom) checkTimeout(now time.Time) bool {
	if now.Sub(r.createTime) >= time.Minute {
		invitePlayer := module.Player.GetPlayer(r.inviter)
		if invitePlayer != nil && invitePlayer.GetAgent() != nil {
			invitePlayer.GetAgent().PushClient(pb.MessageID_S2C_INVITE_BATTLE_RESULT, &pb.InviteBattleResult{
				Result: pb.InviteBattleResult_Timeout,
			})
		}
		return true
	}
	return false
}

func (r *inviteRoom) beginBattle(targetFighter *pb.FighterData) bool {
	invitePlayer := module.Player.GetPlayer(r.inviter)
	if invitePlayer == nil || invitePlayer.IsInBattle() || !invitePlayer.IsOnline() {
		return false
	}

	module.Mission.OnInviteBattle(invitePlayer)
	eventhub.Publish(consts.EvCombat, invitePlayer)

	invitePlayer.GetAgent().PushClient(pb.MessageID_S2C_INVITE_BATTLE_RESULT, &pb.InviteBattleResult{
		Result: pb.InviteBattleResult_Agree,
	})

	evq.CallLater(func() {
		logic.PushBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, &pb.BeginBattleArg{
			BattleType:         int32(consts.BtFriend),
			Fighter1:           invitePlayer.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpFighterData(),
			Fighter2:           targetFighter,
			NeedFortifications: true,
			NeedVideo:          true,
		})
	})

	return true
}

func (r *inviteRoom) refuse() {
	invitePlayer := module.Player.GetPlayer(r.inviter)
	if invitePlayer != nil && invitePlayer.GetAgent() != nil {
		invitePlayer.GetAgent().PushClient(pb.MessageID_S2C_INVITE_BATTLE_RESULT, &pb.InviteBattleResult{
			Result: pb.InviteBattleResult_Refuse,
		})
	}
}
