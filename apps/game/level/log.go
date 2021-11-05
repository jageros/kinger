package level

import (
	"kinger/gopuppy/attribute"
	"strconv"
)

type levelLog struct {
	attr *attribute.MapAttr
}

func newLevelLog(cptAttr *attribute.MapAttr) *levelLog {
	attr := cptAttr.GetMapAttr("record")
	if attr == nil {
		attr = attribute.NewMapAttr()
		cptAttr.SetMapAttr("record", attr)
	}
	return &levelLog{attr: attr}
}

func (l *levelLog) getAttrByID(levelID int) *attribute.MapAttr {
	key := strconv.Itoa(levelID)
	attr := l.attr.GetMapAttr(key)
	if attr == nil {
		attr = attribute.NewMapAttr()
		l.attr.SetMapAttr(key, attr)
	}
	return attr
}

func (l *levelLog) onLevelBegin(levelID int) {
	attr := l.getAttrByID(levelID)
	attr.SetInt("cnt", attr.GetInt("cnt") + 1)
}

func (l *levelLog) onLevelWin(levelID int) {
	attr := l.getAttrByID(levelID)
	attr.SetInt("winCnt", attr.GetInt("winCnt") + 1)
}
