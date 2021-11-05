package player

import (
	"kinger/common/consts"
	"time"
)

// 防止客户端1秒内连续点击多次，其实应该在客户端做的

var type2LimitTime = map[int]int64 {
	consts.FmcMatch: 1,
}

type forbidMultiClick struct {
	type2LastTime map[int]int64
}

func newForbidMultiClick() *forbidMultiClick {
	return &forbidMultiClick{
		type2LastTime: map[int]int64{},
	}
}

func (fmc *forbidMultiClick) isForbid(type_ int) bool {
	limitTime := type2LimitTime[type_]
	if limitTime <= 0 {
		limitTime = 2
	}

	lastTime := fmc.type2LastTime[type_]
	now := time.Now().Unix()
	if lastTime > 0 && now - lastTime < limitTime {
		return true
	}
	fmc.type2LastTime[type_] = now
	return false
}
