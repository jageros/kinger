package logic

import (
	"kinger/gopuppy/common"
	"kinger/gopuppy/attribute"
)

var uid2Region = map[common.UUid]uint32{}

func cacheAgentRegion(uid common.UUid, region uint32) {
	uid2Region[uid] = region
}

func SaveAgentRegion(uid common.UUid, region uint32) {
	uid2Region[uid] = region
	attr := attribute.NewAttrMgr("player_region", uid, true)
	attr.SetUInt32("region", region)
	attr.Save(false)
}

func GetAgentRegion(uid common.UUid) uint32 {
	if region, ok := uid2Region[uid]; ok {
		return region
	}

	attr := attribute.NewAttrMgr("player_region", uid, true)
	if err := attr.Load(); err != nil {
		return 0
	}

	region := attr.GetUInt32("region")
	uid2Region[uid] = region
	return region
}
