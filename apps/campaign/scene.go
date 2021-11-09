package main

import (
	"fmt"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"sort"
	"time"
	//"kinger/gopuppy/common/glog"
	//"encoding/json"
	"kinger/gopuppy/common/glog"
	"strconv"
)

var sceneMgr = &sceneMgrSt{}

type sceneMgrSt struct {
	// 支援、出征的队伍
	id2Team    map[int]*team
	id2DefTeam map[int]*team
	// 守这些城的队伍
	city2DefTeams map[int][]*team
	// 向这些城走去的队伍
	toCity2Teams map[int]map[int]*team
	// 在这些路上的队伍，map[roadID_targetCityID][]*team
	road2Team map[string]teamList
	// 当前每条路上同步给客户端的队伍
	tid2SyncTeam map[int]*team
	// 上次同步给客户端的守城玩家数量
	city2SyncDefAmount map[int]int
	ticker             *timer.Timer
}

func (sm *sceneMgrSt) initialize() {
	sm.id2Team = map[int]*team{}
	sm.id2DefTeam = map[int]*team{}
	sm.city2DefTeams = map[int][]*team{}
	sm.toCity2Teams = map[int]map[int]*team{}
	sm.road2Team = map[string]teamList{}
	sm.tid2SyncTeam = map[int]*team{}
	sm.city2SyncDefAmount = map[int]int{}
	attrs, err := attribute.LoadAll(warMgr.getTeamAttrName())
	if err != nil {
		panic(err)
	}

	var road2SyncTeamAttr *attribute.AttrMgr
	for _, attr := range attrs {
		_, ok := attr.GetAttrID().(string)
		if ok {
			road2SyncTeamAttr = attr
			continue
		}

		id, ok := attr.GetAttrID().(int)
		if !ok {
			panic(fmt.Sprintf("wrong teamID %s", attr.GetAttrID()))
		}

		t := newTeamByAttr(id, attr)
		if t != nil {
			sm.addTeam(t)
		}
	}

	if road2SyncTeamAttr != nil {
		road2SyncTeamAttr.ForEachKey(func(key string) {
			tid, _ := strconv.Atoi(key)
			t := sm.getTeam(tid)
			if t != nil {
				sm.tid2SyncTeam[tid] = t
			}
		})
	}

	if warMgr.isInWar() || warMgr.isReadyWar() {
		sm.onWarReady()
	}
	eventhub.Subscribe(evWarReady, sm.onWarReady)
	eventhub.Subscribe(evWarEnd, sm.onWarEnd)
	eventhub.Subscribe(evUnified, sm.onWarEnd)
}

func (sm *sceneMgrSt) getSyncClientTeams() map[int]*team {
	return sm.tid2SyncTeam
}

func (sm *sceneMgrSt) onWarReady(_ ...interface{}) {
	if sm.ticker != nil && sm.ticker.IsActive() {
		return
	}
	sm.ticker = timer.AddTicker(time.Second, sm.onHeartBeat)
}

func (sm *sceneMgrSt) onWarEnd(_ ...interface{}) {
	if sm.ticker == nil {
		return
	}
	sm.ticker.Cancel()
	sm.ticker = nil

	msg := &pb.CampaignTeams{}
	for _, t := range sm.tid2SyncTeam {
		tmsg := t.packMsg()
		tmsg.State = pb.TeamState_DisappearTS
		msg.Teams = append(msg.Teams, tmsg)
	}

	if len(msg.Teams) > 0 {
		campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_TEAMS, msg)
	}

	sm.id2Team = map[int]*team{}
	sm.id2DefTeam = map[int]*team{}
	sm.city2DefTeams = map[int][]*team{}
	sm.toCity2Teams = map[int]map[int]*team{}
	sm.road2Team = map[string]teamList{}
	sm.tid2SyncTeam = map[int]*team{}
	sm.city2SyncDefAmount = map[int]int{}
}

func (sm *sceneMgrSt) onHeartBeat() {
	if warMgr.isPause() {
		return
	}
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	var tids []int
	for tid, _ := range sm.id2Team {
		tids = append(tids, tid)
	}

	for _, tid := range tids {
		t, ok := sm.id2Team[tid]
		if !ok {
			continue
		}

		t.heartBeat(paramGameData)
	}

	warMgr.syncCityDefense()
	if warMgr.isInWar() {
		fieldMatchMgr.onMatchTick()
		cityMatchMgr.onMatchTick()
	}
	sm.syncTeam()
	sm.syncDefAmount()
}

func (sm *sceneMgrSt) getDefAmount() *pb.CitysDefPlayerAmount {
	if !warMgr.isInWar() {
		return nil
	}

	msg := &pb.CitysDefPlayerAmount{}
	for cityID, amount := range sm.city2SyncDefAmount {
		msg.Amounts = append(msg.Amounts, &pb.SyncCityPlayerAmount{
			CityID: int32(cityID),
			Amount: int32(amount),
		})
	}
	return msg
}

func (sm *sceneMgrSt) syncDefAmount() {
	city2SyncDefAmount := map[int]int{}
	for cityID, ts := range sm.city2DefTeams {
		city2SyncDefAmount[cityID] = len(ts)
	}

	msg := &pb.CitysDefPlayerAmount{}
	for cityID, amount := range city2SyncDefAmount {
		if sm.city2SyncDefAmount[cityID] != amount {
			msg.Amounts = append(msg.Amounts, &pb.SyncCityPlayerAmount{
				CityID: int32(cityID),
				Amount: int32(amount),
			})
		}
	}

	for cityID, _ := range sm.city2SyncDefAmount {
		if _, ok := city2SyncDefAmount[cityID]; !ok {
			msg.Amounts = append(msg.Amounts, &pb.SyncCityPlayerAmount{
				CityID: int32(cityID),
			})
		}
	}

	sm.city2SyncDefAmount = city2SyncDefAmount
	if len(msg.Amounts) > 0 {
		campaignMgr.broadcastClient(pb.MessageID_S2C_SYNC_DEF_CITY_PLAYER_AMOUNT, msg)
	}
}

func (sm *sceneMgrSt) syncTeam() {
	tid2SyncTeam := map[int]*team{}
	roadTeamAmount := map[string]int{}
	for roadKey, ts := range sm.road2Team {
		teamAmount := len(ts)
		if teamAmount <= 0 {
			continue
		}
		sort.Sort(ts)
		t := ts[0]
		tid2SyncTeam[t.getID()] = t
		roadTeamAmount[roadKey] = teamAmount
	}

	msg := &pb.CampaignTeams{}
	for tid, t := range tid2SyncTeam {
		oldt, ok := sm.tid2SyncTeam[tid]
		roadKey := t.getRoadKey()
		if ok {
			if t.dirty || oldt.teamAmount != roadTeamAmount[roadKey] {
				t.teamAmount = roadTeamAmount[roadKey]
				msg.Teams = append(msg.Teams, t.packMsg())
			}
		} else {
			t.teamAmount = roadTeamAmount[roadKey]
			msg.Teams = append(msg.Teams, t.packMsg())
		}

		t.dirty = false
	}

	for tid, t := range sm.tid2SyncTeam {
		if _, ok := tid2SyncTeam[tid]; !ok {
			oldTMsg := t.packMsg()
			oldTMsg.State = pb.TeamState_DisappearTS
			msg.Teams = append(msg.Teams, oldTMsg)
		}
	}

	sm.tid2SyncTeam = tid2SyncTeam

	//logMsg, _ := json.Marshal(msg)
	//glog.Infof("sync team %s", logMsg)

	if len(msg.Teams) > 0 {
		campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_TEAMS, msg)
	}
}

func (sm *sceneMgrSt) delToCityTeam(cityID int, t *team) {
	if ts, ok := sm.toCity2Teams[cityID]; ok {
		delete(ts, t.getID())
	}
}

func (sm *sceneMgrSt) addToCityTeam(cityID int, t *team) {
	ts, ok := sm.toCity2Teams[cityID]
	if !ok {
		ts = map[int]*team{}
		sm.toCity2Teams[cityID] = ts
	}
	ts[t.getID()] = t
}

func (sm *sceneMgrSt) addRoadTeam(t *team) {
	key := t.getRoadKey()
	if key == "" {
		return
	}
	ts := sm.road2Team[key]
	sm.road2Team[key] = append(ts, t)
}

func (sm *sceneMgrSt) delRoadTeam(t *team) {
	key := t.getRoadKey()
	if key == "" {
		return
	}
	ts := sm.road2Team[key]
	for i, t2 := range ts {
		if t.getID() == t2.getID() {
			sm.road2Team[key] = append(ts[:i], ts[i+1:]...)
			return
		}
	}
}

func (sm *sceneMgrSt) getTeam(tid int) *team {
	return sm.id2Team[tid]
}

func (sm *sceneMgrSt) getDefTeam(tid int) *team {
	return sm.id2DefTeam[tid]
}

func (sm *sceneMgrSt) addTeam(t *team) {
	if t.defCityID <= 0 {
		sm.id2Team[t.getID()] = t
	} else if _, ok := sm.id2DefTeam[t.getID()]; !ok {
		sm.id2DefTeam[t.getID()] = t
		ts := sm.city2DefTeams[t.defCityID]
		sm.city2DefTeams[t.defCityID] = append(ts, t)
	}
}

func (sm *sceneMgrSt) delTeam(t *team) {
	tid := t.getID()
	if _, ok := sm.id2Team[tid]; ok {
		delete(sm.id2Team, tid)
		for _, ts := range sm.toCity2Teams {
			delete(ts, tid)
		}
		targetCty := t.getTargetCity()
		if targetCty != nil {
			targetCty.delAttackTeam(t)
		}
	} else if _, ok := sm.id2DefTeam[tid]; ok {
		delete(sm.id2DefTeam, tid)
		ts := sm.city2DefTeams[t.defCityID]
		for i, t2 := range ts {
			if t == t2 {
				sm.city2DefTeams[t.defCityID] = append(ts[:i], ts[i+1:]...)
				break
			}
		}
	}

	sm.delRoadTeam(t)
	t.owner.delTeam(t)
	fieldMatchMgr.stopMatch(t.owner.getUid())
	cityMatchMgr.stopMatch(t.owner.getUid())

	//glog.Infof("del team %s", t)
}

func (sm *sceneMgrSt) save() {
	for _, t := range sm.id2Team {
		t.save()
	}
	for _, t := range sm.id2DefTeam {
		t.save()
	}

	road2SyncTeamAttr := attribute.NewAttrMgr(warMgr.getTeamAttrName(), "road2SyncTeam")
	for tid, t := range sm.tid2SyncTeam {
		road2SyncTeamAttr.SetInt(strconv.Itoa(tid), t.getID())
	}
	road2SyncTeamAttr.Save(true)
}

func (sm *sceneMgrSt) onCityBeOccupy(cty *city) {
	cityID := cty.getCityID()
	ts, ok := sm.toCity2Teams[cityID]
	if ok {
		delete(sm.toCity2Teams, cityID)
		for _, t := range ts {
			t.disappear(pb.MyTeamDisappear_CityBeOccupy, true, cityID)
		}
	}

	ts2 := sm.city2DefTeams[cityID]
	sm.city2DefTeams[cityID] = []*team{}
	for _, t := range ts2 {
		sm.delTeam(t)
	}
}

type roadPath struct {
	distance     int
	r            *gamedata.Road
	targetCityID int
	key          string
}

func newRoadPath(distance, targetCityID int, r *gamedata.Road) *roadPath {
	return &roadPath{
		distance:     distance,
		r:            r,
		targetCityID: targetCityID,
	}
}

func newRoadPathByAttr(attr *attribute.MapAttr) *roadPath {
	return &roadPath{
		distance:     attr.GetInt("distance"),
		targetCityID: attr.GetInt("targetCityID"),
		r:            gamedata.GetGameData(consts.Road).(*gamedata.RoadGameData).ID2Road[attr.GetInt("roadID")],
	}
}

func (rp *roadPath) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("distance", rp.distance)
	attr.SetInt("roadID", rp.r.ID)
	attr.SetInt("targetCityID", rp.targetCityID)
	return attr
}

func (rp *roadPath) getKey() string {
	if rp.key == "" {
		rp.key = fmt.Sprintf("%d_%d", rp.r.ID, rp.targetCityID)
	}
	return rp.key
}

type teamList []*team

func (tl teamList) Less(i, j int) bool {
	t1 := tl[i]
	t2 := tl[j]
	if len(t1.paths) <= 0 {
		return false
	}
	if len(t2.paths) <= 0 {
		return false
	}

	d1 := t1.paths[0].distance - t1.curDistance
	d2 := t2.paths[0].distance - t2.curDistance
	if d1 > d2 {
		return false
	} else if d2 > d1 {
		return true
	}

	return t1.getID() <= t2.getID()
}

func (tl teamList) Swap(i, j int) {
	tl[j], tl[i] = tl[i], tl[j]
}

func (tl teamList) Len() int {
	return len(tl)
}

type team struct {
	id                  int
	type_               int
	forage              int
	owner               *player
	state               pb.TeamState
	cityPath            []int
	paths               []*roadPath
	distance            int
	curDistance         int
	targetCityCountryID uint32
	fighterData         *pb.FighterData
	isMatching          bool
	defCityID           int
	teamAmount          int  // 跟这个队伍在同一条路，同一个方向的队伍数量
	dirty               bool // 状态是否有了重要的改变，用于同步客户端
}

func newTeam(owner *player, type_, forage int, cityPath []int, paths []*roadPath, fighterData *pb.FighterData) *team {
	id := warMgr.genTeamID()
	targetCity := cityMgr.getCity(cityPath[len(cityPath)-1])
	t := newTeamByProp(id, type_, forage, owner, cityPath, paths, 0, paths[len(paths)-1].distance, pb.TeamState_NormalTS,
		targetCity.getCountryID(), fighterData, false, 0)
	sceneMgr.addTeam(t)
	glog.Infof("newTeam %s", t)
	return t
}

func newDefTeam(owner *player, fighterData *pb.FighterData, defCityID, forage int) *team {
	id := warMgr.genTeamID()
	t := newTeamByProp(id, ttDefCity, forage, owner, []int{}, []*roadPath{}, 0, 0, pb.TeamState_NormalTS,
		owner.getCountryID(), fighterData, false, defCityID)
	sceneMgr.addTeam(t)
	glog.Infof("newDefTeam %s", t)
	return t
}

func newTeamByProp(tid, type_, forage int, owner *player, cityPath []int, paths []*roadPath, curDistance, distance int,
	state pb.TeamState, targetCityCountryID uint32, fighterData *pb.FighterData, isMatching bool, defCityID int) *team {

	t := &team{
		id:                  tid,
		type_:               type_,
		owner:               owner,
		state:               state,
		cityPath:            cityPath,
		distance:            distance,
		fighterData:         fighterData,
		curDistance:         curDistance,
		targetCityCountryID: targetCityCountryID,
		isMatching:          isMatching,
		defCityID:           defCityID,
		forage:              forage,
	}

	index := -1
	pathsAmount := len(paths)
	for i, rpath := range paths {
		if curDistance >= rpath.distance && i != pathsAmount-1 {
			sceneMgr.delToCityTeam(rpath.targetCityID, t)
		} else {
			if index < 0 {
				index = i
			}
			sceneMgr.addToCityTeam(rpath.targetCityID, t)
		}
	}

	if index == 0 {
		t.paths = paths
	} else if index > 0 {
		t.paths = paths[index:]
	}

	if !t.isDefTeam() {
		sceneMgr.addRoadTeam(t)
	}

	return t
}

func newTeamByAttr(id int, attr *attribute.AttrMgr) *team {
	ownerUid := common.UUid(attr.GetUInt64("ownerUid"))
	owner := playerMgr.getPlayer(ownerUid)
	if owner == nil {
		return nil
	}

	var cityPath []int
	cityPathAttr := attr.GetListAttr("cityPath")
	cityPathAttr.ForEachIndex(func(index int) bool {
		cityPath = append(cityPath, cityPathAttr.GetInt(index))
		return true
	})

	ok := true
	var roadPaths []*roadPath
	roadPathsAttr := attr.GetListAttr("roadPaths")
	roadPathsAttr.ForEachIndex(func(index int) bool {
		rattr := roadPathsAttr.GetMapAttr(index)
		rp := newRoadPathByAttr(rattr)
		if rp.r == nil {
			ok = false
			return false
		}

		roadPaths = append(roadPaths, rp)
		return true
	})

	if !ok {
		return nil
	}

	var fighterData *pb.FighterData = nil
	fighterDataAttr := attr.GetStr("fighterData")
	if fighterDataAttr != "" {
		fighterData = &pb.FighterData{}
		err := fighterData.Unmarshal([]byte(fighterDataAttr))
		if err != nil {
			return nil
		}
	}

	t := newTeamByProp(id, attr.GetInt("type"), attr.GetInt("forage"), owner, cityPath, roadPaths,
		attr.GetInt("curDistance"), attr.GetInt("distance"), pb.TeamState(attr.GetInt("state")),
		attr.GetUInt32("targetCityCountryID"), fighterData, false, attr.GetInt("defCityID"))
	t.teamAmount = attr.GetInt("teamAmount")
	t.dirty = attr.GetBool("dirty")

	owner.initTeam(t)
	if t.getState() == pb.TeamState_AttackingCityTS {
		targetCty := t.getTargetCity()
		if targetCty != nil {
			t.beginAttack(false)
		}
	} else if t.isDefTeam() && t.getState() == pb.TeamState_NormalTS {
		t.beginDefCity()
	}
	return t
}

func (t *team) save() {
	attr := attribute.NewAttrMgr(warMgr.getTeamAttrName(), t.id)
	attr.SetUInt64("ownerUid", uint64(t.owner.getUid()))
	attr.SetInt("curDistance", t.curDistance)
	attr.SetInt("distance", t.distance)
	attr.SetInt("state", int(t.state))
	attr.SetUInt32("targetCityCountryID", t.targetCityCountryID)
	attr.SetInt("defCityID", t.defCityID)
	attr.SetInt("teamAmount", t.teamAmount)
	attr.SetBool("dirty", t.dirty)
	attr.SetInt("type", t.type_)
	attr.SetInt("forage", t.forage)

	cityPathAttr := attribute.NewListAttr()
	for _, cityID := range t.cityPath {
		cityPathAttr.AppendInt(cityID)
	}
	attr.SetListAttr("cityPath", cityPathAttr)

	roadPathsAttr := attribute.NewListAttr()
	for _, rp := range t.paths {
		roadPathsAttr.AppendMapAttr(rp.packAttr())
	}
	attr.SetListAttr("roadPaths", roadPathsAttr)
	attr.Save(true)
}

func (t *team) String() string {
	return fmt.Sprintf("[tid=%d, type=%s, uid=%d, state=%s, curDistance=%d, isMatching=%v, cityPath=%v]", t.id, t.type_,
		t.owner.getUid(), t.state, t.curDistance, t.isMatching, t.cityPath)
}

func (t *team) getForage() int {
	return t.forage
}

func (t *team) modifyForage(val int) {
	if val == 0 {
		return
	}
	cur := t.forage + val
	if cur < 0 {
		cur = 0
	}
	t.forage = cur

	if t.owner != nil && t.owner.isOnline() {
		t.owner.agent.PushClient(pb.MessageID_S2C_UPDATE_FORAGE, &pb.UpdateForageArg{
			ForageAmount: int32(cur),
		})
	}
}

func (t *team) setState(state pb.TeamState) {
	if t.state == state {
		return
	}
	t.state = state

	//glog.Infof("team setState %s %s", state, t)

	switch state {
	case pb.TeamState_NormalTS:
		fallthrough
	case pb.TeamState_DisappearTS:
		fallthrough
	case pb.TeamState_FieldBattleTS:
		fallthrough
	case pb.TeamState_AttackingCityTS:
		fallthrough
	case pb.TeamState_AtkCityBattleTS:
		t.dirty = true
	}
}

func (t *team) getRoadKey() string {
	if len(t.paths) > 0 {
		return t.paths[0].getKey()
	} else {
		return ""
	}
}

func (t *team) beginAttack(needSync bool) {
	t.setState(pb.TeamState_AttackingCityTS)
	t.isMatching = true

	if t.owner.isOnline() {
		t.owner.agent.PushClient(pb.MessageID_S2C_UPDATE_MY_TEAM_STATE, &pb.UpdateMyTeamStateArg{
			State: pb.TeamState_AttackingCityTS,
		})
	}

	targetCty := t.getTargetCity()
	if targetCty != nil {
		targetCty.addAttackTeam(t, needSync)
	}

	glog.Infof("beginAttack t=%s", t)

	cityMatchMgr.beginMatch(t.getTargetCity().getCityID(), t)
}

func (t *team) beginDefCity() {
	t.setState(pb.TeamState_NormalTS)
	t.isMatching = true

	if t.owner.isOnline() {
		t.owner.agent.PushClient(pb.MessageID_S2C_UPDATE_MY_TEAM_STATE, &pb.UpdateMyTeamStateArg{
			State: pb.TeamState_NormalTS,
		})
	}

	glog.Infof("beginDefCity t=%s", t)

	cityMatchMgr.beginMatch(t.defCityID, t)
}

func (t *team) isDefTeam() bool {
	return t.defCityID > 0
}

func (t *team) getID() int {
	return t.id
}

func (t *team) getOwner() *player {
	return t.owner
}

func (t *team) getDistance() int {
	return t.distance
}

func (t *team) getCurDistance() int {
	return t.curDistance
}

func (t *team) getTargetCity() *city {
	return cityMgr.getCity(t.cityPath[len(t.cityPath)-1])
}

func (t *team) getTargetCityID() int {
	return t.cityPath[len(t.cityPath)-1]
}

func (t *team) getState() pb.TeamState {
	return t.state
}

func (t *team) getFromCity() *city {
	return cityMgr.getCity(t.cityPath[0])
}

func (t *team) getFromCityID() int {
	return t.cityPath[0]
}

func (t *team) onFieldMatchDone() {
	t.setState(pb.TeamState_FieldBattleTS)
	t.isMatching = false
}

func (t *team) onCityMatchDone(isAttack bool) {
	if isAttack {
		t.setState(pb.TeamState_AtkCityBattleTS)
		targetCty := t.getTargetCity()
		if targetCty != nil {
			targetCty.delAttackTeam(t)
		}
	} else {
		t.setState(pb.TeamState_DefCityBattleTS)
	}
	t.isMatching = false
}

func (t *team) packMsg() *pb.TeamData {
	var cityPath []int32
	for _, cityID := range t.cityPath {
		cityPath = append(cityPath, int32(cityID))
	}
	return &pb.TeamData{
		ID:         int32(t.id),
		CountryID:  t.owner.getCountryID(),
		CityPath:   cityPath,
		Trip:       int32(t.curDistance),
		State:      t.state,
		TeamAmount: int32(t.teamAmount),
	}
}

func (t *team) retreat(needSync bool) *pb.TeamRetreat {
	if t.type_ == ttDefCity {
		sceneMgr.delTeam(t)
		return &pb.TeamRetreat{
			OldCity: int32(t.owner.getLocationCityID()),
			NewCity: int32(t.owner.getLocationCityID()),
		}
	} else {
		return t.disappear(pb.MyTeamDisappear_Retreat, needSync)
	}
}

func (t *team) disappear(reason pb.MyTeamDisappear_ReasonEnum, needSync bool, args ...interface{}) *pb.TeamRetreat {
	if t.type_ == ttDefCity {
		return nil
	}

	glog.Infof("team disappear, reason=%s, t=%s", reason, t)
	disappearMsg := &pb.MyTeamDisappear{Reason: reason}
	var ret *pb.TeamRetreat
	var targetCityID int

	t.setState(pb.TeamState_DisappearTS)
	switch reason {
	case pb.MyTeamDisappear_Retreat:
		// 主动撤退
		ret = &pb.TeamRetreat{OldCity: int32(t.getFromCityID())}
		fallthrough
	case pb.MyTeamDisappear_CityBeOccupy:
		// 去打的城被占领
		if len(args) > 0 {
			targetCityID = args[0].(int)
		}
		fallthrough
	//case pb.MyTeamRetreat_SupportCityBeOccupy:
	// 去支援的城被占领
	//	fallthrough
	case pb.MyTeamDisappear_NoForage:
		// 没粮
		toCity := t.owner.getLocationCity()
		if toCity == nil || toCity.getCountryID() != t.owner.getCountryID() {

			toCity = t.owner.getCity()
			if toCity == nil || toCity.getCountryID() != t.owner.getCountryID() {
				cry := t.owner.getCountry()
				if cry == nil {
					break
				}
				toCity = cry.randomCity()

				if toCity == nil || toCity.getCountryID() != t.owner.getCountryID() {
					// 下野，理应不会跑到这里
					t.owner.setJob(pb.CampaignJob_UnknowJob, pb.CampaignJob_UnknowJob, true)
					t.owner.setCity(0, 0, true)
					t.owner.setCountryID(0, true)
					break
				}
			}
		}

		toCity.onTeamEnter(t)
		if ret != nil {
			ret.NewCity = int32(toCity.getCityID())
		}

		if targetCityID == 0 {
			targetCityID = toCity.getCityID()
		}

	case pb.MyTeamDisappear_OccupyCity:
		// 占领目标城
		fallthrough
	case pb.MyTeamDisappear_EnterCity:
		targetCty := t.getTargetCity()
		targetCty.onTeamEnter(t)
		targetCityID = t.getTargetCityID()

	case pb.MyTeamDisappear_CountryDestory:
		// 国家被灭
		t.owner.setJob(pb.CampaignJob_UnknowJob, pb.CampaignJob_UnknowJob, true)
		t.owner.setCity(0, 0, true)
		t.owner.setCountryID(0, true)
	}

	switch reason {
	case pb.MyTeamDisappear_CityBeOccupy:
		fallthrough
	case pb.MyTeamDisappear_OccupyCity:
		fallthrough
	case pb.MyTeamDisappear_EnterCity:
		fallthrough
	case pb.MyTeamDisappear_NoForage:
		disappearMsg.Arg, _ = (&pb.TargetCity{CityID: int32(targetCityID)}).Marshal()
	}

	t.owner.onTeamDisappear(reason, t, disappearMsg, needSync)
	sceneMgr.delTeam(t)

	//disappearLog, _ := json.Marshal(disappearMsg)
	//glog.Infof("team disappear %s, reason=%s", t, disappearLog)

	return ret
}

func (t *team) onArriveTargetCity() {
	//glog.Infof("team onArriveTargetCity")
	targetCity := t.getTargetCity()
	if targetCity != nil && targetCity.getCountryID() == t.owner.getCountryID() {
		// 进城
		t.disappear(pb.MyTeamDisappear_EnterCity, true)
		return
	}

	if targetCity == nil || targetCity.getCountryID() != t.targetCityCountryID {
		// 撤退
		reason := pb.MyTeamDisappear_CityBeOccupy
		t.disappear(reason, true, t.getTargetCityID())
		return
	}

	// 攻城
	t.beginAttack(true)
}

func (t *team) march(paramGameData *gamedata.CampaignParamGameData) {
	if t.curDistance < t.distance {
		t.curDistance += paramGameData.MarchSpeed
		index := -1
		pathsAmount := len(t.paths)
		for i, rpath := range t.paths {
			if t.curDistance < rpath.distance || i == pathsAmount-1 {
				index = i
				break
			} else {
				sceneMgr.delToCityTeam(rpath.targetCityID, t)
			}
		}

		if index > 0 {
			sceneMgr.delRoadTeam(t)
			t.paths = t.paths[index:]
			sceneMgr.addRoadTeam(t)
		}

		if t.curDistance >= t.distance {
			// 到达目的地
			fieldMatchMgr.stopMatch(t.owner.getUid())
			t.isMatching = false
		} else if t.type_ == ttExpedition && !t.isMatching {
			rpath := t.paths[0]
			fieldMatchMgr.beginMatch(rpath.r.ID, t)
			t.isMatching = true

			/*
				// 前面的城如果是敌城，匹配
				cty := cityMgr.getCity(rpath.targetCityID)
				if cty != nil && cty.getCountryID() != t.owner.getCountryID() {
					// 匹配
					if !t.isMatching {
						fieldMatchMgr.beginMatch(rpath.r.ID, t)
						t.isMatching = true
					}
				} else {
					// 取消匹配
					fieldMatchMgr.stopMatch(t.owner.getUid())
					t.isMatching = false
				}
			*/
		}
	}

	//glog.Infof("team march %s", t)

	if t.curDistance >= t.distance {
		t.curDistance = t.distance
		t.onArriveTargetCity()
	}
}

func (t *team) attackCity(paramGameData *gamedata.CampaignParamGameData) {
	//glog.Infof("team attack %s", t)
	targetCity := t.getTargetCity()
	countryID := targetCity.getCountryID()
	if countryID != t.targetCityCountryID {
		// 攻击中的城被占领
		if countryID == t.owner.getCountryID() {
			// 进城
			t.disappear(pb.MyTeamDisappear_OccupyCity, true)
		} else {
			// 撤退
			t.disappear(pb.MyTeamDisappear_CityBeOccupy, true, t.getTargetCityID())
		}
		return
	}

	if targetCity.isBeOccupy() {
		return
	}

	if targetCity.attack(t, len(sceneMgr.city2DefTeams[targetCity.getCityID()])) {
		t.disappear(pb.MyTeamDisappear_OccupyCity, true)
		sceneMgr.onCityBeOccupy(targetCity)
	}
	t.owner.addContribution(paramGameData.SingleMerit*t.owner.getCityGlory(), true, false)
}

func (t *team) onBattleEnd(isWin bool) {
	glog.Infof("onBattleEnd %s", t)
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	switch t.getState() {
	case pb.TeamState_FieldBattleTS:
		t.setState(pb.TeamState_FieldBattleEndTS)
		if isWin {
			t.owner.addContribution(paramGameData.SingleEncounterVic*t.owner.getCityGlory(), true, false)
		} else {
			t.owner.addContribution(paramGameData.SingleLoseVic*t.owner.getCityGlory(), true, false)
			forage := t.getForage()
			if forage <= 0 {
				// 没粮，撤退
				t.owner.beginRestState()
				t.disappear(pb.MyTeamDisappear_NoForage, true)
			} else {
				t.owner.beginRectifyState()
				t.modifyForage(-1)
			}
		}

	case pb.TeamState_AtkCityBattleTS:
		t.setState(pb.TeamState_CanAttackCityTS)
		if isWin {
			t.owner.addContribution(paramGameData.SingleAttackVic*t.owner.getCityGlory(), true, false)
		} else {
			t.owner.addContribution(paramGameData.SingleLoseVic*t.owner.getCityGlory(), true, false)
			forage := t.getForage()
			if forage <= 0 {
				// 没粮，撤退
				t.owner.beginRestState()
				t.disappear(pb.MyTeamDisappear_NoForage, true)
			} else {
				t.owner.beginRectifyState()
				t.modifyForage(-1)
			}
		}

	case pb.TeamState_DefCityBattleTS:
		t.setState(pb.TeamState_DefCityBattleEndTS)
		if isWin {
			t.owner.addContribution(paramGameData.SingleAttackVic*t.owner.getCityGlory(), true, false)
		} else {
			t.owner.addContribution(paramGameData.SingleLoseVic*t.owner.getCityGlory(), true, false)
			forage := t.getForage()
			if forage <= 0 {
				// 没粮，撤退
				t.owner.beginRestState()
				sceneMgr.delTeam(t)
			} else {
				t.owner.beginRectifyState()
				t.modifyForage(-1)
			}
		}
	}
}

func (t *team) heartBeat(paramGameData *gamedata.CampaignParamGameData) {
	switch t.state {
	case pb.TeamState_NormalTS:
		// 向前走
		t.march(paramGameData)

	case pb.TeamState_AttackingCityTS:
		// 正在攻城
		t.attackCity(paramGameData)
	}
}
