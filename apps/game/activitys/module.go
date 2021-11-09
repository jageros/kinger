package activitys

import (
	"kinger/apps/game/activitys/consume"
	"kinger/apps/game/activitys/dailyrecharge"
	"kinger/apps/game/activitys/dailyshare"
	"kinger/apps/game/activitys/fight"
	"kinger/apps/game/activitys/firstrecharge"
	"kinger/apps/game/activitys/growplan"
	"kinger/apps/game/activitys/login"
	"kinger/apps/game/activitys/loginrecharge"
	"kinger/apps/game/activitys/online"
	"kinger/apps/game/activitys/rank"
	"kinger/apps/game/activitys/recharge"
	"kinger/apps/game/activitys/spring"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/activitys/win"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
)

var mod *activityModule
var _ aTypes.IActivity = &activity{}
var _ aTypes.IActivityMod = &activityModule{}

type activity struct {
	//所有活动
	activity *gamedata.Activity
	//活动的开启条件
	openCondition *gamedata.ActivityOpenCondition
	//活动的开启时间限制
	openTime *gamedata.ActivityTime
}

type activityModule struct {
	id2Activity    map[int]*activity
	type2Activitys map[int][]*activity
}

func (m *activityModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr(consts.ActivityCpt)
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr(consts.ActivityCpt, attr)
	}
	return &activityComponent{attr: attr}
}

func (m *activityModule) NewPCM(player types.IPlayer) aTypes.IPlayerCom {
	pcm := player.GetComponent(consts.ActivityCpt).(*activityComponent)
	return pcm
}

//从gameData中读取数据进来
func (m *activityModule) initializeAllData() {
	//活动数据
	actData := gamedata.GetGameData(consts.ActivityConfig)
	if actData == nil {
		err := gamedata.GameError(aTypes.GetGameDataError)
		glog.Errorf("initializeAllData get activity game data err=%s: ", err)
		return
	}
	actData.AddReloadCallback(m.initializeActivityData)
	m.initializeActivityData(actData)
}

func (m *activityModule) initializeActivityData(act gamedata.IGameData) {
	m.id2Activity = map[int]*activity{}
	m.type2Activitys = map[int][]*activity{}

	//活动时间数据
	tim := gamedata.GetGameData(consts.ActivityTime)
	if tim == nil {
		err := gamedata.GameError(aTypes.GetGameDataError)
		glog.Errorf("Init activity time condition get game data err=%s", err)
	}

	//活动开启条件数据
	con := gamedata.GetGameData(consts.ActivityOpenCondition)
	if con == nil {
		err := gamedata.GameError(aTypes.GetGameDataError)
		glog.Errorf("Init activity open condition get game data err=%s", err)
	}

	data := act.(*gamedata.ActivityGameData)
	if data == nil {
		err := gamedata.GameError(aTypes.GetGameDataError)
		glog.Errorf("Activity IGameData to activity game data err=%s", err)
		return
	}
	for aid, a := range data.ActivityMap {
		var tmc *gamedata.ActivityTime
		var opc *gamedata.ActivityOpenCondition
		if t, ok := tim.(*gamedata.ActivityTimeGameData); ok {
			if tm, ok := t.ActivityTimeMap[a.TimeID]; ok {
				tmc = tm
			}
		}

		if o, ok := con.(*gamedata.ActivityOpenConditionGameData); ok {
			if op, ok := o.ActivityOpenConditionMap[a.ConditionID]; ok {
				opc = op
			}
		}

		act := &activity{
			activity:      a,
			openTime:      tmc,
			openCondition: opc,
		}
		m.id2Activity[aid] = act

		activitys := m.type2Activitys[a.ActivityType]
		m.type2Activitys[a.ActivityType] = append(activitys, act)
	}
	initializeActivityDataAfterAddCallBack(m)
	updateTagListAfterReloadConfig()
}

func (m *activityModule) updateRewardDataCallBack() {
	for _, act := range m.id2Activity {
		tbName := act.GetRewardTableName()
		rwData := gamedata.GetGameData(tbName)
		if rwData != nil {
			rwData.AddReloadCallback(func(data gamedata.IGameData) {
				initializeActivityDataAfterAddCallBack(m)
			})
		}
	}
}

//根据活动ID获取单个活动数据
func (m *activityModule) GetActivityByID(activityID int) aTypes.IActivity {
	if val, ok := m.id2Activity[activityID]; ok {
		return val
	}
	return nil
}

func (m *activityModule) GetActivityIdList() []int {
	var idList []int
	for aid, _ := range m.id2Activity {
		idList = append(idList, aid)
	}
	return idList
}

func (m *activityModule) getActivityIdListByType(ty int) []int {
	var idList []int
	for aid, a := range m.id2Activity {
		if a.activity.ActivityType == ty {
			idList = append(idList, aid)
		}
	}
	return idList
}

func (m *activityModule) InitDataByType(t int) aTypes.IActivityMod {
	am := &activityModule{
		id2Activity:    map[int]*activity{},
		type2Activitys: map[int][]*activity{},
	}
	for k, a := range m.id2Activity {
		if a.activity.ActivityType == t {
			am.id2Activity[k] = a
		}
	}

	return am
}

func (m *activityModule) OnGetSpringHuodongItemType(player types.IPlayer, itemAmount int) int {
	return spring.OnGetTreasureRewardItemType(player, itemAmount)
}

func (m *activityModule) ForEachActivityDataByType(atype int, callback func(data aTypes.IActivity)) {
	if m.type2Activitys == nil {
		return
	}
	activitys := m.type2Activitys[atype]
	for _, a := range activitys {
		callback(a)
	}
}

//根据活动id获取限制时间
func (a *activity) GetTimeCondition() *gamedata.ActivityTime {
	return a.openTime
}

//根据活动id获取限制时间
func (a *activity) GetOpenCondition() *gamedata.ActivityOpenCondition {
	return a.openCondition
}

func (a *activity) GetRewardTableName() string {
	var tb string
	if a.activity != nil {
		tb = a.activity.RewardTable
	}
	return tb
}

func (a *activity) GetActivityType() int {
	if a != nil {
		if a.activity != nil {
			return a.activity.ActivityType
		}
	}
	return consts.ActivityUnknow
}

func (a *activity) GetActivityVersion() int {
	return a.activity.Version
}

func (a *activity) GetActivityId() int {
	return a.activity.ID
}

func (a *activity) GetGameData() *gamedata.Activity {
	return a.activity
}

func Initialize() {
	mod = &activityModule{}
	mod.initializeAllData()
	module.Activitys = mod
	initializeEvent()
	mod.updateRewardDataCallBack()
	registerRpc()
}

func updateTagListAfterReloadConfig() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		p := player.GetComponent(consts.ActivityCpt).(*activityComponent)
		p.initTagList(true)
	})
}

func initializeEvent() {
	login.AddEvent()
	recharge.AddEvent()
	online.AddEvent()
	fight.AddEvent()
	win.AddEvent()
	rank.AddEvent()
	consume.AddEvent()
	loginrecharge.AddEvent()
	firstrecharge.AddEvent()
	growplan.AddEvent()
	dailyrecharge.AddEvent()
	dailyshare.AddEvent()
}

func initializeActivityDataAfterAddCallBack(m *activityModule) {
	aTypes.IMod = m
	login.Initialize()
	recharge.Initialize()
	online.Initialize()
	fight.Initialize()
	win.Initialize()
	rank.Initialize()
	consume.Initialize()
	loginrecharge.Initialize()
	firstrecharge.Initialize()
	growplan.Initialize()
	spring.Initialize()
	dailyrecharge.Initialize()
	dailyshare.Initialize()
}
