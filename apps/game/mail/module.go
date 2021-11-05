package mail

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/apps/game/module"
	"kinger/gopuppy/common"
	"kinger/common/consts"
	"time"
	"kinger/common/utils"
	"kinger/proto/pb"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/apps/logic"
)

var mod *mailModule

type mailModule struct {
	wholeServerMails []*wholeServerMailSt
	maxIDAttr *attribute.AttrMgr
}

func newMailModule() *mailModule {
	m := &mailModule{}
	var wholeServerMailAttrs []*attribute.AttrMgr
	for true {
		attrs, err := attribute.LoadAll("wholeServerMail", true)
		if err != nil {
			glog.Errorf("newMailModule LoadAll wholeServerMail %s", err)
			panic(err)
		}

		if len(attrs) > 0 {
			m.maxIDAttr = attrs[0]
			wholeServerMailAttrs = attrs[1:]
			break
		} else if module.Service.GetAppID() == 1 {
			maxIDAttr := attribute.NewAttrMgr("wholeServerMail", "maxID", true)
			maxIDAttr.SetDirty(true)
			maxIDAttr.Save(true)
			m.maxIDAttr = maxIDAttr
			break
		}

		time.Sleep(time.Second)
	}

	now := time.Now().Unix()
	for _, attr := range wholeServerMailAttrs {
		wm := newWholeServerMailByAttr(attr)
		if wm.isTimeout(now) {
			if module.Service.GetAppID() == 1 {
				wm.del()
			}
		} else {
			m.wholeServerMails = append(m.wholeServerMails, wm)
		}
	}

	glog.Infof("init wholeServerMailAttrs ok, %d", len(m.wholeServerMails))

	return m
}

func (m *mailModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("mail")
	isNewbie := false
	if attr == nil {
		isNewbie = true
		attr = attribute.NewMapAttr()
		//attr.SetInt("serverMailID", m.getWholeServerMailMaxID())
		playerAttr.SetMapAttr("mail", attr)
	}
	return &mailComponent{attr: attr, isNewbie: isNewbie}
}

func (m *mailModule) getWholeServerMailMaxID() int {
	if len(m.wholeServerMails) > 0 {
		return m.wholeServerMails[len(m.wholeServerMails) - 1].getMailObj().getID()
	}
	return 0
}

func (m *mailModule) genWholeServerMailID() int {
	id := m.maxIDAttr.GetInt("id")
	id++
	m.maxIDAttr.SetInt("id", id)
	m.maxIDAttr.Save(false)
	return id
}

func (m *mailModule) sendMailToOnlinePlayers(mailID int, title, content, sender string, t int, reward types.IMailReward,
	accountType pb.AccountTypeEnum, area int, mailType pb.MailTypeEnum, args []byte) {

	reward2, ok := reward.(*mailReward)
	isWxgameAccount := module.Player.IsWxgameAccount(accountType)
	module.Player.ForEachOnlinePlayer(func(p types.IPlayer) {

		if area > 0 && p.GetArea() != area {
			return
		}

		playerAccountType := p.GetAccountType()

		if accountType <= 0 || playerAccountType == accountType ||
			(module.Player.IsWxgameAccount(playerAccountType) && isWxgameAccount) {

			var rw types.IMailReward = nil
			if ok && reward2 != nil {
				rw = copyMailRewardFromAttr(reward2.attrs)
			}
			p.GetComponent(consts.MailCpt).(*mailComponent).sendWholeServerMail(mailID, title, content, sender, t, rw,
				mailType, args)
		}

	})
}

func (m *mailModule) onPlayerLogin(cpt *mailComponent, isNewbie bool) {
	serverMailID := cpt.getServerMailID()
	now := time.Now().Unix()
	player := cpt.GetPlayer()
	playerAccountType := player.GetAccountType()
	isWxgameAccount := module.Player.IsWxgameAccount(playerAccountType)
	var mails []*wholeServerMailSt

	for i := len(m.wholeServerMails) - 1; i >= 0; i-- {
		wm := m.wholeServerMails[i]
		if wm.isTimeout(now) {
			continue
		}

		ml := wm.getMailObj()
		if serverMailID >= ml.getID() {
			break
		}

		area := wm.getArea()
		if area > 0 && area != player.GetArea() {
			continue
		}

		accountType := wm.getAccountType()
		if accountType <= 0 || playerAccountType == accountType ||
			(module.Player.IsWxgameAccount(accountType) && isWxgameAccount) {

			if isNewbie {
				mailDeadLine := wm.getNewbieDeadLine()
				if mailDeadLine <= now {
					continue
				}
			}

			mails = append(mails, wm)
		}
	}

	for i := len(mails) - 1; i >= 0; i-- {
		wm := mails[i]
		var reward types.IMailReward = nil
		rewardAttr := wm.getRewardAttr()
		if rewardAttr != nil {
			reward = copyMailRewardFromAttr(rewardAttr)
		}
		ml := wm.getMailObj()
		id := ml.getID()
		t := ml.getTime()
		if wm.getNewbieDeadLine() > 0 {
			t = int(now)
		}
		cpt.sendWholeServerMail(id, ml.getTitle(), ml.getContent(), "", t, reward, ml.getMailType(), ml.getArgs())
	}

	cpt.setServerMailID(m.getWholeServerMailMaxID())
}

func (m *mailModule) sendMailToPlayer(uid common.UUid, title, content string, reward types.IMailReward,
	mailType pb.MailTypeEnum, args []byte) {

	player := module.Player.GetPlayer(uid)
	now := time.Now().Unix()
	if player != nil {
		player.GetComponent(consts.MailCpt).(*mailComponent).SendMail(title, content, "", int(now), reward,
			mailType, args)
	} else {
		mailMsg := &pb.Mail{
			Title: title,
			Content: content,
			Time: int32(now),
			Arg: args,
		}
		if reward != nil {
			mailMsg.Rewards = reward.PackMsg()
		}
		utils.PlayerMqPublish(uid, pb.RmqType_SendMail, mailMsg)
	}
}

func (m *mailModule) sendMailToAllPlayers(title, content string, reward types.IMailReward, accountType pb.AccountTypeEnum,
	newbieDeadLine int64, area int, mailType pb.MailTypeEnum, args []byte) int {

	if module.Service.GetAppID() != 1 {
		glog.Error("send whole server mail not in game1, title=%s, content=%s, reward=%s", title, content, reward)
		return 0
	}

	wm := newWholeServerMail(title, content, reward, accountType, newbieDeadLine, area, mailType, args)
	ml := wm.getMailObj()
	m.sendMailToOnlinePlayers(ml.getID(), title, content, "", ml.getTime(), wm.getRewardObj(), accountType, area,
		mailType, args)
	mod.wholeServerMails = append(mod.wholeServerMails, wm)

	logic.BroadcastBackend(pb.MessageID_G2G_ON_SEND_WHOLE_SERVER_MAIL, wm.packMsg())

	return ml.getID()
}

func (m *mailModule) NewMailRewardByMsg(msg []*pb.MailReward) types.IMailReward {
	return newMailRewardByMsg(msg)
}

func (m *mailModule) NewMailSender(uid common.UUid) types.IMailSender {
	return &mailSender{
		uid: uid,
	}
}

func (m *mailModule) UpdateMailDeadLine(mailID int, newbieDeadLine int64) bool {
	for _, wm := range m.wholeServerMails {
		if wm.getMailObj().getID() == mailID {
			wm.setNewbieDeadLine(newbieDeadLine)
			wm.save()

			logic.BroadcastBackend(pb.MessageID_G2G_ON_UPDATE_WHOLE_SERVER_MAIL, wm.packMsg())

			return true
		}
	}
	return false
}

func (m *mailModule) onSendWholeServerMail(wm *wholeServerMailSt) {
	ml := wm.getMailObj()
	mod.sendMailToOnlinePlayers(ml.getID(), ml.getTitle(), ml.getContent(), ml.getSender(), ml.getTime(), ml.reward,
		wm.getAccountType(), wm.getArea(), ml.getMailType(), ml.getArgs())
	mod.wholeServerMails = append(mod.wholeServerMails, wm)
}

func (m *mailModule) onUpdateWholeServerMail(wm *wholeServerMailSt) {
	id := wm.getMailObj().getID()
	for i, wm2 := range mod.wholeServerMails {
		if wm2.getMailObj().getID() == id {
			mod.wholeServerMails[i] = wm
			return
		}
	}
}

func Initialize() {
	mod = newMailModule()
	module.Mail = mod
	registerRpc()
}
