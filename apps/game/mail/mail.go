package mail

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/proto/pb"
	"kinger/gamedata"
	"fmt"
	"time"
)

const mailTimeout = 7 * 24 * 60 * 60

var _ types.IMailReward = &mailReward{}

type mailSt struct {
	attr *attribute.MapAttr
	player types.IPlayer
	reward *mailReward
}

func newMailByAttr(player types.IPlayer, attr *attribute.MapAttr) *mailSt {
	m := &mailSt{
		attr: attr,
		player: player,
	}
	rewardsAttr := attr.GetListAttr("rewards")
	if rewardsAttr != nil {
		m.reward = newMailRewardByAttr(rewardsAttr)
		m.reward.mailID = m.getID()
	}
	return m
}

func newMail(player types.IPlayer, id int) *mailSt {
	attr := attribute.NewMapAttr()
	attr.SetInt("id", id)
	return &mailSt{
		attr: attr,
		player: player,
	}
}

func newMailByMsg(msg *pb.Mail) *mailSt {
	m := newMail(nil, int(msg.ID))
	m.setReward( newMailRewardByMsg(msg.Rewards) )
	m.setArgs(msg.MailType, msg.Arg)
	m.setTime(int(msg.Time))
	m.setTitle(msg.Title)
	m.setContent(msg.Content)
	m.setSender(msg.SenderName)
	return m
}

func (m *mailSt) String() string {
	return fmt.Sprintf("[mail id=%d, t=%d, reward=%s]", m.getID(), m.getTime(), m.reward)
}

func (m *mailSt) setArgs(mailType pb.MailTypeEnum, args []byte) {
	m.attr.SetInt("mailType", int(mailType))
	if args != nil {
		m.attr.SetStr("args", string(args))
	}
}

func (m *mailSt) getMailType() pb.MailTypeEnum {
	return pb.MailTypeEnum(m.attr.GetInt("mailType"))
}

func (m *mailSt) getArgs() []byte {
	return []byte(m.attr.GetStr("args"))
}

func (m *mailSt) getID() int {
	return m.attr.GetInt("id")
}

func (m *mailSt) getTime() int {
	return m.attr.GetInt("time")
}

func (m *mailSt) setTime(t int) {
	m.attr.SetInt("time", t)
}

func (m *mailSt) isTimeout(now int) bool {
	return m.getTime() + mailTimeout <= now
}

func (m *mailSt) getTitle() string {
	return m.attr.GetStr("title")
}

func (m *mailSt) setTitle(title string) {
	m.attr.SetStr("title", title)
}

func (m *mailSt) getContent() string {
	return m.attr.GetStr("content")
}

func (m *mailSt) setContent(content string) {
	m.attr.SetStr("content", content)
}

func (m *mailSt) isReward() bool {
	return m.attr.GetBool("isReward")
}

func (m *mailSt) isRead() bool {
	return m.attr.GetBool("isRead")
}

func (m *mailSt) read() {
	m.attr.SetBool("isRead", true)
}

func (m *mailSt) getSender() string {
	return m.attr.GetStr("sender")
}

func (m *mailSt) setSender(sender string) {
	m.attr.SetStr("sender", sender)
}

func (m *mailSt) setReward(reward *mailReward) {
	if reward == nil {
		return
	}
	m.attr.SetListAttr("rewards", reward.attrs)
	m.reward = reward
	reward.mailID = m.getID()
}

func (m *mailSt) packMsg() *pb.Mail {
	msg := &pb.Mail{
		ID: int32(m.getID()),
		Title: m.getTitle(),
		Content: m.getContent(),
		Time: int32(m.getTime()),
		IsReward: m.isReward(),
		SenderName: m.getSender(),
		IsRead: m.isRead(),
		MailType: pb.MailTypeEnum(m.attr.GetInt("mailType")),
		Arg: m.getArgs(),
	}
	if m.reward != nil {
		msg.Rewards = m.reward.PackMsg()
	}

	return msg
}

func (m *mailSt) getReward() (err error, amountRewards []*pb.MailRewardAmountArg, treasureRewards []*pb.OpenTreasureReply,
	cards []*pb.MailRewardCard, itemRewards []*pb.MailRewardItemArg, emojis []int32) {

	if m.isReward() {
		err = gamedata.GameError(1)
		return
	}
	if m.reward == nil {
		err = gamedata.GameError(2)
		return
	}

	amountRewards, treasureRewards, cards, itemRewards, emojis = m.reward.getReward(m.player)
	m.attr.SetBool("isReward", true)
	m.read()
	//m.attr.Del("rewards")
	//m.reward = nil
	return
}

type wholeServerMailSt struct {
	attr *attribute.AttrMgr
	m *mailSt
}

func newWholeServerMail(title, content string, reward types.IMailReward, accountType pb.AccountTypeEnum,
	newbieDeadLine int64, area int, mailType pb.MailTypeEnum, args []byte) *wholeServerMailSt {

	ml := newMail(nil, mod.genWholeServerMailID())
	ml.setArgs(mailType, args)
	ml.setTitle(title)
	ml.setContent(content)
	ml.setTime(int(time.Now().Unix()))
	reward2, ok := reward.(*mailReward)
	if ok && reward2 != nil {
		ml.setReward(reward2)
	}
	attr := attribute.NewAttrMgr("wholeServerMail", ml.getID(), true)
	attr.SetMapAttr("mailData", ml.attr)
	attr.SetInt("accountType", int(accountType))
	attr.SetInt64("newbieDeadLine", newbieDeadLine)
	attr.SetInt("area", area)
	attr.Save(true)
	return &wholeServerMailSt{
		attr: attr,
		m: ml,
	}
}

func newWholeServerMailByAttr(attr *attribute.AttrMgr) *wholeServerMailSt {
	mattr := attr.GetMapAttr("mailData")
	return &wholeServerMailSt{
		attr: attr,
		m: newMailByAttr(nil, mattr),
	}
}

func newWholeServerMailByMsg(msg *pb.WholeServerMail) *wholeServerMailSt {
	wm := &wholeServerMailSt{
		m: newMailByMsg(msg.MailData),
	}

	wm.attr = attribute.NewAttrMgr("wholeServerMail", wm.m.getID(), true)
	wm.attr.SetMapAttr("mailData", wm.m.attr)
	wm.attr.SetInt("accountType", int(msg.AccountType))
	wm.attr.SetInt64("newbieDeadLine", msg.NewbieDeadLine)
	wm.attr.SetInt("area", int(msg.Area))
	return wm
}

func (wm *wholeServerMailSt) packMsg() *pb.WholeServerMail {
	return &pb.WholeServerMail{
		MailData: wm.m.packMsg(),
		AccountType: wm.getAccountType(),
		NewbieDeadLine: wm.getNewbieDeadLine(),
		Area: int32(wm.getArea()),
	}
}

func (wm *wholeServerMailSt) getRewardAttr() *attribute.ListAttr {
	if wm.m == nil {
		return nil
	}
	if wm.m.reward == nil {
		return nil
	}
	return wm.m.reward.attrs
}

func (wm *wholeServerMailSt) getMailObj() *mailSt {
	return wm.m
}

func (wm *wholeServerMailSt) getNewbieDeadLine() int64 {
	return wm.attr.GetInt64("newbieDeadLine")
}

func (wm *wholeServerMailSt) setNewbieDeadLine(deadLine int64) {
	wm.attr.SetInt64("newbieDeadLine", deadLine)
}

func (wm *wholeServerMailSt) getArea() int {
	return wm.attr.GetInt("area")
}

func (wm *wholeServerMailSt) del() {
	wm.attr.Delete(false)
}

func (wm *wholeServerMailSt) isTimeout(now int64) bool {
	timeout := int64(wm.m.getTime()) + mailTimeout
	deadLine := wm.getNewbieDeadLine()
	if deadLine > timeout {
		timeout = deadLine
	}
	return now >= timeout
}

func (wm *wholeServerMailSt) getAccountType() pb.AccountTypeEnum {
	return pb.AccountTypeEnum(wm.attr.GetInt("accountType"))
}

func (wm *wholeServerMailSt) getRewardObj() types.IMailReward {
	if wm.m.reward == nil {
		return nil
	}
	return wm.m.reward
}

func (wm *wholeServerMailSt) save() {
	wm.attr.Save(true)
}
