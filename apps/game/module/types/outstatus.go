package types

import "kinger/proto/pb"

type IOutStatus interface {
	GetID() string
	GetRemainTime() int
	Over(leftTime int, args ...interface{})
	PackMsg() *pb.OutStatus
	GetLeftTime() int
	SetLeftTime(leftTime int)
}

type IBuff interface {
	IOutStatus
	GetBuffID() int
}
