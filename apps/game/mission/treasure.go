package mission

import (
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"math/rand"
)

type treasure struct {
	attr *attribute.MapAttr
}

func newTreasureByAttr(attr *attribute.MapAttr) *treasure {
	return &treasure{
		attr: attr,
	}
}

func newTreasure(player types.IPlayer) *treasure {
	gdata := gamedata.GetGameData(consts.MissionTreasure).(*gamedata.MissionTreasureGameData)
	tdata := gdata.MissionTreasures[rand.Intn(len(gdata.MissionTreasures))]
	attr := attribute.NewMapAttr()
	attr.SetInt("loopID", tdata.ID)
	rare := tdata.RareLoop[0]
	treasuresGameData := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData)
	team := player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpTeam()
	if team < 2 {
		team = 2
	}
	ts := treasuresGameData.TreasuresOfTeam[team]
	var t *gamedata.Treasure
	for _, t2 := range ts {
		if t2.Rare == rare {
			t = t2
			break
		}
	}

	if t == nil {
		return nil
	}

	attr.SetStr("id", t.ID)
	return &treasure{
		attr: attr,
	}
}

func (t *treasure) packMsg() *pb.MissionTreasure {
	return &pb.MissionTreasure{
		TreasureModelID: t.attr.GetStr("id"),
		CurCnt:          int32(t.attr.GetInt("cnt")),
	}
}

func (t *treasure) getReward(player types.IPlayer) (*pb.OpenTreasureReply, *treasure, error) {
	modelID := t.attr.GetStr("id")
	treasureGameData := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData)
	tdata := treasureGameData.Treasures[modelID]
	if t.attr.GetInt("cnt") < tdata.QuestUnlockCnt {
		return nil, nil, gamedata.GameError(1)
	}

	loopID := t.attr.GetInt("loopID")
	gdata := gamedata.GetGameData(consts.MissionTreasure).(*gamedata.MissionTreasureGameData)
	mtdata := gdata.ID2MissionTreasures[loopID]
	loopIdx := t.attr.GetInt("loopIdx")
	var newT *treasure
	if loopIdx >= len(mtdata.RareLoop)-1 {
		newT = newTreasure(player)
		if newT == nil {
			return nil, nil, gamedata.GameError(2)
		}
	} else {
		loopIdx++
		t.attr.SetInt("loopIdx", loopIdx)
		ts := treasureGameData.TreasuresOfTeam[player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpTeam()]
		var td *gamedata.Treasure
		rare := mtdata.RareLoop[loopIdx]
		for _, t2 := range ts {
			if t2.Rare == rare {
				td = t2
				break
			}
		}

		if td == nil {
			return nil, nil, gamedata.GameError(3)
		}

		t.attr.SetStr("id", td.ID)
		t.attr.SetInt("cnt", 0)
	}

	glog.Infof("mission treasure get reward, uid=%d, treasureModelID=%s", player.GetUid(), modelID)
	reward := player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(modelID, false)
	return reward, newT, nil
}

func (t *treasure) onMissionComplete(player types.IPlayer) bool {
	modelID := t.attr.GetStr("id")
	treasureGameData := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData)
	tdata := treasureGameData.Treasures[modelID]
	if tdata == nil {
		return false
	}
	cnt := t.attr.GetInt("cnt")
	if cnt >= tdata.QuestUnlockCnt {
		return false
	}

	cnt++
	if cnt >= tdata.QuestUnlockCnt {
		agent := player.GetAgent()
		if agent != nil {
			agent.PushClient(pb.MessageID_S2C_MISSION_SHOW_RED_DOT, nil)
		}
	}

	t.attr.SetInt("cnt", cnt)
	return true
}
