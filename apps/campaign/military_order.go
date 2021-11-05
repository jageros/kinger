package main

import (
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
)

type militaryOrder struct {
	attr *attribute.MapAttr
	cty *city
	cityPath []int
}

func newMilitaryOrder(cty *city, moType pb.MilitaryOrderType, forage, amount int, cityPath []int) *militaryOrder {
	attr := attribute.NewMapAttr()
	attr.SetInt("type", int(moType))
	attr.SetInt("amount", amount)
	attr.SetInt("maxAmount", amount)
	attr.SetInt("forage", forage)
	cityPathAttr := attribute.NewListAttr()
	attr.SetListAttr("cityPath", cityPathAttr)
	for _, cityID := range cityPath {
		cityPathAttr.AppendInt(cityID)
	}
	return newMilitaryOrderByAttr(cty, attr)
}

func newMilitaryOrderByAttr(cty *city, attr *attribute.MapAttr) *militaryOrder {
	mo := &militaryOrder{
		attr: attr,
		cty: cty,
	}
	cityPathAttr := attr.GetListAttr("cityPath")
	if cityPathAttr != nil {
		cityPathAttr.ForEachIndex(func(index int) bool {
			mo.cityPath = append(mo.cityPath, cityPathAttr.GetInt(index))
			return true
		})
	}
	return mo
}

func (mo *militaryOrder) cancel() {
	mo.attr.SetBool("cancel", true)
	mo.cty.modifyResource(resForage, float64(mo.getAmount() * mo.getForage()))
}

func (mo *militaryOrder) getType() pb.MilitaryOrderType {
	return pb.MilitaryOrderType(mo.attr.GetInt("type"))
}

func (mo *militaryOrder) getTargetCity() int {
	pathLen := len(mo.cityPath)
	if pathLen > 0 {
		return mo.cityPath[pathLen-1]
	} else {
		return 0
	}
}

func (mo *militaryOrder) getCityPath() []int {
	return mo.cityPath
}

func (mo *militaryOrder) getForage() int {
	return mo.attr.GetInt("forage")
}

func (mo *militaryOrder) getAmount() int {
	return mo.attr.GetInt("amount")
}

func (mo *militaryOrder) setAmount(val int) {
	mo.attr.SetInt("amount", val)
}

func (mo *militaryOrder) getMaxAmount() int {
	return mo.attr.GetInt("maxAmount")
}

func (mo *militaryOrder) isCancel() bool {
	return mo.attr.GetBool("cancel")
}

func (mo *militaryOrder) packMsg() *pb.MilitaryOrder {
	return &pb.MilitaryOrder{
		Type: mo.getType(),
		Forage: int32(mo.getForage()),
		Amount: int32(mo.getAmount()),
		MaxAmount: int32(mo.getMaxAmount()),
		TargetCity: int32(mo.getTargetCity()),
	}
}
