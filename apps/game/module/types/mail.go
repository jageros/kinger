package types

import (
	"kinger/proto/pb"
)

type IMailReward interface {
	AddGold(amount int)
	AddJade(amount int)
	AddCard(cardID uint32, amount int)
	AddItem(rewardType pb.MailRewardType, itemID string, amount int)
	PackMsg() []*pb.MailReward
	AddAmountByType(rewardType pb.MailRewardType, amount int)
	AddEmoji(emojiTeam int)
}

type IMailComponent interface {
	IPlayerComponent
	SendMail(title, content, sender string, t int, reward IMailReward, mailType pb.MailTypeEnum, args []byte)
}

type IMailSender interface {
	// 设置邮件标题、内容，只有 mailType == pb.MailTypeEnum_CUSTOM 才有效
	SetTitleAndContent(title, content string)
	// 设置邮件类型、参数，类型不能为 pb.MailTypeEnum_CUSTOM
	SetTypeAndArgs(mailType pb.MailTypeEnum, args ...interface{})
	// 设置哪种accountType能收到邮件，uid == 0 （全服） 才有效
	SetAccountType(accountType pb.AccountTypeEnum)
	// 设置过期时间，在这个时间建号的玩家能收到，uid == 0 （全服） 才有效
	SetNewbieDeadLine(t int64)
	GetRewardObj() IMailReward
	Send() int
	// 设置哪个区的玩家能收到，uid == 0 （全服） 才有效
	SetArea(area int)
}
