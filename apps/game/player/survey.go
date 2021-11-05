package player

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
)

var _ types.IPlayerComponent = &surveyComponent{}

type surveyComponent struct {
	attr   *attribute.AttrMgr
	player *Player
}

func newSurveyComponent(player *Player) *surveyComponent {
	return &surveyComponent{
		player: player,
	}
}

func (sc *surveyComponent) OnLogin(isRelogin, isRestore bool) {
}

func (sc *surveyComponent) OnLogout() {
}

func (sc *surveyComponent) ComponentID() string {
	return consts.SurveyCpt
}

func (sc *surveyComponent) GetPlayer() types.IPlayer {
	return sc.player
}

func (sc *surveyComponent) OnInit(player types.IPlayer) {

}

func (sc *surveyComponent) loadAttr() {
	if sc.attr != nil {
		return
	}
	attr := attribute.NewAttrMgr("survey2", sc.player.GetUid())
	err := attr.Load()
	if sc.attr != nil {
		return
	}
	if err == attribute.NotExistsErr {
		sc.attr = attr
		attr.Save(false)
	} else if err != nil {
		glog.Errorf("load Survey error, uid=%d, err=%s")
	} else {
		sc.attr = attr
	}
}

func (sc *surveyComponent) packMsg() *pb.SurveyInfo {
	sc.loadAttr()
	msg := &pb.SurveyInfo{}
	if sc.attr == nil {
		msg.IsComplete = true
		msg.IsReward = true
	} else {
		msg.IsComplete = sc.attr.GetListAttr("answers") != nil
		msg.IsReward = sc.attr.GetBool("isReward")
	}
	return msg
}

func (sc *surveyComponent) answer(answers *pb.SurveyAnswer) {
	sc.loadAttr()
	if sc.attr == nil || sc.attr.GetListAttr("answers") != nil {
		return
	}
	answersAttr := attribute.NewListAttr()
	for _, a := range answers.Answers {
		aAttr := attribute.NewMapAttr()
		aAttr.SetInt("questionID", int(a.QuestionID))
		lAttr := attribute.NewListAttr()
		for _, l := range a.AnswerIDs {
			lAttr.AppendInt(int(l))
		}
		aAttr.SetListAttr("answerIDs", lAttr)
		answersAttr.AppendMapAttr(aAttr)
	}
	sc.attr.SetListAttr("answers", answersAttr)
	sc.attr.Save(false)
}

func (sc *surveyComponent) getReward() *pb.OpenTreasureReply {
	sc.loadAttr()
	if sc.attr == nil || sc.attr.GetBool("isReward") || sc.attr.GetListAttr("answers") == nil {
		return nil
	}

	sc.attr.SetBool("isReward", true)
	sc.attr.Save(false)
	treasureReward := sc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(
		"BX6004", false)
	glog.Infof("survey get reward %d", sc.player.GetUid())
	return treasureReward
}
