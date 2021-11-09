package social

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"time"
)

var nilAdvertProtecter iAdvertProtecter = &nilAdvertProtecterSt{}

// 反广告

type iAdvertProtecter interface {
	onChat(msg string) (isAdvert bool)
}

func newAdvertProtecter(player types.IPlayer, cptAttr *attribute.MapAttr) iAdvertProtecter {
	if player.GetMaxPvpLevel() > advertProtectMaxPvpLevel {
		cptAttr.Del("adPter")
		return nilAdvertProtecter
	}

	attr := cptAttr.GetMapAttr("adPter")
	if attr == nil {
		attr = attribute.NewMapAttr()
		cptAttr.SetMapAttr("adPter", attr)
	}

	return &advertProtecterSt{
		player: player,
		attr:   attr,
	}
}

type nilAdvertProtecterSt struct {
}

func (ap *nilAdvertProtecterSt) onChat(msg string) (isAdvert bool) {
	return
}

type advertProtecterSt struct {
	player types.IPlayer
	attr   *attribute.MapAttr
}

func (ap *advertProtecterSt) getLastMsg() string {
	return ap.attr.GetStr("lastMsg")
}

func (ap *advertProtecterSt) setLastMsg(msg string) {
	ap.attr.SetStr("lastMsg", msg)
}

func (ap *advertProtecterSt) getLastChatDay() int {
	return ap.attr.GetInt("lastDay")
}

func (ap *advertProtecterSt) setLastChatDay(dayno int) {
	ap.attr.SetInt("lastDay", dayno)
}

func (ap *advertProtecterSt) getAdvertCnt() int {
	return ap.attr.GetInt("advertCnt")
}

func (ap *advertProtecterSt) setAdvertCnt(cnt int) {
	ap.attr.SetInt("advertCnt", cnt)
}

func (ap *advertProtecterSt) sendToSelf(msg string) {
	chat := &pb.ChatItem{
		Uid:         uint64(ap.player.GetUid()),
		Name:        ap.player.GetName(),
		HeadImgUrl:  ap.player.GetHeadImgUrl(),
		Time:        int32(time.Now().Unix()),
		Msg:         msg,
		PvpLevel:    int32(ap.player.GetPvpLevel()),
		Country:     ap.player.GetCountry(),
		HeadFrame:   ap.player.GetHeadFrame(),
		ChatPop:     ap.player.GetChatPop(),
		CountryFlag: ap.player.GetCountryFlag(),
	}

	clet := &pb.Chatlet{Type: pb.Chatlet_Normal}
	clet.Data, _ = chat.Marshal()
	chatMsg := &pb.ChatNotify{
		Channel: pb.ChatChannel_World,
		Chat:    clet,
	}
	agent := ap.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_CHAT_NOTIFY, chatMsg)
	}
}

func (ap *advertProtecterSt) onChat(msg string) bool {
	st := module.OutStatus.GetStatus(ap.player, consts.OtAdvertProtecter)
	alreadyInForbid := st != nil
	if alreadyInForbid {
		ap.sendToSelf(msg)
	}

	if st != nil && st.GetRemainTime() < 0 {
		return true
	}

	today := timer.GetDayNo()
	lastMsg := ap.getLastMsg()
	lastDay := ap.getLastChatDay()

	ap.setLastChatDay(today)
	ap.setLastMsg(msg)

	if lastMsg == "" || today != lastDay {
		return alreadyInForbid
	}

	isAdvert2 := float64(utils.EditDistance(lastMsg, msg)) <= float64(len(msg))*0.25
	if !isAdvert2 && mod.lastAdvertChat != "" {
		isAdvert2 = float64(utils.EditDistance(mod.lastAdvertChat, msg)) <= float64(len(msg))*0.25
	}
	if !isAdvert2 {
		return alreadyInForbid
	}

	advertCnt := ap.getAdvertCnt() + 1
	ap.setAdvertCnt(advertCnt)

	var remainTime int
	if st == nil {
		remainTime = 10
		module.OutStatus.AddStatus(ap.player, consts.OtAdvertProtecter, remainTime)
	} else {
		st.Over(30)
		remainTime = st.GetRemainTime()
	}

	if remainTime >= 60 || advertCnt >= 3 {
		module.OutStatus.DelStatus(ap.player, consts.OtAdvertProtecter)
		module.OutStatus.AddStatus(ap.player, consts.OtAdvertProtecter, -1)
	}

	if !alreadyInForbid {
		ap.sendToSelf(msg)
	}
	mod.onSendAdvertChat(msg)
	return true
}
