package main

import (
	"fmt"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"math/rand"
	"strconv"
	"time"
)

var countryMgr = &countryMgrSt{}

type countryMgrSt struct {
	allCountrys map[uint32]*country
}

func (cm *countryMgrSt) initialize() {
	cm.allCountrys = map[uint32]*country{}
	attrs, err := attribute.LoadAll(campaignMgr.genAttrName("country"))
	if err != nil {
		panic(err)
	}

	for _, attr := range attrs {
		id, ok := attr.GetAttrID().(int)
		if !ok {
			panic(fmt.Sprintf("wrong countryID %s", attr.GetAttrID()))
		}
		countryID := uint32(id)
		cry := newCountryByAttr(countryID, attr)
		if cry.isDestory() {
			continue
		}
		cm.allCountrys[countryID] = cry
	}

	timer.RunEveryDay(0, 0, 0, cm.onProduction)
	//timer.AddTicker(5 * time.Minute, cm.onProduction)
	timer.AddTicker(time.Duration(rand.Intn(20)+290)*time.Second, func() {
		cm.save(false)
	})
	timer.AddTicker(10*time.Second, cm.calcPerContribution)

	eventhub.Subscribe(evWarReady, cm.onWarReady)
	eventhub.Subscribe(evWarEnd, cm.onWarEnd)
	eventhub.Subscribe(evUnified, cm.onWarEnd)
}

func (cm *countryMgrSt) onWarEnd(args ...interface{}) {
	for _, cry := range cm.allCountrys {
		campaignMgr.addSortCountry(cry.getID())
		cry.playerNeedSort = true
	}
}

func (cm *countryMgrSt) onWarReady(args ...interface{}) {
	for _, cry := range cm.allCountrys {
		cry.onWarReady()
	}
}

func (cm *countryMgrSt) onProduction() {
	//if time.Now().Weekday() != time.Monday {
	//	return
	//}
	if warMgr.isPause() {
		return
	}
	for _, cry := range cm.allCountrys {
		cry.onProduction()
	}
}

func (cm *countryMgrSt) delCountry(cry *country) {
	delete(cm.allCountrys, cry.getID())
	if len(cm.allCountrys) != 1 {
		return
	}

	var countryID uint32
	var unifiedCry *country
	for cryID, cry := range cm.allCountrys {
		countryID = cryID
		unifiedCry = cry
	}

	isUnified := true
	cityMgr.forEachCity(func(cty *city) bool {
		if cty.getCountryID() != countryID {
			isUnified = false
			return false
		}
		return true
	})

	if isUnified {
		warMgr.onUnified(unifiedCry)
	}
}

func (cm *countryMgrSt) sortPlayers() {
	for _, cry := range cm.allCountrys {
		cry.sortPlayers()
	}
}

func (cm *countryMgrSt) addCountry(cry *country) {
	cm.allCountrys[cry.getID()] = cry
}

func (cm *countryMgrSt) getAllPlayerAmount() int {
	var amount int
	for _, cry := range cm.allCountrys {
		amount += cry.getPlayerAmount()
	}
	return amount
}

func (cm *countryMgrSt) getCountry(countryID uint32) *country {
	return cm.allCountrys[countryID]
}

func (cm *countryMgrSt) genCountryID() (uint32, error) {
	return common.Gen32UUid("countryID")
}

func (cm *countryMgrSt) getAllCountrys() map[uint32]*country {
	return cm.allCountrys
}

func (cm *countryMgrSt) save(isStopServer bool) {
	for _, cry := range cm.allCountrys {
		cry.save(isStopServer)
	}
}

func (cm *countryMgrSt) calcPerContribution() {
	if warMgr.isPause() {
		return
	}
	for _, cry := range cm.allCountrys {
		cry.calcPerContribution()
	}
}

type country struct {
	id            uint32
	attr          *attribute.AttrMgr
	warRecordAttr *attribute.MapAttr

	citys common.IntSet
	// 国家的玩家
	uid2Player map[common.UUid]*player
	players    []*player
	// 国家职位的玩家
	job2Players    map[pb.CampaignJob][]*player
	playerNeedSort bool
	// 这个城获得战功的玩家
	addContributionPlayers map[common.UUid]float64
	addContributions       float64
}

func newCountry(name, flag string, cty *city) (*country, error) {
	countryID, err := countryMgr.genCountryID()
	if err != nil {
		return nil, err
	}
	attr := attribute.NewAttrMgr(campaignMgr.genAttrName("country"), countryID)
	c := newCountryByAttr(countryID, attr)
	c.setName(name)
	c.setFlag(flag)
	c.addCity(cty)
	c.save(false)
	return c, nil
}

func newCountryByAttr(countryID uint32, attr *attribute.AttrMgr) *country {
	c := &country{
		id:                     countryID,
		attr:                   attr,
		citys:                  common.IntSet{},
		uid2Player:             map[common.UUid]*player{},
		job2Players:            map[pb.CampaignJob][]*player{},
		warRecordAttr:          attr.GetMapAttr("warRecord"),
		addContributionPlayers: map[common.UUid]float64{},
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

func (c *country) onLostCity(cty *city) {
	if c.warRecordAttr == nil {
		c.warRecordAttr = attribute.NewMapAttr()
		c.attr.SetMapAttr("warRecord", c.warRecordAttr)
	}

	lostCityAttr := c.warRecordAttr.GetListAttr("lostCity")
	if lostCityAttr == nil {
		lostCityAttr = attribute.NewListAttr()
		c.warRecordAttr.SetListAttr("lostCity", lostCityAttr)
	}
	lostCityAttr.AppendInt(cty.getCityID())
}

func (c *country) onOccupyCity(cty *city) {
	if c.warRecordAttr == nil {
		c.warRecordAttr = attribute.NewMapAttr()
		c.attr.SetMapAttr("warRecord", c.warRecordAttr)
	}

	lostCityAttr := c.warRecordAttr.GetListAttr("occupyCity")
	if lostCityAttr == nil {
		lostCityAttr = attribute.NewListAttr()
		c.warRecordAttr.SetListAttr("occupyCity", lostCityAttr)
	}
	lostCityAttr.AppendInt(cty.getCityID())
}

func (c *country) getLastWarRecord() *pb.CaStateWarEndArg {
	record := &pb.CaStateWarEndArg{}
	if c.warRecordAttr == nil {
		return record
	}

	lostCityAttr := c.warRecordAttr.GetListAttr("lostCity")
	if lostCityAttr != nil {
		lostCityAttr.ForEachIndex(func(index int) bool {
			record.LostCitys = append(record.LostCitys, int32(lostCityAttr.GetInt(index)))
			return true
		})
	}

	occupyCityAttr := c.warRecordAttr.GetListAttr("occupyCity")
	if occupyCityAttr != nil {
		occupyCityAttr.ForEachIndex(func(index int) bool {
			record.OccupyCitys = append(record.OccupyCitys, int32(occupyCityAttr.GetInt(index)))
			return true
		})
	}

	return record
}

func (c *country) onWarReady() {
	if c.warRecordAttr != nil {
		c.attr.Del("warRecord")
		c.warRecordAttr = nil
	}
}

func (c *country) isKingdom() bool {
	return c.attr.GetBool("kingdom")
}

func (c *country) getCityAmount() int {
	return c.citys.Size()
}

func (c *country) isDestory() bool {
	return c.attr.GetBool("isDestory")
}

func (c *country) setDestory() {
	c.attr.SetBool("isDestory", true)
	warMgr.onCountryDestory(c)
	countryMgr.delCountry(c)

	campaignMgr.broadcastClient(pb.MessageID_S2C_COUNTRY_DESTORYED, &pb.CountryDestoryed{
		CountryID: c.id,
	})

	for _, p := range c.uid2Player {
		p.onCountryDestory(c.id)
	}
}

func (c *country) setName(name string) {
	c.attr.SetStr("name", name)
	for _, p := range c.players {
		p.syncInfoToGame()
	}
}

func (c *country) setFlag(flag string) {
	c.attr.SetStr("flag", flag)
}

func (c *country) getAllPlayers() []*player {
	return c.players
}

func (c *country) delPlayer(p *player) {
	uid := p.getUid()
	if _, ok := c.uid2Player[uid]; ok {
		delete(c.uid2Player, uid)
		for i, p2 := range c.players {
			if uid == p2.getUid() {
				c.players = append(c.players[:i], c.players[i+1:]...)
				break
			}
		}

		job := p.getCountryJob()
		if job != pb.CampaignJob_UnknowJob {
			ps := c.job2Players[job]
			for i, p3 := range ps {
				if uid == p3.getUid() {
					c.job2Players[job] = append(ps[:i], ps[i+1:]...)
					break
				}
			}
		}

		if len(c.players) <= 0 {
			c.forEachCity(func(cty *city) {
				if cty.getCountryID() == c.id {
					cty.setCountryID(0)
				}
			})
			c.setDestory()
		}

		if p.isOnline() {
			p.agent.SetClientFilter("campaign_country", "")
		}
	}
}

func (c *country) addPlayer(p *player) {
	uid := p.getUid()
	if _, ok := c.uid2Player[uid]; !ok {
		c.uid2Player[uid] = p
		c.players = append(c.players, p)

		job := p.getCountryJob()
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

		c.playerNeedSort = true
		if p.isOnline() {
			p.agent.SetClientFilter("campaign_country", strconv.Itoa(int(c.id)))
		}
	}
}

func (c *country) save(isStopServer bool) {
	if isStopServer {
		addContributionPlayersAttr := attribute.NewMapAttr()
		for uid, amount := range c.addContributionPlayers {
			addContributionPlayersAttr.SetFloat64(uid.String(), amount)
		}
		c.attr.SetMapAttr("addContributionPlayers", addContributionPlayersAttr)
	}
	c.attr.Save(isStopServer)
}

func (c *country) addCity(cty *city) {
	c.citys.Add(cty.getCityID())
}

func (c *country) delCity(cty *city) {
	c.citys.Remove(cty.getCityID())
	if c.citys.Size() <= 0 {
		c.setDestory()
	}
}

func (c *country) getID() uint32 {
	return c.id
}

func (c *country) getName() string {
	return c.attr.GetStr("name")
}

func (c *country) getFlag() string {
	return c.attr.GetStr("flag")
}

func (c *country) forEachCity(callback func(cty *city)) {
	c.citys.ForEach(func(cityID int) {
		cty := cityMgr.getCity(cityID)
		if cty != nil {
			callback(cty)
		}
	})
}

func (c *country) getCurBusinessGold() float64 {
	var gold float64
	c.forEachCity(func(cty *city) {
		g, _ := cty.getCurBusinessGold()
		gold += g
	})
	return gold
}

func (c *country) getSalaryByJob(job pb.CampaignJob) int {
	var gold int
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	c.forEachCity(func(cty *city) {
		switch job {
		case pb.CampaignJob_YourMajesty:
			g, _ := cty.getCurBusinessGold()
			gold += int(paramGameData.KingSalary * float64(g))
		case pb.CampaignJob_Counsellor:
			g, _ := cty.getCurBusinessGold()
			gold += int(paramGameData.JunshiSalary * float64(g))
		case pb.CampaignJob_General:
			g, _ := cty.getCurBusinessGold()
			gold += int(paramGameData.ZhonglangjiangSalary * float64(g))
		}
	})
	return gold
}

func (c *country) getPlayerAmount() int {
	var amount int
	c.forEachCity(func(cty *city) {
		amount += cty.getPlayerAmount()
	})
	return amount
}

func (c *country) packSimpleMsg() *pb.CountrySimpleData {
	return &pb.CountrySimpleData{
		CountryID: c.id,
		Name:      c.getName(),
		Flag:      c.getFlag(),
	}
}

func (c *country) getCountryJobPlayers() []*pb.CampaignPlayer {
	var ret []*pb.CampaignPlayer
	jobs := campaignMgr.getCountryJobs()
	for _, job := range jobs {
		ps := c.job2Players[job]
		for _, p := range ps {
			ret = append(ret, p.packMsg(false))
		}
	}
	return ret
}

func (c *country) getAllJobPlayers() map[pb.CampaignJob][]*player {
	return c.job2Players
}

func (c *country) onJobUpdate(p *player, oldJob, newJob pb.CampaignJob) {
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
}

func (c *country) recallJob(p, targetPlayer *player, job pb.CampaignJob, isActive, needSync, needNotice bool) error {
	if job == pb.CampaignJob_YourMajesty {
		return gamedata.GameError(9)
	}
	if p.getCountryID() != targetPlayer.getCountryID() {
		return gamedata.GameError(10)
	}

	if campaignMgr.isCityJob(job) {
		return targetPlayer.getCity().recallJob(p, targetPlayer, job, isActive, needSync, needNotice)
	}

	targetJob := targetPlayer.getCountryJob()
	myJob := p.getCountryJob()
	if targetJob != job {
		return gamedata.GameError(11)
	}

	if myJob != pb.CampaignJob_YourMajesty && !isActive {
		return gamedata.GameError(12)
	}

	jobPlayers := c.job2Players[targetJob]
	for i, jp := range jobPlayers {
		if jp.getUid() == targetPlayer.getUid() {
			c.job2Players[targetJob] = append(jobPlayers[:i], jobPlayers[i+1:]...)
			break
		}
	}
	targetPlayer.setJob(targetPlayer.getCityJob(), pb.CampaignJob_UnknowJob, needSync)
	if needNotice && !isActive {
		noticeMgr.sendNoticeToPlayer(targetPlayer.getUid(), pb.CampaignNoticeType_RecallJobNt, job, targetPlayer.getCityID())

	} else if isActive {

		yourMajesty := c.getYourMajesty()
		if yourMajesty != nil {
			noticeMgr.sendNoticeToPlayer(yourMajesty.getUid(), pb.CampaignNoticeType_ResignNt, job, targetPlayer.getName(),
				targetPlayer.getCityID())
		}
	}

	glog.Infof("country recallJob countryID=%d, uid=%d, targetUid=%d, myJob=%s, targetJob=%s", p.getUid(),
		c.id, targetPlayer.getUid(), myJob, targetJob)

	return nil
}

// 玩家p，把玩家oldUid的职位job，任命给玩家targetPlayer
func (c *country) appointJob(p, targetPlayer *player, job pb.CampaignJob, oldUid common.UUid) error {
	targetUid := targetPlayer.getUid()
	if targetUid == oldUid {
		return gamedata.GameError(10)
	}

	if p.getCountryID() != targetPlayer.getCountryID() {
		return gamedata.GameError(10)
	}

	targetJob := targetPlayer.getCountryJob()
	if targetJob == job || targetPlayer.getCityJob() == job {
		return gamedata.GameError(10)
	}

	if job >= pb.CampaignJob_Prefect {
		// 城市官员
		return targetPlayer.getCity().appointJob(p, targetPlayer, job, oldUid)
	}

	myJob := p.getCountryJob()
	if myJob != pb.CampaignJob_YourMajesty {
		return gamedata.GameError(11)
	}

	var jobMaxAmount int
	switch job {
	case pb.CampaignJob_Counsellor:
		jobMaxAmount = maxCounsellorAmount
	case pb.CampaignJob_General:
		jobMaxAmount = maxGeneralAmount
	default:
		return gamedata.GameError(12)
	}

	jobPlayers := c.job2Players[job]
	var oldPlayer *player
	if oldUid > 0 {
		// 罢免oldPlayer
		oldPlayer, _ = playerMgr.loadPlayer(oldUid)
		if oldPlayer == nil || oldPlayer.getCountryJob() != job {
			return gamedata.GameError(13)
		}

		err := c.recallJob(p, oldPlayer, job, false, true, true)
		if err != nil {
			return err
		}
	} else if len(jobPlayers) >= jobMaxAmount {
		return gamedata.GameError(14)
	}

	if targetJob != pb.CampaignJob_UnknowJob {
		// 罢免targetPlayer原来的职位
		err := c.recallJob(p, targetPlayer, targetJob, false, false, false)
		if err != nil {
			return err
		}
	}

	// 任命targetPlayer
	targetPlayer.setJob(targetPlayer.getCityJob(), job, true)
	noticeMgr.sendNoticeToCountry(c.id, pb.CampaignNoticeType_AppointJobNt, p.getName(),
		targetPlayer.getName(), job, targetPlayer.getCityID())

	glog.Infof("country appointJob countryID=%d, uid=%d, targetUid=%d, oldUid=%d, myJob=%s, targetOldJob=%s, targetJob=%s",
		c.id, p.getUid(), targetPlayer.getUid(), oldUid, myJob, targetJob, job)
	return nil
}

func (c *country) sortPlayers() {
	if !c.playerNeedSort {
		return
	}
	c.playerNeedSort = false
	c.players = sortPlayers(c.players)
}

func (c *country) chooseNewYourMajesty() *player {
	var newYourMajesty *player

	jobs := campaignMgr.getCountryJobs()
L:
	for _, job := range jobs {
		if job == pb.CampaignJob_YourMajesty {
			continue
		}
		ps := c.job2Players[job]
		ps = sortPlayers(ps)
		for _, p := range ps {
			if p.getCountryJob() == pb.CampaignJob_YourMajesty {
				continue
			}
			newYourMajesty = p
			break L
		}
	}

	if newYourMajesty == nil {
		var prefect *player
		var playerAmount int
		c.forEachCity(func(cty *city) {
			p := cty.getPrefect()
			if p == nil || p.getCountryJob() == pb.CampaignJob_YourMajesty {
				return
			}
			if prefect == nil || cty.getPlayerAmount() > playerAmount {
				prefect = p
				playerAmount = cty.getPlayerAmount()
			}
		})
		newYourMajesty = prefect
	}

	if newYourMajesty == nil {
		var score int
		for _, p := range c.players {
			if p.getCountryJob() == pb.CampaignJob_YourMajesty {
				continue
			}
			if newYourMajesty == nil || p.getPvpScore() > score {
				newYourMajesty = p
				score = p.getPvpScore()
			}
		}
	}

	return newYourMajesty
}

func (c *country) changeYourMajesty(newYourMajesty, oldYourMajesty *player) {
	newYourMajesty.setJob(newYourMajesty.getCityJob(), pb.CampaignJob_YourMajesty, true)
	noticeMgr.sendNoticeToCountry(c.getID(), pb.CampaignNoticeType_YourMajestyChangeNt, oldYourMajesty.getName(),
		newYourMajesty.getName())

	if !c.isKingdom() {
		c.setName(createCountryMgr.genCountryName(newYourMajesty.getName()))
		campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_COUNTRY_NAME, &pb.UpdateCountryNameArg{
			CountryID: c.getID(),
			Name:      c.getName(),
		})
	}
}

func (c *country) checkPlayerAmount() {
	if c.getPlayerAmount() <= 0 {
		c.setDestory()
		c.forEachCity(func(cty *city) {
			if cty.getCountryID() == c.getID() {
				cty.setCountryID(0)
			}
		})
	}
}

func (c *country) onProduction() {
	cityGoldProfits := map[int]float64{}
	c.forEachCity(func(cty *city) {
		if cty.getCountryID() != c.id {
			return
		}
		cityGoldProfits[cty.getCityID()] = cty.onProduction()
	})

	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	for _, p := range c.players {
		countryJob := p.getCountryJob()
		cityJob := p.getCityJob()
		var gold float64
		if countryJob != pb.CampaignJob_UnknowJob {
			for cityID, cityGold := range cityGoldProfits {
				var gold2 float64
				switch countryJob {
				case pb.CampaignJob_YourMajesty:
					gold2 = paramGameData.KingSalary * cityGold
				case pb.CampaignJob_Counsellor:
					gold2 = paramGameData.JunshiSalary * cityGold
				case pb.CampaignJob_General:
					gold2 = paramGameData.ZhonglangjiangSalary * cityGold
				}
				gold += gold2
				cityMgr.getCity(cityID).modifyResource(resGold, -gold2)
			}
		}

		if cityJob != pb.CampaignJob_UnknowJob {
			cityID := p.getCityID()
			cityGold := cityGoldProfits[cityID]
			var gold2 float64
			switch cityJob {
			case pb.CampaignJob_Prefect:
				gold2 = paramGameData.TaishouSalary * cityGold
			case pb.CampaignJob_DuWei:
				gold2 = paramGameData.DuweiSalary * cityGold
			case pb.CampaignJob_FieldOfficer:
				gold2 = paramGameData.XiaoweiSalary * cityGold
			}
			gold += gold2
			cityMgr.getCity(cityID).modifyResource(resGold, -gold2)
		}

		if gold > 0 {
			utils.PlayerMqPublish(p.getUid(), pb.RmqType_Bonus, &pb.RmqBonus{
				ChangeRes: []*pb.Resource{&pb.Resource{Type: int32(consts.Gold), Amount: int32(gold)}},
			})
		}

		if cityJob != pb.CampaignJob_UnknowJob || countryJob != pb.CampaignJob_UnknowJob {
			noticeMgr.sendNoticeToPlayer(p.getUid(), pb.CampaignNoticeType_SalaryNt, int(gold))
		}
	}
}

func (c *country) getYourMajesty() *player {
	ps := c.job2Players[pb.CampaignJob_YourMajesty]
	if len(ps) <= 0 {
		return nil
	} else {
		return ps[0]
	}
}

func (c *country) randomCity() *city {
	cityIDs := c.citys.ToList()
	if len(cityIDs) <= 0 {
		return nil
	}
	return cityMgr.getCity(cityIDs[rand.Intn(len(cityIDs))])
}

func (c *country) onPlayerAddContribution(p *player, val float64) {
	old := c.addContributionPlayers[p.getUid()]
	c.addContributionPlayers[p.getUid()] = old + val
	c.addContributions += val
}

func (c *country) calcPerContribution() {
	if c.addContributions <= 0 {
		return
	}

	for _, ps := range c.job2Players {
		for _, p := range ps {
			p.addContribution((c.addContributions-c.addContributionPlayers[p.getUid()])*
				campaignMgr.getContributionPerByJob(p.getCountryJob()), false, false)
		}
	}

	c.addContributions = 0
	c.addContributionPlayers = map[common.UUid]float64{}
}

func (c *country) surrender(p *player, targetCry *country) error {
	if p.getCountryJob() != pb.CampaignJob_YourMajesty {
		return gamedata.GameError(10)
	}

	noticeMgr.sendNoticeToCountry(targetCry.getID(), pb.CampaignNoticeType_SurrenderCountry2Nt, p.getName(), c.getName())
	c.forEachCity(func(cty *city) {
		cty.doSurrender(p, targetCry, pb.CampaignNoticeType_SurrenderCountry1Nt)
	})

	p.subContribution(p.getMaxContribution()*0.2, true)
	c.setDestory()
	p.setLastCountryID(0, true)
	glog.Infof("country surrender, countryID=%d, uid=%d", c.id, p.getUid())
	return nil
}
