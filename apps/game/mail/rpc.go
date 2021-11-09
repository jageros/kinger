package mail

import (
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/network"
	"kinger/proto/pb"
)

func rpc_C2S_FetchMailList(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.FetchMailListArg)
	minMailID := int(arg2.MinMailID)

	reply := &pb.MailList{}
	mails := player.GetComponent(consts.MailCpt).(*mailComponent).getMailList()
	mailAmount := len(mails)
	i := mailAmount - 1

	for amount := 0; i >= 0 && amount < 10; i-- {
		m := mails[i]
		if minMailID > 0 && m.getID() >= minMailID {
			continue
		}
		amount++
		reply.Mails = append(reply.Mails, m.packMsg())
	}
	reply.HasMore = i > 0

	return reply, nil
}

func rpc_C2S_GetMailReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.GetMailRewardArg)
	m := player.GetComponent(consts.MailCpt).(*mailComponent).getMail(int(arg2.ID))
	if m == nil {
		return nil, gamedata.GameError(3)
	}

	err, amountRewards, treasureRewards, cards, itemRewards, emojis := m.getReward()
	if err != nil {
		return nil, err
	}
	return &pb.MailRewardReply{
		AmountRewards:   amountRewards,
		TreasureRewards: treasureRewards,
		Cards:           cards,
		ItemRewards:     itemRewards,
		EmojiTeams:      emojis,
	}, err
}

func rpc_C2S_ReadMail(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.ReadMailArg)
	m := player.GetComponent(consts.MailCpt).(*mailComponent).getMail(int(arg2.ID))
	if m == nil {
		return nil, gamedata.GameError(1)
	}
	m.read()

	return nil, nil
}

func rpc_C2S_GetAllReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply := &pb.MailRewardReply{}
	mails := player.GetComponent(consts.MailCpt).(*mailComponent).getMailList()
	for _, m := range mails {
		err, amountRewards, treasureRewards, cards, itemRewards, emojis := m.getReward()
		if err == nil {
			reply.TreasureRewards = append(reply.TreasureRewards, treasureRewards...)
			reply.EmojiTeams = append(reply.EmojiTeams, emojis...)
			reply.ItemRewards = append(reply.ItemRewards, itemRewards...)

		L1:
			for _, r2 := range amountRewards {
				for _, r := range reply.AmountRewards {
					if r.Type == r2.Type {
						r.Amount += r2.Amount
						continue L1
					}
				}
				reply.AmountRewards = append(reply.AmountRewards, r2)
			}

		L2:
			for _, c2 := range cards {
				for _, c := range reply.Cards {
					if c.CardID == c2.CardID {
						c.Amount += c2.Amount
						continue L2
					}
				}
				reply.Cards = append(reply.Cards, c2)
			}
		}
	}
	return reply, nil
}

func rpc_G2G_OnUpdateWholeServerMail(_ *network.Session, arg interface{}) (interface{}, error) {
	mod.onUpdateWholeServerMail(newWholeServerMailByMsg(arg.(*pb.WholeServerMail)))
	return nil, nil
}

func rpc_G2G_OnSendWholeServerMail(_ *network.Session, arg interface{}) (interface{}, error) {
	mod.onSendWholeServerMail(newWholeServerMailByMsg(arg.(*pb.WholeServerMail)))
	return nil, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_MAIL_LIST, rpc_C2S_FetchMailList)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_MAIL_REWARD, rpc_C2S_GetMailReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_READ_MAIL, rpc_C2S_ReadMail)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_ALL_MAIL_REWARD, rpc_C2S_GetAllReward)

	if module.Service.GetAppID() != 1 {
		logic.RegisterRpcHandler(pb.MessageID_G2G_ON_UPDATE_WHOLE_SERVER_MAIL, rpc_G2G_OnUpdateWholeServerMail)
		logic.RegisterRpcHandler(pb.MessageID_G2G_ON_SEND_WHOLE_SERVER_MAIL, rpc_G2G_OnSendWholeServerMail)
	}
}
