package main

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"time"
	"kinger/gopuppy/common/timer"
	"fmt"
	"strings"
	"unicode"
	"kinger/gamedata"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
)

var createCountryMgr = &createCountryMgrSt{}

type createCountryMgrSt struct {
	canApplyCreateCountry_ bool
	nextApplyCreateCountryStopTime time.Time
	uid2AyApply map[common.UUid]*autocephalyApply
	city2AyApply map[int]*autocephalyApply
}

func (ccm *createCountryMgrSt) canApplyCreateCountry() bool {
	return ccm.canApplyCreateCountry_
}

func (ccm *createCountryMgrSt) initialize() {
	ccm.canApplyCreateCountry_ = false
	ccm.uid2AyApply = map[common.UUid]*autocephalyApply{}
	ccm.city2AyApply = map[int]*autocephalyApply{}
	now := time.Now()
	if warMgr.isNormalState() {
		stopTime := time.Date(now.Year(), now.Month(), now.Day(), createCountryEndHour, createCountryEndMin, 0,
			0, now.Location())
		beginTime := time.Date(now.Year(), now.Month(), now.Day(), coutryWarEndHour, coutryWarEndMin, 0, 0,
			now.Location())
		if now.Before(stopTime) || now.After(beginTime) {
			ccm.canApplyCreateCountry_ = true
		}
	}

	attrs, err := attribute.LoadAll(campaignMgr.genAttrName("autocephaly"))
	if err != nil {
		panic(err)
	}
	for _, attr := range attrs {
		id, ok := attr.GetAttrID().(int64)
		if !ok {
			panic(fmt.Sprintf("wrong autocephaly uid %s", attr.GetAttrID()))
		}
		uid := common.UUid(id)
		if _, ok := ccm.uid2AyApply[uid]; ok {
			continue
		}

		p, _ := playerMgr.loadPlayer(uid)
		if p == nil {
			continue
		}

		aa := newAutocephalyApplyByAttr(uid, attr)
		ccm.uid2AyApply[uid] = aa
		ccm.city2AyApply[aa.getCityID()] = aa
	}

	//ccm.checkAutocephalyApply()
	ccm.nextApplyCreateCountryStopTime = now.Add(timer.TimeDelta(createCountryEndHour, createCountryEndMin, 0))
	timer.RunEveryDay(createCountryEndHour, createCountryEndMin, 0, ccm.onApplyCreateCountryEnd)
	eventhub.Subscribe(evPlayerChangeCity, ccm.onPlayerChangeCity)
	eventhub.Subscribe(evPlayerChangeCountry, ccm.onPlayerChangeCity)
	eventhub.Subscribe(evPlayerChangeCityJob, ccm.onPlayerChangeCityJob)
	//eventhub.Subscribe(evWarEnd, ccm.checkAutocephalyApply)
}

/*
func (ccm *createCountryMgrSt) checkAutocephalyApply(_ ...interface{}) {
	if !warMgr.isNormalState() {
		return
	}
	for _, aa := range ccm.uid2AyApply {
		if aa.isPass() {
			ccm.onAutocephalySuccess(aa)
		} else if aa.isCancel() {
			ccm.onAutocephalyFail(aa)
		}
	}
}
*/

func (ccm *createCountryMgrSt) onApplyCreateCountryBegin() {
	ccm.canApplyCreateCountry_ = true
}

func (ccm *createCountryMgrSt) onApplyCreateCountryEnd() {
	if warMgr.isPause() {
		return
	}
	ccm.canApplyCreateCountry_ = true
	cityMgr.citysCreateCountry()
	ccm.nextApplyCreateCountryStopTime = time.Now().Add(timer.TimeDelta(createCountryEndHour, createCountryEndMin, 0))
}

func (ccm *createCountryMgrSt) getApplyCreateCountryRemainTime() time.Duration {
	return ccm.nextApplyCreateCountryStopTime.Sub(time.Now())
}

func (ccm *createCountryMgrSt) onPlayerChangeCity(args ...interface{}) {
	p := args[0].(*player)
	uid := p.getUid()
	if aa, ok := ccm.uid2AyApply[uid]; ok {
		ccm.onAutocephalyFail(aa)
	}
}

func (ccm *createCountryMgrSt) onPlayerChangeCityJob(args ...interface{}) {
	p := args[0].(*player)
	oldCityJob := args[1].(pb.CampaignJob)
	newCityJob := args[2].(pb.CampaignJob)
	if oldCityJob != pb.CampaignJob_Prefect || newCityJob == pb.CampaignJob_Prefect {
		return
	}

	uid := p.getUid()
	if aa, ok := ccm.uid2AyApply[uid]; ok {
		ccm.onAutocephalyFail(aa)
	}
}

func (ccm *createCountryMgrSt) getCityAutocephaly(cityID int) *autocephalyApply {
	return ccm.city2AyApply[cityID]
}

func (ccm *createCountryMgrSt) applyAutocephaly(uid common.UUid, cty *city) error {
	aa := newAutocephalyApply(uid, cty)
	if aa == nil {
		return gamedata.GameError(10)
	}
	glog.Infof("applyAutocephaly uid=%d, cityID=%d", uid, cty.getCityID())
	ccm.uid2AyApply[uid] = aa
	ccm.city2AyApply[cty.getCityID()] = aa
	return nil
}

func (ccm* createCountryMgrSt) onAutocephalyFail(aa *autocephalyApply) {
	aa.attr.Delete(false)
	delete(ccm.uid2AyApply, aa.getUid())
	delete(ccm.city2AyApply, aa.getCityID())
}

func (ccm *createCountryMgrSt) onAutocephalySuccess(aa *autocephalyApply) {
	if !warMgr.isNormalState() {
		return
	}
	uid := aa.getUid()
	cityID := aa.getCityID()
	if ccm.uid2AyApply[uid] != aa {
		return
	}
	if ccm.city2AyApply[cityID] != aa {
		return
	}

	aa.attr.Delete(false)
	delete(ccm.uid2AyApply, uid)
	delete(ccm.city2AyApply, cityID)

	p := playerMgr.getPlayer(uid)
	if p == nil || p.getCityJob() != pb.CampaignJob_Prefect || p.getCountryJob() == pb.CampaignJob_YourMajesty {
		return
	}
	cry := p.getCountry()
	if cry == nil {
		return
	}

	cty := p.getCity()
	if cty == nil || cty.getCityID() != cityID {
		return
	}

	cty.autocephaly(false)
}

func (ccm *createCountryMgrSt) genCountryName(playerName string) string {
	if strings.HasPrefix(playerName, "#c") && len(playerName) > 10 {
		playerName = playerName[8:]
		playerName = playerName[:len(playerName) - 2]
	}
	var runes []rune
	for i, r := range playerName {
		runes = append(runes, r)
		if i == 0 {
			if unicode.Is(unicode.Scripts["Han"], r) {
				break
			}
		} else {
			break
		}
	}
	return string(runes)
}


type createCountryApply struct {
	uid common.UUid
	attr *attribute.MapAttr
}

func newCreateCountryApply(p *player, gold int) *createCountryApply {
	attr := attribute.NewMapAttr()
	attr.SetInt("gold", gold)
	attr.SetStr("countryName", createCountryMgr.genCountryName(p.getName()))
	return newCreateCountryApplyByAttr(p.getUid(), attr)
}

func newCreateCountryApplyByAttr(uid common.UUid, attr *attribute.MapAttr) *createCountryApply {
	return &createCountryApply{
		uid: uid,
		attr: attr,
	}
}

func (ca *createCountryApply) getCountryName() string {
	return ca.attr.GetStr("countryName")
}

func (ca *createCountryApply) getGold() int {
	return ca.attr.GetInt("gold")
}

func (ca *createCountryApply) addGold(gold int) {
	ca.attr.SetInt("gold", ca.getGold() + gold)
}

func (ca *createCountryApply) getUid() common.UUid {
	return ca.uid
}

func (ca *createCountryApply) battleThan(oth *createCountryApply) bool {
	gold1 := ca.getGold()
	gold2 := oth.getGold()
	if gold1 > gold2 {
		return true
	} else if gold1 < gold2 {
		return false
	}

	p1, _ := playerMgr.loadPlayer(ca.getUid())
	if p1 == nil {
		return false
	}
	p2, _ := playerMgr.loadPlayer(oth.getUid())
	if p2 == nil {
		return true
	}

	score1 := p1.getPvpScore()
	score2 := p2.getPvpScore()
	if score1 > score2 {
		return true
	} else if score1 < score2 {
		return false
	}
	return ca.getUid() >= oth.getUid()
}

func (ca *createCountryApply) packMsg() *pb.ApplyCreateCountryPlayer {
	p, _ := playerMgr.loadPlayer(ca.getUid())
	if p == nil {
		return nil
	}
	return &pb.ApplyCreateCountryPlayer{
		Player: p.packMsg(false),
		Gold: int32(ca.getGold()),
	}
}

type autocephalyApply struct {
	uid common.UUid
	attr *attribute.AttrMgr
	voteAttr *attribute.MapAttr
}

func newAutocephalyApplyByAttr(uid common.UUid, attr *attribute.AttrMgr) *autocephalyApply {
	return &autocephalyApply{
		uid: uid,
		attr: attr,
		voteAttr: attr.GetMapAttr("vote"),
	}
}

func newAutocephalyApply(uid common.UUid, cty *city) *autocephalyApply {
	attr := attribute.NewAttrMgr(campaignMgr.genAttrName("autocephaly"), uid)
	attr.SetInt("cityID", cty.getCityID())
	voteAttr := attribute.NewMapAttr()
	attr.SetMapAttr("vote", voteAttr)
	allPlayers := cty.getAllPlayers()
	var ps []*player
	for _, p := range allPlayers {
		countryJob := p.getCountryJob()
		cityJob := p.getCityJob()
		if countryJob == pb.CampaignJob_YourMajesty {
			return nil
		}
		if cityJob == pb.CampaignJob_Prefect {
			continue
		}
		if cityJob == pb.CampaignJob_UnknowJob && countryJob == pb.CampaignJob_UnknowJob {
			continue
		}

		voteAttr.SetInt(p.getUid().String(), 1)
		ps = append(ps, p)
	}

	for _, p := range ps {
		noticeMgr.sendNoticeToPlayer(p.getUid(), pb.CampaignNoticeType_AutocephalyVoteNt)
	}

	attr.Save(false)
	return newAutocephalyApplyByAttr(uid, attr)
}

func (aa *autocephalyApply) packMsg() *pb.AutocephalyInfo {
	reply := &pb.AutocephalyInfo{
		CountryName: "1",
	}
	aa.voteAttr.ForEachKey(func(key string) {
		isAgree := aa.voteAttr.GetInt(key) == 2
		if isAgree {
			p := playerMgr.getPlayer(common.ParseUUidFromString(key))
			if p != nil {
				reply.AgreePlayers = append(reply.AgreePlayers, p.packMsg(false))
			}
		}
	})
	return reply
}

func (aa *autocephalyApply) getUid() common.UUid {
	return aa.uid
}

func (aa *autocephalyApply) getCityID() int {
	return aa.attr.GetInt("cityID")
}

func (aa *autocephalyApply) vote(uid common.UUid, isAgree bool) {
	key := uid.String()
	if !aa.voteAttr.HasKey(key) {
		return
	}

	val := 3
	if isAgree {
		val = 2
	}
	aa.voteAttr.SetInt(key, val)
	aa.attr.Save(false)

	if !isAgree {
		createCountryMgr.onAutocephalyFail(aa)
	} else if aa.isPass() {
		createCountryMgr.onAutocephalySuccess(aa)
	}
}

func (aa *autocephalyApply) isPass() bool {
	pass := true
	aa.attr.ForEachKey(func(key string) {
		if !pass {
			return
		}

		vote := aa.attr.GetInt(key)
		if vote == 3 {
			pass = false
			return
		}

		if vote == 0 {
			p := playerMgr.getPlayer(common.ParseUUidFromString(key))
			if p == nil || p.getCityID() != aa.getCityID() {
				return
			} else {
				pass = false
				return
			}
		}
	})
	return pass
}

func (aa *autocephalyApply) isCancel() bool {
	is := aa.attr.GetBool("isCancel")
	if is {
		return true
	}

	aa.attr.ForEachKey(func(key string) {
		if is {
			return
		}
		vote := aa.attr.GetInt(key)
		if vote == 3 {
			is = true
			return
		}
	})
	return is
}
