package api

import (
	"kinger/gopuppy/common"
)

func hashUUid(id common.UUid) int {
	h := uint64(id)
	h = (^h) + (h << 18)
	h = h ^ (h >> 31)
	h = h * 2
	h = h ^ (h >> 11)
	h = h + (h << 6)
	h = h ^ (h >> 22)
	h2 := int(h)
	if h2 < 0 {
		h2 = -h2
	}
	return h2
}

func hashAppID(appID uint32) int {
	h := ^appID + (appID << 15)
	h = h ^ (h >> 12)
	h = h + (h << 2)
	h = h ^ (h >> 4)
	h = h * 2057
	h = h ^ (h >> 16)
	h2 := int(h)
	if h2 < 0 {
		h2 = -h2
	}
	return h2
}

func hashString(s string) int {
	var h int
	for _, c := range s {
		h = h*131 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}
