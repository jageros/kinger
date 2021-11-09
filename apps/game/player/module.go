package player

import (
	"errors"
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/apps/center/mq"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	gconfig "kinger/gopuppy/common/config"
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/lru"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/gopuppy/network"
	kpb "kinger/proto/pb"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var mod *playerModule

type onlineInfoLog struct {
	onlineAmount int
	onlineTime   int
	accountType  string
	area         int
}

func (ol *onlineInfoLog) MarshalLogObject(encoder glog.ObjectEncoder) error {
	encoder.AddInt("onlineAmount", ol.onlineAmount)
	encoder.AddInt("totalOnlineTime", ol.onlineTime)
	encoder.AddString("accountType", ol.accountType)
	encoder.AddInt("area", ol.area)
	return nil
}

type playerModule struct {
	playersMap                  map[common.UUid]types.IPlayer
	playerCache                 *lru.LruCache
	loadingPlayer               map[common.UUid]chan struct{}
	accountID2BindAccountRegion map[string]uint32
	pong                        *kpb.Pong
}

func newPlayerModule() *playerModule {
	c, err := lru.NewLruCache(1000, func(key interface{}, value interface{}) {
		uid := key.(common.UUid)
		if mod.GetPlayer(uid) == nil {
			api.DelClientDispatchInfo(uid, logic.GetAgentRegion(uid))
		}
	})
	if err != nil {
		panic(err)
	}

	m := &playerModule{
		playersMap:                  make(map[common.UUid]types.IPlayer),
		playerCache:                 c,
		loadingPlayer:               map[common.UUid]chan struct{}{},
		accountID2BindAccountRegion: map[string]uint32{},
		pong:                        &kpb.Pong{ServerTime: int32(time.Now().Unix())},
	}
	m.beginHeartBeat()
	m.loadBindAccountRegion()
	return m
}

func (m *playerModule) loadBindAccountRegion() {
	attrs, err := attribute.LoadAll("bindAccountRegion", true)
	if err != nil {
		glog.Errorf("loadBindAccountRegion %s", err)
		return
	}

	for _, attr := range attrs {
		accountID := attr.GetAttrID().(string)
		m.accountID2BindAccountRegion[accountID] = attr.GetUInt32("region")
	}
}

func (m *playerModule) onBindAccount(accountID string, region uint32) {
	m.accountID2BindAccountRegion[accountID] = region
}

func (m *playerModule) getBindRegion(accountID string) uint32 {
	return m.accountID2BindAccountRegion[accountID]
}

func (m *playerModule) saveAllPlayer() {
	for _, p := range m.playersMap {
		p2 := p.(*Player)
		if attr := p2.getAttr(); attr != nil {
			if attr.Dirty() {
				attr.Save(false)
				glog.Debugf("save player %d attr", p.GetUid())
			}
		}
	}
}

func (m *playerModule) dumpOnlineInfo() {
	area2TypeOnlineInfo := map[int]map[kpb.AccountTypeEnum]*onlineInfoLog{}
	cfgs := gconfig.GetConfig().GetLogicConfigsByName(gconsts.AppGame)
	for _, cfg := range cfgs {
		if cfg.GetAppID() == 1 {
			continue
		}

		reply, err := logic.CallBackend(gconsts.AppGame, uint32(cfg.GetAppID()), kpb.MessageID_G2G_GET_ONLINE_INFO, nil)
		if err == nil {
			reply2 := reply.(*kpb.OnlineInfo)
			for _, msg := range reply2.Infos {
				type2OnlineInfo, ok := area2TypeOnlineInfo[int(msg.Area)]

				if !ok {
					type2OnlineInfo = map[kpb.AccountTypeEnum]*onlineInfoLog{}
					info := &onlineInfoLog{}
					type2OnlineInfo[msg.AccountType] = info
					area2TypeOnlineInfo[int(msg.Area)] = type2OnlineInfo
				}

				info, ok := type2OnlineInfo[msg.AccountType]

				if !ok {
					info = &onlineInfoLog{}
					type2OnlineInfo[msg.AccountType] = info
				}

				info.onlineTime += int(msg.TotalOnlineTime)
				info.onlineAmount += int(msg.PlayerAmount)
				info.accountType = msg.AccountType.String()
				info.area = int(msg.Area)
			}
		}
	}

	onlineInfos := m.getOnlineInfo()
	for area, type2OInfo := range onlineInfos {
		for accountType, oInfo := range type2OInfo {
			type2OnlineInfo, ok := area2TypeOnlineInfo[area]
			if !ok {
				type2OnlineInfo = map[kpb.AccountTypeEnum]*onlineInfoLog{}
				info := &onlineInfoLog{}
				type2OnlineInfo[accountType] = info
				area2TypeOnlineInfo[area] = type2OnlineInfo
			}

			info, ok := type2OnlineInfo[accountType]
			if !ok {
				info = &onlineInfoLog{}
				type2OnlineInfo[accountType] = info
			}

			info.onlineTime += oInfo.onlineTime
			info.onlineAmount += oInfo.onlineAmount
			info.accountType = accountType.String()
			info.area = area
		}
	}

	var areaAccountTypeInfos []glog.ObjectMarshaler
	for _, info := range area2TypeOnlineInfo {
		for _, info2 := range info {
			areaAccountTypeInfos = append(areaAccountTypeInfos, info2)
		}
	}
	glog.JsonInfo("online", glog.Objects("AreaAccountTypes", areaAccountTypeInfos))
}

func (m *playerModule) beginHeartBeat() {
	timer.AfterFunc(time.Duration(rand.Intn(50)+10)*time.Second, func() {
		timer.AddTicker(5*time.Minute, m.saveAllPlayer)
	})

	if module.Service.GetAppID() == 1 {
		timer.AddTicker(time.Minute, m.dumpOnlineInfo)
	}
}

func (m *playerModule) addPlayer(player *Player) {
	m.playersMap[player.GetUid()] = player
}

func (m *playerModule) delPlayer(uid common.UUid) {
	delete(m.playersMap, uid)
}

func (m *playerModule) onPlayerLogout(uid common.UUid) {
	player, ok := m.playersMap[uid]
	if ok {
		player2 := player.(*Player)
		if player2.isNetAlive() {
			//glog.Infof("onPlayerLogout uid=%d", uid)
			player2.onNetDisconnect()
		}
		delete(m.playersMap, uid)
		module.Social.OnLogout(player)
	}
}

func (m *playerModule) onPlayerLogin(player *Player, isRelogin, isRestore bool) bool {
	ok := player.onLogin()
	if !ok {
		return false
	}

	for _, cpt := range player.componentList {
		cpt.OnLogin(isRelogin, isRestore)
	}

	player.onCrossDay(timer.GetDayNo())
	doCompensate(player, isRelogin)

	return true
}

func (m *playerModule) GetPlayer(uid common.UUid) types.IPlayer {
	if p, ok := m.playersMap[uid]; ok {
		return p
	} else {
		return nil
	}
}

func (m *playerModule) GetAllPlayer() []types.IPlayer {
	var players []types.IPlayer
	for _, p := range m.playersMap {
		players = append(players, p)
	}
	return players
}

func (m *playerModule) GetPlayerMap() map[common.UUid]types.IPlayer {
	return m.playersMap
}

func (m *playerModule) isNameExist(name string) bool {
	exist, err := attribute.NewAttrMgr("names", name, true).Exists()
	if err != nil {
		return true
	}
	return exist
}

func (m *playerModule) addName(name string, uid common.UUid) {
	attr := attribute.NewAttrMgr("names", name, true)
	attr.SetUInt64("uid", uint64(uid))
	attr.Save(false)
}

func (m *playerModule) delName(name string) {
	attribute.NewAttrMgr("names", name, true).Delete(false)
}

func (m *playerModule) GetSimplePlayerInfo(uid common.UUid) *kpb.SimplePlayerInfo {
	if p, ok := m.playersMap[uid]; ok {
		return p.(*Player).packSimpleMsg()
	}

	if p, ok := m.playerCache.Get(uid); ok {
		return p.(*Player).packSimpleMsg()
	}

	payload, err := logic.LoadPlayer(uid)
	if err != nil {
		return nil
	}

	msg := &kpb.SimplePlayerInfo{}
	err = msg.Unmarshal(payload)
	if err != nil {
		return nil
	}

	return msg
}

func (m *playerModule) LoadSimplePlayerInfoAsync(uid common.UUid) chan *kpb.SimplePlayerInfo {
	c := make(chan *kpb.SimplePlayerInfo, 1)
	resultChan := logic.LoadPlayerAsync(uid)
	go func() {
		payload := <-resultChan
		if payload == nil {
			c <- nil
			return
		}

		msg := &kpb.SimplePlayerInfo{}
		err := msg.Unmarshal(payload)
		if err != nil {
			c <- nil
			return
		}

		c <- msg
	}()

	return c
}

func (m *playerModule) addCachePlayer(p *Player) {
	uid := p.GetUid()
	if _, ok := m.playersMap[uid]; ok {
		return
	}

	if _, ok := m.playerCache.Get(uid); ok {
		return
	}

	p.agent = nil
	m.playerCache.Add(uid, p)
}

func (m *playerModule) GetCachePlayer(uid common.UUid) types.IPlayer {
	if p, ok := m.playerCache.Get(uid); ok {
		return p.(*Player)
	} else {
		return nil
	}
}

func (m *playerModule) delCachePlayer(uid common.UUid) {
	m.playerCache.RemoveWithoutCallback(uid)
	if c, ok := m.loadingPlayer[uid]; ok {
		evq.Await(func() {
			<-c
		})
		m.playerCache.RemoveWithoutCallback(uid)
	}
}

func (m *playerModule) NewPlayerByAttr(uid common.UUid, attr *attribute.AttrMgr) types.IPlayer {
	return newPlayer(uid, nil, attr)
}

func (m *playerModule) IsWxgameAccount(accountType kpb.AccountTypeEnum) bool {
	return accountType == kpb.AccountTypeEnum_Wxgame || accountType == kpb.AccountTypeEnum_WxgameIos
}

func (m *playerModule) getNewbieNamePrefix() string {
	cfg := config.GetConfig()
	if cfg.IsMultiLan || cfg.IsXfMultiLan {
		return "guest"
	} else {
		return "新兵"
	}
}

func (m *playerModule) ModifyResource(player types.IPlayer, resType int, amount int, args ...interface{}) {
	player.GetComponent(consts.ResourceCpt).(*ResourceComponent).ModifyResource(resType, amount, args...)
}

func (m *playerModule) GetResource(player types.IPlayer, resType int) int {
	return player.GetComponent(consts.ResourceCpt).(*ResourceComponent).GetResource(resType)
}

func (m *playerModule) SetResource(player types.IPlayer, resType, amount int) {
	player.GetComponent(consts.ResourceCpt).(*ResourceComponent).SetResource(resType, amount)
}

func (m *playerModule) createNewbiePlayer(uid common.UUid, agent *logic.PlayerAgent, channel,
	channelID string, area int) *Player {

	now := time.Now().Unix()
	createTime := uint32(now)
	playerAttr := attribute.NewAttrMgr("player", uid)
	playerAttr.SetUInt32("createTime", createTime)
	playerAttr.SetStr("serverID", utils.TimeToString(int64(createTime), utils.TimeFormat1))
	playerAttr.SetStr("channel", channel)
	playerAttr.SetStr("channelID", channelID)
	playerAttr.SetStr("name", genNewbieName(uid))
	playerAttr.SetInt("lastOnlineTime", int(now))

	if area <= 0 {
		area = gamedata.GetGameData(consts.AreaConfig).(*gamedata.AreaConfigGameData).GetCurArea().Area
	}
	playerAttr.SetInt("area", area)

	playerAttr.SetDirty(true)
	playerAttr.Save(true)
	p := newPlayer(uid, agent, playerAttr)
	p.setCompensateVersion(compensateVersion)
	resCpt := p.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	resCpt.SetResource(consts.Gold, 500)
	resCpt.SetResource(consts.AccTreasureCnt, 30)
	if config.GetConfig().IsMultiLan {
		resCpt.SetResource(consts.Bowlder, 50)
	} else {
		//resCpt.SetResource(consts.Jade, 50)
	}

	region := agent.GetRegion()
	if region <= 0 {
		region = module.Service.GetRegion()
	}
	logic.SaveAgentRegion(uid, region)

	return p
}

func (m *playerModule) getOnlineInfo() map[int]map[kpb.AccountTypeEnum]*onlineInfoLog {
	reply := map[int]map[kpb.AccountTypeEnum]*onlineInfoLog{}
	now := time.Now().Unix()
	for _, p := range m.playersMap {
		p2 := p.(*Player)
		area := p2.GetArea()
		accountType := p2.GetLogAccountType()
		accountTypeInfo, ok := reply[area]
		if !ok {
			accountTypeInfo = map[kpb.AccountTypeEnum]*onlineInfoLog{}
			info := &onlineInfoLog{}
			accountTypeInfo[accountType] = info
			reply[area] = accountTypeInfo
		}

		info, ok := accountTypeInfo[accountType]
		if !ok {
			info := &onlineInfoLog{}
			accountTypeInfo[accountType] = info
		}

		info.onlineAmount++
		loginTime := p2.getLoginTime()
		if loginTime > 0 && loginTime < now {
			info.onlineTime += int(now - loginTime)
		}
	}
	return reply
}

func (m *playerModule) loadPlayer(uid common.UUid, loadDb bool) ([]byte, error) {
	player := mod.GetPlayer(uid)
	var payload []byte
	var err error
	if player == nil {
		cPlayer := mod.GetCachePlayer(uid)
		if cPlayer == nil {
			if !loadDb {
				return payload, errors.New("doLoadPlayer not found")
			}

			if c, ok := m.loadingPlayer[uid]; ok {
				evq.Await(func() {
					<-c
				})
				return m.loadPlayer(uid, false)
			}

			c := make(chan struct{})
			m.loadingPlayer[uid] = c
			defer func() {
				delete(m.loadingPlayer, uid)
				close(c)
			}()

			region := logic.GetAgentRegion(uid)
			playerAttr := attribute.NewAttrMgr("player", uid, false, region)
			err = playerAttr.Load()
			if err != nil {
				glog.Errorf("rpc_LoadPlayer uid=%d, err=%s", uid, err)
				return nil, network.InternalErr
			}

			cPlayer2 := newPlayer(uid, nil, playerAttr)
			msg := cPlayer2.packSimpleMsg()
			payload, err = msg.Marshal()
			if err != nil {
				return nil, err
			}

			mod.addCachePlayer(cPlayer2)
		} else {
			msg := cPlayer.(*Player).packSimpleMsg()
			payload, err = msg.Marshal()
		}
	} else {
		msg := player.(*Player).packSimpleMsg()
		payload, err = msg.Marshal()
	}

	return payload, err
}

func (m *playerModule) LogMission(player types.IPlayer, missionID string, event int) {
	glog.JsonInfo("mission", glog.Uint64("uid", uint64(player.GetUid())), glog.String("channel",
		player.GetChannel()), glog.String("accountType", player.GetAccountType().String()), glog.String(
		"missionID", missionID), glog.Int("event", event), glog.Int("area", player.GetArea()),
		glog.String("subChannel", player.GetSubChannel()))
}

func (m *playerModule) HasResource(p types.IPlayer, resType, amount int) bool {
	return p.GetComponent(consts.ResourceCpt).(*ResourceComponent).HasResource(resType, amount)
}

func (m *playerModule) PackUpdateRankMsg(p types.IPlayer, battleHandCards []*kpb.SkinGCard, battleCamp int) *kpb.UpdatePvpScoreArg {
	resCpt := p.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	return &kpb.UpdatePvpScoreArg{
		Uid:            uint64(p.GetUid()),
		Name:           p.GetName(),
		HandCards:      battleHandCards,
		Camp:           int32(battleCamp),
		PvpScore:       int32(resCpt.GetResource(consts.Score)),
		WinDiff:        int32(resCpt.GetResource(consts.WinDiff)),
		WinCnt:         int32(module.Huodong.GetSeasonPvpWinCnt(p)),
		RebornCnt:      int32(module.Reborn.GetRebornCnt(p)),
		Area:           int32(p.GetArea()),
		CrossAreaHonor: int32(resCpt.GetResource(consts.CrossAreaHonor)),
		RankScore:      int32(resCpt.GetResource(consts.MatchScore)),
	}
}

func (m *playerModule) getResourceMaxAmount(resType int) int {
	switch resType {
	case consts.AccTreasureCnt:
		return 50
	//case consts.MatchScore:
	//	return 4000
	default:
		return math.MaxInt32
	}
}

func (m *playerModule) getResourceMinAmount(resType int, player types.IPlayer) int {
	if resType == consts.CrossAreaHonor {
		return math.MinInt32
	} else if resType == consts.MatchScore && player.GetPvpTeam() >= 9 {
		minScore := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetScoreById(1)
		return minScore
	} else {
		return 0
	}
}

func (m *playerModule) GetResourceName(resType int) string {
	if name, ok := needLogRes[resType]; ok {
		return name
	} else {
		return strconv.Itoa(resType)
	}
}

func (m *playerModule) ForEachOnlinePlayer(callback func(player types.IPlayer)) {
	for _, p := range m.playersMap {
		callback(p)
	}
}

func OnServerStop() {
	for _, p := range mod.playersMap {
		mq.RemoveConsumer(fmt.Sprintf("player:%d", p.GetUid()), module.Service.GetRegion())
	}

	mod.saveAllPlayer()
	mod.playersMap = map[common.UUid]types.IPlayer{}
}

func onCrossDay() {
	dayno := timer.GetDayNo()
	for _, player := range mod.playersMap {
		player2 := player.(*Player)
		for _, cpt := range player2.components {
			if cpt2, ok := cpt.(types.ICrossDayComponent); ok {
				cpt2.OnCrossDay(dayno)
			}
		}
		player2.onCrossDay(dayno)
	}

	if module.Service.GetAppID() == 1 {
		api.BroadcastClient(kpb.MessageID_S2C_ON_CROSS_DAY, nil, nil)
	}
}

func onHeartBeat() {
	mod.pong = &kpb.Pong{ServerTime: int32(time.Now().Unix())}
	for _, player := range mod.playersMap {
		player2 := player.(*Player)
		for _, cpt := range player2.components {
			if cpt2, ok := cpt.(types.IHeartBeatComponent); ok {
				cpt2.OnHeartBeat()
			}
		}
	}
}

func Initialize() {
	registerRpc()
	mod = newPlayerModule()
	module.Player = mod
	timer.RunEveryDay(0, 0, 1, onCrossDay)
	timer.AddTicker(10*time.Second, onHeartBeat)
	checkAccountCode()
	//fixRecharge(19750, "1fa8a8b512ace0aac94ea0bab29d7167", "197501552274490634748", "10035a0d1b38853b39d51jyohkcvns4o", 30)
}
