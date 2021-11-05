package main

import (
	"kinger/gopuppy/attribute"
	"kinger/gamedata"
	"kinger/common/consts"
	"fmt"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"kinger/gopuppy/common"
	"sort"
	"kinger/gopuppy/common/eventhub"
	"math"
	"kinger/gopuppy/apps/center/api"
	"strconv"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/gopuppy/common/timer"
	"time"
	"math/rand"
)

var cityMgr = &cityMgrSt{}

type cityMgrSt struct {
	allCitys map[int]*city
}

func (cm *cityMgrSt) initialize() {
	cm.allCitys = map[int]*city{}
	attrs, err := attribute.LoadAll(campaignMgr.genAttrName("city"))
	if err != nil {
		panic(err)
	}

	cityGameData := gamedata.GetGameData(consts.City).(*gamedata.CityGameData)
	for _, attr := range attrs {
		cityID, ok := attr.GetAttrID().(int)
		if !ok {
			panic(fmt.Sprintf("wrong cityID %s", attr.GetAttrID()))
		}
		cityData := cityGameData.ID2City[cityID]
		if cityData == nil {
			continue
		}

		cty := newCityByAttr(cityID, attr)
		cty.checkVersion()
		cty.fixResource()
		cm.allCitys[cityID] = cty

		cry := cty.getCountry()
		if cry != nil {
			cry.addCity(cty)
		}
	}

	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	for cityID, data := range cityGameData.ID2City {
		if _, ok := cm.allCitys[cityID]; ok {
			continue
		}
		cty := newCity(cityID)
		cty.calcGlory(paramGameData, 0, 0, 0, 0)
		cty.modifyResource(resDefense, data.DefenseMax * paramGameData.InitialDefense)
		cm.allCitys[cityID] = cty
		cty.attr.Save(false)
	}

	eventhub.Subscribe(evWarBegin, cm.onWarBegin)
	eventhub.Subscribe(evWarEnd, cm.onWarEnd)
	eventhub.Subscribe(evUnified, cm.onWarEnd)
	eventhub.Subscribe(evCityChangeCountry, cm.onCityChangeCountry)
	timer.AddTicker(15 * time.Minute, func() {
		cm.calcGlory()
	})
	timer.AddTicker(time.Duration(rand.Intn(20) + 290) * time.Second, func() {
		cm.save(false)
	})
	timer.AddTicker(time.Minute, cm.checkCaptiveTimeout)
	timer.AddTicker(10 * time.Second, cm.calcPerContribution)
}

func (cm *cityMgrSt) checkCaptiveTimeout() {
	if warMgr.isPause() {
		return
	}
	for _, cty := range cm.allCitys {
		cty.checkCaptiveTimeout()
	}
}

func (cm *cityMgrSt) onCityChangeCountry(args ...interface{}) {
	cityID := args[0].(int)
	for _, cty := range cm.allCitys {
		if cty.getCityID() != cityID {
			cty.onCityChangeCountry(cityID)
		}
	}
}

func (cm *cityMgrSt) calcGlory() {
	if warMgr.isPause() {
		return
	}

	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	var allCountryPlayerAmount int
	var maxCountryPlayerAmount int
	country2PlayerAmount := map[uint32]int{}
	allCrys := countryMgr.getAllCountrys()
	for _, cry := range allCrys {
		amount := cry.getPlayerAmount()
		allCountryPlayerAmount += amount
		country2PlayerAmount[cry.getID()] = amount
		if amount > maxCountryPlayerAmount {
			maxCountryPlayerAmount = amount
		}
	}

	cityAmount := cm.getCityAmount()
	for _, cty := range cm.allCitys {
		cty.calcGlory(paramGameData, allCountryPlayerAmount, maxCountryPlayerAmount,
			country2PlayerAmount[cty.getCountryID()], cityAmount)
	}
}

func (cm *cityMgrSt) onWarBegin(args ...interface{}) {
	for _, cty := range cm.allCitys {
		cty.onWarBegin()
	}
}

func (cm *cityMgrSt) onWarEnd(args ...interface{})  {
	for _, cty := range cm.allCitys {
		cty.onWarEnd()
		campaignMgr.addSortCity(cty.getCityID())
		cty.playersNeedSort = true
	}
}

func (cm *cityMgrSt) save(isStopServer bool) {
	for _, cty := range cm.allCitys {
		cty.save(isStopServer)
	}
}

func (cm *cityMgrSt) sortPlayers() {
	for _, cty := range cm.allCitys {
		cty.sortInCityPlayers()
		cty.sortPlayers()
		cty.sortCaptives()
	}
}

func (cm *cityMgrSt) sortCaPlayers() {
	for _, cty := range cm.allCitys {
		cty.sortCaPlayers()
	}
}

func (cm *cityMgrSt) getCity(cityID int) *city {
	return cm.allCitys[cityID]
}

func (cm *cityMgrSt) citysCreateCountry() {
	for _, cty := range cm.allCitys {
		cty.createCountry()
	}
}

func (cm *cityMgrSt) forEachCity(callback func(cty *city) bool) {
	for _, cty := range cm.allCitys {
		if !callback(cty) {
			break
		}
	}
}

func (cm *cityMgrSt) getCityAmount() int {
	return len(cm.allCitys)
}

func (cm *cityMgrSt) calcPerContribution() {
	if warMgr.isPause() {
		return
	}
	for _, cty := range cm.allCitys {
		cty.calcPerContribution()
	}
}

type city struct {
	id int
	attr *attribute.AttrMgr
	resourceAttr *attribute.MapAttr
	missionsAttr *attribute.ListAttr
	militaryOrderAttr *attribute.ListAttr
	// 注资记录
	capitalInjectionAttr *attribute.ListAttr
	// 平时不用存盘，但停服时需要存的，用户停服热更
	memAttr *attribute.MapAttr

	// 属于这个城的玩家
	uid2Player map[common.UUid]*player
	players   []*player
	playersNeedSort bool
	// 在这个城的玩家，包括属于这个和不属于这个城
	uid2InCityPlayer map[common.UUid]*player
	inCityPlayers   []*player
	// 在这个城的俘虏
	uid2Captive map[common.UUid]*player
	captives   []*player
	// 在这个城申请创建国家
	ccApplys []*createCountryApply
	uid2CcApply map[common.UUid]*createCountryApply
	// 已发布的任务
	missions []*mission
	// 城市的官员
	job2Players map[pb.CampaignJob][]*player
	// 正在攻打这个城的队伍
	attackingTeams common.IntSet
	// 注资记录
	capitalInjections []*attribute.MapAttr
	uid2CapitalInjection map[common.UUid]*attribute.MapAttr
	// 军令
	militaryOrders []*militaryOrder
	// 这个城获得战功的玩家
	addContributionPlayers map[common.UUid]float64
	addContributions float64
}

func newCityByAttr(cityID int, attr *attribute.AttrMgr) *city {
	c := &city{
		id:                   cityID,
		attr:                 attr,
		capitalInjectionAttr: attr.GetListAttr("capitalInjection"),
		uid2Player:           map[common.UUid]*player{},
		uid2InCityPlayer: map[common.UUid]*player{},
		uid2Captive: map[common.UUid]*player{},
		uid2CcApply:          map[common.UUid]*createCountryApply{},
		job2Players:          map[pb.CampaignJob][]*player{},
		uid2CapitalInjection: map[common.UUid]*attribute.MapAttr{},
		addContributionPlayers: map[common.UUid]float64{},
	}

	// 资源
	c.resourceAttr = attr.GetMapAttr("res2")
	if c.resourceAttr == nil {
		c.resourceAttr = attribute.NewMapAttr()
		attr.SetMapAttr("res2", c.resourceAttr)

		oldResourceAttr := attr.GetMapAttr("res")
		if oldResourceAttr != nil {
			// 兼容老数据
			oldResourceAttr.ForEachKey(func(key string) {
				c.resourceAttr.SetFloat64(key, float64(oldResourceAttr.GetInt(key)))
			})
		}
	}

	/*
	c.resourceAttr = attr.GetMapAttr("res")
	if c.resourceAttr == nil {
		c.resourceAttr = attribute.NewMapAttr()
		attr.SetMapAttr("res", c.resourceAttr)
	}
	*/

	// 任务
	c.missionsAttr = attr.GetListAttr("missions")
	if c.missionsAttr == nil {
		c.missionsAttr = attribute.NewListAttr()
		attr.SetListAttr("missions", c.missionsAttr)
	}
	c.missionsAttr.ForEachIndex(func(index int) bool {
		mAttr := c.missionsAttr.GetMapAttr(index)
		c.missions = append(c.missions, newMissionByAttr(c, mAttr))
		return true
	})
	//c.sortMissions()

	if c.capitalInjectionAttr != nil {
		c.capitalInjectionAttr.ForEachIndex(func(index int) bool {
			cattr := c.capitalInjectionAttr.GetMapAttr(index)
			c.uid2CapitalInjection[common.UUid(cattr.GetUInt64("uid"))] = cattr
			c.capitalInjections = append(c.capitalInjections, cattr)
			return true
		})
		c.sortCapitalInjections()
	}

	c.militaryOrderAttr = attr.GetListAttr("militaryOrder")
	if c.militaryOrderAttr == nil {
		c.militaryOrderAttr = attribute.NewListAttr()
		attr.SetListAttr("militaryOrder", c.militaryOrderAttr)
	}
	c.militaryOrderAttr.ForEachIndex(func(index int) bool {
		mattr := c.militaryOrderAttr.GetMapAttr(index)
		c.militaryOrders = append(c.militaryOrders, newMilitaryOrderByAttr(c, mattr))
		return true
	})

	// 停服热更暂存数据
	memAttr := attr.GetMapAttr("__memAttr")
	c.memAttr = attribute.NewMapAttr()
	if memAttr != nil {
		attr.Del("__memAttr")
		mem := memAttr.ToMap()
		c.memAttr.AssignMap(mem)
	}

	addContributionPlayersAttr := c.attr.GetMapAttr("addContributionPlayers")
	if addContributionPlayersAttr != nil {
		addContributionPlayersAttr.ForEachKey(func(key string) {
			uid := common.ParseUUidFromString(key)
			amount := addContributionPlayersAttr.GetFloat64(key)
			c.addContributionPlayers[uid] = amount
			c.addContributions += amount
		})
		c.attr.Del("addContributionPlayers")
	}

	return c
}

func newCity(cityID int) *city {
	attr := attribute.NewAttrMgr(campaignMgr.genAttrName("city"), cityID)
	attr.SetInt("version", cdVersion)
	return newCityByAttr(cityID, attr)
}

func (c *city) checkCaptiveTimeout() {
	now := time.Now().Unix()
	for _, p := range c.uid2Captive {
		if now >= p.getCaptiveTimeout() {
			p.setCaptive(nil, true)
			if p.getCountry() != nil {
				p.setLastCountryID(0, true)
				p.quitCountry(0)
			} else {
				p.subContribution(p.getMaxContribution() * 0.1, true)
			}
		}
	}
}

func (c *city) checkVersion() {
	if c.attr.GetInt("version") == cdVersion {
		return
	}

	c.attr.SetInt("version", cdVersion)
	defense := c.getResource(resDefense) * 10
	c.setResource(resDefense, defense)
}

func (c *city) fixResource() {
	if c.getResource(resAgriculture) < 0 {
		c.setResource(resAgriculture, 0)
	}
	if c.getResource(resBusiness) < 0 {
		c.setResource(resBusiness, 0)
	}
}

func (c *city) sortCapitalInjections() {
	sort.Slice(c.capitalInjections, func(i, j int) bool {
		cattr1 := c.capitalInjections[i]
		cattr2 := c.capitalInjections[j]
		gold1 := cattr1.GetInt("gold")
		gold2 := cattr2.GetInt("gold")
		if gold1 > gold2 {
			return true
		} else if gold2 > gold1 {
			return false
		}

		return cattr1.GetInt64("time") >= cattr2.GetInt64("time")
	})
}

func (c *city) onTeamEnter(t *team) {
	p := t.getOwner()
	cityID := p.getCityID()
	if cityID == 0 {
		cityID = c.id
	}
	p.setCity(cityID, c.id, true)
	c.modifyResource(resForage, float64(t.getForage()))

	if t.fighterData == nil {
		return
	}
	if cityID == c.id {
		return
	}
	supportCards := p.getSupportCards()
	if len(supportCards) > 0 {
		return
	}

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, card := range t.fighterData.HandCards {
		cardData := poolGameData.GetCardByGid(card.GCardID)
		if cardData != nil {
			supportCards = append(supportCards, cardData.CardID)
		}
	}
	p.updateSupportCards(supportCards)
}

func (c *city) save(isStopServer bool) {
	if isStopServer {
		if c.memAttr != nil && !c.attr.HasKey("__memAttr") {
			c.attr.SetMapAttr("__memAttr", c.memAttr)
		}

		addContributionPlayersAttr := attribute.NewMapAttr()
		for uid, amount := range c.addContributionPlayers {
			addContributionPlayersAttr.SetFloat64(uid.String(), amount)
		}
		c.attr.SetMapAttr("addContributionPlayers", addContributionPlayersAttr)
	}
	c.attr.Save(isStopServer)
}

func (c *city) isBeOccupy() bool {
	return c.memAttr != nil && c.memAttr.GetBool("isBeOccupy")
}

func (c *city) addCcApply(ccApply *createCountryApply) {
	uid := ccApply.getUid()
	if _, ok := c.uid2CcApply[uid]; !ok {
		c.uid2CcApply[uid] = ccApply
		c.ccApplys = append(c.ccApplys, ccApply)
	}
}

func (c *city) delPlayer(p *player) {
	uid := p.getUid()
	if _, ok := c.uid2Player[uid]; ok {
		delete(c.uid2Player, uid)
		for i, p2 := range c.players {
			if p2.getUid() == uid {
				c.players = append(c.players[:i], c.players[i+1:]...)
				break
			}
		}

		if len(c.uid2Player) <= 0 {
			c.cancelAllMission()
		}

		job := p.getCityJob()
		if job != pb.CampaignJob_UnknowJob {
			ps := c.job2Players[job]
			for i, p3 := range ps {
				if uid == p3.getUid() {
					c.job2Players[job] = append(ps[:i], ps[i+1:]...)
					return
				}
			}
		}
	}
}

func (c *city) addPlayer(p *player) {
	uid := p.getUid()
	if _, ok := c.uid2Player[uid]; !ok {
		c.uid2Player[uid] = p
		c.players = append(c.players, p)
		c.playersNeedSort = true

		job := p.getCityJob()
		if job != pb.CampaignJob_UnknowJob {
			ps := c.job2Players[job]
			for _, p2 := range ps {
				if p.getUid() == p2.getUid() {
					return
				}
			}

			ps = append(ps, p)
			c.job2Players[job] = sortPlayers(ps)
		}
	}
}

func (c *city) delInCityPlayer(p *player) {
	uid := p.getUid()
	if _, ok := c.uid2InCityPlayer[uid]; ok {
		delete(c.uid2InCityPlayer, uid)
		if p.isOnline() {
			p.agent.SetClientFilter("campaign_lcity", "")
		}

		for i, p2 := range c.inCityPlayers {
			if p2.getUid() == uid {
				c.inCityPlayers = append(c.inCityPlayers[:i], c.inCityPlayers[i+1:]...)
				break
			}
		}

		c.broadcastInCityPlayer(pb.MessageID_S2C_SYNC_CITY_PLYAER_AMOUNT, &pb.SyncCityPlayerAmount{
			CityID: int32(c.id),
			Amount: int32(len(c.inCityPlayers)),
		})
	}
}

func (c *city) delCaptive(p *player) {
	uid := p.getUid()
	if _, ok := c.uid2Captive[uid]; ok {
		delete(c.uid2Captive, uid)
		for i, p2 := range c.captives {
			if uid == p2.getUid() {
				c.captives = append(c.captives[:i], c.captives[i+1:]...)
				return
			}
		}
	}
}

func (c *city) addInCityPlayer(p *player) {
	uid := p.getUid()
	if p.isCaptive() {
		if _, ok := c.uid2Captive[uid]; !ok {
			c.uid2Captive[uid] = p
			c.captives = append(c.captives, p)
			if p.getCaptiveTimeout() <= 0 {
				p.setCaptiveTimeout(time.Now().Unix() + captiveTimeout)
			}
		}
		return
	}

	if _, ok := c.uid2InCityPlayer[uid]; !ok {
		c.uid2InCityPlayer[uid] = p
		c.inCityPlayers = append(c.inCityPlayers, p)
		campaignMgr.addSortCity(c.id)
		if p.isOnline() {
			p.agent.SetClientFilter("campaign_lcity", strconv.Itoa(c.id))
		}

		c.broadcastInCityPlayer(pb.MessageID_S2C_SYNC_CITY_PLYAER_AMOUNT, &pb.SyncCityPlayerAmount{
			CityID: int32(c.id),
			Amount: int32(len(c.inCityPlayers)),
		})
	}
}

func (c *city) getCityID() int {
	return c.id
}

func (c *city) getGameData() *gamedata.City {
	return gamedata.GetGameData(consts.City).(*gamedata.CityGameData).ID2City[c.id]
}

func (c *city) getInCityPlayerAmount() int {
	return len(c.inCityPlayers)
}

func (c *city) getPlayerAmount() int {
	return len(c.players)
}

func (c *city) getResource(resType string) float64 {
	return c.resourceAttr.GetFloat64(resType)
}

func (c *city) setResource(resType string, val float64) {
	if resType == resDefense && c.getResource(resDefense) != val {
		warMgr.addBeAttackCity(c.id)
	}
	c.resourceAttr.SetFloat64(resType, val)
}

func (c *city) modifyResource(resType string, val float64) {
	if val == 0 {
		return
	}
	old := c.getResource(resType)
	cur := old + val
	if cur < 0 && resType != resAgriculture && resType != resBusiness {
		cur = 0
	}

	if val > 0 && resType != resForage && resType != resGold {
		var max float64
		data := c.getGameData()
		switch resType {
		case resAgriculture:
			max = data.AgricultureMax
		case resBusiness:
			max = data.BusinessMax
		case resDefense:
			max = data.DefenseMax
		}

		if max > 0 && cur > max {
			cur = max
		}
	}

	c.setResource(resType, cur)
	glog.Infof("city modify resource, cityID=%d, resType=%s, val=%f, old=%f, new=%f", c.id, resType, val, old, cur)
}

func (c *city) calcGlory(paramGameData *gamedata.CampaignParamGameData, allCountryPlayerAmount, maxCountryPlayerAmount,
	myCountryPlayerAmount, cityAmount int) {

	var glory float64
	data := c.getGameData()
	playerAmount := c.getPlayerAmount()
	if cityAmount == 0 || playerAmount == 0 || myCountryPlayerAmount == 0 {
		glory = 10.0
	} else {
		// 基础荣耀值＊所有势力玩家数／本城市玩家数／总城池数＊（（最大势力人数／本势力人数－1）＊honor_revise＋1）
		glory = ( data.GloryBase * float64(allCountryPlayerAmount) / float64(playerAmount) / float64(cityAmount) * (
			(float64(maxCountryPlayerAmount)/float64(myCountryPlayerAmount)-1) * paramGameData.HonorRevise + 1) ) * 10
		glory = math.Round(glory)
		glory /= 10
	}

	if glory < 0.1 {
		glory = 0.1
	} else if glory > 10.0 {
		glory = 10.0
	}
	c.attr.SetFloat64("glory", glory)
}

func (c *city) getGlory() float64 {
	return c.attr.GetFloat64("glory")
}

func (c *city) getCountry() *country {
	countryID := c.getCountryID()
	if countryID <= 0 {
		return nil
	}
	return countryMgr.getCountry(countryID)
}

func (c *city) getCountryID() uint32 {
	return c.attr.GetUInt32("country")
}

func (c *city) setCountryID(countryID uint32) {
	oldCryID := c.getCountryID()
	if oldCryID == countryID {
		return
	}
	oldCry := countryMgr.getCountry(oldCryID)
	if oldCry != nil {
		oldCry.delCity(c)
	}
	newCry := countryMgr.getCountry(countryID)
	if newCry != nil {
		newCry.addCity(c)
	}
	c.attr.SetUInt32("country", countryID)

	glog.Infof("city setCountryID, cityID=%d, oldCryID=%d, countryID=%d", c.id, oldCryID, countryID)

	c.cancelAllMission()
	c.cancelAllMilitaryOrder()
	eventhub.Publish(evCityChangeCountry, c.id, oldCryID, countryID)

	campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CITY_COUNTRY, &pb.UpdateCityCountryArg{
		CityID: int32(c.id),
		CountryID: countryID,
	})
}

func (c *city) onCityChangeCountry(cityID int) {
	for i := 0; i < len(c.militaryOrders); {
		mo := c.militaryOrders[i]
		if mo.getTargetCity() == cityID {
			mo.cancel()
			c.militaryOrders = append(c.militaryOrders[:i], c.militaryOrders[i+1:]...)
			c.militaryOrderAttr.DelByIndex(i)
		} else {
			i++
		}
	}

	for i := 0; i < len(c.missions); {
		ms := c.missions[i]
		if ms.getTransportCity() == cityID && ms.getType() == pb.CampaignMsType_Dispatch {
			ms.cancel()
			c.missions = append(c.missions[:i], c.missions[i+1:]...)
			c.militaryOrderAttr.DelByIndex(i)
		} else {
			i++
		}
	}
}

func (c *city) autocephaly(isFix bool) error {

	jobPs := c.job2Players[pb.CampaignJob_Prefect]
	if len(jobPs) <= 0 {
		return gamedata.GameError(10)
	}

	p := jobPs[0]
	oldCry := c.getCountry()
	oldCryID := oldCry.getID()

	// TODO
	flag := ""
	cry, err := newCountry(createCountryMgr.genCountryName(p.getName()), "", c)
	if err != nil {
		glog.Errorf("autocephaly error, cityID=%d, err=%d", c.id, err)
		return err
	}

	countryID := cry.getID()
	campaignMgr.broadcastClient(pb.MessageID_S2C_COUNTRY_CREATED, &pb.CountryCreatedArg{
		CountryID: countryID,
		Name: cry.getName(),
		Flag: flag,
		CityID: int32(c.id),
		YourMajesty: p.packSimpleMsg(),
	})

	c.setCountryID(countryID)
	countryMgr.addCountry(cry)
	for _, p2 := range c.uid2Player {
		p2.setCountryID(countryID, true)
		if p != p2 {
			p2.setLastCountryID(oldCryID, true)
			noticeMgr.sendNoticeToPlayer(p2.getUid(), pb.CampaignNoticeType_AutocephalyNt3, p.getName(), c.id, cry.getName())
		}
	}

	if !isFix {
		p.subContribution( p.getMaxContribution() * 0.2, true )
	}
	p.setJob(pb.CampaignJob_Prefect, pb.CampaignJob_YourMajesty, true)
	noticeMgr.sendNoticeToCountry(oldCryID, pb.CampaignNoticeType_AutocephalyNt2, c.id, pb.CampaignJob_Prefect, p.getName(),
		oldCry.getName(), cry.getName())
	return nil
}

func (c *city) playerSettle(p *player, needSync bool) error {
	uid := p.getUid()
	if _, ok := c.uid2Player[uid]; ok {
		return gamedata.GameError(2)
	}
	if p.getCity() != nil {
		return gamedata.GameError(4)
	}

	p.setCity(c.id, c.id, needSync)
	cry := c.getCountry()
	var countryID uint32
	if cry != nil {
		countryID = cry.getID()
		p.setCountryID(countryID, needSync)
		cry.addPlayer(p)
	}
	glog.Infof("city playerSettle, uid=%d, cityID=%d, countryID=%d", p.getUid(), c.id, countryID)
	return nil
}

func (c *city) sortCaPlayers() {
	sort.Slice(c.ccApplys, func(i, j int) bool {
		return c.ccApplys[i].battleThan(c.ccApplys[j])
	})
}

func (c *city) setForagePrice(price int) {
	c.attr.SetInt("foragePrice", price)
}

func (c *city) getForagePrice() int {
	return c.attr.GetInt("foragePrice")
}

func (c *city) getAllPlayers() []*player  {
	return c.players
}

func (c *city) getCaptives() []*player {
	return c.captives
}

func (c *city) getAllInCityPlayers() []*player  {
	var ps []*player
	for _, p := range c.inCityPlayers {
		if _, ok := c.uid2Player[p.getUid()]; !ok {
			ps = append(ps, p)
		}
	}
	return ps
}

func (c *city) onJobUpdate(p *player, oldJob, newJob pb.CampaignJob) {
	if oldJob != pb.CampaignJob_UnknowJob {
		ps := c.job2Players[oldJob]
		for i, p2 := range ps {
			if p2.getUid() == p.getUid() {
				c.job2Players[oldJob] = append(ps[:i], ps[i+1:]...)
				break
			}
		}
	}

	if newJob != pb.CampaignJob_UnknowJob {
		ps := c.job2Players[newJob]
		c.job2Players[newJob] = sortPlayers(append(ps, p))
	}
	c.playersNeedSort = true
}

func (c *city) delCcApply(ccApply *createCountryApply) {
	uid := ccApply.getUid()
	if ca, ok := c.uid2CcApply[uid]; ok {
		delete(c.uid2CcApply, uid)
		for i, ca2 := range c.ccApplys {
			if ca == ca2 {
				c.ccApplys = append(c.ccApplys[:i], c.ccApplys[i+1:]...)
				break
			}
		}
	}
}

func (c *city) applyCreateCountry(p *player, gold int) {
	uid := p.getUid()
	if ca, ok := c.uid2CcApply[uid]; ok {
		ca.addGold(gold)
	} else {
		ca := newCreateCountryApply(p, gold)
		c.uid2CcApply[uid] = ca
		c.ccApplys = append(c.ccApplys, ca)
		p.addCcApply(ca)
	}
	c.sortCaPlayers()
}

func (c *city) getCreateCountryApply(uid common.UUid) *createCountryApply {
	return c.uid2CcApply[uid]
}

func (c *city) getCreateCoutryApplysByPage(page int) []*pb.ApplyCreateCountryPlayer {
	var ret []*pb.ApplyCreateCountryPlayer
	beginIdx := page * pageAmount
	endIdx := beginIdx + pageAmount
	totalAmount := len(c.ccApplys)
	if beginIdx >= totalAmount {
		return ret
	}
	if endIdx > totalAmount {
		endIdx = totalAmount
	}

	for i := beginIdx; i < endIdx; i++ {
		ret = append(ret, c.ccApplys[i].packMsg())
	}
	return ret
}

func (c *city) getTopCreateCoutryApply() *createCountryApply {
	if len(c.ccApplys) > 0 {
		return c.ccApplys[0]
	}
	return nil
}

func (c *city) createCountry() {
	if len(c.ccApplys) <= 0 {
		return
	}
	if c.getCountry() != nil {
		return
	}

	ca := c.ccApplys[0]
	p, _ := playerMgr.loadPlayer(ca.getUid())
	if p == nil {
		glog.Errorf("createCountry no player %s, cityID=%d", ca.getUid(), c.id)
		return
	}

	if p.getCityID() != c.id {
		glog.Errorf("createCountry error uid=%d, cityID=%d, playerCityID=%d", ca.getUid(), c.id, p.getCityID())
		return
	}

	// TODO
	flag := ""
	cry, err := newCountry(ca.getCountryName(), "", c)
	if err != nil {
		glog.Errorf("createCountry error, cityID=%d, err=%d", c.id, err)
		return
	}

	countryID := cry.getID()
	campaignMgr.broadcastClient(pb.MessageID_S2C_COUNTRY_CREATED, &pb.CountryCreatedArg{
		CountryID: countryID,
		Name: ca.getCountryName(),
		Flag: flag,
		CityID: int32(c.id),
		YourMajesty: p.packSimpleMsg(),
	})

	c.setCountryID(countryID)
	c.ccApplys = []*createCountryApply{}
	c.uid2CcApply = map[common.UUid]*createCountryApply{}
	countryMgr.addCountry(cry)
	c.modifyResource(resGold, float64(ca.getGold()) / 2)

	for _, p2 := range c.players {
		p2.delCcApply(p.getUid() != p2.getUid())
		cry.addPlayer(p2)
		p2.setCountryID(countryID, true)
	}

	p.setJob(pb.CampaignJob_Prefect, pb.CampaignJob_YourMajesty, true)
	noticeMgr.sendNoticeToCity(c.id, pb.CampaignNoticeType_NewCountryNt, p.getName(), ca.getCountryName(), c.id)
}

func (c *city) packMsg(uid common.UUid) *pb.CityData {
	msg := &pb.CityData{
		CountryID: c.getCountryID(),
		PlayerAmount: int32(c.getPlayerAmount()),
		Agriculture: int32(c.getResource(resAgriculture)),
		Business: int32(c.getResource(resBusiness)),
		Defense: int32(c.getResource(resDefense)),
		Forage: int32(c.getResource(resForage)),
		Gold: int32(c.getResource(resGold)),
		Glory: int32(c.getGlory() * 10),
		InCityPlayerAmount: int32(c.getInCityPlayerAmount()),
	}

	allCityJobs := campaignMgr.getCityJobs()
	for _, job := range allCityJobs {
		ps := c.job2Players[job]
		for _, p := range ps {
			msg.Players = append(msg.Players, p.packMsg(false))
		}
	}

	cry := c.getCountry()
	if cry == nil {
		msg.ApplyCreateCountry = &pb.ApplyCreateCountryData{
			RemainTime: int32(createCountryMgr.getApplyCreateCountryRemainTime().Seconds()),
			Players: c.getCreateCoutryApplysByPage(0),
		}
		ca := c.getCreateCountryApply(uid)
		if ca != nil {
			msg.ApplyCreateCountry.MyApplyMoney = int32(ca.getGold())
		}
	} else {

		yourMajesty := cry.getYourMajesty()
		if yourMajesty != nil {
			msg.YourMajesty = yourMajesty.packSimpleMsg()
		}
	}

	return msg
}

func (c *city) packSimpleMsg() *pb.CitySimpleData {
	msg := &pb.CitySimpleData{
		CityID: int32(c.id),
		CountryID: c.getCountryID(),
		Defense: int32(c.getResource(resDefense)),
		State: pb.CityState_NormalCS,
	}

	if warMgr.isInWar() {
		if c.isBeOccupy() {
			msg.State = pb.CityState_BeOccupyCS
		} else if c.isBeAttack() {
			msg.State = pb.CityState_BeAttackCS
		}
	}
	return msg
}

func (c *city) onMissionCancel(m *playerMission) {
	for _, ms := range c.missions {
		if ms.equal(m) {
			if ms.getAmount() + 1 <= ms.getMaxAmount() {
				ms.setAmount(ms.getAmount() + 1)
				return
			} else {
				break
			}
		}
	}
	c.modifyResource(resGold, float64(m.getGoldReward() + m.getTransportGold()))
	c.modifyResource(resForage, float64(m.getTransportForage()))
}

func (c *city) getAllMissions() []*mission {
	return c.missions
}

func (c *city) acceptMission(p *player, msType pb.CampaignMsType, transportTargetCity int, gcardIDs []uint32) *playerMission {
	for _, ms := range c.missions {
		if ms.getType() == msType && ms.getTransportCity() == transportTargetCity {
			if ms.getAmount() > 0 {
				return ms.accept(p, gcardIDs)
			}
		}
	}
	return nil
}

func (c *city) publishMission(uid common.UUid, msType pb.CampaignMsType, gold, amount int,
	transportType pb.TransportTypeEnum, transportCityPath []int) error {

	var transportForage int
	var transportGold int
	var maxTime int
	var targetCity int
	var dispatchGold int
	var distance int

	if msType == pb.CampaignMsType_Transport || msType == pb.CampaignMsType_Dispatch {
		if len(transportCityPath) < 2 {
			return gamedata.GameError(4)
		}
		if transportCityPath[0] != c.id {
			return gamedata.GameError(5)
		}

		roadGameData := gamedata.GetGameData(consts.Road).(*gamedata.RoadGameData)
		maxIndex := len(transportCityPath) - 1
		targetCity = transportCityPath[maxIndex]
		if targetCity == c.id {
			return gamedata.GameError(5)
		}

		if msType == pb.CampaignMsType_Dispatch {
			targetCty := cityMgr.getCity(targetCity)
			if targetCty == nil || targetCty.getCountryID() != c.getCountryID() {
				return gamedata.GameError(5)
			}
		}

		for i := 0; i < maxIndex; i++ {
			city1 := transportCityPath[i]
			city2 := transportCityPath[i+1]
			rs, ok := roadGameData.City2Road[city1]
			if !ok {
				return gamedata.GameError(5)
			}
			r, ok := rs[city2]
			if !ok {
				return gamedata.GameError(5)
			}
			distance += r.Distance
		}

		paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
		if msType == pb.CampaignMsType_Dispatch {

			dispatchGold = int( paramGameData.TransferCost * float64(distance) )
		} else {
			maxTime = int(float64(distance) * paramGameData.TransportTime)
			if transportType == pb.TransportTypeEnum_ForageTT {
				transportForage = paramGameData.TaskTransportForage
			} else {
				transportGold = paramGameData.TaskTransportGold
			}
		}
	}

	for i, m := range c.missions {
		if m.getType() == msType && m.getTransportCity() == targetCity {
			m.cancel()
			c.missions = append(c.missions[:i], c.missions[i+1:]...)
			c.missionsAttr.DelByIndex(i)
			break
		}
	}

	needGold := float64(amount * (gold + transportGold + dispatchGold))
	needForage := float64(amount * transportForage)
	if c.getResource(resGold) < needGold {
		return gamedata.GameError(4)
	}
	if c.getResource(resForage) < needForage {
		return gamedata.GameError(4)
	}

	c.modifyResource(resGold, - needGold)
	c.modifyResource(resForage, - needForage)
	ms := newMission(c, msType, amount, gold)
	ms.setMaxTime(maxTime)
	ms.setTransportCity(targetCity)
	ms.setTransportForage(transportForage)
	ms.setTransportGold(transportGold)
	ms.setTransportDistance(distance)
	c.missionsAttr.AppendMapAttr(ms.attr)
	c.missions = append(c.missions, ms)
	//c.sortMissions()

	glog.Infof("publishMission cityID=%d, uid=%d, msType=%s, gold=%d, transportGold=%d, transportForage=%d, " +
		"amount=%d, targetCity=%d", c.id, uid, msType, gold, transportGold, transportForage, amount, targetCity)
	return nil
}

func (c *city) cancelMission(type_ pb.CampaignMsType, transportTargetCity int) {
	for i, m := range c.missions {
		if m.getType() == type_ && m.getTransportCity() == transportTargetCity {
			m.cancel()
			c.missions = append(c.missions[:i], c.missions[i+1:]...)
			c.missionsAttr.DelByIndex(i)
			break
		}
	}
}

func (c *city) cancelAllMission() {
	for _, m := range c.missions {
		m.cancel()
	}
	c.missions = []*mission{}
	c.missionsAttr = attribute.NewListAttr()
	c.attr.SetListAttr("missions", c.missionsAttr)
}

func (c *city) sortMissions() {
	sort.Slice(c.missions, func(i, j int) bool {
		return c.missions[i].getPublishTime() >= c.missions[j].getPublishTime()
	})
}

func (c *city) getJobPlayers(job pb.CampaignJob) []*player {
	return c.job2Players[job]
}

func (c *city) getAllJobPlayers() map[pb.CampaignJob][]*player {
	return c.job2Players
}

func (c *city) sortInCityPlayers() {
	c.inCityPlayers = sortPlayers(c.inCityPlayers)
}

func (c *city) sortPlayers() {
	if c.playersNeedSort {
		c.playersNeedSort = false
		c.players = sortPlayers(c.players)
	}
}

func (c *city) sortCaptives() {
	c.captives = sortPlayers(c.captives)
}

func (c *city) recallJob(p, targetPlayer *player, job pb.CampaignJob, isActive, needSync, needNotice bool) error {
	targetJob := targetPlayer.getCityJob()
	myJob := p.getCityJob()
	myCountryJob := p.getCountryJob()
	if targetJob != job {
		return gamedata.GameError(20)
	}

	if myCountryJob == pb.CampaignJob_YourMajesty {

		if p.getCountryID() != targetPlayer.getCountryID() {
			return gamedata.GameError(11)
		}

	} else {

		if job != pb.CampaignJob_Prefect {
			if p.getCityID() != targetPlayer.getCityID() {
				return gamedata.GameError(11)
			}
			if myJob != pb.CampaignJob_Prefect && !isActive {
				return gamedata.GameError(12)
			}
		} else {
			if p.getCountryID() != targetPlayer.getCountryID() {
				return gamedata.GameError(11)
			}
			if myCountryJob != pb.CampaignJob_YourMajesty && !isActive {
				return gamedata.GameError(12)
			}
		}
	}

	jobPlayers := c.job2Players[targetJob]
	for i, jp := range jobPlayers {
		if jp.getUid() == targetPlayer.getUid() {
			c.job2Players[targetJob] = append(jobPlayers[:i], jobPlayers[i+1:]...)
			break
		}
	}
	targetPlayer.setJob(pb.CampaignJob_UnknowJob, targetPlayer.getCountryJob(), needSync)
	if needNotice && !isActive {
		noticeMgr.sendNoticeToPlayer(targetPlayer.getUid(), pb.CampaignNoticeType_RecallJobNt, job, c.id)

	} else if isActive {

		if job != pb.CampaignJob_Prefect {
			prefect := c.getPrefect()
			if prefect != nil {
				noticeMgr.sendNoticeToPlayer(prefect.getUid(), pb.CampaignNoticeType_ResignNt, job, targetPlayer.getName(),
					c.id)
			}
		} else {

			cry := c.getCountry()
			if cry != nil {
				yourMajesty := cry.getYourMajesty()
				if yourMajesty != nil {
					noticeMgr.sendNoticeToPlayer(yourMajesty.getUid(), pb.CampaignNoticeType_ResignNt, job,
						targetPlayer.getName(), c.id)
				}
			}
		}
	}

	glog.Infof("city recallJob cityID=%d, uid=%d, targetUid=%d, myJob=%s, targetJob=%s", c.id, p.getUid(),
		targetPlayer.getUid(), myJob, targetJob)

	return nil
}

// 玩家p，把玩家oldUid的职位job，任命给玩家targetPlayer
func (c *city) appointJob(p, targetPlayer *player, job pb.CampaignJob, oldUid common.UUid) error {

	targetJob := targetPlayer.getCityJob()
	myJob := p.getCountryJob()
	myCityJob := p.getCityJob()

	if job == pb.CampaignJob_Prefect {
		if p.getCountryID() != targetPlayer.getCountryID() {
			return gamedata.GameError(11)
		}
		if myJob != pb.CampaignJob_YourMajesty {
			return gamedata.GameError(12)
		}
	} else if myCityJob == pb.CampaignJob_YourMajesty {

		if p.getCountryID() != targetPlayer.getCountryID() {
			return gamedata.GameError(11)
		}

	} else {
		if p.getCityID() != targetPlayer.getCityID() {
			return gamedata.GameError(11)
		}
		if myCityJob != pb.CampaignJob_Prefect {
			return gamedata.GameError(12)
		}
	}

	var jobMaxAmount int
	switch job {
	case pb.CampaignJob_Prefect:
		jobMaxAmount = maxPrefectAmount
	case pb.CampaignJob_DuWei:
		jobMaxAmount = maxDuWeiAmount
	case pb.CampaignJob_FieldOfficer:
		jobMaxAmount = maxFieldOfficerAmount
	default:
		return gamedata.GameError(13)
	}

	jobPlayers := c.job2Players[job]
	var oldPlayer *player
	if oldUid > 0 {
		// 罢免oldPlayer
		oldPlayer, _ = playerMgr.loadPlayer(oldUid)
		if oldPlayer == nil || oldPlayer.getCityJob() != job {
			return gamedata.GameError(14)
		}

		err := c.recallJob(p, oldPlayer, job, false,true, true)
		if err != nil {
			return err
		}
	} else if len(jobPlayers) >= jobMaxAmount {
		return gamedata.GameError(15)
	}

	if targetJob != pb.CampaignJob_UnknowJob {
		// 罢免targetPlayer原来的职位
		err := c.recallJob(p, targetPlayer, targetJob, false,false, false)
		if err != nil {
			return err
		}
	}

	// 任命targetPlayer
	targetPlayer.setJob(job, targetPlayer.getCountryJob(), true)
	noticeMgr.sendNoticeToCity(targetPlayer.getCityID(), pb.CampaignNoticeType_AppointJobNt, p.getName(),
		targetPlayer.getName(), job, c.id)

	glog.Infof("city appointJob cityID=%d, uid=%d, targetUid=%d, oldUid=%d, myJob=%s, targetOldJob=%s, targetJob=%s",
		c.id, p.getUid(), targetPlayer.getUid(), oldUid, myJob, targetJob, job)
	return nil
}

func (c *city) isBeAttack() bool {
	return c.attackingTeams != nil && c.attackingTeams.Size() > 0
}

func (c *city) beOccupy(t *team) {
	cry := t.owner.getCountry()
	if cry == nil {
		return
	}

	cityCry := c.getCountry()
	if cityCry != nil {
		cityCry.onLostCity(c)
	}
	cityCryID := c.getCountryID()
	cry.onOccupyCity(c)

	c.attackingTeams.Remove(t.getID())
	c.setCountryID(cry.getID())

	glog.Infof("city beOccupy, tid=%d, uid=%d, oldCountryID=%d, countryID=%d", t.getID(), t.owner.getUid(),
		cityCryID, cry.getID())

	var captiveAmount int
	for _, p := range c.uid2InCityPlayer {
		if p.getTeam() == nil || p.getTeam().isDefTeam() {
			p.setCaptive(c, true)
			captiveAmount++
		}
	}

	cityCry = countryMgr.getCountry(cityCryID)
	if cityCry != nil {
		noticeMgr.sendNoticeToCountry(cityCryID, pb.CampaignNoticeType_BeOccupyNt, cry.getName(), c.id, captiveAmount)
	}
	noticeMgr.sendNoticeToCountry(cry.getID(), pb.CampaignNoticeType_OccupyNt, c.id, captiveAmount)

	for _, p := range c.uid2Player {
		p.setJob(pb.CampaignJob_UnknowJob, p.getCountryJob(), true)
		lCityID := p.getLocationCityID()
		cityID := p.getCityID()

		if lCityID != c.id {
			// 属于这个城，但现在不在这个城
			cityID = lCityID
			lCty := cityMgr.getCity(lCityID)
			if lCty == nil || lCty.getCountryID() != cityCryID {
				cityID = 0
				if cityCry != nil {
					newCty := cityCry.randomCity()
					if newCty != nil {
						cityID = newCty.getCityID()
					}
				}
			}
			p.setCity(cityID, lCityID, true)

		} else if cityID == c.id {
			// 属于这个城，现在在这个城
			cityID = 0
			if cityCry != nil {
				newCty := cityCry.randomCity()
				if newCty != nil {
					cityID = newCty.getCityID()
				}
			}
			p.setCity(cityID, lCityID, true)
		}
	}

	c.uid2Player = map[common.UUid]*player{}
	c.players = []*player{}
	c.uid2InCityPlayer = map[common.UUid]*player{}
	c.inCityPlayers = []*player{}
	c.job2Players = map[pb.CampaignJob][]*player{}

	if c.memAttr == nil {
		c.memAttr = attribute.NewMapAttr()
	}
	c.memAttr.SetBool("isBeOccupy", true)
	c.attackingTeams = nil

	for _, p := range c.uid2Captive {
		if p.getCountryID() == cry.getID() {
			p.setCaptive(nil, true)
			if p.getCity() == nil {
				p.setCity(c.id, c.id, true)
			}

			if p.getLocationCityID() == c.id {
				c.addInCityPlayer(p)
			}
		}
	}

	c.captives = sortPlayers(c.captives)

	campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CITY_STATE, &pb.UpdateCityStateArg{
		State: pb.CityState_BeOccupyCS,
		CityID: int32(c.id),
		OccupyCountryID: cry.getID(),
	})
}

func (c *city) broadcastInCityPlayer(msgID pb.MessageID, arg interface{})  {
	api.BroadcastClient(msgID, arg, &gpb.BroadcastClientFilter{
		OP: gpb.BroadcastClientFilter_EQ,
		Key: "campaign_lcity",
		Val: strconv.Itoa(c.id),
	})
}

func (c *city) delAttackTeam(t *team) {
	tid := t.getID()
	if c.attackingTeams == nil || !c.attackingTeams.Contains(tid) {
		return
	}
	c.attackingTeams.Remove(tid)
	if c.attackingTeams.Size() == 0 && !c.isBeOccupy() {
		campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CITY_STATE, &pb.UpdateCityStateArg{
			State: pb.CityState_NormalCS,
			CityID: int32(c.id),
		})
	}
}

func (c *city) addAttackTeam(t *team, needSync bool) {
	if c.attackingTeams == nil {
		c.attackingTeams = common.IntSet{}
	}

	if c.attackingTeams.Size() == 0 && needSync {
		campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CITY_STATE, &pb.UpdateCityStateArg{
			State: pb.CityState_BeAttackCS,
			CityID: int32(c.id),
		})
	}

	c.attackingTeams.Add(t.getID())
}

func (c *city) attack(t *team, defTeamAmount int) bool {
	c.addAttackTeam(t, true)
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	atkTeamAmount := float64(c.attackingTeams.Size())
	damage := paramGameData.SingleDamage * (1 - paramGameData.DefenseRevise *
		(1 - atkTeamAmount / (atkTeamAmount + float64(defTeamAmount))))
	c.modifyResource(resDefense, - damage)
	//c.modifyResource(resBusiness, - paramGameData.SingleDamage)
	//c.modifyResource(resAgriculture, - paramGameData.SingleDamage)
	if c.getResource(resDefense) <= 0 {
		c.beOccupy(t)
		return true
	} else {
		return false
	}
}

func (c *city) getCurBusinessGold() (float64, float64) {
	business := c.getResource(resBusiness)
	if business <= 0 {
		return 0, business
	}
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	return business * paramGameData.GoldConversion, business
}

func (c *city) onProduction() float64 {
	gold, business := c.getCurBusinessGold()
	c.modifyResource(resGold, gold)
	if business > 0 {
		c.modifyResource(resBusiness, - business)
	}

	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	agriculture := c.getResource(resAgriculture)
	var forage float64
	if agriculture > 0 {
		forage = agriculture * paramGameData.ForageConversion
		c.modifyResource(resForage, forage)
		c.modifyResource(resAgriculture, - agriculture)
	}

	for _, ps := range c.job2Players {
		for _, p := range ps {
			noticeMgr.sendNoticeToPlayer(p.getUid(), pb.CampaignNoticeType_ProductionNt, int(gold), int(forage))
		}
	}
	return gold
}

func (c *city) getPrefect() *player {
	ps := c.job2Players[pb.CampaignJob_Prefect]
	if len(ps) <= 0 {
		return nil
	}
	return ps[0]
}

func (c *city) capitalInjection(p *player, gold int) {
	c.modifyResource(resGold, float64(gold))
	noticeMgr.sendNoticeToCity(c.id, pb.CampaignNoticeType_CapitalInjectionNt, p.getCityJob(), p.getName(), gold)
	if c.capitalInjectionAttr == nil {
		c.capitalInjectionAttr = attribute.NewListAttr()
		c.attr.SetListAttr("capitalInjection", c.capitalInjectionAttr)
	}

	attr, ok := c.uid2CapitalInjection[p.getUid()]
	if !ok {
		attr = attribute.NewMapAttr()
		attr.SetUInt64("uid", uint64(p.getUid()))
		c.capitalInjectionAttr.AppendMapAttr(attr)
		c.uid2CapitalInjection[p.getUid()] = attr
		c.capitalInjections = append(c.capitalInjections, attr)
	}

	attr.SetInt64("time", time.Now().Unix())
	attr.SetInt("gold", attr.GetInt("gold") + gold)
	c.sortCapitalInjections()
}

func (c *city) getCapitalInjections(page int) []*pb.CapitalInjectionRecord {
	var ret []*pb.CapitalInjectionRecord
	beginIdx := page * pageAmount
	endIdx := beginIdx + pageAmount
	totalAmount := len(c.capitalInjections)
	if beginIdx >= totalAmount {
		return ret
	}
	if endIdx > totalAmount {
		endIdx = totalAmount
	}

	for i := beginIdx; i < endIdx; i++ {
		attr := c.capitalInjections[i]
		p := playerMgr.getPlayer(common.UUid(attr.GetUInt64("uid")))
		if p == nil {
			continue
		}

		ret = append(ret, &pb.CapitalInjectionRecord{
			Player: p.packMsg(false),
			Gold: int32(attr.GetInt("gold")),
			Time: int32(attr.GetInt64("time")),
		})
	}
	return ret
}

func (c *city) setNotice(notice string) {
	c.attr.SetStr("notice", notice)
}

func (c *city) getNotice() string {
	return c.attr.GetStr("notice")
}

func (c *city) cancelMilitaryOrder(moType pb.MilitaryOrderType, targetCity int) {
	for i, mo := range c.militaryOrders {
		if mo.getType() == moType && mo.getTargetCity() == targetCity {
			mo.cancel()
			c.militaryOrders = append(c.militaryOrders[:i], c.militaryOrders[i+1:]...)
			c.militaryOrderAttr.DelByIndex(i)
			return
		}
	}
}

func (c *city) cancelAllMilitaryOrder() {
	for _, mo := range c.militaryOrders {
		mo.cancel()
	}
	c.militaryOrders = []*militaryOrder{}
	c.militaryOrderAttr = attribute.NewListAttr()
	c.attr.SetListAttr("militaryOrder", c.militaryOrderAttr)
}

func (c *city) getMilitaryOrdersInfo() *pb.MilitaryOrderInfo {
	msg := &pb.MilitaryOrderInfo{}
	for _, mo := range c.militaryOrders {
		msg.Orders = append(msg.Orders, mo.packMsg())
	}
	return msg
}

func (c *city) publishMilitaryOrder(moType pb.MilitaryOrderType, forage, amount int, cityPath []int) (
	*pb.PublishMilitaryOrderReply, error) {

	totalForage := float64(forage * amount)
	if c.getResource(resForage) < totalForage {
		return nil, gamedata.GameError(4)
	}

	var targetCityID int
	switch moType {
	case pb.MilitaryOrderType_SupportMT:
		forage = 0
		pathLen := len(cityPath)
		myCountryID := c.getCountryID()
		if pathLen < 2 {
			return nil, gamedata.GameError(5)
		}
		targetCityID = cityPath[pathLen-1]
		if cityPath[0] != c.id || targetCityID == c.id {
			return nil, gamedata.GameError(6)
		}

		targetCity := cityMgr.getCity(targetCityID)
		if targetCity == nil || targetCity.getCountryID() != myCountryID {
			return nil, gamedata.GameError(11)
		}

		roadGameData := gamedata.GetGameData(consts.Road).(*gamedata.RoadGameData)
		for i, cityID := range cityPath {
			cty := cityMgr.getCity(cityID)
			if cty == nil {
				return nil, gamedata.GameError(7)
			}
			if myCountryID != cty.getCountryID() {
				return nil, gamedata.GameError(8)
			}

			if i <= pathLen - 2 {
				rs, ok := roadGameData.City2Road[cityID]
				if !ok {
					return nil, gamedata.GameError(9)
				}
				cityID2 := int(cityPath[i+1])
				_, ok = rs[cityID2]
				if !ok {
					return nil, gamedata.GameError(10)
				}
			} else {
				break
			}
		}

	case pb.MilitaryOrderType_ExpeditionMT:
		if len(cityPath) != 2 {
			return nil, gamedata.GameError(15)
		}
		targetCityID = cityPath[1]
		if cityPath[0] != c.id || targetCityID == c.id {
			return nil, gamedata.GameError(16)
		}
		targetCty := cityMgr.getCity(targetCityID)
		if targetCty == nil || targetCty.getCountryID() == c.getCountryID() {
			return nil, gamedata.GameError(17)
		}

	case pb.MilitaryOrderType_DefCityMT:
		cityPath = []int{}
	default:
		return nil, gamedata.GameError(5)
	}

	c.modifyResource(resForage, - totalForage)
	c.cancelMilitaryOrder(moType, targetCityID)
	mo := newMilitaryOrder(c, moType, forage, amount, cityPath)
	c.militaryOrders = append(c.militaryOrders, mo)
	c.militaryOrderAttr.AppendMapAttr(mo.attr)

	return &pb.PublishMilitaryOrderReply{
		Orders: c.getMilitaryOrdersInfo().Orders,
		Forage: int32(c.getResource(resForage)),
	}, nil
}

func (c *city) acceptMilitaryOrder(p *player, moType pb.MilitaryOrderType, targetCity int, cardIDs []uint32) (
	*pb.AcceptMilitaryOrderReply, error) {

	if p.getLocationCityID() != c.id {
		return nil, gamedata.GameError(4)
	}

	var mo *militaryOrder
	for _, mo2 := range c.militaryOrders {
		if mo2.getType() == moType && mo2.getTargetCity() == targetCity {
			mo = mo2
			break
		}
	}

	if mo == nil || mo.getAmount() <= 0 {
		return nil, gamedata.GameError(5)
	}

	var needFighterData bool
	var needPath bool
	var teamType int
	switch moType {
	case pb.MilitaryOrderType_SupportMT:
		needPath = true
		teamType = ttSupport
	case pb.MilitaryOrderType_ExpeditionMT:
		if c.isBeAttack() {
			return nil, gamedata.GameError(11)
		}
		needPath = true
		needFighterData = true
		teamType = ttExpedition
	case pb.MilitaryOrderType_DefCityMT:
		needFighterData = true
	default:
		return nil, gamedata.GameError(6)
	}

	if p.getLocationCityID() != p.getCityID() {

		supportCards := p.getSupportCards()
		for _, cardID := range supportCards {
			isExist := false
			for _, cardID2 := range cardIDs {
				if cardID == cardID2 {
					isExist = true
					break
				}
			}

			if !isExist {
				return nil, gamedata.GameError(7)
			}
		}

	} else if moType == pb.MilitaryOrderType_SupportMT {
		p.updateSupportCards(cardIDs)
	}

	var distance int
	var paths []*roadPath
	var fighterData *pb.FighterData
	cityPath := mo.getCityPath()

	if needPath {

		pathLen := len(cityPath)
		roadGameData := gamedata.GetGameData(consts.Road).(*gamedata.RoadGameData)
		for i, cityID := range cityPath {
			cty := cityMgr.getCity(cityID)
			if cty == nil {
				return nil, gamedata.GameError(8)
			}

			if i <= pathLen-2 {
				rs, ok := roadGameData.City2Road[cityID]
				if !ok {
					return nil, gamedata.GameError(10)
				}
				cityID2 := int(cityPath[i+1])
				r, ok := rs[cityID2]
				if !ok {
					return nil, gamedata.GameError(11)
				}

				distance += r.Distance
				paths = append(paths, newRoadPath(distance, cityID2, r))
			} else {
				break
			}
		}

	}


	if needFighterData {
		reply, err := p.agent.CallBackend(pb.MessageID_L2G_GET_PVP_FIGHTER_DATA, &pb.GetFighterDataArg{
			CardIDs: cardIDs,
		})
		if err != nil {
			return nil, err
		}

		if mo.getAmount() <= 0 || mo.isCancel() {
			return nil, gamedata.GameError(12)
		}

		fighterData = reply.(*pb.FighterData)
	}

	moAmount := mo.getAmount() - 1
	mo.setAmount(moAmount)
	if moAmount <= 0 {
		for i, mo2 := range c.militaryOrders {
			if mo2 == mo {
				c.militaryOrders = append(c.militaryOrders[:i], c.militaryOrders[i+1:]...)
				c.militaryOrderAttr.DelByIndex(i)
				break
			}
		}
	}

	var t *team
	if moType == pb.MilitaryOrderType_DefCityMT {
		t = newDefTeam(p, fighterData, c.id, 0)
		cityMatchMgr.beginMatch(c.id, t)
	} else {
		t = newTeam(p, teamType, 0, cityPath, paths, fighterData)
	}
	p.setTeam(t)
	t.modifyForage(mo.getForage())

	glog.Infof("acceptMilitaryOrder uid=%d, t=%s", p.getUid(), t)

	return &pb.AcceptMilitaryOrderReply{
		State: p.getMyState(),
		Team: t.packMsg(),
	}, nil
}

func (c *city) onWarEnd() {
	if c.memAttr != nil {
		c.memAttr = nil
		c.attr.Del("__memAttr")
	}

	c.cancelAllMilitaryOrder()
	c.attackingTeams = nil
}

func (c *city) onPlayerAddContribution(p *player, val float64) {
	old := c.addContributionPlayers[p.getUid()]
	c.addContributionPlayers[p.getUid()] = old + val
	c.addContributions += val
}

func (c *city) calcPerContribution() {
	if c.addContributions <= 0 {
		return
	}

	for _, ps := range c.job2Players {
		for _, p := range ps {
			p.addContribution( (c.addContributions - c.addContributionPlayers[p.getUid()]) *
				campaignMgr.getContributionPerByJob(p.getCityJob()), false, false )
		}
	}

	c.addContributions = 0
	c.addContributionPlayers = map[common.UUid]float64{}
}

func (c *city) getMaxMissionReward() int {
	var val int
	for _, ms := range c.missions {
		if ms.getAmount() <= 0 {
			continue
		}
		msType := ms.getType()
		if msType == pb.CampaignMsType_Dispatch || msType == pb.CampaignMsType_Transport {
			continue
		}

		gold := ms.getGoldReward()
		if gold > val {
			val = gold
		}
	}
	return val
}

func (c *city) onWarBegin() {
	if len(c.militaryOrders) > 0 {
		return
	}
	cry := c.getCountry()
	if cry == nil {
		return
	}

	amount := c.getInCityPlayerAmount()
	if amount <= 0 {
		return
	}

	c.publishMilitaryOrder(pb.MilitaryOrderType_DefCityMT, 0, amount * 5, []int{})
	job2Players := cry.getAllJobPlayers()
	for _, ps := range job2Players {
		for _, p := range ps {
			if p.getLocationCityID() == c.id {
				continue
			}
			noticeMgr.sendNoticeToPlayer(p.getUid(), pb.CampaignNoticeType_AutoDefOrderNt, c.id)
		}
	}

	for _, p := range c.inCityPlayers {
		noticeMgr.sendNoticeToPlayer(p.getUid(), pb.CampaignNoticeType_AutoDefOrderNt, c.id)
	}
}

func (c *city) doSurrender(p2 *player, targetCry *country, noticeType pb.CampaignNoticeType) {
	countryID := targetCry.getID()
	var pName string
	if p2 != nil {
		pName = p2.getName()
	}
	for _, p := range c.uid2Player {
		oldCountryID := p.getCountryID()
		p.setJob(p.getCityJob(), pb.CampaignJob_UnknowJob, true)
		p.setCountryID(countryID, true)
		if p2 != p {
			p.setLastCountryID(oldCountryID, true)
			switch noticeType {
			case pb.CampaignNoticeType_SurrenderCity3Nt:
				noticeMgr.sendNoticeToPlayer(p.getUid(), noticeType, pName, c.id, targetCry.getName())
			case pb.CampaignNoticeType_SurrenderCountry1Nt:
				noticeMgr.sendNoticeToPlayer(p.getUid(), noticeType, pName, targetCry.getName())
			}
		}
	}

	c.setCountryID(countryID)

	for _, p := range c.uid2Captive {
		if p.getCountryID() == countryID {
			p.setCaptive(nil, true)
			cityID := p.getCityID()
			if cityID == 0 {
				cityID = c.id
			}
			p.setCity(cityID, cityID, true)
			lcty := p.getLocationCity()
			if lcty != nil {
				lcty.addInCityPlayer(p)
			}
		}
	}
}

func (c *city) surrender(p *player, targetCry *country, isFix bool) error {
	var pName string
	var pUid common.UUid
	if !isFix && p.getCityJob() != pb.CampaignJob_Prefect {
		return gamedata.GameError(10)
	}

	if p != nil {
		pName = p.getName()
		pUid = p.getUid()
	}

	cry := c.getCountry()
	if cry == nil {
		return gamedata.GameError(11)
	}

	yourMajesty := cry.getYourMajesty()
	if yourMajesty == nil || yourMajesty.getCityID() == c.id {
		return gamedata.GameError(12)
	}

	noticeMgr.sendNoticeToCountry(targetCry.getID(), pb.CampaignNoticeType_SurrenderCity2Nt, pName, c.id)
	c.doSurrender(p, targetCry, pb.CampaignNoticeType_SurrenderCity3Nt)
	if !isFix {
		p.subContribution(p.getMaxContribution() * 0.2, true)
	}
	noticeMgr.sendNoticeToCountry(cry.getID(), pb.CampaignNoticeType_SurrenderCity1Nt, pName, c.id, targetCry.getName())
	glog.Infof("city surrender, cityID=%d, uid=%d", c.id, pUid)
	return nil
}
