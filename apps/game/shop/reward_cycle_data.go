package shop

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"time"
)

var (
	crd *cycleRewardData
	cfg = &gamedata.RecruitRefreshConfig{}
)

const (
	cpt_             = "rewardCycleDataAttr"
	curCardVer_      = "cur_card_version"
	curSkinVer_      = "cur_skin_version"
	exchangeCardVer_ = "max_exchange_card_version"
	exchangeSkinVer_ = "max_exchange_skin_version"
	switchVer_       = "switch_ver"
	remainWeekNum_   = "remain_week_num"
)

type cycleReward struct {
	attr *attribute.AttrMgr //存进数据库的数据
	area int                //区号
}

type cycleRewardData struct {
	area2CycleReward map[int]*cycleReward
}

func (c *cycleRewardData) initAttr() {
	areaData, ok := gamedata.GetGameData(consts.AreaConfig).(*gamedata.AreaConfigGameData)
	if !ok {
		return
	}
	areaData.ForEachOpenedArea(func(config *gamedata.AreaConfig) {
		cycData, ok := c.area2CycleReward[config.Area]
		if !ok {
			cycData = &cycleReward{}
			c.area2CycleReward[config.Area] = cycData
		}
		id := fmt.Sprintf("%s_%d", cpt_, config.Area)
		cycData.attr = attribute.NewAttrMgr(cpt_, id)
		cycData.attr.Load()
		cycData.area = config.Area
	})
}

func (c *cycleReward) getCurCardVer() int {
	if vs := c.attr.GetInt(curCardVer_); vs == 0 {
		return 1
	} else {
		return vs
	}
}

func (c *cycleReward) getCurSkinVer() int {
	if vs := c.attr.GetInt(curSkinVer_); vs == 0 {
		return 1
	} else {
		return vs
	}
}

func (c *cycleReward) getMaxExchangeCardVer() int {
	if vs := c.attr.GetInt(exchangeCardVer_); vs == 0 {
		if c.getRewardType() == "card" {
			return 1
		} else {
			return 0
		}
	} else {
		return vs
	}
}

func (c *cycleReward) getMaxExchangeSkinVer() int {
	if vs := c.attr.GetInt(exchangeSkinVer_); vs == 0 {
		if c.getRewardType() == "skin" {
			return 1
		} else {
			return 0
		}
	} else {
		return vs
	}
}

func (c *cycleReward) getCurWeeks() int {
	return c.attr.GetInt(remainWeekNum_)
}

func (c *cycleReward) setCurWeeks(num int) {
	c.attr.SetInt(remainWeekNum_, num)
}

func (c *cycleReward) setCurCardVer(vs int) {
	c.attr.SetInt(curCardVer_, vs)
}

func (c *cycleReward) setCurSkinVer(vs int) {
	c.attr.SetInt(curSkinVer_, vs)
}

func (c *cycleReward) setMaxExchangeCardVer(ver int) {
	c.attr.SetInt(exchangeCardVer_, ver)
}

func (c *cycleReward) setMaxExchangeSkinVer(ver int) {
	c.attr.SetInt(exchangeSkinVer_, ver)
}

func (c *cycleReward) getSwitchVer() int {
	if ver := c.attr.GetInt(switchVer_); ver <= 0 {
		return 1
	} else {
		return ver
	}
}

func (c *cycleReward) setSwitchVer(ver int) {
	c.attr.SetInt(switchVer_, ver)
}

func (c *cycleReward) save() {
	c.attr.Save(false)
}

func (c *cycleReward) getIDs(tblName string) (pb.RecruitTreasureData_RewardType, []int32) {
	var rwType pb.RecruitTreasureData_RewardType
	gdata, ok := gamedata.GetGameData(tblName).(*gamedata.RewardTblGameData)
	if !ok {
		return rwType, nil
	}
	maxVer := gdata.Area2MaxVer[c.area]
	var ids []int32
	var ver int
	if gdata.Type == "card" {
		ver = c.getCurCardVer()
		if ver > maxVer {
			ver = 1
			c.setCurCardVer(ver)
			c.save()
		}
		rwType = pb.RecruitTreasureData_Card
	}
	if gdata.Type == "skin" {
		ver = c.getCurSkinVer()
		if ver > maxVer {
			ver = 1
			c.setCurSkinVer(ver)
			c.save()
		}
		rwType = pb.RecruitTreasureData_Skin
	}

	for _, gd := range gdata.Team2Rewards {
		for _, rw := range gd.Rewards {
			if rw.Version == ver && rw.AreaLimit.IsEffective(c.area) {
				ids = append(ids, int32(rw.ID))
			}
		}
	}
	return rwType, ids
}

func (c *cycleReward) getMaxCardVer() int {
	if gdata, ok := gamedata.GetGameData(consts.RecruitTreausreCardRewardTbl).(*gamedata.RewardTblGameData); ok {
		if mVer, ok := gdata.Area2MaxVer[c.area]; ok {
			return mVer
		}
	}
	return 0
}

func (c *cycleReward) getMaxSkinVer() int {
	if gdata, ok := gamedata.GetGameData(consts.RecruitTreausreSkinRewardTbl).(*gamedata.RewardTblGameData); ok {
		if mVer, ok := gdata.Area2MaxVer[c.area]; ok {
			return mVer
		}
	}
	return 0
}

func (cd *cycleRewardData) getRecruitIDs(area int, tblName string) (pb.RecruitTreasureData_RewardType, []int32, int64) {
	if cycData, ok := cd.area2CycleReward[area]; ok {
		wd := (cfg.Weeks - cycData.getCurWeeks() - 1) * 7
		td := time.Now().Weekday() + 1
		cfgDay := cfg.Day
		if cfgDay == 0 {
			cfgDay = 7
		}
		var dayNum int
		if int(td) > cfgDay {
			dayNum = 7 - int(td) + cfgDay + wd
		} else {
			dayNum = cfgDay - int(td) + wd
		}
		tim := int64(timer.TimeDelta(cfg.Hours, cfg.Min, cfg.Sec).Seconds() + float64(dayNum*86400))
		rwTy, ids := cycData.getIDs(tblName)
		return rwTy, ids, tim
	}
	return pb.RecruitTreasureData_Unknow, nil, 0
}

func (cd *cycleRewardData) getCurVerByArea(area int, ty pb.RecruitTreasureData_RewardType) int {
	if cycData, ok := cd.area2CycleReward[area]; ok {
		if ty == pb.RecruitTreasureData_Card {
			return cycData.getCurCardVer()
		}
		if ty == pb.RecruitTreasureData_Skin {
			return cycData.getCurSkinVer()
		}
	}
	return 0
}

func (cd *cycleRewardData) getRecruiteSwitchVer(area int) int {
	if cycData, ok := cd.area2CycleReward[area]; ok {
		return cycData.getSwitchVer()
	}
	return 1
}

func (c *cycleReward) addCardVer() {
	cvs := c.getCurCardVer()
	mcvs := c.getMaxExchangeCardVer()
	if cvs >= mcvs {
		c.setMaxExchangeCardVer(cvs)
	}
	cvs += 1
	if cvs > c.getMaxCardVer() {
		cvs = 1
	}
	c.setCurCardVer(cvs)
}

func (c *cycleReward) addSkinVer() {
	svs := c.getCurSkinVer()
	msvs := c.getMaxExchangeSkinVer()
	if svs >= msvs {
		c.setMaxExchangeSkinVer(svs)
	}
	svs += 1
	if svs > c.getMaxSkinVer() {
		svs = 1
	}
	c.setCurSkinVer(svs)
}

func (c *cycleReward) getRewardType() string {
	var tbName string
	gdata, ok := gamedata.GetGameData(consts.RecruitTreausre).(*gamedata.RecruitTreasureGameData)
	if !ok {
		glog.Infof("refreshVer get RecruitTreausre max switch error!")
		return tbName
	}

	swVer := c.getSwitchVer()
	team2Treasure := gdata.GetTeam2Treausre(c.area)
	for _, treasures := range team2Treasure {
		if treasure, ok := treasures[swVer]; ok {
			if treData, ok := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).Treasures[treasure.TreasureID]; ok {
				tbName = treData.RewardTbl
				if tbName == consts.RecruitTreausreCardRewardTbl {
					return "card"
				} else if tbName == consts.RecruitTreausreSkinRewardTbl {
					return "skin"
				} else {
					return tbName
				}
			}

		}
	}
	return tbName
}

func (cd *cycleRewardData) rfVer() {
	cd.refreshVer(cfg.Day)
}

func (cd *cycleRewardData) refreshVer(day int) {
	if int(time.Now().Weekday()) != day {
		return
	}
	gdata, ok := gamedata.GetGameData(consts.RecruitTreausre).(*gamedata.RecruitTreasureGameData)
	if !ok {
		glog.Infof("refreshVer get RecruitTreausre max switch error!")
		return
	}

	for area, c := range cd.area2CycleReward {
		wks := c.getCurWeeks() + 1
		if wks < cfg.Weeks {
			c.setCurWeeks(wks)
			c.save()
			continue
		}
		c.setCurWeeks(0)

		swVer := c.getSwitchVer()
		maxSwitch := gdata.GetMaxSwitchByArea(area)
		if c.getRewardType() == "card" {
			c.addCardVer()
		}
		if c.getRewardType() == "skin" {
			c.addSkinVer()
		}
		swVer += 1
		if swVer > maxSwitch {
			swVer = 1
		}
		c.setSwitchVer(swVer)
		c.save()

		cv := c.getCurCardVer()
		sv := c.getCurSkinVer()
		ssv := c.getSwitchVer()
		c.pushRecruitIdsToClientOnceArea()
		glog.Infof("Area=%d, After refreshVer switchVer=%d, cardVer=%d, skinVer=%d", area, ssv, cv, sv)
	}
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		p := player.GetComponent(consts.ShopCpt).(*shopComponent)
		p.recruitTreasure.syncToClient()
	})
}

func (c *cycleReward) getExchangeCardIds() []int32 {
	gdata, ok := gamedata.GetGameData(consts.RecruitTreausreCardRewardTbl).(*gamedata.RewardTblGameData)
	if !ok {
		return nil
	}
	maxVer := gdata.Area2MaxVer[c.area]
	var ids []int32
	var ver int
	ver = c.getMaxExchangeCardVer()
	if ver > maxVer {
		ver = maxVer
		c.setMaxExchangeCardVer(ver)
		c.save()
	}

	for _, gd := range gdata.Team2Rewards {
		for _, rw := range gd.Rewards {
			if rw.Version <= ver && rw.AreaLimit.IsEffective(c.area) {
				ids = append(ids, int32(rw.ID))
			}
		}
	}
	return ids
}

func (c *cycleReward) getExchangeSkinIds() []int32 {
	gdata, ok := gamedata.GetGameData(consts.RecruitTreausreSkinRewardTbl).(*gamedata.RewardTblGameData)
	if !ok {
		return nil
	}
	maxVer := gdata.Area2MaxVer[c.area]
	var ids []int32
	var ver int
	ver = c.getMaxExchangeSkinVer()
	if ver > maxVer {
		ver = maxVer
		c.setMaxExchangeSkinVer(ver)
		c.save()
	}

	for _, gd := range gdata.Team2Rewards {
		for _, rw := range gd.Rewards {
			if rw.Version <= ver && rw.AreaLimit.IsEffective(c.area) {
				ids = append(ids, int32(rw.ID))
			}
		}
	}
	return ids
}

func (cd *cycleRewardData) getExchangeIdsByArea(area int) ([]int32, []int32) {
	var cids []int32
	var sids []int32
	if cycData, ok := cd.area2CycleReward[area]; ok {
		cids = cycData.getExchangeCardIds()
		sids = cycData.getExchangeSkinIds()
	}
	return cids, sids
}

func (cd *cycleRewardData) pushRecruitIdsToClient(p types.IPlayer) {
	msg := &pb.PieceExchangeIds{}
	area := p.GetArea()
	cids, sids := cd.getExchangeIdsByArea(area)
	msg.ExchangeCardIds = cids
	msg.ExchangeSkinIds = sids
	p.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS, msg)
}

func (c *cycleReward) pushRecruitIdsToClientOnceArea() {
	msg := &pb.PieceExchangeIds{}
	msg.ExchangeCardIds = c.getExchangeCardIds()
	msg.ExchangeSkinIds = c.getExchangeSkinIds()
	module.Player.ForEachOnlinePlayer(func(p types.IPlayer) {
		if c.area == p.GetArea() {
			p.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS, msg)
		}
	})
}

func (cd *cycleRewardData) crossWeek() {
	cd.refreshVer(int(time.Now().Weekday()))
}

func crdInitialized() {
	gamedata.GetGameData(consts.RecruitRefreshConfig).AddReloadCallback(func(data gamedata.IGameData) {
		cfg.InitCfg()
	})
	cfg.InitCfg()
	crd = &cycleRewardData{
		area2CycleReward: map[int]*cycleReward{},
	}
	crd.initAttr()
	timer.RunEveryDay(cfg.Hours, cfg.Min, cfg.Sec, crd.rfVer)
}
