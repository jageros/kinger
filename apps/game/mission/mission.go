package mission

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module"
	"kinger/proto/pb"
	"kinger/gamedata"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/common/glog"
	"kinger/common/consts"
	"fmt"
)

type missionSt struct {
	attr *attribute.MapAttr
}

func newMissionByAttr(attr *attribute.MapAttr) *missionSt {
	return &missionSt{
		attr: attr,
	}
}

func newMission(mdata *gamedata.Mission, id int) *missionSt {
	attr := attribute.NewMapAttr()
	attr.SetInt("missionID", mdata.ID)
	attr.SetInt("id", id)
	return &missionSt{
		attr: attr,
	}
}

func (m *missionSt) String() string {
	return fmt.Sprintf("[mid=%d, gid=%d]", m.attr.GetInt("missionID"), m.attr.GetInt("id"))
}

func (m *missionSt) packMsg() *pb.Mission {
	msg := &pb.Mission{
		MissionID: int32(m.attr.GetInt("missionID")),
		CurCnt: int32(m.attr.GetInt("cnt")),
		IsReward: m.attr.GetBool("isReward"),
		ID: int32(m.attr.GetInt("id")),
	}
	return msg
}

func (m *missionSt) getMissionTpl() iMissionTemplate {
	return templates[m.attr.GetInt("missionID")]
}

func (m *missionSt) isReward() bool {
	return m.attr.GetBool("isReward")
}

func (m *missionSt) getMissionID() int {
	return m.attr.GetInt("missionID")
}

func (m *missionSt) getID() int {
	return m.attr.GetInt("id")
}

func (m *missionSt) getReward(player types.IPlayer) (jade, gold, bowlder int, err error) {
	if m.isReward() {
		return 0, 0, 0, gamedata.GameError(1)
	}
	tpl := m.getMissionTpl()
	mdata := tpl.getData()
	if m.attr.GetInt("cnt") < mdata.Process {
		return 0, 0, 0, gamedata.GameError(2)
	}

	m.attr.SetBool("isReward", true)

	glog.Infof("mission get reward, uid=%d, missionID=%d, jade=%d, gold=%s", player.GetUid(), mdata.ID,
		mdata.Jade, mdata.Gold)
	module.Player.LogMission(player, fmt.Sprintf("doMission_%d",  mdata.ID), 2)

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	resReward := map[int]int{}
	if mdata.Jade > 0 {
		jade = mdata.Jade + funcPrice.Team2MissionExtReward[player.GetPvpTeam()]
		resReward[consts.Jade] = jade
	}
	if mdata.Bowlder > 0 {
		bowlder = mdata.Bowlder + funcPrice.Team2MissionExtReward[player.GetPvpTeam()]
		resReward[consts.Bowlder] = bowlder
	}
	if mdata.Gold > 0 {
		gold = mdata.Gold + funcPrice.Team2MissionExtReward[player.GetPvpTeam()]
		resReward[consts.Gold] = gold
	}
	resCpt.BatchModifyResource(resReward, consts.RmrMission)
	return
}

func (m *missionSt) isComplete() bool {
	tpl := m.getMissionTpl()
	mdata := tpl.getData()
	curCnt := m.attr.GetInt("cnt")
	return curCnt >= mdata.Process
}

func (m *missionSt) addCnt(cnt int) (bool, bool) {
	isComplete := false
	tpl := m.getMissionTpl()
	mdata := tpl.getData()
	oldCnt := m.attr.GetInt("cnt")
	newCnt := oldCnt + cnt
	if newCnt >= mdata.Process {
		isComplete = true
		newCnt = mdata.Process
	}
	m.attr.SetInt("cnt", newCnt)

	return isComplete, oldCnt != newCnt
}
