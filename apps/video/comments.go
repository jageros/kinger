package main

import (
	"kinger/gopuppy/attribute"
	"sort"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"time"
)

type comments struct {
	attr *attribute.MapAttr
	battleIDKey string
}

func (c *comments) packMsg(uid common.UUid) *pb.VideoComments {
	p := videoMgr.getPlayer(uid)
	return &pb.VideoComments{
		ID: int32(c.getID()),
		Name: c.getName(),
		Content: c.getContent(),
		Like: int32(c.getLike()),
		HeadImgUrl: c.getHeadImgUrl(),
		Time: int32(c.getTime()),
		Country: c.getCountry(),
		Uid: uint64(c.getUid()),
		HeadFrame: c.getHeadFrame(),
		IsLike: p.isLikeComments(c.battleIDKey, c.getID()),
		CountryFlag: c.getCountryFlag(),
	}
}

func (c *comments) getID() int {
	return c.attr.GetInt("id")
}

func (c *comments) getLike() int {
	return c.attr.GetInt("like")
}
func (c *comments) setLike(amount int) {
	c.attr.SetInt("like", amount)
}

func (c *comments) getName() string {
	return c.attr.GetStr("name")
}

func (c *comments) getContent() string {
	return c.attr.GetStr("content")
}

func (c *comments) getHeadImgUrl() string {
	return c.attr.GetStr("headImgUrl")
}

func (c *comments) getCountryFlag() string {
	return c.attr.GetStr("countryFlag")
}

func (c *comments) getTime() int {
	return c.attr.GetInt("time")
}

func (c *comments) getCountry() string {
	return c.attr.GetStr("country")
}

func (c *comments) getUid() common.UUid {
	return common.UUid(c.attr.GetUInt64("uid"))
}

func (c *comments) getHeadFrame() string {
	return c.attr.GetStr("headFrame")
}

type baseCommentsList struct {
	battleIDKey string
	id2Comments map[int]*comments
	list []*comments
}

func (tl *baseCommentsList) getComments(id int) *comments {
	return tl.id2Comments[id]
}


func (tl *baseCommentsList) newCommentsByAttr(attr *attribute.MapAttr) *comments {
	return &comments{
		attr: attr,
		battleIDKey: tl.battleIDKey,
	}
}

func (tl *baseCommentsList) add(c *comments) {
	id := c.getID()
	if _, ok := tl.id2Comments[id]; ok {
		return
	}
	tl.id2Comments[id] = c
	tl.list = append(tl.list, c)
}

func (tl *baseCommentsList) newComments(id int, uid common.UUid, name, content, headImgUrl, country,
	headFrame, countryFlag string) *comments {
	attr := attribute.NewMapAttr()
	t := int(time.Now().Unix())
	attr.SetInt("id", id)
	attr.SetStr("name", name)
	attr.SetStr("content", content)
	attr.SetInt("time", t)
	attr.SetStr("headImgUrl", headImgUrl)
	attr.SetStr("countryFlag", countryFlag)
	attr.SetStr("country", country)
	attr.SetStr("headFrame", headFrame)
	attr.SetUInt64("uid", uint64(uid))
	return &comments{
		attr: attr,
		battleIDKey: tl.battleIDKey,
	}
}

type topCommentsList struct {
	baseCommentsList
}

func newTopCommentsList(battleID common.UUid, list []*attribute.MapAttr) *topCommentsList {
	tl := &topCommentsList{}
	tl.battleIDKey = battleID.String()
	tl.id2Comments = map[int]*comments{}
	for _, attr := range list {
		c := tl.newCommentsByAttr(attr)
		cid := c.getID()
		tl.id2Comments[cid] = c
		tl.list = append(tl.list, c)
	}
	tl.sort()
	return tl
}

func (tl *topCommentsList) likeComments(id int) int {
	if c, ok := tl.id2Comments[id]; ok {
		curLike := c.getLike() + 1
		c.setLike(curLike)
		tl.sort()
		return curLike
	} else {
		return 0
	}
}

func (tl *topCommentsList) sort() {
	sort.Sort(tl)
}

func (tl *topCommentsList) Len() int {
	if tl == nil {
		return 0
	}
	return len(tl.list)
}

func (tl *topCommentsList) Swap(i, j int) {
	tl.list[i], tl.list[j] = tl.list[j], tl.list[i]
}

func (tl *topCommentsList) Less(i, j int) bool {
	return tl.list[i].getLike() >= tl.list[j].getLike()
}

func (tl *topCommentsList) getPageComments(uid common.UUid, begin, amount int) []*pb.VideoComments {
	end := begin + amount
	size := tl.Len()
	if end > size {
		end = size
	}
	l := tl.list[begin:end]

	var ret []*pb.VideoComments
	for _, c := range l {
		msg :=  c.packMsg(uid)
		ret = append(ret, msg)
	}
	return ret
}

type newCommentsList struct {
	baseCommentsList
	top *topCommentsList
}

func newNewCommentsList(battleID common.UUid, top *topCommentsList, attrList []*attribute.MapAttr) *newCommentsList {
	nl := &newCommentsList{}
	nl.battleIDKey = battleID.String()
	nl.top = top
	nl.id2Comments = map[int]*comments{}

	for _, attr := range attrList {
		c := nl.newCommentsByAttr(attr)
		nl.list = append(nl.list, c)
		nl.id2Comments[c.getID()] = c
	}

	return nl
}

func (nl *newCommentsList) size() int {
	if nl == nil {
		return 0
	}
	return len(nl.list)
}

func (nl *newCommentsList) likeComments(id int) int {
	if c, ok := nl.id2Comments[id]; ok {
		curLike := c.getLike() + 1
		c.setLike(curLike)
		if curLike >= topCommentsLikeLimit {
			delete(nl.id2Comments, id)
			for i, c2 := range nl.list {
				if c == c2 {
					nl.list = append(nl.list[:i], nl.list[i+1:]...)
					break
				}
			}
			nl.top.add(c)
		}
		return curLike
	} else {
		return 0
	}
}

func (nl *newCommentsList) getPageComments(uid common.UUid, begin, amount int) []*pb.VideoComments {
	size := nl.size()
	begin = size - begin - 1
	if begin < 0 {
		return []*pb.VideoComments{}
	}
	end := begin - amount
	if end < 0 {
		end = 0
	}

	var ret []*pb.VideoComments
	for begin >= end {
		c := nl.list[begin]
		begin--
		ret = append(ret, c.packMsg(uid))
	}
	return ret
}
