package mail

import (
	"kinger/apps/game/module/types"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
)

var _ types.IMailSender = &mailSender{}

type mailSender struct {
	uid            common.UUid
	isSend         bool
	reward         types.IMailReward
	accountType    pb.AccountTypeEnum
	mailType       pb.MailTypeEnum
	title          string
	content        string
	newbieDeadLine int64
	area           int
	args           []interface{}
}

func (ms *mailSender) SetArea(area int) {
	if ms.uid > 0 {
		return
	}
	ms.area = area
}

func (ms *mailSender) SetTitleAndContent(title, content string) {
	ms.title = title
	ms.content = content
	ms.mailType = pb.MailTypeEnum_CUSTOM
}

func (ms *mailSender) SetTypeAndArgs(mailType pb.MailTypeEnum, args ...interface{}) {
	ms.mailType = mailType
	ms.args = args
}

func (ms *mailSender) SetAccountType(accountType pb.AccountTypeEnum) {
	if ms.uid > 0 {
		return
	}
	ms.accountType = accountType
}

func (ms *mailSender) SetNewbieDeadLine(t int64) {
	if ms.uid > 0 {
		return
	}
	ms.newbieDeadLine = t
}

func (ms *mailSender) GetRewardObj() types.IMailReward {
	if ms.reward == nil {
		ms.reward = newMailReward()
	}
	return ms.reward
}

func (ms *mailSender) Send() int {
	if ms.isSend {
		return 0
	}
	ms.isSend = true

	var reward types.IMailReward = nil
	if ms.reward != nil {
		reward = ms.reward
	}

	if ms.uid == 0 {
		return mod.sendMailToAllPlayers(ms.title, ms.content, reward, ms.accountType, ms.newbieDeadLine, ms.area,
			ms.mailType, ms.encodeArgs())
	} else {
		mod.sendMailToPlayer(ms.uid, ms.title, ms.content, reward, ms.mailType, ms.encodeArgs())
		return 0
	}
}

func (ms *mailSender) encodeArgs() []byte {
	var payload []byte
	if len(ms.args) <= 0 {
		return payload
	}

	switch ms.mailType {
	case pb.MailTypeEnum_SeasonPvpBegin:
		arg, _ := (&pb.MailSeasonPvpBeginArg{
			PvpLevel: int32(ms.args[0].(int)),
		}).Marshal()
		return arg

	case pb.MailTypeEnum_SeasonPvpEnd:
		arg, _ := (&pb.MailSeasonPvpEndArg{
			WinDiff: int32(ms.args[0].(int)),
		}).Marshal()
		return arg

	case pb.MailTypeEnum_CampaignUnified:
		arg, _ := (&pb.MailCampaignUnifiedArg{
			YourMajestyName: ms.args[0].(string),
			CountryName:     ms.args[1].(string),
		}).Marshal()
		return arg

	case pb.MailTypeEnum_RankHonorReward:
		arg, _ := (&pb.MailRankHonorRewardArg{
			Rank:  int32(ms.args[0].(int)),
			Honor: ms.args[1].(int32),
		}).Marshal()
		return arg
	case pb.MailTypeEnum_BackPrivReimburse:
		arg, _ := (&pb.ReimburseBackMailPrivID{
			PrivID: int32(ms.args[0].(int)),
		}).Marshal()
		return arg
	case pb.MailTypeEnum_LeagueEnd:
		arg, _ := (&pb.LeagueSeasonEndArg{
			OldScore: int32(ms.args[0].(int)),
			NewScore: int32(ms.args[1].(int)),
			Rank:     int32(ms.args[2].(int)),
		}).Marshal()
		return arg
	default:
		return payload
	}
}
