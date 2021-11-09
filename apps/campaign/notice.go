package main

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"math/rand"
	"sort"
	"time"
)

var noticeMgr = &noticeMgrSt{}

type noticeMgrSt struct {
	// 全世界玩家
	worldNotices *noticeBoard
	// 全城玩家
	cityNotices map[int]*cityNoticeBoard
	// 全国玩家
	countryNotices map[uint32]*countryNoticeBoard
	// 某个玩家
	playerNotices map[common.UUid]*playerNoticeBoard
	// 玩家不在线时收到的
	offlinePlayerNotices        map[common.UUid]*noticeBoard
	offlinePlayerNoticesLoading map[common.UUid]chan struct{}
}

func (nm *noticeMgrSt) initialize() {
	nm.cityNotices = map[int]*cityNoticeBoard{}
	nm.countryNotices = map[uint32]*countryNoticeBoard{}
	nm.playerNotices = map[common.UUid]*playerNoticeBoard{}
	nm.offlinePlayerNotices = map[common.UUid]*noticeBoard{}
	nm.offlinePlayerNoticesLoading = map[common.UUid]chan struct{}{}

	wordNoticesAttr := attribute.NewAttrMgr(campaignMgr.genAttrName("worldNotices2"), 1)
	err := wordNoticesAttr.Load()
	if err != nil && err != attribute.NotExistsErr {
		panic(err)
	}
	nm.worldNotices = newNoticeBoard(wordNoticesAttr)

	cityMgr.forEachCity(func(cty *city) bool {
		attr := attribute.NewAttrMgr(campaignMgr.genAttrName("cityNotices2"), cty.getCityID())
		err := attr.Load()
		if err != nil && err != attribute.NotExistsErr {
			panic(err)
		}
		nm.cityNotices[cty.getCityID()] = newCityNoticeBoard(cty.getCityID(), attr)
		return true
	})

	allCountrys := countryMgr.getAllCountrys()
	for _, cry := range allCountrys {
		attr := attribute.NewAttrMgr(campaignMgr.genAttrName("countryNotices2"), cry.getID())
		err := attr.Load()
		if err != nil && err != attribute.NotExistsErr {
			panic(err)
		}
		nm.countryNotices[cry.getID()] = newCountryNoticeBoard(cry.getID(), attr)
	}

	eventhub.Subscribe(evPlayerLogin, nm.onPlayerLogin)
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, nm.onPlayerLogut)
	eventhub.Subscribe(evPlayerChangeCity, nm.onPlayerChangeCity)
	eventhub.Subscribe(evPlayerChangeCountry, nm.onPlayerChangeCountry)

	timer.AddTicker(time.Duration(rand.Intn(20)+290)*time.Second, func() {
		nm.save(false)
	})
	timer.AddTicker(20*time.Second, nm.checkTimeout)
}

func (nm *noticeMgrSt) loadOfflinePlayerNotices(uid common.UUid) (*noticeBoard, bool) {
	nb, ok := nm.offlinePlayerNotices[uid]
	if ok {
		return nb, true
	}

	c, ok := nm.offlinePlayerNoticesLoading[uid]
	if ok {
		evq.Await(func() {
			<-c
		})
		return nm.offlinePlayerNotices[uid], true
	}

	c = make(chan struct{})
	nm.offlinePlayerNoticesLoading[uid] = c
	defer func() {
		delete(nm.offlinePlayerNoticesLoading, uid)
		close(c)
	}()

	oAttr := attribute.NewAttrMgr(campaignMgr.genAttrName("offlinePlayerNotices2"), uid)
	err := oAttr.Load()
	nb, ok = nm.offlinePlayerNotices[uid]
	if ok {
		return nb, true
	}

	if err != nil && err != attribute.NotExistsErr {
		return nil, false
	}

	isExist := true
	if err == attribute.NotExistsErr {
		isExist = false
	}

	nb = newNoticeBoard(oAttr)
	nm.offlinePlayerNotices[uid] = nb
	return nb, isExist
}

func (nm *noticeMgrSt) sendNoticeToCountry(countryID uint32, type_ pb.CampaignNoticeType, args ...interface{}) {
	nb, ok := nm.countryNotices[countryID]
	if !ok {
		nb = newCountryNoticeBoard(countryID, attribute.NewAttrMgr(campaignMgr.genAttrName("countryNotices2"), countryID))
		nm.countryNotices[countryID] = nb
	}
	nb.sendNotice(type_, args...)
}

func (nm *noticeMgrSt) sendNoticeToCity(cityID int, type_ pb.CampaignNoticeType, args ...interface{}) {
	nb, ok := nm.cityNotices[cityID]
	if ok {
		nb.sendNotice(type_, args...)
	}
}

func (nm *noticeMgrSt) getWorldNotice() *noticeBoard {
	return nm.worldNotices
}

func (nm *noticeMgrSt) getCountryNotice(countryID uint32) *countryNoticeBoard {
	return nm.countryNotices[countryID]
}

func (nm *noticeMgrSt) getCityNotice(cityID int) *cityNoticeBoard {
	return nm.cityNotices[cityID]
}

func (nm *noticeMgrSt) getPlayerNotice(uid common.UUid) *playerNoticeBoard {
	return nm.playerNotices[uid]
}

func (nm *noticeMgrSt) getAllPlayerNotices() map[common.UUid]*playerNoticeBoard {
	return nm.playerNotices
}

func (nm *noticeMgrSt) onPlayerChangeCity(args ...interface{}) {
	p := args[0].(*player)
	uid := p.getUid()
	nb, ok := nm.playerNotices[uid]
	if !ok {
		return
	}

	if cnb, ok := nm.cityNotices[p.getCityID()]; ok {
		nb.setCityMaxID(cnb.getMaxID())
	}
}

func (nm *noticeMgrSt) onPlayerChangeCountry(args ...interface{}) {
	p := args[0].(*player)
	uid := p.getUid()
	nb, ok := nm.playerNotices[uid]
	if !ok {
		return
	}

	if cnb, ok := nm.countryNotices[p.getCountryID()]; ok {
		nb.setCountryMaxID(cnb.getMaxID())
	}
}

func (nm *noticeMgrSt) onPlayerLogut(args ...interface{}) {
	uid := args[0].(*logic.PlayerAgent).GetUid()
	if nb, ok := nm.playerNotices[uid]; ok {
		delete(nm.playerNotices, uid)
		nb.save(false)
	}
}

func (nm *noticeMgrSt) onPlayerLogin(args ...interface{}) {
	p := args[0].(*player)
	uid := p.getUid()
	if _, ok := nm.playerNotices[uid]; ok {
		return
	}

	attr := attribute.NewAttrMgr(campaignMgr.genAttrName("playerNotices2"), uid)
	err := attr.Load()
	cityID := p.getCityID()
	countryID := p.getCountryID()
	nb := newPlayerNoticeBoard(uid, attr)
	if err != nil {
		if err == attribute.NotExistsErr {
			nb = newPlayerNoticeBoard(uid, attr)
			nb.setWorldMaxID(nm.worldNotices.getMaxID())

			if cnb, ok := nm.cityNotices[cityID]; ok {
				nb.setCityMaxID(cnb.getMaxID())
			}

			if cnb, ok := nm.countryNotices[countryID]; ok {
				nb.setCountryMaxID(cnb.getMaxID())
			}
			attr.Save(false)
		} else {
			glog.Errorf("noticeMgrSt onPlayerLogin load error, uid=%d, err=%s", uid, err)
			return
		}
	}
	nm.playerNotices[uid] = nb

	var offlineNotices []*notice
	wMaxID := nb.getWorldMaxID()
	crMaxID := nb.getCountryMaxID()
	ctMaxID := nb.getCityMaxID()
	offlineNotices = append(offlineNotices, nm.worldNotices.getOfflineNotices(wMaxID)...)
	nb.setWorldMaxID(nm.worldNotices.getMaxID())
	if cnb, ok := nm.cityNotices[cityID]; ok {
		offlineNotices = append(offlineNotices, cnb.getOfflineNotices(ctMaxID)...)
		nb.setCityMaxID(cnb.getMaxID())
	}
	if cnb, ok := nm.countryNotices[countryID]; ok {
		offlineNotices = append(offlineNotices, cnb.getOfflineNotices(crMaxID)...)
		nb.setCountryMaxID(cnb.getMaxID())
	}

	onb, exist := nm.loadOfflinePlayerNotices(uid)
	if onb != nil {
		delete(nm.offlinePlayerNotices, uid)
		offlineNotices = append(offlineNotices, onb.getNotices()...)
		if exist {
			onb.attr.Delete(false)
		}
	}

	sort.Slice(offlineNotices, func(i, j int) bool {
		return offlineNotices[i].getTime() <= offlineNotices[j].getTime()
	})
	for _, n := range offlineNotices {
		nb.addOfflineNotice(n)
	}
}

func (nm *noticeMgrSt) checkTimeout() {
	now := time.Now().Unix()
	nm.worldNotices.checkTimeout(now)
	for _, nb := range nm.cityNotices {
		nb.checkTimeout(now)
	}
	for _, nb := range nm.countryNotices {
		nb.checkTimeout(now)
	}
	for _, nb := range nm.offlinePlayerNotices {
		nb.checkTimeout(now)
	}
}

func (nm *noticeMgrSt) save(isStopServer bool) {
	nm.worldNotices.save(isStopServer)
	for _, nb := range nm.cityNotices {
		nb.save(isStopServer)
	}
	for _, nb := range nm.countryNotices {
		nb.save(isStopServer)
	}
	for _, nb := range nm.playerNotices {
		nb.save(isStopServer)
	}
	for _, nb := range nm.offlinePlayerNotices {
		nb.save(isStopServer)
	}
}

func (nm *noticeMgrSt) sendNoticeToPlayer(uid common.UUid, type_ pb.CampaignNoticeType, args ...interface{}) {
	if nb, ok := nm.playerNotices[uid]; ok {
		nb.sendNotice(type_, args...)
	} else {
		onb, _ := nm.loadOfflinePlayerNotices(uid)
		if nb, ok := nm.playerNotices[uid]; ok {
			nb.sendNotice(type_, args...)
		} else if onb != nil {
			onb.newNotice(type_, args...)
		}
	}
}

type notice struct {
	attr    *attribute.MapAttr
	lowData map[string]interface{}
}

func newNotice(id int, type_ pb.CampaignNoticeType, args string) *notice {
	attr := attribute.NewMapAttr()
	attr.SetInt("type", int(type_))
	attr.SetInt("id", id)
	attr.SetInt("t", int(time.Now().Unix()))
	attr.SetStr("args", args)
	return newNoticeByAttr(attr)
}

func newNoticeByAttr(attr *attribute.MapAttr) *notice {
	return &notice{
		attr: attr,
	}
}

func (n *notice) setAutocephalyCity(cityID int) {
	n.attr.SetInt("autocephalyCity", cityID)
}

func (n *notice) getAutocephalyCity() int {
	return n.attr.GetInt("autocephalyCity")
}

func (n *notice) getAutocephalyPlayer() common.UUid {
	return common.UUid(n.attr.GetUInt64("autocephalyPlayer"))
}

func (n *notice) setAutocephalyPlayer(uid common.UUid) {
	n.attr.SetUInt64("autocephalyPlayer", uint64(uid))
}

func (n *notice) copy() *notice {
	if n.lowData == nil {
		n.lowData = n.attr.ToMap()
	}
	attr := attribute.NewMapAttr()
	attr.AssignMap(n.lowData)
	return newNoticeByAttr(attr)
}

func (n *notice) getID() int {
	return n.attr.GetInt("id")
}

func (n *notice) setID(id int) {
	n.attr.SetInt("id", id)
}

func (n *notice) getType() pb.CampaignNoticeType {
	return pb.CampaignNoticeType(n.attr.GetInt("type"))
}

func (n *notice) getTime() int {
	return n.attr.GetInt("t")
}

func (n *notice) isOp() bool {
	return n.attr.GetBool("isOp")
}

func (n *notice) op() {
	n.attr.SetBool("isOp", true)
}

func (n *notice) packMsg() *pb.CampaignNotice {
	return &pb.CampaignNotice{
		ID:   int32(n.getID()),
		Type: n.getType(),
		Time: int32(n.getTime()),
		Args: []byte(n.attr.GetStr("args")),
	}
}

func (n *notice) isTimeut(now int64) bool {
	return int64(n.getTime()+noticeTimeout) <= now
}

type noticeBoard struct {
	attr        *attribute.AttrMgr
	noticesAttr *attribute.ListAttr
	notices     []*notice
}

func newNoticeBoard(attr *attribute.AttrMgr) *noticeBoard {
	nb := &noticeBoard{
		attr: attr,
	}
	nb.initNotices()
	return nb
}

func (nb *noticeBoard) initNotices() {
	nb.noticesAttr = nb.attr.GetListAttr("notices")
	if nb.noticesAttr == nil {
		nb.noticesAttr = attribute.NewListAttr()
		nb.attr.SetListAttr("notices", nb.noticesAttr)
	}

	now := time.Now().Unix()
	i := -1
	nb.noticesAttr.ForEachIndex(func(index int) bool {
		nAttr := nb.noticesAttr.GetMapAttr(index)
		n := newNoticeByAttr(nAttr)
		if !n.isTimeut(now) {
			nb.notices = append(nb.notices, n)
		} else {
			i = index
		}
		return true
	})

	if i >= 0 {
		nb.noticesAttr.DelBySection(0, i+1)
	}
}

func (nb *noticeBoard) checkTimeout(now int64) {
	index := -1
	for i, n := range nb.notices {
		if !n.isTimeut(now) {
			break
		} else {
			index = i
		}
	}

	if index >= 0 {
		nb.notices = nb.notices[index+1:]
		nb.noticesAttr.DelBySection(0, index+1)
	}
}

func (nb *noticeBoard) save(isStopServer bool) {
	nb.attr.Save(isStopServer)
}

func (nb *noticeBoard) getMaxID() int {
	return nb.attr.GetInt("maxID")
}

func (nb *noticeBoard) genMaxID() int {
	id := nb.getMaxID() + 1
	nb.attr.SetInt("maxID", id)
	return id
}

func (nb *noticeBoard) getOfflineNotices(maxID int) []*notice {
	var ns []*notice
	for _, n := range nb.notices {
		if n.getID() > maxID {
			ns = append(ns, n)
		}
	}
	return ns
}

func (nb *noticeBoard) getNotices() []*notice {
	return nb.notices
}

func (nb *noticeBoard) getNoticeByID(id int) *notice {
	for _, n := range nb.notices {
		if n.getID() == id {
			return n
		}
	}
	return nil
}

func (nb *noticeBoard) newNotice(type_ pb.CampaignNoticeType, args ...interface{}) *notice {
	var argsData []byte
	switch type_ {
	case pb.CampaignNoticeType_NewCountryNt:
		arg := &pb.NewCountryNtArg{
			PlayerName:  args[0].(string),
			CountryName: args[1].(string),
			CityID:      int32(args[2].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_AppointJobNt:
		arg := &pb.AppointJobNtArg{
			PlayerName:       args[0].(string),
			TargetPlayerName: args[1].(string),
			Job:              args[2].(pb.CampaignJob),
			CityID:           int32(args[3].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_RecallJobNt:
		arg := &pb.RecallJobNtArg{
			Job:    args[0].(pb.CampaignJob),
			CityID: int32(args[1].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_AutocephalyNt:
		arg := &pb.AutocephalyNtArg{
			Job:            args[0].(pb.CampaignJob),
			PlayerName:     args[1].(string),
			CountryName:    args[2].(string),
			NewCountryName: args[3].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_KickOutNt:
		arg := &pb.KickOutNtArg{
			Job:        args[0].(pb.CampaignJob),
			PlayerName: args[1].(string),
			CityID:     int32(args[2].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_YourMajestyChangeNt:
		arg := &pb.YourMajestyChangeNtArg{
			YourMajestyName:    args[0].(string),
			NewYourMajestyName: args[1].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_ResignNt:
		arg := &pb.ResignNtArg{
			Job:        args[0].(pb.CampaignJob),
			PlayerName: args[1].(string),
			CityID:     int32(args[2].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_BeOccupyNt:
		arg := &pb.BeOccupyNtArg{
			CountryName:    args[0].(string),
			BeOccupyCityID: int32(args[1].(int)),
			CaptiveAmount:  int32(args[2].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_DestoryCountryNt:
		arg := &pb.DestoryCountryNtArg{
			CountryName:          args[0].(string),
			BeDestoryCountryName: args[1].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_UnifiedWordNt:
		arg := &pb.UnifiedWordNtArg{
			CountryName:     args[0].(string),
			YourMajestyName: args[1].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_CapitalInjectionNt:
		arg := &pb.CapitalInjectionNtArg{
			Job:        args[0].(pb.CampaignJob),
			PlayerName: args[1].(string),
			Gold:       int32(args[2].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_ProductionNt:
		arg := &pb.ProductionNtArg{
			Gold:   int32(args[0].(int)),
			Forage: int32(args[1].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_SalaryNt:
		arg := &pb.SalaryNtArg{
			Gold: int32(args[0].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_TransportNt:
		arg := &pb.TransportNtArg{
			FromCity:      int32(args[0].(int)),
			TargetCity:    int32(args[1].(int)),
			TransportType: args[2].(pb.TransportTypeEnum),
			Amount:        int32(args[3].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_OccupyNt:
		arg := &pb.OccupyNtArg{
			OccupyCityID:  int32(args[0].(int)),
			CaptiveAmount: int32(args[1].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_SurrenderNt:
		arg := &pb.SurrenderNtArg{
			PlayerName: args[0].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_BetrayNt:
		arg := &pb.BetrayNtArg{
			PlayerName: args[0].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_EscapedNt:
		arg := &pb.EscapedNtArg{
			PlayerName: args[0].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_EscapedReturnNt:
		arg := &pb.EscapedReturnNtArg{
			PlayerName: args[0].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_AutoDefOrderNt:
		arg := &pb.TargetCity{
			CityID: int32(args[0].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_SurrenderCity1Nt:
		fallthrough
	case pb.CampaignNoticeType_SurrenderCity3Nt:
		arg := &pb.SurrenderCity1NtArg{
			PlayerName:        args[0].(string),
			CityID:            int32(args[1].(int)),
			TargetCountryName: args[2].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_SurrenderCity2Nt:
		arg := &pb.SurrenderCity2NtArg{
			PlayerName: args[0].(string),
			CityID:     int32(args[1].(int)),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_SurrenderCountry1Nt:
		fallthrough
	case pb.CampaignNoticeType_SurrenderCountry2Nt:
		arg := &pb.SurrenderCountry1NtArg{
			PlayerName:        args[0].(string),
			TargetCountryName: args[1].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_AutocephalyNt2:
		arg := &pb.AutocephalyNt2Arg{
			CityID:         int32(args[0].(int)),
			Job:            args[1].(pb.CampaignJob),
			PlayerName:     args[2].(string),
			OldCountryName: args[3].(string),
			NewCountryName: args[4].(string),
		}
		argsData, _ = arg.Marshal()
	case pb.CampaignNoticeType_AutocephalyNt3:
		arg := &pb.AutocephalyNt3Arg{
			PlayerName:     args[0].(string),
			CityID:         int32(args[1].(int)),
			NewCountryName: args[2].(string),
		}
		argsData, _ = arg.Marshal()
	}

	n := newNotice(nb.genMaxID(), type_, string(argsData))
	if type_ == pb.CampaignNoticeType_AutocephalyVoteNt {
		n.setAutocephalyCity(args[0].(int))
		n.setAutocephalyPlayer(args[1].(common.UUid))
	}
	nb.notices = append(nb.notices, n)
	nb.noticesAttr.AppendMapAttr(n.attr)
	return n
}

func (nb *noticeBoard) sendNotice(type_ pb.CampaignNoticeType, args ...interface{}) {
	n := nb.newNotice(type_, args...)
	pnbs := noticeMgr.getAllPlayerNotices()
	for _, nb := range pnbs {
		nb.addNotice(n)
		nb.setWorldMaxID(n.getID())
	}
}

type countryNoticeBoard struct {
	noticeBoard
	countryID uint32
}

func newCountryNoticeBoard(countryID uint32, attr *attribute.AttrMgr) *countryNoticeBoard {
	nb := &countryNoticeBoard{
		countryID: countryID,
	}
	nb.attr = attr
	nb.initNotices()
	return nb
}

func (nb *countryNoticeBoard) sendNotice(type_ pb.CampaignNoticeType, args ...interface{}) {
	n := nb.newNotice(type_, args...)
	cry := countryMgr.getCountry(nb.countryID)
	if cry == nil {
		return
	}

	players := cry.getAllPlayers()
	id := n.getID()
	for _, p := range players {
		nb := noticeMgr.getPlayerNotice(p.getUid())
		if nb != nil {
			nb.addNotice(n)
			nb.setCountryMaxID(id)
		}
	}
}

type cityNoticeBoard struct {
	noticeBoard
	cityID int
}

func newCityNoticeBoard(cityID int, attr *attribute.AttrMgr) *cityNoticeBoard {
	nb := &cityNoticeBoard{
		cityID: cityID,
	}
	nb.attr = attr
	nb.initNotices()
	return nb
}

func (nb *cityNoticeBoard) sendNotice(type_ pb.CampaignNoticeType, args ...interface{}) {
	n := nb.newNotice(type_, args...)
	cty := cityMgr.getCity(nb.cityID)
	if cty == nil {
		return
	}

	players := cty.getAllPlayers()
	id := n.getID()
	for _, p := range players {
		nb := noticeMgr.getPlayerNotice(p.getUid())
		if nb != nil {
			nb.addNotice(n)
			nb.setCityMaxID(id)
		}
	}
}

type playerNoticeBoard struct {
	noticeBoard
	uid common.UUid
}

func newPlayerNoticeBoard(uid common.UUid, attr *attribute.AttrMgr) *playerNoticeBoard {
	nb := &playerNoticeBoard{
		uid: uid,
	}
	nb.attr = attr
	nb.initNotices()
	return nb
}

func (nb *playerNoticeBoard) getWorldMaxID() int {
	return nb.attr.GetInt("wMaxID")
}

func (nb *playerNoticeBoard) setWorldMaxID(id int) {
	nb.attr.SetInt("wMaxID", id)
}

func (nb *playerNoticeBoard) getCountryMaxID() int {
	return nb.attr.GetInt("crMaxID")
}

func (nb *playerNoticeBoard) setCountryMaxID(id int) {
	nb.attr.SetInt("crMaxID", id)
}

func (nb *playerNoticeBoard) getCityMaxID() int {
	return nb.attr.GetInt("ctMaxID")
}

func (nb *playerNoticeBoard) setCityMaxID(id int) {
	nb.attr.SetInt("ctMaxID", id)
}

func (nb *playerNoticeBoard) hasNew() bool {
	return nb.attr.GetBool("hasNew")
}

func (nb *playerNoticeBoard) addOfflineNotice(n *notice) {
	cn := n.copy()
	cn.setID(nb.genMaxID())
	nb.notices = append(nb.notices, cn)
	nb.noticesAttr.AppendMapAttr(cn.attr)
	nb.attr.SetBool("hasNew", true)
}

func (nb *playerNoticeBoard) addNotice(n *notice) {
	nb.addOfflineNotice(n)
	p := playerMgr.getPlayer(nb.uid)
	if p != nil {
		if p.isOnline() {
			p.agent.PushClient(pb.MessageID_S2C_CAMPAIGN_NOTIFY_RED_DOT, &pb.CampaignNotifyRedDotArg{
				Type: pb.CampaignNotifyRedDotArg_Notice,
			})
		}

		if p.agent != nil {
			p.agent.PushBackend(pb.MessageID_CA2G_CAMPAIGN_NOTICE_NOTIFY, n.packMsg())
		}
	}
}

func (nb *playerNoticeBoard) sendNotice(type_ pb.CampaignNoticeType, args ...interface{}) {
	n := nb.newNotice(type_, args...)
	nb.attr.SetBool("hasNew", true)
	p := playerMgr.getPlayer(nb.uid)
	if p != nil {
		if p.isOnline() {
			p.agent.PushClient(pb.MessageID_S2C_CAMPAIGN_NOTIFY_RED_DOT, &pb.CampaignNotifyRedDotArg{
				Type: pb.CampaignNotifyRedDotArg_Notice,
			})
		}

		if p.agent != nil {
			p.agent.PushBackend(pb.MessageID_CA2G_CAMPAIGN_NOTICE_NOTIFY, n.packMsg())
		}
	}
}

func (nb *playerNoticeBoard) fetchNotices() []*notice {
	nb.attr.SetBool("hasNew", false)
	nb.checkTimeout(time.Now().Unix())
	return nb.notices
}

func (nb *playerNoticeBoard) getNotices() []*notice {
	nb.checkTimeout(time.Now().Unix())
	return nb.notices
}
