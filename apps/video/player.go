package main

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
)

type playerSt struct {
	uid              common.UUid
	historyVideo     *attribute.AttrMgr
	likeCommentsAttr *attribute.AttrMgr
}

func newPlayer(uid common.UUid) *playerSt {
	return &playerSt{
		uid: uid,
	}
}

func (p *playerSt) save() {
	if p.historyVideo != nil {
		p.historyVideo.Save(false)
	}
	if p.likeCommentsAttr != nil {
		p.likeCommentsAttr.Save(false)
	}
}

func (p *playerSt) getHistoryVideos(page int) []*videoItem {
	p.loadHistoryVideoAttr()
	if p.historyVideo == nil {
		return []*videoItem{}
	}

	battleIDsAttr := p.historyVideo.GetListAttr("battleIDs")
	if battleIDsAttr == nil {
		return []*videoItem{}
	}
	amount := battleIDsAttr.Size()
	if amount <= 0 {
		return []*videoItem{}
	}
	var vis []*videoItem
	beginIdx := amount - 1 - ((page - 1) * 10)
	if beginIdx < 0 {
		return vis
	}
	endIdx := amount - 1 - page*10

	for i := beginIdx; i >= 0 && i > endIdx; i-- {
		battleID := common.UUid(battleIDsAttr.GetUInt64(i))
		vi := videoMgr.loadVideoItem(battleID, false)
		if vi != nil {
			vis = append(vis, vi)
		} else {
			battleIDsAttr.DelBySection(0, beginIdx+1)
			break
		}
	}
	return vis
}

func (p *playerSt) loadHistoryVideoAttr() {
	if p.historyVideo != nil {
		return
	}

	historyVideo := attribute.NewAttrMgr("historyBattles", p.uid)
	err := historyVideo.Load()
	if p.historyVideo != nil {
		return
	}

	if err != nil && err != attribute.NotExistsErr {
		return
	}

	p.historyVideo = historyVideo
	if err == attribute.NotExistsErr {
		p.historyVideo.SetListAttr("battleIDs", attribute.NewListAttr())
		//historyVideoAttr.Save(false)
	}
}

func (p *playerSt) saveHistoryVideo(videoID common.UUid) {
	p.loadHistoryVideoAttr()
	if p.historyVideo == nil {
		return
	}

	attr := p.historyVideo.GetListAttr("battleIDs")
	if attr == nil {
		attr = attribute.NewListAttr()
		p.historyVideo.SetListAttr("battleIDs", attr)
	}
	attr.AppendUInt64(uint64(videoID))
	p.historyVideo.Save(false)
}

func (p *playerSt) loadLikeCommentsAttr() {
	if p.likeCommentsAttr != nil {
		return
	}

	likeCommentsAttr := attribute.NewAttrMgr("likeComments", p.uid)
	err := likeCommentsAttr.Load()
	if err != nil && err != attribute.NotExistsErr {
		return
	}

	if p.likeCommentsAttr != nil {
		return
	}
	p.likeCommentsAttr = likeCommentsAttr
}

func (p *playerSt) isLikeComments(battleIDKey string, commentsID int) bool {
	p.loadLikeCommentsAttr()
	if p.likeCommentsAttr == nil {
		return false
	}

	vcAttr := p.likeCommentsAttr.GetListAttr(battleIDKey)
	if vcAttr == nil {
		return false
	}

	isLike := false
	vcAttr.ForEachIndex(func(index int) bool {
		if vcAttr.GetInt(index) == commentsID {
			isLike = true
			return false
		}
		return true
	})
	return isLike
}

func (p *playerSt) likeComments(battleIDKey string, commentsID int) {
	p.loadLikeCommentsAttr()
	if p.likeCommentsAttr == nil {
		return
	}

	vcAttr := p.likeCommentsAttr.GetListAttr(battleIDKey)
	if vcAttr == nil {
		vcAttr = attribute.NewListAttr()
		p.likeCommentsAttr.SetListAttr(battleIDKey, vcAttr)
	}
	vcAttr.AppendInt(commentsID)
}
