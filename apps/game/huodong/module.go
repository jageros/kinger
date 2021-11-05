package huodong

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/apps/game/module"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"time"
	"kinger/gopuppy/common/evq"
	"kinger/common/consts"
	"kinger/proto/pb"
	"kinger/gamedata"
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/huodong/seasonpvp"
	"kinger/apps/game/huodong/spring"
	"kinger/apps/game/huodong/event"
	"kinger/apps/game/huodong/springskin"
	"fmt"
	"kinger/gopuppy/apps/logic"
)

var mod *huodongModule

type huodongModule struct {
	allHuodong map[int]map[pb.HuodongTypeEnum]htypes.IHuodong
	refreshingHuodong map[int]map[pb.HuodongTypeEnum]chan struct{}
}

func (m *huodongModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("huodong")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("huodong", attr)
	}
	return &huodongComponent{attr: attr}
}

func (m *huodongModule) GetSeasonPvpLimitTime(player types.IPlayer) int32 {
	hd := m.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
	if hd != nil && hd.IsOpen() {
		limitTime := hd.GetEndTime().Sub(time.Now())
		limits := limitTime.Seconds()
		if limits < 0 {
			limits = 0
		}
		return int32(limits)
	} else {
		return 0
	}
}

func (m *huodongModule) AddHuodong(area int, hd htypes.IHuodong) {
	if hd == nil {
		return
	}
	type2Huodong, ok := m.allHuodong[area]
	if !ok {
		type2Huodong = map[pb.HuodongTypeEnum]htypes.IHuodong{}
		m.allHuodong[area] = type2Huodong
	}

	type2Huodong[hd.GetHtype()] = hd
}

func (m *huodongModule) getGameData(area int, htype pb.HuodongTypeEnum) interface{} {
	switch htype {
	case pb.HuodongTypeEnum_HSeasonPvp:
		data := gamedata.GetSeasonPvpGameData().GetSeasonData(area)
		if data == nil {
			return nil
		}
		return data
	case pb.HuodongTypeEnum_HSpringExchange:
		fallthrough
	case pb.HuodongTypeEnum_HSpringSkin:
		huodongGameData := gamedata.GetGameData(consts.HuodongConfig).(*gamedata.HuodongGameData)
		if huodongGameData.ID2Huodong == nil {
			return nil
		}

		data := huodongGameData.ID2Huodong[htype]
		if data == nil {
			return nil
		}
		return data
	default:
		return nil
	}
}

func (m *huodongModule) NewHuodong(area int, htype pb.HuodongTypeEnum) htypes.IHuodong {
	attr := attribute.NewAttrMgr(fmt.Sprintf("huodong%d", area), htype, true)
	var hd htypes.IHuodong = nil
	switch htype {
	case pb.HuodongTypeEnum_HSeasonPvp:
		hd = seasonpvp.NewSeasonPvpHd(area, attr, m.getGameData(area, htype))
	case pb.HuodongTypeEnum_HSpringExchange:
		hd = spring.NewSpringHd(area, attr, m.getGameData(area, htype))
	case pb.HuodongTypeEnum_HSpringSkin:
		hd = springskin.NewSpringSkinHd(area, attr, m.getGameData(area, htype))
	default:
		glog.Errorf("newHuodong what the fuck id=%d, area=%d", htype, area)
		return nil
	}
	return hd
}

func (m *huodongModule) newHuodongByAttr(area int, attr *attribute.AttrMgr) htypes.IHuodong {
	attrID := attr.GetAttrID()
	var htype pb.HuodongTypeEnum
	switch htype1 := attrID.(type) {
	case int:
		htype = pb.HuodongTypeEnum(htype1)
	case int32:
		htype = pb.HuodongTypeEnum(htype1)
	case pb.HuodongTypeEnum:
		htype = pb.HuodongTypeEnum(htype1)
	}

	switch htype {
	case pb.HuodongTypeEnum_HSeasonPvp:
		hd := seasonpvp.NewSeasonPvpHdByAttr(area, attr)
		return hd
	case pb.HuodongTypeEnum_HSpringExchange:
		hd := spring.NewSpringHdByAttr(area, attr)
		return hd
	case pb.HuodongTypeEnum_HSpringSkin:
		hd := springskin.NewSpringSkinHdByAttr(area, attr)
		return hd
	default:
		glog.Errorf("newHuodongByAttr what the fuck id=%d")
		return nil
	}
}

func (m *huodongModule) initAreaHuodong(area int) {
	attrs, err := attribute.LoadAll(fmt.Sprintf("huodong%d", area))
	if err != nil {
		panic(err)
	}

	type2huodong, ok := m.allHuodong[area]
	if !ok {
		type2huodong = map[pb.HuodongTypeEnum]htypes.IHuodong{}
		m.allHuodong[area] = type2huodong
	}

	for _, attr := range attrs {
		hd := m.newHuodongByAttr(area, attr)
		if hd != nil {
			type2huodong[hd.GetHtype()] = hd
		}
	}

	if module.Service.GetAppID() == 1 {
		var eventArgs []*pb.HuodongEvent
		for id, _ := range pb.HuodongTypeEnum_name {
			htype := pb.HuodongTypeEnum(id)
			if htype == pb.HuodongTypeEnum_HUnknow {
				continue
			}

			var needRefresh bool
			hd, ok := type2huodong[htype]
			if ok {
				needRefresh = hd.Refresh(m.getGameData(area, htype))
			} else {
				hd = m.NewHuodong(area, htype)
				if hd == nil {
					glog.Errorf("newHuodong error, htype=%d, area=%d", htype, area)
					continue
				}
				hd.Save()
				type2huodong[htype] = hd
				needRefresh = true
			}

			if needRefresh {
				eventArgs = append(eventArgs, &pb.HuodongEvent{
					HdType: htype,
					Event:  pb.HuodongEventType_HetRefresh,
					Areas:  []int32{int32(area)},
				})
			}
		}

		if len(eventArgs) > 0 {
			timer.AfterFunc(2 * time.Second, func() {
				for _, arg := range eventArgs {
					logic.BroadcastBackend(pb.MessageID_G2G_ON_HUODONG_EVENT, arg)
				}
			})
		}
	}
}

func (m *huodongModule) initAllAreaHuodong(gdata gamedata.IGameData) {
	areaGameData := gdata.(*gamedata.AreaConfigGameData)
	for _, areaCfg := range areaGameData.Areas {
		area := areaCfg.Area
		if _, ok := m.allHuodong[area]; ok {
			continue
		}

		m.initAreaHuodong(area)
	}
}

func (m *huodongModule) initializeHuodong() {
	m.allHuodong = map[int]map[pb.HuodongTypeEnum]htypes.IHuodong{}
	m.refreshingHuodong = map[int]map[pb.HuodongTypeEnum]chan struct{}{}

	areaGameData := gamedata.GetGameData(consts.AreaConfig)
	m.initAllAreaHuodong(areaGameData)

	if module.Service.GetAppID() == 1 {
		areaGameData.AddReloadCallback(m.initAllAreaHuodong)
		timer.AddTicker(10 * time.Second, onHeartBeat)
	}
}

func (m *huodongModule) GetHuodong(area int, htype pb.HuodongTypeEnum) htypes.IHuodong {
	if type2Huodong, ok := m.allHuodong[area]; !ok {
		return nil
	} else {
		if hd, ok := type2Huodong[htype]; ok {
			return hd
		} else {
			return nil
		}
	}
}

func (m *huodongModule) GetSeasonPvpHandCardInfo(player types.IPlayer) (gamedata.ISeasonPvp, int, pb.BattleHandType,
	*pb.SeasonPvpChooseCardData, *pb.FetchSeasonHandCardReply) {

	return seasonpvp.GetSeasonPvpHandCardInfo(player)
}

func (m *huodongModule) SeasonPvpChooseCamp(player types.IPlayer, camp int) *pb.SeasonPvpChooseCardData {
	return seasonpvp.SeasonPvpChooseCamp(player, camp)
}

func (m *huodongModule) SeasonPvpChooseCard(player types.IPlayer, cards []uint32) (randCards []uint32, err error) {
	return seasonpvp.SeasonPvpChooseCard(player, cards)
}

func (m *huodongModule) RefreshSeasonPvpChooseCard(player types.IPlayer) (*pb.SeasonPvpChooseCardData, error) {
	return seasonpvp.RefreshSeasonPvpChooseCard(player)
}

func (m *huodongModule) GetSeasonPvpWinCnt(player types.IPlayer) int {
	return seasonpvp.GetSeasonPvpWinCnt(player)
}

func (m *huodongModule) OnRecharge(player types.IPlayer, oldJade, money int) int {
	now := time.Now()
	if now.Before(htypes.ChristmasBegin) || !now.Before(htypes.ChristmasEnd) {
		return 0
	}

	cpt := player.GetComponent(consts.HuodongCpt).(*huodongComponent)
	return cpt.onChristmasRecharge(oldJade, money)
}

func (m *huodongModule) PackEventHuodongs(player types.IPlayer) []*pb.HuodongData {
	area := player.GetArea()
	var datas []*pb.HuodongData
	type2huodong, ok := m.allHuodong[area]
	if !ok {
		return datas
	}

	for _, hd := range type2huodong {
		if hd.IsOpen() {
			eventHd, ok := hd.(htypes.IEventHuodong)
			if ok {
				datas = append(datas, eventHd.PackEventMsg())
			}
		}
	}
	return datas
}

func (m *huodongModule) GetEventHuodongItemType(player types.IPlayer) int {
	return event.GetHuodongItemType(player, pb.HuodongTypeEnum_HSpringExchange)
}

func (m *huodongModule) GetTreasureHuodongSkin(player types.IPlayer, treasureData *gamedata.Treasure) (skins []string, eventItem int) {
	hd := m.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSpringSkin)
	if hd == nil || !hd.IsOpen() {
		return
	}
	return springskin.TreasureRandomSkin(player, treasureData)
}

func (m *huodongModule) refreshHuodong(area int, htype pb.HuodongTypeEnum) {
	c := make(chan struct{})
	type2refreshing, ok := m.refreshingHuodong[area]
	if !ok {
		type2refreshing = map[pb.HuodongTypeEnum]chan struct{} {}
		m.refreshingHuodong[area] = type2refreshing
	}
	type2refreshing[htype] = c

	defer func() {
		delete(type2refreshing, htype)
		close(c)
	}()

	attr := attribute.NewAttrMgr(fmt.Sprintf("huodong%d", area), htype, true)
	err := attr.Load()
	if err != nil {
		glog.Errorf("handlerHuodongEvent refresh load error, htype=%d, area=%d, err=%s", htype, area, err)
		return
	}

	hd := mod.newHuodongByAttr(area, attr)
	if hd != nil {
		glog.Infof("handlerHuodongEvent refresh area=%d, htype=%d", area, htype)
		type2huodong, ok := mod.allHuodong[area]
		if !ok {
			type2huodong = map[pb.HuodongTypeEnum]htypes.IHuodong{}
			mod.allHuodong[area] = type2huodong
		}
		type2huodong[htype] = hd
	}
}

func (m *huodongModule) handlerHuodongEvent(arg *pb.HuodongEvent) {
	for _, area2 := range arg.Areas {
		area := int(area2)
		type2refreshing, ok := mod.refreshingHuodong[area]
		if ok {
			refreshing, ok := type2refreshing[arg.HdType]
			if ok {
				evq.Await(func() {
					<-refreshing
				})
			}
		}

		m.refreshHuodong(area, arg.HdType)

		switch arg.Event {
		case pb.HuodongEventType_HetRefresh:

		case pb.HuodongEventType_HetStart:
			hd := mod.GetHuodong(area, arg.HdType)
			if hd == nil {
				glog.Errorf("handlerHuodongEvent start no huodong, htype=%d, area=%d", arg.HdType, area)
				return
			}
			hd.OnStart()

		case pb.HuodongEventType_HetStop:
			hd := mod.GetHuodong(area, arg.HdType)
			if hd == nil {
				glog.Errorf("handlerHuodongEvent stop no huodong, htype=%d, area=%d", arg.HdType, area)
				return
			}
			hd.OnStop()

		default:
			glog.Errorf("handlerHuodongEvent unknow eventType, htype=%d, event=%s, area=%d", arg.HdType,
				arg.Event, area)
		}

	}
}

func onHeartBeat() {
	now := time.Now()
	htype2Event := map[pb.HuodongTypeEnum]map[int]chan pb.HuodongEventType{}
	for area, type2huodong := range mod.allHuodong {
		for htype, hd := range type2huodong {
			c := hd.HeartBeat(now)
			if c == nil {
				continue
			}

			area2Event, ok := htype2Event[htype]
			if !ok {
				area2Event = map[int]chan pb.HuodongEventType{}
				htype2Event[htype] = area2Event
			}

			if c != nil {
				area2Event[area] = c
			}
		}
	}

	for htype, area2Event := range htype2Event {

		var arg *pb.HuodongEvent
		for area, c := range area2Event {
			var eventType pb.HuodongEventType
			evq.Await(func() {
				eventType = <-c
			})

			if eventType == pb.HuodongEventType_HetUnknow {
				continue
			}

			if arg == nil {
				arg = &pb.HuodongEvent{
					HdType: htype,
					Event: eventType,
				}
			}
			arg.Areas = append(arg.Areas, int32(area))
		}

		if arg != nil {
			logic.BroadcastBackend(pb.MessageID_G2G_ON_HUODONG_EVENT, arg)
		}
	}
}

func Initialize() {
	mod = &huodongModule{}
	mod.initializeHuodong()
	module.Huodong = mod
	htypes.Mod = mod
	seasonpvp.InitializeSeasonPvpHd()
	event.Initialize()
	spring.Initialize()
	springskin.Initialize()
	registerRpc()
}
