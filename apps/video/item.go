package main

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"sort"
	"time"
	"strconv"
)

type baseVideoItemList struct {
	attrMgr           *attribute.AttrMgr
	itemsAttr         *attribute.ListAttr
	itemsSet          common.UInt64Set
}

func (vl *baseVideoItemList) initAttr() {
	vl.itemsSet = common.UInt64Set{}
	itemsAttr := vl.attrMgr.GetListAttr("data")
	if itemsAttr == nil {
		itemsAttr = attribute.NewListAttr()
		vl.attrMgr.SetListAttr("data", itemsAttr)
	}

	itemsAttr2 := attribute.NewListAttr()
	itemsAttr.ForEachIndex(func(index int) bool {
		id := itemsAttr.GetUInt64(index)
		battleID := common.UUid(id)
		if vl.itemsSet.Contains(id) {
			return true
		}

		vi := videoMgr.loadVideoItem(battleID, false)
		if vi != nil {
			itemsAttr2.AppendUInt64(id)
			vl.itemsSet.Add(id)
			videoMgr.addTopOrNewVideoItem(vi)
		}
		return true
	})

	if itemsAttr2.Size() != itemsAttr.Size() {
		vl.itemsAttr = itemsAttr2
		vl.attrMgr.SetListAttr("data", itemsAttr2)
	} else {
		vl.itemsAttr = itemsAttr
	}
}

func (vl *baseVideoItemList) save() {
	vl.attrMgr.Save(false)
}

func (vl *baseVideoItemList) append(vi *videoItem) {
	battleID := uint64(vi.getBattleID())
	if vl.itemsSet.Contains(battleID) {
		return
	}

	vl.itemsAttr.AppendUInt64(battleID)
	vl.itemsSet.Add(battleID)
	videoMgr.addTopOrNewVideoItem(vi)
}

func (vl *baseVideoItemList) getItems() ([]*videoItem, common.UInt64Set) {
	var items []*videoItem
	vl.itemsAttr.ForEachIndex(func(index int) bool {
		vi := videoMgr.loadVideoItem(common.UUid(vl.itemsAttr.GetUInt64(index)), true)
		if vi != nil {
			items = append(items, vi)
		}
		return true
	})
	return items, vl.itemsSet
}

type topVideoItemList struct {
	baseVideoItemList
	hotChangeItems    []common.UUid
	hotChangeItemsSet common.UInt64Set
}

func newTopVideoItemList(area int) *topVideoItemList {
	attrMgr := attribute.NewAttrMgr("topVideos", area)
	attrMgr.Load()

	tl := &topVideoItemList{
		hotChangeItemsSet: common.UInt64Set{},
	}
	tl.attrMgr = attrMgr
	tl.initAttr()

	return tl
}

func (tl *topVideoItemList) size() int {
	if tl == nil {
		return 0
	}
	return tl.itemsAttr.Size()
}

func (tl *topVideoItemList) Len() int {
	if tl == nil {
		return 0
	}
	return len(tl.hotChangeItems)
}

func (tl *topVideoItemList) Swap(i, j int) {
	tl.hotChangeItems[i], tl.hotChangeItems[j] =  tl.hotChangeItems[j], tl.hotChangeItems[i]
}

func (tl *topVideoItemList) Less(i, j int) bool {
	vi1 := videoMgr.loadVideoItem(tl.hotChangeItems[i], true)
	if vi1 == nil {
		return false
	}
	vi2 := videoMgr.loadVideoItem(tl.hotChangeItems[j], true)
	if vi2 == nil {
		return true
	}
	return vi1.isBetterThan(vi2)
}

func (tl *topVideoItemList) sort() {
	sort.Sort(tl)
	newItemsAttr := attribute.NewListAttr()
	newItemSet := common.UInt64Set{}
	for _, battleID := range tl.hotChangeItems {
		vi := videoMgr.loadVideoItem(battleID, true)
		if vi != nil {
			vi.saveLastHot()

			if newItemSet.Size() < topVideoMaxAmount {
				newItemsAttr.AppendUInt64(uint64(battleID))
				newItemSet.Add(uint64(battleID))
			} else {
				// 没入选这次的高热度录像，移除常驻缓存
				videoMgr.cacheVideoItem(vi)
			}
		}
	}

	hotChangeItemLen := newItemsAttr.Size()
	needAmount := topVideoMaxAmount - hotChangeItemLen
	tl.itemsAttr.ForEachIndex(func(index int) bool {
		battleID := tl.itemsAttr.GetUInt64(index)
		if newItemSet.Contains(battleID) {

		} else if needAmount > 0 {
			// 最新高热度的录像不够，用之前高热度的录像补
			needAmount --
			newItemsAttr.AppendUInt64(uint64(battleID))
			newItemSet.Add(uint64(battleID))
		} else {
			// 移除上一次高热度录像的常驻缓存
			vi := videoMgr.loadVideoItem(common.UUid(battleID), true)
			if vi != nil {
				videoMgr.cacheVideoItem(vi)
			}
		}
		return true
	})

	tl.itemsAttr = newItemsAttr
	tl.itemsSet = newItemSet
	tl.hotChangeItems = []common.UUid{}
	tl.hotChangeItemsSet = common.UInt64Set{}
	tl.attrMgr.SetListAttr("data", tl.itemsAttr)
}

func (tl *topVideoItemList) onVideoHotUpdate(vi *videoItem) {
	battleID := uint64(vi.getBattleID())
	if tl.hotChangeItemsSet.Contains(battleID) {
		return
	}

	tl.hotChangeItemsSet.Add(battleID)
	tl.hotChangeItems = append(tl.hotChangeItems, common.UUid(battleID))
	videoMgr.addTopOrNewVideoItem(vi)
}

func (tl *topVideoItemList) getLastItem() *videoItem {
	n := tl.itemsAttr.Size()
	if n <= 0 {
		return nil
	}
	return videoMgr.loadVideoItem(common.UUid(tl.itemsAttr.GetUInt64(n - 1)), true)
}

type newVideoItemList struct {
	baseVideoItemList
}

func newNewVideoItemList(area int) *newVideoItemList {
	attrMgr := attribute.NewAttrMgr("newVideos", area)
	attrMgr.Load()

	nl := &newVideoItemList{}
	nl.attrMgr = attrMgr
	nl.initAttr()
	return nl
}

func (nl *newVideoItemList) checkNewTimeout() {
	now := time.Now().Unix()
	var timeOutItems []*videoItem
	nl.itemsAttr.ForEachIndex(func(index int) bool {
		battleID := nl.itemsAttr.GetUInt64(index)
		vi := videoMgr.loadVideoItem(common.UUid(battleID), true)
		if vi != nil && now - int64(vi.getTime()) < newVideoLimitTime {
			if index > 0 {
				nl.itemsAttr.DelBySection(0, index)
			}
			return false
		}

		if vi != nil {
			timeOutItems = append(timeOutItems, vi)
		} else {
			nl.itemsSet.Remove(battleID)
		}
		return true
	})

	for _, vi := range timeOutItems {
		videoMgr.cacheVideoItem(vi)
	}
}

type videoItem struct {
	attr     *attribute.AttrMgr
	battleID common.UUid
	commentsAttr *attribute.ListAttr
	topComments *topCommentsList
	newComments *newCommentsList
	hot int
	ref int
}

func newVideoItemByAttr(battleID common.UUid, attr *attribute.AttrMgr) *videoItem {
	vi := &videoItem{
		attr:     attr,
		battleID: battleID,
		hot: -1,
	}
	vi.commentsAttr = attr.GetListAttr("comments2")
	if vi.commentsAttr == nil {
		vi.commentsAttr = attribute.NewListAttr()
		attr.SetListAttr("comments2", vi.commentsAttr)
	}
	return vi
}

func newVideoItem(battleID common.UUid, fighter1, fighter2 *pb.VideoFighterData, winner common.UUid, team int) *videoItem {
	attr := attribute.NewAttrMgr("videoItem", battleID)
	attr.SetUInt64("winner", uint64(winner))
	attr.SetInt("team", team)
	attr.SetUInt64("fighter1", fighter1.Uid)
	attr.SetUInt64("fighter2", fighter2.Uid)
	attr.SetStr("fighter1Name", fighter1.Name)
	attr.SetStr("fighter2Name", fighter2.Name)
	attr.SetInt("f1NameText", int(fighter1.NameText))
	attr.SetInt("f2NameText", int(fighter2.NameText))
	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	attr.SetInt("fighter1PvpLevel", rankGameData.GetPvpLevelByStar(int(fighter1.PvpScore)))
	attr.SetInt("fighter2PvpLevel", rankGameData.GetPvpLevelByStar(int(fighter2.PvpScore)))
	attr.SetInt("time", int(time.Now().Unix()))

	fighter1CardsAttr := attribute.NewListAttr()
	attr.SetListAttr("f1Cards", fighter1CardsAttr)
	for _, c := range fighter1.HandCards {
		cAttr := attribute.NewMapAttr()
		cAttr.SetUInt32("gcardID", c.GCardID)
		cAttr.SetStr("skin", c.Skin)
		cAttr.SetStr("equip", c.Equip)
		fighter1CardsAttr.AppendMapAttr(cAttr)
	}

	fighter2CardsAttr := attribute.NewListAttr()
	attr.SetListAttr("f2Cards", fighter2CardsAttr)
	for _, c := range fighter2.HandCards {
		cAttr := attribute.NewMapAttr()
		cAttr.SetUInt32("gcardID", c.GCardID)
		cAttr.SetStr("skin", c.Skin)
		cAttr.SetStr("equip", c.Equip)
		fighter2CardsAttr.AppendMapAttr(cAttr)
	}

	likePlayersAttr := attribute.NewMapAttr()
	attr.SetMapAttr("likePlayers", likePlayersAttr)

	vi := newVideoItemByAttr(battleID, attr)
	return vi
}

func (vi *videoItem) getRef() int {
	return vi.ref
}

func (vi *videoItem) setRef(val int) {
	vi.ref = val
}

func (vi *videoItem) getTeam() int {
	return vi.attr.GetInt("team")
}

func (vi *videoItem) saveLastHot() {
	vi.attr.SetInt("lastHot", vi.getHot())
}

func (vi *videoItem) getLastHot() int {
	return vi.attr.GetInt("lastHot")
}

func (vi *videoItem) getName() string {
	return vi.attr.GetStr("name")
}

func (vi *videoItem) getCommentsAmount() int {
	return vi.commentsAttr.Size()
}

func (vi *videoItem) isBetterThan(oth *videoItem) bool {
	hot1 := vi.getHot()
	lastHot1 := vi.getLastHot()
	change1 := hot1 - lastHot1
	hot2 := oth.getHot()
	lastHot2 := oth.getLastHot()
	change2 := hot2 - lastHot2

	if change1 > change2 {
		return true
	} else if change2 > change1 {
		return false
	}

	return vi.getTime() > oth.getTime()
}

func (vi *videoItem) isLike(uid common.UUid) bool {
	likePlayersAttr := vi.getLikePlayers()
	return likePlayersAttr.Get(uid.String()) != nil
}

func (vi *videoItem) packMsg(uid common.UUid) *pb.VideoItem {
	msg := &pb.VideoItem{
		VideoID: uint64(vi.getBattleID()),
		Fighter1: &pb.VideoFighter{
			Name:       vi.getFighter1Name(),
			Uid:        uint64(vi.getFighter1Uid()),
			FightCards: vi.getFighterCards(true),
			NameText: int32(vi.getFighter1NameText()),
		},
		Fighter2: &pb.VideoFighter{
			Name:       vi.getFighter2Name(),
			Uid:        uint64(vi.getFighter2Uid()),
			FightCards: vi.getFighterCards(false),
			NameText: int32(vi.getFighter2NameText()),
		},
		WinnerUid:  uint64(vi.getWinner()),
		WatchTimes: int32(vi.getWatchTimes()),
		Like:       int32(vi.getLikeAmount()),
		Time:       int32(vi.getTime()),
		IsLike:     vi.isLike(uid),
		Name: vi.getName(),
		CommentsAmount: int32(vi.getCommentsAmount()),
	}

	sharePalyer := vi.getSharePlayer()
	if sharePalyer > 0 {
		if sharePalyer == vi.getFighter1Uid() {
			msg.SharePlayerName = vi.getFighter1Name()
		} else {
			msg.SharePlayerName = vi.getFighter2Name()
		}
	}

	return msg
}

func (vi *videoItem) getHot() int {
	if vi.hot < 0 {
		vi.hot = vi.getWatchTimes() + vi.getLikeAmount()*2 + (vi.topComments.Len()+vi.newComments.size())*5
	}
	return vi.hot
}

func (vi *videoItem) getBattleID() common.UUid {
	return vi.battleID
}

func (vi *videoItem) getFighter1Uid() common.UUid {
	return common.UUid(vi.attr.GetUInt64("fighter1"))
}

func (vi *videoItem) getFighter2Uid() common.UUid {
	return common.UUid(vi.attr.GetUInt64("fighter2"))
}

func (vi *videoItem) getWinner() common.UUid {
	return common.UUid(vi.attr.GetUInt64("winner"))
}

func (vi *videoItem) oldGetFighterCards(isFighter1 bool) []*pb.SkinGCard {
	// 兼容老数据
	var cardsAttr *attribute.ListAttr
	var cardSkinsAttr *attribute.MapAttr
	if isFighter1 {
		cardsAttr = vi.attr.GetListAttr("fighter1Cards")
		cardSkinsAttr = vi.attr.GetMapAttr("fighter1CardSkins")
	} else {
		cardsAttr = vi.attr.GetListAttr("fighter2Cards")
		cardSkinsAttr = vi.attr.GetMapAttr("fighter2CardSkins")
	}

	cards := make([]*pb.SkinGCard, cardsAttr.Size(), cardsAttr.Size())
	cardsAttr.ForEachIndex(func(index int) bool {
		c := &pb.SkinGCard{
			GCardID: cardsAttr.GetUInt32(index),
		}
		if cardSkinsAttr != nil {
			c.Skin = cardSkinsAttr.GetStr(strconv.Itoa(int(c.GCardID)))
		}
		cards[index] = c
		return true
	})
	return cards
}

func (vi *videoItem) getFighterCards(isFighter1 bool) []*pb.SkinGCard {
	var cardsAttr *attribute.ListAttr
	if isFighter1 {
		cardsAttr = vi.attr.GetListAttr("f1Cards")
	} else {
		cardsAttr = vi.attr.GetListAttr("f2Cards")
	}
	if cardsAttr == nil {
		return vi.oldGetFighterCards(isFighter1)
	}

	cards := make([]*pb.SkinGCard, cardsAttr.Size(), cardsAttr.Size())
	cardsAttr.ForEachIndex(func(index int) bool {
		cAttr := cardsAttr.GetMapAttr(index)
		cards[index] = &pb.SkinGCard{
			GCardID: cAttr.GetUInt32("gcardID"),
			Skin: cAttr.GetStr("skin"),
			Equip: cAttr.GetStr("equip"),
		}
		return true
	})
	return cards
}

func (vi *videoItem) getFighter1Name() string {
	return vi.attr.GetStr("fighter1Name")
}

func (vi *videoItem) getFighter2Name() string {
	return vi.attr.GetStr("fighter2Name")
}

func (vi *videoItem) getFighter1NameText() int {
	return vi.attr.GetInt("f1NameText")
}

func (vi *videoItem) getFighter2NameText() int {
	return vi.attr.GetInt("f2NameText")
}

func (vi *videoItem) getTime() int {
	return vi.attr.GetInt("time")
}

func (vi *videoItem) getWatchTimes() int {
	return vi.attr.GetInt("watchTimes")
}

func (vi *videoItem) getArea() int {
	return videoMgr.wrapArea(vi.attr.GetInt("area"))
}

func (vi *videoItem) getLikePlayers() *attribute.MapAttr {
	return vi.attr.GetMapAttr("likePlayers")
}

func (vi *videoItem) getLikeAmount() int {
	return vi.getLikePlayers().Size()
}

func (vi *videoItem) getSharePlayer() common.UUid {
	return common.UUid(vi.attr.GetUInt64("sharePlayer"))
}

func (vi *videoItem) isTimeout(args ...int64) bool {
	var now int64
	if len(args) > 0 {
		now = args[0]
	} else {
		now = time.Now().Unix()
	}
	return now-int64(vi.getTime()) >= videoMaxLife
}

func (vi *videoItem) watch() (videoData *pb.VideoBattleData, curWatchTimes, curLike int) {
	videoData = videoMgr.loadVideo(vi.getBattleID())
	if videoData == nil {
		return
	}
	curWatchTimes = vi.getWatchTimes() + 1
	vi.attr.SetInt("watchTimes", curWatchTimes)
	curLike = vi.getLikeAmount()
	videoMgr.onVideoHotUpdate(vi)
	return
}

func (vi *videoItem) like(uid common.UUid) (curWatchTimes, curLike int) {
	likePlayersAttr := vi.getLikePlayers()
	key := uid.String()
	curWatchTimes = vi.getWatchTimes()
	if likePlayersAttr.Get(key) != nil {
		curLike = likePlayersAttr.Size()
		return
	} else {
		likePlayersAttr.Set(key, true)
		curLike = likePlayersAttr.Size()
		videoMgr.onVideoHotUpdate(vi)
		return
	}
}

func (vi *videoItem) share(uid common.UUid, name string, area int) {
	if vi.getSharePlayer() > 0 {
		return
	}
	vi.attr.SetUInt64("sharePlayer", uint64(uid))
	vi.attr.SetStr("name", name)
	vi.attr.SetInt("area", area)
	videoMgr.shareVideo(vi)
}

func (vi *videoItem) save() {
	vi.attr.Save(false)
}

func (vi *videoItem) loadComments() {
	if vi.newComments != nil && vi.topComments != nil {
		return
	}

	var nl []*attribute.MapAttr
	var tl []*attribute.MapAttr
	vi.commentsAttr.ForEachIndex(func(index int) bool {
		attr := vi.commentsAttr.GetMapAttr(index)
		if attr.GetInt("like") >= topCommentsLikeLimit {
			tl = append(tl, attr)
		} else {
			nl = append(nl, attr)
		}
		return true
	})

	vi.topComments = newTopCommentsList(vi.getBattleID(), tl)
	vi.newComments = newNewCommentsList(vi.getBattleID(), vi.topComments, nl)
}

func (vi *videoItem) genCommentsID() int {
	maxID := vi.attr.GetInt("maxCommentsID")
	if maxID <= 0 {
		maxID = 1
	}
	vi.attr.SetInt("maxCommentsID", maxID + 1)
	return maxID
}

func (vi *videoItem) comments(uid common.UUid, name, content, headImgUrl, country, headFrame, countryFlag string) *comments  {
	vi.loadComments()
	id := vi.genCommentsID()
	c := vi.newComments.newComments(id, uid, name, content, headImgUrl, country, headFrame, countryFlag)
	vi.newComments.add(c)
	vi.commentsAttr.AppendMapAttr(c.attr)
	videoMgr.onVideoHotUpdate(vi)
	return c
}

func (vi *videoItem) likeComments(uid common.UUid, id int) int {
	vi.loadComments()
	var curLike int
	if vi.newComments.getComments(id) == nil {
		curLike = vi.topComments.likeComments(id)
	} else {
		curLike = vi.newComments.likeComments(id)
	}

	videoMgr.getPlayer(uid).likeComments(vi.getBattleID().String(), id)
	return curLike
}

func (vi *videoItem) getComments(uid common.UUid, curAmount int) ([]*pb.VideoComments, bool) {
	vi.loadComments()
	var l []*pb.VideoComments
	newLen := vi.newComments.size()
	topLen := vi.topComments.Len()
	if curAmount >= newLen + topLen {
		return l, false
	}

	totalAmount := curAmount + pageCommentsAmount
	needAmount := pageCommentsAmount
	if curAmount < topLen {
		l = vi.topComments.getPageComments(uid, curAmount, needAmount)
		curAmount = 0
		needAmount -= len(l)
	} else {
		curAmount -= topLen
	}

	if needAmount > 0 {
		l2 := vi.newComments.getPageComments(uid, curAmount, needAmount)
		needAmount -= len(l2)
		l = append(l, l2...)
	}

	return l, totalAmount < newLen + topLen
}
