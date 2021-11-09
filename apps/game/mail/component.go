package mail

import (
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"time"
)

var (
	_ types.IMailComponent     = &mailComponent{}
	_ types.ICrossDayComponent = &mailComponent{}
)

type mailComponent struct {
	player    types.IPlayer
	isNewbie  bool
	attr      *attribute.MapAttr
	mailsAttr *attribute.ListAttr
	mails     []*mailSt
	id2Mail   map[int]*mailSt
}

func (mc *mailComponent) ComponentID() string {
	return consts.MailCpt
}

func (mc *mailComponent) GetPlayer() types.IPlayer {
	return mc.player
}

func (mc *mailComponent) OnInit(player types.IPlayer) {
	mc.player = player
	mailsAttr := mc.attr.GetListAttr("mails")
	if mailsAttr == nil {
		mailsAttr = attribute.NewListAttr()
		mc.attr.SetListAttr("mails", mailsAttr)
	}
	mc.mailsAttr = mailsAttr
	mc.id2Mail = map[int]*mailSt{}

	delIndex := -1
	now := int(time.Now().Unix())
	mailsAttr.ForEachIndex(func(index int) bool {
		attr := mailsAttr.GetMapAttr(index)
		ml := newMailByAttr(player, attr)
		if ml.isTimeout(now) {
			return true
		} else if delIndex < 0 {
			delIndex = index
		}
		mc.mails = append(mc.mails, ml)
		mc.id2Mail[ml.getID()] = ml
		return true
	})

	if delIndex > 0 {
		mailsAttr.DelBySection(0, delIndex)
	}
}

func (mc *mailComponent) OnLogin(isRelogin, isRestore bool) {
	if isRelogin {
		return
	}
	mod.onPlayerLogin(mc, mc.isNewbie)
	mc.isNewbie = false
	if mc.attr.GetBool("hasNew") {
		timer.AfterFunc(2*time.Second, func() {
			if mc.attr.GetBool("hasNew") {
				agent := mc.player.GetAgent()
				if agent != nil {
					agent.PushClient(pb.MessageID_S2C_NOTIFY_NEW_MAIL, nil)
				}
			}
		})
	}
}

func (mc *mailComponent) OnLogout() {

}

func (mc *mailComponent) OnCrossDay(curDayno int) {
	delIndex := -1
	now := int(time.Now().Unix())
	for i, ml := range mc.mails {
		if ml.isTimeout(now) {
			delete(mc.id2Mail, ml.getID())
			continue
		}

		if i > 0 {
			delIndex = i
			break
		}
	}

	if delIndex > 0 {
		mc.mails = mc.mails[delIndex:]
		mc.mailsAttr.DelBySection(0, delIndex)
	}
}

func (mc *mailComponent) genMailID() int {
	id := mc.attr.GetInt("maxID") + 1
	mc.attr.SetInt("maxID", id)
	return id
}

func (mc *mailComponent) getMail(id int) *mailSt {
	return mc.id2Mail[id]
}

func (mc *mailComponent) getMailList() []*mailSt {
	mc.attr.SetBool("hasNew", false)
	return mc.mails
}

func (mc *mailComponent) sendMailByID(mailID int, title, content, sender string, t int, reward types.IMailReward,
	mailType pb.MailTypeEnum, args []byte) {

	m := newMail(mc.player, mailID)
	m.setArgs(mailType, args)
	m.setTitle(title)
	m.setContent(content)
	m.setTime(t)
	reward2, ok := reward.(*mailReward)
	if ok && reward2 != nil {
		m.setReward(reward2)
	}
	mc.mails = append(mc.mails, m)
	mc.id2Mail[m.getID()] = m
	mc.mailsAttr.AppendMapAttr(m.attr)
	if !mc.attr.GetBool("hasNew") {
		mc.attr.SetBool("hasNew", true)

		timer.AfterFunc(150*time.Millisecond, func() {
			agent := mc.player.GetAgent()
			if agent != nil {
				agent.PushClient(pb.MessageID_S2C_NOTIFY_NEW_MAIL, nil)
			}
		})
	}
	glog.Infof("sendMail uid=%d, mail=%s", mc.player.GetUid(), m)
}

func (mc *mailComponent) SendMail(title, content, sender string, t int, reward types.IMailReward, mailType pb.MailTypeEnum,
	args []byte) {
	mc.sendMailByID(mc.genMailID(), title, content, sender, t, reward, mailType, args)
}

func (mc *mailComponent) sendWholeServerMail(mailID int, title, content, sender string, t int, reward types.IMailReward,
	mailType pb.MailTypeEnum, args []byte) {
	serverMailID := mc.attr.GetInt("serverMailID")
	if serverMailID >= mailID {
		return
	}
	mc.attr.SetInt("serverMailID", mailID)
	mc.sendMailByID(mc.genMailID(), title, content, sender, t, reward, mailType, args)
}

func (mc *mailComponent) getServerMailID() int {
	return mc.attr.GetInt("serverMailID")
}

func (mc *mailComponent) setServerMailID(mailID int) {
	if mailID > mc.getServerMailID() {
		mc.attr.SetInt("serverMailID", mailID)
	}
}
