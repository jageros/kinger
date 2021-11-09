package wxgame

import (
	"container/list"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"time"
)

var mod *wxgameModule

type wxgameModule struct {
	inviteRooms     map[common.UUid]*list.Element
	inviteRoomsList *list.List
}

func (wm *wxgameModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("wxgame")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("wxgame", attr)
	}
	return &wxgameComponent{attr: attr}
}

func (wm *wxgameModule) newInviteBattleRoom(uid common.UUid) {
	r := newInviteRoom(uid)
	elem := wm.inviteRoomsList.PushBack(r)
	if elem2, ok := wm.inviteRooms[uid]; ok {
		wm.inviteRoomsList.Remove(elem2)
	}
	wm.inviteRooms[uid] = elem
}

func (wm *wxgameModule) CancelInviteBattle(uid common.UUid) {
	if elem, ok := wm.inviteRooms[uid]; ok {
		wm.inviteRoomsList.Remove(elem)
		delete(wm.inviteRooms, uid)
	}
}

func (wm *wxgameModule) IsInInviteBattle(uid common.UUid) bool {
	_, ok := wm.inviteRooms[uid]
	return ok
}

func (wm *wxgameModule) GetDailyShareState(player types.IPlayer) int {
	return player.GetComponent(consts.WxgameCpt).(*wxgameComponent).getDailyShareState()
}

func (wm *wxgameModule) ReturnDailyShareReward(player types.IPlayer, playerName string) {
	player.GetComponent(consts.WxgameCpt).(*wxgameComponent).returnDailyShareReward(playerName)
}

func (wm *wxgameModule) replyInviteBattle(inviter common.UUid, beInviterData *pb.FighterData) bool {
	if elem, ok := wm.inviteRooms[inviter]; ok {
		p := module.Player.GetPlayer(inviter)
		if p == nil || module.Campaign.IsInCampaignMatch(p) || module.Social.IsInInviteBattle(inviter) {
			return false
		}

		r := elem.Value.(*inviteRoom)
		wm.inviteRoomsList.Remove(elem)
		delete(wm.inviteRooms, inviter)
		return r.beginBattle(beInviterData)
	} else {
		return false
	}
}

func checkInviteTimeout() {
	now := time.Now()
	elem := mod.inviteRoomsList.Front()
	for elem != nil {
		r := elem.Value.(*inviteRoom)
		if r.isTimeout(now) {
			delete(mod.inviteRooms, r.inviter)
			mod.inviteRoomsList.Remove(elem)
			elem = mod.inviteRoomsList.Front()
		} else {
			return
		}
	}
}

func everyDayTask() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		player.GetComponent(consts.WxgameCpt).OnLogin(false, false)
	})
}

func Initialize() {
	mod = &wxgameModule{
		inviteRooms:     make(map[common.UUid]*list.Element),
		inviteRoomsList: list.New(),
	}
	module.WxGame = mod
	registerRpc()
	parseTreasureShareHD()
	//evq.HandleEvent(consts.EvReloadConfig, reloadConfig)
	timer.AddTicker(5*time.Second, checkInviteTimeout)
	timer.RunEveryDay(0, 0, 1, everyDayTask)
}
