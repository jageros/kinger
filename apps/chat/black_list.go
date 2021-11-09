package main

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
)

var blackUidList *common.UInt64Set

func loadBlackList() {
	blackUidList = &common.UInt64Set{}
	attrs, _ := attribute.LoadAll("forbid_chat")
	for _, attr := range attrs {
		id, ok := attr.GetAttrID().(int64)
		if !ok {
			glog.Errorf("loadBlackList wrong uid %s", attr.GetAttrID())
			continue
		}
		blackUidList.Add(uint64(id))
	}
}
