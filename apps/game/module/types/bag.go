package types

import (
	"kinger/gopuppy/attribute"
)

type IItem interface {
	GetType() int
	GetAmount() int
	GetID() string
	ModifyAmount(amount int)
	GetAttr() *attribute.MapAttr
	GetName() string
	GetGmID() string
}

type IEquipItem interface {
	IItem
	IRestAble
	GetOwner() uint32
	SetOwner(cardID uint32)
	SetObtainTime(tim int64)
	GetObtainTime() int64
}

// 可回退的东西
type IRestAble interface {
	SetFrom(from int)
	GetFrom() int
	Reset(player IPlayer)
	GetVersion() int
}
