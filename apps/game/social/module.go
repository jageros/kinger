package social

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	//"kinger/gopuppy/network"
	"container/list"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/eventhub"
	"kinger/proto/pb"
	"time"
)

var mod *socialModule

type socialModule struct {
	inviteRooms     map[common.UUid]*list.Element
	inviteRoomsList *list.List
	beInviteRooms   map[common.UUid]map[common.UUid]*inviteRoom
	lastAdvertChat  string
}

func (m *socialModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("social")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("social", attr)
	}
	return &socialComponent{attr: attr}
}

func (m *socialModule) newInviteBattleRoom(uid, targetUid common.UUid) {
	r := newInviteRoom(uid, targetUid)
	elem := m.inviteRoomsList.PushBack(r)
	m.inviteRooms[uid] = elem
	rs, ok := m.beInviteRooms[r.target]
	if !ok {
		rs = map[common.UUid]*inviteRoom{}
		m.beInviteRooms[r.target] = rs
	}
	rs[uid] = r
}

func (m *socialModule) OnBattleBegin(player types.IPlayer) {
	uid := player.GetUid()
	if elem, ok := m.inviteRooms[uid]; ok {
		r := elem.Value.(*inviteRoom)
		beInviteUid := r.target
		delete(m.inviteRooms, uid)
		m.inviteRoomsList.Remove(elem)
		if rs, ok := m.beInviteRooms[beInviteUid]; ok {
			delete(rs, uid)
		}
	}

	glog.Debugf("OnBattleBegin uid=%d, inviteRooms=%v, beInviteRooms=%v",
		uid, m.inviteRooms, m.beInviteRooms)

	rs, ok := m.beInviteRooms[uid]
	if !ok {
		return
	}
	delete(m.beInviteRooms, uid)
	for _, r := range rs {
		if elem, ok := m.inviteRooms[r.inviter]; ok {
			delete(m.inviteRooms, uid)
			m.inviteRoomsList.Remove(elem)
		}
		r.refuse()
	}
}

func (m *socialModule) OnLogout(player types.IPlayer) {
	m.OnBattleBegin(player)
}

func (m *socialModule) CancelInviteBattle(uid common.UUid) {
	if elem, ok := m.inviteRooms[uid]; ok {
		r := elem.Value.(*inviteRoom)
		beInviteUid := r.target
		delete(m.inviteRooms, uid)
		m.inviteRoomsList.Remove(elem)
		if rs, ok := m.beInviteRooms[beInviteUid]; ok {
			delete(rs, uid)
		}
	}
}

func (m *socialModule) IsInInviteBattle(uid common.UUid) bool {
	_, ok := m.inviteRooms[uid]
	return ok
}

func (m *socialModule) replyInviteBattle(targetFighter *pb.FighterData, targetUid, inviter common.UUid, isAgree bool) bool {
	ok := true
	var r *inviteRoom
	var elem *list.Element
	if isAgree {
		//glog.Infof("replyInviteBattle 11111111 %v", m.inviteRooms)
		elem, ok = m.inviteRooms[inviter]
		if !ok {
			//glog.Infof("replyInviteBattle 2222222222 %v", m.inviteRooms)
			return false
		}
		r = elem.Value.(*inviteRoom)

		if r.target != targetUid {
			//glog.Infof("replyInviteBattle 3333333333 %v, r.target=%d, targetUid=%d", m.inviteRooms, r.target, targetUid)
			return false
		}

		p := module.Player.GetPlayer(inviter)
		if p == nil || module.Campaign.IsInCampaignMatch(p) || module.WxGame.IsInInviteBattle(p.GetUid()) {
			return false
		}

		ok = r.beginBattle(targetFighter)
		if !ok {
			return false
		}

		delete(m.inviteRooms, inviter)
		m.inviteRoomsList.Remove(elem)
		if rs, ok := m.beInviteRooms[targetUid]; ok {
			delete(rs, inviter)
		}

		glog.Debugf("replyInviteBattle agree uid=%d, inviter=%d, inviteRooms=%v, beInviteRooms=%v",
			targetUid, inviter, m.inviteRooms, m.beInviteRooms)
		return true
	} else {
		elem, ok = m.inviteRooms[inviter]
		if !ok {
			return true
		}
		r = elem.Value.(*inviteRoom)

		if r.target != targetUid {
			return true
		}

		r.refuse()
		delete(m.inviteRooms, inviter)
		m.inviteRoomsList.Remove(elem)
		if rs, ok := m.beInviteRooms[targetUid]; ok {
			delete(rs, inviter)
		}
		return true
	}
}

func (m *socialModule) AddFriendApply(player types.IPlayer, targetUid common.UUid, isInvite bool) error {
	if player.GetUid() == targetUid {
		return gamedata.GameError(1)
	}
	if player.GetComponent(consts.SocialCpt).(*socialComponent).getFriend(targetUid) != nil {
		return gamedata.GameError(2)
	}

	targerPlayer := module.Player.GetPlayer(targetUid)
	if targerPlayer != nil {
		socialCpt := targerPlayer.GetComponent(consts.SocialCpt).(*socialComponent)
		socialCpt.AddFriendApply(player.GetUid(), player.GetName(), isInvite)
	} else {

		utils.PlayerMqPublish(targetUid, pb.RmqType_AddFriendApply, &pb.RmqAddFriendApply{
			FromUid:  uint64(player.GetUid()),
			FromName: player.GetName(),
			IsInvite: isInvite,
		})
	}
	return nil
}

func (m *socialModule) WxInviteFriend(inviterUid common.UUid, targetPlayer types.IPlayer) {
	m.AddFriendApply(targetPlayer, inviterUid, true)
	socialCpt := targetPlayer.GetComponent(consts.SocialCpt).(*socialComponent)
	if socialCpt.getWxInviter() > 0 {
		return
	}
	socialCpt.setWxInviter(inviterUid)
	m.onWxInviteFriendUpdate(targetPlayer, inviterUid)
}

func (m *socialModule) onWxInviteFriendUpdate(player types.IPlayer, inviterUid common.UUid) {
	if inviterUid <= 0 {
		return
	}
	inviterPlayer := module.Player.GetPlayer(inviterUid)
	if inviterPlayer != nil {
		socialCpt := inviterPlayer.GetComponent(consts.SocialCpt).(*socialComponent)
		socialCpt.OnWxInviteFriendUpdate(player.GetUid(), player.GetHeadImgUrl(), player.GetMaxPvpLevel())
	} else {
		utils.PlayerMqPublish(inviterUid, pb.RmqType_WxInviteFriendTp, &pb.RmqWxInviteFriend{
			Uid:         uint64(player.GetUid()),
			HeadImgUrl:  player.GetHeadImgUrl(),
			MaxPvpLevel: int32(player.GetMaxPvpLevel()),
		})
	}
}

func (m *socialModule) onSendAdvertChat(msg string) {
	m.lastAdvertChat = msg
	logic.BroadcastBackend(pb.MessageID_G2G_ON_SEND_ADVERT_CHAT, &pb.OnSendAdvertChatArg{Msg: msg})
}

func onMaxPvpLevelUpdate(args ...interface{}) {
	player := args[0].(types.IPlayer)
	socialCpt := player.GetComponent(consts.SocialCpt).(*socialComponent)
	mod.onWxInviteFriendUpdate(player, socialCpt.getWxInviter())
	socialCpt.onMaxPvpLevelUpdate(args[2].(int))
}

func onFixServer1WxInvite(args ...interface{}) {
	player := args[0].(types.IPlayer)
	inviter := player.GetComponent(consts.SocialCpt).(*socialComponent).getWxInviter()
	if inviter <= 0 {
		return
	}

	timer.AfterFunc(time.Second, func() {
		mod.onWxInviteFriendUpdate(player, inviter)
	})
}

func Initialize() {
	mod = &socialModule{
		inviteRooms:     make(map[common.UUid]*list.Element),
		beInviteRooms:   make(map[common.UUid]map[common.UUid]*inviteRoom),
		inviteRoomsList: list.New(),
	}
	module.Social = mod
	registerRpc()
	eventhub.Subscribe(consts.EvMaxPvpLevelUpdate, onMaxPvpLevelUpdate)
	eventhub.Subscribe(consts.EvFixServer1WxInvite, onFixServer1WxInvite)

	timer.AddTicker(5*time.Second, func() {
		now := time.Now()
		elem := mod.inviteRoomsList.Front()
		for elem != nil {
			r := elem.Value.(*inviteRoom)
			if r.checkTimeout(now) {
				delete(mod.inviteRooms, r.inviter)
				mod.inviteRoomsList.Remove(elem)
				elem = mod.inviteRoomsList.Front()
				if rs, ok := mod.beInviteRooms[r.target]; ok {
					delete(rs, r.inviter)
				}
			} else {
				break
			}
		}
	})
}
