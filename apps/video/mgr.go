package main

import (
	"kinger/gopuppy/common/lru"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"kinger/gopuppy/common/evq"
	"kinger/gamedata"
	"kinger/common/consts"
	"time"
	"kinger/gopuppy/common/timer"
	"math/rand"
	"kinger/common/config"
)

const (
	topVideoMaxAmount = 20
	newVideoMinAmount = 20
	videoMaxLife      = 7 * 24 * 60 * 60
	newVideoLimitTime = 2 * 60 * 60
	pageCommentsAmount = 10
	topCommentsLikeLimit = 20
)

var videoMgr *videoMgrSt

type videoMgrSt struct {
	videoItemCache     *lru.LruCache
	videoDataCache     *lru.LruCache

	topOrNewItems map[common.UUid]*videoItem  // 所有区最新或热度变化最大的录像
	loadingVideoItems map[common.UUid]chan struct{}

	areaToTopVideos     map[int]*topVideoItemList  // 热度变化最大的录像
	areaToNewVideos     map[int]*newVideoItemList  // 最新录像

	players map[common.UUid]*playerSt
}

func newVideoMgr() {
	videoMgr = &videoMgrSt{
		topOrNewItems: map[common.UUid]*videoItem{},
		loadingVideoItems: map[common.UUid]chan struct{}{},
		areaToTopVideos: map[int]*topVideoItemList{},
		areaToNewVideos: map[int]*newVideoItemList{},
		players: map[common.UUid]*playerSt{},
	}
	videoMgr.initCache()
	videoMgr.loadVideoItemList()
	videoMgr.addTimer()
}

func (vm *videoMgrSt) addTimer() {
	timer.AddTicker(5*time.Minute, vm.checkNewVideoTimeout)
	timer.RunEveryHour(0, 0, vm.onTopVideosRefresh)
	timer.AddTicker(6 * time.Minute, vm.save)
}

func (vm *videoMgrSt) checkNewVideoTimeout() {
	for _, nl := range vm.areaToNewVideos {
		nl.checkNewTimeout()
	}
}

func (vm *videoMgrSt) onTopVideosRefresh() {
	if time.Now().Hour() % 3 != 0 {
		return
	}
	for _, tl := range vm.areaToTopVideos {
		tl.sort()
	}
}

func (vm *videoMgrSt) initCache() {
	var err error
	videoMgr.videoItemCache, err = lru.NewLruCache(1000, func(key interface{}, value interface{}) {
		value.(*videoItem).save()
	})
	if err != nil {
		panic(err)
	}

	videoMgr.videoDataCache, err = lru.NewLruCache(1000, nil)
	if err != nil {
		panic(err)
	}
}

func (vm *videoMgrSt) loadVideoItemList() {
	//areaGameData := gamedata.GetGameData(consts.AreaConfig).(*gamedata.AreaConfigGameData)
	//for _, areaConfg := range areaGameData.Areas {
	vm.newTopVideoItemList(1)
	vm.newNewVideoItemList(1)
	//}

	if config.GetConfig().IsOldXfServer() {
		vm.newTopVideoItemList(2)
		vm.newNewVideoItemList(2)
	}
}

func (vm *videoMgrSt) wrapArea(area int) int {
	if area == 1 && config.GetConfig().IsOldXfServer() {
		return 1
	} else {
		return 2
	}
}

func (vm *videoMgrSt) newTopVideoItemList(area int) *topVideoItemList {
	area = vm.wrapArea(area)
	if tl, ok := vm.areaToTopVideos[area]; ok {
		return tl
	} else {
		tl = newTopVideoItemList(area)
		vm.areaToTopVideos[area] = tl
		return tl
	}
}

func (vm *videoMgrSt) newNewVideoItemList(area int) *newVideoItemList {
	area = vm.wrapArea(area)
	if nl, ok := vm.areaToNewVideos[area]; ok {
		return nl
	} else {
		nl = newNewVideoItemList(area)
		vm.areaToNewVideos[area] = nl
		return nl
	}
}

func (vm *videoMgrSt) getTopVideos(area int) *topVideoItemList {
	area = vm.wrapArea(area)
	return vm.areaToTopVideos[area]
}

func (vm *videoMgrSt) getNewVideos(area int) *newVideoItemList {
	area = vm.wrapArea(area)
	return vm.areaToNewVideos[area]
}

func (vm *videoMgrSt) loadVideo(battleID common.UUid) *pb.VideoBattleData {
	if videoData, ok := vm.videoDataCache.Get(battleID); ok {
		vd, ok := videoData.(*pb.VideoBattleData)
		if !ok {
			return nil
		} else {
			return vd
		}
	} else {
		attr := attribute.NewAttrMgr("battleVideo", battleID, true)
		err := attr.Load()
		if err != nil {
			if err == attribute.NotExistsErr {
				vm.videoDataCache.Add(battleID, nil)
			}
			return nil
		}

		data := []byte(attr.GetStr("data"))
		videoData := &pb.VideoBattleData{}
		err = videoData.Unmarshal(data)
		if err != nil {
			return nil
		}
		vm.videoDataCache.Add(battleID, videoData)
		return videoData
	}
}

func (vm *videoMgrSt) addTopOrNewVideoItem(vi *videoItem) {
	battleID := vi.getBattleID()
	vm.videoItemCache.Remove(battleID)
	vm.topOrNewItems[battleID] = vi
	vi.setRef( vi.getRef() + 1 )
}

func (vm *videoMgrSt) cacheVideoItem(vi *videoItem) {
	battleID := vi.getBattleID()
	if _, ok := vm.topOrNewItems[battleID]; ok {
		ref := vi.getRef()
		ref--
		if ref < 0 {
			ref = 0
		}
		vi.setRef(ref)
		if ref > 0 {
			return
		}
	}

	delete(vm.topOrNewItems, battleID)
	vm.videoItemCache.Add(battleID, vi)
}

func (vm *videoMgrSt) loadVideoItem(battleID common.UUid, justFromCache bool, ignoreTimeoutArg ...interface{}) *videoItem {
	vi := vm.topOrNewItems[battleID]
	if vi != nil {
		return vi
	}

	vi2, ok := vm.videoItemCache.Get(battleID)
	if ok {
		return vi2.(*videoItem)
	} else if justFromCache {
		return nil
	}

	if loading, ok := vm.loadingVideoItems[battleID]; ok {
		evq.Await(func() {
			<- loading
		})
		return vm.loadVideoItem(battleID, true)
	}

	c := make(chan struct{})
	vm.loadingVideoItems[battleID] = c

	defer func() {
		delete(vm.loadingVideoItems, battleID)
		close(c)
	}()

	attr := attribute.NewAttrMgr("videoItem", battleID)
	if err := attr.Load(); err != nil {
		return nil
	}

	vi = newVideoItemByAttr(battleID, attr)
	var ignoreTimeout bool
	if len(ignoreTimeoutArg) > 0 {
		ignoreTimeout, _ = ignoreTimeoutArg[0].(bool)
	}
	if !ignoreTimeout && vi.isTimeout() {
		return nil
	}

	cacheVi := vm.loadVideoItem(battleID, true)
	if cacheVi != nil {
		return cacheVi
	}

	vm.videoItemCache.Add(battleID, vi)
	return vi
}

func (vm *videoMgrSt) getPlayer(uid common.UUid) *playerSt {
	if p, ok := vm.players[uid]; ok {
		return p
	} else {
		p = newPlayer(uid)
		vm.players[uid] = p
		return p
	}
}

func (vm *videoMgrSt) delPlayer(uid common.UUid) {
	if p, ok := vm.players[uid]; ok {
		p.save()
		delete(vm.players, uid)
	}
}

func (vm *videoMgrSt) save() {
	for _, p := range vm.players {
		p.save()
	}

	for _, vi := range vm.topOrNewItems {
		vi.save()
	}

	for _, vl := range vm.areaToTopVideos {
		vl.save()
	}

	for _, vl := range vm.areaToNewVideos {
		vl.save()
	}

	vm.videoItemCache.ForEach(func(value interface{}) {
		value.(*videoItem).save()
	})
}

func (vm *videoMgrSt) randomVideos(area int) []*videoItem {
	area = vm.wrapArea(area)
	var items []*videoItem
	var itemsSet common.UInt64Set
	topVideos := vm.getTopVideos(area)
	if topVideos != nil {
		items, itemsSet = topVideos.getItems()
	} else {
		itemsSet = common.UInt64Set{}
	}

	if len(items) >= 40 {
		return items
	}

	newVideos := vm.getNewVideos(area)
	if newVideos != nil {
		newVideoItems, _ := newVideos.getItems()
		for i := range newVideoItems {
			j := rand.Intn(i + 1)
			newVideoItems[i], newVideoItems[j] = newVideoItems[j], newVideoItems[i]
		}

		for _, vi := range newVideoItems {
			if itemsSet.Contains(uint64(vi.getBattleID())) {
				continue
			}
			items = append(items, vi)
			if len(items) >= 40 {
				return items
			}
		}
	}

	return items
}

func (vm *videoMgrSt) saveVideoItem(battleID common.UUid, fighter1, fighter2 *pb.VideoFighterData, winnerUid common.UUid) {
	if battleID <= 0 {
		return
	}

	winner := fighter1
	loser := fighter2
	if winner.Uid != uint64(winnerUid) {
		loser = fighter2
		winner = fighter1
	}

	team := 1000
	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	if !winner.IsRobot {
		vm.getPlayer(common.UUid(winner.Uid)).saveHistoryVideo(battleID)
		team = rankGameData.GetTeamByStar(int(winner.PvpScore))
	}
	if !loser.IsRobot {
		vm.getPlayer(common.UUid(loser.Uid)).saveHistoryVideo(battleID)
		team2 := rankGameData.GetTeamByStar(int(loser.PvpScore))
		if team2 < team {
			team = team2
		}
	}

	vi := newVideoItem(battleID, fighter1, fighter2, winnerUid, team)
	vi.save()
}

func (vm *videoMgrSt) onVideoHotUpdate(vi *videoItem) {
	if vi.getSharePlayer() <= 0 {
		return
	}

	area := vi.getArea()
	if area <= 0 {
		return
	}

	vm.newTopVideoItemList(area).onVideoHotUpdate(vi)
}

func (vm *videoMgrSt) shareVideo(vi *videoItem) {
	area := vi.getArea()
	if area <= 0 {
		return
	}

	topVideos := vm.newTopVideoItemList(area)
	if topVideos.size() < topVideoMaxAmount {
		topVideos.append(vi)
	} else {
		vm.newNewVideoItemList(area).append(vi)
	}
}
